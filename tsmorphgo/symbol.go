package tsmorphgo

import (
	"context"
	"fmt"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/lsp"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/checker"
)

// Symbol 代表一个语义符号，它是连接代码中多个引用的核心。
// 例如，一个变量的声明和它的所有使用之处，都指向同一个 Symbol。
// 符号系统是 TypeScript 语义分析的基础，提供了代码元素之间的语义连接。
type Symbol struct {
	inner      *ast.Symbol      // 底层符号对象
	checker    *checker.Checker // 类型检查器实例
	sourceFile *SourceFile      // 所属的源文件
	lspService *lsp.Service     // LSP 服务（用于符号相关操作）
}

// SymbolFlags 表示符号的种类和特性
// 这些标志与底层 typescript-go 库的 ast.SymbolFlags 保持一致
type SymbolFlags uint32

const (
	// 变量相关标志
	SymbolFlagsFunctionScopedVariable SymbolFlags = 1 << iota // 函数作用域变量（var或参数）
	SymbolFlagsBlockScopedVariable                            // 块作用域变量（let或const）
	SymbolFlagsProperty                                       // 属性或枚举成员
	SymbolFlagsEnumMember                                     // 枚举成员
	SymbolFlagsFunction                                       // 函数
	SymbolFlagsClass                                          // 类
	SymbolFlagsInterface                                      // 接口
	SymbolFlagsConstEnum                                      // Const enum
	SymbolFlagsRegularEnum                                    // 枚举
	SymbolFlagsValueModule                                    // Instantiated module
	SymbolFlagsNamespaceModule                                // Uninstantiated module
	SymbolFlagsTypeLiteral                                    // Type Literal or mapped type
	SymbolFlagsObjectLiteral                                  // Object Literal
	SymbolFlagsMethod                                         // Method
	SymbolFlagsConstructor                                    // Constructor
	SymbolFlagsGetAccessor                                    // Get accessor
	SymbolFlagsSetAccessor                                    // Set accessor
	SymbolFlagsSignature                                      // Call, construct, or index signature
	SymbolFlagsTypeParameter                                  // Type parameter
	SymbolFlagsTypeAlias                                      // Type alias
	SymbolFlagsExportValue                                    // Exported value marker
	SymbolFlagsAlias                                          // An alias for another symbol
	SymbolFlagsPrototype                                      // Prototype property (no source representation)
	SymbolFlagsExportStar                                     // Export * declaration
	SymbolFlagsOptional                                       // Optional property

	// 组合标志，方便使用
	SymbolFlagsEnum      = SymbolFlagsRegularEnum | SymbolFlagsConstEnum
	SymbolFlagsVariable  = SymbolFlagsFunctionScopedVariable | SymbolFlagsBlockScopedVariable
	SymbolFlagsValue     = SymbolFlagsVariable | SymbolFlagsProperty | SymbolFlagsEnumMember | SymbolFlagsObjectLiteral | SymbolFlagsFunction | SymbolFlagsClass | SymbolFlagsEnum | SymbolFlagsValueModule | SymbolFlagsMethod | SymbolFlagsGetAccessor | SymbolFlagsSetAccessor
	SymbolFlagsType      = SymbolFlagsClass | SymbolFlagsInterface | SymbolFlagsEnum | SymbolFlagsEnumMember | SymbolFlagsTypeLiteral | SymbolFlagsTypeParameter | SymbolFlagsTypeAlias
	SymbolFlagsNamespace = SymbolFlagsValueModule | SymbolFlagsNamespaceModule | SymbolFlagsEnum
	SymbolFlagsModule    = SymbolFlagsValueModule | SymbolFlagsNamespaceModule
	SymbolFlagsAccessor  = SymbolFlagsGetAccessor | SymbolFlagsSetAccessor
)

// GetName 返回符号的名称。
// 这是符号的基本标识符，如变量名、函数名、类名等。
func (s *Symbol) GetName() string {
	if s.inner == nil {
		return ""
	}
	return s.inner.Name
}

