// Package file_analyzer 提供文件级影响分析功能。
// 这是通用的能力，适用于所有前端项目，不依赖 component-manifest.json。
package file_analyzer

import (
	"container/list"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/Flying-Bird1999/analyzer-ts/pkg/symbol_analysis"
)

// =============================================================================
// 符号级影响传播器
// =============================================================================

// SymbolPropagator 符号级影响传播器
// 基于符号的 import/export 关系传播影响
type SymbolPropagator struct {
	parsingResult *projectParser.ProjectParserResult
}

// NewSymbolPropagator 创建符号影响传播器
func NewSymbolPropagator(parsingResult *projectParser.ProjectParserResult) *SymbolPropagator {
	return &SymbolPropagator{
		parsingResult: parsingResult,
	}
}

// ImpactedFiles 受影响的文件集合
type ImpactedFiles struct {
	Direct   map[string]*FileImpact // 直接变更的文件
	Indirect map[string]*FileImpact // 间接受影响的文件
}

// FileImpact 文件影响信息
type FileImpact struct {
	FilePath    string   // 文件路径
	ImpactLevel int      // 影响层级（0=直接，1=间接，2+=二级）
	ImpactType  string   // 影响类型
	ChangePaths []string // 从变更源头到该文件的路径
	SymbolCount int      // 影响的符号数量
}

// Propagate 传播符号影响
//
// 核心逻辑：
// 1. 对于每个被修改的符号，找到哪些文件导入了它
// 2. 区分不同的导出/导入类型：
//   - export default X ↔ import X from ...
//   - export { X } ↔ import { X } from ...
//   - export * as X ↔ import * as X
//
// 3. 使用 BFS 传播影响（文件可能导入被修改的符号，该符号属于另一个被修改的文件）
func (p *SymbolPropagator) Propagate(changedSymbols []ChangedSymbol, changedNonSymbolFiles []string) *ImpactedFiles {
	result := &ImpactedFiles{
		Direct:   make(map[string]*FileImpact),
		Indirect: make(map[string]*FileImpact),
	}

	if len(changedSymbols) == 0 && len(changedNonSymbolFiles) == 0 {
		return result
	}

	// 步骤 1: 构建符号索引（快速查找符号的导出信息）
	symbolIndex := p.buildSymbolIndex(changedSymbols)

	// 步骤 2: 找出直接导入变更符号的文件
	directImpactedFiles := p.findDirectImpactedFiles(changedSymbols, symbolIndex)

	// 步骤 3: 找出导入非符号文件的文件
	nonSymbolImpactedFiles := p.findFilesImportingNonSymbols(changedNonSymbolFiles)
	for filePath, impacts := range nonSymbolImpactedFiles {
		directImpactedFiles[filePath] = append(directImpactedFiles[filePath], impacts...)
	}

	// 步骤 4: 标记直接变更的文件
	for _, sym := range changedSymbols {
		if _, exists := result.Direct[sym.FilePath]; !exists {
			result.Direct[sym.FilePath] = &FileImpact{
				FilePath:    sym.FilePath,
				ImpactLevel: 0,
				ImpactType:  "internal",
				ChangePaths: []string{sym.FilePath},
				SymbolCount: 1,
			}
		}
	}

	// 步骤 5: 标记直接变更的非符号文件
	for _, filePath := range changedNonSymbolFiles {
		if _, exists := result.Direct[filePath]; !exists {
			result.Direct[filePath] = &FileImpact{
				FilePath:    filePath,
				ImpactLevel: 0,
				ImpactType:  "internal",
				ChangePaths: []string{filePath},
				SymbolCount: 0, // 非符号文件没有符号概念
			}
		}
	}

	// 步骤 6: BFS 传播影响
	p.bfsPropagation(directImpactedFiles, symbolIndex, result)

	return result
}

// SymbolIndex 符号索引用于快速查找符号信息
type SymbolIndex struct {
	// ChangedSymbols 被修改的符号（按文件路径+符号名索引）
	ChangedSymbols map[string]*ChangedSymbolInfo

	// FileExports 每个文件的导出信息（按文件路径索引）
	FileExports map[string][]symbol_analysis.ExportInfo
}

// ChangedSymbolInfo 被修改符号的信息
type ChangedSymbolInfo struct {
	Name       string
	FilePath   string
	ExportType symbol_analysis.ExportType
}

