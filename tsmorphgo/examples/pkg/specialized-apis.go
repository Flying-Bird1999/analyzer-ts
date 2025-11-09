//go:build specialized_apis
// +build specialized_apis

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🛠️ TSMorphGo 专用API - 正确使用姿势")
	fmt.Println("=" + repeat("=", 50))

	// =============================================================================
	// 本文件演示 TSMorphGo 专用API的正确使用方法
	// =============================================================================
	// 学习级别: 中级 → 高级
	// 预计时间: 40-55分钟
	//
	// 功能覆盖:
	// - 函数声明处理: IsFunctionDeclaration, GetFunctionDeclarationNameNode
	// - 调用表达式分析: IsCallExpression, GetCallExpressionExpression
	// - 属性访问表达式: IsPropertyAccessExpression, GetPropertyAccessName
	// - 变量声明分析: IsVariableDeclaration, GetVariableName
	// - 类型声明分析: IsInterfaceDeclaration, IsTypeAliasDeclaration
	// - 导入别名处理: IsImportSpecifier, GetImportSpecifierAliasNode
	// - 二元表达式分析: IsBinaryExpression, GetBinaryExpressionLeft/Right
	//
	// 对齐 ts-morph API:
	// - node.isFunctionDeclaration() → IsFunctionDeclaration()
	// - functionDeclaration.getName() → GetFunctionDeclarationNameNode()
	// - node.isCallExpression() → IsCallExpression()
	// - callExpression.getExpression() → GetCallExpressionExpression()
	// - node.isPropertyAccessExpression() → IsPropertyAccessExpression()
	// - propertyAccessExpression.getName() → GetPropertyAccessName()
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

	sourceFiles := project.GetSourceFiles()
	if len(sourceFiles) == 0 {
		log.Fatal("未找到任何源文件")
	}

	fmt.Printf("📊 项目统计: %d 个TypeScript文件\n", len(sourceFiles))

	// 示例1: 函数声明处理 (中级)
	// 对应 ts-morph: node.isFunctionDeclaration(), functionDeclaration.getName()
	fmt.Println("\n🔧 示例1: 函数声明处理 (中级)")
	fmt.Println("对齐 ts-morph: node.isFunctionDeclaration(), functionDeclaration.getName()")
	fmt.Println("功能: 识别和分析函数声明的关键信息")

	var functions []struct {
		name      string
		line      int
		isExported bool
		file      string
	}

	totalFunctions := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// IsFunctionDeclaration 检查节点是否为函数声明
			if tsmorphgo.IsFunctionDeclaration(node) {
				totalFunctions++
				// GetFunctionDeclarationNameNode 获取函数声明的名称节点
				if funcName, ok := tsmorphgo.GetFunctionDeclarationNameNode(node); ok {
					if totalFunctions <= 8 { // 显示前8个
						// 检查是否导出
						isExported := strings.HasPrefix(strings.TrimSpace(node.GetText()), "export")

						functions = append(functions, struct {
							name      string
							line      int
							isExported bool
							file      string
						}{
							name:      funcName.GetText(),
							line:      node.GetStartLineNumber(),
							isExported: isExported,
							file:      extractFileName(file.GetFilePath()),
						})

						fmt.Printf("  - %s", funcName.GetText())
						if isExported {
							fmt.Printf(" (导出)")
						}
						fmt.Printf(" - 行 %d, 文件: %s\n", node.GetStartLineNumber(), extractFileName(file.GetFilePath()))
					}
				}
			}
		})
	}

	fmt.Printf("✅ 总计发现 %d 个函数声明\n", totalFunctions)

	// 示例2: 调用表达式分析 (中级)
	// 对应 ts-morph: node.isCallExpression(), callExpression.getExpression()
	fmt.Println("\n⚡ 示例2: 调用表达式分析 (中级)")
	fmt.Println("对齐 ts-morph: node.isCallExpression(), callExpression.getExpression()")
	fmt.Println("功能: 分析函数和方法的调用模式")

	var calls []struct {
		target     string
		line       int
		file       string
		isMethod   bool
		argCount   int
	}

	totalCalls := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// IsCallExpression 检查节点是否为函数或方法调用
			if tsmorphgo.IsCallExpression(node) {
				totalCalls++
				// GetCallExpressionExpression 获取被调用的表达式部分
				if target, ok := tsmorphgo.GetCallExpressionExpression(node); ok {
					if totalCalls <= 10 { // 显示前10个
						targetText := strings.TrimSpace(target.GetText())

						// IsPropertyAccessExpression 检查是否为成员方法调用
						isMethod := tsmorphgo.IsPropertyAccessExpression(*target)

						// 获取参数数量
						argCount := len(node.AsCallExpression().Arguments.Nodes)

						calls = append(calls, struct {
							target     string
							line       int
							file       string
							isMethod   bool
							argCount   int
						}{
							target:   targetText,
							line:     node.GetStartLineNumber(),
							file:     extractFileName(file.GetFilePath()),
							isMethod: isMethod,
							argCount: argCount,
						})

						fmt.Printf("  - %s", targetText)
						if isMethod {
							fmt.Printf(" (方法调用)")
						} else {
							fmt.Printf(" (函数调用)")
						}
						fmt.Printf(" - 行 %d, 参数: %d\n", node.GetStartLineNumber(), argCount)
					}
				}
			}
		})
	}

	fmt.Printf("✅ 总计发现 %d 个方法调用\n", totalCalls)

	// 示例3: 属性访问表达式分析 (中级)
	// 对应 ts-morph: node.isPropertyAccessExpression(), propertyAccessExpression.getName()
	fmt.Println("\n🔗 示例3: 属性访问表达式分析 (中级)")
	fmt.Println("对齐 ts-morph: node.isPropertyAccessExpression(), propertyAccessExpression.getName()")
	fmt.Println("功能: 理解对象属性的访问模式")

	var propertyAccesses []struct {
		property  string
		expression string
		line      int
		file      string
	}

	propertyAccessCount := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// 通过节点的 Kind 属性直接判断类型
			if node.Kind == tsmorphgo.KindPropertyAccessExpression {
				propertyAccessCount++
				if propertyAccessCount <= 12 { // 显示前12个
					// GetPropertyAccessName 获取属性访问的名称
					if name, ok := tsmorphgo.GetPropertyAccessName(node); ok {
						fullText := strings.TrimSpace(node.GetText())

						propertyAccesses = append(propertyAccesses, struct {
							property  string
							expression string
							line      int
							file      string
						}{
							property:   name,
							expression: fullText,
							line:       node.GetStartLineNumber(),
							file:       extractFileName(file.GetFilePath()),
						})

						fmt.Printf("  - 属性: %s (完整表达式: %s)\n", name, truncateString(fullText, 40))
						fmt.Printf("    位置: 行 %d, 文件: %s\n", node.GetStartLineNumber(), extractFileName(file.GetFilePath()))
					}
				}
			}
		})
	}

	fmt.Printf("✅ 总计发现 %d 个属性访问\n", propertyAccessCount)

	// 示例4: 变量声明分析 (中级)
	// 对应 ts-morph: node.isVariableDeclaration(), variableDeclaration.getName()
	fmt.Println("\n📦 示例4: 变量声明分析 (中级)")
	fmt.Println("对齐 ts-morph: node.isVariableDeclaration(), variableDeclaration.getName()")
	fmt.Println("功能: 跟踪变量的声明和导出状态")

	var variables []struct {
		name      string
		line      int
		file      string
		isExported bool
	}

	variableCount := 0
	exportedVariables := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// IsVariableDeclaration 检查节点是否为变量声明
			if tsmorphgo.IsVariableDeclaration(node) {
				variableCount++
				// GetVariableName 获取变量名
				if varName, ok := tsmorphgo.GetVariableName(node); ok {
					if variableCount <= 10 { // 显示前10个
						// 检查是否导出
						parent := node.GetParent()
						isExported := false
						for parent != nil {
							parentText := strings.ToLower(strings.TrimSpace(parent.GetText()))
							if strings.HasPrefix(parentText, "export") {
								isExported = true
								exportedVariables++
								break
							}
							parent = parent.GetParent()
						}

						variables = append(variables, struct {
							name      string
							line      int
							file      string
							isExported bool
						}{
							name:      varName,
							line:      node.GetStartLineNumber(),
							file:      extractFileName(file.GetFilePath()),
							isExported: isExported,
						})

						fmt.Printf("  - %s", varName)
						if isExported {
							fmt.Printf(" (导出)")
						}
						fmt.Printf(" - 行 %d, 文件: %s\n", node.GetStartLineNumber(), extractFileName(file.GetFilePath()))
					}
				}
			}
		})
	}

	fmt.Printf("✅ 总计发现 %d 个变量声明，其中 %d 个导出变量\n", variableCount, exportedVariables)

	// 示例5: 类型声明分析 (中级)
	// 对应 ts-morph: node.isInterfaceDeclaration(), node.isTypeAliasDeclaration()
	fmt.Println("\n🏷️ 示例5: 类型声明分析 (中级)")
	fmt.Println("对齐 ts-morph: node.isInterfaceDeclaration(), node.isTypeAliasDeclaration()")
	fmt.Println("功能: 识别接口和类型别名的定义")

	var types []struct {
		kind      string
		name      string
		line      int
		file      string
		detail    string
	}

	interfaceCount := 0
	typeAliasCount := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// 使用 Kind 判断接口声明
			if node.Kind == tsmorphgo.KindInterfaceDeclaration {
				interfaceCount++
				if interfaceCount <= 6 { // 显示前6个
					if nameNode, ok := tsmorphgo.GetFirstChild(node, tsmorphgo.IsIdentifier); ok {
						// 简单统计接口成员
						propertyCount := 0
						node.ForEachDescendant(func(descendant tsmorphgo.Node) {
							if descendant.Kind == tsmorphgo.KindPropertySignature {
								propertyCount++
							}
						})

						types = append(types, struct {
							kind   string
							name   string
							line   int
							file   string
							detail string
						}{
							kind:   "接口",
							name:   nameNode.GetText(),
							line:   node.GetStartLineNumber(),
							file:   extractFileName(file.GetFilePath()),
							detail: fmt.Sprintf("%d个属性", propertyCount),
						})

						fmt.Printf("  - 接口: %s (%d个属性)\n", nameNode.GetText(), propertyCount)
						fmt.Printf("    位置: 行 %d, 文件: %s\n", node.GetStartLineNumber(), extractFileName(file.GetFilePath()))
					}
				}
			} else if node.Kind == tsmorphgo.KindTypeAliasDeclaration { // 使用 Kind 判断类型别名
				typeAliasCount++
				if typeAliasCount <= 6 { // 显示前6个
					text := strings.TrimSpace(node.GetText())
					if len(text) > 50 {
						text = text[:47] + "..."
					}

					// 检查是否是泛型类型
					isGeneric := strings.Contains(text, "<") && strings.Contains(text, ">")
					detail := "类型别名"
					if isGeneric {
						detail += " (泛型)"
					}

					if nameNode, ok := tsmorphgo.GetFirstChild(node, tsmorphgo.IsIdentifier); ok {
						types = append(types, struct {
							kind   string
							name   string
							line   int
							file   string
							detail string
						}{
							kind:   "类型别名",
							name:   nameNode.GetText(),
							line:   node.GetStartLineNumber(),
							file:   extractFileName(file.GetFilePath()),
							detail: detail,
						})

						fmt.Printf("  - 类型别名: %s (%s)\n", nameNode.GetText(), detail)
						fmt.Printf("    位置: 行 %d, 文件: %s\n", node.GetStartLineNumber(), extractFileName(file.GetFilePath()))
					}
				}
			}
		})
	}

	fmt.Printf("✅ 总计发现 %d 个接口声明, %d 个类型别名\n", interfaceCount, typeAliasCount)

	// 示例6: 导入别名分析 (高级 ⭐)
	// 对应 ts-morph: importSpecifier.getAliasNode()
	fmt.Println("\n📛 示例6: 导入别名分析 (高级 ⭐)")
	fmt.Println("对齐 ts-morph: importSpecifier.getAliasNode()")
	fmt.Println("功能: 处理复杂的模块导入和别名模式")

	var importAliases []struct {
		original  string
		alias     string
		line      int
		file      string
		context   string
	}

	aliasCount := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if aliasCount >= 8 { // 只演示前8个
				return
			}

			// IsImportSpecifier 检查节点是否为导入说明符
			if tsmorphgo.IsImportSpecifier(node) {
				// GetImportSpecifierAliasNode 获取导入项的别名节点
				if alias, ok := tsmorphgo.GetImportSpecifierAliasNode(node); ok {
					// 获取原始名称
					originalName := "unknown"
					if prop, ok := tsmorphgo.GetFirstChild(node, func(n tsmorphgo.Node) bool {
						return n.Kind == tsmorphgo.KindIdentifier && n.GetText() != alias.GetText()
					}); ok {
						originalName = prop.GetText()
					}

					// 获取导入语句的上下文
					context := ""
					grandParent := node.GetParent()
					if grandParent != nil {
						context = truncateString(strings.TrimSpace(grandParent.GetText()), 60)
					}

					importAliases = append(importAliases, struct {
						original string
						alias    string
						line     int
						file     string
						context  string
					}{
						original: originalName,
						alias:    alias.GetText(),
						line:     node.GetStartLineNumber(),
						file:     extractFileName(file.GetFilePath()),
						context:  context,
					})

					aliasCount++
					fmt.Printf("  - 导入别名: '%s' as '%s'\n", originalName, alias.GetText())
					fmt.Printf("    位置: 行 %d, 文件: %s\n", node.GetStartLineNumber(), extractFileName(file.GetFilePath()))
					fmt.Printf("    上下文: %s\n", context)
				}
			}
		})
	}

	if aliasCount == 0 {
		fmt.Println("  - 未找到导入别名")
	} else {
		fmt.Printf("✅ 在项目中找到 %d 个导入别名\n", aliasCount)
	}

	// 示例7: 二元表达式分析 (高级 ⭐)
	// 对应 ts-morph: binaryExpression.getLeft(), binaryExpression.getRight(), binaryExpression.getOperatorToken()
	fmt.Println("\n⚖️ 示例7: 二元表达式分析 (高级 ⭐)")
	fmt.Println("对齐 ts-morph: binaryExpression.getLeft(), binaryExpression.getRight(), binaryExpression.getOperatorToken()")
	fmt.Println("功能: 理解赋值、比较和逻辑运算的表达式结构")

	var binaryExpressions []struct {
		left      string
		right     string
		operator  string
		line      int
		file      string
		fullExpr  string
	}

	foundCount := 0
	for _, file := range sourceFiles {
		if foundCount >= 8 { // 只演示前8个
			break
		}

		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if foundCount >= 8 {
				return
			}

			// IsBinaryExpression 检查节点是否为二元表达式
			if tsmorphgo.IsBinaryExpression(node) {
				// GetBinaryExpressionOperatorToken 获取操作符节点
				if operator, ok := tsmorphgo.GetBinaryExpressionOperatorToken(node); ok {
					operatorText := strings.TrimSpace(operator.GetText())

					// 重点关注赋值操作符和逻辑操作符
					if operatorText == "=" || operatorText == "+=" || operatorText == "-=" ||
						operatorText == "&&" || operatorText == "||" || operatorText == "==" ||
						operatorText == "!=" || operatorText == "<" || operatorText == ">" {

						// 获取左右操作数
						leftText := ""
						if left, ok := tsmorphgo.GetBinaryExpressionLeft(node); ok {
							leftText = truncateString(strings.TrimSpace(left.GetText()), 25)
						}

						rightText := ""
						if right, ok := tsmorphgo.GetBinaryExpressionRight(node); ok {
							rightText = truncateString(strings.TrimSpace(right.GetText()), 25)
						}

						fullExpr := truncateString(strings.TrimSpace(node.GetText()), 40)

						binaryExpressions = append(binaryExpressions, struct {
							left     string
							right    string
							operator string
							line     int
							file     string
							fullExpr string
						}{
							left:     leftText,
							right:    rightText,
							operator: operatorText,
							line:     node.GetStartLineNumber(),
							file:     extractFileName(file.GetFilePath()),
							fullExpr: fullExpr,
						})

						foundCount++
						fmt.Printf("  - 表达式: %s\n", fullExpr)
						fmt.Printf("    左操作数: %s\n", leftText)
						fmt.Printf("    操作符: %s\n", operatorText)
						fmt.Printf("    右操作数: %s\n", rightText)
						fmt.Printf("    位置: 行 %d, 文件: %s\n", node.GetStartLineNumber(), extractFileName(file.GetFilePath()))
					}
				}
			}
		})
	}

	if foundCount == 0 {
		fmt.Println("  - 未找到二元表达式")
	} else {
		fmt.Printf("✅ 分析了 %d 个二元表达式\n", foundCount)
	}

	// 示例8: 符号分析应用 (高级 ⭐)
	// 对应 ts-morph: node.getSymbol(), symbol.getName()
	fmt.Println("\n🧬 示例8: 符号分析应用 (高级 ⭐)")
	fmt.Println("对齐 ts-morph: node.getSymbol(), symbol.getName()")
	fmt.Println("功能: 语义级别的代码分析，理解标识符的真实含义")

	// 选择App.tsx进行符号分析演示
	appFile := project.GetSourceFile(realProjectPath + "/src/App.tsx")
	if appFile != nil {
		symbolAnalysisCount := 0
		appFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if symbolAnalysisCount >= 5 { // 只演示前5个
				return
			}

			// 重点关注变量声明的符号
			if tsmorphgo.IsVariableDeclaration(node) {
				if name, ok := tsmorphgo.GetVariableName(node); ok && len(name) > 2 {
					// 获取标识符节点
					if nameNode, ok := tsmorphgo.GetFirstChild(node, tsmorphgo.IsIdentifier); ok {
						// GetSymbol 获取节点的符号信息
						symbol, err := tsmorphgo.GetSymbol(*nameNode)
						if err == nil && symbol != nil {
							symbolAnalysisCount++
							fmt.Printf("  - 变量: '%s'\n", name)
							fmt.Printf("    符号名称: %s\n", symbol.GetName())
							fmt.Printf("    位置: 行 %d\n", node.GetStartLineNumber())

							// 检查符号是否有类型信息
							if symbol.HasType() {
								fmt.Printf("    类型信息: 有\n")
							} else {
								fmt.Printf("    类型信息: 无\n")
							}
						}
					}
				}
			}
		})

		if symbolAnalysisCount == 0 {
			fmt.Println("  - 未找到可分析的符号")
		} else {
			fmt.Printf("✅ 成功分析了 %d 个符号\n", symbolAnalysisCount)
		}
	} else {
		fmt.Println("  - 未找到 App.tsx 文件")
	}

	fmt.Println("\n🎯 专用API使用姿势总结:")
	fmt.Println("1. 函数声明 → IsFunctionDeclaration() + GetFunctionDeclarationNameNode()")
	fmt.Println("2. 调用分析 → IsCallExpression() + GetCallExpressionExpression()")
	fmt.Println("3. 属性访问 → IsPropertyAccessExpression() + GetPropertyAccessName()")
	fmt.Println("4. 变量分析 → IsVariableDeclaration() + GetVariableName()")
	fmt.Println("5. 类型声明 → Kind == KindInterfaceDeclaration/KindTypeAliasDeclaration")
	fmt.Println("6. 导入别名 → IsImportSpecifier() + GetImportSpecifierAliasNode()")
	fmt.Println("7. 二元表达式 → IsBinaryExpression() + GetBinaryExpressionLeft/Right/OperatorToken()")
	fmt.Println("8. 符号分析 → GetSymbol() + symbol.GetName() + symbol.HasType()")

	fmt.Println("\n✅ 专用API示例完成!")
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

// 辅助函数：提取文件名
func extractFileName(filePath string) string {
	parts := strings.Split(filePath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return filePath
}