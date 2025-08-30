// package analyzer_tree 负责将扁平化的解析节点列表，构建成一个能够反映代码作用域和层级关系的树状结构。
package analyzer_tree

import (
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
	"github.com/samber/lo"
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

		// isContainer 标记当前节点是否是一个容器节点（如函数），
		// 在它所有子节点都处理完毕后，需要将它从上下文堆栈中弹出。
		isContainer := tp.dispatch(node)

		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})

		// 如果当前节点是容器，则在遍历完所有子节点后，将其从堆栈中弹出。
		if isContainer {
			tp.nodeStack = tp.nodeStack[:len(tp.nodeStack)-1]
		}
	}

	walk(tp.Ast)
}

// dispatch 是自定义的节点分发器，这是构建树的核心逻辑。
func (tp *TreeParser) dispatch(node *ast.Node) (isContainer bool) {
	isContainer = false
	parent := tp.nodeStack[len(tp.nodeStack)-1]

	switch n := node.AsNode(); n.Kind {
	case ast.KindFunctionDeclaration:
		fnDecl := n.AsFunctionDeclaration()
		declResult := parser.NewFunctionDeclarationResult(fnDecl, tp.SourceCode)
		fnNode := &FunctionNode{Declaration: *declResult, parent: parent}
		parent.AddChild(fnNode)
		tp.nodeStack = append(tp.nodeStack, fnNode)
		isContainer = true

	case ast.KindCallExpression:
		callExpr := n.AsCallExpression()
		if _, ok := tp.Parser.ProcessedDynamicImports[n]; ok {
			return
		}
		if callExpr.Expression.Kind == ast.KindImportKeyword {
			return
		}

		ce := parser.CallExpression{
			Expression: parser.AnalyzeVariableValueNode(callExpr.Expression, tp.SourceCode),
			CallChain:  parser.ReconstructCallChain(callExpr.Expression, tp.SourceCode),
			Arguments: lo.Map(callExpr.Arguments.Nodes, func(arg *ast.Node, _ int) *parser.VariableValue {
				return parser.AnalyzeVariableValueNode(arg, tp.SourceCode)
			}),
			Raw:            utils.GetNodeText(n, tp.SourceCode),
			SourceLocation: parser.NewSourceLocation(n, tp.SourceCode),
		}
		callNode := &CallNode{Call: ce, parent: parent}
		parent.AddChild(callNode)

	case ast.KindVariableStatement:
		varStmt := n.AsVariableStatement()
		declarations := parser.ExtractVariableDeclarations(varStmt, tp.SourceCode)
		for _, decl := range declarations {
			// 检查是否是函数赋值或动态导入，这些在基础 parser 中已经处理，这里要避免重复创建节点
			isFuncAssignment := false
			if len(decl.Declarators) > 0 && decl.Declarators[0].InitValue != nil {
				initType := decl.Declarators[0].InitValue.Type
				if initType == "arrowFunction" || initType == "functionExpression" {
					isFuncAssignment = true
				}
			}

			if !isFuncAssignment {
				varNode := &VariableNode{Declaration: decl, parent: parent}
				parent.AddChild(varNode)
			}
		}
	}

	return
}
