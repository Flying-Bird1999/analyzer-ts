// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（returnStatement.go）专门负责处理和解析 return 语句。
package parser

import (
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// ReturnStatementResult 存储一个解析后的 return 语句信息。
type ReturnStatementResult struct {
	Expression *VariableValue `json:"expression"` // return 后面跟随的表达式
	Node       *ast.Node      `json:"-"`           // 对应的 AST 节点，不在 JSON 中序列化。
}

// AnalyzeReturnStatement 是一个公共的、可复用的函数，用于从 AST 节点中解析 return 语句。
func AnalyzeReturnStatement(node *ast.ReturnStatement, sourceCode string) *ReturnStatementResult {
	if node == nil {
		return nil
	}
	return &ReturnStatementResult{
		Expression: AnalyzeVariableValueNode(node.Expression, sourceCode),
		Node:       node.AsNode(),
	}
}
