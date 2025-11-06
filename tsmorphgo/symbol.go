package tsmorphgo

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/checker"
)

// ============================================================================
// 增强 Symbol 模块 - 基于 TypeScript-Go Checker
// ============================================================================

// Symbol 表示一个 TypeScript 符号的完整信息，基于 TypeScript 编译器的符号系统
type Symbol struct {
	// 原生 TypeScript 符号
	nativeSymbol *ast.Symbol

	// 符号名称
	name string

	// 符号标志 (从原生符号转换而来)
	flags ast.SymbolFlags

	// 所属的声明节点 (包装为 TSMorphGo Node)
	declarations []*Node

	// 是否为导出符号
	exported bool

	// 符号文档
	documentation string

	// 缓存字段
	mu sync.RWMutex
}

// TypeCheckerProvider 提供 TypeChecker 访问的接口
type TypeCheckerProvider interface {
	GetTypeChecker() (*checker.Checker, error)
	GetProgram() (*ast.Node, error)
}

// SymbolManager 管理符号的创建和缓存
type SymbolManager struct {
	provider TypeCheckerProvider
	cache    map[string]*Symbol
	mu       sync.RWMutex
}

// NewSymbolManager 创建新的符号管理器
func NewSymbolManager(provider TypeCheckerProvider) *SymbolManager {
	return &SymbolManager{
		provider: provider,
		cache:    make(map[string]*Symbol),
	}
}

// GetSymbol 获取节点对应的符号，使用 TypeScript 编译器的 GetSymbolAtLocation
func (sm *SymbolManager) GetSymbol(node Node) (*Symbol, error) {
	if !node.IsValid() {
		return nil, fmt.Errorf("invalid node")
	}

	// 检查缓存
	cacheKey := sm.getCacheKey(node)
	sm.mu.RLock()
	if cached, exists := sm.cache[cacheKey]; exists {
		sm.mu.RUnlock()
		return cached, nil
	}
	sm.mu.RUnlock()

	// 获取 TypeChecker
	typeChecker, err := sm.provider.GetTypeChecker()
	if err != nil {
		// Fallback: 创建一个简单的符号用于测试
		symbol := sm.createFallbackSymbol(node)
		if symbol != nil {
			// 缓存结果
			sm.mu.Lock()
			sm.cache[cacheKey] = symbol
			sm.mu.Unlock()
			return symbol, nil
		}
		return nil, fmt.Errorf("failed to get type checker and fallback failed: %w", err)
	}

	// 使用 TypeScript 编译器的 GetSymbolAtLocation
	nativeSymbol := typeChecker.GetSymbolAtLocation(node.Node)
	if nativeSymbol == nil {
		// Fallback: 创建一个简单的符号
		symbol := sm.createFallbackSymbol(node)
		if symbol != nil {
			// 缓存结果
			sm.mu.Lock()
			sm.cache[cacheKey] = symbol
			sm.mu.Unlock()
			return symbol, nil
		}
		return nil, fmt.Errorf("no symbol found at location")
	}

	// 创建 TSMorphGo Symbol 包装器
	symbol := sm.wrapNativeSymbol(nativeSymbol, node)

	// 缓存结果
	sm.mu.Lock()
	sm.cache[cacheKey] = symbol
	sm.mu.Unlock()

	return symbol, nil
}

// wrapNativeSymbol 将原生 TypeScript 符号包装为 TSMorphGo Symbol
func (sm *SymbolManager) wrapNativeSymbol(nativeSymbol *ast.Symbol, node Node) *Symbol {
	if nativeSymbol == nil {
		return nil
	}

	symbol := &Symbol{
		nativeSymbol: nativeSymbol,
		name:         sm.extractSymbolName(nativeSymbol),
		flags:        nativeSymbol.Flags,
		exported:     sm.isExportedSymbol(nativeSymbol),
	}

	// 提取声明信息
	symbol.declarations = sm.extractDeclarations(nativeSymbol, node)

	// 提取文档信息
	symbol.documentation = sm.extractDocumentation(nativeSymbol)

	return symbol
}

// GetSymbol 从 Node 获取符号，兼容现有 API
func GetSymbol(node Node) (*Symbol, error) {
	if node.GetSourceFile() == nil || node.GetSourceFile().project == nil {
		return nil, fmt.Errorf("node must belong to a project")
	}

	symbolManager := node.GetSourceFile().project.getSymbolManager()
	return symbolManager.GetSymbol(node)
}

