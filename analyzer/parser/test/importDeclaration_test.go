package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

func TestAnalyzeImportDeclaration(t *testing.T) {
	type expectedResult struct {
		ImportModules []parser.ImportModule `json:"importModules"`
		Raw           string                `json:"raw"`
		Source        string                `json:"source"`
	}

	testCases := []struct {
		name           string
		code           string
		expectedResult expectedResult
	}{
		{
			name: "Default Import",
			code: "import Bird from './type2';",
			expectedResult: expectedResult{
				ImportModules: []parser.ImportModule{
					{ImportModule: "default", Type: "default", Identifier: "Bird"},
				},
				Raw:    "import Bird from './type2';",
				Source: "./type2",
			},
		},
		{
			name: "Namespace Import",
			code: "import * as allTypes from './type';",
			expectedResult: expectedResult{
				ImportModules: []parser.ImportModule{
					{ImportModule: "allTypes", Type: "namespace", Identifier: "allTypes"},
				},
				Raw:    "import * as allTypes from './type';",
				Source: "./type",
			},
		},
		{
			name: "Named Imports",
			code: "import { School, Teacher } from './school';",
			expectedResult: expectedResult{
				ImportModules: []parser.ImportModule{
					{ImportModule: "School", Type: "named", Identifier: "School"},
					{ImportModule: "Teacher", Type: "named", Identifier: "Teacher"},
				},
				Raw:    "import { School, Teacher } from './school';",
				Source: "./school",
			},
		},
		{
			name: "Named Imports with Alias",
			code: "import { School, School2 as NewSchool } from './school';",
			expectedResult: expectedResult{
				ImportModules: []parser.ImportModule{
					{ImportModule: "School", Type: "named", Identifier: "School"},
					{ImportModule: "School2", Type: "named", Identifier: "NewSchool"},
				},
				Raw:    "import { School, School2 as NewSchool } from './school';",
				Source: "./school",
			},
		},
		{
			name: "Side Effect Import",
			code: "import './setup';",
			expectedResult: expectedResult{
				ImportModules: []parser.ImportModule{},
				Raw:           "import './setup';",
				Source:        "./setup",
			},
		},
		{
			name: "Default and Named Imports with Alias",
			code: "import Bird, { School, Teacher as t2 } from './type2';",
			expectedResult: expectedResult{
				ImportModules: []parser.ImportModule{
					{ImportModule: "default", Type: "default", Identifier: "Bird"},
					{ImportModule: "School", Type: "named", Identifier: "School"},
					{ImportModule: "Teacher", Type: "named", Identifier: "t2"},
				},
				Raw:    "import Bird, { School, Teacher as t2 } from './type2';",
				Source: "./type2",
			},
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
			expectedJSON, err := json.MarshalIndent(tc.expectedResult, "", "\t")
			if err != nil {
				t.Fatalf("Failed to marshal expected result to JSON: %v", err)
			}
			RunTest(t, tc.code, string(expectedJSON), findNode, testParser, marshal)
		})
	}
}
