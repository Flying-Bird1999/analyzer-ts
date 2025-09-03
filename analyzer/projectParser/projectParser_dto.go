package projectParser

import "github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"

// JsFileParserResult 结构体用于存储对单个JS或TS文件进行解析后得到的核心数据。
// 这些数据主要包括文件中的导入和导出声明，为项目级别的依赖分析和代码理解提供基础。
type JsFileParserResult struct {
	// ImportDeclarations 存储了文件中所有的导入声明。
	// 每个导入声明都包含了导入的模块、原始语句以及解析后的来源信息。
	ImportDeclarations []ImportDeclarationResult `json:"importDeclarations"`
	// ExportDeclarations 存储了文件中所有的导出声明。
	// 每个导出声明都包含了导出的模块、原始语句以及可能的来源信息（用于再导出场景）。
	ExportDeclarations    []ExportDeclarationResult                    `json:"exportDeclarations"`
	ExportAssignments     []parser.ExportAssignmentResult              `json:"exportAssignments"`     // 例如 `export default` 声明
	InterfaceDeclarations map[string]parser.InterfaceDeclarationResult `json:"interfaceDeclarations"` // 文件中定义的接口
	TypeDeclarations      map[string]parser.TypeDeclarationResult      `json:"typeDeclarations"`      // 文件中定义的类型别名
	EnumDeclarations      map[string]parser.EnumDeclarationResult      `json:"enumDeclarations"`      // 文件中定义的枚举
	VariableDeclarations  []parser.VariableDeclaration                 `json:"variableDeclarations"`  // 文件中声明的变量
	CallExpressions       []parser.CallExpression                      `json:"callExpressions"`       // 文件中的函数调用表达式
	JsxElements           []JSXElementResult                           `json:"jsxElements"`           // 文件中的JSX元素
	FunctionDeclarations  []parser.FunctionDeclarationResult           `json:"functionsDeclarations"` // 文件中所有函数声明的信息
	ExtractedNodes        parser.ExtractedNodes                        `json:"extractedNodes"`        // 用于存储提取的节点信息
	Errors                []error                                      `json:"errors,omitempty"`      // 新增：用于存储解析过程中遇到的错误
}

// PackageJsonFileParserResult 存储了对 `package.json` 文件解析后的关键信息。
// 这对于理解项目的基本配置、依赖关系和在 monorepo 中的角色至关重要。
type PackageJsonFileParserResult struct {
	// Workspace 表示该 `package.json` 所属的工作区。
	// 在 monorepo 项目中，这通常是子包的目录名。对于根目录或非 monorepo 项目，其值为 "root"。
	Workspace string `json:"workspace"`
	// Path 记录了 `package.json` 文件在文件系统中的绝对路径。
	Path string `json:"path"`
	// Namespace 是包的名称，通常定义在 `package.json` 的 "name" 字段中，例如 "@scope/my-package"。
	Namespace string `json:"namespace"`
	// Version 是包的版本号，定义在 `package.json` 的 "version" 字段中。
	Version string `json:"version"`
	// NpmList 是一个映射，存储了项目的所有NPM依赖（包括 dependencies, devDependencies, peerDependencies）。
	// 键是NPM包的名称，值是包含该包详细信息的 NpmItem 结构体。
	NpmList map[string]NpmItem `json:"npmList"`
}

// NpmItem 包含了单个NPM依赖包的详细信息。
type NpmItem struct {
	// Name 是NPM包的名称，例如 "react"。
	Name string `json:"name"`
	// Type 表示该依赖的类型，例如 "devDependencies", "peerDependencies", 或 "dependencies"。
	Type string `json:"type"`
	// Version 是在 `package.json` 中声明的版本范围，例如 "^18.2.0"。
	Version string `json:"version"`
	// NodeModuleVersion 是在 `node_modules` 目录中实际安装的该包的版本。
	// 这对于解决版本冲突和理解实际使用的依赖版本非常有用。
	NodeModuleVersion string `json:"nodeModuleVersion"`
}

