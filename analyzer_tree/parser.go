// package analyzer_tree 负责将扁平化的解析节点列表，构建成一个能够反映代码作用域和层级关系的树状结构。
package analyzer_tree

import (
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// TreeParser 是一个专门用于构建层级关系树的解析器。
// 它通过嵌入基础的 `parser.Parser` 继承了其通用的 AST 遍历能力，
// 然后通过重新实现 `Traverse` 和 `dispatch` 方法，在遍历过程中加入了上下文感知和节点构建的逻辑。
type TreeParser struct {
	*parser.Parser // 嵌入基础解析器，复用其文件读取和 AST 生成功能

	Tree      *RootNode // 构建完成的树的根节点
	nodeStack []Node    // 用于跟踪当前作用域的上下文堆栈
}

// NewTreeParser 是 TreeParser 的构造函数。
func NewTreeParser(filePath string) (*TreeParser, error) {
	baseParser, err := parser.NewParser(filePath)
	if err != nil {
		return nil, err
	}
	return newTreeParser(baseParser), nil
}

// NewTreeParserFromSource 是一个主要用于测试的构造函数，它直接从源码字符串创建解析器。
func NewTreeParserFromSource(filePath string, sourceCode string) (*TreeParser, error) {
	baseParser, err := parser.NewParserFromSource(filePath, sourceCode)
	if err != nil {
		return nil, err
	}
	return newTreeParser(baseParser), nil
}

// newTreeParser 是一个内部辅助函数，用于完成 TreeParser 的初始化。
func newTreeParser(baseParser *parser.Parser) *TreeParser {
	root := &RootNode{}
	tp := &TreeParser{
		Parser:    baseParser,
		Tree:      root,
		nodeStack: []Node{root}, // 将根节点作为初始上下文压入堆栈
	}
	return tp
}

// Traverse 是一个全新的遍历入口，它会覆盖基础解析器的默认行为。
// 它启动一个递归的 walk 函数，该函数在遍历 AST 的同时，调用自定义的 dispatch 方法来构建树。
func (tp *TreeParser) Traverse() {
	var walk func(node *ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}

		// isContainer 标记当前节点是否是一个容器节点（如函数、JSX元素），
		// 在它所有子节点都处理完毕后，需要将它从上下文堆栈中弹出。
		// continueWalk 控制是否要继续遍历当前节点的子节点。
		isContainer, continueWalk := tp.dispatch(node)

		// 如果分发器指示不要继续，则直接返回，停止对该分支的深入遍历。
		if !continueWalk {
			return
		}

		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false // 返回 false 以确保遍历所有子节点
		})

		// 如果当前节点是容器，则在遍历完所有子节点后，将其从堆栈中弹出，返回到父级作用域。
		if isContainer {
			tp.nodeStack = tp.nodeStack[:len(tp.nodeStack)-1]
		}
	}

	walk(tp.Ast)
}

