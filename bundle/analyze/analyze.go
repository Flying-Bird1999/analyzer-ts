package analyze

import (
	"main/bundle/parser"
	"main/bundle/scanProject"
	"main/bundle/utils"
	"path/filepath"
	"strings"

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
	curAlias := Alias
	if Alias == nil {
		// 如果没有传入 Alias，尝试读取项目中tsconfig.json的 alias
		curAlias = utils.ReadAliasFromTsConfig(rootPath)
	}

	curExtensions := Extensions
	if Extensions == nil {
		curExtensions = []string{".ts", ".tsx", ".js", ".jsx"}
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

// 是否命中别名 alias，如果命中则做替换
// 1. 匹配是否命中，
func (ar *AnalyzeResult) isMatchAlias(filePath string) (string, bool) {
	for alias, realPath := range ar.Alias {
		// 检查路径是否以 alias 开头,并确保 alias 后面是路径分隔符或结束
		if strings.HasPrefix(filePath, alias) && (len(filePath) == len(alias) || filePath[len(alias)] == '/') {
			// 替换 alias 为真实路径
			absolutePath := filepath.Join(realPath, strings.TrimPrefix(filePath, alias))
			return absolutePath, true
		}
	}
	return "", false // 未命中别名
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
			sourceData := ar.matchImportSource(targetPath, importDecl.Source, projectResult.GetFileList())
			importResult = append(importResult, ImportDeclarationResult{
				Modules: lo.Map(importDecl.Modules, func(module parser.Module, _ int) Module {
					return Module{
						Module:     module.Module,
						Type:       module.Type,
						Identifier: module.Identifier,
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

// 匹配 import 的真实绝对路径
func (ar *AnalyzeResult) matchImportSource(targetPath string, filePath string, fileList map[string]scanProject.FileItem) SourceData {
	realPath := filePath

	// 1. 匹配 alias,替换为真实路径
	if absolutePath, matched := ar.isMatchAlias(filePath); matched {
		realPath = absolutePath
	}

	// 2. 需要检查结尾是否有文件后缀，如果没有后缀，需要基于Extensions尝试去匹配
	if !utils.HasValidExtension(realPath, ar.Extensions) {
		for _, ext := range ar.Extensions {
			// 尝试直接拼接扩展名
			extendedPath := realPath + ext
			if _, exists := fileList[extendedPath]; exists {
				realPath = extendedPath
				break
			}

			// 如果找不到文件，尝试将其视为目录并拼接 index 文件
			indexPath := filepath.Join(realPath, "index"+ext)
			if _, exists := fileList[indexPath]; exists {
				realPath = extendedPath
				break
			}
		}
	}

	// 3. 匹配 npm 包
	for npmName, npmItem := range ar.Npm {
		// 检查 filePath 是否包含 npm 包名
		if strings.HasPrefix(filePath, npmName) {
			return SourceData{
				FilePath: filePath,
				NpmPkg:   npmItem.Name,
				Type:     "npm",
			}
		}
	}
	// 4. 将路径转换为绝对路径
	returnPath, _ := filepath.Abs(filepath.Join(filepath.Dir(targetPath), realPath))

	// 5. 如果存在，则返回 SourceData
	if _, exists := fileList[returnPath]; exists {
		return SourceData{
			FilePath: returnPath,
			NpmPkg:   "",
			Type:     "file",
		}
	}
	return SourceData{
		FilePath: filePath,
		NpmPkg:   "",
		Type:     "unknown",
	}
}
