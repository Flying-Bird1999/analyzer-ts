package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

func TestNewEnumDeclarationResult(t *testing.T) {
	type expectedResult struct {
		Identifier string `json:"identifier"`
		Raw        string `json:"raw"`
	}

	testCases := []struct {
		name           string
		code           string
		expectedResult expectedResult
	}{
		{
			name: "Simple Enum",
			code: `enum Color { Red, Green, Blue }`,
			expectedResult: expectedResult{
				Identifier: "Color",
				Raw:        "enum Color { Red, Green, Blue }",
			},
		},
		{
			name: "Enum with Initializer",
			code: `enum Direction { Up = 1, Down, Left, Right }`,
			expectedResult: expectedResult{
				Identifier: "Direction",
				Raw:        "enum Direction { Up = 1, Down, Left, Right }",
			},
		},
	}

	findNode := func(sourceFile *ast.SourceFile) *ast.EnumDeclaration {
		for _, stmt := range sourceFile.Statements.Nodes {
			if stmt.Kind == ast.KindEnumDeclaration {
				return stmt.AsEnumDeclaration()
			}
		}
		return nil
	}

	testParser := func(node *ast.EnumDeclaration, code string) *parser.EnumDeclarationResult {
		return parser.NewEnumDeclarationResult(node, code)
	}

	marshal := func(result *parser.EnumDeclarationResult) ([]byte, error) {
		return json.MarshalIndent(struct {
			Identifier string `json:"identifier"`
			Raw        string `json:"raw"`
		}{
			Identifier: result.Identifier,
			Raw:        result.Raw,
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
