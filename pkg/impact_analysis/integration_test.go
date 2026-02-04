// Package impact_analysis 集成测试
// 使用 testdata/test_project 真实项目进行验证
package impact_analysis_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/Flying-Bird1999/analyzer-ts/pkg/impact_analysis"
	"github.com/Flying-Bird1999/analyzer-ts/pkg/impact_analysis/component_analyzer"
	"github.com/Flying-Bird1999/analyzer-ts/pkg/impact_analysis/file_analyzer"
	"github.com/Flying-Bird1999/analyzer-ts/pkg/symbol_analysis"
)

// =============================================================================
// 集成测试：使用 testdata/test_project 真实项目
// =============================================================================

func TestIntegration_FileLevelAnalysis(t *testing.T) {
	// 获取测试项目路径
	testProjectPath, err := filepath.Abs("../../testdata/test_project")
	if err != nil {
		t.Fatalf("Failed to get test project path: %v", err)
	}

	// 步骤 1: 使用 projectParser 解析项目
	config := projectParser.NewProjectParserConfig(testProjectPath, nil, false, nil)
	parsingResult := projectParser.NewProjectParserResult(config)
	parsingResult.ProjectParser()

	// 验证解析结果
	if len(parsingResult.Js_Data) == 0 {
		t.Fatal("No JS/TS files parsed")
	}

	t.Logf("Parsed %d JS/TS files", len(parsingResult.Js_Data))

	// 步骤 2: 创建文件分析器
	analyzer := file_analyzer.NewAnalyzer(parsingResult)

	// 步骤 3: 模拟符号变更 - 修改 Button 组件
	buttonPath := filepath.Join(testProjectPath, "src/components/Button/Button.tsx")
	input := &file_analyzer.Input{
		ChangedSymbols: []file_analyzer.ChangedSymbol{
			{
				Name:       "Button",
				FilePath:   buttonPath,
				ExportType: symbol_analysis.ExportTypeDefault,
			},
		},
	}

	// 步骤 4: 执行文件级影响分析
	result, err := analyzer.Analyze(input)
	if err != nil {
		t.Fatalf("Failed to analyze file impact: %v", err)
	}

	// 验证结果
	t.Logf("File Analysis Result:")
	t.Logf("  Total files: %d", result.Meta.TotalFileCount)
	t.Logf("  Changed files: %d", result.Meta.ChangedFileCount)
	t.Logf("  Impacted files: %d", result.Meta.ImpactFileCount)

	// 应该有直接变更的文件
	if len(result.Changes) != 1 {
		t.Errorf("Expected 1 changed file, got %d", len(result.Changes))
	}

	// 应该有受影响的文件（Form.tsx 和 Table.tsx 都依赖 Button）
	if len(result.Impact) == 0 {
		t.Error("Expected impacted files, got none")
	}

	// 打印受影响的文件
	for _, impact := range result.Impact {
		t.Logf("  Impacted: %s (level: %d, symbols: %d)", impact.Path, impact.ImpactLevel, impact.SymbolCount)
	}

	// 验证 Form.tsx 被影响
	formPath := filepath.Join(testProjectPath, "src/components/Form/Form.tsx")
	formFound := false
	for _, impact := range result.Impact {
		if impact.Path == formPath {
			formFound = true
			t.Logf("Form.tsx found with impact level %d", impact.ImpactLevel)
		}
	}
	if !formFound {
		t.Error("Form.tsx should be in impacted files")
	}
}

