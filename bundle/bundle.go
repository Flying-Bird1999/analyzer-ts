package bundle

import (
	"fmt"
	"main/bundle/parser"
	"main/bundle/scanProject"
	"main/bundle/utils"
	"os"
	"path/filepath"

	"github.com/samber/lo"
)

type Bundle struct {
	//
}

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
				getCode(Result, refName, importPath, sourceCodeMap)
			} else {
				fmt.Printf("导入路径 %s 未找到对应的解析结果\n", importPath)
			}
		}
	}
	fmt.Printf("引用类型 %s 未找到\n", refName)
}

// 获取代码的主逻辑
func getCode(Result map[string]parser.ParserResult, targetTypeName string, targetPath string, sourceCodeMap *map[string]string) {
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
	filePath, _ := filepath.Abs("./ts/demo")

	Result := make(map[string]parser.ParserResult)

	// 扫描项目
	projectResult := scanProject.NewProjectResult(filePath, []string{})
	projectResult.ScanProject()

	for _, item := range projectResult.GetFileList() {
		// fmt.Printf("开始解析文件: %s\n", item.Path)
		pr := parser.NewBundleResult(item.Path)
		pr.Traverse()
		Result[item.Path] = pr.GetResult()
	}

	// 打印解析结果（调试用）
	fmt.Println("解析完成，结果如下:")
	for path, result := range Result {
		fmt.Printf("文件: %s, 解析结果: %+v, %+v, %+v\n", path, result.ImportDeclarations, result.InterfaceDeclarations, result.TypeDeclarations)
	}

	// 1. 在 Result 中找到 targetPath 的 ParserResult
	// 2. 在 ParserResult 中找到 targetTypeName，可能在 TypeDeclarationResult，也可能在 InterfaceDeclarationResult
	// 3. 看 Reference，是否有值，
	//     - 没有值，输出 Raw
	//     - 有值，遍历 Reference, 查找引用的类型
	//         - 1. 在 InterfaceDeclarationResult / TypeDeclarations 中查找
	//         - 2. 在 ImportDeclarations 中查找, 结合继续 1 的步骤

	targetPath, _ := filepath.Abs("./ts/demo/index.ts")
	targetTypeName := "Class"
	var sourceCodeMap = make(map[string]string)

	getCode(Result, targetTypeName, targetPath, &sourceCodeMap)
	fmt.Println("最终的代码：")

	for _, item := range sourceCodeMap {
		fmt.Println(item)
	}

	// 定义文件路径
	resultFilePath := "./bundle/result.ts"

	// 打开或创建文件
	file, err := os.Create(resultFilePath)
	if err != nil {
		fmt.Printf("创建文件失败: %s\n", err)
		return
	}
	defer file.Close()

	// 写入 sourceCodeMap 数据
	for _, value := range sourceCodeMap {
		_, err := file.WriteString(fmt.Sprintf("%s\n", value))
		if err != nil {
			fmt.Printf("写入文件失败: %s\n", err)
			return
		}
	}
}
