// package countas 包含了用于统计项目中 'as' 类型断言使用情况的核心业务逻辑。
package countas

import (
	projectanalyzer "main/analyzer_plugin/project_analyzer"
)

// Counter 是"统计as"分析器的实现。
type Counter struct{}

var _ projectanalyzer.Analyzer = (*Counter)(nil)

func (c *Counter) Name() string {
	return "count-as"
}

func (c *Counter) Configure(params map[string]string) error {
	return nil
}

func (c *Counter) Analyze(ctx *projectanalyzer.ProjectContext) (projectanalyzer.Result, error) {
	parseResult := ctx.ParsingResult

	totalAsCount := 0
	var fileCounts []FileCount

	for filePath, fileData := range parseResult.Js_Data {
		asCountInFile := len(fileData.ExtractedNodes.AsExpressions)

		if asCountInFile > 0 {
			fileCounts = append(fileCounts, FileCount{
				FilePath: filePath,
				AsCount:  asCountInFile,
				Details:  fileData.ExtractedNodes.AsExpressions,
			})
		}
		totalAsCount += asCountInFile
	}

	finalResult := &CountAsResult{
		TotalAsCount: totalAsCount,
		FileCounts:   fileCounts,
		FilesParsed:  len(parseResult.Js_Data),
	}

	return finalResult, nil
}
