// Package gitlab provides GitLab integration capabilities for analyzer-ts.
package gitlab

import (
	"context"
	"fmt"
)

// =============================================================================
// MR 结果发布器
// =============================================================================

// MRPoster MR 结果发布器
// 负责将分析结果发布到 GitLab MR
type MRPoster struct {
	mrService *MRService
	formatter *Formatter
}

// NewMRPoster 创建 MR 结果发布器
func NewMRPoster(client *Client, projectID, mrIID int) *MRPoster {
	return &MRPoster{
		mrService: NewMRService(client, projectID, mrIID),
		formatter: NewFormatter(CommentStyleDetailed),
	}
}

// PostResult 发布分析结果到 MR
func (p *MRPoster) PostResult(ctx context.Context, result interface{}) error {
	// 尝试将结果转换为 ImpactAnalysisResult
	if impactResult, ok := result.(*AnalysisResult); ok {
		return p.mrService.PostImpactComment(ctx, impactResult)
	}

	// 如果是其他类型，返回错误
	return fmt.Errorf("unsupported result type: %T", result)
}

// PostResultWithFormatter 使用自定义格式化器发布结果
func (p *MRPoster) PostResultWithFormatter(ctx context.Context, result interface{}, formatter *Formatter) error {
	if impactResult, ok := result.(*AnalysisResult); ok {
		comment, err := formatter.FormatImpactResult(impactResult)
		if err != nil {
			return fmt.Errorf("format impact result failed: %w", err)
		}

		// 先尝试更新现有评论
		existingComment, err := p.mrService.FindAnalyzerComment(ctx)
		if err == nil && existingComment != nil {
			return p.mrService.updateComment(ctx, existingComment.NoteID, comment)
		}

		// 创建新评论
		return p.mrService.createComment(ctx, comment)
	}

	return fmt.Errorf("unsupported result type: %T", result)
}

// GetMRService 获取 MR 服务（用于更高级的用法）
func (p *MRPoster) GetMRService() *MRService {
	return p.mrService
}
