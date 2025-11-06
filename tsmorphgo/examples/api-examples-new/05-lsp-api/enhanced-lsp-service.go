// +build lsp-api

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/lsp"
	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

// enhanced-lsp-service.go
//
// 这个示例展示了如何使用增强的 LSP API 功能：
// 1. 基础 LSP 服务管理
// 2. 带缓存的引用查找
// 3. 错误处理和重试机制
// 4. 性能监控和指标收集
// 5. 降级策略
// 6. 配置化管理
//

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run -tags lsp-api enhanced-lsp-service.go <TypeScript项目路径> [配置文件路径]")
		os.Exit(1)
	}

	projectPath := os.Args[1]
	configPath := "configs/references_config.json"
	if len(os.Args) > 2 {
		configPath = os.Args[2]
	}

	fmt.Println("🚀 增强型 LSP API - 高级引用查找和管理")
	fmt.Println("=====================================")

	// 加载配置
	config, err := tsmorphgo.LoadReferencesConfig(configPath)
	if err != nil {
		log.Printf("加载配置失败: %v，使用默认配置", err)
		config = tsmorphgo.DefaultReferencesConfig()
	}

	fmt.Printf("📋 配置加载完成: 缓存(%t) 重试(%t) 指标(%t) 降级(%t)\n",
		config.CacheSettings.Enabled, config.RetrySettings.Enabled,
		config.PerformanceSettings.EnableMetrics, config.FallbackSettings.Enabled)

	// 1. 基础 LSP 服务设置
	setupLSPService(projectPath)

	// 2. 引用查找功能演示
	demonstrateReferenceFinding(projectPath, config)

	// 3. 性能对比测试
	performanceComparison(projectPath, config)

	// 4. 错误处理和降级演示
	errorHandlingDemo(projectPath, config)

	// 5. 批量处理演示
	batchProcessingDemo(projectPath, config)

	// 6. 配置化功能演示
	configurationDemo(projectPath, config)
}

// setupLSPService 设置基础LSP服务
func setupLSPService(projectPath string) {
	fmt.Println("\n🔧 1. LSP 服务基础设置")
	fmt.Println("======================")

	// 创建 LSP 服务
	service, err := lsp.NewService(projectPath)
	if err != nil {
		fmt.Printf("❌ LSP 服务创建失败: %v\n", err)
		return
	}
	defer service.Close()

	fmt.Printf("✅ LSP 服务创建成功\n")
	fmt.Printf("   项目路径: %s\n", projectPath)

	// 创建 TSMorphGo 项目
	tsProject := tsmorphgo.NewProject(".", &tsmorphgo.ProjectOptions{})
	sourceFile := tsProject.AddSourceFile("demo.ts", `
		const sharedVar = "shared value";
		let mutableVar = "mutable";

		interface DemoInterface {
			method(param: string): string;
		}

		class DemoClass implements DemoInterface {
			private property: string;

			constructor() {
				this.property = sharedVar;
			}

			method(param: string): string {
				console.log(sharedVar);
				console.log(param);
				return this.property + " " + param;
			}
		}

		const instance = new DemoClass();
		const result = instance.method(sharedVar);
		console.log(result);
	`)

	fmt.Printf("✅ 创建测试源文件: %s\n", sourceFile.GetFilePath())
	fmt.Printf("   文件大小: %d 字符\n", len(sourceFile.GetText()))
}

