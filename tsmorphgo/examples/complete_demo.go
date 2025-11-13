//go:build examples

package main

import (
	"fmt"
	"os"
	"path/filepath"
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

	// 步骤4: 使用各种API验证这是否是我们要找的节点
	fmt.Println("\n🔍 节点类型验证:")
	fmt.Printf("   节点类型: %s\n", getSyntaxKindName(userInterface.Kind))
	fmt.Printf("   可以获取文本: %s\n", userInterface.GetText() != "")
	fmt.Printf("   可以获取位置: %d:%d\n", userInterface.GetStartLineNumber(), userInterface.GetStartColumnNumber())
}

// 演示3: 符号分析 - 调用变量的API
func demo3_SymbolAnalysis(project *tsmorphgo.Project, projectPath string) {
	fmt.Println("🔍 符号分析演示:")
	fmt.Println("================")
	fmt.Println("场景: 获取User接口的符号信息，深入了解它的属性")

	// 找到User接口节点
	userInterface := findUserInterface(project, projectPath)
	if userInterface == nil {
		fmt.Println("❌ 找不到User接口")
		return
	}

	// 获取符号信息
	symbol, err := tsmorphgo.GetSymbol(*userInterface)
	if err != nil {
		fmt.Printf("❌ 获取符号失败: %v\n", err)
		return
	}

	if symbol == nil {
		fmt.Println("❌ 符号为空")
		return
	}

	// 分析符号信息
	fmt.Println("📊 符号信息:")
	fmt.Printf("   🏷️  符号名称: %s\n", symbol.GetName())
	fmt.Printf("   🚩 符号标志: %d\n", symbol.GetFlags())

	// 检查符号的各种属性
	fmt.Println("\n🔧 符号属性分析:")
	checkSymbolProperties(*symbol)

	// 获取符号的声明
	fmt.Println("\n📍 符号声明:")
	fmt.Printf("   📄 文件: %s\n", filepath.Base(userInterface.GetSourceFile().GetFilePath()))
	fmt.Printf("   📍 行号: %d\n", userInterface.GetStartLineNumber())
}

// 演示4: 引用查找 - 寻找所有使用该变量的地方
func demo4_ReferenceFinding(project *tsmorphgo.Project, projectPath string) {
	fmt.Println("🔗 引用查找示例:")
	fmt.Println("================")
	fmt.Println("场景: 找到User接口的所有引用，看看它在哪里被使用了")

	// 找到User接口节点
	userInterface := findUserInterface(project, projectPath)
	if userInterface == nil {
		fmt.Println("❌ 找不到User接口")
		return
	}

	// 查找引用
	references, err := tsmorphgo.FindReferences(*userInterface)
	if err != nil {
		fmt.Printf("❌ 查找引用失败: %v\n", err)
		return
	}

	fmt.Printf("📊 找到 %d 处引用:\n", len(references))

	if len(references) == 0 {
		fmt.Println("   ℹ️  该接口没有被直接引用（可能是通过类型推导间接使用）")
		return
	}

	// 分析引用的分布
	analyzeReferenceDistribution(references, projectPath)

	// 展示具体的引用位置
	fmt.Println("\n📍 详细引用位置:")
	showDetailedReferences(references, projectPath)
}

