//go:build references_usage_example
// +build references_usage_example

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

// references_usage_example_fixed.go
//
// 这个示例展示了如何使用 TSMorphGo References API 的各种功能：
// 1. 基础引用查找
// 2. 带缓存的优化查找
// 3. 错误处理和重试
//

func main() {
	fmt.Println("=== TSMorphGo References API 使用示例 ===\n")

	// 示例1: 基础引用查找
	basicReferenceExample()

	// 示例2: 缓存优化性能
	cachePerformanceExample()

	// 示例3: 错误处理和重试
	errorHandlingExample()
}

// basicReferenceExample 基础引用查找示例
func basicReferenceExample() {
	fmt.Println("1. 基础引用查找示例")
	fmt.Println("==================")

	// 创建项目 - 使用内存源码方式
	sources := map[string]string{
		"/basic.ts": `
		const sharedVar = "hello world";

		function testFunction() {
			console.log(sharedVar);
			let localVar = sharedVar;
			return localVar;
		}

		testFunction();
		console.log(sharedVar);
		`,
	}
	project := tsmorphgo.NewProjectFromSources(sources)
	defer project.Close()

	// 获取源文件
	sourceFile := project.GetSourceFile("/basic.ts")
	if sourceFile == nil {
		log.Fatal("无法创建测试文件")
	}

	// 找到 sharedVar 的定义节点
	var definitionNode *tsmorphgo.Node
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsIdentifier(node) && node.GetText() == "sharedVar" {
			parent := node.GetParent()
			if parent != nil && parent.Kind == 214 { // VariableDeclaration
				nodeCopy := node
				definitionNode = &nodeCopy
			}
		}
	})

	if definitionNode == nil {
		log.Fatal("找不到 sharedVar 的定义")
	}

	// 查找定义位置
	defs, err := tsmorphgo.GotoDefinition(*definitionNode)
	if err != nil {
		log.Printf("查找定义失败: %v", err)
		return
	}

	fmt.Printf("定义位置: 找到 %d 个定义\n", len(defs))
	for i, def := range defs {
		fmt.Printf("  定义 %d: %s (行 %d, 列 %d)\n",
			i+1, def.GetText(), def.GetStartLineNumber(), def.GetStartColumnNumber())
	}

	// 找到 sharedVar 的使用节点
	var usageNode *tsmorphgo.Node
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsIdentifier(node) && node.GetText() == "sharedVar" {
			parent := node.GetParent()
			if parent != nil && parent.Kind != 214 { // 不是变量声明
				nodeCopy := node
				usageNode = &nodeCopy
				return // 找到第一个使用节点就返回
			}
		}
	})

	if usageNode == nil {
		// 如果没有找到使用节点，就使用定义节点进行测试
		fmt.Println("警告: 找不到使用节点，使用定义节点进行测试")
		usageNode = definitionNode
	}

	// 查找所有引用
	refs, err := tsmorphgo.FindReferences(*usageNode)
	if err != nil {
		log.Printf("查找引用失败: %v", err)
		return
	}

	fmt.Printf("引用位置: 找到 %d 个引用\n", len(refs))
	for i, ref := range refs {
		fmt.Printf("  引用 %d: %s (行 %d, 列 %d)\n",
			i+1, ref.GetText(), ref.GetStartLineNumber(), ref.GetStartColumnNumber())
	}

	fmt.Println()
}

