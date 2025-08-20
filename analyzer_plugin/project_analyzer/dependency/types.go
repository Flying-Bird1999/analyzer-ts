// package dependency 包含了NPM依赖分析功能所需的所有类型定义。
package dependency

// DependencyCheckResult 是依赖检查功能最终输出的完整结果结构体。
// 它整合了隐式依赖、未使用依赖和过期依赖三项检查的结果。
type DependencyCheckResult struct {
	// ImplicitDependencies 是在代码中被使用但未在 package.json 中声明的“幽灵依赖”列表。
	ImplicitDependencies []ImplicitDependency `json:"implicitDependencies"`
	// UnusedDependencies 是在 package.json 中声明但代码中从未被使用过的依赖列表。
	UnusedDependencies []UnusedDependency `json:"unusedDependencies"`
	// OutdatedDependencies 是当前版本落后于NPM仓库最新版本的依赖列表。
	OutdatedDependencies []OutdatedDependency `json:"outdatedDependencies"`
}

// ImplicitDependency 代表一个隐式依赖（或称“幽灵依赖”）。
// 这是指在代码中被 import/require，但没有在当前项目的 package.json 中明确声明的依赖。
// 这种情况通常发生在 monorepo 中，或者由于 npm/yarn 的依赖提升机制导致。
type ImplicitDependency struct {
	// Name 是依赖包的名称，例如 "react"。
	Name string `json:"name"`
	// FilePath 是发现该隐式依赖的代码文件的路径。
	FilePath string `json:"filePath"`
	// Raw 是导致发现该依赖的原始导入语句，用于快速定位。
	Raw string `json:"raw"`
}

// UnusedDependency 代表一个未使用的依赖。
// 这是指在 package.json 中声明了，但在整个项目的任何代码中都没有被实际使用的依赖。
type UnusedDependency struct {
	// Name 是依赖包的名称。
	Name string `json:"name"`
	// Version 是在 package.json 中声明的版本号。
	Version string `json:"version"`
	// PackageJsonPath 是声明该依赖的 package.json 文件路径。
	PackageJsonPath string `json:"packageJsonPath"`
}

// OutdatedDependency 代表一个已过期的依赖。
// 这是指在 package.json 中声明的版本落后于 NPM Registry 中记录的最新版本。
type OutdatedDependency struct {
	// Name 是依赖包的名称。
	Name string `json:"name"`
	// CurrentVersion 是当前在 package.json 中声明的版本。
	CurrentVersion string `json:"currentVersion"`
	// LatestVersion 是从 NPM Registry 获取到的最新版本。
	LatestVersion string `json:"latestVersion"`
	// PackageJsonPath 是声明该依赖的 package.json 文件路径。
	PackageJsonPath string `json:"packageJsonPath"`
}

// packageInfo 是一个辅助结构体，用于解析从 NPM Registry API 返回的 JSON 数据。
// 我们只关心 `dist-tags.latest` 字段以获取最新版本号。
type packageInfo struct {
	DistTags struct {
		Latest string `json:"latest"`
	} `json:"dist-tags"`
}