// ⚠️废弃

package impact_analysis

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

// =============================================================================
// 影响分析器
// =============================================================================

// Analyzer 影响分析器
type Analyzer struct {
	changeInput    *ChangeInput
	depsDataSource *DepsDataSource
	maxDepth       int
	propagator     *Propagator
	riskAssessor   *RiskAssessor
	chainBuilder   *ChainBuilder
}

// NewAnalyzer 创建影响分析器
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		maxDepth:     10, // 默认最大深度
		propagator:   NewPropagator(10),
		riskAssessor: NewRiskAssessor(),
		chainBuilder: NewChainBuilder(nil, nil), // 稍后初始化
	}
}

// Name 返回分析器名称
func (a *Analyzer) Name() string {
	return "impact-analysis"
}

// Configure 配置分析器
func (a *Analyzer) Configure(params map[string]string) error {
	// 解析变更输入
	if changeJSON, ok := params["changes"]; ok {
		input, err := ParseChangeInput(changeJSON)
		if err != nil {
			return fmt.Errorf("failed to parse changes: %w", err)
		}
		a.changeInput = input
	}

	// 或从文件加载变更输入
	if changeFile, ok := params["changeFile"]; ok {
		input, err := LoadChangeInput(changeFile)
		if err != nil {
			return fmt.Errorf("failed to load change file: %w", err)
		}
		a.changeInput = input
	}

	// 加载依赖数据
	if depsFile, ok := params["depsFile"]; ok {
		a.depsDataSource = &DepsDataSource{
			Type:  "file",
			Value: depsFile,
		}
	}

	// 设置最大深度
	if maxDepth, ok := params["maxDepth"]; ok {
		var depth int
		if _, err := fmt.Sscanf(maxDepth, "%d", &depth); err == nil {
			a.maxDepth = depth
			a.propagator = NewPropagator(depth)
		}
	}

	return nil
}

// Analyze 执行影响分析
func (a *Analyzer) Analyze(ctx *projectanalyzer.ProjectContext) (projectanalyzer.Result, error) {
	// 1. 加载依赖数据
	depData, err := a.loadDependencyData()
	if err != nil {
		return nil, fmt.Errorf("failed to load dependency data: %w", err)
	}

	// 2. 解析变更输入
	if a.changeInput == nil {
		return nil, fmt.Errorf("no change input provided")
	}

	// 3. 识别变更的组件
	changedComponents := a.identifyChangedComponents(ctx, depData)
	if len(changedComponents) == 0 {
		return nil, fmt.Errorf("no components changed")
	}

	// 4. 传播影响
	changedComponentNames := extractComponentNames(changedComponents)
	impactMap := a.propagator.PropagateImpact(
		changedComponentNames,
		depData.DepGraph,
		depData.RevDepGraph,
	)

	// 5. 构建结果
	result := a.buildResult(changedComponents, impactMap, depData)

	// 6. 生成建议
	a.generateRecommendations(result, changedComponents)

	return result, nil
}

// =============================================================================
// 数据加载
// =============================================================================

// DependencyData 依赖数据（从 component-deps-v2 加载）
type DependencyData struct {
	DepGraph    map[string][]string `json:"depGraph"`
	RevDepGraph map[string][]string `json:"revDepGraph"`
	Meta        struct {
		Version        string `json:"version"`
		LibraryName    string `json:"libraryName"`
		ComponentCount int    `json:"componentCount"`
	} `json:"meta"`
}

// loadDependencyData 加载依赖数据
func (a *Analyzer) loadDependencyData() (*DependencyData, error) {
	if a.depsDataSource == nil {
		return nil, fmt.Errorf("no dependency data source configured")
	}

	// 从文件加载
	if a.depsDataSource.Type == "file" {
		data, err := os.ReadFile(a.depsDataSource.Value)
		if err != nil {
			return nil, err
		}

		// 尝试解析为包裹格式 {"component-deps-v2": {...}}
		var wrappedData map[string]json.RawMessage
		if err := json.Unmarshal(data, &wrappedData); err == nil {
			// 如果找到了 component-deps-v2 键，提取其内容
			if raw, exists := wrappedData["component-deps-v2"]; exists {
				var depData DependencyData
				if err := json.Unmarshal(raw, &depData); err == nil {
					return &depData, nil
				}
			}
		}

		// 如果不是包裹格式，直接解析
		var depData DependencyData
		if err := json.Unmarshal(data, &depData); err != nil {
			return nil, err
		}

		return &depData, nil
	}

	return nil, fmt.Errorf("unsupported dependency data source type: %s", a.depsDataSource.Type)
}

// =============================================================================
// 组件识别
// =============================================================================

// identifyChangedComponents 识别变更的组件
func (a *Analyzer) identifyChangedComponents(
	ctx *projectanalyzer.ProjectContext,
	depData *DependencyData,
) []ComponentChange {
	changes := make([]ComponentChange, 0)

	// 遍历所有组件
	for componentName := range depData.DepGraph {
		// 检查组件是否受变更影响
		if a.isComponentAffected(ctx, componentName, depData) {
			changes = append(changes, ComponentChange{
				Name:         componentName,
				Action:       a.getChangeAction(componentName),
				ChangedFiles: a.getComponentChangedFiles(componentName),
			})
		}
	}

	return changes
}

