package symbol_analysis

import (
	"fmt"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

// extractSymbolChange 从 AST 节点提取符号变更信息。
func (a *Analyzer) extractSymbolChange(
	node tsmorphgo.Node,
	changedLines map[int]bool,
) *SymbolChange {
	// 提取符号名称
	name := a.extractSymbolName(node)
	if name == "" {
		return nil
	}

	// 确定符号类型
	kind := a.determineSymbolKind(node)

	// 计算变更行
	actualChanges := a.calculateChangedLines(node, changedLines)

	// 确定变更类型
	changeType := a.determineChangeType(node, changedLines)

	return &SymbolChange{
		Name:         name,
		Kind:         kind,
		FilePath:     node.GetSourceFile().GetFilePath(),
		StartLine:    node.GetStartLineNumber(),
		EndLine:      node.GetEndLineNumber(),
		ChangedLines: actualChanges,
		ChangeType:   changeType,
		ExportType:   ExportTypeNone, // 稍后填充
		IsExported:   false,          // 稍后填充
	}
}

// extractSymbolName 从 AST 节点提取符号名称。
func (a *Analyzer) extractSymbolName(node tsmorphgo.Node) string {
	// 对于函数声明
	if node.IsFunctionDeclaration() {
		if fnDecl, ok := node.AsFunctionDeclaration(); ok {
			return fnDecl.GetName()
		}
	}

	// 对于变量声明
	if node.IsVariableDeclaration() {
		if varDecl, ok := node.AsVariableDeclaration(); ok {
			return varDecl.GetName()
		}
	}

	// 对于类、接口、枚举、类型别名声明
	if node.IsClassDeclaration() ||
		node.IsInterfaceDeclaration() ||
		node.IsEnumDeclaration() ||
		node.IsTypeAliasDeclaration() {
		// 查找第一个标识符子节点
		var name string
		node.ForEachChild(func(child tsmorphgo.Node) bool {
			if child.IsIdentifier() {
				name = child.GetText()
				return true // 找到了，停止遍历
			}
			return false
		})
		return name
	}

	// 对于方法声明（类方法）
	if node.IsKind(tsmorphgo.KindMethodDeclaration) {
		var name string
		node.ForEachChild(func(child tsmorphgo.Node) bool {
			if child.IsIdentifier() {
				name = child.GetText()
				return true
			}
			return false
		})
		return name
	}

	return ""
}

// determineSymbolKind 从 AST 节点确定符号类型。
func (a *Analyzer) determineSymbolKind(node tsmorphgo.Node) SymbolKind {
	switch {
	case node.IsFunctionDeclaration():
		return SymbolKindFunction
	case node.IsVariableDeclaration():
		return SymbolKindVariable
	case node.IsClassDeclaration():
		return SymbolKindClass
	case node.IsInterfaceDeclaration():
		return SymbolKindInterface
	case node.IsTypeAliasDeclaration():
		return SymbolKindTypeAlias
	case node.IsEnumDeclaration():
		return SymbolKindEnum
	case node.IsKind(tsmorphgo.KindMethodDeclaration):
		return SymbolKindMethod
	default:
		// 未知类型默认为函数
		return SymbolKindFunction
	}
}

// calculateChangedLines 计算节点内实际变更的行。
func (a *Analyzer) calculateChangedLines(
	node tsmorphgo.Node,
	changedLines map[int]bool,
) []int {
	var changes []int
	startLine := node.GetStartLineNumber()
	endLine := node.GetEndLineNumber()

	for line := startLine; line <= endLine; line++ {
		if changedLines[line] {
			changes = append(changes, line)
		}
	}

	return changes
}

// determineChangeType 根据节点和变更行确定变更类型。
func (a *Analyzer) determineChangeType(
	node tsmorphgo.Node,
	changedLines map[int]bool,
) ChangeType {
	// 检查是否所有行都变更了（可能是新增/删除）
	startLine := node.GetStartLineNumber()
	endLine := node.GetEndLineNumber()

	_ = endLine - startLine + 1 // 总行数（预留供将来使用）
	changedCount := len(a.calculateChangedLines(node, changedLines))

	// 如果没有行变更，则不是变更（不应该发生）
	if changedCount == 0 {
		return ChangeTypeModified
	}

	// 如果全部或大部分行都变更了，可能是新增/删除
	// 目前默认为修改，因为它是最常见的
	return ChangeTypeModified
}

// isDeclarationNode 检查节点是否为我们关心的顶层声明节点。
// 它排除了内部声明，如函数内部和类方法中的变量。
func (a *Analyzer) isDeclarationNode(node tsmorphgo.Node) bool {
	if !a.options.IncludeTypes {
		// 如果不包含类型，则跳过仅类型的声明
		if node.IsInterfaceDeclaration() ||
			node.IsTypeAliasDeclaration() {
			return false
		}
	}

	// 我们关心的顶层声明
	isTopLevelDeclaration := node.IsFunctionDeclaration() ||
		node.IsClassDeclaration() ||
		node.IsInterfaceDeclaration() ||
		node.IsTypeAliasDeclaration() ||
		node.IsEnumDeclaration()

	if isTopLevelDeclaration {
		return true
	}

	// 排除类方法 - 它们应该由其父类处理
	// 当变更发生在方法内部时，我们想要找到类，而不是方法
	if node.IsKind(tsmorphgo.KindMethodDeclaration) {
		return false
	}

	// 对于变量声明，仅当它是顶层变量时才包含
	//（不在函数、类等内部）
	if node.IsVariableDeclaration() {
		return a.isTopLevelVariable(node)
	}

	return false
}

// isTopLevelVariable 检查变量声明是否在顶层
//（不在函数、类、接口等内部）
func (a *Analyzer) isTopLevelVariable(node tsmorphgo.Node) bool {
	parent := node.GetParent()

	// 没有父节点意味着它可能在源文件级别
	if parent == nil || !parent.IsValid() {
		return true
	}

	// 检查父节点是否是顶层容器
	// 如果父节点是 SourceFile，则它在顶层
	if parent.IsKind(tsmorphgo.KindSourceFile) {
		return true
	}

	// 如果父节点是函数、类、接口、块等，则它不在顶层
	// 我们检查常见的容器类型
	if parent.IsFunctionDeclaration() ||
		parent.IsClassDeclaration() ||
		parent.IsInterfaceDeclaration() ||
		parent.IsKind(tsmorphgo.KindBlock) { // Block 是 {...}
		return false
	}

	// 递归检查父节点的父节点（用于嵌套情况）
	p := parent
	return a.isTopLevelVariable(*p)
}

// shouldIncludeSymbol 检查符号是否应该包含在分析中。
func (a *Analyzer) shouldIncludeSymbol(node tsmorphgo.Node) bool {
	// 始终包含导出符号
	if a.isExportedNode(node) {
		return true
	}

	// 仅当选项设置时才包含内部符号
	return a.options.IncludeInternal
}

// isExportedNode 检查节点是否已导出。
func (a *Analyzer) isExportedNode(node tsmorphgo.Node) bool {
	// 检查任何父节点是否是导出声明
	parent := node.GetParent()
	for parent != nil {
		if parent.IsExportDeclaration() {
			return true
		}
		// 同时检查 export 关键字修饰符
		text := parent.GetText()
		if len(text) > 6 && text[:6] == "export" {
			return true
		}
		parent = parent.GetParent()
	}
	return false
}

// FormatSymbolID 为符号创建唯一标识符。
func FormatSymbolID(filePath, symbolName string, startLine int) string {
	return fmt.Sprintf("%s:%s:%d", filePath, symbolName, startLine)
}
