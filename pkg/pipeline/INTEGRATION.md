# pkg/pipeline 接入文档

本文档帮助业务方快速接入代码影响分析能力。

## 快速开始

### 1. 确认项目类型

首先确认你的项目类型：

| 项目类型 | 特征 | 需要准备 |
|---------|------|---------|
| 普通项目 | 无组件清单 | 无特殊准备 |
| 组件库项目 | 有明确的组件定义 | 需要创建 `component-manifest.json` |

### 2. 准备工作

#### 2.1 确认项目路径

```bash
# 项目根目录（包含 package.json 的目录）
PROJECT_ROOT="/path/to/your/project"

# Git 仓库根目录
# - 如果是单体仓库：通常等于 PROJECT_ROOT
# - 如果是 monorepo：通常是 monorepo 的根目录
GIT_ROOT="/path/to/git/repository"
```

#### 2.2 （可选）创建组件清单

如果你的项目是组件库，创建组件清单：

```json
// .analyzer/component-manifest.json
{
  "version": "1.0",
  "components": [
    {
      "name": "Button",
      "entry": "src/components/Button/index.tsx",
      "dependencies": {
        "Icon": "src/components/Icon/index.tsx"
      }
    },
    {
      "name": "Input",
      "entry": "src/components/Input/index.tsx",
      "dependencies": {}
    }
  ]
}
```

**组件清单说明：**

| 字段 | 必填 | 说明 |
|------|------|------|
| `name` | 是 | 组件名称 |
| `entry` | 是 | 组件入口文件（相对于项目根） |
| `dependencies` | 否 | 组件依赖的其他组件（`组件名: 入口文件`） |

### 3. 选择接入方式

#### 方式一：使用 CLI（推荐）

最简单的方式，适合快速验证和 CI/CD 集成：

```bash
# 安装
go install github.com/Flying-Bird1999/analyzer-ts/cmd/analyzer-ts@latest

# 使用 diff 文件
analyzer-ts impact \
  --project-root /path/to/project \
  --diff-file /path/to/changes.patch \
  --output impact-result.json

# 使用 git diff
analyzer-ts impact \
  --project-root /path/to/project \
  --git-diff "HEAD~1 HEAD" \
  --output impact-result.json

# 使用 diff 字符串
analyzer-ts impact \
  --project-root /path/to/project \
  --diff-string "$(git diff HEAD~1 HEAD)" \
  --output impact-result.json
```

#### 方式二：Go 代码集成

适合需要自定义处理逻辑的场景：

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/Flying-Bird1999/analyzer-ts/pkg/pipeline"
)

func main() {
    // 配置管道
    config := &pipeline.GitLabPipelineConfig{
        DiffSource:   pipeline.DiffSourceFile,
        DiffFile:     "/path/to/mr.patch",
        ProjectRoot:  "/path/to/project",
        GitRoot:      "/path/to/git/root",  // monorepo 场景需要
        ManifestPath: "/path/to/component-manifest.json",  // 可选
        MaxDepth:     10,
    }

    // 创建上下文
    ctx := context.Background()
    analysisCtx := pipeline.NewAnalysisContext(ctx, config.ProjectRoot, nil)

    // 执行
    pipe := pipeline.NewGitLabPipeline(config)
    result, err := pipe.Execute(analysisCtx)
    if err != nil {
        fmt.Printf("分析失败: %v\n", err)
        os.Exit(1)
    }

    // 获取结果
    impactResult, _ := result.GetResult("影响分析（文件级）")
    if impact, ok := impactResult.(*pipeline.ImpactAnalysisResult); ok {
        fmt.Printf("变更文件: %d\n", impact.FileResult.Meta.ChangedFileCount)
        fmt.Printf("受影响文件: %d\n", impact.FileResult.Meta.ImpactFileCount)

        // 处理受影响文件
        for _, file := range impact.FileResult.Impact {
            fmt.Printf("  - %s (层级 %d)\n", file.Path, file.ImpactLevel)
        }
    }
}
```

## CI/CD 集成示例

### GitLab CI

```yaml
# .gitlab-ci.yml
analyze:
  stage: test
  script:
    # 获取 MR 的 diff
    DIFF=$(git diff --diff-filter=d origin/main...HEAD)

    # 执行影响分析
    analyzer-ts impact \
      --project-root ${CI_PROJECT_DIR} \
      --git-root ${CI_PROJECT_DIR} \
      --diff-string "$DIFF" \
      --output impact-report.json

    # 解析结果（可选）
    IMPACT_COUNT=$(cat impact-report.json | jq '.fileAnalysis.meta.impactFileCount')
    echo "受影响文件数: $IMPACT_COUNT"

    # 如果影响范围过大，可以阻止合并
    if [ "$IMPACT_COUNT" -gt 20 ]; then
      echo "⚠️  影响范围过大，建议人工审查"
      exit 1
    fi

  artifacts:
    paths:
      - impact-report.json
  only:
    - merge_requests