// demonstrateReferenceFinding 演示引用查找功能
func demonstrateReferenceFinding(projectPath string, config *tsmorphgo.ReferencesConfig) {
	fmt.Println("\n🔍 2. 引用查找功能演示")
	fmt.Println("======================")

	// 创建项目
	project := tsmorphgo.NewProject(".", &tsmorphgo.ProjectOptions{})

	// 添加更复杂的测试文件
	sourceFile := project.AddSourceFile("references_demo.ts", `
		// 变量声明
		const globalVar = "global";
		const reusedVar = "reused";

		// 接口定义
	 interface IDataProcessor {
			process(data: string): string;
			validate(data: string): boolean;
		}

		// 类定义
		class DataProcessor implements IDataProcessor {
			private readonly name: string;

			constructor(name: string) {
				this.name = name;
				console.log("Processor created:", this.name);
			}

			process(data: string): string {
				console.log("Processing:", data);
				return data + "_processed_" + globalVar;
			}

			validate(data: string): boolean {
				return data.length > 0 && data.includes(reusedVar);
			}
		}

		// 函数定义
		function createProcessor(name: string): IDataProcessor {
			return new DataProcessor(name);
		}

		// 使用示例
		const processor = createProcessor("main");
		const result = processor.process(reusedVar);
		const isValid = processor.validate(result);

		console.log("Result:", result);
		console.log("Valid:", isValid);
		console.log("Global:", globalVar);
	`)

	// 收集所有标识符用于测试
	var testNodes []struct {
		name string
		node *tsmorphgo.Node
	}

	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsIdentifier(node) {
			nodeText := node.GetText()
			// 选择一些代表性的标识符
			if nodeText == "globalVar" || nodeText == "reusedVar" ||
			   nodeText == "DataProcessor" || nodeText == "processor" {
				parent := node.GetParent()
				// 排除定义位置，只测试引用位置
				if parent != nil && parent.Kind != 164 { // VariableDeclaration
					nodeCopy := node
					testNodes = append(testNodes, struct {
						name string
						node *tsmorphgo.Node
					}{name: nodeText, node: &nodeCopy})
				}
			}
		}
	})

	fmt.Printf("找到 %d 个测试节点\n", len(testNodes))

	// 对每个测试节点进行引用查找
	for i, testNode := range testNodes {
		fmt.Printf("\n📍 测试节点 %d: %s\n", i+1, testNode.name)

		// 使用带缓存的引用查找
		retryConfig := config.RetrySettings.ToRetryConfig()
		refs, fromCache, err := tsmorphgo.FindReferencesWithCacheAndRetry(*testNode.node, retryConfig)

		if err != nil {
			fmt.Printf("❌ 引用查找失败: %v\n", err)

			// 尝试降级策略
			if config.FallbackSettings.Enabled {
				fallbackRefs := tsmorphgo.FindReferencesFallback(*testNode.node)
				fmt.Printf("🔄 降级策略找到 %d 个引用\n", len(fallbackRefs))
			}
		} else {
			fmt.Printf("✅ 找到 %d 个引用", len(refs))
			if fromCache {
				fmt.Printf(" (来自缓存)")
			} else {
				fmt.Printf(" (来自LSP)")
			}
			fmt.Println()

			// 显示引用详情
			for j, ref := range refs {
				fmt.Printf("   引用 %d: 行 %d, 列 %d, 内容: %s\n",
					j+1, ref.GetStartLineNumber(), ref.GetStartColumnNumber(), ref.GetText())
			}
		}
	}
}

