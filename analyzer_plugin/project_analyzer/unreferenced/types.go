// Package unreferenced 包含了查找项目中未使用文件的功能所需的所有类型定义。
//
// 该包定义了未引用文件分析器的完整类型系统，用于：
// - 存储分析配置信息和参数设置
// - 提供分析结果的统计数据和分类信息
// - 支持结果序列化和输出格式化
// - 便于后续的数据处理和集成
//
// 设计理念：
// 类型定义采用结构化的设计，既包含量化的统计数据，
// 也包含配置信息，便于分析结果的追溯和验证。
//
// 应用场景：
// - 代码清理：识别可以安全删除的文件
// - 架构分析：了解项目的文件依赖结构
// - 性能优化：减少编译和打包的文件数量
// - 维护性改进：简化项目结构，提高可维护性
package unreferenced

// SummaryStats 提供了关于未引用文件分析的统计数据。
//
// 设计用途：
// 该结构体提供分析结果的量化统计，用于快速了解项目的整体状况。
// 每个字段都对应分析过程中的关键指标，便于生成报告和监控。
//
// 字段说明：
// - TotalFiles: 项目中分析的总文件数量
// - ReferencedFiles: 被其他文件引用的文件数量
// - TrulyUnreferencedFiles: 真正未引用的文件数量（可以安全删除）
// - SuspiciousFiles: 可疑文件数量（需要人工确认）
//
// 计算方式：
// - TotalFiles: 所有解析的 TypeScript/TSX 文件数量
// - ReferencedFiles: 在其他文件中被引用的文件数量
// - TrulyUnreferencedFiles: 智能分类后的真正未引用文件数量
// - SuspiciousFiles: 智能分类后的可疑文件数量
type SummaryStats struct {
	TotalFiles             int `json:"totalFiles"`              // 项目中分析的总文件数量
	ReferencedFiles        int `json:"referencedFiles"`         // 被其他文件引用的文件数量
	TrulyUnreferencedFiles int `json:"trulyUnreferencedFiles"` // 真正未引用的文件数量（可以安全删除）
	SuspiciousFiles        int `json:"suspiciousFiles"`        // 可疑文件数量（需要人工确认）
}

// AnalysisConfiguration 记录了用于本次分析的输入配置，便于追溯结果。
//
// 设计用途：
// 该结构体存储分析器的配置信息，确保分析结果的可追溯性和可重现性。
// 当需要验证分析结果或复现分析过程时，这些配置信息非常有价值。
//
// 配置字段：
// - InputDir: 分析的输入目录路径
// - EntrypointsSpecified: 是否指定了自定义入口文件
// - IncludeEntryDirs: 是否包含常见的入口目录模式
//
// 使用场景：
// - 结果验证：确认分析时使用的配置参数
// - 报告生成：在报告中包含分析配置信息
// - 问题排查：诊断分析结果异常时的参考依据
// - 性能优化：不同配置下分析结果的对比
type AnalysisConfiguration struct {
	InputDir             string `json:"inputDir"`             // 分析的输入目录路径
	EntrypointsSpecified bool   `json:"entrypointsSpecified"` // 是否指定了自定义入口文件
	IncludeEntryDirs     bool   `json:"includeEntryDirs"`     // 是否包含常见的入口目录模式
}
