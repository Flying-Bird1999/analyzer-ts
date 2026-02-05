// Package cmd 提供代码影响分析命令
package cmd

// ImpactCmd 代码影响分析命令
//
// 快速开始：
//   ./analyzer-ts impact --project-root $(pwd) --diff-string "$(git diff)"
//
// 常用示例：
//   1. 本地未提交变更：
//      ./analyzer-ts impact --project-root $(pwd) --diff-string "$(git diff)" --format summary
//
//   2. 分支对比：
//      ./analyzer-ts impact --project-root $(pwd) --git-diff "main...HEAD"
//
//   3. package.json 配置：
//      "analyze:impact": "analyzer-ts impact --project-root $(pwd) --diff-string \"$(git diff)\""
//
//   4. Monorepo：
//      ./analyzer-ts impact --project-root $(pwd) --git-root /path/to/monorepo --git-diff "main...HEAD"
//
//   5. 排除文件：
//      ./analyzer-ts impact --project-root $(pwd) --diff-string "$(git diff)" --exclude "**/lib/**" --exclude "**/tests/**"
//
// Glob 模式说明：
//   dist/**         仅匹配根目录的 dist 文件夹
//   **/dist/**      仅匹配嵌套的 dist 文件夹（不匹配根目录）
//   **/dist         匹配任意层级的 dist 文件夹（包括根目录）
//   **/*.test.tsx   匹配所有 .test.tsx 文件
//   node_modules    始终自动排除（硬编码）
//
// 参数说明：
//   必需：
//     --project-root <path>    项目根目录（绝对路径）
//   输入（三选一）：
//     --diff-string "$(git diff)"     直接传入 diff（注意：必须用双引号）
//     --diff-file <path>              从文件读取 diff
//     --git-diff "main...HEAD"        git 命令对比
//   可选：
//     --git-root <path>         Git 仓库根（默认=project-root）
//     --manifest <path>         组件清单路径
//     --format json|pretty|summary  输出格式（默认 json）
//     --output <path>           输出文件
//     --exclude <pattern>       排除 glob 模式
//     --max-depth <n>           最大深度（默认 10）
//     --quiet                   静默模式
//
// 输出格式：
//   json    - 紧凑 JSON，程序解析（默认）
//   pretty  - 美化 JSON，人工阅读
//   summary - 简要摘要，快速查看

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Flying-Bird1999/analyzer-ts/pkg/pipeline"
	"github.com/spf13/cobra"
)

// =============================================================================
// 影响分析命令配置
// =============================================================================

var (
	// 输入配置
	diffString   string // 直接传入 diff 字符串
	diffFile     string // 从文件读取 diff
	gitDiffArgs  string // git diff 参数 (如 "HEAD~1 HEAD")
	gitlabAPI    bool   // 是否使用 GitLab API
	gitlabProjID int    // GitLab 项目 ID
	gitlabMRIID  int    // GitLab MR IID
	gitlabToken  string // GitLab API token
	gitlabURL    string // GitLab API URL

	// 项目配置
	projectRoot  string // 项目根目录（必需）
	gitRoot      string // Git 仓库根目录（可选，默认等于 projectRoot）
	manifestPath string // 组件清单路径（可选）
	maxDepth     int    // 影响分析最大深度
	// excludePaths 已在 scan.go 中声明（包级别共享变量）

	// 输出配置
	outputFile   string // 输出文件路径（可选，默认 stdout）
	outputFormat string // 输出格式：json | pretty | summary
	verbose      bool   // 详细输出
	showSymbols  bool   // 显示符号级分析结果
	quiet        bool   // 静默模式，只输出结果
)

// ImpactCmd 代码影响分析命令
var ImpactCmd = &cobra.Command{
	Use:   "impact",
	Short: "分析代码变更的影响范围",
	Long: `分析代码变更（Git diff）对项目的影响范围，包括文件级和组件级影响。

输入方式（任选其一）：
  --diff-string "$(git diff)"    直接传入 diff
  --diff-file <path>             从文件读取
  --git-diff "main...HEAD"       git 命令对比

输出：变更文件列表、受影响文件、受影响组件（如果有组件清单）

示例：
  # 本地未提交变更
  analyzer-ts impact --project-root $(pwd) --diff-string "$(git diff)"

  # 分支对比
  analyzer-ts impact --project-root $(pwd) --git-diff "main...HEAD"
`,
	RunE: runImpact,
}

