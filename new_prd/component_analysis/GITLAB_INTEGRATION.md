# GitLab 集成技术方案

> **版本**: v2.0.0
> **状态**: 设计阶段
> **更新日期**: 2024-01-31
> **目标**: 独立的 GitLab 集成包，提供 MR 评论能力

---

## 执行摘要

创建独立的 `pkg/gitlab` 包，为 impact-analysis 提供 GitLab MR 评论能力。

| 功能 | 状态 | 说明 |
|------|------|------|
| Git diff 解析 | 🆕 设计中 | 支持文件/API/CI 三种模式 |
| JSON 报告 | 🆕 设计中 | 复用现有 impact-analysis 输出 |
| MR 评论发布 | 🆕 设计中 | 格式化 JSON 为 Markdown 评论 |
| AI 代码审查 | 🆕 Phase 3 | Breaking Changes 检测 |

---

## 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                     GitLab CI/CD Pipeline                     │
│  触发 MR → CI Job 执行 analyzer-ts → 分析结果               │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                      1. Git Diff 解析                         │
│  pkg/gitlab/diff_parser → 提取变更文件列表                    │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                      2. 依赖分析                               │
│  component-deps-v2 → 构建组件依赖图                           │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                      3. 影响传播                               │
│  impact-analysis → BFS 计算影响范围                           │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                      4. JSON 输出                              │
│  impact-analysis → 现有 JSON 格式输出                         │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                      5. MR 评论发布                           │
│  pkg/gitlab/mr_service → JSON 转 Markdown → MR 评论           │
└─────────────────────────────────────────────────────────────┘
```

---

## 文件结构

### 新建文件

```
pkg/gitlab/
├── client.go              # GitLab API 客户端
├── types.go               # GitLab 类型定义
├── mr_service.go          # MR 服务
├── diff_parser.go         # Git diff 解析器
├── formatter.go           # JSON → Markdown 转换
└── integration.go         # 与 impact-analysis 集成
```

### 修改文件

```
cmd/root.go                   # 注册 gitlab 子命令
main.go                       # 导入 pkg/gitlab 包
```

**命令注册方式**：
```go
// cmd/root.go
import (
    "github.com/Flying-Bird1999/analyzer-ts/pkg/gitlab"
)

func init() {
    RootCmd.AddCommand(gitlab.GetCommand())
}
```

---

## 核心模块

### 0. pkg/gitlab/command.go

提供 cobra.Command 接口

```go
func GetCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "gitlab",
        Short: "GitLab 集成命令",
        RunE:  runGitLabCommand,
    }
    // ... 参数定义
    return cmd
}
```

### 1. pkg/gitlab/client.go

GitLab API 低层客户端

```go
type Client struct {
    baseURL    string
    token      string
    httpClient *http.Client
}

// API 方法
GetMergeRequest(ctx, projectID, mrIID) (*MergeRequest, error)
GetMergeRequestDiff(ctx, projectID, mrIID) ([]DiffFile, error)
CreateMRComment(ctx, projectID, mrIID, body) error
UpdateMRComment(ctx, projectID, mrIID, noteID, body) error
ListMRComments(ctx, projectID, mrIID) ([]Comment, error)
```

### 2. pkg/gitlab/diff_parser.go

解析 git diff 输出（行级别）

```go
// ChangedLineSetOfFiles 跟踪每个文件变更的行号
type ChangedLineSetOfFiles map[string]map[int]bool

// DiffParser 解析 git diff
type DiffParser struct {
    baseDir string
}

// ParseDiffOutput 解析 diff 输出，提取变更行号
// 参考 merge-request-impact-reviewer/git-diff-plugin.ts
ParseDiffOutput(diffOutput string) (ChangedLineSetOfFiles, error)

// ParseFromGit 执行 git diff 并解析
ParseFromGit(baseSHA, headSHA string) (ChangedLineSetOfFiles, error)

// ParseDiffFiles 解析 GitLab API diff 格式
ParseDiffFiles(diffFiles []DiffFile) (ChangedLineSetOfFiles, error)

// GetChangedFiles 提取变更文件列表（兼容现有接口）
GetChangedFiles(lineSet ChangedLineSetOfFiles) []string
```

### 3. pkg/gitlab/mr_service.go

MR 高层服务

```go
type MRService struct {
    client    *Client
    projectID int
    mrIID     int
}

FindAnalyzerComment(ctx) (*Comment, error)
PostImpactComment(ctx, result *ImpactAnalysisResult) error
DeleteOldComments(ctx) error
```

### 4. pkg/gitlab/formatter.go

JSON 转 Markdown

```go
type Formatter struct {
    style CommentStyle
}

