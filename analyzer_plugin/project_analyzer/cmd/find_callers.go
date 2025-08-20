package cmd

// example： go run main.go find-callers -i /Users/bird/company/sc1.0/live/shopline-live-sale -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin/project_analyzer/result -x "examples/**" -x "tests/**" -f /Users/bird/company/sc1.0/live/shopline-live-sale/src/utils/downloadFile.ts -f /Users/bird/company/sc1.0/live/shopline-live-sale/src/utils/string-utils.ts

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"main/analyzer_plugin/project_analyzer/callgraph"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// 定义该命令所需的标志变量
var (
	findCallersInputDir    string
	findCallersTargetFiles []string
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
		// RunE 是命令的执行入口，它现在只负责参数的传递和结果的输出。
		RunE: func(cmd *cobra.Command, args []string) error {
			// 步骤 1: 将从命令行解析出的参数，打包成一个结构体，传递给业务逻辑层。
			params := callgraph.Params{
				RootPath:    findCallersInputDir,
				TargetFiles: findCallersTargetFiles,
				Exclude:     findCallersExclude,
				IsMonorepo:  findCallersIsMonorepo,
			}

			// 步骤 2: 调用 callgraph 包中的核心业务逻辑函数。
			result, err := callgraph.Find(params)
			if err != nil {
				return err
			}

			// 步骤 3: 将业务逻辑层返回的结构化结果，序列化为 JSON 并输出到文件或控制台。
			outputBytes, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return fmt.Errorf("无法将结果序列化为 JSON: %w", err)
			}

			if findCallersOutputDir != "" {
				if err := os.MkdirAll(findCallersOutputDir, os.ModePerm); err != nil {
					return fmt.Errorf("无法创建输出目录 %s: %w", findCallersOutputDir, err)
				}
				baseName := filepath.Base(findCallersInputDir)
				outputFileName := fmt.Sprintf("%s_find_callers.json", baseName)
				fullOutputPath := filepath.Join(findCallersOutputDir, outputFileName)

				if err := ioutil.WriteFile(fullOutputPath, outputBytes, 0644); err != nil {
					return fmt.Errorf("无法将输出写入文件 %s: %w", fullOutputPath, err)
				}
				fmt.Printf("调用链分析结果已写入: %s", fullOutputPath)
			} else {
				fmt.Println(string(outputBytes))
			}

			return nil
		},
	}

	// 定义所有命令行标志
	cmd.Flags().StringVarP(&findCallersInputDir, "input", "i", "", "要分析的 TypeScript 项目目录的路径 (必需)")
	cmd.Flags().StringSliceVarP(&findCallersTargetFiles, "file", "f", []string{}, "要追踪其调用链的文件路径 (必需, 可多次使用)")
	cmd.Flags().StringVarP(&findCallersOutputDir, "output", "o", "", "用于存储 JSON 结果的输出目录路径 (可选, 默认为标准输出)")
	cmd.Flags().StringSliceVarP(&findCallersExclude, "exclude", "x", []string{}, "要从分析中排除的 Glob 模式 (可多次使用)")
	cmd.Flags().BoolVarP(&findCallersIsMonorepo, "monorepo", "m", false, "如果分析的是 monorepo 项目，请设置为 true")

	cmd.MarkFlagRequired("input")
	cmd.MarkFlagRequired("file")

	return cmd
}
