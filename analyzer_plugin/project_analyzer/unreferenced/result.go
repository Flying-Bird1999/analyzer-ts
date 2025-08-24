package unreferenced

import (
	"fmt"
	projectanalyzer "main/analyzer_plugin/project_analyzer"
	"strings"
)

// FindUnreferencedFilesResult 保存了“未引用文件”分析的完整结果。
// 它实现了 projectanalyzer.Result 接口。
type FindUnreferencedFilesResult struct {
	// Configuration 记录了本次分析所使用的配置参数。
	Configuration AnalysisConfiguration `json:"configuration"`
	// Stats 包含了本次分析的各项统计数据。
	Stats SummaryStats `json:"stats"`
	// EntrypointFiles 是在本次分析中被当作入口点的文件列表。
	EntrypointFiles []string `json:"entrypointFiles"`
	// SuspiciousFiles 是一些虽然未被直接引用，但根据其命名或位置，可能很重要的文件（例如配置文件），需要人工检查。
	SuspiciousFiles []string `json:"suspiciousFiles"`
	// TrulyUnreferencedFiles 是被认为是“真正”未被引用的文件列表，可以相对安全地删除。
	TrulyUnreferencedFiles []string `json:"trulyUnreferencedFiles"`
}

// 确保 Result 结构体实现了 projectanalyzer.Result 接口。
var _ projectanalyzer.Result = (*FindUnreferencedFilesResult)(nil)

// Name 返回该结果对应的分析器的名称。
func (r *FindUnreferencedFilesResult) Name() string {
	return "Find Unreferenced Files"
}

// Summary 返回对结果的简短、人类可读的摘要。
func (r *FindUnreferencedFilesResult) Summary() string {
	return fmt.Sprintf(
		"扫描文件 %d 个，发现 %d 个真正未引用文件和 %d 个可疑文件。",
		r.Stats.TotalFiles,
		r.Stats.TrulyUnreferencedFiles,
		r.Stats.SuspiciousFiles,
	)
}

// ToJSON 将结果的完整数据序列化为 JSON 格式。
func (r *FindUnreferencedFilesResult) ToJSON(indent bool) ([]byte, error) {
	return projectanalyzer.ToJSONBytes(r, indent)
}

// ToConsole 将结果格式化为适合在控制台（终端）中打印的字符串。
func (r *FindUnreferencedFilesResult) ToConsole() string {
	totalUnreferenced := len(r.TrulyUnreferencedFiles) + len(r.SuspiciousFiles)
	if totalUnreferenced == 0 {
		return fmt.Sprintf("✅ %s 没有发现任何未引用文件。", r.Summary())
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("⚠️ %s\n", r.Summary()))

	if len(r.TrulyUnreferencedFiles) > 0 {
		builder.WriteString("\n--- 🗑️ 真正未引用的文件 (可以安全删除) ---\n")
		for _, file := range r.TrulyUnreferencedFiles {
			builder.WriteString(fmt.Sprintf("  - %s\n", file))
		}
	}

	if len(r.SuspiciousFiles) > 0 {
		builder.WriteString("\n--- 🤔 可疑的未引用文件 (请人工检查) ---\n")
		for _, file := range r.SuspiciousFiles {
			builder.WriteString(fmt.Sprintf("  - %s\n", file))
		}
	}

	return builder.String()
}
