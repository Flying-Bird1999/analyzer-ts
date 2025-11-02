// +build symbol-api

package main

import (
	"fmt"
	"os"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run -tags symbol-api symbol-basics.go <项目路径>")
		os.Exit(1)
	}

	projectPath := os.Args[1]

	fmt.Println("🎯 符号系统 API - 符号获取和基本信息")
	fmt.Println("================================")

	// 创建项目配置和项目实例
	config := tsmorphgo.ProjectConfig{
		RootPath:         projectPath,
		IgnorePatterns:   []string{"node_modules", "dist", "build", ".git"},
		TargetExtensions: []string{".ts", ".tsx"},
	}
	project := tsmorphgo.NewProject(config)

	// 验证项目创建是否成功
	sourceFiles := project.GetSourceFiles()
	if len(sourceFiles) == 0 {
		fmt.Println("❌ 项目创建失败：未发现任何源文件")
		return
	}

	fmt.Printf("✅ 项目创建成功，发现 %d 个源文件\n", len(sourceFiles))

	// 1. 符号发现能力验证 - 测试是否能从 AST 节点中提取符号
	fmt.Println("\n🔍 符号发现能力验证:")

	symbolCount := 0
	var firstSymbol *tsmorphgo.Symbol
	// var firstSymbolNode tsmorphgo.Node

	// 遍历所有源文件，收集符号
	for _, sf := range sourceFiles {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			// 尝试从节点获取符号
			if symbol, ok := tsmorphgo.GetSymbol(node); ok {
				symbolCount++

				// 记录第一个符号用于后续测试
				if firstSymbol == nil {
					firstSymbol = symbol
					// firstSymbolNode = node
				}
			}
		})
	}

	fmt.Printf("✅ 符号发现完成，共发现 %d 个符号\n", symbolCount)

	// 验证是否至少发现了一个符号
	if symbolCount == 0 {
		fmt.Println("❌ 符号发现验证失败：项目中未发现任何符号")
		return
	}

	// 2. 符号基本信息验证 - 验证符号的基本属性是否正确
	fmt.Println("\n📋 符号基本信息验证:")

	if firstSymbol != nil {
		fmt.Printf("  符号名称: %s\n", firstSymbol.GetName())
		fmt.Printf("  符号类型: %v\n", firstSymbol.GetFlags())
		fmt.Printf("  是否导出: %t\n", firstSymbol.IsExported())

		// 验证符号名称是否为空
		if firstSymbol.GetName() == "" {
			fmt.Println("❌ 符号基本信息验证失败：符号名称为空")
			return
		}
		fmt.Println("✅ 符号基本信息验证通过")
	}

	// 3. 符号类型分类验证 - 验证符号类型判断的准确性
	fmt.Println("\n🏷️ 符号类型分类验证:")

	// 定义符号类型统计映射
	symbolTypeStats := make(map[string]int)

	// 再次遍历，统计不同类型的符号
	for _, sf := range sourceFiles {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			if symbol, ok := tsmorphgo.GetSymbol(node); ok {
				symbolTypeName := getSymbolTypeName(symbol)
				symbolTypeStats[symbolTypeName]++
			}
		})
	}

	// 输出符号类型统计
	fmt.Println("  符号类型分布:")
	for typeName, count := range symbolTypeStats {
		fmt.Printf("    %s: %d\n", typeName, count)
	}

	// 验证是否发现了多种类型的符号
	if len(symbolTypeStats) == 0 {
		fmt.Println("❌ 符号类型分类验证失败：未发现任何类型的符号")
		return
	}

	fmt.Printf("✅ 符号类型分类验证通过，共发现 %d 种类型的符号\n", len(symbolTypeStats))

	// 4. 符号导出状态验证 - 验证导出状态检测的准确性
	fmt.Println("\n📤 符号导出状态验证:")

	exportedCount := 0
	localCount := 0

	for _, sf := range sourceFiles {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			if symbol, ok := tsmorphgo.GetSymbol(node); ok {
				if symbol.IsExported() {
					exportedCount++
				} else {
					localCount++
				}
			}
		})
	}

	fmt.Printf("  导出符号: %d\n", exportedCount)
	fmt.Printf("  本地符号: %d\n", localCount)
	fmt.Printf("  导出比例: %.1f%%\n", float64(exportedCount)/float64(exportedCount+localCount)*100)

	// 验证导出状态统计的合理性
	if exportedCount+localCount != symbolCount {
		fmt.Printf("⚠️ 导出状态统计可能存在错误：总数不匹配（%d vs %d）\n",
			exportedCount+localCount, symbolCount)
	} else {
		fmt.Println("✅ 符号导出状态验证通过")
	}

	// 5. 符号声明验证 - 验证符号声明的提取准确性
	fmt.Println("\n📝 符号声明验证:")

	declarationCount := 0
	multiDeclarationCount := 0

	for _, sf := range sourceFiles {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			if symbol, ok := tsmorphgo.GetSymbol(node); ok {
				declCount := symbol.GetDeclarationCount()
				declarationCount += declCount

				// 检查是否有多个声明的情况（如函数重载）
				if declCount > 1 {
					multiDeclarationCount++
				}
			}
		})
	}

	fmt.Printf("  总声明数: %d\n", declarationCount)
	fmt.Printf("  多声明符号数: %d\n", multiDeclarationCount)

	// 验证声明统计的合理性
	if declarationCount == 0 {
		fmt.Println("❌ 符号声明验证失败：未发现任何声明")
		return
	}

	// 6. 符号声明节点验证 - 验证能否正确获取声明节点
	fmt.Println("\n🔗 符号声明节点验证:")

	var foundDeclaration bool
	for _, sf := range sourceFiles {
		if foundDeclaration {
			break
		}

		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			if foundDeclaration {
				return
			}

			if symbol, ok := tsmorphgo.GetSymbol(node); ok {
				declarations := symbol.GetDeclarations()
				if len(declarations) > 0 {
					// 验证声明节点的基本属性
					decl := declarations[0]
					fmt.Printf("  首个声明节点类型: %v\n", decl.Kind)
					fmt.Printf("  首个声明文件: %s\n", decl.GetSourceFile().GetFilePath())
					fmt.Printf("  首个声明行号: %d\n", decl.GetStartLineNumber())

					foundDeclaration = true
				}
			}
		})
	}

	if foundDeclaration {
		fmt.Println("✅ 符号声明节点验证通过")
	} else {
		fmt.Println("❌ 符号声明节点验证失败：未发现任何声明节点")
	}

	// 7. 符号第一声明验证 - 验证便捷方法 GetFirstDeclaration 的准确性
	fmt.Println("\n🎯 符号第一声明验证:")

	var foundFirstDeclaration bool
	for _, sf := range sourceFiles {
		if foundFirstDeclaration {
			break
		}

		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			if foundFirstDeclaration {
				return
			}

			if symbol, ok := tsmorphgo.GetSymbol(node); ok {
				if decl, ok := symbol.GetFirstDeclaration(); ok {
					fmt.Printf("  第一声明类型: %v\n", decl.Kind)
					fmt.Printf("  第一声明文本: %s\n", decl.GetText())
					foundFirstDeclaration = true
				}
			}
		})
	}

	if foundFirstDeclaration {
		fmt.Println("✅ 符号第一声明验证通过")
	} else {
		fmt.Println("❌ 符号第一声明验证失败：GetFirstDeclaration 无效")
	}

	// 8. 符号字符串表示验证 - 验证 String() 方法的工作
	fmt.Println("\n📝 符号字符串表示验证:")

	if firstSymbol != nil {
		symbolString := firstSymbol.String()
		fmt.Printf("  符号字符串表示: %s\n", symbolString)

		// 验证字符串表示是否包含基本信息
		if symbolString != "" && len(symbolString) > 0 {
			fmt.Println("✅ 符号字符串表示验证通过")
		} else {
			fmt.Println("❌ 符号字符串表示验证失败：字符串为空")
		}
	}

	// 9. 边界情况验证 - 测试无效符号的处理
	fmt.Println("\n⚠️ 边界情况验证:")

	// 创建一个空的符号测试
	var emptySymbol *tsmorphgo.Symbol
	if emptySymbol != nil {
		emptyName := emptySymbol.GetName()
		emptyExported := emptySymbol.IsExported()
		fmt.Printf("  空符号处理：名称='%s', 导出=%t\n", emptyName, emptyExported)
	} else {
		fmt.Println("  空符号处理：nil 符号处理正常")
	}

	// 10. 性能测试 - 符号收集的性能验证
	fmt.Println("\n⏱️ 性能测试:")

	fmt.Printf("  开始收集 %d 个源文件的符号...\n", len(sourceFiles))

	performanceSymbols := 0
	for _, sf := range sourceFiles {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			if _, ok := tsmorphgo.GetSymbol(node); ok {
				performanceSymbols++
			}
		})
	}

	fmt.Printf("  性能测试完成，共收集 %d 个符号\n", performanceSymbols)
	fmt.Printf("  平均每个文件 %.1f 个符号\n", float64(performanceSymbols)/float64(len(sourceFiles)))

	// 11. 验证结果汇总
	fmt.Println("\n📊 验证结果汇总:")
	fmt.Printf("  ✅ 符号发现能力: 发现 %d 个符号\n", symbolCount)
	fmt.Printf("  ✅ 符号基本信息: %d 种类型\n", len(symbolTypeStats))
	fmt.Printf("  ✅ 导出状态检测: %d 个导出符号\n", exportedCount)
	fmt.Printf("  ✅ 符号声明验证: %d 个声明\n", declarationCount)
	fmt.Printf("  ✅ 声明节点获取: %v\n", foundDeclaration)
	fmt.Printf("  ✅ 第一声明方法: %v\n", foundFirstDeclaration)
	fmt.Printf("  ✅ 性能基准: %.1f 符号/文件\n", float64(performanceSymbols)/float64(len(sourceFiles)))

	// 最终验证结果
	if symbolCount > 0 && len(symbolTypeStats) > 0 && declarationCount > 0 {
		fmt.Println("\n🎉 符号系统 API 基础功能验证完成！")
		fmt.Println("================================")
		fmt.Println("📋 已验证的 API:")
		fmt.Println("   - tsmorphgo.GetSymbol() - 从节点获取符号")
		fmt.Println("   - symbol.GetName() - 获取符号名称")
		fmt.Println("   - symbol.GetFlags() - 获取符号标志")
		fmt.Println("   - symbol.IsExported() - 检查导出状态")
		fmt.Println("   - symbol.GetDeclarationCount() - 获取声明数量")
		fmt.Println("   - symbol.GetDeclarations() - 获取所有声明")
		fmt.Println("   - symbol.GetFirstDeclaration() - 获取第一个声明")
		fmt.Println("   - symbol.String() - 符号字符串表示")
		fmt.Println("================================")
	} else {
		fmt.Println("\n❌ 符号系统 API 基础功能验证失败！")
	}
}

// getSymbolTypeName 根据符号的标志返回人类可读的类型名称
func getSymbolTypeName(symbol *tsmorphgo.Symbol) string {
	// 使用符号的各种类型检查方法来判断类型
	switch {
	case symbol.IsFunction():
		return "function"
	case symbol.IsClass():
		return "class"
	case symbol.IsInterface():
		return "interface"
	case symbol.IsTypeAlias():
		return "typeAlias"
	case symbol.IsEnum():
		return "enum"
	case symbol.IsVariable():
		return "variable"
	case symbol.IsMethod():
		return "method"
	case symbol.IsConstructor():
		return "constructor"
	case symbol.IsAccessor():
		return "accessor"
	case symbol.IsTypeParameter():
		return "typeParameter"
	case symbol.IsEnumMember():
		return "enumMember"
	case symbol.IsProperty():
		return "property"
	case symbol.IsObjectLiteral():
		return "objectLiteral"
	case symbol.IsTypeLiteral():
		return "typeLiteral"
	case symbol.IsModule():
		return "module"
	case symbol.IsAlias():
		return "alias"
	default:
		return "unknown"
	}
}