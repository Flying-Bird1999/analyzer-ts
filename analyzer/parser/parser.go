// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（parser.go）是解析器的核心，定义了主解析结构、遍历逻辑和结果收集。
package parser

import (
	"fmt"
	"main/analyzer/utils"
	"strings"

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
	processedDynamicImports map[*ast.Node]bool
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
		processedDynamicImports: make(map[*ast.Node]bool),
	}, nil
}

// Traverse 是解析器的核心驱动函数。
// 它通过启动一个递归的 `walk` 函数来深度优先遍历整个 AST，从而识别和解析各种类型的节点。
func (p *Parser) Traverse() {
	var walk func(node *ast.Node)
	// walk 是一个递归函数，用于遍历 AST 树。
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}

		// 提取 any 信息
		if node.Kind == ast.KindAnyKeyword {
			p.Result.AnyDeclarations = append(p.Result.AnyDeclarations, AnyInfo{
				SourceLocation: SourceLocation{
					Start: func() NodePosition {
						line, character := utils.GetLineAndCharacterOfPosition(p.SourceCode, node.Loc.Pos())
						return NodePosition{Line: line + 1, Column: character + 1}
					}(),
					End: func() NodePosition {
						line, character := utils.GetLineAndCharacterOfPosition(p.SourceCode, node.Loc.End())
						return NodePosition{Line: line + 1, Column: character}
					}(),
				},
				Raw: func() string {
					line, _ := utils.GetLineAndCharacterOfPosition(p.SourceCode, node.Loc.Pos())
					lines := strings.Split(p.SourceCode, "\n")
					if line >= 0 && line < len(lines) {
						return strings.TrimSpace(lines[line])
					}
					return ""
				}(),
			})
		}

		// switch 语句是节点类型分发器。
		// 它根据当前节点的类型，调用相应的 `analyze...` 方法进行处理。
		switch node.Kind {
		case ast.KindImportDeclaration:
			p.analyzeImportDeclaration(node.AsImportDeclaration())
			return // 导入声明不需深入遍历其子节点。

		case ast.KindExportDeclaration:
			p.analyzeExportDeclaration(node.AsExportDeclaration())
			return // 导出声明同样不需深入遍历。

		case ast.KindExportAssignment:
			p.analyzeExportAssignment(node.AsExportAssignment())
			return // `export default` 也不需深入遍历。

		case ast.KindInterfaceDeclaration:
			p.analyzeInterfaceDeclaration(node.AsInterfaceDeclaration())

		case ast.KindTypeAliasDeclaration:
			p.analyzeTypeAliasDeclaration(node.AsTypeAliasDeclaration())

		case ast.KindEnumDeclaration:
			p.analyzeEnumDeclaration(node.AsEnumDeclaration())

		case ast.KindVariableStatement:
			p.analyzeVariableStatement(node.AsVariableStatement())

		case ast.KindCallExpression:
			p.analyzeCallExpression(node.AsCallExpression())

		case ast.KindJsxElement, ast.KindJsxSelfClosingElement:
			p.analyzeJsxElement(node)

		case ast.KindFunctionDeclaration:
			p.analyzeFunctionDeclaration(node.AsFunctionDeclaration())
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

// analyzeImportDeclaration 解析静态导入声明。
func (p *Parser) analyzeImportDeclaration(node *ast.ImportDeclaration) {
	idr := NewImportDeclarationResult()
	idr.Raw = utils.GetNodeText(node.AsNode(), p.SourceCode)
	idr.Source = node.ModuleSpecifier.Text()
	pos, end := node.Pos(), node.End()
	idr.SourceLocation = SourceLocation{
		Start: NodePosition{Line: pos, Column: 0},
		End:   NodePosition{Line: end, Column: 0},
	}

	if node.ImportClause == nil { // 处理副作用导入，例如 `import './setup';`
		p.Result.ImportDeclarations = append(p.Result.ImportDeclarations, *idr)
		return
	}

	importClause := node.ImportClause.AsImportClause()

	if ast.IsDefaultImport(node.AsNode()) {
		name := importClause.Name().Text()
		idr.addModule("default", "default", name)
	}

	if namespaceNode := ast.GetNamespaceDeclarationNode(node.AsNode()); namespaceNode != nil {
		name := namespaceNode.Name().Text()
		idr.addModule("namespace", name, name)
	}

	if importClause.NamedBindings != nil && importClause.NamedBindings.Kind == ast.KindNamedImports {
		namedImports := importClause.NamedBindings.AsNamedImports()
		for _, element := range namedImports.Elements.Nodes {
			importSpecifier := element.AsImportSpecifier()
			identifier := importSpecifier.Name().Text()
			importModule := identifier
			if importSpecifier.PropertyName != nil {
				importModule = importSpecifier.PropertyName.Text()
			}
			idr.addModule("named", importModule, identifier)
		}
	}
	p.Result.ImportDeclarations = append(p.Result.ImportDeclarations, *idr)
}

// analyzeExportDeclaration 解析导出声明。
func (p *Parser) analyzeExportDeclaration(node *ast.ExportDeclaration) {
	edr := NewExportDeclarationResult(node)
	edr.Raw = utils.GetNodeText(node.AsNode(), p.SourceCode)

	if node.ModuleSpecifier != nil {
		edr.Source = node.ModuleSpecifier.Text()
		edr.Type = "re-export"
	} else {
		edr.Type = "named-export"
	}

	if node.ExportClause != nil {
		if node.ExportClause.Kind == ast.KindNamedExports {
			namedExports := node.ExportClause.AsNamedExports()
			for _, element := range namedExports.Elements.Nodes {
				specifier := element.AsExportSpecifier()
				identifier := specifier.Name().Text()
				moduleName := identifier
				if specifier.PropertyName != nil {
					moduleName = specifier.PropertyName.Text()
				}
				edr.ExportModules = append(edr.ExportModules, ExportModule{
					ModuleName: moduleName,
					Type:       "named",
					Identifier: identifier,
				})
			}
		} else if node.ExportClause.Kind == ast.KindNamespaceExport {
			namespaceExport := node.ExportClause.AsNamespaceExport()
			identifier := namespaceExport.Name().Text()
			edr.ExportModules = append(edr.ExportModules, ExportModule{
				ModuleName: "*",
				Type:       "namespace",
				Identifier: identifier,
			})
		}
	} else {
		if edr.Source != "" {
			edr.ExportModules = append(edr.ExportModules, ExportModule{
				ModuleName: "*",
				Type:       "namespace",
				Identifier: "*",
			})
		}
	}
	p.Result.ExportDeclarations = append(p.Result.ExportDeclarations, *edr)
}

// analyzeExportAssignment 解析 `export default` 声明。
func (p *Parser) analyzeExportAssignment(node *ast.ExportAssignment) {
	ear := NewExportAssignmentResult(node)
	ear.Raw = utils.GetNodeText(node.AsNode(), p.SourceCode)
	ear.Expression = strings.TrimSpace(utils.GetNodeText(node.Expression, p.SourceCode))
	p.Result.ExportAssignments = append(p.Result.ExportAssignments, *ear)
}

// analyzeInterfaceDeclaration 解析接口声明。
func (p *Parser) analyzeInterfaceDeclaration(node *ast.InterfaceDeclaration) {
	inter := NewInterfaceDeclarationResult(node.AsNode(), p.SourceCode)
	interfaceName := node.Name().Text()
	inter.Identifier = interfaceName

	// 检查导出关键字
	if modifiers := node.Modifiers(); modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if modifier != nil && modifier.Kind == ast.KindExportKeyword {
				inter.Exported = true
				break
			}
		}
	}

	// 分析 `extends` 子句
	extendsElements := ast.GetExtendsHeritageClauseElements(node.AsNode())
	for _, element := range extendsElements {
		expression := element.Expression()
		if ast.IsIdentifier(expression) {
			name := expression.AsIdentifier().Text
			if !(utils.IsUtilityType(name)) {
				inter.addTypeReference(name, "", true)
			}
		} else if ast.IsPropertyAccessExpression(expression) {
			name := entityNameToString(expression)
			inter.addTypeReference(name, "", true)
		}

		if len(element.TypeArguments()) > 0 {
			for _, typeArg := range element.TypeArguments() {
				results := AnalyzeType(typeArg, "")
				for _, res := range results {
					inter.addTypeReference(res.TypeName, res.Location, true)
				}
			}
		}
	}

	// 分析接口成员
	if node.Members != nil {
		for _, member := range node.Members.Nodes {
			results := AnalyzeMember(member, interfaceName)
			for _, res := range results {
				inter.addTypeReference(res.TypeName, res.Location, false)
			}
		}
	}
	p.Result.InterfaceDeclarations[inter.Identifier] = *inter
}

