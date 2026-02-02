// Package gitlab provides GitLab integration capabilities for analyzer-ts.
package gitlab

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// =============================================================================
// 常量定义
// =============================================================================

const (
	// BinaryFileMarker 二进制文件标记
	// 语义: 行 0 表示"文件级别"变更，而非具体行号
	// 用途: 二进制文件变更时，无法确定具体行号，使用此标记表示整个文件变更
	BinaryFileMarker = 0
)

// =============================================================================
// Parser - Git Diff 解析器
// =============================================================================

// Parser Git diff 解析器
// 参考 merge-request-impact-reviewer/git-diff-plugin.ts 实现
type Parser struct {
	baseDir string
}

// NewParser 创建 Parser
func NewParser(baseDir string) *Parser {
	if baseDir == "" {
		// 如果未指定 baseDir，使用当前目录
		baseDir = "."
	}
	return &Parser{
		baseDir: baseDir,
	}
}

// =============================================================================
// 公共解析方法
// =============================================================================

// ParseDiffString 解析 diff 字符串
// 返回 ChangedLineSetOfFiles: map[filePath]map[lineNumber]bool
func (p *Parser) ParseDiffString(diffOutput string) (ChangedLineSetOfFiles, error) {
	result := make(ChangedLineSetOfFiles)

	// 1. 按文件块分割: split(/^diff --git\s+/m)
	fileBlocks := p.splitDiffBlocks(diffOutput)

	// 2. 解析每个文件块
	for _, block := range fileBlocks {
		filePath, addedLines, err := p.parseFileBlock(block)
		if err != nil {
			// 记录错误但继续处理其他文件
			fmt.Printf("Warning: failed to parse file block: %v\n", err)
			continue
		}

		if len(addedLines) > 0 {
			result[filePath] = addedLines
		}
	}

	return result, nil
}

// ParseProvider 从 Provider 解析 diff
func (p *Parser) ParseProvider(ctx context.Context, provider Provider) (ChangedLineSetOfFiles, error) {
	diffString, err := provider.GetDiffString(ctx)
	if err != nil {
		return nil, fmt.Errorf("get diff string failed: %w", err)
	}

	return p.ParseDiffString(diffString)
}

// ParseFromGit 执行 git diff 命令并解析输出
func (p *Parser) ParseFromGit(baseSHA, headSHA string) (ChangedLineSetOfFiles, error) {
	// 执行 git diff 命令
	cmd := exec.Command("git", "diff", baseSHA, headSHA)
	cmd.Dir = p.baseDir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}

	return p.ParseDiffString(string(output))
}

// ParseDiffFile 从文件读取并解析 diff
func (p *Parser) ParseDiffFile(diffFilePath string) (ChangedLineSetOfFiles, error) {
	content, err := os.ReadFile(diffFilePath)
	if err != nil {
		return nil, fmt.Errorf("read diff file failed: %w", err)
	}

	return p.ParseDiffString(string(content))
}

// ParseDiffFiles 解析 GitLab API diff 格式
func (p *Parser) ParseDiffFiles(diffFiles []DiffFile) (ChangedLineSetOfFiles, error) {
	result := make(ChangedLineSetOfFiles)

	for _, diffFile := range diffFiles {
		addedLines, err := p.parseDiffText(diffFile.Diff, diffFile.OldPath, diffFile.NewPath)
		if err != nil {
			continue
		}

		if len(addedLines) > 0 {
			// 使用 NewPath，如果是新文件则 OldPath 可能为空
			filePath := diffFile.NewPath
			if filePath == "" {
				filePath = diffFile.OldPath
			}
			result[filePath] = addedLines
		}
	}

	return result, nil
}

// =============================================================================
// 内部解析方法
// =============================================================================

