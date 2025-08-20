package cmd

// go run main.go find-unreferenced-files -i /Users/bird/company/sc1.0/live/shopline-live-sale -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin/project_analyzer/result -x "node_modules/**" -x "bffApiDoc/**"

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"main/analyzer_plugin/project_analyzer"
	"os"
	"path/filepath"
	"strings"

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
				// 从 import 声明中收集（包括动态导入）
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
				// 从 JSX 元素中收集（React 组件引用）
				for _, jsx := range fileDeps.JsxElements {
					if jsx.Source.FilePath != "" {
						referencedFiles[jsx.Source.FilePath] = true
					}
				}
			}

			// 步骤 3: 处理入口文件（entrypoints）
			entrypointFiles := make(map[string]bool)
			if len(findUnreferencedEntrypoints) > 0 {
				// 如果指定了入口文件，则只考虑从这些入口文件可达的文件
				for _, entrypoint := range findUnreferencedEntrypoints {
					absEntrypoint, err := filepath.Abs(entrypoint)
					if err != nil {
						return fmt.Errorf("无法解析入口文件路径 %s: %w", entrypoint, err)
					}
					entrypointFiles[absEntrypoint] = true
				}
			} else if findUnreferencedIncludeEntryDirs {
				// 如果指定包含入口目录，则查找常见的入口文件
				commonEntrypoints := []string{
					"index.ts", "index.tsx", "main.ts", "main.tsx",
					"App.ts", "App.tsx", "src/index.ts", "src/index.tsx",
				}
				for _, entryName := range commonEntrypoints {
					entryPath := filepath.Join(findUnreferencedInputDir, entryName)
					if _, exists := deps.Js_Data[entryPath]; exists {
						entrypointFiles[entryPath] = true
					}
				}
			}

			// 步骤 4: 找出所有未被引用的文件
			var unreferencedFiles []string
			allFiles := make(map[string]bool)
			for filePath := range deps.Js_Data {
				allFiles[filePath] = true
			}

			// 如果指定了入口文件，则使用可达性分析
			if len(entrypointFiles) > 0 {
				// 从入口文件开始进行深度优先搜索，标记所有可达的文件
				visited := make(map[string]bool)
				var dfs func(string)
				dfs = func(filePath string) {
					if visited[filePath] {
						return
					}
					visited[filePath] = true

					// 遍历当前文件引用的所有文件
					fileDeps := deps.Js_Data[filePath]
					for _, dep := range fileDeps.ImportDeclarations {
						if dep.Source.FilePath != "" {
							dfs(dep.Source.FilePath)
						}
					}
					for _, dep := range fileDeps.ExportDeclarations {
						if dep.Source != nil && dep.Source.FilePath != "" {
							dfs(dep.Source.FilePath)
						}
					}
					for _, jsx := range fileDeps.JsxElements {
						if jsx.Source.FilePath != "" {
							dfs(jsx.Source.FilePath)
						}
					}
				}

				// 从所有入口文件开始搜索
				for entrypoint := range entrypointFiles {
					dfs(entrypoint)
				}

				// 只有既未被引用又不可达的文件才被认为是未引用的
				for filePath := range allFiles {
					if !referencedFiles[filePath] && !visited[filePath] {
						unreferencedFiles = append(unreferencedFiles, filePath)
					}
				}
			} else {
				// 默认行为：找出所有未被任何文件引用的文件
				for filePath := range allFiles {
					// 排除入口文件（它们可能不被引用，但是程序的入口）
					isEntrypoint := strings.HasSuffix(filePath, "index.ts") ||
						strings.HasSuffix(filePath, "index.tsx") ||
						strings.HasSuffix(filePath, "main.ts") ||
						strings.HasSuffix(filePath, "main.tsx")

					if !referencedFiles[filePath] && !isEntrypoint {
						unreferencedFiles = append(unreferencedFiles, filePath)
					}
				}
			}

			// 步骤 5: 分类文件并过滤
			var trulyUnreferencedFiles []string
			var suspiciousFiles []string

			ignoredPatterns := []string{
				".test.", ".spec.", "__tests__", "__mocks__", ".d.ts",
				".story.", ".stories.",
			}

			// 常见的配置文件名模式
			configFilePatterns := []string{
				"webpack.config", "vite.config", "rollup.config", "babel.config",
				"prettier.config", ".prettierrc", "eslint.config", ".eslintrc",
				"jest.config", "karma.conf", "gulpfile", "gruntfile",
				"tsconfig", "jsconfig", "postcss.config", "tailwind.config",
				"commitlint.config", ".commitlintrc", "lint-staged.config",
				"stylelint.config", ".stylelintrc", "nodemon.json", "nodemon-debug.json",
				"build.config", "vitest.config", "cypress.config", "playwright.config",
			}

			// 常见的入口文件模式
			entryFilePatterns := []string{
				"index.", "main.", "app.", "root.", "entry.",
			}

			// 常见的配置文件名（在src目录下）
			srcConfigPatterns := []string{
				"router", "route", "store", "state", "theme", "i18n", "locale",
				"config", "setting", "constant", "util", "helper", "service",
				"api", "http", "request", "axios", "fetch", "polyfill",
			}

			for _, filePath := range unreferencedFiles {
				// 检查是否应该忽略该文件
				shouldIgnore := false

				// 检查是否匹配忽略模式
				for _, pattern := range ignoredPatterns {
					if strings.Contains(filePath, pattern) {
						shouldIgnore = true
						break
					}
				}

				// 检查是否为配置文件
				if !shouldIgnore {
					fileName := filepath.Base(filePath)
					for _, pattern := range configFilePatterns {
						if strings.HasPrefix(fileName, pattern) {
							shouldIgnore = true
							break
						}
					}
				}

				// 检查是否在排除目录中
				if !shouldIgnore && !isInExcludedDir(filePath, findUnreferencedExclude) {
					// 分类文件
					isSuspicious := false

					// 检查是否为可疑文件（需要人工检查的文件）
					dir := filepath.Dir(filePath)

					// 简单而安全的策略：所有非src目录下的文件都视为可疑文件
					// 因为这些通常是项目配置文件，即使未被直接引用也可能很重要
					if !strings.Contains(dir, "/src/") {
						isSuspicious = true
					} else {
						// 对于src目录下的文件，检查是否为特殊文件
						fileName := filepath.Base(filePath)

						// 检查是否为入口文件
						for _, pattern := range entryFilePatterns {
							if strings.HasPrefix(fileName, pattern) {
								isSuspicious = true
								break
							}
						}

						// 检查是否为配置相关文件
						if !isSuspicious {
							for _, pattern := range srcConfigPatterns {
								if strings.Contains(fileName, pattern) {
									isSuspicious = true
									break
								}
							}
						}
					}

					if isSuspicious {
						suspiciousFiles = append(suspiciousFiles, filePath)
					} else {
						trulyUnreferencedFiles = append(trulyUnreferencedFiles, filePath)
					}
				}
			}

			// 步骤 6: 格式化并输出结果
			outputData := map[string]interface{}{
				"trulyUnreferencedFiles": trulyUnreferencedFiles,
				"suspiciousFiles":        suspiciousFiles,
				"entrypointFiles":        getKeys(entrypointFiles),
				"summary": map[string]int{
					"totalFiles":             len(allFiles),
					"referencedFiles":        len(referencedFiles),
					"trulyUnreferencedFiles": len(trulyUnreferencedFiles),
					"suspiciousFiles":        len(suspiciousFiles),
				},
				"configuration": map[string]interface{}{
					"inputDir":             findUnreferencedInputDir,
					"entrypointsSpecified": len(findUnreferencedEntrypoints) > 0,
					"includeEntryDirs":     findUnreferencedIncludeEntryDirs,
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
				outputFileName := fmt.Sprintf("%s_find_unreferenced_files.json", baseName)
				fullOutputPath := filepath.Join(findUnreferencedOutputDir, outputFileName)

				if err := ioutil.WriteFile(fullOutputPath, outputBytes, 0644); err != nil {
					return fmt.Errorf("无法将输出写入文件 %s: %w", fullOutputPath, err)
				}
				fmt.Printf("未引用文件分析结果已写入: %s\n", fullOutputPath)

				// 同时输出简要的分类文件列表
				fmt.Printf("\n发现 %d 个真正未引用的文件:\n", len(trulyUnreferencedFiles))
				for _, file := range trulyUnreferencedFiles {
					fmt.Printf("  - %s\n", file)
				}

				if len(suspiciousFiles) > 0 {
					fmt.Printf("\n发现 %d 个可疑文件 (可能重要但未被引用):\n", len(suspiciousFiles))
					for _, file := range suspiciousFiles {
						fmt.Printf("  - %s\n", file)
					}
				}
			} else {
				// 仅输出文件列表到控制台
				fmt.Printf("真正未引用的文件:\n")
				for _, file := range trulyUnreferencedFiles {
					fmt.Println(file)
				}

				if len(suspiciousFiles) > 0 {
					fmt.Printf("\n可疑文件 (可能重要但未被引用):\n")
					for _, file := range suspiciousFiles {
						fmt.Println(file)
					}
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
	cmd.Flags().StringSliceVarP(&findUnreferencedEntrypoints, "entrypoints", "e", []string{}, "指定入口文件路径 (可多次使用)")
	cmd.Flags().BoolVar(&findUnreferencedIncludeEntryDirs, "include-entry-dirs", false, "自动包含常见的入口目录文件")

	// 将 --input 标记为必需
	cmd.MarkFlagRequired("input")

	return cmd
}

// isInExcludedDir 检查文件是否在排除的目录中
func isInExcludedDir(filePath string, excludePatterns []string) bool {
	// 简单实现，可以扩展为更复杂的 glob 模式匹配
	for _, pattern := range excludePatterns {
		if strings.Contains(filePath, strings.TrimSuffix(pattern, "/**")) {
			return true
		}
	}
	return false
}

// getKeys 获取 map 的所有键
func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
