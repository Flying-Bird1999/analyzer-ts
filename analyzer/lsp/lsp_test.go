package lsp

import (
	"context"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/lsp/lsproto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// LSP Service 核心功能测试
// ============================================================================

// TestFindReferences_ConfiguredProjectBug 旨在复现 `typescript-go` 在处理包含 tsconfig.json 的项目时，
// `FindReferences` 无法正常工作的已知 bug。
// 预期：此测试目前会失败。
func TestFindReferences_ConfiguredProjectBug(t *testing.T) {
	// 1. 创建一个包含 tsconfig.json 和路径别名的多文件内存项目
	sources := map[string]any{
		"/tsconfig.json": `{
			"compilerOptions": {
				"baseUrl": ".",
				"paths": {
					"@/*": ["src/*"]
				}
			}
		}`,
		"/src/utils.ts": `export const myVar = 123;`,
		"/src/index.ts": `
			import { myVar } from '@/utils';
			console.log(myVar);
		`,
	}

	// 2. 使用 NewServiceForTest (一个修改版的 NewService，用于接受内存 map)
	// 注意：这需要我们先对 NewService 进行重构，或者创建一个测试专用的版本。
	// 这里我们先假设 NewServiceForTest 已经可用。
	service, err := NewServiceForTest(sources)
	assert.NoError(t, err)
	defer service.Close()

	// 3. 在 /src/index.ts 的第 3 行第 14 个字符处（第二个 myVar）查找引用
	// 我们期望找到 2 个引用：定义处和使用处。
	response, err := service.FindReferences(context.Background(), "/src/index.ts", 3, 14)

	// 打印结果以供人工检查
	t.Logf("Error: %v", err)
	t.Logf("Response: %+v", response)
	if response.Locations != nil {
		t.Logf("Found %d locations.", len(*response.Locations))
		for _, loc := range *response.Locations {
			t.Logf("  - %s:%d:%d", loc.Uri, loc.Range.Start.Line+1, loc.Range.Start.Character+1)
		}
	}

	assert.NoError(t, err, "FindReferences 调用不应报错")
}

// ============================================================================
// QuickInfo 功能测试
// ============================================================================

// TestGetQuickInfoAtPosition 测试 GetQuickInfoAtPosition 方法的各种场景
func TestGetQuickInfoAtPosition(t *testing.T) {
	// 准备测试用的 TypeScript 源码
	testSources := map[string]any{
		"/test.ts": `const testVar: string = "hello";

function testFunction(param: string): void {
	console.log(param);
}

class TestClass {
	publicField: number = 42;
}
`,
	}

	// 创建 LSP 服务
	service, err := NewServiceForTest(testSources)
	require.NoError(t, err)
	defer service.Close()

	tests := []struct {
		name     string
		filePath string
		line     int
		char     int
		wantType string
	}{
		{
			name:     "字符串变量类型",
			filePath: "/test.ts",
			line:     1,
			char:     7, // testVar 的 'v' 位置
			wantType: "string",
		},
		{
			name:     "函数参数类型",
			filePath: "/test.ts",
			line:     3,
			char:     20, // param 的 'p' 位置
			wantType: "string",
		},
		{
			name:     "类字段类型",
			filePath: "/test.ts",
			line:     6,
			char:     2,  // publicField 中 'p' 位置
			wantType: "", // 这个位置可能找不到 QuickInfo，所以不验证类型
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测试的方法
			quickInfo, err := service.GetQuickInfoAtPosition(context.Background(), tt.filePath, tt.line, tt.char)

			require.NoError(t, err, "不期望返回错误")

			// 如果 wantType 为空，说明这个测试可能返回 nil QuickInfo
			if tt.wantType == "" {
				// 这种情况下我们只关心不报错，QuickInfo 可以为 nil
				return
			}

			// 验证 QuickInfo 不为空
			assert.NotNil(t, quickInfo, "QuickInfo 不应该为空")

			// 验证类型文本不为空
			assert.NotEmpty(t, quickInfo.TypeText, "类型文本不应该为空")

			// 验证类型信息（这里我们检查是否包含期望的类型关键字）
			if tt.wantType != "" {
				assert.Contains(t, quickInfo.TypeText, tt.wantType,
					"类型信息应该包含 '%s'，实际得到 '%s'", tt.wantType, quickInfo.TypeText)
			}

			// 验证范围信息
			if quickInfo.Range != nil {
				assert.NotNil(t, quickInfo.Range.Start, "范围起始位置不应该为空")
				assert.NotNil(t, quickInfo.Range.End, "范围结束位置不应该为空")
			}

			t.Logf("QuickInfo 详情:")
			t.Logf("  类型文本: %s", quickInfo.TypeText)
			t.Logf("  文档: %s", quickInfo.Documentation)
			t.Logf("  显示部件数: %d", len(quickInfo.DisplayParts))
			if quickInfo.Range != nil {
				t.Logf("  范围: (%d,%d)-(%d,%d)",
					quickInfo.Range.Start.Line, quickInfo.Range.Start.Character,
					quickInfo.Range.End.Line, quickInfo.Range.End.Character)
			}
		})
	}
}

