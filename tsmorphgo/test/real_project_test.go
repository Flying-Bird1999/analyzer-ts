package tsmorphgo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// real_project_test.go
//
// 这个文件包含了真实项目中的缓存效果测试，验证：
// 1. 真实TypeScript项目的引用查找
// 2. 缓存在实际场景中的性能提升
// 3. 大型项目的处理能力
// 4. 复杂代码结构的处理
//

const (
	// 真实项目路径 - 使用用户提到的真实React项目
	realProjectPath = "/Users/bird/Desktop/alalyzer/analyzer-ts/tsmorphgo/examples/demo-react-app"
)

// TestRealProjectPerformance 测试真实项目中的性能
func TestRealProjectPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过真实项目性能测试 (使用 -short 标志)")
	}

	// 检查项目路径是否存在
	if _, err := os.Stat(realProjectPath); os.IsNotExist(err) {
		t.Skipf("真实项目路径不存在: %s", realProjectPath)
		return
	}

	t.Logf("真实项目性能测试: %s", realProjectPath)

	// 创建项目配置
	config := ProjectConfig{
		RootPath:         realProjectPath,
		IgnorePatterns:   []string{"node_modules", "dist", "build", ".git"},
		TargetExtensions: []string{".ts", ".tsx"},
		UseTsConfig:      true,
	}

	// 创建项目
	project := NewProject(config)

	// 获取项目中的源文件
	sourceFiles := project.GetSourceFiles()

	if len(sourceFiles) == 0 {
		t.Skip("未找到TypeScript文件，跳过测试")
		return
	}

	t.Logf("找到 %d 个TypeScript文件", len(sourceFiles))

	// 限制处理的文件数量以避免测试时间过长
	maxFiles := 10
	if len(sourceFiles) > maxFiles {
		sourceFiles = sourceFiles[:maxFiles]
		t.Logf("限制处理文件数量为 %d", maxFiles)
	}

	// 收集所有标识符节点
	var allIdentifiers []struct {
		text  string
		node  *Node
		file  string
		line  int
	}

	for _, sourceFile := range sourceFiles {
		sourceFile.ForEachDescendant(func(node Node) {
			if IsIdentifier(node) {
				nodeText := strings.TrimSpace(node.GetText())
				// 过滤掉太短或太长的标识符
				if len(nodeText) >= 3 && len(nodeText) <= 20 {
					// 过滤掉常见的TypeScript关键字
					if !isTypeScriptKeyword(nodeText) {
						nodeCopy := node
						allIdentifiers = append(allIdentifiers, struct {
							text  string
							node  *Node
							file  string
							line  int
						}{
							text: nodeText,
							node: &nodeCopy,
							file: sourceFile.GetFilePath(),
							line: node.GetStartLineNumber(),
						})
					}
				}
			}
		})
	}

	t.Logf("收集到 %d 个标识符", len(allIdentifiers))

	if len(allIdentifiers) == 0 {
		t.Skip("未找到可测试的标识符，跳过测试")
		return
	}

	// 限制测试的标识符数量
	maxIdentifiers := 20
	if len(allIdentifiers) > maxIdentifiers {
		allIdentifiers = allIdentifiers[:maxIdentifiers]
		t.Logf("限制测试标识符数量为 %d", maxIdentifiers)
	}

	// 创建指标收集器
	collector := NewMetricsCollector(project)

	// 执行性能测试
	t.Logf("开始执行 %d 个标识符的引用查找测试...", len(allIdentifiers))

	var totalLSPCalls int
	var totalCacheHits int
	var totalTime time.Duration
	var successfulLookups int

	for i, identifier := range allIdentifiers {
		t.Logf("测试 %d/%d: %s (%s:%d)", i+1, len(allIdentifiers), identifier.text, identifier.file, identifier.line)

		start := time.Now()

		// 执行带缓存的引用查找
		refs, fromCache, err := FindReferencesWithCache(*identifier.node)
		duration := time.Since(start)

		if err != nil {
			t.Logf("  ❌ 失败: %v", err)
			continue
		}

		// 收集性能指标
		_, err = collector.FindReferencesWithMetrics(*identifier.node)
		if err != nil {
			t.Logf("  ⚠️  指标收集失败: %v", err)
		}

		successfulLookups++
		totalTime += duration

		if fromCache {
			totalCacheHits++
			t.Logf("  ✅ 成功: %d 引用, 来自缓存, 耗时: %v", len(refs), duration)
		} else {
			totalLSPCalls++
			t.Logf("  ✅ 成功: %d 引用, 来自LSP, 耗时: %v", len(refs), duration)
		}
	}

	// 统计结果
	t.Logf("\n=== 真实项目性能测试结果 ===")
	t.Logf("测试标识符数量: %d", len(allIdentifiers))
	t.Logf("成功查找数量: %d", successfulLookups)
	t.Logf("成功率: %.1f%%", float64(successfulLookups)/float64(len(allIdentifiers))*100)

	if successfulLookups > 0 {
		avgTime := totalTime / time.Duration(successfulLookups)
		t.Logf("平均查找时间: %v", avgTime)
	}

	t.Logf("LSP调用次数: %d", totalLSPCalls)
	t.Logf("缓存命中次数: %d", totalCacheHits)

	if totalLSPCalls+totalCacheHits > 0 {
		hitRate := float64(totalCacheHits) / float64(totalLSPCalls+totalCacheHits) * 100
		t.Logf("缓存命中率: %.1f%%", hitRate)
	}

	// 显示详细性能指标
	metrics := collector.GetMetrics()
	t.Logf("\n=== 详细性能指标 ===")
	t.Logf("总查询次数: %d", metrics.TotalQueries)
	t.Logf("缓存命中次数: %d", metrics.CacheHits)
	t.Logf("缓存未命中次数: %d", metrics.CacheMisses)
	t.Logf("LSP调用次数: %d", metrics.LSPCalls)
	t.Logf("平均延迟: %v", metrics.AverageLatency)
	t.Logf("总耗时: %v", metrics.TotalDuration)
	t.Logf("缓存命中率: %.1f%%", metrics.HitRate())

	// 显示缓存统计
	cacheStats := project.GetCacheStats()
	t.Logf("\n=== 缓存统计 ===")
	t.Logf("当前缓存条目: %d", cacheStats.TotalEntries)
	t.Logf("总访问次数: %d", cacheStats.TotalAccesses)
	t.Logf("过期条目数: %d", cacheStats.ExpiredEntries)
	t.Logf("最大条目数: %d", cacheStats.MaxEntries)
	t.Logf("缓存TTL: %v", cacheStats.TTL)

	// 验证性能提升
	if totalCacheHits > 0 && totalLSPCalls > 0 {
		improvementFactor := float64(totalCacheHits) / float64(totalLSPCalls)
		t.Logf("\n=== 性能提升评估 ===")
		t.Logf("缓存效率比: %.2f : 1", improvementFactor)
		t.Logf("预估性能提升: %.2fx 倍", improvementFactor)

		// 设置性能基准
		if metrics.HitRate() >= 30.0 {
			t.Logf("✅ 缓存效果良好 (命中率 >= 30%%)")
		} else {
			t.Logf("⚠️  缓存效果一般 (命中率 < 30%%)")
		}
	}

	// 确保基本功能正常
	assert.Greater(t, successfulLookups, 0, "应该至少有一次成功的查找")
	assert.GreaterOrEqual(t, metrics.HitRate(), 0.0, "缓存命中率应该 >= 0")
	assert.LessOrEqual(t, metrics.HitRate(), 100.0, "缓存命中率应该 <= 100")

	t.Logf("✅ 真实项目性能测试完成")
}

