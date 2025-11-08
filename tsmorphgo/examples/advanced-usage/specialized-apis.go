//go:build specialized_apis
// +build specialized_apis

package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🛠️ TSMorphGo 专用API使用示例")
	fmt.Println("=" + repeat("=", 50))

	// 初始化项目，指向一个真实的React项目目录
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
	fmt.Printf("项目包含 %d 个TypeScript文件。\n", len(sourceFiles))

	// 示例1: 函数声明处理 (FunctionDeclaration)
	fmt.Println("\n🔧 示例1: 函数声明处理")
	totalFunctions := 0
	for _, file := range sourceFiles {
		// 遍历文件中的所有节点
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// IsFunctionDeclaration 检查节点是否为函数声明
			if tsmorphgo.IsFunctionDeclaration(node) {
				// GetFunctionDeclarationNameNode 获取函数声明的名称节点
				if funcName, ok := tsmorphgo.GetFunctionDeclarationNameNode(node); ok {
					totalFunctions++
					if totalFunctions <= 5 { // 仅显示前5个以保持简洁
						fmt.Printf("函数: %s (行 %d)\n", funcName.GetText(), node.GetStartLineNumber())
						// 通过检查文本前缀来简单判断是否导出
						fmt.Printf("  - 是否导出: %v\n", strings.HasPrefix(node.GetText(), "export"))
					}
				}
			}
		})
	}
	fmt.Printf("总计发现 %d 个函数声明\n", totalFunctions)

	// 示例2: 调用表达式处理 (CallExpression)
	fmt.Println("\n⚡ 示例2: 调用表达式分析")
	totalCalls := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// IsCallExpression 检查节点是否为函数或方法调用
			if tsmorphgo.IsCallExpression(node) {
				totalCalls++
				// GetCallExpressionExpression 获取被调用的表达式部分 (例如 `foo.bar` in `foo.bar()`)
				if target, ok := tsmorphgo.GetCallExpressionExpression(node); ok {
					if totalCalls <= 10 { // 仅显示前10个
						fmt.Printf("方法调用: %s (行 %d)\n", target.GetText(), node.GetStartLineNumber())

						// IsPropertyAccessExpression 检查被调用的是否为成员访问表达式 (例如 `obj.method`)
						if tsmorphgo.IsPropertyAccessExpression(*target) {
							fmt.Printf("  - 调用类型: 成员方法调用\n")
						} else {
							fmt.Printf("  - 调用类型: 普通函数调用\n")
						}

						// AsCallExpression().Arguments.Nodes 获取调用的参数列表
						argCount := len(node.AsCallExpression().Arguments.Nodes)
						fmt.Printf("  - 参数数量: %d\n", argCount)
					}
				}
			}
		})
	}
	fmt.Printf("总计发现 %d 个方法调用\n", totalCalls)

	// 示例3: 属性访问表达式处理 (PropertyAccessExpression)
	fmt.Println("\n🔗 示例3: 属性访问表达式分析")
	propertyAccessCount := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// 通过节点的 Kind 属性直接判断类型
			if node.Kind == tsmorphgo.KindPropertyAccessExpression {
				propertyAccessCount++
				if propertyAccessCount <= 10 {
					// GetPropertyAccessName 获取属性访问的名称 (例如 `bar` in `foo.bar`)
					if name, ok := tsmorphgo.GetPropertyAccessName(node); ok {
						fmt.Printf("属性访问: %s (来自: %s)\n", name, node.GetText())
					}
				}
			}
		})
	}
	fmt.Printf("总计发现 %d 个属性访问\n", propertyAccessCount)

	// 示例4: 变量声明分析 (VariableDeclaration)
	fmt.Println("\n📦 示例4: 变量声明分析")
	variableCount := 0
	exportedVariables := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// IsVariableDeclaration 检查节点是否为变量声明
			if tsmorphgo.IsVariableDeclaration(node) {
				variableCount++
				// GetVariableName 获取变量名
				if varName, ok := tsmorphgo.GetVariableName(node); ok {
					if variableCount <= 10 {
						fmt.Printf("变量: %s (行 %d)\n", varName, node.GetStartLineNumber())
						// 简单检查是否导出
						isExported := strings.HasPrefix(node.GetParent().GetParent().GetText(), "export")
						fmt.Printf("  - 是否导出: %v\n", isExported)
						if isExported {
							exportedVariables++
						}
					}
				}
			}
		})
	}
	fmt.Printf("总计发现 %d 个变量声明，其中约 %d 个导出变量\n", variableCount, exportedVariables)

	// 示例5: 类型声明分析 (InterfaceDeclaration, TypeAliasDeclaration)
	fmt.Println("\n🏷️ 示例5: 类型声明分析")
	interfaceCount := 0
	typeAliasCount := 0
	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// 使用 Kind 判断接口声明
			if node.Kind == tsmorphgo.KindInterfaceDeclaration {
				interfaceCount++
				if interfaceCount <= 5 {
					if ifaceName, ok := tsmorphgo.GetFirstChild(node, tsmorphgo.IsIdentifier); ok {
						fmt.Printf("接口 %d: %s (行 %d)\n",
							interfaceCount, ifaceName.GetText(), node.GetStartLineNumber())
					}
				}
			} else if node.Kind == tsmorphgo.KindTypeAliasDeclaration { // 使用 Kind 判断类型别名
				typeAliasCount++
				if typeAliasCount <= 5 {
					text := strings.TrimSpace(node.GetText())
					if len(text) > 80 {
						text = text[:80] + "..."
					}
					fmt.Printf("类型别名 %d: %s (行 %d)\n", typeAliasCount, text, node.GetStartLineNumber())
				}
			}
		})
	}
	fmt.Printf("总计发现 %d 个接口声明, %d 个类型别名\n", interfaceCount, typeAliasCount)

	// --- 以下为新增的、用于补充文档覆盖范围的示例 ---

	// 示例6: 符号分析 (Symbol)
	// Symbol 是 TypeScript 编译器在语义层面理解代码的方式，比纯文本匹配更准确。
	fmt.Println("\n🧬 示例6: 符号(Symbol)分析")
	appFile := project.GetSourceFile(realProjectPath + "/src/App.tsx")
	if appFile != nil {
		appFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if tsmorphgo.IsVariableDeclaration(node) {
				if name, ok := tsmorphgo.GetVariableName(node); ok && name == "users" {
					fmt.Printf("找到 'users' 变量声明 (行 %d)\n", node.GetStartLineNumber())
					if nameNode, ok := tsmorphgo.GetFirstChild(node, tsmorphgo.IsIdentifier); ok {
						// GetSymbol 从一个节点获取其对应的符号。这可能失败，所以返回一个 error。
						symbol, err := tsmorphgo.GetSymbol(*nameNode)
						if err == nil {
							// GetName 获取符号的名称。
							fmt.Printf("  - 符号名称: %s\n", symbol.GetName())
						} else {
							fmt.Println("  - 未能获取符号")
						}
					}
					return // 只演示一次
				}
			}
		})
	}

	// 示例7: 属性访问表达式的深度分析
	// 对应 ts-morph 的 `propertyAccessExpression.getExpression()`
	fmt.Println("\n🔬 示例7: 属性访问表达式深度分析")
	if appFile != nil {
		appFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if tsmorphgo.IsPropertyAccessExpression(node) && node.GetText() == "response.data" {
				fmt.Printf("找到属性访问: %s (行 %d)\n", node.GetText(), node.GetStartLineNumber())
				// GetPropertyAccessExpression 获取被访问的对象部分 (即 `response`)
				if expr, ok := tsmorphgo.GetPropertyAccessExpression(node); ok {
					fmt.Printf("  - 被访问的对象 (Expression): %s\n", expr.GetText())
				}
				// GetPropertyAccessName 获取被访问的属性名 (即 `data`)
				if name, ok := tsmorphgo.GetPropertyAccessName(node); ok {
					fmt.Printf("  - 访问的属性名 (Name): %s\n", name)
				}
				return // 只演示一次
			}
		})
	}

	// 示例8: 导入别名分析 (ImportSpecifier)
	// 对应 ts-morph 的 `importSpecifier.getAliasNode()`
	fmt.Println("\n📛 示例8: 导入别名分析")
	appFilePath := realProjectPath + "/src/App.tsx"
	originalContent, err := os.ReadFile(appFilePath) // 读取原始文件内容
	if err != nil {
		log.Printf("无法读取 App.tsx: %v", err)
	} else {
		// 在函数结束时，无论如何都恢复文件的原始内容，确保示例不破坏项目文件
		defer os.WriteFile(appFilePath, originalContent, 0644)

		// 动态地在文件内容中添加一个带别名的导入语句
		newContent := strings.Replace(string(originalContent),
			"import _ from 'lodash';",
			"import _ from 'lodash';\nimport { type User as AppUser } from './types';", 1)

		// 使用修改后的内容创建一个（或覆盖）源文件
		aliasedFile, err := project.CreateSourceFile(appFilePath, newContent, tsmorphgo.CreateSourceFileOptions{Overwrite: true})
		if err != nil {
			log.Printf("创建带别名的文件失败: %v", err)
		} else {
			aliasedFile.ForEachDescendant(func(node tsmorphgo.Node) {
				// IsImportSpecifier 检查节点是否为导入说明符 (例如 `{ User as AppUser }` 中的 `User as AppUser`)
				if tsmorphgo.IsImportSpecifier(node) {
					// GetImportSpecifierAliasNode 获取导入项的别名节点
					if alias, ok := tsmorphgo.GetImportSpecifierAliasNode(node); ok {
						// 简单地通过遍历子节点来找到原始名称
						originalName := "unknown"
						if prop, ok := tsmorphgo.GetFirstChild(node, func(n tsmorphgo.Node) bool {
							return n.Kind == tsmorphgo.KindIdentifier && n.GetText() != alias.GetText()
						}); ok {
							originalName = prop.GetText()
						}
						fmt.Printf("找到导入别名: '%s' as '%s' (行 %d)\n", originalName, alias.GetText(), node.GetStartLineNumber())
						return // 只演示一次
					}
				}
			})
		}
	}

	// 示例9: 二元表达式分析 (BinaryExpression)
	// 对应 ts-morph 的 `binaryExpression.getLeft()`, `.getRight()`, `.getOperatorToken()`
	fmt.Println("\n⚖️ 示例9: 二元表达式分析")
	if appFile != nil {
		appFile.ForEachDescendant(func(node tsmorphgo.Node) {
			// IsBinaryExpression 检查节点是否为二元表达式 (例如 `a + b`, `c = d`)
			if tsmorphgo.IsBinaryExpression(node) {
				// GetBinaryExpressionOperatorToken 获取操作符节点
				if operator, ok := tsmorphgo.GetBinaryExpressionOperatorToken(node); ok && operator.Kind == tsmorphgo.KindEqualsToken {
					fmt.Printf("找到赋值表达式: %s (行 %d)\n", node.GetText(), node.GetStartLineNumber())

					// GetBinaryExpressionLeft 获取左侧操作数
					if left, ok := tsmorphgo.GetBinaryExpressionLeft(node); ok {
						fmt.Printf("  - 左侧操作数 (Left): %s\n", left.GetText())
					}
					// GetBinaryExpressionRight 获取右侧操作数
					if right, ok := tsmorphgo.GetBinaryExpressionRight(node); ok {
						fmt.Printf("  - 右侧操作数 (Right): %s\n", right.GetText())
					}
					fmt.Printf("  - 操作符 (Operator): %s (%s)\n", operator.GetText(), operator.GetKindName())

					return // 只演示一次
				}
			}
		})
	}

	fmt.Println("\n✅ 专用API使用示例完成!")
}

// 辅助函数，用于重复字符串
func repeat(s string, count int) string {
	return strings.Repeat(s, count)
}

// 废弃的辅助函数，因为 tsmorphgo 提供了更直接的API
/*
func getPropertyName(node tsmorphgo.Node) (string, bool) {
	if nameNode, ok := tsmorphgo.GetFirstChild(node, func(child tsmorphgo.Node) bool {
		return tsmorphgo.IsIdentifier(child)
	}); ok {
		return strings.TrimSpace(nameNode.GetText()), true
	}
	return "", false
}
*/
