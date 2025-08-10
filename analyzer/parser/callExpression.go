// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（callExpression.go）专门负责处理和解析函数/方法调用表达式。
package parser

import (
	"main/analyzer/utils"
	"strings"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// CallExpression 代表一个函数或方法调用表达式的解析结果。
// 它旨在捕获调用的关键信息，如调用者（Identifier）、被调用的方法（Property）以及参数数量。
type CallExpression struct {
	Identifier     string         `json:"identifier"`         // 调用的主体或命名空间。例如 `myObj.myMethod()` 中的 `myObj`，或 `myFunc()` 中的 `myFunc`。
	Property       string         `json:"property,omitempty"` // 调用的属性或方法名。例如 `myObj.myMethod()` 中的 `myMethod`。对于简单函数调用，此字段为空。
	ArgLen         int            `json:"argLen"`             // 调用时传递的参数数量。
	Type           string         `json:"type"`               // 调用的类型。`call` 表示普通函数调用，`member` 表示对象成员方法调用。
	Raw            string         `json:"raw,omitempty"`      // 节点在源码中的原始文本。
	SourceLocation SourceLocation `json:"sourceLocation"`     // 节点在源码中的位置信息（起止行号和列号）。
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

// reconstructExpression 从表达式节点递归地构建一个干净的、点分隔的标识符字符串。
// 例如，对于 `a.b.c()` 中的 `a.b.c` 部分，此函数会将其重建为字符串 "a.b.c"。
// 这对于处理深度嵌套的属性访问表达式非常有用。
func reconstructExpression(node *ast.Node, sourceCode string) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	// 如果是标识符，直接返回其文本。
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	// 如果是属性访问（如 a.b），则递归地重建左侧部分（a），然后附加右侧部分的名称（b）。
	case ast.KindPropertyAccessExpression:
		propAccess := node.AsPropertyAccessExpression()
		left := reconstructExpression(propAccess.Expression, sourceCode)
		right := propAccess.Name().Text()
		if left != "" {
			return left + "." + right
		}
		return right
	default:
		// 对于其他类型的表达式（例如，函数调用返回的对象），直接获取其源码文本并去除多余空格。
		return strings.TrimSpace(utils.GetNodeText(node, sourceCode))
	}
}

// analyzeCallExpression 从给定的 ast.CallExpression 节点中提取详细信息，并填充到 CallExpression 结构体中。
// 它能区分简单的函数调用（如 `myFunc()`）和成员方法调用（如 `myObj.myMethod()`）。
func (ce *CallExpression) analyzeCallExpression(node *ast.CallExpression, sourceCode string) {
	if node == nil {
		return
	}

	// 获取参数列表的长度。
	ce.ArgLen = len(node.Arguments.Nodes)

	// 分析被调用的表达式（`node.Expression`）以确定标识符和属性。
	expressionNode := node.Expression

	switch expressionNode.Kind {
	// 简单函数调用，例如 `myFunc()`。
	case ast.KindIdentifier:
		ce.Identifier = expressionNode.AsIdentifier().Text
		ce.Property = ""

	// 对象成员方法调用，例如 `myObj.myMethod()`。
	case ast.KindPropertyAccessExpression:
		propAccess := expressionNode.AsPropertyAccessExpression()
		// 递归重建调用主体作为标识符。
		ce.Identifier = reconstructExpression(propAccess.Expression, sourceCode)
		// 获取方法名作为属性。
		ce.Property = propAccess.Name().Text()
		// 将类型标记为 'member'。
		ce.Type = "member"

	default:
		// 对于其他复杂情况（例如，立即执行函数表达式 IIFE），
		// 将整个被调用表达式的文本作为标识符，属性字段留空。
		ce.Identifier = strings.TrimSpace(utils.GetNodeText(expressionNode, sourceCode))
		ce.Property = ""
	}
}
