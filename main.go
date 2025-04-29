package main

import (
	"fmt"
	"path/filepath"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/core"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/parser"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/scanner"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/tspath"
)

func main() {
	// 示例TypeScript代码
	sourceText := `  
		import { add } from './math';
		
		/**
		 * 这是一个示例函数
		 * @param name 要问候的名字
		 * @returns 问候信息
		 */
		function greet(name: string): string {  
				return "Hello, " + name;  
		}  

		interface Person {
			name: string;
		}
			
		const message = greet("World");  
		console.log(message);  
	`

	// 文件名和路径
	fileName := "example.ts"
	path := tspath.Path("")

	absFileName, err := filepath.Abs(fileName)

	if err != nil {
		fmt.Println("err")
	}

	// 使用ParseSourceFile函数解析源代码
	sourceFile := parser.ParseSourceFile(
		absFileName,
		path,
		sourceText,
		core.ScriptTargetES2015,
		scanner.JSDocParsingModeParseAll,
	)

	// 打印AST的基本信息
	fmt.Printf("解析成功！\n")
	fmt.Printf("文件名: %s\n", sourceFile.FileName)
	fmt.Printf("语句数量: %d\n", len(sourceFile.Statements.Nodes))

	// 打印每个顶层语句的类型
	fmt.Println("\n顶层语句:")
	for i, stmt := range sourceFile.Statements.Nodes {
		// fmt.Printf("  %d. 类型: %s\n", i+1, ast.KindToString(stmt.Kind))
		fmt.Printf("  %d. 类型: %s\n", i+1, stmt.Kind)
	}

	// 检查是否有解析错误
	diagnostics := sourceFile.Diagnostics()
	if len(diagnostics) > 0 {
		fmt.Println("\n解析错误:")
		for _, diag := range diagnostics {
			fmt.Printf("  - %s\n", diag.Message())
		}
	} else {
		fmt.Println("\n没有解析错误")
	}
}