// performanceComparison 性能对比测试
func performanceComparison(projectPath string, config *tsmorphgo.ReferencesConfig) {
	fmt.Println("\n⚡ 3. 性能对比测试")
	fmt.Println("==================")

	project := tsmorphgo.NewProject(".", &tsmorphgo.ProjectOptions{})

	// 创建测试文件
	sourceFile := project.AddSourceFile("performance_test.ts", `
		const perfVar = "performance test";

	 function perfFunction() {
		 console.log(perfVar);
		 console.log(perfVar);
		 console.log(perfVar);
		 return perfVar;
		}

	 // 多次调用
	 perfFunction();
	 perfFunction();
	 perfFunction();
	 console.log(perfVar);
	 console.log(perfVar);
	`)

	// 找到测试节点
	var targetNode *tsmorphgo.Node
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsIdentifier(node) && node.GetText() == "perfVar" {
			parent := node.GetParent()
			if parent != nil && parent.Kind != 164 {
				nodeCopy := node
				targetNode = &nodeCopy
			}
		}
	})

	if targetNode == nil {
		fmt.Println("❌ 找不到性能测试节点")
		return
	}

	// 创建指标收集器
	collector := tsmorphgo.NewMetricsCollector(project)

	// 性能测试参数
	const numTests = 5
	fmt.Printf("执行 %d 次性能测试...\n", numTests)

	var lspTotalTime time.Duration
	var cacheTotalTime time.Duration
	var lspSuccessCount, cacheSuccessCount int

	// LSP 直连测试
	fmt.Println("\n📊 LSP 直连测试:")
	for i := 0; i < numTests; i++ {
		start := time.Now()
		refs, err := tsmorphgo.FindReferences(*targetNode)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("  测试 %d: ❌ 失败 (%v)\n", i+1, err)
		} else {
			fmt.Printf("  测试 %d: ✅ 成功 (%v, %d 引用)\n", i+1, duration, len(refs))
			lspTotalTime += duration
			lspSuccessCount++
		}
	}

	// 缓存优化测试
	fmt.Println("\n🚀 缓存优化测试:")
	retryConfig := config.RetrySettings.ToRetryConfig()

	for i := 0; i < numTests; i++ {
		start := time.Now()
		refs, fromCache, err := tsmorphgo.FindReferencesWithCacheAndRetry(*targetNode, retryConfig)
		duration := time.Since(start)

		// 收集指标
		collector.FindReferencesWithMetrics(*targetNode)

		if err != nil {
			fmt.Printf("  测试 %d: ❌ 失败 (%v)\n", i+1, err)
		} else {
			source := "LSP"
			if fromCache {
				source = "缓存"
			}
			fmt.Printf("  测试 %d: ✅ 成功 (%v, %s, %d 引用)\n", i+1, duration, source, len(refs))
			cacheTotalTime += duration
			cacheSuccessCount++
		}
	}

	// 性能统计
	fmt.Println("\n📈 性能统计结果:")
	fmt.Println("==================")

	if lspSuccessCount > 0 {
		avgLSPTime := lspTotalTime / time.Duration(lspSuccessCount)
		fmt.Printf("LSP 平均响应时间: %v\n", avgLSPTime)
	}

	if cacheSuccessCount > 0 {
		avgCacheTime := cacheTotalTime / time.Duration(cacheSuccessCount)
		fmt.Printf("缓存平均响应时间: %v\n", avgCacheTime)

		if lspSuccessCount > 0 {
			speedup := float64(lspTotalTime) / float64(cacheTotalTime)
			fmt.Printf("性能提升倍数: %.2fx\n", speedup)
		}
	}

	// 显示详细指标
	metrics := collector.GetMetrics()
	fmt.Printf("\n📊 详细性能指标:\n")
	fmt.Printf("总查询次数: %d\n", metrics.TotalQueries)
	fmt.Printf("缓存命中次数: %d\n", metrics.CacheHits)
	fmt.Printf("LSP调用次数: %d\n", metrics.LSPCalls)
	fmt.Printf("缓存命中率: %.1f%%\n", metrics.HitRate())
	fmt.Printf("平均延迟: %v\n", metrics.AverageLatency)
}

