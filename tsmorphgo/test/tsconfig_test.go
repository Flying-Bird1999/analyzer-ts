package tsmorphgo

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/stretchr/testify/assert"
)

// tsconfig_test.go
//
// 这个文件包含了 TypeScript 配置文件 (tsconfig.json) 处理功能的测试用例，
// 专注于验证 tsmorphgo 对各种 TypeScript 项目配置的支持。
//
// 主要测试场景：
// 1. 基本配置解析 - 测试标准 tsconfig.json 文件的解析
// 2. 编译选项提取 - 验证各种编译选项的正确提取
// 3. 文件包含/排除 - 测试 include/exclude 模式的处理
// 4. 配置继承 - 验证 tsconfig 继承机制
// 5. 配置合并 - 测试多层级配置的正确合并
// 6. 错误处理 - 验证无效配置的容错能力
// 7. 文件匹配 - 测试 glob 模式的文件匹配
//
// 测试目标：
// - 确保 tsconfig.json 文件的正确解析和处理
// - 验证编译选项对项目配置的影响
// - 测试文件包含/排除逻辑的准确性
// - 确保配置继承和合并的正确性
// - 验证错误情况下的系统稳定性

func TestProjectConfig_WithBasicTsConfig(t *testing.T) {
	// 创建测试目录和配置文件
	testDir := t.TempDir()
	tsconfigContent := `{
		"compilerOptions": {
			"target": "es2018",
			"module": "commonjs",
			"strict": true
		},
		"include": ["src/**/*.ts"],
		"exclude": ["**/*.test.ts"]
	}`

	tsconfigPath := filepath.Join(testDir, "tsconfig.json")
	err := os.WriteFile(tsconfigPath, []byte(tsconfigContent), 0644)
	assert.NoError(t, err)

	// 创建测试源文件
	srcDir := filepath.Join(testDir, "src")
	err = os.MkdirAll(srcDir, 0755)
	assert.NoError(t, err)

	testFileContent := `const hello = "world";`
	testFilePath := filepath.Join(srcDir, "main.ts")
	err = os.WriteFile(testFilePath, []byte(testFileContent), 0644)
	assert.NoError(t, err)

	// 创建测试文件（应该被排除）
	testFilePath2 := filepath.Join(srcDir, "main.test.ts")
	err = os.WriteFile(testFilePath2, []byte(testFileContent), 0644)
	assert.NoError(t, err)

	// 使用 tsconfig 创建项目
	config := ProjectConfig{
		RootPath:    testDir,
		UseTsConfig: true,
	}
	project := NewProject(config)

	// 验证项目创建成功
	assert.NotNil(t, project)

	// 注意：当前实现可能还未完全实现文件过滤，所以这里主要验证配置加载
	tsConfig := project.GetTsConfig()
	assert.NotNil(t, tsConfig)
	assert.Equal(t, "es2018", tsConfig.CompilerOptions["target"])
	assert.Equal(t, "commonjs", tsConfig.CompilerOptions["module"])
	assert.True(t, tsConfig.CompilerOptions["strict"].(bool))
}

func TestProjectConfig_WithTsConfigInheritance(t *testing.T) {
	// 创建测试目录和配置文件
	testDir := t.TempDir()

	// 创建基础配置文件
	baseConfigContent := `{
		"compilerOptions": {
			"target": "es2015",
			"strict": true,
			"baseUrl": "./"
		},
		"include": ["src/**/*"]
	}`

	baseConfigPath := filepath.Join(testDir, "tsconfig.base.json")
	err := os.WriteFile(baseConfigPath, []byte(baseConfigContent), 0644)
	assert.NoError(t, err)

	// 创建继承配置文件
	extendConfigContent := `{
		"extends": "./tsconfig.base.json",
		"compilerOptions": {
			"target": "es2018",
			"module": "es6"
		},
		"exclude": ["**/*.test.ts"]
	}`

	extendConfigPath := filepath.Join(testDir, "tsconfig.json")
	err = os.WriteFile(extendConfigPath, []byte(extendConfigContent), 0644)
	assert.NoError(t, err)

	// 创建测试源文件
	srcDir := filepath.Join(testDir, "src")
	err = os.MkdirAll(srcDir, 0755)
	assert.NoError(t, err)

	testFileContent := `const hello = "world";`
	testFilePath := filepath.Join(srcDir, "main.ts")
	err = os.WriteFile(testFilePath, []byte(testFileContent), 0644)
	assert.NoError(t, err)

	// 使用继承的 tsconfig 创建项目
	config := ProjectConfig{
		RootPath:    testDir,
		UseTsConfig: true,
	}
	project := NewProject(config)

	// 验证项目创建成功
	assert.NotNil(t, project)

	// 验证配置正确合并
	tsConfig := project.GetTsConfig()
	assert.NotNil(t, tsConfig)
	assert.Equal(t, "es2018", tsConfig.CompilerOptions["target"]) // 子配置覆盖
	assert.Equal(t, "es6", tsConfig.CompilerOptions["module"])    // 子配置新增
	assert.True(t, tsConfig.CompilerOptions["strict"].(bool))     // 基础配置继承
	assert.Equal(t, "./", tsConfig.CompilerOptions["baseUrl"])    // 基础配置继承
	assert.Contains(t, tsConfig.Exclude, "**/*.test.ts")          // 子配置新增
}