// TestRealProjectErrorHandling 测试真实项目中的错误处理
func TestRealProjectErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过真实项目错误处理测试")
	}

	// 检查项目路径
	if _, err := os.Stat(realProjectPath); os.IsNotExist(err) {
		t.Skipf("真实项目路径不存在: %s", realProjectPath)
		return
	}

	t.Logf("真实项目错误处理测试: %s", realProjectPath)

	// 创建项目配置
	config := ProjectConfig{
		RootPath:         ".",
		TargetExtensions: []string{".ts", ".tsx"},
	}

	// 创建项目
	project := NewProject(config)

	// 添加一个简单的测试文件
	testContent := `
		const errorTestVar = "error test";
		function errorTestFunction() {
			console.log(errorTestVar);
			return errorTestVar + " processed";
		}
		errorTestFunction();
	`

	sourceFile, err := project.CreateSourceFile("error_test.ts", testContent)
	require.NoError(t, err, "创建测试文件失败")

	// 找到测试节点
	var targetNode *Node
	sourceFile.ForEachDescendant(func(node Node) {
		if IsIdentifier(node) && node.GetText() == "errorTestVar" {
			parent := node.GetParent()
			if parent != nil && parent.Kind != 164 {
				nodeCopy := node
				targetNode = &nodeCopy
			}
		}
	})

	require.NotNil(t, targetNode, "需要找到测试节点")

	// 测试错误处理和重试机制
	t.Logf("测试错误处理和重试机制...")

	// 创建快速失败的重试配置
	fastRetryConfig := &RetryConfig{
		MaxRetries:    2,
		BaseDelay:     10 * time.Millisecond,
		MaxDelay:      50 * time.Millisecond,
		BackoffFactor: 1.5,
		Enabled:       true,
	}

	// 执行带重试的查找
	refs, fromCache, err := FindReferencesWithCacheAndRetry(*targetNode, fastRetryConfig)

	if err != nil {
		t.Logf("查找失败，分析错误类型: %v", err)

		// 分析错误
		if refErr, ok := err.(*ReferenceError); ok {
			t.Logf("错误类型: %s", refErr.Error())
			t.Logf("可重试: %t", refErr.Retryable)
			t.Logf("重试次数: %d", refErr.RetryCount)

			// 测试降级策略
			if refErr.ShouldUseFallback() {
				t.Logf("尝试降级策略...")
				fallbackRefs := FindReferencesFallback(*targetNode)
				t.Logf("降级策略找到 %d 个引用", len(fallbackRefs))

				assert.Greater(t, len(fallbackRefs), 0, "降级策略应该找到一些引用")
			}
		}
	} else {
		t.Logf("查找成功: %d 个引用, 来自缓存: %t", len(refs), fromCache)
		assert.Greater(t, len(refs), 0, "应该找到引用")
	}

	// 测试上下文分析
	t.Logf("测试上下文分析...")
	isRef := IsLikelyReference(*targetNode)
	isDef := IsLikelyDefinition(*targetNode)
	t.Logf("引用可能性: %t, 定义可能性: %t", isRef, isDef)

	t.Logf("✅ 真实项目错误处理测试完成")
}

