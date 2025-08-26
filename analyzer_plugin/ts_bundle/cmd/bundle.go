package cmd

// 示例:
// go run main.go bundle -i /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin/ts_bundle/test_data/complex_exports/index.ts -t Container -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin/ts_bundle/result.ts

import (
	"fmt"
	"os"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/ts_bundle"

	"github.com/spf13/cobra"
)

// 定义命令行参数的变量
var inputFile string
var inputType string
var outputFile string
var projectRoot string

// NewBundleCmd 创建并返回一个 Cobra 命令，用于处理 TypeScript 类型打包。
// 这个命令是主 CLI 工具 (`analyzer-ts`) 的一个子命令。
func NewBundleCmd() *cobra.Command {
	bundleCmd := &cobra.Command{
		Use:   "bundle",
		Short: "打包 TypeScript 类型声明",
		Long:  `从给定的入口文件递归收集所有引用的类型声明，并将它们打包到一个单独的文件中。`,
		Run: func(cmd *cobra.Command, args []string) {
			// 检查必要的参数是否已提供
			if inputFile == "" || inputType == "" {
				fmt.Println("用法: analyzer-ts bundle --input <入口文件> --type <类型名称>")
				cmd.Help()
				os.Exit(1)
			}
			// 调用 ts_bundle 包中的 GenerateBundle 函数执行打包逻辑
			bundledContent := ts_bundle.GenerateBundle(inputFile, inputType, outputFile, projectRoot)

			// 将打包后的内容写入指定输出文件
			err := os.WriteFile(outputFile, []byte(bundledContent), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "写入文件失败: %v\n", err)
				return
			}

			fmt.Printf("打包完成，输出文件: %s\n", outputFile)
		},
	}

	// 定义命令行标志 (flags)
	bundleCmd.Flags().StringVarP(&inputFile, "input", "i", "", "入口文件路径 (必需)")
	bundleCmd.Flags().StringVarP(&inputType, "type", "t", "", "要分析的类型名称 (必需)")
	bundleCmd.Flags().StringVarP(&outputFile, "output", "o", "./output.ts", "输出文件路径")
	bundleCmd.Flags().StringVarP(&projectRoot, "root", "r", "", "项目根路径")

	return bundleCmd
}
