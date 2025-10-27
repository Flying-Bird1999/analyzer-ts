package cmd

// go run main.go find references -i /Users/zxc/Desktop/analyzer/analyzer-ts/analyzer_plugin/ts_bundle/testdata/simple-case /Users/zxc/Desktop/analyzer/analyzer-ts/analyzer_plugin/ts_bundle/testdata/simple-case/simple_test.ts:2:1

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/lsp"
	"github.com/spf13/cobra"
)

var (
	queryInputPath string
)

// lspCmd represents the base command for all lsp-related actions
var findCmd = &cobra.Command{
	Use:   "find",
	Short: "Perform language-aware queries on the TypeScript project",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if queryInputPath == "" {
			return fmt.Errorf("required flag \"input\" not set")
		}
		return nil
	},
}

// referencesCmd represents the command to find references
var referencesCmd = &cobra.Command{
	Use:   "references [file:line:char]",
	Short: "Find all references to a symbol at a given location",
	Args:  cobra.ExactArgs(1), // 要求必须有一个参数
	Run: func(cmd *cobra.Command, args []string) {
		// 1. 解析参数
		location := args[0]
		parts := strings.Split(location, ":")
		if len(parts) != 3 {
			log.Fatalf("invalid location format. expected [file:line:char], got %s", location)
		}

		filePath, err := filepath.Abs(parts[0])
		if err != nil {
			log.Fatalf("invalid file path: %v", err)
		}

		line, err := strconv.Atoi(parts[1])
		if err != nil {
			log.Fatalf("invalid line number: %v", err)
		}
		char, err := strconv.Atoi(parts[2])
		if err != nil {
			log.Fatalf("invalid character number: %v", err)
		}

		// 2. 创建新的 LSP 服务
		lspService, err := lsp.NewService(queryInputPath)
		if err != nil {
			log.Fatalf("failed to create lsp service: %v", err)
		}
		defer lspService.Close()

		// 3. 执行查找引用
		response, err := lspService.FindReferences(context.Background(), filePath, line, char)
		if err != nil {
			log.Fatalf("error finding references: %v", err)
		}

		// 4. 打印结果
		if response.Locations == nil || len(*response.Locations) == 0 {
			fmt.Println("No references found.")
			return
		}

		fmt.Printf("Found %d references:\n", len(*response.Locations))
		for _, loc := range *response.Locations {
			// LSP 的行列号是从 0 开始的，我们转换为从 1 开始，更符合常规
			fmt.Printf("  - %s:%d:%d\n", loc.Uri, loc.Range.Start.Line+1, loc.Range.Start.Character+1)
		}
	},
}

// init 函数在程序启动时被调用，用于注册命令
func init() {
	findCmd.PersistentFlags().StringVarP(&queryInputPath, "input", "i", "", "要分析的项目根目录")
	// 将 referencesCmd 作为 findCmd 的子命令
	findCmd.AddCommand(referencesCmd)
	// 将 queryCmd 注册到根命令 (此行将在 root.go 中处理)
	// rootCmd.AddCommand(queryCmd)
}
