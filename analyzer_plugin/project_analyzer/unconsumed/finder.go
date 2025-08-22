// package unconsumed 实现了查找项目中已导出但未被消费的变量的核心业务逻辑。
package unconsumed

import (
	"fmt"
	"main/analyzer/projectParser"
	"main/analyzer_plugin/project_analyzer"
	"strings"
)

// as export 别名结构体，用于追踪二次导出
type alias struct {
	OrigPath string // 原始文件路径
	OrigName string // 原始导出名
}

// Find 在指定的项目中查找所有已导出但从未被任何其他文件导入的变量、函数、类型等。
// V2版本改进了算法，能够处理二次导出和JSX组件消费，大大提高了准确性。
func Find(params Params) (*Result, error) {
	// 步骤 1: 分析整个项目，获取所有文件的详细AST信息。
	analyzer := project_analyzer.NewProjectAnalyzer(params.RootPath, params.Exclude, params.IsMonorepo)
	deps, err := analyzer.Analyze()
	if err != nil {
		return nil, fmt.Errorf("分析项目失败: %w", err)
	}

	// 步骤 2: 构建“消费记录”和“别名记录”
	// consumedExports 的键格式: "<文件路径>#<变量名>" (e.g., "/path/to/utils.ts#isEmpty")
	// exportAliases 的键格式: "<文件路径>#<别名>" (e.g., "/path/to/index.ts#Button")
	consumedExports := make(map[string]bool)
	exportAliases := make(map[string]alias)

	for filePath, fileData := range deps.Js_Data {
		// 从导入中收集消费记录
		for _, imp := range fileData.ImportDeclarations {
			if imp.Source.FilePath == "" { // 忽略NPM包
				continue
			}
			for _, module := range imp.ImportModules {
				key := ""
				if module.Type == "default" || module.Type == "namespace" || module.Type == "dynamic_variable" {
					key = fmt.Sprintf("%s#*", imp.Source.FilePath) // `*` 代表消费了整个模块，特别是默认导出
				} else {
					key = fmt.Sprintf("%s#%s", imp.Source.FilePath, module.Identifier)
				}
				consumedExports[key] = true
			}
		}

		// 从JSX使用中收集消费记录
		for _, jsx := range fileData.JsxElements {
			if jsx.Source.FilePath != "" {
				// JSX组件的消费通常是消费其默认导出
				key := fmt.Sprintf("%s#*", jsx.Source.FilePath)
				consumedExports[key] = true
			}
		}

		// 从二次导出中收集别名记录 (`export { a as b } from './c'`)
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

	// 步骤 2.5: 解析别名，将间接消费转换为直接消费
	for consumedKey := range consumedExports {
		if alias, ok := exportAliases[consumedKey]; ok {
			// 如果消费的是一个别名，则将原始导出标记为已消费
			finalKey := fmt.Sprintf("%s#%s", alias.OrigPath, alias.OrigName)
			consumedExports[finalKey] = true
		}
	}

	// 步骤 3: 遍历所有直接导出项，检查它们是否在“消费记录”中。
	var unconsumedExports []Export
	totalExportsFound := 0

	for filePath, fileData := range deps.Js_Data {
		if isIgnoredFile(filePath) {
			continue
		}

		// 检查 `export { ... }`
		for _, exp := range fileData.ExportDeclarations {
			if exp.Source != nil { // 只处理直接导出，忽略 `export ... from ...`
				continue
			}
			for _, module := range exp.ExportModules {
				totalExportsFound++
				key := fmt.Sprintf("%s#%s", filePath, module.Identifier)
				if !consumedExports[key] {
					unconsumedExports = append(unconsumedExports, Export{
						FilePath:   filePath,
						ExportName: module.Identifier,
						Line:       0, // FIXME: 解析器未提供此导出类型的行号
						Kind:       string(module.Type),
					})
				}
			}
		}

		// 检查 `export const/var/let/function/class ...`
		addUnconsumedFromDeclarations(&unconsumedExports, &totalExportsFound, filePath, fileData, consumedExports)

		// 检查 `export default ...`
		for _, assign := range fileData.ExportAssignments {
			totalExportsFound++
			key := fmt.Sprintf("%s#*", filePath) // 默认导出被消费，通常是整个模块被消费
			if !consumedExports[key] {
				unconsumedExports = append(unconsumedExports, Export{
					FilePath:   filePath,
					ExportName: "default",
					Line:       assign.SourceLocation.Start.Line,
					Kind:       "default",
				})
			}
		}
	}

	// 步骤 4: 组装最终结果。
	result := &Result{
		UnconsumedExports: unconsumedExports,
		Summary: SummaryStats{
			TotalFilesScanned:      len(deps.Js_Data),
			TotalExportsFound:      totalExportsFound,
			UnconsumedExportsFound: len(unconsumedExports),
		},
	}

	return result, nil
}

// isIgnoredFile 是一个辅助函数，判断文件是否应该在分析中被忽略。
func isIgnoredFile(filePath string) bool {
	return strings.Contains(filePath, ".test.") ||
		strings.Contains(filePath, ".spec.") ||
		strings.HasSuffix(filePath, ".d.ts") ||
		strings.Contains(filePath, "__tests__") ||
		strings.Contains(filePath, "__mocks__")
}

// addUnconsumedFromDeclarations 检查通过变量、接口、枚举、类型等声明导出的实体。
func addUnconsumedFromDeclarations(unconsumedExports *[]Export, totalExportsFound *int, filePath string, fileData projectParser.JsFileParserResult, consumedExports map[string]bool) {
	// 检查 `export var/let/const ...`
	for _, v := range fileData.VariableDeclarations {
		if !v.Exported {
			continue
		}
		for _, declarator := range v.Declarators {
			*totalExportsFound++
			key := fmt.Sprintf("%s#%s", filePath, declarator.Identifier)
			if !consumedExports[key] {
				*unconsumedExports = append(*unconsumedExports, Export{
					FilePath:   filePath,
					ExportName: declarator.Identifier,
					Line:       v.SourceLocation.Start.Line,
					Kind:       string(v.Kind),
				})
			}
		}
	}

	// 检查 `export interface ...`
	for identifier, decl := range fileData.InterfaceDeclarations {
		if !decl.Exported {
			continue
		}
		*totalExportsFound++
		key := fmt.Sprintf("%s#%s", filePath, identifier)
		if !consumedExports[key] {
			*unconsumedExports = append(*unconsumedExports, Export{
				FilePath:   filePath,
				ExportName: identifier,
				Line:       decl.SourceLocation.Start.Line,
				Kind:       "interface",
			})
		}
	}

	// 检查 `export enum ...`
	for identifier, decl := range fileData.EnumDeclarations {
		if !decl.Exported {
			continue
		}
		*totalExportsFound++
		key := fmt.Sprintf("%s#%s", filePath, identifier)
		if !consumedExports[key] {
			*unconsumedExports = append(*unconsumedExports, Export{
				FilePath:   filePath,
				ExportName: identifier,
				Line:       decl.SourceLocation.Start.Line,
				Kind:       "enum",
			})
		}
	}

	// 检查 `export type ...`
	for identifier, decl := range fileData.TypeDeclarations {
		if !decl.Exported {
			continue
		}
		*totalExportsFound++
		key := fmt.Sprintf("%s#%s", filePath, identifier)
		if !consumedExports[key] {
			*unconsumedExports = append(*unconsumedExports, Export{
				FilePath:   filePath,
				ExportName: identifier,
				Line:       decl.SourceLocation.Start.Line,
				Kind:       "type",
			})
		}
	}
}
