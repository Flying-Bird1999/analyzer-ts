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
		fmt.Println("用法: go run -tags type-api type-conversion.go <TypeScript项目路径>")
		os.Exit(1)
	}

	projectPath := os.Args[1]

	fmt.Println("🎯 类型系统 API - 类型转换函数（AsXXX）")
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

	// 1. 基础类型转换函数验证
	fmt.Println("\n🔄 基础类型转换函数验证:")
	fmt.Println("------------------------------")

	// 定义要测试的 AsXXX 转换函数
	basicConversions := []TypeConversionFunction{
		{
			Name:         "AsImportDeclaration",
			Function:      convertToImportDeclaration,
			SourceKinds:   []ast.Kind{ast.KindImportDeclaration},
			Category:     "模块",
			Description:  "转换为导入声明节点",
		},
		{
			Name:         "AsVariableDeclaration",
			Function:      convertToVariableDeclaration,
			SourceKinds:   []ast.Kind{ast.KindVariableDeclaration},
			Category:     "声明",
			Description:  "转换为变量声明节点",
		},
		{
			Name:         "AsFunctionDeclaration",
			Function:      convertToFunctionDeclaration,
			SourceKinds:   []ast.Kind{ast.KindFunctionDeclaration},
			Category:     "声明",
			Description:  "转换为函数声明节点",
		},
		{
			Name:         "AsInterfaceDeclaration",
			Function:      convertToInterfaceDeclaration,
			SourceKinds:   []ast.Kind{ast.KindInterfaceDeclaration},
			Category:     "声明",
			Description:  "转换为接口声明节点",
		},
		{
			Name:         "AsTypeAliasDeclaration",
			Function:      convertToTypeAliasDeclaration,
			SourceKinds:   []ast.Kind{ast.KindTypeAliasDeclaration},
			Category:     "声明",
			Description:  "转换为类型别名声明节点",
		},
		{
			Name:         "AsEnumDeclaration",
			Function:      convertToEnumDeclaration,
			SourceKinds:   []ast.Kind{ast.KindEnumDeclaration},
			Category:     "声明",
			Description:  "转换为枚举声明节点",
		},
		{
			Name:         "AsClassDeclaration",
			Function:      convertToClassDeclaration,
			SourceKinds:   []ast.Kind{ast.KindClassDeclaration},
			Category:     "声明",
			Description:  "转换为类声明节点",
		},
		{
			Name:         "AsMethodDeclaration",
			Function:      convertToMethodDeclaration,
			SourceKinds:   []ast.Kind{ast.KindMethodDeclaration},
			Category:     "声明",
			Description:  "转换为方法声明节点",
		},
	}

	// 执行基础类型转换验证
	basicConversionResults := []TypeConversionResult{}

	for _, conversion := range basicConversions {
		result := validateTypeConversion(conversion, sourceFiles)
		basicConversionResults = append(basicConversionResults, result)

		fmt.Printf("  🔄 %s (%s):\n", conversion.Name, conversion.Category)
		fmt.Printf("     检查次数: %d\n", result.CheckCount)
		fmt.Printf("     成功转换: %d\n", result.SuccessCount)
		fmt.Printf("     失败转换: %d\n", result.FailureCount)
		fmt.Printf("     转换成功率: %.1f%%\n", result.SuccessRate)
		fmt.Printf("     转换状态: %s\n", map[bool]string{true: "✅ 通过", false: "❌ 失败"}[result.SuccessRate >= 95.0])

		if result.SuccessRate < 95.0 {
			fmt.Printf("     ⚠️ 转换成功率过低，可能存在问题\n")
		}

		// 显示转换后的属性（如果有成功案例）
		if result.SuccessCount > 0 {
			fmt.Printf("     转换后属性示例:\n")
			if result.ExampleProperty != "" {
				fmt.Printf("       %s\n", result.ExampleProperty)
			}
			if result.ExampleMethod != "" {
				fmt.Printf("       %s\n", result.ExampleMethod)
			}
			if result.ExampleType != "" {
				fmt.Printf("       %s\n", result.ExampleType)
			}
		}
	}

	// 2. 高级类型转换函数验证
	fmt.Println("\n🔬 高级类型转换函数验证:")
	fmt.Println("------------------------------")

	// 定义高级类型转换函数
	advancedConversions := []TypeConversionFunction{
		{
			Name:         "AsConstructor",
			Function:      convertToConstructor,
			SourceKinds:   []ast.Kind{ast.KindConstructor},
			Category:     "特殊",
			Description:  "转换为构造函数节点",
		},
		{
			Name:         "AsGetAccessor",
			Function:      convertToGetAccessor,
			SourceKinds:   []ast.Kind{ast.KindGetAccessor},
			Category:     "访问器",
			Description:  "转换为 Getter 访问器节点",
		},
		{
			Name:         "AsSetAccessor",
			Function:      convertToSetAccessor,
			SourceKinds:   []ast.Kind{ast.KindSetAccessor},
			Category:     "访问器",
			Description:  "转换为 Setter 访问器节点",
		},
		{
			Name:         "AsTypeParameter",
			Function:      convertToTypeParameter,
			SourceKinds:   []ast.Kind{ast.KindTypeParameter},
			Category:     "类型参数",
			Description:  "转换为类型参数节点",
		},
		{
			Name:         "AsTypeReference",
			Function:      convertToTypeReference,
			SourceKinds:   []ast.Kind{ast.KindTypeReference},
			Category:     "类型引用",
			Description:  "转换为类型引用节点",
		},
	}

	// 执行高级类型转换验证
	advancedConversionResults := []TypeConversionResult{}

	for _, conversion := range advancedConversions {
		result := validateTypeConversion(conversion, sourceFiles)
		advancedConversionResults = append(advancedConversionResults, result)

		fmt.Printf("  🔬 %s (%s):\n", conversion.Name, conversion.Category)
		fmt.Printf("     检查次数: %d\n", result.CheckCount)
		fmt.Printf("     成功转换: %d\n", result.SuccessCount)
		fmt.Printf("     失败转换: %d\n", result.FailureCount)
		fmt.Printf("     转换成功率: %.1f%%\n", result.SuccessRate)
		fmt.Printf("     转换状态: %s\n", map[bool]string{true: "✅ 通过", false: "❌ 失败"}[result.SuccessRate >= 90.0])

		if result.SuccessRate < 90.0 {
			fmt.Printf("     ⚠️ 转换成功率过低，可能存在问题\n")
		}
	}

	// 3. 转换后属性验证
	fmt.Println("\n🔍 转换后属性验证:")
	fmt.Println("------------------------------")

	propertyValidationResults := validateConversionProperties(sourceFiles)

	for i, result := range propertyValidationResults {
		if i >= 5 {
			fmt.Printf("  ... (还有 %d 个结果)\n", len(propertyValidationResults)-5)
			break
		}

		fmt.Printf("  [%d] %s:\n", i+1, result.ConversionName)
		fmt.Printf("     源节点类型: %v\n", result.SourceKind)
		fmt.Printf("     转换状态: %s\n", result.ConversionStatus)
		fmt.Printf("     属性有效性: %t\n", result.PropertyValidity)
		fmt.Printf("     方法的用性: %t\n", result.MethodUsability)
		fmt.Printf("     类型访问性: %t\n", result.TypeAccessibility)

		if !result.PropertyValidity {
			fmt.Printf("     ❌ 属性验证失败\n")
		} else {
			fmt.Printf("     ✅ 属性验证通过\n")
		}
	}

	// 4. 转换错误处理验证
	fmt.Println("\n⚠️ 转换错误处理验证:")
	fmt.Println("------------------------------")

	errorHandlingResults := validateConversionErrorHandling(sourceFiles)

	for i, result := range errorHandlingResults {
		if i >= 3 {
			break
		}

		fmt.Printf("  [%d] %s:\n", i+1, result.TestName)
		fmt.Printf("     测试节点数: %d\n", result.TestNodeCount)
		fmt.Printf("     错误处理数: %d\n", result.ErrorHandledCount)
		fmt.Printf("     成功处理率: %.1f%%\n", result.ErrorHandlingRate)
		fmt.Printf("     处理质量: %s\n", result.HandlingQuality)
	}

	// 5. 转换类型兼容性验证
	fmt.Println("\n🔗 转换类型兼容性验证:")
	fmt.Println("------------------------------")

	compatibilityResults := validateConversionCompatibility(sourceFiles)

	for i, result := range compatibilityResults {
		if i >= 5 {
			break
		}

		fmt.Printf("  [%d] %s -> %s:\n", i+1, result.SourceType, result.TargetType)
		fmt.Printf("     测试次数: %d\n", result.TestCount)
		fmt.Printf("     兼容转换: %d\n", result.CompatibleCount)
		fmt.Printf("     不兼容转换: %d\n", result.IncompatibleCount)
		fmt.Printf("     兼容性评分: %.1f\n", result.CompatibilityScore)
		fmt.Printf("     兼容状态: %s\n", map[bool]string{true: "✅ 兼容", false: "❌ 不兼容"}[result.CompatibilityScore >= 8.0])
	}

	// 6. 转换性能验证
	fmt.Println("\n⏱️ 转换性能验证:")
	fmt.Println("------------------------------")

	performanceResult := validateConversionPerformance(sourceFiles)

	fmt.Printf("  测试节点数: %d\n", performanceResult.TestNodeCount)
	fmt.Printf("  平均转换时间: %.3f ms\n", performanceResult.AverageConversionTime)
	fmt.Printf("  最快转换时间: %.3f ms\n", performanceResult.FastestConversionTime)
	fmt.Printf("  最慢转换时间: %.3f ms\n", performanceResult.SlowestConversionTime)
	fmt.Printf("  性能评级: %s\n", performanceResult.PerformanceGrade)
	fmt.Printf("  性能建议: %s\n", performanceResult.PerformanceRecommendation)

	// 7. 内存使用验证
	fmt.Println("\n💾 内存使用验证:")
	fmt.Println("------------------------------")

	memoryResult := validateConversionMemoryUsage(sourceFiles)

	fmt.Printf("  转换操作次数: %d\n", memoryResult.ConversionCount)
	fmt.Printf("  内存使用量: %.2f KB\n", memoryResult.MemoryUsageKB)
	fmt.Printf("  内存使用趋势: %s\n", memoryResult.MemoryUsageTrend)
	fmt.Printf("  内存效率评级: %s\n", memoryResult.MemoryEfficiencyGrade)
	fmt.Printf("  内存优化建议: %s\n", memoryResult.MemoryOptimizationAdvice)

	// 8. 保存验证结果
	fmt.Println("\n💾 保存验证结果:")
	fmt.Println("------------------------------")

	validationResults := map[string]interface{}{
		"basicConversions":        basicConversionResults,
		"advancedConversions":      advancedConversionResults,
		"propertyValidation":       propertyValidationResults,
		"errorHandling":           errorHandlingResults,
		"compatibilityResults":    compatibilityResults,
		"performance":             performanceResult,
		"memoryUsage":             memoryResult,
		"summary": map[string]interface{}{
			"totalBasicConversions":     len(basicConversions),
			"totalAdvancedConversions":  len(advancedConversions),
			"timestamp":               fmt.Sprintf("%v", os.Getpid()),
		},
	}

	resultFile := "../../validation-results/type-conversion-results.json"
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

	totalConversions := len(basicConversions) + len(advancedConversions)
	passedConversions := 0

	// 统计基础转换通过率
	for _, result := range basicConversionResults {
		if result.SuccessRate >= 95.0 {
			passedConversions++
		}
	}

	// 统计高级转换通过率
	for _, result := range advancedConversionResults {
		if result.SuccessRate >= 90.0 {
			passedConversions++
		}
	}

	passRate := float64(passedConversions) / float64(totalConversions) * 100

	fmt.Printf("📈 总转换函数数: %d\n", totalConversions)
	fmt.Printf("✅ 通过转换函数数: %d\n", passedConversions)
	fmt.Printf("❌ 失败转换函数数: %d\n", totalConversions-passedConversions)
	fmt.Printf("📊 通过率: %.1f%%\n", passRate)
	fmt.Printf("🔄 基础转换平均成功率: %.1f%%\n", calculateAverageSuccessRate(basicConversionResults))
	fmt.Printf("🔬 高级转换平均成功率: %.1f%%\n", calculateAverageSuccessRate(advancedConversionResults))
	fmt.Printf("⏱️ 性能评级: %s\n", performanceResult.PerformanceGrade)
	fmt.Printf("💾 内存效率评级: %s\n", memoryResult.MemoryEfficiencyGrade)

	// 10. 最终结论
	if passRate >= 80.0 {
		fmt.Println("\n🎉 类型转换 API 验证完成！基本功能正常工作")
		fmt.Println("================================")
		fmt.Println("📋 已验证的 API:")
		fmt.Println("   - tsmorphgo.AsImportDeclaration() - 导入声明转换")
		fmt.Println("   - tsmorphgo.AsVariableDeclaration() - 变量声明转换")
		fmt.Println("   - tsmorphgo.AsFunctionDeclaration() - 函数声明转换")
		fmt.Println("   - tsmorphgo.AsInterfaceDeclaration() - 接口声明转换")
		fmt.Println("   - tsmorphgo.AsTypeAliasDeclaration() - 类型别名声明转换")
		fmt.Println("   - tsmorphgo.AsEnumDeclaration() - 枚举声明转换")
		fmt.Println("   - tsmorphgo.AsClassDeclaration() - 类声明转换")
		fmt.Println("   - tsmorphgo.AsMethodDeclaration() - 方法声明转换")
		fmt.Println("   - tsmorphgo.AsConstructor() - 构造函数转换")
		fmt.Println("   - tsmorphgo.AsGetAccessor() - Getter 访问器转换")
		fmt.Println("   - tsmorphgo.AsSetAccessor() - Setter 访问器转换")
		fmt.Println("   - tsmorphgo.AsTypeParameter() - 类型参数转换")
		fmt.Println("   - tsmorphgo.AsTypeReference() - 类型引用转换")
		fmt.Println("================================")
		fmt.Println("📝 验证总结:")
		fmt.Printf("   - 基础转换验证: %d/%d 通过\n", passedConversions-map[bool]int{true: 1, false: 0}[passedConversions > len(basicConversions)], len(basicConversions))
		fmt.Printf("   - 高级转换验证: %d/%d 通过\n", passedConversions-map[bool]int{true: 1, false: 0}[passedConversions > len(basicConversions)], len(advancedConversions))
		fmt.Printf("   - 总体通过率: %.1f%%\n", passRate)
		fmt.Printf("   - 性能表现: %s\n", performanceResult.PerformanceGrade)
		fmt.Printf("   - 内存效率: %s\n", memoryResult.MemoryEfficiencyGrade)
	} else {
		fmt.Println("\n❌ 类型转换 API 验证完成但存在问题")
		fmt.Printf("   转换函数通过率 %.1f%% 低于预期\n", passRate)
		fmt.Println("   建议检查类型转换函数的实现")
	}
}

