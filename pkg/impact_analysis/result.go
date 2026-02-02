// Package impact_analysis 提供符号级影响分析功能。
package impact_analysis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// =============================================================================
// 分析结果类型定义
// =============================================================================

// AnalysisResult 影响分析结果
type AnalysisResult struct {
	Meta              ImpactMeta           `json:"meta"`
	Changes           []ComponentChange    `json:"changes"`           // 变更的组件列表
	Impact            []ImpactComponent    `json:"impact"`            // 受影响的组件列表
	SymbolChanges     []SymbolImpactChange `json:"symbolChanges"`     // 符号级变更详情
	SymbolImpactPaths []SymbolImpactPath   `json:"symbolImpactPaths"` // 符号级影响路径
	RiskAssessment    RiskAssessment       `json:"riskAssessment"`    // 风险评估
	Recommendations   []Recommendation     `json:"recommendations"`   // 建议列表
}

// =============================================================================
// 元数据
// =============================================================================

// ImpactMeta 影响分析元数据
type ImpactMeta struct {
	AnalyzedAt       string `json:"analyzedAt"`       // 分析时间
	ComponentCount   int    `json:"componentCount"`   // 组件总数
	ChangedFileCount int    `json:"changedFileCount"` // 变更文件数
	ChangeSource     string `json:"changeSource"`     // 变更来源
	SymbolCount      int    `json:"symbolCount"`      // 变更符号数量（新增）
}

// =============================================================================
// 组件变更
// =============================================================================

// ComponentChange 组件变更信息
type ComponentChange struct {
	Name         string   `json:"name"`         // 组件名称
	Action       string   `json:"action"`       // 变更类型: modified/added/deleted
	ChangedFiles []string `json:"changedFiles"` // 变更的文件列表
	SymbolCount  int      `json:"symbolCount"`  // 变更的符号数量（新增）
}

// =============================================================================
// 组件影响
// =============================================================================

// ImpactComponent 受影响的组件
type ImpactComponent struct {
	Name        string   `json:"name"`        // 组件名称
	ImpactLevel int      `json:"impactLevel"` // 影响层级（0=直接，1=间接，2=二级间接...）
	RiskLevel   string   `json:"riskLevel"`   // 风险等级: low/medium/high/critical
	ChangePaths []string `json:"changePaths"` // 从变更组件到该组件的路径
	SymbolCount int      `json:"symbolCount"` // 影响的符号数量（新增）
}

// =============================================================================
// 符号级影响变更
// =============================================================================

// SymbolImpactChange 符号级影响变更
type SymbolImpactChange struct {
	// 符号信息
	Symbol SymbolChange `json:"symbol"`

	// 所属组件
	ComponentName string `json:"componentName"`

	// 影响类型
	ImpactType ImpactType `json:"impactType"`
	// "breaking" - 破坏性变更（导出 API 修改/删除）
	// "internal" - 内部变更（仅内部实现）
	// "additive" - 增强性变更（新增导出）

	// 影响的下游组件
	AffectedComponents []string `json:"affectedComponents"`
}

// =============================================================================
// 符号级影响路径
// =============================================================================

// SymbolImpactPath 符号级影响路径
type SymbolImpactPath struct {
	// 源头符号
	SourceSymbol SymbolRef `json:"sourceSymbol"`

	// 传播路径
	Path []ImpactStep `json:"path"`

	// 最终影响的组件
	TargetComponent string `json:"targetComponent"`
}

// ImpactStep 影响传播步骤
type ImpactStep struct {
	Component string    `json:"component"` // 当前组件
	Symbol    SymbolRef `json:"symbol"`    // 涉及的符号
	Relation  string    `json:"relation"`  // "exports"/"imports"/"re-exports"
}

// =============================================================================
// 影响类型（定义在 types.go 中）
// =============================================================================

// =============================================================================
// 风险评估
// =============================================================================

// RiskAssessment 风险评估
type RiskAssessment struct {
	OverallRisk    string `json:"overallRisk"`    // low/medium/high/critical
	BreakingChange int    `json:"breakingChange"` // 破坏性变更数量
	InternalChange int    `json:"internalChange"` // 内部变更数量
	AdditiveChange int    `json:"additiveChange"` // 增强性变更数量
}

// =============================================================================
// 建议
// =============================================================================

// Recommendation 建议
type Recommendation struct {
	Type        string `json:"type"`        // 建议类型: test/review/refactor/document
	Priority    string `json:"priority"`    // 优先级: low/medium/high/critical
	Description string `json:"description"` // 建议描述
	Target      string `json:"target"`      // 目标组件（可选）
}

// =============================================================================
// 符号引用
// =============================================================================

// SymbolRef 符号引用
type SymbolRef struct {
	Name       string     `json:"name"`       // 符号名称
	Kind       SymbolKind `json:"kind"`       // 符号类型
	FilePath   string     `json:"filePath"`   // 文件路径
	ExportType ExportType `json:"exportType"` // 导出类型
}

// =============================================================================
// 结果方法
// =============================================================================

// Summary 返回分析结果摘要
func (r *AnalysisResult) Summary() string {
	return fmt.Sprintf("影响分析完成，发现 %d 个组件变更，影响 %d 个组件。",
		len(r.Changes), len(r.Impact))
}

