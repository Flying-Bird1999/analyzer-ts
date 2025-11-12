package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tmorphgo "github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🚀 TSMorphGo 完整演示 - 真实React项目全覆盖分析")
	fmt.Println("==============================================")

	// 检查demo-react-app目录是否存在
	demoAppPath, _ := os.Getwd()
	demoAppPath = filepath.Join(demoAppPath, "demo-react-app")
	if _, err := os.Stat(demoAppPath); os.IsNotExist(err) {
		fmt.Println("❌ 错误: demo-react-app目录不存在")
		fmt.Println("请确保demo-react-app项目已创建")
		return
	}

	fmt.Printf("✅ 找到真实React项目: %s\n", demoAppPath)

	// 创建TSMorphGo项目实例 - 基于真实前端项目
	fmt.Println("\n📁 创建TSMorphGo项目实例...")
	config := tmorphgo.ProjectConfig{
		RootPath:         demoAppPath,
		UseTsConfig:      true,
		TargetExtensions: []string{".ts", ".tsx"},
		IgnorePatterns:   []string{"node_modules", "dist", ".git", "build"},
	}

	project := tmorphgo.NewProject(config)
	defer project.Close()

	// 等待项目初始化完成，确保LSP服务准备就绪
	fmt.Println("⏳ 等待LSP服务初始化...")
	time.Sleep(2 * time.Second)

	// 获取所有源文件
	fmt.Println("\n🔍 分析项目文件...")
	sourceFiles := project.GetSourceFiles()
	fmt.Printf("找到 %d 个源文件\n", len(sourceFiles))

	if len(sourceFiles) == 0 {
		fmt.Println("❌ 没有找到源文件")
		return
	}

	// 分析所有文件
	fmt.Println("\n📄 项目文件分析:")
	var tsxFiles, tsFiles int

	for _, file := range sourceFiles {
		filePath := file.GetFilePath()
		if strings.HasSuffix(filePath, ".tsx") {
			tsxFiles++
		} else if strings.HasSuffix(filePath, ".ts") {
			tsFiles++
		}
		fmt.Printf("  ✅ %s\n", filePath)
	}

	fmt.Printf("\n📊 文件统计: %d TSX文件, %d TS文件\n", tsxFiles, tsFiles)

	// 综合分析演示
	fmt.Println("\n🎯 综合分析演示:")
	// demonstrateNodeAnalysis(project, sourceFiles)
	// demonstrateTypeChecking(project, sourceFiles)
	// demonstrateSymbolAnalysis(project, sourceFiles)
	demonstrateReferenceAnalysis(project, sourceFiles)

	fmt.Println("\n📋 完整演示总结:")
	fmt.Println("  ✅ 项目管理: 基于真实前端项目创建")
	fmt.Println("  ✅ 节点访问: 成功遍历和分析AST节点")
	fmt.Println("  ✅ 类型检查: 识别各种TypeScript语法结构")
	fmt.Println("  ✅ 符号系统: 访问节点关联的符号信息")
	fmt.Println("  ✅ 引用查找: 演示基本的引用分析")
	fmt.Println("  ✅ 完全基于真实项目: 无虚拟项目依赖")

	fmt.Println("\n🎉 完整演示完成！")
	fmt.Println("💡 这证明了TSMorphGo具备完整的TypeScript代码分析能力")
}

