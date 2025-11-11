package tsmorphgo_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	. "github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

// TestSymbol_BasicAPIs 测试 Symbol 基础 API
func TestSymbol_BasicAPIs(t *testing.T) {
	project := NewProjectFromSources(map[string]string{
		"/symbols.ts": `
			export function exportedFunction(): string {
				return "test";
			}
		`,
	})
	defer project.Close()

	sourceFile := project.GetSourceFile("/symbols.ts")
	require.NotNil(t, sourceFile)

	// 查找 exportedFunction 符号
	sourceFile.ForEachDescendant(func(node Node) {
		if node.IsIdentifier() && node.GetText() == "exportedFunction" {
			symbol, err := GetSymbol(node)
			if err != nil {
				t.Logf("Warning: Could not get symbol: %v", err)
				return
			}

			if symbol != nil {
				assert.Equal(t, "exportedFunction", symbol.GetName())
				t.Logf("Found symbol: %s", symbol.String())
			}
		}
	})
}

// TestSymbol_TypeChecking 测试 Symbol 基础 API
func TestSymbol_TypeChecking(t *testing.T) {
	project := NewProjectFromSources(map[string]string{
		"/types.ts": `
			const variableSymbol = "test";
			function functionSymbol(): void {}
			class ClassSymbol {}
			interface InterfaceSymbol {}
		`,
	})
	defer project.Close()

	sourceFile := project.GetSourceFile("/types.ts")
	require.NotNil(t, sourceFile)

	// 测试基础符号功能
	sourceFile.ForEachDescendant(func(node Node) {
		text := node.GetText()
		symbol, err := GetSymbol(node)
		if err != nil || symbol == nil {
			return
		}

		// 验证符号名称正确性
		assert.Equal(t, text, symbol.GetName())
		t.Logf("Symbol found: %s", symbol.String())
	})
}

// TestSymbol_ComprehensiveTypes 全面测试各种 TypeScript 节点类型的 symbol
func TestSymbol_ComprehensiveTypes(t *testing.T) {
	project := NewProjectFromSources(map[string]string{
		"/comprehensive.ts": `
			// 变量声明
			const constVariable = "const";
			let letVariable = "let";
			var varVariable = "var";

			// 函数声明
			function functionDeclaration() {}
			async function asyncFunction() {}
			function* generatorFunction() {}

			// 箭头函数
			const arrowFunction = () => {};
			const asyncArrowFunction = async () => {};

			// 类声明
			class RegularClass {
				constructor() {}
				method() {}
				get getter() { return ""; }
				set setter(value) {}
				static staticMethod() {}
			}

			// 抽象类
			abstract class AbstractClass {
				abstract abstractMethod(): void;
			}

			// 接口声明
		interface SimpleInterface {
				method(): void;
			}

			interface GenericInterface<T> {
				value: T;
				method(value: T): T;
			}

			// 类型别名
			type TypeAlias = string;
			type GenericType<T> = T[];
			type UnionType = string | number;
			type IntersectionType = { a: string } & { b: number };

			// 枚举
			enum StringEnum {
				A = "a",
				B = "b"
			}

			enum NumericEnum {
				A = 0,
				B = 1
			}

			// 命名空间
			namespace MyNamespace {
				export const exportedVar = "namespace";
			}

			// 导入导出
			import { ImportInterface } from "./types";
			import DefaultImport from "./types";
			import * as NamespaceImport from "./types";

			export const exportedVar = "exported";
			export function exportedFunction() {}
			export class ExportedClass {}
			export interface ExportedInterface {}
		`,
		"/types.ts": `
			export interface ImportInterface {
				property: string;
			}

			export default class DefaultExport {
				method() {}
			}
		`,
	})
	defer project.Close()

	sourceFile := project.GetSourceFile("/comprehensive.ts")
	require.NotNil(t, sourceFile)

	t.Log("=== 全面 Symbol 测试开始 ===")

	// 统计找到的符号数量
	symbolCount := 0
	symbolDetails := make(map[string][]string)

	sourceFile.ForEachDescendant(func(node Node) {
		text := strings.TrimSpace(node.GetText())
		if len(text) > 50 { // 忽略过长的文本
			return
		}

		// 获取符号信息
		symbol, err := GetSymbol(node)
		if err != nil {
			t.Logf("❌ 获取符号失败: %s, 错误: %v", text, err)
			return
		}

		if symbol == nil {
			return // 忽略没有符号的节点
		}

		symbolCount++
		symbolName := symbol.GetName()

		// 记录符号详细信息
		detail := fmt.Sprintf("  📍 位置: 行%d,列%d | 节点类型: %v | 文本: '%s'",
			node.GetStartLineNumber(),
			node.GetStartColumnNumber(),
			node.Kind,
			text)
		symbolDetails[symbolName] = append(symbolDetails[symbolName], detail)

		// 输出符号详情
		t.Logf("✅ Symbol[%d] - 名称: '%s' | %s", symbolCount, symbolName, symbol.String())

		// 尝试使用增强版的 GetSymbolAtLocation
		if enhancedSymbol, err := GetSymbolAtLocation(node); err == nil && enhancedSymbol != nil {
			t.Logf("🔍 Enhanced Symbol: %s", enhancedSymbol.String())
		} else {
			t.Logf("⚠️ Enhanced Symbol 未找到或失败: %v", err)
		}
	})

	t.Log("\n=== Symbol 汇总 ===")
	t.Logf("总共找到 %d 个符号", symbolCount)

	for name, details := range symbolDetails {
		t.Logf("\n🔖 Symbol: '%s'", name)
		for _, detail := range details {
			t.Log(detail)
		}
	}
}

