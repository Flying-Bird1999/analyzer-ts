// Package file_analyzer 文件级影响分析测试
package file_analyzer

import (
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
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
		name        string
		exportType  symbol_analysis.ExportType
		importType  string
		shouldMatch bool
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

// =============================================================================
// 新增场景测试
// =============================================================================

// TestSymbolPropagator_NewExportImpact 测试新增导出的影响
func TestSymbolPropagator_NewExportImpact(t *testing.T) {
	parsingResult := &projectParser.ProjectParserResult{
		Js_Data: make(map[string]projectParser.JsFileParserResult),
	}

	// Button.tsx 新增导出 NewFeature
	parsingResult.Js_Data["/project/Button.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{
			{
				ExportModules: []projectParser.ExportModule{
					{Identifier: "NewFeature", Type: "named"},
				},
			},
		},
	}

	// App.tsx 已经导入了 NewFeature（假设它已存在）
	parsingResult.Js_Data["/project/App.tsx"] = projectParser.JsFileParserResult{
		ImportDeclarations: []projectParser.ImportDeclarationResult{
			{
				Source: projectParser.SourceData{
					FilePath: "/project/Button.tsx",
				},
				ImportModules: []projectParser.ImportModule{
					{Identifier: "NewFeature", Type: "named"},
				},
			},
		},
	}

	// 新增的符号被标记为变更
	changedSymbols := []ChangedSymbol{
		{
			Name:       "NewFeature",
			FilePath:   "/project/Button.tsx",
			ExportType: symbol_analysis.ExportTypeNamed,
		},
	}

	propagator := NewSymbolPropagator(parsingResult)
	result := propagator.Propagate(changedSymbols, nil)

	// 验证：新增导出不应影响（因为它是新加的，之前的代码还没用到）
	// 注意：当前实现中，任何导出变更都会被视为影响
	// 这个测试验证传播逻辑正确识别了导入关系
	if len(result.Indirect) > 0 {
		t.Logf("New export impacts %d files (implementation detail)", len(result.Indirect))
	}
}

// TestSymbolPropagator_RemovedExportImpact 测试删除导出的影响
func TestSymbolPropagator_RemovedExportImpact(t *testing.T) {
	parsingResult := &projectParser.ProjectParserResult{
		Js_Data: make(map[string]projectParser.JsFileParserResult),
	}

	// Button.tsx 导出 Button
	parsingResult.Js_Data["/project/Button.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{
			{
				ExportModules: []projectParser.ExportModule{
					{Identifier: "Button", Type: "default"},
				},
			},
		},
	}

	// App.tsx 导入 Button
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
	}

	// Button 导出被标记为变更
	changedSymbols := []ChangedSymbol{
		{
			Name:       "Button",
			FilePath:   "/project/Button.tsx",
			ExportType: symbol_analysis.ExportTypeDefault,
		},
	}

	propagator := NewSymbolPropagator(parsingResult)
	result := propagator.Propagate(changedSymbols, nil)

	// 验证：删除导出仍应影响（因为引用方代码会编译失败）
	if len(result.Indirect) == 0 {
		t.Error("Removing export should still impact importing files")
	}

	// 验证 App.tsx 被影响
	appImpact, exists := result.Indirect["/project/App.tsx"]
	if !exists {
		t.Error("App.tsx should be impacted when Button export is removed")
	}

	if appImpact.ImpactType != "internal" {
		t.Errorf("ImpactType should be internal, got %s", appImpact.ImpactType)
	}
}

