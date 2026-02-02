package pipeline

import (
	"context"
	"fmt"

	"github.com/Flying-Bird1999/analyzer-ts/pkg/gitlab"
)

// =============================================================================
// Diff Parser 阶段
// =============================================================================

// DiffSourceType diff 来源类型
type DiffSourceType string

const (
	DiffSourceAuto DiffSourceType = "auto" // 自动检测
	DiffSourceFile DiffSourceType = "file" // 从文件读取
	DiffSourceAPI  DiffSourceType = "api"  // 从 API 获取
	DiffSourceSHA  DiffSourceType = "diff" // 使用 git diff 命令
)

// GitLabClient GitLab API 客户端接口
// 定义接口避免循环依赖
type GitLabClient interface {
	GetMergeRequestDiff(ctx context.Context, projectID, mrIID int) ([]gitlab.DiffFile, error)
}

// DiffParserStage diff 解析阶段
// 该阶段从 API 或文件获取 git diff，解析出行级变更
type DiffParserStage struct {
	// API 客户端（用于 API 模式）
	client GitLabClient

	// Diff 来源配置
	source    DiffSourceType
	diffFile  string
	diffSHA   string
	baseDir   string
	projectID int
	mrIID     int
}

// NewDiffParserStage 创建 diff 解析阶段
func NewDiffParserStage(
	client GitLabClient,
	source DiffSourceType,
	diffFile, diffSHA, baseDir string,
	projectID, mrIID int,
) *DiffParserStage {
	return &DiffParserStage{
		client:    client,
		source:    source,
		diffFile:  diffFile,
		diffSHA:   diffSHA,
		baseDir:   baseDir,
		projectID: projectID,
		mrIID:     mrIID,
	}
}

// Name 返回阶段名称
func (s *DiffParserStage) Name() string {
	return "Diff解析"
}

// Execute 执行 diff 解析
func (s *DiffParserStage) Execute(ctx *AnalysisContext) (interface{}, error) {
	// 使用 gitlab 包的 Parser
	parser := gitlab.NewParser(s.baseDir)

	var lineSet gitlab.ChangedLineSetOfFiles
	var err error

	switch s.source {
	case DiffSourceFile:
		if s.diffFile != "" {
			fmt.Printf("  - 从文件读取 diff: %s\n", s.diffFile)
			lineSet, err = parser.ParseDiffFile(s.diffFile)
		} else {
			return nil, fmt.Errorf("diff file not specified")
		}
	case DiffSourceAPI:
		if s.client == nil {
			return nil, fmt.Errorf("GitLab client not configured")
		}
		diffFiles, err := s.client.GetMergeRequestDiff(ctx.Cancel, s.projectID, s.mrIID)
		if err != nil {
			return nil, fmt.Errorf("get MR diff failed: %w", err)
		}
		fmt.Printf("  - 从 API 获取 diff: %d 个文件\n", len(diffFiles))
		// 转换为 gitlab.DiffFile 类型
		gitlabDiffFiles := make([]gitlab.DiffFile, len(diffFiles))
		for i, df := range diffFiles {
			gitlabDiffFiles[i] = gitlab.DiffFile{
				OldPath:     df.OldPath,
				NewPath:     df.NewPath,
				Diff:        df.Diff,
				NewFile:     df.NewFile,
				RenamedFile: df.RenamedFile,
				DeletedFile: df.DeletedFile,
			}
		}
		lineSet, err = parser.ParseDiffFiles(gitlabDiffFiles)
	case DiffSourceSHA:
		if s.diffSHA != "" {
			shas := s.parseSHA(s.diffSHA)
			if len(shas) == 2 {
				fmt.Printf("  - 执行 git diff %s...%s\n", shas[0][:8], shas[1][:8])
				lineSet, err = parser.ParseFromGit(shas[0], shas[1])
			} else {
				return nil, fmt.Errorf("invalid SHA format")
			}
		} else {
			baseSHA := ctx.GetOption("baseSHA", "").(string)
			if baseSHA != "" {
				fmt.Printf("  - 执行 git diff %s...HEAD\n", baseSHA[:8])
				lineSet, err = parser.ParseFromGit(baseSHA, "HEAD")
			} else {
				return nil, fmt.Errorf("no diff source available")
			}
		}
	case DiffSourceAuto:
		// 自动检测
		return s.autoDetect(ctx)
	default:
		return nil, fmt.Errorf("unknown diff source: %s", s.source)
	}

	if err != nil {
		return nil, fmt.Errorf("parse diff failed: %w", err)
	}

	// 统计结果
	totalFiles := len(lineSet)
	totalLines := 0
	for _, lines := range lineSet {
		totalLines += len(lines)
	}
	fmt.Printf("  - 发现 %d 个文件，%d 行变更\n", totalFiles, totalLines)

	return lineSet, nil
}

// Skip 判断是否跳过此阶段
func (s *DiffParserStage) Skip(ctx *AnalysisContext) bool {
	return false
}

// =============================================================================
// 内部辅助方法
// =============================================================================

// parseSHA 解析 SHA 字符串（支持 "base...head" 格式）
func (s *DiffParserStage) parseSHA(sha string) []string {
	if sha == "" {
		return nil
	}
	// 简单的按 "..." 分割
	for i := 0; i < len(sha)-3; i++ {
		if sha[i:i+3] == "..." {
			return []string{sha[:i], sha[i+3:]}
		}
	}
	return nil
}

// autoDetect 自动检测 diff 来源
func (s *DiffParserStage) autoDetect(ctx *AnalysisContext) (gitlab.ChangedLineSetOfFiles, error) {
	parser := gitlab.NewParser(s.baseDir)

	// 优先级 1: 从 API 获取
	if s.client != nil && s.projectID > 0 && s.mrIID > 0 {
		diffFiles, err := s.client.GetMergeRequestDiff(ctx.Cancel, s.projectID, s.mrIID)
		if err == nil && len(diffFiles) > 0 {
			fmt.Println("  ℹ️  使用 API")
			gitlabDiffFiles := make([]gitlab.DiffFile, len(diffFiles))
			for i, df := range diffFiles {
				gitlabDiffFiles[i] = gitlab.DiffFile{
					OldPath:     df.OldPath,
					NewPath:     df.NewPath,
					Diff:        df.Diff,
					NewFile:     df.NewFile,
					RenamedFile: df.RenamedFile,
					DeletedFile: df.DeletedFile,
				}
			}
			return parser.ParseDiffFiles(gitlabDiffFiles)
		}
	}

	// 优先级 2: 尝试 git diff
	baseSHA := ctx.GetOption("baseSHA", "").(string)
	if baseSHA != "" {
		fmt.Println("  ℹ️  使用 git diff")
		return parser.ParseFromGit(baseSHA, "HEAD")
	}

	// 优先级 3: 从文件读取
	diffFile := ctx.GetOption("diffFile", "").(string)
	if diffFile != "" {
		fmt.Println("  ℹ️  使用 diff 文件")
		return parser.ParseDiffFile(diffFile)
	}

	return nil, fmt.Errorf("no diff source available")
}
