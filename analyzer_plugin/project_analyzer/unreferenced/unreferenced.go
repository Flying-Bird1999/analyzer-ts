// Package unreferenced 实现了查找项目中未被任何其他文件引用的"死代码"的核心业务逻辑。
//
// 功能概述：
// 该分析器专门用于识别项目中没有被任何其他文件引用的"死代码"文件。
// 通过静态分析项目中的导入、导出和引用关系，构建完整的依赖图，
// 然后识别出那些孤立存在、没有被任何其他文件使用的文件。
//
// 技术原理：
// 1. 构建引用关系图：分析所有文件的导入和导出语句
// 2. 识别入口文件：根据用户指定或常见入口文件模式识别项目入口
// 3. 深度优先遍历：从入口文件开始，遍历所有可达的文件
// 4. 分类未引用文件：使用智能分类，区分真正的死代码和潜在重要的文件
//
// 主要用途：
// 1. 代码清理：安全地删除无用的代码文件
// 2. 架构优化：识别孤立的模块和组件
// 3. 包体积优化：减少最终打包的大小
// 4. 维护性改进：简化项目结构，提高代码可维护性
//
// 核心特点：
// - 支持自定义入口文件配置
// - 智能文件分类，避免误删重要文件
// - 深度优先搜索算法，确保分析的完整性
// - 支持复杂的 re-export 和动态导入场景
package unreferenced

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

// =============================================================================
// 分析器主体定义
// =============================================================================

// Finder 是"未引用文件"分析器的实现。
//
// 设计理念：
// 该分析器采用图论算法，将项目中的文件视为图的节点，
// 文件间的引用关系视为边，通过深度优先搜索(DFS)算法
// 识别从入口文件可达的文件，从而发现不可达的"死代码"文件。
//
// 配置选项：
// - 入口文件：支持用户自定义入口文件路径
// - 包含入口目录：自动识别常见的入口文件模式
// - 文件排除：支持 glob 模式排除特定目录
//
// 智能分类：
// 使用多层次的分类策略，将未引用文件分为：
// 1. 真正的未引用文件：可以安全删除的文件
// 2. 可疑文件：可能被间接使用或特殊的配置文件
type Finder struct {
	entrypoints      []string // 自定义入口文件路径列表
	includeEntryDirs bool     // 是否包含常见的入口目录模式
}

// 确保 Finder 实现了 projectanalyzer.Analyzer 接口
var _ projectanalyzer.Analyzer = (*Finder)(nil)

// Name 返回分析器的唯一标识符。
//
// 返回值说明：
// 返回 "find-unreferenced-files" 作为分析器的标识符。
// 这个名称用于在插件系统中注册和识别该分析器。
func (f *Finder) Name() string {
	return "find-unreferenced-files"
}

// Configure 配置分析器的参数。
//
// 支持的配置参数：
// 1. entrypoint: 指定入口文件路径，支持多个入口（逗号分隔）
// 2. include-entry-dirs: 是否自动包含常见的入口目录模式
//
// 参数处理逻辑：
// - entrypoint: 字符串类型，多个入口文件用逗号分隔
// - include-entry-dirs: 布尔类型，控制是否自动识别常见入口文件
//
// 错误处理：
// - 布尔值解析失败时返回详细的错误信息
// - 参数格式错误时提供清晰的错误提示
//
// 使用示例：
// ```bash
// ./analyzer-ts analyze find-unreferenced-files -i /path/to/project -p "unreferenced.entrypoint=src/index.ts"
// ./analyzer-ts analyze find-unreferenced-files -i /path/to/project -p "unreferenced.include-entry-dirs=true"
// ```
func (f *Finder) Configure(params map[string]string) error {
	// 处理入口文件参数
	if entrypoints, ok := params["entrypoint"]; ok {
		f.entrypoints = strings.Split(entrypoints, ",")
		// 清理可能的空值
		for i, entry := range f.entrypoints {
			f.entrypoints[i] = strings.TrimSpace(entry)
		}
	}

	// 处理包含入口目录参数
	if include, ok := params["include-entry-dirs"]; ok {
		includeBool, err := strconv.ParseBool(include)
		if err != nil {
			return fmt.Errorf("无效的布尔值 for include-entry-dirs: %s", include)
		}
		f.includeEntryDirs = includeBool
	}

	return nil
}

