// Package gitlab provides GitLab integration capabilities for analyzer-ts.
package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	componentDepsV2 "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/component_deps_v2"
	impactAnalysis "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/impact_analysis"
	projectAnalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

// =============================================================================
// GitLabIntegration - GitLab 集成器
// =============================================================================

// GitLabIntegration GitLab 集成器
type GitLabIntegration struct {
	client     *Client
	diffParser *DiffParser
	mrService  *MRService
	formatter  *Formatter
	config     *GitLabConfig
}

// NewGitLabIntegration 创建 GitLab 集成器
func NewGitLabIntegration(config *GitLabConfig) *GitLabIntegration {
	client := NewClient(config.URL, config.Token)

	return &GitLabIntegration{
		client:     client,
		diffParser: NewDiffParser(""), // baseDir 将在运行时设置
		mrService:  NewMRService(client, config.ProjectID, config.MRIID),
		formatter:  NewFormatter(CommentStyleDetailed),
		config:     config,
	}
}

// =============================================================================
// 核心分析流程
// =============================================================================

// RunAnalysis 执行完整的分析流程
func (g *GitLabIntegration) RunAnalysis(ctx context.Context, projectRoot string) error {
	// 设置项目根目录
	g.diffParser.baseDir = projectRoot

	// 1. 获取变更信息
	changeInput, err := g.getChangeInput(ctx)
	if err != nil {
		return fmt.Errorf("get change input failed: %w", err)
	}

	// 2. 运行 component-deps-v2 生成依赖图
	depData, err := g.runComponentDepsV2(ctx, projectRoot)
	if err != nil {
		return fmt.Errorf("component-deps-v2 analysis failed: %w", err)
	}

	// 3. 运行 impact-analysis 分析影响范围
	impactResult, err := g.runImpactAnalysis(ctx, changeInput, depData)
	if err != nil {
		return fmt.Errorf("impact-analysis failed: %w", err)
	}

	// 4. 发布 MR 评论
	if err := g.mrService.PostImpactComment(ctx, impactResult); err != nil {
		return fmt.Errorf("post MR comment failed: %w", err)
	}

	fmt.Println("✅ GitLab integration completed successfully!")
	return nil
}

// =============================================================================
// 阶段 1: 获取变更信息
// =============================================================================

// getChangeInput 根据 DiffSource 模式获取变更信息
func (g *GitLabIntegration) getChangeInput(ctx context.Context) (*ChangeInput, error) {
	var lineSet ChangedLineSetOfFiles
	var err error

	switch g.config.DiffSource {
	case "file":
		// 从 diff 文件读取
		lineSet, err = g.diffParser.ParseDiffFile(g.config.DiffFile)
	case "api":
		// 从 GitLab API 获取 MR diff
		var diffFiles []DiffFile
		diffFiles, err = g.client.GetMergeRequestDiff(ctx, g.config.ProjectID, g.config.MRIID)
		if err == nil {
			lineSet, err = g.diffParser.ParseDiffFiles(diffFiles)
		}
	case "diff":
		// 执行 git diff 命令
		if g.config.DiffSHA != "" {
			// 解析 SHA 范围 "baseSHA...headSHA"
			shas := strings.Split(g.config.DiffSHA, "...")
			if len(shas) == 2 {
				lineSet, err = g.diffParser.ParseFromGit(shas[0], shas[1])
			} else {
				return nil, fmt.Errorf("invalid SHA format, expected 'base...head': %s", g.config.DiffSHA)
			}
		} else {
			// 自动检测：从环境变量获取 SHA
			baseSHA := os.Getenv("CI_MERGE_REQUEST_DIFF_BASE_SHA")
			headSHA := "HEAD"
			lineSet, err = g.diffParser.ParseFromGit(baseSHA, headSHA)
		}
	default:
		// auto 模式：自动检测
		lineSet, err = g.autoDetectDiffSource(ctx)
	}

	if err != nil {
		return nil, err
	}

	// 转换为文件级别（兼容当前 impact-analysis）
	return g.diffParser.GetChangedFiles(lineSet), nil
}

