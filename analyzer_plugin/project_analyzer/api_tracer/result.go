package api_tracer

import (
	"fmt"
	"strings"

	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

// ApiTracerResult 存储了 `api-tracer` 分析器的完整结果。
type ApiTracerResult struct {
	// Findings 是一个列表，包含了所有找到的API调用点。
	Findings []ApiCallSite `json:"findings"`
}

// 确保 ApiTracerResult 实现了 projectanalyzer.Result 接口。
var _ projectanalyzer.Result = (*ApiTracerResult)(nil)

// Name 返回分析结果的名称。
func (r *ApiTracerResult) Name() string {
	return "api-tracer-result"
}

// Summary 返回对分析结果的简短总结。
func (r *ApiTracerResult) Summary() string {
	return fmt.Sprintf("找到了 %d 个API调用点。", len(r.Findings))
}

// ToJSON 将结果序列化为JSON格式的字节流。
func (r *ApiTracerResult) ToJSON(indent bool) ([]byte, error) {
	return projectanalyzer.ToJSONBytes(r, indent)
}

// ToConsole 将结果格式化为适合在控制台输出的字符串。
func (r *ApiTracerResult) ToConsole() string {
	if len(r.Findings) == 0 {
		return "没有找到匹配的API调用点。"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("找到了 %d 个API调用点:\n", len(r.Findings)))
	for _, finding := range r.Findings {
		sb.WriteString(fmt.Sprintf("  - API: %s\n", finding.ApiPath))
		sb.WriteString(fmt.Sprintf("    文件: %s\n", finding.FilePath))
		sb.WriteString(fmt.Sprintf("    代码: %s\n", finding.Raw))
	}
	return sb.String()
}

// AnalyzerName 返回对应的分析器名称
func (r *ApiTracerResult) AnalyzerName() string {
	return "api-tracer"
}
