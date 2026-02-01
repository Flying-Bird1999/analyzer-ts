// Package impact_analysis 提供符号级影响分析功能。
package impact_analysis

import (
	"path/filepath"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

// =============================================================================
// 符号匹配器
// =============================================================================

// Matcher 符号匹配器
// 负责将符号变更匹配到对应的组件，并构建符号依赖映射
type Matcher struct {
	project          *tsmorphgo.Project
	parsingResult    *projectParser.ProjectParserResult
	componentManifest *ComponentManifest
}

// NewMatcher 创建符号匹配器
func NewMatcher(
	project *tsmorphgo.Project,
	parsingResult *projectParser.ProjectParserResult,
	manifest *ComponentManifest,
) *Matcher {
	return &Matcher{
		project:          project,
		parsingResult:    parsingResult,
		componentManifest: manifest,
	}
}

// =============================================================================
// 符号依赖映射类型定义
// =============================================================================

// SymbolDependencyMap 符号依赖映射
// 记录组件之间的符号级依赖关系
type SymbolDependencyMap struct {
	// ComponentExports 组件导出的符号
	// map[组件名][]符号引用
	ComponentExports map[string][]SymbolRef

	// SymbolImports 组件导入的符号
	// map[组件名][]导入关系
	SymbolImports map[string][]ImportRelation
}

// ImportRelation 导入关系
type ImportRelation struct {
	// SourceComponent 来源组件（或外部依赖名称）
	SourceComponent string

	// ImportedSymbols 导入的符号列表
	ImportedSymbols []SymbolRef

	// ImportType 导入类型: named/default/namespace
	ImportType string
}

// NewSymbolDependencyMap 创建符号依赖映射
func NewSymbolDependencyMap() *SymbolDependencyMap {
	return &SymbolDependencyMap{
		ComponentExports: make(map[string][]SymbolRef),
		SymbolImports:    make(map[string][]ImportRelation),
	}
}

// =============================================================================
// 符号匹配（将变更符号映射到组件）
// =============================================================================
// 返回: map[组件名][]符号变更
func (m *Matcher) MatchSymbolsToComponents(
	symbols []SymbolChange,
) map[string][]SymbolChange {
	result := make(map[string][]SymbolChange)

	for _, symbol := range symbols {
		// 步骤 1: 根据文件路径确定所属组件
		component := m.findComponentByPath(symbol.FilePath)
		if component == nil {
			continue
		}

		// 步骤 2: 更新符号的组件信息
		symbol.ComponentName = component.Name

		// 步骤 3: 添加到结果
		result[component.Name] = append(result[component.Name], symbol)
	}

	return result
}

// findComponentByPath 根据文件路径查找组件
func (m *Matcher) findComponentByPath(filePath string) *Component {
	if m.componentManifest == nil {
		return nil
	}

	// 规范化文件路径（使用正斜杠）
	normalizedPath := filepath.ToSlash(filePath)

	// 遍历组件清单，检查文件是否在组件范围内
	for _, comp := range m.componentManifest.Components {
		for _, scope := range comp.Scopes {
			// 规范化 scope 路径
			normalizedScope := filepath.ToSlash(scope)
			// 检查文件是否在组件范围内
			if strings.HasPrefix(normalizedPath, normalizedScope) {
				return &comp
			}
		}
	}

	return nil
}

// =============================================================================
// 符号依赖映射构建（使用 projectParser 的解析结果）
// =============================================================================

// BuildSymbolDependencyMap 构建符号依赖映射
// 利用 projectParser.Js_Data 中已解析的 ImportDeclarations
func (m *Matcher) BuildSymbolDependencyMap() *SymbolDependencyMap {
	depMap := NewSymbolDependencyMap()

	if m.parsingResult == nil {
		return depMap
	}

	// 遍历所有已解析的 JS/TS 文件
	for filePath, fileResult := range m.parsingResult.Js_Data {
		// 获取文件所属组件
		component := m.findComponentByPath(filePath)
		if component == nil {
			continue
		}

		// 处理该文件的导入声明
		for _, importDecl := range fileResult.ImportDeclarations {
			m.processImportDeclaration(importDecl, component.Name, depMap)
		}

		// 记录该文件的导出（用于符号依赖匹配）
		for _, exportDecl := range fileResult.ExportDeclarations {
			m.processExportDeclaration(exportDecl, component.Name, depMap)
		}

		// 记录默认导出
		for _, exportAssign := range fileResult.ExportAssignments {
			m.processExportAssignment(exportAssign, component.Name, depMap)
		}
	}

	return depMap
}

// processImportDeclaration 处理导入声明
func (m *Matcher) processImportDeclaration(
	importDecl projectParser.ImportDeclarationResult,
	currentComponent string,
	depMap *SymbolDependencyMap,
) {
	// 检查导入来源是否为项目内文件（非 NPM 包）
	if importDecl.Source.FilePath == "" {
		// 可能是 NPM 包或其他外部依赖
		return
	}

	// 根据解析后的路径确定来源组件
	sourceComponent := m.findComponentByPath(importDecl.Source.FilePath)
	var sourceCompName string
	if sourceComponent != nil {
		sourceCompName = sourceComponent.Name
	} else {
		// 外部依赖（NPM 包或项目外的文件）
		sourceCompName = importDecl.Source.FilePath
	}

	// 构建导入关系
	relation := ImportRelation{
		SourceComponent: sourceCompName,
		ImportedSymbols: make([]SymbolRef, 0),
		ImportType:      m.determineImportType(importDecl),
	}

	// 提取导入的符号
	for _, module := range importDecl.ImportModules {
		relation.ImportedSymbols = append(relation.ImportedSymbols, SymbolRef{
			Name:       module.Identifier,
			Kind:       m.inferSymbolKind(module),
			FilePath:   importDecl.Source.FilePath,
			ExportType: m.inferExportType(module.Type),
		})
	}

	// 记录导入关系
	depMap.SymbolImports[currentComponent] = append(
		depMap.SymbolImports[currentComponent],
		relation,
	)
}

// processExportDeclaration 处理导出声明
func (m *Matcher) processExportDeclaration(
	exportDecl projectParser.ExportDeclarationResult,
	componentName string,
	depMap *SymbolDependencyMap,
) {
	// 记录组件的导出符号
	for _, module := range exportDecl.ExportModules {
		depMap.ComponentExports[componentName] = append(
			depMap.ComponentExports[componentName],
			SymbolRef{
				Name:       module.Identifier,
				Kind:       SymbolKindVariable, // 默认类型，后续可细化
				FilePath:   "", // 同文件内
				ExportType: ExportTypeNamed,
			},
		)
	}
}

// processExportAssignment 处理默认导出
func (m *Matcher) processExportAssignment(
	exportAssign parser.ExportAssignmentResult,
	componentName string,
	depMap *SymbolDependencyMap,
) {
	// 默认导出
	depMap.ComponentExports[componentName] = append(
		depMap.ComponentExports[componentName],
		SymbolRef{
			Name:       m.extractDefaultExportName(exportAssign),
			Kind:       SymbolKindVariable, // 默认类型，后续可细化
			FilePath:   "",
			ExportType: ExportTypeDefault,
		},
	)
}

// determineImportType 确定导入类型
func (m *Matcher) determineImportType(importDecl projectParser.ImportDeclarationResult) string {
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
func (m *Matcher) inferSymbolKind(module projectParser.ImportModule) SymbolKind {
	// TODO: 可以通过查看源文件中该符号的实际类型来推断
	// 当前返回默认类型
	return SymbolKindVariable
}

// inferExportType 推断导出类型
func (m *Matcher) inferExportType(importType string) ExportType {
	switch importType {
	case "default":
		return ExportTypeDefault
	case "namespace":
		return ExportTypeNamespace
	default:
		return ExportTypeNamed
	}
}

// extractDefaultExportName 从默认导出赋值中提取名称
func (m *Matcher) extractDefaultExportName(exportAssign parser.ExportAssignmentResult) string {
	// 使用 projectParser 提供的表达式
	if exportAssign.Expression != "" {
		return exportAssign.Expression
	}
	return "default"
}
