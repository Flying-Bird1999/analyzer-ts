package tsmorphgo

import (
	"fmt"
	"strings"
	"time"
)

// FindReferencesWithCache 带缓存的引用查找，支持错误处理和重试机制
// 首先检查缓存，如果缓存命中则直接返回；否则调用 LSP 服务并缓存结果
// 返回：节点列表、是否来自缓存、错误
func FindReferencesWithCache(node Node) ([]*Node, bool, error) {
	return FindReferencesWithCacheAndRetry(node, DefaultRetryConfig())
}

// FindReferencesWithCacheAndRetry 带缓存和重试机制的引用查找
// 允许自定义重试配置，提供更细粒度的错误控制
func FindReferencesWithCacheAndRetry(node Node, retryConfig *RetryConfig) ([]*Node, bool, error) {
	// 1. 获取项目和缓存
	project := node.GetSourceFile().GetProject()
	cache := project.getReferenceCache()

	// 2. 生成缓存键
	cacheKey := cache.GenerateCacheKey(node)

	// 3. 尝试从缓存获取
	if cached, hit := cache.Get(cacheKey, project); hit {
		return cached, true, nil
	}

	// 4. 缓存未命中，执行带重试的 LSP 调用
	refs, err := findReferencesWithRetry(node, retryConfig)
	if err != nil {
		// 5. 尝试降级策略
		if refErr, ok := err.(*ReferenceError); ok && refErr.ShouldUseFallback() {
			refs = FindReferencesFallback(node)
			if len(refs) > 0 {
				// 降级策略成功，缓存结果
				cache.Set(cacheKey, refs, project)
				return refs, false, nil
			}
		}
		return nil, false, err
	}

	// 6. 将结果存入缓存
	cache.Set(cacheKey, refs, project)

	return refs, false, nil
}

// GotoDefinitionWithCache 带缓存的跳转到定义，支持错误处理和重试机制
// 返回：节点列表、是否来自缓存、错误
func GotoDefinitionWithCache(node Node) ([]*Node, bool, error) {
	return GotoDefinitionWithCacheAndRetry(node, DefaultRetryConfig())
}

// GotoDefinitionWithCacheAndRetry 带缓存和重试机制的跳转到定义
// 允许自定义重试配置，提供更细粒度的错误控制
func GotoDefinitionWithCacheAndRetry(node Node, retryConfig *RetryConfig) ([]*Node, bool, error) {
	// 1. 获取项目和缓存
	project := node.GetSourceFile().GetProject()
	cache := project.getReferenceCache()

	// 2. 生成缓存键
	cacheKey := cache.GenerateCacheKey(node)

	// 3. 尝试从缓存获取
	if cached, hit := cache.Get(cacheKey, project); hit {
		return cached, true, nil
	}

	// 4. 缓存未命中，执行带重试的 LSP 调用
	defs, err := gotoDefinitionWithRetry(node, retryConfig)
	if err != nil {
		// 5. 尝试降级策略
		if refErr, ok := err.(*ReferenceError); ok && refErr.ShouldUseFallback() {
			defs = GotoDefinitionFallback(node)
			if len(defs) > 0 {
				// 降级策略成功，缓存结果
				cache.Set(cacheKey, defs, project)
				return defs, false, nil
			}
		}
		return nil, false, err
	}

	// 6. 将结果存入缓存
	cache.Set(cacheKey, defs, project)

	return defs, false, nil
}

// FindReferencesBatch 批量查找引用，减少 LSP 调用开销
// 对多个节点进行批量引用查找，通过缓存优化提升性能
func FindReferencesBatch(nodes []Node) (map[string][]*Node, error) {
	if len(nodes) == 0 {
		return make(map[string][]*Node), nil
	}

	// 1. 获取项目和缓存
	project := nodes[0].GetSourceFile().GetProject()
	cache := project.getReferenceCache()

	// 2. 结果容器
	results := make(map[string][]*Node)
	pendingNodes := make(map[string]Node) // 待处理的节点

	// 3. 首先检查缓存
	for _, node := range nodes {
		cacheKey := cache.GenerateCacheKey(node)
		if cached, hit := cache.Get(cacheKey, project); hit {
			results[cacheKey] = cached
		} else {
			pendingNodes[cacheKey] = node
		}
	}

	// 4. 处理缓存未命中的节点
	for cacheKey, node := range pendingNodes {
		refs, err := FindReferences(node)
		if err != nil {
			// 记录错误但继续处理其他节点
			continue
		}
		results[cacheKey] = refs
		// 将结果存入缓存
		cache.Set(cacheKey, refs, project)
	}

	return results, nil
}

