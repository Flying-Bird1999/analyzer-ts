package tsmorphgo

import (
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// AsImportDeclaration 尝试将一个通用节点 (Node) 转换为一个具体的导入声明结果。
// 这种方式提供了类型安全的、针对特定节点类型的数据访问。
// 如果转换成功，返回具体的导入声明结果和 true；否则返回零值和 false。
func AsImportDeclaration(node Node) (projectParser.ImportDeclarationResult, bool) {
	if node.sourceFile == nil || node.sourceFile.nodeResultMap == nil {
		return projectParser.ImportDeclarationResult{}, false
	}

	// 从预先构建的 map 中查找当前 ast.Node 对应的解析结果
	if result, ok := node.sourceFile.nodeResultMap[node.Node]; ok {
		// 使用类型断言检查结果是否是我们期望的类型
		if castedResult, ok := result.(projectParser.ImportDeclarationResult); ok {
			return castedResult, true
		}
	}

	return projectParser.ImportDeclarationResult{}, false
}

// AsVariableDeclaration 尝试将一个通用节点 (Node) 转换为一个具体的变量声明结果。
func AsVariableDeclaration(node Node) (parser.VariableDeclaration, bool) {
	if node.sourceFile == nil || node.sourceFile.nodeResultMap == nil {
		return parser.VariableDeclaration{}, false
	}
	if result, ok := node.sourceFile.nodeResultMap[node.Node]; ok {
		if castedResult, ok := result.(parser.VariableDeclaration); ok {
			return castedResult, true
		}
	}
	return parser.VariableDeclaration{}, false
}

// AsFunctionDeclaration 尝试将一个通用节点 (Node) 转换为一个具体的函数声明结果。
func AsFunctionDeclaration(node Node) (parser.FunctionDeclarationResult, bool) {
	if node.sourceFile == nil || node.sourceFile.nodeResultMap == nil {
		return parser.FunctionDeclarationResult{}, false
	}
	if result, ok := node.sourceFile.nodeResultMap[node.Node]; ok {
		if castedResult, ok := result.(parser.FunctionDeclarationResult); ok {
			return castedResult, true
		}
	}
	return parser.FunctionDeclarationResult{}, false
}

// AsInterfaceDeclaration 尝试将一个通用节点 (Node) 转换为一个具体的接口声明结果。
func AsInterfaceDeclaration(node Node) (parser.InterfaceDeclarationResult, bool) {
	if node.sourceFile == nil || node.sourceFile.nodeResultMap == nil {
		return parser.InterfaceDeclarationResult{}, false
	}
	if result, ok := node.sourceFile.nodeResultMap[node.Node]; ok {
		if castedResult, ok := result.(parser.InterfaceDeclarationResult); ok {
			return castedResult, true
		}
	}
	return parser.InterfaceDeclarationResult{}, false
}

// IsIdentifier 检查一个节点是否是标识符 (Identifier)。
func IsIdentifier(node Node) bool {
	return node.Kind == ast.KindIdentifier
}

// IsCallExpression 检查一个节点是否是函数调用表达式 (CallExpression)。
func IsCallExpression(node Node) bool {
	return node.Kind == ast.KindCallExpression
}

// IsPropertyAccessExpression 检查一个节点是否是属性访问表达式 (PropertyAccessExpression)。
func IsPropertyAccessExpression(node Node) bool {
	return node.Kind == ast.KindPropertyAccessExpression
}

// IsPropertyAssignment 检查一个节点是否是属性分配 (PropertyAssignment)。
func IsPropertyAssignment(node Node) bool {
	return node.Kind == ast.KindPropertyAssignment
}

// IsVariableDeclaration 检查一个节点是否是变量声明 (VariableDeclaration)。
// 注意：这对应于 `const a = 1` 中的 `a = 1` 部分，而不是整个语句。
func IsVariableDeclaration(node Node) bool {
	return node.Kind == ast.KindVariableDeclaration
}

// IsFunctionDeclaration 检查一个节点是否是函数声明 (FunctionDeclaration)。
func IsFunctionDeclaration(node Node) bool {
	return node.Kind == ast.KindFunctionDeclaration
}

// IsObjectLiteralExpression 检查一个节点是否是对象字面量表达式 (ObjectLiteralExpression)。
func IsObjectLiteralExpression(node Node) bool {
	return node.Kind == ast.KindObjectLiteralExpression
}

// IsInterfaceDeclaration 检查一个节点是否是接口声明 (InterfaceDeclaration)。
func IsInterfaceDeclaration(node Node) bool {
	return node.Kind == ast.KindInterfaceDeclaration
}

// IsBinaryExpression 检查一个节点是否是二元表达式 (BinaryExpression)。
func IsBinaryExpression(node Node) bool {
	return node.Kind == ast.KindBinaryExpression
}

// IsTypeAliasDeclaration 检查一个节点是否是类型别名声明 (TypeAliasDeclaration)。
func IsTypeAliasDeclaration(node Node) bool {
	return node.Kind == ast.KindTypeAliasDeclaration
}

// AsTypeAliasDeclaration 尝试将一个通用节点 (Node) 转换为一个具体的类型别名声明结果。
func AsTypeAliasDeclaration(node Node) (parser.TypeDeclarationResult, bool) {
	if node.sourceFile == nil || node.sourceFile.nodeResultMap == nil {
		return parser.TypeDeclarationResult{}, false
	}
	if result, ok := node.sourceFile.nodeResultMap[node.Node]; ok {
		if castedResult, ok := result.(parser.TypeDeclarationResult); ok {
			return castedResult, true
		}
	}
	return parser.TypeDeclarationResult{}, false
}

// IsEnumDeclaration 检查一个节点是否是枚举声明 (EnumDeclaration)。
func IsEnumDeclaration(node Node) bool {
	return node.Kind == ast.KindEnumDeclaration
}

// AsEnumDeclaration 尝试将一个通用节点 (Node) 转换为一个具体的枚举声明结果。
func AsEnumDeclaration(node Node) (parser.EnumDeclarationResult, bool) {
	if node.sourceFile == nil || node.sourceFile.nodeResultMap == nil {
		return parser.EnumDeclarationResult{}, false
	}
	if result, ok := node.sourceFile.nodeResultMap[node.Node]; ok {
		if castedResult, ok := result.(parser.EnumDeclarationResult); ok {
			return castedResult, true
		}
	}
	return parser.EnumDeclarationResult{}, false
}

// ... 后续可以按照此模式添加其他 IsXXX 和 AsXXX 函数 ...
