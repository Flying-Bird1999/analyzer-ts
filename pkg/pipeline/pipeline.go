package pipeline

import (
	"fmt"
	"time"
)

// =============================================================================
// AnalysisPipeline 分析管道
// =============================================================================

// AnalysisPipeline 分析管道，按顺序执行多个分析阶段。
type AnalysisPipeline struct {
	stages []Stage
}

// NewPipeline 创建一个新的分析管道。
func NewPipeline(name string) *AnalysisPipeline {
	return &AnalysisPipeline{
		stages: make([]Stage, 0),
	}
}

// AddStage 添加一个分析阶段到管道。
func (p *AnalysisPipeline) AddStage(stage Stage) *AnalysisPipeline {
	p.stages = append(p.stages, stage)
	return p
}

// Execute 执行管道中的所有阶段。
func (p *AnalysisPipeline) Execute(ctx *AnalysisContext) (*PipelineResult, error) {
	results := make([]*StageResult, 0, len(p.stages))

	fmt.Printf("🚀 开始执行分析管道，共 %d 个阶段\n", len(p.stages))

	for i, stage := range p.stages {
		stageName := stage.Name()
		fmt.Printf("\n[阶段 %d/%d] %s\n", i+1, len(p.stages), stageName)

		// 检查是否跳过
		if stage.Skip(ctx) {
			fmt.Printf("  ⊘ 跳过阶段\n")
			results = append(results, NewSkippedResult(stageName, "配置要求未满足"))
			continue
		}

		// 检查取消
		if ctx.IsCanceled() {
			return nil, ctx.Cancel.Err()
		}

		// 执行阶段
		startTime := time.Now()
		result, err := stage.Execute(ctx)
		duration := time.Since(startTime)

		if err != nil {
			fmt.Printf("  ❌ 执行失败 (耗时: %s): %v\n", duration, err)
			results = append(results, NewErrorResult(stageName, err))
			// 继续执行还是中断？这里选择中断
			return nil, fmt.Errorf("stage %s failed: %w", stageName, err)
		}

		// 存储结果
		ctx.SetResult(stageName, result)
		results = append(results, NewSuccessResult(stageName, result))
		fmt.Printf("  ✅ 完成 (耗时: %s)\n", duration)

		// 打印简要统计
		if printer, ok := stage.(ResultPrinter); ok {
			fmt.Print("     ")
			printer.PrintResult(result)
		}
	}

	fmt.Printf("\n✅ 管道执行完成\n")
	return NewPipelineResult(results), nil
}

// =============================================================================
// PipelineResult 管道结果
// =============================================================================

// PipelineResult 表示管道的执行结果。
type PipelineResult struct {
	Results []*StageResult
}

// NewPipelineResult 创建管道结果。
func NewPipelineResult(results []*StageResult) *PipelineResult {
	return &PipelineResult{
		Results: results,
	}
}

// GetResult 获取指定阶段的结果。
func (r *PipelineResult) GetResult(stageName string) (interface{}, bool) {
	for _, result := range r.Results {
		if result.StageName == stageName {
			if result.Error != nil {
				return nil, false
			}
			if result.Skipped {
				return nil, false
			}
			return result.Result, true
		}
	}
	return nil, false
}

// MustGetResult 获取指定阶段的结果，如果不存在则 panic。
func (r *PipelineResult) MustGetResult(stageName string) interface{} {
	result, exists := r.GetResult(stageName)
	if !exists {
		panic(fmt.Sprintf("stage result not found or failed: %s", stageName))
	}
	return result
}

// IsSuccessful 检查管道是否全部成功执行。
// 跳过的阶段（Skipped=true）不影响成功状态。
func (r *PipelineResult) IsSuccessful() bool {
	for _, result := range r.Results {
		if result.Error != nil && !result.Skipped {
			return false
		}
	}
	return true
}

// GetErrors 获取所有阶段的错误。
func (r *PipelineResult) GetErrors() []error {
	errors := make([]error, 0)
	for _, result := range r.Results {
		if result.Error != nil {
			errors = append(errors, result.Error)
		}
	}
	return errors
}

// =============================================================================
// ResultPrinter 结果打印器接口
// =============================================================================

// ResultPrinter 可以打印阶段结果的简要信息。
type ResultPrinter interface {
	Stage
	PrintResult(result interface{})
}
