package ts_bundle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getBatchTestProjectRoot 获取测试项目根目录（避免与 collect_test.go 中的函数冲突）
func getBatchTestProjectRoot(t *testing.T) string {
	projectRoot, err := filepath.Abs("testdata")
	require.NoError(t, err, "获取 testdata 的绝对路径失败")
	return projectRoot
}

// TestGenerateBatchBundlesToFiles 测试批量文件输出功能
func TestGenerateBatchBundlesToFiles(t *testing.T) {
	projectRoot := getBatchTestProjectRoot(t)

	testCases := []struct {
		name            string
		entries         []string
		expectedFiles   []string // 期望生成的文件名
		expectedTypes   []string // 每个文件中应包含的类型
		shouldError     bool
		errorContains   string
	}{
		{
			name: "基础批量打包 - 不同文件的不同类型",
			entries: []string{
				filepath.Join(projectRoot, "src", "utils", "user.ts") + ":User",
				filepath.Join(projectRoot, "src", "utils", "address.ts") + ":Address",
			},
			expectedFiles: []string{"User.d.ts", "Address.d.ts"},
			expectedTypes: []string{
				"interface User",
				"interface Address",
			},
		},
		{
			name: "别名功能测试",
			entries: []string{
				filepath.Join(projectRoot, "src", "utils", "user.ts") + ":User:UserDTO",
				filepath.Join(projectRoot, "src", "index.ts") + ":UserProfile:Profile",
			},
			expectedFiles: []string{"UserDTO.d.ts", "Profile.d.ts"},
			expectedTypes: []string{
				"interface User", // 类型声明仍然是原名
				"interface UserProfile",
			},
		},
		{
			name: "复杂类型依赖测试",
			entries: []string{
				filepath.Join(projectRoot, "src", "index.ts") + ":FullUser",
				filepath.Join(projectRoot, "src", "complex.ts") + ":UserWithoutAddress",
			},
			expectedFiles: []string{"FullUser.d.ts", "UserWithoutAddress.d.ts"},
			expectedTypes: []string{
				"type FullUser",
				"type UserWithoutAddress",
			},
		},
		{
			name: "不存在的类型应被跳过",
			entries: []string{
				filepath.Join(projectRoot, "src", "utils", "user.ts") + ":User",
				filepath.Join(projectRoot, "src", "utils", "user.ts") + ":NonExistentType",
				filepath.Join(projectRoot, "src", "utils", "address.ts") + ":Address",
			},
			expectedFiles: []string{"User.d.ts", "Address.d.ts"}, // 不存在的类型不应生成文件
			expectedTypes: []string{
				"interface User",
				"interface Address",
			},
		},
		{
			name: "错误格式测试",
			entries: []string{
				"invalid_format", // 缺少冒号
			},
			shouldError:   true,
			errorContains: "无效的入口格式",
		},
		{
			name: "命名空间导入测试",
			entries: []string{
				filepath.Join(projectRoot, "src", "index.ts") + ":UserId",
			},
			expectedFiles: []string{"UserId.d.ts"},
			expectedTypes: []string{
				"type UserId",
			},
		},
		{
			name: "路径别名测试",
			entries: []string{
				filepath.Join(projectRoot, "src", "path-alias.ts") + ":PathAliasUser",
			},
			expectedFiles: []string{"PathAliasUser.d.ts"},
			expectedTypes: []string{
				"interface PathAliasUser",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir() // 每个测试用例使用独立的临时目录
			results, err := GenerateBatchBundlesToFiles(tc.entries, projectRoot, tempDir)

			if tc.shouldError {
				assert.Error(t, err, "期望返回错误")
				assert.Contains(t, err.Error(), tc.errorContains, "错误信息应包含预期内容")
				return
			}

			require.NoError(t, err, "不应返回错误")
			assert.NotEmpty(t, results, "应返回结果")

			// 检查生成的文件数量
			if len(tc.expectedFiles) > 0 {
				assert.Equal(t, len(tc.expectedFiles), len(results), "生成的文件数量应匹配")
			}

			// 检查每个生成的文件
			for i, result := range results {
				// 检查文件是否存在
				assert.FileExists(t, result.FilePath, "生成的文件应存在")

				// 检查文件名
				if i < len(tc.expectedFiles) {
					assert.Equal(t, tc.expectedFiles[i], result.FileName, "文件名应匹配")
				}

				// 检查文件内容
				content, err := os.ReadFile(result.FilePath)
				require.NoError(t, err, "读取文件内容失败")

				// 检查内容不为空
				assert.NotEmpty(t, strings.TrimSpace(string(content)), "文件内容不应为空")

				// 检查预期的类型是否存在
				if i < len(tc.expectedTypes) {
					assert.Contains(t, string(content), tc.expectedTypes[i], "文件应包含预期的类型")
				}

				// 验证文件大小与报告的一致
				assert.Equal(t, len(content), result.ContentSize, "报告的内容大小应与实际一致")
			}
		})
	}
}

