package ts_bundle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// getTestProjectRoot 是一个测试辅助函数，用于获取测试数据目录（`testdata`）的绝对路径。
// 为了保证测试的稳定性，我们所有的测试用例都基于这个固定的目录结构进行。
func getTestProjectRoot(t *testing.T) string {
	projectRoot, err := filepath.Abs("testdata")
	assert.NoError(t, err, "获取 testdata 的绝对路径失败")
	return projectRoot
}

// TestGenerateBundle 是一个全面的表驱动测试，它系统地覆盖了 `GenerateBundle` 函数的各种核心功能和边缘场景。
// 表驱动测试（Table-Driven Test）是一种将测试用例的数据和断言逻辑分离的设计模式。
// 所有的测试场景都定义在一个名为 `testCases` 的切片中，每个元素代表一个独立的测试用例。
// 这种结构极大地提高了测试代码的可读性、可维护性和可扩展性。当需要添加新的测试场景时，只需向 `testCases` 切片中增加一个新的结构体实例即可。
func TestGenerateBundle(t *testing.T) {
	projectRoot := getTestProjectRoot(t)

	// testCases 定义了一系列测试场景。
	testCases := []struct {
		name                 string   // name: 测试用例的描述，会显示在 `go test -v` 的日志中，便于快速定位问题。
		entryFile            string   // entryFile: 作为依赖收集起点的文件路径。
		typeName             string   // typeName: 需要收集的目标类型名称。
		expectedToContain    []string // expectedToContain: 一个字符串切片，断言最终生成的 bundle 文件中必须包含这些字符串。
		expectedToNotContain []string // expectedToNotContain: 一个字符串切片，断言最终生成的 bundle 文件中必须不包含这些字符串。
	}{
		{
			name:      "基础场景：简单接口",
			entryFile: filepath.Join(projectRoot, "src", "utils", "user.ts"),
			typeName:  "User",
			expectedToContain: []string{
				"interface User",
				"id: number",
				"name: string",
			},
			expectedToNotContain: []string{
				"UserRole", // 确保不会收集同一文件中的其他无关类型。
			},
		},
		{
			name:      "基础场景：联合类型",
			entryFile: filepath.Join(projectRoot, "src", "utils", "user.ts"),
			typeName:  "UserRole",
			expectedToContain: []string{
				"type UserRole",
				"'admin' | 'user'",
			},
		},
		{
			name:      "基础场景：枚举类型",
			entryFile: filepath.Join(projectRoot, "src", "utils", "user.ts"),
			typeName:  "UserStatus",
			expectedToContain: []string{
				"enum UserStatus",
				"Active = 'active'",
				"Inactive = 'inactive'",
			},
		},
		{
			name:      "依赖收集：接口继承 (extends)",
			entryFile: filepath.Join(projectRoot, "src", "utils", "user.ts"),
			typeName:  "AdminUser",
			expectedToContain: []string{
				"interface AdminUser",
				"extends User",
				"role: UserRole",
				"status: UserStatus",
				// 验证其依赖项（User, UserRole, UserStatus）是否也被正确收集。
				"interface User",
				"type UserRole",
				"enum UserStatus",
			},
		},
		{
			name:      "依赖收集：跨文件导入与命名冲突",
			entryFile: filepath.Join(projectRoot, "src", "index.ts"),
			typeName:  "UserProfile",
			expectedToContain: []string{
				"interface UserProfile",
				"extends User",
				"address: Address",
				"tags: Common_CommonType[]", // 验证 CommonType 因为命名冲突被重命名为 Common_CommonType。
				// 验证所有直接和间接依赖项都被收集。
				"interface User",
				"interface Address",
				"type Common_CommonType", // 验证被重命名的依赖项本身也被正确收集。
			},
		},
		{
			name:      "依赖收集：命名空间导入 (import * as ...)",
			entryFile: filepath.Join(projectRoot, "src", "index.ts"),
			typeName:  "UserId",
			expectedToContain: []string{
				"type UserId",
				"Common_CommonInterface['id']", // 验证通过命名空间访问的类型被正确处理和重命名。
				// 验证相关依赖项被收集。
				"interface Common_CommonInterface",
				"id: CommonType",
				"type CommonType",
			},
		},
		{
			name:      "依赖收集：类型组合 (&)",
			entryFile: filepath.Join(projectRoot, "src", "index.ts"),
			typeName:  "FullUser",
			expectedToContain: []string{
				"type FullUser",
				"UserProfile & AdminUser",
				// 验证组合的所有部分及其深层依赖都被正确收集。
				"interface UserProfile",
				"interface AdminUser",
				"interface User",
				"type UserRole",
				"enum UserStatus",
				"interface Address",
				"type Common_CommonType", // 验证重名类型。
			},
		},
		{
			name:      "模块解析：NPM 包导入",
			entryFile: filepath.Join(projectRoot, "src", "external.ts"),
			typeName:  "LocalTypeWithExternal",
			expectedToContain: []string{
				"interface LocalTypeWithExternal",
				"externalProp: ExternalType",
				// 验证能成功从 node_modules/some-package/index.d.ts 中收集类型。
				"interface ExternalType",
				"externalId: number",
			},
		},
		{
			name:      "高级类型：Omit 工具类型",
			entryFile: filepath.Join(projectRoot, "src", "complex.ts"),
			typeName:  "UserWithoutAddress",
			expectedToContain: []string{
				"type UserWithoutAddress = Omit<FullUser, 'address'>",
				// 验证 Omit 的泛型参数类型被正确收集。
				"type FullUser",
			},
		},
		{
			name:      "高级类型：Pick 工具类型",
			entryFile: filepath.Join(projectRoot, "src", "complex.ts"),
			typeName:  "UserBasicInfo",
			expectedToContain: []string{
				"type UserBasicInfo = Pick<FullUser, 'id' | 'name'>",
				// 验证 Pick 的泛型参数类型被正确收集。
				"type FullUser",
			},
		},
		{
			name:      "高级类型：keyof 操作符",
			entryFile: filepath.Join(projectRoot, "src", "complex.ts"),
			typeName:  "UserFields",
			expectedToContain: []string{
				"[K in keyof FullUser]?: FullUser[K]",
				// 验证 keyof 的操作对象类型被正确收集。
				"type FullUser",
			},
		},
		{
			name:      "高级类型：索引访问类型 (T[K])",
			entryFile: filepath.Join(projectRoot, "src", "complex.ts"),
			typeName:  "UserName",
			expectedToContain: []string{
				"type UserName = FullUser['name']",
				// 验证索引访问的对象类型被正确收集。
				"type FullUser",
			},
		},
		{
			name:      "模块解析：tsconfig.json 路径别名 (@/...)",
			entryFile: filepath.Join(projectRoot, "src", "path-alias.ts"),
			typeName:  "PathAliasUser",
			expectedToContain: []string{
				"interface PathAliasUser",
				"extends AliasUser",
				"role: AliasRole",
				// 验证能通过路径别名找到并收集依赖。
				"interface AliasUser",
				"type AliasRole",
			},
		},
		{
			name:      "边缘场景：循环依赖",
			entryFile: filepath.Join(projectRoot, "src", "circular.ts"),
			typeName:  "CircularType",
			expectedToContain: []string{
				"interface CircularType",
				"b?: CircularBType",
				"interface CircularBType",
				"a?: CircularAType",
				"interface CircularAType",
			},
		},
		{
			name:      "边缘场景：请求一个不存在的类型",
			entryFile: filepath.Join(projectRoot, "src", "utils", "user.ts"),
			typeName:  "NonExistentType",
			// 期望返回空内容且无错误。
			expectedToContain:    []string{},
			expectedToNotContain: []string{"interface", "type", "enum"},
		},
		{
			name:      "导出模式：默认导出 (export default)",
			entryFile: filepath.Join(projectRoot, "src", "default-export.ts"),
			typeName:  "DefaultExportedType",
			expectedToContain: []string{
				"interface DefaultExportedType",
				"id: string",
				"value: boolean",
			},
		},
		{
			name:      "导出模式：重命名重新导出 (export { A as B } from ...)",
			entryFile: filepath.Join(projectRoot, "src", "re-export.ts"),
			typeName:  "ReExportedType",
			expectedToContain: []string{
				// 打包器应该正确地将类型重命名为其导出的别名。
				"interface ReExportedType",
				"message: string",
			},
		},
		{
			name:      "导出模式：通配符重新导出 (export * from ...)",
			entryFile: filepath.Join(projectRoot, "src", "re-export.ts"),
			typeName:  "User", // 这个类型来自通配符导出 `export * from './utils/user'`
			expectedToContain: []string{
				"interface User",
				"id: number",
				"name: string",
			},
			expectedToNotContain: []string{
				"AdminUser", // 确保不会引入无关的类型。
			},
		},
		{
			name:      "全局类型：从 .d.ts 文件收集环境类型",
			entryFile: filepath.Join(projectRoot, "src", "uses-global.ts"),
			typeName:  "UsesGlobal",
			expectedToContain: []string{
				"interface UsesGlobal",
				"prop: MyGlobalType",
				// 验证 `MyGlobalType`（在 global.d.ts 中声明）被成功收集。
				"interface MyGlobalType",
				"id: string",
			},
		},
	}

	// 遍历并执行所有测试用例。
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 调用核心函数 GenerateBundle 进行测试。
			bundledContent, err := GenerateBundle(tc.entryFile, tc.typeName, projectRoot)

			// 所有表中的测试用例都预期成功执行，不应返回错误。
			assert.NoError(t, err, "GenerateBundle 应该成功执行，不产生错误")

			// 特殊处理“不存在类型”的用例，预期返回一个完全空的字符串。
			if tc.typeName == "NonExistentType" {
				assert.Equal(t, "", bundledContent, "对于不存在的类型，捆绑内容应为空")
				return
			}

			// 对于所有其他情况，检查预期包含和不应包含的子字符串。
			// 这种断言方式比精确匹配整个文件内容更健壮，因为它对代码格式（如空格、换行）不敏感。
			for _, expected := range tc.expectedToContain {
				assert.Contains(t, bundledContent, expected, "捆绑内容应包含预期的字符串")
			}

			for _, notExpected := range tc.expectedToNotContain {
				assert.NotContains(t, bundledContent, notExpected, "捆绑内容不应包含意外的字符串")
			}

			// 一个通用检查，确保对于有效类型，生成的内容不为空。
			if len(tc.expectedToContain) > 0 {
				assert.NotEmpty(t, strings.TrimSpace(bundledContent), "对于存在的类型，捆绑内容不应为空")
			}
		})
	}
}

// TestGenerateBundle_ErrorCases 测试预期会返回错误的场景。
func TestGenerateBundle_ErrorCases(t *testing.T) {
	projectRoot := getTestProjectRoot(t)

	t.Run("边缘场景：入口文件不存在", func(t *testing.T) {
		entryFile := filepath.Join(projectRoot, "src", "non-existent-file.ts")
		typeName := "AnyType"

		_, err := GenerateBundle(entryFile, typeName, projectRoot)

		// 对于一个不存在的文件，底层的解析器应该会失败，我们预期这里会收到一个错误。
		assert.Error(t, err, "对于不存在的入口文件，GenerateBundle 应该返回一个错误")
	})
}