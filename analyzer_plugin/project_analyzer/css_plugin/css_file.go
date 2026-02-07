// Package css_plugin 实现列出项目中 CSS 文件的分析器
package css_plugin

import (
	"encoding/json"
	"fmt"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

type CssFile struct{}

var _ projectanalyzer.Analyzer = (*CssFile)(nil)

func (c *CssFile) Name() string {
	return "css-file"
}

func (c *CssFile) Configure(params map[string]string) error {
	return nil
}

func (c *CssFile) Analyze(ctx *projectanalyzer.ProjectContext) (projectanalyzer.Result, error) {
	return &CssFileResult{
		CssData: ctx.ParsingResult.Css_Data,
	}, nil
}

type CssFileResult struct {
	CssData map[string]projectParser.CssFileInfo
}

var _ projectanalyzer.Result = (*CssFileResult)(nil)
var _ json.Marshaler = (*CssFileResult)(nil)

func (r *CssFileResult) Name() string {
	return "CSS Files"
}

func (r *CssFileResult) Summary() string {
	return fmt.Sprintf("%d 个 CSS 文件", len(r.CssData))
}

// MarshalJSON 实现 json.Marshaler 接口，直接输出文件路径 map
func (r *CssFileResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.CssData)
}

func (r *CssFileResult) ToJSON(indent bool) ([]byte, error) {
	if indent {
		return json.MarshalIndent(r.CssData, "", "  ")
	}
	return json.Marshal(r.CssData)
}

func (r *CssFileResult) ToConsole() string {
	var s string
	for path := range r.CssData {
		s += fmt.Sprintf("%s\n", path)
	}
	return s
}
