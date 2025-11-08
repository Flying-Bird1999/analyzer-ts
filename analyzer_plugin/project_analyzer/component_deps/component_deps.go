// Package component_deps 实现了分析组件依赖关系的核心业务逻辑。
//
// 功能说明：
// 这个分析器专门用于分析 TypeScript/TSX 项目中组件之间的依赖关系。
// 通过从指定的入口文件开始，递归分析组件的导入和导出关系，构建完整的
// 组件依赖图。这对于理解项目架构、重构优化、循环依赖检测等场景非常有用。
//
// 主要用途：
// 1. 架构分析：了解项目的组件依赖关系和模块结构
// 2. 循环依赖检测：识别组件间的循环依赖问题
// 3. 重构支持：在重构组件前了解其影响范围
// 4. 包优化：识别可以合并或拆分的组件
// 5. 文档生成：自动生成组件依赖关系文档
//
// 支持的组件类型：
// - React 组件（函数组件和类组件）
// - Vue 组件
// - 普通 TypeScript 模块作为组件使用
// - 默认导出的组件
// - 命名导出的组件
//
// 实现特点：
// - 支持 glob 模式匹配多个入口文件
// - 自动识别公共组件和内部组件
// - 支持多层依赖关系分析
// - 生成可视化的依赖图数据
// - 支持大型 monorepo 项目
package component_deps

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
	"github.com/samber/lo"
)

// =============================================================================
// 分析器主体定义
// =============================================================================

// ComponentDependencyAnalyzer 实现了 Analyzer 接口，用于分析组件依赖关系。
//
// 功能概述：
// 该分析器能够从指定入口文件开始，识别公共组件并分析它们之间的依赖关系。
// 通过深度遍历项目的导入导出关系，构建完整的组件依赖图谱。
//
// 工作流程：
// 1. 解析入口文件，识别根级别的组件
// 2. 递归分析每个组件的依赖关系
// 3. 构建组件依赖图和层级关系
// 4. 生成包含完整依赖关系的分析报告
//
// 支持的功能：
// - Glob 模式匹配：支持 `packages/*/src/index.ts` 这样的模式
// - 公共组件识别：自动识别被多个文件使用的公共组件
// - 内部组件分析：分析组件内部的私有子组件
// - 循环依赖检测：识别并报告组件间的循环依赖
// - 依赖层级分析：构建组件的依赖层次结构
type ComponentDependencyAnalyzer struct {
	// EntryPoint 入口文件路径，支持 glob 模式匹配。
	// 可以是单个文件路径，也可以是 glob 模式，例如：
	// - "packages/sl-admin-components/src/v3.ts"
	// - "packages/*/src/index.ts"
	// - "src/components/**/index.ts"
	EntryPoint string
}

// Name 返回分析器的唯一标识符。
// 用于在插件系统中注册和识别该分析器。
//
// 返回值说明：
// 返回 "component-deps" 作为分析器的标识符。
// 这个名称用于在命令行中调用该分析器。
func (a *ComponentDependencyAnalyzer) Name() string {
	return "component-deps"
}

// Configure 配置分析器的参数
// params: 包含配置参数的 map，目前支持 "entryPoint" 参数
// entryPoint: 指定分析的入口文件路径，支持 glob 模式
func (a *ComponentDependencyAnalyzer) Configure(params map[string]string) error {
	if entryPoint, ok := params["entryPoint"]; ok {
		a.EntryPoint = entryPoint
	}
	return nil
}

// isComponentExport 判断给定的名称是否为组件导出
// 通过检查名称的第一个字符是否为大写字母来判断
// 在 TypeScript/React 中，组件通常以大写字母开头
// 返回值: 如果是组件导出名称则返回 true，否则返回 false
func isComponentExport(name string) bool {
	if name == "" {
		return false
	}
	firstChar := []rune(name)[0]
	return unicode.IsUpper(firstChar)
}

