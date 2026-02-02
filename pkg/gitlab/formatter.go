// Package gitlab provides GitLab integration capabilities for analyzer-ts.
package gitlab

import (
	"fmt"
	"strings"
)

// =============================================================================
// Formatter - JSON 转 Markdown
// =============================================================================

// Formatter 格式化器
type Formatter struct {
	style CommentStyle
}

// CommentStyle 评论风格
type CommentStyle int

const (
	CommentStyleCompact CommentStyle = iota
	CommentStyleDetailed
)

// NewFormatter 创建格式化器
func NewFormatter(style CommentStyle) *Formatter {
	return &Formatter{
		style: style,
	}
}

// =============================================================================
// 格式化方法
// =============================================================================

// FormatImpactResult 格式化影响分析结果为 Markdown
func (f *Formatter) FormatImpactResult(result *ImpactAnalysisResult) (string, error) {
	var builder strings.Builder

	// 标题
	builder.WriteString("## 🔍 代码影响分析报告\n\n")

	// 概要
	builder.WriteString(f.formatSummary(result))

	// 变更组件
	if len(result.Changes) > 0 {
		builder.WriteString("\n### 🎯 变更组件\n\n")
		for _, change := range result.Changes {
			builder.WriteString(f.formatComponentChange(change))
		}
	}

	// 影响范围
	if len(result.Impact) > 0 {
		builder.WriteString("\n### 📈 影响范围\n\n")
		for _, impact := range result.Impact {
			builder.WriteString(f.formatImpactComponent(impact))
		}
	}

	// 建议
	if len(result.Recommendations) > 0 {
		builder.WriteString("\n### 💡 建议\n\n")
		for _, rec := range result.Recommendations {
			builder.WriteString(f.formatRecommendation(rec))
		}
	}

	// 页脚
	builder.WriteString("\n---\n\n")
	builder.WriteString("*由 analyzer-ts 自动生成\n")

	return builder.String(), nil
}

// formatSummary 格式化概要信息
func (f *Formatter) formatSummary(result *ImpactAnalysisResult) string {
	var builder strings.Builder

	builder.WriteString("### 📊 概要\n\n")

	// 统计风险等级
	riskCount := make(map[string]int)
	for _, impact := range result.Impact {
		riskCount[impact.RiskLevel]++
	}

	builder.WriteString("| 指标 | 数值 |\n")
	builder.WriteString("|------|------|\n")
	builder.WriteString(fmt.Sprintf("| 变更组件 | %d |\n", len(result.Changes)))
	builder.WriteString(fmt.Sprintf("| 受影响组件 | %d |\n", len(result.Impact)-len(result.Changes))) // 排除变更组件本身
	builder.WriteString(fmt.Sprintf("| 高风险 | %d |\n", riskCount["high"]))
	builder.WriteString(fmt.Sprintf("| 中风险 | %d |\n", riskCount["medium"]))
	builder.WriteString(fmt.Sprintf("| 低风险 | %d |\n", riskCount["low"]))

	return builder.String()
}

// formatComponentChange 格式化组件变更
func (f *Formatter) formatComponentChange(change ComponentChange) string {
	var builder strings.Builder

	actionIcon := map[string]string{
		"modified": "📝",
		"added":    "✨",
		"deleted":  "❌",
	}[change.Action]

	builder.WriteString(fmt.Sprintf("#### %s %s\n\n", actionIcon, change.Name))

	for _, file := range change.ChangedFiles {
		builder.WriteString(fmt.Sprintf("- `%s`\n", file))
	}
	builder.WriteString("\n")

	return builder.String()
}

// formatImpactComponent 格式化受影响组件
func (f *Formatter) formatImpactComponent(impact ImpactComponent) string {
	var builder strings.Builder

	// 风险图标
	riskIcon := map[string]string{
		"critical": "🔴",
		"high":     "🟠",
		"medium":   "🟡",
		"low":      "🟢",
	}[impact.RiskLevel]

	builder.WriteString(fmt.Sprintf("#### %s %s (风险: %s, 层级: %d)\n\n",
		riskIcon, impact.Name, impact.RiskLevel, impact.ImpactLevel))

	if len(impact.ChangePaths) > 0 {
		builder.WriteString("变更路径:\n")
		for _, path := range impact.ChangePaths {
			builder.WriteString(fmt.Sprintf("- %s\n", path))
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

// formatRecommendation 格式化建议
func (f *Formatter) formatRecommendation(rec Recommendation) string {
	priorityIcon := map[string]string{
		"critical": "🔴",
		"high":     "🟠",
		"medium":   "🟡",
		"low":      "🟢",
	}[rec.Priority]

	typeIcon := map[string]string{
		"review":   "👁",
		"test":     "🧪",
		"document": "📄",
		"refactor": "♻️",
	}[rec.Type]

	return fmt.Sprintf("- [%s%s] **%s %s**: %s\n",
		priorityIcon, typeIcon, rec.Priority, rec.Type, rec.Description)
}

// FormatSummary 简化的摘要格式（用于紧凑模式）
func (f *Formatter) FormatSummary(result *ImpactAnalysisResult) string {
	var builder strings.Builder

	builder.WriteString("## 代码影响分析\n\n")
	builder.WriteString(fmt.Sprintf("- **变更组件**: %d\n", len(result.Changes)))
	builder.WriteString(fmt.Sprintf("- **受影响组件**: %d\n", len(result.Impact)-len(result.Changes)))

	// 统计风险
	criticalCount := 0
	highCount := 0
	for _, impact := range result.Impact {
		if impact.RiskLevel == "critical" {
			criticalCount++
		} else if impact.RiskLevel == "high" {
			highCount++
		}
	}

	builder.WriteString(fmt.Sprintf("- **🔴 严重风险**: %d\n", criticalCount))
	builder.WriteString(fmt.Sprintf("- **🟠 高风险**: %d\n", highCount))

	return builder.String()
}

// FormatRiskTable 格式化风险表格
func (f *Formatter) FormatRiskTable(result *ImpactAnalysisResult) string {
	var builder strings.Builder

	builder.WriteString("### 风险详情\n\n")
	builder.WriteString("| 组件 | 风险等级 | 层级 |\n")
	builder.WriteString("|------|----------|------|\n")

	for _, impact := range result.Impact {
		// 跳过变更组件本身（level 0）
		if impact.ImpactLevel == 0 {
			continue
		}

		riskIcon := map[string]string{
			"critical": "🔴",
			"high":     "🟠",
			"medium":   "🟡",
			"low":      "🟢",
		}[impact.RiskLevel]

		builder.WriteString(fmt.Sprintf("| %s | %s %s | %d |\n",
			impact.Name, riskIcon, impact.RiskLevel, impact.ImpactLevel))
	}

	return builder.String()
}
