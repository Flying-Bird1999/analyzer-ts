package bundle

// 换个思路，这里从入口文件开始解析，递归解析依赖，将依赖的类型、接口、import 都收集起来，最后输出到一个文件中

import (
	"fmt"
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
func (br *BundleResult) analyzeFileAndType(absFilePath string, typeName string) {
	// 解析当前文件
	pr := parser.NewParserResult(absFilePath)
	pr.Traverse()
	parserResult := pr.GetResult()

	// 查找类型声明
	if typeDecl, found := parserResult.TypeDeclarations[typeName]; found {
		br.SourceCodeMap[absFilePath+"_"+typeName] = typeDecl.Raw
		for ref := range typeDecl.Reference {
			br.analyzeFileAndType(absFilePath, ref)
		}
		return
	}
	// 查找接口声明
	if interfaceDecl, found := parserResult.InterfaceDeclarations[typeName]; found {
		br.SourceCodeMap[absFilePath+"_"+typeName] = interfaceDecl.Raw
		for ref := range interfaceDecl.Reference {
			br.analyzeFileAndType(absFilePath, ref)
		}
		return
	}

	// 查找 import 依赖
	for _, importDecl := range parserResult.ImportDeclarations {
		for _, module := range importDecl.ImportModules {
			// 普通命名导入
			if module.Identifier == typeName {
				realTypeName := typeName
				if module.Type == "named" && module.ImportModule != typeName {
					realTypeName = module.ImportModule
				}
				sourceData := analyze.MatchImportSource(absFilePath, importDecl.Source, br.RootPath, br.NpmList, br.Alias, br.Extensions)

				nextFile := ""
				if sourceData.Type == "file" {
					nextFile = sourceData.FilePath
				} else {
					nextFile = br.RootPath + "/node_modules" + importDecl.Source
					fmt.Printf("nextFile: %s\n", nextFile)
				}
				br.analyzeFileAndType(nextFile, realTypeName)
			}
			// 命名空间导入
			if module.Type == "namespace" {
				refNameArr := strings.Split(typeName, ".")
				if len(refNameArr) == 2 && refNameArr[0] == module.Identifier {
					realTypeName := refNameArr[1]
					replaceTypeName := module.Identifier + "_" + realTypeName
					// 替换源码
					key := absFilePath + "_" + typeName
					if raw, ok := br.SourceCodeMap[key]; ok {
						br.SourceCodeMap[key] = strings.ReplaceAll(raw, typeName, replaceTypeName)
					}
					sourceData := analyze.MatchImportSource(absFilePath, importDecl.Source, br.RootPath, br.NpmList, br.Alias, br.Extensions)

					nextFile := ""
					if sourceData.Type == "file" {
						nextFile = sourceData.FilePath
					} else {
						nextFile = br.RootPath + "/node_modules" + importDecl.Source
					}
					br.analyzeFileAndType(nextFile, realTypeName)
				}
			}
		}
	}
}

// 入口方法
func GenerateBundle2() {
	// inputAnalyzeFile := "/Users/zxc/Desktop/shopline-live-sale/src/feature/LiveRoom/components/MainLeft/ProductSet/AddProductSetPicker/index.tsx"
	// inputAnalyzeType := "Name"
	inputAnalyzeFile := "/Users/zxc/Desktop/shopline-order-detail/src/interface/preloadedState/index.ts"
	inputAnalyzeType := "PreloadedState"

	br := NewBundleResult(inputAnalyzeFile, inputAnalyzeType)
	br.analyzeFileAndType(inputAnalyzeFile, inputAnalyzeType)

	resultCode := ""
	for _, value := range br.SourceCodeMap {
		resultCode += value + "\n"
	}
	utils.WriteResultToFile("./ts/output/result.ts", resultCode)
}
