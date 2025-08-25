// package unreferenced 实现了查找项目中未被任何其他文件引用的“死代码”的核心业务逻辑。
package unreferenced

import (
	"fmt"
	"main/analyzer/projectParser"
	projectanalyzer "main/analyzer_plugin/project_analyzer"
	"path/filepath"
	"strconv"
	"strings"
)

// Finder 是“未引用文件”分析器的实现。
type Finder struct {
	entrypoints      []string
	includeEntryDirs bool
}

var _ projectanalyzer.Analyzer = (*Finder)(nil)

func (f *Finder) Name() string {
	return "find-unreferenced-files"
}

func (f *Finder) Configure(params map[string]string) error {
	if entrypoints, ok := params["entrypoint"]; ok {
		f.entrypoints = strings.Split(entrypoints, ",")
	}
	if include, ok := params["include-entry-dirs"]; ok {
		includeBool, err := strconv.ParseBool(include)
		if err != nil {
			return fmt.Errorf("无效的布尔值 for include-entry-dirs: %s", include)
		}
		f.includeEntryDirs = includeBool
	}
	return nil
}

func (f *Finder) Analyze(ctx *projectanalyzer.ProjectContext) (projectanalyzer.Result, error) {
	deps := ctx.ParsingResult

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

	entrypointFiles := make(map[string]bool)
	if len(f.entrypoints) > 0 {
		for _, entrypoint := range f.entrypoints {
			absEntrypoint, err := filepath.Abs(entrypoint)
			if err != nil {
				return nil, fmt.Errorf("无法解析入口文件路径 %s: %w", entrypoint, err)
			}
			entrypointFiles[absEntrypoint] = true
		}
	} else if f.includeEntryDirs {
		commonEntrypoints := []string{
			"index.ts", "index.tsx", "main.ts", "main.tsx",
			"App.ts", "App.tsx", "src/index.ts", "src/index.tsx",
		}
		for _, entryName := range commonEntrypoints {
			entryPath := filepath.Join(ctx.ProjectRoot, entryName)
			if _, exists := deps.Js_Data[entryPath]; exists {
				entrypointFiles[entryPath] = true
			}
		}
	}

	var unreferencedFiles []string
	allFiles := make(map[string]bool)
	for filePath := range deps.Js_Data {
		allFiles[filePath] = true
	}

	if len(entrypointFiles) > 0 {
		visited := performDFS(entrypointFiles, deps)
		for filePath := range allFiles {
			if !referencedFiles[filePath] && !visited[filePath] {
				unreferencedFiles = append(unreferencedFiles, filePath)
			}
		}
	} else {
		for filePath := range allFiles {
			isEntrypoint := strings.HasSuffix(filePath, "index.ts") ||
				strings.HasSuffix(filePath, "index.tsx") ||
				strings.HasSuffix(filePath, "main.ts") ||
				strings.HasSuffix(filePath, "main.tsx")

			if !referencedFiles[filePath] && !isEntrypoint {
				unreferencedFiles = append(unreferencedFiles, filePath)
			}
		}
	}

	trulyUnreferencedFiles, suspiciousFiles := classifyFiles(unreferencedFiles, ctx.Exclude)

	finalResult := &FindUnreferencedFilesResult{
		Configuration: AnalysisConfiguration{
			InputDir:             ctx.ProjectRoot,
			EntrypointsSpecified: len(f.entrypoints) > 0,
			IncludeEntryDirs:     f.includeEntryDirs,
		},
		Stats: SummaryStats{
			TotalFiles:             len(allFiles),
			ReferencedFiles:        len(referencedFiles),
			TrulyUnreferencedFiles: len(trulyUnreferencedFiles),
			SuspiciousFiles:        len(suspiciousFiles),
		},
		EntrypointFiles:        getKeys(entrypointFiles),
		SuspiciousFiles:        suspiciousFiles,
		TrulyUnreferencedFiles: trulyUnreferencedFiles,
	}

	return finalResult, nil
}

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

	for entrypoint := range entrypointFiles {
		dfs(entrypoint)
	}
	return visited
}

func classifyFiles(unreferencedFiles []string, excludePatterns []string) (trulyUnreferenced []string, suspicious []string) {
	ignoredPatterns := []string{
		".test.", ".spec.", "__tests__", "__mocks__", ".d.ts",
		".story.", ".stories.",
	}
	configFilePatterns := []string{
		"webpack.config", "vite.config", "rollup.config", "babel.config",
		"prettier.config", ".prettierrc", "eslint.config", ".eslintrc",
		"jest.config", "karma.conf", "gulpfile", "gruntfile",
		"tsconfig", "jsconfig", "postcss.config", "tailwind.config",
		"commitlint.config", ".commitlintrc", "lint-staged.config",
		"stylelint.config", ".stylelintrc", "nodemon.json", "nodemon-debug.json",
		"build.config", "vitest.config", "cypress.config", "playwright.config",
	}
	entryFilePatterns := []string{
		"index.", "main.", "app.", "root.", "entry.",
	}
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

		if !strings.Contains(dir, "/src/") {
			isSuspicious = true
		} else {
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

func isInExcludedDir(filePath string, excludePatterns []string) bool {
	for _, pattern := range excludePatterns {
		if strings.Contains(filePath, strings.TrimSuffix(pattern, "/**")) {
			return true
		}
	}
	return false
}

func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
