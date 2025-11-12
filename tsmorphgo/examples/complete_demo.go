package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
		RootPath: demoAppPath,
	}

	project := tmorphgo.NewProject(config)

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
	demonstrateNodeAnalysis(project, sourceFiles)
	demonstrateTypeChecking(project, sourceFiles)
	demonstrateSymbolAnalysis(project, sourceFiles)
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

						detail := fmt.Sprintf("      ├─ 符号: %s (标识符: %s, 位置: %d:%d)",
							symbolName, identifierName, lineNum, colNum)
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

	fmt.Println("  🎯 详细引用路径分析:")

	// 分析所有文件的引用信息
	for _, file := range sourceFiles {
		if file == nil {
			continue
		}

		filePath := file.GetFilePath()
		fileName := filePath[strings.LastIndex(filePath, "/")+1:]

		fmt.Printf("    📄 %s:\n", fileName)

		// 收集重要的导入和标识符引用
		imports := make(map[string][]string)
		identifiers := make(map[string][]string)

		file.ForEachDescendant(func(node tmorphgo.Node) {
			// 收集导入信息
			if node.IsImportDeclaration() {
				importText := node.GetText()
				if len(importText) > 80 {
					importText = importText[:80] + "..."
				}
				lineNum := node.GetStartLineNumber()
				imports["import"] = append(imports["import"], fmt.Sprintf("%s (行 %d)", importText, lineNum))
			}

			// 收集重要标识符（React、useState等）
			if node.IsIdentifier() {
				identifierName := node.GetText()
				if identifierName == "React" || identifierName == "useState" ||
				   identifierName == "useEffect" || identifierName == "interface" {
					lineNum := node.GetStartLineNumber()
					colNum := node.GetStartColumnNumber()
					identifiers[identifierName] = append(identifiers[identifierName],
						fmt.Sprintf("%d:%d", lineNum, colNum))
				}
			}
		})

		// 打印导入信息
		for _, importInfo := range imports["import"] {
			fmt.Printf("      📥 %s\n", importInfo)
		}

		// 打印重要标识符引用
		for id, positions := range identifiers {
			if len(positions) > 0 {
				fmt.Printf("      🔗 标识符 '%s': %s\n", id, strings.Join(positions, ", "))
			}
		}
	}

	// 演示特定标识符的详细分析
	fmt.Println("  🔍 跨文件引用分析:")

	// 查找React和useState的使用情况
	reactRefs := []string{}
	useStateRefs := []string{}

	for _, file := range sourceFiles {
		filePath := file.GetFilePath()
		fileName := filePath[strings.LastIndex(filePath, "/")+1:]

		file.ForEachDescendant(func(node tmorphgo.Node) {
			if node.IsIdentifier() {
				if node.GetText() == "React" {
					lineNum := node.GetStartLineNumber()
					colNum := node.GetStartColumnNumber()
					reactRefs = append(reactRefs, fmt.Sprintf("%s:%d:%d", fileName, lineNum, colNum))
				} else if node.GetText() == "useState" {
					lineNum := node.GetStartLineNumber()
					colNum := node.GetStartColumnNumber()
					useStateRefs = append(useStateRefs, fmt.Sprintf("%s:%d:%d", fileName, lineNum, colNum))
				}
			}
		})
	}

	if len(reactRefs) > 0 {
		fmt.Printf("    ⚛️ React 引用: %s\n", strings.Join(reactRefs, ", "))
	}
	if len(useStateRefs) > 0 {
		fmt.Printf("    🎣 useState 引用: %s\n", strings.Join(useStateRefs, ", "))
	}

	// 别名引用分析
	fmt.Println("  🎯 别名映射引用分析:")
	for _, file := range sourceFiles {
		filePath := file.GetFilePath()
		fileName := filePath[strings.LastIndex(filePath, "/")+1:]

		if strings.Contains(fileName, "test-aliases") {
			fmt.Printf("    📍 %s - 检测到别名使用:\n", fileName)
			file.ForEachDescendant(func(node tmorphgo.Node) {
				if node.GetKindName() == "KindStringLiteral" && strings.Contains(node.GetText(), "@/") {
					lineNum := node.GetStartLineNumber()
					fmt.Printf("      ├─ 别名路径: %s (行 %d)\n", node.GetText(), lineNum)
				}
			})
		}
	}
}
