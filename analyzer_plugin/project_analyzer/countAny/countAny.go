// package countany 包含了用于统计项目中 'any' 类型使用情况的核心业务逻辑。
package countany

import (
	"fmt"
	projectanalyzer "main/analyzer_plugin/project_analyzer"
	"main/analyzer_plugin/project_analyzer/internal/parser"
)

// Counter 是“统计any”分析器的实现。
type Counter struct{}

var _ projectanalyzer.Analyzer = (*Counter)(nil)

func (c *Counter) Name() string {
	return "count-any"
}

func (c *Counter) Configure(params map[string]string) error {
	return nil
}

func (c *Counter) Analyze(ctx *projectanalyzer.ProjectContext) (projectanalyzer.Result, error) {
	parseResult, err := parser.ParseProject(ctx.ProjectRoot, ctx.Exclude, ctx.IsMonorepo)
	if err != nil {
		return nil, fmt.Errorf("解析项目失败: %w", err)
	}

	totalAnyCount := 0
	var fileCounts []FileCount

	for filePath, fileData := range parseResult.Js_Data {
		anyCountInFile := len(fileData.AnyDeclarations)

		if anyCountInFile > 0 {
			fileCounts = append(fileCounts, FileCount{
				FilePath: filePath,
				AnyCount: anyCountInFile,
				Details:  fileData.AnyDeclarations,
			})
		}
		totalAnyCount += anyCountInFile
	}

	finalResult := &CountAnyResult{
		TotalAnyCount: totalAnyCount,
		FileCounts:    fileCounts,
		FilesParsed:   len(parseResult.Js_Data),
	}

	return finalResult, nil
}
