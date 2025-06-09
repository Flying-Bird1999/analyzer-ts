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

	File map[string]FileAnalyzeResult
	Npm  map[string]scanProject.NpmItem
}

func NewAnalyzeResult(rootPath string, Alias map[string]string, Extensions []string) *AnalyzeResult {
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

	return &AnalyzeResult{
		RootPath:   newRootPath,
		Alias:      curAlias,
		Extensions: curExtensions,
		File:       make(map[string]FileAnalyzeResult),
		Npm:        make(map[string]scanProject.NpmItem),
	}
}

func (ar *AnalyzeResult) GetFileData() map[string]FileAnalyzeResult {
	return ar.File
}

func (ar *AnalyzeResult) GetNpmData() map[string]scanProject.NpmItem {
	return ar.Npm
}

// 是否命中别名 alias，如果命中则做替换
func (ar *AnalyzeResult) isMatchAlias(filePath string) (string, bool) {
	return IsMatchAlias(filePath, ar.RootPath, ar.Alias)
}

func (ar *AnalyzeResult) Analyze() {
	// 扫描项目
	projectResult := scanProject.NewProjectResult(ar.RootPath, []string{})
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
			sourceData := MatchImportSource(targetPath, importDecl.Source, ar.RootPath, ar.Npm, ar.Alias, ar.Extensions)
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
