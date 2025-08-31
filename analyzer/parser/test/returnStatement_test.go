package parser_test

import (
	"encoding/json"
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/stretchr/testify/assert"
)

func TestReturnStatement(t *testing.T) {
	testCases := []struct {
		name           string
		code           string
		expectedResult parser.ReturnStatementResult
	}{
		{
			name: "返回数字字面量",
			code: `function a() { return 1; }`,
			expectedResult: parser.ReturnStatementResult{
				Expression: &parser.VariableValue{
					Type:       "numericLiteral",
					Expression: "1",
					Data:       "1",
				},
			},
		},
		{
			name: "返回标识符",
			code: `function a() { return myVar; }`,
			expectedResult: parser.ReturnStatementResult{
				Expression: &parser.VariableValue{
					Type:       "identifier",
					Expression: "myVar",
					Data:       "myVar",
				},
			},
		},
		{
			name: "返回箭头函数",
			code: `function a() { return () => {}; }`,
			expectedResult: parser.ReturnStatementResult{
				Expression: &parser.VariableValue{
					Type:       "arrowFunction",
					Expression: "() => {}",
				},
			},
		},
		{
			name: "空的返回",
			code: `function a() { return; }`,
			expectedResult: parser.ReturnStatementResult{
				Expression: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			RunTest(t, tc.code, func() string {
				jsonBytes, err := json.Marshal(tc.expectedResult)
				assert.NoError(t, err)
				return string(jsonBytes)
			}(),
				func(result *parser.ParserResult) parser.ReturnStatementResult {
					// 假设每个测试用例的函数体内只有一个 return 语句
					assert.GreaterOrEqual(t, len(result.ReturnStatements), 1, "应至少找到一个 return 语句")
					return result.ReturnStatements[0]
				},
				func(result parser.ReturnStatementResult) ([]byte, error) {
					return json.Marshal(result)
				})
		})
	}
}
