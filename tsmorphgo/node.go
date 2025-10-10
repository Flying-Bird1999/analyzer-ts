package tsmorphgo

import (
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// Node 是对原始 ast.Node 的包装，提供了丰富的导航和信息获取 API。
type Node struct {
	// 内嵌 typescript-go 的原始 AST 节点
	*ast.Node
	// 指向其所在的 SourceFile，便于访问文件级和项目级信息
	sourceFile *SourceFile
}

// GetSourceFile 返回该节点所属的 SourceFile。
func (n *Node) GetSourceFile() *SourceFile {
	return n.sourceFile
}

// GetParent 返回该节点的直接父节点。
// 它直接利用 `typescript-go` AST 内置的父节点引用，效率高且实现简单。
func (n *Node) GetParent() *Node {
	if parentAstNode := n.Node.Parent; parentAstNode != nil {
		return &Node{
			Node:       parentAstNode,
			sourceFile: n.sourceFile,
		}
	}
	return nil
}

// GetAncestors 返回从当前节点的父节点到根节点的所有祖先节点数组。
func (n *Node) GetAncestors() []*Node {
	ancestors := []*Node{}
	for parent := n.GetParent(); parent != nil; parent = parent.GetParent() {
		ancestors = append(ancestors, parent)
	}
	return ancestors
}

// GetFirstAncestorByKind 向上遍历，返回第一个匹配指定类型的祖先节点。
func (n *Node) GetFirstAncestorByKind(kind ast.Kind) (*Node, bool) {
	for parent := n.GetParent(); parent != nil; parent = parent.GetParent() {
		if parent.Kind == kind {
			return parent, true
		}
	}
	return nil, false
}

// GetText 返回此节点在源码中的原始文本。
func (n *Node) GetText() string {
	if n.sourceFile == nil || n.sourceFile.fileResult == nil {
		return ""
	}
	return utils.GetNodeText(n.Node, n.sourceFile.fileResult.Raw)
}

// GetStartLineNumber 返回节点在源文件中的起始行号 (1-based)。
func (n *Node) GetStartLineNumber() int {
	if n.sourceFile == nil || n.sourceFile.fileResult == nil {
		return 0
	}
	line, _ := utils.GetLineAndCharacterOfPosition(n.sourceFile.fileResult.Raw, n.Pos())
	return line + 1
}

// GetStart 返回节点在文件中的起始字符偏移量 (0-based)。
func (n *Node) GetStart() int {
	return n.Pos()
}

// GetFirstChild 按条件查找并返回第一个匹配的直接子节点。
func GetFirstChild(node Node, predicate func(child Node) bool) (*Node, bool) {
	var foundNode *Node
	node.ForEachChild(func(child *ast.Node) bool {
		wrappedChild := Node{Node: child, sourceFile: node.sourceFile}
		if predicate(wrappedChild) {
			foundNode = &wrappedChild
			return true // 找到了，停止遍历
		}
		return false // 继续遍历
	})

	if foundNode != nil {
		return foundNode, true
	}
	return nil, false
}
