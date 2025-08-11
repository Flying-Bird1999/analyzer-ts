// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（exportDeclaration.go）专门负责处理和解析导出（Export）声明。
package parser

import (
	"main/analyzer/utils"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// ExportModule 代表一个被导出的独立实体。
// TypeScript 的导出语法多样，此结构用于统一表示不同类型的导出项。
// 例如：
// - `export default Bird;` (默认导出)
// - `export { School };` (命名导出)
// - `export { School2 as NewSchool2 };` (带别名的命名导出)
// - `export const name = "bird";` (导出变量)
type ExportModule struct {
	ImportModule string `json:"importModule"` // 模块名, 对应实际导出的内容模块。例如 `export { a as b }` 中的 `a`。
	Type         string `json:"type"`         // 导出类型。可以是 `default` (默认导出), `named` (命名导出), 或 `unknown`。
	Identifier   string `json:"identifier"`   // 导出的标识符。例如 `export { a as b }` 中的 `b`，或者 `export default A` 中的 `A`。
}

// ExportDeclarationResult 存储一个完整的导出声明的解析结果。
// 一个导出声明（例如 `export { a, b } from './mod'`) 可能包含多个导出的模块。
type ExportDeclarationResult struct {
	ExportModules  []ExportModule `json:"exportModules"`  // 该导出声明中包含的所有导出模块的列表。
	Raw            string         `json:"raw"`            // 节点在源码中的原始文本。
	Source         string         `json:"source"`         // 导出来源的模块路径。例如 `export { a } from "../index.ts"` 中的 `"../index.ts"`。
	Type           string         `json:"type"`           // 预留字段，未来可用于表示导出的具体类型（如：变量、函数、类等）。
	SourceLocation SourceLocation `json:"sourceLocation"` // 节点在源码中的位置信息。
}

// NewExportDeclarationResult 基于 AST 节点创建一个新的 ExportDeclarationResult 实例。
// 它初始化了结果结构体，并设置了源码位置。
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

// analyzeExportDeclaration 从给定的 ast.ExportDeclaration 节点中提取信息。
// 注意：当前的实现是基础的，并且有待完善。
func (edr *ExportDeclarationResult) AnalyzeExportDeclaration(node *ast.ExportDeclaration, sourceCode string) {
	// 提取节点在源码中的原始文本。
	raw := utils.GetNodeText(node.AsNode(), sourceCode)
	edr.Raw = raw

	// TODO: 实现完整的导出声明分析逻辑。
	// 需要处理以下几种情况：
	// 1. 命名导出: `export { name1, name2 as alias }`
	// 2. 默认导出: `export default myExpression`
	// 3. 重导出: `export * from './module'` 或 `export { name } from './module'`
	// 4. 导出声明: `export const a = 1;` 或 `export function b() {}` (这部分可能在 `Traverse` 的其他 case 中处理)

	// 预留类型字段，当前未实现。
	edr.Type = ""
}
