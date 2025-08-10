// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（parser_run.go）提供了一个可执行的入口点，用于手动触发对指定文件的解析，
// 并将解析结果以 JSON 格式输出到文件中。这对于调试和验证解析器的功能非常有用。
package parser

import (
	"encoding/json"
	"fmt"
	"os"
)

// Parser_run 是一个独立的运行函数，用于演示和测试文件解析功能。
// 1. 它硬编码了一个输入文件路径（`inputDir`）。
// 2. 调用解析器核心逻辑（`NewParserResult`, `Traverse`）来分析该文件。
// 3. 将解析结果的一部分（变量声明、JSX元素等）序列化为格式化的 JSON。
// 4. 将生成的 JSON 写入到指定的输出文件（`parser_output.json`）中。
func Parser_run() {
	// 定义要解析的目标文件路径。
	inputDir := "/Users/bird/Desktop/alalyzer/analyzer-ts/ts_example/variable.ts"

	// 创建一个新的解析器结果实例并启动遍历解析过程。
	pr := NewParserResult(inputDir)
	pr.Traverse()
	parserResult := pr.GetResult()

	// 定义输出 JSON 文件的路径。
	outputFilePath := "./analyzer/parser/parser_output.json"

	// 创建或覆盖输出文件。
	file, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Printf("创建文件失败: %s\n", err)
		return
	}
	// 确保在函数结束时关闭文件句柄。
	defer file.Close()

	// 创建一个匿名结构体，用于选择性地组织需要输出到 JSON 的解析结果。
	// 这样可以灵活地控制最终输出的内容。
	output := struct {
		VariableDeclarations []VariableDeclaration `json:"variableDeclarations"`
		// CallExpressions      []CallExpression      `json:"callExpressions"` // 此处被注释掉，不会输出到 JSON 中
		JsxElements []JSXElement `json:"jsxElements"`
	}{
		VariableDeclarations: parserResult.VariableDeclarations,
		// CallExpressions:      parserResult.CallExpressions,
		JsxElements: parserResult.JsxElements,
	}

	// 将 output 结构体序列化为易于阅读的、带缩进的 JSON 格式。
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Printf("JSON 序列化失败: %s\n", err)
		return
	}

	// 将 JSON 数据写入到文件中。
	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Printf("写入文件失败: %s\n", err)
	}

	fmt.Println("解析结果已写入到", outputFilePath)
}