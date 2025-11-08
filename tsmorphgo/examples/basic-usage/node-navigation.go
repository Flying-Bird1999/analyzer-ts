//go:build node_navigation
// +build node_navigation

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🔍 TSMorphGo 节点导航示例")
	fmt.Println("=" + repeat("=", 50))

	// 使用真实的demo-react-app项目进行演示
	realProjectPath := "/Users/bird/Desktop/alalyzer/analyzer-ts/tsmorphgo/examples/demo-react-app"

	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:         realProjectPath,
		TargetExtensions: []string{".ts", ".tsx"},
		IgnorePatterns:   []string{"node_modules", "dist", ".git", "build"},
		UseTsConfig:      true,
	})
	defer project.Close()

	// 示例1: 基础节点遍历
	fmt.Println("\n🔁 示例1: 基础节点遍历")

	// 获取项目中的所有源文件
	sourceFiles := project.GetSourceFiles()
	if len(sourceFiles) == 0 {
		log.Fatal("未找到任何源文件")
	}

	fmt.Printf("项目包含 %d 个TypeScript文件:\n", len(sourceFiles))

	// 选择第一个有内容的文件进行演示
	var sourceFile *tsmorphgo.SourceFile
	for _, file := range sourceFiles {
		if file != nil {
			sourceFile = file
			break
		}
	}

	if sourceFile == nil {
		log.Fatal("未找到可用的源文件")
	}

	fmt.Printf("分析文件: %s\n", sourceFile.GetFilePath())

	fmt.Printf("遍历文件中的所有函数声明:\n")
	funcCount := 0
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsFunctionDeclaration(node) {
			funcCount++
			if nameNode, ok := tsmorphgo.GetFunctionDeclarationNameNode(node); ok {
				fmt.Printf("  - 函数: %s (行 %d)\n",
					strings.TrimSpace(nameNode.GetText()), node.GetStartLineNumber())
			}
		}
	})
	fmt.Printf("总计发现 %d 个函数声明\n", funcCount)

	// 示例2: 父节点和祖先节点导航
	fmt.Println("\n👆 示例2: 父节点和祖先节点导航")

	// 查找所有标识符
	var identifiers []tsmorphgo.Node
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "loadUserData" {
			nodeCopy := node
			identifiers = append(identifiers, nodeCopy)
		}
	})

	fmt.Printf("找到 %d 个 'loadUserData' 标识符:\n", len(identifiers))
	for i, identifier := range identifiers {
		fmt.Printf("  %d. 位置: 行 %d, 列 %d\n",
			i+1, identifier.GetStartLineNumber(), identifier.GetStartColumnNumber())

		// 获取父节点
		parent := identifier.GetParent()
		if parent != nil {
			fmt.Printf("     父节点类型: %v\n", parent.Kind)
			if tsmorphgo.IsCallExpression(*parent) {
				fmt.Printf("     父节点文本: %s\n", strings.TrimSpace(parent.GetText()))
			}
		}

		// 获取祖先节点
		ancestors := identifier.GetAncestors()
		fmt.Printf("     祖先节点数量: %d\n", len(ancestors))
	}

	// 示例3: 查找特定类型的祖先节点
	fmt.Println("\n🔍 示例3: 查找特定类型的祖先节点")

	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "useState" {
			fmt.Printf("标识符 'useState' 的信息:\n")
			fmt.Printf("  - 位置: 行 %d, 列 %d\n", node.GetStartLineNumber(), node.GetStartColumnNumber())

			// 查找最近的函数声明祖先
			if funcDecl, found := node.GetFirstAncestorByKind(292); found { // FunctionDeclaration
				text := strings.TrimSpace(funcDecl.GetText())
				if len(text) > 50 {
					text = text[:50] + "..."
				}
				fmt.Printf("  - 在函数声明中: %s\n", text)
			}

			// 查找最近的变量声明祖先
			if varDecl, found := node.GetFirstAncestorByKind(221); found { // VariableDeclaration
				text := strings.TrimSpace(varDecl.GetText())
				if len(text) > 50 {
					text = text[:50] + "..."
				}
				fmt.Printf("  - 在变量声明中: %s\n", text)
			}
		}
	})

	// 示例4: 条件遍历和提前终止
	fmt.Println("\n⚡ 示例4: 条件遍历和提前终止")

	// 查找第一个箭头函数
	var arrowFunc *tsmorphgo.Node
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if arrowFunc != nil {
			return // 提前终止遍历
		}

		if node.Kind == 293 { // ArrowFunction
			text := strings.TrimSpace(node.GetText())
			if len(text) > 80 {
				text = text[:80] + "..."
			}
			fmt.Printf("找到箭头函数 (行 %d): %s\n", node.GetStartLineNumber(), text)
			nodeCopy := node
			arrowFunc = &nodeCopy
		}
	})

	if arrowFunc != nil {
		// 分析箭头函数的参数
		paramCount := 0
		arrowFunc.ForEachDescendant(func(descendant tsmorphgo.Node) {
			if descendant.Kind == 218 { // Parameter
				paramCount++
			}
		})
		fmt.Printf("  - 参数数量: %d\n", paramCount)
	} else {
		fmt.Println("未找到箭头函数")
	}

	// 示例5: 深度分析React组件结构
	fmt.Println("\n⚛️ 示例5: 分析React组件结构")

	var reactComponents []struct {
		name      string
		 propsType string
	 line      int
	}

	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsFunctionDeclaration(node) {
			// 检查是否是React组件（返回JSX）
			text := strings.TrimSpace(node.GetText())
			if strings.Contains(text, "React.FC") || strings.Contains(text, "return (") {
				if nameNode, ok := tsmorphgo.GetFunctionDeclarationNameNode(node); ok {
					componentName := strings.TrimSpace(nameNode.GetText())

					// 查找Props接口
					var propsType string
					funcText := node.GetText()
					if strings.Contains(funcText, "React.FC<") {
						start := strings.Index(funcText, "React.FC<") + 8
						end := strings.Index(funcText[start:], ">")
						if end > 0 {
							propsType = funcText[start : start+end]
						}
					}

					reactComponents = append(reactComponents, struct {
						name      string
						propsType string
						line      int
					}{
						name:      componentName,
						propsType: propsType,
						line:      node.GetStartLineNumber(),
					})
				}
			}
		}
	})

	fmt.Printf("发现 %d 个React组件:\n", len(reactComponents))
	for _, component := range reactComponents {
		fmt.Printf("  - %s (行 %d)\n", component.name, component.line)
		if component.propsType != "" {
			fmt.Printf("    Props类型: %s\n", component.propsType)
		}
	}

	fmt.Println("\n✅ 节点导航示例完成!")
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}