// Package mr_component_impact 提供 MR 组件影响分析功能
package mr_component_impact

import (
	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/component_deps"
)

// =============================================================================
// 组件影响分析器
// =============================================================================

// ComponentImpactAnalyzer 组件影响分析器
// 基于 component_deps 的结果分析组件变更的影响
type ComponentImpactAnalyzer struct {
	componentDeps *component_deps.ComponentDepsResult
}

// NewComponentImpactAnalyzer 创建组件影响分析器
func NewComponentImpactAnalyzer(
	componentDeps *component_deps.ComponentDepsResult,
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
	visited := make(map[string]bool) // 已访问的组件

	// BFS 传播节点：记录组件名、原始源头、当前层级、传播路径
	type propagationNode struct {
		name   string
		source string // 原始变更源头
		level  int    // 当前层级
		path   []string
	}

	queue := []propagationNode{
		{name: componentName, source: componentName, level: 0, path: []string{componentName}},
	}
	visited[componentName] = true

	// BFS 遍历依赖链
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// 查找所有依赖当前组件的组件（下游组件）
		for compName, compInfo := range a.componentDeps.Components {
			// 跳过已访问的组件
			if visited[compName] {
				continue
			}

			// 检查该组件是否依赖当前组件
			for _, dep := range compInfo.ComponentDeps {
				if dep.Name == current.name {
					// 构建传播路径
					newPath := append([]string{}, current.path...)
					newPath = append(newPath, compName)

					// 确定关系类型
					relation := RelationDepends
					if current.level > 0 {
						relation = RelationIndirect
					}

					// 记录影响
					impact := ComponentImpact{
						Component:    compName,
						ChangeSource: current.source, // 始终是原始变更源头
						Relation:     relation,
						Level:        current.level + 1,
					}

					// 对于间接依赖，添加传播路径
					if current.level > 0 {
						impact.Path = newPath
					}

					impacts = append(impacts, impact)

					// 标记为已访问并加入队列
					visited[compName] = true
					queue = append(queue, propagationNode{
						name:   compName,
						source: current.source, // 保持原始源头
						level:  current.level + 1,
						path:   newPath,
					})
					break
				}
			}
		}
	}

	return impacts
}
