package parser

import (
	"encoding/json"
	"fmt"
	"os"
)

func Parser_run() {
	inputDir := "/Users/bird/Desktop/alalyzer/analyzer-ts/ts_example/variable.ts"

	// 解析当前文件
	pr := NewParserResult(inputDir)
	pr.Traverse()
	parserResult := pr.GetResult()

	// 定义输出文件路径
	outputFilePath := "./analyzer/parser/parser_output.json"

	// 打开或创建文件
	file, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Printf("创建文件失败: %s\n", err)
		return
	}
	defer file.Close()

	// 创建一个包含所有你想要输出的结果的结构体
	output := struct {
		VariableDeclarations []VariableDeclaration `json:"variableDeclarations"`
		// CallExpressions      []CallExpression      `json:"callExpressions"`
		JsxElements []JSXElement `json:"jsxElements"`
	}{
		VariableDeclarations: parserResult.VariableDeclarations,
		// CallExpressions:      parserResult.CallExpressions,
		JsxElements: parserResult.JsxElements,
	}

	// 将整个 output 结构体序列化为 JSON
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Printf("JSON 序列化失败: %s\n", err)
		return
	}

	// 将 JSON 数据写入文件
	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Printf("写入文件失败: %s\n", err)
	}

	fmt.Println("解析结果已写入到", outputFilePath)
}
