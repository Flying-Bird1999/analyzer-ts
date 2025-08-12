// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象抽象语法树）解析的功能。
// 本文件（importDeclaration.go）专门负责处理和解析导入（Import）声明。
package parser

import (
	"main/analyzer/utils"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// ImportModule 代表一个被导入的独立实体。
// 它用于表示默认导入、命名导入或命名空间导入中的具体项。
type ImportModule struct {
	ImportModule string `json:"importModule"` // 原始模块名。对于 `import { a as b }` 是 `a`；对于默认导入是 `default`；对于命名空间导入是命名空间名称。
	Type         string `json:"type"`         // 导入类型: `default`, `namespace`, `named`。
	Identifier   string `json:"identifier"`   // 在当前文件中使用的标识符。对于 `import { a as b }` 是 `b`；对于 `import a` 是 `a`。
}

// ImportDeclarationResult 存储一个完整的导入声明的解析结果。
// 一个导入声明（例如 `import a, { b } from './mod'`) 可能包含多个导入的模块。
type ImportDeclarationResult struct {
	ImportModules  []ImportModule `json:"importModules"`  // 该导入声明中包含的所有导入模块的列表。
	Raw            string         `json:"raw"`            // 节点在源码中的原始文本。
	Source         string         `json:"source"`         // 导入来源的模块路径，例如 `'./school'`。
	SourceLocation SourceLocation `json:"sourceLocation"` // 节点在源码中的位置信息。
}

// NewImportDeclarationResult 创建并初始化一个 ImportDeclarationResult 实例。
func NewImportDeclarationResult() *ImportDeclarationResult {
	return &ImportDeclarationResult{
		ImportModules: make([]ImportModule, 0),
	}
}

// addModule 是一个辅助函数，用于向 ImportDeclarationResult 添加一个新的导入模块。
func (idr *ImportDeclarationResult) addModule(moduleType, importModule, identifier string) {
	idr.ImportModules = append(idr.ImportModules, ImportModule{
		Type:         moduleType,
		ImportModule: importModule,
		Identifier:   identifier,
	})
}

// AnalyzeImportDeclaration 从给定的 ast.ImportDeclaration 节点中提取详细信息。
// 它能够处理默认导入、命名空间导入和命名导入（包括带别名的导入）。
func (idr *ImportDeclarationResult) AnalyzeImportDeclaration(node *ast.ImportDeclaration, sourceCode string) {
	// 提取基本信息：原始文本、来源和位置。
	idr.Raw = utils.GetNodeText(node.AsNode(), sourceCode)
	idr.Source = node.ModuleSpecifier.Text()
	pos, end := node.Pos(), node.End()
	idr.SourceLocation = SourceLocation{
		Start: NodePosition{Line: pos, Column: 0},
		End:   NodePosition{Line: end, Column: 0},
	}

	// `ImportClause` 为 nil 表示这是一个纯副作用导入，例如 `import './setup';`，直接返回。
	if node.ImportClause == nil {
		return
	}

	importClause := node.ImportClause.AsImportClause()

	// Case 1: 默认导入, 例如: `import Bird from './type2';`
	if ast.IsDefaultImport(node.AsNode()) {
		name := importClause.Name().Text()
		idr.addModule("default", "default", name)
	}

	// Case 2: 命名空间导入, 例如: `import * as allTypes from './type';`
	if namespaceNode := ast.GetNamespaceDeclarationNode(node.AsNode()); namespaceNode != nil {
		name := namespaceNode.Name().Text()
		idr.addModule("namespace", name, name)
	}

	// Case 3: 命名导入, 例如: `import { School, School2 as NewSchool } from './school';`
	if importClause.NamedBindings != nil && importClause.NamedBindings.Kind == ast.KindNamedImports {
		namedImports := importClause.NamedBindings.AsNamedImports()
		for _, element := range namedImports.Elements.Nodes {
			importSpecifier := element.AsImportSpecifier()

			identifier := importSpecifier.Name().Text()
			importModule := identifier
			// 如果 PropertyName 存在，说明是带别名的导入，原始模块名为 PropertyName。
			if importSpecifier.PropertyName != nil {
				importModule = importSpecifier.PropertyName.Text()
			}
			idr.addModule("named", importModule, identifier)
		}
	}
}