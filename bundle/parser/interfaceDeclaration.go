package parser

import (
	"fmt"
	"main/bundle/utils"
	"strings"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// 解析 interface 声明，递归去查找interface里边的类型
// - 如果有引用外部类型的就找出来
// - 如果有应用到其他的ts语法的也要找出来，比如：extends、omit等

type TypeReference struct {
	Name     string
	Location []string // 保留设计，类型的位置，用.隔开引用的位置，例如：School.student.name
	IsExtend bool     // 是否继承，true表示继承，false表示member中引用的
}

type InterfaceDeclarationResult struct {
	Name      string // 名称
	Raw       string // 源码
	Reference map[string]TypeReference
}

func NewInterfaceDeclarationResult(node *ast.Node, sourceCode string) *InterfaceDeclarationResult {
	raw := utils.GetNodeText(node.AsNode(), sourceCode)

	return &InterfaceDeclarationResult{
		Name:      "",
		Raw:       raw,
		Reference: make(map[string]TypeReference),
	}
}

// 分析接口的主要结构，包括：
// 1. 接口名称。
// 2. 继承关系（通过 analyzeHeritageClause）。
// 3. 接口成员（通过 analyzeMember）。
func (inter *InterfaceDeclarationResult) analyzeInterfaces(interfaceDecl *ast.InterfaceDeclaration) {
	interfaceName := interfaceDecl.Name().AsIdentifier().Text
	inter.Name = interfaceName

	// 分析接口的继承关系
	inter.analyzeHeritageClause(interfaceDecl, interfaceName)

	// 分析接口的成员
	if interfaceDecl.Members != nil {
		for _, member := range interfaceDecl.Members.Nodes {
			memberTypeName, memberLocation := AnalyzeMember(member, interfaceName)
			if memberTypeName != "" && memberLocation != "" {
				for i, typeName := range strings.Split(memberTypeName, ",") {
					inter.addTypeReference(typeName, strings.Split(memberLocation, ",")[i], false)
				}
			}
		}
	}
}

// 分析接口的继承子句（extends）。
// 1. 找出接口继承的其他接口。
// 2. 提取继承的接口名称及其类型参数。
func (inter *InterfaceDeclarationResult) analyzeHeritageClause(interfaceDecl *ast.InterfaceDeclaration, interfaceName string) {
	// 获取 extends 子句元素
	extendsElements := ast.GetExtendsHeritageClauseElements(interfaceDecl.AsNode())

	// 处理每个 extends 元素
	for _, node := range extendsElements {
		expression := node.Expression()
		if ast.IsIdentifier(expression) {
			name := expression.AsIdentifier().Text
			// 如果不是工具类型，直接添加到依赖列表
			if !(utils.IsUtilityType(name)) {
				inter.addTypeReference(name, "", true)
			}
		} else if ast.IsPropertyAccessExpression(expression) {
			// 属性访问表达式，如 "module.Interface"
			name := entityNameToString(expression)
			inter.addTypeReference(name, "", true)
		}

		// 处理类型参数，无论是否是工具类型都提取参数中的依赖
		if node.TypeArguments != nil {
			for _, typeArg := range node.TypeArguments() {
				name, _ := AnalyzeType(typeArg, "")
				inter.addTypeReference(name, "", true)
			}
		}
	}
}

// 填充数据
func (inter *InterfaceDeclarationResult) addTypeReference(typeName string, location string, isExtend bool) {
	// 排除基本类型和已知的内置类型
	if utils.IsBasicType(typeName) {
		return
	}

	if ref, exists := inter.Reference[typeName]; exists {
		// 如果类型引用已存在，追加新的位置
		ref.Location = append(ref.Location, location)
		inter.Reference[typeName] = ref
	} else {
		// 如果类型引用不存在，创建新的引用
		inter.Reference[typeName] = TypeReference{
			Name:     typeName,
			Location: []string{location},
			IsExtend: isExtend,
		}
	}
}

// 分析接口的成员属性类型
func AnalyzeMember(member *ast.Node, interfaceName string) (string, string) {
	// 大多数情况下，成员是属性签名
	if ast.IsPropertySignatureDeclaration(member) {
		propSig := member.AsPropertySignatureDeclaration()
		// 提取属性名称
		var propName string
		if propSig.Name().Kind == ast.KindIdentifier {
			// 常规：interface Person { name: string; };
			propName = propSig.Name().Text()
		} else if propSig.Name().Kind == ast.KindStringLiteral || propSig.Name().Kind == ast.KindNumericLiteral {
			// 字符串/数字 字面量：interface Person { 0: string;  1: number; };
			propName = propSig.Name().Text()
		} else if propSig.Name().Kind == ast.KindComputedPropertyName {
			// 计算属性名 interface Person { ["sss111"]: string; }
			expr := propSig.Name().AsComputedPropertyName().Expression
			if ast.IsStringOrNumericLiteralLike(expr) {
				propName = expr.Text()
			}
		}
		location := fmt.Sprintf("%s.%s", interfaceName, propName)
		if propSig.Type != nil {
			return AnalyzeType(propSig.Type, location)
		}
	}
	return "", ""
}

// 递归分析类型节点。
// 根据类型节点的种类（如类型引用、数组类型、联合类型、交叉类型等）进行不同的处理。
// 如果类型是外部引用，则记录下来返回。
func AnalyzeType(typeNode *ast.Node, location string) (string, string) {
	if typeNode == nil {
		return "", ""
	}
	var typeNames []string
	var locations []string

	switch {
	// 处理类型引用
	case ast.IsTypeReferenceNode(typeNode):
		typeRef := typeNode.AsTypeReferenceNode()
		if ast.IsIdentifier(typeRef.TypeName) {
			typeName := typeRef.TypeName.AsIdentifier().Text
			// 排除基本类型
			if !utils.IsBasicType(typeName) {
				typeNames = append(typeNames, typeName)
				locations = append(locations, location)
			}

			// 分析类型参数 (泛型场景)
			if typeRef.TypeArguments != nil {
				for _, typeArg := range typeRef.TypeArguments.Nodes {
					argTypeName, argLocation := AnalyzeType(typeArg, location+"<>")
					if argTypeName != "" {
						typeNames = append(typeNames, argTypeName)
						locations = append(locations, argLocation)
					}
				}
			}
		} else if ast.IsQualifiedName(typeRef.TypeName) {
			// 处理 namespace.Type
			typeNames = append(typeNames, entityNameToString(typeRef.TypeName))
			locations = append(locations, location)
		}
	// 处理数组类型
	case typeNode.Kind == ast.KindArrayType:
		arrayType := typeNode.AsArrayTypeNode()
		if ast.IsTypeReferenceNode(arrayType.ElementType) {
			elemTypeRef := arrayType.ElementType.AsTypeReferenceNode()
			if ast.IsIdentifier(elemTypeRef.TypeName) {
				typeName := elemTypeRef.TypeName.AsIdentifier().Text
				if !utils.IsBasicType(typeName) {
					typeNames = append(typeNames, typeName)
					locations = append(locations, location)
				}
			}

		} else {
			// 递归处理数组元素类型
			memberTypeName, memberLocation := AnalyzeType(arrayType.ElementType, location)
			if memberTypeName != "" {
				typeNames = append(typeNames, memberTypeName)
				locations = append(locations, memberLocation)
			}
		}
	// 处理联合类型
	case typeNode.Kind == ast.KindUnionType:
		unionType := typeNode.AsUnionTypeNode()
		for _, unionMember := range unionType.Types.Nodes {
			memberTypeName, memberLocation := AnalyzeType(unionMember, location)
			if memberTypeName != "" {
				typeNames = append(typeNames, memberTypeName)
				locations = append(locations, memberLocation)
			}
		}
	// 处理交叉类型
	case typeNode.Kind == ast.KindIntersectionType:
		intersectionType := typeNode.AsIntersectionTypeNode()
		for _, intersectionMember := range intersectionType.Types.Nodes {
			memberTypeName, memberLocation := AnalyzeType(intersectionMember, location)
			if memberTypeName != "" {
				typeNames = append(typeNames, memberTypeName)
				locations = append(locations, memberLocation)
			}
		}
	// 处理元组类型
	case ast.IsTupleTypeNode(typeNode):
		tupleType := typeNode.AsTupleTypeNode()
		for i, elemType := range tupleType.Elements.Nodes {
			elemTypeName, elemLocation := AnalyzeType(elemType, fmt.Sprintf("%s[%d]", location, i))
			if elemTypeName != "" {
				typeNames = append(typeNames, elemTypeName)
				locations = append(locations, elemLocation)
			}
		}
	// 处理内联类型，持续递归调用 {a: {b: c:number}}
	case ast.IsTypeLiteralNode(typeNode):
		typeLiteral := typeNode.AsTypeLiteralNode()
		for _, member := range typeLiteral.Members.Nodes {
			memberTypeName, memberLocation := AnalyzeMember(member, location)
			if memberTypeName != "" {
				typeNames = append(typeNames, memberTypeName)
				locations = append(locations, memberLocation)
			}
		}
	}
	return strings.Join(typeNames, ","), strings.Join(locations, ",")
}

// typescript-go内部方法，从实体名称节点获取完整的字符串表示
func entityNameToString(name *ast.Node) string {
	switch name.Kind {
	case ast.KindThisKeyword:
		return "this"
	case ast.KindIdentifier, ast.KindPrivateIdentifier:
		return name.Text()
	case ast.KindQualifiedName:
		return entityNameToString(name.AsQualifiedName().Left) + "." + entityNameToString(name.AsQualifiedName().Right)
	case ast.KindPropertyAccessExpression:
		return entityNameToString(name.AsPropertyAccessExpression().Expression) + "." + entityNameToString(name.AsPropertyAccessExpression().Name())
	}
	return fmt.Sprintf("UnknownExpression(%s)", name.Kind)
}
