//go:build references_usage_example
// +build references_usage_example

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

// references_usage_example.go
//
// 这个示例展示了如何使用 TSMorphGo References API 的各种功能：
// 1. 基础引用查找
// 2. 带缓存的优化查找
// 3. 错误处理和重试
// 4. 性能监控
// 5. 批量处理
// 6. 降级策略
//

func main() {
	fmt.Println("=== TSMorphGo References API 使用示例 ===\n")

	// 示例1: 基础引用查找
	basicReferenceExample()

	// 示例2: 缓存优化性能
	cachePerformanceExample()

	// 示例3: 错误处理和重试
	errorHandlingExample()

	// 示例4: 批量处理
	batchProcessingExample()

	// 示例5: 性能监控
	performanceMonitoringExample()

	// 示例6: 降级策略
	fallbackStrategyExample()
}

// basicReferenceExample 基础引用查找示例
func basicReferenceExample() {
	fmt.Println("1. 基础引用查找示例")
	fmt.Println("==================")

	// 创建项目
	project := tsmorphgo.NewProject(".", &tsmorphgo.ProjectOptions{})

	// 添加源文件
	sourceFile := project.AddSourceFile("basic.ts", `
		const sharedVar = "hello world";

		function testFunction() {
			console.log(sharedVar);
			let localVar = sharedVar;
			return localVar;
		}

		testFunction();
		console.log(sharedVar);
	`)

	// 找到 sharedVar 的定义节点
	var definitionNode *tsmorphgo.Node
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsIdentifier(node) && node.GetText() == "sharedVar" {
			parent := node.GetParent()
			if parent != nil && parent.Kind == 164 { // VariableDeclaration
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
			if parent != nil && parent.Kind != 164 { // 不是变量声明
				nodeCopy := node
				usageNode = &nodeCopy
			}
		}
	})

	if usageNode == nil {
		log.Fatal("找不到 sharedVar 的使用")
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

	project := tsmorphgo.NewProject(".", &tsmorphgo.ProjectOptions{})

	// 添加复杂的源文件
	sourceFile := project.AddSourceFile("cache.ts", `
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
	`)

	// 找到 sharedVar 的使用节点
	var targetNode *tsmorphgo.Node
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsIdentifier(node) && node.GetText() == "sharedVar" {
			parent := node.GetParent()
			if parent != nil && parent.Kind != 164 {
				nodeCopy := node
				targetNode = &nodeCopy
			}
		}
	})

	if targetNode == nil {
		log.Fatal("找不到目标节点")
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

	// 第三次调用 - 验证缓存一致性
	fmt.Println("\n第三次调用 (验证缓存一致性):")
	start = time.Now()
	refs3, fromCache3, err := tsmorphgo.FindReferencesWithCache(*targetNode)
	duration3 := time.Since(start)

	if err != nil {
		log.Printf("查找失败: %v", err)
		return
	}

	fmt.Printf("  耗时: %v\n", duration3)
	fmt.Printf("  结果来源: %s\n", map[bool]string{true: "缓存", false: "LSP服务"}[fromCache3])
	fmt.Printf("  结果一致: %t\n", len(refs1) == len(refs2) && len(refs2) == len(refs3))

	fmt.Println()
}