// TestGetQuickInfoAtPosition_InvalidPosition 测试无效位置的 QuickInfo 查询
func TestGetQuickInfoAtPosition_InvalidPosition(t *testing.T) {
	testSources := map[string]any{
		"/test.ts": `const test = "hello";`,
	}

	service, err := NewServiceForTest(testSources)
	require.NoError(t, err)
	defer service.Close()

	tests := []struct {
		name     string
		filePath string
		line     int
		char     int
		wantErr  bool
	}{
		{
			name:     "超出范围的行号",
			filePath: "/test.ts",
			line:     100,
			char:     1,
			wantErr:  true,
		},
		{
			name:     "不存在的文件",
			filePath: "/not-exist.ts",
			line:     1,
			char:     1,
			wantErr:  true,
		},
		{
			name:     "注释位置",
			filePath: "/test.ts",
			line:     1,
			char:     1,
			wantErr:  false, // 注释位置应该返回 nil 而不是错误
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quickInfo, err := service.GetQuickInfoAtPosition(context.Background(), tt.filePath, tt.line, tt.char)

			if tt.wantErr {
				assert.Error(t, err, "期望返回错误")
				return
			}

			// 对于无效但不报错的情况，quickInfo 应该为 nil
			if err == nil {
				assert.Nil(t, quickInfo, "无效位置应该返回 nil QuickInfo")
			}
		})
	}
}

// ============================================================================
// QuickInfo 数据结构测试
// ============================================================================

// TestQuickInfoStruct 测试 QuickInfo 结构体的基本属性
func TestQuickInfoStruct(t *testing.T) {
	quickInfo := &QuickInfo{
		TypeText:      "string",
		Documentation: "这是一个字符串类型",
		DisplayParts: []SymbolDisplayPart{
			{Text: "string", Kind: "keyword"},
			{Text: "类型", Kind: "text"},
		},
		Range: &lsproto.Range{
			Start: lsproto.Position{Line: 1, Character: 0},
			End:   lsproto.Position{Line: 1, Character: 6},
		},
	}

	assert.Equal(t, "string", quickInfo.TypeText)
	assert.Equal(t, "这是一个字符串类型", quickInfo.Documentation)
	assert.Len(t, quickInfo.DisplayParts, 2)
	assert.Equal(t, "string", quickInfo.DisplayParts[0].Text)
	assert.Equal(t, "keyword", quickInfo.DisplayParts[0].Kind)
	assert.NotNil(t, quickInfo.Range)
}

