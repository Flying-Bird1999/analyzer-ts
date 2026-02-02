// Package file_analyzer 提供文件级影响分析功能。
// 这是通用的能力，适用于所有前端项目，不依赖 component-manifest.json。
package file_analyzer

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
	return fmt.Sprintf("FileAnalysisResult: %d changed files, %d impacted files",
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
	buffer.WriteString("文件级影响分析报告\n")
	buffer.WriteString("=====================================\n\n")

	// 元数据
	buffer.WriteString(fmt.Sprintf("分析时间: %s\n", r.Meta.AnalyzedAt.Format("2006-01-02 15:04:05")))
	buffer.WriteString(fmt.Sprintf("文件总数: %d\n", r.Meta.TotalFileCount))
	buffer.WriteString(fmt.Sprintf("变更文件: %d\n", r.Meta.ChangedFileCount))
	buffer.WriteString(fmt.Sprintf("影响文件: %d\n\n", r.Meta.ImpactFileCount))

	// 直接变更的文件
	if len(r.Changes) > 0 {
		buffer.WriteString("=====================================\n")
		buffer.WriteString("直接变更的文件\n")
		buffer.WriteString("=====================================\n\n")
		for _, change := range r.Changes {
			buffer.WriteString(fmt.Sprintf("▶ %s (%s, 符号数: %d)\n",
				change.Path, change.ChangeType, change.SymbolCount))
		}
		buffer.WriteString("\n")
	}

	// 受影响的文件
	if len(r.Impact) > 0 {
		buffer.WriteString("=====================================\n")
		buffer.WriteString("受影响的文件\n")
		buffer.WriteString("=====================================\n\n")

		// 按影响层级排序
		sortedByLevel := r.sortByImpactLevel()
		for _, impact := range sortedByLevel {
			buffer.WriteString(fmt.Sprintf("▶ %s (层级: %d, 类型: %s, 符号数: %d)\n",
				impact.Path, impact.ImpactLevel, impact.ImpactType, impact.SymbolCount))
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

// sortByImpactLevel 按影响层级排序受影响文件
func (r *Result) sortByImpactLevel() []FileImpactInfo {
	// 创建副本以避免修改原始数据
	sorted := make([]FileImpactInfo, len(r.Impact))
	copy(sorted, r.Impact)

	sort.Slice(sorted, func(i, j int) bool {
		// 先按影响层级排序（层级越高越靠前）
		if sorted[i].ImpactLevel != sorted[j].ImpactLevel {
			return sorted[i].ImpactLevel > sorted[j].ImpactLevel
		}
		// 再按路径排序
		return sorted[i].Path < sorted[j].Path
	})

	return sorted
}

// GetImpactedFilesByLevel 获取指定影响层级的文件
func (r *Result) GetImpactedFilesByLevel(level int) []FileImpactInfo {
	result := make([]FileImpactInfo, 0)
	for _, impact := range r.Impact {
		if int(impact.ImpactLevel) == level {
			result = append(result, impact)
		}
	}
	return result
}

// GetImpactedFilesByType 获取指定影响类型的文件
func (r *Result) GetImpactedFilesByType(impactType string) []FileImpactInfo {
	result := make([]FileImpactInfo, 0)
	for _, impact := range r.Impact {
		if string(impact.ImpactType) == impactType {
			result = append(result, impact)
		}
	}
	return result
}