// errorHandlingExample 错误处理和重试示例
func errorHandlingExample() {
	fmt.Println("3. 错误处理和重试示例")
	fmt.Println("===================")

	project := tsmorphgo.NewProject(".", &tsmorphgo.ProjectOptions{})

	// 添加测试文件
	sourceFile := project.AddSourceFile("error.ts", `
		const testVar = "error test";
		console.log(testVar);
	`)

	// 找到目标节点
	var targetNode *tsmorphgo.Node
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsIdentifier(node) && node.GetText() == "testVar" {
			parent := node.GetParent()
			if parent != nil && parent.Kind != 164 {
				nodeCopy := node
				targetNode = &nodeCopy
			}
		}
	})

	if targetNode == nil {
		log.Fatal("找不到目标节点")
	}

	// 创建自定义重试配置 - 快速失败用于演示
	fastRetryConfig := &tsmorphgo.RetryConfig{
		MaxRetries:    2,
		BaseDelay:     10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 1.5,
		Enabled:       true,
	}

	fmt.Println("使用快速重试配置进行测试...")

	// 执行带重试的查找
	refs, fromCache, err := tsmorphgo.FindReferencesWithCacheAndRetry(*targetNode, fastRetryConfig)

	if err != nil {
		fmt.Printf("查找失败: %v\n", err)

		// 分析错误类型
		if refErr, ok := err.(*tsmorphgo.ReferenceError); ok {
			fmt.Printf("错误类型: %s\n", refErr.Error())
			fmt.Printf("可重试: %t\n", refErr.Retryable)
			fmt.Printf("重试次数: %d\n", refErr.RetryCount)
			fmt.Printf("文件路径: %s\n", refErr.FilePath)
			fmt.Printf("行号: %d\n", refErr.LineNumber)

			// 检查是否应该使用降级策略
			if refErr.ShouldUseFallback() {
				fmt.Println("尝试降级策略...")
				fallbackRefs := tsmorphgo.FindReferencesFallback(*targetNode)
				fmt.Printf("降级策略找到 %d 个引用\n", len(fallbackRefs))

				if len(fallbackRefs) > 0 {
					for i, ref := range fallbackRefs {
						fmt.Printf("  降级引用 %d: %s (行 %d)\n",
							i+1, ref.GetText(), ref.GetStartLineNumber())
					}
				}
			}
		}
	} else {
		fmt.Printf("查找成功: %d 个引用, 来自缓存: %t\n", len(refs), fromCache)
		for i, ref := range refs {
			fmt.Printf("  引用 %d: %s (行 %d)\n",
				i+1, ref.GetText(), ref.GetStartLineNumber())
		}
	}

	fmt.Println()
}

// batchProcessingExample 批量处理示例
func batchProcessingExample() {
	fmt.Println("4. 批量处理示例")
	fmt.Println("================")

	project := tsmorphgo.NewProject(".", &tsmorphgo.ProjectOptions{})

	// 添加包含多个变量的文件
	sourceFile := project.AddSourceFile("batch.ts", `
		const var1 = "value1";
		const var2 = "value2";
		const var3 = "value3";

		function testFunction() {
		 console.log(var1);
		 console.log(var2);
		 return var3;
		}

		testFunction();
		console.log(var1);
		console.log(var3);
	`)

	// 收集所有变量标识符
	var nodes []tsmorphgo.Node
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsIdentifier(node) {
			nodeText := node.GetText()
			if nodeText == "var1" || nodeText == "var2" || nodeText == "var3" {
				parent := node.GetParent()
				if parent != nil && parent.Kind != 164 { // 不是变量声明
					nodes = append(nodes, node)
				}
			}
		}
	})

	fmt.Printf("收集到 %d 个节点进行批量查找\n", len(nodes))

	// 执行批量查找
	fmt.Println("执行批量查找...")
	start := time.Now()
	results, err := tsmorphgo.FindReferencesBatch(nodes)
	duration := time.Since(start)

	if err != nil {
		log.Printf("批量查找失败: %v", err)
		return
	}

	fmt.Printf("批量查找完成，耗时: %v\n", duration)
	fmt.Printf("结果数量: %d\n", len(results))

	// 显示结果
	for cacheKey, refs := range results {
		fmt.Printf("  节点 %s: 找到 %d 个引用\n", cacheKey, len(refs))
		for i, ref := range refs {
			fmt.Printf("    引用 %d: %s (行 %d)\n",
				i+1, ref.GetText(), ref.GetStartLineNumber())
		}
	}

	// 比较单独查找的性能
	fmt.Println("\n比较单独查找的性能...")
	start = time.Now()
	singleResults := 0
	for _, node := range nodes {
		refs, err := tsmorphgo.FindReferences(node)
		if err == nil {
			singleResults += len(refs)
		}
	}
	singleDuration := time.Since(start)

	fmt.Printf("单独查找耗时: %v, 总引用数: %d\n", singleDuration, singleResults)

	if duration > 0 && singleDuration > 0 {
		improvement := float64(singleDuration) / float64(duration)
		fmt.Printf("批量处理性能提升: %.2fx\n", improvement)
	}

	fmt.Println()
}