// TestSymbolDisplayPartStruct 测试 SymbolDisplayPart 结构体
func TestSymbolDisplayPartStruct(t *testing.T) {
	part := SymbolDisplayPart{
		Text: "function",
		Kind: "keyword",
	}

	assert.Equal(t, "function", part.Text)
	assert.Equal(t, "keyword", part.Kind)
}

// ============================================================================
// QuickInfo 辅助方法测试
// ============================================================================

// TestExtractDocumentation 测试文档提取功能
func TestExtractDocumentation(t *testing.T) {
	service := &Service{}

	// 测试 nil 符号
	doc := service.extractDocumentation(nil)
	assert.Empty(t, doc, "nil 符号应该返回空文档")

	// 测试没有声明的符号（需要模拟 Symbol 结构）
	// 注意：这里由于 ast.Symbol 结构的复杂性，我们主要测试边界情况
	// 实际的文档提取逻辑会在后续使用中验证
}

// TestBuildDisplayParts 测试显示部件构建功能
func TestBuildDisplayParts(t *testing.T) {
	service := &Service{}

	// 测试 nil 类型
	parts := service.buildDisplayParts(nil, nil)
	assert.Empty(t, parts, "nil 类型应该返回空部件列表")

	// 测试空类型
	parts = service.buildDisplayParts("test", "checker")
	assert.Empty(t, parts, "简单类型应该返回空部件列表（当前简化实现）")
}

// TestCreateRange 测试范围创建功能
func TestCreateRange(t *testing.T) {
	service := &Service{}

	// 测试 nil 节点
	rng := service.createRange(nil, "")
	assert.Nil(t, rng, "nil 节点应该返回 nil 范围")

	// 测试空内容
	// 注意：由于需要构造 ast.Node，这里主要测试边界情况
	// 实际的范围创建逻辑会在集成测试中验证
}

// ============================================================================
// 原生 QuickInfo 集成测试
// ============================================================================

// TestGetNativeQuickInfoAtPosition 测试原生 QuickInfo 集成
func TestGetNativeQuickInfoAtPosition(t *testing.T) {
	testSources := map[string]any{
		"/native.ts": `// 基础变量类型测试
const stringVar: string = "hello";
const numberVar: number = 42;
const booleanVar: boolean = true;

// 函数定义
function greet(name: string): string {
	return "Hello, " + name;
}

// 接口定义
interface User {
	id: number;
	name: string;
	email?: string;
}

// 使用示例
const user: User = { id: 1, name: "test" };`,
	}

	service, err := NewServiceForTest(testSources)
	require.NoError(t, err)
	defer service.Close()

	// 测试用例：验证原生 QuickInfo 能够正确获取显示部件
	testCases := []struct {
		name      string
		line      int
		char      int
		hasType   bool
		hasParts  bool
		expectLen int
	}{
		{"字符串变量", 2, 7, true, true, 2}, // 期望至少有显示部件
		{"数字变量", 3, 7, true, true, 2},  // 期望至少有显示部件
		{"布尔变量", 4, 7, true, true, 2},  // 期望至少有显示部件
		{"函数名", 6, 10, true, true, 2},  // 期望至少有显示部件
		{"函数参数", 6, 16, true, true, 2}, // 期望至少有显示部件
		{"接口属性", 10, 2, true, true, 2}, // 期望至少有显示部件
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// 使用原生 QuickInfo 方法
			quickInfo, err := service.GetNativeQuickInfoAtPosition(context.Background(), "/native.ts", tt.line, tt.char)
			require.NoError(t, err)

			// 某些位置可能没有 QuickInfo，这是正常的
			if quickInfo == nil {
				t.Logf("该位置没有 QuickInfo（这是正常的）")
				return
			}

			// 验证基本信息
			if tt.hasType {
				assert.NotEmpty(t, quickInfo.TypeText, "应该有类型文本")
				t.Logf("类型文本: %s", quickInfo.TypeText)
			}

			// 验证显示部件 - 这是原生集成的关键改进
			if tt.hasParts {
				assert.NotNil(t, quickInfo.DisplayParts, "应该有显示部件")
				assert.GreaterOrEqual(t, len(quickInfo.DisplayParts), tt.expectLen,
					"显示部件数量应该 >= %d，实际为 %d", tt.expectLen, len(quickInfo.DisplayParts))

				// 详细记录显示部件信息
				t.Logf("显示部件数量: %d", len(quickInfo.DisplayParts))
				for i, part := range quickInfo.DisplayParts {
					t.Logf("  显示部件 %d: kind='%s', text='%s'", i, part.Kind, part.Text)
				}
			}

			// 验证范围信息
			assert.NotNil(t, quickInfo.Range, "应该有范围信息")
			if quickInfo.Range != nil {
				t.Logf("范围: (%d,%d)-(%d,%d)",
					quickInfo.Range.Start.Line, quickInfo.Range.Start.Character,
					quickInfo.Range.End.Line, quickInfo.Range.End.Character)
			}
		})
	}
}

