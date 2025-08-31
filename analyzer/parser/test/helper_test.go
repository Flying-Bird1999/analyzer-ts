package parser_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"

	"github.com/stretchr/testify/assert"
)

// RunTest 是一个通用的测试运行器，它抽象了通用的解析和比较逻辑。
// 它接收测试代码和预期的 JSON 字符串，然后：
// 1. 创建一个新的解析器实例。
// 2. 对代码进行完整遍历解析。
// 3. 调用一个提取函数（extractFn）从完整的解析结果中提取出我们关心的部分。
// 4. 调用一个序列化函数（marshalFn）将提取出的结果转换为 JSON。
// 5. 将生成的 JSON 与预期的 JSON 进行比较。
func RunTest[R any](t *testing.T, code, expectedJSON string, extractFn func(result *parser.ParserResult) R, marshalFn func(result R) ([]byte, error)) {
	// 获取当前工作目录，用于创建虚拟文件路径
	wd, err := os.Getwd()
	assert.NoError(t, err, "获取当前工作目录失败")
	dummyPath := filepath.Join(wd, "test.ts")

	// 使用源码创建解析器，避免文件 I/O
	p, err := parser.NewParserFromSource(dummyPath, code)
	assert.NoError(t, err, "创建解析器失败")
	p.Traverse()

	// 从完整的解析结果中提取出我们关心的部分
	extractedResult := extractFn(p.Result)

	// 将提取的结果序列化为 JSON
	resultJSON, err := marshalFn(extractedResult)
	assert.NoError(t, err, "将结果序列化为 JSON 失败")

	// 比较生成的 JSON 是否与预期的 JSON 一致
	assert.JSONEq(t, expectedJSON, string(resultJSON), "生成的 JSON 应与预期的 JSON 匹配")
}

// a helper function to trim whitespace from raw fields for robust testing
func trimRaw(s string) string {
	return strings.TrimSpace(s)
}