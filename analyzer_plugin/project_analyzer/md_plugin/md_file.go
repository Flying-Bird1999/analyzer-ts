// Package md_plugin 实现列出项目中 Markdown 文件的分析器
package md_plugin

import (
	"encoding/json"
	"fmt"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

type MdFile struct{}

var _ projectanalyzer.Analyzer = (*MdFile)(nil)

func (m *MdFile) Name() string {
	return "md-file"
}

func (m *MdFile) Configure(params map[string]string) error {
	return nil
}

func (m *MdFile) Analyze(ctx *projectanalyzer.ProjectContext) (projectanalyzer.Result, error) {
	return &MdFileResult{
		MdData: ctx.ParsingResult.Md_Data,
	}, nil
}

type MdFileResult struct {
	MdData map[string]projectParser.MdFileInfo
}

var _ projectanalyzer.Result = (*MdFileResult)(nil)
var _ json.Marshaler = (*MdFileResult)(nil)

func (r *MdFileResult) Name() string {
	return "Markdown Files"
}

func (r *MdFileResult) Summary() string {
	return fmt.Sprintf("%d 个 Markdown 文件", len(r.MdData))
}

// MarshalJSON 实现 json.Marshaler 接口，直接输出文件路径 map
func (r *MdFileResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.MdData)
}

func (r *MdFileResult) ToJSON(indent bool) ([]byte, error) {
	if indent {
		return json.MarshalIndent(r.MdData, "", "  ")
	}
	return json.Marshal(r.MdData)
}

func (r *MdFileResult) ToConsole() string {
	var s string
	for path := range r.MdData {
		s += fmt.Sprintf("%s\n", path)
	}
	return s
}

// AnalyzerName 返回对应的分析器名称
func (r *MdFileResult) AnalyzerName() string {
	return "md-file"
}

// init 在包加载时自动注册分析器
func init() {
	projectanalyzer.RegisterAnalyzer("md-file", func() projectanalyzer.Analyzer {
		return &MdFile{}
	})
}
