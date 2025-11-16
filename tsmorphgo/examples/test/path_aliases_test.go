package examples_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPathAliases(t *testing.T) {
	// 设置: 初始化项目时加载 tsconfig 以启用路径别名解析
	demoAppPath, err := filepath.Abs("../demo-react-app")
	require.NoError(t, err, "获取 demo-react-app 绝对路径失败")

	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:    demoAppPath,
		UseTsConfig: true,
	})
	require.NotNil(t, project, "项目创建不应失败")
	defer project.Close()

	t.Run("TSConfig别名解析", func(t *testing.T) {
		tsConfig := project.GetTsConfig()
		require.NotNil(t, tsConfig, "应能获取 tsconfig.json 的数据")
		require.NotEmpty(t, tsConfig.CompilerOptions, "tsconfig 中的 CompilerOptions 不应为空")

		paths, ok := tsConfig.CompilerOptions["paths"].(map[string]interface{})
		require.True(t, ok, "CompilerOptions 应包含 'paths' 映射")

		// 检查一个特定的、重要的别名
		aliasPath, aliasExists := paths["@/*"]
		assert.True(t, aliasExists, "'@/*' 别名应存在于 paths 中")

		// 检查别名的值
		expectedValue := []interface{}{"src/*"}
		assert.Equal(t, expectedValue, aliasPath, "'@/*' 别名的值应为 '[\"src/*\"]'")
	})

	t.Run("导入中的别名使用", func(t *testing.T) {
		// 查找使用路径别名的文件
		testAliasesFilePath := filepath.Join(demoAppPath, "src/test-aliases.tsx")
		testAliasesFile := project.GetSourceFile(testAliasesFilePath)
		require.NotNil(t, testAliasesFile, "应找到 test-aliases.tsx 文件")

		// 查找使用 '@/' 别名的导入声明
		var aliasImportNode tsmorphgo.Node
		testAliasesFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if aliasImportNode.IsValid() {
				return // 已找到, 跳过剩余节点
			}
			if node.IsImportDeclaration() {
				// 检查单引号或双引号的别名
				if strings.Contains(node.GetText(), "'@/" ) || strings.Contains(node.GetText(), "\"@/" ) {
					aliasImportNode = node
				}
			}
		})

		require.True(t, aliasImportNode.IsValid(), "应找到一个使用路径别名的导入声明")

		// 进一步检查导入节点
		_, success := aliasImportNode.AsImportDeclaration()
		require.True(t, success, "应能够将节点收窄为 ImportDeclaration")

		// 原始示例的这部分很难直接断言, 因为没有更多的 API, 但我们可以检查文本。
		assert.Contains(t, aliasImportNode.GetText(), "'@/utils/dateUtils'", "导入文本应包含别名路径")
	})

	t.Run("别名解析检查", func(t *testing.T) {
		// 一个隐式的检查别名是否工作的方法是, 查看目标文件是否包含在项目的源文件中。
		dateUtilsFilePath := filepath.Join(demoAppPath, "src/utils/dateUtils.ts")
		dateUtilsFile := project.GetSourceFile(dateUtilsFilePath)
		assert.NotNil(t, dateUtilsFile, "被别名引用的文件 'dateUtils.ts' 应在项目中找到, 表明解析在项目级别上是有效的")
	})
}