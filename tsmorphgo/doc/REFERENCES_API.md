# TSMorphGo References API 文档

## 概述

TSMorphGo References 模块提供了强大的 TypeScript 代码引用查找功能，包括：
- 基础引用查找 (FindReferences, GotoDefinition)
- 带缓存的优化查找 (FindReferencesWithCache, GotoDefinitionWithCache)
- 错误处理和重试机制
- 降级策略
- 性能监控和指标收集

## 核心API

### 1. 基础API

#### FindReferences
```go
func FindReferences(node Node) ([]*Node, error)
```
查找指定节点的所有引用位置。

**参数:**
- `node Node`: 要查找引用的节点

**返回:**
- `[]*Node`: 引用节点列表
- `error`: 错误信息

**示例:**
```go
// 找到一个标识符节点
var targetNode *Node
sourceFile.ForEachDescendant(func(node Node) {
    if IsIdentifier(node) && node.GetText() == "myVariable" {
        nodeCopy := node
        targetNode = &nodeCopy
    }
})

// 查找所有引用
refs, err := FindReferences(*targetNode)
if err != nil {
    log.Printf("查找引用失败: %v", err)
    return
}

fmt.Printf("找到 %d 个引用\n", len(refs))
for i, ref := range refs {
    fmt.Printf("引用 %d: %s (行 %d)\n", i+1, ref.GetText(), ref.GetStartLineNumber())
}
```

#### GotoDefinition
```go
func GotoDefinition(node Node) ([]*Node, error)
```
跳转到指定节点的定义位置。

**参数:**
- `node Node`: 要查找定义的节点

**返回:**
- `[]*Node`: 定义节点列表
- `error`: 错误信息

### 2. 缓存优化API

#### FindReferencesWithCache
```go
func FindReferencesWithCache(node Node) ([]*Node, bool, error)
```
带缓存的引用查找，提供显著的性能提升。

**参数:**
- `node Node`: 要查找引用的节点

**返回:**
- `[]*Node`: 引用节点列表
- `bool`: 是否来自缓存
- `error`: 错误信息

**性能优势:**
- 首次调用: ~100ms (LSP服务调用)
- 缓存命中: ~0.3ms (400x+ 速度提升)

**示例:**
```go
// 使用缓存优化的引用查找
refs, fromCache, err := FindReferencesWithCache(*targetNode)
if err != nil {
    log.Printf("查找引用失败: %v", err)
    return
}

if fromCache {
    fmt.Println("结果来自缓存")
} else {
    fmt.Println("结果来自LSP服务")
}

fmt.Printf("找到 %d 个引用\n", len(refs))
```

#### GotoDefinitionWithCache
```go
func GotoDefinitionWithCache(node Node) ([]*Node, bool, error)
```
带缓存的定义查找。

#### FindReferencesWithCacheAndRetry
```go
func FindReferencesWithCacheAndRetry(node Node, retryConfig *RetryConfig) ([]*Node, bool, error)
```
带缓存和自定义重试配置的引用查找。

**参数:**
- `node Node`: 要查找引用的节点
- `retryConfig *RetryConfig`: 重试配置

#### GotoDefinitionWithCacheAndRetry
```go
func GotoDefinitionWithCacheAndRetry(node Node, retryConfig *RetryConfig) ([]*Node, bool, error)
```
带缓存和自定义重试配置的定义查找。

### 3. 批量处理API

#### FindReferencesBatch
```go
func FindReferencesBatch(nodes []Node) (map[string][]*Node, error)
```
批量查找多个节点的引用，减少LSP调用开销。

**参数:**
- `nodes []Node`: 要查找的节点列表

**返回:**
- `map[string][]*Node`: 缓存键到引用列表的映射
- `error`: 错误信息

**示例:**
```go
// 收集所有要查找的节点
var nodes []Node
sourceFile.ForEachDescendant(func(node Node) {
    if IsIdentifier(node) {
        nodes = append(nodes, node)
    }
})

// 批量查找
results, err := FindReferencesBatch(nodes)
if err != nil {
    log.Printf("批量查找失败: %v", err)
    return
}

for cacheKey, refs := range results {
    fmt.Printf("节点 %s: 找到 %d 个引用\n", cacheKey, len(refs))
}
```

### 4. 性能监控API

#### MetricsCollector
```go
type MetricsCollector struct {
    project *Project
    metrics *ReferenceMetrics
}

func NewMetricsCollector(project *Project) *MetricsCollector
func (mc *MetricsCollector) FindReferencesWithMetrics(node Node) ([]*Node, error)
func (mc *MetricsCollector) GetMetrics() *ReferenceMetrics
func (mc *MetricsCollector) ResetMetrics()
```