func TestIntegration_ComponentDependencyGraph(t *testing.T) {
	// 测试组件依赖图构建
	testProjectPath, err := filepath.Abs("../../testdata/test_project")
	if err != nil {
		t.Fatalf("Failed to get test project path: %v", err)
	}

	// 解析项目
	config := projectParser.NewProjectParserConfig(testProjectPath, nil, false, nil)
	parsingResult := projectParser.NewProjectParserResult(config)
	parsingResult.ProjectParser()

	// 加载组件清单
	manifestPath := filepath.Join(testProjectPath, ".analyzer/component-manifest.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("Failed to read component manifest: %v", err)
	}

	var manifest impact_analysis.ComponentManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		t.Fatalf("Failed to parse component manifest: %v", err)
	}

	// 将相对路径转换为绝对路径
	for i := range manifest.Components {
		if !filepath.IsAbs(manifest.Components[i].Entry) {
			manifest.Components[i].Entry = filepath.Join(testProjectPath, manifest.Components[i].Entry)
		}
	}

	t.Logf("Loaded %d components", len(manifest.Components))

	// 构建文件依赖图（使用 GraphBuilder）
	graphBuilder := file_analyzer.NewGraphBuilder(parsingResult)
	fileGraph := graphBuilder.BuildFileDependencyGraph()

	t.Logf("File Dependency Graph has %d entries", len(fileGraph.DepGraph))

	// 创建文件分析器
	analyzer := file_analyzer.NewAnalyzer(parsingResult)

	// 模拟 Button 组件变更
	buttonPath := filepath.Join(testProjectPath, "src/components/Button/Button.tsx")
	input := &file_analyzer.Input{
		ChangedSymbols: []file_analyzer.ChangedSymbol{
			{
				Name:       "Button",
				FilePath:   buttonPath,
				ExportType: symbol_analysis.ExportTypeDefault,
			},
		},
	}

	result, err := analyzer.Analyze(input)
	if err != nil {
		t.Fatalf("Failed to analyze: %v", err)
	}

	t.Logf("File Analysis Result: %d changed files, %d impacted files",
		len(result.Changes), len(result.Impact))

	// 应该有受影响的文件
	if len(result.Impact) == 0 {
		t.Error("Expected impacted files, got none")
	}

	// 打印影响路径
	for _, impact := range result.Impact {
		if len(impact.ChangePaths) > 0 {
			t.Logf("Impact paths for %s:", impact.Path)
			for _, path := range impact.ChangePaths {
				t.Logf("  %s", path)
			}
		}
	}
}

func TestIntegration_FullAnalysis(t *testing.T) {
	// 完整的集成测试：文件级 + 组件级分析
	testProjectPath, err := filepath.Abs("../../testdata/test_project")
	if err != nil {
		t.Fatalf("Failed to get test project path: %v", err)
	}

	// 解析项目
	config := projectParser.NewProjectParserConfig(testProjectPath, nil, false, nil)
	parsingResult := projectParser.NewProjectParserResult(config)
	parsingResult.ProjectParser()

	t.Logf("Project parsed: %d JS/TS files", len(parsingResult.Js_Data))

	// 加载组件清单
	manifestPath := filepath.Join(testProjectPath, ".analyzer/component-manifest.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("Failed to read component manifest: %v", err)
	}

	var manifest impact_analysis.ComponentManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		t.Fatalf("Failed to parse component manifest: %v", err)
	}

	for i := range manifest.Components {
		if !filepath.IsAbs(manifest.Components[i].Entry) {
			manifest.Components[i].Entry = filepath.Join(testProjectPath, manifest.Components[i].Entry)
		}
	}

	t.Logf("Loaded component manifest: %d components", len(manifest.Components))

	// 步骤 1: 文件级分析
	fileAnalyzer := file_analyzer.NewAnalyzer(parsingResult)
	buttonPath := filepath.Join(testProjectPath, "src/components/Button/Button.tsx")
	fileInput := &file_analyzer.Input{
		ChangedSymbols: []file_analyzer.ChangedSymbol{
			{
				Name:       "Button",
				FilePath:   buttonPath,
				ExportType: symbol_analysis.ExportTypeDefault,
			},
		},
	}

	fileResult, err := fileAnalyzer.Analyze(fileInput)
	if err != nil {
		t.Fatalf("Failed to analyze file impact: %v", err)
	}

	t.Logf("File-level analysis:")
	t.Logf("  Changed files: %d", len(fileResult.Changes))
	t.Logf("  Impacted files: %d", len(fileResult.Impact))

	// 步骤 2: 组件级分析
	componentAnalyzer := component_analyzer.NewAnalyzer(&manifest, parsingResult, 10)

	// 构建文件依赖图（使用 GraphBuilder）
	graphBuilder := file_analyzer.NewGraphBuilder(parsingResult)
	fileGraph := graphBuilder.BuildFileDependencyGraph()

	// 将文件级结果转换为代理格式
	fileResultProxy := &component_analyzer.FileAnalysisResultProxy{
		DepGraph:     fileGraph.DepGraph,
		RevDepGraph:  fileGraph.RevDepGraph,
		ExternalDeps: fileGraph.ExternalDeps,
	}

	// 转换 Changes
	for _, change := range fileResult.Changes {
		fileResultProxy.Changes = append(fileResultProxy.Changes, component_analyzer.FileChangeInfoProxy{
			Path:        change.Path,
			ChangeType:  impact_analysis.ChangeTypeModified,
			SymbolCount: change.SymbolCount,
		})
	}

	// 转换 Impact
	for _, impact := range fileResult.Impact {
		fileResultProxy.Impact = append(fileResultProxy.Impact, component_analyzer.FileImpactInfoProxy{
			Path:        impact.Path,
			ImpactLevel: impact_analysis.ImpactLevel(impact.ImpactLevel),
			ChangePaths: impact.ChangePaths,
		})
	}

	compInput := &component_analyzer.Input{
		FileResult: fileResultProxy,
	}

	compResult, err := componentAnalyzer.Analyze(compInput)
	if err != nil {
		t.Fatalf("Failed to analyze component impact: %v", err)
	}

	t.Logf("Component-level analysis:")
	t.Logf("  Changed components: %d", len(compResult.Changes))
	t.Logf("  Impacted components: %d", len(compResult.Impact))

	// 验证 Button 组件被标记为变更
	buttonFound := false
	for _, change := range compResult.Changes {
		if change.Name == "Button" {
			buttonFound = true
			break
		}
	}
	if !buttonFound {
		t.Error("Button component should be in changes")
	}

	// 验证 Form 组件被影响
	formFound := false
	for _, impact := range compResult.Impact {
		if impact.Name == "Form" {
			formFound = true
			t.Logf("Form component found with impact level %d", impact.ImpactLevel)
			break
		}
	}
	if !formFound {
		t.Error("Form component should be in impacts")
	}
}

