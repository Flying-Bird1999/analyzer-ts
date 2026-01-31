package component_deps_v2

import (
	"sort"

	"github.com/samber/lo"
)

// =============================================================================
// 依赖图构建器
// =============================================================================

// GraphBuilder 依赖图构建器
type GraphBuilder struct {
	manifest *ComponentManifest
}

// NewGraphBuilder 创建依赖图构建器
func NewGraphBuilder(manifest *ComponentManifest) *GraphBuilder {
	return &GraphBuilder{
		manifest: manifest,
	}
}

// BuildDepGraph 构建正向依赖图
// 输入: 组件名 -> 依赖组件列表
// 输出: DependencyGraph (key: 组件名, value: 该组件依赖的组件列表)
func (gb *GraphBuilder) BuildDepGraph(dependencies map[string][]string) DependencyGraph {
	// 创建依赖图的副本，确保所有组件都在图中
	graph := make(DependencyGraph)

	// 确保所有组件都在图中（即使没有依赖）
	for _, comp := range gb.manifest.Components {
		graph[comp.Name] = []string{}
	}

	// 填充依赖关系
	for compName, deps := range dependencies {
		// 排序依赖列表，保证输出一致性
		sortedDeps := sortAndUnique(deps)
		graph[compName] = sortedDeps
	}

	return graph
}

// BuildRevDepGraph 构建反向依赖图
// 输入: 正向依赖图
// 输出: ReverseDepGraph (key: 组件名, value: 依赖该组件的其他组件列表)
func (gb *GraphBuilder) BuildRevDepGraph(depGraph DependencyGraph) ReverseDepGraph {
	revGraph := make(ReverseDepGraph)

	// 初始化反向依赖图
	for _, comp := range gb.manifest.Components {
		revGraph[comp.Name] = []string{}
	}

	// 遍历正向依赖图，构建反向关系
	for compName, deps := range depGraph {
		for _, depName := range deps {
			// compName 依赖 depName
			// 所以在反向图中，depName 被 compName 依赖
			revGraph[depName] = append(revGraph[depName], compName)
		}
	}

	// 排序反向依赖列表
	for compName := range revGraph {
		revGraph[compName] = sortAndUnique(revGraph[compName])
	}

	return revGraph
}

// BuildComponentInfo 构建组件信息映射
// 输入: 正向依赖图
// 输出: 组件名 -> ComponentInfo
func (gb *GraphBuilder) BuildComponentInfo(depGraph DependencyGraph) map[string]ComponentInfo {
	result := make(map[string]ComponentInfo)

	for _, comp := range gb.manifest.Components {
		result[comp.Name] = ComponentInfo{
			Name:         comp.Name,
			Entry:        comp.Entry,
			Dependencies: depGraph[comp.Name],
		}
	}

	return result
}

// =============================================================================
// 辅助函数
// =============================================================================

// sortAndUnique 排序并去重
func sortAndUnique(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}

	// 去重
	unique := lo.Uniq(items)

	// 排序
	sort.Strings(unique)

	return unique
}

// DetectCycles 检测循环依赖
// 使用 DFS 算法检测依赖图中的环
func (gb *GraphBuilder) DetectCycles(depGraph DependencyGraph) [][]string {
	cycles := [][]string{}
	visited := make(map[string]bool)
	inPath := make(map[string]bool)
	path := []string{}

	for compName := range depGraph {
		if !visited[compName] {
			cycle := gb.detectCyclesDFS(compName, depGraph, visited, inPath, path)
			if len(cycle) > 0 {
				cycles = append(cycles, cycle)
			}
		}
	}

	return cycles
}

// detectCyclesDFS 使用 DFS 检测循环依赖
func (gb *GraphBuilder) detectCyclesDFS(
	node string,
	graph DependencyGraph,
	visited map[string]bool,
	inPath map[string]bool,
	path []string,
) []string {
	visited[node] = true
	inPath[node] = true
	path = append(path, node)

	for _, dep := range graph[node] {
		if inPath[dep] {
			// 发现环
			cycleStart := -1
			for i, p := range path {
				if p == dep {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cycle := append([]string{}, path[cycleStart:]...)
				cycle = append(cycle, dep) // 闭合环
				return cycle
			}
		} else if !visited[dep] {
			if cycle := gb.detectCyclesDFS(dep, graph, visited, inPath, path); len(cycle) > 0 {
				return cycle
			}
		}
	}

	inPath[node] = false
	return nil
}

// GetDependencyLevel 获取组件的依赖层级
// 返回值: 层级数（0 表示无依赖，1 表示依赖一层，以此类推）
func (gb *GraphBuilder) GetDependencyLevel(compName string, depGraph DependencyGraph) int {
	// 使用 BFS 计算最长路径长度
	levels := make(map[string]int)
	queue := []string{}

	// 找到所有没有依赖的组件，作为起始点
	for name, deps := range depGraph {
		if len(deps) == 0 {
			levels[name] = 0
			queue = append(queue, name)
		}
	}

	// BFS
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// 找到所有依赖当前组件的其他组件
		for name, deps := range depGraph {
			if lo.Contains(deps, current) {
				// 更新层级
				newLevel := levels[current] + 1
				if newLevel > levels[name] {
					levels[name] = newLevel
					queue = append(queue, name)
				}
			}
		}
	}

	return levels[compName]
}
