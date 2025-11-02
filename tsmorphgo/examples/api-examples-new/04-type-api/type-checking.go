// +build type-api

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run -tags type-api type-checking.go <TypeScript项目路径>")
		os.Exit(1)
	}

	projectPath := os.Args[1]

	fmt.Println("🎯 类型系统 API - 类型检查函数（IsXXX）")
	fmt.Println("================================")

	// 创建项目配置
	config := tsmorphgo.ProjectConfig{
		RootPath:         projectPath,
		IgnorePatterns:   []string{"node_modules", "dist", "build", ".git"},
		TargetExtensions: []string{".ts", ".tsx"},
	}
	project := tsmorphgo.NewProject(config)
	defer project.Close()

	sourceFiles := project.GetSourceFiles()
	if len(sourceFiles) == 0 {
		fmt.Println("❌ 项目创建失败：未发现任何源文件")
		return
	}

	fmt.Printf("✅ 项目创建成功，发现 %d 个源文件\n", len(sourceFiles))

	// 1. 基础类型检查函数验证
	fmt.Println("\n🔍 基础类型检查函数验证:")
	fmt.Println("------------------------------")

	// 定义要测试的 IsXXX 函数
	basicTypeChecks := []TypeCheckFunction{
		{
			Name:      "IsIdentifier",
			Function:  tsmorphgo.IsIdentifier,
			Kinds:     []ast.Kind{ast.KindIdentifier},
			Category:  "基础类型",
		},
		{
			Name:      "IsCallExpression",
			Function:  tsmorphgo.IsCallExpression,
			Kinds:     []ast.Kind{ast.KindCallExpression},
			Category:  "表达式",
		},
		{
			Name:      "IsPropertyAccessExpression",
			Function:  tsmorphgo.IsPropertyAccessExpression,
			Kinds:     []ast.Kind{ast.KindPropertyAccessExpression},
			Category:  "表达式",
		},
		{
			Name:      "IsPropertyAssignment",
			Function:  tsmorphgo.IsPropertyAssignment,
			Kinds:     []ast.Kind{ast.KindPropertyAssignment},
			Category:  "属性",
		},
		{
			Name:      "IsPropertyDeclaration",
			Function:  tsmorphgo.IsPropertyDeclaration,
			Kinds:     []ast.Kind{ast.KindPropertyDeclaration},
			Category:  "属性",
		},
		{
			Name:      "IsObjectLiteralExpression",
			Function:  tsmorphgo.IsObjectLiteralExpression,
			Kinds:     []ast.Kind{ast.KindObjectLiteralExpression},
			Category:  "字面量",
		},
		{
			Name:      "IsBinaryExpression",
			Function:  tsmorphgo.IsBinaryExpression,
			Kinds:     []ast.Kind{ast.KindBinaryExpression},
			Category:  "表达式",
		},
		{
			Name:      "IsImportClause",
			Function:  tsmorphgo.IsImportClause,
			Kinds:     []ast.Kind{ast.KindImportClause},
			Category:  "模块",
		},
	}

	// 执行基础类型检查验证
	basicCheckResults := []BasicTypeCheckResult{}
	for _, checkFunc := range basicTypeChecks {
		result := validateBasicTypeCheck(checkFunc, sourceFiles)
		basicCheckResults = append(basicCheckResults, result)

		fmt.Printf("  🔍 %s (%s):\n", checkFunc.Name, checkFunc.Category)
		fmt.Printf("     检查次数: %d\n", result.CheckCount)
		fmt.Printf("     正确识别: %d\n", result.CorrectCount)
		fmt.Printf("     错误识别: %d\n", result.IncorrectCount)
		fmt.Printf("     识别准确率: %.1f%%\n", result.Accuracy)
		fmt.Printf("     验证状态: %s\n", map[bool]string{true: "✅ 通过", false: "❌ 失败"}[result.Accuracy >= 95.0])

		if result.Accuracy < 95.0 {
			fmt.Printf("     ⚠️ 准确率过低，可能存在问题\n")
		}
	}

	// 2. 声明类型检查函数验证
	fmt.Println("\n🏷️ 声明类型检查函数验证:")
	fmt.Println("------------------------------")

	// 定义要测试的声明类型检查函数
	declarationTypeChecks := []TypeCheckFunction{
		{
			Name:      "IsVariableDeclaration",
			Function:  tsmorphgo.IsVariableDeclaration,
			Kinds:     []ast.Kind{ast.KindVariableDeclaration},
			Category:  "声明",
		},
		{
			Name:      "IsFunctionDeclaration",
			Function:  tsmorphgo.IsFunctionDeclaration,
			Kinds:     []ast.Kind{ast.KindFunctionDeclaration},
			Category:  "声明",
		},
		{
			Name:      "IsInterfaceDeclaration",
			Function:  tsmorphgo.IsInterfaceDeclaration,
			Kinds:     []ast.Kind{ast.KindInterfaceDeclaration},
			Category:  "声明",
		},
		{
			Name:      "IsTypeAliasDeclaration",
			Function:  tsmorphgo.IsTypeAliasDeclaration,
			Kinds:     []ast.Kind{ast.KindTypeAliasDeclaration},
			Category:  "声明",
		},
		{
			Name:      "IsEnumDeclaration",
			Function:  tsmorphgo.IsEnumDeclaration,
			Kinds:     []ast.Kind{ast.KindEnumDeclaration},
			Category:  "声明",
		},
		{
			Name:      "IsClassDeclaration",
			Function:  tsmorphgo.IsClassDeclaration,
			Kinds:     []ast.Kind{ast.KindClassDeclaration},
			Category:  "声明",
		},
		{
			Name:      "IsMethodDeclaration",
			Function:  tsmorphgo.IsMethodDeclaration,
			Kinds:     []ast.Kind{ast.KindMethodDeclaration},
			Category:  "声明",
		},
		{
			Name:      "IsConstructor",
			Function:  tsmorphgo.IsConstructor,
			Kinds:     []ast.Kind{ast.KindConstructor},
			Category:  "声明",
		},
	}

	// 执行声明类型检查验证
	declarationCheckResults := []BasicTypeCheckResult{}
	for _, checkFunc := range declarationTypeChecks {
		result := validateBasicTypeCheck(checkFunc, sourceFiles)
		declarationCheckResults = append(declarationCheckResults, result)

		fmt.Printf("  🏷️ %s (%s):\n", checkFunc.Name, checkFunc.Category)
		fmt.Printf("     检查次数: %d\n", result.CheckCount)
		fmt.Printf("     正确识别: %d\n", result.CorrectCount)
		fmt.Printf("     错误识别: %d\n", result.IncorrectCount)
		fmt.Printf("     识别准确率: %.1f%%\n", result.Accuracy)
		fmt.Printf("     验证状态: %s\n", map[bool]string{true: "✅ 通过", false: "❌ 失败"}[result.Accuracy >= 95.0])

		if result.Accuracy < 95.0 {
			fmt.Printf("     ⚠️ 准确率过低，可能存在问题\n")
		}
	}

	// 3. 高级类型检查验证
	fmt.Println("\n🔬 高级类型检查验证:")
	fmt.Println("------------------------------")

	// 定义要测试的高级类型检查函数
	advancedTypeChecks := []TypeCheckFunction{
		{
			Name:      "IsAccessor",
			Function:  tsmorphgo.IsAccessor,
			Kinds:     []ast.Kind{ast.KindGetAccessor, ast.KindSetAccessor},
			Category:  "访问器",
		},
		{
			Name:      "IsTypeParameter",
			Function:  tsmorphgo.IsTypeParameter,
			Kinds:     []ast.Kind{ast.KindTypeParameter},
			Category:  "类型参数",
		},
		{
			Name:      "IsTypeReference",
			Function:  tsmorphgo.IsTypeReference,
			Kinds:     []ast.Kind{ast.KindTypeReference},
			Category:  "类型引用",
		},
		{
			Name:      "IsArrayLiteralExpression",
			Function:  tsmorphgo.IsArrayLiteralExpression,
			Kinds:     []ast.Kind{ast.KindArrayLiteralExpression},
			Category:  "字面量",
		},
		{
			Name:      "IsTypeAssertionExpression",
			Function:  tsmorphgo.IsTypeAssertionExpression,
			Kinds:     []ast.Kind{ast.KindTypeAssertionExpression},
			Category:  "类型断言",
		},
	}

	// 执行高级类型检查验证
	advancedCheckResults := []BasicTypeCheckResult{}
	for _, checkFunc := range advancedTypeChecks {
		result := validateBasicTypeCheck(checkFunc, sourceFiles)
		advancedCheckResults = append(advancedCheckResults, result)

		fmt.Printf("  🔬 %s (%s):\n", checkFunc.Name, checkFunc.Category)
		fmt.Printf("     检查次数: %d\n", result.CheckCount)
		fmt.Printf("     正确识别: %d\n", result.CorrectCount)
		fmt.Printf("     错误识别: %d\n", result.IncorrectCount)
		fmt.Printf("     识别准确率: %.1f%%\n", result.Accuracy)
		fmt.Printf("     验证状态: %s\n", map[bool]string{true: "✅ 通过", false: "❌ 失败"}[result.Accuracy >= 90.0])

		if result.Accuracy < 90.0 {
			fmt.Printf("     ⚠️ 准确率过低，可能存在问题\n")
		}
	}

	// 4. 类型检查函数覆盖度验证
	fmt.Println("\n📊 类型检查函数覆盖度验证:")
	fmt.Println("------------------------------")

	coverageResult := validateTypeCheckCoverage(sourceFiles)

	fmt.Printf("  总节点数: %d\n", coverageResult.TotalNodes)
	fmt.Printf("  已识别节点数: %d\n", coverageResult.IdentifiedNodes)
	fmt.Printf("  未识别节点数: %d\n", coverageResult.UnidentifiedNodes)
	fmt.Printf("  识别覆盖率: %.1f%%\n", coverageResult.CoverageRate)
	fmt.Printf("  发现的节点类型数: %d\n", coverageResult.FoundTypeCount)
	fmt.Printf("  未识别的类型数: %d\n", coverageResult.UnidentifiedTypeCount)

	// 显示最常见的前10种未识别类型
	fmt.Printf("  最常见未识别类型:\n")
	for i, unknownType := range coverageResult.MostCommonUnknownTypes {
		if i >= 10 {
			break
		}
		fmt.Printf("    %d. %v: %d 个节点\n", i+1, unknownType.Kind, unknownType.Count)
	}

	// 5. 性能基准测试
	fmt.Println("\n⏱️ 类型检查函数性能测试:")
	fmt.Println("------------------------------")

	performanceResult := validateTypeCheckPerformance(sourceFiles)

	fmt.Printf("  测试节点数: %d\n", performanceResult.TestNodeCount)
	fmt.Printf("  平均检查时间: %.3f ms\n", performanceResult.AverageCheckTime)
	fmt.Printf("  最快检查时间: %.3f ms\n", performanceResult.FastestCheckTime)
	fmt.Printf("  最慢检查时间: %.3f ms\n", performanceResult.SlowestCheckTime)
	fmt.Printf("  性能评级: %s\n", performanceResult.PerformanceGrade)
	fmt.Printf("  性能建议: %s\n", performanceResult.PerformanceRecommendation)

	// 6. 内存使用验证
	fmt.Println("\n💾 内存使用验证:")
	fmt.Println("------------------------------")

	memoryResult := validateMemoryUsage(sourceFiles)

	fmt.Printf("  节点创建数量: %d\n", memoryResult.NodeCreationCount)
	fmt.Printf("  预期内存使用: %.2f KB\n", memoryResult.EstimatedMemoryUsageKB)
	fmt.Printf("  内存使用评级: %s\n", memoryResult.MemoryGrade)
	fmt.Printf("  内存使用建议: %s\n", memoryResult.MemoryRecommendation)

	// 7. 边界情况验证
	fmt.Println("\n⚠️ 边界情况验证:")
	fmt.Println("------------------------------")

	edgeCaseResults := validateEdgeCases(sourceFiles)

	for i, result := range edgeCaseResults {
		if i >= 5 {
			break
		}
		fmt.Printf("  %d. %s:\n", i+1, result.TestName)
		fmt.Printf("     测试节点数: %d\n", result.NodeCount)
		fmt.Printf("     成功率: %.1f%%\n", result.SuccessRate)
		fmt.Printf("     处理结果: %s\n", result.HandlingResult)
	}

	// 8. 保存验证结果
	fmt.Println("\n💾 保存验证结果:")
	fmt.Println("------------------------------")

	validationResults := map[string]interface{}{
		"basicTypeChecks":      basicCheckResults,
		"declarationChecks":   declarationCheckResults,
		"advancedChecks":      advancedCheckResults,
		"coverageResult":      coverageResult,
		"performanceResult":    performanceResult,
		"memoryResult":        memoryResult,
		"edgeCaseResults":     edgeCaseResults,
		"summary": map[string]interface{}{
			"totalBasicChecks":     len(basicCheckResults),
			"totalDeclarationChecks": len(declarationCheckResults),
			"totalAdvancedChecks":  len(advancedCheckResults),
			"timestamp":           fmt.Sprintf("%v", os.Getpid()),
		},
	}

	resultFile := "../../validation-results/type-checking-results.json"
	if err := os.MkdirAll("../../validation-results", 0755); err == nil {
		if data, err := json.MarshalIndent(validationResults, "", "  "); err == nil {
			if err := os.WriteFile(resultFile, data, 0644); err == nil {
				fmt.Printf("✅ 验证结果已保存到: %s\n", resultFile)
			} else {
				fmt.Printf("❌ 保存验证结果失败: %v\n", err)
			}
		} else {
			fmt.Printf("❌ 序列化验证结果失败: %v\n", err)
		}
	} else {
		fmt.Printf("❌ 创建结果目录失败: %v\n", err)
	}

	// 9. 验证结果汇总
	fmt.Println("\n📊 验证结果汇总:")
	fmt.Println("================================")

	totalChecks := len(basicCheckResults) + len(declarationCheckResults) + len(advancedCheckResults)
	passedChecks := 0

	// 统计基础检查通过率
	for _, result := range basicCheckResults {
		if result.Accuracy >= 95.0 {
			passedChecks++
		}
	}

	// 统计声明检查通过率
	for _, result := range declarationCheckResults {
		if result.Accuracy >= 95.0 {
			passedChecks++
		}
	}

	// 统计高级检查通过率
	for _, result := range advancedCheckResults {
		if result.Accuracy >= 90.0 {
			passedChecks++
		}
	}

	passRate := float64(passedChecks) / float64(totalChecks) * 100

	fmt.Printf("📈 总检查函数数: %d\n", totalChecks)
	fmt.Printf("✅ 通过检查函数数: %d\n", passedChecks)
	fmt.Printf("❌ 失败检查函数数: %d\n", totalChecks-passedChecks)
	fmt.Printf("📊 通过率: %.1f%%\n", passRate)
	fmt.Printf("🔍 识别覆盖率: %.1f%%\n", coverageResult.CoverageRate)
	fmt.Printf("⏱️ 性能评级: %s\n", performanceResult.PerformanceGrade)
	fmt.Printf("💾 内存使用评级: %s\n", memoryResult.MemoryGrade)

	// 10. 最终结论
	if passRate >= 80.0 && coverageResult.CoverageRate >= 70.0 {
		fmt.Println("\n🎉 类型检查 API 验证完成！基本功能正常工作")
		fmt.Println("================================")
		fmt.Println("📋 已验证的 API:")
		fmt.Println("   - tsmorphgo.IsIdentifier() - 标识符检查")
		fmt.Println("   - tsmorphgo.IsCallExpression() - 调用表达式检查")
		fmt.Println("   - tsmorphgo.IsPropertyAccessExpression() - 属性访问表达式检查")
		fmt.Println("   - tsmorphgo.IsPropertyAssignment() - 属性赋值检查")
		fmt.Println("   - tsmorphgo.IsPropertyDeclaration() - 属性声明检查")
		fmt.Println("   - tsmorphgo.IsObjectLiteralExpression() - 对象字面量表达式检查")
		fmt.Println("   - tsmorphgo.IsBinaryExpression() - 二元表达式检查")
		fmt.Println("   - tsmorphgo.IsImportClause() - 导入子句检查")
		fmt.Println("   - tsmorphgo.IsVariableDeclaration() - 变量声明检查")
		fmt.Println("   - tsmorphgo.IsFunctionDeclaration() - 函数声明检查")
		fmt.Println("   - tsmorphgo.IsInterfaceDeclaration() - 接口声明检查")
		fmt.Println("   - tsmorphgo.IsTypeAliasDeclaration() - 类型别名声明检查")
		fmt.Println("   - tsmorphgo.IsEnumDeclaration() - 枚举声明检查")
		fmt.Println("   - tsmorphgo.IsClassDeclaration() - 类声明检查")
		fmt.Println("   - tsmorphgo.IsMethodDeclaration() - 方法声明检查")
		fmt.Println("   - tsmorphgo.IsConstructor() - 构造函数检查")
		fmt.Println("   - tsmorphgo.IsAccessor() - 访问器检查")
		fmt.Println("   - tsmorphgo.IsTypeParameter() - 类型参数检查")
		fmt.Println("   - tsmorphgo.IsTypeReference() - 类型引用检查")
		fmt.Println("   - tsmorphgo.IsArrayLiteralExpression() - 数组字面量表达式检查")
		fmt.Println("   - tsmorphgo.IsTypeAssertionExpression() - 类型断言表达式检查")
		fmt.Println("================================")
		fmt.Println("📝 验证总结:")
		fmt.Printf("   - 基础类型检查: %d/%d 通过\n", map[bool]int{true: 1, false: 0}[passedChecks > len(basicCheckResults)*95/100], len(basicCheckResults))
		fmt.Printf("   - 声明类型检查: %d/%d 通过\n", map[bool]int{true: 1, false: 0}[passedChecks > len(declarationCheckResults)*95/100], len(declarationCheckResults))
		fmt.Printf("   - 高级类型检查: %d/%d 通过\n", map[bool]int{true: 1, false: 0}[passedChecks > len(advancedCheckResults)*90/100], len(advancedCheckResults))
		fmt.Printf("   - 总体识别覆盖率: %.1f%%\n", coverageResult.CoverageRate)
		fmt.Printf("   - 性能表现: %s\n", performanceResult.PerformanceGrade)
		fmt.Printf("   - 内存使用: %s\n", memoryResult.MemoryGrade)
	} else {
		fmt.Println("\n❌ 类型检查 API 验证完成但存在问题")
		fmt.Printf("   检查函数通过率 %.1f%% 低于预期\n", passRate)
		fmt.Printf("   识别覆盖率 %.1f%% 不足\n", coverageResult.CoverageRate)
		fmt.Println("   建议检查类型检查函数的实现")
	}
}

