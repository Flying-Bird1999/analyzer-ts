// Package mr_component_impact 提供 MR 组件影响分析功能
package mr_component_impact

import (
	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/component_deps"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/export_call"
)

// =============================================================================
// MR 组件影响分析器
// =============================================================================

// Analyzer MR 组件影响分析器
// 协调文件分类、组件影响分析和函数影响分析
type Analyzer struct {
	classifier        *Classifier
	componentAnalyzer *ComponentImpactAnalyzer
	functionAnalyzer  *FunctionImpactAnalyzer
}

// AnalyzerConfig 分析器配置
type AnalyzerConfig struct {
	Manifest      *ComponentManifest
	FunctionPaths []string
	ComponentDeps *component_deps.ComponentDepsResult
	ExportCall    *export_call.ExportCallResult
}

// NewAnalyzer 创建 MR 组件影响分析器
func NewAnalyzer(config *AnalyzerConfig) *Analyzer {
	classifier := NewClassifier(config.Manifest, config.FunctionPaths)

	return &Analyzer{
		classifier:        classifier,
		componentAnalyzer: NewComponentImpactAnalyzer(config.ComponentDeps),
		functionAnalyzer:  NewFunctionImpactAnalyzer(config.ExportCall),
	}
}

// Analyze 执行 MR 组件影响分析
// 输入: changedFiles - 变更文件列表
// 输出: 分析结果
func (a *Analyzer) Analyze(changedFiles []string) *AnalysisResult {
	result := &AnalysisResult{
		ChangedComponents:  make(map[string]*ComponentChangeInfo),
		ChangedFunctions:   make(map[string]*FunctionChangeInfo),
		ImpactedComponents: make(map[string][]ComponentImpact),
		OtherFiles:         make([]string, 0),
	}

	for _, file := range changedFiles {
		category, name := a.classifier.ClassifyFile(file)

		switch category {
		case CategoryComponent:
			a.analyzeComponentChange(result, name, file)

		case CategoryFunctions:
			a.analyzeFunctionChange(result, name, file)

		default:
			result.OtherFiles = append(result.OtherFiles, file)
		}
	}

	return result
}

// analyzeComponentChange 分析组件变更
func (a *Analyzer) analyzeComponentChange(
	result *AnalysisResult,
	componentName string,
	filePath string,
) {
	// 记录变更组件
	if _, exists := result.ChangedComponents[componentName]; !exists {
		result.ChangedComponents[componentName] = &ComponentChangeInfo{
			Name:  componentName,
			Files: make([]string, 0),
		}
	}
	result.ChangedComponents[componentName].Files = append(
		result.ChangedComponents[componentName].Files,
		filePath,
	)

	// 分析影响
	impacts := a.componentAnalyzer.AnalyzeComponentChange(componentName)

	// 将变更组件本身也加入到受影响列表中（标记为直接变更）
	selfImpact := ComponentImpact{
		ChangeSource: componentName,
		Relation:     RelationDirect,
		Level:        0,
	}
	result.ImpactedComponents[componentName] = append(
		result.ImpactedComponents[componentName],
		selfImpact,
	)

	// 记录受影响的组件
	for _, impact := range impacts {
		result.ImpactedComponents[impact.Component] = append(
			result.ImpactedComponents[impact.Component],
			impact,
		)
	}
}

// analyzeFunctionChange 分析函数变更
func (a *Analyzer) analyzeFunctionChange(
	result *AnalysisResult,
	functionName string,
	filePath string,
) {
	// 记录变更函数
	if _, exists := result.ChangedFunctions[functionName]; !exists {
		result.ChangedFunctions[functionName] = &FunctionChangeInfo{
			Name:  functionName,
			Files: make([]string, 0),
		}
	}
	result.ChangedFunctions[functionName].Files = append(
		result.ChangedFunctions[functionName].Files,
		filePath,
	)

	// 分析影响
	impacts := a.functionAnalyzer.AnalyzeFunctionChange(filePath, functionName)

	// 记录受影响的组件
	for _, impact := range impacts {
		result.ImpactedComponents[impact.Component] = append(
			result.ImpactedComponents[impact.Component],
			impact,
		)
	}
}
