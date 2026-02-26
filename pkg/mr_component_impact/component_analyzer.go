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
// 返回受影响的所有组件信息列表（包括传递依赖）
func (a *ComponentImpactAnalyzer) AnalyzeComponentChange(
	componentName string,
) []ComponentImpact {
	if a.componentDeps == nil {
		return nil
	}

	impacts := make([]ComponentImpact, 0)
	visited := make(map[string]bool)     // 已访问的组件
	queue := []string{componentName}      // BFS 队列
	sourceMap := make(map[string]string)  // 记录每个组件的影响来源

	visited[componentName] = true
	sourceMap[componentName] = componentName

	// BFS 遍历依赖链
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// 查找所有依赖当前组件的组件
		for compName, compInfo := range a.componentDeps.Components {
			// 跳过已访问的组件
			if visited[compName] {
				continue
			}

			// 检查该组件是否依赖当前组件
			for _, dep := range compInfo.ComponentDeps {
				if dep.Name == current {
					// 记录影响
					source := sourceMap[current]
					impacts = append(impacts, ComponentImpact{
						ComponentName: compName,
						ImpactReason:  fmt.Sprintf("依赖组件 %s", current),
						ChangeType:    "component",
						ChangeSource:  source,
					})

					// 标记为已访问并加入队列
					visited[compName] = true
					sourceMap[compName] = source
					queue = append(queue, compName)
					break
				}
			}
		}
	}

	return impacts
}
