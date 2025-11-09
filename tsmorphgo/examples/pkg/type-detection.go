//go:build type_detection
// +build type_detection

package main

import (
	"fmt"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🏷️ TSMorphGo 类型检测 - 新API演示")
	fmt.Println("=" + strings.Repeat("=", 50))

	// =============================================================================
	// 本文件演示新的统一API在类型检测中的应用
	// =============================================================================
	// 学习级别: 初级 → 高级
	// 预计时间: 15-20分钟
	//
	// 新API的优势:
	// - 统一的接口设计，无需记忆大量函数名
	// - 支持类别检查，可以批量判断节点类型
	// - 更简洁的方法调用
	// - 类型安全的转换接口
	//
	// 新API:
	// - node.IsInterfaceDeclaration() → 接口声明检查
	// - node.IsTypeAliasDeclaration() → 类型别名检查
	// - node.IsFunctionDeclaration() → 函数声明检查
	// - node.IsCallExpr() → 函数调用检查
	// - node.IsDeclaration() → 任何声明检查
	// - node.IsType() → 任何类型检查
	// =============================================================================

	// 使用内存项目进行演示，不依赖外部文件
	project := tsmorphgo.NewProjectFromSources(map[string]string{
		"/types/user.ts": `
			// 用户接口定义
			export interface User {
				id: number;
				name: string;
				email?: string;
				avatar?: string;
			}

			// 用户状态枚举
			export enum UserStatus {
				Active = 'active',
				Inactive = 'inactive',
				Suspended = 'suspended'
			}

			// 用户类型别名
			export type UserRole = 'admin' | 'user' | 'guest';
			export type UserID = number;

			// 响应类型
			export interface ApiResponse<T> {
				data: T;
				status: number;
				message: string;
			}

			// 用户响应类型
			export type UserResponse = ApiResponse<User>;
		`,
		"/services/user-service.ts": `
			import { User, UserStatus, UserRole, UserID } from '../types/user';

			// 用户服务类
			export class UserService {
				private users: Map<UserID, User> = new Map();

				// 创建用户
				create(userData: Omit<User, 'id'>): User {
					const user: User = {
						id: this.generateId(),
						...userData
					};
					this.users.set(user.id, user);
					return user;
				}

				// 查找用户
				findById(id: UserID): User | undefined {
					return this.users.get(id);
				}

				// 获取所有用户
				findAll(): User[] {
					return Array.from(this.users.values());
				}

				// 更新用户状态
				updateStatus(id: UserID, status: UserStatus): boolean {
					const user = this.users.get(id);
					if (user) {
						user.email = user.email || ''; // 确保email字段存在
						this.users.set(id, user);
						return true;
					}
					return false;
				}

				// 根据角色筛选用户
				findByRole(role: UserRole): User[] {
					return this.findAll().filter(user => {
						// 模拟角色检查逻辑
						return role === 'admin' || role === 'user';
					});
				}

				private generateId(): UserID {
					return Math.floor(Math.random() * 10000);
				}
			}

			// 工厂函数
			export function createUserAdmin(name: string): User {
				return {
					id: 0,
					name,
					email: '',
					status: UserStatus.Active
				};
			}
		`,
		"/app/main.ts": `
			import { UserService, createUserAdmin } from '../services/user-service';
			import { User, UserResponse } from '../types/user';

			// 应用主类
			class Application {
				private userService: UserService;

				constructor() {
					this.userService = new UserService();
				}

				// 初始化应用
				async initialize(): Promise<void> {
					console.log('应用初始化中...');

					// 创建管理员用户
					const admin = createUserAdmin('Admin User');
					this.userService.create(admin);

					// 创建普通用户
					const normalUser: User = {
						id: 1,
						name: 'Normal User',
						email: 'user@example.com',
						status: 'active'
					};
					this.userService.create(normalUser);

					console.log('应用初始化完成！');
				}

				// 获取用户统计
				getUserStats(): { total: number; active: number } {
					const users = this.userService.findAll();
					const active = users.filter(u => u.status === 'active').length;
					return {
						total: users.length,
						active
					};
				}
			}

			// 应用入口
			const app = new Application();
			app.initialize().then(() => {
				console.log('应用启动成功！');
			});
		`,
	})

	defer project.Close()

	// 示例1: 基础类型检测
	fmt.Println("\n🔍 示例1: 基础类型检测")
	fmt.Println("展示如何使用新API进行基础类型检测")

	typesFile := project.GetSourceFile("/types/user.ts")
	if typesFile == nil {
		fmt.Println("❌ 未找到 types/user.ts 文件")
		return
	}

	var (
		interfaces = 0
		enums = 0
		typeAliases = 0
	)

	typesFile.ForEachDescendant(func(node tsmorphgo.Node) {
		switch {
		case node.IsInterfaceDeclaration():
			interfaces++
			if name, ok := node.GetNodeName(); ok {
				fmt.Printf("  🎭 发现接口: %s\n", name)
			}
		case node.IsKind(tsmorphgo.KindEnumDeclaration):
			enums++
			if name, ok := node.GetNodeName(); ok {
				fmt.Printf("  🔢 发现枚举: %s\n", name)
			}
		case node.IsKind(tsmorphgo.KindTypeAliasDeclaration):
			typeAliases++
			if name, ok := node.GetNodeName(); ok {
				fmt.Printf("  📝 发现类型别名: %s\n", name)
			}
		}
	})

	fmt.Printf("\n📊 类型统计:\n")
	fmt.Printf("  - 接口声明: %d\n", interfaces)
	fmt.Printf("  - 枚举声明: %d\n", enums)
	fmt.Printf("  - 类型别名: %d\n", typeAliases)

	// 示例2: 类别检测
	fmt.Println("\n🎯 示例2: 类别检测")
	fmt.Println("展示如何使用类别检查进行批量检测")

	serviceFile := project.GetSourceFile("/services/user-service.ts")
	if serviceFile == nil {
		fmt.Println("❌ 未找到 services/user-service.ts 文件")
		return
	}

	var (
		declarations = 0
		expressions = 0
		statements = 0
		types = 0
		modules = 0
	)

	serviceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsDeclaration() {
			declarations++
		}
		if node.IsExpression() {
			expressions++
		}
		if node.IsStatement() {
			statements++
		}
		if node.IsType() {
			types++
		}
		if node.IsModule() {
			modules++
		}
	})

	fmt.Printf("\n📊 类别统计:\n")
	fmt.Printf("  - 声明类节点: %d\n", declarations)
	fmt.Printf("  - 表达式类节点: %d\n", expressions)
	fmt.Printf("  - 语句类节点: %d\n", statements)
	fmt.Printf("  - 类型类节点: %d\n", types)
	fmt.Printf("  - 模块类节点: %d\n", modules)

	// 示例3: 多类型检查
	fmt.Println("\n🔬 示例3: 多类型检查")
	fmt.Println("展示如何一次检查多种类型")

	appFile := project.GetSourceFile("/app/main.ts")
	if appFile == nil {
		fmt.Println("❌ 未找到 app/main.ts 文件")
		return
	}

	var classCount = 0
	var variableOrFunctionCount = 0

	// 检查类声明
	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsClassDeclaration() {
			classCount++
			if name, ok := node.GetNodeName(); ok {
				fmt.Printf("  🏗️ 发现类: %s\n", name)
			}
		}
	})

	// 检查变量或函数声明
	declarationKinds := []tsmorphgo.SyntaxKind{
		tsmorphgo.KindVariableDeclaration,
		tsmorphgo.KindFunctionDeclaration,
	}

	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsAnyKind(declarationKinds...) {
			variableOrFunctionCount++
			if name, ok := node.GetNodeName(); ok {
				fmt.Printf("  📦 发现声明: %s\n", name)
			}
		}
	})

	fmt.Printf("\n📊 多类型统计:\n")
	fmt.Printf("  - 类声明: %d\n", classCount)
	fmt.Printf("  - 变量或函数声明: %d\n", variableOrFunctionCount)

	// 示例4: 精确类型检查
	fmt.Println("\n⚡ 示例4: 精确类型检查")
	fmt.Println("展示如何使用精确的节点类型检查")

	var callExpressions = 0
	var propertyAccess = 0
	var binaryExpressions = 0

	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsCallExpr() {
			callExpressions++
			text := node.GetText()
			if len(text) > 30 {
				text = text[:30] + "..."
			}
			fmt.Printf("  📞 函数调用: %s\n", text)
		}
		if node.IsPropertyAccessExpression() {
			propertyAccess++
			fmt.Printf("  🔗 属性访问: %s\n", strings.TrimSpace(node.GetText()))
		}
		if node.IsKind(tsmorphgo.KindBinaryExpression) {
			binaryExpressions++
			fmt.Printf("  ➕ 二元表达式: %s\n", strings.TrimSpace(node.GetText()))
		}
	})

	fmt.Printf("\n📊 精确类型统计:\n")
	fmt.Printf("  - 函数调用表达式: %d\n", callExpressions)
	fmt.Printf("  - 属性访问表达式: %d\n", propertyAccess)
	fmt.Printf("  - 二元表达式: %d\n", binaryExpressions)

	// 示例5: 类型转换
	fmt.Println("\n🔄 示例5: 类型转换")
	fmt.Println("展示如何使用类型转换API")

	var conversionSuccess = 0

	typesFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsDeclaration() {
			if result, ok := node.AsDeclaration(); ok {
				conversionSuccess++
				fmt.Printf("  ✅ 转换成功: %T\n", result)
			}
		}
	})

	fmt.Printf("\n📊 转换统计:\n")
	fmt.Printf("  - 成功转换: %d\n", conversionSuccess)

	// 示例6: 名称和值提取
	fmt.Println("\n💎 示例6: 名称和值提取")
	fmt.Println("展示如何提取节点名称和字面量值")

	var names []string
	var literals []interface{}

	serviceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		// 提取声明名称
		if node.IsDeclaration() {
			if name, ok := node.GetNodeName(); ok {
				names = append(names, name)
			}
		}

		// 提取字面量值
		if node.IsLiteral() {
			if value, ok := node.GetLiteralValue(); ok {
				literals = append(literals, value)
			}
		}
	})

	fmt.Printf("\n📊 提取统计:\n")
	fmt.Printf("  - 提取的名称: %d个\n", len(names))
	if len(names) > 0 {
		fmt.Printf("    示例: %s\n", strings.Join(names[:min(3, len(names))], ", "))
	}
	fmt.Printf("  - 提取的字面量: %d个\n", len(literals))
	if len(literals) > 0 {
		fmt.Printf("    示例: %v\n", literals[0])
	}

	fmt.Println("\n🎯 新API总结:")
	fmt.Println("1. 使用 node.IsDeclaration() 等类别方法进行批量检查")
	fmt.Println("2. 使用 node.IsKind(KindXxx) 进行精确类型检查")
	fmt.Println("3. 使用 node.IsAnyKind(...) 检查多种类型")
	fmt.Println("4. 使用 node.GetNodeName() 获取节点名称")
	fmt.Println("5. 使用 node.GetLiteralValue() 提取字面量值")
	fmt.Println("6. 使用 node.AsDeclaration() 进行类型转换")

	fmt.Println("\n✅ 类型检测示例完成!")
	fmt.Println("新API大大简化了类型检测的复杂度！")
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}