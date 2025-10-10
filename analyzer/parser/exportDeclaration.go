// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（exportDeclaration.go）专门负责处理和解析导出（Export）声明。
package parser

import (
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"
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
	ModuleName string `json:"moduleName"` // 模块名, 对应实际导出的内容模块。例如 `export { a as b }` 中的 `a`。
	Type       string `json:"type"`       // 导出类型: `named` (命名导出), `namespace` (命名空间导出)。
	Identifier string `json:"identifier"` // 导出的标识符。例如 `export { a as b }` 中的 `b`。
}

// ExportDeclarationResult 存储一个完整的导出声明的解析结果。
// 一个导出声明（例如 `export { a, b } from './mod'`) 可能包含多个导出的模块。
type ExportDeclarationResult struct {
	ExportModules  []ExportModule `json:"exportModules"`            // 该导出声明中包含的所有导出模块的列表。
	Raw            string         `json:"raw,omitempty"`            // 节点在源码中的原始文本。
	Source         string         `json:"source,omitempty"`         // 导出来源的模块路径。例如 `export { a } from "../index.ts"` 中的 `"../index.ts"`。
	Type           string         `json:"type"`                      // 导出类型: `re-export` (重导出) 或 `named-export` (命名导出)。
	SourceLocation *SourceLocation `json:"sourceLocation,omitempty"` // 节点在源码中的位置信息。
	Node           *ast.Node      `json:"-"`                     // 对应的 AST 节点，不在 JSON 中序列化。
}

// AnalyzeExportDeclaration 是一个公共的、可复用的函数，用于从 AST 节点中解析导出声明的详细信息。
func AnalyzeExportDeclaration(node *ast.ExportDeclaration, sourceCode string) *ExportDeclarationResult {
	edr := &ExportDeclarationResult{
		ExportModules:  make([]ExportModule, 0),
		Raw:            utils.GetNodeText(node.AsNode(), sourceCode),
		SourceLocation: NewSourceLocation(node.AsNode(), sourceCode),
		Node:           node.AsNode(),
	}

	// 检查是否存在模块说明符（例如 `from './module'`），如果存在，则为重导出
	if node.ModuleSpecifier != nil {
		edr.Source = node.ModuleSpecifier.Text()
		edr.Type = "re-export"
	} else {
		edr.Type = "named-export"
	}

	// 检查是否存在导出子句（例如 `{ a, b }` 或 `* as ns`）
	if node.ExportClause != nil {
		// 处理命名导出 `export { a, b as c }`
		if node.ExportClause.Kind == ast.KindNamedExports {
			namedExports := node.ExportClause.AsNamedExports()
			for _, element := range namedExports.Elements.Nodes {
				specifier := element.AsExportSpecifier()
				identifier := specifier.Name().Text()
				moduleName := identifier
				// 处理别名 `export { a as b }`
				if specifier.PropertyName != nil {
					moduleName = specifier.PropertyName.Text()
				}
				edr.ExportModules = append(edr.ExportModules, ExportModule{
					ModuleName: moduleName,
					Type:       "named",
					Identifier: identifier,
				})
			}
			// 处理命名空间导出 `export * as ns from './module'`
		} else if node.ExportClause.Kind == ast.KindNamespaceExport {
			namespaceExport := node.ExportClause.AsNamespaceExport()
			identifier := namespaceExport.Name().Text()
			edr.ExportModules = append(edr.ExportModules, ExportModule{
				ModuleName: "*",
				Type:       "namespace",
				Identifier: identifier,
				})
		}
	} else {
		// 处理 `export * from './module'`
		if edr.Source != "" {
			edr.ExportModules = append(edr.ExportModules, ExportModule{
				ModuleName: "*",
				Type:       "namespace",
				Identifier: "*",
			})
		}
	}
	return edr
}

// VisitExportDeclaration 是 `parser.Parser` 的一部分，在 AST 遍历时被调用。
// 它现在将工作委托给可复用的 `AnalyzeExportDeclaration` 函数。
func (p *Parser) VisitExportDeclaration(node *ast.ExportDeclaration) {
	result := AnalyzeExportDeclaration(node, p.SourceCode)
	p.Result.ExportDeclarations = append(p.Result.ExportDeclarations, *result)
}