// TestGenerateFileName 测试文件名生成逻辑
func TestGenerateFileName(t *testing.T) {
	tempDir := t.TempDir()

	testCases := []struct {
		name           string
		entry          TypeEntryPoint
		expectedPrefix string
	}{
		{
			name: "基础类型名",
			entry: TypeEntryPoint{
				TypeName: "User",
			},
			expectedPrefix: "User.d.ts",
		},
		{
			name: "使用别名",
			entry: TypeEntryPoint{
				TypeName: "User",
				Alias:    "UserDTO",
			},
			expectedPrefix: "UserDTO.d.ts",
		},
		{
			name: "特殊字符清理",
			entry: TypeEntryPoint{
				TypeName: "User-Type",
				Alias:    "DTO.Type",
			},
			expectedPrefix: "DTO_Type.d.ts",
		},
		{
			name: "数字和字母组合",
			entry: TypeEntryPoint{
				TypeName: "Type123",
			},
			expectedPrefix: "Type123.d.ts",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fileName := generateFileName(tc.entry, tempDir)

			assert.True(t, strings.HasPrefix(fileName, tc.expectedPrefix), "文件名应以预期前缀开始")
			assert.True(t, strings.HasSuffix(fileName, ".d.ts"), "文件名应以 .d.ts 结尾")

			// 验证文件可以被创建
			fullPath := filepath.Join(tempDir, fileName)
			err := os.WriteFile(fullPath, []byte("test"), 0644)
			assert.NoError(t, err, "文件名应有效且可创建")
		})
	}
}

// TestBatchFileResult 测试批量文件结果结构
func TestBatchFileResult(t *testing.T) {
	result := BatchFileResult{
		EntryPoint: TypeEntryPoint{
			FilePath: "path/to/file.ts",
			TypeName: "TestType",
			Alias:    "TestAlias",
		},
		FileName:    "TestAlias.d.ts",
		FilePath:    "/tmp/TestAlias.d.ts",
		ContentSize: 1024,
	}

	assert.Equal(t, "TestType", result.EntryPoint.TypeName)
	assert.Equal(t, "TestAlias", result.EntryPoint.Alias)
	assert.Equal(t, "TestAlias.d.ts", result.FileName)
	assert.Equal(t, "/tmp/TestAlias.d.ts", result.FilePath)
	assert.Equal(t, 1024, result.ContentSize)
}

// TestGenerateBatchBundleFromStrings_Extended 扩展的字符串格式测试
func TestGenerateBatchBundleFromStrings_Extended(t *testing.T) {
	projectRoot := getBatchTestProjectRoot(t)

	testCases := []struct {
		name          string
		entries       []string
		shouldError   bool
		errorContains string
		expectedTypes []string
	}{
		{
			name: "三段式格式（文件:类型:别名）",
			entries: []string{
				filepath.Join(projectRoot, "src", "utils", "user.ts") + ":User:UserDTO",
			},
			expectedTypes: []string{"interface User"},
		},
		{
			name: "混合格式",
			entries: []string{
				filepath.Join(projectRoot, "src", "utils", "user.ts") + ":User",      // 两段式
				filepath.Join(projectRoot, "src", "utils", "address.ts") + ":Address:Addr", // 三段式
			},
			expectedTypes: []string{"interface User", "interface Addr"}, // 别名会重命名类型
		},
		{
			name: "格式错误 - 缺少类型名",
			entries: []string{
				filepath.Join(projectRoot, "src", "utils", "user.ts") + ":",
			},
			shouldError:   true,
			errorContains: "类型名不能为空",
		},
		{
			name: "格式错误 - 过多段",
			entries: []string{
				filepath.Join(projectRoot, "src", "utils", "user.ts") + ":User:Alias:Extra",
			},
			shouldError:   true,
			errorContains: "无效的入口格式",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content, err := GenerateBatchBundleFromStrings(tc.entries, projectRoot)

			if tc.shouldError {
				assert.Error(t, err, "期望返回错误")
				assert.Contains(t, err.Error(), tc.errorContains, "错误信息应包含预期内容")
				return
			}

			require.NoError(t, err, "不应返回错误")
			assert.NotEmpty(t, strings.TrimSpace(content), "内容不应为空")

			// 检查预期的类型是否存在
			for _, expectedType := range tc.expectedTypes {
				assert.Contains(t, content, expectedType, "内容应包含预期的类型")
			}
		})
	}
}

// TestDirectoryCreation 测试目录创建功能
func TestDirectoryCreation(t *testing.T) {
	projectRoot := getBatchTestProjectRoot(t)
	tempDir := t.TempDir()

	// 测试嵌套目录创建
	nestedDir := filepath.Join(tempDir, "level1", "level2", "level3")
	entries := []string{
		filepath.Join(projectRoot, "src", "utils", "user.ts") + ":User",
	}

	results, err := GenerateBatchBundlesToFiles(entries, projectRoot, nestedDir)
	require.NoError(t, err, "创建嵌套目录不应失败")
	assert.NotEmpty(t, results, "应返回结果")

	// 验证目录被创建
	assert.DirExists(t, nestedDir, "嵌套目录应被创建")

	// 验证文件被创建在正确的位置
	assert.FileExists(t, results[0].FilePath, "文件应在嵌套目录中创建")
	assert.True(t, strings.HasPrefix(results[0].FilePath, nestedDir), "文件路径应以嵌套目录开始")
}

// TestEmptyOutputHandling 测试空输出处理
func TestEmptyOutputHandling(t *testing.T) {
	projectRoot := getBatchTestProjectRoot(t)
	tempDir := t.TempDir()

	// 只有不存在的类型
	entries := []string{
		filepath.Join(projectRoot, "src", "utils", "user.ts") + ":NonExistentType1",
		filepath.Join(projectRoot, "src", "utils", "user.ts") + ":NonExistentType2",
	}

	results, err := GenerateBatchBundlesToFiles(entries, projectRoot, tempDir)
	require.NoError(t, err, "不应返回错误")
	assert.Empty(t, results, "应返回空结果")

	// 验证没有文件被创建
	files, err := os.ReadDir(tempDir)
	require.NoError(t, err, "读取目录不应失败")
	assert.Empty(t, files, "目录应为空")
}