package parser

import (
	"fmt"
	"os"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/core"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/parser"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/scanner"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/tspath"
)

type BundleResult struct {
}

// 读取文件内容
func ReadFileContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// 解析TypeScript文件为AST
func ParseTypeScriptFile(filePath string, sourceCode string) *ast.SourceFile {
	// 创建路径对象
	path := tspath.Path(filePath)

	// 使用ParseSourceFile函数解析源代码
	sourceFile := parser.ParseSourceFile(
		filePath,
		path,
		sourceCode,
		core.ScriptTargetES2015,
		scanner.JSDocParsingModeParseAll,
	)

	return sourceFile
}

func Traverse(filePath string) {
	sourceCode, err := ReadFileContent(filePath)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
	}

	sourceFile := ParseTypeScriptFile(filePath, sourceCode)

	for _, node := range sourceFile.Statements.Nodes {
		// 解析 import
		if node.Kind == ast.KindImportDeclaration {
			idr := NewImportDeclarationResult()
			idr.analyzeImportDeclaration(node.AsImportDeclaration(), sourceCode)
		}

		// 解析 interface
		if node.Kind == ast.KindInterfaceDeclaration {
			inter := NewCusInterfaceDeclaration(node.AsNode(), sourceCode)
			inter.analyzeInterfaces(node.AsInterfaceDeclaration())

			fmt.Printf("\n分析接口: %s\n", inter.Name)
			for _, ref := range inter.Reference {
				if ref.IsExtend {
					fmt.Printf("- %s【继承】", ref.Name)
				} else {
					fmt.Printf("- %s 在", ref.Name)
					for _, location := range ref.Location {
						fmt.Printf("%s, ", location)
					}
				}
				fmt.Printf("\n")
			}
		}

		// // 解析 type
		// if node.Kind == ast.KindTypeAliasDeclaration {
		// 	fmt.Printf("Type: %s\n", node.Kind, node.AsTypeAliasDeclaration())
		// }

	}
}
