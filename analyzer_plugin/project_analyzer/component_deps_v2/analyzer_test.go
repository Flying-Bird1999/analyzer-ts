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
			{Name: "Button", Type: "component", Path: "src/Button"},
			{Name: "Button", Type: "component", Path: "src/Button2"},
		},
	}

	err := validateManifest(manifest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "组件名称重复")
}

func TestGetComponentByName(t *testing.T) {
	manifest := &ComponentManifest{
		Components: []ComponentDefinition{
			{Name: "Button", Type: "component", Path: "src/Button"},
			{Name: "Input", Type: "component", Path: "src/Input"},
		},
	}

	comp, ok := manifest.GetComponentByName("Button")
	assert.True(t, ok)
	assert.Equal(t, "Button", comp.Name)
	assert.Equal(t, "src/Button", comp.Path)

	_, ok = manifest.GetComponentByName("NonExistent")
	assert.False(t, ok)
}

// =============================================================================
// 依赖分析测试
// =============================================================================

func TestIsFileInComponent(t *testing.T) {
	manifest := &ComponentManifest{
		Components: []ComponentDefinition{
			{Name: "Button", Type: "component", Path: "src/Button"},
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
			{Name: "Button", Type: "component", Path: "src/Button"},
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
			{Name: "Button", Type: "component", Path: "src/Button"},
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
			{Name: "Button", Type: "component", Path: "src/Button"},
			{Name: "Input", Type: "component", Path: "src/Input"},
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
			{Name: "Button", Type: "component", Path: "src/Button"},
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

// 去重测试 - 验证合并 ImportModules 的逻辑
func TestAnalyzeComponent_Dedup(t *testing.T) {
	manifest := &ComponentManifest{
		Components: []ComponentDefinition{
			{Name: "Button", Type: "component", Path: "src/Button"},
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

	comp := &ComponentDefinition{Name: "Button", Type: "component", Path: "src/Button"}
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
				Path:         "src/Button",
				Dependencies: []projectParser.ImportDeclarationResult{},
			},
		},
	}

	data, err := result.ToJSON(false)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "Button")
	assert.Contains(t, string(data), "src/Button")
}

func TestComponentDepsV2Result_ToConsole(t *testing.T) {
	result := &ComponentDepsV2Result{
		Meta: Meta{ComponentCount: 1},
		Components: map[string]ComponentInfo{
			"Button": {
				Name:         "Button",
				Path:         "src/Button",
				Dependencies: []projectParser.ImportDeclarationResult{},
			},
		},
	}

	output := result.ToConsole()
	assert.Contains(t, output, "组件依赖分析报告")
	assert.Contains(t, output, "Button")
	assert.Contains(t, output, "src/Button")
	assert.Contains(t, output, "外部依赖: 无")
}

// =============================================================================
// 依赖分类测试 (新增功能)
// =============================================================================

func TestClassifyDependencies_NpmOnly(t *testing.T) {
	manifest := &ComponentManifest{
		Components: []ComponentDefinition{
			{Name: "Button", Type: "component", Path: "src/Button"},
		},
	}
	analyzer := NewDependencyAnalyzer(manifest)

	dependencies := []projectParser.ImportDeclarationResult{
		{Source: projectParser.SourceData{Type: "npm", NpmPkg: "react"}},
		{Source: projectParser.SourceData{Type: "npm", NpmPkg: "lodash"}},
		{Source: projectParser.SourceData{Type: "npm", NpmPkg: "react"}}, // 重复
	}

	classified := analyzer.ClassifyDependencies(dependencies)

	assert.Len(t, classified.NpmDeps, 2)
	assert.Contains(t, classified.NpmDeps, "react")
	assert.Contains(t, classified.NpmDeps, "lodash")
	assert.Empty(t, classified.ComponentDeps)
}

func TestClassifyDependencies_ComponentOnly(t *testing.T) {
	manifest := &ComponentManifest{
		Components: []ComponentDefinition{
			{Name: "Button", Type: "component", Path: "src/Button"},
			{Name: "Input", Type: "component", Path: "src/Input"},
		},
	}
	analyzer := NewDependencyAnalyzer(manifest)

	dependencies := []projectParser.ImportDeclarationResult{
		{Source: projectParser.SourceData{Type: "file", FilePath: "src/Input/index.tsx"}},
		{Source: projectParser.SourceData{Type: "file", FilePath: "src/Input/types.ts"}},
		{Source: projectParser.SourceData{Type: "file", FilePath: "src/Input/index.tsx"}}, // 重复文件
	}

	classified := analyzer.ClassifyDependencies(dependencies)

	assert.Empty(t, classified.NpmDeps)
	assert.Len(t, classified.ComponentDeps, 1)

	inputDep := classified.ComponentDeps["Input"]
	assert.NotNil(t, inputDep)
	assert.Equal(t, "Input", inputDep.Name)
	assert.Equal(t, "src/Input", inputDep.Path)
	assert.Len(t, inputDep.DepFiles, 2) // 去重后只有 2 个文件
}

func TestClassifyDependencies_Mixed(t *testing.T) {
	manifest := &ComponentManifest{
		Components: []ComponentDefinition{
			{Name: "Button", Type: "component", Path: "src/Button"},
			{Name: "Input", Type: "component", Path: "src/Input"},
		},
	}
	analyzer := NewDependencyAnalyzer(manifest)

	dependencies := []projectParser.ImportDeclarationResult{
		{Source: projectParser.SourceData{Type: "npm", NpmPkg: "react"}},
		{Source: projectParser.SourceData{Type: "npm", NpmPkg: "lodash"}},
		{Source: projectParser.SourceData{Type: "file", FilePath: "src/Input/index.tsx"}},
		{Source: projectParser.SourceData{Type: "file", FilePath: "src/utils/helper.ts"}}, // 不属于任何组件
	}

	classified := analyzer.ClassifyDependencies(dependencies)

	assert.Len(t, classified.NpmDeps, 2)
	assert.Len(t, classified.ComponentDeps, 1) // 只有 Input，utils 不在 manifest 中
	assert.Contains(t, classified.NpmDeps, "react")
	assert.Contains(t, classified.NpmDeps, "lodash")
	assert.NotNil(t, classified.ComponentDeps["Input"])
	assert.Nil(t, classified.ComponentDeps["utils"])
}

func TestFindComponentByFile(t *testing.T) {
	manifest := &ComponentManifest{
		Components: []ComponentDefinition{
			{Name: "Button", Type: "component", Path: "src/components/Button"},
			{Name: "Input", Type: "component", Path: "src/components/Input"},
		},
	}
	analyzer := NewDependencyAnalyzer(manifest)

	// 测试找到组件（相对路径）
	comp := analyzer.findComponentByFile("src/components/Button/index.tsx")
	assert.NotNil(t, comp)
	assert.Equal(t, "Button", comp.Name)

	comp = analyzer.findComponentByFile("src/components/Input/types.ts")
	assert.NotNil(t, comp)
	assert.Equal(t, "Input", comp.Name)

	// 测试绝对路径（包含项目根目录前缀）
	comp = analyzer.findComponentByFile("/project/src/components/Button/index.tsx")
	assert.NotNil(t, comp)
	assert.Equal(t, "Button", comp.Name)

	comp = analyzer.findComponentByFile("/project/src/components/Input/types.ts")
	assert.NotNil(t, comp)
	assert.Equal(t, "Input", comp.Name)

	// 测试找不到组件
	comp = analyzer.findComponentByFile("src/utils/helper.ts")
	assert.Nil(t, comp)

	comp = analyzer.findComponentByFile("src/components/ButtonTest/index.tsx")
	assert.Nil(t, comp)

	comp = analyzer.findComponentByFile("/project/src/utils/helper.ts")
	assert.Nil(t, comp)
}

// TestAnalyzeComponent_MergeImportModules 测试同一来源多次引用时合并 ImportModules
// 验证去重时不会丢失导入模块信息
func TestAnalyzeComponent_MergeImportModules(t *testing.T) {
	manifest := &ComponentManifest{
		Components: []ComponentDefinition{
			{Name: "Button", Type: "component", Path: "src/Button"},
		},
	}
	analyzer := NewDependencyAnalyzer(manifest)

	// 模拟同一文件中多个 import 语句引用同一 npm 包
	fileResults := map[string]projectParser.JsFileParserResult{
		"src/Button/Button.tsx": {
			ImportDeclarations: []projectParser.ImportDeclarationResult{
				{
					Source: projectParser.SourceData{Type: "npm", NpmPkg: "react"},
					Raw:    "import React from 'react'",
					ImportModules: []projectParser.ImportModule{
						{Type: "default", ImportModule: "default", Identifier: "React"},
					},
				},
				{
					Source: projectParser.SourceData{Type: "npm", NpmPkg: "react"},
					Raw:    "import { useState } from 'react'",
					ImportModules: []projectParser.ImportModule{
						{Type: "named", ImportModule: "useState", Identifier: "useState"},
					},
				},
				{
					Source: projectParser.SourceData{Type: "npm", NpmPkg: "react"},
					Raw:    "import { useEffect } from 'react'",
					ImportModules: []projectParser.ImportModule{
						{Type: "named", ImportModule: "useEffect", Identifier: "useEffect"},
					},
				},
			},
		},
	}

	comp := &ComponentDefinition{Name: "Button", Type: "component", Path: "src/Button"}
	deps := analyzer.AnalyzeComponent(comp, fileResults)

	// 应该只有 1 个 npm 包依赖（react），但 ImportModules 应该包含所有导入
	assert.Len(t, deps, 1, "应该只有一个 react 依赖记录")

	reactDep := deps[0]
	assert.Equal(t, "npm", reactDep.Source.Type)
	assert.Equal(t, "react", reactDep.Source.NpmPkg)

	// 关键断言：ImportModules 应该包含所有 3 个导入的模块
	assert.Len(t, reactDep.ImportModules, 3, "ImportModules 应该包含所有导入的模块")

	// 验证具体的模块存在
	moduleNames := lo.Map(reactDep.ImportModules, func(mod projectParser.ImportModule, _ int) string {
		return mod.ImportModule
	})
	assert.Contains(t, moduleNames, "default", "应该包含 default 导入")
	assert.Contains(t, moduleNames, "useState", "应该包含 useState 导入")
	assert.Contains(t, moduleNames, "useEffect", "应该包含 useEffect 导入")

	// Raw 应该合并了所有原始语句
	assert.Contains(t, reactDep.Raw, "import React from 'react'")
	assert.Contains(t, reactDep.Raw, "import { useState } from 'react'")
	assert.Contains(t, reactDep.Raw, "import { useEffect } from 'react'")
}

// TestAnalyzeComponent_MergeCrossFile 测试跨文件合并 ImportModules
func TestAnalyzeComponent_MergeCrossFile(t *testing.T) {
	manifest := &ComponentManifest{
		Components: []ComponentDefinition{
			{Name: "Button", Type: "component", Path: "src/Button"},
		},
	}
	analyzer := NewDependencyAnalyzer(manifest)

	// 模拟不同文件中引用同一外部文件
	fileResults := map[string]projectParser.JsFileParserResult{
		"src/Button/Button.tsx": {
			ImportDeclarations: []projectParser.ImportDeclarationResult{
				{
					Source: projectParser.SourceData{Type: "file", FilePath: "src/utils/helpers.ts"},
					Raw:    "import { helper1 } from '../../utils/helpers'",
					ImportModules: []projectParser.ImportModule{
						{Type: "named", ImportModule: "helper1", Identifier: "helper1"},
					},
				},
			},
		},
		"src/Button/ButtonIcon.tsx": {
			ImportDeclarations: []projectParser.ImportDeclarationResult{
				{
					Source: projectParser.SourceData{Type: "file", FilePath: "src/utils/helpers.ts"},
					Raw:    "import { helper2 } from '../../utils/helpers'",
					ImportModules: []projectParser.ImportModule{
						{Type: "named", ImportModule: "helper2", Identifier: "helper2"},
					},
				},
			},
		},
	}

	comp := &ComponentDefinition{Name: "Button", Type: "component", Path: "src/Button"}
	deps := analyzer.AnalyzeComponent(comp, fileResults)

	// 应该只有 1 个文件依赖（helpers.ts），但 ImportModules 应该包含所有导入
	assert.Len(t, deps, 1, "应该只有一个 helpers.ts 依赖记录")

	helpersDep := deps[0]
	assert.Equal(t, "file", helpersDep.Source.Type)
	assert.Equal(t, "src/utils/helpers.ts", helpersDep.Source.FilePath)

	// 关键断言：ImportModules 应该包含两个文件中的所有导入
	assert.Len(t, helpersDep.ImportModules, 2, "ImportModules 应该包含跨文件的所有导入")

	// 验证具体的模块存在
	moduleNames := lo.Map(helpersDep.ImportModules, func(mod projectParser.ImportModule, _ int) string {
		return mod.ImportModule
	})
	assert.Contains(t, moduleNames, "helper1", "应该包含 helper1 导入")
	assert.Contains(t, moduleNames, "helper2", "应该包含 helper2 导入")
}
