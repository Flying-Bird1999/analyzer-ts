// package unreferenced 包含了查找项目中未使用文件的功能所需的所有类型定义。
package unreferenced

// SummaryStats 提供了关于未引用文件分析的统计数据。
type SummaryStats struct {
	TotalFiles             int `json:"totalFiles"`
	ReferencedFiles        int `json:"referencedFiles"`
	TrulyUnreferencedFiles int `json:"trulyUnreferencedFiles"`
	SuspiciousFiles        int `json:"suspiciousFiles"`
}

// AnalysisConfiguration 记录了用于本次分析的输入配置，便于追溯结果。
type AnalysisConfiguration struct {
	InputDir             string `json:"inputDir"`
	EntrypointsSpecified bool   `json:"entrypointsSpecified"`
	IncludeEntryDirs     bool   `json:"includeEntryDirs"`
}
