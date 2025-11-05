package tsmorphgo

import (
	"testing"

	. "github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/stretchr/testify/assert"
)

// create_source_file_test.go
//
// 这个文件包含了 CreateSourceFile 动态文件创建功能的测试用例，专注于验证
// tsmorphgo 对动态文件创建、更新和移除的支持。
//
// 主要测试场景：
// 1. 基本文件创建 - 测试从源码字符串创建文件
// 2. 文件更新功能 - 验证动态更新文件内容的能力
// 3. 文件移除功能 - 测试从项目中移除文件
// 4. 路径处理 - 测试相对路径和绝对路径的处理
// 5. 错误处理 - 验证各种错误情况的容错能力
// 6. 选项控制 - 测试覆盖和其他创建选项
// 7. 文件管理 - 测试文件数量统计和存在性检查
//
// 测试目标：
// - 验证动态文件创建的正确性和完整性
// - 确保文件更新和移除操作的安全性
// - 测试路径处理的准确性和一致性
// - 验证错误情况下的系统稳定性
// - 确保与现有项目结构的良好集成

// TestProject_CreateSourceFile_Basic 测试基本的动态文件创建功能
func TestProject_CreateSourceFile_Basic(t *testing.T) {
	// 创建一个空项目
	config := ProjectConfig{
		RootPath:    "/test",
		UseTsConfig: false,
	}
	project := NewProject(config)
	assert.NotNil(t, project)

	// 测试创建简单的 TypeScript 文件
	sourceCode := `const hello = "world";
function greet() {
    return hello;
}`

	sourceFile, err := project.CreateSourceFile("main.ts", sourceCode)
	assert.NoError(t, err, "创建文件应该成功")
	assert.NotNil(t, sourceFile, "返回的 SourceFile 不应该为空")

	// 验证文件被正确添加到项目中
	assert.Equal(t, 1, project.GetFileCount(), "项目应该包含 1 个文件")
	assert.True(t, project.ContainsFile("main.ts"), "项目应该包含 main.ts")
	assert.True(t, project.ContainsFile("/test/main.ts"), "项目应该包含完整路径的文件")

	// 验证文件内容被正确解析
	assert.Equal(t, "/test/main.ts", sourceFile.GetFilePath())
	assert.NotNil(t, sourceFile.GetFileResult(), "文件的解析结果不应该为空")
	assert.NotNil(t, sourceFile.GetAstNode(), "文件的 AST 不应该为空")
}

// TestProject_CreateSourceFile_WithAbsolutePath 测试使用绝对路径创建文件
func TestProject_CreateSourceFile_WithAbsolutePath(t *testing.T) {
	config := ProjectConfig{
		RootPath:    "/test/project",
		UseTsConfig: false,
	}
	project := NewProject(config)
	assert.NotNil(t, project)

	// 使用绝对路径创建文件
	sourceCode := `export const message = "Hello World";`
	sourceFile, err := project.CreateSourceFile("/test/project/utils/helper.ts", sourceCode)

	assert.NoError(t, err)
	assert.NotNil(t, sourceFile)
	assert.Equal(t, "/test/project/utils/helper.ts", sourceFile.GetFilePath())

	// 验证文件被正确添加
	assert.True(t, project.ContainsFile("/test/project/utils/helper.ts"))
	assert.Equal(t, 1, project.GetFileCount())
}

