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

func TestAnalyzeImportDeclaration(t *testing.T) {
	testCases := []struct {
		name         string
		code         string
		expectedJSON string
	}{
		{
			name: "Default Import",
			code: "import Bird from './type2';",
			expectedJSON: `{
				"importModules": [
					{
						"importModule": "default",
						"type": "default",
						"identifier": "Bird"
					}
				],
				"raw": "import Bird from './type2';",
				"source": "./type2"
			}`,
		},
		{
			name: "Namespace Import",
			code: "import * as allTypes from './type';",
			expectedJSON: `{
				"importModules": [
					{
						"importModule": "allTypes",
						"type": "namespace",
						"identifier": "allTypes"
					}
				],
				"raw": "import * as allTypes from './type';",
				"source": "./type"
			}`,
		},
		{
			name: "Named Imports",
			code: "import { School, Teacher } from './school';",
			expectedJSON: `{
				"importModules": [
					{
						"importModule": "School",
						"type": "named",
						"identifier": "School"
					},
					{
						"importModule": "Teacher",
						"type": "named",
						"identifier": "Teacher"
					}
				],
				"raw": "import { School, Teacher } from './school';",
				"source": "./school"
			}`,
		},
		{
			name: "Named Imports with Alias",
			code: "import { School, School2 as NewSchool } from './school';",
			expectedJSON: `{
				"importModules": [
					{
						"importModule": "School",
						"type": "named",
						"identifier": "School"
					},
					{
						"importModule": "School2",
						"type": "named",
						"identifier": "NewSchool"
					}
				],
				"raw": "import { School, School2 as NewSchool } from './school';",
				"source": "./school"
			}`,
		},
		{
			name: "Side Effect Import",
			code: "import './setup';",
			expectedJSON: `{
				"importModules": [],
				"raw": "import './setup';",
				"source": "./setup"
			}`,
		},
	}

	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current working directory")
	// The parser requires an absolute path, so we create a dummy one.
	dummyPath := filepath.Join(wd, "test.ts")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sourceFile := utils.ParseTypeScriptFile(dummyPath, tc.code)
			rootNode := sourceFile.AsNode()

			var importNode *ast.ImportDeclaration
			rootNode.ForEachChild(func(child *ast.Node) bool {
				if child.Kind == ast.KindImportDeclaration {
					importNode = child.AsImportDeclaration()
					return true // Stop traversal
				}
				return false // Continue traversal
			})

			assert.NotNil(t, importNode, "Import node should not be nil")

			// Call the exported functions from the parser package
			result := parser.NewImportDeclarationResult()
			result.AnalyzeImportDeclaration(importNode, tc.code)

			// Marshal the result to JSON for comparison, ignoring the SourceLocation field.
			resultJSON, err := json.MarshalIndent(struct {
				ImportModules []parser.ImportModule `json:"importModules"`
				Raw           string                `json:"raw"`
				Source        string                `json:"source"`
			}{
				ImportModules: result.ImportModules,
				Raw:           result.Raw,
				Source:        result.Source,
			}, "", "	")
			assert.NoError(t, err, "Failed to marshal result to JSON")

			// Compare the actual JSON with the expected JSON.
			assert.JSONEq(t, tc.expectedJSON, string(resultJSON), "The generated JSON should match the expected JSON")
		})
	}
}
