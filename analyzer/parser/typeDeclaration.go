// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（typeDeclaration.go）专门负责处理和解析 `type` 别名声明。
package parser

import (
	"main/analyzer/utils"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// TypeDeclarationResult 存储一个解析后的 `type` 别名声明信息。
type TypeDeclarationResult struct {
	Identifier     string                   `json:"identifier"`     // 类型别名的名称。
	Raw            string                   `json:"raw"`            // 节点在源码中的原始文本。
	Reference      map[string]TypeReference `json:"reference"`      // 该类型别名所依赖的其他类型的映射。
	SourceLocation SourceLocation           `json:"sourceLocation"` // 节点在源码中的位置信息。
}

// NewTypeDeclarationResult 基于 AST 节点创建一个新的 TypeDeclarationResult 实例。
func NewTypeDeclarationResult(node *ast.Node, sourceCode string) *TypeDeclarationResult {
	raw := utils.GetNodeText(node, sourceCode)
	pos, end := node.Pos(), node.End()

	return &TypeDeclarationResult{
		Identifier: "",
		Raw:        raw,
		Reference:  make(map[string]TypeReference),
		SourceLocation: SourceLocation{
			Start: NodePosition{Line: pos, Column: 0},
			End:   NodePosition{Line: end, Column: 0},
		},
	}
}

// AnalyzeTypeDecl 是分析 `type` 别名声明的入口函数。
// 重构后，此函数变得非常简洁，直接调用共享的 AnalyzeType 函数即可处理所有情况。
func (tr *TypeDeclarationResult) AnalyzeTypeDecl(typeDecl *ast.TypeAliasDeclaration) {
	typeName := typeDecl.Name().Text()
	tr.Identifier = typeName

	// 调用共享的类型分析器来处理别名的具体类型定义。
	results := AnalyzeType(typeDecl.Type, typeName)
	for _, res := range results {
		tr.addTypeReference(res.TypeName, res.Location, false)
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