// analyzeTypeAliasDeclaration 解析 `type` 别名声明。
func (p *Parser) analyzeTypeAliasDeclaration(node *ast.TypeAliasDeclaration) {
	tr := NewTypeDeclarationResult(node.AsNode(), p.SourceCode)
	typeName := node.Name().Text()
	tr.Identifier = typeName

	// 检查导出关键字
	if modifiers := node.Modifiers(); modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if modifier != nil && modifier.Kind == ast.KindExportKeyword {
				tr.Exported = true
				break
			}
		}
	}

	results := AnalyzeType(node.Type, typeName)
	for _, res := range results {
		tr.addTypeReference(res.TypeName, res.Location, false)
	}
	p.Result.TypeDeclarations[tr.Identifier] = *tr
}

// analyzeEnumDeclaration 解析枚举声明。
func (p *Parser) analyzeEnumDeclaration(node *ast.EnumDeclaration) {
	er := NewEnumDeclarationResult(node, p.SourceCode)

	// 检查导出关键字
	if modifiers := node.Modifiers(); modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if modifier != nil && modifier.Kind == ast.KindExportKeyword {
				er.Exported = true
				break
			}
		}
	}

	p.Result.EnumDeclarations[er.Identifier] = *er
}

// analyzeVariableStatement 解析变量声明语句。
// 这里的逻辑经过了重构，以支持对赋值给变量的函数表达式（箭头函数、匿名函数）进行解析。
func (p *Parser) analyzeVariableStatement(node *ast.VariableStatement) {
	// 检查 `export` 修饰符
	isExported := false
	if modifiers := node.Modifiers(); modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if modifier != nil && modifier.Kind == ast.KindExportKeyword {
				isExported = true
				break
			}
		}
	}

	declarationList := node.DeclarationList
	if declarationList == nil {
		return
	}

	// 遍历声明列表中的每一个声明 (例如 `const a = 1, b = 2`)
	for _, decl := range declarationList.AsVariableDeclarationList().Declarations.Nodes {
		variableDecl := decl.AsVariableDeclaration()
		if variableDecl == nil {
			continue
		}

		nameNode := variableDecl.Name()
		initializerNode := variableDecl.Initializer

		// --- 函数表达式检查逻辑 ---
		// 检查变量的初始值是否是一个函数或箭头函数
		if ast.IsIdentifier(nameNode) && initializerNode != nil {
			identifier := nameNode.AsIdentifier().Text
			initKind := initializerNode.Kind

			if initKind == ast.KindArrowFunction || initKind == ast.KindFunctionExpression {
				// 如果是，则使用新的构造函数来解析这个函数表达式
				fr := NewFunctionDeclarationResultFromExpression(identifier, isExported, initializerNode, p.SourceCode)
				p.Result.FunctionDeclarations = append(p.Result.FunctionDeclarations, *fr)
				// 解析为函数后，无需再作为普通变量处理，跳过当前循环
				continue
			}
		}
		// --- 函数表达式检查逻辑结束 ---

		// --- 动态导入检查逻辑 ---
		// 核心目的：将 `const AdminPage = lazy(() => import('./AdminPage'))` 这样的代码
		// 正确解析为 `AdminPage` 标识符和 `./AdminPage` 路径之间的关联。
		if ast.IsIdentifier(nameNode) && initializerNode != nil {
			identifier := nameNode.AsIdentifier().Text
			// 递归地在变量的初始化表达式中查找 `import()` 调用。
			importCallNode, importPath := p.findDynamicImport(initializerNode)

			if importCallNode != nil && importPath != "" {
				// 如果找到了，就创建一个精确的导入记录。
				importResult := &ImportDeclarationResult{
					Source: importPath,
					ImportModules: []ImportModule{
						{
							Identifier:   identifier, // 使用变量名作为导入的标识符
							ImportModule: "default",  // 动态导入可以看作是导入默认模块
							Type:         "dynamic_variable",
						},
					},
					Raw: utils.GetNodeText(importCallNode, p.SourceCode),
				}
				p.Result.ImportDeclarations = append(p.Result.ImportDeclarations, *importResult)
				// 标记此 `import()` 节点已处理，避免在 `analyzeCallExpression` 中重复记录。
				p.processedDynamicImports[importCallNode] = true
			}
		}

		// --- 常规变量和解构变量处理---
		vd := NewVariableDeclaration(node, p.SourceCode)
		vd.Exported = isExported
		if (declarationList.Flags & ast.NodeFlagsConst) != 0 {
			vd.Kind = ConstDeclaration
		} else if (declarationList.Flags & ast.NodeFlagsLet) != 0 {
			vd.Kind = LetDeclaration
		} else {
			vd.Kind = VarDeclaration
		}

		if ast.IsIdentifier(nameNode) {
			declarator := &VariableDeclarator{
				Identifier: nameNode.AsIdentifier().Text,
				Type:       analyzeVariableValueNode(variableDecl.Type, p.SourceCode),
				InitValue:  analyzeVariableValueNode(initializerNode, p.SourceCode),
			}
			vd.Declarators = append(vd.Declarators, declarator)
		} else if ast.IsObjectBindingPattern(nameNode) || ast.IsArrayBindingPattern(nameNode) {
			vd.Source = analyzeVariableValueNode(initializerNode, p.SourceCode)
			p.analyzeBindingPattern(nameNode, vd)
		}
		p.Result.VariableDeclarations = append(p.Result.VariableDeclarations, *vd)
	}
}

