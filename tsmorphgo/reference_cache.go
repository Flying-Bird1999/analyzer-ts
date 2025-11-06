package tsmorphgo

import (
	"crypto/md5"
	"fmt"
	"sync"
	"time"
)

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
	// accessCount 访问次数，用于 LRU 清理
	accessCount int64
	// lastAccessTime 最后访问时间
	lastAccessTime time.Time
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

	// 更新访问统计
	cached.accessCount++
	cached.lastAccessTime = time.Now()

	return cached.nodes, true
}

// Set 将引用查找结果存入缓存
func (rc *ReferenceCache) Set(key string, nodes []*Node, project *Project) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// 检查是否需要清理空间
	if len(rc.cache) >= rc.maxEntries {
		rc.evictLRU()
	}

	// 计算相关文件的哈希
	fileHashes := make(map[string]string)
	affectedFiles := make(map[string]bool)

	// 收集所有相关的文件
	for _, node := range nodes {
		affectedFiles[node.GetSourceFile().GetFilePath()] = true
	}
	// 也包括查询节点所在的文件
	affectedFiles[nodes[0].GetSourceFile().GetFilePath()] = true

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
		nodes:          nodes,
		timestamp:      currentTime,
		fileHashes:     fileHashes,
		accessCount:    1,
		lastAccessTime: currentTime,
	}
}

// evictLRU 清理最少使用的缓存条目
func (rc *ReferenceCache) evictLRU() {
	if len(rc.cache) == 0 {
		return
	}

	// 找到最少使用的条目
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, cached := range rc.cache {
		if first || cached.lastAccessTime.Before(oldestTime) {
			oldestKey = key
			oldestTime = cached.lastAccessTime
			first = false
		}
	}

	// 删除最少使用的条目
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

// Stats 获取缓存统计信息
func (rc *ReferenceCache) Stats() CacheStats {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	var totalAccesses int64
	var expiredCount int

	for _, cached := range rc.cache {
		totalAccesses += cached.accessCount
		if cached.isExpired(rc.ttl) {
			expiredCount++
		}
	}

	return CacheStats{
		TotalEntries:   len(rc.cache),
		TotalAccesses:  totalAccesses,
		ExpiredEntries: expiredCount,
		MaxEntries:     rc.maxEntries,
		TTL:            rc.ttl,
	}
}

// CacheStats 缓存统计信息
type CacheStats struct {
	// TotalEntries 总缓存条目数
	TotalEntries int `json:"totalEntries"`
	// TotalAccesses 总访问次数
	TotalAccesses int64 `json:"totalAccesses"`
	// ExpiredEntries 过期条目数
	ExpiredEntries int `json:"expiredEntries"`
	// MaxEntries 最大条目数
	MaxEntries int `json:"maxEntries"`
	// TTL 缓存生存时间
	TTL time.Duration `json:"ttl"`
}

// HitRate 计算缓存命中率
func (cs CacheStats) HitRate(totalQueries int64) float64 {
	if totalQueries == 0 {
		return 0.0
	}
	return float64(cs.TotalAccesses) / float64(totalQueries) * 100.0
}