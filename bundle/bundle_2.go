package bundle

// 换个思路，这里从入口文件开始解析，递归解析依赖，将依赖的类型、接口、import 都收集起来，最后输出到一个文件中

import (
	"main/bundle/analyze"
	"main/bundle/parser"
	"main/bundle/scanProject"
	"main/bundle/utils"
	"path/filepath"
	"strings"
)

type BundleResult struct {
	RootPath   string
	Alias      map[string]string
	Extensions []string
	NpmList    map[string]scanProject.NpmItem

	SourceCodeMap map[string]string
}

func NewBundleResult(inputAnalyzeFile string, inputAnalyzeType string) BundleResult {
	// 1. 通过截取 inputAnalyzeFile 中的路径，匹配到/src前边的部分，得到 rootPath
	absFilePath, _ := filepath.Abs(inputAnalyzeFile)
	rootPath := strings.Split(absFilePath, "/src")[0]

	// 2. 获取 npm 列表
	pr := scanProject.NewProjectResult(rootPath, []string{})
	pr.ScanNpmList()

	// 3. 获取 tsconfig.json 中的 alias 列表
	ar := analyze.NewAnalyzeResult(rootPath, nil, nil)

	return BundleResult{
		RootPath:      rootPath,
		Alias:         ar.Alias,
		Extensions:    ar.Extensions,
		NpmList:       pr.GetNpmList(),
		SourceCodeMap: make(map[string]string),
	}
}

// 递归解析依赖
// absFilePath必须传入绝对路径
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
				sourceData := analyze.MatchImportSource(absFilePath, importDecl.Source, br.RootPath, br.NpmList, br.Alias, br.Extensions)

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

					sourceData := analyze.MatchImportSource(absFilePath, importDecl.Source, br.RootPath, br.NpmList, br.Alias, br.Extensions)
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