FormatImpactResult(result *ImpactAnalysisResult) (string, error)
FormatSummary(result *ImpactAnalysisResult) string
FormatRiskTable(result *ImpactAnalysisResult) string
```

### 5. pkg/gitlab/integration.go

与 impact-analysis 集成

```go
type GitLabIntegration struct {
    client     *Client
    mrService  *MRService
    diffParser *DiffParser
    formatter  *Formatter
}

RunAnalysis(ctx, config) error
```

---

## 配置设计

### 独立命令

新增 `gitlab` 子命令：

```bash
analyzer-ts gitlab impact [options]
```

### CLI 参数

```bash
# GitLab 连接参数
--gitlab-url string           # GitLab 实例 URL (默认: $CI_SERVER_URL)
--gitlab-token string         # GitLab Token (默认: $GITLAB_TOKEN)
--project-id int             # 项目 ID (默认: $CI_PROJECT_ID)
--mr-id int                   # MR IID (默认: $CI_MERGE_REQUEST_ID)

# Diff 来源
--diff-source string          # diff/api/file (默认: auto-detect)
--diff-file string            # 本地 diff 文件路径
--diff-sha string             # 指定 diff 的 SHA 范围

# 分析参数
--manifest string             # component-manifest.json 路径
--deps-file string            # 依赖数据文件路径
--max-depth int               # 最大传播深度 (默认: 10)
```

### 环境变量（CI 自动检测）

```bash
# GitLab CI 内置变量
CI_MERGE_REQUEST_ID
CI_MERGE_REQUEST_DIFF_BASE_SHA
CI_PROJECT_ID
CI_SERVER_URL
GITLAB_TOKEN

# 自定义变量
ANALYZER_MANIFEST_PATH       # component-manifest.json 路径
```

### 使用示例

```bash
# GitLab CI 模式（自动检测环境变量）
analyzer-ts gitlab impact -i .

# 本地测试模式
analyzer-ts gitlab impact -i /path/to/project \
  --gitlab-url https://gitlab.example.com \
  --gitlab-token $GITLAB_TOKEN \
  --project-id 123 \
  --mr-id 456 \
  --diff-file /path/to/diff.patch

# API 模式
analyzer-ts gitlab impact -i . \
  --diff-source api
```

---

## CI/CD 集成

### .gitlab-ci.yml 示例

```yaml
stages:
  - analyze

impact-analysis:
  stage: analyze
  rules:
    - if: '$CI_PIPELINE_SOURCE == "merge_request_event"'
  variables:
    GIT_DEPTH: 0
  script:
    - analyzer-ts gitlab impact -i .
```

### MR 评论格式

```markdown
## 🔍 代码影响分析报告

### 📊 概要

| 指标 | 数值 |
|------|------|
| 变更组件 | 2 |
| 受影响组件 | 5 |
| 高风险 | 0 |
| 中风险 | 1 |

### 🎯 变更组件

#### Button (modified)
- `src/components/Button/Button.tsx`

### 📈 影响范围

#### Input (风险: low, 层级: 1)
- 变更路径: Button → Input

#### Select (风险: medium, 层级: 2)
- 变更路径: Button → Input → Select
- 变更路径: Button → Select

### 💡 建议

- [test] 发现 1 个中风险组件，建议补充单元测试

---

*由 analyzer-ts 自动生成*
```

---

## 实施阶段

### Phase 1: GitLab 基础集成 (1 周)

- [ ] pkg/gitlab/client.go - API 客户端
- [ ] pkg/gitlab/types.go - 类型定义
- [ ] pkg/gitlab/diff_parser.go - diff 解析
- [ ] pkg/gitlab/mr_service.go - MR 服务
- [ ] pkg/gitlab/formatter.go - Markdown 格式化
- [ ] pkg/gitlab/integration.go - 集成逻辑
- [ ] cmd/gitlab.go - 命令行接口
- [ ] 单元测试

**交付**: 基本的 MR 评论功能

### Phase 2: 完善集成 (3-5 天)

- [ ] CI 环境变量自动检测
- [ ] 错误处理和重试
- [ ] 集成测试
- [ ] 文档完善

**交付**: 完整的 CI/CD 工作流

### Phase 3: AI 集成（可选）(1 周)

- [ ] pkg/gitlab/ai/reviewer.go - AI 审查
- [ ] pkg/gitlab/ai/openai_client.go - OpenAI 客户端
- [ ] Breaking Changes 检测

**交付**: AI 增强的代码审查

---

## 关键文件清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `pkg/gitlab/command.go` | 新建 | 提供 cobra.Command 接口 |
| `pkg/gitlab/client.go` | 新建 | GitLab API 客户端 |
| `pkg/gitlab/types.go` | 新建 | GitLab 类型定义 |
| `pkg/gitlab/mr_service.go` | 新建 | MR 服务 |
| `pkg/gitlab/diff_parser.go` | 新建 | Git diff 解析器 |
| `pkg/gitlab/formatter.go` | 新建 | Markdown 格式化 |
| `pkg/gitlab/integration.go` | 新建 | 集成逻辑 |
| `cmd/root.go` | 修改 | 注册 gitlab 子命令 |
| `main.go` | 修改 | 导入 pkg/gitlab |

---

## 验证计划

### 单元测试
```bash
go test ./pkg/gitlab/... -v
```

### 集成测试
```bash
# 本地测试（模拟 GitLab 环境）
export CI_SERVER_URL="https://gitlab.example.com"
export GITLAB_TOKEN="glpat-xxxxx"
analyzer-ts gitlab impact -i testdata/test_project \
  --diff-file test.diff \
  --project-id 123 \
  --mr-id 456