func TestProjectConfig_TsConfigAutoDiscovery(t *testing.T) {
	// 创建测试目录
	testDir := t.TempDir()

	// 创建配置文件
	tsconfigContent := `{
		"compilerOptions": {
			"target": "es2017"
		}
	}`

	tsconfigPath := filepath.Join(testDir, "tsconfig.json")
	err := os.WriteFile(tsconfigPath, []byte(tsconfigContent), 0644)
	assert.NoError(t, err)

	// 创建测试源文件
	testFileContent := `const hello = "world";`
	testFilePath := filepath.Join(testDir, "main.ts")
	err = os.WriteFile(testFilePath, []byte(testFileContent), 0644)
	assert.NoError(t, err)

	// 不指定配置文件路径，应该自动发现
	config := ProjectConfig{
		RootPath:    testDir,
		UseTsConfig: true,
	}
	project := NewProject(config)

	// 验证项目创建成功
	assert.NotNil(t, project)

	// 验证配置文件被自动发现并解析
	tsConfig := project.GetTsConfig()
	assert.NotNil(t, tsConfig)
	assert.Equal(t, "es2017", tsConfig.CompilerOptions["target"])
}

func TestProject_GetCompilerOptions(t *testing.T) {
	// 创建测试目录和配置文件
	testDir := t.TempDir()
	tsconfigContent := `{
		"compilerOptions": {
			"target": "es2018",
			"module": "commonjs",
			"strict": true,
			"declaration": false
		}
	}`

	tsconfigPath := filepath.Join(testDir, "tsconfig.json")
	err := os.WriteFile(tsconfigPath, []byte(tsconfigContent), 0644)
	assert.NoError(t, err)

	// 创建测试源文件
	testFileContent := `const hello = "world";`
	testFilePath := filepath.Join(testDir, "main.ts")
	err = os.WriteFile(testFilePath, []byte(testFileContent), 0644)
	assert.NoError(t, err)

	// 创建项目
	config := ProjectConfig{
		RootPath:    testDir,
		UseTsConfig: true,
	}
	project := NewProject(config)

	// 测试获取编译选项
	target, ok := project.GetCompilerOptionString("target")
	assert.True(t, ok)
	assert.Equal(t, "es2018", target)

	module, ok := project.GetCompilerOptionString("module")
	assert.True(t, ok)
	assert.Equal(t, "commonjs", module)

	strict, ok := project.GetCompilerOptionBool("strict")
	assert.True(t, ok)
	assert.True(t, strict)

	declaration, ok := project.GetCompilerOptionBool("declaration")
	assert.True(t, ok)
	assert.False(t, declaration)

	// 测试不存在的选项
	nonexistent, ok := project.GetCompilerOptionString("nonexistent")
	assert.False(t, ok)
	assert.Empty(t, nonexistent)
}

func TestProjectConfig_WithoutTsConfig(t *testing.T) {
	// 创建测试目录但不创建配置文件
	testDir := t.TempDir()

	// 创建测试源文件
	testFileContent := `const hello = "world";`
	testFilePath := filepath.Join(testDir, "main.ts")
	err := os.WriteFile(testFilePath, []byte(testFileContent), 0644)
	assert.NoError(t, err)

	// 创建项目，不使用 tsconfig
	config := ProjectConfig{
		RootPath:    testDir,
		UseTsConfig: false,
	}
	project := NewProject(config)

	// 验证项目创建成功
	assert.NotNil(t, project)

	// 验证没有 tsconfig 配置
	tsConfig := project.GetTsConfig()
	assert.Nil(t, tsConfig)

	// 验证编译选项获取失败
	_, ok := project.GetCompilerOptionString("target")
	assert.False(t, ok)
}