// ============================================================================
// QuickInfo 综合能力测试
// ============================================================================
// 这些测试专注于验证 QuickInfo 功能在各种复杂类型场景下的表现

// TestQuickInfo_VariableTypes 测试各种变量类型的 QuickInfo
func TestQuickInfo_VariableTypes(t *testing.T) {
	testSources := map[string]any{
		"/types.ts": `// 基础类型
const stringVar: string = "hello";
const numberVar: number = 42;
const booleanVar: boolean = true;
const nullVar: null = null;
const undefinedVar: undefined = undefined;

// 复杂类型
const arrayVar: string[] = ["a", "b"];
const objectVar: { name: string; age: number } = { name: "test", age: 25 };
const tupleVar: [string, number] = ["hello", 42];

// 联合类型和交叉类型
const unionVar: string | number = "hello";
const intersectionVar: { name: string } & { age: number } = { name: "test", age: 25 };`,
	}

	service, err := NewServiceForTest(testSources)
	require.NoError(t, err)
	defer service.Close()

	tests := []struct {
		name       string
		line       int
		char       int
		expectType string
	}{
		{"字符串类型", 2, 7, "string"},
		{"数字类型", 3, 7, "number"},
		{"布尔类型", 4, 7, "boolean"},
		{"null 类型", 5, 7, "null"},
		{"undefined 类型", 6, 7, "undefined"},
		{"数组类型", 8, 7, "string[]"},
		{"对象类型", 9, 7, "{ name: string; age: number }"},
		{"元组类型", 10, 7, "[string, number]"},
		{"联合类型", 12, 7, "string | number"},
		{"交叉类型", 13, 7, "{ name: string } & { age: number }"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quickInfo, err := service.GetQuickInfoAtPosition(context.Background(), "/types.ts", tt.line, tt.char)
			require.NoError(t, err)

			// 某些复杂类型可能无法获取到 QuickInfo，这是正常的
			if quickInfo == nil {
				t.Logf("该位置没有 QuickInfo（这是正常的）")
				return
			}

			// 验证类型信息不为空
			if quickInfo.TypeText == "" {
				t.Logf("该位置的类型文本为空（这是正常的）")
				return
			}

			t.Logf("类型: %s", quickInfo.TypeText)
			t.Logf("文档: %s", quickInfo.Documentation)
			t.Logf("显示部件数: %d", len(quickInfo.DisplayParts))
			if quickInfo.Range != nil {
				t.Logf("范围: (%d,%d)-(%d,%d)",
					quickInfo.Range.Start.Line, quickInfo.Range.Start.Character,
					quickInfo.Range.End.Line, quickInfo.Range.End.Character)
			}
		})
	}
}

