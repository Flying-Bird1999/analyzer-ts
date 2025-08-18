
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"main/analyzer_plugin/project_analyzer"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	findUnreferencedInputDir   string
	findUnreferencedOutputDir  string
	findUnreferencedExclude    []string
	findUnreferencedIsMonorepo bool
)

// NewFindUnreferencedFilesCmd 创建并返回 `find-unreferenced-files` 命令
func NewFindUnreferencedFilesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "find-unreferenced-files",
		Short: "查找项目中所有未被引用的文件。",
		Long:  `该命令首先对指定的TypeScript项目进行全面分析，构建依赖关系图。然后，它会识别出那些从未被任何其他文件导入或引用的文件，并将这些文件的路径列表输出。这对于清理项目中不再使用的废弃代码非常有用。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 步骤 1: 分析项目，构建完整的依赖图
			analyzer := project_analyzer.NewProjectAnalyzer(findUnreferencedInputDir, findUnreferencedExclude, findUnreferencedIsMonorepo)
			deps, err := analyzer.Analyze()
			if err != nil {
				return fmt.Errorf("分析项目失败: %w", err)
			}

			// 步骤 2: 收集所有被引用的文件
			referencedFiles := make(map[string]bool)
			for _, fileDeps := range deps.Js_Data {
				// 从 import 声明中收集
				for _, dep := range fileDeps.ImportDeclarations {
					if dep.Source.FilePath != "" {
						referencedFiles[dep.Source.FilePath] = true
					}
				}
				// 从 export ... from 声明中收集
				for _, dep := range fileDeps.ExportDeclarations {
					if dep.Source != nil && dep.Source.FilePath != "" {
						referencedFiles[dep.Source.FilePath] = true
					}
				}
			}

			// 步骤 3: 找出所有未被引用的文件
			var unreferencedFiles []string
			allFiles := make(map[string]bool)
			for filePath := range deps.Js_Data {
				allFiles[filePath] = true
			}

			for filePath := range allFiles {
				if !referencedFiles[filePath] {
					unreferencedFiles = append(unreferencedFiles, filePath)
				}
			}

			// 步骤 4: 格式化并输出结果
			outputData := map[string]interface{}{
				"unreferencedFiles": unreferencedFiles,
				"summary": map[string]int{
					"totalFiles":           len(allFiles),
					"referencedFiles":      len(referencedFiles),
					"unreferencedFiles":    len(unreferencedFiles),
				},
			}

			outputBytes, err := json.MarshalIndent(outputData, "", "  ")
			if err != nil {
				return fmt.Errorf("无法将结果序列化为 JSON: %w", err)
			}

			if findUnreferencedOutputDir != "" {
				if err := os.MkdirAll(findUnreferencedOutputDir, os.ModePerm); err != nil {
					return fmt.Errorf("无法创建输出目录 %s: %w", findUnreferencedOutputDir, err)
				}
				baseName := filepath.Base(findUnreferencedInputDir)
				outputFileName := fmt.Sprintf("%s-unreferenced-files.json", baseName)
				fullOutputPath := filepath.Join(findUnreferencedOutputDir, outputFileName)

				if err := ioutil.WriteFile(fullOutputPath, outputBytes, 0644); err != nil {
					return fmt.Errorf("无法将输出写入文件 %s: %w", fullOutputPath, err)
				}
				fmt.Printf("未引用文件分析结果已写入: %s", fullOutputPath)
			} else {
				// 仅输出文件列表到控制台
				for _, file := range unreferencedFiles {
					fmt.Println(file)
				}
			}

			return nil
		},
	}

	// 定义所有标志
	cmd.Flags().StringVarP(&findUnreferencedInputDir, "input", "i", "", "要分析的 TypeScript 项目目录的路径 (必需)")
	cmd.Flags().StringVarP(&findUnreferencedOutputDir, "output", "o", "", "用于存储 JSON 结果的输出目录路径 (可选, 默认为标准输出)")
	cmd.Flags().StringSliceVarP(&findUnreferencedExclude, "exclude", "x", []string{}, "要从分析中排除的 Glob 模式 (可多次使用)")
	cmd.Flags().BoolVarP(&findUnreferencedIsMonorepo, "monorepo", "m", false, "如果分析的是 monorepo 项目，请设置为 true")

	// 将 --input 标记为必需
	cmd.MarkFlagRequired("input")

	return cmd
}
