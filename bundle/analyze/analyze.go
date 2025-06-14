package analyze

import (
	"main/bundle/parser"
	"main/bundle/scanProject"
	"path/filepath"

	"github.com/samber/lo"
)

type AnalyzeResult struct {
	RootPath   string            // 项目根目录
	Alias      map[string]string // 别名映射，key: 别名, value: 实际路径
	Extensions []string          // 扩展名列表，例如: [".ts", ".tsx",".js", ".jsx"]
	IsMonorepo bool              // 是否为 monorepo 项目

	File map[string]FileAnalyzeResult
	Npm  scanProject.ProjectNpmList
}

func NewAnalyzeResult(rootPath string, Alias map[string]string, Extensions []string, IsMonorepo bool) *AnalyzeResult {
	curAlias := FormatAlias(Alias)
	if Alias == nil {
		// 如果没有传入 Alias，尝试读取项目中tsconfig.json的 alias
		curAlias = ReadAliasFromTsConfig(rootPath)
	}

	curExtensions := Extensions
	if Extensions == nil {
		curExtensions = []string{".ts", ".tsx", ".d.ts", ".js", ".jsx"}
	}

	newRootPath, _ := filepath.Abs(rootPath)

	// 这里可以再自行检测一下 IsMonorepo

	return &AnalyzeResult{
		RootPath:   newRootPath,
		Alias:      curAlias,
		Extensions: curExtensions,
		IsMonorepo: IsMonorepo,
		File:       make(map[string]FileAnalyzeResult),
		Npm:        make(scanProject.ProjectNpmList),
	}
}

func (ar *AnalyzeResult) GetFileData() map[string]FileAnalyzeResult {
	return ar.File
}

func (ar *AnalyzeResult) GetNpmData() scanProject.ProjectNpmList {
	return ar.Npm
}

// 是否命中别名 alias，如果命中则做替换
func (ar *AnalyzeResult) isMatchAlias(filePath string) (string, bool) {
	return IsMatchAlias(filePath, ar.RootPath, ar.Alias)
}

func (ar *AnalyzeResult) Analyze() {
	// 扫描项目
	projectResult := scanProject.NewProjectResult(ar.RootPath, []string{}, ar.IsMonorepo)
	projectResult.ScanProject()

	// 赋值扫描的npm列表
	ar.Npm = projectResult.GetNpmList()

	// 扫描文件
	for targetPath, item := range projectResult.GetFileList() {
		pr := parser.NewParserResult(item.Path)
		pr.Traverse()
		result := pr.GetResult()

		importResult := make([]ImportDeclarationResult, 0)

		// 处理每个 import 声明
		for _, importDecl := range result.ImportDeclarations {
			// TODO: 这里的Npm先传入根目录/最外层的，多包的场景需要先看自身的，再看外层的
			sourceData := MatchImportSource(targetPath, importDecl.Source, ar.RootPath, ar.Npm["root"].NpmList, ar.Alias, ar.Extensions)
			importResult = append(importResult, ImportDeclarationResult{
				ImportModules: lo.Map(importDecl.ImportModules, func(module parser.ImportModule, _ int) ImportModule {
					return ImportModule{
						ImportModule: module.ImportModule,
						Type:         module.Type,
						Identifier:   module.Identifier,
					}
				}),
				Raw:    importDecl.Raw,
				Source: sourceData,
			})
		}

		ar.File[item.Path] = FileAnalyzeResult{
			ImportDeclarations:    importResult,
			InterfaceDeclarations: result.InterfaceDeclarations,
			TypeDeclarations:      result.TypeDeclarations,
		}
	}
}
