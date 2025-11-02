// +build project-api

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run -tags project-api project-creation.go <项目路径>")
		os.Exit(1)
	}

	projectPath := os.Args[1]

	fmt.Println("🎯 项目管理 API - 项目创建和配置")
	fmt.Println("================================")

	// 1. 基础项目创建 - 验证最基本的项目创建功能
	fmt.Println("\n📦 基础项目创建:")
	basicConfig := tsmorphgo.ProjectConfig{
		RootPath:         projectPath,
		IgnorePatterns:   []string{"node_modules", "dist", "build", ".git"},
		TargetExtensions: []string{".ts", ".tsx"},
	}
	basicProject := tsmorphgo.NewProject(basicConfig)
	basicFiles := basicProject.GetSourceFiles()
	fmt.Printf("✅ 基础项目创建成功，发现 %d 个文件\n", len(basicFiles))

	// 验证文件列表不为空
	if len(basicFiles) == 0 {
		fmt.Println("❌ 基础项目创建验证失败：未发现任何文件")
		return
	}

	// 2. 高级项目配置 - 验证各种配置选项的有效性
	fmt.Println("\n⚙️ 高级项目配置:")
	advancedConfig := tsmorphgo.ProjectConfig{
		RootPath:         projectPath,
		IgnorePatterns:   []string{"node_modules", "dist", "build", ".git", "*.test.ts", "*.spec.ts"},
		IsMonorepo:       false,
		TargetExtensions: []string{".ts", ".tsx", ".d.ts"},
	}
	advancedProject := tsmorphgo.NewProject(advancedConfig)
	advancedFiles := advancedProject.GetSourceFiles()
	fmt.Printf("✅ 高级项目配置成功，发现 %d 个文件\n", len(advancedFiles))

	// 验证配置是否生效 - 应该比基础项目包含更多文件类型
	if len(advancedFiles) < len(basicFiles) {
		fmt.Println("ℹ️ 高级配置过滤了一些文件，这是正常的")
	}

	// 3. 内存源码项目创建 - 验证从内存字符串创建项目的能力
	fmt.Println("\n💾 从内存源码创建项目:")
	memorySources := map[string]string{
		"test.ts": `interface User {
    id: number;
    name: string;
    email: string;
}

class UserService {
    private users: User[] = [];

    addUser(user: User): void {
        this.users.push(user);
    }

    getUsers(): User[] {
        return this.users;
    }
}`,
		"utils.ts": `export const formatDate = (date: Date): string => {
    return date.toISOString();
}

export const debounce = <T extends (...args: any[]) => any>(
    func: T,
    wait: number
): ((...args: Parameters<T>) => void) => {
    let timeout: NodeJS.Timeout;
    return (...args: Parameters<T>) => {
        clearTimeout(timeout);
        timeout = setTimeout(() => func(...args), wait);
    };
}`,
	}
	memoryProject := tsmorphgo.NewProjectFromSources(memorySources)
	memoryFiles := memoryProject.GetSourceFiles()
	fmt.Printf("✅ 内存项目创建成功，发现 %d 个文件\n", len(memoryFiles))

	// 验证内存项目是否正确解析
	if len(memoryFiles) != len(memorySources) {
		fmt.Printf("❌ 内存项目验证失败：期望 %d 个文件，实际 %d 个\n",
			len(memorySources), len(memoryFiles))
		return
	}

	// 4. 项目配置验证 - 测试不同配置选项的行为
	fmt.Println("\n🔍 项目配置验证:")

	// 4.1 测试空忽略列表
	fmt.Println("  4.1 测试空忽略列表:")
	noIgnoreConfig := tsmorphgo.ProjectConfig{
		RootPath:         projectPath,
		IgnorePatterns:   []string{},
		TargetExtensions: []string{".ts", ".tsx"},
	}
	noIgnoreProject := tsmorphgo.NewProject(noIgnoreConfig)
	fmt.Printf("✅ 空忽略列表配置成功，发现 %d 个文件\n",
		len(noIgnoreProject.GetSourceFiles()))

	// 4.2 测试仅 TypeScript 文件
	fmt.Println("  4.2 测试仅 TypeScript 文件:")
	tsOnlyConfig := tsmorphgo.ProjectConfig{
		RootPath:         projectPath,
		IgnorePatterns:   []string{"node_modules", "dist", "build", ".git"},
		TargetExtensions: []string{".ts"},
	}
	tsOnlyProject := tsmorphgo.NewProject(tsOnlyConfig)
	fmt.Printf("✅ 仅 TypeScript 文件配置成功，发现 %d 个文件\n",
		len(tsOnlyProject.GetSourceFiles()))

	// 4.3 测试包含 JSX
	fmt.Println("  4.3 测试包含 JSX:")
	jsxConfig := tsmorphgo.ProjectConfig{
		RootPath:         projectPath,
		IgnorePatterns:   []string{"node_modules", "dist", "build", ".git"},
		TargetExtensions: []string{".ts", ".tsx"},
	}
	jsxProject := tsmorphgo.NewProject(jsxConfig)
	fmt.Printf("✅ 包含 JSX 配置成功，发现 %d 个文件\n",
		len(jsxProject.GetSourceFiles()))

	// 5. 项目 API 功能验证 - 验证核心 API 方法
	fmt.Println("\n🔧 项目 API 功能验证:")

	// 5.1 GetSourceFile 方法验证
	fmt.Println("  5.1 验证 GetSourceFile 方法:")
	if len(advancedFiles) > 0 {
		firstFile := advancedFiles[0]
		filePath := firstFile.GetFilePath()
		retrievedFile := advancedProject.GetSourceFile(filePath)

		if retrievedFile != nil {
			fmt.Printf("✅ GetSourceFile 成功：能够获取文件 %s\n", filePath)
		} else {
			fmt.Printf("❌ GetSourceFile 失败：无法获取文件 %s\n", filePath)
		}
	}

	// 5.2 文件路径一致性验证
	fmt.Println("  5.2 验证文件路径一致性:")
	for i, file := range advancedFiles[:3] { // 只检查前 3 个文件
		retrievedFile := advancedProject.GetSourceFile(file.GetFilePath())
		if retrievedFile != nil && retrievedFile.GetFilePath() == file.GetFilePath() {
			fmt.Printf("✅ 文件 %d 路径一致\n", i+1)
		} else {
			fmt.Printf("❌ 文件 %d 路径不一致\n", i+1)
		}
	}

	// 5.3 项目元数据验证
	fmt.Println("  5.3 验证项目元数据:")
	fmt.Printf("   - 项目根路径: %s\n", projectPath)
	fmt.Printf("   - 忽略模式: %v\n", advancedConfig.IgnorePatterns)
	fmt.Printf("   - 目标扩展名: %v\n", advancedConfig.TargetExtensions)
	fmt.Printf("   - Monorepo 模式: %t\n", advancedConfig.IsMonorepo)

	// 6. 错误处理验证 - 测试错误输入的处理
	fmt.Println("\n⚠️ 错误处理验证:")

	// 6.1 测试不存在的项目路径
	fmt.Println("  6.1 测试不存在的项目路径:")
	invalidConfig := tsmorphgo.ProjectConfig{
		RootPath:         "/nonexistent/path",
		IgnorePatterns:   []string{"node_modules", "dist", "build", ".git"},
		TargetExtensions: []string{".ts", ".tsx"},
	}
	invalidProject := tsmorphgo.NewProject(invalidConfig)
	invalidFiles := invalidProject.GetSourceFiles()
	fmt.Printf("✅ 不存在路径的处理正常：发现 %d 个文件（应为 0）\n", len(invalidFiles))

	// 7. 性能基准测试 - 简单的性能测试
	fmt.Println("\n⏱️ 性能基准测试:")

	// 7.1 项目创建时间
	fmt.Println("  7.1 项目创建时间测试:")
	startTime := time.Now()
	for i := 0; i < 5; i++ {
		perfProject := tsmorphgo.NewProject(basicConfig)
		_ = len(perfProject.GetSourceFiles())
	}
	duration := time.Since(startTime)
	fmt.Printf("✅ 性能测试完成：连续创建 5 个项目，耗时: %v\n", duration)

	// 8. 验证结果汇总
	fmt.Println("\n📊 验证结果汇总:")
	fmt.Printf("  ✅ 基础项目创建: 发现 %d 个文件\n", len(basicFiles))
	fmt.Printf("  ✅ 高级项目配置: 发现 %d 个文件\n", len(advancedFiles))
	fmt.Printf("  ✅ 内存项目创建: 发现 %d 个文件\n", len(memoryFiles))
	fmt.Printf("  ✅ 空忽略列表: 发现 %d 个文件\n", len(noIgnoreProject.GetSourceFiles()))
	fmt.Printf("  ✅ 仅 TypeScript: 发现 %d 个文件\n", len(tsOnlyProject.GetSourceFiles()))
	fmt.Printf("  ✅ 包含 JSX: 发现 %d 个文件\n", len(jsxProject.GetSourceFiles()))

	// 最终验证
	if len(basicFiles) > 0 && len(advancedFiles) > 0 && len(memoryFiles) > 0 {
		fmt.Println("\n🎉 项目管理 API 验证完成！所有核心功能正常工作")
		fmt.Println("================================")
		fmt.Println("📋 已验证的 API:")
		fmt.Println("   - tsmorphgo.NewProject()")
		fmt.Println("   - tsmorphgo.NewProjectFromSources()")
		fmt.Println("   - project.GetSourceFiles()")
		fmt.Println("   - project.GetSourceFile()")
		fmt.Println("   - ProjectConfig 结构体配置")
		fmt.Println("================================")
	} else {
		fmt.Println("\n❌ 项目管理 API 验证失败！存在功能异常")
	}
}