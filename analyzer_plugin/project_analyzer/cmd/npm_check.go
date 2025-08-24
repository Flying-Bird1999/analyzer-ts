package cmd

// go run main.go npm-check -i /Users/bird/company/sc1.0/live/shopline-live-sale -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin/project_analyzer/result -x "node_modules/**" -x "bffApiDoc/**"

import (
	"fmt"
	"main/analyzer_plugin/project_analyzer/dependency"
	"main/analyzer_plugin/project_analyzer/internal/filenamer"
	"main/analyzer_plugin/project_analyzer/internal/writer"
	"os"

	"github.com/spf13/cobra"
)

var (
	npmCheckInputDir   string
	npmCheckOutputDir  string
	npmCheckExclude    []string
	npmCheckIsMonorepo bool
)

func NewNpmCheckCmd() *cobra.Command {
	npmCheckCmd := &cobra.Command{
		Use:   "npm-check",
		Short: "检查项目的NPM依赖，类似npm-check。",
		Long:  `分析一个TypeScript项目，识别隐式依赖（幽灵依赖）、未使用的依赖和过期的依赖。`,
		// Run 是命令的执行入口，它现在只负责参数的传递和结果的输出。
		Run: func(cmd *cobra.Command, args []string) {
			if npmCheckInputDir == "" {
				fmt.Println("需要输入路径。")
				cmd.Help()
				os.Exit(1)
			}

			// 直接调用 dependency 包中的核心业务逻辑函数。
			depCheckResult := dependency.Check(npmCheckInputDir, npmCheckExclude, npmCheckIsMonorepo)

			// 根据用户指定的输出方式，格式化并输出结果。
			if npmCheckOutputDir != "" {
				// 使用新的 writer 和 filenamer 包将结果写入文件。
				outputFileName := filenamer.GenerateOutputFileName(npmCheckInputDir, "npm_check")
				err := writer.WriteJSONResult(npmCheckOutputDir, outputFileName, depCheckResult)
				if err != nil {
					fmt.Printf("Error writing JSON to file: %s", err)
					return
				}
			} else {
				//
			}
		},
	}

	npmCheckCmd.Flags().StringVarP(&npmCheckInputDir, "input", "i", "", "要分析的 TypeScript 项目目录的路径")
	npmCheckCmd.Flags().StringVarP(&npmCheckOutputDir, "output", "o", "", "用于存储 JSON 输出文件的目录路径")
	npmCheckCmd.Flags().StringSliceVarP(&npmCheckExclude, "exclude", "x", []string{}, "要从分析中排除的 Glob 模式")
	npmCheckCmd.Flags().BoolVarP(&npmCheckIsMonorepo, "monorepo", "m", false, "如果要分析的是 monorepo，则设置为 true")

	npmCheckCmd.MarkFlagRequired("input")

	return npmCheckCmd
}
