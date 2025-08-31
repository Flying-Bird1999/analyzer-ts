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
	Raw            string         `json:"raw"`            // `export default ...` 语句的完整原始文本。
	Expression     string         `json:"expression"`     // 被导出的表达式本身的文本。
	SourceLocation SourceLocation `json:"sourceLocation"` // 节点在源码中的位置信息。
}

// AnalyzeExportAssignment 是一个公共的、可复用的函数，用于从 AST 节点中解析 `export default` 声明。
func AnalyzeExportAssignment(node *ast.ExportAssignment, sourceCode string) *ExportAssignmentResult {
	return &ExportAssignmentResult{
		Raw:            utils.GetNodeText(node.AsNode(), sourceCode),
		Expression:     strings.TrimSpace(utils.GetNodeText(node.Expression, sourceCode)),
		SourceLocation: NewSourceLocation(node.AsNode(), sourceCode),
	}
}

// VisitExportAssignment 是 `parser.Parser` 的一部分，在 AST 遍历时被调用。
// 它现在将工作委托给可复用的 `AnalyzeExportAssignment` 函数。
func (p *Parser) VisitExportAssignment(node *ast.ExportAssignment) {
	result := AnalyzeExportAssignment(node, p.SourceCode)
	p.Result.ExportAssignments = append(p.Result.ExportAssignments, *result)
}