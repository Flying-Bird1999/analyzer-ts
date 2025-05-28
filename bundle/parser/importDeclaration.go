package parser

import (
	"main/bundle/utils"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// 解析导入模块
// - 默认导入: import Bird from './type2';
// - 命名空间导入: import * as allTypes from './type';
// - 命名导入: import { School, School2 } from './school';
// 					- import type { CurrentRes } from './type';
//      		- import { School as NewSchool } from './school';

// ==> 解析结果:
// [
//   {
//     "modules": [
//       {
//         "module": "default",
//         "type": "default",
//         "identifier": "Bird"
//       }
//     ],
//     "raw": "import Bird from './type2';",
//     "source": "./type2"
//   },
//   {
//     "modules": [
//       {
//         "module": "allTypes",
//         "type": "namespace",
//         "identifier": "allTypes"
//       }
//     ],
//     "raw": "import * as allTypes from './type';",
//     "source": "./type"
//   },
//   {
//     "modules": [
//       {
//         "module": "School",
//         "type": "named",
//         "identifier": "School"
//       },
//       {
//         "module": "School2",
//         "type": "named",
//         "identifier": "School2"
//       }
//     ],
//     "raw": "import { School, School2 } from './school';",
//     "source": "./school"
//   },
//   {
//     "modules": [
//       {
//         "module": "CurrentRes",
//         "type": "named",
//         "identifier": "CurrentRes"
//       }
//     ],
//     "raw": "import type { CurrentRes } from './type';",
//     "source": "./type"
//   },
//   {
//     "modules": [
//       {
//         "module": "School",
//         "type": "named",
//         "identifier": "NewSchool"
//       }
//     ],
//     "raw": "import { School as NewSchool } from './school';",
//     "source": "./school"
//   }
// ]

type ImportModule struct {
	ImportModule string // 模块名, 对应实际导出的内容模块
	Type         string // 默认导入: default、命名空间导入: namespace、命名导入:named、unknown
	Identifier   string //唯一标识
}

type ImportDeclarationResult struct {
	ImportModules []ImportModule // 导入的模块内容
	Raw           string         // 源码
	Source        string         // 路径
}

func NewImportDeclarationResult() *ImportDeclarationResult {
	return &ImportDeclarationResult{
		ImportModules: make([]ImportModule, 0),
		Raw:           "",
		Source:        "",
	}
}

func (idr *ImportDeclarationResult) analyzeImportDeclaration(node *ast.ImportDeclaration, sourceCode string) {
	initImportModule := ImportDeclarationResult{
		ImportModules: make([]ImportModule, 0),
		Raw:           "",
		Source:        "",
	}

	// ✅ 解析 import 的源代码
	raw := utils.GetNodeText(node.AsNode(), sourceCode)
	initImportModule.Raw = raw

	// ✅ 解析 import 的模块路径
	moduleSpecifier := node.ModuleSpecifier
	initImportModule.Source = moduleSpecifier.Text()

	if node.ImportClause != nil {
		// ✅ 解析 import 的模块内容
		importClause := node.ImportClause.AsImportClause()

		// 默认导入: import Bird from './type2';
		if ast.IsDefaultImport(node.AsNode()) {
			Name := importClause.Name().Text()
			initImportModule.ImportModules = append(initImportModule.ImportModules, ImportModule{
				ImportModule: "default",
				Type:         "default",
				Identifier:   Name,
			})
		}

		// - 命名空间导入: import * as allTypes from './type';
		namespaceNode := ast.GetNamespaceDeclarationNode(node.AsNode())
		if namespaceNode != nil {
			Name := namespaceNode.Name().Text()
			initImportModule.ImportModules = append(initImportModule.ImportModules, ImportModule{
				ImportModule: Name,
				Type:         "namespace",
				Identifier:   Name,
			})
		}

		// - 命名导入: import { School, School2 } from './school';
		// 					- import type { CurrentRes } from './type';
		//      		- import { School as NewSchool } from './school';
		if importClause.NamedBindings != nil && importClause.NamedBindings.Kind == ast.KindNamedImports {
			namedImports := importClause.NamedBindings.AsNamedImports()
			for _, element := range namedImports.Elements.Nodes {
				importSpecifier := element.AsImportSpecifier()

				if importSpecifier.PropertyName != nil {
					// import { School as NewSchool } from './school';
					Name := importSpecifier.PropertyName.Text()
					Alias := importSpecifier.Name().Text()
					initImportModule.ImportModules = append(initImportModule.ImportModules, ImportModule{
						ImportModule: Name,
						Type:         "named",
						Identifier:   Alias,
					})

				} else {
					// import { School, School2 } from './school';
					// import type { CurrentRes } from './type';
					Name := importSpecifier.Name().Text()
					initImportModule.ImportModules = append(initImportModule.ImportModules, ImportModule{
						ImportModule: Name,
						Type:         "named",
						Identifier:   Name,
					})
				}
			}
		}
	}

	idr.ImportModules = initImportModule.ImportModules
	idr.Raw = initImportModule.Raw
	idr.Source = initImportModule.Source
}
