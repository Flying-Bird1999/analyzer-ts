package tsmorphgo

import (
	"strings"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// =============================================================================
// Node 结构定义 - 来自原 node.go
// =============================================================================

// Node 是对原始 ast.Node 的包装，提供了丰富的导航和信息获取 API。
type Node struct {
	// 内嵌 typescript-go 的原始 AST 节点
	*ast.Node
	// 指向其所在的 SourceFile，便于访问文件级和项目级信息
	sourceFile *SourceFile
}

// =============================================================================
// 节点类别定义 - 来自原 node_unified.go
// =============================================================================

// NodeCategory 定义节点类别，便于批量判断
type NodeCategory struct {
	name  string
	kinds map[SyntaxKind]bool
}

// 常用节点类别
var (
	// CategoryDeclarations 声明类节点
	CategoryDeclarations = NodeCategory{
		name: "declarations",
		kinds: map[SyntaxKind]bool{
			KindFunctionDeclaration:  true,
			KindVariableDeclaration:  true,
			KindInterfaceDeclaration: true,
			KindClassDeclaration:     true,
			KindTypeAliasDeclaration: true,
			KindEnumDeclaration:      true,
			KindMethodDeclaration:    true,
			KindConstructor:          true,
			KindGetAccessor:          true,
			KindSetAccessor:          true,
		},
	}

	// CategoryExpressions 表达式类节点
	CategoryExpressions = NodeCategory{
		name: "expressions",
		kinds: map[SyntaxKind]bool{
			KindCallExpression:           true,
			KindPropertyAccessExpression: true,
			KindBinaryExpression:         true,
			KindConditionalExpression:    true,
			KindUnaryExpression:          true,
			KindObjectLiteralExpression:  true,
			KindArrayLiteralExpression:   true,
			KindTemplateExpression:       true,
			KindYieldExpression:          true,
			KindAwaitExpression:          true,
			KindTypeAssertionExpression:  true,
			KindSpreadElement:            true,
		},
	}

	// CategoryStatements 语句类节点
	CategoryStatements = NodeCategory{
		name: "statements",
		kinds: map[SyntaxKind]bool{
			KindVariableStatement: true,
			KindReturnStatement:   true,
			KindIfStatement:       true,
			KindForStatement:      true,
			KindWhileStatement:    true,
			KindTryStatement:      true,
			KindCatchClause:       true,
		},
	}

	// CategoryTypes 类型相关节点
	CategoryTypes = NodeCategory{
		name: "types",
		kinds: map[SyntaxKind]bool{
			KindInterfaceDeclaration: true,
			KindTypeAliasDeclaration: true,
			KindEnumDeclaration:      true,
			KindTypeReference:        true,
			KindTypeParameter:        true,
		},
	}

	// CategoryLiterals 字面量节点
	CategoryLiterals = NodeCategory{
		name: "literals",
		kinds: map[SyntaxKind]bool{
			KindStringLiteral:     true,
			KindNumericLiteral:    true,
			KindTrueKeyword:       true,
			KindFalseKeyword:      true,
			KindNullKeyword:       true,
			KindUndefinedKeyword:  true,
		},
	}

	// CategoryModules 模块相关节点
	CategoryModules = NodeCategory{
		name: "modules",
		kinds: map[SyntaxKind]bool{
			KindImportDeclaration: true,
			KindExportDeclaration: true,
			KindImportClause:      true,
			KindImportSpecifier:   true,
			KindExportSpecifier:   true,
		},
	}

	// CategoryIdentifiers 标识符相关节点
	CategoryIdentifiers = NodeCategory{
		name: "identifiers",
		kinds: map[SyntaxKind]bool{
			KindIdentifier: true,
		},
	}
)

// Contains 检查指定类型是否属于此类别
func (c NodeCategory) Contains(kind SyntaxKind) bool {
	return c.kinds[kind]
}

// Name 返回类别名称
func (c NodeCategory) Name() string {
	return c.name
}

// Kinds 返回类别中包含的所有类型
func (c NodeCategory) Kinds() []SyntaxKind {
	kinds := make([]SyntaxKind, 0, len(c.kinds))
	for kind := range c.kinds {
		kinds = append(kinds, kind)
	}
	return kinds
}

// =============================================================================
// 基础导航和信息获取方法 - 来自原 node.go
// =============================================================================

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
	text := n.sourceFile.fileResult.Raw[:n.Pos()]
	return len(strings.Split(text, "\n"))
}

// GetStartLineCharacter 获取起始列号（简化版本）
func (n *Node) GetStartLineCharacter() int {
	if !n.IsValid() {
		return 0
	}
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
		Node:       n.Node.Parent,
		sourceFile: n.sourceFile,
	}
}