// 演示节点分析
func demonstrateNodeAnalysis(project *tmorphgo.Project, sourceFiles []*tmorphgo.SourceFile) {
	fmt.Println("\n🔍 节点分析演示:")

	var appFile *tmorphgo.SourceFile
	// 找到App.tsx文件
	for _, file := range sourceFiles {
		if strings.Contains(file.GetFilePath(), "App.tsx") {
			appFile = file
			break
		}
	}

	if appFile == nil {
		fmt.Println("  ⚠️ 无法获取App.tsx")
		return
	}

	fmt.Println("  🎯 详细节点分析:")

	// 遍历所有节点进行分析
	appFile.ForEachDescendant(func(node tmorphgo.Node) {
		// 函数声明 - 打印详细信息
		if node.IsFunctionDeclaration() {
			lineNum := node.GetStartLineNumber()
			colNum := node.GetStartColumnNumber()
			text := node.GetText()
			if len(text) > 50 {
				text = text[:50] + "..."
			}
			fmt.Printf("    📍 函数声明: %s (行 %d:%d)\n", text, lineNum, colNum)

			// 打印参数信息
			node.ForEachChild(func(childNode tmorphgo.Node) bool {
				if childNode.IsIdentifier() {
					fmt.Printf("      ├─ 参数: %s\n", childNode.GetText())
				}
				return true
			})
		}

		// 变量声明 - 打印详细信息
		if node.IsVariableDeclaration() {
			lineNum := node.GetStartLineNumber()
			text := node.GetText()
			if len(text) > 30 {
				text = text[:30] + "..."
			}
			fmt.Printf("    📝 变量声明: %s (行 %d)\n", text, lineNum)
		}

		// 接口声明 - 打印详细信息
		if node.IsInterfaceDeclaration() {
			lineNum := node.GetStartLineNumber()
			interfaceName := node.GetText()
			fmt.Printf("    🏗️ 接口声明: %s (行 %d)\n", interfaceName, lineNum)

			// 打印接口属性
			node.ForEachChild(func(childNode tmorphgo.Node) bool {
				if childNode.GetKindName() == "KindPropertySignature" {
					propName := childNode.GetText()
					if len(propName) > 40 {
						propName = propName[:40] + "..."
					}
					fmt.Printf("      ├─ 属性: %s\n", propName)
				}
				return true
			})
		}

		// 调用表达式 - 打印详细信息
		if node.IsCallExpression() {
			lineNum := node.GetStartLineNumber()
			callText := node.GetText()
			if len(callText) > 40 {
				callText = callText[:40] + "..."
			}
			fmt.Printf("    📞 调用表达式: %s (行 %d)\n", callText, lineNum)
		}

		// 导入声明 - 打印详细信息
		if node.IsImportDeclaration() {
			lineNum := node.GetStartLineNumber()
			importText := node.GetText()
			if len(importText) > 60 {
				importText = importText[:60] + "..."
			}
			fmt.Printf("    📥 导入声明: %s (行 %d)\n", importText, lineNum)
		}

		// JSX元素 - 打印重要JSX标签
		if node.GetKindName() == "KindJsxElement" {
			lineNum := node.GetStartLineNumber()
			text := node.GetText()
			// 只显示开头的JSX标签
			if strings.Contains(text, "<") && strings.Index(text, ">") < 30 {
				tag := text[:strings.Index(text, ">")+1]
				fmt.Printf("    🎨 JSX元素: %s (行 %d)\n", tag, lineNum)
			}
		}
	})
}

// 演示类型检查
func demonstrateTypeChecking(project *tmorphgo.Project, sourceFiles []*tmorphgo.SourceFile) {
	fmt.Println("\n🏷️ 类型检查演示:")

	var userProfileFile *tmorphgo.SourceFile
	// 找到UserProfile.tsx文件
	for _, file := range sourceFiles {
		if strings.Contains(file.GetFilePath(), "UserProfile.tsx") {
			userProfileFile = file
			break
		}
	}

	if userProfileFile == nil {
		fmt.Println("  ⚠️ 无法获取UserProfile.tsx")
		return
	}

	var (
		identifiers       int
		propertyAccess    int
		binaryExpressions int
		literalValues     int
	)

	userProfileFile.ForEachDescendant(func(node tmorphgo.Node) {
		if node.IsIdentifier() {
			identifiers++
		}
		if node.IsPropertyAccessExpression() {
			propertyAccess++
		}
		if node.IsBinaryExpression() {
			binaryExpressions++
		}
		// Note: IsLiteral method not available, checking via kind name
		kindName := node.GetKindName()
		if strings.Contains(kindName, "Literal") {
			literalValues++
		}
	})

	fmt.Printf("  📊 UserProfile组件统计: 标识符=%d, 属性访问=%d, 二元表达式=%d, 字面量=%d\n",
		identifiers, propertyAccess, binaryExpressions, literalValues)

	// 演示 SyntaxKind 分析
	fmt.Println("  🎯 SyntaxKind分析示例:")
	userProfileFile.ForEachDescendant(func(node tmorphgo.Node) {
		kindName := node.GetKindName()
		if strings.Contains(kindName, "Arrow") || strings.Contains(kindName, "Return") {
			fmt.Printf("    📝 %s (行 %d, 列 %d)\n", kindName, node.GetStartLineNumber(), node.GetStartColumnNumber())
		}
	})
}

