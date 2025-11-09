//go:build type_detection
// +build type_detection

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🏷️ TSMorphGo 类型检测 - 正确使用姿势")
	fmt.Println("=" + repeat("=", 50))

	// =============================================================================
	// 本文件演示 TSMorphGo 类型检测和类型守卫的正确使用方法
	// =============================================================================
	// 学习级别: 初级 → 高级
	// 预计时间: 35-50分钟
	//
	// 功能覆盖:
	// - 基础: 接口、类型别名、函数声明识别
	// - 高级: 复合类型守卫、代码质量分析 ⭐、依赖关系分析 ⭐
	// - 应用: 代码重构、静态分析、IDE功能
	//
	// ⭐ = 高级功能，初学者可先跳过
	//
	// 对齐 ts-morph API:
	// - Node.isInterfaceDeclaration() → IsInterfaceDeclaration()
	// - Node.isTypeAliasDeclaration() → IsTypeAliasDeclaration()
	// - Node.isFunctionDeclaration() → IsFunctionDeclaration()
	// - Node.isCallExpression() → IsCallExpression()
	// - Node.isImportDeclaration() → node.Kind == KindImportDeclaration
	// - Node.isExportDeclaration() → node.Kind == KindExportDeclaration
	// =============================================================================

	// 初始化项目
	realProjectPath := "/Users/bird/Desktop/alalyzer/analyzer-ts/tsmorphgo/examples/demo-react-app"
	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:         realProjectPath,
		TargetExtensions: []string{".ts", ".tsx"},
		IgnorePatterns:   []string{"node_modules", "dist", ".git", "build"},
		UseTsConfig:      true,
	})
	defer project.Close()

	// 选择types.ts作为主要分析文件
	typesFile := project.GetSourceFile(realProjectPath + "/src/types.ts")
	if typesFile == nil {
		log.Fatal("未找到 types.ts 文件")
	}

	fmt.Printf("📄 分析文件: %s\n", typesFile.GetFilePath())
	fmt.Println("=" + repeat("=", 30))

	// 示例1: 基础类型检测 (初级)
	// 对应 ts-morph: Node.isInterfaceDeclaration(), Node.isTypeAliasDeclaration()
	fmt.Println("\n🔍 示例1: 基础类型检测 (初级)")
	fmt.Println("对齐 ts-morph: Node.isInterfaceDeclaration(), Node.isTypeAliasDeclaration()")
	fmt.Println("功能: 识别TypeScript中的类型定义")

	// 统计各种类型定义
	var stats = make(map[string]int)
	var typeDetails []struct {
		kind   string
		name   string
		line   int
		detail string
	}

	typesFile.ForEachDescendant(func(node tsmorphgo.Node) {
		switch {
		// 对应 ts-morph: node.isInterfaceDeclaration()
		case tsmorphgo.IsInterfaceDeclaration(node):
			stats["InterfaceDeclaration"]++
			if stats["InterfaceDeclaration"] <= 5 { // 只记录前5个
				if nameNode, ok := tsmorphgo.GetFirstChild(node, tsmorphgo.IsIdentifier); ok {
					detail := "接口定义"
					// 简单统计接口成员
					propertyCount := 0
					methodCount := 0
					node.ForEachDescendant(func(descendant tsmorphgo.Node) {
						switch descendant.Kind {
						case tsmorphgo.KindPropertySignature:
							propertyCount++
						case tsmorphgo.KindMethodSignature:
							methodCount++
						}
					})
					if propertyCount > 0 || methodCount > 0 {
						detail += fmt.Sprintf(" (属性:%d, 方法:%d)", propertyCount, methodCount)
					}

					typeDetails = append(typeDetails, struct {
						kind   string
						name   string
						line   int
						detail string
					}{
						kind:   "Interface",
						name:   nameNode.GetText(),
						line:   node.GetStartLineNumber(),
						detail: detail,
					})
				}
			}

		// 对应 ts-morph: node.isTypeAliasDeclaration()
		case tsmorphgo.IsTypeAliasDeclaration(node):
			stats["TypeAliasDeclaration"]++
			if stats["TypeAliasDeclaration"] <= 5 {
				if nameNode, ok := tsmorphgo.GetFirstChild(node, tsmorphgo.IsIdentifier); ok {
					text := strings.TrimSpace(node.GetText())
					// 检查是否是泛型
					isGeneric := strings.Contains(text, "<") && strings.Contains(text, ">")
					detail := "类型别名"
					if isGeneric {
						detail += " (泛型)"
					}

					typeDetails = append(typeDetails, struct {
						kind   string
						name   string
						line   int
						detail string
					}{
						kind:   "TypeAlias",
						name:   nameNode.GetText(),
						line:   node.GetStartLineNumber(),
						detail: detail,
					})
				}
			}
		}
	})

	fmt.Printf("📊 类型定义统计:\n")
	for kind, count := range stats {
		switch kind {
		case "InterfaceDeclaration":
			fmt.Printf("  - 接口声明: %d 个\n", count)
		case "TypeAliasDeclaration":
			fmt.Printf("  - 类型别名: %d 个\n", count)
		}
	}

	fmt.Printf("\n📋 详细类型信息:\n")
	for i, detail := range typeDetails {
		fmt.Printf("  %d. %s: %s (行 %d) - %s\n", i+1, detail.kind, detail.name, detail.line, detail.detail)
	}

	// 示例2: 函数和方法的类型检测 (中级)
	fmt.Println("\n⚡ 示例2: 函数和方法的类型检测 (中级)")
	fmt.Println("对齐 ts-morph: Node.isFunctionDeclaration(), Node.isMethodDeclaration()")

	// 分析services/api.ts中的函数
	apiFile := project.GetSourceFile(realProjectPath + "/src/services/api.ts")
	if apiFile != nil {
		fmt.Printf("\n分析 %s 中的函数:\n", extractFileName(apiFile.GetFilePath()))

		var functions []struct {
			name       string
			line       int
			isAsync    bool
			isExported bool
			params     int
		}

		apiFile.ForEachDescendant(func(node tsmorphgo.Node) {
			// 对应 ts-morph: node.isFunctionDeclaration()
			if tsmorphgo.IsFunctionDeclaration(node) {
				if funcName, ok := tsmorphgo.GetFunctionDeclarationNameNode(node); ok {
					// 检查是否是异步函数
					text := strings.ToLower(node.GetText())
					isAsync := strings.Contains(text, "async")

					// 检查是否导出
					parent := node.GetParent()
					isExported := false
					for parent != nil {
						if strings.ToLower(parent.GetText()) == "export" {
							isExported = true
							break
						}
						parent = parent.GetParent()
					}

					// 统计参数数量 (简化统计)
					paramCount := strings.Count(node.GetText(), ",") + 1

					functions = append(functions, struct {
						name       string
						line       int
						isAsync    bool
						isExported bool
						params     int
					}{
						name:       funcName.GetText(),
						line:       node.GetStartLineNumber(),
						isAsync:    isAsync,
						isExported: isExported,
						params:     paramCount,
					})
				}
			}
		})

		for _, fn := range functions {
			fmt.Printf("  - %s", fn.name)
			if fn.isExported {
				fmt.Printf(" (导出)")
			}
			if fn.isAsync {
				fmt.Printf(" (异步)")
			}
			fmt.Printf(" - 行 %d, 参数: %d\n", fn.line, fn.params)
		}

		fmt.Printf("✅ 共找到 %d 个函数\n", len(functions))
	}

	// 示例3: 复合类型守卫 (高级 ⭐)
	// 对应 ts-morph: 组合多个 isXxx() 函数进行复杂判断
	fmt.Println("\n🛡️ 示例3: 复合类型守卫 (高级 ⭐)")
	fmt.Println("对齐 ts-morph: 组合多个 Node.isXxx() 函数")
	fmt.Println("功能: 复杂的类型判断，精确的代码分析")

	// 分析项目中的复杂类型模式
	var complexPatterns []struct {
		description string
		count       int
		examples    []string
	}

	// 确保我们预先创建所有需要的模式
	complexPatterns = append(complexPatterns, struct {
		description string
		count       int
		examples    []string
	}{description: "导出接口", count: 0, examples: []string{}})

	complexPatterns = append(complexPatterns, struct {
		description string
		count       int
		examples    []string
	}{description: "异步函数", count: 0, examples: []string{}})

	complexPatterns = append(complexPatterns, struct {
		description string
		count       int
		examples    []string
	}{description: "对象方法调用", count: 0, examples: []string{}})

	// 查找导出的接口
	for _, file := range project.GetSourceFiles() {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// 导出接口: IsInterfaceDeclaration + 父节点有export
			if tsmorphgo.IsInterfaceDeclaration(node) {
				parent := node.GetParent()
				for parent != nil {
					parentText := strings.ToLower(parent.GetText())
					if strings.Contains(parentText, "export") {
						if nameNode, ok := tsmorphgo.GetFirstChild(node, tsmorphgo.IsIdentifier); ok {
							complexPatterns[0].count++
							if len(complexPatterns[0].examples) < 5 {
								complexPatterns[0].examples = append(complexPatterns[0].examples,
									fmt.Sprintf("%s (行 %d)", nameNode.GetText(), node.GetStartLineNumber()))
							}
						}
					}
					break
				}
			}

			// 异步函数: IsFunctionDeclaration + 异步关键字
			if tsmorphgo.IsFunctionDeclaration(node) {
				text := strings.ToLower(node.GetText())
				if strings.Contains(text, "async") {
					if funcName, ok := tsmorphgo.GetFunctionDeclarationNameNode(node); ok {
						complexPatterns[1].count++
						if len(complexPatterns[1].examples) < 5 {
							complexPatterns[1].examples = append(complexPatterns[1].examples,
								fmt.Sprintf("%s (行 %d)", funcName.GetText(), node.GetStartLineNumber()))
						}
					}
				}
			}

			// 对象方法调用: IsCallExpression + GetPropertyAccessExpression
			if tsmorphgo.IsCallExpression(node) {
				if expr, ok := tsmorphgo.GetCallExpressionExpression(node); ok {
					if tsmorphgo.IsPropertyAccessExpression(*expr) {
						complexPatterns[2].count++
						if len(complexPatterns[2].examples) < 5 {
							callText := strings.TrimSpace(node.GetText())
							if len(callText) > 30 {
								callText = callText[:27] + "..."
							}
							complexPatterns[2].examples = append(complexPatterns[2].examples,
								fmt.Sprintf("%s (行 %d)", callText, node.GetStartLineNumber()))
						}
					}
				}
			}
		})
	}

	fmt.Printf("🔍 复合类型模式分析:\n")
	for _, pattern := range complexPatterns {
		fmt.Printf("  - %s: %d 个\n", pattern.description, pattern.count)
		if len(pattern.examples) > 0 {
			for i, example := range pattern.examples {
				if i >= 3 { // 只显示前3个
					fmt.Printf("    %d. %s\n", i+1, example)
				}
				if len(pattern.examples) > 3 {
					fmt.Printf("    ... 还有 %d 个\n", len(pattern.examples)-3)
				}
			}
		}
	}

	// 示例4: 代码质量分析 (高级 ⭐)
	fmt.Println("\n📊 示例4: 代码质量分析 (高级 ⭐)")
	fmt.Println("应用: 静态代码分析、质量检查、重构建议")

	var qualityIssues []struct {
		issueType string
		location  string
		details   string
		file      string
	}

	// 分析所有文件
	for _, file := range project.GetSourceFiles() {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// 长函数检测
			if tsmorphgo.IsFunctionDeclaration(node) {
				funcText := node.GetText()
				if len(funcText) > 500 { // 超过500字符的函数
					if funcName, ok := tsmorphgo.GetFunctionDeclarationNameNode(node); ok {
						qualityIssues = append(qualityIssues, struct {
							issueType string
							location  string
							details   string
							file      string
						}{
							issueType: "长函数",
							location:  fmt.Sprintf("行 %d", node.GetStartLineNumber()),
							details:   fmt.Sprintf("函数 '%s' 过长 (%d 字符)", funcName.GetText(), len(funcText)),
							file:      extractFileName(file.GetFilePath()),
						})
					}
				}
			}

			// 深层嵌套检测
			ancestors := node.GetAncestors()
			if len(ancestors) > 15 { // 嵌套深度超过15层
				if tsmorphgo.IsIdentifier(node) || tsmorphgo.IsCallExpression(node) {
					nodeText := strings.TrimSpace(node.GetText())
					if len(nodeText) > 0 && len(nodeText) < 50 {
						qualityIssues = append(qualityIssues, struct {
							issueType string
							location  string
							details   string
							file      string
						}{
							issueType: "深层嵌套",
							location:  fmt.Sprintf("行 %d", node.GetStartLineNumber()),
							details:   fmt.Sprintf("嵌套深度: %d 层", len(ancestors)),
							file:      extractFileName(file.GetFilePath()),
						})
					}
				}
			}

			// 复杂条件表达式检测
			if tsmorphgo.IsBinaryExpression(node) {
				nodeText := strings.ToLower(node.GetText())
				andCount := strings.Count(nodeText, "&&")
				orCount := strings.Count(nodeText, "||")
				if andCount+orCount > 4 { // 超过4个逻辑操作符
					qualityIssues = append(qualityIssues, struct {
						issueType string
						location  string
						details   string
						file      string
					}{
						issueType: "复杂条件",
						location:  fmt.Sprintf("行 %d", node.GetStartLineNumber()),
						details:   fmt.Sprintf("逻辑操作符数量: %d (AND: %d, OR: %d)", andCount+orCount, andCount, orCount),
						file:      extractFileName(file.GetFilePath()),
					})
				}
			}
		})
	}

	fmt.Printf("🔍 代码质量问题分析:\n")
	if len(qualityIssues) == 0 {
		fmt.Printf("  ✅ 未发现明显的代码质量问题\n")
	} else {
		fmt.Printf("  ⚠️  发现 %d 个潜在质量问题:\n", len(qualityIssues))

		// 按类型分组显示
		issueTypes := make(map[string][]struct {
			location string
			details  string
			file     string
		})

		for _, issue := range qualityIssues {
			issueTypes[issue.issueType] = append(issueTypes[issue.issueType], struct {
				location string
				details  string
				file     string
			}{
				location: issue.location,
				details:  issue.details,
				file:     issue.file,
			})
		}

		for issueType, issues := range issueTypes {
			fmt.Printf("  - %s (%d个):\n", issueType, len(issues))
			for i, issue := range issues {
				if i >= 3 { // 只显示前3个
					fmt.Printf("    %d. %s - %s (%s)\n", i+1, issue.file, issue.location, issue.details)
				}
				if len(issues) > 3 {
					fmt.Printf("    ... 还有 %d 个\n", len(issues)-3)
				}
			}
		}
	}

	// 示例5: 依赖关系分析 (高级 ⭐)
	fmt.Println("\n🔗 示例5: 依赖关系分析 (高级 ⭐)")
	fmt.Println("应用: 模块依赖图、循环依赖检测、重构影响分析")

	// 分析导入依赖
	importDependencies := make(map[string][]string)
	for _, file := range project.GetSourceFiles() {
		fileName := extractFileName(file.GetFilePath())
		var imports []string

		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// 对应 ts-morph: node.kind === SyntaxKind.ImportDeclaration
			if node.Kind == tsmorphgo.KindImportDeclaration {
				importText := strings.TrimSpace(node.GetText())
				// 提取导入路径
				if strings.Contains(importText, "from") {
					parts := strings.Split(importText, "from")
					if len(parts) >= 2 {
						importPath := strings.TrimSpace(parts[1])
						importPath = strings.Trim(importPath, `"'`)
						// 只保留相对路径导入
						if strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") {
							imports = append(imports, importPath)
						}
					}
				}
			}
		})

		if len(imports) > 0 {
			importDependencies[fileName] = imports
		}
	}

	fmt.Printf("📦 模块导入依赖:\n")
	for file, deps := range importDependencies {
		fmt.Printf("  - %s 依赖于:\n", file)
		for _, dep := range deps {
			fmt.Printf("    - %s\n", dep)
		}
	}

	// 分析类型依赖
	typeDependencies := make(map[string][]string)
	for _, file := range project.GetSourceFiles() {
		fileName := extractFileName(file.GetFilePath())
		var types []string

		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// 查找类型引用
			if tsmorphgo.IsIdentifier(node) {
				// 简单启发式：检查是否可能是类型名称
				text := strings.TrimSpace(node.GetText())
				if isTypeName(text) {
					// 避免重复
					found := false
					for _, t := range types {
						if t == text {
							found = true
							break
						}
					}
					if !found {
						types = append(types, text)
					}
				}
			}
		})

		if len(types) > 5 { // 只显示类型依赖较多的文件
			typeDependencies[fileName] = types[:5]
		}
	}

	fmt.Printf("\n🏷️ 类型依赖分析 (前5个文件):\n")
	for file, types := range typeDependencies {
		fmt.Printf("  - %s 使用类型: %s\n", file, strings.Join(types, ", "))
	}

	fmt.Println("\n🎯 类型检测使用姿势总结:")
	fmt.Println("1. 类型识别 → 使用 IsXxx() 系列函数 (IsInterfaceDeclaration 等)")
	fmt.Println("2. 复合判断 → 组合多个类型检查函数")
	fmt.Println("3. 节点种类 → 使用 node.Kind == KindXxx 进行精确匹配")
	fmt.Println("4. 性能优化 → 在回调中提前 return 避免无效遍历")
	fmt.Println("5. 实际应用 → 代码质量分析、依赖关系分析、重构建议")

	fmt.Println("\n✅ 类型检测示例完成!")
}

// 辅助函数：重复字符串
func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// 辅助函数：提取文件名
func extractFileName(filePath string) string {
	parts := strings.Split(filePath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return filePath
}

// 辅助函数：判断是否是类型名称（启发式）
func isTypeName(text string) bool {
	// 简单的启发式规则
	if len(text) <= 1 {
		return false
	}

	// 检查是否以大写字母开头 (PascalCase)
	if text[0] >= 'A' && text[0] <= 'Z' {
		return true
	}

	// 检查是否是常见的类型模式
	commonTypes := []string{
		"User", "UserType", "AppConfig", "Response", "Request", "Service", "Controller",
		"Model", "Entity", "DTO", "Interface", "Type", "Enum", "Class",
	}

	for _, commonType := range commonTypes {
		if text == commonType || strings.HasPrefix(text, commonType) {
			return true
		}
	}

	return false
}
