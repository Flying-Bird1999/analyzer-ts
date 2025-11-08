// Package unconsumed 实现了查找项目中已导出但未被消费的变量的核心业务逻辑。
//
// 功能说明：
// 这个分析器专门用于检测 TypeScript 项目中的"死导出"——即那些被导出但在其他任何地方
// 都没有被使用的符号。这些未使用的导出会增加代码包的大小，影响项目的维护性，
// 并且可能表明存在遗留的、不再需要的代码。
//
// 主要用途：
// 1. 代码清理：识别和移除未使用的导出，减少包体积
// 2. 维护性提升：清理死代码，提高代码库的可维护性
// 3. 依赖关系分析：了解项目内部的实际使用关系
// 4. 重构支持：在进行大型重构前识别可能受影响的部分
//
// 支持的导出类型：
// - 函数和函数声明（function declarations 和 function expressions）
// - 变量声明（var, const, let）
// - 类声明（class declarations）
// - 接口声明（interface declarations）
// - 类型声明（type declarations）
// - 枚举声明（enum declarations）
// - 默认导出（default exports）
// - 命名导出（named exports）
// - 重导出（re-exports 和 export aliases）
//
// 实现特点：
// - 智能排除：自动忽略测试文件、类型定义文件等
// - 重导出追踪：能够追踪通过 export from 语法的二次导出
// - JSX 支持：能够识别 React 组件的导入使用
// - 详细信息：提供每个未使用导出的位置和类型信息
package unconsumed

