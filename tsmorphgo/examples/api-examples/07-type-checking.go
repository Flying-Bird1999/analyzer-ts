//go:build example07

package main

import (
	"fmt"
	"os"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run 08-type-checking.go <TypeScript项目路径>")
		os.Exit(1)
	}

	projectPath := os.Args[1]

	fmt.Println("🔍 类型检查示例 - 节点类型识别和转换")
	fmt.Println("==================================================")

	// 创建项目
	config := tsmorphgo.ProjectConfig{
		RootPath:         projectPath,
		IgnorePatterns:   []string{"node_modules", "dist", "build", ".git"},
		TargetExtensions: []string{".ts", ".tsx"},
	}
	project := tsmorphgo.NewProject(config)

	// 统计所有类型检查函数的使用
	typeCheckStats := make(map[string]int)

	// 分析各种节点类型
	var (
		variables     []TypeCheckResult
		functions     []TypeCheckResult
		interfaces    []TypeCheckResult
		typeAliases   []TypeCheckResult
		enums         []TypeCheckResult
		classes       []TypeCheckResult
		identifiers   []TypeCheckResult
	)

	for _, sf := range project.GetSourceFiles() {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			// 使用各种类型检查函数
			if tsmorphgo.IsIdentifier(node) {
				typeCheckStats["IsIdentifier"]++
				identifiers = append(identifiers, TypeCheckResult{
					Kind:      node.Kind,
					Text:      node.GetText(),
					File:      sf.GetFilePath(),
					Line:      node.GetStartLineNumber(),
					CheckType: "identifier",
				})
			}

			if tsmorphgo.IsCallExpression(node) {
				typeCheckStats["IsCallExpression"]++
			}

			if tsmorphgo.IsPropertyAccessExpression(node) {
				typeCheckStats["IsPropertyAccessExpression"]++
			}

			if tsmorphgo.IsPropertyAssignment(node) {
				typeCheckStats["IsPropertyAssignment"]++
			}

			if tsmorphgo.IsPropertyDeclaration(node) {
				typeCheckStats["IsPropertyDeclaration"]++
			}

			if tsmorphgo.IsObjectLiteralExpression(node) {
				typeCheckStats["IsObjectLiteralExpression"]++
			}

			if tsmorphgo.IsBinaryExpression(node) {
				typeCheckStats["IsBinaryExpression"]++
			}

			// 使用声明类型检查函数
			if tsmorphgo.IsVariableDeclaration(node) {
				typeCheckStats["IsVariableDeclaration"]++
				result := analyzeVariableDeclaration(node, sf)
				variables = append(variables, *result)
			}

			if tsmorphgo.IsFunctionDeclaration(node) {
				typeCheckStats["IsFunctionDeclaration"]++
				result := analyzeFunctionDeclaration(node, sf)
				functions = append(functions, *result)
			}

			if tsmorphgo.IsInterfaceDeclaration(node) {
				typeCheckStats["IsInterfaceDeclaration"]++
				result := analyzeInterfaceDeclaration(node, sf)
				interfaces = append(interfaces, *result)
			}

			if tsmorphgo.IsTypeAliasDeclaration(node) {
				typeCheckStats["IsTypeAliasDeclaration"]++
				result := analyzeTypeAliasDeclaration(node, sf)
				typeAliases = append(typeAliases, *result)
			}

			if tsmorphgo.IsEnumDeclaration(node) {
				typeCheckStats["IsEnumDeclaration"]++
				result := analyzeEnumDeclaration(node, sf)
				enums = append(enums, *result)
			}

			if tsmorphgo.IsClassDeclaration(node) {
				typeCheckStats["IsClassDeclaration"]++
				result := analyzeClassDeclaration(node, sf)
				classes = append(classes, *result)
			}

			if tsmorphgo.IsImportClause(node) {
				typeCheckStats["IsImportClause"]++
			}
		})
	}

	// 打印类型检查统计
	fmt.Println("📊 类型检查函数使用统计:")
	for checkFunc, count := range typeCheckStats {
		fmt.Printf("  %s: %d\n", checkFunc, count)
	}

	// 显示各种声明分析结果
	fmt.Println("\n📦 变量声明分析 (前 3 个):")
	for i, result := range variables {
		if i >= 3 {
			break
		}
		fmt.Printf("  %d. %s: %s (%s:%d)\n", i+1, result.Name, result.Type, result.File, result.Line)
	}

	fmt.Println("\n⚡ 函数声明分析 (前 3 个):")
	for i, result := range functions {
		if i >= 3 {
			break
		}
		fmt.Printf("  %d. %s (%s:%d)\n", i+1, result.Name, result.File, result.Line)
	}

	fmt.Println("\n🔷 接口声明分析 (前 3 个):")
	for i, result := range interfaces {
		if i >= 3 {
			break
		}
		fmt.Printf("  %d. %s (%s:%d)\n", i+1, result.Name, result.File, result.Line)
	}

	fmt.Println("\n🏷️ 类型别名分析 (前 3 个):")
	for i, result := range typeAliases {
		if i >= 3 {
			break
		}
		fmt.Printf("  %d. %s = %s (%s:%d)\n", i+1, result.Name, result.Type, result.File, result.Line)
	}

	fmt.Println("\n🏗️ 类声明分析 (前 3 个):")
	for i, result := range classes {
		if i >= 3 {
			break
		}
		fmt.Printf("  %d. %s (%s:%d)\n", i+1, result.Name, result.File, result.Line)
	}

	// 测试类型转换函数
	fmt.Println("\n🔄 类型转换函数测试:")
	testTypeConversions(project)

	fmt.Println("\n✅ 类型检查分析完成！")
}

