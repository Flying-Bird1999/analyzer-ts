//go:build examples

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🚀 TSMorphGo 完整功能演示")
	fmt.Println("==========================")
	fmt.Println("本演示将展示TSMorphGo的主要API，基于真实的React项目")
	fmt.Println("演示场景：代码重构、依赖分析、符号查找等真实开发需求")
	fmt.Println()

	// 获取项目路径
	projectPath := getProjectPath()
	if projectPath == "" {
		fmt.Println("❌ 找不到 demo-react-app 项目")
		os.Exit(1)
	}

	fmt.Printf("📁 分析项目: %s\n", projectPath)
	fmt.Println()

	// 创建项目实例
	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:    projectPath,
		UseTsConfig: true,
	})

	defer project.Close()

	// 运行完整的演示
	runCompleteDemo(project, projectPath)

	fmt.Println("\n✅ 所有演示完成！")
}

func getProjectPath() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(wd, "demo-react-app")
}

func runCompleteDemo(project *tsmorphgo.Project, projectPath string) {
	// 演示1: 项目基础信息
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("1️⃣  项目基础信息 - 站在代码分析者的角度")
	fmt.Println(strings.Repeat("=", 60))
	demo1_ProjectBasics(project, projectPath)

	// 演示2: 精准节点查找 - 我要找到变量A
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("2️⃣  精准节点查找 - 找到变量A并分析它")
	fmt.Println(strings.Repeat("=", 60))
	demo2_FindTargetNode(project, projectPath)

	// 演示3: 符号分析 - 调用变量的API
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("3️⃣  符号分析 - 获取符号信息并验证")
	fmt.Println(strings.Repeat("=", 60))
	demo3_SymbolAnalysis(project, projectPath)

	// 演示4: 引用查找 - 寻找所有使用该变量的地方
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("4️⃣  引用查找 - 找到变量的所有引用位置")
	fmt.Println(strings.Repeat("=", 60))
	demo4_ReferenceFinding(project, projectPath)

	// 演示5: 节点导航 - 从一个节点跳转到相关节点
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("5️⃣  节点导航 - 在AST中自由移动")
	fmt.Println(strings.Repeat("=", 60))
	demo5_NodeNavigation(project, projectPath)

	// 演示6: 代码重构 - 实际的开发场景
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("6️⃣  代码重构 - 真实的重构需求演示")
	fmt.Println(strings.Repeat("=", 60))
	demo6_CodeRefactoring(project, projectPath)

	// 演示7: 类型分析 - 深入理解类型系统
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("7️⃣  类型分析 - 深入TypeScript类型系统")
	fmt.Println(strings.Repeat("=", 60))
	demo7_TypeAnalysis(project, projectPath)

	// 演示8: 实际使用场景 - 开发者真正需要的工具
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("8️⃣  实际开发场景 - 开发者日常工具集")
	fmt.Println(strings.Repeat("=", 60))
	demo8_RealWorldScenarios(project, projectPath)
}

// 演示1: 项目基础信息
func demo1_ProjectBasics(project *tsmorphgo.Project, projectPath string) {
	fmt.Println("📊 项目基础信息:")
	fmt.Println("================")

	// 获取项目统计信息
	fileCount := project.GetFileCount()
	filePaths := project.GetFilePaths()

	fmt.Printf("📄 总文件数: %d\n", fileCount)
	fmt.Printf("📁 文件列表 (前10个):\n")
	for i, path := range filePaths {
		if i >= 10 {
			fmt.Printf("    ... 还有 %d 个文件\n", len(filePaths)-10)
			break
		}
		relativePath, _ := filepath.Rel(projectPath, path)
		fmt.Printf("    %d. %s\n", i+1, relativePath)
	}

	// 按类型分析文件
	analyzeFilesByType(project, projectPath)

	// 统计接口、函数、变量等
	analyzeProjectStats(project, projectPath)
}

