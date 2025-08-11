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

func TestNewVariableDeclaration(t *testing.T) {
	testCases := []struct {
		name         string
		code         string
		expectedJSON string
	}{
		{
			name: "Simple const declaration",
			code: `const myVar = "hello";`,
			expectedJSON: `{
				"exported": false,
				"kind": "const",
				"declarators": [
					{
						"identifier": "myVar",
						"initValue": {
							"type": "stringLiteral",
							"expression": "\"hello\"",
							"data": "hello"
						}
					}
				]
			}`,
		},
		{
			name: "Export let declaration",
			code: `export let myVar: number = 123;`,
			expectedJSON: `{
				"exported": true,
				"kind": "let",
				"declarators": [
					{
						"identifier": "myVar",
						"type": {
							"type": "typeNode",
							"expression": "number"
						},
						"initValue": {
							"type": "numericLiteral",
							"expression": "123",
							"data": "123"
						}
					}
				]
			}`,
		},
		{
			name: "Object destructuring",
			code: `const { a, b: myB } = myObj;`,
			expectedJSON: `{
				"exported": false,
				"kind": "const",
				"source": {
					"type": "identifier",
					"expression": "myObj",
					"data": "myObj"
				},
				"declarators": [
					{
						"identifier": "a",
						"propName": "a"
					},
					{
						"identifier": "myB",
						"propName": "b"
					}
				]
			}`,
		},
	}

	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current working directory")
	dummyPath := filepath.Join(wd, "test.ts")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sourceFile := utils.ParseTypeScriptFile(dummyPath, tc.code)

			var varNode *ast.VariableStatement
			for _, stmt := range sourceFile.Statements.Nodes {
				if stmt.Kind == ast.KindVariableStatement {
					varNode = stmt.AsVariableStatement()
					break
				}
			}

			assert.NotNil(t, varNode, "VariableStatement node should not be nil")

			result := parser.NewVariableDeclaration(varNode, tc.code)

			// Marshal the result to JSON for comparison, ignoring Raw and SourceLocation fields.
			resultJSON, err := json.MarshalIndent(struct {
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
			assert.NoError(t, err, "Failed to marshal result to JSON")

			assert.JSONEq(t, tc.expectedJSON, string(resultJSON), "The generated JSON should match the expected JSON")
		})
	}
}
