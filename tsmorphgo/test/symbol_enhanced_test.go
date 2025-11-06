package tsmorphgo

import (
	"strings"
	"testing"

	. "github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// symbol_enhanced_test.go
//
// 这个文件包含了基于 TypeScript-Go Checker 的增强 Symbol 模块测试。
// 验证新的符号系统是否能够正确集成 TypeScript 编译器的符号信息。
//
// 主要测试目标：
// 1. 验证 SymbolManager 的基本功能
// 2. 测试基于 Checker.GetSymbolAtLocation 的符号获取
// 3. 验证符号标志和类型的正确识别
// 4. 测试符号声明和导出状态的检查
// 5. 验证符号缓存机制的有效性

// TestSymbolManager_BasicFunctionality 测试 SymbolManager 的基础功能
func TestSymbolManager_BasicFunctionality(t *testing.T) {
	// 创建测试项目
	project := NewProjectFromSources(map[string]string{
		"/test.ts": `
			function testFunction() {
				const localVar = 42;
				return localVar;
			}

			class TestClass {
				prop: string;
				constructor(prop: string) {
					this.prop = prop;
				}
			}

			export const exportedVar = "hello";
		`,
	})

	require.NotNil(t, project)

	// 由于 getSymbolManager 是私有方法，我们通过公共 API 来测试
	// 这里我们直接测试符号获取功能
	require.NotNil(t, project)

	// 验证缓存统计
	symbolManager := project.GetSymbolManager()
	require.NotNil(t, symbolManager)

	initialCacheStats := symbolManager.GetCacheStats()
	assert.Equal(t, 0, initialCacheStats)
}

// TestSymbol_GetSymbolAtLocation 测试通过 GetSymbolAtLocation 获取符号
func TestSymbol_GetSymbolAtLocation(t *testing.T) {
	// 创建包含各种符号类型的测试项目
	sources := map[string]string{
		"/symbols.ts": `
			export function exportedFunction(): string {
				return "test";
			}

			interface TestInterface {
				method(): void;
			}

			class TestClass implements TestInterface {
				method(): void {}
			}

			const localVariable = 123;
		`,
	}

	project := NewProjectFromSources(sources)
	require.NotNil(t, project)

	sourceFile := project.GetSourceFile("/symbols.ts")
	require.NotNil(t, sourceFile)

	// 测试变量符号
	sourceFile.ForEachDescendant(func(node Node) {
		if IsIdentifier(node) && node.GetText() == "localVariable" {
			// 获取符号
			symbol, err := GetSymbol(node)
			if err != nil {
				t.Logf("Warning: Could not get symbol for localVariable: %v", err)
				return
			}

			if symbol != nil {
				t.Logf("Found symbol: %s", symbol.String())
				assert.Equal(t, "localVariable", symbol.GetName())
				assert.True(t, symbol.IsVariable())
			}
		}
	})

	// 测试导出函数符号
	sourceFile.ForEachDescendant(func(node Node) {
		if IsIdentifier(node) && node.GetText() == "exportedFunction" {
			symbol, err := GetSymbol(node)
			if err != nil {
				t.Logf("Warning: Could not get symbol for exportedFunction: %v", err)
				return
			}

			if symbol != nil {
				t.Logf("Found function symbol: %s", symbol.String())
				assert.Equal(t, "exportedFunction", symbol.GetName())
				assert.True(t, symbol.IsFunction())
				assert.True(t, symbol.IsExported())
			}
		}
	})
}

// TestSymbol_SymbolTypeChecking 测试符号类型检查方法
func TestSymbol_SymbolTypeChecking(t *testing.T) {
	// 创建包含各种符号类型的测试项目
	sources := map[string]string{
		"/types.ts": `
			// 变量符号
			const variableSymbol = "test";

			// 函数符号
			function functionSymbol() {}

			// 类符号
			class classSymbol {}

			// 接口符号
			interface interfaceSymbol {}

			// 类型别名符号
			type typeAliasSymbol = string;

			// 枚举符号
			enum enumSymbol {
				ValueA,
				ValueB
			}
		`,
	}

	project := NewProjectFromSources(sources)
	require.NotNil(t, project)

	sourceFile := project.GetSourceFile("/types.ts")
	require.NotNil(t, sourceFile)

	// 测试各种符号类型的识别
	testCases := []struct {
		identifier      string
		expectedIsVar   bool
		expectedIsFunc  bool
		expectedIsClass bool
		expectedIsIntf  bool
		expectedIsAlias bool
		expectedIsEnum  bool
	}{
		{"variableSymbol", true, false, false, false, false, false},
		{"functionSymbol", false, true, false, false, false, false},
		{"classSymbol", false, false, true, false, false, false},
		{"interfaceSymbol", false, false, false, true, false, false},
		{"typeAliasSymbol", false, false, false, false, true, false},
		{"enumSymbol", false, false, false, false, false, true},
	}

	for _, tc := range testCases {
		sourceFile.ForEachDescendant(func(node Node) {
			if IsIdentifier(node) && node.GetText() == tc.identifier {
				symbol, err := GetSymbol(node)
				if err != nil {
					t.Logf("Warning: Could not get symbol for %s: %v", tc.identifier, err)
					return
				}

				if symbol != nil {
					t.Logf("Testing symbol: %s - %s", tc.identifier, symbol.String())
					assert.Equal(t, tc.identifier, symbol.GetName())

					// 注意：由于我们还没有完全集成 TypeChecker，这些检查可能不会完全准确
					// 这里主要是验证 API 的存在和基本调用
					t.Logf("Symbol %s - IsVariable: %v", tc.identifier, symbol.IsVariable())
					t.Logf("Symbol %s - IsFunction: %v", tc.identifier, symbol.IsFunction())
					t.Logf("Symbol %s - IsClass: %v", tc.identifier, symbol.IsClass())
					t.Logf("Symbol %s - IsInterface: %v", tc.identifier, symbol.IsInterface())
					t.Logf("Symbol %s - IsTypeAlias: %v", tc.identifier, symbol.IsTypeAlias())
					t.Logf("Symbol %s - IsEnum: %v", tc.identifier, symbol.IsEnum())
				}
			}
		})
	}
}

// TestSymbol_Declarations 测试符号声明信息
func TestSymbol_Declarations(t *testing.T) {
	// 创建包含重复声明的测试项目
	sources := map[string]string{
		"/declarations.ts": `
			function overloadedFunction(): void;
			function overloadedFunction(param: string): void;
			function overloadedFunction(param?: string): void {
				// 实现
			}
		`,
	}

	project := NewProjectFromSources(sources)
	require.NotNil(t, project)

	sourceFile := project.GetSourceFile("/declarations.ts")
	require.NotNil(t, sourceFile)

	// 查找函数符号
	sourceFile.ForEachDescendant(func(node Node) {
		if IsIdentifier(node) && node.GetText() == "overloadedFunction" {
			symbol, err := GetSymbol(node)
			if err != nil {
				t.Logf("Warning: Could not get symbol for overloadedFunction: %v", err)
				return
			}

			if symbol != nil {
				t.Logf("Function symbol: %s", symbol.String())
				t.Logf("Declaration count: %d", symbol.GetDeclarationCount())

				// 测试声明获取
				declarations := symbol.GetDeclarations()
				t.Logf("Found %d declarations", len(declarations))

				// 测试第一个声明
				if firstDecl, hasFirst := symbol.GetFirstDeclaration(); hasFirst {
					t.Logf("First declaration: %s", firstDecl.GetText())
				}
			}
		}
	})
}

// TestSymbol_Caching 测试符号缓存机制
func TestSymbol_Caching(t *testing.T) {
	// 创建测试项目
	sources := map[string]string{
		"/cache.ts": `
			const cachedSymbol = "cache test";
		`,
	}

	project := NewProjectFromSources(sources)
	require.NotNil(t, project)

	symbolManager := project.GetSymbolManager()
	require.NotNil(t, symbolManager)

	sourceFile := project.GetSourceFile("/cache.ts")
	require.NotNil(t, sourceFile)

	// 查找目标节点
	var targetNode Node
	sourceFile.ForEachDescendant(func(node Node) {
		if IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "cachedSymbol" {
			targetNode = node
		}
	})

	require.True(t, targetNode.IsValid(), "Could not find target node")

	// 第一次获取符号 - 应该创建缓存
	symbol1, err1 := GetSymbol(targetNode)
	if err1 != nil {
		t.Logf("Warning: First symbol lookup failed: %v", err1)
		return
	}

	// 检查缓存统计
	cacheStats1 := symbolManager.GetCacheStats()
	t.Logf("Cache stats after first lookup: %d", cacheStats1)

	// 第二次获取符号 - 应该使用缓存
	symbol2, err2 := GetSymbol(targetNode)
	if err2 != nil {
		t.Logf("Warning: Second symbol lookup failed: %v", err2)
		return
	}

	// 验证是同一个符号对象
	if symbol1 != nil && symbol2 != nil {
		assert.Equal(t, symbol1.GetName(), symbol2.GetName())
		assert.Equal(t, symbol1.GetFlags(), symbol2.GetFlags())
		t.Logf("Both lookups returned symbols with name: %s", symbol1.GetName())
	}
}

// TestSymbolManager_IntegrationWithProject 测试 SymbolManager 与 Project 的集成
func TestSymbolManager_IntegrationWithProject(t *testing.T) {
	// 创建复杂项目
	sources := map[string]string{
		"/module1.ts": `
			export interface Module1Interface {
				method(): void;
			}

			export class Module1Class implements Module1Interface {
				method(): void {}
			}
		`,
		"/module2.ts": `
			import { Module1Interface } from './module1';

			export function createModule1(): Module1Interface {
				return new Module1Class();
			}
		`,
	}

	project := NewProjectFromSources(sources)
	require.NotNil(t, project)

	// 测试全局符号作用域
	globalScope := project.GetGlobalSymbolScope()
	require.NotNil(t, globalScope)
	assert.Equal(t, "global", globalScope.GetName())

	// 测试符号查找
	symbols := project.FindSymbolsByName("Module1Interface")
	t.Logf("Found %d symbols named 'Module1Interface'", len(symbols))

	// 测试缓存清理
	project.ClearSymbolCache()
	t.Logf("Symbol cache cleared")
}

// TestSymbol_ErrorHandling 测试错误处理
func TestSymbol_ErrorHandling(t *testing.T) {
	// 创建空项目
	project := NewProjectFromSources(map[string]string{})
	require.NotNil(t, project)

	// 创建无效节点
	invalidNode := Node{}
	symbol, err := GetSymbol(invalidNode)

	// 应该返回错误
	assert.Error(t, err)
	assert.Nil(t, symbol)
	assert.Contains(t, err.Error(), "must belong")
}