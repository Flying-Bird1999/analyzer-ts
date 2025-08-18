package parser

import (
	"fmt"
	"main/analyzer/utils"
	"strings"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// ... (rest of the file is correct)

// Parser 定义了解析器的主要结构，包含了源码、AST 和最终的解析结果。
type Parser struct {
	SourceCode              string             // 文件的源码内容
	Ast                     *ast.Node          // 从源码解析出的 AST
	Result                  *ParserResult      // 存储解析结果的容器
	processedDynamicImports map[*ast.Node]bool // 用于标记已处理的动态导入节点，防止重复解析
}

// NewParser 创建并返回一个新的 Parser 实例。
func NewParser(filePath string) (*Parser, error) {
	sourceCode, err := utils.ReadFileContent(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	sourceFile := utils.ParseTypeScriptFile(filePath, sourceCode)
	return &Parser{
		SourceCode:              sourceCode,
		Ast:                     sourceFile.AsNode(),
		Result:                  NewParserResult(filePath),
		processedDynamicImports: make(map[*ast.Node]bool),
	}, nil
}

// Traverse 是解析器的核心驱动函数。
// 它通过深度优先遍历整个 AST 来识别和解析各种类型的节点。
func (p *Parser) Traverse() {
	var walk func(node *ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}

		switch node.Kind {
		case ast.KindImportDeclaration:
			p.analyzeImportDeclaration(node.AsImportDeclaration())
			return // 导入声明不需深入遍历

		case ast.KindExportDeclaration:
			p.analyzeExportDeclaration(node.AsExportDeclaration())
			return // 导出声明不需深入遍历

		case ast.KindExportAssignment:
			p.analyzeExportAssignment(node.AsExportAssignment())
			return // `export default` 不需深入遍历

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
		}

		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false // 继续遍历
		})
	}

	walk(p.Ast)
}

