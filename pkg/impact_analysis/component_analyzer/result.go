// Package component_analyzer 提供组件级影响分析功能。
// 这是组件库专用能力，基于 file_analyzer 的结果进行组件映射。
package component_analyzer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// =============================================================================
// Result 方法扩展
// =============================================================================

// String 返回结果的字符串表示
func (r *Result) String() string {
	return fmt.Sprintf("ComponentAnalysisResult: %d changed components, %d impacted components",
		len(r.Changes), len(r.Impact))
}

// ToJSON 将结果序列化为 JSON
func (r *Result) ToJSON(indent bool) ([]byte, error) {
	if indent {
		return json.MarshalIndent(r, "", "  ")
	}
	return json.Marshal(r)
}

// ToConsole 将结果格式化为控制台输出
func (r *Result) ToConsole() string {
	var buffer bytes.Buffer

	// 标题
	buffer.WriteString("=====================================\n")
	buffer.WriteString("组件级影响分析报告\n")
	buffer.WriteString("=====================================\n\n")

	// 元数据
	buffer.WriteString(fmt.Sprintf("分析时间: %s\n", r.Meta.AnalyzedAt.Format("2006-01-02 15:04:05")))
	buffer.WriteString(fmt.Sprintf("组件总数: %d\n", r.Meta.TotalComponentCount))
	buffer.WriteString(fmt.Sprintf("变更组件: %d\n", r.Meta.ChangedComponentCount))
	buffer.WriteString(fmt.Sprintf("影响组件: %d\n\n", r.Meta.ImpactComponentCount))

	// 直接变更的组件
	if len(r.Changes) > 0 {
		buffer.WriteString("=====================================\n")
		buffer.WriteString("直接变更的组件\n")
		buffer.WriteString("=====================================\n\n")
		for _, change := range r.Changes {
			buffer.WriteString(fmt.Sprintf("▶ %s (%s)\n",
				change.Name, change.Action))
			if len(change.ChangedFiles) > 0 {
				buffer.WriteString("  变更文件:\n")
				for _, file := range change.ChangedFiles {
					buffer.WriteString(fmt.Sprintf("    - %s\n", file))
				}
			}
		}
		buffer.WriteString("\n")
	}

	// 受影响的组件
	if len(r.Impact) > 0 {
		buffer.WriteString("=====================================\n")
		buffer.WriteString("受影响的组件\n")
		buffer.WriteString("=====================================\n\n")

		// 按影响层级排序
		sortedByLevel := r.sortByImpactLevel()
		for _, impact := range sortedByLevel {
			buffer.WriteString(fmt.Sprintf("▶ %s (层级: %d)\n",
				impact.Name, impact.ImpactLevel))
			if len(impact.ChangePaths) > 0 {
				buffer.WriteString("  影响路径:\n")
				for _, path := range impact.ChangePaths {
					buffer.WriteString(fmt.Sprintf("    %s\n", path))
				}
			}
			buffer.WriteString("\n")
		}
	}

	return buffer.String()
}

// sortByImpactLevel 按影响层级排序受影响组件
func (r *Result) sortByImpactLevel() []ComponentImpactInfo {
	// 创建副本以避免修改原始数据
	sorted := make([]ComponentImpactInfo, len(r.Impact))
	copy(sorted, r.Impact)

	sort.Slice(sorted, func(i, j int) bool {
		// 先按影响层级排序（层级越高越靠前）
		if sorted[i].ImpactLevel != sorted[j].ImpactLevel {
			return sorted[i].ImpactLevel > sorted[j].ImpactLevel
		}
		// 再按名称排序
		return sorted[i].Name < sorted[j].Name
	})

	return sorted
}

// GetImpactedComponentsByLevel 获取指定影响层级的组件
func (r *Result) GetImpactedComponentsByLevel(level int) []ComponentImpactInfo {
	result := make([]ComponentImpactInfo, 0)
	for _, impact := range r.Impact {
		if int(impact.ImpactLevel) == level {
			result = append(result, impact)
		}
	}
	return result
}

// GetImpactPathsForComponent 获取指定组件的影响路径
func (r *Result) GetImpactPathsForComponent(componentName string) []ComponentImpactPathInfo {
	result := make([]ComponentImpactPathInfo, 0)
	for _, path := range r.Paths {
		if path.TargetComponent == componentName {
			result = append(result, path)
		}
	}
	return result
}

// GetComponentChange 获取指定组件的变更信息
func (r *Result) GetComponentChange(componentName string) *ComponentChange {
	for _, change := range r.Changes {
		if change.Name == componentName {
			return &change
		}
	}
	return nil
}

// GetComponentImpact 获取指定组件的影响信息
func (r *Result) GetComponentImpact(componentName string) *ComponentImpactInfo {
	for _, impact := range r.Impact {
		if impact.Name == componentName {
			return &impact
		}
	}
	return nil
}

// GetHighestImpactLevel 获取最高的影响层级
func (r *Result) GetHighestImpactLevel() int {
	maxLevel := 0
	for _, impact := range r.Impact {
		if int(impact.ImpactLevel) > maxLevel {
			maxLevel = int(impact.ImpactLevel)
		}
	}
	return maxLevel
}

// GetCriticalImpacts 获取所有高风险影响（层级 >= 2）
func (r *Result) GetCriticalImpacts() []ComponentImpactInfo {
	result := make([]ComponentImpactInfo, 0)
	for _, impact := range r.Impact {
		if impact.ImpactLevel >= 2 {
			result = append(result, impact)
		}
	}
	return result
}

// GetDirectImpacts 获取所有直接影响（层级 = 1）
func (r *Result) GetDirectImpacts() []ComponentImpactInfo {
	result := make([]ComponentImpactInfo, 0)
	for _, impact := range r.Impact {
		if impact.ImpactLevel == 1 {
			result = append(result, impact)
		}
	}
	return result
}

// GetTransitiveImpacts 获取所有传递影响（层级 > 1）
func (r *Result) GetTransitiveImpacts() []ComponentImpactInfo {
	result := make([]ComponentImpactInfo, 0)
	for _, impact := range r.Impact {
		if impact.ImpactLevel > 1 {
			result = append(result, impact)
		}
	}
	return result
}

// CountByImpactLevel 统计各影响层级的组件数量
func (r *Result) CountByImpactLevel() map[int]int {
	counts := make(map[int]int)
	for _, impact := range r.Impact {
		level := int(impact.ImpactLevel)
		counts[level]++
	}
	return counts
}
