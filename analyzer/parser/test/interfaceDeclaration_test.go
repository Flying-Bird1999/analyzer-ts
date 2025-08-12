
package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
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

	findNode := func(sourceFile *ast.SourceFile) *ast.InterfaceDeclaration {
		for _, stmt := range sourceFile.Statements.Nodes {
			if stmt.Kind == ast.KindInterfaceDeclaration {
				return stmt.AsInterfaceDeclaration()
			}
		}
		return nil
	}

	testParser := func(node *ast.InterfaceDeclaration, code string) *parser.InterfaceDeclarationResult {
		result := parser.NewInterfaceDeclarationResult(node.AsNode(), code)
		result.AnalyzeInterfaces(node)
		return result
	}

	marshal := func(result *parser.InterfaceDeclarationResult) ([]byte, error) {
		return json.MarshalIndent(struct {
			Identifier string                            `json:"identifier"`
			Raw        string                            `json:"raw"`
			Reference  map[string]parser.TypeReference `json:"reference"`
		}{
			Identifier: result.Identifier,
			Raw:        result.Raw,
			Reference:  result.Reference,
		}, "", "\t")
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			RunTest(t, tc.code, tc.expectedJSON, findNode, testParser, marshal)
		})
	}
}
