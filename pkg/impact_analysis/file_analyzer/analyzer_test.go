// Package file_analyzer 文件级影响分析测试
package file_analyzer

import (
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/Flying-Bird1999/analyzer-ts/pkg/symbol_analysis"
)

// =============================================================================
// SymbolPropagator 测试
// =============================================================================

func TestSymbolPropagator_Propagate(t *testing.T) {
	// 构造测试用的解析结果
	parsingResult := &projectParser.ProjectParserResult{
		Js_Data: make(map[string]projectParser.JsFileParserResult),
	}

	// 模拟文件结构：
	// src/App.tsx imports: Button (default) from Button.tsx
	// src/components/Button/Button.tsx exports: default Button

	// src/components/Button/Button.tsx - 导出 Button
	parsingResult.Js_Data["/project/src/components/Button/Button.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{
			{
				ExportModules: []projectParser.ExportModule{
					{Identifier: "Button", Type: "default"},
				},
			},
		},
	}

	// src/App.tsx - 导入 Button
	parsingResult.Js_Data["/project/src/App.tsx"] = projectParser.JsFileParserResult{
		ImportDeclarations: []projectParser.ImportDeclarationResult{
			{
				Source: projectParser.SourceData{
					FilePath: "/project/src/components/Button/Button.tsx",
				},
				ImportModules: []projectParser.ImportModule{
					{Identifier: "Button", Type: "default"},
				},
			},
		},
	}

	// 构造被修改的符号（Button 被修改）
	changedSymbols := []ChangedSymbol{
		{
			Name:       "Button",
			FilePath:   "/project/src/components/Button/Button.tsx",
			ExportType: symbol_analysis.ExportTypeDefault,
		},
	}

	// 先检查符号索引是否正确构建
	propagator := NewSymbolPropagator(parsingResult)
	symbolIndex := propagator.buildSymbolIndex(changedSymbols)

	// 调试：检查符号索引
	t.Logf("Changed symbols in index: %d", len(symbolIndex.ChangedSymbols))
	for key, sym := range symbolIndex.ChangedSymbols {
		t.Logf("  Key: %s, Name: %s, FilePath: %s, ExportType: %v",
			key, sym.Name, sym.FilePath, sym.ExportType)
	}

	t.Logf("File exports in index: %d", len(symbolIndex.FileExports))
	for filePath, exports := range symbolIndex.FileExports {
		t.Logf("  File: %s", filePath)
		for _, exp := range exports {
			t.Logf("    Export: %s (type: %v)", exp.Name, exp.ExportType)
		}
	}

	// 检查 App.tsx 的导入
	appFile := parsingResult.Js_Data["/project/src/App.tsx"]
	t.Logf("App.tsx import declarations: %d", len(appFile.ImportDeclarations))
	for i, imp := range appFile.ImportDeclarations {
		t.Logf("  Import %d: Source=%s, Modules=%d",
			i, imp.Source.FilePath, len(imp.ImportModules))
		for j, mod := range imp.ImportModules {
			t.Logf("    Module %d: Identifier=%s, Type=%s",
				j, mod.Identifier, mod.Type)
		}
	}

	// 执行传播
	result := propagator.Propagate(changedSymbols, nil)

	// 验证结果
	if len(result.Direct) == 0 {
		t.Fatal("Direct changes should not be empty")
	}

	if len(result.Indirect) == 0 {
		t.Fatal("Indirect impacts should not be empty")
	}

	// 验证 App.tsx 被影响
	appImpact, exists := result.Indirect["/project/src/App.tsx"]
	if !exists {
		t.Fatal("App.tsx should be impacted")
	}

	if appImpact.ImpactLevel != 1 {
		t.Errorf("App.tsx should have impact level 1, got %d", appImpact.ImpactLevel)
	}

	if appImpact.SymbolCount != 1 {
		t.Errorf("App.tsx should have 1 impacted symbol, got %d", appImpact.SymbolCount)
	}
}

func TestSymbolPropagator_MultipleImpacts(t *testing.T) {
	// 测试多个符号同时被修改的场景
	parsingResult := &projectParser.ProjectParserResult{
		Js_Data: make(map[string]projectParser.JsFileParserResult),
	}

	// components.tsx exports: Button, Input
	parsingResult.Js_Data["/project/components.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{
			{
				ExportModules: []projectParser.ExportModule{
					{Identifier: "Button", Type: "default"},
					{Identifier: "Input", Type: "default"},
				},
			},
		},
	}

	// App.tsx imports: Button, Input
	parsingResult.Js_Data["/project/App.tsx"] = projectParser.JsFileParserResult{
		ImportDeclarations: []projectParser.ImportDeclarationResult{
			{
				Source: projectParser.SourceData{
					FilePath: "/project/components.tsx",
				},
				ImportModules: []projectParser.ImportModule{
					{Identifier: "Button", Type: "default"},
					{Identifier: "Input", Type: "default"},
				},
			},
		},
	}

	// 两个符号都被修改
	changedSymbols := []ChangedSymbol{
		{
			Name:       "Button",
			FilePath:   "/project/components.tsx",
			ExportType: symbol_analysis.ExportTypeDefault,
		},
		{
			Name:       "Input",
			FilePath:   "/project/components.tsx",
			ExportType: symbol_analysis.ExportTypeDefault,
		},
	}

	propagator := NewSymbolPropagator(parsingResult)
	result := propagator.Propagate(changedSymbols, nil)

	// 验证 App.tsx 被两个符号影响
	appImpact, exists := result.Indirect["/project/App.tsx"]
	if !exists {
		t.Fatal("App.tsx should be impacted")
	}

	if appImpact.SymbolCount != 2 {
		t.Errorf("App.tsx should have 2 impacted symbols, got %d", appImpact.SymbolCount)
	}
}

