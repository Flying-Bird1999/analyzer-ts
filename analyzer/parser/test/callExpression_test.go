package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

func TestAnalyzeCallExpression(t *testing.T) {
	testCases := []struct {
		name         string
		code         string
		expectedJSON string
	}{
		{
			name: "Simple function call",
			code: `myFunction();`,
			expectedJSON: `{
				"callChain": ["myFunction"],
				"arguments": [],
				"type": "call"
			}`,
		},
		{
			name: "Call with arguments",
			code: `myFunction(1, "hello", true, myVar);`,
			expectedJSON: `{
				"callChain": ["myFunction"],
				"arguments": [
					{"type": "number", "text": "1"},
					{"type": "string", "text": "\"hello\""},
					{"type": "boolean", "text": "true"},
					{"type": "identifier", "text": "myVar"}
				],
				"type": "call"
			}`,
		},
		{
			name: "Member access call",
			code: `myObj.myMethod();`,
			expectedJSON: `{
				"callChain": ["myObj", "myMethod"],
				"arguments": [],
				"type": "member"
			}`,
		},
		{
			name: "Chained member access call",
			code: `this.a.b.c(123);`,
			expectedJSON: `{
				"callChain": ["this", "a", "b", "c"],
				"arguments": [
					{"type": "number", "text": "123"}
				],
				"type": "member"
			}`,
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
			RunTest(t, tc.code, tc.expectedJSON, findNode, testParser, marshal)
		})
	}
}