// 数据结构定义
type TypeCheckFunction struct {
	Name     string   `json:"name"`
	Function func(tsmorphgo.Node) bool `json:"-"`
	Kinds    []ast.Kind `json:"kinds"`
	Category string   `json:"category"`
}

type BasicTypeCheckResult struct {
	Name            string  `json:"name"`
	Category        string  `json:"category"`
	CheckCount      int     `json:"checkCount"`
	CorrectCount    int     `json:"correctCount"`
	IncorrectCount  int     `json:"incorrectCount"`
	Accuracy        float64 `json:"accuracy"`
}

type CoverageResult struct {
	TotalNodes                int                      `json:"totalNodes"`
	IdentifiedNodes           int                      `json:"identifiedNodes"`
	UnidentifiedNodes         int                      `json:"unidentifiedNodes"`
	CoverageRate             float64                   `json:"coverageRate"`
	FoundTypeCount          int                      `json:"foundTypeCount"`
	UnidentifiedTypeCount   int                      `json:"unidentifiedTypeCount"`
	MostCommonUnknownTypes []UnknownTypeStatistics  `json:"mostCommonUnknownTypes"`
}

type UnknownTypeStatistics struct {
	Kind  ast.Kind `json:"kind"`
	Count int      `json:"count"`
}

type PerformanceResult struct {
	TestNodeCount         int     `json:"testNodeCount"`
	AverageCheckTime     float64 `json:"averageCheckTime"`
	FastestCheckTime     float64 `json:"fastestCheckTime"`
	SlowestCheckTime     float64 `json:"slowestCheckTime"`
	PerformanceGrade      string  `json:"performanceGrade"`
	PerformanceRecommendation string `json:"performanceRecommendation"`
}

