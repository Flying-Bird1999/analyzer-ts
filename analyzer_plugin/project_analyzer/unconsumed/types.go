// Package unconsumed 包含了查找项目中已导出但未被消费的变量所需的所有类型定义。
//
// 这个包定义了 unconsumed 分析器的完整类型系统，包括：
// - Finding: 单个未使用导出项的详细信息
// - SummaryStats: 分析过程的统计数据
// - Result: 完整的分析结果结构
//
// 所有类型都支持 JSON 序列化，便于数据导出和集成到其他系统。
package unconsumed

import (
	"fmt"
	"strings"

	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

// Finding 代表一个具体的、已导出但未被消费的实体（变量、函数、类型等）。
// 为了保持命名一致性，我们将原来的 Export 结构体更名为 Finding。
type Finding struct {
	// FilePath 是这个导出项所在的文件路径。
	FilePath string `json:"filePath"`
	// ExportName 是导出项的名称（标识符）。"default" 代表默认导出。
	ExportName string `json:"exportName"`
	// Line 是导出语句所在的行号，便于快速定位。
	Line int `json:"line"`
	// Kind 描述了导出项的类型（如 var, const, function, interface, type, enum）。
	Kind string `json:"kind"`
}

// SummaryStats 提供了关于本次分析的统计数据。
type SummaryStats struct {
	// TotalFilesScanned 是本次分析扫描的总文件数。
	TotalFilesScanned int `json:"totalFilesScanned"`
	// TotalExportsFound 是在项目中找到的总导出项数量。
	TotalExportsFound int `json:"totalExportsFound"`
	// UnconsumedExportsFound 是未被消费的导出项的数量。
	UnconsumedExportsFound int `json:"unconsumedExportsFound"`
}

// Result 保存了“未消费导出”分析的完整结果。
// 它实现了 projectanalyzer.Result 接口，因此可以被上层统一处理。
type Result struct {
	// Findings 是找到的所有未被消费的导出项的列表。
	Findings []Finding `json:"findings"`
	// Stats 包含了本次分析的统计数据。
	Stats SummaryStats `json:"stats"`
}

// 确保 Result 结构体实现了 projectanalyzer.Result 接口。
// 这是一个编译时检查，如果接口实现不完整，编译会失败。
var _ projectanalyzer.Result = (*Result)(nil)

// Name 返回该结果对应的分析器的名称。
func (r *Result) Name() string {
	return "Unconsumed Exports Finder"
}

// Summary 返回对结果的简短、人类可读的摘要。
func (r *Result) Summary() string {
	return fmt.Sprintf(`扫描文件 %d 个，发现导出 %d 个，其中未使用导出 %d 个。`,
		r.Stats.TotalFilesScanned,
		r.Stats.TotalExportsFound,
		r.Stats.UnconsumedExportsFound,
	)
}

// ToJSON 将结果的完整数据序列化为 JSON 格式。
func (r *Result) ToJSON(indent bool) ([]byte, error) {
	return projectanalyzer.ToJSONBytes(r, indent)
}

// ToConsole 将结果格式化为适合在控制台（终端）中打印的字符串。
func (r *Result) ToConsole() string {
	if len(r.Findings) == 0 {
		return "✅ " + r.Summary() + " 没有发现未使用的导出。"
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("⚠️ %s\n", r.Summary()))
	builder.WriteString("--------------------------------------------------\n")
	for _, f := range r.Findings {
		builder.WriteString(
			fmt.Sprintf(`  - [%s] %s:%d 	 (%s)`+"\n", f.Kind, f.FilePath, f.Line, f.ExportName),
		)
	}
	builder.WriteString("--------------------------------------------------\n")

	return builder.String()
}

// AnalyzerName 返回对应的分析器名称
func (r *Result) AnalyzerName() string {
	return "unconsumed"
}
