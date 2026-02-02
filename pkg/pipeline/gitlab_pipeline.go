// Package pipeline 提供代码分析管道协调功能。
package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/Flying-Bird1999/analyzer-ts/pkg/impact_analysis"
	"github.com/Flying-Bird1999/analyzer-ts/pkg/impact_analysis/component_analyzer"
	"github.com/Flying-Bird1999/analyzer-ts/pkg/impact_analysis/file_analyzer"
	"github.com/Flying-Bird1999/analyzer-ts/pkg/symbol_analysis"
)

// =============================================================================
// GitLab Pipeline 工厂函数
// =============================================================================

// GitLabPipelineConfig GitLab 管道配置
type GitLabPipelineConfig struct {
	// Diff 配置
	DiffSource  DiffSourceType
	DiffFile    string
	DiffSHA     string
	ProjectRoot string
	ProjectID   int
	MRIID       int

	// 组件依赖配置
	ManifestPath string // 组件清单路径（可选）
	DepsFile     string

	// 影响分析配置
	MaxDepth int

	// GitLab 客户端
	Client GitLabClient
}

// NewGitLabPipeline 创建完整的 GitLab 分析管道
// 自动检测是否为组件库项目，并执行相应的影响分析
func NewGitLabPipeline(config *GitLabPipelineConfig) *AnalysisPipeline {
	pipe := NewPipeline("GitLab Analysis Pipeline")

	// 阶段 1: Diff 解析
	pipe.AddStage(NewDiffParserStage(
		config.Client,
		config.DiffSource,
		config.DiffFile,
		config.DiffSHA,
		config.ProjectRoot,
		config.ProjectID,
		config.MRIID,
	))

	// 阶段 2: 符号分析
	pipe.AddStage(NewSymbolAnalysisStage())

	// 阶段 3: 项目解析
	pipe.AddStage(NewProjectParserStage())

	// 阶段 4: 影响分析（自动检测组件库）
	pipe.AddStage(NewImpactAnalysisStage(
		config.ManifestPath,
		config.MaxDepth,
	))

	return pipe
}

// =============================================================================
// 项目解析阶段
// =============================================================================

// ProjectParserStage 项目解析阶段
// 解析项目的所有 JS/TS 文件，构建 AST
type ProjectParserStage struct{}

// NewProjectParserStage 创建项目解析阶段
func NewProjectParserStage() *ProjectParserStage {
	return &ProjectParserStage{}
}

// Name 返回阶段名称
func (s *ProjectParserStage) Name() string {
	return "项目解析"
}

// Execute 执行项目解析
func (s *ProjectParserStage) Execute(ctx *AnalysisContext) (interface{}, error) {
	fmt.Println("  - 解析项目 AST...")

	// 创建项目解析器配置
	parserConfig := projectParser.NewProjectParserConfig(
		ctx.ProjectRoot,
		[]string{}, // exclude patterns
		false,      // isMonorepo
		nil,        // target extensions (use default)
	)

	// 解析项目
	parsingResult := projectParser.NewProjectParserResult(parserConfig)
	parsingResult.ProjectParser()

	fileCount := len(parsingResult.Js_Data)
	fmt.Printf("  - 发现 %d 个 JS/TS 文件\n", fileCount)

	if fileCount == 0 {
		return nil, fmt.Errorf("no JS/TS files found in project")
	}

	// 存储解析结果到上下文（供后续阶段使用）
	ctx.SetResult("projectParser", parsingResult)

	return parsingResult, nil
}

// Skip 判断是否跳过此阶段
func (s *ProjectParserStage) Skip(ctx *AnalysisContext) bool {
	return false
}

// =============================================================================
// 影响分析阶段（支持组件库自动检测）
// =============================================================================

// ImpactAnalysisStage 影响分析阶段
// 自动检测是否为组件库项目，执行相应的影响分析
type ImpactAnalysisStage struct {
	manifestPath       string // 组件清单路径（可选，如果不提供则自动检测）
	maxDepth           int
	isComponentLibrary bool // 是否为组件库项目
}

// NewImpactAnalysisStage 创建影响分析阶段
func NewImpactAnalysisStage(manifestPath string, maxDepth int) *ImpactAnalysisStage {
	if maxDepth <= 0 {
		maxDepth = 10
	}
	return &ImpactAnalysisStage{
		manifestPath: manifestPath,
		maxDepth:     maxDepth,
	}
}

// Name 返回阶段名称
func (s *ImpactAnalysisStage) Name() string {
	if s.isComponentLibrary {
		return "影响分析（组件级）"
	}
	return "影响分析（文件级）"
}