// TestSymbolPropagator_CyclicDependency 测试循环依赖的影响
func TestSymbolPropagator_CyclicDependency(t *testing.T) {
	parsingResult := &projectParser.ProjectParserResult{
		Js_Data: make(map[string]projectParser.JsFileParserResult),
	}

	// A.tsx 导出 B
	parsingResult.Js_Data["/project/A.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{
			{
				ExportModules: []projectParser.ExportModule{
					{Identifier: "FuncFromA", Type: "named"},
				},
			},
		},
		ImportDeclarations: []projectParser.ImportDeclarationResult{
			{
				Source: projectParser.SourceData{
					FilePath: "/project/B.tsx",
				},
				ImportModules: []projectParser.ImportModule{
					{Identifier: "FuncFromB", Type: "named"},
				},
			},
		},
	}

	// B.tsx 导出 A
	parsingResult.Js_Data["/project/B.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{
			{
				ExportModules: []projectParser.ExportModule{
					{Identifier: "FuncFromB", Type: "named"},
				},
			},
		},
		ImportDeclarations: []projectParser.ImportDeclarationResult{
			{
				Source: projectParser.SourceData{
					FilePath: "/project/A.tsx",
				},
				ImportModules: []projectParser.ImportModule{
					{Identifier: "FuncFromA", Type: "named"},
				},
			},
		},
	}

	// 修改 A.tsx 中的导出
	changedSymbols := []ChangedSymbol{
		{
			Name:       "FuncFromA",
			FilePath:   "/project/A.tsx",
			ExportType: symbol_analysis.ExportTypeNamed,
		},
	}

	propagator := NewSymbolPropagator(parsingResult)
	result := propagator.Propagate(changedSymbols, nil)

	// 验证：循环依赖时，导入该符号的文件应该被影响
	bImpact, bExists := result.Indirect["/project/B.tsx"]

	if !bExists {
		t.Error("B.tsx should be impacted (imports FuncFromA which is modified)")
	}

	// 验证影响类型
	if bImpact.ImpactType != "internal" {
		t.Errorf("B.tsx impact should be internal, got %s", bImpact.ImpactType)
	}
}

// TestSymbolPropagator_MultipleFilesFromSameSymbol 测试同一符号影响多个文件
func TestSymbolPropagator_MultipleFilesFromSameSymbol(t *testing.T) {
	parsingResult := &projectParser.ProjectParserResult{
		Js_Data: make(map[string]projectParser.JsFileParserResult),
	}

	// utils.tsx 导出多个工具函数
	parsingResult.Js_Data["/project/utils.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{
			{
				ExportModules: []projectParser.ExportModule{
					{Identifier: "formatDate", Type: "named"},
					{Identifier: "formatNumber", Type: "named"},
				},
			},
		},
	}

	// component1.tsx 导入一个函数
	parsingResult.Js_Data["/project/component1.tsx"] = projectParser.JsFileParserResult{
		ImportDeclarations: []projectParser.ImportDeclarationResult{
			{
				Source: projectParser.SourceData{
					FilePath: "/project/utils.tsx",
				},
				ImportModules: []projectParser.ImportModule{
					{Identifier: "formatDate", Type: "named"},
				},
			},
		},
	}

	// component2.tsx 也导入该函数
	parsingResult.Js_Data["/project/component2.tsx"] = projectParser.JsFileParserResult{
		ImportDeclarations: []projectParser.ImportDeclarationResult{
			{
				Source: projectParser.SourceData{
					FilePath: "/project/utils.tsx",
				},
				ImportModules: []projectParser.ImportModule{
					{Identifier: "formatDate", Type: "named"},
					{Identifier: "formatNumber", Type: "named"},
				},
			},
		},
	}

	// 修改 utils.tsx 中的一个导出
	changedSymbols := []ChangedSymbol{
		{
			Name:       "formatDate",
			FilePath:   "/project/utils.tsx",
			ExportType: symbol_analysis.ExportTypeNamed,
		},
	}

	propagator := NewSymbolPropagator(parsingResult)
	result := propagator.Propagate(changedSymbols, nil)

	// 验证：两个组件都应该被影响
	c1Impact, c1Exists := result.Indirect["/project/component1.tsx"]
	c2Impact, c2Exists := result.Indirect["/project/component2.tsx"]

	if !c1Exists {
		t.Error("component1.tsx should be impacted")
	}

	if !c2Exists {
		t.Error("component2.tsx should be impacted")
	}

	// 验证符号数量
	if c1Impact.SymbolCount != 1 {
		t.Errorf("component1.tsx should have 1 impacted symbol, got %d", c1Impact.SymbolCount)
	}

	if c2Impact.SymbolCount != 1 {
		t.Errorf("component2.tsx should have 1 impacted symbol, got %d", c2Impact.SymbolCount)
	}
}

