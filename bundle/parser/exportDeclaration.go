package parser

import (
	"main/bundle/utils"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// 解析导出模块
// - 默认导出 default: export default Bird;
// - 命名导出 named:
// 		- export { School, School2 as NewSchool2 };
// 		- export type { CurrentRes };
//  	- export const name = "bird"
//  	- export function name() {}

type ExportModule struct {
	ImportModule string // 模块名, 对应实际导出的内容模块
	Type         string // 默认导出: default、命名导出:named、unknown
	Identifier   string // 唯一标识
}

type ExportDeclarationResult struct {
	ExportModules []ExportModule // 导出的模块内容
	Raw           string         // 源码
	Source        string         // 源文件路径  case: export { a } from "../index.ts"
	Type          string         // 类型, 预留字段，代表导出的类型，例如：变量/函数/类型等
}

func NewExportDeclarationResult() *ExportDeclarationResult {
	return &ExportDeclarationResult{
		ExportModules: make([]ExportModule, 0),
		Raw:           "",
		Type:          "",
	}
}

func (edr *ExportDeclarationResult) analyzeExportDeclaration(node *ast.ExportDeclaration, sourceCode string) {
	initExportModule := ExportDeclarationResult{
		ExportModules: make([]ExportModule, 0),
		Raw:           "",
		Type:          "",
	}

	// ✅ 解析 import 的源代码
	raw := utils.GetNodeText(node.AsNode(), sourceCode)
	initExportModule.Raw = raw

	// 解析export的模块内容
	//

	edr.ExportModules = initExportModule.ExportModules
	edr.Raw = initExportModule.Raw
	edr.Type = ""
}