// performanceMonitoringExample 性能监控示例
func performanceMonitoringExample() {
	fmt.Println("5. 性能监控示例")
	fmt.Println("=================")

	project := tsmorphgo.NewProject(".", &tsmorphgo.ProjectOptions{})

	// 添加测试文件
	sourceFile := project.AddSourceFile("monitoring.ts", `
		const monitoredVar = "monitor test";

	 function monitoredFunction() {
		 console.log(monitoredVar);
		 return monitoredVar + " processed";
		}

	 // 多次调用
	 for(let i = 0; i < 3; i++) {
		 monitoredFunction();
		 console.log(monitoredVar);
	 }
	`)

	// 找到目标节点
	var targetNode *tsmorphgo.Node
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsIdentifier(node) && node.GetText() == "monitoredVar" {
			parent := node.GetParent()
			if parent != nil && parent.Kind != 164 {
				nodeCopy := node
				targetNode = &nodeCopy
			}
		}
	})

	if targetNode == nil {
		log.Fatal("找不到目标节点")
	}

	// 创建指标收集器
	collector := tsmorphgo.NewMetricsCollector(project)

	// 执行多次查询并收集指标
	fmt.Println("执行多次查询并收集性能指标...")
	const numQueries = 5

	for i := 0; i < numQueries; i++ {
		fmt.Printf("查询 %d/%d\n", i+1, numQueries)

		refs, err := collector.FindReferencesWithMetrics(*targetNode)
		if err != nil {
			log.Printf("查询 %d 失败: %v", i+1, err)
			continue
		}

		fmt.Printf("  找到 %d 个引用\n", len(refs))

		// 获取实时指标
		metrics := collector.GetMetrics()
		fmt.Printf("  当前缓存命中率: %.1f%%\n", metrics.HitRate())
	}

	// 显示最终统计
	fmt.Println("\n=== 最终性能统计 ===")
	metrics := collector.GetMetrics()

	fmt.Printf("总查询次数: %d\n", metrics.TotalQueries)
	fmt.Printf("缓存命中次数: %d\n", metrics.CacheHits)
	fmt.Printf("缓存未命中次数: %d\n", metrics.CacheMisses)
	fmt.Printf("LSP调用次数: %d\n", metrics.LSPCalls)
	fmt.Printf("缓存命中率: %.1f%%\n", metrics.HitRate())
	fmt.Printf("平均延迟: %v\n", metrics.AverageLatency)
	fmt.Printf("总耗时: %v\n", metrics.TotalDuration)

	// 显示指标字符串表示
	fmt.Printf("\n指标摘要: %s\n", metrics.String())

	// 重置指标并再次测试
	fmt.Println("\n重置指标...")
	collector.ResetMetrics()

	// 执行一次查询
	_, err := collector.FindReferencesWithMetrics(*targetNode)
	if err != nil {
		log.Printf("重置后查询失败: %v", err)
	} else {
		resetMetrics := collector.GetMetrics()
		fmt.Printf("重置后查询 - 总查询: %d, 缓存命中: %d, 命中率: %.1f%%\n",
			resetMetrics.TotalQueries, resetMetrics.CacheHits, resetMetrics.HitRate())
	}

	fmt.Println()
}

