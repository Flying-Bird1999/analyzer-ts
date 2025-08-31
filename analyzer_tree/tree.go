// package analyzer_tree 负责将扁平化的解析节点列表，构建成一个能够反映代码作用域和层级关系的树状结构。
// 这是一个更高层次的分析包，它依赖于 `parser` 包提供的基础解析能力。
package analyzer_tree

import "github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"

// Node 是树中任何一个节点的通用接口。
// 它提供了一种统一的方式来访问节点的子节点和父节点。
// 这是为未来更复杂的树操作（如统一的遍历算法）预留的扩展点。
type Node interface {
	GetChildren() []Node
	GetParent() Node
	SetParent(Node)
	AddChild(Node)
}

// RootNode 代表了一个文件或项目的顶层作用域，是整棵树的根节点。
// 它包含了所有在全局作用域下定义的各种声明和表达式。
type RootNode struct {
	Functions         []*FunctionNode         `json:"functions"`
	Calls             []*CallNode             `json:"calls"`
	Variables         []*VariableNode         `json:"variables"`
	Interfaces        []*InterfaceNode        `json:"interfaces"`
	Types             []*TypeAliasNode        `json:"types"`
	Enums             []*EnumNode             `json:"enums"`
	JsxElements       []*JsxNode              `json:"jsxElements"`
	Imports           []*ImportNode           `json:"imports"`
	Exports           []*ExportNode           `json:"exports"`
	ExportAssignments []*ExportAssignmentNode `json:"exportAssignments"`
}

// FunctionNode 代表一个函数声明节点。
// 它不仅包含了函数自身的声明信息，还通过 Children 字段，递归地包含了所有在其作用域内部定义的节点。
// 它是一个【容器节点】。
type FunctionNode struct {
	Declaration parser.FunctionDeclarationResult `json:"declaration"` // 存储了从 parser 包解析出的原始函数声明信息。
	Children    []Node                           `json:"children"`    // 存储了所有直接定义在此函数作用域内的节点。
	parent      Node                             `json:"-"`           // 指向父节点的指针，用于在构建树时进行回溯。
}

// CallNode 代表一个函数调用节点。
// 在当前的树模型中，函数调用被视为【叶子节点】。
type CallNode struct {
	Call   parser.CallExpression `json:"call"`  // 存储了从 parser 包解析出的原始函数调用信息。
	parent Node                  `json:"-"`     // 指向其所属的父节点（根节点或函数节点）。
}

// VariableNode 代表一个变量声明节点。
// 在当前的树模型中，变量声明被视为【叶子节点】。
type VariableNode struct {
	Declaration parser.VariableDeclaration `json:"declaration"` // 存储了从 parser 包解析出的原始变量声明信息。
	parent      Node                       `json:"-"`           // 指向其所属的父节点（根节点或函数节点）。
}

// ImportNode 代表一个导入声明节点。
// 它是一个【叶子节点】。
type ImportNode struct {
	Declaration parser.ImportDeclarationResult `json:"declaration"`
	parent      Node                           `json:"-"`
}

// ExportNode 代表一个命名导出或重导出声明节点。
// 它是一个【叶子节点】。
type ExportNode struct {
	Declaration parser.ExportDeclarationResult `json:"declaration"`
	parent      Node                           `json:"-"`
}

// ExportAssignmentNode 代表一个 `export default` 声明节点。
// 它是一个【叶子节点】。
type ExportAssignmentNode struct {
	Declaration parser.ExportAssignmentResult `json:"declaration"`
	parent      Node                          `json:"-"`
}

// InterfaceNode 代表一个接口声明节点。
// 它是一个【容器节点】，理论上可以包含子节点（例如内联定义的类型），但目前我们将其作为叶子节点简化处理。
type InterfaceNode struct {
	Declaration parser.InterfaceDeclarationResult `json:"declaration"`
	Children    []Node                            `json:"children"`
	parent      Node                              `json:"-"`
}

// TypeAliasNode 代表一个类型别名（`type`）声明节点。
// 它是一个【叶子节点】。
type TypeAliasNode struct {
	Declaration parser.TypeDeclarationResult `json:"declaration"`
	parent      Node                         `json:"-"`
}

