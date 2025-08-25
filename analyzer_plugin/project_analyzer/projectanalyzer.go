// package project_analyzer 定义了分析器插件系统的核心接口和类型。
// 它作为所有具体分析器模块的统一入口和契约。
package project_analyzer

import (
	"encoding/json"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
)

// --- 核心接口 ---

// Analyzer 是所有分析器模块都必须实现的接口。
type Analyzer interface {
	Name() string
	Configure(params map[string]string) error
	Analyze(ctx *ProjectContext) (Result, error)
}

// Result 是所有分析结果都必须实现的接口。
type Result interface {
	Name() string
	Summary() string
	ToJSON(indent bool) ([]byte, error)
	ToConsole() string
}

// --- 共享类型 ---

// ProjectContext 包含了执行一次分析所需的所有项目上下文信息。
type ProjectContext struct {
	ProjectRoot   string
	Exclude       []string
	IsMonorepo    bool
	ParsingResult *projectParser.ProjectParserResult
}

// --- 辅助函数 ---

// ToJSONBytes 是一个辅助函数，用于简化各种 Result 类型对 ToJSON 的实现。
func ToJSONBytes(v interface{}, indent bool) ([]byte, error) {
	if indent {
		return json.MarshalIndent(v, "", "  ")
	}
	return json.Marshal(v)
}
