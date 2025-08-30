// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（variableDeclaration.go）专门负责处理和解析变量声明。
package parser

import (
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"

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
	return &VariableDeclaration{
		Declarators:    make([]*VariableDeclarator, 0),
		Raw:            utils.GetNodeText(node.AsNode(), sourceCode),
		SourceLocation: NewSourceLocation(node.AsNode(), sourceCode),
	}
}

// AnalyzeVariableValueNode 是一个核心辅助函数，用于从 AST 节点中解析出结构化的值信息。
func AnalyzeVariableValueNode(node *ast.Node, sourceCode string) *VariableValue {
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

// ExtractVariableDeclarations 从一个变量声明语句中提取出所有的声明，并作为单独的 VariableDeclaration 对象返回。
// 一个语句（如 `export const a = 1, b = 2`）可能包含多个声明，此函数将其拆分。
func ExtractVariableDeclarations(node *ast.VariableStatement, sourceCode string) []VariableDeclaration {
	results := []VariableDeclaration{}

	isExported := false
	if modifiers := node.Modifiers(); modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if modifier != nil && modifier.Kind == ast.KindExportKeyword {
				isExported = true
				break
			}
		}
	}

	declarationList := node.DeclarationList
	if declarationList == nil {
		return results
	}

	// 遍历声明列表中的每一个声明 (例如 `const a = 1, b = 2`)
	for _, decl := range declarationList.AsVariableDeclarationList().Declarations.Nodes {
		variableDecl := decl.AsVariableDeclaration()
		if variableDecl == nil {
			continue
		}
		// --- 常规变量和解构变量处理---
		vd := NewVariableDeclaration(node, sourceCode)
		vd.Exported = isExported
		if (declarationList.Flags & ast.NodeFlagsConst) != 0 {
			vd.Kind = ConstDeclaration
		} else if (declarationList.Flags & ast.NodeFlagsLet) != 0 {
			vd.Kind = LetDeclaration
		} else {
			vd.Kind = VarDeclaration
		}

		nameNode := variableDecl.Name()
		initializerNode := variableDecl.Initializer

		if ast.IsIdentifier(nameNode) {
			declarator := &VariableDeclarator{
				Identifier: nameNode.AsIdentifier().Text,
				Type:       AnalyzeVariableValueNode(variableDecl.Type, sourceCode),
				InitValue:  AnalyzeVariableValueNode(initializerNode, sourceCode),
			}
			vd.Declarators = append(vd.Declarators, declarator)
		} else if ast.IsObjectBindingPattern(nameNode) || ast.IsArrayBindingPattern(nameNode) {
			vd.Source = AnalyzeVariableValueNode(initializerNode, sourceCode)
			analyzeBindingPattern(nameNode, vd, sourceCode)
		}
		results = append(results, *vd)
	}
	return results
}

// VisitVariableStatement 解析变量声明语句。
// 它现在将主要工作委托给 ExtractVariableDeclarations，并处理函数赋值和动态导入等特殊情况。
func (p *Parser) VisitVariableStatement(node *ast.VariableStatement) {
	isExported := false
	if modifiers := node.Modifiers(); modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if modifier != nil && modifier.Kind == ast.KindExportKeyword {
				isExported = true
				break
			}
		}
	}

	// 遍历所有声明，检查是否是函数赋值或动态导入
	for _, decl := range node.DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
		variableDecl := decl.AsVariableDeclaration()
		if p.analyzeFunctionAssignment(variableDecl, isExported) {
			continue
		}
		if p.analyzeDynamicImportAssignment(variableDecl) {
			continue
		}
	}

	// 将常规变量声明添加到结果中
	decls := ExtractVariableDeclarations(node, p.SourceCode)
	p.Result.VariableDeclarations = append(p.Result.VariableDeclarations, decls...)
}

// analyzeFunctionAssignment 专门处理赋值为函数表达式的变量声明。
// 如果成功解析了一个函数表达式，则返回 true。
func (p *Parser) analyzeFunctionAssignment(variableDecl *ast.VariableDeclaration, isExported bool) bool {
	nameNode := variableDecl.Name()
	initializerNode := variableDecl.Initializer

	if ast.IsIdentifier(nameNode) && initializerNode != nil {
		identifier := nameNode.AsIdentifier().Text
		initKind := initializerNode.Kind

		if initKind == ast.KindArrowFunction || initKind == ast.KindFunctionExpression {
			fr := NewFunctionDeclarationResultFromExpression(identifier, isExported, initializerNode, p.SourceCode)
			p.Result.FunctionDeclarations = append(p.Result.FunctionDeclarations, *fr)
			return true
		}
	}
	return false
}

