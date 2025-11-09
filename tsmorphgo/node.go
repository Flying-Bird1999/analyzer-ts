package tsmorphgo

import (
	"strings"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// Node 是对原始 ast.Node 的包装，提供了丰富的导航和信息获取 API。
type Node struct {
	// 内嵌 typescript-go 的原始 AST 节点
	*ast.Node
	// 指向其所在的 SourceFile，便于访问文件级和项目级信息
	sourceFile *SourceFile
	// 声明访问器，用于高效访问解析结果（懒加载）
	declarationAccessor DeclarationAccessor
}

// GetSourceFile 返回该节点所属的 SourceFile。
func (n *Node) GetSourceFile() *SourceFile {
	return n.sourceFile
}

// IsValid 检查节点是否有效
func (n *Node) IsValid() bool {
	return n != nil && n.Node != nil
}

// GetText 获取节点的文本内容
func (n *Node) GetText() string {
	if !n.IsValid() {
		return ""
	}
	return n.sourceFile.fileResult.Raw[n.Pos():n.End()]
}

// GetStartLineNumber 获取起始行号（简化版本）
func (n *Node) GetStartLineNumber() int {
	if !n.IsValid() {
		return 0
	}
	// 简单计算行号，适用于基本使用场景
	text := n.sourceFile.fileResult.Raw[:n.Pos()]
	return len(strings.Split(text, "\n"))
}

// GetStartLineCharacter 获取起始列号（简化版本）
func (n *Node) GetStartLineCharacter() int {
	if !n.IsValid() {
		return 0
	}
	// 简单计算列号，适用于基本使用场景
	text := n.sourceFile.fileResult.Raw[:n.Pos()]
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return 0
	}
	return len(lines[len(lines)-1])
}

// GetStart 获取起始位置
func (n *Node) GetStart() int {
	if !n.IsValid() {
		return 0
	}
	return n.Pos()
}

// GetParent 获取父节点
func (n *Node) GetParent() *Node {
	if !n.IsValid() || n.Node.Parent == nil {
		return nil
	}
	return &Node{
		Node:              n.Node.Parent,
		sourceFile:        n.sourceFile,
		declarationAccessor: n.declarationAccessor,
	}
}

// GetStartColumnNumber 获取起始列号（1-based）
func (n *Node) GetStartColumnNumber() int {
	if !n.IsValid() {
		return 0
	}
	// 简单计算列号，适用于基本使用场景
	text := n.sourceFile.fileResult.Raw[:n.Pos()]
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return 1
	}
	return len(lines[len(lines)-1]) + 1
}

// IsIdentifier 检查节点是否为标识符
func IsIdentifier(node Node) bool {
	return node.IsKind(KindIdentifier)
}

// GetAncestors 获取所有祖先节点
func (n *Node) GetAncestors() []*Node {
	if !n.IsValid() {
		return nil
	}

	var ancestors []*Node
	current := n.GetParent()
	for current != nil {
		ancestors = append(ancestors, current)
		current = current.GetParent()
	}
	return ancestors
}

// GetFirstAncestorByKind 根据节点类型查找第一个匹配的祖先节点
func (n *Node) GetFirstAncestorByKind(kind SyntaxKind) (*Node, bool) {
	if !n.IsValid() {
		return nil, false
	}

	current := n.GetParent()
	for current != nil {
		if current.IsKind(kind) {
			return current, true
		}
		current = current.GetParent()
	}
	return nil, false
}

// GetStartLinePos 获取行起始位置
func (n *Node) GetStartLinePos() int {
	if !n.IsValid() {
		return 0
	}
	// 简单计算行起始位置
	text := n.sourceFile.fileResult.Raw[:n.Pos()]
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return 0
	}
	return len(text) - len(lines[len(lines)-1])
}

// ForEachDescendant 遍历所有子孙节点
func (n *Node) ForEachDescendant(callback func(Node)) {
	if !n.IsValid() {
		return
	}

	// 使用SourceFile的ForEachDescendant方法，并过滤出当前节点的子孙节点
	if n.sourceFile != nil {
		n.sourceFile.ForEachDescendant(func(node Node) {
			if isDescendantOf(node, *n) {
				callback(node)
			}
		})
	}
}

// 辅助函数：检查node是否是ancestor的子孙节点
func isDescendantOf(node, ancestor Node) bool {
	current := node.GetParent()
	for current != nil {
		if current.GetStart() == ancestor.GetStart() &&
		   current.GetEnd() == ancestor.GetEnd() {
			return true
		}
		current = current.GetParent()
	}
	return false
}

// GetEnd 获取结束位置
func (n *Node) GetEnd() int {
	if !n.IsValid() {
		return 0
	}
	return n.End()
}