// import (
// 	"encoding/json"
// 	"os"
// 	"path/filepath"
// 	"strings"
// 	"testing"

// 	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
// 	"github.com/Flying-Bird1999/analyzer-ts/pkg/impact_analysis"
// 	"github.com/Flying-Bird1999/analyzer-ts/pkg/impact_analysis/component_analyzer"
// 	"github.com/Flying-Bird1999/analyzer-ts/pkg/impact_analysis/file_analyzer"
// 	"github.com/Flying-Bird1999/analyzer-ts/pkg/symbol_analysis"
// )

// // =============================================================================
// // 测试辅助函数
// // =============================================================================

// // testProjectContext 包含集成测试所需的上下文
// type testProjectContext struct {
// 	projectPath    string
// 	parsingResult  *projectParser.ProjectParserResult
// 	manifest       *impact_analysis.ComponentManifest
// 	fileAnalyzer   *file_analyzer.Analyzer
// 	componentAnalyzer *component_analyzer.Analyzer
// }

// // setupTestProject 初始化测试项目上下文
// func setupTestProject(t *testing.T) *testProjectContext {
// 	testProjectPath, err := filepath.Abs("../../testdata/test_project")
// 	if err != nil {
// 		t.Fatalf("Failed to get test project path: %v", err)
// 	}

// 	// 解析项目
// 	config := projectParser.NewProjectParserConfig(testProjectPath, nil, false, nil)
// 	parsingResult := projectParser.NewProjectParserResult(config)
// 	parsingResult.ProjectParser()

// 	if len(parsingResult.Js_Data) == 0 {
// 		t.Fatal("No JS/TS files parsed")
// 	}

// 	t.Logf("✓ Parsed %d JS/TS files", len(parsingResult.Js_Data))

// 	// 加载组件清单
// 	manifestPath := filepath.Join(testProjectPath, ".analyzer/component-manifest.json")
// 	manifestData, err := os.ReadFile(manifestPath)
// 	if err != nil {
// 		t.Fatalf("Failed to read component manifest: %v", err)
// 	}

// 	var manifest impact_analysis.ComponentManifest
// 	if err := json.Unmarshal(manifestData, &manifest); err != nil {
// 		t.Fatalf("Failed to parse component manifest: %v", err)
// 	}

// 	// 将相对路径转换为绝对路径
// 	for i := range manifest.Components {
// 		if !filepath.IsAbs(manifest.Components[i].Entry) {
// 			manifest.Components[i].Entry = filepath.Join(testProjectPath, manifest.Components[i].Entry)
// 		}
// 	}