func (p *Parser) analyzeImportDeclaration(node *ast.ImportDeclaration) {
	idr := NewImportDeclarationResult()
	idr.Raw = utils.GetNodeText(node.AsNode(), p.SourceCode)
	idr.Source = node.ModuleSpecifier.Text()
	pos, end := node.Pos(), node.End()
	idr.SourceLocation = SourceLocation{
		Start: NodePosition{Line: pos, Column: 0},
		End:   NodePosition{Line: end, Column: 0},
	}

	if node.ImportClause == nil {
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

func (p *Parser) analyzeExportAssignment(node *ast.ExportAssignment) {
	ear := NewExportAssignmentResult(node)
	ear.Raw = utils.GetNodeText(node.AsNode(), p.SourceCode)
	ear.Expression = strings.TrimSpace(utils.GetNodeText(node.Expression, p.SourceCode))
	p.Result.ExportAssignments = append(p.Result.ExportAssignments, *ear)
}

func (p *Parser) analyzeInterfaceDeclaration(node *ast.InterfaceDeclaration) {
	inter := NewInterfaceDeclarationResult(node.AsNode(), p.SourceCode)
	interfaceName := node.Name().Text()
	inter.Identifier = interfaceName

	// Analyze heritage clauses (extends)
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

	// Analyze members
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

func (p *Parser) analyzeTypeAliasDeclaration(node *ast.TypeAliasDeclaration) {
	tr := NewTypeDeclarationResult(node.AsNode(), p.SourceCode)
	typeName := node.Name().Text()
	tr.Identifier = typeName

	results := AnalyzeType(node.Type, typeName)
	for _, res := range results {
		tr.addTypeReference(res.TypeName, res.Location, false)
	}
	p.Result.TypeDeclarations[tr.Identifier] = *tr
}

func (p *Parser) analyzeEnumDeclaration(node *ast.EnumDeclaration) {
	er := NewEnumDeclarationResult(node, p.SourceCode)
	p.Result.EnumDeclarations[er.Identifier] = *er
}

func (p *Parser) analyzeVariableStatement(node *ast.VariableStatement) {
	vd := NewVariableDeclaration(node, p.SourceCode)

	if modifiers := node.Modifiers(); modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if modifier != nil && modifier.Kind == ast.KindExportKeyword {
				vd.Exported = true
				break
			}
		}
	}

	declarationList := node.DeclarationList
	if declarationList == nil {
		return
	}
	if (declarationList.Flags & ast.NodeFlagsConst) != 0 {
		vd.Kind = ConstDeclaration
	} else if (declarationList.Flags & ast.NodeFlagsLet) != 0 {
		vd.Kind = LetDeclaration
	} else {
		vd.Kind = VarDeclaration
	}

	for _, decl := range declarationList.AsVariableDeclarationList().Declarations.Nodes {
		variableDecl := decl.AsVariableDeclaration()
		if variableDecl == nil {
			continue
		}

		nameNode := variableDecl.Name()
		initializerNode := variableDecl.Initializer

		// --- 新增逻辑：检查变量赋值中是否包含动态导入 ---
		if ast.IsIdentifier(nameNode) && initializerNode != nil {
			identifier := nameNode.AsIdentifier().Text
			importCallNode, importPath := p.findDynamicImport(initializerNode)

			if importCallNode != nil && importPath != "" {
				// 找到了一个动态导入赋值给变量，创建一个精确的导入记录
				importResult := &ImportDeclarationResult{
					Source: importPath,
					ImportModules: []ImportModule{
						{
							Identifier:   identifier, // 使用变量名作为 identifier
							ImportModule: "default",  // 动态导入可以看作是导入默认模块
							Type:         "dynamic_variable",
						},
					},
					Raw: utils.GetNodeText(importCallNode, p.SourceCode),
				}
				p.Result.ImportDeclarations = append(p.Result.ImportDeclarations, *importResult)
				// 标记此 import() 节点已处理，避免在 analyzeCallExpression 中重复处理
				p.processedDynamicImports[importCallNode] = true
			}
		}
		// --- 新增逻辑结束 ---

		if ast.IsIdentifier(nameNode) {
			declarator := &VariableDeclarator{
				Identifier: nameNode.AsIdentifier().Text,
				Type:       analyzeVariableValueNode(variableDecl.Type, p.SourceCode),
				InitValue:  analyzeVariableValueNode(initializerNode, p.SourceCode),
			}
			vd.Declarators = append(vd.Declarators, declarator)
			continue
		}

		if ast.IsObjectBindingPattern(nameNode) || ast.IsArrayBindingPattern(nameNode) {
			vd.Source = analyzeVariableValueNode(initializerNode, p.SourceCode)
			p.analyzeBindingPattern(nameNode, vd)
		}
	}
	p.Result.VariableDeclarations = append(p.Result.VariableDeclarations, *vd)
}

// findDynamicImport 递归地在给定的 AST 节点中查找第一个 `import()` 调用。
// 返回找到的 `import()` 的 ast.Node 和导入的路径字符串。
func (p *Parser) findDynamicImport(node *ast.Node) (*ast.Node, string) {
	if node == nil {
		return nil, ""
	}

	// 基本情况：当前节点就是 import() 调用
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

	// 递归情况：遍历子节点查找
	var foundNode *ast.Node
	var foundPath string
	node.ForEachChild(func(child *ast.Node) bool {
		// 如果已经找到了，就停止遍历
		if foundNode != nil {
			return true // stop traversal
		}
		// 递归查找
		foundNode, foundPath = p.findDynamicImport(child)
		// 如果在子节点中找到了，返回 true 停止遍历
		return foundNode != nil
	})

	return foundNode, foundPath
}

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

func (p *Parser) analyzeJsxElement(node *ast.Node) {
	jsxNode := NewJSXNode(*node, p.SourceCode)
	if node.Kind == ast.KindJsxElement {
		p.analyzeJsxOpeningElement(jsxNode, node.AsJsxElement().OpeningElement)
	} else if node.Kind == ast.KindJsxSelfClosingElement {
		p.analyzeJsxSelfClosingElement(jsxNode, node.AsJsxSelfClosingElement())
	}
	p.Result.JsxElements = append(p.Result.JsxElements, *jsxNode)
}

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
