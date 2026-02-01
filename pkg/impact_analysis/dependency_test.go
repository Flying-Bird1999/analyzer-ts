// Package impact_analysis 单元测试
package impact_analysis

import (
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 符号依赖映射构建测试
// =============================================================================

// TestSymbolDependencyMap_BuildDependencyMap 测试构建完整的符号依赖映射
func TestSymbolDependencyMap_BuildDependencyMap(t *testing.T) {
	// 模拟一个简单的项目结构：
	// Button.tsx 导出 handleClick 和默认导出 Button
	// Input.tsx 导出 validateInput
	// App.tsx 导入 Button 和 Input
	// Form.tsx 导入 Input

	jsData := make(map[string]projectParser.JsFileParserResult)

	// Button.tsx - 导出符号
	jsData["src/Button.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{
			{
				ExportModules: []projectParser.ExportModule{
					{Identifier: "handleClick", ModuleName: "handleClick", Type: "named"},
				},
			},
		},
		ExportAssignments: []parser.ExportAssignmentResult{
			{Expression: "Button"},
		},
		ImportDeclarations: []projectParser.ImportDeclarationResult{},
	}

	// Input.tsx - 导出符号
	jsData["src/Input.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{
			{
				ExportModules: []projectParser.ExportModule{
					{Identifier: "validateInput", ModuleName: "validateInput", Type: "named"},
				},
			},
		},
		ExportAssignments: []parser.ExportAssignmentResult{
			{Expression: "Input"},
		},
		ImportDeclarations: []projectParser.ImportDeclarationResult{},
	}

	// App.tsx - 导入 Button 和 Input
	jsData["src/App.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{},
		ImportDeclarations: []projectParser.ImportDeclarationResult{
			{
				ImportModules: []projectParser.ImportModule{
					{Identifier: "Button", ImportModule: "Button", Type: "default"},
				},
				Source: projectParser.SourceData{
					FilePath: "src/Button.tsx",
					Type:     "file",
				},
			},
			{
				ImportModules: []projectParser.ImportModule{
					{Identifier: "validateInput", ImportModule: "validateInput", Type: "named"},
				},
				Source: projectParser.SourceData{
					FilePath: "src/Input.tsx",
					Type:     "file",
				},
			},
		},
	}

	// Form.tsx - 导入 Input
	jsData["src/Form.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{},
		ImportDeclarations: []projectParser.ImportDeclarationResult{
			{
				ImportModules: []projectParser.ImportModule{
					{Identifier: "Input", ImportModule: "Input", Type: "default"},
				},
				Source: projectParser.SourceData{
					FilePath: "src/Input.tsx",
					Type:     "file",
				},
			},
		},
	}

	parsingResult := &projectParser.ProjectParserResult{
		Js_Data: jsData,
	}

	// 创建组件清单
	manifest := &ComponentManifest{
		Components: []Component{
			{Name: "Button", Scopes: []string{"src/Button.tsx", "src/components/Button"}},
			{Name: "Input", Scopes: []string{"src/Input.tsx", "src/components/Input"}},
			{Name: "App", Scopes: []string{"src/App.tsx"}},
			{Name: "Form", Scopes: []string{"src/Form.tsx"}},
		},
	}

	// 创建 Matcher 并构建依赖映射
	matcher := NewMatcher(nil, parsingResult, manifest)
	depMap := matcher.BuildSymbolDependencyMap()

	// 验证导出符号
	t.Run("验证导出符号", func(t *testing.T) {
		require.NotEmpty(t, depMap.ComponentExports)

		// Button 组件导出
		buttonExports := depMap.ComponentExports["Button"]
		assert.Len(t, buttonExports, 2)
		assert.Contains(t, getSymbolNames(buttonExports), "handleClick")
		assert.Contains(t, getSymbolNames(buttonExports), "Button")

		// Input 组件导出
		inputExports := depMap.ComponentExports["Input"]
		assert.Len(t, inputExports, 2)
		assert.Contains(t, getSymbolNames(inputExports), "validateInput")
		assert.Contains(t, getSymbolNames(inputExports), "Input")
	})

	// 验证导入关系
	t.Run("验证导入关系", func(t *testing.T) {
		require.NotEmpty(t, depMap.SymbolImports)

		// App 组件的导入
		appImports := depMap.SymbolImports["App"]
		require.Len(t, appImports, 2)

		// 验证从 Button 导入
		buttonImport := findImportRelation(appImports, "Button")
		require.NotNil(t, buttonImport)
		assert.Equal(t, "Button", buttonImport.SourceComponent)
		assert.Len(t, buttonImport.ImportedSymbols, 1)
		assert.Equal(t, "Button", buttonImport.ImportedSymbols[0].Name)
		assert.Equal(t, ExportTypeDefault, buttonImport.ImportedSymbols[0].ExportType)

		// 验证从 Input 导入
		inputImport := findImportRelation(appImports, "Input")
		require.NotNil(t, inputImport)
		assert.Equal(t, "Input", inputImport.SourceComponent)
		assert.Len(t, inputImport.ImportedSymbols, 1)
		assert.Equal(t, "validateInput", inputImport.ImportedSymbols[0].Name)
		assert.Equal(t, ExportTypeNamed, inputImport.ImportedSymbols[0].ExportType)

		// Form 组件的导入
		formImports := depMap.SymbolImports["Form"]
		require.Len(t, formImports, 1)
		assert.Equal(t, "Input", formImports[0].SourceComponent)
	})
}