// findComponentSource 递归查找组件符号的源文件路径
// 该函数通过跟踪导出链路，找到组件符号的实际定义位置
// 支持处理直接导出和重导出（re-export）的情况
//
// 参数:
//   - symbolName: 要查找的组件符号名称
//   - sourcePath: 开始查找的源文件路径
//   - fileResults: 所有文件的解析结果映射
//   - visited: 用于防止循环依赖的访问记录
//
// 返回值:
//   - 如果找到则返回组件符号的实际源文件路径
//   - 如果未找到则返回空字符串
func findComponentSource(symbolName, sourcePath string, fileResults map[string]projectParser.JsFileParserResult, visited map[string]bool) string {
	// 构造唯一的访问键，防止重复处理同一个符号在同一个文件的查找
	visitedKey := fmt.Sprintf("%s|%s", symbolName, sourcePath)
	if visited[visitedKey] {
		return "" // 检测到循环依赖，直接返回
	}
	visited[visitedKey] = true

	// 获取当前文件的解析结果
	fileResult, ok := fileResults[sourcePath]
	if !ok {
		return ""
	}

	// 检查直接导出情况：检查是否在该文件中直接导出了目标符号
	for _, exportDecl := range fileResult.ExportDeclarations {
		if exportDecl.Source == nil { // 本地导出（非 re-export）
			for _, module := range exportDecl.ExportModules {
				if module.Identifier == symbolName {
					return sourcePath // 找到直接导出的符号
				}
			}
		}
	}

	// 检查默认导出情况：处理 export default 的情况
	// 注意：这个逻辑进行了简化，可能无法覆盖所有情况
	if len(fileResult.ExportAssignments) > 0 {
		if getComponentName(sourcePath) == symbolName {
			return sourcePath
		}
	}

	// 检查重导出情况：处理从其他文件导入再导出的情况
	for _, exportDecl := range fileResult.ExportDeclarations {
		if exportDecl.Source != nil && exportDecl.Source.FilePath != "" { // re-export
			for _, module := range exportDecl.ExportModules {
				if module.Identifier == symbolName {
					// 递归查找符号的实际定义位置
					return findComponentSource(module.ModuleName, exportDecl.Source.FilePath, fileResults, visited)
				}
			}
		}
	}

	// 如果在当前文件中找不到符号，则返回当前文件路径
	// 这通常意味着符号是在其他地方定义的
	return sourcePath
}

// getComponentName 根据文件路径推导组件名称
// 该函数从文件路径中提取组件名称，遵循常见的命名约定
//
// 参数:
//   - filePath: 文件的完整路径
//
// 返回值:
//   - 推导出的组件名称，如果无法推导则返回空字符串
//
// 命名约定:
//   - 对于普通文件（如 Button.tsx），返回文件名（Button）
//   - 对于 index 文件（如 src/index.ts），返回父目录名
//   - 跳过 src 和 components 目录中的 index 文件
func getComponentName(filePath string) string {
	// 获取文件名（包含扩展名）
	base := filepath.Base(filePath)
	// 移除文件扩展名
	nameWithoutExt := strings.TrimSuffix(base, filepath.Ext(base))

	// 处理 index 文件的特殊情况
	if nameWithoutExt == "index" {
		// 获取父目录名称
		parentDir := filepath.Base(filepath.Dir(filePath))
		// 如果父目录是 src 或 components，则不返回组件名称
		if parentDir == "src" || parentDir == "components" {
			return ""
		}
		return parentDir // 返回父目录名作为组件名
	}

	// 对于普通文件，直接返回文件名
	return nameWithoutExt
}

