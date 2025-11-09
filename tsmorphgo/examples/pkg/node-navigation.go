//go:build node_navigation
// +build node_navigation

package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🔍 TSMorphGo 节点导航 - 正确使用姿势")
	fmt.Println("=" + repeat("=", 50))

	// =============================================================================
	// 本文件演示 TSMorphGo 节点导航和位置信息的正确使用方法
	// =============================================================================
	// 学习级别: 初级 → 高级
	// 预计时间: 40-60分钟
	//
	// 功能覆盖:
	// - 基础: 节点遍历、父子关系、祖先查找
	// - 高级: 精确位置计算 ⭐、IDE链接生成 ⭐
	// - 应用: 代码分析、IDE开发、重构工具
	//
	// ⭐ = 高级功能，初学者可先跳过
	//
	// 对齐 ts-morph API:
	// - node.forEachDescendant() → node.ForEachDescendant()
	// - node.getParent() → node.GetParent()
	// - node.getAncestors() → node.GetAncestors()
	// - node.getFirstAncestorByKind() → node.GetFirstAncestorByKind()
	// - node.getStart() → node.GetStart()
	// - node.getStartLineNumber() → node.GetStartLineNumber()
	// - node.getStartLinePos() → node.GetStartLinePos()
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

	// 选择App.tsx作为主要分析文件
	appFile := project.GetSourceFile(realProjectPath + "/src/App.tsx")
	if appFile == nil {
		log.Fatal("未找到 App.tsx 文件")
	}

	fmt.Printf("📄 分析文件: %s\n", appFile.GetFilePath())
	fmt.Println("=" + repeat("=", 30))

	// 示例1: 基础节点遍历 (初级)
	// 对应 ts-morph: node.forEachDescendant(callback)
	fmt.Println("\n🔄 示例1: 基础节点遍历 (初级)")
	fmt.Println("对齐 ts-morph: node.forEachDescendant(callback)")
	fmt.Println("功能: 深度优先遍历所有子节点")

	var totalNodes, functionNodes, variableNodes, callNodes int
	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		totalNodes++

		switch {
		case tsmorphgo.IsFunctionDeclaration(node):
			functionNodes++
		case tsmorphgo.IsVariableDeclaration(node):
			variableNodes++
		case tsmorphgo.IsCallExpression(node):
			callNodes++
		}
	})

	fmt.Printf("📊 节点统计:\n")
	fmt.Printf("  - 总节点数: %d\n", totalNodes)
	fmt.Printf("  - 函数声明: %d\n", functionNodes)
	fmt.Printf("  - 变量声明: %d\n", variableNodes)
	fmt.Printf("  - 函数调用: %d\n", callNodes)

	// 示例2: 父节点和祖先节点导航 (初级)
	// 对应 ts-morph: node.getParent(), node.getAncestors()
	fmt.Println("\n👆 示例2: 父节点和祖先节点导航 (初级)")
	fmt.Println("对齐 ts-morph: node.getParent(), node.getAncestors()")
	fmt.Println("功能: 向上遍历节点树，理解节点间的关系")

	// 查找useState标识符并分析其上下文
	useStateCount := 0
	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if useStateCount >= 3 { // 只分析前3个
			return
		}

		if tsmorphgo.IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "useState" {
			useStateCount++
			fmt.Printf("\nuseState 使用 %d:\n", useStateCount)
			fmt.Printf("  - 位置: 行 %d, 列 %d\n", node.GetStartLineNumber(), node.GetStartColumnNumber())
			fmt.Printf("  - 完整文本: %s\n", node.GetText())

			// 获取父节点
			parent := node.GetParent()
			if parent != nil {
				fmt.Printf("  - 父节点类型: %s\n", parent.GetKindName())
				fmt.Printf("  - 父节点内容: %s\n", truncateString(parent.GetText(), 50))
			}

			// 获取所有祖先节点
			ancestors := node.GetAncestors()
			fmt.Printf("  - 祖先节点数量: %d\n", len(ancestors))
			if len(ancestors) > 0 {
				// 显示前3个祖先节点
				fmt.Printf("  - 部分祖先链: ")
				for i, ancestor := range ancestors {
					if i >= 3 {
						fmt.Printf("... (共%d个)", len(ancestors))
						break
					}
					fmt.Printf("%s → ", ancestor.GetKindName())
				}
				fmt.Printf("\n")
			}
		}
	})

	// 示例3: 条件祖先查找 (中级)
	// 对应 ts-morph: node.getFirstAncestorByKind()
	fmt.Println("\n🎯 示例3: 条件祖先查找 (中级)")
	fmt.Println("对齐 ts-morph: node.getFirstAncestorByKind(kind)")
	fmt.Println("功能: 根据节点类型查找特定祖先")

	// 查找函数声明中的useState调用
	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "useState" {
			// 查找最近的函数声明祖先
			if funcAncestor, ok := node.GetFirstAncestorByKind(tsmorphgo.KindFunctionDeclaration); ok {
				fmt.Printf("找到 useState 在函数中:\n")
				fmt.Printf("  - useState位置: 行 %d\n", node.GetStartLineNumber())
				fmt.Printf("  - 函数位置: 行 %d\n", funcAncestor.GetStartLineNumber())

				// 获取函数名
				if funcName, ok := tsmorphgo.GetFirstChild(*funcAncestor, tsmorphgo.IsIdentifier); ok {
					fmt.Printf("  - 函数名: %s\n", funcName.GetText())
				}
				return
			}
		}
	})

	// 示例4: 条件遍历和性能优化 (中级)
	fmt.Println("\n⚡ 示例4: 条件遍历和性能优化 (中级)")
	fmt.Println("功能: 提前终止遍历，提高性能")

	// 查找第一个类声明，找到后立即停止
	var foundClass *tsmorphgo.Node

	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if foundClass != nil {
			return // 已经找到，停止遍历
		}

		if tsmorphgo.IsClassDeclaration(node) {
			foundClass = &node
			fmt.Printf("✅ 找到第一个类声明:\n")
			fmt.Printf("  - 位置: 行 %d\n", node.GetStartLineNumber())
			fmt.Printf("  - 节点类型: %s\n", node.GetKindName())

			// 获取类名
			if className, ok := tsmorphgo.GetFirstChild(node, tsmorphgo.IsIdentifier); ok {
				fmt.Printf("  - 类名: %s\n", className.GetText())
			}
			return
		}
	})

	if foundClass == nil {
		fmt.Printf("❌ 未找到类声明\n")
	}

	// 示例5: 精确位置信息 (高级 ⭐)
	// 对应 ts-morph: node.getStart(), node.getStartLinePos(), node.getStartLineNumber(), node.getStartColumnNumber()
	fmt.Println("\n📍 示例5: 精确位置信息 (高级 ⭐)")
	fmt.Println("对齐 ts-morph: node.getStart(), node.getStartLinePos(), node.getStartLineNumber(), node.getStartColumnNumber()")
	fmt.Println("应用: IDE开发、代码高亮、错误定位、跳转定义")

	positionCount := 0
	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if positionCount >= 5 { // 只演示前5个
			return
		}

		// 重点分析变量声明的位置信息
		if tsmorphgo.IsVariableDeclaration(node) {
			if varName, ok := tsmorphgo.GetVariableName(node); ok && len(varName) > 0 {
				positionCount++
				fmt.Printf("\n位置信息 %d - 变量: '%s'\n", positionCount, varName)

				// GetStart() 获取节点在文件中的起始字符位置 (0-based)
				// 对应 ts-morph: node.getStart()
				startPos := node.GetStart()
				fmt.Printf("  - 起始位置(文件偏移): %d\n", startPos)

				// GetStartLineNumber() 获取起始行号 (1-based)
				// 对应 ts-morph: node.getStartLineNumber()
				lineNumber := node.GetStartLineNumber()
				fmt.Printf("  - 起始行号: %d\n", lineNumber)

				// GetStartColumnNumber() 获取起始列号 (1-based)
				// 对应 ts-morph: node.getStartLineCharacter()
				columnNumber := node.GetStartColumnNumber()
				fmt.Printf("  - 起始列号: %d\n", columnNumber)

				// GetStartLinePos() 获取节点所在行的起始字符位置 (0-based)
				// 对应 ts-morph: node.getStartLinePos()
				startLinePos := node.GetStartLinePos()
				fmt.Printf("  - 行起始位置: %d\n", startLinePos)

				// 计算相对列位置
				relativeColumn := startPos - startLinePos
				fmt.Printf("  - 行内相对位置: %d (0-based)\n", relativeColumn)

				// 完整位置信息结构
				if posInfo := node.GetPositionInfo(); posInfo != nil {
					fmt.Printf("  - 完整位置信息: Line=%d, Column=%d, Offset=%d\n",
						posInfo.Line, posInfo.Column, posInfo.StartOffset)
				}

				// 验证位置计算的正确性
				fmt.Printf("  - 验证: 列号-1 = %d, 相对位置 = %d, 相等吗? %v\n",
					columnNumber-1, relativeColumn, columnNumber-1 == relativeColumn)
			}
		}
	})

	// 示例6: IDE功能应用示例 (高级 ⭐)
	fmt.Println("\n💻 示例6: IDE功能应用示例 (高级 ⭐)")
	fmt.Println("应用: 生成编辑器链接、代码上下文提取、跳转定义")

	// 查找fetchUsers函数并生成IDE链接
	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsFunctionDeclaration(node) {
			if funcName, ok := tsmorphgo.GetFunctionDeclarationNameNode(node); ok {
				if funcName.GetText() == "fetchUsers" {
					fmt.Printf("找到函数: %s\n", funcName.GetText())

					// 生成IDE链接格式
					filePath := node.GetSourceFile().GetFilePath()
					line := node.GetStartLineNumber()

					fmt.Printf("🔗 IDE跳转链接:\n")
					fmt.Printf("  - VSCode: %s:%d:%d\n", filePath, line, 1)
					fmt.Printf("  - IntelliJ: %s:%d\n", filePath, line)
					fmt.Printf("  - WebStorm: %s:%d\n", filePath, line)
					fmt.Printf("  - GitHub: %s#L%d\n", extractRelativePath(realProjectPath, filePath), line)

					// 提取代码上下文
					if fileResult := node.GetSourceFile().GetFileResult(); fileResult != nil {
						lines := strings.Split(fileResult.Raw, "\n")
						if line > 0 && line <= len(lines) {
							fmt.Printf("📝 代码上下文:\n")

							// 显示前后各2行
							start := line - 2
							if start < 1 {
								start = 1
							}
							end := line + 2
							if end > len(lines) {
								end = len(lines)
							}

							for i := start; i <= end; i++ {
								prefix := "    "
								if i == line {
									prefix = ">>> " // 标记目标行
								}
								fmt.Printf("%s%d: %s\n", prefix, i, lines[i-1])
							}
						}
					}

					return
				}
			}
		}
	})

	// 示例7: 多节点位置比较 (高级 ⭐)
	fmt.Println("\n⚖️ 示例7: 多节点位置比较 (高级 ⭐)")
	fmt.Println("应用: 代码重构、依赖分析、影响评估")

	// 收集App.tsx中的前5个函数调用
	var callPositions []struct {
		name   string
		start  int
		line   int
		column int
		text   string
	}

	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if len(callPositions) >= 5 {
			return
		}

		if tsmorphgo.IsCallExpression(node) {
			if expr, ok := tsmorphgo.GetCallExpressionExpression(node); ok {
				text := strings.TrimSpace(expr.GetText())
				if len(text) > 0 && len(text) <= 20 { // 避免太长的表达式
					callPositions = append(callPositions, struct {
						name   string
						start  int
						line   int
						column int
						text   string
					}{
						name:   text,
						start:  node.GetStart(),
						line:   node.GetStartLineNumber(),
						column: node.GetStartColumnNumber(),
						text:   truncateString(node.GetText(), 30),
					})
				}
			}
		}
	})

	if len(callPositions) > 0 {
		fmt.Printf("函数调用位置分析 (前%d个):\n", len(callPositions))

		// 按位置排序
		for i, call := range callPositions {
			fmt.Printf("  %d. %s\n", i+1, call.text)
			fmt.Printf("     位置: 行 %d, 列 %d, 偏移 %d\n", call.line, call.column, call.start)
		}

		// 分析相邻调用
		if len(callPositions) >= 2 {
			fmt.Printf("\n📊 相邻调用分析:\n")
			for i := 1; i < len(callPositions); i++ {
				prev := callPositions[i-1]
				curr := callPositions[i]
				distance := curr.start - prev.start
				lineDiff := curr.line - prev.line

				fmt.Printf("  %s → %s:\n", prev.text, curr.text)
				fmt.Printf("    - 字符距离: %d\n", distance)
				fmt.Printf("    - 行距离: %d\n", lineDiff)
				if lineDiff == 0 {
					fmt.Printf("    - 关系: 同一行\n")
				} else if lineDiff == 1 {
					fmt.Printf("    - 关系: 相邻行\n")
				}
			}
		}
	}

	fmt.Println("\n🎯 节点导航使用姿势总结:")
	fmt.Println("1. 遍历节点 → 使用 ForEachDescendant() + 条件判断")
	fmt.Println("2. 查找父节点 → 使用 GetParent() 检查关系")
	fmt.Println("3. 查找祖先 → 使用 GetAncestors() 或 GetFirstAncestorByKind()")
	fmt.Println("4. 性能优化 → 在回调中及时 return 避免不必要遍历")
	fmt.Println("5. 位置信息 → GetStart() + GetStartLineNumber() + GetStartColumnNumber()")
	fmt.Println("6. IDE集成 → 结合位置信息生成跳转链接和上下文")

	fmt.Println("\n✅ 节点导航示例完成!")
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
