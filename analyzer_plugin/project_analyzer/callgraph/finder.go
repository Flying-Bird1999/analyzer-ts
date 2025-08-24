package callgraph

import (
	"errors"
	"fmt"
	projectanalyzer "main/analyzer_plugin/project_analyzer"
	"path/filepath"
)

// Finder 是“查找调用方”分析器的实现。
type Finder struct {
	targetFile string
}

var _ projectanalyzer.Analyzer = (*Finder)(nil)

func (f *Finder) Name() string {
	return "find-callers"
}

func (f *Finder) Configure(params map[string]string) error {
	target, ok := params["targetFile"]
	if !ok || target == "" {
		return errors.New("缺少必需的参数 'targetFile'")
	}
	f.targetFile = target
	return nil
}

func (f *Finder) Analyze(ctx *projectanalyzer.ProjectContext) (projectanalyzer.Result, error) {
	if f.targetFile == "" {
		return nil, errors.New("分析前必须通过 Configure 方法或 -p 标志设置 'targetFile' 参数")
	}

	deps := ctx.ParsingResult

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

	lookupPath, err := filepath.Abs(f.targetFile)
	if err != nil {
		fmt.Printf("警告: 无法获取 %s 的绝对路径: %v", f.targetFile, err)
		lookupPath = f.targetFile
	}

	visited := make(map[string]bool)
	callTree := buildCallerTree(lookupPath, callerGraph, visited)

	affectedSet := make(map[string]struct{})
	collectAffectedFiles(&callTree, affectedSet)
	affectedList := setToSortedSlice(affectedSet)

	singleResult := SingleFileResult{
		Summary: SingleFileSummary{
			TargetFile:         f.targetFile,
			TotalAffectedFiles: len(affectedList),
			AffectedFilesList:  affectedList,
		},
		CallTree: callTree,
	}

	finalResult := &FindCallersResult{
		OverallSummary: OverallSummary{
			TargetFiles:        []string{f.targetFile},
			TotalAffectedFiles: len(affectedList),
			AffectedFilesList:  affectedList,
		},
		PerFileResults: []SingleFileResult{singleResult},
	}

	return finalResult, nil
}
