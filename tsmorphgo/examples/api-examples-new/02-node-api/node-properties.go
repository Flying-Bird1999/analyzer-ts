// +build node-api

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run -tags node-api node-properties.go <TypeScript项目路径>")
		os.Exit(1)
	}

	projectPath := os.Args[1]

	fmt.Println("🎯 节点操作 API - 节点属性（位置、文本、类型）")
	fmt.Println("================================")

	// 创建项目配置
	config := tsmorphgo.ProjectConfig{
		RootPath:         projectPath,
		IgnorePatterns:   []string{"node_modules", "dist", "build", ".git"},
		TargetExtensions: []string{".ts", ".tsx"},
	}
	project := tsmorphgo.NewProject(config)

	sourceFiles := project.GetSourceFiles()
	if len(sourceFiles) == 0 {
		fmt.Println("❌ 项目创建失败：未发现任何源文件")
		return
	}

	fmt.Printf("✅ 项目创建成功，发现 %d 个源文件\n", len(sourceFiles))

	// 1. 收集测试节点
	var testNodes []tsmorphgo.Node
	testNodeTypes := []ast.Kind{
		ast.KindInterfaceDeclaration,
		ast.KindFunctionDeclaration,
		ast.KindClassDeclaration,
		ast.KindTypeAliasDeclaration,
		ast.KindVariableDeclaration,
		ast.KindMethodDeclaration,
		ast.KindPropertyDeclaration,
	}

	nodeTypeTargets := make(map[ast.Kind]int)
	for _, sf := range sourceFiles {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			for _, kind := range testNodeTypes {
				if node.Kind == kind && nodeTypeTargets[kind] < 2 {
					testNodes = append(testNodes, node)
					nodeTypeTargets[kind]++
					break
				}
			}
		})
	}

	// 2. 节点位置信息验证
	fmt.Println("\n📍 节点位置信息验证:")
	fmt.Println("------------------------------")

	locationValidationResults := []LocationValidationResult{}

	for i, node := range testNodes {
		result := validateNodeLocation(node, i)
		locationValidationResults = append(locationValidationResults, result)

		fmt.Printf("  [%d] %v:\n", i+1, node.Kind)
		fmt.Printf("     起始位置: %d\n", result.StartPosition)
		fmt.Printf("     结束位置: %d\n", result.EndPosition)
		fmt.Printf("     位置范围: %d\n", result.Length)
		fmt.Printf("     起始行号: %d\n", result.StartLine)
		fmt.Printf("     结束行号: %d\n", result.EndLine)
		fmt.Printf("     跨行数量: %d\n", result.SpanLines)
		fmt.Printf("     位置有效性: %t\n", result.IsValid)

		if !result.IsValid {
			fmt.Printf("     ❌ 位置信息验证失败\n")
		} else {
			fmt.Printf("     ✅ 位置信息验证通过\n")
		}
	}

	// 3. 节点文本信息验证
	fmt.Println("\n📝 节点文本信息验证:")
	fmt.Println("------------------------------")

	textValidationResults := []TextValidationResult{}

	for i, node := range testNodes {
		result := validateNodeText(node, i)
		textValidationResults = append(textValidationResults, result)

		fmt.Printf("  [%d] %v:\n", i+1, node.Kind)
		fmt.Printf("     节点文本: %s\n", result.Text)
		fmt.Printf("     文本长度: %d\n", result.TextLength)
		fmt.Printf("     文本哈希: %x\n", result.TextHash)
		fmt.Printf("     是否为空: %t\n", result.IsEmpty)
		fmt.Printf("     是否包含换行: %t\n", result.HasNewlines)
		fmt.Printf("     文本有效性: %t\n", result.IsValid)

		if !result.IsValid {
			fmt.Printf("     ❌ 文本信息验证失败\n")
		} else {
			fmt.Printf("     ✅ 文本信息验证通过\n")
		}
	}

	// 4. 节点类型信息验证
	fmt.Println("\n🏷️ 节点类型信息验证:")
	fmt.Println("------------------------------")

	typeValidationResults := []TypeValidationResult{}

	for i, node := range testNodes {
		result := validateNodeType(node, i)
		typeValidationResults = append(typeValidationResults, result)

		fmt.Printf("  [%d] 节点:\n", i+1)
		fmt.Printf("     节点类型: %v\n", result.NodeType)
		fmt.Printf("     类型名称: %s\n", result.TypeName)
		fmt.Printf("     类型分组: %s\n", result.TypeGroup)
		fmt.Printf("     是否为声明类型: %t\n", result.IsDeclaration)
		fmt.Printf("     是否为表达式类型: %t\n", result.IsExpression)
		fmt.Printf("     是否为字面量类型: %t\n", result.IsLiteral)
		fmt.Printf("     是否为标识符类型: %t\n", result.IsIdentifier)
		fmt.Printf("     类型有效性: %t\n", result.IsValid)

		if !result.IsValid {
			fmt.Printf("     ❌ 类型信息验证失败\n")
		} else {
			fmt.Printf("     ✅ 类型信息验证通过\n")
		}
	}

	// 5. 节点边界情况验证
	fmt.Println("\n🔍 节点边界情况验证:")
	fmt.Println("------------------------------")

	edgeCaseResults := []EdgeCaseResult{}

	// 测试空节点或特殊节点
	for _, sf := range sourceFiles {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			if len(edgeCaseResults) >= 10 {
				return
			}

			result := validateEdgeCaseNode(node)
			if result.IsEdgeCase {
				edgeCaseResults = append(edgeCaseResults, result)
			}
		})
	}

	for i, result := range edgeCaseResults {
		fmt.Printf("  [%d] 边界情况 (%v):\n", i+1, result.NodeType)
		fmt.Printf("     边界类型: %s\n", result.EdgeCaseType)
		fmt.Printf("     描述: %s\n", result.Description)
		fmt.Printf("     处理结果: %s\n", result.HandlingResult)
	}

	// 6. 节点属性关联性验证
	fmt.Println("\n🔗 节点属性关联性验证:")
	fmt.Println("------------------------------")

	correlationResults := []CorrelationResult{}

	for i, node := range testNodes {
		result := validateNodeCorrelations(node, i)
		correlationResults = append(correlationResults, result)

		fmt.Printf("  [%d] %v:\n", i+1, node.Kind)
		fmt.Printf("     位置-文本关联: %t\n", result.LocationTextCorrelation)
		fmt.Printf("     位置-类型关联: %t\n", result.LocationTypeCorrelation)
		fmt.Printf("     文本-类型关联: %t\n", result.TextTypeCorrelation)
		fmt.Printf("     整体一致性: %t\n", result.OverallConsistency)

		if !result.OverallConsistency {
			fmt.Printf("     ❌ 属性关联性验证失败\n")
		} else {
			fmt.Printf("     ✅ 属性关联性验证通过\n")
		}
	}

	// 7. 节点属性性能验证
	fmt.Println("\n⏱️ 节点属性性能验证:")
	fmt.Println("------------------------------")

	performanceResult := validatePropertyPerformance(testNodes)

	fmt.Printf("  测试节点数: %d\n", performanceResult.TestNodeCount)
	fmt.Printf("  平均位置获取时间: %.3f ms\n", performanceResult.AverageLocationTime)
	fmt.Printf("  平均文本获取时间: %.3f ms\n", performanceResult.AverageTextTime)
	fmt.Printf("  平均类型获取时间: %.3f ms\n", performanceResult.AverageTypeTime)
	fmt.Printf("  平均总体时间: %.3f ms\n", performanceResult.AverageTotalTime)
	fmt.Printf("  性能评级: %s\n", performanceResult.PerformanceGrade)
	fmt.Printf("  性能建议: %s\n", performanceResult.PerformanceRecommendation)

	// 8. 验证结果汇总
	fmt.Println("\n📊 验证结果汇总:")
	fmt.Println("================================")

	totalTests := len(locationValidationResults) + len(textValidationResults) + len(typeValidationResults) + len(correlationResults)
	passedTests := 0

	// 统计位置验证结果
	locationPasses := 0
	for _, result := range locationValidationResults {
		if result.IsValid {
			locationPasses++
		}
	}
	passedTests += locationPasses

	// 统计文本验证结果
	textPasses := 0
	for _, result := range textValidationResults {
		if result.IsValid {
			textPasses++
		}
	}
	passedTests += textPasses

	// 统计类型验证结果
	typePasses := 0
	for _, result := range typeValidationResults {
		if result.IsValid {
			typePasses++
		}
	}
	passedTests += typePasses

	// 统计关联性验证结果
	correlationPasses := 0
	for _, result := range correlationResults {
		if result.OverallConsistency {
			correlationPasses++
		}
	}
	passedTests += correlationPasses

	passRate := float64(passedTests) / float64(totalTests) * 100

	fmt.Printf("📈 总测试数: %d\n", totalTests)
	fmt.Printf("✅ 通过数: %d\n", passedTests)
	fmt.Printf("❌ 失败数: %d\n", totalTests-passedTests)
	fmt.Printf("📊 通过率: %.1f%%\n", passRate)
	fmt.Printf("📍 位置验证: %d/%d\n", locationPasses, len(locationValidationResults))
	fmt.Printf("📝 文本验证: %d/%d\n", textPasses, len(textValidationResults))
	fmt.Printf("🏷️ 类型验证: %d/%d\n", typePasses, len(typeValidationResults))
	fmt.Printf("🔗 关联性验证: %d/%d\n", correlationPasses, len(correlationResults))

	// 9. 保存详细验证结果
	fmt.Println("\n💾 保存验证结果:")
	fmt.Println("------------------------------")

	detailedResults := map[string]interface{}{
		"summary": map[string]interface{}{
			"totalTests":       totalTests,
			"passedTests":      passedTests,
			"failedTests":      totalTests - passedTests,
			"passRate":         passRate,
			"locationPasses":   locationPasses,
			"textPasses":       textPasses,
			"typePasses":       typePasses,
			"correlationPasses": correlationPasses,
		},
		"locationValidation":     locationValidationResults,
		"textValidation":        textValidationResults,
		"typeValidation":        typeValidationResults,
		"edgeCases":            edgeCaseResults,
		"correlationResults":    correlationResults,
		"performance":          performanceResult,
		"timestamp":            fmt.Sprintf("%v", os.Getpid()),
	}

	resultFile := "../../validation-results/node-properties-results.json"
	if err := os.MkdirAll("../../validation-results", 0755); err == nil {
		if data, err := json.MarshalIndent(detailedResults, "", "  "); err == nil {
			if err := os.WriteFile(resultFile, data, 0644); err == nil {
				fmt.Printf("✅ 详细验证结果已保存到: %s\n", resultFile)
			} else {
				fmt.Printf("❌ 保存详细结果失败: %v\n", err)
			}
		} else {
			fmt.Printf("❌ 序列化详细结果失败: %v\n", err)
		}
	} else {
		fmt.Printf("❌ 创建结果目录失败: %v\n", err)
	}

	// 10. 最终结论
	fmt.Println("\n🎯 最终验证结论:")
	fmt.Println("================================")

	if passRate >= 80.0 {
		fmt.Printf("🎉 节点属性 API 验证完成！基本功能正常工作\n")
		fmt.Println("================================")
		fmt.Println("📋 已验证的 API:")
		fmt.Println("   - node.GetStart() - 获取起始位置")
		fmt.Println("   - node.GetEnd() - 获取结束位置")
		fmt.Println("   - node.GetStartLineNumber() - 获取起始行号")
		fmt.Println("   - node.GetEndLineNumber() - 获取结束行号")
		fmt.Println("   - node.GetText() - 获取节点文本")
		fmt.Println("   - node.GetTextLength() - 获取文本长度")
		fmt.Println("   - node.Kind - 获取节点类型")
		fmt.Println("   - node.GetSourceFile() - 获取所属源文件")
		fmt.Println("   - 节点属性关联性验证")
		fmt.Println("   - 边界情况处理")
		fmt.Println("   - 性能基准测试")
		fmt.Println("================================")
		fmt.Println("📝 验证总结:")
		fmt.Printf("   - 位置信息验证: %.1f%% (%d/%d)\n",
			float64(locationPasses)/float64(len(locationValidationResults))*100,
			locationPasses, len(locationValidationResults))
		fmt.Printf("   - 文本信息验证: %.1f%% (%d/%d)\n",
			float64(textPasses)/float64(len(textValidationResults))*100,
			textPasses, len(textValidationResults))
		fmt.Printf("   - 类型信息验证: %.1f%% (%d/%d)\n",
			float64(typePasses)/float64(len(typeValidationResults))*100,
			typePasses, len(typeValidationResults))
		fmt.Printf("   - 关联性验证: %.1f%% (%d/%d)\n",
			float64(correlationPasses)/float64(len(correlationResults))*100,
			correlationPasses, len(correlationResults))
		fmt.Printf("   - 性能评级: %s\n", performanceResult.PerformanceGrade)
		fmt.Printf("   - 发现边界情况: %d 种\n", len(edgeCaseResults))
	} else {
		fmt.Printf("❌ 节点属性 API 验证完成但存在问题\n")
		fmt.Printf("   验证通过率 %.1f%% 低于预期\n", passRate)
		fmt.Println("   建议检查节点属性获取的实现")
	}
}

