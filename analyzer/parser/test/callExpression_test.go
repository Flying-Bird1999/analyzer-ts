package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

func TestAnalyzeCallExpression(t *testing.T) {
	type expectedResult struct {
		CallChain []string          `json:"callChain"`
		Arguments []parser.Argument `json:"arguments"`
		Type      string            `json:"type"`
	}

	testCases := []struct {
		name           string
		code           string
		expectedResult expectedResult
	}{
		{
			name: "Simple function call",
			code: `myFunction();`,
			expectedResult: expectedResult{
				CallChain: []string{"myFunction"},
				Arguments: []parser.Argument{},
				Type:      "call",
			},
		},
		{
			name: "Call with arguments",
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
			name: "Member access call",
			code: `myObj.myMethod();`,
			expectedResult: expectedResult{
				CallChain: []string{"myObj", "myMethod"},
				Arguments: []parser.Argument{},
				Type:      "member",
			},
		},
		{
			name: "Chained member access call",
			code: `this.a.b.c(123);`,
			expectedResult: expectedResult{
				CallChain: []string{"this", "a", "b", "c"},
				Arguments: []parser.Argument{
					{Type: "number", Text: "123"},
				},
				Type: "member",
			},
		},
	}

	findNode := func(sourceFile *ast.SourceFile) *ast.CallExpression {
		var callNode *ast.CallExpression
		var walk func(node *ast.Node)
		walk = func(node *ast.Node) {
			if callNode != nil {
				return
			}
			if node.Kind == ast.KindCallExpression {
				callNode = node.AsCallExpression()
				return
			}
			node.ForEachChild(func(child *ast.Node) bool {
				walk(child)
				return callNode != nil
			})
		}
		walk(sourceFile.AsNode())
		return callNode
	}

	testParser := func(node *ast.CallExpression, code string) *parser.CallExpression {
		result := parser.NewCallExpression(node, code)
		result.AnalyzeCallExpression(node, code)
		return result
	}

	marshal := func(result *parser.CallExpression) ([]byte, error) {
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expectedJSON, err := json.MarshalIndent(tc.expectedResult, "", "\t")
			if err != nil {
				t.Fatalf("Failed to marshal expected result to JSON: %v", err)
			}
			RunTest(t, tc.code, string(expectedJSON), findNode, testParser, marshal)
		})
	}
}
