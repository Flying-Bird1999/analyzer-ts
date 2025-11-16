package examples_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComprehensiveVerification(t *testing.T) {
	project, demoAppPath := setupProject(t)
	defer project.Close()

	appFile := project.GetSourceFile(filepath.Join(demoAppPath, "src/components/App.tsx"))
	require.NotNil(t, appFile, "应找到 App.tsx 文件")

	// 查找目标导入声明节点
	var importNode tsmorphgo.Node
	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsImportDeclaration() && strings.Contains(node.GetText(), "Header") && strings.Contains(node.GetText(), "@/components/Header") {
			importNode = node
			return
		}
	})
	require.NotNil(t, importNode, "应找到 'Header' 导入声明")

	t.Run("节点信息", func(t *testing.T) {
		assert.Equal(t, tsmorphgo.KindImportDeclaration, importNode.GetKind())
		assert.Contains(t, importNode.GetText(), "@/components/Header")
		assert.Equal(t, 1, importNode.GetStartLineNumber())
	})

	t.Run("类型检查与收窄", func(t *testing.T) {
		assert.True(t, importNode.IsImportDeclaration())
		assert.True(t, importNode.IsKind(tsmorphgo.KindImportDeclaration))
		assert.False(t, importNode.IsKind(tsmorphgo.KindExportDeclaration))

		_, success := importNode.AsImportDeclaration()
		assert.True(t, success, "应成功将节点收窄为 ImportDeclaration 类型")
	})

	t.Run("导入说明符分析", func(t *testing.T) {
		var importSpecifier tsmorphgo.Node
		importNode.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsKind(tsmorphgo.KindImportSpecifier) && node.GetText() == "Header" {
				importSpecifier = node
				return
			}
		})
		require.NotNil(t, importSpecifier, "应找到 'Header' 导入说明符")
		assert.Equal(t, "Header", importSpecifier.GetText())
	})

	t.Run("符号信息", func(t *testing.T) {
		var headerIdentifier tsmorphgo.Node
		importNode.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsIdentifier() && node.GetText() == "Header" {
				headerIdentifier = node
			}
		})
		require.NotNil(t, headerIdentifier, "应在导入中找到 'Header' 标识符")

		symbol, err := headerIdentifier.GetSymbol()
		require.NoError(t, err)
		require.NotNil(t, symbol)
		assert.Equal(t, "Header", symbol.GetName())
	})

	t.Run("解析器数据", func(t *testing.T) {
		assert.True(t, importNode.HasParserData(), "节点应包含解析器数据")
		assert.NotEmpty(t, importNode.GetParserDataType(), "解析器数据类型不应为空")

		parserData, ok := importNode.GetParserData()
		assert.True(t, ok, "应成功获取解析器数据")
		assert.NotNil(t, parserData, "解析器数据不应为 nil")
	})

	t.Run("位置信息", func(t *testing.T) {
		assert.Equal(t, 1, importNode.GetStartLineNumber())
		assert.Equal(t, 2, importNode.GetEndLineNumber())
		assert.True(t, importNode.GetStart() > 0, "起始位置应大于 0")
		assert.True(t, importNode.GetEnd() > importNode.GetStart(), "结束位置应大于起始位置")
	})
}