// TestSymbol_ImportsExports 测试导入导出场景下的 symbol 行为
func TestSymbol_ImportsExports(t *testing.T) {
	project := NewProjectFromSources(map[string]string{
		"/module.ts": `
			// 导出不同的类型
			export const exportedConst = "const";
			export let exportedLet = "let";
			export var exportedVar = "var";
			export function exportedFunction() {}
			export class ExportedClass {}
			export interface ExportedInterface {}
			export type ExportedType = string;
			export enum ExportedEnum { A = "a", B = "b" }

			// 默认导出
			export default class DefaultExport {}
		`,
		"/consumer.ts": `
			// 导入不同的类型
			import { exportedConst, exportedFunction, ExportedClass, ExportedInterface } from "./module";
			import DefaultExport from "./module";
			import * as Namespace from "./module";

			// 使用导入的符号
			const localVar = exportedConst;
			const localFunc = exportedFunction;
			const localClass = new ExportedClass();
			const localDefault = new DefaultExport();

			// 本地符号
			function localFunction() {}
			class LocalClass {}
			const localConst = "local";
		`,
	})
	defer project.Close()

	t.Log("=== 导入导出 Symbol 测试开始 ===")

	// 测试模块文件中的导出符号
	moduleFile := project.GetSourceFile("/module.ts")
	require.NotNil(t, moduleFile)

	t.Log("\n📤 模块文件中的导出符号:")
	moduleFile.ForEachDescendant(func(node Node) {
		text := strings.TrimSpace(node.GetText())
		if len(text) > 30 {
			return
		}

		symbol, err := GetSymbol(node)
		if err != nil || symbol == nil {
			return
		}

		symbolName := symbol.GetName()
		if text == symbolName {
			t.Logf("  🎯 导出符号: '%s' | 位置: 行%d,列%d | %s",
				symbolName, node.GetStartLineNumber(), node.GetStartColumnNumber(), symbol.String())
		}
	})

	// 测试消费者文件中的导入符号
	consumerFile := project.GetSourceFile("/consumer.ts")
	require.NotNil(t, consumerFile)

	t.Log("\n📥 消费者文件中的符号:")
	consumerFile.ForEachDescendant(func(node Node) {
		text := strings.TrimSpace(node.GetText())
		if len(text) > 30 {
			return
		}

		symbol, err := GetSymbol(node)
		if err != nil || symbol == nil {
			return
		}

		symbolName := symbol.GetName()
		t.Logf("  🔗 符号: '%s' | 文本: '%s' | 位置: 行%d,列%d | %s",
			symbolName, text, node.GetStartLineNumber(), node.GetStartColumnNumber(), symbol.String())

		// 尝试使用增强版方法
		if enhancedSymbol, err := GetSymbolAtLocation(node); err == nil && enhancedSymbol != nil {
			t.Logf("      🔍 Enhanced: %s", enhancedSymbol.String())
		}
	})
}

