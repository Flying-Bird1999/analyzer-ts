package trace

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

func TestTracerConfigure(t *testing.T) {
	testCases := []struct {
		name         string
		params       map[string]string
		expectErr    bool
		expectedPkgs map[string]struct{}
	}{
		{
			name: "正常情况 - 单个包",
			params: map[string]string{
				"targetPkgs": "antd",
			},
			expectErr:    false,
			expectedPkgs: map[string]struct{}{"antd": {}},
		},
		{
			name: "正常情况 - 多个包，逗号分隔",
			params: map[string]string{
				"targetPkgs": "antd,lodash,moment",
			},
			expectErr:    false,
			expectedPkgs: map[string]struct{}{"antd": {}, "lodash": {}, "moment": {}},
		},
		{
			name: "正常情况 - 带有多余的空格",
			params: map[string]string{
				"targetPkgs": " antd ,  lodash  ",
			},
			expectErr:    false,
			expectedPkgs: map[string]struct{}{"antd": {}, "lodash": {}},
		},
		{
			name:      "失败情况 - 缺少 targetPkgs 参数",
			params:    map[string]string{},
			expectErr: true,
		},
		{
			name: "失败情况 - targetPkgs 参数为空字符串",
			params: map[string]string{
				"targetPkgs": "",
			},
			expectErr: true,
		},
		{
			name: "失败情况 - targetPkgs 参数只包含空格和逗号",
			params: map[string]string{
				"targetPkgs": " , , ",
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tracer := &Tracer{}
			err := tracer.Configure(tc.params)

			if tc.expectErr {
				if err == nil {
					t.Errorf("预期出现错误，但实际没有")
				}
			} else {
				if err != nil {
					t.Errorf("预期没有错误，但实际出现了: %v", err)
				}
				if !reflect.DeepEqual(tracer.TargetPkgs, tc.expectedPkgs) {
					t.Errorf("TargetPkgs 不匹配，预期是 %v, 实际是 %v", tc.expectedPkgs, tracer.TargetPkgs)
				}
			}
		})
	}
}

func TestTracerAnalyze(t *testing.T) {
	// 准备路径
	projectRoot, _ := filepath.Abs("/test-project")
	componentPath := filepath.Join(projectRoot, "component.ts")
	unrelatedPath := filepath.Join(projectRoot, "unrelated.ts")

	// 1. 准备测试数据
	mockParsingResult := &projectParser.ProjectParserResult{
		Js_Data: map[string]projectParser.JsFileParserResult{
			// component.ts: 从 antd 导入 Button，并创建 MyButton
			componentPath: {
				ImportDeclarations: []projectParser.ImportDeclarationResult{
					{
						Source:        projectParser.SourceData{Type: "npm", NpmPkg: "antd"},
						ImportModules: []projectParser.ImportModule{{Identifier: "Button"}},
					},
				},
				VariableDeclarations: []parser.VariableDeclaration{
					{
						Declarators: []*parser.VariableDeclarator{
							{Identifier: "MyButton", InitValue: &parser.VariableValue{Type: "identifier", Expression: "Button"}},
						},
					},
				},
			},
			// unrelated.ts: 无关文件
			unrelatedPath: {},
		},
	}

	// 2. 创建分析器和上下文
	analyzer := &Tracer{}
	ctx := &projectanalyzer.ProjectContext{
		ParsingResult: mockParsingResult,
	}

	// 3. 配置分析器
	params := map[string]string{"targetPkgs": "antd"}
	if err := analyzer.Configure(params); err != nil {
		t.Fatalf("Configure() failed: %v", err)
	}

	// 4. 执行分析
	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze() returned an unexpected error: %v", err)
	}

	// 5. 断言结果
	traceResult, ok := result.(*TraceResult)
	if !ok {
		t.Fatalf("Analyze() returned result of wrong type: got %T, want *TraceResult", result)
	}

	// 检查结果中是否只包含 component.ts
	if len(traceResult.Data) != 1 {
		t.Fatalf("Expected 1 file in trace result, but got %d", len(traceResult.Data))
	}
	if _, exists := traceResult.Data[componentPath]; !exists {
		t.Fatalf("Expected trace result to contain file %s, but it did not", componentPath)
	}

	// 检查 component.ts 的结果内容
	fileData, ok := traceResult.Data[componentPath].(map[string]interface{})
	if !ok {
		t.Fatalf("Unexpected data type for file result")
	}

	// 检查是否包含相关的导入声明
	imports, exists := fileData["importDeclarations"].([]projectParser.ImportDeclarationResult)
	if !exists || len(imports) != 1 {
		t.Errorf("Expected to find 1 relevant import declaration, but got %d", len(imports))
	} else if imports[0].Source.NpmPkg != "antd" {
		t.Errorf("Expected import source to be 'antd', but got %s", imports[0].Source.NpmPkg)
	}

	// 检查是否包含相关的变量声明
	vars, exists := fileData["variableDeclarations"].([]parser.VariableDeclaration)
	if !exists || len(vars) != 1 {
		t.Errorf("Expected to find 1 relevant variable declaration, but got %d", len(vars))
	} else if vars[0].Declarators[0].Identifier != "MyButton" {
		t.Errorf("Expected tainted variable to be 'MyButton', but got %s", vars[0].Declarators[0].Identifier)
	}
}