// isComponentAffected 检查组件是否受影响
func (a *Analyzer) isComponentAffected(
	ctx *projectanalyzer.ProjectContext,
	componentName string,
	depData *DependencyData,
) bool {
	allFiles := a.changeInput.GetAllFiles()

	for _, file := range allFiles {
		// 简单的字符串匹配（实际可以根据组件作用域进行更精确的匹配）
		if a.fileBelongsToComponent(file, componentName, depData) {
			return true
		}
	}

	return false
}

// fileBelongsToComponent 判断文件是否属于某个组件
func (a *Analyzer) fileBelongsToComponent(
	file string,
	componentName string,
	depData *DependencyData,
) bool {
	// 简单实现：检查文件路径是否包含组件名
	// 实际应该根据组件 manifest 中的 scope 进行匹配
	return contains(file, componentName)
}

// getChangeAction 获取变更类型
func (a *Analyzer) getChangeAction(componentName string) string {
	// 根据变更文件列表判断变更类型
	// TODO: 实现更精确的判断逻辑
	return "modified"
}

// getComponentChangedFiles 获取组件的变更文件列表
func (a *Analyzer) getComponentChangedFiles(componentName string) []string {
	allFiles := a.changeInput.GetAllFiles()
	files := make([]string, 0)

	for _, file := range allFiles {
		if contains(file, componentName) {
			files = append(files, file)
		}
	}

	return files
}

// =============================================================================
// 结果构建
// =============================================================================

// buildResult 构建分析结果
func (a *Analyzer) buildResult(
	changedComponents []ComponentChange,
	impactMap map[string]*ImpactInfoInternal,
	depData *DependencyData,
) *ImpactAnalysisResult {
	result := &ImpactAnalysisResult{
		Meta: ImpactMeta{
			AnalyzedAt:       time.Now().Format(time.RFC3339),
			ComponentCount:   depData.Meta.ComponentCount,
			ChangedFileCount: a.changeInput.GetFileCount(),
			ChangeSource:     "manual",
		},
		Changes:     changedComponents,
		Impact:      make([]ImpactComponent, 0),
		ChangePaths: make([]ChangePath, 0),
	}

	// 转换 impactMap 到 ImpactComponent
	changedComponentNames := extractComponentNames(changedComponents)
	for _, info := range impactMap {
		impactType := a.riskAssessor.GetImpactType(info.ComponentName, changedComponentNames)
		riskLevel := a.riskAssessor.AssessRisk(info.ImpactLevel, impactType)

		// 转换 ChangePaths 为字符串格式
		pathStrings := make([]string, len(info.ChangePaths))
		for i, path := range info.ChangePaths {
			pathStrings[i] = formatPath(path.Path)
		}

		result.Impact = append(result.Impact, ImpactComponent{
			Name:        info.ComponentName,
			ImpactLevel: info.ImpactLevel,
			RiskLevel:   riskLevel,
			ChangePaths: pathStrings,
		})

		// 添加到 ChangePaths
		result.ChangePaths = append(result.ChangePaths, info.ChangePaths...)
	}

	return result
}

// generateRecommendations 生成建议
func (a *Analyzer) generateRecommendations(
	result *ImpactAnalysisResult,
	changedComponents []ComponentChange,
) {
	result.Recommendations = make([]Recommendation, 0)

	// 统计风险等级
	riskCount := make(map[string]int)
	for _, impact := range result.Impact {
		riskCount[impact.RiskLevel]++
	}

	// 根据风险等级生成建议
	if riskCount["critical"] > 0 {
		result.Recommendations = append(result.Recommendations, Recommendation{
			Type:        "review",
			Priority:    "critical",
			Description: fmt.Sprintf("发现 %d 个严重风险的组件，请进行代码审查", riskCount["critical"]),
		})
	}

	if riskCount["high"] > 0 {
		result.Recommendations = append(result.Recommendations, Recommendation{
			Type:        "test",
			Priority:    "high",
			Description: fmt.Sprintf("发现 %d 个高风险组件，建议补充单元测试", riskCount["high"]),
		})
	}

	if len(changedComponents) > 3 {
		result.Recommendations = append(result.Recommendations, Recommendation{
			Type:        "document",
			Priority:    "medium",
			Description: fmt.Sprintf("本次变更涉及 %d 个组件，建议更新相关文档", len(changedComponents)),
		})
	}
}

// =============================================================================
// 辅助函数
// =============================================================================

// contains 检查字符串是否包含子串（忽略大小写）
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// formatPath 格式化路径
func formatPath(path []string) string {
	if len(path) <= 1 {
		return path[0]
	}
	result := ""
	for i, p := range path {
		if i > 0 {
			result += " → "
		}
		result += p
	}
	return result
}

// extractComponentNames 从 ComponentChange 列表中提取组件名
func extractComponentNames(changes []ComponentChange) []string {
	names := make([]string, len(changes))
	for i, c := range changes {
		names[i] = c.Name
	}
	return names
}
