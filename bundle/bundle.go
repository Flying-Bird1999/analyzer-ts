package bundle

import (
	"fmt"
	"main/bundle/parser"
	"main/bundle/scanProject"
	"main/bundle/utils"
	"path/filepath"

	"github.com/samber/lo"
)

// 处理引用的逻辑
func processReference(refName string, parserResult parser.ParserResult, Result map[string]parser.ParserResult, targetPath string, sourceCodeMap *map[string]string) {
	// 在 TypeDeclarations 中查找引用的类型
	if refTypeDecl, found := parserResult.TypeDeclarations[refName]; found {
		(*sourceCodeMap)[targetPath+"_"+refName] = refTypeDecl.Raw
		// 在目标文件中递归查找引用的类型
		if len(refTypeDecl.Reference) == 0 {
		} else {
			for refName := range refTypeDecl.Reference {
				processReference(refName, parserResult, Result, targetPath, sourceCodeMap)
			}
		}
	}

	// 在 InterfaceDeclarations 中查找引用的接口
	if refInterfaceDecl, found := parserResult.InterfaceDeclarations[refName]; found {
		(*sourceCodeMap)[targetPath+"_"+refName] = refInterfaceDecl.Raw
		// 在目标文件中递归查找引用的类型
		if len(refInterfaceDecl.Reference) == 0 {
		} else {
			for refName := range refInterfaceDecl.Reference {
				processReference(refName, parserResult, Result, targetPath, sourceCodeMap)
			}
		}
	}

	// 在 ImportDeclarations 中查找引用的类型
	for _, importDecl := range parserResult.ImportDeclarations {
		if utils.Contains(lo.Map(importDecl.Modules, func(it parser.Module, index int) string {
			return it.Identifier
		}), refName) {
			// 根据导入路径查找目标文件
			importPath, _ := filepath.Abs(filepath.Join(filepath.Dir(targetPath), importDecl.Source))
			if _, exists := Result[importPath]; exists {
				analyze(Result, refName, importPath, sourceCodeMap)
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
func analyze(Result map[string]parser.ParserResult, targetTypeName string, targetPath string, sourceCodeMap *map[string]string) {
	// 在 Result 中找到 targetPath 的 ParserResult
	parserResult, exists := Result[targetPath]
	if !exists {
		fmt.Printf("目标文件 %s 未在解析结果中找到\n", targetPath)
	}

	// 在 ParserResult 中找到 targetTypeName
	if typeDecl, found := parserResult.TypeDeclarations[targetTypeName]; found {
		(*sourceCodeMap)[targetPath+"_"+targetTypeName] = typeDecl.Raw
		if len(typeDecl.Reference) == 0 {
		} else {
			for refName := range typeDecl.Reference {
				processReference(refName, parserResult, Result, targetPath, sourceCodeMap)
			}
		}
	} else if interfaceDecl, found := parserResult.InterfaceDeclarations[targetTypeName]; found {
		(*sourceCodeMap)[targetPath+"_"+targetTypeName] = interfaceDecl.Raw
		if len(interfaceDecl.Reference) == 0 {
		} else {
			for refName := range interfaceDecl.Reference {
				processReference(refName, parserResult, Result, targetPath, sourceCodeMap)
			}
		}
	} else {
		fmt.Printf("目标类型 %s 未在文件 %s 中找到\n", targetTypeName, targetPath)
	}
}

func GenerateBundle() {
	// inputAnalyzeDir := "/Users/zxc/Desktop/shopline-order-detail"
	// inputAnalyzeFile := "/Users/zxc/Desktop/shopline-order-detail/src/interface/preloadedState/index.ts"
	// inputAnalyzeType := "PreloadedState"

	inputAnalyzeDir := "./ts/demo"
	inputAnalyzeFile := "./ts/demo/index.ts"
	inputAnalyzeType := "Class"

	filePath, _ := filepath.Abs(inputAnalyzeDir)

	Result := make(map[string]parser.ParserResult)

	// 扫描项目
	projectResult := scanProject.NewProjectResult(filePath, []string{})
	projectResult.ScanProject()

	for _, item := range projectResult.GetFileList() {
		pr := parser.NewParserResult(item.Path)
		pr.Traverse()
		Result[item.Path] = pr.GetResult()
	}

	// fmt.Println("解析完成，结果如下:")
	// for path, result := range Result {
	// 	fmt.Printf("文件: %s, 解析结果: %+v, %+v, %+v\n", path, result.ImportDeclarations, result.InterfaceDeclarations, result.TypeDeclarations)
	// }

	var sourceCodeMap = make(map[string]string)
	targetPath, _ := filepath.Abs(inputAnalyzeFile)
	analyze(Result, inputAnalyzeType, targetPath, &sourceCodeMap)

	resultCode := ""
	for _, value := range sourceCodeMap {
		resultCode += value + "\n"
	}
	utils.WriteResultToFile("./ts/output/result.ts", resultCode)
}