// Execute 执行影响分析
func (s *ImpactAnalysisStage) Execute(ctx *AnalysisContext) (interface{}, error) {
	// 1. 获取项目解析结果
	parsingResult, exists := ctx.GetResult("projectParser")
	if !exists {
		return nil, fmt.Errorf("project parser result not found")
	}

	parsedResult, ok := parsingResult.(*projectParser.ProjectParserResult)
	if !ok {
		return nil, fmt.Errorf("invalid project parser result type")
	}

	// 2. 检测是否为组件库项目
	manifest, err := s.detectComponentLibrary(ctx)
	if err != nil {
		return nil, fmt.Errorf("detect component library failed: %w", err)
	}

	s.isComponentLibrary = (manifest != nil)
	if s.isComponentLibrary {
		fmt.Printf("  - 检测到组件库项目 (%d 个组件)\n", len(manifest.Components))
	} else {
		fmt.Println("  - 非组件库项目，执行文件级影响分析")
	}

	// 3. 获取符号分析结果（文件变更）
	symbolResults, exists := ctx.GetResult("符号分析")
	if !exists {
		return nil, fmt.Errorf("symbol analysis result not found")
	}

	// 4. 转换为变更符号列表
	changedSymbols := s.convertToChangedSymbols(symbolResults)
	if len(changedSymbols) == 0 {
		fmt.Println("  - 没有检测到符号变更")
		return &ImpactAnalysisResult{
			IsComponentLibrary: s.isComponentLibrary,
			FileResult:         nil,
			ComponentResult:    nil,
		}, nil
	}

	fmt.Printf("  - 检测到 %d 个变更符号\n", len(changedSymbols))

	// 5. 执行文件级影响分析
	fileResult, err := s.runFileLevelAnalysis(parsedResult, changedSymbols)
	if err != nil {
		return nil, fmt.Errorf("file level analysis failed: %w", err)
	}

	fmt.Printf("  - 文件级分析完成: %d 个直接变更, %d 个间接受影响\n",
		len(fileResult.Changes), len(fileResult.Impact))

	// 6. 如果是组件库，执行组件级影响分析
	var componentResult *component_analyzer.Result
	if s.isComponentLibrary {
		componentResult, err = s.runComponentLevelAnalysis(parsedResult, manifest, fileResult)
		if err != nil {
			return nil, fmt.Errorf("component level analysis failed: %w", err)
		}

		fmt.Printf("  - 组件级分析完成: %d 个组件变更, %d 个组件受影响\n",
			len(componentResult.Changes), len(componentResult.Impact))
	}

	// 7. 返回综合结果
	return &ImpactAnalysisResult{
		IsComponentLibrary: s.isComponentLibrary,
		FileResult:         fileResult,
		ComponentResult:    componentResult,
	}, nil
}

// Skip 判断是否跳过此阶段
func (s *ImpactAnalysisStage) Skip(ctx *AnalysisContext) bool {
	_, exists := ctx.GetResult("projectParser")
	return !exists
}

// detectComponentLibrary 检测是否为组件库项目
func (s *ImpactAnalysisStage) detectComponentLibrary(ctx *AnalysisContext) (*impact_analysis.ComponentManifest, error) {
	// 确定组件清单路径
	manifestPath := s.manifestPath
	if manifestPath == "" {
		// 默认路径: {projectRoot}/.analyzer/component-manifest.json
		manifestPath = filepath.Join(ctx.ProjectRoot, ".analyzer", "component-manifest.json")
	}

	// 检查文件是否存在
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, nil // 不是组件库项目
	}

	// 读取并解析组件清单
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read component manifest failed: %w", err)
	}

	var manifest impact_analysis.ComponentManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parse component manifest failed: %w", err)
	}

	// 将相对路径转换为绝对路径
	for i := range manifest.Components {
		if !filepath.IsAbs(manifest.Components[i].Entry) {
			manifest.Components[i].Entry = filepath.Join(ctx.ProjectRoot, manifest.Components[i].Entry)
		}
	}

	return &manifest, nil
}

