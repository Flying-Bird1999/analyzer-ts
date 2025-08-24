// package project_analyzer 是整个分析器插件的核心包。
// 它作为业务逻辑层的根，提供了项目分析的统一入口，并组织了如此下各种专业的分析器子包：
// - callgraph: 调用链分析
// - dependency: NPM依赖分析
// - unreferenced: 未引用文件分析
package project_analyzer

import (
	"fmt"
	"main/analyzer/parser"
	"main/analyzer/projectParser"
	"main/analyzer_plugin/project_analyzer/internal/filenamer"
	internalparser "main/analyzer_plugin/project_analyzer/internal/parser"
	"main/analyzer_plugin/project_analyzer/internal/writer"

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
	return ar, nil
}

// AnalyzeProject 是为 `analyze` 命令提供的处理器。
// 它执行完整的项目分析，并将结果以过滤和序列化后的JSON格式写入文件。
func AnalyzeProject(rootPath string, outputDir string, ignore []string, isMonorepo bool) {
	// 步骤 1: parser解析项目。
	ar, err := internalparser.ParseProject(rootPath, ignore, isMonorepo)
	if err != nil {
		// 可能需要更优雅的错误处理，比如返回错误给调用者。
		fmt.Printf("解析项目失败: %s", err)
		return
	}

	// 步骤 2: 在序列化之前，将完整的分析结果转换为一个更简洁、过滤后的版本。
	filteredResult := toFilteredResult(ar)

	// 步骤 3: 使用新的 filenamer 包生成标准化的输出文件名。
	outputFileName := filenamer.GenerateOutputFileName(rootPath, "analyze")

	// 步骤 4: 使用新的 writer 包将结果写入文件。
	err = writer.WriteJSONResult(outputDir, outputFileName, filteredResult)
	if err != nil {
		fmt.Printf("写入分析结果失败: %s", err)
		return
	}
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
			AnyDeclarations: jsData.AnyDeclarations,
		}
	}

	return &FilteredProjectParserResult{
		Js_Data:      filteredJsData,
		Package_Data: ar.Package_Data,
	}
}
