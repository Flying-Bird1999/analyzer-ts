# Project Analyzer 架构

`project_analyzer` 是 `analyzer-ts` 工具的核心分析引擎。它采用插件式架构，允许开发者轻松添加新的分析器来扩展工具的功能。

## 核心设计理念

该架构的核心设计思想是将**项目解析**与**代码分析**分离。

1. **一次解析 (Parse Once)**: `analyzer/projectParser` 模块负责对整个 TypeScript 项目进行一次性的深度解析。它会遍历所有相关的 `.ts`/`.tsx` 文件，构建每个文件的抽象语法树 (AST)，并提取出所有关键信息（如导入、导出、函数、变量、依赖等）。这个过程成本较高，但只需要执行一次。
2. **多次分析 (Analyze Many Times)**: 解析完成后，会生成一个包含整个项目所有信息的 `ProjectParserResult` 对象。这个对象被封装在 `ProjectContext` 中，然后被传递给一系列独立的、专业的**分析器 (Analyzer)**。
3. **插件式分析器 (Pluggable Analyzers)**: 每个分析器都是一个独立的模块，它接收 `ProjectContext`，并从中读取预先解析好的数据来执行特定的分析任务，例如：

   * 检查NPM依赖 (`dependency`)
   * 统计 `any` 类型的使用 (`countAny`)
   * 查找未使用的导出 (`unconsumed`)
   * 构建调用链 (`callgraph`)

这种设计使得添加新功能变得非常高效，因为开发者可以专注于分析逻辑，而无需关心如何解析 TypeScript 代码。

## 文件命名规范

为了保持项目结构的一致性和可读性，所有分析器插件都应遵循以下文件命名约定：

*   **主逻辑文件**: 每个插件的主 Go 文件应与其父目录同名。
    *   例如：`countAny` 插件的逻辑应放在 `countAny/countAny.go` 文件中。
*   **结果文件**: 定义分析结果数据结构的文件应命名为 `result.go`。
*   **类型文件**: 定义插件内部使用的其他数据结构的文件应命名为 `types.go`。
*   **辅助函数文件**: 插件内部使用的辅助函数可以放在 `utils.go` 或 `helpers.go` 文件中。

**示例结构:**

```
analyzer_plugin/project_analyzer/
└── my_analyzer/
    ├── my_analyzer.go   <- 主逻辑 (实现 Analyzer 接口)
    ├── result.go        <- 结果 (实现 Result 接口)
    ├── types.go         <- 内部类型
    └── README.md        <- 插件说明
```

## 核心接口

为了确保所有分析器都遵循统一的规范，我们定义了两个核心接口：

### `Analyzer`

所有分析器都必须实现此接口。

```go
// Analyzer 是所有分析器模块都必须实现的接口。
type Analyzer interface {
    // Name 返回分析器的唯一名称。
    Name() string
    // Configure 用于从命令行接收参数并配置分析器。
    Configure(params map[string]string) error
    // Analyze 是执行分析的核心方法。
    // 它接收包含项目解析结果的上下文，并返回一个 Result 对象。
    Analyze(ctx *ProjectContext) (Result, error)
}
```

### `Result`

每个分析器的 `Analyze` 方法都必须返回一个实现了 `Result` 接口的对象。

```go
// Result 是所有分析结果都必须实现的接口。
type Result interface {
    // Name 返回结果的名称，通常与分析器名称对应。
    Name() string
    // Summary 返回一个单行的、易于阅读的摘要信息。
    Summary() string
    // ToJSON 将完整的分析结果序列化为 JSON 格式。
    ToJSON(indent bool) ([]byte, error)
    // ToConsole 返回一个适合在控制台打印的、格式化的字符串。
    ToConsole() string
}
```

## 如何新增一个分析器

按照以下步骤，您可以轻松地为 `analyzer-ts` 添加一个新的分析器。

### 步骤 1: 创建分析器逻辑目录

在 `analyzer_plugin/project_analyzer/` 目录下，为您的新分析器创建一个新的子目录来存放其核心逻辑。例如，我们要创建一个名为 `deadcode` 的分析器来查找无效代码。

正确的目录结构如下：

