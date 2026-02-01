// Package impact_analysis 提供符号级影响分析功能。
package impact_analysis

import (
	"container/list"
	"fmt"
)

// =============================================================================
// 影响传播器
// =============================================================================

// Propagator 影响传播器
// 负责将符号变更沿着依赖链传播，找出所有受影响的组件
type Propagator struct {
	depMap            *SymbolDependencyMap
	depGraph          map[string][]string // 组件依赖图 [组件][依赖组件列表]
	revDepGraph       map[string][]string // 反向依赖图 [组件][被依赖组件列表]
	symbolChanges     []SymbolChange      // 原始符号变更
	componentChanges  map[string][]SymbolChange // 按组件分组的符号变更
	maxDepth          int
}

// NewPropagator 创建影响传播器
func NewPropagator(
	depMap *SymbolDependencyMap,
	depGraph map[string][]string,
	revDepGraph map[string][]string,
	symbolChanges []SymbolChange,
	componentChanges map[string][]SymbolChange,
	maxDepth int,
) *Propagator {
	return &Propagator{
		depMap:           depMap,
		depGraph:         depGraph,
		revDepGraph:      revDepGraph,
		symbolChanges:    symbolChanges,
		componentChanges: componentChanges,
		maxDepth:         maxDepth,
	}
}

// Propagate 执行影响传播分析
// 返回: 所有受影响的组件及其影响详情
func (p *Propagator) Propagate() map[string]*ComponentImpact {
	impacts := make(map[string]*ComponentImpact)

	// 步骤 1: 初始化 - 标记变更组件本身为受影响
	for compName := range p.componentChanges {
		impacts[compName] = &ComponentImpact{
			ComponentName:    compName,
			ImpactLevel:      0,
			ChangedSymbols:   p.componentChanges[compName],
			AffectedSymbols:  make(map[string]bool),
			ChangePaths:      []string{compName},
			ImpactType:       classifyComponentChange(p.componentChanges[compName]),
		}

		// 标记所有变更符号为受影响
		for _, sym := range p.componentChanges[compName] {
			impacts[compName].AffectedSymbols[sym.Name] = true
		}
	}

	// 步骤 2: BFS 传播影响
	p.bfsPropagation(impacts)

	// 步骤 3: 符号级精确传播
	p.symbolLevelPropagation(impacts)

	return impacts
}

// bfsPropagation 广度优先搜索传播影响
func (p *Propagator) bfsPropagation(impacts map[string]*ComponentImpact) {
	// 使用 BFS 从变更组件出发，沿依赖链传播影响
	queue := list.New()
	visited := make(map[string]bool)

	// 初始化队列：将所有变更组件加入队列
	for compName := range p.componentChanges {
		queue.PushBack(&propagationNode{
			component:  compName,
			path:       []string{compName},
			depth:      0,
			impactType: impacts[compName].ImpactType,
		})
		visited[compName] = true
	}

	for queue.Len() > 0 {
		current := queue.Remove(queue.Front()).(*propagationNode)

		// 检查深度限制
		if current.depth >= p.maxDepth {
			continue
		}

		// 获取依赖当前组件的其他组件（下游组件）
		downstreamComponents, exists := p.revDepGraph[current.component]
		if !exists {
			continue
		}

		for _, downstream := range downstreamComponents {
			// 构建新的传播路径
			newPath := append(current.path, downstream)

			// 计算影响类型（传播后影响类型会衰减）
			newImpactType := propagateImpactType(current.impactType)

			// 更新或创建下游组件的影响信息
			if existingImpact, exists := impacts[downstream]; exists {
				// 更新影响层级（取最小值）
				if existingImpact.ImpactLevel > current.depth+1 {
					existingImpact.ImpactLevel = current.depth + 1
				}

				// 添加传播路径
				existingImpact.ChangePaths = append(existingImpact.ChangePaths, formatPath(newPath))

				// 更新影响类型（取更严重的）
				if isMoreSevere(newImpactType, existingImpact.ImpactType) {
					existingImpact.ImpactType = newImpactType
				}
			} else {
				// 创建新的影响记录
				impacts[downstream] = &ComponentImpact{
					ComponentName:   downstream,
					ImpactLevel:     current.depth + 1,
					ChangedSymbols:  []SymbolChange{}, // 没有直接变更
					AffectedSymbols: make(map[string]bool),
					ChangePaths:     []string{formatPath(newPath)},
					ImpactType:      newImpactType,
				}
			}

			// 将下游组件加入队列（如果未访问过）
			if !visited[downstream] {
				queue.PushBack(&propagationNode{
					component:  downstream,
					path:       newPath,
					depth:      current.depth + 1,
					impactType: newImpactType,
				})
				visited[downstream] = true
			}
		}
	}
}

