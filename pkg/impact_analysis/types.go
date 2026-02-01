// Package impact_analysis 提供符号级影响分析功能。
// 它基于 symbol_analysis 的输出（导出符号变更）和组件依赖图，分析代码变更的影响范围。
package impact_analysis

import (
	"encoding/json"
	"os"
)

// =============================================================================
// 输入类型定义（与 symbol_analysis 对齐）
// =============================================================================

// SymbolChangeInput 符号级变更输入
type SymbolChangeInput struct {
	// 变更的符号列表（只包含导出符号）
	ChangedSymbols []SymbolChange `json:"changedSymbols"`

	// 组件依赖图（从 component-deps-v2 获取）
	DepGraph    map[string][]string `json:"depGraph"`    // [组件][依赖组件列表]
	RevDepGraph map[string][]string `json:"revDepGraph"` // [组件][被依赖组件列表]

	// 组件清单（用于符号到组件的映射）
	ComponentManifest *ComponentManifest `json:"componentManifest,omitempty"`
}

// SymbolChange 符号变更（从 symbol_analysis.SymbolChange 转换而来）
type SymbolChange struct {
	// 符号标识
	Name string     `json:"name"` // 符号名称，如 "handleClick"
	Kind SymbolKind `json:"kind"` // function/class/variable/etc.

	// 位置信息
	FilePath  string `json:"filePath"`  // 源文件路径
	StartLine int    `json:"startLine"` // 起始行号
	EndLine   int    `json:"endLine"`   // 结束行号

	// 变更信息
	ChangedLines []int      `json:"changedLines"` // 符号内部的实际变更行号
	ChangeType   ChangeType `json:"changeType"`  // modified/added/deleted

	// 导出信息（关键字段）
	ExportType ExportType `json:"exportType"` // named/default/namespace
	IsExported  bool       `json:"isExported"`  // 是否导出

	// 所属组件（在匹配后填充）
	ComponentName string `json:"componentName,omitempty"` // 所属组件名称
}

// =============================================================================
// 符号类型
// =============================================================================

// SymbolKind 符号类型
type SymbolKind string

const (
	SymbolKindFunction  SymbolKind = "function"  // 函数声明
	SymbolKindVariable  SymbolKind = "variable"  // 变量声明
	SymbolKindClass     SymbolKind = "class"     // 类声明
	SymbolKindInterface SymbolKind = "interface" // 接口声明
	SymbolKindTypeAlias SymbolKind = "type-alias" // 类型别名
	SymbolKindEnum      SymbolKind = "enum"      // 枚举声明
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
	ExportTypeNone      ExportType = ""        // 非导出
	ExportTypeNamed     ExportType = "named"   // 命名导出：export const A = 1
	ExportTypeDefault   ExportType = "default" // 默认导出：export default A
	ExportTypeNamespace ExportType = "namespace" // 命名空间导出：export * as A
)

// =============================================================================
// 组件清单
// =============================================================================

// ComponentManifest 组件清单
type ComponentManifest struct {
	Meta       ManifestMeta            `json:"meta"`
	Components []Component             `json:"components"`
}

// ManifestMeta 清单元数据
type ManifestMeta struct {
	Version     string `json:"version"`     // 清单版本
	LibraryName string `json:"libraryName"` // 库名称
}

// Component 组件定义
type Component struct {
	Name   string   `json:"name"`   // 组件名称
	Scopes []string `json:"scopes"` // 组件文件路径范围
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
// 兼容性：保留旧的文件级输入
// =============================================================================

// FileChangeInput 文件级变更输入（旧格式，向后兼容）
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
