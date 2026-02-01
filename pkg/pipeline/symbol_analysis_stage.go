package pipeline

import (
	"fmt"

	"github.com/Flying-Bird1999/analyzer-ts/pkg/symbol_analysis"
)

// =============================================================================
// 符号分析阶段
// =============================================================================

// SymbolAnalysisStage 符号分析阶段。
// 该阶段将行级变更转换为符号级变更。
type SymbolAnalysisStage struct {
	// 分析选项
	IncludeTypes    bool // 是否包含类型声明（接口、类型别名）
	IncludeInternal bool // 是否包含非导出符号
}

// NewSymbolAnalysisStage 创建符号分析阶段。
func NewSymbolAnalysisStage() *SymbolAnalysisStage {
	return &SymbolAnalysisStage{
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
	changedLines, exists := ctx.GetResult("diff_parser")
	if !exists {
		return nil, fmt.Errorf("diff parser result not found in context")
	}

	// 类型断言：检查是否为 ChangedLineSetOfFiles 类型
	lineSet, ok := changedLines.(map[string]map[int]bool)
	if !ok {
		return nil, fmt.Errorf("invalid diff parser result type")
	}

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
	fmt.Printf("  - 分析 %d 个文件的变更\n", len(lineSet))
	results := analyzer.AnalyzeChangedLines(lineSet)

	// 统计结果
	totalSymbols := 0
	exportedSymbols := 0
	for _, result := range results {
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
	changedLines, exists := ctx.GetResult("diff_parser")
	if !exists {
		return true
	}

	lineSet, ok := changedLines.(map[string]map[int]bool)
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
		for _, s := range fileResult.AffectedSymbols {
			if s.IsExported {
				exportedCount++
			}
		}
		fmt.Printf("%s: %d 个符号 (%d 已导出)", filePath, len(fileResult.AffectedSymbols), exportedCount)
	}
}
