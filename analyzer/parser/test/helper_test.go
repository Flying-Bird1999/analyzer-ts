package parser_test

import (
	"main/analyzer/utils"
	"os"
	"path/filepath"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
	"github.com/stretchr/testify/assert"
)

// findNodeFunc 定义了一个泛型函数类型，用于查找特定的 AST 节点。
// 它接收一个根 AST 节点，并返回特定类型的节点 T。
type findNodeFunc[T any] func(root *ast.SourceFile) T

// testParserFunc 定义了一个泛型函数类型，用于运行解析器逻辑。
// 它接收找到的类型为 T 的 AST 节点和源代码，并返回类型为 R 的结果。
type testParserFunc[T any, R any] func(node T, code string) R

// marshalFunc 定义了一个泛型函数类型，用于将结果序列化为 JSON。
// 它接收一个类型为 R 的结果，并返回其 JSON 表示形式的字节切片。
type marshalFunc[R any] func(result R) ([]byte, error)

// RunTest 是一个通用的测试运行器，它抽象了通用的解析、节点查找、
// 执行解析器和比较 JSON 输出的逻辑。
func RunTest[T any, R any](t *testing.T, code, expectedJSON string, findNode findNodeFunc[T], testParser testParserFunc[T, R], marshal marshalFunc[R]) {
	// 获取当前工作目录，用于创建虚拟文件路径
	wd, err := os.Getwd()
	assert.NoError(t, err, "获取当前工作目录失败")
	dummyPath := filepath.Join(wd, "test.ts")

	// 解析 TypeScript 代码
	sourceFile := utils.ParseTypeScriptFile(dummyPath, code)
	// 查找目标节点
	node := findNode(sourceFile)

	// 使用 'any' 来检查 nil，因为一个有类型的 nil 接口不等于一个原始的 nil
	var nodeAsAny any = node
	assert.NotNil(t, nodeAsAny, "AST 节点不应为 nil")

	// 执行解析
	result := testParser(node, code)
	// 将结果序列化为 JSON
	resultJSON, err := marshal(result)
	assert.NoError(t, err, "将结果序列化为 JSON 失败")

	// 比较生成的 JSON 是否与预期的 JSON 一致
	assert.JSONEq(t, expectedJSON, string(resultJSON), "生成的 JSON 应与预期的 JSON 匹配")
}