package countas

import (
	"fmt"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

// CountAsResult 是 'as' 类型断言分析的最终结果的顶层结构体。
// 它实现了 projectanalyzer.Result 接口，提供完整的分析结果数据。
//
// 设计理念：
// 该结果结构体采用了聚合模式，既包含总体统计数据（总数、文件数），
// 也包含详细的文件级别信息，支持不同层次的分析需求。
//
// 支持的输出格式：
// - 控制台：人类可读的格式化报告，包含文件路径和代码片段
// - JSON：机器可读的结构化数据，便于集成到 CI/CD 系统
type CountAsResult struct {
	FilesParsed  int         `json:"filesParsed"`  // 成功解析的 JS/TS 文件数量
	TotalAsCount int         `json:"totalAsCount"` // 项目中 'as' 类型断言的总数
	FileCounts   []FileCount `json:"fileCounts"`   // 每个文件的 'as' 类型断言统计列表
}

// FileCount 存储单个文件中 'as' 类型断言的使用情况统计。
//
// 设计用途：
// 该结构体支持文件级别的详细分析，用于：
// - 识别类型断言使用集中的文件
// - 提供具体的代码片段和位置信息
// - 支持按文件分类的优化工作
// - 追踪类型断言的历史变化
//
// 数据完整性：
// - FilePath 提供文件定位信息
// - AsCount 提供快速的量化统计
// - Details 提供完整的代码上下文信息
type FileCount struct {
	FilePath string                `json:"filePath"` // 文件的绝对路径
	AsCount  int                   `json:"asCount"`  // 该文件中的 'as' 类型断言总数
	Details  []parser.AsExpression `json:"details"`  // 该文件中所有 'as' 类型断言的详细信息列表
}

// 确保 CountAsResult 结构体实现了 projectanalyzer.Result 接口。
// 这是一个编译时检查，确保结构体满足接口要求。
var _ projectanalyzer.Result = (*CountAsResult)(nil)

// Name 返回该结果对应的分析器的名称。
//
// 返回值说明：
// 返回 "Count As Usage" 作为分析器的标识符。
// 这个名称用于在结果输出中标识该分析器的分析结果。
func (r *CountAsResult) Name() string {
	return "Count As Usage"
}

// Summary 返回对结果的简短、人类可读的摘要。
//
// 设计目的：
// 该方法提供快速概览信息，适用于：
// - 控制台输出的首行摘要
// - 日志记录和监控
// - CI/CD 状态报告
// - 快速判断分析结果状态
//
// 输出格式：
// 采用中文友好的格式，清晰显示扫描文件数量和发现的类型断言总数。
//
// 返回值说明：
// 返回包含关键统计信息的字符串，格式为："扫描文件 X 个，共发现 Y 处 'as' 类型断言使用。"
func (r *CountAsResult) Summary() string {
	return fmt.Sprintf(
		"扫描文件 %d 个，共发现 %d 处 'as' 类型断言使用。",
		r.FilesParsed,
		r.TotalAsCount,
	)
}

// ToJSON 将结果的完整数据序列化为 JSON 格式。
//
// 功能说明：
// 该方法将完整的分析结果序列化为 JSON 格式，支持：
// - 机器可读的结构化数据输出
// - 集成到 CI/CD 流程
// - 数据库存储和后续分析
// - 自动化报告生成
//
// 参数说明：
// - indent: 控制是否格式化 JSON 输出，true 为格式化输出，false 为紧凑输出
//
// 返回值说明：
// - []byte: JSON 格式的结果数据
// - error: 序列化过程中可能出现的错误
func (r *CountAsResult) ToJSON(indent bool) ([]byte, error) {
	return projectanalyzer.ToJSONBytes(r, indent)
}

// ToConsole 将结果格式化为适合在控制台（终端）中打印的字符串。
//
// 输出设计：
// 采用清晰的层次结构，使用表情符号和分隔线增强可读性：
// - ✅ 无类型断言：显示成功消息
// - ⚠️ 发现类型断言：显示警告消息和详细列表
// - 分隔线：分隔摘要内容和详细列表
//
// 详细信息展示：
// 对于每个包含类型断言的文件，显示：
// - 文件路径和断言数量
// - 每个断言的行号和代码片段
//
// 使用场景：
// - 命令行工具的直接输出
// - 开发者快速定位问题
// - 代码审查会议的讨论材料
// - 构建过程的即时反馈
func (r *CountAsResult) ToConsole() string {
	// 处理没有类型断言的情况
	if r.TotalAsCount == 0 {
		return fmt.Sprintf("✅ %s 太棒了，项目中没有发现 'as' 类型断言！", r.Summary())
	}

	// 构建包含详细信息的输出
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("⚠️ %s\n", r.Summary()))
	builder.WriteString("--------------------------------------------------\n")

	// 遍历每个文件，输出详细信息
	for _, fc := range r.FileCounts {
		builder.WriteString(fmt.Sprintf("  - %s (%d 处):\n", fc.FilePath, fc.AsCount))
		for _, detail := range fc.Details {
			builder.WriteString(fmt.Sprintf("    - Line %d: %s\n", detail.SourceLocation.Start.Line, detail.Raw))
		}
	}
	builder.WriteString("--------------------------------------------------\n")

	return builder.String()
}

// AnalyzerName 返回对应的分析器名称
func (r *CountAsResult) AnalyzerName() string {
	return "count-as"
}
