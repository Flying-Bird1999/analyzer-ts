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
	fmt.Println("🔍 TSMorphGo 节点导航 - 新API演示")
	fmt.Println("=" + strings.Repeat("=", 50))

	// =============================================================================
	// 本文件演示新的统一API在节点导航中的应用
	// =============================================================================
	// 学习级别: 初级 → 高级
	// 预计时间: 15-20分钟
	//
	// 新API的优势:
	// - 统一的接口设计，无需记忆大量函数名
	// - 支持分析真实文件系统项目
	// - 更简洁的方法调用
	//
	// 新API功能:
	// - node.IsFunctionDeclaration() → 函数声明检查
	// - node.IsVariableDeclaration() → 变量声明检查
	// - node.IsCallExpr() → 函数调用检查
	// - node.IsIdentifierNode() → 标识符检查
	// - node.GetParent() → 父节点访问
	// - node.GetAncestors() → 祖先节点列表
	// =============================================================================

	// 获取 demo-react-app 的绝对路径
	realProjectPath, err := filepath.Abs("../demo-react-app")
	if err != nil {
		log.Fatalf("无法解析项目路径: %v", err)
	}

	// 使用 NewProject 加载真实的 React 项目进行演示
	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:    realProjectPath,
		UseTsConfig: true,
	})
	defer project.Close()

	// 获取演示文件
	appFilePath := filepath.Join(realProjectPath, "src/App.tsx")
	appFile := project.GetSourceFile(appFilePath)
	if appFile == nil {
		fmt.Printf("❌ 未找到 App.tsx 文件 at %s\n", appFilePath)
		return
	}

	fmt.Printf("📄 分析文件: %s\n", appFile.GetFilePath())
	fmt.Println("=" + strings.Repeat("=", 30))

	// 示例1: 基础节点遍历 (初级)
	fmt.Println("\n🔄 示例1: 基础节点遍历 (初级)")
	fmt.Println("展示如何使用新API遍历和分析节点")

	var (
		totalNodes      int
		functionNodes   int
		variableNodes   int
		callNodes       int
		identifierNodes int
	)

	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		totalNodes++

		switch {
		case node.IsFunctionDeclaration():
			functionNodes++
		case node.IsVariableDeclaration():
			variableNodes++
		case node.IsCallExpr():
			callNodes++
		case node.IsIdentifierNode():
			identifierNodes++
		}
	})

	fmt.Printf("📊 节点统计:\n")
	fmt.Printf("  - 总节点数: %d\n", totalNodes)
	fmt.Printf("  - 函数声明: %d\n", functionNodes)
	fmt.Printf("  - 变量声明: %d\n", variableNodes)
	fmt.Printf("  - 函数调用: %d\n", callNodes)
	fmt.Printf("  - 标识符: %d\n", identifierNodes)

	// 示例2: 父节点和祖先节点导航 (初级)
	fmt.Println("\n👆 示例2: 父节点和祖先节点导航 (初级)")
	fmt.Println("展示如何向上遍历节点树")

	// 分析useState的使用情况
	useStateCount := 0
	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if useStateCount >= 3 { // 只分析前3个
			return
		}

		if node.IsIdentifierNode() && strings.TrimSpace(node.GetText()) == "useState" {
			useStateCount++
			fmt.Printf("\nuseState 使用 %d:\n", useStateCount)
			fmt.Printf("  - 位置: 行 %d, 列 %d\n", node.GetStartLineNumber(), node.GetStartColumnNumber())
			fmt.Printf("  - 完整文本: %s\n", node.GetText())

			// 获取父节点
			parent := node.GetParent()
			if parent != nil {
				fmt.Printf("  - 父节点类型: %s\n", parent.GetKind().String())
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
					fmt.Printf("%s → ", ancestor.GetKind().String())
				}
				fmt.Printf("\n")
			}
		}
	})

	// 示例3: 条件祖先查找 (中级)
	fmt.Println("\n🎯 示例3: 条件祖先查找 (中级)")
	fmt.Println("展示如何根据节点类型查找特定祖先")

	// 查找函数声明中的标识符
	var foundFetchUsers = false
	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if foundFetchUsers {
			return
		}
		if node.IsIdentifierNode() && strings.TrimSpace(node.GetText()) == "fetchUsers" {
			// 查找最近的函数声明祖先
			if funcAncestor, ok := node.GetFirstAncestorByKind(tsmorphgo.KindFunctionDeclaration); ok {
				fmt.Printf("找到 fetchUsers 在函数中:\n")
				fmt.Printf("  - 标识符位置: 行 %d\n", node.GetStartLineNumber())
				fmt.Printf("  - 函数位置: 行 %d\n", funcAncestor.GetStartLineNumber())

				// 获取函数名
				if funcName, ok := funcAncestor.GetNodeName(); ok {
					fmt.Printf("  - 函数名: %s\n", funcName)
				}
				foundFetchUsers = true
			}
		}
	})

	// 示例4: 条件遍历和性能优化 (中级)
	fmt.Println("\n⚡ 示例4: 条件遍历和性能优化 (中级)")
	fmt.Println("展示如何提前终止遍历，提高性能")

	// 查找第一个类声明
	var foundClass *tsmorphgo.Node

	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if foundClass != nil {
			return // 已经找到，停止遍历
		}

		if node.IsClassDeclaration() {
			foundClass = &node
			fmt.Printf("✅ 找到第一个类声明:\n")
			fmt.Printf("  - 位置: 行 %d\n", node.GetStartLineNumber())
			fmt.Printf("  - 节点类型: %s\n", node.GetKind().String())

			// 获取类名
			if className, ok := node.GetNodeName(); ok {
				fmt.Printf("  - 类名: %s\n", className)
			}
			return
		}
	})

	if foundClass == nil {
		fmt.Printf("ℹ️ 未找到类声明 (这很正常，因为这是函数组件)\n")
	}

	// 示例5: 精确位置信息 (高级)
	fmt.Println("\n📍 示例5: 精确位置信息 (高级)")
	fmt.Println("展示如何获取节点的详细位置信息")

	positionCount := 0
	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if positionCount >= 5 { // 只演示前5个
			return
		}

		// 重点分析变量声明的位置信息
		if node.IsVariableDeclaration() {
			if varName, ok := node.GetNodeName(); ok && len(varName) > 0 {
				positionCount++
				fmt.Printf("\n位置信息 %d - 变量: '%s'\n", positionCount, varName)

				// 获取各种位置信息
				startPos := node.GetStart()
				lineNumber := node.GetStartLineNumber()
				columnNumber := node.GetStartColumnNumber()
				startLinePos := node.GetStartLinePos()

				fmt.Printf("  - 起始位置(文件偏移): %d\n", startPos)
				fmt.Printf("  - 起始行号: %d\n", lineNumber)
				fmt.Printf("  - 起始列号: %d\n", columnNumber)
				fmt.Printf("  - 行起始位置: %d\n", startLinePos)

				// 计算相对列位置
				relativeColumn := startPos - startLinePos
				fmt.Printf("  - 行内相对位置: %d (0-based)\n", relativeColumn)

				// 验证位置计算的正确性
				fmt.Printf("  - 验证: 列号-1 = %d, 相对位置 = %d, 相等吗? %v\n",
					columnNumber-1, relativeColumn, columnNumber-1 == relativeColumn)
			}
		}
	})

	// 示例6: 函数调用分析
	fmt.Println("\n📞 示例6: 函数调用分析")
	fmt.Println("展示如何分析函数调用和其上下文")

	callCount := 0
	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if callCount >= 5 { // 只分析前5个
			return
		}

		if node.IsCallExpr() {
			callCount++
			text := strings.TrimSpace(node.GetText())
			if len(text) > 30 {
				text = text[:30] + "..."
			}

			fmt.Printf("\n函数调用 %d:\n", callCount)
			fmt.Printf("  - 调用表达式: %s\n", text)
			fmt.Printf("  - 位置: 行 %d, 列 %d\n", node.GetStartLineNumber(), node.GetStartColumnNumber())

			// 分析调用上下文
			parent := node.GetParent()
			if parent != nil {
				fmt.Printf("  - 父节点类型: %s\n", parent.GetKind().String())
			}

			// 获取被调用的函数名 - 使用遍历找到第一个子节点
			node.ForEachDescendant(func(child tsmorphgo.Node) {
				if child.IsIdentifierNode() {
					fmt.Printf("  - 函数名: %s\n", child.GetText())
				}
			})
		}
	})

	// 示例7: 节点层次分析
	fmt.Println("\n🌳 示例7: 节点层次分析")
	fmt.Println("展示如何分析节点的层次结构")

	// 找到export default语句并分析其层次
	var foundExport = false
	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if foundExport {
			return
		}
		if node.IsKind(tsmorphgo.KindExportDeclaration) {
			fmt.Printf("\n找到导出语句:\n")
			fmt.Printf("  - 位置: 行 %d\n", node.GetStartLineNumber())
			fmt.Printf("  - 内容: %s\n", truncateString(node.GetText(), 50))

			// 分析祖先节点，了解语句在文件中的位置
			ancestors := node.GetAncestors()
			fmt.Printf("  - 层次深度: %d\n", len(ancestors))

			// 显示层次路径
			fmt.Printf("  - 层次路径: ")
			pathParts := []string{}
			for i := len(ancestors) - 1; i >= 0; i-- {
				ancestor := ancestors[i]
				pathParts = append(pathParts, ancestor.GetKind().String())
			}
			pathParts = append(pathParts, node.GetKind().String())
			fmt.Printf("%s\n", strings.Join(pathParts, " → "))
			foundExport = true
		}
	})

	// 示例8: 导入导出分析
	fmt.Println("\n📦 示例8: 导入导出分析")
	fmt.Println("展示如何分析模块的导入导出")

	var importCount, exportCount int
	var imports []string
	var exports []string

	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsImportDeclaration() {
			importCount++
			// 提取导入的模块名
			text := strings.TrimSpace(node.GetText())
			imports = append(imports, text)
		}
		if node.IsKind(tsmorphgo.KindExportDeclaration) {
			exportCount++
			text := strings.TrimSpace(node.GetText())
			exports = append(exports, text)
		}
	})

	fmt.Printf("\n📊 模块统计:\n")
	fmt.Printf("  - 导入语句: %d\n", importCount)
	fmt.Printf("  - 导出语句: %d\n", exportCount)

	if len(imports) > 0 {
		fmt.Printf("\n📥 导入模块:\n")
		for i, imp := range imports {
			if i >= 3 { // 只显示前3个
				fmt.Printf("  ... (共%d个导入)\n", len(imports))
				break
			}
			fmt.Printf("  %d. %s\n", i+1, truncateString(imp, 60))
		}
	}

	if len(exports) > 0 {
		fmt.Printf("\n📤 导出内容:\n")
		for i, exp := range exports {
			fmt.Printf("  %d. %s\n", i+1, truncateString(exp, 60))
		}
	}

	fmt.Println("\n🎯 新API使用总结:")
	fmt.Println("1. 节点遍历 → 使用 ForEachDescendant() + 统一的IsXxx()方法")
	fmt.Println("2. 类型检查 → 使用 node.IsKind() 或具体的IsXxx()方法")
	fmt.Println("3. 节点导航 → 使用 GetParent(), GetAncestors(), GetFirstAncestorByKind()")
	fmt.Println("4. 位置信息 → 使用 GetStart(), GetStartLineNumber(), GetStartColumnNumber()")
	fmt.Println("5. 名称提取 → 使用 GetNodeName() 获取节点名称")
	fmt.Println("6. 性能优化 → 在遍历回调中及时 return 终止遍历")

	fmt.Println("\n✅ 节点导航示例完成!")
	fmt.Println("新API大大简化了节点导航的复杂度！")
}

// 辅助函数：截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
