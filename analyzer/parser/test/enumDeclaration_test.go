package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
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

	// findNode 是一个辅助函数，用于在 AST 中查找第一个枚举声明节点
	findNode := func(sourceFile *ast.SourceFile) *ast.EnumDeclaration {
		for _, stmt := range sourceFile.Statements.Nodes {
			if stmt.Kind == ast.KindEnumDeclaration {
				return stmt.AsEnumDeclaration()
			}
		}
		return nil
	}

	// testParser 是一个辅助函数，用于执行解析操作
	testParser := func(node *ast.EnumDeclaration, code string) *parser.EnumDeclarationResult {
		return parser.NewEnumDeclarationResult(node, code)
	}

	// marshal 是一个辅助函数，用于将解析结果序列化为 JSON
	marshal := func(result *parser.EnumDeclarationResult) ([]byte, error) {
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
			RunTest(t, tc.code, string(expectedJSON), findNode, testParser, marshal)
		})
	}
}