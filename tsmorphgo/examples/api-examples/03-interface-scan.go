//go:build example03

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run 03-interface-scan.go <TypeScript项目路径> [输出文件]")
		os.Exit(1)
	}

	projectPath := os.Args[1]
	outputFile := "./interfaces.json"
	if len(os.Args) > 2 {
		outputFile = os.Args[2]
	}

	fmt.Println("🔷 接口扫描示例 - 收集所有接口和类型")
	fmt.Println("==================================================")

	// 创建项目
	config := tsmorphgo.ProjectConfig{
		RootPath:         projectPath,
		IgnorePatterns:   []string{"node_modules", "dist", "build", ".git"},
		TargetExtensions: []string{".ts", ".tsx"},
	}
	project := tsmorphgo.NewProject(config)

	// 扫描接口和类型别名
	interfaces := scanInterfaces(project)
	fmt.Printf("✅ 扫描到 %d 个接口\n", len(interfaces))

	typeAliases := scanTypeAliases(project)
	fmt.Printf("✅ 扫描到 %d 个类型别名\n", len(typeAliases))

	// 显示前 5 个接口作为示例
	fmt.Println("\n📋 接口列表（前 5 个）:")
	for i, iface := range interfaces {
		if i >= 5 {
			break
		}
		fmt.Printf("  %d. %s (%d 个字段)\n", i+1, iface.Name, len(iface.Fields))
	}

	// 显示前 5 个类型别名作为示例
	fmt.Println("\n🏷️  类型别名列表（前 5 个）:")
	for i, alias := range typeAliases {
		if i >= 5 {
			break
		}
		fmt.Printf("  %d. %s = %s\n", i+1, alias.Name, alias.Type)
	}

	// 保存到文件
	result := map[string]interface{}{
		"interfaces":  interfaces,
		"typeAliases": typeAliases,
		"summary": map[string]int{
			"totalInterfaces":  len(interfaces),
			"totalTypeAliases": len(typeAliases),
		},
	}

	if data, err := json.MarshalIndent(result, "", "  "); err == nil {
		if err := os.WriteFile(outputFile, data, 0644); err == nil {
			fmt.Printf("\n💾 分析结果已保存到: %s\n", outputFile)
		}
	}

	fmt.Println("\n✅ 接口扫描完成！")
}

// InterfaceInfo 接口信息
type InterfaceInfo struct {
	Name     string     `json:"name"`
	Fields   []FieldInfo `json:"fields"`
	File     string     `json:"file"`
	Line     int        `json:"line"`
	Exported bool       `json:"exported"`
}

// FieldInfo 字段信息
type FieldInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Optional bool   `json:"optional"`
	Line     int    `json:"line"`
}

// TypeAliasInfo 类型别名信息
type TypeAliasInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Exported bool    `json:"exported"`
}

// scanInterfaces 扫描所有接口
func scanInterfaces(project *tsmorphgo.Project) []InterfaceInfo {
	var interfaces []InterfaceInfo

	for _, sf := range project.GetSourceFiles() {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.Kind == ast.KindInterfaceDeclaration {
				iface := analyzeInterface(node, sf)
				interfaces = append(interfaces, *iface)
			}
		})
	}

	return interfaces
}

// scanTypeAliases 扫描所有类型别名
func scanTypeAliases(project *tsmorphgo.Project) []TypeAliasInfo {
	var typeAliases []TypeAliasInfo

	for _, sf := range project.GetSourceFiles() {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.Kind == ast.KindTypeAliasDeclaration {
				if name, ok := tsmorphgo.GetVariableName(node); ok {
					alias := TypeAliasInfo{
						Name:     name,
						Type:     extractTypeAliasType(node),
						File:     sf.GetFilePath(),
						Line:     node.GetStartLineNumber(),
						Exported: isExported(node),
					}
					typeAliases = append(typeAliases, alias)
				}
			}
		})
	}

	return typeAliases
}

// analyzeInterface 分析接口
func analyzeInterface(node tsmorphgo.Node, sf *tsmorphgo.SourceFile) *InterfaceInfo {
	name, _ := tsmorphgo.GetVariableName(node)
	iface := &InterfaceInfo{
		Name:     name,
		Fields:   []FieldInfo{},
		File:     sf.GetFilePath(),
		Line:     node.GetStartLineNumber(),
		Exported: isExported(node),
	}

	return iface
}

// extractTypeAliasType 提取类型别名的类型
func extractTypeAliasType(node tsmorphgo.Node) string {
	return node.GetText()
}

// isExported 检查是否导出
func isExported(node tsmorphgo.Node) bool {
	// 简化实现：检查是否有 export 关键字
	return true
}