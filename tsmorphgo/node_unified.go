package tsmorphgo


// =============================================================================
// 统一的 Node API 设计 - 替换分散的 IsXXX 和 AsXXX 函数
// 提供更简洁、一致的接口
// =============================================================================

// NodeMethods 为 Node 类型添加统一的检查方法
type NodeMethods struct {
	node Node
}

// Methods 返回节点的统一方法接口
func (n Node) Methods() *NodeMethods {
	return &NodeMethods{node: n}
}

// Is 检查节点是否为指定类型
// 替代所有 IsXXX 函数，提供统一的接口
func (m *NodeMethods) Is(kind SyntaxKind) bool {
	return m.node.Kind == kind
}

// IsAny 检查节点是否为任意一个指定类型
func (m *NodeMethods) IsAny(kinds ...SyntaxKind) bool {
	for _, kind := range kinds {
		if m.node.Kind == kind {
			return true
		}
	}
	return false
}

// IsCategory 检查节点是否属于指定类别
func (m *NodeMethods) IsCategory(category NodeCategory) bool {
	return category.Contains(m.node.Kind)
}

// As 尝试将节点转换为指定类型的结果
// 替代所有 AsXXX 函数，提供统一的转换接口
func (m *NodeMethods) As(kind SyntaxKind) (interface{}, bool) {
	if !m.Is(kind) {
		return nil, false
	}

	// 根据类型调用对应的转换函数
	switch kind {
	case KindImportDeclaration:
		return AsImportDeclaration(m.node)
	case KindVariableDeclaration:
		return AsVariableDeclaration(m.node)
	case KindFunctionDeclaration:
		return AsFunctionDeclaration(m.node)
	case KindInterfaceDeclaration:
		return AsInterfaceDeclaration(m.node)
	case KindTypeAliasDeclaration:
		return AsTypeAliasDeclaration(m.node)
	case KindEnumDeclaration:
		return AsEnumDeclaration(m.node)
	case KindClassDeclaration:
		return AsClassDeclaration(m.node)
	default:
		return m.node, true
	}
}

// TryAs 尝试转换，如果失败则返回指定的默认值
func (m *NodeMethods) TryAs(kind SyntaxKind, defaultValue interface{}) interface{} {
	if result, ok := m.As(kind); ok {
		return result
	}
	return defaultValue
}

// =============================================================================
// 节点类别定义
// =============================================================================

// NodeCategory 定义节点类别，便于批量判断
type NodeCategory struct {
	name  string
	kinds map[SyntaxKind]bool
}

// 常用节点类别
var (
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

	CategoryLiterals = NodeCategory{
		name: "literals",
		kinds: map[SyntaxKind]bool{
			KindStringLiteral:  true,
			KindNumericLiteral: true,
			KindTrueKeyword:    true,
			KindFalseKeyword:   true,
			KindNullKeyword:    true,
			KindUndefinedKeyword: true,
		},
	}

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

	CategoryIdentifiers = NodeCategory{
		name: "identifiers",
		kinds: map[SyntaxKind]bool{
			KindIdentifier: true,
		},
	}
)

// Contains 检查类别是否包含指定的节点类型
func (c NodeCategory) Contains(kind SyntaxKind) bool {
	return c.kinds[kind]
}

// Name 返回类别名称
func (c NodeCategory) Name() string {
	return c.name
}

// Kinds 返回类别中所有的节点类型
func (c NodeCategory) Kinds() []SyntaxKind {
	var kinds []SyntaxKind
	for k := range c.kinds {
		kinds = append(kinds, k)
	}
	return kinds
}

// =============================================================================
// 便捷的判断函数
// =============================================================================

// IsDeclaration 检查是否为声明类节点
func (m *NodeMethods) IsDeclaration() bool {
	return m.IsCategory(CategoryDeclarations)
}

// IsExpression 检查是否为表达式类节点
func (m *NodeMethods) IsExpression() bool {
	return m.IsCategory(CategoryExpressions)
}

// IsStatement 检查是否为语句类节点
func (m *NodeMethods) IsStatement() bool {
	return m.IsCategory(CategoryStatements)
}

// IsType 检查是否为类型相关节点
func (m *NodeMethods) IsType() bool {
	return m.IsCategory(CategoryTypes)
}

// IsLiteral 检查是否为字面量节点
func (m *NodeMethods) IsLiteral() bool {
	return m.IsCategory(CategoryLiterals)
}

// IsModule 检查是否为模块相关节点
func (m *NodeMethods) IsModule() bool {
	return m.IsCategory(CategoryModules)
}

// IsIdentifier 检查是否为标识符
func (m *NodeMethods) IsIdentifier() bool {
	return m.Is(KindIdentifier)
}

// =============================================================================
// 名称和值获取接口
// =============================================================================

// GetName 获取节点的名称（统一接口）
// 适用于各种具有名称的节点类型
func (m *NodeMethods) GetName() (string, bool) {
	text := m.node.GetText()
	if text == "" {
		return "", false
	}

	switch m.node.Kind {
	case KindFunctionDeclaration, KindInterfaceDeclaration,
		 KindClassDeclaration, KindTypeAliasDeclaration, KindEnumDeclaration,
		 KindVariableDeclaration:
		// 从节点文本中提取名称
		return extractNameFromText(text, m.node.Kind)
	case KindIdentifier:
		return text, true
	default:
		return "", false
	}
}

// GetValue 获取字面量节点的值
func (m *NodeMethods) GetValue() (interface{}, bool) {
	if !m.IsLiteral() {
		return nil, false
	}

	text := m.node.GetText()
	switch m.node.Kind {
	case KindStringLiteral:
		return extractStringValue(text), true
	case KindNumericLiteral:
		return extractNumericValue(text), true
	case KindTrueKeyword:
		return true, true
	case KindFalseKeyword:
		return false, true
	case KindNullKeyword:
		return nil, true
	case KindUndefinedKeyword:
		return nil, true
	default:
		return text, true
	}
}

// =============================================================================
// 辅助函数
// =============================================================================

// extractNameFromText 从节点文本中提取名称
func extractNameFromText(text string, kind SyntaxKind) (string, bool) {
	// 简单的名称提取逻辑
	// 实际实现中可以根据需要优化
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
		// 变量声明比较复杂，可能包含解构等
		return extractFirstIdentifier(text), true
	}
	return "", false
}

// 简单的字符串处理函数
func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func trimPrefix(s, prefix string) string {
	if startsWith(s, prefix) {
		return s[len(prefix):]
	}
	return s
}

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

func extractFirstIdentifier(s string) string {
	// 简单实现：提取第一个标识符
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

func isIdentifierChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '$'
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

func extractStringValue(text string) string {
	// 移除引号
	if len(text) >= 2 && (text[0] == '"' || text[0] == '\'') {
		return text[1 : len(text)-1]
	}
	return text
}

func extractNumericValue(text string) interface{} {
	// 简单实现，返回字符串
	// 实际中可以解析为 int 或 float
	return text
}