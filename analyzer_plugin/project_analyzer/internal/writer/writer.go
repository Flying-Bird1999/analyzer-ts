// Package writer 提供将分析结果序列化为 JSON 并写入文件的公共逻辑，以供各个分析器复用。
package writer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WriteJSONResult 是一个公共函数，用于将分析结果序列化为 JSON 并写入文件。
// 它封装了 JSON 序列化和文件写入的通用逻辑，供各个分析器调用，避免代码重复。
// 参数:
//   - outputDir: 用于存储 JSON 输出文件的目录路径。
//   - fileName: 输出文件的名称（不包含路径）。
//   - result:   - error: 写入过程中可能发生的错误。
func WriteJSONResult(outputDir, fileName string, result interface{}) error {
	// 步骤 1: 将结果序列化为格式化的 JSON 字节。
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化JSON时出错: %w", err)
	}

	// 步骤 2: 确保输出目录存在。
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 步骤 3: 构建完整的输出文件路径。
	outputFile := filepath.Join(outputDir, fileName)

	// 步骤 4: 将 JSON 数据写入文件。
	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		return fmt.Errorf("写入JSON文件失败: %w", err)
	}

	// 步骤 5: 打印成功信息到标准输出。
	fmt.Printf("结果已成功写入到 %s\n", outputFile)
	return nil
}