```

### GitHub Actions

```yaml
# .github/workflows/impact-analysis.yml
name: Impact Analysis

on:
  pull_request:
    types: [opened, synchronize]

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0  # 获取完整历史用于 diff

      - name: Install analyzer-ts
        run: go install github.com/Flying-Bird1999/analyzer-ts/cmd/analyzer-ts@latest

      - name: Run Impact Analysis
        run: |
          DIFF=$(git diff --diff-filter=d origin/main...HEAD)
          analyzer-ts impact \
            --project-root ${{ github.workspace }} \
            --diff-string "$DIFF" \
            --output impact-report.json

      - name: Upload Report
        uses: actions/upload-artifact@v3
        with:
          name: impact-report
          path: impact-report.json

      - name: Comment PR
        uses: actions/github-script@v6
        with:
          script: |
            const fs = require('fs');
            const report = JSON.parse(fs.readFileSync('impact-report.json', 'utf8'));
            const impactCount = report.fileAnalysis.meta.impactFileCount;

            const body = `## 📊 影响分析报告
            - 变更文件: ${report.fileAnalysis.meta.changedFileCount}
            - 受影响文件: ${impactCount}`;

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: body
            });
```

### npm scripts 集成

```json
// package.json
{
  "scripts": {
    "analyze:impact": "analyzer-ts impact --project-root $(pwd) --git-diff \"HEAD~1 HEAD\"",
    "analyze:impact:file": "analyzer-ts impact --project-root $(pwd) --diff-file ./changes.patch --output result.json",
    "precommit": "analyzer-ts impact --project-root $(pwd) --git-diff \"HEAD\" --format summary"
  }
}
```

## 输出结果说明

### JSON 格式（默认）

```json
{
  "meta": {
    "projectRoot": "/path/to/project",
    "analyzedAt": "2024-01-01T00:00:00Z",
    "inputSource": "diff 文件: /path/to/changes.patch"
  },
  "input": {
    "files": ["src/components/Button/Button.tsx"]
  },
  "fileAnalysis": {
    "meta": {
      "totalFileCount": 100,
      "changedFileCount": 1,
      "impactFileCount": 7
    },
    "changes": [
      {
        "path": "src/components/Button/Button.tsx",
        "type": "modified",
        "symbolCount": 3
      }
    ],
    "impact": [
      {
        "path": "src/components/Form/Form.tsx",
        "impactLevel": 1,
        "impactType": "internal",
        "changePaths": ["src/components/Button/Button.tsx"]
      }
    ]
  },
  "componentAnalysis": {
    "meta": {
      "totalComponentCount": 10,
      "changedComponentCount": 1,
      "impactComponentCount": 8
    },
    "changes": [
      {
        "name": "Button",
        "changedFiles": ["src/components/Button/Button.tsx"]
      }
    ],
    "impact": [
      {
        "name": "Form",
        "impactLevel": 2,
        "changePaths": ["src/components/Button/Button.tsx"]
      }
    ]
  }
}
```

### 字段说明

#### 文件级分析

| 字段 | 说明 |
|------|------|
| `changedFileCount` | 直接变更的文件数 |
| `impactFileCount` | 间接受影响的文件数 |
| `impactLevel` | 影响层级（1 = 直接依赖，2+ = 间接依赖） |
| `impactType` | 影响类型（`internal` = 项目内部，`external` = 外部依赖） |

#### 组件级分析

| 字段 | 说明 |
|------|------|
| `changedComponentCount` | 直接变更的组件数 |
| `impactComponentCount` | 间接受影响的组件数 |
| `impactLevel` | 组件影响层级 |

## 常见场景

### 场景一：Monorepo 项目

```bash
# 目录结构
# /repo
#   ├── packages/
#   │   └── my-package/     # 项目根
#   └── .git/               # Git 仓库根

