package parser

import (
	"main/analyzer/utils"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// 解析导出模块
// - 默认导出 default: export default Bird;
// - 命名导出 named:
// 		- export { School, School2 as NewSchool2 };
// 		- export type { CurrentRes };
//  	- export const name = "bird"
//  	- export function name() {}

// ExportModule 导出模块
type ExportModule struct {
	ImportModule string `json:"importModule"` // 模块名, 对应实际导出的内容模块
	Type         string `json:"type"`         // 默认导出: default、命名导出:named、unknown
	Identifier   string `json:"identifier"`   // 唯一标识
}

// ExportDeclarationResult 导出声明结果
type ExportDeclarationResult struct {
	ExportModules  []ExportModule `json:"exportModules"` // 导出的模块内容
	Raw            string         `json:"raw"`           // 源码
	Source         string         `json:"source"`        // 源文件路径  case: export { a } from "../index.ts"
	Type           string         `json:"type"`          // 类型, 预留字段，代表导出的类型，例如：变量/函数/类型等
	SourceLocation SourceLocation `json:"sourceLocation"`
}

func NewExportDeclarationResult(node *ast.ExportDeclaration) *ExportDeclarationResult {
	pos, end := node.Pos(), node.End()
	return &ExportDeclarationResult{
		ExportModules: make([]ExportModule, 0),
		Raw:           "",
		Type:          "",
		SourceLocation: SourceLocation{
			Start: NodePosition{Line: pos, Column: 0},
			End:   NodePosition{Line: end, Column: 0},
		},
	}
}

func (edr *ExportDeclarationResult) analyzeExportDeclaration(node *ast.ExportDeclaration, sourceCode string) {
	// ✅ 解析 import 的源代码
	raw := utils.GetNodeText(node.AsNode(), sourceCode)
	edr.Raw = raw

	// 解析export的模块内容
	// TODO: Implement the logic to analyze the export declaration details

	edr.Type = ""
}