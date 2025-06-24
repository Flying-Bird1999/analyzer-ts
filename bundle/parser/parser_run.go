package parser

import (
	"fmt"
	"os"
)

func Parser_run() {
	// inputDir := "/Users/bird/Desktop/alalyzer/analyzer-ts/ts/import.ts"
	inputDir := "/Users/zxc/Desktop/analyzer-ts/ts/import.ts"

	// 解析当前文件
	pr := NewParserResult(inputDir)
	pr.Traverse()
	parserResult := pr.GetResult()
	// 定义输出文件路径
	outputFilePath := "./bundle/parser/parser_output.txt"

	// 打开或创建文件
	file, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Printf("创建文件失败: %s\n", err)
		return
	}
	defer file.Close()

	// 遍历分析结果并写入文件
	// for _, v := range parserResult.ImportDeclarations {
	// 	// 写入文件路径
	// 	file.WriteString(fmt.Sprintf("Source: %s, Raw: %s\n", v.Source, v.Raw))
	// 	// 写入 ImportSpecifiers
	// 	for _, specifier := range v.ImportModules {
	// 		file.WriteString(fmt.Sprintf("  - ImportModule: %s, Type: %s, Identifier: %s\n", specifier.ImportModule, specifier.Type, specifier.Identifier))
	// 	}

	// 	file.WriteString(fmt.Sprintf("\n\n\n"))
	// }

	// 遍历分析结果并写入文件
	for _, v := range parserResult.TypeDeclarations {
		file.WriteString(fmt.Sprintf("Source: %s, Raw: %s\n", v.Identifier, v.Raw))
		for _, specifier := range v.Reference {
			file.WriteString(fmt.Sprintf("  - Identifier: %s, Location: %s\n", specifier.Identifier, specifier.Location))
		}
		file.WriteString(fmt.Sprintf("\n\n\n\n\n\n"))

	}
}
