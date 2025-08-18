package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"
)

// TestAnalyzeCallExpression 测试分析调用表达式的功能
func TestAnalyzeCallExpression(t *testing.T) {
	// expectedResult 定义了测试期望的结果结构体
	type expectedResult struct {
		CallChain []string          `json:"callChain"` // 调用链
		Arguments []parser.Argument `json:"arguments"` // 参数
		Type      string            `json:"type"`      // 类型
	}

	// testCases 定义了一系列的测试用例
	testCases := []struct {
		name           string         // 测试用例名称
		code           string         // 需要被解析的代码
		expectedResult expectedResult // 期望的解析结果
	}{
		{
			name: "简单的函数调用",
			code: `myFunction();`,
			expectedResult: expectedResult{
				CallChain: []string{"myFunction"},
				Arguments: []parser.Argument{},
				Type:      "call",
			},
		},
		{
			name: "带参数的函数调用",
			code: `myFunction(1, "hello", true, myVar);`,
			expectedResult: expectedResult{
				CallChain: []string{"myFunction"},
				Arguments: []parser.Argument{
					{Type: "number", Text: "1"},
					{Type: "string", Text: "\"hello\""},
					{Type: "boolean", Text: "true"},
					{Type: "identifier", Text: "myVar"},
				},
				Type: "call",
			},
		},
		{
			name: "成员访问调用",
			code: `myObj.myMethod();`,
			expectedResult: expectedResult{
				CallChain: []string{"myObj", "myMethod"},
				Arguments: []parser.Argument{},
				Type:      "member",
			},
		},
		{
			name: "链式成员访问调用",
			code: `this.a.b.c(123);`,
			expectedResult: expectedResult{
				CallChain: []string{"this", "a", "b", "c"},
				Arguments: []parser.Argument{
					{Type: "number", Text: "123"},
				},
				Type:      "member",
			},
		},
	}

	// extractFn 定义了如何从完整的解析结果中提取我们关心的部分
	extractFn := func(result *parser.ParserResult) parser.CallExpression {
		if len(result.CallExpressions) > 0 {
			return result.CallExpressions[0]
		}
		return parser.CallExpression{}
	}

	// marshalFn 定义了如何将提取出的结果序列化为 JSON
	marshalFn := func(result parser.CallExpression) ([]byte, error) {
		return json.MarshalIndent(struct {
			CallChain []string          `json:"callChain"`
			Arguments []parser.Argument `json:"arguments"`
			Type      string            `json:"type"`
		}{
			CallChain: result.CallChain,
			Arguments: result.Arguments,
			Type:      result.Type,
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