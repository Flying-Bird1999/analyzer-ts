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
	fmt.Println("🔗 TSMorphGo 引用查找 - 新API演示")
	fmt.Println("=" + strings.Repeat("=", 50))

	// =============================================================================
	// 本文件演示新的统一API在引用查找和符号分析中的应用
	// =============================================================================
	// 学习级别: 中级 → 高级
	// 预计时间: 15-20分钟
	//
	// 新API的优势:
	// - 统一的接口设计，简化引用查找逻辑
	// - 更好的错误处理和调试信息
	// - 简化的符号访问接口
	// - 性能优化的遍历机制
	//
	// 新API功能:
	// - node.IsIdentifierNode() → 标识符检查
	// - node.IsVariableDeclaration() → 变量声明检查
	// - node.GetNodeName() → 获取节点名称
	// - node.IsMethodDeclaration() → 方法声明检查 (使用IsKind)
	// =============================================================================

	// 使用真实的demo-react-app项目
	realProjectPath, err := filepath.Abs("../demo-react-app")
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

	// 获取所有源文件
	files := project.GetSourceFiles()
	fmt.Printf("📄 项目文件数量: %d\n", len(files))

	for _, file := range files {
		fmt.Printf("  - %s\n", file.GetFilePath())
	}

	// 示例1: 基础引用查找 (中级)
	fmt.Println("\n🔍 示例1: 基础引用查找 (中级)")
	fmt.Println("展示如何查找符号在项目中的所有引用")

	// 查找User接口的引用
	typesFile := project.GetSourceFile(realProjectPath + "/src/types.ts")
	if typesFile == nil {
		fmt.Println("❌ 未找到 types.ts 文件")
		return
	}

	var userInterfaceNode *tsmorphgo.Node
	var foundInterfaces []string
	typesFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsKind(tsmorphgo.KindInterfaceDeclaration) {
			if name, ok := node.GetNodeName(); ok {
				foundInterfaces = append(foundInterfaces, name)
				fmt.Printf("  找到接口: %s\n", name)
				if name == "User" {
					userInterfaceNode = &node
				}
			}
		}
	})

	fmt.Printf("📋 找到的所有接口: %v\n", foundInterfaces)

	if userInterfaceNode == nil {
		fmt.Println("❌ 未找到 User 接口")
		return
	}

	fmt.Printf("✅ 找到 User 接口定义:\n")
	fmt.Printf("  - 位置: %s:%d\n", userInterfaceNode.GetSourceFile().GetFilePath(), userInterfaceNode.GetStartLineNumber())
	fmt.Printf("  - 节点类型: %s\n", userInterfaceNode.GetKind().String())

	// 计时引用查找
	start := time.Now()

	// 在所有文件中搜索User的引用
	var userReferences []struct {
		file   string
		line   int
		text   string
		node   tsmorphgo.Node
	}

	for _, file := range files {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsIdentifierNode() && node.GetText() == "User" {
				// 排除定义本身
				if node.GetStart() != userInterfaceNode.GetStart() {
					userReferences = append(userReferences, struct {
						file   string
						line   int
						text   string
						node   tsmorphgo.Node
					}{
						file:   file.GetFilePath(),
						line:   node.GetStartLineNumber(),
						text:   truncateString(node.GetText(), 50),
						node:   node,
					})
				}
			}
		})
	}

	duration := time.Since(start)

	fmt.Printf("\n📊 User 引用统计:\n")
	fmt.Printf("  - 总引用数: %d\n", len(userReferences))
	fmt.Printf("  - 查找耗时: %v\n", duration)

	// 按文件分组显示引用
	referencesByFile := make(map[string][]struct {
		line int
		text string
		node tsmorphgo.Node
	})

	for _, ref := range userReferences {
		referencesByFile[ref.file] = append(referencesByFile[ref.file], struct {
			line int
			text string
			node tsmorphgo.Node
		}{
			line: ref.line,
			text: ref.text,
			node: ref.node,
		})
	}

	fmt.Printf("\n📁 按文件分组的引用:\n")
	for filePath, refs := range referencesByFile {
		fmt.Printf("  📄 %s (%d个引用)\n", filepath.Base(filePath), len(refs))
		for i, ref := range refs {
			if i >= 3 { // 只显示前3个
				fmt.Printf("    ... (还有%d个)\n", len(refs)-3)
				break
			}
			fmt.Printf("    %d. 行%d: %s\n", i+1, ref.line, ref.text)
		}
	}

	// 示例2: 变量引用分析 (中级)
	fmt.Println("\n🎯 示例2: 变量引用分析 (中级)")
	fmt.Println("展示如何分析变量的使用情况")

	// 分析useState的使用
	var useStateUsages []struct {
		file string
		line int
		text string
	}

	for _, file := range files {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsIdentifierNode() && node.GetText() == "useState" {
				useStateUsages = append(useStateUsages, struct {
					file string
					line int
					text string
				}{
					file: file.GetFilePath(),
					line: node.GetStartLineNumber(),
					text: extractContext(node, 40),
				})
			}
		})
	}

	fmt.Printf("\n📊 useState 使用统计:\n")
	fmt.Printf("  - 使用次数: %d\n", len(useStateUsages))

	for _, usage := range useStateUsages {
		fmt.Printf("  📄 %s:%d\n", filepath.Base(usage.file), usage.line)
		fmt.Printf("    代码: %s\n", usage.text)
	}

	// 示例3: 函数调用链分析 (高级)
	fmt.Println("\n🔗 示例3: 函数调用链分析 (高级)")
	fmt.Println("展示如何分析函数的调用关系")

	// 分析fetchUsers函数的调用
	var fetchUsages []struct {
		file     string
		line     int
		callExpr string
		caller   string
	}

	for _, file := range files {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsIdentifierNode() && node.GetText() == "fetchUsers" {
				// 检查是否是函数调用
				parent := node.GetParent()
				if parent != nil && parent.IsCallExpr() {
					// 找到调用上下文
					callerInfo := findCallerContext(node)
					fetchUsages = append(fetchUsages, struct {
						file     string
						line     int
						callExpr string
						caller   string
					}{
						file:     file.GetFilePath(),
						line:     node.GetStartLineNumber(),
						callExpr: truncateString(parent.GetText(), 50),
						caller:   callerInfo,
					})
				}
			}
		})
	}

	fmt.Printf("\n📊 fetchUsers 调用分析:\n")
	fmt.Printf("  - 调用次数: %d\n", len(fetchUsages))

	for _, call := range fetchUsages {
		fmt.Printf("  📄 %s:%d\n", filepath.Base(call.file), call.line)
		fmt.Printf("    调用者: %s\n", call.caller)
		fmt.Printf("    表达式: %s\n", call.callExpr)
	}

	// 示例4: 属性访问分析 (高级)
	fmt.Println("\n🏷️ 示例4: 属性访问分析 (高级)")
	fmt.Println("展示如何分析对象属性的使用")

	// 分析_.toUpper和_.filter的使用
	var lodashUsages []struct {
		file      string
		line      int
		method    string
		context   string
	}

	for _, file := range files {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsPropertyAccessExpression() {
				// 检查是否访问了lodash的方法
				node.ForEachDescendant(func(child tsmorphgo.Node) {
					if child.IsIdentifierNode() && child.GetText() == "_" {
						// 获取完整的属性访问表达式
						fullExpr := node.GetText()
						if strings.Contains(fullExpr, "_.") {
							// 提取方法名
							parts := strings.Split(fullExpr, ".")
							if len(parts) >= 2 {
								method := parts[len(parts)-1]

								lodashUsages = append(lodashUsages, struct {
									file      string
									line      int
									method    string
									context   string
								}{
									file:    file.GetFilePath(),
									line:    node.GetStartLineNumber(),
									method:  method,
									context: extractContext(node, 30),
								})
							}
						}
					}
				})
			}
		})
	}

	fmt.Printf("\n📊 lodash 方法使用分析:\n")
	fmt.Printf("  - 使用次数: %d\n", len(lodashUsages))

	// 按方法分组
	usageByMethod := make(map[string][]struct {
		file    string
		line    int
		context string
	})

	for _, usage := range lodashUsages {
		usageByMethod[usage.method] = append(usageByMethod[usage.method], struct {
			file    string
			line    int
			context string
		}{
			file:    usage.file,
			line:    usage.line,
			context: usage.context,
		})
	}

	for method, usages := range usageByMethod {
		fmt.Printf("\n  🔸 _.%s (%d次使用)\n", method, len(usages))
		for _, usage := range usages {
			fmt.Printf("    📄 %s:%d\n", filepath.Base(usage.file), usage.line)
			fmt.Printf("       %s\n", usage.context)
		}
	}

	// 示例5: 导入导出分析 (高级)
	fmt.Println("\n📦 示例5: 导入导出分析 (高级)")
	fmt.Println("展示如何分析模块间的依赖关系")

	var importAnalysis []struct {
		importer string
		imported string
		items    []string
	}

	for _, file := range files {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsImportDeclaration() {
				// 分析导入语句
				importText := node.GetText()
				if strings.Contains(importText, "from") {
					parts := strings.Split(importText, "from")
					if len(parts) == 2 {
						importer := file.GetFilePath()
						imported := strings.TrimSpace(strings.Trim(parts[1], `'"`))

						// 提取导入项
						importItems := extractImportItems(parts[0])

						importAnalysis = append(importAnalysis, struct {
							importer string
							imported string
							items    []string
						}{
							importer: importer,
							imported: imported,
							items:    importItems,
						})
					}
				}
			}
		})
	}

	fmt.Printf("\n📊 模块依赖分析:\n")
	fmt.Printf("  - 导入关系: %d 个\n", len(importAnalysis))

	fmt.Printf("\n📥 导入详情:\n")
	for _, imp := range importAnalysis {
		fmt.Printf("  📄 %s → %s\n", filepath.Base(imp.importer), imp.imported)
		for _, item := range imp.items {
			fmt.Printf("    - %s\n", item)
		}
	}

	// 示例6: 类型引用分析 (高级)
	fmt.Println("\n🎭 示例6: 类型引用分析 (高级)")
	fmt.Println("展示如何分析类型的使用情况")

	// 分析interface的使用
	var interfaceUsages []struct {
		file    string
		line    int
		iface   string
		context string
		usage   string // "type_annotation", "extends", "implements"
	}

	for _, file := range files {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsIdentifierNode() {
				text := node.GetText()
				// 检查是否是已知接口
				if text == "User" || text == "Msg" || text == "ApiResponse" {
					// 确定使用类型
					usageType := "identifier"
					parent := node.GetParent()
					if parent != nil {
						if parent.IsKind(tsmorphgo.KindTypeReference) {
							usageType = "type_annotation"
						}
					}

					interfaceUsages = append(interfaceUsages, struct {
						file    string
						line    int
						iface   string
						context string
						usage   string
					}{
						file:    file.GetFilePath(),
						line:    node.GetStartLineNumber(),
						iface:   text,
						context: extractContext(node, 30),
						usage:   usageType,
					})
				}
			}
		})
	}

	fmt.Printf("\n📊 接口使用分析:\n")
	fmt.Printf("  - 使用次数: %d\n", len(interfaceUsages))

	// 按接口分组
	usageByInterface := make(map[string][]struct {
		file    string
		line    int
		context string
		usage   string
	})

	for _, usage := range interfaceUsages {
		usageByInterface[usage.iface] = append(usageByInterface[usage.iface], struct {
			file    string
			line    int
			context string
			usage   string
		}{
			file:    usage.file,
			line:    usage.line,
			context: usage.context,
			usage:   usage.usage,
		})
	}

	for iface, usages := range usageByInterface {
		fmt.Printf("\n  🔸 %s 接口 (%d次使用)\n", iface, len(usages))
		for _, usage := range usages {
			fmt.Printf("    📄 %s:%d [%s]\n", filepath.Base(usage.file), usage.line, usage.usage)
			fmt.Printf("       %s\n", usage.context)
		}
	}

	fmt.Println("\n🎯 新API使用总结:")
	fmt.Println("1. 符号查找 → 使用 ForEachDescendant() + IsIdentifierNode()")
	fmt.Println("2. 类型分析 → 使用 IsInterfaceDeclaration(), IsKind() 等")
	fmt.Println("3. 引用计数 → 遍历所有文件统计标识符使用")
	fmt.Println("4. 调用链分析 → 结合 IsCallExpr() 和上下文分析")
	fmt.Println("5. 模块分析 → 使用 IsImportDeclaration(), IsKind()")
	fmt.Println("6. 属性访问 → 使用 IsPropertyAccessExpression()")

	fmt.Println("\n✅ 引用查找示例完成!")
	fmt.Println("新API让符号分析变得简单高效！")
}

