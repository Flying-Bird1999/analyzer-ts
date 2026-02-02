// Package gitlab provides GitLab integration capabilities for analyzer-ts.
// It includes GitLab API client, diff parser, and comment service.
package gitlab

import (
	impactAnalysis "github.com/Flying-Bird1999/analyzer-ts/pkg/impact_analysis"
)

// =============================================================================
// Core Types - GitLab API 数据结构
// =============================================================================

// ChangedLineSetOfFiles 每个文件的变更行集合
// Key: 文件路径, Value: 变更行号集合
type ChangedLineSetOfFiles map[string]map[int]bool

// MergeRequest GitLab MR 信息
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

// DiffFile GitLab diff 文件
type DiffFile struct {
	OldPath     string `json:"old_path"`
	NewPath     string `json:"new_path"`
	Diff        string `json:"diff"`
	NewFile     bool   `json:"new_file"`
	RenamedFile bool   `json:"renamed_file"`
	DeletedFile bool   `json:"deleted_file"`
}

// Comment GitLab 评论
type Comment struct {
	ID        string `json:"id"`
	NoteID    int    `json:"note_id"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Author GitLab 用户信息
type Author struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

// Config GitLab 配置
type Config struct {
	URL       string
	Token     string
	ProjectID int
	MRIID     int
}

// =============================================================================
// Impact Analysis 类型别名 (复用 pkg/impact_analysis)
// =============================================================================

// AnalysisResult 影响分析结果（别名）
type AnalysisResult = impactAnalysis.AnalysisResult

// ImpactMeta 分析元数据（别名）
type ImpactMeta = impactAnalysis.ImpactMeta

// ComponentChange 组件变更（别名）
type ComponentChange = impactAnalysis.ComponentChange

// ImpactComponent 受影响组件（别名）
type ImpactComponent = impactAnalysis.ImpactComponent

// Recommendation 建议（别名）
type Recommendation = impactAnalysis.Recommendation

// RiskAssessment 风险评估（别名）
type RiskAssessment = impactAnalysis.RiskAssessment

// ImpactAnalysisResult 保持向后兼容的别名
type ImpactAnalysisResult = AnalysisResult
