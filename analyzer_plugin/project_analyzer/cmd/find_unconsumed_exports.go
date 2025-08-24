package cmd

// go run main.go find-unconsumed-exports -i /Users/bird/company/sc1.0/live/shopline-live-sale -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin/project_analyzer/result -x "node_modules/**" -x "bffApiDoc/**" -x "**/apidoc/**"

import (
	"fmt"
	"main/analyzer_plugin/project_analyzer/internal/filenamer"
	"main/analyzer_plugin/project_analyzer/internal/writer"
	"main/analyzer_plugin/project_analyzer/unconsumed"

	"github.com/spf13/cobra"
)

var (
	findUnconsumedInputDir   string
	findUnconsumedOutputDir  string
	findUnconsumedExclude    []string
	findUnconsumedIsMonorepo bool
)

// NewFindUnconsumedExportsCmd 创建并返回 `find-unconsumed-exports` 命令
func NewFindUnconsumedExportsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "find-unconsumed-exports",
		Short: "查找项目中已导出但未被消费的变量、函数或类型。",
		Long:  `该命令会全面分析一个项目，找出所有被导出的实体，并检查它们是否在项目的其他地方被导入。这对于清理废弃的、不再需要的导出非常有用。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 步骤 1: 将命令行参数打包到参数结构体中。
			params := unconsumed.Params{
				RootPath:   findUnconsumedInputDir,
				Exclude:    findUnconsumedExclude,
				IsMonorepo: findUnconsumedIsMonorepo,
			}

			// 步骤 2: 调用核心业务逻辑。
			result, err := unconsumed.Find(params)
			if err != nil {
				return err
			}

			// 步骤 3: 使用新的 writer 和 filenamer 包来格式化并输出结果。
			if findUnconsumedOutputDir != "" {
				// 如果指定了输出目录，则将结果写入文件。
				outputFileName := filenamer.GenerateOutputFileName(findUnconsumedInputDir, "find_unconsumed_exports")
				err = writer.WriteJSONResult(findUnconsumedOutputDir, outputFileName, result)
				if err != nil {
					// 包装错误信息，提供更清晰的上下文。
					return fmt.Errorf("无法将输出写入文件: %w", err)
				}
			}
			return nil
		},
	}

	// 定义所有命令行标志。
	cmd.Flags().StringVarP(&findUnconsumedInputDir, "input", "i", "", "要分析的 TypeScript 项目目录的路径 (必需)")
	cmd.Flags().StringVarP(&findUnconsumedOutputDir, "output", "o", "", "用于存储 JSON 结果的输出目录路径 (可选, 默认为标准输出)")
	cmd.Flags().StringSliceVarP(&findUnconsumedExclude, "exclude", "x", []string{}, "要从分析中排除的 Glob 模式 (可多次使用)")
	cmd.Flags().BoolVarP(&findUnconsumedIsMonorepo, "monorepo", "m", false, "如果分析的是 monorepo 项目，请设置为 true")

	cmd.MarkFlagRequired("input")

	return cmd
}
