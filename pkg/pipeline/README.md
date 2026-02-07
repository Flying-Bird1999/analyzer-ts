# pkg/pipeline - 代码影响分析管道

> 一个基于管道模式的代码影响分析系统，用于分析 TypeScript 项目的代码变更影响范围

## 目录

- [概述](#概述)
- [架构设计](#架构设计)
- [数据流向](#数据流向)
- [核心组件](#核心组件)
- [执行流程](#执行流程)
- [配置选项](#配置选项)
- [使用示例](#使用示例)
- [结果结构](#结果结构)
- [设计模式](#设计模式)
- [扩展指南](#扩展指南)

---

## 概述

`pkg/pipeline` 是一个可扩展的、分阶段的代码分析管道，用于：

- **解析 Git diff** - 支持多种输入源（GitLab API、diff 文件、git 命令、字符串）
- **符号级变更分析** - 将行级变更转换为符号级变更
- **影响范围计算** - 分析文件级和组件级影响范围
- **智能检测** - 自动识别组件库项目并执行相应分析

### 核心特性

| 特性 | 说明 |
|------|------|
| **管道模式** | 将复杂分析流程拆分为独立阶段，易于扩展和维护 |
| **多输入源** | 支持 GitLab API、diff 文件、git 命令、直接字符串 |
| **智能检测** | 自动识别组件库项目，动态选择分析策略 |
| **上下文传递** | 通过 `AnalysisContext` 实现阶段间数据共享 |
| **错误处理** | 完善的错误处理和取消机制 |

---

## 架构设计

### 整体架构图

```mermaid
graph TB
    subgraph INPUT["输入源 (4种)"]
        API["GitLab API<br/>GetMergeRequestDiff"]
        FILE["Diff 文件<br/>.patch"]
        SHA["Git 命令<br/>git diff"]
        STR["Diff 字符串<br/>直接传入"]
    end

    subgraph STAGE1["阶段1: Diff 解析<br/>(diff_parser_stage.go)"]
        S1_IN["输入: Diff 内容<br/>diff --git a/src/utils.ts b/src/utils.ts"]
        S1_PROC["gitlab.Parser<br/>解析 git diff 格式"]
        S1_OUT["输出: ChangedLineSetOfFiles<br/>map[string]map[int]bool<br/>{'src/utils.ts': {5, 6, 7}}"]
    end

    subgraph STAGE2["阶段2: 项目解析<br/>(project_parser_stage.go)"]
        S2_IN["输入: 项目根目录<br/>/path/to/project"]
        S2_PROC["tsmorphgo.Project<br/>构建 AST + 符号表"]
        S2_OUT["输出: ProjectParserResult<br/>*tsmorphgo.Project<br/>包含所有源文件的 AST"]
    end

    subgraph STAGE3["阶段3: 符号分析<br/>(symbol_analysis_stage.go)"]
        S3_IN["输入: ChangedLineSetOfFiles + Project<br/>变更行 + AST"]
        S3_PROC["symbol_analysis.Analyzer<br/>行级变更 → 符号级变更"]
        S3_OUT["输出: FileAnalysisResults<br/>SymbolChange[]<br/>[{name:'formatDate', file:'src/utils.ts'}]"]
    end

    subgraph STAGE4["阶段4: 影响分析<br/>(impact_analysis_stage.go)"]
        S4_IN["输入: SymbolChange[]<br/>符号级变更列表"]
        S4_DECIDE{"检测项目类型"}
        S4_FILE["file_analyzer<br/>文件级影响分析"]
        S4_COMP["component_analyzer<br/>组件级影响分析"]
        S4_OUT["输出: ImpactAnalysisResult<br/>FileResult + ComponentResult<br/>影响链路 + 影响层级"]
    end

    subgraph OUTPUT["输出结果"]
        FI["文件影响列表<br/>直接变更 / 间接受影响"]
        CI["组件影响列表<br/>变更组件 / 受影响组件"]
        CHAIN["影响链路<br/>A → B → C"]
    end

    API --> S1_IN
    FILE --> S1_IN
    SHA --> S1_IN
    STR --> S1_IN

    S1_IN --> S1_PROC
    S1_PROC --> S1_OUT

    S1_OUT -.->|依赖| S3_IN
    S2_IN --> S2_PROC
    S2_PROC --> S2_OUT
    S2_OUT -.->|依赖| S3_IN

    S3_IN --> S3_PROC
    S3_PROC --> S3_OUT
    S3_OUT --> S4_IN

    S4_IN --> S4_DECIDE
    S4_DECIDE -->|普通项目| S4_FILE
    S4_DECIDE -->|组件库| S4_COMP
    S4_FILE --> S4_OUT
    S4_COMP --> S4_OUT

    S4_OUT --> FI
    S4_OUT --> CI
    S4_OUT --> CHAIN

    style INPUT fill:#e3f2fd
    style STAGE1 fill:#fff9c4
    style STAGE2 fill:#ffccbc
    style STAGE3 fill:#d1c4e9
    style STAGE4 fill:#c8e6c9
    style OUTPUT fill:#f8bbd0

    style S1_IN fill:#fffde7
    style S1_OUT fill:#fff59d
    style S2_IN fill:#ffe0b2
    style S2_OUT fill:#ffb74d
    style S3_IN fill:#e1bee7
    style S3_OUT fill:#ba68c8
    style S4_IN fill:#c8e6c9
    style S4_OUT fill:#81c784
```

### 模块调用关系图

```mermaid
flowchart TD
    subgraph Entry["入口: cmd/impact.go"]
        IMPACT["impact 命令"]
    end

    subgraph PipelineCore["管道核心: pkg/pipeline"]
        GP["gitlab_pipeline.go<br/>GitLabPipeline"]
        P["pipeline.go<br/>AnalysisPipeline"]
        S["stage.go<br/>Stage接口"]
        AC["context.go<br/>AnalysisContext"]
    end

    subgraph Stages["阶段实现"]
        DPS["diff_parser_stage.go<br/>DiffParserStage"]
        PPS["project_parser_stage<br/>ProjectParserStage"]
        SAS["symbol_analysis_stage<br/>SymbolAnalysisStage"]
        IAS["impact_analysis_stage<br/>ImpactAnalysisStage"]
    end

    subgraph Algorithms["算法实现"]
        GPAG["pkg/gitlab/parser.go"]
        SA["pkg/symbol_analysis/"]
        FA["pkg/impact_analysis/file_analyzer"]
        CA["pkg/impact_analysis/component_analyzer"]
    end

    IMPACT --> GP
    GP --> P
    P --> S
    P --> AC

    P --> DPS
    P --> PPS
    P --> SAS
    P --> IAS

    DPS --> GPAG
    PPS --> TSM["tsmorphgo.Project"]
    SAS --> SA
    IAS --> FA
    IAS --> CA

    style Entry fill:#e1f5fe
    style PipelineCore fill:#fff3e0
    style Stages fill:#c8e6c9
    style Algorithms fill:#f3e5f5
```

---

## 数据流向

### 完整数据流图

```mermaid
flowchart LR
    subgraph Input["输入源 (4种)"]
        API["GitLab API<br/>GetMergeRequestDiff"]
        FILE["Diff 文件<br/>.patch"]
        SHA["Git SHA<br/>git diff"]
        STR["Diff 字符串<br/>直接传入"]
    end

    subgraph Stage1["阶段1: Diff解析"]
        DP["DiffParserStage"]
        CL["ChangedLineSetOfFiles<br/>map[string]map[int]bool<br/>文件 → 变更行号"]
    end

    subgraph Stage2["阶段2: 项目解析"]
        PP["ProjectParserStage"]
        PR["ProjectParserResult<br/>*tsmorphgo.Project<br/>AST + 符号表"]
    end

    subgraph Stage3["阶段3: 符号分析"]
        SA["SymbolAnalysisStage"]
        FR["FileAnalysisResults<br/>SymbolChange[]<br/>符号级变更"]
    end

    subgraph Stage4["阶段4: 影响分析"]
        IA["ImpactAnalysisStage"]
        IR["ImpactAnalysisResult<br/>FileResult + ComponentResult"]
    end

    subgraph Output["输出结果"]
        FI["文件影响列表<br/>直接 + 间接"]
        CI["组件影响列表<br/>变更 + 受影响"]
        SR["符号变更详情<br/>类型 + 导出状态"]
    end

    API --> DP
    FILE --> DP
    SHA --> DP
    STR --> DP

    DP --> CL
    CL --> PP
    PP --> PR
    PR --> SA
    SA --> FR
    FR --> IA
    IA --> IR

    IR --> FI
    IR --> CI
    IR --> SR

    style Input fill:#e3f2fd
    style Stage1 fill:#fff9c4
    style Stage2 fill:#ffccbc
    style Stage3 fill:#d1c4e9
    style Stage4 fill:#c8e6c9
    style Output fill:#f8bbd0
```

### 阶段间数据传递

```mermaid
sequenceDiagram
    participant CLI as cmd/impact.go
    participant Pipe as Pipeline
    participant S1 as DiffParserStage
    participant S2 as ProjectParserStage
    participant S3 as SymbolAnalysisStage
    participant S4 as ImpactAnalysisStage
    participant CTX as AnalysisContext

    CLI->>Pipe: Execute(ctx)
    Pipe->>CTX: 初始化上下文

    Pipe->>S1: Execute(ctx)
    S1->>S1: 解析 diff
    S1->>CTX: SetResult("diff", ChangedLineSetOfFiles)
    S1-->>Pipe: StageResult{Success}

    Pipe->>S2: Execute(ctx)
    S2->>CTX: GetResult("diff")
    S2->>S2: 解析 TS 项目
    S2->>CTX: SetResult("project", *tsmorphgo.Project)
    S2-->>Pipe: StageResult{Success}

    Pipe->>S3: Execute(ctx)
    S3->>CTX: GetResult("diff", "project")
    S3->>S3: 分析符号变更
    S3->>CTX: SetResult("symbols", FileAnalysisResults)
    S3-->>Pipe: StageResult{Success}

    Pipe->>S4: Execute(ctx)
    S4->>CTX: GetResult("symbols")
    S4->>S4: 分析影响范围
    S4->>CTX: SetResult("impact", ImpactAnalysisResult)
    S4-->>Pipe: StageResult{Success}

    Pipe-->>CLI: PipelineResult
```

---

## 核心组件

### 1. AnalysisContext（分析上下文）

贯穿整个管道的共享上下文，用于阶段间数据传递。

```go
// pkg/pipeline/context.go
type AnalysisContext struct {
    // Go 标准上下文（用于取消和超时）
    context.Context

    // 项目根目录
    projectRoot string

    // AST 项目实例（阶段2填充）
    project *tsmorphgo.Project

    // 排除路径模式
    excludePaths []string

    // 配置选项
    options map[string]interface{}

    // 阶段结果存储（key: 阶段名称, value: 阶段输出）
    results map[string]interface{}
}
```

| 字段 | 类型 | 说明 | 填充阶段 |
|------|------|------|----------|
| `Context` | `context.Context` | Go 标准上下文 | 初始化 |
| `projectRoot` | `string` | 项目根目录 | 初始化 |
| `project` | `*tsmorphgo.Project` | AST 项目实例 | 阶段2 |
| `excludePaths` | `[]string` | 排除路径模式 | 初始化 |
| `options` | `map[string]interface{}` | 额外配置选项 | 初始化 |
| `results` | `map[string]interface{}` | 阶段结果存储 | 各阶段 |

**核心方法**：
```go
// 获取项目实例
func (c *AnalysisContext) GetProject() *tsmorphgo.Project

// 获取阶段结果
func (c *AnalysisContext) GetResult(key string) (interface{}, bool)

// 设置阶段结果
func (c *AnalysisContext) SetResult(key string, result interface{})

// 获取配置选项
func (c *AnalysisContext) GetOption(key string) (interface{}, bool)

// 设置配置选项
func (c *AnalysisContext) SetOption(key string, value interface{})
```

### 2. Pipeline（管道）

#### 2.1 AnalysisPipeline（通用管道）

```go
// pkg/pipeline/pipeline.go
type AnalysisPipeline struct {
    stages []Stage
}

// 添加阶段
func (p *AnalysisPipeline) AddStage(stage Stage)

// 执行管道
func (p *AnalysisPipeline) Execute(ctx *AnalysisContext) *PipelineResult
```

#### 2.2 GitLabPipeline（GitLab MR 专用管道）

```go
// pkg/pipeline/gitlab_pipeline.go
type GitLabPipeline struct {
    *AnalysisPipeline
    config *GitLabPipelineConfig
}

// 创建 GitLab 管道
func NewGitLabPipeline(config *GitLabPipelineConfig) *GitLabPipeline
```

### 3. Stage（阶段接口）

```go
// pkg/pipeline/stage.go
type Stage interface {
    // 阶段名称
    Name() string

    // 执行阶段逻辑
    Execute(ctx *AnalysisContext) (*StageResult, error)

    // 是否跳过此阶段
    Skip(ctx *AnalysisContext) bool
}
```

#### 3.1 DiffParserStage

**输入**：Diff 数据（来自 API/文件/命令/字符串）
**输出**：`ChangedLineSetOfFiles`

```go
// pkg/pipeline/diff_parser_stage.go
type DiffParserStage struct {
    diffSource  DiffSourceType
    diffFile    string
    diffSHA     string
    client      GitLabClient
    projectID   int
    mrIID       int
}
```

#### 3.2 ProjectParserStage

**输入**：项目路径
**输出**：`*tsmorphgo.Project`

```go
// pkg/pipeline/gitlab_pipeline.go (内嵌阶段)
type ProjectParserStage struct {
    projectRoot string
    excludePaths []string
}
```

#### 3.3 SymbolAnalysisStage

**输入**：`ChangedLineSetOfFiles` + `*tsmorphgo.Project`
**输出**：`FileAnalysisResults`

```go
// pkg/pipeline/symbol_analysis_stage.go
type SymbolAnalysisStage struct {
    gitRoot string
}
```

#### 3.4 ImpactAnalysisStage

**输入**：`FileAnalysisResults`
**输出**：`ImpactAnalysisResult`

```go
// pkg/pipeline/gitlab_pipeline.go (内嵌阶段)
type ImpactAnalysisStage struct {
    manifestPath string
    maxDepth     int
}
```

---

## 执行流程

### 命令执行流程（cmd/impact.go）

```mermaid
flowchart TD
    START([impact 命令开始]) --> VALIDATE{参数验证}

    VALIDATE -->|失败| ERROR1([返回错误])
    VALIDATE -->|成功| SOURCE{确定输入源}

    SOURCE -->|API| API_CFG[检查 GitLab 参数]
    SOURCE -->|File| FILE_CFG[检查文件路径]
    SOURCE -->|SHA| SHA_CFG[检查 SHA 格式]
    SOURCE -->|String| STR_CFG[检查 diff 字符串]

    API_CFG -->|失败| ERROR1
    API_CFG -->|成功| BUILD[构建管道配置]
    FILE_CFG --> BUILD
    SHA_CFG --> BUILD
    STR_CFG --> BUILD

    BUILD --> PIPELINE[创建 GitLabPipeline]
    PIPELINE --> CTX[创建 AnalysisContext]
    CTX --> EXEC[执行管道]

    EXEC -->|失败| FORMAT_ERR[格式化错误]
    EXEC -->|成功| FORMAT_OK[格式化结果]

    FORMAT_ERR --> OUTPUT([输出到控制台])
    FORMAT_OK --> OUTPUT

    style VALIDATE fill:#fff9c4
    style SOURCE fill:#ffccbc
    style BUILD fill:#d1c4e9
    style EXEC fill:#c8e6c9
    style OUTPUT fill:#f8bbd0
    style ERROR1 fill:#ffcdd2
```

### 管道执行流程（pkg/pipeline）

```mermaid
flowchart TD
    START([Pipeline.Execute]) --> INIT{初始化上下文}
    INIT --> CREATE[创建 PipelineResult]
    CREATE --> LOOP{遍历阶段}

    LOOP -->|有阶段| CHECK_SKIP{检查 Skip}
    CHECK_SKIP -->|跳过| NEXT[记录跳过状态]
    CHECK_SKIP -->|执行| EXECUTE[调用 Stage.Execute]

    EXECUTE -->|成功| SAVE[保存结果到上下文]
    EXECUTE -->|失败| ERROR[记录错误]

    SAVE --> NEXT
    ERROR --> CHECK_STOP{是否继续?}
    CHECK_STOP -->|停止| RETURN_ERR([返回失败结果])
    CHECK_STOP -->|继续| NEXT

    NEXT --> LOOP
    LOOP -->|无阶段| RETURN_OK([返回成功结果])

    style INIT fill:#e3f2fd
    style EXECUTE fill:#fff9c4
    style SAVE fill:#c8e6c9
    style ERROR fill:#ffcdd2
    style RETURN_OK fill:#a5d6a7
    style RETURN_ERR fill:#ef9a9a
```

### 各阶段详细流程

#### 阶段1: Diff 解析

```mermaid
flowchart TD
    START([DiffParserStage.Execute]) --> CHECK{检查输入源类型}

    CHECK -->|API| API[调用 GitLab API]
    CHECK -->|File| FILE[读取 diff 文件]
    CHECK -->|SHA| SHA[执行 git 命令]
    CHECK -->|String| STR[使用 diff 字符串]

    API --> PARSE[解析 diff]
    FILE --> PARSE
    SHA --> PARSE
    STR --> PARSE

    PARSE --> VALID{验证解析结果}
    VALID -->|失败| ERROR([返回错误])
    VALID -->|成功| BUILD[构建 ChangedLineSetOfFiles]

    BUILD --> SAVE[保存到上下文]
    SAVE --> RETURN([返回成功])

    style CHECK fill:#fff9c4
    style PARSE fill:#ffccbc
    style BUILD fill:#c8e6c9
    style ERROR fill:#ffcdd2
```

#### 阶段2: 项目解析

```mermaid
flowchart TD
    START([ProjectParserStage.Execute]) --> CREATE[创建 tsmorphgo.ProjectConfig]
    CREATE --> NEW[调用 tsmorphgo.NewProject]
    NEW --> LOAD[加载源文件和 AST]
    LOAD --> SAVE[保存 project 到上下文]
    SAVE --> RETURN([返回成功])

    style CREATE fill:#e3f2fd
    style NEW fill:#fff9c4
    style LOAD fill:#c8e6c9
```

#### 阶段3: 符号分析

```mermaid
flowchart TD
    START([SymbolAnalysisStage.Execute]) --> GET_DIFF[获取 ChangedLineSetOfFiles]
    GET_DIFF --> GET_PROJ[获取 tsmorphgo.Project]
    GET_PROJ --> CREATE_ANALYZER[创建符号分析器]
    CREATE_ANALYZER --> ANALYZE[分析变更行]
    ANALYZE --> MAP[映射到符号]
    MAP --> BUILD[构建 FileAnalysisResults]
    BUILD --> SAVE[保存到上下文]
    SAVE --> RETURN([返回成功])

    style GET_DIFF fill:#e3f2fd
    style GET_PROJ fill:#e3f2fd
    style CREATE_ANALYZER fill:#fff9c4
    style ANALYZE fill:#ffccbc
    style BUILD fill:#c8e6c9
```

#### 阶段4: 影响分析

```mermaid
flowchart TD
    START([ImpactAnalysisStage.Execute]) --> GET_SYM[获取 FileAnalysisResults]
    GET_SYM --> CONVERT[转换为 ChangedSymbol 列表]
    CONVERT --> CHECK_LIB{检测组件库}

    CHECK_LIB -->|是组件库| COMP_ANALYZER[执行组件级分析]
    CHECK_LIB -->|非组件库| FILE_ANALYZER[执行文件级分析]

    FILE_ANALYZER --> SAVE[保存结果]
    COMP_ANALYZER --> FILE_ANALYZER

    SAVE --> BUILD[构建 ImpactAnalysisResult]
    BUILD --> RETURN([返回成功])

    style GET_SYM fill:#e3f2fd
    style CONVERT fill:#fff9c4
    style CHECK_LIB fill:#ffccbc
    style FILE_ANALYZER fill:#c8e6c9
    style COMP_ANALYZER fill:#d1c4e9
```

---

## 配置选项

### GitLabPipelineConfig

```go
// pkg/pipeline/gitlab_pipeline.go
type GitLabPipelineConfig struct {
    // ========== Diff 输入源 ==========
    DiffSource  DiffSourceType  // 输入源类型: API, File, SHA, String
    DiffFile    string          // Diff 文件路径 (DiffSourceFile)
    DiffSHA     string          // Git SHA 或分支 (DiffSourceSHA)
    ProjectRoot string          // 项目根目录（必需）
    GitRoot     string          // Git 仓库根（可选，默认 = ProjectRoot）

    // ========== GitLab API（DiffSourceAPI 时必需）==========
    ProjectID   int             // GitLab 项目 ID
    MRIID       int             // GitLab MR IID
    Client      GitLabClient    // GitLab API 客户端

    // ========== 组件分析 ==========
    ManifestPath string         // 组件清单路径（自动检测组件库）
    DepsFile     string         // 依赖配置文件

    // ========== 分析配置 ==========
    MaxDepth     int            // 影响分析最大深度（默认 10）
}
```

### DiffSourceType

```go
// pkg/pipeline/gitlab_pipeline.go
const (
    DiffSourceString DiffSourceType = "string" // 直接传入 diff 字符串
    DiffSourceFile   DiffSourceType = "file"   // 从文件读取 diff
    DiffSourceSHA    DiffSourceType = "sha"    // 执行 git diff 命令
    DiffSourceAPI    DiffSourceType = "api"    // 从 GitLab API 获取
)
```

### 配置示例

#### 使用 Diff 字符串

```go
config := &pipeline.GitLabPipelineConfig{
    DiffSource:  pipeline.DiffSourceString,
    ProjectRoot: "/path/to/project",
    MaxDepth:    10,
}
// 通过 context.SetOption("diffString", diffContent) 传入
```

#### 使用 Diff 文件

```go
config := &pipeline.GitLabPipelineConfig{
    DiffSource:  pipeline.DiffSourceFile,
    DiffFile:    "/path/to/mr.patch",
    ProjectRoot: "/path/to/project",
    MaxDepth:    10,
}
```

#### 使用 Git Diff

```go
config := &pipeline.GitLabPipelineConfig{
    DiffSource:  pipeline.DiffSourceSHA,
    DiffSHA:     "HEAD~1 HEAD",  // 或 "main...feature-branch"
    ProjectRoot: "/path/to/project",
    GitRoot:     "/path/to/git/repo",  // monorepo 场景需要
    MaxDepth:    10,
}
```

#### 使用 GitLab API

```go
config := &pipeline.GitLabPipelineConfig{
    DiffSource:  pipeline.DiffSourceAPI,
    ProjectRoot: "/path/to/project",
    ProjectID:   123,
    MRIID:       456,
    Client:      gitLabClient,
    MaxDepth:    10,
}
```

---

## 使用示例

### 完整使用流程

```go
package main

import (
    "context"
    "fmt"
    "github.com/Flying-Bird1999/analyzer-ts/pkg/pipeline"
)

func main() {
    // 1. 创建管道配置
    config := &pipeline.GitLabPipelineConfig{
        DiffSource:  pipeline.DiffSourceFile,
        DiffFile:    "/path/to/mr.patch",
        ProjectRoot: "/path/to/project",
        ManifestPath: ".analyzer/component-manifest.json", // 可选，启用组件级分析
        MaxDepth:    10,
    }

    // 2. 创建分析上下文
    ctx := context.Background()
    analysisCtx := pipeline.NewAnalysisContext(ctx, config.ProjectRoot, nil)

    // 3. 创建管道
    pipe := pipeline.NewGitLabPipeline(config)

    // 4. 执行管道
    result, err := pipe.Execute(analysisCtx)
    if err != nil {
        panic(err)
    }

    // 5. 获取结果
    if !result.IsSuccessful() {
        fmt.Printf("管道执行失败: %v\n", result.GetErrors())
        return
    }

    // 6. 处理影响分析结果
    impactResult, _ := result.GetResult("影响分析（文件级）")
    if impact, ok := impactResult.(*pipeline.ImpactAnalysisResult); ok {
        fmt.Printf("受影响文件数: %d\n", impact.FileResult.Meta.ImpactFileCount)
        if impact.IsComponentLibrary {
            fmt.Printf("受影响组件数: %d\n", impact.ComponentResult.Meta.ImpactComponentCount)
        }
    }
}
```

### 命令行使用

```bash
# 使用 diff 文件
analyzer-ts impact \
  --project-root /path/to/project \
  --diff-file /path/to/mr.patch

# 使用 git diff
analyzer-ts impact \
  --project-root /path/to/project \
  --diff-sha "HEAD~1 HEAD"

# 使用 GitLab API
analyzer-ts impact \
  --project-root /path/to/project \
  --diff-source api \
  --project-id 123 \
  --mr-iid 456 \
  --gitlab-token $GITLAB_TOKEN

# 指定组件清单（启用组件级分析）
analyzer-ts impact \
  --project-root /path/to/project \
  --diff-file /path/to/mr.patch \
  --manifest-path .analyzer/component-manifest.json
```

---

## 结果结构

### ImpactAnalysisResult

```go
// pkg/pipeline/types.go
type ImpactAnalysisResult struct {
    // 文件级结果（始终存在）
    FileResult *FileImpactResult `json:"fileResult,omitempty"`

    // 组件级结果（仅组件库项目）
    ComponentResult *ComponentImpactResult `json:"componentResult,omitempty"`

    // 是否为组件库项目
    IsComponentLibrary bool `json:"isComponentLibrary"`
}
```

### FileImpactResult

```go
type FileImpactResult struct {
    // 元数据
    Meta struct {
        TotalFileCount    int  // 项目总文件数
        ChangedFileCount  int  // 直接变更的文件数
        ImpactFileCount   int  // 间接受影响的文件数
    }

    // 直接变更的文件（影响层级 = 0）
    Changes []FileChange

    // 间接受影响的文件（影响层级 ≥ 1）
    Impact []FileImpact
}

type FileChange struct {
    FilePath     string
    ChangeType   string  // "modified", "added", "deleted"
    ChangedLines []int   // 变更的行号
}

type FileImpact struct {
    FilePath     string
    ImpactLevel  int     // 影响层级（0=直接，1=间接，2+=传递）
    ImpactType   string  // "direct", "indirect", "transitive"
    DependedBy   []string // 依赖链
}
```

### ComponentImpactResult

```go
type ComponentImpactResult struct {
    // 元数据
    Meta struct {
        TotalComponentCount   int  // 组件库总组件数
        ChangedComponentCount int  // 直接变更的组件数
        ImpactComponentCount  int  // 间接受影响的组件数
    }

    // 直接变更的组件
    Changes []ComponentChange

    // 间接受影响的组件
    Impact []ComponentImpact
}

type ComponentChange struct {
    ComponentName string
    EntryFile     string
    ChangeType    string
    ChangedFiles  []string
}

type ComponentImpact struct {
    ComponentName string
    ImpactLevel   int
    ImpactReason  string  // "direct", "dependency", "transitive"
    DependencyPath []string
}
```

---

## 设计模式

### 1. 管道模式（Pipeline Pattern）

将复杂处理流程分解为多个独立的处理阶段：

```mermaid
flowchart LR
    Input[输入数据] --> S1[阶段1]
    S1 --> S2[阶段2]
    S2 --> S3[阶段3]
    S3 --> S4[阶段4]
    S4 --> Output[输出结果]

    style S1 fill:#fff9c4
    style S2 fill:#ffccbc
    style S3 fill:#d1c4e9
    style S4 fill:#c8e6c9
```

**优势**：
- 各阶段独立，易于测试
- 灵活组合，可动态添加/移除阶段
- 易于扩展，添加新阶段不影响现有代码

### 2. 上下文模式（Context Pattern）

使用 `AnalysisContext` 在阶段间传递数据：

```mermaid
flowchart TD
    CTX[AnalysisContext]
    S1[阶段1]
    S2[阶段2]
    S3[阶段3]

    S1 --> CTX
    CTX --> S2
    S2 --> CTX
    CTX --> S3
    S3 --> CTX

    style CTX fill:#e1f5fe
```

**优势**：
- 解耦阶段间依赖
- 统一的数据访问接口
- 支持阶段跳过和错误恢复

### 3. 策略模式（Strategy Pattern）

根据项目类型动态选择分析策略：

```mermaid
flowchart TD
    START[影响分析阶段] --> CHECK{检测项目类型}

    CHECK -->|组件库| COMP_STRATEGY[组件级分析策略]
    CHECK -->|普通项目| FILE_STRATEGY[文件级分析策略]

    COMP_STRATEGY --> COMP_EXEC[执行组件分析器]
    FILE_STRATEGY --> FILE_EXEC[执行文件分析器]

    COMP_EXEC --> MERGE[合并结果]
    FILE_EXEC --> MERGE
    MERGE --> OUTPUT[返回结果]

    style CHECK fill:#fff9c4
    style COMP_STRATEGY fill:#d1c4e9
    style FILE_STRATEGY fill:#c8e6c9
```

---

## 扩展指南

### 添加自定义阶段

```go
// 1. 实现 Stage 接口
type CustomStage struct {
    name string
    config map[string]interface{}
}

func (s *CustomStage) Name() string {
    return s.name
}

func (s *CustomStage) Execute(ctx *pipeline.AnalysisContext) (*pipeline.StageResult, error) {
    // 从上下文获取前置阶段的结果
    prevResult, ok := ctx.GetResult("前一阶段名称")
    if !ok {
        return nil, fmt.Errorf("缺少前置结果")
    }

    // 执行自定义逻辑
    result := doSomething(prevResult)

    // 保存结果到上下文
    ctx.SetResult(s.Name(), result)

    return &pipeline.StageResult{
        Status: pipeline.StageStatusSuccess,
        Data:   result,
    }, nil
}

func (s *CustomStage) Skip(ctx *pipeline.AnalysisContext) bool {
    // 根据上下文决定是否跳过
    return false
}

// 2. 添加到管道
pipe := pipeline.NewGitLabPipeline(config)
pipe.AddStage(&CustomStage{name: "自定义阶段"})
```

### 添加新的输入源

```go
// 1. 定义新的 DiffSourceType
const DiffSourceCustom DiffSourceType = "custom"

// 2. 在 DiffParserStage 中添加处理逻辑
func (s *DiffParserStage) Execute(ctx *pipeline.AnalysisContext) (*StageResult, error) {
    var diffContent string
    var err error

    switch s.diffSource {
    case DiffSourceCustom:
        diffContent, err = s.parseCustomSource(ctx)
    // ... 其他 case
    }

    // ... 后续处理
}

func (s *DiffParserStage) parseCustomSource(ctx *pipeline.AnalysisContext) (string, error) {
    // 自定义解析逻辑
    return "", nil
}
```

### 添加新的分析器

```go
// 1. 创建分析器
type CustomAnalyzer struct {
    project *tsmorphgo.Project
    config  map[string]interface{}
}

func (a *CustomAnalyzer) Analyze(changes []SymbolChange) (*CustomResult, error) {
    // 分析逻辑
    return &CustomResult{}, nil
}

// 2. 在新阶段中使用分析器
func (s *CustomAnalysisStage) Execute(ctx *pipeline.AnalysisContext) (*StageResult, error) {
    analyzer := NewCustomAnalyzer(ctx.GetProject(), nil)
    result, err := analyzer.Analyze(changes)
    // ...
}
```

---

## 关键文件索引

| 组件 | 文件路径 | 说明 |
|------|----------|------|
| **命令入口** | `cmd/impact.go` | CLI 命令定义和执行 |
| **管道核心** | `pkg/pipeline/pipeline.go` | 通用管道执行器 |
| **GitLab 管道** | `pkg/pipeline/gitlab_pipeline.go` | GitLab MR 专用管道 |
| **阶段接口** | `pkg/pipeline/stage.go` | Stage 接口定义 |
| **上下文** | `pkg/pipeline/context.go` | AnalysisContext 定义 |
| **Diff 解析** | `pkg/pipeline/diff_parser_stage.go` | Diff 解析阶段 |
| **符号分析** | `pkg/pipeline/symbol_analysis_stage.go` | 符号分析阶段 |
| **GitLab 解析器** | `pkg/gitlab/parser.go` | Git diff 解析器 |
| **符号分析** | `pkg/symbol_analysis/analyzer.go` | 符号分析算法 |
| **文件影响** | `pkg/impact_analysis/file_analyzer` | 文件级影响分析 |
| **组件影响** | `pkg/impact_analysis/component_analyzer` | 组件级影响分析 |

---

## 附录：错误处理

### 管道错误处理流程

```mermaid
flowchart TD
    START([阶段执行]) --> TRY{尝试执行}
    TRY -->|成功| SUCCESS[记录成功状态]
    TRY -->|失败| ERROR{错误类型?}

    ERROR -->|致命错误| FATAL[停止管道]
    ERROR -->|可恢复| CONTINUE[记录错误继续]

    SUCCESS --> NEXT{下一阶段?}
    FATAL --> END([返回失败结果])
    CONTINUE --> NEXT

    NEXT -->|有| TRY
    NEXT -->|无| CHECK{检查整体状态}

    CHECK -->|全部成功| OK([返回成功结果])
    CHECK -->|有错误| WARN([返回警告结果])

    style SUCCESS fill:#c8e6c9
    style FATAL fill:#ef9a9a
    style CONTINUE fill:#fff9c4
    style OK fill:#a5d6a7
    style WARN fill:#ffcc80
```

### 错误处理示例

```go
result, err := pipe.Execute(analysisCtx)

// 检查整体成功状态
if !result.IsSuccessful() {
    fmt.Printf("管道执行失败:\n")
    for stageName, stageErr := range result.GetErrors() {
        fmt.Printf("  - %s: %v\n", stageName, stageErr)
    }
}

// 检查特定阶段
if stageResult, ok := result.GetStageResult("Diff解析"); ok {
    if stageResult.Status != pipeline.StageStatusSuccess {
        fmt.Printf("Diff解析阶段失败: %s\n", stageResult.Error)
    }
}

// 检查跳过的阶段
skipped := result.GetSkippedStages()
if len(skipped) > 0 {
    fmt.Printf("跳过的阶段: %v\n", skipped)
}
```

---

**文档版本**: v2.0
**最后更新**: 2025-01-07
**维护者**: analyzer-ts 团队
