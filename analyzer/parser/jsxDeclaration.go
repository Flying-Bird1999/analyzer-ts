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
	Raw            string         `json:"raw"`
	SourceLocation SourceLocation `json:"sourceLocation"`
}

// NewJSXNode 是创建和解析 JSXElement 实例的工厂函数。
// 它接收一个 AST 节点和文件源码作为输入，根据节点的具体类型
// （自闭合标签或非自闭合标签）分发给相应的解析函数，并返回一个填充好信息的 JSXElement 实例。
func NewJSXNode(node ast.Node, sourceCode string) *JSXElement {
	pos, end := node.Pos(), node.End()
	return &JSXElement{
		Raw: utils.GetNodeText(node.AsNode(), sourceCode),
		SourceLocation: SourceLocation{
			Start: NodePosition{Line: pos, Column: 0},
			End:   NodePosition{Line: end, Column: 0},
		},
	}
}

// analyzeAttributeValue 从属性的 AST 节点中解析出其结构化信息。
// 这是实现“从 AST 直接采集数据”的核心函数，取代了旧的、依赖源码字符串的方法。
// node: 属性的初始值节点 (Initializer)。
// sourceCode: 完整的文件源码，用于获取节点原始文本。
// 返回一个填充好的 JSXAttributeValue 实例。
func analyzeAttributeValue(node *ast.Node, sourceCode string) *JSXAttributeValue {
	// 对于布尔属性，其 Initializer 节点为 nil。
	if node == nil {
		return nil
	}

	value := &JSXAttributeValue{
		// 首先，无论值的类型是什么，都记录其原始文本表达式。
		Expression: utils.GetNodeText(node.AsNode(), sourceCode),
	}

	// 属性值通常被包裹在 JsxExpression 中（例如 `attr={...}`），
	// 我们需要解开这层包裹，获取真正的表达式节点。
	actualValueNode := node
	if node.Kind == ast.KindJsxExpression {
		if expr := node.AsJsxExpression(); expr.Expression != nil {
			actualValueNode = expr.Expression
		}
	}

	// 根据真实值节点的类型，填充 Type 和 Data 字段。
	switch actualValueNode.Kind {
	case ast.KindStringLiteral:
		value.Type = "stringLiteral"
		value.Data = actualValueNode.AsStringLiteral().Text // 存储不带引号的纯字符串
	case ast.KindIdentifier:
		value.Type = "identifier"
		value.Data = actualValueNode.AsIdentifier().Text // 存储变量名
	case ast.KindPropertyAccessExpression:
		value.Type = "propertyAccess"
		// 未来可以扩展，例如将 `styles.foo` 解析为 { object: "styles", property: "foo" } 并存入 Data 字段。
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
		value.Type = "other" // 其他未覆盖的复杂类型
	}

	return value
}

// reconstructJSXName 从 JSX 标签名节点递归地构建一个组件调用链。
func reconstructJSXName(node *ast.Node) []string {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return []string{node.AsIdentifier().Text}
	case ast.KindPropertyAccessExpression:
		propAccess := node.AsPropertyAccessExpression()
		// 递归地构建左侧部分的调用链
		left := reconstructJSXName(propAccess.Expression)
		// 将右侧的名称追加到链上
		return append(left, propAccess.Name().Text())
	default:
		// 对于非预期的节点类型，返回一个空切片
		return []string{}
	}
}
