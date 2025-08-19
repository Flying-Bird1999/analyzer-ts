package project_analyzer

import (
	"encoding/json"
	"fmt"
	"io"
	"main/analyzer/projectParser"
	"net/http"
	"strings"
	"sync"
	"time"
)

// findImplicitAndUsedDependencies 在一次遍历中同时查找隐式依赖和所有被使用的依赖。
// @param ar 项目的完整解析结果
// @param declaredDependencies 一个包含所有在 package.json 中声明的依赖的 map
// @return (隐式依赖列表, 所有被使用依赖的集合)
func findImplicitAndUsedDependencies(ar *projectParser.ProjectParserResult, declaredDependencies map[string]bool) ([]ImplicitDependency, map[string]bool) {
	usedDependencies := make(map[string]bool)
	implicitDependencies := []ImplicitDependency{}

	for path, jsData := range ar.Js_Data {
		for _, imp := range jsData.ImportDeclarations {
			if imp.Source.Type == "npm" {
				// 将找到的 npm 包添加到“已使用”列表中
				usedDependencies[imp.Source.NpmPkg] = true

				// 如果这个包不在声明列表里，并且也不是 Node.js 内置模块，那么它就是隐式依赖
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

// findUnusedDependencies 查找在 package.json 中声明但代码中未被使用的依赖。
// @param ar 项目的完整解析结果
// @param usedDependencies 一个包含所有在代码中实际使用过的依赖的 map
// @return 未使用依赖的列表
func findUnusedDependencies(ar *projectParser.ProjectParserResult, usedDependencies map[string]bool) []UnusedDependency {
	unusedDependencies := []UnusedDependency{}
	processedDependencies := make(map[string]bool) // 用于在 monorepo 中避免重复报告同一个包

	for path, pkgData := range ar.Package_Data {
		for _, dep := range pkgData.NpmList {
			// 检查该依赖是否应被忽略 (例如：开发工具、@types包)
			isIgnored := devDependencyIgnoreList[dep.Name] || strings.HasPrefix(dep.Name, "@types/")

			// 如果一个依赖未被使用，之前也未处理过，且不应被忽略，则将其标记为“未使用”
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

// findOutdatedDependencies 通过查询 NPM registry 来查找过期的依赖。
// @param ar 项目的完整解析结果
// @return 过期依赖的列表
func findOutdatedDependencies(ar *projectParser.ProjectParserResult) []OutdatedDependency {
	outdatedDependencies := []OutdatedDependency{}

	// 使用 map 避免在 monorepo 中重复检查同一个包
	checkedPackages := make(map[string]bool)

	// 使用 channel 从并发的 goroutine 中收集结果
	resultsChan := make(chan OutdatedDependency)
	var wg sync.WaitGroup

	client := &http.Client{Timeout: 10 * time.Second}

	for path, pkgData := range ar.Package_Data {
		for _, dep := range pkgData.NpmList {
			if !checkedPackages[dep.Name] {
				checkedPackages[dep.Name] = true
				wg.Add(1)

				// 为每个依赖检查启动一个 goroutine
				go func(dep projectParser.NpmItem, path string) {
					defer wg.Done()
					url := fmt.Sprintf("https://registry.npmjs.org/%s", dep.Name)
					resp, err := client.Get(url)
					if err != nil {
						return
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						return
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
					// 简单的版本号对比，对于大多数情况有效，但未处理复杂的 semver 范围
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

	// 等待所有 goroutine 完成后关闭 channel
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// 从 channel 收集所有结果
	for res := range resultsChan {
		outdatedDependencies = append(outdatedDependencies, res)
	}

	return outdatedDependencies
}