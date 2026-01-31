package impact_analysis

// =============================================================================
// 影响传播算法
// =============================================================================

// Propagator 影响传播器
type Propagator struct {
	maxDepth int
}

// NewPropagator 创建影响传播器
func NewPropagator(maxDepth int) *Propagator {
	return &Propagator{
		maxDepth: maxDepth,
	}
}

// PropagateImpact 传播影响范围
// 输入: 变更的组件列表、依赖图、反向依赖图
// 输出: 受影响的组件列表（带影响层级）
func (p *Propagator) PropagateImpact(
	changedComponents []string,
	depGraph map[string][]string,
	revDepGraph map[string][]string,
) map[string]*ImpactInfoInternal {
	// 初始化
	impactMap := make(map[string]*ImpactInfoInternal)
	queue := make([]string, 0)

	// 将变更组件加入队列
	for _, comp := range changedComponents {
		impactMap[comp] = &ImpactInfoInternal{
			ComponentName: comp,
			ImpactLevel:   0,
			ChangePaths: []ChangePath{
				{From: comp, To: comp, Path: []string{comp}},
			},
		}
		queue = append(queue, comp)
	}

	// BFS 传播
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		currentInfo := impactMap[current]

		// 如果超过最大深度，停止传播
		if currentInfo.ImpactLevel >= p.maxDepth {
			continue
		}

		// 获取所有依赖当前组件的其他组件（下游）
		// 通过反向依赖图找到所有依赖 current 的组件
		downstreamComps := revDepGraph[current]
		for _, downstream := range downstreamComps {
			// 如果还没有处理过
			if _, exists := impactMap[downstream]; !exists {
				// 计算影响层级
				newLevel := currentInfo.ImpactLevel + 1

				// 构建变更路径
				newPaths := make([]ChangePath, 0, len(currentInfo.ChangePaths))
				for _, path := range currentInfo.ChangePaths {
					newPath := ChangePath{
						From: path.From,
						To:   downstream,
						Path: append(path.Path, downstream),
					}
					newPaths = append(newPaths, newPath)
				}

				impactMap[downstream] = &ImpactInfoInternal{
					ComponentName: downstream,
					ImpactLevel:   newLevel,
					ChangePaths:   newPaths,
				}

				// 加入队列继续传播
				queue = append(queue, downstream)
			}
		}
	}

	return impactMap
}

// ImpactInfoInternal 影响信息（内部使用）
type ImpactInfoInternal struct {
	ComponentName string        // 组件名称
	ImpactLevel   int           // 影响层级
	ChangePaths   []ChangePath  // 从变更组件到该组件的所有路径
}

// =============================================================================
// 风险评估
// =============================================================================

// RiskLevel 风险等级
type RiskLevel int

const (
	RiskLow      RiskLevel = iota // 低风险
	RiskMedium                    // 中风险
	RiskHigh                      // 高风险
	RiskCritical                  // 严重风险
)

// RiskAssessor 风险评估器
type RiskAssessor struct{}

// NewRiskAssessor 创建风险评估器
func NewRiskAssessor() *RiskAssessor {
	return &RiskAssessor{}
}

// AssessRisk 评估风险等级
func (r *RiskAssessor) AssessRisk(
	impactLevel int,
	impactType string,
) string {
	// 根据影响层级和类型评估风险
	switch {
	case impactLevel >= 4:
		return "critical"
	case impactLevel >= 3:
		return "high"
	case impactLevel >= 2:
		return "medium"
	default:
		return "low"
	}
}

// CalculateRiskLevel 计算风险等级（数值）
func (r *RiskAssessor) CalculateRiskLevel(impactLevel int, isDirectChange bool) int {
	risk := impactLevel
	if isDirectChange {
		risk += 1 // 直接变更的组件风险加一
	}
	return risk
}

// GetImpactType 获取影响类型
func (r *RiskAssessor) GetImpactType(componentName string, changedComponents []string) string {
	for _, changed := range changedComponents {
		if changed == componentName {
			return "direct"
		}
	}
	return "indirect"
}
