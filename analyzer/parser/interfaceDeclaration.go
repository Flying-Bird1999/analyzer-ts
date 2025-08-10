// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（interfaceDeclaration.go）专门负责处理和解析接口（Interface）声明及其复杂的类型依赖关系。
package parser

import (
	"fmt"
	"main/analyzer/utils"
	"strings"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// TypeReference 代表一个类型引用。
// 它记录了在接口中引用的其他类型（非基础类型）的信息。
type TypeReference struct {
	Identifier string   `json:"identifier"` // 被引用的类型名称，例如 `User`。
	Location   []string `json:"location"`   // 类型被引用的具体位置路径，例如 `School.student.name`。
	IsExtend   bool     `json:"isExtend"`   // 标记此引用是否来自 `extends` 子句。
}

// InterfaceDeclarationResult 存储一个完整的接口声明的解析结果。
type InterfaceDeclarationResult struct {
	Identifier     string                   `json:"identifier"` // 接口的名称。
	Raw            string                   `json:"raw"`        // 节点在源码中的原始文本。
	Reference      map[string]TypeReference `json:"reference"`  // 接口所依赖的其他类型的映射，以类型名作为 key。
	SourceLocation SourceLocation           `json:"sourceLocation"` // 节点在源码中的位置信息。
}

// NewInterfaceDeclarationResult 基于 AST 节点创建一个新的 InterfaceDeclarationResult 实例。
func NewInterfaceDeclarationResult(node *ast.Node, sourceCode string) *InterfaceDeclarationResult {
	raw := utils.GetNodeText(node, sourceCode)
	pos, end := node.Pos(), node.End()

	return &InterfaceDeclarationResult{
		Identifier: "",
		Raw:        raw,
		Reference:  make(map[string]TypeReference),
		SourceLocation: SourceLocation{
			Start: NodePosition{Line: pos, Column: 0},
			End:   NodePosition{Line: end, Column: 0},
		},
	}
}

// analyzeInterfaces 是分析接口声明的入口函数。
// 它负责提取接口名称，并分别调用函数处理继承关系和成员属性。
func (inter *InterfaceDeclarationResult) analyzeInterfaces(interfaceDecl *ast.InterfaceDeclaration) {
	interfaceName := interfaceDecl.Name().AsIdentifier().Text
	inter.Identifier = interfaceName

	// 分析接口的 `extends` 继承关系。
	inter.analyzeHeritageClause(interfaceDecl, interfaceName)

	// 遍历并分析接口的所有成员（属性、方法等）。
	if interfaceDecl.Members != nil {
		for _, member := range interfaceDecl.Members.Nodes {
			memberTypeName, memberLocation := AnalyzeMember(member, interfaceName)
			// 如果成员分析返回了有效的类型名和位置，则将其添加到引用中。
			if memberTypeName != "" && memberLocation != "" {
				// 一个成员可能依赖多个类型（例如联合类型 A | B），因此需要分割处理。
				for i, typeName := range strings.Split(memberTypeName, ",") {
					inter.addTypeReference(typeName, strings.Split(memberLocation, ",")[i], false)
				}
			}
		}
	}
}

// analyzeHeritageClause 分析接口的 `extends` 子句，提取所有被继承的类型。
func (inter *InterfaceDeclarationResult) analyzeHeritageClause(interfaceDecl *ast.InterfaceDeclaration, interfaceName string) {
	extendsElements := ast.GetExtendsHeritageClauseElements(interfaceDecl.AsNode())

	for _, node := range extendsElements {
		expression := node.Expression()
		// Case 1: 简单标识符继承, e.g., `extends MyInterface`
		if ast.IsIdentifier(expression) {
			name := expression.AsIdentifier().Text
			// 忽略 TypeScript 的内置工具类型（如 Omit, Pick），但仍会分析其泛型参数。
			if !(utils.IsUtilityType(name)) {
				inter.addTypeReference(name, "", true)
			}
		// Case 2: 带命名空间的继承, e.g., `extends MyNamespace.MyInterface`
		} else if ast.IsPropertyAccessExpression(expression) {
			name := entityNameToString(expression)
			inter.addTypeReference(name, "", true)
		}

		// 分析 `extends` 中的泛型参数, e.g., `extends MyGeneric<TypeA, TypeB>`
			if node.TypeArguments != nil {
			for _, typeArg := range node.TypeArguments() {
				name, _ := AnalyzeType(typeArg, "")
				inter.addTypeReference(name, "", true)
			}
		}
	}
}

// addTypeReference 是一个辅助函数，用于向结果的 Reference 映射中添加或更新类型引用。
func (inter *InterfaceDeclarationResult) addTypeReference(typeName string, location string, isExtend bool) {
	// 忽略基础类型（string, number, boolean 等）和已知的内置类型。
	if utils.IsBasicType(typeName) {
		return
	}

	// 忽略对接口自身的引用。
	if typeName == inter.Identifier {
		return
	}

	// 如果引用已存在，则追加新的位置信息。
	if ref, exists := inter.Reference[typeName]; exists {
		ref.Location = append(ref.Location, location)
		inter.Reference[typeName] = ref
	} else {
		// 否则，创建一个新的类型引用条目。
		inter.Reference[typeName] = TypeReference{
			Identifier: typeName,
			Location:   []string{location},
			IsExtend:   isExtend,
		}
	}
}

// AnalyzeMember 分析接口中的单个成员（通常是属性签名），并返回其引用的类型名称和位置。
func AnalyzeMember(member *ast.Node, interfaceName string) (string, string) {
	if ast.IsPropertySignatureDeclaration(member) {
		propSig := member.AsPropertySignatureDeclaration()
		var propName string
		// 确定属性名称，处理常规标识符、字符串/数字字面量和计算属性名。
		switch propSig.Name().Kind {
		case ast.KindIdentifier: // e.g., `name: string;`
			propName = propSig.Name().Text()
		case ast.KindStringLiteral, ast.KindNumericLiteral: // e.g., `'0': string;`
			propName = propSig.Name().Text()
		case ast.KindComputedPropertyName: // e.g., `[myVar]: string;`
			expr := propSig.Name().AsComputedPropertyName().Expression
			if ast.IsStringOrNumericLiteralLike(expr) {
				propName = expr.Text()
			}
		}
		// 构建属性的访问路径。
		location := fmt.Sprintf("%s.%s", interfaceName, propName)
		// 递归分析属性的类型。
		if propSig.Type != nil {
			return AnalyzeType(propSig.Type, location)
		}
	}
	return "", ""
}

// AnalyzeType 是一个核心的递归函数，用于深度分析类型节点，找出所有非基础类型的引用。
// 它能够处理各种复杂的 TypeScript 类型，如泛型、数组、联合/交叉类型、元组、内联对象、索引访问和映射类型。
func AnalyzeType(typeNode *ast.Node, location string) (string, string) {
	if typeNode == nil {
		return "", ""
	}
	var typeNames []string
	var locations []string

	switch {
	// Case: 类型引用, e.g., `User`, `MyNamespace.User`, `Promise<User>`
	case ast.IsTypeReferenceNode(typeNode):
		typeRef := typeNode.AsTypeReferenceNode()
		if ast.IsIdentifier(typeRef.TypeName) {
			typeName := typeRef.TypeName.AsIdentifier().Text
			if !utils.IsBasicType(typeName) {
				typeNames = append(typeNames, typeName)
				locations = append(locations, location)
			}
			// 递归分析泛型参数。
			if typeRef.TypeArguments != nil {
				for _, typeArg := range typeRef.TypeArguments.Nodes {
					argTypeName, argLocation := AnalyzeType(typeArg, location+"<>") // 使用 <> 标记泛型位置
					if argTypeName != "" {
						typeNames = append(typeNames, argTypeName)
						locations = append(locations, argLocation)
					}
				}
			}
		} else if ast.IsQualifiedName(typeRef.TypeName) { // e.g., `Namespace.Type`
			typeNames = append(typeNames, entityNameToString(typeRef.TypeName))
			locations = append(locations, location)
		}
	// Case: 数组类型, e.g., `User[]`
	case typeNode.Kind == ast.KindArrayType:
		arrayType := typeNode.AsArrayTypeNode()
		// 递归分析数组成员的类型。
		memberTypeName, memberLocation := AnalyzeType(arrayType.ElementType, location)
		if memberTypeName != "" {
			typeNames = append(typeNames, memberTypeName)
			locations = append(locations, memberLocation)
		}
	// Case: 联合类型, e.g., `string | User | null`
	case typeNode.Kind == ast.KindUnionType:
		unionType := typeNode.AsUnionTypeNode()
		for _, unionMember := range unionType.Types.Nodes {
			memberTypeName, memberLocation := AnalyzeType(unionMember, location)
			if memberTypeName != "" {
				typeNames = append(typeNames, memberTypeName)
				locations = append(locations, memberLocation)
			}
		}
	// Case: 交叉类型, e.g., `User & { id: number }`
	case typeNode.Kind == ast.KindIntersectionType:
		intersectionType := typeNode.AsIntersectionTypeNode()
		for _, intersectionMember := range intersectionType.Types.Nodes {
			memberTypeName, memberLocation := AnalyzeType(intersectionMember, location)
			if memberTypeName != "" {
				typeNames = append(typeNames, memberTypeName)
				locations = append(locations, memberLocation)
			}
		}
	// Case: 元组类型, e.g., `[string, number, User]`
	case ast.IsTupleTypeNode(typeNode):
		tupleType := typeNode.AsTupleTypeNode()
		for i, elemType := range tupleType.Elements.Nodes {
			elemTypeName, elemLocation := AnalyzeType(elemType, fmt.Sprintf("%s[%d]", location, i))
			if elemTypeName != "" {
				typeNames = append(typeNames, elemTypeName)
				locations = append(locations, elemLocation)
			}
		}
	// Case: 内联类型/对象字面量类型, e.g., `{ name: string; data: User }`
	case ast.IsTypeLiteralNode(typeNode):
		typeLiteral := typeNode.AsTypeLiteralNode()
		for _, member := range typeLiteral.Members.Nodes {
			// location 作为父级路径传入，AnalyzeMember 会拼接自己的属性名。
			memberTypeName, memberLocation := AnalyzeMember(member, location)
			if memberTypeName != "" {
				typeNames = append(typeNames, memberTypeName)
				locations = append(locations, memberLocation)
			}
		}
	// Case: 索引访问类型, e.g., `Translations["name"]`
	case typeNode.Kind == ast.KindIndexedAccessType:
		indexedAccessType := typeNode.AsIndexedAccessTypeNode()
		// 分析被索引的对象类型。
		elemTypeName, elemLocation := AnalyzeType(indexedAccessType.ObjectType, location)
		if elemTypeName != "" {
			typeNames = append(typeNames, elemTypeName)
			locations = append(locations, elemLocation)
		}
	// Case: 映射类型, e.g., `[key in ImagesType]: ImagesAttribute`
	case typeNode.Kind == ast.KindMappedType:
		mappedTypeNode := typeNode.AsMappedTypeNode()
		if mappedTypeNode.TypeParameter != nil {
			typeParam := mappedTypeNode.TypeParameter.AsTypeParameter()
			// 分析 `in` 后面的约束类型。
			if typeParam.Constraint != nil {
				elemTypeName, elemLocation := AnalyzeType(typeParam.Constraint, "")
				typeNames = append(typeNames, elemTypeName)
				locations = append(locations, elemLocation)
			}
			// 分析映射的值类型。
			if mappedTypeNode.Type != nil {
				elemTypeName, elemLocation := AnalyzeType(mappedTypeNode.Type, "")
				typeNames = append(typeNames, elemTypeName)
				locations = append(locations, elemLocation)
			}
		}
	}
	// 将收集到的所有类型和位置用逗号连接成字符串返回。
	return strings.Join(typeNames, ","), strings.Join(locations, ",")
}

// entityNameToString 将一个实体名称节点（可能是一个标识符或一个属性访问表达式）转换为一个点分隔的字符串。
// 例如，`a.b.c` 节点会被转换为字符串 "a.b.c"。
func entityNameToString(name *ast.Node) string {
	switch name.Kind {
	case ast.KindThisKeyword:
		return "this"
	case ast.KindIdentifier, ast.KindPrivateIdentifier:
		return name.Text()
	case ast.KindQualifiedName: // e.g., `A.B`
		return entityNameToString(name.AsQualifiedName().Left) + "." + entityNameToString(name.AsQualifiedName().Right)
	case ast.KindPropertyAccessExpression: // e.g., `a.b`
		return entityNameToString(name.AsPropertyAccessExpression().Expression) + "." + entityNameToString(name.AsPropertyAccessExpression().Name())
	}
	return fmt.Sprintf("UnknownExpression(%s)", name.Kind)
}
