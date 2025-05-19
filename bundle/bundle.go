package bundle

import (
	"fmt"
	"main/bundle/parser"
	"main/bundle/scanProject"
	"main/bundle/utils"
	"path/filepath"

	"github.com/samber/lo"
)

type Bundle struct {
	//
}

// 处理引用的逻辑
func processReference(refName string, parserResult parser.ParserResult, Result map[string]parser.ParserResult, targetPath string) string {
	// 在 TypeDeclarations 中查找引用的类型
	if refTypeDecl, found := parserResult.TypeDeclarations[refName]; found {
		fmt.Printf("引用类型 %s 的原始代码: %s\n", refName, refTypeDecl.Raw)
		return refTypeDecl.Raw
	}

	// 在 InterfaceDeclarations 中查找引用的接口
	if refInterfaceDecl, found := parserResult.InterfaceDeclarations[refName]; found {
		fmt.Printf("引用接口 %s 的原始代码: %s\n", refName, refInterfaceDecl.Raw)
		return refInterfaceDecl.Raw
	}

	// 在 ImportDeclarations 中查找引用的类型
	for _, importDecl := range parserResult.ImportDeclarations {
		if utils.Contains(lo.Map(importDecl.Modules, func(it parser.Module, index int) string {
			return it.Identifier
		}), refName) {
			fmt.Printf("引用类型 %s 的导入路径: %s\n", refName, importDecl.Source)

			// 根据导入路径查找目标文件
			importPath, _ := filepath.Abs(filepath.Join(filepath.Dir(targetPath), importDecl.Source))
			if importedParserResult, exists := Result[importPath]; exists {
				// 在目标文件中递归查找引用的类型
				return processReference(refName, importedParserResult, Result, importPath)
			} else {
				fmt.Printf("导入路径 %s 未找到对应的解析结果\n", importPath)
			}
		}
	}

	// 如果没有找到，返回空字符串
	fmt.Printf("引用类型 %s 未找到\n", refName)
	return ""
}

// 获取代码的主逻辑
func getCode(Result map[string]parser.ParserResult, targetTypeName string, targetPath string) string {
	var sourceCode string

	// 在 Result 中找到 targetPath 的 ParserResult
	parserResult, exists := Result[targetPath]
	if !exists {
		fmt.Printf("目标文件 %s 未在解析结果中找到\n", targetPath)
		return ""
	}

	// 在 ParserResult 中找到 targetTypeName
	if typeDecl, found := parserResult.TypeDeclarations[targetTypeName]; found {
		sourceCode += typeDecl.Raw
		if len(typeDecl.Reference) == 0 {
			fmt.Printf("目标类型 %s 的原始代码: %s\n", targetTypeName, typeDecl.Raw)
		} else {
			fmt.Printf("目标类型 %s 的引用信息:\n", targetTypeName)
			for refName := range typeDecl.Reference {
				sourceCode += processReference(refName, parserResult, Result, targetPath)
			}
		}
	} else if interfaceDecl, found := parserResult.InterfaceDeclarations[targetTypeName]; found {
		sourceCode += interfaceDecl.Raw
		if len(interfaceDecl.Reference) == 0 {
			fmt.Printf("目标接口 %s 的原始代码: %s\n", targetTypeName, interfaceDecl.Raw)
		} else {
			fmt.Printf("目标接口 %s 的引用信息:\n", targetTypeName)
			for refName := range interfaceDecl.Reference {
				sourceCode += processReference(refName, parserResult, Result, targetPath)
			}
		}
	} else {
		fmt.Printf("目标类型 %s 未在文件 %s 中找到\n", targetTypeName, targetPath)
	}

	return sourceCode
}

func GenerateBundle() {
	filePath, _ := filepath.Abs("./ts/demo")

	Result := make(map[string]parser.ParserResult)

	// 扫描项目
	projectResult := scanProject.NewProjectResult(filePath, []string{})
	projectResult.ScanProject()

	for _, item := range projectResult.GetFileList() {
		fmt.Printf("开始解析文件: %s\n", item.Path)
		pr := parser.NewBundleResult(item.Path)
		pr.Traverse()
		Result[item.Path] = pr.GetResult()
	}

	// 打印解析结果（调试用）
	fmt.Println("解析完成，结果如下:")
	for path, result := range Result {
		fmt.Printf("文件: %s, 解析结果: %+v, %+v, %+v\n", path, result.ImportDeclarations, result.InterfaceDeclarations, result.TypeDeclarations)
	}

	targetPath, _ := filepath.Abs("./ts/demo/index.ts")
	targetTypeName := "Class"

	// 1. 在 Result 中找到 targetPath 的 ParserResult
	// 2. 在 ParserResult 中找到 targetTypeName，可能在 TypeDeclarationResult，也可能在 InterfaceDeclarationResult
	// 3. 看 Reference，是否有值，
	//     - 没有值，输出 Raw
	//     - 有值，遍历 Reference, 查找引用的类型
	//         - 1. 在 InterfaceDeclarationResult / TypeDeclarations 中查找
	//         - 2. 在 ImportDeclarations 中查找, 结合继续 1 的步骤

	code := getCode(Result, targetTypeName, targetPath)
	fmt.Println("最终的代码：")
	fmt.Println(code)
}
