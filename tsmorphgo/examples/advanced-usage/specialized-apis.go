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

	// 分析所有文件中的函数声明
	totalFunctions := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if tsmorphgo.IsFunctionDeclaration(node) {
				if nameNode, ok := tsmorphgo.GetFunctionDeclarationNameNode(node); ok {
					funcName := strings.TrimSpace(nameNode.GetText())
					fmt.Printf("函数: %s (行 %d)\n", funcName, node.GetStartLineNumber())
					totalFunctions++

					// 获取参数数量
					paramCount := 0
					node.ForEachDescendant(func(descendant tsmorphgo.Node) {
						if descendant.Kind == 218 { // Parameter
							paramCount++
						}
					})
					fmt.Printf("  - 参数数量: %d\n", paramCount)

					// 检查返回类型
					text := strings.TrimSpace(node.GetText())
					if strings.Contains(text, ": Promise<") {
						fmt.Printf("  - 异步函数\n")
					}
				}
			}
		})
	}

	fmt.Printf("总计发现 %d 个函数声明\n", totalFunctions)

	// 示例2: 调用表达式处理
	fmt.Println("\n⚡ 示例2: 调用表达式分析")

	// 分析所有文件中的方法调用
	totalCalls := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if tsmorphgo.IsCallExpression(node) {
				// 获取被调用的表达式
				if expr, ok := tsmorphgo.GetCallExpressionExpression(node); ok {
					callText := strings.TrimSpace(expr.GetText())
					totalCalls++

					// 只显示前10个调用以避免输出过多
					if totalCalls <= 10 {
						fmt.Printf("方法调用: %s (行 %d)\n", callText, node.GetStartLineNumber())

						// TODO: IsMemberExpression API not available yet, showing basic call analysis
						fmt.Printf("  - 调用类型: 方法调用\n")

						// 获取参数
						argCount := 0
						node.ForEachDescendant(func(descendant tsmorphgo.Node) {
							if descendant.Kind == 215 { // Argument
								argCount++
							}
						})
						fmt.Printf("  - 参数数量: %d\n", argCount)
					}
				}
			}
		})
	}

	fmt.Printf("总计发现 %d 个方法调用\n", totalCalls)

	// 示例3: 属性访问表达式处理
	fmt.Println("\n🔗 示例3: 属性访问表达式分析")

	// 分析所有文件中的属性访问
	propertyAccessCount := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// TODO: IsMemberExpression API not available yet
			// Using node.Kind == 193 as a workaround for MemberExpression
			if node.Kind == 193 { // MemberExpression
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

	// 示例4: 变量声明处理
	fmt.Println("\n📦 示例4: 变量声明分析")

	variableCount := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if tsmorphgo.IsVariableDeclaration(node) {
				variableCount++
				if variableCount <= 10 { // 只显示前10个
					fmt.Printf("变量声明 (行 %d)\n", node.GetStartLineNumber())
				}
			}
		})
	}

	fmt.Printf("总计发现 %d 个变量声明\n", variableCount)

	// 示例5: 类型声明处理
	fmt.Println("\n🏷️ 示例5: 类型声明分析")

	interfaceCount := 0
	typeAliasCount := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.Kind == 257 { // InterfaceDeclaration
				interfaceCount++
				if interfaceCount <= 5 { // 只显示前5个
					text := strings.TrimSpace(node.GetText())
					if len(text) > 50 {
						text = text[:50] + "..."
					}
					fmt.Printf("接口声明 %d: %s (行 %d)\n", interfaceCount, text, node.GetStartLineNumber())
				}
			} else if node.Kind == 258 { // TypeAliasDeclaration
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
			if node.Kind == 268 { // ConditionalExpression
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
