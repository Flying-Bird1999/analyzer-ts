package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"
)

// TestNewEnumDeclarationResult 测试分析枚举声明的功能
func TestNewEnumDeclarationResult(t *testing.T) {
	// expectedResult 定义了测试期望的结果结构体
	type expectedResult struct {
		Identifier string `json:"identifier"` // 枚举的标识符
		Raw        string `json:"raw"`        // 原始代码文本
	}

	// testCases 定义了一系列的测试用例
	testCases := []struct {
		name           string         // 测试用例名称
		code           string         // 需要被解析的代码
		expectedResult expectedResult // 期望的解析结果
	}{
		{
			name: "简单的枚举",
			code: `enum Color { Red, Green, Blue }`,
			expectedResult: expectedResult{
				Identifier: "Color",
				Raw:        "enum Color { Red, Green, Blue }",
			},
		},
		{
			name: "带初始化值的枚举",
			code: `enum Direction { Up = 1, Down, Left, Right }`,
			expectedResult: expectedResult{
				Identifier: "Direction",
				Raw:        "enum Direction { Up = 1, Down, Left, Right }",
			},
		},
	}

	// extractFn 定义了如何从完整的解析结果中提取我们关心的部分
	extractFn := func(result *parser.ParserResult) parser.EnumDeclarationResult {
		for _, enum := range result.EnumDeclarations {
			return enum
		}
		return parser.EnumDeclarationResult{}
	}

	// marshalFn 定义了如何将提取出的结果序列化为 JSON
	marshalFn := func(result parser.EnumDeclarationResult) ([]byte, error) {
		return json.MarshalIndent(struct {
			Identifier string `json:"identifier"`
			Raw        string `json:"raw"`
		}{
			Identifier: result.Identifier,
			Raw:        result.Raw,
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
