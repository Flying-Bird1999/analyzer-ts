package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// TestNewVariableDeclaration 测试分析变量声明的功能
func TestNewVariableDeclaration(t *testing.T) {
	// expectedResult 定义了测试期望的结果结构体
	type expectedResult struct {
		Exported    bool                         `json:"exported"`          // 是否导出
		Kind        parser.DeclarationKind       `json:"kind"`              // 声明类型 (const, let, var)
		Source      *parser.VariableValue        `json:"source,omitempty"`   // 解构赋值的来源
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

	// findNode 是一个辅助函数，用于在 AST 中查找第一个变量声明语句节点
	findNode := func(sourceFile *ast.SourceFile) *ast.VariableStatement {
		for _, stmt := range sourceFile.Statements.Nodes {
			if stmt.Kind == ast.KindVariableStatement {
				return stmt.AsVariableStatement()
			}
		}
		return nil
	}

	// testParser 是一个辅助函数，用于执行解析操作
	testParser := func(node *ast.VariableStatement, code string) *parser.VariableDeclaration {
		return parser.NewVariableDeclaration(node, code)
	}

	// marshal 是一个辅助函数，用于将解析结果序列化为 JSON
	marshal := func(result *parser.VariableDeclaration) ([]byte, error) {
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
			RunTest(t, tc.code, string(expectedJSON), findNode, testParser, marshal)
		})
	}
}