package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "analyzer-ts",
	Short: "一个用于分析 TypeScript 项目的命令行工具。",
	Long:  `analyzer-ts 是一个功能强大的命令行工具，旨在解析和分析 TypeScript 代码库。它可以生成报告、打包代码，并将分析结果存储在数据库中。`,
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
