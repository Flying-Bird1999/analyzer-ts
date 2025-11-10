package tsmorphgo

import (
	"fmt"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
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
			KindStringLiteral:    true,
			KindNumericLiteral:   true,
			KindTrueKeyword:      true,
			KindFalseKeyword:     true,
			KindNullKeyword:      true,
			KindUndefinedKeyword: true,
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
		return n.AsImportDeclaration()
	case KindVariableDeclaration:
		return n.AsVariableDeclaration()
	case KindFunctionDeclaration:
		return n.AsFunctionDeclaration()
	case KindInterfaceDeclaration:
		return n.AsInterfaceDeclaration()
	case KindTypeAliasDeclaration:
		return n.AsTypeAliasDeclaration()
	case KindEnumDeclaration:
		return n.AsEnumDeclaration()
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

// =============================================================================
// 透传API实现 - 透明访问analyzer/parser解析数据
// =============================================================================

// GetParserData 获取节点对应的analyzer/parser解析数据
// 这是透传API的核心方法，提供对底层解析结构的直接访问
// 返回值: 解析结果接口和是否成功找到对应的解析数据
func (node Node) GetParserData() (interface{}, bool) {
	// 安全检查：确保节点和源文件有效
	if node.sourceFile == nil || node.sourceFile.nodeResultMap == nil {
		return nil, false
	}

	// 从节点-结果映射中查找对应的解析数据
	// 这个映射在sourcefile.go的buildNodeResultMap()方法中构建
	if result, ok := node.sourceFile.nodeResultMap[node.Node]; ok {
		return result, true
	}

	// 未找到对应的解析数据
	return nil, false
}

// GetParserDataWithFallback 带降级策略的解析数据获取
// 当无法从缓存获取时，提供实时解析或基础信息降级
func (node Node) GetParserDataWithFallback() (interface{}, bool, error) {
	// 策略1: 从缓存获取
	if data, ok := node.GetParserData(); ok {
		return data, true, nil
	}

	// 策略2: 实时解析 (如果可能)
	if data, err := parseNodeOnDemand(node); err == nil {
		return data, false, nil
	}

	// 策略3: 返回基础AST信息作为降级
	return node.getBasicInfo(), false, fmt.Errorf("no parser data available, using fallback")
}

// getBasicInfo 获取节点的基础信息作为降级策略
func (node Node) getBasicInfo() map[string]interface{} {
	info := make(map[string]interface{})

	// 基础节点信息
	info["kind"] = node.GetKind().String()
	info["text"] = node.GetText()
	info["start"] = node.GetStart()
	info["end"] = node.GetEnd()
	info["line"] = node.GetStartLineNumber()
	info["column"] = node.GetStartColumnNumber()

	// 如果有源文件，添加文件信息
	if node.sourceFile != nil {
		info["filePath"] = node.sourceFile.GetFilePath()
	}

	return info
}

// TryGetParserData 带错误处理的类型安全获取
// 这是一个泛型辅助方法，提供更友好的错误信息
func TryGetParserData[T any](node Node) (T, error) {
	data, ok := GetParserData[T](node)
	if !ok {
		var zero T
		return zero, fmt.Errorf("node is not of expected type %T, actual type: %T", zero, getNodeActualType(node))
	}
	return data, nil
}

// getNodeActualType 获取节点的实际类型（用于错误信息）
func getNodeActualType(node Node) interface{} {
	data, ok := node.GetParserData()
	if ok {
		return fmt.Sprintf("%T", data)
	}
	return fmt.Sprintf("ast.Node kind: %s", node.GetKind().String())
}

// =============================================================================
// 常用类型的便利方法 - 基于透传API的快捷访问
// =============================================================================

// AsCallExpression 获取函数调用表达式的解析数据
// 返回: parser.CallExpression结构和是否成功
func (node Node) AsCallExpression() (parser.CallExpression, bool) {
	return GetParserData[parser.CallExpression](node)
}

// AsVariableDeclaration 获取变量声明的解析数据
// 返回: parser.VariableDeclaration结构和是否成功
func (node Node) AsVariableDeclaration() (parser.VariableDeclaration, bool) {
	return GetParserData[parser.VariableDeclaration](node)
}

// AsInterfaceDeclaration 获取接口声明的解析数据
// 返回: parser.InterfaceDeclarationResult结构和是否成功
func (node Node) AsInterfaceDeclaration() (parser.InterfaceDeclarationResult, bool) {
	return GetParserData[parser.InterfaceDeclarationResult](node)
}

// AsFunctionDeclaration 获取函数声明的解析数据
// 返回: parser.FunctionDeclarationResult结构和是否成功
func (node Node) AsFunctionDeclaration() (parser.FunctionDeclarationResult, bool) {
	return GetParserData[parser.FunctionDeclarationResult](node)
}

// AsImportDeclaration 获取导入声明的解析数据
// 返回: projectParser.ImportDeclarationResult结构和是否成功
func (node Node) AsImportDeclaration() (projectParser.ImportDeclarationResult, bool) {
	return GetParserData[projectParser.ImportDeclarationResult](node)
}

// AsTypeAliasDeclaration 获取类型别名声明的解析数据
// 返回: parser.TypeDeclarationResult结构和是否成功
func (node Node) AsTypeAliasDeclaration() (parser.TypeDeclarationResult, bool) {
	return GetParserData[parser.TypeDeclarationResult](node)
}

// AsEnumDeclaration 获取枚举声明的解析数据
// 返回: parser.EnumDeclarationResult结构和是否成功
func (node Node) AsEnumDeclaration() (parser.EnumDeclarationResult, bool) {
	return GetParserData[parser.EnumDeclarationResult](node)
}

// AsExportDeclaration 获取导出声明的解析数据
// 返回: projectParser.ExportDeclarationResult结构和是否成功
func (node Node) AsExportDeclaration() (projectParser.ExportDeclarationResult, bool) {
	return GetParserData[projectParser.ExportDeclarationResult](node)
}

// AsPropertyAssignment 获取属性赋值的解析数据
// 这个通常在对象字面量中使用
func (node Node) AsPropertyAssignment() (interface{}, bool) {
	// PropertyAssignment可能在不同的解析结构中
	// 我们先尝试从通用的ExtractedNodes中获取
	if extractedNodes, ok := node.GetParserData(); ok {
		// 检查是否是ExtractedNodes.AnyDeclarations类型
		if nodes, ok := extractedNodes.([]interface{}); ok {
			for _, item := range nodes {
				// 这里可以根据实际的数据结构进行更精确的匹配
				if item != nil {
					return item, true
				}
			}
		}
	}
	return nil, false
}

// HasParserData 检查节点是否有对应的解析数据
// 这是一个便捷的检查方法，避免直接使用GetParserData()的第二个返回值
func (node Node) HasParserData() bool {
	_, ok := node.GetParserData()
	return ok
}

// GetParserDataType 获取解析数据的类型名称
// 用于调试和日志记录
func (node Node) GetParserDataType() string {
	data, ok := node.GetParserData()
	if !ok {
		return "none"
	}
	return fmt.Sprintf("%T", data)
}

// =============================================================================
// 泛型辅助函数 - 透传API的类型安全支持
// =============================================================================

// GetParserData 类型安全的获取解析数据
// 利用Go 1.18+的泛型特性，在编译时提供类型安全检查
// 这是透传API的核心泛型函数，为所有类型安全访问提供基础
func GetParserData[T any](node Node) (T, bool) {
	var zero T // 声明泛型的零值

	data, ok := node.GetParserData()
	if !ok {
		return zero, false
	}

	// 类型断言：确保返回的是预期的类型
	// 如果类型不匹配，返回零值和false
	if typed, ok := data.(T); ok {
		return typed, true
	}

	return zero, false
}

// parseNodeOnDemand 按需解析节点
// 当缓存中没有解析数据时，提供实时解析功能
// 这是一个降级策略，确保即使缓存失效也能获得基础解析结果
func parseNodeOnDemand(node Node) (interface{}, error) {
	// 检查节点类型，尝试使用analyzer/parser的实时解析功能
	switch node.GetKind() {
	case KindCallExpression:
		// 尝试实时解析函数调用表达式
		if node.IsCallExpr() {
			// 这里需要调用analyzer/parser的实时解析函数
			// 暂时返回基础信息，后续可以集成完整的解析逻辑
			return map[string]interface{}{
				"type":       "CallExpression",
				"expression": node.GetText(),
				"runtime":    true,
			}, nil
		}

	case KindVariableDeclaration:
		// 尝试实时解析变量声明
		if node.IsVariableDeclaration() {
			return map[string]interface{}{
				"type":     "VariableDeclaration",
				"variable": node.GetText(),
				"runtime":  true,
			}, nil
		}

	default:
		// 对于不支持的节点类型，返回基础信息
		return node.getBasicInfo(), nil
	}

	return nil, fmt.Errorf("unsupported node type for on-demand parsing: %s", node.GetKind().String())
}

// =============================================================================
// 透传API使用示例和最佳实践函数
// =============================================================================

// DebugParserData 调试输出解析数据
// 用于开发阶段查看节点对应的解析数据内容
func (node Node) DebugParserData() {
	fmt.Printf("=== 节点解析数据调试 ===\n")
	fmt.Printf("节点类型: %s\n", node.GetKind().String())
	fmt.Printf("节点文本: %s\n", node.GetText())
	fmt.Printf("位置: %d:%d\n", node.GetStartLineNumber(), node.GetStartColumnNumber())

	if data, ok := node.GetParserData(); ok {
		fmt.Printf("解析数据类型: %T\n", data)
		fmt.Printf("解析数据内容: %+v\n", data)
	} else {
		fmt.Printf("解析数据: 无\n")
	}
	fmt.Printf("========================\n")
}

// ForEachWithParserData 遍历节点并检查是否有解析数据
// 这是一个便利方法，结合了节点遍历和解析数据检查
func (node Node) ForEachWithParserData(callback func(Node, interface{})) {
	// 检查当前节点是否有解析数据
	if data, ok := node.GetParserData(); ok {
		callback(node, data)
	}

	// 递归遍历子节点
	node.ForEachChild(func(child *ast.Node) bool {
		childNode := Node{
			Node:       child,
			sourceFile: node.sourceFile,
		}
		childNode.ForEachWithParserData(callback)
		return false
	})
}

// CountNodesWithParserData 统计有解析数据的子节点数量
// 用于性能分析和调试
func (node Node) CountNodesWithParserData() int {
	count := 0

	if node.HasParserData() {
		count++
	}

	node.ForEachChild(func(child *ast.Node) bool {
		childNode := Node{
			Node:       child,
			sourceFile: node.sourceFile,
		}
		count += childNode.CountNodesWithParserData()
		return false
	})

	return count
}
