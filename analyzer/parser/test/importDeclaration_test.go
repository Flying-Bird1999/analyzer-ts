package parser_test

import (
	"encoding/json"
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
)

// TestAnalyzeImportDeclaration 测试分析导入声明的功能
func TestAnalyzeImportDeclaration(t *testing.T) {
	// expectedResult 定义了测试期望的结果结构体
	type expectedResult struct {
		ImportModules []parser.ImportModule `json:"importModules"` // 导入的模块列表
		Raw           string                `json:"raw"`           // 原始代码文本
		Source        string                `json:"source"`        // 导入来源
	}

	// testCases 定义了一系列的测试用例
	testCases := []struct {
		name           string         // 测试用例名称
		code           string         // 需要被解析的代码
		expectedResult expectedResult // 期望的解析结果
	}{
		{
			name: "默认导入",
			code: "import Bird from './type2';",
			expectedResult: expectedResult{
				ImportModules: []parser.ImportModule{
					{ImportModule: "default", Type: "default", Identifier: "Bird"},
				},
				Raw:    "import Bird from './type2';",
				Source: "./type2",
			},
		},
		{
			name: "命名空间导入",
			code: "import * as allTypes from './type';",
			expectedResult: expectedResult{
				ImportModules: []parser.ImportModule{
					{ImportModule: "allTypes", Type: "namespace", Identifier: "allTypes"},
				},
				Raw:    "import * as allTypes from './type';",
				Source: "./type",
			},
		},
		{
			name: "命名导入",
			code: "import { School, Teacher } from './school';",
			expectedResult: expectedResult{
				ImportModules: []parser.ImportModule{
					{ImportModule: "School", Type: "named", Identifier: "School"},
					{ImportModule: "Teacher", Type: "named", Identifier: "Teacher"},
				},
				Raw:    "import { School, Teacher } from './school';",
				Source: "./school",
			},
		},
		{
			name: "带别名的命名导入",
			code: "import { School, School2 as NewSchool } from './school';",
			expectedResult: expectedResult{
				ImportModules: []parser.ImportModule{
					{ImportModule: "School", Type: "named", Identifier: "School"},
					{ImportModule: "School2", Type: "named", Identifier: "NewSchool"},
				},
				Raw:    "import { School, School2 as NewSchool } from './school';",
				Source: "./school",
			},
		},
		{
			name: "副作用导入",
			code: "import './setup';",
			expectedResult: expectedResult{
				ImportModules: []parser.ImportModule{},
				Raw:           "import './setup';",
				Source:        "./setup",
			},
		},
		{
			name: "默认导入和带别名的命名导入",
			code: "import Bird, { School, Teacher as t2 } from './type2';",
			expectedResult: expectedResult{
				ImportModules: []parser.ImportModule{
					{ImportModule: "default", Type: "default", Identifier: "Bird"},
					{ImportModule: "School", Type: "named", Identifier: "School"},
					{ImportModule: "Teacher", Type: "named", Identifier: "t2"},
				},
				Raw:    "import Bird, { School, Teacher as t2 } from './type2';",
				Source: "./type2",
			},
		},
		{
			name: "动态导入",
			code: "const AdminPage = lazy(() => import('./AdminPage'))",
			expectedResult: expectedResult{
				ImportModules: []parser.ImportModule{
					{ImportModule: "default", Type: "dynamic_variable", Identifier: "AdminPage"},
				},
				Raw:    " import('./AdminPage')",
				Source: "./AdminPage",
			},
		},
	}

	// 遍历所有测试用例并执行测试
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// extractFn 定义了如何从完整的解析结果中提取我们关心的部分
			extractFn := func(result *parser.ParserResult) parser.ImportDeclarationResult {
				if len(result.ImportDeclarations) > 0 {
					// 为了测试的稳定性，我们总是返回最后一个找到的导入声明
					// （因为动态导入可能会被加到列表末尾）
					return result.ImportDeclarations[len(result.ImportDeclarations)-1]
				}
				return parser.ImportDeclarationResult{}
			}

			// marshalFn 定义了如何将提取出的结果序列化为 JSON
			marshalFn := func(result parser.ImportDeclarationResult) ([]byte, error) {
				return json.MarshalIndent(struct {
					ImportModules []parser.ImportModule `json:"importModules"`
					Raw           string                `json:"raw"`
					Source        string                `json:"source"`
				}{
					ImportModules: result.ImportModules,
					Raw:           result.Raw,
					Source:        result.Source,
				}, "", "\t")
			}

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
