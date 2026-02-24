// Package component_analyzer 提供组件级影响分析功能。
// 这是组件库专用能力，基于 file_analyzer 的结果进行组件映射。
package component_analyzer

import (
	"time"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/Flying-Bird1999/analyzer-ts/pkg/impact_analysis"
)

// =============================================================================
// 组件级分析器
// =============================================================================

// Analyzer 组件级影响分析器
// 基于文件级分析结果，将影响映射到组件级别
type Analyzer struct {
	mapper        *ComponentMapper
	propagator    *Propagator
	parsingResult *projectParser.ProjectParserResult
}

// NewAnalyzer 创建组件分析器
func NewAnalyzer(
	manifest *impact_analysis.ComponentManifest,
	parsingResult *projectParser.ProjectParserResult,
	maxDepth int,
) *Analyzer {
	mapper := NewComponentMapper(manifest)

	return &Analyzer{
		mapper:        mapper,
		propagator:    nil, // 将在分析时创建
		parsingResult: parsingResult,
	}
}

// Input 输入参数
type Input struct {
	// FileResult 文件级分析结果（必需）
	FileResult *FileAnalysisResultProxy

	// SymbolChanges 可选：符号级变更信息
	SymbolChanges []SymbolChangeProxy
}

// Result 组件级分析结果
type Result struct {
	Meta    ComponentAnalysisMeta     `json:"meta"`
	Changes []ComponentChange         `json:"changes"` // 直接变更的组件
	Impact  []ComponentImpactInfo     `json:"impact"`  // 受影响的组件
	Paths   []ComponentImpactPathInfo `json:"paths"`   // 影响路径
}

// Analyze 执行组件级影响分析
// 返回：受影响的组件列表 + 影响路径
func (a *Analyzer) Analyze(input *Input) (*Result, error) {
	// 步骤 1: 构建组件依赖图
	fileGraphProxy := NewFileDependencyGraphProxy(
		input.FileResult.DepGraph,
		input.FileResult.RevDepGraph,
		input.FileResult.ExternalDeps,
	)
	componentGraph := a.mapper.BuildComponentDependencyGraph(fileGraphProxy, a.parsingResult)

	// 创建传播器
	a.propagator = NewPropagator(componentGraph, 10) // 默认深度 10

	// 步骤 2: 从文件级影响结果中提取变更组件
	changedComponents := a.extractChangedComponents(input)

	// 步骤 3: 传播影响
	impactedComponents := a.propagator.Propagate(changedComponents)

	// 步骤 4: 构建结果
	return a.buildResult(input, impactedComponents, componentGraph), nil
}

// extractChangedComponents 从文件级结果中提取变更组件
// 注意：只从直接变更的文件（Changes）中提取，不包括间接受影响的文件
func (a *Analyzer) extractChangedComponents(input *Input) []string {
	seen := make(map[string]bool)
	components := make([]string, 0)

	// 只从直接变更的文件中提取组件
	for _, fileChange := range input.FileResult.Changes {
		compName := a.mapper.MapFileToComponent(fileChange.Path)
		if compName != "" && !seen[compName] {
			seen[compName] = true
			components = append(components, compName)
		}
	}

	return components
}

// buildResult 构建分析结果
func (a *Analyzer) buildResult(
	input *Input,
	impactedComponents *ImpactedComponents,
	graph *ComponentDependencyGraph,
) *Result {
	result := &Result{
		Meta: ComponentAnalysisMeta{
			AnalyzedAt:            time.Now(),
			TotalComponentCount:   len(a.mapper.componentManifest.Components),
			ChangedComponentCount: len(impactedComponents.Direct),
			ImpactComponentCount:  len(impactedComponents.Indirect),
		},
		Changes: a.buildComponentChanges(impactedComponents.Direct, input.FileResult),
		Impact:  a.buildComponentImpacts(impactedComponents),
		Paths:   a.buildComponentImpactPaths(impactedComponents.ImpactPaths),
	}

	return result
}

// buildComponentChanges 构建组件变更列表
func (a *Analyzer) buildComponentChanges(
	direct map[string]*ComponentImpact,
	fileResult *FileAnalysisResultProxy,
) []ComponentChange {
	changes := make([]ComponentChange, 0, len(direct))

	for _, impact := range direct {
		// 获取该组件的变更文件
		changedFiles := a.getComponentChangedFiles(impact.ComponentName, fileResult)

		changes = append(changes, ComponentChange{
			Name:         impact.ComponentName,
			Action:       "modified", // TODO: 根据实际情况判断
			ChangedFiles: changedFiles,
			SymbolCount:  0, // TODO: 从符号变更中统计
		})
	}

	return changes
}

// getComponentChangedFiles 获取组件的变更文件列表
func (a *Analyzer) getComponentChangedFiles(
	componentName string,
	fileResult *FileAnalysisResultProxy,
) []string {
	comp := a.mapper.GetComponentByName(componentName)
	if comp == nil {
		return []string{}
	}

	componentDir := comp.Path
	files := make([]string, 0)

	// 从直接变更的文件中筛选
	for _, change := range fileResult.Changes {
		if len(change.Path) >= len(componentDir) && change.Path[:len(componentDir)] == componentDir {
			files = append(files, change.Path)
		}
	}

	return files
}

