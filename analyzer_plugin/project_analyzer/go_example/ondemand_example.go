// Package main 提供了按需调用分析器的示例
// 说明：此文件仅作参考，展示了如何在不同业务方法中按需执行 analyzer
// 根目录下执行：go run -tags=ondemand ./analyzer_plugin/project_analyzer/go_example/ondemand_example.go
//
// 架构说明：
//   NewProjectAnalyzer 构造时会自动解析项目并内部持有 ProjectContext
//   用户可以通过 analyzer.Context() 获取并在 service 中持有
//   两种使用方式：
//     1. 直接使用 RunOneT（无需持有 context）:
//        project_analyzer.RunOneT[...](s.analyzer, ...)
//     2. 持有 context 后传递给其他函数:
//        otherFunc(s.analyzerCtx)

//go:build ondemand
// +build ondemand

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	// 导入 analyzer 包以触发注册
	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
	component_deps_v2_pkg "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/component_deps_v2"
	export_call_pkg "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/export_call"
	list_deps_pkg "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/list_deps"
)

// MyBusinessService 模拟你的业务服务
// 持有分析器和上下文，可以在业务流程中按需调用各个 analyzer
type MyBusinessService struct {
	analyzer     *project_analyzer.ProjectAnalyzer
	analyzerCtx  *project_analyzer.ProjectContext // 可选：持有上下文供其他地方使用
	absPath      string                           // 项目绝对路径
	manifestPath string
	results      map[string]project_analyzer.Result // 收集的分析结果
}

// NewMyBusinessService 创建业务服务
// NewProjectAnalyzer 构造时会自动解析项目（耗时操作）
func NewMyBusinessService(projectPath string) (*MyBusinessService, error) {
	// 转换为绝对路径
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, fmt.Errorf("无法解析项目路径: %w", err)
	}

	// 检查项目路径是否存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("项目路径不存在: %s", absPath)
	}

	// 创建 ProjectAnalyzer（会自动解析项目，这是耗时操作）
	fmt.Println("===========================================")
	fmt.Println("正在创建 ProjectAnalyzer 并解析项目...")
	fmt.Println("===========================================")

	analyzer, err := project_analyzer.NewProjectAnalyzer(project_analyzer.Config{
		ProjectRoot: absPath,
		Exclude:     []string{"node_modules/**", "dist/**", "**/*.test.ts", "**/*.spec.ts"},
		IsMonorepo:  false,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 ProjectAnalyzer 失败: %w", err)
	}

	fmt.Printf("✓ 项目解析完成: %s\n\n", absPath)

	// 获取分析上下文（可选：可以在业务中传递和复用）
	analyzerCtx := analyzer.Context()

	// manifest 路径
	manifestPath := filepath.Join(absPath, ".analyzer/component-manifest.json")

	return &MyBusinessService{
		analyzer:     analyzer,
		analyzerCtx:  analyzerCtx,
		absPath:      absPath,
		manifestPath: manifestPath,
		results:      make(map[string]project_analyzer.Result),
	}, nil
}

// ListDependencies 列出项目依赖
// 调用 list-deps analyzer 获取项目的所有 NPM 依赖
func (s *MyBusinessService) ListDependencies() error {
	fmt.Println("===========================================")
	fmt.Println("列出项目依赖: list-deps")
	fmt.Println("===========================================")

	// 使用 RunOneT 泛型函数，直接返回具体类型，无需类型断言
	result, err := project_analyzer.RunOneT[*list_deps_pkg.ListDepsResult](
		s.analyzer,
		project_analyzer.AnalyzerListDeps,
		project_analyzer.ListDepsConfig{},
	)
	if err != nil {
		return fmt.Errorf("执行 list-deps 失败: %w", err)
	}

	// 存储结果
	s.results[result.Name()] = result

	printListDepsResultOndemand(result)
	return nil
}

// AnalyzeComponentDeps 分析组件依赖
// 调用 component-deps-v2 analyzer 分析组件之间的依赖关系
func (s *MyBusinessService) AnalyzeComponentDeps() error {
	fmt.Println("===========================================")
	fmt.Println("分析组件依赖: component-deps-v2")
	fmt.Println("===========================================")

	// 使用 RunOneT 泛型函数，直接返回具体类型，无需类型断言
	result, err := project_analyzer.RunOneT[*component_deps_v2_pkg.ComponentDepsV2Result](
		s.analyzer,
		project_analyzer.AnalyzerComponentDepsV2,
		project_analyzer.ComponentDepsV2Config{
			Manifest: s.manifestPath,
		},
	)
	if err != nil {
		return fmt.Errorf("执行 component-deps-v2 失败: %w", err)
	}

	// 存储结果
	s.results[result.Name()] = result

	printComponentDepsResultOndemand(result)
	return nil
}

