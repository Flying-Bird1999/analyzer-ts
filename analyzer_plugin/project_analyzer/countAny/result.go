package countany

import (
	"fmt"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

// CountAnyResult 是 'any' 类型分析的最终结果的顶层结构体。
// 它实现了 projectanalyzer.Result 接口。
type CountAnyResult struct {
	FilesParsed   int         `json:"filesParsed"`   // 成功解析的 JS/TS 文件数量
	TotalAnyCount int         `json:"totalAnyCount"` // 项目中 'any' 类型的总数
	FileCounts    []FileCount `json:"fileCounts"`    // 每个文件的 'any' 类型统计列表
}

// FileCount 存储单个文件中 'any' 类型的使用情况统计。
type FileCount struct {
	FilePath string           `json:"filePath"` // 文件的绝对路径
	AnyCount int              `json:"anyCount"` // 该文件中的 'any' 类型总数
	Details  []parser.AnyInfo `json:"details"`  // 该文件中所有 'any' 类型的详细信息列表
}

// 确保 Result 结构体实现了 projectanalyzer.Result 接口。
var _ projectanalyzer.Result = (*CountAnyResult)(nil)

// Name 返回该结果对应的分析器的名称。
func (r *CountAnyResult) Name() string {
	return "Count Any Usage"
}

// Summary 返回对结果的简短、人类可读的摘要。
func (r *CountAnyResult) Summary() string {
	return fmt.Sprintf(
		"扫描文件 %d 个，共发现 %d 处 'any' 类型使用。",
		r.FilesParsed,
		r.TotalAnyCount,
	)
}

// ToJSON 将结果的完整数据序列化为 JSON 格式。
func (r *CountAnyResult) ToJSON(indent bool) ([]byte, error) {
	return projectanalyzer.ToJSONBytes(r, indent)
}

// ToConsole 将结果格式化为适合在控制台（终端）中打印的字符串。
func (r *CountAnyResult) ToConsole() string {
	if r.TotalAnyCount == 0 {
		return fmt.Sprintf("✅ %s 太棒了，项目中没有发现 'any' 类型！", r.Summary())
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("⚠️ %s\n", r.Summary()))
	builder.WriteString("--------------------------------------------------\n")
	for _, fc := range r.FileCounts {
		builder.WriteString(fmt.Sprintf("  - %s (%d 处):\n", fc.FilePath, fc.AnyCount))
		for _, detail := range fc.Details {
			builder.WriteString(fmt.Sprintf("    - Line %d: %s\n", detail.SourceLocation.Start.Line, detail.Raw))
		}
	}
	builder.WriteString("--------------------------------------------------\n")

	return builder.String()
}