// TestSymbol_ComplexScenarios 测试复杂场景下的 symbol 行为
func TestSymbol_ComplexScenarios(t *testing.T) {
	project := NewProjectFromSources(map[string]string{
		"/complex.ts": `
			// 嵌套函数和类
			class OuterClass {
				private privateField: string;
				protected protectedField: number;
				public publicField: boolean;

				constructor(private paramField: string) {}

				public method(): void {
					const localInMethod = "local";

					// 嵌套函数
					function nestedFunction() {
						const nestedLocal = "nested";
						return nestedLocal;
					}
				}

				get getter(): string {
					return "";
				}

				set setter(value: string) {}
			}

			// 命名空间嵌套
			namespace OuterNamespace {
				export namespace InnerNamespace {
					export const nestedConst = "nested";

					export class NestedClass {
						method() {}
					}
				}
			}

			// 泛型类
			class GenericClass<T, U extends string> {
				private genericField: T;

				constructor(field: T) {
					this.genericField = field;
				}

				public genericMethod(value: T): U {
					return value as U;
				}
			}

			// 装饰器（如果支持）
			// @decorated
			class DecoratedClass {}

			// 解构赋值
			const { prop1, prop2 } = { prop1: "a", prop2: "b" };
			const [arr1, arr2] = [1, 2];

			// 对象字面量
			const objectLiteral = {
				prop: "value",
				method() { return "method"; }
			};
		`,
	})
	defer project.Close()

	t.Log("=== 复杂场景 Symbol 测试开始 ===")

	sourceFile := project.GetSourceFile("/complex.ts")
	require.NotNil(t, sourceFile)

	symbolCount := 0
	categories := make(map[string][]string)

	sourceFile.ForEachDescendant(func(node Node) {
		text := strings.TrimSpace(node.GetText())
		if len(text) == 0 || len(text) > 40 {
			return
		}

		symbol, err := GetSymbol(node)
		if err != nil || symbol == nil {
			return
		}

		symbolCount++
		symbolName := symbol.GetName()

		// 根据上下文分类
		var category string
		switch {
		case strings.Contains(text, "OuterClass"):
			category = "🏛️ 外层类"
		case strings.Contains(text, "nested") || strings.Contains(text, "Nested"):
			category = "🪆 嵌套元素"
		case strings.Contains(text, "Generic"):
			category = "🧬 泛型"
		case strings.Contains(text, "Namespace"):
			category = "📦 命名空间"
		case strings.Contains(text, "private") || strings.Contains(text, "protected") || strings.Contains(text, "public"):
			category = "🔐 访问修饰符"
		case strings.Contains(text, "prop") || strings.Contains(text, "arr"):
			category = "📋 解构"
		case strings.Contains(text, "objectLiteral"):
			category = "📝 对象字面量"
		default:
			category = "🔧 其他"
		}

		detail := fmt.Sprintf("    '%s' (行%d,列%d) | %s",
			symbolName, node.GetStartLineNumber(), node.GetStartColumnNumber(), symbol.String())
		categories[category] = append(categories[category], detail)
	})

	t.Logf("📊 总共找到 %d 个符号", symbolCount)

	for category, details := range categories {
		t.Logf("\n%s:", category)
		for _, detail := range details {
			t.Log(detail)
		}
	}
}

