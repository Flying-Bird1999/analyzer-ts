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

type ProjectAnalyzer struct {
	rootPath   string
	ignore     []string
	isMonorepo bool
}

func NewProjectAnalyzer(rootPath string, ignore []string, isMonorepo bool) *ProjectAnalyzer {
	return &ProjectAnalyzer{
		rootPath:   rootPath,
		ignore:     ignore,
		isMonorepo: isMonorepo,
	}
}

func (pa *ProjectAnalyzer) Analyze() (*projectParser.ProjectParserResult, error) {
	config := projectParser.NewProjectParserConfig(pa.rootPath, pa.ignore, pa.isMonorepo)
	ar := projectParser.NewProjectParserResult(config)
	ar.ProjectParser()
	return ar, nil
}

// ... (rest of the file remains the same)

// Filtered structs definition

type FilteredInterfaceDeclarationResult struct {
	Identifier string                          `json:"identifier"`
	Reference  map[string]parser.TypeReference `json:"reference"`
}

type FilteredTypeDeclarationResult struct {
	Identifier string                          `json:"identifier"`
	Reference  map[string]parser.TypeReference `json:"reference"`
}

type FilteredEnumDeclarationResult struct {
	Identifier string `json:"identifier"`
}

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

type FilteredJsFileParserResult struct {
	ImportDeclarations []FilteredImportDeclaration      `json:"importDeclarations"`
	ExportDeclarations []FilteredExportDeclaration      `json:"exportDeclarations"`
	ExportAssignments  []FilteredExportAssignmentResult `json:"exportAssignments"`
	// InterfaceDeclarations map[string]FilteredInterfaceDeclarationResult `json:"interfaceDeclarations"`
	// TypeDeclarations      map[string]FilteredTypeDeclarationResult      `json:"typeDeclarations"`
	// EnumDeclarations      map[string]FilteredEnumDeclarationResult      `json:"enumDeclarations"`
	VariableDeclarations []FilteredVariableDeclaration `json:"variableDeclarations"`
	CallExpressions      []FilteredCallExpression      `json:"callExpressions"`
	JsxElements          []FilteredJSXElement          `json:"jsxElements"`
}

type FilteredProjectParserResult struct {
	Js_Data      map[string]FilteredJsFileParserResult                `json:"js_Data"`
	Package_Data map[string]projectParser.PackageJsonFileParserResult `json:"package_Data"`
}

// Conversion function
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
			// InterfaceDeclarations: lo.MapValues(jsData.InterfaceDeclarations, func(inter parser.InterfaceDeclarationResult, _ string) FilteredInterfaceDeclarationResult {
			// 	return FilteredInterfaceDeclarationResult{Identifier: inter.Identifier, Reference: inter.Reference}
			// }),
			// TypeDeclarations: lo.MapValues(jsData.TypeDeclarations, func(typeDecl parser.TypeDeclarationResult, _ string) FilteredTypeDeclarationResult {
			// 	return FilteredTypeDeclarationResult{Identifier: typeDecl.Identifier, Reference: typeDecl.Reference}
			// }),
			// EnumDeclarations: lo.MapValues(jsData.EnumDeclarations, func(enumDecl parser.EnumDeclarationResult, _ string) FilteredEnumDeclarationResult {
			// 	return FilteredEnumDeclarationResult{Identifier: enumDecl.Identifier}
			// }),
			VariableDeclarations: lo.Map(jsData.VariableDeclarations, func(decl parser.VariableDeclaration, _ int) FilteredVariableDeclaration {
				return FilteredVariableDeclaration{Exported: decl.Exported, Kind: decl.Kind, Source: decl.Source, Declarators: decl.Declarators}
			}),
			CallExpressions: lo.Map(jsData.CallExpressions, func(expr parser.CallExpression, _ int) FilteredCallExpression {
				return FilteredCallExpression{CallChain: expr.CallChain, Arguments: expr.Arguments, Type: expr.Type}
			}),
			JsxElements: lo.Map(jsData.JsxElements, func(elem parser.JSXElement, _ int) FilteredJSXElement {
				return FilteredJSXElement{ComponentChain: elem.ComponentChain, Attrs: elem.Attrs}
			}),
		}
	}

	return &FilteredProjectParserResult{
		Js_Data:      filteredJsData,
		Package_Data: ar.Package_Data,
	}
}

func parseProject(rootPath string, ignore []string, isMonorepo bool) *projectParser.ProjectParserResult {
	config := projectParser.NewProjectParserConfig(rootPath, ignore, isMonorepo)
	ar := projectParser.NewProjectParserResult(config)
	ar.ProjectParser()
	return ar
}

func AnalyzeProject(rootPath string, outputDir string, ignore []string, isMonorepo bool) {
	ar := parseProject(rootPath, ignore, isMonorepo)

	// Convert to filtered result before marshalling
	filteredResult := toFilteredResult(ar)

	jsonData, err := json.MarshalIndent(filteredResult, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling to JSON: %s\n", err)
		return
	}

	// Write to file
	outputFile := filepath.Join(outputDir, filepath.Base(rootPath)+".json")
	err = os.WriteFile(outputFile, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing JSON to file: %s\n", err)
		return
	}

	fmt.Printf("分析结果已写入文件: %s\n", outputFile)
}

type ImplicitDependency struct {
	Name     string `json:"name"`
	FilePath string `json:"filePath"`
	Raw      string `json:"raw"`
}

func FindImplicitDependencies(rootPath string, ignore []string, isMonorepo bool) []ImplicitDependency {
	ar := parseProject(rootPath, ignore, isMonorepo)

	declaredDependencies := make(map[string]bool)
	for _, pkgData := range ar.Package_Data {
		for _, dep := range pkgData.NpmList {
			declaredDependencies[dep.Name] = true
		}
	}

	implicitDependencies := []ImplicitDependency{}
	for path, jsData := range ar.Js_Data {
		for _, imp := range jsData.ImportDeclarations {
			if imp.Source.Type == "npm" {
				if !declaredDependencies[imp.Source.NpmPkg] {
					implicitDependencies = append(implicitDependencies, ImplicitDependency{
						Name:     imp.Source.NpmPkg,
						FilePath: path,
						Raw:      imp.Raw,
					})
				}
			}
		}
	}

	return implicitDependencies
}