// dispatch 是自定义的节点分发器，这是构建树的核心逻辑。
// 它根据 AST 节点的类型，创建对应的树节点，并正确处理容器和叶子节点，维护作用域堆栈。
// 返回 isContainer 和 continueWalk 两个布尔值。
func (tp *TreeParser) dispatch(node *ast.Node) (isContainer bool, continueWalk bool) {
	// 默认行为：不是容器，并继续遍历子节点
	isContainer = false
	continueWalk = true

	// 从堆栈顶部获取当前父节点
	parent := tp.nodeStack[len(tp.nodeStack)-1]

	switch n := node.AsNode(); n.Kind {

	// --- 容器节点: Function, JsxElement ---
	// 对于容器节点，我们将 isContainer 设为 true，并在处理后压入堆栈。
	// 通常我们希望继续遍历其子节点，所以 continueWalk 保持 true。

	case ast.KindFunctionDeclaration:
		fnDecl := n.AsFunctionDeclaration()
		declResult := parser.NewFunctionDeclarationResult(fnDecl, tp.SourceCode)
		fnNode := &FunctionNode{Declaration: *declResult}
		parent.AddChild(fnNode)
		tp.nodeStack = append(tp.nodeStack, fnNode)
		isContainer = true

	case ast.KindJsxElement, ast.KindJsxSelfClosingElement:
		declResult := parser.AnalyzeJsxElement(n, tp.SourceCode)
		jsxNode := &JsxNode{Declaration: *declResult}
		parent.AddChild(jsxNode)
		tp.nodeStack = append(tp.nodeStack, jsxNode)
		isContainer = true

	// --- 块级叶子节点: Variable, Interface, Enum, TypeAlias, Import, Export ---
	// 这些节点我们只关心其本身，不关心其内部的子节点（因为信息已经在 Analyze... 函数中提取）。
	// 所以我们将 continueWalk 设为 false，以停止对这些分支的深入遍历。

	case ast.KindVariableStatement:
		varStmt := n.AsVariableStatement()
		declarations := parser.ExtractVariableDeclarations(varStmt, tp.SourceCode)
		for _, decl := range declarations {
			// 检查是否是函数赋值
			if len(decl.Declarators) > 0 && decl.Declarators[0].InitValue != nil {
				initType := decl.Declarators[0].InitValue.Type
				if initType == "arrowFunction" || initType == "functionExpression" {
					// 如果是函数表达式，则作为 FunctionNode 处理
					// 为了简化，我们暂时跳过，但这是一个可以完善的点。
					continue
				}
			}
			varNode := &VariableNode{Declaration: decl}
			parent.AddChild(varNode)
		}
		continueWalk = false

	case ast.KindInterfaceDeclaration:
		interfaceDecl := n.AsInterfaceDeclaration()
		declResult := parser.AnalyzeInterfaceDeclaration(interfaceDecl, tp.SourceCode)
		interfaceNode := &InterfaceNode{Declaration: *declResult}
		parent.AddChild(interfaceNode)
		continueWalk = false

	case ast.KindEnumDeclaration:
		enumDecl := n.AsEnumDeclaration()
		declResult := parser.AnalyzeEnumDeclaration(enumDecl, tp.SourceCode)
		enumNode := &EnumNode{Declaration: *declResult}
		parent.AddChild(enumNode)
		continueWalk = false

	case ast.KindTypeAliasDeclaration:
		typeAliasDecl := n.AsTypeAliasDeclaration()
		declResult := parser.AnalyzeTypeAliasDeclaration(typeAliasDecl, tp.SourceCode)
		typeAliasNode := &TypeAliasNode{Declaration: *declResult}
		parent.AddChild(typeAliasNode)
		continueWalk = false

	case ast.KindImportDeclaration:
		importDecl := n.AsImportDeclaration()
		declResult := parser.AnalyzeImportDeclaration(importDecl, tp.SourceCode)
		importNode := &ImportNode{Declaration: *declResult}
		parent.AddChild(importNode)
		continueWalk = false

	case ast.KindExportDeclaration:
		exportDecl := n.AsExportDeclaration()
		declResult := parser.AnalyzeExportDeclaration(exportDecl, tp.SourceCode)
		exportNode := &ExportNode{Declaration: *declResult}
		parent.AddChild(exportNode)
		continueWalk = false

	case ast.KindExportAssignment:
		exportAssign := n.AsExportAssignment()
		declResult := parser.AnalyzeExportAssignment(exportAssign, tp.SourceCode)
		exportAssignNode := &ExportAssignmentNode{Declaration: *declResult}
		parent.AddChild(exportAssignNode)
		continueWalk = false

	// --- 表达式级叶子节点: CallExpression ---
	// 只有当 CallExpression 是一个独立的表达式语句的一部分时，我们才处理它。
	// 其他情况（如在变量声明中）已经被包含在父节点的解析逻辑中。
	case ast.KindCallExpression:
		callExpr := n.AsCallExpression()
		callResult, importResult := parser.AnalyzeCallExpression(callExpr, tp.SourceCode, tp.ProcessedDynamicImports)

		if importResult != nil {
			importNode := &ImportNode{Declaration: *importResult}
			parent.AddChild(importNode)
		} else if callResult != nil {
			callNode := &CallNode{Call: *callResult}
			parent.AddChild(callNode)
		}
	}

	return
}