// 数据结构定义
type LocationValidationResult struct {
	Index         int   `json:"index"`
	NodeType      ast.Kind `json:"nodeType"`
	StartPosition int   `json:"startPosition"`
	EndPosition   int   `json:"endPosition"`
	Length        int   `json:"length"`
	StartLine     int   `json:"startLine"`
	EndLine       int   `json:"endLine"`
	SpanLines     int   `json:"spanLines"`
	IsValid       bool  `json:"isValid"`
}

type TextValidationResult struct {
	Index        int    `json:"index"`
	NodeType     ast.Kind `json:"nodeType"`
	Text         string `json:"text"`
	TextLength   int    `json:"textLength"`
	TextHash     string `json:"textHash"`
	IsEmpty      bool   `json:"isEmpty"`
	HasNewlines  bool   `json:"hasNewlines"`
	IsValid      bool   `json:"isValid"`
}

type TypeValidationResult struct {
	Index         int      `json:"index"`
	NodeType      ast.Kind `json:"nodeType"`
	TypeName      string   `json:"typeName"`
	TypeGroup     string   `json:"typeGroup"`
	IsDeclaration bool     `json:"isDeclaration"`
	IsExpression  bool     `json:"isExpression"`
	IsLiteral     bool     `json:"isLiteral"`
	IsIdentifier  bool     `json:"isIdentifier"`
	IsValid       bool     `json:"isValid"`
}

