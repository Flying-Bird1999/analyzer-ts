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
	fmt.Println("🔧 TSMorphGo 真实使用演示")
	fmt.Println("====================")

	// 获取项目路径
	projectPath := getProjectPath()
	if projectPath == "" {
		log.Fatal("❌ 找不到 demo-react-app 项目")
	}

	fmt.Printf("📁 项目路径: %s\n", projectPath)

	// 创建项目实例
	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:    projectPath,
		UseTsConfig: true,
	})

	defer project.Close()

	// === 演示 1: 分析特定文件 ===
	fmt.Println("\n1️⃣  分析特定文件 - App.tsx")
	fmt.Println("---------------------------")

	analyzeFile(project, "src/components/App.tsx")

	// === 演示 2: 手动查找和分析符号 ===
	fmt.Println("\n2️⃣  手动查找 User 类型定义")
	fmt.Println("-----------------------")

	userInterface := findInterface(project, "User")
	if userInterface == nil {
		fmt.Println("❌ 找不到 User 接口")
	} else {
		analyzeInterface(userInterface)
	}

	// === 演示 3: 分析函数符号 ===
	fmt.Println("\n3️⃣  分析 useUserData Hook")
	fmt.Println("-------------------------")

	useUserDataFunc := findFunction(project, "useUserData")
	if useUserDataFunc == nil {
		fmt.Println("❌ 找不到 useUserData 函数")
	} else {
		analyzeFunction(useUserDataFunc)
	}

	// === 演示 4: 查找引用 ===
	fmt.Println("\n4️⃣  查找符号引用 - User 类型")
	fmt.Println("-----------------------------")

	if userInterface != nil {
		findReferences(userInterface, "User")
	}

	// === 演示 5: 分析组件结构 ===
	fmt.Println("\n5️⃣  分析组件结构 - UserProfile")
	fmt.Println("---------------------------------")

	userProfile := findComponent(project, "UserProfile")
	if userProfile == nil {
		fmt.Println("❌ 找不到 UserProfile 组件")
	} else {
		analyzeComponent(userProfile)
	}

	// === 演示 6: 通过文件路径和行列号获取节点 ===
	fmt.Println("\n6️⃣  通过文件路径和行列号获取节点")
	fmt.Println("-------------------------------")

	// 获取 types.ts 中第 2 行的 User 接口定义
	nodeByLocation := getNodeByLocation(project, "src/types/types.ts", 2, 10)
	if nodeByLocation != nil {
		fmt.Printf("✅ 通过位置找到节点: %s\n", getNodeTypeName(nodeByLocation.Kind))
		fmt.Printf("   文本: %s\n", extractContext(*nodeByLocation))
		analyzeNode(nodeByLocation)
	} else {
		fmt.Println("❌ 通过位置未找到节点")
	}

	// === 演示 7: 路径别名分析 ===
	fmt.Println("\n7️⃣  分析路径别名使用")
	fmt.Println("---------------------")

	analyzePathAliases(project)

	fmt.Println("\n✅ 演示完成！这就是 TSMorphGo 的实际使用方式")
}

func getProjectPath() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(wd, "demo-react-app")
}

// analyzeFile 分析单个文件
func analyzeFile(project *tsmorphgo.Project, relativePath string) {
	sourceFile := project.GetSourceFile(filepath.Join(getProjectPath(), relativePath))
	if sourceFile == nil {
		fmt.Printf("❌ 找不到文件: %s\n", relativePath)
		return
	}

	fmt.Printf("📄 文件: %s\n", relativePath)
	fmt.Printf("   📏 行数: %d\n", countFileLines(sourceFile))
	fmt.Printf("   🌟 AST 节点: %d\n", countNodes(sourceFile))

	// 统计不同类型的节点
	counts := make(map[tsmorphgo.SyntaxKind]int)
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		counts[node.Kind]++
	})

	fmt.Printf("   📊 节点类型分布:\n")
	for kind, count := range counts {
		if count > 0 {
			fmt.Printf("      %s: %d\n", getNodeTypeName(kind), count)
		}
	}
}

