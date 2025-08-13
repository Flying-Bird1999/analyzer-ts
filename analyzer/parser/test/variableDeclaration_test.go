package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

func TestNewVariableDeclaration(t *testing.T) {
	type expectedResult struct {
		Exported    bool                         `json:"exported"`
		Kind        parser.DeclarationKind       `json:"kind"`
		Source      *parser.VariableValue        `json:"source,omitempty"`
		Declarators []*parser.VariableDeclarator `json:"declarators"`
	}

	testCases := []struct {
		name           string
		code           string
		expectedResult expectedResult
	}{
		{
			name: "Simple const declaration",
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
			name: "Export let declaration",
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
			name: "Object destructuring",
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
			name: "Object destructuring with computed property",
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
			expectedJSON, err := json.MarshalIndent(tc.expectedResult, "", "\t")
			if err != nil {
				t.Fatalf("Failed to marshal expected result to JSON: %v", err)
			}
			RunTest(t, tc.code, string(expectedJSON), findNode, testParser, marshal)
		})
	}
}
