package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
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

	findNode := func(sourceFile *ast.SourceFile) *ast.VariableStatement {
		for _, stmt := range sourceFile.Statements.Nodes {
			if stmt.Kind == ast.KindVariableStatement {
				return stmt.AsVariableStatement()
			}
		}
		return nil
	}

	testParser := func(node *ast.VariableStatement, code string) *parser.VariableDeclaration {
		return parser.NewVariableDeclaration(node, code)
	}

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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			RunTest(t, tc.code, tc.expectedJSON, findNode, testParser, marshal)
		})
	}
}
