package ts_bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GenerateBundle 单类型打包入口方法（向后兼容）
func GenerateBundle(inputAnalyzeFile string, inputAnalyzeType string, projectRoot string) (string, error) {
	br := NewCollectResult(inputAnalyzeFile, inputAnalyzeType, projectRoot)
	if err := br.collectFileType(inputAnalyzeFile, inputAnalyzeType, "", ""); err != nil {
		return "", err
	}
	bundler := NewTypeBundler()
	bundledContent, err := bundler.Bundle(br.SourceCodeMap)
	return bundledContent, err
}

// GenerateBatchBundle 多类型批量打包入口方法
// 支持一次处理多个入口文件和类型，通过缓存优化性能
func GenerateBatchBundle(entryPoints []TypeEntryPoint, projectRoot string) (string, error) {
	if len(entryPoints) == 0 {
		return "", nil
	}

	// 创建批量收集器
	bcr := NewBatchCollectResult(entryPoints, projectRoot)

	// 执行批量收集
	if err := bcr.CollectBatch(); err != nil {
		return "", err
	}

	// 创建打包器并打包所有类型
	bundler := NewTypeBundler()
	bundledContent, err := bundler.Bundle(bcr.GetAllTypes())
	return bundledContent, err
}

// GenerateBatchBundleFromStrings 从字符串切片创建入口点并打包
// 便捷方法：传入 "文件路径:类型名[:别名]" 格式的字符串切片
func GenerateBatchBundleFromStrings(entryStrings []string, projectRoot string) (string, error) {
	var entryPoints []TypeEntryPoint

	for _, entryStr := range entryStrings {
		parts := strings.Split(entryStr, ":")
		if len(parts) < 2 || len(parts) > 3 {
			return "", fmt.Errorf("无效的入口格式: %s，期望格式为 '文件路径:类型名[:别名]'", entryStr)
		}

		filePath := strings.TrimSpace(parts[0])
		typeName := strings.TrimSpace(parts[1])

		// 验证必要字段
		if filePath == "" {
			return "", fmt.Errorf("文件路径不能为空: %s", entryStr)
		}
		if typeName == "" {
			return "", fmt.Errorf("类型名不能为空: %s", entryStr)
		}

		entry := TypeEntryPoint{
			FilePath: filePath,
			TypeName: typeName,
		}

		// 如果提供了别名
		if len(parts) == 3 {
			entry.Alias = strings.TrimSpace(parts[2])
		}

		entryPoints = append(entryPoints, entry)
	}

	return GenerateBatchBundle(entryPoints, projectRoot)
}

// BatchFileResult 批量文件输出结果
type BatchFileResult struct {
	EntryPoint  TypeEntryPoint // 入口点信息
	FileName    string         // 输出的文件名
	FilePath    string         // 输出的完整文件路径
	ContentSize int            // 内容大小（字符数）
}

// GenerateBatchBundlesToFiles 批量生成类型包到独立文件
// 每个类型生成一个独立的 .d.ts 文件，避免命名冲突
func GenerateBatchBundlesToFiles(entryStrings []string, projectRoot string, outputDir string) ([]BatchFileResult, error) {
	var results []BatchFileResult

	// 确保输出目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 解析入口点
	var entryPoints []TypeEntryPoint
	for _, entryStr := range entryStrings {
		parts := strings.Split(entryStr, ":")
		if len(parts) < 2 || len(parts) > 3 {
			return nil, fmt.Errorf("无效的入口格式: %s，期望格式为 '文件路径:类型名[:别名]'", entryStr)
		}

		filePath := strings.TrimSpace(parts[0])
		typeName := strings.TrimSpace(parts[1])

		// 验证必要字段
		if filePath == "" {
			return nil, fmt.Errorf("文件路径不能为空: %s", entryStr)
		}
		if typeName == "" {
			return nil, fmt.Errorf("类型名不能为空: %s", entryStr)
		}

		entry := TypeEntryPoint{
			FilePath: filePath,
			TypeName: typeName,
		}

		// 如果提供了别名
		if len(parts) == 3 {
			entry.Alias = strings.TrimSpace(parts[2])
		}

		entryPoints = append(entryPoints, entry)
	}

	// 为每个入口点生成独立的类型包
	for _, entry := range entryPoints {
		// 确保文件路径是绝对路径
		absFilePath, err := filepath.Abs(entry.FilePath)
		if err != nil {
			return nil, fmt.Errorf("解析文件路径失败 %s: %v", entry.FilePath, err)
		}

		// 生成单个类型包
		bundledContent, err := GenerateBundle(absFilePath, entry.TypeName, projectRoot)
		if err != nil {
			return nil, fmt.Errorf("生成类型包失败 %s:%s: %v", entry.FilePath, entry.TypeName, err)
		}

		// 如果内容为空（类型不存在），跳过
		if strings.TrimSpace(bundledContent) == "" {
			continue
		}

		// 生成文件名
		fileName := generateFileName(entry, outputDir)
		filePath := filepath.Join(outputDir, fileName)

		// 写入文件
		if err := os.WriteFile(filePath, []byte(bundledContent), 0644); err != nil {
			return nil, fmt.Errorf("写入文件失败 %s: %v", filePath, err)
		}

		results = append(results, BatchFileResult{
			EntryPoint:  entry,
			FileName:    fileName,
			FilePath:    filePath,
			ContentSize: len(bundledContent),
		})
	}

	return results, nil
}

// generateFileName 根据入口点信息生成文件名
// 规则：别名 > 类型名，使用 .d.ts 扩展名
func generateFileName(entry TypeEntryPoint, outputDir string) string {
	// 确定基础文件名
	baseName := entry.TypeName
	if entry.Alias != "" {
		baseName = entry.Alias
	}

	// 清理文件名，移除非法字符
	cleanName := regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(baseName, "_")
	fileName := fmt.Sprintf("%s.d.ts", cleanName)

	// 确保文件名唯一
	counter := 1
	for {
		fullPath := filepath.Join(outputDir, fileName)
		if _, exists := os.Stat(fullPath); os.IsNotExist(exists) {
			break // 文件不存在，可以使用
		}

		// 文件已存在，添加计数器
		fileName = fmt.Sprintf("%s_%d.d.ts", cleanName, counter)
		counter++
	}

	return fileName
}
