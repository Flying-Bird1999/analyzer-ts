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
	return strings.TrimLeft(n.sourceFile.fileResult.Raw[n.Pos():n.End()], " ")
}

// GetStartLineNumber 获取起始行号（1-based）
func (n *Node) GetStartLineNumber() int {
	if !n.IsValid() {
		return 0
	}
	line, _ := n.getLineAndColumn()
	return line
}

// GetStartLineCharacter 获取起始列号（0-based）
func (n *Node) GetStartLineCharacter() int {
	if !n.IsValid() {
		return 0
	}
	_, col := n.getLineAndColumn()
	return col
}

// GetStartColumnNumber 获取起始列号（1-based）
func (n *Node) GetStartColumnNumber() int {
	if !n.IsValid() {
		return 0
	}
	_, col := n.getLineAndColumn()
	return col + 1
}

// GetStart 获取起始位置
func (n *Node) GetStart() int {
	if !n.IsValid() {
		return 0
	}
	return n.Pos()
}

// GetStartLinePos 获取行起始位置
func (n *Node) GetStartLinePos() int {
	if !n.IsValid() {
		return 0
	}
	_, col := n.getLineAndColumn()
	return n.GetStart() - col
}

// GetEnd 获取结束位置
func (n *Node) GetEnd() int {
	if !n.IsValid() {
		return 0
	}
	return n.End()
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

// getLineAndColumn 统一的位置计算方法，避免重复计算
func (n *Node) getLineAndColumn() (line, column int) {
	if !n.IsValid() || n.sourceFile == nil || n.sourceFile.fileResult == nil {
		return 0, 0
	}

	pos := n.Pos()
	text := n.sourceFile.fileResult.Raw[:pos]
	lines := strings.Split(text, "\n")

	line = len(lines)
	if len(lines) == 0 {
		column = 0
	} else {
		column = len(lines[len(lines)-1])
	}

	return line, column
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

// GetChildren 获取所有直接子节点
func (n *Node) GetChildren() []*Node {
	if !n.IsValid() {
		return nil
	}

	var children []*Node
	n.Node.ForEachChild(func(child *ast.Node) bool {
		children = append(children, &Node{
			Node:       child,
			sourceFile: n.sourceFile,
		})
		return false // 继续遍历其他子节点
	})

	return children
}

// GetFirstChild 根据条件获取第一个匹配的子节点
func (n *Node) GetFirstChild(predicate func(Node) bool) *Node {
	if !n.IsValid() {
		return nil
	}

	var foundChild *Node
	n.Node.ForEachChild(func(child *ast.Node) bool {
		childNode := &Node{
			Node:       child,
			sourceFile: n.sourceFile,
		}
		if predicate != nil && predicate(*childNode) {
			foundChild = childNode
			return true // 找到了，停止遍历
		}
		return false // 继续遍历其他子节点
	})

	return foundChild
}

// ForEachChild 遍历所有直接子节点
func (n *Node) ForEachChild(callback func(Node) bool) {
	if !n.IsValid() || callback == nil {
		return
	}

	n.Node.ForEachChild(func(child *ast.Node) bool {
		childNode := Node{
			Node:       child,
			sourceFile: n.sourceFile,
		}
		return callback(childNode)
	})
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
// 统一的类型检查API
// =============================================================================

// GetKind 返回节点的类型（统一接口）
func (n Node) GetKind() SyntaxKind {
	return SyntaxKind(n.Kind)
}

// IsKind 检查节点是否为指定类型（基础类型检查）
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

// =============================================================================
// 位置信息 API - 补充缺失的位置方法
// =============================================================================

// GetEndLineNumber 获取结束行号（1-based）
func (n *Node) GetEndLineNumber() int {
	if !n.IsValid() {
		return 0
	}
	line, _ := n.getEndLineAndColumn()
	return line
}

// GetEndColumnNumber 获取结束列号（1-based）
func (n *Node) GetEndColumnNumber() int {
	if !n.IsValid() {
		return 0
	}
	_, col := n.getEndLineAndColumn()
	return col + 1
}

// GetWidth 获取节点文本宽度
func (n *Node) GetWidth() int {
	if !n.IsValid() {
		return 0
	}
	return n.GetEnd() - n.GetStart()
}

// GetKindName 获取语法类型的字符串名称
func (n *Node) GetKindName() string {
	if !n.IsValid() {
		return ""
	}
	return n.Kind.String()
}

// getEndLineAndColumn 获取结束位置的行号和列号
func (n *Node) getEndLineAndColumn() (line, column int) {
	if !n.IsValid() || n.sourceFile == nil || n.sourceFile.fileResult == nil {
		return 0, 0
	}

	pos := n.End() - 1 // 结束位置是后开区间，所以-1
	text := n.sourceFile.fileResult.Raw[:pos]
	lines := strings.Split(text, "\n")

	line = len(lines)
	if len(lines) == 0 {
		column = 0
	} else {
		column = len(lines[len(lines)-1])
	}

	return line, column
}

// =============================================================================
// Node 类型检查方法 - 对应文档中的 namespace Node
// =============================================================================

// 基础类型检查
func (n Node) IsIdentifier() bool { return n.IsKind(KindIdentifier) }

// 声明类型
func (n Node) IsFunctionDeclaration() bool  { return n.IsKind(KindFunctionDeclaration) }
func (n Node) IsVariableDeclaration() bool  { return n.IsKind(KindVariableDeclaration) }
func (n Node) IsInterfaceDeclaration() bool { return n.IsKind(KindInterfaceDeclaration) }
func (n Node) IsClassDeclaration() bool     { return n.IsKind(KindClassDeclaration) }
func (n Node) IsEnumDeclaration() bool      { return n.IsKind(KindEnumDeclaration) }
func (n Node) IsTypeAliasDeclaration() bool { return n.IsKind(KindTypeAliasDeclaration) }

// 表达式类型
func (n Node) IsCallExpression() bool           { return n.IsKind(KindCallExpression) }
func (n Node) IsPropertyAccessExpression() bool { return n.IsKind(KindPropertyAccessExpression) }
func (n Node) IsBinaryExpression() bool         { return n.IsKind(KindBinaryExpression) }
func (n Node) IsObjectLiteralExpression() bool  { return n.IsKind(KindObjectLiteralExpression) }
func (n Node) IsArrayLiteralExpression() bool   { return n.IsKind(KindArrayLiteralExpression) }

// 其他类型
func (n Node) IsPropertyAssignment() bool { return n.IsKind(KindPropertyAssignment) }
func (n Node) IsImportSpecifier() bool    { return n.IsKind(KindImportSpecifier) }
func (n Node) IsImportDeclaration() bool  { return n.IsKind(KindImportDeclaration) }
func (n Node) IsExportDeclaration() bool  { return n.IsKind(KindExportDeclaration) }

// =============================================================================
// 全局辅助函数 - 来自原 node.go
// =============================================================================

// GetSymbol 获取节点对应的符号
func (n Node) GetSymbol() (*Symbol, error) {
	return GetSymbol(n)
}

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

	// 策略2: 返回基础AST信息作为降级
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

// AsInterfaceDeclaration 获取接口声明的解析数据
// 返回: parser.InterfaceDeclarationResult结构和是否成功
func (node Node) AsInterfaceDeclaration() (parser.InterfaceDeclarationResult, bool) {
	return GetParserData[parser.InterfaceDeclarationResult](node)
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
// 特定节点类型的专有 API - 类型安全的高级接口
// =============================================================================

// =============================================================================
// 辅助方法 - 支持特定节点类型 API 的内部实现
// =============================================================================

// getTextFromChild 从指定类型的子节点中提取文本
func (n Node) getTextFromChild(kind SyntaxKind) string {
	child := n.getFirstChildByKind(kind)
	if child != nil {
		return child.GetText()
	}
	return ""
}

// isOperatorKind 判断语法类型是否为操作符
func isOperatorKind(kind SyntaxKind) bool {
	switch kind {
	case KindEqualsToken, KindPlusToken, KindMinusToken,
		KindAsteriskToken, KindSlashToken, KindEqualsEqualsEqualsToken,
		KindExclamationEqualsEqualsToken:
		return true
	default:
		return false
	}
}

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

// =============================================================================
// 特定节点类型的专有 API - 类型安全的高级接口
// =============================================================================

// NodeWrapper 基础接口，所有特定节点类型都应该实现这个接口
type NodeWrapper interface {
	GetNode() *Node
	GetKind() SyntaxKind
}

// =============================================================================
// VariableDeclaration 特定API
// =============================================================================

// VariableDeclaration 提供变量声明节点的专有API
type VariableDeclaration struct {
	*Node
}

// GetNode 返回基础Node节点
func (v *VariableDeclaration) GetNode() *Node {
	return v.Node
}

// GetKind 返回节点类型
func (v *VariableDeclaration) GetKind() SyntaxKind {
	return KindVariableDeclaration
}

// AsVariableDeclaration 将Node转换为VariableDeclaration，提供类型安全的转换
func (n *Node) AsVariableDeclaration() (*VariableDeclaration, bool) {
	if !n.IsVariableDeclaration() {
		return nil, false
	}
	return &VariableDeclaration{Node: n}, true
}

// GetNameNode 获取变量名节点
func (v *VariableDeclaration) GetNameNode() *Node {
	if v.Node == nil || !v.Node.IsValid() {
		return nil
	}

	// 查找第一个标识符子节点
	return v.Node.getFirstChildByKind(KindIdentifier)
}

// GetName 获取变量名
func (v *VariableDeclaration) GetName() string {
	nameNode := v.GetNameNode()
	if nameNode == nil {
		return ""
	}
	return strings.TrimSpace(nameNode.GetText())
}

// GetInitializer 获取初始值表达式节点
func (v *VariableDeclaration) GetInitializer() *Node {
	if !v.Node.IsValid() {
		return nil
	}

	// 对于变量声明 const x = 1，初始值通常是 VariableDeclaration 的第二个子节点
	// 测试显示子节点是: [Identifier("x"), NumericLiteral("1")]
	children := v.Node.GetChildren()
	if len(children) >= 2 {
		return children[1] // 第二个子节点通常就是初始值
	}

	// 如果只有一个子节点，可能没有初始值
	return nil
}

// HasInitializer 检查是否有初始值
func (v *VariableDeclaration) HasInitializer() bool {
	return v.GetInitializer() != nil
}

// GetParserData 获取透传API的解析数据
func (v *VariableDeclaration) GetParserData() (parser.VariableDeclaration, bool) {
	return GetParserData[parser.VariableDeclaration](*v.Node)
}

// =============================================================================
// CallExpression 特定API
// =============================================================================

// CallExpression 提供函数调用节点的专有API
type CallExpression struct {
	*Node
}

// GetNode 返回基础Node节点
func (c *CallExpression) GetNode() *Node {
	return c.Node
}

// GetKind 返回节点类型
func (c *CallExpression) GetKind() SyntaxKind {
	return KindCallExpression
}

// AsCallExpression 将Node转换为CallExpression，提供类型安全的转换
func (n *Node) AsCallExpression() (*CallExpression, bool) {
	if !n.IsCallExpression() {
		return nil, false
	}
	return &CallExpression{Node: n}, true
}

// GetExpression 获取被调用的表达式（函数名或函数对象）
func (c *CallExpression) GetExpression() *Node {
	if !c.Node.IsValid() {
		return nil
	}

	// 对于函数调用 foo()，第一个子节点通常是被调用的表达式
	children := c.Node.GetChildren()
	if len(children) > 0 {
		return children[0]
	}
	return nil
}

// GetArguments 获取参数列表
func (c *CallExpression) GetArguments() []*Node {
	if !c.Node.IsValid() {
		return nil
	}

	var arguments []*Node
	children := c.Node.GetChildren()
	if len(children) > 1 {
		// 跳过第一个子节点（表达式），其余的是参数
		for i := 1; i < len(children); i++ {
			arguments = append(arguments, children[i])
		}
	}
	return arguments
}

// GetArgumentCount 获取参数数量
func (c *CallExpression) GetArgumentCount() int {
	return len(c.GetArguments())
}

// GetArgument 获取指定索引的参数
func (c *CallExpression) GetArgument(index int) *Node {
	args := c.GetArguments()
	if index >= 0 && index < len(args) {
		return args[index]
	}
	return nil
}

// IsMethodCall 检查是否为方法调用（是否为 obj.method() 形式）
func (c *CallExpression) IsMethodCall() bool {
	expr := c.GetExpression()
	return expr != nil && expr.IsPropertyAccessExpression()
}

// IsConstructorCall 检查是否为构造函数调用（new 调用）
func (c *CallExpression) IsConstructorCall() bool {
	// 简化版本：检查调用者的文本中是否包含 "new"
	text := c.GetText()
	return strings.Contains(text, "new ")
}

// GetCalleeName 获取被调用函数的名称
func (c *CallExpression) GetCalleeName() string {
	expr := c.GetExpression()
	if expr == nil {
		return ""
	}

	if expr.IsIdentifier() {
		return expr.GetText()
	}

	if expr.IsPropertyAccessExpression() {
		// 对于 obj.method()，返回 method
		if propAccess, ok := expr.AsPropertyAccessExpression(); ok {
			return propAccess.GetName()
		}
	}

	return ""
}

// GetParserData 获取透传API的解析数据
func (c *CallExpression) GetParserData() (parser.CallExpression, bool) {
	return GetParserData[parser.CallExpression](*c.Node)
}

// =============================================================================
// PropertyAccessExpression 特定API
// =============================================================================

// PropertyAccessExpression 提供属性访问节点的专有API
type PropertyAccessExpression struct {
	*Node
}

// GetNode 返回基础Node节点
func (p *PropertyAccessExpression) GetNode() *Node {
	return p.Node
}

// GetKind 返回节点类型
func (p *PropertyAccessExpression) GetKind() SyntaxKind {
	return KindPropertyAccessExpression
}

// AsPropertyAccessExpression 将Node转换为PropertyAccessExpression，提供类型安全的转换
func (n *Node) AsPropertyAccessExpression() (*PropertyAccessExpression, bool) {
	if !n.IsPropertyAccessExpression() {
		return nil, false
	}
	return &PropertyAccessExpression{Node: n}, true
}

// GetName 获取属性名
func (p *PropertyAccessExpression) GetName() string {
	if !p.Node.IsValid() {
		return ""
	}

	// 查找标识符子节点作为属性名
	children := p.Node.GetChildren()
	for i := len(children) - 1; i >= 0; i-- {
		if children[i].IsIdentifier() {
			return strings.TrimSpace(children[i].GetText())
		}
	}
	return ""
}

// GetExpression 获取被访问的对象表达式
func (p *PropertyAccessExpression) GetExpression() *Node {
	if !p.Node.IsValid() {
		return nil
	}

	// 对于 obj.key，获取 obj 部分
	children := p.Node.GetChildren()
	if len(children) >= 2 {
		return children[0] // 第一个子节点通常是被访问的对象
	}
	return nil
}

// IsOptionalAccess 检查是否为可选链访问（obj?.prop）
func (p *PropertyAccessExpression) IsOptionalAccess() bool {
	// 检查是否包含问号token（简化版本）
	text := p.GetText()
	return strings.Contains(text, "?.")
}

// IsElementAccess 检查是否为元素访问（obj[prop]）
func (p *PropertyAccessExpression) IsElementAccess() bool {
	// 简化版本：检查文本是否包含方括号
	text := p.GetText()
	return strings.Contains(text, "[") && strings.Contains(text, "]")
}

// GetObjectExpression 获取对象表达式（与GetExpression相同，语义更明确）
func (p *PropertyAccessExpression) GetObjectExpression() *Node {
	return p.GetExpression()
}

// GetPropertyNameNode 获取属性名节点
func (p *PropertyAccessExpression) GetPropertyNameNode() *Node {
	// 查找属性名节点，通常是最后一个标识符子节点
	children := p.GetChildren()
	for i := len(children) - 1; i >= 0; i-- {
		if children[i].IsIdentifier() {
			return children[i]
		}
	}
	return nil
}

// GetParserData 获取透传API的解析数据（PropertyAssignment在透传API中没有直接对应）
func (p *PropertyAccessExpression) GetParserData() (interface{}, bool) {
	return p.Node.GetParserData()
}

// =============================================================================
// FunctionDeclaration 特定API
// =============================================================================

// FunctionDeclaration 提供函数声明节点的专有API
type FunctionDeclaration struct {
	*Node
}

// GetNode 返回基础Node节点
func (f *FunctionDeclaration) GetNode() *Node {
	return f.Node
}

// GetKind 返回节点类型
func (f *FunctionDeclaration) GetKind() SyntaxKind {
	return KindFunctionDeclaration
}

// AsFunctionDeclaration 将Node转换为FunctionDeclaration，提供类型安全的转换
func (n *Node) AsFunctionDeclaration() (*FunctionDeclaration, bool) {
	if !n.IsFunctionDeclaration() {
		return nil, false
	}
	return &FunctionDeclaration{Node: n}, true
}

// GetNameNode 获取函数名节点（根据 ts-morph.md 场景 7.4）
func (f *FunctionDeclaration) GetNameNode() *Node {
	if !f.Node.IsValid() {
		return nil
	}

	// 函数名通常是第一个标识符子节点
	return f.Node.getFirstChildByKind(KindIdentifier)
}

// GetName 获取函数名（便利方法）
func (f *FunctionDeclaration) GetName() string {
	nameNode := f.GetNameNode()
	if nameNode != nil {
		return strings.TrimSpace(nameNode.GetText())
	}
	return ""
}

// IsAnonymous 检查是否为匿名函数
func (f *FunctionDeclaration) IsAnonymous() bool {
	return f.GetName() == ""
}

// GetParserData 获取透传API的解析数据
func (f *FunctionDeclaration) GetParserData() (parser.FunctionDeclarationResult, bool) {
	return GetParserData[parser.FunctionDeclarationResult](*f.Node)
}

// =============================================================================
// BinaryExpression 特定API
// =============================================================================

// BinaryExpression 提供二元表达式节点的专有API
type BinaryExpression struct {
	*Node
}

// GetNode 返回基础Node节点
func (b *BinaryExpression) GetNode() *Node {
	return b.Node
}

// GetKind 返回节点类型
func (b *BinaryExpression) GetKind() SyntaxKind {
	return KindBinaryExpression
}

// AsBinaryExpression 将Node转换为BinaryExpression，提供类型安全的转换
func (n *Node) AsBinaryExpression() (*BinaryExpression, bool) {
	if !n.IsBinaryExpression() {
		return nil, false
	}
	return &BinaryExpression{Node: n}, true
}

// GetLeft 获取左操作数
func (b *BinaryExpression) GetLeft() *Node {
	if !b.Node.IsValid() {
		return nil
	}
	children := b.Node.GetChildren()
	if len(children) > 0 {
		return children[0]
	}
	return nil
}

// GetRight 获取右操作数
func (b *BinaryExpression) GetRight() *Node {
	if !b.Node.IsValid() {
		return nil
	}
	children := b.Node.GetChildren()
	if len(children) >= 2 {
		return children[1]
	}
	return nil
}

// GetOperatorToken 获取操作符节点
func (b *BinaryExpression) GetOperatorToken() *Node {
	if !b.Node.IsValid() {
		return nil
	}
	children := b.Node.GetChildren()
	if len(children) > 1 {
		// 寻找操作符
		for _, child := range children {
			if isOperatorKind(child.GetKind()) {
				return child
			}
		}
	}
	// for some binary expressions, the operator is not a child
	if len(children) > 1 {
		return children[1]
	}
	return nil
}

// =============================================================================
// 辅助方法
// =============================================================================

// getFirstChildByKind 根据语法类型获取第一个匹配的子节点
// 注意：这个方法已经移到这里，因为它主要是特定节点类型API的辅助方法
func (n Node) getFirstChildByKind(kind SyntaxKind) *Node {
	if !n.IsValid() {
		return nil
	}

	children := n.GetChildren()
	for _, child := range children {
		if child.IsKind(kind) {
			return child
		}
	}
	return nil
}

// =============================================================================
// ImportSpecifier 特定API - ts-morph兼容性支持
// =============================================================================

// ImportSpecifier 提供导入说明符节点的专有API
// 用于处理 import { foo as bar } from 'module' 中的具体导入项
type ImportSpecifier struct {
	*Node
}

// GetNode 返回基础Node节点
func (i *ImportSpecifier) GetNode() *Node {
	return i.Node
}

// GetKind 返回节点类型
func (i *ImportSpecifier) GetKind() SyntaxKind {
	return KindImportSpecifier
}

// GetParserData 获取透传API的解析数据
// 返回底层的parser.ImportModule数据结构
func (i *ImportSpecifier) GetParserData() (parser.ImportModule, bool) {
	return GetParserData[parser.ImportModule](*i.Node)
}

// AsImportSpecifier 将Node转换为ImportSpecifier，提供类型安全的转换
// API兼容性：对应ts-morph的import specifier节点操作
func (n *Node) AsImportSpecifier() (*ImportSpecifier, bool) {
	if !n.IsImportSpecifier() {
		return nil, false
	}
	return &ImportSpecifier{Node: n}, true
}

// GetAliasNode 获取导入别名节点
// API兼容性：对应ts-morph的ImportSpecifier.getAliasNode()方法
//
// 使用场景：
//   - import { foo } from 'module'      -> GetAliasNode() 返回 nil (无别名)
//   - import { foo as bar } from 'module' -> GetAliasNode() 返回 "bar" 标识符节点
//
// 返回值：
//   - *Node: 别名标识符节点，如果存在别名的话
//   - nil: 没有别名时返回 nil
func (i *ImportSpecifier) GetAliasNode() *Node {
	// 安全检查：防止空指针
	if i == nil || i.Node == nil || !i.Node.IsValid() {
		return nil
	}

	// 将Node转换为typescript-go的ImportSpecifier
	importSpecAST, ok := i.Node.AsImportSpecifier()
	if !ok {
		return nil
	}

	// 检查是否有别名：如果PropertyName()返回的节点不为nil，说明有别名
	// 语法：import { original as alias }
	// PropertyName = "original", Name = "alias"
	if propertyNameNode := importSpecAST.PropertyName(); propertyNameNode != nil {
		// 有别名：返回代表别名的Name节点
		nameNode := importSpecAST.Name()
		if nameNode != nil {
			return &Node{
				Node:       nameNode.AsNode(),
				sourceFile: i.Node.sourceFile,
			}
		}
	}

	// 没有别名
	return nil
}

// GetOriginalName 获取原始导入名称（无别名时的名称）
// API兼容性：辅助方法，用于获取原始导入的模块名
//
// 使用场景：
//   - import { foo } from 'module'      -> GetOriginalName() 返回 "foo"
//   - import { foo as bar } from 'module' -> GetOriginalName() 返回 "foo"
func (i *ImportSpecifier) GetOriginalName() string {
	if !i.Node.IsValid() {
		return ""
	}

	// 将Node转换为typescript-go的ImportSpecifier
	importSpecAST, ok := i.Node.AsImportSpecifier()
	if !ok {
		return i.Node.GetText() // 回退到节点文本
	}

	// 如果有别名，PropertyName是原始名称
	if propertyNameNode := importSpecAST.PropertyName(); propertyNameNode != nil {
		return propertyNameNode.Text()
	}

	// 没有别名，Name就是原始名称
	if importSpecAST.Name() != nil {
		return importSpecAST.Name().Text()
	}

	return i.Node.GetText() // 回退到节点文本
}

// GetLocalName 获取本地使用名称（可能和原始名称相同）
// API兼容性：辅助方法，用于获取在当前文件中实际使用的名称
//
// 使用场景：
//   - import { foo } from 'module'      -> GetLocalName() 返回 "foo"
//   - import { foo as bar } from 'module' -> GetLocalName() 返回 "bar"
func (i *ImportSpecifier) GetLocalName() string {
	if !i.Node.IsValid() {
		return ""
	}

	// 将Node转换为typescript-go的ImportSpecifier
	importSpecAST, ok := i.Node.AsImportSpecifier()
	if !ok {
		return i.Node.GetText() // 回退到节点文本
	}

	// 无论是否有别名，Name都是本地使用的名称
	if importSpecAST.Name() != nil {
		return importSpecAST.Name().Text()
	}

	return i.Node.GetText() // 回退到节点文本
}

// HasAlias 判断是否有别名
// API兼容性：便捷方法，用于判断导入项是否有别名
func (i *ImportSpecifier) HasAlias() bool {
	if !i.Node.IsValid() {
		return false
	}

	// 将Node转换为typescript-go的ImportSpecifier
	importSpecAST, ok := i.Node.AsImportSpecifier()
	if !ok {
		return false
	}

	// 直接检查PropertyName()返回的节点是否存在，这是最可靠的方式
	propertyNameNode := importSpecAST.PropertyName()
	return propertyNameNode != nil
}

// =============================================================================
// ImportSpecifier 辅助方法
// =============================================================================


// getChildrenOfNode 获取AST节点的所有直接子节点
func getChildrenOfNode(node *ast.Node) []*ast.Node {
	if node == nil {
		return nil
	}

	var children []*ast.Node
	node.ForEachChild(func(child *ast.Node) bool {
		children = append(children, child)
		return false // 继续遍历所有子节点
	})
	return children
}

// =============================================================================
// 引用查找 API - 调用层 (委托给 references.go 中的核心实现)
// =============================================================================

// FindReferences 查找给定节点所代表的符号的所有引用
func (n *Node) FindReferences() ([]*Node, error) {
	return findReferencesCore(*n)
}

// FindReferencesWithCache 带缓存的引用查找，支持错误处理和重试机制
func (n *Node) FindReferencesWithCache() ([]*Node, bool, error) {
	return n.FindReferencesWithCacheAndRetry(DefaultRetryConfig())
}

// FindReferencesWithCacheAndRetry 带缓存和重试机制的引用查找
func (n *Node) FindReferencesWithCacheAndRetry(retryConfig *RetryConfig) ([]*Node, bool, error) {
	return findReferencesWithCacheAndRetry(*n, retryConfig)
}

// GotoDefinition 查找给定节点所代表的符号的定义位置
func (n *Node) GotoDefinition() ([]*Node, error) {
	return gotoDefinitionCore(*n)
}

// FindAllReferences 查找所有引用，包括定义位置
func (n *Node) FindAllReferences() ([]*Node, error) {
	var allReferences []*Node

	// 1. 查找引用
	refs, err := n.FindReferences()
	if err != nil {
		return nil, err
	}
	allReferences = append(allReferences, refs...)

	// 2. 查找定义
	defs, err := n.GotoDefinition()
	if err != nil {
		// 定义查找失败不影响引用查找结果
		return allReferences, nil
	}
	allReferences = append(allReferences, defs...)

	return allReferences, nil
}

// CountReferences 统计引用数量
func (n *Node) CountReferences() (int, error) {
	refs, err := n.FindReferences()
	if err != nil {
		return 0, err
	}
	return len(refs), nil
}
