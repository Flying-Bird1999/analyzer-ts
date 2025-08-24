// package project_analyzer 是整个分析器插件的核心包。
package project_analyzer

import (
	"main/analyzer/parser"
	"main/analyzer/projectParser"
)

// --- 以下是为 `analyze` 命令服务的、经过滤和简化的数据结构 ---
// 这些结构体的目的是为了在 `analyze` 命令的JSON输出中，
// 只保留最关键的信息，去除不必要的原始数据，以减小输出文件的大小。

// FilteredInterfaceDeclarationResult 代表一个简化的接口声明信息。
type FilteredInterfaceDeclarationResult struct {
	Identifier string                          `json:"identifier"`
	Reference  map[string]parser.TypeReference `json:"reference"`
}

// FilteredTypeDeclarationResult 代表一个简化的类型定义信息。
type FilteredTypeDeclarationResult struct {
	Identifier string                          `json:"identifier"`
	Reference  map[string]parser.TypeReference `json:"reference"`
}

// FilteredEnumDeclarationResult 代表一个简化的枚举声明信息。
type FilteredEnumDeclarationResult struct {
	Identifier string `json:"identifier"`
}

// FilteredVariableDeclaration 代表一个简化的变量声明信息。
type FilteredVariableDeclaration struct {
	Exported    bool                         `json:"exported"`
	Kind        parser.DeclarationKind       `json:"kind"`
	Source      *parser.VariableValue        `json:"source,omitempty"`
	Declarators []*parser.VariableDeclarator `json:"declarators"`
}

// FilteredCallExpression 代表一个简化的函数调用表达式信息。
type FilteredCallExpression struct {
	CallChain []string          `json:"callChain"`
	Arguments []parser.Argument `json:"arguments"`
	Type      string            `json:"type"`
}

// FilteredJSXElement 代表一个简化的JSX元素信息。
type FilteredJSXElement struct {
	ComponentChain []string              `json:"componentChain"`
	Attrs          []parser.JSXAttribute `json:"attrs"`
}

// FilteredExportAssignmentResult 代表一个简化的 `export =` 表达式信息。
type FilteredExportAssignmentResult struct {
	Expression string `json:"expression"`
}

// FilteredImportDeclaration 代表一个简化的导入声明信息。
type FilteredImportDeclaration struct {
	ImportModules []parser.ImportModule    `json:"importModules"`
	Source        projectParser.SourceData `json:"source"`
}

// FilteredExportDeclaration 代表一个简化的导出声明信息。
type FilteredExportDeclaration struct {
	ExportModules []parser.ExportModule     `json:"exportModules"`
	Source        *projectParser.SourceData `json:"source,omitempty"`
}

type FilteredFunctionDeclarationResult struct {
	Exported bool `json:"exported"` // 标记此函数是否被导出。

	Identifier string                   `json:"identifier"` // 函数的名称。对于匿名函数可能为空。
	IsAsync    bool                     `json:"isAsync"`    // 标记此函数是否为异步函数 (async)。
	Parameters []parser.ParameterResult `json:"parameters"` // 函数的参数列表。
	ReturnType string                   `json:"returnType"` // 函数的返回类型文本。
}

// FilteredJsFileParserResult 代表一个被简化和过滤后的单个文件的解析结果。
type FilteredJsFileParserResult struct {
	ImportDeclarations   []FilteredImportDeclaration         `json:"importDeclarations"`
	ExportDeclarations   []FilteredExportDeclaration         `json:"exportDeclarations"`
	ExportAssignments    []FilteredExportAssignmentResult    `json:"exportAssignments"`
	VariableDeclarations []FilteredVariableDeclaration       `json:"variableDeclarations"`
	CallExpressions      []FilteredCallExpression            `json:"callExpressions"`
	JsxElements          []FilteredJSXElement                `json:"jsxElements"`
	FunctionDeclarations []FilteredFunctionDeclarationResult `json:"functionDeclarations"`
	AnyDeclarations      []parser.AnyInfo                    `json:"anyDeclarations"`
}

// FilteredProjectParserResult 代表整个项目被简化和过滤后的解析结果。
type FilteredProjectParserResult struct {
	Js_Data      map[string]FilteredJsFileParserResult                `json:"js_Data"`
	Package_Data map[string]projectParser.PackageJsonFileParserResult `json:"package_Data"`
}
