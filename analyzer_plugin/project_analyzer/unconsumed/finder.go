// package unconsumed 实现了查找项目中已导出但未被消费的变量的核心业务逻辑。
package unconsumed

import (
	"fmt"
	"main/analyzer/projectParser"
	projectanalyzer "main/analyzer_plugin/project_analyzer"
	"strings"
)

// Finder 是“未消费导出”分析器的实现。
type Finder struct{}

var _ projectanalyzer.Analyzer = (*Finder)(nil)

func (f *Finder) Name() string {
	return "unconsumed-exports-finder"
}

func (f *Finder) Configure(params map[string]string) error {
	return nil
}

// as export 别名结构体，用于追踪二次导出
type alias struct {
	OrigPath string // 原始文件路径
	OrigName string // 原始导出名
}

func (f *Finder) Analyze(ctx *projectanalyzer.ProjectContext) (projectanalyzer.Result, error) {
	deps := ctx.ParsingResult

	consumedExports := make(map[string]bool)
	exportAliases := make(map[string]alias)

	for filePath, fileData := range deps.Js_Data {
		for _, imp := range fileData.ImportDeclarations {
			if imp.Source.FilePath == "" {
				continue
			}
			for _, module := range imp.ImportModules {
				key := ""
				if module.Type == "default" || module.Type == "namespace" || module.Type == "dynamic_variable" {
					key = fmt.Sprintf("%s#*", imp.Source.FilePath)
				} else {
					key = fmt.Sprintf("%s#%s", imp.Source.FilePath, module.Identifier)
				}
				consumedExports[key] = true
			}
		}

		for _, jsx := range fileData.JsxElements {
			if jsx.Source.FilePath != "" {
				key := fmt.Sprintf("%s#*", jsx.Source.FilePath)
				consumedExports[key] = true
			}
		}

		for _, exp := range fileData.ExportDeclarations {
			if exp.Source == nil || exp.Source.FilePath == "" {
				continue
			}
			for _, module := range exp.ExportModules {
				aliasKey := fmt.Sprintf("%s#%s", filePath, module.Identifier)
				originalName := module.ModuleName
				if originalName == "" {
					originalName = module.Identifier
				}
				exportAliases[aliasKey] = alias{
					OrigPath: exp.Source.FilePath,
					OrigName: originalName,
				}
			}
		}
	}

	for consumedKey := range consumedExports {
		if alias, ok := exportAliases[consumedKey]; ok {
			finalKey := fmt.Sprintf("%s#%s", alias.OrigPath, alias.OrigName)
			consumedExports[finalKey] = true
		}
	}

	var findings []Finding
	totalExportsFound := 0

	for filePath, fileData := range deps.Js_Data {
		if isIgnoredFile(filePath) {
			continue
		}

		for _, exp := range fileData.ExportDeclarations {
			if exp.Source != nil {
				continue
			}
			for _, module := range exp.ExportModules {
				totalExportsFound++
				key := fmt.Sprintf("%s#%s", filePath, module.Identifier)
				if !consumedExports[key] {
					findings = append(findings, Finding{
						FilePath:   filePath,
						ExportName: module.Identifier,
						Line:       0,
						Kind:       string(module.Type),
					})
				}
			}
		}

		addUnconsumedFromDeclarations(&findings, &totalExportsFound, filePath, fileData, consumedExports)

		for _, assign := range fileData.ExportAssignments {
			totalExportsFound++
			key := fmt.Sprintf("%s#*", filePath)
			if !consumedExports[key] {
				findings = append(findings, Finding{
					FilePath:   filePath,
					ExportName: "default",
					Line:       assign.SourceLocation.Start.Line,
					Kind:       "default",
				})
			}
		}
	}

	finalResult := &Result{
		Findings: findings,
		Stats: SummaryStats{
			TotalFilesScanned:      len(deps.Js_Data),
			TotalExportsFound:      totalExportsFound,
			UnconsumedExportsFound: len(findings),
		},
	}

	return finalResult, nil
}

func isIgnoredFile(filePath string) bool {
	return strings.Contains(filePath, ".test.") ||
		strings.Contains(filePath, ".spec.") ||
		strings.HasSuffix(filePath, ".d.ts") ||
		strings.Contains(filePath, "__tests__") ||
		strings.Contains(filePath, "__mocks__")
}

func addUnconsumedFromDeclarations(findings *[]Finding, totalExportsFound *int, filePath string, fileData projectParser.JsFileParserResult, consumedExports map[string]bool) {
	for _, v := range fileData.VariableDeclarations {
		if !v.Exported {
			continue
		}
		for _, declarator := range v.Declarators {
			*totalExportsFound++
			key := fmt.Sprintf("%s#%s", filePath, declarator.Identifier)
			if !consumedExports[key] {
				*findings = append(*findings, Finding{
					FilePath:   filePath,
					ExportName: declarator.Identifier,
					Line:       v.SourceLocation.Start.Line,
					Kind:       string(v.Kind),
				})
			}
		}
	}

	for identifier, decl := range fileData.InterfaceDeclarations {
		if !decl.Exported {
			continue
		}
		*totalExportsFound++
		key := fmt.Sprintf("%s#%s", filePath, identifier)
		if !consumedExports[key] {
			*findings = append(*findings, Finding{
				FilePath:   filePath,
				ExportName: identifier,
				Line:       decl.SourceLocation.Start.Line,
				Kind:       "interface",
			})
		}
	}

	for identifier, decl := range fileData.EnumDeclarations {
		if !decl.Exported {
			continue
		}
		*totalExportsFound++
		key := fmt.Sprintf("%s#%s", filePath, identifier)
		if !consumedExports[key] {
			*findings = append(*findings, Finding{
				FilePath:   filePath,
				ExportName: identifier,
				Line:       decl.SourceLocation.Start.Line,
				Kind:       "enum",
			})
		}
	}

	for identifier, decl := range fileData.TypeDeclarations {
		if !decl.Exported {
			continue
		}
		*totalExportsFound++
		key := fmt.Sprintf("%s#%s", filePath, identifier)
		if !consumedExports[key] {
			*findings = append(*findings, Finding{
				FilePath:   filePath,
				ExportName: identifier,
				Line:       decl.SourceLocation.Start.Line,
				Kind:       "type",
			})
		}
	}
}