// TestSymbolDependencyMap_CrossComponentDependency 测试跨组件依赖
func TestSymbolDependencyMap_CrossComponentDependency(t *testing.T) {
	// 测试场景：
	// ComponentA 导出 symbol1
	// ComponentB 导出 symbol2，并导入 ComponentA.symbol1
	// ComponentC 导入 ComponentB.symbol2

	jsData := make(map[string]projectParser.JsFileParserResult)

	// ComponentA.tsx
	jsData["src/ComponentA.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{
			{
				ExportModules: []projectParser.ExportModule{
					{Identifier: "symbol1", ModuleName: "symbol1", Type: "named"},
				},
			},
		},
		ExportAssignments: []parser.ExportAssignmentResult{},
		ImportDeclarations: []projectParser.ImportDeclarationResult{},
	}

	// ComponentB.tsx - 导入 ComponentA
	jsData["src/ComponentB.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{
			{
				ExportModules: []projectParser.ExportModule{
					{Identifier: "symbol2", ModuleName: "symbol2", Type: "named"},
				},
			},
		},
		ExportAssignments: []parser.ExportAssignmentResult{},
		ImportDeclarations: []projectParser.ImportDeclarationResult{
			{
				ImportModules: []projectParser.ImportModule{
					{Identifier: "symbol1", ImportModule: "symbol1", Type: "named"},
				},
				Source: projectParser.SourceData{
					FilePath: "src/ComponentA.tsx",
					Type:     "file",
				},
			},
		},
	}

	// ComponentC.tsx - 导入 ComponentB
	jsData["src/ComponentC.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{},
		ImportDeclarations: []projectParser.ImportDeclarationResult{
			{
				ImportModules: []projectParser.ImportModule{
					{Identifier: "symbol2", ImportModule: "symbol2", Type: "named"},
				},
				Source: projectParser.SourceData{
					FilePath: "src/ComponentB.tsx",
					Type:     "file",
				},
			},
		},
	}

	parsingResult := &projectParser.ProjectParserResult{
		Js_Data: jsData,
	}

	manifest := &ComponentManifest{
		Components: []Component{
			{Name: "ComponentA", Scopes: []string{"src/ComponentA.tsx"}},
			{Name: "ComponentB", Scopes: []string{"src/ComponentB.tsx"}},
			{Name: "ComponentC", Scopes: []string{"src/ComponentC.tsx"}},
		},
	}

	matcher := NewMatcher(nil, parsingResult, manifest)
	depMap := matcher.BuildSymbolDependencyMap()

	// 验证依赖链
	t.Run("验证依赖链", func(t *testing.T) {
		// ComponentB 应该有从 ComponentA 的导入
		bImports := depMap.SymbolImports["ComponentB"]
		require.Len(t, bImports, 1)
		assert.Equal(t, "ComponentA", bImports[0].SourceComponent)

		// ComponentC 应该有从 ComponentB 的导入
		cImports := depMap.SymbolImports["ComponentC"]
		require.Len(t, cImports, 1)
		assert.Equal(t, "ComponentB", cImports[0].SourceComponent)
	})
}

// TestSymbolDependencyMap_NamespaceImport 测试命名空间导入
func TestSymbolDependencyMap_NamespaceImport(t *testing.T) {
	jsData := make(map[string]projectParser.JsFileParserResult)

	// Utils.tsx - 导出多个工具函数
	jsData["src/Utils.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{
			{
				ExportModules: []projectParser.ExportModule{
					{Identifier: "func1", ModuleName: "func1", Type: "named"},
					{Identifier: "func2", ModuleName: "func2", Type: "named"},
				},
			},
		},
		ExportAssignments: []parser.ExportAssignmentResult{},
		ImportDeclarations: []projectParser.ImportDeclarationResult{},
	}

	// App.tsx - 使用命名空间导入
	jsData["src/App.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{},
		ImportDeclarations: []projectParser.ImportDeclarationResult{
			{
				ImportModules: []projectParser.ImportModule{
					{Identifier: "Utils", ImportModule: "*", Type: "namespace"},
				},
				Source: projectParser.SourceData{
					FilePath: "src/Utils.tsx",
					Type:     "file",
				},
			},
		},
	}

	parsingResult := &projectParser.ProjectParserResult{
		Js_Data: jsData,
	}

	manifest := &ComponentManifest{
		Components: []Component{
			{Name: "Utils", Scopes: []string{"src/Utils.tsx"}},
			{Name: "App", Scopes: []string{"src/App.tsx"}},
		},
	}

	matcher := NewMatcher(nil, parsingResult, manifest)
	depMap := matcher.BuildSymbolDependencyMap()

	// 验证命名空间导入
	t.Run("验证命名空间导入", func(t *testing.T) {
		appImports := depMap.SymbolImports["App"]
		require.Len(t, appImports, 1)

		utilsImport := appImports[0]
		assert.Equal(t, "Utils", utilsImport.SourceComponent)
		assert.Equal(t, "namespace", utilsImport.ImportType)
		assert.Len(t, utilsImport.ImportedSymbols, 1)
		assert.Equal(t, "Utils", utilsImport.ImportedSymbols[0].Name)
		assert.Equal(t, ExportTypeNamespace, utilsImport.ImportedSymbols[0].ExportType)
	})
}

