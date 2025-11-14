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
	fmt.Println("🔍 TSMorphGo 引用查找测试")
	fmt.Println("========================")
	fmt.Println("验证符号引用查找能力，确保从1个引用恢复到3个引用")
	fmt.Println()

	// 获取当前工作目录并构建绝对路径
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("❌ 获取当前目录失败")
	}

	projectPath := filepath.Join(wd, "demo-react-app")
	fmt.Printf("📂 项目路径: %s\n", projectPath)

	// 创建项目实例
	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:     projectPath,
		UseTsConfig:  true,
		TsConfigPath: filepath.Join(projectPath, "tsconfig.json"),
	})

	if project == nil {
		log.Fatal("❌ 项目创建失败")
	}

	defer project.Close()

	// 1. 查找useUserData函数的引用
	fmt.Println("🎯 1. useUserData 引用查找")
	fmt.Println("==========================")

	// 直接使用已知位置找到useUserData的定义
	useUserDataDef := project.FindNodeAt(filepath.Join(projectPath, "src/hooks/useUserData.ts"), 10, 13)
	if useUserDataDef == nil {
		log.Fatal("❌ 没有找到useUserData函数定义")
	}

	fmt.Printf("✅ 找到useUserData定义: %s:%d\n",
		useUserDataDef.GetSourceFile().GetFilePath(),
		useUserDataDef.GetStartLineNumber())

	// 查找所有引用
	references, err := tsmorphgo.FindReferences(*useUserDataDef)
	if err != nil {
		log.Printf("❌ 查找引用失败: %v", err)
		return
	}

	fmt.Printf("📊 找到 %d 处引用:\n", len(references))

	// 分析引用详情
	for i, ref := range references {
		fmt.Printf("   %d. %s:%d - %s\n",
			i+1,
			ref.GetSourceFile().GetFilePath(),
			ref.GetStartLineNumber(),
			ref.GetText())
	}

	// 验证引用数量
	if len(references) == 3 {
		fmt.Println("✅ 引用查找正常！找到了预期的3个引用")
	} else if len(references) == 1 {
		fmt.Println("❌ 引用查找有问题！只找到1个引用，应该找到3个")
		fmt.Println("   这表明tsconfig.json配置没有正确传递给LSP服务")
	} else {
		fmt.Printf("⚠️  引用数量异常: %d，预期是3个\n", len(references))
	}

	// 2. 分析引用类型
	fmt.Println()
	fmt.Println("📋 2. 引用类型分析")
	fmt.Println("==================")

	importRefs := 0
	callRefs := 0
	typeRefs := 0

	for _, ref := range references {
		// 简化的上下文判断：基于文本内容
		text := ref.GetText()
		if contains(text, "import") {
			importRefs++
		} else if contains(text, "(") {
			callRefs++
		} else {
			typeRefs++
		}
	}

	fmt.Printf("📥 导入引用: %d\n", importRefs)
	fmt.Printf("⚡ 调用引用: %d\n", callRefs)
	fmt.Printf("🔤 类型引用: %d\n", typeRefs)

	// 3. 跨文件引用分析
	fmt.Println()
	fmt.Println("🔄 3. 跨文件引用分析")
	fmt.Println("====================")

	fileRefs := make(map[string]int)
	for _, ref := range references {
		filePath := ref.GetSourceFile().GetFilePath()
		fileRefs[filePath]++
	}

	fmt.Printf("📁 涉及文件数: %d\n", len(fileRefs))
	for filePath, count := range fileRefs {
		fmt.Printf("   %s: %d 处引用\n", filepath.Base(filePath), count)
	}

	// 4. 路径别名引用验证
	fmt.Println()
	fmt.Println("🔗 4. 路径别名引用验证")
	fmt.Println("======================")

	// 查找使用路径别名导入useUserData的文件
	aliasImports := 0
	for _, ref := range references {
		refText := ref.GetText()
		if contains(refText, "import") && isAliasImport(refText) {
			aliasImports++
			fmt.Printf("✅ 找到路径别名导入: %s\n", refText)
		}
	}

	if aliasImports > 0 {
		fmt.Printf("✅ 路径别名解析正常，找到 %d 个别名导入\n", aliasImports)
	} else {
		fmt.Println("⚠️  没有找到路径别名导入，可能存在解析问题")
	}

	// 5. 查找其他符号的引用进行对比
	fmt.Println()
	fmt.Println("🔍 5. 其他符号引用对比")
	fmt.Println("======================")

	// 查找User接口的引用
	userInterfaceDef := findInterfaceDefinition(project, "User")
	if userInterfaceDef != nil {
		userRefs, err := tsmorphgo.FindReferences(*userInterfaceDef)
		if err == nil {
			fmt.Printf("User接口引用: %d 处\n", len(userRefs))
		}
	}

	// 查找formatDate函数的引用
	formatDateDef := findFunctionDefinition(project, "formatDate")
	if formatDateDef != nil {
		formatDateRefs, err := tsmorphgo.FindReferences(*formatDateDef)
		if err == nil {
			fmt.Printf("formatDate函数引用: %d 处\n", len(formatDateRefs))
		}
	}

	// 6. 重构影响分析
	fmt.Println()
	fmt.Println("🔧 6. 重构影响分析")
	fmt.Println("==================")

	if len(references) > 0 {
		fmt.Printf("📊 如果重命名 'useUserData' → 'getUserInfo'：\n")
		fmt.Printf("   📝 需要修改的文件数: %d\n", len(fileRefs))
		fmt.Printf("   🔄 需要更新的引用数: %d\n", len(references))

		fmt.Printf("\n📋 详细修改计划:\n")
		for i, ref := range references {
			fmt.Printf("   %d. %s:%d - %s\n",
				i+1,
				ref.GetSourceFile().GetFilePath(),
				ref.GetStartLineNumber(),
				ref.GetText())
		}

		// 风险评估
		if len(references) <= 5 {
			fmt.Println("\n✅ 重构风险: 低 - 引用数量较少，可以安全修改")
		} else if len(references) <= 20 {
			fmt.Println("\n⚠️  重构风险: 中 - 引用数量适中，需要仔细测试")
		} else {
			fmt.Println("\n🚨 重构风险: 高 - 引用数量较多，需要全面测试")
		}
	}

	fmt.Println()
	fmt.Println("✅ 引用查找测试完成！")
	fmt.Printf("🎯 关键指标: 找到 %d 个引用（预期: 3个）\n", len(references))

	if len(references) == 3 {
		fmt.Println("🎉 测试通过！引用查找功能正常工作")
	} else {
		fmt.Println("❌ 测试失败！引用查找功能存在问题")
	}
}

