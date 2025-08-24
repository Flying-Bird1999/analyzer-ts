package cmd

import (
	"fmt"
	"os"

	projectAnalyzerCmd "main/analyzer_plugin/project_analyzer/cmd"
	tsBundleCmd "main/analyzer_plugin/ts_bundle/cmd"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "analyzer-ts",
	Short: "一个用于分析 TypeScript 项目的命令行工具。",
	Long:  `analyzer-ts 是一个功能强大的命令行工具，旨在解析和分析 TypeScript 代码库。它可以生成报告、打包代码，并将分析结果存储在数据库中。`,
}

func init() {
	// 注册重构后的统一分析命令
	RootCmd.AddCommand(projectAnalyzerCmd.GetAnalyzeCmd())

	// 保留其他尚未重构的命令
	RootCmd.AddCommand(projectAnalyzerCmd.NewStoreDbCmd())
	RootCmd.AddCommand(tsBundleCmd.NewBundleCmd())
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