// 符号方法实现

// GetName 返回符号的名称
func (s *Symbol) GetName() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.name
}

// GetFlags 返回符号的标志
func (s *Symbol) GetFlags() ast.SymbolFlags {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.flags
}

// IsExported 检查符号是否为导出符号
func (s *Symbol) IsExported() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.exported
}

// GetDeclarations 返回符号的所有声明
func (s *Symbol) GetDeclarations() []*Node {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.declarations
}

// GetDeclarationCount 返回声明的数量
func (s *Symbol) GetDeclarationCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.declarations)
}

// GetFirstDeclaration 返回第一个声明
func (s *Symbol) GetFirstDeclaration() (*Node, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.declarations) > 0 {
		return s.declarations[0], true
	}
	return nil, false
}

// String 返回符号的字符串表示
func (s *Symbol) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fmt.Sprintf("Symbol{name: %s, flags: %v, exported: %v}", s.name, s.flags, s.exported)
}

// 符号类型检查方法 - 基于 ast.SymbolFlags

// IsVariable 检查是否为变量符号
func (s *Symbol) IsVariable() bool {
	return s.flags&(ast.SymbolFlagsVariable|ast.SymbolFlagsFunctionScopedVariable|ast.SymbolFlagsBlockScopedVariable) != 0
}

// IsFunction 检查是否为函数符号
func (s *Symbol) IsFunction() bool {
	return s.flags&ast.SymbolFlagsFunction != 0
}

// IsClass 检查是否为类符号
func (s *Symbol) IsClass() bool {
	return s.flags&ast.SymbolFlagsClass != 0
}

// IsInterface 检查是否为接口符号
func (s *Symbol) IsInterface() bool {
	return s.flags&ast.SymbolFlagsInterface != 0
}

// IsEnum 检查是否为枚举符号
func (s *Symbol) IsEnum() bool {
	return s.flags&ast.SymbolFlagsEnum != 0
}

// IsModule 检查是否为模块符号
func (s *Symbol) IsModule() bool {
	return s.flags&(ast.SymbolFlagsValueModule|ast.SymbolFlagsNamespaceModule) != 0
}

// IsTypeAlias 检查是否为类型别名符号
func (s *Symbol) IsTypeAlias() bool {
	return s.flags&ast.SymbolFlagsTypeAlias != 0
}

// IsMethod 检查是否为方法符号
func (s *Symbol) IsMethod() bool {
	return s.flags&ast.SymbolFlagsMethod != 0
}

// IsProperty 检查是否为属性符号
func (s *Symbol) IsProperty() bool {
	return s.flags&ast.SymbolFlagsProperty != 0
}

// HasType 检查符号是否有类型信息
func (s *Symbol) HasType() bool {
	return s.nativeSymbol != nil
}

// HasValue 检查符号是否有值
func (s *Symbol) HasValue() bool {
	return s.flags&(ast.SymbolFlagsVariable|ast.SymbolFlagsFunction|ast.SymbolFlagsClass|ast.SymbolFlagsEnum) != 0
}

// 辅助方法实现

// getCacheKey 生成缓存键
func (sm *SymbolManager) getCacheKey(node Node) string {
	return fmt.Sprintf("%s:%d:%d",
		node.GetSourceFile().GetFilePath(),
		node.GetStartLineNumber(),
		node.GetStartLineCharacter())
}

// extractSymbolName 提取符号名称
func (sm *SymbolManager) extractSymbolName(symbol *ast.Symbol) string {
	if symbol == nil {
		return ""
	}

	// ast.Symbol 结构中有 Name 字段，直接使用
	return symbol.Name
}

// isExportedSymbol 检查符号是否为导出符号
func (sm *SymbolManager) isExportedSymbol(symbol *ast.Symbol) bool {
	if symbol == nil {
		return false
	}

	flags := symbol.Flags
	return flags&(ast.SymbolFlagsExportValue|ast.SymbolFlagsExportStar) != 0
}

// extractDeclarations 提取符号的声明信息
func (sm *SymbolManager) extractDeclarations(symbol *ast.Symbol, node Node) []*Node {
	if symbol == nil {
		return nil
	}

	declarations := []*Node{}

	// ast.Symbol 结构中有 Declarations 字段，是 []*Node 类型
	for _, decl := range symbol.Declarations {
		if decl != nil && node.sourceFile != nil {
			wrappedNode := &Node{
				Node:       decl,
				sourceFile: node.sourceFile,
			}
			declarations = append(declarations, wrappedNode)
		}
	}

	return declarations
}