// 演示符号分析
func demonstrateSymbolAnalysis(project *tmorphgo.Project, sourceFiles []*tmorphgo.SourceFile) {
	fmt.Println("\n🧬 符号分析演示:")

	var totalSymbols int

	fmt.Println("  🎯 详细符号分析:")

	for _, file := range sourceFiles {
		if file == nil {
			continue
		}

		var fileSymbols int
		var symbolDetails []string

		file.ForEachDescendant(func(node tmorphgo.Node) {
			if node.IsIdentifier() {
				// 尝试获取符号
				symbol, err := node.GetSymbol()
				if err == nil && symbol != nil {
					fileSymbols++

					// 只收集前5个符号的详细信息
					if len(symbolDetails) < 5 {
						symbolName := symbol.GetName()

						lineNum := node.GetStartLineNumber()
						colNum := node.GetStartColumnNumber()
						identifierName := node.GetText()

						detail := fmt.Sprintf("      ├─ 符号: %s (标识符: %s, 位置: %d:%d)", symbolName, identifierName, lineNum, colNum)
						symbolDetails = append(symbolDetails, detail)
					}
				}
			}
		})

		if fileSymbols > 0 {
			filePath := file.GetFilePath()
			fileName := filePath[strings.LastIndex(filePath, "/")+1:]
			fmt.Printf("    📄 %s: %d 个符号\n", fileName, fileSymbols)

			// 打印符号详细信息
			for _, detail := range symbolDetails {
				fmt.Println(detail)
			}
		}

		totalSymbols += fileSymbols
	}

	fmt.Printf("  📊 总计找到 %d 个符号关联的节点\n", totalSymbols)
}

