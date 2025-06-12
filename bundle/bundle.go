package bundle

// 该文件实现了基于入口文件和类型名的 TypeScript 类型依赖递归收集与输出。
// 主要流程为：
// 1. 以入口文件和类型为起点，递归解析类型、接口及 import 依赖，收集所有相关类型声明源码。
// 2. 支持 alias、npm 包、命名空间导入等常见 TypeScript 导入场景。
// 3. 最终将所有依赖类型源码合并输出到指定文件。

import (
	"main/bundle/analyze"
	"main/bundle/parser"
	"main/bundle/scanProject"
	"main/bundle/utils"
	"path/filepath"
	"strings"
)

type BundleResult struct {
	RootPath      string                     // 项目根目录
	Alias         map[string]string          // tsconfig.json 中的路径别名
	Extensions    []string                   // 支持的文件扩展名
	NpmList       scanProject.ProjectNpmList // 依赖列表
	SourceCodeMap map[string]string          // 已收集的类型源码
}

// NewBundleResult 构造函数，初始化 BundleResult。
// 通过入口文件路径自动推断项目根目录、npm 列表、alias、扩展名等信息。
func NewBundleResult(inputAnalyzeFile string, inputAnalyzeType string) BundleResult {
	// 1. 通过截取 inputAnalyzeFile 中的路径，匹配到/src前边的部分，得到 rootPath
	absFilePath, _ := filepath.Abs(inputAnalyzeFile)
	rootPath := strings.Split(absFilePath, "/src")[0]

	// 2. 获取 npm 列表
	pr := scanProject.NewProjectResult(rootPath, []string{}, false)
	pr.ScanNpmList()

	// 3. 获取 tsconfig.json 中的 alias 列表
	ar := analyze.NewAnalyzeResult(rootPath, nil, nil, false)

	return BundleResult{
		RootPath:      rootPath,
		Alias:         ar.Alias,
		Extensions:    ar.Extensions,
		NpmList:       pr.GetNpmList(),
		SourceCodeMap: make(map[string]string),
	}
}

// analyzeFileAndType 递归解析指定文件中的类型依赖。
// absFilePath: 当前解析的文件绝对路径
// typeName: 当前要查找的类型名
// replaceTypeName: 类型重命名（如 import {A as B}）时的替换名
// parentTypeName: 父类型名（用于命名空间类型替换）
func (br *BundleResult) analyzeFileAndType(absFilePath string, typeName string, replaceTypeName string, parentTypeName string) {
	// 解析当前文件
	pr := parser.NewParserResult(absFilePath)
	pr.Traverse()
	parserResult := pr.GetResult()

	// 查找类型声明
	if typeDecl, found := parserResult.TypeDeclarations[typeName]; found {
		realRaw := typeDecl.Raw
		if replaceTypeName != "" {
			realRaw = strings.ReplaceAll(typeDecl.Raw, typeName, replaceTypeName)
		}

		br.SourceCodeMap[absFilePath+"_"+typeName] = realRaw
		for ref := range typeDecl.Reference {
			br.analyzeFileAndType(absFilePath, ref, "", typeName)
		}
		return
	}
	// 查找接口声明
	if interfaceDecl, found := parserResult.InterfaceDeclarations[typeName]; found {
		realRaw := interfaceDecl.Raw
		if replaceTypeName != "" {
			realRaw = strings.ReplaceAll(interfaceDecl.Raw, typeName, replaceTypeName)
		}
		br.SourceCodeMap[absFilePath+"_"+typeName] = realRaw
		for ref := range interfaceDecl.Reference {
			br.analyzeFileAndType(absFilePath, ref, "", typeName)
		}
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
				sourceData := analyze.MatchImportSource(absFilePath, importDecl.Source, br.RootPath, br.NpmList["root"].NpmList, br.Alias, br.Extensions)

				nextFile := ""
				if sourceData.Type == "file" {
					nextFile = sourceData.FilePath
				} else {
					// TODO： 待优化： npm的case
					nextFile = br.RootPath + "/node_modules/" + importDecl.Source
					// 检查结尾是否有文件后缀，如果没有后缀，需要基于Extensions尝试去匹配
					if !utils.HasExtension(nextFile) {
						nextFile = utils.FindRealFilePath(nextFile, br.Extensions)
					}
				}
				br.analyzeFileAndType(nextFile, realTypeName, replaceTypeName, typeName)
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

					sourceData := analyze.MatchImportSource(absFilePath, importDecl.Source, br.RootPath, br.NpmList["root"].NpmList, br.Alias, br.Extensions)
					nextFile := ""
					if sourceData.Type == "file" {
						nextFile = sourceData.FilePath
					} else {
						// TODO： 待优化： npm的case
						nextFile = br.RootPath + "/node_modules/" + importDecl.Source
						// 检查结尾是否有文件后缀，如果没有后缀，需要基于Extensions尝试去匹配
						if !utils.HasExtension(nextFile) {
							nextFile = utils.FindRealFilePath(nextFile, br.Extensions)
						}
					}
					br.analyzeFileAndType(nextFile, realRefName, replaceTypeName, typeName)
				}
			}
		}
	}
}

// 入口方法
func GenerateBundle2() {
	inputAnalyzeFile := "/Users/zxc/Desktop/shopline-live-sale/src/feature/LiveRoom/components/MainLeft/ProductSet/AddProductSetPicker/index.tsx"
	inputAnalyzeType := "Name"
	// inputAnalyzeFile := "/Users/zxc/Desktop/shopline-order-detail/src/interface/preloadedState/index.ts"
	// inputAnalyzeType := "PreloadedState"

	br := NewBundleResult(inputAnalyzeFile, inputAnalyzeType)
	br.analyzeFileAndType(inputAnalyzeFile, inputAnalyzeType, "", "")

	resultCode := ""
	for _, value := range br.SourceCodeMap {
		resultCode += value + "\n"
	}
	utils.WriteResultToFile("./ts/output/result.ts", resultCode)
}