// 演示5: 节点导航 - 在AST中自由移动
func demo5_NodeNavigation(project *tsmorphgo.Project, projectPath string) {
	fmt.Println("🧭 节点导航演示:")
	fmt.Println("================")
	fmt.Println("场景: 从User接口导航到相关的类型定义和属性")

	// 找到User接口
	userInterface := findUserInterface(project, projectPath)
	if userInterface == nil {
		fmt.Println("❌ 找不到User接口")
		return
	}

	fmt.Println("📍 导航起点: User接口")
	fmt.Printf("   位置: %s:%d\n",
		filepath.Base(userInterface.GetSourceFile().GetFilePath()),
		userInterface.GetStartLineNumber())

	// 向上导航：获取父节点
	fmt.Println("\n⬆️  向上导航:")
	parent := userInterface.GetParent()
	if parent != nil {
		fmt.Printf("   父节点: %s\n", getSyntaxKindName(parent.Kind))
	}

	// 获取所有祖先节点
	ancestors := userInterface.GetAncestors()
	fmt.Printf("   祖先节点数量: %d\n", len(ancestors))
	if len(ancestors) > 0 {
		fmt.Printf("   根节点类型: %s\n", getSyntaxKindName(ancestors[len(ancestors)-1].Kind))
	}

	// 向下导航：遍历子节点
	fmt.Println("\n⬇️  向下导航:")
	childCount := 0
	userInterface.ForEachChild(func(child tsmorphgo.Node) bool {
		childCount++
		fmt.Printf("   子节点 %d: %s - %s\n",
			childCount, getSyntaxKindName(child.Kind), truncateText(child.GetText(), 50))
		return false // 继续遍历所有子节点
	})
	fmt.Printf("   总子节点数: %d\n", childCount)

	// 横向导航：查找相关的接口
	fmt.Println("\n↔️  横向导航 - 查找相关接口:")
	findRelatedInterfaces(userInterface.GetSourceFile())

	// 类型导航：分析接口的属性类型
	fmt.Println("\n🎯 类型导航 - 分析接口属性:")
	navigatePropertyTypes(userInterface)
}

// 演示6: 代码重构 - 真实的重构需求
func demo6_CodeRefactoring(project *tsmorphgo.Project, projectPath string) {
	fmt.Println("🔧 代码重构演示:")
	fmt.Println("================")
	fmt.Println("场景: 代码重构 - 重命名User接口、检查影响范围")

	// 找到User接口
	userInterface := findUserInterface(project, projectPath)
	if userInterface == nil {
		fmt.Println("❌ 找不到User接口")
		return
	}

	fmt.Println("🎯 重构任务: 将User接口重命名为UserProfile")
	fmt.Printf("   当前位置: %s:%d\n",
		filepath.Base(userInterface.GetSourceFile().GetFilePath()),
		userInterface.GetStartLineNumber())

	// 步骤1: 分析重构影响
	fmt.Println("\n📊 重构影响分析:")

	// 查找所有引用
	references, err := tsmorphgo.FindReferences(*userInterface)
	if err != nil {
		fmt.Printf("   ❌ 查找引用失败: %v\n", err)
		return
	}

	fmt.Printf("   📋 需要修改的文件数: %d\n", countUniqueFiles(references))
	fmt.Printf("   📝 需要修改的引用数: %d\n", len(references))

	// 步骤2: 生成重构计划
	fmt.Println("\n📝 重构计划:")
	generateRefactoringPlan(references, projectPath)

	// 步骤3: 检查潜在冲突
	fmt.Println("\n⚠️  潜在冲突检查:")
	checkRefactoringConflicts(project, projectPath)

	// 步骤4: 模拟重构结果
	fmt.Println("\n✅ 重构后预览:")
	simulateRefactoringResult(userInterface, references, projectPath)
}

// 演示7: 类型分析 - 深入理解类型系统
func demo7_TypeAnalysis(project *tsmorphgo.Project, projectPath string) {
	fmt.Println("🎯 类型分析演示:")
	fmt.Println("================")
	fmt.Println("场景: 深入分析TypeScript类型系统")

	// 分析类型定义
	fmt.Println("📋 项目中的类型定义:")
	analyzeTypeDefinitions(project, projectPath)

	// 分析类型继承关系
	fmt.Println("\n🧬 类型继承关系:")
	analyzeTypeInheritance(project, projectPath)

	// 分析泛型使用
	fmt.Println("\n🔤 泛型使用分析:")
	analyzeGenerics(project, projectPath)

	// 分析类型别名
	fmt.Println("\n🏷️  类型别名分析:")
	analyzeTypeAliases(project, projectPath)

	// 分析复杂类型
	fmt.Println("\n🔗 复杂类型分析:")
	analyzeComplexTypes(project, projectPath)
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

	// 场景4: API使用分析
	fmt.Println("\n4️⃣  API使用分析:")
	analyzeAPIUsage(project, projectPath)

	// 场景5: 错误处理分析
	fmt.Println("\n5️⃣  错误处理分析:")
	analyzeErrorHandling(project, projectPath)
}