// 	t.Logf("✓ Loaded %d components", len(manifest.Components))

// 	// 创建分析器
// 	fileAnalyzer := file_analyzer.NewAnalyzer(parsingResult)
// 	componentAnalyzer := component_analyzer.NewAnalyzer(&manifest, parsingResult, 10)

// 	return &testProjectContext{
// 		projectPath:       testProjectPath,
// 		parsingResult:     parsingResult,
// 		manifest:          &manifest,
// 		fileAnalyzer:      fileAnalyzer,
// 		componentAnalyzer: componentAnalyzer,
// 	}
// }

// // fileImpactExpectation 定义对文件影响的预期
// type fileImpactExpectation struct {
// 	path        string
// 	impactLevel int // 使用 -1 表示"不检查"
// 	symbolCount int // 使用 -1 表示"不检查"
// }

// // componentImpactExpectation 定义对组件影响的预期
// type componentImpactExpectation struct {
// 	name        string
// 	impactLevel int // 使用 -1 表示"不检查"
// 	symbolCount int // 使用 -1 表示"不检查"
// }

// // verifyFileImpacts 验证文件影响结果是否符合预期
// func verifyFileImpacts(t *testing.T, impacts []file_analyzer.FileImpactInfo, expected []fileImpactExpectation, allowUnexpected bool) {
// 	// 创建预期文件的映射
// 	expectedMap := make(map[string]fileImpactExpectation)
// 	for _, exp := range expected {
// 		expectedMap[exp.path] = exp
// 	}

// 	// 验证所有预期的影响都存在
// 	verifiedPaths := make(map[string]bool)
// 	for _, impact := range impacts {
// 		expected, exists := expectedMap[impact.Path]
// 		if exists {
// 			verifiedPaths[impact.Path] = true

// 			// 验证影响层级
// 			if expected.impactLevel >= 0 && int(impact.ImpactLevel) != expected.impactLevel {
// 				t.Errorf("❌ %s: expected impact level %d, got %d",
// 					impact.Path, expected.impactLevel, impact.ImpactLevel)
// 			}

// 			// 验证符号数量
// 			if expected.symbolCount >= 0 && impact.SymbolCount != expected.symbolCount {
// 				t.Errorf("❌ %s: expected %d symbols, got %d",
// 					impact.Path, expected.symbolCount, impact.SymbolCount)
// 			}
// 		} else if !allowUnexpected {
// 			t.Errorf("❌ Unexpected impacted file: %s (level %d, symbols %d)",
// 				impact.Path, impact.ImpactLevel, impact.SymbolCount)
// 		}
// 	}

// 	// 检查缺失的预期影响
// 	for _, exp := range expected {
// 		if !verifiedPaths[exp.path] {
// 			t.Errorf("❌ Expected impacted file not found: %s (level %d)",
// 				exp.path, exp.impactLevel)
// 		}
// 	}
// }

// // verifyComponentImpacts 验证组件影响结果是否符合预期
// func verifyComponentImpacts(t *testing.T, impacts []component_analyzer.ComponentImpactInfo, expected []componentImpactExpectation, allowUnexpected bool) {
// 	expectedMap := make(map[string]componentImpactExpectation)
// 	for _, exp := range expected {
// 		expectedMap[exp.name] = exp
// 	}

// 	verifiedNames := make(map[string]bool)
// 	for _, impact := range impacts {
// 		expected, exists := expectedMap[impact.Name]
// 		if exists {
// 			verifiedNames[impact.Name] = true

// 			if expected.impactLevel >= 0 && int(impact.ImpactLevel) != expected.impactLevel {
// 				t.Errorf("❌ %s: expected impact level %d, got %d",
// 					impact.Name, expected.impactLevel, impact.ImpactLevel)
// 			}

// 			if expected.symbolCount >= 0 && impact.SymbolCount != expected.symbolCount {
// 				t.Errorf("❌ %s: expected %d symbols, got %d",
// 					impact.Name, expected.symbolCount, impact.SymbolCount)
// 			}
// 		} else if !allowUnexpected {
// 			t.Errorf("❌ Unexpected impacted component: %s (level %d, symbols %d)",
// 				impact.Name, impact.ImpactLevel, impact.SymbolCount)
// 		}
// 	}

