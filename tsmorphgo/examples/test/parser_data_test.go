package examples_test

import (
	"path/filepath"
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserData(t *testing.T) {
	project, demoAppPath := setupProject(t)
	defer project.Close()

	helpersFile := project.GetSourceFile(filepath.Join(demoAppPath, "src/utils/helpers.ts"))
	require.NotNil(t, helpersFile, "应找到 helpers.ts 文件")

	// 查找目标函数声明节点
	var debounceNode tsmorphgo.Node
	helpersFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsFunctionDeclaration() {
			node.ForEachChild(func(child tsmorphgo.Node) bool {
				// 使用 IsValid 确保我们只赋值一次 (找到第一个就停止)
				if !debounceNode.IsValid() && child.IsKind(tsmorphgo.KindIdentifier) && child.GetText() == "debounce" {
					debounceNode = node
					return true // 停止内层循环
				}
				return false
			})
		}
	})
	require.NotNil(t, debounceNode, "应找到 'debounce' 函数声明节点")

	t.Run("解析器数据验证", func(t *testing.T) {
		// 检查是否存在解析器数据
		assert.True(t, debounceNode.HasParserData(), "节点应包含解析器数据")

		// 检查解析器数据的类型
		dataType := debounceNode.GetParserDataType()
		// 注意: 此方法返回底层解析器数据结构的 Go 类型名称。
		assert.Equal(t, "parser.FunctionDeclarationResult", dataType, "解析器数据类型应为 'parser.FunctionDeclarationResult'")

		// 获取实际的解析器数据
		parserData, ok := debounceNode.GetParserData()
		assert.True(t, ok, "应成功获取解析器数据")
		assert.NotNil(t, parserData, "解析器数据不应为 nil")
	})
}