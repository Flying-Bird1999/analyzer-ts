package analyze

type FileAnalyzeResult struct {
	ImportDeclarations    []ImportDeclarationResult
	InterfaceDeclarations map[string]InterfaceDeclarationResult
	TypeDeclarations      map[string]TypeDeclarationResult
}

type ImportDeclarationResult struct {
	Modules []Module
	Raw     string
	Source  string
}

type Module struct {
	Module     string // 模块名, 对应实际导出的内容模块
	Type       string // 默认导入: default、命名空间导入: namespace、命名导入:named、unknown
	Identifier string //
}

type TypeDeclarationResult struct {
	Name      string // 名称
	Raw       string // 源码
	Reference map[string]TypeReference
}

type InterfaceDeclarationResult struct {
	Name      string // 名称
	Raw       string // 源码
	Reference map[string]TypeReference
}

type TypeReference struct {
	Name     string
	Location []string // 保留设计，类型的位置，用.隔开引用的位置，例如：School.student.name
	IsExtend bool     // 是否继承，true表示继承，false表示member中引用的
}
