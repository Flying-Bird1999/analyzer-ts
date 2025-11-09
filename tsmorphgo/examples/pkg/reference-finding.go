//go:build reference_finding
// +build reference_finding

package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🔗 TSMorphGo 引用查找 - 正确使用姿势")
	fmt.Println("=" + repeat("=", 50))

	// =============================================================================
	// 本文件演示 TSMorphGo 引用查找和符号系统的正确使用方法
	// =============================================================================
	// 学习级别: 中级 → 高级
	// 预计时间: 45-60分钟
	//
	// 功能覆盖:
	// - 基础: FindReferences() 引用查找、GotoDefinition() 跳转定义
	// - 高级: 缓存机制 ⭐、符号系统 ⭐、重命名安全性分析 ⭐
	// - 应用: IDE功能、重构工具、代码导航
	//
	// ⭐ = 高级功能，需要LSP服务支持
	//
	// 对齐 ts-morph API:
	// - identifier.findReferencesAsNodes() → FindReferences()
	// - identifier.getDefinitionNodes() → GotoDefinition()
	// - node.getSymbol() → GetSymbol()
	// - symbol.getName() → symbol.GetName()
	// =============================================================================

	// 计算 demo-react-app 的绝对路径
	realProjectPath, err := filepath.Abs(filepath.Join("..", "demo-react-app"))
	if err != nil {
		log.Fatalf("无法解析项目路径: %v", err)
	}
	fmt.Printf("✅ 项目路径: %s\n", realProjectPath)

	// 初始化项目
	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:         realProjectPath,
		TargetExtensions: []string{".ts", ".tsx"},
		IgnorePatterns:   []string{"node_modules", "dist", ".git", "build"},
		UseTsConfig:      true,
	})
	defer project.Close()

	fmt.Printf("📄 项目路径: %s\n", realProjectPath)
	fmt.Printf("📊 源文件数量: %d\n", len(project.GetSourceFiles()))

	// 示例1: 基础引用查找 (中级)
	// 对应 ts-morph: identifier.findReferencesAsNodes()
	fmt.Println("\n🔍 示例1: 基础引用查找 (中级)")
	fmt.Println("对齐 ts-morph: identifier.findReferencesAsNodes()")
	fmt.Println("功能: 查找变量、函数、类型在整个项目中的所有引用")

	// 查找useUsers变量在项目中的引用
	fmt.Println("\n查找 'useUsers' 变量的引用:")

	// 在hooks/useUserQuery.ts中查找useUsers声明
	useUserQueryFile := project.GetSourceFile(realProjectPath + "/src/hooks/useUserQuery.ts")
	if useUserQueryFile == nil {
		log.Fatal("未找到 useUserQuery.ts 文件")
	}

	var useUsersNode *tsmorphgo.Node
	useUserQueryFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if useUsersNode != nil {
			return // 已经找到，停止遍历
		}

		if tsmorphgo.IsVariableDeclaration(node) {
			if varName, ok := tsmorphgo.GetVariableName(node); ok && varName == "useUsers" {
				// GetFirstChild 获取变量名对应的标识符节点
				if nameNode, ok := tsmorphgo.GetFirstChild(node, tsmorphgo.IsIdentifier); ok {
					useUsersNode = nameNode
				}
			}
		}
	})

	if useUsersNode == nil {
		log.Fatal("未找到 useUsers 变量声明")
	}

	// GetSymbol 获取节点的符号信息
	symbol, err := tsmorphgo.GetSymbol(*useUsersNode)
	if err != nil {
		log.Printf("获取符号失败: %v", err)
	} else {
		fmt.Printf("✅ 找到符号: %s\n", symbol.GetName())
		fmt.Printf("📍 声明位置: %s:%d\n", useUsersNode.GetSourceFile().GetFilePath(), useUsersNode.GetStartLineNumber())
	}

	// FindReferences 查找所有引用
	// 对应 ts-morph: identifier.findReferencesAsNodes()
	fmt.Println("\n🔍 执行引用查找...")
	start := time.Now()
	refs, err := tsmorphgo.FindReferences(*useUsersNode)
	duration := time.Since(start)

	if err != nil {
		log.Printf("查找引用失败: %v", err)
		return
	}

	fmt.Printf("✅ 引用查找完成!\n")
	fmt.Printf("📊 查找统计:\n")
	fmt.Printf("  - 查找耗时: %v\n", duration)
	fmt.Printf("  - 引用数量: %d\n", len(refs))

	if len(refs) == 0 {
		fmt.Println("  - 结果: 未找到任何引用")
	} else {
		fmt.Printf("  - 引用列表:\n")
		for i, ref := range refs {
			parent := ref.GetParent()
			context := ""
			if parent != nil {
				parentText := strings.TrimSpace(parent.GetText())
				if len(parentText) > 60 {
					parentText = parentText[:57] + "..."
				}
				context = parentText
			}

			filePath := ref.GetSourceFile().GetFilePath()
			relativePath := extractRelativePath(realProjectPath, filePath)

			fmt.Printf("    %d. %s:%d - %s\n", i+1, relativePath, ref.GetStartLineNumber(), context)
		}
	}

	// 示例2: 跳转到定义 (中级)
	// 对应 ts-morph: identifier.getDefinitionNodes()
	fmt.Println("\n📍 示例2: 跳转到定义 (中级)")
	fmt.Println("对齐 ts-morph: identifier.getDefinitionNodes()")
	fmt.Println("功能: 从引用点跳转到声明位置")

	// 在App.tsx中查找useUsers的使用，然后跳转到定义
	appFile := project.GetSourceFile(realProjectPath + "/src/App.tsx")
	if appFile != nil {
		var foundUsage *tsmorphgo.Node
		appFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if foundUsage != nil {
				return // 已经找到使用点，停止遍历
			}

			// 查找useUsers标识符
			if tsmorphgo.IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "useUsers" {
				// 确保不是它自己的声明
				parent := node.GetParent()
				if parent != nil && parent.Kind == tsmorphgo.KindVariableDeclaration {
					// 这是声明，不是使用，跳过
					return
				}

				foundUsage = &node
			}
		})

		if foundUsage != nil {
			fmt.Printf("📍 找到使用点: %s:%d\n",
				extractRelativePath(realProjectPath, foundUsage.GetSourceFile().GetFilePath()),
				foundUsage.GetStartLineNumber())

			// GotoDefinition 跳转到定义
			// 对应 ts-morph: identifier.getDefinitionNodes()
			fmt.Println("\n🎯 执行跳转到定义...")
			start = time.Now()
			defs, err := tsmorphgo.GotoDefinition(*foundUsage)
			duration = time.Since(start)

			if err != nil {
				log.Printf("跳转定义失败: %v", err)
			} else {
				fmt.Printf("✅ 跳转定义完成! 耗时: %v\n", duration)
				fmt.Printf("📍 找到 %d 个定义:\n", len(defs))

				for i, def := range defs {
					defPath := extractRelativePath(realProjectPath, def.GetSourceFile().GetFilePath())
					fmt.Printf("    %d. %s:%d - %s\n", i+1, defPath, def.GetStartLineNumber(),
						truncateString(strings.TrimSpace(def.GetText()), 50))
				}
			}
		} else {
			fmt.Println("⚠️  未找到 useUsers 的使用点")
		}
	}

	// 示例3: 缓存机制和性能优化 (高级 ⭐)
	fmt.Println("\n⚡ 示例3: 缓存机制和性能优化 (高级 ⭐)")
	fmt.Println("功能: 提高重复查找的性能，避免重复的LSP调用")

	if len(refs) > 0 {
		testRef := refs[0] // 使用第一个引用进行测试

		fmt.Printf("🔬 缓存性能测试 (使用第一个引用):\n")

		var source1, source2 string

		// 第一次查找 - 应该调用LSP服务
		fmt.Printf("  第一次查找:")
		start = time.Now()
		refs1, fromCache1, err := tsmorphgo.FindReferencesWithCache(*testRef)
		duration1 := time.Since(start)
		if err != nil {
			log.Printf("    - 查找失败: %v\n", err)
		} else {
			source1 = "LSP服务"
			if fromCache1 {
				source1 = "缓存"
			}
			fmt.Printf("    - 耗时: %v\n", duration1)
			fmt.Printf("    - 数据源: %s\n", source1)
			fmt.Printf("    - 引用数: %d\n", len(refs1))
		}

		// 第二次查找 - 应该使用缓存
		fmt.Printf("  第二次查找:")
		start = time.Now()
		refs2, fromCache2, err := tsmorphgo.FindReferencesWithCache(*testRef)
		duration2 := time.Since(start)
		if err != nil {
			log.Printf("    - 查找失败: %v\n", err)
		} else {
			source2 = "LSP服务"
			if fromCache2 {
				source2 = "缓存"
			}
			fmt.Printf("    - 耗时: %v\n", duration2)
			fmt.Printf("    - 数据源: %s\n", source2)
			fmt.Printf("    - 引用数: %d\n", len(refs2))
		}

		// 计算性能提升
		if duration1 > 0 && duration2 > 0 {
			speedup := float64(duration1) / float64(duration2)
			fmt.Printf("\n📊 性能对比:\n")
			fmt.Printf("  - 第一次查找: %v (来自 %s)\n", duration1, source1)
			fmt.Printf("  - 第二次查找: %v (来自 %s)\n", duration2, source2)
			fmt.Printf("  - 性能提升: %.1fx 倍\n", speedup)
			fmt.Printf("  - 节省时间: %v\n", duration1-duration2)

			if speedup > 10 {
				fmt.Printf("  🚀 缓存效果显著！\n")
			} else if speedup > 2 {
				fmt.Printf("  ✅ 缓存效果良好\n")
			} else {
				fmt.Printf("  ⚠️  缓存效果一般\n")
			}
		}
	}

	// 示例4: 符号系统深度分析 (高级 ⭐)
	// 对应 ts-morph: node.getSymbol(), symbol.getName()
	fmt.Println("\n🧬 示例4: 符号系统深度分析 (高级 ⭐)")
	fmt.Println("对齐 ts-morph: node.getSymbol(), symbol.getName()")
	fmt.Println("功能: 语义级别的代码分析，比文本匹配更准确")

	// 分析types.ts中的符号
	typesFile := project.GetSourceFile(realProjectPath + "/src/types.ts")
	if typesFile != nil {
		fmt.Printf("\n📋 分析 %s 中的符号:\n", extractFileName(typesFile.GetFilePath()))

		var symbols []struct {
			name     string
			node     *tsmorphgo.Node
			line     int
			exports  bool
			typeInfo string
		}

		typesFile.ForEachDescendant(func(node tsmorphgo.Node) {
			// 重点关注标识符节点
			if tsmorphgo.IsIdentifier(node) {
				text := strings.TrimSpace(node.GetText())
				// 跳过太短或太长的标识符
				if len(text) < 2 || len(text) > 20 {
					return
				}

				// GetSymbol 获取符号信息
				symbol, err := tsmorphgo.GetSymbol(node)
				if err == nil && symbol != nil {
					// 检查是否导出
					isExported := false
					parent := node.GetParent()
					for parent != nil {
						parentText := strings.ToLower(parent.GetText())
						if strings.Contains(parentText, "export") {
							isExported = true
							break
						}
						parent = parent.GetParent()
					}

					// 获取类型信息
					typeInfo := "未知"
					if symbol.HasType() {
						typeInfo = "有类型信息"
					}

					symbols = append(symbols, struct {
						name     string
						node     *tsmorphgo.Node
						line     int
						exports  bool
						typeInfo string
					}{
						name:     symbol.GetName(),
						node:     &node,
						line:     node.GetStartLineNumber(),
						exports:  isExported,
						typeInfo: typeInfo,
					})
				}
			}
		})

		fmt.Printf("  - 符号总数: %d\n", len(symbols))

		// 按名称排序显示
		symbolMap := make(map[string]struct {
			node     *tsmorphgo.Node
			line     int
			exports  bool
			typeInfo string
		})

		for _, sym := range symbols {
			if _, exists := symbolMap[sym.name]; !exists {
				symbolMap[sym.name] = struct {
					node     *tsmorphgo.Node
					line     int
					exports  bool
					typeInfo string
				}{
					node:     sym.node,
					line:     sym.line,
					exports:  sym.exports,
					typeInfo: sym.typeInfo,
				}
			}
		}

		fmt.Printf("  - 符号列表 (按名称排序):\n")
		count := 0
		for name, info := range symbolMap {
			if count >= 8 { // 只显示前8个
				fmt.Printf("    ... 还有 %d 个符号\n", len(symbolMap)-count)
				break
			}
			count++
			status := "私有"
			if info.exports {
				status = "导出"
			}
			fmt.Printf("    - %s (%s, 行 %d, %s)\n", name, status, info.line, info.typeInfo)
		}
	}

	// 示例5: 重命名安全性分析 (高级 ⭐)
	// 对应 ts-morph: 基于符号的重命名影响分析
	fmt.Println("\n🛡️ 示例5: 重命名安全性分析 (高级 ⭐)")
	fmt.Println("应用: 重构工具的安全性评估、影响范围分析")

	// 在App.tsx中找到合适的符号进行重命名测试
	var targetSymbol *tsmorphgo.Symbol
	var targetName string
	var targetFile string

	appFile = project.GetSourceFile(realProjectPath + "/src/App.tsx")
	if appFile != nil {
		appFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if targetSymbol != nil {
				return
			}

			// 查找合适的变量进行重命名测试
			if tsmorphgo.IsIdentifier(node) {
				text := strings.TrimSpace(node.GetText())
				// 选择一个合适的变量进行测试
				if text == "users" || text == "loading" || text == "fetchUsers" {
					symbol, err := tsmorphgo.GetSymbol(node)
					if err == nil && symbol != nil {
						targetSymbol = symbol
						targetName = text
						targetFile = extractFileName(appFile.GetFilePath())
						return
					}
				}
			}
		})
	}

	if targetSymbol != nil {
		fmt.Printf("🎯 重命名安全性分析: '%s'\n", targetName)
		fmt.Printf("📍 目标文件: %s\n", targetFile)

		// 统计所有文件中的引用
		refCount := 0
		fileRefs := make(map[string]int)

		for _, file := range project.GetSourceFiles() {
			fileRefCount := 0
			file.ForEachDescendant(func(node tsmorphgo.Node) {
				if tsmorphgo.IsIdentifier(node) && strings.TrimSpace(node.GetText()) == targetName {
					symbol, err := tsmorphgo.GetSymbol(node)
					if err == nil && symbol != nil && symbol.GetName() == targetSymbol.GetName() {
						fileRefCount++
						refCount++
					}
				}
			})

			if fileRefCount > 0 {
				fileRefs[file.GetFilePath()] = fileRefCount
			}
		}

		fmt.Printf("\n📊 重命名影响分析:\n")
		fmt.Printf("  - 总引用数: %d\n", refCount)
		fmt.Printf("  - 影响文件数: %d\n", len(fileRefs))
		fmt.Printf("  - 文件引用分布:\n")

		for filePath, count := range fileRefs {
			relativePath := extractRelativePath(realProjectPath, filePath)
			fmt.Printf("    - %s: %d 个引用\n", relativePath, count)
		}

		// 安全性评估
		fmt.Printf("\n🔒 安全性评估:\n")
		if refCount > 20 {
			fmt.Printf("  ❌ 高风险: 引用过多 (%d个)\n", refCount)
			fmt.Printf("     建议: 重命名前请仔细测试，考虑分批处理\n")
		} else if refCount > 10 {
			fmt.Printf("  ⚠️  中风险: 引用较多 (%d个)\n", refCount)
			fmt.Printf("     建议: 重命名后请运行完整测试套件\n")
		} else if refCount > 5 {
			fmt.Printf("  ✅ 低风险: 引用适中 (%d个)\n", refCount)
			fmt.Printf("     建议: 重命名后运行相关测试即可\n")
		} else {
			fmt.Printf("  ✅ 很安全: 引用很少 (%d个)\n", refCount)
			fmt.Printf("     建议: 可以安全重命名\n")
		}

		// 具体建议
		fmt.Printf("\n💡 重命名建议:\n")
		fmt.Printf("  1. 使用IDE的重构功能 (如VS Code的 F2 重命名)\n")
		fmt.Printf("  2. 运行完整测试套件确保功能正确\n")
		if refCount > 10 {
			fmt.Printf("  3. 考虑分批次重命名，降低风险\n")
		}
		fmt.Printf("  4. 重命名后检查编译是否成功\n")

	} else {
		fmt.Printf("⚠️  未找到合适的符号进行重命名分析\n")
		fmt.Printf("     尝试查找: users, loading, fetchUsers 等变量\n")
	}

	// 示例6: 错误处理和边界情况 (中级)
	fmt.Println("\n🛡️ 示例6: 错误处理和边界情况 (中级)")
	fmt.Println("功能: 处理各种异常情况，提高代码健壮性")

	// 测试查找不存在符号的引用
	fmt.Println("\n🔍 测试不存在的符号:")

	// 创建临时项目来测试错误处理
	testProject := tsmorphgo.NewProjectFromSources(map[string]string{
		"/test.ts": `
			const unknownVar = "test";
			console.log(unknownVar);
		`,
	})
	defer testProject.Close()

	testFile := testProject.GetSourceFile("/test.ts")
	if testFile != nil {
		var unknownNode *tsmorphgo.Node
		testFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if unknownNode != nil {
				return // 已经找到
			}

			if tsmorphgo.IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "unknownVar" {
				unknownNode = &node
			}
		})

		if unknownNode != nil {
			fmt.Printf("  - 找到未定义标识符: '%s'\n", unknownNode.GetText())
			fmt.Printf("  - 位置: 行 %d\n", unknownNode.GetStartLineNumber())

			// 尝试查找引用
			refs, err := tsmorphgo.FindReferences(*unknownNode)
			if err != nil {
				fmt.Printf("  - 引用查找失败 (预期): %v\n", err)
				fmt.Printf("  - 原因: '%s' 未定义，没有符号信息\n", unknownNode.GetText())
			} else {
				fmt.Printf("  - 意外找到引用: %d 个\n", len(refs))
			}
		}
	}

	// 测试空节点
	fmt.Println("\n🔍 测试空节点处理:")
	var emptyNode tsmorphgo.Node
	_, err = tsmorphgo.FindReferences(emptyNode)
	if err != nil {
		fmt.Printf("  - 空节点查找失败 (预期): %v\n", err)
	}

	// 测试无效位置
	// 这里可以添加更多边界情况的测试

	fmt.Println("\n🎯 引用查找使用姿势总结:")
	fmt.Println("1. 基础查找 → FindReferences(node) 获取所有引用")
	fmt.Println("2. 跳转定义 → GotoDefinition(node) 跳转到声明位置")
	fmt.Println("3. 性能优化 → FindReferencesWithCache(node) 使用缓存")
	fmt.Println("4. 符号分析 → GetSymbol(node) + symbol.GetName() 获取语义信息")
	fmt.Println("5. 重命名安全 → 基于符号统计引用，评估影响范围")
	fmt.Println("6. 错误处理 → 检查返回值，处理不存在的符号")
	fmt.Println("7. 性能考虑 → 缓存重复查找，避免重复LSP调用")

	fmt.Println("\n✅ 引用查找示例完成!")
}

// 辅助函数：重复字符串
func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// 辅助函数：截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// 辅助函数：提取相对路径
func extractRelativePath(basePath, fullPath string) string {
	if strings.HasPrefix(fullPath, basePath) {
		return fullPath[len(basePath):]
	}
	return fullPath
}

// 辅助函数：提取文件名
func extractFileName(filePath string) string {
	parts := strings.Split(filePath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return filePath
}