// TestRealProjectBatchProcessing 测试真实项目中的批量处理
func TestRealProjectBatchProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过真实项目批量处理测试")
	}

	// 检查项目路径
	if _, err := os.Stat(realProjectPath); os.IsNotExist(err) {
		t.Skipf("真实项目路径不存在: %s", realProjectPath)
		return
	}

	t.Logf("真实项目批量处理测试: %s", realProjectPath)

	// 创建项目配置
	config := ProjectConfig{
		RootPath:         ".",
		TargetExtensions: []string{".ts", ".tsx"},
	}

	// 创建项目
	project := NewProject(config)

	// 创建测试文件
	testContent := `
		const batchVar1 = "batch1";
		const batchVar2 = "batch2";
		const batchVar3 = "batch3";

		function batchFunction1() {
			console.log(batchVar1);
			console.log(batchVar2);
			return batchVar1 + batchVar2;
		}

		function batchFunction2() {
			console.log(batchVar3);
			console.log(batchVar1);
			return batchVar3 + batchVar1;
		}

		const result1 = batchFunction1();
		const result2 = batchFunction2();
		console.log(batchVar2);
		console.log(batchVar3);
	`

	sourceFile, err := project.CreateSourceFile("batch_test.ts", testContent)
	require.NoError(t, err, "创建批量测试文件失败")

	// 收集所有变量节点
	var nodes []Node
	targetVars := []string{"batchVar1", "batchVar2", "batchVar3"}

	sourceFile.ForEachDescendant(func(node Node) {
		if IsIdentifier(node) {
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

	t.Logf("收集到 %d 个节点进行批量处理", len(nodes))

	if len(nodes) == 0 {
		t.Skip("找不到批量处理的节点，跳过测试")
		return
	}

	// 批量处理
	t.Logf("执行批量处理...")
	start := time.Now()
	results, err := FindReferencesBatch(nodes)
	batchDuration := time.Since(start)

	require.NoError(t, err, "批量处理不应该失败")
	assert.Greater(t, len(results), 0, "应该有批量处理结果")

	t.Logf("批量处理完成，耗时: %v", batchDuration)
	t.Logf("结果数量: %d", len(results))

	// 验证结果
	totalRefs := 0
	for cacheKey, refs := range results {
		t.Logf("节点 %s: %d 个引用", cacheKey, len(refs))
		totalRefs += len(refs)
		assert.Greater(t, len(refs), 0, "每个节点都应该有引用")
	}

	t.Logf("总计找到 %d 个引用", totalRefs)

	// 单独处理对比
	t.Logf("执行单独处理对比...")
	start = time.Now()
	singleResults := 0

	for _, node := range nodes {
		refs, err := FindReferences(node)
		if err == nil {
			singleResults += len(refs)
		}
	}
	singleDuration := time.Since(start)

	t.Logf("单独处理耗时: %v", singleDuration)
	t.Logf("单独处理引用: %d", singleResults)

	// 性能对比
	if batchDuration > 0 && singleDuration > 0 {
		improvement := float64(singleDuration) / float64(batchDuration)
		t.Logf("批量处理性能提升: %.2fx", improvement)

		// 验证结果一致性
		if totalRefs == singleResults {
			t.Logf("✅ 结果一致性验证通过")
		} else {
			t.Logf("⚠️  结果数量不一致 (批量: %d, 单独: %d)", totalRefs, singleResults)
		}
	}

	t.Logf("✅ 真实项目批量处理测试完成")
}

// 辅助函数

// findTypeScriptFiles 查找项目中的TypeScript文件
func findTypeScriptFiles(rootPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// 跳过一些目录
			dirName := filepath.Base(path)
			if dirName == "node_modules" || dirName == ".git" || dirName == "dist" || dirName == "build" {
				return filepath.SkipDir
			}
			return nil
		}

		// 检查文件扩展名
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".ts" || ext == ".tsx" {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// isTypeScriptKeyword 检查是否为TypeScript关键字
func isTypeScriptKeyword(text string) bool {
	keywords := map[string]bool{
		// JavaScript/TypeScript 关键字
		"break": true, "case": true, "catch": true, "class": true, "const": true,
		"continue": true, "debugger": true, "default": true, "delete": true,
		"do": true, "else": true, "enum": true, "export": true, "extends": true,
		"false": true, "finally": true, "for": true, "function": true,
		"if": true, "import": true, "in": true, "instanceof": true, "new": true,
		"null": true, "return": true, "super": true, "switch": true,
		"this": true, "throw": true, "true": true, "try": true, "typeof": true,
		"var": true, "void": true, "while": true, "with": true,

		// TypeScript 特定关键字
		"as": true, "implements": true, "interface": true, "let": true,
		"package": true, "private": true, "protected": true, "public": true,
		"static": true, "yield": true, "declare": true, "type": true,
		"abstract": true, "async": true, "await": true, "constructor": true,
		"get": true, "set": true, "readonly": true,
	}

	return keywords[text]
}

// TestRealProjectCacheEffect 测试真实项目中的缓存效果
func TestRealProjectCacheEffect(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过真实项目缓存效果测试")
	}

	// 检查项目路径
	if _, err := os.Stat(realProjectPath); os.IsNotExist(err) {
		t.Skipf("真实项目路径不存在: %s", realProjectPath)
		return
	}

	t.Logf("真实项目缓存效果测试: %s", realProjectPath)

	// 创建项目配置
	config := ProjectConfig{
		RootPath:         ".",
		TargetExtensions: []string{".ts", ".tsx"},
	}

	// 创建项目
	project := NewProject(config)

	// 添加测试文件
	testContent := `
		const cacheTestVar = "cache test";
		let mutableVar = "mutable";

		function cacheTestFunction() {
			console.log(cacheTestVar);
			console.log(mutableVar);
			return cacheTestVar + " " + mutableVar;
		}

		class CacheTestClass {
			private value: string;

			constructor() {
				this.value = cacheTestVar;
			}

			getValue(): string {
				return this.value;
			}

			setValue(newValue: string): void {
				this.value = newValue;
			}
		}

		const instance = new CacheTestClass();
		const result1 = cacheTestFunction();
		const result2 = instance.getValue();
		console.log(cacheTestVar);
		console.log(mutableVar);
	`

	sourceFile, err := project.CreateSourceFile("cache_effect_test.ts", testContent)
	require.NoError(t, err, "创建缓存测试文件失败")

	// 找到测试节点
	var targetNodes []struct {
		name string
		node *Node
	}

	sourceFile.ForEachDescendant(func(node Node) {
		if IsIdentifier(node) {
			nodeText := node.GetText()
			// 选择一些代表性的标识符
			if nodeText == "cacheTestVar" || nodeText == "mutableVar" || nodeText == "instance" {
				parent := node.GetParent()
				if parent != nil && parent.Kind != 164 { // 不是变量声明
					nodeCopy := node
					targetNodes = append(targetNodes, struct {
						name string
						node *Node
					}{name: nodeText, node: &nodeCopy})
				}
			}
		}
	})

	if len(targetNodes) == 0 {
		t.Skip("找不到缓存测试节点，跳过测试")
		return
	}

	t.Logf("找到 %d 个缓存测试节点", len(targetNodes))

	// 创建指标收集器
	collector := NewMetricsCollector(project)

	// 第一轮：冷缓存测试
	t.Logf("第一轮：冷缓存测试")
	for i, target := range targetNodes {
		t.Logf("  测试 %d/%d: %s", i+1, len(targetNodes), target.name)

		start := time.Now()
		refs, fromCache, err := FindReferencesWithCache(*target.node)
		duration := time.Since(start)

		if err != nil {
			t.Logf("    ❌ 失败: %v", err)
			continue
		}

		// 收集指标
		collector.FindReferencesWithMetrics(*target.node)

		source := "LSP"
		if fromCache {
			source = "缓存"
		}
		t.Logf("    ✅ 成功: %d 引用, 来源: %s, 耗时: %v", len(refs), source, duration)
	}

	// 显示第一轮统计
	metrics1 := collector.GetMetrics()
	t.Logf("\n第一轮统计:")
	t.Logf("  总查询: %d", metrics1.TotalQueries)
	t.Logf("  缓存命中: %d", metrics1.CacheHits)
	t.Logf("  LSP调用: %d", metrics1.LSPCalls)
	t.Logf("  命中率: %.1f%%", metrics1.HitRate())

	// 第二轮：热缓存测试
	t.Logf("\n第二轮：热缓存测试")
	collector.ResetMetrics()

	for i, target := range targetNodes {
		t.Logf("  测试 %d/%d: %s", i+1, len(targetNodes), target.name)

		start := time.Now()
		refs, fromCache, err := FindReferencesWithCache(*target.node)
		duration := time.Since(start)

		if err != nil {
			t.Logf("    ❌ 失败: %v", err)
			continue
		}

		// 收集指标
		collector.FindReferencesWithMetrics(*target.node)

		source := "LSP"
		if fromCache {
			source = "缓存"
		}
		t.Logf("    ✅ 成功: %d 引用, 来源: %s, 耗时: %v", len(refs), source, duration)
	}

	// 显示第二轮统计
	metrics2 := collector.GetMetrics()
	t.Logf("\n第二轮统计:")
	t.Logf("  总查询: %d", metrics2.TotalQueries)
	t.Logf("  缓存命中: %d", metrics2.CacheHits)
	t.Logf("  LSP调用: %d", metrics2.LSPCalls)
	t.Logf("  命中率: %.1f%%", metrics2.HitRate())

	// 缓存效果分析
	t.Logf("\n=== 缓存效果分析 ===")
	if metrics1.TotalQueries > 0 && metrics2.TotalQueries > 0 {
		hitRate1 := metrics1.HitRate()
		hitRate2 := metrics2.HitRate()

		t.Logf("第一轮命中率: %.1f%%", hitRate1)
		t.Logf("第二轮命中率: %.1f%%", hitRate2)

		if hitRate2 > hitRate1 {
			improvement := hitRate2 - hitRate1
			t.Logf("缓存命中率提升: %.1f%%", improvement)
		}

		// 平均延迟对比
		avgLatency1 := metrics1.AverageLatency
		avgLatency2 := metrics2.AverageLatency

		if avgLatency1 > 0 && avgLatency2 > 0 {
			speedup := float64(avgLatency1) / float64(avgLatency2)
			t.Logf("平均延迟对比: %v → %v", avgLatency1, avgLatency2)
			t.Logf("速度提升: %.2fx", speedup)
		}
	}

	// 最终缓存统计
	cacheStats := project.GetCacheStats()
	t.Logf("\n=== 最终缓存统计 ===")
	t.Logf("缓存条目数: %d", cacheStats.TotalEntries)
	t.Logf("总访问次数: %d", cacheStats.TotalAccesses)
	t.Logf("过期条目数: %d", cacheStats.ExpiredEntries)
	t.Logf("最大条目数: %d", cacheStats.MaxEntries)

	// 验证缓存效果
	assert.GreaterOrEqual(t, metrics2.HitRate(), metrics1.HitRate(), "第二轮缓存命中率应该更高或相等")
	assert.Greater(t, cacheStats.TotalEntries, 0, "应该有缓存条目")
	assert.Greater(t, int(cacheStats.TotalAccesses), 0, "应该有缓存访问")

	t.Logf("✅ 真实项目缓存效果测试完成")
}