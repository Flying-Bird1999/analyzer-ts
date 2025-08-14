package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// TestAnalyzeTypeDecl 测试分析类型别名声明的功能
func TestAnalyzeTypeDecl(t *testing.T) {
	// expectedResult 定义了测试期望的结果结构体
	type expectedResult struct {
		Identifier string                            `json:"identifier"` // 类型别名的标识符
		Raw        string                            `json:"raw"`        // 原始代码文本
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

	// findNode 是一个辅助函数，用于在 AST 中查找第一个类型别名声明节点
	findNode := func(sourceFile *ast.SourceFile) *ast.TypeAliasDeclaration {
		for _, stmt := range sourceFile.Statements.Nodes {
			if stmt.Kind == ast.KindTypeAliasDeclaration {
				return stmt.AsTypeAliasDeclaration()
			}
		}
		return nil
	}

	// testParser 是一个辅助函数，用于执行解析操作
	testParser := func(node *ast.TypeAliasDeclaration, code string) *parser.TypeDeclarationResult {
		result := parser.NewTypeDeclarationResult(node.AsNode(), code)
		result.AnalyzeTypeDecl(node)
		return result
	}

	// marshal 是一个辅助函数，用于将解析结果序列化为 JSON
	marshal := func(result *parser.TypeDeclarationResult) ([]byte, error) {
		return json.MarshalIndent(struct {
			Identifier string                            `json:"identifier"`
			Raw        string                            `json:"raw"`
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
			RunTest(t, tc.code, string(expectedJSON), findNode, testParser, marshal)
		})
	}
}