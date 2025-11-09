//go:build unified_api_demo
// +build unified_api_demo

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🚀 TSMorphGo 统一 API 演示")
	fmt.Println("=" + strings.Repeat("=", 50))

	// =============================================================================
	// 本文件演示新的统一 API 设计，替换原来分散的 IsXXX 和 AsXXX 函数
	// =============================================================================
	// 学习级别: 初级 → 中级
	// 预计时间: 15-20分钟
	//
	// 新 API 的优势:
	// - 统一的接口设计，无需记忆几十个函数名
	// - 支持类别检查，可以批量判断节点类型
	// - 更简洁的方法链调用
	// - 类型安全的转换接口
	//
	// 旧 API (已弃用):
	// - IsFunctionDeclaration(node)
	// - IsVariableDeclaration(node)
	// - AsFunctionDeclaration(node)
	// - AsVariableDeclaration(node)
	//
	// 新 API:
	// - node.IsDeclaration()
	// - node.IsKind(KindFunctionDeclaration)
	// - node.GetNodeName()
	// - node.AsDeclaration()
	// =============================================================================

	// 使用内存项目进行演示，不依赖外部文件
	project := tsmorphgo.NewProjectFromSources(map[string]string{
		"/src/types.ts": `
			// 用户接口定义
			export interface User {
				id: number;
				name: string;
				email: string;
				avatar?: string;
			}

			// API响应类型
			export interface ApiResponse<T> {
				data: T;
				status: number;
				message: string;
			}

			// 用户类型枚举
			export enum UserType {
				ADMIN = 'admin',
				USER = 'user',
				GUEST = 'guest'
			}

			// 用户类型别名
			export type UserRole = 'admin' | 'user' | 'guest';

			// 工具函数
			export function createUser(userData: Omit<User, 'id'>): User {
				return {
					id: Math.random(),
					...userData
				};
			}

			// 常量定义
			export const API_URL = 'https://api.example.com';
			export const MAX_USERS = 1000;
			export const DEFAULT_AVATAR = '/default-avatar.png';

			// 导入其他模块
			import { Logger } from './logger';
			import { Database } from './database';
		`,
		"/src/utils.ts": `
			// 工具函数集合
			export function formatDate(date: Date): string {
				return date.toISOString();
			}

			export function validateEmail(email: string): boolean {
				return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
			}

			// 默认导出
			export default {
				formatDate,
				validateEmail
			};
		`,
	})
	defer project.Close()

	// 获取源文件进行演示
	typesFile := project.GetSourceFile("/src/types.ts")
	if typesFile == nil {
		log.Fatal("未找到 types.ts 文件")
	}

	fmt.Printf("📄 分析文件: %s\n", typesFile.GetFilePath())
	fmt.Println("=" + strings.Repeat("=", 30))

	// 示例1: 统一的类型检查 API
	fmt.Println("\n🔍 示例1: 统一的类型检查 API")
	fmt.Println("展示如何使用新的统一接口进行类型检查")

	var (
		declarations = 0
		expressions  = 0
		types        = 0
		modules      = 0
		literals     = 0
	)

	typesFile.ForEachDescendant(func(node tsmorphgo.Node) {
		switch {
		case node.IsDeclaration():
			declarations++
			fmt.Printf("  ✅ 声明: %s\n", getNodeTypeDescription(node.GetKind()))
		case node.IsExpression():
			expressions++
			fmt.Printf("  🔵 表达式: %s\n", getNodeTypeDescription(node.GetKind()))
		case node.IsType():
			types++
			fmt.Printf("  🏷️ 类型: %s\n", getNodeTypeDescription(node.GetKind()))
		case node.IsModule():
			modules++
			fmt.Printf("  📦 模块: %s\n", getNodeTypeDescription(node.GetKind()))
		case node.IsLiteral():
			literals++
			if name, ok := node.GetNodeName(); ok {
				if value, ok := node.GetLiteralValue(); ok {
					fmt.Printf("  💎 字面量: %s = %v\n", name, value)
				}
			}
		}
	})

	fmt.Printf("\n📊 统计结果:\n")
	fmt.Printf("  - 声明类节点: %d\n", declarations)
	fmt.Printf("  - 表达式类节点: %d\n", expressions)
	fmt.Printf("  - 类型类节点: %d\n", types)
	fmt.Printf("  - 模块类节点: %d\n", modules)
	fmt.Printf("  - 字面量节点: %d\n", literals)

	// 示例2: 便捷的精确类型检查
	fmt.Println("\n🎯 示例2: 便捷的精确类型检查")
	fmt.Println("展示常用类型的便捷检查方法")

	var (
		functions    = 0
		interfaces   = 0
		classes      = 0
		variables    = 0
		imports      = 0
		calls        = 0
	)

	typesFile.ForEachDescendant(func(node tsmorphgo.Node) {
		switch {
		case node.IsFunctionDeclaration():
			functions++
			if name, ok := node.GetNodeName(); ok {
				fmt.Printf("  📝 函数: %s\n", name)
			}
		case node.IsInterfaceDeclaration():
			interfaces++
			if name, ok := node.GetNodeName(); ok {
				fmt.Printf("  🎭 接口: %s\n", name)
			}
		case node.IsClassDeclaration():
			classes++
			if name, ok := node.GetNodeName(); ok {
				fmt.Printf("  🏗️ 类: %s\n", name)
			}
		case node.IsVariableDeclaration():
			variables++
			if name, ok := node.GetNodeName(); ok {
				fmt.Printf("  📦 变量: %s\n", name)
			}
		case node.IsImportDeclaration():
			imports++
			text := strings.TrimSpace(node.GetText())
			if len(text) > 50 {
				text = text[:50] + "..."
			}
			fmt.Printf("  📥 导入: %s\n", text)
		case node.IsCallExpr():
			calls++
			text := strings.TrimSpace(node.GetText())
			if len(text) > 30 {
				text = text[:30] + "..."
			}
			fmt.Printf("  📞 调用: %s\n", text)
		}
	})

	fmt.Printf("\n📊 精确类型统计:\n")
	fmt.Printf("  - 函数声明: %d\n", functions)
	fmt.Printf("  - 接口声明: %d\n", interfaces)
	fmt.Printf("  - 类声明: %d\n", classes)
	fmt.Printf("  - 变量声明: %d\n", variables)
	fmt.Printf("  - 导入声明: %d\n", imports)
	fmt.Printf("  - 函数调用: %d\n", calls)

	// 示例3: 类型转换的统一接口
	fmt.Println("\n🔄 示例3: 类型转换的统一接口")
	fmt.Println("展示如何使用统一的转换接口")

	typesFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsDeclaration() {
			if result, ok := node.AsDeclaration(); ok {
				fmt.Printf("  🎯 转换声明成功: %T\n", result)
			}
		}
	})

	// 示例4: 多类型检查和复杂查询
	fmt.Println("\n🔬 示例4: 多类型检查和复杂查询")
	fmt.Println("展示如何进行复杂的类型查询")

	// 查找所有可能的声明类型
	declarationKinds := []tsmorphgo.SyntaxKind{
		tsmorphgo.KindFunctionDeclaration,
		tsmorphgo.KindInterfaceDeclaration,
		tsmorphgo.KindClassDeclaration,
		tsmorphgo.KindTypeAliasDeclaration,
		tsmorphgo.KindEnumDeclaration,
	}

	var complexDeclarations = 0
	typesFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsAnyKind(declarationKinds...) {
			complexDeclarations++
			if name, ok := node.GetNodeName(); ok {
				fmt.Printf("  🎯 复杂声明: %s (%s)\n", name, getNodeTypeDescription(node.GetKind()))
			}
		}
	})

	fmt.Printf("\n找到 %d 个复杂声明\n", complexDeclarations)

	// 示例5: 字面量值提取
	fmt.Println("\n💎 示例5: 字面量值提取")
	fmt.Println("展示如何从字面量节点提取值")

	var literalsFound = 0
	typesFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsLiteral() {
			literalsFound++
			if value, ok := node.GetLiteralValue(); ok {
				text := node.GetText()
				if len(text) > 30 {
					text = text[:30] + "..."
				}
				fmt.Printf("  💎 %s = %v\n", text, value)
			}
		}
	})

	fmt.Printf("\n找到 %d 个字面量值\n", literalsFound)

	fmt.Println("\n🎯 新 API 总结:")
	fmt.Println("1. 使用 node.IsDeclaration() 等类别方法进行批量检查")
	fmt.Println("2. 使用 node.IsKind(KindXxx) 进行精确类型检查")
	fmt.Println("3. 使用 node.GetNodeName() 获取节点名称")
	fmt.Println("4. 使用 node.AsDeclaration() 进行类型转换")
	fmt.Println("5. 使用 node.GetLiteralValue() 提取字面量值")
	fmt.Println("6. 使用 node.IsAnyKind(...) 检查多种类型")

	fmt.Println("\n✅ 统一 API 演示完成!")
	fmt.Println("新 API 大大简化了类型检查和转换的复杂度！")
}

// 辅助函数
func getNodeTypeDescription(kind tsmorphgo.SyntaxKind) string {
	switch kind {
	case tsmorphgo.KindFunctionDeclaration:
		return "函数声明"
	case tsmorphgo.KindInterfaceDeclaration:
		return "接口声明"
	case tsmorphgo.KindClassDeclaration:
		return "类声明"
	case tsmorphgo.KindVariableDeclaration:
		return "变量声明"
	case tsmorphgo.KindTypeAliasDeclaration:
		return "类型别名"
	case tsmorphgo.KindImportDeclaration:
		return "导入声明"
	case tsmorphgo.KindCallExpression:
		return "函数调用"
	case tsmorphgo.KindStringLiteral:
		return "字符串字面量"
	case tsmorphgo.KindNumericLiteral:
		return "数字字面量"
	default:
		return kind.String()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}