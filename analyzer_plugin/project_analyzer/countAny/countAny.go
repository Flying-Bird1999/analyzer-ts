// Package countany 包含了用于统计项目中 'any' 类型使用情况的核心业务逻辑。
//
// 功能说明：
// 这个分析器专门用于检测和统计 TypeScript 项目中所有 'any' 类型的使用情况。
// 'any' 类型是 TypeScript 中的一个特殊类型，它会绕过类型检查，可能导致运行时错误。
// 通过统计 'any' 类型的使用，可以帮助开发者：
// 1. 了解项目的类型安全性程度
// 2. 识别需要重构的代码区域
// 3. 追踪类型安全性改进的进展
// 4. 制定逐步消除 'any' 类型的计划
//
// 实现特点：
// - 无需配置参数，开箱即用
// - 提供详细的位置信息和原始代码片段
// - 支持按文件分类统计
// - 生成清晰的文本报告和结构化的JSON输出
package countany

import (
	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

// Counter 是"统计any"分析器的实现。
// 这个分析器会遍历项目中所有的 TypeScript/TSX 文件，统计其中 'any' 类型的使用次数。
//
// 工作原理：
// 1. 接收项目解析结果作为输入
// 2. 遍历每个文件的 ExtractedNodes.AnyDeclarations 数据
// 3. 统计每个文件中 'any' 类型的使用次数
// 4. 汇总所有文件的统计结果
// 5. 生成包含总数和详细信息的分析报告
type Counter struct{}

// 确保 Counter 实现了 projectanalyzer.Analyzer 接口
var _ projectanalyzer.Analyzer = (*Counter)(nil)

// Name 返回分析器的唯一标识符。
// 这个名称用于在命令行中调用该分析器：
// ./analyzer-ts analyze count-any -i /path/to/project
func (c *Counter) Name() string {
	return "count-any"
}

// Configure 配置分析器的运行参数。
// 由于 'count-any' 分析器不需要任何配置参数，这个方法直接返回 nil。
// 这种设计保证了接口的一致性，同时保持了简单性。
func (c *Counter) Configure(params map[string]string) error {
	// 该分析器不需要任何配置参数
	return nil
}

// Analyze 执行核心的分析逻辑。
// 这个方法会遍历项目中的所有文件，统计 'any' 类型的使用情况。
//
// 处理流程：
// 1. 从项目上下文中获取解析结果
// 2. 初始化统计数据结构
// 3. 遍历每个文件的解析数据：
//   - 获取该文件中 'any' 类型的声明数量
//   - 如果有 'any' 类型使用，记录详细信息
//   - 累加到总数统计中
//
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
	totalAnyCount := 0         // 项目中 'any' 类型的总数
	var fileCounts []FileCount // 每个文件的统计信息

	// 遍历所有已解析的文件
	// parseResult.Js_Data 是一个 map，key是文件路径，value是文件解析结果
	for filePath, fileData := range parseResult.Js_Data {
		// 获取当前文件中 'any' 类型的数量
		anyCountInFile := len(fileData.ExtractedNodes.AnyDeclarations)

		// 如果该文件中有 'any' 类型使用，记录详细信息
		if anyCountInFile > 0 {
			fileCounts = append(fileCounts, FileCount{
				FilePath: filePath,                                // 文件路径
				AnyCount: anyCountInFile,                          // 该文件中的 'any' 类型数量
				Details:  fileData.ExtractedNodes.AnyDeclarations, // 详细的位置和源码信息
			})
		}

		// 累加到总数统计
		totalAnyCount += anyCountInFile
	}

	// 构建最终的分析结果对象
	finalResult := &CountAnyResult{
		TotalAnyCount: totalAnyCount,            // 项目中 'any' 类型的总数
		FileCounts:    fileCounts,               // 每个文件的详细统计
		FilesParsed:   len(parseResult.Js_Data), // 总共解析的文件数量
	}

	return finalResult, nil
}

// init 在包加载时自动注册分析器
func init() {
	projectanalyzer.RegisterAnalyzer("count-any", func() projectanalyzer.Analyzer {
		return &Counter{}
	})
}
