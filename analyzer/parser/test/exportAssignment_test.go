package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
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
		{
			name: "导出一个对象字面量",
			code: "export default { key: 'value' };",
			expectedResult: expectedResult{
				Raw:        "export default { key: 'value' };",
				Expression: "{ key: 'value' }",
			},
		},
	}

	// findNode 是一个辅助函数，用于在 AST 中查找第一个默认导出节点
	findNode := func(sourceFile *ast.SourceFile) *ast.ExportAssignment {
		for _, stmt := range sourceFile.Statements.Nodes {
			if stmt.Kind == ast.KindExportAssignment {
				return stmt.AsExportAssignment()
			}
		}
		return nil
	}

	// testParser 是一个辅助函数，用于执行解析操作
	testParser := func(node *ast.ExportAssignment, code string) *parser.ExportAssignmentResult {
		result := parser.NewExportAssignmentResult(node)
		result.AnalyzeExportAssignment(node, code)
		return result
	}

	// marshal 是一个辅助函数，用于将解析结果序列化为 JSON
	marshal := func(result *parser.ExportAssignmentResult) ([]byte, error) {
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
			RunTest(t, tc.code, string(expectedJSON), findNode, testParser, marshal)
		})
	}
}