// buildSymbolIndex 构建符号索引
func (p *SymbolPropagator) buildSymbolIndex(changedSymbols []ChangedSymbol) *SymbolIndex {
	index := &SymbolIndex{
		ChangedSymbols: make(map[string]*ChangedSymbolInfo),
		FileExports:    make(map[string][]symbol_analysis.ExportInfo),
	}

	// 索引被修改的符号
	for i := range changedSymbols {
		sym := &changedSymbols[i]
		key := sym.FilePath + "::" + sym.Name
		index.ChangedSymbols[key] = &ChangedSymbolInfo{
			Name:       sym.Name,
			FilePath:   sym.FilePath,
			ExportType: sym.ExportType,
		}
	}

	// 索引所有文件的导出信息
	for filePath, fileResult := range p.parsingResult.Js_Data {
		exports := make([]symbol_analysis.ExportInfo, 0)

		// 从 ExportDeclarations 提取
		for _, exportDecl := range fileResult.ExportDeclarations {
			for _, module := range exportDecl.ExportModules {
				exports = append(exports, symbol_analysis.ExportInfo{
					Name:       module.Identifier,
					ExportType: exportTypeFromString(module.Type),
					DeclLine:   0, // TODO: 填充
					DeclNode:   "ExportDeclaration",
				})
			}
		}

		// 从 ExportAssignments 提取（export default）
		for _, exportAssign := range fileResult.ExportAssignments {
			exports = append(exports, symbol_analysis.ExportInfo{
				Name:       p.extractDefaultExportName(exportAssign),
				ExportType: symbol_analysis.ExportTypeDefault,
				DeclLine:   0,
				DeclNode:   "ExportAssignment",
			})
		}

		// 从带有内联导出的声明中提取
		for _, varDecl := range fileResult.VariableDeclarations {
			if varDecl.Exported {
				for _, declarator := range varDecl.Declarators {
					if declarator.Identifier != "" {
						exports = append(exports, symbol_analysis.ExportInfo{
							Name:       declarator.Identifier,
							ExportType: symbol_analysis.ExportTypeNamed,
							DeclLine:   0,
							DeclNode:   "VariableDeclaration",
						})
					}
				}
			}
		}

		for _, fnDecl := range fileResult.FunctionDeclarations {
			if fnDecl.Exported {
				exports = append(exports, symbol_analysis.ExportInfo{
					Name:       fnDecl.Identifier,
					ExportType: symbol_analysis.ExportTypeNamed,
					DeclLine:   0,
					DeclNode:   "FunctionDeclaration",
				})
			}
		}

		if len(exports) > 0 {
			index.FileExports[filePath] = exports
		}
	}

	return index
}

// findDirectImpactedFiles 找出直接导入变更符号的文件
//
// 关键：区分不同的导出/导入类型
// - export default X → 只有 import X from ... 的文件受影响
// - export { X } → 只有 import { X } from ... 的文件受影响
// - export * as X → 只有 import * as X from ... 的文件受影响
func (p *SymbolPropagator) findDirectImpactedFiles(
	changedSymbols []ChangedSymbol,
	symbolIndex *SymbolIndex,
) map[string][]*SymbolImpact {
	// result[文件路径] = 受影响的符号列表
	result := make(map[string][]*SymbolImpact)

	for filePath, fileResult := range p.parsingResult.Js_Data {
		// 遍历该文件的所有导入声明
		for _, importDecl := range fileResult.ImportDeclarations {
			sourceFile := importDecl.Source.FilePath

			// 只处理项目内文件的导入
			if sourceFile == "" {
				continue
			}

			// 获取该文件的导出信息
			exports, exists := symbolIndex.FileExports[sourceFile]
			if !exists || len(exports) == 0 {
				continue
			}

			// 检查该导入了哪些被修改的符号
			impactedSymbols := p.matchImportsWithChangedSymbols(
				importDecl, exports, symbolIndex.ChangedSymbols,
			)

			if len(impactedSymbols) > 0 {
				result[filePath] = append(result[filePath], impactedSymbols...)
			}
		}
	}

	return result
}

// EnableDebug enables debug logging for testing
var EnableDebug = false

// matchImportsWithChangedSymbols 匹配导入与被修改的符号
//
// 这是符号级分析的核心：区分不同的导出/导入类型
func (p *SymbolPropagator) matchImportsWithChangedSymbols(
	importDecl projectParser.ImportDeclarationResult,
	exports []symbol_analysis.ExportInfo,
	changedSymbols map[string]*ChangedSymbolInfo,
) []*SymbolImpact {
	impacts := make([]*SymbolImpact, 0)

	// 获取导入源文件
	sourceFile := importDecl.Source.FilePath
	if sourceFile == "" {
		return impacts
	}

	// 遍历该导入声明的所有导入模块
	for _, module := range importDecl.ImportModules {
		importedName := module.Identifier

		// 检查这个导入是否匹配任何一个被修改的符号
		for _, export := range exports {
			// 特殊处理：对于 export default 的情况
			// 当导出是 "default" 且类型是 ExportTypeDefault 时
			// 应该匹配所有 default 类型的导入（不管导入名是什么）
			isDefaultExport := export.Name == "default" && export.ExportType == symbol_analysis.ExportTypeDefault
			isDefaultImport := module.Type == "default"

			if isDefaultExport && isDefaultImport {
				// 对于 export default，不管导入名是什么都匹配
				// 因为 import Button from ... 和 import MyButton from ... 都引用同一个默认导出
			} else if export.Name != importedName {
				// 对于非 default 导出，检查名称是否匹配
				continue
			}

			// 找到匹配的符号
			symbolKey := sourceFile + "::" + export.Name
			_, exists := changedSymbols[symbolKey]
			if !exists {
				continue
			}

			// 检查导出/导入类型是否匹配
			typeMatches := p.isExportImportMatch(export.ExportType, module.Type, importDecl)
			if !typeMatches {
				continue
			}

			// 添加匹配的符号影响
			impacts = append(impacts, &SymbolImpact{
				SymbolName: export.Name,
				SourceFile: sourceFile, // 符号所属的文件
				ImportType: module.Type,
				ExportType: export.ExportType,
			})
		}
	}

	return impacts
}