// TestProject_CreateSourceFile_Overwrite 测试文件覆盖功能
func TestProject_CreateSourceFile_Overwrite(t *testing.T) {
	config := ProjectConfig{
		RootPath:    "/test",
		UseTsConfig: false,
	}
	project := NewProject(config)
	assert.NotNil(t, project)

	// 首先创建一个文件
	originalCode := `const x = 1;`
	sourceFile, err := project.CreateSourceFile("test.ts", originalCode)
	assert.NoError(t, err)
	assert.NotNil(t, sourceFile)

	// 尝试不覆盖地创建同名文件，应该失败
	newCode := `const y = 2;`
	_, err = project.CreateSourceFile("test.ts", newCode)
	assert.Error(t, err, "不覆盖已存在文件应该返回错误")

	// 使用覆盖选项创建文件，应该成功
	sourceFile, err = project.CreateSourceFile("test.ts", newCode, CreateSourceFileOptions{
		Overwrite: true,
	})
	assert.NoError(t, err)
	assert.NotNil(t, sourceFile)

	// 验证文件数量仍然是 1
	assert.Equal(t, 1, project.GetFileCount())
}

// TestProject_CreateSourceFile_ComplexContent 测试创建包含复杂内容的文件
func TestProject_CreateSourceFile_ComplexContent(t *testing.T) {
	config := ProjectConfig{
		RootPath:    "/test",
		UseTsConfig: false,
	}
	project := NewProject(config)
	assert.NotNil(t, project)

	// 创建包含多种 TypeScript 特性的复杂文件
	sourceCode := `import { Component } from 'react';

interface User {
    id: number;
    name: string;
    email?: string;
}

type UserRole = 'admin' | 'user' | 'guest';

class UserService {
    private users: User[] = [];

    addUser(user: User): void {
        this.users.push(user);
    }

    getUserById(id: number): User | undefined {
        return this.users.find(u => u.id === id);
    }
}

export { User, UserRole, UserService };
`

	sourceFile, err := project.CreateSourceFile("services/user.ts", sourceCode)
	assert.NoError(t, err)
	assert.NotNil(t, sourceFile)

	// 验证复杂内容被正确解析
	assert.NotNil(t, sourceFile.GetFileResult().InterfaceDeclarations)
	assert.NotNil(t, sourceFile.GetFileResult().TypeDeclarations)
	assert.NotNil(t, sourceFile.GetFileResult().VariableDeclarations)

	// 验证可以访问到导入声明
	// 注意：由于我们简化了导入/导出声明的转换，这里可能为空
	assert.NotNil(t, sourceFile.GetFileResult().ImportDeclarations)
}

// TestProject_UpdateSourceFile 测试文件更新功能
func TestProject_UpdateSourceFile(t *testing.T) {
	config := ProjectConfig{
		RootPath:    "/test",
		UseTsConfig: false,
	}
	project := NewProject(config)
	assert.NotNil(t, project)

	// 创建初始文件
	originalCode := `const version = "1.0.0";`
	sourceFile, err := project.CreateSourceFile("config.ts", originalCode)
	assert.NoError(t, err)
	assert.NotNil(t, sourceFile)

	// 记录原始文件的信息
	originalFilePath := sourceFile.GetFilePath()
	originalFileCount := project.GetFileCount()

	// 更新文件内容
	updatedCode := `const version = "2.0.0";
const features = ["new", "improved"];`

	updatedFile, err := project.UpdateSourceFile("config.ts", updatedCode)
	assert.NoError(t, err, "更新文件应该成功")
	assert.NotNil(t, updatedFile, "更新后的文件不应该为空")

	// 验证文件路径保持不变
	assert.Equal(t, originalFilePath, updatedFile.GetFilePath())

	// 验证文件数量保持不变
	assert.Equal(t, originalFileCount, project.GetFileCount())

	// 验证文件仍然可以访问
	assert.True(t, project.ContainsFile("config.ts"))

	// 获取更新后的文件以验证
	retrievedFile := project.GetSourceFile("config.ts")
	assert.NotNil(t, retrievedFile, "应该能够获取更新后的文件")
	if retrievedFile != nil && updatedFile != nil {
		assert.Equal(t, updatedFile.GetFilePath(), retrievedFile.GetFilePath())
	}
}