// 查找函数定义
func findFunctionDefinition(project *tsmorphgo.Project, functionName string) *tsmorphgo.Node {
	sourceFiles := project.GetSourceFiles()

	for _, file := range sourceFiles {
		// 使用ForEachDescendant遍历文件中的所有节点
		var foundNode *tsmorphgo.Node
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsIdentifier() && node.GetText() == functionName {
				// 检查父节点是否是函数定义或变量声明
				parent := node.GetParent()
				if parent != nil && (parent.IsFunctionDeclaration() || parent.IsVariableDeclaration()) {
					// 简单启发式：检查是否是导出的函数
					if isFunctionDefinition(&node, functionName) {
						foundNode = &node
					}
				}
			}
		})
		if foundNode != nil {
			return foundNode
		}
	}

	return nil
}

// 查找接口定义
func findInterfaceDefinition(project *tsmorphgo.Project, interfaceName string) *tsmorphgo.Node {
	sourceFiles := project.GetSourceFiles()

	for _, file := range sourceFiles {
		// 使用ForEachDescendant遍历文件中的所有节点
		var foundNode *tsmorphgo.Node
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsInterfaceDeclaration() && node.GetText() == interfaceName {
				foundNode = &node
			}
		})
		if foundNode != nil {
			return foundNode
		}
	}

	return nil
}

// 判断是否是函数定义
func isFunctionDefinition(node *tsmorphgo.Node, name string) bool {
	text := node.GetText()

	// 简单的启发式判断
	if node.IsFunctionDeclaration() && contains(text, name) {
		return true
	}

	if node.IsVariableDeclaration() && contains(text, name) && contains(text, "=>") {
		return true
	}

	return false
}

// 判断是否是路径别名导入
func isAliasImport(text string) bool {
	return contains(text, "@/") ||
		contains(text, "@components/") ||
		contains(text, "@hooks/") ||
		contains(text, "@utils/") ||
		contains(text, "@types/")
}

// 简单的字符串包含检查
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
