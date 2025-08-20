// package unreferenced 实现了查找项目中未被任何其他文件引用的“死代码”的核心业务逻辑。
package unreferenced

import (
	"fmt"
	"main/analyzer/projectParser"
	"main/analyzer_plugin/project_analyzer"
	"path/filepath"
	"strings"
)

// Find 在指定的项目中查找所有未被引用的文件。
// 它首先构建出完整的依赖关系图，然后找出所有入度为 0 的文件节点。
// 如果用户指定了入口文件，它将执行可达性分析，找出所有从入口文件出发不可达的节点。
// 最后，它会对结果进行分类，区分出“真正”的未使用文件和“可疑”的未使用文件。
func Find(params Params) (*Result, error) {
	// 步骤 1: 分析项目，构建完整的依赖图
	analyzer := project_analyzer.NewProjectAnalyzer(params.RootPath, params.Exclude, params.IsMonorepo)
	deps, err := analyzer.Analyze()
	if err != nil {
		return nil, fmt.Errorf("分析项目失败: %w", err)
	}

	// 步骤 2: 收集所有被其他文件引用过的文件，存入一个集合中以便快速查找。
	referencedFiles := make(map[string]bool)
	for _, fileDeps := range deps.Js_Data {
		for _, dep := range fileDeps.ImportDeclarations {
			if dep.Source.FilePath != "" {
				referencedFiles[dep.Source.FilePath] = true
			}
		}
		for _, dep := range fileDeps.ExportDeclarations {
			if dep.Source != nil && dep.Source.FilePath != "" {
				referencedFiles[dep.Source.FilePath] = true
			}
		}
		for _, jsx := range fileDeps.JsxElements {
			if jsx.Source.FilePath != "" {
				referencedFiles[jsx.Source.FilePath] = true
			}
		}
	}

	// 步骤 3: 根据用户输入，确定分析的入口文件。
	entrypointFiles := make(map[string]bool)
	if len(params.Entrypoints) > 0 {
		for _, entrypoint := range params.Entrypoints {
			absEntrypoint, err := filepath.Abs(entrypoint)
			if err != nil {
				return nil, fmt.Errorf("无法解析入口文件路径 %s: %w", entrypoint, err)
			}
			entrypointFiles[absEntrypoint] = true
		}
	} else if params.IncludeEntryDirs {
		commonEntrypoints := []string{
			"index.ts", "index.tsx", "main.ts", "main.tsx",
			"App.ts", "App.tsx", "src/index.ts", "src/index.tsx",
		}
		for _, entryName := range commonEntrypoints {
			entryPath := filepath.Join(params.RootPath, entryName)
			if _, exists := deps.Js_Data[entryPath]; exists {
				entrypointFiles[entryPath] = true
			}
		}
	}

	// 步骤 4: 找出所有未被引用的文件。
	var unreferencedFiles []string
	allFiles := make(map[string]bool)
	for filePath := range deps.Js_Data {
		allFiles[filePath] = true
	}

	// 如果指定了入口文件，则使用可达性分析。
	if len(entrypointFiles) > 0 {
		// 从入口文件开始进行深度优先搜索，标记所有可达的文件。
		visited := performDFS(entrypointFiles, deps)
		// 如果一个文件既没有被其他文件引用，也无法从入口点访问到，则认为它是未引用的。
		for filePath := range allFiles {
			if !referencedFiles[filePath] && !visited[filePath] {
				unreferencedFiles = append(unreferencedFiles, filePath)
			}
		}
	} else {
		// 默认行为：找出所有未被任何其他文件引用的文件。
		for filePath := range allFiles {
			// 排除常见的入口文件名，因为它们本身可能不会被引用。
			isEntrypoint := strings.HasSuffix(filePath, "index.ts") ||
				strings.HasSuffix(filePath, "index.tsx") ||
				strings.HasSuffix(filePath, "main.ts") ||
				strings.HasSuffix(filePath, "main.tsx")

			if !referencedFiles[filePath] && !isEntrypoint {
				unreferencedFiles = append(unreferencedFiles, filePath)
			}
		}
	}

	// 步骤 5: 使用启发式规则对未引用文件进行分类和过滤。
	trulyUnreferencedFiles, suspiciousFiles := classifyFiles(unreferencedFiles, params.Exclude)

	// 步骤 6: 格式化并返回最终的结构化结果。
	result := &Result{
		Configuration: AnalysisConfiguration{
			InputDir:             params.RootPath,
			EntrypointsSpecified: len(params.Entrypoints) > 0,
			IncludeEntryDirs:     params.IncludeEntryDirs,
		},
		Summary: SummaryStats{
			TotalFiles:             len(allFiles),
			ReferencedFiles:        len(referencedFiles),
			TrulyUnreferencedFiles: len(trulyUnreferencedFiles),
			SuspiciousFiles:        len(suspiciousFiles),
		},
		EntrypointFiles:        getKeys(entrypointFiles),
		SuspiciousFiles:        suspiciousFiles,
		TrulyUnreferencedFiles: trulyUnreferencedFiles,
	}

	return result, nil
}

