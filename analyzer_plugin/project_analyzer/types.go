package project_analyzer

import (
	"main/analyzer/parser"
	"main/analyzer/projectParser"
)

// DependencyCheckResult 是依赖检查功能最终输出的完整结果结构体。
// 它整合了隐式依赖、未使用依赖和过期依赖三项检查的结果。
type DependencyCheckResult struct {
	ImplicitDependencies []ImplicitDependency `json:"implicitDependencies"`
	UnusedDependencies   []UnusedDependency   `json:"unusedDependencies"`
	OutdatedDependencies []OutdatedDependency `json:"outdatedDependencies"`
}

// ImplicitDependency 代表一个隐式依赖（幽灵依赖）。
// 即在代码中被使用，但未在 package.json 中声明的包。
type ImplicitDependency struct {
	Name     string `json:"name"`     // 依赖包的名称
	FilePath string `json:"filePath"` // 在哪个文件中发现了该依赖
	Raw      string `json:"raw"`      // 发现该依赖的原始导入语句
}

// UnusedDependency 代表一个未使用的依赖。
// 即在 package.json 中声明了，但代码中并未导入的包。
type UnusedDependency struct {
	Name            string `json:"name"`            // 依赖包的名称
	Version         string `json:"version"`         // 在 package.json 中声明的版本
	PackageJsonPath string `json:"packageJsonPath"` // 该依赖所在的 package.json 文件路径
}

// OutdatedDependency 代表一个已过期的依赖。
// 即在 package.json 中声明的版本落后于 NPM Registry 中的最新版本。
type OutdatedDependency struct {
	Name            string `json:"name"`            // 依赖包的名称
	CurrentVersion  string `json:"currentVersion"`  // 当前在 package.json 中声明的版本
	LatestVersion   string `json:"latestVersion"`   // NPM Registry 中的最新版本
	PackageJsonPath string `json:"packageJsonPath"` // 该依赖所在的 package.json 文件路径
}

// packageInfo 是用于解析从 NPM Registry API 返回的 JSON 数据的结构体。
// 我们主要关心 `dist-tags.latest` 字段来获取最新版本号。
type packageInfo struct {
	DistTags struct {
		Latest string `json:"latest"`
	} `json:"dist-tags"`
}

// --- 以下是为旧 `analyze` 命令服务的过滤后的数据结构 ---

type FilteredInterfaceDeclarationResult struct {
	Identifier string                          `json:"identifier"`
	Reference  map[string]parser.TypeReference `json:"reference"`
}

type FilteredTypeDeclarationResult struct {
	Identifier string                          `json:"identifier"`
	Reference  map[string]parser.TypeReference `json:"reference"`
}

type FilteredEnumDeclarationResult struct {
	Identifier string `json:"identifier"`
}

type FilteredVariableDeclaration struct {
	Exported    bool                         `json:"exported"`
	Kind        parser.DeclarationKind       `json:"kind"`
	Source      *parser.VariableValue        `json:"source,omitempty"`
	Declarators []*parser.VariableDeclarator `json:"declarators"`
}

type FilteredCallExpression struct {
	CallChain []string          `json:"callChain"`
	Arguments []parser.Argument `json:"arguments"`
	Type      string            `json:"type"`
}

type FilteredJSXElement struct {
	ComponentChain []string              `json:"componentChain"`
	Attrs          []parser.JSXAttribute `json:"attrs"`
}

type FilteredExportAssignmentResult struct {
	Expression string `json:"expression"`
}

type FilteredImportDeclaration struct {
	ImportModules []parser.ImportModule    `json:"importModules"`
	Source        projectParser.SourceData `json:"source"`
}

type FilteredExportDeclaration struct {
	ExportModules []parser.ExportModule     `json:"exportModules"`
	Source        *projectParser.SourceData `json:"source,omitempty"`
}

type FilteredJsFileParserResult struct {
	ImportDeclarations []FilteredImportDeclaration      `json:"importDeclarations"`
	ExportDeclarations []FilteredExportDeclaration      `json:"exportDeclarations"`
	ExportAssignments  []FilteredExportAssignmentResult `json:"exportAssignments"`
	VariableDeclarations []FilteredVariableDeclaration `json:"variableDeclarations"`
	CallExpressions      []FilteredCallExpression      `json:"callExpressions"`
	JsxElements          []FilteredJSXElement          `json:"jsxElements"`
}

type FilteredProjectParserResult struct {
	Js_Data      map[string]FilteredJsFileParserResult                `json:"js_Data"`
	Package_Data map[string]projectParser.PackageJsonFileParserResult `json:"package_Data"`
}
