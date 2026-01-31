# 组件库资产采集流程方案
## 一、核心设计理念
**一次扫描，多次复用**

```mermaid
flowchart TD
    Start[开始采集] --> Scan[analyzer-ts scan<br/>获取所有文件]

    Scan --> FileList[完整文件列表<br/>核心数据基础]

    FileList --> Layer1[层1: 基础资产提取]
    Layer1 --> Layer2[层2: 关联资产采集]
    Layer2 --> Layer3[层3: 派生资产计算]

    style FileList fill:#fff4e1
```

**关键点**：

+ `analyzer-ts scan` 是唯一的数据入口
+ 文件列表是所有后续采集的基础
+ 避免重复扫描，最大化数据复用

---

## 二、采集流程总览
```mermaid
flowchart TD
    P0[阶段0: 准备] --> P1[阶段1: 统一扫描]
    P1 --> P2[阶段2: 组件识别]
    P2 --> P3[阶段3: 资产采集]
    P3 --> P4[阶段4: 影响分析]

    P0 --> Config[加载配置<br/>.asset-repos.json]
    P0 --> Clone[克隆关联仓库]

    P1 --> ATS[analyzer-ts scan<br/>完整文件列表]
    P1 --> GitLog[git log]

    P2 --> Comp[识别组件<br/>标准/过渡/废弃]

    P3 --> A1[文档资产]
    P3 --> A2[UX资产]
    P3 --> A3[工具资产]
    P3 --> A4[代码质量]

    P4 --> Impact[影响范围评估]

    A1 --> Output[资产JSON]
    A2 --> Output
    A3 --> Output
    A4 --> Output
    Impact --> Output

    style P1 fill:#e8f5e9
    style Output fill:#c8e6c9
```

---

## 三、数据流转与关联
### 3.1 核心数据流
```mermaid
flowchart LR
    Scan[analyzer-ts scan] --> Files[FileList<br/>完整文件列表]

    Files --> Comp[组件识别]
    Files --> Type[文件类型分类]

    Comp --> CompFiles[组件文件列表<br/>按组件分组]

    CompFiles --> Dep[依赖分析<br/>component-deps]
    CompFiles --> AST[AST解析<br/>query]

    Dep --> Impact[影响分析<br/>git diff]
    AST --> Quality[质量分析<br/>count-any/as]
```

### 3.2 数据复用关系
| 基础数据 | 派生资产 | 复用方式 |
| --- | --- | --- |
| **FileList** | 所有资产 | 唯一数据源 |
| 组件文件列表 | 内部依赖、体积、Token、单测 | 基于组件名过滤 |
| AST 解析结果 | 依赖、API 文档、质量 | 复用解析缓存 |
| Git 日志 | Changelog、影响分析 | 复用 commit 列表 |


---

## 四、组件库维度采集
### 4.1 文档资产 (Markdown)
```mermaid
flowchart TD
    Files[FileList] --> Filter[过滤 .md 文件]

    Filter --> Docs[文档列表]

    Docs --> Parse[解析 Markdown]

    Parse --> Categorize{分类}

    Categorize --> Dev[开发规范]
    Categorize --> Best[最佳实践]
    Categorize --> Guide[使用指南]

    Dev --> DocAssets[文档资产]
    Best --> DocAssets
    Guide --> DocAssets

    style Files fill:#fff4e1
```

### 4.2 UX 规范资产 (语雀)
```mermaid
flowchart LR
    Config[UX配置] --> Yuque[语雀 API]

    Yuque --> Font[字体规范]
    Yuque --> Date[日期规范]
    Yuque --> Color[色彩规范]

    Font --> UXAssets[UX规范资产]
    Date --> UXAssets
    Color --> UXAssets
```

### 4.3 工具资产 (跨仓库)
```mermaid
flowchart TD
    Config[.asset-repos.json] --> Repos[关联仓库列表]

    Repos --> ESL[eslint-config]
    Repos --> Style[stylelint-config]
    Repos --> Play[playground]

    ESL --> Clone[并行克隆]
    Style --> Clone
    Play --> Clone

    Clone --> Extract[提取配置]

    Extract --> ToolAssets[工具资产]

    style Config fill:#e1f5fe
```

### 4.4 组件状态识别
```mermaid
flowchart TD
    Files[FileList] --> Entry[入口文件分析<br/>src/index.ts]

    Entry --> Exported[已导出组件]

    Exported --> Check1{检查状态}

    Check1 --> Deprecated[废弃组件]
    Check1 --> Experimental[过渡组件]
    Check1 --> Standard[规范组件]

    Deprecated --> CompAssets[组件资产]
    Experimental --> CompAssets
    Standard --> CompAssets

    style CompAssets fill:#c8e6c9
```

