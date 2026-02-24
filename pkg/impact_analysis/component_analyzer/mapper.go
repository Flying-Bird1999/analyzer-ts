// Package component_analyzer 提供组件级影响分析功能。
// 这是组件库专用能力，基于 file_analyzer 的结果进行组件映射。
package component_analyzer

import (
	"path/filepath"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/Flying-Bird1999/analyzer-ts/pkg/impact_analysis"
)

// =============================================================================
// 组件映射器
// =============================================================================

// ComponentMapper 组件映射器
// 负责将文件映射到组件，并构建组件级依赖关系
type ComponentMapper struct {
	componentManifest *impact_analysis.ComponentManifest
}

// NewComponentMapper 创建组件映射器
func NewComponentMapper(manifest *impact_analysis.ComponentManifest) *ComponentMapper {
	return &ComponentMapper{
		componentManifest: manifest,
	}
}

// =============================================================================
// 文件到组件映射
// =============================================================================

// MapFileToComponent 将文件路径映射到组件名
// 返回组件名，如果文件不属于任何组件则返回空字符串
func (m *ComponentMapper) MapFileToComponent(filePath string) string {
	if m.componentManifest == nil {
		return ""
	}

	// 遍历组件清单，检查文件是否在组件范围内
	// 使用 path 作为组件作用域
	for _, comp := range m.componentManifest.Components {
		// 检查文件是否在组件目录下
		if strings.HasPrefix(filePath, comp.Path) {
			return comp.Name
		}
	}

	return ""
}

// MapFilesToComponents 批量映射文件到组件
// 返回: map[文件路径]组件名
func (m *ComponentMapper) MapFilesToComponents(filePaths []string) map[string]string {
	result := make(map[string]string)
	for _, path := range filePaths {
		result[path] = m.MapFileToComponent(path)
	}
	return result
}

// GetComponentByEntry 根据 entry 获取组件
func (m *ComponentMapper) GetComponentByEntry(entry string) *impact_analysis.Component {
	if m.componentManifest == nil {
		return nil
	}

	for _, comp := range m.componentManifest.Components {
		// 将 path 转换为 entry 格式进行比较
		componentEntry := filepath.Join(comp.Path, "index.tsx")
		if componentEntry == entry {
			return &comp
		}
	}
	return nil
}

// GetComponentByName 根据名称获取组件
func (m *ComponentMapper) GetComponentByName(name string) *impact_analysis.Component {
	if m.componentManifest == nil {
		return nil
	}

	for _, comp := range m.componentManifest.Components {
		if comp.Name == name {
			return &comp
		}
	}
	return nil
}

// GetComponentEntry 获取组件的 entry 路径
func (m *ComponentMapper) GetComponentEntry(componentName string) string {
	comp := m.GetComponentByName(componentName)
	if comp == nil {
		return ""
	}
	// 从 path 构造 entry 路径
	return filepath.Join(comp.Path, "index.tsx")
}

// GetComponentFiles 获取组件的所有文件
// 基于 file_analyzer 的结果，筛选出属于指定组件的文件
func (m *ComponentMapper) GetComponentFiles(componentName string, fileGraph *FileDependencyGraphProxy) []string {
	if fileGraph == nil {
		return []string{}
	}

	componentEntry := m.GetComponentEntry(componentName)
	if componentEntry == "" {
		return []string{}
	}

	componentDir := filepath.Dir(componentEntry)
	files := make([]string, 0)

	// 从正向依赖图中获取所有文件
	for filePath := range fileGraph.DepGraph {
		if strings.HasPrefix(filePath, componentDir) {
			files = append(files, filePath)
		}
	}

	// 从反向依赖图中获取所有文件（避免遗漏）
	for filePath := range fileGraph.RevDepGraph {
		if strings.HasPrefix(filePath, componentDir) {
			// 去重
			found := false
			for _, f := range files {
				if f == filePath {
					found = true
					break
				}
			}
			if !found {
				files = append(files, filePath)
			}
		}
	}

	return files
}

// =============================================================================
// 组件依赖图构建
// =============================================================================

// ComponentDependencyGraph 组件依赖图
type ComponentDependencyGraph struct {
	// DepGraph 正向依赖图：组件 → 依赖的组件
	DepGraph map[string][]string

	// RevDepGraph 反向依赖图：组件 → 被依赖的组件（下游组件）
	RevDepGraph map[string][]string

	// SymbolImports 组件间的符号导入关系
	// map[组件名][]符号导入关系
	SymbolImports map[string][]ComponentSymbolImport
}

// ComponentSymbolImport 组件符号导入
type ComponentSymbolImport struct {
	// SourceComponent 来源组件（或外部依赖名称）
	SourceComponent string

	// ImportedSymbols 导入的符号列表
	ImportedSymbols []impact_analysis.SymbolRef

	// ImportType 导入类型: named/default/namespace
	ImportType string

	// SourceFiles 涉及的源文件列表
	SourceFiles []string
}