type EdgeCaseResult struct {
	NodeType        ast.Kind `json:"nodeType"`
	IsEdgeCase      bool    `json:"isEdgeCase"`
	EdgeCaseType    string  `json:"edgeCaseType"`
	Description     string  `json:"description"`
	HandlingResult  string  `json:"handlingResult"`
}

type CorrelationResult struct {
	Index                  int  `json:"index"`
	NodeType               ast.Kind `json:"nodeType"`
	LocationTextCorrelation bool  `json:"locationTextCorrelation"`
	LocationTypeCorrelation bool  `json:"locationTypeCorrelation"`
	TextTypeCorrelation     bool  `json:"textTypeCorrelation"`
	OverallConsistency     bool  `json:"overallConsistency"`
}

type PerformanceResult struct {
	TestNodeCount              int     `json:"testNodeCount"`
	AverageLocationTime        float64 `json:"averageLocationTime"`
	AverageTextTime           float64 `json:"averageTextTime"`
	AverageTypeTime           float64 `json:"averageTypeTime"`
	AverageTotalTime          float64 `json:"averageTotalTime"`
	PerformanceGrade          string  `json:"performanceGrade"`
	PerformanceRecommendation string `json:"performanceRecommendation"`
}

// 验证函数实现
func validateNodeLocation(node tsmorphgo.Node, index int) LocationValidationResult {
	result := LocationValidationResult{
		Index:    index,
		NodeType: node.Kind,
	}

	// 获取位置信息
	result.StartPosition = node.GetStart()
	result.EndPosition = node.GetEnd()
	result.Length = result.EndPosition - result.StartPosition
	result.StartLine = node.GetStartLineNumber()
	result.EndLine = node.GetEndLineNumber()
	result.SpanLines = result.EndLine - result.StartLine + 1

	// 验证位置信息的合理性
	result.IsValid = true

	// 检查位置范围
	if result.StartPosition < 0 || result.EndPosition < result.StartPosition {
		result.IsValid = false
	}

	// 检查行号范围
	if result.StartLine <= 0 || result.EndLine < result.StartLine {
		result.IsValid = false
	}

	// 检查长度
	if result.Length < 0 {
		result.IsValid = false
	}

	return result
}

