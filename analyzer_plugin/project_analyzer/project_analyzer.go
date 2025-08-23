// package project_analyzer 是整个分析器插件的核心包。
// 它作为业务逻辑层的根，提供了项目分析的统一入口，并组织了如此下各种专业的分析器子包：
// - callgraph: 调用链分析
// - dependency: NPM依赖分析
// - unreferenced: 未引用文件分析
package project_analyzer

import (
	"encoding/json"
	"fmt"
	"main/analyzer/parser"
	"main/analyzer/projectParser"
	"os"
	"path/filepath"

	"github.com/samber/lo"
)

// ProjectAnalyzer 是项目分析器的主要结构体。
// 它封装了执行项目分析所需的所有配置和状态。
type ProjectAnalyzer struct {
	rootPath   string
	ignore     []string
	isMonorepo bool
}

// NewProjectAnalyzer 创建一个新的 ProjectAnalyzer 实例。
// rootPath: 要分析的项目根目录。
// ignore: 需要从分析中排除的文件/目录的 glob 模式列表。
// isMonorepo: 指示项目是否为 monorepo。
func NewProjectAnalyzer(rootPath string, ignore []string, isMonorepo bool) *ProjectAnalyzer {
	return &ProjectAnalyzer{
		rootPath:   rootPath,
		ignore:     ignore,
		isMonorepo: isMonorepo,
	}
}

// Analyze 是一个核心方法，它调用底层的 projectParser 来对整个项目进行深度分析，
// 并返回包含所有文件AST信息的详细结果。这个原始结果是所有上层专业分析器的基础。
func (pa *ProjectAnalyzer) Analyze() (*projectParser.ProjectParserResult, error) {
	config := projectParser.NewProjectParserConfig(pa.rootPath, pa.ignore, pa.isMonorepo)
	ar := projectParser.NewProjectParserResult(config)
	ar.ProjectParser()
	// 在未来的版本中，这里可以增加错误处理的逻辑。
	return ar, nil
}

// parseProject 是一个辅助函数，封装了执行底层项目解析操作的具体步骤。
func parseProject(rootPath string, ignore []string, isMonorepo bool) *projectParser.ProjectParserResult {
	config := projectParser.NewProjectParserConfig(rootPath, ignore, isMonorepo)
	ar := projectParser.NewProjectParserResult(config)
	ar.ProjectParser()
	return ar
}

// AnalyzeProject 是为 `analyze` 命令提供的处理器。
// 它执行完整的项目分析，并将结果以过滤和序列化后的JSON格式写入文件。
func AnalyzeProject(rootPath string, outputDir string, ignore []string, isMonorepo bool) {
	ar := parseProject(rootPath, ignore, isMonorepo)

	// 在序列化之前，将完整的分析结果转换为一个更简洁、过滤后的版本。
	filteredResult := toFilteredResult(ar)

	jsonData, err := json.MarshalIndent(filteredResult, "", "  ")
	if err != nil {
		fmt.Printf("序列化JSON时出错: %s", err)
		return
	}

	// 将JSON数据写入以项目名命名的输出文件中。
	outputFile := filepath.Join(outputDir, filepath.Base(rootPath)+"_analyze.json")
	err = os.WriteFile(outputFile, jsonData, 0644)
	if err != nil {
		fmt.Printf("写入JSON文件时出错: %s", err)
		return
	}

	fmt.Printf("分析结果已写入文件: %s", outputFile)
}

// toFilteredResult 是一个转换函数，用于将完整的、包含大量原始信息的分析结果 (ProjectParserResult)
// 转换为一个更简洁、过滤后的版本 (FilteredProjectParserResult)，以供 `analyze` 命令输出。
// 这样可以减少输出文件的大小，使其更具可读性。
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
			FunctionDeclarations: lo.Map(jsData.FunctionDeclarations, func(elem parser.FunctionDeclarationResult, _ int) FilteredFunctionDeclarationResult {
				return FilteredFunctionDeclarationResult{
					Exported:   elem.Exported,
					Identifier: elem.Identifier,
					IsAsync:    elem.IsAsync,
					Parameters: elem.Parameters,
					ReturnType: elem.ReturnType,
				}
			}),
		}
	}

	return &FilteredProjectParserResult{
		Js_Data:      filteredJsData,
		Package_Data: ar.Package_Data,
	}
}
