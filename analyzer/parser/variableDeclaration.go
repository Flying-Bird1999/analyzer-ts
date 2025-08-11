// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（variableDeclaration.go）专门负责处理和解析变量声明。
package parser

import (
	"main/analyzer/utils"
	"strings"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// DeclarationKind 用于表示变量声明的类型 (const, let, var)。
type DeclarationKind string

const (
	ConstDeclaration DeclarationKind = "const"
	LetDeclaration   DeclarationKind = "let"
	VarDeclaration   DeclarationKind = "var"
)

// VariableValue 用于结构化地表示变量的类型、初始值或解构源。
// 它取代了之前使用简单字符串的方式，提供了更丰富、更精确的 AST 信息。
type VariableValue struct {
	// Type 字段用于标识值的具体类型，例如 "stringLiteral", "identifier", "callExpression", "objectLiteral" 等。
	Type string `json:"type"`

	// Expression 字段存储了节点在源码中的原始文本，主要用于展示或简单分析。
	Expression string `json:"expression"`

	// Data 字段用于存储解析后的结构化数据，提供了比原始文本更丰富的信息。
	Data interface{} `json:"data,omitempty"`
}

// VariableDeclarator 代表一个独立的变量声明器。
// 在 `const a = 1, b = 2` 中，`a = 1` 和 `b = 2` 分别是两个声明器。
type VariableDeclarator struct {
	// Identifier 是声明的变量名（在解构中是绑定的本地变量名）。
	Identifier string `json:"identifier,omitempty"`

	// PropName 是解构赋值时的属性名（源属性名）。如果存在别名，Identifier 是别名，PropName 是原名。
	// 例如 `const { name: myName } = user` 中，Identifier 是 `myName`，PropName 是 `name`。
	// 如果没有别名，则与 Identifier 相同。
	PropName string `json:"propName,omitempty"`

	// Type 是变量的类型注解的结构化表示。
	Type *VariableValue `json:"type,omitempty"`

	// InitValue 是变量初始值的结构化表示。
	InitValue *VariableValue `json:"initValue,omitempty"`
}

// VariableDeclaration 代表一个完整的变量声明语句，例如 `export const a = 1;`。
type VariableDeclaration struct {
	// Exported 标记此变量声明是否被导出。
	Exported bool `json:"exported"`

	// Kind 表示声明的类型 (const, let, var)。
	Kind DeclarationKind `json:"kind"`

	// Source 是解构赋值的源的结构化表示。
	// 例如 `const { name } = user` 中，Source 代表 `user`。
	Source *VariableValue `json:"source,omitempty"`

	// Declarators 包含此语句中所有的变量声明器。
	Declarators []*VariableDeclarator `json:"declarators"`

	// Raw 存储了该节点在源码中的原始文本。
	Raw string `json:"raw,omitempty"`

	// SourceLocation 记录了该节点在源码中的精确位置。
	SourceLocation SourceLocation `json:"sourceLocation"`
}

// NewVariableDeclaration 是创建和解析 VariableDeclaration 实例的工厂函数。
func NewVariableDeclaration(node *ast.VariableStatement, sourceCode string) *VariableDeclaration {
	pos, end := node.Pos(), node.End()
	vd := &VariableDeclaration{
		Declarators: make([]*VariableDeclarator, 0),
		Raw:         utils.GetNodeText(node.AsNode(), sourceCode),
		SourceLocation: SourceLocation{
			Start: NodePosition{Line: pos, Column: 0},
			End:   NodePosition{Line: end, Column: 0},
		},
	}
	vd.AnalyzeVariableDeclaration(node, sourceCode)
	return vd
}

// analyzeVariableValueNode 是一个核心辅助函数，用于从 AST 节点中解析出结构化的值信息。
func analyzeVariableValueNode(node *ast.Node, sourceCode string) *VariableValue {
	if node == nil {
		return nil
	}

	value := &VariableValue{
		Expression: strings.TrimSpace(utils.GetNodeText(node.AsNode(), sourceCode)),
	}

	switch node.Kind {
	case ast.KindStringLiteral:
		value.Type = "stringLiteral"
		value.Data = node.AsStringLiteral().Text
	case ast.KindNumericLiteral:
		value.Type = "numericLiteral"
		value.Data = node.AsNumericLiteral().Text
	case ast.KindIdentifier:
		value.Type = "identifier"
		value.Data = node.AsIdentifier().Text
	case ast.KindPropertyAccessExpression:
		value.Type = "propertyAccess"
	case ast.KindCallExpression:
		value.Type = "callExpression"
	case ast.KindArrowFunction:
		value.Type = "arrowFunction"
	case ast.KindObjectLiteralExpression:
		value.Type = "objectLiteral"
	case ast.KindArrayLiteralExpression:
		value.Type = "arrayLiteral"
	case ast.KindNewExpression:
		value.Type = "newExpression"
	case ast.KindTrueKeyword, ast.KindFalseKeyword:
		value.Type = "booleanLiteral"
	default:
		if ast.IsTypeNode(node) {
			value.Type = "typeNode"
		} else {
			value.Type = "other"
		}
	}

	return value
}

// analyzeBindingPattern 是一个递归函数，用于解析（可能嵌套的）解构模式。
// 它会将解析出的所有变量声明器追加到 vd.Declarators 中。
func (vd *VariableDeclaration) analyzeBindingPattern(node *ast.Node, sourceCode string) {
	if node == nil {
		return
	}

	// 使用通用的 AsBindingPattern 获取元素列表
	bindingPattern := node.AsBindingPattern()
	if bindingPattern == nil || bindingPattern.Elements == nil {
		return
	}
	elements := bindingPattern.Elements.Nodes

	// 遍历所有解构元素
	for _, element := range elements {
		bindingElement := element.AsBindingElement()
		if bindingElement == nil {
			continue
		}

		nameNode := bindingElement.Name()
		if nameNode == nil {
			continue
		}

		// 如果解构的元素本身又是一个解构模式（嵌套解构），则递归调用
		if ast.IsObjectBindingPattern(nameNode) || ast.IsArrayBindingPattern(nameNode) {
			vd.analyzeBindingPattern(nameNode, sourceCode)
		} else if ast.IsIdentifier(nameNode) {
			// 如果是最终的标识符，则创建并添加声明器
			identifier := nameNode.AsIdentifier().Text
			propName := identifier // 默认情况下，属性名和标识符相同

			// 处理别名: `const { name: myName } = user`
			if propertyNameNode := bindingElement.PropertyName; propertyNameNode != nil {
				if propIdentifier := propertyNameNode.AsIdentifier(); propIdentifier != nil {
					propName = propIdentifier.Text
				}
			}

			declarator := &VariableDeclarator{
				Identifier: identifier,
				PropName:   propName,
				// 处理默认值: `const { name = "guest" } = user`
				InitValue: analyzeVariableValueNode(bindingElement.Initializer, sourceCode),
			}
			vd.Declarators = append(vd.Declarators, declarator)
		}
	}
}

// analyzeVariableDeclaration 是解析变量声明语句的核心逻辑。
func (vd *VariableDeclaration) AnalyzeVariableDeclaration(node *ast.VariableStatement, sourceCode string) {
	// 1. 检查导出关键字
	if modifiers := node.Modifiers(); modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if modifier != nil && modifier.Kind == ast.KindExportKeyword {
				vd.Exported = true
				break
			}
		}
	}

	// 2. 解析声明类型 (const, let, var)
	declarationList := node.DeclarationList
	if declarationList == nil {
		return
	}
	if (declarationList.Flags & ast.NodeFlagsConst) != 0 {
		vd.Kind = ConstDeclaration
	} else if (declarationList.Flags & ast.NodeFlagsLet) != 0 {
		vd.Kind = LetDeclaration
	} else {
		vd.Kind = VarDeclaration
	}

	// 3. 遍历所有声明器
	for _, decl := range declarationList.AsVariableDeclarationList().Declarations.Nodes {
		variableDecl := decl.AsVariableDeclaration()
		if variableDecl == nil {
			continue
		}

		nameNode := variableDecl.Name()
		initializerNode := variableDecl.Initializer

		// 3.1. 处理常规变量声明 (e.g., `const a = 1`)
		if ast.IsIdentifier(nameNode) {
			declarator := &VariableDeclarator{
				Identifier: nameNode.AsIdentifier().Text,
				Type:       analyzeVariableValueNode(variableDecl.Type, sourceCode),
				InitValue:  analyzeVariableValueNode(initializerNode, sourceCode),
			}
			vd.Declarators = append(vd.Declarators, declarator)
			continue
		}

		// 3.2. 处理解构声明 (e.g., `const {a, b: c} = {a: 1, b: 2}`)
		if ast.IsObjectBindingPattern(nameNode) || ast.IsArrayBindingPattern(nameNode) {
			// 解析解构的源
			vd.Source = analyzeVariableValueNode(initializerNode, sourceCode)
			// 使用新的递归函数来解析解构模式
			vd.analyzeBindingPattern(nameNode, sourceCode)
		}
	}
}
