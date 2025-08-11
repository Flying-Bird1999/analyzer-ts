

package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"main/analyzer/utils"
	"os"
	"path/filepath"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
	"github.com/stretchr/testify/assert"
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

	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current working directory")
	dummyPath := filepath.Join(wd, "test.ts")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sourceFile := utils.ParseTypeScriptFile(dummyPath, tc.code)

			var callNode *ast.CallExpression
			var walk func(node *ast.Node)
			walk = func(node *ast.Node) {
				// If we've already found the node, no need to keep walking
				if callNode != nil {
					return
				}
				if node.Kind == ast.KindCallExpression {
					callNode = node.AsCallExpression()
					return
				}
				node.ForEachChild(func(child *ast.Node) bool {
					walk(child)
					// Return true if we've found the node to stop further iteration
					return callNode != nil
				})
			}
			walk(sourceFile.AsNode())

			assert.NotNil(t, callNode, "CallExpression node should not be nil")

			result := parser.NewCallExpression(callNode, tc.code)
			result.AnalyzeCallExpression(callNode, tc.code)

			// Marshal the result to JSON for comparison, ignoring Raw and SourceLocation fields.
			resultJSON, err := json.MarshalIndent(struct {
				CallChain []string           `json:"callChain"`
				Arguments []parser.Argument `json:"arguments"`
				Type      string            `json:"type"`
			}{
				CallChain: result.CallChain,
				Arguments: result.Arguments,
				Type:      result.Type,
			}, "", "\t")
			assert.NoError(t, err, "Failed to marshal result to JSON")

			assert.JSONEq(t, tc.expectedJSON, string(resultJSON), "The generated JSON should match the expected JSON")
		})
	}
}
