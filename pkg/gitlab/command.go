// Package gitlab provides GitLab integration capabilities for analyzer-ts.
package gitlab

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

// =============================================================================
// Command - GitLab 命令
// =============================================================================

// GetCommand 返回 gitlab 子命令
func GetCommand() *cobra.Command {
	// 创建 gitlab 命令
	cmd := &cobra.Command{
		Use:   "gitlab",
		Short: "GitLab 集成命令",
		Long: `GitLab 集成命令，提供代码影响分析和 MR 评论功能

支持在 GitLab CI/CD 流程中自动分析代码变更的影响范围，
并在 Merge Request 中发布分析结果。

示例:
  # GitLab CI 模式（自动检测环境变量）
  analyzer-ts gitlab impact -i .

  # 本地测试模式
  analyzer-ts gitlab impact -i /path/to/project \
    --gitlab-url https://gitlab.example.com \
    --gitlab-token $GITLAB_TOKEN \
    --project-id 123 --mr-id 456 \
    --diff-file /path/to/diff.patch`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGitLabCommand(cmd, args)
		},
	}

	// 注册子命令
	cmd.AddCommand(getImpactCommand())

	return cmd
}

// =============================================================================
// impact 子命令
// =============================================================================

// getImpactCommand 返回 impact 子命令
func getImpactCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "impact",
		Short: "分析代码影响并发布 MR 评论",
		Long:  `分析代码变更的影响范围并在 GitLab MR 中发布评论

工作流程：
1. 解析 git diff（文件/API/自动检测）
2. 运行 component-deps-v2 生成依赖图
3. 运行 impact-analysis 计算影响范围
4. 格式化结果为 Markdown 并发布到 MR`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImpactCommand(cmd, args)
		},
	}

	// GitLab 连接参数
	cmd.Flags().String("gitlab-url", "", "GitLab 实例 URL (默认: $CI_SERVER_URL)")
	cmd.Flags().String("gitlab-token", "", "GitLab Token (默认: $GITLAB_TOKEN)")
	cmd.Flags().Int("project-id", 0, "项目 ID (默认: $CI_PROJECT_ID)")
	cmd.Flags().Int("mr-id", 0, "MR IID (默认: $CI_MERGE_REQUEST_ID)")

	// Diff 来源参数
	cmd.Flags().String("diff-source", "auto", "Diff 来源: diff/api/file/auto (默认: auto-detect)")
	cmd.Flags().String("diff-file", "", "本地 diff 文件路径 (diff-source=file)")
	cmd.Flags().String("diff-sha", "", "指定 diff 的 SHA 范围 (格式: base...head)")

	// 分析参数
	cmd.Flags().String("manifest", "", "component-manifest.json 路径")
	cmd.Flags().String("deps-file", "", "依赖数据文件路径")
	cmd.Flags().Int("max-depth", 10, "最大传播深度")

	return cmd
}

// =============================================================================
// 命令执行函数
// =============================================================================

// runGitLabCommand 执行 gitlab 命令（入口）
func runGitLabCommand(cmd *cobra.Command, args []string) error {
	// 显示帮助信息
	return cmd.Help()
}

// runImpactCommand 执行 impact 命令
func runImpactCommand(cmd *cobra.Command, args []string) error {
	// 获取项目根目录
	inputPath, err := cmd.Flags().GetString("input")
	if err != nil || inputPath == "" {
		return fmt.Errorf("请指定项目根目录 (-i)")
	}

	// 检测或创建配置
	config, err := detectConfigFromFlags(cmd)
	if err != nil {
		return fmt.Errorf("failed to detect config: %w", err)
	}

	// 创建集成器
	integration := NewGitLabIntegration(config)

	// 执行分析
	ctx := cmd.Context()
	if err := integration.RunAnalysis(ctx, inputPath); err != nil {
		return err
	}

	return nil
}

// =============================================================================
// 配置检测
// =============================================================================

