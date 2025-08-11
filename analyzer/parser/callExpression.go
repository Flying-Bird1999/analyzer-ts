// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（callExpression.go）专门负责处理和解析函数/方法调用表达式。
package parser

import (
	"main/analyzer/utils"
	"strings"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// Argument 代表函数调用中的一个参数。
type Argument struct {
	Type string `json:"type"` // 参数的类型, e.g., "string", "number", "identifier", "object", "function", "array", "boolean"
	Text string `json:"text"` // 参数在源码中的原始文本。
}

// CallExpression 代表一个函数或方法调用表达式的解析结果。
type CallExpression struct {
	CallChain      []string       `json:"callChain"`          // 调用的完整路径，例如 `myObj.methods.myMethod` 会被解析为 `["myObj", "methods", "myMethod"]`。
	Arguments      []Argument     `json:"arguments"`          // 调用时传递的参数列表。
	Type           string         `json:"type"`               // 调用的类型。`call` 表示普通函数调用，`member` 表示对象成员方法调用。
	Raw            string         `json:"raw,omitempty"`      // 节点在源码中的原始文本。
	SourceLocation SourceLocation `json:"sourceLocation"`     // 节点在源码中的位置信息。
}

// NewCallExpression 基于 AST 节点创建一个新的 CallExpression 实例。
// 它初始化了表达式的源码位置、原始文本和默认类型。
func NewCallExpression(node *ast.CallExpression, sourceCode string) *CallExpression {
	pos, end := node.Pos(), node.End()
	return &CallExpression{
		SourceLocation: SourceLocation{
			Start: NodePosition{Line: pos, Column: 0},
			End:   NodePosition{Line: end, Column: 0},
		},
		Raw:  utils.GetNodeText(node.AsNode(), sourceCode),
		Type: "call", // 默认为 'call'，在后续分析中可能会被修改为 'member'。
	}
}

// reconstructExpression 从表达式节点递归地构建一个调用链。
func reconstructExpression(node *ast.Node, sourceCode string) []string {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return []string{node.AsIdentifier().Text}
	case ast.KindPropertyAccessExpression:
		propAccess := node.AsPropertyAccessExpression()
		// 递归地构建左侧部分的调用链
		left := reconstructExpression(propAccess.Expression, sourceCode)
		// 将右侧的属性名追加到链上
		return append(left, propAccess.Name().Text())
	default:
		// 对于其他复杂情况，返回其源码文本作为唯一标识
		return []string{strings.TrimSpace(utils.GetNodeText(node, sourceCode))}
	}
}

// getArgumentType 是一个辅助函数，用于从 AST 节点确定参数类型。
func getArgumentType(node *ast.Node) string {
	switch node.Kind {
	case ast.KindStringLiteral:
		return "string"
	case ast.KindNumericLiteral:
		return "number"
	case ast.KindIdentifier:
		return "identifier"
	case ast.KindObjectLiteralExpression:
		return "object"
	case ast.KindArrowFunction, ast.KindFunctionExpression:
		return "function"
	case ast.KindArrayLiteralExpression:
		return "array"
	case ast.KindTrueKeyword, ast.KindFalseKeyword:
		return "boolean"
	default:
		return "unknown"
	}
}

// analyzeCallExpression 从给定的 ast.CallExpression 节点中提取详细信息，并填充到 CallExpression 结构体中。
// 它能区分简单的函数调用（如 `myFunc()`）和成员方法调用（如 `myObj.myMethod()`）。
func (ce *CallExpression) analyzeCallExpression(node *ast.CallExpression, sourceCode string) {
	if node == nil {
		return
	}

	// 填充参数信息
	ce.Arguments = make([]Argument, len(node.Arguments.Nodes))
	for i, argNode := range node.Arguments.Nodes {
		ce.Arguments[i] = Argument{
			Type: getArgumentType(argNode.AsNode()),
			Text: utils.GetNodeText(argNode.AsNode(), sourceCode),
		}
	}

	expressionNode := node.Expression

	switch expressionNode.Kind {
	case ast.KindIdentifier:
		ce.Type = "call"
		ce.CallChain = []string{expressionNode.AsIdentifier().Text}

	case ast.KindPropertyAccessExpression:
		ce.Type = "member"
		// 直接从属性访问表达式重建整个调用链
		ce.CallChain = reconstructExpression(expressionNode, sourceCode)

	default:
		ce.Type = "call"
		ce.CallChain = []string{strings.TrimSpace(utils.GetNodeText(expressionNode, sourceCode))}
	}
}
