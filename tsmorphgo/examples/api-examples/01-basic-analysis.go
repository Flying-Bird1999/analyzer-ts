//go:build example01

package main

import (
	"fmt"
	"os"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run 01-basic-analysis.go <TypeScript项目路径>")
		os.Exit(1)
	}

	projectPath := os.Args[1]

	fmt.Println("🔍 基础分析示例 - 项目解析和 AST 遍历")
	fmt.Println("==================================================")

	// 1. 创建项目配置
	config := tsmorphgo.ProjectConfig{
		RootPath:         projectPath,
		IgnorePatterns:   []string{"node_modules", "dist", "build", ".git"},
		IsMonorepo:       false,
		TargetExtensions: []string{".ts", ".tsx"},
	}

	// 2. 初始化项目
	project := tsmorphgo.NewProject(config)

	// 3. 获取所有源文件
	sourceFiles := project.GetSourceFiles()
	fmt.Printf("✅ 发现 %d 个 TypeScript 文件\n", len(sourceFiles))

	// 4. 遍历所有文件，分析基本结构
	var (
		interfaceCount  int
		typeAliasCount  int
		functionCount   int
		classCount      int
		variableCount   int
		importCount     int
	)

	for _, sf := range sourceFiles {
		fmt.Printf("\n📄 分析文件: %s\n", sf.GetFilePath())

		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			switch node.Kind {
			case ast.KindInterfaceDeclaration:
				interfaceCount++
				if interfaceCount <= 3 { // 只打印前 3 个作为示例
					fmt.Printf("  🔷 接口: %s (行: %d)\n", node.GetText(), node.GetStartLineNumber())
				}

			case ast.KindTypeAliasDeclaration:
				typeAliasCount++

			case ast.KindFunctionDeclaration:
				functionCount++

			case ast.KindClassDeclaration:
				classCount++

			case ast.KindVariableDeclaration:
				variableCount++

			case ast.KindImportDeclaration:
				importCount++
			}
		})
	}

	// 5. 打印统计信息
	fmt.Println("\n📊 项目统计摘要:")
	fmt.Printf("  📋 接口数量: %d\n", interfaceCount)
	fmt.Printf("  🏷️  类型别名: %d\n", typeAliasCount)
	fmt.Printf("  ⚡ 函数数量: %d\n", functionCount)
	fmt.Printf("  🏗️  类数量: %d\n", classCount)
	fmt.Printf("  📦 变量数量: %d\n", variableCount)
	fmt.Printf("  📄 总文件数: %d\n", len(sourceFiles))

	fmt.Println("\n✅ 基础分析完成！")
}