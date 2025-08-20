// package unconsumed 包含了查找项目中已导出但未被消费的变量所需的所有类型定义。
package unconsumed

// Params 定义了 Find 函数的输入参数。
type Params struct {
	// RootPath 是要分析的项目根目录的绝对路径。
	RootPath string
	// Exclude 是一个 glob 模式列表，用于在分析时排除特定的文件或目录。
	Exclude []string
	// IsMonorepo 指示正在分析的项目是否是一个 monorepo 仓库。
	IsMonorepo bool
}

// Result 保存了未消费导出变量的分析结果。
type Result struct {
	// UnconsumedExports 是找到的所有未被消费的导出变量的列表。
	UnconsumedExports []Export `json:"unconsumedExports"`
	// Summary 包含了本次分析的统计数据。
	Summary SummaryStats `json:"summary"`
}

// Export 代表一个具体的、已导出但未被消费的变量、函数或类型等。
type Export struct {
	// FilePath 是这个导出项所在的文件路径。
	FilePath string `json:"filePath"`
	// ExportName 是导出项的名称（标识符）。
	ExportName string `json:"exportName"`
	// Line 是导出语句所在的行号，便于快速定位。
	Line int `json:"line"`
	// Kind 描述了导出项的类型（如 var, const, let, function, class, interface, type, enum）。
	Kind string `json:"kind"`
}

// SummaryStats 提供了关于分析的统计数据。
type SummaryStats struct {
	// TotalFilesScanned 是本次分析扫描的总文件数。
	TotalFilesScanned int `json:"totalFilesScanned"`
	// TotalExportsFound 是在项目中找到的总导出项数量。
	TotalExportsFound int `json:"totalExportsFound"`
	// UnconsumedExportsFound 是未被消费的导出项的数量。
	UnconsumedExportsFound int `json:"unconsumedExportsFound"`
}
