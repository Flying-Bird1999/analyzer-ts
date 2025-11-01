//go:build example09

package main

import (
	"fmt"
	"os"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run 10-advanced-symbols.go <TypeScript项目路径>")
		os.Exit(1)
	}

	projectPath := os.Args[1]

	fmt.Println("🔣 高级符号分析示例 - 深度符号关系分析")
	fmt.Println("==================================================")

	// 创建项目
	config := tsmorphgo.ProjectConfig{
		RootPath:         projectPath,
		IgnorePatterns:   []string{"node_modules", "dist", "build", ".git"},
		TargetExtensions: []string{".ts", ".tsx"},
	}
	project := tsmorphgo.NewProject(config)

	// 高级符号分析
	analysis := performAdvancedSymbolAnalysis(project)

	// 显示分析结果
	fmt.Println("\n📊 高级符号分析结果:")
	fmt.Printf("  总符号数: %d\n", analysis.TotalSymbols)
	fmt.Printf("  导出符号: %d\n", analysis.ExportedSymbols)
	fmt.Printf("  类型分布: %v\n", analysis.TypeDistribution)

	// 显示符号层次结构
	fmt.Println("\n🌳 符号层次结构:")
	printSymbolHierarchy(analysis.SymbolHierarchy, 0)

	// 显示符号关系
	fmt.Println("\n🔗 符号关系分析:")
	printSymbolRelationships(analysis.SymbolRelationships)

	// 显示引用分析
	fmt.Println("\n📚 引用分析 (前 5 个):")
	for i, refAnalysis := range analysis.ReferenceAnalyses {
		if i >= 5 {
			break
		}
		printReferenceAnalysis(refAnalysis)
	}

	// 显示模块分析
	fmt.Println("\n📦 模块分析:")
	printModuleAnalysis(analysis.ModuleAnalysis)

	// 显示复杂度分析
	fmt.Println("\n🧩 复杂度分析 (前 5 个):")
	for i, complexity := range analysis.ComplexityAnalyses {
		if i >= 5 {
			break
		}
		printComplexityAnalysis(complexity)
	}

	fmt.Println("\n✅ 高级符号分析完成！")
}

// AdvancedSymbolAnalysis 高级符号分析结果
type AdvancedSymbolAnalysis struct {
	TotalSymbols        int                    `json:"totalSymbols"`
	ExportedSymbols    int                    `json:"exportedSymbols"`
	TypeDistribution   map[string]int         `json:"typeDistribution"`
	SymbolHierarchy    []*SymbolHierarchyNode  `json:"symbolHierarchy"`
	SymbolRelationships []*SymbolRelationship  `json:"symbolRelationships"`
	ReferenceAnalyses   []*ReferenceAnalysis   `json:"referenceAnalyses"`
	ModuleAnalysis      *ModuleAnalysis        `json:"moduleAnalysis"`
	ComplexityAnalyses  []*ComplexityAnalysis  `json:"complexityAnalyses"`
}

// SymbolHierarchyNode 符号层次节点
type SymbolHierarchyNode struct {
	Symbol   *tsmorphgo.Symbol `json:"symbol"`
	Children []*SymbolHierarchyNode `json:"children"`
	Depth    int               `json:"depth"`
}

// SymbolRelationship 符号关系
type SymbolRelationship struct {
	FromSymbol *tsmorphgo.Symbol `json:"fromSymbol"`
	ToSymbol   *tsmorphgo.Symbol `json:"toSymbol"`
	RelationshipType string        `json:"relationshipType"`
	Strength   int               `json:"strength"`
}

// ReferenceAnalysis 引用分析
type ReferenceAnalysis struct {
	Symbol         *tsmorphgo.Symbol `json:"symbol"`
	References     []tsmorphgo.Node  `json:"references"`
	ReferenceCount int             `json:"referenceCount"`
	CrossFileRefs  int             `json:"crossFileRefs"`
	SameFileRefs   int             `json:"sameFileRefs"`
}

// ModuleAnalysis 模块分析
type ModuleAnalysis struct {
	Modules        []*ModuleInfo      `json:"modules"`
	Dependencies   []ModuleDependency `json:"dependencies"`
	ExportMap      map[string][]string `json:"exportMap"`
}

