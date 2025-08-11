

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

	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current working directory")
	dummyPath := filepath.Join(wd, "test.ts")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sourceFile := utils.ParseTypeScriptFile(dummyPath, tc.code)

			var exportNode *ast.ExportDeclaration
			for _, stmt := range sourceFile.Statements.Nodes {
				if stmt.Kind == ast.KindExportDeclaration {
					exportNode = stmt.AsExportDeclaration()
					break
				}
			}

			assert.NotNil(t, exportNode, "Export node should not be nil")

			result := parser.NewExportDeclarationResult(exportNode)
			result.AnalyzeExportDeclaration(exportNode, tc.code)

			// Marshal the result to JSON for comparison, ignoring the SourceLocation field.
			resultJSON, err := json.MarshalIndent(struct {
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
			assert.NoError(t, err, "Failed to marshal result to JSON")

			assert.JSONEq(t, tc.expectedJSON, string(resultJSON), "The generated JSON should match the expected JSON")
		})
	}
}