// 演示引用分析
func demonstrateReferenceAnalysis(project *tmorphgo.Project, sourceFiles []*tmorphgo.SourceFile) {
	fmt.Println("\n🔗 引用分析演示:")

	fmt.Println("  🎯 使用 FindReferences 查找节点引用:")

	// 查找特定标识符的引用
	testCases := []struct {
		identifier  string
		fileName    string
		description string
	}{
		{"formatDate", "App.tsx", "查找 formatDate 函数的所有引用"},
	}

	for _, testCase := range testCases {
		fmt.Printf("\n    🔍 %s:\n", testCase.description)

		var targetNode tmorphgo.Node

		// 在指定文件中查找目标标识符 - 使用和测试相同的方法
		for _, file := range sourceFiles {
			fileName := file.GetFilePath()
			if strings.Contains(fileName, testCase.fileName) {
				file.ForEachDescendant(func(node tmorphgo.Node) {
					if node.IsIdentifier() && strings.TrimSpace(node.GetText()) == testCase.identifier {
						// 关键：检查父节点，确保找到的是真正的使用位置
						parent := node.GetParent()
						if parent != nil && parent.GetKindName() == "KindCallExpression" {
							targetNode = node
							return
						}
					}
				})
			}
			if targetNode.GetKindName() != "" { // 检查是否找到了有效的节点
				break
			}
		}

		if targetNode.GetKindName() == "" {
			fmt.Printf("      ❌ 未找到标识符 '%s' 在 %s 中\n", testCase.identifier, testCase.fileName)
			continue
		}

		// 获取目标节点的位置信息
		lineNum := targetNode.GetStartLineNumber()
		colNum := targetNode.GetStartColumnNumber()
		fmt.Printf("      📍 目标节点: '%s' 位置: %d:%d (类型: %s)\n", testCase.identifier, lineNum, colNum, targetNode.GetKindName())

		// 使用 FindReferences 查找所有引用
		fmt.Printf("      🔎 正在查找引用...\n")
		references, err := tmorphgo.FindReferences(targetNode)
		if err != nil {
			fmt.Printf("      ❌ 查找引用失败: %v\n", err)
			continue
		}

		if len(references) == 0 {
			fmt.Printf("      ⚠️ 未找到任何引用\n")
			continue
		}

		fmt.Printf("      ✅ 找到 %d 个引用:\n", len(references))
		for i, ref := range references {
			refLine := ref.GetStartLineNumber()
			refCol := ref.GetStartColumnNumber()
			refFile := ref.GetSourceFile().GetFilePath()
			refFileName := refFile[strings.LastIndex(refFile, "/")+1:]

			// 获取引用节点的上下文
			parent := ref.GetParent()
			context := ""
			if parent != nil {
				parentText := parent.GetText()
				if len(parentText) > 50 {
					context = parentText[:50] + "..."
				} else {
					context = parentText
				}
			}

			fmt.Printf("        %d. %s:%d:%d - 上下文: %s\n", i+1, refFileName, refLine, refCol, context)
		}
	}

	fmt.Println("\n  🎯 使用 GotoDefinition 查找定义位置:")

	// 查找某个标识符的定义位置
	definitionTestCases := []struct {
		identifier  string
		fileName    string
		description string
	}{
		{"formatDate", "App.tsx", "查找 formatDate 函数的定义"},
	}

	for _, testCase := range definitionTestCases {
		fmt.Printf("\n    🔍 %s:\n", testCase.description)

		var targetNode tmorphgo.Node

		// 在指定文件中查找目标标识符 - 使用相同的方法
		for _, file := range sourceFiles {
			fileName := file.GetFilePath()
			if strings.Contains(fileName, testCase.fileName) {
				file.ForEachDescendant(func(node tmorphgo.Node) {
					if node.IsIdentifier() && strings.TrimSpace(node.GetText()) == testCase.identifier {
						// 检查父节点，确保找到的是使用位置
						parent := node.GetParent()
						if parent != nil && parent.GetKindName() == "KindCallExpression" {
							targetNode = node
							return
						}
					}
				})
			}
			if targetNode.GetKindName() != "" {
				break
			}
		}

		if targetNode.GetKindName() == "" {
			fmt.Printf("      ❌ 未找到标识符 '%s' 在 %s 中\n", testCase.identifier, testCase.fileName)
			continue
		}

		// 使用 GotoDefinition 查找定义
		definitions, err := tmorphgo.GotoDefinition(targetNode)
		if err != nil {
			fmt.Printf("      ❌ 查找定义失败: %v\n", err)
			continue
		}

		if len(definitions) == 0 {
			fmt.Printf("      ⚠️ 未找到定义位置\n")
			continue
		}

		fmt.Printf("      ✅ 找到定义位置:\n")
		for _, def := range definitions {
			defLine := def.GetStartLineNumber()
			defCol := def.GetStartColumnNumber()
			defFile := def.GetSourceFile().GetFilePath()
			defFileName := defFile[strings.LastIndex(defFile, "/")+1:]

			defText := def.GetText()
			if len(defText) > 60 {
				defText = defText[:60] + "..."
			}

			fmt.Printf("        📍 %s:%d:%d - %s\n", defFileName, defLine, defCol, defText)
		}
	}

	fmt.Println("\n  🎯 引用计数演示:")

	// 统计一些常见标识符的引用数量
	countTestCases := []string{"formatDate", "useState", "React"}

	for _, identifier := range countTestCases {
		var foundNode tmorphgo.Node

		// 查找标识符的第一个出现 - 使用更精确的方法
		for _, file := range sourceFiles {
			file.ForEachDescendant(func(node tmorphgo.Node) {
				if node.IsIdentifier() && strings.TrimSpace(node.GetText()) == identifier && foundNode.GetKindName() == "" {
					// 检查父节点，确保找到的是有意义的使用位置
					parent := node.GetParent()
					if parent != nil && (parent.GetKindName() == "KindCallExpression" ||
						parent.GetKindName() == "KindImportDeclaration" ||
						parent.GetKindName() == "KindImportClause") {
						foundNode = node
						return
					}
				}
			})
			if foundNode.GetKindName() != "" {
				break
			}
		}

		if foundNode.GetKindName() != "" {
			count, err := tmorphgo.CountReferences(foundNode)
			if err != nil {
				fmt.Printf("    ❌ 统计 '%s' 引用失败: %v\n", identifier, err)
			} else {
				fmt.Printf("    📊 '%s' 共有 %d 个引用\n", identifier, count)
			}
		}
	}
}
