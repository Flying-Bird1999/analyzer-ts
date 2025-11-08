//go:build ignore
// +build ignore

// 在当前目录执行即可： go run example_simple

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/ts_bundle"
)

func main() {
	// 获取绝对路径，确保 projectRoot 是绝对路径
	absProjectRoot, err := filepath.Abs("./testdata")
	if err != nil {
		log.Fatalf("获取项目根目录失败: %v", err)
	}
	projectRoot := absProjectRoot
	tempDir := "./temp_output"

	// 清理之前的输出
	os.RemoveAll(tempDir)

	fmt.Println("=== TypeScript 批量类型打包功能演示 ===\n")

	// 示例1：基础批量打包（合并到单个文件）
	fmt.Println("1. 基础批量打包（合并模式）")
	entryStrings1 := []string{
		filepath.Join(projectRoot, "src", "utils", "user.ts") + ":User",
		filepath.Join(projectRoot, "src", "utils", "user.ts") + ":AdminUser",
		filepath.Join(projectRoot, "src", "utils", "address.ts") + ":Address",
	}

	// 转换为绝对路径
	for i := range entryStrings1 {
		parts := strings.Split(entryStrings1[i], ":")
		absPath, err := filepath.Abs(parts[0])
		if err != nil {
			log.Fatalf("获取绝对路径失败: %v", err)
		}
		entryStrings1[i] = absPath + ":" + strings.Join(parts[1:], ":")
	}

	for i, entry := range entryStrings1 {
		fmt.Printf("  入口点 %d: %s\n", i+1, entry)
	}

	bundledContent1, err := ts_bundle.GenerateBatchBundleFromStrings(entryStrings1, projectRoot)
	if err != nil {
		log.Fatalf("批量打包失败: %v", err)
	}
	fmt.Printf("✅ 合并模式成功！内容长度: %d 字符\n", len(bundledContent1))

	// 示例2：带别名的批量打包
	fmt.Println("\n2. 带别名的批量打包（合并模式）")
	entryStrings2 := []string{
		filepath.Join(projectRoot, "src", "utils", "user.ts") + ":User:UserDTO",
		filepath.Join(projectRoot, "src", "utils", "common.ts") + ":CommonType:ConfigType",
		filepath.Join(projectRoot, "src", "index.ts") + ":UserProfile:Profile",
	}

	// 转换为绝对路径
	for i := range entryStrings2 {
		parts := strings.Split(entryStrings2[i], ":")
		absPath, err := filepath.Abs(parts[0])
		if err != nil {
			log.Fatalf("获取绝对路径失败: %v", err)
		}
		entryStrings2[i] = absPath + ":" + strings.Join(parts[1:], ":")
	}

	for i, entry := range entryStrings2 {
		fmt.Printf("  入口点 %d: %s\n", i+1, entry)
	}

	bundledContent2, err := ts_bundle.GenerateBatchBundleFromStrings(entryStrings2, projectRoot)
	if err != nil {
		log.Fatalf("带别名批量打包失败: %v", err)
	}
	fmt.Printf("✅ 别名合并模式成功！内容长度: %d 字符\n", len(bundledContent2))

	// 示例3：批量文件输出（每个类型独立文件）
	fmt.Println("\n3. 批量文件输出（独立文件模式）")
	entryStrings3 := []string{
		filepath.Join(projectRoot, "src", "utils", "user.ts") + ":User",
		filepath.Join(projectRoot, "src", "index.ts") + ":UserProfile:UserProfileDTO",
		filepath.Join(projectRoot, "src", "complex.ts") + ":UserWithoutAddress",
		filepath.Join(projectRoot, "src", "path-alias.ts") + ":PathAliasUser",
	}

	// 转换为绝对路径
	for i := range entryStrings3 {
		parts := strings.Split(entryStrings3[i], ":")
		absPath, err := filepath.Abs(parts[0])
		if err != nil {
			log.Fatalf("获取绝对路径失败: %v", err)
		}
		entryStrings3[i] = absPath + ":" + strings.Join(parts[1:], ":")
	}

	for i, entry := range entryStrings3 {
		fmt.Printf("  入口点 %d: %s\n", i+1, entry)
	}

	results, err := ts_bundle.GenerateBatchBundlesToFiles(entryStrings3, projectRoot, tempDir)
	if err != nil {
		log.Fatalf("批量文件输出失败: %v", err)
	}

	fmt.Printf("✅ 独立文件模式成功！生成了 %d 个文件到目录: %s\n", len(results), tempDir)
	for _, result := range results {
		fmt.Printf("  - %s (%d 字符)\n", result.FileName, result.ContentSize)
	}

	// 展示文件内容
	fmt.Println("\n4. 生成的文件内容预览:")
	for _, result := range results {
		fmt.Printf("\n📄 %s:\n", result.FileName)
		content, err := os.ReadFile(result.FilePath)
		if err != nil {
			fmt.Printf("  读取失败: %v\n", err)
			continue
		}

		contentStr := string(content)
		if len(contentStr) > 300 {
			fmt.Printf("  内容预览 (前300字符):\n%s\n", contentStr[:300])
		} else {
			fmt.Printf("  完整内容:\n%s\n", contentStr)
		}
	}

	// 清理临时文件
	os.RemoveAll(tempDir)

	fmt.Println("\n=== 演示完成 ===")
	fmt.Println("🎉 所有批量打包功能测试通过！")
}
