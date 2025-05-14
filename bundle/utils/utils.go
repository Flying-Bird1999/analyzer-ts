package utils

import (
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// 获取ast节点的原始源代码文本
func GetNodeText(node *ast.Node, sourceCode string) string {
	start := node.Pos()
	end := node.End()
	if start >= 0 && end >= start && end <= len(sourceCode) {
		return sourceCode[start:end]
	}
	return ""
}
