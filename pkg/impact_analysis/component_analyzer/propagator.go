// Package component_analyzer 提供组件级影响分析功能。
// 这是组件库专用能力，基于 file_analyzer 的结果进行组件映射。
package component_analyzer

import (
	"container/list"

	"github.com/Flying-Bird1999/analyzer-ts/pkg/impact_analysis"
)

// =============================================================================
// 组件级影响传播器
// =============================================================================

// Propagator 组件级影响传播器
// 使用 BFS 算法沿组件依赖链传播影响
type Propagator struct {
	depGraph *ComponentDependencyGraph
	maxDepth int
}

// NewPropagator 创建组件影响传播器
func NewPropagator(depGraph *ComponentDependencyGraph, maxDepth int) *Propagator {
	if maxDepth <= 0 {
		maxDepth = 10 // 默认最大深度
	}
	return &Propagator{
		depGraph: depGraph,
		maxDepth: maxDepth,
	}
}

// Propagate 传播组件影响
// 输入：直接变更的组件列表
// 输出：受影响的组件集合
func (p *Propagator) Propagate(changedComponents []string) *ImpactedComponents {
	result := &ImpactedComponents{
		Direct:      make(map[string]*ComponentImpact),
		Indirect:    make(map[string]*ComponentImpact),
		ImpactPaths: make([]ComponentImpactPath, 0),
	}

	if len(changedComponents) == 0 {
		return result
	}

	// 步骤 1: 标记直接变更的组件
	for _, componentName := range changedComponents {
		result.Direct[componentName] = &ComponentImpact{
			ComponentName: componentName,
			ImpactLevel:   impact_analysis.ImpactLevelDirect,
			ChangePaths:   []string{componentName},
		}
	}

	// 步骤 2: BFS 传播影响
	p.bfsPropagation(changedComponents, result)

	return result
}

// bfsPropagation 广度优先搜索传播影响
func (p *Propagator) bfsPropagation(
	changedComponents []string,
	result *ImpactedComponents,
) {
	queue := list.New()
	visited := make(map[string]bool)

	// 初始化队列：将所有变更组件加入队列
	for _, comp := range changedComponents {
		queue.PushBack(&componentPropagationNode{
			componentName: comp,
			path:          []string{comp},
			depth:         0,
		})
		visited[comp] = true
	}

	// BFS 遍历
	for queue.Len() > 0 {
		current := queue.Remove(queue.Front()).(*componentPropagationNode)

		// 检查深度限制
		if current.depth >= p.maxDepth {
			continue
		}

		// 获取依赖当前组件的其他组件（下游组件）
		downstreamComponents := p.depGraph.GetDependants(current.componentName)
		if len(downstreamComponents) == 0 {
			continue
		}

		for _, downstream := range downstreamComponents {
			newPath := append(current.path, downstream)

			// 更新或创建影响记录
			if existing, exists := result.Indirect[downstream]; exists {
				// 添加新的传播路径
				existing.ChangePaths = append(existing.ChangePaths, formatPath(newPath))
				// 更新影响层级（取最小值）
				if existing.ImpactLevel < impact_analysis.ImpactLevel(current.depth+1) {
					existing.ImpactLevel = impact_analysis.ImpactLevel(current.depth + 1)
				}
			} else {
				// 创建新的影响记录
				result.Indirect[downstream] = &ComponentImpact{
					ComponentName: downstream,
					ImpactLevel:   impact_analysis.ImpactLevel(current.depth + 1),
					ChangePaths:   []string{formatPath(newPath)},
				}
			}

			// 将下游组件加入队列（如果未访问过）
			if !visited[downstream] {
				queue.PushBack(&componentPropagationNode{
					componentName: downstream,
					path:          newPath,
					depth:         current.depth + 1,
				})
				visited[downstream] = true
			}
		}
	}
}

// componentPropagationNode 传播节点（用于 BFS）
type componentPropagationNode struct {
	componentName string
	path          []string
	depth         int
}

// =============================================================================
// 受影响的组件集合
// =============================================================================

// ImpactedComponents 受影响的组件集合
type ImpactedComponents struct {
	// Direct 直接变更的组件
	Direct map[string]*ComponentImpact

	// Indirect 间接受影响的组件
	Indirect map[string]*ComponentImpact

	// ImpactPaths 影响路径列表
	ImpactPaths []ComponentImpactPath
}

// ComponentImpact 组件影响信息
type ComponentImpact struct {
	ComponentName string                      // 组件名称
	ImpactLevel   impact_analysis.ImpactLevel // 影响层级（0=直接变更，>0=传播层级）
	ChangePaths   []string                    // 从变更源头到该组件的路径
}

// ComponentImpactPath 组件影响路径
type ComponentImpactPath struct {
	SourceComponent string   // 变更源头组件
	TargetComponent string   // 受影响组件
	Path            []string // 影响传播路径
}

// GetDirectChangedComponents 获取直接变更的组件列表
func (c *ImpactedComponents) GetDirectChangedComponents() []string {
	components := make([]string, 0, len(c.Direct))
	for compName := range c.Direct {
		components = append(components, compName)
	}
	return components
}

// GetImpactedComponents 获取所有受影响的组件（包括直接和间接）
func (c *ImpactedComponents) GetImpactedComponents() []string {
	components := make([]string, 0, len(c.Direct)+len(c.Indirect))
	for compName := range c.Direct {
		components = append(components, compName)
	}
	for compName := range c.Indirect {
		components = append(components, compName)
	}
	return components
}

// GetComponentImpact 获取指定组件的影响信息
func (c *ImpactedComponents) GetComponentImpact(componentName string) (*ComponentImpact, bool) {
	if impact, exists := c.Direct[componentName]; exists {
		return impact, true
	}
	if impact, exists := c.Indirect[componentName]; exists {
		return impact, true
	}
	return nil, false
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
// ComponentDependencyGraph 方法扩展
// =============================================================================

// GetDependants 获取依赖指定组件的所有组件
func (g *ComponentDependencyGraph) GetDependants(componentName string) []string {
	return g.RevDepGraph[componentName]
}

// GetDependencies 获取指定组件依赖的所有组件
func (g *ComponentDependencyGraph) GetDependencies(componentName string) []string {
	return g.DepGraph[componentName]
}

// GetSymbolImports 获取指定组件的符号导入
func (g *ComponentDependencyGraph) GetSymbolImports(componentName string) []ComponentSymbolImport {
	return g.SymbolImports[componentName]
}