// findInterface 查找接口定义
func findInterface(project *tsmorphgo.Project, name string) *tsmorphgo.Node {
	// 首先在types.ts中查找接口声明用于显示信息
	typesFile := project.GetSourceFile(filepath.Join(getProjectPath(), "src/types/types.ts"))
	if typesFile != nil {
		typesFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.Kind == tsmorphgo.KindInterfaceDeclaration {
				nodeText := strings.TrimSpace(node.GetText())
				if strings.Contains(nodeText, "export interface "+name+" ") ||
				   strings.Contains(nodeText, "export interface "+name+" {") {
					fmt.Printf("✅ 找到接口 '%s': %s:%d\n", name,
						filepath.Base(node.GetSourceFile().GetFilePath()),
						node.GetStartLineNumber())
				}
			}
		})
	}

	// 然后查找一个能够成功进行引用分析的User标识符节点
	sourceFiles := project.GetSourceFiles()
	var referenceNode *tsmorphgo.Node
	for _, sourceFile := range sourceFiles {
		sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if tsmorphgo.IsIdentifier(node) && strings.TrimSpace(node.GetText()) == name {
				// 尝试获取符号信息，看是否可以成功引用分析
				symbol, err := tsmorphgo.GetSymbol(node)
				if err == nil && symbol != nil {
					references, _ := tsmorphgo.FindReferences(node)
					if len(references) > 0 {
						// 找到了可以成功引用分析的节点
						referenceNode = &node
					}
				}
			}
		})
		if referenceNode != nil {
			return referenceNode
		}
	}

	// 如果上面的方法没找到，回退到使用接口声明节点
	if typesFile != nil {
		var found *tsmorphgo.Node
		typesFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.Kind == tsmorphgo.KindInterfaceDeclaration {
				nodeText := strings.TrimSpace(node.GetText())
				if strings.Contains(nodeText, "export interface "+name+" ") ||
				   strings.Contains(nodeText, "export interface "+name+" {") {
					found = &node
					return
				}
			}
		})
		return found
	}

	fmt.Printf("❌ 找不到接口 '%s'\n", name)
	return nil
}

// findFunction 查找函数定义
func findFunction(project *tsmorphgo.Project, name string) *tsmorphgo.Node {
	// 直接在useUserData.ts中查找useUserData函数
	sourceFile := project.GetSourceFile(filepath.Join(getProjectPath(), "src/hooks/useUserData.ts"))
	if sourceFile == nil {
		fmt.Printf("❌ 找不到 src/hooks/useUserData.ts\n")
		return nil
	}

	var found *tsmorphgo.Node
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		// 查找 export const useUserData
		if node.Kind == tsmorphgo.KindVariableDeclaration {
			nodeText := strings.TrimSpace(node.GetText())
			// 检查是否包含 useUserData
			if strings.Contains(nodeText, name+" =") ||
			   strings.Contains(nodeText, "export const "+name) ||
			   strings.HasPrefix(nodeText, name+":") {
				found = &node
				return
			}
		}
	})

	if found != nil {
		fmt.Printf("✅ 找到函数 '%s': %s:%d\n", name,
			filepath.Base(found.GetSourceFile().GetFilePath()),
			found.GetStartLineNumber())
		return found
	}

	fmt.Printf("❌ 在 useUserData.ts 中找不到函数 '%s'\n", name)
	return nil
}