// GetFlags 返回符号的标志，用于判断符号的种类和特性。
// 返回的标志可以用来确定符号是变量、函数、类还是其他类型。
func (s *Symbol) GetFlags() SymbolFlags {
	if s.inner == nil {
		return 0
	}
	return SymbolFlags(s.inner.Flags)
}

// IsExported 检查符号是否被导出（export）。
// 导出的符号可以被其他文件引用。
func (s *Symbol) IsExported() bool {
	if s.inner == nil {
		return false
	}
	return s.inner.Flags&ast.SymbolFlagsExportValue != 0
}

// IsVariable 检查符号是否是变量。
// 包括函数作用域变量和块作用域变量。
func (s *Symbol) IsVariable() bool {
	flags := s.GetFlags()
	return flags&SymbolFlagsFunctionScopedVariable != 0 ||
		flags&SymbolFlagsBlockScopedVariable != 0
}

// IsFunction 检查符号是否是函数。
func (s *Symbol) IsFunction() bool {
	return s.GetFlags()&SymbolFlagsFunction != 0
}

// IsClass 检查符号是否是类。
func (s *Symbol) IsClass() bool {
	return s.GetFlags()&SymbolFlagsClass != 0
}

// IsInterface 检查符号是否是接口。
func (s *Symbol) IsInterface() bool {
	return s.GetFlags()&SymbolFlagsInterface != 0
}

// IsEnum 检查符号是否是枚举。
func (s *Symbol) IsEnum() bool {
	return s.GetFlags()&SymbolFlagsEnum != 0
}

// IsTypeAlias 检查符号是否是类型别名。
func (s *Symbol) IsTypeAlias() bool {
	return s.GetFlags()&SymbolFlagsTypeAlias != 0
}

// IsModule 检查符号是否是模块。
func (s *Symbol) IsModule() bool {
	return s.GetFlags()&SymbolFlagsModule != 0
}

// IsAlias 检查符号是否是另一个符号的别名。
func (s *Symbol) IsAlias() bool {
	return s.GetFlags()&SymbolFlagsAlias != 0
}

// IsMethod 检查符号是否是方法。
func (s *Symbol) IsMethod() bool {
	return s.GetFlags()&SymbolFlagsMethod != 0
}

// IsConstructor 检查符号是否是构造函数。
func (s *Symbol) IsConstructor() bool {
	return s.GetFlags()&SymbolFlagsConstructor != 0
}

// IsAccessor 检查符号是否是访问器（getter/setter）。
func (s *Symbol) IsAccessor() bool {
	return s.GetFlags()&SymbolFlagsAccessor != 0
}

// IsOptional 检查符号是否是可选的。
func (s *Symbol) IsOptional() bool {
	return s.GetFlags()&SymbolFlagsOptional != 0
}

// HasValue 检查符号是否具有值（不仅仅是类型）。
func (s *Symbol) HasValue() bool {
	return s.GetFlags()&SymbolFlagsValue != 0
}

// HasType 检查符号是否具有类型信息。
func (s *Symbol) HasType() bool {
	return s.GetFlags()&SymbolFlagsType != 0
}

// IsTypeParameter 检查符号是否是类型参数。
func (s *Symbol) IsTypeParameter() bool {
	return s.GetFlags()&SymbolFlagsTypeParameter != 0
}

// IsEnumMember 检查符号是否是枚举成员。
func (s *Symbol) IsEnumMember() bool {
	return s.GetFlags()&SymbolFlagsEnumMember != 0
}

// IsProperty 检查符号是否是属性。
func (s *Symbol) IsProperty() bool {
	return s.GetFlags()&SymbolFlagsProperty != 0
}

// IsObjectLiteral 检查符号是否是对象字面量。
func (s *Symbol) IsObjectLiteral() bool {
	return s.GetFlags()&SymbolFlagsObjectLiteral != 0
}

