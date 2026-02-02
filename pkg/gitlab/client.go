// Package gitlab provides GitLab integration capabilities for analyzer-ts.
package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// =============================================================================
// Client - GitLab API 客户端
// =============================================================================

// Client GitLab API 客户端
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient 创建 GitLab 客户端
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// =============================================================================
// API 方法
// =============================================================================

// GetMergeRequest 获取 MR 详情
func (c *Client) GetMergeRequest(ctx context.Context, projectID int, mrIID int) (*MergeRequest, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/merge_requests/%d", c.baseURL, projectID, mrIID)

	req, err := c.createRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var mr MergeRequest
	if err := json.NewDecoder(resp.Body).Decode(&mr); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	return &mr, nil
}

// GetMergeRequestDiff 获取 MR 的 diff
func (c *Client) GetMergeRequestDiff(ctx context.Context, projectID int, mrIID int) ([]DiffFile, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/merge_requests/%d/diff", c.baseURL, projectID, mrIID)

	req, err := c.createRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var diffs []DiffFile
	if err := json.NewDecoder(resp.Body).Decode(&diffs); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	return diffs, nil
}

// CreateMRComment 在 MR 上创建评论
func (c *Client) CreateMRComment(ctx context.Context, projectID int, mrIID int, body string) error {
	url := fmt.Sprintf("%s/api/v4/projects/%d/merge_requests/%d/notes", c.baseURL, projectID, mrIID)

	payload := map[string]interface{}{
		"body": body,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload failed: %w", err)
	}

	req, err := c.createRequestWithBody(ctx, "POST", url, payloadBytes)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// =============================================================================
// 内部方法
// =============================================================================

// createRequest 创建 HTTP 请求
func (c *Client) createRequest(ctx context.Context, method, url string, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// createRequestWithBody 创建带 body 的 HTTP 请求
func (c *Client) createRequestWithBody(ctx context.Context, method, url string, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// =============================================================================
// 工具方法
// =============================================================================

// GetProjectIDFromString 从字符串解析项目 ID
func GetProjectIDFromString(s string) (int, error) {
	return strconv.Atoi(s)
}

// GetMRIIDFromString 从字符串解析 MR IID
func GetMRIIDFromString(s string) (int, error) {
	return strconv.Atoi(s)
}
