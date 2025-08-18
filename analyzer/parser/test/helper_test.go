package parser_test

import (
	"main/analyzer/parser"
	"os"
	"path/filepath"
	"testing"

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

	// 创建一个临时文件来写入测试代码，因为我们的解析器现在直接从文件路径读取
	tempFile, err := os.Create(dummyPath)
	assert.NoError(t, err, "创建临时文件失败")
	defer os.Remove(dummyPath) // 测试结束后清理文件

	_, err = tempFile.WriteString(code)
	assert.NoError(t, err, "写入临时文件失败")
	tempFile.Close()

	// 使用新的 API 创建和运行解析器
	p, err := parser.NewParser(dummyPath)
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
