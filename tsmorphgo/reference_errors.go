package tsmorphgo

import (
	"fmt"
	"strings"
	"time"
)

// ReferenceError 引用查找错误的详细分类
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
	if e.Type != ErrorTypeUnknown {
		parts = append(parts, fmt.Sprintf("[%s]", e.getErrorTypeName()))
	}
	parts = append(parts, e.Message)

	if e.NodeInfo != "" {
		parts = append(parts, fmt.Sprintf("(节点: %s)", e.NodeInfo))
	}

	if e.FilePath != "" {
		parts = append(parts, fmt.Sprintf("(文件: %s:%d)", e.FilePath, e.LineNumber))
	}

	if e.Retryable {
		parts = append(parts, fmt.Sprintf("(可重试, 已重试 %d 次)", e.RetryCount))
	}

	return strings.Join(parts, " ")
}

// getErrorTypeName 获取错误类型名称
func (e ReferenceError) getErrorTypeName() string {
	switch e.Type {
	case ErrorTypeLSPService:
		return "LSP服务错误"
	case ErrorTypeLSPTimeout:
		return "LSP超时"
	case ErrorTypeLSPUnavailable:
		return "LSP不可用"
	case ErrorTypeProjectNotFound:
		return "项目未找到"
	case ErrorTypeFileNotFound:
		return "文件未找到"
	case ErrorTypeInvalidNode:
		return "无效节点"
	case ErrorTypeCacheCorruption:
		return "缓存损坏"
	case ErrorTypeCacheExpired:
		return "缓存过期"
	case ErrorTypeSyntaxError:
		return "语法错误"
	case ErrorTypeTypeCheckError:
		return "类型检查错误"
	default:
		return "未知错误"
	}
}

// IsRetryable 判断错误是否可重试
func (e ReferenceError) IsRetryable() bool {
	return e.Retryable && e.RetryCount < 3 // 最多重试3次
}

// ShouldUseFallback 判断是否应该使用降级策略
func (e ReferenceError) ShouldUseFallback() bool {
	// 某些错误类型应该触发降级策略
	switch e.Type {
	case ErrorTypeLSPTimeout, ErrorTypeLSPUnavailable, ErrorTypeLSPService:
		return true
	case ErrorTypeFileNotFound:
		return false // 文件不存在不应该降级
	default:
		return e.RetryCount >= 3 // 重试次数过多才降级
	}
}

// NewReferenceError 创建新的引用错误
func NewReferenceError(errorType ReferenceErrorType, message string, cause error) *ReferenceError {
	return &ReferenceError{
		Type:      errorType,
		Message:   message,
		Cause:     cause,
		Retryable: isRetryableErrorType(errorType),
		Timestamp: time.Now(),
	}
}

// NewReferenceErrorWithNode 创建带节点信息的引用错误
func NewReferenceErrorWithNode(errorType ReferenceErrorType, message string, node Node, cause error) *ReferenceError {
	nodeInfo := ""
	filePath := ""
	lineNumber := 0

	if node.IsValid() {
		nodeInfo = strings.TrimSpace(node.GetText())
		if nodeInfo == "" {
			nodeInfo = fmt.Sprintf("节点类型: %v", node.Kind)
		}

		if sourceFile := node.GetSourceFile(); sourceFile != nil {
			filePath = sourceFile.GetFilePath()
			lineNumber = node.GetStartLineNumber()
		}
	}

	return &ReferenceError{
		Type:       errorType,
		Message:    message,
		Cause:      cause,
		NodeInfo:   nodeInfo,
		FilePath:   filePath,
	LineNumber: lineNumber,
		Retryable:  isRetryableErrorType(errorType),
		Timestamp:  time.Now(),
	}
}

// isRetryableErrorType 判断错误类型是否可重试
func isRetryableErrorType(errorType ReferenceErrorType) bool {
	switch errorType {
	case ErrorTypeLSPService, ErrorTypeLSPTimeout, ErrorTypeLSPUnavailable:
		return true
	case ErrorTypeFileNotFound, ErrorTypeInvalidNode, ErrorTypeCacheCorruption:
		return false
	default:
		return true // 默认可重试
	}
}

// IncrementRetryCount 增加重试计数
func (e *ReferenceError) IncrementRetryCount() {
	e.RetryCount++
	e.Timestamp = time.Now()
}

// WithRetryable 设置重试标志
func (e *ReferenceError) WithRetryable(retryable bool) *ReferenceError {
	e.Retryable = retryable
	return e
}

// WrapError 包装通用错误为引用错误
func WrapError(err error, context string, node Node) *ReferenceError {
	if err == nil {
		return nil
	}

	errorMsg := err.Error()
	errorType := ErrorTypeUnknown

	// 根据错误消息推断错误类型
	errorMsgLower := strings.ToLower(errorMsg)
	if strings.Contains(errorMsgLower, "timeout") {
		errorType = ErrorTypeLSPTimeout
	} else if strings.Contains(errorMsgLower, "unavailable") {
		errorType = ErrorTypeLSPUnavailable
	} else if strings.Contains(errorMsgLower, "service") {
		errorType = ErrorTypeLSPService
	} else if strings.Contains(errorMsgLower, "file not found") {
		errorType = ErrorTypeFileNotFound
	} else if strings.Contains(errorMsgLower, "syntax") {
		errorType = ErrorTypeSyntaxError
	}

	return NewReferenceErrorWithNode(errorType, fmt.Sprintf("%s: %s", context, errorMsg), node, err)
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries    int           `json:"maxRetries"`    // 最大重试次数
	BaseDelay     time.Duration `json:"baseDelay"`     // 基础延迟
	MaxDelay      time.Duration `json:"maxDelay"`      // 最大延迟
	BackoffFactor  float64       `json:"backoffFactor"`  // 退避因子
	Enabled       bool          `json:"enabled"`       // 是否启用重试
}

// DefaultRetryConfig 默认重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
	MaxRetries:   3,
		BaseDelay:    100 * time.Millisecond,
	MaxDelay:     5 * time.Second,
		BackoffFactor: 2.0,
		Enabled:      true,
	}
}

// CalculateDelay 计算重试延迟（指数退避）
func (rc *RetryConfig) CalculateDelay(retryCount int) time.Duration {
	if !rc.Enabled {
		return 0
	}

	if retryCount <= 0 {
		return rc.BaseDelay
	}

	delay := time.Duration(float64(rc.BaseDelay) * (rc.BackoffFactor * float64(retryCount)))
	if delay > rc.MaxDelay {
		delay = rc.MaxDelay
	}

	return delay
}