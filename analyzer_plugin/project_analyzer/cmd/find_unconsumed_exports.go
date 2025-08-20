package cmd

// go run main.go find-unconsumed-exports -i /Users/bird/company/sc1.0/live/shopline-live-sale -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin/project_analyzer/result -x "node_modules/**" -x "bffApiDoc/**" -x "**/apidoc/**"

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"main/analyzer_plugin/project_analyzer/unconsumed"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	findUnconsumedInputDir   string
	findUnconsumedOutputDir  string
	findUnconsumedExclude    []string
	findUnconsumedIsMonorepo bool
)

// NewFindUnconsumedExportsCmd 创建并返回 `find-unconsumed-exports` 命令
func NewFindUnconsumedExportsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "find-unconsumed-exports",
		Short: "查找项目中已导出但未被消费的变量、函数或类型。",
		Long:  `该命令会全面分析一个项目，找出所有被导出的实体，并检查它们是否在项目的其他地方被导入。这对于清理废弃的、不再需要的导出非常有用。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 步骤 1: 将命令行参数打包到参数结构体中。
			params := unconsumed.Params{
				RootPath:   findUnconsumedInputDir,
				Exclude:    findUnconsumedExclude,
				IsMonorepo: findUnconsumedIsMonorepo,
			}

			// 步骤 2: 调用核心业务逻辑。
			result, err := unconsumed.Find(params)
			if err != nil {
				return err
			}

			// 步骤 3: 格式化并输出结果。
			if findUnconsumedOutputDir != "" {
				// 输出为 JSON 文件。
				return printUnconsumedExportsToFile(result, findUnconsumedInputDir, findUnconsumedOutputDir)
			} else {
				// 在控制台打印摘要。
				printUnconsumedExportsToConsole(result)
				return nil
			}
		},
	}

	// 定义所有命令行标志。
	cmd.Flags().StringVarP(&findUnconsumedInputDir, "input", "i", "", "要分析的 TypeScript 项目目录的路径 (必需)")
	cmd.Flags().StringVarP(&findUnconsumedOutputDir, "output", "o", "", "用于存储 JSON 结果的输出目录路径 (可选, 默认为标准输出)")
	cmd.Flags().StringSliceVarP(&findUnconsumedExclude, "exclude", "x", []string{}, "要从分析中排除的 Glob 模式 (可多次使用)")
	cmd.Flags().BoolVarP(&findUnconsumedIsMonorepo, "monorepo", "m", false, "如果分析的是 monorepo 项目，请设置为 true")

	cmd.MarkFlagRequired("input")

	return cmd
}

// printUnconsumedExportsToFile 将分析结果序列化为JSON并写入文件。
func printUnconsumedExportsToFile(result *unconsumed.Result, inputDir, outputDir string) error {
	outputBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("无法将结果序列化为 JSON: %w", err)
	}

	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("无法创建输出目录 %s: %w", outputDir, err)
	}

	baseName := filepath.Base(inputDir)
	outputFileName := fmt.Sprintf("%s_find_unconsumed_exports.json", baseName)
	fullOutputPath := filepath.Join(outputDir, outputFileName)

	if err := ioutil.WriteFile(fullOutputPath, outputBytes, 0644); err != nil {
		return fmt.Errorf("无法将输出写入文件 %s: %w", fullOutputPath, err)
	}

	fmt.Printf("未消费导出分析结果已写入: %s", fullOutputPath)
	return nil
}

// printUnconsumedExportsToConsole 在控制台打印易于阅读的分析结果摘要。
func printUnconsumedExportsToConsole(result *unconsumed.Result) {
	fmt.Printf("--- 未消费的导出项分析摘要 ---\n")
	fmt.Printf("扫描文件总数: %d\n", result.Summary.TotalFilesScanned)
	fmt.Printf("发现导出项总数: %d\n", result.Summary.TotalExportsFound)
	fmt.Printf("发现未消费导出项: %d\n", result.Summary.UnconsumedExportsFound)
	fmt.Println("----------------------------------")

	if len(result.UnconsumedExports) > 0 {
		fmt.Println("\n未消费的导出项列表:")
		for _, export := range result.UnconsumedExports {
			fmt.Printf("  - %s:%d - %s (类型: %s)\n", export.FilePath, export.Line, export.ExportName, export.Kind)
		}
	}
}
