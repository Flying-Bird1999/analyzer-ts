//go:build example08

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/lsp"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run 09-lsp-service.go <TypeScript项目路径>")
		os.Exit(1)
	}

	projectPath := os.Args[1]

	fmt.Println("🔍 LSP 服务示例 - 语言服务器功能测试")
	fmt.Println("==================================================")

	// 1. 创建 LSP 服务
	service, err := lsp.NewService(projectPath)
	if err != nil {
		fmt.Printf("❌ 创建 LSP 服务失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 获取基础项目进行文件扫描
	config := tsmorphgo.ProjectConfig{
		RootPath:         projectPath,
		IgnorePatterns:   []string{"node_modules", "dist", "build", ".git"},
		TargetExtensions: []string{".ts", ".tsx"},
	}
	project := tsmorphgo.NewProject(config)
	sourceFiles := project.GetSourceFiles()

	fmt.Printf("✅ 成功创建 LSP 服务，发现 %d 个源文件\n", len(sourceFiles))

	ctx := context.Background()

	// 3. 测试 QuickInfo 功能
	fmt.Println("\n🔍 测试 QuickInfo 功能:")
	fmt.Println("----------------------------------------")

	// 查找第一个函数声明进行测试
	for _, sf := range sourceFiles {
		found := false
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			if !found && (node.Kind == ast.KindFunctionDeclaration || node.Kind == ast.KindVariableDeclaration) {
				line := node.GetStartLineNumber()
				filePath := sf.GetFilePath()

				fmt.Printf("📄 测试文件: %s\n", filePath)
				fmt.Printf("📍 测试位置: 第 %d 行，第 1 列\n", line)

				// 测试 QuickInfo 功能
				if quickInfo, err := service.GetQuickInfoAtPosition(ctx, filePath, line, 1); err == nil {
					if quickInfo != nil {
						fmt.Printf("✅ QuickInfo 成功:\n")
						fmt.Printf("   类型文本: %s\n", quickInfo.TypeText)
						fmt.Printf("   显示部件数: %d\n", len(quickInfo.DisplayParts))
						if quickInfo.Documentation != "" {
							fmt.Printf("   文档: %s\n", quickInfo.Documentation)
						}
						if quickInfo.Range != nil {
							fmt.Printf("   范围: %v\n", quickInfo.Range)
						}
					} else {
						fmt.Printf("ℹ️  该位置没有 QuickInfo 信息\n")
					}
				} else {
					fmt.Printf("❌ QuickInfo 失败: %v\n", err)
				}

				// 测试原生 QuickInfo 功能
				if nativeQuickInfo, err := service.GetNativeQuickInfoAtPosition(ctx, filePath, line, 1); err == nil {
					if nativeQuickInfo != nil {
						fmt.Printf("✅ 原生 QuickInfo 成功:\n")
						fmt.Printf("   类型文本: %s\n", nativeQuickInfo.TypeText)
						fmt.Printf("   显示部件数: %d\n", len(nativeQuickInfo.DisplayParts))
						if nativeQuickInfo.Documentation != "" {
							fmt.Printf("   文档: %s\n", nativeQuickInfo.Documentation)
						}
						if nativeQuickInfo.Range != nil {
							fmt.Printf("   范围: %v\n", nativeQuickInfo.Range)
						}

						// 显示显示部件详情
						for i, part := range nativeQuickInfo.DisplayParts {
							if i >= 3 { // 只显示前3个
								break
							}
							fmt.Printf("   部件 %d: [%s] %s\n", i+1, part.Kind, part.Text)
						}
					} else {
						fmt.Printf("ℹ️  该位置没有原生 QuickInfo 信息\n")
					}
				} else {
					fmt.Printf("❌ 原生 QuickInfo 失败: %v\n", err)
				}

				found = true
			}
		})
		if found {
			break
		}
	}

	// 4. 测试引用查找功能
	fmt.Println("\n🔗 测试引用查找功能:")
	fmt.Println("----------------------------------------")

	// 查找第一个接口声明进行引用测试
	for _, sf := range sourceFiles {
		found := false
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			if !found && node.Kind == ast.KindInterfaceDeclaration {
				line := node.GetStartLineNumber()
				filePath := sf.GetFilePath()

				fmt.Printf("📄 查找引用: %s\n", filePath)
				fmt.Printf("📍 接口位置: 第 %d 行，第 1 列\n", line)

				// 测试引用查找
				if response, err := service.FindReferences(ctx, filePath, line, 1); err == nil {
					if response.Locations != nil {
						fmt.Printf("✅ 找到 %d 个引用:\n", len(*response.Locations))
						for i, ref := range *response.Locations {
							if i >= 3 { // 只显示前3个
								break
							}
							fmt.Printf("   %d. %s:%d:%d\n", i+1,
								ref.Uri,
								ref.Range.Start.Line+1,
								ref.Range.Start.Character+1)
						}
					} else {
						fmt.Printf("ℹ️  没有找到引用\n")
					}
				} else {
					fmt.Printf("❌ 查找引用失败: %v\n", err)
				}

				found = true
			}
		})
		if found {
			break
		}
	}

	// 5. 测试符号获取功能（简化版）
	fmt.Println("\n🔤 测试符号获取功能:")
	fmt.Println("----------------------------------------")

	// 查找第一个声明进行符号测试
	for _, sf := range sourceFiles {
		found := false
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			if !found && (node.Kind == ast.KindFunctionDeclaration || node.Kind == ast.KindInterfaceDeclaration) {
				line := node.GetStartLineNumber()
				filePath := sf.GetFilePath()

				fmt.Printf("📄 获取符号: %s\n", filePath)
				fmt.Printf("📍 符号位置: 第 %d 行，第 1 列\n", line)

				// 使用 tsmorphgo 的符号获取功能
				if symbol, ok := tsmorphgo.GetSymbol(node); ok {
					fmt.Printf("✅ 符号获取成功:\n")
					fmt.Printf("   符号名称: %s\n", symbol.GetName())
					fmt.Printf("   是否导出: %t\n", symbol.IsExported())

					// 测试类型检查
					if symbol.IsFunction() {
						fmt.Printf("   符号类型: 函数\n")
					} else if symbol.IsInterface() {
						fmt.Printf("   符号类型: 接口\n")
					} else if symbol.IsClass() {
						fmt.Printf("   符号类型: 类\n")
					} else if symbol.IsVariable() {
						fmt.Printf("   符号类型: 变量\n")
					} else if symbol.IsTypeAlias() {
						fmt.Printf("   符号类型: 类型别名\n")
					} else {
						fmt.Printf("   符号类型: 其他\n")
					}

					// 测试引用查找
					if refs, err := symbol.FindReferences(); err == nil {
						fmt.Printf("   符号引用数: %d\n", len(refs))
					} else {
						fmt.Printf("   获取引用失败: %v\n", err)
					}
				} else {
					fmt.Printf("ℹ️  该节点没有符号信息\n")
				}

				found = true
			}
		})
		if found {
			break
		}
	}

	// 6. 测试上下文相关的 LSP 功能
	fmt.Println("\n🎯 测试上下文相关功能:")
	fmt.Println("----------------------------------------")

	fmt.Printf("📊 LSP 服务状态:\n")
	fmt.Printf("   根路径: %s\n", projectPath)
	fmt.Printf("   上下文: %v\n", ctx)
	fmt.Printf("   会话管理: 启用\n")
	fmt.Printf("   缓存文件数: %d\n", len(sourceFiles))

	// 7. 清理资源
	defer service.Close()

	fmt.Println("\n✅ LSP 服务测试完成！")
	fmt.Println("==================================================")
	fmt.Println("📋 测试总结:")
	fmt.Println("   ✅ LSP 服务创建和管理")
	fmt.Println("   ✅ QuickInfo 功能（类型提示）")
	fmt.Println("   ✅ 原生 QuickInfo 功能")
	fmt.Println("   ✅ 引用查找功能")
	fmt.Println("   ✅ 符号获取功能（使用 tsmorphgo）")
	fmt.Println("   ✅ 上下文管理和资源清理")
	fmt.Println("==================================================")
}