// cachePerformanceExample 缓存优化性能示例
func cachePerformanceExample() {
	fmt.Println("2. 缓存优化性能示例")
	fmt.Println("==================")

	// 创建项目 - 使用内存源码方式
	sources := map[string]string{
		"/cache.ts": `
		const sharedVar = "hello";
		const anotherVar = "world";

	 function helperFunction() {
		 console.log(sharedVar);
		 console.log(anotherVar);
		 return sharedVar + " " + anotherVar;
	}

	 function mainFunction() {
		 const result = helperFunction();
		 console.log(result);
		 console.log(sharedVar);
		 return result;
	}

	 // 多次调用
	 mainFunction();
	 mainFunction();
	 helperFunction();
	 console.log(sharedVar);
	 console.log(anotherVar);
		`,
	}
	project := tsmorphgo.NewProjectFromSources(sources)
	defer project.Close()

	// 获取源文件
	sourceFile := project.GetSourceFile("/cache.ts")
	if sourceFile == nil {
		log.Fatal("无法创建测试文件")
	}

	// 找到 sharedVar 的使用节点
	var targetNode *tsmorphgo.Node
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsIdentifier(node) && node.GetText() == "sharedVar" {
			parent := node.GetParent()
			if parent != nil && parent.Kind != 214 {
				nodeCopy := node
				targetNode = &nodeCopy
			}
		}
	})

	if targetNode == nil {
		fmt.Println("警告: 找不到sharedVar的目标节点，跳过缓存性能测试")
		fmt.Println()
		return
	}

	// 第一次调用 - 应该来自LSP服务
	fmt.Println("第一次调用 (LSP服务):")
	start := time.Now()
	refs1, fromCache1, err := tsmorphgo.FindReferencesWithCache(*targetNode)
	duration1 := time.Since(start)

	if err != nil {
		log.Printf("查找失败: %v", err)
		return
	}

	fmt.Printf("  耗时: %v\n", duration1)
	fmt.Printf("  结果来源: %s\n", map[bool]string{true: "缓存", false: "LSP服务"}[fromCache1])
	fmt.Printf("  找到引用: %d 个\n", len(refs1))

	// 第二次调用 - 应该来自缓存
	fmt.Println("\n第二次调用 (缓存):")
	start = time.Now()
	refs2, fromCache2, err := tsmorphgo.FindReferencesWithCache(*targetNode)
	duration2 := time.Since(start)

	if err != nil {
		log.Printf("查找失败: %v", err)
		return
	}

	fmt.Printf("  耗时: %v\n", duration2)
	fmt.Printf("  结果来源: %s\n", map[bool]string{true: "缓存", false: "LSP服务"}[fromCache2])
	fmt.Printf("  找到引用: %d 个\n", len(refs2))

	// 计算性能提升
	if duration1 > 0 && duration2 > 0 {
		speedup := float64(duration1) / float64(duration2)
		fmt.Printf("\n性能提升: %.2fx 倍\n", speedup)
	}

	fmt.Println()
}

// errorHandlingExample 错误处理和重试示例
func errorHandlingExample() {
	fmt.Println("3. 错误处理和重试示例")
	fmt.Println("===================")

	// 创建项目 - 使用内存源码方式
	sources := map[string]string{
		"/error.ts": `
		const testVar = "error test";
		console.log(testVar);
		`,
	}
	project := tsmorphgo.NewProjectFromSources(sources)
	defer project.Close()

	// 获取源文件
	sourceFile := project.GetSourceFile("/error.ts")
	if sourceFile == nil {
		log.Fatal("无法创建测试文件")
	}

	// 找到目标节点
	var targetNode *tsmorphgo.Node
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsIdentifier(node) && node.GetText() == "testVar" {
			parent := node.GetParent()
			if parent != nil && parent.Kind != 214 {
				nodeCopy := node
				targetNode = &nodeCopy
			}
		}
	})

	if targetNode == nil {
		fmt.Println("警告: 找不到testVar的目标节点，跳过错误处理测试")
		fmt.Println()
		return
	}

	// 执行查找
	refs, err := tsmorphgo.FindReferences(*targetNode)
	if err != nil {
		fmt.Printf("查找失败: %v\n", err)
	} else {
		fmt.Printf("查找成功: %d 个引用\n", len(refs))
		for i, ref := range refs {
			fmt.Printf("  引用 %d: %s (行 %d)\n",
				i+1, ref.GetText(), ref.GetStartLineNumber())
		}
	}

	fmt.Println()
}