// IsTypeLiteral 检查符号是否是类型字面量。
func (s *Symbol) IsTypeLiteral() bool {
	return s.GetFlags()&SymbolFlagsTypeLiteral != 0
}

// GetDeclarationCount 返回符号的声明数量。
// 对于大多数简单符号，这个值为1，但对于函数重载等情况可能大于1。
func (s *Symbol) GetDeclarationCount() int {
	if s.inner == nil {
		return 0
	}
	return len(s.inner.Declarations)
}

// GetDeclarations 返回符号的所有声明节点。
// 一个符号可能有多个声明（如函数重载、命名空间合并等）。
func (s *Symbol) GetDeclarations() []Node {
	if s.inner == nil || len(s.inner.Declarations) == 0 {
		return nil
	}

	declarations := make([]Node, 0, len(s.inner.Declarations))
	for _, decl := range s.inner.Declarations {
		if decl != nil {
			declarations = append(declarations, Node{
				Node:       decl,
				sourceFile: s.sourceFile,
			})
		}
	}
	return declarations
}

// GetFirstDeclaration 返回符号的第一个声明节点。
// 对于大多数简单符号，这是它们唯一的声明。
func (s *Symbol) GetFirstDeclaration() (*Node, bool) {
	declarations := s.GetDeclarations()
	if len(declarations) == 0 {
		return nil, false
	}
	return &declarations[0], true
}

// GetParent 返回符号的父符号。
// 例如，类方法的父符号是类本身。
func (s *Symbol) GetParent() (*Symbol, bool) {
	if s.inner == nil || s.inner.Parent == nil {
		return nil, false
	}
	return &Symbol{
		inner:      s.inner.Parent,
		checker:    s.checker,
		sourceFile: s.sourceFile,
		lspService: s.lspService,
	}, true
}

// GetMembers 返回符号的成员符号表。
// 主要用于类、接口、对象字面量等具有成员的符号。
func (s *Symbol) GetMembers() map[string]*Symbol {
	if s.inner == nil || s.inner.Members == nil {
		return make(map[string]*Symbol) // 返回空 map 而不是 nil
	}

	members := make(map[string]*Symbol)
	for name, memberSymbol := range s.inner.Members {
		if memberSymbol != nil {
			members[name] = &Symbol{
				inner:      memberSymbol,
				checker:    s.checker,
				sourceFile: s.sourceFile,
				lspService: s.lspService,
			}
		}
	}
	return members
}

// GetExports 返回符号的导出符号表。
// 主要用于模块/命名空间的导出成员。
func (s *Symbol) GetExports() map[string]*Symbol {
	if s.inner == nil || s.inner.Exports == nil {
		return make(map[string]*Symbol) // 返回空 map 而不是 nil
	}

	exports := make(map[string]*Symbol)
	for name, exportSymbol := range s.inner.Exports {
		if exportSymbol != nil {
			exports[name] = &Symbol{
				inner:      exportSymbol,
				checker:    s.checker,
				sourceFile: s.sourceFile,
				lspService: s.lspService,
			}
		}
	}
	return exports
}

// GetSymbolAtLocation 通过 LanguageService 获取指定位置的符号。
// 这是一个更可靠的符号获取方法，利用了 LSP 服务的能力。
func (s *Symbol) GetSymbolAtLocation(node Node) (*Symbol, bool) {
	if s.lspService == nil {
		return nil, false
	}

	filePath := node.GetSourceFile().GetFilePath()
	startLine := node.GetStartLineNumber()
	// 简化处理，使用节点的起始位置作为字符位置
	char := 0

	// 使用 query service 获取符号
	symbol, err := s.lspService.GetSymbolAt(context.Background(), filePath, startLine, char)
	if err != nil || symbol == nil {
		return nil, false
	}

	return &Symbol{
		inner:      symbol,
		checker:    s.checker,
		sourceFile: node.sourceFile,
		lspService: s.lspService,
	}, true
}

