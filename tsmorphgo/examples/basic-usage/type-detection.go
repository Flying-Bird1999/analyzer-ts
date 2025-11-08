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

	// 初始化项目
	realProjectPath := "/Users/bird/Desktop/alalyzer/analyzer-ts/tsmorphgo/examples/demo-react-app"
	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:         realProjectPath,
		TargetExtensions: []string{".ts", ".tsx"},
		IgnorePatterns:   []string{"node_modules", "dist", ".git", "build"},
		UseTsConfig:      true,
	})
	defer project.Close()

	// 我们选择 types.ts 文件作为分析的起点，因为它包含了丰富的类型定义。
	typesFile := project.GetSourceFile(realProjectPath + "/src/types.ts")
	if typesFile == nil {
		log.Fatal("未找到 types.ts 文件")
	}
	fmt.Printf("分析文件: %s\n", typesFile.GetFilePath())

	// 示例1: 遍历并统计节点类型
	// 演示如何获取节点的类型名称(KindName)并进行统计。
	fmt.Println("\n🔍 示例1: 统计文件中的节点类型")
	typeStats := make(map[string]int)
	typesFile.ForEachDescendant(func(node tsmorphgo.Node) {
		// GetKindName() 获取节点类型的可读名称，如 "InterfaceDeclaration"。
		// 对应 ts-morph 的 `node.getKindName()`。
		typeName := node.GetKindName()
		if typeName != "" {
			typeStats[typeName]++
		}
	})

	fmt.Println("文件中的主要节点类型统计:")
	for _, typeName := range []string{"InterfaceDeclaration", "TypeAliasDeclaration", "PropertySignature", "Identifier"} {
		count := typeStats[typeName]
		if count > 0 {
			fmt.Printf("  - %s: %d 个\n", typeName, count)
		}
	}

	// 示例2: 接口检测与分析 (InterfaceDeclaration)
	// 演示如何找到所有的接口声明，并分析其内部结构。
	fmt.Println("\n🔧 示例2: 接口检测与分析")
	typesFile.ForEachDescendant(func(node tsmorphgo.Node) {
		// IsInterfaceDeclaration 判断节点是否为接口声明。
		// 对应 ts-morph 的 `Node.isInterfaceDeclaration(node)`。
		if tsmorphgo.IsInterfaceDeclaration(node) {
			// GetFirstChild 获取第一个子节点，这里用它来获取接口的名称。
			if nameNode, ok := tsmorphgo.GetFirstChild(node, tsmorphgo.IsIdentifier); ok {
				fmt.Printf("\n找到接口: %s (行 %d)\n", nameNode.GetText(), node.GetStartLineNumber())
			}

			// 遍历接口内部，统计属性和方法的数量
			propertyCount := 0
			methodCount := 0
			node.ForEachDescendant(func(descendant tsmorphgo.Node) {
				// KindPropertySignature 和 KindMethodSignature 是属性和方法签名的类型枚举。
				if descendant.Kind == tsmorphgo.KindPropertySignature {
					propertyCount++
				} else if descendant.Kind == tsmorphgo.KindMethodSignature {
					methodCount++
				}
			})

			fmt.Printf("  - 属性数量: %d\n", propertyCount)
			fmt.Printf("  - 方法数量: %d\n", methodCount)
		}
	})

	// 示例3: 类型别名检测 (TypeAliasDeclaration)
	// 演示如何找到 `type` 关键字定义的类型别名。
	fmt.Println("\n📜 示例3: 类型别名检测")
	typesFile.ForEachDescendant(func(node tsmorphgo.Node) {
		// IsTypeAliasDeclaration 判断节点是否为类型别名声明。
		// 对应 ts-morph 的 `Node.isTypeAliasDeclaration(node)`。
		if tsmorphgo.IsTypeAliasDeclaration(node) {
			if nameNode, ok := tsmorphgo.GetFirstChild(node, tsmorphgo.IsIdentifier); ok {
				fmt.Printf("\n找到类型别名: %s (行 %d)\n", nameNode.GetText(), node.GetStartLineNumber())
				fullText := strings.TrimSpace(node.GetText())
				if len(fullText) > 80 {
					fullText = fullText[:80] + "..."
				}
				fmt.Printf("  - 完整定义: %s\n", fullText)
			}
		}
	})

	// 示例4: 函数和方法检测 (FunctionDeclaration, MethodDeclaration)
	fmt.Println("\n⚡ 示例4: 函数和方法检测")
	serviceFile := project.GetSourceFile(realProjectPath + "/src/services/api.ts")
	if serviceFile != nil {
		fmt.Printf("\n分析文件: %s\n", serviceFile.GetFilePath())
		// IsMethodDeclaration 判断节点是否为类的方法声明。
		serviceFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if tsmorphgo.IsMethodDeclaration(node) {
				if nameNode, ok := tsmorphgo.GetFirstChild(node, tsmorphgo.IsIdentifier); ok {
					fmt.Printf("找到方法: %s (行 %d)\n", nameNode.GetText(), node.GetStartLineNumber())

					// 检查方法是否为异步 (async)
					isAsync := false
					if _, ok := tsmorphgo.GetFirstChild(node, func(n tsmorphgo.Node) bool { return n.Kind == tsmorphgo.KindAsyncKeyword }); ok {
						isAsync = true
					}
					fmt.Printf("  - 是否异步: %v\n", isAsync)
				}
			}
		})
	}

	// 示例5: 导入和导出检测 (ImportDeclaration, ExportKeyword)
	fmt.Println("\n📦 示例5: 导入导出检测")
	totalImports, totalExports := 0, 0
	for _, file := range project.GetSourceFiles() {
		fileImports, fileExports := 0, 0
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// KindImportDeclaration 是导入声明的类型枚举。
			if node.Kind == tsmorphgo.KindImportDeclaration {
				fileImports++
			}
			// KindExportKeyword 是 `export` 关键字的类型枚举。
			if node.Kind == tsmorphgo.KindExportKeyword {
				fileExports++
			}
		})

		if fileImports > 0 || fileExports > 0 {
			totalImports += fileImports
			totalExports += fileExports
		}
	}
	fmt.Printf("项目总计: %d 个导入声明, %d 个导出关键字\n", totalImports, totalExports)

	// 示例6: 类型守卫检测 (Type Guard)
	// 类型守卫是一种特殊的函数，它会返回一个 `parameterName is Type` 形式的布尔值。
	fmt.Println("\n🛡️ 示例6: 类型守卫分析")
	utilsFile := project.GetSourceFile(realProjectPath + "/src/utils.ts")
	if utilsFile != nil {
		fmt.Printf("\n分析文件: %s\n", utilsFile.GetFilePath())
		utilsFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if tsmorphgo.IsFunctionDeclaration(node) {
				// 这是一个简化的检查，通过检查函数文本中是否包含 `is User` 来判断。
				// 在实际应用中，需要更精确地分析函数的返回类型节点。
				if strings.Contains(node.GetText(), "is User") {
					if name, ok := tsmorphgo.GetFunctionDeclarationNameNode(node); ok {
						fmt.Printf("可能是一个类型守卫函数: %s\n", name.GetText())
					}
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