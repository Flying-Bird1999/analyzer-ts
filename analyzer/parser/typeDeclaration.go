// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（typeDeclaration.go）专门负责处理和解析 `type` 别名声明。
package parser

import (
	"main/analyzer/utils"
	"strings"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// TypeDeclarationResult 存储一个解析后的 `type` 别名声明信息。
// 它类似于 InterfaceDeclarationResult，也需要追踪其依赖的其他类型。
type TypeDeclarationResult struct {
	Identifier     string                   `json:"identifier"` // 类型别名的名称。
	Raw            string                   `json:"raw"`        // 节点在源码中的原始文本。
	Reference      map[string]TypeReference `json:"reference"`  // 该类型别名所依赖的其他类型的映射。
	SourceLocation SourceLocation           `json:"sourceLocation"` // 节点在源码中的位置信息。
}

// NewTypeDeclarationResult 基于 AST 节点创建一个新的 TypeDeclarationResult 实例。
func NewTypeDeclarationResult(node *ast.Node, sourceCode string) *TypeDeclarationResult {
	raw := utils.GetNodeText(node, sourceCode)
	pos, end := node.Pos(), node.End()

	return &TypeDeclarationResult{
		Identifier: "",
		Raw:        raw,
		Reference:  make(map[string]TypeReference),
		SourceLocation: SourceLocation{
			Start: NodePosition{Line: pos, Column: 0},
			End:   NodePosition{Line: end, Column: 0},
		},
	}
}

// analyzeTypeDecl 是分析 `type` 别名声明的入口函数。
// 它能根据别名的具体定义（对象字面量、映射类型、联合类型等）进行分发处理。
func (tr *TypeDeclarationResult) AnalyzeTypeDecl(typeDecl *ast.TypeAliasDeclaration) {
	typeName := typeDecl.Name().Text()
	tr.Identifier = typeName

	switch typeDecl.Type.Kind {
	// Case 1: 对象字面量类型, e.g., `type MyType = { name: string; data: User; };`
	case ast.KindTypeLiteral:
		if typeDecl.Type.Members() != nil {
			for _, member := range typeDecl.Type.Members() {
				// 复用接口成员的分析逻辑来分析对象字面量的属性。
				memberTypeName, memberLocation := AnalyzeMember(member, typeName)
				if memberTypeName != "" && memberLocation != "" {
					memberTypeNameArray := strings.Split(memberTypeName, ",")
					memberLocationArray := strings.Split(memberLocation, ",")
					for i, name := range memberTypeNameArray {
						tr.addTypeReference(name, memberLocationArray[i], false)
					}
				}
			}
		}
	// Case 2: 映射类型, e.g., `type Translations = { [key in SupportedLanguages]: string; }`
	case ast.KindMappedType:
		mappedTypeNode := typeDecl.Type.AsMappedTypeNode()
		if mappedTypeNode.TypeParameter != nil {
			typeParam := mappedTypeNode.TypeParameter.AsTypeParameter()
			// 分析 `in` 后面的约束类型 (e.g., `SupportedLanguages`)。
			if typeParam.Constraint != nil {
				memberTypeName, _ := AnalyzeType(typeParam.Constraint, "")
				tr.addTypeReference(memberTypeName, "", false)
			}

			// 分析映射的值类型 (e.g., `string`)。
			if mappedTypeNode.Type != nil {
				memberTypeName, memberLocation := AnalyzeType(mappedTypeNode.Type, "")
				if memberTypeName != "" {
					memberTypeNameArray := strings.Split(memberTypeName, ",")
					memberLocationArray := strings.Split(memberLocation, ",")
					for i, name := range memberTypeNameArray {
						tr.addTypeReference(name, memberLocationArray[i], false)
					}
				}
			}
		}
	// Case 3: 其他类型，如联合类型、交叉类型、直接的类型引用等。
	// e.g., `type MyUnion = TypeA | TypeB;`
	default:
		// 直接使用通用的类型分析函数来处理。
		memberTypeName, memberLocation := AnalyzeType(typeDecl.Type, typeName)
		if memberTypeName != "" {
			memberTypeNameArray := strings.Split(memberTypeName, ",")
			memberLocationArray := strings.Split(memberLocation, ",")
			for i, name := range memberTypeNameArray {
				tr.addTypeReference(name, memberLocationArray[i], false)
			}
		}
	}
}

// addTypeReference 是一个辅助函数，用于向结果的 Reference 映射中添加或更新类型引用。
// (此函数与 interfaceDeclaration.go 中的版本逻辑相同)
func (tr *TypeDeclarationResult) addTypeReference(typeName string, location string, isExtend bool) {
	// 忽略基础类型（string, number, boolean 等）。
	if utils.IsBasicType(typeName) {
		return
	}

	// 忽略对类型别名自身的引用。
	if typeName == tr.Identifier {
		return
	}

	if ref, exists := tr.Reference[typeName]; exists {
		// 如果引用已存在，则追加新的位置信息。
		ref.Location = append(ref.Location, location)
		tr.Reference[typeName] = ref
	} else {
		// 否则，创建一个新的类型引用条目。
		tr.Reference[typeName] = TypeReference{
			Identifier: typeName,
			Location:   []string{location},
			IsExtend:   isExtend, // 对于 type 别名，isExtend 通常为 false
		}
	}
}
