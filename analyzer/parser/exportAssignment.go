// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（exportAssignment.go）专门负责处理 `export default` 语句。
package parser

import (
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
