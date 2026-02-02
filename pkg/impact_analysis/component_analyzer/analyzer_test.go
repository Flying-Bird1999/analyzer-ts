// Package component_analyzer 组件级影响分析测试
package component_analyzer

import (
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/pkg/impact_analysis"
)

// =============================================================================
// ComponentMapper 测试
// =============================================================================

func TestComponentMapper_MapFileToComponent(t *testing.T) {
	manifest := &impact_analysis.ComponentManifest{
		Components: []impact_analysis.Component{
			{Name: "Button", Entry: "/project/src/components/Button/index.tsx"},
			{Name: "Input", Entry: "/project/src/components/Input/index.tsx"},
		},
	}

	mapper := NewComponentMapper(manifest)

	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{
			name:     "Button 组件内部文件",
			filePath: "/project/src/components/Button/Button.tsx",
			expected: "Button",
		},
		{
			name:     "Button 组件入口文件",
			filePath: "/project/src/components/Button/index.tsx",
			expected: "Button",
		},
		{
			name:     "Input 组件内部文件",
			filePath: "/project/src/components/Input/Input.tsx",
			expected: "Input",
		},
		{
			name:     "不属于任何组件的文件",
			filePath: "/project/src/utils.ts",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.MapFileToComponent(tt.filePath)
			if result != tt.expected {
				t.Errorf("MapFileToComponent(%s) = %s, want %s", tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestComponentMapper_MapFilesToComponents(t *testing.T) {
	manifest := &impact_analysis.ComponentManifest{
		Components: []impact_analysis.Component{
			{Name: "Button", Entry: "/project/src/components/Button/index.tsx"},
		},
	}

	mapper := NewComponentMapper(manifest)

	paths := []string{
		"/project/src/components/Button/Button.tsx",
		"/project/src/utils.ts",
		"/project/src/components/Button/types.ts",
	}

	result := mapper.MapFilesToComponents(paths)

	if len(result) != 3 {
		t.Errorf("Expected 3 mappings, got %d", len(result))
	}

	if result["/project/src/components/Button/Button.tsx"] != "Button" {
		t.Error("Button.tsx should map to Button component")
	}

	if result["/project/src/utils.ts"] != "" {
		t.Error("utils.ts should not map to any component")
	}
}

func TestComponentMapper_GetComponentByName(t *testing.T) {
	manifest := &impact_analysis.ComponentManifest{
		Components: []impact_analysis.Component{
			{Name: "Button", Entry: "/project/src/components/Button/index.tsx"},
		},
	}

	mapper := NewComponentMapper(manifest)

	// 测试存在的组件
	button := mapper.GetComponentByName("Button")
	if button == nil {
		t.Fatal("GetComponentByName(Button) should return a component")
	}
	if button.Name != "Button" {
		t.Errorf("Expected component name 'Button', got '%s'", button.Name)
	}

	// 测试不存在的组件
	nonExistent := mapper.GetComponentByName("NonExistent")
	if nonExistent != nil {
		t.Error("GetComponentByName(NonExistent) should return nil")
	}
}

// =============================================================================
// ComponentDependencyGraph 测试
// =============================================================================

func TestBuildComponentDependencyGraph(t *testing.T) {
	manifest := &impact_analysis.ComponentManifest{
		Components: []impact_analysis.Component{
			{Name: "Form", Entry: "/project/src/components/Form/index.tsx"},
			{Name: "Button", Entry: "/project/src/components/Button/index.tsx"},
			{Name: "Input", Entry: "/project/src/components/Input/index.tsx"},
		},
	}

	mapper := NewComponentMapper(manifest)

	// 构造文件依赖图代理
	fileGraph := &FileDependencyGraphProxy{
		DepGraph: map[string][]string{
			"/project/src/components/Form/Form.tsx": {
				"/project/src/components/Button/Button.tsx",
				"/project/src/components/Input/Input.tsx",
			},
		},
		RevDepGraph: map[string][]string{
			"/project/src/components/Button/Button.tsx": {"/project/src/components/Form/Form.tsx"},
			"/project/src/components/Input/Input.tsx":   {"/project/src/components/Form/Form.tsx"},
		},
	}

	// 构建组件依赖图
	componentGraph := mapper.BuildComponentDependencyGraph(fileGraph, nil)

	// 验证正向依赖图
	if len(componentGraph.DepGraph) == 0 {
		t.Fatal("DepGraph should not be empty")
	}

	formDeps := componentGraph.DepGraph["Form"]
	if len(formDeps) != 2 {
		t.Errorf("Form should depend on 2 components, got %d", len(formDeps))
	}

	// 验证反向依赖图
	buttonDependants := componentGraph.RevDepGraph["Button"]
	if len(buttonDependants) != 1 || buttonDependants[0] != "Form" {
		t.Errorf("Button should be depended on by Form, got %v", buttonDependants)
	}
}

// =============================================================================
// Propagator 测试
// =============================================================================

func TestPropagator_Propagate(t *testing.T) {
	// 构造测试组件依赖图：
	// Button ← Form ← App
	// Input ← Form
	graph := &ComponentDependencyGraph{
		DepGraph: map[string][]string{
			"Form": {"Button", "Input"},
			"App":  {"Form"},
		},
		RevDepGraph: map[string][]string{
			"Button": {"Form"},
			"Input":  {"Form"},
			"Form":   {"App"},
		},
	}

	propagator := NewPropagator(graph, 10)

	// 测试：修改 Button
	changedComponents := []string{"Button"}
	result := propagator.Propagate(changedComponents)

	// 验证直接变更
	if len(result.Direct) != 1 {
		t.Errorf("Expected 1 direct change, got %d", len(result.Direct))
	}
	if _, exists := result.Direct["Button"]; !exists {
		t.Error("Button should be in Direct changes")
	}

	// 验证间接受影响（Form 和 App）
	if len(result.Indirect) != 2 {
		t.Errorf("Expected 2 indirect impacts, got %d", len(result.Indirect))
	}

	// Form 应该被影响（层级 1）
	if formImpact, exists := result.Indirect["Form"]; exists {
		if formImpact.ImpactLevel != 1 {
			t.Errorf("Form should have impact level 1, got %d", formImpact.ImpactLevel)
		}
	} else {
		t.Error("Form should be in Indirect impacts")
	}

	// App 应该被影响（层级 2）
	if appImpact, exists := result.Indirect["App"]; exists {
		if appImpact.ImpactLevel != 2 {
			t.Errorf("App should have impact level 2, got %d", appImpact.ImpactLevel)
		}
	} else {
		t.Error("App should be in Indirect impacts")
	}
}

func TestPropagator_Propagate_EmptyChanges(t *testing.T) {
	graph := &ComponentDependencyGraph{
		DepGraph:    make(map[string][]string),
		RevDepGraph: make(map[string][]string),
	}

	propagator := NewPropagator(graph, 10)
	result := propagator.Propagate([]string{})

	if len(result.Direct) != 0 || len(result.Indirect) != 0 {
		t.Error("Expected no impacts for empty changes")
	}
}

// =============================================================================
// ComponentDependencyGraph 方法测试
// =============================================================================

func TestComponentDependencyGraph_GetDependants(t *testing.T) {
	graph := &ComponentDependencyGraph{
		DepGraph: map[string][]string{
			"App": {"Button", "Input"},
		},
		RevDepGraph: map[string][]string{
			"Button": {"App", "Form"},
			"Input":  {"App"},
		},
	}

	// 测试获取依赖者
	buttonDependants := graph.GetDependants("Button")
	if len(buttonDependants) != 2 {
		t.Errorf("Expected 2 dependants for Button, got %d", len(buttonDependants))
	}

	inputDependants := graph.GetDependants("Input")
	if len(inputDependants) != 1 || inputDependants[0] != "App" {
		t.Errorf("Expected [App] for Input dependants, got %v", inputDependants)
	}

	// 测试不存在的组件
	nonExistent := graph.GetDependants("NonExistent")
	if nonExistent != nil {
		t.Errorf("Expected nil for non-existent component, got %v", nonExistent)
	}
}

func TestComponentDependencyGraph_GetDependencies(t *testing.T) {
	graph := &ComponentDependencyGraph{
		DepGraph: map[string][]string{
			"App": {"Button", "Input"},
		},
		RevDepGraph: map[string][]string{},
	}

	appDeps := graph.GetDependencies("App")
	if len(appDeps) != 2 {
		t.Errorf("Expected 2 dependencies for App, got %d", len(appDeps))
	}
}

// =============================================================================
// ImpactedComponents 方法测试
// =============================================================================

func TestImpactedComponents_GetDirectChangedComponents(t *testing.T) {
	impacted := &ImpactedComponents{
		Direct: map[string]*ComponentImpact{
			"Button": {ComponentName: "Button"},
			"Input":  {ComponentName: "Input"},
		},
		Indirect: map[string]*ComponentImpact{},
	}

	direct := impacted.GetDirectChangedComponents()
	if len(direct) != 2 {
		t.Errorf("Expected 2 direct components, got %d", len(direct))
	}
}

func TestImpactedComponents_GetImpactedComponents(t *testing.T) {
	impacted := &ImpactedComponents{
		Direct: map[string]*ComponentImpact{
			"Button": {ComponentName: "Button"},
		},
		Indirect: map[string]*ComponentImpact{
			"Form": {ComponentName: "Form"},
		},
	}

	all := impacted.GetImpactedComponents()
	if len(all) != 2 {
		t.Errorf("Expected 2 total impacted components, got %d", len(all))
	}
}

func TestImpactedComponents_GetComponentImpact(t *testing.T) {
	directImpact := &ComponentImpact{
		ComponentName: "Button",
		ImpactLevel:   0,
	}

	indirectImpact := &ComponentImpact{
		ComponentName: "Form",
		ImpactLevel:   1,
	}

	impacted := &ImpactedComponents{
		Direct:   map[string]*ComponentImpact{"Button": directImpact},
		Indirect: map[string]*ComponentImpact{"Form": indirectImpact},
	}

	// 测试获取直接影响
	button, exists := impacted.GetComponentImpact("Button")
	if !exists {
		t.Error("Button impact should exist")
	}
	if button.ComponentName != "Button" {
		t.Errorf("Expected component name 'Button', got '%s'", button.ComponentName)
	}

	// 测试获取间接影响
	form, exists := impacted.GetComponentImpact("Form")
	if !exists {
		t.Error("Form impact should exist")
	}
	if form.ImpactLevel != 1 {
		t.Errorf("Expected impact level 1 for Form, got %d", form.ImpactLevel)
	}

	// 测试获取不存在的组件
	_, exists = impacted.GetComponentImpact("NonExistent")
	if exists {
		t.Error("NonExistent component should not exist")
	}
}
