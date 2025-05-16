package parser

import (
	"main/bundle/utils"
	"strings"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// 解析 type 声明，递归去查找 type 里边的类型
// - 如果有引用外部类型的就找出来
// - case1: type Name3 = LinearModel | Person;
// - case2: type Name = { name: string; age: LinearModel; };

type TypeDeclarationResult struct {
	Name      string // 名称
	Raw       string // 源码
	Reference map[string]TypeReference
}

func NewTypeDeclarationResult(node *ast.Node, sourceCode string) *TypeDeclarationResult {
	raw := utils.GetNodeText(node.AsNode(), sourceCode)

	return &TypeDeclarationResult{
		Name:      "",
		Raw:       raw,
		Reference: make(map[string]TypeReference),
	}
}

// 分析接口的主要结构，包括：
// 1. 接口名称。
// 2. 类型成员（通过 analyzeMember）。
func (tr *TypeDeclarationResult) analyzeTypeDecl(typeDecl *ast.TypeAliasDeclaration) {
	typeName := typeDecl.Name().Text()
	tr.Name = typeName

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

	if ref, exists := tr.Reference[typeName]; exists {
		// 如果类型引用已存在，追加新的位置
		ref.Location = append(ref.Location, location)
		tr.Reference[typeName] = ref
	} else {
		// 如果类型引用不存在，创建新的引用
		tr.Reference[typeName] = TypeReference{
			Name:     typeName,
			Location: []string{location},
			IsExtend: isExtend,
		}
	}
}
