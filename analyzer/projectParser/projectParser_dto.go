package projectParser

import "main/analyzer/parser"

// JsFileParserResult JS 文件解析结果
type JsFileParserResult struct {
	ImportDeclarations    []ImportDeclarationResult                    `json:"importDeclarations"` // 导入声明
	ExportDeclarations    []parser.ExportDeclarationResult             // 导出声明
	ExportAssignments     []parser.ExportAssignmentResult              // `export default` 声明
	InterfaceDeclarations map[string]parser.InterfaceDeclarationResult `json:"interfaceDeclarations"` // 接口声明
	TypeDeclarations      map[string]parser.TypeDeclarationResult      `json:"typeDeclarations"`      // 类型声明
	EnumDeclarations      map[string]parser.EnumDeclarationResult      `json:"enumDeclarations"`      // 枚举声明
	VariableDeclarations  []parser.VariableDeclaration                 `json:"variableDeclarations"`  // 变量声明
	CallExpressions       []parser.CallExpression                      `json:"callExpressions"`       // 函数调用
	JsxElements           []parser.JSXElement                          `json:"jsxElements"`           // JSX 元素
}

// PackageJsonFileParserResult package.json 文件解析结果
type PackageJsonFileParserResult struct {
	Workspace string             `json:"workspace"` // 如果是 monorepo 项目，则表示所在的 workspace, 最外层或非monorepo项目否则为 "root"
	Path      string             `json:"path"`      // package.json 的路径
	Namespace string             `json:"namespace"` // 包名的命名空间，例如 @sl/sc-product
	Version   string             `json:"version"`   // 包的版本号
	NpmList   map[string]NpmItem `json:"npmList"`   // npm列表，key为包名
}

// NpmItem npm 包信息
type NpmItem struct {
	Name              string `json:"name"`              // 包名
	Type              string `json:"type"`              // 包类型: "devDependencies"、“peerDependencies”、“dependencies”
	Version           string `json:"version"`           // 包版本号
	NodeModuleVersion string `json:"nodeModuleVersion"` // 项目真实安装的包版本
}

// ImportDeclarationResult 导入声明结果
type ImportDeclarationResult struct {
	ImportModules []ImportModule `json:"importModules"` // 导入的模块
	Raw           string         `json:"raw"`           // 原始导入语句
	Source        SourceData     `json:"source"`        // 导入来源
}

// ImportModule 导入的模块
type ImportModule struct {
	ImportModule string `json:"importModule"` // 模块名, 对应实际导出的内容模块
	Type         string `json:"type"`         // 默认导入: default、命名空间导入: namespace、命名导入:named、unknown
	Identifier   string `json:"identifier"`   // 标识符
}

// SourceData 导入来源
type SourceData struct {
	FilePath string `json:"filePath"` // 绝对路径
	NpmPkg   string `json:"npmPkg"`   // npm 包名，如果是 npm 包则有值
	Type     string `json:"type"`     // file | npm | unknown
}
