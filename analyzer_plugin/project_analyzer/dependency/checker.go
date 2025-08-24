// package dependency 实现了检查项目NPM依赖健康状况的核心业务逻辑。
package dependency

import (
	"encoding/json"
	"fmt"
	"io"
	"main/analyzer/projectParser"
	projectanalyzer "main/analyzer_plugin/project_analyzer"
	"main/analyzer_plugin/project_analyzer/internal/parser"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Checker 是“NPM依赖检查”分析器的实现。
type Checker struct{}

var _ projectanalyzer.Analyzer = (*Checker)(nil)

func (c *Checker) Name() string {
	return "npm-check"
}

func (c *Checker) Configure(params map[string]string) error {
	return nil
}

func (c *Checker) Analyze(ctx *projectanalyzer.ProjectContext) (projectanalyzer.Result, error) {
	ar, err := parser.ParseProject(ctx.ProjectRoot, ctx.Exclude, ctx.IsMonorepo)
	if err != nil {
		return nil, fmt.Errorf("解析项目失败: %w", err)
	}

	declaredDependencies := make(map[string]bool)
	for _, pkgData := range ar.Package_Data {
		for _, dep := range pkgData.NpmList {
			declaredDependencies[dep.Name] = true
		}
	}

	var wg sync.WaitGroup
	var implicitDeps []ImplicitDependency
	var unusedDeps []UnusedDependency
	var outdatedDeps []OutdatedDependency

	wg.Add(2)

	go func() {
		defer wg.Done()
		var usedDependencies map[string]bool
		implicitDeps, usedDependencies = findImplicitAndUsedDependencies(ar, declaredDependencies)
		unusedDeps = findUnusedDependencies(ar, usedDependencies)
	}()

	go func() {
		defer wg.Done()
		outdatedDeps = findOutdatedDependencies(ar)
	}()

	wg.Wait()

	finalResult := &DependencyCheckResult{
		ImplicitDependencies: implicitDeps,
		UnusedDependencies:   unusedDeps,
		OutdatedDependencies: outdatedDeps,
	}

	return finalResult, nil
}

func findImplicitAndUsedDependencies(ar *projectParser.ProjectParserResult, declaredDependencies map[string]bool) ([]ImplicitDependency, map[string]bool) {
	usedDependencies := make(map[string]bool)
	implicitDependencies := []ImplicitDependency{}

	for path, jsData := range ar.Js_Data {
		for _, imp := range jsData.ImportDeclarations {
			if imp.Source.Type == "npm" {
				usedDependencies[imp.Source.NpmPkg] = true
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

func findUnusedDependencies(ar *projectParser.ProjectParserResult, usedDependencies map[string]bool) []UnusedDependency {
	unusedDependencies := []UnusedDependency{}
	processedDependencies := make(map[string]bool)

	for path, pkgData := range ar.Package_Data {
		for _, dep := range pkgData.NpmList {
			isIgnored := devDependencyIgnoreList[dep.Name] || strings.HasPrefix(dep.Name, "@types/")
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

func findOutdatedDependencies(ar *projectParser.ProjectParserResult) []OutdatedDependency {
	outdatedDependencies := []OutdatedDependency{}
	checkedPackages := make(map[string]bool)

	resultsChan := make(chan OutdatedDependency)
	var wg sync.WaitGroup
	client := &http.Client{Timeout: 10 * time.Second}

	for path, pkgData := range ar.Package_Data {
		for _, dep := range pkgData.NpmList {
			if !checkedPackages[dep.Name] {
				checkedPackages[dep.Name] = true
				wg.Add(1)

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

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for res := range resultsChan {
		outdatedDependencies = append(outdatedDependencies, res)
	}

	return outdatedDependencies
}
