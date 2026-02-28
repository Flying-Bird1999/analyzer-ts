// Package cmd 提供 MR 组件影响分析命令
package cmd

// MrComponentImpactCmd MR 组件影响分析命令
//
// 这是一个专门用于 Merge Request 场景的组件影响分析工具。
// 它基于 git diff 变更，分析代码变更对组件库的影响范围。
//
// 快速开始：
//   ./analyzer-ts mr-component-impact --project-root $(pwd) --diff-file changes.diff
//
// 常用示例：
//   1. 分析 git diff 文件：
//      ./analyzer-ts mr-component-impact --project-root $(pwd) --diff-file changes.diff
//
//   2. 使用 git diff 输出：
//      git diff main...HEAD > changes.diff
//      ./analyzer-ts mr-component-impact --project-root $(pwd) --diff-file changes.diff
//
//   3. 指定 manifest 路径：
//      ./analyzer-ts mr-component-impact --project-root $(pwd) --diff-file changes.diff --manifest .analyzer/component-manifest.json
//
//   4. 输出为 JSON：
//      ./analyzer-ts mr-component-impact --project-root $(pwd) --diff-file changes.diff --output result.json
//
//   5. 排除特定文件：
//      ./analyzer-ts mr-component-impact --project-root $(pwd) --diff-file changes.diff --exclude "**/tests/**" --exclude "**/*.test.ts"
//
// 参数说明：
//   必需：
//     --project-root <path>    项目根目录（绝对路径）
//     --diff-file <path>        diff 文件路径
//   可选：
//     --manifest <path>         组件清单路径（默认 .analyzer/component-manifest.json）
//     --output <path>           输出文件路径（默认 stdout，控制台输出）
//     --exclude <pattern>       排除 glob 模式（可多次使用）
//     --format json|console     输出格式（默认 console）
//
// 输出格式：
//   console - 人类可读的控制台格式（默认）
//   json    - JSON 格式，用于程序解析
//
// 工作原理：
//   1. 解析 git diff 文件，提取变更文件列表
//   2. 根据 manifest 配置，将文件分类为：
//      - component: 组件文件
//      - functions: 函数/工具文件
//      - other: 其他文件
//   3. 对于组件变更：
//      - 查询 component_deps 的依赖关系
//      - 找出所有依赖该组件的其他组件
//   4. 对于函数变更：
//      - 查询 export_call 的引用关系
//      - 找出所有引用该函数的组件
//   5. 输出完整的影响分析报告
//
// 注意：
//   - 组件必须在 manifest.json 中声明
//   - 函数影响分析基于 export_call 的 RefComponents 字段
//   - 未在 manifest 中声明的组件不会被追踪

import (
	"fmt"
	"os"
	"path/filepath"

	mrcomponentimpact "github.com/Flying-Bird1999/analyzer-ts/pkg/mr_component_impact"
	"github.com/spf13/cobra"
)

// =============================================================================
// 命令配置变量
// =============================================================================

var (
	// 必需参数
	mrProjectRoot string // 项目根目录
	mrDiffFile    string // diff 文件路径

	// 可选参数
	mrManifestPath string   // manifest 路径
	mrOutputFile   string   // 输出文件路径
	mrOutputFormat string   // 输出格式
	mrExcludePaths []string // 排除路径
)

// =============================================================================
// 命令定义
// =============================================================================

// MrComponentImpactCmd MR 组件影响分析命令
var MrComponentImpactCmd = &cobra.Command{
	Use:   "mr-component-impact",
	Short: "分析 MR 变更对组件的影响范围",
	Long: `MR 组件影响分析命令 - 专门用于 Merge Request 场景的组件级影响分析

此命令基于 git diff 变更，分析代码变更对组件库的影响范围。
它会识别直接变更的组件和函数，以及间接受影响的所有组件。

工作流程：
  1. 解析 diff 文件，提取变更文件列表
  2. 将文件分类为 component/functions/other
  3. 分析组件变更的影响（基于组件依赖关系）
  4. 分析函数变更的影响（基于函数引用关系）
  5. 生成完整的影响分析报告

示例：
  # 分析 diff 文件
  ./analyzer-ts mr-component-impact --project-root $(pwd) --diff-file changes.diff

  # 使用 git diff
  git diff main...HEAD > changes.diff
  ./analyzer-ts mr-component-impact --project-root $(pwd) --diff-file changes.diff

  # 输出为 JSON
  ./analyzer-ts mr-component-impact --project-root $(pwd) --diff-file changes.diff --format json --output result.json

  # 排除测试文件
  ./analyzer-ts mr-component-impact --project-root $(pwd) --diff-file changes.diff --exclude "**/tests/**" --exclude "**/*.test.ts"
`,
	RunE: runMrComponentImpact,
}

// =============================================================================
// 初始化
// =============================================================================

