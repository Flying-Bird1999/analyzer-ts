// Package mr_component_impact 提供 MR 组件影响分析功能
package mr_component_impact

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/component_deps"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/export_call"
)

// =============================================================================
// 便捷 API - 外部调用接口
// =============================================================================

// AnalyzeConfig 分析配置
// 提供简化的配置方式，内部自动处理项目解析等复杂逻辑
type AnalyzeConfig struct {
	// ProjectRoot 项目根目录（必需）
	ProjectRoot string `json:"projectRoot"`

	// ManifestPath 组件清单文件路径（可选）
	// 如果为空，默认为 {projectRoot}/.analyzer/component-manifest.json
	ManifestPath string `json:"manifestPath,omitempty"`

	// ExcludePaths 要排除的 glob 模式（可选）
	ExcludePaths []string `json:"excludePaths,omitempty"`

	// DiffFilePath diff 文件路径（可选）
	// 指向一个 git diff 文件，可以通过 `git diff main...HEAD > changes.diff` 生成
	// 如果设置了此字段，将自动从 diff 文件解析变更文件，changedFiles 参数将被忽略
	DiffFilePath string `json:"diffFilePath,omitempty"`
}

// LoadManifest 加载组件清单
// 这是外部调用的辅助函数，允许外部预先加载 manifest
// 注意：加载后会自动将相对路径转换为绝对路径（相对于 manifest 文件所在目录的父目录，即项目根目录）
func LoadManifest(manifestPath string) (*ComponentManifest, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	var manifest ComponentManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	// 获取项目根目录（manifest 文件在 .analyzer 目录中，项目根目录是其父目录）
	projectRoot := filepath.Clean(filepath.Join(filepath.Dir(manifestPath), ".."))

	// 转换为绝对路径，并复制 Name 字段
	for name, comp := range manifest.Components {
		comp.Name = name
		if !filepath.IsAbs(comp.Path) {
			comp.Path = filepath.Join(projectRoot, comp.Path)
		}
		manifest.Components[name] = comp
	}

	for name, fn := range manifest.Functions {
		fn.Name = name
		if !filepath.IsAbs(fn.Path) {
			fn.Path = filepath.Join(projectRoot, fn.Path)
		}
		manifest.Functions[name] = fn
	}

	return &manifest, nil
}

// GetChangedFilesFromDiff 从 diff 文件中提取变更文件列表
// 这是一个独立的辅助函数，方便外部调用
func GetChangedFilesFromDiff(diffFilePath string) ([]string, error) {
	return parseDiffFile(diffFilePath)
}

// AnalyzeFromDiff 分析 Git diff 变更的影响
// 这是外部调用的主要入口点，内部集成所有必要的分析步骤
//
// 参数：
//   - config: 分析配置（必须包含 DiffFilePath 指定 diff 文件路径）
//
// 返回：
//   - 分析结果
//   - 错误（如果有）
//
// 示例：
//
//	// 1. 先通过 git 命令生成 diff 文件
//	//    git diff main...HEAD > changes.diff
//	//    git diff HEAD~1 > changes.diff
//	//
//	// 2. 调用分析 API
//	result, err := mr_component_impact.AnalyzeFromDiff(&mr_component_impact.AnalyzeConfig{
//	    ProjectRoot:  "/path/to/project",
//	    DiffFilePath: "/path/to/changes.diff",
//	})
func AnalyzeFromDiff(config *AnalyzeConfig) (*AnalysisResult, error) {
	// 1. 从 diff 文件解析变更文件
	if config.DiffFilePath == "" {
		return nil, fmt.Errorf("必须指定 DiffFilePath 参数")
	}

	changedFiles, err := parseDiffFile(config.DiffFilePath)
	if err != nil {
		return nil, fmt.Errorf("解析 diff 文件失败: %w", err)
	}

	// 如果没有变更文件，返回空结果
	if len(changedFiles) == 0 {
		return &AnalysisResult{
			ChangedComponents:  make(map[string]*ComponentChangeInfo),
			ChangedFunctions:   make(map[string]*FunctionChangeInfo),
			ImpactedComponents: make(map[string][]ComponentImpact),
			OtherFiles:         []string{},
		}, nil
	}

	// 2. 设置默认值
	if config.ManifestPath == "" {
		config.ManifestPath = filepath.Join(config.ProjectRoot, ".analyzer", "component-manifest.json")
	}

	// 3. 加载 manifest
	manifest, err := LoadManifest(config.ManifestPath)
	if err != nil {
		return nil, err
	}

	// 4. 解析项目
	parsingResult, err := parseProject(config)
	if err != nil {
		return nil, err
	}

	// 5. 运行 component_deps 分析
	componentDeps := runComponentDepsAnalysis(manifest, parsingResult, config.ProjectRoot, config.ManifestPath)

	// 6. 运行 export_call 分析
	exportCallResult := runExportCallAnalysis(manifest, parsingResult, config.ProjectRoot, config.ManifestPath)

	// 7. 创建分析器并执行分析
	analyzer := NewAnalyzer(&AnalyzerConfig{
		Manifest:      manifest,
		FunctionPaths: extractFunctionPaths(manifest, config.ProjectRoot),
		ComponentDeps: componentDeps,
		ExportCall:    exportCallResult,
	})

	// 8. 规范化变更文件路径
	normalizedFiles := normalizeFilePaths(changedFiles, config.ProjectRoot)

	// 9. 执行分析
	result := analyzer.Analyze(normalizedFiles)

	return result, nil
}

