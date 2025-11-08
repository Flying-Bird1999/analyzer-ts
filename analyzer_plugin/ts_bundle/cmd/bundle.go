package cmd

// TypeScript 类型声明打包工具
//
// 1. 单类型模式: bundle 命令 - 从单个入口文件收集指定类型及其所有依赖
//    analyzer-ts bundle -i /path/to/file.ts -t TypeName -o output.ts
//
// 2. 批量类型模式: batch-bundle 命令 - 批量处理多个类型，每个类型生成独立文件
//    analyzer-ts batch-bundle -e "/path/to/file1.ts:Type1" -e "/path/to/file2.ts:Type2" --output-dir ./output/
//
// 批量模式支持别名: "文件路径:类型名[:别名]" 格式

import (
	"fmt"
	"os"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/ts_bundle"

	"github.com/spf13/cobra"
)

// 定义命令行参数的变量 - 单类型模式
var inputFile string
var inputType string
var outputFile string
var projectRoot string

// 定义命令行参数的变量 - 批量类型模式
var entries []string // 批量入口点
var outputDir string // 批量模式输出目录

// NewBundleCmd 创建单类型打包命令
func NewBundleCmd() *cobra.Command {
	bundleCmd := &cobra.Command{
		Use:   "bundle",
		Short: "打包单个 TypeScript 类型声明",
		Long: `从给定的入口文件递归收集指定类型及其所有依赖的声明，并将它们打包到单个文件中。

这是单类型打包模式，适合处理一个入口类型及其依赖。如需处理多个类型，请使用 batch-bundle 命令。

示例:
  # 基础用法
  analyzer-ts bundle -i ./src/index.ts -t MyType -o ./dist/types.d.ts

  # 使用项目根路径
  analyzer-ts bundle -i ./src/index.ts -t MyType -o ./dist/types.d.ts -r ./my-project

功能特性:
  • 智能依赖收集: 自动递归收集类型的所有依赖
  • 路径别名解析: 完整支持 tsconfig.json 中的 paths 配置
  • 多种导入方式: 支持命名导入、默认导入、命名空间导入等`,
		Run: func(cmd *cobra.Command, args []string) {
			// 验证必需参数
			if inputFile == "" || inputType == "" {
				fmt.Fprintf(os.Stderr, "错误: 需要指定 --input 和 --type 参数\n")
				fmt.Println("用法: analyzer-ts bundle -i <入口文件> -t <类型名称> -o <输出文件>")
				cmd.Help()
				os.Exit(1)
			}

			fmt.Printf("打包类型: %s:%s\n", inputFile, inputType)
			bundledContent, err := ts_bundle.GenerateBundle(inputFile, inputType, projectRoot)
			if err != nil {
				fmt.Fprintf(os.Stderr, "打包失败: %v\n", err)
				os.Exit(1)
			}

			// 将打包后的内容写入指定输出文件
			err = os.WriteFile(outputFile, []byte(bundledContent), 0644)
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
	bundleCmd.Flags().StringVarP(&projectRoot, "root", "r", "", "项目根路径 (可选，自动检测)")

	return bundleCmd
}

// NewBatchBundleCmd 创建批量类型打包命令
func NewBatchBundleCmd() *cobra.Command {
	bundleCmd := &cobra.Command{
		Use:   "batch-bundle",
		Short: "批量打包多个 TypeScript 类型声明",
		Long: `批量处理多个类型，每个类型生成独立的 .d.ts 文件，完美解决命名冲突。

这是批量类型打包模式，适合处理多个入口类型。如需处理单个类型，请使用 bundle 命令。

示例:
  # 基础批量打包 - 每个类型生成独立文件
  analyzer-ts batch-bundle -e "./src/user.ts:User" -e "./src/product.ts:Product" --output-dir ./dist/types/

  # 支持逗号分隔（简写形式）
  analyzer-ts batch-bundle -e "./src/user.ts:User,./src/product.ts:Product" --output-dir ./dist/types/

  # 带别名 - 自定义输出类型名
  analyzer-ts batch-bundle -e "./src/user.ts:User:UserDTO" -e "./src/common.ts:CommonType:ConfigType" --output-dir ./dist/types/

功能特性:
  • 独立文件输出: 每个类型生成独立文件，避免命名冲突
  • 类型别名支持: 支持为类型定义别名，便于自定义输出
  • 智能文件命名: 自动生成唯一且清晰的文件名
  • 智能依赖收集: 自动递归收集类型的所有依赖
  • 路径别名解析: 完整支持 tsconfig.json 中的 paths 配置
  • 多种导入方式: 支持命名导入、默认导入、命名空间导入等`,
		Run: func(cmd *cobra.Command, args []string) {
			// 验证必需参数
			if len(entries) == 0 {
				fmt.Fprintf(os.Stderr, "错误: 需要指定 --entries 参数\n")
				fmt.Println("用法: analyzer-ts batch-bundle -e <入口点列表> --output-dir <目录>")
				cmd.Help()
				os.Exit(1)
			}

			if outputDir == "" {
				fmt.Fprintf(os.Stderr, "错误: 需要指定 --output-dir 参数\n")
				os.Exit(1)
			}

			// 处理逗号分隔的入口点
			var allEntries []string
			for _, entry := range entries {
				if strings.Contains(entry, ",") {
					// 支持逗号分隔的多入口点
					parts := strings.Split(entry, ",")
					for _, part := range parts {
						part = strings.TrimSpace(part)
						if part != "" {
							allEntries = append(allEntries, part)
						}
					}
				} else {
					allEntries = append(allEntries, entry)
				}
			}

			fmt.Printf("批量打包 %d 个入口类型到目录: %s\n", len(allEntries), outputDir)
			for i, entry := range allEntries {
				fmt.Printf("  %d. %s\n", i+1, entry)
			}

			results, err := ts_bundle.GenerateBatchBundlesToFiles(allEntries, projectRoot, outputDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "批量打包失败: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("批量打包完成，输出目录: %s\n", outputDir)
			fmt.Printf("成功生成 %d 个类型文件:\n", len(results))
			for _, result := range results {
				fmt.Printf("  - %s (%d 字符)\n", result.FileName, result.ContentSize)
			}
		},
	}

	// 定义命令行标志 (flags)
	bundleCmd.Flags().StringSliceVarP(&entries, "entries", "e", []string{}, "入口点列表，格式为 '文件路径:类型名[:别名]' (必需，可多次使用或逗号分隔)")
	bundleCmd.Flags().StringVar(&outputDir, "output-dir", "", "输出目录路径 (必需，每个类型生成独立文件)")
	bundleCmd.Flags().StringVarP(&projectRoot, "root", "r", "", "项目根路径 (可选，自动检测)")

	return bundleCmd
}