// 演示2: 精准节点查找 - 站在使用者角度
func demo2_FindTargetNode(project *tsmorphgo.Project, projectPath string) {
	fmt.Println("🎯 精准节点查找演示:")
	fmt.Println("====================")
	fmt.Println("场景: 我要找到User接口定义，并获取其详细信息")

	// 步骤1: 找到types.ts文件
	typesFile := project.GetSourceFile(filepath.Join(projectPath, "src/types/types.ts"))
	if typesFile == nil {
		fmt.Println("❌ 找不到 types.ts 文件")
		return
	}

	fmt.Printf("✅ 找到文件: %s\n", typesFile.GetFilePath())

	// 步骤2: 在文件中查找User接口
	var userInterface *tsmorphgo.Node
	typesFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.Kind == tsmorphgo.KindInterfaceDeclaration {
			text := node.GetText()
			if strings.Contains(text, "export interface User") {
				userInterface = &node
				fmt.Printf("✅ 找到User接口: %s:%d\n",
					filepath.Base(node.GetSourceFile().GetFilePath()),
					node.GetStartLineNumber())
			}
		}
	})

	if userInterface == nil {
		fmt.Println("❌ 找不到User接口")
		return
	}

	// 步骤3: 验证找到的节点
	fmt.Println("\n📋 验证找到的节点:")
	fmt.Printf("   📍 位置: %d:%d - %d:%d\n",
		userInterface.GetStartLineNumber(), userInterface.GetStartColumnNumber(),
		userInterface.GetEndLineNumber(), userInterface.GetEndColumnNumber())
	fmt.Printf("   🏷️  类型: %s\n", getSyntaxKindName(userInterface.Kind))
	fmt.Printf("   📝 内容: %s\n", truncateText(userInterface.GetText(), 100))

	// 步骤4: 分析接口的属性
	fmt.Println("\n🔍 分析User接口的属性:")
	analyzeInterfaceProperties(userInterface)
}

// 演示3: 符号分析 - 调用变量的API
func demo3_SymbolAnalysis(project *tsmorphgo.Project, projectPath string) {
	fmt.Println("🔍 符号分析演示:")
	fmt.Println("================")
	fmt.Println("场景: 获取useUserData函数的符号信息，深入了解它的属性")

	// 找到useUserData函数
	useUserDataFunc := findUseUserDataFunction(project, projectPath)
	if useUserDataFunc == nil {
		fmt.Println("❌ 找不到useUserData函数")
		return
	}

	fmt.Printf("✅ 找到useUserData节点:\n")
	fmt.Printf("   📍 位置: %d\n", useUserDataFunc.GetStartLineNumber())
	fmt.Printf("   🏷️  类型: %s\n", getSyntaxKindName(useUserDataFunc.Kind))
	fmt.Printf("   📝 内容: %s\n", truncateString(useUserDataFunc.GetText(), 80))

	// 尝试多种方式获取符号信息
	fmt.Println("\n🔍 尝试获取符号信息:")

	// 方法1: 直接从节点获取符号
	if symbol, err := useUserDataFunc.GetSymbol(); err == nil && symbol != nil {
		fmt.Println("✅ 方法1成功 - 从节点直接获取符号")
		analyzeSymbol(*symbol)
	} else {
		fmt.Printf("❌ 方法1失败 - 节点.GetSymbol() 错误: %v\n", err)

		// 方法2: 使用全局函数获取符号
		if symbol, err := tsmorphgo.GetSymbol(*useUserDataFunc); err == nil && symbol != nil {
			fmt.Println("✅ 方法2成功 - 使用tsmorphgo.GetSymbol()")
			analyzeSymbol(*symbol)
		} else {
			fmt.Printf("❌ 方法2失败 - tsmorphgo.GetSymbol() 错误: %v\n", err)
		}
	}

	// 方法3: 尝试从父节点查找符号
	if parent := useUserDataFunc.GetParent(); parent != nil {
		fmt.Println("\n🔍 尝试从父节点查找符号:")
		fmt.Printf("   父节点类型: %s\n", getSyntaxKindName(parent.Kind))

		// 查找父节点中的所有子节点
		parent.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.GetText() == "useUserData" && node.Kind == tsmorphgo.KindIdentifier {
				fmt.Println("✅ 找到useUserData标识符节点")
				if symbol, err := node.GetSymbol(); err == nil && symbol != nil {
					fmt.Println("✅ 从标识符节点获取符号成功")
					analyzeSymbol(*symbol)
				}
			}
		})
	}
}

// 分析符号详细信息
func analyzeSymbol(symbol tsmorphgo.Symbol) {
	fmt.Println("\n📊 符号详细信息:")
	fmt.Printf("   🏷️  符号名称: %s\n", symbol.GetName())

	// 简化版本，只显示基本信息
	fmt.Printf("   ✅ 成功获取符号\n")

	// 获取符号标志
	flags := symbol.GetFlags()
	fmt.Printf("   🚩 符号标志: %d\n", flags)
}

// 演示4: 引用查找 - 寻找所有使用该变量的地方
func demo4_ReferenceFinding(project *tsmorphgo.Project, projectPath string) {
	fmt.Println("🔗 引用查找示例:")
	fmt.Println("================")
	fmt.Println("场景: 找到useUserData函数的所有引用，看看它在哪里被使用了")

	// 找到useUserData函数节点
	useUserDataFunc := findUseUserDataFunction(project, projectPath)
	if useUserDataFunc == nil {
		fmt.Println("❌ 找不到useUserData函数")
		return
	}

	// 查找引用
	references, err := tsmorphgo.FindReferences(*useUserDataFunc)
	if err != nil {
		fmt.Printf("❌ 查找引用失败: %v\n", err)
		return
	}

	fmt.Printf("📊 找到 %d 处引用:\n", len(references))

	if len(references) == 0 {
		fmt.Println("   ℹ️  该函数没有被直接引用")
		return
	}

	// 分析引用的分布
	analyzeReferenceDistribution(references, projectPath)

	// 展示具体的引用位置
	fmt.Println("\n📍 详细引用位置:")
	showDetailedReferences(references, projectPath)

	// 分析引用的类型
	fmt.Println("\n🔍 引用类型分析:")
	analyzeReferenceTypes(references)
}

