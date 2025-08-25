package parser_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"

	"github.com/stretchr/testify/assert"
)

// testParameterResult 是一个临时结构体，用于在测试中模拟 ParameterResult 的 JSON 序列化行为。
// 它被定义在包级别，以便在 expectedResult 中使用。
type testParameterResult struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Raw          string `json:"raw"`
	Optional     bool   `json:"optional"`
	DefaultValue string `json:"defaultValue,omitempty"` // 添加 omitempty
	IsRest       bool   `json:"isRest"`
}

// TestFunctionDeclarations 使用表驱动模式测试所有与函数声明相关的场景。
// 这种方式使得添加新测试用例变得简单，并且风格与项目中其他测试保持一致。
func TestFunctionDeclarations(t *testing.T) {

	// expectedResult 定义了测试期望的结果结构体，与 parser.FunctionDeclarationResult 对应
	// 这样做可以避免在测试用例中手写 JSON 字符串，使得测试更具类型安全和可读性。
	type expectedResult struct {
		Identifier  string                `json:"identifier"`
		Exported    bool                  `json:"exported"`
		IsAsync     bool                  `json:"isAsync"`
		IsGenerator bool                  `json:"isGenerator"`
		Generics    []string              `json:"generics"`
		Parameters  []testParameterResult `json:"parameters"` // 使用临时结构体
		ReturnType  string                `json:"returnType"`
	}

	// testCases 定义了所有测试用例。
	testCases := []struct {
		name     string           // name 是测试用例的描述性中文名称。
		code     string           // code 是要被解析的 TypeScript 源码。
		expected []expectedResult // expected 是期望从源码中解析出的 Go 结构体结果。
	}{
		{
			name: "基础箭头函数",
			code: `const myFunc = (a: string, b: number) => a + b;`,
			expected: []expectedResult{
				{
					Identifier: "myFunc",
					Generics:   []string{},
					Parameters: []testParameterResult{
						{Name: "a", Type: "string", Raw: "a: string"},
						{Name: "b", Type: "number", Raw: "b: number"},
					},
					ReturnType: "",
				},
			},
		},
		{
			name: "导出的异步箭头函数",
			code: `export const fetchUsers = async (): Promise<User[]> => { return []; };`,
			expected: []expectedResult{
				{
					Identifier: "fetchUsers",
					Exported:   true,
					IsAsync:    true,
					Generics:   []string{},
					Parameters: []testParameterResult{},
					ReturnType: "Promise<User[]>",
				},
			},
		},
		{
			name: "包含复杂参数的函数声明",
			code: `function configure(port: number = 3000, options?: Config, ...args: any[]) {}`,
			expected: []expectedResult{
				{
					Identifier: "configure",
					Generics:   []string{},
					Parameters: []testParameterResult{
						{Name: "port", Type: "number", Raw: "port: number = 3000", DefaultValue: "3000"},
						{Name: "options", Type: "Config", Raw: "options?: Config", Optional: true},
						{Name: "args", Type: "any[]", Raw: "...args: any[]", IsRest: true},
					},
					ReturnType: "",
				},
			},
		},
		{
			name: "带泛型的箭头函数",
			code: `export const createGeneric = <T, K extends string>(arg1: T, arg2: K): [T, K] => [arg1, arg2];`,
			expected: []expectedResult{
				{
					Identifier: "createGeneric",
					Exported:   true,
					Generics:   []string{"T", "K extends string"},
					Parameters: []testParameterResult{
						{Name: "arg1", Type: "T", Raw: "arg1: T"},
						{Name: "arg2", Type: "K", Raw: "arg2: K"},
					},
					ReturnType: "[T, K]",
				},
			},
		},
		{
			name: "生成器函数",
			code: `function* idMaker() { var index = 0; while(true) yield index++; }`,
			expected: []expectedResult{
				{
					Identifier:  "idMaker",
					IsGenerator: true,
					Generics:    []string{},
					Parameters:  []testParameterResult{},
					ReturnType:  "",
				},
			},
		},
	}

	// 遍历并执行所有测试用例
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 1. 从源码进行解析
			wd, err := os.Getwd()
			assert.NoError(t, err)
			dummyPath := filepath.Join(wd, "test.ts")
			p, err := parser.NewParserFromSource(dummyPath, tc.code)
			assert.NoError(t, err)
			p.Traverse()

			// 2. 提取实际解析结果，并进行清理以方便比较
			actualResults := p.Result.FunctionDeclarations
			cleanedActuals := []expectedResult{}
			for _, r := range actualResults {
				// 清理泛型中的空格
				cleanedGenerics := []string{}
				for _, g := range r.Generics {
					cleanedGenerics = append(cleanedGenerics, strings.TrimSpace(g))
				}

				// 手动转换 ParameterResult 到 testParameterResult
				convertedParameters := []testParameterResult{}
				for _, param := range r.Parameters {
					convertedParameters = append(convertedParameters, testParameterResult{
						Name:         param.Name,
						Type:         param.Type,
						Raw:          strings.TrimSpace(param.Raw),
						Optional:     param.Optional,
						DefaultValue: param.DefaultValue,
						IsRest:       param.IsRest,
					})
				}

				cleanedActuals = append(cleanedActuals, expectedResult{
					Identifier:  r.Identifier,
					Exported:    r.Exported,
					IsAsync:     r.IsAsync,
					IsGenerator: r.IsGenerator,
					Generics:    cleanedGenerics,
					Parameters:  convertedParameters,
					ReturnType:  r.ReturnType,
				})
			}

			// 3. 将期望结果和实际结果都序列化为 JSON 进行比较
			expectedJSON, err := json.Marshal(tc.expected)
			assert.NoError(t, err)
			actualJSON, err := json.Marshal(cleanedActuals)
			assert.NoError(t, err)

			assert.JSONEq(t, string(expectedJSON), string(actualJSON), "函数解析结果与预期不符")
		})
	}
}