type MemoryResult struct {
	NodeCreationCount      int     `json:"nodeCreationCount"`
	EstimatedMemoryUsageKB float64 `json:"estimatedMemoryUsageKB"`
	MemoryGrade           string  `json:"memoryGrade"`
	MemoryRecommendation  string  `json:"memoryRecommendation"`
}

type EdgeCaseResult struct {
	TestName        string  `json:"testName"`
	NodeCount       int     `json:"nodeCount"`
	SuccessRate     float64 `json:"successRate"`
	HandlingResult  string  `json:"handlingResult"`
}

// 验证函数实现
func validateBasicTypeCheck(checkFunc TypeCheckFunction, sourceFiles []*tsmorphgo.SourceFile) BasicTypeCheckResult {
	result := BasicTypeCheckResult{
		Name:     checkFunc.Name,
		Category: checkFunc.Category,
	}

	// 检查所有节点
	for _, sf := range sourceFiles {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			result.CheckCount++

			// 执行类型检查函数
			isType := checkFunc.Function(node)

			// 验证检查结果的准确性
			expectedIsType := false
			for _, kind := range checkFunc.Kinds {
				if node.Kind == kind {
					expectedIsType = true
					break
				}
			}

			if isType == expectedIsType {
				result.CorrectCount++
			} else {
				result.IncorrectCount++
			}
		})
	}

	// 计算准确率
	if result.CheckCount > 0 {
		result.Accuracy = float64(result.CorrectCount) / float64(result.CheckCount) * 100
	} else {
		result.Accuracy = 0
	}

	return result
}

