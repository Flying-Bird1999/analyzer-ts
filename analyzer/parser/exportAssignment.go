// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（exportAssignment.go）专门负责处理 `export default` 语句。
package parser

import (
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// ExportAssignmentResult 存储一个 `export default` 声明的解析结果。
type ExportAssignmentResult struct {
	Expression     string         `json:"expression,omitempty"`     // 导出的表达式的文本。
	Name           string         `json:"name,omitempty"`           // 导出的符号名（如 export default foo → "foo"）
	Raw            string         `json:"raw,omitempty"`            // 节点在源码中的原始文本。
	SourceLocation *SourceLocation `json:"sourceLocation,omitempty"` // 节点在源码中的位置信息。
	Node           *ast.Node      `json:"-"`                     // 对应的 AST 节点，不在 JSON 中序列化。
}

// AnalyzeExportAssignment 是一个公共的、可复用的函数，用于从 AST 节点中解析 `export default` 声明。
func AnalyzeExportAssignment(node *ast.ExportAssignment, sourceCode string) *ExportAssignmentResult {
	expr := strings.TrimSpace(utils.GetNodeText(node.Expression, sourceCode))
	return &ExportAssignmentResult{
		Raw:            utils.GetNodeText(node.AsNode(), sourceCode),
		Expression:     expr,
		Name:           extractNameFromExpression(expr),
		SourceLocation: NewSourceLocation(node.AsNode(), sourceCode),
		Node:           node.AsNode(),
	}
}

// extractNameFromExpression 从表达式中提取符号名
// export default formatDate → "formatDate"
// export default function Button → "Button"
// export default class Foo → "Foo"
// export default function useCounter(...) → "useCounter"
// export default () => {} → ""（匿名导出）
func extractNameFromExpression(expr string) string {
	expr = strings.TrimSpace(expr)

	// 移除 "function " 和 "class " 关键字
	expr = strings.TrimPrefix(expr, "function ")
	expr = strings.TrimPrefix(expr, "class ")

	// 移除前导空格
	expr = strings.TrimSpace(expr)

	// 提取第一个标识符（在遇到任何非标识符字符之前）
	// 标识符可以包含：字母、数字、下划线
	var result strings.Builder
	for i, r := range expr {
		if i == 0 {
			// 第一个字符必须是字母或下划线
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' {
				result.WriteRune(r)
			} else {
				break
			}
		} else {
			// 后续字符可以是字母、数字或下划线
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
				result.WriteRune(r)
			} else {
				break
			}
		}
	}

	return result.String()
}

// VisitExportAssignment 是 `parser.Parser` 的一部分，在 AST 遍历时被调用。
// 它现在将工作委托给可复用的 `AnalyzeExportAssignment` 函数。
func (p *Parser) VisitExportAssignment(node *ast.ExportAssignment) {
	result := AnalyzeExportAssignment(node, p.SourceCode)
	p.Result.ExportAssignments = append(p.Result.ExportAssignments, *result)
}