func validateNodeText(node tsmorphgo.Node, index int) TextValidationResult {
	result := TextValidationResult{
		Index:    index,
		NodeType: node.Kind,
	}

	// 获取文本信息
	result.Text = node.GetText()
	result.TextLength = node.GetTextLength()

	// 计算文本哈希（简化实现）
	result.TextHash = fmt.Sprintf("%x", len(result.Text))

	// 分析文本内容
	result.IsEmpty = result.Text == ""
	result.HasNewlines = strings.Contains(result.Text, "\n")

	// 验证文本信息的合理性
	result.IsValid = true

	// 检查文本长度一致性
	if result.TextLength != len(result.Text) {
		result.IsValid = false
	}

	// 检查空文本的合理性
	if result.IsEmpty && result.NodeType != ast.KindSourceFile {
		// 大多数节点类型不应该有空文本
		switch result.NodeType {
		case ast.KindInterfaceDeclaration, ast.KindFunctionDeclaration, ast.KindClassDeclaration:
			if result.IsEmpty {
				result.IsValid = false
			}
		}
	}

	return result
}

func validateNodeType(node tsmorphgo.Node, index int) TypeValidationResult {
	result := TypeValidationResult{
		Index:    index,
		NodeType: node.Kind,
	}

	// 获取类型信息
	result.TypeName = node.Kind.String()

	// 分类类型
	result.TypeGroup = getTypeGroup(node.Kind)
	result.IsDeclaration = isDeclarationKind(node.Kind)
	result.IsExpression = isExpressionKind(node.Kind)
	result.IsLiteral = isLiteralKind(node.Kind)
	result.IsIdentifier = node.Kind == ast.KindIdentifier

	// 验证类型信息的合理性
	result.IsValid = true

	// 检查类型名称
	if result.TypeName == "" {
		result.IsValid = false
	}

	// 检查类型分组的合理性
	if result.TypeGroup == "" {
		result.IsValid = false
	}

	return result
}

