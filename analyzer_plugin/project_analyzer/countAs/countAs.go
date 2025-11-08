// Package countas 包含了用于统计项目中 'as' 类型断言使用情况的核心业务逻辑。
//
// 功能说明：
// 这个分析器专门用于检测和统计 TypeScript 项目中所有 'as' 类型断言的使用情况。
// 类型断言（Type Assertion）是 TypeScript 中的一个特性，允许开发者覆盖类型推断，
// 告诉编译器"我知道这个值是什么类型"。虽然类型断言在某些情况下是必要的，
// 但过度使用可能表明类型定义不够完善，或者存在强制类型转换的代码异味。
//
// 主要用途：
// 1. 代码质量监控：跟踪类型断言的使用数量和分布
// 2. 类型安全改进：识别可能需要改进类型定义的区域
// 3. 重构指导：在重构过程中帮助减少不必要的类型断言
// 4. 团队规范：制定和执行类型断言使用规范
//
// 支持的语法类型：
// - 尖括号语法：`<Type>value`
// - as 关键字语法：`value as Type`
// - 非空断言：`value!`
// - const 类型断言：`value as const`
//
// 实现特点：
// - 无需配置参数，开箱即用
// - 提供详细的位置信息和代码片段
// - 支持按文件分类统计
// - 生成清晰的文本报告和结构化的JSON输出
package countas

import (
	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

// =============================================================================
// 分析器主体定义
// =============================================================================

// Counter 是"统计as"分析器的实现。
// 这个分析器会遍历项目中所有的 TypeScript/TSX 文件，统计其中 'as' 类型断言的使用次数。
//
// 工作原理：
// 1. 接收项目解析结果作为输入
// 2. 遍历每个文件的 ExtractedNodes.AsExpressions 数据
// 3. 统计每个文件中类型断言的使用次数
// 4. 汇总所有文件的统计结果
// 5. 生成包含总数和详细信息的分析报告
type Counter struct{}

// 确保 Counter 实现了 projectanalyzer.Analyzer 接口
var _ projectanalyzer.Analyzer = (*Counter)(nil)

// Name 返回分析器的唯一标识符。
// 这个名称用于在命令行中调用该分析器：
// ./analyzer-ts analyze count-as -i /path/to/project
func (c *Counter) Name() string {
	return "count-as"
}

// Configure 配置分析器的运行参数。
// 由于 'count-as' 分析器不需要任何配置参数，这个方法直接返回 nil。
// 这种设计保证了接口的一致性，同时保持了简单性。
func (c *Counter) Configure(params map[string]string) error {
	// 该分析器不需要任何配置参数
	return nil
}

// Analyze 执行核心的分析逻辑。
// 这个方法会遍历项目中的所有文件，统计 'as' 类型断言的使用情况。
//
// 处理流程：
// 1. 从项目上下文中获取解析结果
// 2. 初始化统计数据结构
// 3. 遍历每个文件的解析数据：
//    - 获取该文件中类型断言的数量
//    - 如果有类型断言使用，记录详细信息
//    - 累加到总数统计中
// 4. 生成最终的分析结果对象
//
// 参数说明：
// - ctx: 项目上下文，包含完整的解析结果
//
// 返回值说明：
// - projectanalyzer.Result: 包含分析统计结果的对象
// - error: 分析过程中出现的错误（通常不会出错）
func (c *Counter) Analyze(ctx *projectanalyzer.ProjectContext) (projectanalyzer.Result, error) {
	// 获取项目解析结果，包含了所有文件的AST数据
	parseResult := ctx.ParsingResult

	// 初始化统计数据
	totalAsCount := 0       // 项目中类型断言的总数
	var fileCounts []FileCount // 每个文件的统计信息

	// 遍历所有已解析的文件
	// parseResult.Js_Data 是一个 map，key是文件路径，value是文件解析结果
	for filePath, fileData := range parseResult.Js_Data {
		// 获取当前文件中类型断言的数量
		asCountInFile := len(fileData.ExtractedNodes.AsExpressions)

		// 如果该文件中有类型断言使用，记录详细信息
		if asCountInFile > 0 {
			fileCounts = append(fileCounts, FileCount{
				FilePath: filePath, // 文件路径
				AsCount:  asCountInFile, // 该文件中的类型断言数量
				Details:  fileData.ExtractedNodes.AsExpressions, // 详细的位置和源码信息
			})
		}

		// 累加到总数统计
		totalAsCount += asCountInFile
	}

	// 构建最终的分析结果对象
	finalResult := &CountAsResult{
		TotalAsCount: totalAsCount,    // 项目中类型断言的总数
		FileCounts:   fileCounts,       // 每个文件的详细统计
		FilesParsed:  len(parseResult.Js_Data), // 总共解析的文件数量
	}

	return finalResult, nil
}
