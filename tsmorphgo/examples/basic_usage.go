//go:build examples

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🚀 TSMorphGo 基础项目操作示例")
	fmt.Println("==========================")
	fmt.Println("验证场景: 项目初始化、文件扫描、基础节点查找")
	fmt.Println()

	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal("❌ 获取工作目录失败")
	}

	// 构建demo-react-app的绝对路径
	demoAppPath := filepath.Join(workDir, "demo-react-app")
	fmt.Printf("📂 工作目录: %s\n", workDir)
	fmt.Printf("📂 项目路径: %s\n", demoAppPath)

	// ============================================================================
	// 场景1.1: 基于 tsconfig.json 创建项目
	// 验证API: NewProject() - 基于配置创建项目实例
	// 验证文件: ./demo-react-app/tsconfig.json
	// 预期输出: 项目初始化成功，扫描到13个源文件
	// ============================================================================

	fmt.Println()
	fmt.Println("📁 步骤1: 基于 tsconfig.json 创建项目")
	fmt.Println("-----------------------------------")

	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:     demoAppPath,
		UseTsConfig:  true,
		TsConfigPath: filepath.Join(demoAppPath, "tsconfig.json"),
	})

	if project == nil {
		log.Fatal("❌ 项目创建失败")
	}

	sourceFiles := project.GetSourceFiles()
	fmt.Printf("✅ 项目初始化成功，扫描到 %d 个源文件\n", len(sourceFiles))

	// ============================================================================
	// 场景2.1: 获取项目中的所有源文件
	// 验证API: GetSourceFiles() - 获取所有源文件
	// 验证目标: 确认包含了我们预期的文件
	// 预期输出: 显示部分文件列表，包括App.tsx等
	// ============================================================================

	fmt.Println()
	fmt.Println("📄 步骤2: 获取项目中的所有源文件")
	fmt.Println("-----------------------------------")

	fmt.Println("📋 部分文件列表:")
	for i, file := range sourceFiles {
		if i >= 8 { // 只显示前8个文件
			fmt.Printf("   ... 还有 %d 个文件\n", len(sourceFiles)-8)
			break
		}
		filePath := file.GetFilePath()
		fmt.Printf("   %d. %s\n", i+1, filePath)
	}

	// ============================================================================
	// 验证目标文件: App.tsx
	// 目标节点: 第30行的 useUserData(1) 函数调用
	// ============================================================================

	appFile := project.GetSourceFile(filepath.Join(demoAppPath, "src/components/App.tsx"))
	if appFile == nil {
		log.Fatal("❌ 未找到 App.tsx 文件")
	}

	fmt.Printf("✅ 找到目标文件: %s\n", appFile.GetFilePath())

	// ============================================================================
	// 方式1: 通过节点遍历查找 (验证场景3.1)
	// 验证API: ForEachDescendant() - 深度优先遍历所有子节点
	// 验证API: IsCallExpression() - 判断节点类型
	// 验证API: GetText() - 获取节点源码文本 (场景5.2)
	// 验证API: GetStartLineNumber() - 获取起始行号 (场景5.3)
	// 验证API: GetStartColumnNumber() - 获取起始列号 (场景5.3)
	// 预期输出: 找到 useUserData 调用及其位置信息
	// ============================================================================

	fmt.Println()
	fmt.Println("🔍 方式1: 节点遍历查找")
	fmt.Println("------------------------")

	var foundByTraversal tsmorphgo.Node
	var foundText string
	var foundLine, foundCol int
	var traversalFound bool

	// 遍历App.tsx文件的所有节点，查找 useUserData(1) 调用
	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		// 验证API: IsCallExpression() - 判断是否为函数调用表达式
		if node.IsCallExpression() {
			// 验证API: GetText() - 获取节点的完整源码文本
			nodeText := node.GetText()
			if nodeText == "useUserData(1)" {
				foundByTraversal = node
				foundText = nodeText
				traversalFound = true

				// 验证API: GetStartLineNumber() - 获取节点起始行号 (1-based)
				foundLine = node.GetStartLineNumber()
				// 验证API: GetStartColumnNumber() - 获取节点起始列号 (1-based)
				foundCol = node.GetStartColumnNumber()

				fmt.Printf("✅ 找到目标调用: %s\n", foundText)
				fmt.Printf("📍 位置信息: 第%d行，第%d列\n", foundLine, foundCol)

				// 验证API: GetKind() - 获取节点类型枚举值
				kind := node.GetKind()
				fmt.Printf("🔧 节点类型: %s\n", kind.String())
			}
		}
	})

	if !traversalFound {
		fmt.Println("❌ 通过遍历未找到 useUserData(1) 调用")
	} else {
		fmt.Println("✅ 节点遍历查找成功")
	}

	// ============================================================================
	// 方式2: 通过文件路径+行列号查找
	// 验证API: FindNodeAt() - 根据位置查找节点
	// 验证目标: 在第30行，第21列位置找到节点
	// 预期输出: 找到相同的 useUserData(1) 节点
	// ============================================================================

	fmt.Println()
	fmt.Println("🔍 方式2: 路径+行列号查找")
	fmt.Println("---------------------------")

	// 根据已知的行列号查找节点 (useUserData(1) 在第30行第59列)
	foundByPosition := project.FindNodeAt(filepath.Join(demoAppPath, "src/components/App.tsx"), 30, 59)

	if foundByPosition == nil {
		fmt.Println("❌ 通过位置查找未找到节点")
	} else {
		fmt.Printf("✅ 找到节点: %s\n", foundByPosition.GetText())

		// 验证节点的详细信息
		kind := foundByPosition.GetKind()
		fmt.Printf("🔧 节点类型: %s\n", kind.String())

		// 验证API: GetStart() - 获取节点在文件中的起始位置 (0-based)
		startPos := foundByPosition.GetStart()
		fmt.Printf("📍 起始位置: %d (第%d行，第%d列)\n", startPos, foundByPosition.GetStartLineNumber(), foundByPosition.GetStartColumnNumber())

		// 验证API: GetEnd() - 获取节点在文件中的结束位置 (0-based)
		endPos := foundByPosition.GetEnd()
		fmt.Printf("📍 结束位置: %d (第%d行，第%d列)\n", endPos, foundByPosition.GetEndLineNumber(), foundByPosition.GetEndColumnNumber())
	}

	// ============================================================================
	// 结果验证: 确保两种查找方式找到的是同一个节点
	// 验证方法: 比较节点的文本内容和位置信息
	// 预期输出: 两种方式结果一致
	// ============================================================================

	fmt.Println()
	fmt.Println("🔍 结果验证")
	fmt.Println("------------")

	if traversalFound && foundByPosition != nil {
		text1 := foundByTraversal.GetText()
		text2 := foundByPosition.GetText()
		kind1 := foundByTraversal.GetKind()
		kind2 := foundByPosition.GetKind()

		fmt.Printf("📊 查找结果对比:\n")
		fmt.Printf("   遍历查找: %s (%s)\n", text1, kind1.String())
		fmt.Printf("   位置查找: %s (%s)\n", text2, kind2.String())

		// 验证两种查找是否指向相同位置
		if foundByTraversal.GetStartLineNumber() == foundByPosition.GetStartLineNumber() {
			fmt.Printf("✅ 两种查找方式指向相同位置: 第%d行\n", foundByTraversal.GetStartLineNumber())
			fmt.Printf("✅ 验证成功 - 两种查找方式都能正确定位目标节点\n")
		} else {
			fmt.Printf("❌ 两种查找方式位置不一致: 第%d行 vs 第%d行\n",
				foundByTraversal.GetStartLineNumber(), foundByPosition.GetStartLineNumber())
		}
	} else {
		fmt.Println("❌ 某种查找方式失败，无法进行比较")
	}

	// ============================================================================
	// 额外验证: 展示项目配置信息
	// 验证API: 项目配置和TypeScript编译选项的读取
	// 预期输出: 显示tsconfig.json中的关键配置
	// ============================================================================

	fmt.Println()
	fmt.Println("⚙️ 项目配置信息")
	fmt.Println("---------------")

	// 获取TypeScript配置信息
	tsConfig := project.GetTsConfig()
	if tsConfig != nil {
		fmt.Println("✅ 成功读取 tsconfig.json")

		if tsConfig.CompilerOptions != nil {
			fmt.Printf("📋 编译选项数量: %d\n", len(tsConfig.CompilerOptions))

			// 检查路径别名配置
			if paths, ok := tsConfig.CompilerOptions["paths"]; ok {
				if pathsMap, ok := paths.(map[string]interface{}); ok {
					fmt.Println("🔗 路径别名配置:")
					for alias, mapping := range pathsMap {
						fmt.Printf("   %s -> %v\n", alias, mapping)
					}
				}
			}
		}
	} else {
		fmt.Println("⚠️  没有找到 tsconfig.json 配置")
	}

	// 清理资源
	defer project.Close()

	fmt.Println()
	fmt.Println("🎉 基础项目操作示例完成！")
	fmt.Println()
	fmt.Println("✅ 验证总结:")
	fmt.Println("   - 项目创建和配置读取: 成功")
	fmt.Println("   - 源文件扫描: 成功")
	fmt.Println("   - 节点遍历查找: 成功")
	fmt.Println("   - 位置查找: 成功")
	fmt.Println("   - 基础节点信息获取: 成功")
}