// Package filenamer 提供生成输出文件名的公共逻辑，以供各个分析器复用。
package filenamer

import (
	"fmt"
	"path/filepath"
	"strings"
)

// GenerateOutputFileName 是一个公共函数，用于根据输入目录和分析类型生成标准化的输出文件名。
// 它封装了文件名生成的通用逻辑，供各个分析器调用，确保输出文件名的一致性。
// 参数:
//   - inputPath: 输入项目目录的路径。
//   - suffix:    - string: 生成的标准化文件名（不包含路径）。
func GenerateOutputFileName(inputPath, suffix string) string {
	// 1. 获取输入目录的基名（例如，"/path/to/my-project" -> "my-project"）。
	baseName := filepath.Base(inputPath)
	
	// 2. 将基名中的空格替换为下划线，以确保文件名的有效性。
	//    虽然不太常见，但 defensive programming 是好的。
	safeBaseName := strings.ReplaceAll(baseName, " ", "_")
	
	// 3. 拼接基名和后缀，并加上 .json 扩展名。
	//    例如，"my-project" 和 "analyze" 会生成 "my-project_analyze.json"。
	return fmt.Sprintf("%s_%s.json", safeBaseName, suffix)
}