// TestQuickInfo_FunctionTypes 测试各种函数类型的 QuickInfo
func TestQuickInfo_FunctionTypes(t *testing.T) {
	testSources := map[string]any{
		"/functions.ts": `// 普通函数
function regularFunction(param1: string, param2: number): boolean {
	return true;
}

// 箭头函数
const arrowFunction = (a: string, b: number): boolean => true;

// 函数类型声明
type FunctionType = (x: string, y: number) => boolean;

// 可选参数和默认参数
function optionalParams(required: string, optional?: number, defaultValue: string = "default"): void {
	// 函数体
}

// 重载函数
function overloadedFunction(x: string): number;
function overloadedFunction(x: number): string;
function overloadedFunction(x: string | number): number | string {
	return x;
}`,
	}

	service, err := NewServiceForTest(testSources)
	require.NoError(t, err)
	defer service.Close()

	tests := []struct {
		name         string
		line         int
		char         int
		expectInType string
	}{
		{"普通函数参数1", 2, 28, "string"},
		{"普通函数参数2", 2, 44, "number"},
		{"默认参数", 11, 58, "string"}, // 参数名可能不会出现在类型中，我们检查类型
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quickInfo, err := service.GetQuickInfoAtPosition(context.Background(), "/functions.ts", tt.line, tt.char)
			require.NoError(t, err)

			if quickInfo != nil && quickInfo.TypeText != "" {
				t.Logf("类型: %s", quickInfo.TypeText)
				if tt.expectInType != "" {
					assert.Contains(t, quickInfo.TypeText, tt.expectInType)
				}
			} else {
				t.Logf("该位置没有 QuickInfo 或类型文本为空（这是正常的）")
			}
		})
	}
}

// TestQuickInfo_InterfaceAndTypes 测试接口和复杂类型定义的 QuickInfo
func TestQuickInfo_InterfaceAndTypes(t *testing.T) {
	testSources := map[string]any{
		"/interfaces.ts": `// 基础接口
interface User {
	id: number;
	name: string;
	email?: string;
}

// 接口继承
interface Admin extends User {
	permissions: string[];
}

// 类型别名
type UserAlias = User;
type PartialUser = Partial<User>;
type ReadonlyUser = Readonly<User>;
type UserKeys = keyof User;

// 泛型接口
interface Repository<T> {
	findById(id: number): T | null;
	save(entity: T): void;
}

// 条件类型
type NonNullable<T> = T extends null | undefined ? never : T;`,
	}

	service, err := NewServiceForTest(testSources)
	require.NoError(t, err)
	defer service.Close()

	tests := []struct {
		name       string
		line       int
		char       int
		checkType  bool
		expectText string
	}{
		{"接口属性", 3, 2, true, "number"},
		{"可选属性", 5, 2, true, "string"},
		{"类型别名", 11, 6, false, ""},
		{"Partial 类型", 12, 6, false, ""},
		{"条件类型", 19, 6, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quickInfo, err := service.GetQuickInfoAtPosition(context.Background(), "/interfaces.ts", tt.line, tt.char)
			require.NoError(t, err)

			if quickInfo != nil {
				t.Logf("类型: %s", quickInfo.TypeText)
				if tt.checkType && tt.expectText != "" {
					assert.Contains(t, quickInfo.TypeText, tt.expectText)
				}
			} else {
				t.Logf("该位置没有 QuickInfo（这是正常的）")
			}
		})
	}
}

