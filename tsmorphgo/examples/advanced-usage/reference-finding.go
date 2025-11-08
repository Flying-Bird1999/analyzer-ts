//go:build reference_finding
// +build reference_finding

package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🔗 TSMorphGo 引用查找示例 (基于真实项目)")
	fmt.Println("=" + repeat("=", 50))

	// 1. 初始化项目
	// 通过 tsmorphgo.NewProject 创建一个项目实例。
	// 这里我们指向一个真实的React项目目录，并设置 UseTsConfig: true 来自动加载 tsconfig.json 文件。
	// 这与 ts-morph 中的 `new Project({ tsConfigFilePath: ... })` 思想一致。
	realProjectPath := "/Users/bird/Desktop/alalyzer/analyzer-ts/tsmorphgo/examples/demo-react-app"
	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:         realProjectPath,
		TargetExtensions: []string{".ts", ".tsx"},
		IgnorePatterns:   []string{"node_modules", "dist", ".git", "build"},
		UseTsConfig:      true,
	})
	defer project.Close() // 确保在函数结束时释放项目资源

	// 示例1: 基础引用查找
	fmt.Println("\n🔍 示例1: 基础引用查找")

	// 2. 获取源文件
	// 使用 project.GetSourceFile 获取项目中特定的源文件。
	hooksFile := project.GetSourceFile(realProjectPath + "/src/hooks/useUserQuery.ts")
	if hooksFile == nil {
		log.Fatal("useUserQuery.ts 文件未找到")
	}

	// 3. 查找目标节点
	// 遍历AST（抽象语法树）来找到我们想要分析的节点。
	// 在这里，我们想找到 `useUsers` 这个自定义Hook的声明位置。
	var useUsersNode *tsmorphgo.Node
	// ForEachDescendant 会深度优先遍历一个节点下的所有子孙节点。
	hooksFile.ForEachDescendant(func(node tsmorphgo.Node) {
		// IsVariableDeclaration 检查当前节点是否是一个变量声明。
		if tsmorphgo.IsVariableDeclaration(node) {
			// GetVariableName 是一个辅助函数，用于获取变量声明的名称。
			if name, ok := tsmorphgo.GetVariableName(node); ok && name == "useUsers" {
				// GetFirstChild 用来获取符合条件的第一个子节点，这里我们用它来获取变量名对应的标识符(Identifier)节点。
				if nameNode, ok := tsmorphgo.GetFirstChild(node, tsmorphgo.IsIdentifier); ok {
					useUsersNode = nameNode // 直接赋值指针
					return                  // 找到后提前终止遍历
				}
			}
		}
	})

	if useUsersNode == nil {
		log.Fatal("未找到 useUsers 变量声明")
	}

	// 4. 执行引用查找
	// GetSourceFile() 和 GetStartLineNumber() 用于获取节点的位置信息。
	fmt.Printf("`useUsers` 变量声明位置: %s:%d\n", useUsersNode.GetSourceFile().GetFilePath(), useUsersNode.GetStartLineNumber())

	// tsmorphgo.FindReferences 是核心功能，它会利用LSP服务查找一个节点在整个项目中的所有引用。
	// 这对应 ts-morph 中的 `identifier.findReferencesAsNodes()`。
	refs, err := tsmorphgo.FindReferences(*useUsersNode)
	if err != nil {
		log.Printf("查找引用失败: %v", err)
		return
	}

	// 5. 处理和展示引用结果
	fmt.Printf("找到 %d 个 `useUsers` 引用:\n", len(refs))
	for i, ref := range refs {
		// GetParent() 获取节点的父节点，用于展示引用的上下文。
		parent := ref.GetParent()
		context := ""
		if parent != nil {
			// GetText() 获取节点在源码中的原始文本。
			parentText := strings.TrimSpace(parent.GetText())
			if len(parentText) > 80 {
				parentText = parentText[:80] + "..."
			}
			context = parentText
		}

		fmt.Printf("  %d. %s:%d - 上下文: %s\n",
			i+1, ref.GetSourceFile().GetFilePath(), ref.GetStartLineNumber(), context)
	}

	// 示例2: 带缓存的引用查找
	fmt.Println("\n⚡ 示例2: 带缓存的引用查找性能对比")

	if len(refs) > 0 {
		testRef := refs[0] // 使用第一个引用进行测试

		// 第一次查找会调用底层的LSP服务，耗时较长。
		start := time.Now()
		refs1, fromCache1, err := tsmorphgo.FindReferencesWithCache(*testRef)
		duration1 := time.Since(start)
		if err != nil {
			log.Printf("查找失败: %v", err)
			return
		}
		source1 := "LSP服务"
		if fromCache1 {
			source1 = "缓存"
		}
		fmt.Printf("第一次查找:\n")
		fmt.Printf("  - 耗时: %v\n", duration1)
		fmt.Printf("  - 来源: %s\n", source1)
		fmt.Printf("  - 引用数: %d\n", len(refs1))

		// 第二次查找同一个节点的引用，应该会命中缓存，速度极快。
		start = time.Now()
		refs2, fromCache2, err := tsmorphgo.FindReferencesWithCache(*testRef)
		duration2 := time.Since(start)
		if err != nil {
			log.Printf("查找失败: %v", err)
			return
		}
		source2 := "LSP服务"
		if fromCache2 {
			source2 = "缓存"
		}
		fmt.Printf("第二次查找:\n")
		fmt.Printf("  - 耗时: %v\n", duration2)
		fmt.Printf("  - 来源: %s\n", source2)
		fmt.Printf("  - 引用数: %d\n", len(refs2))

		if duration1 > 0 && duration2 > 0 {
			speedup := float64(duration1) / float64(duration2)
			fmt.Printf("  - 性能提升: %.1fx 倍\n", speedup)
		}
	}

	// 示例3: 跳转到定义
	// 对应 ts-morph 中的 `identifier.getDefinitionNodes()`。
	fmt.Println("\n📍 示例3: 跳转到定义")

	// 在 App.tsx 文件中找到对 `useUsers` 的使用，然后跳转到它的定义。
	appFile := project.GetSourceFile(realProjectPath + "/src/App.tsx")
	if appFile != nil {
		appFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if tsmorphgo.IsIdentifier(node) &&
				strings.TrimSpace(node.GetText()) == "useUsers" {
				// 确保我们找到的不是它自己的声明
				parent := node.GetParent()
				if parent != nil && tsmorphgo.IsVariableDeclaration(*parent) {
					return
				}

				// tsmorphgo.GotoDefinition 是核心功能，用于从一个使用点跳转到其定义点。
				defs, err := tsmorphgo.GotoDefinition(node)
				if err != nil {
					log.Printf("跳转到定义失败: %v", err)
					return
				}

				fmt.Printf("在 %s:%d 找到对 `useUsers` 的引用\n",
					node.GetSourceFile().GetFilePath(),
					node.GetStartLineNumber())

				fmt.Printf("跳转到定义结果:\n")
				for i, def := range defs {
					fmt.Printf("  %d. %s:%d - %s\n",
						i+1, def.GetSourceFile().GetFilePath(),
						def.GetStartLineNumber(),
						func() string {
							text := strings.TrimSpace(def.GetText())
							if len(text) > 80 {
								text = text[:80] + "..."
							}
							return text
						}())
				}
				return // 只演示一次
			}
		})
	}

	// 示例4: 错误处理和降级策略
	fmt.Println("\n🛡️ 示例4: 错误处理和降级策略")

	// 尝试查找一个不存在的变量的引用，预期会收到一个错误。
	// 这是一个好的实践，展示了当符号找不到时库如何优雅地失败。
	var nonExistentNode *tsmorphgo.Node
	// 动态创建一个用于测试的源文件和节点
	tempProject := tsmorphgo.NewProjectFromSources(map[string]string{
		"/temp.ts": "const a = nonExistentVar;",
	})
	defer tempProject.Close()
	tempFile := tempProject.GetSourceFile("/temp.ts")
	tempFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "nonExistentVar" {
			nonExistentNode = &node
		}
	})
	if nonExistentNode != nil {
		_, err := tsmorphgo.FindReferences(*nonExistentNode)
		if err != nil {
			fmt.Printf("预期内的错误处理: %v\n", err)
			fmt.Println("这种错误是正常的，因为我们查找的是一个不存在的变量的引用。")
		}
	} else {
		fmt.Println("未能创建用于错误处理的测试节点。")
	}

	fmt.Println("\n✅ 引用查找示例完成!")
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