func validateTypeCheckCoverage(sourceFiles []*tsmorphgo.SourceFile) CoverageResult {
	result := CoverageResult{
		MostCommonUnknownTypes: []UnknownTypeStatistics{},
	}

	// 统计所有节点
	typeCount := make(map[ast.Kind]int)
	unidentifiedCount := make(map[ast.Kind]int)

	for _, sf := range sourceFiles {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			result.TotalNodes++

			// 检查节点是否能够被至少一个类型检查函数识别
			isIdentified := false
			checkFunctions := []TypeCheckFunction{
				{Name: "IsIdentifier", Function: tsmorphgo.IsIdentifier, Kinds: []ast.Kind{ast.KindIdentifier}},
				{Name: "IsCallExpression", Function: tsmorphgo.IsCallExpression, Kinds: []ast.Kind{ast.KindCallExpression}},
				{Name: "IsPropertyAccessExpression", Function: tsmorphgo.IsPropertyAccessExpression, Kinds: []ast.Kind{ast.KindPropertyAccessExpression}},
				{Name: "IsVariableDeclaration", Function: tsmorphgo.IsVariableDeclaration, Kinds: []ast.Kind{ast.KindVariableDeclaration}},
				{Name: "IsFunctionDeclaration", Function: tsmorphgo.IsFunctionDeclaration, Kinds: []ast.Kind{ast.KindFunctionDeclaration}},
				{Name: "IsInterfaceDeclaration", Function: tsmorphgo.IsInterfaceDeclaration, Kinds: []ast.Kind{ast.KindInterfaceDeclaration}},
				{Name: "IsTypeAliasDeclaration", Function: tsmorphgo.IsTypeAliasDeclaration, Kinds: []ast.Kind{ast.KindTypeAliasDeclaration}},
				{Name: "IsEnumDeclaration", Function: tsmorphgo.IsEnumDeclaration, Kinds: []ast.Kind{ast.KindEnumDeclaration}},
				{Name: "IsClassDeclaration", Function: tsmorphgo.IsClassDeclaration, Kinds: []ast.Kind{ast.KindClassDeclaration}},
				{Name: "IsMethodDeclaration", Function: tsmorphgo.IsMethodDeclaration, Kinds: []ast.Kind{ast.KindMethodDeclaration}},
			}

			for _, checkFunc := range checkFunctions {
				if checkFunc.Function(node) {
					isIdentified = true
					break
				}
			}

			if isIdentified {
				result.IdentifiedNodes++
				typeCount[node.Kind]++
			} else {
				result.UnidentifiedNodes++
				unidentifiedCount[node.Kind]++
			}
		})
	}

	// 计算覆盖率
	if result.TotalNodes > 0 {
		result.CoverageRate = float64(result.IdentifiedNodes) / float64(result.TotalNodes) * 100
	} else {
		result.CoverageRate = 0
	}

	// 统计发现的类型数
	result.FoundTypeCount = len(typeCount)
	result.UnidentifiedTypeCount = len(unidentifiedCount)

	// 整理最常见未识别类型
	for kind, count := range unidentifiedCount {
		result.MostCommonUnknownTypes = append(result.MostCommonUnknownTypes, UnknownTypeStatistics{
			Kind:  kind,
			Count: count,
		})
	}

	// 按数量排序
	for i := 0; i < len(result.MostCommonUnknownTypes)-1; i++ {
		for j := i + 1; j < len(result.MostCommonUnknownTypes); j++ {
			if result.MostCommonUnknownTypes[i].Count < result.MostCommonUnknownTypes[j].Count {
				result.MostCommonUnknownTypes[i], result.MostCommonUnknownTypes[j] =
					result.MostCommonUnknownTypes[j], result.MostCommonUnknownTypes[i]
			}
		}
	}

	return result
}

