// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（exportAssignment.go）专门负责处理 `export default` 语句。
package parser

import (
	"main/analyzer/utils"
	"strings"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// ExportAssignmentResult 存储一个 `export default` 声明的解析结果。
type ExportAssignmentResult struct {
	Raw            string         `json:"raw"`            // `export default ...` 语句的完整原始文本。
	Expression     string         `json:"expression"`     // 被导出的表达式本身的文本。
	SourceLocation SourceLocation `json:"sourceLocation"` // 节点在源码中的位置信息。
}

// NewExportAssignmentResult 基于 AST 节点创建一个新的 ExportAssignmentResult 实例。
func NewExportAssignmentResult(node *ast.ExportAssignment) *ExportAssignmentResult {
	pos, end := node.Pos(), node.End()
	return &ExportAssignmentResult{
		SourceLocation: SourceLocation{
			Start: NodePosition{Line: pos, Column: 0},
			End:   NodePosition{Line: end, Column: 0},
		},
	}
}

// AnalyzeExportAssignment 从 `export default` 节点中提取信息。
func (ear *ExportAssignmentResult) AnalyzeExportAssignment(node *ast.ExportAssignment, sourceCode string) {
	ear.Raw = utils.GetNodeText(node.AsNode(), sourceCode)
	// 直接从源码中获取表达式的文本，以避免第三方库中 .Text() 方法可能存在的 bug (例如处理函数调用时)。
	// 同时，使用 TrimSpace 清理可能存在的前后多余的空白字符。
	ear.Expression = strings.TrimSpace(utils.GetNodeText(node.Expression, sourceCode))
}
