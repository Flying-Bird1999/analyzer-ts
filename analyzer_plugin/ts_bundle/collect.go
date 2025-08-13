// 该文件实现了基于入口文件和类型名的 TypeScript 类型依赖递归收集与输出。
// 主要流程为：
// 1. 以入口文件和类型为起点，递归解析类型、接口及 import 依赖，收集所有相关类型声明源码。
// 2. 支持 alias、npm 包、命名空间导入等常见 TypeScript 导入场景。
// 3. 最终将所有依赖类型源码合并输出到指定文件。

package ts_bundle

import (
	"fmt"
	"main/analyzer/parser"
	"main/analyzer/projectParser"
	"main/analyzer/utils"
	"path/filepath"
	"strings"
)

type CollectResult struct {
	RootPath   string            // 项目根目录
	Alias      map[string]string // tsconfig.json 中的路径别名
	Extensions []string          // 支持的文件扩展名
	// NpmList       scanProject.ProjectNpmList // 依赖列表
	SourceCodeMap map[string]string // 已收集的类型源码
}

// NewCollectResult 构造函数，初始化 CollectResult。
// 通过入口文件路径自动推断项目根目录、npm 列表、alias、扩展名等信息。
func NewCollectResult(inputAnalyzeFile string, inputAnalyzeType string, projectRootPath string) CollectResult {
	var rootPath string = projectRootPath

	// TODO: 逻辑待优化
	// 1. 通过截取 inputAnalyzeFile 中的路径，匹配到/src前边的部分，得到 rootPath
	if projectRootPath == "" {
		absFilePath, _ := filepath.Abs(inputAnalyzeFile)
		rootPath = strings.Split(absFilePath, "/src")[0]
	}

	// 2. 获取 tsconfig.json 中的 alias 列表
	config := projectParser.NewProjectParserConfig(rootPath, nil, nil, []string{}, false)
	ar := projectParser.NewProjectParserResult(config)

	return CollectResult{
		RootPath:   rootPath,
		Alias:      ar.Config.RootAlias,
		Extensions: ar.Config.Extensions,
		// NpmList:       pr.GetNpmList(),
		SourceCodeMap: make(map[string]string),
	}
}

// collectFileType 递归解析指定文件中的类型依赖。
// absFilePath: 当前解析的文件绝对路径path:/Users/bird/Desktop/sp/smart-push-new/tsconfig.base.json
// typeName: 当前要查找的类型名
// replaceTypeName: 类型重命名（如 import {A as B}）时的替换名
// parentTypeName: 父类型名（用于命名空间类型替换）
func (br *CollectResult) collectFileType(absFilePath string, typeName string, replaceTypeName string, parentTypeName string) {
	// TODO: 已经解析过的文件可以做缓存
	fmt.Printf("开始解析当前文件: %s \n", absFilePath)

	// 解析当前文件
	pr := parser.NewParserResult(absFilePath)
	pr.Traverse()
	parserResult := pr.GetResult()

	// 查找类型声明 type
	if typeDecl, found := parserResult.TypeDeclarations[typeName]; found {
		realRaw := typeDecl.Raw
		if replaceTypeName != "" {
			realRaw = strings.ReplaceAll(typeDecl.Raw, typeName, replaceTypeName)
		}

		br.SourceCodeMap[absFilePath+"_"+typeName] = realRaw
		for ref := range typeDecl.Reference {
			br.collectFileType(absFilePath, ref, "", typeName)
		}
		return
	}

	// 查找接口声明 interface
	if interfaceDecl, found := parserResult.InterfaceDeclarations[typeName]; found {
		realRaw := interfaceDecl.Raw
		if replaceTypeName != "" {
			realRaw = strings.ReplaceAll(interfaceDecl.Raw, typeName, replaceTypeName)
		}
		br.SourceCodeMap[absFilePath+"_"+typeName] = realRaw
		for ref := range interfaceDecl.Reference {
			br.collectFileType(absFilePath, ref, "", typeName)
		}
		return
	}

	// 查找枚举声明 enum
	if enumDecl, found := parserResult.EnumDeclarations[typeName]; found {
		realRaw := enumDecl.Raw
		if replaceTypeName != "" {
			realRaw = strings.ReplaceAll(enumDecl.Raw, typeName, replaceTypeName)
		}
		br.SourceCodeMap[absFilePath+"_"+typeName] = realRaw
		return
	}

	// 查找 import 依赖
	for _, importDecl := range parserResult.ImportDeclarations {
		for _, module := range importDecl.ImportModules {
			// 普通命名导入
			if module.Identifier == typeName {
				realTypeName := typeName
				var replaceTypeName string
				if module.Type == "named" && module.ImportModule != typeName {
					realTypeName = module.ImportModule
					replaceTypeName = typeName
				}
				sourceData := projectParser.MatchImportSource(absFilePath, importDecl.Source, br.RootPath, br.Alias, br.Extensions)

				nextFile := ""
				if sourceData.Type == "file" {
					nextFile = sourceData.FilePath
				} else if sourceData.Type == "npm" {
					nextFile = utils.ResolveNpmPath(absFilePath, br.RootPath, importDecl.Source, true)
					// 检查结尾是否有文件后缀，如果没有后缀，需要基于Extensions尝试去匹配
					if !utils.HasExtension(nextFile, br.Extensions) {
						nextFile = utils.FindRealFilePath(nextFile, br.Extensions)
					}
				}
				br.collectFileType(nextFile, realTypeName, replaceTypeName, typeName)
			}

			// case: import * as allTypes from './type';
			if module.Type == "namespace" {
				// 解析typeName: allTypes.MerchantData。提取出 allTypes.MerchantData 中的 MerchantData
				refNameArr := strings.Split(typeName, ".")
				realRefName := refNameArr[len(refNameArr)-1] // MerchantData
				if refNameArr[0] == module.Identifier {      // allTypes
					var replaceTypeName = module.Identifier + "_" + realRefName // allTypes_MerchantData
					// 替换源码的类型，PreloadedState中的 allTypes.MerchantData -> allTypes_MerchantData
					realTargetTypeRaw := strings.ReplaceAll(br.SourceCodeMap[absFilePath+"_"+parentTypeName], typeName, replaceTypeName)
					br.SourceCodeMap[absFilePath+"_"+parentTypeName] = realTargetTypeRaw

					sourceData := projectParser.MatchImportSource(absFilePath, importDecl.Source, br.RootPath, br.Alias, br.Extensions)
					nextFile := ""
					if sourceData.Type == "file" {
						nextFile = sourceData.FilePath
					} else {
						nextFile = utils.ResolveNpmPath(absFilePath, br.RootPath, importDecl.Source, true)
						// 检查结尾是否有文件后缀，如果没有后缀，需要基于Extensions尝试去匹配
						if !utils.HasExtension(nextFile, br.Extensions) {
							nextFile = utils.FindRealFilePath(nextFile, br.Extensions)
						}
					}
					br.collectFileType(nextFile, realRefName, replaceTypeName, typeName)
				}
			}
		}
	}
}
