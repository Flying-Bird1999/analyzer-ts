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
	fmt.Println("🛠️ TSMorphGo 专用API使用示例")
	fmt.Println("=" + repeat("=", 50))

	// 使用真实的demo-react-app项目进行演示
	realProjectPath := "/Users/bird/Desktop/alalyzer/analyzer-ts/tsmorphgo/examples/demo-react-app"

	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:         realProjectPath,
		TargetExtensions: []string{".ts", ".tsx"},
		IgnorePatterns:   []string{"node_modules", "dist", ".git", "build"},
		UseTsConfig:      true,
	})
	defer project.Close()

	// 示例1: 函数声明处理
	fmt.Println("\n🔧 示例1: 函数声明处理")

	// 获取项目中的所有源文件
	sourceFiles := project.GetSourceFiles()
	if len(sourceFiles) == 0 {
		log.Fatal("未找到任何源文件")
	}

	fmt.Printf("项目包含 %d 个TypeScript文件:\n", len(sourceFiles))

	// 使用新的API分析函数声明
	totalFunctions := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsFunction() {
				if funcName, ok := node.GetName(); ok {
					fmt.Printf("函数: %s (行 %d)\n", funcName, node.GetStartLineNumber())
					totalFunctions++

					// 检查函数属性
					fmt.Printf("  - 是否导出: %v\n", node.IsExported())
					fmt.Printf("  - 是否异步: %v\n", node.IsAsyncFunction())
					fmt.Printf("  - 返回类型: %s\n", node.GetType())

					// 简单的参数统计
					paramCount := 0
					node.ForEachDescendant(func(descendant tsmorphgo.Node) {
						if descendant.Kind == tsmorphgo.KindParameter {
							paramCount++
						}
					})
					fmt.Printf("  - 参数数量: %d\n", paramCount)
				}
			}
		})
	}

	fmt.Printf("总计发现 %d 个函数声明\n", totalFunctions)

	// 示例2: 调用表达式处理
	fmt.Println("\n⚡ 示例2: 调用表达式分析")

	// 使用新的API分析方法调用
	totalCalls := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsCallExpression() {
				// 使用新的常量，替换魔法数字
				if target, ok := tsmorphgo.GetCallExpressionExpression(node); ok {
					totalCalls++

					// 只显示前10个调用以避免输出过多
					if totalCalls <= 10 {
						fmt.Printf("方法调用: %s (行 %d)\n", target.GetText(), node.GetStartLineNumber())

						if node.IsMemberAccess() {
							fmt.Printf("  - 调用类型: 成员方法调用\n")
						} else {
							fmt.Printf("  - 调用类型: 普通函数调用\n")
						}

						// 获取参数（使用新的常量）
						argCount := len(node.AsCallExpression().Arguments.Nodes)
						fmt.Printf("  - 参数数量: %d\n", argCount)
					}
				}
			}
		})
	}

	fmt.Printf("总计发现 %d 个方法调用\n", totalCalls)

	// 示例3: 属性访问表达式处理
	fmt.Println("\n🔗 示例3: 属性访问表达式分析")

	// 使用新的API分析属性访问
	propertyAccessCount := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// 使用新的API和常量，替换魔法数字
			if node.Kind == tsmorphgo.KindPropertyAccessExpression {
				text := strings.TrimSpace(node.GetText())
				// 只处理简单的属性访问，排除方法调用
				if !strings.Contains(text, "()") {
					propertyAccessCount++

					// 只显示前10个属性访问以避免输出过多
					if propertyAccessCount <= 10 {
						fmt.Printf("属性访问: %s (行 %d)\n", text, node.GetStartLineNumber())
					}
				}
			}
		})
	}

	fmt.Printf("总计发现 %d 个属性访问\n", propertyAccessCount)

	// 示例4: 变量声明分析 - 使用新的API
	fmt.Println("\n📦 示例4: 变量声明分析")

	variableCount := 0
	exportedVariables := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsVariable() {
				variableCount++
				if variableCount <= 10 { // 只显示前10个
					if varName, ok := node.GetName(); ok {
						fmt.Printf("变量: %s (行 %d)\n", varName, node.GetStartLineNumber())
						fmt.Printf("  - 类型: %s\n", node.GetType())
						fmt.Printf("  - 是否导出: %v\n", node.IsExported())
						fmt.Printf("  - 声明方式: ", "")
						if node.IsConst() {
							fmt.Printf("const\n")
						} else if node.IsLet() {
							fmt.Printf("let\n")
						} else {
							fmt.Printf("var\n")
						}
					}
				}
				if node.IsExported() {
					exportedVariables++
				}
			}
		})
	}

	fmt.Printf("总计发现 %d 个变量声明，其中 %d 个导出变量\n", variableCount, exportedVariables)

	// 示例5: 类型声明分析 - 使用新的常量
	fmt.Println("\n🏷️ 示例5: 类型声明分析")

	interfaceCount := 0
	typeAliasCount := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// 使用新的常量，替换魔法数字
			if node.Kind == tsmorphgo.KindInterfaceDeclaration {
				interfaceCount++
				if interfaceCount <= 5 { // 只显示前5个
					if ifaceName, ok := node.GetName(); ok {
						fmt.Printf("接口 %d: %s (行 %d, 是否导出: %v)\n",
							interfaceCount, ifaceName, node.GetStartLineNumber(), node.IsExported())
					}
				}
			} else if node.Kind == tsmorphgo.KindTypeAliasDeclaration {
				typeAliasCount++
				if typeAliasCount <= 5 { // 只显示前5个
					text := strings.TrimSpace(node.GetText())
					if len(text) > 50 {
						text = text[:50] + "..."
					}
					fmt.Printf("类型别名 %d: %s (行 %d)\n", typeAliasCount, text, node.GetStartLineNumber())
				}
			}
		})
	}

	fmt.Printf("总计发现 %d 个接口声明, %d 个类型别名\n", interfaceCount, typeAliasCount)

	// 示例6: 条件表达式处理
	fmt.Println("\n🤔 示例6: 条件表达式分析")

	conditionalCount := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.Kind == tsmorphgo.KindConditionalExpression {
				conditionalCount++
				if conditionalCount <= 5 { // 只显示前5个
					text := strings.TrimSpace(node.GetText())
					if len(text) > 80 {
						text = text[:80] + "..."
					}
					fmt.Printf("条件表达式 %d: %s (行 %d)\n", conditionalCount, text, node.GetStartLineNumber())

					// TODO: Conditional expression APIs not available yet
					fmt.Printf("  - 条件表达式结构: 三元运算符\n")
				}
			}
		})
	}

	fmt.Printf("总计发现 %d 个条件表达式\n", conditionalCount)

	fmt.Println("\n✅ 专用API使用示例完成!")
}

// 辅助函数
func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

func getPropertyName(node tsmorphgo.Node) (string, bool) {
	if nameNode, ok := tsmorphgo.GetFirstChild(node, func(child tsmorphgo.Node) bool {
		return tsmorphgo.IsIdentifier(child)
	}); ok {
		return strings.TrimSpace(nameNode.GetText()), true
	}
	return "", false
}
