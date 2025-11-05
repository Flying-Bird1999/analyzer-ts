package tsmorphgo

import (
	"context"
	"fmt"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/lsp"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"
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

// GetStartLineCharacter 返回节点在源文件中的起始列号 (0-based)。
func (n *Node) GetStartLineCharacter() int {
	if n.sourceFile == nil || n.sourceFile.fileResult == nil {
		return 0
	}
	_, char := utils.GetLineAndCharacterOfPosition(n.sourceFile.fileResult.Raw, n.Pos())
	return char
}

// GetStart 返回节点在文件中的起始字符偏移量 (0-based)。
func (n *Node) GetStart() int {
	return n.Pos()
}

// GetFirstChild 按条件查找并返回第一个匹配的直接子节点。
func GetFirstChild(node Node, predicate func(child Node) bool) (*Node, bool) {
	var foundNode *Node
	node.Node.ForEachChild(func(child *ast.Node) bool {
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
//
//	quickInfo, err := node.GetQuickInfo()
//	if err != nil {
//	    return err
//	}
//	if quickInfo != nil {
//	    fmt.Printf("类型: %s\n", quickInfo.TypeText)
//	    fmt.Printf("显示部件数: %d\n", len(quickInfo.DisplayParts))
//	}
func (n *Node) GetQuickInfo() (*lsp.QuickInfo, error) {
	if n.sourceFile == nil || n.sourceFile.project == nil {
		return nil, fmt.Errorf("node must belong to a source file and project")
	}

	// 从项目获取共享的 LSP 服务
	lspService, err := n.sourceFile.project.getLspService()
	if err != nil {
		return nil, fmt.Errorf("failed to get LSP service: %w", err)
	}

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

// GetEndLineNumber 返回节点在源文件中的结束行号 (1-based)。
func (n *Node) GetEndLineNumber() int {
	if n.sourceFile == nil || n.sourceFile.fileResult == nil {
		return 0
	}
	line, _ := utils.GetLineAndCharacterOfPosition(n.sourceFile.fileResult.Raw, n.End())
	return line + 1
}

// GetEnd 返回节点在文件中的结束字符偏移量 (0-based)。
func (n *Node) GetEnd() int {
	return n.End()
}

// GetTextLength 返回节点文本的长度。
func (n *Node) GetTextLength() int {
	return n.End() - n.Pos()
}

// GetChildCount 返回子节点的数量。
func (n *Node) GetChildCount() int {
	count := 0
	n.Node.ForEachChild(func(child *ast.Node) bool {
		count++
		return false // 继续遍历
	})
	return count
}

// GetChildAt 返回指定索引位置的子节点。
func (n *Node) GetChildAt(index int) (*Node, bool) {
	if index < 0 {
		return nil, false
	}

	currentIndex := 0
	var foundNode *ast.Node
	n.ForEachChild(func(child *ast.Node) bool {
		if currentIndex == index {
			foundNode = child
			return true // 找到了，停止遍历
		}
		currentIndex++
		return false // 继续遍历
	})

	if foundNode != nil {
		return &Node{Node: foundNode, sourceFile: n.sourceFile}, true
	}
	return nil, false
}

// ForEachDescendant 遍历所有后代节点（包括自身）。
// 这是对底层 ForEachChild 的封装，提供更直观的API。
func (n *Node) ForEachDescendant(callback func(Node)) {
	// 遍历自身
	callback(*n)

	// 遍历所有子节点
	n.Node.ForEachChild(func(child *ast.Node) bool {
		wrappedChild := Node{Node: child, sourceFile: n.sourceFile}
		wrappedChild.ForEachDescendant(callback)
		return false // 继续遍历其他子节点
	})
}

// ContainsString 检查节点文本是否包含指定的字符串。
// 这是对 strings.Contains 的包装，方便在AST分析中使用。
func (n *Node) ContainsString(substr string) bool {
	nodeText := n.GetText()
	return strings.Contains(nodeText, substr)
}

// GetNextSibling 返回下一个兄弟节点。
func (n *Node) GetNextSibling() (*Node, bool) {
	if n.GetParent() == nil {
		return nil, false
	}

	parent := n.GetParent()
	foundCurrent := false

	for i := 0; i < parent.GetChildCount(); i++ {
		sibling, ok := parent.GetChildAt(i)
		if !ok {
			continue
		}

		if foundCurrent {
			return sibling, true
		}

		if sibling.Node == n.Node {
			foundCurrent = true
		}
	}

	return nil, false
}

// GetPreviousSibling 返回上一个兄弟节点。
func (n *Node) GetPreviousSibling() (*Node, bool) {
	if n.GetParent() == nil {
		return nil, false
	}

	parent := n.GetParent()
	var previousSibling *Node

	for i := 0; i < parent.GetChildCount(); i++ {
		sibling, ok := parent.GetChildAt(i)
		if !ok {
			continue
		}

		if sibling.Node == n.Node {
			if previousSibling != nil {
				return previousSibling, true
			}
			break
		}

		previousSibling = sibling
	}

	return nil, false
}

// FindChildren 查找所有匹配指定类型的子节点。
func (n *Node) FindChildren(kind ast.Kind) []Node {
	var children []Node
	n.Node.ForEachChild(func(child *ast.Node) bool {
		if child.Kind == kind {
			children = append(children, Node{Node: child, sourceFile: n.sourceFile})
		}
		return false // 继续遍历
	})
	return children
}

// GetFirstChild 返回第一个子节点。
func (n *Node) GetFirstChild() (*Node, bool) {
	var firstChild *ast.Node
	n.Node.ForEachChild(func(child *ast.Node) bool {
		firstChild = child
		return true // 找到第一个就停止
	})
	if firstChild != nil {
		return &Node{Node: firstChild, sourceFile: n.sourceFile}, true
	}
	return nil, false
}

// GetLastChild 返回最后一个子节点。
func (n *Node) GetLastChild() (*Node, bool) {
	var lastChild *ast.Node
	n.Node.ForEachChild(func(child *ast.Node) bool {
		lastChild = child
		return false // 继续遍历，最后一个会被保留
	})
	if lastChild != nil {
		return &Node{Node: lastChild, sourceFile: n.sourceFile}, true
	}
	return nil, false
}

// FindFirstChild 查找第一个匹配指定类型的子节点。
func (n *Node) FindFirstChild(kind ast.Kind) (*Node, bool) {
	var foundNode *ast.Node
	n.Node.ForEachChild(func(child *ast.Node) bool {
		if child.Kind == kind {
			foundNode = child
			return true // 找到了，停止遍历
		}
		return false // 继续遍历
	})
	if foundNode != nil {
		return &Node{Node: foundNode, sourceFile: n.sourceFile}, true
	}
	return nil, false
}

// IsValid 检查节点是否有效（非nil）。
func (n *Node) IsValid() bool {
	return n != nil && n.Node != nil
}

// Contains 检查该节点是否包含另一个节点。
func (n *Node) Contains(other Node) bool {
	return n.Pos() <= other.Pos() && other.End() <= n.End()
}

// GetStartColumnNumber 返回节点在源文件中的起始列号 (1-based)。
func (n *Node) GetStartColumnNumber() int {
	if n.sourceFile == nil || n.sourceFile.fileResult == nil {
		return 0
	}
	_, char := utils.GetLineAndCharacterOfPosition(n.sourceFile.fileResult.Raw, n.Pos())
	return char + 1
}

// GetEndColumnNumber 返回节点在源文件中的结束列号 (1-based)。
func (n *Node) GetEndColumnNumber() int {
	if n.sourceFile == nil || n.sourceFile.fileResult == nil {
		return 0
	}
	_, char := utils.GetLineAndCharacterOfPosition(n.sourceFile.fileResult.Raw, n.End())
	return char + 1
}

// GetKindName 将 ast.Kind 转换为可读的字符串名称。
// 这个方法提供了对调试和日志输出非常有用的节点类型信息。
// 支持所有常用的 TypeScript 语法节点类型。
//
// 返回值：
//   - string: 节点类型的可读名称，如 "Identifier"、"CallExpression" 等
//
// 示例：
//
//	node.GetKindName() // 返回 "Identifier"
func (n *Node) GetKindName() string {
	kind := n.Kind
	switch kind {
	case ast.KindIdentifier:
		return "Identifier"
	case ast.KindCallExpression:
		return "CallExpression"
	case ast.KindPropertyAccessExpression:
		return "PropertyAccessExpression"
	case ast.KindVariableDeclaration:
		return "VariableDeclaration"
	case ast.KindFunctionDeclaration:
		return "FunctionDeclaration"
	case ast.KindMethodDeclaration:
		return "MethodDeclaration"
	case ast.KindInterfaceDeclaration:
		return "InterfaceDeclaration"
	case ast.KindTypeAliasDeclaration:
		return "TypeAliasDeclaration"
	case ast.KindClassDeclaration:
		return "ClassDeclaration"
	case ast.KindEnumDeclaration:
		return "EnumDeclaration"
	case ast.KindBinaryExpression:
		return "BinaryExpression"
	case ast.KindObjectLiteralExpression:
		return "ObjectLiteralExpression"
	case ast.KindArrayLiteralExpression:
		return "ArrayLiteralExpression"
	case ast.KindPropertyAssignment:
		return "PropertyAssignment"
	case ast.KindPropertyDeclaration:
		return "PropertyDeclaration"
	case ast.KindImportSpecifier:
		return "ImportSpecifier"
	case ast.KindImportClause:
		return "ImportClause"
	case ast.KindImportDeclaration:
		return "ImportDeclaration"
	case ast.KindExportDeclaration:
		return "ExportDeclaration"
	case ast.KindExportAssignment:
		return "ExportAssignment"
	case ast.KindEqualsToken:
		return "EqualsToken"
	case ast.KindPlusToken:
		return "PlusToken"
	case ast.KindMinusToken:
		return "MinusToken"
	case ast.KindAsteriskToken:
		return "AsteriskToken"
	case ast.KindSlashToken:
		return "SlashToken"
	case ast.KindSourceFile:
		return "SourceFile"
	case ast.KindTypeReference:
		return "TypeReference"
	case ast.KindTypeParameter:
		return "TypeParameter"
	case ast.KindConstructor:
		return "Constructor"
	case ast.KindGetAccessor:
		return "GetAccessor"
	case ast.KindSetAccessor:
		return "SetAccessor"
	case ast.KindTypeAssertionExpression:
		return "TypeAssertionExpression"
	case ast.KindParameter:
		return "Parameter"
	// 添加更多类型支持...
	default:
		return fmt.Sprintf("Unknown(%d)", kind)
	}
}

// GetStartLinePos 返回节点所在行的起始字符位置 (0-based)。
// 这个方法对于计算节点的精确列位置非常重要，通常与
// GetStartLineNumber() 和 GetStartColumnNumber() 配合使用。
//
// 算法说明：
//   - 从节点的起始位置 (GetStart()) 向前查找
//   - 找到最近的换行符 (\n)
//   - 如果没有找到换行符，说明节点在第一行，返回 0
//
// 返回值：
//   - int: 节点所在行的起始字符位置 (0-based)
//
// 示例：
//
//	文件内容: "const x = 1;\nconst y = 2;"
//	对于第二个 const 声明:
//	- GetStart() 返回 14
//	- GetStartLinePos() 返回 13 (第二行起始位置)
//	- 列号 = 14 - 13 = 1
func (n *Node) GetStartLinePos() int {
	if n.sourceFile == nil || n.sourceFile.fileResult == nil {
		return 0
	}

	content := n.sourceFile.fileResult.Raw
	startPos := n.GetStart()

	// 从起始位置向前查找，找到最近的换行符
	for i := startPos - 1; i >= 0; i-- {
		if content[i] == '\n' {
			return i + 1 // 返回换行符的下一个位置
		}
	}

	// 如果没有找到换行符，说明节点在第一行
	return 0
}

// PositionInfo 包含节点的完整位置信息，提供比单独方法
// 更全面的节点位置描述。这对于调试、日志记录和位置
// 相关的操作非常有用。
type PositionInfo struct {
	Line        int // 节点所在行号 (1-based)
	Column      int // 节点所在列号 (1-based)
	StartOffset int // 节点在文件中的起始字符偏移量 (0-based)
	EndOffset   int // 节点在文件中的结束字符偏移量 (0-based)
	StartLinePos int // 节点所在行的起始字符偏移量 (0-based)
}

// GetPositionInfo 返回节点的完整位置信息。
// 这个方法整合了所有位置相关的计算，返回一个结构体
// 包含了节点的行、列、偏移量等所有位置信息。
//
// 返回值：
//   - *PositionInfo: 包含完整位置信息的结构体，如果节点无效则返回 nil
//
// 示例：
//
//	pos := node.GetPositionInfo()
//	if pos != nil {
//	    fmt.Printf("位置: %d:%d, 偏移量: %d-%d\n",
//	        pos.Line, pos.Column, pos.StartOffset, pos.EndOffset)
//	}
func (n *Node) GetPositionInfo() *PositionInfo {
	if n.sourceFile == nil || n.sourceFile.fileResult == nil {
		return nil
	}

	content := n.sourceFile.fileResult.Raw
	start := n.GetStart()
	end := n.GetEnd()

	// 获取行号和列号
	line, char := utils.GetLineAndCharacterOfPosition(content, start)
	startLinePos := n.GetStartLinePos()

	return &PositionInfo{
		Line:        line + 1,    // 转换为 1-based
		Column:      char + 1,    // 转换为 1-based
		StartOffset: start,      // 已是 0-based
		EndOffset:   end,        // 已是 0-based
		StartLinePos: startLinePos, // 已是 0-based
	}
}


// GetDeclarationAccessor 获取节点的声明访问器。
// 如果访问器尚未初始化，则会创建一个新的访问器实例。
// 这个方法提供了对 analyzer/parser 能力的统一访问接口。
//
// 返回值：
//   - DeclarationAccessor: 声明访问器实例
func (n *Node) GetDeclarationAccessor() DeclarationAccessor {
	if n.declarationAccessor == nil {
		n.declarationAccessor = NewDeclarationAccessor(n.sourceFile)
	}
	return n.declarationAccessor
}

// AsVariableDeclarationOptimized 使用优化的声明访问器获取变量声明信息。
// 这个方法集成了 analyzer/parser 的能力，提供更好的性能和功能。
//
// 返回值：
//   - *parser.VariableDeclaration: 变量声明信息，如果节点不是变量声明则返回 nil
//   - bool: 转换是否成功
func (n *Node) AsVariableDeclarationOptimized() (*parser.VariableDeclaration, bool) {
	if n.sourceFile == nil {
		return nil, false
	}
	accessor := n.GetDeclarationAccessor()
	return accessor.GetVariableDeclaration(n.Node)
}

// AsFunctionDeclarationOptimized 使用优化的声明访问器获取函数声明信息。
// 这个方法集成了 analyzer/parser 的能力，提供更好的性能和功能。
//
// 返回值：
//   - *parser.FunctionDeclarationResult: 函数声明信息，如果节点不是函数声明则返回 nil
//   - bool: 转换是否成功
func (n *Node) AsFunctionDeclarationOptimized() (*parser.FunctionDeclarationResult, bool) {
	if n.sourceFile == nil {
		return nil, false
	}
	accessor := n.GetDeclarationAccessor()
	return accessor.GetFunctionDeclaration(n.Node)
}

// AsInterfaceDeclarationOptimized 使用优化的声明访问器获取接口声明信息。
// 这个方法集成了 analyzer/parser 的能力，提供更好的性能和功能。
//
// 返回值：
//   - *parser.InterfaceDeclarationResult: 接口声明信息，如果节点不是接口声明则返回 nil
//   - bool: 转换是否成功
func (n *Node) AsInterfaceDeclarationOptimized() (*parser.InterfaceDeclarationResult, bool) {
	if n.sourceFile == nil {
		return nil, false
	}
	accessor := n.GetDeclarationAccessor()
	return accessor.GetInterfaceDeclaration(n.Node)
}

// AsImportDeclarationOptimized 使用优化的声明访问器获取导入声明信息。
// 这个方法集成了 analyzer/parser 的能力，提供更好的性能和功能。
//
// 返回值：
//   - *projectParser.ImportDeclarationResult: 导入声明信息，如果节点不是导入声明则返回 nil
//   - bool: 转换是否成功
func (n *Node) AsImportDeclarationOptimized() (*projectParser.ImportDeclarationResult, bool) {
	if n.sourceFile == nil {
		return nil, false
	}
	accessor := n.GetDeclarationAccessor()
	return accessor.GetImportDeclaration(n.Node)
}

// AsTypeDeclarationOptimized 使用优化的声明访问器获取类型别名声明信息。
// 这个方法集成了 analyzer/parser 的能力，提供更好的性能和功能。
//
// 返回值：
//   - *parser.TypeDeclarationResult: 类型别名声明信息，如果节点不是类型别名声明则返回 nil
//   - bool: 转换是否成功
func (n *Node) AsTypeDeclarationOptimized() (*parser.TypeDeclarationResult, bool) {
	if n.sourceFile == nil {
		return nil, false
	}
	accessor := n.GetDeclarationAccessor()
	return accessor.GetTypeDeclaration(n.Node)
}

// AsEnumDeclarationOptimized 使用优化的声明访问器获取枚举声明信息。
// 这个方法集成了 analyzer/parser 的能力，提供更好的性能和功能。
//
// 返回值：
//   - *parser.EnumDeclarationResult: 枚举声明信息，如果节点不是枚举声明则返回 nil
//   - bool: 转换是否成功
func (n *Node) AsEnumDeclarationOptimized() (*parser.EnumDeclarationResult, bool) {
	if n.sourceFile == nil {
		return nil, false
	}
	accessor := n.GetDeclarationAccessor()
	return accessor.GetEnumDeclaration(n.Node)
}

// GetDeclaration 通用声明获取方法。
// 根据节点类型自动选择相应的声明获取方法，提供统一的访问接口。
//
// 返回值：
//   - interface{}: 声明信息（具体类型取决于节点类型）
//   - bool: 获取是否成功
//   - string: 声明类型名称（如 "VariableDeclaration"、"FunctionDeclaration" 等）
func (n *Node) GetDeclaration() (interface{}, bool, string) {
	if n.sourceFile == nil {
		return nil, false, "Unknown"
	}
	accessor := n.GetDeclarationAccessor()
	return accessor.GetDeclaration(n.Node)
}

// IsDeclarationType 检查节点是否是指定类型的声明。
// 这个方法提供了统一的类型检查接口，集成了 analyzer/parser 的能力。
//
// 参数：
//   - kind: ast.Kind 要检查的节点类型
//
// 返回值：
//   - bool: 节点是否是指定类型的声明
func (n *Node) IsDeclarationType(kind ast.Kind) bool {
	if n.sourceFile == nil {
		return false
	}
	accessor := n.GetDeclarationAccessor()

	switch kind {
	case ast.KindVariableDeclaration, ast.KindVariableDeclarationList:
		return accessor.IsVariableDeclaration(n.Node)
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression:
		return accessor.IsFunctionDeclaration(n.Node)
	case ast.KindInterfaceDeclaration:
		return accessor.IsInterfaceDeclaration(n.Node)
	case ast.KindImportDeclaration:
		return accessor.IsImportDeclaration(n.Node)
	case ast.KindTypeAliasDeclaration:
		return accessor.IsTypeDeclaration(n.Node)
	case ast.KindEnumDeclaration:
		return accessor.IsEnumDeclaration(n.Node)
	default:
		return false
	}
}

// GetDeclarationType 获取节点的声明类型名称。
// 这个方法提供了统一的方式来获取节点的声明类型。
//
// 返回值：
//   - string: 声明类型名称，如果节点不是声明类型则返回 "Unknown"
func (n *Node) GetDeclarationType() string {
	if n.sourceFile == nil {
		return "Unknown"
	}
	accessor := n.GetDeclarationAccessor()
	_, _, typeName := accessor.GetDeclaration(n.Node)
	return typeName
}

// CreateTestProject 创建 LSP 服务的辅助函数。
// 这个函数为给定的项目创建一个 LSP 服务实例，用于执行 QuickInfo 查询等操作。
//
// 参数：
//   - project: tsmorphgo 项目实例
//
// 返回值：
//   - *lsp.Service: LSP 服务实例
//   - error: 错误信息
func CreateTestProject(project *Project) (*lsp.Service, error) {
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
