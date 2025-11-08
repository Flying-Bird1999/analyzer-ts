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

	// 示例1: 从内存中的源码创建项目
	// 这在测试或需要快速分析独立代码片段时非常有用。
	// 对应 ts-morph 的 `new Project({ useInMemoryFileSystem: true })` 和 `project.createSourceFile()`。
	fmt.Println("\n🧠 示例1: 基于内存源码创建项目")
	memoryProject := tsmorphgo.NewProjectFromSources(map[string]string{
		"/main.ts": `
			import { Greeter } from './greeter';
			const greeter = new Greeter('World');
			console.log(greeter.greet());
		`,
		"/greeter.ts": `
			export class Greeter {
				private greeting: string;
				constructor(message: string) {
					this.greeting = message;
				}
				public greet() {
					return 'Hello, ' + this.greeting;
				}
			}
		`,
	})
	defer memoryProject.Close() // 确保资源被释放

	// GetSourceFiles 获取项目中的所有源文件。
	memFiles := memoryProject.GetSourceFiles()
	fmt.Printf("内存项目包含 %d 个文件:\n", len(memFiles))
	for _, file := range memFiles {
		// GetFilePath 获取源文件的完整路径。
		fmt.Printf("  - %s\n", file.GetFilePath())
	}

	// 示例2: 基于真实文件系统和 tsconfig.json 创建项目
	// 这是最常用的方式，可以利用 TypeScript 项目的完整配置。
	// 对应 ts-morph 的 `new Project({ tsConfigFilePath: 'path/to/tsconfig.json' })`。
	fmt.Println("\n📁 示例2: 基于文件系统和 tsconfig.json 创建项目")
	realProjectPath := "/Users/bird/Desktop/alalyzer/analyzer-ts/tsmorphgo/examples/demo-react-app"
	realProject := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:         realProjectPath,
		IgnorePatterns:   []string{"node_modules", "dist", ".git", "build"},
		TargetExtensions: []string{".ts", ".tsx"},
		UseTsConfig:      true, // 自动查找并使用 tsconfig.json
	})
	defer realProject.Close()

	realFiles := realProject.GetSourceFiles()
	fmt.Printf("真实React项目包含 %d 个TypeScript文件:\n", len(realFiles))

	// 简单地分析项目中的模块构成
	fileCategories := make(map[string]int)
	for _, file := range realFiles {
		relPath, _ := filepath.Rel(realProjectPath, file.GetFilePath())
		// 根据文件路径将文件分类
		if strings.Contains(relPath, "components/") {
			fileCategories["Components"]++
		} else if strings.Contains(relPath, "utils") || strings.Contains(relPath, "services") {
			fileCategories["Utils/Services"]++
		} else if strings.Contains(relPath, "types") {
			fileCategories["Types"]++
		} else {
			fileCategories["Other"]++
		}
	}
	fmt.Printf("\n项目文件分类:\n")
	for category, count := range fileCategories {
		fmt.Printf("  - %s: %d 个\n", category, count)
	}

	// 示例3: 获取并分析特定的单个文件
	// 对应 ts-morph 的 `project.getSourceFile('path/to/file.ts')`。
	fmt.Println("\n🔍 示例3: 分析特定文件")
	appFile := realProject.GetSourceFile(filepath.Join(realProjectPath, "src/App.tsx"))
	if appFile != nil {
		fmt.Printf("App.tsx 文件信息:\n")
		fmt.Printf("  - 完整路径: %s\n", appFile.GetFilePath())

		// 统计文件中的导入语句数量
		importCount := 0
		// ForEachDescendant 遍历文件中的所有节点。
		appFile.ForEachDescendant(func(node tsmorphgo.Node) {
			// 通过节点的 Kind 属性判断其语法类型。
			if node.Kind == tsmorphgo.KindImportDeclaration {
				importCount++
			}
		})
		fmt.Printf("  - 导入语句数量: %d\n", importCount)
	} else {
		fmt.Println("未找到 App.tsx 文件")
	}

	// 示例4: 在运行时动态地向项目中添加新文件
	// 对应 ts-morph 的 `project.createSourceFile(filePath, content)`。
	fmt.Println("\n➕ 示例4: 动态添加文件到项目")
	dynamicContent := `
		// 这是一个在运行时动态添加的配置文件
		export const DYNAMIC_CONFIG = { version: '1.0.0' };
	`
	newFile, err := realProject.CreateSourceFile(
		filepath.Join(realProjectPath, "src/dynamic-config.ts"),
		dynamicContent,
		tsmorphgo.CreateSourceFileOptions{Overwrite: true}, // 如果文件已存在，则覆盖
	)
	if err != nil {
		log.Printf("创建动态文件失败: %v", err)
	} else {
		fmt.Printf("成功创建动态文件: %s\n", newFile.GetFilePath())
		// 验证文件是否真的被添加到项目中
		finalFileCount := len(realProject.GetSourceFiles())
		fmt.Printf("添加后项目文件总数: %d\n", finalFileCount)
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