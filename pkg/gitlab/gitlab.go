// Package gitlab provides GitLab-specific capabilities for analyzer-ts.
//
// This package provides:
//   - GitLabClient: API client for GitLab
//   - DiffProvider: Get diff from GitLab MR
//   - MRPoster: Post results to GitLab MR
//
// The package does NOT contain orchestration logic - that belongs in
// the application layer (cmd/) or pipeline layer (pkg/pipeline).
package gitlab

import (
	"context"
	"fmt"
	"os"
	"strconv"
)

// =============================================================================
// 便捷入口函数
// =============================================================================

// AnalyzeMR 在 GitLab CI 环境中分析 MR
// 这是一个便捷函数，用于最常见的 GitLab CI 场景
//
// 注意：这只是一个薄封装，真正的编排逻辑应该在应用层实现
func AnalyzeMR(ctx context.Context, projectRoot string) error {
	// 1. 从环境变量读取配置
	config, err := ReadConfigFromEnv()
	if err != nil {
		return fmt.Errorf("read config from env: %w", err)
	}

	// 2. 创建客户端
	client := NewClient(config.URL, config.Token)

	// 3. 获取 diff
	diffProvider := NewDiffProvider(client, config.ProjectID, config.MRIID)
	diffFiles, err := diffProvider.GetDiffFiles(ctx)
	if err != nil {
		return fmt.Errorf("get diff: %w", err)
	}

	// 注意：这里只是演示如何使用 gitlab 包的能力
	// 真正的分析编排应该由 pipeline 或应用层完成
	_ = diffFiles

	// 4. （示例）分析完成后发布结果
	// poster := NewMRPoster(client, config.ProjectID, config.MRIID)
	// poster.PostResult(ctx, result)

	return fmt.Errorf("analysis not implemented - use pipeline instead")
}

// =============================================================================
// 配置读取
// =============================================================================

// ReadConfigFromEnv 从环境变量读取 GitLab 配置
func ReadConfigFromEnv() (*GitLabConfig, error) {
	config := &GitLabConfig{
		DiffSource: string(DiffSourceAuto),
		MaxDepth:   10,
	}

	// GitLab 连接信息
	if url := os.Getenv("CI_SERVER_URL"); url != "" {
		config.URL = url
	}
	if token := os.Getenv("GITLAB_TOKEN"); token != "" {
		config.Token = token
	}

	// MR 信息
	if projectID := os.Getenv("CI_PROJECT_ID"); projectID != "" {
		id, err := strconv.Atoi(projectID)
		if err != nil {
			return nil, fmt.Errorf("invalid CI_PROJECT_ID: %w", err)
		}
		config.ProjectID = id
	}
	if mrIID := os.Getenv("CI_MERGE_REQUEST_IID"); mrIID != "" {
		id, err := strconv.Atoi(mrIID)
		if err != nil {
			return nil, fmt.Errorf("invalid CI_MERGE_REQUEST_IID: %w", err)
		}
		config.MRIID = id
	}

	// 分析参数
	if manifest := os.Getenv("ANALYZER_MANIFEST_PATH"); manifest != "" {
		config.ManifestPath = manifest
	}
	if depsFile := os.Getenv("ANALYZER_DEPS_FILE"); depsFile != "" {
		config.DepsFile = depsFile
	}

	return config, nil
}

// =============================================================================
// 旧的 Integration 类型（向后兼容，已废弃）
// =============================================================================

// GitLabIntegration GitLab 集合器
//
// Deprecated: 这个类型将被移除。请使用：
//   - gitlab.DiffProvider 获取 diff
//   - pkg/pipeline 进行分析编排
//   - gitlab.MRPoster 发布结果
type GitLabIntegration struct {
	client    *Client
	config    *GitLabConfig
}

// NewGitLabIntegration 创建 GitLab 集成器
//
// Deprecated: 请直接使用 pipeline 层进行编排
func NewGitLabIntegration(config *GitLabConfig) *GitLabIntegration {
	return &GitLabIntegration{
		client: NewClient(config.URL, config.Token),
		config: config,
	}
}

// RunAnalysis 运行分析
//
// Deprecated: 请使用 pkg/pipeline 代替
func (g *GitLabIntegration) RunAnalysis(ctx context.Context, projectRoot string) error {
	// 这个方法现在只是一个占位符，指向新的实现方式
	return fmt.Errorf("GitLabIntegration.RunAnalysis is deprecated. " +
		"Please use pkg/pipeline for orchestration instead. " +
		"See pkg/gitlab/example_usage.go for examples.")
}

// GetClient 获取 GitLab 客户端
func (g *GitLabIntegration) GetClient() *Client {
	return g.client
}

// GetConfig 获取配置
func (g *GitLabIntegration) GetConfig() *GitLabConfig {
	return g.config
}
