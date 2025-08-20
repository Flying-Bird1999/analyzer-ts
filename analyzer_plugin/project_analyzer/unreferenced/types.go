// package unreferenced 包含了查找项目中未使用文件的功能所需的所有类型定义。
package unreferenced

// Params 定义了 Find 函数的输入参数。
type Params struct {
	// RootPath 是要分析的项目根目录的绝对路径。
	RootPath string
	// Exclude 是一个 glob 模式列表，用于在分析时排除特定的文件或目录。
	Exclude []string
	// IsMonorepo 指示正在分析的项目是否是一个 monorepo 仓库。
	IsMonorepo bool
	// Entrypoints 是一个可选的入口文件列表。如果指定，分析将变为“可达性分析”，
	// 只报告从这些入口文件出发无法访问到的文件。
	Entrypoints []string
	// IncludeEntryDirs 指示是否应自动将项目中常见的入口文件（如 index.ts, main.ts）视为入口点。
	IncludeEntryDirs bool
}

// Result 保存了未引用文件分析的完整结果。
type Result struct {
	// Configuration 记录了本次分析所使用的配置参数。
	Configuration AnalysisConfiguration `json:"configuration"`
	// Summary 包含了本次分析的各项统计数据。
	Summary SummaryStats `json:"summary"`
	// EntrypointFiles 是在本次分析中被当作入口点的文件列表。
	EntrypointFiles []string `json:"entrypointFiles"`
	// SuspiciousFiles 是一些虽然未被直接引用，但根据其命名或位置，可能很重要的文件（例如配置文件），需要人工检查。
	SuspiciousFiles []string `json:"suspiciousFiles"`
	// TrulyUnreferencedFiles 是被认为是“真正”未被引用的文件列表，可以相对安全地删除。
	TrulyUnreferencedFiles []string `json:"trulyUnreferencedFiles"`
}

// SummaryStats 提供了关于未引用文件分析的统计数据。
type SummaryStats struct {
	// TotalFiles 是项目中被分析的 JS/TS 文件的总数。
	TotalFiles int `json:"totalFiles"`
	// ReferencedFiles 是被其他文件引用过的文件的总数。
	ReferencedFiles int `json:"referencedFiles"`
	// TrulyUnreferencedFiles 是“真正”未引用文件的数量。
	TrulyUnreferencedFiles int `json:"trulyUnreferencedFiles"`
	// SuspiciousFiles 是需要人工检查的可疑文件的数量。
	SuspiciousFiles int `json:"suspiciousFiles"`
}

// AnalysisConfiguration 记录了用于本次分析的输入配置，便于追溯结果。
type AnalysisConfiguration struct {
	// InputDir 是分析的根目录。
	InputDir string `json:"inputDir"`
	// EntrypointsSpecified 指示用户是否通过命令行指定了入口文件。
	EntrypointsSpecified bool `json:"entrypointsSpecified"`
	// IncludeEntryDirs 指示是否启用了自动检测入口文件的功能。
	IncludeEntryDirs bool `json:"includeEntryDirs"`
}