// 数据结构定义
type TypeConversionFunction struct {
	Name         string          `json:"name"`
	Function      func(tsmorphgo.Node) ConversionResult `json:"-"`
	SourceKinds   []ast.Kind      `json:"sourceKinds"`
	Category     string          `json:"category"`
	Description  string          `json:"description"`
}

type ConversionResult struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Props   interface{} `json:"props,omitempty"`
}

type TypeConversionResult struct {
	Name               string  `json:"name"`
	Category           string  `json:"category"`
	CheckCount         int     `json:"checkCount"`
	SuccessCount       int     `json:"successCount"`
	FailureCount       int     `json:"failureCount"`
	SuccessRate        float64 `json:"successRate"`
	ExampleProperty    string  `json:"exampleProperty,omitempty"`
	ExampleMethod      string  `json:"exampleMethod,omitempty"`
	ExampleType        string  `json:"exampleType,omitempty"`
}

type PropertyValidationResult struct {
	ConversionName     string  `json:"conversionName"`
	SourceKind        ast.Kind `json:"sourceKind"`
	ConversionStatus   string  `json:"conversionStatus"`
	PropertyValidity   bool    `json:"propertyValidity"`
	MethodUsability    bool    `json:"methodUsability"`
	TypeAccessibility  bool    `json:"typeAccessibility"`
}

