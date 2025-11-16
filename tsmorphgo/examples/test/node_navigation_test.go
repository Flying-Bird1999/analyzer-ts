package examples_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeNavigationAndTypeNarrowing(t *testing.T) {
	// 设置: 初始化项目并获取目标文件
	demoAppPath, err := filepath.Abs("../demo-react-app")
	require.NoError(t, err, "获取 demo-react-app 绝对路径失败")

	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:    demoAppPath,
		UseTsConfig: true,
	})
	require.NotNil(t, project, "项目创建不应失败")
	defer project.Close()

	useUserDataFilePath := filepath.Join(demoAppPath, "src/hooks/useUserData.ts")
	useUserDataFile := project.GetSourceFile(useUserDataFilePath)
	require.NotNil(t, useUserDataFile, "应找到 useUserData.ts 文件")

	// 查找用于测试的目标节点
	var targetNode tsmorphgo.Node
	useUserDataFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsVariableDeclaration() && strings.Contains(node.GetText(), "useUserData =") {
			targetNode = node
			return // 找到后停止
		}
	})
	require.NotNil(t, targetNode, "应找到 'useUserData' 变量声明节点")
	assert.Equal(t, tsmorphgo.KindVariableDeclaration, targetNode.GetKind())

	t.Run("父节点导航", func(t *testing.T) {
		parentNode := targetNode.GetParent()
		require.NotNil(t, parentNode, "父节点不应为 nil")
		// 在 TypeScript AST 中, 一个 VariableDeclaration 位于一个 VariableDeclarationList 内部,
		// 而该 List 则位于一个 VariableStatement 内部。因此直接父节点是 List。
		assert.True(t, parentNode.IsKind(tsmorphgo.KindVariableDeclarationList), "父节点应为 VariableDeclarationList")
	})

	t.Run("祖先节点导航", func(t *testing.T) {
		ancestors := targetNode.GetAncestors()
		assert.NotEmpty(t, ancestors, "应有祖先节点")

		// 最后一个祖先节点应该是 SourceFile
		if len(ancestors) > 0 {
			lastAncestor := ancestors[len(ancestors)-1]
			assert.Equal(t, tsmorphgo.KindSourceFile, lastAncestor.GetKind(), "最后的祖先节点应为 SourceFile")
		}

		// 按类型查找特定的祖先节点
		varStatement, found := targetNode.GetFirstAncestorByKind(tsmorphgo.KindVariableStatement)
		assert.True(t, found, "应找到一个 VariableStatement 类型的祖先")
		require.NotNil(t, varStatement)
		assert.Equal(t, tsmorphgo.KindVariableStatement, varStatement.GetKind())
	})

	t.Run("类型收窄为VariableDeclaration", func(t *testing.T) {
		varDecl, success := targetNode.AsVariableDeclaration()
		require.True(t, success, "应成功将节点收窄为 VariableDeclaration")

		// 测试 VariableDeclaration 特有的 API
		assert.Equal(t, "useUserData", varDecl.GetName(), "变量名应为 'useUserData'")
		assert.True(t, varDecl.HasInitializer(), "变量应有初始化器")

		initializer := varDecl.GetInitializer()
		require.NotNil(t, initializer, "初始化器不应为 nil")
		assert.True(t, strings.Contains(initializer.GetText(), "=>"), "初始化器应为箭头函数 (通过文本推断)")

		nameNode := varDecl.GetNameNode()
		require.NotNil(t, nameNode, "名称节点不应为 nil")
		assert.True(t, nameNode.IsKind(tsmorphgo.KindIdentifier), "名称节点应为 Identifier")
		assert.Equal(t, "useUserData", nameNode.GetText())
	})
}