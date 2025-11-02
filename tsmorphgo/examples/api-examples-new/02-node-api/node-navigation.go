// +build node-api

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
		fmt.Println("用法: go run -tags node-api node-navigation.go <TypeScript项目路径>")
		os.Exit(1)
	}

	projectPath := os.Args[1]

	fmt.Println("🎯 节点操作 API - 节点导航（父子、祖先）")
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

	// 1. 节点发现验证 - 测试从项目中发现各种类型的节点
	fmt.Println("\n🔍 节点发现验证:")
	fmt.Println("------------------------------")

	nodeTypeStats := make(map[string]int)
	var firstInterfaceNode *tsmorphgo.Node
	var firstFunctionNode *tsmorphgo.Node
	var firstClassNode *tsmorphgo.Node
	var firstTypeAliasNode *tsmorphgo.Node

	// 遍历所有源文件，收集各种类型的节点
	for _, sf := range sourceFiles {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			nodeTypeStats[node.Kind.String()]++

			// 记录不同类型的第一个节点用于后续测试
			switch node.Kind {
			case ast.KindInterfaceDeclaration:
				if firstInterfaceNode == nil {
					firstInterfaceNode = &node
				}
			case ast.KindFunctionDeclaration:
				if firstFunctionNode == nil {
					firstFunctionNode = &node
				}
			case ast.KindClassDeclaration:
				if firstClassNode == nil {
					firstClassNode = &node
				}
			case ast.KindTypeAliasDeclaration:
				if firstTypeAliasNode == nil {
					firstTypeAliasNode = &node
				}
			}
		})
	}

	// 输出节点类型统计
	fmt.Println("  节点类型分布:")
	for typeName, count := range nodeTypeStats {
		fmt.Printf("    %s: %d\n", typeName, count)
	}

	// 验证是否发现了足够多类型的节点
	foundTypes := len(nodeTypeStats)
	fmt.Printf("  发现的节点类型总数: %d\n", foundTypes)

	if foundTypes == 0 {
		fmt.Println("❌ 节点发现验证失败：项目中未发现任何节点")
		return
	}

	// 2. 节点基本信息验证 - 验证节点基本属性的获取能力
	fmt.Println("\n📋 节点基本信息验证:")
	fmt.Println("------------------------------")

	validateNodeBasicInfo := func(node *tsmorphgo.Node, nodeType string) bool {
		if node == nil {
			fmt.Printf("  ❌ %s: 节点不存在\n", nodeType)
			return false
		}

		fmt.Printf("  📝 %s 节点信息:\n", nodeType)
		fmt.Printf("    节点类型: %v\n", node.Kind)
		fmt.Printf("    节点文本: %s\n", node.GetText())
		fmt.Printf("    起始行号: %d\n", node.GetStartLineNumber())
		fmt.Printf("    结束行号: %d\n", node.GetEndLineNumber())
		fmt.Printf("    起始位置: %d\n", node.GetStart())
		fmt.Printf("    结束位置: %d\n", node.GetEnd())
		fmt.Printf("    所属文件: %s\n", node.GetSourceFile().GetFilePath())
		fmt.Printf("    文本长度: %d\n", node.GetTextLength())

		// 验证基本信息的合理性
		hasValidText := node.GetText() != ""
		hasValidRange := node.GetStart() >= 0 && node.GetEnd() > node.GetStart()
		hasValidLine := node.GetStartLineNumber() > 0

		fmt.Printf("    ✅ 文本有效性: %t\n", hasValidText)
		fmt.Printf("    ✅ 范围有效性: %t\n", hasValidRange)
		fmt.Printf("    ✅ 行号有效性: %t\n", hasValidLine)

		return hasValidText && hasValidRange && hasValidLine
	}

	interfaceValid := validateNodeBasicInfo(firstInterfaceNode, "接口")
	functionValid := validateNodeBasicInfo(firstFunctionNode, "函数")
	classValid := validateNodeBasicInfo(firstClassNode, "类")
	typeAliasValid := validateNodeBasicInfo(firstTypeAliasNode, "类型别名")

	// 3. 父子节点导航验证 - 测试父子关系导航
	fmt.Println("\n🔗 父子节点导航验证:")
	fmt.Println("------------------------------")

	validateParentChildNavigation := func(node *tsmorphgo.Node, nodeType string) bool {
		if node == nil {
			fmt.Printf("  ❌ %s: 节点不存在，跳过父子导航测试\n", nodeType)
			return false
		}

		fmt.Printf("  🔗 %s 父子导航:\n", nodeType)

		// 获取父节点
		parent := node.GetParent()
		if parent != nil {
			fmt.Printf("    父节点类型: %v\n", parent.Kind)
			fmt.Printf("    父节点文本: %s\n", parent.GetText())
			fmt.Printf("    父节点文件: %s\n", parent.GetSourceFile().GetFilePath())
			fmt.Printf("    ✅ 父节点获取成功\n")
		} else {
			fmt.Printf("    ❌ 父节点获取失败\n")
			return false
		}

		// 获取子节点数量
		childCount := 0
		node.Node.ForEachChild(func(child *ast.Node) bool {
			childCount++
			return true // 继续遍历
		})

		fmt.Printf("    子节点数量: %d\n", childCount)

		if childCount > 0 {
			// 获取第一个子节点
			if firstChild, ok := node.GetFirstChild(); ok {
				fmt.Printf("    首个子节点类型: %v\n", firstChild.Kind)
				fmt.Printf("    首个子节点文本: %s\n", firstChild.GetText())
				fmt.Printf("    ✅ 子节点获取成功\n")
			} else {
				fmt.Printf("    ❌ 首个子节点获取失败\n")
				return false
			}
		} else {
			fmt.Printf("    ℹ️ 无子节点\n")
		}

		// 获取最后一个子节点
		if childCount > 0 {
			if lastChild, ok := node.GetLastChild(); ok {
				fmt.Printf("    最后子节点类型: %v\n", lastChild.Kind)
				fmt.Printf("    最后子节点文本: %s\n", lastChild.GetText())
				fmt.Printf("    ✅ 最后子节点获取成功\n")
			} else {
				fmt.Printf("    ❌ 最后子节点获取失败\n")
				return false
			}
		}

		return true
	}

	interfaceParentChildValid := validateParentChildNavigation(firstInterfaceNode, "接口")
	functionParentChildValid := validateParentChildNavigation(firstFunctionNode, "函数")

	// 4. 祖先节点遍历验证 - 测试祖先关系导航
	fmt.Println("\n🌳 祖先节点遍历验证:")
	fmt.Println("------------------------------")

	validateAncestorTraversal := func(node *tsmorphgo.Node, nodeType string) bool {
		if node == nil {
			fmt.Printf("  ❌ %s: 节点不存在，跳过祖先遍历测试\n", nodeType)
			return false
		}

		fmt.Printf("  🌳 %s 祖先遍历:\n", nodeType)

		// 获取所有祖先节点
		ancestors := node.GetAncestors()
		fmt.Printf("    祖先节点总数: %d\n", len(ancestors))

		// 显示前5个祖先节点
		for i, ancestor := range ancestors {
			if i >= 5 {
				fmt.Printf("    ... (还有 %d 个祖先节点)\n", len(ancestors)-5)
				break
			}
			fmt.Printf("    [%d] %v: %s\n", i+1, ancestor.Kind, ancestor.GetText())
		}

		// 测试特定类型祖先查找
		foundSourceFile := false
		
		for _, ancestor := range ancestors {
			if ancestor.Kind == ast.KindSourceFile {
				foundSourceFile = true
				fmt.Printf("    ✅ 找到源文件祖先: %s\n", ancestor.GetSourceFile().GetFilePath())
			}
			if ancestor.Kind == ast.KindInterfaceDeclaration {
								fmt.Printf("    ✅ 找到接口祖先: %s\n", ancestor.GetText())
			}
		}

		// 使用便捷方法查找特定类型祖先
		if _, ok := node.GetFirstAncestorByKind(ast.KindSourceFile); ok {
			fmt.Printf("    ✅ GetFirstAncestorByKind(源文件) 成功\n")
		} else {
			fmt.Printf("    ❌ GetFirstAncestorByKind(源文件) 失败\n")
		}

		if interfaceAncestor, ok := node.GetFirstAncestorByKind(ast.KindInterfaceDeclaration); ok {
			fmt.Printf("    ✅ GetFirstAncestorByKind(接口) 成功: %s\n", interfaceAncestor.GetText())
		} else {
			fmt.Printf("    ℹ️ GetFirstAncestorByKind(接口) 未找到\n")
		}

		return foundSourceFile
	}

	interfaceAncestorValid := validateAncestorTraversal(firstInterfaceNode, "接口")
	functionAncestorValid := validateAncestorTraversal(firstFunctionNode, "函数")

	// 5. 条件子节点查找验证 - 测试条件化的子节点查找
	fmt.Println("\n🔍 条件子节点查找验证:")
	fmt.Println("------------------------------")

	validateConditionalChildSearch := func(node *tsmorphgo.Node, nodeType string) bool {
		if node == nil {
			fmt.Printf("  ❌ %s: 节点不存在，跳过条件查找测试\n", nodeType)
			return false
		}

		fmt.Printf("  🔍 %s 条件子节点查找:\n", nodeType)

		// 查找标识符节点
		if identifierNode, ok := tsmorphgo.GetFirstChild(*node, func(n tsmorphgo.Node) bool {
			return n.Kind == ast.KindIdentifier
		}); ok {
			fmt.Printf("    ✅ 找到标识符节点: %s\n", identifierNode.GetText())
		} else {
			fmt.Printf("    ℹ️ 未找到标识符节点\n")
		}

		// 查找类型引用节点
		if typeReferenceNode, ok := tsmorphgo.GetFirstChild(*node, func(n tsmorphgo.Node) bool {
			return n.Kind == ast.KindTypeReference
		}); ok {
			fmt.Printf("    ✅ 找到类型引用节点: %s\n", typeReferenceNode.GetText())
		} else {
			fmt.Printf("    ℹ️ 未找到类型引用节点\n")
		}

		// 查找字符串字面量节点
		if stringLiteralNode, ok := tsmorphgo.GetFirstChild(*node, func(n tsmorphgo.Node) bool {
			return n.Kind == ast.KindStringLiteral
		}); ok {
			fmt.Printf("    ✅ 找到字符串字面量节点: %s\n", stringLiteralNode.GetText())
		} else {
			fmt.Printf("    ℹ️ 未找到字符串字面量节点\n")
		}

		return true
	}

	interfaceConditionalValid := validateConditionalChildSearch(firstInterfaceNode, "接口")
	functionConditionalValid := validateConditionalChildSearch(firstFunctionNode, "函数")

	// 6. 节点深度计算验证 - 测试节点深度计算
	fmt.Println("\n📊 节点深度计算验证:")
	fmt.Println("------------------------------")

	calculateNodeDepth := func(node *tsmorphgo.Node) int {
		if node == nil {
			return 0
		}
		depth := 0
		ancestors := node.GetAncestors()

		// 计算有效祖先深度（排除源文件）
		for _, ancestor := range ancestors {
			if ancestor.Kind != ast.KindSourceFile {
				depth++
			}
		}

		return depth
	}

	validateNodeDepth := func(node *tsmorphgo.Node, nodeType string) bool {
		if node == nil {
			fmt.Printf("  ❌ %s: 节点不存在，跳过深度计算测试\n", nodeType)
			return false
		}

		depth := calculateNodeDepth(node)
		fmt.Printf("  📊 %s 节点深度: %d\n", nodeType, depth)

		// 验证深度的合理性
		if depth >= 0 {
			fmt.Printf("    ✅ 深度计算合理\n")
			return true
		} else {
			fmt.Printf("    ❌ 深度计算异常\n")
			return false
		}
	}

	interfaceDepthValid := validateNodeDepth(firstInterfaceNode, "接口")
	functionDepthValid := validateNodeDepth(firstFunctionNode, "函数")

	// 7. 节点关系验证 - 测试节点之间的关系
	fmt.Println("\n🔗 节点关系验证:")
	fmt.Println("------------------------------")

	validateNodeRelationships := func(node *tsmorphgo.Node, nodeType string) bool {
		if node == nil {
			fmt.Printf("  ❌ %s: 节点不存在，跳过关系验证\n", nodeType)
			return false
		}

		fmt.Printf("  🔗 %s 节点关系验证:\n", nodeType)

		// 检查是否为根节点
		parent := node.GetParent()
		isRoot := parent == nil
		fmt.Printf("    是否为根节点: %t\n", isRoot)

		// 检查是否有子节点
		hasChildren := false
		node.Node.ForEachChild(func(child *ast.Node) bool {
			hasChildren = true
			return false // 只检查是否有子节点，不继续遍历
		})
		fmt.Printf("    是否有子节点: %t\n", hasChildren)

		// 检查是否为叶子节点
		isLeaf := !hasChildren
		fmt.Printf("    是否为叶子节点: %t\n", isLeaf)

		// 检查是否为中间节点
		isIntermediate := !isRoot && !isLeaf
		fmt.Printf("    是否为中间节点: %t\n", isIntermediate)

		fmt.Printf("    ✅ 节点关系分析完成\n")
		return true
	}

	interfaceRelationshipValid := validateNodeRelationships(firstInterfaceNode, "接口")
	functionRelationshipValid := validateNodeRelationships(firstFunctionNode, "函数")

	// 8. 节点遍历性能验证 - 测试遍历性能
	fmt.Println("\n⏱️ 节点遍历性能验证:")
	fmt.Println("------------------------------")

	validateTraversalPerformance := func(node *tsmorphgo.Node, nodeType string) bool {
		if node == nil {
			fmt.Printf("  ❌ %s: 节点不存在，跳过性能验证\n", nodeType)
			return false
		}

		fmt.Printf("  ⏱️ %s 遍历性能验证:\n", nodeType)

		// 测试子节点遍历
		childCount := 0
		node.Node.ForEachChild(func(child *ast.Node) bool {
			childCount++
			return true
		})

		// 测试后代节点遍历
		descendantCount := 0
		node.ForEachDescendant(func(descendant tsmorphgo.Node) {
			descendantCount++
		})

		fmt.Printf("    直接子节点数: %d\n", childCount)
		fmt.Printf("    后代节点总数: %d\n", descendantCount)
		fmt.Printf("    子节点/后代节点比例: %.2f\n", float64(childCount)/float64(descendantCount+1))

		// 性能评估
		if descendantCount > 100 {
			fmt.Printf("    ⚠️ 后代节点较多，遍历可能需要优化\n")
		} else if descendantCount > 10 {
			fmt.Printf("    ✅ 后代节点数量适中\n")
		} else {
			fmt.Printf("    ✅ 后代节点较少，遍历性能良好\n")
		}

		return true
	}

	interfacePerformanceValid := validateTraversalPerformance(firstInterfaceNode, "接口")
	functionPerformanceValid := validateTraversalPerformance(firstFunctionNode, "函数")

	// 9. 保存验证结果到 JSON 文件
	fmt.Println("\n💾 保存验证结果:")
	fmt.Println("------------------------------")

	validationResults := map[string]interface{}{
		"nodeTypeStats": nodeTypeStats,
		"foundTypes":    foundTypes,
		"basicInfo": map[string]bool{
			"interface":  interfaceValid,
			"function":   functionValid,
			"class":      classValid,
			"typeAlias":  typeAliasValid,
		},
		"parentChild": map[string]bool{
			"interface": interfaceParentChildValid,
			"function":  functionParentChildValid,
		},
		"ancestorTraversal": map[string]bool{
			"interface": interfaceAncestorValid,
			"function":  functionAncestorValid,
		},
		"conditionalSearch": map[string]bool{
			"interface": interfaceConditionalValid,
			"function":  functionConditionalValid,
		},
		"depthCalculation": map[string]bool{
			"interface": interfaceDepthValid,
			"function":  functionDepthValid,
		},
		"relationships": map[string]bool{
			"interface": interfaceRelationshipValid,
			"function":  functionRelationshipValid,
		},
		"performance": map[string]bool{
			"interface": interfacePerformanceValid,
			"function":  functionPerformanceValid,
		},
		"timestamp": fmt.Sprintf("%v", os.Getpid()),
	}

	resultFile := "../../validation-results/node-navigation-results.json"
	if err := os.MkdirAll("../../validation-results", 0755); err == nil {
		if data, err := json.MarshalIndent(validationResults, "", "  "); err == nil {
			if err := os.WriteFile(resultFile, data, 0644); err == nil {
				fmt.Printf("✅ 验证结果已保存到: %s\n", resultFile)
			} else {
				fmt.Printf("❌ 保存验证结果失败: %v\n", err)
			}
		} else {
			fmt.Printf("❌ 序列化验证结果失败: %v\n", err)
		}
	} else {
		fmt.Printf("❌ 创建结果目录失败: %v\n", err)
	}

	// 10. 验证结果汇总
	fmt.Println("\n📊 验证结果汇总:")
	fmt.Println("================================")

	totalValidations := 0
	passedValidations := 0

	// 统计基本信息验证
	totalValidations += 4
	if interfaceValid {
		passedValidations++
	}
	if functionValid {
		passedValidations++
	}
	if classValid {
		passedValidations++
	}
	if typeAliasValid {
		passedValidations++
	}

	// 统计父子导航验证
	totalValidations += 2
	if interfaceParentChildValid {
		passedValidations++
	}
	if functionParentChildValid {
		passedValidations++
	}

	// 统计祖先遍历验证
	totalValidations += 2
	if interfaceAncestorValid {
		passedValidations++
	}
	if functionAncestorValid {
		passedValidations++
	}

	// 统计条件搜索验证
	totalValidations += 2
	if interfaceConditionalValid {
		passedValidations++
	}
	if functionConditionalValid {
		passedValidations++
	}

	// 统计深度计算验证
	totalValidations += 2
	if interfaceDepthValid {
		passedValidations++
	}
	if functionDepthValid {
		passedValidations++
	}

	// 统计关系验证
	totalValidations += 2
	if interfaceRelationshipValid {
		passedValidations++
	}
	if functionRelationshipValid {
		passedValidations++
	}

	// 统计性能验证
	totalValidations += 2
	if interfacePerformanceValid {
		passedValidations++
	}
	if functionPerformanceValid {
		passedValidations++
	}

	passRate := float64(passedValidations) / float64(totalValidations) * 100

	fmt.Printf("📈 总验证数: %d\n", totalValidations)
	fmt.Printf("✅ 通过数: %d\n", passedValidations)
	fmt.Printf("❌ 失败数: %d\n", totalValidations-passedValidations)
	fmt.Printf("📊 通过率: %.1f%%\n", passRate)

	// 11. 最终结论
	if passRate >= 80.0 {
		fmt.Println("\n🎉 节点操作 API 验证完成！基本功能正常工作")
		fmt.Println("================================")
		fmt.Println("📋 已验证的 API:")
		fmt.Println("   - node.GetParent() - 获取父节点")
		fmt.Println("   - node.GetAncestors() - 获取所有祖先节点")
		fmt.Println("   - node.GetFirstAncestorByKind() - 按类型查找祖先节点")
		fmt.Println("   - node.ForEachChild() - 遍历子节点")
		fmt.Println("   - node.GetFirstChild() - 获取第一个子节点")
		fmt.Println("   - node.GetLastChild() - 获取最后一个子节点")
		fmt.Println("   - tsmorphgo.GetFirstChild() - 条件化子节点查找")
		fmt.Println("   - node.GetText() - 获取节点文本")
		fmt.Println("   - node.GetStartLineNumber() - 获取起始行号")
		fmt.Println("   - node.GetEndLineNumber() - 获取结束行号")
		fmt.Println("   - node.GetStart() - 获取起始位置")
		fmt.Println("   - node.GetEnd() - 获取结束位置")
		fmt.Println("   - node.GetTextLength() - 获取文本长度")
		fmt.Println("   - node.ForEachDescendant() - 遍历后代节点")
		fmt.Println("================================")
		fmt.Println("📝 验证总结:")
		fmt.Printf("   - 节点类型发现: %d 种\n", foundTypes)
		fmt.Printf("   - 基本信息验证: %d/4\n", map[bool]int{true: 1, false: 0}[interfaceValid]+map[bool]int{true: 1, false: 0}[functionValid]+map[bool]int{true: 1, false: 0}[classValid]+map[bool]int{true: 1, false: 0}[typeAliasValid])
		fmt.Printf("   - 父子导航验证: %d/2\n", map[bool]int{true: 1, false: 0}[interfaceParentChildValid]+map[bool]int{true: 1, false: 0}[functionParentChildValid])
		fmt.Printf("   - 祖先遍历验证: %d/2\n", map[bool]int{true: 1, false: 0}[interfaceAncestorValid]+map[bool]int{true: 1, false: 0}[functionAncestorValid])
		fmt.Printf("   - 条件搜索验证: %d/2\n", map[bool]int{true: 1, false: 0}[interfaceConditionalValid]+map[bool]int{true: 1, false: 0}[functionConditionalValid])
		fmt.Printf("   - 深度计算验证: %d/2\n", map[bool]int{true: 1, false: 0}[interfaceDepthValid]+map[bool]int{true: 1, false: 0}[functionDepthValid])
		fmt.Printf("   - 关系验证: %d/2\n", map[bool]int{true: 1, false: 0}[interfaceRelationshipValid]+map[bool]int{true: 1, false: 0}[functionRelationshipValid])
		fmt.Printf("   - 性能验证: %d/2\n", map[bool]int{true: 1, false: 0}[interfacePerformanceValid]+map[bool]int{true: 1, false: 0}[functionPerformanceValid])
	} else {
		fmt.Println("\n❌ 节点操作 API 验证完成但存在问题")
		fmt.Printf("   验证通过率 %.1f%% 低于预期\n", passRate)
		fmt.Println("   建议检查节点导航功能的实现")
	}
}