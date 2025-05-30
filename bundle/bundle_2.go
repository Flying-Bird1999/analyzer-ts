package bundle

// 换个思路，这里从入口文件开始解析，递归解析依赖，将依赖的类型、接口、import 都收集起来，最后输出到一个文件中

import (
	"fmt"
	"main/bundle/parser"
	"main/bundle/utils"
	"path/filepath"
	"strings"
)

// 递归解析依赖
func analyzeFileAndType(rootPath, filePath, typeName string, sourceCodeMap map[string]string, visited map[string]bool) {
	// 这里精准处理导入的路径，就可以了
	absFilePath, _ := filepath.Abs(filePath)

	fmt.Printf("absFilePath: %s\n", absFilePath)

	visitKey := absFilePath + "::" + typeName
	if visited[visitKey] {
		return
	}
	visited[visitKey] = true

	// 解析当前文件
	pr := parser.NewParserResult(absFilePath)
	pr.Traverse()
	parserResult := pr.GetResult()

	// 查找类型声明
	if typeDecl, found := parserResult.TypeDeclarations[typeName]; found {
		sourceCodeMap[absFilePath+"_"+typeName] = typeDecl.Raw
		for ref := range typeDecl.Reference {
			analyzeFileAndType(rootPath, absFilePath, ref, sourceCodeMap, visited)
		}
		return
	}
	// 查找接口声明
	if interfaceDecl, found := parserResult.InterfaceDeclarations[typeName]; found {
		sourceCodeMap[absFilePath+"_"+typeName] = interfaceDecl.Raw
		for ref := range interfaceDecl.Reference {
			analyzeFileAndType(rootPath, absFilePath, ref, sourceCodeMap, visited)
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
				nextFile := importDecl.Source
				if !filepath.IsAbs(nextFile) {
					nextFile = filepath.Join(filepath.Dir(absFilePath), nextFile)
				}
				analyzeFileAndType(rootPath, nextFile, realTypeName, sourceCodeMap, visited)
			}
			// 命名空间导入
			if module.Type == "namespace" {
				refNameArr := strings.Split(typeName, ".")
				if len(refNameArr) == 2 && refNameArr[0] == module.Identifier {
					realTypeName := refNameArr[1]
					replaceTypeName := module.Identifier + "_" + realTypeName
					// 替换源码
					key := absFilePath + "_" + typeName
					if raw, ok := sourceCodeMap[key]; ok {
						sourceCodeMap[key] = strings.ReplaceAll(raw, typeName, replaceTypeName)
					}
					nextFile := importDecl.Source
					if !filepath.IsAbs(nextFile) {
						nextFile = filepath.Join(filepath.Dir(absFilePath), nextFile)
					}
					analyzeFileAndType(rootPath, nextFile, realTypeName, sourceCodeMap, visited)
				}
			}
		}
	}
}

// 入口方法
func GenerateBundle2() {
	rootPath := "/Users/zxc/Desktop/shopline-live-sale"
	inputAnalyzeFile := "/Users/zxc/Desktop/shopline-live-sale/src/feature/LiveRoom/components/MainLeft/ProductSet/AddProductSetPicker/index.tsx"
	inputAnalyzeType := "Name"

	sourceCodeMap := make(map[string]string)
	visited := make(map[string]bool)

	analyzeFileAndType(rootPath, inputAnalyzeFile, inputAnalyzeType, sourceCodeMap, visited)

	resultCode := ""
	for _, value := range sourceCodeMap {
		resultCode += value + "\n"
	}
	utils.WriteResultToFile("./ts/output/result.ts", resultCode)
}
