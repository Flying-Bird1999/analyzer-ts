# analyzer-ts 架构文档

本文档描述 analyzer-ts 的整体架构设计，包括各层的职责、依赖关系和使用方式。

## 目录

- [架构概览](#架构概览)
- [分层设计](#分层设计)
- [核心组件](#核心组件)
- [依赖关系](#依赖关系)
- [数据流](#数据流)
- [使用示例](#使用示例)

---

## 架构概览

analyzer-ts 采用分层架构设计，将代码分析能力按职责划分为独立的模块：

```mermaid
graph TB
    subgraph "应用层 Application Layer"
        APP[cmd/gitlab-analyzer<br/>cmd/local-analyzer<br/>CLI 工具]
    end

    subgraph "编排层 Orchestration Layer"
        PIPELINE[pkg/pipeline<br/>AnalysisPipeline<br/>Stage 接口]
    end

    subgraph "平台能力层 Platform Capabilities"
        GITLAB[pkg/gitlab<br/>DiffProvider<br/>MRPoster]
        GITHUB[未来: pkg/github<br/>DiffProvider<br/>PRPoster]
    end

    subgraph "分析层 Analysis Layer"
        IMPACT[pkg/impact_analysis<br/>影响分析]
        SUB_IMPACT_FILE[file_analyzer<br/>文件级分析]
        SUB_IMPACT_COMP[component_analyzer<br/>组件级分析]
        SYMBOL[pkg/symbol_analysis<br/>符号提取与分析]
        PARSER[analyzer/projectParser<br/>AST 解析]
    end

    subgraph "基础设施层 Infrastructure"
        TSMORPHGO[tsmorphgo<br/>TS/JS AST]
    end

    APP --> PIPELINE
    APP --> GITLAB

    PIPELINE --> GITLAB
    PIPELINE --> IMPACT
    PIPELINE --> SYMBOL
    PIPELINE --> PARSER

    IMPACT --> SUB_IMPACT_FILE
    IMPACT --> SUB_IMPACT_COMP
    SUB_IMPACT_COMP --> SUB_IMPACT_FILE
    IMPACT --> SYMBOL
    IMPACT --> PARSER
    SYMBOL --> PARSER
    PARSER --> TSMORPHGO

    style APP fill:#e1f5ff
    style PIPELINE fill:#fff4e6
    style GITLAB fill:#f0f0f0
    style IMPACT fill:#f9f9f9
    style SUB_IMPACT_FILE fill:#e8f5e9
    style SUB_IMPACT_COMP fill:#c8e6c9
    style SYMBOL fill:#f9f9f9
    style PARSER fill:#f9f9f9
    style TSMORPHGO fill:#f0f0f0
```

---

## 分层设计

### 1. 应用层 (Application Layer)

**位置**: `cmd/`

**职责**:
- 作为程序的入口点
- 负责组装各个模块
- 处理命令行参数
- 协调平台能力层和编排层

**特点**:
- 不包含业务逻辑
- 只负责组装和调用

### 2. 编排层 (Orchestration Layer)

**位置**: `pkg/pipeline/`

**职责**:
- 定义分析流程（Stage）
- 提供工厂函数创建完整管道
- 执行管道并收集结果
- 不依赖具体平台

**核心接口**:
```go
type Stage interface {
    Name() string
    Execute(ctx *AnalysisContext) (interface{}, error)
    Skip(ctx *AnalysisContext) bool
}
```

**提供的 Stage**:
- `DiffParserStage` - 解析 git diff
- `SymbolAnalysisStage` - 符号分析
- `ProjectParserStage` - 项目解析
- `ImpactAnalysisStage` - 影响分析（自动检测组件库）

### 3. 平台能力层 (Platform Capabilities)

**位置**: `pkg/gitlab/`, 未来 `pkg/github/`

**职责**:
- 提供特定平台的 API 客户端
- 提供 diff 获取能力
- 提供结果发布功能（MR/PR 评论）
- **不包含编排逻辑**

**设计原则**:
- ✅ 只提供能力
- ❌ 不依赖 pipeline
- ✅ 可以被 pipeline 使用

### 4. 分析层 (Analysis Layer)

**位置**: `pkg/impact_analysis/`, `pkg/symbol_analysis/`, `analyzer/projectParser/`

**职责**:
- 具体的分析逻辑
- 不关心数据来源
- 不关心结果去向

### 5. 基础设施层 (Infrastructure)

**位置**: `tsmorphgo/`

**职责**:
- 提供底层 AST 解析能力
- 被多个分析模块共享

---

## 核心组件

### pkg/pipeline

**Stage 接口**:

```mermaid
classDiagram
    class Stage {
        <<interface>>
        +Name() string
        +Execute(ctx) Result
        +Skip(ctx) bool
    }

    class DiffParserStage {
        +client GitLabClient
        +source DiffSourceType
        +Name() string
        +Execute(ctx) LineSet
    }

    class SymbolAnalysisStage {
        +includeTypes bool
        +Name() string
        +Execute(ctx) SymbolChanges
    }

    class ProjectParserStage {
        +Name() string
        +Execute(ctx) ParsingResult
    }

    class ImpactAnalysisStage {
        +manifestPath string
        +maxDepth int
        +isComponentLibrary bool
        +Name() string
        +Execute(ctx) ImpactResult
        +detectComponentLibrary()
    }

    Stage <|.. DiffParserStage
    Stage <|.. SymbolAnalysisStage
    Stage <|.. ProjectParserStage
    Stage <|.. ImpactAnalysisStage
```

**Pipeline 执行流程**:

```mermaid
sequenceDiagram
    participant App as 应用层
    participant Pipe as AnalysisPipeline
    participant S1 as Stage 1: DiffParser
    participant S2 as Stage 2: SymbolAnalysis
    participant S3 as Stage 3: ProjectParser
    participant S4 as Stage 4: ImpactAnalysis

    App->>Pipe: Execute(ctx)
    Pipe->>S1: Execute(ctx)
    S1-->>Pipe: Result1
    Pipe->>Pipe: Store Result1

    Pipe->>S2: Execute(ctx)
    S2-->>Pipe: Result2
    Pipe->>Pipe: Store Result2

    Pipe->>S3: Execute(ctx)
    S3-->>Pipe: Result3
    Pipe->>Pipe: Store Result3

    Pipe->>S4: Execute(ctx)
    S4->>S4: detectComponentLibrary()
    S4->>S4: runFileLevelAnalysis()
    alt 组件库项目
        S4->>S4: runComponentLevelAnalysis()
    end
    S4-->>Pipe: Result4
    Pipe->>Pipe: Store Result4

    Pipe-->>App: PipelineResult
```

### pkg/gitlab

**提供的能力**:

```mermaid
classDiagram
    class Client {
        +GetMergeRequestDiff()
        +ListMRComments()
        +CreateMRComment()
        +UpdateMRComment()
    }

    class DiffProvider {
        +client Client
        +projectID int
        +mrIID int
        +GetDiffFiles()
        +GetDiffAsPatch()
    }

    class MRPoster {
        +mrService MRService
        +formatter Formatter
        +PostResult()
    }

    class DiffInputSource {
        <<interface>>
        +GetDiffFiles()
        +GetPatch()
    }

    class GitLabDiffSource {
        +provider DiffProvider
        +GetDiffFiles()
        +GetPatch()
    }

    Client --> DiffProvider
    Client --> MRService
    MRService --> MRPoster
    DiffProvider --> GitLabDiffSource
    DiffInputSource <|.. GitLabDiffSource
```

### pkg/impact_analysis

**架构设计**:

impact_analysis 采用两层架构设计，将影响分析分为文件级和组件级：

```mermaid
graph TB
    subgraph "impact_analysis"
        TYPES[types.go<br/>共享类型定义]

        subgraph "file_analyzer"
            FA_ANALYZER[analyzer.go<br/>文件级分析器]
            FA_GRAPH[graph_builder.go<br/>文件依赖图构建]
            FA_PROP[propagator.go<br/>影响传播]
            FA_RESULT[result.go<br/>结果类型]
        end

        subgraph "component_analyzer"
            CA_ANALYZER[analyzer.go<br/>组件级分析器]
            CA_MAPPER[mapper.go<br/>文件到组件映射]
            CA_PROP[propagator.go<br/>组件影响传播]
            CA_RESULT[result.go<br/>结果类型]
        end
    end

    TYPES --> FA_ANALYZER
    TYPES --> CA_ANALYZER

    FA_ANALYZER --> FA_GRAPH
    FA_ANALYZER --> FA_PROP
    FA_GRAPH --> FA_PROP
    FA_PROP --> FA_RESULT

    CA_ANALYZER --> CA_MAPPER
    CA_ANALYZER --> CA_PROP
    CA_MAPPER --> CA_PROP
    CA_PROP --> CA_RESULT
    CA_MAPPER -.->|依赖| FA_ANALYZER

    style TYPES fill:#fff9c4
    style FA_ANALYZER fill:#e8f5e9
    style CA_ANALYZER fill:#c8e6c9
```

**file_analyzer (文件级分析 - 通用能力)**:

```mermaid
classDiagram
    class Analyzer {
        +graphBuilder GraphBuilder
        +propagator Propagator
        +NewAnalyzer()
        +Analyze(input) Result
    }

    class GraphBuilder {
        +parsingResult ProjectParserResult
        +BuildFileDependencyGraph() FileDependencyGraph
    }

    class FileDependencyGraph {
        +DepGraph map[string][]string
        +RevDepGraph map[string][]string
        +ExternalDeps map[string][]string
        +GetDependencies(path)
        +GetDependants(path)
    }

    class Propagator {
        +depGraph FileDependencyGraph
        +maxDepth int
        +Propagate(changedFiles) ImpactedFiles
    }

    class Result {
        +Meta FileAnalysisMeta
        +Changes []FileChangeInfo
        +Impact []FileImpactInfo
        +Paths []FileImpactPath
    }

    Analyzer --> GraphBuilder
    Analyzer --> Propagator
    GraphBuilder --> FileDependencyGraph
    Propagator --> FileDependencyGraph
    Propagator --> Result
```

**component_analyzer (组件级分析 - 组件库专用)**:

```mermaid
classDiagram
    class ComponentAnalyzer {
        +mapper ComponentMapper
        +propagator Propagator
        +NewAnalyzer(manifest, parsingResult)
        +Analyze(input) Result
    }

    class ComponentMapper {
        +componentManifest ComponentManifest
        +MapFileToComponent(path) string
        +BuildComponentDependencyGraph() ComponentDependencyGraph
    }

    class ComponentDependencyGraph {
        +DepGraph map[string][]string
        +RevDepGraph map[string][]string
        +SymbolImports map[string][]ComponentSymbolImport
        +GetDependencies(comp)
        +GetDependants(comp)
    }

    class ComponentPropagator {
        +depGraph ComponentDependencyGraph
        +maxDepth int
        +Propagate(changedComponents) ImpactedComponents
    }

    class ComponentResult {
        +Meta ComponentAnalysisMeta
        +Changes []ComponentChange
        +Impact []ComponentImpactInfo
        +Paths []ComponentImpactPathInfo
    }

    ComponentAnalyzer --> ComponentMapper
    ComponentAnalyzer --> ComponentPropagator
    ComponentMapper --> ComponentDependencyGraph
    ComponentPropagator --> ComponentDependencyGraph
    ComponentPropagator --> ComponentResult
    ComponentMapper ..> FileAnalyzer : uses
```

**组件依赖查找机制**:

```mermaid
flowchart TD
    A[开始: BuildComponentDependencyGraph] --> B[加载 ComponentManifest]
    B --> C[初始化 ComponentMapper<br/>构建 file → component 映射]

    C --> D[遍历所有文件级依赖]
    D --> E[获取源文件所属组件]
    E --> F{是否属于组件?}

    F -->|否| G[跳过]
    F -->|是| H[获取目标文件所属组件]

    H --> I{目标是否属于组件?}
    I -->|否| G
    I -->|是| J{源组件 ≠ 目标组件?}

    J -->|否| G
    J -->|是| K[记录跨组件依赖]

    K --> L[添加到 DepGraph<br/>sourceComp → targetComp]
    L --> M[添加到 RevDepGraph<br/>targetComp → sourceComp]
    M --> N[记录符号导入关系<br/>SymbolImports]

    G --> O[还有更多依赖?]
    N --> O
    O -->|是| D
    O -->|否| P[构建完成]

    style K fill:#FFD700
    style P fill:#87CEEB
```

---

## 依赖关系

### 正确的依赖方向

```mermaid
graph TD
    A[应用层 cmd/] --> B[编排层 pkg/pipeline]
    A --> C[平台层 pkg/gitlab]

    B --> C
    B --> D[分析层 pkg/impact_analysis]
    B --> E[分析层 pkg/symbol_analysis]

    D --> D1[file_analyzer]
    D --> D2[component_analyzer]
    D2 --> D1

    E --> F[analyzer/projectParser]
    F --> G[tsmorphgo]

    C -.->|不依赖| B

    style A fill:#e1f5ff
    style B fill:#fff4e6
    style C fill:#f0f0f0
    style D fill:#f9f9f9
    style D1 fill:#e8f5e9
    style D2 fill:#c8e6c9
    style E fill:#f9f9f9
    style F fill:#f9f9f9
    style G fill:#f0f0f0
```

**关键原则**:
- ✅ 高层可以依赖低层
- ✅ 编排层可以平台能力层
- ❌ 平台能力层不依赖编排层
- ✅ component_analyzer 可以依赖 file_analyzer

### 模块依赖图

```mermaid
graph LR
    subgraph "无依赖"
        tsmorphgo[tsmorphgo]
    end

    subgraph "第一层"
        parser[analyzer/projectParser]
    end

    subgraph "第二层"
        symbol[pkg/symbol_analysis]
        file_analyzer[file_analyzer]
        gitlab[pkg/gitlab]
    end

    subgraph "第三层"
        component_analyzer[component_analyzer]
    end

    subgraph "第四层"
        pipeline[pkg/pipeline]
    end

    parser --> tsmorphgo
    symbol --> parser
    symbol --> tsmorphgo
    file_analyzer --> parser
    component_analyzer --> file_analyzer
    component_analyzer --> parser
    pipeline --> symbol
    pipeline --> file_analyzer
    pipeline --> component_analyzer
    pipeline --> gitlab

    style tsmorphgo fill:#e8f5e9
    style parser fill:#c8e6c9
    style symbol fill:#a5d6a7
    style file_analyzer fill:#81c784
    style component_analyzer fill:#66bb6a
    style gitlab fill:#fff59d
    style pipeline fill:#4fc3f7
```

---

## 数据流

### GitLab MR 分析完整流程

```mermaid
flowchart TD
    START([开始: GitLab MR]) --> ENV[读取环境变量]
    ENV --> CONFIG[创建 GitLabConfig]

    CONFIG --> CLIENT[创建 GitLabClient]
    CONFIG --> DIFF[创建 DiffProvider]

    DIFF --> PIPELINE[创建 AnalysisPipeline]
    PIPELINE --> S1[Stage 1: DiffParser]
    PIPELINE --> S2[Stage 2: SymbolAnalysis]
    PIPELINE --> S3[Stage 3: ProjectParser]
    PIPELINE --> S4[Stage 4: ImpactAnalysis]

    S1 --> |diff 数据| S2
    S2 --> |符号变更| S3
    S3 --> |AST 数据| S4

    S4 --> DETECT{检测组件库?}
    DETECT -->|是| BOTH[文件级 + 组件级分析]
    DETECT -->|否| FILE_ONLY[仅文件级分析]

    BOTH --> RESULT[获取 ImpactAnalysisResult]
    FILE_ONLY --> RESULT

    RESULT --> POSTER[创建 MRPoster]
    POSTER --> POST[发布 MR 评论]

    POST --> END([完成])

    style START fill:#e1f5ff
    style END fill:#c8e6c9
    style PIPELINE fill:#fff4e6
    style RESULT fill:#f9f9f9
    style POST fill:#ffe0b2
    style BOTH fill:#c8e6c9
    style FILE_ONLY fill:#e8f5e9
```

### 数据在各层之间的流转

```mermaid
stateDiagram-v2
    [*] --> GitLabAPI: GitLab API 获取 diff
    GitLabAPI --> DiffFiles: DiffFile[] 格式
    DiffFiles --> LineSet: DiffParser 解析
    LineSet --> SymbolChanges: SymbolAnalysis 提取符号
    SymbolChanges --> ProjectParser: 项目解析
    ProjectParser --> FileImpact: file_analyzer 文件级分析

    FileImpact --> CheckManifest: 检查 component-manifest.json
    CheckManifest --> ComponentImpact: 存在: component_analyzer
    CheckManifest --> Markdown: 不存在: 直接格式化

    ComponentImpact --> Markdown: 格式化结果
    Markdown --> MREndpoint: 发布到 GitLab MR
    MREndpoint --> [*]
```

### 影响分析数据流

```mermaid
flowchart LR
    INPUT[文件变更列表] --> PARSE[ProjectParser<br/>解析 Import 声明]

    PARSE --> FILE_GRAPH[GraphBuilder<br/>构建文件依赖图]
    FILE_GRAPH --> FILE_PROP[Propagator<br/>BFS 传播影响]
    FILE_PROP --> FILE_RESULT[FileAnalysisResult]

    FILE_RESULT --> CHECK{component-manifest.json<br/>存在?}

    CHECK -->|是| MAP[ComponentMapper<br/>文件 → 组件映射]
    CHECK -->|否| OUTPUT[输出文件级结果]

    MAP --> COMP_GRAPH[构建组件依赖图]
    COMP_GRAPH --> COMP_PROP[ComponentPropagator<br/>BFS 传播影响]
    COMP_PROP --> COMP_RESULT[ComponentAnalysisResult]

    COMP_RESULT --> OUTPUT
    FILE_RESULT --> OUTPUT

    style INPUT fill:#e1f5ff
    style FILE_RESULT fill:#e8f5e9
    style COMP_RESULT fill:#c8e6c9
    style OUTPUT fill:#fff9c4
```

---

## 使用示例

### 示例 1: GitLab CI 环境

```go
// cmd/gitlab-analyzer/main.go

package main

import (
    "context"
    "github.com/Flying-Bird1999/analyzer-ts/pkg/gitlab"
    "github.com/Flying-Bird1999/analyzer-ts/pkg/pipeline"
    "github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
    ctx := context.Background()
    projectRoot := "/path/to/project"

    // 1. 从环境变量读取配置
    config, _ := gitlab.ReadConfigFromEnv()

    // 2. 创建 GitLab 客户端和 diff 提供者
    client := gitlab.NewClient(config.URL, config.Token)
    diffSource := gitlab.NewGitLabDiffSource(client, config.ProjectID, config.MRIID)

    // 3. 创建 tsmorphgo 项目
    project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
        RootPath: projectRoot,
    })
    defer project.Close()

    // 4. 创建分析上下文
    analysisCtx := pipeline.NewAnalysisContext(ctx, projectRoot, project)

    // 5. 创建并配置 pipeline（自动检测组件库）
    pipe := pipeline.NewGitLabPipeline(&pipeline.GitLabPipelineConfig{
        Client:      client,
        DiffSource:  pipeline.DiffSourceGitLab,
        ProjectRoot: projectRoot,
        ProjectID:   config.ProjectID,
        MRIID:       config.MRIID,
        MaxDepth:    10,
    })

    // 6. 执行 pipeline
    result, _ := pipe.Execute(analysisCtx)

    // 7. 发布结果到 MR
    poster := gitlab.NewMRPoster(client, config.ProjectID, config.MRIID)
    impactResult, _ := result.GetResult("影响分析")
    poster.PostResult(ctx, impactResult)
}
```

### 示例 2: 本地文件分析

```go
// cmd/local-analyzer/main.go

func main() {
    ctx := context.Background()
    projectRoot := "/path/to/project"

    // 1. 从文件读取 diff
    diffSource := gitlab.NewFileDiffSource("changes.patch")

    // 2. 创建 pipeline
    project := tsmorphgo.NewProject(...)
    defer project.Close()

    analysisCtx := pipeline.NewAnalysisContext(ctx, projectRoot, project)

    pipe := pipeline.NewPipeline("Local Analysis")
    pipe.AddStage(pipeline.NewDiffParserStage(diffSource, ...))
    pipe.AddStage(pipeline.NewSymbolAnalysisStage())
    pipe.AddStage(pipeline.NewProjectParserStage())
    pipe.AddStage(pipeline.NewImpactAnalysisStage("", 10))

    // 3. 执行并输出到控制台
    result, _ := pipe.Execute(analysisCtx)
    impactResult, _ := result.GetResult("影响分析")
    fmt.Println(impactResult.ToConsole())
}
```

### 示例 3: 直接使用 file_analyzer

```go
// 适用于任何前端项目（不需要 component-manifest.json）

import (
    "github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
    "github.com/Flying-Bird1999/analyzer-ts/pkg/impact_analysis"
    "github.com/Flying-Bird1999/analyzer-ts/pkg/impact_analysis/file_analyzer"
)

func main() {
    projectRoot := "/path/to/project"

    // 1. 解析项目
    config := projectParser.NewProjectParserConfig(projectRoot, nil, false, nil)
    parsingResult := projectParser.NewProjectParserResult(config)
    parsingResult.ProjectParser()

    // 2. 创建文件级分析器
    analyzer := file_analyzer.NewAnalyzer(parsingResult, 20)

    // 3. 定义变更文件
    input := &file_analyzer.Input{
        ChangedFiles: []impact_analysis.FileChange{
            {Path: "/path/to/changed/file.ts", Type: impact_analysis.ChangeTypeModified},
        },
    }

    // 4. 执行分析
    result, _ := analyzer.Analyze(input)

    // 5. 输出结果
    fmt.Println(result.ToConsole())
}
```

### 示例 4: 直接使用 component_analyzer

```go
// 适用于组件库项目（需要 component-manifest.json）

import (
    "encoding/json"
    "github.com/Flying-Bird1999/analyzer-ts/pkg/impact_analysis"
    "github.com/Flying-Bird1999/analyzer-ts/pkg/impact_analysis/component_analyzer"
    "github.com/Flying-Bird1999/analyzer-ts/pkg/impact_analysis/file_analyzer"
)

func main() {
    projectRoot := "/path/to/project"

    // 1. 解析项目
    config := projectParser.NewProjectParserConfig(projectRoot, nil, false, nil)
    parsingResult := projectParser.NewProjectParserResult(config)
    parsingResult.ProjectParser()

    // 2. 加载组件清单
    manifestData, _ := os.ReadFile(".analyzer/component-manifest.json")
    var manifest impact_analysis.ComponentManifest
    json.Unmarshal(manifestData, &manifest)

    // 3. 先执行文件级分析
    fileAnalyzer := file_analyzer.NewAnalyzer(parsingResult, 20)
    fileInput := &file_analyzer.Input{...}
    fileResult, _ := fileAnalyzer.Analyze(fileInput)

    // 4. 再执行组件级分析
    compAnalyzer := component_analyzer.NewAnalyzer(&manifest, parsingResult, 10)
    compInput := &component_analyzer.Input{
        FileResult: convertToFileResultProxy(fileResult),
    }
    compResult, _ := compAnalyzer.Analyze(compInput)

    // 5. 输出结果
    fmt.Println(compResult.ToConsole())
}
```

---

## 设计原则

### 1. 依赖倒置原则 (DIP)

- 高层模块不依赖低层模块
- 抽象不依赖具体实现
- `pipeline` 定义 `Stage` 接口，具体 Stage 实现该接口

### 2. 单一职责原则 (SRP)

- `pkg/gitlab` 只提供 GitLab 相关能力
- `pkg/pipeline` 只负责编排
- `file_analyzer` 只负责文件级影响分析
- `component_analyzer` 只负责组件级影响分析

### 3. 开闭原则 (OCP)

- 通过 `Stage` 接口，可以扩展新的分析阶段
- 通过 `DiffInputSource` 接口，可以扩展新的 diff 来源
- 通过两层影响分析架构，可以支持不同类型的项目

### 4. 接口隔离原则 (ISP)

- `DiffInputSource` 只定义获取 diff 的方法
- `MRPoster` 只定义发布结果的方法

---

## 常见问题

### Q: 为什么 gitlab 不依赖 pipeline？

**A**: 因为 `gitlab` 是能力提供者，应该保持独立。如果 `gitlab` 依赖 `pipeline`，那么：
- 无法在非 pipeline 场景使用 gitlab 的能力
- 造成循环依赖的风险
- 违反了依赖倒置原则

### Q: 如何添加新的平台支持（如 GitHub）？

**A**: 创建 `pkg/github` 包，实现与 `pkg/gitlab` 相同的接口：
- `Client` - GitHub API 客户端
- `DiffProvider` - 从 GitHub 获取 diff
- `PRPoster` - 发布 PR 评论

然后在应用层使用时，只需替换导入即可。

### Q: file_analyzer 和 component_analyzer 的区别是什么？

**A**:
- **file_analyzer**: 通用能力，适用于所有前端项目
  - 输入：文件变更列表 + 项目解析结果
  - 输出：受影响的文件列表
  - 不依赖 component-manifest.json

- **component_analyzer**: 组件库专用能力
  - 输入：file_analyzer 结果 + component-manifest.json
  - 输出：受影响的组件列表
  - 依赖 file_analyzer 的结果

### Q: 如何添加新的分析阶段？

**A**: 实现 `Stage` 接口：
```go
type MyCustomStage struct {}

func (s *MyCustomStage) Name() string {
    return "MyCustomStage"
}

func (s *MyCustomStage) Execute(ctx *pipeline.AnalysisContext) (interface{}, error) {
    // 分析逻辑
}

func (s *MyCustomStage) Skip(ctx *pipeline.AnalysisContext) bool {
    // 跳过条件
}
```

然后添加到 pipeline：
```go
pipe.AddStage(&MyCustomStage{})
```

---

## 版本历史

### v3.0 (当前架构 - 2026)

**重大重构：影响分析分层架构**

- `pkg/impact_analysis` 分为两层：
  - `file_analyzer` - 文件级影响分析（通用能力）
  - `component_analyzer` - 组件级影响分析（组件库专用）
- 移除 `analyzer.go`, `matcher.go`, `propagator.go`, `assessor.go` 等旧文件
- 使用 `entry` 字段替代 `scopes`（协议统一）
- `pkg/pipeline/gitlab_pipeline.go` 重构：
  - 新增 `ProjectParserStage` 统一项目解析
  - `ImpactAnalysisStage` 自动检测组件库项目
  - 支持非组件库项目的影响分析

### v2.0 (2025)

- 重构为分层架构
- `pkg/impact_analysis` 从 `analyzer_plugin` 迁移到 `pkg/`
- `pkg/pipeline` 作为独立编排层
- `pkg/gitlab` 只提供能力，不依赖 pipeline

### v1.0 (旧架构)

- `analyzer_plugin/project_analyzer/impact_analysis` 独立命令行工具
- `pkg/gitlab/integration.go` 包含编排逻辑
- 存在循环依赖风险

---

## 相关文档

- [impact_analysis_refactor_plan.md](../impact_analysis_refactor_plan.md) - 影响分析重构计划
- [pkg/impact_analysis/README.md](./impact_analysis/README.md) - 影响分析模块文档
- [README.md](../README.md) - 项目总览