// splitDiffBlocks 分割 diff 为文件块
// 按/^diff --git\s+/m 分割
func (p *Parser) splitDiffBlocks(diffOutput string) []string {
	// 使用正则分割 diff --git 行
	pattern := regexp.MustCompile(`(?m)^diff --git\s+`)
	matches := pattern.FindAllStringIndex(diffOutput, -1)

	blocks := make([]string, 0, len(matches))

	for i, match := range matches {
		// 当前块的开始位置
		start := match[0]

		// 当前块的结束位置（下一个 diff --git 之前，或文件末尾）
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
// 返回: 文件路径, 新增行集合, 错误
func (p *Parser) parseFileBlock(block string) (string, map[int]bool, error) {
	// 1. 检测二进制文件变更
	// 格式: "Binary files a/path/to/file and b/path/to/file differ"
	binaryPattern := regexp.MustCompile(`(?m)^Binary files\s+(\S+)\s+and\s+(\S+)\s+differ`)
	if match := binaryPattern.FindStringSubmatch(block); match != nil && len(match) > 2 {
		filePath := match[2] // b/path/to/file
		// 移除 b/ 前缀，与文本文件路径保持一致
		if strings.HasPrefix(filePath, "b/") {
			filePath = filePath[2:]
		}
		// 使用特殊标记表示二进制文件整个文件变更
		return filePath, map[int]bool{BinaryFileMarker: true}, nil
	}

	// 2. 解析文本文件路径
	// 格式: diff --git a/path/to/file b/path/to/file
	// 使用 (?m) 多行模式，使 ^ 匹配每行开头
	pathPattern := regexp.MustCompile(`(?m)^diff --git\s+a/(\S+)\s+b/(\S+)`)
	pathMatch := pathPattern.FindStringSubmatch(block)

	if pathMatch == nil || len(pathMatch) < 3 {
		return "", nil, fmt.Errorf("cannot parse file path from block")
	}

	// 使用新文件路径
	filePath := pathMatch[2]

	// 3. 提取 hunk 块并解析新增行
	addedLines := p.extractAddedLines(block)

	return filePath, addedLines, nil
}

// parseDiffText 解析 diff 文本，提取新增行
func (p *Parser) parseDiffText(diffText, oldPath, newPath string) (map[int]bool, error) {
	scanner := bufio.NewScanner(bytes.NewBufferString(diffText))
	addedLines := make(map[int]bool)

	lineNumber := 0
	currentHunk := false

	for scanner.Scan() {
		line := scanner.Text()

		// 检测 hunk 头: @@ -old,old +new +new @@
		hunkPattern := regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)
		if hunkMatch := hunkPattern.FindStringSubmatch(line); hunkMatch != nil {
			// 解析新文件的起始行号
			fmt.Sscanf(hunkMatch[3], "%d", &lineNumber)
			currentHunk = true
			continue
		}

		// 只处理新增的行（以 + 开头）
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
// 参考 merge-request-impact-reviewer 的逻辑
func (p *Parser) extractAddedLines(block string) map[int]bool {
	addedLines := make(map[int]bool)

	scanner := bufio.NewScanner(bytes.NewBufferString(block))
	lineNumber := 0
	currentHunk := false

	// 正则匹配 hunk 头: @@ -old,old +new +new @@
	hunkPattern := regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+)?)[^@]*@@`)

	for scanner.Scan() {
		line := scanner.Text()

		// 检测 hunk 头
		if hunkMatch := hunkPattern.FindStringSubmatch(line); hunkMatch != nil {
			// 解析新文件的起始行号
			fmt.Sscanf(hunkMatch[3], "%d", &lineNumber)
			currentHunk = true
			continue
		}

		// 在 hunk 块中处理行
		if currentHunk {
			// 新增行（以 + 开头，但不是 +++）
			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "++") {
				// 只记录非空的新增行
				if len(strings.TrimSpace(line)) > 0 {
					addedLines[lineNumber] = true
				}
				lineNumber++
			} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
				// 删除行：不影响新文件的行号
				continue
			} else if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
				// 上下文行：新文件也有这行，需要增加行号
				lineNumber++
			}
			// 其他行（如 ---, +++）不影响行号
		}
	}

	if err := scanner.Err(); err != nil {
		return nil
	}

	return addedLines
}

// =============================================================================
// 工具方法
// =============================================================================

// resolveToProjectRoot 将文件路径解析为相对于项目根的路径
func (p *Parser) resolveToProjectRoot(filePath string) string {
	// 如果是绝对路径，转换为相对路径
	if filepath.IsAbs(filePath) {
		// 尝试转换为相对路径
		relPath, err := filepath.Rel(p.baseDir, filePath)
		if err == nil {
			return filepath.ToSlash(relPath)
		}
	}

	// 如果是相对路径，直接返回
	return filepath.ToSlash(filePath)
}

// isModified 判断文件是否被修改（既有删除又有新增）
func (p *Parser) isModified(diffText string) bool {
	// 检查是否同时包含删除和新增行
	hasDeletion := regexp.MustCompile(`^-\s*[^\s]`).MatchString(diffText)
	hasAddition := regexp.MustCompile(`^\+\s*[^\s]`).MatchString(diffText)
	return hasDeletion && hasAddition
}

// isAdded 判断文件是否是新增文件
func (p *Parser) isAdded(diffText string) bool {
	return regexp.MustCompile(`^diff --git\s+a/`).MatchString(diffText)
}

// isDeleted 判断文件是否被删除
func (p *Parser) isDeleted(diffText string) bool {
	return regexp.MustCompile(`^diff --git\s+d/`).MatchString(diffText)
}

// =============================================================================
// 向后兼容的别名
// =============================================================================

// DiffParser 向后兼容的别名
type DiffParser = Parser

// NewDiffParser 向后兼容的别名
func NewDiffParser(baseDir string) *DiffParser {
	return NewParser(baseDir)
}

// ParseDiffOutput 向后兼容的别名
func (p *Parser) ParseDiffOutput(diffOutput string) (ChangedLineSetOfFiles, error) {
	return p.ParseDiffString(diffOutput)
}

// GetChangedFiles 向后兼容的别名（返回文件列表）
// 注意：这个方法现在只返回修改的文件路径列表，不再使用 ChangeInput
func (p *Parser) GetChangedFiles(lineSet ChangedLineSetOfFiles) []string {
	files := make([]string, 0, len(lineSet))
	for filePath := range lineSet {
		files = append(files, filePath)
	}
	return files
}
