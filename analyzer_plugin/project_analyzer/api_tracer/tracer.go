// package api_tracer 实现了追踪API调用链路的分析器插件。
//
// 使用示例:
// go run main.go analyze api-tracer -i /path/to/your/project \
//   -p "api-tracer.apiPaths=GET /api/v1/users" \
//   -p "api-tracer.apiPaths=POST /api/v1/orders"
package api_tracer

import (
	"errors"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

// Tracer 是追踪API调用的分析器实现。
type Tracer struct {
	// apiPaths 是一个由待搜索的API路径字符串组成的集合，用于快速查找。
	apiPaths map[string]bool
}

// 确保 Tracer 实现了 projectanalyzer.Analyzer 接口。
var _ projectanalyzer.Analyzer = (*Tracer)(nil)

// Name 返回分析器的唯一名称。
func (t *Tracer) Name() string {
	return "api-tracer"
}

// Configure 根据传入的参数配置分析器。
// "apiPaths" 参数是必需的，它应该是一个由逗号分隔的API路径字符串。
func (t *Tracer) Configure(params map[string]string) error {
	apiPathsStr, ok := params["apiPaths"]
	if !ok || apiPathsStr == "" {
		return errors.New("缺少必需的参数: apiPaths")
	}

	t.apiPaths = make(map[string]bool)
	paths := strings.Split(apiPathsStr, ",")
	for _, path := range paths {
		trimmedPath := strings.TrimSpace(path)
		if trimmedPath != "" {
			t.apiPaths[trimmedPath] = true
		}
	}

	return nil
}

// normalizeString 将字符串中连续的空白字符替换为单个空格，用于健壮的字符串比较。
func normalizeString(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// Analyze 对项目进行扫描，查找对已配置API路径的调用。
func (t *Tracer) Analyze(ctx *projectanalyzer.ProjectContext) (projectanalyzer.Result, error) {
	if len(t.apiPaths) == 0 {
		return nil, errors.New("没有配置任何用于分析的API路径")
	}

	result := &ApiTracerResult{
		Findings: []ApiCallSite{},
	}

	// 为待查找的API路径创建一个标准化的版本，以进行健壮匹配。
	normalizedApiPaths := make(map[string]string)
	for path := range t.apiPaths {
		normalizedApiPaths[normalizeString(path)] = path
	}

	// 遍历所有已解析的JS/TS文件
	for filePath, jsData := range ctx.ParsingResult.Js_Data {
		// 遍历文件中的所有函数调用表达式
		for _, callExpr := range jsData.CallExpressions {
			if t.isFetchCall(callExpr) {
				// 检查第一个参数是否为类字符串字面量
				if len(callExpr.Arguments) > 0 {
					arg := callExpr.Arguments[0]

					var potentialApiPath string
					if arg.Type == "stringLiteral" {
						if path, ok := arg.Data.(string); ok {
							potentialApiPath = path
						}
					} else {
						// 对模板字符串等其他类型进行回退处理。
						// 从原始表达式中剔除引号和首尾空格。
						potentialApiPath = strings.TrimSpace(arg.Expression)
						potentialApiPath = strings.Trim(potentialApiPath, "`'\"")
					}

					if potentialApiPath != "" {
						normalizedPath := normalizeString(potentialApiPath)
						// 检查标准化后的路径是否存在于我们的查找集合中
						if originalPath, exists := normalizedApiPaths[normalizedPath]; exists {
							// 找到了一个匹配项！
							finding := ApiCallSite{
								ApiPath:  originalPath, // 报告原始的、未标准化的路径
								FilePath: filePath,
								Raw:      callExpr.Raw,
							}
							result.Findings = append(result.Findings, finding)
						}
					}
				}
			}
		}
	}

	return result, nil
}

// isFetchCall 检查一个调用表达式是否匹配 XX.fetch() 的模式。
func (t *Tracer) isFetchCall(callExpr parser.CallExpression) bool {
	// CallChain 是一个类似 ["request", "fetch"] 的列表
	chain := callExpr.CallChain
	if len(chain) >= 2 && chain[len(chain)-1] == "fetch" {
		return true
	}
	return false
}
