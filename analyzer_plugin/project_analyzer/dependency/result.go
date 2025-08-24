package dependency

import (
	"fmt"
	projectanalyzer "main/analyzer_plugin/project_analyzer"
	"strings"
)

// DependencyCheckResult 是依赖检查功能最终输出的完整结果结构体。
// 它整合了隐式依赖、未使用依赖和过期依赖三项检查的结果，并实现了 projectanalyzer.Result 接口。
type DependencyCheckResult struct {
	ImplicitDependencies []ImplicitDependency `json:"implicitDependencies"`
	UnusedDependencies   []UnusedDependency   `json:"unusedDependencies"`
	OutdatedDependencies []OutdatedDependency `json:"outdatedDependencies"`
}

// 确保 Result 结构体实现了 projectanalyzer.Result 接口。
var _ projectanalyzer.Result = (*DependencyCheckResult)(nil)

// Name 返回该结果对应的分析器的名称。
func (r *DependencyCheckResult) Name() string {
	return "NPM Dependency Check"
}

// Summary 返回对结果的简短、人类可读的摘要。
func (r *DependencyCheckResult) Summary() string {
	return fmt.Sprintf(
		"发现 %d 个隐式依赖, %d 个未使用依赖, %d 个过期依赖。",
		len(r.ImplicitDependencies),
		len(r.UnusedDependencies),
		len(r.OutdatedDependencies),
	)
}

// ToJSON 将结果的完整数据序列化为 JSON 格式。
func (r *DependencyCheckResult) ToJSON(indent bool) ([]byte, error) {
	return projectanalyzer.ToJSONBytes(r, indent)
}

// ToConsole 将结果格式化为适合在控制台（终端）中打印的字符串。
func (r *DependencyCheckResult) ToConsole() string {
	totalIssues := len(r.ImplicitDependencies) + len(r.UnusedDependencies) + len(r.OutdatedDependencies)
	if totalIssues == 0 {
		return "✅ NPM 依赖健康检查通过，没有发现任何问题。"
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("⚠️ %s\n", r.Summary()))

	if len(r.ImplicitDependencies) > 0 {
		builder.WriteString("\n--- 👻 隐式依赖 (幽灵依赖) ---\n")
		for _, dep := range r.ImplicitDependencies {
			builder.WriteString(fmt.Sprintf("  - %s (在 %s 中使用)\n", dep.Name, dep.FilePath))
		}
	}

	if len(r.UnusedDependencies) > 0 {
		builder.WriteString("\n--- 🗑️ 未使用依赖 ---\n")
		for _, dep := range r.UnusedDependencies {
			builder.WriteString(fmt.Sprintf("  - %s@%s (在 %s 中声明)\n", dep.Name, dep.Version, dep.PackageJsonPath))
		}
	}

	if len(r.OutdatedDependencies) > 0 {
		builder.WriteString("\n--- ⬆️ 过期依赖 ---\n")
		for _, dep := range r.OutdatedDependencies {
			builder.WriteString(fmt.Sprintf("  - %s: %s -> %s (在 %s 中声明)\n", dep.Name, dep.CurrentVersion, dep.LatestVersion, dep.PackageJsonPath))
		}
	}

	return builder.String()
}
