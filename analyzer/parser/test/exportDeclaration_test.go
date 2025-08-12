package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

func TestAnalyzeExportDeclaration(t *testing.T) {
	type expectedResult struct {
		ExportModules []parser.ExportModule `json:"exportModules"`
		Raw           string                `json:"raw"`
		Source        string                `json:"source"`
		Type          string                `json:"type"`
	}

	testCases := []struct {
		name           string
		code           string
		expectedResult expectedResult
	}{
		{
			name: "Named Export",
			code: "export { name1, name2 as alias };",
			expectedResult: expectedResult{
				ExportModules: []parser.ExportModule{
					{ModuleName: "name1", Type: "named", Identifier: "name1"},
					{ModuleName: "name2", Type: "named", Identifier: "alias"},
				},
				Raw:    "export { name1, name2 as alias };",
				Source: "",
				Type:   "named-export",
			},
		},
		{
			name: "Re-export from module",
			code: `export { name1 } from "./mod";`,
			expectedResult: expectedResult{
				ExportModules: []parser.ExportModule{
					{ModuleName: "name1", Type: "named", Identifier: "name1"},
				},
				Raw:    "export { name1 } from \"./mod\";",
				Source: "./mod",
				Type:   "re-export",
			},
		},
		{
			name: "Wildcard re-export",
			code: `export * from "./mod";`,
			expectedResult: expectedResult{
				ExportModules: []parser.ExportModule{
					{ModuleName: "*", Type: "namespace", Identifier: "*"},
				},
				Raw:    "export * from \"./mod\";",
				Source: "./mod",
				Type:   "re-export",
			},
		},
		{
			name: "Namespace re-export",
			code: `export * as ns from "./mod";`,
			expectedResult: expectedResult{
				ExportModules: []parser.ExportModule{
					{ModuleName: "*", Type: "namespace", Identifier: "ns"},
				},
				Raw:    "export * as ns from \"./mod\";",
				Source: "./mod",
				Type:   "re-export",
			},
		},
	}

	findNode := func(sourceFile *ast.SourceFile) *ast.ExportDeclaration {
		for _, stmt := range sourceFile.Statements.Nodes {
			if stmt.Kind == ast.KindExportDeclaration {
				return stmt.AsExportDeclaration()
			}
		}
		return nil
	}

	testParser := func(node *ast.ExportDeclaration, code string) *parser.ExportDeclarationResult {
		result := parser.NewExportDeclarationResult(node)
		result.AnalyzeExportDeclaration(node, code)
		return result
	}

	marshal := func(result *parser.ExportDeclarationResult) ([]byte, error) {
		return json.MarshalIndent(struct {
			ExportModules []parser.ExportModule `json:"exportModules"`
			Raw           string                `json:"raw"`
			Source        string                `json:"source"`
			Type          string                `json:"type"`
		}{
			ExportModules: result.ExportModules,
			Raw:           result.Raw,
			Source:        result.Source,
			Type:          result.Type,
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