// 	for _, exp := range expected {
// 		if !verifiedNames[exp.name] {
// 			t.Errorf("❌ Expected impacted component not found: %s (level %d)",
// 				exp.name, exp.impactLevel)
// 		}
// 	}
// }

// // =============================================================================
// // 场景 1: 基础组件变更的影响分析
// // =============================================================================

// func TestIntegration_Scenario_BasicComponentChange(t *testing.T) {
// 	ctx := setupTestProject(t)

// 	t.Run("File Level - Button Component Change", func(t *testing.T) {
// 		buttonPath := filepath.Join(ctx.projectPath, "src/components/Button/Button.tsx")

// 		input := &file_analyzer.Input{
// 			ChangedSymbols: []file_analyzer.ChangedSymbol{
// 				{
// 					Name:       "Button",
// 					FilePath:   buttonPath,
// 					ExportType: symbol_analysis.ExportTypeDefault,
// 				},
// 			},
// 		}

// 		result, err := ctx.fileAnalyzer.Analyze(input)
// 		if err != nil {
// 			t.Fatalf("Failed to analyze: %v", err)
// 		}

// 		t.Logf("File Analysis Result:")
// 		t.Logf("  Changed: %d files", len(result.Changes))
// 		t.Logf("  Impacted: %d files", len(result.Impact))

// 		// 验证直接变更
// 		if len(result.Changes) != 1 {
// 			t.Errorf("❌ Expected 1 changed file, got %d", len(result.Changes))
// 		}

// 		// 验证间接受影响的文件
// 		// Button 被 Form, Table, Modal, Card 直接引用 (Level 1)
// 		expectedImpacts := []fileImpactExpectation{
// 			{path: filepath.Join(ctx.projectPath, "src/components/Form/Form.tsx"), impactLevel: 1, symbolCount: -1},
// 			{path: filepath.Join(ctx.projectPath, "src/components/Table/Table.tsx"), impactLevel: 1, symbolCount: -1},
// 			{path: filepath.Join(ctx.projectPath, "src/components/Modal/Modal.tsx"), impactLevel: 1, symbolCount: -1},
// 			{path: filepath.Join(ctx.projectPath, "src/components/Card/Card.tsx"), impactLevel: 1, symbolCount: -1},
// 		}

// 		verifyFileImpacts(t, result.Impact, expectedImpacts, false)

// 		// 验证不应该被影响的文件
// 		impactedPaths := make(map[string]bool)
// 		for _, impact := range result.Impact {
// 			impactedPaths[impact.Path] = true
// 		}

// 		notImpactedFiles := []string{
// 			filepath.Join(ctx.projectPath, "src/components/Badge/Badge.tsx"), // Badge 只依赖 Input
// 			filepath.Join(ctx.projectPath, "src/components/Dropdown/Dropdown.tsx"),
// 			filepath.Join(ctx.projectPath, "src/components/Tabs/Tabs.tsx"),
// 			filepath.Join(ctx.projectPath, "src/components/Tooltip/Tooltip.tsx"),
// 		}

// 		for _, path := range notImpactedFiles {
// 			if impactedPaths[path] {
// 				t.Errorf("❌ File should NOT be impacted: %s", path)
// 			}
// 		}

// 		t.Logf("✓ All assertions passed")
// 	})

// 	t.Run("Component Level - Button Component Change", func(t *testing.T) {
// 		buttonPath := filepath.Join(ctx.projectPath, "src/components/Button/Button.tsx")

// 		// 文件级分析
// 		fileInput := &file_analyzer.Input{
// 			ChangedSymbols: []file_analyzer.ChangedSymbol{
// 				{
// 					Name:       "Button",
// 					FilePath:   buttonPath,
// 					ExportType: symbol_analysis.ExportTypeDefault,
// 				},
// 			},
// 		}

// 		fileResult, err := ctx.fileAnalyzer.Analyze(fileInput)
// 		if err != nil {
// 			t.Fatalf("Failed to analyze file impact: %v", err)
// 		}

// 		// 构建文件依赖图
// 		graphBuilder := file_analyzer.NewGraphBuilder(ctx.parsingResult)
// 		fileGraph := graphBuilder.BuildFileDependencyGraph()