analyzer-ts impact \
  --project-root /repo/packages/my-package \
  --git-root /repo \
  --diff-file /path/to/changes.patch
```

**注意**：diff 文件中的路径必须相对于 `git-root`。

### 场景二：前端 npm 包集成

```javascript
// scripts/analyze-impact.js
const { execSync } = require('child_process');

function analyzeImpact(options = {}) {
  const {
    projectRoot = process.cwd(),
    diffString,
    outputFile
  } = options;

  const cmd = `analyzer-ts impact \
    --project-root ${projectRoot} \
    --diff-string "${diffString}" \
    --output ${outputFile}`;

  return execSync(cmd, { encoding: 'utf-8' });
}

// 使用
const diff = execSync('git diff HEAD~1 HEAD', { encoding: 'utf-8' });
const result = analyzeImpact({
  diffString: diff,
  outputFile: './impact-report.json'
});

console.log('分析完成！');
```

### 场景三：只获取简要摘要

```bash
analyzer-ts impact \
  --project-root /path/to/project \
  --git-diff "HEAD~1 HEAD" \
  --format summary
```

输出：
```
代码影响分析结果
==================

变更文件: 1
受影响文件: 7
变更组件: 1
受影响组件: 8

变更的文件:
  - src/components/Button/Button.tsx

受影响的文件:
  - src/components/Form/Form.tsx (层级 1)
  - src/components/Table/Table.tsx (层级 1)
  ...
```

## 业务方需要做什么？

### 必做事项

1. **确定项目路径**
   - 确认项目根目录（包含 `package.json` 或 `tsconfig.json`）
   - 确认 Git 仓库根目录（monorepo 需特别确认）

2. **准备 diff 数据**
   - 选择合适的 diff 输入方式（文件、字符串、或 git 命令）

3. **配置输出**
   - 决定输出格式（JSON、Pretty、Summary）

### 可做事项

1. **创建组件清单**（组件库项目）
   - 在项目根创建 `.analyzer/component-manifest.json`
   - 定义组件及其依赖关系

2. **配置分析深度**
   - 根据项目大小调整 `--max-depth`（默认 10）

3. **集成到 CI/CD**
   - 添加影响分析步骤到 pipeline
   - 配置影响范围阈值检查

## 故障排查

### 问题：找不到文件

```
Error: 项目根目录不存在
```

**解决**：确认 `--project-root` 使用绝对路径

### 问题：解析结果为空

```
发现 0 个文件，0 行变更
```

**解决**：
1. 检查 diff 文件格式是否正确
2. monorepo 场景检查 `--git-root` 配置
3. diff 中的路径必须是相对于 git root 的路径

### 问题：符号分析为空

```
没有检测到符号变更
```

**解决**：
1. 确认变更文件包含实际代码变更（不只是注释或空行）
2. 检查变更是否影响导出的符号

## 获取帮助

- CLI 帮助：`analyzer-ts impact --help`
- 查看示例：`examples/impact/`
- 查看测试用例：`pkg/pipeline/scenario_test.go`
