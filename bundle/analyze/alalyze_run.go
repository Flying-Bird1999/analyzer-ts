package analyze

import (
	"fmt"
	"os"
)

func Analyze() {
	inputAnalyzeDir := "/Users/zxc/Desktop/shopline-order-detail"

	ar := NewAnalyzeResult(inputAnalyzeDir, nil, nil)

	ar.Analyze()
	// 定义输出文件路径
	outputFilePath := "./bundle/analyze/analyze_output.txt"

	// 打开或创建文件
	file, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Printf("创建文件失败: %s\n", err)
		return
	}
	defer file.Close()

	// 遍历分析结果并写入文件
	for k, v := range ar.File {
		// 写入文件路径
		_, err := file.WriteString(fmt.Sprintf("file: %s\n", k))
		if err != nil {
			fmt.Printf("写入文件失败: %s\n", err)
			return
		}

		// 写入 ImportDeclarations
		for _, v2 := range v.ImportDeclarations {
			_, err := file.WriteString(fmt.Sprintf("FilePath: %s, Type: %s\n", v2.Source.FilePath, v2.Source.Type))
			if err != nil {
				fmt.Printf("写入文件失败: %s\n", err)
				return
			}
		}

		file.WriteString(fmt.Sprintf("\n\n\n"))
	}

	fmt.Printf("分析结果已写入文件: %s\n", outputFilePath)
}
