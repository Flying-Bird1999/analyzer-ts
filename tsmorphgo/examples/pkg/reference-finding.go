//go:build reference_finding
// +build reference_finding

package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🔗 TSMorphGo - 引用查找与跳转定义")
	fmt.Println("=" + strings.Repeat("=", 50))

	// =============================================================================
	// 本文件演示如何正确使用 tsmorphgo 的引用查找、跳转定义和缓存功能。
	//
	// 核心 API:
	// - FindReferencesWithCache(node): 查找符号引用（带缓存）。
	// - GotoDefinition(node): 从引用跳转到定义。
	// =============================================================================

	// 1. 初始化项目
	realProjectPath, err := filepath.Abs("../demo-react-app")
	if err != nil {
		log.Fatalf("无法解析项目路径: %v", err)
	}

	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:         realProjectPath,
		TargetExtensions: []string{".ts", ".tsx"},
		IgnorePatterns:   []string{"node_modules", "dist", ".git", "build"},
		UseTsConfig:      true,
	})
	defer project.Close()

	fmt.Printf("✅ 成功加载项目: %s\n", realProjectPath)
	fmt.Printf("📊 分析 %d 个源文件...\n", len(project.GetSourceFiles()))

	// 2. 定位一个用于分析的起始节点
	// 我们将查找 `src/types.ts` 文件中的 `User` 接口声明
	typesFilePath := filepath.Join(realProjectPath, "src/types.ts")
	typesFile := project.GetSourceFile(typesFilePath)
	if typesFile == nil {
		log.Fatalf("未找到目标文件: %s", typesFilePath)
	}

	var userInterfaceIdentifier *tsmorphgo.Node
	typesFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if userInterfaceIdentifier != nil {
			return // 已经找到
		}
		// 找到名为 "User" 的标识符
		if node.IsIdentifierNode() && strings.TrimSpace(node.GetText()) == "User" {
			parent := node.GetParent()
			if parent != nil {
				if parent.IsInterfaceDeclaration() {
					userInterfaceIdentifier = &node
				}
			}
		}
	})

	if userInterfaceIdentifier == nil {
		log.Fatal("在 src/types.ts 中未找到 'User' 接口的标识符节点")
	}

	fmt.Printf("\n🎯 查找到分析起点: 'User' 接口 (位于 %s:%d)\n",
		filepath.Base(userInterfaceIdentifier.GetSourceFile().GetFilePath()),
		userInterfaceIdentifier.GetStartLineNumber())

	// 3. 示例 1: 查找 'User' 接口的所有引用
	fmt.Println("\n" + strings.Repeat("-", 20) + " 示例 1: 查找引用 " + strings.Repeat("-", 20))
	findAndPrintReferences(userInterfaceIdentifier, realProjectPath)

	// 4. 示例 2: 演示缓存带来的性能提升
	fmt.Println("\n" + strings.Repeat("-", 20) + " 示例 2: 缓存性能 " + strings.Repeat("-", 20))
	demonstrateCaching(userInterfaceIdentifier)

	// 5. 示例 3: 从一个引用点跳转到定义
	fmt.Println("\n" + strings.Repeat("-", 20) + " 示例 3: 跳转到定义 " + strings.Repeat("-", 20))
	findUsageAndGoToDefinition(project, realProjectPath)
}

// findAndPrintReferences 查找并打印给定节点的引用
func findAndPrintReferences(node *tsmorphgo.Node, basePath string) {
	fmt.Println("执行 FindReferencesWithCache...")
	start := time.Now()
	refs, fromCache, err := tsmorphgo.FindReferencesWithCache(*node)
	duration := time.Since(start)

	if err != nil {
		log.Fatalf("查找引用失败: %v", err)
	}

	fmt.Printf("✅ 查找完成! (耗时: %v, 来自缓存: %v)\n", duration, fromCache)
	fmt.Printf("📊 共找到 %d 个引用。\n", len(refs))

	// 按文件对引用进行分组
	refsByFile := make(map[string][]tsmorphgo.Node)
	for _, ref := range refs {
		path := ref.GetSourceFile().GetFilePath()
		refsByFile[path] = append(refsByFile[path], *ref)
	}

	fmt.Println("📄 引用分布:")
	for path, fileRefs := range refsByFile {
		relPath, _ := filepath.Rel(basePath, path)
		fmt.Printf("  - %s (%d 个引用)\n", relPath, len(fileRefs))
		for i, r := range fileRefs {
			if i >= 3 { // 最多显示3个
				fmt.Printf("    ... 等\n")
				break
			}
			fmt.Printf("    - 第 %d 行: '%s'\n", r.GetStartLineNumber(), truncateString(r.GetParent().GetText(), 60))
		}
	}
}

