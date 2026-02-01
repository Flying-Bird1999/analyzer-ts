// Package impact_analysis 提供符号级影响分析功能。
// 它基于 symbol_analysis 的输出（导出符号变更）和组件依赖图，分析代码变更的影响范围。
package impact_analysis

import (
	"fmt"
	"time"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

// =============================================================================
// 符号级影响分析器
// =============================================================================

// Analyzer 符号级影响分析器
// 负责协调整个影响分析流程
type Analyzer struct {
	project          *tsmorphgo.Project
	parsingResult    *projectParser.ProjectParserResult
	componentManifest *ComponentManifest
	maxDepth          int
}

// NewAnalyzer 创建影响分析器
func NewAnalyzer(
	project *tsmorphgo.Project,
	parsingResult *projectParser.ProjectParserResult,
	manifest *ComponentManifest,
	maxDepth int,
) *Analyzer {
	if maxDepth <= 0 {
		maxDepth = 10 // 默认最大深度
	}

	return &Analyzer{
		project:          project,
		parsingResult:    parsingResult,
		componentManifest: manifest,
		maxDepth:         maxDepth,
	}
}

// Analyze 执行符号级影响分析
// 输入：符号变更列表、组件依赖图、反向依赖图
// 输出：完整的影响分析结果
func (a *Analyzer) Analyze(
	symbolChanges []SymbolChange,
	depGraph map[string][]string,
	revDepGraph map[string][]string,
) (*AnalysisResult, error) {
	// 步骤 1: 创建 Matcher 并匹配符号到组件
	matcher := NewMatcher(a.project, a.parsingResult, a.componentManifest)
	componentChanges := matcher.MatchSymbolsToComponents(symbolChanges)

	// 步骤 2: 构建符号依赖映射
	depMap := matcher.BuildSymbolDependencyMap()

	// 步骤 3: 创建 Propagator 并执行影响传播
	propagator := NewPropagator(
		depMap,
		depGraph,
		revDepGraph,
		symbolChanges,
		componentChanges,
		a.maxDepth,
	)
	impacts := propagator.Propagate()

	// 步骤 4: 创建 Assessor 并评估风险
	assessor := NewAssessor()
	assessment := assessor.Assess(impacts)

	// 步骤 5: 构建最终结果
	result := a.buildResult(
		symbolChanges,
		componentChanges,
		impacts,
		depMap,
		assessment,
	)

	return result, nil
}

// buildResult 构建最终的分析结果
func (a *Analyzer) buildResult(
	symbolChanges []SymbolChange,
	componentChanges map[string][]SymbolChange,
	impacts map[string]*ComponentImpact,
	depMap *SymbolDependencyMap,
	assessment *RiskAssessmentResult,
) *AnalysisResult {
	result := &AnalysisResult{
		Meta: ImpactMeta{
			AnalyzedAt:      time.Now().Format(time.RFC3339),
			ComponentCount:   len(a.componentManifest.Components),
			ChangedFileCount: countChangedFiles(symbolChanges),
			ChangeSource:     "symbol_analysis",
			SymbolCount:      len(symbolChanges),
		},
		Changes:        a.buildComponentChanges(componentChanges, impacts),
		Impact:         a.buildImpactComponents(impacts, assessment),
		SymbolChanges:  a.buildSymbolImpactChanges(symbolChanges, impacts, depMap),
		RiskAssessment: a.buildRiskAssessment(assessment),
		Recommendations: a.generateRecommendations(impacts, assessment),
	}

	return result
}

// buildComponentChanges 构建组件变更列表
func (a *Analyzer) buildComponentChanges(
	componentChanges map[string][]SymbolChange,
	impacts map[string]*ComponentImpact,
) []ComponentChange {
	changes := make([]ComponentChange, 0, len(componentChanges))

	for compName, symbols := range componentChanges {
		change := ComponentChange{
			Name:         compName,
			Action:       determineComponentAction(symbols),
			ChangedFiles: extractChangedFiles(symbols),
			SymbolCount:  len(symbols),
		}
		changes = append(changes, change)
	}

	return changes
}

// buildImpactComponents 构建受影响组件列表
func (a *Analyzer) buildImpactComponents(
	impacts map[string]*ComponentImpact,
	assessment *RiskAssessmentResult,
) []ImpactComponent {
	components := make([]ImpactComponent, 0, len(impacts))

	for compName, impact := range impacts {
		riskLevel, exists := assessment.ComponentRisks[compName]
		if !exists {
			riskLevel = "low"
		}

		comp := ImpactComponent{
			Name:        compName,
			ImpactLevel: impact.ImpactLevel,
			RiskLevel:   riskLevel,
			ChangePaths: impact.ChangePaths,
			SymbolCount: len(impact.AffectedSymbols),
		}
		components = append(components, comp)
	}

	return components
}

// buildSymbolImpactChanges 构建符号级影响变更列表
func (a *Analyzer) buildSymbolImpactChanges(
	symbolChanges []SymbolChange,
	impacts map[string]*ComponentImpact,
	depMap *SymbolDependencyMap,
) []SymbolImpactChange {
	changes := make([]SymbolImpactChange, 0, len(symbolChanges))

	for _, symbol := range symbolChanges {
		// 确定影响类型
		impactType := classifySymbolImpact(symbol)

		// 查找受影响的组件
		affectedComponents := a.findAffectedComponents(symbol, impacts, depMap)

		change := SymbolImpactChange{
			Symbol:             symbol,
			ComponentName:      symbol.ComponentName,
			ImpactType:         impactType,
			AffectedComponents: affectedComponents,
		}
		changes = append(changes, change)
	}

	return changes
}

// findAffectedComponents 查找受符号变更影响的组件
func (a *Analyzer) findAffectedComponents(
	symbol SymbolChange,
	impacts map[string]*ComponentImpact,
	depMap *SymbolDependencyMap,
) []string {
	affected := make([]string, 0)

	// 遍历所有受影响的组件
	for compName, impact := range impacts {
		// 检查该组件是否受此符号影响
		if impact.AffectedSymbols[symbol.Name] {
			if compName != symbol.ComponentName {
				affected = append(affected, compName)
			}
		}
	}

	return affected
}

// buildRiskAssessment 构建风险评估
func (a *Analyzer) buildRiskAssessment(assessment *RiskAssessmentResult) RiskAssessment {
	return RiskAssessment{
		OverallRisk:    assessment.OverallRisk,
		BreakingChange: assessment.BreakingChange,
		InternalChange: assessment.InternalChange,
		AdditiveChange: assessment.AdditiveChange,
	}
}

// generateRecommendations 生成建议
func (a *Analyzer) generateRecommendations(
	impacts map[string]*ComponentImpact,
	assessment *RiskAssessmentResult,
) []Recommendation {
	recommendations := make([]Recommendation, 0)

	// 根据风险等级生成建议
	for compName, risk := range assessment.ComponentRisks {
		switch risk {
		case "critical":
			recommendations = append(recommendations, Recommendation{
				Type:        "review",
				Priority:    "critical",
				Description: fmt.Sprintf("组件 %s 存在严重风险，建议立即进行代码审查", compName),
				Target:      compName,
			})
			recommendations = append(recommendations, Recommendation{
				Type:        "test",
				Priority:    "critical",
				Description: fmt.Sprintf("组件 %s 需要补充完整的单元测试和集成测试", compName),
				Target:      compName,
			})

		case "high":
			recommendations = append(recommendations, Recommendation{
				Type:        "review",
				Priority:    "high",
				Description: fmt.Sprintf("组件 %s 存在较高风险，建议进行代码审查", compName),
				Target:      compName,
			})
			recommendations = append(recommendations, Recommendation{
				Type:        "test",
				Priority:    "high",
				Description: fmt.Sprintf("组件 %s 建议补充测试用例", compName),
				Target:      compName,
			})

		case "medium":
			recommendations = append(recommendations, Recommendation{
				Type:        "test",
				Priority:    "medium",
				Description: fmt.Sprintf("组件 %s 建议进行基本的测试验证", compName),
				Target:      compName,
			})
		}
	}

	// 根据影响类型生成建议
	if assessment.BreakingChange > 0 {
		recommendations = append(recommendations, Recommendation{
			Type:        "document",
			Priority:    "high",
			Description: fmt.Sprintf("检测到 %d 个破坏性变更，请更新相关文档和版本说明", assessment.BreakingChange),
		})
	}

	return recommendations
}

// =============================================================================
// 辅助函数
// =============================================================================

// countChangedFiles 统计变更文件数量
func countChangedFiles(symbolChanges []SymbolChange) int {
	fileSet := make(map[string]bool)
	for _, change := range symbolChanges {
		fileSet[change.FilePath] = true
	}
	return len(fileSet)
}

// determineComponentAction 确定组件的变更类型
func determineComponentAction(symbols []SymbolChange) string {
	hasModified := false
	hasAdded := false
	hasDeleted := false

	for _, sym := range symbols {
		switch sym.ChangeType {
		case ChangeTypeModified:
			hasModified = true
		case ChangeTypeAdded:
			hasAdded = true
		case ChangeTypeDeleted:
			hasDeleted = true
		}
	}

	if hasDeleted {
		return "deleted"
	}
	if hasModified {
		return "modified"
	}
	if hasAdded {
		return "added"
	}
	return "modified"
}

// extractChangedFiles 从符号变更中提取变更文件列表
func extractChangedFiles(symbols []SymbolChange) []string {
	fileSet := make(map[string]bool)
	for _, sym := range symbols {
		fileSet[sym.FilePath] = true
	}

	files := make([]string, 0, len(fileSet))
	for file := range fileSet {
		files = append(files, file)
	}
	return files
}