// ReferenceMetrics 引用查找的性能指标
type ReferenceMetrics struct {
	// TotalQueries 总查询次数
	TotalQueries int64 `json:"totalQueries"`
	// CacheHits 缓存命中次数
	CacheHits int64 `json:"cacheHits"`
	// CacheMisses 缓存未命中次数
	CacheMisses int64 `json:"cacheMisses"`
	// TotalDuration 总耗时
	TotalDuration time.Duration `json:"totalDuration"`
	// AverageLatency 平均延迟
	AverageLatency time.Duration `json:"averageLatency"`
	// LSPCalls LSP 调用次数
	LSPCalls int64 `json:"lspCalls"`
}

// HitRate 计算缓存命中率
func (rm *ReferenceMetrics) HitRate() float64 {
	if rm.TotalQueries == 0 {
		return 0.0
	}
	return float64(rm.CacheHits) / float64(rm.TotalQueries) * 100.0
}

// String 返回性能指标的字符串表示
func (rm *ReferenceMetrics) String() string {
	return strings.Join([]string{
		fmt.Sprintf("总查询: %d", rm.TotalQueries),
		fmt.Sprintf("缓存命中: %d (%.1f%%)", rm.CacheHits, rm.HitRate()),
		fmt.Sprintf("LSP调用: %d", rm.LSPCalls),
		fmt.Sprintf("平均延迟: %v", rm.AverageLatency),
	}, ", ")
}

// MetricsCollector 性能指标收集器
type MetricsCollector struct {
	project *Project
	metrics *ReferenceMetrics
}

// NewMetricsCollector 创建新的指标收集器
func NewMetricsCollector(project *Project) *MetricsCollector {
	return &MetricsCollector{
		project: project,
		metrics: &ReferenceMetrics{},
	}
}

// FindReferencesWithMetrics 带性能指标收集的引用查找
func (mc *MetricsCollector) FindReferencesWithMetrics(node Node) ([]*Node, error) {
	start := time.Now()
	defer func() {
		mc.metrics.TotalQueries++
		mc.metrics.TotalDuration += time.Since(start)
		mc.updateAverageLatency()
	}()

	// 1. 获取缓存
	cache := mc.project.getReferenceCache()
	cacheKey := cache.GenerateCacheKey(node)

	// 2. 检查缓存
	if cached, hit := cache.Get(cacheKey, mc.project); hit {
		mc.metrics.CacheHits++
		return cached, nil
	}

	// 3. 缓存未命中，调用 LSP
	mc.metrics.CacheMisses++
	mc.metrics.LSPCalls++

	refs, err := FindReferences(node)
	if err != nil {
		return nil, err
	}

	// 4. 缓存结果
	cache.Set(cacheKey, refs, mc.project)

	return refs, nil
}

// updateAverageLatency 更新平均延迟
func (mc *MetricsCollector) updateAverageLatency() {
	if mc.metrics.TotalQueries > 0 {
		mc.metrics.AverageLatency = mc.metrics.TotalDuration / time.Duration(mc.metrics.TotalQueries)
	}
}

// GetMetrics 获取当前性能指标
func (mc *MetricsCollector) GetMetrics() *ReferenceMetrics {
	return mc.metrics
}

// ResetMetrics 重置性能指标
func (mc *MetricsCollector) ResetMetrics() {
	mc.metrics = &ReferenceMetrics{}
}

// GetCacheStats 获取项目的缓存统计信息
func (p *Project) GetCacheStats() CacheStats {
	cache := p.getReferenceCache()
	return cache.Stats()
}

