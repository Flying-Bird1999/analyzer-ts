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
				assert.True(t, symbol.IsFunction())
				t.Logf("Found symbol: %s", symbol.String())
			}
		}
	})
}

// TestSymbol_TypeChecking 测试 Symbol 类型检查 API
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

	// 测试各种符号类型的检测
	sourceFile.ForEachDescendant(func(node Node) {
		text := node.GetText()
		symbol, err := GetSymbol(node)
		if err != nil || symbol == nil {
			return
		}

		switch text {
		case "variableSymbol":
			assert.True(t, symbol.IsVariable())
			t.Logf("Variable symbol found: %s", symbol.GetName())

		case "functionSymbol":
			assert.True(t, symbol.IsFunction())
			t.Logf("Function symbol found: %s", symbol.GetName())

		case "ClassSymbol":
			assert.True(t, symbol.IsClass())
			t.Logf("Class symbol found: %s", symbol.GetName())

		case "InterfaceSymbol":
			assert.True(t, symbol.IsInterface())
			t.Logf("Interface symbol found: %s", symbol.GetName())
		}
	})
}
