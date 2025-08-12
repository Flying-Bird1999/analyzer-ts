package parser_test

import (
	"main/analyzer/utils"
	"os"
	"path/filepath"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
	"github.com/stretchr/testify/assert"
)

// findNodeFunc defines a generic function type for finding a specific AST node.
// It takes a root AST node and returns the specific node type T.
type findNodeFunc[T any] func(root *ast.SourceFile) T

// testParserFunc defines a generic function type for running the parser logic.
// It takes the found AST node of type T and the source code, and returns a result of type R.
type testParserFunc[T any, R any] func(node T, code string) R

// marshalFunc defines a generic function type for marshaling the result into JSON.
// It takes a result of type R and returns its JSON representation as a byte slice.
type marshalFunc[R any] func(result R) ([]byte, error)

// runTest is a generic test runner that abstracts the common logic of parsing, node finding,
// executing the parser, and comparing the JSON output.
func RunTest[T any, R any](t *testing.T, code, expectedJSON string, findNode findNodeFunc[T], testParser testParserFunc[T, R], marshal marshalFunc[R]) {
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current working directory")
	dummyPath := filepath.Join(wd, "test.ts")

	sourceFile := utils.ParseTypeScriptFile(dummyPath, code)
	node := findNode(sourceFile)

	// Use 'any' to check for nil because a typed nil interface is not equal to a raw nil.
	var nodeAsAny any = node
	assert.NotNil(t, nodeAsAny, "AST node should not be nil")

	result := testParser(node, code)
	resultJSON, err := marshal(result)
	assert.NoError(t, err, "Failed to marshal result to JSON")

	assert.JSONEq(t, expectedJSON, string(resultJSON), "The generated JSON should match the expected JSON")
}