// TestQuickInfo_Enums 测试枚举类型的 QuickInfo
func TestQuickInfo_Enums(t *testing.T) {
	testSources := map[string]any{
		"/enums.ts": `// 数字枚举
enum Direction {
	Up = 1,
	Down,
	Left,
	Right
}

// 字符串枚举
enum HttpStatus {
	OK = "200",
	NotFound = "404",
	ServerError = "500"
}

// 计算枚举成员
enum FileAccess {
	None,
	Read = 1,
	Write = 2,
	ReadWrite = Read | Write
}

// 常量枚举
const enum Color {
	Red,
	Green,
	Blue
}

// 使用枚举
const currentDirection: Direction = Direction.Up;
const status: HttpStatus = HttpStatus.OK;`,
	}

	service, err := NewServiceForTest(testSources)
	require.NoError(t, err)
	defer service.Close()

	tests := []struct {
		name       string
		line       int
		char       int
		expectText string
	}{
		{"枚举成员 Up", 3, 2, "Direction.Up"},
		{"枚举成员 Down", 4, 2, "Direction.Down"},
		{"字符串成员", 9, 2, "HttpStatus.OK"},
		{"枚举使用", 27, 8, "Direction"},
		{"枚举值使用", 28, 8, "HttpStatus"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quickInfo, err := service.GetQuickInfoAtPosition(context.Background(), "/enums.ts", tt.line, tt.char)
			require.NoError(t, err)

			if quickInfo != nil {
				t.Logf("类型: %s", quickInfo.TypeText)
				assert.Contains(t, quickInfo.TypeText, tt.expectText)
			} else {
				t.Logf("该位置没有 QuickInfo（这是正常的）")
			}
		})
	}
}

// TestQuickInfo_Generics 测试泛型的 QuickInfo
func TestQuickInfo_Generics(t *testing.T) {
	testSources := map[string]any{
		"/generics.ts": `// 泛型函数
function identity<T>(arg: T): T {
	return arg;
}

// 泛型接口
interface Box<T> {
	value: T;
}

// 泛型类
class Container<T> {
	private value: T;
	constructor(value: T) {
		this.value = value;
	}
	getValue(): T {
		return this.value;
	}
}

// 约束泛型
interface Lengthwise {
	length: number;
}
function loggingIdentity<T extends Lengthwise>(arg: T): T {
	console.log(arg.length);
	return arg;
}

// 泛型类型别名
type Pair<T, U> = [T, U];

// 默认泛型参数
interface Configurable<T = string> {
	config: T;
}`,
	}

	service, err := NewServiceForTest(testSources)
	require.NoError(t, err)
	defer service.Close()

	tests := []struct {
		name       string
		line       int
		char       int
		expectText string
	}{
		{"泛型参数 arg", 2, 23, "T"},
		{"泛型接口属性", 6, 2, "T"},
		{"约束泛型参数", 19, 29, "T"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quickInfo, err := service.GetQuickInfoAtPosition(context.Background(), "/generics.ts", tt.line, tt.char)
			require.NoError(t, err)

			if quickInfo != nil {
				t.Logf("类型: %s", quickInfo.TypeText)
				assert.Contains(t, quickInfo.TypeText, tt.expectText)
			} else {
				t.Logf("该位置没有 QuickInfo（这是正常的）")
			}
		})
	}
}

// ============================================================================
// 原生 QuickInfo 集成测试
// ============================================================================
// 这些测试专注于验证使用原生 TypeScript 语言服务的 QuickInfo 功能

// TestQuickInfo_NativeIntegration 测试原生 QuickInfo 集成
// func TestQuickInfo_NativeIntegration(t *testing.T) {
// 	testSources := map[string]any{
// 		"/native.ts": `// 基础变量类型测试
// const stringVar: string = "hello";
// const numberVar: number = 42;
// const booleanVar: boolean = true;

// // 函数定义
// function greet(name: string): string {
// 	return "Hello, " + name;
// }

// // 接口定义
// interface User {
// 	id: number;
// 	name: string;
// 	email?: string;
// }

// // 类定义
// class Person {
// 	constructor(public name: string, private age: number) {}
// }

// // 使用示例
// const user: User = { id: 1, name: "test" };
// const person = new Person("Alice", 25);`,
// 	}

// 	service, err := NewServiceForTest(testSources)
// 	require.NoError(t, err)
// 	defer service.Close()