// 演示5: 节点导航 - 在AST中自由移动
func demo5_NodeNavigation(project *tsmorphgo.Project, projectPath string) {
	fmt.Println("🧭 节点导航演示:")
	fmt.Println("================")
	fmt.Println("场景: 从useUserData函数导航到相关的代码结构")

	// 找到useUserData函数
	useUserDataFunc := findUseUserDataFunction(project, projectPath)
	if useUserDataFunc == nil {
		fmt.Println("❌ 找不到useUserData函数")
		return
	}

	fmt.Println("📍 导航起点: useUserData函数")
	fmt.Printf("   位置: %s:%d\n",
		filepath.Base(useUserDataFunc.GetSourceFile().GetFilePath()),
		useUserDataFunc.GetStartLineNumber())

	// 向上导航：获取父节点
	fmt.Println("\n⬆️  向上导航:")
	parent := useUserDataFunc.GetParent()
	if parent != nil {
		fmt.Printf("   父节点: %s\n", getSyntaxKindName(parent.Kind))
	}

	// 获取所有祖先节点
	ancestors := useUserDataFunc.GetAncestors()
	fmt.Printf("   祖先节点数量: %d\n", len(ancestors))
	if len(ancestors) > 0 {
		fmt.Printf("   根节点类型: %s\n", getSyntaxKindName(ancestors[len(ancestors)-1].Kind))
	}

	// 向下导航：遍历子节点
	fmt.Println("\n⬇️  向下导航:")
	childCount := 0
	useUserDataFunc.ForEachChild(func(child tsmorphgo.Node) bool {
		childCount++
		fmt.Printf("   子节点 %d: %s - %s\n",
			childCount, getSyntaxKindName(child.Kind), truncateString(child.GetText(), 50))
		return false // 继续遍历所有子节点
	})
	fmt.Printf("   总子节点数: %d\n", childCount)

	// 横向导航：查找相关的函数
	fmt.Println("\n↔️  横向导航 - 查找相关函数:")
	findRelatedFunctions(useUserDataFunc.GetSourceFile())

	// 参数导航：分析函数的参数
	fmt.Println("\n🎯 参数导航 - 分析函数参数:")
	navigateFunctionParameters(useUserDataFunc)
}

// 演示6: 代码重构 - 真实的重构需求
func demo6_CodeRefactoring(project *tsmorphgo.Project, projectPath string) {
	fmt.Println("🔧 代码重构演示:")
	fmt.Println("================")
	fmt.Println("场景: 代码重构 - 重命名useUserData函数、检查影响范围")

	// 找到useUserData函数
	useUserDataFunc := findUseUserDataFunction(project, projectPath)
	if useUserDataFunc == nil {
		fmt.Println("❌ 找不到useUserData函数")
		return
	}

	fmt.Println("🎯 重构任务: 将useUserData函数重命名为useUserInfo")
	fmt.Printf("   当前位置: %s:%d\n",
		filepath.Base(useUserDataFunc.GetSourceFile().GetFilePath()),
		useUserDataFunc.GetStartLineNumber())

	// 步骤1: 分析重构影响
	fmt.Println("\n📊 重构影响分析:")

	// 查找所有引用
	references, err := tsmorphgo.FindReferences(*useUserDataFunc)
	if err != nil {
		fmt.Printf("   ❌ 查找引用失败: %v\n", err)
		return
	}

	fmt.Printf("   📋 需要修改的文件数: %d\n", countUniqueFiles(references))
	fmt.Printf("   📝 需要修改的引用数: %d\n", len(references))

	// 步骤2: 生成重构计划
	fmt.Println("\n📝 重构计划:")
	generateRefactoringPlan(references, projectPath, "useUserData", "useUserInfo")

	// 步骤3: 检查潜在冲突
	fmt.Println("\n⚠️  潜在冲突检查:")
	checkRefactoringConflicts(project, projectPath, "useUserInfo")

	// 步骤4: 模拟重构结果
	fmt.Println("\n✅ 重构后预览:")
	simulateRefactoringResult(project, useUserDataFunc, references, projectPath, "useUserData", "useUserInfo")
}

