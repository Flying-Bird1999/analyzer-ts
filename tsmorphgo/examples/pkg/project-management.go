//go:build project_management
// +build project_management

package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🏗️ TSMorphGo 项目管理 - 新API演示")
	fmt.Println("=" + strings.Repeat("=", 50))

	// =============================================================================
	// 本文件演示新的统一API在项目管理中的应用
	// =============================================================================
	// 学习级别: 初级 → 高级
	// 预计时间: 15-20分钟
	//
	// 新API的优势:
	// - 统一的接口设计，更简洁的方法调用
	// - 支持内存文件系统，便于测试和原型开发
	// - 支持动态文件创建和修改
	// - 更好的资源管理
	//
	// 新API功能:
	// - NewProjectFromSources() → 内存项目创建
	// - project.CreateSourceFile() → 动态文件创建
	// - project.GetSourceFiles() → 获取所有源文件
	// - project.Close() → 资源清理
	// =============================================================================

	// 示例1: 内存文件系统项目 (基础)
	fmt.Println("\n🧠 示例1: 内存文件系统项目 (基础)")
	fmt.Println("展示如何创建和管理内存中的TypeScript项目")

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

	// 验证内存项目
	memFiles := memoryProject.GetSourceFiles()
	fmt.Printf("✅ 内存项目创建成功！\n")
	fmt.Printf("📊 内存项目统计:\n")
	fmt.Printf("  - 文件数量: %d\n", len(memFiles))

	for _, file := range memFiles {
		fileName := extractFileName(file.GetFilePath())
		lineCount := strings.Count(file.GetFileResult().Raw, "\n") + 1
		fmt.Printf("  - %s (%d行)\n", fileName, lineCount)
	}

	// 示例2: 动态文件管理 (高级)
	fmt.Println("\n➕ 示例2: 动态文件管理 (高级)")
	fmt.Println("展示如何动态创建和管理项目文件")

	// 在内存项目中动态创建配置文件
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
		logLevel: "info"
	}
};

