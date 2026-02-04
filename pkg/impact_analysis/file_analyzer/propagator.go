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
	// reverseImportIndex 反向导入索引：filePath -> importers[]
	// 预构建后可将 getFilesImporting 从 O(N) 优化到 O(1)
	reverseImportIndex map[string][]string
	// indexBuilt 标记索引是否已构建
	indexBuilt bool
	// maxDepth 最大传播深度（0=不限制，推荐值：3-5）
	maxDepth int
}

// DefaultMaxDepth 默认最大传播深度
const DefaultMaxDepth = 5

// NewSymbolPropagator 创建符号影响传播器（使用默认深度）
func NewSymbolPropagator(parsingResult *projectParser.ProjectParserResult) *SymbolPropagator {
	return NewSymbolPropagatorWithMaxDepth(parsingResult, DefaultMaxDepth)
}

// NewSymbolPropagatorWithMaxDepth 创建符号影响传播器（指定最大深度）
func NewSymbolPropagatorWithMaxDepth(parsingResult *projectParser.ProjectParserResult, maxDepth int) *SymbolPropagator {
	p := &SymbolPropagator{
		parsingResult:      parsingResult,
		reverseImportIndex: make(map[string][]string),
		indexBuilt:         false,
		maxDepth:           maxDepth,
	}
	// 在创建时预构建索引
	p.buildReverseIndex()
	return p
}

