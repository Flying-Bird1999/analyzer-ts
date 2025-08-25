# Analyzer 架构

`analyzer` 目录是 `analyzer-ts` 工具的解析核心。它负责将 TypeScript/JavaScript 源代码转换为结构化的、可供上层分析器使用的数据。本文档旨在阐述其内部架构，并指导开发者如何扩展其功能。

## 核心架构

本解析器采用三层分层设计，清晰地划分了从文件发现到项目聚合的各个阶段。

### 第 1 层: 文件扫描 (`analyzer/scanProject`)

这是整个解析流程的入口。`scanProject` 是一个文件发现工具，其唯一职责是高效地扫描项目目录，并根据过滤规则找出现需要被分析的文件。

- **工作原理**: `scanProject` 从一个根目录开始，使用 `filepath.Walk` 递归地遍历所有文件和子目录。它会根据用户提供的 `ignore` 模式（glob 格式）以及内置的规则（如忽略 `node_modules`）来过滤条目。为了提升效率，当一个目录被忽略时，整个目录都会被跳过。
- **产出**: `scanProject` 的产出是一个文件列表，其中包含了所有未被忽略的文件的绝对路径和元数据。这个列表是下一层解析器的输入。

### 第 2 层: 文件级解析器 (`analyzer/parser`)

`parser` 目录负责对**单个** TypeScript/JavaScript 文件进行深度解析。

- **核心驱动**: `parser.go` 中的 `Parser` 结构体是此层的核心。它接收一个文件路径，并利用 `github.com/Zzzen/typescript-go`（一个 TypeScript 官方解析器的 Go 语言绑定）生成一个完整的抽象语法树 (AST)。
- **节点遍历**: `Traverse()` 方法会深度优先遍历 AST。在遍历过程中，一个巨大的 `switch node.Kind` 语句会将不同类型的 AST 节点（如函数声明、导入语句、JSX 元素等）分发给各自专门的 `analyze...` 处理函数。
- **结果提取**: 每个 `analyze...` 函数（例如 `analyzeFunctionDeclaration`）负责从对应的 AST 节点中提取所有相关信息，并将其填充到一个专门的 Go 结构体中（例如 `FunctionDeclarationResult`）。
- **产出**: 单文件解析的最终产出是一个 `ParserResult` 结构体，它包含了从该文件中提取出的所有声明、表达式和其他重要信息的集合。此阶段的路径（如导入源）是未经处理的原始字符串。

### 第 3 层: 项目级解析器 (`analyzer/projectParser`)

`projectParser` 目录是最高层，负责编排和整合整个项目的解析过程。

- **循环解析**: 它接收由 `scanProject` 生成的文件列表，然后遍历这个列表，为每个文件调用 `parser` 来获取其文件级的 `ParserResult`。
- **路径解析与转换**: 这是 `projectParser` 的核心价值所在。它会读取项目中的 `tsconfig.json` 文件以获取路径别名（`paths` alias）。然后，它会调用一系列 `transform...` 函数（如 `transformImportDeclarations`），将 `parser` 产出的原始导入路径（例如 `@/components/Button`）转换为相对于项目根目录的绝对路径。
- **最终产出**: 整个 `analyzer` 模块的最终产出是一个 `ProjectParserResult` 对象。该对象包含了项目中所有文件的解析数据，并且所有的模块间引用都已被解析为绝对路径，为上层的分析器插件提供了可以直接使用的数据基础。

## 如何扩展解析器

### 新增一个 AST 节点类型解析

假设您需要解析一个新的 TypeScript 语法或 AST 节点，例如 `for...of` 循环。请遵循以下步骤：

**1. 定义结果结构体**

在 `analyzer/parser/` 目录下创建一个新文件，例如 `forOfStatement.go`。在该文件中，定义一个结构体来存储您想从节点中提取的信息。

```go
// analyzer/parser/forOfStatement.go
package parser

// ForOfStatementResult 存储 for...of 循环的解析结果
type ForOfStatementResult struct {
    Identifier string `json:"identifier"` // 循环中的变量名
    Iterable   string `json:"iterable"`   // 被迭代的对象
    Raw        string `json:"raw"`        // 原始代码文本
}
```

