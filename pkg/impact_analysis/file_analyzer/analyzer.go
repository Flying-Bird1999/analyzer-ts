// Package file_analyzer 提供文件级影响分析功能。
// 这是通用的能力，适用于所有前端项目，不依赖 component-manifest.json。
//
// 核心设计：基于符号级别的变更进行影响传播
// 输入：symbol_analysis 的输出（哪些符号被修改）
// 输出：哪些文件受影响（基于符号导入关系）
package file_analyzer

import (
	"time"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/Flying-Bird1999/analyzer-ts/pkg/symbol_analysis"
)

// =============================================================================
// 文件级分析器
// =============================================================================

// Analyzer 文件级影响分析器
// 通用的前端项目影响分析能力，不依赖 component-manifest.json
type Analyzer struct {
	parsingResult *projectParser.ProjectParserResult
}

// NewAnalyzer 创建文件分析器
func NewAnalyzer(parsingResult *projectParser.ProjectParserResult) *Analyzer {
	return &Analyzer{
		parsingResult: parsingResult,
	}
}

// Input 输入参数
type Input struct {
	// ChangedSymbols 来自 symbol_analysis 的分析结果
	// 表示哪些符号被修改了（仅符号文件）
	ChangedSymbols []ChangedSymbol

	// ChangedNonSymbolFiles 非符号文件的变更列表（如 CSS、图片等）
	// 对于这些文件，任何导入它们的文件都被视为受影响
	ChangedNonSymbolFiles []string
}

// ChangedSymbol 被修改的符号
type ChangedSymbol struct {
	Name       string                     // 符号名称，如 "ButtonProps", "Button"
	FilePath   string                     // 所属文件路径
	ExportType symbol_analysis.ExportType // 导出类型
}

// Result 文件级分析结果
type Result struct {
	Meta             FileAnalysisMeta   `json:"meta"`
	Changes          []FileChangeInfo    `json:"changes"`          // 直接变更的文件
	Impact           []FileImpactInfo     `json:"impact"`           // 受影响的文件
	DependencyGraph *FileDependencyGraph `json:"dependencyGraph"` // 文件依赖图（供组件级分析使用）
}

// Analyze 执行文件级影响分析
// 返回：受影响的文件列表
//
// 核心逻辑：
// 1. 接收符号级别的变更（来自 symbol_analysis）
// 2. 接收非符号文件的变更（如 CSS、图片等）
// 3. 对于符号文件：基于符号的 import/export 关系传播影响
// 4. 对于非符号文件：任何导入它们的文件都被视为受影响
func (a *Analyzer) Analyze(input *Input) (*Result, error) {
	// 构建符号影响传播器
	propagator := NewSymbolPropagator(a.parsingResult)

	// 执行符号级传播
	impactedFiles := propagator.Propagate(input.ChangedSymbols, input.ChangedNonSymbolFiles)

	// 构建文件依赖图（供组件级分析使用）
	graphBuilder := NewGraphBuilder(a.parsingResult)
	depGraph := graphBuilder.BuildFileDependencyGraph()

	// 构建结果
	return a.buildResult(input, impactedFiles, depGraph), nil
}

// buildResult 构建分析结果
func (a *Analyzer) buildResult(
	input *Input,
	impactedFiles *ImpactedFiles,
	depGraph *FileDependencyGraph,
) *Result {
	result := &Result{
		Meta: FileAnalysisMeta{
			AnalyzedAt:       time.Now(),
			TotalFileCount:   len(a.parsingResult.Js_Data),
			ChangedFileCount: countUniqueFiles(input.ChangedSymbols),
			ImpactFileCount:  len(impactedFiles.Indirect),
		},
		Changes:          buildFileChangeInfos(input.ChangedSymbols),
		Impact:           buildFileImpactInfos(impactedFiles),
		DependencyGraph: depGraph,
	}

	return result
}

// countUniqueFiles 统计符号变更涉及的唯一文件数
func countUniqueFiles(symbols []ChangedSymbol) int {
	seen := make(map[string]bool)
	for _, sym := range symbols {
		seen[sym.FilePath] = true
	}
	return len(seen)
}

// buildFileChangeInfos 构建文件变更信息列表
func buildFileChangeInfos(symbols []ChangedSymbol) []FileChangeInfo {
	// 按文件分组
	fileSymbols := make(map[string][]ChangedSymbol)
	for _, sym := range symbols {
		fileSymbols[sym.FilePath] = append(fileSymbols[sym.FilePath], sym)
	}

	infos := make([]FileChangeInfo, 0, len(fileSymbols))
	for filePath, syms := range fileSymbols {
		infos = append(infos, FileChangeInfo{
			Path:        filePath,
			ChangeType:  "modified", // TODO: 根据实际情况判断
			SymbolCount: len(syms),
		})
	}
	return infos
}

// buildFileImpactInfos 构建文件影响信息列表
func buildFileImpactInfos(impactedFiles *ImpactedFiles) []FileImpactInfo {
	infos := make([]FileImpactInfo, 0, len(impactedFiles.Indirect))

	for _, impact := range impactedFiles.Indirect {
		infos = append(infos, FileImpactInfo{
			Path:        impact.FilePath,
			ImpactLevel: impact.ImpactLevel,
			ChangePaths: impact.ChangePaths,
			SymbolCount: impact.SymbolCount,
		})
	}

	return infos
}

// =============================================================================
// 文件变更信息
// =============================================================================

// FileChangeInfo 文件变更信息（用于输出）
type FileChangeInfo struct {
	Path        string `json:"path"`
	ChangeType  string `json:"changeType"`
	SymbolCount int    `json:"symbolCount"` // 该文件中变更的符号数量
}

// FileImpactInfo 文件影响信息（用于输出）
type FileImpactInfo struct {
	Path        string                      `json:"path"`
	ImpactLevel int                         `json:"impactLevel"`
	ChangePaths []string                    `json:"changePaths"`
	SymbolCount int                         `json:"symbolCount"` // 影响的符号数量
}

// =============================================================================
// 文件分析元数据
// =============================================================================

// FileAnalysisMeta 文件分析元数据
type FileAnalysisMeta struct {
	AnalyzedAt       time.Time `json:"analyzedAt"`
	TotalFileCount   int       `json:"totalFileCount"`
	ChangedFileCount int       `json:"changedFileCount"`
	ImpactFileCount  int       `json:"impactFileCount"`
}

// GetImpactedFiles 获取所有受影响的文件路径（去重）
func (r *Result) GetImpactedFiles() []string {
	seen := make(map[string]bool)
	paths := make([]string, 0)

	for _, change := range r.Changes {
		if !seen[change.Path] {
			seen[change.Path] = true
			paths = append(paths, change.Path)
		}
	}
	for _, impact := range r.Impact {
		if !seen[impact.Path] {
			seen[impact.Path] = true
			paths = append(paths, impact.Path)
		}
	}
	return paths
}

// GetDirectChangedFiles 获取直接变更的文件路径
func (r *Result) GetDirectChangedFiles() []string {
	paths := make([]string, len(r.Changes))
	for i, change := range r.Changes {
		paths[i] = change.Path
	}
	return paths
}