---

## 五、单组件维度采集
### 5.1 依赖采集 (基于文件列表)
```mermaid
flowchart TD
    CompFiles[组件文件列表] --> AST[analyzer-ts AST解析]

    AST --> Imports[提取 import 语句]

    Imports --> Classify{分类}

    Classify -->|来自 src/components| Internal[内部组件依赖]
    Classify -->|来自其他路径| External[NPM 包依赖]

    Internal --> Dep[依赖资产]
    External --> Dep

    style CompFiles fill:#fff4e1
```

### 5.2 组件体积 (基于文件列表)
```mermaid
flowchart LR
    CompFiles[组件文件列表] --> Size[计算文件大小]

    Size --> Lines[统计代码行数]

    Size --> Total[总体积]
    Lines --> Total

    Total --> Volume[组件体积资产]
```

**复用逻辑**：

+ 组件文件列表已包含所有相关文件
+ 直接遍历累加大小和行数
+ 无需再次扫描

### 5.3 Figma 链接 (配置关联)
```mermaid
flowchart LR
    Config[.asset-figma.json] --> Figma[Figma 链接配置]

    Figma --> Match[按组件名匹配]

    Match --> Links[Figma 资产]
```

### 5.4 CSS Token (基于文件列表 + 样式文件)
```mermaid
flowchart TD
    CompFiles[组件文件列表] --> Filter[过滤样式文件]

    Filter --> StyleFiles[.less/.css 文件]

    StyleFiles --> Extract[提取 Token]

    Extract --> Tokens[Token 资产]

    style CompFiles fill:#fff4e1
```

**复用逻辑**：

+ 从组件文件列表中筛选样式文件
+ 只解析组件相关的样式
+ 无需扫描全项目

### 5.5 单测 + 质量 (基于文件列表)
```mermaid
flowchart TD
    CompFiles[组件文件列表] --> TestFilter[过滤 .test/.spec 文件]

    TestFile --> Test[运行单测]

    CompFiles --> AST[AST 解析]

    AST --> CountAny[count-any]
    AST --> CountAs[count-as]

    Test --> TestReport[单测资产]
    CountAny --> Quality[质量资产]
    CountAs --> Quality

    style CompFiles fill:#fff4e1
```

**复用逻辑**：

+ 组件文件列表直接定位测试文件
+ AST 结果复用于质量分析
+ 避免全量扫描

---

## 六、影响分析 (基于 Changelog)
```mermaid
flowchart TD
    GitLog[git log] --> Commits[提交列表]

    Commits --> Diff[git diff<br/>本次变更]

    Diff --> Changed[变更组件列表]

    Changed --> Dep[查询依赖关系<br/>已采集]

    Dep --> Downstream[下游组件]

    Downstream --> Impact[影响范围报告]

    style Dep fill:#fff4e1
```

**复用逻辑**：

+ 复用已采集的依赖关系图
+ 变更组件 → 依赖查询 → 下游组件
+ 无需重新解析依赖

---

## 七、执行顺序
```mermaid
flowchart TD
    Start[开始采集] --> P0[准备阶段<br/>并行执行]

    P0 --> Config[加载配置]
    P0 --> Clone[克隆关联仓库]

    Config --> P1[扫描阶段]
    Clone --> P1

    P1 --> ATS[analyzer-ts scan<br/>一次性获取所有文件]
    P1 --> Log[git log]

    ATS --> Files[FileList<br/>核心数据]

    Files --> P2[识别阶段<br/>组件识别]

    P2 --> P3[采集阶段<br/>并行执行]

    P3 --> C1[组件库维度]
    P3 --> C2[单组件维度]

    C1 --> Assets[资产集合]
    C2 --> Assets

    Assets --> P4[分析阶段]

    P4 --> Diff[git diff]
    P4 --> Dep[依赖查询<br/>复用已有数据]

    Diff --> Impact[影响分析]
    Dep --> Impact

    Impact --> Output[输出 JSON]

    style Files fill:#fff4e1
    style Output fill:#c8e6c9
```

---

## 八、资产清单
### 组件库维度
| 资产类型 | 数据源 | 采集方式 | 依赖 |
| --- | --- | --- | --- |
| 文档资产 | .md 文件 | FileList 过滤 | FileList |
| UX 规范 | 语雀 API | 配置驱动 | - |
| 工具资产 | 关联仓库 | Git Clone + 解析 | 配置 |
| 组件状态 | 入口文件 | AST 分析 | FileList |
| 依赖信息 | package.json | JSON 解析 | - |
| Changelog | git log | 日志解析 | - |
| 影响分析 | git diff + 依赖图 | 差异计算 | 依赖资产 |


