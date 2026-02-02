// Package file_analyzer 提供文件级影响分析功能。
// 这是通用的能力，适用于所有前端项目，不依赖 component-manifest.json。
package file_analyzer

import (
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
)

// =============================================================================
// 文件依赖图构建器（用于 component_analyzer）
// =============================================================================

// GraphBuilder 文件依赖图构建器
// 通过解析 Import 声明构建文件间的依赖关系图
// 注意：这是一个独立的工具，主要用于 component_analyzer 构建组件级依赖
type GraphBuilder struct {
	// 项目解析结果，包含所有文件的导入/导出信息
	parsingResult *projectParser.ProjectParserResult
}

// NewGraphBuilder 创建文件依赖图构建器
func NewGraphBuilder(parsingResult *projectParser.ProjectParserResult) *GraphBuilder {
	return &GraphBuilder{
		parsingResult: parsingResult,
	}
}

// FileDependencyGraph 文件依赖图
// 记录项目内文件之间的依赖关系
type FileDependencyGraph struct {
	// DepGraph 正向依赖图：文件 → 依赖的文件
	DepGraph map[string][]string

	// RevDepGraph 反向依赖图：文件 → 被依赖的文件（下游文件）
	RevDepGraph map[string][]string

	// ExternalDeps 外部依赖：文件 → npm 包列表
	ExternalDeps map[string][]string
}

// BuildFileDependencyGraph 构建文件依赖图
func (b *GraphBuilder) BuildFileDependencyGraph() *FileDependencyGraph {
	graph := &FileDependencyGraph{
		DepGraph:     make(map[string][]string),
		RevDepGraph:  make(map[string][]string),
		ExternalDeps: make(map[string][]string),
	}

	if b.parsingResult == nil {
		return graph
	}

	// 遍历所有已解析的 JS/TS 文件
	for sourceFile, fileResult := range b.parsingResult.Js_Data {
		// 处理该文件的导入声明
		for _, importDecl := range fileResult.ImportDeclarations {
			b.processImport(sourceFile, importDecl, graph)
		}
	}

	// 构建反向依赖图
	b.buildReverseGraph(graph)

	return graph
}

// processImport 处理单个导入声明
func (b *GraphBuilder) processImport(
	sourceFile string,
	importDecl projectParser.ImportDeclarationResult,
	graph *FileDependencyGraph,
) {
	// 检查导入来源是否为项目内文件
	// importDecl.Source.FilePath 是解析后的绝对路径
	if importDecl.Source.FilePath == "" {
		// 可能是 npm 包或其他外部依赖
		b.recordExternalDependency(sourceFile, importDecl, graph)
		return
	}

	// 内部文件依赖：添加到依赖图
	// importDecl.Source.FilePath 已经是绝对路径
	graph.DepGraph[sourceFile] = appendUnique(graph.DepGraph[sourceFile], importDecl.Source.FilePath)
}

// recordExternalDependency 记录外部依赖（npm 包）
func (b *GraphBuilder) recordExternalDependency(
	sourceFile string,
	importDecl projectParser.ImportDeclarationResult,
	graph *FileDependencyGraph,
) {
	// 从导入路径提取包名
	packageName := b.extractPackageName(importDecl)
	if packageName != "" {
		graph.ExternalDeps[sourceFile] = appendUnique(graph.ExternalDeps[sourceFile], packageName)
	}
}

// extractPackageName 从导入路径提取包名
func (b *GraphBuilder) extractPackageName(importDecl projectParser.ImportDeclarationResult) string {
	// 获取第一个导入模块来推断包名
	for _, module := range importDecl.ImportModules {
		importPath := module.ImportModule
		if importPath == "" {
			continue
		}

		// 跳过相对路径
		if strings.HasPrefix(importPath, ".") {
			return ""
		}

		// 提取包名（例如：@scope/name 或 name）
		// 取路径的第一段
		parts := strings.Split(importPath, "/")
		if len(parts) > 0 {
			return parts[0]
		}
	}

	return ""
}

// buildReverseGraph 构建反向依赖图
func (b *GraphBuilder) buildReverseGraph(graph *FileDependencyGraph) {
	for sourceFile, targets := range graph.DepGraph {
		for _, target := range targets {
			graph.RevDepGraph[target] = appendUnique(graph.RevDepGraph[target], sourceFile)
		}
	}
}

// appendUnique 唯一添加元素到切片
func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}

// GetDependants 获取依赖指定文件的所有文件
func (g *FileDependencyGraph) GetDependants(filePath string) []string {
	return g.RevDepGraph[filePath]
}

// GetDependencies 获取指定文件依赖的所有文件
func (g *FileDependencyGraph) GetDependencies(filePath string) []string {
	return g.DepGraph[filePath]
}

// GetExternalDeps 获取指定文件的外部依赖
func (g *FileDependencyGraph) GetExternalDeps(filePath string) []string {
	return g.ExternalDeps[filePath]
}
