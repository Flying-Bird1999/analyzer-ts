//go:build config_usage_example
// +build config_usage_example

package main

import (
	"fmt"
	"log"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

// config_usage_example.go
//
// 这个示例展示了如何使用配置化的引用查找功能：
// 1. 加载默认配置
// 2. 从文件加载配置
// 3. 自定义配置
// 4. 配置验证
// 5. 配置应用
//

func main() {
	fmt.Println("=== TSMorphGo 配置化使用示例 ===\n")

	// 示例1: 使用默认配置
	defaultConfigExample()

	// 示例2: 从文件加载配置
	loadConfigFromFileExample()

	// 示例3: 自定义配置
	customConfigExample()

	// 示例4: 配置验证
	configValidationExample()

	// 示例5: 在项目中使用配置
	projectConfigExample()
}

// defaultConfigExample 默认配置示例
func defaultConfigExample() {
	fmt.Println("1. 默认配置示例")
	fmt.Println("================")

	// 获取默认配置
	config := tsmorphgo.DefaultReferencesConfig()

	fmt.Printf("缓存设置:\n")
	fmt.Printf("  启用缓存: %t\n", config.CacheSettings.Enabled)
	fmt.Printf("  最大条目数: %d\n", config.CacheSettings.MaxEntries)
	fmt.Printf("  TTL: %s\n", config.CacheSettings.TTL)
	fmt.Printf("  文件哈希检查: %t\n", config.CacheSettings.EnableFileHashCheck)

	fmt.Printf("\n重试设置:\n")
	fmt.Printf("  启用重试: %t\n", config.RetrySettings.Enabled)
	fmt.Printf("  最大重试次数: %d\n", config.RetrySettings.MaxRetries)
	fmt.Printf("  基础延迟: %s\n", config.RetrySettings.BaseDelay)
	fmt.Printf("  最大延迟: %s\n", config.RetrySettings.MaxDelay)
	fmt.Printf("  退避因子: %.1f\n", config.RetrySettings.BackoffFactor)

	fmt.Printf("\n性能设置:\n")
	fmt.Printf("  启用指标: %t\n", config.PerformanceSettings.EnableMetrics)
	fmt.Printf("  批量处理: %t\n", config.PerformanceSettings.EnableBatching)
	fmt.Printf("  批量大小: %d\n", config.PerformanceSettings.BatchSize)
	fmt.Printf("  超时时间: %s\n", config.PerformanceSettings.Timeout)

	fmt.Printf("\n降级设置:\n")
	fmt.Printf("  启用降级: %t\n", config.FallbackSettings.Enabled)
	fmt.Printf("  上下文分析: %t\n", config.FallbackSettings.EnableContextAnalysis)
	fmt.Printf("  降级超时: %s\n", config.FallbackSettings.FallbackTimeout)

	fmt.Printf("\n日志设置:\n")
	fmt.Printf("  启用日志: %t\n", config.LoggingSettings.Enabled)
	fmt.Printf("  日志级别: %s\n", config.LoggingSettings.Level)
	fmt.Printf("  记录缓存操作: %t\n", config.LoggingSettings.LogCacheOperations)
	fmt.Printf("  记录重试尝试: %t\n", config.LoggingSettings.LogRetryAttempts)

	fmt.Println()
}

// loadConfigFromFileExample 从文件加载配置示例
func loadConfigFromFileExample() {
	fmt.Println("2. 从文件加载配置示例")
	fmt.Println("====================")

	configPath := "configs/references_config.json"

	// 从文件加载配置
	config, err := tsmorphgo.LoadReferencesConfig(configPath)
	if err != nil {
		log.Printf("加载配置失败: %v", err)
		fmt.Println("使用默认配置代替")
		config = tsmorphgo.DefaultReferencesConfig()
	} else {
		fmt.Printf("成功从 %s 加载配置\n", configPath)
	}

	// 显示加载的配置中的特定设置
	fmt.Printf("\n加载的配置详情:\n")
	fmt.Printf("  缓存TTL: %s (解析后: %v)\n",
		config.CacheSettings.TTL, config.CacheSettings.ToCacheTTL())
	fmt.Printf("  重试基础延迟: %s (解析后: %v)\n",
		config.RetrySettings.BaseDelay, config.RetrySettings.ToRetryConfig().BaseDelay)
	fmt.Printf("  性能超时: %s (解析后: %v)\n",
		config.PerformanceSettings.Timeout, config.PerformanceSettings.ToPerformanceTimeout())

	// 检查可重试的错误类型
	fmt.Printf("  可重试的错误类型: %v\n", config.RetrySettings.RetryableErrors)
	fmt.Printf("  'timeout' 是否可重试: %t\n", config.RetrySettings.IsRetryableError("timeout"))
	fmt.Printf("  'syntax' 是否可重试: %t\n", config.RetrySettings.IsRetryableError("syntax"))

	fmt.Println()
}

// customConfigExample 自定义配置示例
func customConfigExample() {
	fmt.Println("3. 自定义配置示例")
	fmt.Println("==================")

	// 创建自定义配置
	config := tsmorphgo.DefaultReferencesConfig()

	// 修改缓存设置
	config.CacheSettings.MaxEntries = 2000
	config.CacheSettings.TTL = "30m"
	config.CacheSettings.CleanupInterval = "10m"

	// 修改重试设置
	config.RetrySettings.MaxRetries = 5
	config.RetrySettings.BaseDelay = "200ms"
	config.RetrySettings.MaxDelay = "10s"
	config.RetrySettings.BackoffFactor = 1.5

	// 添加自定义可重试错误类型
	config.RetrySettings.RetryableErrors = append(config.RetrySettings.RetryableErrors,
		"temp_error", "rate_limit")

	// 修改性能设置
	config.PerformanceSettings.BatchSize = 100
	config.PerformanceSettings.Timeout = "60s"

	// 修改日志设置
	config.LoggingSettings.Level = "debug"
	config.LoggingSettings.LogCacheOperations = true
	config.LoggingSettings.LogRetryAttempts = true

	fmt.Printf("自定义配置:\n")
	fmt.Printf("  缓存条目数: %d (默认: %d)\n", config.CacheSettings.MaxEntries, 1000)
	fmt.Printf("  缓存TTL: %s (默认: 10m)\n", config.CacheSettings.TTL)
	fmt.Printf("  重试次数: %d (默认: 3)\n", config.RetrySettings.MaxRetries)
	fmt.Printf("  批量大小: %d (默认: 50)\n", config.PerformanceSettings.BatchSize)
	fmt.Printf("  日志级别: %s (默认: info)\n", config.LoggingSettings.Level)
	fmt.Printf("  日志级别数值: %d\n", config.LoggingSettings.GetLogLevelInt())

	// 保存自定义配置
	customConfigPath := "configs/custom_references_config.json"
	err := config.SaveReferencesConfig(customConfigPath)
	if err != nil {
		log.Printf("保存配置失败: %v", err)
	} else {
		fmt.Printf("自定义配置已保存到: %s\n", customConfigPath)
	}

	fmt.Println()
}

// configValidationExample 配置验证示例
func configValidationExample() {
	fmt.Println("4. 配置验证示例")
	fmt.Println("================")

	// 创建有效配置
	validConfig := tsmorphgo.DefaultReferencesConfig()
	err := validConfig.Validate()
	if err != nil {
		fmt.Printf("有效配置验证失败: %v\n", err)
	} else {
		fmt.Println("有效配置验证通过")
	}

	// 创建无效配置示例
	invalidConfigs := []struct {
		name   string
		config func() *tsmorphgo.ReferencesConfig
	}{
		{
			name: "负数缓存条目",
			config: func() *tsmorphgo.ReferencesConfig {
				c := tsmorphgo.DefaultReferencesConfig()
				c.CacheSettings.MaxEntries = -1
				return c
			},
		},
		{
			name: "无效TTL格式",
			config: func() *tsmorphgo.ReferencesConfig {
				c := tsmorphgo.DefaultReferencesConfig()
				c.CacheSettings.TTL = "invalid"
				return c
			},
		},
		{
			name: "退避因子过小",
			config: func() *tsmorphgo.ReferencesConfig {
				c := tsmorphgo.DefaultReferencesConfig()
				c.RetrySettings.BackoffFactor = 0.5
				return c
			},
		},
		{
			name: "无效日志级别",
			config: func() *tsmorphgo.ReferencesConfig {
				c := tsmorphgo.DefaultReferencesConfig()
				c.LoggingSettings.Level = "invalid"
				return c
			},
		},
	}

	for _, tc := range invalidConfigs {
		fmt.Printf("\n测试 %s:\n", tc.name)
		config := tc.config()
		err := config.Validate()
		if err != nil {
			fmt.Printf("  验证失败: %v\n", err)
		} else {
			fmt.Printf("  验证通过 (不应该发生)\n")
		}
	}

	fmt.Println()
}

// projectConfigExample 项目配置使用示例
func projectConfigExample() {
	fmt.Println("5. 项目配置使用示例")
	fmt.Println("===================")

	// 加载配置
	config, err := tsmorphgo.LoadReferencesConfig("configs/references_config.json")
	if err != nil {
		log.Printf("加载配置失败: %v", err)
		config = tsmorphgo.DefaultReferencesConfig()
	}

	// 创建项目
	project := tsmorphgo.NewProject(".", &tsmorphgo.ProjectOptions{})

	// 添加测试文件
	sourceFile := project.AddSourceFile("config_example.ts", `
		const configVar = "config test";

	 function configFunction() {
		 console.log(configVar);
		 return configVar + " processed";
		}

	 configFunction();
	 console.log(configVar);
	`)

	// 找到测试节点
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

	if targetNode == nil {
		log.Fatal("找不到目标节点")
	}

	// 根据配置执行查找
	fmt.Printf("使用配置进行引用查找:\n")
	fmt.Printf("  缓存启用: %t\n", config.CacheSettings.Enabled)
	fmt.Printf("  重试启用: %t\n", config.RetrySettings.Enabled)
	fmt.Printf("  性能指标启用: %t\n", config.PerformanceSettings.EnableMetrics)
	fmt.Printf("  降级策略启用: %t\n", config.FallbackSettings.Enabled)

	// 执行查找（实际实现中会根据配置选择不同的API）
	if config.CacheSettings.Enabled {
		fmt.Println("\n使用缓存查找...")
		retryConfig := config.RetrySettings.ToRetryConfig()
		refs, fromCache, err := tsmorphgo.FindReferencesWithCacheAndRetry(*targetNode, retryConfig)

		if err != nil {
			fmt.Printf("查找失败: %v\n", err)

			// 如果启用了降级策略
			if config.FallbackSettings.Enabled {
				fmt.Println("尝试降级策略...")
				fallbackRefs := tsmorphgo.FindReferencesFallback(*targetNode)
				fmt.Printf("降级策略找到 %d 个引用\n", len(fallbackRefs))
			}
		} else {
			fmt.Printf("查找成功: %d 个引用, 来自缓存: %t\n", len(refs), fromCache)
		}
	} else {
		fmt.Println("\n直接LSP查找...")
		refs, err := tsmorphgo.FindReferences(*targetNode)
		if err != nil {
			fmt.Printf("查找失败: %v\n", err)
		} else {
			fmt.Printf("查找成功: %d 个引用\n", len(refs))
		}
	}

	// 如果启用了性能指标
	if config.PerformanceSettings.EnableMetrics {
		fmt.Println("\n性能指标收集...")
		collector := tsmorphgo.NewMetricsCollector(project)

		// 执行几次查询收集指标
		for i := 0; i < 3; i++ {
			_, err := collector.FindReferencesWithMetrics(*targetNode)
			if err != nil {
				fmt.Printf("指标查询 %d 失败: %v\n", i+1, err)
			}
		}

		metrics := collector.GetMetrics()
		fmt.Printf("指标统计:\n")
		fmt.Printf("  总查询: %d\n", metrics.TotalQueries)
		fmt.Printf("  缓存命中: %d (%.1f%%)\n", metrics.CacheHits, metrics.HitRate())
		fmt.Printf("  平均延迟: %v\n", metrics.AverageLatency)
	}

	// 显示项目缓存统计
	if config.CacheSettings.Enabled {
		cacheStats := project.GetCacheStats()
		fmt.Printf("\n缓存统计:\n")
		fmt.Printf("  条目数: %d / %d\n", cacheStats.TotalEntries, config.CacheSettings.MaxEntries)
		fmt.Printf("  访问次数: %d\n", cacheStats.TotalAccesses)
		fmt.Printf("  过期条目: %d\n", cacheStats.ExpiredEntries)
	}

	fmt.Println()
}