**2. 实现节点分析函数**

在同一个文件中，创建一个分析函数，它接收一个 `ast.Node` 并返回您定义的结构体。

```go
// analyzer/parser/forOfStatement.go

// ... (struct definition) ...

func analyzeForOfStatement(node *ast.ForOfStatement, sourceCode string) *ForOfStatementResult {
    // 从 node 对象中提取所需信息
    identifier := ""
    if nameNode := node.Initializer(); nameNode != nil {
        identifier = utils.GetNodeText(nameNode, sourceCode)
    }

    iterable := ""
    if iterableNode := node.Expression(); iterableNode != nil {
        iterable = utils.GetNodeText(iterableNode, sourceCode)
    }

    return &ForOfStatementResult{
        Identifier: identifier,
        Iterable:   iterable,
        Raw:        utils.GetNodeText(node.AsNode(), sourceCode),
    }
}
```

**3. 在主遍历函数中注册**

打开 `analyzer/parser/parser.go` 文件，在 `Traverse()` 方法的 `switch` 语句中，添加一个新的 `case` 来处理 `for...of` 节点。

```go
// analyzer/parser/parser.go

// ... in Parser.Traverse() ...
func (p *Parser) Traverse() {
    // ...
    walk = func(node *ast.Node) {
        // ...
        switch node.Kind {
        // ... (其他 case)
        case ast.KindForOfStatement: // <-- 新增的 case
            // 调用您新创建的分析函数
            res := analyzeForOfStatement(node.AsForOfStatement(), p.SourceCode)
            // 将结果存入 ParserResult (可能需要先在 ParserResult 中添加新的字段)
            p.Result.ForOfStatements = append(p.Result.ForOfStatements, *res)

        // ... (其他 case)
        }
        // ...
    }
    // ...
}
```

**4. (可选) 在 `ParserResult` 中添加字段**

如果需要，请在 `analyzer/parser/parser.go` 的 `ParserResult` 结构体中添加一个新字段来存储解析结果。

```go
// analyzer/parser/parser.go
type ParserResult struct {
    // ... (其他字段)
    ForOfStatements []ForOfStatementResult // <-- 新增的字段
}
```

### 新增一个单元测试

为确保您的解析逻辑正确无误，添加单元测试至关重要。测试代码位于 `analyzer/parser/test/` 目录。

**1. 创建测试文件**

为您的 `forOfStatement.go` 创建一个对应的测试文件 `forOfStatement_test.go`。

**2. 编写测试用例**

我们采用**表驱动测试**的模式。您可以参考 `functionDeclaration_test.go` 的写法。

```go
// analyzer/parser/test/forOfStatement_test.go
package parser_test

import (
    "encoding/json"
    "github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestForOfStatement(t *testing.T) {
    // 定义测试用例
    testCases := []struct {
        name         string // 测试用例描述
        code         string // 要解析的 TS 代码片段
        expectedJSON string // 期望从代码中解析出的 JSON 结果
    }{
        {
            name: "基础 for...of 循环",
            code: `const items = [1, 2]; for (const item of items) { console.log(item); }`,
            expectedJSON: `[
                {
                    "identifier": "const item",
                    "iterable": "items",
                    "raw": "for (const item of items) { console.log(item); }"
                }
            ]`,
        },
        // ... 可以添加更多测试用例
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // 1. 从源码进行解析
            p, err := parser.NewParserFromSource("test.ts", tc.code)
            assert.NoError(t, err)
            p.Traverse()

            // 2. 提取实际解析结果
            // 假设您已在 ParserResult 中添加了 ForOfStatements 字段
            actualResults := p.Result.ForOfStatements

            // 3. 将实际结果序列化为 JSON
            actualJSON, err := json.Marshal(actualResults)
            assert.NoError(t, err)

            // 4. 使用 assert.JSONEq 进行比较，它能忽略格式差异
            assert.JSONEq(t, tc.expectedJSON, string(actualJSON), "解析结果与预期不符")
        })
    }
}
```
