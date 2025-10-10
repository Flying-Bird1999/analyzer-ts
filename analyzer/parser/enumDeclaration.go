// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（enumDeclaration.go）专门负责处理和解析枚举（Enum）声明。
package parser

import (
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// EnumDeclarationResult 存储一个解析后的枚举声明信息。
type EnumDeclarationResult struct {
	Identifier     string         `json:"identifier"`     // 枚举的名称。
	Exported       bool           `json:"exported"`       // 新增：标记此枚举是否被导出。
	Raw            string         `json:"raw,omitempty"`            // 节点在源码中的原始文本
	SourceLocation *SourceLocation `json:"sourceLocation,omitempty"` // 节点在源码中的位置信息
	Node           *ast.Node      `json:"-"`                     // 对应的 AST 节点，不在 JSON 中序列化。
}

// AnalyzeEnumDeclaration 是一个公共的、可复用的函数，用于从 AST 节点中解析枚举声明。
func AnalyzeEnumDeclaration(node *ast.EnumDeclaration, sourceCode string) *EnumDeclarationResult {
	raw := utils.GetNodeText(node.AsNode(), sourceCode)

	result := &EnumDeclarationResult{
		Exported:       false, // 默认为 false
		Raw:            raw,
		SourceLocation: NewSourceLocation(node.AsNode(), sourceCode),
		Node:           node.AsNode(),
	}

	nameNode := node.Name()
	// 确认名称节点是一个标识符，然后提取其文本作为枚举的名称。
	if ast.IsIdentifier(nameNode) {
		result.Identifier = nameNode.AsIdentifier().Text
	} else {
		// 如果名称不是一个简单的标识符（异常情况），则将标识符设置为空字符串。
		result.Identifier = ""
	}

	// 检查导出关键字
	if modifiers := node.Modifiers(); modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if modifier != nil && modifier.Kind == ast.KindExportKeyword {
				result.Exported = true
				break
			}
		}
	}

	return result
}

// VisitEnumDeclaration 解析枚举声明。
func (p *Parser) VisitEnumDeclaration(node *ast.EnumDeclaration) {
	result := AnalyzeEnumDeclaration(node, p.SourceCode)
	p.Result.EnumDeclarations[result.Identifier] = *result
}