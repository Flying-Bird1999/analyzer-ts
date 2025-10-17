package component_deps

import (
	"bytes"
	"fmt"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

// ComponentInfo 包含了单个公共组件的详细分析信息
// 该结构体存储了每个组件的源文件路径和依赖关系
// 用于构建完整的组件依赖图和可视化展示
//
// JSON 标签说明:
//   - sourcePath: 组件源文件的完整路径
//   - dependencies: 该组件依赖的其他公共组件名称列表
type ComponentInfo struct {
	SourcePath   string   `json:"sourcePath"`   // 组件源文件的完整路径
	Dependencies []string `json:"dependencies"` // 依赖的其他公共组件名称列表
}

// Result 保存了对组件库的完整分析结果，以 package 分组
// 该结构体是分析器的最终输出结果，包含了所有组件的依赖关系信息
// 支持多包分析，每个包可以包含多个组件
//
// JSON 标签说明:
//   - packages: 包名 -> 组件名 -> 组件信息的嵌套映射结构
type Result struct {
	Packages map[string]map[string]ComponentInfo `json:"packages"` // 包名到组件信息的映射
}

// Name 返回分析结果的标识符
// 用于在插件系统中识别和分类不同的分析结果
// 该值与分析器的名称保持一致
func (r *Result) Name() string {
	return "component-deps"
}

// Summary 返回分析结果的摘要信息
// 提供分析结果的统计概览，包括:
//   - 分析的包总数
//   - 发现的公共组件总数
//
// 返回值:
//   - 包含统计信息的格式化字符串
func (r *Result) Summary() string {
	packageCount := len(r.Packages) // 包总数
	componentCount := 0
	for _, components := range r.Packages {
		componentCount += len(components) // 累计组件总数
	}
	return fmt.Sprintf("分析完成，共找到 %d 个包中的 %d 个公共组件。", packageCount, componentCount)
}

// ToJSON 将分析结果序列化为 JSON 格式
// 支持带缩进和不带缩进的 JSON 输出格式
// 便于机器处理和数据持久化
//
// 参数:
//   - indent: 是否格式化 JSON 输出（带缩进和换行）
//
// 返回值:
//   - []byte: JSON 格式的字节数据
//   - error: 序列化过程中的错误
func (r *Result) ToJSON(indent bool) ([]byte, error) {
	return project_analyzer.ToJSONBytes(r, indent)
}

// ToConsole 以易于阅读的格式在控制台打印分析结果
// 生成的报告包含以下内容：
//   - 总体标题和概览
//   - 每个包的详细信息（带图标装饰）
//   - 每个组件的详细信息，包括源文件路径和依赖关系
//   - 使用清晰的层级结构和视觉分隔线
//
// 报告格式特点：
//   - 使用 Unicode 图标增强可读性
//   - 清晰的层级缩进
//   - 分隔线区分不同的包和组件
//   - 依赖列表使用嵌套格式显示
//
// 返回值:
//   - 包含完整分析报告的格式化字符串
func (r *Result) ToConsole() string {
	var buffer bytes.Buffer
	// 报告标题
	buffer.WriteString(fmt.Sprintf("组件依赖分析报告:\n"))

	// 遍历每个包，生成包级别的信息
	for pkgName, components := range r.Packages {
		// 包分隔线
		buffer.WriteString("\n=====================================\n")
		// 包标题，包含包名和组件数量统计
		buffer.WriteString(fmt.Sprintf("📦 包: %s (%d 个组件)\n", pkgName, len(components)))
		buffer.WriteString("=====================================\n")

		// 遍历包中的每个组件，生成组件级别的信息
		for name, info := range components {
			// 组件名称标题
			buffer.WriteString(fmt.Sprintf("\n▶ 组件: %s\n", name))
			// 组件源文件路径
			buffer.WriteString(fmt.Sprintf("  - 源文件: %s\n", info.SourcePath))

			// 处理依赖关系信息
			if len(info.Dependencies) > 0 {
				// 如果有依赖，显示依赖列表
				buffer.WriteString("  - 依赖的组件:\n")
				for _, dep := range info.Dependencies {
					buffer.WriteString(fmt.Sprintf("    - %s\n", dep))
				}
			} else {
				// 如果没有依赖，显示无依赖信息
				buffer.WriteString("  - 依赖的组件: 无\n")
			}
		}
	}

	return buffer.String()
}
