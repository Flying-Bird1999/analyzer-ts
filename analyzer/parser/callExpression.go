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
// 它同时包含结构化的 Expression 字段和易于使用的 CallChain 字段。
type CallExpression struct {
	Expression     *VariableValue   `json:"expression"`     // [权威] 被调用的表达式的完整结构化信息
	CallChain      []string         `json:"callChain"`      // [便利] 表达式的调用链视图，例如 ["myObj", "method", "call"]
	Arguments      []*VariableValue `json:"arguments"`      // 调用时传递的参数列表。
	Raw            string           `json:"raw,omitempty"`  // 节点在源码中的原始文本。
	SourceLocation SourceLocation   `json:"sourceLocation"` // 节点在源码中的位置信息。
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

// VisitCallExpression 从给定的 ast.CallExpression 节点中提取详细信息。
func (p *Parser) VisitCallExpression(node *ast.CallExpression) {
	if node == nil {
		return
	}

	// 动态导入（赋值给变量的）已在变量声明处处理，这里跳过以避免重复
	if _, ok := p.ProcessedDynamicImports[node.AsNode()]; ok {
		return
	}

	// 检查是否是独立的动态导入 `import(...)`
	if node.Expression.Kind == ast.KindImportKeyword {
		// 这是一个独立的、未赋值给变量的动态导入
		if len(node.Arguments.Nodes) > 0 {
			arg := node.Arguments.Nodes[0]
			var importPath string
			if arg.Kind == ast.KindStringLiteral {
				importPath = arg.AsStringLiteral().Text
			} else if arg.Kind == ast.KindIdentifier {
				importPath = arg.AsIdentifier().Text
			} else {
				return // 不支持的动态导入参数类型
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
				Raw: utils.GetNodeText(node.AsNode(), p.SourceCode),
			}
			p.Result.ImportDeclarations = append(p.Result.ImportDeclarations, *importResult)
		}
		return // 处理完毕，不再作为常规 CallExpression 添加
	}

	ce := CallExpression{
		Expression: AnalyzeVariableValueNode(node.Expression, p.SourceCode),
		CallChain:  ReconstructCallChain(node.Expression, p.SourceCode),
		Arguments: lo.Map(node.Arguments.Nodes, func(arg *ast.Node, _ int) *VariableValue {
			return AnalyzeVariableValueNode(arg, p.SourceCode)
		}),
		Raw:            utils.GetNodeText(node.AsNode(), p.SourceCode),
		SourceLocation: NewSourceLocation(node.AsNode(), p.SourceCode),
	}

	p.Result.CallExpressions = append(p.Result.CallExpressions, ce)
}
