//go:build project_management
// +build project_management

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🏗️ TSMorphGo 项目管理 - 正确使用姿势")
	fmt.Println("=" + repeat("=", 50))

	// =============================================================================
	// 本文件演示 TSMorphGo 项目管理的正确使用方法
	// =============================================================================
	// 学习级别: 初级 → 高级
	// 预计时间: 30-45分钟
	//
	// 功能覆盖:
	// - 基础: 项目初始化、文件管理、tsconfig支持
	// - 高级: 内存文件系统 ⭐、动态文件创建 ⭐
	// - 应用: 测试场景、原型开发
	//
	// ⭐ = 高级功能，初学者可先跳过
	//
	// 对齐 ts-morph API:
	// - new Project({tsConfigFilePath}) → NewProject(ProjectConfig{UseTsConfig: true})
	// - new Project({useInMemoryFileSystem: true}) → NewProjectFromSources()
	// - project.createSourceFile() → project.CreateSourceFile()
	// =============================================================================

	// 示例1: 基于真实项目的初始化 (初级)
	// 对应 ts-morph: new Project({tsConfigFilePath: "path/to/tsconfig.json"})
	fmt.Println("\n📁 示例1: 基于tsconfig.json的项目初始化 (初级)")
	fmt.Println("对齐 ts-morph: new Project({tsConfigFilePath})")

	// 初始化项目，自动加载tsconfig.json配置
	realProjectPath := "/Users/bird/Desktop/alalyzer/analyzer-ts/tsmorphgo/examples/demo-react-app"
	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:         realProjectPath,
		TargetExtensions: []string{".ts", ".tsx"},
		IgnorePatterns:   []string{"node_modules", "dist", ".git", "build"},
		UseTsConfig:      true, // 对应 ts-morph 的 tsConfigFilePath 配置
	})
	defer project.Close()

	// 验证项目创建成功
	sourceFiles := project.GetSourceFiles()
	if len(sourceFiles) == 0 {
		log.Fatal("项目初始化失败：未找到任何源文件")
	}

	fmt.Printf("✅ 项目初始化成功！\n")
	fmt.Printf("📊 项目统计:\n")
	fmt.Printf("  - 项目路径: %s\n", realProjectPath)
	fmt.Printf("  - 源文件数量: %d\n", len(sourceFiles))

	// 按类型分类文件
	var types, components, utils, other int
	for _, file := range sourceFiles {
		filePath := file.GetFilePath()
		switch {
		case strings.Contains(filePath, "types"):
			types++
		case strings.Contains(filePath, "components"):
			components++
		case strings.Contains(filePath, "utils") || strings.Contains(filePath, "services"):
			utils++
		default:
			other++
		}
	}

	fmt.Printf("  - 类型文件: %d\n", types)
	fmt.Printf("  - 组件文件: %d\n", components)
	fmt.Printf("  - 工具文件: %d\n", utils)
	fmt.Printf("  - 其他文件: %d\n", other)

	// 示例2: 内存文件系统项目 (高级 ⭐)
	// 对应 ts-morph: new Project({useInMemoryFileSystem: true, skipAddingFilesFromTsConfig: true})
	fmt.Println("\n🧠 示例2: 内存文件系统项目 (高级 ⭐)")
	fmt.Println("对齐 ts-morph: new Project({useInMemoryFileSystem: true})")
	fmt.Println("应用场景: 单元测试、原型开发、代码生成")

	// 创建内存项目，完全在内存中操作，不依赖真实文件系统
	memoryProject := tsmorphgo.NewProjectFromSources(map[string]string{
		"/models/User.ts": `
			// 用户模型定义
			export interface User {
				id: number;
				name: string;
				email: string;
				avatar?: string;
				createdAt: Date;
				updatedAt: Date;
			}

			// 用户状态枚举
			export enum UserStatus {
				Active = 'active',
				Inactive = 'inactive',
				Suspended = 'suspended'
			}

			// 用户类型
			export type UserType = 'admin' | 'user' | 'guest';
		`,
		"/services/UserService.ts": `
			// 用户服务层
			import { User, UserStatus, UserType } from '../models/User';

			// 用户服务类
			export class UserService {
				private users: User[] = [];

				// 创建用户
				create(userData: Omit<User, 'id' | 'createdAt' | 'updatedAt'>): User {
					const user: User = {
						...userData,
						id: this.users.length + 1,
						createdAt: new Date(),
						updatedAt: new Date()
					};
					this.users.push(user);
					return user;
				}

				// 查找用户
				findById(id: number): User | undefined {
					return this.users.find(user => user.id === id);
				}

				// 获取所有用户
				findAll(): User[] {
					return [...this.users];
				}

				// 更新用户
				update(id: number, updates: Partial<User>): User | null {
					const userIndex = this.users.findIndex(user => user.id === id);
					if (userIndex === -1) return null;

					this.users[userIndex] = {
						...this.users[userIndex],
						...updates,
						updatedAt: new Date()
					};
					return this.users[userIndex];
				}

				// 删除用户
				delete(id: number): boolean {
					const userIndex = this.users.findIndex(user => user.id === id);
					if (userIndex === -1) return false;

					this.users.splice(userIndex, 1);
					return true;
				}
			}
		`,
		"/tests/UserService.test.ts": `
			// 用户服务测试
			import { UserService, User } from '../services/UserService';

			// 测试数据
			const testUserData = {
				name: '测试用户',
				email: 'test@example.com'
			};

			// 测试函数
			export function testUserService(): void {
				console.log('开始测试 UserService...');

				const service = new UserService();

				// 测试创建用户
				const user = service.create(testUserData);
				console.log('✅ 创建用户成功:', user.name);

				// 测试查找用户
				const foundUser = service.findById(user.id);
				console.log('✅ 查找用户成功:', foundUser?.name);

				// 测试更新用户
				const updatedUser = service.update(user.id, { name: '更新后的用户' });
				console.log('✅ 更新用户成功:', updatedUser?.name);

				// 测试删除用户
				const deleted = service.delete(user.id);
				console.log('✅ 删除用户成功:', deleted);

				// 测试查找不存在的用户
				const notFoundUser = service.findById(999);
				console.log('✅ 查找不存在用户返回:', notFoundUser);

				console.log('UserService 测试完成！');
			}
		`,
	})
	defer memoryProject.Close()

	// 验证内存项目
	memFiles := memoryProject.GetSourceFiles()
	fmt.Printf("✅ 内存项目创建成功！\n")
	fmt.Printf("📊 内存项目统计:\n")
	fmt.Printf("  - 文件数量: %d\n", len(memFiles))

	for _, file := range memFiles {
		fileName := extractFileName(file.GetFilePath())
		fmt.Printf("  - %s (%d行)\n", fileName, countLines(file))
	}

	// 示例3: 动态文件管理 (高级 ⭐)
	// 对应 ts-morph: project.createSourceFile(filePath, content)
	fmt.Println("\n➕ 示例3: 动态文件管理 (高级 ⭐)")
	fmt.Println("对齐 ts-morph: project.createSourceFile(filePath, content)")
	fmt.Println("应用场景: 配置文件生成、临时文件创建、动态内容注入")

	// 在真实项目中动态创建配置文件
	configContent := `
// 动态生成的配置文件
// 生成时间: ${new Date().toISOString()}

export const APP_CONFIG = {
	// 应用基础配置
	name: "TSMorphGo Demo App",
	version: "1.0.0",
	environment: "development",

	// API配置
	api: {
		baseUrl: "https://api.example.com",
		timeout: 10000,
		retries: 3
	},

	// 功能开关
	features: {
		userManagement: true,
		dataExport: true,
		advancedSearch: false
	},

	// 调试配置
	debug: {
		enabled: true,
		logLevel: "info",
		showPerformanceMetrics: true
	}
};

// 导出配置类型
export type AppConfig = typeof APP_CONFIG;
`

	// 动态创建文件到真实项目中
	configFile, err := project.CreateSourceFile(
		realProjectPath+"/src/config/app-config.ts",
		configContent,
		tsmorphgo.CreateSourceFileOptions{Overwrite: true},
	)
	if err != nil {
		log.Printf("❌ 创建配置文件失败: %v", err)
	} else {
		fmt.Printf("✅ 配置文件创建成功: %s\n", configFile.GetFilePath())
		fmt.Printf("  - 文件行数: %d\n", countLines(configFile))
	}

	// 验证文件已创建
	updatedFiles := project.GetSourceFiles()
	fmt.Printf("📊 更新后项目统计: %d 个文件\n", len(updatedFiles))

	// 示例4: 文件内容操作和验证 (中级)
	fmt.Println("\n📖 示例4: 文件内容操作和验证 (中级)")

	// 读取并分析特定文件
	userTypesFile := project.GetSourceFile(realProjectPath + "/src/types.ts")
	if userTypesFile != nil {
		content := userTypesFile.GetFileResult().Raw
		interfaceCount := strings.Count(content, "export interface")
		typeCount := strings.Count(content, "export type")

		fmt.Printf("📋 types.ts 文件分析:\n")
		fmt.Printf("  - 接口数量: %d\n", interfaceCount)
		fmt.Printf("  - 类型别名数量: %d\n", typeCount)
		fmt.Printf("  - 总行数: %d\n", strings.Count(content, "\n")+1)
	}

	// 示例5: 错误处理和最佳实践 (中级)
	fmt.Println("\n🛡️ 示例5: 错误处理和最佳实践 (中级)")

	// 演示错误处理
	nonExistentFile := project.GetSourceFile(realProjectPath + "/src/non-existent.ts")
	if nonExistentFile == nil {
		fmt.Printf("✅ 正确处理不存在的文件: 返回 nil\n")
	}

	// 演示安全的项目关闭
	fmt.Printf("✅ 项目资源管理: 使用 defer 确保资源正确释放\n")

	fmt.Println("\n🎯 项目管理使用姿势总结:")
	fmt.Println("1. 基础项目 → 使用 NewProject() + UseTsConfig: true")
	fmt.Println("2. 测试项目 → 使用 NewProjectFromSources() + 内存文件")
	fmt.Println("3. 动态文件 → 使用 CreateSourceFile() + Overwrite 选项")
	fmt.Println("4. 资源管理 → 始终使用 defer 关闭项目")
	fmt.Println("5. 错误处理 → 检查返回值是否为 nil")

	fmt.Println("\n✅ 项目管理示例完成!")
}

// 辅助函数：重复字符串
func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// 辅助函数：提取文件名
func extractFileName(filePath string) string {
	parts := strings.Split(filePath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return filePath
}

// 辅助函数：统计文件行数
func countLines(file *tsmorphgo.SourceFile) int {
	if fileResult := file.GetFileResult(); fileResult != nil && fileResult.Raw != "" {
		return len(strings.Split(fileResult.Raw, "\n"))
	}
	return 0
}