// buildReverseIndex 构建反向导入索引
// 将"哪些文件导入了指定文件"的关系预先建立索引
// 时间复杂度: O(N)，N 为文件数
func (p *SymbolPropagator) buildReverseIndex() {
	if p.parsingResult == nil {
		return
	}

	p.reverseImportIndex = make(map[string][]string)

	for sourceFile, fileResult := range p.parsingResult.Js_Data {
		for _, importDecl := range fileResult.ImportDeclarations {
			targetFile := importDecl.Source.FilePath
			if targetFile == "" {
				continue // 跳过外部依赖
			}

			// 记录：targetFile 被 sourceFile 导入
			p.reverseImportIndex[targetFile] = appendUnique(
				p.reverseImportIndex[targetFile],
				sourceFile,
			)
		}
	}

	p.indexBuilt = true
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

		// 检查深度限制（maxDepth=0 表示不限制）
		if p.maxDepth > 0 && current.depth > p.maxDepth {
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
// 使用预构建的反向索引，时间复杂度从 O(N) 优化到 O(1)
func (p *SymbolPropagator) getFilesImporting(filePath string) []string {
	// 使用预构建的索引
	if p.indexBuilt {
		return p.reverseImportIndex[filePath]
	}

	// 降级处理：如果索引未构建，使用原始方法（不应发生）
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

// =============================================================================
// Re-export 链追踪支持
// =============================================================================

// SymbolOrigin 符号来源信息
// 用于追踪 re-export 链，找出符号的实际来源文件
type SymbolOrigin struct {
	OriginalFile  string   // 实际来源文件（符号最初定义的文件）
	SymbolName    string   // 符号名称
	ReexportChain []string // Re-export 链（从源头到当前文件的路径）
}

// SymbolOriginMap 符号来源映射
// key: "文件路径::符号名称"
// value: 符号的实际来源信息
type SymbolOriginMap map[string]*SymbolOrigin

// BuildSymbolOriginMap 构建符号来源映射表
// 用于追踪 re-export 链，将所有符号映射到它们的实际源头文件
func BuildSymbolOriginMap(parsingResult *projectParser.ProjectParserResult) *SymbolOriginMap {
	originMap := make(SymbolOriginMap)

	// 步骤 1: 首先建立直接导出映射（自身定义的符号）
	// A.ts::X → A.ts
	for file, fileResult := range parsingResult.Js_Data {
		// 从符号分析结果中获取导出信息
		exports := make([]symbol_analysis.ExportInfo, 0)

		// 从 ExportDeclarations 提取（跳过 re-export）
		for _, exportDecl := range fileResult.ExportDeclarations {
			// 跳过 re-export，只处理直接导出
			if exportDecl.Source != nil && exportDecl.Source.FilePath != "" {
				continue
			}

			for _, module := range exportDecl.ExportModules {
				exports = append(exports, symbol_analysis.ExportInfo{
					Name:       module.Identifier,
					ExportType: exportTypeFromString(module.Type),
					DeclLine:   0,
					DeclNode:   "ExportDeclaration",
				})
			}
		}

		// 从 ExportAssignments 提取（export default）
		for _, exportAssign := range fileResult.ExportAssignments {
			exports = append(exports, symbol_analysis.ExportInfo{
				Name:       extractDefaultExportNameFromAssign(exportAssign),
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

		// 记录符号来源
		for _, export := range exports {
			key := file + "::" + export.Name
			originMap[key] = &SymbolOrigin{
				OriginalFile:  file,
				SymbolName:    export.Name,
				ReexportChain: []string{},
			}
		}
	}

	// 步骤 2: 迭代处理 Re-export，直到收敛
	// B.ts re-export X from A.ts
	// originMap["B.ts::X"] = originMap["A.ts::X"]
	const MaxReexportIterations = 10 // 防止无限循环

	for iteration := 0; iteration < MaxReexportIterations; iteration++ {
		changed := false

		for file, fileResult := range parsingResult.Js_Data {
			// 从符号分析结果中获取 Re-export 信息
			// 这里我们需要从 FileAnalysisResult 中获取，但由于当前结构限制，
			// 我们先使用 ExportDeclaration 中的 Source 信息
			for _, exportDecl := range fileResult.ExportDeclarations {
				// 检查是否为 Re-export（Source 不为空）
				if exportDecl.Source == nil || exportDecl.Source.FilePath == "" {
					continue
				}

				sourceFile := exportDecl.Source.FilePath
				exportedNames := make([]string, 0, len(exportDecl.ExportModules))
				for _, module := range exportDecl.ExportModules {
					exportedNames = append(exportedNames, module.Identifier)
				}

				// 处理每个 re-export 的符号
				for _, symbolName := range exportedNames {
					key := file + "::" + symbolName

					// 如果已经存在映射，跳过
					if _, exists := originMap[key]; exists {
						continue
					}

					// 查找源文件的符号来源
					sourceKey := sourceFile + "::" + symbolName
					if origin, found := originMap[sourceKey]; found {
						// 创建新的符号来源，记录 Re-export 链
						originMap[key] = &SymbolOrigin{
							OriginalFile:  origin.OriginalFile,
							SymbolName:    symbolName,
							ReexportChain: append([]string{file}, origin.ReexportChain...),
						}
						changed = true
					}
				}
			}
		}

		// 如果没有变化，提前退出
		if !changed {
			break
		}
	}

	return &originMap
}

// extractDefaultExportNameFromAssign 从默认导出赋值中提取名称
func extractDefaultExportNameFromAssign(exportAssign interface{}) string {
	// 简化版本，直接返回 "default"
	return "default"
}

// GetSymbolOrigin 获取符号的实际来源
// 返回：实际来源文件，如果符号不存在则返回空字符串
func (m SymbolOriginMap) GetSymbolOrigin(filePath, symbolName string) *SymbolOrigin {
	key := filePath + "::" + symbolName
	if origin, exists := m[key]; exists {
		return origin
	}
	return nil
}

// =============================================================================
// 支持 Re-export 的传播方法
// =============================================================================

// PropagateWithReexport 使用符号来源映射进行传播
// 支持 Re-export 链追踪：当 A.ts 的符号通过 B.ts re-export 到 C.ts
// 如果 A.ts 变更，C.ts 也会被标记为受影响
func (p *SymbolPropagator) PropagateWithReexport(
	changedSymbols []ChangedSymbol,
	changedNonSymbolFiles []string,
	originMap *SymbolOriginMap,
) *ImpactedFiles {
	result := &ImpactedFiles{
		Direct:   make(map[string]*FileImpact),
		Indirect: make(map[string]*FileImpact),
	}

	if len(changedSymbols) == 0 && len(changedNonSymbolFiles) == 0 {
		return result
	}

	// 构建符号索引（使用 re-export 增强版）
	symbolIndex := p.buildSymbolIndexWithReexport(changedSymbols, originMap)

	// 找出直接导入变更符号的文件（使用 re-export 增强版）
	directImpactedFiles := p.findDirectImpactedFilesWithReexport(changedSymbols, symbolIndex, originMap)

	// 找出 re-export 变更符号的文件
	// 例如：B.ts export { X } from './A'，当 A.ts 的 X 变更时，B.ts 也应该被标记为受影响
	reexportImpactedFiles := p.findFilesReexportingChangedSymbols(changedSymbols, originMap)
	for filePath, impacts := range reexportImpactedFiles {
		directImpactedFiles[filePath] = append(directImpactedFiles[filePath], impacts...)
	}

	// 找出导入非符号文件的文件
	nonSymbolImpactedFiles := p.findFilesImportingNonSymbols(changedNonSymbolFiles)
	for filePath, impacts := range nonSymbolImpactedFiles {
		directImpactedFiles[filePath] = append(directImpactedFiles[filePath], impacts...)
	}

	// 标记直接变更的文件
	for _, sym := range changedSymbols {
		if _, exists := result.Direct[sym.FilePath]; !exists {
			result.Direct[sym.FilePath] = &FileImpact{
				FilePath:    sym.FilePath,
				ImpactLevel: 0,
				ChangePaths: []string{sym.FilePath},
				SymbolCount: 1,
			}
		}
	}

	// 标记直接变更的非符号文件
	for _, filePath := range changedNonSymbolFiles {
		if _, exists := result.Direct[filePath]; !exists {
			result.Direct[filePath] = &FileImpact{
				FilePath:    filePath,
				ImpactLevel: 0,
				ChangePaths: []string{filePath},
				SymbolCount: 0,
			}
		}
	}

	// BFS 传播影响
	p.bfsPropagation(directImpactedFiles, symbolIndex, result)

	return result
}

// buildSymbolIndexWithReexport 构建包含 re-export 信息的符号索引
func (p *SymbolPropagator) buildSymbolIndexWithReexport(
	changedSymbols []ChangedSymbol,
	originMap *SymbolOriginMap,
) *SymbolIndex {
	index := &SymbolIndex{
		ChangedSymbols: make(map[string]*ChangedSymbolInfo),
		FileExports:    make(map[string][]symbol_analysis.ExportInfo),
	}

	// 索引被修改的符号（使用实际来源）
	for i := range changedSymbols {
		sym := &changedSymbols[i]
		originalFile := sym.FilePath
		symbolName := sym.Name

		// 检查是否是 re-export 的符号
		if originMap != nil {
			if origin := originMap.GetSymbolOrigin(sym.FilePath, sym.Name); origin != nil {
				originalFile = origin.OriginalFile
				symbolName = origin.SymbolName
			}
		}

		key := originalFile + "::" + symbolName
		index.ChangedSymbols[key] = &ChangedSymbolInfo{
			Name:       symbolName,
			FilePath:   originalFile,
			ExportType: sym.ExportType,
		}
	}

	// 索引所有文件的导出信息
	for filePath, fileResult := range p.parsingResult.Js_Data {
		exports := make([]symbol_analysis.ExportInfo, 0)

		// 从 ExportDeclarations 提取（包括 re-export）
		for _, exportDecl := range fileResult.ExportDeclarations {
			// Re-export 也需要被包含，因为其他文件可能从该文件导入 re-export 的符号
			// 例如：C.ts import { X } from './B'，B.ts re-export X from A.ts
			// B.ts 的 exports 需要包含 X，这样 C.ts 才能匹配到

			for _, module := range exportDecl.ExportModules {
				exports = append(exports, symbol_analysis.ExportInfo{
					Name:       module.Identifier,
					ExportType: exportTypeFromString(module.Type),
					DeclLine:   0,
					DeclNode:   "ExportDeclaration",
				})
			}
		}

		// 从 ExportAssignments 提取（export default）
		for _, exportAssign := range fileResult.ExportAssignments {
			exports = append(exports, symbol_analysis.ExportInfo{
				Name:       extractDefaultExportNameFromAssign(exportAssign),
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

// findDirectImpactedFilesWithReexport 找出直接导入变更符号的文件（支持 Re-export 追溯）
func (p *SymbolPropagator) findDirectImpactedFilesWithReexport(
	changedSymbols []ChangedSymbol,
	symbolIndex *SymbolIndex,
	originMap *SymbolOriginMap,
) map[string][]*SymbolImpact {
	result := make(map[string][]*SymbolImpact)

	// 构建变更文件集合，用于快速查找
	changedFileSet := make(map[string]bool)
	for _, sym := range changedSymbols {
		changedFileSet[sym.FilePath] = true
	}

	for filePath, fileResult := range p.parsingResult.Js_Data {
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

			// 检查该导入了哪些被修改的符号（支持 Re-export 追溯）
			impactedSymbols := p.matchImportsWithChangedSymbolsWithReexport(
				importDecl, sourceFile, exports, symbolIndex.ChangedSymbols, originMap,
			)

			if len(impactedSymbols) > 0 {
				result[filePath] = append(result[filePath], impactedSymbols...)
			}
		}
	}

	return result
}

// findFilesReexportingChangedSymbols 找出 re-export 变更符号的文件
// 例如：B.ts export { X } from './A'，当 A.ts 的 X 变更时，B.ts 也应该被标记为受影响
func (p *SymbolPropagator) findFilesReexportingChangedSymbols(
	changedSymbols []ChangedSymbol,
	originMap *SymbolOriginMap,
) map[string][]*SymbolImpact {
	result := make(map[string][]*SymbolImpact)

	// 构建变更符号集合，用于快速查找
	changedSet := make(map[string]bool)
	for _, sym := range changedSymbols {
		key := sym.FilePath + "::" + sym.Name
		changedSet[key] = true
	}

	// 遍历所有文件，检查是否有文件 re-export 了变更的符号
	for filePath, fileResult := range p.parsingResult.Js_Data {
		for _, exportDecl := range fileResult.ExportDeclarations {
			// 只处理 re-export（有 Source 的 ExportDeclaration）
			if exportDecl.Source == nil || exportDecl.Source.FilePath == "" {
				continue
			}

			sourceFile := exportDecl.Source.FilePath

			// 检查 re-export 的符号是否包含变更的符号
			for _, module := range exportDecl.ExportModules {
				symbolName := module.Identifier

				// 检查这个符号是否是变更的符号（直接匹配）
				key := sourceFile + "::" + symbolName
				if !changedSet[key] {
					// 如果不是直接匹配，尝试通过 originMap 查找实际来源
					if originMap != nil {
						// 检查 sourceFile 的这个符号是否来自变更文件
						if origin := originMap.GetSymbolOrigin(sourceFile, symbolName); origin != nil {
							originKey := origin.OriginalFile + "::" + origin.SymbolName
							if !changedSet[originKey] {
								continue
							}
							// 更新 sourceFile 为实际来源文件
							sourceFile = origin.OriginalFile
							symbolName = origin.SymbolName
						} else {
							continue
						}
					} else {
						continue
					}
				}

				// 找到了！这个文件 re-export 了变更的符号
				result[filePath] = append(result[filePath], &SymbolImpact{
					SymbolName: symbolName,
					SourceFile: sourceFile,
					ImportType: module.Type,
				})
			}
		}
	}

	return result
}

// matchImportsWithChangedSymbolsWithReexport 匹配导入与被修改的符号（支持 Re-export 追溯）
func (p *SymbolPropagator) matchImportsWithChangedSymbolsWithReexport(
	importDecl projectParser.ImportDeclarationResult,
	sourceFile string,
	exports []symbol_analysis.ExportInfo,
	changedSymbols map[string]*ChangedSymbolInfo,
	originMap *SymbolOriginMap,
) []*SymbolImpact {
	impacts := make([]*SymbolImpact, 0)

	// 遍历该导入声明的所有导入模块
	for _, module := range importDecl.ImportModules {
		importedName := module.Identifier

		// 检查这个导入是否匹配任何一个被修改的符号
		for _, export := range exports {
			// 特殊处理：对于 export default 的情况
			isDefaultExport := export.Name == "default" && export.ExportType == symbol_analysis.ExportTypeDefault
			isDefaultImport := module.Type == "default"

			if isDefaultExport && isDefaultImport {
				// 对于 export default，不管导入名是什么都匹配
			} else if export.Name != importedName {
				// 对于非 default 导出，检查名称是否匹配
				continue
			}

			// 找到符号名称
			symbolName := export.Name

			// 【Re-export 追溯】检查符号是否通过 re-export 来自变更的文件
			actualSourceFile := sourceFile
			if originMap != nil {
				if origin := originMap.GetSymbolOrigin(sourceFile, symbolName); origin != nil {
					// 使用实际来源文件
					actualSourceFile = origin.OriginalFile
					symbolName = origin.SymbolName
				}
			}

			// 检查实际来源文件和符号是否变更
			key := actualSourceFile + "::" + symbolName
			_, exists := changedSymbols[key]
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
				SymbolName: symbolName,
				SourceFile: actualSourceFile, // 使用实际来源文件
				ImportType: module.Type,
				ExportType: export.ExportType,
			})
		}
	}

	return impacts
}
