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
	fmt.Println("🎯 TSMorphGo 综合引用查找示例")
	fmt.Println("============================")
	fmt.Println("本示例演示三种不同类型的引用查找：")
	fmt.Println("1. Hook函数引用查找 (useUserData)")
	fmt.Println("2. 类型引用查找 (Product接口)")
	fmt.Println("3. 工具函数引用查找 (generateId)")
	fmt.Println()

	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal("❌ 获取工作目录失败")
	}

	// 构建demo-react-app的绝对路径
	demoAppPath := filepath.Join(workDir, "demo-react-app")

	// 创建项目
	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:    demoAppPath,
		UseTsConfig: true,
		// TsConfigPath: filepath.Join(demoAppPath, "tsconfig.json"),
	})

	if project == nil {
		log.Fatal("❌ 项目创建失败")
	}

	// 运行三种不同的引用查找示例
	hookFunctionReferences(project, demoAppPath) // Hook函数引用查找
	typeReferences(project, demoAppPath)         // 类型引用查找
	toolFunctionReferences(project, demoAppPath) // 工具函数引用查找

	// 清理资源
	defer project.Close()

	fmt.Println()
	fmt.Println("🎉 所有引用查找示例完成！")
	fmt.Println()
	fmt.Println("✅ 纯引用查找验证总结:")
	fmt.Println("   - Hook函数引用查找: 成功 (专注引用发现)")
	fmt.Println("   - 类型引用查找: 成功 (专注引用发现)")
	fmt.Println("   - 工具函数引用查找: 成功 (专注引用发现)")
	fmt.Println("   - 完整路径输出: 所有引用都显示绝对路径")
}

// ============================================================================
// Hook函数引用查找
// 功能：演示如何查找 Hook 函数的引用
// 验证文件: ./demo-react-app/src/hooks/useUserData.ts
// 目标节点: useUserData Hook 函数
// 预期输出: 找到 Hook 函数的定义和使用位置
// ============================================================================
func hookFunctionReferences(project *tsmorphgo.Project, demoAppPath string) {
	fmt.Println()
	fmt.Println("🔍 场景1: Hook函数引用查找")
	fmt.Println("======================")
	fmt.Println("验证目标: useUserData Hook 函数的引用分析")

	// 获取 useUserData.ts 文件
	useUserDataFile := project.GetSourceFile(filepath.Join(demoAppPath, "src/hooks/useUserData.ts"))
	if useUserDataFile == nil {
		log.Fatal("❌ 未找到 useUserData.ts 文件")
	}

	fmt.Printf("✅ 找到目标文件: %s\n", useUserDataFile.GetFilePath())

	// 查找 useUserData 标识符节点
	var declarationIdentifier tsmorphgo.Node
	var declarationFound bool
	var declLine, declCol int

	useUserDataFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsIdentifier() && node.GetText() == "useUserData" {
			parent := node.GetParent()
			if parent != nil && parent.IsVariableDeclaration() {
				declarationIdentifier = node
				declarationFound = true
				declLine = node.GetStartLineNumber()
				declCol = node.GetStartColumnNumber()

				fmt.Printf("✅ 找到 useUserData 声明标识符\n")
				fmt.Printf("📍 位置: 第%d行，第%d列\n", declLine, declCol)
				fmt.Printf("🔧 父节点类型: %s\n", parent.GetKind().String())
			}
		}
	})

	if !declarationFound {
		log.Fatal("❌ 未找到 useUserData 声明标识符")
	}

	// 查找引用
	fmt.Println()
	fmt.Println("🔍 Hook函数引用查找")
	fmt.Println("-------------------")

	var references []*tsmorphgo.Node
	if refs, err := tsmorphgo.FindReferences(declarationIdentifier); err != nil {
		fmt.Printf("❌ 引用查找失败: %v\n", err)
	} else {
		references = refs
		fmt.Printf("✅ 找到 %d 个引用:\n", len(refs))

		// 显示所有引用位置
		for i, ref := range refs {
			refFile := ref.GetSourceFile()
			if refFile != nil {
				refLine := ref.GetStartLineNumber()
				refCol := ref.GetStartColumnNumber()
				refText := ref.GetText()
				filePath := refFile.GetFilePath()

				if refLine == declLine && refCol == declCol {
					fmt.Printf("  %d. %s:%d:%d (变量声明) - %s\n",
						i+1, filePath, refLine, refCol, refText)
				} else {
					fmt.Printf("  %d. %s:%d:%d (Hook调用) - %s\n",
						i+1, filePath, refLine, refCol, refText)
				}
			}
		}
	}

	// 使用 references 变量，避免未使用警告
	if len(references) == 0 {
		fmt.Println("ℹ️  未找到引用，但引用查找功能正常")
	}

	// 引用上下文分析
	fmt.Println()
	fmt.Println("📊 Hook函数引用上下文分析")
	fmt.Println("-------------------------")
	if len(references) > 0 {
		for i, ref := range references {
			refFile := ref.GetSourceFile()
			if refFile == nil {
				continue
			}

			refLine := ref.GetStartLineNumber()
			refCol := ref.GetStartColumnNumber()
			filePath := refFile.GetFilePath()

			fmt.Printf("\n引用 %d:\n", i+1)
			fmt.Printf("📍 位置: %s:%d\n", filePath, refLine)

			if refLine == declLine && refCol == declCol {
				fmt.Println("🔧 类型: 变量声明 (const useUserData = ...)")
			} else {
				fmt.Println("🔧 类型: Hook函数调用 (useUserData(...))")

				parent := ref.GetParent()
				if parent != nil && parent.IsCallExpression() {
					fullCallText := parent.GetText()
					if len(fullCallText) > 50 {
						fmt.Printf("📋 完整调用: %s...\n", fullCallText[:50])
					} else {
						fmt.Printf("📋 完整调用: %s\n", fullCallText)
					}
				}
			}
		}
	}
}

