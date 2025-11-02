// +build accuracy-validation

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// SymbolAccuracyTestCase 符号准确性测试用例
type SymbolAccuracyTestCase struct {
	Name        string                 `json:"name"`        // 测试用例名称
	Description string                 `json:"description"` // 测试用例描述
	Input       SymbolAccuracyInput   `json:"input"`       // 输入参数
	Expected    SymbolAccuracyExpected `json:"expected"`    // 期望结果
}

// SymbolAccuracyInput 符号准确性测试输入
type SymbolAccuracyInput struct {
	FilePath string `json:"filePath"` // 文件路径
	Line     int    `json:"line"`     // 行号
	Char     int    `json:"char"`     // 列号
	Symbol   string `json:"symbol"`   // 期望的符号名称（可选）
}

// SymbolAccuracyExpected 符号准确性期望结果
type SymbolAccuracyExpected struct {
	Name        string   `json:"name"`        // 期望的符号名称
	Kind        string   `json:"kind"`        // 期望的符号类型
	IsExported  bool     `json:"isExported"`  // 期望的导出状态
	Line        int      `json:"line"`        // 期望的行号
	Members     []string `json:"members"`     // 期望的成员列表（可选）
	Declaration  string   `json:"declaration"` // 期望的声明类型
}

// SymbolAccuracyResult 符号准确性测试结果
type SymbolAccuracyResult struct {
	TestCase  SymbolAccuracyTestCase `json:"testCase"`   // 测试用例
	Actual    SymbolAccuracyActual   `json:"actual"`      // 实际结果
	Success   bool                  `json:"success"`     // 是否成功
	Error     error                 `json:"error"`       // 错误信息
	Diff      SymbolAccuracyDiff    `json:"diff"`        // 差异详情
}

// SymbolAccuracyActual 符号准确性实际结果
type SymbolAccuracyActual struct {
	Name       string   `json:"name"`       // 实际的符号名称
	Kind       string   `json:"kind"`       // 实际的符号类型
	IsExported bool     `json:"isExported"` // 实际的导出状态
	Line       int      `json:"line"`       // 实际的行号
	Members    []string `json:"members"`    // 实际的成员列表
	Declaration string   `json:"declaration"` // 实际的声明类型
}

// SymbolAccuracyDiff 符号准确性差异
type SymbolAccuracyDiff struct {
	Name       *string `json:"name,omitempty"`       // 名称差异
	Kind       *string `json:"kind,omitempty"`       // 类型差异
	IsExported *bool   `json:"isExported,omitempty"` // 导出状态差异
	Line       *int    `json:"line,omitempty"`       // 行号差异
	Members    *string `json:"members,omitempty"`    // 成员列表差异
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run -tags accuracy-validation symbol-accuracy.go <项目路径>")
		os.Exit(1)
	}

	projectPath := os.Args[1]

	fmt.Println("🎯 符号 API 准确性验证")
	fmt.Println("================================")

	// 1. 加载测试用例
	testCases, err := loadSymbolAccuracyTestCases()
	if err != nil {
		fmt.Printf("❌ 加载测试用例失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 加载 %d 个测试用例\n", len(testCases))

	// 2. 创建 TSMorphGo 项目
	fmt.Println("\n🔧 创建项目...")
	config := tsmorphgo.ProjectConfig{
		RootPath:         projectPath,
		IgnorePatterns:   []string{"node_modules", "dist", "build", ".git"},
		TargetExtensions: []string{".ts", ".tsx"},
	}
	project := tsmorphgo.NewProject(config)
	defer project.Close()

	sourceFiles := project.GetSourceFiles()
	fmt.Printf("✅ 项目创建成功，发现 %d 个源文件\n", len(sourceFiles))

	if len(sourceFiles) == 0 {
		fmt.Println("❌ 项目中没有发现任何源文件")
		return
	}

	// 3. 执行准确性验证
	fmt.Println("\n🧪 执行准确性验证...")
	fmt.Println("================================")

	results := []SymbolAccuracyResult{}
	passedCount := 0
	failedCount := 0

	for i, testCase := range testCases {
		fmt.Printf("\n🔍 [%d/%d] 测试: %s\n", i+1, len(testCases), testCase.Name)
		fmt.Printf("   描述: %s\n", testCase.Description)
		fmt.Printf("   位置: %s:%d:%d\n", testCase.Input.FilePath, testCase.Input.Line, testCase.Input.Char)

		// 执行单个测试用例
		result := executeSymbolAccuracyTest(project, testCase)
		results = append(results, result)

		// 输出测试结果
		if result.Success {
			fmt.Printf("   ✅ 通过\n")
			passedCount++
		} else {
			fmt.Printf("   ❌ 失败\n")
			if result.Error != nil {
				fmt.Printf("      错误: %v\n", result.Error)
			}
			if result.Diff.Name != nil || result.Diff.Kind != nil || result.Diff.Line != nil {
				fmt.Printf("      差异: 名称=%v, 类型=%v, 行号=%v\n",
					result.Diff.Name, result.Diff.Kind, result.Diff.Line)
			}
			failedCount++
		}
	}

	// 4. 生成验证报告
	fmt.Println("\n📊 验证结果汇总")
	fmt.Println("================================")

	totalTests := len(testCases)
	successRate := float64(passedCount) / float64(totalTests) * 100

	fmt.Printf("   总测试数: %d\n", totalTests)
	fmt.Printf("   通过数量: %d\n", passedCount)
	fmt.Printf("   失败数量: %d\n", failedCount)
	fmt.Printf("   成功率: %.1f%%\n", successRate)

	// 5. 分析失败原因
	if failedCount > 0 {
		fmt.Println("\n🔍 失败原因分析:")
		fmt.Println("------------------------------")
		analyzeFailures(results)
	}

	// 6. 保存详细结果
	fmt.Println("\n💾 保存验证结果...")
	resultFile := "../../validation-results/symbol-accuracy-results.json"
	if err := saveSymbolAccuracyResults(results, resultFile); err != nil {
		fmt.Printf("❌ 保存结果失败: %v\n", err)
	} else {
		fmt.Printf("✅ 结果已保存到: %s\n", resultFile)
	}

	// 7. 最终结论
	fmt.Println("\n🎯 验证结论")
	fmt.Println("================================")

	if successRate >= 90.0 {
		fmt.Printf("🎉 符号 API 准确性验证通过！成功率 %.1f%%\n", successRate)
		fmt.Println("   API 准确性达到可接受水平")
	} else if successRate >= 70.0 {
		fmt.Printf("⚠️ 符号 API 准确性验证部分通过！成功率 %.1f%%\n", successRate)
		fmt.Println("   API 存在一些准确性问题，需要改进")
	} else {
		fmt.Printf("❌ 符号 API 准确性验证未通过！成功率 %.1f%%\n", successRate)
		fmt.Println("   API 准确性问题严重，需要重点关注")
	}
}