// EnumNode 代表一个枚举声明节点。
// 它是一个【容器节点】，可以包含枚举成员。
type EnumNode struct {
	Declaration parser.EnumDeclarationResult `json:"declaration"`
	Children    []Node                       `json:"children"`
	parent      Node                         `json:"-"`
}

// JsxNode 代表一个 JSX 元素节点。
// 它是一个【容器节点】，因为 JSX 元素可以嵌套其他元素。
type JsxNode struct {
	Declaration parser.JSXElement `json:"declaration"`
	Children    []Node            `json:"children"`
	parent      Node              `json:"-"`
}

// --- 通用节点实现 ---

// baseNode 提供了所有节点共享的 parent 指针。
type baseNode struct {
	parent Node `json:"-"`
}

func (b *baseNode) GetParent() Node {
	return b.parent
}

func (b *baseNode) SetParent(p Node) {
	b.parent = p
}

// leafNode 提供了所有叶子节点的通用实现。
type leafNode struct {
	baseNode
}

func (l *leafNode) GetChildren() []Node {
	return nil
}

func (l *leafNode) AddChild(child Node) {
	// 叶子节点不添加子节点
}

// containerNode 提供了所有容器节点的通用实现。
type containerNode struct {
	baseNode
	Children []Node `json:"children"`
}

func (c *containerNode) GetChildren() []Node {
	return c.Children
}

func (c *containerNode) AddChild(child Node) {
	child.SetParent(c)
	c.Children = append(c.Children, child)
}

// --- 根节点实现 ---

// GetChildren 实现 Node 接口，它收集所有类型的子节点并作为一个统一的切片返回。
func (rn *RootNode) GetChildren() []Node {
	var children []Node
	for _, n := range rn.Functions {
		children = append(children, n)
	}
	for _, n := range rn.Calls {
		children = append(children, n)
	}
	for _, n := range rn.Variables {
		children = append(children, n)
	}
	for _, n := range rn.Interfaces {
		children = append(children, n)
	}
	for _, n := range rn.Types {
		children = append(children, n)
	}
	for _, n := range rn.Enums {
		children = append(children, n)
	}
	for _, n := range rn.JsxElements {
		children = append(children, n)
	}
	for _, n := range rn.Imports {
		children = append(children, n)
	}
	for _, n := range rn.Exports {
		children = append(children, n)
	}
	for _, n := range rn.ExportAssignments {
		children = append(children, n)
	}
	return children
}

// GetParent 实现 Node 接口，根节点没有父节点，总是返回 nil。
func (rn *RootNode) GetParent() Node { return nil }

// SetParent 实现 Node 接口，根节点无法设置父节点，此方法为空操作。
func (rn *RootNode) SetParent(p Node) {}

// AddChild 将子节点添加到 RootNode。它通过类型断言将子节点放入对应的切片中。
func (rn *RootNode) AddChild(child Node) {
	child.SetParent(rn)
	switch v := child.(type) {
	case *FunctionNode:
		rn.Functions = append(rn.Functions, v)
	case *CallNode:
		rn.Calls = append(rn.Calls, v)
	case *VariableNode:
		rn.Variables = append(rn.Variables, v)
	case *InterfaceNode:
		rn.Interfaces = append(rn.Interfaces, v)
	case *TypeAliasNode:
		rn.Types = append(rn.Types, v)
	case *EnumNode:
		rn.Enums = append(rn.Enums, v)
	case *JsxNode:
		rn.JsxElements = append(rn.JsxElements, v)
	case *ImportNode:
		rn.Imports = append(rn.Imports, v)
	case *ExportNode:
		rn.Exports = append(rn.Exports, v)
	case *ExportAssignmentNode:
		rn.ExportAssignments = append(rn.ExportAssignments, v)
	}
}

// --- 容器节点实现 ---

func (fn *FunctionNode) GetChildren() []Node { return fn.Children }
func (fn *FunctionNode) GetParent() Node     { return fn.parent }
func (fn *FunctionNode) SetParent(p Node)    { fn.parent = p }
func (fn *FunctionNode) AddChild(child Node) {
	child.SetParent(fn)
	fn.Children = append(fn.Children, child)
}

