//go:build type_detection
// +build type_detection

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🏷️ TSMorphGo 类型检测示例")
	fmt.Println("=" + repeat("=", 50))

	// 使用真实的demo-react-app项目进行演示
	realProjectPath := "/Users/bird/Desktop/alalyzer/analyzer-ts/tsmorphgo/examples/demo-react-app"

	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:         realProjectPath,
		TargetExtensions: []string{".ts", ".tsx"},
		IgnorePatterns:   []string{"node_modules", "dist", ".git", "build"},
		UseTsConfig:      true,
	})
	defer project.Close()

	// 示例1: 基础类型检测
	fmt.Println("\n🔍 示例1: 基础类型检测")

	// 获取项目中的所有源文件
	sourceFiles := project.GetSourceFiles()
	if len(sourceFiles) == 0 {
		log.Fatal("未找到任何源文件")
	}

	fmt.Printf("项目包含 %d 个TypeScript文件:\n", len(sourceFiles))

	// 选择第一个文件进行演示
	var typesFile *tsmorphgo.SourceFile
	for _, file := range sourceFiles {
		if file != nil && strings.HasSuffix(file.GetFilePath(), ".ts") {
			typesFile = file
			break
		}
	}

	if typesFile == nil {
		log.Fatal("未找到可用的TypeScript文件")
	}

	fmt.Printf("分析文件: %s\n", typesFile.GetFilePath())

	// 统计各种节点类型
	typeStats := make(map[string]int)
	typesFile.ForEachDescendant(func(node tsmorphgo.Node) {
		typeName := node.GetKindName()
		if typeName != "" {
			typeStats[typeName]++
		}
	})

	fmt.Println("文件中的节点类型统计:")
	for _, typeName := range []string{"InterfaceDeclaration", "EnumDeclaration", "TypeAliasDeclaration", "PropertySignature"} {
		count := typeStats[typeName]
		if count > 0 {
			fmt.Printf("  - %s: %d 个\n", typeName, count)
		}
	}

	// 示例2: 接口检测
	fmt.Println("\n🔧 示例2: 接口检测与分析")

	typesFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsInterfaceDeclaration(node) {
			fmt.Printf("接口: %s (行 %d)\n",
				strings.TrimSpace(node.GetText()[:30])+"...",
				node.GetStartLineNumber())

			// 统计接口属性数量
			propertyCount := 0
			methodCount := 0
			node.ForEachDescendant(func(descendant tsmorphgo.Node) {
				switch descendant.Kind {
				case 298: // PropertySignature
					propertyCount++
				case 299: // MethodSignature
					methodCount++
				}
			})

			fmt.Printf("  - 属性数量: %d\n", propertyCount)
			fmt.Printf("  - 方法数量: %d\n", methodCount)

			// 获取接口名称
			if nameNode, ok := tsmorphgo.GetFirstChild(node, func(child tsmorphgo.Node) bool {
				return tsmorphgo.IsIdentifier(child)
			}); ok {
				fmt.Printf("  - 接口名: %s\n", strings.TrimSpace(nameNode.GetText()))
			}
		}
	})

	// 示例3: 枚举检测
	fmt.Println("\n🔤 示例3: 枚举检测")

	typesFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsEnumDeclaration(node) {
			fmt.Printf("枚举: %s (行 %d)\n",
				strings.TrimSpace(node.GetText()[:50])+"...",
				node.GetStartLineNumber())

			// 获取枚举成员
			memberCount := 0
			node.ForEachDescendant(func(descendant tsmorphgo.Node) {
				if descendant.Kind == 258 { // EnumMember
					memberCount++
					if memberCount <= 5 { // 只显示前5个成员
						fmt.Printf("  - 成员: %s\n", strings.TrimSpace(descendant.GetText()))
					}
				}
			})
			if memberCount > 5 {
				fmt.Printf("  - ... 还有 %d 个成员\n", memberCount-5)
			}
		}
	})

	// 示例4: 函数和方法检测
	fmt.Println("\n⚡ 示例4: 函数和方法检测")

	serviceFile := project.GetSourceFile("/src/services/userService.ts")
	if serviceFile != nil {
		var functions, methods, asyncFunctions []tsmorphgo.Node

		serviceFile.ForEachDescendant(func(node tsmorphgo.Node) {
			switch {
			case tsmorphgo.IsFunctionDeclaration(node):
				nodeCopy := node
				functions = append(functions, nodeCopy)
			case tsmorphgo.IsMethodDeclaration(node):
				nodeCopy := node
				methods = append(methods, nodeCopy)
			}

			// 检查异步函数
			node.ForEachDescendant(func(descendant tsmorphgo.Node) {
				if descendant.Kind == 164 { // AsyncKeyword
					if len(asyncFunctions) == 0 || asyncFunctions[len(asyncFunctions)-1].GetStartLineNumber() != node.GetStartLineNumber() {
						nodeCopy := node
						asyncFunctions = append(asyncFunctions, nodeCopy)
					}
				}
			})
		})

		fmt.Printf("发现 %d 个函数:\n", len(functions))
		for i, fn := range functions {
			if nameNode, ok := tsmorphgo.GetFunctionDeclarationNameNode(fn); ok {
				fmt.Printf("  %d. %s (行 %d)\n", i+1, strings.TrimSpace(nameNode.GetText()), fn.GetStartLineNumber())
			}
		}

		fmt.Printf("\n发现 %d 个方法:\n", len(methods))
		for i, method := range methods {
			if nameNode, ok := tsmorphgo.GetFirstChild(method, func(child tsmorphgo.Node) bool {
				return tsmorphgo.IsIdentifier(child)
			}); ok {
				fmt.Printf("  %d. %s() (行 %d)\n", i+1, strings.TrimSpace(nameNode.GetText()), method.GetStartLineNumber())
			}
		}

		fmt.Printf("\n发现 %d 个异步函数/方法:\n", len(asyncFunctions))
		for i, asyncFn := range asyncFunctions {
			text := strings.TrimSpace(asyncFn.GetText()[:60]) + "..."
			fmt.Printf("  %d. async (行 %d): %s\n", i+1, asyncFn.GetStartLineNumber(), text)
		}
	}

	// 示例5: 类型导入和导出检测
	fmt.Println("\n📦 示例5: 导入导出检测")

	allFiles := project.GetSourceFiles()
	totalImports, totalExports := 0, 0

	for _, file := range allFiles {
		fileImports, fileExports := 0, 0

		file.ForEachDescendant(func(node tsmorphgo.Node) {
			switch {
			case node.Kind == 266: // ImportDeclaration
				fileImports++
			case node.Kind == 148: // ExportKeyword
				fileExports++
			}
		})

		if fileImports > 0 || fileExports > 0 {
			fmt.Printf("文件 %s: %d 个导入, %d 个导出\n",
				file.GetFilePath(), fileImports, fileExports)
		}

		totalImports += fileImports
		totalExports += fileExports
	}

	fmt.Printf("\n总计: %d 个导入, %d 个导出\n", totalImports, totalExports)

	// 示例6: 复杂类型分析
	fmt.Println("\n🎯 示例6: 复杂类型分析")

	helperFile := project.GetSourceFile("/src/utils/helpers.ts")
	if helperFile != nil {
		fmt.Println("分析高级类型工具...")

		helperFile.ForEachDescendant(func(node tsmorphgo.Node) {
			// 查找类型别名
			if tsmorphgo.IsTypeAliasDeclaration(node) {
				text := strings.TrimSpace(node.GetText())
				if strings.Contains(text, "Optional<") || strings.Contains(text, "RequiredKeys<") {
					fmt.Printf("高级类型工具: %s\n", text[:80]+"...")
				}
			}

			// 查找函数重载
			if tsmorphgo.IsFunctionDeclaration(node) {
				text := strings.TrimSpace(node.GetText())
				if strings.Contains(text, "export function formatUserInfo") {
					fmt.Printf("函数重载示例: %s\n", text[:80]+"...")
				}
			}

			// 查找类型守卫
			if tsmorphgo.IsFunctionDeclaration(node) {
				text := strings.TrimSpace(node.GetText())
				if strings.Contains(text, "is User") {
					fmt.Printf("类型守卫函数: %s\n", text[:80]+"...")
				}
			}
		})
	}

	fmt.Println("\n✅ 类型检测示例完成!")
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}