// isExportImportMatch 检查导出/导入类型是否匹配
//
// 核心规则：
// - export default X 匹配 import X (default import)
// - export { X } 匹配 import { X } (named import)
// - export * as X 匹配 import * as X (namespace import)
func (p *SymbolPropagator) isExportImportMatch(
	exportType symbol_analysis.ExportType,
	importType string,
	importDecl projectParser.ImportDeclarationResult,
) bool {
	// 检查匹配规则
	switch exportType {
	case symbol_analysis.ExportTypeDefault:
		// export default X 应该匹配 import X (default import)
		// 或者匹配 import { default as X } (named import with rename)
		return importType == "default" ||
			(importType == "named" && importedNameIsDefault(importDecl))

	case symbol_analysis.ExportTypeNamed:
		// export { X } 应该匹配 import { X }
		return importType == "named"

	case symbol_analysis.ExportTypeNamespace:
		// export * as X 应该匹配 import * as X
		return importType == "namespace"

	default:
		return false
	}
}

// importedNameIsDefault 检查是否为 default 重命名导入
func importedNameIsDefault(importDecl projectParser.ImportDeclarationResult) bool {
	// TODO: 实现检查 import { default as X } 的逻辑
	return false
}

// findFilesImportingNonSymbols 找出导入非符号文件的文件
//
// 对于非符号文件（CSS、图片等），任何导入它们的文件都被视为受影响
// 这是与非符号文件交互的核心逻辑
func (p *SymbolPropagator) findFilesImportingNonSymbols(
	nonSymbolFiles []string,
) map[string][]*SymbolImpact {
	result := make(map[string][]*SymbolImpact)

	// 构建非符号文件集合以便快速查找
	nonSymbolSet := make(map[string]bool)
	for _, file := range nonSymbolFiles {
		nonSymbolSet[file] = true
	}

	// 遍历所有文件，查找导入非符号文件的文件
	for filePath, fileResult := range p.parsingResult.Js_Data {
		for _, importDecl := range fileResult.ImportDeclarations {
			sourceFile := importDecl.Source.FilePath
			if sourceFile == "" {
				continue
			}

			// 检查是否导入的是非符号文件
			if nonSymbolSet[sourceFile] {
				// 创建一个特殊的符号影响，标记为非符号文件导入
				result[filePath] = append(result[filePath], &SymbolImpact{
					SymbolName: "",        // 非符号文件没有符号名
					SourceFile: sourceFile, // 导入的非符号文件路径
					ImportType: "non-symbol",
					ExportType: symbol_analysis.ExportTypeNone,
				})
			}
		}
	}

	return result
}