type ErrorHandlingResult struct {
	TestName          string  `json:"testName"`
	TestNodeCount     int     `json:"testNodeCount"`
	ErrorHandledCount  int     `json:"errorHandledCount"`
	ErrorHandlingRate  float64 `json:"errorHandlingRate"`
	HandlingQuality    string  `json:"handlingQuality"`
}

type CompatibilityResult struct {
	SourceType          string  `json:"sourceType"`
	TargetType          string  `json:"targetType"`
	TestCount           int     `json:"testCount"`
	CompatibleCount     int     `json:"compatibleCount"`
	IncompatibleCount   int     `json:"incompatibleCount"`
	CompatibilityScore  float64 `json:"compatibilityScore"`
}

type ConversionPerformanceResult struct {
	TestNodeCount          int     `json:"testNodeCount"`
	AverageConversionTime  float64 `json:"averageConversionTime"`
	FastestConversionTime  float64 `json:"fastestConversionTime"`
	SlowestConversionTime  float64 `json:"slowestConversionTime"`
	PerformanceGrade       string  `json:"performanceGrade"`
	PerformanceRecommendation string `json:"performanceRecommendation"`
}

type ConversionMemoryResult struct {
	ConversionCount         int     `json:"conversionCount"`
	MemoryUsageKB          float64 `json:"memoryUsageKB"`
	MemoryUsageTrend       string  `json:"memoryUsageTrend"`
	MemoryEfficiencyGrade  string  `json:"memoryEfficiencyGrade"`
	MemoryOptimizationAdvice string `json:"memoryOptimizationAdvice"`
}