// 		// 转换为组件级输入
// 		fileResultProxy := &component_analyzer.FileAnalysisResultProxy{
// 			DepGraph:    fileGraph.DepGraph,
// 			RevDepGraph: fileGraph.RevDepGraph,
// 			ExternalDeps: fileGraph.ExternalDeps,
// 		}

// 		for _, change := range fileResult.Changes {
// 			fileResultProxy.Changes = append(fileResultProxy.Changes, component_analyzer.FileChangeInfoProxy{
// 				Path:        change.Path,
// 				ChangeType:  impact_analysis.ChangeTypeModified,
// 				SymbolCount: change.SymbolCount,
// 			})
// 		}

// 		for _, impact := range fileResult.Impact {
// 			fileResultProxy.Impact = append(fileResultProxy.Impact, component_analyzer.FileImpactInfoProxy{
// 				Path:        impact.Path,
// 				ImpactLevel: impact_analysis.ImpactLevel(impact.ImpactLevel),
// 				ImpactType:  impact_analysis.ImpactType(impact.ImpactType),
// 				ChangePaths: impact.ChangePaths,
// 			})
// 		}

// 		compInput := &component_analyzer.Input{FileResult: fileResultProxy}
// 		compResult, err := ctx.componentAnalyzer.Analyze(compInput)
// 		if err != nil {
// 			t.Fatalf("Failed to analyze component impact: %v", err)
// 		}

// 		t.Logf("Component Analysis Result:")
// 		t.Logf("  Changed: %d components", len(compResult.Changes))
// 		t.Logf("  Impacted: %d components", len(compResult.Impact))

// 		// 验证变更的组件
// 		if len(compResult.Changes) != 1 {
// 			t.Errorf("❌ Expected 1 changed component, got %d", len(compResult.Changes))
// 		}
// 		if len(compResult.Changes) > 0 && compResult.Changes[0].Name != "Button" {
// 			t.Errorf("❌ Expected Button component changed, got %s", compResult.Changes[0].Name)
// 		}

// 		// 验证受影响的组件
// 		expectedImpacts := []componentImpactExpectation{
// 			{name: "Form", impactLevel: 1, symbolCount: -1},
// 			{name: "Table", impactLevel: 1, symbolCount: -1},
// 			{name: "Modal", impactLevel: 1, symbolCount: -1},
// 			{name: "Card", impactLevel: 1, symbolCount: -1},
// 		}

// 		verifyComponentImpacts(t, compResult.Impact, expectedImpacts, false)

// 		// 验证不应该被影响的组件
// 		impactedNames := make(map[string]bool)
// 		for _, impact := range compResult.Impact {
// 			impactedNames[impact.Name] = true
// 		}

// 		notImpactedComponents := []string{
// 			"Badge",      // Badge 只依赖 Input
// 			"Dropdown", "Tabs", "Tooltip", "Select", "Input",
// 		}

// 		for _, name := range notImpactedComponents {
// 			if impactedNames[name] {
// 				t.Errorf("❌ Component should NOT be impacted: %s", name)
// 			}
// 		}

// 		t.Logf("✓ All assertions passed")
// 	})
// }

// // =============================================================================
// // 场景 2: 多组件同时变更
// // =============================================================================

// func TestIntegration_Scenario_MultipleComponentChanges(t *testing.T) {
// 	ctx := setupTestProject(t)

// 	t.Run("Button and Input Components Changed", func(t *testing.T) {
// 		buttonPath := filepath.Join(ctx.projectPath, "src/components/Button/Button.tsx")
// 		inputPath := filepath.Join(ctx.projectPath, "src/components/Input/Input.tsx")

// 		input := &file_analyzer.Input{
// 			ChangedSymbols: []file_analyzer.ChangedSymbol{
// 				{
// 					Name:       "Button",
// 					FilePath:   buttonPath,
// 					ExportType: symbol_analysis.ExportTypeDefault,
// 				},
// 				{
// 					Name:       "Input",
// 					FilePath:   inputPath,
// 					ExportType: symbol_analysis.ExportTypeDefault,
// 				},
// 			},
// 		}

// 		result, err := ctx.fileAnalyzer.Analyze(input)
// 		if err != nil {
// 			t.Fatalf("Failed to analyze: %v", err)
// 		}

// 		t.Logf("File Analysis Result:")
// 		t.Logf("  Changed: %d files", len(result.Changes))
// 		t.Logf("  Impacted: %d files", len(result.Impact))

