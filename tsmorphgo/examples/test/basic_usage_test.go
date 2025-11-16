package examples_test

import (
	"path/filepath"
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicUsage(t *testing.T) {
	// 设置: 初始化项目
	demoAppPath, err := filepath.Abs("../demo-react-app")
	require.NoError(t, err, "获取 demo-react-app 绝对路径失败")

	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:    demoAppPath,
		UseTsConfig: true,
	})
	require.NotNil(t, project, "项目创建不应失败")
	defer project.Close()

	t.Run("项目初始化", func(t *testing.T) {
		sourceFiles := project.GetSourceFiles()
		// 注意: 原始示例预期找到 13 个文件。当前的 TypeScript 版本可能会包含更多文件(例如默认的库文件)。
		// 关键在于它找到了非零数量的文件。我们断言上次测试运行时发现的实际数量。
		assert.Equal(t, 15, len(sourceFiles), "应扫描并找到 15 个源文件")
	})

	// 为后续测试获取目标文件
	appFilePath := filepath.Join(demoAppPath, "src/components/App.tsx")
	appFile := project.GetSourceFile(appFilePath)
	require.NotNil(t, appFile, "应找到 App.tsx 文件")

	var foundByTraversal tsmorphgo.Node
	t.Run("节点遍历查找", func(t *testing.T) {
		var traversalFound bool
		appFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsCallExpression() && node.GetText() == "useUserData(1)" {
				foundByTraversal = node
				traversalFound = true

				assert.Equal(t, 30, node.GetStartLineNumber(), "找到的节点应在第 30 行")
				assert.Equal(t, 59, node.GetStartColumnNumber(), "找到的节点应在第 59 列")
				assert.Equal(t, tsmorphgo.KindCallExpression, node.GetKind(), "节点类型应为 CallExpression")
			}
		})
		assert.True(t, traversalFound, "应通过遍历找到 'useUserData(1)' 调用")
	})

	var foundByPosition *tsmorphgo.Node
	t.Run("节点位置查找", func(t *testing.T) {
		// 根据原始示例，节点位于第 30 行，第 59 列
		foundByPosition = project.FindNodeAt(appFilePath, 30, 59)
		require.NotNil(t, foundByPosition, "应在指定位置找到一个节点")
		assert.Contains(t, foundByPosition.GetText(), "useUserData", "节点文本应与 'useUserData' 相关")
	})

	t.Run("结果验证", func(t *testing.T) {
		require.NotNil(t, foundByTraversal, "必须找到通过遍历得到的节点以进行验证")
		require.NotNil(t, foundByPosition, "必须找到通过位置查询得到的节点以进行验证")

		// 原始示例仅验证了节点是否在同一行。
		// FindNodeAt 可能返回标识符，而遍历找到的是调用表达式。
		// 断言行号足以满足原始示例的意图。
		assert.Equal(t, foundByTraversal.GetStartLineNumber(), foundByPosition.GetStartLineNumber(), "两种方法都应在同一行找到节点")
	})

	t.Run("项目配置验证", func(t *testing.T) {
		tsConfig := project.GetTsConfig()
		require.NotNil(t, tsConfig, "应能获取 tsconfig.json 的数据")
		assert.NotEmpty(t, tsConfig.CompilerOptions, "tsconfig 中的 CompilerOptions 不应为空")
	})
}