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
export * from './other';
`
	wd, _ := os.Getwd()
	dummyPath := filepath.Join(wd, "test.ts")

	tp, err := NewTreeParserFromSource(dummyPath, code)
	assert.NoError(t, err)
	tp.Traverse()
	tree := tp.Tree

	// 断言根节点的子节点数量是否正确
	assert.Equal(t, 6, len(tree.GetChildren()), "根节点应包含6个声明")

	// 使用辅助函数查找并断言各个节点
	importNode := findNode(t, tree, func(n *ImportNode) bool {
		return n.Declaration.Source == "react"
	})
	assert.Same(t, tree, importNode.GetParent(), "导入节点的父节点应该是根节点")

	variableNode := findNode(t, tree, func(n *VariableNode) bool {
		return n.Declaration.Declarators[0].Identifier == "PI" && n.Declaration.Exported
	})
	assert.Same(t, tree, variableNode.GetParent(), "变量节点的父节点应该是根节点")

	interfaceNode := findNode(t, tree, func(n *InterfaceNode) bool {
		return n.Declaration.Identifier == "User" && n.Declaration.Exported
	})
	assert.Same(t, tree, interfaceNode.GetParent(), "接口节点的父节点应该是根节点")

	enumNode := findNode(t, tree, func(n *EnumNode) bool {
		return n.Declaration.Identifier == "Role" && n.Declaration.Exported
	})
	assert.Same(t, tree, enumNode.GetParent(), "枚举节点的父节点应该是根节点")

	fnNode := findNode(t, tree, func(n *FunctionNode) bool {
		return n.Declaration.Identifier == "App" && n.Declaration.Exported
	})
	assert.Same(t, tree, fnNode.GetParent(), "函数节点的父节点应该是根节点")

	exportNode := findNode(t, tree, func(n *ExportNode) bool {
		return n.Declaration.Source == "./other"
	})
	assert.Same(t, tree, exportNode.GetParent(), "导出节点的父节点应该是根节点")
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
	varNode := findNode(t, parentFunc, func(n *VariableNode) bool {
		return n.Declaration.Declarators[0].Identifier == "parentVar"
	})
	assert.Same(t, parentFunc, varNode.GetParent(), "parentVar 的父节点应该是 parentFunc")

	// 查找并断言 childFunc 的定义
	childFunc := findNode(t, parentFunc, func(n *FunctionNode) bool {
		return n.Declaration.Identifier == "childFunc"
	})
	assert.Same(t, parentFunc, childFunc.GetParent(), "childFunc 的父节点应该是 parentFunc")

	// 断言孙子节点
	assert.Equal(t, 1, len(childFunc.GetChildren()), "childFunc 应有1个子节点")
	callNode := findNode(t, childFunc, func(n *CallNode) bool {
		return n.Call.CallChain[0] == "console"
	})
	assert.Same(t, childFunc, callNode.GetParent(), "console.log 调用的父节点应该是 childFunc")
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

	// MyComponent 应该只包含一个 return 语句
	assert.Equal(t, 1, len(myComponent.GetChildren()), "MyComponent 应该只包含一个 return 语句")
	returnNode := findNode(t, myComponent, func(n *ReturnNode) bool { return true })
	assert.Same(t, myComponent, returnNode.GetParent(), "return 语句的父节点应该是 MyComponent")

	// return 语句应该只包含一个 div
	assert.Equal(t, 1, len(returnNode.GetChildren()), "return 语句应该只包含一个 div")
	div := findNode(t, returnNode, func(n *JsxNode) bool {
		return n.Declaration.ComponentChain[0] == "div"
	})
	assert.Same(t, returnNode, div.GetParent(), "div 的父节点应该是 return 语句")

	// 断言 div 的子节点
	assert.Equal(t, 2, len(div.GetChildren()), "div 应该有两个子节点 (h1, Button)")
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

// TestHookScope 测试对 Hooks（如 useEffect）这类将函数作为参数的调用表达式的解析。
func TestHookScope(t *testing.T) {
	code := `
import { useEffect } from 'react';

function MyComponent() {
    useEffect(() => {
        const subscription = API.subscribe();
        return () => {
            API.unsubscribe(subscription);
        };
    });
}
`
	wd, _ := os.Getwd()
	dummyPath := filepath.Join(wd, "test.tsx")

	tp, err := NewTreeParserFromSource(dummyPath, code)
	assert.NoError(t, err)
	tp.Traverse()
	tree := tp.Tree

	// 1. 找到 MyComponent 函数节点
	myComponentFunc := findNode(t, tree, func(n *FunctionNode) bool {
		return n.Declaration.Identifier == "MyComponent"
	})

	// 2. 在 MyComponent 内部找到 useEffect 调用节点
	useEffectCall := findNode(t, myComponentFunc, func(n *CallNode) bool {
		return n.Call.CallChain[0] == "useEffect"
	})
	assert.Same(t, myComponentFunc, useEffectCall.GetParent(), "useEffect 调用的父节点应该是 MyComponent")

	// 3. useEffect 调用节点现在应该是一个容器，包含一个匿名的函数子节点
	assert.Equal(t, 1, len(useEffectCall.GetChildren()), "useEffect 调用应包含一个函数作为子节点")
	anonFunc := findNode(t, useEffectCall, func(n *FunctionNode) bool {
		return n.Declaration.Identifier == ""
	})
	assert.Same(t, useEffectCall, anonFunc.GetParent(), "匿名函数的父节点应该是 useEffect 调用")

	// 4. 断言匿名函数内部的结构
	assert.Equal(t, 2, len(anonFunc.GetChildren()), "useEffect 的回调函数应有两个子节点")

	// 5. 查找内部的 return 语句
	returnNode := findNode(t, anonFunc, func(n *ReturnNode) bool {
		return n.Expression.Type == "arrowFunction"
	})
	assert.Same(t, anonFunc, returnNode.GetParent(), "return 语句的父节点应该是匿名函数")
}
