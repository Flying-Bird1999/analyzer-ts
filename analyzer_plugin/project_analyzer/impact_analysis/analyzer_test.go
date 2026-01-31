package impact_analysis

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Propagator 测试
// =============================================================================

func TestPropagator_PropagateImpact(t *testing.T) {
	tests := []struct {
		name             string
		changedComponents []string
		depGraph         map[string][]string
		revDepGraph      map[string][]string
		expectedCount    int
		expectedLevels   map[string]int
	}{
		{
			name:             "单个组件变更",
			changedComponents: []string{"Button"},
			depGraph: map[string][]string{
				"Button": {},
				"Input":  {"Button"},
				"Select": {"Button", "Input"},
			},
			revDepGraph: map[string][]string{
				"Button": {"Input", "Select"},
				"Input":  {"Select"},
				"Select": {},
			},
			expectedCount: 3,
			expectedLevels: map[string]int{
				"Button": 0,
				"Input":  1,
				"Select": 1, // Select directly depends on Button, so level 1
			},
		},
		{
			name:             "多个组件变更",
			changedComponents: []string{"Button", "Input"},
			depGraph: map[string][]string{
				"Button": {},
				"Input":  {"Button"},
				"Select": {"Button", "Input"},
			},
			revDepGraph: map[string][]string{
				"Button": {"Input", "Select"},
				"Input":  {"Select"},
				"Select": {},
			},
			expectedCount: 3,
			expectedLevels: map[string]int{
				"Button": 0,
				"Input":  0,
				"Select": 1,
			},
		},
		{
			name:             "无影响传播",
			changedComponents: []string{"Button"},
			depGraph: map[string][]string{
				"Button": {},
				"Input":  {},
				"Select": {},
			},
			revDepGraph: map[string][]string{
				"Button": {},
				"Input":  {},
				"Select": {},
			},
			expectedCount: 1,
			expectedLevels: map[string]int{
				"Button": 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			propagator := NewPropagator(10)
			result := propagator.PropagateImpact(
				tt.changedComponents,
				tt.depGraph,
				tt.revDepGraph,
			)

			assert.Equal(t, tt.expectedCount, len(result), "影响组件数量不匹配")

			for comp, expectedLevel := range tt.expectedLevels {
				info, exists := result[comp]
				assert.True(t, exists, "组件 %s 不在结果中", comp)
				assert.Equal(t, expectedLevel, info.ImpactLevel, "组件 %s 影响层级不匹配", comp)
			}
		})
	}
}

func TestPropagator_MaxDepth(t *testing.T) {
	propagator := NewPropagator(1) // 最大深度为 1

	depGraph := map[string][]string{
		"A": {},
		"B": {"A"},
		"C": {"B"},
	}
	revDepGraph := map[string][]string{
		"A": {"B"},
		"B": {"C"},
		"C": {},
	}

	result := propagator.PropagateImpact([]string{"A"}, depGraph, revDepGraph)

	// A (level 0) -> B (level 1) -> C 不应该被传播（超过最大深度）
	assert.Equal(t, 2, len(result))
	assert.Equal(t, 0, result["A"].ImpactLevel)
	assert.Equal(t, 1, result["B"].ImpactLevel)
	assert.NotContains(t, result, "C")
}

// =============================================================================
// RiskAssessor 测试
// =============================================================================

func TestRiskAssessor_AssessRisk(t *testing.T) {
	assessor := NewRiskAssessor()

	tests := []struct {
		level        int
		impactType   string
		expectedRisk string
	}{
		{0, "direct", "low"},
		{1, "indirect", "low"},
		{2, "indirect", "medium"},
		{3, "indirect", "high"},
		{4, "indirect", "critical"},
		{5, "indirect", "critical"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedRisk, func(t *testing.T) {
			risk := assessor.AssessRisk(tt.level, tt.impactType)
			assert.Equal(t, tt.expectedRisk, risk)
		})
	}
}

func TestRiskAssessor_GetImpactType(t *testing.T) {
	assessor := NewRiskAssessor()
	changedComponents := []string{"Button", "Input"}

	tests := []struct {
		component      string
		expectedType   string
	}{
		{"Button", "direct"},
		{"Input", "direct"},
		{"Select", "indirect"},
	}

	for _, tt := range tests {
		t.Run(tt.component, func(t *testing.T) {
			impactType := assessor.GetImpactType(tt.component, changedComponents)
			assert.Equal(t, tt.expectedType, impactType)
		})
	}
}

// =============================================================================
// ChainBuilder 测试
// =============================================================================

func TestChainBuilder_BuildImpactChains(t *testing.T) {
	depGraph := map[string][]string{
		"Button": {},
		"Input":  {"Button"},
		"Select": {"Button", "Input"},
	}
	revDepGraph := map[string][]string{
		"Button": {"Input", "Select"},
		"Input":  {"Select"},
		"Select": {},
	}

	builder := NewChainBuilder(depGraph, revDepGraph)

	changed := []string{"Button"}
	impacted := []string{"Select"}

	chains := builder.BuildImpactChains(changed, impacted)

	// 应该找到两条路径：Button -> Select 和 Button -> Input -> Select
	assert.GreaterOrEqual(t, len(chains), 1)
}