// findDynamicImport 递归地在给定的 AST 节点中查找第一个 `import()` 调用。
// 它会深入常见的包装函数（如 `lazy`, `() => ...`）内部进行查找。
// 返回找到的 `import()` 对应的 ast.Node 和导入的路径字符串。
func (p *Parser) findDynamicImport(node *ast.Node) (*ast.Node, string) {
	if node == nil {
		return nil, ""
	}

	// 基本情况：当前节点就是 `import()` 调用。
	if node.Kind == ast.KindCallExpression {
		callExpr := node.AsCallExpression()
		if callExpr.Expression.Kind == ast.KindImportKeyword {
			if len(callExpr.Arguments.Nodes) > 0 {
				arg := callExpr.Arguments.Nodes[0]
				if arg.Kind == ast.KindStringLiteral {
					return node, arg.AsStringLiteral().Text
				} else if arg.Kind == ast.KindIdentifier {
					return node, arg.AsIdentifier().Text
				}
			}
		}
	}

	// 递归情况：遍历子节点查找。
	var foundNode *ast.Node
	var foundPath string
	node.ForEachChild(func(child *ast.Node) bool {
		// 如果已经找到了，就停止遍历，防止找到更深层的无关 `import`。
		if foundNode != nil {
			return true // stop traversal
		}
		foundNode, foundPath = p.findDynamicImport(child)
		return foundNode != nil // 如果在子节点中找到了，返回 true 停止遍历
	})

	return foundNode, foundPath
}

