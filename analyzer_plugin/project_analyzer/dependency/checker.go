// package dependency 实现了检查项目NPM依赖健康状况的核心业务逻辑。
package dependency

import (
	"encoding/json"
	"fmt"
	"io"
	"main/analyzer/projectParser"
	"main/analyzer_plugin/project_analyzer/internal/parser"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Check 是NPM依赖检查功能的主入口函数。
// 它负责协调整个检查流程：解析项目、并发地执行各项检查（隐式、未使用、过期），并最终返回整合后的结果。
// rootPath: 要分析的项目根目录。
// ignore: 需要从分析中排除的文件/目录的 glob 模式列表。
// isMonorepo: 指示项目是否为 monorepo。
func Check(rootPath string, ignore []string, isMonorepo bool) *DependencyCheckResult {
	// 步骤 1: 使用新的 parser 包解析整个项目，获取所有代码文件的AST和所有 package.json 的信息。
	ar, err := parser.ParseProject(rootPath, ignore, isMonorepo)
	if err != nil {
		// 在实际的生产代码中，应该更优雅地处理这个错误，例如记录日志或返回错误。
		// 但为了保持与原函数签名一致（返回 *DependencyCheckResult），这里直接panic或返回空结果。
		// 考虑到这是一个分析工具，遇到解析错误通常意味着无法继续，所以打印错误并返回空结果是合理的。
		fmt.Printf("解析项目失败: %v\n", err)
		// 返回一个空的结果，而不是nil，以避免调用者出现空指针异常。
		return &DependencyCheckResult{} 
	}

	// 步骤 2: 准备数据，提取所有在 package.json 中声明的依赖项，存入一个集合中以便快速查找。
	declaredDependencies := make(map[string]bool)
	for _, pkgData := range ar.Package_Data {
		for _, dep := range pkgData.NpmList {
			declaredDependencies[dep.Name] = true
		}
	}

	// 步骤 3: 并行执行各项检查以提高效率。
	// 本地分析（CPU密集型）和网络请求（IO密集型）可以很好地并行。
	var wg sync.WaitGroup
	var implicitDeps []ImplicitDependency
	var unusedDeps []UnusedDependency
	var outdatedDeps []OutdatedDependency

	wg.Add(2)

	// 任务一: 在本地查找隐式依赖和未使用的依赖。
	go func() {
		defer wg.Done()
		var usedDependencies map[string]bool
		implicitDeps, usedDependencies = findImplicitAndUsedDependencies(ar, declaredDependencies)
		unusedDeps = findUnusedDependencies(ar, usedDependencies)
	}()

	// 任务二: 通过网络请求查找所有过期的依赖。
	go func() {
		defer wg.Done()
		outdatedDeps = findOutdatedDependencies(ar)
	}()

	wg.Wait() // 等待所有检查任务完成。

	// 步骤 4: 整合并返回最终的检查结果。
	return &DependencyCheckResult{
		ImplicitDependencies: implicitDeps,
		UnusedDependencies:   unusedDeps,
		OutdatedDependencies: outdatedDeps,
	}
}

// findImplicitAndUsedDependencies 在一次遍历中同时查找隐式依赖和所有被实际使用过的依赖。
// ar: 完整的项目分析结果。
// declaredDependencies: 在所有 package.json 中声明的依赖集合。
// returns: (找到的隐式依赖列表, 所有被使用过的依赖的集合)
func findImplicitAndUsedDependencies(ar *projectParser.ProjectParserResult, declaredDependencies map[string]bool) ([]ImplicitDependency, map[string]bool) {
	usedDependencies := make(map[string]bool)
	implicitDependencies := []ImplicitDependency{}

	for path, jsData := range ar.Js_Data {
		for _, imp := range jsData.ImportDeclarations {
			if imp.Source.Type == "npm" {
				// 将所有在代码中导入的 npm 包标记为“已使用”。
				usedDependencies[imp.Source.NpmPkg] = true

				// 如果这个包不在声明列表里，并且也不是 Node.js 内置模块，那么它就是隐式依赖。
				if !declaredDependencies[imp.Source.NpmPkg] && !nodeBuiltInModules[imp.Source.NpmPkg] {
					implicitDependencies = append(implicitDependencies, ImplicitDependency{
						Name:     imp.Source.NpmPkg,
						FilePath: path,
						Raw:      imp.Raw,
					})
				}
			}
		}
	}
	return implicitDependencies, usedDependencies
}

// findUnusedDependencies 查找在 package.json 中声明但代码中从未被使用过的依赖。
// ar: 完整的项目分析结果。
// usedDependencies: 所有在代码中实际使用过的依赖的集合。
// returns: 未使用依赖的列表。
func findUnusedDependencies(ar *projectParser.ProjectParserResult, usedDependencies map[string]bool) []UnusedDependency {
	unusedDependencies := []UnusedDependency{}
	processedDependencies := make(map[string]bool) // 用于在 monorepo 中避免重复报告同一个未使用的包。

	for path, pkgData := range ar.Package_Data {
		for _, dep := range pkgData.NpmList {
			// 检查该依赖是否应被忽略 (例如：常见的开发工具、@types 类型定义包)。
			isIgnored := devDependencyIgnoreList[dep.Name] || strings.HasPrefix(dep.Name, "@types/")

			// 如果一个依赖未被使用，之前也未处理过，且不应被忽略，则将其标记为“未使用”。
			if !usedDependencies[dep.Name] && !processedDependencies[dep.Name] && !isIgnored {
				unusedDependencies = append(unusedDependencies, UnusedDependency{
					Name:            dep.Name,
					Version:         dep.Version,
					PackageJsonPath: path,
				})
				processedDependencies[dep.Name] = true
			}
		}
	}
	return unusedDependencies
}

// findOutdatedDependencies 通过并发地查询 NPM registry 来查找所有过期的依赖。
// ar: 完整的项目分析结果。
// returns: 过期依赖的列表。
func findOutdatedDependencies(ar *projectParser.ProjectParserResult) []OutdatedDependency {
	outdatedDependencies := []OutdatedDependency{}
	checkedPackages := make(map[string]bool) // 用于在 monorepo 中避免重复检查同一个包。

	// 使用 channel 从并发的 goroutine 中安全地收集结果。
	resultsChan := make(chan OutdatedDependency)
	var wg sync.WaitGroup
	client := &http.Client{Timeout: 10 * time.Second}

	for path, pkgData := range ar.Package_Data {
		for _, dep := range pkgData.NpmList {
			if !checkedPackages[dep.Name] {
				checkedPackages[dep.Name] = true
				wg.Add(1)

				// 为每个依赖检查启动一个独立的 goroutine。
				go func(dep projectParser.NpmItem, path string) {
					defer wg.Done()
					url := fmt.Sprintf("https://registry.npmjs.org/%s", dep.Name)
					resp, err := client.Get(url)
					if err != nil {
						return // 网络错误，静默失败
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						return // 包不存在或API错误，静默失败
					}

					body, err := io.ReadAll(resp.Body)
					if err != nil {
						return
					}

					var info packageInfo
					if err := json.Unmarshal(body, &info); err != nil {
						return
					}

					latestVersion := info.DistTags.Latest
					// 简单的版本号对比，对于大多数情况有效。
					if latestVersion != "" && dep.Version != latestVersion {
						resultsChan <- OutdatedDependency{
							Name:            dep.Name,
							CurrentVersion:  dep.Version,
							LatestVersion:   latestVersion,
							PackageJsonPath: path,
						}
					}
				}(dep, path)
			}
		}
	}

	// 启动一个 goroutine 来等待所有网络请求完成后关闭 channel。
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// 从 channel 中收集所有 goroutine 的结果。
	for res := range resultsChan {
		outdatedDependencies = append(outdatedDependencies, res)
	}

	return outdatedDependencies
}