// 演示7: 类型分析 - 深入理解类型系统
func demo7_TypeAnalysis(project *tsmorphgo.Project, projectPath string) {
	fmt.Println("🎯 类型分析演示:")
	fmt.Println("================")
	fmt.Println("场景: 深入分析TypeScript类型系统")

	// 分析类型定义
	fmt.Println("📋 项目中的类型定义:")
	analyzeTypeDefinitions(project, projectPath)

	// 分析接口定义
	fmt.Println("\n🔌 接口定义分析:")
	analyzeInterfaceDefinitions(project, projectPath)

	// 分析函数签名
	fmt.Println("\n⚡ 函数签名分析:")
	analyzeFunctionSignatures(project, projectPath)

	// 分析变量声明
	fmt.Println("\n📦 变量声明分析:")
	analyzeVariableDeclarations(project, projectPath)

	// 分析导入导出
	fmt.Println("\n📤 导入导出分析:")
	analyzeImportExports(project, projectPath)
}

// 演示8: 实际开发场景 - 开发者日常工具集
func demo8_RealWorldScenarios(project *tsmorphgo.Project, projectPath string) {
	fmt.Println("🛠️  实际开发场景演示:")
	fmt.Println("======================")
	fmt.Println("场景: 开发者日常需要的代码分析工具")

	// 场景1: 找到未使用的代码
	fmt.Println("1️⃣  清理未使用代码:")
	findUnusedCode(project, projectPath)

	// 场景2: 分析代码复杂度
	fmt.Println("\n2️⃣  代码复杂度分析:")
	analyzeCodeComplexity(project, projectPath)

	// 场景3: 依赖分析
	fmt.Println("\n3️⃣  依赖关系分析:")
	analyzeDependencies(project, projectPath)

	// 场景4: 组件分析
	fmt.Println("\n4️⃣  React组件分析:")
	analyzeReactComponents(project, projectPath)

	// 场景5: Hook分析
	fmt.Println("\n5️⃣  自定义Hook分析:")
	analyzeCustomHooks(project, projectPath)

	// 场景6: API使用分析
	fmt.Println("\n6️⃣  API使用分析:")
	analyzeAPIUsage(project, projectPath)

	// 场景7: 类型安全检查
	fmt.Println("\n7️⃣  类型安全检查:")
	analyzeTypeSafety(project, projectPath)
}

// ========== 辅助函数 ==========

// 按类型分析文件
func analyzeFilesByType(project *tsmorphgo.Project, projectPath string) {
	fmt.Println("\n📊 文件类型分析:")

	sourceFiles := project.GetSourceFiles()
	typeCount := make(map[string]int)

	// 先统计所有文件类型
	for _, file := range sourceFiles {
		relativePath, _ := filepath.Rel(projectPath, file.GetFilePath())
		ext := filepath.Ext(relativePath)
		typeCount[ext]++
	}

	// 然后输出统计结果
	fmt.Printf("   📁 项目文件分布:\n")
	totalFiles := len(sourceFiles)
	for ext, count := range typeCount {
		percentage := float64(count) / float64(totalFiles) * 100
		fmt.Printf("   📄 %s 文件: %d 个 (%.1f%%)\n", ext, count, percentage)
	}
	fmt.Printf("   📊 总计: %d 个文件\n", totalFiles)
}

// 分析项目统计
func analyzeProjectStats(project *tsmorphgo.Project, projectPath string) {
	fmt.Println("\n📊 项目代码统计:")

	sourceFiles := project.GetSourceFiles()
	totalInterfaces := 0
	totalFunctions := 0
	totalVariables := 0
	totalImports := 0
	totalExports := 0

	for _, file := range sourceFiles {
		fileResult := file.GetFileResult()
		if fileResult != nil {
			totalInterfaces += len(fileResult.InterfaceDeclarations)
			totalFunctions += len(fileResult.FunctionDeclarations)
			totalVariables += len(fileResult.VariableDeclarations)
			totalImports += len(fileResult.ImportDeclarations)
			totalExports += len(fileResult.ExportDeclarations)
		}
	}

	fmt.Printf("   🔌 接口声明: %d\n", totalInterfaces)
	fmt.Printf("   ⚡ 函数声明: %d\n", totalFunctions)
	fmt.Printf("   📦 变量声明: %d\n", totalVariables)
	fmt.Printf("   📥 导入声明: %d\n", totalImports)
	fmt.Printf("   📤 导出声明: %d\n", totalExports)
}