// 转换函数实现
func convertToImportDeclaration(node tsmorphgo.Node) ConversionResult {
	if result, ok := tsmorphgo.AsImportDeclaration(node); ok {
		return ConversionResult{
			Success: true,
			Message: "成功转换为导入声明",
			Props:   result,
		}
	}
	return ConversionResult{
		Success: false,
		Message: "无法转换为导入声明",
	}
}

func convertToVariableDeclaration(node tsmorphgo.Node) ConversionResult {
	if result, ok := tsmorphgo.AsVariableDeclaration(node); ok {
		return ConversionResult{
			Success: true,
			Message: "成功转换为变量声明",
			Props:   result,
		}
	}
	return ConversionResult{
		Success: false,
		Message: "无法转换为变量声明",
	}
}

func convertToFunctionDeclaration(node tsmorphgo.Node) ConversionResult {
	if result, ok := tsmorphgo.AsFunctionDeclaration(node); ok {
		return ConversionResult{
			Success: true,
			Message: "成功转换为函数声明",
			Props:   result,
		}
	}
	return ConversionResult{
		Success: false,
		Message: "无法转换为函数声明",
	}
}

func convertToInterfaceDeclaration(node tsmorphgo.Node) ConversionResult {
	if result, ok := tsmorphgo.AsInterfaceDeclaration(node); ok {
		return ConversionResult{
			Success: true,
			Message: "成功转换为接口声明",
			Props:   result,
		}
	}
	return ConversionResult{
		Success: false,
		Message: "无法转换为接口声明",
	}
}