// loadSymbolAccuracyTestCases 加载符号准确性测试用例
func loadSymbolAccuracyTestCases() ([]SymbolAccuracyTestCase, error) {
	// 这里使用硬编码的测试用例，实际项目中可以从 JSON 文件加载
	return []SymbolAccuracyTestCase{
		{
			Name:        "User interface symbol",
			Description: "验证 User 接口的符号信息",
			Input: SymbolAccuracyInput{
				FilePath: "src/types.ts",
				Line:     8,
				Char:     1,
				Symbol:   "User",
			},
			Expected: SymbolAccuracyExpected{
				Name:       "User",
				Kind:       "interface",
				IsExported: true,
				Line:       8,
				Declaration: "InterfaceDeclaration",
			},
		},
		{
			Name:        "UserRole type alias symbol",
			Description: "验证 UserRole 类型别名的符号信息",
			Input: SymbolAccuracyInput{
				FilePath: "src/types.ts",
				Line:     29,
				Char:     1,
				Symbol:   "UserRole",
			},
			Expected: SymbolAccuracyExpected{
				Name:       "UserRole",
				Kind:       "typeAlias",
				IsExported: true,
				Line:       29,
				Declaration: "TypeAliasDeclaration",
			},
		},
		{
			Name:        "UserService class symbol",
			Description: "验证 UserService 类的符号信息",
			Input: SymbolAccuracyInput{
				FilePath: "src/services/api.ts",
				Line:     1,
				Char:     1,
				Symbol:   "UserService",
			},
			Expected: SymbolAccuracyExpected{
				Name:       "UserService",
				Kind:       "class",
				IsExported: true,
				Line:       1,
				Declaration: "ClassDeclaration",
			},
		},
	}, nil
}

