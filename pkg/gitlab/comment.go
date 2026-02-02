// Package gitlab provides GitLab integration capabilities for analyzer-ts.
package gitlab

import (
	"context"
)

// =============================================================================
// Service - 评论发布服务
// =============================================================================

// Service 评论服务
// 提供纯字符串发布功能，不耦合格式化逻辑
type Service struct {
	client    *Client
	projectID int
	mrIID     int
}

// NewService 创建评论服务
func NewService(client *Client, projectID, mrIID int) *Service {
	return &Service{
		client:    client,
		projectID: projectID,
		mrIID:     mrIID,
	}
}

// =============================================================================
// 公共方法
// =============================================================================

// PostComment 发布评论（纯字符串）
func (s *Service) PostComment(ctx context.Context, body string) error {
	return s.client.CreateMRComment(ctx, s.projectID, s.mrIID, body)
}
