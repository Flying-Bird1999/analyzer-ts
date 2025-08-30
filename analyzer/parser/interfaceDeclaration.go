// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（interfaceDeclaration.go）专门负责处理和解析接口（Interface）声明。
package parser

import (
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// TypeReference 代表一个类型引用。
// 它记录了在接口或类型别名中引用的其他类型（非基础类型）的信息。
type TypeReference struct {
	Identifier string   `json:"identifier"` // 被引用的类型名称，例如 `User`。
	Location   []string `json:"location"`   // 类型被引用的所有具体位置路径的列表，例如 `["School.student.name"]`。
	IsExtend   bool     `json:"isExtend"`   // 标记此引用是否来自 `extends` 子句。
}

// InterfaceDeclarationResult 存储一个完整的接口声明的解析结果。
type InterfaceDeclarationResult struct {
	Identifier     string                   `json:"identifier"`     // 接口的名称。
	Exported       bool                     `json:"exported"`       // 新增：标记此接口是否被导出。
	Raw            string                   `json:"raw"`            // 节点在源码中的原始文本。
	Reference      map[string]TypeReference `json:"reference"`      // 接口所依赖的其他类型的映射，以类型名作为 key。
	SourceLocation SourceLocation           `json:"sourceLocation"` // 节点在源码中的位置信息。
}

// NewInterfaceDeclarationResult 基于 AST 节点创建一个新的 InterfaceDeclarationResult 实例。
func NewInterfaceDeclarationResult(node *ast.Node, sourceCode string) *InterfaceDeclarationResult {
	raw := utils.GetNodeText(node, sourceCode)
	pos, end := node.Pos(), node.End()

	return &InterfaceDeclarationResult{
		Identifier: "",
		Exported:   false, // 默认为 false
		Raw:        raw,
		Reference:  make(map[string]TypeReference),
		SourceLocation: SourceLocation{
			Start: NodePosition{Line: pos, Column: 0},
			End:   NodePosition{Line: end, Column: 0},
		},
	}
}

// addTypeReference 是一个辅助函数，用于向结果的 Reference 映射中添加或更新类型引用。
func (inter *InterfaceDeclarationResult) addTypeReference(typeName string, location string, isExtend bool) {
	// 忽略空、基础类型或对接口自身的引用。
	if utils.IsBasicType(typeName) || typeName == "" || typeName == inter.Identifier {
		return
	}

	// 如果引用已存在，则追加新的位置信息。
	if ref, exists := inter.Reference[typeName]; exists {
		ref.Location = append(ref.Location, location)
		inter.Reference[typeName] = ref
	} else {
		// 否则，创建一个新的类型引用条目。
		inter.Reference[typeName] = TypeReference{
			Identifier: typeName,
			Location:   []string{location},
			IsExtend:   isExtend,
		}
	}
}

// VisitInterfaceDeclaration 解析接口声明。
func (p *Parser) VisitInterfaceDeclaration(node *ast.InterfaceDeclaration) {
	inter := NewInterfaceDeclarationResult(node.AsNode(), p.SourceCode)
	// 接口名称通常是一个标识符，如果不是，则记录一个错误。
	if !ast.IsIdentifier(node.Name()) {
		p.addError(node.Name(), "Expected interface name to be an identifier, but got %s", node.Name().Kind)
		return
	}
	interfaceName := node.Name().Text()
	inter.Identifier = interfaceName

	// 检查导出关键字
	if modifiers := node.Modifiers(); modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if modifier != nil && modifier.Kind == ast.KindExportKeyword {
				inter.Exported = true
				break
			}
		}
	}

	// 分析 `extends` 子句
	extendsElements := ast.GetExtendsHeritageClauseElements(node.AsNode())
	for _, element := range extendsElements {
		expression := element.Expression()
		if ast.IsIdentifier(expression) {
			name := expression.AsIdentifier().Text
			if !(utils.IsUtilityType(name)) {
				inter.addTypeReference(name, "", true)
			}
		} else if ast.IsPropertyAccessExpression(expression) {
			name := entityNameToString(expression)
			inter.addTypeReference(name, "", true)
		}

		if len(element.TypeArguments()) > 0 {
			for _, typeArg := range element.TypeArguments() {
				results := AnalyzeType(typeArg, "")
				for _, res := range results {
					inter.addTypeReference(res.TypeName, res.Location, true)
				}
			}
		}
	}

	// 分析接口成员
	if node.Members != nil {
		for _, member := range node.Members.Nodes {
			results := AnalyzeMember(member, interfaceName)
			for _, res := range results {
				inter.addTypeReference(res.TypeName, res.Location, false)
			}
		}
	}
	p.Result.InterfaceDeclarations[inter.Identifier] = *inter
}