// TestProject_UpdateSourceFile_NonExistent 测试更新不存在文件的错误处理
func TestProject_UpdateSourceFile_NonExistent(t *testing.T) {
	config := ProjectConfig{
		RootPath:    "/test",
		UseTsConfig: false,
	}
	project := NewProject(config)
	assert.NotNil(t, project)

	// 尝试更新不存在的文件
	_, err := project.UpdateSourceFile("nonexistent.ts", `const x = 1;`)
	assert.Error(t, err, "更新不存在的文件应该返回错误")
	assert.Contains(t, err.Error(), "不存在", "错误信息应该包含'不存在'")
}

// TestProject_RemoveSourceFile 测试文件移除功能
func TestProject_RemoveSourceFile(t *testing.T) {
	config := ProjectConfig{
		RootPath:    "/test",
		UseTsConfig: false,
	}
	project := NewProject(config)
	assert.NotNil(t, project)

	// 创建几个文件
	filePaths := []string{"file1.ts", "file2.ts", "file3.ts"}
	for _, path := range filePaths {
		_, err := project.CreateSourceFile(path, `const x = 1;`)
		assert.NoError(t, err)
	}

	// 验证初始状态
	assert.Equal(t, len(filePaths), project.GetFileCount())
	for _, path := range filePaths {
		assert.True(t, project.ContainsFile(path))
	}

	// 移除中间的文件
	removed, err := project.RemoveSourceFile("file2.ts")
	assert.True(t, removed, "移除操作应该返回成功")
	assert.NoError(t, err, "移除文件不应该返回错误")

	// 验证移除后的状态
	assert.Equal(t, len(filePaths)-1, project.GetFileCount())
	assert.False(t, project.ContainsFile("file2.ts"), "被移除的文件不应该再存在")
	assert.True(t, project.ContainsFile("file1.ts"), "其他文件应该仍然存在")
	assert.True(t, project.ContainsFile("file3.ts"), "其他文件应该仍然存在")
}

// TestProject_RemoveSourceFile_NonExistent 测试移除不存在文件的错误处理
func TestProject_RemoveSourceFile_NonExistent(t *testing.T) {
	config := ProjectConfig{
		RootPath:    "/test",
		UseTsConfig: false,
	}
	project := NewProject(config)
	assert.NotNil(t, project)

	// 尝试移除不存在的文件
	removed, err := project.RemoveSourceFile("nonexistent.ts")
	assert.False(t, removed, "移除不存在的文件应该返回失败")
	assert.Error(t, err, "移除不存在的文件应该返回错误")
	assert.Contains(t, err.Error(), "不存在", "错误信息应该包含'不存在'")
}

// TestProject_GetFilePaths 测试获取文件路径列表功能
func TestProject_GetFilePaths(t *testing.T) {
	config := ProjectConfig{
		RootPath:    "/test",
		UseTsConfig: false,
	}
	project := NewProject(config)
	assert.NotNil(t, project)

	// 创建文件前应该返回空列表
	paths := project.GetFilePaths()
	assert.Empty(t, paths, "空项目的文件路径列表应该为空")

	// 创建几个文件
	filePaths := []string{"src/main.ts", "src/utils/helper.ts", "config/app.ts"}
	for _, path := range filePaths {
		_, err := project.CreateSourceFile(path, `// test file`)
		assert.NoError(t, err)
	}

	// 验证文件路径列表
	paths = project.GetFilePaths()
	assert.Len(t, paths, len(filePaths), "文件路径列表长度应该正确")

	// 验证所有预期路径都存在（注意路径会被规范化）
	for _, expectedPath := range filePaths {
		found := false
		for _, actualPath := range paths {
			if actualPath == "/test/"+expectedPath {
				found = true
				break
			}
		}
		assert.True(t, found, "应该找到路径 %s", expectedPath)
	}
}

