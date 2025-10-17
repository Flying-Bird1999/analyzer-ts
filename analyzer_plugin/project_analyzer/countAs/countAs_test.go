package countas

import (
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

func TestCountAsAnalyzer(t *testing.T) {
	// 1. 准备测试数据
	mockParsingResult := &projectParser.ProjectParserResult{
		Js_Data: map[string]projectParser.JsFileParserResult{
			"file1.ts": {
				ExtractedNodes: parser.ExtractedNodes{
					AsExpressions: []parser.AsExpression{
						{SourceLocation: parser.SourceLocation{}, Raw: "const a = 1 as number;"},
						{SourceLocation: parser.SourceLocation{}, Raw: "<string>b;"},
					},
				},
			},
			"file2.ts": {
				ExtractedNodes: parser.ExtractedNodes{
					AsExpressions: []parser.AsExpression{
						{SourceLocation: parser.SourceLocation{}, Raw: "c as any;"},
					},
				},
			},
			"file3.ts": {
				ExtractedNodes: parser.ExtractedNodes{
					AsExpressions: []parser.AsExpression{}, // No 'as' expressions here
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
	countAsResult, ok := result.(*CountAsResult)
	if !ok {
		t.Fatalf("Analyze() returned result of wrong type: got %T, want *CountAsResult", result)
	}

	// 检查解析的文件总数
	if countAsResult.FilesParsed != 3 {
		t.Errorf("Expected FilesParsed to be 3, but got %d", countAsResult.FilesParsed)
	}

	// 检查 'as' 的总数
	if countAsResult.TotalAsCount != 3 {
		t.Errorf("Expected TotalAsCount to be 3, but got %d", countAsResult.TotalAsCount)
	}

	// 检查包含 'as' 的文件数量
	if len(countAsResult.FileCounts) != 2 {
		t.Errorf("Expected FileCounts to have length 2, but got %d", len(countAsResult.FileCounts))
	}

	// 检查摘要信息
	expectedSummary := "扫描文件 3 个，共发现 3 处 'as' 类型断言使用。"
	if summary := result.Summary(); summary != expectedSummary {
		t.Errorf("Expected Summary() to be '%s', but got '%s'", expectedSummary, summary)
	}
}
