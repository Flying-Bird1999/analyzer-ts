//go:build example06

package main

import (
	"fmt"
	"os"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run 07-expression-analysis.go <TypeScript项目路径>")
		os.Exit(1)
	}

	projectPath := os.Args[1]

	fmt.Println("🔍 表达式分析示例 - 代码表达式解析")
	fmt.Println("==================================================")

	// 创建项目
	config := tsmorphgo.ProjectConfig{
		RootPath:         projectPath,
		IgnorePatterns:   []string{"node_modules", "dist", "build", ".git"},
		TargetExtensions: []string{".ts", ".tsx"},
	}
	project := tsmorphgo.NewProject(config)

	// 统计各种表达式类型
	stats := make(map[ast.Kind]int)

	// 分析各种表达式
	var (
		callExpressions       []ExpressionInfo
		propertyAccessExprs  []ExpressionInfo
		binaryExpressions    []ExpressionInfo
		objectLiterals       []ExpressionInfo
		identifiers          []ExpressionInfo
		stringLiterals      []ExpressionInfo
		numericLiterals     []ExpressionInfo
	)

	for _, sf := range project.GetSourceFiles() {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			stats[node.Kind]++

			switch node.Kind {
			case ast.KindCallExpression:
				info := analyzeCallExpression(node)
				callExpressions = append(callExpressions, *info)

			case ast.KindPropertyAccessExpression:
				info := analyzePropertyAccessExpression(node)
				propertyAccessExprs = append(propertyAccessExprs, *info)

			case ast.KindBinaryExpression:
				info := analyzeBinaryExpression(node)
				binaryExpressions = append(binaryExpressions, *info)

			case ast.KindObjectLiteralExpression:
				info := analyzeObjectLiteral(node)
				objectLiterals = append(objectLiterals, *info)

			case ast.KindIdentifier:
				identifiers = append(identifiers, ExpressionInfo{
					Kind:      node.Kind,
					Text:      node.GetText(),
					File:      sf.GetFilePath(),
					Line:      node.GetStartLineNumber(),
				})

			case ast.KindStringLiteral:
				stringLiterals = append(stringLiterals, ExpressionInfo{
					Kind:      node.Kind,
					Text:      node.GetText(),
					File:      sf.GetFilePath(),
					Line:      node.GetStartLineNumber(),
				})

			case ast.KindNumericLiteral:
				numericLiterals = append(numericLiterals, ExpressionInfo{
					Kind:      node.Kind,
					Text:      node.GetText(),
					File:      sf.GetFilePath(),
					Line:      node.GetStartLineNumber(),
				})
			}
		})
	}

	// 打印统计信息
	fmt.Println("📊 表达式统计:")
	fmt.Printf("  调用表达式: %d\n", len(callExpressions))
	fmt.Printf("  属性访问表达式: %d\n", len(propertyAccessExprs))
	fmt.Printf("  二元表达式: %d\n", len(binaryExpressions))
	fmt.Printf("  对象字面量: %d\n", len(objectLiterals))
	fmt.Printf("  标识符: %d\n", len(identifiers))
	fmt.Printf("  字符串字面量: %d\n", len(stringLiterals))
	fmt.Printf("  数字字面量: %d\n", len(numericLiterals))

	// 显示调用表达式分析
	fmt.Println("\n📞 调用表达式分析 (前 5 个):")
	for i, expr := range callExpressions {
		if i >= 5 {
			break
		}
		fmt.Printf("  %d. %s -> %s (行: %d)\n", i+1, expr.Function, expr.Details, expr.Line)
	}

	// 显示属性访问分析
	fmt.Println("\n🔗 属性访问表达式分析 (前 5 个):")
	for i, expr := range propertyAccessExprs {
		if i >= 5 {
			break
		}
		fmt.Printf("  %d. %s.%s (行: %d)\n", i+1, expr.Object, expr.Property, expr.Line)
	}

	// 显示二元表达式分析
	fmt.Println("\n⚖️  二元表达式分析 (前 5 个):")
	for i, expr := range binaryExpressions {
		if i >= 5 {
			break
		}
		fmt.Printf("  %d. %s %s %s (行: %d)\n", i+1, expr.Left, expr.Operator, expr.Right, expr.Line)
	}

	// 显示各种 Kind 统计
	fmt.Println("\n🏷️  AST 节点种类统计:")
	for kind, count := range stats {
		fmt.Printf("  %v: %d\n", kind, count)
	}

	// 分析模式
	fmt.Println("\n🔍 表达式模式分析:")
	analyzeExpressionPatterns(callExpressions, propertyAccessExprs, binaryExpressions)

	fmt.Println("\n✅ 表达式分析完成！")
}

