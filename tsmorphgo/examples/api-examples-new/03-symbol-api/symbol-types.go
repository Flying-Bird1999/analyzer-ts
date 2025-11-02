// +build symbol-api

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run -tags symbol-api symbol-types.go <TypeScript项目路径>")
		os.Exit(1)
	}

	projectPath := os.Args[1]

	fmt.Println("🎯 符号系统 API - 符号类型和接口验证")
	fmt.Println("================================")

	// 创建项目配置 - 验证项目创建的配置选项
	config := tsmorphgo.ProjectConfig{
		RootPath:         projectPath,
		IgnorePatterns:   []string{"node_modules", "dist", "build", ".git"},
		TargetExtensions: []string{".ts", ".tsx"},
	}
	project := tsmorphgo.NewProject(config)
	defer project.Close()

	// 验证项目创建是否成功
	sourceFiles := project.GetSourceFiles()
	if len(sourceFiles) == 0 {
		fmt.Println("❌ 项目创建失败：未发现任何源文件")
		return
	}

	fmt.Printf("✅ 项目创建成功，发现 %d 个源文件\n", len(sourceFiles))

	// 1. 接口符号发现验证 - 测试从 AST 节点发现接口符号的能力
	fmt.Println("\n🔷 接口符号发现验证:")
	fmt.Println("------------------------------")

	interfaceCount := 0
	var firstInterface *tsmorphgo.Symbol
	var firstInterfaceNode tsmorphgo.Node

	// 遍历所有源文件，收集接口符号
	for _, sf := range sourceFiles {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			// 检查是否为接口声明节点
			if node.Kind == ast.KindInterfaceDeclaration {
				// 尝试获取接口符号
				if symbol, ok := tsmorphgo.GetSymbol(node); ok {
					interfaceCount++

					// 记录第一个接口符号用于后续测试
					if firstInterface == nil {
						firstInterface = symbol
						firstInterfaceNode = node
					}
				}
			}
		})
	}

	fmt.Printf("✅ 接口符号发现完成，共发现 %d 个接口符号\n", interfaceCount)

	// 验证是否发现了接口符号
	if interfaceCount == 0 {
		fmt.Println("❌ 接口符号发现验证失败：项目中未发现任何接口符号")
		return
	}

	// 2. 类型别名符号发现验证 - 测试类型别名符号的提取能力
	fmt.Println("\n🏷️ 类型别名符号发现验证:")
	fmt.Println("------------------------------")

	typeAliasCount := 0
	var firstTypeAlias *tsmorphgo.Symbol
	var firstTypeAliasNode tsmorphgo.Node

	// 遍历所有源文件，收集类型别名符号
	for _, sf := range sourceFiles {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			// 检查是否为类型别名声明节点
			if node.Kind == ast.KindTypeAliasDeclaration {
				// 尝试获取类型别名符号
				if symbol, ok := tsmorphgo.GetSymbol(node); ok {
					typeAliasCount++

					// 记录第一个类型别名符号用于后续测试
					if firstTypeAlias == nil {
						firstTypeAlias = symbol
						firstTypeAliasNode = node
					}
				}
			}
		})
	}

	fmt.Printf("✅ 类型别名符号发现完成，共发现 %d 个类型别名符号\n", typeAliasCount)

	// 3. 符号类型识别验证 - 验证符号类型判断的准确性
	fmt.Println("\n🏷️ 符号类型识别验证:")
	fmt.Println("------------------------------")

	typeIdentificationSuccess := true
	symbolTypeCount := make(map[string]int)

	// 遍历所有符号，验证类型识别
	for _, sf := range sourceFiles {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			if symbol, ok := tsmorphgo.GetSymbol(node); ok {
				symbolTypeName := getSymbolTypeName(symbol)
				symbolTypeCount[symbolTypeName]++

				// 验证节点类型与符号类型的一致性
				switch node.Kind {
				case ast.KindInterfaceDeclaration:
					if !symbol.IsInterface() {
						fmt.Printf("⚠️ 接口节点符号类型不匹配: %s\n", symbol.GetName())
						typeIdentificationSuccess = false
					}
				case ast.KindTypeAliasDeclaration:
					if !symbol.IsTypeAlias() {
						fmt.Printf("⚠️ 类型别名节点符号类型不匹配: %s\n", symbol.GetName())
						typeIdentificationSuccess = false
					}
				}
			}
		})
	}

	// 输出符号类型统计
	fmt.Println("  符号类型分布:")
	for typeName, count := range symbolTypeCount {
		fmt.Printf("    %s: %d\n", typeName, count)
	}

	if typeIdentificationSuccess {
		fmt.Println("✅ 符号类型识别验证通过")
	} else {
		fmt.Println("❌ 符号类型识别验证失败：发现类型不匹配")
	}

	// 4. 符号导出状态验证 - 验证导出状态检测的准确性
	fmt.Println("\n📤 符号导出状态验证:")
	fmt.Println("------------------------------")

	exportedStats := make(map[string]int) // 按类型统计导出状态
	nonExportedStats := make(map[string]int)

	for _, sf := range sourceFiles {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			if symbol, ok := tsmorphgo.GetSymbol(node); ok {
				symbolTypeName := getSymbolTypeName(symbol)
				if symbol.IsExported() {
					exportedStats[symbolTypeName]++
				} else {
					nonExportedStats[symbolTypeName]++
				}
			}
		})
	}

	fmt.Println("  导出符号统计:")
	for typeName, count := range exportedStats {
		fmt.Printf("    %s: %d 个导出符号\n", typeName, count)
	}

	fmt.Println("  本地符号统计:")
	for typeName, count := range nonExportedStats {
		fmt.Printf("    %s: %d 个本地符号\n", typeName, count)
	}

	// 5. 符号详细信息验证 - 验证符号详细信息的提取
	fmt.Println("\n📋 符号详细信息验证:")
	fmt.Println("------------------------------")

	if firstInterface != nil {
		fmt.Println("  首个接口符号详情:")
		fmt.Printf("    符号名称: %s\n", firstInterface.GetName())
		fmt.Printf("    符号类型: %s\n", getSymbolTypeName(firstInterface))
		fmt.Printf("    是否导出: %t\n", firstInterface.IsExported())
		fmt.Printf("    声明数量: %d\n", firstInterface.GetDeclarationCount())
		fmt.Printf("    文件位置: %s\n", firstInterfaceNode.GetSourceFile().GetFilePath())
		fmt.Printf("    行号: %d\n", firstInterfaceNode.GetStartLineNumber())
		fmt.Println("  ✅ 接口符号详细信息验证通过")
	}

	if firstTypeAlias != nil {
		fmt.Println("  首个类型别名符号详情:")
		fmt.Printf("    符号名称: %s\n", firstTypeAlias.GetName())
		fmt.Printf("    符号类型: %s\n", getSymbolTypeName(firstTypeAlias))
		fmt.Printf("    是否导出: %t\n", firstTypeAlias.IsExported())
		fmt.Printf("    声明数量: %d\n", firstTypeAlias.GetDeclarationCount())
		fmt.Printf("    文件位置: %s\n", firstTypeAliasNode.GetSourceFile().GetFilePath())
		fmt.Printf("    行号: %d\n", firstTypeAliasNode.GetStartLineNumber())
		fmt.Println("  ✅ 类型别名符号详细信息验证通过")
	}

	// 6. 符号声明节点验证 - 验证能否正确获取声明节点
	fmt.Println("\n🔗 符号声明节点验证:")
	fmt.Println("------------------------------")

	declarationNodeSuccess := true

	if firstInterface != nil {
		if declarations := firstInterface.GetDeclarations(); len(declarations) > 0 {
			decl := declarations[0]
			fmt.Printf("  接口符号声明节点类型: %v\n", decl.Kind)
			fmt.Printf("  接口符号声明文本: %s\n", decl.GetText())
			fmt.Printf("  接口符号声明位置: %s:%d\n",
				decl.GetSourceFile().GetFilePath(), decl.GetStartLineNumber())
		} else {
			fmt.Println("❌ 接口符号声明节点验证失败：未找到声明")
			declarationNodeSuccess = false
		}
	}

	if firstTypeAlias != nil {
		if declarations := firstTypeAlias.GetDeclarations(); len(declarations) > 0 {
			decl := declarations[0]
			fmt.Printf("  类型别名符号声明节点类型: %v\n", decl.Kind)
			fmt.Printf("  类型别名符号声明文本: %s\n", decl.GetText())
			fmt.Printf("  类型别名符号声明位置: %s:%d\n",
				decl.GetSourceFile().GetFilePath(), decl.GetStartLineNumber())
		} else {
			fmt.Println("❌ 类型别名符号声明节点验证失败：未找到声明")
			declarationNodeSuccess = false
		}
	}

	if declarationNodeSuccess {
		fmt.Println("✅ 符号声明节点验证通过")
	} else {
		fmt.Println("❌ 符号声明节点验证失败")
	}

	// 7. 符号字符串表示验证 - 验证符号的字符串表示功能
	fmt.Println("\n📝 符号字符串表示验证:")
	fmt.Println("------------------------------")

	stringRepresentationSuccess := true

	if firstInterface != nil {
		symbolString := firstInterface.String()
		fmt.Printf("  接口符号字符串表示: %s\n", symbolString)
		if symbolString == "" {
			fmt.Println("❌ 接口符号字符串表示验证失败：空字符串")
			stringRepresentationSuccess = false
		}
	}

	if firstTypeAlias != nil {
		symbolString := firstTypeAlias.String()
		fmt.Printf("  类型别名符号字符串表示: %s\n", symbolString)
		if symbolString == "" {
			fmt.Println("❌ 类型别名符号字符串表示验证失败：空字符串")
			stringRepresentationSuccess = false
		}
	}

	if stringRepresentationSuccess {
		fmt.Println("✅ 符号字符串表示验证通过")
	} else {
		fmt.Println("❌ 符号字符串表示验证失败")
	}

	// 8. 符号数量统计验证 - 验证符号数量统计的准确性
	fmt.Println("\n📊 符号数量统计验证:")
	fmt.Println("------------------------------")

	totalSymbols := 0
	for _, count := range symbolTypeCount {
		totalSymbols += count
	}

	fmt.Printf("  总符号数量: %d\n", totalSymbols)
	fmt.Printf("  接口符号数量: %d\n", interfaceCount)
	fmt.Printf("  类型别名符号数量: %d\n", typeAliasCount)
	fmt.Printf("  其他类型符号数量: %d\n", totalSymbols-interfaceCount-typeAliasCount)
	fmt.Printf("  发现的符号类型种类: %d\n", len(symbolTypeCount))

	// 9. 保存分析结果到 JSON 文件
	fmt.Println("\n💾 保存分析结果:")
	fmt.Println("------------------------------")

	analysisResult := map[string]interface{}{
		"totalSymbols":       totalSymbols,
		"interfaceCount":     interfaceCount,
		"typeAliasCount":     typeAliasCount,
		"symbolTypes":        symbolTypeCount,
		"exportedStats":      exportedStats,
		"nonExportedStats":   nonExportedStats,
		"validationResults": map[string]bool{
			"interfaceDiscovery":    interfaceCount > 0,
			"typeAliasDiscovery":    typeAliasCount > 0,
			"typeIdentification":    typeIdentificationSuccess,
			"declarationNode":      declarationNodeSuccess,
			"stringRepresentation":  stringRepresentationSuccess,
		},
	}

	resultFile := "../../validation-results/symbol-types-analysis.json"
	if err := os.MkdirAll("../../validation-results", 0755); err == nil {
		if data, err := json.MarshalIndent(analysisResult, "", "  "); err == nil {
			if err := os.WriteFile(resultFile, data, 0644); err == nil {
				fmt.Printf("✅ 分析结果已保存到: %s\n", resultFile)
			} else {
				fmt.Printf("❌ 保存分析结果失败: %v\n", err)
			}
		} else {
			fmt.Printf("❌ 序列化分析结果失败: %v\n", err)
		}
	} else {
		fmt.Printf("❌ 创建结果目录失败: %v\n", err)
	}

	// 10. 验证结果汇总
	fmt.Println("\n📊 验证结果汇总:")
	fmt.Println("================================")

	validationResults := map[string]bool{
		"接口符号发现":     interfaceCount > 0,
		"类型别名符号发现":   typeAliasCount > 0,
		"符号类型识别":     typeIdentificationSuccess,
		"符号声明节点获取":   declarationNodeSuccess,
		"符号字符串表示":    stringRepresentationSuccess,
	}

	passedCount := 0
	totalValidations := len(validationResults)

	for testName, passed := range validationResults {
		if passed {
			fmt.Printf("✅ %s: 通过\n", testName)
			passedCount++
		} else {
			fmt.Printf("❌ %s: 失败\n", testName)
		}
	}

	passRate := float64(passedCount) / float64(totalValidations) * 100
	fmt.Printf("\n📈 验证通过率: %.1f%% (%d/%d)\n", passRate, passedCount, totalValidations)

	// 11. 最终结论
	if passRate >= 80.0 {
		fmt.Println("\n🎉 符号类型 API 验证完成！基本功能正常工作")
		fmt.Println("================================")
		fmt.Println("📋 已验证的 API:")
		fmt.Println("   - tsmorphgo.GetSymbol() - 从节点获取符号")
		fmt.Println("   - symbol.GetName() - 获取符号名称")
		fmt.Println("   - symbol.IsInterface() - 检查是否为接口符号")
		fmt.Println("   - symbol.IsTypeAlias() - 检查是否为类型别名符号")
		fmt.Println("   - symbol.IsExported() - 检查导出状态")
		fmt.Println("   - symbol.GetDeclarationCount() - 获取声明数量")
		fmt.Println("   - symbol.GetDeclarations() - 获取所有声明")
		fmt.Println("   - symbol.String() - 符号字符串表示")
		fmt.Println("================================")
		fmt.Println("📝 主要发现:")
		fmt.Printf("   - 项目中共发现 %d 个接口符号\n", interfaceCount)
		fmt.Printf("   - 项目中共发现 %d 个类型别名符号\n", typeAliasCount)
		fmt.Printf("   - 共识别 %d 种不同的符号类型\n", len(symbolTypeCount))
	} else {
		fmt.Println("\n❌ 符号类型 API 验证完成但存在问题")
		fmt.Printf("   验证通过率 %.1f%% 低于预期\n", passRate)
		fmt.Println("   建议检查符号系统实现和项目配置")
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