// package structuresimple 包含一个分析器，用于生成项目整体结构的简化版报告。
package structuresimple

import (
	"fmt"
	"main/analyzer/parser"
	"main/analyzer/projectParser"
	projectanalyzer "main/analyzer_plugin/project_analyzer"

	"github.com/samber/lo"
)

// --- Analyzer Implementation ---

// StructureSimpleAnalyzer 实现了 Analyzer 接口，用于生成简化的项目结构报告。
type StructureSimpleAnalyzer struct{}

func (s *StructureSimpleAnalyzer) Name() string {
	return "structure-simple"
}

func (s *StructureSimpleAnalyzer) Configure(params map[string]string) error {
	// 本分析器无需配置
	return nil
}

func (s *StructureSimpleAnalyzer) Analyze(ctx *projectanalyzer.ProjectContext) (projectanalyzer.Result, error) {
	// 直接使用上下文中的完整解析结果，并将其转换为过滤后的简化版本
	filteredResult := toFilteredResult(ctx.ParsingResult)
	return &StructureSimpleResult{Data: filteredResult}, nil
}

// --- Result Implementation ---

// StructureSimpleResult 存储了简化的项目结构数据。
type StructureSimpleResult struct {
	Data *FilteredProjectParserResult
}

func (r *StructureSimpleResult) Name() string {
	return "structure-simple"
}

func (r *StructureSimpleResult) Summary() string {
	return fmt.Sprintf("Processed %d JS/TS files and %d package.json files.", len(r.Data.Js_Data), len(r.Data.Package_Data))
}

func (r *StructureSimpleResult) ToJSON(indent bool) ([]byte, error) {
	// 直接返回过滤后的数据作为 JSON
	return projectanalyzer.ToJSONBytes(r.Data, indent)
}

func (r *StructureSimpleResult) ToConsole() string {
	// 对于如此复杂的结果，控制台输出仅显示摘要
	return r.Summary()
}

// --- Filtering Logic and Types ---

// toFilteredResult 是一个转换函数，用于将完整的、包含大量原始信息的分析结果 (ProjectParserResult)
// 转换为一个更简洁、过滤后的版本 (FilteredProjectParserResult)，以供 `analyze` 命令输出。
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

// --- Filtered Types ---

type FilteredVariableDeclaration struct {
	Exported    bool                         `json:"exported"`
	Kind        parser.DeclarationKind       `json:"kind"`
	Source      *parser.VariableValue        `json:"source,omitempty"`
	Declarators []*parser.VariableDeclarator `json:"declarators"`
}

type FilteredCallExpression struct {
	CallChain []string          `json:"callChain"`
	Arguments []parser.Argument `json:"arguments"`
	Type      string            `json:"type"`
}

type FilteredJSXElement struct {
	ComponentChain []string              `json:"componentChain"`
	Attrs          []parser.JSXAttribute `json:"attrs"`
}

type FilteredExportAssignmentResult struct {
	Expression string `json:"expression"`
}

type FilteredImportDeclaration struct {
	ImportModules []parser.ImportModule    `json:"importModules"`
	Source        projectParser.SourceData `json:"source"`
}

type FilteredExportDeclaration struct {
	ExportModules []parser.ExportModule     `json:"exportModules"`
	Source        *projectParser.SourceData `json:"source,omitempty"`
}

type FilteredFunctionDeclarationResult struct {
	Exported   bool                     `json:"exported"`
	Identifier string                   `json:"identifier"`
	IsAsync    bool                     `json:"isAsync"`
	Parameters []parser.ParameterResult `json:"parameters"`
	ReturnType string                   `json:"returnType"`
}

type FilteredJsFileParserResult struct {
	ImportDeclarations   []FilteredImportDeclaration         `json:"importDeclarations"`
	ExportDeclarations   []FilteredExportDeclaration         `json:"exportDeclarations"`
	ExportAssignments    []FilteredExportAssignmentResult    `json:"exportAssignments"`
	VariableDeclarations []FilteredVariableDeclaration       `json:"variableDeclarations"`
	CallExpressions      []FilteredCallExpression            `json:"callExpressions"`
	JsxElements          []FilteredJSXElement                `json:"jsxElements"`
	FunctionDeclarations []FilteredFunctionDeclarationResult `json:"functionDeclarations"`
	AnyDeclarations      []parser.AnyInfo                    `json:"anyDeclarations"`
}

type FilteredProjectParserResult struct {
	Js_Data      map[string]FilteredJsFileParserResult                `json:"js_Data"`
	Package_Data map[string]projectParser.PackageJsonFileParserResult `json:"package_Data"`
}
