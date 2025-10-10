# `tsmorphgo` 包使用指引

## 1. 简介

`tsmorphgo` 包是 `analyzer-ts` 项目的核心交互层。它的设计目标是提供一个高级、易用、类型安全且与 `ts-morph` 风格类似的 API，用于在 Go 语言环境中对 TypeScript/TSX 代码进行静态分析。

**核心设计原则：**

*   **简单易用**：封装了底层复杂的 AST 操作，提供面向对象的、符合直觉的 API。
*   **高性能**：通过预计算和缓存机制，最大化地复用底层 `analyzer` 包的解析结果，避免重复计算。
*   **类型安全**：提供丰富的 `IsXXX` 和 `AsXXX` 辅助函数，帮助开发者安全地进行节点类型判断和转换。

## 2. 架构与设计哲学

### 核心概念

`tsmorphgo` 的设计围绕三个核心结构体展开：`Project`, `SourceFile`, `Node`。

*   `Project`：整个 TypeScript 项目的入口和管理者。它持有项目中所有文件的解析结果。
*   `SourceFile`：代表一个独立的 `.ts`/`.tsx` 源文件。你可以从 `Project` 中获取它。
*   `Node`：代表 AST（抽象语法树）中的任意一个节点。它是进行代码导航、分析和信息获取的主要对象。

它们之间的关系如下图所示：

```mermaid
graph TD
    Project -- "1..*" --> SourceFile;
    SourceFile -- "持有引用" --> Project;
    Node -- "持有引用" --> SourceFile;

    subgraph Project [Project 结构]
        direction LR
        P_ParserResult[parserResult]
        P_SourceFiles[sourceFiles<br/>(文件缓存)]
    end

    subgraph SourceFile [SourceFile 结构]
        direction LR
        SF_FileResult[fileResult]
        SF_NodeResultMap[nodeResultMap<br/>(节点结果映射)]
    end

    subgraph Node [Node 结构]
        direction LR
        N_AstNode[ast.Node]
    end

    style Project fill:#e6ffcd
    style SourceFile fill:#cde4ff
    style Node fill:#ffcdd2
```

### 包结构设计

你可能会注意到，`tsmorphgo` 包内的所有 `.go` 文件都位于同一个目录下，呈现“平铺”的结构。这是我们遵循 Go 语言最佳实践而 **有意为之** 的设计。

在 Go 中，一个目录代表一个独立的包。创建子目录会形成新的包，从而割裂 API 的统一性，导致用户需要 `import` 多个路径。为了提供一个像 `ts-morph` 那样内聚、统一的 API 体验，我们将所有功能都归于 `package tsmorphgo` 之下。

我们通过清晰的文件命名来在逻辑上组织代码：

*   **核心对象层**: `project.go`, `sourcefile.go`, `node.go` (定义了核心的三个对象)
*   **类型系统层**: `types.go` (提供了 `IsXXX`, `AsXXX` 等类型工具)
*   **专用 API 层**: `declaration.go`, `expression.go` (按节点类型提供了专用的便捷 API)
*   **语义分析层**: `references.go`, `symbol.go` (提供了 `FindReferences` 等高级功能)

这种方式是 Go 语言中管理包内复杂度的标准做法。

## 3. 快速上手 (使用姿势)

以下是一个完整的示例，展示了如何使用 `tsmorphgo` 来分析一段代码，并查找引用。

```go
package main

import (
	"fmt"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

func main() {
	// 1. 使用 NewProjectFromSources 从内存中的源码创建项目
	// 注意：为了让 FindReferences 等语义功能正常工作，建议包含一个 tsconfig.json。
	project := tsmorphgo.NewProjectFromSources(map[string]string{
		"/tsconfig.json": `{
			"compilerOptions": {
				"baseUrl": ".",
				"paths": {
					"@/*": ["src/*"]
				}
			}
		}`,
		"/src/api.ts":   `export const getUser = () => {};`,
		"/src/index.ts": `
			import { getUser } from '@/api';
			getUser(); // 使用处
		`,
	})

	// 2. 获取要分析的源文件
	indexFile := project.GetSourceFile("/src/index.ts")

	// 3. 查找目标节点：找到 `getUser()` 这次调用的标识符
	var usageNode *tsmorphgo.Node
	indexFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "getUser" {
			if parent := node.GetParent(); parent != nil && parent.Kind == ast.KindCallExpression {
				usageNode = &node
			}
		}
	})

	if usageNode == nil {
		panic("未能找到 getUser 的使用节点")
	}

	// 4. 调用核心语义 API：FindReferences
	fmt.Println("为 getUser() 调用查找所有引用...")
	refs, err := tsmorphgo.FindReferences(*usageNode)
	if err != nil {
		panic(err)
	}

	// 5. 处理并打印结果
	fmt.Printf("成功找到 %d 个引用:\n", len(refs))
	for _, refNode := range refs {
		fmt.Printf("- 文件: %s, 行号: %d, 文本: '%s'\n",
			refNode.GetSourceFile().GetFilePath(),
			refNode.GetStartLineNumber(),
			strings.TrimSpace(refNode.GetText()),
		)
	}
}
```

## 4. API 指南

*   **项目创建**: `NewProject`, `NewProjectFromSources`
    *   *位置*: `tsmorphgo/project.go`
*   **文件与节点获取**: `project.GetSourceFile`, `sourceFile.ForEachDescendant`
    *   *位置*: `tsmorphgo/project.go`, `tsmorphgo/sourcefile.go`
*   **节点导航**: `node.GetParent`, `node.GetAncestors`, `sdk.GetFirstChild`
    *   *位置*: `tsmorphgo/node.go`
*   **类型查询**: `IsIdentifier`, `IsCallExpression`, `AsImportDeclaration`, `AsVariableDeclaration` 等。
    *   *位置*: `tsmorphgo/types.go`
*   **专用 API**: `GetVariableName`, `GetCallExpressionExpression`, `GetImportSpecifierAliasNode` 等。
    *   *位置*: `tsmorphgo/declaration.go`, `tsmorphgo/expression.go`
*   **语义分析**: `FindReferences`
    *   *位置*: `tsmorphgo/references.go`

## 5. 当前状态与遗留问题

*   **已实现**: `migration_api.md` 中定义的绝大部分 API（约 95%）均已实现、测试和验证，包括最关键的 `FindReferences` 功能。

*   **待办/遗留问题**: 
    *   **`getSymbol()`**: 这是唯一剩下的主要 API。由于底层 `typescript-go` 库并未提供稳定、公开的接口来直接获取节点的语义符号，我们决定 **暂时搁置** 此功能的实现，以规避不必要的风险和复杂度。我们将在 `typescript-go` 库更新或有更明确的实现路径时，再重新审视此功能。

## 6. 总结

`tsmorphgo` 包目前功能强大、API 稳定且经过了充分测试，完全可以作为项目代码分析的核心工具投入使用。