// TestSymbol_FlagsAnalysis 详细分析不同类型的 Symbol Flags
func TestSymbol_FlagsAnalysis(t *testing.T) {
	project := NewProjectFromSources(map[string]string{
		"/flags.ts": `
			// 变量声明
			const constVar = "const";
			let letVar = "let";
			var varVar = "var";

			// 函数声明
			function functionDeclaration() {}
			async function asyncFunction() {}
			function* generatorFunction() {}

			// 箭头函数
			const arrowFunction = () => {};

			// 类声明
			class RegularClass {
				constructor() {}
				method() {}
				get getter() { return ""; }
				set setter(value) {}
				static staticMethod() {}
			}

			// 接口声明
			interface InterfaceDeclaration {
				method(): void;
			}

			// 类型别名
			type TypeAlias = string;

			// 枚举
			enum EnumDeclaration {
				A = "a",
				B = "b"
			}

			// 模块/命名空间
			namespace NamespaceDeclaration {
				export const namespacedVar = "namespace";
			}

			// 参数
			function functionWithParameters(param1: string, param2: number) {}

			// 属性访问
			const obj = { property: "value" };
			console.log(obj.property);
		`,
	})
	defer project.Close()

	sourceFile := project.GetSourceFile("/flags.ts")
	require.NotNil(t, sourceFile)

	t.Log("=== Symbol Flags 详细分析开始 ===")

	// 按flags值分组分析
	flagsGroups := make(map[uint32][]string)
	symbolCount := 0

	sourceFile.ForEachDescendant(func(node Node) {
		text := strings.TrimSpace(node.GetText())
		if len(text) > 30 || len(text) == 0 {
			return
		}

		symbol, err := GetSymbol(node)
		if err != nil || symbol == nil {
			return
		}

		symbolCount++
		symbolName := symbol.GetName()

		// 解析flags值
		flags := symbol.GetFlags()
		desc := fmt.Sprintf("  '%s' (行%d,列%d) | %s",
			symbolName, node.GetStartLineNumber(), node.GetStartColumnNumber(), text)
		flagsGroups[flags] = append(flagsGroups[flags], desc)
	})

	t.Logf("📊 总共找到 %d 个符号，按 Flags 分组:", symbolCount)

	// 按flags值从大到小排序，便于观察
	var sortedFlags []uint32
	for flags := range flagsGroups {
		sortedFlags = append(sortedFlags, flags)
	}
	for i := 0; i < len(sortedFlags); i++ {
		for j := i + 1; j < len(sortedFlags); j++ {
			if sortedFlags[i] < sortedFlags[j] {
				sortedFlags[i], sortedFlags[j] = sortedFlags[j], sortedFlags[i]
			}
		}
	}

	for _, flags := range sortedFlags {
		details := flagsGroups[flags]
		t.Logf("\n🏷️  Flags = %d (二进制: %032b)", flags, flags)
		t.Logf("   含义推测: %s", interpretFlags(flags))
		for _, detail := range details {
			t.Log(detail)
		}
	}
}

// interpretFlags 解释常见的 flags 值含义
func interpretFlags(flags uint32) string {

	// 常见的标志位组合（基于TypeScript源码）
	switch flags {
	case 0:
		return "未知或无特殊标志"
	case 1:
		return "函数作用域变量 (FunctionScopedVariable)"
	case 2:
		return "块作用域变量 (BlockScopedVariable)"
	case 4:
		return "属性 (Property)"
	case 8:
		return "枚举成员 (EnumMember)"
	case 16:
		return "函数 (Function)"
	case 32:
		return "类 (Class)"
	case 64:
		return "接口 (Interface)"
	case 128:
		return "常量枚举 (ConstEnum)"
	case 256:
		return "常规枚举 (RegularEnum)"
	case 512:
		return "值模块 (ValueModule)"
	case 1024:
		return "命名空间模块 (NamespaceModule)"
	case 2048:
		return "类型别名 (TypeAlias)"
	case 4096:
		return "方法 (Method)"
	case 8192:
		return "构造函数 (Constructor)"
	case 16384:
		return "get 访问器 (GetAccessor)"
	case 32768:
		return "set 访问器 (SetAccessor)"
	case 65536:
		return "签名 (Signature)"
	case 131072:
		return "类型参数 (TypeParameter)"
	case 262144:
		return "类型 (Type)"
	case 524288:
		return "类型字面量 (TypeLiteral)"
	case 1048576:
		return "对象字面量 (ObjectLiteral)"
	case 2097152:
		return "演进方法 (EvictedMethod)"
	case 4194304:
		return "传递泛型 (TransitiveGeneric)"
	case 8388608:
		return "可选类型参数 (OptionalTypeParameter)"
	case 16777216:
		return "类表达式的隐式引用 (ClassThisImplicitThis)"
	case 33554432:
		return "类型谓词 (TypePredicate)"
	case 67108864:
		return "多态类型 (Polymorphic)"
	case 134217728:
		return "导出值 (ExportValue)"
	case 268435456:
		// 组合标志，例如: 262144 (Type) + 134217728 (ExportValue)
		return "导出类型 (Exported Type)"
	default:
		// 分析组合标志
		var components []string
		if flags&1 != 0 {
			components = append(components, "FunctionScoped")
		}
		if flags&2 != 0 {
			components = append(components, "BlockScoped")
		}
		if flags&134217728 != 0 {
			components = append(components, "Exported")
		}
		if flags&262144 != 0 {
			components = append(components, "Type")
		}
		if len(components) > 0 {
			return "组合: " + strings.Join(components, " + ")
		}
		return fmt.Sprintf("未知组合 (0x%x)", flags)
	}
}

