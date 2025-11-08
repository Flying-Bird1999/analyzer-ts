package countany

import (
	"fmt"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

// =============================================================================
// 结果数据结构定义
// =============================================================================

// CountAnyResult 是 'any' 类型分析的最终结果的顶层结构体。
// 这个结构体包含了整个项目的 'any' 类型统计信息，实现了 projectanalyzer.Result 接口。
//
// 数据结构设计：
// - FilesParsed: 项目中成功解析的 TypeScript/TSX 文件总数
// - TotalAnyCount: 整个项目中 'any' 类型的使用总数
// - FileCounts: 每个文件的详细统计信息列表，只包含有 'any' 类型使用的文件
//
// 这个结构体通过 JSON 标签支持序列化，便于输出和存储。
type CountAnyResult struct {
	FilesParsed   int         `json:"filesParsed"`   // 成功解析的 TypeScript/TSX 文件数量
	TotalAnyCount int         `json:"totalAnyCount"` // 项目中 'any' 类型的总数
	FileCounts    []FileCount `json:"fileCounts"`    // 每个文件的 'any' 类型统计列表
}

// FileCount 存储单个文件中 'any' 类型的使用情况统计。
// 这个结构体提供了文件级别的详细统计信息，便于按文件进行分类查看和处理。
//
// 数据字段说明：
// - FilePath: 文件的绝对路径，用于定位具体的文件位置
// - AnyCount: 该文件中 'any' 类型的总数，快速了解该文件的类型安全性
// - Details: 该文件中所有 'any' 类型的详细信息列表，包含位置和源码片段
type FileCount struct {
	FilePath string           `json:"filePath"` // 文件的绝对路径
	AnyCount int              `json:"anyCount"` // 该文件中的 'any' 类型总数
	Details  []parser.AnyInfo `json:"details"`  // 该文件中所有 'any' 类型的详细信息列表
}

// =============================================================================
// 接口实现
// =============================================================================

// 确保 CountAnyResult 结构体实现了 projectanalyzer.Result 接口。
// 这是一个编译时检查，确保结构体符合接口规范。
var _ projectanalyzer.Result = (*CountAnyResult)(nil)

// Name 返回该结果对应的分析器的名称。
// 返回一个描述性的名称，用于标识这个分析结果的来源和类型。
func (r *CountAnyResult) Name() string {
	return "Count Any Usage"
}

// Summary 返回对结果的简短、人类可读的摘要。
// 这个摘要提供了分析结果的快速概览，便于在报告或日志中显示。
//
// 格式说明：
// 显示扫描的文件总数和发现的 'any' 类型总数，例如：
// "扫描文件 150 个，共发现 23 处 'any' 类型使用。"
func (r *CountAnyResult) Summary() string {
	return fmt.Sprintf(
		"扫描文件 %d 个，共发现 %d 处 'any' 类型使用。",
		r.FilesParsed,
		r.TotalAnyCount,
	)
}

// ToJSON 将结果的完整数据序列化为 JSON 格式。
// 支持格式化输出，便于进一步处理或与其他系统集成。
//
// 参数说明：
// - indent: 是否格式化输出（使用缩进和换行），设置为 true 时便于人类阅读
//
// 返回值说明：
// - []byte: JSON 格式的结果数据
// - error: 序列化过程中出现的错误（通常不会出错）
func (r *CountAnyResult) ToJSON(indent bool) ([]byte, error) {
	return projectanalyzer.ToJSONBytes(r, indent)
}

// ToConsole 将结果格式化为适合在控制台（终端）中打印的字符串。
// 这个方法提供了清晰、易读的文本输出，包含视觉提示和详细的代码片段。
//
// 输出格式设计：
// 1. 当没有 'any' 类型时：显示成功消息 ✅
// 2. 当有 'any' 类型时：显示警告消息 ⚠️，并按文件分类列出详情
// 3. 每个文件显示文件路径和 'any' 类型数量
// 4. 每个 'any' 类型显示具体的行号和代码片段
// 5. 使用分隔线增强可读性
//
// 输出示例：
// ⚠️ 扫描文件 150 个，共发现 23 处 'any' 类型使用。
// --------------------------------------------------
//   - /path/to/file.ts (5 处):
//     - Line 42: const data: any = response;
//     - Line 58: function processData(input: any): void { ... }
//   - /path/to/another.ts (18 处):
//     - Line 12: let config: any;
// --------------------------------------------------
func (r *CountAnyResult) ToConsole() string {
	// 特殊情况处理：项目中没有 'any' 类型，显示成功消息
	if r.TotalAnyCount == 0 {
		return fmt.Sprintf("✅ %s 太棒了，项目中没有发现 'any' 类型！", r.Summary())
	}

	// 使用 strings.Builder 高效构建字符串输出
	var builder strings.Builder

	// 显示警告标题和摘要
	builder.WriteString(fmt.Sprintf("⚠️ %s\n", r.Summary()))
	builder.WriteString("--------------------------------------------------\n")

	// 遍历每个包含 'any' 类型的文件
	for _, fc := range r.FileCounts {
		// 显示文件路径和该文件中的 'any' 类型数量
		builder.WriteString(fmt.Sprintf("  - %s (%d 处):\n", fc.FilePath, fc.AnyCount))

		// 显示该文件中每个 'any' 类型的详细信息
		for _, detail := range fc.Details {
			builder.WriteString(fmt.Sprintf("    - Line %d: %s\n",
				detail.SourceLocation.Start.Line, detail.Raw))
		}
	}

	builder.WriteString("--------------------------------------------------\n")

	return builder.String()
}