// ModuleInfo 模块信息
type ModuleInfo struct {
	Path           string                   `json:"path"`
	ExportedCount  int                      `json:"exportedCount"`
	Symbols        map[string]*tsmorphgo.Symbol `json:"symbols"`
}

// ModuleDependency 模块依赖
type ModuleDependency struct {
	FromModule string `json:"fromModule"`
	ToModule   string `json:"toModule"`
	Strength   int    `json:"strength"`
	DependencyType string `json:"dependencyType"`
}

// ComplexityAnalysis 复杂度分析
type ComplexityAnalysis struct {
	Symbol      *tsmorphgo.Symbol `json:"symbol"`
	Complexity  int               `json:"complexity"`
	Depth       int               `json:"depth"`
	Children    int               `json:"children"`
	Members     int               `json:"members"`
}

// performAdvancedSymbolAnalysis 执行高级符号分析
func performAdvancedSymbolAnalysis(project *tsmorphgo.Project) *AdvancedSymbolAnalysis {
	analysis := &AdvancedSymbolAnalysis{
		TypeDistribution: make(map[string]int),
		ModuleAnalysis:  &ModuleAnalysis{
			Modules:      []*ModuleInfo{},
			Dependencies: []ModuleDependency{},
			ExportMap:    make(map[string][]string),
		},
	}

	// 收集所有符号
	symbolMap := make(map[string]*tsmorphgo.Symbol)
	for _, sf := range project.GetSourceFiles() {
		fileSymbols := collectFileSymbols(sf)
		for name, symbol := range fileSymbols {
			symbolMap[name] = symbol
		}
	}
	analysis.TotalSymbols = len(symbolMap)

	// 构建符号层次结构
	hierarchy, exportedCount := buildSymbolHierarchy(symbolMap)
	analysis.SymbolHierarchy = hierarchy
	analysis.ExportedSymbols = exportedCount

	// 分析类型分布
	analyzeTypeDistribution(analysis)

	// 分析符号关系
	analysis.SymbolRelationships = analyzeSymbolRelationships(symbolMap)

	// 分析引用关系
	analysis.ReferenceAnalyses = analyzeReferences(symbolMap)

	// 分析模块结构
	analyzeModuleStructure(project, analysis)

	// 分析复杂度
	analysis.ComplexityAnalyses = analyzeComplexity(symbolMap)

	return analysis
}

// collectFileSymbols 收集文件中的符号
func collectFileSymbols(sf *tsmorphgo.SourceFile) map[string]*tsmorphgo.Symbol {
	symbols := make(map[string]*tsmorphgo.Symbol)

	sf.ForEachDescendant(func(node tsmorphgo.Node) {
		if symbol, ok := tsmorphgo.GetSymbol(node); ok {
			name := symbol.GetName()
			if name != "" {
				symbols[name] = symbol
			}
		}
	})

	return symbols
}

// buildSymbolHierarchy 构建符号层次结构
func buildSymbolHierarchy(symbols map[string]*tsmorphgo.Symbol) ([]*SymbolHierarchyNode, int) {
	var hierarchy []*SymbolHierarchyNode
	exportedCount := 0

	// 构建层次树
	for _, symbol := range symbols {
		node := &SymbolHierarchyNode{
			Symbol:  symbol,
			Depth:   0,
			Children: []*SymbolHierarchyNode{},
		}

	// 检查是否是导出的
		if symbol.IsExported() {
			exportedCount++
		}

	// 查找子符号
	if members := symbol.GetMembers(); len(members) > 0 {
			for _, member := range members {
				childNode := &SymbolHierarchyNode{
					Symbol:  member,
					Depth:   1,
					Children: []*SymbolHierarchyNode{},
				}
				node.Children = append(node.Children, childNode)
			}
		}

		hierarchy = append(hierarchy, node)
	}

	return hierarchy, exportedCount
}

// analyzeTypeDistribution 分析类型分布
func analyzeTypeDistribution(analysis *AdvancedSymbolAnalysis) {
	for _, node := range analysis.SymbolHierarchy {
		symbol := node.Symbol
		classifySymbolType(symbol, analysis.TypeDistribution)
	}
}