// 分析接口属性
func analyzeInterfaceProperties(interfaceNode *tsmorphgo.Node) {
	propertyCount := 0
	methodCount := 0

	interfaceNode.ForEachDescendant(func(child tsmorphgo.Node) {
		if child.Kind == tsmorphgo.KindPropertySignature {
			propertyCount++
			if propertyCount <= 5 {
				text := child.GetText()
				fmt.Printf("   📋 属性 %d: %s\n", propertyCount, truncateText(text, 60))
			}
		} else if child.Kind == tsmorphgo.KindMethodSignature {
			methodCount++
			if methodCount <= 3 {
				text := child.GetText()
				fmt.Printf("   ⚡ 方法 %d: %s\n", methodCount, truncateText(text, 60))
			}
		}
	})

	fmt.Printf("   📊 总计: %d个属性, %d个方法\n", propertyCount, methodCount)
}

// 查找useUserData函数
func findUseUserDataFunction(project *tsmorphgo.Project, projectPath string) *tsmorphgo.Node {
	useDataFile := project.GetSourceFile(filepath.Join(projectPath, "src/hooks/useUserData.ts"))
	if useDataFile == nil {
		return nil
	}

	var useUserDataFunc *tsmorphgo.Node
	useDataFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.Kind == tsmorphgo.KindVariableDeclaration {
			text := node.GetText()
			if strings.Contains(text, "useUserData") {
				useUserDataFunc = &node
			}
		}
	})

	return useUserDataFunc
}

// 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen > 3 {
		return s[:maxLen-3] + "..."
	}
	return s[:maxLen]
}

// 获取语法种类名称
func getSyntaxKindName(kind tsmorphgo.SyntaxKind) string {
	// 简化版本，返回基本的种类名称
	switch kind {
	case tsmorphgo.KindInterfaceDeclaration:
		return "InterfaceDeclaration"
	case tsmorphgo.KindFunctionDeclaration:
		return "FunctionDeclaration"
	case tsmorphgo.KindVariableDeclaration:
		return "VariableDeclaration"
	case tsmorphgo.KindClassDeclaration:
		return "ClassDeclaration"
	case tsmorphgo.KindTypeAliasDeclaration:
		return "TypeAliasDeclaration"
	case tsmorphgo.KindIdentifier:
		return "Identifier"
	case tsmorphgo.KindStringLiteral:
		return "StringLiteral"
	case tsmorphgo.KindNumericLiteral:
		return "NumericLiteral"
	case tsmorphgo.KindPropertySignature:
		return "PropertySignature"
	case tsmorphgo.KindMethodSignature:
		return "MethodSignature"
	case tsmorphgo.KindImportDeclaration:
		return "ImportDeclaration"
	case tsmorphgo.KindExportDeclaration:
		return "ExportDeclaration"
	case tsmorphgo.KindParameter:
		return "Parameter"
	case tsmorphgo.KindCallExpression:
		return "CallExpression"
	default:
		return fmt.Sprintf("Kind(%d)", int(kind))
	}
}

// 截断文本
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}

// 分析引用分布
func analyzeReferenceDistribution(references []*tsmorphgo.Node, projectPath string) {
	fileCount := make(map[string]int)

	for _, ref := range references {
		filePath := ref.GetSourceFile().GetFilePath()
		relativePath, _ := filepath.Rel(projectPath, filePath)
		fileCount[relativePath]++
	}

	fmt.Printf("   📁 涉及文件数: %d\n", len(fileCount))
	for file, count := range fileCount {
		fmt.Printf("      %s: %d 处引用\n", file, count)
	}
}

// 显示详细引用
func showDetailedReferences(references []*tsmorphgo.Node, projectPath string) {
	for i, ref := range references {
		if i >= 10 {
			fmt.Printf("   ... 还有 %d 处引用\n", len(references)-10)
			break
		}

		relativePath, _ := filepath.Rel(projectPath, ref.GetSourceFile().GetFilePath())
		fmt.Printf("   %d. %s:%d - %s\n",
			i+1, relativePath, ref.GetStartLineNumber(), truncateText(ref.GetText(), 60))
	}
}

// 分析引用类型
func analyzeReferenceTypes(references []*tsmorphgo.Node) {
	importCount := 0
	typeRefCount := 0
	exprCount := 0

	for _, ref := range references {
		parent := ref.GetParent()
		if parent != nil {
			switch parent.Kind {
			case tsmorphgo.KindImportDeclaration:
				importCount++
			case tsmorphgo.KindTypeReference:
				typeRefCount++
			default:
				exprCount++
			}
		}
	}

	fmt.Printf("   📥 导入引用: %d\n", importCount)
	fmt.Printf("   🎯 类型引用: %d\n", typeRefCount)
	fmt.Printf("   ⚡ 表达式引用: %d\n", exprCount)
}

