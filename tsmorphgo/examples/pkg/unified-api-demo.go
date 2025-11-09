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
	// 本文件旨在清晰地演示“统一API”的设计理念和核心用法。
	// 为了聚焦API本身，我们使用一个精简的、自包含的内存项目。
	//
	// 核心概念:
	// 1. 类别检查: 使用 IsDeclaration(), IsExpression() 等方法对节点进行分类。
	// 2. 精确检查: 使用 IsKind() 和 IsAnyKind() 进行精确的节点类型匹配。
	// 3. 统一访问: 使用 GetNodeName() 和 GetLiteralValue() 从不同类型的节点获取信息。
	// 4. 统一转换: 使用 AsDeclaration() 等方法安全地转换节点类型。
	// =============================================================================

	// 创建一个精心设计的内存项目，用于演示
	project := tsmorphgo.NewProjectFromSources(map[string]string{
		"/main.ts": `
import { Greeter } from './utils';

const PI = 3.14;
let message = "Hello, World!";

interface User {
    name: string;
    id: number;
}

function greet(user: User) {
    const greeter = new Greeter(message);
    return greeter.greet(user.name);
}

const result = greet({ name: "TypeScript", id: 1 });
`,
	})
	defer project.Close()

	mainFile := project.GetSourceFile("/main.ts")
	if mainFile == nil {
		log.Fatal("未能加载用于演示的 main.ts 文件")
	}

	fmt.Printf("📄 分析文件: %s\n", mainFile.GetFilePath())

	// 示例1: 类别检查 (Category Checking)
	fmt.Println("\n" + strings.Repeat("-", 20) + " 示例1: 类别检查 " + strings.Repeat("-", 20))
	fmt.Println("演示如何使用 IsDeclaration(), IsExpression(), IsLiteral() 等方法对节点进行分类。")

	mainFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsDeclaration() {
			if name, ok := node.GetNodeName(); ok {
				fmt.Printf("  [声明] 发现 '%s' (%s)\n", name, node.GetKind().String())
			}
		}
		if node.IsExpression() {
			fmt.Printf("  [表达式] 发现: '%s'\n", truncateString(node.GetText(), 30))
		}
		if node.IsLiteral() {
			if value, ok := node.GetLiteralValue(); ok {
				fmt.Printf("  [字面量] 发现: %v\n", value)
			}
		}
	})

	// 示例2: 精确和多类型检查 (Precise & Multi-Kind Checking)
	fmt.Println("\n" + strings.Repeat("-", 20) + " 示例2: 精确检查 " + strings.Repeat("-", 20))
	fmt.Println("演示如何使用 IsKind() 和 IsAnyKind() 进行精确匹配。")

	// 使用 IsKind 查找所有接口声明
	fmt.Println("\n  --- 使用 IsKind() 查找接口 ---")
	mainFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsKind(tsmorphgo.KindInterfaceDeclaration) {
			if name, ok := node.GetNodeName(); ok {
				fmt.Printf("    - 找到接口声明: %s\n", name)
			}
		}
	})

	// 使用 IsAnyKind 查找所有常量或变量
	fmt.Println("\n  --- 使用 IsAnyKind() 查找常量和变量 ---")
	mainFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsAnyKind(tsmorphgo.KindVariableDeclaration, tsmorphgo.KindVariableStatement) {
			if name, ok := node.GetNodeName(); ok {
				fmt.Printf("    - 找到变量/常量: %s\n", name)
			}
		}
	})

	// 示例3: 统一的名称和值获取 (Unified Name & Value Getters)
	fmt.Println("\n" + strings.Repeat("-", 20) + " 示例3: 统一访问 " + strings.Repeat("-", 20))
	fmt.Println("演示 GetNodeName() 和 GetLiteralValue() 如何适用于多种节点。")

	mainFile.ForEachDescendant(func(node tsmorphgo.Node) {
		// GetNodeName() 适用于多种声明
		if node.IsDeclaration() {
			if name, ok := node.GetNodeName(); ok {
				fmt.Printf("  [名称] GetNodeName() 从 %s 中提取到: %s\n", node.GetKind().String(), name)
			}
		}
		// GetLiteralValue() 适用于多种字面量
		if node.IsLiteral() {
			if value, ok := node.GetLiteralValue(); ok {
				fmt.Printf("  [值] GetLiteralValue() 从 %s 中提取到: %v\n", node.GetKind().String(), value)
			}
		}
	})

	// 示例4: 统一的类型转换 (Unified Type Conversion)
	fmt.Println("\n" + strings.Repeat("-", 20) + " 示例4: 统一转换 " + strings.Repeat("-", 20))
	fmt.Println("演示 AsDeclaration() 如何安全地转换节点。")

	mainFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsDeclaration() {
			if specificDecl, ok := node.AsDeclaration(); ok {
				fmt.Printf("  [转换] 节点 %s 成功转换为类型 %T\n", node.GetKind().String(), specificDecl)
			}
		}
	})

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("✅ 统一 API 演示完成!")
	fmt.Println("这个示例清晰地展示了统一API如何让代码分析更简洁、更具可读性。")
}

func truncateString(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}