// runFileLevelAnalysis 执行文件级影响分析
func (s *ImpactAnalysisStage) runFileLevelAnalysis(
	parsingResult *projectParser.ProjectParserResult,
	changedSymbols []file_analyzer.ChangedSymbol,
) (*file_analyzer.Result, error) {
	// 创建文件级分析器
	analyzer := file_analyzer.NewAnalyzer(parsingResult)

	// 构建输入
	input := &file_analyzer.Input{
		ChangedSymbols: changedSymbols,
	}

	// 执行分析
	result, err := analyzer.Analyze(input)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// runComponentLevelAnalysis 执行组件级影响分析
func (s *ImpactAnalysisStage) runComponentLevelAnalysis(
	parsingResult *projectParser.ProjectParserResult,
	manifest *impact_analysis.ComponentManifest,
	fileResult *file_analyzer.Result,
) (*component_analyzer.Result, error) {
	// 创建组件级分析器
	analyzer := component_analyzer.NewAnalyzer(manifest, parsingResult, s.maxDepth)

	// 转换文件结果为代理类型
	input := &component_analyzer.Input{
		FileResult: &component_analyzer.FileAnalysisResultProxy{
			Changes:      convertFileChangeInfos(fileResult.Changes),
			Impact:       convertFileImpactInfos(fileResult.Impact),
			DepGraph:     extractDepGraphFromContext(fileResult),
			RevDepGraph:  extractRevDepGraphFromContext(fileResult),
			ExternalDeps: extractExternalDepsFromContext(fileResult),
		},
	}

	// 执行分析
	result, err := analyzer.Analyze(input)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// convertToChangedSymbols 转换符号分析结果为变更符号列表
// 注意：这里从 diff 结果中提取文件变更，并创建 ChangedSymbol 条目
// 实际的符号级分析应该由 symbol_analysis stage 提供
func (s *ImpactAnalysisStage) convertToChangedSymbols(symbolResults interface{}) []file_analyzer.ChangedSymbol {
	symbols := make([]file_analyzer.ChangedSymbol, 0)

	resultMap, ok := symbolResults.(map[string]interface{})
	if !ok {
		return symbols
	}

	diffResults, ok := resultMap["diff_results"].([]interface{})
	if !ok {
		return symbols
	}

	for _, item := range diffResults {
		diffResult, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// 提取文件路径
		newPath, _ := diffResult["new_path"].(string)
		oldPath, _ := diffResult["old_path"].(string)

		// 确定文件路径
		filePath := newPath
		if filePath == "" || filePath == "/dev/null" {
			filePath = oldPath
		}

		if filePath == "" || filePath == "/dev/null" {
			continue
		}

		// 从文件路径提取符号名称（简化处理：使用文件名作为符号名）
		// 例如: src/components/Button/Button.tsx -> Button
		// 实际应该从 symbol_analysis 结果获取准确的符号列表
		symbolName := extractSymbolNameFromPath(filePath)

		// 创建 ChangedSymbol 条目
		// 注意：这里使用默认的导出类型，实际应该从 symbol_analysis 获取
		symbols = append(symbols, file_analyzer.ChangedSymbol{
			Name:       symbolName,
			FilePath:   filePath,
			ExportType: symbol_analysis.ExportTypeDefault, // 默认使用 default 导出
		})
	}

	return symbols
}

// extractSymbolNameFromPath 从文件路径提取符号名称
// 例如: src/components/Button/Button.tsx -> Button
func extractSymbolNameFromPath(filePath string) string {
	// 获取文件名（不含扩展名）
	parts := strings.Split(filepath.Base(filePath), ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return "Unknown"
}

// =============================================================================
// 影响分析结果
// =============================================================================

// ImpactAnalysisResult 影响分析结果
type ImpactAnalysisResult struct {
	IsComponentLibrary bool                       // 是否为组件库项目
	FileResult         *file_analyzer.Result      // 文件级分析结果（所有项目）
	ComponentResult    *component_analyzer.Result // 组件级分析结果（仅组件库）
}

// GetSummary 获取分析摘要
func (r *ImpactAnalysisResult) GetSummary() string {
	if r.IsComponentLibrary && r.ComponentResult != nil {
		return fmt.Sprintf("组件级影响分析: %d 个组件变更, %d 个组件受影响",
			len(r.ComponentResult.Changes), len(r.ComponentResult.Impact))
	}

	if r.FileResult != nil {
		return fmt.Sprintf("文件级影响分析: %d 个文件变更, %d 个文件受影响",
			len(r.FileResult.Changes), len(r.FileResult.Impact))
	}

	return "影响分析: 无变更"
}

// =============================================================================
// 辅助函数
// =============================================================================

// convertFileChangeInfos 转换文件变更信息
func convertFileChangeInfos(changes []file_analyzer.FileChangeInfo) []component_analyzer.FileChangeInfoProxy {
	result := make([]component_analyzer.FileChangeInfoProxy, len(changes))
	for i, c := range changes {
		result[i] = component_analyzer.FileChangeInfoProxy{
			Path:        c.Path,
			ChangeType:  impact_analysis.ChangeType(c.ChangeType),
			SymbolCount: c.SymbolCount,
		}
	}
	return result
}

// convertFileImpactInfos 转换文件影响信息
func convertFileImpactInfos(impacts []file_analyzer.FileImpactInfo) []component_analyzer.FileImpactInfoProxy {
	result := make([]component_analyzer.FileImpactInfoProxy, len(impacts))
	for i, imp := range impacts {
		result[i] = component_analyzer.FileImpactInfoProxy{
			Path:        imp.Path,
			ImpactLevel: impact_analysis.ImpactLevel(imp.ImpactLevel),
			ImpactType:  impact_analysis.ImpactType(imp.ImpactType),
			ChangePaths: imp.ChangePaths,
		}
	}
	return result
}

// extractDepGraphFromContext 从上下文提取依赖图
func extractDepGraphFromContext(result *file_analyzer.Result) map[string][]string {
	// TODO: 从结果中提取依赖图
	// 目前需要重新构建，或者修改 file_analyzer.Result 以包含依赖图
	return make(map[string][]string)
}

// extractRevDepGraphFromContext 从上下文提取反向依赖图
func extractRevDepGraphFromContext(result *file_analyzer.Result) map[string][]string {
	return make(map[string][]string)
}

// extractExternalDepsFromContext 从上下文提取外部依赖
func extractExternalDepsFromContext(result *file_analyzer.Result) map[string][]string {
	return make(map[string][]string)
}
