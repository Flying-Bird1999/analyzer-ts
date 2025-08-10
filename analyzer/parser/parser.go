// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（parser.go）是解析器的核心，定义了主解析结构、遍历逻辑和结果收集。
package parser

import (
	"fmt"
	"main/analyzer/utils"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// ParserResult 是单文件解析的最终结果容器。
// 它存储了从文件中提取出的所有顶层声明和表达式，
// 包括导入、导出、接口、类型、枚举、变量、函数调用和 JSX 元素。
type ParserResult struct {
	filePath string // 被解析文件的路径，仅内部使用。
	// 以下字段存储了文件中提取出的所有声明和表达式。
	ImportDeclarations    []ImportDeclarationResult             // 文件中所有的导入声明
	ExportDeclarations    []ExportDeclarationResult             // 文件中所有的导出声明
	InterfaceDeclarations map[string]InterfaceDeclarationResult // 文件中所有的接口声明，以接口名作为 key
	TypeDeclarations      map[string]TypeDeclarationResult      // 文件中所有的类型别名声明，以类型名作为 key
	EnumDeclarations      map[string]EnumDeclarationResult      // 文件中所有的枚举声明，以枚举名作为 key
	VariableDeclarations  []VariableDeclaration                 // 文件中所有的变量声明
	CallExpressions       []CallExpression                      // 文件中所有的函数/方法调用表达式
	JsxElements           []JSXElement                          // 文件中所有的 JSX 元素
}

// NodePosition 用于精确记录代码在源文件中的位置。
type NodePosition struct {
	Line   int `json:"line"`   // 行号
	Column int `json:"column"` // 列号
}

// SourceLocation 定义了一个节点在源码中的范围。
type SourceLocation struct {
	Start NodePosition `json:"start"` // 节点起始位置
	End   NodePosition `json:"end"`   // 节点结束位置
}

// NewParserResult 创建并初始化一个 ParserResult 实例。
// 它为所有切片和映射进行了初始化，以避免在后续添加数据时出现空指针错误。
func NewParserResult(filePath string) ParserResult {
	return ParserResult{
		filePath:              filePath,
		ImportDeclarations:    []ImportDeclarationResult{},
		InterfaceDeclarations: make(map[string]InterfaceDeclarationResult),
		TypeDeclarations:      make(map[string]TypeDeclarationResult),
		EnumDeclarations:      make(map[string]EnumDeclarationResult),
		VariableDeclarations:  []VariableDeclaration{},
		CallExpressions:       []CallExpression{},
		JsxElements:           []JSXElement{},
	}
}

// AddImportDeclaration 向结果中添加一个解析后的导入声明。
func (pr *ParserResult) AddImportDeclaration(idr *ImportDeclarationResult) {
	pr.ImportDeclarations = append(pr.ImportDeclarations, *idr)
}

// AddExportDeclaration 向结果中添加一个解析后的导出声明。
func (pr *ParserResult) AddExportDeclaration(edr *ExportDeclarationResult) {
	pr.ExportDeclarations = append(pr.ExportDeclarations, *edr)
}

// AddInterfaceDeclaration 向结果中添加一个解析后的接口声明。
func (pr *ParserResult) AddInterfaceDeclaration(inter *InterfaceDeclarationResult) {
	pr.InterfaceDeclarations[inter.Identifier] = *inter
}

// addTypeDeclaration 向结果中添加一个解析后的类型别名声明。
func (pr *ParserResult) addTypeDeclaration(tr *TypeDeclarationResult) {
	pr.TypeDeclarations[tr.Identifier] = *tr
}

// addEnumDeclaration 向结果中添加一个解析后的枚举声明。
func (pr *ParserResult) addEnumDeclaration(er *EnumDeclarationResult) {
	pr.EnumDeclarations[er.Identifier] = *er
}

// addVariableDeclaration 向结果中添加一个解析后的变量声明。
func (pr *ParserResult) addVariableDeclaration(vd *VariableDeclaration) {
	pr.VariableDeclarations = append(pr.VariableDeclarations, *vd)
}

// addCallExpression 向结果中添加一个解析后的函数调用表达式。
func (pr *ParserResult) addCallExpression(ce *CallExpression) {
	pr.CallExpressions = append(pr.CallExpressions, *ce)
}

// addJsxNode 向结果中添加一个解析后的 JSX 元素。
func (pr *ParserResult) addJsxNode(jsxNode *JSXElement) {
	pr.JsxElements = append(pr.JsxElements, *jsxNode)
}

// GetResult 返回一个不包含文件路径的解析结果副本，用于外部使用。
func (pr *ParserResult) GetResult() ParserResult {
	return ParserResult{
		ImportDeclarations:    pr.ImportDeclarations,
		ExportDeclarations:    pr.ExportDeclarations,
		InterfaceDeclarations: pr.InterfaceDeclarations,
		TypeDeclarations:      pr.TypeDeclarations,
		EnumDeclarations:      pr.EnumDeclarations,
		VariableDeclarations:  pr.VariableDeclarations,
		CallExpressions:       pr.CallExpressions,
		JsxElements:           pr.JsxElements,
	}
}

// Traverse 是解析器的核心驱动函数。
//  1. 它首先读取并解析指定路径的 TypeScript 文件，生成一个顶层的 AST 节点（SourceFile）。
//  2. 然后，它定义了一个名为 `walk` 的递归函数，用于深度优先遍历整个 AST。
//  3. 在 `walk` 函数中，通过一个 `switch` 语句来识别感兴趣的节点类型（如导入、接口、JSX 元素等）。
//  4. 当匹配到特定类型的节点时，它会调用相应的 `New...` 和 `analyze...` 函数（在其他文件中定义）
//     来提取该节点的详细信息，并将结果添加到 `ParserResult` 中。
//  5. 对于某些节点（如导入声明），遍历会提前终止（`return`），因为我们通常不关心其内部细节。
//  6. 遍历通过调用 `node.ForEachChild(walk)` 来递归地访问所有子节点，从而实现对整个树的访问。
//  7. 最终，从根节点 `sourceFile` 开始调用 `walk`，启动整个遍历过程。
func (pr *ParserResult) Traverse() {
	sourceCode, err := utils.ReadFileContent(pr.filePath)
	if err != nil {
		fmt.Printf("Failed to read file: %s\n", pr.filePath)
		return
	}

	sourceFile := utils.ParseTypeScriptFile(pr.filePath, sourceCode)

	var walk func(node *ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}

		switch node.Kind {
		// 匹配导入声明，例如: import { a } from 'b'
		case ast.KindImportDeclaration:
			idr := NewImportDeclarationResult()
			idr.analyzeImportDeclaration(node.AsImportDeclaration(), sourceCode)
			pr.AddImportDeclaration(idr)
			// 导入声明通常不需要深入遍历其子节点，因此在此处返回。
			return

		// 匹配接口声明，例如: interface MyInterface { ... }
		case ast.KindInterfaceDeclaration:
			inter := NewInterfaceDeclarationResult(node, sourceCode)
			inter.analyzeInterfaces(node.AsInterfaceDeclaration())
			pr.AddInterfaceDeclaration(inter)

		// 匹配类型别名声明，例如: type MyType = string;
		case ast.KindTypeAliasDeclaration:
			tr := NewTypeDeclarationResult(node, sourceCode)
			tr.analyzeTypeDecl(node.AsTypeAliasDeclaration())
			pr.addTypeDeclaration(tr)

		// 匹配枚举声明，例如: enum MyEnum { ... }
		case ast.KindEnumDeclaration:
			er := NewEnumDeclarationResult(node.AsEnumDeclaration(), sourceCode)
			pr.addEnumDeclaration(er)

		// 匹配变量声明语句，例如: const a = 1; let b = '2';
		case ast.KindVariableStatement:
			vd := NewVariableDeclaration(node.AsVariableStatement(), sourceCode)
			pr.addVariableDeclaration(vd)

		// 匹配函数或方法调用，例如: myFunction(); obj.myMethod();
		case ast.KindCallExpression:
			callExpr := node.AsCallExpression()
			ce := NewCallExpression(callExpr, sourceCode)
			ce.analyzeCallExpression(callExpr, sourceCode)
			pr.addCallExpression(ce)

		// 匹配 JSX 元素，包括自闭合和非自闭合的，例如: <MyComponent /> 或 <div>...</div>
		case ast.KindJsxElement, ast.KindJsxSelfClosingElement:
			jsxNode := NewJSXNode(*node, sourceCode)
			pr.addJsxNode(jsxNode)
		}

		// 使用 ForEachChild 方法以标准方式递归遍历子节点。
		// 返回 false 确保遍历会继续到所有同级和子级的节点。
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false // 继续遍历
		})
	}

	// 从 AST 的根节点开始遍历。
	walk(sourceFile.AsNode())
}
