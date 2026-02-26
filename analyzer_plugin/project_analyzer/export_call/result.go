// Package export_call 数据结构定义
package export_call

import (
	"bytes"
	"fmt"
	"sort"

	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

// NodeType 导出节点类型（按内容分类）
type NodeType string

const (
	NodeTypeFunction  NodeType = "function"
	NodeTypeVariable  NodeType = "variable"
	NodeTypeType      NodeType = "type"
	NodeTypeInterface NodeType = "interface"
	NodeTypeEnum      NodeType = "enum"
)

// ExportType 导出方式（按导出语法分类）
type ExportType string

const (
	ExportTypeNamed   ExportType = "named"
	ExportTypeDefault ExportType = "default"
)

// ExportNode 导出节点信息
type ExportNode struct {
	ID         string     `json:"id"`         // "assetName:nodeName:exportType"
	Name       string     `json:"name"`       // 节点名称
	AssetName  string     `json:"assetName"`  // 所属资产
	NodeType   NodeType   `json:"nodeType"`   // 节点类型
	ExportType ExportType `json:"exportType"` // 导出方式
	SourceFile string     `json:"sourceFile"` // 定义所在文件
}

// FileExportRecord 文件导出记录
type FileExportRecord struct {
	File  string         `json:"file"`  // 文件路径
	Nodes []NodeWithRefs `json:"nodes"` // 该文件的导出节点
}

// ComponentRef 组件级引用信息
// 表示某个组件引用了该导出节点
type ComponentRef struct {
	ComponentName string   `json:"componentName"` // 组件名称
	RefFiles      []string `json:"refFiles"`      // 该组件中引用该节点的文件列表
}

// NodeWithRefs 带引用信息的节点
type NodeWithRefs struct {
	Name           string         `json:"name"`           // 节点名称
	NodeType       NodeType       `json:"nodeType"`       // 节点类型
	ExportType     ExportType     `json:"exportType"`     // 导出方式
	RefFiles       []string       `json:"refFiles"`       // 引用该节点的文件路径列表
	RefComponents  []ComponentRef `json:"refComponents"`  // 引用该节点的组件列表（基于 manifest 中的 components 配置）
}

// ModuleExportRecord 模块导出记录（按模块分组）
type ModuleExportRecord struct {
	ModuleName string             `json:"moduleName"` // 模块名称
	Path       string             `json:"path"`       // 资产配置路径
	Files      []FileExportRecord `json:"files"`      // 该模块的文件列表
}

// ExportCallResult 分析结果
type ExportCallResult struct {
	ModuleExports []ModuleExportRecord `json:"moduleExports"` // 按模块分组的导出记录
}

// =============================================================================
// Result 接口实现
// =============================================================================

// Name 返回分析结果标识符
func (r *ExportCallResult) Name() string {
	return "export-call"
}

// Summary 返回分析结果摘要
func (r *ExportCallResult) Summary() string {
	totalModules := len(r.ModuleExports)
	totalFiles := 0
	totalNodes := 0
	totalRefs := 0
	for _, m := range r.ModuleExports {
		totalFiles += len(m.Files)
		for _, f := range m.Files {
			totalNodes += len(f.Nodes)
			for _, n := range f.Nodes {
				totalRefs += len(n.RefFiles)
			}
		}
	}
	return fmt.Sprintf("分析完成，共 %d 个模块，%d 个文件，%d 个导出节点，%d 条引用关系。",
		totalModules, totalFiles, totalNodes, totalRefs)
}

// ToJSON 将结果序列化为 JSON
func (r *ExportCallResult) ToJSON(indent bool) ([]byte, error) {
	return projectanalyzer.ToJSONBytes(r, indent)
}

// ToConsole 将结果格式化为控制台输出
func (r *ExportCallResult) ToConsole() string {
	var buffer bytes.Buffer

	// 标题
	buffer.WriteString("=====================================\n")
	buffer.WriteString("导出节点引用关系分析报告\n")
	buffer.WriteString("=====================================\n\n")

	// 按模块显示
	for _, module := range r.ModuleExports {
		buffer.WriteString(fmt.Sprintf("━━━━ 模块: %s ━━━\n", module.ModuleName))
		buffer.WriteString(fmt.Sprintf("  路径: %s\n\n", module.Path))

		// 按文件排序
		sortedFiles := make([]string, 0, len(module.Files))
		for _, record := range module.Files {
			sortedFiles = append(sortedFiles, record.File)
		}
		sort.Strings(sortedFiles)

		// 文件详情
		for _, file := range sortedFiles {
			for _, record := range module.Files {
				if record.File != file {
					continue
				}

				// 显示相对路径（去掉模块前缀，更简洁）
				relPath := record.File
				if idx := indexOf(relPath, module.ModuleName); idx >= 0 {
					relPath = relPath[idx:] // 保留模块名称开始的部分
				}

				buffer.WriteString(fmt.Sprintf("  ▶ %s\n", relPath))
				for _, node := range record.Nodes {
					refCount := len(node.RefFiles)
					if refCount > 0 {
						buffer.WriteString(fmt.Sprintf("    - %s [%s, %s] 被引用 %d 次\n",
							node.Name, node.NodeType, node.ExportType, refCount))
					} else {
						buffer.WriteString(fmt.Sprintf("    - %s [%s, %s] 未被引用\n",
							node.Name, node.NodeType, node.ExportType))
					}

					// 显示组件级引用
					if len(node.RefComponents) > 0 {
						buffer.WriteString("      影响组件:\n")
						for _, compRef := range node.RefComponents {
							buffer.WriteString(fmt.Sprintf("        • %s (%d 个文件)\n",
								compRef.ComponentName, len(compRef.RefFiles)))
						}
					}
				}
				buffer.WriteString("\n")
			}
		}
	}

	return buffer.String()
}

// indexOf 查找子字符串首次出现的位置
func indexOf(s, substr string) int {
	idx := sort.SearchStrings([]string{substr}, s)
	if idx < 0 {
		return -1
	}
	// 验证找到的位置是否真的包含子串
	if len(s) < len(substr) || s[idx:idx+len(substr)] != substr {
		return -1
	}
	return idx
}

// AnalyzerName 返回对应的分析器名称
func (r *ExportCallResult) AnalyzerName() string {
	return "export-call"
}