func TestChainBuilder_HasCycle(t *testing.T) {
	tests := []struct {
		name     string
		depGraph map[string][]string
		hasCycle bool
	}{
		{
			name: "无循环",
			depGraph: map[string][]string{
				"A": {},
				"B": {"A"},
				"C": {"B"},
			},
			hasCycle: false,
		},
		{
			name: "有循环",
			depGraph: map[string][]string{
				"A": {"B"},
				"B": {"C"},
				"C": {"A"},
			},
			hasCycle: true,
		},
		{
			name: "自循环",
			depGraph: map[string][]string{
				"A": {"A"},
			},
			hasCycle: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			revDepGraph := buildReverseDepGraph(tt.depGraph)
			builder := NewChainBuilder(tt.depGraph, revDepGraph)
			assert.Equal(t, tt.hasCycle, builder.HasCycle())
		})
	}
}

func TestChainBuilder_DetectCycles(t *testing.T) {
	depGraph := map[string][]string{
		"A": {"B"},
		"B": {"C"},
		"C": {"A"},
	}
	revDepGraph := buildReverseDepGraph(depGraph)

	builder := NewChainBuilder(depGraph, revDepGraph)
	cycles := builder.DetectCycles()

	assert.Equal(t, 1, len(cycles))
	assert.Contains(t, cycles[0], "A")
	assert.Contains(t, cycles[0], "B")
	assert.Contains(t, cycles[0], "C")
}

// =============================================================================
// ChangeInput 测试
// =============================================================================

func TestChangeInput_GetAllFiles(t *testing.T) {
	input := &ChangeInput{
		ModifiedFiles: []string{"a.ts", "b.ts"},
		AddedFiles:    []string{"c.ts"},
		DeletedFiles:  []string{"d.ts"},
	}

	files := input.GetAllFiles()
	assert.Equal(t, 4, len(files))
	assert.Contains(t, files, "a.ts")
	assert.Contains(t, files, "b.ts")
	assert.Contains(t, files, "c.ts")
	assert.Contains(t, files, "d.ts")
}

func TestChangeInput_GetFileCount(t *testing.T) {
	input := &ChangeInput{
		ModifiedFiles: []string{"a.ts", "b.ts"},
		AddedFiles:    []string{"c.ts"},
		DeletedFiles:  []string{"d.ts"},
	}

	assert.Equal(t, 4, input.GetFileCount())
}

func TestChangeInput_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    *ChangeInput
		expected bool
	}{
		{
			name:     "空变更",
			input:    &ChangeInput{},
			expected: true,
		},
		{
			name: "有修改文件",
			input: &ChangeInput{
				ModifiedFiles: []string{"a.ts"},
			},
			expected: false,
		},
		{
			name: "有新增文件",
			input: &ChangeInput{
				AddedFiles: []string{"b.ts"},
			},
			expected: false,
		},
		{
			name: "有删除文件",
			input: &ChangeInput{
				DeletedFiles: []string{"c.ts"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.input.IsEmpty())
		})
	}
}

func TestParseChangeInput(t *testing.T) {
	jsonStr := `{
		"modifiedFiles": ["a.ts", "b.ts"],
		"addedFiles": ["c.ts"],
		"deletedFiles": ["d.ts"]
	}`

	input, err := ParseChangeInput(jsonStr)
	assert.NoError(t, err)
	assert.Equal(t, []string{"a.ts", "b.ts"}, input.ModifiedFiles)
	assert.Equal(t, []string{"c.ts"}, input.AddedFiles)
	assert.Equal(t, []string{"d.ts"}, input.DeletedFiles)
}

// =============================================================================
// Result 测试
// =============================================================================

func TestImpactAnalysisResult_Name(t *testing.T) {
	result := &ImpactAnalysisResult{}
	assert.Equal(t, "impact-analysis", result.Name())
}

func TestImpactAnalysisResult_Summary(t *testing.T) {
	result := &ImpactAnalysisResult{
		Changes: []ComponentChange{
			{Name: "Button"},
			{Name: "Input"},
		},
		Impact: []ImpactComponent{
			{Name: "Button"},
			{Name: "Input"},
			{Name: "Select"},
		},
	}

	summary := result.Summary()
	assert.Contains(t, summary, "2")
	assert.Contains(t, summary, "3")
}

func TestImpactAnalysisResult_ToJSON(t *testing.T) {
	result := &ImpactAnalysisResult{
		Meta: ImpactMeta{
			AnalyzedAt:       "2024-01-31T00:00:00Z",
			ComponentCount:   3,
			ChangedFileCount: 2,
			ChangeSource:     "manual",
		},
		Changes: []ComponentChange{
			{Name: "Button", Action: "modified"},
		},
		Impact: []ImpactComponent{
			{Name: "Button", ImpactLevel: 0, RiskLevel: "low"},
		},
	}

	data, err := result.ToJSON(true)
	assert.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	// Verify the meta field exists and has correct values
	meta, ok := decoded["meta"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "2024-01-31T00:00:00Z", meta["analyzedAt"])
	assert.Equal(t, float64(3), meta["componentCount"])

	// Verify the changes field exists
	changes, ok := decoded["changes"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 1, len(changes))

	// Verify the impact field exists
	impact, ok := decoded["impact"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 1, len(impact))
}

// =============================================================================
// 辅助函数
// =============================================================================

func buildReverseDepGraph(depGraph map[string][]string) map[string][]string {
	revDepGraph := make(map[string][]string)
	for from, tos := range depGraph {
		for _, to := range tos {
			revDepGraph[to] = append(revDepGraph[to], from)
		}
	}
	return revDepGraph
}

func TestExtractComponentNames(t *testing.T) {
	changes := []ComponentChange{
		{Name: "Button"},
		{Name: "Input"},
	}

	names := extractComponentNames(changes)
	assert.Equal(t, []string{"Button", "Input"}, names)
}
