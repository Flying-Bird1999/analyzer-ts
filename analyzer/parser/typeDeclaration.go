// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（typeDeclaration.go）专门负责处理和解析 `type` 别名声明。
package parser

import (
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// TypeDeclarationResult 存储一个解析后的 `type` 别名声明信息。
type TypeDeclarationResult struct {
	Identifier     string                   `json:"identifier"`     // 类型别名的名称。
	Exported       bool                     `json:"exported"`       // 新增：标记此类型别名是否被导出。
	Raw            string                   `json:"raw"`            // 节点在源码中的原始文本。
	Reference      map[string]TypeReference `json:"reference"`      // 该类型别名所依赖的其他类型的映射。
	SourceLocation SourceLocation           `json:"sourceLocation"` // 节点在源码中的位置信息。
}

// NewTypeDeclarationResult 基于 AST 节点创建一个新的 TypeDeclarationResult 实例。
func NewTypeDeclarationResult(node *ast.Node, sourceCode string) *TypeDeclarationResult {
	raw := utils.GetNodeText(node, sourceCode)

	return &TypeDeclarationResult{
		Identifier:     "",
		Exported:       false,
		Raw:            raw,
		Reference:      make(map[string]TypeReference),
		SourceLocation: NewSourceLocation(node, sourceCode),
	}
}

// addTypeReference 是一个辅助函数，用于向结果的 Reference 映射中添加或更新类型引用。
func (tr *TypeDeclarationResult) addTypeReference(typeName string, location string, isExtend bool) {
	// 忽略空、基础类型或对类型别名自身的引用。
	if utils.IsBasicType(typeName) || typeName == "" || typeName == tr.Identifier {
		return
	}

	// 如果引用已存在，则追加新的位置信息。
	if ref, exists := tr.Reference[typeName]; exists {
		ref.Location = append(ref.Location, location)
		tr.Reference[typeName] = ref
	} else {
		// 否则，创建一个新的类型引用条目。
		tr.Reference[typeName] = TypeReference{
			Identifier: typeName,
			Location:   []string{location},
			IsExtend:   isExtend,
		}
	}
}

// AnalyzeTypeAliasDeclaration 是一个公共的、可复用的函数，用于从 AST 节点中解析 `type` 别名声明。
func AnalyzeTypeAliasDeclaration(node *ast.TypeAliasDeclaration, sourceCode string) *TypeDeclarationResult {
	tr := NewTypeDeclarationResult(node.AsNode(), sourceCode)
	typeName := node.Name().Text()
	tr.Identifier = typeName

	// 检查导出关键字
	if modifiers := node.Modifiers(); modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if modifier != nil && modifier.Kind == ast.KindExportKeyword {
				tr.Exported = true
				break
			}
		}
	}

	// 使用核心的类型分析器来递归地查找所有依赖的类型
	results := AnalyzeType(node.Type, typeName)
	for _, res := range results {
		tr.addTypeReference(res.TypeName, res.Location, false)
	}

	return tr
}

// VisitTypeAliasDeclaration 解析 `type` 别名声明。
func (p *Parser) VisitTypeAliasDeclaration(node *ast.TypeAliasDeclaration) {
	result := AnalyzeTypeAliasDeclaration(node, p.SourceCode)
	p.Result.TypeDeclarations[result.Identifier] = *result
}