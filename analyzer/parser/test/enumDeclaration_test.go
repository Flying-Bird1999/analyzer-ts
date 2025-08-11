
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

func TestNewEnumDeclarationResult(t *testing.T) {
	testCases := []struct {
		name         string
		code         string
		expectedJSON string
	}{
		{
			name: "Simple Enum",
			code: `enum Color { Red, Green, Blue }`,
			expectedJSON: `{
				"identifier": "Color",
				"raw": "enum Color { Red, Green, Blue }"
			}`,
		},
		{
			name: "Enum with Initializer",
			code: `enum Direction { Up = 1, Down, Left, Right }`,
			expectedJSON: `{
				"identifier": "Direction",
				"raw": "enum Direction { Up = 1, Down, Left, Right }"
			}`,
		},
	}

	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current working directory")
	dummyPath := filepath.Join(wd, "test.ts")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sourceFile := utils.ParseTypeScriptFile(dummyPath, tc.code)

			var enumNode *ast.EnumDeclaration
			for _, stmt := range sourceFile.Statements.Nodes {
				if stmt.Kind == ast.KindEnumDeclaration {
					enumNode = stmt.AsEnumDeclaration()
					break
				}
			}

			assert.NotNil(t, enumNode, "Enum node should not be nil")

			result := parser.NewEnumDeclarationResult(enumNode, tc.code)

			// Marshal the result to JSON for comparison, ignoring the SourceLocation field.
			resultJSON, err := json.MarshalIndent(struct {
				Identifier string `json:"identifier"`
				Raw        string `json:"raw"`
			}{
				Identifier: result.Identifier,
				Raw:        result.Raw,
			}, "", "	")
			assert.NoError(t, err, "Failed to marshal result to JSON")

			assert.JSONEq(t, tc.expectedJSON, string(resultJSON), "The generated JSON should match the expected JSON")
		})
	}
}