func init() {
	// 输入方式（互斥，使用时会校验）
	ImpactCmd.Flags().StringVar(&diffString, "diff-string", "", "直接传入 diff 字符串")
	ImpactCmd.Flags().StringVar(&diffFile, "diff-file", "", "从文件读取 diff")
	ImpactCmd.Flags().StringVar(&gitDiffArgs, "git-diff", "", "执行 git diff 命令（参数如 'HEAD~1 HEAD'）")
	ImpactCmd.Flags().BoolVar(&gitlabAPI, "gitlab-api", false, "从 GitLab API 获取 diff")

	// GitLab API 配置
	ImpactCmd.Flags().IntVar(&gitlabProjID, "gitlab-project-id", 0, "GitLab 项目 ID")
	ImpactCmd.Flags().IntVar(&gitlabMRIID, "gitlab-mr-iid", 0, "GitLab MR IID")
	ImpactCmd.Flags().StringVar(&gitlabToken, "gitlab-token", "", "GitLab API Token（也可通过 GITLAB_TOKEN 环境变量）")
	ImpactCmd.Flags().StringVar(&gitlabURL, "gitlab-url", "https://gitlab.com", "GitLab API URL")

	// 项目配置（必需）
	ImpactCmd.Flags().StringVar(&projectRoot, "project-root", "", "项目根目录（必需）")
	ImpactCmd.Flags().StringVar(&gitRoot, "git-root", "", "Git 仓库根目录（可选，默认等于 projectRoot）")
	ImpactCmd.Flags().StringVar(&manifestPath, "manifest", "", "组件清单路径（可选，用于组件级分析）")
	ImpactCmd.Flags().StringSliceVarP(&excludePaths, "exclude", "x", []string{}, "要排除的 glob 模式（如 **/*.test.tsx, **/stories/**）")

	// 分析配置
	ImpactCmd.Flags().IntVar(&maxDepth, "max-depth", 10, "影响分析最大深度")

	// 输出配置
	ImpactCmd.Flags().StringVarP(&outputFile, "output", "o", "", "输出文件路径（可选，默认 stdout）")
	ImpactCmd.Flags().StringVar(&outputFormat, "format", "json", "输出格式：json | pretty | summary")
	ImpactCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "详细输出")
	ImpactCmd.Flags().BoolVar(&showSymbols, "show-symbols", false, "显示符号级分析结果")
	ImpactCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "静默模式，只输出结果")

	// 标记必需参数
	ImpactCmd.MarkFlagRequired("project-root")
}

// =============================================================================
// 命令执行逻辑
// =============================================================================

func runImpact(cmd *cobra.Command, args []string) error {
	// 1. 参数校验
	if err := validateFlags(); err != nil {
		return fmt.Errorf("参数校验失败: %w", err)
	}

	// 2. 确定输入源
	source, client, err := determineDiffSource()
	if err != nil {
		return fmt.Errorf("确定输入源失败: %w", err)
	}

	// 3. 构建 Pipeline 配置
	config := buildPipelineConfig(source, client)

	// 4. 执行分析
	if !quiet {
		fmt.Printf("🔍 开始分析代码变更影响...\n")
		fmt.Printf("📁 项目路径: %s\n", projectRoot)
		if gitRoot != "" {
			fmt.Printf("📁 Git 仓库根: %s\n", gitRoot)
		}
		fmt.Printf("📥 输入方式: %s\n\n", sourceDesc(source))
	}

	ctx := context.Background()
	analysisCtx := pipeline.NewAnalysisContext(ctx, projectRoot, nil)

	// 设置排除路径
	if len(excludePaths) > 0 {
		analysisCtx.ExcludePaths = excludePaths
		if !quiet {
			fmt.Printf("🚫 排除模式: %v\n", excludePaths)
		}
	}

	// 如果是 diff 字符串输入，通过 context 传递
	if source == pipeline.DiffSourceString && diffString != "" {
		analysisCtx.SetOption("diffString", diffString)
	}

	pipe := pipeline.NewGitLabPipeline(config)

	startTime := time.Now()
	result, err := pipe.Execute(analysisCtx)
	elapsed := time.Since(startTime)

	if err != nil {
		return fmt.Errorf("分析执行失败: %w", err)
	}

	// 5. 构建输出
	output, err := buildOutput(result)
	if err != nil {
		return fmt.Errorf("构建输出失败: %w", err)
	}

	// 6. 输出结果
	if !quiet {
		fmt.Printf("\n✅ 分析完成! (耗时: %s)\n", elapsed)
	}

	if err := writeOutput(output); err != nil {
		return fmt.Errorf("写入输出失败: %w", err)
	}

	return nil
}

