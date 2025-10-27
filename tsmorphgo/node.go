package tsmorphgo

import (
	"context"
	"fmt"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/lsp"
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

// GetQuickInfo 获取该节点的 QuickInfo（类型提示）信息。
// 这个方法提供了类似 VSCode 中悬停提示的功能，显示符号的类型、文档等信息。
// 现在使用原生 TypeScript QuickInfo 集成，可以获取完整的显示部件信息。
//
// 返回值：
//   - *lsp.QuickInfo: 类型提示信息，如果节点没有有效的符号则返回 nil
//   - error: 错误信息
//
// 示例：
//   quickInfo, err := node.GetQuickInfo()
//   if err != nil {
//       return err
//   }
//   if quickInfo != nil {
//       fmt.Printf("类型: %s\n", quickInfo.TypeText)
//       fmt.Printf("显示部件数: %d\n", len(quickInfo.DisplayParts))
//   }
func (n *Node) GetQuickInfo() (*lsp.QuickInfo, error) {
	if n.sourceFile == nil || n.sourceFile.project == nil {
		return nil, fmt.Errorf("node must belong to a source file and project")
	}

	// 创建 LSP 服务来获取 QuickInfo
	lspService, err := createLSPService(n.sourceFile.project)
	if err != nil {
		return nil, fmt.Errorf("failed to create LSP service: %w", err)
	}
	defer lspService.Close()

	// 获取节点的位置信息
	line := n.GetStartLineNumber()
	char := 0 // 使用节点的起始字符位置

	// 使用原生 QuickInfo 集成，获取更完整的显示部件信息
	quickInfo, err := lspService.GetNativeQuickInfoAtPosition(
		context.Background(),
		n.sourceFile.GetFilePath(),
		line,
		char,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get native quick info: %w", err)
	}

	return quickInfo, nil
}

// createLSPService 创建 LSP 服务的辅助函数。
// 这个函数为给定的项目创建一个 LSP 服务实例，用于执行 QuickInfo 查询等操作。
//
// 参数：
//   - project: tsmorphgo 项目实例
//
// 返回值：
//   - *lsp.Service: LSP 服务实例
//   - error: 错误信息
func createLSPService(project *Project) (*lsp.Service, error) {
	if project == nil || project.parserResult == nil {
		return nil, fmt.Errorf("invalid project or parser result")
	}

	// 构建源码映射，使用项目的解析结果
	sources := make(map[string]any, len(project.parserResult.Js_Data))
	for path, jsResult := range project.parserResult.Js_Data {
		sources[path] = jsResult.Raw
	}

	// 创建 LSP 服务（使用测试构造函数，传入内存中的源码映射）
	return lsp.NewServiceForTest(sources)
}
