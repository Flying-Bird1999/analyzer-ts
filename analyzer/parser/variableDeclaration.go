package parser

import (
	"main/analyzer/utils"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// DeclarationKind 用于表示变量声明的类型 (const, let, var)
type DeclarationKind string

const (
	ConstDeclaration DeclarationKind = "const"
	LetDeclaration   DeclarationKind = "let"
	VarDeclaration   DeclarationKind = "var"
)

// VariableDeclarator 代表一个独立的变量声明
type VariableDeclarator struct {
	Identifier string `json:"identifier,omitempty"` // 标识符
	PropName   string `json:"propName,omitempty"`   // 属性名(别名)
	Type       string `json:"type,omitempty"`       // 类型
	InitValue  string `json:"initValue,omitempty"`  // 初始值
}

// NodePosition 用于记录代码中的位置信息
type NodePosition struct {
	Line   int `json:"line"`   // 行号
	Column int `json:"column"` // 列号
}

// SourceLocation 源码位置
type SourceLocation struct {
	Start NodePosition `json:"start"` // 节点起始位置
	End   NodePosition `json:"end"`   // 节点结束位置
}

// VariableDeclaration 代表一个完整的变量声明语句
type VariableDeclaration struct {
	Exported       bool                  `json:"exported"`         // 是否导出
	Kind           DeclarationKind       `json:"kind"`             // 声明类型
	Source         string                `json:"source,omitempty"` // 解构赋值的源，这里先记录下源码，因为可能为比较复杂的变量，比如 const { name } = user.profile； const { id } = getResponse().data;等
	Declarators    []*VariableDeclarator `json:"declarators"`      // 声明的变量
	Raw            string                `json:"raw,omitempty"`    // 源码
	SourceLocation SourceLocation        `json:"sourceLocation"`   // 源码位置
}

func NewVariableDeclaration(node *ast.VariableStatement, sourceCode string) *VariableDeclaration {
	pos, end := node.Pos(), node.End()
	return &VariableDeclaration{
		Declarators: make([]*VariableDeclarator, 0),
		SourceLocation: SourceLocation{
			Start: NodePosition{Line: pos, Column: 0},
			End:   NodePosition{Line: end, Column: 0},
		},
	}
}

// analyzeVariableDeclaration 解析变量声明语句
func (vd *VariableDeclaration) analyzeVariableDeclaration(node *ast.VariableStatement, sourceCode string, sourceFile *ast.SourceFile) {
	if node == nil {
		return
	}

	// 检查是否有 export 修饰符
	modifiers := node.Modifiers()
	if modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if modifier != nil && modifier.Kind == ast.KindExportKeyword {
				vd.Exported = true
				break
			}
		}
	}

	// 解析声明类型 (const, let, var)
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

	// 遍历所有声明
	for _, decl := range declarationList.AsVariableDeclarationList().Declarations.Nodes {
		if decl.Kind != ast.KindVariableDeclaration {
			continue
		}
		variableDecl := decl.AsVariableDeclaration()
		nameNode := variableDecl.Name()
		initializerNode := variableDecl.Initializer

		// 常规变量声明
		if ast.IsIdentifier(nameNode) {
			identifier := nameNode.AsIdentifier().Text
			var varType string
			if variableDecl.Type != nil {
				varType = utils.GetNodeText(variableDecl.Type.AsNode(), sourceCode)
			}
			var initValue string
			if initializerNode != nil {
				initValue = utils.GetNodeText(initializerNode.AsNode(), sourceCode)
			}
			declarator := &VariableDeclarator{
				Identifier: identifier,
				Type:       varType,
				InitValue:  initValue,
			}
			vd.Declarators = append(vd.Declarators, declarator)
			continue
		}

		// 数组解构
		if ast.IsArrayBindingPattern(nameNode) {
			if initializerNode != nil && initializerNode.Kind != ast.KindArrayLiteralExpression {
				vd.Source = utils.GetNodeText(initializerNode.AsNode(), sourceCode)
			}
			arrayBinding := nameNode.AsBindingPattern()
			// Case 1: Initializer is an ArrayLiteralExpression, e.g. [a, b] = [1, 2]
			if initializerNode != nil && initializerNode.Kind == ast.KindArrayLiteralExpression {
				arrayLiteral := initializerNode.AsArrayLiteralExpression()
				for i, element := range arrayBinding.Elements.Nodes {
					if element.Kind != ast.KindBindingElement {
						continue
					}
					bindingElement := element.AsBindingElement()
					identifier := bindingElement.Name().AsIdentifier().Text
					var initValue string
					if i < len(arrayLiteral.Elements.Nodes) && arrayLiteral.Elements.Nodes[i] != nil {
						// Get value from RHS array literal
						initValue = utils.GetNodeText(arrayLiteral.Elements.Nodes[i].AsNode(), sourceCode)
					} else if bindingElement.Initializer != nil {
						// Get default value
						initValue = utils.GetNodeText(bindingElement.Initializer.AsNode(), sourceCode)
					}
					declarator := &VariableDeclarator{Identifier: identifier, InitValue: initValue}
					vd.Declarators = append(vd.Declarators, declarator)
				}
			} else { // Case 2: Initializer is not an array literal, e.g. [a, b] = someArray
				// Fallback to original behavior: only capture default values
				for _, element := range arrayBinding.Elements.Nodes {
					if element.Kind != ast.KindBindingElement {
						continue
					}
					bindingElement := element.AsBindingElement()
					identifier := bindingElement.Name().AsIdentifier().Text
					var initValue string
					if bindingElement.Initializer != nil {
						initValue = utils.GetNodeText(bindingElement.Initializer.AsNode(), sourceCode)
					}
					declarator := &VariableDeclarator{Identifier: identifier, InitValue: initValue}
					vd.Declarators = append(vd.Declarators, declarator)
				}
			}
			continue
		}

		// 对象解构
		if ast.IsObjectBindingPattern(nameNode) {
			if initializerNode != nil && initializerNode.Kind != ast.KindObjectLiteralExpression {
				vd.Source = utils.GetNodeText(initializerNode.AsNode(), sourceCode)
			}
			objectBinding := nameNode.AsBindingPattern()
			// Case 1: Initializer is an ObjectLiteralExpression, e.g. {a, b} = {a: 1, b: 2}
			if initializerNode != nil && initializerNode.Kind == ast.KindObjectLiteralExpression {
				objectLiteral := initializerNode.AsObjectLiteralExpression()
				propertyValues := make(map[string]string)
				for _, prop := range objectLiteral.Properties.Nodes {
					if prop.Kind == ast.KindPropertyAssignment {
						propAssignment := prop.AsPropertyAssignment()
						name := propAssignment.Name()
						var propName string
						if ast.IsIdentifier(name) {
							propName = name.AsIdentifier().Text
						} else {
							propName = utils.GetNodeText(name, sourceCode)
						}
						propValue := utils.GetNodeText(propAssignment.Initializer.AsNode(), sourceCode)
						propertyValues[propName] = propValue
					} else if prop.Kind == ast.KindShorthandPropertyAssignment {
						shorthand := prop.AsShorthandPropertyAssignment()
						propName := shorthand.Name().Text()
						// The value is the name itself (as an identifier)
						propertyValues[propName] = propName
					}
				}

				for _, element := range objectBinding.Elements.Nodes {
					if element.Kind != ast.KindBindingElement {
						continue
					}
					bindingElement := element.AsBindingElement()
					nameNode := bindingElement.Name()

					// Handle nested destructuring by checking if the name is an identifier or another pattern
					if ast.IsIdentifier(nameNode) {
						identifier := nameNode.AsIdentifier().Text
						var lookupName string
						if bindingElement.PropertyName != nil {
							// Handle aliasing e.g. { name: myName }
							propNameNode := bindingElement.PropertyName
							if ast.IsIdentifier(propNameNode) {
								lookupName = propNameNode.AsIdentifier().Text
							} else {
								lookupName = utils.GetNodeText(propNameNode.AsNode(), sourceCode)
							}
						} else {
							lookupName = identifier
						}

						initValue, ok := propertyValues[lookupName]
						if !ok && bindingElement.Initializer != nil {
							initValue = utils.GetNodeText(bindingElement.Initializer.AsNode(), sourceCode)
						}
						declarator := &VariableDeclarator{Identifier: identifier, PropName: lookupName, InitValue: initValue}
						vd.Declarators = append(vd.Declarators, declarator)
					} else if ast.IsObjectBindingPattern(nameNode) || ast.IsArrayBindingPattern(nameNode) {
						// This is a nested pattern.
						// We will treat the entire nested pattern as a single "identifier" for now.
						// A more advanced implementation would recursively parse this.
						identifier := utils.GetNodeText(nameNode.AsNode(), sourceCode)
						var lookupName string
						if bindingElement.PropertyName != nil {
							propNameNode := bindingElement.PropertyName
							if ast.IsIdentifier(propNameNode) {
								lookupName = propNameNode.AsIdentifier().Text
							} else {
								lookupName = utils.GetNodeText(propNameNode.AsNode(), sourceCode)
							}
						} else {
							// This case should be syntactically invalid in JS for nested patterns, but we handle it defensively.
							lookupName = identifier
						}

						initValue, ok := propertyValues[lookupName]
						if !ok && bindingElement.Initializer != nil {
							// If there's a default value for the whole nested pattern
							initValue = utils.GetNodeText(bindingElement.Initializer.AsNode(), sourceCode)
						}
						declarator := &VariableDeclarator{Identifier: identifier, PropName: lookupName, InitValue: initValue}
						vd.Declarators = append(vd.Declarators, declarator)
					}
				}
			} else {
				// Case 2: Initializer is not an object literal, e.g. {a, b} = someObject
				// Fallback to original behavior: only capture default values
				for _, element := range objectBinding.Elements.Nodes {
					if element.Kind != ast.KindBindingElement {
						continue
					}
					bindingElement := element.AsBindingElement()
					nameNode := bindingElement.Name()

					// Handle nested destructuring by checking if the name is an identifier or another pattern
					if ast.IsIdentifier(nameNode) {
						identifier := nameNode.AsIdentifier().Text
						var propName string
						if bindingElement.PropertyName != nil {
							// Handle aliasing e.g. { name: myName }
							propNameNode := bindingElement.PropertyName
							if ast.IsIdentifier(propNameNode) {
								propName = propNameNode.AsIdentifier().Text
							} else {
								propName = utils.GetNodeText(propNameNode.AsNode(), sourceCode)
							}
						} else {
							propName = identifier
						}

						var initValue string
						if bindingElement.Initializer != nil {
							initValue = utils.GetNodeText(bindingElement.Initializer.AsNode(), sourceCode)
						}
						declarator := &VariableDeclarator{Identifier: identifier, PropName: propName, InitValue: initValue}
						vd.Declarators = append(vd.Declarators, declarator)
					} else if ast.IsObjectBindingPattern(nameNode) || ast.IsArrayBindingPattern(nameNode) {
						// This is a nested pattern.
						identifier := utils.GetNodeText(nameNode.AsNode(), sourceCode)
						var propName string
						if bindingElement.PropertyName != nil {
							propNameNode := bindingElement.PropertyName
							if ast.IsIdentifier(propNameNode) {
								propName = propNameNode.AsIdentifier().Text
							} else {
								propName = utils.GetNodeText(propNameNode.AsNode(), sourceCode)
							}
						} else {
							// This case should be syntactically invalid in JS for nested patterns, but we handle it defensively.
							propName = identifier
						}

						var initValue string
						if bindingElement.Initializer != nil {
							// If there's a default value for the whole nested pattern
							initValue = utils.GetNodeText(bindingElement.Initializer.AsNode(), sourceCode)
						}
						declarator := &VariableDeclarator{Identifier: identifier, PropName: propName, InitValue: initValue}
						vd.Declarators = append(vd.Declarators, declarator)
					}
				}
			}
			continue
		}
	}
	vd.Raw = utils.GetNodeText(node.AsNode(), sourceCode)
}
