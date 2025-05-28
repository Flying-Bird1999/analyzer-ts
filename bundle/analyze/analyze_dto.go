package analyze

import "main/bundle/parser"

type FileAnalyzeResult struct {
	ImportDeclarations    []ImportDeclarationResult
	InterfaceDeclarations map[string]parser.InterfaceDeclarationResult
	TypeDeclarations      map[string]parser.TypeDeclarationResult
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
