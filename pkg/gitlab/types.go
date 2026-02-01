// Package gitlab provides GitLab integration capabilities for analyzer-ts.
// It includes GitLab API client, diff parser, MR service, and command interface.
package gitlab

import (
	"context"

	impactAnalysis "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/impact_analysis"
)

// =============================================================================
// Types - GitLab API 数据结构
// =============================================================================

// MergeRequest represents a GitLab merge request
type MergeRequest struct {
	IID          int    `json:"iid"`
	ID           int    `json:"id"`
	ProjectID    int    `json:"project_id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	WebURL       string `json:"web_url"`
}

// DiffFile represents a file diff from GitLab API
type DiffFile struct {
	OldPath     string `json:"old_path"`
	NewPath     string `json:"new_path"`
	Diff        string `json:"diff"`
	NewFile     bool   `json:"new_file"`
	RenamedFile bool   `json:"renamed_file"`
	DeletedFile bool   `json:"deleted_file"`
}

// Comment represents a merge request comment
type Comment struct {
	ID           string `json:"id"`
	NoteID       int    `json:"note_id"`
	DiscussionID string `json:"discussion_id"`
	Body         string `json:"body"`
	Author       Author `json:"author"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// Author represents a GitLab user
type Author struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

// =============================================================================
// Diff 数据结构
// =============================================================================

// ChangedLineSetOfFiles 跟踪每个文件变更的行号
// 与 merge-request-impact-reviewer/git-diff-plugin.ts 保持一致
type ChangedLineSetOfFiles map[string]map[int]bool

// ChangeInput 兼容现有 impact-analysis 的文件级别输入
type ChangeInput struct {
	ModifiedFiles []string `json:"modifiedFiles"`
	AddedFiles    []string `json:"addedFiles"`
	DeletedFiles  []string `json:"deletedFiles"`
}

// GetFileCount 返回变更文件总数
func (c *ChangeInput) GetFileCount() int {
	return len(c.ModifiedFiles) + len(c.AddedFiles) + len(c.DeletedFiles)
}

// GetAllFiles 返回所有变更文件的列表
func (c *ChangeInput) GetAllFiles() []string {
	files := make([]string, 0, c.GetFileCount())
	files = append(files, c.ModifiedFiles...)
	files = append(files, c.AddedFiles...)
	files = append(files, c.DeletedFiles...)
	return files
}

// =============================================================================
// GitLab 配置
// =============================================================================

// GitLabConfig GitLab 连接配置
type GitLabConfig struct {
	// GitLab 实例配置
	URL    string
	Token  string

	// MR 信息
	ProjectID int
	MRIID     int

	// Diff 来源
	DiffSource string // "diff", "api", "file", "auto"
	DiffFile   string // 当 DiffSource="file" 时的文件路径
	DiffSHA    string // 可选：指定 diff 的 SHA 范围

	// 分析参数
	ManifestPath string
	DepsFile     string
	MaxDepth     int
}

// DiffSourceMode diff 来源模式
type DiffSourceMode string

const (
	DiffSourceAuto  DiffSourceMode = "auto"  // 自动检测
	DiffSourceFile  DiffSourceMode = "file"  // 从文件读取
	DiffSourceAPI   DiffSourceMode = "api"   // 从 GitLab API 获取
	DiffSourceDiff  DiffSourceMode = "diff"  // 执行 git diff 命令
)

// =============================================================================
// Impact Analysis 结果类型 (复用)
// =============================================================================

// 使用 impact-analysis 插件的结果类型
type (
	// ImpactAnalysisResult 影响分析结果
	ImpactAnalysisResult = impactAnalysis.ImpactAnalysisResult
	// ImpactMeta 分析元数据
	ImpactMeta = impactAnalysis.ImpactMeta
	// ComponentChange 组件变更
	ComponentChange = impactAnalysis.ComponentChange
	// ImpactComponent 受影响组件
	ImpactComponent = impactAnalysis.ImpactComponent
	// ChangePath 变更路径
	ChangePath = impactAnalysis.ChangePath
	// Recommendation 建议
	Recommendation = impactAnalysis.Recommendation
)

// =============================================================================
// Context 用于传递分析上下文
// =============================================================================

// AnalysisContext 分析上下文
type AnalysisContext struct {
	Context context.Context
	Config   *GitLabConfig
	// 可选：取消信号
	CancelChan <-chan struct{}
}