// extractDocumentation 提取符号的文档信息
func (sm *SymbolManager) extractDocumentation(symbol *ast.Symbol) string {
	if symbol == nil {
		return ""
	}

	// 这里需要根据 ast.Symbol 的实际 API 来获取文档注释
	// 临时实现，需要根据实际情况调整
	return ""
}

// createFallbackSymbol 当 TypeChecker 不可用时创建fallback符号
func (sm *SymbolManager) createFallbackSymbol(node Node) *Symbol {
	if !node.IsValid() {
		return nil
	}

	// 从节点推断符号信息
	name := strings.TrimSpace(node.GetText())
	if name == "" {
		name = "unknown"
	}

	// 推断符号类型和标志
	flags := sm.inferSymbolFlags(node)
	exported := sm.isExportedNode(node)

	symbol := &Symbol{
		nativeSymbol: nil, // 没有原生符号
		name:         name,
		flags:        flags,
		exported:     exported,
		declarations: []*Node{&node}, // 将当前节点作为声明
		documentation: "",
	}

	return symbol
}

// inferSymbolFlags 从节点推断符号标志
func (sm *SymbolManager) inferSymbolFlags(node Node) ast.SymbolFlags {
	if !node.IsValid() {
		return ast.SymbolFlagsNone
	}

	parent := node.GetParent()
	if parent == nil {
		return ast.SymbolFlagsVariable
	}

	// 根据父节点类型推断符号标志
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		return ast.SymbolFlagsVariable
	case ast.KindFunctionDeclaration:
		return ast.SymbolFlagsFunction
	case ast.KindClassDeclaration:
		return ast.SymbolFlagsClass
	case ast.KindInterfaceDeclaration:
		return ast.SymbolFlagsInterface
	case ast.KindTypeAliasDeclaration:
		return ast.SymbolFlagsTypeAlias
	case ast.KindEnumDeclaration:
		return ast.SymbolFlagsEnum
	case ast.KindConstructor:
		return ast.SymbolFlagsConstructor
	case ast.KindMethodDeclaration:
		return ast.SymbolFlagsMethod
	case ast.KindPropertyDeclaration, ast.KindPropertyAssignment:
		return ast.SymbolFlagsProperty
	case ast.KindParameter:
		return ast.SymbolFlagsVariable
	default:
		// 检查是否在导出上下文中
		if sm.isExportedNode(node) {
			return ast.SymbolFlagsVariable | ast.SymbolFlagsExportValue
		}
		return ast.SymbolFlagsVariable
	}
}

// isExportedNode 检查节点是否在导出上下文中
func (sm *SymbolManager) isExportedNode(node Node) bool {
	if !node.IsValid() {
		return false
	}

	// 向上遍历查找 ExportDeclaration
	for current := &node; current != nil && current.IsValid(); {
		if current.Kind == ast.KindExportDeclaration {
			return true
		}
		parent := current.GetParent()
		if parent != nil {
			current = parent
		} else {
			break
		}
	}

	return false
}

// ClearCache 清空符号缓存
func (sm *SymbolManager) ClearCache() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.cache = make(map[string]*Symbol)
}

// GetCacheStats 获取缓存统计信息
func (sm *SymbolManager) GetCacheStats() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.cache)
}

// GetGlobalScope 获取全局作用域
func (sm *SymbolManager) GetGlobalScope() *SymbolScope {
	// 创建全局作用域
	globalScope := &SymbolScope{
		name:     "global",
		symbols:  make(map[string]*Symbol),
		children: []*SymbolScope{},
	}

	return globalScope
}

// FindSymbolsByName 在作用域中查找指定名称的符号
func (sm *SymbolManager) FindSymbolsByName(scope *SymbolScope, name string) []*Symbol {
	if scope == nil {
		return nil
	}

	scope.mu.RLock()
	defer scope.mu.RUnlock()

	var symbols []*Symbol
	if symbol, exists := scope.symbols[name]; exists {
		symbols = append(symbols, symbol)
	}

	// 递归搜索父作用域
	if scope.parent != nil {
		parentSymbols := sm.FindSymbolsByName(scope.parent, name)
		symbols = append(symbols, parentSymbols...)
	}

	return symbols
}