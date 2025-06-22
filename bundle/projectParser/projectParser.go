package projectParser

import (
	"fmt"
	"main/bundle/parser"
	"main/bundle/scanProject"
	"path/filepath"

	"github.com/samber/lo"
)

type ProjectParserResult struct {
	RootPath   string            // 项目根目录
	Alias      map[string]string // 别名映射，key: 别名, value: 实际路径
	Extensions []string          // 扩展名列表，例如: [".ts", ".tsx",".js", ".jsx"]
	Ignore     []string          // 指定忽略的文件/文件夹
	IsMonorepo bool              // 是否为 monorepo 项目

	Js_Data      map[string]JsFileParserResult
	Package_Data map[string]PackageJsonFileParserResult
}

func NewProjectParserResult(rootPath string, Alias map[string]string, Extensions []string, Ignore []string, IsMonorepo bool) *ProjectParserResult {
	curAlias := FormatAlias(Alias)
	if Alias == nil {
		// 如果没有传入 Alias，尝试读取项目中tsconfig.json的 alias
		curAlias = ReadAliasFromTsConfig(rootPath)
	}

	curIgnore := Ignore
	if Ignore == nil {
		curIgnore = []string{"node_modules", "dist", "build", "public", "static", "docs"}
	}

	curExtensions := Extensions
	if Extensions == nil {
		curExtensions = []string{".ts", ".tsx", ".d.ts", ".js", ".jsx"}
	}

	newRootPath, _ := filepath.Abs(rootPath)

	// 这里可以再自行检测一下是否为 IsMonorepo

	return &ProjectParserResult{
		RootPath:   newRootPath,
		Alias:      curAlias,
		Extensions: curExtensions,
		Ignore:     curIgnore,
		IsMonorepo: IsMonorepo,

		Js_Data:      make(map[string]JsFileParserResult),
		Package_Data: make(map[string]PackageJsonFileParserResult),
	}
}

func (ar *ProjectParserResult) ProjectParser() {
	// 扫描项目
	projectResult := scanProject.NewProjectResult(ar.RootPath, ar.Ignore, ar.IsMonorepo)
	projectResult.ScanProject()

	// 扫描文件
	for targetPath, fileDetail := range projectResult.GetFileList() {
		// 判断是否为 js 文件, 解析文件名后缀为 js、jsx、ts、tsx、d.ts
		if lo.Contains([]string{".js", ".jsx", ".ts", ".tsx", ".d.ts"}, fileDetail.Ext) {
			pr := parser.NewParserResult(targetPath)
			pr.Traverse()
			result := pr.GetResult()

			importResult := make([]ImportDeclarationResult, 0)

			// 处理每个 import 声明
			for _, importDecl := range result.ImportDeclarations {
				// TODO: 这里的Npm先传入根目录/最外层的，多包的场景需要先看自身的，再看外层的
				sourceData := MatchImportSource(targetPath, importDecl.Source, ar.RootPath, ar.Alias, ar.Extensions)
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

			ar.Js_Data[targetPath] = JsFileParserResult{
				ImportDeclarations:    importResult,
				InterfaceDeclarations: result.InterfaceDeclarations,
				TypeDeclarations:      result.TypeDeclarations,
			}
		}

		// 判断是否为 package.json 文件
		if fileDetail.FileName == "package.json" {
			packageJsonInfo, err := GetPackageJson(targetPath)
			if err != nil {
				fmt.Printf("解析 package.json 文件失败: %v\n", err)
			}
			// 如果是最外层的 package.json，位于根目录下，则Workspace为root
			if filepath.Dir(targetPath) == ar.RootPath {
				ar.Package_Data["root"] = PackageJsonFileParserResult{
					Workspace: "root",
					Path:      targetPath,
					Namespace: packageJsonInfo.Name,
					Version:   packageJsonInfo.Version,
					NpmList:   packageJsonInfo.NpmList,
				}
			} else {
				ar.Package_Data[filepath.Base(filepath.Dir(targetPath))] = PackageJsonFileParserResult{
					Workspace: filepath.Base(filepath.Dir(targetPath)),
					Path:      targetPath,
					Namespace: packageJsonInfo.Name,
					Version:   packageJsonInfo.Version,
					NpmList:   packageJsonInfo.NpmList,
				}
			}
		}
	}
}
