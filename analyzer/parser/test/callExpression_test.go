package parser_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/stretchr/testify/assert"
)

// expectedCallResult 是一个简化的结构，仅用于在测试中断言核心字段。
// 这样可以忽略像 SourceLocation 这样容易变动且与核心逻辑无关的字段。
type expectedCallResult struct {
	Expression      *parser.VariableValue              `json:"expression"`
	CallChain       []string                           `json:"callChain"`
	Arguments       []*parser.VariableValue            `json:"arguments"`
	InlineFunctions []parser.FunctionDeclarationResult `json:"inlineFunctions,omitempty"` // omitempty 用于不测试此字段的旧用例
	Raw             string                             `json:"raw"`
}

func TestCallExpression(t *testing.T) {
	testCases := []struct {
		name               string
		code               string
		expectedCalls      []expectedCallResult
		expectedImportDecl int // 用于测试动态导入
	}{
		{
			name: "简单的函数调用",
			code: `foo(1, 'bar');`,
			expectedCalls: []expectedCallResult{
				{
					Expression: &parser.VariableValue{
						Type:       "identifier",
						Expression: "foo",
						Data:       "foo",
					},
					CallChain: []string{"foo"},
					Arguments: []*parser.VariableValue{
						{
							Type:       "numericLiteral",
							Expression: "1",
							Data:       "1",
						},
						{
							Type:       "stringLiteral",
							Expression: "'bar'",
							Data:       "bar",
						},
					},
					Raw: "foo(1, 'bar')",
				},
			},
		},
		{
			name: "成员方法调用",
			code: `myObj.method.call(null, a, b);`,
			expectedCalls: []expectedCallResult{
				{
					Expression: &parser.VariableValue{
						Type:       "propertyAccess",
						Expression: "myObj.method.call",
					},
					CallChain: []string{"myObj", "method", "call"},
					Arguments: []*parser.VariableValue{
						{
							Type:       "other",
							Expression: "null",
						},
						{
							Type:       "identifier",
							Expression: "a",
							Data:       "a",
						},
						{
							Type:       "identifier",
							Expression: "b",
							Data:       "b",
						},
					},
					Raw: "myObj.method.call(null, a, b)",
				},
			},
		},
		{
			name: "无参数调用",
			code: `run();`,
			expectedCalls: []expectedCallResult{
				{
					Expression: &parser.VariableValue{
						Type:       "identifier",
						Expression: "run",
						Data:       "run",
					},
					CallChain: []string{"run"},
					Arguments:  []*parser.VariableValue{},
					Raw:        "run()",
				},
			},
		},
		{
			name: "复杂参数调用",
			code: `register({ user: 'test' }, () => { return true; });`,
			expectedCalls: []expectedCallResult{
				{
					Expression: &parser.VariableValue{
						Type:       "identifier",
						Expression: "register",
						Data:       "register",
					},
					CallChain: []string{"register"},
					Arguments: []*parser.VariableValue{
						{
							Type:       "objectLiteral",
							Expression: "{ user: 'test' }",
						},
						{
							Type:       "arrowFunction",
							Expression: "() => { return true; }",
						},
					},
					InlineFunctions: []parser.FunctionDeclarationResult{
						{
							Identifier: "", // 匿名函数
							Parameters: []parser.ParameterResult{},
						},
					},
					Raw: "register({ user: 'test' }, () => { return true; })",
				},
			},
		},
		{
			name: "被调用者是函数调用",
			code: `getHandler()();`,
			// 注意：顺序已根据深度优先遍历的实际结果调整
			expectedCalls: []expectedCallResult{
				{
					Expression: &parser.VariableValue{
						Type:       "callExpression",
						Expression: "getHandler()",
					},
					CallChain:  []string{"getHandler()"},
					Arguments:  []*parser.VariableValue{},
					Raw:        "getHandler()()",
				},
				{
					Expression: &parser.VariableValue{
						Type:       "identifier",
						Expression: "getHandler",
						Data:       "getHandler",
					},
					CallChain:  []string{"getHandler"},
					Arguments:  []*parser.VariableValue{},
					Raw:        "getHandler()",
				},
			},
		},
		{
			name:               "独立的动态导入",
			code:               `import('./module');`,
			expectedCalls:      []expectedCallResult{},
			expectedImportDecl: 1, // 期望生成一个导入声明
		},
		{
			name: "带有内联函数的Hook调用",
			code: `useEffect(() => { console.log("mounted"); });`,
			expectedCalls: []expectedCallResult{
				{
					Expression: &parser.VariableValue{
						Type:       "identifier",
						Expression: "useEffect",
						Data:       "useEffect",
					},
					CallChain: []string{"useEffect"},
					Arguments: []*parser.VariableValue{
						{
							Type:       "arrowFunction",
							Expression: "() => { console.log(\"mounted\"); }",
						},
					},
					InlineFunctions: []parser.FunctionDeclarationResult{
						{
							Identifier: "", // 匿名函数
							Parameters: []parser.ParameterResult{},
						},
					},
					Raw: "useEffect(() => { console.log(\"mounted\"); })",
				},
				{
					Expression: &parser.VariableValue{
						Type:       "propertyAccess",
						Expression: "console.log",
					},
					CallChain: []string{"console", "log"},
					Arguments: []*parser.VariableValue{
						{
							Type:       "stringLiteral",
							Expression: "\"mounted\"",
							Data:       "mounted",
						},
					},
					Raw: "console.log(\"mounted\")",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 检查调用表达式
			RunTest(t, tc.code, func() string {
				jsonBytes, err := json.Marshal(tc.expectedCalls)
				assert.NoError(t, err)
				return string(jsonBytes)
			}(),
				func(result *parser.ParserResult) []expectedCallResult {
					cleanedActuals := []expectedCallResult{}
					for _, callExpr := range result.CallExpressions {
						// 为了测试稳定性，我们只比较关键字段
						cleanedFuncs := []parser.FunctionDeclarationResult{}
						for _, f := range callExpr.InlineFunctions {
							cleanedFuncs = append(cleanedFuncs, parser.FunctionDeclarationResult{
								Identifier: f.Identifier,
								Parameters: f.Parameters,
							})
						}

						cleanedActuals = append(cleanedActuals, expectedCallResult{
							Expression:      callExpr.Expression,
							CallChain:       callExpr.CallChain,
							Arguments:       callExpr.Arguments,
							InlineFunctions: cleanedFuncs,
							Raw:             strings.TrimSpace(callExpr.Raw),
						})
					}
					return cleanedActuals
				},
				func(result []expectedCallResult) ([]byte, error) {
					return json.Marshal(result)
				})

			// 如果需要，检查动态导入声明的数量
			if tc.expectedImportDecl > 0 {
				wd, err := os.Getwd()
				assert.NoError(t, err)
				dummyPath := filepath.Join(wd, "test.ts")

				p, err := parser.NewParserFromSource(dummyPath, tc.code)
				assert.NoError(t, err)
				p.Traverse()
				assert.Equal(t, tc.expectedImportDecl, len(p.Result.ImportDeclarations), "动态导入声明的数量应匹配")
			}
		})
	}
}
