package tsmorphgo

import (
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// GetVariableName 获取变量声明节点的名称。
// - 对于 `const foo = 1`，它返回 "foo"。
// - 对于 `const { a, b } = c`，它返回 "{ a, b }"。
// 如果节点不是一个有效的变量声明，则返回空字符串和 false。
func GetVariableName(node Node) (string, bool) {
	if node.Kind != ast.KindVariableDeclaration {
		return "", false
	}
	decl := node.AsVariableDeclaration()
	if decl == nil || decl.Name() == nil {
		return "", false
	}
	return decl.Name().Text(), true
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
