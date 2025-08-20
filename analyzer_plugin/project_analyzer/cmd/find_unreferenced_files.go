package cmd

// go run main.go find-unreferenced-files -i /Users/bird/company/sc1.0/live/shopline-live-sale -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin/project_analyzer/result -x "node_modules/**" -x "bffApiDoc/**"

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"main/analyzer_plugin/project_analyzer/unreferenced"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	findUnreferencedInputDir         string
	findUnreferencedOutputDir        string
	findUnreferencedExclude          []string
	findUnreferencedIsMonorepo       bool
	findUnreferencedEntrypoints      []string
	findUnreferencedIncludeEntryDirs bool
)

// NewFindUnreferencedFilesCmd 创建并返回 `find-unreferenced-files` 命令
func NewFindUnreferencedFilesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "find-unreferenced-files",
		Short: "查找项目中所有未被引用的文件。",
		Long:  `该命令首先对指定的TypeScript项目进行全面分析，构建依赖关系图。然后，它会识别出那些从未被任何其他文件导入或引用的文件，并将这些文件的路径列表输出。这对于清理项目中不再使用的废弃代码非常有用。`,
		Args:  cobra.NoArgs,
		// RunE 是命令的执行入口，它现在只负责参数的传递和结果的输出。
		RunE: func(cmd *cobra.Command, args []string) error {
			// 步骤 1: 将从命令行解析出的参数，打包成一个结构体，传递给业务逻辑层。
			params := unreferenced.Params{
				RootPath:         findUnreferencedInputDir,
				Exclude:          findUnreferencedExclude,
				IsMonorepo:       findUnreferencedIsMonorepo,
				Entrypoints:      findUnreferencedEntrypoints,
				IncludeEntryDirs: findUnreferencedIncludeEntryDirs,
			}

			// 步骤 2: 调用 unreferenced 包中的核心业务逻辑函数。
			result, err := unreferenced.Find(params)
			if err != nil {
				return err
			}

			// 步骤 3: 根据用户指定的输出方式，格式化并输出结果。
			if findUnreferencedOutputDir != "" {
				// 输出为 JSON 文件
				outputBytes, err := json.MarshalIndent(result, "", "  ")
				if err != nil {
					return fmt.Errorf("无法将结果序列化为 JSON: %w", err)
				}

				if err := os.MkdirAll(findUnreferencedOutputDir, os.ModePerm); err != nil {
					return fmt.Errorf("无法创建输出目录 %s: %w", findUnreferencedOutputDir, err)
				}
				baseName := filepath.Base(findUnreferencedInputDir)
				outputFileName := fmt.Sprintf("%s_find_unreferenced_files.json", baseName)
				fullOutputPath := filepath.Join(findUnreferencedOutputDir, outputFileName)

				if err := ioutil.WriteFile(fullOutputPath, outputBytes, 0644); err != nil {
					return fmt.Errorf("无法将输出写入文件 %s: %w", fullOutputPath, err)
				}
				fmt.Printf("未引用文件分析结果已写入: %s", fullOutputPath)
			}

			return nil
		},
	}

	// 定义所有命令行标志
	cmd.Flags().StringVarP(&findUnreferencedInputDir, "input", "i", "", "要分析的 TypeScript 项目目录的路径 (必需)")
	cmd.Flags().StringVarP(&findUnreferencedOutputDir, "output", "o", "", "用于存储 JSON 结果的输出目录路径 (可选, 默认为标准输出)")
	cmd.Flags().StringSliceVarP(&findUnreferencedExclude, "exclude", "x", []string{}, "要从分析中排除的 Glob 模式 (可多次使用)")
	cmd.Flags().BoolVarP(&findUnreferencedIsMonorepo, "monorepo", "m", false, "如果分析的是 monorepo 项目，请设置为 true")
	cmd.Flags().StringSliceVarP(&findUnreferencedEntrypoints, "entrypoints", "e", []string{}, "指定入口文件路径 (可多次使用)")
	cmd.Flags().BoolVar(&findUnreferencedIncludeEntryDirs, "include-entry-dirs", false, "自动包含常见的入口目录文件")

	cmd.MarkFlagRequired("input")

	return cmd
}