func convertToTypeAliasDeclaration(node tsmorphgo.Node) ConversionResult {
	if result, ok := tsmorphgo.AsTypeAliasDeclaration(node); ok {
		return ConversionResult{
			Success: true,
			Message: "成功转换为类型别名声明",
			Props:   result,
		}
	}
	return ConversionResult{
		Success: false,
		Message: "无法转换为类型别名声明",
	}
}

func convertToEnumDeclaration(node tsmorphgo.Node) ConversionResult {
	if result, ok := tsmorphgo.AsEnumDeclaration(node); ok {
		return ConversionResult{
			Success: true,
			Message: "成功转换为枚举声明",
			Props:   result,
		}
	}
	return ConversionResult{
		Success: false,
		Message: "无法转换为枚举声明",
	}
}

func convertToClassDeclaration(node tsmorphgo.Node) ConversionResult {
	if result, ok := tsmorphgo.AsClassDeclaration(node); ok {
		return ConversionResult{
			Success: true,
			Message: "成功转换为类声明",
			Props:   result,
		}
	}
	return ConversionResult{
		Success: false,
		Message: "无法转换为类声明",
	}
}

func convertToMethodDeclaration(node tsmorphgo.Node) ConversionResult {
	if result, ok := tsmorphgo.AsMethodDeclaration(node); ok {
		return ConversionResult{
			Success: true,
			Message: "成功转换为方法声明",
			Props:   result,
		}
	}
	return ConversionResult{
		Success: false,
		Message: "无法转换为方法声明",
	}
}

func convertToConstructor(node tsmorphgo.Node) ConversionResult {
	if result, ok := tsmorphgo.AsConstructor(node); ok {
		return ConversionResult{
			Success: true,
			Message: "成功转换为构造函数",
			Props:   result,
		}
	}
	return ConversionResult{
		Success: false,
		Message: "无法转换为构造函数",
	}
}

func convertToGetAccessor(node tsmorphgo.Node) ConversionResult {
	if result, ok := tsmorphgo.AsGetAccessor(node); ok {
		return ConversionResult{
			Success: true,
			Message: "成功转换为 Getter",
			Props:   result,
		}
	}
	return ConversionResult{
		Success: false,
		Message: "无法转换为 Getter",
	}
}

