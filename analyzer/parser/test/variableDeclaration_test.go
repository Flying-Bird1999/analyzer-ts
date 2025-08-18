package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"
)

// TestNewVariableDeclaration 测试分析变量声明的功能
func TestNewVariableDeclaration(t *testing.T) {
	// expectedResult 定义了测试期望的结果结构体
	type expectedResult struct {
		Exported    bool                         `json:"exported"`         // 是否导出
		Kind        parser.DeclarationKind       `json:"kind"`             // 声明类型 (const, let, var)
		Source      *parser.VariableValue        `json:"source,omitempty"` // 解构赋值的来源
		Declarators []*parser.VariableDeclarator `json:"declarators"`      // 声明的变量列表
	}

	// testCases 定义了一系列的测试用例
	testCases := []struct {
		name           string         // 测试用例名称
		code           string         // 需要被解析的代码
		expectedResult expectedResult // 期望的解析结果
	}{
		{
			name: "简单的 const 声明",
			code: `const myVar = "hello";`,
			expectedResult: expectedResult{
				Exported: false,
				Kind:     "const",
				Declarators: []*parser.VariableDeclarator{
					{
						Identifier: "myVar",
						InitValue: &parser.VariableValue{
							Type:       "stringLiteral",
							Expression: "\"hello\"",
							Data:       "hello",
						},
					},
				},
			},
		},
		{
			name: "导出的 let 声明",
			code: `export let myVar: number = 123;`,
			expectedResult: expectedResult{
				Exported: true,
				Kind:     "let",
				Declarators: []*parser.VariableDeclarator{
					{
						Identifier: "myVar",
						Type: &parser.VariableValue{
							Type:       "typeNode",
							Expression: "number",
						},
						InitValue: &parser.VariableValue{
							Type:       "numericLiteral",
							Expression: "123",
							Data:       "123",
						},
					},
				},
			},
		},
		{
			name: "对象解构",
			code: `const { a, b: myB } = myObj;`,
			expectedResult: expectedResult{
				Exported: false,
				Kind:     "const",
				Source: &parser.VariableValue{
					Type:       "identifier",
					Expression: "myObj",
					Data:       "myObj",
				},
				Declarators: []*parser.VariableDeclarator{
					{
						Identifier: "a",
						PropName:   "a",
					},
					{
						Identifier: "myB",
						PropName:   "b",
					},
				},
			},
		},
		{
			name: "带计算属性的对象解构",
			code: `const { [key]: value } = obj;`,
			expectedResult: expectedResult{
				Exported: false,
				Kind:     "const",
				Source: &parser.VariableValue{
					Type:       "identifier",
					Expression: "obj",
					Data:       "obj",
				},
				Declarators: []*parser.VariableDeclarator{
					{
						Identifier: "value",
						PropName:   "[key]",
					},
				},
			},
		},
	}

	// extractFn 定义了如何从完整的解析结果中提取我们关心的部分
	extractFn := func(result *parser.ParserResult) parser.VariableDeclaration {
		if len(result.VariableDeclarations) > 0 {
			return result.VariableDeclarations[0]
		}
		return parser.VariableDeclaration{}
	}

	// marshalFn 定义了如何将提取出的结果序列化为 JSON
	marshalFn := func(result parser.VariableDeclaration) ([]byte, error) {
		return json.MarshalIndent(struct {
			Exported    bool                         `json:"exported"`
			Kind        parser.DeclarationKind       `json:"kind"`
			Source      *parser.VariableValue        `json:"source,omitempty"`
			Declarators []*parser.VariableDeclarator `json:"declarators"`
		}{
			Exported:    result.Exported,
			Kind:        result.Kind,
			Source:      result.Source,
			Declarators: result.Declarators,
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