// findComponent 查找 React 组件
func findComponent(project *tsmorphgo.Project, name string) *tsmorphgo.Node {
	// 直接在UserProfile.tsx中查找UserProfile组件
	sourceFile := project.GetSourceFile(filepath.Join(getProjectPath(), "src/components/UserProfile.tsx"))
	if sourceFile == nil {
		fmt.Printf("❌ 找不到 src/components/UserProfile.tsx\n")
		return nil
	}

	var found *tsmorphgo.Node
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.Kind == tsmorphgo.KindVariableDeclaration {
			nodeText := strings.TrimSpace(node.GetText())
			// 检查是否包含 UserProfile
			if strings.Contains(nodeText, name+": React.FC") ||
			   strings.Contains(nodeText, "export const "+name) ||
			   strings.HasPrefix(nodeText, name+":") {
				found = &node
				return
			}
		}
	})

	if found != nil {
		fmt.Printf("✅ 找到组件 '%s': %s:%d\n", name,
			filepath.Base(found.GetSourceFile().GetFilePath()),
			found.GetStartLineNumber())
		return found
	}

	fmt.Printf("❌ 在 UserProfile.tsx 中找不到组件 '%s'\n", name)
	return nil
}

// analyzeInterface 分析接口
func analyzeInterface(node *tsmorphgo.Node) {
	fmt.Printf("\n📋 接口信息:\n")
	fmt.Printf("   名称: %s\n", getNameFromNode(node))
	fmt.Printf("   位置: %s:%d\n",
		filepath.Base(node.GetSourceFile().GetFilePath()),
		node.GetStartLineNumber())

	// 分析接口成员
	fmt.Printf("   成员:\n")
	node.ForEachDescendant(func(child tsmorphgo.Node) {
		if child.Kind == tsmorphgo.KindPropertySignature ||
		   child.Kind == tsmorphgo.KindMethodSignature {
			memberText := strings.TrimSpace(child.GetText())
			if len(memberText) > 0 && len(memberText) <= 50 {
				fmt.Printf("      - %s\n", memberText)
			}
		}
	})
}

// analyzeFunction 分析函数
func analyzeFunction(node *tsmorphgo.Node) {
	fmt.Printf("\n⚡ 函数信息:\n")
	fmt.Printf("   名称: %s\n", getNameFromNode(node))
	fmt.Printf("   类型: %s\n", getNodeTypeName(node.Kind))
	fmt.Printf("   位置: %s:%d\n",
		filepath.Base(node.GetSourceFile().GetFilePath()),
		node.GetStartLineNumber())

	// 分析参数
	fmt.Printf("   参数:\n")
	node.ForEachDescendant(func(child tsmorphgo.Node) {
		if child.Kind == tsmorphgo.KindParameter {
			paramText := strings.TrimSpace(child.GetText())
			if len(paramText) > 0 {
				fmt.Printf("      - %s\n", paramText)
			}
		}
	})
}

// analyzeComponent 分析 React 组件
func analyzeComponent(node *tsmorphgo.Node) {
	fmt.Printf("\n⚛️  组件信息:\n")
	fmt.Printf("   名称: %s\n", getNameFromNode(node))
	fmt.Printf("   类型: %s\n", getNodeTypeName(node.Kind))
	fmt.Printf("   位置: %s:%d\n",
		filepath.Base(node.GetSourceFile().GetFilePath()),
		node.GetStartLineNumber())

	// 检查是否导出
	isExported := hasReactExport(node)
	fmt.Printf("   导出: %v\n", isExported)

	// 分析 props
	fmt.Printf("   Props:\n")
	node.ForEachDescendant(func(child tsmorphgo.Node) {
		if child.Kind == tsmorphgo.KindParameter {
			propText := strings.TrimSpace(child.GetText())
			if len(propText) > 0 {
				fmt.Printf("      - %s\n", propText)
			}
		}
	})
}

