// package callgraph 包含了调用链分析功能所需的所有类型定义。
package callgraph

// Params 定义了 Find 函数的输入参数。
type Params struct {
	// RootPath 是要分析的项目根目录的绝对路径。
	RootPath string
	// TargetFiles 是一个或多个要追踪其调用链的目标文件的绝对路径列表。
	TargetFiles []string
	// Exclude 是一个 glob 模式列表，用于在分析时排除特定的文件或目录。
	Exclude []string
	// IsMonorepo 指示正在分析的项目是否是一个 monorepo 仓库。
	IsMonorepo bool
}

// Result 是用于多文件调用链分析的顶级结果结构体。
type Result struct {
	// OverallSummary 包含了对所有目标文件分析结果的聚合汇总。
	OverallSummary OverallSummary `json:"overallSummary"`
	// PerFileResults 包含了每个目标文件的独立分析结果。
	PerFileResults []SingleFileResult `json:"perFileResults"`
}

// OverallSummary 包含所有目标文件的聚合汇总数据。
type OverallSummary struct {
	// TargetFiles 列出了本次分析的所有目标文件。
	TargetFiles []string `json:"targetFiles"`
	// TotalAffectedFiles 是指所有受目标文件变更影响的上游文件的总数（去重后）。
	TotalAffectedFiles int `json:"totalAffectedFiles"`
	// AffectedFilesList 是所有受影响的上游文件的路径列表（去重并排序后）。
	AffectedFilesList []string `json:"affectedFilesList"`
}

// SingleFileResult 代表对单个目标文件的完整分析结果。
type SingleFileResult struct {
	// Summary 包含了对这单个目标文件分析结果的汇总。
	Summary SingleFileSummary `json:"summary"`
	// CallTree 是为这单个目标文件生成的完整上游调用树。
	CallTree CallerNode `json:"callTree"`
}

// SingleFileSummary 包含单个目标文件的汇总数据。
type SingleFileSummary struct {
	// TargetFile 是本次分析的目标文件。
	TargetFile string `json:"targetFile"`
	// TotalAffectedFiles 是指受该目标文件变更影响的上游文件的总数。
	TotalAffectedFiles int `json:"totalAffectedFiles"`
	// AffectedFilesList 是所有受影响的上游文件的路径列表。
	AffectedFilesList []string `json:"affectedFilesList"`
}

// CallerNode 定义了调用树的数据结构，代表调用关系图中的一个节点。
type CallerNode struct {
	// FilePath 是当前节点代表的文件的绝对路径。
	FilePath string `json:"filePath"`
	// Callers 是一个 CallerNode 列表，代表所有直接引用了当前文件的上游文件节点。
	Callers []CallerNode `json:"callers"`
}