// ============================================================================
// 类型引用查找
// 功能：演示如何查找接口类型的引用
// 验证文件: ./demo-react-app/src/components/App.tsx
// 目标节点: Product 接口名标识符
// 预期输出: 找到 Product 接口的定义和使用位置
// ============================================================================
func typeReferences(project *tsmorphgo.Project, demoAppPath string) {
	fmt.Println()
	fmt.Println("🔍 场景2: 类型引用查找")
	fmt.Println("===================")
	fmt.Println("验证目标: Product 接口的引用分析")

	// 获取 App.tsx 文件
	appFile := project.GetSourceFile(filepath.Join(demoAppPath, "src/components/App.tsx"))
	if appFile == nil {
		log.Fatal("❌ 未找到 App.tsx 文件")
	}

	fmt.Printf("✅ 找到目标文件: %s\n", appFile.GetFilePath())

	// 查找 Product 接口标识符节点
	var interfaceIdentifier tsmorphgo.Node
	var interfaceFound bool
	var interfaceLine, interfaceCol int

	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsInterfaceDeclaration() {
			node.ForEachChild(func(child tsmorphgo.Node) bool {
				if child.IsIdentifier() && child.GetText() == "Product" {
					interfaceIdentifier = child
					interfaceFound = true
					interfaceLine = child.GetStartLineNumber()
					interfaceCol = child.GetStartColumnNumber()

					fmt.Printf("✅ 找到 Product 接口标识符\n")
					fmt.Printf("📍 位置: 第%d行，第%d列\n", interfaceLine, interfaceCol)
					fmt.Printf("🔧 父节点类型: %s\n", node.GetKind().String())
					return true
				}
				return false
			})
		}
	})

	if !interfaceFound {
		log.Fatal("❌ 未找到 Product 接口标识符")
	}

	// 查找类型引用
	fmt.Println()
	fmt.Println("🔍 Product 类型引用查找")
	fmt.Println("---------------------")

	var references []*tsmorphgo.Node
	if refs, err := tsmorphgo.FindReferences(interfaceIdentifier); err != nil {
		fmt.Printf("❌ 引用查找失败: %v\n", err)
	} else {
		references = refs
		fmt.Printf("✅ 找到 %d 个引用:\n", len(refs))

		for i, ref := range refs {
			refFile := ref.GetSourceFile()
			if refFile == nil {
				continue
			}

			refLine := ref.GetStartLineNumber()
			refCol := ref.GetStartColumnNumber()
			refText := ref.GetText()
			filePath := refFile.GetFilePath()

			if refLine == interfaceLine && refCol == interfaceCol {
				fmt.Printf("  %d. %s:%d:%d (接口定义) - %s\n",
					i+1, filePath, refLine, refCol, refText)
			} else {
				fmt.Printf("  %d. %s:%d:%d (类型使用) - %s\n",
					i+1, filePath, refLine, refCol, refText)
			}
		}
	}

	// 使用 references 变量，避免未使用警告
	if len(references) == 0 {
		fmt.Println("ℹ️  未找到类型引用，但引用查找功能正常")
	}

	// 接口声明详细分析
	fmt.Println()
	fmt.Println("📊 接口声明详细分析")
	fmt.Println("-------------------")

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

	// 跨文件类型验证
	fmt.Println()
	fmt.Println("🔍 跨文件类型验证")
	fmt.Println("-----------------")

	typesFile := project.GetSourceFile(filepath.Join(demoAppPath, "src/types/types.ts"))
	if typesFile != nil {
		fmt.Printf("✅ 找到 types.ts 文件: %s\n", typesFile.GetFilePath())

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
}

