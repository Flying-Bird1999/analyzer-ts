// Package component_deps_v2 实现了基于配置文件的组件依赖分析器（V2版本）。
//
// 与 component-deps 的区别：
// - component-deps: 从入口文件自动识别组件
// - component-deps-v2: 基于配置文件显式声明组件
//
// 核心特性：
// 1. 配置驱动：通过 component-manifest.json 显式声明组件
// 2. 精确作用域：每个组件可以定义自己的文件作用域
// 3. 完整依赖图：生成正向和反向依赖关系
// 4. 支持大型项目：适用于复杂的组件库项目
package component_deps_v2

import (
	"bytes"
	"fmt"
	"sort"

	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

// =============================================================================
// 数据结构定义
// =============================================================================

// ComponentInfo 单个组件的依赖信息
type ComponentInfo struct {
	Name         string   `json:"name"`         // 组件名称
	Entry        string   `json:"entry"`        // 组件入口文件
	Dependencies []string `json:"dependencies"` // 该组件依赖的其他组件
}

// DependencyGraph 正向依赖图
// key: 组件名称, value: 该组件直接依赖的组件列表
type DependencyGraph map[string][]string

// ReverseDepGraph 反向依赖图
// key: 组件名称, value: 依赖该组件的其他组件列表
type ReverseDepGraph map[string][]string

// Meta 分析元数据
type Meta struct {
	Version    string `json:"version"`    // 配置文件版本
	LibraryName string `json:"libraryName"` // 组件库名称
	ComponentCount int  `json:"componentCount"` // 组件总数
}

// ComponentDepsV2Result 组件依赖分析结果
type ComponentDepsV2Result struct {
	Meta           Meta                      `json:"meta"`                       // 元数据
	Components     map[string]ComponentInfo  `json:"components"`                 // 组件信息
	DepGraph       DependencyGraph           `json:"depGraph"`                   // 正向依赖图
	RevDepGraph    ReverseDepGraph            `json:"revDepGraph"`                // 反向依赖图
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
	for _, deps := range r.DepGraph {
		totalDeps += len(deps)
	}
	return fmt.Sprintf("分析完成，共发现 %d 个组件，%d 条依赖关系。",
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
	buffer.WriteString(fmt.Sprintf("组件库: %s\n", r.Meta.LibraryName))
	buffer.WriteString(fmt.Sprintf("配置版本: %s\n", r.Meta.Version))
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
		buffer.WriteString(fmt.Sprintf("  入口: %s\n", comp.Entry))
		if len(comp.Dependencies) > 0 {
			buffer.WriteString("  依赖:\n")
			for _, dep := range comp.Dependencies {
				buffer.WriteString(fmt.Sprintf("    - %s\n", dep))
			}
		} else {
			buffer.WriteString("  依赖: 无\n")
		}
		buffer.WriteString("\n")
	}

	// 反向依赖（被依赖情况）
	buffer.WriteString("=====================================\n")
	buffer.WriteString("反向依赖（被谁依赖）\n")
	buffer.WriteString("=====================================\n\n")

	for _, name := range sortedNames {
		if revDeps, ok := r.RevDepGraph[name]; ok && len(revDeps) > 0 {
			buffer.WriteString(fmt.Sprintf("▶ %s 被 %d 个组件依赖:\n", name, len(revDeps)))
			for _, dep := range revDeps {
				buffer.WriteString(fmt.Sprintf("    - %s\n", dep))
			}
			buffer.WriteString("\n")
		}
	}

	return buffer.String()
}
