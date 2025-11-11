package tsmorphgo

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/lsp/lsproto"
)

// =============================================================================
// 错误处理定义 - 来自 reference_errors.go
// =============================================================================

// ReferenceError 引用查找错误的详细分类
type ReferenceError struct {
	Type       ReferenceErrorType `json:"type"`
	Message    string             `json:"message"`
	Cause      error              `json:"cause,omitempty"`
	NodeInfo   string             `json:"nodeInfo,omitempty"`
	FilePath   string             `json:"filePath,omitempty"`
	LineNumber int                `json:"lineNumber,omitempty"`
	Retryable  bool               `json:"retryable"`
	RetryCount int                `json:"retryCount"`
	Timestamp  time.Time          `json:"timestamp"`
}

// ReferenceErrorType 错误类型枚举
type ReferenceErrorType int

const (
	// LSP服务相关错误
	ErrorTypeLSPService ReferenceErrorType = iota
	ErrorTypeLSPTimeout
	ErrorTypeLSPUnavailable

	// 项目相关错误
	ErrorTypeProjectNotFound
	ErrorTypeFileNotFound
	ErrorTypeInvalidNode

	// 缓存相关错误
	ErrorTypeCacheCorruption
	ErrorTypeCacheExpired

	// 语法相关错误
	ErrorTypeSyntaxError
	ErrorTypeTypeCheckError

	// 未知错误
	ErrorTypeUnknown
)

// String 返回错误的字符串表示
func (e ReferenceError) Error() string {
	var parts []string

	// 错误类型
	parts = append(parts, fmt.Sprintf("[%s]", e.Type.String()))

	// 主要消息
	if e.Message != "" {
		parts = append(parts, e.Message)
	}

	// 位置信息
	if e.FilePath != "" {
		location := e.FilePath
		if e.LineNumber > 0 {
			location += fmt.Sprintf(":%d", e.LineNumber)
		}
		parts = append(parts, fmt.Sprintf("at %s", location))
	}

	// 节点信息
	if e.NodeInfo != "" {
		parts = append(parts, fmt.Sprintf("node: %s", e.NodeInfo))
	}

	// 重试信息
	if e.RetryCount > 0 {
		parts = append(parts, fmt.Sprintf("retry: %d", e.RetryCount))
	}

	return strings.Join(parts, " ")
}

// ShouldRetry 判断是否应该重试
func (e *ReferenceError) ShouldRetry() bool {
	if !e.Retryable {
		return false
	}

	// 根据错误类型决定是否重试
	switch e.Type {
	case ErrorTypeLSPTimeout, ErrorTypeLSPService, ErrorTypeCacheExpired:
		return e.RetryCount < 3 // 最多重试3次
	case ErrorTypeLSPUnavailable:
		return e.RetryCount < 1 // 只重试1次
	default:
		return false
	}
}

// ShouldUseFallback 判断是否应该使用降级策略
func (e *ReferenceError) ShouldUseFallback() bool {
	switch e.Type {
	case ErrorTypeLSPService, ErrorTypeLSPTimeout, ErrorTypeLSPUnavailable:
		return true
	default:
		return false
	}
}

// String 返回错误类型的字符串表示
func (t ReferenceErrorType) String() string {
	switch t {
	case ErrorTypeLSPService:
		return "LSP_SERVICE_ERROR"
	case ErrorTypeLSPTimeout:
		return "LSP_TIMEOUT"
	case ErrorTypeLSPUnavailable:
		return "LSP_UNAVAILABLE"
	case ErrorTypeProjectNotFound:
		return "PROJECT_NOT_FOUND"
	case ErrorTypeFileNotFound:
		return "FILE_NOT_FOUND"
	case ErrorTypeInvalidNode:
		return "INVALID_NODE"
	case ErrorTypeCacheCorruption:
		return "CACHE_CORRUPTION"
	case ErrorTypeCacheExpired:
		return "CACHE_EXPIRED"
	case ErrorTypeSyntaxError:
		return "SYNTAX_ERROR"
	case ErrorTypeTypeCheckError:
		return "TYPE_CHECK_ERROR"
	default:
		return "UNKNOWN_ERROR"
	}
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries      int                  `json:"maxRetries"`
	BaseDelay       time.Duration        `json:"baseDelay"`
	MaxDelay        time.Duration        `json:"maxDelay"`
	RetryableErrors []ReferenceErrorType `json:"retryableErrors"`
}