// analyzeDynamicImportAssignment 专门处理赋值为动态导入的变量声明。
// 如果成功解析了一个动态导入，则返回 true。
func (p *Parser) analyzeDynamicImportAssignment(variableDecl *ast.VariableDeclaration) bool {
	nameNode := variableDecl.Name()
	initializerNode := variableDecl.Initializer

	if ast.IsIdentifier(nameNode) && initializerNode != nil {
		identifier := nameNode.AsIdentifier().Text
		importCallNode, importPath := p.findDynamicImport(initializerNode)

		if importCallNode != nil && importPath != "" {
			importResult := &ImportDeclarationResult{
				Source: importPath,
				ImportModules: []ImportModule{
					{
						Identifier:   identifier,
						ImportModule: "default",
						Type:         "dynamic_variable",
					},
				},
				Raw: utils.GetNodeText(importCallNode, p.SourceCode),
			}
			p.Result.ImportDeclarations = append(p.Result.ImportDeclarations, *importResult)
			p.ProcessedDynamicImports[importCallNode] = true
			return true
		}
	}
	return false
}

// findDynamicImport 递归地在给定的 AST 节点中查找第一个 `import()` 调用。
// 它会深入常见的包装函数（如 `lazy`, `() => ...`）内部进行查找。
// 返回找到的 `import()` 对应的 ast.Node 和导入的路径字符串。
func (p *Parser) findDynamicImport(node *ast.Node) (*ast.Node, string) {
	if node == nil {
		return nil, ""
	}
	// 基本情况：当前节点就是 `import()` 调用

	if node.Kind == ast.KindCallExpression {
		callExpr := node.AsCallExpression()
		if callExpr.Expression.Kind == ast.KindImportKeyword {
			if len(callExpr.Arguments.Nodes) > 0 {
				arg := callExpr.Arguments.Nodes[0]
				if arg.Kind == ast.KindStringLiteral {
					return node, arg.AsStringLiteral().Text
				} else if arg.Kind == ast.KindIdentifier {
					return node, arg.AsIdentifier().Text
				}
			}
		}
	}

	// 递归情况：遍历子节点查找。
	var foundNode *ast.Node
	var foundPath string
	node.ForEachChild(func(child *ast.Node) bool {
		// 如果已经找到了，就停止遍历，防止找到更深层的无关 `import`。
		if foundNode != nil {
			return true // stop traversal
		}
		foundNode, foundPath = p.findDynamicImport(child)
		return foundNode != nil // 如果在子节点中找到了，返回 true 停止遍历
	})

	return foundNode, foundPath
}

// analyzeBindingPattern 解析解构模式。
func analyzeBindingPattern(node *ast.Node, vd *VariableDeclaration, sourceCode string) {
	if node == nil {
		return
	}

	bindingPattern := node.AsBindingPattern()
	if bindingPattern == nil || bindingPattern.Elements == nil {
		return
	}
	elements := bindingPattern.Elements.Nodes

	for _, element := range elements {
		bindingElement := element.AsBindingElement()
		if bindingElement == nil {
			continue
		}

		nameNode := bindingElement.Name()
		if nameNode == nil {
			continue
		}

		if ast.IsObjectBindingPattern(nameNode) || ast.IsArrayBindingPattern(nameNode) {
			analyzeBindingPattern(nameNode, vd, sourceCode)
		} else if ast.IsIdentifier(nameNode) {
			identifier := nameNode.AsIdentifier().Text
			propName := identifier

			if propertyNameNode := bindingElement.PropertyName; propertyNameNode != nil {
				switch propertyNameNode.Kind {
				case ast.KindIdentifier:
					if propIdentifier := propertyNameNode.AsIdentifier(); propIdentifier != nil {
						propName = propIdentifier.Text
					}
				case ast.KindStringLiteral:
					if strLit := propertyNameNode.AsStringLiteral(); strLit != nil {
						propName = strLit.Text
					}
				case ast.KindNumericLiteral:
					if numLit := propertyNameNode.AsNumericLiteral(); numLit != nil {
						propName = numLit.Text
					}
				case ast.KindComputedPropertyName:
					propName = strings.TrimSpace(utils.GetNodeText(propertyNameNode.AsNode(), sourceCode))
				default:
					propName = strings.TrimSpace(utils.GetNodeText(propertyNameNode.AsNode(), sourceCode))
				}
			}

			declarator := &VariableDeclarator{
				Identifier: identifier,
				PropName:   propName,
				InitValue:  AnalyzeVariableValueNode(bindingElement.Initializer, sourceCode),
			}
			vd.Declarators = append(vd.Declarators, declarator)
		}
	}
}