// =============================================================================
// 辅助函数
// =============================================================================

// validateFlags 校验命令行参数
func validateFlags() error {
	// 检查项目根目录是否存在
	if projectRoot == "" {
		return fmt.Errorf("--project-root 是必需参数")
	}

	if _, err := os.Stat(projectRoot); os.IsNotExist(err) {
		return fmt.Errorf("项目根目录不存在: %s", projectRoot)
	}

	// 检查输入方式（必须有且仅有一种）
	inputCount := 0
	if diffString != "" {
		inputCount++
	}
	if diffFile != "" {
		inputCount++
	}
	if gitDiffArgs != "" {
		inputCount++
	}
	if gitlabAPI {
		inputCount++
	}

	if inputCount == 0 {
		return fmt.Errorf("必须指定一种输入方式：--diff-string, --diff-file, --git-diff, 或 --gitlab-api")
	}

	if inputCount > 1 {
		return fmt.Errorf("只能指定一种输入方式")
	}

	// 如果使用 GitLab API，检查相关参数
	if gitlabAPI {
		if gitlabProjID == 0 {
			return fmt.Errorf("使用 --gitlab-api 时必须指定 --gitlab-project-id")
		}
		if gitlabMRIID == 0 {
			return fmt.Errorf("使用 --gitlab-api 时必须指定 --gitlab-mr-iid")
		}
		if gitlabToken == "" {
			gitlabToken = os.Getenv("GITLAB_TOKEN")
			if gitlabToken == "" {
				return fmt.Errorf("使用 --gitlab-api 时必须指定 --gitlab-token 或设置 GITLAB_TOKEN 环境变量")
			}
		}
	}

	// 检查输出格式
	if outputFormat != "json" && outputFormat != "pretty" && outputFormat != "summary" {
		return fmt.Errorf("无效的输出格式: %s，必须是 json、pretty 或 summary", outputFormat)
	}

	return nil
}

// determineDiffSource 确定使用哪种 diff 输入源
func determineDiffSource() (pipeline.DiffSourceType, pipeline.GitLabClient, error) {
	var source pipeline.DiffSourceType
	var client pipeline.GitLabClient

	// GitLab API
	if gitlabAPI {
		source = pipeline.DiffSourceAPI
		// TODO: 创建实际的 GitLab 客户端
		client = nil
		return source, client, nil
	}

	// diff 字符串
	if diffString != "" {
		source = pipeline.DiffSourceString
		return source, nil, nil
	}

	// diff 文件
	if diffFile != "" {
		source = pipeline.DiffSourceFile
		return source, nil, nil
	}

	// git diff 命令
	if gitDiffArgs != "" {
		source = pipeline.DiffSourceSHA
		return source, nil, nil
	}

	return source, nil, fmt.Errorf("未知的输入源")
}

// buildPipelineConfig 构建 Pipeline 配置
func buildPipelineConfig(source pipeline.DiffSourceType, client pipeline.GitLabClient) *pipeline.GitLabPipelineConfig {
	config := &pipeline.GitLabPipelineConfig{
		DiffSource:   source,
		DiffFile:     diffFile,
		DiffSHA:      gitDiffArgs,
		ProjectRoot:  projectRoot,
		GitRoot:      gitRoot,
		ProjectID:    gitlabProjID,
		MRIID:        gitlabMRIID,
		ManifestPath: manifestPath,
		MaxDepth:     maxDepth,
		Client:       client,
	}

	return config
}

