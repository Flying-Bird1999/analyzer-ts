// Package mr_component_impact 提供 MR 组件影响分析功能
package mr_component_impact

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// =============================================================================
// 结果输出方法
// =============================================================================

// ToJSON 将结果序列化为 JSON 格式
func (r *AnalysisResult) ToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToConsole 将结果格式化为控制台输出
func (r *AnalysisResult) ToConsole() string {
	var sb strings.Builder

	// 标题
	sb.WriteString("========================================\n")
	sb.WriteString("MR 组件影响分析报告\n")
	sb.WriteString("========================================\n\n")

	// 变更组件
	if len(r.ChangedComponents) > 0 {
		sb.WriteString("📦 变更组件:\n")
		names := sortedStringKeys(r.ChangedComponents)
		for _, name := range names {
			info := r.ChangedComponents[name]
			sb.WriteString(fmt.Sprintf("  • %s\n", name))
			for _, file := range info.Files {
				sb.WriteString(fmt.Sprintf("    - %s\n", file))
			}
		}
		sb.WriteString("\n")
	}

	// 变更函数
	if len(r.ChangedFunctions) > 0 {
		sb.WriteString("🔧 变更函数:\n")
		names := sortedStringKeys(r.ChangedFunctions)
		for _, name := range names {
			info := r.ChangedFunctions[name]
			sb.WriteString(fmt.Sprintf("  • %s\n", name))
			for _, file := range info.Files {
				sb.WriteString(fmt.Sprintf("    - %s\n", file))
			}
		}
		sb.WriteString("\n")
	}

	// 受影响组件
	if len(r.ImpactedComponents) > 0 {
		sb.WriteString("⚠️  受影响组件:\n")
		names := sortedStringKeys(r.ImpactedComponents)
		for _, name := range names {
			impacts := r.ImpactedComponents[name]
			sb.WriteString(fmt.Sprintf("  • %s\n", name))
			for _, impact := range impacts {
				sb.WriteString(fmt.Sprintf("    - %s\n", impact.ImpactReason))
			}
		}
		sb.WriteString("\n")
	}

	// 其他文件
	if len(r.OtherFiles) > 0 {
		sb.WriteString("📄 其他文件:\n")
		for _, file := range r.OtherFiles {
			sb.WriteString(fmt.Sprintf("  - %s\n", file))
		}
		sb.WriteString("\n")
	}

	// 摘要
	sb.WriteString("========================================\n")
	sb.WriteString(r.GetSummary())
	sb.WriteString("\n========================================\n")

	return sb.String()
}

// GetSummary 获取分析结果摘要
func (r *AnalysisResult) GetSummary() string {
	return fmt.Sprintf(
		"分析完成: %d 个组件变更, %d 个函数变更, %d 个组件受影响, %d 个其他文件",
		len(r.ChangedComponents),
		len(r.ChangedFunctions),
		len(r.ImpactedComponents),
		len(r.OtherFiles),
	)
}

// GetImpactedComponentNames 获取所有受影响组件的名称列表
func (r *AnalysisResult) GetImpactedComponentNames() []string {
	names := make([]string, 0, len(r.ImpactedComponents))
	for name := range r.ImpactedComponents {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetChangedComponentNames 获取所有变更组件的名称列表
func (r *AnalysisResult) GetChangedComponentNames() []string {
	names := make([]string, 0, len(r.ChangedComponents))
	for name := range r.ChangedComponents {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// =============================================================================
// 辅助函数
// =============================================================================

// sortedStringKeys 对 map 的键进行排序
func sortedStringKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
