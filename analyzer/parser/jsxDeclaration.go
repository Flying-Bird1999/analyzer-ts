// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（jsxDeclaration.go）专门负责处理和解析 JSX 相关的节点，
// 例如 <div className="App"></div> 或 <Component {...props} /> 等。
package parser

import (
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// JSXAttributeValue 用于结构化地表示 JSX 属性的值。
// 由于属性值可以是字符串、变量、函数调用或更复杂的表达式，
// 使用此结构体可以更精确地描述值的类型和内容。
type JSXAttributeValue struct {
	// Type 字段用于标识属性值的具体类型。
	// 例如："stringLiteral", "identifier", "propertyAccess", "callExpression" 等。
	Type string `json:"type"`

	// Expression 字段存储了属性值在源码中的原始文本，主要用于展示或简单分析。
	Expression string `json:"expression"`

	// Data 字段用于存储解析后的结构化数据，提供了比原始文本更丰富的信息。
	// 例如，如果 Type 是 "stringLiteral"，Data 会存储不带引号的字符串值。
	// 对于复杂的表达式，未来可以扩展此字段以存储更详细的结构。
	Data interface{} `json:"data,omitempty"`
}

// JSXAttribute 表示 JSX 元素中的一个属性。
// 它可以是一个常规的键值对属性（如 `className="App"`），
// 一个布尔属性（如 `disabled`），或一个展开属性（如 `{...props}`）。
type JSXAttribute struct {
	// Name 是属性的名称。对于展开属性，名称会以 "..." 开头，后跟展开的表达式文本。
	Name string `json:"name"`

	// Value 是一个指向 JSXAttributeValue 的指针，用于存储属性值的结构化信息。
	// 对于布尔属性（例如 `<Button disabled />`）或展开属性，此字段为 nil。
	Value *JSXAttributeValue `json:"value"`

	// IsSpread 标记该属性是否为展开属性（Spread Attribute）。
	IsSpread bool `json:"isSpread"`
}

// JSXElement 代表一个解析后的 JSX 节点。
type JSXElement struct {
	// ComponentChain 表示组件的完整路径。
	// 例如 <myComponent.Name.SingleSelect /> 会被解析为 ["myComponent", "Name", "SingleSelect"]。
	// 对于简单标签如 <div />，则为 ["div"]。
	ComponentChain []string       `json:"componentChain"`
	Attrs          []JSXAttribute `json:"attrs"`
	Raw            string         `json:"raw,omitempty"`            // 节点在源码中的原始文本
	SourceLocation *SourceLocation `json:"sourceLocation,omitempty"` // 节点在源码中的位置信息
}

// NewJSXNode 是创建和解析 JSXElement 实例的工厂函数。
// 它接收一个 AST 节点和文件源码作为输入，根据节点的具体类型
// （自闭合标签或非自闭合标签）分发给相应的解析函数，并返回一个填充好信息的 JSXElement 实例。
func NewJSXNode(node *ast.Node, sourceCode string) *JSXElement {
	return &JSXElement{
		Raw:            utils.GetNodeText(node, sourceCode),
		SourceLocation: NewSourceLocation(node, sourceCode),
	}
}

// analyzeAttributeValue 从属性的 AST 节点中解析出其结构化信息。
func analyzeAttributeValue(node *ast.Node, sourceCode string) *JSXAttributeValue {
	if node == nil {
		return nil
	}

	value := &JSXAttributeValue{
		Expression: utils.GetNodeText(node, sourceCode),
	}

	actualValueNode := node
	if node.Kind == ast.KindJsxExpression {
		if expr := node.AsJsxExpression(); expr.Expression != nil {
			actualValueNode = expr.Expression
		}
	}

	switch actualValueNode.Kind {
	case ast.KindStringLiteral:
		value.Type = "stringLiteral"
		value.Data = actualValueNode.AsStringLiteral().Text
	case ast.KindIdentifier:
		value.Type = "identifier"
		value.Data = actualValueNode.AsIdentifier().Text
	case ast.KindPropertyAccessExpression:
		value.Type = "propertyAccess"
	case ast.KindCallExpression:
		value.Type = "callExpression"
	case ast.KindArrowFunction:
		value.Type = "arrowFunction"
	case ast.KindTemplateExpression:
		value.Type = "templateExpression"
	case ast.KindNumericLiteral:
		value.Type = "numericLiteral"
	case ast.KindTrueKeyword, ast.KindFalseKeyword:
		value.Type = "booleanLiteral"
	default:
		value.Type = "other"
	}

	return value
}

// ReconstructJSXName 从 JSX 标签名节点递归地构建一个组件调用链。
// 此函数被设为公共，以便在 analyzer_tree 包中复用。
func ReconstructJSXName(node *ast.Node) []string {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return []string{node.AsIdentifier().Text}
	case ast.KindPropertyAccessExpression:
		propAccess := node.AsPropertyAccessExpression()
		left := ReconstructJSXName(propAccess.Expression)
		return append(left, propAccess.Name().Text())
	default:
		return []string{}
	}
}

// analyzeJsxAttributes 是一个辅助函数，用于解析 JSX 元素的属性。
func analyzeJsxAttributes(jsxNode *JSXElement, attributesNode *ast.Node, sourceCode string) {
	if attributesNode == nil {
		return
	}
	if jsxAttrs := attributesNode.AsJsxAttributes(); jsxAttrs != nil && jsxAttrs.Properties != nil {
		for _, attr := range jsxAttrs.Properties.Nodes {
			if attr.Kind == ast.KindJsxAttribute {
				jsxAttr := attr.AsJsxAttribute()
				jsxNode.Attrs = append(jsxNode.Attrs, JSXAttribute{
					Name:     jsxAttr.Name().Text(),
					Value:    analyzeAttributeValue(jsxAttr.Initializer, sourceCode),
					IsSpread: false,
				})
			} else if attr.Kind == ast.KindJsxSpreadAttribute {
				jsxSpreadAttr := attr.AsJsxSpreadAttribute()
				jsxNode.Attrs = append(jsxNode.Attrs, JSXAttribute{
					Name:     "..." + utils.GetNodeText(jsxSpreadAttr.Expression, sourceCode),
					Value:    nil,
					IsSpread: true,
				})
			}
		}
	}
}

// AnalyzeJsxElement 是一个公共的、可复用的函数，用于从 AST 节点中解析 JSX 元素。
func AnalyzeJsxElement(node *ast.Node, sourceCode string) *JSXElement {
	jsxNode := NewJSXNode(node, sourceCode)
	if node.Kind == ast.KindJsxElement {
		openingElement := node.AsJsxElement().OpeningElement.AsJsxOpeningElement()
		jsxNode.ComponentChain = ReconstructJSXName(openingElement.TagName)
		analyzeJsxAttributes(jsxNode, openingElement.Attributes, sourceCode)
	} else if node.Kind == ast.KindJsxSelfClosingElement {
		selfClosingElement := node.AsJsxSelfClosingElement()
		jsxNode.ComponentChain = ReconstructJSXName(selfClosingElement.TagName)
		analyzeJsxAttributes(jsxNode, selfClosingElement.Attributes, sourceCode)
	}
	return jsxNode
}

// VisitJsxElement 解析 JSX 元素（包括自闭合和非自闭合的）。
func (p *Parser) VisitJsxElement(node *ast.Node) {
	result := AnalyzeJsxElement(node, p.SourceCode)
	p.Result.JsxElements = append(p.Result.JsxElements, *result)
}