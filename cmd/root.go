package cmd

import (
	"fmt"
	"os"

	projectAnalyzerCmd "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/cmd"
	tsBundleCmd "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/ts_bundle/cmd"
	"github.com/Flying-Bird1999/analyzer-ts/pkg/gitlab"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "analyzer-ts",
	Short: "一个用于分析 TypeScript 项目的命令行工具。",
	Long:  `analyzer-ts 是一个功能强大的命令行工具，旨在解析和分析 TypeScript 代码库。它可以生成报告、打包代码，并将分析结果存储在数据库中。`,
}

func init() {
	// 添加 project_analyzer 的子命令
	RootCmd.AddCommand(projectAnalyzerCmd.GetAnalyzeCmd())
	RootCmd.AddCommand(projectAnalyzerCmd.GetQueryCmd()) // 新增: 添加 query 命令

	RootCmd.AddCommand(projectAnalyzerCmd.NewStoreDbCmd())

	// 添加 ts_bundle 的子命令
	RootCmd.AddCommand(tsBundleCmd.NewBundleCmd())
	RootCmd.AddCommand(tsBundleCmd.NewBatchBundleCmd())

	// 添加 gitlab 子命令
	RootCmd.AddCommand(gitlab.GetCommand())

	// 添加 find 命令
	RootCmd.AddCommand(findCmd)

	// 添加其他顶级命令
	RootCmd.AddCommand(ScanCmd)
}

// Execute 将所有子命令添加到根命令并适当设置标志。
// 这是由 main.main() 调用的。它只需要对 rootCmd 执行一次。
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "哎呀，执行您的命令行时出错 '%s'", err)
		os.Exit(1)
	}
}