// errorHandlingDemo 错误处理和降级演示
func errorHandlingDemo(projectPath string, config *tsmorphgo.ReferencesConfig) {
	fmt.Println("\n🛡️ 4. 错误处理和降级演示")
	fmt.Println("========================")

	project := tsmorphgo.NewProject(".", &tsmorphgo.ProjectOptions{})

	// 创建包含潜在问题的测试文件
	sourceFile := project.AddSourceFile("error_handling_test.ts", `
		const errorVar = "error test";
		let dynamicVar: any = undefined;

		function errorProneFunction(param: any) {
		 try {
			 console.log(errorVar);
			 if (dynamicVar.method) {
				 return dynamicVar.method(param);
			 }
			 return param + errorVar;
		 } catch (e) {
			 console.error("Error occurred:", e);
			 return "fallback";
		 }
		}

		errorProneFunction(dynamicVar);
	`)

	// 找到测试节点
	var targetNode *tsmorphgo.Node
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsIdentifier(node) && node.GetText() == "errorVar" {
			parent := node.GetParent()
			if parent != nil && parent.Kind != 164 {
				nodeCopy := node
				targetNode = &nodeCopy
			}
		}
	})

	if targetNode == nil {
		fmt.Println("❌ 找不到错误处理测试节点")
		return
	}

	// 测试不同的错误场景
	testScenarios := []struct {
		name        string
		retryConfig *tsmorphgo.RetryConfig
	}{
		{
			name:        "默认重试配置",
			retryConfig: config.RetrySettings.ToRetryConfig(),
		},
		{
			name: "快速失败配置",
			retryConfig: &tsmorphgo.RetryConfig{
				MaxRetries:    1,
				BaseDelay:     10 * time.Millisecond,
				MaxDelay:      50 * time.Millisecond,
				BackoffFactor: 1.2,
				Enabled:       true,
			},
		},
		{
			name: "禁用重试配置",
			retryConfig: &tsmorphgo.RetryConfig{
				Enabled: false,
			},
		},
	}

	for _, scenario := range testScenarios {
		fmt.Printf("\n🔬 测试场景: %s\n", scenario.name)

		// 执行引用查找
		refs, fromCache, err := tsmorphgo.FindReferencesWithCacheAndRetry(*targetNode, scenario.retryConfig)

		if err != nil {
			fmt.Printf("❌ 引用查找失败: %v\n", err)

			// 分析错误类型
			if refErr, ok := err.(*tsmorphgo.ReferenceError); ok {
				fmt.Printf("   错误类型: %s\n", refErr.Error())
				fmt.Printf("   可重试: %t\n", refErr.Retryable)
				fmt.Printf("   重试次数: %d\n", refErr.RetryCount)
				fmt.Printf("   文件: %s:%d\n", refErr.FilePath, refErr.LineNumber)

				// 测试降级策略
				if refErr.ShouldUseFallback() && config.FallbackSettings.Enabled {
					fmt.Printf("🔄 启用降级策略...\n")
					fallbackRefs := tsmorphgo.FindReferencesFallback(*targetNode)
					fmt.Printf("   降级策略找到 %d 个引用\n", len(fallbackRefs))

					if len(fallbackRefs) > 0 {
						fmt.Printf("   降级成功，引用位置:\n")
						for i, ref := range fallbackRefs {
							fmt.Printf("     %d. 行 %d: %s\n", i+1, ref.GetStartLineNumber(), ref.GetText())
						}
					}
				}
			}
		} else {
			fmt.Printf("✅ 查找成功: %d 个引用", len(refs))
			if fromCache {
				fmt.Printf(" (来自缓存)")
			} else {
				fmt.Printf(" (来自LSP)")
			}
			fmt.Println()
		}
	}

	// 测试上下文分析
	fmt.Println("\n🧠 上下文分析测试:")
	testContextAnalysis(sourceFile)
}

// testContextAnalysis 测试上下文分析功能
func testContextAnalysis(sourceFile *tsmorphgo.SourceFile) {
	var definitionNodes, referenceNodes []*tsmorphgo.Node

	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if !tsmorphgo.IsIdentifier(node) {
			return
		}

		nodeText := node.GetText()
		// 只分析关键标识符
		if nodeText == "errorVar" || nodeText == "errorProneFunction" || nodeText == "dynamicVar" {
			nodeCopy := node

			if tsmorphgo.IsLikelyDefinition(node) {
				definitionNodes = append(definitionNodes, &nodeCopy)
			} else if tsmorphgo.IsLikelyReference(node) {
				referenceNodes = append(referenceNodes, &nodeCopy)
			}
		}
	})

	fmt.Printf("上下文分析结果:\n")
	fmt.Printf("  定义节点: %d\n", len(definitionNodes))
	for i, def := range definitionNodes {
		fmt.Printf("    %d. %s (行 %d)\n", i+1, def.GetText(), def.GetStartLineNumber())
	}

	fmt.Printf("  引用节点: %d\n", len(referenceNodes))
	for i, ref := range referenceNodes {
		fmt.Printf("    %d. %s (行 %d)\n", i+1, ref.GetText(), ref.GetStartLineNumber())
	}
}

