package parser_test

import (
	"encoding/json"
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
)

// TestAnalyzeTypeDecl 测试分析类型别名声明的功能
func TestAnalyzeTypeDecl(t *testing.T) {
	// expectedResult 定义了测试期望的结果结构体
	type expectedResult struct {
		Identifier string                          `json:"identifier"` // 类型别名的标识符
		Raw        string                          `json:"raw"`        // 原始代码文本
		Reference  map[string]parser.TypeReference `json:"reference"`  // 类型引用
	}

	// testCases 定义了一系列的测试用例
	testCases := []struct {
		name           string         // 测试用例名称
		code           string         // 需要被解析的代码
		expectedResult expectedResult // 期望的解析结果
	}{
		{
			name: "简单的类型别名",
			code: `type MyString = string;`,
			expectedResult: expectedResult{
				Identifier: "MyString",
				Raw:        "type MyString = string;",
				Reference:  map[string]parser.TypeReference{},
			},
		},
		{
			name: "带自定义类型的类型别名",
			code: `type UserResponse = Response<User>;`,
			expectedResult: expectedResult{
				Identifier: "UserResponse",
				Raw:        "type UserResponse = Response<User>;",
				Reference: map[string]parser.TypeReference{
					"Response": {
						Identifier: "Response",
						Location:   []string{"UserResponse"},
						IsExtend:   false,
					},
					"User": {
						Identifier: "User",
						Location:   []string{"UserResponse<>"},
						IsExtend:   false,
					},
				},
			},
		},
		{
			name: "带联合类型的类型别名",
			code: `type MyUnion = TypeA | TypeB;`,
			expectedResult: expectedResult{
				Identifier: "MyUnion",
				Raw:        "type MyUnion = TypeA | TypeB;",
				Reference: map[string]parser.TypeReference{
					"TypeA": {
						Identifier: "TypeA",
						Location:   []string{"MyUnion"},
						IsExtend:   false,
					},
					"TypeB": {
						Identifier: "TypeB",
						Location:   []string{"MyUnion"},
						IsExtend:   false,
					},
				},
			},
		},
		{
			name: "映射类型",
			code: `type MappedType = { [key in SupportedLanguages]?: string[] | string }`,
			expectedResult: expectedResult{
				Identifier: "MappedType",
				Raw:        "type MappedType = { [key in SupportedLanguages]?: string[] | string }",
				Reference: map[string]parser.TypeReference{
					"SupportedLanguages": {
						Identifier: "SupportedLanguages",
						Location:   []string{""},
						IsExtend:   false,
					},
				},
			},
		},
		{
			name: "索引访问类型",
			code: `type PersonName = Translations["name"];`,
			expectedResult: expectedResult{
				Identifier: "PersonName",
				Raw:        "type PersonName = Translations[\"name\"];",
				Reference: map[string]parser.TypeReference{
					"Translations": {
						Identifier: "Translations",
						Location:   []string{"PersonName"},
						IsExtend:   false,
					},
				},
			},
		},
		{
			name: "带字符串键的类型",
			code: `type A = { "name": string };`,
			expectedResult: expectedResult{
				Identifier: "A",
				Raw:        "type A = { \"name\": string };",
				Reference:  map[string]parser.TypeReference{},
			},
		},
	}

	// extractFn 定义了如何从完整的解析结果中提取我们关心的部分
	extractFn := func(result *parser.ParserResult) parser.TypeDeclarationResult {
		for _, typeDecl := range result.TypeDeclarations {
			return typeDecl
		}
		return parser.TypeDeclarationResult{}
	}

	// marshalFn 定义了如何将提取出的结果序列化为 JSON
	marshalFn := func(result parser.TypeDeclarationResult) ([]byte, error) {
		return json.MarshalIndent(struct {
			Identifier string                          `json:"identifier"`
			Raw        string                          `json:"raw"`
			Reference  map[string]parser.TypeReference `json:"reference"`
		}{
			Identifier: result.Identifier,
			Raw:        result.Raw,
			Reference:  result.Reference,
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