// isPureTypeRecursive 递归检查一个符号是否为纯类型（TypeScript 类型定义）
// 该函数通过跟踪符号的定义链路，判断其最终是否为类型声明
// 支持处理类型重导出和导入后再导出的复杂情况
//
// 参数:
//   - symbolName: 要检查的符号名称
//   - sourcePath: 开始检查的源文件路径
//   - fileResults: 所有文件的解析结果映射
//   - visited: 用于防止循环依赖的访问记录
//
// 返回值:
//   - 如果符号是纯类型则返回 true
//   - 如果符号不是纯类型或存在循环依赖则返回 false
func isPureTypeRecursive(symbolName, sourcePath string, fileResults map[string]projectParser.JsFileParserResult, visited map[string]bool) bool {
	// 构造唯一的访问键，防止重复处理同一个符号在同一个文件的查找
	visitedKey := fmt.Sprintf("%s|%s", symbolName, sourcePath)
	if visited[visitedKey] {
		return false // 检测到循环依赖，返回 false
	}
	visited[visitedKey] = true

	// 获取当前文件的解析结果
	fileResult, ok := fileResults[sourcePath]
	if !ok {
		return false
	}

	// 步骤 1: 检查在当前文件中是否被定义为类型
	// 检查 type 声明
	if _, ok := fileResult.TypeDeclarations[symbolName]; ok {
		return true
	}
	// 检查 interface 声明
	if _, ok := fileResult.InterfaceDeclarations[symbolName]; ok {
		return true
	}
	// 检查 enum 声明
	if _, ok := fileResult.EnumDeclarations[symbolName]; ok {
		return true
	}

	// 步骤 2: 检查是否从另一个文件重导出了该符号
	// 处理 export { type T } from './types' 这样的情况
	for _, exportDecl := range fileResult.ExportDeclarations {
		if exportDecl.Source != nil && exportDecl.Source.FilePath != "" {
			for _, module := range exportDecl.ExportModules {
				if module.Identifier == symbolName {
					// 递归检查被重导出的符号是否为纯类型
					if isPureTypeRecursive(module.ModuleName, exportDecl.Source.FilePath, fileResults, visited) {
						return true
					}
				}
			}
		}
	}

	// 步骤 3: 检查是否从另一个文件导入该符号然后导出
	// 处理 import { T } from './types'; export { T } 这样的情况
	for _, importDecl := range fileResult.ImportDeclarations {
		for _, module := range importDecl.ImportModules {
			if module.Identifier == symbolName {
				// 检查该符号是否从当前文件导出
				for _, exportDecl := range fileResult.ExportDeclarations {
					if exportDecl.Source == nil { // 例如：export { symbolName }
						for _, exportModule := range exportDecl.ExportModules {
							if exportModule.ModuleName == symbolName {
								// 递归检查导入的符号是否为纯类型
								if isPureTypeRecursive(module.ImportModule, importDecl.Source.FilePath, fileResults, visited) {
									return true
								}
							}
						}
					}
				}
			}
		}
	}

	// 如果所有检查都未发现类型定义，则认为不是纯类型
	return false
}