// ClearReferenceCache 清空引用缓存
func (p *Project) ClearReferenceCache() {
	if p.referenceCache != nil {
		p.referenceCache.Clear()
	}
}

// findReferencesWithRetry 带重试机制的引用查找
// 使用指数退避算法进行重试，提供对 LSP 服务故障的容错能力
func findReferencesWithRetry(node Node, retryConfig *RetryConfig) ([]*Node, error) {
	var lastErr error

	for attempt := 0; attempt <= retryConfig.MaxRetries; attempt++ {
		// 执行引用查找
		refs, err := FindReferences(node)
		if err == nil {
			return refs, nil
		}

		// 记录错误
		lastErr = err

		// 将通用错误包装为 ReferenceError
		refErr := WrapError(err, fmt.Sprintf("引用查找失败 (尝试 %d/%d)", attempt+1, retryConfig.MaxRetries+1), node)

		// 如果这是第一次错误，创建可重试的错误
		if attempt == 0 {
			refErr.RetryCount = attempt
			refErr.Retryable = refErr.IsRetryable()
		} else if refErr.Retryable {
			// 增加重试计数
			refErr.IncrementRetryCount()
		}

		// 检查是否应该重试
		if !refErr.IsRetryable() || attempt >= retryConfig.MaxRetries {
			return nil, refErr
		}

		// 等待重试延迟
		delay := retryConfig.CalculateDelay(attempt)
		time.Sleep(delay)
	}

	// 所有重试都失败，返回最后的错误
	if refErr, ok := lastErr.(*ReferenceError); ok {
		return nil, refErr
	}
	return nil, WrapError(lastErr, fmt.Sprintf("引用查找失败，已重试 %d 次", retryConfig.MaxRetries), node)
}

// gotoDefinitionWithRetry 带重试机制的跳转到定义
// 使用指数退避算法进行重试，提供对 LSP 服务故障的容错能力
func gotoDefinitionWithRetry(node Node, retryConfig *RetryConfig) ([]*Node, error) {
	var lastErr error

	for attempt := 0; attempt <= retryConfig.MaxRetries; attempt++ {
		// 执行定义查找
		defs, err := GotoDefinition(node)
		if err == nil {
			return defs, nil
		}

		// 记录错误
		lastErr = err

		// 将通用错误包装为 ReferenceError
		refErr := WrapError(err, fmt.Sprintf("定义查找失败 (尝试 %d/%d)", attempt+1, retryConfig.MaxRetries+1), node)

		// 如果这是第一次错误，创建可重试的错误
		if attempt == 0 {
			refErr.RetryCount = attempt
			refErr.Retryable = refErr.IsRetryable()
		} else if refErr.Retryable {
			// 增加重试计数
			refErr.IncrementRetryCount()
		}

		// 检查是否应该重试
		if !refErr.IsRetryable() || attempt >= retryConfig.MaxRetries {
			return nil, refErr
		}

		// 等待重试延迟
		delay := retryConfig.CalculateDelay(attempt)
		time.Sleep(delay)
	}

	// 所有重试都失败，返回最后的错误
	if refErr, ok := lastErr.(*ReferenceError); ok {
		return nil, refErr
	}
	return nil, WrapError(lastErr, fmt.Sprintf("定义查找失败，已重试 %d 次", retryConfig.MaxRetries), node)
}

// FindReferencesFallback 引用查找的降级策略
// 当 LSP 服务不可用时，使用本地 AST 分析进行简单的引用查找
func FindReferencesFallback(node Node) []*Node {
	var refs []*Node

	// 检查节点是否有效
	if !node.IsValid() {
		return refs
	}

	// 获取节点文本作为查找目标
	nodeText := strings.TrimSpace(node.GetText())
	if nodeText == "" {
		return refs
	}

	// 获取源文件
	sourceFile := node.GetSourceFile()
	if sourceFile == nil {
		return refs
	}

	// 在当前文件中查找相同的标识符
	sourceFile.ForEachDescendant(func(descendant Node) {
		// 跳过节点本身
		if descendant == node {
			return
		}

		// 检查是否是标识符且文本匹配
		if IsIdentifier(descendant) && descendant.GetText() == nodeText {
			// 简单的上下文验证 - 确保是真正的引用而不是定义
			if IsLikelyReference(descendant) {
				nodeCopy := descendant
				refs = append(refs, &nodeCopy)
			}
		}
	})

	return refs
}

