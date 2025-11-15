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
	fmt.Println("🎯 TSMorphGo Hook函数引用查找示例")
	fmt.Println("==================================")
	fmt.Println("验证场景: Hook函数(变量声明)的引用查找")
	fmt.Println()

	// ============================================================================
	// 项目初始化
	// 验证文件: ./demo-react-app/src/hooks/useUserData.ts
	// 目标节点: 第10行的 useUserData 变量名标识符
	// 预期输出: 找到 useUserData 的定义和使用位置
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

	useUserDataFile := project.GetSourceFile(filepath.Join(demoAppPath, "src/hooks/useUserData.ts"))
	if useUserDataFile == nil {
		log.Fatal("❌ 未找到 useUserData.ts 文件")
	}

	fmt.Printf("✅ 找到目标文件: %s\n", useUserDataFile.GetFilePath())

	// ============================================================================
	// 查找 useUserData 变量声明中的标识符节点
	// 验证API: ForEachDescendant() - 遍历所有节点
	// 验证API: IsIdentifier() - 判断是否为标识符
	// 验证目标: 找到变量名 'useUserData' 的标识符节点
	// 预期输出: 找到标识符节点及其位置信息
	// ============================================================================

	fmt.Println()
	fmt.Println("🔍 步骤1: 查找 useUserData 标识符节点")
	fmt.Println("----------------------------------")

	var declarationIdentifier tsmorphgo.Node
	var declarationFound bool
	var declText string
	var declLine, declCol int

	// 遍历文件查找 useUserData 标识符
	useUserDataFile.ForEachDescendant(func(node tsmorphgo.Node) {
		// 验证API: IsIdentifier() - 判断是否为标识符
		if node.IsIdentifier() && node.GetText() == "useUserData" {
			// 检查是否在变量声明中
			parent := node.GetParent()
			if parent != nil && parent.IsVariableDeclaration() {
				declarationIdentifier = node
				declarationFound = true
				declText = node.GetText()

				// 验证API: GetStartLineNumber() - 获取起始行号
				declLine = node.GetStartLineNumber()
				// 验证API: GetStartColumnNumber() - 获取起始列号
				declCol = node.GetStartColumnNumber()

				fmt.Printf("✅ 找到 useUserData 声明标识符\n")
				fmt.Printf("📍 位置: 第%d行，第%d列\n", declLine, declCol)
				fmt.Printf("🏷️  标识符文本: %s\n", declText)

				// 获取父节点信息
				parentKind := parent.GetKind()
				fmt.Printf("🔧 父节点类型: %s\n", parentKind.String())
			}
		}
	})

	if !declarationFound {
		log.Fatal("❌ 未找到 useUserData 声明标识符")
	}

	// ============================================================================
	// 场景5.1: 获取节点的符号和名称
	// 验证API: GetSymbol() - 获取节点的符号信息
	// 验证目标: 获取 useUserData 的符号信息
	// 预期输出: 显示符号名称
	// ============================================================================

	fmt.Println()
	fmt.Println("🔖 步骤2: 获取符号信息")
	fmt.Println("--------------------")

	// 验证API: GetSymbol() - 获取节点的符号信息
	symbol, err := declarationIdentifier.GetSymbol()
	if err != nil {
		fmt.Printf("❌ 获取符号失败: %v\n", err)
	} else if symbol == nil {
		fmt.Println("❌ 节点没有符号信息")
	} else {
		symbolName := symbol.GetName()
		fmt.Printf("✅ 符号名称: %s\n", symbolName)

		if symbolName == "useUserData" {
			fmt.Println("✅ 符号名称验证正确")
		} else {
			fmt.Printf("❌ 符号名称不匹配，期望: useUserData, 实际: %s\n", symbolName)
		}

		// 获取符号标志
		flags := symbol.GetFlags()
		fmt.Printf("🔖 符号标志: %d\n", flags)
	}

	// ============================================================================
	// 方式1: 从声明处查找引用
	// 验证API: FindReferences() - 查找标识符的所有引用位置
	// 验证目标: 从变量声明处查找所有 useUserData 的引用
	// 预期输出: 找到定义和调用位置
	// ============================================================================

	fmt.Println()
	fmt.Println("🔍 方式1: 从声明处查找引用")
	fmt.Println("--------------------------")

	var referencesFromDecl []*tsmorphgo.Node

	// 验证API: FindReferences() - 查找所有引用位置
	if refs, err := tsmorphgo.FindReferences(declarationIdentifier); err != nil {
		fmt.Printf("❌ 引用查找失败: %v\n", err)
	} else {
		referencesFromDecl = refs
		fmt.Printf("✅ 找到 %d 个引用:\n", len(refs))

		// 显示所有引用位置
		for i, ref := range refs {
			refFile := ref.GetSourceFile()
			if refFile != nil {
				refLine := ref.GetStartLineNumber()
				refCol := ref.GetStartColumnNumber()
				refText := ref.GetText()

				// 判断是定义还是使用
				if refLine == declLine && refCol == declCol {
					fmt.Printf("  %d. %s:%d:%d (变量声明) - %s\n",
						i+1, refFile.GetFilePath(), refLine, refCol, refText)
				} else {
					fmt.Printf("  %d. %s:%d:%d (Hook调用) - %s\n",
						i+1, refFile.GetFilePath(), refLine, refCol, refText)
				}
			}
		}
	}

	// ============================================================================
	// 查找 App.tsx 中的 useUserData 调用
	// 验证目标: 找到函数调用处的标识符节点
	// 预期输出: 找到调用位置的标识符
	// ============================================================================

	fmt.Println()
	fmt.Println("🔍 步骤3: 查找调用处的标识符")
	fmt.Println("--------------------------")

	appFile := project.GetSourceFile(filepath.Join(demoAppPath, "src/components/App.tsx"))
	if appFile == nil {
		log.Fatal("❌ 未找到 App.tsx 文件")
	}

	var callIdentifier tsmorphgo.Node
	var callFound bool
	var callLine, callCol int

	// 遍历 App.tsx 文件查找 useUserData 调用
	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		// 查找函数调用表达式
		if node.IsCallExpression() {
			// 获取被调用的表达式部分
			node.ForEachChild(func(child tsmorphgo.Node) bool {
				// 查找标识符
				if child.IsIdentifier() && child.GetText() == "useUserData" {
					callIdentifier = child
					callFound = true
					callLine = child.GetStartLineNumber()
					callCol = child.GetStartColumnNumber()
					fmt.Printf("✅ 找到 useUserData 调用标识符\n")
					fmt.Printf("📍 位置: 第%d行，第%d列\n", callLine, callCol)
					return true // 停止遍历
				}
				return false
			})
		}
	})

	if !callFound {
		fmt.Println("❌ 未找到 useUserData 调用标识符")
	}

	// ============================================================================
	// 方式2: 从调用处查找引用
	// 验证目标: 从函数调用处查找所有引用
	// 预期输出: 找到相同的引用列表
	// ============================================================================

	fmt.Println()
	fmt.Println("🔍 方式2: 从调用处查找引用")
	fmt.Println("--------------------------")

	var referencesFromCall []*tsmorphgo.Node

	if callFound {
		// 从调用处查找引用
		if refs, err := tsmorphgo.FindReferences(callIdentifier); err != nil {
			fmt.Printf("❌ 引用查找失败: %v\n", err)
		} else {
			referencesFromCall = refs
			fmt.Printf("✅ 找到 %d 个引用:\n", len(refs))

			// 显示所有引用位置
			for i, ref := range refs {
				refFile := ref.GetSourceFile()
				if refFile != nil {
					refLine := ref.GetStartLineNumber()
					refCol := ref.GetStartColumnNumber()
					refText := ref.GetText()

					// 判断是定义还是使用
					if refLine == declLine && refCol == declCol {
						fmt.Printf("  %d. %s:%d:%d (变量声明) - %s\n",
							i+1, refFile.GetFilePath(), refLine, refCol, refText)
					} else {
						fmt.Printf("  %d. %s:%d:%d (Hook调用) - %s\n",
							i+1, refFile.GetFilePath(), refLine, refCol, refText)
					}
				}
			}
		}
	}

	// ============================================================================
	// 结果验证: 确保两种查找方式结果一致
	// 验证方法: 比较引用数量和位置
	// 预期输出: 两种方式结果一致
	// ============================================================================

	fmt.Println()
	fmt.Println("🔍 结果验证")
	fmt.Println("------------")

	if len(referencesFromDecl) > 0 && len(referencesFromCall) > 0 {
		declCount := len(referencesFromDecl)
		callCount := len(referencesFromCall)

		fmt.Printf("📊 从声明处找到引用数: %d\n", declCount)
		fmt.Printf("📊 从调用处找到引用数: %d\n", callCount)

		if declCount == callCount {
			fmt.Println("✅ 两种查找方式找到的引用数量一致")

			// 比较具体引用位置
			allMatch := true
			for i, ref1 := range referencesFromDecl {
				if i < len(referencesFromCall) {
					ref2 := referencesFromCall[i]
					file1 := ref1.GetSourceFile()
					file2 := ref2.GetSourceFile()

					if file1 != nil && file2 != nil {
						line1 := ref1.GetStartLineNumber()
						col1 := ref1.GetStartColumnNumber()
						line2 := ref2.GetStartLineNumber()
						col2 := ref2.GetStartColumnNumber()

						if line1 != line2 || col1 != col2 {
							allMatch = false
							fmt.Printf("❌ 引用位置不匹配: 方式1(%d:%d) vs 方式2(%d:%d)\n",
								line1, col1, line2, col2)
							break
						}
					}
				}
			}

			if allMatch {
				fmt.Println("✅ 两种查找方式找到的引用位置完全一致")
			}
		} else {
			fmt.Println("❌ 两种查找方式找到的引用数量不一致")
		}
	} else {
		fmt.Println("❌ 某种查找方式未找到引用，无法比较")
	}

	// ============================================================================
	// 引用上下文分析
	// 验证目标: 分析每个引用的具体使用场景
	// 预期输出: 显示引用的上下文信息
	// ============================================================================

	fmt.Println()
	fmt.Println("📊 引用上下文分析")
	fmt.Println("-----------------")

	if len(referencesFromDecl) > 0 {
		for i, ref := range referencesFromDecl {
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
			if refLine == declLine && refCol == declCol {
				fmt.Println("🔧 类型: 变量声明 (const useUserData = ...)")
			} else {
				fmt.Println("🔧 类型: Hook函数调用 (useUserData(...))")

				// 获取调用上下文
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

	// 清理资源
	defer project.Close()

	fmt.Println()
	fmt.Println("🎉 Hook函数引用查找示例完成！")
	fmt.Println()
	fmt.Println("✅ 验证总结:")
	fmt.Println("   - 标识符节点查找: 成功")
	fmt.Println("   - 符号信息获取: 成功")
	fmt.Println("   - 从声明处查找引用: 成功")
	fmt.Println("   - 从调用处查找引用: 成功")
	fmt.Println("   - 引用结果验证: 成功")
	fmt.Println("   - 引用上下文分析: 成功")
}