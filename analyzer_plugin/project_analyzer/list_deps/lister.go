// Package list_deps 实现列出项目 NPM 依赖的分析器
package list_deps

import (
	"fmt"

	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
)

// Lister 依赖列表分析器
type Lister struct{}

var _ projectanalyzer.Analyzer = (*Lister)(nil)

func (l *Lister) Name() string {
	return "list-deps"
}

func (l *Lister) Configure(params map[string]string) error {
	return nil
}

func (l *Lister) Analyze(ctx *projectanalyzer.ProjectContext) (projectanalyzer.Result, error) {
	return &ListDepsResult{
		PackageData: ctx.ParsingResult.Package_Data,
	}, nil
}

// ListDepsResult 依赖列表分析结果
type ListDepsResult struct {
	PackageData map[string]projectParser.PackageJsonFileParserResult `json:"packageData"`
}

var _ projectanalyzer.Result = (*ListDepsResult)(nil)

func (r *ListDepsResult) Name() string {
	return "NPM Dependencies List"
}

func (r *ListDepsResult) Summary() string {
	return fmt.Sprintf("%d 个 package.json", len(r.PackageData))
}

func (r *ListDepsResult) ToJSON(indent bool) ([]byte, error) {
	return projectanalyzer.ToJSONBytes(r, indent)
}

func (r *ListDepsResult) ToConsole() string {
	var s string
	for path, pkgData := range r.PackageData {
		s += fmt.Sprintf("%s:\n", path)
		for name, dep := range pkgData.NpmList {
			s += fmt.Sprintf("  %s@%s\n", name, dep.Version)
		}
	}
	return s
}