// ExpressionInfo 表达式信息
type ExpressionInfo struct {
	Kind      ast.Kind `json:"kind"`
	Text      string    `json:"text"`
	File      string    `json:"file"`
	Line      int       `json:"line"`

	// 特定字段
	Function  string `json:"function,omitempty"`
	Object    string `json:"object,omitempty"`
	Property  string `json:"property,omitempty"`
	Left      string `json:"left,omitempty"`
	Right     string `json:"right,omitempty"`
	Operator  string `json:"operator,omitempty"`
	Details   string `json:"details,omitempty"`
}

// analyzeCallExpression 分析调用表达式
func analyzeCallExpression(node tsmorphgo.Node) *ExpressionInfo {
	info := &ExpressionInfo{
		Kind:   node.Kind,
		Text:   node.GetText(),
		File:   node.GetSourceFile().GetFilePath(),
		Line:   node.GetStartLineNumber(),
	}

	// 获取调用的函数
	if expr, ok := tsmorphgo.GetCallExpressionExpression(node); ok {
		info.Function = expr.GetText()
		info.Details = fmt.Sprintf("函数调用: %s", info.Function)
	}

	return info
}

// analyzePropertyAccessExpression 分析属性访问表达式
func analyzePropertyAccessExpression(node tsmorphgo.Node) *ExpressionInfo {
	info := &ExpressionInfo{
		Kind:   node.Kind,
		Text:   node.GetText(),
		File:   node.GetSourceFile().GetFilePath(),
		Line:   node.GetStartLineNumber(),
	}

	// 获取属性名称
	if name, ok := tsmorphgo.GetPropertyAccessName(node); ok {
		info.Property = name
	}

	// 获取访问对象
	if obj, ok := tsmorphgo.GetPropertyAccessExpression(node); ok {
		info.Object = obj.GetText()
	}

	return info
}

// analyzeBinaryExpression 分析二元表达式
func analyzeBinaryExpression(node tsmorphgo.Node) *ExpressionInfo {
	info := &ExpressionInfo{
		Kind:   node.Kind,
		Text:   node.GetText(),
		File:   node.GetSourceFile().GetFilePath(),
		Line:   node.GetStartLineNumber(),
	}

	// 获取左右操作数
	if left, ok := tsmorphgo.GetBinaryExpressionLeft(node); ok {
		info.Left = left.GetText()
	}

	if right, ok := tsmorphgo.GetBinaryExpressionRight(node); ok {
		info.Right = right.GetText()
	}

	// 获取操作符
	if op, ok := tsmorphgo.GetBinaryExpressionOperatorToken(node); ok {
		info.Operator = op.GetText()
	}

	return info
}

// analyzeObjectLiteral 分析对象字面量
func analyzeObjectLiteral(node tsmorphgo.Node) *ExpressionInfo {
	info := &ExpressionInfo{
		Kind:   node.Kind,
		Text:   node.GetText(),
		File:   node.GetSourceFile().GetFilePath(),
		Line:   node.GetStartLineNumber(),
	}

	// 计算属性数量
	propertyCount := 0
	node.ForEachChild(func(child *ast.Node) bool {
		if child.Kind == ast.KindPropertyAssignment {
			propertyCount++
		}
		return true
	})

	info.Details = fmt.Sprintf("包含 %d 个属性", propertyCount)

	return info
}

// analyzeExpressionPatterns 分析表达式模式
func analyzeExpressionPatterns(calls, propertyAccesses, binaries []ExpressionInfo) {
	// 分析函数调用模式
	functionCalls := make(map[string]int)
	for _, call := range calls {
		functionCalls[call.Function]++
	}

	// 分析属性访问模式
	propertyChains := make(map[string]int)
	for _, access := range propertyAccesses {
		if access.Object != "" && access.Property != "" {
			chain := fmt.Sprintf("%s.%s", access.Object, access.Property)
			propertyChains[chain]++
		}
	}

	// 分析操作符使用
	operators := make(map[string]int)
	for _, binary := range binaries {
		operators[binary.Operator]++
	}

	// 显示最常用的函数调用
	fmt.Println("  最常用的函数调用:")
	count := 0
	for funcName, callCount := range functionCalls {
		if count >= 3 {
			break
		}
		if callCount > 1 {
			fmt.Printf("    %s: %d 次\n", funcName, callCount)
			count++
		}
	}

	// 显示最常用的属性访问
	fmt.Println("  最常用的属性访问:")
	count = 0
	for chain, accessCount := range propertyChains {
		if count >= 3 {
			break
		}
		if accessCount > 1 {
			fmt.Printf("    %s: %d 次\n", chain, accessCount)
			count++
		}
	}

	// 显示最常用的操作符
	fmt.Println("  最常用的二元操作符:")
	count = 0
	for op, opCount := range operators {
		if count >= 3 {
			break
		}
		if opCount > 1 {
			fmt.Printf("    %s: %d 次\n", op, opCount)
			count++
		}
	}
}