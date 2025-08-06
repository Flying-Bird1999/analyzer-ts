// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（jsxDeclaration.go）专门负责处理和解析 JSX 相关的节点。
package parser

import (
	"main/analyzer/utils"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// JSXAttribute 表示 JSX 元素中的一个属性，例如在 <div className="App"> 中，className="App" 就是一个属性。
type JSXAttribute struct {
	Name     string `json:"name"`     // 属性名，例如 "className" 或在展开属性中为 "...props"。
	Value    string `json:"value"`    // 属性值，例如 "App"。对于布尔属性（如 `disabled`）或展开属性，此字段可能为空。
	IsSpread bool   `json:"isSpread"` // 标记该属性是否为展开属性（Spread Attribute），例如 <Component {...props} /> 中的 {...props}。
}

// JSXElement 代表一个解析后的 JSX 节点，包括其类型、标识符、属性和源码位置等信息。
type JSXElement struct {
	Type               string         `json:"type"`               // 节点类型，用于区分是简单标识符还是成员表达式。值为 "JSXIdentifier" 或 "JSXMemberExpression"。
	ModuleIdentifier   string         `json:"moduleIdentifier"`   // 模块标识符，仅在成员表达式（如 <Antd.Button />）中有效，此例中为 "Antd"。
	PropertyIdentifier string         `json:"propertyIdentifier"` // 属性或组件标识符。例如在 <Button /> 中为 "Button"，在 <Antd.Button /> 中也为 "Button"。
	Attrs              []JSXAttribute `json:"attrs"`              // JSX 属性列表，存储该元素的所有属性。
	Raw                string         `json:"raw"`                // 节点在源码中的原始文本。
	SourceLocation     SourceLocation `json:"sourceLocation"`     // 节点在源码中的精确位置（起始和结束的行列号）。
}

// NewJSXNode 是创建 JSXElement 实例的工厂函数。
// 它接收一个通用的 ast.Node 和源码字符串，首先提取节点的原始文本和源码位置。
// 然后根据节点的具体类型（ast.KindJsxElement 或 ast.KindJsxSelfClosingElement），
// 调用相应的分析函数来填充 JSXElement 的详细信息。
func NewJSXNode(node ast.Node, sourceCode string) *JSXElement {
	pos, end := node.Pos(), node.End()
	jsxNode := &JSXElement{
		Raw: utils.GetNodeText(node.AsNode(), sourceCode),
		SourceLocation: SourceLocation{
			Start: NodePosition{Line: pos, Column: 0},
			End:   NodePosition{Line: end, Column: 0},
		},
	}
	if node.Kind == ast.KindJsxElement {
		// 处理非自闭合的 JSX 元素, 例如 <div>...</div>
		jsxNode.analyzeJsxElement(node.AsJsxElement(), sourceCode)
	} else if node.Kind == ast.KindJsxSelfClosingElement {
		// 处理自闭合的 JSX 元素, 例如 <div />
		jsxNode.analyzeJsxSelfClosingElement(node.AsJsxSelfClosingElement(), sourceCode)
	}
	return jsxNode
}

// analyzeJsxElement 处理非自闭合的 JSX 元素（例如 <div>...</div>）。
// 它主要关注元素的开标签（OpeningElement），因此直接将分析任务委托给 analyzeJsxOpeningElement。
func (j *JSXElement) analyzeJsxElement(node *ast.JsxElement, sourceCode string) {
	j.analyzeJsxOpeningElement(node.OpeningElement, sourceCode)
}

// analyzeJsxOpeningElement 负责分析 JSX 开标签（<...>）或自闭合元素标签的内部结构。
// 1. 分析标签名（TagName）：
//    - 如果是简单标识符（ast.KindIdentifier），如 <Button>，则 Type 为 "JSXIdentifier"，PropertyIdentifier 为 "Button"。
//    - 如果是属性访问表达式（ast.KindPropertyAccessExpression），如 <Antd.Button>，则 Type 为 "JSXMemberExpression"，
//      ModuleIdentifier 为 "Antd"，PropertyIdentifier 为 "Button"。
// 2. 分析属性（Attributes）：
//    - 遍历所有属性节点。
//    - 如果是普通属性（ast.KindJsxAttribute），则提取其名称和值，并设置 IsSpread 为 false。
//    - 如果是展开属性（ast.KindJsxSpreadAttribute），如 {...props}，则提取表达式文本（如 "props"），
//      并将其作为 Name（前面加上 "..." 以便区分），同时设置 IsSpread 为 true。
func (j *JSXElement) analyzeJsxOpeningElement(node *ast.Node, sourceCode string) {
	openingElement := node.AsJsxOpeningElement()

	// 分析标签名
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

	// 分析属性列表
	if attributes := openingElement.Attributes; attributes != nil {
		if jsxAttrs := attributes.AsJsxAttributes(); jsxAttrs != nil && jsxAttrs.Properties != nil {
			for _, attr := range jsxAttrs.Properties.Nodes {
				if attr.Kind == ast.KindJsxAttribute {
					jsxAttr := attr.AsJsxAttribute()
					name := jsxAttr.Name().Text()
					var value string
					if jsxAttr.Initializer != nil {
						value = getTextFromNode(jsxAttr.Initializer)
					}
					j.Attrs = append(j.Attrs, JSXAttribute{Name: name, Value: value, IsSpread: false})
				} else if attr.Kind == ast.KindJsxSpreadAttribute {
					jsxSpreadAttr := attr.AsJsxSpreadAttribute()
					// 对于展开属性，名称记录为 `...` 加上表达式的文本
					name := "..." + jsxSpreadAttr.Expression.Text()
					j.Attrs = append(j.Attrs, JSXAttribute{Name: name, IsSpread: true})
				}
			}
		}
	}
}

// analyzeJsxSelfClosingElement 负责分析自闭合的 JSX 元素（例如 <Component />）。
// 其逻辑与 analyzeJsxOpeningElement 非常相似，同样需要分析标签名和属性列表。
func (j *JSXElement) analyzeJsxSelfClosingElement(node *ast.JsxSelfClosingElement, sourceCode string) {
	// 分析标签名
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

	// 分析属性列表
	if attributes := node.Attributes; attributes != nil {
		if jsxAttrs := attributes.AsJsxAttributes(); jsxAttrs != nil && jsxAttrs.Properties != nil {
			for _, attr := range jsxAttrs.Properties.Nodes {
				if attr.Kind == ast.KindJsxAttribute {
					jsxAttr := attr.AsJsxAttribute()
					name := jsxAttr.Name().Text()
					var value string
					if jsxAttr.Initializer != nil {
						value = getTextFromNode(jsxAttr.Initializer)
					}
					j.Attrs = append(j.Attrs, JSXAttribute{Name: name, Value: value, IsSpread: false})
				} else if attr.Kind == ast.KindJsxSpreadAttribute {
					jsxSpreadAttr := attr.AsJsxSpreadAttribute()
					// 对于展开属性，名称记录为 `...` 加上表达式的文本
					name := "..." + jsxSpreadAttr.Expression.Text()
					j.Attrs = append(j.Attrs, JSXAttribute{Name: name, IsSpread: true})
				}
			}
		}
	}
}

// getTextFromNode 是一个辅助函数，用于从 AST 节点中提取文本值。
// 它特别处理了 JSX 表达式（ast.KindJsxExpression），例如在 attr={value} 中，它会提取 `value` 的文本。
// 对于其他类型的节点（如字符串字面量），它直接返回节点的原始文本。
func getTextFromNode(node *ast.Node) string {
	if node.Kind == ast.KindJsxExpression {
		if expr := node.AsJsxExpression(); expr.Expression != nil {
			return expr.Expression.Text()
		}
		return ""
	}
	return node.Text()
}
