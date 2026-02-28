package component_deps

import (
	"path/filepath"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
)

// =============================================================================
// 重导出解析器
// =============================================================================

// ReExportResolver 重导出解析器
// 用于解析 export { xxx } from './path' 这类重导出语句
// 将依赖重定向到真实的源文件，而不是中转的 index.ts
type ReExportResolver struct {
	// fileResults 所有文件的解析结果，用于查找导出声明
	fileResults map[string]projectParser.JsFileParserResult

	// reExportCache 缓存已解析的重导出映射
	// key: "文件路径:导出名称" → value: 真实源文件路径
	reExportCache map[string]string

	// visited 记录访问过的文件，防止循环重导出
	visited map[string]bool
}

// NewReExportResolver 创建重导出解析器
func NewReExportResolver(fileResults map[string]projectParser.JsFileParserResult) *ReExportResolver {
	return &ReExportResolver{
		fileResults:  fileResults,
		reExportCache: make(map[string]string),
		visited:      make(map[string]bool),
	}
}

// ResolveDependency 解析单个依赖，返回其真实的依赖列表
//
// 例如：
//   输入: import { Button, Input } from './exports'
//   输出: [
//     { source: src/components/Button/index.ts, modules: [Button] },
//     { source: src/components/Input/index.ts, modules: [Input] }
//   ]
//
// 如果文件没有重导出，返回原依赖
func (r *ReExportResolver) ResolveDependency(
	dep projectParser.ImportDeclarationResult,
) []projectParser.ImportDeclarationResult {
	// 只处理文件类型的依赖
	if dep.Source.Type != "file" {
		return []projectParser.ImportDeclarationResult{dep}
	}

	targetFile := dep.Source.FilePath
	if targetFile == "" {
		return []projectParser.ImportDeclarationResult{dep}
	}

	// 检查目标文件是否存在
	fileResult, exists := r.fileResults[targetFile]
	if !exists {
		// 文件不存在，返回原依赖
		return []projectParser.ImportDeclarationResult{dep}
	}

	// 解析该文件的重导出映射
	exportMapping := r.buildExportMapping(targetFile, fileResult)

	// 如果没有重导出映射，返回原依赖
	if len(exportMapping) == 0 {
		// 没有重导出，返回原依赖
		return []projectParser.ImportDeclarationResult{dep}
	}

	// 根据导入的模块，重定向到真实的源文件
	resultMap := make(map[string]*projectParser.ImportDeclarationResult)

	for _, importedModule := range dep.ImportModules {
		// 查找该模块的真实源文件
		realSource, found := exportMapping[importedModule.ImportModule]
		if !found {
			// 未找到重导出映射，保留原依赖
			key := r.getDependencyKey(dep)
			if _, exists := resultMap[key]; !exists {
				resultMap[key] = &dep
			}
			continue
		}

		// 找到真实源文件，创建新的依赖
		key := "file:" + realSource
		if existing, exists := resultMap[key]; exists {
			// 已存在该源的依赖，合并 ImportModules
			existing.ImportModules = append(existing.ImportModules, importedModule)
		} else {
			// 创建新依赖
			resultMap[key] = &projectParser.ImportDeclarationResult{
				ImportModules: []projectParser.ImportModule{importedModule},
				Source: projectParser.SourceData{
					Type:     "file",
					FilePath: realSource,
				},
				Raw: dep.Raw + " (重导出自: " + realSource + ")",
			}
		}
	}

	// 转换为切片
	result := make([]projectParser.ImportDeclarationResult, 0, len(resultMap))
	for _, dep := range resultMap {
		result = append(result, *dep)
	}

	return result
}

// buildExportMapping 构建文件的重导出映射
// 返回: map[导出名称]真实源文件路径
// 例如: {"Button": "src/components/Button/index.ts", "Input": "src/components/Input/index.ts"}
func (r *ReExportResolver) buildExportMapping(
	_ string, // filePath 保留参数以便未来扩展
	fileResult projectParser.JsFileParserResult,
) map[string]string {
	mapping := make(map[string]string)

	// 遍历所有导出声明
	for _, exportDecl := range fileResult.ExportDeclarations {
		// 只处理重导出（Source 不为 nil）
		if exportDecl.Source == nil {
			continue
		}

		// 只处理文件类型的重导出
		if exportDecl.Source.Type != "file" {
			continue
		}

		sourceFile := exportDecl.Source.FilePath
		if sourceFile == "" {
			continue
		}

		// 构建导出映射
		for _, exportedModule := range exportDecl.ExportModules {
			// 使用 Identifier（外部名称）作为 key，而不是 ModuleName（原始名称）
			// 这样可以正确处理 export { default as Popcard } 这类语法
			// exportName 是外部导入时使用的名称
			var exportName string
			if exportedModule.Identifier != "" && exportedModule.Identifier != exportedModule.ModuleName {
				// 有别名的情况：export { xxx as yyy } 或 export { default as yyy }
				exportName = exportedModule.Identifier
			} else if exportedModule.ModuleName == "default" {
				// export default xxx 或 export { default } from './xxx'
				exportName = "default"
			} else {
				// 普通导出：export { Button }
				exportName = exportedModule.ModuleName
			}

			// 处理 export * 的情况
			if exportedModule.ModuleName == "*" {
				// 递归解析被导出的文件
				sourceFileResult, exists := r.fileResults[sourceFile]
				if exists {
					// 防止循环依赖
					if r.visited[sourceFile] {
						continue
					}
					r.visited[sourceFile] = true

					// 递归获取子映射
					subMapping := r.buildExportMapping(sourceFile, sourceFileResult)
					for k, v := range subMapping {
						mapping[k] = v
					}
				}
				continue
			}

			// 建立映射：外部名称 → 真实源文件
			mapping[exportName] = sourceFile
		}
	}

	return mapping
}