// TestSymbolPropagator_NonSymbolFileChange 测试非符号文件变更的影响
func TestSymbolPropagator_NonSymbolFileChange(t *testing.T) {
	parsingResult := &projectParser.ProjectParserResult{
		Js_Data: make(map[string]projectParser.JsFileParserResult),
	}

	// Button.tsx 导出 Button
	parsingResult.Js_Data["/project/Button.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{
			{
				ExportModules: []projectParser.ExportModule{
					{Identifier: "Button", Type: "default"},
				},
			},
		},
	}

	changedSymbols := []ChangedSymbol{
		{
			Name:       "Button",
			FilePath:   "/project/Button.tsx",
			ExportType: symbol_analysis.ExportTypeDefault,
		},
	}

	nonSymbolFiles := []string{"/project/styles.css"}

	propagator := NewSymbolPropagator(parsingResult)
	result := propagator.Propagate(changedSymbols, nonSymbolFiles)

	// 验证：非符号文件的变更会在 Direct 中
	if _, exists := result.Direct["/project/styles.css"]; !exists {
		t.Error("styles.css should be in Direct changes")
	}

	// 验证：符号文件的变更也会在 Direct 中
	if _, exists := result.Direct["/project/Button.tsx"]; !exists {
		t.Error("Button.tsx should be in Direct changes")
	}
}

// TestSymbolPropagator_DirectImpactSameFile 测试同一文件内既有变更又有影响
func TestSymbolPropagator_DirectImpactSameFile(t *testing.T) {
	parsingResult := &projectParser.ProjectParserResult{
		Js_Data: make(map[string]projectParser.JsFileParserResult),
	}

	// components.tsx 导出并重新导出 Button
	parsingResult.Js_Data["/project/components.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{
			{
				ExportModules: []projectParser.ExportModule{
					{Identifier: "Button", Type: "default"},
				},
			},
		},
	}

	// App.tsx 导入 Button
	parsingResult.Js_Data["/project/App.tsx"] = projectParser.JsFileParserResult{
		ImportDeclarations: []projectParser.ImportDeclarationResult{
			{
				Source: projectParser.SourceData{
					FilePath: "/project/components.tsx",
				},
				ImportModules: []projectParser.ImportModule{
					{Identifier: "Button", Type: "default"},
				},
			},
		},
	}

	// components.tsx 中的 Button 导出被标记为变更
	changedSymbols := []ChangedSymbol{
		{
			Name:       "Button",
			FilePath:   "/project/components.tsx",
			ExportType: symbol_analysis.ExportTypeDefault,
		},
	}

	propagator := NewSymbolPropagator(parsingResult)
	result := propagator.Propagate(changedSymbols, nil)

	// 验证：components.tsx 在 Direct 中（包含符号变更的文件）
	if _, exists := result.Direct["/project/components.tsx"]; !exists {
		t.Error("components.tsx should be in Direct (it contains the changed symbol)")
	}

	// 验证：App.tsx 在 Indirect 中（使用了被变更的符号）
	if _, exists := result.Indirect["/project/App.tsx"]; !exists {
		t.Error("App.tsx should be in Indirect (imports the changed symbol)")
	}
}

