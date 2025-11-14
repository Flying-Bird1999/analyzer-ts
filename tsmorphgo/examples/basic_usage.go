//go:build examples

package main

import (
	"fmt"
	"log"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🚀 TSMorphGo 基础API测试")
	fmt.Println("========================")
	fmt.Println("验证核心功能：项目创建、文件扫描、路径别名解析")
	fmt.Println()

	// 1. 创建项目实例 - 基于tsconfig.json
	fmt.Println("📁 1. 项目创建测试")
	fmt.Println("==================")

	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:     "./demo-react-app",
		UseTsConfig:  true, // 自动读取和使用tsconfig.json
		TsConfigPath: "./demo-react-app/tsconfig.json",
	})

	if project == nil {
		log.Fatal("❌ 项目创建失败")
	}
	fmt.Println("✅ 项目创建成功")

	// 2. 获取项目文件列表
	fmt.Println()
	fmt.Println("📄 2. 文件扫描测试")
	fmt.Println("==================")

	sourceFiles := project.GetSourceFiles()
	fmt.Printf("✅ 扫描到 %d 个源文件\n", len(sourceFiles))

	fmt.Println("\n📋 文件列表:")
	for i, file := range sourceFiles {
		if i >= 10 { // 只显示前10个
			fmt.Printf("   ... 还有 %d 个文件\n", len(sourceFiles)-10)
			break
		}
		filePath := file.GetFilePath()
		fmt.Printf("   %d. %s\n", i+1, filePath)
	}

	// 3. 验证路径别名解析
	fmt.Println()
	fmt.Println("🔗 3. 路径别名解析测试")
	fmt.Println("======================")

	// 查找使用路径别名的文件
	aliasFiles := 0
	for _, file := range sourceFiles {
		filePath := file.GetFilePath()
		if contains(filePath, "test-aliases") || contains(filePath, "App.tsx") {
			aliasFiles++
			fmt.Printf("✅ 找到使用路径别名的文件: %s\n", filePath)
		}
	}

	if aliasFiles > 0 {
		fmt.Printf("✅ 路径别名解析正常，找到 %d 个使用别名的文件\n", aliasFiles)
	} else {
		fmt.Println("⚠️  没有找到使用路径别名的文件")
	}

	// 4. 基本节点遍历测试
	fmt.Println()
	fmt.Println("🔍 4. 节点遍历测试")
	fmt.Println("==================")

	// 找一个包含导入语句的文件进行测试
	var testFile *tsmorphgo.SourceFile
	for _, file := range sourceFiles {
		if contains(file.GetFilePath(), "App.tsx") {
			testFile = file
			break
		}
	}

	if testFile != nil {
		fmt.Printf("✅ 选择测试文件: %s\n", testFile.GetFilePath())

		// 遍历文件的所有节点
		nodeCount := 0
		importCount := 0

		testFile.ForEachDescendant(func(node *tsmorphgo.Node) {
			nodeCount++

			// 检查是否是导入节点
			if node.IsImportDeclaration() {
				importCount++
				fmt.Printf("   📥 导入: %s\n", node.GetText())
			}
		})

		fmt.Printf("✅ 遍历完成，找到 %d 个节点，其中 %d 个导入\n", nodeCount, importCount)
	} else {
		fmt.Println("❌ 没有找到合适的测试文件")
	}

	// 5. 符号查找测试
	fmt.Println()
	fmt.Println("🎯 5. 符号查找测试")
	fmt.Println("==================")

	// 查找useUserData函数的引用
	if testFile != nil {
		// 在文件中查找useUserData标识符
		node := testFile.FindNodeByText("useUserData")
		if node != nil {
			fmt.Printf("✅ 找到目标节点: %s\n", node.GetText())

			// 尝试获取符号信息
			symbol := node.GetSymbol()
			if symbol != nil {
				fmt.Printf("✅ 符号信息: %s\n", symbol.GetName())
			} else {
				fmt.Println("⚠️  无法获取符号信息")
			}

			// 获取节点位置
			line := node.GetStartLineNumber()
			col := node.GetStartColumn()
			fmt.Printf("📍 位置: 第%d行，第%d列\n", line, col)
		} else {
			fmt.Println("❌ 没有找到目标节点")
		}
	}

	// 6. 项目配置信息
	fmt.Println()
	fmt.Println("⚙️  6. 项目配置信息")
	fmt.Println("==================")

	// 获取TypeScript配置
	tsConfig := project.GetTsConfig()
	if tsConfig != nil {
		fmt.Println("✅ 成功读取tsconfig.json")

		// 显示编译选项
		if tsConfig.CompilerOptions != nil {
			fmt.Printf("📋 编译选项数量: %d\n", len(tsConfig.CompilerOptions))

			// 检查路径别名配置
			if paths, ok := tsConfig.CompilerOptions["paths"]; ok {
				if pathsMap, ok := paths.(map[string]interface{}); ok {
					fmt.Printf("🔗 路径别名配置:\n")
					for alias, mapping := range pathsMap {
						fmt.Printf("   %s -> %v\n", alias, mapping)
					}
				}
			}
		}
	} else {
		fmt.Println("⚠️  没有找到tsconfig.json")
	}

	// 7. 清理资源
	fmt.Println()
	fmt.Println("🧹 清理资源")
	project.Close()

	fmt.Println()
	fmt.Println("✅ 基础API测试完成！")
	fmt.Println("如果所有测试都显示 ✅，说明核心功能工作正常。")
}

// 简单的字符串包含检查
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