// TypeCheckResult 类型检查结果
type TypeCheckResult struct {
	Kind      ast.Kind `json:"kind"`
	Text      string   `json:"text"`
	File      string   `json:"file"`
	Line      int      `json:"line"`
	CheckType string   `json:"checkType"`

	// 特定字段
	Name      string `json:"name,omitempty"`
	Type      string `json:"type,omitempty"`
	Exported  bool   `json:"exported,omitempty"`
}

// analyzeVariableDeclaration 分析变量声明
func analyzeVariableDeclaration(node tsmorphgo.Node, sf *tsmorphgo.SourceFile) *TypeCheckResult {
	result := &TypeCheckResult{
		Kind:      node.Kind,
		Text:      node.GetText(),
		File:      sf.GetFilePath(),
		Line:      node.GetStartLineNumber(),
		CheckType: "variable",
	}

	// 获取变量名
	if name, ok := tsmorphgo.GetVariableName(node); ok {
		result.Name = name
	}

	// 获取变量类型
	if nameNode, ok := tsmorphgo.GetVariableDeclarationNameNode(node); ok {
		result.Type = nameNode.GetText()
	}

	return result
}

// analyzeFunctionDeclaration 分析函数声明
func analyzeFunctionDeclaration(node tsmorphgo.Node, sf *tsmorphgo.SourceFile) *TypeCheckResult {
	result := &TypeCheckResult{
		Kind:      node.Kind,
		Text:      node.GetText(),
		File:      sf.GetFilePath(),
		Line:      node.GetStartLineNumber(),
		CheckType: "function",
	}

	// 获取函数名
	if nameNode, ok := tsmorphgo.GetFunctionDeclarationNameNode(node); ok {
		result.Name = nameNode.GetText()
	}

	return result
}

// analyzeInterfaceDeclaration 分析接口声明
func analyzeInterfaceDeclaration(node tsmorphgo.Node, sf *tsmorphgo.SourceFile) *TypeCheckResult {
	result := &TypeCheckResult{
		Kind:      node.Kind,
		Text:      node.GetText(),
		File:      sf.GetFilePath(),
		Line:      node.GetStartLineNumber(),
		CheckType: "interface",
	}

	// 获取接口名
	if name, ok := tsmorphgo.GetVariableName(node); ok {
		result.Name = name
	}

	return result
}

// analyzeTypeAliasDeclaration 分析类型别名声明
func analyzeTypeAliasDeclaration(node tsmorphgo.Node, sf *tsmorphgo.SourceFile) *TypeCheckResult {
	result := &TypeCheckResult{
		Kind:      node.Kind,
		Text:      node.GetText(),
		File:      sf.GetFilePath(),
		Line:      node.GetStartLineNumber(),
		CheckType: "typeAlias",
	}

	// 获取类型别名名称
	if name, ok := tsmorphgo.GetVariableName(node); ok {
		result.Name = name
	}

	// 获取类型
	if typeDecl, ok := tsmorphgo.AsTypeAliasDeclaration(node); ok {
		result.Type = typeDecl.Raw
	}

	return result
}

// analyzeEnumDeclaration 分析枚举声明
func analyzeEnumDeclaration(node tsmorphgo.Node, sf *tsmorphgo.SourceFile) *TypeCheckResult {
	result := &TypeCheckResult{
		Kind:      node.Kind,
		Text:      node.GetText(),
		File:      sf.GetFilePath(),
		Line:      node.GetStartLineNumber(),
		CheckType: "enum",
	}

	// 获取枚举名
	if name, ok := tsmorphgo.GetVariableName(node); ok {
		result.Name = name
	}

	return result
}

// analyzeClassDeclaration 分析类声明
func analyzeClassDeclaration(node tsmorphgo.Node, sf *tsmorphgo.SourceFile) *TypeCheckResult {
	result := &TypeCheckResult{
		Kind:      node.Kind,
		Text:      node.GetText(),
		File:      sf.GetFilePath(),
		Line:      node.GetStartLineNumber(),
		CheckType: "class",
	}

	// 获取类名
	if name, ok := tsmorphgo.GetVariableName(node); ok {
		result.Name = name
	}

	return result
}

// testTypeConversions 测试类型转换函数
func testTypeConversions(project *tsmorphgo.Project) {
	fmt.Println("  测试各种 AsXXX 转换函数:")

	conversionCount := make(map[string]int)

	for _, sf := range project.GetSourceFiles() {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			switch node.Kind {
			case ast.KindImportDeclaration:
				if _, ok := tsmorphgo.AsImportDeclaration(node); ok {
					conversionCount["AsImportDeclaration"]++
				}

			case ast.KindVariableDeclaration:
				if _, ok := tsmorphgo.AsVariableDeclaration(node); ok {
					conversionCount["AsVariableDeclaration"]++
				}

			case ast.KindFunctionDeclaration:
				if _, ok := tsmorphgo.AsFunctionDeclaration(node); ok {
					conversionCount["AsFunctionDeclaration"]++
				}

			case ast.KindInterfaceDeclaration:
				if _, ok := tsmorphgo.AsInterfaceDeclaration(node); ok {
					conversionCount["AsInterfaceDeclaration"]++
				}

			case ast.KindTypeAliasDeclaration:
				if _, ok := tsmorphgo.AsTypeAliasDeclaration(node); ok {
					conversionCount["AsTypeAliasDeclaration"]++
				}

			case ast.KindEnumDeclaration:
				if _, ok := tsmorphgo.AsEnumDeclaration(node); ok {
					conversionCount["AsEnumDeclaration"]++
				}
			}
		})
	}

	// 显示转换统计
	for convFunc, count := range conversionCount {
		if count > 0 {
			fmt.Printf("    %s: %d 次转换成功\n", convFunc, count)
		}
	}
}