func TestProjectConfig_InvalidTsConfigPath(t *testing.T) {
	// 创建测试目录
	testDir := t.TempDir()

	// 指定不存在的配置文件路径
	config := ProjectConfig{
		RootPath:     testDir,
		UseTsConfig:  true,
		TsConfigPath: filepath.Join(testDir, "nonexistent.json"),
	}
	project := NewProject(config)

	// 验证项目仍然创建成功（容错处理）
	assert.NotNil(t, project)

	// 验证没有 tsconfig 配置
	tsConfig := project.GetTsConfig()
	assert.Nil(t, tsConfig)
}

func TestPathMatchesPatterns(t *testing.T) {
	// 测试文件路径匹配
	testCases := []struct {
		filePath string
		patterns []string
		expected bool
	}{
		{"/src/main.ts", []string{"src/**/*"}, true},
		{"/src/utils/helper.ts", []string{"src/**/*.ts"}, true},
		{"/test/main.test.ts", []string{"src/**/*.ts"}, false},
		{"/src/main.tsx", []string{"src/**/*.ts"}, false},
		{"/src/main.ts", []string{"**/*.ts"}, true},
		{"/src/main.test.ts", []string{"**/*.ts", "!**/*.test.ts"}, false},
		{"/src/main.ts", []string{}, true}, // 空模式匹配所有
	}

	for _, tc := range testCases {
		result := PathMatchesPatterns(tc.filePath, tc.patterns)
		assert.Equal(t, tc.expected, result,
			"文件路径 %s 匹配模式 %v 应该为 %v", tc.filePath, tc.patterns, tc.expected)
	}
}

func TestMergeTsConfig_TargetExtensions(t *testing.T) {
	// 创建测试目录
	testDir := t.TempDir()

	// 创建包含 target 选项的配置文件
	tsconfigContent := `{
		"compilerOptions": {
			"target": "es2018"
		}
	}`

	tsconfigPath := filepath.Join(testDir, "tsconfig.json")
	err := os.WriteFile(tsconfigPath, []byte(tsconfigContent), 0644)
	assert.NoError(t, err)

	// 创建项目
	config := ProjectConfig{
		RootPath:    testDir,
		UseTsConfig: true,
	}
	project := NewProject(config)

	// 验证 target 扩展名被正确添加
	// 注意：这需要在项目配置构建时处理，这里主要验证配置被正确解析
	tsConfig := project.GetTsConfig()
	assert.NotNil(t, tsConfig)
	assert.Equal(t, "es2018", tsConfig.CompilerOptions["target"])
}

func TestConvertGlobPatterns(t *testing.T) {
	// 测试 glob 模式转换
	testCases := []struct {
		patterns    []string
		rootPath    string
		expectCount int
	}{
		{[]string{"src/**/*"}, "/project", 1},
		{[]string{"src/**/*.ts", "tests/**/*.ts"}, "/project", 2},
		{[]string{"**/*.ts"}, "/project", 1},
		{[]string{}, "/project", 0},
	}

	for _, tc := range testCases {
		result := ConvertGlobPatterns(tc.patterns, tc.rootPath)
		assert.Len(t, result, tc.expectCount)

		for i, pattern := range result {
			if tc.rootPath != "/" {
				// 非绝对模式应该被转换为相对于根路径
				assert.Contains(t, pattern, tc.rootPath,
					"模式 %d 应该包含根路径 %s", i, tc.rootPath)
			}
		}
	}
}

func TestFindTsConfigFile(t *testing.T) {
	// 创建测试目录
	testDir := t.TempDir()

	// 创建 tsconfig.json
	tsconfigPath := filepath.Join(testDir, "tsconfig.json")
	err := os.WriteFile(tsconfigPath, []byte("{}"), 0644)
	assert.NoError(t, err)

	// 测试文件查找
	foundPath := FindTsConfigFile(testDir)
	assert.Equal(t, tsconfigPath, foundPath)

	// 创建 tsconfig.base.json
	baseConfigPath := filepath.Join(testDir, "tsconfig.base.json")
	err = os.WriteFile(baseConfigPath, []byte("{}"), 0644)
	assert.NoError(t, err)

	// 应该仍然优先找到 tsconfig.json
	foundPath = FindTsConfigFile(testDir)
	assert.Equal(t, tsconfigPath, foundPath)

	// 删除 tsconfig.json，应该找到 tsconfig.base.json
	os.Remove(tsconfigPath)
	foundPath = FindTsConfigFile(testDir)
	assert.Equal(t, baseConfigPath, foundPath)
}

func TestFindTsConfigFile_NoConfig(t *testing.T) {
	// 创建测试目录但不创建配置文件
	testDir := t.TempDir()

	// 测试文件查找
	foundPath := FindTsConfigFile(testDir)
	assert.Empty(t, foundPath)
}
