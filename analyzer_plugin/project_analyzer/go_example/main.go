// Package main 提供了使用 Runner 调用分析器的示例
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
	component_deps_v2 "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/component_deps_v2"
	export_call "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/export_call"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/list_deps"
)

func main() {
	// 获取项目路径（从命令行参数或使用默认的测试项目）
	projectPath := "/Users/bird/Desktop/alalyzer/analyzer-ts/testdata/test_project"
	if len(os.Args) > 1 {
		projectPath = os.Args[1]
	}

	// 检查项目路径是否存在
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		log.Fatalf("项目路径不存在: %s\n"+
			"用法: go run main.go [/path/to/typescript-project]\n"+
			"默认: 使用测试项目 ../../testdata/test_project", projectPath)
	}

	// 转换为绝对路径
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		log.Fatalf("无法解析项目路径: %v", err)
	}

	fmt.Printf("===========================================\n")
	fmt.Printf("项目分析示例\n")
	fmt.Printf("项目路径: %s\n", absPath)
	fmt.Printf("===========================================\n\n")

	// 1. 创建 Runner
	runner, err := project_analyzer.NewRunner(project_analyzer.RunnerConfig{
		ProjectRoot: absPath,
		Exclude:     []string{"node_modules/**", "dist/**", "**/*.test.ts", "**/*.spec.ts"},
		IsMonorepo:  false,
	})
	if err != nil {
		log.Fatalf("创建 Runner 失败: %v", err)
	}

	// 2. 注册分析器
	fmt.Println("注册分析器...")
	runner.RegisterBatch(
		&list_deps.Lister{},
		&component_deps_v2.ComponentDepsV2Analyzer{},
		&export_call.ExportCallAnalyzer{},
	)
	fmt.Println("✓ 已注册: list-deps, component-deps-v2, export-call\n")

	// 3. 准备配置
	manifestPath := filepath.Join(absPath, ".analyzer/component-manifest.json")
	configs := map[string]map[string]string{
		"list-deps": {}, // list_deps 不需要配置
		"component-deps-v2": {
			"manifest": manifestPath,
		},
		"export-call": {
			"manifest": manifestPath,
		},
	}

	// 4. 执行分析（批量）
	fmt.Println("开始分析...")
	fmt.Println("───────────────────────────────────────────")

	results, err := runner.RunBatch(configs)
	if err != nil {
		log.Fatalf("分析失败: %v", err)
	}

	// 5. 处理结果
	fmt.Println("\n分析结果:")
	fmt.Println("===========================================\n")

	// 5.1 list-deps 结果
	if result, ok := results["list-deps"]; ok {
		printListDepsResult(result)
	}

	// 5.2 component-deps-v2 结果
	if result, ok := results["component-deps-v2"]; ok {
		printComponentDepsResult(result)
	}

	// 5.3 export-call 结果
	if result, ok := results["export-call"]; ok {
		printExportCallResult(result)
	}

	// 6. 保存 JSON 结果
	outputDir := filepath.Join(absPath, ".analyzer", "output")
	os.MkdirAll(outputDir, 0755)
	fmt.Println("\n保存结果:")
	fmt.Println("───────────────────────────────────────────")

	for name, result := range results {
		jsonFile := filepath.Join(outputDir, fmt.Sprintf("%s.json", name))
		jsonData, _ := result.ToJSON(true)
		if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
			log.Printf("警告: 无法写入 %s: %v", jsonFile, err)
		} else {
			fmt.Printf("✓ %s -> %s\n", name, jsonFile)
		}
	}

	fmt.Println("\n===========================================")
	fmt.Println("分析完成！")
	fmt.Println("===========================================")
}

// printListDepsResult 打印 list-deps 结果
func printListDepsResult(result project_analyzer.Result) {
	fmt.Println("【list-deps】NPM 依赖列表")
	fmt.Println("───────────────────────────────────────────")

	listResult, ok := result.(*list_deps.ListDepsResult)
	if !ok {
		fmt.Printf("错误: 结果类型不匹配\n\n")
		return
	}

	fmt.Printf("摘要: %s\n\n", result.Summary())

	// 显示每个 package.json 的依赖
	for path, pkgData := range listResult.PackageData {
		relPath, _ := filepath.Rel(filepath.Dir(path), path)
		fmt.Printf("📦 %s (%d 个依赖)\n", relPath, len(pkgData.NpmList))

		// 只显示前 5 个依赖
		count := 0
		for name, dep := range pkgData.NpmList {
			if count >= 5 {
				remaining := len(pkgData.NpmList) - 5
				if remaining > 0 {
					fmt.Printf("  ... 还有 %d 个依赖\n", remaining)
				}
				break
			}
			fmt.Printf("  - %s@%s\n", name, dep.Version)
			count++
		}
		fmt.Println()
	}
}