// 辅助函数：截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// 辅助函数：提取代码上下文
func extractContext(node tsmorphgo.Node, maxLen int) string {
	text := node.GetText()
	if len(text) > maxLen {
		text = text[:maxLen] + "..."
	}
	return text
}

// 辅助函数：查找调用者上下文
func findCallerContext(callNode tsmorphgo.Node) string {
	parent := callNode.GetParent()
	for parent != nil {
		if parent.IsFunctionDeclaration() || parent.IsKind(tsmorphgo.KindMethodDeclaration) {
			if name, ok := parent.GetNodeName(); ok {
				return name
			}
		}
		if parent.IsVariableDeclaration() {
			if name, ok := parent.GetNodeName(); ok {
				return "匿名函数 (变量: " + name + ")"
			}
		}
		parent = parent.GetParent()
	}
	return "全局作用域"
}

// 辅助函数：提取导入项
func extractImportItems(importClause string) []string {
	items := []string{}
	importClause = strings.TrimSpace(importClause)

	if strings.HasPrefix(importClause, "import") {
		importClause = strings.TrimSpace(importClause[6:])
	}

	if strings.HasPrefix(importClause, "{") {
		// 具名导入
		importClause = strings.Trim(importClause, "{}")
		parts := strings.Split(importClause, ",")
		for _, part := range parts {
			item := strings.TrimSpace(strings.Split(part, " as ")[0])
			if item != "" {
				items = append(items, item)
			}
		}
	} else {
		// 默认导入
		item := strings.Split(importClause, " as ")[0]
		item = strings.TrimSpace(item)
		if item != "" {
			items = append(items, item)
		}
	}

	return items
}