import (
	"fmt"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

// Finder 是"未消费导出"分析器的实现。
// 这个分析器会遍历项目中所有的导出项，并检查它们是否在其他文件中被实际使用。
//
// 工作原理：
// 1. 第一阶段：收集所有项目中实际被使用的导出项（通过分析 import 语句）
// 2. 第二阶段：收集所有文件中声明的导出项
// 3. 第三阶段：对比两个集合，找出差异即为未使用的导出
// 4. 第四阶段：生成包含详细位置和类型信息的分析报告
type Finder struct{}

// 确保 Finder 实现了 projectanalyzer.Analyzer 接口
var _ projectanalyzer.Analyzer = (*Finder)(nil)

// Name 返回分析器的唯一标识符。
// 注意：这里返回的是内部名称，实际注册在 analyze.go 中使用 "unconsumed"
func (f *Finder) Name() string {
	return "unconsumed-exports-finder"
}

// Configure 配置分析器的运行参数。
// 由于 unconsumed 分析器不需要任何配置参数，这个方法直接返回 nil。
func (f *Finder) Configure(params map[string]string) error {
	// 该分析器不需要任何配置参数
	return nil
}

// alias 结构体用于追踪导出别名信息。
// 当一个文件使用 `export { OriginalName as NewName } from './module'` 语法时，
// 我们需要记录这个映射关系，以便正确追踪原始导出的使用情况。
type alias struct {
	OrigPath string // 原始文件路径，即原始导出所在的文件
	OrigName string // 原始导出名，即被重新导出的原始符号名称
}

// Analyze 执行核心的分析逻辑。
// 这个方法是整个分析器的核心，通过四阶段的算法来识别未使用的导出项。
//
// 分析算法说明：
//
// 第一阶段：收集被消费的导出项
// 遍历所有文件的导入语句，记录哪些导出项被实际使用了：
// - 普通命名导入：import { name } from 'module'
// - 默认导入：import def from 'module'
// - 命名空间导入：import * as ns from 'module'
// - JSX 组件导入：<Component /> 会隐式导入默认导出
//
// 第二阶段：收集重导出映射关系
// 处理 `export { name } from 'module'` 语法，建立别名映射关系，确保能正确追踪
// 重新导出的符号的使用情况。
//
// 第三阶段：解析重导出关系
// 根据第二阶段收集的映射关系，将被重导出的符号标记为已消费。
//
// 第四阶段：识别未消费的导出项
// 对比所有导出项和已消费导出项的集合，找出差异。
//
// 参数说明：
// - ctx: 项目上下文，包含完整的解析结果
//
// 返回值说明：
// - projectanalyzer.Result: 包含未使用导出分析结果的对象
// - error: 分析过程中出现的错误（通常不会出错）
func (f *Finder) Analyze(ctx *projectanalyzer.ProjectContext) (projectanalyzer.Result, error) {
	// 获取项目解析结果
	deps := ctx.ParsingResult

	// === 第一阶段：收集被消费的导出项 ===
	// key 格式：文件路径#导出名，例如："/src/utils.ts#formatDate"
	// 默认导出使用 "*" 作为导出名："/src/utils.ts#*"
	consumedExports := make(map[string]bool)

	// === 第二阶段：收集重导出映射关系 ===
	// 记录 export { name as newName } from 'module' 的映射关系
	exportAliases := make(map[string]alias)

	// 遍历所有已解析的文件，收集导入和重导出信息
	for filePath, fileData := range deps.Js_Data {
		// 处理所有导入声明
		for _, imp := range fileData.ImportDeclarations {
			// 跳过无效的导入声明
			if imp.Source.FilePath == "" {
				continue
			}

			// 处理导入的各个模块
			for _, module := range imp.ImportModules {
				key := ""
				// 根据导入类型构建不同的 key
				if module.Type == "default" || module.Type == "namespace" || module.Type == "dynamic_variable" {
					// 默认导入、命名空间导入等特殊类型
					key = fmt.Sprintf("%s#*", imp.Source.FilePath)
				} else {
					// 普通命名导入
					key = fmt.Sprintf("%s#%s", imp.Source.FilePath, module.Identifier)
				}
				consumedExports[key] = true
			}
		}

		// 处理 JSX 组件导入
		// 当文件包含 JSX 元素时，会隐式导入 React 组件的默认导出
		for _, jsx := range fileData.JsxElements {
			if jsx.Source.FilePath != "" {
				key := fmt.Sprintf("%s#*", jsx.Source.FilePath)
				consumedExports[key] = true
			}
		}

		// 处理重导出声明：export { name } from 'module'
		for _, exp := range fileData.ExportDeclarations {
			// 只处理有来源的导出声明（即重导出）
			if exp.Source == nil || exp.Source.FilePath == "" {
				continue
			}

			// 建立重导出映射关系
			for _, module := range exp.ExportModules {
				aliasKey := fmt.Sprintf("%s#%s", filePath, module.Identifier)
				// 确定原始导出名称
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

	// === 第三阶段：解析重导出关系 ===
	// 将被重导出的符号也标记为已消费
	for consumedKey := range consumedExports {
		if alias, ok := exportAliases[consumedKey]; ok {
			finalKey := fmt.Sprintf("%s#%s", alias.OrigPath, alias.OrigName)
			consumedExports[finalKey] = true
		}
	}

	// === 第四阶段：识别未消费的导出项 ===
	var findings []Finding
	totalExportsFound := 0

	// 遍历所有文件，检查导出项是否被消费
	for filePath, fileData := range deps.Js_Data {
		// 跳过测试文件和类型定义文件等
		if isIgnoredFile(filePath) {
			continue
		}

		// 处理 export { name } 语法（直接导出，非重导出）
		for _, exp := range fileData.ExportDeclarations {
			// 只处理本地的直接导出
			if exp.Source != nil {
				continue
			}

			// 检查每个导出项是否被消费
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

		// 处理各种声明语句的导出（函数、变量、类型等）
		addUnconsumedFromDeclarations(&findings, &totalExportsFound, filePath, fileData, consumedExports)

		// 处理默认导出：export default ...
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

	// 构建最终的分析结果
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

// isIgnoredFile 判断是否应该忽略某个文件。
// 这些文件通常不应该被分析，因为它们：
// 1. 是测试文件，通常不会被生产代码引用
// 2. 是类型定义文件，通常用于提供类型而非实际功能
// 3. 是测试工具相关文件，用于测试环境
//
// 忽略的文件类型：
// - *.test.* 和 *.spec.*：测试文件
// - *.d.ts：TypeScript 类型定义文件
// - __tests__ 目录：Jest 测试目录
// - __mocks__ 目录：模拟文件目录
//
// 参数说明：
// - filePath: 需要判断的文件路径
//
// 返回值说明：
// - bool: 如果文件应该被忽略返回 true，否则返回 false
func isIgnoredFile(filePath string) bool {
	return strings.Contains(filePath, ".test.") ||
		strings.Contains(filePath, ".spec.") ||
		strings.HasSuffix(filePath, ".d.ts") ||
		strings.Contains(filePath, "__tests__") ||
		strings.Contains(filePath, "__mocks__")
}

// addUnconsumedFromDeclarations 从各种声明语句中查找未使用的导出项。
// 这个函数专门处理通过声明语句导出的符号，包括变量、函数、接口、枚举和类型声明。
//
// 支持的声明类型：
// 1. 变量声明（var, const, let）
// 2. 接口声明（interface）
// 3. 枚举声明（enum）
// 4. 类型别名声明（type）
// 5. 函数声明（function）- 通过 VariableDeclarations 处理
//
// 参数说明：
// - findings: 用于存储找到的未使用导出项的切片引用
// - totalExportsFound: 用于统计找到的导出项总数的引用
// - filePath: 当前分析的文件路径
// - fileData: 当前文件的解析结果数据
// - consumedExports: 已消费导出项的映射，用于判断是否被使用
func addUnconsumedFromDeclarations(findings *[]Finding, totalExportsFound *int, filePath string, fileData projectParser.JsFileParserResult, consumedExports map[string]bool) {
	// === 处理变量声明 ===
	// 包括 const, let, var 声明，以及函数声明（function declarations）
	for _, v := range fileData.VariableDeclarations {
		// 只处理被导出的变量声明
		if !v.Exported {
			continue
		}

		// 处理变量声明中的每个声明符（支持多重声明）
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

	// === 处理接口声明 ===
	for identifier, decl := range fileData.InterfaceDeclarations {
		// 只处理被导出的接口
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

	// === 处理枚举声明 ===
	for identifier, decl := range fileData.EnumDeclarations {
		// 只处理被导出的枚举
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

	// === 处理类型别名声明 ===
	for identifier, decl := range fileData.TypeDeclarations {
		// 只处理被导出的类型别名
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