// GetStartColumnNumber 获取起始列号（1-based）
func (n *Node) GetStartColumnNumber() int {
	if !n.IsValid() {
		return 0
	}
	text := n.sourceFile.fileResult.Raw[:n.Pos()]
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return 1
	}
	return len(lines[len(lines)-1]) + 1
}

// GetEnd 获取结束位置
func (n *Node) GetEnd() int {
	if !n.IsValid() {
		return 0
	}
	return n.End()
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

// =============================================================================
// 统一的类型检查API - 来自 node_api_clean.go
// =============================================================================

// GetKind 返回节点的类型（统一接口）
func (n Node) GetKind() SyntaxKind {
	return SyntaxKind(n.Kind)
}

// IsKind 检查节点是否为指定类型
func (n Node) IsKind(kind SyntaxKind) bool {
	return n.Kind == ast.Kind(kind)
}

// IsAnyKind 检查节点是否为任意一个指定类型
func (n Node) IsAnyKind(kinds ...SyntaxKind) bool {
	for _, kind := range kinds {
		if n.Kind == ast.Kind(kind) {
			return true
		}
	}
	return false
}

// IsCategory 检查节点是否属于指定类别
func (n Node) IsCategory(category NodeCategory) bool {
	return category.Contains(SyntaxKind(n.Kind))
}

// =============================================================================
// 便捷的类别检查方法 - 来自 node_api_clean.go
// =============================================================================

// IsDeclaration 检查是否为声明类节点
func (n Node) IsDeclaration() bool {
	return n.IsCategory(CategoryDeclarations)
}

// IsExpression 检查是否为表达式类节点
func (n Node) IsExpression() bool {
	return n.IsCategory(CategoryExpressions)
}

// IsStatement 检查是否为语句类节点
func (n Node) IsStatement() bool {
	return n.IsCategory(CategoryStatements)
}

// IsType 检查是否为类型相关节点
func (n Node) IsType() bool {
	return n.IsCategory(CategoryTypes)
}

// IsLiteral 检查是否为字面量节点
func (n Node) IsLiteral() bool {
	return n.IsCategory(CategoryLiterals)
}

// IsModule 检查是否为模块相关节点
func (n Node) IsModule() bool {
	return n.IsCategory(CategoryModules)
}

// IsIdentifierNode 检查是否为标识符
func (n Node) IsIdentifierNode() bool {
	return n.Kind == ast.KindIdentifier
}

// =============================================================================
// 常用类型的便捷检查方法 - 来自 node_api_clean.go
// =============================================================================

// IsFunctionDeclaration 检查是否为函数声明
func (n Node) IsFunctionDeclaration() bool {
	return n.Kind == ast.KindFunctionDeclaration
}

// IsVariableDeclaration 检查是否为变量声明
func (n Node) IsVariableDeclaration() bool {
	return n.Kind == ast.KindVariableDeclaration
}

// IsInterfaceDeclaration 检查是否为接口声明
func (n Node) IsInterfaceDeclaration() bool {
	return n.Kind == ast.KindInterfaceDeclaration
}

// IsClassDeclaration 检查是否为类声明
func (n Node) IsClassDeclaration() bool {
	return n.Kind == ast.KindClassDeclaration
}

// IsCallExpr 检查是否为函数调用表达式
func (n Node) IsCallExpr() bool {
	return n.Kind == ast.KindCallExpression
}

// IsPropertyAccessExpression 检查是否为属性访问表达式
func (n Node) IsPropertyAccessExpression() bool {
	return n.Kind == ast.KindPropertyAccessExpression
}

// IsImportDeclaration 检查是否为导入声明
func (n Node) IsImportDeclaration() bool {
	return n.Kind == ast.KindImportDeclaration
}

// IsExportDeclaration 检查是否为导出声明
func (n Node) IsExportDeclaration() bool {
	return n.Kind == ast.KindExportDeclaration
}

// =============================================================================
// 类型转换的统一接口 - 来自 node_api_clean.go
// =============================================================================

// AsDeclaration 尝试转换为声明类结果
func (n Node) AsDeclaration() (interface{}, bool) {
	switch n.GetKind() {
	case KindImportDeclaration:
		return AsImportDeclaration(n)
	case KindVariableDeclaration:
		return AsVariableDeclaration(n)
	case KindFunctionDeclaration:
		return AsFunctionDeclaration(n)
	case KindInterfaceDeclaration:
		return AsInterfaceDeclaration(n)
	case KindTypeAliasDeclaration:
		return AsTypeAliasDeclaration(n)
	case KindEnumDeclaration:
		return AsEnumDeclaration(n)
	default:
		return nil, false
	}
}

// AsNode 尝试转换为特定类型的节点
func (n Node) AsNode(kind SyntaxKind) (*Node, bool) {
	if n.Kind == ast.Kind(kind) {
		return &n, true
	}
	return nil, false
}

// =============================================================================
// 名称和值获取的统一接口 - 来自 node_api_clean.go
// =============================================================================

// GetNodeName 获取节点名称（统一接口）
func (n Node) GetNodeName() (string, bool) {
	text := n.GetText()
	if text == "" {
		return "", false
	}

	switch n.GetKind() {
	case KindFunctionDeclaration, KindInterfaceDeclaration,
		KindClassDeclaration, KindTypeAliasDeclaration, KindEnumDeclaration:
		return extractNameFromText(text, n.GetKind())
	case KindIdentifier:
		return text, true
	case KindVariableDeclaration:
		if name, ok := extractVariableNameFromText(text); ok {
			return name, true
		}
		return "", false
	default:
		return "", false
	}
}

// GetLiteralValue 获取字面量值
func (n Node) GetLiteralValue() (interface{}, bool) {
	if !n.IsLiteral() {
		return nil, false
	}

	text := n.GetText()
	switch n.GetKind() {
	case KindStringLiteral:
		return extractStringValue(text), true
	case KindNumericLiteral:
		return extractNumericValue(text), true
	case KindTrueKeyword:
		return true, true
	case KindFalseKeyword:
		return false, true
	case KindNullKeyword, KindUndefinedKeyword:
		return nil, true
	default:
		return text, true
	}
}

// =============================================================================
// 辅助函数 - 来自 node_api_clean.go
// =============================================================================

// extractNameFromText 从节点文本中提取名称
func extractNameFromText(text string, kind SyntaxKind) (string, bool) {
	switch kind {
	case KindFunctionDeclaration:
		if startsWith(text, "function ") {
			name := trimPrefix(text, "function ")
			return extractFirstWord(name), true
		}
	case KindInterfaceDeclaration:
		if startsWith(text, "interface ") {
			name := trimPrefix(text, "interface ")
			return extractFirstWord(name), true
		}
	case KindClassDeclaration:
		if startsWith(text, "class ") {
			name := trimPrefix(text, "class ")
			return extractFirstWord(name), true
		}
	case KindTypeAliasDeclaration:
		if startsWith(text, "type ") {
			name := trimPrefix(text, "type ")
			return extractFirstWord(name), true
		}
	case KindEnumDeclaration:
		if startsWith(text, "enum ") {
			name := trimPrefix(text, "enum ")
			return extractFirstWord(name), true
		}
	case KindVariableDeclaration:
		return extractFirstIdentifier(text), true
	}
	return "", false
}

// extractVariableNameFromText 从文本中提取变量名
func extractVariableNameFromText(text string) (string, bool) {
	var name []rune
	for _, r := range text {
		if isIdentifierChar(r) {
			name = append(name, r)
		} else if len(name) > 0 {
			return string(name), true
		} else if r == '=' || r == '{' || r == '[' {
			return "destructured pattern", true
		}
	}

	if len(name) > 0 {
		return string(name), true
	}
	return "", false
}

// extractStringValue 从字面量文本中提取字符串值
func extractStringValue(text string) string {
	if len(text) >= 2 && (text[0] == '"' || text[0] == '\'') {
		return text[1 : len(text)-1]
	}
	return text
}

// extractNumericValue 从字面量文本中提取数值
func extractNumericValue(text string) interface{} {
	return text // 简单实现，实际中可以解析为 int 或 float
}

// isIdentifierChar 检查字符是否为标识符字符
func isIdentifierChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '$'
}

// startsWith 检查字符串是否以指定前缀开始
func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// trimPrefix 移除字符串的前缀
func trimPrefix(s, prefix string) string {
	if startsWith(s, prefix) {
		return s[len(prefix):]
	}
	return s
}

// extractFirstWord 提取第一个单词
func extractFirstWord(s string) string {
	if len(s) == 0 {
		return s
	}
	for i, r := range s {
		if isSpace(r) || r == '{' || r == '(' || r == '=' || r == ':' {
			if i < len(s) {
				return s[:i]
			}
		}
	}
	return s
}

// extractFirstIdentifier 提取第一个标识符
func extractFirstIdentifier(s string) string {
	var result []rune
	for i, r := range s {
		if isIdentifierChar(r) {
			result = append(result, r)
		} else if len(result) > 0 {
			return s[:i]
		}
	}
	return string(result)
}

// isSpace 检查字符是否为空白字符
func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// =============================================================================
// 全局辅助函数 - 来自原 node.go
// =============================================================================

// IsIdentifier 检查节点是否为标识符 (全局函数)
func IsIdentifier(node Node) bool {
	return node.IsKind(KindIdentifier)
}