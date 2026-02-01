// Package pipeline 提供代码分析管道协调功能。
// 它作为各个分析模块之间的协调层，负责串联符号分析、依赖分析、影响分析等步骤。
package pipeline

import (
	"context"
	"fmt"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

// =============================================================================
// 分析上下文
// =============================================================================

// AnalysisContext 表示分析管道的执行上下文。
// 它在整个分析过程中传递，包含项目信息、配置和中间结果。
type AnalysisContext struct {
	// 项目信息
	ProjectRoot    string                // 项目根目录
	Project        *tsmorphgo.Project     // tsmorphgo 项目实例
	ExcludePaths   []string              // 排除的路径模式

	// 配置选项
	Options map[string]interface{} // 分析器配置选项

	// 中间结果
	intermediateResults map[string]interface{} // 各阶段的输出结果

	// 控制选项
	Cancel context.Context // 取消信号
}

// NewAnalysisContext 创建一个新的分析上下文。
func NewAnalysisContext(ctx context.Context, projectRoot string, project *tsmorphgo.Project) *AnalysisContext {
	return &AnalysisContext{
		ProjectRoot:        projectRoot,
		Project:            project,
		ExcludePaths:       []string{},
		Options:            make(map[string]interface{}),
		intermediateResults: make(map[string]interface{}),
		Cancel:             ctx,
	}
}

// SetOption 设置配置选项。
func (ctx *AnalysisContext) SetOption(key string, value interface{}) {
	ctx.Options[key] = value
}

// GetOption 获取配置选项。
func (ctx *AnalysisContext) GetOption(key string, defaultValue interface{}) interface{} {
	if value, exists := ctx.Options[key]; exists {
		return value
	}
	return defaultValue
}

// SetResult 存储中间结果。
func (ctx *AnalysisContext) SetResult(stageName string, result interface{}) {
	ctx.intermediateResults[stageName] = result
}

// GetResult 获取中间结果。
func (ctx *AnalysisContext) GetResult(stageName string) (interface{}, bool) {
	result, exists := ctx.intermediateResults[stageName]
	return result, exists
}

// MustGetResult 获取中间结果，如果不存在则 panic。
func (ctx *AnalysisContext) MustGetResult(stageName string) interface{} {
	result, exists := ctx.intermediateResults[stageName]
	if !exists {
		panic(fmt.Sprintf("stage result not found: %s", stageName))
	}
	return result
}

// IsCanceled 检查是否已取消。
func (ctx *AnalysisContext) IsCanceled() bool {
	select {
	case <-ctx.Cancel.Done():
		return true
	default:
		return false
	}
}

// =============================================================================
// 阶段接口
// =============================================================================

// Stage 表示分析管道中的一个阶段。
type Stage interface {
	// Name 返回阶段名称
	Name() string

	// Execute 执行该阶段的分析逻辑
	Execute(ctx *AnalysisContext) (interface{}, error)

	// Skip 判断是否跳过该阶段
	Skip(ctx *AnalysisContext) bool
}

// =============================================================================
// StageResult 阶段结果
// =============================================================================

// StageResult 表示一个阶段的执行结果。
type StageResult struct {
	StageName string      // 阶段名称
	Result    interface{} // 阶段结果
	Error     error      // 执行错误（如果有）
	Skipped   bool       // 是否跳过
}

// NewSuccessResult 创建一个成功的结果。
func NewSuccessResult(stageName string, result interface{}) *StageResult {
	return &StageResult{
		StageName: stageName,
		Result:    result,
	}
}

// NewSkippedResult 创建一个跳过的结果。
func NewSkippedResult(stageName string, reason string) *StageResult {
	return &StageResult{
		StageName: stageName,
		Skipped:   true,
		Error:     fmt.Errorf("skipped: %s", reason),
	}
}

// NewErrorResult 创建一个错误的结果。
func NewErrorResult(stageName string, err error) *StageResult {
	return &StageResult{
		StageName: stageName,
		Error:     err,
	}
}
