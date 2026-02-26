// Package mr_component_impact 提供 MR 组件影响分析功能
package mr_component_impact

import (
	"fmt"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/component_deps_v2"
)

// =============================================================================
// 组件影响分析器
// =============================================================================

// ComponentImpactAnalyzer 组件影响分析器
// 基于 component_deps_v2 的结果分析组件变更的影响
type ComponentImpactAnalyzer struct {
	componentDeps *component_deps_v2.ComponentDepsV2Result
}

// NewComponentImpactAnalyzer 创建组件影响分析器
func NewComponentImpactAnalyzer(
	componentDeps *component_deps_v2.ComponentDepsV2Result,
) *ComponentImpactAnalyzer {
	return &ComponentImpactAnalyzer{
		componentDeps: componentDeps,
	}
}

// AnalyzeComponentChange 分析组件变更的影响
// 返回受影响的所有组件信息列表
func (a *ComponentImpactAnalyzer) AnalyzeComponentChange(
	componentName string,
) []ComponentImpact {
	if a.componentDeps == nil {
		return nil
	}

	impacts := make([]ComponentImpact, 0)

	// 遍历所有组件，查找依赖该组件的组件
	for compName, compInfo := range a.componentDeps.Components {
		// 跳过自身
		if compName == componentName {
			continue
		}

		// 检查该组件是否依赖变更的组件
		for _, dep := range compInfo.ComponentDeps {
			if dep.Name == componentName {
				impacts = append(impacts, ComponentImpact{
					ComponentName: compName,
					ImpactReason:  fmt.Sprintf("依赖组件 %s", componentName),
					ChangeType:    "component",
					ChangeSource:  componentName,
				})
				break
			}
		}
	}

	return impacts
}
