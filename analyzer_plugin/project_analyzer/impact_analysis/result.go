package impact_analysis

import (
	"bytes"
	"fmt"
	"sort"

	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

// =============================================================================
// 数据结构定义
// =============================================================================

// ImpactMeta 影响分析元数据
type ImpactMeta struct {
	AnalyzedAt    string `json:"analyzedAt"`    // 分析时间
	ComponentCount int    `json:"componentCount"` // 组件总数
	ChangedFileCount int  `json:"changedFileCount"` // 变更文件数
	ChangeSource   string `json:"changeSource"`   // 变更来源
}

// ComponentChange 组件变更信息
type ComponentChange struct {
	Name         string `json:"name"`         // 组件名称
	Action       string `json:"action"`       // 变更类型: modified/added/deleted
	ChangedFiles []string `json:"changedFiles"` // 变更的文件列表
}

// ImpactComponent 受影响的组件
type ImpactComponent struct {
	Name         string   `json:"name"`         // 组件名称
	ImpactLevel int      `json:"impactLevel"`  // 影响层级（0=直接，1=间接，2=二级间接...）
	RiskLevel    string   `json:"riskLevel"`    // 风险等级: low/medium/high/critical
	ChangePaths []string `json:"changePaths"`   // 从变更组件到该组件的路径
}

// ChangePath 变更路径
type ChangePath struct {
	From string `json:"from"` // 变更组件
	To   string `json:"to"`   // 受影响组件
	Path []string `json:"path"` // 完整路径（如 [Button, Input, Select]）
}

// Recommendation 建议
type Recommendation struct {
	Type        string `json:"type"`        // 建议类型: test/review/refactor/document
	Priority    string `json:"priority"`    // 优先级: low/medium/high/critical
	Description string `json:"description"` // 建议描述
	Target      string `json:"target"`      // 目标组件（可选）
}

// ImpactAnalysisResult 影响分析结果
type ImpactAnalysisResult struct {
	Meta           ImpactMeta         `json:"meta"`           // 元数据
	Changes        []ComponentChange  `json:"changes"`        // 变更的组件列表
	Impact         []ImpactComponent `json:"impact"`         // 受影响的组件列表
	ChangePaths    []ChangePath       `json:"changePaths"`    // 完整变更路径
	Recommendations []Recommendation `json:"recommendations"` // 建议列表
}

// =============================================================================
// Result 接口实现
// =============================================================================

// Name 返回分析结果标识符
func (r *ImpactAnalysisResult) Name() string {
	return "impact-analysis"
}

// Summary 返回分析结果摘要
func (r *ImpactAnalysisResult) Summary() string {
	return fmt.Sprintf("影响分析完成，发现 %d 个组件变更，影响 %d 个组件。",
		len(r.Changes), len(r.Impact))
}

// ToJSON 将结果序列化为 JSON
func (r *ImpactAnalysisResult) ToJSON(indent bool) ([]byte, error) {
	return projectanalyzer.ToJSONBytes(r, indent)
}

// ToConsole 将结果格式化为控制台输出
func (r *ImpactAnalysisResult) ToConsole() string {
	var buffer bytes.Buffer

	// 标题
	buffer.WriteString("=====================================\n")
	buffer.WriteString("影响分析报告\n")
	buffer.WriteString("=====================================\n\n")

	// 元数据
	buffer.WriteString(fmt.Sprintf("分析时间: %s\n", r.Meta.AnalyzedAt))
	buffer.WriteString(fmt.Sprintf("组件总数: %d\n", r.Meta.ComponentCount))
	buffer.WriteString(fmt.Sprintf("变更文件: %d\n\n", r.Meta.ChangedFileCount))

	// 变更的组件
	buffer.WriteString("=====================================\n")
	buffer.WriteString("变更的组件\n")
	buffer.WriteString("=====================================\n\n")
	for _, change := range r.Changes {
		buffer.WriteString(fmt.Sprintf("▶ %s (%s)\n", change.Name, change.Action))
		for _, file := range change.ChangedFiles {
			buffer.WriteString(fmt.Sprintf("  - %s\n", file))
		}
		buffer.WriteString("\n")
	}

	// 受影响的组件
	buffer.WriteString("=====================================\n")
	buffer.WriteString("受影响的组件\n")
	buffer.WriteString("=====================================\n\n")

	// 按风险等级排序
	sortedByRisk := r.sortByRisk()
	for _, item := range sortedByRisk {
		comp := item.comp
		buffer.WriteString(fmt.Sprintf("▶ %s (风险: %s, 层级: %d)\n",
			comp.Name, comp.RiskLevel, comp.ImpactLevel))
		if len(comp.ChangePaths) > 0 {
			buffer.WriteString("  变更路径:\n")
			for _, path := range comp.ChangePaths {
				buffer.WriteString(fmt.Sprintf("    %s\n", path))
			}
		}
		buffer.WriteString("\n")
	}

	// 建议
	if len(r.Recommendations) > 0 {
		buffer.WriteString("=====================================\n")
		buffer.WriteString("建议\n")
		buffer.WriteString("=====================================\n\n")
		for i, rec := range r.Recommendations {
			buffer.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, rec.Priority, rec.Type))
			buffer.WriteString(fmt.Sprintf("   %s\n", rec.Description))
			if rec.Target != "" {
				buffer.WriteString(fmt.Sprintf("   目标: %s\n", rec.Target))
			}
			buffer.WriteString("\n")
		}
	}

	return buffer.String()
}

// sortByRisk 按风险等级排序受影响组件
func (r *ImpactAnalysisResult) sortByRisk() []struct {
	comp ImpactComponent
	risk int
} {
	riskOrder := map[string]int{"critical": 4, "high": 3, "medium": 2, "low": 1}

	items := make([]struct {
		comp ImpactComponent
		risk int
	}, len(r.Impact))

	for i, comp := range r.Impact {
		items[i] = struct {
			comp ImpactComponent
			risk int
		}{
			comp: comp,
			risk: riskOrder[comp.RiskLevel],
		}
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].risk != items[j].risk {
			return items[i].risk > items[j].risk
		}
		if items[i].comp.ImpactLevel != items[j].comp.ImpactLevel {
			return items[i].comp.ImpactLevel > items[j].comp.ImpactLevel
		}
		return items[i].comp.Name < items[j].comp.Name
	})

	return items
}

