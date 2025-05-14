package parser

import (
	"fmt"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// InterfaceDependency 表示接口依赖的类型
type InterfaceDependency struct {
	Name     string
	IsBasic  bool   // 是否是基本类型
	TypeKind string // 类型的种类
}

// InterfaceInfo 表示接口的信息
type InterfaceInfo struct {
	Name         string
	Properties   []string
	Dependencies []InterfaceDependency
}

// ExtractTypeScriptInterface 从TypeScript文件中提取接口信息
func ExtractTypeScriptInterface(interfaceDeclaration *ast.InterfaceDeclaration, sourceCode string) ([]InterfaceInfo, error) {
	// 存储接口信息的结果
	var interfaces []InterfaceInfo

	// 创建接口信息
	interfaceInfo := InterfaceInfo{
		Name: interfaceDeclaration.Name().AsIdentifier().Text,
	}

	// 提取属性和方法
	if interfaceDeclaration.Members != nil {
		for _, member := range interfaceDeclaration.Members.Nodes {
			if ast.IsPropertySignatureDeclaration(member) {
				propSig := member.AsPropertySignatureDeclaration()
				interfaceInfo.Properties = append(interfaceInfo.Properties, propSig.Name().AsIdentifier().Text)

				// 提取属性类型依赖
				if propSig.Type != nil {
					dependency := extractTypeDependency(propSig.Type)
					if dependency != nil {
						interfaceInfo.Dependencies = append(interfaceInfo.Dependencies, *dependency)
					}
				}
			}
		}
	}

	// 提取继承的接口
	if interfaceDeclaration.HeritageClauses != nil {
		for _, clause := range interfaceDeclaration.HeritageClauses.Nodes {
			heritageClause := clause.AsHeritageClause()
			if heritageClause.Token == ast.KindExtendsKeyword {
				for _, type_ := range heritageClause.Types.Nodes {
					expr := type_.AsExpressionWithTypeArguments()
					if ast.IsIdentifier(expr.Expression) {
						interfaceInfo.Dependencies = append(interfaceInfo.Dependencies, InterfaceDependency{
							Name:     expr.Expression.AsIdentifier().Text,
							IsBasic:  false,
							TypeKind: "interface",
						})
					}
				}
			}
		}
	}

	interfaces = append(interfaces, interfaceInfo)

	return interfaces, nil
}

// extractTypeDependency 提取类型依赖
func extractTypeDependency(typeNode *ast.TypeNode) *InterfaceDependency {
	if typeNode == nil {
		return nil
	}

	// 根据类型节点的种类提取依赖
	if ast.IsTypeReferenceNode(typeNode) {
		typeRef := typeNode.AsTypeReferenceNode()
		if ast.IsIdentifier(typeRef.TypeName) {
			return &InterfaceDependency{
				Name:     typeRef.TypeName.AsIdentifier().Text,
				IsBasic:  false,
				TypeKind: "reference",
			}
		}
	} else if ast.IsKeywordKind(typeNode.Kind) {
		// 处理基本类型
		return &InterfaceDependency{
			Name:     typeNode.Kind.String(),
			IsBasic:  true,
			TypeKind: "keyword",
		}
	} else if typeNode.Kind == ast.KindArrayType {
		// 处理数组类型
		arrayType := typeNode.AsArrayTypeNode()
		elemDep := extractTypeDependency(arrayType.ElementType)
		if elemDep != nil {
			elemDep.TypeKind = "array"
			return elemDep
		}
	} else if typeNode.Kind == ast.KindUnionType {
		// 处理联合类型
		unionType := typeNode.AsUnionTypeNode()
		if unionType.Types != nil && len(unionType.Types.Nodes) > 0 {
			// 简化处理，只返回第一个类型
			firstType := unionType.Types.Nodes[0]
			dep := extractTypeDependency(firstType)
			if dep != nil {
				dep.TypeKind = "union"
				return dep
			}
		}
	}

	return nil
}

// 解析当前文件中的 interface 声明
// - 当指定某个 interface 后，找到内部依赖的其他类型
//  	- 需要考虑 继承 extends 的case

func TraverseInterfaceDeclaration(node *ast.InterfaceDeclaration, sourceCode string) {
	// ✅ 解析 interface 的源代码
	// raw := utils.GetNodeText(node.AsNode(), sourceCode)
	// fmt.Printf("源代码 raw: %s\n", raw)

	interfaces, err := ExtractTypeScriptInterface(node, sourceCode)

	fmt.Printf("interfaces: %v\n", interfaces)

	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	// 输出结果
	for _, iface := range interfaces {
		fmt.Printf("接口名称: %s\n", iface.Name)
		fmt.Println("属性:")
		for _, prop := range iface.Properties {
			fmt.Printf("  - %s\n", prop)
		}
		fmt.Println("依赖类型:")
		for _, dep := range iface.Dependencies {
			fmt.Printf("  - %s (基本类型: %v, 类型: %s)\n", dep.Name, dep.IsBasic, dep.TypeKind)
		}
		fmt.Println("\n\n\n")
	}
}