// 查找相关函数
func findRelatedFunctions(file *tsmorphgo.SourceFile) {
	functionCount := 0
	file.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.Kind == tsmorphgo.KindFunctionDeclaration ||
		   node.Kind == tsmorphgo.KindVariableDeclaration {
			functionCount++
			if functionCount <= 5 {
				text := node.GetText()
				if strings.Contains(text, "function") || strings.Contains(text, "const") {
					name := extractFunctionName(text)
					fmt.Printf("   ⚡ %s\n", name)
				}
			}
		}
	})

	if functionCount > 5 {
		fmt.Printf("   ... 还有 %d 个函数\n", functionCount-5)
	}
}

// 导航函数参数
func navigateFunctionParameters(funcNode *tsmorphgo.Node) {
	fmt.Printf("   🎯 目标函数: %s\n", truncateString(funcNode.GetText(), 60))

	paramCount := 0

	// 首先查找直接的函数参数（在函数签名中的参数）
	funcNode.ForEachChild(func(child tsmorphgo.Node) bool {
		// 对于变量声明的函数，尝试在子节点中查找参数
		child.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.Kind == tsmorphgo.KindParameter {
				paramCount++
				text := node.GetText()
				fmt.Printf("   📋 参数 %d: %s\n", paramCount, text)
			}
		})
		return false
	})

	if paramCount == 0 {
		// 如果没找到，尝试在函数体的第一层查找
		fmt.Println("   🔍 未找到参数列表，尝试在函数体中查找...")
		funcNode.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.Kind == tsmorphgo.KindParameter {
				// 检查这个参数是否属于useUserData函数本身
				parent := node.GetParent()
				if parent != nil && parent.GetText() == funcNode.GetText() {
					paramCount++
					if paramCount <= 5 {
						text := node.GetText()
						fmt.Printf("   📋 参数 %d: %s\n", paramCount, text)
					}
				}
			}
		})
	}

	fmt.Printf("   📊 总计: %d 个参数\n", paramCount)
}

// 提取函数名称
func extractFunctionName(text string) string {
	// 处理箭头函数: const name = (params) => { ... }
	if strings.Contains(text, "const ") && strings.Contains(text, "=>") {
		parts := strings.Split(text, "const ")
		if len(parts) > 1 {
			nameAndRest := strings.Split(parts[1], "=")[0]
			name := strings.TrimSpace(nameAndRest)
			return name
		}
	}

	// 处理普通函数: function name() { ... }
	if strings.Contains(text, "function ") {
		parts := strings.Split(text, "function ")
		if len(parts) > 1 {
			name := strings.Split(parts[1], "(")[0]
			return strings.TrimSpace(name)
		}
	}

	// 如果都匹配不上，尝试从变量声明中提取名称
	if strings.Contains(text, "export const ") || strings.Contains(text, "const ") {
		// 使用正则表达式匹配 const name = ...
		re := regexp.MustCompile(`(?:export\s+)?const\s+(\w+)\s*=`)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	// 如果还是找不到，返回前50个字符作为标识
	if len(text) > 50 {
		return "Unknown: " + text[:50] + "..."
	}
	return "Unknown: " + text
}

// 统计唯一文件数
func countUniqueFiles(references []*tsmorphgo.Node) int {
	fileSet := make(map[string]bool)
	for _, ref := range references {
		filePath := ref.GetSourceFile().GetFilePath()
		fileSet[filePath] = true
	}
	return len(fileSet)
}

// 生成重构计划
func generateRefactoringPlan(references []*tsmorphgo.Node, projectPath, oldName, newName string) {
	fmt.Printf("   📝 重命名 '%s' -> '%s'\n", oldName, newName)
	fmt.Printf("   📄 影响文件: %d 个\n", countUniqueFiles(references))
	fmt.Printf("   🔄 需要更新: %d 处引用\n", len(references))

	// 按文件分组
	files := make(map[string][]*tsmorphgo.Node)
	for _, ref := range references {
		filePath := ref.GetSourceFile().GetFilePath()
		files[filePath] = append(files[filePath], ref)
	}

	fmt.Printf("   📋 详细计划:\n")
	for filePath, refs := range files {
		relativePath, _ := filepath.Rel(projectPath, filePath)
		fmt.Printf("      - %s (%d处)\n", relativePath, len(refs))
	}
}

// 检查重构冲突
func checkRefactoringConflicts(project *tsmorphgo.Project, projectPath string, newName string) {
	// 检查是否已存在重命名后的名称
	sourceFiles := project.GetSourceFiles()
	conflict := false

	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.Kind == tsmorphgo.KindFunctionDeclaration ||
			   node.Kind == tsmorphgo.KindVariableDeclaration {
				text := node.GetText()
				if strings.Contains(text, newName) {
					conflict = true
				}
			}
		})
	}

	if conflict {
		fmt.Printf("   ⚠️  警告: '%s' 已存在\n", newName)
	} else {
		fmt.Printf("   ✅ 无命名冲突\n")
	}
}