```

### CI/CD 测试
1. 推送到测试项目
2. 创建 MR
3. 验证自动分析触发
4. 检查 MR 评论内容

---

## 风险与依赖

### 风险
- GitLab API 版本兼容性
- 大型项目 diff 解析性能
- Token 权限配置

### 依赖
- GitLab API v4
- Go 1.22+
- 现有 component-deps-v2 和 impact-analysis 插件

---

## 与 merge-request-impact-reviewer 对比

### Git Diff 处理对比

| 方面 | merge-request-impact-reviewer | analyzer-ts (当前) | analyzer-ts (未来) |
|------|-------------------------------|------------------|------------------|
| diff 精度 | **行级别** ✨ | 文件级别 | 行级别 ✨ |
| 数据结构 | `ChangedLineSetOfFiles` | `ChangeInput` | `ChangedLineSetOfFiles` |
| 解析方式 | **正则匹配 hunk + 行号** ✨ | 文件列表匹配 | 正则匹配 hunk + 行号 |
| 变更追踪 | **知道具体哪些行变了** ✨ | 知道哪些文件变了 | 知道具体哪些行变了 |

### 渐进式设计策略

```go
// ===== 当前实现：diff_parser 提供精确解析 =====

// 行级别数据结构（与 merge-request-impact-reviewer 一致）
type ChangedLineSetOfFiles map[string]map[int]bool

// DiffParser 精确解析 diff
type DiffParser struct {
    baseDir string
}

ParseDiffOutput(diffOutput string) (ChangedLineSetOfFiles, error)
ParseFromGit(baseSHA, headSHA string) (ChangedLineSetOfFiles, error)

// ===== 兼容层：转换为文件级别 =====

// GetChangedFiles 将行级别转换为文件列表（兼容现有 impact-analysis）
func (p *DiffParser) GetChangedFiles(lineSet ChangedLineSetOfFiles) *ChangeInput {
    files := &ChangeInput{
        ModifiedFiles: []string{},
        AddedFiles:    []string{},
        DeletedFiles:  []string{},
    }

    for filePath, lines := range lineSet {
        if len(lines) > 0 {
            files.ModifiedFiles = append(files.ModifiedFiles, filePath)
        }
    }

    return files
}

// ===== 未来优化：impact-analysis 支持行级别 =====

// Phase 2: 扩展 impact-analysis 支持行级别变更输入
// type ChangeInputV2 struct {
//     ChangedFiles map[string]*FileChanges
// }
//
// type FileChanges struct {
//     AddedLines   []int
//     ModifiedLines []int
//     DeletedLines []int
// }
```

### 设计优势

1. **diff_parser 精确实现**：与 merge-request-impact-reviewer 保持一致
2. **当前可用**：通过 GetChangedFiles() 兼容现有 impact-analysis
3. **未来可扩展**：impact-analysis 可升级到行级别分析
4. **降低风险**：分步实施，每步都可验证

---

## 与 types-convertor-app 对比

| 特性 | types-convertor-app | analyzer-ts (Go) |
|------|---------------------|------------------|
| 语言 | TypeScript + Go addon | 纯 Go |
| 分析粒度 | AST 节点级别 | 组件级别 |
| GitLab 集成 | ✅ 已完成 | 🆕 设计中 |
| 报告格式 | HTML + Artifact | JSON + Markdown |
| AI 集成 | ✅ 已完成 | 🆕 Phase 3 |
| 部署 | 需要 node_modules | 单一二进制 |

**选择 Go 的优势**：
- 更简单的 CI/CD 集成
- 更好的性能和内存控制
- 无需 Node.js 运行时
- 更容易维护和调试

---

## 成功标准

1. 能从 GitLab MR 自动解析 diff 并分析影响
2. Markdown 评论清晰展示影响范围和风险
3. 一行配置即可集成到 GitLab CI
4. 中型项目（<1000 组件）分析时间 <30s
5. API 失败时有清晰的错误提示