// AnalyzeExportCall 分析导出节点引用
// 调用 export-call analyzer 分析资产目录的导出节点引用关系
func (s *MyBusinessService) AnalyzeExportCall() error {
	fmt.Println("===========================================")
	fmt.Println("分析导出节点引用: export-call")
	fmt.Println("===========================================")

	// 使用 RunOneT 泛型函数，直接返回具体类型，无需类型断言
	result, err := project_analyzer.RunOneT[*export_call_pkg.ExportCallResult](
		s.analyzer,
		project_analyzer.AnalyzerExportCall,
		project_analyzer.ExportCallConfig{
			Manifest: s.manifestPath,
		},
	)
	if err != nil {
		return fmt.Errorf("执行 export-call 失败: %w", err)
	}

	// 存储结果
	s.results[result.Name()] = result

	printExportCallResultOndemand(result)
	return nil
}

// ProcessUserRequest 模拟处理用户请求
// 演示在业务逻辑中按需调用 analyzer
func (s *MyBusinessService) ProcessUserRequest(userAction string) error {
	fmt.Printf("\n===========================================")
	fmt.Printf("处理用户请求: %s\n", userAction)
	fmt.Printf("===========================================\n")

	// 可以在业务逻辑的任意位置调用 analyzer
	return s.ListDependencies()
}

// SaveResults 保存所有分析结果到 JSON 文件
func (s *MyBusinessService) SaveResults() error {
	outputDir := filepath.Join(s.absPath, ".analyzer", "output")
	os.MkdirAll(outputDir, 0755)

	fmt.Println("\n保存结果:")
	fmt.Println("───────────────────────────────────────────")

	for name, result := range s.results {
		jsonFile := filepath.Join(outputDir, fmt.Sprintf("%s.json", name))
		jsonData, err := result.ToJSON(true)
		if err != nil {
			log.Printf("警告: 无法写入 %s: %v", jsonFile, err)
			continue
		}
		if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
			log.Printf("警告: 无法写入 %s: %v", jsonFile, err)
			continue
		}
		fmt.Printf("✓ %s -> %s\n", name, jsonFile)
	}

	return nil
}

// =============================================================================
// 结果打印函数（复制自 main.go）
// =============================================================================

// printListDepsResultOndemand 打印 list-deps 结果
func printListDepsResultOndemand(result *list_deps_pkg.ListDepsResult) {
	fmt.Println("【list-deps】NPM 依赖列表")
	fmt.Println("───────────────────────────────────────────")

	fmt.Printf("摘要: %s\n\n", result.Summary())

	// 显示每个 package.json 的依赖
	for path, pkgData := range result.PackageData {
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

// printComponentDepsResultOndemand 打印 component-deps-v2 结果
func printComponentDepsResultOndemand(result *component_deps_v2_pkg.ComponentDepsV2Result) {
	fmt.Println("【component-deps-v2】组件依赖分析")
	fmt.Println("───────────────────────────────────────────")

	fmt.Printf("摘要: %s\n\n", result.Summary())
	fmt.Printf("组件数量: %d\n\n", result.Meta.ComponentCount)

	// 显示每个组件的外部依赖
	for _, comp := range result.Components {
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

// printExportCallResultOndemand 打印 export-call 结果
func printExportCallResultOndemand(result *export_call_pkg.ExportCallResult) {
	fmt.Println("【export-call】导出节点引用分析")
	fmt.Println("───────────────────────────────────────────")

	fmt.Printf("摘要: %s\n\n", result.Summary())
	fmt.Printf("模块数量: %d\n\n", len(result.ModuleExports))

	// 统计信息
	totalFiles := 0
	totalNodes := 0
	totalUnreferenced := 0

	for _, module := range result.ModuleExports {
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
	for _, module := range result.ModuleExports {
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

// =============================================================================
// 示例主函数
// =============================================================================

// ExampleMain 示例主函数
//
// 运行方式（从 go_example 目录）:
//
//	go run -tags=ondemand ondemand_example.go [项目路径]
//
// 或者直接复制此函数到你的项目中作为 main() 使用
func ExampleMain() {
	// 获取项目路径
	var projectPath string
	if len(os.Args) > 1 {
		projectPath = os.Args[1]
	} else {
		// 默认使用 testdata/test_project
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("无法获取当前工作目录: %v", err)
		}
		testProjectPath := filepath.Join(wd, "testdata/test_project")
		projectPath = testProjectPath
	}

	// 创建业务服务（项目只会被解析一次）
	svc, err := NewMyBusinessService(projectPath)
	if err != nil {
		log.Fatalf("创建服务失败: %v", err)
	}

	// 在业务流程的不同位置按需调用 analyzer
	svc.ListDependencies()

	svc.AnalyzeComponentDeps()

	svc.AnalyzeExportCall()

	svc.ProcessUserRequest("用户提交代码")

	// 保存 JSON 结果
	if err := svc.SaveResults(); err != nil {
		log.Printf("保存结果失败: %v", err)
	}

	fmt.Println("\n===========================================")
	fmt.Println("所有检查完成！")
	fmt.Println("===========================================")
}

// main 主函数（仅在使用 ondemand build tag 时编译）
func main() {
	ExampleMain()
}
