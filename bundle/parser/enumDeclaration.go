package parser

import (
	"main/bundle/utils"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// 解析 enum 声明
type EnumDeclarationResult struct {
	Name string // 名称
	Raw  string // 源码
}

func NewEnumDeclarationResult(node *ast.EnumDeclaration, sourceCode string) *EnumDeclarationResult {
	raw := utils.GetNodeText(node.AsNode(), sourceCode)

	// 获取枚举的名称节点
	nameNode := node.Name()
	// 如果是标识符节点，返回其文本内容
	if ast.IsIdentifier(nameNode) {
		return &EnumDeclarationResult{
			Name: nameNode.AsIdentifier().Text,
			Raw:  raw,
		}
	}
	return &EnumDeclarationResult{
		Name: "",
		Raw:  raw,
	}
}