// classifySymbolType 分类符号类型
func classifySymbolType(symbol *tsmorphgo.Symbol, distribution map[string]int) {
	switch {
	case symbol.IsFunction():
		distribution["function"]++
	case symbol.IsClass():
		distribution["class"]++
	case symbol.IsInterface():
		distribution["interface"]++
	case symbol.IsTypeAlias():
		distribution["typeAlias"]++
	case symbol.IsVariable():
		distribution["variable"]++
	case symbol.IsMethod():
		distribution["method"]++
	case symbol.IsProperty():
		distribution["property"]++
	case symbol.IsEnum():
		distribution["enum"]++
	case symbol.IsModule():
		distribution["module"]++
	default:
		distribution["unknown"]++
	}
}

// analyzeSymbolRelationships 分析符号关系
func analyzeSymbolRelationships(symbols map[string]*tsmorphgo.Symbol) []*SymbolRelationship {
	var relationships []*SymbolRelationship

	// 分析父子关系
	for _, symbol := range symbols {
		if parent, ok := symbol.GetParent(); ok {
			relationship := &SymbolRelationship{
				FromSymbol:      symbol,
				ToSymbol:        parent,
				RelationshipType: "parent-child",
				Strength:        1,
			}
			relationships = append(relationships, relationship)
		}
	}

	return relationships
}

// analyzeReferences 分析引用关系
func analyzeReferences(symbols map[string]*tsmorphgo.Symbol) []*ReferenceAnalysis {
	var analyses []*ReferenceAnalysis

	for _, symbol := range symbols {
		analysis := &ReferenceAnalysis{
			Symbol: symbol,
		}

		// 查找引用
		if refs, err := symbol.FindReferences(); err == nil {
			analysis.References = refs
			analysis.ReferenceCount = len(refs)

			// 分析跨文件引用
			symbolFile := getSymbolFile(symbol)
			for _, ref := range refs {
				if ref.GetSourceFile().GetFilePath() != symbolFile {
					analysis.CrossFileRefs++
				} else {
					analysis.SameFileRefs++
				}
			}
		}

		analyses = append(analyses, analysis)
	}

	return analyses
}

// analyzeModuleStructure 分析模块结构
func analyzeModuleStructure(project *tsmorphgo.Project, analysis *AdvancedSymbolAnalysis) {
	// 分析每个文件的导出
	for _, sf := range project.GetSourceFiles() {
		module := &ModuleInfo{
			Path:    sf.GetFilePath(),
			Symbols:  make(map[string]*tsmorphgo.Symbol),
		}

		// 收集模块的符号
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			if symbol, ok := tsmorphgo.GetSymbol(node); ok {
				if symbol.IsExported() {
					module.ExportedCount++
					module.Symbols[symbol.GetName()] = symbol
					analysis.ModuleAnalysis.ExportMap[sf.GetFilePath()] = append(
						analysis.ModuleAnalysis.ExportMap[sf.GetFilePath()],
						symbol.GetName(),
					)
				}
			}
		})

		analysis.ModuleAnalysis.Modules = append(analysis.ModuleAnalysis.Modules, module)
	}

	// 分析模块间的依赖关系
	analyzeModuleDependencies(analysis)
}

// analyzeModuleDependencies 分析模块依赖
func analyzeModuleDependencies(analysis *AdvancedSymbolAnalysis) {
	// 这里可以添加模块依赖分析逻辑
	// 通过 import 语句分析文件间的依赖关系
}

// analyzeComplexity 分析复杂度
func analyzeComplexity(symbols map[string]*tsmorphgo.Symbol) []*ComplexityAnalysis {
	var analyses []*ComplexityAnalysis

	for _, symbol := range symbols {
		complexity := &ComplexityAnalysis{
			Symbol: symbol,
		}

		// 计算复杂度
		complexity.Complexity = calculateSymbolComplexity(symbol)
		complexity.Depth = calculateSymbolDepth(symbol)
		complexity.Children = len(symbol.GetMembers())
		complexity.Members = len(symbol.GetDeclarations())

		analyses = append(analyses, complexity)
	}

	return analyses
}

