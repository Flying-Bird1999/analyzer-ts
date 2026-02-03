package pipeline

import (
	"fmt"

	"github.com/Flying-Bird1999/analyzer-ts/pkg/gitlab"
	"github.com/Flying-Bird1999/analyzer-ts/pkg/symbol_analysis"
)

// =============================================================================
// 符号分析阶段
// =============================================================================

// SymbolAnalysisStage 符号分析阶段。
// 该阶段将行级变更转换为符号级变更。
type SymbolAnalysisStage struct {
	// Git 仓库根目录（用于路径转换，当 GitRoot != ProjectRoot 时）
	gitRoot string

	// 分析选项
	IncludeTypes    bool // 是否包含类型声明（接口、类型别名）
	IncludeInternal bool // 是否包含非导出符号
}

// NewSymbolAnalysisStage 创建符号分析阶段。
// gitRoot: Git 仓库根目录，当 GitRoot != ProjectRoot 时用于路径转换
func NewSymbolAnalysisStage(gitRoot string) *SymbolAnalysisStage {
	return &SymbolAnalysisStage{
		gitRoot:         gitRoot,
		IncludeTypes:    true,
		IncludeInternal: false, // 默认只分析导出符号
	}
}

// Name 返回阶段名称。
func (s *SymbolAnalysisStage) Name() string {
	return "符号分析"
}

// Execute 执行符号分析。
func (s *SymbolAnalysisStage) Execute(ctx *AnalysisContext) (interface{}, error) {
	// 从上下文中获取 ChangedLineSetOfFiles
	// 注意：使用 DiffParserStage 的 Name() 作为 key
	changedLines, exists := ctx.GetResult("Diff解析")
	if !exists {
		return nil, fmt.Errorf("diff parser result not found in context")
	}

	// 类型断言：检查是否为 gitlab.ChangedLineSetOfFiles 类型
	lineSet, ok := changedLines.(gitlab.ChangedLineSetOfFiles)
	if !ok {
		return nil, fmt.Errorf("invalid diff parser result type")
	}

	// 路径转换：当 GitRoot != ProjectRoot 时，需要将相对路径转换为绝对路径
	// 如果 GitRoot == ProjectRoot（默认情况），路径已经是正确的
	lineSet = s.normalizePaths(lineSet)

	// 从上下文中获取 tsmorphgo 项目
	if ctx.Project == nil {
		return nil, fmt.Errorf("project not initialized in context")
	}

	// 创建符号分析器
	opts := symbol_analysis.AnalysisOptions{
		IncludeTypes:    s.IncludeTypes,
		IncludeInternal: s.IncludeInternal,
	}
	analyzer := symbol_analysis.NewAnalyzer(ctx.Project, opts)

	// 执行分析
	results := analyzer.AnalyzeChangedLines(lineSet)

	// 统计结果
	totalSymbols := 0
	exportedSymbols := 0
	for filePath, result := range results {
		fmt.Printf("    - %s: IsSymbolFile=%v, 符号数=%d\n", filePath, result.IsSymbolFile, len(result.AffectedSymbols))
		for _, symbol := range result.AffectedSymbols {
			fmt.Printf("      * %s (%s) 导出=%v\n", symbol.Name, symbol.Kind, symbol.IsExported)
		}
		totalSymbols += len(result.AffectedSymbols)
		for _, symbol := range result.AffectedSymbols {
			if symbol.IsExported {
				exportedSymbols++
			}
		}
	}

	fmt.Printf("  - 发现 %d 个符号，其中 %d 个已导出\n", totalSymbols, exportedSymbols)

	return results, nil
}

// Skip 判断是否跳过此阶段。
func (s *SymbolAnalysisStage) Skip(ctx *AnalysisContext) bool {
	// 如果没有变更文件，跳过此阶段
	// 注意：使用 DiffParserStage 的 Name() 作为 key
	changedLines, exists := ctx.GetResult("Diff解析")
	if !exists {
		return true
	}

	lineSet, ok := changedLines.(gitlab.ChangedLineSetOfFiles)
	if !ok || len(lineSet) == 0 {
		return true
	}

	return false
}

// PrintResult 打印阶段结果的简要信息。
func (s *SymbolAnalysisStage) PrintResult(result interface{}) {
	results, ok := result.(map[string]*symbol_analysis.FileAnalysisResult)
	if !ok {
		fmt.Printf("符号分析结果格式错误\n")
		return
	}

	for filePath, fileResult := range results {
		exportedCount := 0
		for _, sym := range fileResult.AffectedSymbols {
			if sym.IsExported {
				exportedCount++
			}
		}
		fmt.Printf("%s: %d 个符号 (%d 已导出)", filePath, len(fileResult.AffectedSymbols), exportedCount)
	}
}

// normalizePaths 标准化路径。
// 当 GitRoot != ProjectRoot 时，将相对路径转换为绝对路径。
// 默认情况下（GitRoot == ProjectRoot），路径不需要转换。
func (s *SymbolAnalysisStage) normalizePaths(lineSet gitlab.ChangedLineSetOfFiles) gitlab.ChangedLineSetOfFiles {
	// 检查是否需要转换：如果路径已经是绝对路径，则不需要转换
	needsConversion := false
	for filePath := range lineSet {
		// 如果第一个文件路径不是绝对路径，则认为需要转换
		// 在 Unix 系统上，绝对路径以 / 开头
		if len(filePath) > 0 && filePath[0] != '/' {
			needsConversion = true
		}
		break // 只检查第一个文件即可
	}

	if !needsConversion {
		return lineSet // 路径已经是正确的格式
	}

	// 需要转换：将相对路径转换为绝对路径
	// 相对路径是相对于 GitRoot 的
	normalized := make(gitlab.ChangedLineSetOfFiles)
	for filePath, lines := range lineSet {
		// 如果是相对路径，拼接 GitRoot
		if len(filePath) > 0 && filePath[0] != '/' {
			// 检查路径是否已经包含 gitRoot 前缀（避免重复拼接）
			if s.gitRoot != "" && len(filePath) > len(s.gitRoot) && filePath[:len(s.gitRoot)] == s.gitRoot {
				// 路径已经包含 gitRoot 前缀，直接使用
				normalized[filePath] = lines
			} else {
				normalized[s.gitRoot+"/"+filePath] = lines
			}
		} else {
			normalized[filePath] = lines
		}
	}
	return normalized
}
