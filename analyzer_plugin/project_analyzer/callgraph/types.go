// package callgraph 包含了调用链分析功能所需的所有类型定义。
package callgraph

// Params 定义了旧的 Find 函数的输入参数。在重构为 Analyzer 接口后，
// 这些参数将通过 ProjectContext 或其他方式传入，此结构体仅为过渡保留。
type Params struct {
	RootPath    string
	TargetFiles []string
	Exclude     []string
	IsMonorepo  bool
}

// OverallSummary 包含所有目标文件的聚合汇总数据。
type OverallSummary struct {
	TargetFiles        []string `json:"targetFiles"`
	TotalAffectedFiles int      `json:"totalAffectedFiles"`
	AffectedFilesList  []string `json:"affectedFilesList"`
}

// SingleFileResult 代表对单个目标文件的完整分析结果。
type SingleFileResult struct {
	Summary  SingleFileSummary `json:"summary"`
	CallTree CallerNode        `json:"callTree"`
}

// SingleFileSummary 包含单个目标文件的汇总数据。
type SingleFileSummary struct {
	TargetFile         string   `json:"targetFile"`
	TotalAffectedFiles int      `json:"totalAffectedFiles"`
	AffectedFilesList  []string `json:"affectedFilesList"`
}

// CallerNode 定义了调用树的数据结构，代表调用关系图中的一个节点。
type CallerNode struct {
	FilePath string       `json:"filePath"`
	Callers  []CallerNode `json:"callers"`
}
