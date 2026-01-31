# 组件分析能力技术方案

> **版本**: v2.0.0 | **状态**: ✅ 核心功能完成 | **更新日期**: 2024-01-31

---

## 📊 执行摘要

为 `analyzer-ts` 实现的组件分析能力，包括：

| 功能 | 状态 | 描述 |
|------|------|------|
| component-deps-v2 | ✅ 完成 | 基于配置文件的组件依赖分析 |
| impact-analysis | ✅ 完成 | 基于 BFS 的代码变更影响评估 |
| 单元测试 | ✅ 全部通过 | 17 + 13 个测试用例 |

---

## 🏗️ 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                        输入层                                │
│  component-manifest.json     changes.json     项目源码       │
│  (组件配置)                   (变更文件)      (TS/JS)        │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                        分析层                                │
│  1. component-deps-v2     →  依赖关系图                      │
│  2. impact-analysis       →  影响分析报告                    │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                        输出层                                │
│  depGraph.json              impact-report.json               │
│  (正反向依赖图)              (影响范围+风险评估)              │
└─────────────────────────────────────────────────────────────┘
```

---

## 📋 配置文件格式

### component-manifest.json

```json
{
  "meta": {
    "version": "1.0.0",
    "libraryName": "@example/ui-components"
  },
  "components": [
    {
      "name": "Button",
      "entry": "src/components/Button/index.tsx"
      // scope 自动推断为: src/components/Button/**
    },
    {
      "name": "Input",
      "entry": "src/components/Input/index.tsx"
    }
  ]
}
```

**字段说明**：
- `meta.version`: 配置协议版本
- `meta.libraryName`: 组件库名称
- `components[].name`: 组件名称（唯一标识）
- `components[].entry`: 组件入口文件路径（相对于项目根目录）

**组件作用域自动推断**：
- `entry` = `src/components/Button/index.tsx`
- `scope` = `src/components/Button/**`（自动推断）

### changes.json

```json
{
  "modifiedFiles": ["src/components/Button/Button.tsx"],
  "addedFiles": [],
  "deletedFiles": []
}
```

---

## 🔧 使用方式

### 1. 组件依赖分析

```bash
./analyzer-ts analyze component-deps-v2 \
  -i /absolute/path/to/project \
  -p component-deps-v2.manifest=.analyzer/component-manifest.json
```

**输出示例**：
```json
{
  "component-deps-v2": {
    "meta": {
      "version": "1.0.0",
      "libraryName": "@test/ui-components",
      "componentCount": 3
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

### 2. 影响范围分析

```bash
./analyzer-ts analyze impact-analysis \
  -i /absolute/path/to/project \
  -p impact-analysis.changeFile=/tmp/changes.json \
  -p impact-analysis.depsFile=/tmp/project_data.json
```

**输出示例**：
```json
{
  "impact-analysis": {
    "meta": {
      "analyzedAt": "2024-01-31T22:06:35+08:00",
      "componentCount": 3,
      "changedFileCount": 1
    },
    "changes": [
      {
        "name": "Button",
        "action": "modified",
        "changedFiles": ["src/components/Button/Button.tsx"]
      }
    ],
    "impact": [
      {
        "name": "Button",
        "impactLevel": 0,
        "riskLevel": "low",
        "changePaths": ["Button"]
      },
      {
        "name": "Input",
        "impactLevel": 1,
        "riskLevel": "low",
        "changePaths": ["Button → Input"]
      },
      {
        "name": "Select",
        "impactLevel": 1,
        "riskLevel": "low",
        "changePaths": ["Button → Select"]
      }
    ]
  }
}
```

---

## 🛠️ 核心实现

### 组件依赖分析 (component-deps-v2)

**核心流程**：
```
1. 加载 manifest.json
2. 为每个组件创建 glob 模式（基于 entry 自动推断 scope）
3. 遍历组件文件，提取 import 声明
4. 解析相对路径 → 匹配组件作用域 → 记录依赖
5. 构建正反向依赖图，检测循环依赖
```

**关键实现**：
- **相对路径解析**: `../Input/Input` → `src/components/Input/Input`
- **作用域匹配**: 使用 glob 模式 `src/components/Button/**`
- **循环检测**: DFS + 递归栈

### 影响分析 (impact-analysis)

**BFS 传播算法**：
```
Level 0: [Button]           ← 变更组件
         ↓
Level 1: [Input, Select]    ← 依赖 Button 的组件
         ↓
Level 2: [...]              ← 继续传播...
```

**风险评估模型**：
```
Level 0 (直接变更)  → low
Level 1 (一级间接)  → low
Level 2 (二级间接)  → medium
Level 3 (三级间接)  → high
Level 4+ (四级+)    → critical
```

---

## 📁 文件结构

```
analyzer_plugin/project_analyzer/
├── component_deps_v2/           ✅ 组件依赖分析插件
│   ├── analyzer.go              # 主分析器
│   ├── manifest.go              # 配置解析
│   ├── scope.go                 # 作用域管理
│   ├── dependency.go            # 依赖分析
│   ├── graph.go                 # 依赖图构建
│   ├── result.go                # 结果定义
│   ├── analyzer_test.go         # 单元测试
│   └── README.md                # 插件文档
│
└── impact_analysis/             ✅ 影响分析插件
    ├── analyzer.go              # 主分析器
    ├── types.go                 # 输入类型
    ├── propagation.go           # BFS 传播算法
    ├── chain.go                 # 链路构建
    ├── result.go                # 结果定义
    ├── analyzer_test.go         # 单元测试
    ├── e2e_test.go              # 端到端测试
    └── README.md                # 插件文档

testdata/                        ✅ 测试项目
└── test_project/
    ├── .analyzer/
    │   └── component-manifest.json
    ├── package.json             # 新增：必需
    ├── tsconfig.json            # 新增：必需
    └── src/components/
        ├── Button/
        ├── Input/
        └── Select/
```

---

## 🧪 验证方式

### 单元测试
```bash
go test ./analyzer_plugin/project_analyzer/... -v
```

### 完整端到端验证
```bash
# 1. 依赖分析
./analyzer-ts analyze component-deps-v2 \
  -i /Users/bird/Desktop/alalyzer/analyzer-ts/testdata/test_project \
  -p component-deps-v2.manifest=.analyzer/component-manifest.json \
  -o /tmp

# 2. 影响分析
./analyzer-ts analyze impact-analysis \
  -i /Users/bird/Desktop/alalyzer/analyzer-ts/testdata/test_project \
  -p impact-analysis.changeFile=/tmp/changes.json \
  -p impact-analysis.depsFile=/tmp/test_project_analyzer_data.json
```

---

## ❓ 常见问题

**Q: 为什么依赖图是空的？**
A: 确保项目根目录包含 `tsconfig.json`，这样解析器才能正确处理 `.tsx` 文件。

**Q: 支持哪些导入路径格式？**
A: 支持相对路径（如 `../Button/Button`）和绝对路径。npm 包会被自动识别为外部依赖。

**Q: 如何处理循环依赖？**
A: 使用 DFS + 递归栈检测，会在结果中标记有循环依赖的组件。

---

## 🔗 相关文档

- [component_deps_v2 README](../../analyzer_plugin/project_analyzer/component_deps_v2/README.md)
- [impact_analysis README](../../analyzer_plugin/project_analyzer/impact_analysis/README.md)
