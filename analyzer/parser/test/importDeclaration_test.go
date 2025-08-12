package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
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

	findNode := func(sourceFile *ast.SourceFile) *ast.ImportDeclaration {
		for _, stmt := range sourceFile.Statements.Nodes {
			if stmt.Kind == ast.KindImportDeclaration {
				return stmt.AsImportDeclaration()
			}
		}
		return nil
	}

	testParser := func(node *ast.ImportDeclaration, code string) *parser.ImportDeclarationResult {
		result := parser.NewImportDeclarationResult()
		result.AnalyzeImportDeclaration(node, code)
		return result
	}

	marshal := func(result *parser.ImportDeclarationResult) ([]byte, error) {
		return json.MarshalIndent(struct {
			ImportModules []parser.ImportModule `json:"importModules"`
			Raw           string                `json:"raw"`
			Source        string                `json:"source"`
		}{
			ImportModules: result.ImportModules,
			Raw:           result.Raw,
			Source:        result.Source,
		}, "", "\t")
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			RunTest(t, tc.code, tc.expectedJSON, findNode, testParser, marshal)
		})
	}
}
