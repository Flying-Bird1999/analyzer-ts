// 示例执行命令:

// 分析 shopline-admin-components 项目
// ./analyzer-ts analyze component-deps -i /Users/zxc/Desktop/analyzer/analyzer-ts/shopline-admin-components -m -p "component-deps.entryPoint=packages/sl-admin-components/src/v3.ts"

// 分析 nova 项目
// ./analyzer-ts analyze component-deps -i /Users/zxc/Desktop/analyzer/analyzer-ts/nova -m -p "component-deps.entryPoint=packages/*/src/index.ts"

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

// ComponentDependencyAnalyzer 实现了 Analyzer 接口，用于分析组件依赖。
type ComponentDependencyAnalyzer struct {
	EntryPoint string
}

func (a *ComponentDependencyAnalyzer) Name() string {
	return "component-deps"
}

func (a *ComponentDependencyAnalyzer) Configure(params map[string]string) error {
	if entryPoint, ok := params["entryPoint"]; ok {
		a.EntryPoint = entryPoint
	}
	return nil
}

func isComponentExport(name string) bool {
	if name == "" {
		return false
	}
	firstChar := []rune(name)[0]
	return unicode.IsUpper(firstChar)
}

func findComponentSource(symbolName, sourcePath string, fileResults map[string]projectParser.JsFileParserResult, visited map[string]bool) string {
	visitedKey := fmt.Sprintf("%s|%s", symbolName, sourcePath)
	if visited[visitedKey] {
		return ""
	}
	visited[visitedKey] = true

	fileResult, ok := fileResults[sourcePath]
	if !ok {
		return ""
	}

	for _, exportDecl := range fileResult.ExportDeclarations {
		if exportDecl.Source == nil {
			for _, module := range exportDecl.ExportModules {
				if module.Identifier == symbolName {
					return sourcePath
				}
			}
		}
	}
	// Note: This logic for default exports is simplified and might not cover all cases.
	if len(fileResult.ExportAssignments) > 0 {
		if getComponentName(sourcePath) == symbolName {
			return sourcePath
		}
	}

	for _, exportDecl := range fileResult.ExportDeclarations {
		if exportDecl.Source != nil && exportDecl.Source.FilePath != "" {
			for _, module := range exportDecl.ExportModules {
				if module.Identifier == symbolName {
					return findComponentSource(module.ModuleName, exportDecl.Source.FilePath, fileResults, visited)
				}
			}
		}
	}

	return sourcePath
}

func getComponentName(filePath string) string {
	base := filepath.Base(filePath)
	nameWithoutExt := strings.TrimSuffix(base, filepath.Ext(base))

	if nameWithoutExt == "index" {
		parentDir := filepath.Base(filepath.Dir(filePath))
		if parentDir == "src" || parentDir == "components" {
			return ""
		}
		return parentDir
	}

	return nameWithoutExt
}

