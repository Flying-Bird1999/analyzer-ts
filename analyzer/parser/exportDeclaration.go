// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（exportDeclaration.go）专门负责处理和解析导出（Export）声明。
package parser

import (
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"
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
	ExportModules  []ExportModule `json:"exportModules"`  // 该导出声明中包含的所有导出模块的列表。
	Raw            string         `json:"raw"`            // 节点在源码中的原始文本。
	Source         string         `json:"source"`         // 导出来源的模块路径。例如 `export { a } from "../index.ts"` 中的 `"../index.ts"`。
	Type           string         `json:"type"`           // 导出类型: `re-export` (重导出) 或 `named-export` (命名导出)。
	SourceLocation SourceLocation `json:"sourceLocation"` // 节点在源码中的位置信息。
}

// NewExportDeclarationResult 基于 AST 节点创建一个新的 ExportDeclarationResult 实例。
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

// VisitExportDeclaration 解析导出声明。
func (p *Parser) VisitExportDeclaration(node *ast.ExportDeclaration) {
	edr := NewExportDeclarationResult(node)
	edr.Raw = utils.GetNodeText(node.AsNode(), p.SourceCode)

	if node.ModuleSpecifier != nil {
		edr.Source = node.ModuleSpecifier.Text()
		edr.Type = "re-export"
	} else {
		edr.Type = "named-export"
	}

	if node.ExportClause != nil {
		if node.ExportClause.Kind == ast.KindNamedExports {
			namedExports := node.ExportClause.AsNamedExports()
			for _, element := range namedExports.Elements.Nodes {
				specifier := element.AsExportSpecifier()
				identifier := specifier.Name().Text()
				moduleName := identifier
				if specifier.PropertyName != nil {
					moduleName = specifier.PropertyName.Text()
				}
				edr.ExportModules = append(edr.ExportModules, ExportModule{
					ModuleName: moduleName,
					Type:       "named",
					Identifier: identifier,
				})
			}
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
		if edr.Source != "" {
			edr.ExportModules = append(edr.ExportModules, ExportModule{
				ModuleName: "*",
				Type:       "namespace",
				Identifier: "*",
			})
		}
	}
	p.Result.ExportDeclarations = append(p.Result.ExportDeclarations, *edr)
}
