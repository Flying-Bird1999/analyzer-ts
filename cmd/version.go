package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version 先搞个假的
var version = "0.0.1"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "打印 analyzer-ts 的版本号",
	Long:  `该命令用于打印当前 analyzer-ts 工具的版本信息。`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("analyzer-ts version %s\n", version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
