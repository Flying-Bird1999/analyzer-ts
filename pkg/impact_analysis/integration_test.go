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
			ImpactType:  impact_analysis.ImpactType(impact.ImpactType),
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
