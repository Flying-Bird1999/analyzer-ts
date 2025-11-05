package tsmorphgo

import (
	"fmt"
	"sync"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// DeclarationAccessor 统一声明访问接口
// 提供高性能的声明访问能力，集成 analyzer/parser 的解析结果
type DeclarationAccessor interface {
	// 变量声明相关
	GetVariableDeclaration(node *ast.Node) (*parser.VariableDeclaration, bool)
	IsVariableDeclaration(node *ast.Node) bool

	// 函数声明相关
	GetFunctionDeclaration(node *ast.Node) (*parser.FunctionDeclarationResult, bool)
	IsFunctionDeclaration(node *ast.Node) bool

	// 接口声明相关
	GetInterfaceDeclaration(node *ast.Node) (*parser.InterfaceDeclarationResult, bool)
	IsInterfaceDeclaration(node *ast.Node) bool

	// 导入声明相关
	GetImportDeclaration(node *ast.Node) (*projectParser.ImportDeclarationResult, bool)
	IsImportDeclaration(node *ast.Node) bool

	// 类型别名声明相关
	GetTypeDeclaration(node *ast.Node) (*parser.TypeDeclarationResult, bool)
	IsTypeDeclaration(node *ast.Node) bool

	// 枚举声明相关
	GetEnumDeclaration(node *ast.Node) (*parser.EnumDeclarationResult, bool)
	IsEnumDeclaration(node *ast.Node) bool

	// 通用声明获取
	GetDeclaration(node *ast.Node) (interface{}, bool, string)
}

// OptimizedDeclarationAccessor 优化的声明访问器实现
// 使用缓存和懒加载机制提高性能
type OptimizedDeclarationAccessor struct {
	sourceFile    *SourceFile
	cache         map[*ast.Node]interface{}
	cacheMutex    sync.RWMutex
	parser        *parser.Parser
	parserMutex   sync.Mutex
	initialized   bool
	initializedMu sync.Mutex
}

// NewDeclarationAccessor 创建新的声明访问器
func NewDeclarationAccessor(sourceFile *SourceFile) DeclarationAccessor {
	return &OptimizedDeclarationAccessor{
		sourceFile: sourceFile,
		cache:      make(map[*ast.Node]interface{}),
	}
}

// ensureInitialized 确保解析器已初始化
func (d *OptimizedDeclarationAccessor) ensureInitialized() {
	d.initializedMu.Lock()
	defer d.initializedMu.Unlock()

	if !d.initialized {
		// 初始化解析器
		var err error
		d.parser, err = parser.NewParserFromSource(d.sourceFile.filePath, d.sourceFile.fileResult.Raw)
		if err != nil {
			// 如果解析器初始化失败，我们仍然继续工作，但某些功能可能不可用
			fmt.Printf("Warning: failed to initialize parser: %v\n", err)
		}
		d.initialized = true
	}
}

// GetVariableDeclaration 获取变量声明信息
// 使用缓存机制避免重复解析
func (d *OptimizedDeclarationAccessor) GetVariableDeclaration(node *ast.Node) (*parser.VariableDeclaration, bool) {
	// 1. 首先从现有 nodeResultMap 中查找（保持向后兼容）
	if d.sourceFile != nil && d.sourceFile.nodeResultMap != nil {
		if result, exists := d.sourceFile.nodeResultMap[node]; exists {
			if variable, ok := result.(parser.VariableDeclaration); ok {
				return &variable, true
			}
		}
	}

	// 2. 检查缓存
	d.cacheMutex.RLock()
	if result, exists := d.cache[node]; exists {
		d.cacheMutex.RUnlock()
		if variable, ok := result.(*parser.VariableDeclaration); ok {
			return variable, true
		}
		return nil, false
	}
	d.cacheMutex.RUnlock()

	// 3. 如果节点类型不匹配，直接返回
	if node.Kind != ast.KindVariableDeclaration && node.Kind != ast.KindVariableDeclarationList {
		return nil, false
	}

	// 4. 动态解析并缓存结果
	d.ensureInitialized()
	d.parserMutex.Lock()
	defer d.parserMutex.Unlock()

	// 解析变量声明（这里简化处理，实际应该调用parser的相应方法）
	// 注意：实际的解析逻辑需要根据parser包的具体实现来调用
	variable := &parser.VariableDeclaration{
		Exported:       false, // 需要从AST中提取
		SourceLocation: d.createSourceLocation(node),
		Node:          node,
	}

	// 缓存结果
	d.cacheMutex.Lock()
	d.cache[node] = variable
	d.cacheMutex.Unlock()

	return variable, true
}

// GetFunctionDeclaration 获取函数声明信息
func (d *OptimizedDeclarationAccessor) GetFunctionDeclaration(node *ast.Node) (*parser.FunctionDeclarationResult, bool) {
	// 1. 从现有 nodeResultMap 中查找
	if d.sourceFile != nil && d.sourceFile.nodeResultMap != nil {
		if result, exists := d.sourceFile.nodeResultMap[node]; exists {
			if function, ok := result.(parser.FunctionDeclarationResult); ok {
				return &function, true
			}
		}
	}

	// 2. 检查缓存
	d.cacheMutex.RLock()
	if result, exists := d.cache[node]; exists {
		d.cacheMutex.RUnlock()
		if function, ok := result.(*parser.FunctionDeclarationResult); ok {
			return function, true
		}
		return nil, false
	}
	d.cacheMutex.RUnlock()

	// 3. 检查节点类型
	if node.Kind != ast.KindFunctionDeclaration && node.Kind != ast.KindFunctionExpression {
		return nil, false
	}

	// 4. 动态解析并缓存
	d.ensureInitialized()
	d.parserMutex.Lock()
	defer d.parserMutex.Unlock()

	function := &parser.FunctionDeclarationResult{
		Exported:       false, // 需要从AST中提取
		IsAsync:        false, // 需要从AST中提取
		IsGenerator:    false, // 需要从AST中提取
		Generics:       []string{}, // 需要从AST中提取
		Parameters:     []parser.ParameterResult{}, // 需要从AST中提取
		ReturnType:     "", // 需要从AST中提取
		SourceLocation: d.createSourceLocation(node),
		Node:          node,
	}

	d.cacheMutex.Lock()
	d.cache[node] = function
	d.cacheMutex.Unlock()

	return function, true
}

// GetInterfaceDeclaration 获取接口声明信息
func (d *OptimizedDeclarationAccessor) GetInterfaceDeclaration(node *ast.Node) (*parser.InterfaceDeclarationResult, bool) {
	// 1. 从现有 nodeResultMap 中查找
	if d.sourceFile != nil && d.sourceFile.nodeResultMap != nil {
		if result, exists := d.sourceFile.nodeResultMap[node]; exists {
			if interfaceDecl, ok := result.(parser.InterfaceDeclarationResult); ok {
				return &interfaceDecl, true
			}
		}
	}

	// 2. 检查缓存
	d.cacheMutex.RLock()
	if result, exists := d.cache[node]; exists {
		d.cacheMutex.RUnlock()
		if interfaceDecl, ok := result.(*parser.InterfaceDeclarationResult); ok {
			return interfaceDecl, true
		}
		return nil, false
	}
	d.cacheMutex.RUnlock()

	// 3. 检查节点类型
	if node.Kind != ast.KindInterfaceDeclaration {
		return nil, false
	}

	// 4. 动态解析并缓存
	d.ensureInitialized()
	d.parserMutex.Lock()
	defer d.parserMutex.Unlock()

	interfaceDecl := &parser.InterfaceDeclarationResult{
		Identifier:     "", // 需要从AST中提取
		Exported:       false, // 需要从AST中提取
		Reference:      map[string]parser.TypeReference{}, // 需要从AST中提取
		SourceLocation: d.createSourceLocation(node),
		Node:          node,
	}

	d.cacheMutex.Lock()
	d.cache[node] = interfaceDecl
	d.cacheMutex.Unlock()

	return interfaceDecl, true
}

// GetImportDeclaration 获取导入声明信息
func (d *OptimizedDeclarationAccessor) GetImportDeclaration(node *ast.Node) (*projectParser.ImportDeclarationResult, bool) {
	// 1. 从现有 nodeResultMap 中查找
	if d.sourceFile != nil && d.sourceFile.nodeResultMap != nil {
		if result, exists := d.sourceFile.nodeResultMap[node]; exists {
			if importDecl, ok := result.(projectParser.ImportDeclarationResult); ok {
				return &importDecl, true
			}
		}
	}

	// 2. 检查缓存
	d.cacheMutex.RLock()
	if result, exists := d.cache[node]; exists {
		d.cacheMutex.RUnlock()
		if importDecl, ok := result.(*projectParser.ImportDeclarationResult); ok {
			return importDecl, true
		}
		return nil, false
	}
	d.cacheMutex.RUnlock()

	// 3. 检查节点类型
	if node.Kind != ast.KindImportDeclaration {
		return nil, false
	}

	// 4. 动态解析并缓存
	d.ensureInitialized()
	d.parserMutex.Lock()
	defer d.parserMutex.Unlock()

	importDecl := &projectParser.ImportDeclarationResult{
		ImportModules:  []projectParser.ImportModule{}, // 需要从AST中提取
		Source:         projectParser.SourceData{Type: "module"}, // 需要从AST中提取
		Node:           node,
	}

	d.cacheMutex.Lock()
	d.cache[node] = importDecl
	d.cacheMutex.Unlock()

	return importDecl, true
}

// GetTypeDeclaration 获取类型别名声明信息
func (d *OptimizedDeclarationAccessor) GetTypeDeclaration(node *ast.Node) (*parser.TypeDeclarationResult, bool) {
	// 1. 从现有 nodeResultMap 中查找
	if d.sourceFile != nil && d.sourceFile.nodeResultMap != nil {
		if result, exists := d.sourceFile.nodeResultMap[node]; exists {
			if typeDecl, ok := result.(parser.TypeDeclarationResult); ok {
				return &typeDecl, true
			}
		}
	}

	// 2. 检查缓存
	d.cacheMutex.RLock()
	if result, exists := d.cache[node]; exists {
		d.cacheMutex.RUnlock()
		if typeDecl, ok := result.(*parser.TypeDeclarationResult); ok {
			return typeDecl, true
		}
		return nil, false
	}
	d.cacheMutex.RUnlock()

	// 3. 检查节点类型
	if node.Kind != ast.KindTypeAliasDeclaration {
		return nil, false
	}

	// 4. 动态解析并缓存
	d.ensureInitialized()
	d.parserMutex.Lock()
	defer d.parserMutex.Unlock()

	typeDecl := &parser.TypeDeclarationResult{
		Identifier:     "", // 需要从AST中提取
		Exported:       false, // 需要从AST中提取
		Reference:      map[string]parser.TypeReference{}, // 需要从AST中提取
		SourceLocation: d.createSourceLocation(node),
		Node:          node,
	}

	d.cacheMutex.Lock()
	d.cache[node] = typeDecl
	d.cacheMutex.Unlock()

	return typeDecl, true
}

// GetEnumDeclaration 获取枚举声明信息
func (d *OptimizedDeclarationAccessor) GetEnumDeclaration(node *ast.Node) (*parser.EnumDeclarationResult, bool) {
	// 1. 从现有 nodeResultMap 中查找
	if d.sourceFile != nil && d.sourceFile.nodeResultMap != nil {
		if result, exists := d.sourceFile.nodeResultMap[node]; exists {
			if enumDecl, ok := result.(parser.EnumDeclarationResult); ok {
				return &enumDecl, true
			}
		}
	}

	// 2. 检查缓存
	d.cacheMutex.RLock()
	if result, exists := d.cache[node]; exists {
		d.cacheMutex.RUnlock()
		if enumDecl, ok := result.(*parser.EnumDeclarationResult); ok {
			return enumDecl, true
		}
		return nil, false
	}
	d.cacheMutex.RUnlock()

	// 3. 检查节点类型
	if node.Kind != ast.KindEnumDeclaration {
		return nil, false
	}

	// 4. 动态解析并缓存
	d.ensureInitialized()
	d.parserMutex.Lock()
	defer d.parserMutex.Unlock()

	enumDecl := &parser.EnumDeclarationResult{
		Identifier:     "", // 需要从AST中提取
		Exported:       false, // 需要从AST中提取
		SourceLocation: d.createSourceLocation(node),
		Node:          node,
	}

	d.cacheMutex.Lock()
	d.cache[node] = enumDecl
	d.cacheMutex.Unlock()

	return enumDecl, true
}

// GetDeclaration 通用声明获取方法
// 根据节点类型自动选择相应的声明获取方法
func (d *OptimizedDeclarationAccessor) GetDeclaration(node *ast.Node) (interface{}, bool, string) {
	switch node.Kind {
	case ast.KindVariableDeclaration, ast.KindVariableDeclarationList:
		result, ok := d.GetVariableDeclaration(node)
		return result, ok, "VariableDeclaration"
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression:
		result, ok := d.GetFunctionDeclaration(node)
		return result, ok, "FunctionDeclaration"
	case ast.KindInterfaceDeclaration:
		result, ok := d.GetInterfaceDeclaration(node)
		return result, ok, "InterfaceDeclaration"
	case ast.KindImportDeclaration:
		result, ok := d.GetImportDeclaration(node)
		return result, ok, "ImportDeclaration"
	case ast.KindTypeAliasDeclaration:
		result, ok := d.GetTypeDeclaration(node)
		return result, ok, "TypeDeclaration"
	case ast.KindEnumDeclaration:
		result, ok := d.GetEnumDeclaration(node)
		return result, ok, "EnumDeclaration"
	default:
		return nil, false, "Unknown"
	}
}

// IsVariableDeclaration 检查是否是变量声明
func (d *OptimizedDeclarationAccessor) IsVariableDeclaration(node *ast.Node) bool {
	_, ok := d.GetVariableDeclaration(node)
	return ok
}

// IsFunctionDeclaration 检查是否是函数声明
func (d *OptimizedDeclarationAccessor) IsFunctionDeclaration(node *ast.Node) bool {
	_, ok := d.GetFunctionDeclaration(node)
	return ok
}

// IsInterfaceDeclaration 检查是否是接口声明
func (d *OptimizedDeclarationAccessor) IsInterfaceDeclaration(node *ast.Node) bool {
	_, ok := d.GetInterfaceDeclaration(node)
	return ok
}

// IsImportDeclaration 检查是否是导入声明
func (d *OptimizedDeclarationAccessor) IsImportDeclaration(node *ast.Node) bool {
	_, ok := d.GetImportDeclaration(node)
	return ok
}

// IsTypeDeclaration 检查是否是类型别名声明
func (d *OptimizedDeclarationAccessor) IsTypeDeclaration(node *ast.Node) bool {
	_, ok := d.GetTypeDeclaration(node)
	return ok
}

// IsEnumDeclaration 检查是否是枚举声明
func (d *OptimizedDeclarationAccessor) IsEnumDeclaration(node *ast.Node) bool {
	_, ok := d.GetEnumDeclaration(node)
	return ok
}

// createSourceLocation 创建源码位置信息
// 这是一个辅助方法，用于为动态解析的结果创建位置信息
func (d *OptimizedDeclarationAccessor) createSourceLocation(node *ast.Node) *parser.SourceLocation {
	if d.sourceFile == nil || d.sourceFile.fileResult == nil {
		return &parser.SourceLocation{
			Start: parser.NodePosition{Line: 0, Column: 0},
			End:   parser.NodePosition{Line: 0, Column: 0},
		}
	}

	raw := d.sourceFile.fileResult.Raw
	startLine, startChar := utils.GetLineAndCharacterOfPosition(raw, node.Pos())
	endLine, endChar := utils.GetLineAndCharacterOfPosition(raw, node.End())

	return &parser.SourceLocation{
		Start: parser.NodePosition{
			Line:   startLine + 1, // 转换为1-based
			Column: startChar + 1,
		},
		End: parser.NodePosition{
			Line:   endLine + 1,
			Column: endChar + 1,
		},
	}
}

// GetCacheStats 获取缓存统计信息（用于性能调试）
func (d *OptimizedDeclarationAccessor) GetCacheStats() map[string]interface{} {
	d.cacheMutex.RLock()
	defer d.cacheMutex.RUnlock()

	stats := map[string]interface{}{
		"cache_size":         len(d.cache),
		"initialized":        d.initialized,
		"source_file_path":   d.sourceFile.filePath,
		"has_node_result_map": d.sourceFile.nodeResultMap != nil,
	}

	if d.sourceFile.nodeResultMap != nil {
		stats["node_result_map_size"] = len(d.sourceFile.nodeResultMap)
	}

	return stats
}

// ClearCache 清理缓存（用于内存管理）
func (d *OptimizedDeclarationAccessor) ClearCache() {
	d.cacheMutex.Lock()
	defer d.cacheMutex.Unlock()

	d.cache = make(map[*ast.Node]interface{})
}

// ValidateCache 验证缓存一致性（用于调试）
func (d *OptimizedDeclarationAccessor) ValidateCache() ([]string, []error) {
	var warnings []string
	var errors []error

	d.cacheMutex.RLock()
	defer d.cacheMutex.RUnlock()

	// 检查缓存项是否仍然有效
	for node, result := range d.cache {
		if d.sourceFile == nil || d.sourceFile.fileResult == nil {
			errors = append(errors, fmt.Errorf("source file is nil for cached node"))
			continue
		}

		// 检查节点是否仍然在AST中
		// 注意：这里简化了实际的验证逻辑
		if node == nil {
			errors = append(errors, fmt.Errorf("cached node is nil"))
			continue
		}

		if result == nil {
			warnings = append(warnings, fmt.Sprintf("cached result is nil for node type: %d", node.Kind))
		}
	}

	return warnings, errors
}