// calculateSymbolComplexity 计算符号复杂度
func calculateSymbolComplexity(symbol *tsmorphgo.Symbol) int {
	complexity := 0

	// 基础复杂度
	if symbol.IsClass() {
		complexity += 5
	}
	if symbol.IsInterface() {
		complexity += 3
	}
	if symbol.IsFunction() {
		complexity += 2
	}

	// 成员数量影响
	members := symbol.GetMembers()
	complexity += len(members)

	// 引用数量影响
	if refs, err := symbol.FindReferences(); err == nil {
		complexity += len(refs) / 10
	}

	return complexity
}

// calculateSymbolDepth 计算符号深度
func calculateSymbolDepth(symbol *tsmorphgo.Symbol) int {
	depth := 0
	current := symbol

	for {
		parent, ok := current.GetParent()
		if !ok {
			break
		}
		depth++
		current = parent
	}

	return depth
}

// getSymbolFile 获取符号所在文件
func getSymbolFile(symbol *tsmorphgo.Symbol) string {
	decls := symbol.GetDeclarations()
	if len(decls) > 0 {
		return decls[0].GetSourceFile().GetFilePath()
	}
	return ""
}

// printSymbolHierarchy 打印符号层次结构
func printSymbolHierarchy(nodes []*SymbolHierarchyNode, indent int) {
	for _, node := range nodes {
		prefix := ""
		for i := 0; i < indent; i++ {
			prefix += "  "
		}

		symbol := node.Symbol
		exported := ""
		if symbol.IsExported() {
			exported = " ✅"
		}

		fmt.Printf("%s- %s%s (%s)%s\n", prefix, symbol.GetName(), getSymbolTypeName(symbol), exported, exported)

		if len(node.Children) > 0 {
			printSymbolHierarchy(node.Children, indent+1)
		}
	}
}

// printSymbolRelationships 打印符号关系
func printSymbolRelationships(relationships []*SymbolRelationship) {
	for i, rel := range relationships {
		if i >= 3 {
			break
		}
		fmt.Printf("  %d. %s -> %s (%s)\n",
			i+1,
			rel.FromSymbol.GetName(),
			rel.ToSymbol.GetName(),
			rel.RelationshipType,
		)
	}
}

// printReferenceAnalysis 打印引用分析
func printReferenceAnalysis(analysis *ReferenceAnalysis) {
	fmt.Printf("  %s: %d 引用 (跨文件: %d, 同文件: %d)\n",
		analysis.Symbol.GetName(),
		analysis.ReferenceCount,
		analysis.CrossFileRefs,
		analysis.SameFileRefs,
	)
}

// printModuleAnalysis 打印模块分析
func printModuleAnalysis(analysis *ModuleAnalysis) {
	fmt.Printf("  模块数量: %d\n", len(analysis.Modules))
	fmt.Printf("  导出映射数量: %d\n", len(analysis.ExportMap))
	fmt.Printf("  依赖数量: %d\n", len(analysis.Dependencies))

	if len(analysis.Modules) > 0 {
		fmt.Printf("  导出最多的模块: %s (%d)\n",
			analysis.Modules[0].Path,
			analysis.Modules[0].ExportedCount,
		)
	}
}

// printComplexityAnalysis 打印复杂度分析
func printComplexityAnalysis(analysis *ComplexityAnalysis) {
	fmt.Printf("  %s: 复杂度=%d, 深度=%d, 成员=%d\n",
		analysis.Symbol.GetName(),
		analysis.Complexity,
		analysis.Depth,
		analysis.Members,
	)
}

// getSymbolTypeName 获取符号类型名称
func getSymbolTypeName(symbol *tsmorphgo.Symbol) string {
	switch {
	case symbol.IsFunction():
		return "function"
	case symbol.IsClass():
		return "class"
	case symbol.IsInterface():
		return "interface"
	case symbol.IsTypeAlias():
		return "typeAlias"
	case symbol.IsVariable():
		return "variable"
	case symbol.IsMethod():
		return "method"
	case symbol.IsProperty():
		return "property"
	case symbol.IsEnum():
		return "enum"
	case symbol.IsModule():
		return "module"
	default:
		return "unknown"
	}
}