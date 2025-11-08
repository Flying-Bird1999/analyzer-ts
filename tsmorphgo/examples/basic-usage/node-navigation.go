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

	// 初始化项目，指向一个真实的React项目目录
	realProjectPath := "/Users/bird/Desktop/alalyzer/analyzer-ts/tsmorphgo/examples/demo-react-app"
	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:         realProjectPath,
		TargetExtensions: []string{".ts", ".tsx"},
		IgnorePatterns:   []string{"node_modules", "dist", ".git", "build"},
		UseTsConfig:      true,
	})
	defer project.Close()

	// 我们选择 App.tsx 作为分析的起点
	sourceFile := project.GetSourceFile(realProjectPath + "/src/App.tsx")
	if sourceFile == nil {
		log.Fatal("未找到 App.tsx 文件")
	}
	fmt.Printf("分析文件: %s\n", sourceFile.GetFilePath())

	// 示例1: 深度优先遍历 (forEachDescendant)
	// ForEachDescendant 是遍历AST最常用的方法之一，它会深度优先访问一个节点下的所有子孙节点。
	// 对应 ts-morph 的 `sourceFile.forEachDescendant(node => { ... })`。
	fmt.Println("\n🔁 示例1: 深度优先遍历")
	fmt.Printf("遍历文件中的所有函数声明:\n")
	funcCount := 0
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		// IsFunctionDeclaration 用于判断节点是否为一个函数声明。
		if tsmorphgo.IsFunctionDeclaration(node) {
			funcCount++
			// GetFunctionDeclarationNameNode 获取函数声明的名称节点。
			if nameNode, ok := tsmorphgo.GetFunctionDeclarationNameNode(node); ok {
				fmt.Printf("  - 函数: %s (行 %d)\n",
					strings.TrimSpace(nameNode.GetText()), node.GetStartLineNumber())
			}
		}
	})
	fmt.Printf("总计发现 %d 个函数声明\n", funcCount)

	// 示例2: 父节点和祖先节点导航 (getParent, getAncestors)
	fmt.Println("\n👆 示例2: 父节点和祖先节点导航")
	// 遍历找到 `useState` 这个标识符，然后查看它的上下文
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "useState" {
			fmt.Printf("\n找到 'useState' 标识符:\n")
			fmt.Printf("  - 位置: 行 %d, 列 %d\n", node.GetStartLineNumber(), node.GetStartColumnNumber())

			// GetParent 获取节点的直接父节点。
			// 对应 ts-morph 的 `node.getParent()`。
			parent := node.GetParent()
			if parent != nil {
				// GetKindName 获取节点类型的可读名称，如 "CallExpression"。
				fmt.Printf("  - 父节点类型: %s\n", parent.GetKindName())
				// IsCallExpression 判断节点是否为函数调用表达式。
				if tsmorphgo.IsCallExpression(*parent) {
					fmt.Printf("  - 父节点是调用表达式: %s\n", strings.TrimSpace(parent.GetText()))
				}
			}

			// GetAncestors 获取从父节点到根节点的所有祖先节点数组。
			// 对应 ts-morph 的 `node.getAncestors()`。
			ancestors := node.GetAncestors()
			fmt.Printf("  - 祖先节点数量: %d\n", len(ancestors))
			if len(ancestors) > 2 {
				fmt.Printf("  - 部分祖先类型: %s -> %s\n", ancestors[0].GetKindName(), ancestors[1].GetKindName())
			}
			return // 只演示一次
		}
	})

	// 示例3: 查找特定类型的祖先节点 (getFirstAncestorByKind)
	// GetFirstAncestorByKind 向上查找第一个匹配指定类型的祖先节点，非常高效。
	// 对应 ts-morph 的 `node.getFirstAncestorByKind(SyntaxKind.Kind)`。
	fmt.Println("\n🔍 示例3: 查找特定类型的祖先节点")
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		// 查找 `users` 变量的声明
		if tsmorphgo.IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "users" {
			parent := node.GetParent()
			if parent != nil && tsmorphgo.IsVariableDeclaration(*parent) {
				fmt.Printf("\n找到 'users' 变量声明:\n")
				fmt.Printf("  - 位置: 行 %d\n", node.GetStartLineNumber())

				// 查找最近的函数声明祖先
				if funcDecl, found := node.GetFirstAncestorByKind(tsmorphgo.KindFunctionDeclaration); found {
					if name, ok := tsmorphgo.GetFunctionDeclarationNameNode(*funcDecl); ok {
						fmt.Printf("  - 它位于函数 '%s' 内部\n", name.GetText())
					}
				}

				// 查找最近的JSX元素祖先
				if _, found := node.GetFirstAncestorByKind(tsmorphgo.KindJsxElement); found {
					fmt.Printf("  - 它位于一个JSX元素内部\n")
				}
				return // 只演示一次
			}
		}
	})

	// 示例4: 条件遍历和提前终止
	fmt.Println("\n⚡ 示例4: 条件遍历和提前终止")
	// 在 ForEachDescendant 的回调中，可以通过返回非nil的error来提前终止遍历。
	// 这里我们通过一个闭包变量来模拟这个过程。
	var jsxAttribute *tsmorphgo.Node
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if jsxAttribute != nil {
			return // 变量非nil，说明已找到，终止后续遍历
		}
		// 通过 Kind 直接判断节点是否为 JSX 属性
		if node.Kind == tsmorphgo.KindJsxAttribute {
			text := strings.TrimSpace(node.GetText())
			fmt.Printf("找到第一个JSX属性 (行 %d): %s\n", node.GetStartLineNumber(), text)
			jsxAttribute = &node
		}
	})

	if jsxAttribute != nil {
		// GetFirstChild 根据回调函数查找第一个匹配的子节点。
		// 对应 ts-morph 的 `node.getFirstChild(predicate)`。
		if name, ok := tsmorphgo.GetFirstChild(*jsxAttribute, func(n tsmorphgo.Node) bool { return tsmorphgo.IsIdentifier(n) }); ok {
			fmt.Printf("  - 属性名: %s\n", name.GetText())
		}
	} else {
		fmt.Println("未找到JSX属性")
	}

	// 示例5: 深度分析React组件结构
	fmt.Println("\n⚛️ 示例5: 分析React组件的返回值")
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		// 找到名为 "App" 的函数
		if tsmorphgo.IsFunctionDeclaration(node) {
			if name, ok := tsmorphgo.GetFunctionDeclarationNameNode(node); ok && name.GetText() == "App" {
				fmt.Println("分析 'App' 组件的 return 语句:")
				node.ForEachDescendant(func(descendant tsmorphgo.Node) {
					// 找到 return 语句
					if descendant.Kind == tsmorphgo.KindReturnStatement {
						fmt.Printf("  - 找到 return 语句 (行 %d)\n", descendant.GetStartLineNumber())
						// 进一步分析 return 的内容
						descendant.ForEachDescendant(func(returnChild tsmorphgo.Node) {
							if returnChild.Kind == tsmorphgo.KindJsxSelfClosingElement {
								fmt.Printf("    - 返回了自闭合JSX元素: %s\n", returnChild.GetText())
							}
						})
					}
				})
			}
		}
	})

	fmt.Println("\n✅ 节点导航示例完成!")
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}