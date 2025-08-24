package callgraph

import (
	"errors"
	"fmt"
	projectanalyzer "main/analyzer_plugin/project_analyzer"
	"path/filepath"
	"strings"
)

// Finder 是"查找调用方"分析器的实现。
type Finder struct {
	targetFiles []string // 存储目标文件路径的切片，支持多个文件
}

// 确保 Finder 实现了 projectanalyzer.Analyzer 接口
var _ projectanalyzer.Analyzer = (*Finder)(nil)

// Name 返回分析器的名称
func (f *Finder) Name() string {
	return "find-callers"
}

// Configure 配置分析器参数
// 参数 params 是一个 map，键为参数名，值为参数值（可能包含多个值，用逗号分隔）
// 例如: {"targetFiles": "/path/to/file1.ts,/path/to/file2.ts"}
func (f *Finder) Configure(params map[string]string) error {
	// 获取 targetFiles 参数
	targetsStr, ok := params["targetFiles"]
	if !ok || targetsStr == "" {
		return errors.New("缺少必需的参数 'targetFiles' (逗号分隔的文件路径)")
	}

	// 将逗号分隔的字符串拆分为文件路径切片
	splitTargets := strings.Split(targetsStr, ",")

	// 遍历拆分后的路径，去除空格并添加到 targetFiles 切片中
	for _, path := range splitTargets {
		trimmedPath := strings.TrimSpace(path)
		if trimmedPath != "" { // 避免添加空字符串
			f.targetFiles = append(f.targetFiles, trimmedPath)
		}
	}

	return nil
}

// Analyze 执行分析逻辑
// ctx 包含项目上下文信息，如项目根路径、排除路径和解析结果
func (f *Finder) Analyze(ctx *projectanalyzer.ProjectContext) (projectanalyzer.Result, error) {
	// 检查是否已配置目标文件
	if len(f.targetFiles) == 0 {
		return nil, errors.New("分析前必须通过 Configure 方法或 -p 标志设置 'targetFiles' 参数")
	}

	// 获取解析结果
	deps := ctx.ParsingResult

	// 构建调用图：map[被调用文件路径][]调用方文件路径
	callerGraph := make(map[string][]string)
	for callerPathRaw, fileDeps := range deps.Js_Data {
		// 标准化调用方路径
		callerPath := filepath.ToSlash(filepath.Clean(callerPathRaw))

		// 遍历导入声明
		for _, dep := range fileDeps.ImportDeclarations {
			if dep.Source.FilePath != "" {
				// 标准化被导入文件路径
				normalizedDepPath := filepath.ToSlash(filepath.Clean(dep.Source.FilePath))
				// 将调用方添加到被调用文件的调用方列表中
				callerGraph[normalizedDepPath] = append(callerGraph[normalizedDepPath], callerPath)
			}
		}

		// 遍历导出声明
		for _, dep := range fileDeps.ExportDeclarations {
			if dep.Source != nil && dep.Source.FilePath != "" {
				// 标准化被导出文件路径
				normalizedDepPath := filepath.ToSlash(filepath.Clean(dep.Source.FilePath))
				// 将调用方添加到被调用文件的调用方列表中
				callerGraph[normalizedDepPath] = append(callerGraph[normalizedDepPath], callerPath)
			}
		}
	}

	// 存储所有文件的分析结果
	var allPerFileResults []SingleFileResult
	// 用于收集所有受影响的唯一文件
	allAffectedFilesSet := make(map[string]struct{})

	// 遍历所有目标文件
	for _, targetFile := range f.targetFiles {
		// 获取目标文件的绝对路径
		absPath, err := filepath.Abs(targetFile)
		if err != nil {
			fmt.Printf("警告: 无法获取 %s 的绝对路径: %v\n", targetFile, err)
			continue // 跳过此文件
		}
		lookupPath := filepath.Clean(absPath) // 标准化路径

		// 为每个文件创建一个新的 visited map，防止不同文件间的调用链互相干扰
		visited := make(map[string]bool)
		// 构建调用树
		callTree := buildCallerTree(lookupPath, callerGraph, visited)

		// 收集受影响的文件
		affectedSet := make(map[string]struct{})
		collectAffectedFiles(&callTree, affectedSet)
		affectedList := setToSortedSlice(affectedSet)

		// 创建单个文件的分析结果
		singleResult := SingleFileResult{
			Summary: SingleFileSummary{
				TargetFile:         targetFile,
				TotalAffectedFiles: len(affectedList),
				AffectedFilesList:  affectedList,
			},
			CallTree: callTree,
		}
		allPerFileResults = append(allPerFileResults, singleResult)

		// 将当前文件的受影响文件合并到总的受影响文件集合中
		for file := range affectedSet {
			allAffectedFilesSet[file] = struct{}{}
		}
	}

	// 创建最终结果
	finalResult := &FindCallersResult{
		OverallSummary: OverallSummary{
			TargetFiles:        f.targetFiles,                         // 所有目标文件
			TotalAffectedFiles: len(allAffectedFilesSet),              // 所有文件的总和
			AffectedFilesList:  setToSortedSlice(allAffectedFilesSet), // 所有文件的唯一列表
		},
		PerFileResults: allPerFileResults, // 包含每个文件的结果
	}

	return finalResult, nil
}