// symbolLevelPropagation 符号级精确传播
// 基于 SymbolDependencyMap 进行更精确的符号级影响分析
func (p *Propagator) symbolLevelPropagation(impacts map[string]*ComponentImpact) {
	if p.depMap == nil {
		return
	}

	// 遍历所有符号变更
	for _, symbolChange := range p.symbolChanges {
		// 只处理导出符号的变更
		if !symbolChange.IsExported {
			continue
		}

		// 获取变更符号的组件
		sourceComponent := symbolChange.ComponentName
		if sourceComponent == "" {
			continue
		}

		// 遍历所有组件，查找哪些组件导入了这个符号
		for consumerComp, importRelations := range p.depMap.SymbolImports {
			// 跳过变更组件自身
			if consumerComp == sourceComponent {
				continue
			}

			// 检查是否有从 sourceComponent 导入的符号
			for _, relation := range importRelations {
				if relation.SourceComponent != sourceComponent {
					continue
				}

				// 检查是否导入了变更的符号
				for _, importedSymbol := range relation.ImportedSymbols {
					if importedSymbol.Name == symbolChange.Name {
						// 找到了符号级的影响传播

						// 更新消费组件的影响信息
						if impact, exists := impacts[consumerComp]; exists {
							// 标记受影响的符号
							impact.AffectedSymbols[symbolChange.Name] = true

							// 根据符号变更类型更新影响类型
							symbolImpactType := classifySymbolImpact(symbolChange)

							// 如果符号级影响更严重，则更新组件影响类型
							if isMoreSevere(symbolImpactType, impact.ImpactType) {
								impact.ImpactType = symbolImpactType
							}

							// 添加符号级传播路径
							symbolPath := fmt.Sprintf("%s.%s → %s",
								sourceComponent, symbolChange.Name, consumerComp)
							if !contains(impact.ChangePaths, symbolPath) {
								impact.ChangePaths = append(impact.ChangePaths, symbolPath)
							}
						}
					}
				}
			}
		}
	}
}

// =============================================================================
// 辅助类型和函数
// =============================================================================

// propagationNode 传播节点（用于 BFS）
type propagationNode struct {
	component  string
	path       []string
	depth      int
	impactType ImpactType
}

// ComponentImpact 组件影响信息
type ComponentImpact struct {
	ComponentName   string             // 组件名称
	ImpactLevel     int                // 影响层级（0=直接变更，>0=传播层级）
	ChangedSymbols  []SymbolChange     // 该组件的直接变更（可能为空）
	AffectedSymbols map[string]bool    // 受影响的符号（来自上游变更）
	ChangePaths     []string           // 从变更组件到该组件的路径
	ImpactType      ImpactType         // 影响类型
}

// classifyComponentChange 根据组件的所有符号变更分类影响类型
func classifyComponentChange(changes []SymbolChange) ImpactType {
	hasBreaking := false
	hasInternal := false

	for _, change := range changes {
		if !change.IsExported {
			hasInternal = true
			continue
		}

		switch change.ChangeType {
		case ChangeTypeDeleted:
			hasBreaking = true
		case ChangeTypeModified:
			hasBreaking = true
		case ChangeTypeAdded:
			// hasAdditive = true // 只新增导出是向后兼容的
		}
	}

	// 优先级：breaking > internal > additive
	if hasBreaking {
		return ImpactTypeBreaking
	}
	if hasInternal {
		return ImpactTypeInternal
	}
	return ImpactTypeAdditive
}

// classifySymbolImpact 根据单个符号变更分类影响类型
func classifySymbolImpact(change SymbolChange) ImpactType {
	// 非导出符号：内部变更
	if !change.IsExported {
		return ImpactTypeInternal
	}

	// 导出符号：根据变更类型分类
	switch change.ChangeType {
	case ChangeTypeDeleted:
		return ImpactTypeBreaking
	case ChangeTypeModified:
		return ImpactTypeBreaking
	case ChangeTypeAdded:
		return ImpactTypeAdditive
	default:
		return ImpactTypeInternal
	}
}

// propagateImpactType 计算传播后的影响类型
func propagateImpactType(sourceType ImpactType) ImpactType {
	// 影响类型在传播时会衰减
	// breaking → internal（破坏性变更传播后变成潜在风险）
	// internal → internal（内部变更保持内部）
	// additive → additive（增强性变更保持增强性）
	switch sourceType {
	case ImpactTypeBreaking:
		return ImpactTypeInternal
	default:
		return sourceType
	}
}

// isMoreSevere 判断 impactType1 是否比 impactType2 更严重
func isMoreSevere(impactType1, impactType2 ImpactType) bool {
	severity := map[ImpactType]int{
		ImpactTypeBreaking: 3,
		ImpactTypeInternal: 2,
		ImpactTypeAdditive: 1,
	}
	return severity[impactType1] > severity[impactType2]
}

// formatPath 格式化路径为字符串
func formatPath(path []string) string {
	result := ""
	for i, comp := range path {
		if i > 0 {
			result += " → "
		}
		result += comp
	}
	return result
}

// contains 检查字符串切片是否包含某个元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
