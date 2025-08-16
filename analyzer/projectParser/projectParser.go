// package projectParser 负责对整个 TypeScript/JavaScript 项目进行高级解析。
// 它整合了底层的文件扫描（scanProject）和单个文件解析（parser）功能，
// 构建出整个项目的依赖关系图和代码结构概览。
package projectParser

import (
	"fmt"
	"main/analyzer/parser"
	"main/analyzer/scanProject"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
)

// ProjectParserConfig 结构体定义了项目解析器所需的配置信息。
// 这些配置项指导解析器如何扫描文件、处理路径别名等。
type ProjectParserConfig struct {
	// RootPath 是待分析项目的根目录的绝对路径。
	RootPath string
	// RootAlias 是从根目录 tsconfig.json 解析出的路径别名映射。
	RootAlias map[string]string
	// PackageAliasMaps 存储了项目中所有找到的 tsconfig.json 的路径别名。
	// 键是 tsconfig.json 所在的目录的绝对路径，值是该配置对应的别名映射。
	PackageAliasMaps map[string]map[string]string
	// Extensions 是一个字符串切片，定义了需要被解析的文件的扩展名。
	Extensions []string
	// Ignore 是一个字符串切片，定义了在文件扫描时需要忽略的目录或文件的模式。
	Ignore []string
	// IsMonorepo 是一个布尔值，指示当前分析的项目是否是一个 monorepo 仓库。
	IsMonorepo bool
}

// ProjectParserResult 结构体是整个项目解析过程的最终结果容器。
// 它存储了配置信息、所有已解析的 JS/TS 文件的信息以及所有 `package.json` 文件的信息。
type ProjectParserResult struct {
	Config       ProjectParserConfig
	Js_Data      map[string]JsFileParserResult
	Package_Data map[string]PackageJsonFileParserResult
}

// NewProjectParserConfig 创建并初始化一个项目解析器的配置对象。
// 它会设置默认值，并根据项目类型（是否为 monorepo）解析路径别名。
func NewProjectParserConfig(rootPath string, alias map[string]string, extensions []string, ignore []string, isMonorepo bool) ProjectParserConfig {
	absRootPath, _ := filepath.Abs(rootPath)

	if ignore == nil || len(ignore) == 0 {
		ignore = []string{"**/node_modules/**", "**/dist/**", "**/build/**", "**/test/**", "**/public/**", "**/static/**"}
	}

	if extensions == nil || len(extensions) == 0 {
		extensions = []string{".ts", ".tsx", ".d.ts", ".js", ".jsx"}
	}

	// 如果用户没有提供别名，则从 tsconfig.json 中解析。
	rootAlias := alias
	if rootAlias == nil {
		rootAlias = ReadAliasFromTsConfig(absRootPath)
	}

	// 为 monorepo 项目查找所有子包的 tsconfig 别名。
	packageAliases := make(map[string]map[string]string)
	if isMonorepo {
		packageAliases = FindAllTsConfigsAndAliases(absRootPath, ignore)
	}

	return ProjectParserConfig{
		RootPath:         absRootPath,
		RootAlias:        rootAlias,
		PackageAliasMaps: packageAliases,
		Extensions:       extensions,
		Ignore:           ignore,
		IsMonorepo:       isMonorepo,
	}
}

// NewProjectParserResult 根据给定的配置，初始化一个用于存储项目解析结果的空容器。
func NewProjectParserResult(config ProjectParserConfig) *ProjectParserResult {
	return &ProjectParserResult{
		Config:       config,
		Js_Data:      make(map[string]JsFileParserResult),
		Package_Data: make(map[string]PackageJsonFileParserResult),
	}
}

// ProjectParser 是项目解析的入口和总调度方法。
func (ppr *ProjectParserResult) ProjectParser() {
	projectScanner := scanProject.NewProjectResult(ppr.Config.RootPath, ppr.Config.Ignore, ppr.Config.IsMonorepo)
	projectScanner.ScanProject()

	for targetPath, fileDetail := range projectScanner.GetFileList() {
		if lo.Contains(ppr.Config.Extensions, fileDetail.Ext) {
			ppr.parseJsFile(targetPath)
		}

		if fileDetail.FileName == "package.json" {
			ppr.parsePackageJson(targetPath)
		}
	}
}

