package cmd

// go run main.go count-any -i /Users/bird/company/sc1.0/live/shopline-live-sale -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin/project_analyzer/result -x "node_modules/**" -x "bffApiDoc/**"

import (
	"fmt"
	"log"
	countany "main/analyzer_plugin/project_analyzer/countAny"
	"os"

	"github.com/spf13/cobra"
)

var (
	countAnyInputDir   string
	countAnyOutputDir  string
	countAnyIsMonorepo bool
	countAnyExclude    []string
)

// NewCountAnyCmd 创建并返回 count-any 命令。
func NewCountAnyCmd() *cobra.Command {
	countAnyCmd := &cobra.Command{
		Use:   "count-any",
		Short: "统计 TypeScript 项目中 any 类型的使用数量。",
		Long:  `统计一个 TypeScript 项目中所有显式 any 类型的使用数量，并可选择将详细结果输出为 JSON 文件。`,
		Run: func(cmd *cobra.Command, args []string) {
			if countAnyInputDir == "" {
				fmt.Println("统计 any 需要输入路径。")
				cmd.Help()
				os.Exit(1)
			}

			// 调用核心业务逻辑
			result := countany.CountAnyUsages(countAnyInputDir, countAnyExclude, countAnyIsMonorepo)

			// 根据是否提供了 output 参数，决定输出到文件还是控制台
			if countAnyOutputDir != "" {
				// 调用写入函数，传入输出目录和输入目录
				err := countany.WriteOutput(countAnyOutputDir, countAnyInputDir, result)
				if err != nil {
					log.Fatalf("写入输出文件时出错: %v", err)
				}
			} else {
				// 如果没有提供 output，则在控制台打印总数
				fmt.Printf("项目中 any 类型的使用总数: %d\n", result.TotalAnyCount)
			}
		},
	}

	// 设置命令的标志
	countAnyCmd.Flags().StringVarP(&countAnyInputDir, "input", "i", "", "要分析的 TypeScript 项目目录的路径")
	countAnyCmd.Flags().StringVarP(&countAnyOutputDir, "output", "o", "", "用于存储 JSON 输出文件的目录路径")
	countAnyCmd.Flags().StringSliceVarP(&countAnyExclude, "exclude", "x", []string{}, "要从分析中排除的 Glob 模式")
	countAnyCmd.Flags().BoolVarP(&countAnyIsMonorepo, "monorepo", "m", false, "如果要分析的是 monorepo，则设置为 true")

	// 将 input 标志设为必需
	countAnyCmd.MarkFlagRequired("input")

	return countAnyCmd
}
