package tsmorphgo

import (
	"strings"
)

// =============================================================================
// TSMorphGo API 简化优化版本
// 专注于核心功能，避免复杂的实现冲突
// =============================================================================

// =============================================================================
// 1. 统一的类型检查API (方法形式)
// =============================================================================

func (n Node) IsFunction() bool {
	return n.Kind == KindFunctionDeclaration
}

func (n Node) IsCallExpression() bool {
	return n.Kind == KindCallExpression
}

func (n Node) IsVariable() bool {
	return n.Kind == KindVariableDeclaration
}

func (n Node) IsInterface() bool {
	return n.Kind == KindInterfaceDeclaration
}

func (n Node) IsClass() bool {
	return n.Kind == KindClassDeclaration
}

func (n Node) IsTypeAlias() bool {
	return n.Kind == KindTypeAliasDeclaration
}

func (n Node) IsEnum() bool {
	return n.Kind == KindEnumDeclaration
}

func (n Node) IsImport() bool {
	return n.Kind == KindImportDeclaration
}

func (n Node) IsExport() bool {
	return n.Kind == KindExportDeclaration
}

func (n Node) IsMemberAccess() bool {
	return n.Kind == KindPropertyAccessExpression
}

func (n Node) IsIdentifier() bool {
	return n.Kind == KindIdentifier
}

// =============================================================================
// 2. 基础信息提取API
// =============================================================================

// GetName 获取节点的名称（统一接口）
func (n Node) GetName() (string, bool) {
	text := strings.TrimSpace(n.GetText())
	if text == "" {
		return "", false
	}

	switch {
	case n.IsFunction():
		// 简单解析函数名
		if strings.HasPrefix(text, "function ") {
			name := strings.TrimPrefix(text, "function ")
			if spaceIdx := strings.IndexAny(name, " ("); spaceIdx != -1 {
				return strings.TrimSpace(name[:spaceIdx]), true
			}
			return strings.TrimSpace(name), true
		}
		// 匿名函数或箭头函数
		return "", false

	case n.IsVariable(), n.IsInterface(), n.IsClass(), n.IsEnum(), n.IsTypeAlias():
		// 查找第一个标识符作为名称
		parts := strings.Fields(text)
		for _, part := range parts {
			if strings.HasPrefix(part, strings.ToLower(part)) && len(part) > 1 {
				// 简单的标识符检查
				if strings.ContainsAny(part, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_") && !strings.ContainsAny(part, "=[]{}();,") {
					return part, true
				}
			}
		}
		return "", false

	case n.IsIdentifier():
		return text, true

	default:
		return "", false
	}
}

// GetType 获取节点的类型信息（简单实现）
func (n Node) GetType() string {
	text := strings.TrimSpace(n.GetText())

	// 查找类型注解
	if strings.Contains(text, ":") {
		parts := strings.SplitN(text, ":", 2)
		if len(parts) >= 2 {
			typePart := strings.TrimSpace(parts[1])
			// 移除初始化部分
			if equalsIdx := strings.Index(typePart, "="); equalsIdx != -1 {
				typePart = strings.TrimSpace(typePart[:equalsIdx])
			}
			// 移除函数体部分
			if braceIdx := strings.Index(typePart, "{"); braceIdx != -1 {
				typePart = strings.TrimSpace(typePart[:braceIdx])
			}
			return typePart
		}
	}

	return "unknown"
}

// IsExported 检查节点是否为导出的
func (n Node) IsExported() bool {
	// 检查节点本身的文本
	text := strings.ToLower(n.GetText())
	if strings.Contains(text, "export") {
		return true
	}

	// 检查父节点
	parent := n.GetParent()
	for parent != nil {
		parentText := strings.ToLower(parent.GetText())
		if strings.Contains(parentText, "export") {
			return true
		}
		parent = parent.GetParent()
	}

	return false
}

// IsAsync 检查是否为异步函数
func (n Node) IsAsync() bool {
	if !n.IsFunction() {
		return false
	}
	text := strings.ToLower(n.GetText())
	return strings.Contains(text, "async")
}