// detectConfigFromFlags 从命令行参数和环境变量检测配置
func detectConfigFromFlags(cmd *cobra.Command) (*GitLabConfig, error) {
	config := &GitLabConfig{
		DiffSource: string(DiffSourceAuto),
		MaxDepth:   10,
	}

	// GitLab 连接参数
	if url, err := cmd.Flags().GetString("gitlab-url"); err == nil && url != "" {
		config.URL = url
	} else if url := os.Getenv("CI_SERVER_URL"); url != "" {
		config.URL = url
	} else {
		return nil, fmt.Errorf("gitlab-url is required (provide --gitlab-url or set $CI_SERVER_URL)")
	}

	if token, err := cmd.Flags().GetString("gitlab-token"); err == nil && token != "" {
		config.Token = token
	} else if token := os.Getenv("GITLAB_TOKEN"); token != "" {
		config.Token = token
	} else {
		return nil, fmt.Errorf("gitlab-token is required (provide --gitlab-token or set $GITLAB_TOKEN)")
	}

	// MR 信息
	if projectID, err := cmd.Flags().GetInt("project-id"); err == nil && projectID > 0 {
		config.ProjectID = projectID
	} else if projectID := os.Getenv("CI_PROJECT_ID"); projectID != "" {
		id, err := strconv.Atoi(projectID)
		if err != nil {
			return nil, fmt.Errorf("invalid CI_PROJECT_ID: %w", err)
		}
		config.ProjectID = id
	} else {
		return nil, fmt.Errorf("project-id is required (provide --project-id or set $CI_PROJECT_ID)")
	}

	if mrIID, err := cmd.Flags().GetInt("mr-id"); err == nil && mrIID > 0 {
		config.MRIID = mrIID
	} else if mrIID := os.Getenv("CI_MERGE_REQUEST_IID"); mrIID != "" {
		id, err := strconv.Atoi(mrIID)
		if err != nil {
			return nil, fmt.Errorf("invalid CI_MERGE_REQUEST_IID: %w", err)
		}
		config.MRIID = id
	} else {
		return nil, fmt.Errorf("mr-id is required (provide --mr-id or set $CI_MERGE_REQUEST_IID)")
	}

	// Diff 来源
	if diffSource, err := cmd.Flags().GetString("diff-source"); err == nil {
		config.DiffSource = diffSource
	}

	if diffFile, err := cmd.Flags().GetString("diff-file"); err == nil && diffFile != "" {
		config.DiffFile = diffFile
		config.DiffSource = "file"
	}

	if diffSHA, err := cmd.Flags().GetString("diff-sha"); err == nil && diffSHA != "" {
		config.DiffSHA = diffSHA
		config.DiffSource = "diff"
	}

	// 分析参数
	if manifest, err := cmd.Flags().GetString("manifest"); err == nil && manifest != "" {
		config.ManifestPath = manifest
	}

	if depsFile, err := cmd.Flags().GetString("deps-file"); err == nil && depsFile != "" {
		config.DepsFile = depsFile
	}

	if maxDepth, err := cmd.Flags().GetInt("max-depth"); err == nil {
		config.MaxDepth = maxDepth
	}

	return config, nil
}

// =============================================================================
// 工具函数
// =============================================================================

// validateConfig 验证配置完整性
func validateConfig(config *GitLabConfig) error {
	if config.URL == "" {
		return fmt.Errorf("gitlab-url is required")
	}
	if config.Token == "" {
		return fmt.Errorf("gitlab-token is required")
	}
	if config.ProjectID == 0 {
		return fmt.Errorf("project-id is required")
	}
	if config.MRIID == 0 {
		return fmt.Errorf("mr-id is required")
	}
	return nil
}

// printConfig 打印配置信息（调试用）
func printConfig(config *GitLabConfig) {
	fmt.Println("📋 GitLab 配置:")
	fmt.Printf("  URL: %s\n", config.URL)
	fmt.Printf("  Project ID: %d\n", config.ProjectID)
	fmt.Printf("  MR IID: %d\n", config.MRIID)
	fmt.Printf("  Diff Source: %s\n", config.DiffSource)
	if config.DiffFile != "" {
		fmt.Printf("  Diff File: %s\n", config.DiffFile)
	}
	fmt.Printf("  Max Depth: %d\n", config.MaxDepth)
}