// printComponentDepsResult 打印 component-deps-v2 结果
func printComponentDepsResult(result project_analyzer.Result) {
	fmt.Println("【component-deps-v2】组件依赖分析")
	fmt.Println("───────────────────────────────────────────")

	compResult, ok := result.(*component_deps_v2.ComponentDepsV2Result)
	if !ok {
		fmt.Printf("错误: 结果类型不匹配\n\n")
		return
	}

	fmt.Printf("摘要: %s\n\n", result.Summary())
	fmt.Printf("组件数量: %d\n\n", compResult.Meta.ComponentCount)

	// 显示每个组件的外部依赖
	for _, comp := range compResult.Components {
		if len(comp.Dependencies) == 0 {
			continue
		}
		fmt.Printf("📦 %s\n", comp.Name)
		fmt.Printf("   路径: %s\n", comp.Path)

		// 按包名分组依赖
		pkgDeps := make(map[string]int)
		for _, dep := range comp.Dependencies {
			// 从 Source 中获取包名
			if dep.Source.Type == "npm" {
				pkgDeps[dep.Source.NpmPkg]++
			}
		}

		// 显示依赖包
		for pkgName, count := range pkgDeps {
			fmt.Printf("   - %s (%d 个导入)\n", pkgName, count)
		}
		fmt.Println()
	}
}

// printExportCallResult 打印 export-call 结果
func printExportCallResult(result project_analyzer.Result) {
	fmt.Println("【export-call】导出节点引用分析")
	fmt.Println("───────────────────────────────────────────")

	exportResult, ok := result.(*export_call.ExportCallResult)
	if !ok {
		fmt.Printf("错误: 结果类型不匹配\n\n")
		return
	}

	fmt.Printf("摘要: %s\n\n", result.Summary())
	fmt.Printf("模块数量: %d\n\n", len(exportResult.ModuleExports))

	// 统计信息
	totalFiles := 0
	totalNodes := 0
	totalUnreferenced := 0

	for _, module := range exportResult.ModuleExports {
		totalFiles += len(module.Files)
		for _, file := range module.Files {
			totalNodes += len(file.Nodes)
			for _, node := range file.Nodes {
				if len(node.RefFiles) == 0 {
					totalUnreferenced++
				}
			}
		}
	}

	fmt.Printf("统计:\n")
	fmt.Printf("  总文件数: %d\n", totalFiles)
	fmt.Printf("  总节点数: %d\n", totalNodes)
	fmt.Printf("  未引用: %d\n\n", totalUnreferenced)

	// 显示每个模块的导出节点
	for _, module := range exportResult.ModuleExports {
		fmt.Printf("📦 模块: %s (路径: %s)\n", module.ModuleName, module.Path)
		fmt.Printf("   文件数: %d\n", len(module.Files))

		for _, file := range module.Files {
			relFile, _ := filepath.Rel(module.Path, file.File)
			fmt.Printf("\n   📄 %s\n", relFile)

			unreferencedCount := 0
			for _, node := range file.Nodes {
				if len(node.RefFiles) == 0 {
					unreferencedCount++
				}
			}

			if unreferencedCount > 0 {
				fmt.Printf("      ⚠️  未引用导出: %d 个\n", unreferencedCount)
			}

			// 只显示前 3 个节点
			count := 0
			for _, node := range file.Nodes {
				if count >= 3 {
					remaining := len(file.Nodes) - 3
					if remaining > 0 {
						fmt.Printf("      ... 还有 %d 个节点\n", remaining)
					}
					break
				}
				refStatus := "✓"
				if len(node.RefFiles) == 0 {
					refStatus = "✗"
				}
				fmt.Printf("      %s [%s] %s - %s\n", refStatus, node.NodeType, node.ExportType, node.Name)
				count++
			}
		}
		fmt.Println()
	}
}

// saveJSONResult 保存 JSON 结果（可选）
func saveJSONResult(result project_analyzer.Result, filename string) error {
	jsonData, err := result.ToJSON(true)
	if err != nil {
		return err
	}

	// 格式化输出
	var prettyJSON map[string]interface{}
	if err := json.Unmarshal(jsonData, &prettyJSON); err == nil {
		jsonData, _ = json.MarshalIndent(prettyJSON, "", "  ")
	}

	return os.WriteFile(filename, jsonData, 0644)
}
