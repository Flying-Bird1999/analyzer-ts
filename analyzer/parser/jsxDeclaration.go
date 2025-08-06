package parser

import (
	"main/analyzer/utils"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// JSXAttribute 表示 JSX 元素中的一个属性
type JSXAttribute struct {
	Name  string `json:"name"`  // 属性名
	Value string `json:"value"` // 属性值
}

// JSXNode 表示一个 JSX 节点
type JSXNode struct {
	Type               string         `json:"type"`               // 节点类型，例如 "JSXIdentifier" 或 "JSXMemberExpression"
	ModuleIdentifier   string         `json:"moduleIdentifier"`   // 模块标识符，例如 <Antd.Button /> 中的 "Antd"
	PropertyIdentifier string         `json:"propertyIdentifier"` // 属性标识符，例如 <Button /> 中的 "Button" 或 <Antd.Button /> 中的 "Button"
	Attrs              []JSXAttribute `json:"attrs"`              // JSX 属性列表
	Raw                string         `json:"raw"`                // 源码
	SourceLocation     SourceLocation `json:"sourceLocation"`     // 源码位置
}

func NewJSXNode(node ast.Node, sourceCode string) *JSXNode {
	pos, end := node.Pos(), node.End()
	jsxNode := &JSXNode{
		Raw: utils.GetNodeText(node.AsNode(), sourceCode),
		SourceLocation: SourceLocation{
			Start: NodePosition{Line: pos, Column: 0},
			End:   NodePosition{Line: end, Column: 0},
		},
	}
	if node.Kind == ast.KindJsxElement {
		jsxNode.analyzeJsxElement(node.AsJsxElement(), sourceCode)
	} else if node.Kind == ast.KindJsxSelfClosingElement {
		jsxNode.analyzeJsxSelfClosingElement(node.AsJsxSelfClosingElement(), sourceCode)
	}
	return jsxNode
}

func (j *JSXNode) analyzeJsxElement(node *ast.JsxElement, sourceCode string) {
	// 修正：直接传递 OpeningElement 节点
	j.analyzeJsxOpeningElement(node.OpeningElement, sourceCode)
}

// 修正：参数类型改为 *ast.Node
func (j *JSXNode) analyzeJsxOpeningElement(node *ast.Node, sourceCode string) {
	// 修正：先转换为 JsxOpeningElement 类型
	openingElement := node.AsJsxOpeningElement()

	switch openingElement.TagName.Kind {
	case ast.KindIdentifier:
		j.Type = "JSXIdentifier"
		j.ModuleIdentifier = openingElement.TagName.AsIdentifier().Text
	case ast.KindPropertyAccessExpression:
		expr := openingElement.TagName.AsPropertyAccessExpression()
		if expr.Expression.Kind == ast.KindIdentifier {
			j.Type = "JSXMemberExpression"
			j.ModuleIdentifier = expr.Expression.AsIdentifier().Text
			j.PropertyIdentifier = expr.Name().Text()
		}
	}

	// 修正：简化属性访问方式
	if attributes := openingElement.Attributes; attributes != nil {
		// 修正：直接访问 Properties 字段
		if jsxAttrs := attributes.AsJsxAttributes(); jsxAttrs != nil && jsxAttrs.Properties != nil {
			for _, attr := range jsxAttrs.Properties.Nodes {
				if attr.Kind == ast.KindJsxAttribute {
					jsxAttr := attr.AsJsxAttribute()
					name := jsxAttr.Name().Text()
					var value string
					if jsxAttr.Initializer != nil {
						// 修正：直接传递节点，不需要解引用
						value = getTextFromNode(jsxAttr.Initializer)
					}
					j.Attrs = append(j.Attrs, JSXAttribute{Name: name, Value: value})
				}
			}
		}
	}
}

func (j *JSXNode) analyzeJsxSelfClosingElement(node *ast.JsxSelfClosingElement, sourceCode string) {
	switch node.TagName.Kind {
	case ast.KindIdentifier:
		j.Type = "JSXIdentifier"
		j.PropertyIdentifier = node.TagName.AsIdentifier().Text
	case ast.KindPropertyAccessExpression:
		expr := node.TagName.AsPropertyAccessExpression()
		if expr.Expression.Kind == ast.KindIdentifier {
			j.Type = "JSXMemberExpression"
			j.ModuleIdentifier = expr.Expression.AsIdentifier().Text
			j.PropertyIdentifier = expr.Name().Text()
		}
	}

	// 修正：简化属性访问方式
	if attributes := node.Attributes; attributes != nil {
		if jsxAttrs := attributes.AsJsxAttributes(); jsxAttrs != nil && jsxAttrs.Properties != nil {
			for _, attr := range jsxAttrs.Properties.Nodes {
				if attr.Kind == ast.KindJsxAttribute {
					jsxAttr := attr.AsJsxAttribute()
					name := jsxAttr.Name().Text()
					var value string
					if jsxAttr.Initializer != nil {
						// 修正：直接传递节点
						value = getTextFromNode(jsxAttr.Initializer)
					}
					j.Attrs = append(j.Attrs, JSXAttribute{Name: name, Value: value})
				}
			}
		}
	}
}

// 修正：参数类型改为 *ast.Node
func getTextFromNode(node *ast.Node) string {
	if node.Kind == ast.KindJsxExpression {
		if expr := node.AsJsxExpression(); expr.Expression != nil {
			return expr.Expression.Text()
		}
		return ""
	}
	return node.Text()
}
