package tsmorphgo

import (
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// GetCallExpressionExpression 获取函数调用表达式中被调用的部分。
// 例如，对于 `foo.bar()`，它返回代表 `foo.bar` 的节点。
func GetCallExpressionExpression(node Node) (*Node, bool) {
	if node.Kind != ast.KindCallExpression {
		return nil, false
	}
	callExpr := node.AsCallExpression()
	if callExpr == nil || callExpr.Expression == nil {
		return nil, false
	}
	return &Node{
		Node:       callExpr.Expression,
		sourceFile: node.sourceFile,
	}, true
}

// GetPropertyAccessName 获取属性访问表达式的属性名称。
// 例如，对于 `foo.bar`，它返回字符串 "bar"。
func GetPropertyAccessName(node Node) (string, bool) {
	if node.Kind != ast.KindPropertyAccessExpression {
		return "", false
	}
	propAccess := node.AsPropertyAccessExpression()
	if propAccess == nil || propAccess.Name() == nil {
		return "", false
	}
	return propAccess.Name().Text(), true
}

// GetPropertyAccessExpression 获取属性访问表达式中被访问的对象部分。
// 例如，对于 `foo.bar`，它返回代表 `foo` 的节点。
func GetPropertyAccessExpression(node Node) (*Node, bool) {
	if node.Kind != ast.KindPropertyAccessExpression {
		return nil, false
	}
	propAccess := node.AsPropertyAccessExpression()
	if propAccess == nil || propAccess.Expression == nil {
		return nil, false
	}
	return &Node{
		Node:       propAccess.Expression,
		sourceFile: node.sourceFile,
	}, true
}

// GetBinaryExpressionLeft 获取二元表达式的左操作数。
func GetBinaryExpressionLeft(node Node) (*Node, bool) {
	if node.Kind != ast.KindBinaryExpression {
		return nil, false
	}
	binaryExpr := node.AsBinaryExpression()
	if binaryExpr == nil || binaryExpr.Left == nil {
		return nil, false
	}
	return &Node{
		Node:       binaryExpr.Left,
		sourceFile: node.sourceFile,
	}, true
}

// GetBinaryExpressionRight 获取二元表达式的右操作数。
func GetBinaryExpressionRight(node Node) (*Node, bool) {
	if node.Kind != ast.KindBinaryExpression {
		return nil, false
	}
	binaryExpr := node.AsBinaryExpression()
	if binaryExpr == nil || binaryExpr.Right == nil {
		return nil, false
	}
	return &Node{
		Node:       binaryExpr.Right,
		sourceFile: node.sourceFile,
	}, true
}

// GetBinaryExpressionOperatorToken 获取二元表达式的操作符节点。
func GetBinaryExpressionOperatorToken(node Node) (*Node, bool) {
	if node.Kind != ast.KindBinaryExpression {
		return nil, false
	}
	binaryExpr := node.AsBinaryExpression()
	if binaryExpr == nil || binaryExpr.OperatorToken == nil {
		return nil, false
	}
	return &Node{
		Node:       binaryExpr.OperatorToken,
		sourceFile: node.sourceFile,
	}, true
}
