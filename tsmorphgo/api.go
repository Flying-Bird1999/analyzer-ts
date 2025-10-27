/*
*

	当前只是调试而已！！！！！！
*/
package tsmorphgo

import (
	"fmt"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/lsp"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// ============================================================================
// API 分析核心数据结构
// ============================================================================

// API 表示一个 TypeScript API 定义，可以是 interface 或 typeAlias
type API struct {
	// 基本信息
	Name       string  `json:"name"`       // API 名称
	APIType    APIType `json:"apiType"`    // API 类型：interface 或 typeAlias
	ImportPath string  `json:"importPath"` // 文件路径

	// 字段信息
	Fields      []Field `json:"fields"`      // API 字段列表
	TypeText    string  `json:"typeText"`    // 当没有字段时使用（如简单类型）
	DisplayName string  `json:"displayName"` // 显示名称（考虑 @apiNameAlias）
	Depth       int     `json:"depth"`       // API 深度（用于过滤）

	// 类型别名特有信息
	TypeAliasCategory TypeAliasCategory `json:"typeAliasCategory,omitempty"` // 类型别名类别
}

// APIType 表示 API 的类型
type APIType string

const (
	APITypeInterface APIType = "interface"
	APITypeTypeAlias APIType = "typeAlias"
)

// TypeAliasCategory 表示类型别名的类别
type TypeAliasCategory string

const (
	TypeAliasCategoryIntersection      TypeAliasCategory = "intersection"      // 交叉类型：Type1 & Type2
	TypeAliasCategoryUnion             TypeAliasCategory = "union"             // 联合类型：Type1 | Type2
	TypeAliasCategoryObjectLiteral     TypeAliasCategory = "objectLiteral"     // 对象字面量：{ ... }
	TypeAliasCategoryReference         TypeAliasCategory = "reference"         // 类型引用：Type1
	TypeAliasCategoryReferencePartial  TypeAliasCategory = "referencePartial"  // Partial<T>
	TypeAliasCategoryReferenceReadonly TypeAliasCategory = "referenceReadonly" // Readonly<T>
	TypeAliasCategoryReferencePick     TypeAliasCategory = "referencePick"     // Pick<T, K>
	TypeAliasCategoryReferenceOmit     TypeAliasCategory = "referenceOmit"     // Omit<T, K>
)

// Field 表示 API 字段
type Field struct {
	Name         string `json:"name"`         // 字段名称
	Type         string `json:"type"`         // 字段类型文本
	Description  string `json:"description"`  // 字段描述
	Required     bool   `json:"required"`     // 是否必需
	DefaultValue string `json:"defaultValue"` // 默认值
	Internal     bool   `json:"internal"`     // 内部字段，不对外展示
	Deprecated   bool   `json:"deprecated"`   // 是否废弃
	Version      string `json:"version"`      // 添加版本
	Readonly     bool   `json:"readonly"`     // 是否只读

	// 原始符号信息，用于深度计算和衍生
	Symbol       *Symbol                 `json:"-"` // 原始符号
	Declaration  *Node                   `json:"-"` // 声明节点
	DisplayParts []lsp.SymbolDisplayPart `json:"-"` // 显示部件（用于类型推导）
}

// SymbolDisplayPartKind 表示显示部件的类型
type SymbolDisplayPartKind string

const (
	SymbolDisplayPartKindClassName     SymbolDisplayPartKind = "className"
	SymbolDisplayPartKindEnumName      SymbolDisplayPartKind = "enumName"
	SymbolDisplayPartKindInterfaceName SymbolDisplayPartKind = "interfaceName"
	SymbolDisplayPartKindTypeName      SymbolDisplayPartKind = "typeName"
	SymbolDisplayPartKindParameterName SymbolDisplayPartKind = "parameterName"
	SymbolDisplayPartKindPropertyName  SymbolDisplayPartKind = "propertyName"
	SymbolDisplayPartKindText          SymbolDisplayPartKind = "text"
	SymbolDisplayPartKindKeyword       SymbolDisplayPartKind = "keyword"
)

// APIAnalyzer API 分析器，提供完整的 API 分析和导出功能
type APIAnalyzer struct {
	project    *Project
	lspService *lsp.Service
}

// NewAPIAnalyzer 创建新的 API 分析器
func NewAPIAnalyzer(project *Project) (*APIAnalyzer, error) {
	lspService, err := createLSPService(project)
	if err != nil {
		return nil, fmt.Errorf("failed to create LSP service: %w", err)
	}

	return &APIAnalyzer{
		project:    project,
		lspService: lspService,
	}, nil
}

// Close 关闭分析器并释放资源
func (a *APIAnalyzer) Close() {
	if a.lspService != nil {
		a.lspService.Close()
	}
}

// CollectAPIs 收集项目中的所有 API
func (a *APIAnalyzer) CollectAPIs() ([]API, error) {
	var apis []API

	// 遍历所有源文件
	for _, sourceFile := range a.project.GetSourceFiles() {
		fileAPIs, err := a.collectAPIsFromFile(sourceFile)
		if err != nil {
			return nil, fmt.Errorf("failed to collect APIs from file %s: %w",
				sourceFile.GetFilePath(), err)
		}
		apis = append(apis, fileAPIs...)
	}

	return apis, nil
}

// collectAPIsFromFile 从单个文件中收集 API
func (a *APIAnalyzer) collectAPIsFromFile(sourceFile *SourceFile) ([]API, error) {
	var apis []API

	// 收集 interface 类型的 API
	interfaceAPIs := a.collectInterfaceAPIs(sourceFile)
	apis = append(apis, interfaceAPIs...)

	// 收集 typeAlias 类型的 API
	typeAliasAPIs := a.collectTypeAliasAPIs(sourceFile)
	apis = append(apis, typeAliasAPIs...)

	return apis, nil
}

// collectInterfaceAPIs 收集 interface 类型的 API
func (a *APIAnalyzer) collectInterfaceAPIs(sourceFile *SourceFile) []API {
	var apis []API

	sourceFile.ForEachDescendant(func(node Node) {
		if node.Kind == ast.KindInterfaceDeclaration {
			api := a.analyzeInterfaceDeclaration(node, sourceFile)
			if api != nil {
				apis = append(apis, *api)
			}
		}
	})

	return apis
}

// collectTypeAliasAPIs 收集 typeAlias 类型的 API
func (a *APIAnalyzer) collectTypeAliasAPIs(sourceFile *SourceFile) []API {
	var apis []API

	sourceFile.ForEachDescendant(func(node Node) {
		if node.Kind == ast.KindTypeAliasDeclaration {
			api := a.analyzeTypeAliasDeclaration(node, sourceFile)
			if api != nil {
				apis = append(apis, *api)
			}
		}
	})

	return apis
}

// analyzeInterfaceDeclaration 分析 interface 声明
func (a *APIAnalyzer) analyzeInterfaceDeclaration(node Node, sourceFile *SourceFile) *API {
	interfaceName := a.getNodeName(node)
	if interfaceName == "" {
		return nil
	}

	api := &API{
		Name:        interfaceName,
		APIType:     APITypeInterface,
		ImportPath:  sourceFile.GetFilePath(),
		Depth:       1, // interface 默认深度为 1
		DisplayName: a.getAPIDisplayName(node, interfaceName),
	}

	// 收集 interface 的字段
	fields := a.collectFieldsFromType(node, api.Depth)
	api.Fields = fields

	// 如果没有收集到字段，尝试使用 QuickInfo 作为兜底
	if len(fields) == 0 {
		if quickInfo, err := node.GetQuickInfo(); err == nil && quickInfo != nil {
			api.TypeText = quickInfo.TypeText
		}
	}

	return api
}

// analyzeTypeAliasDeclaration 分析 typeAlias 声明
func (a *APIAnalyzer) analyzeTypeAliasDeclaration(node Node, sourceFile *SourceFile) *API {
	typeAliasName := a.getNodeName(node)
	if typeAliasName == "" {
		return nil
	}

	api := &API{
		Name:        typeAliasName,
		APIType:     APITypeTypeAlias,
		ImportPath:  sourceFile.GetFilePath(),
		Depth:       2, // typeAlias 默认深度为 2
		DisplayName: a.getAPIDisplayName(node, typeAliasName),
	}

	// 分析类型别名类别
	api.TypeAliasCategory = a.analyzeTypeAliasCategory(node)

	// 根据不同类别收集字段
	fields := a.collectFieldsFromType(node, api.Depth)
	api.Fields = fields

	// 如果没有收集到字段，尝试使用 QuickInfo 作为兜底
	if len(fields) == 0 {
		if quickInfo, err := node.GetQuickInfo(); err == nil && quickInfo != nil {
			api.TypeText = quickInfo.TypeText
		}
	}

	return api
}

// getNodeName 获取节点的名称
func (a *APIAnalyzer) getNodeName(node Node) string {
	var nameNode *Node
	node.ForEachChild(func(child *ast.Node) bool {
		if child.Kind == ast.KindIdentifier {
			nameNode = &Node{Node: child, sourceFile: node.sourceFile}
			return true
		}
		return false
	})

	if nameNode != nil {
		return nameNode.GetText()
	}
	return ""
}

// getAPIDisplayName 获取 API 的显示名称，考虑 @apiNameAlias 标签
func (a *APIAnalyzer) getAPIDisplayName(node Node, defaultName string) string {
	// TODO: 实现从 JSDoc 中读取 @apiNameAlias 标签
	// 当前暂时返回默认名称
	return defaultName
}

// analyzeTypeAliasCategory 分析类型别名的类别
func (a *APIAnalyzer) analyzeTypeAliasCategory(node Node) TypeAliasCategory {
	// TODO: 实现复杂的类型分析逻辑
	// 当前暂时返回对象字面量类型
	return TypeAliasCategoryObjectLiteral
}

// collectFieldsFromType 收集类型的字段，考虑深度过滤
func (a *APIAnalyzer) collectFieldsFromType(node Node, maxDepth int) []Field {
	// TODO: 实现基于 TypeChecker.getPropertiesOfType 的字段收集
	// 并实现深度过滤逻辑
	// 当前暂时返回空切片
	return []Field{}
}
