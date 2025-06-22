package projectParser

import (
	"fmt"
	"os"
)

func ProjectParser_run() {
	// inputDir := "/Users/zxc/Desktop/shopline-live-sale"
	// ar := NewAnalyzeResult(inputDir, nil, nil, false)

	// inputDir := "/Users/zxc/Desktop/message-center/client"
	// inputDir := "/Users/bird/company/sc1.0/mc/message-center/client"
	inputDir := "/Users/bird/company/sc1.0/components/nova"
	ar := NewAnalyzeResult(inputDir, nil, nil, []string{"node_modules/**", "sc-components/**"}, false)

	ar.ProjectParser()
	// 定义输出文件路径
	outputFilePath := "./bundle/projectParser/projectParser_output.txt"

	// 打开或创建文件
	file, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Printf("创建文件失败: %s\n", err)
		return
	}
	defer file.Close()

	// // 遍历分析结果并写入文件
	// for k, v := range ar.Js_File {
	// 	// 写入文件路径
	// 	_, err := file.WriteString(fmt.Sprintf("file: %s\n", k))
	// 	if err != nil {
	// 		fmt.Printf("写入文件失败: %s\n", err)
	// 		return
	// 	}

	// 	file.WriteString(fmt.Sprintf("ImportDeclarations👇👇👇\n"))

	// 	// 写入 ImportDeclarations
	// 	for _, v2 := range v.ImportDeclarations {
	// 		_, err := file.WriteString(fmt.Sprintf("FilePath: %s, Type: %s\n", v2.Source.FilePath, v2.Source.Type))
	// 		if err != nil {
	// 			fmt.Printf("写入文件失败: %s\n", err)
	// 			return
	// 		}
	// 	}

	// 	file.WriteString(fmt.Sprintf("\n\n\n"))
	// }

	// 遍历分析结果并写入文件
	for k, v := range ar.Package_Data {
		// 写入文件路径
		_, err := file.WriteString(fmt.Sprintf("file: %s\n", k))
		if err != nil {
			fmt.Printf("写入文件失败: %s\n", err)
			return
		}

		// 写入 具体信息
		file.WriteString(fmt.Sprintf("Namespace: %s, Version: %s, Workspace: %s\n", v.Namespace, v.Version, v.Workspace))
		for _, v2 := range v.NpmList {
			file.WriteString(fmt.Sprintf("Name: %s, Version: %s,RealVersion: %s, Type: %s\n", v2.Name, v2.Version, v2.NodeModuleVersion, v2.Type))
		}
		file.WriteString(fmt.Sprintf("\n\n\n"))
	}

	fmt.Printf("分析结果已写入文件: %s\n", outputFilePath)
}
