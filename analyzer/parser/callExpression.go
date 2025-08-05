package parser

import (
	"main/analyzer/utils"
	"strings"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// CallExpression 代表一个函数或方法调用，根据新的需求定制
type CallExpression struct {
	Identifier     string         `json:"identifier"`         // 标识符，例如 myObj.myMethod() 中的 myObj
	Property       string         `json:"property,omitempty"` // 属性，例如 myObj.myMethod() 中的 myMethod
	ArgLen         int            `json:"argLen"`             // 参数数量
	Type           string         `json:"type"`               // 类型，固定为 'call' 或 'member'
	Raw            string         `json:"raw,omitempty"`      // 原始源码
	SourceLocation SourceLocation `json:"sourceLocation"`     // 源码位置
}

// NewCallExpression 创建一个新的 CallExpression 实例
func NewCallExpression(node *ast.CallExpression, sourceCode string) *CallExpression {
	pos, end := node.Pos(), node.End()
	return &CallExpression{
		SourceLocation: SourceLocation{
			Start: NodePosition{Line: pos, Column: 0},
			End:   NodePosition{Line: end, Column: 0},
		},
		Raw:  utils.GetNodeText(node.AsNode(), sourceCode),
		Type: "call", // 默认为 'call'，在解析时可能会被修改
	}
}

// reconstructExpression 从表达式节点递归地构建一个干净的标识符字符串
func reconstructExpression(node *ast.Node, sourceCode string) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	case ast.KindPropertyAccessExpression:
		propAccess := node.AsPropertyAccessExpression()
		left := reconstructExpression(propAccess.Expression, sourceCode)
		right := propAccess.Name().Text()
		if left != "" {
			return left + "." + right
		}
		return right
	default:
		// 对其他表达式类型进行回退，并清理空格
		return strings.TrimSpace(utils.GetNodeText(node, sourceCode))
	}
}

// analyzeCallExpression 根据新的结构从 ast.CallExpression 节点中提取信息
func (ce *CallExpression) analyzeCallExpression(node *ast.CallExpression, sourceCode string) {
	if node == nil {
		return
	}

	// 获取参数数量
	ce.ArgLen = len(node.Arguments.Nodes)

	// 分析被调用的表达式以确定标识符和属性
	expressionNode := node.Expression

	switch expressionNode.Kind {
	case ast.KindIdentifier:
		// 这是一个简单的函数调用, 例如 myFunc()
		ce.Identifier = expressionNode.AsIdentifier().Text
		ce.Property = ""

	case ast.KindPropertyAccessExpression:
		// 这是一个对象上的方法调用, 例如 myObj.myMethod()
		propAccess := expressionNode.AsPropertyAccessExpression()
		ce.Identifier = reconstructExpression(propAccess.Expression, sourceCode)
		ce.Property = propAccess.Name().Text()
		ce.Type = "member"

	default:
		// 对其他类型的表达式（例如，IIFE、new表达式）进行回退
		// 我们将整个表达式记录为标识符，并使属性为空
		ce.Identifier = strings.TrimSpace(utils.GetNodeText(expressionNode, sourceCode))
		ce.Property = ""
	}
}