// IsConst 检查是否为const声明
func (n Node) IsConst() bool {
	if !n.IsVariable() {
		return false
	}
	parent := n.GetParent()
	for parent != nil {
		parentText := strings.ToLower(parent.GetText())
		if strings.Contains(parentText, "const") {
			return true
		}
		parent = parent.GetParent()
	}
	return false
}

// IsLet 检查是否为let声明
func (n Node) IsLet() bool {
	if !n.IsVariable() {
		return false
	}
	parent := n.GetParent()
	for parent != nil {
		parentText := strings.ToLower(parent.GetText())
		if strings.Contains(parentText, "let") && !strings.Contains(parentText, "const") {
			return true
		}
		parent = parent.GetParent()
	}
	return false
}

// =============================================================================
// 3. 项目级搜索API
// =============================================================================

// FindFunctions 在项目中查找所有函数
func (p *Project) FindFunctions() []Node {
	var functions []Node

	for _, file := range p.GetSourceFiles() {
		file.ForEachDescendant(func(node Node) {
			if node.IsFunction() {
				functions = append(functions, node)
			}
		})
	}

	return functions
}

// FindExportedFunctions 在项目中查找所有导出的函数
func (p *Project) FindExportedFunctions() []Node {
	var functions []Node

	for _, file := range p.GetSourceFiles() {
		file.ForEachDescendant(func(node Node) {
			if node.IsFunction() && node.IsExported() {
				functions = append(functions, node)
			}
		})
	}

	return functions
}

// FindVariables 在项目中查找所有变量
func (p *Project) FindVariables() []Node {
	var variables []Node

	for _, file := range p.GetSourceFiles() {
		file.ForEachDescendant(func(node Node) {
			if node.IsVariable() {
				variables = append(variables, node)
			}
		})
	}

	return variables
}

// FindInterfaces 在项目中查找所有接口
func (p *Project) FindInterfaces() []Node {
	var interfaces []Node

	for _, file := range p.GetSourceFiles() {
		file.ForEachDescendant(func(node Node) {
			if node.IsInterface() {
				interfaces = append(interfaces, node)
			}
		})
	}

	return interfaces
}

// FindClasses 在项目中查找所有类
func (p *Project) FindClasses() []Node {
	var classes []Node

	for _, file := range p.GetSourceFiles() {
		file.ForEachDescendant(func(node Node) {
			if node.IsClass() {
				classes = append(classes, node)
			}
		})
	}

	return classes
}

// FindNodesByKind 根据节点类型查找节点
func (p *Project) FindNodesByKind(kind SyntaxKind) []Node {
	var nodes []Node

	for _, file := range p.GetSourceFiles() {
		file.ForEachDescendant(func(node Node) {
			if node.Kind == kind {
				nodes = append(nodes, node)
			}
		})
	}

	return nodes
}

// =============================================================================
// 4. 便捷的复合检查API
// =============================================================================

func (n Node) IsExportedFunction() bool {
	return n.IsFunction() && n.IsExported()
}

func (n Node) IsExportedVariable() bool {
	return n.IsVariable() && n.IsExported()
}

func (n Node) IsExportedInterface() bool {
	return n.IsInterface() && n.IsExported()
}

func (n Node) IsExportedClass() bool {
	return n.IsClass() && n.IsExported()
}

func (n Node) IsArrowFunction() bool {
	if !n.IsFunction() {
		return false
	}
	text := strings.TrimSpace(n.GetText())
	return strings.Contains(text, "=>")
}

func (n Node) IsAsyncFunction() bool {
	return n.IsFunction() && n.IsAsync()
}

// =============================================================================
// 5. 简单的搜索辅助方法
// =============================================================================

// FindFirstFunctionInParent 在父节点中查找第一个函数
func (n Node) FindFirstFunctionInParent() *Node {
	parent := n.GetParent()
	for parent != nil {
		if parent.IsFunction() {
			return parent
		}
		parent = parent.GetParent()
	}
	return nil
}

// IsEmpty 检查节点是否为空
func (n Node) IsEmpty() bool {
	return n.Node == nil || strings.TrimSpace(n.GetText()) == ""
}