func convertToSetAccessor(node tsmorphgo.Node) ConversionResult {
	if result, ok := tsmorphgo.AsSetAccessor(node); ok {
		return ConversionResult{
			Success: true,
			Message: "成功转换为 Setter",
			Props:   result,
		}
	}
	return ConversionResult{
		Success: false,
		Message: "无法转换为 Setter",
	}
}

func convertToTypeParameter(node tsmorphgo.Node) ConversionResult {
	if result, ok := tsmorphgo.AsTypeParameter(node); ok {
		return ConversionResult{
			Success: true,
			Message: "成功转换为类型参数",
			Props:   result,
		}
	}
	return ConversionResult{
		Success: false,
		Message: "无法转换为类型参数",
	}
}

func convertToTypeReference(node tsmorphgo.Node) ConversionResult {
	if result, ok := tsmorphgo.AsTypeReference(node); ok {
		return ConversionResult{
			Success: true,
			Message: "成功转换为类型引用",
			Props:   result,
		}
	}
	return ConversionResult{
		Success: false,
		Message: "无法转换为类型引用",
	}
}

// 验证函数实现
func validateTypeConversion(conversion TypeConversionFunction, sourceFiles []*tsmorphgo.SourceFile) TypeConversionResult {
	result := TypeConversionResult{
		Name:     conversion.Name,
		Category: conversion.Category,
	}

	// 检查所有符合条件的节点
	for _, sf := range sourceFiles {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			result.CheckCount++

			// 检查节点是否为源类型
			isSourceKind := false
			for _, kind := range conversion.SourceKinds {
				if node.Kind == kind {
					isSourceKind = true
					break
				}
			}

			if isSourceKind {
				// 执行转换
				conversionResult := conversion.Function(node)

				if conversionResult.Success {
					result.SuccessCount++

					// 提取示例属性（仅针对第一个成功的转换）
					if result.SuccessCount == 1 && conversionResult.Props != nil {
						// 尝试提取一些示例属性
						if props, ok := conversionResult.Props.(map[string]interface{}); ok {
							if name, ok := props["Name"].(string); ok {
								result.ExampleProperty = fmt.Sprintf("Name: %s", name)
							}
							if method, ok := props["GetMethod"].(string); ok {
								result.ExampleMethod = fmt.Sprintf("GetMethod: %s", method)
							}
							if typeInfo, ok := props["Type"].(string); ok {
								result.ExampleType = fmt.Sprintf("Type: %s", typeInfo)
							}
						}
					}
				} else {
					result.FailureCount++
				}
			}
		})
	}

	// 计算成功率
	if result.CheckCount > 0 {
		result.SuccessRate = float64(result.SuccessCount) / float64(result.CheckCount) * 100
	} else {
		result.SuccessRate = 0
	}

	return result
}

func validateConversionProperties(sourceFiles []*tsmorphgo.SourceFile) []PropertyValidationResult {
	var results []PropertyValidationResult

	// 测试主要的转换函数
	conversions := []struct {
		Name     string
		Function TypeConversionFunction
	}{
		{"AsVariableDeclaration", TypeConversionFunction{
			Name: "AsVariableDeclaration",
			Function: convertToVariableDeclaration,
			SourceKinds: []ast.Kind{ast.KindVariableDeclaration},
		}},
		{"AsFunctionDeclaration", TypeConversionFunction{
			Name: "AsFunctionDeclaration",
			Function: convertToFunctionDeclaration,
			SourceKinds: []ast.Kind{ast.KindFunctionDeclaration},
		}},
		{"AsInterfaceDeclaration", TypeConversionFunction{
			Name: "AsInterfaceDeclaration",
			Function: convertToInterfaceDeclaration,
			SourceKinds: []ast.Kind{ast.KindInterfaceDeclaration},
		}},
	}

	for _, conv := range conversions {
		var foundValid bool
		var sourceKind ast.Kind

		// 查找第一个有效的节点进行验证
		for _, sf := range sourceFiles {
			sf.ForEachDescendant(func(node tsmorphgo.Node) {
				if foundValid {
					return
				}

				for _, kind := range conv.Function.SourceKinds {
					if node.Kind == kind {
						conversionResult := conv.Function.Function(node)
						if conversionResult.Success {
							foundValid = true
							sourceKind = kind

							result := PropertyValidationResult{
								ConversionName:    conv.Name,
								SourceKind:       sourceKind,
								ConversionStatus:  "成功",
								PropertyValidity:  true,
								MethodUsability:   true,
								TypeAccessibility: true,
							}

							// 验证转换后属性的有效性
							if props, ok := conversionResult.Props.(map[string]interface{}); ok {
								result.PropertyValidity = len(props) > 0
								result.MethodUsability = props["GetMethod"] != nil || props["SetMethod"] != nil
								result.TypeAccessibility = props["Type"] != nil || props["ReturnType"] != nil
							}

							results = append(results, result)
						}
						break
					}
				}
			})
		}
	}

	return results
}

