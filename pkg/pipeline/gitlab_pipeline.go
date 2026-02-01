// Package pipeline 提供代码分析管道协调功能。
package pipeline

import (
	"fmt"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/Flying-Bird1999/analyzer-ts/pkg/impact_analysis"
)

// =============================================================================
// GitLab Pipeline 工厂函数
// =============================================================================

// GitLabPipelineConfig GitLab 管道配置
type GitLabPipelineConfig struct {
	// Diff 配置
	DiffSource  DiffSourceType
	DiffFile    string
	DiffSHA     string
	ProjectRoot string
	ProjectID   int
	MRIID       int

	// 组件依赖配置
	ManifestPath  string
	DepsFile      string

	// 影响分析配置
	MaxDepth int

	// GitLab 客户端
	Client GitLabClient
}

// NewGitLabPipeline 创建完整的 GitLab 分析管道
// 包含：Diff解析 → 符号分析 → 组件依赖分析 → 影响分析
func NewGitLabPipeline(config *GitLabPipelineConfig) *AnalysisPipeline {
	pipe := NewPipeline("GitLab Analysis Pipeline")

	// 阶段 1: Diff 解析
	pipe.AddStage(NewDiffParserStage(
		config.Client,
		config.DiffSource,
		config.DiffFile,
		config.DiffSHA,
		config.ProjectRoot,
		config.ProjectID,
		config.MRIID,
	))

	// 阶段 2: 符号分析
	pipe.AddStage(NewSymbolAnalysisStage())

	// 阶段 3: 组件依赖分析
	pipe.AddStage(NewComponentDepsStage(
		config.ManifestPath,
		config.DepsFile,
	))

	// 阶段 4: 影响分析
	pipe.AddStage(NewImpactAnalysisStage(
		config.MaxDepth,
	))

	return pipe
}

// NewGitLabPipelineWithoutDeps 创建 GitLab 管道（不包含组件依赖分析）
// 用于已预生成依赖数据的场景
func NewGitLabPipelineWithoutDeps(config *GitLabPipelineConfig) *AnalysisPipeline {
	pipe := NewPipeline("GitLab Analysis Pipeline (No Deps Analysis)")

	// 阶段 1: Diff 解析
	pipe.AddStage(NewDiffParserStage(
		config.Client,
		config.DiffSource,
		config.DiffFile,
		config.DiffSHA,
		config.ProjectRoot,
		config.ProjectID,
		config.MRIID,
	))

	// 阶段 2: 符号分析
	pipe.AddStage(NewSymbolAnalysisStage())

	// 阶段 3: 影响分析（使用预加载的依赖数据）
	pipe.AddStage(NewImpactAnalysisStageWithPreloadedDeps(
		config.DepsFile,
		config.MaxDepth,
	))

	return pipe
}

// =============================================================================
// 组件依赖分析阶段
// =============================================================================

// ComponentDepsStage 组件依赖分析阶段
// 分析项目中的组件依赖关系
type ComponentDepsStage struct {
	manifestPath string
	depsFile     string
}

// NewComponentDepsStage 创建组件依赖分析阶段
func NewComponentDepsStage(manifestPath, depsFile string) *ComponentDepsStage {
	return &ComponentDepsStage{
		manifestPath: manifestPath,
		depsFile:     depsFile,
	}
}

// Name 返回阶段名称
func (s *ComponentDepsStage) Name() string {
	return "组件依赖分析"
}

// Execute 执行组件依赖分析
func (s *ComponentDepsStage) Execute(ctx *AnalysisContext) (interface{}, error) {
	// 如果提供了依赖文件，直接加载
	if s.depsFile != "" {
		fmt.Printf("  - 从文件加载依赖数据: %s\n", s.depsFile)
		return s.loadDependencyData(s.depsFile)
	}

	// 否则运行组件依赖分析
	fmt.Println("  - 运行组件依赖分析...")

	// 获取项目解析器配置
	if ctx.Project == nil {
		return nil, fmt.Errorf("project not initialized in context")
	}

	// 创建项目解析器配置
	parserConfig := projectParser.NewProjectParserConfig(
		ctx.ProjectRoot,
		[]string{}, // exclude patterns
		false,     // isMonorepo
		[]string{},// strip paths
	)

	// 解析项目
	fmt.Println("  - 解析项目 AST...")
	parsingResult := projectParser.NewProjectParserResult(parserConfig)
	parsingResult.ProjectParser()
	fmt.Printf("  - 发现 %d 个 JS/TS 文件\n", len(parsingResult.Js_Data))

	// 存储解析结果到上下文（供后续阶段使用）
	ctx.SetResult("projectParser", parsingResult)

	// 返回解析结果
	return parsingResult, nil
}

