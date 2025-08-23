package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"

	"github.com/stretchr/testify/assert"
)

type expectedAnyResult struct {
	SourceLocation parser.SourceLocation `json:"sourceLocation"`
	Raw            string                `json:"raw"`
}

func TestAnyDeclarations(t *testing.T) {
	testCases := []struct {
		name     string
		code     string
		expected []expectedAnyResult
	}{
		{
			name: "单个 any 类型",
			code: `let a: any;`,
			expected: []expectedAnyResult{
				{
					SourceLocation: parser.SourceLocation{
						Start: parser.NodePosition{Line: 1, Column: 7},
						End:   parser.NodePosition{Line: 1, Column: 10},
					},
					Raw: "let a: any;",
				},
			},
		},
		{
			name: "函数参数中的 any 类型",
			code: `function greet(name: any) { console.log(name); }`,
			expected: []expectedAnyResult{
				{
					SourceLocation: parser.SourceLocation{
						Start: parser.NodePosition{Line: 1, Column: 21},
						End:   parser.NodePosition{Line: 1, Column: 24},
					},
					Raw: "function greet(name: any) { console.log(name); }",
				},
			},
		},
		{
			name: "函数返回类型中的 any 类型",
			code: `function getData(): any { return {}; }`,
			expected: []expectedAnyResult{
				{
					SourceLocation: parser.SourceLocation{
						Start: parser.NodePosition{Line: 1, Column: 20},
						End:   parser.NodePosition{Line: 1, Column: 23},
					},
					Raw: "function getData(): any { return {}; }",
				},
			},
		},
		{
			name: "多个 any 类型",
			code: `let a: any; function b(c: any): any { return c; }`,
			expected: []expectedAnyResult{
				{
					SourceLocation: parser.SourceLocation{
						Start: parser.NodePosition{Line: 1, Column: 7},
						End:   parser.NodePosition{Line: 1, Column: 10},
					},
					Raw: "let a: any; function b(c: any): any { return c; }",
				},
				{
					SourceLocation: parser.SourceLocation{
						Start: parser.NodePosition{Line: 1, Column: 26},
						End:   parser.NodePosition{Line: 1, Column: 29},
					},
					Raw: "let a: any; function b(c: any): any { return c; }",
				},
				{
					SourceLocation: parser.SourceLocation{
						Start: parser.NodePosition{Line: 1, Column: 32},
						End:   parser.NodePosition{Line: 1, Column: 35},
					},
					Raw: "let a: any; function b(c: any): any { return c; }",
				},
			},
		},
		{
			name:     "没有 any 类型",
			code:     `let a: string; function b(c: number): string { return "hello"; }`,
			expected: []expectedAnyResult{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			RunTest(t, tc.code, func() string {
				jsonBytes, err := json.Marshal(tc.expected)
				assert.NoError(t, err)
				return string(jsonBytes)
			}(),
				func(result *parser.ParserResult) []expectedAnyResult {
					cleanedActuals := []expectedAnyResult{}
					for _, anyInfo := range result.AnyDeclarations {
						cleanedActuals = append(cleanedActuals, expectedAnyResult{
							SourceLocation: anyInfo.SourceLocation,
							Raw:            anyInfo.Raw,
						})
					}
					return cleanedActuals
				},
				func(result []expectedAnyResult) ([]byte, error) {
					return json.Marshal(result)
				})
		})
	}
}
