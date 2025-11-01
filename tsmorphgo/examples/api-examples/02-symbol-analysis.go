//go:build example02

package main

import (
	"fmt"
	"os"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run 02-symbol-analysis.go <TypeScript项目路径>")
		os.Exit(1)
	}

	projectPath := os.Args[1]

	fmt.Println("🔣 符号分析示例 - 查找定义和引用")
	fmt.Println("==================================================")

	// 创建项目
	config := tsmorphgo.ProjectConfig{
		RootPath:         projectPath,
		IgnorePatterns:   []string{"node_modules", "dist", "build", ".git"},
		TargetExtensions: []string{".ts", ".tsx"},
	}
	project := tsmorphgo.NewProject(config)

	// 收集所有符号
	symbols := collectAllSymbols(project)
	fmt.Printf("✅ 收集到 %d 个符号\n", len(symbols))

	// 分析前 5 个符号的引用情况
	for i, symbolInfo := range symbols {
		if i >= 5 {
			break
		}

		fmt.Printf("\n🔍 符号 %d: %s\n", i+1, symbolInfo.Name)
		fmt.Printf("   类型: %s\n", symbolInfo.Kind)
		fmt.Printf("   位置: %s:%d\n", symbolInfo.File, symbolInfo.Line)
		fmt.Printf("   可导出: %t\n", symbolInfo.IsExported)

		// 查找引用
		if refs, err := symbolInfo.Symbol.FindReferences(); err == nil {
			fmt.Printf("   引用数: %d\n", len(refs))
			for j, ref := range refs {
				if j >= 3 { // 只显示前 3 个引用
					break
				}
				fmt.Printf("     -> %s:%d\n", ref.GetSourceFile().GetFilePath(), ref.GetStartLineNumber())
			}
		}
	}

	fmt.Println("\n✅ 符号分析完成！")
}

// SymbolInfo 符号信息
type SymbolInfo struct {
	Name      string
	Kind      string
	File      string
	Line      int
	IsExported bool
	Symbol    *tsmorphgo.Symbol
}

// collectAllSymbols 收集项目中的所有符号
func collectAllSymbols(project *tsmorphgo.Project) []SymbolInfo {
	var symbols []SymbolInfo

	for _, sf := range project.GetSourceFiles() {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			// 只处理特定类型的节点
			if isSymbolNode(node) {
				if symbol, ok := tsmorphgo.GetSymbol(node); ok {
					info := SymbolInfo{
						Name:      symbol.GetName(),
						Kind:      getKindName(node.Kind),
						File:      sf.GetFilePath(),
						Line:      node.GetStartLineNumber(),
						IsExported: symbol.IsExported(),
						Symbol:    symbol,
					}
					symbols = append(symbols, info)
				}
			}
		})
	}

	return symbols
}

// isSymbolNode 判断节点是否包含符号
func isSymbolNode(node tsmorphgo.Node) bool {
	switch node.Kind {
	case ast.KindInterfaceDeclaration,
		ast.KindTypeAliasDeclaration,
		ast.KindFunctionDeclaration,
		ast.KindClassDeclaration,
		ast.KindVariableDeclaration,
		ast.KindEnumDeclaration:
		return true
	default:
		return false
	}
}

// getKindName 获取节点类型名称
func getKindName(kind ast.Kind) string {
	switch kind {
	case ast.KindInterfaceDeclaration:
		return "interface"
	case ast.KindTypeAliasDeclaration:
		return "type"
	case ast.KindFunctionDeclaration:
		return "function"
	case ast.KindClassDeclaration:
		return "class"
	case ast.KindVariableDeclaration:
		return "variable"
	case ast.KindEnumDeclaration:
		return "enum"
	default:
		return "unknown"
	}
}