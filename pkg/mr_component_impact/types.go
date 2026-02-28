// Package mr_component_impact 提供 MR 组件影响分析功能
//
// 核心功能：
// - 基于 git diff 分析代码变更对组件的影响范围
// - 支持 component_deps 和 export_call 的组件级引用分析
// - 简单直接的影响传播，无需复杂的 BFS 算法
package mr_component_impact

// =============================================================================
// 文件分类相关类型
// =============================================================================

// FileCategory 文件分类类型
type FileCategory string

const (
	CategoryComponent FileCategory = "component" // 组件文件
	CategoryFunctions FileCategory = "functions" // 函数/工具文件
	CategoryOther     FileCategory = "other"     // 其他文件
)

// =============================================================================
// 分析结果相关类型
// =============================================================================

// AnalysisResult 分析结果
type AnalysisResult struct {
	ChangedComponents  map[string]*ComponentChangeInfo `json:"changedComponents"`  // 变更的组件
	ChangedFunctions   map[string]*FunctionChangeInfo  `json:"changedFunctions"`   // 变更的函数
	ImpactedComponents map[string][]ComponentImpact    `json:"impactedComponents"` // 受影响的组件
	OtherFiles         []string                        `json:"otherFiles"`         // 其他文件
}

// ComponentChangeInfo 组件变更信息
type ComponentChangeInfo struct {
	Name  string   `json:"name"`  // 组件名称
	Files []string `json:"files"` // 变更的文件列表
}

// FunctionChangeInfo 函数变更信息
type FunctionChangeInfo struct {
	Name  string   `json:"name"`  // 函数名称（如 utils, hooks 等）
	Files []string `json:"files"` // 变更的文件列表
}

// ComponentImpact 组件影响信息
type ComponentImpact struct {
	ComponentName string `json:"componentName"` // 受影响的组件名称
	ImpactReason  string `json:"impactReason"`  // 影响原因说明
	ChangeType    string `json:"changeType"`    // 变更类型: "component" 或 "function"
	ChangeSource  string `json:"changeSource"`  // 变更来源（组件名或函数路径）
}

// =============================================================================
// 配置相关类型
// =============================================================================

// ComponentManifest 组件清单配置
// 对应 component-manifest.json 文件内容
type ComponentManifest struct {
	Components []ComponentInfo `json:"components"` // 组件列表
	Functions  []FunctionInfo  `json:"functions"`  // 函数列表（可选）
}

// ComponentInfo 组件信息
type ComponentInfo struct {
	Name string `json:"name"` // 组件名称
	Path string `json:"path"` // 组件路径
	Type string `json:"type"` // 类型: "component"
}

// FunctionInfo 函数信息
type FunctionInfo struct {
	Name string `json:"name"` // 函数组名称（如 utils, hooks 等）
	Path string `json:"path"` // 函数目录路径
	Type string `json:"type"` // 类型: "functions"
}

// =============================================================================
// 辅助类型
// =============================================================================

// StringSet 字符串集合，用于去重
type StringSet map[string]struct{}

// Add 添加元素
func (s StringSet) Add(item string) {
	s[item] = struct{}{}
}

// Contains 检查是否包含元素
func (s StringSet) Contains(item string) bool {
	_, ok := s[item]
	return ok
}

// ToSlice 转换为切片
func (s StringSet) ToSlice() []string {
	result := make([]string, 0, len(s))
	for item := range s {
		result = append(result, item)
	}
	return result
}
