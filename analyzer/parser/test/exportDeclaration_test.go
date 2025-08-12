package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

func TestAnalyzeExportDeclaration(t *testing.T) {
	testCases := []struct {
		name         string
		code         string
		expectedJSON string
	}{
		{
			name: "Named Export",
			code: "export { name1, name2 };",
			expectedJSON: `{
				"exportModules": [],
				"raw": "export { name1, name2 };",
				"source": "",
				"type": ""
			}`,
		},
		{
			name: "Re-export from module",
			code: "export { name1 } from \"./mod\";",
			expectedJSON: `{
				"exportModules": [],
				"raw": "export { name1 } from \"./mod\";",
				"source": "",
				"type": ""
			}`,
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
			RunTest(t, tc.code, tc.expectedJSON, findNode, testParser, marshal)
		})
	}
}