func init() {
	// 必需参数
	MrComponentImpactCmd.Flags().StringVar(&mrProjectRoot, "project-root", "", "项目根目录（必需，绝对路径）")
	MrComponentImpactCmd.Flags().StringVar(&mrDiffFile, "diff-file", "", "diff 文件路径（必需）")
	MrComponentImpactCmd.MarkFlagRequired("project-root")
	MrComponentImpactCmd.MarkFlagRequired("diff-file")

	// 可选参数
	MrComponentImpactCmd.Flags().StringVar(&mrManifestPath, "manifest", "", "组件清单路径（默认 .analyzer/component-manifest.json）")
	MrComponentImpactCmd.Flags().StringVarP(&mrOutputFile, "output", "o", "", "输出文件路径（默认控制台输出）")
	MrComponentImpactCmd.Flags().StringVarP(&mrOutputFormat, "format", "f", "console", "输出格式：console 或 json（默认 console）")
	MrComponentImpactCmd.Flags().StringSliceVarP(&mrExcludePaths, "exclude", "x", []string{}, "排除 glob 模式（可多次使用）")
}

// =============================================================================
// 命令执行逻辑
// =============================================================================

// runMrComponentImpact 执行 MR 组件影响分析
func runMrComponentImpact(cmd *cobra.Command, args []string) error {
	// 1. 验证输入参数
	if err := validateInput(); err != nil {
		return fmt.Errorf("参数验证失败: %w", err)
	}

	// 2. 准备配置
	config := &mrcomponentimpact.AnalyzeConfig{
		ProjectRoot:  mrProjectRoot,
		DiffFilePath: mrDiffFile,
		ManifestPath: mrManifestPath,
		ExcludePaths: mrExcludePaths,
	}

	// 3. 执行分析
	fmt.Fprintf(os.Stderr, "🔍 开始分析 MR 组件影响...\n")
	fmt.Fprintf(os.Stderr, "   项目根目录: %s\n", mrProjectRoot)
	fmt.Fprintf(os.Stderr, "   Diff 文件: %s\n\n", mrDiffFile)

	result, err := mrcomponentimpact.AnalyzeFromDiff(config)
	if err != nil {
		return fmt.Errorf("分析失败: %w", err)
	}

	// 4. 输出结果
	if err := outputResult(result); err != nil {
		return fmt.Errorf("输出结果失败: %w", err)
	}

	return nil
}

// =============================================================================
// 辅助函数
// =============================================================================

// validateInput 验证输入参数
func validateInput() error {
	// 检查 project-root
	if mrProjectRoot == "" {
		return fmt.Errorf("--project-root 参数不能为空")
	}
	if !filepath.IsAbs(mrProjectRoot) {
		return fmt.Errorf("--project-root 必须是绝对路径")
	}
	if _, err := os.Stat(mrProjectRoot); os.IsNotExist(err) {
		return fmt.Errorf("项目根目录不存在: %s", mrProjectRoot)
	}

	// 检查 diff-file
	if mrDiffFile == "" {
		return fmt.Errorf("--diff-file 参数不能为空")
	}
	if !filepath.IsAbs(mrDiffFile) {
		// 转换为绝对路径
		mrDiffFile = filepath.Join(mrProjectRoot, mrDiffFile)
	}
	if _, err := os.Stat(mrDiffFile); os.IsNotExist(err) {
		return fmt.Errorf("diff 文件不存在: %s", mrDiffFile)
	}

	// 设置默认 manifest 路径
	if mrManifestPath == "" {
		mrManifestPath = filepath.Join(mrProjectRoot, ".analyzer", "component-manifest.json")
	}
	if !filepath.IsAbs(mrManifestPath) {
		mrManifestPath = filepath.Join(mrProjectRoot, mrManifestPath)
	}
	// 检查 manifest 是否存在
	if _, err := os.Stat(mrManifestPath); os.IsNotExist(err) {
		return fmt.Errorf("manifest 文件不存在: %s", mrManifestPath)
	}

	// 验证输出格式
	if mrOutputFormat != "console" && mrOutputFormat != "json" {
		return fmt.Errorf("无效的输出格式: %s（必须是 console 或 json）", mrOutputFormat)
	}

	return nil
}

// outputResult 输出分析结果
func outputResult(result *mrcomponentimpact.AnalysisResult) error {
	var output string
	var err error

	switch mrOutputFormat {
	case "json":
		output, err = result.ToJSON()
		if err != nil {
			return fmt.Errorf("生成 JSON 失败: %w", err)
		}
	case "console":
		output = result.ToConsole()
	}

	// 输出到文件或控制台
	if mrOutputFile != "" {
		if err := os.WriteFile(mrOutputFile, []byte(output), 0644); err != nil {
			return fmt.Errorf("写入输出文件失败: %w", err)
		}
		fmt.Fprintf(os.Stderr, "\n✅ 结果已保存到: %s\n", mrOutputFile)
		fmt.Fprintf(os.Stderr, "   %s\n", result.GetSummary())
	} else {
		// 控制台输出直接输出到 stdout
		fmt.Print(output)
	}

	return nil
}
