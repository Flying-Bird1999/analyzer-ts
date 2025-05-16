package parser

import (
	"fmt"
	"main/bundle/utils"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

type BundleResult struct {
	ImportDeclarations    []ImportDeclarationResult
	InterfaceDeclarations []InterfaceDeclarationResult
	TypeDeclarations      []TypeDeclarationResult
}

func NewBundleResult() BundleResult {
	return BundleResult{
		ImportDeclarations:    []ImportDeclarationResult{},
		InterfaceDeclarations: []InterfaceDeclarationResult{},
		TypeDeclarations:      []TypeDeclarationResult{},
	}
}

func (br *BundleResult) AddImportDeclaration(idr *ImportDeclarationResult) {
	br.ImportDeclarations = append(br.ImportDeclarations, *idr)
}

func (br *BundleResult) AddInterfaceDeclaration(inter *InterfaceDeclarationResult) {
	br.InterfaceDeclarations = append(br.InterfaceDeclarations, *inter)
}

func (br *BundleResult) addTypeDeclaration(tr *TypeDeclarationResult) {
	br.TypeDeclarations = append(br.TypeDeclarations, *tr)
}

func Traverse(filePath string) BundleResult {
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
	return bundle
}