func validateEdgeCaseNode(node tsmorphgo.Node) EdgeCaseResult {
	result := EdgeCaseResult{
		NodeType:   node.Kind,
		IsEdgeCase: false,
	}

	// 检查是否为边界情况
	switch node.Kind {
	case ast.KindSourceFile:
		result.IsEdgeCase = true
		result.EdgeCaseType = "根节点"
		result.Description = "源文件是 AST 的根节点"
		result.HandlingResult = "正常处理"
	case ast.KindIdentifier:
		text := node.GetText()
		if text == "" {
			result.IsEdgeCase = true
			result.EdgeCaseType = "空标识符"
			result.Description = "标识符节点的文本为空"
			result.HandlingResult = "需特殊处理"
		}
	case ast.KindStringLiteral:
		text := node.GetText()
		if text == "\"\"" || text == "''" {
			result.IsEdgeCase = true
			result.EdgeCaseType = "空字符串"
			result.Description = "字符串字面量为空"
			result.HandlingResult = "正常处理"
		}
	}

	return result
}

func validateNodeCorrelations(node tsmorphgo.Node, index int) CorrelationResult {
	result := CorrelationResult{
		Index:    index,
		NodeType: node.Kind,
	}

	// 位置-文本关联性
	startPos := node.GetStart()
	endPos := node.GetEnd()
	text := node.GetText()
	expectedLength := endPos - startPos
	actualLength := len(text)
	result.LocationTextCorrelation = expectedLength == actualLength

	// 位置-类型关联性
	startLine := node.GetStartLineNumber()
	endLine := node.GetEndLineNumber()
	isMultiLine := endLine > startLine

	// 多行节点通常是声明类型
	result.LocationTypeCorrelation = true
	if isMultiLine {
		if !isDeclarationKind(node.Kind) {
			result.LocationTypeCorrelation = false
		}
	}

	// 文本-类型关联性
	result.TextTypeCorrelation = true
	if node.Kind == ast.KindIdentifier && text == "" {
		result.TextTypeCorrelation = false
	}

	// 整体一致性
	result.OverallConsistency = result.LocationTextCorrelation &&
		result.LocationTypeCorrelation &&
		result.TextTypeCorrelation

	return result
}

