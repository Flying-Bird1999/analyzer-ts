package component_deps_v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 配置加载测试
// =============================================================================

func TestLoadManifest_Success(t *testing.T) {
	// TODO: 创建测试配置文件并验证加载
	t.Skip("需要创建测试配置文件")
}

func TestLoadManifest_FileNotFound(t *testing.T) {
	_, err := LoadManifest("/non/existent/path.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "配置文件不存在")
}

func TestLoadManifest_InvalidJSON(t *testing.T) {
	// TODO: 创建无效 JSON 文件测试
	t.Skip("需要创建无效 JSON 测试文件")
}

func TestValidateManifest_EmptyComponents(t *testing.T) {
	manifest := &ComponentManifest{
		Components: []ComponentDefinition{},
	}

	err := validateManifest(manifest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "components 列表不能为空")
}

func TestValidateManifest_DuplicateComponentNames(t *testing.T) {
	manifest := &ComponentManifest{

		Components: []ComponentDefinition{
			{Name: "Button", Entry: "src/Button/index.tsx"},
			{Name: "Button", Entry: "src/Button2/index.tsx"},
		},
	}

	err := validateManifest(manifest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "组件名称重复")
}

func TestGetComponentByName(t *testing.T) {
	manifest := &ComponentManifest{

		Components: []ComponentDefinition{
			{Name: "Button", Entry: "src/Button/index.tsx"},
			{Name: "Input", Entry: "src/Input/index.tsx"},
		},
	}

	comp, ok := manifest.GetComponentByName("Button")
	assert.True(t, ok)
	assert.Equal(t, "Button", comp.Name)
	assert.Equal(t, "src/Button/index.tsx", comp.Entry)

	_, ok = manifest.GetComponentByName("NonExistent")
	assert.False(t, ok)
}

func TestGetComponentCount(t *testing.T) {
	manifest := &ComponentManifest{
		Components: []ComponentDefinition{
			{Name: "Button", Entry: "src/Button/index.tsx"},
			{Name: "Input", Entry: "src/Input/index.tsx"},
		},
	}

	assert.Equal(t, 2, manifest.GetComponentCount())
}

func TestGetComponentNames(t *testing.T) {
	manifest := &ComponentManifest{
		Components: []ComponentDefinition{
			{Name: "Button", Entry: "src/Button/index.tsx"},
			{Name: "Input", Entry: "src/Input/index.tsx"},
		},
	}

	names := manifest.GetComponentNames()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "Button")
	assert.Contains(t, names, "Input")
}

// =============================================================================
// 作用域测试
// =============================================================================

func TestComponentScope_Contains(t *testing.T) {
	comp := &ComponentDefinition{
		Name:  "Button",
		Entry: "src/Button/index.tsx",
	}

	scope := NewComponentScope(comp)

	// 测试匹配
	assert.True(t, scope.Contains("src/Button/index.tsx"))
	assert.True(t, scope.Contains("src/Button/Button.tsx"))
	assert.True(t, scope.Contains("src/Button/components/ButtonIcon.tsx"))
	assert.True(t, scope.Contains("src/Button/utils/helpers.ts"))

	// 测试不匹配
	assert.False(t, scope.Contains("src/Input/index.tsx"))
	assert.False(t, scope.Contains("src/ButtonTest/index.tsx"))
}

func TestMultiComponentScope_FindComponentByFile(t *testing.T) {
	manifest := &ComponentManifest{

		Components: []ComponentDefinition{
			{
				Name:  "Button",
				Entry: "src/Button/index.tsx",
			},
			{
				Name:  "Input",
				Entry: "src/Input/index.tsx",
			},
		},
	}

	scope := NewMultiComponentScope(manifest, "/test/project")

	// 测试查找
	compName, ok := scope.FindComponentByFile("src/Button/index.tsx")
	assert.True(t, ok)
	assert.Equal(t, "Button", compName)

	compName, ok = scope.FindComponentByFile("src/Input/Input.tsx")
	assert.True(t, ok)
	assert.Equal(t, "Input", compName)

	// 测试未找到
	_, ok = scope.FindComponentByFile("src/Select/index.tsx")
	assert.False(t, ok)
}

