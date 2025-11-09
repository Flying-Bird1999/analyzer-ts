//go:build type_detection
// +build type_detection

package main

import (
	"fmt"
	"log"
	"path/filepath"
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

	// 获取 demo-react-app 的绝对路径
	realProjectPath, err := filepath.Abs("../demo-react-app")
	if err != nil {
		log.Fatalf("无法解析项目路径: %v", err)
	}

	// 使用真实项目进行演示
	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:         realProjectPath,
		TargetExtensions: []string{".ts", ".tsx"},
		IgnorePatterns:   []string{"node_modules", "dist", ".git", "build"},
		UseTsConfig:      true,
	})
	defer project.Close()

	fmt.Printf("✅ 成功加载真实项目: %s\n", realProjectPath)
	fmt.Printf("📊 分析 %d 个文件...\n", len(project.GetSourceFiles()))

	// 在一个循环中执行所有分析
	runFullAnalysis(project)

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

// runFullAnalysis 对整个项目进行全面的类型分析
func runFullAnalysis(project *tsmorphgo.Project) {
	// 统计数据容器
	stats := struct {
		interfaces      int
		enums           int
		typeAliases     int
		declarations    int
		expressions     int
		statements      int
		types           int
		modules         int
		classCount      int
		varFuncCount    int
		callExpressions int
		propertyAccess  int
		binaryExprs     int
		names           []string
		literals        []interface{}
	}{}

	// 定义要查找的声明类型
	declarationKinds := []tsmorphgo.SyntaxKind{
		tsmorphgo.KindVariableDeclaration,
		tsmorphgo.KindFunctionDeclaration,
	}

	// 遍历项目中的所有文件
	for _, file := range project.GetSourceFiles() {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// --- 基础类型检测 ---
			switch {
			case node.IsInterfaceDeclaration():
				stats.interfaces++
			case node.IsKind(tsmorphgo.KindEnumDeclaration):
				stats.enums++
			case node.IsKind(tsmorphgo.KindTypeAliasDeclaration):
				stats.typeAliases++
			}

			// --- 类别检测 ---
			if node.IsDeclaration() {
				stats.declarations++
			}
			if node.IsExpression() {
				stats.expressions++
			}
			if node.IsStatement() {
				stats.statements++
			}
			if node.IsType() {
				stats.types++
			}
			if node.IsModule() {
				stats.modules++
			}

			// --- 多类型检查 ---
			if node.IsClassDeclaration() {
				stats.classCount++
			}
			if node.IsAnyKind(declarationKinds...) {
				stats.varFuncCount++
			}

			// --- 精确类型检查 ---
			if node.IsCallExpr() {
				stats.callExpressions++
			}
			if node.IsPropertyAccessExpression() {
				stats.propertyAccess++
			}
			if node.IsKind(tsmorphgo.KindBinaryExpression) {
				stats.binaryExprs++
			}

			// --- 名称和值提取 ---
			if node.IsDeclaration() {
				if name, ok := node.GetNodeName(); ok {
					stats.names = append(stats.names, name)
				}
			}
			if node.IsLiteral() {
				if value, ok := node.GetLiteralValue(); ok {
					stats.literals = append(stats.literals, value)
				}
			}
		})
	}

	// --- 打印所有统计结果 ---
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("📊 全项目类型分析统计结果")
	fmt.Println(strings.Repeat("-", 50))

	fmt.Println("\n🔍 基础类型统计:")
	fmt.Printf("  - 接口声明 (Interfaces): %d\n", stats.interfaces)
	fmt.Printf("  - 枚举声明 (Enums): %d\n", stats.enums)
	fmt.Printf("  - 类型别名 (Type Aliases): %d\n", stats.typeAliases)

	fmt.Println("\n🎯 节点类别统计:")
	fmt.Printf("  - 声明类节点 (Declarations): %d\n", stats.declarations)
	fmt.Printf("  - 表达式类节点 (Expressions): %d\n", stats.expressions)
	fmt.Printf("  - 语句类节点 (Statements): %d\n", stats.statements)
	fmt.Printf("  - 类型类节点 (Types): %d\n", stats.types)
	fmt.Printf("  - 模块类节点 (Modules): %d\n", stats.modules)

	fmt.Println("\n🔬 多类型检查统计:")
	fmt.Printf("  - 类声明 (Classes): %d\n", stats.classCount)
	fmt.Printf("  - 变量或函数声明 (Variables/Functions): %d\n", stats.varFuncCount)

	fmt.Println("\n⚡ 精确类型统计:")
	fmt.Printf("  - 函数调用 (Call Expressions): %d\n", stats.callExpressions)
	fmt.Printf("  - 属性访问 (Property Access): %d\n", stats.propertyAccess)
	fmt.Printf("  - 二元表达式 (Binary Expressions): %d\n", stats.binaryExprs)

	fmt.Println("\n💎 名称和值提取统计:")
	fmt.Printf("  - 提取的声明名称总数: %d\n", len(stats.names))
	if len(stats.names) > 0 {
		fmt.Printf("    - 示例名称: %s\n", strings.Join(stats.names[:min(5, len(stats.names))], ", "))
	}
	fmt.Printf("  - 提取的字面量总数: %d\n", len(stats.literals))
	if len(stats.literals) > 0 {
		fmt.Printf("    - 示例字面量: %v\n", stats.literals[0])
	}
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
