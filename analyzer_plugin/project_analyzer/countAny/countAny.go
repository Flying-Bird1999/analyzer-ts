// package countany 包含了用于统计项目中 'any' 类型使用情况的核心业务逻辑。
package countany

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"main/analyzer/parser"
	"main/analyzer_plugin/project_analyzer/internal/filenamer"
	internalparser "main/analyzer_plugin/project_analyzer/internal/parser"
	"os"
	"path/filepath"
)

// CountResult 是 'any' 类型分析的最终结果的顶层结构体。
type CountResult struct {
	FilesParsed   int         `json:"filesParsed"`   // 成功解析的 JS/TS 文件数量
	TotalAnyCount int         `json:"totalAnyCount"` // 项目中 'any' 类型的总数
	FileCounts    []FileCount `json:"fileCounts"`    // 每个文件的 'any' 类型统计列表
}

// FileCount 存储单个文件中 'any' 类型的使用情况统计。
type FileCount struct {
	FilePath string           `json:"filePath"` // 文件的绝对路径
	AnyCount int              `json:"anyCount"` // 该文件中的 'any' 类型总数
	Details  []parser.AnyInfo `json:"details"`  // 该文件中所有 'any' 类型的详细信息列表
}

// CountAnyUsages 分析一个 TypeScript 项目，以统计 'any' 类型的使用情况。
// 它返回一个包含详细信息的 CountResult 结构体。
// rootPath: 要分析的项目根目录。
// excludePatterns: 需要从分析中排除的文件/目录的 glob 模式列表。
// isMonorepo: 指示项目是否为 monorepo。
func CountAnyUsages(rootPath string, excludePatterns []string, isMonorepo bool) *CountResult {
	// 1. 使用新的 parser 包解析项目
	result, err := internalparser.ParseProject(rootPath, excludePatterns, isMonorepo)
	if err != nil {
		// 在实际的生产代码中，应该更优雅地处理这个错误。
		// 但为了保持与原函数签名一致（返回 *CountResult），这里直接panic或返回空结果。
		fmt.Printf("解析项目失败: %v\n", err)
		// 返回一个空的结果，而不是nil，以避免调用者出现空指针异常。
		return &CountResult{}
	}

	totalAnyCount := 0
	var fileCounts []FileCount

	// 4. 遍历已解析的数据，提取 'any' 的信息
	for filePath, fileData := range result.Js_Data {
		anyCountInFile := len(fileData.AnyDeclarations)

		if anyCountInFile > 0 {
			fileCounts = append(fileCounts, FileCount{
				FilePath: filePath,
				AnyCount: anyCountInFile,
				Details:  fileData.AnyDeclarations,
			})
		}
		totalAnyCount += anyCountInFile
	}

	return &CountResult{
		TotalAnyCount: totalAnyCount,
		FileCounts:    fileCounts,
		FilesParsed:   len(result.Js_Data),
	}
}

// WriteOutput 将分析结果以 JSON 格式写入指定的输出目录。
// 文件名将根据输入目录的名称动态生成，例如 "my-project_any_count.json"。
// inputPath: 输入项目目录的路径。
// outputDir: 用于存储 JSON 输出文件的目录路径。
// result: 要写入的分析结果。
func WriteOutput(outputDir string, inputPath string, result *CountResult) error {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化JSON时出错: %w", err)
	}

	// 确保输出目录存在
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("创建输出目录失败: %w", err)
		}
	}

	// 使用新的 filenamer 包构建输出文件路径
	outputFileName := filenamer.GenerateOutputFileName(inputPath, "any_count")
	outputFile := filepath.Join(outputDir, outputFileName)

	// 写入文件
	if err := ioutil.WriteFile(outputFile, jsonData, 0644); err != nil {
		return fmt.Errorf("写入JSON文件失败: %w", err)
	}

	fmt.Printf("结果已成功写入到 %s\n", outputFile)
	return nil
}
