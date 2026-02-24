// Package export_call 符号解析器
package export_call

import (
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
)

// SymbolResolver 符号解析器
// 用于解析 export { foo } 和 export default foo 中的符号，找到真实定义
type SymbolResolver struct {
	jsData map[string]projectParser.JsFileParserResult
}

// NewSymbolResolver 创建符号解析器
func NewSymbolResolver(jsData map[string]projectParser.JsFileParserResult) *SymbolResolver {
	return &SymbolResolver{
		jsData: jsData,
	}
}

// ResolveExportDeclaration 解析 export { foo, bar } 中的符号
// 返回每个符号对应的真实节点类型
func (r *SymbolResolver) ResolveExportDeclaration(
	fileData *projectParser.JsFileParserResult,
	exportDecl *projectParser.ExportDeclarationResult,
) map[string]NodeType {
	result := make(map[string]NodeType)

	for _, module := range exportDecl.ExportModules {
		symbolName := module.ModuleName

		// 在同一文件中查找符号的定义
		if nodeType := r.findSymbolDefinition(fileData, symbolName); nodeType != "" {
			result[symbolName] = nodeType
		} else {
			result[symbolName] = NodeTypeVariable // 默认作为 variable
		}
	}

	return result
}

// ResolveExportAssignment 解析 export default foo 中的符号
// 返回 default 导出的真实节点类型
func (r *SymbolResolver) ResolveExportAssignment(
	fileData *projectParser.JsFileParserResult,
	exportAssign *parser.ExportAssignmentResult,
) NodeType {
	// 直接使用 parser 预先提取的 Name
	symbolName := exportAssign.Name

	if nodeType := r.findSymbolDefinition(fileData, symbolName); nodeType != "" {
		return nodeType
	}

	// 根据表达式特征推断
	expr := exportAssign.Expression
	if strings.Contains(expr, "function") {
		return NodeTypeFunction
	}
	if strings.Contains(expr, "class") {
		return NodeTypeFunction
	}
	return NodeTypeVariable
}

// findSymbolDefinition 在文件中查找符号定义
func (r *SymbolResolver) findSymbolDefinition(
	fileData *projectParser.JsFileParserResult,
	symbolName string,
) NodeType {
	// 1. 查找函数声明
	for _, fn := range fileData.FunctionDeclarations {
		if fn.Identifier == symbolName {
			return NodeTypeFunction
		}
	}

	// 2. 查找变量声明
	for _, v := range fileData.VariableDeclarations {
		for _, decl := range v.Declarators {
			if decl.Identifier == symbolName {
				return NodeTypeVariable
			}
		}
	}

	// 3. 查找类型声明
	if _, ok := fileData.TypeDeclarations[symbolName]; ok {
		return NodeTypeType
	}

	// 4. 查找接口声明
	if _, ok := fileData.InterfaceDeclarations[symbolName]; ok {
		return NodeTypeInterface
	}

	// 5. 查找枚举声明
	if _, ok := fileData.EnumDeclarations[symbolName]; ok {
		return NodeTypeEnum
	}

	return ""
}
