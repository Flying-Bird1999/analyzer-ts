// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象抽象语法树）解析的功能。
// 本文件（importDeclaration.go）专门负责处理和解析导入（Import）声明。
package parser

import (
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"
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
	ImportModules  []ImportModule `json:"importModules"`            // 该导入声明中包含的所有导入模块的列表。
	Raw            string         `json:"raw,omitempty"`            // 节点在源码中的原始文本。
	Source         string         `json:"source"`                   // 导入来源的模块路径，例如 `'./school'`。
	SourceLocation *SourceLocation `json:"sourceLocation,omitempty"` // 节点在源码中的位置信息。
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

// AnalyzeImportDeclaration 是一个公共的、可复用的函数，用于从 AST 节点中解析导入声明的详细信息。
// 它将核心解析逻辑与 `Parser` 的状态解耦，使其可以在其他包（如 analyzer_tree）中被安全地调用。
// node: 要分析的 `ast.ImportDeclaration` 节点。
// sourceCode: 完整的源文件代码，用于提取原始文本。
// 返回一个填充了信息的 `ImportDeclarationResult` 结构体。
func AnalyzeImportDeclaration(node *ast.ImportDeclaration, sourceCode string) *ImportDeclarationResult {
	idr := NewImportDeclarationResult()
	idr.Raw = utils.GetNodeText(node.AsNode(), sourceCode)
	idr.Source = node.ModuleSpecifier.Text()
	idr.SourceLocation = NewSourceLocation(node.AsNode(), sourceCode)

	// 处理副作用导入，例如 `import './setup';`
	if node.ImportClause == nil {
		return idr
	}

	importClause := node.ImportClause.AsImportClause()

	// 处理默认导入: `import MyDefault from '...'`
	if ast.IsDefaultImport(node.AsNode()) {
		name := importClause.Name().Text()
		idr.addModule("default", "default", name)
	}

	// 处理命名空间导入: `import * as ns from '...'`
	if namespaceNode := ast.GetNamespaceDeclarationNode(node.AsNode()); namespaceNode != nil {
		name := namespaceNode.Name().Text()
		idr.addModule("namespace", name, name)
	}

	// 处理命名导入: `import { a, b as c } from '...'`
	if importClause.NamedBindings != nil && importClause.NamedBindings.Kind == ast.KindNamedImports {
		namedImports := importClause.NamedBindings.AsNamedImports()
		for _, element := range namedImports.Elements.Nodes {
			importSpecifier := element.AsImportSpecifier()
			identifier := importSpecifier.Name().Text()
			importModule := identifier
			if importSpecifier.PropertyName != nil {
				importModule = importSpecifier.PropertyName.Text()
			}
			idr.addModule("named", importModule, identifier)
		}
	}

	return idr
}

// VisitImportDeclaration 是 `parser.Parser` 的一部分，它在 AST 遍历期间被调用。
// 它现在将工作委托给可复用的 `AnalyzeImportDeclaration` 函数，然后将结果存入解析器的结果列表中。
func (p *Parser) VisitImportDeclaration(node *ast.ImportDeclaration) {
	result := AnalyzeImportDeclaration(node, p.SourceCode)
	p.Result.ImportDeclarations = append(p.Result.ImportDeclarations, *result)
}