// autoDetectDiffSource 自动检测 diff 来源
func (g *GitLabIntegration) autoDetectDiffSource(ctx context.Context) (ChangedLineSetOfFiles, error) {
	// 优先级 1: 从 GitLab API 获取
	diffFiles, err := g.client.GetMergeRequestDiff(ctx, g.config.ProjectID, g.config.MRIID)
	if err == nil && len(diffFiles) > 0 {
		fmt.Println("ℹ️  Using GitLab API for diff")
		return g.diffParser.ParseDiffFiles(diffFiles)
	}

	// 优先级 2: 执行 git diff 命令
	baseSHA := os.Getenv("CI_MERGE_REQUEST_DIFF_BASE_SHA")
	if baseSHA != "" {
		fmt.Println("ℹ️  Using git diff for diff")
		return g.diffParser.ParseFromGit(baseSHA, "HEAD")
	}

	// 优先级 3: 从环境变量读取 diff 文件
	diffFile := os.Getenv("CI_DIFF_FILE")
	if diffFile != "" {
		fmt.Println("ℹ️  Using diff file from environment")
		return g.diffParser.ParseDiffFile(diffFile)
	}

	return nil, fmt.Errorf("no diff source available")
}

// =============================================================================
// 阶段 2 & 3: 运行分析
// =============================================================================

// runComponentDepsV2 运行组件依赖分析
func (g *GitLabIntegration) runComponentDepsV2(ctx context.Context, projectRoot string) (*ComponentDepsData, error) {
	// 如果提供了依赖文件，直接加载（避免重复解析）
	if g.config.DepsFile != "" {
		fmt.Println("📦 从文件加载依赖数据:", g.config.DepsFile)
		return g.loadDependencyDataFromFile(g.config.DepsFile)
	}

	// 运行 component-deps-v2 分析器
	fmt.Println("🔍 运行组件依赖分析...")

	// 1. 创建项目解析器配置
	parserConfig := projectParser.NewProjectParserConfig(
		projectRoot,
		[]string{}, // exclude patterns
		false,     // isMonorepo
		[]string{},// strip paths
	)

	// 2. 解析项目
	fmt.Println("  - 解析项目 AST...")
	parsingResult := projectParser.NewProjectParserResult(parserConfig)
	parsingResult.ProjectParser()
	fmt.Printf("  - 发现 %d 个 JS/TS 文件\n", len(parsingResult.Js_Data))

	// 3. 创建项目上下文
	projectCtx := &projectAnalyzer.ProjectContext{
		ProjectRoot:   projectRoot,
		Exclude:       []string{},
		IsMonorepo:    false,
		ParsingResult: parsingResult,
	}

	// 4. 创建并配置 component-deps-v2 分析器
	analyzer := &componentDepsV2.ComponentDepsV2Analyzer{}
	manifestPath := g.config.ManifestPath
	if manifestPath == "" {
		manifestPath = "component-manifest.json"
	}
	params := map[string]string{"manifest": manifestPath}
	if err := analyzer.Configure(params); err != nil {
		return nil, fmt.Errorf("configure component-deps-v2 failed: %w", err)
	}

	// 5. 运行分析
	result, err := analyzer.Analyze(projectCtx)
	if err != nil {
		return nil, fmt.Errorf("component-deps-v2 analysis failed: %w", err)
	}

	// 6. 转换结果为 ComponentDepsData
	depsResult := result.(*componentDepsV2.ComponentDepsV2Result)
	return &ComponentDepsData{
		DepGraph:    depsResult.DepGraph,
		RevDepGraph: depsResult.RevDepGraph,
		Meta: struct {
			Version        string `json:"version"`
			LibraryName    string `json:"libraryName"`
			ComponentCount int    `json:"componentCount"`
		}{
			Version:        depsResult.Meta.Version,
			LibraryName:    depsResult.Meta.LibraryName,
			ComponentCount: depsResult.Meta.ComponentCount,
		},
	}, nil
}

