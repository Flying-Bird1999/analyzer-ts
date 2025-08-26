// package cmd 存放所有命令行相关的代码。
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/scanProject"
	"github.com/spf13/cobra"
)

// 定义一组变量，用于存储从命令行标志（flags）中捕获的用户输入。
var (
	// inputDir 存储用户通过 --input 或 -i 标志指定的项目根目录路径。
	inputDir string
	// excludePaths 存储用户通过 --exclude 或 -x 标志指定的需要排除的文件或目录的 glob 模式列表。
	excludePaths []string
	// isMonorepo 标记项目是否为一个 monorepo，通过 --monorepo 或 -m 标志设置。
	isMonorepo bool
	// outputDir 存储用户通过 --output 或 -o 标志指定的输出目录。如果为空，则结果将打印到标准输出。
	outputDir string
)

// ScanCmd 定义了 `scan` 命令的所有行为和属性。
var ScanCmd = &cobra.Command{
	// Use 是命令的名称，用户在终端中输入 `go run main.go scan` 来调用它。
	Use: "scan",
	// Short 是命令的简短描述，会显示在帮助列表（-h）中。
	Short: "扫描项目仓库以列出所有文件",
	// Long 是命令的详细描述，当用户运行 `go run main.go help scan` 时显示。
	Long: `扫描给定的项目仓库，列出所有文件，并根据排除模式进行过滤。支持 JSON 格式输出。`,
	// Run 是 `scan` 命令的核心逻辑。当命令被调用时，这个函数会被执行。
	Run: func(cmd *cobra.Command, args []string) {
		// 1. 基于用户输入的参数，初始化项目扫描器。
		pr := scanProject.NewProjectResult(inputDir, excludePaths, isMonorepo)
		// 执行文件列表的扫描。
		pr.ScanFileList()

		// 2. 检查用户是否指定了输出目录。
		if outputDir != "" {
			// 2.1. 如果指定了输出目录，则将结果以 JSON 格式写入文件。

			// 确保输出目录存在，如果不存在则创建它。
			if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
				fmt.Printf("错误：创建输出目录失败: %s\n", err)
				os.Exit(1)
			}

			// 获取扫描结果。
			result := pr.GetFileList()
			// 将结果数据序列化为易于阅读的 JSON 格式（带有缩进）。
			jsonData, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				fmt.Printf("错误：序列化JSON失败: %s\n", err)
				os.Exit(1)
			}

			// 构建最终的输出文件路径，文件名为 scan_result.json。
			outputPath := filepath.Join(outputDir, "scan_result.json")
			// 将 JSON 数据写入文件。
			if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
				fmt.Printf("错误：写入文件失败: %s\n", err)
				os.Exit(1)
			}
			// 提示用户操作成功。
			fmt.Printf("扫描结果已成功写入: %s\n", outputPath)
		} else {
			// 2.2. 如果未指定输出目录，则将结果逐行打印到标准输出（控制台）。
			for path, item := range pr.GetFileList() {
				fmt.Printf("file: %s, fileName: %s, size: %s, ext: %s\n", path, item.FileName, item.Size, item.Ext)
			}
		}
	},
}

// init 函数在程序启动时由 Go 运行时自动调用。
// 它用于初始化命令、定义标志以及将命令添加到根命令中。
func init() {
	// 为 scan 命令定义命令行标志。
	ScanCmd.Flags().StringVarP(&inputDir, "input", "i", "", "要扫描的仓库路径")
	ScanCmd.Flags().StringSliceVarP(&excludePaths, "exclude", "x", []string{}, "要排除的 glob 模式")
	ScanCmd.Flags().BoolVarP(&isMonorepo, "monorepo", "m", false, "是否为 monorepo 项目？")
	ScanCmd.Flags().StringVarP(&outputDir, "output", "o", "", "用于存放 scan_result.json 的输出目录（可选，默认为标准输出）")

	// 将 --input 标志设置为必需项，如果用户未提供此标志，Cobra 将会报错。
	ScanCmd.MarkFlagRequired("input")
}
