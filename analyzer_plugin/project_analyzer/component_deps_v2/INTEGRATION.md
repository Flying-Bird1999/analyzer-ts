# component-deps-v2 接入文档

本文档帮助业务方快速接入 `component-deps-v2` 组件依赖分析能力。

- **V1 vs V2**: `component-deps` (V1) 通过入口文件自动推断组件边界，适合简单项目；`component-deps-v2` (V2) 通过配置文件显式声明组件，适合复杂组件库项目
- **作用域推断**: 基于入口文件所在目录自动推断组件作用域（如 `src/Button/index.tsx` → `src/Button/**`），无需手动配置
- **配置极简**: 只需声明 `components` 数组，每个组件包含 `name` 和 `entry` 两个字段

---

## 目录

1. [快速开始](#快速开始)
2. [配置文件](#配置文件)
3. [使用方式](#使用方式)
4. [结果说明](#结果说明)
5. [集成示例](#集成示例)
6. [常见问题](#常见问题)

---

## 快速开始

### 1. 确认项目类型

`component-deps-v2` 适用于以下项目类型：

| 项目类型 | 特征 | 是否适用 |
|---------|------|---------|
| 组件库项目 | 有明确的组件定义和入口 | ✅ 强烈推荐 |
| Monorepo | 多个包，每个包有独立组件 | ✅ 推荐 |
| 普通项目 | 无明确组件边界 | ❌ 建议使用 `component-deps` |

### 2. 准备工作

#### 2.1 创建组件配置文件

在项目根目录创建 `component-manifest.json`：

```bash
# 推荐位置
project_root/.analyzer/component-manifest.json

# 或直接放在根目录
project_root/component-manifest.json
```

#### 2.2 编写配置内容

```json
{
  "components": [
    {
      "name": "Button",
      "entry": "src/components/Button/index.tsx"
    },
    {
      "name": "Input",
      "entry": "src/components/Input/index.tsx"
    },
    {
      "name": "Select",
      "entry": "src/components/Select/index.tsx"
    }
  ]
}
```

---

## 配置文件

### 完整配置结构

```json
{
  "components": [
    {
      "name": "Button",
      "entry": "src/components/Button/index.tsx"
      // scope 自动推断为: src/components/Button/**
    }
  ],
  "rules": {
    "ignorePatterns": [
      "**/*.test.tsx",
      "**/*.spec.tsx",
      "**/node_modules/**"
    ]
  }
}
```

### 字段说明

#### components (必填)

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | string | 组件名称（唯一标识符） |
| `entry` | string | 组件入口文件（相对于项目根目录） |

**作用域推断规则**：
- `entry` = `src/components/Button/index.tsx`
- `scope` = `src/components/Button/**`（自动推断，无需显式配置）

#### rules (可选)

| 字段 | 类型 | 说明 |
|------|------|------|
| `ignorePatterns` | string[] | 忽略的文件模式（glob 格式） |

### 配置示例

#### 简单组件库

```json
{
  "components": [
    {"name": "Button", "entry": "src/Button.tsx"}
  ]
}
```

#### 大型组件库

```json
{
  "components": [
    {"name": "Button", "entry": "src/components/Button/index.tsx"},
    {"name": "Input", "entry": "src/components/Input/index.tsx"},
    {"name": "Select", "entry": "src/components/Select/index.tsx"},
    {"name": "Table", "entry": "src/components/Table/index.tsx"},
    {"name": "Form", "entry": "src/components/Form/index.tsx"},
    {"name": "Modal", "entry": "src/components/Modal/index.tsx"}
  ]
}
```

---

## 使用方式

### 方式一：CLI 命令（推荐）

#### 基本用法

```bash
analyzer-ts analyze component-deps-v2 \
  -i /path/to/project \
  -p "component-deps-v2.manifest=.analyzer/component-manifest.json"
```

#### 使用绝对路径

```bash
analyzer-ts analyze component-deps-v2 \
  -i /Users/user/my-project \
  -p "component-deps-v2.manifest=/absolute/path/to/component-manifest.json"
```

#### CLI 参数说明

| 参数 | 说明 | 示例 |
|------|------|------|
| `-i, --input` | 项目根目录（绝对路径） | `/Users/user/my-project` |
| `-o, --output` | 输出目录（可选，默认当前目录） | `/Users/user/output` |
| `-x, --exclude` | 排除的文件模式（可选） | `node_modules/**` |
| `-p, --param` | 分析器参数 | `component-deps-v2.manifest=...` |

### 方式二：Go 代码集成

```go
package main

import (
    "fmt"
    "os"

    component_deps_v2 "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/component_deps_v2"
    projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

func main() {
    // 创建分析器实例
    analyzer := &component_deps_v2.ComponentDepsV2Analyzer{}

    // 配置分析器
    params := map[string]string{
        "manifest": ".analyzer/component-manifest.json",
    }
    if err := analyzer.Configure(params); err != nil {
        fmt.Printf("配置失败: %v\n", err)
        os.Exit(1)
    }

    // 创建项目上下文（需要先解析项目）
    ctx := &projectanalyzer.ProjectContext{
        ProjectRoot:   "/path/to/project",
        Exclude:       []string{"node_modules/**"},
        IsMonorepo:    false,
        ParsingResult: parsingResult, // 从 projectParser 获取
    }

    // 执行分析
    result, err := analyzer.Analyze(ctx)
    if err != nil {
        fmt.Printf("分析失败: %v\n", err)
        os.Exit(1)
    }

    // 输出结果
    jsonData, _ := result.ToJSON(true)
    fmt.Println(string(jsonData))
}
```

### 方式三：结合 impact-analysis 使用

```bash
# 步骤 1: 生成组件依赖数据
analyzer-ts analyze component-deps-v2 \
  -i /path/to/project \
  -p "component-deps-v2.manifest=.analyzer/component-manifest.json" \
  -o /tmp/analyzer-output

# 步骤 2: 使用依赖数据进行影响分析
analyzer-ts impact \
  --project-root /path/to/project \
  --diff-file changes.patch \
  --deps-file /tmp/analyzer-output/test_project_analyzer_data.json \
  --output impact-result.json
```

---

## 结果说明

### 输出结构

```json
{
  "component-deps-v2": {
    "meta": {
      "componentCount": 3
    },
    "components": {
      "Button": {
        "name": "Button",
        "entry": "src/components/Button/index.tsx",
        "dependencies": []
      },
      "Input": {
        "name": "Input",
        "entry": "src/components/Input/index.tsx",
        "dependencies": ["Button"]
      },
      "Select": {
        "name": "Select",
        "entry": "src/components/Select/index.tsx",
        "dependencies": ["Button", "Input"]
      }
    },
    "depGraph": {
      "Button": [],
      "Input": ["Button"],
      "Select": ["Button", "Input"]
    },
    "revDepGraph": {
      "Button": ["Input", "Select"],
      "Input": ["Select"],
      "Select": []
    }
  }
}
```

### 字段说明

#### meta

| 字段 | 说明 |
|------|------|
| `componentCount` | 组件总数 |

#### components

| 字段 | 说明 |
|------|------|
| `name` | 组件名称 |
| `entry` | 组件入口文件 |
| `dependencies` | 该组件依赖的其他组件列表 |

#### depGraph (正向依赖图)

- **key**: 组件名称
- **value**: 该组件直接依赖的组件列表

**用途**: 查找组件的上游依赖

```json
"depGraph": {
  "Select": ["Button", "Input"]  // Select 依赖 Button 和 Input
}
```

#### revDepGraph (反向依赖图)

- **key**: 组件名称
- **value**: 依赖该组件的其他组件列表

**用途**: 查找组件的下游影响范围

```json
"revDepGraph": {
  "Button": ["Input", "Select"]  // Button 被 Input 和 Select 依赖
}
```

### 依赖关系可视化

```
Button (基础组件)
  │
  ├─→ Input (依赖 Button)
  │     │
  │     └─→ Select (依赖 Input, Button)
  │
  └─→ Select (直接依赖 Button)
```

---

## 集成示例

### GitLab CI/CD 集成

```yaml
# .gitlab-ci.yml
stages:
  - analyze

component-deps-analysis:
  stage: analyze
  image: golang:1.22
  script:
    - go install github.com/Flying-Bird1999/analyzer-ts/cmd/analyzer-ts@latest
    - analyzer-ts analyze component-deps-v2 \
        -i $CI_PROJECT_DIR \
        -p "component-deps-v2.manifest=.analyzer/component-manifest.json" \
        -o $CI_PROJECT_DIR/artifacts
  artifacts:
    paths:
      - artifacts/*.json
    expire_in: 1 week
```

### 结合 GitLab MR 评论

```yaml
# .gitlab-ci.yml
impact-analysis:
  stage: analyze
  rules:
    - if: '$CI_PIPELINE_SOURCE == "merge_request_event"'
  script:
    # 分析组件依赖
    - analyzer-ts analyze component-deps-v2 \
        -i . \
        -p "component-deps-v2.manifest=.analyzer/component-manifest.json" \
        -o /tmp

    # 执行影响分析
    - analyzer-ts impact \
        --project-root . \
        --git-diff "$CI_MERGE_REQUEST_DIFF_BASE_SHA $CI_COMMIT_SHA" \
        --deps-file /tmp/*_analyzer_data.json \
        --output /tmp/impact-result.json

    # 发布 MR 评论（需要 GitLab 集成）
    - analyzer-ts gitlab impact \
        --result-file /tmp/impact-result.json
```

### GitHub Actions 集成

```yaml
# .github/workflows/component-analysis.yml
name: Component Dependency Analysis

on:
  pull_request:
    types: [opened, synchronize]

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Install analyzer-ts
        run: go install github.com/Flying-Bird1999/analyzer-ts/cmd/analyzer-ts@latest

      - name: Analyze component dependencies
        run: |
          analyzer-ts analyze component-deps-v2 \
            -i $GITHUB_WORKSPACE \
            -p "component-deps-v2.manifest=.analyzer/component-manifest.json" \
            -o $GITHUB_WORKSPACE/artifacts

      - name: Upload results
        uses: actions/upload-artifact@v3
        with:
          name: component-deps
          path: artifacts/*.json
```

---

## 常见问题

### Q1: component-deps 和 component-deps-v2 如何选择？

| 特性 | component-deps (V1) | component-deps-v2 (V2) |
|------|---------------------|------------------------|
| 配置方式 | 通过入口文件自动推断 | 通过配置文件显式声明 |
| 适用场景 | 简单项目、快速验证 | 复杂组件库、需要精确控制 |
| 组件识别 | 基于 import 关系 | 基于 manifest 声明 |
| 灵活性 | 较低 | 高 |

**建议**: 组件库项目使用 V2，普通项目使用 V1。

### Q2: 如何处理共享代码？

共享代码（如 utils、hooks）不会被识别为组件，它们的作用是支持组件实现。

**解决方案**: 将共享代码放在组件外部，或单独创建一个 "Shared" 组件：

```json
{
  "components": [
    {"name": "Button", "entry": "src/components/Button/index.tsx"},
    {"name": "Shared", "entry": "src/shared/index.ts"}
  ]
}
```

### Q3: 如何处理跨目录的组件？

组件作用域基于 `entry` 所在目录自动推断。如果组件文件分散，可以：

1. 将所有文件放在同一目录下
2. 使用多个 entry 文件（如果组件有多个入口）

### Q4: 配置文件路径找不到？

检查：
1. 路径是否正确（支持相对路径和绝对路径）
2. 文件是否存在
3. JSON 格式是否正确

```bash
# 验证配置文件
cat .analyzer/component-manifest.json | jq .
```

### Q5: 如何验证分析结果？

```bash
# 查看控制台输出
analyzer-ts analyze component-deps-v2 \
  -i . \
  -p "component-deps-v2.manifest=.analyzer/component-manifest.json" | jq .

# 检查依赖关系是否正确
jq '.["component-deps-v2"].depGraph' artifacts/*.json
```

### Q6: 大型项目性能如何？

针对大型项目（>100 组件）：

1. **使用排除模式**: 减少需要解析的文件
   ```bash
   -x "node_modules/**" -x "**/*.test.tsx"
   ```

2. **分离分析**: 按模块分组分析
   ```json
   // 只分析变更的模块
   "components": [
     {"name": "Button", "entry": "src/modules/Button/index.tsx"},
     {"name": "Input", "entry": "src/modules/Input/index.tsx"}
   ]
   ```

---

## 相关文档

- [技术方案概述](../README.md)
- [实施计划](../../IMPLEMENTATION_PLAN.md)
- [impact-analysis README](../../../pkg/impact_analysis/README.md)
- [pkg/pipeline 接入文档](../../../pkg/pipeline/INTEGRATION.md)
