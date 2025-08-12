package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

func TestAnalyzeInterfaces(t *testing.T) {
	type expectedResult struct {
		Identifier string                          `json:"identifier"`
		Raw        string                          `json:"raw"`
		Reference  map[string]parser.TypeReference `json:"reference"`
	}

	testCases := []struct {
		name           string
		code           string
		expectedResult expectedResult
	}{
		{
			name: "Simple Interface",
			code: `interface MyInterface { name: string; age: number; }`,
			expectedResult: expectedResult{
				Identifier: "MyInterface",
				Raw:        "interface MyInterface { name: string; age: number; }",
				Reference:  map[string]parser.TypeReference{},
			},
		},
		{
			name: "Interface with Custom Type",
			code: `interface MyInterface { user: User; }`,
			expectedResult: expectedResult{
				Identifier: "MyInterface",
				Raw:        "interface MyInterface { user: User; }",
				Reference: map[string]parser.TypeReference{
					"User": {
						Identifier: "User",
						Location:   []string{"MyInterface.user"},
						IsExtend:   false,
					},
				},
			},
		},
		{
			name: "Interface with Extends",
			code: `interface MyInterface extends AnotherInterface { id: number; }`,
			expectedResult: expectedResult{
				Identifier: "MyInterface",
				Raw:        "interface MyInterface extends AnotherInterface { id: number; }",
				Reference: map[string]parser.TypeReference{
					"AnotherInterface": {
						Identifier: "AnotherInterface",
						Location:   []string{""},
						IsExtend:   true,
					},
				},
			},
		},
		{
			name: "Complex Interface",
			code: `interface Class extends A {
				name: string;
				age: number | Class3;
				// 学校
				school: School;
				['class2']: Class2;
				pack: allTypes.Size;
			}`,
			expectedResult: expectedResult{
				Identifier: "Class",
				Raw: `interface Class extends A {
				name: string;
				age: number | Class3;
				// 学校
				school: School;
				['class2']: Class2;
				pack: allTypes.Size;
			}`,
				Reference: map[string]parser.TypeReference{
					"A": {
						Identifier: "A",
						Location:   []string{""},
						IsExtend:   true,
					},
					"Class3": {
						Identifier: "Class3",
						Location:   []string{"Class.age"},
						IsExtend:   false,
					},
					"School": {
						Identifier: "School",
						Location:   []string{"Class.school"},
						IsExtend:   false,
					},
					"Class2": {
						Identifier: "Class2",
						Location:   []string{"Class.class2"},
						IsExtend:   false,
					},
					"allTypes.Size": {
						Identifier: "allTypes.Size",
						Location:   []string{"Class.pack"},
						IsExtend:   false,
					},
				},
			},
		},
		{
			name: "Extends with Utility Type",
			code: `interface Class8 extends Omit<Class2, 'age'> {name:string}`,
			expectedResult: expectedResult{
				Identifier: "Class8",
				Raw:        "interface Class8 extends Omit<Class2, 'age'> {name:string}",
				Reference: map[string]parser.TypeReference{
					"Class2": {
						Identifier: "Class2",
						Location:   []string{""},
						IsExtend:   true,
					},
				},
			},
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
			Identifier string                          `json:"identifier"`
			Raw        string                          `json:"raw"`
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
