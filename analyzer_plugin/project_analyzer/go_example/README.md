# Runner 使用示例

本示例演示如何使用 `Runner` 在 Go 项目中直接调用分析器插件。

## 运行示例

### 1. 编译示例

```bash
cd analyzer_plugin/project_analyzer/example
go build -o example main.go
```

### 2. 运行示例

```bash
# 使用测试项目运行
./example /path/to/testdata/test_project

# 或使用自己的 TypeScript 项目
./example /path/to/your/typescript-project
```

## 示例代码

`main.go` 演示了以下功能：

### 1. 创建 Runner

```go
runner, err := project_analyzer.NewRunner(project_analyzer.RunnerConfig{
    ProjectRoot: projectPath,
    Exclude:     []string{"node_modules/**", "dist/**"},
    IsMonorepo:  false,
})
```

### 2. 注册分析器

```go
runner.RegisterBatch(
    &list_deps.Lister{},
    &component_deps_v2.ComponentDepsV2Analyzer{},
    &export_call.ExportCallAnalyzer{},
)
```

### 3. 配置并执行

```go
manifestPath := filepath.Join(projectPath, ".analyzer/component-manifest.json")
configs := map[string]map[string]string{
    "list-deps": {}, // 无需配置
    "component-deps-v2": {
        "manifest": manifestPath,
    },
    "export-call": {
        "manifest": manifestPath,
    },
}

results, err := runner.RunBatch(configs)
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

### list-deps
列出项目的 NPM 依赖。

- **配置**: 无需配置
- **结果**: `*list_deps.ListDepsResult`

### component-deps-v2
分析组件的外部依赖关系。

- **配置**: `manifest` - 组件清单文件路径
- **结果**: `*component_deps_v2.ComponentDepsV2Result`

### export-call
分析导出节点的引用关系。

- **配置**: `manifest` - 资产清单文件路径
- **结果**: `*export_call.ExportCallResult`

## 输出

### 控制台输出

示例程序会在控制台输出格式化的分析结果，包括：

1. **list-deps**: 每个 package.json 的 NPM 依赖列表
2. **component-deps-v2**: 每个组件的外部依赖包
3. **export-call**: 每个模块的导出节点及引用状态

### JSON 文件

结果也会保存为 JSON 文件到 `.analyzer/output/` 目录：

```
.analyzer/output/
├── list-deps.json
├── component-deps-v2.json
└── export-call.json
```

## 自定义使用

您可以基于 `main.go` 修改来满足自己的需求：

### 修改分析器列表

```go
runner.RegisterBatch(
    &export_call.ExportCallAnalyzer{},
    &unconsumed.Finder{},           // 添加未使用导出分析器
    &trace.Tracer{},                 // 添加追踪分析器
)
```

### 修改配置

```go
configs := map[string]map[string]string{
    "export-call": {
        "manifest": "/path/to/custom-manifest.json",
        "verbose":  "true",  // 如果插件支持
    },
    "unconsumed": {
        "targetFiles": "/path/to/file1.ts,/path/to/file2.ts",
    },
}
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
    case "list-deps":
        r := result.(*list_deps.ListDepsResult)
        // 处理 list-deps 结果...
    }
}
```

## 更多信息

- [Runner API 文档](../README_GO_INTEGRATION.md)
- [插件开发指南](../../README.md)