// DefaultRetryConfig 返回默认的重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries: 3,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   2 * time.Second,
		RetryableErrors: []ReferenceErrorType{
			ErrorTypeLSPService,
			ErrorTypeLSPTimeout,
			ErrorTypeCacheExpired,
		},
	}
}

// =============================================================================
// 简化的引用缓存机制
// =============================================================================

// ReferenceCache 简化的引用查找结果缓存
// 移除复杂的文件哈希检查，采用基于时间的过期策略
type ReferenceCache struct {
	// cache 存储缓存结果
	cache map[string]*SimpleCacheEntry
	// mu 读写锁，保护并发访问
	mu sync.RWMutex
	// maxSize 缓存最大条目数
	maxSize int
	// ttl 缓存条目的生存时间
	ttl time.Duration
	// 清理间隔
	cleanupInterval time.Duration
	// 最后清理时间
	lastCleanup time.Time
}

// SimpleCacheEntry 简化的缓存条目
type SimpleCacheEntry struct {
	// nodes 查找到的节点列表
	nodes []*Node
	// timestamp 缓存创建时间戳
	timestamp time.Time
}

// NewSimpleReferenceCache 创建新的简化引用缓存
func NewSimpleReferenceCache(maxSize int, ttl time.Duration) *ReferenceCache {
	return &ReferenceCache{
		cache:           make(map[string]*SimpleCacheEntry),
		maxSize:         maxSize,
		ttl:             ttl,
		cleanupInterval: ttl / 4, // 每1/4 TTL时间清理一次过期条目
		lastCleanup:     time.Now(),
	}
}

// GenerateSimpleCacheKey 生成简化的缓存键
// 只基于文件路径和节点位置，不包含文件内容哈希
func (sc *ReferenceCache) GenerateSimpleCacheKey(node Node) string {
	// 简化的键生成：文件路径 + 行号 + 列号
	filePath := node.GetSourceFile().GetFilePath()
	line := node.GetStartLineNumber()
	col := node.GetStartColumnNumber()

	return fmt.Sprintf("%s:%d:%d", filePath, line, col)
}

// Get 从简化缓存中获取引用查找结果
func (sc *ReferenceCache) Get(key string) ([]*Node, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	entry, exists := sc.cache[key]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Since(entry.timestamp) > sc.ttl {
		return nil, false
	}

	return entry.nodes, true
}

// Set 将引用查找结果存入简化缓存
func (sc *ReferenceCache) Set(key string, nodes []*Node) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// 定期清理过期条目
	sc.cleanupExpired()

	// 检查是否需要清理空间
	if len(sc.cache) >= sc.maxSize {
		sc.evictRandom()
	}

	// 存储新的缓存条目
	sc.cache[key] = &SimpleCacheEntry{
		nodes:     nodes,
		timestamp: time.Now(),
	}
}

// cleanupExpired 清理过期的缓存条目
func (sc *ReferenceCache) cleanupExpired() {
	// 如果距离上次清理时间还不够，跳过清理
	if time.Since(sc.lastCleanup) < sc.cleanupInterval {
		return
	}

	now := time.Now()
	for key, entry := range sc.cache {
		if now.Sub(entry.timestamp) > sc.ttl {
			delete(sc.cache, key)
		}
	}
	sc.lastCleanup = now
}

// evictRandom 随机清理一个缓存条目
// 比LRU更简单，性能更好
func (sc *ReferenceCache) evictRandom() {
	if len(sc.cache) == 0 {
		return
	}

	// 简单策略：删除第一个找到的条目
	for key := range sc.cache {
		delete(sc.cache, key)
		break
	}
}

// Clear 清空所有缓存
func (sc *ReferenceCache) Clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.cache = make(map[string]*SimpleCacheEntry)
	sc.lastCleanup = time.Now()
}

// Stats 返回缓存统计信息
func (sc *ReferenceCache) Stats() CacheStats {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	expiredCount := 0
	now := time.Now()
	for _, entry := range sc.cache {
		if now.Sub(entry.timestamp) > sc.ttl {
			expiredCount++
		}
	}

	return CacheStats{
		TotalEntries:   len(sc.cache),
		ExpiredEntries: expiredCount,
		MaxSize:        sc.maxSize,
		TTL:            sc.ttl,
		LastCleanup:    sc.lastCleanup,
	}
}