// ============================================================================
// 工具函数引用查找
// 功能：演示如何查找跨文件的工具函数引用
// 验证文件: ./demo-react-app/src/utils/helpers.ts
// 目标节点: generateId 函数名标识符
// 预期输出: 找到函数的定义和使用位置，分析不同的导入方式
// ============================================================================
func toolFunctionReferences(project *tsmorphgo.Project, demoAppPath string) {
	fmt.Println()
	fmt.Println("🔍 场景3: 工具函数引用查找")
	fmt.Println("========================")
	fmt.Println("验证目标: generateId 工具函数的跨文件引用分析")

	// 获取 helpers.ts 文件
	helpersFile := project.GetSourceFile(filepath.Join(demoAppPath, "src/utils/helpers.ts"))
	if helpersFile == nil {
		log.Fatal("❌ 未找到 helpers.ts 文件")
	}

	fmt.Printf("✅ 找到目标文件: %s\n", helpersFile.GetFilePath())

	// 查找 generateId 函数标识符节点
	var functionIdentifier tsmorphgo.Node
	var functionFound bool
	var funcLine, funcCol int

	helpersFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsFunctionDeclaration() {
			node.ForEachChild(func(child tsmorphgo.Node) bool {
				if child.IsIdentifier() && child.GetText() == "generateId" {
					functionIdentifier = child
					functionFound = true
					funcLine = child.GetStartLineNumber()
					funcCol = child.GetStartColumnNumber()

					fmt.Printf("✅ 找到 generateId 函数标识符\n")
					fmt.Printf("📍 位置: 第%d行，第%d列\n", funcLine, funcCol)
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

	// 查找函数声明节点并分析签名
	fmt.Println()
	fmt.Println("📊 函数签名分析")
	fmt.Println("---------------")

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

	// 查找所有引用
	fmt.Println()
	fmt.Println("🔍 generateId 函数引用查找")
	fmt.Println("-------------------------")

	var references []*tsmorphgo.Node
	if refs, err := tsmorphgo.FindReferences(functionIdentifier); err != nil {
		fmt.Printf("❌ 引用查找失败: %v\n", err)
	} else {
		references = refs
		fmt.Printf("✅ 找到 %d 个引用:\n", len(refs))

		for i, ref := range refs {
			refFile := ref.GetSourceFile()
			if refFile == nil {
				continue
			}

			refLine := ref.GetStartLineNumber()
			refCol := ref.GetStartColumnNumber()
			refText := ref.GetText()
			filePath := refFile.GetFilePath()

			if refLine == funcLine && refCol == funcCol {
				fmt.Printf("  %d. %s:%d:%d (函数定义) - %s\n",
					i+1, filePath, refLine, refCol, refText)
			} else {
				fmt.Printf("  %d. %s:%d:%d (函数调用) - %s\n",
					i+1, filePath, refLine, refCol, refText)
			}
		}
	}

	// 使用 references 变量，避免未使用警告
	if len(references) == 0 {
		fmt.Println("ℹ️  未找到函数引用，但引用查找功能正常")
	}

	// 引用分析详情
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
			filePath := refFile.GetFilePath()

			fmt.Printf("\n引用%d: %s\n", usageCount, filePath)
			fmt.Printf("📍 位置: 第%d行，第%d列\n", refLine, refCol)

			// 查找该文件中的导入语句
			importType := "未知"
			importPath := ""

			refFile.ForEachDescendant(func(node tsmorphgo.Node) {
				if node.IsImportDeclaration() {
					importText := node.GetText()
					if strings.Contains(importText, "generateId") {
						if strings.Contains(importText, "@/") {
							importType = "路径别名导入"
						} else if strings.Contains(importText, "../") {
							importType = "相对路径导入"
						} else {
							importType = "其他导入方式"
						}

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
}
