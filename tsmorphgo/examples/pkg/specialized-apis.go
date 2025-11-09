//go:build specialized_apis
// +build specialized_apis

package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🛠️ TSMorphGo 专用API - 新API演示")
	fmt.Println("=" + strings.Repeat("=", 50))

	// =============================================================================
	// 本文件演示新的统一API在专用场景中的应用
	// =============================================================================
	// 学习级别: 中级 → 高级
	// 预计时间: 15-20分钟
	//
	// 新API的优势:
	// - 统一的接口设计，无需记忆大量专用函数
	// - 支持类别检查，简化类型判断
	// - 更好的错误处理和调试信息
	// - 性能优化的遍历机制
	//
	// 新API功能:
	// - node.IsFunctionDeclaration() → 函数声明检查
	// - node.IsCallExpr() → 函数调用检查
	// - node.IsPropertyAccessExpression() → 属性访问检查
	// - node.IsVariableDeclaration() → 变量声明检查
	// - node.IsKind() → 精确类型检查
	// - node.GetNodeName() → 获取节点名称
	// =============================================================================

	// 使用真实的demo-react-app项目
	realProjectPath, err := filepath.Abs("../demo-react-app")
	if err != nil {
		fmt.Printf("无法解析项目路径: %v\n", err)
		return
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

	sourceFiles := project.GetSourceFiles()
	if len(sourceFiles) == 0 {
		fmt.Println("❌ 未找到任何源文件")
		return
	}

	fmt.Printf("📊 项目统计: %d 个TypeScript文件\n", len(sourceFiles))

	// 示例1: 函数声明处理 (中级)
	fmt.Println("\n🔧 示例1: 函数声明处理 (中级)")
	fmt.Println("展示如何使用新API识别和分析函数声明")

	var functions []struct {
		name       string
		line       int
		isExported bool
		file       string
	}

	totalFunctions := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// 使用新API检查函数声明
			if node.IsFunctionDeclaration() {
				totalFunctions++
				// 使用新API获取函数名
				if funcName, ok := node.GetNodeName(); ok {
					if totalFunctions <= 8 { // 显示前8个
						// 检查是否导出
						isExported := strings.HasPrefix(strings.TrimSpace(node.GetText()), "export")

						functions = append(functions, struct {
							name       string
							line       int
							isExported bool
							file       string
						}{
							name:       funcName,
							line:       node.GetStartLineNumber(),
							isExported: isExported,
							file:       extractFileName(file.GetFilePath()),
						})
					}
				}
			}
		})
	}

	fmt.Printf("📊 函数声明统计: 找到 %d 个函数\n", totalFunctions)
	if len(functions) > 0 {
		fmt.Printf("前 %d 个函数:\n", len(functions))
		for i, fn := range functions {
			fmt.Printf("  %d. %s() - 行 %d - %s - %s\n",
				i+1, fn.name, fn.line, fn.file,
				map[bool]string{true: "导出", false: "内部"}[fn.isExported])
		}
	}

	// 示例2: 调用表达式分析 (中级)
	fmt.Println("\n📞 示例2: 调用表达式分析 (中级)")
	fmt.Println("展示如何分析函数调用表达式")

	var calls []struct {
		expr    string
		line    int
		file    string
		context string
	}

	totalCalls := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// 使用新API检查函数调用
			if node.IsCallExpr() {
				totalCalls++
				if totalCalls <= 10 { // 显示前10个
					expr := node.GetText()
					if len(expr) > 40 {
						expr = expr[:40] + "..."
					}

					// 获取调用上下文
					parent := node.GetParent()
					context := "表达式"
					if parent != nil {
						if parent.IsVariableDeclaration() {
							context = "变量声明"
						} else if parent.IsKind(tsmorphgo.KindReturnStatement) {
							context = "返回语句"
						} else if parent.IsKind(tsmorphgo.KindBinaryExpression) {
							context = "赋值表达式"
						}
					}

					calls = append(calls, struct {
						expr    string
						line    int
						file    string
						context string
					}{
						expr:    expr,
						line:    node.GetStartLineNumber(),
						file:    extractFileName(file.GetFilePath()),
						context: context,
					})
				}
			}
		})
	}

	fmt.Printf("📊 函数调用统计: 找到 %d 个调用\n", totalCalls)
	if len(calls) > 0 {
		fmt.Printf("前 %d 个调用:\n", len(calls))
		for i, call := range calls {
			fmt.Printf("  %d. %s - 行 %d - %s - %s\n",
				i+1, call.expr, call.line, call.file, call.context)
		}
	}

	// 示例3: 属性访问表达式分析 (中级)
	fmt.Println("\n🔗 示例3: 属性访问表达式分析 (中级)")
	fmt.Println("展示如何分析对象属性访问")

	var propertyAccess []struct {
		object  string
		property string
		line     int
		file     string
	}

	totalPropertyAccess := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// 使用新API检查属性访问
			if node.IsPropertyAccessExpression() {
				totalPropertyAccess++
				if totalPropertyAccess <= 15 { // 显示前15个
					expr := node.GetText()
					parts := strings.Split(expr, ".")
					if len(parts) >= 2 {
						object := parts[0]
						property := strings.Join(parts[1:], ".")

						propertyAccess = append(propertyAccess, struct {
							object   string
							property string
							line     int
							file     string
						}{
							object:   object,
							property: property,
							line:     node.GetStartLineNumber(),
							file:     extractFileName(file.GetFilePath()),
						})
					}
				}
			}
		})
	}

	fmt.Printf("📊 属性访问统计: 找到 %d 个访问\n", totalPropertyAccess)
	if len(propertyAccess) > 0 {
		fmt.Printf("前 %d 个属性访问:\n", len(propertyAccess))
		for i, access := range propertyAccess {
			fmt.Printf("  %d. %s.%s - 行 %d - %s\n",
				i+1, access.object, access.property, access.line, access.file)
		}
	}

	// 示例4: 变量声明分析 (中级)
	fmt.Println("\n📦 示例4: 变量声明分析 (中级)")
	fmt.Println("展示如何分析变量声明")

	var variables []struct {
		name     string
		typeHint string
		line     int
		file     string
	}

	totalVariables := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// 使用新API检查变量声明
			if node.IsVariableDeclaration() {
				totalVariables++
				if totalVariables <= 12 { // 显示前12个
					if varName, ok := node.GetNodeName(); ok {
						// 尝试提取类型信息
						typeHint := "any"
						parent := node.GetParent()
						if parent != nil {
							parentText := parent.GetText()
							if strings.Contains(parentText, ":") {
								// 简单的类型提取
								parts := strings.Split(parentText, ":")
								if len(parts) >= 2 {
									typePart := strings.TrimSpace(parts[1])
									if idx := strings.Index(typePart, "="); idx != -1 {
										typeHint = strings.TrimSpace(typePart[:idx])
									} else {
										typeHint = strings.Split(typePart, ";")[0]
										typeHint = strings.TrimSpace(typeHint)
									}
								}
							}
						}

						variables = append(variables, struct {
							name     string
							typeHint string
							line     int
							file     string
						}{
							name:     varName,
							typeHint: typeHint,
							line:     node.GetStartLineNumber(),
							file:     extractFileName(file.GetFilePath()),
						})
					}
				}
			}
		})
	}

	fmt.Printf("📊 变量声明统计: 找到 %d 个变量\n", totalVariables)
	if len(variables) > 0 {
		fmt.Printf("前 %d 个变量:\n", len(variables))
		for i, v := range variables {
			fmt.Printf("  %d. %s: %s - 行 %d - %s\n",
				i+1, v.name, v.typeHint, v.line, v.file)
		}
	}

	// 示例5: 类型声明分析 (高级)
	fmt.Println("\n🏷️ 示例5: 类型声明分析 (高级)")
	fmt.Println("展示如何分析接口和类型别名")

	var typeDeclarations []struct {
		kind    string // "interface" 或 "type"
		name    string
		line    int
		file    string
		members int
	}

	totalTypes := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// 使用新API检查接口声明
			if node.IsKind(tsmorphgo.KindInterfaceDeclaration) {
				totalTypes++
				if typeName, ok := node.GetNodeName(); ok {
					// 计算成员数量
					memberCount := 0
					node.ForEachDescendant(func(child tsmorphgo.Node) {
						if child.IsKind(tsmorphgo.KindPropertySignature) || child.IsKind(tsmorphgo.KindMethodSignature) {
							memberCount++
						}
					})

					typeDeclarations = append(typeDeclarations, struct {
						kind    string
						name    string
						line    int
						file    string
						members int
					}{
						kind:    "interface",
						name:    typeName,
						line:    node.GetStartLineNumber(),
						file:    extractFileName(file.GetFilePath()),
						members: memberCount,
					})
				}
			}

			// 使用新API检查类型别名
			if node.IsKind(tsmorphgo.KindTypeAliasDeclaration) {
				totalTypes++
				if typeName, ok := node.GetNodeName(); ok {
					typeDeclarations = append(typeDeclarations, struct {
						kind    string
						name    string
						line    int
						file    string
						members int
					}{
						kind:    "type",
						name:    typeName,
						line:    node.GetStartLineNumber(),
						file:    extractFileName(file.GetFilePath()),
						members: 0,
					})
				}
			}
		})
	}

	fmt.Printf("📊 类型声明统计: 找到 %d 个类型\n", totalTypes)
	if len(typeDeclarations) > 0 {
		fmt.Printf("类型声明详情:\n")
		for i, td := range typeDeclarations {
			membersInfo := ""
			if td.kind == "interface" && td.members > 0 {
				membersInfo = fmt.Sprintf(" (%d个成员)", td.members)
			}
			fmt.Printf("  %d. %s %s%s - 行 %d - %s\n",
				i+1, td.kind, td.name, membersInfo, td.line, td.file)
		}
	}

	// 示例6: 导入语句分析 (高级)
	fmt.Println("\n📥 示例6: 导入语句分析 (高级)")
	fmt.Println("展示如何分析模块导入")

	var imports []struct {
		source     string
		items      []string
		line       int
		file       string
		importType string // "default", "named", "namespace", "side-effect"
	}

	totalImports := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// 使用新API检查导入声明
			if node.IsImportDeclaration() {
				totalImports++
				importText := node.GetText()

				// 分析导入类型
				importType := "named"
				if strings.Contains(importText, "import * as") {
					importType = "namespace"
				} else if strings.Contains(importText, "import") && !strings.Contains(importText, "{") && !strings.Contains(importText, "*") {
					importType = "default"
				} else if strings.Contains(importText, "import") && !strings.Contains(importText, "from") {
					importType = "side-effect"
				}

				// 提取导入源
				source := ""
				items := []string{}
				if strings.Contains(importText, "from") {
					parts := strings.Split(importText, "from")
					if len(parts) == 2 {
						source = strings.TrimSpace(strings.Trim(parts[1], `'"`))
						items = extractImportItems(parts[0])
					}
				}

				imports = append(imports, struct {
					source     string
					items      []string
					line       int
					file       string
					importType string
				}{
					source:     source,
					items:      items,
					line:       node.GetStartLineNumber(),
					file:       extractFileName(file.GetFilePath()),
					importType: importType,
				})
			}
		})
	}

	fmt.Printf("📊 导入语句统计: 找到 %d 个导入\n", totalImports)
	if len(imports) > 0 {
		fmt.Printf("导入语句详情:\n")
		for i, imp := range imports {
			itemsStr := ""
			if len(imp.items) > 0 {
				if len(imp.items) <= 3 {
					itemsStr = fmt.Sprintf(" [%s]", strings.Join(imp.items, ", "))
				} else {
					itemsStr = fmt.Sprintf(" [%s, ... (%d more)]", strings.Join(imp.items[:3], ", "), len(imp.items)-3)
				}
			}
			fmt.Printf("  %d. %s %s%s - 行 %d - %s\n",
				i+1, imp.importType, imp.source, itemsStr, imp.line, imp.file)
		}
	}

	// 示例7: 控制流分析 (高级)
	fmt.Println("\n🌊 示例7: 控制流分析 (高级)")
	fmt.Println("展示如何分析程序的控制流结构")

	var controlFlow []struct {
		kind      string
		condition string
		line      int
		file      string
	}

	totalControlFlow := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// 分析不同类型的控制流
			kind := ""
			condition := ""

			if node.IsKind(tsmorphgo.KindIfStatement) {
				kind = "if"
				// 提取条件
				condition = truncateString(node.GetText(), 30)
			} else if node.IsKind(tsmorphgo.KindForStatement) {
				kind = "for"
				// 提取循环条件
				condition = truncateString(node.GetText(), 30)
			} else if node.IsKind(tsmorphgo.KindWhileStatement) {
				kind = "while"
				condition = truncateString(node.GetText(), 30)
			}

			if kind != "" {
				totalControlFlow++
				if totalControlFlow <= 10 {
					controlFlow = append(controlFlow, struct {
						kind      string
						condition string
						line      int
						file      string
					}{
						kind:      kind,
						condition: condition,
						line:      node.GetStartLineNumber(),
						file:      extractFileName(file.GetFilePath()),
					})
				}
			}
		})
	}

	fmt.Printf("📊 控制流统计: 找到 %d 个控制流语句\n", totalControlFlow)
	if len(controlFlow) > 0 {
		fmt.Printf("控制流语句详情:\n")
		for i, cf := range controlFlow {
			fmt.Printf("  %d. %s (%s) - 行 %d - %s\n",
				i+1, cf.kind, cf.condition, cf.line, cf.file)
		}
	}

	fmt.Println("\n🎯 新API使用总结:")
	fmt.Println("1. 函数分析 → 使用 IsFunctionDeclaration() + GetNodeName()")
	fmt.Println("2. 调用分析 → 使用 IsCallExpr() + 遍历子节点")
	fmt.Println("3. 属性访问 → 使用 IsPropertyAccessExpression() + GetText()")
	fmt.Println("4. 变量分析 → 使用 IsVariableDeclaration() + GetNodeName()")
	fmt.Println("5. 类型分析 → 使用 IsKind(KindXxx) + 精确匹配")
	fmt.Println("6. 导入分析 → 使用 IsImportDeclaration() + 文本解析")
	fmt.Println("7. 控制流 → 使用 IsKind() + 条件提取")

	fmt.Println("\n✅ 专用API示例完成!")
	fmt.Println("新API让复杂的AST分析变得简单直观！")
}

// 辅助函数：提取文件名
func extractFileName(filePath string) string {
	parts := strings.Split(filePath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return filePath
}

// 辅助函数：截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
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
	} else if strings.HasPrefix(importClause, "* as") {
		// 命名空间导入
		namespace := strings.TrimSpace(importClause[4:])
		if namespace != "" {
			items = append(items, namespace)
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