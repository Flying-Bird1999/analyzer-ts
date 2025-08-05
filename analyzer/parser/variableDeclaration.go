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
				varType = utils.GetNodeText(variableDecl.Type, sourceCode)
			}
			var initValue string
			if initializerNode != nil {
				initValue = utils.GetNodeText(initializerNode, sourceCode)
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
				vd.Source = utils.GetNodeText(initializerNode, sourceCode)
			}
			arrayBinding := nameNode.AsBindingPattern()
			// 情况1: 初始化器是数组字面量表达式, 例如 [a, b] = [1, 2]
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
						// 从右侧数组字面量获取值
						initValue = utils.GetNodeText(arrayLiteral.Elements.Nodes[i], sourceCode)
					} else if bindingElement.Initializer != nil {
						// 获取默认值
						initValue = utils.GetNodeText(bindingElement.Initializer, sourceCode)
					}
					declarator := &VariableDeclarator{Identifier: identifier, InitValue: initValue}
					vd.Declarators = append(vd.Declarators, declarator)
				}
			} else {
				// 情况2: 初始化器不是数组字面量, 例如 [a, b] = someArray
				// 回退到原始行为：只捕获默认值
				for _, element := range arrayBinding.Elements.Nodes {
					if element.Kind != ast.KindBindingElement {
						continue
					}
					bindingElement := element.AsBindingElement()
					identifier := bindingElement.Name().AsIdentifier().Text
					var initValue string
					if bindingElement.Initializer != nil {
						initValue = utils.GetNodeText(bindingElement.Initializer, sourceCode)
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
				vd.Source = utils.GetNodeText(initializerNode, sourceCode)
			}
			objectBinding := nameNode.AsBindingPattern()
			// 情况1: 初始化器是对象字面量表达式, 例如 {a, b} = {a: 1, b: 2}
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
						propValue := utils.GetNodeText(propAssignment.Initializer, sourceCode)
						propertyValues[propName] = propValue
					} else if prop.Kind == ast.KindShorthandPropertyAssignment {
						shorthand := prop.AsShorthandPropertyAssignment()
						propName := shorthand.Name().Text()
						// 值就是名称本身（作为标识符）
						propertyValues[propName] = propName
					}
				}

				for _, element := range objectBinding.Elements.Nodes {
					if element.Kind != ast.KindBindingElement {
						continue
					}
					bindingElement := element.AsBindingElement()
					nameNode := bindingElement.Name()

					// 通过检查名称是标识符还是其他模式来处理嵌套解构
					if ast.IsIdentifier(nameNode) {
						identifier := nameNode.AsIdentifier().Text
						var lookupName string
						if bindingElement.PropertyName != nil {
							// 处理别名, 例如 { name: myName }
							propNameNode := bindingElement.PropertyName
							if ast.IsIdentifier(propNameNode) {
								lookupName = propNameNode.AsIdentifier().Text
							} else {
								lookupName = utils.GetNodeText(propNameNode, sourceCode)
							}
						} else {
							lookupName = identifier
						}

						initValue, ok := propertyValues[lookupName]
						if !ok && bindingElement.Initializer != nil {
							initValue = utils.GetNodeText(bindingElement.Initializer, sourceCode)
						}
						declarator := &VariableDeclarator{Identifier: identifier, PropName: lookupName, InitValue: initValue}
						vd.Declarators = append(vd.Declarators, declarator)
					} else if ast.IsObjectBindingPattern(nameNode) || ast.IsArrayBindingPattern(nameNode) {
						// 这是一个嵌套模式
						// 我们暂时将整个嵌套模式视为单个“标识符”
						// 更高级的实现会递归地解析它
						identifier := utils.GetNodeText(nameNode, sourceCode)
						var lookupName string
						if bindingElement.PropertyName != nil {
							propNameNode := bindingElement.PropertyName
							if ast.IsIdentifier(propNameNode) {
								lookupName = propNameNode.AsIdentifier().Text
							} else {
								lookupName = utils.GetNodeText(propNameNode, sourceCode)
							}
						} else {
							// 对于嵌套模式，这种情况在JS中应该是语法无效的，但我们进行防御性处理
							lookupName = identifier
						}

						initValue, ok := propertyValues[lookupName]
						if !ok && bindingElement.Initializer != nil {
							// 如果整个嵌套模式有默认值
							initValue = utils.GetNodeText(bindingElement.Initializer, sourceCode)
						}
						declarator := &VariableDeclarator{Identifier: identifier, PropName: lookupName, InitValue: initValue}
						vd.Declarators = append(vd.Declarators, declarator)
					}
				}
			} else {
				// 情况2: 初始化器不是对象字面量, 例如 {a, b} = someObject
				// 回退到原始行为：只捕获默认值
				for _, element := range objectBinding.Elements.Nodes {
					if element.Kind != ast.KindBindingElement {
						continue
					}
					bindingElement := element.AsBindingElement()
					nameNode := bindingElement.Name()

					// 通过检查名称是标识符还是其他模式来处理嵌套解构
					if ast.IsIdentifier(nameNode) {
						identifier := nameNode.AsIdentifier().Text
						var propName string
						if bindingElement.PropertyName != nil {
							// 处理别名, 例如 { name: myName }
							propNameNode := bindingElement.PropertyName
							if ast.IsIdentifier(propNameNode) {
								propName = propNameNode.AsIdentifier().Text
							} else {
								propName = utils.GetNodeText(propNameNode, sourceCode)
							}
						} else {
							propName = identifier
						}

						var initValue string
						if bindingElement.Initializer != nil {
							initValue = utils.GetNodeText(bindingElement.Initializer, sourceCode)
						}
						declarator := &VariableDeclarator{Identifier: identifier, PropName: propName, InitValue: initValue}
						vd.Declarators = append(vd.Declarators, declarator)
					} else if ast.IsObjectBindingPattern(nameNode) || ast.IsArrayBindingPattern(nameNode) {
						// 这是一个嵌套模式
						identifier := utils.GetNodeText(nameNode, sourceCode)
						var propName string
						if bindingElement.PropertyName != nil {
							propNameNode := bindingElement.PropertyName
							if ast.IsIdentifier(propNameNode) {
								propName = propNameNode.AsIdentifier().Text
							} else {
								propName = utils.GetNodeText(propNameNode, sourceCode)
							}
						} else {
							// 对于嵌套模式，这种情况在JS中应该是语法无效的，但我们进行防御性处理
							propName = identifier
						}

						var initValue string
						if bindingElement.Initializer != nil {
							// 如果整个嵌套模式有默认值
							initValue = utils.GetNodeText(bindingElement.Initializer, sourceCode)
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