**示例:**
```go
// 创建指标收集器
collector := NewMetricsCollector(project)

// 执行带指标收集的引用查找
refs, err := collector.FindReferencesWithMetrics(*targetNode)
if err != nil {
    log.Printf("查找失败: %v", err)
    return
}

// 获取性能指标
metrics := collector.GetMetrics()
fmt.Printf("总查询次数: %d\n", metrics.TotalQueries)
fmt.Printf("缓存命中率: %.1f%%\n", metrics.HitRate())
fmt.Printf("平均延迟: %v\n", metrics.AverageLatency)
fmt.Printf("LSP调用次数: %d\n", metrics.LSPCalls)
```

### 5. 缓存管理API

#### 项目级缓存管理
```go
func (p *Project) GetCacheStats() CacheStats
func (p *Project) ClearReferenceCache()
```

**示例:**
```go
// 获取缓存统计
stats := project.GetCacheStats()
fmt.Printf("缓存条目数: %d\n", stats.TotalEntries)
fmt.Printf("总访问次数: %d\n", stats.TotalAccesses)
fmt.Printf("过期条目数: %d\n", stats.ExpiredEntries)

// 清空缓存
project.ClearReferenceCache()
```

#### 缓存配置
```go
func NewReferenceCache(maxEntries int, ttl time.Duration) *ReferenceCache
```

**示例:**
```go
// 创建自定义缓存
cache := NewReferenceCache(1000, 30*time.Minute) // 1000条目，30分钟TTL
```

## 错误处理

### ReferenceError 类型
```go
type ReferenceError struct {
    Type        ReferenceErrorType `json:"type"`
    Message     string              `json:"message"`
    Cause       error               `json:"cause,omitempty"`
    NodeInfo    string              `json:"nodeInfo,omitempty"`
    FilePath    string              `json:"filePath,omitempty"`
    LineNumber  int                 `json:"lineNumber,omitempty"`
    Retryable   bool                `json:"retryable"`
    RetryCount  int                 `json:"retryCount"`
    Timestamp   time.Time           `json:"timestamp"`
}
```

### 错误类型
- `ErrorTypeLSPService`: LSP服务错误
- `ErrorTypeLSPTimeout`: LSP超时
- `ErrorTypeLSPUnavailable`: LSP不可用
- `ErrorTypeProjectNotFound`: 项目未找到
- `ErrorTypeFileNotFound`: 文件未找到
- `ErrorTypeInvalidNode`: 无效节点
- `ErrorTypeCacheCorruption`: 缓存损坏
- `ErrorTypeCacheExpired`: 缓存过期
- `ErrorTypeSyntaxError`: 语法错误
- `ErrorTypeTypeCheckError`: 类型检查错误

### 错误处理示例
```go
refs, fromCache, err := FindReferencesWithCache(*targetNode)
if err != nil {
    if refErr, ok := err.(*ReferenceError); ok {
        log.Printf("错误类型: %s", refErr.getErrorTypeName())
        log.Printf("可重试: %t, 重试次数: %d", refErr.Retryable, refErr.RetryCount)

        if refErr.ShouldUseFallback() {
            // 尝试降级策略
            fallbackRefs := FindReferencesFallback(*targetNode)
            if len(fallbackRefs) > 0 {
                fmt.Printf("降级策略找到 %d 个引用\n", len(fallbackRefs))
            }
        }
    }
    return
}
```

## 重试机制

### RetryConfig
```go
type RetryConfig struct {
    MaxRetries    int           `json:"maxRetries"`
    BaseDelay     time.Duration `json:"baseDelay"`
    MaxDelay      time.Duration `json:"maxDelay"`
    BackoffFactor float64       `json:"backoffFactor"`
    Enabled       bool          `json:"enabled"`
}
```

### 默认配置
```go
config := DefaultRetryConfig()
// MaxRetries: 3
// BaseDelay: 100ms
// MaxDelay: 5s
// BackoffFactor: 2.0
// Enabled: true
```

### 自定义重试配置
```go
config := &RetryConfig{
    MaxRetries:    5,
    BaseDelay:     200 * time.Millisecond,
    MaxDelay:      10 * time.Second,
    BackoffFactor: 1.5,
    Enabled:       true,
}

refs, fromCache, err := FindReferencesWithCacheAndRetry(*targetNode, config)
```

## 降级策略

当LSP服务不可用时，系统会自动降级到本地AST分析：

```go
// 手动调用降级策略
refs := FindReferencesFallback(*targetNode)
defs := GotoDefinitionFallback(*targetNode)
```

### 上下文分析
```go
// 判断节点是否可能是引用或定义
isRef := IsLikelyReference(node)
isDef := IsLikelyDefinition(node)
```

## 性能优化建议

### 1. 使用缓存
对于频繁的引用查找，始终使用带缓存的API：
```go
// ✅ 推荐：使用缓存
refs, fromCache, err := FindReferencesWithCache(*targetNode)

// ❌ 不推荐：直接调用LSP
refs, err := FindReferences(*targetNode)
```

