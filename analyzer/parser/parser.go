// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（parser.go）是解析器的核心，定义了主解析结构、遍历逻辑和结果收集。
package parser

import (
	"fmt"
	"runtime/debug"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// Parser 定义了解析器的主要结构，包含了源码、AST 和最终的解析结果。
type Parser struct {
	// SourceCode 是当前被解析文件的源码内容。
	SourceCode string
	// Ast 是从源码解析出的 AST 的根节点。
	Ast *ast.Node
	// SourceFile 是从源码解析出的 AST 的根节点对应的 SourceFile。
	SourceFile *ast.SourceFile
	// Result 用于存储和累积解析过程中提取出的所有信息。
	Result *ParserResult
	// processedDynamicImports 用于标记在变量声明中找到的动态导入节点。
	// 这样做是为了防止在后续的 `analyzeCallExpression` 中对同一个 `import()` 调用进行重复处理。
	ProcessedDynamicImports map[*ast.Node]bool
}

// NewParser 创建并返回一个新的 Parser 实例。
// 它负责读取文件内容、生成 AST，并初始化解析器结构。
func NewParser(filePath string) (*Parser, error) {
	sourceCode, err := utils.ReadFileContent(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return NewParserFromSource(filePath, sourceCode)
}

// NewParserFromSource 使用源码字符串创建并返回一个新的 Parser 实例。
// 这个构造函数对于测试非常有用，可以避免文件系统的 I/O 操作。
func NewParserFromSource(filePath string, sourceCode string) (*Parser, error) {
	sourceFile := utils.ParseTypeScriptFile(filePath, sourceCode)
	return &Parser{
		SourceCode:              sourceCode,
		Ast:                     sourceFile.AsNode(),
		SourceFile:              sourceFile, // Populate SourceFile
		Result:                  NewParserResult(filePath),
		ProcessedDynamicImports: make(map[*ast.Node]bool),
	}, nil
}

// Visitor 定义了 AST 遍历期间访问不同类型节点的接口。
// 每个 Visit 方法对应一种我们关心的 AST 节点类型。
// 这种设计模式（访问者模式）将节点处理逻辑从遍历逻辑中解耦，
// 使得添加对新节点类型的支持变得容易，而无需修改核心的遍历代码。
type Visitor interface {
	VisitImportDeclaration(*ast.ImportDeclaration)
	VisitExportDeclaration(*ast.ExportDeclaration)
	VisitExportAssignment(*ast.ExportAssignment)
	VisitInterfaceDeclaration(*ast.InterfaceDeclaration)
	VisitTypeAliasDeclaration(*ast.TypeAliasDeclaration)
	VisitEnumDeclaration(*ast.EnumDeclaration)
	VisitVariableStatement(*ast.VariableStatement)
	VisitCallExpression(*ast.CallExpression)
	VisitJsxElement(*ast.Node) // JsxElement 和 JsxSelfClosingElement 没有独立的类型，使用 Node
	VisitFunctionDeclaration(*ast.FunctionDeclaration)
	VisitAnyKeyword(*ast.Node)
	VisitAsExpression(*ast.AsExpression)
	VisitReturnStatement(*ast.ReturnStatement)
}

// Traverse 是解析器的核心驱动函数。
// 它通过启动一个递归的 `walk` 函数来深度优先遍历整个 AST。
// 在每个节点上，它会调用 `dispatch` 方法，该方法会根据节点类型调用合适的 Visitor 方法。
func (p *Parser) Traverse() {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("recovered from panic: %v\n%s", r, debug.Stack())
			p.Result.Errors = append(p.Result.Errors, err)
		}
	}()

	var walk func(node *ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}

		// dispatch 会调用此节点对应的 Visit 方法。
		// continueWalk 控制是否需要继续遍历该节点的子节点。
		continueWalk := p.dispatch(node)
		if !continueWalk {
			return
		}

		// 递归地访问所有子节点。
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false // 返回 false 以确保遍历继续。
		})
	}

	// 从 AST 的根节点开始遍历。
	walk(p.Ast)
}

// dispatch 是节点分发器，它取代了旧的 switch 语句。
// 它检查节点类型，并调用 Parser 上实现的相应 Visitor 方法。
// 返回一个布尔值，指示是否应该继续遍历当前节点的子节点。
func (p *Parser) dispatch(node *ast.Node) (continueWalk bool) {
	// 默认继续遍历子节点
	continueWalk = true

	switch node.Kind {
	case ast.KindImportDeclaration:
		p.VisitImportDeclaration(node.AsImportDeclaration())
		continueWalk = false // 导入声明不需深入遍历其子节点。
	case ast.KindExportDeclaration:
		p.VisitExportDeclaration(node.AsExportDeclaration())
		continueWalk = false // 导出声明同样不需深入遍历。
	case ast.KindExportAssignment:
		p.VisitExportAssignment(node.AsExportAssignment())
		continueWalk = false // `export default` 也不需深入遍历。
	case ast.KindInterfaceDeclaration:
		p.VisitInterfaceDeclaration(node.AsInterfaceDeclaration())
	case ast.KindTypeAliasDeclaration:
		p.VisitTypeAliasDeclaration(node.AsTypeAliasDeclaration())
	case ast.KindEnumDeclaration:
		p.VisitEnumDeclaration(node.AsEnumDeclaration())
	case ast.KindVariableStatement:
		p.VisitVariableStatement(node.AsVariableStatement())
	case ast.KindCallExpression:
		p.VisitCallExpression(node.AsCallExpression())
	case ast.KindJsxElement, ast.KindJsxSelfClosingElement:
		p.VisitJsxElement(node)
	case ast.KindFunctionDeclaration:
		p.VisitFunctionDeclaration(node.AsFunctionDeclaration())
	case ast.KindAnyKeyword:
		p.VisitAnyKeyword(node)
	case ast.KindAsExpression:
		p.VisitAsExpression(node.AsAsExpression())
	case ast.KindReturnStatement:
		p.VisitReturnStatement(node.AsReturnStatement())
	}

	return continueWalk
}

