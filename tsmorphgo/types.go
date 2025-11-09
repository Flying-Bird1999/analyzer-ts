package tsmorphgo

import (
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
)

// =============================================================================
// 简化的类型转换函数
// 配合 node_unified.go 中的统一 API 使用
// =============================================================================

// AsImportDeclaration 转换为导入声明结果
func AsImportDeclaration(node Node) (projectParser.ImportDeclarationResult, bool) {
	if node.sourceFile == nil || node.sourceFile.nodeResultMap == nil {
		return projectParser.ImportDeclarationResult{}, false
	}

	if result, ok := node.sourceFile.nodeResultMap[node.Node]; ok {
		if castedResult, ok := result.(projectParser.ImportDeclarationResult); ok {
			return castedResult, true
		}
	}
	return projectParser.ImportDeclarationResult{}, false
}

// AsVariableDeclaration 转换为变量声明结果
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

// AsFunctionDeclaration 转换为函数声明结果
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

// AsInterfaceDeclaration 转换为接口声明结果
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

// AsTypeAliasDeclaration 转换为类型别名声明结果
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

// AsEnumDeclaration 转换为枚举声明结果
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

// AsClassDeclaration 转换为类声明节点
func AsClassDeclaration(node Node) (*Node, bool) {
	if node.Kind == KindClassDeclaration {
		return &node, true
	}
	return nil, false
}

// AsMethodDeclaration 转换为方法声明节点
func AsMethodDeclaration(node Node) (*Node, bool) {
	if node.Kind == KindMethodDeclaration {
		return &node, true
	}
	return nil, false
}

// AsConstructor 转换为构造函数节点
func AsConstructor(node Node) (*Node, bool) {
	if node.Kind == KindConstructor {
		return &node, true
	}
	return nil, false
}

// AsGetAccessor 转换为 getter 访问器节点
func AsGetAccessor(node Node) (*Node, bool) {
	if node.Kind == KindGetAccessor {
		return &node, true
	}
	return nil, false
}

// AsSetAccessor 转换为 setter 访问器节点
func AsSetAccessor(node Node) (*Node, bool) {
	if node.Kind == KindSetAccessor {
		return &node, true
	}
	return nil, false
}

// AsTypeParameter 转换为类型参数节点
func AsTypeParameter(node Node) (*Node, bool) {
	if node.Kind == KindTypeParameter {
		return &node, true
	}
	return nil, false
}

// AsTypeReference 转换为类型引用节点
func AsTypeReference(node Node) (*Node, bool) {
	if node.Kind == KindTypeReference {
		return &node, true
	}
	return nil, false
}

// AsArrayLiteralExpression 转换为数组字面量表达式节点
func AsArrayLiteralExpression(node Node) (*Node, bool) {
	if node.Kind == KindArrayLiteralExpression {
		return &node, true
	}
	return nil, false
}

// AsTypeAssertionExpression 转换为类型断言表达式节点
func AsTypeAssertionExpression(node Node) (*Node, bool) {
	if node.Kind == KindTypeAssertionExpression {
		return &node, true
	}
	return nil, false
}