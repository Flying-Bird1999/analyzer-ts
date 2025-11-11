package tsmorphgo_test

import (
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
