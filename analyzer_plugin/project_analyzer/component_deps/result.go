package component_deps

import (
	"bytes"
	"fmt"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

// ComponentInfo 包含了单个公共组件的详细分析信息
type ComponentInfo struct {
	SourcePath   string   `json:"sourcePath"`
	Dependencies []string `json:"dependencies"`
}

// Result 保存了对组件库的完整分析结果，以 package 分组
type Result struct {
	Packages map[string]map[string]ComponentInfo `json:"packages"`
}

func (r *Result) Name() string {
	return "component-deps"
}

func (r *Result) Summary() string {
	packageCount := len(r.Packages)
	componentCount := 0
	for _, components := range r.Packages {
		componentCount += len(components)
	}
	return fmt.Sprintf("分析完成，共找到 %d 个包中的 %d 个公共组件。", packageCount, componentCount)
}

func (r *Result) ToJSON(indent bool) ([]byte, error) {
	return project_analyzer.ToJSONBytes(r, indent)
}

// ToConsole 以易于阅读的格式在控制台打印分析结果
func (r *Result) ToConsole() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("组件依赖分析报告:\n"))

	for pkgName, components := range r.Packages {
		buffer.WriteString("\n=====================================\n")
		buffer.WriteString(fmt.Sprintf("📦 包: %s (%d 个组件)\n", pkgName, len(components)))
		buffer.WriteString("=====================================\n")

		for name, info := range components {
			buffer.WriteString(fmt.Sprintf("\n▶ 组件: %s\n", name))
			buffer.WriteString(fmt.Sprintf("  - 源文件: %s\n", info.SourcePath))

			if len(info.Dependencies) > 0 {
				buffer.WriteString("  - 依赖的组件:\n")
				for _, dep := range info.Dependencies {
					buffer.WriteString(fmt.Sprintf("    - %s\n", dep))
				}
			} else {
				buffer.WriteString("  - 依赖的组件: 无\n")
			}
		}
	}

	return buffer.String()
}