// performDFS 从一组入口文件开始，通过深度优先搜索遍历依赖关系图，返回所有可达的文件集合。
func performDFS(entrypointFiles map[string]bool, deps *projectParser.ProjectParserResult) map[string]bool {
	visited := make(map[string]bool)
	var dfs func(string)
	dfs = func(filePath string) {
		if visited[filePath] {
			return
		}
		visited[filePath] = true

		fileDeps, exists := deps.Js_Data[filePath]
		if !exists {
			return
		}

		// 递归访问当前文件所依赖的所有文件。
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

	// 从所有指定的入口文件开始遍历。
	for entrypoint := range entrypointFiles {
		dfs(entrypoint)
	}
	return visited
}

// classifyFiles 使用一系列启发式规则，将未引用的文件分为“真正未引用”和“可疑”两类。
// “可疑”文件通常是配置文件、路由文件或一些公共的工具函数，它们虽然未被直接引用，但可能很重要。
func classifyFiles(unreferencedFiles []string, excludePatterns []string) (trulyUnreferenced []string, suspicious []string) {
	// 忽略测试文件、类型定义文件、故事书等。
	ignoredPatterns := []string{
		".test.", ".spec.", "__tests__", "__mocks__", ".d.ts",
		".story.", ".stories.",
	}
	// 忽略项目根目录下的常见配置文件。
	configFilePatterns := []string{
		"webpack.config", "vite.config", "rollup.config", "babel.config",
		"prettier.config", ".prettierrc", "eslint.config", ".eslintrc",
		"jest.config", "karma.conf", "gulpfile", "gruntfile",
		"tsconfig", "jsconfig", "postcss.config", "tailwind.config",
		"commitlint.config", ".commitlintrc", "lint-staged.config",
		"stylelint.config", ".stylelintrc", "nodemon.json", "nodemon-debug.json",
		"build.config", "vitest.config", "cypress.config", "playwright.config",
	}
	// 常见的入口文件模式，这些文件即使未被引用也应被视为可疑。
	entryFilePatterns := []string{
		"index.", "main.", "app.", "root.", "entry.",
	}
	// src 目录下的常见配置文件或功能性文件，这些也应被视为可疑。
	srcConfigPatterns := []string{
		"router", "route", "store", "state", "theme", "i18n", "locale",
		"config", "setting", "constant", "util", "helper", "service",
		"api", "http", "request", "axios", "fetch", "polyfill",
	}

	for _, filePath := range unreferencedFiles {
		shouldIgnore := false
		fileName := filepath.Base(filePath)

		for _, pattern := range ignoredPatterns {
			if strings.Contains(filePath, pattern) {
				shouldIgnore = true
				break
			}
		}
		if shouldIgnore {
			continue
		}

		for _, pattern := range configFilePatterns {
			if strings.HasPrefix(fileName, pattern) {
				shouldIgnore = true
				break
			}
		}
		if shouldIgnore {
			continue
		}

		if isInExcludedDir(filePath, excludePatterns) {
			continue
		}

		isSuspicious := false
		dir := filepath.Dir(filePath)

		// 位于 src 目录外的文件通常是项目级的配置，标记为可疑。
		if !strings.Contains(dir, "/src/") {
			isSuspicious = true
		} else {
			// 检查是否匹配 src 内的可疑文件模式。
			for _, pattern := range entryFilePatterns {
				if strings.HasPrefix(fileName, pattern) {
					isSuspicious = true
					break
				}
			}
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
			suspicious = append(suspicious, filePath)
		} else {
			trulyUnreferenced = append(trulyUnreferenced, filePath)
		}
	}
	return
}

// isInExcludedDir 检查给定的文件路径是否匹配任何排除模式。
func isInExcludedDir(filePath string, excludePatterns []string) bool {
	for _, pattern := range excludePatterns {
		if strings.Contains(filePath, strings.TrimSuffix(pattern, "/**")) {
			return true
		}
	}
	return false
}

// getKeys 从一个 map[string]bool 中提取所有的键，并返回一个字符串切片。
func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