// 		// 验证两个文件被变更
// 		if len(result.Changes) != 2 {
// 			t.Errorf("❌ Expected 2 changed files, got %d", len(result.Changes))
// 		}

// 		// Form 和 Table 都引用了 Button 和 Input，所以它们应该受影响
// 		// Badge 只引用 Input，所以也应该受影响
// 		expectedImpacts := []fileImpactExpectation{
// 			{path: filepath.Join(ctx.projectPath, "src/components/Form/Form.tsx"), impactLevel: 1, symbolCount: 2}, // 同时引用 Button 和 Input
// 			{path: filepath.Join(ctx.projectPath, "src/components/Table/Table.tsx"), impactLevel: 1, symbolCount: 2},
// 			{path: filepath.Join(ctx.projectPath, "src/components/Modal/Modal.tsx"), impactLevel: 1, symbolCount: 1}, // 只引用 Button
// 			{path: filepath.Join(ctx.projectPath, "src/components/Card/Card.tsx"), impactLevel: 1, symbolCount: 1},  // 只引用 Button
// 			{path: filepath.Join(ctx.projectPath, "src/components/Badge/Badge.tsx"), impactLevel: 1, symbolCount: 1}, // 只引用 Input
// 		}

// 		verifyFileImpacts(t, result.Impact, expectedImpacts, false)

// 		t.Logf("✓ All assertions passed")
// 	})
// }

// // =============================================================================
// // 场景 3: 无依赖组件变更（负向测试）
// // =============================================================================

// func TestIntegration_Scenario_IsolatedComponentChange(t *testing.T) {
// 	ctx := setupTestProject(t)

// 	// Tooltip 是一个独立的组件，不依赖其他任何组件
// 	t.Run("Tooltip Component (No Dependencies)", func(t *testing.T) {
// 		tooltipPath := filepath.Join(ctx.projectPath, "src/components/Tooltip/Tooltip.tsx")

// 		input := &file_analyzer.Input{
// 			ChangedSymbols: []file_analyzer.ChangedSymbol{
// 				{
// 					Name:       "Tooltip",
// 					FilePath:   tooltipPath,
// 					ExportType: symbol_analysis.ExportTypeDefault,
// 				},
// 			},
// 		}

// 		result, err := ctx.fileAnalyzer.Analyze(input)
// 		if err != nil {
// 			t.Fatalf("Failed to analyze: %v", err)
// 		}

// 		t.Logf("File Analysis Result:")
// 		t.Logf("  Changed: %d files", len(result.Changes))
// 		t.Logf("  Impacted: %d files", len(result.Impact))

// 		// 验证只有 Tooltip 被变更
// 		if len(result.Changes) != 1 {
// 			t.Errorf("❌ Expected 1 changed file, got %d", len(result.Changes))
// 		}

// 		// 验证没有其他文件受影响
// 		if len(result.Impact) != 0 {
// 			t.Errorf("❌ Expected 0 impacted files for isolated component, got %d", len(result.Impact))
// 			for _, impact := range result.Impact {
// 				t.Errorf("  Unexpectedly impacted: %s", impact.Path)
// 			}
// 		}

// 		t.Logf("✓ All assertions passed - isolated component correctly identified")
// 	})
// }

// // =============================================================================
// // 场景 4: 影响路径验证
// // =============================================================================

// func TestIntegration_Scenario_ImpactPathTracking(t *testing.T) {
// 	ctx := setupTestProject(t)

// 	t.Run("Verify Impact Paths from Button to Dependents", func(t *testing.T) {
// 		buttonPath := filepath.Join(ctx.projectPath, "src/components/Button/Button.tsx")

// 		input := &file_analyzer.Input{
// 			ChangedSymbols: []file_analyzer.ChangedSymbol{
// 				{
// 					Name:       "Button",
// 					FilePath:   buttonPath,
// 					ExportType: symbol_analysis.ExportTypeDefault,
// 				},
// 			},
// 		}

// 		result, err := ctx.fileAnalyzer.Analyze(input)
// 		if err != nil {
// 			t.Fatalf("Failed to analyze: %v", err)
// 		}

