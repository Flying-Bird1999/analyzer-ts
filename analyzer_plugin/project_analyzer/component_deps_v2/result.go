// Package component_deps_v2 实现了基于配置文件的组件依赖分析器（V2版本）。
//
// 核心特性：
// 1. 配置驱动：通过 component-manifest.json 显式声明组件
// 2. 精确过滤：过滤掉组件内部依赖，只保留外部依赖
// 3. 完整信息：保留原始 import 解析结果，包含导入的详细内容
package component_deps_v2

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

// =============================================================================
// 数据结构定义
// =============================================================================

// ComponentInfo 单个组件的依赖信息
type ComponentInfo struct {
	Name         string                              `json:"name"`         // 组件名称
	Path         string                              `json:"path"`         // 组件目录路径
	Dependencies []projectParser.ImportDeclarationResult `json:"dependencies"` // 外部依赖列表
}

// Meta 分析元数据
type Meta struct {
	ComponentCount int `json:"componentCount"` // 组件总数
}

// ComponentDepsV2Result 组件依赖分析结果
type ComponentDepsV2Result struct {
	Meta       Meta                        `json:"meta"`
	Components map[string]ComponentInfo `json:"components"`
}

// =============================================================================
// Result 接口实现
// =============================================================================

// Name 返回分析结果标识符
func (r *ComponentDepsV2Result) Name() string {
	return "component-deps-v2"
}

// Summary 返回分析结果摘要
func (r *ComponentDepsV2Result) Summary() string {
	totalDeps := 0
	for _, comp := range r.Components {
		totalDeps += len(comp.Dependencies)
	}
	return fmt.Sprintf("分析完成，共发现 %d 个组件，%d 条外部依赖。",
		r.Meta.ComponentCount, totalDeps)
}

// ToJSON 将结果序列化为 JSON
func (r *ComponentDepsV2Result) ToJSON(indent bool) ([]byte, error) {
	return projectanalyzer.ToJSONBytes(r, indent)
}

// ToConsole 将结果格式化为控制台输出
func (r *ComponentDepsV2Result) ToConsole() string {
	var buffer bytes.Buffer

	// 标题
	buffer.WriteString("=====================================\n")
	buffer.WriteString("组件依赖分析报告 (V2)\n")
	buffer.WriteString("=====================================\n\n")

	// 元数据
	buffer.WriteString(fmt.Sprintf("组件总数: %d\n\n", r.Meta.ComponentCount))

	// 按名称排序组件列表
	sortedNames := make([]string, 0, len(r.Components))
	for name := range r.Components {
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	// 组件详情
	for _, name := range sortedNames {
		comp := r.Components[name]
		buffer.WriteString(fmt.Sprintf("▶ %s\n", name))
		buffer.WriteString(fmt.Sprintf("  路径: %s\n", comp.Path))
		if len(comp.Dependencies) > 0 {
			buffer.WriteString("  外部依赖:\n")
			for _, dep := range comp.Dependencies {
				// 根据 type 显示不同信息
				if dep.Source.Type == "npm" {
					buffer.WriteString(fmt.Sprintf("    - npm: %s\n", dep.Source.NpmPkg))
				} else {
					targetFile := dep.Source.FilePath
					buffer.WriteString(fmt.Sprintf("    - file: %s\n", targetFile))
				}
			}
		} else {
			buffer.WriteString("  外部依赖: 无\n")
		}
		buffer.WriteString("\n")
	}

	return buffer.String()
}
