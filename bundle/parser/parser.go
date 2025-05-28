package parser

import (
	"fmt"
	"main/bundle/utils"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

type ParserResult struct {
	filePath string

	ImportDeclarations    []ImportDeclarationResult
	ExportDeclarations    []ExportDeclarationResult
	InterfaceDeclarations map[string]InterfaceDeclarationResult
	TypeDeclarations      map[string]TypeDeclarationResult
}

func NewParserResult(filePath string) ParserResult {
	return ParserResult{
		filePath:              filePath,
		ImportDeclarations:    []ImportDeclarationResult{},
		InterfaceDeclarations: make(map[string]InterfaceDeclarationResult),
		TypeDeclarations:      make(map[string]TypeDeclarationResult),
	}
}

func (pr *ParserResult) AddImportDeclaration(idr *ImportDeclarationResult) {
	pr.ImportDeclarations = append(pr.ImportDeclarations, *idr)
}

func (pr *ParserResult) AddExportDeclaration(edr *ExportDeclarationResult) {
	pr.ExportDeclarations = append(pr.ExportDeclarations, *edr)
}

func (pr *ParserResult) AddInterfaceDeclaration(inter *InterfaceDeclarationResult) {
	pr.InterfaceDeclarations[inter.Name] = *inter
}

func (pr *ParserResult) addTypeDeclaration(tr *TypeDeclarationResult) {
	pr.TypeDeclarations[tr.Name] = *tr
}

func (pr *ParserResult) GetResult() ParserResult {
	return ParserResult{
		ImportDeclarations:    pr.ImportDeclarations,
		ExportDeclarations:    pr.ExportDeclarations,
		InterfaceDeclarations: pr.InterfaceDeclarations,
		TypeDeclarations:      pr.TypeDeclarations,
	}
}

func (pr *ParserResult) Traverse() {
	sourceCode, err := utils.ReadFileContent(pr.filePath)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
	}

	sourceFile := utils.ParseTypeScriptFile(pr.filePath, sourceCode)

	for _, node := range sourceFile.Statements.Nodes {
		// 解析 import
		if node.Kind == ast.KindImportDeclaration {
			idr := NewImportDeclarationResult()
			idr.analyzeImportDeclaration(node.AsImportDeclaration(), sourceCode)
			pr.AddImportDeclaration(idr)
		}

		// 解析 export
		if node.Kind == ast.KindExportAssignment || node.Kind == ast.KindExportDeclaration || node.Kind == ast.KindNamedExports || node.Kind == ast.KindNamespaceExport || node.Kind == ast.KindExportSpecifier {
			fmt.Print("export...")
			edr := NewExportDeclarationResult()
			edr.analyzeExportDeclaration(node.AsExportDeclaration(), sourceCode)
			pr.AddExportDeclaration(edr)
		}

		// 解析 interface
		if node.Kind == ast.KindInterfaceDeclaration {
			inter := NewInterfaceDeclarationResult(node.AsNode(), sourceCode)
			inter.analyzeInterfaces(node.AsInterfaceDeclaration())
			pr.AddInterfaceDeclaration(inter)
		}

		// 解析 type
		if node.Kind == ast.KindTypeAliasDeclaration {
			tr := NewTypeDeclarationResult(node.AsNode(), sourceCode)
			tr.analyzeTypeDecl(node.AsTypeAliasDeclaration())
			pr.addTypeDeclaration(tr)
		}
	}

}
