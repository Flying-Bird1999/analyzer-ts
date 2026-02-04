package component_deps_v2

import (
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// 配置加载测试
// =============================================================================

func TestLoadManifest_FileNotFound(t *testing.T) {
	_, err := LoadManifest("/non/existent/path.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "配置文件不存在")
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

// =============================================================================
// 依赖分析测试
// =============================================================================

func TestIsFileInComponent(t *testing.T) {
	manifest := &ComponentManifest{
		Components: []ComponentDefinition{
			{Name: "Button", Entry: "src/Button/index.tsx"},
		},
	}
	analyzer := NewDependencyAnalyzer(manifest)
	compDir := "src/Button"

	// 测试在组件内的文件
	assert.True(t, analyzer.isFileInComponent("src/Button/Button.tsx", compDir))
	assert.True(t, analyzer.isFileInComponent("src/Button/components/ButtonIcon.tsx", compDir))
	assert.True(t, analyzer.isFileInComponent("src/Button/utils/helpers.ts", compDir))

	// 测试不在组件内的文件
	assert.False(t, analyzer.isFileInComponent("src/Input/index.tsx", compDir))
	assert.False(t, analyzer.isFileInComponent("src/ButtonTest/index.tsx", compDir))
}

func TestIsExternalDependency_NpmPackage(t *testing.T) {
	manifest := &ComponentManifest{
		Components: []ComponentDefinition{
			{Name: "Button", Entry: "src/Button/index.tsx"},
		},
	}
	analyzer := NewDependencyAnalyzer(manifest)
	compDir := "src/Button"

	// npm 包应该被视为外部依赖
	importDecl := projectParser.ImportDeclarationResult{
		Source: projectParser.SourceData{
			Type:   "npm",
			NpmPkg: "react",
		},
	}

	assert.True(t, analyzer.isExternalDependency(importDecl, compDir))
}

func TestIsExternalDependency_InternalFile(t *testing.T) {
	manifest := &ComponentManifest{
		Components: []ComponentDefinition{
			{Name: "Button", Entry: "src/Button/index.tsx"},
		},
	}
	analyzer := NewDependencyAnalyzer(manifest)
	compDir := "src/Button"

	// 组件内部文件不应该被视为外部依赖
	importDecl := projectParser.ImportDeclarationResult{
		Source: projectParser.SourceData{
			Type:     "file",
			FilePath: "src/Button/utils/helper.ts",
		},
	}

	assert.False(t, analyzer.isExternalDependency(importDecl, compDir))
}

func TestIsExternalDependency_CrossComponent(t *testing.T) {
	manifest := &ComponentManifest{
		Components: []ComponentDefinition{
			{Name: "Button", Entry: "src/Button/index.tsx"},
			{Name: "Input", Entry: "src/Input/index.tsx"},
		},
	}
	analyzer := NewDependencyAnalyzer(manifest)
	compDir := "src/Button"

	// 跨组件引用应该被视为外部依赖
	importDecl := projectParser.ImportDeclarationResult{
		Source: projectParser.SourceData{
			Type:     "file",
			FilePath: "src/Input/index.tsx",
		},
	}

	assert.True(t, analyzer.isExternalDependency(importDecl, compDir))
}

func TestIsExternalDependency_ExternalFile(t *testing.T) {
	manifest := &ComponentManifest{
		Components: []ComponentDefinition{
			{Name: "Button", Entry: "src/Button/index.tsx"},
		},
	}
	analyzer := NewDependencyAnalyzer(manifest)
	compDir := "src/Button"

	// 不属于任何组件的文件应该被视为外部依赖
	importDecl := projectParser.ImportDeclarationResult{
		Source: projectParser.SourceData{
			Type:     "file",
			FilePath: "src/utils/helper.ts",
		},
	}

	assert.True(t, analyzer.isExternalDependency(importDecl, compDir))
}

// 去重测试
func TestAnalyzeComponent_Dedup(t *testing.T) {
	manifest := &ComponentManifest{
		Components: []ComponentDefinition{
			{Name: "Button", Entry: "src/Button/index.tsx"},
		},
	}
	analyzer := NewDependencyAnalyzer(manifest)

	// 模拟多个文件引用同一个 npm 包和同一个文件
	fileResults := map[string]projectParser.JsFileParserResult{
		"src/Button/Button.tsx": {
			ImportDeclarations: []projectParser.ImportDeclarationResult{
				{
					Source: projectParser.SourceData{Type: "npm", NpmPkg: "react"},
					Raw:    "import React from 'react'",
				},
				{
					Source: projectParser.SourceData{Type: "file", FilePath: "src/Input/index.ts"},
					Raw:    "import { Input } from '../Input'",
				},
			},
		},
		"src/Button/ButtonIcon.tsx": {
			ImportDeclarations: []projectParser.ImportDeclarationResult{
				{
					Source: projectParser.SourceData{Type: "npm", NpmPkg: "react"}, // 重复
					Raw:    "import React from 'react'",
				},
				{
					Source: projectParser.SourceData{Type: "file", FilePath: "src/Input/index.ts"}, // 重复
					Raw:    "import { Input } from '../Input'",
				},
				{
					Source: projectParser.SourceData{Type: "npm", NpmPkg: "lodash"},
					Raw:    "import { debounce } from 'lodash'",
				},
			},
		},
	}

	comp := &ComponentDefinition{Name: "Button", Entry: "src/Button/index.tsx"}
	deps := analyzer.AnalyzeComponent(comp, fileResults)

	// 应该去重，只有 3 个依赖：react、Input/index.ts、lodash
	assert.Len(t, deps, 3)

	// 验证去重后的结果
	npmDeps := lo.Filter(deps, func(d projectParser.ImportDeclarationResult, _ int) bool {
		return d.Source.Type == "npm"
	})
	assert.Len(t, npmDeps, 2)

	fileDeps := lo.Filter(deps, func(d projectParser.ImportDeclarationResult, _ int) bool {
		return d.Source.Type == "file"
	})
	assert.Len(t, fileDeps, 1)
}

// =============================================================================
// 结果结构测试
// =============================================================================

func TestComponentDepsV2Result_Name(t *testing.T) {
	result := &ComponentDepsV2Result{}
	assert.Equal(t, "component-deps-v2", result.Name())
}

func TestComponentDepsV2Result_Summary(t *testing.T) {
	result := &ComponentDepsV2Result{
		Meta: Meta{ComponentCount: 2},
		Components: map[string]ComponentInfo{
			"Button": {Dependencies: []projectParser.ImportDeclarationResult{{}, {}}},
			"Input":  {Dependencies: []projectParser.ImportDeclarationResult{{}}},
		},
	}

	summary := result.Summary()
	assert.Contains(t, summary, "2 个组件")
	assert.Contains(t, summary, "3 条外部依赖")
}

func TestComponentDepsV2Result_ToJSON(t *testing.T) {
	result := &ComponentDepsV2Result{
		Meta: Meta{ComponentCount: 1},
		Components: map[string]ComponentInfo{
			"Button": {
				Name:         "Button",
				Entry:        "src/Button/index.tsx",
				Dependencies: []projectParser.ImportDeclarationResult{},
			},
		},
	}

	data, err := result.ToJSON(false)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "Button")
	assert.Contains(t, string(data), "src/Button/index.tsx")
}

func TestComponentDepsV2Result_ToConsole(t *testing.T) {
	result := &ComponentDepsV2Result{
		Meta: Meta{ComponentCount: 1},
		Components: map[string]ComponentInfo{
			"Button": {
				Name:         "Button",
				Entry:        "src/Button/index.tsx",
				Dependencies: []projectParser.ImportDeclarationResult{},
			},
		},
	}

	output := result.ToConsole()
	assert.Contains(t, output, "组件依赖分析报告")
	assert.Contains(t, output, "Button")
	assert.Contains(t, output, "src/Button/index.tsx")
	assert.Contains(t, output, "外部依赖: 无")
}
