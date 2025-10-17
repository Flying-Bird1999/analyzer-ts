// Package dependency 包含了 NPM 依赖分析功能所需的所有类型定义。
//
// 这个包定义了依赖检查分析器的完整类型系统，用于：
// - 隐式依赖（幽灵依赖）的检测和报告
// - 未使用依赖的识别和分类
// - 过时依赖的版本信息记录
// - NPM Registry API 响应数据的解析
//
// 所有类型都支持 JSON 序列化，便于数据导出和集成到其他系统。
package dependency

// ImplicitDependency 代表一个隐式依赖（或称“幽灵依赖”）。
type ImplicitDependency struct {
	Name     string `json:"name"`
	FilePath string `json:"filePath"`
	Raw      string `json:"raw"`
}

// UnusedDependency 代表一个未使用的依赖。
type UnusedDependency struct {
	Name            string `json:"name"`
	Version         string `json:"version"`
	PackageJsonPath string `json:"packageJsonPath"`
}

// OutdatedDependency 代表一个已过期的依赖。
type OutdatedDependency struct {
	Name            string `json:"name"`
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion"`
	PackageJsonPath string `json:"packageJsonPath"`
}

// packageInfo 是一个辅助结构体，用于解析从 NPM Registry API 返回的 JSON 数据。
type packageInfo struct {
	DistTags struct {
		Latest string `json:"latest"`
	} `json:"dist-tags"`
}
