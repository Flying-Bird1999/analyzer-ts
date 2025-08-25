// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（typeAnalyzer.go）包含了用于分析复杂类型节点的共享逻辑，
// 被 interfaceDeclaration.go 和 typeDeclaration.go 等文件使用，是类型依赖分析的核心。
package parser

import (
	"fmt"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// TypeAnalysisResult 封装了从单个类型节点中分析出的一个类型引用的结果。
type TypeAnalysisResult struct {
	TypeName string // 分析出的类型名称，例如 "User" 或 "Namespace.Type"。
	Location string // 该类型在父级结构中的位置路径，例如 "MyInterface.user" 或 "MyType.items[]"。
}

// AnalyzeMember 分析接口或类型字面量中的单个成员（通常是属性签名），并返回其引用的类型信息。
// parentLocation 是父级的访问路径，例如接口名 "MyInterface" 或类型字面量的路径 "MyType.config"。
func AnalyzeMember(member *ast.Node, parentLocation string) []TypeAnalysisResult {
	// 目前只处理属性签名（PropertySignature），未来可以扩展到方法签名等。
	if ast.IsPropertySignatureDeclaration(member) {
		propSig := member.AsPropertySignatureDeclaration()
		var propName string
		// 确定属性名称，需要处理常规标识符、字符串/数字字面量和计算属性名。
		switch propSig.Name().Kind {
		case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral:
			// 例如: `name: string;` 或 `'0': number;`
			propName = propSig.Name().Text()
		case ast.KindComputedPropertyName:
			// 例如: `[myVar]: string;`
			expr := propSig.Name().AsComputedPropertyName().Expression
			if ast.IsStringOrNumericLiteralLike(expr) {
				propName = expr.Text()
			}
		}
		// 构建此属性的完整访问路径。
		location := fmt.Sprintf("%s.%s", parentLocation, propName)
		// 如果属性有明确的类型注解，则递归分析该类型。
		if propSig.Type != nil {
			return AnalyzeType(propSig.Type, location)
		}
	}
	return nil
}

// AnalyzeType 是一个核心的递归函数，用于深度分析给定的类型节点，找出所有非基础类型的引用。
// 它能够处理各种复杂的 TypeScript 类型，如泛型、数组、联合/交叉类型、元组、内联对象、索引访问和映射类型。
// location 参数代表当前分析的类型节点在最终结构中的位置。
func AnalyzeType(typeNode *ast.Node, location string) []TypeAnalysisResult {
	if typeNode == nil {
		return nil
	}
	var results []TypeAnalysisResult

	switch {
	// Case: 类型引用，例如 `User`, `MyNamespace.User`, `Promise<User>`
	case ast.IsTypeReferenceNode(typeNode):
		typeRef := typeNode.AsTypeReferenceNode()
		// 处理简单标识符或带命名空间的限定名
		if ast.IsIdentifier(typeRef.TypeName) {
			typeName := typeRef.TypeName.AsIdentifier().Text
			// 过滤掉 string, number 等基础类型
			if !utils.IsBasicType(typeName) {
				results = append(results, TypeAnalysisResult{TypeName: typeName, Location: location})
			}
			// 递归分析泛型参数，例如 `<User, number>`
			if typeRef.TypeArguments != nil {
				for _, typeArg := range typeRef.TypeArguments.Nodes {
					// 为泛型参数的位置添加特殊标记 "<>"
					results = append(results, AnalyzeType(typeArg, location+"<>")...)
				}
			}
		} else if ast.IsQualifiedName(typeRef.TypeName) { // 例如 `Namespace.Type`
			results = append(results, TypeAnalysisResult{TypeName: entityNameToString(typeRef.TypeName), Location: location})
		}
	// Case: 数组类型，例如 `User[]`
	case typeNode.Kind == ast.KindArrayType:
		arrayType := typeNode.AsArrayTypeNode()
		// 递归分析数组成员的类型
		results = append(results, AnalyzeType(arrayType.ElementType, location)...)
	// Case: 联合类型，例如 `string | User | null`
	case typeNode.Kind == ast.KindUnionType:
		unionType := typeNode.AsUnionTypeNode()
		for _, unionMember := range unionType.Types.Nodes {
			results = append(results, AnalyzeType(unionMember, location)...)
		}
	// Case: 交叉类型，例如 `User & { id: number }`
	case typeNode.Kind == ast.KindIntersectionType:
		intersectionType := typeNode.AsIntersectionTypeNode()
		for _, intersectionMember := range intersectionType.Types.Nodes {
			results = append(results, AnalyzeType(intersectionMember, location)...)
		}
	// Case: 元组类型，例如 `[string, number, User]`
	case ast.IsTupleTypeNode(typeNode):
		tupleType := typeNode.AsTupleTypeNode()
		for i, elemType := range tupleType.Elements.Nodes {
			// 为元组成员的位置添加索引标记 "[i]"
			elemLocation := fmt.Sprintf("%s[%d]", location, i)
			results = append(results, AnalyzeType(elemType, elemLocation)...)
		}
	// Case: 内联类型/对象字面量类型，例如 `{ name: string; data: User }`
	case ast.IsTypeLiteralNode(typeNode):
		typeLiteral := typeNode.AsTypeLiteralNode()
		for _, member := range typeLiteral.Members.Nodes {
			// location 作为父级路径传入，AnalyzeMember 会在内部拼接自己的属性名
			results = append(results, AnalyzeMember(member, location)...)
		}
	// Case: 索引访问类型，例如 `Translations["name"]`
	case typeNode.Kind == ast.KindIndexedAccessType:
		indexedAccessType := typeNode.AsIndexedAccessTypeNode()
		// 递归分析被索引的对象类型
		results = append(results, AnalyzeType(indexedAccessType.ObjectType, location)...)
	// Case: 映射类型，例如 `[key in ImagesType]: ImagesAttribute`
	case typeNode.Kind == ast.KindMappedType:
		mappedTypeNode := typeNode.AsMappedTypeNode()
		if mappedTypeNode.TypeParameter != nil {
			typeParam := mappedTypeNode.TypeParameter.AsTypeParameter()
			// 分析 `in` 后面的约束类型
			if typeParam.Constraint != nil {
				results = append(results, AnalyzeType(typeParam.Constraint, "")...)
			}
			// 分析映射的值类型
			if mappedTypeNode.Type != nil {
				results = append(results, AnalyzeType(mappedTypeNode.Type, "")...)
			}
		}
	}
	return results
}

// entityNameToString 将一个实体名称节点（可能是一个标识符或一个属性访问表达式）递归地转换为一个点分隔的字符串。
// 例如，`a.b.c` 节点会被转换为字符串 "a.b.c"。
func entityNameToString(name *ast.Node) string {
	switch name.Kind {
	case ast.KindThisKeyword:
		return "this"
	case ast.KindIdentifier, ast.KindPrivateIdentifier:
		return name.Text()
	case ast.KindQualifiedName: // 例如 `A.B`
		return entityNameToString(name.AsQualifiedName().Left) + "." + entityNameToString(name.AsQualifiedName().Right)
	case ast.KindPropertyAccessExpression: // 例如 `a.b`
		return entityNameToString(name.AsPropertyAccessExpression().Expression) + "." + entityNameToString(name.AsPropertyAccessExpression().Name())
	}
	return fmt.Sprintf("UnknownExpression(%s)", name.Kind)
}
