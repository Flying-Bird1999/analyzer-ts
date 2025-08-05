// 单文件解析AST
package parser

import (
	"fmt"
	"main/analyzer/utils"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

type ParserResult struct {
	filePath string

	ImportDeclarations    []ImportDeclarationResult
	ExportDeclarations    []ExportDeclarationResult
	InterfaceDeclarations map[string]InterfaceDeclarationResult
	TypeDeclarations      map[string]TypeDeclarationResult
	EnumDeclarations      map[string]EnumDeclarationResult
	VariableDeclarations  []VariableDeclaration
	CallExpressions       []CallExpression
}

func NewParserResult(filePath string) ParserResult {
	return ParserResult{
		filePath:              filePath,
		ImportDeclarations:    []ImportDeclarationResult{},
		InterfaceDeclarations: make(map[string]InterfaceDeclarationResult),
		TypeDeclarations:      make(map[string]TypeDeclarationResult),
		EnumDeclarations:      make(map[string]EnumDeclarationResult),
		VariableDeclarations:  []VariableDeclaration{},
		CallExpressions:       []CallExpression{},
	}
}

func (pr *ParserResult) AddImportDeclaration(idr *ImportDeclarationResult) {
	pr.ImportDeclarations = append(pr.ImportDeclarations, *idr)
}

func (pr *ParserResult) AddExportDeclaration(edr *ExportDeclarationResult) {
	pr.ExportDeclarations = append(pr.ExportDeclarations, *edr)
}

func (pr *ParserResult) AddInterfaceDeclaration(inter *InterfaceDeclarationResult) {
	pr.InterfaceDeclarations[inter.Identifier] = *inter
}

func (pr *ParserResult) addTypeDeclaration(tr *TypeDeclarationResult) {
	pr.TypeDeclarations[tr.Identifier] = *tr
}

func (pr *ParserResult) addEnumDeclaration(er *EnumDeclarationResult) {
	pr.EnumDeclarations[er.Identifier] = *er
}

func (pr *ParserResult) addVariableDeclaration(vd *VariableDeclaration) {
	pr.VariableDeclarations = append(pr.VariableDeclarations, *vd)
}

func (pr *ParserResult) addCallExpression(ce *CallExpression) {
	pr.CallExpressions = append(pr.CallExpressions, *ce)
}

func (pr *ParserResult) GetResult() ParserResult {
	return ParserResult{
		ImportDeclarations:    pr.ImportDeclarations,
		ExportDeclarations:    pr.ExportDeclarations,
		InterfaceDeclarations: pr.InterfaceDeclarations,
		TypeDeclarations:      pr.TypeDeclarations,
		EnumDeclarations:      pr.EnumDeclarations,
		VariableDeclarations:  pr.VariableDeclarations,
		CallExpressions:       pr.CallExpressions,
	}
}

func (pr *ParserResult) Traverse() {
	sourceCode, err := utils.ReadFileContent(pr.filePath)
	if err != nil {
		fmt.Printf("Failed to read file: %s\n", pr.filePath)
		return
	}

	sourceFile := utils.ParseTypeScriptFile(pr.filePath, sourceCode)

	var walk func(node *ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}

		switch node.Kind {
		case ast.KindImportDeclaration:
			idr := NewImportDeclarationResult()
			idr.analyzeImportDeclaration(node.AsImportDeclaration(), sourceCode)
			pr.AddImportDeclaration(idr)
			// Stop recursion for imports
			return

		case ast.KindInterfaceDeclaration:
			inter := NewInterfaceDeclarationResult(node, sourceCode)
			inter.analyzeInterfaces(node.AsInterfaceDeclaration())
			pr.AddInterfaceDeclaration(inter)

		case ast.KindTypeAliasDeclaration:
			tr := NewTypeDeclarationResult(node, sourceCode)
			tr.analyzeTypeDecl(node.AsTypeAliasDeclaration())
			pr.addTypeDeclaration(tr)

		case ast.KindEnumDeclaration:
			er := NewEnumDeclarationResult(node.AsEnumDeclaration(), sourceCode)
			pr.addEnumDeclaration(er)

		case ast.KindVariableStatement:
			vd := NewVariableDeclaration(node.AsVariableStatement(), sourceCode)
			vd.analyzeVariableDeclaration(node.AsVariableStatement(), sourceCode, sourceFile)
			pr.addVariableDeclaration(vd)

		case ast.KindCallExpression:
			callExpr := node.AsCallExpression()
			ce := NewCallExpression(callExpr, sourceCode)
			ce.analyzeCallExpression(callExpr, sourceCode)
			pr.addCallExpression(ce)
		}

		// Correctly recurse using the library's ForEachChild method
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false // continue traversal
		})
	}

	walk(sourceFile.AsNode())
}
