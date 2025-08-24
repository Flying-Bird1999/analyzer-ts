// package cmd 存放了所有暴露给用户的命令行接口 (CLI) 的定义。
// 这个包使用了流行的 `cobra` 库来构建子命令。
//
// 它的核心职责是：
// 1. 定义命令的名称、帮助信息和命令行标志 (flags)。
// 2. 解析用户从命令行输入的参数。
// 3. 调用相应业务逻辑包中的核心函数。
// 4. 接收业务逻辑的返回结果，并将其格式化后输出到控制台或文件。
//
// 按照设计原则，这个包应该保持为一个“瘦层”，不包含任何复杂的业务逻辑。
package cmd

// example: go run main.go analyze -i /Users/bird/company/sc1.0/live/shopline-live-sale -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin/project_analyzer/result -x "node_modules/**" -x "bffApiDoc/**"

import (
	"fmt"
	project_analyzer "main/analyzer_plugin/project_analyzer/projectAnalyzer"
	"os"

	"github.com/spf13/cobra"
)

var (
	analyzeInputDir   string
	analyzeOutputDir  string
	analyzeIsMonorepo bool
	analyzeExclude    []string
)

func NewAnalyzeCmd() *cobra.Command {
	analyzeCmd := &cobra.Command{
		Use:   "analyze",
		Short: "分析 TypeScript 项目并将结果输出为 JSON 文件。",
		Long:  `分析一个 TypeScript 项目，解析所有相关文件以构建每个文件的 AST。然后将结构化数据作为单独的 JSON 文件输出到一个目录中。`,
		Run: func(cmd *cobra.Command, args []string) {
			if analyzeInputDir == "" || analyzeOutputDir == "" {
				fmt.Println("分析需要输入和输出路径。")
				cmd.Help()
				os.Exit(1)
			}
			// 直接调用业务逻辑层的 AnalyzeProject 函数来执行核心任务。
			project_analyzer.AnalyzeProject(analyzeInputDir, analyzeOutputDir, analyzeExclude, analyzeIsMonorepo)
		},
	}

	analyzeCmd.Flags().StringVarP(&analyzeInputDir, "input", "i", "", "要分析的 TypeScript 项目目录的路径")
	analyzeCmd.Flags().StringVarP(&analyzeOutputDir, "output", "o", "", "用于存储 JSON 输出文件的目录路径")
	analyzeCmd.Flags().StringSliceVarP(&analyzeExclude, "exclude", "x", []string{}, "要从分析中排除的 Glob 模式")
	analyzeCmd.Flags().BoolVarP(&analyzeIsMonorepo, "monorepo", "m", false, "如果要分析的是 monorepo，则设置为 true")

	analyzeCmd.MarkFlagRequired("input")
	analyzeCmd.MarkFlagRequired("output")

	return analyzeCmd
}
