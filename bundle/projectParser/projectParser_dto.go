package projectParser

import "main/bundle/parser"

type JsFileParserResult struct {
	ImportDeclarations    []ImportDeclarationResult
	InterfaceDeclarations map[string]parser.InterfaceDeclarationResult
	TypeDeclarations      map[string]parser.TypeDeclarationResult
}

type PackageJsonFileParserResult struct {
	Workspace string             // 如果是 monorepo 项目，则表示所在的 workspace, 最外层或非monorepo项目否则为 "root"
	Path      string             // package.json 的路径
	Namespace string             // 包名的命名空间，例如 @sl/sc-product
	Version   string             // 包的版本号
	NpmList   map[string]NpmItem // npm列表，key为包名
}

type NpmItem struct {
	Name              string // 包名
	Type              string // 包类型: "devDependencies"、“peerDependencies”、“dependencies”
	Version           string // 包版本号
	NodeModuleVersion string // 项目真实安装的包版本
}
type ImportDeclarationResult struct {
	ImportModules []ImportModule
	Raw           string
	Source        SourceData
}

type ImportModule struct {
	ImportModule string // 模块名, 对应实际导出的内容模块
	Type         string // 默认导入: default、命名空间导入: namespace、命名导入:named、unknown
	Identifier   string //
}

type SourceData struct {
	FilePath string // 绝对路径
	NpmPkg   string // npm 包名，如果是 npm 包则有值
	Type     string // file | npm | unknown
}
