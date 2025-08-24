package callgraph

import (
	"fmt"
	projectanalyzer "main/analyzer_plugin/project_analyzer"
)

// FindCallersResult 保存了“查找调用方”分析的完整结果。
type FindCallersResult struct {
	OverallSummary OverallSummary     `json:"overallSummary"`
	PerFileResults []SingleFileResult `json:"perFileResults"`
}

var _ projectanalyzer.Result = (*FindCallersResult)(nil)

func (r *FindCallersResult) Name() string {
	return "Find Callers"
}

func (r *FindCallersResult) Summary() string {
	return fmt.Sprintf(
		"分析 %d 个目标文件，共发现 %d 个上游调用文件。",
		len(r.OverallSummary.TargetFiles),
		r.OverallSummary.TotalAffectedFiles,
	)
}

func (r *FindCallersResult) ToJSON(indent bool) ([]byte, error) {
	return projectanalyzer.ToJSONBytes(r, indent)
}

func (r *FindCallersResult) ToConsole() string {
	// ... (省略) ...
	return ""
}

func formatTree(node *CallerNode, prefix string, isLast bool) string {
	// ... (省略) ...
	return ""
}