// demonstrateCaching 演示引用查找的缓存效果
func demonstrateCaching(node *tsmorphgo.Node) {
	// 第一次查找，应该会比较慢，因为需要调用LSP
	fmt.Println("第一次查找 (预期调用 LSP)...")
	start1 := time.Now()
	_, fromCache1, err1 := tsmorphgo.FindReferencesWithCache(*node)
	duration1 := time.Since(start1)
	if err1 != nil {
		log.Printf("第一次查找失败: %v", err1)
		return
	}
	fmt.Printf("  - 耗时: %v, 来自缓存: %v\n", duration1, fromCache1)

	// 第二次查找，应该非常快，因为直接从缓存读取
	fmt.Println("第二次查找 (预期来自缓存)...")
	start2 := time.Now()
	_, fromCache2, err2 := tsmorphgo.FindReferencesWithCache(*node)
	duration2 := time.Since(start2)
	if err2 != nil {
		log.Printf("第二次查找失败: %v", err2)
		return
	}
	fmt.Printf("  - 耗时: %v, 来自缓存: %v\n", duration2, fromCache2)

	if !fromCache1 && fromCache2 && duration1 > duration2 {
		fmt.Printf("✅ 缓存工作正常! 性能提升约 %.2fx\n", float64(duration1)/float64(duration2))
	} else {
		fmt.Println("⚠️ 缓存效果不明显或未按预期工作。")
	}
}

// findUsageAndGoToDefinition 找到一个使用点，并从中跳转到定义
func findUsageAndGoToDefinition(project *tsmorphgo.Project, basePath string) {
	// 在 `src/App.tsx` 中找到 `User` 类型的使用点
	appFilePath := filepath.Join(basePath, "src/App.tsx")
	appFile := project.GetSourceFile(appFilePath)
	if appFile == nil {
		log.Printf("警告: 未找到 App.tsx 文件，跳过跳转定义示例。")
		return
	}

	var userUsageNode *tsmorphgo.Node
	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if userUsageNode != nil {
			return
		}
		// 查找 `useState<User[]>` 中的 `User`
		if node.IsIdentifierNode() && strings.TrimSpace(node.GetText()) == "User" {
			if parent := node.GetParent(); parent != nil && parent.IsKind(tsmorphgo.KindTypeReference) {
				userUsageNode = &node
			}
		}
	})

	if userUsageNode == nil {
		log.Printf("警告: 在 App.tsx 中未找到 'User' 的使用点，跳过跳转定义示例。")
		return
	}

	fmt.Printf("\n从 'User' 的一个使用点 (%s:%d) 跳转到定义...\n",
		filepath.Base(userUsageNode.GetSourceFile().GetFilePath()),
		userUsageNode.GetStartLineNumber())

	defs, err := tsmorphgo.GotoDefinition(*userUsageNode)
	if err != nil {
		log.Fatalf("跳转到定义失败: %v", err)
	}

	fmt.Printf("✅ 跳转成功! 找到 %d 个定义位置:\n", len(defs))
	for _, def := range defs {
		relPath, _ := filepath.Rel(basePath, def.GetSourceFile().GetFilePath())
		fmt.Printf("  - %s (第 %d 行)\n", relPath, def.GetStartLineNumber())
	}
}

// truncateString 是一个辅助函数，用于截断长字符串
func truncateString(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}