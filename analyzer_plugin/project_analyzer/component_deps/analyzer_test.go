package component_deps

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

func TestComponentDepsResult_Name(t *testing.T) {
	result := &ComponentDepsResult{}
	assert.Equal(t, "component-deps", result.Name())
}

func TestComponentDepsResult_Summary(t *testing.T) {
	result := &ComponentDepsResult{
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

func TestComponentDepsResult_ToJSON(t *testing.T) {
	result := &ComponentDepsResult{
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

func TestComponentDepsResult_ToConsole(t *testing.T) {
	result := &ComponentDepsResult{
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

// TestReExportResolver_BasicReExport 测试基本的重导出解析
func TestReExportResolver_BasicReExport(t *testing.T) {
	// 模拟文件结构：
	// src/exports/index.ts - 统一重导出
	//   export { Button } from '../components/Button'
	//   export { Input } from '../components/Input'
	// src/features/Form/index.ts - 使用重导出
	//   import { Button, Input } from '../../exports'

	fileResults := map[string]projectParser.JsFileParserResult{
		"src/exports/index.ts": {
			ExportDeclarations: []projectParser.ExportDeclarationResult{
				{
					ExportModules: []projectParser.ExportModule{
						{ModuleName: "Button", Type: "named"},
					},
					Source: &projectParser.SourceData{
						Type:     "file",
						FilePath: "src/components/Button/index.ts",
					},
				},
				{
					ExportModules: []projectParser.ExportModule{
						{ModuleName: "Input", Type: "named"},
					},
					Source: &projectParser.SourceData{
						Type:     "file",
						FilePath: "src/components/Input/index.ts",
					},
				},
			},
		},
	}

	resolver := NewReExportResolver(fileResults)

	// 测试 IsReExportFile
	assert.True(t, resolver.IsReExportFile("src/exports/index.ts"))

	// 测试 GetReExportMapping
	mapping := resolver.GetReExportMapping("src/exports/index.ts")
	assert.Equal(t, "src/components/Button/index.ts", mapping["Button"])
	assert.Equal(t, "src/components/Input/index.ts", mapping["Input"])

	// 测试 ResolveDependency
	dep := projectParser.ImportDeclarationResult{
		ImportModules: []projectParser.ImportModule{
			{ImportModule: "Button", Type: "named", Identifier: "Button"},
			{ImportModule: "Input", Type: "named", Identifier: "Input"},
		},
		Source: projectParser.SourceData{
			Type:     "file",
			FilePath: "src/exports/index.ts",
		},
		Raw: "import { Button, Input } from '../../exports'",
	}

	resolved := resolver.ResolveDependency(dep)

	// 应该解析为两个独立的依赖
	assert.Len(t, resolved, 2)

	// 验证 Button 依赖
	buttonDep := lo.Filter(resolved, func(d projectParser.ImportDeclarationResult, _ int) bool {
		return d.Source.FilePath == "src/components/Button/index.ts"
	})
	assert.Len(t, buttonDep, 1)
	assert.Len(t, buttonDep[0].ImportModules, 1)
	assert.Equal(t, "Button", buttonDep[0].ImportModules[0].ImportModule)

	// 验证 Input 依赖
	inputDep := lo.Filter(resolved, func(d projectParser.ImportDeclarationResult, _ int) bool {
		return d.Source.FilePath == "src/components/Input/index.ts"
	})
	assert.Len(t, inputDep, 1)
	assert.Len(t, inputDep[0].ImportModules, 1)
	assert.Equal(t, "Input", inputDep[0].ImportModules[0].ImportModule)
}

// TestReExportResolver_ClassifyWithReExport 测试在分类时解析重导出
func TestReExportResolver_ClassifyWithReExport(t *testing.T) {
	manifest := &ComponentManifest{
		Components: []ComponentDefinition{
			{Name: "Form", Type: "component", Path: "src/features/Form"},
			{Name: "Button", Type: "component", Path: "src/components/Button"},
			{Name: "Input", Type: "component", Path: "src/components/Input"},
		},
	}

	fileResults := map[string]projectParser.JsFileParserResult{
		// Form 组件通过 exports 重导出引用 Button 和 Input
		"src/features/Form/index.ts": {
			ImportDeclarations: []projectParser.ImportDeclarationResult{
				{
					ImportModules: []projectParser.ImportModule{
						{ImportModule: "Button", Type: "named", Identifier: "Button"},
						{ImportModule: "Input", Type: "named", Identifier: "Input"},
					},
					Source: projectParser.SourceData{
						Type:     "file",
						FilePath: "src/exports/index.ts",
					},
					Raw: "import { Button, Input } from '../../exports'",
				},
			},
		},
		// exports 统一导出文件
		"src/exports/index.ts": {
			ExportDeclarations: []projectParser.ExportDeclarationResult{
				{
					ExportModules: []projectParser.ExportModule{
						{ModuleName: "Button", Type: "named"},
					},
					Source: &projectParser.SourceData{
						Type:     "file",
						FilePath: "src/components/Button/index.ts",
					},
				},
				{
					ExportModules: []projectParser.ExportModule{
						{ModuleName: "Input", Type: "named"},
					},
					Source: &projectParser.SourceData{
						Type:     "file",
						FilePath: "src/components/Input/index.ts",
					},
				},
			},
		},
	}

	analyzer := NewDependencyAnalyzer(manifest)
	analyzer.fileResults = fileResults
	analyzer.reexportResolver = NewReExportResolver(fileResults)

	// 模拟已经分析好的依赖列表
	dependencies := fileResults["src/features/Form/index.ts"].ImportDeclarations

	// 分类依赖（应该解析重导出）
	classified := analyzer.ClassifyDependencies(dependencies)

	// 验证：应该正确识别为依赖 Button 和 Input 组件
	assert.Len(t, classified.ComponentDeps, 2)
	assert.Contains(t, classified.ComponentDeps, "Button")
	assert.Contains(t, classified.ComponentDeps, "Input")

	// 验证：不应该识别为依赖 exports（因为它不是组件）
	assert.NotContains(t, classified.ComponentDeps, "exports")

	// 验证 DepFiles 应该指向真实的源文件
	assert.Equal(t, []string{"src/components/Button/index.ts"}, classified.ComponentDeps["Button"].DepFiles)
	assert.Equal(t, []string{"src/components/Input/index.ts"}, classified.ComponentDeps["Input"].DepFiles)
}

// TestReExportResolver_ExportStar 测试 export * 语法
func TestReExportResolver_ExportStar(t *testing.T) {
	// 模拟：
	// src/exports/all.ts
	//   export * from '../components/Button'
	// src/components/Button/index.ts
	//   export { ButtonCore } from './ButtonCore'

	fileResults := map[string]projectParser.JsFileParserResult{
		"src/exports/all.ts": {
			ExportDeclarations: []projectParser.ExportDeclarationResult{
				{
					ExportModules: []projectParser.ExportModule{
						{ModuleName: "*", Type: "namespace"},
					},
					Source: &projectParser.SourceData{
						Type:     "file",
						FilePath: "src/components/Button/index.ts",
					},
				},
			},
		},
		"src/components/Button/index.ts": {
			ExportDeclarations: []projectParser.ExportDeclarationResult{
				{
					ExportModules: []projectParser.ExportModule{
						{ModuleName: "ButtonCore", Type: "named"},
					},
					Source: &projectParser.SourceData{
						Type:     "file",
						FilePath: "src/components/Button/ButtonCore.ts",
					},
				},
			},
		},
	}

	resolver := NewReExportResolver(fileResults)

	// 获取重导出映射（应该递归解析）
	mapping := resolver.GetReExportMapping("src/exports/all.ts")

	// 应该递归解析出 ButtonCore 的真实来源
	assert.Equal(t, "src/components/Button/ButtonCore.ts", mapping["ButtonCore"])
}

// TestReExportResolver_GetStats 测试统计信息
func TestReExportResolver_GetStats(t *testing.T) {
	fileResults := map[string]projectParser.JsFileParserResult{
		"src/exports/index.ts": {
			ExportDeclarations: []projectParser.ExportDeclarationResult{
				{
					ExportModules: []projectParser.ExportModule{{ModuleName: "Button", Type: "named"}},
					Source:       &projectParser.SourceData{Type: "file", FilePath: "src/Button/index.ts"},
				},
				{
					ExportModules: []projectParser.ExportModule{{ModuleName: "Input", Type: "named"}},
					Source:       &projectParser.SourceData{Type: "file", FilePath: "src/Input/index.ts"},
				},
			},
		},
		"src/Button/index.ts": {
			ImportDeclarations: []projectParser.ImportDeclarationResult{},
		},
	}

	resolver := NewReExportResolver(fileResults)
	stats := resolver.GetStats()

	assert.Equal(t, 2, stats.TotalFiles)
	assert.Equal(t, 1, stats.ReExportFiles)
	assert.Equal(t, 2, stats.TotalMappings)
}

// TestReExportResolver_EmptyExportDeclarations 测试当 ExportDeclarations 为空时的情况
func TestReExportResolver_EmptyExportDeclarations(t *testing.T) {
	// 模拟：文件有 import，但导出文件的 ExportDeclarations 为空
	fileResults := map[string]projectParser.JsFileParserResult{
		"src/exports/index.ts": {
			ImportDeclarations: []projectParser.ImportDeclarationResult{},
			ExportDeclarations: []projectParser.ExportDeclarationResult{}, // 空导出声明！
		},
	}

	resolver := NewReExportResolver(fileResults)

	// 测试重导出映射
	mapping := resolver.GetReExportMapping("src/exports/index.ts")

	// 应该返回空映射
	assert.Empty(t, mapping)

	// 测试解析依赖
	dep := projectParser.ImportDeclarationResult{
		ImportModules: []projectParser.ImportModule{
			{ImportModule: "Button", Type: "named", Identifier: "Button"},
		},
		Source: projectParser.SourceData{
			Type:     "file",
			FilePath: "src/exports/index.ts",
		},
	}

	resolved := resolver.ResolveDependency(dep)

	// 应该返回原依赖（因为无法解析重导出）
	assert.Len(t, resolved, 1)
	assert.Equal(t, "src/exports/index.ts", resolved[0].Source.FilePath)
}

// TestReExportResolver_EndToEnd 测试完整的重导出解析流程
// 模拟真实场景：组件通过统一导出目录导入其他组件
func TestReExportResolver_EndToEnd(t *testing.T) {
	manifest := &ComponentManifest{
		Components: []ComponentDefinition{
			{Name: "ProDatePicker", Type: "component", Path: "packages/ProDatePicker"},
			{Name: "Button", Type: "component", Path: "packages/atlas/src/core/Button"},
			{Name: "Popcard", Type: "component", Path: "packages/atlas/src/core/Popcard"},
			{Name: "Input", Type: "component", Path: "packages/atlas/src/core/Input"},
		},
	}

	// 模拟完整的文件解析结果
	fileResults := map[string]projectParser.JsFileParserResult{
		// ProDatePicker 组件从 atlas 统一导出导入
		"packages/ProDatePicker/index.ts": {
			ImportDeclarations: []projectParser.ImportDeclarationResult{
				{
					ImportModules: []projectParser.ImportModule{
						{ImportModule: "Button", Type: "named", Identifier: "Button"},
						{ImportModule: "Popcard", Type: "named", Identifier: "Popcard"},
						{ImportModule: "Input", Type: "named", Identifier: "Input"},
					},
					Source: projectParser.SourceData{
						Type:     "file",
						FilePath: "packages/atlas/src/index.ts",
					},
					Raw: "import { Button, Popcard, Input } from '@atlas'",
				},
			},
		},
		// atlas 统一导出文件（关键！）
		"packages/atlas/src/index.ts": {
			ImportDeclarations: []projectParser.ImportDeclarationResult{},
			ExportDeclarations: []projectParser.ExportDeclarationResult{
				{
					ExportModules: []projectParser.ExportModule{
						{ModuleName: "default", Type: "default", Identifier: "Button"},
					},
					Source: &projectParser.SourceData{
						Type:     "file",
						FilePath: "packages/atlas/src/core/Button",
					},
					Raw: "export { default as Button } from './core/Button'",
				},
				{
					ExportModules: []projectParser.ExportModule{
						{ModuleName: "default", Type: "default", Identifier: "Popcard"},
					},
					Source: &projectParser.SourceData{
						Type:     "file",
						FilePath: "packages/atlas/src/core/Popcard",
					},
					Raw: "export { default as Popcard } from './core/Popcard'",
				},
				{
					ExportModules: []projectParser.ExportModule{
						{ModuleName: "default", Type: "default", Identifier: "Input"},
					},
					Source: &projectParser.SourceData{
						Type:     "file",
						FilePath: "packages/atlas/src/core/Input",
					},
					Raw: "export { default as Input } from './core/Input'",
				},
			},
		},
	}

	analyzer := NewDependencyAnalyzer(manifest)
	analyzer.fileResults = fileResults
	analyzer.reexportResolver = NewReExportResolver(fileResults)

	// 模拟已经分析好的依赖列表
	dependencies := fileResults["packages/ProDatePicker/index.ts"].ImportDeclarations

	// 分类依赖（应该解析重导出）
	classified := analyzer.ClassifyDependencies(dependencies)

	// 验证：应该正确识别为依赖 Button、Popcard、Input 组件
	assert.Len(t, classified.ComponentDeps, 3, "应该有 3 个组件依赖")
	assert.Contains(t, classified.ComponentDeps, "Button")
	assert.Contains(t, classified.ComponentDeps, "Popcard")
	assert.Contains(t, classified.ComponentDeps, "Input")

	// 验证 DepFiles 应该指向真实的源文件路径
	assert.Equal(t, []string{"packages/atlas/src/core/Button"}, classified.ComponentDeps["Button"].DepFiles)
	assert.Equal(t, []string{"packages/atlas/src/core/Popcard"}, classified.ComponentDeps["Popcard"].DepFiles)
	assert.Equal(t, []string{"packages/atlas/src/core/Input"}, classified.ComponentDeps["Input"].DepFiles)
}

// TestReExportResolver_WithAlias 测试带别名的重导出
func TestReExportResolver_WithAlias(t *testing.T) {
	// 模拟：
	// src/exports/index.ts
	//   export { default as Popcard } from './core/Popcard'

	fileResults := map[string]projectParser.JsFileParserResult{
		"src/exports/index.ts": {
			ExportDeclarations: []projectParser.ExportDeclarationResult{
				{
					ExportModules: []projectParser.ExportModule{
						{ModuleName: "default", Type: "default", Identifier: "Popcard"},
					},
					Source: &projectParser.SourceData{
						Type:     "file",
						FilePath: "src/core/Popcard.ts",
					},
					Raw: "export { default as Popcard } from './core/Popcard'",
				},
			},
		},
	}

	resolver := NewReExportResolver(fileResults)

	// 测试重导出映射
	mapping := resolver.GetReExportMapping("src/exports/index.ts")

	// 应该使用 "Popcard" 作为 key（外部名称），而不是 "default"
	assert.Equal(t, "src/core/Popcard.ts", mapping["Popcard"])

	// 测试解析依赖
	dep := projectParser.ImportDeclarationResult{
		ImportModules: []projectParser.ImportModule{
			{ImportModule: "Popcard", Type: "default", Identifier: "Popcard"},
		},
		Source: projectParser.SourceData{
			Type:     "file",
			FilePath: "src/exports/index.ts",
		},
		Raw: "import { Popcard } from '../exports'",
	}

	resolved := resolver.ResolveDependency(dep)

	assert.Len(t, resolved, 1)
	assert.Equal(t, "src/core/Popcard.ts", resolved[0].Source.FilePath)
	assert.Equal(t, "Popcard", resolved[0].ImportModules[0].ImportModule)
}

// TestReExportResolver_MixedAlias 测试混合重导出场景
func TestReExportResolver_MixedAlias(t *testing.T) {
	// 模拟：
	// src/exports/index.ts
	//   export { Button } from '../components/Button'
	//   export { default as Popcard } from './core/Popcard'
	//   export { helpers as utils } from '../utils/helpers'

	fileResults := map[string]projectParser.JsFileParserResult{
		"src/exports/index.ts": {
			ExportDeclarations: []projectParser.ExportDeclarationResult{
				{
					ExportModules: []projectParser.ExportModule{
						{ModuleName: "Button", Type: "named", Identifier: "Button"},
					},
					Source: &projectParser.SourceData{Type: "file", FilePath: "src/components/Button/index.ts"},
				},
				{
					ExportModules: []projectParser.ExportModule{
						{ModuleName: "default", Type: "default", Identifier: "Popcard"},
					},
					Source: &projectParser.SourceData{Type: "file", FilePath: "src/core/Popcard.ts"},
				},
				{
					ExportModules: []projectParser.ExportModule{
						{ModuleName: "helpers", Type: "named", Identifier: "utils"},
					},
					Source: &projectParser.SourceData{Type: "file", FilePath: "src/utils/helpers.ts"},
				},
			},
		},
	}

	resolver := NewReExportResolver(fileResults)
	mapping := resolver.GetReExportMapping("src/exports/index.ts")

	// 验证所有映射
	assert.Equal(t, "src/components/Button/index.ts", mapping["Button"])
	assert.Equal(t, "src/core/Popcard.ts", mapping["Popcard"])
	assert.Equal(t, "src/utils/helpers.ts", mapping["utils"])

	// 测试解析混合导入
	dep := projectParser.ImportDeclarationResult{
		ImportModules: []projectParser.ImportModule{
			{ImportModule: "Button", Type: "named", Identifier: "Button"},
			{ImportModule: "Popcard", Type: "default", Identifier: "Popcard"},
			{ImportModule: "utils", Type: "named", Identifier: "utils"},
		},
		Source: projectParser.SourceData{
			Type:     "file",
			FilePath: "src/exports/index.ts",
		},
	}

	resolved := resolver.ResolveDependency(dep)

	// 应该解析为 3 个独立的依赖
	assert.Len(t, resolved, 3)
}
