// package callgraph 实现了查找文件上游调用方的核心业务逻辑。
package callgraph

import (
	"fmt"
	"main/analyzer_plugin/project_analyzer/internal/parser"
	"path/filepath"
	"sort"
	"strings"
)

// Find 分析一个项目，为一组目标文件查找所有上游的调用方。
// 它首先通过底层解析器构建出整个项目的依赖关系，然后反转这个关系图来构建调用图，
// 最后从每个目标文件出发，递归地遍历调用图，从而构建出完整的调用树并汇总结果。
// params: 包含分析所需所有参数的结构体。
// returns: 返回一个包含详细调用链信息的结构化结果，或在发生错误时返回 error。
func Find(params Params) (*Result, error) {
	// 步骤 1: 使用新的 parser 包分析整个项目，以构建一个完整的依赖关系图。
	deps, err := parser.ParseProject(params.RootPath, params.Exclude, params.IsMonorepo)
	if err != nil {
		return nil, fmt.Errorf("分析项目失败: %w", err)
	}

	// 步骤 2: 构建调用关系图 (callerGraph)，这是一个反向的依赖关系图。
	// 键 (key) 是被引用的文件，值 (value) 是引用了该文件的文件列表。
	callerGraph := make(map[string][]string)
	for callerPath, fileDeps := range deps.Js_Data {
		for _, dep := range fileDeps.ImportDeclarations {
			if dep.Source.FilePath != "" {
				callerGraph[dep.Source.FilePath] = append(callerGraph[dep.Source.FilePath], callerPath)
			}
		}
		for _, dep := range fileDeps.ExportDeclarations {
			if dep.Source != nil && dep.Source.FilePath != "" {
				callerGraph[dep.Source.FilePath] = append(callerGraph[dep.Source.FilePath], callerPath)
			}
		}
	}

	// 步骤 3: 循环处理每一个目标文件，并最终将所有结果进行汇总。
	perFileResults := make([]SingleFileResult, 0, len(params.TargetFiles))
	overallAffectedSet := make(map[string]struct{}) // 使用 map 来存储所有受影响的文件，以实现自动去重。

	for _, targetFile := range params.TargetFiles {
		lookupPath, err := filepath.Abs(targetFile)
		if err != nil {
			fmt.Printf("警告: 无法获取 %s 的绝对路径: %v", targetFile, err)
			lookupPath = targetFile
		}

		// 为当前目标文件构建其上游调用树。
		visited := make(map[string]bool)
		callTree := buildCallerTree(lookupPath, callerGraph, visited)

		// 从调用树中收集所有受影响的文件列表。
		affectedSet := make(map[string]struct{})
		collectAffectedFiles(&callTree, affectedSet)
		affectedList := setToSortedSlice(affectedSet)

		// 将当前文件所影响的范围合并到总的影响范围中。
		for path := range affectedSet {
			overallAffectedSet[path] = struct{}{}
		}

		// 为当前目标文件创建独立的分析结果。
		singleResult := SingleFileResult{
			Summary: SingleFileSummary{
				TargetFile:         targetFile,
				TotalAffectedFiles: len(affectedList),
				AffectedFilesList:  affectedList,
			},
			CallTree: callTree,
		}
		perFileResults = append(perFileResults, singleResult)
	}

	// 步骤 4: 创建最终的、包含所有目标文件信息的聚合结果。
	overallAffectedList := setToSortedSlice(overallAffectedSet)
	finalResult := &Result{
		OverallSummary: OverallSummary{
			TargetFiles:        params.TargetFiles,
			TotalAffectedFiles: len(overallAffectedList),
			AffectedFilesList:  overallAffectedList,
		},
		PerFileResults: perFileResults,
	}

	return finalResult, nil
}

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