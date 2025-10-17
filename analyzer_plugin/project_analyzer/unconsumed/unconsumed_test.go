package unconsumed

import (
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

func TestUnconsumedFinder(t *testing.T) {
	// 准备一个绝对路径用于测试
	projectRoot, _ := filepath.Abs("/test-project")
	providerPath := filepath.Join(projectRoot, "src/provider.ts")
	consumerPath := filepath.Join(projectRoot, "src/consumer.ts")

	// 1. 准备测试数据
	mockParsingResult := &projectParser.ProjectParserResult{
		Js_Data: map[string]projectParser.JsFileParserResult{
			// provider.ts: 导出一个使用的变量和一个未使用的函数
			providerPath: {
				VariableDeclarations: []parser.VariableDeclaration{
					{
						Exported: true,
						Kind:     "const",
						Declarators: []*parser.VariableDeclarator{
							{Identifier: "usedVar"},
						},
						SourceLocation: &parser.SourceLocation{Start: parser.NodePosition{Line: 1}},
					},
					{
						Exported: true,
						Kind:     "function",
						Declarators: []*parser.VariableDeclarator{
							{Identifier: "unconsumedFunc"},
						},
						SourceLocation: &parser.SourceLocation{Start: parser.NodePosition{Line: 2}},
					},
				},
				// 默认导出，也是未使用的
				ExportAssignments: []parser.ExportAssignmentResult{
					{SourceLocation: &parser.SourceLocation{Start: parser.NodePosition{Line: 3}}},
				},
			},
			// consumer.ts: 只导入并使用了 usedVar
			consumerPath: {
				ImportDeclarations: []projectParser.ImportDeclarationResult{
					{
						Source: projectParser.SourceData{FilePath: providerPath},
						ImportModules: []projectParser.ImportModule{
							{Identifier: "usedVar", Type: "named"},
						},
					},
				},
			},
		},
	}

	// 2. 创建分析器和上下文
	analyzer := &Finder{}
	ctx := &projectanalyzer.ProjectContext{
		ParsingResult: mockParsingResult,
	}

	// 3. 执行分析
	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze() returned an unexpected error: %v", err)
	}

	// 4. 断言结果
	findResult, ok := result.(*Result)
	if !ok {
		t.Fatalf("Analyze() returned result of wrong type: got %T, want *Result", result)
	}

	// 检查统计数据
	if findResult.Stats.TotalFilesScanned != 2 {
		t.Errorf("Expected TotalFilesScanned to be 2, but got %d", findResult.Stats.TotalFilesScanned)
	}
	// 2个变量导出 + 1个默认导出 = 3个总导出
	if findResult.Stats.TotalExportsFound != 3 {
		t.Errorf("Expected TotalExportsFound to be 3, but got %d", findResult.Stats.TotalExportsFound)
	}
	// 1个未使用的函数 + 1个未使用的默认导出 = 2个未使用
	if findResult.Stats.UnconsumedExportsFound != 2 {
		t.Errorf("Expected UnconsumedExportsFound to be 2, but got %d", findResult.Stats.UnconsumedExportsFound)
	}

	// 检查找到的未使用导出的具体内容
	expectedFindings := []Finding{
		{FilePath: providerPath, ExportName: "unconsumedFunc", Line: 2, Kind: "function"},
		{FilePath: providerPath, ExportName: "default", Line: 3, Kind: "default"},
	}

	// 为了稳定比较，对两个切片都进行排序
	sort.Slice(findResult.Findings, func(i, j int) bool {
		return findResult.Findings[i].ExportName < findResult.Findings[j].ExportName
	})
	sort.Slice(expectedFindings, func(i, j int) bool {
		return expectedFindings[i].ExportName < expectedFindings[j].ExportName
	})

	if !reflect.DeepEqual(findResult.Findings, expectedFindings) {
		t.Errorf("Expected Findings to be %v, but got %v", expectedFindings, findResult.Findings)
	}
}
