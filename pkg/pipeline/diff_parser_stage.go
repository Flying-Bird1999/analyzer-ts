package pipeline

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/Flying-Bird1999/analyzer-ts/pkg/gitlab"
)

// =============================================================================
// Diff Parser 阶段
// =============================================================================

// DiffSourceType diff 来源类型
type DiffSourceType string

const (
	DiffSourceAuto    DiffSourceType = "auto"    // 自动检测
	DiffSourceFile    DiffSourceType = "file"    // 从文件读取
	DiffSourceAPI     DiffSourceType = "api"     // 从 API 获取
	DiffSourceSHA     DiffSourceType = "sha"     // 使用 git diff 命令（SHA 或分支名）
	DiffSourceBranch  DiffSourceType = "branch"  // 使用分支名对比
	DiffSourceString  DiffSourceType = "string"  // 直接传入 diff 字符串
	DiffSourceStdin   DiffSourceType = "stdin"   // 从标准输入读取 diff
)

// GitLabClient GitLab API 客户端接口
// 定义接口避免循环依赖
type GitLabClient interface {
	GetMergeRequestDiff(ctx context.Context, projectID, mrIID int) ([]gitlab.DiffFile, error)
}

// DiffParserStage diff 解析阶段
// 该阶段从多种来源获取 git diff，解析出行级变更
type DiffParserStage struct {
	// API 客户端（用于 API 模式）
	client GitLabClient

	// Diff 来源配置
	source    DiffSourceType
	diffFile  string
	diffSHA   string
	diffString string // 直接传入的 diff 字符串
	baseDir   string
	projectID int
	mrIID     int

	// 分支名配置（用于 DiffSourceBranch）
	baseBranch   string // 基础分支（如 "main", "develop"）
	targetBranch string // 目标分支（默认为 "HEAD"）
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
				// 安全截取：如果长度 < 8，显示完整字符串
				displayBase := shas[0]
				displayHead := shas[1]
				if len(displayBase) > 8 {
					displayBase = displayBase[:8]
				}
				if len(displayHead) > 8 {
					displayHead = displayHead[:8]
				}
				fmt.Printf("  - 执行 git diff %s...%s\n", displayBase, displayHead)
				lineSet, err = parser.ParseFromGit(shas[0], shas[1])
			} else {
				return nil, fmt.Errorf("invalid SHA format, expected 'base...head'")
			}
		} else {
			baseSHA := ctx.GetOption("baseSHA", "").(string)
			if baseSHA != "" {
				displayBase := baseSHA
				if len(displayBase) > 8 {
					displayBase = displayBase[:8]
				}
				fmt.Printf("  - 执行 git diff %s...HEAD\n", displayBase)
				lineSet, err = parser.ParseFromGit(baseSHA, "HEAD")
			} else {
				return nil, fmt.Errorf("no diff source available")
			}
		}
	case DiffSourceBranch:
		// 使用分支名对比（如 "main...HEAD" 或 "develop...feature-branch"）
		base := s.baseBranch
		target := s.targetBranch
		if base == "" {
			base = ctx.GetOption("baseBranch", "main").(string)
		}
		if target == "" {
			target = ctx.GetOption("targetBranch", "HEAD").(string)
		}
		fmt.Printf("  - 执行 git diff %s...%s (分支对比)\n", base, target)
		lineSet, err = parser.ParseFromGit(base, target)
	case DiffSourceString:
		// 直接传入 diff 字符串
		diffString := s.diffString
		if diffString == "" {
			diffString = ctx.GetOption("diffString", "").(string)
		}
		if diffString == "" {
			return nil, fmt.Errorf("diff string not specified")
		}
		fmt.Println("  - 从字符串解析 diff")
		lineSet, err = parser.ParseDiffString(diffString)
	case DiffSourceStdin:
		// 从标准输入读取 diff
		fmt.Println("  - 从标准输入读取 diff...")
		lineSet, err = s.parseDiffFromStdin(parser)
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

// parseDiffFromStdin 从标准输入读取 diff 内容
func (s *DiffParserStage) parseDiffFromStdin(parser *gitlab.Parser) (gitlab.ChangedLineSetOfFiles, error) {
	// 读取所有输入直到 EOF
	var buf bytes.Buffer
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		buf.WriteString(scanner.Text())
		buf.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read from stdin failed: %w", err)
	}

	diffContent := buf.String()
	if diffContent == "" {
		return nil, fmt.Errorf("no diff content from stdin")
	}

	return parser.ParseDiffString(diffContent)
}