// 导出配置类型
export type AppConfig = typeof APP_CONFIG;
`

	// 动态创建文件到内存项目中
	configFile, err := memoryProject.CreateSourceFile(
		"/src/config/app-config.ts",
		configContent,
		tsmorphgo.CreateSourceFileOptions{Overwrite: true},
	)
	if err != nil {
		fmt.Printf("❌ 创建配置文件失败: %v\n", err)
	} else {
		fmt.Printf("✅ 配置文件创建成功: %s\n", configFile.GetFilePath())
		lineCount := strings.Count(configFile.GetFileResult().Raw, "\n") + 1
		fmt.Printf("  - 文件行数: %d\n", lineCount)
	}

	// 验证文件已创建
	updatedFiles := memoryProject.GetSourceFiles()
	fmt.Printf("📊 更新后项目统计: %d 个文件\n", len(updatedFiles))

	// 示例3: 项目分析和统计
	fmt.Println("\n📊 示例3: 项目分析和统计")
	fmt.Println("展示如何分析项目结构和统计信息")

	// 分析所有文件
	var totalLines = 0
	var totalNodes = 0
	var fileStats = make(map[string]int)

	for _, file := range updatedFiles {
		filePath := file.GetFilePath()
		content := file.GetFileResult().Raw
		lineCount := strings.Count(content, "\n") + 1
		totalLines += lineCount

		// 按目录分类统计
		dir := extractDirectory(filePath)
		fileStats[dir]++

		// 统计节点数量
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			totalNodes++
		})
	}

	fmt.Printf("\n📈 项目统计:\n")
	fmt.Printf("  - 总文件数: %d\n", len(updatedFiles))
	fmt.Printf("  - 总行数: %d\n", totalLines)
	fmt.Printf("  - 总节点数: %d\n", totalNodes)

	fmt.Printf("\n📁 目录统计:\n")
	for dir, count := range fileStats {
		fmt.Printf("  - %s: %d 个文件\n", dir, count)
	}

	// 示例4: 节点类型分析
	fmt.Println("\n🔍 示例4: 节点类型分析")
	fmt.Println("展示如何分析项目中的节点类型分布")

	var nodeTypeStats = make(map[tsmorphgo.SyntaxKind]int)

	for _, file := range updatedFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			kind := node.GetKind()
			nodeTypeStats[kind]++
		})
	}

	fmt.Printf("\n🏷️ 节点类型分布:\n")
	// 显示最常见的10种节点类型
	count := 0
	for kind, num := range nodeTypeStats {
		if count >= 10 {
			break
		}
		fmt.Printf("  - %s: %d 个\n", kind.String(), num)
		count++
	}

	// 示例5: 声用和引用分析
	fmt.Println("\n🔗 示例5: 调用和引用分析")
	fmt.Println("展示如何分析函数调用和引用关系")

	var callExpressions = 0
	var importStatements = 0
	var exportStatements = 0

	for _, file := range updatedFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsCallExpr() {
				callExpressions++
			}
			if node.IsImportDeclaration() {
				importStatements++
			}
			if node.IsKind(tsmorphgo.KindExportDeclaration) {
				exportStatements++
			}
		})
	}

	fmt.Printf("\n📞 调用和引用统计:\n")
	fmt.Printf("  - 函数调用: %d\n", callExpressions)
	fmt.Printf("  - 导入语句: %d\n", importStatements)
	fmt.Printf("  - 导出语句: %d\n", exportStatements)

	// 清理资源
	memoryProject.Close()
	fmt.Printf("✅ 内存项目资源已清理\n")

	// =============================================================================
	// 示例6: 分析真实项目 (demo-react-app)
	// =============================================================================
	fmt.Println("\n\n" + strings.Repeat("=", 50))
	fmt.Println("🚀 示例6: 分析真实前端项目 (demo-react-app)")
	fmt.Println("展示如何使用 NewProject 加载和分析一个真实的文件系统项目")
	analyzeRealProject()

	fmt.Println("\n\n" + strings.Repeat("=", 50))
	fmt.Println("\n🎯 项目管理使用总结:")
	fmt.Println("1. 内存项目 → 使用 NewProjectFromSources() 创建，用于测试和原型开发")
	fmt.Println("2. 真实项目 → 使用 NewProject() 加载，用于分析实际代码库")
	fmt.Println("3. 文件管理 → 使用 CreateSourceFile() 动态创建文件")
	fmt.Println("4. 项目分析 → 使用 GetSourceFiles() 和 ForEachDescendant() 遍历")
	fmt.Println("5. 资源管理 → 始终使用 defer project.Close() 清理资源")

	fmt.Println("\n✅ 项目管理示例完成!")
	fmt.Println("新API让项目管理变得更加简单和高效！")
}

// analyzeRealProject 分析一个真实的文件系统项目
func analyzeRealProject() {
	// 获取 demo-react-app 的绝对路径
	realProjectPath, err := filepath.Abs("../demo-react-app")
	if err != nil {
		log.Fatalf("无法解析项目路径: %v", err)
	}

	// 使用 NewProject 加载真实项目
	realProject := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:         realProjectPath,
		TargetExtensions: []string{".ts", ".tsx"},
		UseTsConfig:      true, // 使用项目中的 tsconfig.json
	})
	defer realProject.Close()

	allFiles := realProject.GetSourceFiles()
	fmt.Printf("✅ 真实项目加载成功！\n")
	fmt.Printf("📊 项目统计:\n")
	fmt.Printf("  - 总文件数: %d\n", len(allFiles))

	// 分析组件目录
	fmt.Println("\n🔍 分析 'src/components' 目录:")
	var components []string
	var interfaceCount = 0
	for _, file := range allFiles {
		// 查找组件文件
		if strings.Contains(file.GetFilePath(), "/src/components/") && strings.HasSuffix(file.GetFilePath(), ".tsx") {
			components = append(components, extractFileName(file.GetFilePath()))
		}

		// 统计项目中的接口数量
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsInterfaceDeclaration() {
				interfaceCount++
			}
		})
	}

	if len(components) > 0 {
		fmt.Printf("  - 找到 %d 个组件:\n", len(components))
		for _, component := range components {
			fmt.Printf("    - %s\n", component)
		}
	} else {
		fmt.Println("  - 未在 'src/components' 目录中找到组件文件。")
	}

	fmt.Printf("\n🏷️  项目中总共找到 %d 个 'interface' 声明。\n", interfaceCount)
}

// 辅助函数

// extractFileName 提取文件名
func extractFileName(filePath string) string {
	parts := strings.Split(filePath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return filePath
}

// extractDirectory 提取目录路径
func extractDirectory(filePath string) string {
	parts := strings.Split(filePath, "/")
	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], "/")
	}
	return "/"
}