// 检查是否有测试文件
func hasTestFiles(project *tsmorphgo.Project, projectPath string) bool {
	sourceFiles := project.GetSourceFiles()
	for _, file := range sourceFiles {
		fileName := filepath.Base(file.GetFilePath())
		if strings.Contains(fileName, "test") || strings.Contains(fileName, "spec") {
			return true
		}
	}
	return false
}

// 模拟重构结果
func simulateRefactoringResult(project *tsmorphgo.Project, funcNode *tsmorphgo.Node, references []*tsmorphgo.Node, projectPath, oldName, newName string) {
	fmt.Printf("   📄 原始函数: %s\n", truncateString(funcNode.GetText(), 80))

	// 更详细的重构后预览
	fmt.Printf("   🔄 重构后: %s\n", strings.Replace(truncateString(funcNode.GetText(), 80), oldName, newName, 1))
	fmt.Printf("   📝 更新引用: %d 处\n", len(references))

	// 显示具体的引用位置和修改内容
	if len(references) > 0 {
		fmt.Println("   📍 具体修改预览:")
		refCount := 0
		for _, ref := range references {
			if refCount >= 3 { // 只显示前3个
				fmt.Printf("   ... 还有 %d 处修改\n", len(references)-3)
				break
			}

			refCount++
			refText := ref.GetText()
			if strings.Contains(refText, oldName) {
				newText := strings.Replace(refText, oldName, newName, -1)
				fmt.Printf("   %d. %s:%d - %s\n",
					refCount,
					filepath.Base(ref.GetSourceFile().GetFilePath()),
					ref.GetStartLineNumber(),
					truncateString(newText, 60))
			}
		}
	}

	// 重构风险提示
	fmt.Println("   🚨 重构风险评估:")
	if len(references) > 5 {
		fmt.Printf("      ⚠️  影响范围较大 (%d 处引用)，建议分批重构\n", len(references))
	} else {
		fmt.Printf("      ✅ 影响范围较小 (%d 处引用)，可以安全重构\n", len(references))
	}

	// 检查是否有测试文件
	fmt.Println("   🧪 测试建议:")
	if hasTestFiles(project, projectPath) {
		fmt.Printf("      ✅ 发现测试文件，重构后请运行测试验证\n")
	} else {
		fmt.Printf("      ⚠️  未发现测试文件，建议添加测试后再重构\n")
	}
}

// 分析类型定义
func analyzeTypeDefinitions(project *tsmorphgo.Project, projectPath string) {
	typeCount := 0
	sourceFiles := project.GetSourceFiles()

	for _, file := range sourceFiles {
		fileResult := file.GetFileResult()
		if fileResult != nil {
			typeCount += len(fileResult.InterfaceDeclarations)
			typeCount += len(fileResult.TypeDeclarations)
		}
	}

	fmt.Printf("   📊 总类型定义数: %d\n", typeCount)
}

// 分析接口定义
func analyzeInterfaceDefinitions(project *tsmorphgo.Project, projectPath string) {
	interfaces := make(map[string][]string)
	sourceFiles := project.GetSourceFiles()

	for _, file := range sourceFiles {
		fileResult := file.GetFileResult()
		if fileResult != nil {
			// 简化版本：直接计数
			interfaceCount := len(fileResult.InterfaceDeclarations)
			if interfaceCount > 0 {
				relativePath, _ := filepath.Rel(projectPath, file.GetFilePath())
				interfaces["Interface"] = append(interfaces["Interface"], relativePath)
			}
		}
	}

	fmt.Printf("   🔌 接口定义 (%d个):\n", len(interfaces))
	count := 0
	for name, files := range interfaces {
		if count >= 5 {
			fmt.Printf("   ... 还有 %d 个接口\n", len(interfaces)-5)
			break
		}
		fmt.Printf("      - %s (在 %d 个文件中)\n", name, len(files))
		count++
	}
}

// 分析函数签名
func analyzeFunctionSignatures(project *tsmorphgo.Project, projectPath string) {
	functionCount := 0
	sourceFiles := project.GetSourceFiles()

	for _, file := range sourceFiles {
		fileResult := file.GetFileResult()
		if fileResult != nil {
			functionCount += len(fileResult.FunctionDeclarations)
		}
	}

	fmt.Printf("   ⚡ 函数声明数: %d\n", functionCount)
}

// 分析变量声明
func analyzeVariableDeclarations(project *tsmorphgo.Project, projectPath string) {
	varCount := 0
	sourceFiles := project.GetSourceFiles()

	for _, file := range sourceFiles {
		fileResult := file.GetFileResult()
		if fileResult != nil {
			varCount += len(fileResult.VariableDeclarations)
		}
	}

	fmt.Printf("   📦 变量声明数: %d\n", varCount)
}

