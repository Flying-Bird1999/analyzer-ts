package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

func TestAnalyzeExportAssignment(t *testing.T) {
	type expectedResult struct {
		Raw        string `json:"raw"`
		Expression string `json:"expression"`
	}

	testCases := []struct {
		name           string
		code           string
		expectedResult expectedResult
	}{
		{
			name: "Export default identifier",
			code: "export default myVar;",
			expectedResult: expectedResult{
				Raw:        "export default myVar;",
				Expression: "myVar",
			},
		},
		{
			name: "Export default function call",
			code: "export default myFunction();",
			expectedResult: expectedResult{
				Raw:        "export default myFunction();",
				Expression: "myFunction()",
			},
		},
		{
			name: "Export default object literal",
			code: "export default { key: 'value' };",
			expectedResult: expectedResult{
				Raw:        "export default { key: 'value' };",
				Expression: "{ key: 'value' }",
			},
		},
	}

	findNode := func(sourceFile *ast.SourceFile) *ast.ExportAssignment {
		for _, stmt := range sourceFile.Statements.Nodes {
			if stmt.Kind == ast.KindExportAssignment {
				return stmt.AsExportAssignment()
			}
		}
		return nil
	}

	testParser := func(node *ast.ExportAssignment, code string) *parser.ExportAssignmentResult {
		result := parser.NewExportAssignmentResult(node)
		result.AnalyzeExportAssignment(node, code)
		return result
	}

	marshal := func(result *parser.ExportAssignmentResult) ([]byte, error) {
		return json.MarshalIndent(struct {
			Raw        string `json:"raw"`
			Expression string `json:"expression"`
		}{
			Raw:        result.Raw,
			Expression: result.Expression,
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