// bfsPropagation BFS 传播影响（二级传播）
//
// 场景：文件A 修改了符号 → 文件B 导入了该符号 → 文件C 导入了文件B
// 在这里，文件C 也会间接受影响
func (p *SymbolPropagator) bfsPropagation(
	directImpactedFiles map[string][]*SymbolImpact,
	symbolIndex *SymbolIndex,
	result *ImpactedFiles,
) {
	// 已访问的文件
	visited := make(map[string]bool)

	// 队列：存储（文件路径，影响路径，影响层级）
	queue := list.New()

	// 初始化队列：将所有直接受影响的文件加入队列
	// 同时也将这些文件加入 result.Indirect（ImpactLevel = 1）
	for filePath, impacts := range directImpactedFiles {
		if _, alreadyDirect := result.Direct[filePath]; !alreadyDirect {
			// 将直接受影响的文件加入 Indirect 结果
			result.Indirect[filePath] = &FileImpact{
				FilePath:    filePath,
				ImpactLevel: 1,
				ImpactType:  "internal",
				ChangePaths: []string{filePath},
				SymbolCount: len(impacts),
			}

			// 加入队列用于 BFS 传播到下游文件
			queue.PushBack(&propagationNode{
				filePath:    filePath,
				symbols:     impacts,
				path:        []string{filePath},
				depth:       1,
				sourceFiles: p.getSourceFilesFromImpacts(impacts),
			})
			visited[filePath] = true
		}
	}

	// BFS 遍历
	for queue.Len() > 0 {
		current := queue.Remove(queue.Front()).(*propagationNode)

		// 检查深度限制（最多传播2级）
		if current.depth > 2 {
			continue
		}

		// 获取导入当前文件的文件（下游文件）
		downstreamFiles := p.getFilesImporting(current.filePath)
		for _, downstream := range downstreamFiles {
			// 构建新的影响路径
			newPath := append(current.path, downstream)

			// 计算新的影响层级：总是递增1级
			newDepth := current.depth + 1

			// 添加到结果
			if existing, exists := result.Indirect[downstream]; exists {
				// 更新影响层级（取最小值）
				if newDepth < existing.ImpactLevel {
					existing.ImpactLevel = newDepth
				}
				// 添加新的传播路径
				existing.ChangePaths = append(existing.ChangePaths, formatPath(newPath))
				// 添加新的影响符号
				existing.SymbolCount += len(current.symbols)
			} else {
				result.Indirect[downstream] = &FileImpact{
					FilePath:    downstream,
					ImpactLevel: newDepth,
					ImpactType:  "internal",
					ChangePaths: []string{formatPath(newPath)},
					SymbolCount: len(current.symbols),
				}

				// 将下游文件加入队列（如果未访问过）
				if !visited[downstream] {
					queue.PushBack(&propagationNode{
						filePath:    downstream,
						symbols:     current.symbols, // 继承影响符号
						path:        newPath,
						depth:       newDepth,
						sourceFiles: current.sourceFiles,
					})
					visited[downstream] = true
				}
			}
		}
	}
}

// propagationNode 传播节点（用于 BFS）
type propagationNode struct {
	filePath    string          // 当前文件路径
	symbols     []*SymbolImpact // 影响的符号列表
	path        []string        // 影响路径
	depth       int             // 当前深度
	sourceFiles map[string]bool // 符号来源文件（用于计算影响层级）
}

// SymbolImpact 符号影响信息
type SymbolImpact struct {
	SymbolName string                     // 符号名称
	SourceFile string                     // 符号所属的文件
	ImportType string                     // 导入类型
	ExportType symbol_analysis.ExportType // 导出类型
}

// getFilesImporting 获取导入指定文件的所有文件
func (p *SymbolPropagator) getFilesImporting(filePath string) []string {
	importers := make([]string, 0)

	for sourceFile, fileResult := range p.parsingResult.Js_Data {
		for _, importDecl := range fileResult.ImportDeclarations {
			if importDecl.Source.FilePath == filePath {
				importers = append(importers, sourceFile)
				break
			}
		}
	}

	return importers
}

// getSourceFilesFromImpacts 从符号影响中提取来源文件
func (p *SymbolPropagator) getSourceFilesFromImpacts(impacts []*SymbolImpact) map[string]bool {
	sourceFiles := make(map[string]bool)
	for _, impact := range impacts {
		sourceFiles[impact.SourceFile] = true
	}
	return sourceFiles
}

// extractDefaultExportName 从默认导出赋值中提取名称
func (p *SymbolPropagator) extractDefaultExportName(exportAssign parser.ExportAssignmentResult) string {
	// 与 symbol_analysis 的 extractDefaultExportNameFromAssign 逻辑保持一致
	// 对于有效标识符（如 Button、helper）返回原值
	// 对于匿名表达式（如 () => {}）返回 "default"
	if exportAssign.Expression != "" {
		expr := exportAssign.Expression
		// 检查是否是有效标识符
		if p.isValidIdentifier(expr) {
			return expr
		}
	}
	return "default"
}

// isValidIdentifier 检查字符串是否是有效的标识符
func (p *SymbolPropagator) isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}
	// 简单检查：标识符应该以字母或下划线开头
	for i, c := range s {
		if i == 0 {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_') {
				return false
			}
		} else {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
				return false
			}
		}
	}
	return true
}

// exportTypeFromString 从字符串转换导出类型
func exportTypeFromString(t string) symbol_analysis.ExportType {
	switch t {
	case "default":
		return symbol_analysis.ExportTypeDefault
	case "named":
		return symbol_analysis.ExportTypeNamed
	case "namespace":
		return symbol_analysis.ExportTypeNamespace
	default:
		return symbol_analysis.ExportTypeNone
	}
}

// formatPath 格式化路径为字符串
func formatPath(path []string) string {
	result := ""
	for i, p := range path {
		if i > 0 {
			result += " → "
		}
		result += p
	}
	return result
}
