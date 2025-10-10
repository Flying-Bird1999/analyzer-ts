package parser

import (
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// ExtractedNodes 用于存储从文件中提取出的各种节点信息。
// 当需要添加新的节点类型时，只需在此结构体中添加新的字段。
type ExtractedNodes struct {
	AnyDeclarations []AnyInfo      `json:"anyDeclarations"` // 存储找到的所有 any 类型的信息
	AsExpressions   []AsExpression `json:"asExpressions"`   // 存储找到的所有 as 表达式的信息
	// 后续新增其他节点类型时，同步添加在下方
}

// AnyInfo 存储了在文件中找到的 any 类型的信息。
type AnyInfo struct {
	SourceLocation SourceLocation `json:"sourceLocation"`
	Raw            string         `json:"raw"` // 存储 any 关键字的原始文本
	Node           *ast.Node      `json:"-"`   // 对应的 AST 节点，不在 JSON 中序列化。
}

// AsExpression 代表一个解析后的 'as' 类型断言表达式。
type AsExpression struct {
	Raw            string         `json:"raw"`            // 节点在源码中的原始文本。
	SourceLocation SourceLocation `json:"sourceLocation"` // 节点在源码中的位置信息。
	Node           *ast.Node      `json:"-"`              // 对应的 AST 节点，不在 JSON 中序列化。
}

// VisitAnyKeyword 解析 any 关键字。
func (p *Parser) VisitAnyKeyword(node *ast.Node) {
	anyInfo := AnyInfo{
		SourceLocation: SourceLocation{
			Start: func() NodePosition {
				line, character := utils.GetLineAndCharacterOfPosition(p.SourceCode, node.Loc.Pos())
				return NodePosition{Line: line + 1, Column: character + 1}
			}(),
			End: func() NodePosition {
				line, character := utils.GetLineAndCharacterOfPosition(p.SourceCode, node.Loc.End())
				return NodePosition{Line: line + 1, Column: character}
			}(),
		},
		Raw: func() string {
			line, _ := utils.GetLineAndCharacterOfPosition(p.SourceCode, node.Loc.Pos())
			lines := strings.Split(p.SourceCode, "\n")
			if line >= 0 && line < len(lines) {
				return strings.TrimSpace(lines[line])
			}
			return ""
		}(),
		Node: node,
	}
	p.Result.ExtractedNodes.AnyDeclarations = append(p.Result.ExtractedNodes.AnyDeclarations, anyInfo)
}

// VisitAsExpression 解析 as 表达式。
func (p *Parser) VisitAsExpression(node *ast.AsExpression) {
	asExpr := AsExpression{
		Raw: func() string {
			line, _ := utils.GetLineAndCharacterOfPosition(p.SourceCode, node.AsNode().Loc.Pos())
			lines := strings.Split(p.SourceCode, "\n")
			if line >= 0 && line < len(lines) {
				return strings.TrimSpace(lines[line])
			}
			return ""
		}(),
		SourceLocation: SourceLocation{
			Start: func() NodePosition {
				line, character := utils.GetLineAndCharacterOfPosition(p.SourceCode, node.Expression.Loc.Pos())
				return NodePosition{Line: line + 1, Column: character + 1}
			}(),
			End: func() NodePosition {
				line, character := utils.GetLineAndCharacterOfPosition(p.SourceCode, node.Type.Loc.End())
				return NodePosition{Line: line + 1, Column: character}
			}(),
		},
		Node: node.AsNode()}
	p.Result.ExtractedNodes.AsExpressions = append(p.Result.ExtractedNodes.AsExpressions, asExpr)
}