// executeSymbolAccuracyTest 执行单个符号准确性测试
func executeSymbolAccuracyTest(project *tsmorphgo.Project, testCase SymbolAccuracyTestCase) SymbolAccuracyResult {
	result := SymbolAccuracyResult{
		TestCase: testCase,
	}

	// 在指定位置查找节点
	node := project.FindNodeAt(testCase.Input.FilePath, testCase.Input.Line, testCase.Input.Char)
	if node == nil {
		result.Success = false
		result.Error = fmt.Errorf("未找到指定位置的节点: %s:%d:%d",
			testCase.Input.FilePath, testCase.Input.Line, testCase.Input.Char)
		return result
	}

	// 获取节点的符号
	symbol, ok := tsmorphgo.GetSymbol(*node)
	if !ok {
		result.Success = false
		result.Error = fmt.Errorf("未找到节点的符号: %s", testCase.Input.Symbol)
		return result
	}

	// 提取符号的实际信息
	actualName := symbol.GetName()
	actualKind := getSymbolKindName(symbol)
	actualIsExported := symbol.IsExported()
	actualLine := node.GetStartLineNumber()
	actualDeclaration := getNodeKindName(*node)

	// 构建实际结果对象
	result.Actual = SymbolAccuracyActual{
		Name:       actualName,
		Kind:       actualKind,
		IsExported: actualIsExported,
		Line:       actualLine,
		Declaration: actualDeclaration,
	}

	// 验证准确性
	expected := testCase.Expected
	result.Diff = SymbolAccuracyDiff{}

	result.Success = true

	// 验证名称
	if actualName != expected.Name {
		result.Diff.Name = &expected.Name
		result.Success = false
	}

	// 验证类型
	if actualKind != expected.Kind {
		result.Diff.Kind = &expected.Kind
		result.Success = false
	}

	// 验证导出状态
	if actualIsExported != expected.IsExported {
		result.Diff.IsExported = &expected.IsExported
		result.Success = false
	}

	// 验证行号（允许一定的误差范围）
	lineDiff := actualLine - expected.Line
	if lineDiff < -1 || lineDiff > 1 {
		result.Diff.Line = &expected.Line
		result.Success = false
	}

	return result
}

// analyzeFailures 分析失败原因
func analyzeFailures(results []SymbolAccuracyResult) {
	nameErrors := 0
	kindErrors := 0
	exportedErrors := 0
	lineErrors := 0
	otherErrors := 0

	for _, result := range results {
		if result.Success {
			continue
		}

		if result.Error != nil {
			otherErrors++
			continue
		}

		if result.Diff.Name != nil {
			nameErrors++
		}
		if result.Diff.Kind != nil {
			kindErrors++
		}
		if result.Diff.IsExported != nil {
			exportedErrors++
		}
		if result.Diff.Line != nil {
			lineErrors++
		}
	}

	fmt.Printf("   名称错误: %d 次\n", nameErrors)
	fmt.Printf("   类型错误: %d 次\n", kindErrors)
	fmt.Printf("   导出状态错误: %d 次\n", exportedErrors)
	fmt.Printf("   行号错误: %d 次\n", lineErrors)
	fmt.Printf("   其他错误: %d 次\n", otherErrors)

	// 给出改进建议
	if nameErrors > 0 {
		fmt.Println("   💡 建议：检查符号名称提取逻辑")
	}
	if kindErrors > 0 {
		fmt.Println("   💡 建议：检查符号类型判断逻辑")
	}
	if exportedErrors > 0 {
		fmt.Println("   💡 建议：检查导出状态检测逻辑")
	}
	if lineErrors > 0 {
		fmt.Println("   💡 建议：检查位置计算和行号映射")
	}
}

// saveSymbolAccuracyResults 保存验证结果到文件
func saveSymbolAccuracyResults(results []SymbolAccuracyResult, filename string) error {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}

	// 序列化结果为 JSON
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(filename, data, 0644)
}

// getSymbolKindName 获取符号类型的人类可读名称
func getSymbolKindName(symbol *tsmorphgo.Symbol) string {
	switch {
	case symbol.IsFunction():
		return "function"
	case symbol.IsClass():
		return "class"
	case symbol.IsInterface():
		return "interface"
	case symbol.IsTypeAlias():
		return "typeAlias"
	case symbol.IsEnum():
		return "enum"
	case symbol.IsVariable():
		return "variable"
	case symbol.IsMethod():
		return "method"
	case symbol.IsConstructor():
		return "constructor"
	case symbol.IsAccessor():
		return "accessor"
	default:
		return "unknown"
	}
}

// getNodeKindName 获取节点类型的人类可读名称
func getNodeKindName(node tsmorphgo.Node) string {
	switch node.Kind {
	case ast.KindInterfaceDeclaration:
		return "InterfaceDeclaration"
	case ast.KindTypeAliasDeclaration:
		return "TypeAliasDeclaration"
	case ast.KindClassDeclaration:
		return "ClassDeclaration"
	case ast.KindFunctionDeclaration:
		return "FunctionDeclaration"
	case ast.KindVariableDeclaration:
		return "VariableDeclaration"
	case ast.KindEnumDeclaration:
		return "EnumDeclaration"
	default:
		return "Unknown"
	}
}