// VisitReturnStatement 解析 return 语句。
func (p *Parser) VisitReturnStatement(node *ast.ReturnStatement) {
	result := AnalyzeReturnStatement(node, p.SourceCode)
	if result != nil {
		p.Result.ReturnStatements = append(p.Result.ReturnStatements, *result)
	}
}

// addError 是一个辅助函数，用于向结果中添加一个格式化的解析错误。
func (p *Parser) addError(node *ast.Node, format string, args ...interface{}) {
	line, col := utils.GetLineAndCharacterOfPosition(p.SourceCode, node.Pos())
	msg := fmt.Sprintf(format, args...)
	err := fmt.Errorf("Error at %s:%d:%d: %s", p.Result.filePath, line+1, col+1, msg)
	p.Result.Errors = append(p.Result.Errors, err)
}

// NewSourceLocation 是一个辅助函数，用于从 AST 节点中创建并返回一个准确的 SourceLocation。
// 它将节点的字符偏移位置转换为行列号。
func NewSourceLocation(node *ast.Node, sourceCode string) *SourceLocation {
	startPos, endPos := node.Pos(), node.End()
	startLine, startChar := utils.GetLineAndCharacterOfPosition(sourceCode, startPos)
	endLine, endChar := utils.GetLineAndCharacterOfPosition(sourceCode, endPos)

	return &SourceLocation{
		Start: NodePosition{Line: startLine + 1, Column: startChar + 1},
		End:   NodePosition{Line: endLine + 1, Column: endChar + 1},
	}
}

// ParserResult 是单文件解析的最终结果容器。
// 它存储了从文件中提取出的所有顶层声明和表达式。
type ParserResult struct {
	filePath              string // 被解析文件的路径，仅内部使用。
	ImportDeclarations    []ImportDeclarationResult
	ExportDeclarations    []ExportDeclarationResult
	ExportAssignments     []ExportAssignmentResult
	InterfaceDeclarations map[string]InterfaceDeclarationResult
	TypeDeclarations      map[string]TypeDeclarationResult
	EnumDeclarations      map[string]EnumDeclarationResult
	VariableDeclarations  []VariableDeclaration
	CallExpressions       []CallExpression
	JsxElements           []JSXElement
	FunctionDeclarations  []FunctionDeclarationResult
	ReturnStatements      []ReturnStatementResult // 新增：用于存储 return 语句
	ExtractedNodes        ExtractedNodes
	Errors                []error
}

// NodePosition 用于精确记录代码在源文件中的位置。
type NodePosition struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// SourceLocation 定义了一个节点在源码中的范围。
type SourceLocation struct {
	Start NodePosition `json:"start"`
	End   NodePosition `json:"end"`
}

// NewParserResult 创建并初始化一个 ParserResult 实例。
func NewParserResult(filePath string) *ParserResult {
	return &ParserResult{
		filePath:              filePath,
		ImportDeclarations:    []ImportDeclarationResult{},
		ExportDeclarations:    []ExportDeclarationResult{},
		ExportAssignments:     []ExportAssignmentResult{},
		InterfaceDeclarations: make(map[string]InterfaceDeclarationResult),
		TypeDeclarations:      make(map[string]TypeDeclarationResult),
		EnumDeclarations:      make(map[string]EnumDeclarationResult),
		VariableDeclarations:  []VariableDeclaration{},
		CallExpressions:       []CallExpression{},
		JsxElements:           []JSXElement{},
		FunctionDeclarations:  []FunctionDeclarationResult{},
		ReturnStatements:      []ReturnStatementResult{},
		ExtractedNodes: ExtractedNodes{
			AnyDeclarations: []AnyInfo{},
			AsExpressions:   []AsExpression{},
		},
		Errors: []error{},
	}
}

// GetResult 返回一个不包含文件路径的解析结果副本，用于外部使用。
func (pr *ParserResult) GetResult() ParserResult {
	return ParserResult{
		ImportDeclarations:    pr.ImportDeclarations,
		ExportDeclarations:    pr.ExportDeclarations,
		ExportAssignments:     pr.ExportAssignments,
		InterfaceDeclarations: pr.InterfaceDeclarations,
		TypeDeclarations:      pr.TypeDeclarations,
		EnumDeclarations:      pr.EnumDeclarations,
		VariableDeclarations:  pr.VariableDeclarations,
		CallExpressions:       pr.CallExpressions,
		JsxElements:           pr.JsxElements,
		FunctionDeclarations:  pr.FunctionDeclarations,
		ReturnStatements:      pr.ReturnStatements,
		ExtractedNodes: ExtractedNodes{
			AnyDeclarations: pr.ExtractedNodes.AnyDeclarations,
			AsExpressions:   pr.ExtractedNodes.AsExpressions,
		},
	}
}

// Traverse 是旧的入口点，现在它将工作委托给新的 Parser 结构。
// 这样做是为了保持对外的 API 兼容性。
func (pr *ParserResult) Traverse() error {
	p, err := NewParser(pr.filePath)
	if err != nil {
		return fmt.Errorf("Error creating parser: %w", err)
	}
	p.Traverse()
	*pr = *p.Result
	return nil
}