package tsmorphgo

import (
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// GetVariableName 获取变量声明节点的名称。
// - 对于 `const foo = 1`，它返回 "foo"。
// - 对于 `const { a, b } = c`，它返回 "destructured pattern"。
// 如果节点不是一个有效的变量声明，则返回空字符串和 false。
func GetVariableName(node Node) (string, bool) {
	if node.Kind != ast.KindVariableDeclaration {
		return "", false
	}
	decl := node.AsVariableDeclaration()
	if decl == nil || decl.Name() == nil {
		return "", false
	}

	// Check if it's an identifier (simple variable name)
	if decl.Name().Kind == ast.KindIdentifier {
		return decl.Name().Text(), true
	}

	// Check if it's a binding pattern (destructuring)
	if decl.Name().Kind == ast.KindObjectBindingPattern ||
		decl.Name().Kind == ast.KindArrayBindingPattern {
		return "destructured pattern", true
	}

	// For other cases, try to get text but catch any panics
	defer func() {
		if r := recover(); r != nil {
			// Silently recover and return false
		}
	}()

	text := decl.Name().Text()
	if text != "" {
		return text, true
	}

	return "", false
}

// GetVariableDeclarationNameNode 获取变量声明节点的名称节点 (NameNode)。
// 这对于需要进一步分析名称（例如，它是一个标识符还是一个解构模式）的场景很有用。
// 如果节点不是一个有效的变量声明，则返回 nil 和 false。
func GetVariableDeclarationNameNode(node Node) (*Node, bool) {
	if node.Kind != ast.KindVariableDeclaration {
		return nil, false
	}
	decl := node.AsVariableDeclaration()
	if decl == nil || decl.Name() == nil {
		return nil, false
	}
	nameNode := &Node{
		Node:       decl.Name(),
		sourceFile: node.sourceFile,
	}
	return nameNode, true
}

// GetFunctionDeclarationNameNode 获取函数声明的名称节点。
// 对于匿名函数，返回 nil 和 false。
func GetFunctionDeclarationNameNode(node Node) (*Node, bool) {
	if node.Kind != ast.KindFunctionDeclaration {
		return nil, false
	}
	fnDecl := node.AsFunctionDeclaration()
	if fnDecl == nil || fnDecl.Name() == nil {
		return nil, false
	}
	return &Node{
		Node:       fnDecl.Name(),
		sourceFile: node.sourceFile,
	}, true
}

// GetImportSpecifierAliasNode 获取导入说明符中的别名节点。
// 例如，对于 `import { foo as bar } from './mod'`，它将返回 `bar` 节点。
// 如果没有别名，则返回 nil 和 false。
func GetImportSpecifierAliasNode(node Node) (*Node, bool) {
	if node.Kind != ast.KindImportSpecifier {
		return nil, false
	}
	spec := node.AsImportSpecifier()
	// 在 typescript-go 的 AST 中，如果存在别名，`PropertyName` 是原始名，`Name` 是别名。
	if spec == nil || spec.PropertyName == nil {
		return nil, false
	}
	return &Node{
		Node:       spec.Name(),
		sourceFile: node.sourceFile,
	}, true
}

// IsClassDeclaration 检查节点是否是类声明。
func IsClassDeclaration(node Node) bool {
	return node.Kind == ast.KindClassDeclaration
}

// IsSetAccessor 检查节点是否是setter访问器。
func IsSetAccessor(node Node) bool {
	return node.Kind == ast.KindSetAccessor
}

// IsGetAccessor 检查节点是否是getter访问器。
func IsGetAccessor(node Node) bool {
	return node.Kind == ast.KindGetAccessor
}

// 节点级别的类型检查函数

// IsConstructor 检查节点是否是构造函数。
func IsConstructor(node Node) bool {
	return node.Kind == ast.KindConstructor
}

// IsAccessor 检查节点是否是访问器（getter/setter）。
func IsAccessor(node Node) bool {
	return node.Kind == ast.KindGetAccessor || node.Kind == ast.KindSetAccessor
}

// IsTypeParameter 检查节点是否是类型参数。
func IsTypeParameter(node Node) bool {
	return node.Kind == ast.KindTypeParameter
}

// IsTypeReference 检查节点是否是类型引用。
func IsTypeReference(node Node) bool {
	return node.Kind == ast.KindTypeReference
}

// IsArrayLiteralExpression 检查节点是否是数组字面量表达式。
func IsArrayLiteralExpression(node Node) bool {
	return node.Kind == ast.KindArrayLiteralExpression
}

// IsTypeAssertionExpression 检查节点是否是类型断言表达式。
func IsTypeAssertionExpression(node Node) bool {
	return node.Kind == ast.KindTypeAssertionExpression
}

// IsConstructorDeclaration 检查节点是否是构造函数声明。
func IsConstructorDeclaration(node Node) bool {
	return node.Kind == ast.KindConstructor
}

// IsMethodDeclaration 检查节点是否是方法声明。
func IsMethodDeclaration(node Node) bool {
	return node.Kind == ast.KindMethodDeclaration
}

// IsGetAccessorDeclaration 检查节点是否是getter访问器声明。
func IsGetAccessorDeclaration(node Node) bool {
	return node.Kind == ast.KindGetAccessor
}

// IsSetAccessorDeclaration 检查节点是否是setter访问器声明。
func IsSetAccessorDeclaration(node Node) bool {
	return node.Kind == ast.KindSetAccessor
}

// IsTypeAliasDeclaration 检查节点是否是类型别名声明。
func IsTypeAliasDeclaration(node Node) bool {
	return node.Kind == ast.KindTypeAliasDeclaration
}

// IsThisExpression 检查节点是否是this表达式。
func IsThisExpression(node Node) bool {
	return node.Kind == ast.KindThisKeyword
}

// IsSuperExpression 检查节点是否是super表达式。
func IsSuperExpression(node Node) bool {
	return node.Kind == ast.KindSuperKeyword
}

// IsTemplateExpression 检查节点是否是模板表达式。
func IsTemplateExpression(node Node) bool {
	return node.Kind == ast.KindTemplateExpression
}

// IsSpreadElement 检查节点是否是展开元素。
func IsSpreadElement(node Node) bool {
	return node.Kind == ast.KindSpreadElement
}

// IsYieldExpression 检查节点是否是yield表达式。
func IsYieldExpression(node Node) bool {
	return node.Kind == ast.KindYieldExpression
}

// IsAwaitExpression 检查节点是否是await表达式。
func IsAwaitExpression(node Node) bool {
	return node.Kind == ast.KindAwaitExpression
}
