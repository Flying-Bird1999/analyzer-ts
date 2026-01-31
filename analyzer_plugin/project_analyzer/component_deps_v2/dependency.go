package component_deps_v2

import (
	"path/filepath"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/samber/lo"
)

// =============================================================================
// 依赖分析器
// =============================================================================

// DependencyAnalyzer 组件依赖分析器
type DependencyAnalyzer struct {
	scope    *MultiComponentScope // 组件作用域管理
	manifest *ComponentManifest  // 组件配置
	projectRoot string           // 项目根目录
}

// NewDependencyAnalyzer 创建依赖分析器
func NewDependencyAnalyzer(manifest *ComponentManifest, scope *MultiComponentScope, projectRoot string) *DependencyAnalyzer {
	return &DependencyAnalyzer{
		scope:    scope,
		manifest: manifest,
		projectRoot: projectRoot,
	}
}

// AnalyzeComponent 分析单个组件的依赖关系
// 返回该组件依赖的其他组件列表
func (da *DependencyAnalyzer) AnalyzeComponent(
	comp *ComponentDefinition,
	fileResults map[string]projectParser.JsFileParserResult,
) []string {
	// 获取所有文件路径，然后筛选属于该组件的文件
	allFiles := getFilePaths(fileResults)
	componentFiles := lo.Filter(allFiles, func(path string, _ int) bool {
		// 检查文件是否在该组件的作用域内
		compName, _ := da.scope.FindComponentByFile(path)
		return compName == comp.Name
	})

	// 分析依赖：查找跨组件导入
	dependencies := make(map[string]bool)
	for _, filePath := range componentFiles {
		fileResult, ok := fileResults[filePath]
		if !ok {
			continue
		}

		// 遍历该文件的所有导入
		for _, importDecl := range fileResult.ImportDeclarations {
			importPath := importDecl.Source.FilePath
			if importPath == "" {
				continue
			}

			// 将相对路径解析为绝对路径
			resolvedPath := da.resolveImportPath(importPath, filePath)

			// 检查是否为跨组件导入
			targetComp, isCross, isExternal := da.scope.DetectCrossComponentImports(
				resolvedPath, filePath)

			// 只记录跨组件的依赖
			if isCross && !isExternal && targetComp != "" {
				dependencies[targetComp] = true
			}
		}
	}

	// 返回依赖列表（去重后）
	return lo.Keys(dependencies)
}

// resolveImportPath 解析导入路径
// 如果是相对路径，基于源文件所在目录解析为绝对路径
// 如果是绝对路径或 node_modules，直接返回
func (da *DependencyAnalyzer) resolveImportPath(importPath, sourceFilePath string) string {
	// 如果是相对路径
	if isRelativePath(importPath) {
		// 获取源文件所在目录
		sourceDir := filepath.Dir(sourceFilePath)
		// 拼接相对路径得到绝对路径
		resolved := filepath.Join(sourceDir, importPath)
		// 标准化路径（移除 .. 和 .）
		resolved = filepath.Clean(resolved)
		// 转换为正斜杠（Windows 兼容）
		resolved = filepath.ToSlash(resolved)

		// 如果是绝对路径，去掉项目根目录前缀，得到相对于项目根的路径
		if filepath.IsAbs(resolved) && len(resolved) >= len(da.projectRoot) {
			relativeToRoot := resolved
			if len(resolved) > len(da.projectRoot) && resolved[len(da.projectRoot)] == '/' {
				relativeToRoot = resolved[len(da.projectRoot)+1:]
			} else if resolved == da.projectRoot {
				relativeToRoot = "."
			}
			return relativeToRoot
		}

		return resolved
	}

	// 如果不是相对路径（如 node_modules 或绝对路径），直接返回
	return importPath
}

// isRelativePath 检查是否为相对路径
func isRelativePath(path string) bool {
	return len(path) >= 3 && (path[0] == '.' && (path[1] == '/' || (path[1] == '.' && path[2] == '/')))
}

// =============================================================================
// 辅助函数
// =============================================================================

// getFilePaths 从解析结果中提取所有文件路径
func getFilePaths(fileResults map[string]projectParser.JsFileParserResult) []string {
	paths := make([]string, 0, len(fileResults))
	for path := range fileResults {
		paths = append(paths, path)
	}
	return paths
}

// AnalyzeAllComponents 分析所有组件的依赖关系
// 返回组件名 -> 依赖列表的映射
func (da *DependencyAnalyzer) AnalyzeAllComponents(
	fileResults map[string]projectParser.JsFileParserResult,
) map[string][]string {
	result := make(map[string][]string)

	for i := range da.manifest.Components {
		comp := &da.manifest.Components[i]
		deps := da.AnalyzeComponent(comp, fileResults)
		result[comp.Name] = deps
	}

	return result
}