// batchProcessingDemo 批量处理演示
func batchProcessingDemo(projectPath string, config *tsmorphgo.ReferencesConfig) {
	fmt.Println("\n📦 5. 批量处理演示")
	fmt.Println("==================")

	project := tsmorphgo.NewProject(".", &tsmorphgo.ProjectOptions{})

	// 创建包含多个变量的测试文件
	sourceFile := project.AddSourceFile("batch_test.ts", `
		const batchVar1 = "batch1";
		const batchVar2 = "batch2";
		const batchVar3 = "batch3";

	 function batchFunction() {
		 console.log(batchVar1);
		 console.log(batchVar2);
		 console.log(batchVar3);
		 return batchVar1 + batchVar2 + batchVar3;
		}

	 const result1 = batchFunction();
	 const result2 = batchFunction();
	 console.log(batchVar1);
	 console.log(batchVar2);
	`)

	// 收集所有变量节点
	var nodes []tsmorphgo.Node
	targetVars := []string{"batchVar1", "batchVar2", "batchVar3"}

	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsIdentifier(node) {
			nodeText := node.GetText()
			for _, target := range targetVars {
				if nodeText == target {
					parent := node.GetParent()
					if parent != nil && parent.Kind != 164 { // 不是变量声明
						nodes = append(nodes, node)
					}
					break
				}
			}
		}
	})

	fmt.Printf("收集到 %d 个节点进行批量处理\n", len(nodes))

	if len(nodes) == 0 {
		fmt.Println("❌ 找不到批量处理的节点")
		return
	}

	// 批量处理测试
	fmt.Println("\n🔄 执行批量处理...")
	start := time.Now()
	results, err := tsmorphgo.FindReferencesBatch(nodes)
	batchDuration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ 批量处理失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 批量处理完成，耗时: %v\n", batchDuration)
	fmt.Printf("结果数量: %d\n", len(results))

	// 显示批量处理结果
	totalRefs := 0
	for cacheKey, refs := range results {
		fmt.Printf("  节点 %s: %d 个引用\n", cacheKey, len(refs))
		totalRefs += len(refs)

		// 限制显示的引用数量
		maxDisplay := 3
		for i, ref := range refs {
			if i >= maxDisplay {
				fmt.Printf("    ... 还有 %d 个引用\n", len(refs)-maxDisplay)
				break
			}
			fmt.Printf("    %d. 行 %d: %s\n", i+1, ref.GetStartLineNumber(), ref.GetText())
		}
	}

	fmt.Printf("总计找到 %d 个引用\n", totalRefs)

	// 与单独处理对比
	fmt.Println("\n⏱️  单独处理对比...")
	start = time.Now()
	singleTotalRefs := 0
	successCount := 0

	for _, node := range nodes {
		refs, err := tsmorphgo.FindReferences(node)
		if err == nil {
			singleTotalRefs += len(refs)
			successCount++
		}
	}
	singleDuration := time.Since(start)

	fmt.Printf("单独处理耗时: %v\n", singleDuration)
	fmt.Printf("单独处理成功: %d/%d\n", successCount, len(nodes))
	fmt.Printf("单独处理引用: %d\n", singleTotalRefs)

	// 计算性能提升
	if batchDuration > 0 && singleDuration > 0 {
		improvement := float64(singleDuration) / float64(batchDuration)
		fmt.Printf("\n🚀 批量处理性能提升: %.2fx\n", improvement)

		// 验证结果一致性
		if totalRefs == singleTotalRefs {
			fmt.Printf("✅ 结果一致性验证通过\n")
		} else {
			fmt.Printf("⚠️  结果数量不一致 (批量: %d, 单独: %d)\n", totalRefs, singleTotalRefs)
		}
	}
}

