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
	Identifier string `json:"identifier,omitempty"`
	Type       string `json:"type,omitempty"`
	InitValue  string `json:"initValue,omitempty"`
}

// NodePosition 用于记录代码中的位置信息
type NodePosition struct {
	Line   int // 行号
	Column int // 列号
}

type SourceLocation struct {
	Start NodePosition // 节点起始位置
	End   NodePosition // 节点结束位置
}

// VariableDeclaration 代表一个完整的变量声明语句
type VariableDeclaration struct {
	Exported       bool                  `json:"exported"`
	Kind           DeclarationKind       `json:"kind"`
	Declarators    []*VariableDeclarator `json:"declarators"`
	Raw            string                `json:"raw,omitempty"`
	SourceLocation SourceLocation        `json:"sourceLocation"`
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
		if decl.Kind == ast.KindVariableDeclaration {
			variableDecl := decl.AsVariableDeclaration()

			// 获取变量名
			nameNode := variableDecl.Name()
			if ast.IsArrayBindingPattern(nameNode) {
				// 处理数组解构
				for _, element := range nameNode.AsBindingPattern().Elements.Nodes {
					if element.Kind == ast.KindBindingElement {
						bindingElement := element.AsBindingElement()
						identifier := bindingElement.Name().AsIdentifier().Text

						// 获取初始化表达式
						var initValue string
						if bindingElement.Initializer != nil {
							initValue = utils.GetNodeText(bindingElement.Initializer.AsNode(), sourceCode)
						}

						declarator := &VariableDeclarator{
							Identifier: identifier,
							InitValue:  initValue,
						}
						vd.Declarators = append(vd.Declarators, declarator)
					}
				}
			} else if ast.IsObjectBindingPattern(nameNode) {
				// 处理对象解构
				for _, element := range nameNode.AsBindingPattern().Elements.Nodes {
					if element.Kind == ast.KindBindingElement {
						bindingElement := element.AsBindingElement()
						identifier := bindingElement.Name().AsIdentifier().Text

						// 获取初始化表达式
						var initValue string
						if bindingElement.Initializer != nil {
							initValue = utils.GetNodeText(bindingElement.Initializer.AsNode(), sourceCode)
						}

						declarator := &VariableDeclarator{
							Identifier: identifier,
							InitValue:  initValue,
						}
						vd.Declarators = append(vd.Declarators, declarator)
					}
				}
			} else if ast.IsIdentifier(nameNode) {
				identifier := nameNode.AsIdentifier().Text

				// 获取类型
				var varType string
				if variableDecl.Type != nil {
					varType = utils.GetNodeText(variableDecl.Type.AsNode(), sourceCode)
				}

				// 获取初始化表达式
				var initValue string
				if variableDecl.Initializer != nil {
					initValue = utils.GetNodeText(variableDecl.Initializer.AsNode(), sourceCode)
				}

				declarator := &VariableDeclarator{
					Identifier: identifier,
					Type:       varType,
					InitValue:  initValue,
				}
				vd.Declarators = append(vd.Declarators, declarator)
			}
		}
	}
	vd.Raw = utils.GetNodeText(node.AsNode(), sourceCode)
}