// GotoDefinitionFallback 跳转到定义的降级策略
// 当 LSP 服务不可用时，使用本地 AST 分析进行简单的定义查找
func GotoDefinitionFallback(node Node) []*Node {
	var defs []*Node

	// 检查节点是否有效
	if !node.IsValid() {
		return defs
	}

	// 获取节点文本作为查找目标
	nodeText := strings.TrimSpace(node.GetText())
	if nodeText == "" {
		return defs
	}

	// 获取源文件
	sourceFile := node.GetSourceFile()
	if sourceFile == nil {
		return defs
	}

	// 在当前文件中查找定义
	sourceFile.ForEachDescendant(func(descendant Node) {
		// 跳过节点本身
		if descendant == node {
			return
		}

		// 检查是否是标识符且文本匹配
		if IsIdentifier(descendant) && descendant.GetText() == nodeText {
			// 检查是否可能是定义
			if IsLikelyDefinition(descendant) {
				nodeCopy := descendant
				defs = append(defs, &nodeCopy)
			}
		}
	})

	return defs
}

// IsLikelyReference 判断节点是否可能是引用
// 通过检查父节点的上下文来判断
func IsLikelyReference(node Node) bool {
	parent := node.GetParent()
	if parent == nil {
		return false
	}

	// 检查节点的位置来判断是否是定义
	// 如果节点在父节点的开始位置，可能是定义
	nodeStart := node.GetStart()
	parentStart := parent.GetStart()

	// 如果父节点是变量声明，且节点在父节点开始位置，则可能是定义
	if parent.Kind == 164 { // VariableDeclaration
		if nodeStart-parentStart < 50 { // 简单的位置检查
			return false // 这很可能是定义，不是引用
		}
	}

	// 如果父节点是函数声明，且节点紧跟在function关键字后，则是定义
	if parent.Kind == 259 { // FunctionDeclaration
		parentText := parent.GetText()
		nodeText := node.GetText()
		if strings.HasPrefix(parentText, "function "+nodeText) {
			return false // 这是定义，不是引用
		}
	}

	// 如果父节点是类声明，且节点紧跟在class关键字后，则是定义
	if parent.Kind == 256 { // ClassDeclaration
		parentText := parent.GetText()
		nodeText := node.GetText()
		if strings.HasPrefix(parentText, "class "+nodeText) {
			return false // 这是定义，不是引用
		}
	}

	// 其他情况认为是引用
	return true
}

// IsLikelyDefinition 判断节点是否可能是定义
// 通过检查父节点的上下文来判断
func IsLikelyDefinition(node Node) bool {
	parent := node.GetParent()
	if parent == nil {
		return false
	}

	// 检查节点的位置来判断是否是定义
	nodeStart := node.GetStart()
	parentStart := parent.GetStart()

	// 如果父节点是变量声明，且节点在父节点开始位置，则可能是定义
	if parent.Kind == 164 { // VariableDeclaration
		if nodeStart-parentStart < 50 { // 简单的位置检查
			return true
		}
	}

	// 如果父节点是函数声明，且节点紧跟在function关键字后，则可能是定义
	if parent.Kind == 259 { // FunctionDeclaration
		parentText := parent.GetText()
		nodeText := node.GetText()
		if strings.HasPrefix(parentText, "function "+nodeText) {
			return true
		}
	}

	// 如果父节点是类声明，且节点紧跟在class关键字后，则可能是定义
	if parent.Kind == 256 { // ClassDeclaration
		parentText := parent.GetText()
		nodeText := node.GetText()
		if strings.HasPrefix(parentText, "class "+nodeText) {
			return true
		}
	}

	// 如果父节点是参数列表，则可能是参数定义
	if parent.Kind == 175 { // ParameterList
		return true
	}

	// 如果父节点是属性声明，则可能是属性定义
	if parent.Kind == 141 { // PropertyDeclaration
		return true
	}

	return false
}