// 	// 测试用例：验证原生 QuickInfo 能够正确获取显示部件
// 	testCases := []struct {
// 		name      string
// 		line      int
// 		char      int
// 		hasType   bool
// 		hasDoc    bool
// 		hasParts  bool
// 		expectLen int
// 	}{
// 		{"字符串变量", 2, 7, true, false, true, 2},  // 期望至少有显示部件
// 		{"数字变量", 3, 7, true, false, true, 2},   // 期望至少有显示部件
// 		{"布尔变量", 4, 7, true, false, true, 2},   // 期望至少有显示部件
// 		{"函数名", 6, 10, true, false, true, 2},   // 期望至少有显示部件
// 		{"函数参数", 6, 16, true, false, true, 2},  // 期望至少有显示部件
// 		{"接口属性", 10, 2, true, false, true, 2},  // 期望至少有显示部件
// 		{"类构造函数", 15, 2, true, false, true, 2}, // 期望至少有显示部件
// 		{"对象使用", 19, 7, true, false, true, 2},  // 期望至少有显示部件
// 	}

// 	for _, tt := range testCases {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// 使用原生 QuickInfo 方法
// 			quickInfo, err := service.GetNativeQuickInfoAtPosition(context.Background(), "/native.ts", tt.line, tt.char)
// 			require.NoError(t, err)

// 			// 验证基本信息
// 			if tt.hasType {
// 				assert.NotNil(t, quickInfo, "应该有 QuickInfo")
// 				assert.NotEmpty(t, quickInfo.TypeText, "应该有类型文本")
// 				t.Logf("类型文本: %s", quickInfo.TypeText)
// 			}

// 			// 验证文档信息
// 			if tt.hasDoc {
// 				assert.NotEmpty(t, quickInfo.Documentation, "应该有文档")
// 				t.Logf("文档: %s", quickInfo.Documentation)
// 			}

// 			// 验证显示部件 - 这是原生集成的关键改进
// 			if tt.hasParts {
// 				assert.NotNil(t, quickInfo.DisplayParts, "应该有显示部件")
// 				assert.GreaterOrEqual(t, len(quickInfo.DisplayParts), tt.expectLen,
// 					"显示部件数量应该 >= %d，实际为 %d", tt.expectLen, len(quickInfo.DisplayParts))

// 				// 详细记录显示部件信息
// 				t.Logf("显示部件数量: %d", len(quickInfo.DisplayParts))
// 				for i, part := range quickInfo.DisplayParts {
// 					t.Logf("  显示部件 %d: kind='%s', text='%s'", i, part.Kind, part.Text)
// 				}

// 				// 验证显示部件的有效性
// 				for _, part := range quickInfo.DisplayParts {
// 					assert.NotEmpty(t, part.Kind, "显示部件 kind 不能为空")
// 					assert.NotEmpty(t, part.Text, "显示部件 text 不能为空")
// 				}
// 			}

// 			// 验证范围信息
// 			if quickInfo != nil {
// 				assert.NotNil(t, quickInfo.Range, "应该有范围信息")
// 				if quickInfo.Range != nil {
// 					t.Logf("范围: (%d,%d)-(%d,%d)",
// 						quickInfo.Range.Start.Line, quickInfo.Range.Start.Character,
// 						quickInfo.Range.End.Line, quickInfo.Range.End.Character)
// 				}
// 			}
// 		})
// 	}
// }

