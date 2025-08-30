package analyzer_tree

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTreeParser(t *testing.T) {
	code := `
let globalVar = 1;

function parentFunc() {
    let parentVar = 2;
    childFunc();

    function childFunc() {
        let childVar = 3;
        console.log(childVar);
    }
}

parentFunc();
`

	wd, err := os.Getwd()
	assert.NoError(t, err)
	dummyPath := filepath.Join(wd, "test_tree_parser.ts")
	err = os.WriteFile(dummyPath, []byte(code), 0644)
	assert.NoError(t, err)
	defer os.Remove(dummyPath)

	// 1. 使用新的 TreeParser 进行解析和构建
	tp, err := NewTreeParser(dummyPath)
	assert.NoError(t, err)
	tp.Traverse()
	tree := tp.Tree

	// 2. 断言树的结构是否正确
	assert.NotNil(t, tree, "树的根节点不应为 nil")

	// 2.1. 检查根节点的直接子节点
	assert.Equal(t, 1, len(tree.Variables), "根作用域应包含一个变量 (globalVar)")
	assert.Equal(t, "globalVar", tree.Variables[0].Declaration.Declarators[0].Identifier)

	assert.Equal(t, 1, len(tree.Functions), "根作用域应包含一个函数 (parentFunc)")
	assert.Equal(t, "parentFunc", tree.Functions[0].Declaration.Identifier)

	assert.Equal(t, 1, len(tree.Calls), "根作用域应包含一个函数调用 (parentFunc())")
	assert.Equal(t, "parentFunc", tree.Calls[0].Call.CallChain[0])

	// 2.2. 深入检查 parentFunc 的内部结构
	parentFunc := tree.Functions[0]
	// 为了方便断言，我们将子节点分类
	var childVars []*VariableNode
	var childCalls []*CallNode
	var childFuncs []*FunctionNode

	for _, child := range parentFunc.Children {
		switch c := child.(type) {
		case *VariableNode:
			childVars = append(childVars, c)
		case *CallNode:
			childCalls = append(childCalls, c)
		case *FunctionNode:
			childFuncs = append(childFuncs, c)
		}
	}

	assert.Equal(t, 1, len(childVars), "parentFunc 内部应有1个变量声明 (parentVar)")
	assert.Equal(t, "parentVar", childVars[0].Declaration.Declarators[0].Identifier)

	assert.Equal(t, 1, len(childCalls), "parentFunc 内部应有1个函数调用 (childFunc())")
	assert.Equal(t, "childFunc", childCalls[0].Call.CallChain[0])

	assert.Equal(t, 1, len(childFuncs), "parentFunc 内部应有1个函数定义 (childFunc)")
	assert.Equal(t, "childFunc", childFuncs[0].Declaration.Identifier)

	// 2.3. 深入检查 childFunc 的内部结构
	childFunc := childFuncs[0]
	var grandChildVars []*VariableNode
	var grandChildCalls []*CallNode

	for _, child := range childFunc.Children {
		switch c := child.(type) {
		case *VariableNode:
			grandChildVars = append(grandChildVars, c)
		case *CallNode:
			grandChildCalls = append(grandChildCalls, c)
		}
	}

	assert.Equal(t, 1, len(grandChildVars), "childFunc 内部应有1个变量声明 (childVar)")
	assert.Equal(t, "childVar", grandChildVars[0].Declaration.Declarators[0].Identifier)

	assert.Equal(t, 1, len(grandChildCalls), "childFunc 内部应有1个函数调用 (console.log)")
	assert.Equal(t, "console", grandChildCalls[0].Call.CallChain[0])
}
