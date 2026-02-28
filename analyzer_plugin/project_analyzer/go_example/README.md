# ProjectAnalyzer 使用示例

本示例演示如何使用 `ProjectAnalyzer` 在 Go 项目中直接调用分析器插件。

## 运行示例

### 1. 编译示例

```bash
cd analyzer_plugin/project_analyzer/go_example
go build -o example main.go
```

### 2. 运行示例

```bash
# 默认分析 ./testdata/test_project
./example

# 或指定其他项目路径
./example /path/to/your/typescript-project
```

## 示例代码

`main.go` 演示了以下功能：

### 1. 创建 ProjectAnalyzer

```go
analyzer, err := project_analyzer.NewProjectAnalyzer(project_analyzer.Config{
    ProjectRoot: projectPath,
    Exclude:     []string{"node_modules/**", "dist/**"},
    IsMonorepo:  false,
})
```

### 2. 准备执行配置

```go
execConfig := project_analyzer.NewExecutionConfig().
    AddAnalyzer(&pkg_deps.PkgDepsAnalyzer{}, nil). // pkg_deps 不需要配置
    AddAnalyzer(&component_deps.ComponentDepsAnalyzer{}, map[string]string{
        "manifest": manifestPath,
    }).
    AddAnalyzer(&export_call.ExportCallAnalyzer{}, map[string]string{
        "manifest": manifestPath,
    })
```

### 3. 执行分析

```go
results, err := analyzer.ExecuteWithConfig(execConfig)
```

### 4. 直接使用结果

```go
// 类型断言为具体结果类型
exportCallResult := results["export-call"].(*export_call.ExportCallResult)

// 直接访问结果字段
for _, module := range exportCallResult.ModuleExports {
    fmt.Printf("Module: %s\n", module.ModuleName)
}
```

## 支持的分析器

本示例演示了三个内置分析器的使用：

### pkg-deps
列出项目的 NPM 依赖。

- **配置**: 无需配置
- **结果**: `*pkg_deps.PkgDepsResult`

### component-deps
分析组件的外部依赖关系。

- **配置**: `manifest` - 组件清单文件路径
- **结果**: `*component_deps.ComponentDepsResult`

### export-call
分析导出节点的引用关系。

- **配置**: `manifest` - 资产清单文件路径
- **结果**: `*export_call.ExportCallResult`

## 输出

### 控制台输出

示例程序会在控制台输出格式化的分析结果，包括：

1. **pkg-deps**: 每个 package.json 的 NPM 依赖列表
2. **component-deps**: 每个组件的外部依赖包
3. **export-call**: 每个模块的导出节点及引用状态

### JSON 文件

结果也会保存为 JSON 文件到 `.analyzer/output/` 目录：

```
.analyzer/output/
├── pkg-deps.json
├── component-deps.json
└── export-call.json
```

## 自定义使用

您可以基于 `main.go` 修改来满足自己的需求：

### 修改分析器列表

```go
execConfig := project_analyzer.NewExecutionConfig().
    AddAnalyzer(&export_call.ExportCallAnalyzer{}, config).
    AddAnalyzer(&unconsumed.Finder{}, config).           // 添加未使用导出分析器
    AddAnalyzer(&trace.Tracer{}, nil)                    // 添加追踪分析器
```

### 修改配置

```go
execConfig := project_analyzer.NewExecutionConfig().
    AddAnalyzer(&export_call.ExportCallAnalyzer{}, map[string]string{
        "manifest": "/path/to/custom-manifest.json",
        "verbose":  "true",  // 如果插件支持
    }).
    AddAnalyzer(&unconsumed.Finder{}, map[string]string{
        "targetFiles": "/path/to/file1.ts,/path/to/file2.ts",
    })
```

### 处理结果

```go
for name, result := range results {
    // 方式1: 调用 ToJSON() 序列化
    jsonData, _ := result.ToJSON(true)
    fmt.Println(string(jsonData))

    // 方式2: 类型断言后直接使用
    switch name {
    case "export-call":
        r := result.(*export_call.ExportCallResult)
        // 处理 export-call 结果...
    case "pkg-deps":
        r := result.(*pkg_deps.PkgDepsResult)
        // 处理 pkg-deps 结果...
    }
}
```

## API 概览

### 核心类型

| 类型 | 说明 |
|------|------|
| `ProjectAnalyzer` | 项目分析器，封装解析和执行流程 |
| `Config` | 分析器配置 (项目路径、排除规则等) |
| `ExecutionConfig` | 执行配置 (分析器列表及各自配置) |

### 核心方法

| 方法 | 说明 |
|------|------|
| `NewProjectAnalyzer(Config)` | 创建分析器实例 |
| `NewExecutionConfig()` | 创建执行配置 |
| `AddAnalyzer(Analyzer, Config)` | 添加分析器 (链式调用) |
| `ExecuteWithConfig(ExecutionConfig)` | 执行分析 |

## 更多信息

- [Runner API 文档](../README_GO_INTEGRATION.md)
- [插件开发指南](../../README.md)