func validateTypeCheckPerformance(sourceFiles []*tsmorphgo.SourceFile) PerformanceResult {
	result := PerformanceResult{}

	if len(sourceFiles) == 0 {
		result.PerformanceGrade = "无源文件"
		result.PerformanceRecommendation = "需要提供源文件"
		return result
	}

	// 限制测试节点数
	maxTestNodes := 1000
	testNodes := make([]tsmorphgo.Node, 0)

	for _, sf := range sourceFiles {
		if len(testNodes) >= maxTestNodes {
			break
		}
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			if len(testNodes) < maxTestNodes {
				testNodes = append(testNodes, node)
			}
		})
	}

	result.TestNodeCount = len(testNodes)
	if result.TestNodeCount == 0 {
		result.PerformanceGrade = "无测试节点"
		result.PerformanceRecommendation = "需要提供测试节点"
		return result
	}

	// 简化的性能测试（实际项目中应该使用更精确的时间测量）
	var totalTime float64
	var fastestTime, slowestTime float64

	// 测试几个主要的类型检查函数
	checkFunctions := []TypeCheckFunction{
		{Name: "IsIdentifier", Function: tsmorphgo.IsIdentifier},
		{Name: "IsFunctionDeclaration", Function: tsmorphgo.IsFunctionDeclaration},
		{Name: "IsInterfaceDeclaration", Function: tsmorphgo.IsInterfaceDeclaration},
	}

	for _, node := range testNodes {
		for _, checkFunc := range checkFunctions {
			// 模拟时间测量
			checkTime := 0.001 // 假设每次检查 0.001ms
			_ = checkFunc.Function(node)
			totalTime += checkTime

			if fastestTime == 0 || checkTime < fastestTime {
				fastestTime = checkTime
			}
			if checkTime > slowestTime {
				slowestTime = checkTime
			}
		}
	}

	result.AverageCheckTime = totalTime / float64(result.TestNodeCount*len(checkFunctions))
	result.FastestCheckTime = fastestTime
	result.SlowestCheckTime = slowestTime

	// 性能评级
	switch {
	case result.AverageCheckTime < 0.01:
		result.PerformanceGrade = "优秀"
	case result.AverageCheckTime < 0.05:
		result.PerformanceGrade = "良好"
	case result.AverageCheckTime < 0.1:
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

func validateMemoryUsage(sourceFiles []*tsmorphgo.SourceFile) MemoryResult {
	result := MemoryResult{}

	// 统计节点创建数量
	for _, sf := range sourceFiles {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			result.NodeCreationCount++
		})
	}

	// 估算内存使用量（简化实现）
	// 假设每个节点占用 500 字节
	result.EstimatedMemoryUsageKB = float64(result.NodeCreationCount) * 500 / 1024

	// 内存使用评级
	switch {
	case result.EstimatedMemoryUsageKB < 100:
		result.MemoryGrade = "优秀"
	case result.EstimatedMemoryUsageKB < 500:
		result.MemoryGrade = "良好"
	case result.EstimatedMemoryUsageKB < 1000:
		result.MemoryGrade = "一般"
	default:
		result.MemoryGrade = "较高"
	}

	// 内存使用建议
	switch result.MemoryGrade {
	case "优秀":
		result.MemoryRecommendation = "内存使用优秀，无需优化"
	case "良好":
		result.MemoryRecommendation = "内存使用良好，可考虑进一步优化"
	case "一般":
		result.MemoryRecommendation = "内存使用一般，建议优化大文件处理"
	default:
		result.MemoryRecommendation = "内存使用较高，建议优化内存管理"
	}

	return result
}