// getNodeByLocation 通过文件路径和行列号获取节点
func getNodeByLocation(project *tsmorphgo.Project, relativePath string, targetLine, targetColumn int) *tsmorphgo.Node {
	sourceFile := project.GetSourceFile(filepath.Join(getProjectPath(), relativePath))
	if sourceFile == nil {
		fmt.Printf("❌ 找不到文件: %s\n", relativePath)
		return nil
	}

	var closestNode *tsmorphgo.Node
	minDistance := int(^uint(0) >> 1) // 最大整数值

	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		startLine := node.GetStartLineNumber()
		startCol := node.GetStartColumnNumber()
		endLine := node.GetEndLineNumber()
		endCol := node.GetEndColumnNumber()

		// 检查目标位置是否在节点范围内
		if (targetLine > startLine || (targetLine == startLine && targetColumn >= startCol)) &&
		   (targetLine < endLine || (targetLine == endLine && targetColumn <= endCol)) {

			// 计算到节点起始位置的距离
			distance := (targetLine-startLine)*(targetLine-startLine) + (targetColumn-startCol)*(targetColumn-startCol)

			if closestNode == nil || distance < minDistance {
				closestNode = &node
				minDistance = distance
			}
		}
	})

	if closestNode != nil {
		fmt.Printf("✅ 找到最匹配的节点: %s:%d:%d - %s\n",
			filepath.Base(relativePath),
			closestNode.GetStartLineNumber(),
			closestNode.GetStartColumnNumber(),
			getNodeTypeName(closestNode.Kind))
		return closestNode
	}

	fmt.Printf("❌ 在 %s:%d:%d 未找到匹配的节点\n", relativePath, targetLine, targetColumn)
	return nil
}

// analyzeNode 分析任意节点
func analyzeNode(node *tsmorphgo.Node) {
	fmt.Printf("\n🔍 节点详细信息:\n")
	fmt.Printf("   类型: %s\n", getNodeTypeName(node.Kind))
	fmt.Printf("   文件: %s\n", filepath.Base(node.GetSourceFile().GetFilePath()))
	fmt.Printf("   位置: %d:%d - %d:%d\n",
		node.GetStartLineNumber(), node.GetStartColumnNumber(),
		node.GetEndLineNumber(), node.GetEndColumnNumber())
	fmt.Printf("   文本: %s\n", extractContext(*node))

	// 尝试获取符号信息
	symbol, err := tsmorphgo.GetSymbol(*node)
	if err != nil {
		fmt.Printf("   符号信息: 获取失败 - %v\n", err)
	} else if symbol != nil {
		fmt.Printf("   符号名称: %s\n", symbol.GetName())
		fmt.Printf("   符号标志: %d\n", symbol.GetFlags())

		// 查找引用
		references, err := tsmorphgo.FindReferences(*node)
		if err != nil {
			fmt.Printf("   引用信息: 查找失败 - %v\n", err)
		} else {
			fmt.Printf("   引用数量: %d\n", len(references))
			if len(references) > 0 {
				fmt.Printf("   引用文件: ")
				fileSet := make(map[string]bool)
				for _, ref := range references {
					fileSet[filepath.Base(ref.GetSourceFile().GetFilePath())] = true
				}
				for file := range fileSet {
					fmt.Printf("%s ", file)
				}
				fmt.Printf("\n")
			}
		}
	} else {
		fmt.Printf("   符号信息: 无符号\n")
	}
}