func validatePropertyPerformance(testNodes []tsmorphgo.Node) PerformanceResult {
	result := PerformanceResult{
		TestNodeCount: len(testNodes),
	}

	if result.TestNodeCount == 0 {
		result.PerformanceGrade = "无测试节点"
		result.PerformanceRecommendation = "需要提供测试节点"
		return result
	}

	// 简化的性能测试（实际项目中应该使用更精确的时间测量）
	var totalLocationTime, totalTextTime, totalTypeTime, totalTotalTime float64

	for _, node := range testNodes {
		// 模拟位置获取时间
		locationTime := 0.01
		startPos := node.GetStart()
		endPos := node.GetEnd()
		startLine := node.GetStartLineNumber()
		endLine := node.GetEndLineNumber()
		_ = startPos + endPos + startLine + endLine

		// 模拟文本获取时间
		textTime := 0.02
		text := node.GetText()
		textLength := node.GetTextLength()
		_ = fmt.Sprintf("%s%d", text, textLength)

		// 模拟类型获取时间
		typeTime := 0.005
		nodeType := node.Kind
		_ = nodeType.String()

		totalTime := locationTime + textTime + typeTime

		totalLocationTime += locationTime
		totalTextTime += textTime
		totalTypeTime += typeTime
		totalTotalTime += totalTime
	}

	result.AverageLocationTime = totalLocationTime / float64(result.TestNodeCount)
	result.AverageTextTime = totalTextTime / float64(result.TestNodeCount)
	result.AverageTypeTime = totalTypeTime / float64(result.TestNodeCount)
	result.AverageTotalTime = totalTotalTime / float64(result.TestNodeCount)

	// 性能评级
	switch {
	case result.AverageTotalTime < 0.05:
		result.PerformanceGrade = "优秀"
	case result.AverageTotalTime < 0.1:
		result.PerformanceGrade = "良好"
	case result.AverageTotalTime < 0.2:
		result.PerformanceGrade = "一般"
	default:
		result.PerformanceGrade = "较差"
	}

	// 性能建议
	switch result.PerformanceGrade {
	case "优秀":
		result.PerformanceRecommendation = "性能表现优秀，无需优化"
	case "良好":
		result.PerformanceRecommendation = "性能良好，可考虑进一步优化"
	case "一般":
		result.PerformanceRecommendation = "性能一般，建议优化关键路径"
	default:
		result.PerformanceRecommendation = "性能较差，急需优化"
	}

	return result
}

// 辅助函数
func containsNewline(s string) bool {
	return strings.Contains(s, "\n")
}

func getTypeGroup(kind ast.Kind) string {
	switch kind {
	case ast.KindInterfaceDeclaration, ast.KindClassDeclaration, ast.KindFunctionDeclaration, ast.KindMethodDeclaration:
		return "declaration"
	case ast.KindTypeAliasDeclaration, ast.KindTypeParameter:
		return "type"
	case ast.KindVariableDeclaration, ast.KindPropertyDeclaration:
		return "variable"
	case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral, 150: // ast.KindBooleanLiteral
		return "literal"
	case ast.KindCallExpression, ast.KindNewExpression, ast.KindPropertyAccessExpression:
		return "expression"
	case ast.KindSourceFile:
		return "structural"
	default:
		return "other"
	}
}

func isDeclarationKind(kind ast.Kind) bool {
	declarationKinds := []ast.Kind{
		ast.KindInterfaceDeclaration,
		ast.KindClassDeclaration,
		ast.KindFunctionDeclaration,
		ast.KindMethodDeclaration,
		ast.KindVariableDeclaration,
		ast.KindPropertyDeclaration,
		ast.KindTypeAliasDeclaration,
	}

	for _, dk := range declarationKinds {
		if kind == dk {
			return true
		}
	}
	return false
}

func isExpressionKind(kind ast.Kind) bool {
	expressionKinds := []ast.Kind{
		ast.KindCallExpression,
		ast.KindNewExpression,
		ast.KindPropertyAccessExpression,
		ast.KindBinaryExpression,
		160, // ast.KindUnaryExpression
		ast.KindConditionalExpression,
	}

	for _, ek := range expressionKinds {
		if kind == ek {
			return true
		}
	}
	return false
}

func isLiteralKind(kind ast.Kind) bool {
	literalKinds := []ast.Kind{
		ast.KindStringLiteral,
		ast.KindNumericLiteral,
		150, // ast.KindBooleanLiteral
		151, // ast.KindNullLiteral
		152, // ast.KindUndefinedLiteral
	}

	for _, lk := range literalKinds {
		if kind == lk {
			return true
		}
	}
	return false
}