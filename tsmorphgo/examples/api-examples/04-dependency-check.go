//go:build example04

package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run 04-dependency-check.go <TypeScript项目路径>")
		os.Exit(1)
	}

	projectPath := os.Args[1]

	fmt.Println("🔗 依赖检查示例 - Import/Export 分析")
	fmt.Println("==================================================")

	// 创建项目
	config := tsmorphgo.ProjectConfig{
		RootPath:         projectPath,
		IgnorePatterns:   []string{"node_modules", "dist", "build", ".git"},
		TargetExtensions: []string{".ts", ".tsx", ".js", ".jsx"},
	}
	project := tsmorphgo.NewProject(config)

	// 分析依赖
	analysis := analyzeDependencies(project)

	// 打印摘要
	fmt.Printf("✅ 分析完成，发现 %d 个依赖关系\n", len(analysis.Dependencies))
	fmt.Printf("📦 涉及 %d 个文件\n", len(analysis.Files))
	fmt.Printf("🔍 第三方依赖: %d 个\n", len(analysis.ThirdPartyDeps))

	// 显示依赖最多的文件
	fmt.Println("\n📊 依赖最多的文件 (前 5 个):")
	for i, file := range analysis.TopFilesByDeps {
		if i >= 5 {
			break
		}
		fmt.Printf("  %d. %s (%d 个依赖)\n", i+1, file.File, file.DepsCount)
	}

	// 显示第三方依赖
	fmt.Println("\n📦 第三方依赖列表:")
	for _, dep := range analysis.ThirdPartyDeps {
		fmt.Printf("  - %s\n", dep)
	}

	// 显示模块分类统计
	fmt.Println("\n🗂️ 依赖分类统计:")
	for category, count := range analysis.Categories {
		fmt.Printf("  %s: %d\n", category, count)
	}

	fmt.Println("\n✅ 依赖检查完成！")
}

// DependencyAnalysis 依赖分析结果
type DependencyAnalysis struct {
	Dependencies     []Dependency `json:"dependencies"`
	Files            []string     `json:"files"`
	ThirdPartyDeps   []string     `json:"thirdPartyDeps"`
	TopFilesByDeps   []FileDeps   `json:"topFilesByDeps"`
	Categories       map[string]int `json:"categories"`
}

// Dependency 依赖关系
type Dependency struct {
	FromFile  string `json:"fromFile"`
	ToFile    string `json:"toFile"`
	Type      string `json:"type"`      // local, third-party, scoped
	ImportType string `json:"importType"` // default, named, namespace
	Line      int    `json:"line"`
}

// FileDeps 文件依赖统计
type FileDeps struct {
	File      string `json:"file"`
	DepsCount int    `json:"depsCount"`
}

// analyzeDependencies 分析项目依赖
func analyzeDependencies(project *tsmorphgo.Project) *DependencyAnalysis {
	analysis := &DependencyAnalysis{
		Dependencies: []Dependency{},
		Files:        []string{},
		Categories:    make(map[string]int),
	}

	// 收集所有文件
	fileMap := make(map[string]int)
	for _, sf := range project.GetSourceFiles() {
		fileMap[sf.GetFilePath()] = 0
		analysis.Files = append(analysis.Files, sf.GetFilePath())
	}

	// 分析每个文件的依赖
	thirdPartySet := make(map[string]bool)
	for _, sf := range project.GetSourceFiles() {
		deps := analyzeFileDependencies(sf)
		analysis.Dependencies = append(analysis.Dependencies, deps...)

		// 统计文件依赖数
		fileMap[sf.GetFilePath()] = len(deps)

		// 收集第三方依赖
		for _, dep := range deps {
			if dep.Type == "third-party" {
				pkg := extractPackageName(dep.ToFile)
				thirdPartySet[pkg] = true
			}
			analysis.Categories[dep.Type]++
		}
	}

	// 收集第三方依赖列表
	for dep := range thirdPartySet {
		analysis.ThirdPartyDeps = append(analysis.ThirdPartyDeps, dep)
	}
	sort.Strings(analysis.ThirdPartyDeps)

	// 找出依赖最多的文件
	for file, count := range fileMap {
		analysis.TopFilesByDeps = append(analysis.TopFilesByDeps, FileDeps{
			File:      file,
			DepsCount: count,
		})
	}

	// 按依赖数排序
	sort.Slice(analysis.TopFilesByDeps, func(i, j int) bool {
		return analysis.TopFilesByDeps[i].DepsCount > analysis.TopFilesByDeps[j].DepsCount
	})

	return analysis
}

// analyzeFileDependencies 分析单个文件的依赖
func analyzeFileDependencies(sf *tsmorphgo.SourceFile) []Dependency {
	var dependencies []Dependency

	sf.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.Kind == ast.KindImportDeclaration {
			dep := analyzeImportDeclaration(node, sf)
			if dep.ToFile != "" {
				dependencies = append(dependencies, dep)
			}
		}
	})

	return dependencies
}

// analyzeImportDeclaration 分析 import 声明
func analyzeImportDeclaration(node tsmorphgo.Node, sf *tsmorphgo.SourceFile) Dependency {
	dep := Dependency{
		FromFile:  sf.GetFilePath(),
		Line:      node.GetStartLineNumber(),
		ImportType: "default", // 默认值
	}

	importDecl, ok := tsmorphgo.AsImportDeclaration(node)
	if !ok || importDecl.Source.FilePath == "" && importDecl.Source.NpmPkg == "" {
		return dep
	}

	// 提取导入路径
	if importDecl.Source.FilePath != "" {
		dep.ToFile = importDecl.Source.FilePath
	} else if importDecl.Source.NpmPkg != "" {
		dep.ToFile = importDecl.Source.NpmPkg
	}

	// 判断导入类型
	if len(importDecl.ImportModules) > 0 {
		// 检查是否有命名导入
		hasNamed := false
		for _, module := range importDecl.ImportModules {
			if module.Type == "named" {
				hasNamed = true
				break
			}
		}
		if hasNamed {
			dep.ImportType = "named"
		} else {
			dep.ImportType = "default"
		}
	}

	// 分类依赖类型
	dep.Type = classifyDependencyType(dep.ToFile)

	return dep
}

// classifyDependencyType 分类依赖类型
func classifyDependencyType(path string) string {
	if strings.HasPrefix(path, ".") {
		return "local"
	}
	if strings.HasPrefix(path, "@") {
		return "scoped"
	}
	return "third-party"
}

// extractPackageName 提取包名
func extractPackageName(importPath string) string {
	// 处理 scoped packages
	if strings.HasPrefix(importPath, "@") {
		re := regexp.MustCompile(`^(@[^/]+/[^/]+)`)
		if match := re.FindStringSubmatch(importPath); match != nil {
			return match[1]
		}
	}

	// 处理普通包
	parts := strings.Split(importPath, "/")
	if len(parts) > 0 {
		return parts[0]
	}

	return importPath
}