// sourceDesc 获取输入源的描述
func sourceDesc(source pipeline.DiffSourceType) string {
	switch source {
	case pipeline.DiffSourceString:
		return "diff 字符串"
	case pipeline.DiffSourceFile:
		return fmt.Sprintf("diff 文件: %s", diffFile)
	case pipeline.DiffSourceSHA:
		return fmt.Sprintf("git diff: %s", gitDiffArgs)
	case pipeline.DiffSourceAPI:
		return fmt.Sprintf("GitLab API: Project %d, MR !%d", gitlabProjID, gitlabMRIID)
	default:
		return "未知"
	}
}

// =============================================================================
// 输出构建
// =============================================================================

// AnalysisOutput 分析结果输出结构
type AnalysisOutput struct {
	Meta struct {
		Version     string `json:"version"`     // 工具版本
		ProjectRoot string `json:"projectRoot"` // 项目根目录
		GitRoot     string `json:"gitRoot"`     // Git 仓库根目录
		AnalyzedAt  string `json:"analyzedAt"`  // 分析时间
		Duration    string `json:"duration"`    // 分析耗时
		InputSource string `json:"inputSource"` // 输入方式
	} `json:"meta"`

	Input struct {
		DiffSummary string   `json:"diffSummary"` // diff 摘要
		Files       []string `json:"files"`       // 变更的文件列表
	} `json:"input"`

	SymbolAnalysis *SymbolAnalysisOutput `json:"symbolAnalysis,omitempty"` // 符号分析结果（可选）

	FileAnalysis struct {
		Meta struct {
			TotalFileCount   int `json:"totalFileCount"`   // 总文件数
			ChangedFileCount int `json:"changedFileCount"` // 变更文件数
			ImpactFileCount  int `json:"impactFileCount"`  // 受影响文件数
		} `json:"meta"`
		Changes []FileChangeOutput `json:"changes"` // 直接变更的文件
		Impact  []FileImpactOutput `json:"impact"`  // 间接受影响的文件
	} `json:"fileAnalysis"`

	ComponentAnalysis *ComponentAnalysisOutput `json:"componentAnalysis,omitempty"` // 组件分析结果（可选）
}

// SymbolAnalysisOutput 符号分析输出
type SymbolAnalysisOutput struct {
	Meta struct {
		AnalyzedFileCount int `json:"analyzedFileCount"` // 分析的文件数
		TotalSymbolCount  int `json:"totalSymbolCount"`  // 总符号数
	} `json:"meta"`
	Files []SymbolFileOutput `json:"files"` // 符号文件列表
}

// SymbolFileOutput 符号文件输出
type SymbolFileOutput struct {
	FilePath        string         `json:"filePath"`
	IsSymbolFile    bool           `json:"isSymbolFile"`
	AffectedSymbols []SymbolOutput `json:"affectedSymbols"`
	ChangedLines    []int          `json:"changedLines"`
}

// SymbolOutput 符号输出
type SymbolOutput struct {
	Name         string `json:"name"`
	Kind         string `json:"kind"`
	StartLine    int    `json:"startLine"`
	EndLine      int    `json:"endLine"`
	ChangedLines []int  `json:"changedLines"`
	ChangeType   string `json:"changeType"`
	IsExported   bool   `json:"isExported"`
}

// FileChangeOutput 文件变更输出
type FileChangeOutput struct {
	Path        string `json:"path"`
	Type        string `json:"type"`
	SymbolCount int    `json:"symbolCount"`
}

// FileImpactOutput 文件影响输出
type FileImpactOutput struct {
	Path        string   `json:"path"`
	ImpactLevel int      `json:"impactLevel"`
	ImpactType  string   `json:"impactType"`
	ChangePaths []string `json:"changePaths"`
}

