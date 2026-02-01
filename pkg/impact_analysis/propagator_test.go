// Package impact_analysis 单元测试
package impact_analysis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Propagator 测试
// =============================================================================

func TestNewPropagator(t *testing.T) {
	depMap := NewSymbolDependencyMap()
	depGraph := map[string][]string{}
	revDepGraph := map[string][]string{}
	symbolChanges := []SymbolChange{}
	componentChanges := map[string][]SymbolChange{}

	propagator := NewPropagator(depMap, depGraph, revDepGraph, symbolChanges, componentChanges, 10)

	assert.NotNil(t, propagator)
	assert.Equal(t, depMap, propagator.depMap)
	assert.Equal(t, 10, propagator.maxDepth)
}

// TestPropagate 测试影响传播
func TestPropagate(t *testing.T) {
	// 创建测试数据
	depMap := NewSymbolDependencyMap()
	depMap.ComponentExports["Button"] = []SymbolRef{
		{Name: "handleClick", Kind: SymbolKindFunction, ExportType: ExportTypeNamed},
	}
	depMap.SymbolImports["App"] = []ImportRelation{
		{
			SourceComponent: "Button",
			ImportedSymbols: []SymbolRef{
				{Name: "handleClick", Kind: SymbolKindFunction},
			},
			ImportType: "named",
		},
	}

	depGraph := map[string][]string{
		"App": []string{"Button"},
	}
	revDepGraph := map[string][]string{
		"Button": []string{"App"},
	}

	symbolChanges := []SymbolChange{
		{
			Name:         "handleClick",
			Kind:         SymbolKindFunction,
			FilePath:     "src/Button.tsx",
			IsExported:   true,
			ExportType:   ExportTypeNamed,
			ChangeType:   ChangeTypeModified,
			ComponentName: "Button",
		},
	}

	componentChanges := map[string][]SymbolChange{
		"Button": {symbolChanges[0]},
	}

	propagator := NewPropagator(depMap, depGraph, revDepGraph, symbolChanges, componentChanges, 10)

	// 执行传播
	impacts := propagator.Propagate()

	assert.NotNil(t, impacts)
	assert.Len(t, impacts, 2) // Button 和 App

	// 验证 Button 组件（直接变更）
	buttonImpact := impacts["Button"]
	assert.NotNil(t, buttonImpact)
	assert.Equal(t, 0, buttonImpact.ImpactLevel)
	assert.Equal(t, ImpactTypeBreaking, buttonImpact.ImpactType)

	// 验证 App 组件（间接影响）
	appImpact := impacts["App"]
	assert.NotNil(t, appImpact)
	assert.Equal(t, 1, appImpact.ImpactLevel)
	// 注意：由于符号级传播检测到 App 导入了 Button 的 handleClick（破坏性变更）
	// 所以 App 的 ImpactType 保持为 breaking 而不是衰减后的 internal
	assert.Equal(t, ImpactTypeBreaking, appImpact.ImpactType)
}

// TestPropagate_EmptyChanges 测试空变更情况
func TestPropagate_EmptyChanges(t *testing.T) {
	depMap := NewSymbolDependencyMap()
	depGraph := map[string][]string{}
	revDepGraph := map[string][]string{}
	symbolChanges := []SymbolChange{}
	componentChanges := map[string][]SymbolChange{}

	propagator := NewPropagator(depMap, depGraph, revDepGraph, symbolChanges, componentChanges, 10)
	impacts := propagator.Propagate()

	assert.Empty(t, impacts)
}

// TestClassifyComponentChange 测试组件变更分类
func TestClassifyComponentChange(t *testing.T) {
	tests := []struct {
		name     string
		changes  []SymbolChange
		wantType ImpactType
	}{
		{
			name: "破坏性变更（删除导出）",
			changes: []SymbolChange{
				{Name: "foo", IsExported: true, ChangeType: ChangeTypeDeleted},
			},
			wantType: ImpactTypeBreaking,
		},
		{
			name: "破坏性变更（修改导出）",
			changes: []SymbolChange{
				{Name: "foo", IsExported: true, ChangeType: ChangeTypeModified},
			},
			wantType: ImpactTypeBreaking,
		},
		{
			name: "内部变更",
			changes: []SymbolChange{
				{Name: "internalHelper", IsExported: false},
			},
			wantType: ImpactTypeInternal,
		},
		{
			name: "增强性变更（只有新增）",
			changes: []SymbolChange{
				{Name: "newFeature", IsExported: true, ChangeType: ChangeTypeAdded},
			},
			wantType: ImpactTypeAdditive,
		},
		{
			name: "混合变更（破坏性优先）",
			changes: []SymbolChange{
				{Name: "deleted", IsExported: true, ChangeType: ChangeTypeDeleted},
				{Name: "added", IsExported: true, ChangeType: ChangeTypeAdded},
			},
			wantType: ImpactTypeBreaking,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifyComponentChange(tt.changes)
			assert.Equal(t, tt.wantType, result)
		})
	}
}

