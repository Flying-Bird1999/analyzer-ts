package tsmorphgo

import (
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// =============================================================================
// 统一的 Node API - 简洁版本
// 这个文件提供了统一的接口来替代所有分散的 IsXXX 和 AsXXX 函数
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
// 便捷的类别检查方法
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
// 常用类型的便捷检查方法
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
// 类型转换的统一接口
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
// 名称和值获取的统一接口
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
// 内部辅助函数
// =============================================================================

// extractVariableNameFromText 从文本中提取变量名
func extractVariableNameFromText(text string) (string, bool) {
	// 简单实现：查找第一个标识符
	var name []rune
	for _, r := range text {
		if isIdentifierChar(r) {
			name = append(name, r)
		} else if len(name) > 0 {
			return string(name), true
		} else if r == '=' || r == '{' || r == '[' {
			// 遇到赋值或解构模式
			return "destructured pattern", true
		}
	}

	if len(name) > 0 {
		return string(name), true
	}
	return "", false
}