// 		// 验证影响路径包含变更源头
// 		formPath := filepath.Join(ctx.projectPath, "src/components/Form/Form.tsx")
// 		formFound := false
// 		for _, impact := range result.Impact {
// 			if impact.Path == formPath {
// 				formFound = true
// 				if len(impact.ChangePaths) == 0 {
// 					t.Errorf("❌ Form.tsx should have change paths, got none")
// 				} else {
// 					// 验证路径包含 Button
// 					pathContainsButton := false
// 					for _, path := range impact.ChangePaths {
// 						if strings.Contains(path, "Button") {
// 							pathContainsButton = true
// 							break
// 						}
// 					}
// 					if !pathContainsButton {
// 						t.Errorf("❌ Form.tsx change path should contain Button, got: %v", impact.ChangePaths)
// 					}
// 					t.Logf("✓ Form.tsx impact paths: %v", impact.ChangePaths)
// 				}
// 				break
// 			}
// 		}

// 		if !formFound {
// 			t.Error("❌ Form.tsx should be in impacted files")
// 		}

// 		t.Logf("✓ Impact path tracking verified")
// 	})
// }

// // =============================================================================
// // 场景 5: 组件清单完整性验证
// // =============================================================================

// func TestIntegration_ComponentManifestCompleteness(t *testing.T) {
// 	ctx := setupTestProject(t)

// 	t.Run("All Manifest Components Have Valid Entries", func(t *testing.T) {
// 		for _, comp := range ctx.manifest.Components {
// 			// 验证入口文件存在
// 			if _, err := os.Stat(comp.Entry); os.IsNotExist(err) {
// 				t.Errorf("❌ Component %s entry file does not exist: %s", comp.Name, comp.Entry)
// 				continue
// 			}

// 			// 验证入口文件可以被解析
// 			found := false
// 			for filePath := range ctx.parsingResult.Js_Data {
// 				if strings.HasPrefix(filePath, filepath.Dir(comp.Entry)) {
// 					found = true
// 					break
// 				}
// 			}

// 			if !found {
// 				t.Errorf("❌ Component %s entry not found in parsed files: %s", comp.Name, comp.Entry)
// 			} else {
// 				t.Logf("✓ Component %s: entry verified", comp.Name)
// 			}
// 		}

// 		t.Logf("✓ All %d components verified", len(ctx.manifest.Components))
// 	})
// }

// // =============================================================================
// // 场景 6: 边界条件测试
// // =============================================================================

// func TestIntegration_EdgeCases(t *testing.T) {
// 	ctx := setupTestProject(t)

// 	t.Run("Empty Changed Symbols List", func(t *testing.T) {
// 		input := &file_analyzer.Input{
// 			ChangedSymbols: []file_analyzer.ChangedSymbol{},
// 		}

// 		result, err := ctx.fileAnalyzer.Analyze(input)
// 		if err != nil {
// 			t.Fatalf("Failed to analyze: %v", err)
// 		}

// 		if len(result.Changes) != 0 {
// 			t.Errorf("❌ Expected 0 changed files, got %d", len(result.Changes))
// 		}
// 		if len(result.Impact) != 0 {
// 			t.Errorf("❌ Expected 0 impacted files, got %d", len(result.Impact))
// 		}

// 		t.Logf("✓ Empty input handled correctly")
// 	})

// 	t.Run("Non-Existent Symbol Change", func(t *testing.T) {
// 		input := &file_analyzer.Input{
// 			ChangedSymbols: []file_analyzer.ChangedSymbol{
// 				{
// 					Name:       "NonExistentSymbol",
// 					FilePath:   "/fake/path/NonExistent.tsx",
// 					ExportType: symbol_analysis.ExportTypeDefault,
// 				},
// 			},
// 		}

// 		result, err := ctx.fileAnalyzer.Analyze(input)
// 		if err != nil {
// 			t.Fatalf("Failed to analyze: %v", err)
// 		}

// 		// 应该有 1 个变更文件（即使它不在项目中）
// 		if len(result.Changes) != 1 {
// 			t.Errorf("❌ Expected 1 changed file (non-existent), got %d", len(result.Changes))
// 		}
// 		// 但不应该有任何影响（因为它不存在）
// 		if len(result.Impact) != 0 {
// 			t.Errorf("❌ Expected 0 impacted files for non-existent symbol, got %d", len(result.Impact))
// 		}

// 		t.Logf("✓ Non-existent symbol handled correctly")
// 	})
// }
