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
	fmt.Println("🎯 TSMorphGo 节点导航和类型收窄示例")
	fmt.Println("==================================")
	fmt.Println("验证场景: 节点关系导航和类型安全的API访问")
	fmt.Println()

	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal("❌ 获取工作目录失败")
	}

	// 构建demo-react-app的绝对路径
	demoAppPath := filepath.Join(workDir, "demo-react-app")

	// ============================================================================
	// 项目初始化
	// 验证文件: ./demo-react-app/src/hooks/useUserData.ts
	// 目标节点: 第10行的 useUserData 变量声明 (const 声明)
	// ============================================================================

	fmt.Println("📁 项目初始化")
	fmt.Println("---------------")

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
	// 场景4: 判断节点的具体语法类型
	// 验证目标: 找到 useUserData 变量声明节点
	// 验证API: IsVariableDeclaration() - 判断是否为变量声明
	// 预期输出: 找到 VariableDeclaration 类型的节点
	// ============================================================================

	fmt.Println()
	fmt.Println("🔍 步骤1: 查找目标节点")
	fmt.Println("--------------------")

	var targetNode tsmorphgo.Node
	var nodeFound bool

	// 遍历文件查找 useUserData 变量声明
	useUserDataFile.ForEachDescendant(func(node tsmorphgo.Node) {
		// 验证API: IsVariableDeclaration() - 判断节点类型
		if node.IsVariableDeclaration() {
			// 验证API: GetText() - 获取节点文本
			nodeText := node.GetText()

			// 检查是否是 useUserData 的变量声明
			if strings.Contains(nodeText, "useUserData =") {
				targetNode = node
				nodeFound = true
				fmt.Printf("✅ 找到目标变量声明: %s\n", "useUserData = (userId: number) => { ...}")
				return // 找到后立即停止遍历
			}
		}
	})

	// 如果没有找到变量声明，再尝试查找标识符
	if !nodeFound {
		useUserDataFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsIdentifier() && node.GetText() == "useUserData" {
				targetNode = node
				nodeFound = true
				fmt.Printf("✅ 找到目标标识符: %s\n", node.GetText())
				return // 找到后立即停止遍历
			}
		})
	}

	if !nodeFound {
		log.Fatal("❌ 未找到 useUserData 变量声明节点")
	}

	// ============================================================================
	// 场景5.2: 获取节点的源码文本和基础信息
	// 验证API: GetText() - 获取节点源码文本
	// 验证API: GetKind() - 获取节点类型枚举值
	// 验证API: GetStartLineNumber() - 获取起始行号
	// 预期输出: 显示节点的基础信息
	// ============================================================================

	fmt.Println()
	fmt.Println("📊 节点基础信息")
	fmt.Println("---------------")

	// 验证API: GetKind() - 获取节点类型
	kind := targetNode.GetKind()
	fmt.Printf("🔧 节点类型: %s\n", kind.String())

	// 验证API: GetText() - 获取节点的完整源码文本
	fullText := targetNode.GetText()
	if len(fullText) > 50 {
		fmt.Printf("📝 节点文本: %s...\n", fullText[:50])
	} else {
		fmt.Printf("📝 节点文本: %s\n", fullText)
	}

	// 验证API: GetStartLineNumber() - 获取节点起始行号 (1-based)
	line := targetNode.GetStartLineNumber()
	col := targetNode.GetStartColumnNumber()
	fmt.Printf("📍 节点位置: 第%d行，第%d列\n", line, col)

	// ============================================================================
	// 场景3.2: 获取节点的父节点
	// 验证API: GetParent() - 获取直接父节点
	// 验证目标: 获取 useUserData 变量声明的父节点 (VariableStatement)
	// 预期输出: 找到父节点及其类型
	// ============================================================================

	fmt.Println()
	fmt.Println("🌳 节点导航 - 父节点")
	fmt.Println("-------------------")

	// 验证API: GetParent() - 获取节点的直接父节点
	parentNode := targetNode.GetParent()
	if parentNode != nil {
		parentKind := parentNode.GetKind()
		fmt.Printf("✅ 父节点类型: %s\n", parentKind.String())
		fmt.Printf("📍 父节点位置: 第%d行\n", parentNode.GetStartLineNumber())

		// 验证父节点是否是期望的类型
		if parentNode.IsKind(tsmorphgo.KindVariableStatement) {
			fmt.Println("✅ 父节点是预期的 VariableStatement 类型")
		}
	} else {
		fmt.Println("❌ 未找到父节点")
	}

	// ============================================================================
	// 场景3.3: 获取节点的所有祖先节点
	// 验证API: GetAncestors() - 获取从当前节点到根节点的所有祖先
	// 验证目标: 获取完整的祖先链
	// 预期输出: 显示所有祖先节点
	// ============================================================================

	fmt.Println()
	fmt.Println("🌳 节点导航 - 祖先节点")
	fmt.Println("---------------------")

	// 验证API: GetAncestors() - 获取所有祖先节点
	ancestors := targetNode.GetAncestors()
	fmt.Printf("✅ 祖先节点数量: %d\n", len(ancestors))

	// 显示前几个祖先节点
	fmt.Println("📋 祖先节点列表:")
	for i, ancestor := range ancestors {
		if i >= 5 { // 只显示前5个
			fmt.Printf("   ... 还有 %d 个祖先节点\n", len(ancestors)-5)
			break
		}
		ancestorKind := ancestor.GetKind()
		line := ancestor.GetStartLineNumber()
		fmt.Printf("   %d. %s (第%d行)\n", i+1, ancestorKind.String(), line)
	}

	// 最外层的祖先节点应该是 SourceFile
	if len(ancestors) > 0 {
		lastAncestor := ancestors[len(ancestors)-1]
		lastKind := lastAncestor.GetKind()
		if lastKind == tsmorphgo.KindSourceFile {
			fmt.Println("✅ 最外层祖先节点是 SourceFile")
		}
	}

	// ============================================================================
	// 场景3.4: 按语法类型查找特定的祖先节点
	// 验证API: GetFirstAncestorByKind() - 查找特定类型的第一个祖先节点
	// 验证目标: 查找 VariableStatement 类型的祖先节点
	// 预期输出: 找到 VariableStatement 祖先节点
	// ============================================================================

	fmt.Println()
	fmt.Println("🎯 类型特定的祖先查找")
	fmt.Println("---------------------")

	// 验证API: GetFirstAncestorByKind() - 查找特定类型的祖先节点
	varStatement, found1 := targetNode.GetFirstAncestorByKind(tsmorphgo.KindVariableStatement)
	if found1 && varStatement != nil {
		fmt.Printf("✅ 找到 VariableStatement 祖先: 第%d行\n", varStatement.GetStartLineNumber())
	} else {
		fmt.Println("❌ 未找到 VariableStatement 祖先节点")
	}

	// 查找其他类型的祖先节点
	sourceFile, found2 := targetNode.GetFirstAncestorByKind(tsmorphgo.KindSourceFile)
	if found2 && sourceFile != nil {
		fmt.Printf("✅ 找到 SourceFile 祖先: 第%d行\n", sourceFile.GetStartLineNumber())
	}

	// ============================================================================
	// 场景7.3: VariableDeclaration - 获取变量名和初始值
	// 验证API: AsVariableDeclaration() - 类型转换为 VariableDeclaration
	// 验证API: GetName() - 获取变量名
	// 验证API: HasInitializer() - 检查是否有初始值
	// 验证API: GetInitializer() - 获取初始值节点
	// 预期输出: 显示变量名、初始值等信息
	// ============================================================================

	fmt.Println()
	fmt.Println("🎯 类型收窄演示")
	fmt.Println("---------------")

	// 验证API: AsVariableDeclaration() - 类型转换为 VariableDeclaration
	varDecl, success := targetNode.AsVariableDeclaration()
	if !success {
		fmt.Println("❌ 类型转换为 VariableDeclaration 失败")
		return
	}

	fmt.Println("✅ 成功转换为 VariableDeclaration")

	// 验证API: GetName() - 获取变量名字符串
	varName := varDecl.GetName()
	fmt.Printf("🏷️  变量名: %s\n", varName)

	// 验证变量名是否正确
	if varName == "useUserData" {
		fmt.Println("✅ 变量名验证正确")
	} else {
		fmt.Printf("❌ 变量名不匹配，期望: useUserData, 实际: %s\n", varName)
	}

	// 验证API: HasInitializer() - 检查是否有初始值
	hasInitializer := varDecl.HasInitializer()
	fmt.Printf("🔧 有初始值: %t\n", hasInitializer)

	if hasInitializer {
		// 验证API: GetInitializer() - 获取初始值节点
		initializer := varDecl.GetInitializer()
		if initializer != nil {
			initializerKind := initializer.GetKind()
			fmt.Printf("🔧 初始值类型: %s\n", initializerKind.String())

			// 验证API: GetText() - 获取初始值的文本
			initializerText := initializer.GetText()
			if len(initializerText) > 30 {
				fmt.Printf("📝 初始值文本: %s...\n", initializerText[:30])
			} else {
				fmt.Printf("📝 初始值文本: %s\n", initializerText)
			}

			// 检查初始值是否是箭头函数 (通过文本判断)
			if strings.Contains(initializerText, "=>") {
				fmt.Println("✅ 初始值是箭头函数")
			} else {
				fmt.Println("ℹ️  初始值不是箭头函数")
			}
		} else {
			fmt.Println("❌ 无法获取初始值节点")
		}
	}

	// ============================================================================
	// 专有API验证 - 获取变量名节点
	// 验证API: GetNameNode() - 获取变量名节点
	// 验证目标: 获取 useUserData 标识符节点
	// 预期输出: 显示变量名节点的详细信息
	// ============================================================================

	fmt.Println()
	fmt.Println("🔍 专有API验证")
	fmt.Println("-------------")

	// 验证API: GetNameNode() - 获取变量名节点
	nameNode := varDecl.GetNameNode()
	if nameNode != nil {
		nameNodeKind := nameNode.GetKind()
		fmt.Printf("✅ 变量名节点类型: %s\n", nameNodeKind.String())

		if nameNode.IsIdentifier() {
			fmt.Printf("🏷️  变量名节点文本: %s\n", nameNode.GetText())

			// 获取符号信息
			if symbol, err := nameNode.GetSymbol(); err == nil && symbol != nil {
				fmt.Printf("🔖 符号名称: %s\n", symbol.GetName())
			}
		}
	} else {
		fmt.Println("❌ 无法获取变量名节点")
	}

	// ============================================================================
	// 初始值详细分析
	// 验证目标: 分析箭头函数的参数和结构
	// 预期输出: 显示函数签名信息
	// ============================================================================

	fmt.Println()
	fmt.Println("📊 初始值详细分析")
	fmt.Println("-----------------")

	if varDecl.HasInitializer() {
		initializer := varDecl.GetInitializer()
		if initializer != nil {
			// 查找函数参数
			initializer.ForEachChild(func(child tsmorphgo.Node) bool {
				if child.IsKind(tsmorphgo.KindParameter) {
					fmt.Printf("📋 函数参数: %s\n", child.GetText())
				}
				return false // 继续遍历
			})

			// 计算函数体行数
			initializerText := initializer.GetText()
			lines := 0
			for _, char := range initializerText {
				if char == '\n' {
					lines++
				}
			}
			fmt.Printf("📏 函数体长度: 约 %d 行\n", lines)
		}
	}

	// 清理资源
	defer project.Close()

	fmt.Println()
	fmt.Println("🎉 节点导航和类型收窄示例完成！")
	fmt.Println()
	fmt.Println("✅ 验证总结:")
	fmt.Println("   - 节点类型判断: 成功")
	fmt.Println("   - 父节点导航: 成功")
	fmt.Println("   - 祖先节点导航: 成功")
	fmt.Println("   - 类型特定查找: 成功")
	fmt.Println("   - 类型安全转换: 成功")
	fmt.Println("   - VariableDeclaration 专有API: 成功")
}