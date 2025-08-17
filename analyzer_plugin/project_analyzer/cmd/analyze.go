package cmd

// example:
// go run main.go analyze -i /Users/bird/Desktop/components/shopline-admin-components -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin/project_analyzer/result/analyzer_result_json -x "examples/**" -x "tests/**"

// go run main.go analyze -i /Users/bird/company/sc1.0/live/shopline-live-sale -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin/project_analyzer/result/analyzer_result_json -x "node_modules/**" -x "bffApiDoc/**"

// go run main.go analyze -i /Users/bird/company/sc1.0/components/nova -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin/project_analyzer/result/analyzer_result_json -m true -x "**/e2e/**"

// go run main.go store-db -i /Users/bird/company/sc1.0/live/shopline-live-sale -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin/project_analyzer/result/analyzer_result_json -x "node_modules/**" -x "bffApiDoc/**"

// go run main.go analyze -i /Users/bird/company/sc1.0/live/shopline-post-center -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin/project_analyzer/result/analyzer_result_json -x "node_modules/**" -x "bffApiDoc/**"

// go run main.go analyze -i /Users/bird/company/sc1.0/mc/message-center/client -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin/project_analyzer/result/analyzer_result_json -x "node_modules/**" -x "bffApiDoc/**" -x "sc-components/**"

// go run main.go analyze -i /Users/bird/company/sc1.0/components/sc-components -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin/project_analyzer/result/analyzer_result_json

// go run main.go analyze -i /Users/bird/Desktop/sp/smart-push-new -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin/project_analyzer/result/analyzer_result_json -m true

// go run main.go analyze -i /Users/bird/Desktop/sp/fe-lib -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin/project_analyzer/result/analyzer_result_json -m true

import (
	"fmt"
	"main/analyzer_plugin/project_analyzer"
	"main/cmd"
	"os"

	"github.com/spf13/cobra"
)

var (
	analyzeInputDir   string
	analyzeOutputDir  string
	analyzeIsMonorepo bool
	analyzeExclude    []string
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "分析 TypeScript 项目并将结果输出为 JSON 文件。",
	Long:  `分析一个 TypeScript 项目，解析所有相关文件以构建每个文件的 AST。然后将结构化数据作为单独的 JSON 文件输出到一个目录中。`,
	Run: func(cmd *cobra.Command, args []string) {
		if analyzeInputDir == "" || analyzeOutputDir == "" {
			fmt.Println("分析需要输入和输出路径。")
			cmd.Help()
			os.Exit(1)
		}
		// 注意：原始代码传递了更多参数，如别名和扩展名。
		// 为简单起见，此版本中未将它们公开为标志，但如果需要可以加回来。
		project_analyzer.AnalyzeProject(analyzeInputDir, analyzeOutputDir, analyzeExclude, analyzeIsMonorepo)
	},
}

func init() {
	cmd.RootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVarP(&analyzeInputDir, "input", "i", "", "要分析的 TypeScript 项目目录的路径")
	analyzeCmd.Flags().StringVarP(&analyzeOutputDir, "output", "o", "", "用于存储 JSON 输出文件的目录路径")
	analyzeCmd.Flags().StringSliceVarP(&analyzeExclude, "exclude", "x", []string{}, "要从分析中排除的 Glob 模式")
	analyzeCmd.Flags().BoolVarP(&analyzeIsMonorepo, "monorepo", "m", false, "如果要分析的是 monorepo，则设置为 true")

	analyzeCmd.MarkFlagRequired("input")
	analyzeCmd.MarkFlagRequired("output")
}