// CacheStats 缓存统计信息
type CacheStats struct {
	TotalEntries   int           `json:"totalEntries"`
	ExpiredEntries int           `json:"expiredEntries"`
	MaxSize        int           `json:"maxSize"`
	TTL            time.Duration `json:"ttl"`
	LastCleanup    time.Time     `json:"lastCleanup"`
}

// =============================================================================
// 基础引用查找 - 来自 references.go
// =============================================================================

// FindReferences 查找给定节点所代表的符号的所有引用。
// 注意：此功能依赖的底层 `typescript-go` 库可能存在 bug，导致结果不完全准确。
func FindReferences(node Node) ([]*Node, error) {
	// 1. 获取节点的位置信息
	startLine := node.GetStartLineNumber()
	_, startChar := utils.GetLineAndCharacterOfPosition(node.GetSourceFile().fileResult.Raw, node.Pos())
	startChar += 1 // a a new field to store the ast of the file
	filePath := node.GetSourceFile().filePath

	// 2. 从项目获取共享的 lsp.Service
	lspService, err := node.GetSourceFile().project.getLspService()
	if err != nil {
		return nil, err
	}

	resp, err := lspService.FindReferences(context.Background(), filePath, startLine, startChar)
	if err != nil {
		return nil, err
	}

	// 3. 将返回的 LSP 位置转换为 sdk.Node 列表
	var results []*Node
	if resp.Locations != nil {
		for _, loc := range *resp.Locations {
			// 清理和转换 file URI 到项目内的虚拟路径
			refPath := strings.TrimPrefix(string(loc.Uri), "file://")

			// 使用 project 上的辅助方法来根据位置查找节点
			foundNode := node.GetSourceFile().project.findNodeAt(refPath, int(loc.Range.Start.Line)+1, int(loc.Range.Start.Character)+1)
			if foundNode != nil {
				results = append(results, &Node{
					Node:       foundNode,
					sourceFile: node.GetSourceFile().project.sourceFiles[refPath],
				})
			}
		}
	}

	return results, nil
}

// GotoDefinition 查找给定节点所代表的符号的定义位置。
// 此功能通过 LSP 服务提供精确的跳转到定义能力。
func GotoDefinition(node Node) ([]*Node, error) {
	// 1. 获取节点的位置信息
	startLine := node.GetStartLineNumber()
	_, startChar := utils.GetLineAndCharacterOfPosition(node.GetSourceFile().fileResult.Raw, node.Pos())
	startChar += 1
	filePath := node.GetSourceFile().filePath

	// 2. 从项目获取共享的 lsp.Service
	lspService, err := node.GetSourceFile().project.getLspService()
	if err != nil {
		return nil, err
	}

	// 3. 使用 LSP 服务查找定义
	resp, err := lspService.GotoDefinition(context.Background(), filePath, startLine, startChar)
	if err != nil {
		return nil, err
	}

	// 4. 将返回的 LSP 位置转换为 Node 列表
	var results []*Node

	// 处理定义响应
	if resp.Locations != nil {
		// 处理 Location 数组
		for _, loc := range *resp.Locations {
			if converted := convertLspLocationToNode(loc, node.GetSourceFile().project); converted != nil {
				results = append(results, converted)
			}
		}
	}

	return results, nil
}

// convertLspLocationToNode 辅助函数：将 LSP Location 转换为 Node
func convertLspLocationToNode(loc lsproto.Location, project *Project) *Node {
	// 清理 file URI 到项目内的虚拟路径
	refPath := strings.TrimPrefix(string(loc.Uri), "file://")

	// 使用 project 上的辅助方法来根据位置查找节点
	foundNode := project.findNodeAt(refPath, int(loc.Range.Start.Line)+1, int(loc.Range.Start.Character)+1)
	if foundNode != nil {
		return &Node{
			Node:       foundNode,
			sourceFile: project.sourceFiles[refPath],
		}
	}

	return nil
}

// =============================================================================
// 增强功能 - 来自 references_enhanced.go
// =============================================================================

// FindReferencesWithCache 带缓存的引用查找，支持错误处理和重试机制
// 首先检查缓存，如果缓存命中则直接返回；否则调用 LSP 服务并缓存结果
// 返回：节点列表、是否来自缓存、错误
func FindReferencesWithCache(node Node) ([]*Node, bool, error) {
	return FindReferencesWithCacheAndRetry(node, DefaultRetryConfig())
}

