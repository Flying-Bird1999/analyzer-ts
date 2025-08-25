package parser_test

import (
	"encoding/json"
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
)

// TestAnalyzeExportAssignment 测试分析默认导出（export default）的功能
func TestAnalyzeExportAssignment(t *testing.T) {
	// expectedResult 定义了测试期望的结果结构体
	type expectedResult struct {
		Raw        string `json:"raw"`        // 原始代码文本
		Expression string `json:"expression"` // 导出的表达式
	}

	// testCases 定义了一系列的测试用例
	testCases := []struct {
		name           string         // 测试用例名称
		code           string         // 需要被解析的代码
		expectedResult expectedResult // 期望的解析结果
	}{
		{
			name: "导出一个标识符",
			code: "export default myVar;",
			expectedResult: expectedResult{
				Raw:        "export default myVar;",
				Expression: "myVar",
			},
		},
		{
			name: "导出一个函数调用",
			code: "export default myFunction();",
			expectedResult: expectedResult{
				Raw:        "export default myFunction();",
				Expression: "myFunction()",
			},
		},
	}

	// extractFn 定义了如何从完整的解析结果中提取我们关心的部分
	extractFn := func(result *parser.ParserResult) parser.ExportAssignmentResult {
		if len(result.ExportAssignments) > 0 {
			return result.ExportAssignments[0]
		}
		return parser.ExportAssignmentResult{}
	}

	// marshalFn 定义了如何将提取出的结果序列化为 JSON
	marshalFn := func(result parser.ExportAssignmentResult) ([]byte, error) {
		return json.MarshalIndent(struct {
			Raw        string `json:"raw"`
			Expression string `json:"expression"`
		}{
			Raw:        result.Raw,
			Expression: result.Expression,
		}, "", "\t")
	}

	// 遍历所有测试用例并执行测试
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 将期望结果序列化为 JSON
			expectedJSON, err := json.MarshalIndent(tc.expectedResult, "", "\t")
			if err != nil {
				t.Fatalf("无法将期望结果序列化为 JSON: %v", err)
			}
			// 运行测试
			RunTest(t, tc.code, string(expectedJSON), extractFn, marshalFn)
		})
	}
}
