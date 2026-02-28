// Package impact_analysis 提供符号级影响分析功能。
// 将影响分析拆分为两个独立能力：
// - file_analyzer: 文件级影响分析（通用能力，适用于所有前端项目）
// - component_analyzer: 组件级影响分析（组件库专用）
package impact_analysis

import (
	"encoding/json"
	"os"
	"time"
)

// =============================================================================
// 文件变更类型
// =============================================================================

// FileChange 文件变更信息
type FileChange struct {
	Path string     `json:"path"` // 文件绝对路径
	Type ChangeType `json:"type"` // 变更类型
}

// =============================================================================
// 符号信息（轻量级）
// =============================================================================

// SymbolReference 符号引用
type SymbolReference struct {
	Name       string     `json:"name"`       // 符号名称
	Kind       string     `json:"kind"`       // function/class/variable
	FilePath   string     `json:"filePath"`   // 文件绝对路径
	IsExported bool       `json:"isExported"` // 是否导出
	ExportType ExportType `json:"exportType"` // named/default/namespace
}

// =============================================================================
// 影响级别和类型
// =============================================================================

// ImpactLevel 影响级别（0=直接变更，1=间接影响，2+=传递影响）
type ImpactLevel int

const (
	ImpactLevelDirect     ImpactLevel = 0 // 直接变更
	ImpactLevelIndirect   ImpactLevel = 1 // 间接受影响
	ImpactLevelTransitive ImpactLevel = 2 // 传递影响
)

// =============================================================================
// 变更类型
// =============================================================================

// ChangeType 变更类型
type ChangeType string

const (
	ChangeTypeModified ChangeType = "modified" // 修改
	ChangeTypeAdded    ChangeType = "added"    // 新增
	ChangeTypeDeleted  ChangeType = "deleted"  // 删除
)

// =============================================================================
// 导出类型
// =============================================================================

// ExportType 导出类型
type ExportType string

const (
	ExportTypeNone      ExportType = ""          // 非导出
	ExportTypeNamed     ExportType = "named"     // 命名导出：export const A = 1
	ExportTypeDefault   ExportType = "default"   // 默认导出：export default A
	ExportTypeNamespace ExportType = "namespace" // 命名空间导出：export * as A
)

// =============================================================================
// 符号类型
// =============================================================================

// SymbolKind 符号类型
type SymbolKind string

const (
	SymbolKindFunction  SymbolKind = "function"   // 函数声明
	SymbolKindVariable  SymbolKind = "variable"   // 变量声明
	SymbolKindClass     SymbolKind = "class"      // 类声明
	SymbolKindInterface SymbolKind = "interface"  // 接口声明
	SymbolKindTypeAlias SymbolKind = "type-alias" // 类型别名
	SymbolKindEnum      SymbolKind = "enum"       // 枚举声明
)

// =============================================================================
// 符号变更（与 symbol_analysis 对齐）
// =============================================================================

// SymbolChange 符号变更
type SymbolChange struct {
	// 符号标识
	Name string     `json:"name"` // 符号名称，如 "handleClick"
	Kind SymbolKind `json:"kind"` // function/class/variable/etc.

	// 位置信息
	FilePath  string `json:"filePath"`  // 文件绝对路径
	StartLine int    `json:"startLine"` // 起始行号
	EndLine   int    `json:"endLine"`   // 结束行号

	// 变更信息
	ChangedLines []int      `json:"changedLines"` // 符号内部的实际变更行号
	ChangeType   ChangeType `json:"changeType"`   // modified/added/deleted

	// 导出信息
	ExportType ExportType `json:"exportType"` // named/default/namespace
	IsExported bool       `json:"isExported"` // 是否导出

	// 所属组件（在匹配后填充）
	ComponentName string `json:"componentName,omitempty"` // 所属组件名称
}

// =============================================================================
// 组件清单（与 component_deps 兼容）
// =============================================================================

// ComponentManifest 组件清单
type ComponentManifest struct {
	Meta       ManifestMeta       `json:"meta"`
	Components map[string]Component `json:"components"`
}

// ManifestMeta 清单元数据
type ManifestMeta struct {
	Version     string `json:"version"`     // 清单版本
	LibraryName string `json:"libraryName"` // 库名称
}

// Component 组件定义
type Component struct {
	Name string `json:"name"` // 组件名称
	Path string `json:"path"` // 组件目录路径（相对路径）
	Type string `json:"type"` // 组件类型: "component" 或 "functions"
}

// =============================================================================
// 配置解析
// =============================================================================

// ParseSymbolChangeInput 从 JSON 解析符号变更输入
func ParseSymbolChangeInput(jsonStr string) (*SymbolChangeInput, error) {
	var input SymbolChangeInput
	if err := json.Unmarshal([]byte(jsonStr), &input); err != nil {
		return nil, err
	}
	return &input, nil
}

// LoadSymbolChangeInput 从文件加载符号变更输入
func LoadSymbolChangeInput(filePath string) (*SymbolChangeInput, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return ParseSymbolChangeInput(string(data))
}

// =============================================================================
// 兼容性：保留旧的输入格式
// =============================================================================

// SymbolChangeInput 符号级变更输入（旧格式）
type SymbolChangeInput struct {
	ChangedSymbols []SymbolChange `json:"changedSymbols"`

	// 组件依赖图（从 component-deps-v2 获取）
	DepGraph    map[string][]string `json:"depGraph"`    // [组件][依赖组件列表]
	RevDepGraph map[string][]string `json:"revDepGraph"` // [组件][被依赖组件列表]

	// 组件清单
	ComponentManifest *ComponentManifest `json:"componentManifest,omitempty"`
}

// FileChangeInput 文件级变更输入（旧格式）
type FileChangeInput struct {
	ModifiedFiles []string `json:"modifiedFiles"` // 修改的文件列表
	AddedFiles    []string `json:"addedFiles"`    // 新增的文件列表
	DeletedFiles  []string `json:"deletedFiles"`  // 删除的文件列表
}

// ParseFileChangeInput 从 JSON 解析文件变更输入
func ParseFileChangeInput(jsonStr string) (*FileChangeInput, error) {
	var input FileChangeInput
	if err := json.Unmarshal([]byte(jsonStr), &input); err != nil {
		return nil, err
	}
	return &input, nil
}

// GetAllFiles 获取变更涉及的所有文件
func (c *FileChangeInput) GetAllFiles() []string {
	files := make([]string, 0)
	files = append(files, c.ModifiedFiles...)
	files = append(files, c.AddedFiles...)
	files = append(files, c.DeletedFiles...)
	return files
}

// GetFileCount 获取变更文件总数
func (c *FileChangeInput) GetFileCount() int {
	return len(c.ModifiedFiles) + len(c.AddedFiles) + len(c.DeletedFiles)
}

// IsEmpty 检查是否为空变更
func (c *FileChangeInput) IsEmpty() bool {
	return len(c.ModifiedFiles) == 0 &&
		len(c.AddedFiles) == 0 &&
		len(c.DeletedFiles) == 0
}

// =============================================================================
// 分析结果元数据
// =============================================================================

// AnalysisMeta 分析元数据
type AnalysisMeta struct {
	AnalyzedAt       time.Time `json:"analyzedAt"`
	TotalFileCount   int       `json:"totalFileCount"`
	ChangedFileCount int       `json:"changedFileCount"`
	ImpactFileCount  int       `json:"impactFileCount"`
	ComponentCount   int       `json:"componentCount,omitempty"`
}
