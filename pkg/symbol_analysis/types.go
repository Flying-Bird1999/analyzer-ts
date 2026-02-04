// Package symbol_analysis 提供符号级代码分析功能。
// 它分析 git diff 变更对符号（函数、变量、类等）的影响，并确定它们的导出状态。
package symbol_analysis

import (
	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

// =============================================================================
// 类型别名
// =============================================================================

// ChangedLineSetOfFiles 表示每个文件的变更行集合。
// 格式：map[文件路径]map[行号]bool
type ChangedLineSetOfFiles = map[string]map[int]bool

// =============================================================================
// 分析选项
// =============================================================================

// AnalysisOptions 配置符号分析器的行为。
type AnalysisOptions struct {
	// IncludeTypes 表示是否包含类型声明（接口、类型别名）
	IncludeTypes bool
	// IncludeInternal 表示是否包含非导出符号
	IncludeInternal bool
}

// DefaultAnalysisOptions 返回默认的分析选项。
func DefaultAnalysisOptions() AnalysisOptions {
	return AnalysisOptions{
		IncludeTypes:    true,
		IncludeInternal: true,
	}
}

// =============================================================================
// 分析结果
// =============================================================================

// FileType 表示文件的类型。
type FileType string

const (
	FileTypeTypeScript FileType = "typescript" // TypeScript 文件
	FileTypeJavaScript FileType = "javascript" // JavaScript 文件
	FileTypeBinary     FileType = "binary"     // 二进制文件（图片、字体等）
	FileTypeStyle      FileType = "style"      // 样式文件（CSS、SCSS、LESS等）
	FileTypeMarkup     FileType = "markup"     // 标记文件（HTML、XML等）
	FileTypeData       FileType = "data"       // 数据文件（JSON、YAML等）
	FileTypeUnknown    FileType = "unknown"    // 未知类型
)

// FileAnalysisResult 包含单个文件的分析结果。
type FileAnalysisResult struct {
	FilePath        string         // 文件路径
	FileType        FileType       // 文件类型
	AffectedSymbols []SymbolChange // 该文件中受影响的符号（仅符号文件）
	FileExports     []ExportInfo   // 该文件的所有导出（仅符号文件）
	ReExports       []ReExportInfo // 该文件的所有重新导出（export { X } from './Y'）
	ChangedLines    []int          // 变更行号（所有文件）
	IsSymbolFile    bool           // 是否为符号文件（可以进行符号分析）
}

// =============================================================================
// 符号变更
// =============================================================================

// SymbolChange 表示符号（函数、变量、类等）的变更。
type SymbolChange struct {
	// 符号标识
	Name string     // 符号名称，如 "handleClick"、"Button"
	Kind SymbolKind // 符号类型

	// 位置信息
	FilePath  string // 源文件路径
	StartLine int    // 起始行号（从1开始）
	EndLine   int    // 结束行号（从1开始）

	// 变更信息
	ChangedLines []int      // 符号内部的实际变更行号
	ChangeType   ChangeType // 变更类型（修改、新增、删除）

	// 导出信息（在导出分析后填充）
	ExportType ExportType // 导出类型（命名、默认、命名空间）
	IsExported bool       // 是否导出
}

// =============================================================================
// 符号类型
// =============================================================================

// SymbolKind 表示符号的类型。
type SymbolKind string

const (
	SymbolKindFunction  SymbolKind = "function"   // 函数声明
	SymbolKindVariable  SymbolKind = "variable"   // 变量声明
	SymbolKindClass     SymbolKind = "class"      // 类声明
	SymbolKindInterface SymbolKind = "interface"  // 接口声明
	SymbolKindTypeAlias SymbolKind = "type-alias" // 类型别名
	SymbolKindEnum      SymbolKind = "enum"       // 枚举声明
	SymbolKindMethod    SymbolKind = "method"     // 类方法
	SymbolKindProperty  SymbolKind = "property"   // 类属性
	SymbolKindParameter SymbolKind = "parameter"  // 参数
)

// String 返回 SymbolKind 的字符串表示。
func (k SymbolKind) String() string {
	return string(k)
}

// =============================================================================
// 变更类型
// =============================================================================

// ChangeType 表示变更的类型。
type ChangeType string

const (
	ChangeTypeModified ChangeType = "modified" // 修改
	ChangeTypeAdded    ChangeType = "added"    // 新增
	ChangeTypeDeleted  ChangeType = "deleted"  // 删除
)

// String 返回 ChangeType 的字符串表示。
func (c ChangeType) String() string {
	return string(c)
}

// =============================================================================
// 导出类型
// =============================================================================

// ExportType 表示导出的类型。
type ExportType string

const (
	ExportTypeNone      ExportType = ""          // 非导出
	ExportTypeNamed     ExportType = "named"     // 命名导出：export const A = 1
	ExportTypeDefault   ExportType = "default"   // 默认导出：export default A
	ExportTypeNamespace ExportType = "namespace" // 命名空间导出：export * as A
)

// String 返回 ExportType 的字符串表示。
func (e ExportType) String() string {
	return string(e)
}

// =============================================================================
// 导出信息
// =============================================================================

// ExportInfo 表示导出符号的信息。
type ExportInfo struct {
	Name       string     // 符号名称
	ExportType ExportType // 导出类型
	DeclLine   int        // 声明行号
	DeclNode   string     // 声明节点类型（用于调试）
}

// ReExportInfo 表示重新导出的信息。
type ReExportInfo struct {
	OriginalPath  string   // 原始导出路径
	ExportedNames []string // 重新导出的符号名称
}

// =============================================================================
// 符号分析器
// =============================================================================

// Analyzer 对 TypeScript 代码执行符号级分析。
type Analyzer struct {
	project *tsmorphgo.Project
	options AnalysisOptions
}

// NewAnalyzer 创建一个新的符号分析器。
func NewAnalyzer(project *tsmorphgo.Project, opts AnalysisOptions) *Analyzer {
	return &Analyzer{
		project: project,
		options: opts,
	}
}

// NewAnalyzerWithDefaults 使用默认选项创建一个新的符号分析器。
func NewAnalyzerWithDefaults(project *tsmorphgo.Project) *Analyzer {
	return NewAnalyzer(project, DefaultAnalysisOptions())
}
