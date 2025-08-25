package parser_test

import (
	"encoding/json"
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"

	"github.com/stretchr/testify/assert"
)

type expectedAsExpressionResult struct {
	Raw            string                `json:"raw"`
	SourceLocation parser.SourceLocation `json:"sourceLocation"`
}

func TestAsExpressions(t *testing.T) {
	testCases := []struct {
		name     string
		code     string
		expected []expectedAsExpressionResult
	}{
		{
			name: "基本的 as 表达式",
			code: `const a = b as string;`,
			expected: []expectedAsExpressionResult{
				{
					Raw: "const a = b as string;",
					SourceLocation: parser.SourceLocation{
						Start: parser.NodePosition{Line: 1, Column: 10},
						End:   parser.NodePosition{Line: 1, Column: 21},
					},
				},
			},
		},
		{
			name: "复杂的 as 表达式",
			code: `const a = (b + c) as number;`,
			expected: []expectedAsExpressionResult{
				{
					Raw: "const a = (b + c) as number;",
					SourceLocation: parser.SourceLocation{
						Start: parser.NodePosition{Line: 1, Column: 10},
						End:   parser.NodePosition{Line: 1, Column: 27},
					},
				},
			},
		},
		{
			name: "嵌套的 as 表达式",
			code: `const a = b as string as unknown;`,
			expected: []expectedAsExpressionResult{
				{
					Raw: "const a = b as string as unknown;",
					SourceLocation: parser.SourceLocation{
						Start: parser.NodePosition{Line: 1, Column: 10},
						End:   parser.NodePosition{Line: 1, Column: 32},
					},
				},
				{
					Raw: "const a = b as string as unknown;",
					SourceLocation: parser.SourceLocation{
						Start: parser.NodePosition{Line: 1, Column: 10},
						End:   parser.NodePosition{Line: 1, Column: 21},
					},
				},
			},
		},
		{
			name:     "没有 as 表达式",
			code:     `const a = b;`,
			expected: []expectedAsExpressionResult{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			RunTest(t, tc.code, func() string {
				jsonBytes, err := json.Marshal(tc.expected)
				assert.NoError(t, err)
				return string(jsonBytes)
			}(),
				func(result *parser.ParserResult) []expectedAsExpressionResult {
					cleanedActuals := []expectedAsExpressionResult{}
					for _, asExpr := range result.ExtractedNodes.AsExpressions {
						cleanedActuals = append(cleanedActuals, expectedAsExpressionResult{
							Raw:            asExpr.Raw,
							SourceLocation: asExpr.SourceLocation,
						})
					}
					return cleanedActuals
				},
				func(result []expectedAsExpressionResult) ([]byte, error) {
					return json.Marshal(result)
				})
		})
	}
}
