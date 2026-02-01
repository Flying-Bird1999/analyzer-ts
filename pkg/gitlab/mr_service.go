// Package gitlab provides GitLab integration capabilities for analyzer-ts.
package gitlab

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// =============================================================================
// MRService - MR 高层服务
// =============================================================================

// MRService MR 服务
type MRService struct {
	client    *Client
	projectID int
	mrIID     int
}

// NewMRService 创建 MR 服务
func NewMRService(client *Client, projectID int, mrIID int) *MRService {
	return &MRService{
		client:    client,
		projectID: projectID,
		mrIID:     mrIID,
	}
}

// =============================================================================
// 公共方法
// =============================================================================

// FindAnalyzerComment 查找分析器之前发布的评论
// 通过评论内容中的标记来识别
func (s *MRService) FindAnalyzerComment(ctx context.Context) (*Comment, error) {
	comments, err := s.client.ListMRComments(ctx, s.projectID, s.mrIID)
	if err != nil {
		return nil, err
	}

	// 查找包含标记的评论
	marker := "<!-- analyzer-ts-impact-report -->"
	for i := len(comments) - 1; i >= 0; i-- {
		if strings.Contains(comments[i].Body, marker) {
			return &comments[i], nil
		}
	}

	return nil, nil
}

// PostImpactComment 发布影响分析评论到 MR
func (s *MRService) PostImpactComment(ctx context.Context, result *ImpactAnalysisResult) error {
	// 格式化为 Markdown
	formatter := NewFormatter(CommentStyleDetailed)
	comment, err := formatter.FormatImpactResult(result)
	if err != nil {
		return fmt.Errorf("format impact result failed: %w", err)
	}

	// 先尝试更新现有评论
	existingComment, err := s.FindAnalyzerComment(ctx)
	if err == nil && existingComment != nil {
		return s.updateComment(ctx, existingComment.NoteID, comment)
	}

	// 如果没有找到现有评论，创建新评论
	return s.createComment(ctx, comment)
}

// updateComment 更新现有评论
func (s *MRService) updateComment(ctx context.Context, noteID int, body string) error {
	return s.client.UpdateMRComment(ctx, s.projectID, s.mrIID, noteID, body)
}

// createComment 创建新评论
func (s *MRService) createComment(ctx context.Context, body string) error {
	return s.client.CreateMRComment(ctx, s.projectID, s.mrIID, body)
}

// DeleteOldComments 删除之前的分析器评论（可选，用于清理）
func (s *MRService) DeleteOldComments(ctx context.Context) error {
	comments, err := s.client.ListMRComments(ctx, s.projectID, s.mrIID)
	if err != nil {
		return err
	}

	marker := "<!-- analyzer-ts-impact-report -->"

	for _, comment := range comments {
		if strings.Contains(comment.Body, marker) {
			// 删除评论
			url := fmt.Sprintf("/api/v4/projects/%d/merge_requests/%d/notes/%d",
				s.projectID, s.mrIID, comment.NoteID)
			// TODO: 实现 DELETE 请求
			_ = url
		}
	}

	return nil
}

// =============================================================================
// 内部方法
// =============================================================================

// extractNoteID 从评论 URL 中提取 note ID
func extractNoteID(commentURL string) (int, error) {
	// URL 格式: /api/v4/projects/123/merge_requests/456/notes/789
	pattern := regexp.MustCompile(`/notes/(\d+)$`)
	matches := pattern.FindStringSubmatch(commentURL)
	if len(matches) < 2 {
		return 0, fmt.Errorf("cannot extract note ID from URL: %s", commentURL)
	}

	return strconv.Atoi(matches[1])
}

// getCommentID 从 Comment 结构中获取 note ID
func getCommentID(comment *Comment) (int, error) {
	if comment.NoteID != 0 {
		return comment.NoteID, nil
	}

	// 如果 NoteID 为 0，尝试从 URL 中提取
	return extractNoteID(comment.ID)
}