func (in *InterfaceNode) GetChildren() []Node { return in.Children }
func (in *InterfaceNode) GetParent() Node     { return in.parent }
func (in *InterfaceNode) SetParent(p Node)    { in.parent = p }
func (in *InterfaceNode) AddChild(child Node) {
	child.SetParent(in)
	in.Children = append(in.Children, child)
}

func (en *EnumNode) GetChildren() []Node { return en.Children }
func (en *EnumNode) GetParent() Node     { return en.parent }
func (en *EnumNode) SetParent(p Node)    { en.parent = p }
func (en *EnumNode) AddChild(child Node) {
	child.SetParent(en)
	en.Children = append(en.Children, child)
}

func (jn *JsxNode) GetChildren() []Node { return jn.Children }
func (jn *JsxNode) GetParent() Node     { return jn.parent }
func (jn *JsxNode) SetParent(p Node)    { jn.parent = p }
func (jn *JsxNode) AddChild(child Node) {
	child.SetParent(jn)
	jn.Children = append(jn.Children, child)
}

// --- 叶子节点实现 ---

func (cn *CallNode) GetChildren() []Node { return nil }
func (cn *CallNode) GetParent() Node     { return cn.parent }
func (cn *CallNode) SetParent(p Node)    { cn.parent = p }
func (cn *CallNode) AddChild(child Node) {}

func (vn *VariableNode) GetChildren() []Node { return nil }
func (vn *VariableNode) GetParent() Node     { return vn.parent }
func (vn *VariableNode) SetParent(p Node)    { vn.parent = p }
func (vn *VariableNode) AddChild(child Node) {}

func (in *ImportNode) GetChildren() []Node { return nil }
func (in *ImportNode) GetParent() Node     { return in.parent }
func (in *ImportNode) SetParent(p Node)    { in.parent = p }
func (in *ImportNode) AddChild(child Node) {}

func (en *ExportNode) GetChildren() []Node { return nil }
func (en *ExportNode) GetParent() Node     { return en.parent }
func (en *ExportNode) SetParent(p Node)    { en.parent = p }
func (en *ExportNode) AddChild(child Node) {}

func (ean *ExportAssignmentNode) GetChildren() []Node { return nil }
func (ean *ExportAssignmentNode) GetParent() Node     { return ean.parent }
func (ean *ExportAssignmentNode) SetParent(p Node)    { ean.parent = p }
func (ean *ExportAssignmentNode) AddChild(child Node) {}

func (tan *TypeAliasNode) GetChildren() []Node { return nil }
func (tan *TypeAliasNode) GetParent() Node     { return tan.parent }
func (tan *TypeAliasNode) SetParent(p Node)    { tan.parent = p }
func (tan *TypeAliasNode) AddChild(child Node) {}

// --- 辅助查找函数 ---

// FindFunctionNode 在给定的节点列表中查找具有特定名称的函数节点。
func FindFunctionNode(nodes []Node, name string) *FunctionNode {
	for _, node := range nodes {
		if fn, ok := node.(*FunctionNode); ok && fn.Declaration.Identifier == name {
			return fn
		}
	}
	return nil
}

// FindCallNode 在给定的节点列表中查找特定的函数调用节点。
func FindCallNode(nodes []Node, callChain []string) *CallNode {
	for _, node := range nodes {
		if cn, ok := node.(*CallNode); ok {
			if len(cn.Call.CallChain) == len(callChain) {
				match := true
				for i := range callChain {
					if cn.Call.CallChain[i] != callChain[i] {
						match = false
						break
					}
				}
				if match {
					return cn
				}
			}
		}
	}
	return nil
}

// FindVariableNode 在给定的节点列表中查找具有特定标识符的变量节点。
func FindVariableNode(nodes []Node, identifier string) *VariableNode {
	for _, node := range nodes {
		if vn, ok := node.(*VariableNode); ok {
			for _, decl := range vn.Declaration.Declarators {
				if decl.Identifier == identifier {
					return vn
				}
			}
		}
	}
	return nil
}