// parseDiffFile 解析 diff 文件，提取变更的文件列表
// diff 文件格式示例（git diff 输出）：
// diff --git a/src/components/Button/Button.tsx b/src/components/Button/Button.tsx
// index 123..456 789
// --- a/src/components/Button/Button.tsx
// +++ b/src/components/Button/Button.tsx
func parseDiffFile(diffFilePath string) ([]string, error) {
	file, err := os.Open(diffFilePath)
	if err != nil {
		return nil, fmt.Errorf("打开 diff 文件失败: %w", err)
	}
	defer file.Close()

	// Git diff 文件中，文件路径通常出现在 "diff --git" 行中
	// 格式：diff --git a/path/to/file b/path/to/file
	diffPattern := regexp.MustCompile(`^diff --git a/(.+?) b/`)

	seenFiles := make(map[string]bool)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		matches := diffPattern.FindStringSubmatch(line)
		if len(matches) > 1 {
			filePath := matches[1]
			if !seenFiles[filePath] {
				seenFiles[filePath] = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取 diff 文件失败: %w", err)
	}

	// 转换为切片
	result := make([]string, 0, len(seenFiles))
	for filePath := range seenFiles {
		result = append(result, filePath)
	}

	return result, nil
}

// =============================================================================
// 内部辅助函数
// =============================================================================

// parseProject 解析项目
func parseProject(config *AnalyzeConfig) (*projectParser.ProjectParserResult, error) {
	// 设置默认排除路径
	excludePaths := config.ExcludePaths
	if len(excludePaths) == 0 {
		excludePaths = []string{"**/node_modules/**", "**/dist/**", "**/build/**"}
	}

	// 创建项目解析配置
	parserConfig := projectParser.NewProjectParserConfig(
		config.ProjectRoot,
		excludePaths,
		false, // isMonorepo
		nil,   // targetExtensions
	)

	// 创建并执行项目解析
	parsingResult := projectParser.NewProjectParserResult(parserConfig)
	parsingResult.ProjectParser()

	return parsingResult, nil
}

// runComponentDepsAnalysis 运行组件依赖分析
func runComponentDepsAnalysis(manifest *ComponentManifest, parsingResult *projectParser.ProjectParserResult, projectRoot string, manifestPath string) *component_deps.ComponentDepsResult {
	analyzer := &component_deps.ComponentDepsAnalyzer{}

	params := map[string]string{
		"manifest": manifestPath,
	}

	_ = analyzer.Configure(params)

	ctx := &projectanalyzer.ProjectContext{
		ProjectRoot:   projectRoot,
		ParsingResult: parsingResult,
		Exclude:       []string{},
		IsMonorepo:    false,
	}

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		return createEmptyComponentDepsResult(manifest)
	}

	componentDepsResult, ok := result.(*component_deps.ComponentDepsResult)
	if !ok {
		return createEmptyComponentDepsResult(manifest)
	}

	return componentDepsResult
}

// runExportCallAnalysis 运行导出引用分析
func runExportCallAnalysis(manifest *ComponentManifest, parsingResult *projectParser.ProjectParserResult, projectRoot string, manifestPath string) *export_call.ExportCallResult {
	analyzer := &export_call.ExportCallAnalyzer{}

	params := map[string]string{
		"manifest": manifestPath,
	}

	_ = analyzer.Configure(params)

	ctx := &projectanalyzer.ProjectContext{
		ProjectRoot:   projectRoot,
		ParsingResult: parsingResult,
		Exclude:       []string{},
		IsMonorepo:    false,
	}

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		return createEmptyExportCallResult()
	}

	exportCallResult, ok := result.(*export_call.ExportCallResult)
	if !ok {
		return createEmptyExportCallResult()
	}

	return exportCallResult
}

// extractFunctionPaths 从 manifest 提取函数路径列表
func extractFunctionPaths(manifest *ComponentManifest, projectRoot string) []string {
	paths := make([]string, 0, len(manifest.Functions))
	for _, fn := range manifest.Functions {
		if filepath.IsAbs(fn.Path) {
			paths = append(paths, fn.Path)
		} else {
			paths = append(paths, filepath.Join(projectRoot, fn.Path))
		}
	}
	return paths
}

// normalizeFilePaths 规范化文件路径
func normalizeFilePaths(files []string, projectRoot string) []string {
	normalized := make([]string, len(files))
	for i, file := range files {
		if filepath.IsAbs(file) {
			normalized[i] = file
		} else {
			normalized[i] = filepath.Join(projectRoot, file)
		}
	}
	return normalized
}

// createEmptyComponentDepsResult 创建空的组件依赖结果
func createEmptyComponentDepsResult(manifest *ComponentManifest) *component_deps.ComponentDepsResult {
	components := make(map[string]component_deps.ComponentInfo)
	for name, comp := range manifest.Components {
		components[name] = component_deps.ComponentInfo{
			Name:          name,
			Path:          comp.Path,
			Dependencies:  []projectParser.ImportDeclarationResult{},
			ComponentDeps: []component_deps.ComponentDep{},
		}
	}

	return &component_deps.ComponentDepsResult{
		Meta:       component_deps.Meta{ComponentCount: len(manifest.Components)},
		Components: components,
	}
}

// createEmptyExportCallResult 创建空的导出引用结果
func createEmptyExportCallResult() *export_call.ExportCallResult {
	return &export_call.ExportCallResult{
		ModuleExports: []export_call.ModuleExportRecord{},
	}
}