// findReferences 查找符号引用
func findReferences(node *tsmorphgo.Node, symbolName string) {
	// 获取符号信息
	symbol, err := tsmorphgo.GetSymbol(*node)
	if err != nil {
		fmt.Printf("❌ 获取符号失败: %v\n", err)
		return
	}

	if symbol == nil {
		fmt.Printf("❌ 符号为空\n")
		return
	}

	fmt.Printf("🔗 符号信息:\n")
	fmt.Printf("   名称: %s\n", symbol.GetName())
	fmt.Printf("   标志: %d\n", symbol.GetFlags())

	// 查找引用
	fmt.Printf("   查找引用...\n")
	references, err := tsmorphgo.FindReferences(*node)
	if err != nil {
		fmt.Printf("❌ 查找引用失败: %v\n", err)
		return
	}

	fmt.Printf("   找到 %d 处引用:\n", len(references))

	if len(references) == 0 {
		fmt.Printf("   ℹ️  该符号没有找到引用\n")
		return
	}

	// 按文件分组
	fileRefs := make(map[string][]*tsmorphgo.Node)
	for _, ref := range references {
		filePath := ref.GetSourceFile().GetFilePath()
		fileRefs[filePath] = append(fileRefs[filePath], ref)
	}

	fmt.Printf("   文件分布:\n")
	for filePath, refs := range fileRefs {
		fmt.Printf("      📁 %s (%d 处):\n", filepath.Base(filePath), len(refs))
		for _, ref := range refs {
			context := extractContext(*ref)
			fmt.Printf("         %d: %s\n", ref.GetStartLineNumber(), context)
		}
	}
}

// analyzePathAliases 分析路径别名使用
func analyzePathAliases(project *tsmorphgo.Project) {
	fmt.Printf("📦 分析路径别名使用:\n")

	aliasCount := 0
	aliasExamples := []string{}

	sourceFiles := project.GetSourceFiles()
	for _, sourceFile := range sourceFiles {
		sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.Kind == tsmorphgo.KindImportDeclaration {
				text := node.GetText()
				if strings.Contains(text, "@/") {
					aliasCount++
					if len(aliasExamples) < 3 {
						// 提取前几行
						lines := strings.Split(text, "\n")
						for _, line := range lines {
							line = strings.TrimSpace(line)
							if strings.Contains(line, "@/") && line != "" {
								aliasExamples = append(aliasExamples, line)
								break
							}
						}
					}
				}
			}
		})
	}

	fmt.Printf("   总计: %d 个使用路径别名的导入\n", aliasCount)
	if len(aliasExamples) > 0 {
		fmt.Printf("   示例:\n")
		for _, example := range aliasExamples {
			fmt.Printf("     %s\n", example)
		}
	}
}

// 辅助函数
func getNameFromNode(node *tsmorphgo.Node) string {
	text := strings.TrimSpace(node.GetText())
	if text == "" {
		// 如果节点文本为空，尝试从子节点获取
		var firstChild *tsmorphgo.Node
		node.ForEachChild(func(child tsmorphgo.Node) bool {
			firstChild = &child
			return true // 只要第一个子节点
		})
		if firstChild != nil {
			text = strings.TrimSpace(firstChild.GetText())
		}
	}
	return text
}

func hasReactExport(node *tsmorphgo.Node) bool {
	// 检查父节点是否有导出
	parent := node.GetParent()
	if parent == nil {
		return false
	}

	// 检查附近的导出语句
	context := parent.GetText()
	return strings.Contains(context, "export") &&
		   strings.Contains(context, getNameFromNode(node))
}

func extractContext(node tsmorphgo.Node) string {
	text := strings.TrimSpace(node.GetText())
	if len(text) > 30 {
		return text[:27] + "..."
	}
	return text
}

func getNodeTypeName(kind tsmorphgo.SyntaxKind) string {
	switch kind {
	case tsmorphgo.KindInterfaceDeclaration:
		return "InterfaceDeclaration"
	case tsmorphgo.KindFunctionDeclaration:
		return "FunctionDeclaration"
	case tsmorphgo.KindClassDeclaration:
		return "ClassDeclaration"
	case tsmorphgo.KindVariableDeclaration:
		return "VariableDeclaration"
	case tsmorphgo.KindParameter:
		return "Parameter"
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

func countFileLines(sourceFile *tsmorphgo.SourceFile) int {
	maxLine := 0
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		line := node.GetStartLineNumber()
		if line > maxLine {
			maxLine = line
		}
	})
	return maxLine
}

func countNodes(sourceFile *tsmorphgo.SourceFile) int {
	count := 0
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		count++
	})
	return count
}