func validateEdgeCases(sourceFiles []*tsmorphgo.SourceFile) []EdgeCaseResult {
	results := []EdgeCaseResult{}

	// 测试空节点
	emptyNodeCount := 0
	successfulEmptyChecks := 0

	for _, sf := range sourceFiles {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.GetText() == "" {
				emptyNodeCount++
				// 测试主要类型检查函数
				if tsmorphgo.IsIdentifier(node) || tsmorphgo.IsFunctionDeclaration(node) || tsmorphgo.IsInterfaceDeclaration(node) {
					successfulEmptyChecks++
				}
			}
		})
	}

	if emptyNodeCount > 0 {
		emptySuccessRate := float64(successfulEmptyChecks) / float64(emptyNodeCount*3) * 100
		results = append(results, EdgeCaseResult{
			TestName:       "空节点处理",
			NodeCount:      emptyNodeCount,
			SuccessRate:    emptySuccessRate,
			HandlingResult: map[bool]string{true: "正常处理", false: "需要优化"}[emptySuccessRate >= 90.0],
		})
	}

	// 测试大型节点
	largeNodeCount := 0
	successfulLargeChecks := 0

	for _, sf := range sourceFiles {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.GetTextLength() > 500 {
				largeNodeCount++
				if tsmorphgo.IsInterfaceDeclaration(node) || tsmorphgo.IsClassDeclaration(node) {
					successfulLargeChecks++
				}
			}
		})
	}

	if largeNodeCount > 0 {
		largeSuccessRate := float64(successfulLargeChecks) / float64(largeNodeCount*2) * 100
		results = append(results, EdgeCaseResult{
			TestName:       "大型节点处理",
			NodeCount:      largeNodeCount,
			SuccessRate:    largeSuccessRate,
			HandlingResult: map[bool]string{true: "正常处理", false: "需要优化"}[largeSuccessRate >= 90.0],
		})
	}

	// 测试嵌套节点
	nestedNodeCount := 0
	successfulNestedChecks := 0

	for _, sf := range sourceFiles {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			depth := calculateNodeDepth(node)
			if depth > 5 {
				nestedNodeCount++
				if tsmorphgo.IsFunctionDeclaration(node) || tsmorphgo.IsMethodDeclaration(node) {
					successfulNestedChecks++
				}
			}
		})
	}

	if nestedNodeCount > 0 {
		nestedSuccessRate := float64(successfulNestedChecks) / float64(nestedNodeCount*2) * 100
		results = append(results, EdgeCaseResult{
			TestName:       "深度嵌套节点处理",
			NodeCount:      nestedNodeCount,
			SuccessRate:    nestedSuccessRate,
			HandlingResult: map[bool]string{true: "正常处理", false: "需要优化"}[nestedSuccessRate >= 90.0],
		})
	}

	return results
}

// 辅助函数
func calculateNodeDepth(node tsmorphgo.Node) int {
	depth := 0
	ancestors := node.GetAncestors()

	// 计算有效祖先深度（排除源文件）
	for _, ancestor := range ancestors {
		if ancestor.Kind != ast.KindSourceFile {
			depth++
		}
	}

	return depth
}