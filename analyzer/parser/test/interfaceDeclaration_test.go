package parser_test

import (
	"encoding/json"
	"main/analyzer/parser"
	"testing"
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

	// extractFn 定义了如何从完整的解析结果中提取我们关心的部分
	extractFn := func(result *parser.ParserResult) parser.InterfaceDeclarationResult {
		// 从 map 中提取我们需要的那个接口进行测试
		// 注意：由于 map 遍历顺序不确定，这在有多个接口的测试中可能不稳定
		// 但在这里，每个测试用例都只定义了一个目标接口
		for _, iface := range result.InterfaceDeclarations {
			return iface
		}
		return parser.InterfaceDeclarationResult{}
	}

	// marshalFn 定义了如何将提取出的结果序列化为 JSON
	marshalFn := func(result parser.InterfaceDeclarationResult) ([]byte, error) {
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
			RunTest(t, tc.code, string(expectedJSON), extractFn, marshalFn)
		})
	}
}