// TestSymbol_TsMorphAPIScenarios 测试 ts-morph.md 文档中提到的 getSymbol 使用场景
func TestSymbol_TsMorphAPIScenarios(t *testing.T) {
	project := NewProjectFromSources(map[string]string{
		"/scenarios.ts": `
			// 场景1: 变量声明和使用
			const myVariable = "original";
			const anotherReference = myVariable;

			// 场景2: 函数声明和调用
			function myFunction(param: string): string {
				return param;
			}

			const functionResult = myFunction("test");

			// 场景3: 类声明和实例化
			class MyClass {
				constructor(public value: string) {}
				getValue(): string {
					return this.value;
				}
			}

			const myInstance = new MyClass("instance");

			// 场景4: 对象属性访问
			const myObject = {
				property: "value",
				method(): string {
					return "method result";
				}
			};

			const propertyAccess = myObject.property;
			const methodCall = myObject.method();

			// 场景5: 函数参数和重命名
			function withParameters(param1: number, param2: string) {
				const local1 = param1;
				const local2 = param2;
				return { local1, local2 };
			}

			const result = withParameters(42, "answer");
		`,
	})
	defer project.Close()

	sourceFile := project.GetSourceFile("/scenarios.ts")
	require.NotNil(t, sourceFile)

	t.Log("=== ts-morph API 使用场景测试开始 ===")

	// 场景1: 验证相同变量的符号一致性
	t.Log("\n🔍 场景1: 变量符号一致性检查")
	var myVariableDeclarations []*Node
	var myVariableUsages []*Node

	sourceFile.ForEachDescendant(func(node Node) {
		text := strings.TrimSpace(node.GetText())

		// 找到 myVariable 的声明和使用
		if text == "myVariable" {
			symbol, err := GetSymbol(node)
			if err == nil && symbol != nil {
				if node.GetParent().IsVariableDeclaration() {
					myVariableDeclarations = append(myVariableDeclarations, &node)
					t.Logf("  📍 声明: '%s' (行%d,列%d) | %s",
						symbol.GetName(), node.GetStartLineNumber(), node.GetStartColumnNumber(), symbol.String())
				} else {
					myVariableUsages = append(myVariableUsages, &node)
					t.Logf("  🔗 使用: '%s' (行%d,列%d) | %s",
						symbol.GetName(), node.GetStartLineNumber(), node.GetStartColumnNumber(), symbol.String())
				}
			}
		}
	})

	// 验证所有 myVariable 的引用是否指向同一个符号
	if len(myVariableDeclarations) > 0 && len(myVariableUsages) > 0 {
		declSymbol, _ := GetSymbol(*myVariableDeclarations[0])
		allMatch := true

		for i, usage := range myVariableUsages {
			usageSymbol, err := GetSymbol(*usage)
			if err != nil || usageSymbol == nil {
				t.Logf("  ❌ 使用点 %d 无法获取符号", i)
				allMatch = false
				continue
			}

			// 比较符号名称和flags
			if declSymbol.GetName() == usageSymbol.GetName() &&
			   declSymbol.GetFlags() == usageSymbol.GetFlags() {
				t.Logf("  ✅ 使用点 %d 符号匹配: %s", i, usageSymbol.GetName())
			} else {
				t.Logf("  ❌ 使用点 %d 符号不匹配: 声明=%s vs 使用=%s",
					i, declSymbol.GetName(), usageSymbol.GetName())
				allMatch = false
			}
		}

		if allMatch {
			t.Log("  🎯 所有 myVariable 引用都指向同一个符号！")
		}
	}

	// 场景2: 函数符号测试
	t.Log("\n🔍 场景2: 函数符号分析")
	var functionDeclarations []*Node
	var functionCalls []*Node

	sourceFile.ForEachDescendant(func(node Node) {
		text := strings.TrimSpace(node.GetText())

		if text == "myFunction" {
			symbol, err := GetSymbol(node)
			if err == nil && symbol != nil {
				if node.GetParent().IsFunctionDeclaration() {
					functionDeclarations = append(functionDeclarations, &node)
					t.Logf("  📋 函数声明: '%s' (行%d) | Flags: %d",
						symbol.GetName(), node.GetStartLineNumber(), symbol.GetFlags())
				} else if node.GetParent().IsCallExpression() {
					functionCalls = append(functionCalls, &node)
					t.Logf("  📞 函数调用: '%s' (行%d) | Flags: %d",
						symbol.GetName(), node.GetStartLineNumber(), symbol.GetFlags())
				}
			}
		}
	})

	// 场景3: 类和实例测试
	t.Log("\n🔍 场景3: 类符号和实例化")
	var classDeclarations []*Node
	var classReferences []*Node

	sourceFile.ForEachDescendant(func(node Node) {
		text := strings.TrimSpace(node.GetText())

		if text == "MyClass" {
			symbol, err := GetSymbol(node)
			if err == nil && symbol != nil {
				parent := node.GetParent()
				if parent.IsClassDeclaration() {
					classDeclarations = append(classDeclarations, &node)
					t.Logf("  🏗️ 类声明: '%s' (行%d) | Flags: %d",
						symbol.GetName(), node.GetStartLineNumber(), symbol.GetFlags())
				} else {
					classReferences = append(classReferences, &node)
					t.Logf("  🔗 类引用: '%s' (行%d, 列%d) | 父类型: %v | Flags: %d",
						symbol.GetName(), node.GetStartLineNumber(), node.GetStartColumnNumber(),
						parent.Kind, symbol.GetFlags())
				}
			}
		}
	})

	// 场景4: 对象属性和方法
	t.Log("\n🔍 场景4: 对象属性和方法")
	var propertyDeclarations []*Node
	var propertyAccesses []*Node

	sourceFile.ForEachDescendant(func(node Node) {
		text := strings.TrimSpace(node.GetText())

		if text == "property" || text == "method" {
			symbol, err := GetSymbol(node)
			if err == nil && symbol != nil {
				parent := node.GetParent()
				if parent.Kind == KindPropertyAssignment || parent.Kind == KindMethodDeclaration {
					propertyDeclarations = append(propertyDeclarations, &node)
					t.Logf("  📝 属性声明: '%s' (行%d) | 父类型: %v | Flags: %d",
						symbol.GetName(), node.GetStartLineNumber(), parent.Kind, symbol.GetFlags())
				} else {
					propertyAccesses = append(propertyAccesses, &node)
					t.Logf("  🔍 属性访问: '%s' (行%d) | 父类型: %v | Flags: %d",
						symbol.GetName(), node.GetStartLineNumber(), parent.Kind, symbol.GetFlags())
				}
			}
		}
	})

	// 场景5: 函数参数符号
	t.Log("\n🔍 场景5: 函数参数符号分析")
	var parameterSymbols []*Node

	sourceFile.ForEachDescendant(func(node Node) {
		text := strings.TrimSpace(node.GetText())

		// 查找函数参数
		if text == "param1" || text == "param2" {
			symbol, err := GetSymbol(node)
			if err == nil && symbol != nil {
				parameterSymbols = append(parameterSymbols, &node)
				t.Logf("  🎯 参数: '%s' (行%d,列%d) | Flags: %d | %s",
					symbol.GetName(), node.GetStartLineNumber(), node.GetStartColumnNumber(),
					symbol.GetFlags(), symbol.String())

				// 使用增强版方法获取符号
				if enhancedSymbol, err := GetSymbolAtLocation(node); err == nil && enhancedSymbol != nil {
					t.Logf("      🔍 Enhanced: %s", enhancedSymbol.String())
				}
			}
		}
	})

	// 汇总统计
	t.Logf("\n📊 场景测试汇总:")
	t.Logf("  变量符号: %d 个声明, %d 个使用", len(myVariableDeclarations), len(myVariableUsages))
	t.Logf("  函数符号: %d 个声明, %d 个调用", len(functionDeclarations), len(functionCalls))
	t.Logf("  类符号: %d 个声明, %d 个引用", len(classDeclarations), len(classReferences))
	t.Logf("  属性符号: %d 个声明, %d 个访问", len(propertyDeclarations), len(propertyAccesses))
	t.Logf("  参数符号: %d 个参数", len(parameterSymbols))
}

