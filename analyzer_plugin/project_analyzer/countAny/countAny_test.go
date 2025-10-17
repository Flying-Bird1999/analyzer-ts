
package countany

import (
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

func TestCountAnyAnalyzer(t *testing.T) {
	// 1. 准备测试数据
	mockParsingResult := &projectParser.ProjectParserResult{
		Js_Data: map[string]projectParser.JsFileParserResult{
			"file1.ts": {
				ExtractedNodes: parser.ExtractedNodes{
					AnyDeclarations: []parser.AnyInfo{
						{SourceLocation: parser.SourceLocation{}, Raw: "const a: any = 1;"},
						{SourceLocation: parser.SourceLocation{}, Raw: "let b: any;"},
					},
				},
			},
			"file2.ts": {
				ExtractedNodes: parser.ExtractedNodes{
					AnyDeclarations: []parser.AnyInfo{
						{SourceLocation: parser.SourceLocation{}, Raw: "function fn(p: any) {}"},
					},
				},
			},
			"file3.ts": {
				ExtractedNodes: parser.ExtractedNodes{
					AnyDeclarations: []parser.AnyInfo{}, // No 'any' here
				},
			},
		},
	}

	// 2. 创建分析器和上下文
	analyzer := &Counter{}
	ctx := &projectanalyzer.ProjectContext{
		ParsingResult: mockParsingResult,
	}

	// 3. 执行分析
	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze() returned an unexpected error: %v", err)
	}

	// 4. 断言结果
	countAnyResult, ok := result.(*CountAnyResult)
	if !ok {
		t.Fatalf("Analyze() returned result of wrong type: got %T, want *CountAnyResult", result)
	}

	// 检查解析的文件总数
	if countAnyResult.FilesParsed != 3 {
		t.Errorf("Expected FilesParsed to be 3, but got %d", countAnyResult.FilesParsed)
	}

	// 检查 'any' 的总数
	if countAnyResult.TotalAnyCount != 3 {
		t.Errorf("Expected TotalAnyCount to be 3, but got %d", countAnyResult.TotalAnyCount)
	}

	// 检查包含 'any' 的文件数量
	if len(countAnyResult.FileCounts) != 2 {
		t.Errorf("Expected FileCounts to have length 2, but got %d", len(countAnyResult.FileCounts))
	}

	// 检查摘要信息
	expectedSummary := "扫描文件 3 个，共发现 3 处 'any' 类型使用。"
	if summary := result.Summary(); summary != expectedSummary {
		t.Errorf("Expected Summary() to be '%s', but got '%s'", expectedSummary, summary)
	}

	// (可选) 检查具体某个文件的 'any' 数量
	for _, fc := range countAnyResult.FileCounts {
		if fc.FilePath == "file1.ts" && fc.AnyCount != 2 {
			t.Errorf("Expected file1.ts to have 2 'any' declarations, but got %d", fc.AnyCount)
		}
		if fc.FilePath == "file2.ts" && fc.AnyCount != 1 {
			t.Errorf("Expected file2.ts to have 1 'any' declaration, but got %d", fc.AnyCount)
		}
	}
}
