// package cmd 包含了所有命令行工具的定义
package cmd

// example： go run main.go find-callers -i /Users/bird/company/sc1.0/live/shopline-live-sale -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin/project_analyzer/result -x "examples/**" -x "tests/**" -f /Users/bird/company/sc1.0/live/shopline-live-sale/src/utils/downloadFile.ts -f /Users/bird/company/sc1.0/live/shopline-live-sale/src/utils/string-utils.ts

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	// 引入我们核心的项目分析器包
	project_analyzer "main/analyzer_plugin/project_analyzer"

	"github.com/spf13/cobra"
)

// --- 新的数据结构定义 ---

// FindCallersOverallResult 是用于多文件分析的顶级结构
type FindCallersOverallResult struct {
	OverallSummary OverallSummary     `json:"overallSummary"`
	PerFileResults []SingleFileResult `json:"perFileResults"`
}

// OverallSummary 包含所有目标文件的聚合汇总数据
type OverallSummary struct {
	TargetFiles        []string `json:"targetFiles"`
	TotalAffectedFiles int      `json:"totalAffectedFiles"`
	AffectedFilesList  []string `json:"affectedFilesList"`
}

// SingleFileResult 代表对单个目标文件的完整分析结果
type SingleFileResult struct {
	Summary  SingleFileSummary `json:"summary"`
	CallTree CallerNode        `json:"callTree"`
}

// SingleFileSummary 包含单个目标文件的汇总数据
type SingleFileSummary struct {
	TargetFile         string   `json:"targetFile"`
	TotalAffectedFiles int      `json:"totalAffectedFiles"`
	AffectedFilesList  []string `json:"affectedFilesList"`
}

// CallerNode 定义了调用树的数据结构 (保持不变)
type CallerNode struct {
	FilePath string       `json:"filePath"`
	Callers  []CallerNode `json:"callers"`
}

// 定义该命令所需的标志变量
var (
	findCallersInputDir    string
	findCallersTargetFiles []string // 修改为字符串切片以接受多个文件
	findCallersOutputDir   string
	findCallersExclude     []string
	findCallersIsMonorepo  bool
)

