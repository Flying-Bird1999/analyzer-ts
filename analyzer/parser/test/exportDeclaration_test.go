package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"
)

// TestAnalyzeExportDeclaration 测试分析导出声明的功能
func TestAnalyzeExportDeclaration(t *testing.T) {
	// expectedResult 定义了测试期望的结果结构体
	type expectedResult struct {
		ExportModules []parser.ExportModule `json:"exportModules"` // 导出的模块列表
		Raw           string                `json:"raw"`            // 原始代码文本
		Source        string                `json:"source"`         // 导出来源
		Type          string                `json:"type"`           // 导出类型
	}

	// testCases 定义了一系列的测试用例
	testCases := []struct {
		name           string         // 测试用例名称
		code           string         // 需要被解析的代码
		expectedResult expectedResult // 期望的解析结果
	}{
		{
			name: "命名导出",
			code: "export { name1, name2 as alias };",
			expectedResult: expectedResult{
				ExportModules: []parser.ExportModule{
					{ModuleName: "name1", Type: "named", Identifier: "name1"},
					{ModuleName: "name2", Type: "named", Identifier: "alias"},
				},
				Raw:    "export { name1, name2 as alias };",
				Source: "",
				Type:   "named-export",
			},
		},
		{
			name: "从模块中再次导出",
			code: `export { name1 } from "./mod";`,
			expectedResult: expectedResult{
				ExportModules: []parser.ExportModule{
					{ModuleName: "name1", Type: "named", Identifier: "name1"},
				},
				Raw:    "export { name1 } from \"./mod\";",
				Source: "./mod",
				Type:   "re-export",
			},
		},
		{
			name: "通配符再次导出",
			code: `export * from "./mod";`,
			expectedResult: expectedResult{
				ExportModules: []parser.ExportModule{
					{ModuleName: "*", Type: "namespace", Identifier: "*"},
				},
				Raw:    "export * from \"./mod\";",
				Source: "./mod",
				Type:   "re-export",
			},
		},
		{
			name: "命名空间再次导出",
			code: `export * as ns from "./mod";`,
			expectedResult: expectedResult{
				ExportModules: []parser.ExportModule{
					{ModuleName: "*", Type: "namespace", Identifier: "ns"},
				},
				Raw:    "export * as ns from \"./mod\";",
				Source: "./mod",
				Type:   "re-export",
			},
		},
	}

	// extractFn 定义了如何从完整的解析结果中提取我们关心的部分
	extractFn := func(result *parser.ParserResult) parser.ExportDeclarationResult {
		if len(result.ExportDeclarations) > 0 {
			return result.ExportDeclarations[0]
		}
		return parser.ExportDeclarationResult{}
	}

	// marshalFn 定义了如何将提取出的结果序列化为 JSON
	marshalFn := func(result parser.ExportDeclarationResult) ([]byte, error) {
		return json.MarshalIndent(struct {
			ExportModules []parser.ExportModule `json:"exportModules"`
			Raw           string                `json:"raw"`
			Source        string                `json:"source"`
			Type          string                `json:"type"`
		}{
			ExportModules: result.ExportModules,
			Raw:           result.Raw,
			Source:        result.Source,
			Type:          result.Type,
		}, "", "\t")
	}

	// 遍历所有测试用例并执行测试
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 将期望结果序列化为 JSON
			expectedJSON, err := json.MarshalIndent(tc.expectedResult, "", "\t")
			if err != nil {
				t.Fatalf("无法将期望结果序列化为 JSON: %v", err)
			}
			// 运行测试
			RunTest(t, tc.code, string(expectedJSON), extractFn, marshalFn)
		})
	}
}