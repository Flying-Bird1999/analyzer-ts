//go:build example05

package main

import (
	"fmt"
	"os"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run 06-node-navigation.go <TypeScript项目路径>")
		os.Exit(1)
	}

	projectPath := os.Args[1]

	fmt.Println("🔍 节点导航示例 - AST 树遍历和导航")
	fmt.Println("==================================================")

	// 创建项目
	config := tsmorphgo.ProjectConfig{
		RootPath:         projectPath,
		IgnorePatterns:   []string{"node_modules", "dist", "build", ".git"},
		TargetExtensions: []string{".ts", ".tsx"},
	}
	project := tsmorphgo.NewProject(config)

	// 查找第一个函数声明
	var foundFunction bool
	var firstFunction tsmorphgo.Node
	for _, sf := range project.GetSourceFiles() {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			if !foundFunction && (node.Kind == ast.KindFunctionDeclaration || node.Kind == ast.KindInterfaceDeclaration) {
				foundFunction = true
				firstFunction = node
			}
		})
		if foundFunction {
			break
		}
	}

	if !foundFunction {
		fmt.Println("❌ 未找到函数声明")
		return
	}

	fmt.Printf("✅ 找到函数声明: %s (行: %d)\n", firstFunction.GetText(), firstFunction.GetStartLineNumber())

	// 1. 节点信息
	fmt.Println("\n📋 节点基本信息:")
	fmt.Printf("  节点类型: %v\n", firstFunction.Kind)
	fmt.Printf("  节点文本: %s\n", firstFunction.GetText())
	fmt.Printf("  起始行号: %d\n", firstFunction.GetStartLineNumber())
	fmt.Printf("  起始位置: %d\n", firstFunction.GetStart())
	fmt.Printf("  所属文件: %s\n", firstFunction.GetSourceFile().GetFilePath())

	// 2. 父子节点导航
	fmt.Println("\n🔗 父子节点导航:")
	parent := firstFunction.GetParent()
	if parent != nil {
		fmt.Printf("  父节点类型: %v\n", parent.Kind)
		fmt.Printf("  父节点文本: %s\n", parent.GetText())
	}

	// 3. 祖先节点遍历
	fmt.Println("\n🌳 祖先节点遍历:")
	ancestors := firstFunction.GetAncestors()
	fmt.Printf("  祖先节点数: %d\n", len(ancestors))
	for i, ancestor := range ancestors {
		if i >= 3 { // 只显示前3个
			break
		}
		fmt.Printf("  [%d] %v: %s\n", i+1, ancestor.Kind, ancestor.GetText())
	}

	// 4. 特定类型祖先查找
	fmt.Println("\n🔍 特定类型祖先查找:")
	if interfaceNode, ok := firstFunction.GetFirstAncestorByKind(ast.KindInterfaceDeclaration); ok {
		fmt.Printf("  找到接口祖先: %s\n", interfaceNode.GetText())
	} else {
		fmt.Println("  未找到接口祖先")
	}

	if sourceFileNode, ok := firstFunction.GetFirstAncestorByKind(ast.KindSourceFile); ok {
		fmt.Printf("  找到源文件祖先: %s\n", sourceFileNode.GetSourceFile().GetFilePath())
	}

	// 5. 条件查找子节点
	fmt.Println("\n🔍 条件查找子节点:")
	foundIdentifier := false
	firstFunction.ForEachChild(func(child *ast.Node) bool {
		if !foundIdentifier && child.Kind == ast.KindIdentifier {
			fmt.Printf("  找到标识符: %s\n", child.Text())
			foundIdentifier = true
			return false // 停止遍历
		}
		return true
	})

	// 6. 使用 GetFirstChild 查找
	fmt.Println("\n🎯 使用 GetFirstChild 查找:")
	if identifierNode, ok := tsmorphgo.GetFirstChild(firstFunction, func(n tsmorphgo.Node) bool {
		return n.Kind == ast.KindIdentifier
	}); ok {
		fmt.Printf("  找到标识符节点: %s\n", identifierNode.GetText())
	}

	// 7. 深度分析
	fmt.Println("\n📊 节点深度分析:")
	depth := calculateNodeDepth(firstFunction)
	fmt.Printf("  节点深度: %d\n", depth)

	// 8. QuickInfo 测试
	fmt.Println("\n💡 QuickInfo 测试:")
	if quickInfo, err := firstFunction.GetQuickInfo(); err == nil && quickInfo != nil {
		fmt.Printf("  QuickInfo 类型: %s\n", quickInfo.TypeText)
		if quickInfo.DisplayParts != nil {
			fmt.Printf("  显示内容: %v\n", quickInfo.DisplayParts)
		}
	} else {
		fmt.Printf("  无法获取 QuickInfo: %v\n", err)
	}

	// 9. 查找节点所在文件的符号
	fmt.Println("\n🔣 文件符号分析:")
	sf := firstFunction.GetSourceFile()
	var symbols []string
	sf.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.Kind == ast.KindFunctionDeclaration || node.Kind == ast.KindInterfaceDeclaration {
			symbols = append(symbols, node.GetText())
		}
	})
	fmt.Printf("  文件中的声明: %v\n", symbols)

	fmt.Println("\n✅ 节点导航分析完成！")
}

// calculateNodeDepth 计算节点深度
func calculateNodeDepth(node tsmorphgo.Node) int {
	depth := 0
	ancestors := node.GetAncestors()

	// 计算有效祖先深度
	for _, ancestor := range ancestors {
		if ancestor.Kind != ast.KindSourceFile {
			depth++
		}
	}

	return depth
}