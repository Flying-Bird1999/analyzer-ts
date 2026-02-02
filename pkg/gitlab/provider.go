// Package gitlab provides GitLab integration capabilities for analyzer-ts.
package gitlab

import (
	"context"
	"fmt"
	"io"
	"os"
)

// =============================================================================
// Provider - Diff 数据提供者接口
// =============================================================================

// Provider diff 数据提供者接口
// 支持多种输入方式：字符串、API、文件
type Provider interface {
	// GetDiffString 获取 diff 字符串
	GetDiffString(ctx context.Context) (string, error)
}

// =============================================================================
// StringProvider - 从字符串获取 diff
// =============================================================================

// StringProvider 从字符串获取 diff
type StringProvider struct {
	diff string
}

// NewStringProvider 创建字符串提供者
func NewStringProvider(diff string) *StringProvider {
	return &StringProvider{diff: diff}
}

// GetDiffString 实现 Provider 接口
func (p *StringProvider) GetDiffString(ctx context.Context) (string, error) {
	return p.diff, nil
}

// =============================================================================
// APIProvider - 从 GitLab API 获取 diff
// =============================================================================

// APIProvider 从 GitLab API 获取 diff
type APIProvider struct {
	client    *Client
	projectID int
	mrIID     int
}

// NewAPIProvider 创建 API 提供者
func NewAPIProvider(client *Client, projectID, mrIID int) *APIProvider {
	return &APIProvider{
		client:    client,
		projectID: projectID,
		mrIID:     mrIID,
	}
}

// GetDiffString 实现 Provider 接口
func (p *APIProvider) GetDiffString(ctx context.Context) (string, error) {
	diffFiles, err := p.client.GetMergeRequestDiff(ctx, p.projectID, p.mrIID)
	if err != nil {
		return "", fmt.Errorf("get MR diff failed: %w", err)
	}

	// 将 DiffFile 列表转换为 patch 格式字符串
	return p.convertToPatch(diffFiles), nil
}

// convertToPatch 将 DiffFile 列表转换为 patch 格式字符串
func (p *APIProvider) convertToPatch(diffFiles []DiffFile) string {
	var result string
	for _, df := range diffFiles {
		result += df.Diff + "\n"
	}
	return result
}

// =============================================================================
// FileProvider - 从文件获取 diff
// =============================================================================

// FileProvider 从文件获取 diff
type FileProvider struct {
	filePath string
}

// NewFileProvider 创建文件提供者
func NewFileProvider(filePath string) *FileProvider {
	return &FileProvider{filePath: filePath}
}

// GetDiffString 实现 Provider 接口
func (p *FileProvider) GetDiffString(ctx context.Context) (string, error) {
	content, err := os.ReadFile(p.filePath)
	if err != nil {
		return "", fmt.Errorf("read diff file failed: %w", err)
	}

	return string(content), nil
}

// =============================================================================
// ReaderProvider - 从 io.Reader 获取 diff
// =============================================================================

// ReaderProvider 从 io.Reader 获取 diff
type ReaderProvider struct {
	reader io.Reader
}

// NewReaderProvider 创建 Reader 提供者
func NewReaderProvider(r io.Reader) *ReaderProvider {
	return &ReaderProvider{reader: r}
}

// GetDiffString 实现 Provider 接口
func (p *ReaderProvider) GetDiffString(ctx context.Context) (string, error) {
	data, err := io.ReadAll(p.reader)
	if err != nil {
		return "", fmt.Errorf("read diff failed: %w", err)
	}

	return string(data), nil
}
