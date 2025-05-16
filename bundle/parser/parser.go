package parser

import (
	"fmt"
	"main/bundle/utils"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

type ParserResult struct {
	ImportDeclarations    []ImportDeclarationResult
	InterfaceDeclarations []InterfaceDeclarationResult
	TypeDeclarations      []TypeDeclarationResult
}

func NewBundleResult() ParserResult {
	return ParserResult{
		ImportDeclarations:    []ImportDeclarationResult{},
		InterfaceDeclarations: []InterfaceDeclarationResult{},
		TypeDeclarations:      []TypeDeclarationResult{},
	}
}

func (pr *ParserResult) AddImportDeclaration(idr *ImportDeclarationResult) {
	pr.ImportDeclarations = append(pr.ImportDeclarations, *idr)
}

func (pr *ParserResult) AddInterfaceDeclaration(inter *InterfaceDeclarationResult) {
	pr.InterfaceDeclarations = append(pr.InterfaceDeclarations, *inter)
}

func (pr *ParserResult) addTypeDeclaration(tr *TypeDeclarationResult) {
	pr.TypeDeclarations = append(pr.TypeDeclarations, *tr)
}

func Traverse(filePath string) ParserResult {
	sourceCode, err := utils.ReadFileContent(filePath)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
	}

	sourceFile := utils.ParseTypeScriptFile(filePath, sourceCode)
	bundle := NewBundleResult()

	for _, node := range sourceFile.Statements.Nodes {
		// 解析 import
		if node.Kind == ast.KindImportDeclaration {
			idr := NewImportDeclarationResult()
			idr.analyzeImportDeclaration(node.AsImportDeclaration(), sourceCode)
			bundle.AddImportDeclaration(idr)
		}

		// 解析 interface
		if node.Kind == ast.KindInterfaceDeclaration {
			inter := NewInterfaceDeclarationResult(node.AsNode(), sourceCode)
			inter.analyzeInterfaces(node.AsInterfaceDeclaration())
			bundle.AddInterfaceDeclaration(inter)
		}

		// 解析 type
		if node.Kind == ast.KindTypeAliasDeclaration {
			tr := NewTypeDeclarationResult(node.AsNode(), sourceCode)
			tr.analyzeTypeDecl(node.AsTypeAliasDeclaration())
			bundle.addTypeDeclaration(tr)
		}
	}

	// // 解析 Interface 中的 type
	// for _, inter := range bundle.InterfaceDeclarations {
	// 	fmt.Printf("Name: %s\n", inter.Name)
	// 	fmt.Printf("Raw: %s\n", inter.Raw)
	// 	for _, ref := range inter.Reference {
	// 		fmt.Printf("Reference: %s, %v, %b \n", ref.Name, ref.Location, ref.IsExtend)
	// 	}
	// 	fmt.Print("\n\n\n")
	// }

	// // 解析 Type 中的 type
	// for _, tr := range bundle.TypeDeclarations {
	// 	fmt.Printf("Name: %s\n", tr.Name)
	// 	fmt.Printf("Raw: %s\n", tr.Raw)
	// 	for _, ref := range tr.Reference {
	// 		fmt.Printf("Reference: %s, %v, %b \n", ref.Name, ref.Location, ref.IsExtend)
	// 	}
	// 	fmt.Print("\n\n\n")
	// }

	return bundle
}