// NewFindCallersCmd 创建并返回 `find-callers` 命令
func NewFindCallersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "find-callers",
		Short: "查找一个或多个指定文件的所有上游调用方",
		Long:  `该命令首先分析 --input 指定的 TypeScript 项目以构建完整的依赖关系图，然后追踪 --file 指定的一个或多个文件的上游调用链路，并以 JSON 格式输出结果，其中包含每个文件的独立报告和所有文件的最终汇总。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 步骤 1: 分析项目，构建完整的依赖图
			analyzer := project_analyzer.NewProjectAnalyzer(findCallersInputDir, findCallersExclude, findCallersIsMonorepo)
			deps, err := analyzer.Analyze()
			if err != nil {
				return fmt.Errorf("分析项目失败: %w", err)
			}

			// 步骤 2: 构建调用关系图 (反向依赖图)
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

			// 步骤 3: 循环处理每个目标文件，并汇总结果
			perFileResults := make([]SingleFileResult, 0, len(findCallersTargetFiles))
			overallAffectedSet := make(map[string]struct{})

			for _, targetFile := range findCallersTargetFiles {
				lookupPath, err := filepath.Abs(targetFile)
				if err != nil {
					fmt.Printf("警告: 无法获取 %s 的绝对路径: %v\n", targetFile, err)
					lookupPath = targetFile
				}

				// 为每个文件构建调用树
				visited := make(map[string]bool)
				callTree := buildCallerTree(lookupPath, callerGraph, visited)

				// 为单个文件收集受影响的文件列表
				affectedSet := make(map[string]struct{})
				collectAffectedFiles(&callTree, affectedSet)
				affectedList := setToSortedSlice(affectedSet)

				// 将当前文件的受影响列表合并到总列表中
				for path := range affectedSet {
					overallAffectedSet[path] = struct{}{}
				}

				// 创建单个文件的结果
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

			// 步骤 4: 创建最终的汇总结果
			overallAffectedList := setToSortedSlice(overallAffectedSet)
			finalResult := FindCallersOverallResult{
				OverallSummary: OverallSummary{
					TargetFiles:        findCallersTargetFiles,
					TotalAffectedFiles: len(overallAffectedList),
					AffectedFilesList:  overallAffectedList,
				},
				PerFileResults: perFileResults,
			}

			// 步骤 5: 将结果格式化为 JSON 并输出
			outputBytes, err := json.MarshalIndent(finalResult, "", "  ")
			if err != nil {
				return fmt.Errorf("无法将结果序列化为 JSON: %w", err)
			}

			if findCallersOutputDir != "" {
				if err := os.MkdirAll(findCallersOutputDir, os.ModePerm); err != nil {
					return fmt.Errorf("无法创建输出目录 %s: %w", findCallersOutputDir, err)
				}
				// 使用输入目录的名称作为输出文件名，并加上命令名称
				baseName := filepath.Base(findCallersInputDir)
				outputFileName := fmt.Sprintf("%s_find_callers.json", baseName)
				fullOutputPath := filepath.Join(findCallersOutputDir, outputFileName)

				if err := ioutil.WriteFile(fullOutputPath, outputBytes, 0644); err != nil {
					return fmt.Errorf("无法将输出写入文件 %s: %w", fullOutputPath, err)
				}
				fmt.Printf("调用链分析结果已写入: %s\n", fullOutputPath)
			} else {
				fmt.Println(string(outputBytes))
			}

			return nil
		},
	}

	// 定义所有标志
	cmd.Flags().StringVarP(&findCallersInputDir, "input", "i", "", "要分析的 TypeScript 项目目录的路径 (必需)")
	// 修改为 StringSliceVarP 以支持多个 --file 参数
	cmd.Flags().StringSliceVarP(&findCallersTargetFiles, "file", "f", []string{}, "要追踪其调用链的文件路径 (必需, 可多次使用)")
	cmd.Flags().StringVarP(&findCallersOutputDir, "output", "o", "", "用于存储 JSON 结果的输出目录路径 (可选, 默认为标准输出)")
	cmd.Flags().StringSliceVarP(&findCallersExclude, "exclude", "x", []string{}, "要从分析中排除的 Glob 模式 (可多次使用)")
	cmd.Flags().BoolVarP(&findCallersIsMonorepo, "monorepo", "m", false, "如果分析的是 monorepo 项目，请设置为 true")

	// 将 --input 和 --file 标记为必需
	cmd.MarkFlagRequired("input")
	cmd.MarkFlagRequired("file")

	return cmd
}

// buildCallerTree 是一个递归函数，用于构建单个文件的调用树。
func buildCallerTree(filePath string, callerGraph map[string][]string, visited map[string]bool) CallerNode {
	if visited[filePath] {
		return CallerNode{FilePath: filePath + " (循环依赖)", Callers: []CallerNode{}}
	}
	visited[filePath] = true

	node := CallerNode{
		FilePath: filePath,
		Callers:  []CallerNode{},
	}

	if callers, exists := callerGraph[filePath]; exists {
		for _, callerPath := range callers {
			node.Callers = append(node.Callers, buildCallerTree(callerPath, callerGraph, visited))
		}
	}

	delete(visited, filePath)

	return node
}

// collectAffectedFiles 递归遍历调用树以收集一组唯一的文件路径。
func collectAffectedFiles(node *CallerNode, visited map[string]struct{}) {
	if node == nil {
		return
	}
	if !strings.HasSuffix(node.FilePath, " (循环依赖)") {
		visited[node.FilePath] = struct{}{}
	}
	for i := range node.Callers {
		collectAffectedFiles(&node.Callers[i], visited)
	}
}

// setToSortedSlice 将一个 map[string]struct{} (集合) 转换为一个排序后的字符串切片。
func setToSortedSlice(set map[string]struct{}) []string {
	slice := make([]string, 0, len(set))
	for item := range set {
		slice = append(slice, item)
	}
	sort.Strings(slice)
	return slice
}