// Skip 判断是否跳过此阶段
func (s *ComponentDepsStage) Skip(ctx *AnalysisContext) bool {
	return false
}

// loadDependencyData 从文件加载依赖数据
func (s *ComponentDepsStage) loadDependencyData(filePath string) (*ComponentDepsResult, error) {
	// TODO: 实现从文件加载依赖数据的逻辑
	// 这里暂时返回占位符
	return &ComponentDepsResult{
		DepGraph:    make(map[string][]string),
		RevDepGraph: make(map[string][]string),
		Manifest:    nil,
		ParsingResult: nil,
	}, nil
}

// ComponentDepsResult 组件依赖分析结果
type ComponentDepsResult struct {
	DepGraph      map[string][]string          // 组件依赖图
	RevDepGraph   map[string][]string          // 反向依赖图
	Manifest      *impact_analysis.ComponentManifest // 组件清单
	ParsingResult *projectParser.ProjectParserResult // 项目解析结果
}

// =============================================================================
// 影响分析阶段
// =============================================================================

// ImpactAnalysisStage 影响分析阶段
// 使用符号分析和组件依赖分析的结果进行影响分析
type ImpactAnalysisStage struct {
	maxDepth int
}

// NewImpactAnalysisStage 创建影响分析阶段
func NewImpactAnalysisStage(maxDepth int) *ImpactAnalysisStage {
	if maxDepth <= 0 {
		maxDepth = 10
	}
	return &ImpactAnalysisStage{
		maxDepth: maxDepth,
	}
}

// Name 返回阶段名称
func (s *ImpactAnalysisStage) Name() string {
	return "影响分析"
}

// Execute 执行影响分析
func (s *ImpactAnalysisStage) Execute(ctx *AnalysisContext) (interface{}, error) {
	// 1. 获取符号分析结果
	symbolResults, exists := ctx.GetResult("符号分析")
	if !exists {
		return nil, fmt.Errorf("symbol analysis result not found")
	}

	// 2. 获取项目解析结果
	parsingResult, exists := ctx.GetResult("projectParser")
	if !exists {
		return nil, fmt.Errorf("project parser result not found")
	}

	parsedResult, ok := parsingResult.(*projectParser.ProjectParserResult)
	if !ok {
		return nil, fmt.Errorf("invalid project parser result type")
	}

	// 3. 获取组件清单
	manifest, err := s.loadComponentManifest(ctx)
	if err != nil {
		return nil, fmt.Errorf("load component manifest failed: %w", err)
	}

	// 4. 转换符号分析结果
	symbolChanges := s.convertToSymbolChanges(symbolResults)

	// 5. 构建组件依赖图
	depGraph, revDepGraph := s.buildDependencyGraphs(parsedResult, manifest)

	// 6. 创建影响分析器并执行分析
	analyzer := impact_analysis.NewAnalyzer(
		ctx.Project,
		parsedResult,
		manifest,
		s.maxDepth,
	)

	result, err := analyzer.Analyze(symbolChanges, depGraph, revDepGraph)
	if err != nil {
		return nil, fmt.Errorf("impact analysis failed: %w", err)
	}

	// 7. 打印摘要
	fmt.Printf("  - 发现 %d 个组件变更\n", len(result.Changes))
	fmt.Printf("  - 影响 %d 个组件\n", len(result.Impact))
	fmt.Printf("  - 整体风险: %s\n", result.RiskAssessment.OverallRisk)

	return result, nil
}

// Skip 判断是否跳过此阶段
func (s *ImpactAnalysisStage) Skip(ctx *AnalysisContext) bool {
	// 如果没有符号分析结果，跳过此阶段
	_, exists := ctx.GetResult("符号分析")
	return !exists
}

// loadComponentManifest 加载组件清单
func (s *ImpactAnalysisStage) loadComponentManifest(ctx *AnalysisContext) (*impact_analysis.ComponentManifest, error) {
	// TODO: 实现从配置路径加载组件清单的逻辑
	// 这里暂时返回一个空的清单
	return &impact_analysis.ComponentManifest{
		Meta: impact_analysis.ManifestMeta{
			Version:     "1.0",
			LibraryName: "default",
		},
		Components: []impact_analysis.Component{},
	}, nil
}

