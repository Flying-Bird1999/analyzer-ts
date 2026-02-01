# pkg/gitlab

GitLab CI/CD 集成包，为 analyzer-ts 提供代码影响分析和 MR 评论功能。

## 功能特性

- **自动 diff 解析**: 从 GitLab API 或本地 git diff 获取代码变更
- **行级精确追踪**: 精准追踪变更的文件和行
- **组件影响分析**: 运行 component-deps-v2 和 impact-analysis 插件分析影响范围
- **MR 评论发布**: 自动在 GitLab MR 中发布分析结果评论
- **CI/CD 集成**: 自动检测 GitLab CI 环境变量，无需额外配置

## 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                     GitLabIntegration                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   DiffParser │──│ComponentDeps │──│ImpactAnalysis│      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│         │                  │                  │              │
│         ▼                  ▼                  ▼              │
│  ┌──────────────────────────────────────────────────┐      │
│  │                   Formatter                       │      │
│  └──────────────────────────────────────────────────┘      │
│                            │                               │
│                            ▼                               │
│  ┌──────────────────────────────────────────────────┐      │
│  │                   MRService                       │      │
│  └──────────────────────────────────────────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

### 核心模块

| 模块          | 文件               | 职责                        |
| ------------- | ------------------ | --------------------------- |
| GitLab 集成器 | `integration.go` | 编排完整分析流程            |
| Diff 解析器   | `diff_parser.go` | 解析 git diff，支持行级追踪 |
| GitLab 客户端 | `client.go`      | GitLab API v4 客户端        |
| MR 服务       | `mr_service.go`  | MR 评论操作                 |
| 格式化器      | `formatter.go`   | JSON 转 Markdown            |
| 命令接口      | `command.go`     | Cobra 命令接口              |

## 使用方式

### GitLab CI 模式（自动检测）

在 `.gitlab-ci.yml` 中配置：

```yaml
impact-analysis:
  stage: analyze
  rules:
    - if: '$CI_PIPELINE_SOURCE == "merge_request_event"'
  variables:
    GIT_DEPTH: 0
  script:
    - analyzer-ts gitlab impact -i .
  artifacts:
    when: always
    paths:
      - impact-report.html
    expire_in: 30 days
```

需要配置的环境变量：

```bash
# GitLab Token（在项目 Settings > CI/CD > Variables 中配置）
GITLAB_TOKEN = glpat-xxxxxxxxxxxxxx

# GitLab CI 自动注入的变量（无需手动配置）
CI_SERVER_URL
CI_PROJECT_ID
CI_MERGE_REQUEST_IID
CI_MERGE_REQUEST_DIFF_BASE_SHA
```

### 本地测试模式

```bash
analyzer-ts gitlab impact -i /path/to/project \
  --gitlab-url https://gitlab.example.com \
  --gitlab-token $GITLAB_TOKEN \
  --project-id 123 \
  --mr-id 456 \
  --diff-file /path/to/diff.patch \
  --manifest component-manifest.json \
  --max-depth 10
```

### 使用预生成的依赖数据

如果已经有 component-deps-v2 的输出结果，可以直接使用：

```bash
analyzer-ts gitlab impact -i /path/to/project \
  --gitlab-url https://gitlab.example.com \
  --gitlab-token $GITLAB_TOKEN \
  --project-id 123 \
  --mr-id 456 \
  --deps-file /path/to/deps.json
```

## 命令参数

### `gitlab impact`

分析代码变更并发布 MR 评论。

| 参数               | 类型   | 默认值                      | 说明                            |
| ------------------ | ------ | --------------------------- | ------------------------------- |
| `-i, --input`    | string | 必需                        | 项目根目录路径                  |
| `--gitlab-url`   | string | `$CI_SERVER_URL`          | GitLab 实例 URL                 |
| `--gitlab-token` | string | `$GITLAB_TOKEN`           | GitLab 访问令牌                 |
| `--project-id`   | int    | `$CI_PROJECT_ID`          | 项目 ID                         |
| `--mr-id`        | int    | `$CI_MERGE_REQUEST_IID`   | MR IID                          |
| `--diff-source`  | string | `auto`                    | Diff 来源: auto/file/api/diff   |
| `--diff-file`    | string | -                           | 本地 diff 文件路径              |
| `--diff-sha`     | string | -                           | Git diff SHA 范围 (base...head) |
| `--manifest`     | string | `component-manifest.json` | 组件清单路径                    |
| `--deps-file`    | string | -                           | 依赖数据文件路径                |
| `--max-depth`    | int    | `10`                      | 最大传播深度                    |

## Diff 来源模式

| 模式     | 说明               | 使用场景          |
| -------- | ------------------ | ----------------- |
| `auto` | 自动检测           | GitLab CI 环境    |
| `file` | 从本地文件读取     | 本地测试          |
| `api`  | 从 GitLab API 获取 | 需要 GitLab Token |
| `diff` | 执行 git diff 命令 | 本地 Git 仓库     |