```
analyzer_plugin/project_analyzer/
├── cmd/
│   └── ... (其他命令)
└── deadcode/            <- 在这里创建新目录
    ├── deadcode.go      <- 分析器实现
    └── result.go        <- 结果实现
```

### 步骤 2: 实现 `Analyzer` 和 `Result` 接口

在 `deadcode.go` 和 `result.go` 中，实现核心接口。

**`deadcode/result.go`:**

```go
package deadcode

import "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"

// Result 保存分析结果
type Result struct {
    DeadFiles []string
}

func (r *Result) Name() string { return "Dead Code Result" }
func (r *Result) Summary() string { return "Found X dead files." }
func (r *Result) ToJSON(indent bool) ([]byte, error) {
    return project_analyzer.ToJSONBytes(r, indent)
}
func (r *Result) ToConsole() string {
    // ... 返回格式化的控制台输出
    return "..."
}
```

**`deadcode/deadcode.go`:**

```go
package deadcode

import "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"

// DeadCodeAnalyzer 实现了 Analyzer 接口
type DeadCodeAnalyzer struct {
    // 可在此处添加配置字段
}

func (a *DeadCodeAnalyzer) Name() string { return "Dead Code Analyzer" }
func (a *DeadCodeAnalyzer) Configure(params map[string]string) error {
    // ... 从 params 解析并设置配置
    return nil
}

func (a *DeadCodeAnalyzer) Analyze(ctx *project_analyzer.ProjectContext) (project_analyzer.Result, error) {
    parsingResult := ctx.ParsingResult
    // 在这里实现您的核心分析逻辑...
    result := &Result{
        DeadFiles: []string{"path/to/dead/file.ts"},
    }
    return result, nil
}
```

### 步骤 3: 创建并注册 Cobra 命令

为了让用户能够从命令行调用您的分析器，您需要在**中央命令目录** `analyzer_plugin/project_analyzer/cmd/` 下为其创建一个新文件。

**`analyzer_plugin/project_analyzer/cmd/deadcode.go` (新文件):**

```go
package cmd

import (
    "fmt"
    "github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
    "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
    "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/deadcode" // 导入您的分析器逻辑包
    "github.com/Flying-Bird1999/analyzer-ts/cmd" // 确保导入根 cmd 包
    "github.com/spf13/cobra"
)

// DeadCodeCmd 定义了新的 cobra 命令
var DeadCodeCmd = &cobra.Command{
    Use:   "find-dead-code",
    Short: "Finds dead code in the project",
    Run: func(cmd *cobra.Command, args []string) {
        // ... 获取命令行标志 (input, output, etc.)

        // 1. 解析项目
        config := projectParser.NewProjectParserConfig(input, exclude, isMonorepo)
        ar := projectParser.NewProjectParserResult(config)
        ar.ProjectParser()

        // 2. 创建分析器实例
        analyzer := &deadcode.DeadCodeAnalyzer{}
        // analyzer.Configure(...)

        // 3. 创建上下文并执行分析
        ctx := &project_analyzer.ProjectContext{
            ProjectRoot:   input,
            Exclude:       exclude,
            IsMonorepo:    isMonorepo,
            ParsingResult: ar,
        }
        result, err := analyzer.Analyze(ctx)
        if err != nil {
            fmt.Println("Error:", err)
            return
        }

        // 4. 处理并输出结果
        fmt.Println(result.ToConsole())
        // ... 或写入文件
    },
}

// 使用 init() 函数将新命令自动注册到根命令
func init() {
    // 将命令添加到 project_analyzer 的根命令下
    AnalyzeCmd.AddCommand(DeadCodeCmd)
    // 在这里为新命令定义标志
    DeadCodeCmd.Flags().StringVarP(&input, "input", "i", "", "Path to the project root")
    // ... 其他标志
}
```

### 步骤 4: 验证

您**无需**修改 `main.go`。因为 `main.go` 已经导入了 `analyzer_plugin/project_analyzer/cmd`，所以您在 `cmd` 目录中创建的任何带有 `init()` 函数的新 `.go` 文件都将被自动加载和注册。

完成以上步骤后，重新构建 (`go build`) 项目，新的 `find-dead-code` 命令即可使用。