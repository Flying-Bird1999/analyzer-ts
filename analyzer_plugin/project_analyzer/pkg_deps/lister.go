// Package pkg_deps 实现列出项目 NPM 依赖的分析器
package pkg_deps

import (
	"fmt"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

func init() {
	// 注册分析器到工厂
	projectanalyzer.RegisterAnalyzer("pkg-deps", func() projectanalyzer.Analyzer {
		return &PkgDepsAnalyzer{}
	})
}

// PkgDepsAnalyzer NPM 依赖列表分析器
type PkgDepsAnalyzer struct{}

var _ projectanalyzer.Analyzer = (*PkgDepsAnalyzer)(nil)

func (l *PkgDepsAnalyzer) Name() string {
	return "pkg-deps"
}

func (l *PkgDepsAnalyzer) Configure(params map[string]string) error {
	return nil
}

func (l *PkgDepsAnalyzer) Analyze(ctx *projectanalyzer.ProjectContext) (projectanalyzer.Result, error) {
	return &PkgDepsResult{
		PackageData: ctx.ParsingResult.Package_Data,
	}, nil
}

// PkgDepsResult NPM 依赖列表分析结果
type PkgDepsResult struct {
	PackageData map[string]projectParser.PackageJsonFileParserResult `json:"packageData"`
}

var _ projectanalyzer.Result = (*PkgDepsResult)(nil)

func (r *PkgDepsResult) Name() string {
	return "NPM Dependencies List"
}

func (r *PkgDepsResult) Summary() string {
	return fmt.Sprintf("%d 个 package.json", len(r.PackageData))
}

func (r *PkgDepsResult) ToJSON(indent bool) ([]byte, error) {
	return projectanalyzer.ToJSONBytes(r, indent)
}

func (r *PkgDepsResult) ToConsole() string {
	var s string
	for path, pkgData := range r.PackageData {
		s += fmt.Sprintf("%s:\n", path)
		for name, dep := range pkgData.NpmList {
			s += fmt.Sprintf("  %s@%s\n", name, dep.Version)
		}
	}
	return s
}

func (r *PkgDepsResult) AnalyzerName() string {
	return "pkg-deps"
}