// ComponentAnalysisOutput 组件分析输出
type ComponentAnalysisOutput struct {
	Meta struct {
		TotalComponentCount   int `json:"totalComponentCount"`   // 总组件数
		ChangedComponentCount int `json:"changedComponentCount"` // 变更组件数
		ImpactComponentCount  int `json:"impactComponentCount"`  // 受影响组件数
	} `json:"meta"`
	Changes []ComponentChangeOutput `json:"changes"` // 变更的组件
	Impact  []ComponentImpactOutput `json:"impact"`  // 受影响的组件
}

// ComponentChangeOutput 组件变更输出
type ComponentChangeOutput struct {
	Name         string   `json:"name"`
	ChangedFiles []string `json:"changedFiles"`
	SymbolCount  int      `json:"symbolCount"`
}

// ComponentImpactOutput 组件影响输出
type ComponentImpactOutput struct {
	Name        string   `json:"name"`
	ImpactLevel int      `json:"impactLevel"`
	ImpactType  string   `json:"impactType"`
	ChangePaths []string `json:"changePaths"`
}

// buildOutput 构建输出结构
func buildOutput(result *pipeline.PipelineResult) (*AnalysisOutput, error) {
	output := &AnalysisOutput{}

	// 填充元数据
	output.Meta.Version = "1.0.0" // TODO: 从版本信息获取
	output.Meta.ProjectRoot = projectRoot
	output.Meta.GitRoot = gitRoot
	output.Meta.AnalyzedAt = time.Now().Format(time.RFC3339)
	output.Meta.InputSource = sourceDesc(determineSourceType())

	// 获取影响分析结果
	impactResult, ok := result.GetResult("影响分析（文件级）")
	if !ok {
		impactResult, ok = result.GetResult("影响分析（组件级）")
		if !ok {
			return output, nil
		}
	}

	impactAnalysisResult, ok := impactResult.(*pipeline.ImpactAnalysisResult)
	if !ok {
		return output, nil
	}

	// 填充文件级分析结果
	if impactAnalysisResult.FileResult != nil {
		output.FileAnalysis.Meta.TotalFileCount = impactAnalysisResult.FileResult.Meta.TotalFileCount
		output.FileAnalysis.Meta.ChangedFileCount = impactAnalysisResult.FileResult.Meta.ChangedFileCount
		output.FileAnalysis.Meta.ImpactFileCount = impactAnalysisResult.FileResult.Meta.ImpactFileCount

		// 转换相对路径
		for _, change := range impactAnalysisResult.FileResult.Changes {
			relPath, _ := filepath.Rel(projectRoot, change.Path)
			output.FileAnalysis.Changes = append(output.FileAnalysis.Changes, FileChangeOutput{
				Path:        relPath,
				Type:        string(change.ChangeType),
				SymbolCount: change.SymbolCount,
			})
			output.Input.Files = append(output.Input.Files, relPath)
		}

		for _, impact := range impactAnalysisResult.FileResult.Impact {
			relPath, _ := filepath.Rel(projectRoot, impact.Path)
			changePaths := make([]string, len(impact.ChangePaths))
			for i, p := range impact.ChangePaths {
				changePaths[i], _ = filepath.Rel(projectRoot, p)
			}
			output.FileAnalysis.Impact = append(output.FileAnalysis.Impact, FileImpactOutput{
				Path:        relPath,
				ImpactLevel: impact.ImpactLevel,
				ChangePaths: changePaths,
			})
		}
	}

	// 填充组件级分析结果
	if impactAnalysisResult.ComponentResult != nil && impactAnalysisResult.IsComponentLibrary {
		output.ComponentAnalysis = &ComponentAnalysisOutput{}
		output.ComponentAnalysis.Meta.TotalComponentCount = impactAnalysisResult.ComponentResult.Meta.TotalComponentCount
		output.ComponentAnalysis.Meta.ChangedComponentCount = impactAnalysisResult.ComponentResult.Meta.ChangedComponentCount
		output.ComponentAnalysis.Meta.ImpactComponentCount = impactAnalysisResult.ComponentResult.Meta.ImpactComponentCount

		for _, change := range impactAnalysisResult.ComponentResult.Changes {
			changedFiles := make([]string, len(change.ChangedFiles))
			for i, f := range change.ChangedFiles {
				changedFiles[i], _ = filepath.Rel(projectRoot, f)
			}
			output.ComponentAnalysis.Changes = append(output.ComponentAnalysis.Changes, ComponentChangeOutput{
				Name:         change.Name,
				ChangedFiles: changedFiles,
				SymbolCount:  change.SymbolCount,
			})
		}

		for _, impact := range impactAnalysisResult.ComponentResult.Impact {
			changePaths := make([]string, len(impact.ChangePaths))
			for i, p := range impact.ChangePaths {
				changePaths[i], _ = filepath.Rel(projectRoot, p)
			}
			output.ComponentAnalysis.Impact = append(output.ComponentAnalysis.Impact, ComponentImpactOutput{
				Name:        impact.Name,
				ImpactLevel: int(impact.ImpactLevel),
				ChangePaths: changePaths,
			})
		}
	}

	// 填充符号分析结果（可选）
	if showSymbols {
		if symbolResult, ok := result.GetResult("符号分析"); ok {
			if symbolResults, ok := symbolResult.(map[string]interface{}); ok {
				output.SymbolAnalysis = &SymbolAnalysisOutput{}
				output.SymbolAnalysis.Meta.AnalyzedFileCount = len(symbolResults)
				for _, fileResult := range symbolResults {
					// 转换符号结果
					// TODO: 实现符号结果的转换
					if results, ok := fileResult.([]interface{}); ok {
						output.SymbolAnalysis.Meta.TotalSymbolCount += len(results)
					}
				}
			}
		}
	}

	return output, nil
}