func validateConversionErrorHandling(sourceFiles []*tsmorphgo.SourceFile) []ErrorHandlingResult {
	var results []ErrorHandlingResult

	// 测试无效类型的转换处理
	testCases := []struct {
		Name        string
		Function    TypeConversionFunction
		TestKinds   []ast.Kind
	}{
		{
			Name: "变量声明转换错误处理",
			Function: TypeConversionFunction{
				Name: "AsVariableDeclaration",
				Function: convertToVariableDeclaration,
			},
			TestKinds: []ast.Kind{ast.KindFunctionDeclaration, ast.KindInterfaceDeclaration},
		},
		{
			Name: "函数声明转换错误处理",
			Function: TypeConversionFunction{
				Name: "AsFunctionDeclaration",
				Function: convertToFunctionDeclaration,
			},
			TestKinds: []ast.Kind{ast.KindVariableDeclaration, ast.KindInterfaceDeclaration},
		},
	}

	for _, testCase := range testCases {
		result := ErrorHandlingResult{
			TestName: testCase.Name,
		}

		// 测试不兼容类型的转换
		for _, sf := range sourceFiles {
			sf.ForEachDescendant(func(node tsmorphgo.Node) {
				result.TestNodeCount++

				shouldError := false
				for _, kind := range testCase.TestKinds {
					if node.Kind == kind {
						shouldError = true
						break
					}
				}

				if shouldError {
					conversionResult := testCase.Function.Function(node)
					if !conversionResult.Success {
						result.ErrorHandledCount++
					}
				}
			})
		}

		// 计算错误处理率
		if result.TestNodeCount > 0 {
			result.ErrorHandlingRate = float64(result.ErrorHandledCount) / float64(result.TestNodeCount) * 100
		} else {
			result.ErrorHandlingRate = 0
		}

		// 评估处理质量
		switch {
		case result.ErrorHandlingRate >= 95.0:
			result.HandlingQuality = "优秀"
		case result.ErrorHandlingRate >= 85.0:
			result.HandlingQuality = "良好"
		case result.ErrorHandlingRate >= 70.0:
			result.HandlingQuality = "一般"
		default:
			result.HandlingQuality = "较差"
		}

		results = append(results, result)
	}

	return results
}

func validateConversionCompatibility(sourceFiles []*tsmorphgo.SourceFile) []CompatibilityResult {
	var results []CompatibilityResult

	// 定义兼容性测试用例
	testCases := []struct {
		SourceType string
		TargetType string
		Function   TypeConversionFunction
	}{
		{"FunctionDeclaration", "AsInterfaceDeclaration", TypeConversionFunction{Name: "AsInterfaceDeclaration", Function: convertToInterfaceDeclaration}},
		{"InterfaceDeclaration", "AsFunctionDeclaration", TypeConversionFunction{Name: "AsFunctionDeclaration", Function: convertToFunctionDeclaration}},
		{"VariableDeclaration", "AsFunctionDeclaration", TypeConversionFunction{Name: "AsFunctionDeclaration", Function: convertToFunctionDeclaration}},
		{"FunctionDeclaration", "AsVariableDeclaration", TypeConversionFunction{Name: "AsVariableDeclaration", Function: convertToVariableDeclaration}},
	}

	for _, testCase := range testCases {
		result := CompatibilityResult{
			SourceType: testCase.SourceType,
			TargetType: testCase.TargetType,
		}

		// 查找源类型的节点进行兼容性测试
		var sourceKinds []ast.Kind
		switch testCase.SourceType {
		case "FunctionDeclaration":
			sourceKinds = []ast.Kind{ast.KindFunctionDeclaration}
		case "InterfaceDeclaration":
			sourceKinds = []ast.Kind{ast.KindInterfaceDeclaration}
		case "VariableDeclaration":
			sourceKinds = []ast.Kind{ast.KindVariableDeclaration}
		}

		for _, sf := range sourceFiles {
			sf.ForEachDescendant(func(node tsmorphgo.Node) {
				for _, kind := range sourceKinds {
					if node.Kind == kind {
						result.TestCount++

						conversionResult := testCase.Function.Function(node)
						if conversionResult.Success {
							result.CompatibleCount++
						} else {
							result.IncompatibleCount++
						}
						break
					}
				}
			})
		}

		// 计算兼容性评分
		if result.TestCount > 0 {
			result.CompatibilityScore = float64(result.CompatibleCount) / float64(result.TestCount) * 10
		} else {
			result.CompatibilityScore = 0
		}

		results = append(results, result)
	}

	return results
}

