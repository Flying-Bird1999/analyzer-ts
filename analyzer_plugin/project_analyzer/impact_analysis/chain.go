package impact_analysis

import (
	"strings"
)

// =============================================================================
// 依赖链路构建
// =============================================================================

// ChainBuilder 依赖链路构建器
type ChainBuilder struct {
	depGraph    map[string][]string
	revDepGraph map[string][]string
}

// NewChainBuilder 创建依赖链路构建器
func NewChainBuilder(depGraph, revDepGraph map[string][]string) *ChainBuilder {
	return &ChainBuilder{
		depGraph:    depGraph,
		revDepGraph: revDepGraph,
	}
}

// BuildImpactChains 构建从变更组件到受影响组件的所有链路
func (cb *ChainBuilder) BuildImpactChains(
	changedComponents []string,
	impactedComponents []string,
) [][]string {
	chains := make([][]string, 0)

	for _, changed := range changedComponents {
		for _, impacted := range impactedComponents {
			if changed == impacted {
				// 直接变更的组件
				chains = append(chains, []string{changed})
				continue
			}
			// 查找从 changed 到 impacted 的所有路径
			paths := cb.findAllPaths(changed, impacted, make(map[string]bool), []string{})
			chains = append(chains, paths...)
		}
	}

	return chains
}

// findAllPaths 查找从 start 到 end 的所有路径（DFS）
func (cb *ChainBuilder) findAllPaths(
	start string,
	end string,
	visited map[string]bool,
	currentPath []string,
) [][]string {
	// 避免循环
	if visited[start] {
		return nil
	}

	// 添加当前节点到路径
	newPath := append(currentPath, start)
	visited[start] = true

	// 找到目标
	if start == end {
		// 复制 visited 避免影响其他路径
		newVisited := make(map[string]bool)
		for k, v := range visited {
			newVisited[k] = v
		}
		newVisited[start] = false // 回溯
		return [][]string{newPath}
	}

	// 继续遍历下游组件
	paths := make([][]string, 0)
	downstream := cb.revDepGraph[start]
	for _, next := range downstream {
		subPaths := cb.findAllPaths(next, end, visited, newPath)
		paths = append(paths, subPaths...)
	}

	return paths
}

// GetCriticalPath 获取关键路径（影响层级最深的路径）
func (cb *ChainBuilder) GetCriticalPath(
	changedComponent string,
	impactedComponent string,
) []string {
	paths := cb.findAllPaths(changedComponent, impactedComponent, make(map[string]bool), []string{})

	if len(paths) == 0 {
		return nil
	}

	// 返回最长的路径（影响层级最深）
	criticalPath := paths[0]
	for _, path := range paths {
		if len(path) > len(criticalPath) {
			criticalPath = path
		}
	}

	return criticalPath
}

// FormatPath 格式化路径为字符串
func FormatPath(path []string) string {
	if len(path) == 0 {
		return ""
	}
	return strings.Join(path, " → ")
}

// =============================================================================
// 环检测
// =============================================================================

// DetectCycles 检测依赖图中的循环依赖
func (cb *ChainBuilder) DetectCycles() [][]string {
	cycles := make([][]string, 0)
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	for component := range cb.depGraph {
		if !visited[component] {
			if cycle := cb.detectCycleDFS(component, visited, recursionStack, []string{}); cycle != nil {
				cycles = append(cycles, cycle)
			}
		}
	}

	return cycles
}

// detectCycleDFS 使用 DFS 检测循环
func (cb *ChainBuilder) detectCycleDFS(
	component string,
	visited map[string]bool,
	recursionStack map[string]bool,
	path []string,
) []string {
	visited[component] = true
	recursionStack[component] = true
	newPath := append(path, component)

	// 遍历所有依赖
	for _, dep := range cb.depGraph[component] {
		if !visited[dep] {
			if cycle := cb.detectCycleDFS(dep, visited, recursionStack, newPath); cycle != nil {
				return cycle
			}
		} else if recursionStack[dep] {
			// 找到循环，提取循环部分
			cycleStart := -1
			for i, p := range newPath {
				if p == dep {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cycle := append(newPath[cycleStart:], dep)
				return cycle
			}
		}
	}

	recursionStack[component] = false
	return nil
}

// HasCycle 检查是否存在循环依赖
func (cb *ChainBuilder) HasCycle() bool {
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	for component := range cb.depGraph {
		if !visited[component] {
			if cb.hasCycleDFS(component, visited, recursionStack) {
				return true
			}
		}
	}

	return false
}

// hasCycleDFS DFS 检测循环（返回 bool）
func (cb *ChainBuilder) hasCycleDFS(
	component string,
	visited map[string]bool,
	recursionStack map[string]bool,
) bool {
	visited[component] = true
	recursionStack[component] = true

	for _, dep := range cb.depGraph[component] {
		if !visited[dep] {
			if cb.hasCycleDFS(dep, visited, recursionStack) {
				return true
			}
		} else if recursionStack[dep] {
			return true
		}
	}

	recursionStack[component] = false
	return false
}