// TestClassifySymbolImpact 测试符号影响分类
func TestClassifySymbolImpact(t *testing.T) {
	tests := []struct {
		name   string
		change SymbolChange
		want   ImpactType
	}{
		{
			name: "导出符号被删除",
			change: SymbolChange{
				Name:       "foo",
				IsExported: true,
				ChangeType: ChangeTypeDeleted,
			},
			want: ImpactTypeBreaking,
		},
		{
			name: "导出符号被修改",
			change: SymbolChange{
				Name:       "foo",
				IsExported: true,
				ChangeType: ChangeTypeModified,
			},
			want: ImpactTypeBreaking,
		},
		{
			name: "导出符号被新增",
			change: SymbolChange{
				Name:       "foo",
				IsExported: true,
				ChangeType: ChangeTypeAdded,
			},
			want: ImpactTypeAdditive,
		},
		{
			name: "非导出符号变更",
			change: SymbolChange{
				Name:       "internal",
				IsExported: false,
			},
			want: ImpactTypeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifySymbolImpact(tt.change)
			assert.Equal(t, tt.want, result)
		})
	}
}

// TestPropagateImpactType 测试影响类型传播衰减
func TestPropagateImpactType(t *testing.T) {
	tests := []struct {
		name     string
		source   ImpactType
		expected ImpactType
	}{
		{
			name:     "破坏性变更传播后变为内部",
			source:   ImpactTypeBreaking,
			expected: ImpactTypeInternal,
		},
		{
			name:     "内部变更保持内部",
			source:   ImpactTypeInternal,
			expected: ImpactTypeInternal,
		},
		{
			name:     "增强性变更保持增强",
			source:   ImpactTypeAdditive,
			expected: ImpactTypeAdditive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := propagateImpactType(tt.source)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsMoreSevere 测试影响类型严重程度比较
func TestIsMoreSevere(t *testing.T) {
	tests := []struct {
		name     string
		type1    ImpactType
		type2    ImpactType
		expected bool
	}{
		{
			name:     "breaking > internal",
			type1:    ImpactTypeBreaking,
			type2:    ImpactTypeInternal,
			expected: true,
		},
		{
			name:     "breaking > additive",
			type1:    ImpactTypeBreaking,
			type2:    ImpactTypeAdditive,
			expected: true,
		},
		{
			name:     "internal > additive",
			type1:    ImpactTypeInternal,
			type2:    ImpactTypeAdditive,
			expected: true,
		},
		{
			name:     "additive not > breaking",
			type1:    ImpactTypeAdditive,
			type2:    ImpactTypeBreaking,
			expected: false,
		},
		{
			name:     "same type",
			type1:    ImpactTypeInternal,
			type2:    ImpactTypeInternal,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isMoreSevere(tt.type1, tt.type2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatPath 测试路径格式化
func TestFormatPath(t *testing.T) {
	tests := []struct {
		name string
		path []string
		want string
	}{
		{
			name: "单节点路径",
			path: []string{"Button"},
			want: "Button",
		},
		{
			name: "双节点路径",
			path: []string{"Button", "Input"},
			want: "Button → Input",
		},
		{
			name: "三节点路径",
			path: []string{"A", "B", "C"},
			want: "A → B → C",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPath(tt.path)
			assert.Equal(t, tt.want, result)
		})
	}
}

// TestContains 测试字符串切片包含
func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}

	assert.True(t, contains(slice, "a"))
	assert.True(t, contains(slice, "b"))
	assert.True(t, contains(slice, "c"))
	assert.False(t, contains(slice, "d"))
}

// TestComponentImpact 测试组件影响结构
func TestComponentImpact(t *testing.T) {
	impact := &ComponentImpact{
		ComponentName:    "Button",
		ImpactLevel:      0,
		ChangedSymbols:   []SymbolChange{},
		AffectedSymbols:  make(map[string]bool),
		ChangePaths:      []string{"Button"},
		ImpactType:       ImpactTypeBreaking,
	}

	assert.Equal(t, "Button", impact.ComponentName)
	assert.Equal(t, 0, impact.ImpactLevel)
	assert.Equal(t, ImpactTypeBreaking, impact.ImpactType)
}