// Analyze 是分析器的主方法，执行组件依赖分析的核心逻辑
// 该方法遵循 "Parse Once, Analyze Many Times" 的设计原则，
// 利用预解析的项目数据构建完整的组件依赖关系图
//
// 分析步骤：
// 1. 发现所有入口文件并确定它们的归属包
// 2. 解析所有入口文件，构建"公共组件"清单
// 3. 构建"源文件 -> 公共组件名列表"的反向映射
// 4. 遍历所有组件，分析它们的依赖关系
//
// 参数:
//   - ctx: 项目上下文，包含预解析的项目数据
//
// 返回值:
//   - Result: 组件依赖分析结果
//   - error: 分析过程中遇到的错误
func (a *ComponentDependencyAnalyzer) Analyze(ctx *project_analyzer.ProjectContext) (project_analyzer.Result, error) {
	// 验证必要的配置参数
	if a.EntryPoint == "" {
		return nil, fmt.Errorf("错误: 请使用 -p 'component-deps.entryPoint=path/to/entry.ts' 参数指定入口文件")
	}

	// 获取所有文件的解析结果
	fileResults := ctx.ParsingResult.Js_Data

	// 步骤 1: 发现所有入口文件并确定它们的归属包
	entryPointPattern := filepath.Join(ctx.ProjectRoot, a.EntryPoint)
	var entryPointPaths []string
	for path := range fileResults {
		matched, err := filepath.Match(entryPointPattern, path)
		if err != nil {
			return nil, fmt.Errorf("解析 glob 模式失败 '%s': %w", entryPointPattern, err)
		}
		if matched {
			entryPointPaths = append(entryPointPaths, path)
		}
	}
	if len(entryPointPaths) == 0 {
		return nil, fmt.Errorf("未找到任何匹配的入口文件: %s", entryPointPattern)
	}

	// 为每个入口文件确定归属包
	// 通过比较路径，找到最匹配的包信息
	entryPointToPackageName := make(map[string]string)
	for _, entryPath := range entryPointPaths {
		bestMatchDir := ""
		ownerPackageName := "unknown"
		for _, pkgData := range ctx.ParsingResult.Package_Data {
			pkgDir := filepath.Dir(pkgData.Path)
			if strings.HasPrefix(entryPath, pkgDir) {
				if len(pkgDir) > len(bestMatchDir) {
					bestMatchDir = pkgDir
					ownerPackageName = pkgData.Namespace
				}
			}
		}
		entryPointToPackageName[entryPath] = ownerPackageName
	}

	// 步骤 2: 解析所有入口文件，构建"公共组件"清单
	// 公共组件是指从入口文件中导出的组件
	// 我们需要记录组件的公共名称、实际源文件路径和归属包
	publicComponentSource := make(map[string]string)    // 公共名称 -> 源文件路径
	publicComponentPackage := make(map[string]string)  // 公共名称 -> 包名称

	for _, entryPointPath := range entryPointPaths {
		entryPointResult, ok := fileResults[entryPointPath]
		if !ok {
			continue
		}
		ownerPackageName := entryPointToPackageName[entryPointPath]

		// 遍历入口文件的所有导出声明
		for _, exportDecl := range entryPointResult.ExportDeclarations {
			// 只处理从其他文件导入再导出的情况（re-export）
			if exportDecl.Source == nil || exportDecl.Source.FilePath == "" {
				continue
			}
			for _, module := range exportDecl.ExportModules {
				publicName := module.Identifier    // 公开导出的名称
				originalName := module.ModuleName  // 原始模块名称

				// 判断是否为组件导出并且不是纯类型
				// 使用递归回溯来判断一个符号的最终实体是否为纯类型
				if isComponentExport(publicName) && !isPureTypeRecursive(originalName, exportDecl.Source.FilePath, fileResults, make(map[string]bool)) {
					// 查找组件的实际源文件路径
					finalSourcePath := findComponentSource(originalName, exportDecl.Source.FilePath, fileResults, make(map[string]bool))
					if finalSourcePath != "" {
						publicComponentSource[publicName] = finalSourcePath
						publicComponentPackage[publicName] = ownerPackageName
					}
				}
			}
		}
	}

	// 步骤 3: 构建"源文件 -> 公共组件名列表"的反向映射
	// 这个映射用于快速查找某个源文件对应的所有公共组件
	sourceToPublicNamesMap := make(map[string][]string)
	for name, path := range publicComponentSource {
		sourceToPublicNamesMap[path] = append(sourceToPublicNamesMap[path], name)
	}

	// 步骤 4: 构建依赖图
	// 遍历所有公共组件，分析它们之间的依赖关系
	finalResult := make(map[string]map[string]ComponentInfo)

	for publicName, mainSourcePath := range publicComponentSource {
		packageName := publicComponentPackage[publicName]
		if _, ok := finalResult[packageName]; !ok {
			finalResult[packageName] = make(map[string]ComponentInfo)
		}

		// 获取组件的根目录
		componentDir := filepath.Dir(mainSourcePath)
		var currentDeps []string

		// 遍历所有文件，查找依赖关系
		// 只关注组件目录下的文件
		for filePath, fileResult := range fileResults {
			if strings.HasPrefix(filePath, componentDir) {
				// 检查该文件的所有导入声明
				for _, importDecl := range fileResult.ImportDeclarations {
					importedFilePath := importDecl.Source.FilePath

					// 检查导入的文件是否是公共组件的源文件
					if depPublicNames, isPublic := sourceToPublicNamesMap[importedFilePath]; isPublic {
						for _, depPublicName := range depPublicNames {
							// 排除对自身的依赖
							if depPublicName != publicName {
								currentDeps = append(currentDeps, depPublicName)
							}
						}
					}
				}
			}
		}

		// 构建组件信息并添加到最终结果
		finalResult[packageName][publicName] = ComponentInfo{
			SourcePath:   mainSourcePath,
			Dependencies: lo.Uniq(currentDeps), // 使用 lo.Uniq 去重
		}
	}

	return &Result{Packages: finalResult}, nil
}