### 2. 批量处理
对于多个节点的查找，使用批量API：
```go
// ✅ 推荐：批量处理
results, err := FindReferencesBatch(nodes)

// ❌ 不推荐：逐个处理
for _, node := range nodes {
    refs, err := FindReferences(node)
}
```

### 3. 性能监控
在生产环境中使用指标收集器监控性能：
```go
collector := NewMetricsCollector(project)
// ... 执行多次查询
metrics := collector.GetMetrics()
log.Printf("缓存命中率: %.1f%%", metrics.HitRate())
```

### 4. 缓存管理
根据应用场景调整缓存配置：
```go
// 对于大型项目，增加缓存大小
cache := NewReferenceCache(5000, 30*time.Minute)

// 对于频繁变更的项目，减少TTL
cache := NewReferenceCache(1000, 5*time.Minute)
```

## 完整示例

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
    // 创建项目
    project := tsmorphgo.NewProject(".", &tsmorphgo.ProjectOptions{})

    // 添加源文件
    sourceFile := project.AddSourceFile("example.ts", `
        const sharedVar = "hello";

        function testFunction() {
            console.log(sharedVar);
        }

        testFunction();
        console.log(sharedVar);
    `)

    // 找到目标节点
    var targetNode *tsmorphgo.Node
    sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
        if tsmorphgo.IsIdentifier(node) && node.GetText() == "sharedVar" {
            parent := node.GetParent()
            if parent != nil && parent.Kind != 164 { // 不是变量声明
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

    // 配置重试
    retryConfig := &tsmorphgo.RetryConfig{
        MaxRetries:    3,
        BaseDelay:     100 * time.Millisecond,
        MaxDelay:      5 * time.Second,
        BackoffFactor: 2.0,
        Enabled:       true,
    }

    // 执行多次查找测试性能
    for i := 0; i < 5; i++ {
        fmt.Printf("\n=== 第 %d 次查找 ===\n", i+1)

        // 使用缓存和重试机制
        refs, fromCache, err := tsmorphgo.FindReferencesWithCacheAndRetry(*targetNode, retryConfig)
        if err != nil {
            if refErr, ok := err.(*tsmorphgo.ReferenceError); ok {
                fmt.Printf("错误: %s (重试 %d 次)\n", refErr.Error(), refErr.RetryCount)

                // 尝试降级策略
                fallbackRefs := tsmorphgo.FindReferencesFallback(*targetNode)
                if len(fallbackRefs) > 0 {
                    fmt.Printf("降级策略找到 %d 个引用\n", len(fallbackRefs))
                }
            }
            continue
        }

        fmt.Printf("找到 %d 个引用", len(refs))
        if fromCache {
            fmt.Printf(" (来自缓存)")
        } else {
            fmt.Printf(" (来自LSP)")
        }
        fmt.Println()

        // 显示引用位置
        for j, ref := range refs {
            fmt.Printf("  引用 %d: 行 %d, 列 %d\n",
                j+1, ref.GetStartLineNumber(), ref.GetStartColumnNumber())
        }

        // 收集性能指标
        _, err = collector.FindReferencesWithMetrics(*targetNode)
        if err != nil {
            log.Printf("指标收集失败: %v", err)
        }
    }

    // 显示性能统计
    metrics := collector.GetMetrics()
    fmt.Printf("\n=== 性能统计 ===\n")
    fmt.Printf("总查询次数: %d\n", metrics.TotalQueries)
    fmt.Printf("缓存命中次数: %d\n", metrics.CacheHits)
    fmt.Printf("LSP调用次数: %d\n", metrics.LSPCalls)
    fmt.Printf("缓存命中率: %.1f%%\n", metrics.HitRate())
    fmt.Printf("平均延迟: %v\n", metrics.AverageLatency)

    // 显示缓存统计
    cacheStats := project.GetCacheStats()
    fmt.Printf("\n=== 缓存统计 ===\n")
    fmt.Printf("缓存条目数: %d\n", cacheStats.TotalEntries)
    fmt.Printf("总访问次数: %d\n", cacheStats.TotalAccesses)
    fmt.Printf("过期条目数: %d\n", cacheStats.ExpiredEntries)
}
```

## 最佳实践

1. **优先使用缓存API**: 对于生产环境，始终使用 `FindReferencesWithCache` 和 `GotoDefinitionWithCache`

2. **合理配置重试**: 根据网络环境和LSP服务稳定性调整重试配置

3. **监控性能**: 使用 `MetricsCollector` 持续监控缓存命中率和响应时间

4. **处理降级**: 实现适当的降级策略处理LSP服务不可用的情况

5. **批量操作**: 对于大量节点的查找，使用 `FindReferencesBatch` 减少LSP调用

6. **缓存管理**: 根据项目规模和变更频率调整缓存大小和TTL

7. **错误处理**: 实现完善的错误处理，区分可重试和不可重试的错误类型