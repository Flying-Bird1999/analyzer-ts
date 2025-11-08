//go:build project_management
// +build project_management

package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🏗️ TSMorphGo 项目管理示例")
	fmt.Println("=" + repeat("=", 50))

	// 示例1: 从真实React项目创建项目
	fmt.Println("\n📁 示例1: 基于文件系统创建项目")
	projectPath := "./demo-react-app"

	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:         projectPath,
		TargetExtensions: []string{".ts", ".tsx"},
		IgnorePatterns:   []string{"node_modules", "dist", ".git"},
		UseTsConfig:      true,
	})
	defer project.Close()

	files := project.GetSourceFiles()
	fmt.Printf("发现 %d 个 TypeScript 文件:\n", len(files))

	for i, file := range files {
		if i < 5 { // 只显示前5个
			relPath, _ := filepath.Rel(projectPath, file.GetFilePath())
			fmt.Printf("  - %s\n", relPath)
		}
	}
	if len(files) > 5 {
		fmt.Printf("  ... 还有 %d 个文件\n", len(files)-5)
	}

	// 示例2: 基于真实React项目创建项目
	fmt.Println("\n🚀 示例2: 基于真实React项目创建项目")

	realProjectPath := "/Users/bird/Desktop/alalyzer/analyzer-ts/tsmorphgo/examples/demo-react-app"

	realProject := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:         realProjectPath,
		IgnorePatterns:   []string{"node_modules", "dist", ".git", "build"},
		TargetExtensions: []string{".ts", ".tsx"},
		UseTsConfig:      true,
	})
	defer realProject.Close()

	realFiles := realProject.GetSourceFiles()
	fmt.Printf("真实React项目包含 %d 个TypeScript文件:\n", len(realFiles))

	// 分析项目中的模块
	fileCategories := make(map[string]int)
	componentFiles, utilFiles, typeFiles := 0, 0, 0

	for _, file := range realFiles {
		relPath, _ := filepath.Rel(realProjectPath, file.GetFilePath())

		// 分类文件
		if strings.Contains(relPath, "components/") {
			componentFiles++
			fileCategories["Components"]++
		} else if strings.Contains(relPath, "utils") || strings.Contains(relPath, "services") {
			utilFiles++
			fileCategories["Utils/Services"]++
		} else if strings.Contains(relPath, "types") {
			typeFiles++
			fileCategories["Types"]++
		} else {
			fileCategories["Other"]++
		}

		if len(realFiles) <= 10 { // 只显示前10个文件
			fmt.Printf("  - %s\n", relPath)
		}
	}

	if len(realFiles) > 10 {
		fmt.Printf("  ... 还有 %d 个文件\n", len(realFiles)-10)
	}

	fmt.Printf("\n项目文件分类:\n")
	for category, count := range fileCategories {
		fmt.Printf("  - %s: %d 个\n", category, count)
	}

	// 示例3: 获取特定文件并分析
	fmt.Println("\n🔍 示例3: 分析特定文件")

	utilsFile := project.GetSourceFile(filepath.Join(projectPath, "src/utils.ts"))
	if utilsFile != nil {
		fmt.Printf("utils.ts 文件信息:\n")
		fmt.Printf("  - 完整路径: %s\n", utilsFile.GetFilePath())

		// 统计导出的函数数量
		exportCount := 0
		utilsFile.ForEachDescendant(func(node tsmorphgo.Node) {
			// 查找export关键字
			if node.Kind == 148 { // ExportKeyword
				exportCount++
			}
		})
		fmt.Printf("  - 导出数量: %d\n", exportCount)
	}

	// 示例4: 动态添加文件到项目
	fmt.Println("\n➕ 示例4: 动态添加文件")
	dynamicContent := `
		// 动态添加的类型文件
		export interface ApiConfig {
			baseUrl: string;
			timeout: number;
			headers?: Record<string, string>;
		}

		export interface ApiResponse<T> {
			data: T;
			status: number;
			message: string;
		}
	`

	newFile, err := project.CreateSourceFile(
		filepath.Join(projectPath, "src/dynamic/types.ts"),
		dynamicContent,
		tsmorphgo.CreateSourceFileOptions{Overwrite: true},
	)
	if err != nil {
		log.Printf("创建文件失败: %v", err)
	} else {
		fmt.Printf("成功创建动态文件: %s\n", newFile.GetFilePath())
	}

	fmt.Println("\n✅ 项目管理示例完成!")
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}