// analyzeBindingPattern 解析解构模式。
func (p *Parser) analyzeBindingPattern(node *ast.Node, vd *VariableDeclaration) {
	if node == nil {
		return
	}

	bindingPattern := node.AsBindingPattern()
	if bindingPattern == nil || bindingPattern.Elements == nil {
		return
	}
	elements := bindingPattern.Elements.Nodes

	for _, element := range elements {
		bindingElement := element.AsBindingElement()
		if bindingElement == nil {
			continue
		}

		nameNode := bindingElement.Name()
		if nameNode == nil {
			continue
		}

		if ast.IsObjectBindingPattern(nameNode) || ast.IsArrayBindingPattern(nameNode) {
			p.analyzeBindingPattern(nameNode, vd)
		} else if ast.IsIdentifier(nameNode) {
			identifier := nameNode.AsIdentifier().Text
			propName := identifier

			if propertyNameNode := bindingElement.PropertyName; propertyNameNode != nil {
				switch propertyNameNode.Kind {
				case ast.KindIdentifier:
					if propIdentifier := propertyNameNode.AsIdentifier(); propIdentifier != nil {
						propName = propIdentifier.Text
					}
				case ast.KindStringLiteral:
					if strLit := propertyNameNode.AsStringLiteral(); strLit != nil {
						propName = strLit.Text
					}
				case ast.KindNumericLiteral:
					if numLit := propertyNameNode.AsNumericLiteral(); numLit != nil {
						propName = numLit.Text
					}
				case ast.KindComputedPropertyName:
					propName = strings.TrimSpace(utils.GetNodeText(propertyNameNode.AsNode(), p.SourceCode))
				default:
					propName = strings.TrimSpace(utils.GetNodeText(propertyNameNode.AsNode(), p.SourceCode))
				}
			}

			declarator := &VariableDeclarator{
				Identifier: identifier,
				PropName:   propName,
				InitValue:  analyzeVariableValueNode(bindingElement.Initializer, p.SourceCode),
			}
			vd.Declarators = append(vd.Declarators, declarator)
		}
	}
}

// analyzeFunctionDeclaration 解析函数声明。
// 此函数不仅提取函数的基本信息，还负责检查参数和返回类型中的显式 any。
func (p *Parser) analyzeFunctionDeclaration(node *ast.FunctionDeclaration) {
	// 1. 解析函数声明本身的信息
	fr := NewFunctionDeclarationResult(node, p.SourceCode)
	// 3. 将解析结果存入
	p.Result.FunctionDeclarations = append(p.Result.FunctionDeclarations, *fr)
}