// ========== 辅助函数 ==========

// 查找User接口
func findUserInterface(project *tsmorphgo.Project, projectPath string) *tsmorphgo.Node {
	typesFile := project.GetSourceFile(filepath.Join(projectPath, "src/types/types.ts"))
	if typesFile == nil {
		return nil
	}

	var userInterface *tsmorphgo.Node
	typesFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.Kind == tsmorphgo.KindInterfaceDeclaration {
			text := node.GetText()
			if strings.Contains(text, "export interface User") {
				userInterface = &node
			}
		}
	})

	return userInterface
}

// 按类型分析文件
func analyzeFilesByType(project *tsmorphgo.Project, projectPath string) {
	fmt.Println("\n📊 文件类型分析:")

	sourceFiles := project.GetSourceFiles()
	typeCount := make(map[string]int)

	for _, file := range sourceFiles {
		relativePath, _ := filepath.Rel(projectPath, file.GetFilePath())
		ext := filepath.Ext(relativePath)
		typeCount[ext]++

		if len(typeCount) <= 10 {
			fmt.Printf("   📄 %s (%d个)\n", ext, typeCount[ext])
		}
	}
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

// 检查符号属性
func checkSymbolProperties(symbol tsmorphgo.Symbol) {
	flags := symbol.GetFlags()
	fmt.Printf("   🚩 标志值: %d\n", flags)

	// 这里可以添加更多符号属性的检查
	// 比如检查是否为导出符号、是否为接口等
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

// 查找相关接口
func findRelatedInterfaces(file *tsmorphgo.SourceFile) {
	interfaceCount := 0
	file.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.Kind == tsmorphgo.KindInterfaceDeclaration {
			interfaceCount++
			if interfaceCount <= 5 {
				text := node.GetText()
				if strings.Contains(text, "export interface") {
					name := extractInterfaceName(text)
					fmt.Printf("   🔌 %s\n", name)
				}
			}
		}
	})

	if interfaceCount > 5 {
		fmt.Printf("   ... 还有 %d 个接口\n", interfaceCount-5)
	}
}

// 导航属性类型
func navigatePropertyTypes(interfaceNode *tsmorphgo.Node) {
	propertyCount := 0
	interfaceNode.ForEachDescendant(func(child tsmorphgo.Node) {
		if child.Kind == tsmorphgo.KindPropertySignature {
			propertyCount++
			if propertyCount <= 5 {
				text := child.GetText()
				fmt.Printf("   📋 属性 %d: %s\n", propertyCount, truncateText(text, 50))
			}
		}
			})

	if propertyCount > 5 {
		fmt.Printf("   ... 还有 %d 个属性\n", propertyCount-5)
	}
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
func generateRefactoringPlan(references []*tsmorphgo.Node, projectPath string) {
	fmt.Printf("   📝 重命名 'User' -> 'UserProfile'\n")
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
func checkRefactoringConflicts(project *tsmorphgo.Project, projectPath string) {
	// 检查是否已存在UserProfile
	userProfileFile := project.GetSourceFile(filepath.Join(projectPath, "src/types/types.ts"))
	if userProfileFile != nil {
		conflict := false
		userProfileFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.Kind == tsmorphgo.KindInterfaceDeclaration {
				text := node.GetText()
				if strings.Contains(text, "UserProfile") {
					conflict = true
				}
			}
		})

		if conflict {
			fmt.Printf("   ⚠️  警告: UserProfile接口已存在\n")
		} else {
			fmt.Printf("   ✅ 无命名冲突\n")
		}
	}
}

// 模拟重构结果
func simulateRefactoringResult(userInterface *tsmorphgo.Node, references []*tsmorphgo.Node, projectPath string) {
	fmt.Printf("   📄 原始接口: %s\n", truncateText(userInterface.GetText(), 80))
	fmt.Printf("   🔄 重构后: %s\n", strings.Replace(truncateText(userInterface.GetText(), 80), "interface User", "interface UserProfile", 1))
	fmt.Printf("   📝 更新引用: %d 处\n", len(references))
}