// TestSymbol_VariableConsistency 验证相同变量在不同场景下的 symbol 一致性
func TestSymbol_VariableConsistency(t *testing.T) {
	project := NewProjectFromSources(map[string]string{
		"/consistency.ts": `
			// 场景1: 变量声明和基本使用
			const globalVariable = "global";
			const usedInExpression = globalVariable + " suffix";

			// 场景2: 函数参数中的使用
			function functionWithParam(param: string = globalVariable) {
				return param;
			}

			// 场景3: 对象属性中使用
			const objectWithRef = {
				property: globalVariable,
				method() {
					return globalVariable;
				}
			};

			// 场景4: 条件语句中使用
			if (globalVariable) {
				console.log(globalVariable);
			}

			// 场景5: 循环中使用
			for (let i = 0; i < 1; i++) {
				const loopRef = globalVariable;
			}

			// 场景6: 函数返回值中使用
			function returnGlobalVar() {
				return globalVariable;
			}

			// 场景7: 类中使用
			class UsingGlobalVar {
				method() {
					return globalVariable;
				}
				constructor() {
					console.log(globalVariable);
				}
			}

			// 场景8: 数组中使用
			const arrayWithRef = [globalVariable, 1, 2];

			// 场景9: 作为函数参数传递
			function processVar(inputVar: string) {
				return inputVar;
			}
			const processed = processVar(globalVariable);

			// 场景10: 字符串拼接中使用
			const templateResult = globalVariable + " in template";

			// 场景11: 重新赋值（测试相同变量名的新声明）
			let letVariable = "original";
			letVariable = "modified";  // 这里应该还是同一个符号
			const constRef = letVariable;
		`,
	})
	defer project.Close()

	sourceFile := project.GetSourceFile("/consistency.ts")
	require.NotNil(t, sourceFile)

	t.Log("=== 变量符号一致性测试开始 ===")

	// 测试 globalVariable 的一致性
	t.Log("\n🔍 测试 'globalVariable' 在所有使用场景中的一致性")
	var globalVariableNodes []*Node
	var globalVariableSymbols []*Symbol

	sourceFile.ForEachDescendant(func(node Node) {
		text := strings.TrimSpace(node.GetText())

		if text == "globalVariable" {
			symbol, err := GetSymbol(node)
			if err == nil && symbol != nil {
				globalVariableNodes = append(globalVariableNodes, &node)
				globalVariableSymbols = append(globalVariableSymbols, symbol)

				parent := node.GetParent()
				parentType := "未知"
				if parent != nil {
					parentType = fmt.Sprintf("%v", parent.Kind)
				}

				t.Logf("  📍 使用点[%d]: 位置(行%d,列%d) | 父类型: %s | 符号: %s",
					len(globalVariableNodes)-1,
					node.GetStartLineNumber(),
					node.GetStartColumnNumber(),
					parentType,
					symbol.String())
			}
		}
	})

	// 验证所有 globalVariable 的符号是否一致
	if len(globalVariableSymbols) > 1 {
		t.Logf("\n🔍 验证 'globalVariable' 符号一致性 (共 %d 个引用):", len(globalVariableSymbols))

		firstSymbol := globalVariableSymbols[0]
		firstName := firstSymbol.GetName()
		firstFlags := firstSymbol.GetFlags()

		allConsistent := true
		var inconsistentPositions []string

		for i, symbol := range globalVariableSymbols {
			currentName := symbol.GetName()
			currentFlags := symbol.GetFlags()

			if currentName == firstName && currentFlags == firstFlags {
				t.Logf("  ✅ 引用点[%d]: 名称='%s', Flags=%d (一致)", i, currentName, currentFlags)
			} else {
				t.Logf("  ❌ 引用点[%d]: 名称='%s', Flags=%d (不一致！期望: 名称='%s', Flags=%d)",
					i, currentName, currentFlags, firstName, firstFlags)
				inconsistentPositions = append(inconsistentPositions, fmt.Sprintf("引用点%d(行%d)", i, globalVariableNodes[i].GetStartLineNumber()))
				allConsistent = false
			}
		}

		if allConsistent {
			t.Log("  🎯 所有 'globalVariable' 引用的符号完全一致！")
		} else {
			t.Logf("  ⚠️ 发现不一致的符号引用: %v", inconsistentPositions)
		}
	} else {
		t.Log("  ❌ 未找到足够的 'globalVariable' 引用进行一致性检查")
	}

	// 测试 letVariable 的一致性
	t.Log("\n🔍 测试 'letVariable' 在重新赋值场景中的符号一致性")
	var letVariableNodes []*Node
	var letVariableSymbols []*Symbol

	sourceFile.ForEachDescendant(func(node Node) {
		text := strings.TrimSpace(node.GetText())

		if text == "letVariable" {
			symbol, err := GetSymbol(node)
			if err == nil && symbol != nil {
				letVariableNodes = append(letVariableNodes, &node)
				letVariableSymbols = append(letVariableSymbols, symbol)

				t.Logf("  📍 letVariable[%d]: 位置(行%d,列%d) | %s",
					len(letVariableNodes)-1,
					node.GetStartLineNumber(),
					node.GetStartColumnNumber(),
					symbol.String())
			}
		}
	})

	if len(letVariableSymbols) > 1 {
		t.Logf("\n🔍 验证 'letVariable' 符号一致性 (共 %d 个引用):", len(letVariableSymbols))

		firstSymbol := letVariableSymbols[0]
		firstName := firstSymbol.GetName()
		firstFlags := firstSymbol.GetFlags()

		allConsistent := true

		for i, symbol := range letVariableSymbols {
			currentName := symbol.GetName()
			currentFlags := symbol.GetFlags()

			if currentName == firstName && currentFlags == firstFlags {
				t.Logf("  ✅ 引用点[%d]: 名称='%s', Flags=%d (一致)", i, currentName, currentFlags)
			} else {
				t.Logf("  ❌ 引用点[%d]: 名称='%s', Flags=%d (不一致)", i, currentName, currentFlags)
				allConsistent = false
			}
		}

		if allConsistent {
			t.Log("  🎯 所有 'letVariable' 引用的符号完全一致（包括重新赋值）！")
		} else {
			t.Log("  ❌ 'letVariable' 符号不一致")
		}
	}

	// 测试 param 参数的符号一致性
	t.Log("\n🔍 测试函数参数 'param' 的符号一致性")
	var paramNodes []*Node
	var paramSymbols []*Symbol

	sourceFile.ForEachDescendant(func(node Node) {
		text := strings.TrimSpace(node.GetText())

		if text == "param" {
			symbol, err := GetSymbol(node)
			if err == nil && symbol != nil {
				paramNodes = append(paramNodes, &node)
				paramSymbols = append(paramSymbols, symbol)

				t.Logf("  📍 param[%d]: 位置(行%d,列%d) | %s",
					len(paramNodes)-1,
					node.GetStartLineNumber(),
					node.GetStartColumnNumber(),
					symbol.String())
			}
		}
	})

	if len(paramSymbols) > 1 {
		t.Logf("\n🔍 验证 'param' 符号一致性 (共 %d 个引用):", len(paramSymbols))

		// 注意：不同的函数中的同名参数应该是不同的符号
		for i, symbol := range paramSymbols {
			t.Logf("  📊 param[%d]: 名称='%s', Flags=%d", i, symbol.GetName(), symbol.GetFlags())
		}

		// 验证相同函数内的参数符号是否一致
		symbolGroups := make(map[string][]*Symbol)
		symbolPositions := make(map[string][]int)

		for i, symbol := range paramSymbols {
			symbolKey := fmt.Sprintf("%s_%d", symbol.GetName(), symbol.GetFlags())
			symbolGroups[symbolKey] = append(symbolGroups[symbolKey], symbol)
			symbolPositions[symbolKey] = append(symbolPositions[symbolKey], i)
		}

		t.Log("  📋 按 Symbol 分组:")
		for key, symbols := range symbolGroups {
			positions := symbolPositions[key]
			t.Logf("    符号组: %s (共 %d 个引用，位置: %v)", key, len(symbols), positions)
		}
	}

	// 最终汇总
	t.Log("\n📊 变量符号一致性测试汇总:")
	t.Logf("  'globalVariable': %d 个引用", len(globalVariableSymbols))
	t.Logf("  'letVariable': %d 个引用", len(letVariableSymbols))
	t.Logf("  'param': %d 个引用", len(paramSymbols))
}