// BuildComponentDependencyGraph 构建组件依赖图
// 基于文件依赖图和组件清单，将文件级依赖聚合为组件级依赖
func (m *ComponentMapper) BuildComponentDependencyGraph(
	fileGraph *FileDependencyGraphProxy,
	parsingResult *projectParser.ProjectParserResult,
) *ComponentDependencyGraph {
	graph := &ComponentDependencyGraph{
		DepGraph:      make(map[string][]string),
		RevDepGraph:   make(map[string][]string),
		SymbolImports: make(map[string][]ComponentSymbolImport),
	}

	if fileGraph == nil {
		return graph
	}

	// 步骤 1: 遍历文件级依赖，构建组件级依赖
	for sourceFile, targetFiles := range fileGraph.DepGraph {
		// 获取源文件所属组件
		sourceComponent := m.MapFileToComponent(sourceFile)
		if sourceComponent == "" {
			continue // 源文件不属于任何组件，跳过
		}

		for _, targetFile := range targetFiles {
			// 获取目标文件所属组件
			targetComponent := m.MapFileToComponent(targetFile)
			if targetComponent == "" {
				continue // 目标文件不属于任何组件，跳过
			}

			// 跳过组件内部依赖
			if sourceComponent == targetComponent {
				continue
			}

			// 添加跨组件依赖（去重）
			graph.DepGraph[sourceComponent] = appendUnique(graph.DepGraph[sourceComponent], targetComponent)
		}
	}

	// 步骤 2: 构建反向依赖图
	m.buildReverseComponentGraph(graph)

	// 步骤 3: 构建符号导入关系（仅在 parsingResult 非空时）
	if parsingResult != nil {
		m.buildSymbolImports(graph, parsingResult)
	}

	return graph
}

// buildReverseComponentGraph 构建反向组件依赖图
func (m *ComponentMapper) buildReverseComponentGraph(graph *ComponentDependencyGraph) {
	for sourceComp, targetComps := range graph.DepGraph {
		for _, target := range targetComps {
			graph.RevDepGraph[target] = appendUnique(graph.RevDepGraph[target], sourceComp)
		}
	}
}

// buildSymbolImports 构建组件间的符号导入关系
func (m *ComponentMapper) buildSymbolImports(
	graph *ComponentDependencyGraph,
	parsingResult *projectParser.ProjectParserResult,
) {
	if parsingResult == nil {
		return
	}

	// 遍历所有已解析的 JS/TS 文件
	for filePath, fileResult := range parsingResult.Js_Data {
		// 获取文件所属组件
		currentComponent := m.MapFileToComponent(filePath)
		if currentComponent == "" {
			continue
		}

		// 处理该文件的导入声明
		for _, importDecl := range fileResult.ImportDeclarations {
			m.processImportForSymbols(importDecl, currentComponent, filePath, graph)
		}
	}
}

// processImportForSymbols 处理导入声明，构建符号导入关系
func (m *ComponentMapper) processImportForSymbols(
	importDecl projectParser.ImportDeclarationResult,
	currentComponent string,
	sourceFile string,
	graph *ComponentDependencyGraph,
) {
	// 检查导入来源是否为项目内文件（非 NPM 包）
	if importDecl.Source.FilePath == "" {
		// 可能是 NPM 包或其他外部依赖
		return
	}

	// 根据解析后的路径确定来源组件
	sourceComponent := m.MapFileToComponent(importDecl.Source.FilePath)

	// 如果来源文件不属于任何组件，或者是当前组件内部，跳过
	if sourceComponent == "" || sourceComponent == currentComponent {
		return
	}

	// 构建符号导入关系
	symbolImport := ComponentSymbolImport{
		SourceComponent: sourceComponent,
		ImportedSymbols: make([]impact_analysis.SymbolRef, 0),
		ImportType:      m.determineImportType(importDecl),
		SourceFiles:     []string{sourceFile},
	}

	// 提取导入的符号
	for _, module := range importDecl.ImportModules {
		symbolImport.ImportedSymbols = append(symbolImport.ImportedSymbols, impact_analysis.SymbolRef{
			Name:       module.Identifier,
			Kind:       m.inferSymbolKind(module),
			FilePath:   importDecl.Source.FilePath,
			ExportType: m.inferExportType(module.Type),
		})
	}

	// 记录符号导入关系
	graph.SymbolImports[currentComponent] = append(
		graph.SymbolImports[currentComponent],
		symbolImport,
	)
}

// determineImportType 确定导入类型
func (m *ComponentMapper) determineImportType(importDecl projectParser.ImportDeclarationResult) string {
	// 检查是否有命名空间导入
	for _, module := range importDecl.ImportModules {
		if module.Type == "namespace" {
			return "namespace"
		}
		if module.Type == "default" {
			return "default"
		}
	}
	return "named"
}

// inferSymbolKind 推断符号类型
func (m *ComponentMapper) inferSymbolKind(module projectParser.ImportModule) impact_analysis.SymbolKind {
	// TODO: 可以通过查看源文件中该符号的实际类型来推断
	// 当前返回默认类型
	return impact_analysis.SymbolKindVariable
}

// inferExportType 推断导出类型
func (m *ComponentMapper) inferExportType(importType string) impact_analysis.ExportType {
	switch importType {
	case "default":
		return impact_analysis.ExportTypeDefault
	case "namespace":
		return impact_analysis.ExportTypeNamespace
	default:
		return impact_analysis.ExportTypeNamed
	}
}

// =============================================================================
// FileDependencyGraphProxy 代理类型
// =============================================================================

// FileDependencyGraphProxy 文件依赖图代理
// 用于解耦 component_analyzer 对 file_analyzer 的直接依赖
type FileDependencyGraphProxy struct {
	DepGraph     map[string][]string
	RevDepGraph  map[string][]string
	ExternalDeps map[string][]string
}

// NewFileDependencyGraphProxy 从 file_analyzer 的图创建代理
func NewFileDependencyGraphProxy(depGraph, revDepGraph, externalDeps map[string][]string) *FileDependencyGraphProxy {
	return &FileDependencyGraphProxy{
		DepGraph:     depGraph,
		RevDepGraph:  revDepGraph,
		ExternalDeps: externalDeps,
	}
}

// =============================================================================
// 辅助函数
// =============================================================================

// appendUnique 唯一添加元素到切片
func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}