// analyzeJsxElement 解析 JSX 元素（包括自闭合和非自闭合的）。
func (p *Parser) analyzeJsxElement(node *ast.Node) {
	jsxNode := NewJSXNode(*node, p.SourceCode)
	if node.Kind == ast.KindJsxElement {
		p.analyzeJsxOpeningElement(jsxNode, node.AsJsxElement().OpeningElement)
	} else if node.Kind == ast.KindJsxSelfClosingElement {
		p.analyzeJsxSelfClosingElement(jsxNode, node.AsJsxSelfClosingElement())
	}
	p.Result.JsxElements = append(p.Result.JsxElements, *jsxNode)
}

// analyzeJsxOpeningElement 解析 JSX 的开标签。
func (p *Parser) analyzeJsxOpeningElement(jsxNode *JSXElement, node *ast.Node) {
	openingElement := node.AsJsxOpeningElement()
	jsxNode.ComponentChain = reconstructJSXName(openingElement.TagName)

	if attributes := openingElement.Attributes; attributes != nil {
		if jsxAttrs := attributes.AsJsxAttributes(); jsxAttrs != nil && jsxAttrs.Properties != nil {
			for _, attr := range jsxAttrs.Properties.Nodes {
				if attr.Kind == ast.KindJsxAttribute {
					jsxAttr := attr.AsJsxAttribute()
					jsxNode.Attrs = append(jsxNode.Attrs, JSXAttribute{
						Name:     jsxAttr.Name().Text(),
						Value:    analyzeAttributeValue(jsxAttr.Initializer, p.SourceCode),
						IsSpread: false,
					})
				} else if attr.Kind == ast.KindJsxSpreadAttribute {
					jsxSpreadAttr := attr.AsJsxSpreadAttribute()
					jsxNode.Attrs = append(jsxNode.Attrs, JSXAttribute{
						Name:     "..." + utils.GetNodeText(jsxSpreadAttr.Expression, p.SourceCode),
						Value:    nil,
						IsSpread: true,
					})
				}
			}
		}
	}
}

// analyzeJsxSelfClosingElement 解析 JSX 的自闭合标签。
func (p *Parser) analyzeJsxSelfClosingElement(jsxNode *JSXElement, node *ast.JsxSelfClosingElement) {
	jsxNode.ComponentChain = reconstructJSXName(node.TagName)

	if attributes := node.Attributes; attributes != nil {
		if jsxAttrs := attributes.AsJsxAttributes(); jsxAttrs != nil && jsxAttrs.Properties != nil {
			for _, attr := range jsxAttrs.Properties.Nodes {
				if attr.Kind == ast.KindJsxAttribute {
					jsxAttr := attr.AsJsxAttribute()
					jsxNode.Attrs = append(jsxNode.Attrs, JSXAttribute{
						Name:     jsxAttr.Name().Text(),
						Value:    analyzeAttributeValue(jsxAttr.Initializer, p.SourceCode),
						IsSpread: false,
					})
				} else if attr.Kind == ast.KindJsxSpreadAttribute {
					jsxSpreadAttr := attr.AsJsxSpreadAttribute()
					jsxNode.Attrs = append(jsxNode.Attrs, JSXAttribute{
						Name:     "..." + utils.GetNodeText(jsxSpreadAttr.Expression, p.SourceCode),
						Value:    nil,
						IsSpread: true,
					})
				}
			}
		}
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
	FunctionDeclarations  []FunctionDeclarationResult // 新增：用于存储找到的所有函数声明的信息
	AnyDeclarations       []AnyInfo                   // 新增：用于存储找到的所有 any 类型的信息
}

// AnyInfo 存储了在文件中找到的 any 类型的信息。
type AnyInfo struct {
	SourceLocation SourceLocation
	Raw            string // 存储 any 关键字的原始文本
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
		AnyDeclarations:       []AnyInfo{},
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
		AnyDeclarations:       pr.AnyDeclarations,
	}
}

// Traverse 是旧的入口点，现在它将工作委托给新的 Parser 结构。
// 这样做是为了保持对外的 API 兼容性。
func (pr *ParserResult) Traverse() {
	p, err := NewParser(pr.filePath)
	if err != nil {
		fmt.Printf("Error creating parser: %v\n", err)
		return
	}
	p.Traverse()
	*pr = *p.Result
}
