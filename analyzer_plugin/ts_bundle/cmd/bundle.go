package cmd

// 示例:
// go run main.go bundle -i /Users/bird/Desktop/alalyzer/analyzer-ts/ts_example/bundle/index1.ts -t Class -o /Users/bird/Desktop/alalyzer/analyzer-ts/ts_example/output/result.ts

import (
	"fmt"
	"os"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/ts_bundle"

	"github.com/spf13/cobra"
)

var inputFile string
var inputType string
var outputFile string
var projectRoot string

func NewBundleCmd() *cobra.Command {
	bundleCmd := &cobra.Command{
		Use:   "bundle",
		Short: "打包 TypeScript 类型声明",
		Long:  `从给定的入口文件递归收集所有引用的类型声明，并将它们打包到一个单独的文件中。`,
		Run: func(cmd *cobra.Command, args []string) {
			if inputFile == "" || inputType == "" {
				fmt.Println("用法: analyzer-ts bundle --input <入口文件> --type <类型名称>")
				cmd.Help()
				os.Exit(1)
			}
			ts_bundle.GenerateBundle(inputFile, inputType, outputFile, projectRoot)
		},
	}

	bundleCmd.Flags().StringVarP(&inputFile, "input", "i", "", "入口文件路径 (必需)")
	bundleCmd.Flags().StringVarP(&inputType, "type", "t", "", "要分析的类型名称 (必需)")
	bundleCmd.Flags().StringVarP(&outputFile, "output", "o", "./output.ts", "输出文件路径")
	bundleCmd.Flags().StringVarP(&projectRoot, "root", "r", "", "项目根路径")

	return bundleCmd
}
