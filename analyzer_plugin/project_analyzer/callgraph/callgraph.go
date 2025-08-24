// package callgraph 实现了查找文件上游调用方的核心业务逻辑。
package callgraph

import (
	"sort"
	"strings"
)

// buildCallerTree 是一个递归辅助函数，用于为单个文件构建其上游调用树。
// 它通过深度优先搜索 (DFS) 遍历调用关系图来完成此任务。
// filePath: 当前正在构建树的节点所对应的文件路径。
// callerGraph: 完整的调用关系图。
// visited: 用于在遍历过程中检测循环依赖的访问记录图。
func buildCallerTree(filePath string, callerGraph map[string][]string, visited map[string]bool) CallerNode {
	// 如果在当前遍历路径中再次遇到同一个文件，说明存在循环依赖。
	if visited[filePath] {
		return CallerNode{FilePath: filePath + " (循环依赖)", Callers: []CallerNode{}}
	}
	visited[filePath] = true

	node := CallerNode{
		FilePath: filePath,
		Callers:  []CallerNode{},
	}

	// 查找所有直接引用了当前文件的上游文件，并为它们递归地构建子树。
	if callers, exists := callerGraph[filePath]; exists {
		for _, callerPath := range callers {
			node.Callers = append(node.Callers, buildCallerTree(callerPath, callerGraph, visited))
		}
	}

	// 回溯时，将当前节点从“已访问”记录中移除，以便其他遍历路径可以正常访问它。
	delete(visited, filePath)

	return node
}

// collectAffectedFiles 递归地遍历调用树，以收集一组唯一的文件路径。
// node: 当前遍历到的调用树节点。
// visited: 用于存储所有遇到过的文件路径，利用 map 的特性实现自动去重。
func collectAffectedFiles(node *CallerNode, visited map[string]struct{}) {
	if node == nil {
		return
	}
	// 忽略标记为循环依赖的伪节点。
	if !strings.HasSuffix(node.FilePath, " (循环依赖)") {
		visited[node.FilePath] = struct{}{}
	}
	for i := range node.Callers {
		collectAffectedFiles(&node.Callers[i], visited)
	}
}

// setToSortedSlice 将一个 map[string]struct{} (模拟的集合) 转换为一个经过排序的字符串切片。
func setToSortedSlice(set map[string]struct{}) []string {
	slice := make([]string, 0, len(set))
	for item := range set {
		slice = append(slice, item)
	}
	sort.Strings(slice)
	return slice
}
