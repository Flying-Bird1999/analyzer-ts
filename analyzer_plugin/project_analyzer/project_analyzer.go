package project_analyzer

import (
	"encoding/json"
	"fmt"
	"main/analyzer/parser"
	"main/analyzer/projectParser"
	"os"
	"path/filepath"
	"sync"

	"github.com/samber/lo"
)

// ProjectAnalyzer 是项目分析器的主要结构体。
type ProjectAnalyzer struct {
	rootPath   string
	ignore     []string
	isMonorepo bool
}

// NewProjectAnalyzer 创建一个新的 ProjectAnalyzer 实例。
func NewProjectAnalyzer(rootPath string, ignore []string, isMonorepo bool) *ProjectAnalyzer {
	return &ProjectAnalyzer{
		rootPath:   rootPath,
		ignore:     ignore,
		isMonorepo: isMonorepo,
	}
}

// Analyze 是为旧 `analyze` 命令提供的分析功能，它会解析整个项目并返回详细的AST信息。
func (pa *ProjectAnalyzer) Analyze() (*projectParser.ProjectParserResult, error) {
	config := projectParser.NewProjectParserConfig(pa.rootPath, pa.ignore, pa.isMonorepo)
	ar := projectParser.NewProjectParserResult(config)
	ar.ProjectParser()
	return ar, nil
}

// CheckDependencies 是依赖检查功能的主入口。
// 它负责协调整个检查流程，包括解析项目、并发执行各项检查，并最终返回整合后的结果。
func CheckDependencies(rootPath string, ignore []string, isMonorepo bool) *DependencyCheckResult {
	// 1. 解析整个项目，获取AST和package.json信息
	ar := parseProject(rootPath, ignore, isMonorepo)

	// 2. 准备数据：提取所有已声明的依赖项
	declaredDependencies := make(map[string]bool)
	for _, pkgData := range ar.Package_Data {
		for _, dep := range pkgData.NpmList {
			declaredDependencies[dep.Name] = true
		}
	}

	// 3. 并行执行各项检查
	var wg sync.WaitGroup
	var implicitDeps []ImplicitDependency
	var unusedDeps []UnusedDependency
	var outdatedDeps []OutdatedDependency

	wg.Add(2) // 两个并行的任务：(1)CPU密集型的本地分析 (2)网络密集型的过期检查

	// 任务一: 查找隐式和未使用的依赖 (本地分析)
	go func() {
		defer wg.Done()
		var usedDependencies map[string]bool
		implicitDeps, usedDependencies = findImplicitAndUsedDependencies(ar, declaredDependencies)
		unusedDeps = findUnusedDependencies(ar, usedDependencies)
	}()

	// 任务二: 查找过期的依赖 (网络请求)
	go func() {
		defer wg.Done()
		outdatedDeps = findOutdatedDependencies(ar)
	}()

	wg.Wait() // 等待所有检查任务完成

	// 4. 整合并返回最终结果
	return &DependencyCheckResult{
		ImplicitDependencies: implicitDeps,
		UnusedDependencies:   unusedDeps,
		OutdatedDependencies: outdatedDeps,
	}
}

// parseProject 是一个辅助函数，用于执行底层的项目解析操作。
func parseProject(rootPath string, ignore []string, isMonorepo bool) *projectParser.ProjectParserResult {
	config := projectParser.NewProjectParserConfig(rootPath, ignore, isMonorepo)
	ar := projectParser.NewProjectParserResult(config)
	ar.ProjectParser()
	return ar
}

// AnalyzeProject 是旧 `analyze` 命令的处理器，它将完整的项目分析结果写入JSON文件。
func AnalyzeProject(rootPath string, outputDir string, ignore []string, isMonorepo bool) {
	ar := parseProject(rootPath, ignore, isMonorepo)

	// 在序列化前转换为过滤后的结果
	filteredResult := toFilteredResult(ar)

	jsonData, err := json.MarshalIndent(filteredResult, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling to JSON: %s\n", err)
		return
	}

	// 写入文件，添加命令名称后缀
	outputFile := filepath.Join(outputDir, filepath.Base(rootPath)+"_analyze.json")
	err = os.WriteFile(outputFile, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing JSON to file: %s\n", err)
		return
	}

	fmt.Printf("分析结果已写入文件: %s\n", outputFile)
}

// toFilteredResult 是一个转换函数，用于将完整的分析结果转换为一个更简洁、过滤后的版本，以供 `analyze` 命令输出。
func toFilteredResult(ar *projectParser.ProjectParserResult) *FilteredProjectParserResult {
	filteredJsData := make(map[string]FilteredJsFileParserResult)
	for path, jsData := range ar.Js_Data {
		filteredJsData[path] = FilteredJsFileParserResult{
			ImportDeclarations: lo.Map(jsData.ImportDeclarations, func(decl projectParser.ImportDeclarationResult, _ int) FilteredImportDeclaration {
				return FilteredImportDeclaration{
					ImportModules: lo.Map(decl.ImportModules, func(module projectParser.ImportModule, _ int) parser.ImportModule {
						return parser.ImportModule{
							ImportModule: module.ImportModule,
							Type:         module.Type,
							Identifier:   module.Identifier,
						}
					}),
					Source: decl.Source,
				}
			}),
			ExportDeclarations: lo.Map(jsData.ExportDeclarations, func(decl projectParser.ExportDeclarationResult, _ int) FilteredExportDeclaration {
				return FilteredExportDeclaration{
					ExportModules: lo.Map(decl.ExportModules, func(module projectParser.ExportModule, _ int) parser.ExportModule {
						return parser.ExportModule{
							ModuleName: module.ModuleName,
							Type:       module.Type,
							Identifier: module.Identifier,
						}
					}),
					Source: decl.Source,
				}
			}),
			ExportAssignments: lo.Map(jsData.ExportAssignments, func(assign parser.ExportAssignmentResult, _ int) FilteredExportAssignmentResult {
				return FilteredExportAssignmentResult{Expression: assign.Expression}
			}),
			VariableDeclarations: lo.Map(jsData.VariableDeclarations, func(decl parser.VariableDeclaration, _ int) FilteredVariableDeclaration {
				return FilteredVariableDeclaration{Exported: decl.Exported, Kind: decl.Kind, Source: decl.Source, Declarators: decl.Declarators}
			}),
			CallExpressions: lo.Map(jsData.CallExpressions, func(expr parser.CallExpression, _ int) FilteredCallExpression {
				return FilteredCallExpression{CallChain: expr.CallChain, Arguments: expr.Arguments, Type: expr.Type}
			}),
			JsxElements: lo.Map(jsData.JsxElements, func(elem projectParser.JSXElementResult, _ int) FilteredJSXElement {
				return FilteredJSXElement{ComponentChain: elem.ComponentChain, Attrs: elem.Attrs}
			}),
		}
	}

	return &FilteredProjectParserResult{
		Js_Data:      filteredJsData,
		Package_Data: ar.Package_Data,
	}
}