// TestSymbolDependencyMap_DefaultImport 测试默认导入
func TestSymbolDependencyMap_DefaultImport(t *testing.T) {
	jsData := make(map[string]projectParser.JsFileParserResult)

	// Button.tsx
	jsData["src/Button.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{},
		ExportAssignments: []parser.ExportAssignmentResult{
			{Expression: "Button"},
		},
		ImportDeclarations: []projectParser.ImportDeclarationResult{},
	}

	// App.tsx - 默认导入
	jsData["src/App.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{},
		ImportDeclarations: []projectParser.ImportDeclarationResult{
			{
				ImportModules: []projectParser.ImportModule{
					{Identifier: "Button", ImportModule: "Button", Type: "default"},
				},
				Source: projectParser.SourceData{
					FilePath: "src/Button.tsx",
					Type:     "file",
				},
			},
		},
	}

	parsingResult := &projectParser.ProjectParserResult{
		Js_Data: jsData,
	}

	manifest := &ComponentManifest{
		Components: []Component{
			{Name: "Button", Scopes: []string{"src/Button.tsx"}},
			{Name: "App", Scopes: []string{"src/App.tsx"}},
		},
	}

	matcher := NewMatcher(nil, parsingResult, manifest)
	depMap := matcher.BuildSymbolDependencyMap()

	// 验证默认导入
	t.Run("验证默认导入", func(t *testing.T) {
		appImports := depMap.SymbolImports["App"]
		require.Len(t, appImports, 1)

		buttonImport := appImports[0]
		assert.Equal(t, "Button", buttonImport.SourceComponent)
		assert.Equal(t, "default", buttonImport.ImportType)
		assert.Len(t, buttonImport.ImportedSymbols, 1)
		assert.Equal(t, "Button", buttonImport.ImportedSymbols[0].Name)
		assert.Equal(t, ExportTypeDefault, buttonImport.ImportedSymbols[0].ExportType)
	})
}

// TestSymbolDependencyMap_NpmPackageImport 测试 NPM 包导入（应被忽略）
func TestSymbolDependencyMap_NpmPackageImport(t *testing.T) {
	jsData := make(map[string]projectParser.JsFileParserResult)

	// App.tsx - 导入 React
	jsData["src/App.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{},
		ImportDeclarations: []projectParser.ImportDeclarationResult{
			{
				ImportModules: []projectParser.ImportModule{
					{Identifier: "React", ImportModule: "React", Type: "default"},
				},
				Source: projectParser.SourceData{
					FilePath: "", // 空路径表示 NPM 包
					NpmPkg:   "react",
					Type:     "npm",
				},
			},
		},
	}

	parsingResult := &projectParser.ProjectParserResult{
		Js_Data: jsData,
	}

	manifest := &ComponentManifest{
		Components: []Component{
			{Name: "App", Scopes: []string{"src/App.tsx"}},
		},
	}

	matcher := NewMatcher(nil, parsingResult, manifest)
	depMap := matcher.BuildSymbolDependencyMap()

	// NPM 包导入应该被忽略（不记录到依赖图中）
	t.Run("NPM 包导入应被忽略", func(t *testing.T) {
		// 对于 NPM 包，sourceComponent 应该是包路径而不是组件名
		// 在当前实现中，FilePath 为空的导入会被跳过
		appImports := depMap.SymbolImports["App"]
		// 由于没有 FilePath，这个导入应该被跳过
		assert.Empty(t, appImports)
	})
}

// =============================================================================
// 辅助函数
// =============================================================================

func getSymbolNames(symbols []SymbolRef) []string {
	names := make([]string, len(symbols))
	for i, s := range symbols {
		names[i] = s.Name
	}
	return names
}

func findImportRelation(relations []ImportRelation, sourceComponent string) *ImportRelation {
	for _, rel := range relations {
		if rel.SourceComponent == sourceComponent {
			return &rel
		}
	}
	return nil
}
