
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

func TestAnalyzeInterfaces(t *testing.T) {
	testCases := []struct {
		name         string
		code         string
		expectedJSON string
	}{
		{
			name: "Simple Interface",
			code: `interface MyInterface { name: string; age: number; }`,
			expectedJSON: `{
				"identifier": "MyInterface",
				"raw": "interface MyInterface { name: string; age: number; }",
				"reference": {}
			}`,
		},
		{
			name: "Interface with Custom Type",
			code: `interface MyInterface { user: User; }`,
			expectedJSON: `{
				"identifier": "MyInterface",
				"raw": "interface MyInterface { user: User; }",
				"reference": {
					"User": {
						"identifier": "User",
						"location": ["MyInterface.user"],
						"isExtend": false
					}
				}
			}`,
		},
		{
			name: "Interface with Extends",
			code: `interface MyInterface extends AnotherInterface { id: number; }`,
			expectedJSON: `{
				"identifier": "MyInterface",
				"raw": "interface MyInterface extends AnotherInterface { id: number; }",
				"reference": {
					"AnotherInterface": {
						"identifier": "AnotherInterface",
						"location": [""],
						"isExtend": true
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

			var interfaceNode *ast.InterfaceDeclaration
			for _, stmt := range sourceFile.Statements.Nodes {
				if stmt.Kind == ast.KindInterfaceDeclaration {
					interfaceNode = stmt.AsInterfaceDeclaration()
					break
				}
			}

			assert.NotNil(t, interfaceNode, "Interface node should not be nil")

			result := parser.NewInterfaceDeclarationResult(interfaceNode.AsNode(), tc.code)
			result.AnalyzeInterfaces(interfaceNode)

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