func isPureTypeRecursive(symbolName, sourcePath string, fileResults map[string]projectParser.JsFileParserResult, visited map[string]bool) bool {
	visitedKey := fmt.Sprintf("%s|%s", symbolName, sourcePath)
	if visited[visitedKey] {
		return false // Cycle detected
	}
	visited[visitedKey] = true

	fileResult, ok := fileResults[sourcePath]
	if !ok {
		return false
	}

	// 1. Check if defined as a type in the current file
	if _, ok := fileResult.TypeDeclarations[symbolName]; ok {
		return true
	}
	if _, ok := fileResult.InterfaceDeclarations[symbolName]; ok {
		return true
	}
	if _, ok := fileResult.EnumDeclarations[symbolName]; ok {
		return true
	}

	// 2. Check if it's a re-export from another file
	for _, exportDecl := range fileResult.ExportDeclarations {
		if exportDecl.Source != nil && exportDecl.Source.FilePath != "" {
			for _, module := range exportDecl.ExportModules {
				if module.Identifier == symbolName {
					if isPureTypeRecursive(module.ModuleName, exportDecl.Source.FilePath, fileResults, visited) {
						return true
					}
				}
			}
		}
	}

	// 3. Check if it's imported from another file and then exported
	for _, importDecl := range fileResult.ImportDeclarations {
		for _, module := range importDecl.ImportModules {
			if module.Identifier == symbolName {
				// Now check if this identifier is exported from the current file
				for _, exportDecl := range fileResult.ExportDeclarations {
					if exportDecl.Source == nil { // e.g., export { symbolName }
						for _, exportModule := range exportDecl.ExportModules {
							if exportModule.ModuleName == symbolName {
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

	return false
}

// Analyze 是分析器的主方法。
func (a *ComponentDependencyAnalyzer) Analyze(ctx *project_analyzer.ProjectContext) (project_analyzer.Result, error) {
	if a.EntryPoint == "" {
		return nil, fmt.Errorf("错误: 请使用 -p 'component-deps.entryPoint=path/to/entry.ts' 参数指定入口文件")
	}

	fileResults := ctx.ParsingResult.Js_Data

	// 步骤 1: 发现所有入口文件并确定它们的归属包
	entryPointPattern := filepath.Join(ctx.ProjectRoot, a.EntryPoint)
	entryPointPaths, err := filepath.Glob(entryPointPattern)
	if err != nil {
		return nil, fmt.Errorf("解析 glob 模式失败: %w", err)
	}
	if len(entryPointPaths) == 0 {
		return nil, fmt.Errorf("未找到任何匹配的入口文件: %s", entryPointPattern)
	}

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

	// 步骤 2: 解析所有入口文件，构建“公共组件”清单
	publicComponentSource := make(map[string]string)
	publicComponentPackage := make(map[string]string)

	for _, entryPointPath := range entryPointPaths {
		entryPointResult, ok := fileResults[entryPointPath]
		if !ok {
			continue
		}
		ownerPackageName := entryPointToPackageName[entryPointPath]

		for _, exportDecl := range entryPointResult.ExportDeclarations {
			if exportDecl.Source == nil || exportDecl.Source.FilePath == "" {
				continue
			}
			for _, module := range exportDecl.ExportModules {
				publicName := module.Identifier
				originalName := module.ModuleName

				// 使用递归回溯来判断一个符号的最终实体是否为纯类型
				if isComponentExport(publicName) && !isPureTypeRecursive(originalName, exportDecl.Source.FilePath, fileResults, make(map[string]bool)) {
					finalSourcePath := findComponentSource(originalName, exportDecl.Source.FilePath, fileResults, make(map[string]bool))
					if finalSourcePath != "" {
						publicComponentSource[publicName] = finalSourcePath
						publicComponentPackage[publicName] = ownerPackageName
					}
				}
			}
		}
	}

	// 步骤 3: 构建“源文件 -> 公共组件名列表”的反向地图
	sourceToPublicNamesMap := make(map[string][]string)
	for name, path := range publicComponentSource {
		sourceToPublicNamesMap[path] = append(sourceToPublicNamesMap[path], name)
	}

	// 步骤 4: 构建依赖图
	finalResult := make(map[string]map[string]ComponentInfo)

	for publicName, mainSourcePath := range publicComponentSource {
		packageName := publicComponentPackage[publicName]
		if _, ok := finalResult[packageName]; !ok {
			finalResult[packageName] = make(map[string]ComponentInfo)
		}

		componentDir := filepath.Dir(mainSourcePath)
		var currentDeps []string

		for filePath, fileResult := range fileResults {
			if strings.HasPrefix(filePath, componentDir) {
				for _, importDecl := range fileResult.ImportDeclarations {
					importedFilePath := importDecl.Source.FilePath
					if depPublicNames, isPublic := sourceToPublicNamesMap[importedFilePath]; isPublic {
						for _, depPublicName := range depPublicNames {
							if depPublicName != publicName {
								currentDeps = append(currentDeps, depPublicName)
							}
						}
					}
				}
			}
		}

		finalResult[packageName][publicName] = ComponentInfo{
			SourcePath:   mainSourcePath,
			Dependencies: lo.Uniq(currentDeps),
		}
	}

	return &Result{Packages: finalResult}, nil
}
