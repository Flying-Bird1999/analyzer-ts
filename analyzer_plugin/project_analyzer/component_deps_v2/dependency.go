package component_deps_v2

import (
	"path/filepath"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
)

// =============================================================================
// 依赖分析器
// =============================================================================

// DependencyAnalyzer 组件依赖分析器
type DependencyAnalyzer struct {
	manifest *ComponentManifest // 组件配置
}

// NewDependencyAnalyzer 创建依赖分析器
func NewDependencyAnalyzer(manifest *ComponentManifest) *DependencyAnalyzer {
	return &DependencyAnalyzer{
		manifest: manifest,
	}
}

// AnalyzeComponent 分析单个组件的外部依赖
// 返回该组件的所有外部依赖（过滤掉组件内部依赖，并去重）
func (da *DependencyAnalyzer) AnalyzeComponent(
	comp *ComponentDefinition,
	fileResults map[string]projectParser.JsFileParserResult,
) []projectParser.ImportDeclarationResult {
	// 直接使用目录路径
	compDir := comp.Path

	// 使用 map 去重：key -> ImportDeclarationResult
	// npm 包: key = "npm:" + npmPkg
	// 文件: key = "file:" + filePath
	seen := make(map[string]projectParser.ImportDeclarationResult)

	// 遍历所有文件
	for sourceFile, fileResult := range fileResults {
		// 检查源文件是否属于当前组件
		if !da.isFileInComponent(sourceFile, compDir) {
			continue
		}

		// 遍历该文件的所有导入
		for _, importDecl := range fileResult.ImportDeclarations {
			// 判断是否为外部依赖
			if !da.isExternalDependency(importDecl, compDir) {
				continue
			}

			// 计算去重 key
			key := da.getDependencyKey(importDecl)
			if key == "" {
				continue
			}

			// 如果该依赖尚未记录，则添加
			if _, exists := seen[key]; !exists {
				seen[key] = importDecl
			}
		}
	}

	// 转换为切片返回
	externalDeps := make([]projectParser.ImportDeclarationResult, 0, len(seen))
	for _, dep := range seen {
		externalDeps = append(externalDeps, dep)
	}

	return externalDeps
}

// getDependencyKey 获取依赖的唯一标识，用于去重
func (da *DependencyAnalyzer) getDependencyKey(importDecl projectParser.ImportDeclarationResult) string {
	switch importDecl.Source.Type {
	case "npm":
		return "npm:" + importDecl.Source.NpmPkg
	case "file":
		return "file:" + importDecl.Source.FilePath
	default:
		return ""
	}
}

// isFileInComponent 判断文件是否在组件目录下
func (da *DependencyAnalyzer) isFileInComponent(filePath, compDir string) bool {
	// 标准化路径为正斜杠格式
	normalizedDir := filepath.ToSlash(compDir)
	normalizedPath := filepath.ToSlash(filePath)

	// 首先尝试精确前缀匹配（处理相对路径情况）
	if strings.HasPrefix(normalizedPath, normalizedDir+"/") || normalizedPath == normalizedDir {
		return true
	}

	// 如果是绝对路径，尝试提取相对路径部分后再匹配
	// 例如: /project/src/Button/xxx.tsx → src/Button/xxx.tsx
	parts := strings.Split(normalizedPath, "/")
	for i := 0; i < len(parts); i++ {
		// 尝试从每个位置开始，看是否能匹配组件目录
		candidatePath := strings.Join(parts[i:], "/")
		if strings.HasPrefix(candidatePath, normalizedDir+"/") || candidatePath == normalizedDir {
			return true
		}
	}

	return false
}

// isExternalDependency 判断是否为外部依赖
// 规则：
// 1. npm 包 → 外部依赖
// 2. 文件类型：目标文件不在当前组件目录下 → 外部依赖
// 3. 文件类型：目标文件在当前组件目录下 → 内部依赖（忽略）
func (da *DependencyAnalyzer) isExternalDependency(
	importDecl projectParser.ImportDeclarationResult,
	sourceCompDir string,
) bool {
	// npm 包，直接视为外部依赖
	if importDecl.Source.Type == "npm" {
		return true
	}

	// 文件类型，判断目标文件是否在当前组件目录下
	if importDecl.Source.Type == "file" {
		targetFilePath := importDecl.Source.FilePath
		if targetFilePath == "" {
			return false
		}

		// 检查目标文件是否在当前组件目录下
		if da.isFileInComponent(targetFilePath, sourceCompDir) {
			// 在同一组件内，是内部依赖
			return false
		}

		// 不在当前组件目录下，是外部依赖
		return true
	}

	// 未知类型，不处理
	return false
}

// AnalyzeAllComponents 分析所有组件的外部依赖
// 返回组件名 -> 外部依赖列表的映射
func (da *DependencyAnalyzer) AnalyzeAllComponents(
	fileResults map[string]projectParser.JsFileParserResult,
) map[string][]projectParser.ImportDeclarationResult {
	result := make(map[string][]projectParser.ImportDeclarationResult)

	for i := range da.manifest.Components {
		comp := &da.manifest.Components[i]
		deps := da.AnalyzeComponent(comp, fileResults)
		result[comp.Name] = deps
	}

	return result
}