// convertToSymbolChanges 转换符号分析结果为 SymbolChange 列表
func (s *ImpactAnalysisStage) convertToSymbolChanges(symbolResults interface{}) []impact_analysis.SymbolChange {
	// 类型断言
	_, ok := symbolResults.(map[string]interface{})
	if !ok {
		return []impact_analysis.SymbolChange{}
	}

	changes := make([]impact_analysis.SymbolChange, 0)

	// TODO: 实现完整的转换逻辑
	// 这里需要将 symbol_analysis 的结果转换为 impact_analysis.SymbolChange 格式

	return changes
}

// buildDependencyGraphs 构建组件依赖图
func (s *ImpactAnalysisStage) buildDependencyGraphs(
	parsingResult *projectParser.ProjectParserResult,
	manifest *impact_analysis.ComponentManifest,
) (map[string][]string, map[string][]string) {
	depGraph := make(map[string][]string)
	revDepGraph := make(map[string][]string)

	// TODO: 实现完整的依赖图构建逻辑
	// 这里需要使用 projectParser 的 ImportDeclarations 来构建依赖关系

	return depGraph, revDepGraph
}

// =============================================================================
// 影响分析阶段（使用预加载的依赖数据）
// =============================================================================

// ImpactAnalysisStageWithPreloadedDeps 影响分析阶段（使用预加载数据）
// 用于已预生成依赖数据的场景
type ImpactAnalysisStageWithPreloadedDeps struct {
	depsFile string
	maxDepth int
}

// NewImpactAnalysisStageWithPreloadedDeps 创建影响分析阶段（使用预加载数据）
func NewImpactAnalysisStageWithPreloadedDeps(depsFile string, maxDepth int) *ImpactAnalysisStageWithPreloadedDeps {
	if maxDepth <= 0 {
		maxDepth = 10
	}
	return &ImpactAnalysisStageWithPreloadedDeps{
		depsFile: depsFile,
		maxDepth: maxDepth,
	}
}

// Name 返回阶段名称
func (s *ImpactAnalysisStageWithPreloadedDeps) Name() string {
	return "影响分析"
}

// Execute 执行影响分析（使用预加载数据）
func (s *ImpactAnalysisStageWithPreloadedDeps) Execute(ctx *AnalysisContext) (interface{}, error) {
	// 1. 加载预生成的依赖数据
	depsResult, err := s.loadDependencyData(s.depsFile)
	if err != nil {
		return nil, fmt.Errorf("load dependency data failed: %w", err)
	}

	// 2. 获取符号分析结果
	symbolResults, exists := ctx.GetResult("符号分析")
	if !exists {
		return nil, fmt.Errorf("symbol analysis result not found")
	}

	// 3. 转换符号分析结果
	symbolChanges := s.convertToSymbolChanges(symbolResults)

	// 4. 创建影响分析器并执行分析
	analyzer := impact_analysis.NewAnalyzer(
		ctx.Project,
		depsResult.ParsingResult,
		depsResult.Manifest,
		s.maxDepth,
	)

	result, err := analyzer.Analyze(
		symbolChanges,
		depsResult.DepGraph,
		depsResult.RevDepGraph,
	)
	if err != nil {
		return nil, fmt.Errorf("impact analysis failed: %w", err)
	}

	// 5. 打印摘要
	fmt.Printf("  - 发现 %d 个组件变更\n", len(result.Changes))
	fmt.Printf("  - 影响 %d 个组件\n", len(result.Impact))
	fmt.Printf("  - 整体风险: %s\n", result.RiskAssessment.OverallRisk)

	return result, nil
}

// Skip 判断是否跳过此阶段
func (s *ImpactAnalysisStageWithPreloadedDeps) Skip(ctx *AnalysisContext) bool {
	// 如果没有符号分析结果，跳过此阶段
	_, exists := ctx.GetResult("符号分析")
	return !exists
}

// loadDependencyData 加载依赖数据
func (s *ImpactAnalysisStageWithPreloadedDeps) loadDependencyData(filePath string) (*ComponentDepsResult, error) {
	// TODO: 实现从文件加载依赖数据的逻辑
	return &ComponentDepsResult{
		DepGraph:      make(map[string][]string),
		RevDepGraph:   make(map[string][]string),
		Manifest:      nil,
		ParsingResult: nil,
	}, nil
}

// convertToSymbolChanges 转换符号分析结果
func (s *ImpactAnalysisStageWithPreloadedDeps) convertToSymbolChanges(symbolResults interface{}) []impact_analysis.SymbolChange {
	// TODO: 实现转换逻辑
	return []impact_analysis.SymbolChange{}
}