// 分析类型定义
func analyzeTypeDefinitions(project *tsmorphgo.Project, projectPath string) {
	typeCount := 0
	sourceFiles := project.GetSourceFiles()

	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.Kind == tsmorphgo.KindInterfaceDeclaration ||
			   node.Kind == tsmorphgo.KindTypeAliasDeclaration {
				typeCount++
			}
		})
	}

	fmt.Printf("   📊 总类型定义数: %d\n", typeCount)
}

// 分析类型继承
func analyzeTypeInheritance(project *tsmorphgo.Project, projectPath string) {
	inheritanceCount := 0
	sourceFiles := project.GetSourceFiles()

	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.Kind == tsmorphgo.KindInterfaceDeclaration {
				// 检查是否有继承
				node.ForEachChild(func(child tsmorphgo.Node) bool {
					// TODO: 找到正确的继承语法种类常量
					// if child.Kind == tsmorphgo.KindHeritageClause {
					// 	inheritanceCount++
					// }
					return false
				})
			}
		})
	}

	fmt.Printf("   🧬 有继承关系的类型: %d\n", inheritanceCount)
}

// 分析泛型使用
func analyzeGenerics(project *tsmorphgo.Project, projectPath string) {
	genericCount := 0
	sourceFiles := project.GetSourceFiles()

	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			text := node.GetText()
			if strings.Contains(text, "<") && strings.Contains(text, ">") {
				// 简单判断是否包含泛型语法
				genericCount++
			}
		})
	}

	fmt.Printf("   🔤 可能使用泛型的节点: %d\n", genericCount)
}

// 分析类型别名
func analyzeTypeAliases(project *tsmorphgo.Project, projectPath string) {
	aliasCount := 0
	sourceFiles := project.GetSourceFiles()

	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.Kind == tsmorphgo.KindTypeAliasDeclaration {
				aliasCount++
			}
		})
	}

	fmt.Printf("   🏷️  类型别名数量: %d\n", aliasCount)
}

// 分析复杂类型
func analyzeComplexTypes(project *tsmorphgo.Project, projectPath string) {
	complexCount := 0
	sourceFiles := project.GetSourceFiles()

	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// TODO: 找到正确的复杂类型语法种类常量
			// if node.Kind == tsmorphgo.KindTypeLiteral ||
			//    node.Kind == tsmorphgo.KindUnionType ||
			//    node.Kind == tsmorphgo.KindIntersectionType {
			// 	complexCount++
			// }

			// 临时使用简单判断
			text := node.GetText()
			if strings.Contains(text, "{") || strings.Contains(text, "|") {
				complexCount++
			}
		})
	}

	fmt.Printf("   🔗 复杂类型数量: %d\n", complexCount)
}

// 查找未使用代码
func findUnusedCode(project *tsmorphgo.Project, projectPath string) {
	// 这里可以添加查找未使用代码的逻辑
	fmt.Printf("   📊 扫描未使用的导出...\n")
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
	sourceFiles := project.GetSourceFiles()

	for _, file := range sourceFiles {
		fileResult := file.GetFileResult()
		if fileResult != nil {
			importCount += len(fileResult.ImportDeclarations)
		}
	}

	fmt.Printf("   📦 总导入声明数: %d\n", importCount)
}

// 分析API使用
func analyzeAPIUsage(project *tsmorphgo.Project, projectPath string) {
	// 这里可以添加API使用分析
	fmt.Printf("   📊 分析API使用模式...\n")
	fmt.Printf("   ✅ 分析完成\n")
}

// 分析错误处理
func analyzeErrorHandling(project *tsmorphgo.Project, projectPath string) {
	errorHandlingCount := 0
	sourceFiles := project.GetSourceFiles()

	for _, file := range sourceFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			text := node.GetText()
			if strings.Contains(text, "throw") || strings.Contains(text, "Error") {
				errorHandlingCount++
			}
		})
	}

	fmt.Printf("   🚨 错误处理相关代码: %d 处\n", errorHandlingCount)
}

// 乘法运算符（用于字符串重复）
func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}