// FindReferences 查找该符号的所有引用位置。
// 返回包含该符号引用的所有节点。
func (s *Symbol) FindReferences() ([]Node, error) {
	if s.inner == nil || len(s.inner.Declarations) == 0 {
		return []Node{}, nil // 返回空 slice 而不是 nil
	}

	// 使用第一个声明节点来查找引用
	firstDecl := s.inner.Declarations[0]
	if firstDecl == nil {
		return []Node{}, nil // 返回空 slice 而不是 nil
	}

	declNode := Node{
		Node:       firstDecl,
		sourceFile: s.sourceFile,
	}

	// 重用现有的 FindReferences 实现
	referenceNodes, err := FindReferences(declNode)
	if err != nil {
		return nil, fmt.Errorf("failed to find references: %w", err)
	}

	// 转换为 Node 数组
	result := make([]Node, len(referenceNodes))
	for i, refNode := range referenceNodes {
		if refNode != nil {
			result[i] = *refNode
		}
	}

	return result, nil
}

// String 返回符号的字符串表示，用于调试。
func (s *Symbol) String() string {
	if s.inner == nil {
		return "<nil symbol>"
	}
	return fmt.Sprintf("Symbol{name: %s, flags: %d}", s.inner.Name, s.inner.Flags)
}

// GetSymbol 获取给定节点关联的语义符号。
// 这是一个模拟实现，主要用于测试目的，它会尝试构建一个包含父子和成员关系的符号层级结构。
// 完整的符号获取功能需要更复杂的底层API集成。
//
// 参数:
//   - node: 要获取符号的AST节点
//
// 返回:
//   - *Symbol: 找到的符号对象
//   - bool: 是否成功找到符号
//
// 示例:
//
//	symbol, found := GetSymbol(identifierNode)
//	if found {
//	    fmt.Printf("Symbol name: %s\n", symbol.GetName())
//	    fmt.Printf("Is exported: %v\n", symbol.IsExported())
//	}
func GetSymbol(node Node) (*Symbol, bool) {
	if node.sourceFile == nil || node.sourceFile.project == nil {
		return nil, false
	}
	// 调用内部的递归构建函数，初始时没有父符号
	return buildMockSymbol(node, nil)
}

