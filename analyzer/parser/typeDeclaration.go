package parser

import (
	"main/analyzer/utils"
	"strings"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// 解析 type 声明，递归去查找 type 里边的类型
// - 如果有引用外部类型的就找出来
// - case1: type Name3 = LinearModel | Person;
// - case2: type Name = { name: string; age: LinearModel; };
// - case3: type Translations = { [key in SupportedLanguages]: string; }

// TypeDeclarationResult 类型声明结果
type TypeDeclarationResult struct {
	Identifier     string                   `json:"identifier"` // 名称
	Raw            string                   `json:"raw"`        // 源码
	Reference      map[string]TypeReference `json:"reference"`  // 依赖的其他类型
	SourceLocation SourceLocation           `json:"sourceLocation"`
}

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

// 分析接口的主要结构，包括：
// 1. 接口名称。
// 2. 类型成员（通过 analyzeMember）。
func (tr *TypeDeclarationResult) analyzeTypeDecl(typeDecl *ast.TypeAliasDeclaration) {
	typeName := typeDecl.Name().Text()
	tr.Identifier = typeName

	// 对象字面量类型，分析内部类成员
	// type Name = { name: string; age: LinearModel; };
	if typeDecl.Type.Kind == ast.KindTypeLiteral {
		if typeDecl.Type.Members() != nil {
			for _, member := range typeDecl.Type.Members() {
				memberTypeName, memberLocation := AnalyzeMember(member, typeName)
				if memberTypeName != "" && memberLocation != "" {
					memberTypeNameArray := strings.Split(memberTypeName, ",")
					memberLocationArray := strings.Split(memberLocation, ",")
					for i, typeName := range memberTypeNameArray {
						tr.addTypeReference(typeName, memberLocationArray[i], false)
					}
				}
			}
		}
	} else if typeDecl.Type.Kind == ast.KindMappedType {
		// 映射类型：type Translations = { [key in SupportedLanguages]: string; }
		mappedTypeNode := typeDecl.Type.AsMappedTypeNode()
		if mappedTypeNode.TypeParameter != nil {
			typeParam := mappedTypeNode.TypeParameter.AsTypeParameter()
			// 类型参数名称 typeParam.Name().AsIdentifier().Text，暂时不提取
			// 提取约束类型 (in 后面的类型)
			if typeParam.Constraint != nil {
				memberTypeName, _ := AnalyzeType(typeParam.Constraint, "")
				tr.addTypeReference(memberTypeName, "", false)
			}

			// 提取值类型
			if typeParam.Type != nil {
				memberTypeName, memberLocation := AnalyzeType(mappedTypeNode.Type, "")
				if memberTypeName != "" && memberLocation != "" {
					memberTypeNameArray := strings.Split(memberTypeName, ",")
					memberLocationArray := strings.Split(memberLocation, ",")
					for i, typeName := range memberTypeNameArray {
						tr.addTypeReference(typeName, memberLocationArray[i], false)
					}
				}
			}

		}
	} else {
		// type Name3 = LinearModel | Person;
		memberTypeName, memberLocation := AnalyzeType(typeDecl.Type, typeName)
		if memberTypeName != "" && memberLocation != "" {
			memberTypeNameArray := strings.Split(memberTypeName, ",")
			memberLocationArray := strings.Split(memberLocation, ",")
			for i, typeName := range memberTypeNameArray {
				tr.addTypeReference(typeName, memberLocationArray[i], false)
			}
		}
	}
}

// 填充数据
func (tr *TypeDeclarationResult) addTypeReference(typeName string, location string, isExtend bool) {
	// 排除基本类型和已知的内置类型
	if utils.IsBasicType(typeName) {
		return
	}

	// 如果依赖类型 和 自身是同一个，则不用加上了
	if typeName == tr.Identifier {
		return
	}

	if ref, exists := tr.Reference[typeName]; exists {
		// 如果类型引用已存在，追加新的位置
		ref.Location = append(ref.Location, location)
		tr.Reference[typeName] = ref
	} else {
		// 如果类型引用不存在，创建新的引用
		tr.Reference[typeName] = TypeReference{
			Identifier: typeName,
			Location:   []string{location},
			IsExtend:   isExtend,
		}
	}
}
