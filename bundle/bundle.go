package bundle

import (
	"fmt"
	analyzeModule "main/bundle/analyze"
	"main/bundle/utils"
	"path/filepath"
	"strings"
)

// 处理引用的逻辑
func processReference(refName string, parserResult analyzeModule.FileAnalyzeResult, Result map[string]analyzeModule.FileAnalyzeResult, targetPath string, targetTypeName string, sourceCodeMap *map[string]string) {
	// 在 TypeDeclarations 中查找引用的类型
	if refTypeDecl, found := parserResult.TypeDeclarations[refName]; found {
		(*sourceCodeMap)[targetPath+"_"+refName] = refTypeDecl.Raw
		// 在目标文件中递归查找引用的类型
		if len(refTypeDecl.Reference) != 0 {
			for refName := range refTypeDecl.Reference {
				processReference(refName, parserResult, Result, targetPath, targetTypeName, sourceCodeMap)
			}
		}
	}

	// 在 InterfaceDeclarations 中查找引用的接口
	if refInterfaceDecl, found := parserResult.InterfaceDeclarations[refName]; found {
		(*sourceCodeMap)[targetPath+"_"+refName] = refInterfaceDecl.Raw
		// 在目标文件中递归查找引用的类型
		if len(refInterfaceDecl.Reference) != 0 {
			for refName := range refInterfaceDecl.Reference {
				processReference(refName, parserResult, Result, targetPath, targetTypeName, sourceCodeMap)
			}
		}
	}

	// 在 ImportDeclarations 中查找引用的类型
	for _, importDecl := range parserResult.ImportDeclarations {
		// fmt.Printf("refName: %s\n", refName)
		// fmt.Printf("importDecl.Raw: %s\n", importDecl.Raw)
		for _, module := range importDecl.ImportModules {
			if module.Identifier == refName {
				realRefName := refName
				var replaceTypeName *string
				// case: import { School as NewSchool } from './school';
				if module.Type == "named" && module.ImportModule != refName {
					realRefName = module.ImportModule
					replaceTypeName = &refName
				}

				// 根据导入路径查找目标文件
				if _, exists := Result[importDecl.Source.FilePath]; exists {
					analyze(Result, realRefName, replaceTypeName, importDecl.Source.FilePath, sourceCodeMap)
				}
			}

			// case: import * as allTypes from './type';
			if module.Type == "namespace" {
				// 解析refName: allTypes.MerchantData。提取出 allTypes.MerchantData 中的 MerchantData
				refNameArr := strings.Split(refName, ".")
				realRefName := refNameArr[len(refNameArr)-1] // MerchantData
				if refNameArr[0] == module.Identifier {      // allTypes
					var replaceTypeName = module.Identifier + "_" + realRefName // allTypes_MerchantData
					// 替换源码的类型， allTypes.MerchantData -> allTypes_MerchantData
					realTargetTypeRaw := strings.ReplaceAll((*sourceCodeMap)[targetPath+"_"+targetTypeName], refName, replaceTypeName)
					(*sourceCodeMap)[targetPath+"_"+targetTypeName] = realTargetTypeRaw
					// 根据导入路径查找目标文件
					if _, exists := Result[importDecl.Source.FilePath]; exists {
						analyze(Result, realRefName, &replaceTypeName, importDecl.Source.FilePath, sourceCodeMap)
					}
				}
			}
		}
	}
}

// 依赖分析逻辑
// 1. 在 Result 中找到 targetPath 的 ParserResult
// 2. 在 ParserResult 中找到 targetTypeName，可能在 TypeDeclarationResult，也可能在 InterfaceDeclarationResult
// 3. 看 Reference，是否有值，
//   - 没有值，输出 Raw
//   - 有值，遍历 Reference, 查找引用的类型
//   - 1. 在 InterfaceDeclarationResult / TypeDeclarations 中查找
//   - 2. 在 ImportDeclarations 中查找, 结合继续 1 的步骤
func analyze(Result map[string]analyzeModule.FileAnalyzeResult, targetTypeName string, replaceTypeName *string, targetPath string, sourceCodeMap *map[string]string) {
	// 在 Result 中找到 targetPath 的 ParserResult
	parserResult, exists := Result[targetPath]
	if !exists {
		fmt.Printf("目标文件 %s 未在解析结果中找到\n", targetPath)
	}

	// 在 ParserResult 中找到 targetTypeName
	if typeDecl, found := parserResult.TypeDeclarations[targetTypeName]; found {
		realRaw := typeDecl.Raw
		if replaceTypeName != nil {
			realRaw = strings.ReplaceAll(typeDecl.Raw, targetTypeName, *replaceTypeName)
		}
		(*sourceCodeMap)[targetPath+"_"+targetTypeName] = realRaw
		if len(typeDecl.Reference) != 0 {
			for refName := range typeDecl.Reference {
				processReference(refName, parserResult, Result, targetPath, targetTypeName, sourceCodeMap)
			}
		}
	} else if interfaceDecl, found := parserResult.InterfaceDeclarations[targetTypeName]; found {
		realRaw := interfaceDecl.Raw
		if replaceTypeName != nil {
			realRaw = strings.ReplaceAll(interfaceDecl.Raw, targetTypeName, *replaceTypeName)
		}
		(*sourceCodeMap)[targetPath+"_"+targetTypeName] = realRaw
		if len(interfaceDecl.Reference) != 0 {
			for refName := range interfaceDecl.Reference {
				processReference(refName, parserResult, Result, targetPath, targetTypeName, sourceCodeMap)
			}
		}
	} else {
		fmt.Printf("目标类型 %s 未在文件 %s 中找到\n", targetTypeName, targetPath)
	}
}

func GenerateBundle() {
	inputAnalyzeDir := "/Users/zxc/Desktop/shopline-order-detail"
	inputAnalyzeFile := "/Users/zxc/Desktop/shopline-order-detail/src/interface/preloadedState/index.ts"
	inputAnalyzeType := "PreloadedState"

	ar := analyzeModule.NewAnalyzeResult(inputAnalyzeDir, nil, nil)
	ar.Analyze()
	fileData := ar.GetFileData()

	var sourceCodeMap = make(map[string]string)
	targetPath, _ := filepath.Abs(inputAnalyzeFile)
	analyze(fileData, inputAnalyzeType, nil, targetPath, &sourceCodeMap)

	resultCode := ""
	for _, value := range sourceCodeMap {
		resultCode += value + "\n"
	}
	utils.WriteResultToFile("./ts/output/result.ts", resultCode)
}