// ToJSON 将结果序列化为 JSON
func (r *AnalysisResult) ToJSON(indent bool) ([]byte, error) {
	if indent {
		return json.MarshalIndent(r, "", "  ")
	}
	return json.Marshal(r)
}

// ToConsole 将结果格式化为控制台输出
func (r *AnalysisResult) ToConsole() string {
	var buffer bytes.Buffer

	// 标题
	buffer.WriteString("=====================================\n")
	buffer.WriteString("符号级影响分析报告\n")
	buffer.WriteString("=====================================\n\n")

	// 元数据
	buffer.WriteString(fmt.Sprintf("分析时间: %s\n", r.Meta.AnalyzedAt))
	buffer.WriteString(fmt.Sprintf("组件总数: %d\n", r.Meta.ComponentCount))
	buffer.WriteString(fmt.Sprintf("变更文件: %d\n", r.Meta.ChangedFileCount))
	buffer.WriteString(fmt.Sprintf("变更符号: %d\n\n", r.Meta.SymbolCount))

	// 符号级变更详情
	if len(r.SymbolChanges) > 0 {
		buffer.WriteString("=====================================\n")
		buffer.WriteString("符号级变更详情\n")
		buffer.WriteString("=====================================\n\n")

		// 按影响类型分组
		breakingChanges := filterByImpactType(r.SymbolChanges, ImpactTypeBreaking)
		internalChanges := filterByImpactType(r.SymbolChanges, ImpactTypeInternal)
		additiveChanges := filterByImpactType(r.SymbolChanges, ImpactTypeAdditive)

		if len(breakingChanges) > 0 {
			buffer.WriteString(fmt.Sprintf("🔴 破坏性变更 (%d)\n", len(breakingChanges)))
			for _, sc := range breakingChanges {
				buffer.WriteString(fmt.Sprintf("  • %s.%s (%s)\n",
					sc.ComponentName, sc.Symbol.Name, sc.Symbol.Kind))
			}
			buffer.WriteString("\n")
		}

		if len(internalChanges) > 0 {
			buffer.WriteString(fmt.Sprintf("🟡 内部变更 (%d)\n", len(internalChanges)))
			for _, sc := range internalChanges {
				buffer.WriteString(fmt.Sprintf("  • %s.%s (%s)\n",
					sc.ComponentName, sc.Symbol.Name, sc.Symbol.Kind))
			}
			buffer.WriteString("\n")
		}

		if len(additiveChanges) > 0 {
			buffer.WriteString(fmt.Sprintf("🟢 增强性变更 (%d)\n", len(additiveChanges)))
			for _, sc := range additiveChanges {
				buffer.WriteString(fmt.Sprintf("  • %s.%s (%s)\n",
					sc.ComponentName, sc.Symbol.Name, sc.Symbol.Kind))
			}
			buffer.WriteString("\n")
		}
	}

	// 受影响的组件
	buffer.WriteString("=====================================\n")
	buffer.WriteString("受影响的组件\n")
	buffer.WriteString("=====================================\n\n")

	sortedByRisk := r.sortByRisk()
	for _, item := range sortedByRisk {
		comp := item.comp
		buffer.WriteString(fmt.Sprintf("▶ %s (风险: %s, 层级: %d, 符号: %d)\n",
			comp.Name, comp.RiskLevel, comp.ImpactLevel, comp.SymbolCount))
		if len(comp.ChangePaths) > 0 {
			buffer.WriteString("  变更路径:\n")
			for _, path := range comp.ChangePaths {
				buffer.WriteString(fmt.Sprintf("    %s\n", path))
			}
		}
		buffer.WriteString("\n")
	}

	// 风险评估
	buffer.WriteString("=====================================\n")
	buffer.WriteString("风险评估\n")
	buffer.WriteString("=====================================\n\n")
	buffer.WriteString(fmt.Sprintf("整体风险: %s\n", r.RiskAssessment.OverallRisk))
	buffer.WriteString(fmt.Sprintf("  - 破坏性变更: %d\n", r.RiskAssessment.BreakingChange))
	buffer.WriteString(fmt.Sprintf("  - 内部变更: %d\n", r.RiskAssessment.InternalChange))
	buffer.WriteString(fmt.Sprintf("  - 增强性变更: %d\n\n", r.RiskAssessment.AdditiveChange))

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
func (r *AnalysisResult) sortByRisk() []struct {
	comp ImpactComponent
	risk int
} {
	riskOrder := map[string]int{"critical": 4, "high": 3, "medium": 2, "low": 1}

	items := make([]struct {
		comp ImpactComponent
		risk int
	}, len(r.Impact))

	for i, comp := range r.Impact {
		items[i].comp = comp
		items[i].risk = riskOrder[comp.RiskLevel]
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

// filterByImpactType 按影响类型过滤符号变更
func filterByImpactType(changes []SymbolImpactChange, impactType ImpactType) []SymbolImpactChange {
	result := make([]SymbolImpactChange, 0)
	for _, change := range changes {
		if change.ImpactType == impactType {
			result = append(result, change)
		}
	}
	return result
}