// buildMockSymbol 是一个递归函数，用于构建模拟的符号及其层级关系。
// node: 当前要为其构建符号的节点 (通常是一个标识符)。
// parentAstSymbol: 父级的 ast.Symbol，用于构建层级关系。
func buildMockSymbol(node Node, parentAstSymbol *ast.Symbol) (*Symbol, bool) {
	// 1. 获取当前节点的基本信息
	nodeText := strings.TrimSpace(node.GetText())
	declarationNode := node.GetParent()
	if declarationNode == nil {
		return nil, false // 如果没有父节点（声明节点），则无法创建符号
	}

	// 2. 根据声明节点的类型，确定符号的标志 (flags)
	flags := ast.SymbolFlagsNone
	switch declarationNode.Kind {
	case ast.KindVariableDeclaration:
		flags |= ast.SymbolFlagsBlockScopedVariable | ast.SymbolFlagsValue
	case ast.KindFunctionDeclaration:
		flags |= ast.SymbolFlagsFunction | ast.SymbolFlagsValue
	case ast.KindClassDeclaration:
		flags |= ast.SymbolFlagsClass | ast.SymbolFlagsValue | ast.SymbolFlagsType
	case ast.KindInterfaceDeclaration:
		flags |= ast.SymbolFlagsInterface | ast.SymbolFlagsType
		// 接口只有类型，没有值
		flags &^= ast.SymbolFlagsValue
	case ast.KindMethodDeclaration:
		flags |= ast.SymbolFlagsMethod | ast.SymbolFlagsValue
	case ast.KindPropertyDeclaration, ast.KindPropertySignature:
		flags |= ast.SymbolFlagsProperty | ast.SymbolFlagsValue
	case ast.KindGetAccessor:
		flags |= ast.SymbolFlagsGetAccessor | ast.SymbolFlagsValue
	case ast.KindSetAccessor:
		flags |= ast.SymbolFlagsSetAccessor | ast.SymbolFlagsValue
	}

	// 检查导出标志 (更鲁棒的检查)
	if isNodeExported(*declarationNode) {
		flags |= ast.SymbolFlagsExportValue
	}

	// 3. 创建底层的 ast.Symbol
	innerSymbol := &ast.Symbol{
		Name:         nodeText,
		Flags:        flags,
		Declarations: []*ast.Node{declarationNode.Node},
		Parent:       parentAstSymbol, // 设置父符号
	}

	// 4. 如果是类或接口，则递归构建其成员符号
	if declarationNode.Kind == ast.KindClassDeclaration || declarationNode.Kind == ast.KindInterfaceDeclaration {
		members := make(map[string]*ast.Symbol)
		// 遍历声明节点的子节点以查找成员
		declarationNode.ForEachChild(func(memberNode *ast.Node) bool {
			var memberNameNode *Node
			// 找到成员的名称节点 (Identifier)
			switch memberNode.Kind {
			case ast.KindMethodDeclaration, ast.KindPropertyDeclaration, ast.KindPropertySignature, ast.KindGetAccessor, ast.KindSetAccessor:
				// 查找第一个作为名称的 Identifier 子节点
				memberNode.ForEachChild(func(child *ast.Node) bool {
					if child.Kind == ast.KindIdentifier {
						memberNameNode = &Node{Node: child, sourceFile: node.sourceFile}
						return true // 找到后停止
					}
					return false
				})
			}

			// 如果找到了成员名称，为其构建符号
			if memberNameNode != nil {
				// 递归调用，并传入当前类/接口的符号作为父符号
				if memberSymbol, ok := buildMockSymbol(*memberNameNode, innerSymbol); ok {
					members[memberSymbol.GetName()] = memberSymbol.inner
				}
			}
			return false // 继续遍历其他成员
		})
		innerSymbol.Members = members
	}

	// 5. 返回包装后的 tsmorphgo.Symbol
	return &Symbol{
		inner:      innerSymbol,
		sourceFile: node.sourceFile,
	}, true
}

// isNodeExported 检查一个声明节点是否被导出。
func isNodeExported(declarationNode Node) bool {
	// 方式1: 检查声明节点自身是否有 'export' 修饰符
	if hasExportModifier(declarationNode) {
		return true
	}

	// 方式2: 检查父节点是否是 ExportDeclaration
	parent := declarationNode.GetParent()
	if parent != nil {
		if parent.Kind == ast.KindExportDeclaration {
			return true
		}

		// 方式3: 针对 `export const x = 1` 的情况
		// 结构是 VariableStatement -> VariableDeclarationList -> VariableDeclaration
		if declarationNode.Kind == ast.KindVariableDeclaration {
			// parent 是 VariableDeclarationList
			grandParent := parent.GetParent()
			if grandParent != nil && grandParent.Kind == ast.KindVariableStatement {
				if hasExportModifier(*grandParent) {
					return true
				}
			}
		}
	}

	return false
}

// hasExportModifier 检查节点是否具有 export 修饰符
func hasExportModifier(node Node) bool {
	// 在 typescript-go 中，修饰符通常可以通过特定的方法或属性获取
	// 这里使用一个简化的实现，检查节点的文本是否包含 export 关键字

	// 简化的检查：查看节点的完整文本是否包含 export
	nodeText := node.GetText()
	return strings.Contains(nodeText, "export ")
}

// createSymbolService 创建符号查询服务的辅助函数
func createSymbolService(project *Project) (*lsp.Service, error) {
	if project == nil || project.parserResult == nil {
		return nil, fmt.Errorf("invalid project or parser result")
	}

	// 构建源码映射
	sources := make(map[string]any, len(project.parserResult.Js_Data))
	for path, jsResult := range project.parserResult.Js_Data {
		sources[path] = jsResult.Raw
	}

	// 创建查询服务
	return lsp.NewServiceForTest(sources)
}
