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

## Go 项目调用方式

`project_analyzer` 支持两种调用方式，适用于不同的使用场景。

### 方式一：批量执行（ExecuteWithConfig）

**适用场景**：一次性执行多个分析器，生成完整的分析报告。

```go
import (
    "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
    // 导入 analyzer 包以触发注册
    _ "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/component_deps_v2"
    _ "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/export_call"
    _ "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/list_deps"
)

func main() {
    // 1. 创建 ProjectAnalyzer
    analyzer, _ := project_analyzer.NewProjectAnalyzer(project_analyzer.Config{
        ProjectRoot: "/path/to/project",
        Exclude:     []string{"node_modules/**", "dist/**"},
    })

    // 2. 准备执行配置（使用 AnalyzerType 常量，IDE 会自动补全）
    manifestPath := "/path/to/component-manifest.json"
    execConfig := project_analyzer.NewExecutionConfig().
        AddAnalyzer(project_analyzer.AnalyzerListDeps, project_analyzer.ListDepsConfig{}).
        AddAnalyzer(project_analyzer.AnalyzerComponentDepsV2, project_analyzer.ComponentDepsV2Config{
            Manifest: manifestPath,
        }).
        AddAnalyzer(project_analyzer.AnalyzerExportCall, project_analyzer.ExportCallConfig{
            Manifest: manifestPath,
        })

    // 3. 执行分析（项目只会解析一次）
    results, _ := analyzer.ExecuteWithConfig(execConfig)

    // 4. 处理结果
    listResult, _ := project_analyzer.GetResult[*list_deps.ListDepsResult](results)
    // ... 消费 listResult
}
```

**特点**：
- 一次性执行所有 analyzer
- 项目只解析一次
- 结果统一返回，便于处理

### 方式二：按需执行（RunOneT）

**适用场景**：在业务流程的不同位置按需执行特定分析器，或与业务逻辑深度集成。

#### 方式 A：不持有 context（简单场景）

```go
type MyService struct {
    analyzer *project_analyzer.ProjectAnalyzer
}

func NewService(projectPath string) (*MyService, error) {
    // NewProjectAnalyzer 会自动解析项目（耗时操作）
    analyzer, _ := project_analyzer.NewProjectAnalyzer(project_analyzer.Config{
        ProjectRoot: projectPath,
        Exclude:     []string{"node_modules/**", "dist/**"},
    })
    // 项目已解析完毕，可直接使用

    return &MyService{analyzer: analyzer}, nil
}

// 直接使用 RunOneT
func (s *MyService) CheckCodeQuality() error {
    result, err := project_analyzer.RunOneT[*countAny.CountAnyResult](
        s.analyzer,
        project_analyzer.AnalyzerCountAny,
        project_analyzer.CountAnyConfig{},
    )
    // result 直接是具体类型，无需断言
    fmt.Printf("Any count: %d\n", result.TotalCount)
    return nil
}
```

#### 方式 B：持有 context（需要传递给其他函数）

```go
type MyService struct {
    analyzer     *project_analyzer.ProjectAnalyzer
    analyzerCtx  *project_analyzer.ProjectContext  // 持有 context 供其他地方使用
}

func NewService(projectPath string) (*MyService, error) {
    // NewProjectAnalyzer 会自动解析项目并创建 context
    analyzer, _ := project_analyzer.NewProjectAnalyzer(...)

    // 获取 context（可以在业务中传递）
    ctx := analyzer.Context()

    return &MyService{
        analyzer:    analyzer,
        analyzerCtx: ctx,
    }, nil
}

// 将 context 传递给其他需要它的函数
func (s *MyService) SomeBusinessMethod() {
    // 可以将 s.analyzerCtx 传递给其他需要 ProjectContext 的函数
    otherPackage.ProcessData(s.analyzerCtx)
}
```

**特点**：
- **一步初始化**：`NewProjectAnalyzer` 自动完成解析（耗时操作）
- 项目只解析一次（结果和 context 自动缓存）
- 可以在不同业务方法中按需调用 analyzer
- 分析器调用可以与业务逻辑深度集成
- 支持多次调用不同 analyzer
- **可选持有 `context`**：通过 `analyzer.Context()` 获取
- **RunOneT 泛型函数提供类型安全，无需手动类型断言**

### 核心方法对比

| 方法 | 说明 | 使用场景 |
|------|------|----------|
| `NewProjectAnalyzer()` | 创建分析器并自动解析项目（一步完成） | 初始化时调用，会自动执行耗时操作 |
| `Context()` | 获取分析上下文 | 需要在业务中传递 context 时调用 |
| `ExecuteWithConfig()` | 批量执行多个 analyzer | 一次性获取多个分析结果 |
| `RunOneT[T]()` | 按需执行单个 analyzer，泛型函数直接返回具体类型 | **推荐**：类型安全，无需手动断言 |

### 完整示例代码

完整示例代码位于 `analyzer_plugin/project_analyzer/go_example/` 目录：

- `main.go` - 批量执行方式示例
- `ondemand_example.go` - 按需执行方式示例（使用 `go run -tags=ondemand` 运行）

## 如何新增一个分析器

按照以下步骤，您可以轻松地为 `analyzer-ts` 添加一个新的分析器。

### 步骤 1: 创建分析器逻辑目录

在 `analyzer_plugin/project_analyzer/` 目录下，为您的新分析器创建一个新的子目录来存放其核心逻辑。

### 步骤 2: 实现 `Analyzer` 和 `Result` 接口

在新目录中创建 `xxx.go` 和 `result.go` 文件，实现核心接口。

### 步骤 3: 添加注册代码

在 `xxx.go` 中添加 `init()` 函数，将分析器注册到中央注册表：

```go
func init() {
    projectanalyzer.RegisterAnalyzer("your-analyzer", func() projectanalyzer.Analyzer {
        return &YourAnalyzer{}
    })
}
```

### 步骤 4: 更新枚举常量

在 `runner.go` 的 `AnalyzerType` 枚举中添加对应的常量：

```go
const (
    // ... 其他常量
    AnalyzerYourAnalyzer AnalyzerType = "your-analyzer"
)
```

### 步骤 5: 在 cmd/analyze.go 中添加导入

在 `cmd/analyze.go` 的空白导入部分添加：

```go
import (
    // ... 其他导入
    _ "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/your_analyzer"
)
```

完成以上步骤后，新的分析器即可在命令行和 Go 代码中使用。

### 步骤 6: 定义配置类型（可选）

如果 analyzer 需要配置参数，在 `runner.go` 中定义对应的配置结构体：

```go
// YourAnalyzerConfig your-analyzer 分析器配置
type YourAnalyzerConfig struct {
    // 配置字段
}

func (c YourAnalyzerConfig) ToMap() map[string]string {
    // 转换为 map 的逻辑
}
```
