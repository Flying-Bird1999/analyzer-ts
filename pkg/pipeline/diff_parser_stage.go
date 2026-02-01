package pipeline

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
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
	GetMergeRequestDiff(ctx context.Context, projectID, mrIID int) ([]DiffFile, error)
}

// DiffFile GitLab diff 文件格式
type DiffFile struct {
	Diff    string // diff 内容
	OldPath string // 旧文件路径
	NewPath string // 新文件路径
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
	var lineSet map[string]map[int]bool
	var err error

	switch s.source {
	case DiffSourceFile:
		if s.diffFile != "" {
			fmt.Printf("  - 从文件读取 diff: %s\n", s.diffFile)
			lineSet, err = s.parseDiffFile(s.diffFile)
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
		lineSet, err = s.parseDiffFiles(diffFiles)
	case DiffSourceSHA:
		if s.diffSHA != "" {
			shas := s.parseSHA(s.diffSHA)
			if len(shas) == 2 {
				fmt.Printf("  - 执行 git diff %s...%s\n", shas[0][:8], shas[1][:8])
				lineSet, err = s.parseFromGit(shas[0], shas[1])
			} else {
				return nil, fmt.Errorf("invalid SHA format")
			}
		} else {
			baseSHA := ctx.GetOption("baseSHA", "").(string)
			if baseSHA != "" {
				fmt.Printf("  - 执行 git diff %s...HEAD\n", baseSHA[:8])
				lineSet, err = s.parseFromGit(baseSHA, "HEAD")
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
// 内部解析方法
// =============================================================================

// parseSHA 解析 SHA 字符串（支持 "base...head" 格式）
func (s *DiffParserStage) parseSHA(sha string) []string {
	return strings.Split(sha, "...")
}

// parseDiffFile 从文件读取并解析 diff
func (s *DiffParserStage) parseDiffFile(diffFilePath string) (map[string]map[int]bool, error) {
	content, err := os.ReadFile(diffFilePath)
	if err != nil {
		return nil, fmt.Errorf("read diff file failed: %w", err)
	}

	return s.parseDiffOutput(string(content))
}

// parseFromGit 执行 git diff 命令并解析输出
func (s *DiffParserStage) parseFromGit(baseSHA, headSHA string) (map[string]map[int]bool, error) {
	cmd := exec.Command("git", "diff", baseSHA, headSHA)
	cmd.Dir = s.baseDir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}

	return s.parseDiffOutput(string(output))
}

// parseDiffFiles 解析 API diff 格式
func (s *DiffParserStage) parseDiffFiles(diffFiles []DiffFile) (map[string]map[int]bool, error) {
	result := make(map[string]map[int]bool)

	for _, diffFile := range diffFiles {
		addedLines, err := s.parseDiffText(diffFile.Diff, diffFile.OldPath, diffFile.NewPath)
		if err != nil {
			continue
		}

		if len(addedLines) > 0 {
			filePath := diffFile.NewPath
			if filePath == "" {
				filePath = diffFile.OldPath
			}
			result[filePath] = addedLines
		}
	}

	return result, nil
}

// parseDiffOutput 解析 git diff 输出（字符串格式）
func (s *DiffParserStage) parseDiffOutput(diffOutput string) (map[string]map[int]bool, error) {
	result := make(map[string]map[int]bool)

	// 按文件块分割
	fileBlocks := s.splitDiffBlocks(diffOutput)

	// 解析每个文件块
	for _, block := range fileBlocks {
		filePath, addedLines, err := s.parseFileBlock(block)
		if err != nil {
			fmt.Printf("Warning: failed to parse file block: %v\n", err)
			continue
		}

		if len(addedLines) > 0 {
			result[filePath] = addedLines
		}
	}

	return result, nil
}

// splitDiffBlocks 分割 diff 为文件块
func (s *DiffParserStage) splitDiffBlocks(diffOutput string) []string {
	pattern := regexp.MustCompile(`(?m)^diff --git\s+`)
	matches := pattern.FindAllStringIndex(diffOutput, -1)

	blocks := make([]string, 0, len(matches))

	for i, match := range matches {
		start := match[0]
		end := len(diffOutput)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}

		block := strings.TrimSpace(diffOutput[start:end])
		if block != "" {
			blocks = append(blocks, block)
		}
	}

	return blocks
}

// parseFileBlock 解析单个文件的 diff 块
func (s *DiffParserStage) parseFileBlock(block string) (string, map[int]bool, error) {
	// 检测二进制文件
	binaryPattern := regexp.MustCompile(`(?m)^Binary files\s+(\S+)\s+and\s+(\S+)\s+differ`)
	if match := binaryPattern.FindStringSubmatch(block); match != nil && len(match) > 2 {
		filePath := match[2]
		if strings.HasPrefix(filePath, "b/") {
			filePath = filePath[2:]
		}
		return filePath, map[int]bool{0: true}, nil // 行 0 表示整个文件变更
	}

	// 解析文本文件路径
	pathPattern := regexp.MustCompile(`(?m)^diff --git\s+a/(\S+)\s+b/(\S+)`)
	pathMatch := pathPattern.FindStringSubmatch(block)

	if pathMatch == nil || len(pathMatch) < 3 {
		return "", nil, fmt.Errorf("cannot parse file path from block")
	}

	filePath := pathMatch[2]
	addedLines := s.extractAddedLines(block)

	return filePath, addedLines, nil
}

// parseDiffText 解析 diff 文本，提取新增行
func (s *DiffParserStage) parseDiffText(diffText, oldPath, newPath string) (map[int]bool, error) {
	scanner := bufio.NewScanner(bytes.NewBufferString(diffText))
	addedLines := make(map[int]bool)

	lineNumber := 0
	currentHunk := false

	for scanner.Scan() {
		line := scanner.Text()

		// 检测 hunk 头: @@ -old,old +new +new @@
		hunkPattern := regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)
		if hunkMatch := hunkPattern.FindStringSubmatch(line); hunkMatch != nil {
			fmt.Sscanf(hunkMatch[3], "%d", &lineNumber)
			currentHunk = true
			continue
		}

		// 处理新增的行（以 + 开头）
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "++") && len(strings.TrimSpace(line)) > 0 {
			if currentHunk && lineNumber > 0 {
				addedLines[lineNumber] = true
			}
			lineNumber++
		}
	}

	return addedLines, scanner.Err()
}

// extractAddedLines 从文件块中提取新增的行
func (s *DiffParserStage) extractAddedLines(block string) map[int]bool {
	addedLines := make(map[int]bool)

	scanner := bufio.NewScanner(bytes.NewBufferString(block))
	lineNumber := 0
	currentHunk := false

	hunkPattern := regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+)?)[^@]*@@`)

	for scanner.Scan() {
		line := scanner.Text()

		// 检测 hunk 头
		if hunkMatch := hunkPattern.FindStringSubmatch(line); hunkMatch != nil {
			fmt.Sscanf(hunkMatch[3], "%d", &lineNumber)
			currentHunk = true
			continue
		}

		// 在 hunk 块中处理行
		if currentHunk {
			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "++") {
				if len(strings.TrimSpace(line)) > 0 {
					addedLines[lineNumber] = true
				}
				lineNumber++
			} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
				continue
			} else if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
				lineNumber++
			}
		}
	}

	return addedLines
}

// autoDetect 自动检测 diff 来源
func (s *DiffParserStage) autoDetect(ctx *AnalysisContext) (map[string]map[int]bool, error) {
	// 优先级 1: 从 API 获取
	if s.client != nil && s.projectID > 0 && s.mrIID > 0 {
		diffFiles, err := s.client.GetMergeRequestDiff(ctx.Cancel, s.projectID, s.mrIID)
		if err == nil && len(diffFiles) > 0 {
			fmt.Println("  ℹ️  使用 API")
			return s.parseDiffFiles(diffFiles)
		}
	}

	// 优先级 2: 尝试 git diff
	baseSHA := ctx.GetOption("baseSHA", "").(string)
	if baseSHA != "" {
		fmt.Println("  ℹ️  使用 git diff")
		return s.parseFromGit(baseSHA, "HEAD")
	}

	// 优先级 3: 从文件读取
	diffFile := ctx.GetOption("diffFile", "").(string)
	if diffFile != "" {
		fmt.Println("  ℹ️  使用 diff 文件")
		return s.parseDiffFile(diffFile)
	}

	return nil, fmt.Errorf("no diff source available")
}