// configurationDemo 配置化功能演示
func configurationDemo(projectPath string, config *tsmorphgo.ReferencesConfig) {
	fmt.Println("\n⚙️ 6. 配置化功能演示")
	fmt.Println("==================")

	// 显示当前配置
	fmt.Printf("当前配置详情:\n")
	fmt.Printf("缓存: 启用=%t, 最大=%d, TTL=%s\n",
		config.CacheSettings.Enabled, config.CacheSettings.MaxEntries, config.CacheSettings.TTL)
	fmt.Printf("重试: 启用=%t, 最大=%d, 延迟=%s\n",
		config.RetrySettings.Enabled, config.RetrySettings.MaxRetries, config.RetrySettings.BaseDelay)
	fmt.Printf("性能: 指标=%t, 批量=%t, 超时=%s\n",
		config.PerformanceSettings.EnableMetrics, config.PerformanceSettings.EnableBatching, config.PerformanceSettings.Timeout)
	fmt.Printf("降级: 启用=%t, 上下文分析=%t\n",
		config.FallbackSettings.Enabled, config.FallbackSettings.EnableContextAnalysis)
	fmt.Printf("日志: 启用=%t, 级别=%s\n",
		config.LoggingSettings.Enabled, config.LoggingSettings.Level)

	// 测试不同配置的效果
	fmt.Println("\n🧪 配置效果测试:")

	// 测试1: 禁用缓存
	fmt.Println("\n📋 测试1: 禁用缓存")
	config1 := config.Clone()
	config1.CacheSettings.Enabled = false

	project := tsmorphgo.NewProject(".", &tsmorphgo.ProjectOptions{})
	sourceFile := project.AddSourceFile("config_test.ts", `
		const configVar = "config test";
		console.log(configVar);
	`)

	var targetNode *tsmorphgo.Node
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsIdentifier(node) && node.GetText() == "configVar" {
			parent := node.GetParent()
			if parent != nil && parent.Kind != 164 {
				nodeCopy := node
				targetNode = &nodeCopy
			}
		}
	})

	if targetNode != nil {
		retryConfig := config1.RetrySettings.ToRetryConfig()
		refs, fromCache, err := tsmorphgo.FindReferencesWithCacheAndRetry(*targetNode, retryConfig)
		if err != nil {
			fmt.Printf("  禁用缓存测试失败: %v\n", err)
		} else {
			fmt.Printf("  禁用缓存: %d 引用, 来源: %s\n", len(refs), map[bool]string{true: "缓存", false: "LSP"}[fromCache])
		}
	}

	// 测试2: 优化重试配置
	fmt.Println("\n📋 测试2: 优化重试配置")
	config2 := config.Clone()
	config2.RetrySettings.MaxRetries = 1
	config2.RetrySettings.BaseDelay = "50ms"

	if targetNode != nil {
		retryConfig := config2.RetrySettings.ToRetryConfig()
		refs, fromCache, err := tsmorphgo.FindReferencesWithCacheAndRetry(*targetNode, retryConfig)
		if err != nil {
			fmt.Printf("  优化重试测试失败: %v\n", err)
		} else {
			fmt.Printf("  优化重试: %d 引用, 来源: %s\n", len(refs), map[bool]string{true: "缓存", false: "LSP"}[fromCache])
		}
	}

	// 配置验证
	fmt.Println("\n✅ 配置验证:")
	err = config.Validate()
	if err != nil {
		fmt.Printf("❌ 配置验证失败: %v\n", err)
	} else {
		fmt.Printf("✅ 当前配置验证通过\n")
	}

	// 显示项目缓存统计
	cacheStats := project.GetCacheStats()
	fmt.Printf("\n📊 项目缓存统计:\n")
	fmt.Printf("条目数: %d\n", cacheStats.TotalEntries)
	fmt.Printf("访问次数: %d\n", cacheStats.TotalAccesses)
	fmt.Printf("过期条目: %d\n", cacheStats.ExpiredEntries)
	fmt.Printf("最大条目: %d\n", cacheStats.MaxEntries)
	fmt.Printf("TTL: %v\n", cacheStats.TTL)

	fmt.Println("\n🎉 增强型 LSP API 演示完成!")
	fmt.Println("================================")
	fmt.Println("✅ 已演示的功能:")
	fmt.Println("   - 🔧 基础 LSP 服务设置")
	fmt.Println("   - 🔍 增强引用查找")
	fmt.Println("   - ⚡ 性能优化和缓存")
	fmt.Println("   - 🛡️ 错误处理和重试")
	fmt.Println("   - 🔄 降级策略")
	fmt.Println("   - 📦 批量处理")
	fmt.Println("   - ⚙️ 配置化管理")
	fmt.Println("   - 📊 性能监控和指标")
	fmt.Println("================================")
	fmt.Println("🚀 TSMorphGo References API 已准备就绪!")
}