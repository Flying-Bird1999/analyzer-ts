package countas

import (
	"fmt"
	"main/analyzer/parser"
	projectanalyzer "main/analyzer_plugin/project_analyzer"
	"strings"
)

// CountAsResult 是 'as' 类型断言分析的最终结果的顶层结构体。
// 它实现了 projectanalyzer.Result 接口。
type CountAsResult struct {
	FilesParsed  int            `json:"filesParsed"`  // 成功解析的 JS/TS 文件数量
	TotalAsCount int            `json:"totalAsCount"` // 项目中 'as' 类型断言的总数
	FileCounts   []FileCount    `json:"fileCounts"`   // 每个文件的 'as' 类型断言统计列表
}

// FileCount 存储单个文件中 'as' 类型断言的使用情况统计。
type FileCount struct {
	FilePath string              `json:"filePath"` // 文件的绝对路径
	AsCount  int                 `json:"asCount"`  // 该文件中的 'as' 类型断言总数
	Details  []parser.AsExpression `json:"details"`  // 该文件中所有 'as' 类型断言的详细信息列表
}

// 确保 Result 结构体实现了 projectanalyzer.Result 接口。
var _ projectanalyzer.Result = (*CountAsResult)(nil)

// Name 返回该结果对应的分析器的名称。
func (r *CountAsResult) Name() string {
	return "Count As Usage"
}

// Summary 返回对结果的简短、人类可读的摘要。
func (r *CountAsResult) Summary() string {
	return fmt.Sprintf(
		"扫描文件 %d 个，共发现 %d 处 'as' 类型断言使用。",
		r.FilesParsed,
		r.TotalAsCount,
	)
}

// ToJSON 将结果的完整数据序列化为 JSON 格式。
func (r *CountAsResult) ToJSON(indent bool) ([]byte, error) {
	return projectanalyzer.ToJSONBytes(r, indent)
}

// ToConsole 将结果格式化为适合在控制台（终端）中打印的字符串。
func (r *CountAsResult) ToConsole() string {
	if r.TotalAsCount == 0 {
		return fmt.Sprintf("✅ %s 太棒了，项目中没有发现 'as' 类型断言！", r.Summary())
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("⚠️ %s\n", r.Summary()))
	builder.WriteString("--------------------------------------------------\n")
	for _, fc := range r.FileCounts {
		builder.WriteString(fmt.Sprintf("  - %s (%d 处):\n", fc.FilePath, fc.AsCount))
		for _, detail := range fc.Details {
			builder.WriteString(fmt.Sprintf("    - Line %d: %s\n", detail.SourceLocation.Start.Line, detail.Raw))
		}
	}
	builder.WriteString("--------------------------------------------------\n")

	return builder.String()
}