// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（enumDeclaration.go）专门负责处理和解析枚举（Enum）声明。
package parser

import (
	"main/analyzer/utils"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// EnumDeclarationResult 存储一个解析后的枚举声明信息。
type EnumDeclarationResult struct {
	Identifier     string         `json:"identifier"`     // 枚举的名称。
	Exported       bool           `json:"exported"`       // 新增：标记此枚举是否被导出。
	Raw            string         `json:"raw"`            // 节点在源码中的原始文本。
	SourceLocation SourceLocation `json:"sourceLocation"` // 节点在源码中的位置信息。
}

// NewEnumDeclarationResult 基于 AST 节点创建一个新的 EnumDeclarationResult 实例。
// 它从 ast.EnumDeclaration 节点中提取枚举的名称、原始源码和位置信息。
func NewEnumDeclarationResult(node *ast.EnumDeclaration, sourceCode string) *EnumDeclarationResult {
	raw := utils.GetNodeText(node.AsNode(), sourceCode)
	pos, end := node.Pos(), node.End()

	result := &EnumDeclarationResult{
		Exported: false, // 默认为 false
		Raw:      raw,
		SourceLocation: SourceLocation{
			Start: NodePosition{Line: pos, Column: 0},
			End:   NodePosition{Line: end, Column: 0},
		},
	}

	// 获取枚举的名称节点。
	nameNode := node.Name()
	// 确认名称节点是一个标识符，然后提取其文本作为枚举的名称。
	if ast.IsIdentifier(nameNode) {
		result.Identifier = nameNode.AsIdentifier().Text
	} else {
		// 如果名称不是一个简单的标识符（异常情况），则将标识符设置为空字符串。
		result.Identifier = ""
	}

	return result
}