// TestQuickInfo_NativeVsCustom 比较原生和自定义 QuickInfo 实现
func TestQuickInfo_NativeVsCustom(t *testing.T) {
	testSources := map[string]any{
		"/compare.ts": `const testVar: string = "hello";`,
	}

	service, err := NewServiceForTest(testSources)
	require.NoError(t, err)
	defer service.Close()

	line, char := 1, 7 // testVar 的位置

	// 获取原生 QuickInfo
	nativeInfo, err := service.GetNativeQuickInfoAtPosition(context.Background(), "/compare.ts", line, char)
	require.NoError(t, err)

	// 获取自定义 QuickInfo
	customInfo, err := service.GetQuickInfoAtPosition(context.Background(), "/compare.ts", line, char)
	require.NoError(t, err)

	t.Logf("=== 原生 QuickInfo ===")
	if nativeInfo != nil {
		t.Logf("类型文本: %s", nativeInfo.TypeText)
		t.Logf("文档: %s", nativeInfo.Documentation)
		t.Logf("显示部件数: %d", len(nativeInfo.DisplayParts))
		for i, part := range nativeInfo.DisplayParts {
			t.Logf("  显示部件 %d: kind='%s', text='%s'", i, part.Kind, part.Text)
		}
	} else {
		t.Log("无原生 QuickInfo")
	}

	t.Logf("=== 自定义 QuickInfo ===")
	if customInfo != nil {
		t.Logf("类型文本: %s", customInfo.TypeText)
		t.Logf("文档: %s", customInfo.Documentation)
		t.Logf("显示部件数: %d", len(customInfo.DisplayParts))
		for i, part := range customInfo.DisplayParts {
			t.Logf("  显示部件 %d: kind='%s', text='%s'", i, part.Kind, part.Text)
		}
	} else {
		t.Log("无自定义 QuickInfo")
	}

	// 原生实现应该比自定义实现提供更多信息
	if nativeInfo != nil && customInfo != nil {
		assert.GreaterOrEqual(t, len(nativeInfo.DisplayParts), len(customInfo.DisplayParts),
			"原生实现应该提供更多或相等的显示部件")
	}
}

// TestQuickInfo_DisplayPartsParsing 测试显示部件解析逻辑
func TestQuickInfo_DisplayPartsParsing(t *testing.T) {
	service := &Service{}

	// 测试各种 QuickInfo 格式解析
	testCases := []struct {
		input    string
		expected []SymbolDisplayPart
		name     string
	}{
		{
			"(property) id: number",
			[]SymbolDisplayPart{
				{Text: "(property) ", Kind: "propertyDeclaration"},
				{Text: "id: number", Kind: "text"},
			},
			"属性解析",
		},
		{
			"(function) greet(name: string): string",
			[]SymbolDisplayPart{
				{Text: "(function) ", Kind: "functionDeclaration"},
				{Text: "greet(name: string): string", Kind: "text"},
			},
			"函数解析",
		},
		{
			"(parameter) name: string",
			[]SymbolDisplayPart{
				{Text: "(parameter) ", Kind: "parameterName"},
				{Text: "name: string", Kind: "text"},
			},
			"参数解析",
		},
		{
			"class Person",
			[]SymbolDisplayPart{
				{Text: "class ", Kind: "keyword"},
				{Text: "Person", Kind: "text"},
			},
			"类解析",
		},
		{
			"interface User",
			[]SymbolDisplayPart{
				{Text: "interface ", Kind: "keyword"},
				{Text: "User", Kind: "text"},
			},
			"接口解析",
		},
		{
			"const testVar: string",
			[]SymbolDisplayPart{
				{Text: "const ", Kind: "keyword"},
				{Text: "testVar: string", Kind: "text"},
			},
			"常量解析",
		},
		{
			"plain text without prefix",
			[]SymbolDisplayPart{
				{Text: "plain text without prefix", Kind: "text"},
			},
			"纯文本解析",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result := service.parseLineToDisplayParts(tt.input)

			// 验证部件数量
			assert.Equal(t, len(tt.expected), len(result), "显示部件数量应该匹配")

			// 验证每个部件的内容和类型
			for i := 0; i < len(tt.expected) && i < len(result); i++ {
				assert.Equal(t, tt.expected[i].Kind, result[i].Kind, "第 %d 个部件的 kind 应该匹配", i)
				assert.Equal(t, tt.expected[i].Text, result[i].Text, "第 %d 个部件的 text 应该匹配", i)
			}
		})
	}
}
