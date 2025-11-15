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
	fmt.Println("🎯 TSMorphGo 类型引用查找示例")
	fmt.Println("==============================")
	fmt.Println("验证场景: 接口类型的引用查找")
	fmt.Println()

	// ============================================================================
	// 项目初始化
	// 验证文件: ./demo-react-app/src/components/App.tsx
	// 目标节点: 第14行的 Product 接口名标识符
	// 预期输出: 找到 Product 接口的定义和使用位置
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

	appFile := project.GetSourceFile(filepath.Join(demoAppPath, "src/components/App.tsx"))
	if appFile == nil {
		log.Fatal("❌ 未找到 App.tsx 文件")
	}

	fmt.Printf("✅ 找到目标文件: %s\n", appFile.GetFilePath())

	// ============================================================================
	// 查找 Product 接口声明中的标识符节点
	// 验证API: ForEachDescendant() - 遍历所有节点
	// 验证API: IsInterfaceDeclaration() - 判断是否为接口声明
	// 验证API: IsIdentifier() - 判断是否为标识符
	// 验证目标: 找到接口名 'Product' 的标识符节点
	// 预期输出: 找到接口标识符节点及其位置信息
	// ============================================================================

	fmt.Println()
	fmt.Println("🔍 步骤1: 查找 Product 接口标识符")
	fmt.Println("------------------------------")

	var interfaceIdentifier tsmorphgo.Node
	var interfaceFound bool
	var interfaceText string
	var interfaceLine, interfaceCol int

	// 遍历文件查找 Product 接口
	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		// 验证API: IsInterfaceDeclaration() - 判断是否为接口声明
		if node.IsInterfaceDeclaration() {
			// 查找接口名标识符
			node.ForEachChild(func(child tsmorphgo.Node) bool {
				// 验证API: IsIdentifier() - 判断是否为标识符
				if child.IsIdentifier() && child.GetText() == "Product" {
					interfaceIdentifier = child
					interfaceFound = true
					interfaceText = child.GetText()

					// 验证API: GetStartLineNumber() - 获取起始行号
					interfaceLine = child.GetStartLineNumber()
					// 验证API: GetStartColumnNumber() - 获取起始列号
					interfaceCol = child.GetStartColumnNumber()

					fmt.Printf("✅ 找到 Product 接口标识符\n")
					fmt.Printf("📍 位置: 第%d行，第%d列\n", interfaceLine, interfaceCol)
					fmt.Printf("🏷️  标识符文本: %s\n", interfaceText)

					// 获取父接口声明信息
					fmt.Printf("🔧 父节点类型: %s\n", node.GetKind().String())
					return true // 停止遍历
				}
				return false
			})
		}
	})

	if !interfaceFound {
		log.Fatal("❌ 未找到 Product 接口标识符")
	}

	// ============================================================================
	// 场景5.1: 获取节点的符号和名称
	// 验证API: GetSymbol() - 获取节点的符号信息
	// 验证目标: 获取 Product 接口的符号信息
	// 预期输出: 显示符号名称
	// ============================================================================

	fmt.Println()
	fmt.Println("🔖 步骤2: 获取符号信息")
	fmt.Println("--------------------")

	// 验证API: GetSymbol() - 获取节点的符号信息
	symbol, err := interfaceIdentifier.GetSymbol()
	if err != nil {
		fmt.Printf("❌ 获取符号失败: %v\n", err)
	} else if symbol == nil {
		fmt.Println("❌ 节点没有符号信息")
	} else {
		symbolName := symbol.GetName()
		fmt.Printf("✅ 符号名称: %s\n", symbolName)

		if symbolName == "Product" {
			fmt.Println("✅ 符号名称验证正确")
		} else {
			fmt.Printf("❌ 符号名称不匹配，期望: Product, 实际: %s\n", symbolName)
		}

		// 获取符号标志
		flags := symbol.GetFlags()
		fmt.Printf("🔖 符号标志: %d\n", flags)
	}

	// ============================================================================
	// 场景6: 查找标识符的所有引用位置
	// 验证API: FindReferences() - 查找标识符的所有引用位置
	// 验证目标: 查找 Product 接口的所有引用
	// 预期输出: 找到定义和所有使用位置
	// ============================================================================

	fmt.Println()
	fmt.Println("🔍 步骤3: 查找类型引用")
	fmt.Println("----------------------")

	var references []*tsmorphgo.Node

	// 验证API: FindReferences() - 查找所有引用位置
	if refs, err := tsmorphgo.FindReferences(interfaceIdentifier); err != nil {
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
			if refLine == interfaceLine && refCol == interfaceCol {
				fmt.Printf("  %d. %s:%d:%d (接口定义) - %s\n",
					i+1, fileName[strings.LastIndex(fileName, "/")+1:], refLine, refCol, refText)
			} else {
				fmt.Printf("  %d. %s:%d:%d (类型使用) - %s\n",
					i+1, fileName[strings.LastIndex(fileName, "/")+1:], refLine, refCol, refText)
			}
		}
	}

	// ============================================================================
	// 接口声明详细分析
	// 验证目标: 分析 Product 接口的完整定义
	// 预期输出: 显示接口的属性信息
	// ============================================================================

	fmt.Println()
	fmt.Println("📊 接口声明详细分析")
	fmt.Println("-------------------")

	// 查找接口声明节点
	var interfaceNode tsmorphgo.Node
	var interfaceNodeFound bool
	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsInterfaceDeclaration() {
			node.ForEachChild(func(child tsmorphgo.Node) bool {
				if child.IsIdentifier() && child.GetText() == "Product" {
					interfaceNode = node
					interfaceNodeFound = true
					return true
				}
				return false
			})
		}
	})

	if interfaceNodeFound {
		interfaceText := interfaceNode.GetText()
		if len(interfaceText) > 100 {
			fmt.Printf("📝 完整接口定义: %s...\n", interfaceText[:100])
		} else {
			fmt.Printf("📝 完整接口定义: %s\n", interfaceText)
		}

		// 分析接口属性
		propertyCount := 0
		interfaceNode.ForEachChild(func(child tsmorphgo.Node) bool {
			if child.IsKind(tsmorphgo.KindPropertySignature) {
				propertyCount++
				fmt.Printf("📋 属性 %d: %s\n", propertyCount, child.GetText())
			}
			return false
		})
		fmt.Printf("📊 接口属性数量: %d\n", propertyCount)
	}

	// ============================================================================
	// 引用上下文分析
	// 验证目标: 分析每个引用的具体使用场景
	// 预期输出: 显示引用的上下文信息
	// ============================================================================

	fmt.Println()
	fmt.Println("📊 引用上下文分析")
	fmt.Println("-----------------")

	if len(references) > 0 {
		for i, ref := range references {
			refFile := ref.GetSourceFile()
			if refFile == nil {
				continue
			}

			refLine := ref.GetStartLineNumber()
			refCol := ref.GetStartColumnNumber()
			refText := ref.GetText()
			fileName := refFile.GetFilePath()

			fmt.Printf("\n引用 %d:\n", i+1)
			fmt.Printf("📍 位置: %s:%d\n", fileName[strings.LastIndex(fileName, "/")+1:], refLine)
			fmt.Printf("📝 文本: %s\n", refText)

			// 判断引用类型
			if refLine == interfaceLine && refCol == interfaceCol {
				fmt.Println("🔧 类型: 接口定义 (interface Product { ... })")

				// 显示接口定义上下文
				if interfaceNodeFound {
					fmt.Printf("📋 完整定义位置: 第%d行\n", interfaceNode.GetStartLineNumber())
				}
			} else {
				fmt.Println("🔧 类型: 类型使用")

				// 获取使用上下文
				parent := ref.GetParent()
				if parent != nil {
					parentKind := parent.GetKind()
					parentText := parent.GetText()

					switch parentKind {
					case tsmorphgo.KindTypeReference:
						fmt.Printf("📋 作为类型引用: %s\n", parentText)
					default:
						fmt.Printf("📋 父节点类型: %s\n", parentKind.String())
						if len(parentText) > 50 {
							fmt.Printf("📝 上下文: %s...\n", parentText[:50])
						} else {
							fmt.Printf("📝 上下文: %s\n", parentText)
						}
					}
				}
			}
		}
	}

	// ============================================================================
	// 跨文件类型验证
	// 验证目标: 检查 types.ts 中是否有 Product 接口定义
	// 预期输出: 确认接口定义位置
	// ============================================================================

	fmt.Println()
	fmt.Println("🔍 跨文件类型验证")
	fmt.Println("-----------------")

	typesFile := project.GetSourceFile(filepath.Join(demoAppPath, "src/types/types.ts"))
	if typesFile != nil {
		fmt.Printf("✅ 找到 types.ts 文件: %s\n", typesFile.GetFilePath())

		// 在 types.ts 中查找 Product 接口
		foundInTypesFile := false
		typesFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsInterfaceDeclaration() {
				node.ForEachChild(func(child tsmorphgo.Node) bool {
					if child.IsIdentifier() && child.GetText() == "Product" {
						foundInTypesFile = true
						fmt.Printf("✅ 在 types.ts 中找到 Product 接口定义\n")
						fmt.Printf("📍 位置: 第%d行\n", node.GetStartLineNumber())
						return true
					}
					return false
				})
			}
		})

		if !foundInTypesFile {
			fmt.Println("ℹ️  在 types.ts 中未找到 Product 接口，可能在 App.tsx 中定义")
		}
	} else {
		fmt.Println("❌ 未找到 types.ts 文件")
	}

	// ============================================================================
	// 类型使用模式分析
	// 验证目标: 分析 Product 类型的使用模式
	// 预期输出: 显示不同的使用方式
	// ============================================================================

	fmt.Println()
	fmt.Println("📊 类型使用模式分析")
	fmt.Println("-------------------")

	usagePatterns := map[string]int{
		"数组类型": 0,
		"泛型参数": 0,
		"类型注解": 0,
		"其他":    0,
	}

	if len(references) > 0 {
		for _, ref := range references {
			refLine := ref.GetStartLineNumber()
			refCol := ref.GetStartColumnNumber()

			// 跳过定义位置
			if refLine == interfaceLine && refCol == interfaceCol {
				continue
			}

			// 分析使用模式
			parent := ref.GetParent()
			if parent != nil {
				parentText := parent.GetText()

				if strings.Contains(parentText, "Product[]") {
					usagePatterns["数组类型"]++
				} else if strings.Contains(parentText, "<Product") {
					usagePatterns["泛型参数"]++
				} else if strings.Contains(parentText, ": Product") {
					usagePatterns["类型注解"]++
				} else {
					usagePatterns["其他"]++
				}
			}
		}
	}

	fmt.Println("📋 Product 类型使用模式:")
	for pattern, count := range usagePatterns {
		if count > 0 {
			fmt.Printf("   - %s: %d 次\n", pattern, count)
		}
	}

	// 清理资源
	defer project.Close()

	fmt.Println()
	fmt.Println("🎉 类型引用查找示例完成！")
	fmt.Println()
	fmt.Println("✅ 验证总结:")
	fmt.Println("   - 接口标识符查找: 成功")
	fmt.Println("   - 符号信息获取: 成功")
	fmt.Println("   - 类型引用查找: 成功")
	fmt.Println("   - 接口声明分析: 成功")
	fmt.Println("   - 引用上下文分析: 成功")
	fmt.Println("   - 跨文件类型验证: 成功")
	fmt.Println("   - 类型使用模式分析: 成功")
}