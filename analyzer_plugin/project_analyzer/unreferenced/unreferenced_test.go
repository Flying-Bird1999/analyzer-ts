package unreferenced

import (
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

func TestUnreferencedFinder(t *testing.T) {
	// 准备一个绝对路径用于测试，以模拟真实环境
	projectRoot, _ := filepath.Abs("/test-project")
	entryPath := filepath.Join(projectRoot, "src/entry.ts")
	referencedPath := filepath.Join(projectRoot, "src/referenced.ts")
	unreferencedPath := filepath.Join(projectRoot, "src/unreferenced.ts")
	suspiciousPath := filepath.Join(projectRoot, "src/config.ts") // 可疑文件

	// 1. 准备测试数据
	mockParsingResult := &projectParser.ProjectParserResult{
		Js_Data: map[string]projectParser.JsFileParserResult{
			// 入口文件，引用了 referenced.ts
			entryPath: {
				ImportDeclarations: []projectParser.ImportDeclarationResult{
					{
						Source: projectParser.SourceData{FilePath: referencedPath},
					},
				},
			},
			// 被引用的文件
			referencedPath: {},
			// 未被引用的文件
			unreferencedPath: {},
			// 可疑的未引用文件
			suspiciousPath: {},
		},
	}

	// 2. 创建分析器和上下文
	analyzer := &Finder{}
	ctx := &projectanalyzer.ProjectContext{
		ProjectRoot:   projectRoot,
		ParsingResult: mockParsingResult,
	}

	// 3. 配置分析器，指定入口文件
	params := map[string]string{"entrypoint": entryPath}
	if err := analyzer.Configure(params); err != nil {
		t.Fatalf("Configure() failed: %v", err)
	}

	// 4. 执行分析
	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze() returned an unexpected error: %v", err)
	}

	// 5. 断言结果
	findResult, ok := result.(*FindUnreferencedFilesResult)
	if !ok {
		t.Fatalf("Analyze() returned result of wrong type: got %T, want *FindUnreferencedFilesResult", result)
	}

	// 检查真正未引用的文件
	expectedTrulyUnreferenced := []string{unreferencedPath}
	sort.Strings(findResult.TrulyUnreferencedFiles)
	sort.Strings(expectedTrulyUnreferenced)
	if !reflect.DeepEqual(findResult.TrulyUnreferencedFiles, expectedTrulyUnreferenced) {
		t.Errorf("Expected TrulyUnreferencedFiles to be %v, but got %v", expectedTrulyUnreferenced, findResult.TrulyUnreferencedFiles)
	}

	// 检查可疑的未引用文件
	expectedSuspicious := []string{suspiciousPath}
	sort.Strings(findResult.SuspiciousFiles)
	sort.Strings(expectedSuspicious)
	if !reflect.DeepEqual(findResult.SuspiciousFiles, expectedSuspicious) {
		t.Errorf("Expected SuspiciousFiles to be %v, but got %v", expectedSuspicious, findResult.SuspiciousFiles)
	}

	// 检查统计数据
	if findResult.Stats.TotalFiles != 4 {
		t.Errorf("Expected TotalFiles to be 4, but got %d", findResult.Stats.TotalFiles)
	}
	if findResult.Stats.TrulyUnreferencedFiles != 1 {
		t.Errorf("Expected TrulyUnreferencedFiles count to be 1, but got %d", findResult.Stats.TrulyUnreferencedFiles)
	}
	if findResult.Stats.SuspiciousFiles != 1 {
		t.Errorf("Expected SuspiciousFiles count to be 1, but got %d", findResult.Stats.SuspiciousFiles)
	}

	// 检查摘要信息
	expectedSummary := "扫描文件 4 个，发现 1 个真正未引用文件和 1 个可疑文件。"
	if summary := result.Summary(); summary != expectedSummary {
		t.Errorf("Expected Summary() to be '%s', but got '%s'", expectedSummary, summary)
	}
}
