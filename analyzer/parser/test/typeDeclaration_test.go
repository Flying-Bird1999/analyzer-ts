
package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

func TestAnalyzeTypeDecl(t *testing.T) {
	type expectedResult struct {
		Identifier string                            `json:"identifier"`
		Raw        string                            `json:"raw"`
		Reference  map[string]parser.TypeReference `json:"reference"`
	}

	testCases := []struct {
		name           string
		code           string
		expectedResult expectedResult
	}{
		{
			name: "Simple Type Alias",
			code: `type MyString = string;`,
			expectedResult: expectedResult{
				Identifier: "MyString",
				Raw:        "type MyString = string;",
				Reference:  map[string]parser.TypeReference{},
			},
		},
		{
			name: "Type Alias with Custom Type",
			code: `type UserResponse = Response<User>;`,
			expectedResult: expectedResult{
				Identifier: "UserResponse",
				Raw:        "type UserResponse = Response<User>;",
				Reference: map[string]parser.TypeReference{
					"Response": {
						Identifier: "Response",
						Location:   []string{"UserResponse"},
						IsExtend:   false,
					},
					"User": {
						Identifier: "User",
						Location:   []string{"UserResponse<>"},
						IsExtend:   false,
					},
				},
			},
		},
		{
			name: "Type Alias with Union",
			code: `type MyUnion = TypeA | TypeB;`,
			expectedResult: expectedResult{
				Identifier: "MyUnion",
				Raw:        "type MyUnion = TypeA | TypeB;",
				Reference: map[string]parser.TypeReference{
					"TypeA": {
						Identifier: "TypeA",
						Location:   []string{"MyUnion"},
						IsExtend:   false,
					},
					"TypeB": {
						Identifier: "TypeB",
						Location:   []string{"MyUnion"},
						IsExtend:   false,
					},
				},
			},
		},
		{
			name: "Mapped Type",
			code: `type MappedType = { [key in SupportedLanguages]?: string[] | string }`,
			expectedResult: expectedResult{
				Identifier: "MappedType",
				Raw:        "type MappedType = { [key in SupportedLanguages]?: string[] | string }",
				Reference: map[string]parser.TypeReference{
					"SupportedLanguages": {
						Identifier: "SupportedLanguages",
						Location:   []string{""},
						IsExtend:   false,
					},
				},
			},
		},
		{
			name: "Indexed Access Type",
			code: `type PersonName = Translations["name"];`,
			expectedResult: expectedResult{
				Identifier: "PersonName",
				Raw:        "type PersonName = Translations[\"name\"];",
				Reference: map[string]parser.TypeReference{
					"Translations": {
						Identifier: "Translations",
						Location:   []string{"PersonName"},
						IsExtend:   false,
					},
				},
			},
		},
		{
			name: "Type with String Key",
			code: `type A = { "name": string };`,
			expectedResult: expectedResult{
				Identifier: "A",
				Raw:        "type A = { \"name\": string };",
				Reference:  map[string]parser.TypeReference{},
			},
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
			expectedJSON, err := json.MarshalIndent(tc.expectedResult, "", "\t")
			if err != nil {
				t.Fatalf("Failed to marshal expected result to JSON: %v", err)
			}
			RunTest(t, tc.code, string(expectedJSON), findNode, testParser, marshal)
		})
	}
}
