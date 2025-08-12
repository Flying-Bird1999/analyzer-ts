
package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
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

	findNode := func(sourceFile *ast.SourceFile) *ast.TypeAliasDeclaration {
		for _, stmt := range sourceFile.Statements.Nodes {
			if stmt.Kind == ast.KindTypeAliasDeclaration {
				return stmt.AsTypeAliasDeclaration()
			}
		}
		return nil
	}

	testParser := func(node *ast.TypeAliasDeclaration, code string) *parser.TypeDeclarationResult {
		result := parser.NewTypeDeclarationResult(node.AsNode(), code)
		result.AnalyzeTypeDecl(node)
		return result
	}

	marshal := func(result *parser.TypeDeclarationResult) ([]byte, error) {
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