// runImpactAnalysis 运行影响分析
func (g *GitLabIntegration) runImpactAnalysis(ctx context.Context, changeInput *ChangeInput, depData *ComponentDepsData) (*ImpactAnalysisResult, error) {
	fmt.Println("📊 运行影响分析...")

	// 1. 创建临时文件存储依赖数据
	tmpFile, err := os.CreateTemp("", "deps-*.json")
	if err != nil {
		return nil, fmt.Errorf("create temp file failed: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// 2. 序列化依赖数据并写入临时文件
	depsJSON := map[string]interface{}{"component-deps-v2": depData}
	depsBytes, err := json.Marshal(depsJSON)
	if err != nil {
		return nil, fmt.Errorf("marshal deps data failed: %w", err)
	}
	if err := os.WriteFile(tmpFile.Name(), depsBytes, 0644); err != nil {
		return nil, fmt.Errorf("write deps file failed: %w", err)
	}

	// 3. 创建 impact-analysis 分析器
	analyzer := impactAnalysis.NewAnalyzer()

	// 4. 序列化 changeInput
	changeBytes, err := json.Marshal(changeInput)
	if err != nil {
		return nil, fmt.Errorf("marshal change input failed: %w", err)
	}

	// 5. 配置分析器
	params := map[string]string{
		"changes":  string(changeBytes),
		"depsFile": tmpFile.Name(),
		"maxDepth": fmt.Sprintf("%d", g.config.MaxDepth),
	}
	if err := analyzer.Configure(params); err != nil {
		return nil, fmt.Errorf("configure impact-analysis failed: %w", err)
	}

	// 6. 运行分析（使用简单的 ProjectContext，因为 impact-analysis 不需要 ParsingResult）
	projectCtx := &projectAnalyzer.ProjectContext{
		ProjectRoot:   g.diffParser.baseDir,
		Exclude:       []string{},
		IsMonorepo:    false,
		ParsingResult: nil, // impact-analysis 目前不需要 ParsingResult
	}

	result, err := analyzer.Analyze(projectCtx)
	if err != nil {
		return nil, fmt.Errorf("impact-analysis failed: %w", err)
	}

	// 7. 转换结果
	impactResult := result.(*impactAnalysis.ImpactAnalysisResult)
	return impactResult, nil
}

// =============================================================================
// 数据加载
// =============================================================================

// ComponentDepsData 依赖数据结构（简化版）
type ComponentDepsData struct {
	DepGraph    map[string][]string `json:"depGraph"`
	RevDepGraph map[string][]string `json:"revDepGraph"`
	Meta        struct {
		Version       string `json:"version"`
		LibraryName   string `json:"libraryName"`
		ComponentCount int   `json:"componentCount"`
	} `json:"meta"`
}

// loadDependencyDataFromFile 从文件加载依赖数据
func (g *GitLabIntegration) loadDependencyDataFromFile(filePath string) (*ComponentDepsData, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read deps file failed: %w", err)
	}

	// 尝试解析包裹格式 {"component-deps-v2": {...}}
	var wrappedData map[string]json.RawMessage
	if err := json.Unmarshal(data, &wrappedData); err == nil {
		if raw, exists := wrappedData["component-deps-v2"]; exists {
			var depData ComponentDepsData
			if err := json.Unmarshal(raw, &depData); err == nil {
				return &depData, nil
			}
		}
	}

	// 直接解析
	var depData ComponentDepsData
	if err := json.Unmarshal(data, &depData); err != nil {
		return nil, fmt.Errorf("parse deps data failed: %w", err)
	}

	return &depData, nil
}

// =============================================================================
// 工厂函数
// =============================================================================

// DetectAndCreateConfig 从环境变量自动检测并创建 GitLab 配置
func DetectAndCreateConfig() (*GitLabConfig, error) {
	config := &GitLabConfig{
		DiffSource: string(DiffSourceAuto),
		MaxDepth:  10,
	}

	// GitLab 连接信息
	if url := os.Getenv("CI_SERVER_URL"); url != "" {
		config.URL = url
	}
	if token := os.Getenv("GITLAB_TOKEN"); token != "" {
		config.Token = token
	}

	// MR 信息
	if projectID := os.Getenv("CI_PROJECT_ID"); projectID != "" {
		id, err := strconv.Atoi(projectID)
		if err != nil {
			return nil, fmt.Errorf("invalid CI_PROJECT_ID: %w", err)
		}
		config.ProjectID = id
	}
	if mrIID := os.Getenv("CI_MERGE_REQUEST_IID"); mrIID != "" {
		id, err := strconv.Atoi(mrIID)
		if err != nil {
			return nil, fmt.Errorf("invalid CI_MERGE_REQUEST_IID: %w", err)
		}
		config.MRIID = id
	}

	// 分析参数
	if manifest := os.Getenv("ANALYZER_MANIFEST_PATH"); manifest != "" {
		config.ManifestPath = manifest
	}
	if depsFile := os.Getenv("ANALYZER_DEPS_FILE"); depsFile != "" {
		config.DepsFile = depsFile
	}

	return config, nil
}

// =============================================================================
// JSON 序列化支持
// =============================================================================

func init() {
	// 确保导入但未使用的包不会导致编译错误
	_ = json.Unmarshal
	_ = strconv.Atoi
}