## 分析流程

```
1. 解析 Git Diff
   │
   ├──> auto: 优先级 API > git command > 文件
   ├──> file: 从本地 diff 文件解析
   ├──> api: 从 GitLab API 获取 MR diff
   └──> diff: 执行 git diff 命令
   │
2. 运行组件依赖分析 (component-deps-v2)
   │
   ├──> 解析项目 AST
   ├──> 加载 component-manifest.json
   ├──> 构建依赖图和反向依赖图
   └──> 或从 --deps-file 加载预生成的数据
   │
3. 运行影响分析 (impact-analysis)
   │
   ├──> 识别变更的组件
   ├──> BFS 传播影响
   ├──> 评估风险等级
   └──> 生成建议
   │
4. 发布 MR 评论
   │
   ├──> 格式化为 Markdown
   ├──> 查找已有的分析器评论
   ├──> 更新或创建评论
   └──> 包含风险概要、受影响组件、建议
```

## MR 评论格式

```markdown
## 🔍 代码影响分析报告

### 📊 概要

| 指标 | 数值 |
|------|------|
| 变更组件 | 3 |
| 受影响组件 | 12 |
| 高风险 | 2 |
| 中风险 | 5 |
| 低风险 | 5 |

### 🎯 变更组件

#### 📝 Button

- `src/components/Button/index.tsx`
- `src/components/Button/styles.ts`

### 📈 影响范围

#### 🟠 Form (风险: high, 层级: 2)

变更路径:
- Button → Form

#### 🟡 LoginPage (风险: medium, 层级: 3)

变更路径:
- Button → Form → LoginPage

### 💡 建议

- [🟠🧪] **high**: 发现 2 个高风险组件，建议补充单元测试
- [🟡📄] **medium**: 本次变更涉及 3 个组件，建议更新相关文档

---
*由 analyzer-ts 自动生成
```

## 行级 Diff 解析

兼容层支持文件级别（当前 impact-analysis 使用）：

```go
// 文件级变更输入
type ChangeInput struct {
    ModifiedFiles []string `json:"modifiedFiles"`
    AddedFiles    []string `json:"addedFiles"`
    DeletedFiles  []string `json:"deletedFiles"`
}
```

## 配置示例

### component-manifest.json

```json
{
  "meta": {
    "version": "1.0.0",
    "libraryName": "my-ui-lib"
  },
  "components": [
    {
      "name": "Button",
      "scope": ["src/components/Button/**/*"]
    },
    {
      "name": "Form",
      "scope": ["src/components/Form/**/*"]
    }
  ]
}
```

## 开发指南

### 作为独立包使用

```go
import "github.com/Flying-Bird1999/analyzer-ts/pkg/gitlab"

// 获取命令
cmd := gitlab.GetCommand()
rootCmd.AddCommand(cmd)

// 或直接使用集成器
config := &gitlab.GitLabConfig{
    URL:        "https://gitlab.example.com",
    Token:      "glpat-xxxxx",
    ProjectID:  123,
    MRIID:      456,
    DiffSource: "auto",
    MaxDepth:   10,
}
integration := gitlab.NewGitLabIntegration(config)
err := integration.RunAnalysis(ctx, "/path/to/project")
```

### 内部 API

```go
// 创建 GitLab 客户端
client := gitlab.NewClient(baseURL, token)

// 解析 diff
parser := gitlab.NewDiffParser(projectRoot)
lineSet, err := parser.ParseFromGit("baseSHA", "HEAD")

// 格式化结果
formatter := gitlab.NewFormatter(gitlab.CommentStyleDetailed)
markdown, err := formatter.FormatImpactResult(result)
```

## 与 types-convertor-app 对比

| 特性      | types-convertor-app   | analyzer-ts (pkg/gitlab) |
| --------- | --------------------- | ------------------------ |
| 语言      | TypeScript + Go addon | 纯 Go                    |
| 分析粒度  | AST 节点级别          | 组件级别                 |
| Diff 解析 | 文件级                | 行级 + 兼容层            |
| 命令模式  | 独立命令              | 集成到 analyzer-ts       |
| 部署      | 需要 node_modules     | 单一二进制               |

## 故障排查

### 常见问题

**Q: 提示 "gitlab-url is required"**

A: 请确保设置了 `--gitlab-url` 参数或 `$CI_SERVER_URL` 环境变量。

**Q: 提示 "deps-file is required"**

A: 请提供 `--deps-file` 参数，或确保 `component-manifest.json` 存在以便运行 component-deps-v2。

**Q: MR 评论没有更新**

A: 检查 GitLab Token 是否有 `api` 权限，确认 Project ID 和 MR IID 正确。

**Q: Diff 解析失败**

A: 确保 `GIT_DEPTH=0` 设置在 CI 配置中，以便获取完整的 git 历史。

## 许可证

与 analyzer-ts 项目保持一致。
