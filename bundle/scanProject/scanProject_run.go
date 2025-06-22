package scanProject

import (
	"fmt"
	"os"
)

func ScanProject() {
	inputDir := "/Users/bird/company/sc1.0/mc/message-center/client"
	// 定义输出文件路径
	outputFilePath := "./bundle/scanProject/scanProject_output.txt"

	pr := NewProjectResult(inputDir, []string{"sc-components/**", "sl-utils-node/**"}, false)

	pr.ScanFileList()

	// 打开或创建文件
	file, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Printf("创建文件失败: %s\n", err)
		return
	}
	defer file.Close()

	// 遍历分析结果并写入文件
	for k, v := range pr.GetFileList() {
		// 写入文件路径
		_, err := file.WriteString(fmt.Sprintf("file: %s, fileName: %s, size: %d, 后缀: %s\n", k, v.FileName, v.Size, v.Ext))
		if err != nil {
			fmt.Printf("写入文件失败: %s\n", err)
			return
		}

		file.WriteString(fmt.Sprintf("\n\n\n"))
	}

	fmt.Printf("分析结果已写入文件: %s\n", outputFilePath)
}