// getDependencyKey 获取依赖的唯一标识
func (r *ReExportResolver) getDependencyKey(dep projectParser.ImportDeclarationResult) string {
	switch dep.Source.Type {
	case "npm":
		return "npm:" + dep.Source.NpmPkg
	case "file":
		return "file:" + dep.Source.FilePath
	default:
		return ""
	}
}

// ResolveDependencies 批量解析依赖列表
func (r *ReExportResolver) ResolveDependencies(
	dependencies []projectParser.ImportDeclarationResult,
) []projectParser.ImportDeclarationResult {
	result := make([]projectParser.ImportDeclarationResult, 0, len(dependencies))

	for _, dep := range dependencies {
		resolved := r.ResolveDependency(dep)
		result = append(result, resolved...)
	}

	return result
}

// IsReExportFile 判断文件是否包含重导出语句
func (r *ReExportResolver) IsReExportFile(filePath string) bool {
	fileResult, exists := r.fileResults[filePath]
	if !exists {
		return false
	}

	for _, exportDecl := range fileResult.ExportDeclarations {
		if exportDecl.Source != nil && exportDecl.Source.Type == "file" {
			return true
		}
	}

	return false
}

// GetReExportChain 获取重导出链（用于调试）
// 例如: ["src/exports/index.ts", "src/components/Button/index.ts"]
func (r *ReExportResolver) GetReExportChain(filePath string) []string {
	chain := []string{}
	visited := make(map[string]bool)

	current := filePath
	for current != "" && !visited[current] {
		visited[current] = true
		chain = append(chain, current)

		if !r.IsReExportFile(current) {
			break
		}

		// 获取下一个文件
		fileResult := r.fileResults[current]
		for _, exportDecl := range fileResult.ExportDeclarations {
			if exportDecl.Source != nil && exportDecl.Source.Type == "file" {
				current = exportDecl.Source.FilePath
				break
			}
		}
	}

	return chain
}

// NormalizeFilePath 标准化文件路径
func (r *ReExportResolver) NormalizeFilePath(path string) string {
	return filepath.ToSlash(path)
}

// FindExportByModule 在文件中查找指定模块的导出
// 返回该模块的真实源文件路径
func (r *ReExportResolver) FindExportByModule(filePath, moduleName string) (string, bool) {
	fileResult, exists := r.fileResults[filePath]
	if !exists {
		return "", false
	}

	for _, exportDecl := range fileResult.ExportDeclarations {
		if exportDecl.Source == nil || exportDecl.Source.Type != "file" {
			continue
		}

		for _, exportedModule := range exportDecl.ExportModules {
			if exportedModule.ModuleName == moduleName {
				return exportDecl.Source.FilePath, true
			}
		}
	}

	return "", false
}

// GetAllReExports 获取所有重导出文件列表
func (r *ReExportResolver) GetAllReExports() []string {
	result := make([]string, 0)

	for filePath := range r.fileResults {
		if r.IsReExportFile(filePath) {
			result = append(result, filePath)
		}
	}

	return result
}

// GetReExportMapping 获取文件的重导出映射
func (r *ReExportResolver) GetReExportMapping(filePath string) map[string]string {
	fileResult, exists := r.fileResults[filePath]
	if !exists {
		return nil
	}

	return r.buildExportMapping(filePath, fileResult)
}

// ClearCache 清空缓存
func (r *ReExportResolver) ClearCache() {
	r.reExportCache = make(map[string]string)
	r.visited = make(map[string]bool)
}

// GetStats 获取重导出解析统计信息
func (r *ReExportResolver) GetStats() ReExportStats {
	totalFiles := len(r.fileResults)
	reExportFiles := 0
	totalMappings := 0

	for filePath := range r.fileResults {
		if r.IsReExportFile(filePath) {
			reExportFiles++
			mapping := r.GetReExportMapping(filePath)
			totalMappings += len(mapping)
		}
	}

	return ReExportStats{
		TotalFiles:    totalFiles,
		ReExportFiles: reExportFiles,
		TotalMappings: totalMappings,
	}
}

// ReExportStats 重导出统计信息
type ReExportStats struct {
	TotalFiles    int // 总文件数
	ReExportFiles int // 包含重导出的文件数
	TotalMappings int // 重导出映射总数
}
