// Package project_analyzer 定义了分析器插件系统的核心接口和类型。
// 它作为所有具体分析器模块的统一入口和契约，实现了"解析一次，分析多次"的设计理念。
//
// 核心设计原则：
// 1. 分离关注点：项目解析与代码分析完全分离
// 2. 可扩展性：新的分析器可以轻松添加而无需修改核心逻辑
// 3. 性能优化：避免重复解析，所有分析器共享同一个解析结果
// 4. 统一接口：所有分析器都遵循相同的接口规范
package project_analyzer

import (
	"encoding/json"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
)

// =============================================================================
// 核心接口定义
// =============================================================================

// Analyzer 是所有分析器模块都必须实现的接口。
// 这个接口定义了分析器的标准生命周期和行为规范。
//
// 接口方法说明：
// - Name(): 返回分析器的唯一标识符，用于注册和识别
// - Configure(): 在分析前进行参数配置和初始化
// - Analyze(): 执行具体的分析逻辑并返回结果
type Analyzer interface {
	// Name 返回分析器的唯一标识符，用于在插件系统中注册和识别该分析器。
	// 返回的名称应当简短、具有描述性，并且在整个系统中是唯一的。
	// 例如："unconsumed"、"count-any"、"npm-check" 等。
	Name() string

	// Configure 在分析开始前对分析器进行参数配置。
	// 这个方法会在 Analyze() 方法之前被调用，用于设置分析器的运行参数。
	//
	// 参数说明：
	// - params: 包含分析器所需的配置参数，格式为 map[string]string
	//   例如：{"targetFiles": "/path/to/file1.ts,/path/to/file2.ts"}
	//
	// 返回值说明：
	// - error: 如果配置参数无效，返回相应的错误信息
	Configure(params map[string]string) error

	// Analyze 执行具体的分析逻辑并返回分析结果。
	// 这是分析器的核心方法，包含了所有分析算法的实现。
	//
	// 参数说明：
	// - ctx: ProjectContext 包含了执行分析所需的完整项目上下文信息
	//   包括：项目根目录、排除规则、是否为monorepo、以及完整的项目解析结果
	//
	// 返回值说明：
	// - Result: 包含分析结果的对象，实现了标准的结果接口
	// - error: 分析过程中出现的错误，如数据不一致、算法异常等
	Analyze(ctx *ProjectContext) (Result, error)
}

// Result 是所有分析结果都必须实现的接口。
// 这个接口定义了分析结果的标准格式和输出方式。
//
// 接口方法说明：
// - Name(): 返回结果的名称，通常与分析器名称一致
// - Summary(): 提供分析结果的文本摘要，便于快速了解分析结果
// - ToJSON(): 将结果序列化为JSON格式，支持缩进格式化
// - ToConsole(): 将结果格式化为适合控制台显示的字符串
type Result interface {
	// Name 返回分析结果的名称，通常与分析器名称一致。
	// 这个名称用于标识结果数据的类型和来源。
	Name() string

	// Summary 返回分析结果的文本摘要。
	// 提供一个人类可读的简短描述，便于快速了解分析结果的主要内容。
	// 例如："找到 15 个未使用的导出符号"、"发现 3 个孤岛文件"等。
	Summary() string

	// ToJSON 将分析结果序列化为JSON格式。
	// 支持格式化输出，便于进一步处理或存储。
	//
	// 参数说明：
	// - indent: 是否格式化输出（使用缩进和换行）
	//
	// 返回值说明：
	// - []byte: JSON格式的结果数据
	// - error: 序列化过程中出现的错误
	ToJSON(indent bool) ([]byte, error)

	// ToConsole 将分析结果格式化为适合控制台显示的字符串。
	// 这个方法应当提供清晰、易读的格式，包含关键信息和高亮显示。
	ToConsole() string
}

// =============================================================================
// 共享类型定义
// =============================================================================

// ProjectContext 包含了执行一次分析所需的所有项目上下文信息。
// 这个结构体作为所有分析器的统一数据源，确保所有分析器在相同的上下文中工作。
type ProjectContext struct {
	// ProjectRoot 项目的根目录路径，绝对路径形式。
	// 所有文件路径都基于此路径进行解析和处理。
	ProjectRoot string

	// Exclude 需要从分析中排除的文件或目录的 glob 模式列表。
	// 支持多个模式，例如："node_modules/**", "**/*.test.ts", "dist/**"
	Exclude []string

	// IsMonorepo 指示当前项目是否为 monorepo 结构。
	// 如果是 monorepo，会采用特殊的解析策略来处理包间的依赖关系。
	IsMonorepo bool

	// ParsingResult 项目解析结果，包含完整的AST和项目信息。
	// 这个数据由 projectParser 模块生成，包含了所有 TypeScript/TSX 文件的解析结果。
	// 这是所有分析器的核心数据源。
	ParsingResult *projectParser.ProjectParserResult
}

// =============================================================================
// 辅助函数
// =============================================================================

// ToJSONBytes 是一个辅助函数，用于简化各种 Result 类型对 ToJSON 方法的实现。
// 提供了标准的JSON序列化功能，支持格式化输出。
//
// 参数说明：
// - v: 需要序列化的任意类型数据
// - indent: 是否使用缩进格式化输出
//
// 返回值说明：
// - []byte: JSON格式的字节数组
// - error: 序列化过程中出现的错误
func ToJSONBytes(v interface{}, indent bool) ([]byte, error) {
	if indent {
		return json.MarshalIndent(v, "", "  ")
	}
	return json.Marshal(v)
}
