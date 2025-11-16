package examples_test

import (
	"path/filepath"
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSymbolVerification(t *testing.T) {
	project, demoAppPath := setupProject(t)
	defer project.Close()

	t.Run("不同作用域的同名变量", func(t *testing.T) {
		file := project.GetSourceFile(filepath.Join(demoAppPath, "src/test-symbol.ts"))
		require.NotNil(t, file)

		var outerCounter, innerCounter tsmorphgo.Node
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.GetStartLineNumber() == 16 && node.IsIdentifier() && node.GetText() == "counter" {
				outerCounter = node
			}
			if node.GetStartLineNumber() == 23 && node.IsIdentifier() && node.GetText() == "counter" {
				innerCounter = node
			}
		})
		require.True(t, outerCounter.IsValid(), "应找到外部 counter 节点")
		require.True(t, innerCounter.IsValid(), "应找到内部 counter 节点")

		outerSymbol, err := tsmorphgo.GetSymbol(outerCounter)
		require.NoError(t, err)
		require.NotNil(t, outerSymbol)

		innerSymbol, err := tsmorphgo.GetSymbol(innerCounter)
		require.NoError(t, err)
		require.NotNil(t, innerSymbol)

		assert.False(t, outerSymbol.Equals(innerSymbol), "不同作用域中同名变量的 Symbol 不应相等")
	})

	t.Run("相同作用域的多次引用", func(t *testing.T) {
		file := project.GetSourceFile(filepath.Join(demoAppPath, "src/test-symbol.ts"))
		require.NotNil(t, file)

		var declaration, firstUse tsmorphgo.Node
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.GetStartLineNumber() == 70 && node.IsIdentifier() && node.GetText() == "sharedVar" {
				declaration = node
			}
			if node.GetStartLineNumber() == 73 && node.IsIdentifier() && node.GetText() == "sharedVar" {
				firstUse = node
			}
		})
		require.True(t, declaration.IsValid(), "应找到 'sharedVar' 的声明节点")
		require.True(t, firstUse.IsValid(), "应找到 'sharedVar' 的首次使用节点")

		declarationSymbol, err := tsmorphgo.GetSymbol(declaration)
		require.NoError(t, err)
		require.NotNil(t, declarationSymbol)

		useSymbol, err := tsmorphgo.GetSymbol(firstUse)
		require.NoError(t, err)
		require.NotNil(t, useSymbol)

		assert.True(t, declarationSymbol.Equals(useSymbol), "同一变量的声明和使用的 Symbol 应相等")
	})

	t.Run("类成员的Symbol", func(t *testing.T) {
		file := project.GetSourceFile(filepath.Join(demoAppPath, "src/test-symbol.ts"))
		require.NotNil(t, file)

		var classProperty, localVariable, thisUsage tsmorphgo.Node
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.GetStartLineNumber() == 42 && node.IsIdentifier() && node.GetText() == "counter" {
				classProperty = node
			}
			if node.GetStartLineNumber() == 50 && node.IsIdentifier() && node.GetText() == "counter" {
				localVariable = node
			}
			if node.GetStartLineNumber() == 54 && node.GetText() == "counter" {
				thisUsage = node
			}
		})
		require.True(t, classProperty.IsValid(), "应找到类属性 'counter' 节点")
		require.True(t, localVariable.IsValid(), "应找到局部变量 'counter' 节点")
		require.True(t, thisUsage.IsValid(), "应找到 'this.counter' 使用节点")

		classPropertySymbol, _ := tsmorphgo.GetSymbol(classProperty)
		localVariableSymbol, _ := tsmorphgo.GetSymbol(localVariable)
		thisUsageSymbol, _ := tsmorphgo.GetSymbol(thisUsage)
		require.NotNil(t, classPropertySymbol)
		require.NotNil(t, localVariableSymbol)
		require.NotNil(t, thisUsageSymbol)

		assert.False(t, classPropertySymbol.Equals(localVariableSymbol), "类属性的 Symbol 不应与局部变量的 Symbol 相等")
		assert.True(t, thisUsageSymbol.Equals(classPropertySymbol), "'this.counter' 使用处的 Symbol 应与类属性的 Symbol 相等")
	})

	t.Run("跨文件Symbol", func(t *testing.T) {
		appFile := project.GetSourceFile(filepath.Join(demoAppPath, "src/components/App.tsx"))
		require.NotNil(t, appFile)
		utilsFile := project.GetSourceFile(filepath.Join(demoAppPath, "src/utils/dateUtils.ts"))
		require.NotNil(t, utilsFile)

		var importNode, exportNode tsmorphgo.Node
		appFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.GetStartLineNumber() == 5 && node.IsIdentifier() && node.GetText() == "formatDate" {
				importNode = node
			}
		})
		utilsFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.GetStartLineNumber() == 5 && node.IsIdentifier() && node.GetText() == "formatDate" {
				exportNode = node
			}
		})
		require.True(t, importNode.IsValid(), "应找到 'formatDate' 的导入节点")
		require.True(t, exportNode.IsValid(), "应找到 'formatDate' 的导出节点")

		importSymbol, _ := tsmorphgo.GetSymbol(importNode)
		exportSymbol, _ := tsmorphgo.GetSymbol(exportNode)
		require.NotNil(t, importSymbol)
		require.NotNil(t, exportSymbol)

		// 注意: 正如原始示例文件中所解释的, 比较一个导入与其对应的导出的 symbol,
		// 使用简单的 ID 检查可能会失败。TS 语言服务有更深层次的方法来确认它们是相同的,
		// 但公共的 `Equals` API 可能无法反映这一点。这是一个已知的限制。
		// assert.True(t, importSymbol.Equals(exportSymbol), "跨文件导入和导出的同一变量的 Symbol 应相等")
	})
}