### 单组件维度
| 资产类型 | 数据源 | 采集方式 | 依赖 |
| --- | --- | --- | --- |
| 内部依赖 | 组件文件 | AST import 分析 | 组件文件列表 |
| 组件体积 | 组件文件 | 大小累加 | 组件文件列表 |
| Figma 链接 | 配置文件 | 配置匹配 | - |
| CSS Token | 样式文件 | Token 提取 | 组件文件列表 |
| 单测情况 | 测试文件 | Vitest 报告 | 组件文件列表 |
| 代码质量 | 组件文件 | count-any/count-as | 组件文件列表 |


---

## 九、配置文件
### .asset-repos.json (关联仓库)
```json
{
  "relatedRepos": [
    {
      "name": "eslint-config",
      "url": "git@gitlab.com:yy/eslint-config.git",
      "branch": "master",
      "assets": ["TOOL_ESLINT"]
    },
    {
      "name": "stylelint-config",
      "url": "git@gitlab.com:yy/stylelint-config.git",
      "branch": "master",
      "assets": ["TOOL_STYLELINT"]
    },
    {
      "name": "playground",
      "url": "git@gitlab.com:yy/playground.git",
      "branch": "develop",
      "assets": ["TOOL_PLAYGROUND"]
    }
  ]
}
```

### .asset-figma.json (Figma 配置)
```json
{
  "mappings": {
    "Button": "https://figma.com/file/xxx/Button",
    "Form": "https://figma.com/file/xxx/Form",
    "Table": "https://figma.com/file/xxx/Table"
  }
}
```

### .asset-yuque.json (语雀配置)
```json
{
  "baseUrl": "https://www.yuque.com/api/v2",
  "token": "${YUQUE_TOKEN}",
  "repos": [
    {
      "name": "字体规范",
      "id": "xxx/wiki/yyy",
      "type": "font"
    },
    {
      "name": "日期规范",
      "id": "xxx/wiki/zzz",
      "type": "date"
    }
  ]
}
```

---

## 十、输出格式
```json
{
  "collectedAt": "2024-01-28T10:00:00Z",
  "version": "1.0.0",

  "library": {
    "文档资产": [...],
    "UX规范资产": [...],
    "工具资产": [...],
    "依赖信息": {...},
    "changelog": {...},
    "影响分析": {...}
  },

  "components": {
    "Button": {
      "status": "standard",
      "内部依赖": ["Icon"],
      "组件体积": { "files": 3, "size": 15360, "lines": 245 },
      "Figma链接": "https://figma.com/...",
      "CSS Token": ["--color-primary", "--spacing-base"],
      "单测情况": { "覆盖": "95%", "通过": 48, "失败": 0 },
      "代码质量": { "anyCount": 0, "asCount": 2 }
    },
    "Form": {
      "status": "standard",
      "内部依赖": ["Button", "Input", "Icon"],
      "组件体积": { "files": 5, "size": 25600, "lines": 420 },
      "Figma链接": "https://figma.com/...",
      "CSS Token": ["--color-primary", "--border-radius"],
      "单测情况": { "覆盖": "88%", "通过": 120, "失败": 2 },
      "代码质量": { "anyCount": 5, "asCount": 8 }
    }
  },

  "impactAnalysis": {
    "changedComponents": ["Button"],
    "affectedComponents": ["Form", "Table", "Modal"],
    "riskLevel": "low"
  }
}
```

---

## 十一、关键优化点
### 11.1 最小化扫描次数
```plain
analyzer-ts scan (1次) → FileList → 所有后续采集
```

### 11.2 最大化数据复用
```mermaid
flowchart LR
    FileList[FileList] --> A[组件文件列表]
    A --> B[内部依赖]
    A --> C[组件体积]
    A --> D[CSS Token]
    A --> E[单测文件]
    A --> F[质量分析]

    style FileList fill:#fff4e1
    style A fill:#e8f5e9
```

### 11.3 并行化执行
+ 仓库克隆：并行
+ 采集器执行：并行（无依赖情况下）
+ Git 操作：批量调用

---

## 十二、CLI 命令
```bash
# 完整采集
asset-collector collect --project-root /path/to/project

# 仅组件库维度
asset-collector collect --scope library

# 仅单组件维度
asset-collector collect --scope component --component Button

# 包含影响分析
asset-collector collect --with-impact --base-ref HEAD~1

# 输出路径
asset-collector collect --output ./assets/assets.json
```