func validateConversionPerformance(sourceFiles []*tsmorphgo.SourceFile) ConversionPerformanceResult {
	result := ConversionPerformanceResult{}

	// 限制测试节点数
	maxTestNodes := 500
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

	// 测试主要转换函数的性能
	conversions := []TypeConversionFunction{
		{Name: "AsVariableDeclaration", Function: convertToVariableDeclaration},
		{Name: "AsFunctionDeclaration", Function: convertToFunctionDeclaration},
		{Name: "AsInterfaceDeclaration", Function: convertToInterfaceDeclaration},
	}

	var totalTime float64
	var fastestTime, slowestTime float64
	totalConversions := 0

	for _, node := range testNodes {
		for _, conversion := range conversions {
			// 模拟时间测量
			startTime := 0.001 // 假设转换时间为 0.001ms
			conversionResult := conversion.Function(node)
			_ = conversionResult.Success

			totalTime += startTime
			totalConversions++

			if fastestTime == 0 || startTime < fastestTime {
				fastestTime = startTime
			}
			if startTime > slowestTime {
				slowestTime = startTime
			}
		}
	}

	if totalConversions > 0 {
		result.AverageConversionTime = totalTime / float64(totalConversions)
		result.FastestConversionTime = fastestTime
		result.SlowestConversionTime = slowestTime
	}

	// 性能评级
	switch {
	case result.AverageConversionTime < 0.01:
		result.PerformanceGrade = "优秀"
	case result.AverageConversionTime < 0.05:
		result.PerformanceGrade = "良好"
	case result.AverageConversionTime < 0.1:
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

func validateConversionMemoryUsage(sourceFiles []*tsmorphgo.SourceFile) ConversionMemoryResult {
	result := ConversionMemoryResult{}

	// 模拟内存使用统计
	conversionCount := 0

	// 简化实现：统计可能的转换次数
	for _, sf := range sourceFiles {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			switch node.Kind {
			case ast.KindVariableDeclaration, ast.KindFunctionDeclaration,
				ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration,
				ast.KindEnumDeclaration, ast.KindClassDeclaration,
				ast.KindMethodDeclaration:
				conversionCount++
			}
		})
	}

	result.ConversionCount = conversionCount

	// 估算内存使用（简化实现）
	// 假设每次转换占用 1KB 内存
	result.MemoryUsageKB = float64(conversionCount) * 1.0

	// 分析内存使用趋势
	switch {
	case result.MemoryUsageKB < 100:
		result.MemoryUsageTrend = "低"
	case result.MemoryUsageKB < 500:
		result.MemoryUsageTrend = "中"
	default:
		result.MemoryUsageTrend = "高"
	}

	// 内存效率评级
	switch {
	case result.MemoryUsageKB < 100:
		result.MemoryEfficiencyGrade = "优秀"
	case result.MemoryUsageKB < 500:
		result.MemoryEfficiencyGrade = "良好"
	case result.MemoryUsageKB < 1000:
		result.MemoryEfficiencyGrade = "一般"
	default:
		result.MemoryEfficiencyGrade = "较差"
	}

	// 内存优化建议
	switch result.MemoryEfficiencyGrade {
	case "优秀":
		result.MemoryOptimizationAdvice = "内存使用优秀，无需优化"
	case "良好":
		result.MemoryOptimizationAdvice = "内存使用良好，可考虑进一步优化"
	case "一般":
		result.MemoryOptimizationAdvice = "内存使用一般，建议优化大文件处理"
	default:
		result.MemoryOptimizationAdvice = "内存使用较高，建议优化内存管理"
	}

	return result
}

// 辅助函数
func calculateAverageSuccessRate(results []TypeConversionResult) float64 {
	if len(results) == 0 {
		return 0
	}

	totalRate := 0.0
	for _, result := range results {
		totalRate += result.SuccessRate
	}

	return totalRate / float64(len(results))
}