func TestMultiComponentScope_CrossComponentDetection(t *testing.T) {
	manifest := &ComponentManifest{

		Components: []ComponentDefinition{
			{
				Name:  "Button",
				Entry: "src/Button/index.tsx",
			},
			{
				Name:  "Input",
				Entry: "src/Input/index.tsx",
			},
		},
	}

	scope := NewMultiComponentScope(manifest, "/test/project")

	// 测试跨组件检测
	targetComp, isCross, isExternal := scope.DetectCrossComponentImports(
		"src/Input/index.tsx", "src/Button/Button.tsx")

	assert.Equal(t, "Input", targetComp)
	assert.True(t, isCross)
	assert.False(t, isExternal)

	// 测试同组件内
	targetComp, isCross, isExternal = scope.DetectCrossComponentImports(
		"src/Button/utils.ts", "src/Button/index.tsx")

	assert.Equal(t, "Button", targetComp)
	assert.False(t, isCross)
	assert.False(t, isExternal)

	// 测试外部导入
	targetComp, isCross, isExternal = scope.DetectCrossComponentImports(
		"src/External/index.tsx", "src/Button/index.tsx")

	assert.Equal(t, "", targetComp)
	assert.False(t, isCross)
	assert.True(t, isExternal)
}

// =============================================================================
// 依赖图测试
// =============================================================================

func TestGraphBuilder_BuildDepGraph(t *testing.T) {
	manifest := &ComponentManifest{

		Components: []ComponentDefinition{
			{Name: "Button", Entry: "src/Button/index.tsx"},
			{Name: "Input", Entry: "src/Input/index.tsx"},
			{Name: "Select", Entry: "src/Select/index.tsx"},
		},
	}

	builder := NewGraphBuilder(manifest)

	dependencies := map[string][]string{
		"Button": {},
		"Input":  {"Button"},
		"Select": {"Input", "Button"},
	}

	graph := builder.BuildDepGraph(dependencies)

	assert.Len(t, graph, 3)
	assert.Empty(t, graph["Button"])
	assert.Equal(t, []string{"Button"}, graph["Input"])
	assert.Equal(t, []string{"Button", "Input"}, graph["Select"])
}

func TestGraphBuilder_BuildRevDepGraph(t *testing.T) {
	manifest := &ComponentManifest{

		Components: []ComponentDefinition{
			{Name: "Button", Entry: "src/Button/index.tsx"},
			{Name: "Input", Entry: "src/Input/index.tsx"},
			{Name: "Select", Entry: "src/Select/index.tsx"},
		},
	}

	builder := NewGraphBuilder(manifest)

	depGraph := DependencyGraph{
		"Button": {},
		"Input":  {"Button"},
		"Select": {"Input", "Button"},
	}

	revGraph := builder.BuildRevDepGraph(depGraph)

	assert.Len(t, revGraph, 3)
	assert.Equal(t, []string{"Input", "Select"}, revGraph["Button"])
	assert.Equal(t, []string{"Select"}, revGraph["Input"])
	assert.Empty(t, revGraph["Select"])
}

func TestGraphBuilder_DetectCycles(t *testing.T) {
	manifest := &ComponentManifest{

		Components: []ComponentDefinition{
			{Name: "A", Entry: "src/A/index.tsx"},
			{Name: "B", Entry: "src/B/index.tsx"},
			{Name: "C", Entry: "src/C/index.tsx"},
		},
	}

	builder := NewGraphBuilder(manifest)

	// 测试无环
	depGraph := DependencyGraph{
		"A": {},
		"B": {"A"},
		"C": {"B"},
	}
	cycles := builder.DetectCycles(depGraph)
	assert.Empty(t, cycles)

	// 测试有环
	depGraphWithCycle := DependencyGraph{
		"A": {"B"},
		"B": {"C"},
		"C": {"A"},
	}
	cycles = builder.DetectCycles(depGraphWithCycle)
	assert.NotEmpty(t, cycles)
}

