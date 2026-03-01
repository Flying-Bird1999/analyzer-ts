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

	// 受影响组件（包含所有需要关注/测试的组件）
	if len(r.ImpactedComponents) > 0 {
		sb.WriteString("⚠️  受影响组件（需要测试）:\n")
		names := sortedStringKeys(r.ImpactedComponents)
		for _, name := range names {
			impacts := r.ImpactedComponents[name]
			sb.WriteString(fmt.Sprintf("  • %s\n", name))
			for _, impact := range impacts {
				sb.WriteString(fmt.Sprintf("    - %s\n", impact.DisplayReason()))
			}
		}
		sb.WriteString("\n")
	}

	// 变更详情（可选信息）
	if len(r.ChangedComponents) > 0 {
		sb.WriteString("📦 变更组件详情:\n")
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

	// 变更函数详情（可选信息）
	if len(r.ChangedFunctions) > 0 {
		sb.WriteString("🔧 变更函数详情:\n")
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
	directCount := 0
	for _, impacts := range r.ImpactedComponents {
		for _, impact := range impacts {
			if impact.Relation == RelationDirect {
				directCount++
				break
			}
		}
	}
	return fmt.Sprintf(
		"分析完成: %d 个组件受影响（其中 %d 个直接变更）",
		len(r.ImpactedComponents),
		directCount,
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

// DisplayReason 返回人类可读的影响原因描述
func (c *ComponentImpact) DisplayReason() string {
	switch c.Relation {
	case RelationDirect:
		return fmt.Sprintf("直接变更 %s", c.ChangeSource)
	case RelationDepends:
		return fmt.Sprintf("依赖组件 %s", c.ChangeSource)
	case RelationImports:
		return fmt.Sprintf("引用函数 %s", c.ChangeSource)
	case RelationIndirect:
		if len(c.Path) > 0 {
			// 显示完整传播路径
			return fmt.Sprintf("间接受 %s 影响（路径: %s）", c.ChangeSource, formatPath(c.Path))
		}
		return fmt.Sprintf("间接受 %s 影响", c.ChangeSource)
	default:
		return fmt.Sprintf("受 %s 影响", c.ChangeSource)
	}
}

// formatPath 格式化路径为字符串
func formatPath(path []string) string {
	result := ""
	for i, p := range path {
		if i > 0 {
			result += " → "
		}
		result += p
	}
	return result
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
