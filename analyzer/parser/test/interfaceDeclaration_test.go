package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// TestAnalyzeInterfaces 测试分析接口声明的功能
func TestAnalyzeInterfaces(t *testing.T) {
	// expectedResult 定义了测试期望的结果结构体
	type expectedResult struct {
		Identifier string                          `json:"identifier"` // 接口的标识符
		Raw        string                          `json:"raw"`        // 原始代码文本
		Reference  map[string]parser.TypeReference `json:"reference"`  // 类型引用
	}

	// testCases 定义了一系列的测试用例
	testCases := []struct {
		name           string         // 测试用例名称
		code           string         // 需要被解析的代码
		expectedResult expectedResult // 期望的解析结果
	}{
		{
			name: "简单的接口",
			code: `interface MyInterface { name: string; age: number; }`,
			expectedResult: expectedResult{
				Identifier: "MyInterface",
				Raw:        "interface MyInterface { name: string; age: number; }",
				Reference:  map[string]parser.TypeReference{},
			},
		},
		{
			name: "带自定义类型的接口",
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
			name: "带继承的接口",
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
			name: "复杂的接口",
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
			name: "继承了工具类型的接口",
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

	// findNode 是一个辅助函数，用于在 AST 中查找第一个接口声明节点
	findNode := func(sourceFile *ast.SourceFile) *ast.InterfaceDeclaration {
		for _, stmt := range sourceFile.Statements.Nodes {
			if stmt.Kind == ast.KindInterfaceDeclaration {
				return stmt.AsInterfaceDeclaration()
			}
		}
		return nil
	}

	// testParser 是一个辅助函数，用于执行解析操作
	testParser := func(node *ast.InterfaceDeclaration, code string) *parser.InterfaceDeclarationResult {
		result := parser.NewInterfaceDeclarationResult(node.AsNode(), code)
		result.AnalyzeInterfaces(node)
		return result
	}

	// marshal 是一个辅助函数，用于将解析结果序列化为 JSON
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