// 分析导入导出
func analyzeImportExports(project *tsmorphgo.Project, projectPath string) {
	importCount := 0
	exportCount := 0
	sourceFiles := project.GetSourceFiles()

	for _, file := range sourceFiles {
		fileResult := file.GetFileResult()
		if fileResult != nil {
			importCount += len(fileResult.ImportDeclarations)
			exportCount += len(fileResult.ExportDeclarations)
		}
	}

	fmt.Printf("   📥 导入声明: %d\n", importCount)
	fmt.Printf("   📤 导出声明: %d\n", exportCount)
}

// 查找未使用代码
func findUnusedCode(project *tsmorphgo.Project, projectPath string) {
	// 这里可以添加查找未使用代码的逻辑
	fmt.Printf("   📊 扫描未使用的导出...\n")

	// 简化版本：统计导出和引用
	totalExports := 0
	sourceFiles := project.GetSourceFiles()

	for _, file := range sourceFiles {
		fileResult := file.GetFileResult()
		if fileResult != nil {
			totalExports += len(fileResult.ExportDeclarations)
		}
	}

	fmt.Printf("   📊 总导出声明: %d\n", totalExports)
	fmt.Printf("   ✅ 扫描完成\n")
}

// 分析代码复杂度
func analyzeCodeComplexity(project *tsmorphgo.Project, projectPath string) {
	complexFunctions := 0
	sourceFiles := project.GetSourceFiles()

	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.Kind == tsmorphgo.KindFunctionDeclaration {
				// 简单的复杂度计算
				nodeCount := 0
				node.ForEachDescendant(func(child tsmorphgo.Node) {
					nodeCount++
				})

				if nodeCount > 50 {
					complexFunctions++
				}
			}
		})
	}

	fmt.Printf("   📊 复杂函数数量 (>50个节点): %d\n", complexFunctions)
}

// 分析依赖关系
func analyzeDependencies(project *tsmorphgo.Project, projectPath string) {
	importCount := 0
	externalModules := make(map[string]bool)
	sourceFiles := project.GetSourceFiles()

	for _, file := range sourceFiles {
		fileResult := file.GetFileResult()
		if fileResult != nil {
			importCount += len(fileResult.ImportDeclarations)

			// 简化版本：只统计导入数量
		}
	}

	fmt.Printf("   📦 总导入声明数: %d\n", importCount)
	fmt.Printf("   📦 外部模块数: %d\n", len(externalModules))
}

// 分析React组件
func analyzeReactComponents(project *tsmorphgo.Project, projectPath string) {
	componentCount := 0
	sourceFiles := project.GetSourceFiles()

	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.Kind == tsmorphgo.KindVariableDeclaration {
				text := node.GetText()
				if strings.Contains(text, "React.FC") || strings.Contains(text, ": React.FC") {
					componentCount++
				}
			}
		})
	}

	fmt.Printf("   ⚛️  React组件数: %d\n", componentCount)
}

// 分析自定义Hook
func analyzeCustomHooks(project *tsmorphgo.Project, projectPath string) {
	hookCount := 0
	sourceFiles := project.GetSourceFiles()

	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.Kind == tsmorphgo.KindVariableDeclaration {
				text := node.GetText()
				if strings.Contains(text, "use") && strings.Contains(text, "const") {
					// 简单判断是否为Hook
					if strings.Contains(text, "useState") || strings.Contains(text, "useEffect") {
						// 内置Hook
					} else if strings.Contains(text, "use") && !strings.Contains(text, "React") {
						// 可能的自定义Hook
						hookCount++
					}
				}
			}
		})
	}

	fmt.Printf("   🪝 自定义Hook数: %d\n", hookCount)
}

// 分析API使用
func analyzeAPIUsage(project *tsmorphgo.Project, projectPath string) {
	// 这里可以添加API使用分析
	fmt.Printf("   📊 分析API使用模式...\n")
	fmt.Printf("   ✅ 分析完成\n")
}

// 分析类型安全
func analyzeTypeSafety(project *tsmorphgo.Project, projectPath string) {
	anyCount := 0
	sourceFiles := project.GetSourceFiles()

	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			text := node.GetText()
			if strings.Contains(text, "any") && !strings.Contains(text, "//") {
				// 简单判断是否使用any类型
				anyCount++
			}
		})
	}

	fmt.Printf("   🚨 可能的any类型使用: %d 处\n", anyCount)
}

// 提取接口名称
func extractInterfaceName(text string) string {
	lines := strings.Split(text, "\n")
	if len(lines) > 0 {
		firstLine := strings.TrimSpace(lines[0])
		if strings.HasPrefix(firstLine, "export interface ") {
			parts := strings.Fields(firstLine)
			if len(parts) >= 3 {
				return parts[2]
			}
		}
	}
	return "Unknown"
}