// getAliasForFile 根据给定的文件路径，从已解析的所有 tsconfig 别名中找到最匹配的一个。
// 它返回最匹配的别名映射以及该别名配置所在的目录路径。
func (ppr *ProjectParserResult) getAliasForFile(targetPath string) (map[string]string, string) {
	bestMatchPath := ""
	// 默认使用根别名和根路径
	bestMatchAlias := ppr.Config.RootAlias
	bestMatchDir := ppr.Config.RootPath

	for tsconfigDir, aliasMap := range ppr.Config.PackageAliasMaps {
		// 检查 tsconfig 的目录是否是当前文件路径的前缀
		if strings.HasPrefix(targetPath, tsconfigDir) {
			// 我们寻找最长的匹配路径，即最深的子目录
			if len(tsconfigDir) > len(bestMatchPath) {
				bestMatchPath = tsconfigDir
				bestMatchAlias = aliasMap
				bestMatchDir = tsconfigDir // 更新为当前匹配的 tsconfig 目录
			}
		}
	}
	return bestMatchAlias, bestMatchDir
}

// parseJsFile 负责处理单个 JS/TS 文件的解析流程。
func (ppr *ProjectParserResult) parseJsFile(targetPath string) {
	fileParserResult := parser.NewParserResult(targetPath)
	fileParserResult.Traverse()
	result := fileParserResult.GetResult()

	// 为当前文件获取最匹配的路径别名配置和其所在目录
	aliasForFile, tsconfigDir := ppr.getAliasForFile(targetPath)

	ppr.Js_Data[targetPath] = JsFileParserResult{
		ImportDeclarations:    ppr.transformImportDeclarations(targetPath, result.ImportDeclarations, aliasForFile, tsconfigDir),
		ExportDeclarations:    ppr.transformExportDeclarations(targetPath, result.ExportDeclarations, aliasForFile, tsconfigDir),
		ExportAssignments:     result.ExportAssignments,
		InterfaceDeclarations: result.InterfaceDeclarations,
		TypeDeclarations:      result.TypeDeclarations,
		EnumDeclarations:      result.EnumDeclarations,
		VariableDeclarations:  result.VariableDeclarations,
		CallExpressions:       result.CallExpressions,
		JsxElements:           result.JsxElements,
	}
}

// parsePackageJson 负责处理单个 `package.json` 文件的解析。
func (ppr *ProjectParserResult) parsePackageJson(targetPath string) {
	packageJsonInfo, err := GetPackageJson(targetPath)
	if err != nil {
		fmt.Printf("解析 package.json 失败: %v\n", err)
		return
	}

	workspaceKey := "root"
	if filepath.Dir(targetPath) != ppr.Config.RootPath {
		workspaceKey = filepath.Base(filepath.Dir(targetPath))
	}

	ppr.Package_Data[workspaceKey] = PackageJsonFileParserResult{
		Workspace: workspaceKey,
		Path:      targetPath,
		Namespace: packageJsonInfo.Name,
		Version:   packageJsonInfo.Version,
		NpmList:   packageJsonInfo.NpmList,
	}
}

// transformImportDeclarations 将导入声明转换为高级格式，并使用给定的别名映射来解析模块源。
func (ppr *ProjectParserResult) transformImportDeclarations(importerPath string, decls []parser.ImportDeclarationResult, alias map[string]string, tsconfigDir string) []ImportDeclarationResult {
	return lo.Map(decls, func(decl parser.ImportDeclarationResult, _ int) ImportDeclarationResult {
		sourceData := MatchImportSource(
			importerPath,
			decl.Source,
			tsconfigDir, // 使用 tsconfig 所在的目录作为解析基准
			alias,
			ppr.Config.Extensions,
		)
		return ImportDeclarationResult{
			ImportModules: lo.Map(decl.ImportModules, func(module parser.ImportModule, _ int) ImportModule {
				return ImportModule{
					ImportModule: module.ImportModule,
					Type:         module.Type,
					Identifier:   module.Identifier,
				}
			}),
			Raw:    decl.Raw,
			Source: sourceData,
		}
	})
}

// transformExportDeclarations 将导出声明转换为高级格式，并使用给定的别名映射来解析模块源。
func (ppr *ProjectParserResult) transformExportDeclarations(importerPath string, decls []parser.ExportDeclarationResult, alias map[string]string, tsconfigDir string) []ExportDeclarationResult {
	return lo.Map(decls, func(decl parser.ExportDeclarationResult, _ int) ExportDeclarationResult {
		var sourceData *SourceData
		if decl.Source != "" {
			data := MatchImportSource(
				importerPath,
				decl.Source,
				tsconfigDir, // 使用 tsconfig 所在的目录作为解析基准
				alias,
				ppr.Config.Extensions,
			)
			sourceData = &data
		}

		return ExportDeclarationResult{
			ExportModules: lo.Map(decl.ExportModules, func(module parser.ExportModule, _ int) ExportModule {
				return ExportModule{
					ModuleName: module.ModuleName,
					Type:       module.Type,
					Identifier: module.Identifier,
				}
			}),
			Raw:    decl.Raw,
			Source: sourceData,
		}
	})
}
