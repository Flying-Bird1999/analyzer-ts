// Package component_deps 实现了基于配置文件的组件依赖分析器。
//
// 核心特性：
// 1. 配置驱动：通过 component-manifest.json 显式声明组件
// 2. 精确过滤：过滤掉组件内部依赖，只保留外部依赖
// 3. 完整信息：保留原始 import 解析结果，包含导入的详细内容
package component_deps

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
	Name         string                                  `json:"name"`         // 组件名称
	Path         string                                  `json:"path"`         // 组件目录路径
	Dependencies []projectParser.ImportDeclarationResult `json:"dependencies"` // 外部依赖列表（原始扫描数据，保持不变）

	// NpmDeps 本组件依赖的 npm 包列表（去重）
	// 例如: ["react", "lodash", "dayjs"]
	NpmDeps []string `json:"npmDeps,omitempty"`

	// ComponentDeps 本组件依赖的其他组件列表
	// 例如: Button 组件依赖 Input 组件的多个文件
	ComponentDeps []ComponentDep `json:"componentDeps,omitempty"`
}

// ComponentDep 组件依赖信息
// 表示当前组件依赖了某个其他组件的具体情况
type ComponentDep struct {
	// Name 被依赖的组件名称（来自 manifest）
	Name string `json:"name"`

	// Path 被依赖组件在 manifest 中声明的路径
	Path string `json:"path"`

	// DepFiles 具体依赖的文件路径列表
	// 表示当前组件中哪些文件引用了目标组件的文件
	DepFiles []string `json:"depFiles,omitempty"`
}

// Meta 分析元数据
type Meta struct {
	ComponentCount int `json:"componentCount"` // 组件总数
}

// ComponentDepsResult 组件依赖分析结果
type ComponentDepsResult struct {
	Meta       Meta                     `json:"meta"`
	Components map[string]ComponentInfo `json:"components"`
}

// =============================================================================
// Result 接口实现
// =============================================================================

// Name 返回分析结果标识符
func (r *ComponentDepsResult) Name() string {
	return "component-deps"
}

// Summary 返回分析结果摘要
func (r *ComponentDepsResult) Summary() string {
	totalDeps := 0
	for _, comp := range r.Components {
		totalDeps += len(comp.Dependencies)
	}
	return fmt.Sprintf("分析完成，共发现 %d 个组件，%d 条外部依赖。",
		r.Meta.ComponentCount, totalDeps)
}

// ToJSON 将结果序列化为 JSON
func (r *ComponentDepsResult) ToJSON(indent bool) ([]byte, error) {
	return projectanalyzer.ToJSONBytes(r, indent)
}

// ToConsole 将结果格式化为控制台输出
func (r *ComponentDepsResult) ToConsole() string {
	var buffer bytes.Buffer

	// 标题
	buffer.WriteString("=====================================\n")
	buffer.WriteString("组件依赖分析报告\n")
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

		// 显示 npm 依赖
		if len(comp.NpmDeps) > 0 {
			buffer.WriteString("  NPM 依赖:\n")
			for _, pkg := range comp.NpmDeps {
				buffer.WriteString(fmt.Sprintf("    - %s\n", pkg))
			}
		}

		// 显示组件依赖
		if len(comp.ComponentDeps) > 0 {
			buffer.WriteString("  组件依赖:\n")
			for _, dep := range comp.ComponentDeps {
				buffer.WriteString(fmt.Sprintf("    - %s (%s)\n", dep.Name, dep.Path))
				if len(dep.DepFiles) > 0 {
					for _, file := range dep.DepFiles {
						buffer.WriteString(fmt.Sprintf("      → %s\n", file))
					}
				}
			}
		}

		// 显示完整依赖列表
		if len(comp.Dependencies) > 0 {
			buffer.WriteString("  完整依赖列表:\n")
			for _, dep := range comp.Dependencies {
				if dep.Source.Type == "npm" {
					buffer.WriteString(fmt.Sprintf("    - npm: %s\n", dep.Source.NpmPkg))
				} else {
					buffer.WriteString(fmt.Sprintf("    - file: %s\n", dep.Source.FilePath))
				}
			}
		} else {
			buffer.WriteString("  外部依赖: 无\n")
		}
		buffer.WriteString("\n")
	}

	return buffer.String()
}

// AnalyzerName 返回对应的分析器名称
func (r *ComponentDepsResult) AnalyzerName() string {
	return "component-deps"
}
