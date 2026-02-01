// Package gitlab provides GitLab-specific capabilities for analyzer-ts.
package gitlab

import (
	"context"
	"fmt"
	"io"
)

// =============================================================================
// GitLab Diff 提供者
// =============================================================================

// DiffProvider GitLab diff 提供者
// 负责从 GitLab API 获取 MR 的 diff
type DiffProvider struct {
	client    *Client
	projectID int
	mrIID     int
}

// NewDiffProvider 创建 diff 提供者
func NewDiffProvider(client *Client, projectID, mrIID int) *DiffProvider {
	return &DiffProvider{
		client:    client,
		projectID: projectID,
		mrIID:     mrIID,
	}
}

// GetDiffFiles 从 GitLab API 获取 diff 文件列表
func (p *DiffProvider) GetDiffFiles(ctx context.Context) ([]DiffFile, error) {
	return p.client.GetMergeRequestDiff(ctx, p.projectID, p.mrIID)
}

// GetDiffAsPatch 从 GitLab API 获取 diff 并转换为 patch 格式
func (p *DiffProvider) GetDiffAsPatch(ctx context.Context) (string, error) {
	diffFiles, err := p.GetDiffFiles(ctx)
	if err != nil {
		return "", fmt.Errorf("get diff files failed: %w", err)
	}

	// 将 DiffFile 列表转换为 patch 格式
	patch := p.convertToPatch(diffFiles)
	return patch, nil
}

// WriteDiffToFile 将 diff 写入文件
func (p *DiffProvider) WriteDiffToFile(ctx context.Context, filePath string) error {
	_, err := p.GetDiffAsPatch(ctx)
	if err != nil {
		return err
	}

	// TODO: 写入文件
	return fmt.Errorf("not implemented")
}

// convertToPatch 将 DiffFile 列表转换为 patch 格式字符串
func (p *DiffProvider) convertToPatch(diffFiles []DiffFile) string {
	var result string
	for _, df := range diffFiles {
		result += df.Diff + "\n"
	}
	return result
}

// =============================================================================
// Diff 输入源（供 pipeline 使用）
// =============================================================================

// DiffInputSource diff 输入源类型
type DiffInputSource interface {
	// GetDiffFiles 获取 diff 文件列表
	GetDiffFiles(ctx context.Context) ([]DiffFile, error)

	// GetPatch 获取 patch 格式的 diff
	GetPatch(ctx context.Context) (string, error)
}

// GitLabDiffSource GitLab diff 输入源
type GitLabDiffSource struct {
	provider *DiffProvider
}

// NewGitLabDiffSource 创建 GitLab diff 输入源
func NewGitLabDiffSource(client *Client, projectID, mrIID int) *GitLabDiffSource {
	return &GitLabDiffSource{
		provider: NewDiffProvider(client, projectID, mrIID),
	}
}

// GetDiffFiles 实现 DiffInputSource 接口
func (s *GitLabDiffSource) GetDiffFiles(ctx context.Context) ([]DiffFile, error) {
	return s.provider.GetDiffFiles(ctx)
}

// GetPatch 实现 DiffInputSource 接口
func (s *GitLabDiffSource) GetPatch(ctx context.Context) (string, error) {
	return s.provider.GetDiffAsPatch(ctx)
}

// =============================================================================
// 文件 Diff 输入源
// =============================================================================

// FileDiffSource 文件 diff 输入源
type FileDiffSource struct {
	filePath string
}

// NewFileDiffSource 创建文件 diff 输入源
func NewFileDiffSource(filePath string) *FileDiffSource {
	return &FileDiffSource{filePath: filePath}
}

// GetDiffFiles 实现 DiffInputSource 接口
func (s *FileDiffSource) GetDiffFiles(ctx context.Context) ([]DiffFile, error) {
	// 读取文件并解析为 DiffFile 列表
	// TODO: 实现文件解析
	return nil, fmt.Errorf("not implemented")
}

// GetPatch 实现 DiffInputSource 接口
func (s *FileDiffSource) GetPatch(ctx context.Context) (string, error) {
	// 读取文件内容
	// TODO: 实现文件读取
	return "", fmt.Errorf("not implemented")
}

// =============================================================================
// Reader Diff 输入源（从 io.Reader 读取）
// =============================================================================

// ReaderDiffSource 从 io.Reader 读取 diff
type ReaderDiffSource struct {
	reader io.Reader
}

// NewReaderDiffSource 创建 Reader diff 输入源
func NewReaderDiffSource(r io.Reader) *ReaderDiffSource {
	return &ReaderDiffSource{reader: r}
}

// GetDiffFiles 实现 DiffInputSource 接口
func (s *ReaderDiffSource) GetDiffFiles(ctx context.Context) ([]DiffFile, error) {
	// TODO: 实现 Reader 解析
	return nil, fmt.Errorf("not implemented")
}

// GetPatch 实现 DiffInputSource 接口
func (s *ReaderDiffSource) GetPatch(ctx context.Context) (string, error) {
	// 读取全部内容
	data, err := io.ReadAll(s.reader)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