// determineSourceType 确定输入源类型（用于元数据）
func determineSourceType() pipeline.DiffSourceType {
	if diffString != "" {
		return pipeline.DiffSourceString
	}
	if diffFile != "" {
		return pipeline.DiffSourceFile
	}
	if gitDiffArgs != "" {
		return pipeline.DiffSourceSHA
	}
	if gitlabAPI {
		return pipeline.DiffSourceAPI
	}
	return ""
}

// writeOutput 写入输出
func writeOutput(output *AnalysisOutput) error {
	var data []byte
	var err error

	switch outputFormat {
	case "json":
		data, err = json.Marshal(output)
	case "pretty":
		data, err = json.MarshalIndent(output, "", "  ")
	case "summary":
		data = []byte(buildSummary(output))
	default:
		data, err = json.Marshal(output)
	}

	if err != nil {
		return err
	}

	// 输出到文件或 stdout
	if outputFile != "" {
		// 确保输出目录存在
		dir := filepath.Dir(outputFile)
		if dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("创建输出目录失败: %w", err)
			}
		}
		return os.WriteFile(outputFile, data, 0644)
	}

	// 输出到 stdout
	fmt.Println(string(data))
	return nil
}

// buildSummary 构建简要摘要
func buildSummary(output *AnalysisOutput) string {
	var summary string

	summary += fmt.Sprintf("代码影响分析结果\n")
	summary += fmt.Sprintf("==================\n\n")
	summary += fmt.Sprintf("变更文件: %d\n", len(output.Input.Files))
	summary += fmt.Sprintf("受影响文件: %d\n", len(output.FileAnalysis.Impact))

	if output.ComponentAnalysis != nil {
		summary += fmt.Sprintf("变更组件: %d\n", len(output.ComponentAnalysis.Changes))
		summary += fmt.Sprintf("受影响组件: %d\n", len(output.ComponentAnalysis.Impact))
	}

	summary += fmt.Sprintf("\n变更的文件:\n")
	for _, file := range output.Input.Files {
		summary += fmt.Sprintf("  - %s\n", file)
	}

	if len(output.FileAnalysis.Impact) > 0 {
		summary += fmt.Sprintf("\n受影响的文件:\n")
		for _, impact := range output.FileAnalysis.Impact {
			summary += fmt.Sprintf("  - %s (层级 %d)\n", impact.Path, impact.ImpactLevel)
		}
	}

	return summary
}