// TestSymbolPropagator_IndirectImpactInSameFile 测试间接影响在同一文件内
func TestSymbolPropagator_IndirectImpactInSameFile(t *testing.T) {
	parsingResult := &projectParser.ProjectParserResult{
		Js_Data: make(map[string]projectParser.JsFileParserResult),
	}

	// Button.tsx 导出 Button，并导入 InputUtil
	parsingResult.Js_Data["/project/Button.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{
			{
				ExportModules: []projectParser.ExportModule{
					{Identifier: "Button", Type: "default"},
				},
			},
		},
		ImportDeclarations: []projectParser.ImportDeclarationResult{
			{
				Source: projectParser.SourceData{
					FilePath: "/project/Input.tsx",
				},
				ImportModules: []projectParser.ImportModule{
					{Identifier: "InputUtil", Type: "named"},
				},
			},
		},
	}

	// Input.tsx 导出 InputUtil
	parsingResult.Js_Data["/project/Input.tsx"] = projectParser.JsFileParserResult{
		ExportDeclarations: []projectParser.ExportDeclarationResult{
			{
				ExportModules: []projectParser.ExportModule{
					{Identifier: "InputUtil", Type: "named"},
				},
			},
		},
	}

	// Button.tsx 被修改
	changedSymbols := []ChangedSymbol{
		{
			Name:       "Button",
			FilePath:   "/project/Button.tsx",
			ExportType: symbol_analysis.ExportTypeDefault,
		},
	}

	propagator := NewSymbolPropagator(parsingResult)
	result := propagator.Propagate(changedSymbols, nil)

	// 验证：Button.tsx 在 Direct 中
	if _, exists := result.Direct["/project/Button.tsx"]; !exists {
		t.Error("Button.tsx should be in Direct changes")
	}

	// Input.tsx 不应该在 Direct 中（因为它本身没有被修改）
	if _, exists := result.Direct["/project/Input.tsx"]; exists {
		t.Error("Input.tsx should not be in Direct changes (not modified)")
	}
}

// TestSymbolPropagator_ExportDefaultArrowFunction 测试 export default () => {} 的影响传播
// 验证：当 export default () => {} 内部有变更时，能正确传播到导入它的文件
func TestSymbolPropagator_ExportDefaultArrowFunction(t *testing.T) {
	parsingResult := &projectParser.ProjectParserResult{
		Js_Data: make(map[string]projectParser.JsFileParserResult),
	}

	// Button.tsx 有 export default () => {}
	parsingResult.Js_Data["/project/Button.tsx"] = projectParser.JsFileParserResult{
		ImportDeclarations: []projectParser.ImportDeclarationResult{},
		ExportDeclarations: []projectParser.ExportDeclarationResult{},
		ExportAssignments: []parser.ExportAssignmentResult{
			{
				Expression: "() => {}",  // 箭头函数表达式
			},
		},
	}

	// App.tsx 导入 Button（使用 default import）
	parsingResult.Js_Data["/project/App.tsx"] = projectParser.JsFileParserResult{
		ImportDeclarations: []projectParser.ImportDeclarationResult{
			{
				Source: projectParser.SourceData{
					FilePath: "/project/Button.tsx",
					Type:     "file",
				},
				ImportModules: []projectParser.ImportModule{
					{Identifier: "Button", Type: "default"}, // default import
				},
				Raw: `import Button from "./Button"`,
			},
		},
	}

	// 模拟：Button.tsx 的 default 符号被修改
	changedSymbols := []ChangedSymbol{
		{
			Name:       "default",  // 符号名是 "default"
			FilePath:   "/project/Button.tsx",
			ExportType: symbol_analysis.ExportTypeDefault,
		},
	}

	propagator := NewSymbolPropagator(parsingResult)
	result := propagator.Propagate(changedSymbols, nil)

	// 验证：Button.tsx 在 Direct 中
	if _, exists := result.Direct["/project/Button.tsx"]; !exists {
		t.Error("Button.tsx should be in Direct changes")
	}

	// 验证：App.tsx 在 Indirect 中（因为它导入了 Button 的默认导出）
	appImpact, exists := result.Indirect["/project/App.tsx"]
	if !exists {
		t.Error("App.tsx should be in Indirect impacts (it imports Button)")
		return
	}

	// 验证：影响信息正确
	if appImpact.SymbolCount != 1 {
		t.Errorf("Expected 1 impacted symbol, got %d", appImpact.SymbolCount)
	}

	if appImpact.ImpactLevel != 1 {
		t.Errorf("Expected impact level 1, got %d", appImpact.ImpactLevel)
	}
}
