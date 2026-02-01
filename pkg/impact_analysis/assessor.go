// Package impact_analysis 提供符号级影响分析功能。
package impact_analysis

import (
	"math"
)

// =============================================================================
// 风险评估器
// =============================================================================

// Assessor 风险评估器
// 负责评估影响分析结果的风险等级
type Assessor struct {
	// 影响权重配置
	breakingWeight float64
	internalWeight float64
	additiveWeight float64

	// 层级衰减系数
	levelDecay float64
}

// NewAssessor 创建风险评估器
func NewAssessor() *Assessor {
	return &Assessor{
		breakingWeight: 10.0, // 破坏性变更权重最高
		internalWeight: 3.0,  // 内部变更权重较低
		additiveWeight: 1.0,  // 增强性变更权重最低
		levelDecay:     0.8,  // 每层传播衰减20%的影响
	}
}

// Assess 评估整体风险
func (a *Assessor) Assess(impacts map[string]*ComponentImpact) *RiskAssessmentResult {
	assessment := &RiskAssessmentResult{
		ImpactCount:  len(impacts),
		ComponentRisks: make(map[string]string),
	}

	breakingCount := 0
	internalCount := 0
	additiveCount := 0

	// 统计各类变更数量
	for _, impact := range impacts {
		switch impact.ImpactType {
		case ImpactTypeBreaking:
			breakingCount++
		case ImpactTypeInternal:
			internalCount++
		case ImpactTypeAdditive:
			additiveCount++
		}
	}

	assessment.BreakingChange = breakingCount
	assessment.InternalChange = internalCount
	assessment.AdditiveChange = additiveCount

	// 计算每个组件的风险等级
	for compName, impact := range impacts {
		riskLevel := a.assessComponentRisk(impact)
		assessment.ComponentRisks[compName] = riskLevel
	}

	// 计算整体风险等级
	assessment.OverallRisk = a.calculateOverallRisk(assessment)

	return assessment
}

// assessComponentRisk 评估单个组件的风险等级
func (a *Assessor) assessComponentRisk(impact *ComponentImpact) string {
	// 计算风险分数
	score := a.calculateRiskScore(impact)

	// 根据分数确定风险等级
	if score >= 8.0 {
		return "critical"
	}
	if score >= 5.0 {
		return "high"
	}
	if score >= 3.0 {
		return "medium"
	}
	return "low"
}

// calculateRiskScore 计算组件的风险分数
func (a *Assessor) calculateRiskScore(impact *ComponentImpact) float64 {
	score := 0.0

	// 基础分数：根据影响类型
	switch impact.ImpactType {
	case ImpactTypeBreaking:
		score = a.breakingWeight
	case ImpactTypeInternal:
		score = a.internalWeight
	case ImpactTypeAdditive:
		score = a.additiveWeight
	}

	// 层级衰减：影响层级越高，风险越低
	levelFactor := math.Pow(a.levelDecay, float64(impact.ImpactLevel))
	score *= levelFactor

	// 符号数量加成：受影响的符号越多，风险越高
	symbolCount := len(impact.AffectedSymbols)
	if symbolCount > 0 {
		symbolBonus := math.Log(float64(symbolCount + 1))
		score += symbolBonus
	}

	// 直接变更加成：组件本身有变更，风险更高
	if len(impact.ChangedSymbols) > 0 {
		score += 2.0
	}

	return score
}

// calculateOverallRisk 计算整体风险等级
func (a *Assessor) calculateOverallRisk(assessment *RiskAssessmentResult) string {
	// 统计各风险等级的组件数量
	criticalCount := 0
	highCount := 0
	mediumCount := 0
	lowCount := 0

	for _, risk := range assessment.ComponentRisks {
		switch risk {
		case "critical":
			criticalCount++
		case "high":
			highCount++
		case "medium":
			mediumCount++
		case "low":
			lowCount++
		}
	}

	// 如果有任何 critical 组件，整体风险为 critical
	if criticalCount > 0 {
		return "critical"
	}

	// 如果有多个 high 组件，整体风险为 critical
	if highCount >= 3 {
		return "critical"
	}

	// 如果有任何 high 组件，整体风险为 high
	if highCount > 0 {
		return "high"
	}

	// 如果有多个 medium 组件，整体风险为 high
	if mediumCount >= 5 {
		return "high"
	}

	// 如果有任何 medium 组件，整体风险为 medium
	if mediumCount > 0 {
		return "medium"
	}

	// 默认为 low
	return "low"
}

// =============================================================================
// 风险评估结果类型
// =============================================================================

// RiskAssessmentResult 风险评估结果
type RiskAssessmentResult struct {
	// OverallRisk 整体风险等级
	OverallRisk string

	// ImpactCount 受影响的组件总数
	ImpactCount int

	// BreakingChange 破坏性变更数量
	BreakingChange int

	// InternalChange 内部变更数量
	InternalChange int

	// AdditiveChange 增强性变更数量
	AdditiveChange int

	// ComponentRisks 每个组件的风险等级
	ComponentRisks map[string]string
}