// buildComponentImpacts 构建组件影响列表
func (a *Analyzer) buildComponentImpacts(
	impactedComponents *ImpactedComponents,
) []ComponentImpactInfo {
	impacts := make([]ComponentImpactInfo, 0,
		len(impactedComponents.Direct)+len(impactedComponents.Indirect))

	// 添加直接变更的组件
	for _, impact := range impactedComponents.Direct {
		impacts = append(impacts, ComponentImpactInfo{
			Name:        impact.ComponentName,
			ImpactLevel: impact.ImpactLevel,
			ChangePaths: impact.ChangePaths,
			SymbolCount: 0, // TODO: 从符号变更中统计
		})
	}

	// 添加间接受影响的组件
	for _, impact := range impactedComponents.Indirect {
		impacts = append(impacts, ComponentImpactInfo{
			Name:        impact.ComponentName,
			ImpactLevel: impact.ImpactLevel,
			ChangePaths: impact.ChangePaths,
			SymbolCount: 0, // TODO: 从符号变更中统计
		})
	}

	return impacts
}

// buildComponentImpactPaths 构建组件影响路径列表
func (a *Analyzer) buildComponentImpactPaths(paths []ComponentImpactPath) []ComponentImpactPathInfo {
	result := make([]ComponentImpactPathInfo, len(paths))
	for i, path := range paths {
		result[i] = ComponentImpactPathInfo{
			SourceComponent: path.SourceComponent,
			TargetComponent: path.TargetComponent,
			Path:            path.Path,
		}
	}
	return result
}

// =============================================================================
// 代理类型（用于解耦）
// =============================================================================

// FileAnalysisResultProxy 文件分析结果代理
type FileAnalysisResultProxy struct {
	Changes      []FileChangeInfoProxy
	Impact       []FileImpactInfoProxy
	DepGraph     map[string][]string
	RevDepGraph  map[string][]string
	ExternalDeps map[string][]string
}

// FileChangeInfoProxy 文件变更信息代理
type FileChangeInfoProxy struct {
	Path        string
	ChangeType  impact_analysis.ChangeType
	SymbolCount int
}

// FileImpactInfoProxy 文件影响信息代理
type FileImpactInfoProxy struct {
	Path        string
	ImpactLevel impact_analysis.ImpactLevel
	ChangePaths []string
}

// SymbolChangeProxy 符号变更代理
type SymbolChangeProxy struct {
	Name       string
	Kind       string
	FilePath   string
	IsExported bool
	ExportType impact_analysis.ExportType
}

// =============================================================================
// 结果类型定义
// =============================================================================

// ComponentAnalysisMeta 组件分析元数据
type ComponentAnalysisMeta struct {
	AnalyzedAt            time.Time `json:"analyzedAt"`
	TotalComponentCount   int       `json:"totalComponentCount"`
	ChangedComponentCount int       `json:"changedComponentCount"`
	ImpactComponentCount  int       `json:"impactComponentCount"`
}

// ComponentChange 组件变更信息
type ComponentChange struct {
	Name         string   `json:"name"`         // 组件名称
	Action       string   `json:"action"`       // 变更类型: modified/added/deleted
	ChangedFiles []string `json:"changedFiles"` // 变更的文件列表
	SymbolCount  int      `json:"symbolCount"`  // 变更的符号数量
}

// ComponentImpactInfo 组件影响信息
type ComponentImpactInfo struct {
	Name        string                      `json:"name"`        // 组件名称
	ImpactLevel impact_analysis.ImpactLevel `json:"impactLevel"` // 影响层级
	ChangePaths []string                    `json:"changePaths"` // 影响路径
	SymbolCount int                         `json:"symbolCount"` // 影响的符号数量
}

// ComponentImpactPathInfo 组件影响路径信息
type ComponentImpactPathInfo struct {
	SourceComponent string   `json:"sourceComponent"` // 源头组件
	TargetComponent string   `json:"targetComponent"` // 目标组件
	Path            []string `json:"path"`            // 传播路径
}

// =============================================================================
// Result 方法扩展
// =============================================================================

// GetImpactedComponents 获取所有受影响的组件名称（去重）
func (r *Result) GetImpactedComponents() []string {
	seen := make(map[string]bool)
	components := make([]string, 0)

	for _, change := range r.Changes {
		if !seen[change.Name] {
			seen[change.Name] = true
			components = append(components, change.Name)
		}
	}
	for _, impact := range r.Impact {
		if !seen[impact.Name] {
			seen[impact.Name] = true
			components = append(components, impact.Name)
		}
	}
	return components
}

// GetDirectChangedComponents 获取直接变更的组件名称
func (r *Result) GetDirectChangedComponents() []string {
	components := make([]string, len(r.Changes))
	for i, change := range r.Changes {
		components[i] = change.Name
	}
	return components
}