// ImportDeclarationResult 存储了单个导入声明（`import ... from ...`）的完整解析结果。
type ImportDeclarationResult struct {
	// ImportModules 是一个切片，包含了该导入语句中所有被导入的模块。
	// 例如，`import { a, b as c } from './mod'` 会产生两个 ImportModule。
	ImportModules []ImportModule `json:"importModules"`
	// Raw 存储了该导入声明在源代码中的原始、未修改的文本。
	Raw string `json:"raw"`
	// Source 包含了对导入来源模块的解析结果，包括其绝对路径、类型（文件或NPM包）等。
	Source SourceData `json:"source"`
}

// ImportModule 代表一个被导入的独立实体。
type ImportModule struct {
	// ImportModule 是导入的模块的原始名称。
	// - 对于命名导入 `import { a as b }`，它是 `a`。
	// - 对于默认导入 `import a from ...`，它是 "default"。
	// - 对于命名空间导入 `import * as ns from ...`，它是 `ns`。
	ImportModule string `json:"importModule"`
	// Type 表示导入的类型，可以是 "default"（默认导入）, "namespace"（命名空间导入）, 或 "named"（命名导入）。
	Type string `json:"type"`
	// Identifier 是该模块在当前文件中使用的标识符（本地名称）。
	// - 对于 `import { a as b }`，它是 `b`。
	// - 对于 `import a from ...`，它是 `a`。
	Identifier string `json:"identifier"`
}

// ExportDeclarationResult 存储了单个导出声明（`export ...`）的完整解析结果。
type ExportDeclarationResult struct {
	// ExportModules 是一个切片，包含了该导出语句中所有被导出的模块。
	ExportModules []ExportModule `json:"exportModules"`
	// Raw 存储了该导出声明在源代码中的原始、未修改的文本。
	Raw string `json:"raw"`
	// Source 在 "re-export"（再导出）场景下（例如 `export { a } from './mod'`）不为 nil。
	// 它包含了对来源模块的解析结果。对于常规的命名导出，此字段为 nil。
	Source *SourceData `json:"source,omitempty"`
}

// ExportModule 代表一个被导出的独立实体。
type ExportModule struct {
	// ModuleName 是导出的模块的原始名称。
	// - 对于命名导出 `export { a as b }`，它是 `a`。
	// - 对于命名空间导入 `export * as ns from ...`，它是 `*`。
	ModuleName string `json:"moduleName"`
	// Type 表示导出的类型，可以是 "named"（命名导出）或 "namespace"（命名空间导出）。
	Type string `json:"type"`
	// Identifier 是导出的标识符（外部名称）。
	// - 对于 `export { a as b }`，它是 `b`。
	Identifier string `json:"identifier"`
}

// JSXElementResult 存储了单个JSX元素的解析结果，包括其来源信息。
type JSXElementResult struct {
	ComponentChain []string              `json:"componentChain"` // 组件的完整路径
	Attrs          []parser.JSXAttribute `json:"attrs"`          // JSX 属性
	Raw            string                `json:"raw"`            // 节点在源码中的原始文本
	Source         SourceData            `json:"source"`         // 解析后的来源信息
}

// SourceData 结构体用于存储对导入或再导出来源模块路径的解析结果。
type SourceData struct {
	// FilePath 是解析后的模块的绝对文件路径。如果来源是NPM包，则此字段为空。
	FilePath string `json:"filePath"`
	// NpmPkg 是NPM包的名称。如果来源是本地文件，则此字段为空。
	NpmPkg string `json:"npmPkg"`
	// Type 表示来源的类型，可以是 "file"（本地文件）, "npm"（NPM包）, 或 "unknown"（未知）。
	Type string `json:"type"`
}

// TsConfig holds the parsed information from a tsconfig.json file,
// including path aliases and the base URL for module resolution.
type TsConfig struct {
	Alias   map[string]string `json:"alias"`
	BaseUrl string            `json:"baseUrl"`
}
