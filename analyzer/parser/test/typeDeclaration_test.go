
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

func TestAnalyzeTypeDecl(t *testing.T) {
	testCases := []struct {
		name         string
		code         string
		expectedJSON string
	}{
		{
			name: "Simple Type Alias",
			code: `type MyString = string;`,
			expectedJSON: `{
				"identifier": "MyString",
				"raw": "type MyString = string;",
				"reference": {}
			}`,
		},
		{
			name: "Type Alias with Custom Type",
			code: `type UserResponse = Response<User>;`,
			expectedJSON: `{
				"identifier": "UserResponse",
				"raw": "type UserResponse = Response<User>;",
				"reference": {
					"Response": {
						"identifier": "Response",
						"location": ["UserResponse"],
						"isExtend": false
					},
					"User": {
						"identifier": "User",
						"location": ["UserResponse<>"],
						"isExtend": false
					}
				}
			}`,
		},
		{
			name: "Type Alias with Union",
			code: `type MyUnion = TypeA | TypeB;`,
			expectedJSON: `{
				"identifier": "MyUnion",
				"raw": "type MyUnion = TypeA | TypeB;",
				"reference": {
					"TypeA": {
						"identifier": "TypeA",
						"location": ["MyUnion"],
						"isExtend": false
					},
					"TypeB": {
						"identifier": "TypeB",
						"location": ["MyUnion"],
						"isExtend": false
					}
				}
			}`,
		},
	}

	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current working directory")
	dummyPath := filepath.Join(wd, "test.ts")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sourceFile := utils.ParseTypeScriptFile(dummyPath, tc.code)

			var typeNode *ast.TypeAliasDeclaration
			for _, stmt := range sourceFile.Statements.Nodes {
				if stmt.Kind == ast.KindTypeAliasDeclaration {
					typeNode = stmt.AsTypeAliasDeclaration()
					break
				}
			}

			assert.NotNil(t, typeNode, "TypeAliasDeclaration node should not be nil")

			result := parser.NewTypeDeclarationResult(typeNode.AsNode(), tc.code)
			result.AnalyzeTypeDecl(typeNode)

			// Marshal the result to JSON for comparison, ignoring the SourceLocation field.
			resultJSON, err := json.MarshalIndent(struct {
				Identifier string                            `json:"identifier"`
				Raw        string                            `json:"raw"`
				Reference  map[string]parser.TypeReference `json:"reference"`
			}{
				Identifier: result.Identifier,
				Raw:        result.Raw,
				Reference:  result.Reference,
			}, "", "	")
			assert.NoError(t, err, "Failed to marshal result to JSON")

			assert.JSONEq(t, tc.expectedJSON, string(resultJSON), "The generated JSON should match the expected JSON")
		})
	}
}
