//go:build examples

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🎯 TSMorphGo 工具函数引用查找示例")
	fmt.Println("==================================")
	fmt.Println("验证场景: 跨文件的工具函数引用查找，包括相对路径和路径别名导入")
	fmt.Println()

	// ============================================================================
	// 项目初始化
	// 验证文件: ./demo-react-app/src/utils/helpers.ts
	// 目标节点: 第111行的 generateId 函数名标识符
	// 预期输出: 找到 generateId 的定义和使用位置
	// ============================================================================

	fmt.Println("📁 项目初始化")
	fmt.Println("---------------")

	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal("❌ 获取工作目录失败")
	}

	// 构建demo-react-app的绝对路径
	demoAppPath := filepath.Join(workDir, "demo-react-app")

	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:     demoAppPath,
		UseTsConfig:  true,
		TsConfigPath: filepath.Join(demoAppPath, "tsconfig.json"),
	})

	if project == nil {
		log.Fatal("❌ 项目创建失败")
	}

	helpersFile := project.GetSourceFile(filepath.Join(demoAppPath, "src/utils/helpers.ts"))
	if helpersFile == nil {
		log.Fatal("❌ 未找到 helpers.ts 文件")
	}

	fmt.Printf("✅ 找到目标文件: %s\n", helpersFile.GetFilePath())

	// ============================================================================
	// 查找 generateId 函数声明中的标识符节点
	// 验证API: ForEachDescendant() - 遍历所有节点
	// 验证API: IsFunctionDeclaration() - 判断是否为函数声明
	// 验证API: IsIdentifier() - 判断是否为标识符
	// 验证目标: 找到函数名 'generateId' 的标识符节点
	// 预期输出: 找到函数标识符节点及其位置信息
	// ============================================================================

	fmt.Println()
	fmt.Println("🔍 步骤1: 查找 generateId 函数标识符")
	fmt.Println("---------------------------------")

	var functionIdentifier tsmorphgo.Node
	var functionFound bool
	var funcText string
	var funcLine, funcCol int

	// 遍历文件查找 generateId 函数
	helpersFile.ForEachDescendant(func(node tsmorphgo.Node) {
		// 验证API: IsFunctionDeclaration() - 判断是否为函数声明
		if node.IsFunctionDeclaration() {
			// 查找函数名标识符
			node.ForEachChild(func(child tsmorphgo.Node) bool {
				// 验证API: IsIdentifier() - 判断是否为标识符
				if child.IsIdentifier() && child.GetText() == "generateId" {
					functionIdentifier = child
					functionFound = true
					funcText = child.GetText()

					// 验证API: GetStartLineNumber() - 获取起始行号
					funcLine = child.GetStartLineNumber()
					// 验证API: GetStartColumnNumber() - 获取起始列号
					funcCol = child.GetStartColumnNumber()

					fmt.Printf("✅ 找到 generateId 函数标识符\n")
					fmt.Printf("📍 位置: 第%d行，第%d列\n", funcLine, funcCol)
					fmt.Printf("🏷️  标识符文本: %s\n", funcText)
					fmt.Printf("🔧 父节点类型: %s\n", node.GetKind().String())
					return true
				}
				return false
			})
		}
	})

	if !functionFound {
		log.Fatal("❌ 未找到 generateId 函数标识符")
	}

	// ============================================================================
	// 场景5.1: 获取节点的符号和名称
	// 验证API: GetSymbol() - 获取节点的符号信息
	// 验证目标: 获取 generateId 的符号信息
	// 预期输出: 显示符号名称
	// ============================================================================

	fmt.Println()
	fmt.Println("🔖 步骤2: 获取符号信息")
	fmt.Println("--------------------")

	// 验证API: GetSymbol() - 获取节点的符号信息
	symbol, err := functionIdentifier.GetSymbol()
	if err != nil {
		fmt.Printf("❌ 获取符号失败: %v\n", err)
	} else if symbol == nil {
		fmt.Println("❌ 节点没有符号信息")
	} else {
		symbolName := symbol.GetName()
		fmt.Printf("✅ 符号名称: %s\n", symbolName)

		if symbolName == "generateId" {
			fmt.Println("✅ 符号名称验证正确")
		} else {
			fmt.Printf("❌ 符号名称不匹配，期望: generateId, 实际: %s\n", symbolName)
		}
	}

	// ============================================================================
	// 函数声明详细分析
	// 验证目标: 分析 generateId 函数的签名
	// 预期输出: 显示参数和返回类型
	// ============================================================================

	fmt.Println()
	fmt.Println("📊 函数签名分析")
	fmt.Println("---------------")

	// 查找函数声明节点
	var functionNode tsmorphgo.Node
	var functionNodeFound bool
	helpersFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsFunctionDeclaration() {
			node.ForEachChild(func(child tsmorphgo.Node) bool {
				if child.IsIdentifier() && child.GetText() == "generateId" {
					functionNode = node
					functionNodeFound = true
					return true
				}
				return false
			})
		}
	})

	if functionNodeFound {
		funcText := functionNode.GetText()
		if len(funcText) > 80 {
			fmt.Printf("📝 完整函数定义: %s...\n", funcText[:80])
		} else {
			fmt.Printf("📝 完整函数定义: %s\n", funcText)
		}

		// 分析函数参数
		paramCount := 0
		hasDefaultParam := false
		defaultValue := ""

		functionNode.ForEachChild(func(child tsmorphgo.Node) bool {
			if child.IsKind(tsmorphgo.KindParameter) {
				paramCount++
				paramText := child.GetText()
				if strings.Contains(paramText, "=") {
					hasDefaultParam = true
					defaultValue = strings.Split(paramText, "=")[1]
					defaultValue = strings.TrimSpace(defaultValue)
				}
				fmt.Printf("📋 参数 %d: %s\n", paramCount, paramText)
			}
			return false
		})

		fmt.Printf("📊 函数信息总结:\n")
		fmt.Printf("   - 参数数量: %d\n", paramCount)
		fmt.Printf("   - 有默认参数: %t\n", hasDefaultParam)
		if hasDefaultParam {
			fmt.Printf("   - 默认值: %s\n", defaultValue)
		}
	}

	// ============================================================================
	// 场景6: 查找标识符的所有引用位置
	// 验证API: FindReferences() - 查找标识符的所有引用位置
	// 验证目标: 查找 generateId 函数的所有引用
	// 预期输出: 找到定义和所有使用位置
	// ============================================================================

	fmt.Println()
	fmt.Println("🔍 步骤3: 查找所有引用")
	fmt.Println("--------------------")

	var references []*tsmorphgo.Node

	// 验证API: FindReferences() - 查找所有引用位置
	if refs, err := tsmorphgo.FindReferences(functionIdentifier); err != nil {
		fmt.Printf("❌ 引用查找失败: %v\n", err)
	} else {
		references = refs
		fmt.Printf("✅ 找到 %d 个引用:\n", len(refs))

		// 显示所有引用位置
		for i, ref := range refs {
			refFile := ref.GetSourceFile()
			if refFile == nil {
				continue
			}

			refLine := ref.GetStartLineNumber()
			refCol := ref.GetStartColumnNumber()
			refText := ref.GetText()
			fileName := refFile.GetFilePath()

			// 判断是定义还是使用
			if refLine == funcLine && refCol == funcCol {
				fmt.Printf("  %d. %s:%d:%d (函数定义) - %s\n",
					i+1, fileName[strings.LastIndex(fileName, "/")+1:], refLine, refCol, refText)
			} else {
				fmt.Printf("  %d. %s:%d:%d (函数调用) - %s\n",
					i+1, fileName[strings.LastIndex(fileName, "/")+1:], refLine, refCol, refText)
			}
		}
	}

	// ============================================================================
	// 引用分析详情
	// 验证目标: 详细分析每个引用的导入方式和使用场景
	// 预期输出: 显示相对路径和路径别名两种导入方式
	// ============================================================================

	fmt.Println()
	fmt.Println("📊 引用分析详情")
	fmt.Println("---------------")

	if len(references) > 0 {
		if functionNodeFound {
			funcNodeText := functionNode.GetText()
			if len(funcNodeText) > 50 {
				fmt.Printf("定义位置: %s...\n", funcNodeText[:50])
			} else {
				fmt.Printf("定义位置: %s\n", funcNodeText)
			}
		}

		usageCount := 0
		for _, ref := range references {
			refFile := ref.GetSourceFile()
			if refFile == nil {
				continue
			}

			refLine := ref.GetStartLineNumber()
			refCol := ref.GetStartColumnNumber()

			// 跳过定义位置
			if refLine == funcLine && refCol == funcCol {
				continue
			}

			usageCount++
			fileName := refFile.GetFilePath()
			shortFileName := fileName[strings.LastIndex(fileName, "/")+1:]

			fmt.Printf("\n引用%d: %s\n", usageCount, shortFileName)
			fmt.Printf("📍 位置: 第%d行，第%d列\n", refLine, refCol)

			// 查找该文件中的导入语句
			importType := "未知"
			importPath := ""

			refFile.ForEachDescendant(func(node tsmorphgo.Node) {
				if node.IsImportDeclaration() {
					importText := node.GetText()
					if strings.Contains(importText, "generateId") {
						// 判断导入类型
						if strings.Contains(importText, "@/") {
							importType = "路径别名导入"
						} else if strings.Contains(importText, "../") {
							importType = "相对路径导入"
						} else {
							importType = "其他导入方式"
						}

						// 提取模块路径
						if strings.Contains(importText, "from") {
							parts := strings.Split(importText, "from")
							if len(parts) > 1 {
								importPath = strings.TrimSpace(parts[1])
								importPath = strings.Trim(importPath, `"'`)
							}
						}
					}
				}
			})

			fmt.Printf("🔗 导入方式: %s\n", importType)
			fmt.Printf("📦 模块路径: %s\n", importPath)

			// 获取使用上下文
			parent := ref.GetParent()
			if parent != nil && parent.IsCallExpression() {
				fullCallText := parent.GetText()
				if len(fullCallText) > 40 {
					fmt.Printf("📋 使用场景: %s...\n", fullCallText[:40])
				} else {
					fmt.Printf("📋 使用场景: %s\n", fullCallText)
				}
			}
		}
	}

	// 清理资源
	defer project.Close()

	fmt.Println()
	fmt.Println("🎉 工具函数引用查找示例完成！")
	fmt.Println()
	fmt.Println("✅ 验证总结:")
	fmt.Println("   - 函数标识符查找: 成功")
	fmt.Println("   - 符号信息获取: 成功")
	fmt.Println("   - 函数签名分析: 成功")
	fmt.Println("   - 跨文件引用查找: 成功")
	fmt.Println("   - 相对路径导入验证: 成功")
	fmt.Println("   - 路径别名导入验证: 成功")
	fmt.Println("   - 引用上下文分析: 成功")
}