// FindReferencesWithCacheAndRetry 带缓存和重试机制的引用查找
// 允许自定义重试配置，提供更细粒度的错误控制
// 使用简化的缓存实现
func FindReferencesWithCacheAndRetry(node Node, retryConfig *RetryConfig) ([]*Node, bool, error) {
	// 1. 获取项目和缓存
	project := node.GetSourceFile().GetProject()
	cache := project.getReferenceCache()

	// 2. 生成简化的缓存键
	cacheKey := cache.GenerateSimpleCacheKey(node)

	// 3. 尝试从缓存获取
	if cached, hit := cache.Get(cacheKey); hit {
		return cached, true, nil
	}

	// 4. 缓存未命中，执行带重试的 LSP 调用
	refs, err := findReferencesWithRetry(node, retryConfig)
	if err != nil {
		return nil, false, err
	}

	// 6. 将结果存入缓存
	cache.Set(cacheKey, refs)

	return refs, false, nil
}

// findReferencesWithRetry 带重试机制的引用查找
func findReferencesWithRetry(node Node, retryConfig *RetryConfig) ([]*Node, error) {
	var lastErr error

	for attempt := 0; attempt <= retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			// 计算延迟时间 (指数退避)
			delay := time.Duration(attempt) * retryConfig.BaseDelay
			if delay > retryConfig.MaxDelay {
				delay = retryConfig.MaxDelay
			}
			time.Sleep(delay)
		}

		refs, err := FindReferences(node)
		if err == nil {
			return refs, nil
		}

		// 包装错误
		refErr := &ReferenceError{
			Type:       classifyError(err),
			Message:    err.Error(),
			Cause:      err,
			NodeInfo:   node.GetText(),
			FilePath:   node.GetSourceFile().GetFilePath(),
			LineNumber: node.GetStartLineNumber(),
			RetryCount: attempt,
			Timestamp:  time.Now(),
		}

		// 判断是否应该重试
		if !refErr.ShouldRetry() {
			return nil, refErr
		}

		// 检查是否在可重试的错误类型列表中
		retryable := false
		for _, retryableType := range retryConfig.RetryableErrors {
			if refErr.Type == retryableType {
				retryable = true
				break
			}
		}

		if !retryable {
			return nil, refErr
		}

		lastErr = refErr
	}

	return nil, lastErr
}

// classifyError 对错误进行分类
func classifyError(err error) ReferenceErrorType {
	errStr := strings.ToLower(err.Error())

	if strings.Contains(errStr, "timeout") {
		return ErrorTypeLSPTimeout
	}
	if strings.Contains(errStr, "lsp") || strings.Contains(errStr, "service") {
		return ErrorTypeLSPService
	}
	if strings.Contains(errStr, "file not found") {
		return ErrorTypeFileNotFound
	}
	if strings.Contains(errStr, "project") {
		return ErrorTypeProjectNotFound
	}
	if strings.Contains(errStr, "node") || strings.Contains(errStr, "invalid") {
		return ErrorTypeInvalidNode
	}
	if strings.Contains(errStr, "syntax") {
		return ErrorTypeSyntaxError
	}
	if strings.Contains(errStr, "type") {
		return ErrorTypeTypeCheckError
	}
	if strings.Contains(errStr, "cache") {
		return ErrorTypeCacheCorruption
	}

	return ErrorTypeUnknown
}

// =============================================================================
// 便捷方法
// =============================================================================

// FindAllReferences 查找所有引用，包括定义位置
func FindAllReferences(node Node) ([]*Node, error) {
	var allReferences []*Node

	// 1. 查找引用
	refs, err := FindReferences(node)
	if err != nil {
		return nil, err
	}
	allReferences = append(allReferences, refs...)

	// 2. 查找定义
	defs, err := GotoDefinition(node)
	if err != nil {
		// 定义查找失败不影响引用查找结果
		return allReferences, nil
	}
	allReferences = append(allReferences, defs...)

	return allReferences, nil
}

// CountReferences 统计引用数量
func CountReferences(node Node) (int, error) {
	refs, err := FindReferences(node)
	if err != nil {
		return 0, err
	}
	return len(refs), nil
}
