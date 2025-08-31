package analyzer_tree

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- 测试辅助函数 ---

// findNode 是一个泛型辅助函数，用于在节点的直接子节点中查找特定类型的节点。
// node: 父节点
// predicate: 一个返回布尔值的函数，用于判断子节点是否匹配
func findNode[T Node](t *testing.T, node Node, predicate func(n T) bool) T {
	for _, child := range node.GetChildren() {
		if typedChild, ok := child.(T); ok {
			if predicate(typedChild) {
				return typedChild
			}
		}
	}
	// 如果找不到，返回 T 的零值，并让测试失败
	var zero T
	t.Fatalf("在父节点中未找到匹配的子节点")
	return zero
}

// TestGlobalScopeDeclarations 测试解析器是否能正确处理在全局作用域下的各种声明。
func TestGlobalScopeDeclarations(t *testing.T) {
	code := `
import React from 'react';
export const PI = 3.14;
export interface User {}
export enum Role {}
export default function App() {}
`
	wd, _ := os.Getwd()
	dummyPath := filepath.Join(wd, "test.ts")

	tp, err := NewTreeParserFromSource(dummyPath, code)
	assert.NoError(t, err)
	tp.Traverse()
	tree := tp.Tree

	// 断言根节点的子节点数量是否正确
	assert.Equal(t, 5, len(tree.GetChildren()), "根节点应包含5个声明")

	// 使用辅助函数查找并断言各个节点
	findNode(t, tree, func(n *ImportNode) bool {
		return n.Declaration.Source == "react"
	})
	findNode(t, tree, func(n *VariableNode) bool {
		return n.Declaration.Declarators[0].Identifier == "PI" && n.Declaration.Exported
	})
	findNode(t, tree, func(n *InterfaceNode) bool {
		return n.Declaration.Identifier == "User" && n.Declaration.Exported
	})
	findNode(t, tree, func(n *EnumNode) bool {
		return n.Declaration.Identifier == "Role" && n.Declaration.Exported
	})
	findNode(t, tree, func(n *FunctionNode) bool {
		return n.Declaration.Identifier == "App" && n.Declaration.Exported
	})
}

// TestFunctionScope 测试解析器是否能正确处理函数内部的声明和调用。
func TestFunctionScope(t *testing.T) {
	code := `
function parentFunc() {
    let parentVar = 2;
    childFunc();
    function childFunc() {
        console.log("hello");
    }
}
`
	wd, _ := os.Getwd()
	dummyPath := filepath.Join(wd, "test.ts")

	tp, err := NewTreeParserFromSource(dummyPath, code)
	assert.NoError(t, err)
	tp.Traverse()
	tree := tp.Tree

	// 查找 parentFunc
	parentFunc := findNode(t, tree, func(n *FunctionNode) bool {
		return n.Declaration.Identifier == "parentFunc"
	})

	// 断言 parentFunc 的子节点
	assert.Equal(t, 3, len(parentFunc.GetChildren()), "parentFunc 应有3个子节点")

	// 查找并断言 parentVar
	findNode(t, parentFunc, func(n *VariableNode) bool {
		return n.Declaration.Declarators[0].Identifier == "parentVar"
	})

	// 查找并断言 childFunc 调用
	findNode(t, parentFunc, func(n *CallNode) bool {
		return n.Call.CallChain[0] == "childFunc"
	})

	// 查找并断言 childFunc 的定义
	childFunc := findNode(t, parentFunc, func(n *FunctionNode) bool {
		return n.Declaration.Identifier == "childFunc"
	})

	// 断言孙子节点
	assert.Equal(t, 1, len(childFunc.GetChildren()), "childFunc 应有1个子节点")
	findNode(t, childFunc, func(n *CallNode) bool {
		return n.Call.CallChain[0] == "console"
	})
}

// TestJsxNesting 测试解析器是否能正确处理 JSX 的嵌套结构。
func TestJsxNesting(t *testing.T) {
	code := `
function MyComponent() {
    return (
        <div className="App">
            <h1>Title</h1>
            <Button>Click Me</Button>
        </div>
    );
}
`
	wd, _ := os.Getwd()
	dummyPath := filepath.Join(wd, "test.tsx")

	tp, err := NewTreeParserFromSource(dummyPath, code)
	assert.NoError(t, err)
	tp.Traverse()
	tree := tp.Tree

	// 查找 MyComponent 函数
	myComponent := findNode(t, tree, func(n *FunctionNode) bool {
		return n.Declaration.Identifier == "MyComponent"
	})

	// 查找顶层的 div
	div := findNode(t, myComponent, func(n *JsxNode) bool {
		return n.Declaration.ComponentChain[0] == "div"
	})
	assert.Equal(t, "App", div.Declaration.Attrs[0].Value.Data, "div 的 className 应该是 App")

	// 断言 div 的子节点
	assert.Equal(t, 2, len(div.GetChildren()), "div 应该有两个子节点 (h1, Button)")

	// 查找 h1 和 Button
	findNode(t, div, func(n *JsxNode) bool {
		return n.Declaration.ComponentChain[0] == "h1"
	})
	findNode(t, div, func(n *JsxNode) bool {
		return n.Declaration.ComponentChain[0] == "Button"
	})
}

// TestNoDoubleCounting 验证变量声明中的调用表达式不会被重复计为独立的调用节点。
func TestNoDoubleCounting(t *testing.T) {
	code := `
function App() {
    const [count, setCount] = useState(0);
}
`
	wd, _ := os.Getwd()
	dummyPath := filepath.Join(wd, "test.tsx")

	tp, err := NewTreeParserFromSource(dummyPath, code)
	assert.NoError(t, err)
	tp.Traverse()
	tree := tp.Tree

	appFunc := findNode(t, tree, func(n *FunctionNode) bool {
		return n.Declaration.Identifier == "App"
	})

	// App 函数应该只包含一个 VariableDeclaration 子节点。
	// 它不应该包含一个独立的 CallNode `useState(0)`。
	assert.Equal(t, 1, len(appFunc.GetChildren()), "App 函数应该只包含一个子节点")
	assert.IsType(t, &VariableNode{}, appFunc.GetChildren()[0], "唯一的子节点应该是 VariableNode")
}