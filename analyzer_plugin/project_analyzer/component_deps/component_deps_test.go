package component_deps

import (
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

func TestComponentDepsAnalyzer(t *testing.T) {
	// 准备测试路径
	projectRoot, _ := filepath.Abs("/test-project")
	pkgAPath := filepath.Join(projectRoot, "packages/pkg-a")
	buttonPath := filepath.Join(pkgAPath, "src/button/index.tsx")
	cardPath := filepath.Join(pkgAPath, "src/card/index.tsx")
	entryPath := filepath.Join(pkgAPath, "src/index.ts")
	pkgAJsonPath := filepath.Join(pkgAPath, "package.json")

	// 1. 准备测试数据
	mockParsingResult := &projectParser.ProjectParserResult{
		// JS/TS 文件解析结果
		Js_Data: map[string]projectParser.JsFileParserResult{
			// Button 组件: 自身定义并导出 Button
			buttonPath: {
				ExportDeclarations: []projectParser.ExportDeclarationResult{
					{ExportModules: []projectParser.ExportModule{{Identifier: "Button", ModuleName: "Button"}}},
				},
			},
			// Card 组件: 导入 Button，并导出 Card
			cardPath: {
				ImportDeclarations: []projectParser.ImportDeclarationResult{
					{Source: projectParser.SourceData{FilePath: buttonPath}},
				},
				ExportDeclarations: []projectParser.ExportDeclarationResult{
					{ExportModules: []projectParser.ExportModule{{Identifier: "Card", ModuleName: "Card"}}},
				},
			},
			// 入口文件: 再导出 Button 和 Card
			entryPath: {
				ExportDeclarations: []projectParser.ExportDeclarationResult{
					{
						Source:        &projectParser.SourceData{FilePath: buttonPath},
						ExportModules: []projectParser.ExportModule{{Identifier: "Button", ModuleName: "Button"}},
					},
					{
						Source:        &projectParser.SourceData{FilePath: cardPath},
						ExportModules: []projectParser.ExportModule{{Identifier: "Card", ModuleName: "Card"}},
					},
				},
			},
		},
		// package.json 解析结果
		Package_Data: map[string]projectParser.PackageJsonFileParserResult{
			"pkg-a": {Path: pkgAJsonPath, Namespace: "pkg-a"},
		},
	}

	// 2. 创建分析器和上下文
	analyzer := &ComponentDependencyAnalyzer{}
	ctx := &projectanalyzer.ProjectContext{
		ProjectRoot:   projectRoot,
		ParsingResult: mockParsingResult,
	}

	// 3. 配置分析器
	// 注意：EntryPoint 是相对于项目根目录的
	params := map[string]string{"entryPoint": "packages/pkg-a/src/index.ts"}
	if err := analyzer.Configure(params); err != nil {
		t.Fatalf("Configure() failed: %v", err)
	}

	// 4. 执行分析
	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze() returned an unexpected error: %v", err)
	}

	// 5. 断言结果
	depsResult, ok := result.(*Result)
	if !ok {
		t.Fatalf("Analyze() returned result of wrong type: got %T, want *Result", result)
	}

	// 检查包是否存在
	pkgAComponents, pkgExists := depsResult.Packages["pkg-a"]
	if !pkgExists {
		t.Fatalf("Expected package 'pkg-a' to exist in results, but it did not")
	}

	// 检查组件数量
	if len(pkgAComponents) != 2 {
		t.Fatalf("Expected 2 components in 'pkg-a', but got %d", len(pkgAComponents))
	}

	// 检查 Card 组件的依赖
	cardInfo, cardExists := pkgAComponents["Card"]
	if !cardExists {
		t.Fatalf("Expected component 'Card' to exist, but it did not")
	}
	expectedCardDeps := []string{"Button"}
	sort.Strings(cardInfo.Dependencies)
	if !reflect.DeepEqual(cardInfo.Dependencies, expectedCardDeps) {
		t.Errorf("Expected Card dependencies to be %v, but got %v", expectedCardDeps, cardInfo.Dependencies)
	}

	// 检查 Button 组件的依赖
	buttonInfo, buttonExists := pkgAComponents["Button"]
	if !buttonExists {
		t.Fatalf("Expected component 'Button' to exist, but it did not")
	}
	if len(buttonInfo.Dependencies) != 0 {
		t.Errorf("Expected Button to have 0 dependencies, but got %d", len(buttonInfo.Dependencies))
	}
}
