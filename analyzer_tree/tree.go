// package analyzer_tree 负责将扁平化的解析节点列表，构建成一个能够反映代码作用域和层级关系的树状结构。
// 这是一个更高层次的分析包，它依赖于 `parser` 包提供的基础解析能力。
package analyzer_tree

import "github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"

// Node 是树中任何一个节点的通用接口。
// 它提供了一种统一的方式来访问节点的子节点和父节点，尽管在当前实现中我们未使用它。
// 这是一个为未来更复杂的树操作（如统一的遍历算法）预留的扩展点。
type Node interface {
	GetChildren() []Node
	GetParent() Node
	SetParent(Node)
	AddChild(Node)
}

// RootNode 代表了一个文件或项目的顶层作用域，是整棵树的根节点。
// 它包含了所有在全局作用域下定义的函数、变量和执行的函数调用。
type RootNode struct {
	Functions []*FunctionNode `json:"functions"`
	Calls     []*CallNode     `json:"calls"`
	Variables []*VariableNode `json:"variables"`
}

// FunctionNode 代表一个函数声明节点。
// 它不仅包含了函数自身的声明信息，还通过 Children 字段，递归地包含了所有在其作用域内部定义的节点。
type FunctionNode struct {
	// Declaration 存储了从 parser 包解析出的原始函数声明信息。
	Declaration parser.FunctionDeclarationResult `json:"declaration"`

	// Children 列表体现了父子关系，存储了所有直接定义在此函数作用域内的节点。
	Children []Node `json:"children"`

	// parent 是一个指向父节点的指针（根节点或外层函数节点），用于在构建树时进行回溯。
	// 它在最终的 JSON 输出中被忽略，以防止循环引用。
	parent Node `json:"-"`
}

// CallNode 代表一个函数调用节点。
// 在当前的树模型中，函数调用被视为叶子节点，不包含子节点。
type CallNode struct {
	// Call 存储了从 parser 包解析出的原始函数调用信息。
	Call parser.CallExpression `json:"call"`

	// parent 指向其所属的父节点（根节点或函数节点）。
	parent Node `json:"-"`
}

// VariableNode 代表一个变量声明节点。
// 在当前的树模型中，变量声明也被视为叶子节点。
type VariableNode struct {
	// Declaration 存储了从 parser 包解析出的原始变量声明信息。
	Declaration parser.VariableDeclaration `json:"declaration"`

	// parent 指向其所属的父节点（根节点或函数节点）。
	parent Node `json:"-"`
}

// --- 接口实现 ---

// GetChildren 实现 Node 接口，它收集所有类型的子节点并作为一个统一的切片返回。
func (rn *RootNode) GetChildren() []Node {
	children := make([]Node, 0, len(rn.Functions)+len(rn.Calls)+len(rn.Variables))
	for _, fn := range rn.Functions {
		children = append(children, fn)
	}
	for _, call := range rn.Calls {
		children = append(children, call)
	}
	for _, v := range rn.Variables {
		children = append(children, v)
	}
	return children
}

// GetParent 实现 Node 接口，根节点没有父节点，总是返回 nil。
func (rn *RootNode) GetParent() Node {
	return nil
}

// SetParent 实现 Node 接口，根节点无法设置父节点，此方法为空操作。
func (rn *RootNode) SetParent(p Node) {
	// 根节点没有父节点，此方法为空操作
}

// AddChild 将子节点添加到 RootNode。
func (rn *RootNode) AddChild(child Node) {
	child.SetParent(rn)
	switch v := child.(type) {
	case *FunctionNode:
		rn.Functions = append(rn.Functions, v)
	case *CallNode:
		rn.Calls = append(rn.Calls, v)
	case *VariableNode:
		rn.Variables = append(rn.Variables, v)
	}
}

// GetChildren 实现 Node 接口，返回函数节点的子节点列表。
func (fn *FunctionNode) GetChildren() []Node {
	return fn.Children
}

// AddChild 将子节点添加到 FunctionNode。
func (fn *FunctionNode) AddChild(child Node) {
	child.SetParent(fn)
	fn.Children = append(fn.Children, child)
}

// GetParent 实现 Node 接口，返回节点的父节点。
func (fn *FunctionNode) GetParent() Node {
	return fn.parent
}

// SetParent 实现 Node 接口，设置节点的父节点。
func (fn *FunctionNode) SetParent(p Node) {
	fn.parent = p
}

// GetChildren 对于叶子节点，总是返回 nil。
func (cn *CallNode) GetChildren() []Node {
	return nil
}

// GetParent 实现 Node 接口，返回节点的父节点。
func (cn *CallNode) GetParent() Node {
	return cn.parent
}

// SetParent 实现 Node 接口，设置节点的父节点。
func (cn *CallNode) SetParent(p Node) {
	cn.parent = p
}

// AddChild 对于叶子节点，此方法为空操作。
func (cn *CallNode) AddChild(child Node) {
	// CallNode 是叶子节点，不添加子节点
}

// GetChildren 对于叶子节点，总是返回 nil。
func (vn *VariableNode) GetChildren() []Node {
	return nil
}

// GetParent 实现 Node 接口，返回节点的父节点。
func (vn *VariableNode) GetParent() Node {
	return vn.parent
}

// SetParent 实现 Node 接口，设置节点的父节点。
func (vn *VariableNode) SetParent(p Node) {
	vn.parent = p
}

// AddChild 对于叶子节点，此方法为空操作。
func (vn *VariableNode) AddChild(child Node) {
	// VariableNode 是叶子节点，不添加子节点
}