// fallbackStrategyExample 降级策略示例
func fallbackStrategyExample() {
	fmt.Println("6. 降级策略示例")
	fmt.Println("================")

	project := tsmorphgo.NewProject(".", &tsmorphgo.ProjectOptions{})

	// 添加测试文件
	sourceFile := project.AddSourceFile("fallback.ts", `
		const fallbackVar = "fallback test";
		let mutableVar = "mutable";

	 function fallbackFunction(param: string) {
		 console.log(fallbackVar);
		 console.log(param);
		 return mutableVar;
		}

	 const result = fallbackFunction(fallbackVar);
	 console.log(result);
	`)

	// 测试上下文分析
	fmt.Println("测试上下文分析功能...")

	var definitionNodes, referenceNodes []*tsmorphgo.Node
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if !tsmorphgo.IsIdentifier(node) {
			return
		}

		nodeText := node.GetText()
		if nodeText != "fallbackVar" && nodeText != "mutableVar" && nodeText != "param" && nodeText != "result" {
			return
		}

		nodeCopy := node

		if tsmorphgo.IsLikelyDefinition(node) {
			definitionNodes = append(definitionNodes, &nodeCopy)
		} else if tsmorphgo.IsLikelyReference(node) {
			referenceNodes = append(referenceNodes, &nodeCopy)
		}
	})

	fmt.Printf("上下文分析结果:\n")
	fmt.Printf("  潜在定义节点: %d\n", len(definitionNodes))
	for i, def := range definitionNodes {
		fmt.Printf("    定义 %d: %s (行 %d)\n", i+1, def.GetText(), def.GetStartLineNumber())
	}

	fmt.Printf("  潜在引用节点: %d\n", len(referenceNodes))
	for i, ref := range referenceNodes {
		fmt.Printf("    引用 %d: %s (行 %d)\n", i+1, ref.GetText(), ref.GetStartLineNumber())
	}

	// 测试降级策略
	fmt.Println("\n测试降级策略...")

	// 找到测试节点
	var testNode *tsmorphgo.Node
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsIdentifier(node) && node.GetText() == "fallbackVar" {
			parent := node.GetParent()
			if parent != nil && parent.Kind != 164 {
				nodeCopy := node
				testNode = &nodeCopy
			}
		}
	})

	if testNode == nil {
		log.Fatal("找不到测试节点")
	}

	// 使用降级策略查找引用
	fmt.Println("使用降级策略查找引用...")
	fallbackRefs := tsmorphgo.FindReferencesFallback(*testNode)
	fmt.Printf("降级策略找到 %d 个引用\n", len(fallbackRefs))

	for i, ref := range fallbackRefs {
		fmt.Printf("  降级引用 %d: %s (行 %d, 列 %d)\n",
			i+1, ref.GetText(), ref.GetStartLineNumber(), ref.GetStartColumnNumber())
	}

	// 使用降级策略查找定义
	fmt.Println("\n使用降级策略查找定义...")
	fallbackDefs := tsmorphgo.GotoDefinitionFallback(*testNode)
	fmt.Printf("降级策略找到 %d 个定义\n", len(fallbackDefs))

	for i, def := range fallbackDefs {
		fmt.Printf("  降级定义 %d: %s (行 %d, 列 %d)\n",
			i+1, def.GetText(), def.GetStartLineNumber(), def.GetStartColumnNumber())
	}

	// 比较LSP和降级策略的结果
	fmt.Println("\n比较LSP和降级策略结果...")

	// LSP查找
	lspRefs, err := tsmorphgo.FindReferences(*testNode)
	if err != nil {
		fmt.Printf("LSP查找失败: %v\n", err)
		fmt.Printf("降级策略作为备用方案提供了 %d 个结果\n", len(fallbackRefs))
	} else {
		fmt.Printf("LSP找到 %d 个引用, 降级策略找到 %d 个引用\n", len(lspRefs), len(fallbackRefs))

		// 检查结果一致性
		if len(lspRefs) == len(fallbackRefs) {
			fmt.Println("LSP和降级策略结果数量一致")
		} else {
			fmt.Printf("结果数量不一致，降级策略准确率: %.1f%%\n",
				float64(len(fallbackRefs))/float64(len(lspRefs))*100)
		}
	}

	fmt.Println()
}