func TestSymbolPropagator_ExportImportMatch(t *testing.T) {
	// 测试导出/导入类型匹配
	tests := []struct {
		name         string
		exportType   symbol_analysis.ExportType
		importType   string
		shouldMatch  bool
	}{
		{
			name:        "default export matches default import",
			exportType:  symbol_analysis.ExportTypeDefault,
			importType:  "default",
			shouldMatch: true,
		},
		{
			name:        "named export matches named import",
			exportType:  symbol_analysis.ExportTypeNamed,
			importType:  "named",
			shouldMatch: true,
		},
		{
			name:        "namespace export matches namespace import",
			exportType:  symbol_analysis.ExportTypeNamespace,
			importType:  "namespace",
			shouldMatch: true,
		},
		{
			name:        "default export does not match named import",
			exportType:  symbol_analysis.ExportTypeDefault,
			importType:  "named",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			propagator := &SymbolPropagator{}

			// 构造一个简单的 importDecl 用于测试
			importDecl := projectParser.ImportDeclarationResult{}

			result := propagator.isExportImportMatch(tt.exportType, tt.importType, importDecl)

			if result != tt.shouldMatch {
				t.Errorf("isExportImportMatch() = %v, want %v", result, tt.shouldMatch)
			}
		})
	}
}

func TestSymbolPropagator_NoImpactForUnmatchedImports(t *testing.T) {
	// 测试不匹配的导入不应该被影响
	parsingResult := &projectParser.ProjectParserResult{
		Js_Data: make(map[string]projectParser.JsFileParserResult),
	}

	// Button.tsx exports: Button
	parsingResult.Js_Data["/project/Button.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{
			{
				ExportModules: []projectParser.ExportModule{
					{Identifier: "Button", Type: "default"},
				},
			},
		},
	}

	// App.tsx imports: Input (不是 Button)
	parsingResult.Js_Data["/project/App.tsx"] = projectParser.JsFileParserResult{
		ImportDeclarations: []projectParser.ImportDeclarationResult{
			{
				Source: projectParser.SourceData{
					FilePath: "/project/Button.tsx",
				},
				ImportModules: []projectParser.ImportModule{
					{Identifier: "Input", Type: "default"}, // 不同的符号
				},
			},
		},
	}

	// Button 被修改
	changedSymbols := []ChangedSymbol{
		{
			Name:       "Button",
			FilePath:   "/project/Button.tsx",
			ExportType: symbol_analysis.ExportTypeDefault,
		},
	}

	propagator := NewSymbolPropagator(parsingResult)
	result := propagator.Propagate(changedSymbols, nil)

	// App.tsx 不应该被影响（因为它导入的是 Input，不是 Button）
	if _, exists := result.Indirect["/project/App.tsx"]; exists {
		t.Error("App.tsx should not be impacted (imports Input, not Button)")
	}
}

func TestSymbolPropagator_TransitiveImpact(t *testing.T) {
	// 测试传播影响：Button.tsx -> App.tsx -> Main.tsx
	parsingResult := &projectParser.ProjectParserResult{
		Js_Data: make(map[string]projectParser.JsFileParserResult),
	}

	// Button.tsx exports: Button
	parsingResult.Js_Data["/project/Button.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{
			{
				ExportModules: []projectParser.ExportModule{
					{Identifier: "Button", Type: "default"},
				},
			},
		},
	}

	// App.tsx imports Button from Button.tsx, exports: default App
	parsingResult.Js_Data["/project/App.tsx"] = projectParser.JsFileParserResult{
		ImportDeclarations: []projectParser.ImportDeclarationResult{
			{
				Source: projectParser.SourceData{
					FilePath: "/project/Button.tsx",
				},
				ImportModules: []projectParser.ImportModule{
					{Identifier: "Button", Type: "default"},
				},
			},
		},
		ExportDeclarations: []projectParser.ExportDeclarationResult{
			{
				ExportModules: []projectParser.ExportModule{
					{Identifier: "App", Type: "default"},
				},
			},
		},
	}

	// Main.tsx imports: App from App.tsx
	parsingResult.Js_Data["/project/Main.tsx"] = projectParser.JsFileParserResult{
		ImportDeclarations: []projectParser.ImportDeclarationResult{
			{
				Source: projectParser.SourceData{
					FilePath: "/project/App.tsx",
				},
				ImportModules: []projectParser.ImportModule{
					{Identifier: "App", Type: "default"},
				},
			},
		},
	}

	// Button 被修改
	changedSymbols := []ChangedSymbol{
		{
			Name:       "Button",
			FilePath:   "/project/Button.tsx",
			ExportType: symbol_analysis.ExportTypeDefault,
		},
	}

	propagator := NewSymbolPropagator(parsingResult)
	result := propagator.Propagate(changedSymbols, nil)

	// App.tsx 应该被影响（层级 1）
	appImpact, exists := result.Indirect["/project/App.tsx"]
	if !exists {
		t.Fatal("App.tsx should be impacted (level 1)")
	}
	if appImpact.ImpactLevel != 1 {
		t.Errorf("App.tsx should have impact level 1, got %d", appImpact.ImpactLevel)
	}

	// Main.tsx 应该被影响（层级 2）
	mainImpact, exists := result.Indirect["/project/Main.tsx"]
	if !exists {
		t.Fatal("Main.tsx should be impacted (level 2)")
	}
	if mainImpact.ImpactLevel != 2 {
		t.Errorf("Main.tsx should have impact level 2, got %d", mainImpact.ImpactLevel)
	}
}