// =============================================================================
// 分析器接口测试
// =============================================================================

func TestComponentDepsV2Analyzer_Name(t *testing.T) {
	analyzer := &ComponentDepsV2Analyzer{}
	assert.Equal(t, "component-deps-v2", analyzer.Name())
}

func TestComponentDepsV2Analyzer_Configure_MissingParam(t *testing.T) {
	analyzer := &ComponentDepsV2Analyzer{}
	err := analyzer.Configure(map[string]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "缺少必需参数")
}

func TestComponentDepsV2Analyzer_Configure_Success(t *testing.T) {
	analyzer := &ComponentDepsV2Analyzer{}
	err := analyzer.Configure(map[string]string{
		"manifest": "component-manifest.json",
	})
	assert.NoError(t, err)
	assert.Equal(t, "component-manifest.json", analyzer.ManifestPath)
}

// =============================================================================
// 结果接口测试
// =============================================================================

func TestComponentDepsV2Result_Name(t *testing.T) {
	result := &ComponentDepsV2Result{}
	assert.Equal(t, "component-deps-v2", result.Name())
}

func TestComponentDepsV2Result_Summary(t *testing.T) {
	result := &ComponentDepsV2Result{
		Meta: Meta{
			ComponentCount: 3,
		},
		DepGraph: DependencyGraph{
			"Button": {},
			"Input":  {"Button"},
			"Select": {"Input", "Button"},
		},
	}

	summary := result.Summary()
	assert.Contains(t, summary, "3 个组件")
	assert.Contains(t, summary, "3 条依赖")
}

func TestComponentDepsV2Result_ToJSON(t *testing.T) {
	result := &ComponentDepsV2Result{
		Meta: Meta{
			ComponentCount: 1,
		},
		Components: map[string]ComponentInfo{
			"Button": {
				Name:         "Button",
				Entry:        "src/Button/index.tsx",
				Dependencies: []string{},
			},
		},
		DepGraph:    DependencyGraph{"Button": {}},
		RevDepGraph: ReverseDepGraph{"Button": {}},
	}

	// 测试带缩进
	jsonWithIndent, err := result.ToJSON(true)
	require.NoError(t, err)
	assert.Contains(t, string(jsonWithIndent), "Button")
	assert.Contains(t, string(jsonWithIndent), "\n")

	// 测试不带缩进
	jsonWithoutIndent, err := result.ToJSON(false)
	require.NoError(t, err)
	assert.Contains(t, string(jsonWithoutIndent), "Button")
}

func TestComponentDepsV2Result_ToConsole(t *testing.T) {
	result := &ComponentDepsV2Result{
		Meta: Meta{
			ComponentCount: 2,
		},
		Components: map[string]ComponentInfo{
			"Button": {
				Name:         "Button",
				Entry:        "src/Button/index.tsx",
				Dependencies: []string{},
			},
			"Input": {
				Name:         "Input",
				Entry:        "src/Input/index.tsx",
				Dependencies: []string{"Button"},
			},
		},
		DepGraph: DependencyGraph{
			"Button": {},
			"Input":  {"Button"},
		},
		RevDepGraph: ReverseDepGraph{
			"Button": {"Input"},
			"Input":  {},
		},
	}

	output := result.ToConsole()
	assert.Contains(t, output, "组件依赖分析报告")
	assert.Contains(t, output, "组件总数: 2")
	assert.Contains(t, output, "Button")
	assert.Contains(t, output, "Input")
	assert.Contains(t, output, "反向依赖")
}

// =============================================================================
// 集成测试
// =============================================================================

func TestComponentDepsV2Analyzer_Analyze_Integration(t *testing.T) {
	// TODO: 完整的集成测试需要：
	// 1. 创建测试项目结构
	// 2. 创建配置文件
	// 3. 创建测试文件
	// 4. 运行分析器
	// 5. 验证结果
	t.Skip("需要完整的测试项目设置")
}
