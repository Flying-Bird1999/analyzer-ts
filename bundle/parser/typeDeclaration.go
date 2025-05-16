package parser

import (
	"fmt"
	"main/bundle/utils"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// 解析 type 声明，递归去查找 type 里边的类型
// - 如果有引用外部类型的就找出来
// - case1: type Name3 = LinearModel | Person;
// - case2: type Name = { name: string; age: LinearModel; };

type TypeDeclarationResult struct {
	Name      string // 名称
	Raw       string // 源码
	Reference map[string]TypeReference
}

func NewTypeDeclarationResult(node *ast.Node, sourceCode string) *TypeDeclarationResult {
	raw := utils.GetNodeText(node.AsNode(), sourceCode)

	return &TypeDeclarationResult{
		Name:      "",
		Raw:       raw,
		Reference: make(map[string]TypeReference),
	}
}

// 分析接口的主要结构，包括：
// 1. 接口名称。
// 2. 类型成员（通过 analyzeMember）。
func (inter *TypeDeclarationResult) analyzeTypeDecl(typeDecl *ast.TypeAliasDeclaration) {
	typeName := typeDecl.Name().Text()
	inter.Name = typeName

	fmt.Printf("Type.Kind: %s\n", typeDecl.Type.Kind)

	// 对象字面量类型，分析内部类成员
	// type Name = { name: string; age: LinearModel; };
	if typeDecl.Type.Kind == ast.KindTypeLiteral {
		if typeDecl.Type.Members() != nil {
			for _, member := range typeDecl.Type.Members() {
				inter.analyzeMember(member, typeName)
			}
		}
	} else {
		// type Name3 = LinearModel | Person;
		inter.analyzeType(typeDecl.Type, typeName)
	}
}

// 分析接口的成员属性类型
func (inter *TypeDeclarationResult) analyzeMember(member *ast.Node, typeName string) {
	if ast.IsPropertySignatureDeclaration(member) {
		propSig := member.AsPropertySignatureDeclaration()
		propName := propSig.Name().AsIdentifier().Text
		location := fmt.Sprintf("%s.%s", typeName, propName)
		if propSig.Type != nil {
			inter.analyzeType(propSig.Type, location)
		}
	}
}

// 递归分析类型节点。
// 根据类型节点的种类（如类型引用、数组类型、联合类型、交叉类型等）进行不同的处理。
// 如果类型是外部引用，则调用 inter.addTypeReference 记录。
func (inter *TypeDeclarationResult) analyzeType(typeNode *ast.Node, location string) {
	if typeNode == nil {
		return
	}

	switch {
	// 处理类型引用
	case ast.IsTypeReferenceNode(typeNode):
		typeRef := typeNode.AsTypeReferenceNode()
		if ast.IsIdentifier(typeRef.TypeName) {
			typeName := typeRef.TypeName.AsIdentifier().Text
			// 排除基本类型
			if !utils.IsBasicType(typeName) {
				inter.addTypeReference(typeName, location, false)
			}

			// 分析类型参数 (泛型场景)
			if typeRef.TypeArguments != nil {
				for _, typeArg := range typeRef.TypeArguments.Nodes {
					inter.analyzeType(typeArg, location+"<>")
				}
			}
		} else if ast.IsQualifiedName(typeRef.TypeName) {
			// 处理 namespace.Type
			qualifiedName := typeRef.TypeName.AsQualifiedName()
			right := qualifiedName.Right.AsIdentifier().Text
			left := ""
			if ast.IsIdentifier(qualifiedName.Left) {
				left = qualifiedName.Left.AsIdentifier().Text
			}
			inter.addTypeReference(left+"."+right, location, false)
		}
	// 处理数组类型
	case typeNode.Kind == ast.KindArrayType:
		arrayType := typeNode.AsArrayTypeNode()
		if ast.IsTypeReferenceNode(arrayType.ElementType) {
			elemTypeRef := arrayType.ElementType.AsTypeReferenceNode()
			if ast.IsIdentifier(elemTypeRef.TypeName) {
				typeName := elemTypeRef.TypeName.AsIdentifier().Text
				if !utils.IsBasicType(typeName) {
					inter.addTypeReference(typeName, location, false)
				}
			}

		} else {
			// 递归处理数组元素类型
			inter.analyzeType(arrayType.ElementType, location)
		}
	// 处理联合类型
	case typeNode.Kind == ast.KindUnionType:
		unionType := typeNode.AsUnionTypeNode()
		for _, unionMember := range unionType.Types.Nodes {
			inter.analyzeType(unionMember, location)
		}
	// 处理交叉类型
	case typeNode.Kind == ast.KindIntersectionType:
		intersectionType := typeNode.AsIntersectionTypeNode()
		for _, intersectionMember := range intersectionType.Types.Nodes {
			inter.analyzeType(intersectionMember, location)
		}
	// 处理元组类型
	case ast.IsTupleTypeNode(typeNode):

		tupleType := typeNode.AsTupleTypeNode()
		for i, elemType := range tupleType.Elements.Nodes {
			inter.analyzeType(elemType, fmt.Sprintf("%s[%d]", location, i))
		}
	// 处理内联类型，持续递归调用 {a: {b: c:number}}
	case ast.IsTypeLiteralNode(typeNode):
		typeLiteral := typeNode.AsTypeLiteralNode()
		for _, member := range typeLiteral.Members.Nodes {
			inter.analyzeMember(member, location)
		}
	}
}

// 填充数据
func (inter *TypeDeclarationResult) addTypeReference(typeName string, location string, isExtend bool) {
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
