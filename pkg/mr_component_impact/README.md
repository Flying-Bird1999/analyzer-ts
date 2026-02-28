# MR 组件影响分析

本包提供 Merge Request 场景下的组件影响分析功能。通过分析 git diff 变更，识别直接变更的组件和函数，以及间接受影响的组件。

## 核心功能

### 1. 文件分类
- 自动识别变更文件属于组件、函数还是其他类型
- 支持通过 manifest 配置定义组件和函数路径

### 2. 组件影响分析
- 基于 component_deps 的结果，直接查询组件依赖关系
- 无需复杂的传播算法，简单高效

### 3. 函数影响分析
- 基于 export_call 的结果，直接获取组件级引用信息
- export_call 已原生支持 RefComponents 字段

### 4. 结果输出
- 提供 JSON 和控制台两种输出格式

## 使用方式

### 代码调用

```go
import (
    mrcomponentimpact "github.com/Flying-Bird1999/analyzer-ts/pkg/mr_component_impact"
    "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/component_deps"
    "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/export_call"
)

// 创建分析器
analyzer := mrcomponentimpact.NewAnalyzer(&mrcomponentimpact.AnalyzerConfig{
    Manifest:      manifest,
    FunctionPaths: []string{"src/functions", "src/utils"},
    ComponentDeps: componentDepsResult,  // component_deps 的结果
    ExportCall:    exportCallResult,    // export_call 的结果
})

// 执行分析
changedFiles := []string{
    "src/components/Button/Button.tsx",
    "src/functions/utils/date.ts",
}
result := analyzer.Analyze(changedFiles)

// 输出结果
fmt.Println(result.ToConsole())
```

## 输出示例

### JSON 输出

```json
{
  "changedComponents": {
    "Button": {
      "name": "Button",
      "files": ["src/components/Button/Button.tsx"]
    }
  },
  "changedFunctions": {
    "utils": {
      "name": "utils",
      "files": ["src/functions/utils/date.ts"]
    }
  },
  "impactedComponents": {
    "Form": [
      {
        "componentName": "Form",
        "impactReason": "依赖组件 Button",
        "changeType": "component",
        "changeSource": "Button"
      }
    ],
    "Calendar": [
      {
        "componentName": "Calendar",
        "impactReason": "引用函数 utils/formatDate",
        "changeType": "function",
        "changeSource": "utils/date.ts"
      }
    ]
  },
  "otherFiles": []
}
```

### 控制台输出

```
========================================
MR 组件影响分析报告
========================================

📦 变更组件:
  • Button
    - src/components/Button/Button.tsx

🔧 变更函数:
  • utils
    - src/functions/utils/date.ts

⚠️  受影响组件:
  • Form
    - 依赖组件 Button
  • Calendar
    - 引用函数 utils/formatDate

========================================
分析完成: 1 个组件变更, 1 个函数变更, 2 个组件受影响, 0 个其他文件
========================================
```

## 架构设计

```
Changed Files (git diff)
    ↓
┌─────────────────────────────────┐
│  Classifier (文件分类器)          │
│  - 判断文件类型                  │
│  - component / functions / other │
└─────────────────────────────────┘
    ↓
┌──────────────────────┬──────────────────────┐
│  ComponentAnalyzer   │  FunctionAnalyzer    │
│  - component_deps │  - export_call       │
│  - 查询组件依赖      │  - RefComponents     │
└──────────────────────┴──────────────────────┘
    ↓
AnalysisResult
```

## 文件结构

```
pkg/mr_component_impact/
├── types.go              # 核心数据结构
├── result.go             # 结果输出
├── classifier.go         # 文件分类器
├── component_analyzer.go # 组件影响分析
├── function_analyzer.go  # 函数影响分析
├── analyzer.go           # 主分析器
├── README.md             # 说明文档
└── USAGE.md              # 使用指南
```

## 依赖说明

本包依赖以下分析器：

| 分析器 | 用途 | 组件级支持 |
|--------|------|-----------|
| **component_deps** | 组件依赖分析 | ✅ ComponentDeps |
| **export_call** | 函数引用分析 | ✅ RefComponents |

这两个分析器均已原生支持组件级影响分析，无需在上层做文件→组件映射。
