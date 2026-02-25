package trace

import (
	"fmt"

	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

// =============================================================================
// 结果定义
// =============================================================================

// TraceResult 封装了链路追踪的分析结果，并实现了 projectanalyzer.Result 接口。
//
// 设计理念：
// 该结果结构体采用扁平化的数据结构，直接使用 map[string]interface{}
// 来存储分析数据，这样可以更好地适应复杂的树状数据结构。
//
// 数据组织：
// - 按 file path 分组：每个文件作为一个独立的单元
// - 嵌套结构：每个文件内包含相关的 imports、variables、jsx、calls 等节点
// - 过滤后数据：只包含与目标包相关的代码节点，减少噪音
//
// 输出格式：
// - JSON：适用于机器处理和集成到其他系统
// - 控制台：格式化的 JSON 输出，便于人工查看
//
// 接口实现：
// 完整实现 projectanalyzer.Result 接口的所有方法，确保与框架的兼容性。
type TraceResult struct {
	// Data 存储了最终过滤后的分析数据，其结构是一个以文件路径为键，
	// 以包含该文件内相关代码节点（如imports, jsx等）的map为值的嵌套map。
	// 这种设计支持灵活的数据结构，同时保持清晰的层级关系。
	Data map[string]interface{}
}

// Name 返回结果的名称，与分析器名称一致。
//
// 返回值说明：
// 返回 "trace" 作为结果对象的标识符。
// 这个名称用于在输出中标识该分析器生成的结果。
func (r *TraceResult) Name() string {
	return "trace"
}

// Summary 返回对结果的简短描述。
//
// 设计目的：
// 该方法提供快速概览信息，显示受影响文件的数量，
// 用于评估目标包在项目中的影响范围。
//
// 输出格式：
// 采用中文友好的格式，清晰显示包含相关使用链路的文件数量。
//
// 返回值说明：
// 返回包含关键统计信息的字符串，格式为："成功追踪到 X 个文件中存在相关的使用链路。"
func (r *TraceResult) Summary() string {
	return fmt.Sprintf("成功追踪到 %d 个文件中存在相关的使用链路。", len(r.Data))
}

// ToJSON 将结果序列化为 JSON 格式的字节数组。
//
// 设计决策：
// 直接序列化核心的 Data 字段，而不是整个 TraceResult 结构体，
// 这样可以输出更纯净的 JSON 结果，不包含包装层的元数据。
//
// 功能说明：
// 该方法将完整的分析结果序列化为 JSON 格式，支持：
// - 机器可读的结构化数据输出
// - 集成到 CI/CD 流程
// - 后续数据处理和分析
// - 自动化报告生成
//
// 参数说明：
// - indent: 控制是否格式化 JSON 输出，true 为格式化输出，false 为紧凑输出
//
// 返回值说明：
// - []byte: JSON 格式的结果数据
// - error: 序列化过程中可能出现的错误
func (r *TraceResult) ToJSON(indent bool) ([]byte, error) {
	// 直接将核心的Data字段进行序列化，而不是整个TraceResult结构体，
	// 以便输出更纯净的JSON结果。
	return projectanalyzer.ToJSONBytes(r.Data, indent)
}

// ToConsole 将结果转换为适合在控制台输出的字符串格式。
//
// 输出策略：
// 对于 trace 这种复杂的树状结果，直接输出格式化的 JSON 是最清晰的，
// 因为：
// - 树状结构天然适合 JSON 的层级表示
// - 便于开发者复制和进一步处理
// - 保持了数据的完整性和准确性
//
// 错误处理：
// 如果 JSON 序列化失败，返回错误信息而不是崩溃，
// 确保工具的稳定性和用户体验。
//
// 使用场景：
// - 命令行工具的直接输出
// - 开发者快速查看分析结果
// - 调试和验证分析逻辑
func (r *TraceResult) ToConsole() string {
	// 对于 trace 这种复杂的树状结果，直接输出格式化的 JSON 是最清晰的，所以我们复用 ToJSON。
	jsonData, err := r.ToJSON(true)
	if err != nil {
		return fmt.Sprintf("无法将结果序列化为JSON: %v", err)
	}
	return string(jsonData)
}

// AnalyzerName 返回对应的分析器名称
func (r *TraceResult) AnalyzerName() string {
	return "trace"
}