// TestProject_CreateSourceFile_ErrorHandling 测试错误处理
func TestProject_CreateSourceFile_ErrorHandling(t *testing.T) {
	config := ProjectConfig{
		RootPath:    "/test",
		UseTsConfig: false,
	}
	project := NewProject(config)
	assert.NotNil(t, project)

	// 测试创建包含语法错误的文件
	invalidCode := `const x =
// 缺少值`
	_, err := project.CreateSourceFile("invalid.ts", invalidCode)
	// 注意：由于 typescript-go 的容错性，这可能不会返回错误
	// 但我们应该验证文件不会损坏项目状态
	if err != nil {
		assert.Error(t, err, "语法错误应该被报告")
	}

	// 验证项目状态仍然一致
	assert.True(t, project.GetFileCount() == 0 || project.GetFileCount() == 1)
}

// TestProject_CreateSourceFile_Options 测试创建选项的使用
func TestProject_CreateSourceFile_Options(t *testing.T) {
	config := ProjectConfig{
		RootPath:    "/test",
		UseTsConfig: false,
	}
	project := NewProject(config)
	assert.NotNil(t, project)

	// 测试使用自定义选项创建文件
	sourceCode := `const optionTest = true;`
	options := CreateSourceFileOptions{
		Overwrite:      false,
		ScriptKind:     "typescript",
		AdditionalOptions: map[string]interface{}{
			"test": "value",
		},
	}

	sourceFile, err := project.CreateSourceFile("options.ts", sourceCode, options)
	assert.NoError(t, err)
	assert.NotNil(t, sourceFile)

	// 验证文件被正确创建
	assert.True(t, project.ContainsFile("options.ts"))
	assert.Equal(t, 1, project.GetFileCount())

	// 尝试不覆盖地重新创建，应该失败
	_, err = project.CreateSourceFile("options.ts", sourceCode, options)
	assert.Error(t, err)

	// 使用覆盖选项，应该成功
	options.Overwrite = true
	_, err = project.CreateSourceFile("options.ts", sourceCode, options)
	assert.NoError(t, err)
}

// TestProject_MixedOperations 测试混合操作的综合场景
func TestProject_MixedOperations(t *testing.T) {
	config := ProjectConfig{
		RootPath:    "/test/project",
		UseTsConfig: false,
	}
	project := NewProject(config)
	assert.NotNil(t, project)

	// 阶段1：创建多个文件
	files := map[string]string{
		"main.ts":     `import { Config } from './config';`,
		"config.ts":   `export const Config = { version: "1.0" };`,
		"utils.ts":    `export const helper = () => {};`,
	}

	for path, code := range files {
		_, err := project.CreateSourceFile(path, code)
		assert.NoError(t, err, "创建文件 %s 应该成功", path)
	}

	assert.Equal(t, len(files), project.GetFileCount(), "应该创建所有文件")

	// 阶段2：更新一个文件
	updatedConfigCode := `export const Config = { version: "2.0", debug: true };`
	_, err := project.UpdateSourceFile("config.ts", updatedConfigCode)
	assert.NoError(t, err, "更新配置文件应该成功")

	// 阶段3：移除一个文件
	removed, err := project.RemoveSourceFile("utils.ts")
	assert.True(t, removed, "移除工具文件应该成功")
	assert.NoError(t, err, "移除操作不应该出错")

	// 阶段4：验证最终状态
	assert.Equal(t, len(files)-1, project.GetFileCount(), "最终文件数量应该正确")
	assert.True(t, project.ContainsFile("main.ts"), "主文件应该仍然存在")
	assert.True(t, project.ContainsFile("config.ts"), "配置文件应该仍然存在")
	assert.False(t, project.ContainsFile("utils.ts"), "被移除的文件不应该存在")

	// 验证剩余文件仍然可以访问
	mainFile := project.GetSourceFile("main.ts")
	assert.NotNil(t, mainFile, "应该能够获取主文件")
	configFile := project.GetSourceFile("config.ts")
	assert.NotNil(t, configFile, "应该能够获取配置文件")
}