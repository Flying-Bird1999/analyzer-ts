// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（callExpression.go）专门负责处理和解析函数/方法调用表达式。
package parser

import (
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
	"github.com/samber/lo"
)

// CallExpression 代表一个函数或方法调用表达式的解析结果。
// 它现在可以包含在参数中发现的内联函数声明。
type CallExpression struct {
	Expression      *VariableValue              `json:"expression"`      // [权威] 被调用的表达式的完整结构化信息
	CallChain       []string                    `json:"callChain"`       // [便利] 表达式的调用链视图，例如 ["myObj", "method", "call"]
	Arguments       []*VariableValue            `json:"arguments"`       // 调用时传递的参数列表。
	InlineFunctions []FunctionDeclarationResult `json:"inlineFunctions"` // [新增] 在参数中发现的内联函数（例如 useEffect 的回调）
	Raw             string                      `json:"raw,omitempty"`   // 节点在源码中的原始文本。
	SourceLocation *SourceLocation `json:"sourceLocation,omitempty"` // 表达式在源码中的位置
}

// ReconstructCallChain 是一个辅助函数，用于从表达式节点递归地构建一个简单的字符串调用链。
func ReconstructCallChain(node *ast.Node, sourceCode string) []string {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return []string{node.AsIdentifier().Text}
	case ast.KindPropertyAccessExpression:
		propAccess := node.AsPropertyAccessExpression()
		left := ReconstructCallChain(propAccess.Expression, sourceCode)
		return append(left, propAccess.Name().Text())
	default:
		// 对于其他复杂情况（例如 getFunc()()），返回其源码文本作为唯一标识
		return []string{strings.TrimSpace(utils.GetNodeText(node, sourceCode))}
	}
}

// AnalyzeCallExpression 是一个公共的、可复用的函数，用于从 AST 节点中解析函数调用表达式。
// 它现在还会检查参数，以查找并解析内联的箭头函数或函数表达式。
func AnalyzeCallExpression(node *ast.CallExpression, sourceCode string, processedDynamicImports map[*ast.Node]bool) (*CallExpression, *ImportDeclarationResult) {
	if node == nil {
		return nil, nil
	}

	// 动态导入（赋值给变量的）已在变量声明处处理，这里跳过以避免重复
	if _, ok := processedDynamicImports[node.AsNode()]; ok {
		return nil, nil
	}

	// 检查是否是独立的动态导入 `import(...)`
	if node.Expression.Kind == ast.KindImportKeyword {
		if len(node.Arguments.Nodes) > 0 {
			arg := node.Arguments.Nodes[0]
			var importPath string
			if arg.Kind == ast.KindStringLiteral {
				importPath = arg.AsStringLiteral().Text
			} else if arg.Kind == ast.KindIdentifier {
				importPath = arg.AsIdentifier().Text
			} else {
				return nil, nil // 不支持的动态导入参数类型
			}

			importResult := &ImportDeclarationResult{
				Source: importPath,
				ImportModules: []ImportModule{
					{
						Identifier:   "default", // 独立的动态导入，我们将其视为默认导入
						ImportModule: "default",
						Type:         "dynamic",
					},
				},
				Raw: utils.GetNodeText(node.AsNode(), sourceCode),
			}
			return nil, importResult // 返回一个导入声明结果，而不是调用表达式
		}
		return nil, nil // 参数不合法的动态导入
	}

	// --- 新增逻辑：检查参数中的内联函数 ---
	inlineFunctions := []FunctionDeclarationResult{}
	for _, arg := range node.Arguments.Nodes {
		if arg.Kind == ast.KindArrowFunction || arg.Kind == ast.KindFunctionExpression {
			// 为这个匿名的内联函数创建一个函数声明结果
			// 标识符为空，因为它没有名字
			fnResult := NewFunctionDeclarationResultFromExpression("", false, arg, sourceCode)
			inlineFunctions = append(inlineFunctions, *fnResult)
		}
	}

	ce := &CallExpression{
		Expression:      AnalyzeVariableValueNode(node.Expression, sourceCode),
		CallChain:       ReconstructCallChain(node.Expression, sourceCode),
		Arguments:       lo.Map(node.Arguments.Nodes, func(arg *ast.Node, _ int) *VariableValue {
			return AnalyzeVariableValueNode(arg, sourceCode)
		}),
		InlineFunctions: inlineFunctions, // 存储找到的内联函数
		Raw:             utils.GetNodeText(node.AsNode(), sourceCode),
		SourceLocation:  NewSourceLocation(node.AsNode(), sourceCode),
	}

	return ce, nil
}

// VisitCallExpression 从给定的 ast.CallExpression 节点中提取详细信息。
func (p *Parser) VisitCallExpression(node *ast.CallExpression) {
	callExpr, importDecl := AnalyzeCallExpression(node, p.SourceCode, p.ProcessedDynamicImports)

	if importDecl != nil {
		p.Result.ImportDeclarations = append(p.Result.ImportDeclarations, *importDecl)
	}

	if callExpr != nil {
		p.Result.CallExpressions = append(p.Result.CallExpressions, *callExpr)
	}
}
