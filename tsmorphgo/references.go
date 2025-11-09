package tsmorphgo

import (
	"context"
	"crypto/md5"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/lsp/lsproto"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"
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
	MaxRetries      int           `json:"maxRetries"`
	BaseDelay       time.Duration `json:"baseDelay"`
	MaxDelay        time.Duration `json:"maxDelay"`
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
// 缓存机制 - 来自 reference_cache.go (简化版)
// =============================================================================

// ReferenceCache 引用查找结果的缓存
// 用于缓存 FindReferences 和 GotoDefinition 的结果，避免重复的 LSP 调用
type ReferenceCache struct {
	// cache 存储缓存结果，key 为文件路径+节点位置+内容哈希
	cache map[string]*CachedReference
	// mu 读写锁，保护并发访问
	mu sync.RWMutex
	// maxEntries 缓存最大条目数
	maxEntries int
	// ttl 缓存条目的生存时间
	ttl time.Duration
}

// CachedReference 缓存的引用查找结果
type CachedReference struct {
	// nodes 查找到的节点列表
	nodes []*Node
	// timestamp 缓存创建时间戳
	timestamp time.Time
	// fileHashes 相关文件的内容哈希，用于检测文件变化
	fileHashes map[string]string
}

// NewReferenceCache 创建新的引用缓存
func NewReferenceCache(maxEntries int, ttl time.Duration) *ReferenceCache {
	return &ReferenceCache{
		cache:      make(map[string]*CachedReference),
		maxEntries: maxEntries,
		ttl:        ttl,
	}
}

// GenerateCacheKey 生成缓存键
// 基于文件路径、行列号和文件内容生成唯一键
func (rc *ReferenceCache) GenerateCacheKey(node Node) string {
	// 1. 获取基本信息
	filePath := node.GetSourceFile().GetFilePath()
	line := node.GetStartLineNumber()
	col := node.GetStartColumnNumber()

	// 2. 计算文件内容哈希
	content := node.GetSourceFile().fileResult.Raw
	contentHash := rc.md5Hash(content)

	// 3. 生成唯一键
	return fmt.Sprintf("%s:%d:%d:%s", filePath, line, col, contentHash[:8])
}

// md5Hash 计算 MD5 哈希
func (rc *ReferenceCache) md5Hash(content string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(content)))
}

// isExpired 检查缓存条目是否过期
func (cr *CachedReference) isExpired(ttl time.Duration) bool {
	return time.Since(cr.timestamp) > ttl
}

// isValid 检查缓存条目是否仍然有效
// 通过比较文件内容哈希来检测文件是否发生变化
func (cr *CachedReference) isValid(project *Project, ttl time.Duration) bool {
	// 1. 检查时间过期
	if cr.isExpired(ttl) {
		return false
	}

	// 2. 检查文件内容是否发生变化
	for filePath, expectedHash := range cr.fileHashes {
		if sourceFile := project.GetSourceFile(filePath); sourceFile != nil {
			currentContent := sourceFile.fileResult.Raw
			currentHash := md5.Sum([]byte(currentContent))
			currentHashStr := fmt.Sprintf("%x", currentHash)

			if currentHashStr != expectedHash {
				return false
			}
		} else {
			// 文件不存在了，缓存失效
			return false
		}
	}

	return true
}

// Get 从缓存中获取引用查找结果
func (rc *ReferenceCache) Get(key string, project *Project) ([]*Node, bool) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	cached, exists := rc.cache[key]
	if !exists {
		return nil, false
	}

	// 检查缓存是否有效
	if !cached.isValid(project, rc.ttl) {
		// 缓存已过期，需要清理
		// 注意：这里不能直接删除，因为只有读锁
		return nil, false
	}

	return cached.nodes, true
}

// Set 将引用查找结果存入缓存
func (rc *ReferenceCache) Set(key string, nodes []*Node, project *Project) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// 检查是否需要清理空间
	if len(rc.cache) >= rc.maxEntries {
		rc.evictOldest()
	}

	// 计算相关文件的哈希
	fileHashes := make(map[string]string)
	affectedFiles := make(map[string]bool)

	// 收集所有相关的文件
	for _, node := range nodes {
		affectedFiles[node.GetSourceFile().GetFilePath()] = true
	}

	// 如果nodes不为空，也包括查询节点所在的文件
	if len(nodes) > 0 {
		affectedFiles[nodes[0].GetSourceFile().GetFilePath()] = true
	}

	// 计算文件哈希
	for filePath := range affectedFiles {
		if sourceFile := project.GetSourceFile(filePath); sourceFile != nil {
			content := sourceFile.fileResult.Raw
			hash := rc.md5Hash(content)
			fileHashes[filePath] = hash
		}
	}

	// 创建缓存条目
	currentTime := time.Now()
	rc.cache[key] = &CachedReference{
		nodes:       nodes,
		timestamp:   currentTime,
		fileHashes:  fileHashes,
	}
}

// evictOldest 清理最旧的缓存条目
func (rc *ReferenceCache) evictOldest() {
	if len(rc.cache) == 0 {
		return
	}

	// 找到最旧的条目
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, cached := range rc.cache {
		if first || cached.timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = cached.timestamp
			first = false
		}
	}

	// 删除最旧的条目
	if oldestKey != "" {
		delete(rc.cache, oldestKey)
	}
}

// Clear 清空所有缓存
func (rc *ReferenceCache) Clear() {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.cache = make(map[string]*CachedReference)
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

// FindReferencesFallback 降级策略：当 LSP 服务不可用时使用的简单查找方法
func FindReferencesFallback(node Node) []*Node {
	var results []*Node

	// 1. 获取节点的名称
	nodeName, ok := node.GetNodeName()
	if !ok || nodeName == "" {
		return results
	}

	// 2. 在同一文件中查找同名的标识符
	sourceFile := node.GetSourceFile()
	if sourceFile == nil {
		return results
	}

	sourceFile.ForEachDescendant(func(descendant Node) {
		// 跳过自身
		if descendant.GetStart() == node.GetStart() && descendant.GetEnd() == node.GetEnd() {
			return
		}

		// 检查是否为标识符且名称匹配
		if descendant.IsIdentifierNode() {
			if descName, ok := descendant.GetNodeName(); ok && descName == nodeName {
				results = append(results, &descendant)
			}
		}
	})

	return results
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