// Analyze 执行未引用文件分析的核心逻辑。
//
// 分析流程：
// 该方法实现了完整的未引用文件检测流程，包含以下主要步骤：
//
// 1. 构建引用关系图：
//    - 遍历所有文件的导入语句
//    - 遍历所有文件的导出语句
//    - 遍历所有文件的 JSX 组件引用
//    - 建立文件间的完整引用关系
//
// 2. 识别入口文件：
//    - 使用用户指定的入口文件
//    - 或自动识别常见的入口文件模式
//    - 将入口文件作为图的起始节点
//
// 3. 执行可达性分析：
//    - 从入口文件开始执行深度优先搜索
//    - 标记所有从入口可达的文件
//    - 识别不可达的未引用文件
//
// 4. 智能文件分类：
//    - 将未引用文件分为真正的未引用文件和可疑文件
//    - 应用各种过滤规则避免误删重要文件
//    - 考虑排除模式和文件类型
//
// 参数说明：
// - ctx: 项目上下文，包含完整的解析结果和项目信息
//
// 返回值说明：
// - projectanalyzer.Result: 包含未引用文件分析结果的对象
// - error: 分析过程中遇到的错误（路径解析错误等）
func (f *Finder) Analyze(ctx *projectanalyzer.ProjectContext) (projectanalyzer.Result, error) {
	deps := ctx.ParsingResult

	// 步骤 1: 构建引用关系图
	// 收集所有被其他文件引用的文件
	referencedFiles := make(map[string]bool)
	for _, fileDeps := range deps.Js_Data {
		// 分析导入语句
		for _, dep := range fileDeps.ImportDeclarations {
			if dep.Source.FilePath != "" {
				referencedFiles[dep.Source.FilePath] = true
			}
		}
		// 分析导出语句
		for _, dep := range fileDeps.ExportDeclarations {
			if dep.Source != nil && dep.Source.FilePath != "" {
				referencedFiles[dep.Source.FilePath] = true
			}
		}
		// 分析 JSX 组件引用
		for _, jsx := range fileDeps.JsxElements {
			if jsx.Source.FilePath != "" {
				referencedFiles[jsx.Source.FilePath] = true
			}
		}
	}

	// 步骤 2: 识别入口文件
	entrypointFiles := make(map[string]bool)
	if len(f.entrypoints) > 0 {
		// 使用用户指定的入口文件
		for _, entrypoint := range f.entrypoints {
			absEntrypoint, err := filepath.Abs(entrypoint)
			if err != nil {
				return nil, fmt.Errorf("无法解析入口文件路径 %s: %w", entrypoint, err)
			}
			entrypointFiles[absEntrypoint] = true
		}
	} else if f.includeEntryDirs {
		// 自动识别常见的入口文件模式
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

	// 步骤 3: 执行可达性分析
	// 识别未引用的文件
	var unreferencedFiles []string
	allFiles := make(map[string]bool)
	for filePath := range deps.Js_Data {
		allFiles[filePath] = true
	}

	if len(entrypointFiles) > 0 {
		// 使用深度优先搜索分析可达性
		visited := performDFS(entrypointFiles, deps)
		for filePath := range allFiles {
			if !referencedFiles[filePath] && !visited[filePath] {
				unreferencedFiles = append(unreferencedFiles, filePath)
			}
		}
	} else {
		// 简单的启发式分析（无入口文件时）
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

	// 步骤 4: 智能文件分类
	// 使用启发式规则分类未引用文件
	trulyUnreferencedFiles, suspiciousFiles := classifyFiles(unreferencedFiles, ctx.Exclude)

	// 构建最终结果对象
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

// performDFS 执行深度优先搜索算法，识别从入口文件可达的所有文件。
//
// 算法原理：
// 深度优先搜索（DFS）是一种图遍历算法，从入口文件开始，
// 递归地访问所有通过导入、导出或JSX引用可达的文件。
//
// 核心功能：
// - 构建可达性图：从入口文件开始，标记所有可达的文件
// - 避免循环引用：通过 visited 集合防止无限递归
// - 多类型引用：支持导入、导出、JSX等多种引用类型
// - 完整遍历：确保从入口开始的所有文件都被访问到
//
// 遍历顺序：
// 1. 从所有入口文件开始遍历
// 2. 对每个文件，分析其所有的引用关系
// 3. 递归访问被引用的文件
// 4. 标记所有访问过的文件
//
// 参数说明：
// - entrypointFiles: 入口文件的集合，作为遍历的起点
// - deps: 项目解析结果，包含所有文件的依赖关系
//
// 返回值说明：
// - map[string]bool: 从入口文件可达的文件集合
//   key 为文件路径，value 为可达性标记（true 表示可达）
func performDFS(entrypointFiles map[string]bool, deps *projectParser.ProjectParserResult) map[string]bool {
	visited := make(map[string]bool)

	// 递归遍历函数
	var dfs func(string)
	dfs = func(filePath string) {
		// 如果已经访问过，直接返回（防止循环引用）
		if visited[filePath] {
			return
		}
		// 标记当前文件为已访问
		visited[filePath] = true

		// 获取当前文件的依赖关系
		fileDeps, exists := deps.Js_Data[filePath]
		if !exists {
			return
		}

		// 遍历导入引用
		for _, dep := range fileDeps.ImportDeclarations {
			if dep.Source.FilePath != "" {
				dfs(dep.Source.FilePath)
			}
		}
		// 遍历导出引用
		for _, dep := range fileDeps.ExportDeclarations {
			if dep.Source != nil && dep.Source.FilePath != "" {
				dfs(dep.Source.FilePath)
			}
		}
		// 遍历 JSX 组件引用
		for _, jsx := range fileDeps.JsxElements {
			if jsx.Source.FilePath != "" {
				dfs(jsx.Source.FilePath)
			}
		}
	}

	// 从所有入口文件开始遍历
	for entrypoint := range entrypointFiles {
		dfs(entrypoint)
	}
	return visited
}

// classifyFiles 智能分类未引用文件，区分真正的死代码和潜在重要的文件。
//
// 分类策略：
// 该函数采用多层次的启发式规则，将未引用文件分为两类：
// 1. trulyUnreferenced: 真正的未引用文件，可以安全删除
// 2. suspicious: 可疑文件，可能被间接使用或具有特殊用途
//
// 分类规则层次：
//
// 层次 1: 忽略规则（直接忽略）
// - 测试文件：包含 .test.、.spec.、__tests__ 等模式
// - 故事文件：包含 .story.、.stories. 等模式
// - 类型声明：包含 .d.ts 文件
//
// 层次 2: 配置文件识别（标记为可疑）
// - 构建配置：webpack、vite、rollup、babel 等配置
// - 代码质量：prettier、eslint、stylelint 等配置
// - 测试配置：jest、karma、cypress 等配置
// - 项目配置：tsconfig、postcss、tailwind 等配置
//
// 层次 3: 位置和命名分析
// - 非 src 目录：不在 src 目录下的文件标记为可疑
// - 入口文件：包含 index、main、app 等模式的文件标记为可疑
// - 核心模块：包含路由、状态、主题等模式的文件标记为可疑
//
// 参数说明：
// - unreferencedFiles: 需要分类的未引用文件列表
// - excludePatterns: 用户指定的排除模式
//
// 返回值说明：
// - trulyUnreferenced: 真正的未引用文件，可以安全删除
// - suspicious: 可疑文件，需要人工确认
func classifyFiles(unreferencedFiles []string, excludePatterns []string) (trulyUnreferenced []string, suspicious []string) {
	// 层次 1: 忽略规则（直接忽略）
	ignoredPatterns := []string{
		".test.", ".spec.", "__tests__", "__mocks__", ".d.ts",
		".story.", ".stories.",
	}

	// 层次 2: 配置文件识别（标记为可疑）
	configFilePatterns := []string{
		"webpack.config", "vite.config", "rollup.config", "babel.config",
		"prettier.config", ".prettierrc", "eslint.config", ".eslintrc",
		"jest.config", "karma.conf", "gulpfile", "gruntfile",
		"tsconfig", "jsconfig", "postcss.config", "tailwind.config",
		"commitlint.config", ".commitlintrc", "lint-staged.config",
		"stylelint.config", ".stylelintrc", "nodemon.json", "nodemon-debug.json",
		"build.config", "vitest.config", "cypress.config", "playwright.config",
	}

	// 层次 3: 入口文件模式识别
	entryFilePatterns := []string{
		"index.", "main.", "app.", "root.", "entry.",
	}

	// 层次 3: 核心模块模式识别
	srcConfigPatterns := []string{
		"router", "route", "store", "state", "theme", "i18n", "locale",
		"config", "setting", "constant", "util", "helper", "service",
		"api", "http", "request", "axios", "fetch", "polyfill",
	}

	// 遍历所有未引用文件进行分类
	for _, filePath := range unreferencedFiles {
		shouldIgnore := false
		fileName := filepath.Base(filePath)

		// 检查是否应该忽略
		for _, pattern := range ignoredPatterns {
			if strings.Contains(filePath, pattern) {
				shouldIgnore = true
				break
			}
		}
		if shouldIgnore {
			continue
		}

		// 检查是否为配置文件
		for _, pattern := range configFilePatterns {
			if strings.HasPrefix(fileName, pattern) {
				shouldIgnore = true
				break
			}
		}
		if shouldIgnore {
			continue
		}

		// 检查是否在排除目录中
		if isInExcludedDir(filePath, excludePatterns) {
			continue
		}

		// 进行可疑文件分析
		isSuspicious := false
		dir := filepath.Dir(filePath)

		// 规则 1: 不在 src 目录下的文件标记为可疑
		isInChildrenOfSrc := false
		for _, part := range strings.Split(dir, string(filepath.Separator)) {
			if part == "src" {
				isInChildrenOfSrc = true
				break
			}
		}
		if !isInChildrenOfSrc {
			isSuspicious = true
		} else {
			// 规则 2: 入口文件模式标记为可疑
			for _, pattern := range entryFilePatterns {
				if strings.HasPrefix(fileName, pattern) {
					isSuspicious = true
					break
				}
			}
			// 规则 3: 核心模块模式标记为可疑
			if !isSuspicious {
				for _, pattern := range srcConfigPatterns {
					if strings.Contains(fileName, pattern) {
						isSuspicious = true
						break
					}
				}
			}
		}

		// 根据分类结果分别存储
		if isSuspicious {
			suspicious = append(suspicious, filePath)
		} else {
			trulyUnreferenced = append(trulyUnreferenced, filePath)
		}
	}
	return
}

// isInExcludedDir 检查文件路径是否匹配用户指定的排除模式。
//
// 功能概述：
// 该函数检查给定的文件路径是否被包含在用户指定的排除模式中。
// 支持简单的字符串包含匹配，适用于常见的目录排除场景。
//
// 匹配逻辑：
// 对于每个排除模式，移除 "/**" 后缀（如果存在），
// 然后检查文件路径是否包含该模式字符串。
//
// 参数说明：
// - filePath: 需要检查的文件路径
// - excludePatterns: 用户指定的排除模式列表
//
// 返回值说明：
// - bool: 如果文件路径匹配任何排除模式，返回 true；否则返回 false
//
// 使用示例：
// ```go
// isInExcludedDir("/src/components/Button.tsx", []string{"node_modules/**"}) // 返回 false
// isInExcludedDir("/node_modules/lodash/index.js", []string{"node_modules/**"}) // 返回 true
// ```
func isInExcludedDir(filePath string, excludePatterns []string) bool {
	for _, pattern := range excludePatterns {
		if strings.Contains(filePath, strings.TrimSuffix(pattern, "/**")) {
			return true
		}
	}
	return false
}

// getKeys 提取 map 中所有的键，返回为字符串切片。
//
// 功能概述：
// 这是一个通用的辅助函数，用于将 map 的键提取为有序的字符串切片。
// 在未引用文件分析器中，主要用于提取入口文件列表。
//
// 性能考虑：
// 预先分配切片容量以减少内存分配，提高性能。
// 虽然是简单的工具函数，但遵循了 Go 的最佳实践。
//
// 参数说明：
// - m: 需要提取键的 map，键类型为 string，值类型为 bool
//
// 返回值说明：
// - []string: 包含所有 map 键的字符串切片，顺序不确定
func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
