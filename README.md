# analyzer-ts

<div align="center">

**一个高性能、可扩展的 TypeScript/JavaScript 项目分析工具**

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/tests-2660+-green.svg)](./analyzer/parser/test/)

[功能特性](#功能特性) • [快速开始](#快速开始) • [核心能力](#核心能力) • [架构设计](#架构设计) • [扩展开发](#扩展开发)

</div>

---

## 📖 简介

`analyzer-ts` 是一个使用 Go 语言编写的高性能 TypeScript/JavaScript 项目分析命令行工具。它采用创新的**插件式架构**，将项目解析与代码分析完全分离，实现了"**一次解析，多次分析**"的高效模式。

### 🎯 核心价值

- **🚀 高性能**: 基于 Go 语言的性能优势，处理大型 TypeScript 项目效率极高
- **🔌 可扩展**: 插件式架构，新增分析器无需修改核心代码
- **✅ 完整性**: 基于 TypeScript 官方解析器的 Go 绑定，保证解析准确性
- **🛠️ 实用性**: 提供多种开箱即用的分析器，解决实际开发痛点
- **🏗️ Monorepo 友好**: 原生支持 Monorepo 项目结构
- **📦 类型打包**: 独特的 TypeScript 类型声明打包工具

### 📊 技术指标

- **主项目代码量**: 约 25,000+ 行 Go 代码
- **测试文件数**: 2,660+ 个测试用例
- **核心 Go 文件**: 70+ 个源文件
- **analyzer 模块**: 约 6,830 行代码
- **Go 版本要求**: 1.25+

---

## ✨ 功能特性

### 🔍 代码质量分析

- **[count-any](#count-any---统计-any-类型)**: 统计项目中所有 `any` 类型的使用情况，评估类型安全性
- **[count-as](#count-as---统计-as-断言)**: 统计所有 `as` 类型断言的使用，识别潜在的类型转换问题
- **[unconsumed](#unconsumed---查找未使用的导出)**: 查找已导出但从未被导入的符号，清理死代码
- **[find-unreferenced-files](#find-unreferenced-files---查找未引用的文件)**: 查找从未被引用的"孤岛"文件

### 📦 依赖管理

- **[npm-check](#npm-check---npm-依赖检查)**: 检查隐式依赖、未使用依赖和过期依赖
- **[trace](#trace---npm-包使用追踪)**: 追踪特定 NPM 包在项目中的使用情况
- **[find-callers](#find-callers---查找调用者)**: 查找指定文件的所有上游调用方

### 🏗️ 架构分析

- **[component-deps-v2](#component-deps-v2---组件依赖分析-v2)**: 基于配置文件的组件依赖关系分析
- **[component-deps](#component-deps---组件依赖分析)**: 分析组件之间的依赖关系
- **[api-tracer](#api-tracer---api-调用链追踪)**: 追踪 API 的完整调用链路

### 🔥 代码影响分析 (Pipeline)

- **[impact](#impact---代码变更影响分析)**: 完整的代码变更影响分析管道，支持多种输入源

### 🛠️ 开发工具

- **[bundle](#bundle---类型声明打包)**: 递归收集类型及其依赖，生成独立的 `.d.ts` 文件
- **[batch-bundle](#batch-bundle---批量类型打包)**: 批量打包多个类型，自动解决命名冲突
- **[query](#query---jmespath-查询)**: 使用 JMESPath 查询语法灵活查询项目数据
- **[scan](#scan---项目文件扫描)**: 扫描项目文件，生成文件列表

---

## 🚀 快速开始

### 安装

我们提供两种安装方式：`go install`（推荐）或从源码构建。

#### 方式一：全局安装 (推荐)

```bash
# 确保已安装 Go 1.25 或更高版本
go install github.com/Flying-Bird1999/analyzer-ts@latest
```

#### 方式二：从源码构建

```bash
# 克隆仓库
git clone https://github.com/Flying-Bird1999/analyzer-ts.git
cd analyzer-ts

# 构建项目
go build -o analyzer-ts
```

### 第一个分析

```bash
# 分析项目中的 any 类型使用
analyzer-ts analyze count-any -i /path/to/your/project

# 检查 NPM 依赖健康
analyzer-ts analyze npm-check -i /path/to/your/project

# 组合多个分析器
analyzer-ts analyze count-any unconsumed npm-check \
  -i /path/to/your/project \
  -o ./output
```

---

## 🎯 核心能力

### count-any - 统计 any 类型

统计项目中所有 `any` 类型的使用情况，帮助评估项目的类型安全性。

**功能特性**:
- 精确统计每个文件的 `any` 使用次数
- 提供详细的位置信息（行号、列号）
- 显示原始代码片段
- 生成总体统计报告

**使用示例**:

```bash
analyzer-ts analyze count-any -i /path/to/project -o ./output
```

**输出示例**:

```json
{
  "totalAnyCount": 42,
  "filesParsed": 150,
  "fileCounts": [
    {
      "filePath": "/src/utils/helper.ts",
      "anyCount": 5,
      "details": [
        {
          "sourceLocation": {"start": {"line": 10, "column": 5}},
          "raw": "const data: any = response;"
        }
      ]
    }
  ]
}
```

**使用场景**:
- 评估项目类型安全性
- 追踪类型改进进展
- 识别需要重构的代码区域

---

### count-as - 统计 as 断言

统计项目中所有 `as` 类型断言的使用情况。

**使用示例**:

```bash
analyzer-ts analyze count-as -i /path/to/project
```

---

### unconsumed - 查找未使用的导出

识别已导出但从未被导入的符号，帮助清理死代码。

**支持导出类型**:
- 函数声明 (`export function foo() {}`)
- 变量声明 (`export const bar = 1`)
- 接口声明 (`export interface Baz {}`)
- 类型声明 (`export type Qux = {}`)
- 枚举声明 (`export enum Quux {}`)
- 默认导出 (`export default ...`)
- 重导出 (`export { X } from './module'`)

**智能过滤**:
- 自动忽略测试文件 (`.test.ts`, `.spec.ts`)
- 忽略类型声明文件 (`.d.ts`)
- 忽略测试目录 (`__tests__`, `__mocks__`)

**使用示例**:

```bash
analyzer-ts analyze unconsumed -i /path/to/project -o ./output
```

**输出示例**:

```json
{
  "unconsumedExports": [
    {
      "filePath": "/src/utils/helper.ts",
      "symbolName": "unusedFunction",
      "symbolType": "function",
      "exportedAt": {"line": 15, "column": 1}
    }
  ]
}
```

---

### find-unreferenced-files - 查找未引用的文件

在项目中查找所有从未被任何其他文件导入或引用的"孤岛"文件。

**使用示例**:

```bash
analyzer-ts analyze find-unreferenced-files -i /path/to/project
```

**使用场景**:
- 清理冗余文件
- 减少项目维护成本
- 优化构建时间

---

### npm-check - NPM 依赖检查

检查隐式依赖、未使用依赖和过期依赖。

**检查项**:

1. **隐式依赖检测**: 识别在代码中使用但未在 `package.json` 中声明的依赖
2. **未使用依赖检测**: 识别在 `package.json` 中声明但从未在代码中使用的依赖
3. **过期依赖检测**: 检查依赖是否有新版本可用

**使用示例**:

```bash
analyzer-ts analyze npm-check -i /path/to/project
```

**输出示例**:

```json
{
  "implicitDependencies": [
    {
      "name": "lodash",
      "filePath": "/src/utils.ts",
      "raw": "import { debounce } from 'lodash';"
    }
  ],
  "unusedDependencies": [
    {
      "name": "moment",
      "version": "^1.2.3",
      "packageJsonPath": "/package.json"
    }
  ],
  "outdatedDependencies": [
    {
      "name": "react",
      "currentVersion": "^17.0.0",
      "latestVersion": "18.2.0"
    }
  ]
}
```

**使用场景**:
- 保持 `package.json` 的准确性
- 减少不必要的依赖
- 及时更新过期依赖

---

### trace - NPM 包使用追踪

追踪特定 NPM 包在项目中的使用情况。

**使用示例**:

```bash
# 追踪单个包
analyzer-ts analyze trace \
  -i /path/to/project \
  -p "trace.targetPkgs=lodash"

# 追踪多个包
analyzer-ts analyze trace \
  -i /path/to/project \
  -p "trace.targetPkgs=antd" \
  -p "trace.targetPkgs=@yy/sl-admin-components"
```

**使用场景**:
- 评估替换某个包的影响
- 了解第三方包的使用分布
- 优化依赖结构

---

### find-callers - 查找调用者

查找一个或多个指定文件的所有上游调用方。

**使用示例**:

```bash
# 查找单个文件的调用者
analyzer-ts analyze find-callers \
  -i /path/to/project \
  -p "find-callers.targetFiles=/path/to/file1.ts"

# 查找多个文件的调用者
analyzer-ts analyze find-callers \
  -i /path/to/project \
  -p "find-callers.targetFiles=/path/to/file1.ts" \
  -p "find-callers.targetFiles=/path/to/file2.ts"
```

**使用场景**:
- 重构前了解影响范围
- 分析代码调用链
- 文档化 API 使用情况

---

### component-deps - 组件依赖分析

分析组件之间的依赖关系。

**使用示例**:

```bash
analyzer-ts analyze component-deps \
  -i /path/to/project \
  -p "component-deps.entryPoint=./src/index.tsx"
```

**使用场景**:
- 优化组件结构
- 减少循环依赖
- 可视化组件依赖图

---

### api-tracer - API 调用链追踪

追踪 API 的完整调用链路。

**使用示例**:

```bash
analyzer-ts analyze api-tracer \
  -i /path/to/project \
  -p "api-tracer.apiPaths=/api/users" \
  -p "api-tracer.apiPaths=/api/orders"
```

**使用场景**:
- 文档化 API 使用情况
- 分析 API 调用链
- 优化 API 设计

---

### bundle - 类型声明打包

递归收集类型及其所有依赖，生成独立的 `.d.ts` 文件。

**核心特性**:
- 递归分析类型依赖
- 处理循环依赖
- 自动解决命名冲突
- 生成完整的类型声明

**使用示例**:

```bash
# 单类型打包
analyzer-ts bundle \
  -i ./src/api/user.ts \
  -t UserProfile \
  -o ./dist/types/user.d.ts
```

**使用场景**:
- 微服务架构中的类型共享
- 生成 SDK 类型定义
- 提取 API 类型文档

---

### batch-bundle - 批量类型打包

批量打包多个类型，自动解决命名冲突。

**使用示例**:

```bash
# 批量打包多个类型，使用别名避免命名冲突
analyzer-ts batch-bundle \
  -e "./src/user.ts:User:UserDTO" \
  -e "./src/admin.ts:User:AdminDTO" \
  -e "./src/product.ts:Product:ProductDTO" \
  --output-dir ./dist/types/
```

**特性**:
- 文件级缓存优化
- 支持类型别名
- 独立文件输出
- 自动命名冲突解决

---

### query - JMESPath 查询

使用 JMESPath 查询语法灵活查询项目数据。

**使用示例**:

```bash
# 查找所有包含 'any' 类型的文件
analyzer-ts query \
  -i /path/to/project \
  -j "js_data.*[?contains(@.extractedNodes.anyDeclarations, `true`)]"

# 统计每个文件的导入数量
analyzer-ts query \
  -i /path/to/project \
  -j "js_data.{filePath: keys(@), importCount: @.*.importDeclarations | length(@)}"

# 提取所有导出的接口
analyzer-ts query \
  -i /path/to/project \
  -j "js_data.*.interfaceDeclarations.*"
```

**使用场景**:
- 灵活的数据提取
- 自定义分析脚本
- 生成定制化报告

---

### scan - 项目文件扫描

扫描项目文件，生成文件列表。

**使用示例**:

```bash
analyzer-ts scan \
  -i /path/to/project \
  -o ./output \
  -x "node_modules/**" \
  -x "**/*.test.ts"
```

---

### impact - 代码变更影响分析

完整的代码变更影响分析管道，支持多种输入源和输出格式。

**功能特性**:
- 支持多种 diff 输入源（文件、字符串、git diff、GitLab API）
- 自动解析项目 AST 并分析符号级变更
- 计算文件级和组件级影响范围
- 支持 Monorepo 项目（显式指定 git-root）
- 支持组件库项目（通过 component-manifest.json）

**使用示例**:

```bash
# 使用 diff 文件
analyzer-ts impact \
  --project-root /path/to/project \
  --diff-file ./changes.patch \
  --output impact-result.json

# 使用 git diff
analyzer-ts impact \
  --project-root /path/to/project \
  --git-diff "HEAD~1 HEAD"

# 使用 diff 字符串（适合 CI/CD）
analyzer-ts impact \
  --project-root /path/to/project \
  --diff-string "$(git diff HEAD~1 HEAD)" \
  --format summary

# Monorepo 项目
analyzer-ts impact \
  --project-root /path/to/project \
  --git-root /path/to/git/root \
  --manifest .analyzer/component-manifest.json \
  --diff-file ./changes.patch
```

**输出示例**:

```json
{
  "meta": {
    "projectRoot": "/path/to/project",
    "analyzedAt": "2024-01-01T00:00:00Z"
  },
  "fileAnalysis": {
    "meta": {
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
        "impactType": "internal"
      }
    ]
  },
  "componentAnalysis": {
    "meta": {
      "changedComponentCount": 1,
      "impactComponentCount": 8
    },
    "changes": [{"name": "Button"}],
    "impact": [{"name": "Form", "impactLevel": 2}]
  }
}
```

**使用场景**:
- Code Review 前了解变更影响范围
- CI/CD 质量门禁（影响范围过大则阻止合并）
- 重构前风险评估
- 发布前回归测试范围评估

---

### component-deps-v2 - 组件依赖分析 (v2)

基于 `component-manifest.json` 配置文件的组件依赖关系分析。

**功能特性**:
- 配置驱动：通过 manifest.json 显式声明组件
- 作用域自动推断：基于 entry 文件自动推断组件作用域
- 相对路径解析：正确处理跨组件的相对路径导入
- 循环依赖检测：自动检测并报告循环依赖
- 双向依赖图：生成正向和反向依赖关系图

**配置文件**:

```json
// .analyzer/component-manifest.json
{
  "meta": {
    "version": "1.0.0",
    "libraryName": "@your-org/ui-components"
  },
  "components": [
    {
      "name": "Button",
      "entry": "src/components/Button/index.tsx"
    },
    {
      "name": "Input",
      "entry": "src/components/Input/index.tsx"
    }
  ]
}
```

**使用示例**:

```bash
analyzer-ts analyze component-deps-v2 \
  -i /path/to/project \
  -p component-deps-v2.manifest=.analyzer/component-manifest.json \
  -o ./output
```

**输出示例**:

```json
{
  "component-deps-v2": {
    "meta": {
      "libraryName": "@your-org/ui-components",
      "componentCount": 2
    },
    "components": {
      "Button": {
        "entry": "src/components/Button/index.tsx",
        "dependencies": []
      },
      "Input": {
        "entry": "src/components/Input/index.tsx",
        "dependencies": ["Button"]
      }
    },
    "depGraph": {
      "Button": [],
      "Input": ["Button"]
    },
    "revDepGraph": {
      "Button": ["Input"],
      "Input": []
    }
  }
}
```

**使用场景**:
- 组件库架构优化
- 循环依赖检测和解决
- 组件拆分/合并前的依赖分析
- 生成组件依赖可视化

---

## 🏗️ 系统架构与核心能力

### 架构分层设计

`analyzer-ts 采用五层架构设计，从底层到上层逐层构建 TypeScript 项目分析能力：

```
┌─────────────────────────────────────────────────────┐
│    第5层: 高级应用层 (tsmorphgo API)                 │
│  - ts-morph 风格的 Go API                             │
│  - 类型安全的节点操作                                 │
│  - 符号分析与引用查找                                 │
├─────────────────────────────────────────────────────┤
│    第4层: 语言服务层 (LSP Integration)                │
│  - TypeScript 语言服务协议                            │
│  - 类型提示与定义跳转                                 │
│  - 跨文件引用查找                                     │
├─────────────────────────────────────────────────────┤
│    第3层: 项目分析层 (ProjectParser)                  │
│  - 路径别名解析 (@/components → src/components)      │
│  - Monorepo 多包支持                                  │
│  - 依赖关系图构建                                     │
│  - package.json 解析                                  │
├─────────────────────────────────────────────────────┤
│    第2层: 文件解析层 (Parser)                         │
│  - AST 遍历与节点提取                                 │
│  - 19 种语法节点支持                                  │
│  - 导入/导出/函数/类型声明解析                        │
├─────────────────────────────────────────────────────┤
│    第1层: 文件扫描层 (ScanProject)                    │
│  - 文件系统遍历                                       │
│  - Glob 模式过滤                                      │
│  - 文件元数据收集                                     │
└─────────────────────────────────────────────────────┘
           ↓ (依赖)
┌─────────────────────────────────────────────────────┐
│  底层: typescript-go (TypeScript 官方解析器 Go 绑定)  │
└─────────────────────────────────────────────────────┘
```

---

### 核心能力详解

#### 📁 第1层: 文件扫描能力 (scanProject)

**能力描述**: 高效的项目文件发现与过滤

```go
// 输入: 项目路径 + 忽略规则
scanner := scanProject.NewProjectResult(rootPath, ignorePatterns, isMonorepo)
scanner.ScanProject()

// 输出: 完整的文件清单
type ProjectResult struct {
    Root       string
    FileList   map[string]FileItem  // 文件名、大小、扩展名
}
```

**核心特性**:
- ✅ **智能过滤**: 支持 glob 模式 (`node_modules/**`, `**/*.test.ts`)
- ✅ **元数据提取**: 自动收集文件大小、扩展名
- ✅ **Monorepo 优化**: 针对多包项目的扫描策略
- ✅ **性能优化**: 使用 `filepath.SkipDir` 提前跳过忽略目录

---

#### 🔍 第2层: AST 解析能力 (parser)

**能力描述**: 单文件的深度语法分析与节点提取

**支持的节点类型** (19 种):

| 节点类型 | 解析能力 | 应用场景 |
|---------|---------|---------|
| **ImportDeclaration** | 默认导入、命名导入、命名空间导入、副作用导入 | 依赖分析 |
| **ExportDeclaration** | 导出语句、重导出 | API 文档生成 |
| **ExportAssignment** | 默认导出 | 模块分析 |
| **FunctionDeclaration** | 函数名、参数、返回值、泛型、async/generator | API 文档、代码质量 |
| **VariableDeclaration** | const/let/var、解构、类型注解 | 代码分析 |
| **InterfaceDeclaration** | 属性、方法、继承 | 类型系统分析 |
| **TypeAliasDeclaration** | 类型别名、泛型 | 类型提取 |
| **EnumDeclaration** | 枚举成员 | 代码分析 |
| **CallExpression** | 调用者、参数、动态导入 | 调用链分析 |
| **JsxElement** | 组件路径、属性 | React 组件分析 |
| **ReturnStatement** | 返回值表达式 | 控制流分析 |
| **AnyKeyword** | `any` 类型位置定位 | 类型安全检查 |
| **AsExpression** | 类型断言 | 类型质量分析 |

**解析结果示例**:

```go
type ParserResult struct {
    FilePath              string
    ImportDeclarations    []ImportDeclarationResult
    ExportDeclarations    []ExportDeclarationResult
    FunctionDeclarations  []FunctionDeclarationResult
    InterfaceDeclarations map[string]InterfaceDeclarationResult
    VariableDeclarations  []VariableDeclaration
    CallExpressions       []CallExpression
    JsxElements           []JSXElement
    ExtractedNodes        ExtractedNodes  // any、as 等特殊节点
}
```

**核心优势**:
- 🎯 **访问者模式**: 解耦遍历逻辑与节点处理
- 🎯 **精确位置**: 行号、列号、偏移量级别定位
- 🎯 **完整类型**: 提取所有类型注解信息
- 🎯 **容错机制**: panic 恢复与错误收集

---

#### 🌐 第3层: 项目分析能力 (projectParser)

**能力描述**: 项目级依赖分析与路径别名解析

**核心能力**:

##### A. 路径别名解析

```typescript
// tsconfig.json
{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"],
      "@components/*": ["src/components/*"]
    }
  }
}

// 自动解析
import { Button } from '@/components/Button'
// ↓ 解析结果
// { FilePath: "/project/src/components/Button", Type: "file" }
```

##### B. Monorepo 支持

```go
// 自动查找所有子包的 tsconfig
PackageTsConfigMaps: map[string]TsConfig{
    "/packages/admin":  {Alias: {"@admin/*": "src/*"}},
    "/packages/app":    {Alias: {"@app/*": "src/*"}},
}

// 为每个文件选择最相关的 tsconfig
func (ppr *ProjectParserResult) getTsConfigForFile(targetPath string)
```

##### C. 依赖关系构建

```go
// 完整的导入信息
type ImportDeclarationResult struct {
    ImportModules []ImportModule
    Source        SourceData  // 解析后的绝对路径
    // Source.Type: "file" | "npm" | "unknown"
    // Source.FilePath: /absolute/path/to/file
    // Source.NpmPkg: package-name
}
```

##### D. package.json 解析

```go
type PackageJsonFileParserResult struct {
    Workspace string              // "root" 或子包名
    NpmList   map[string]NpmItem  // 依赖详情
}

type NpmItem struct {
    Name              string  // "react"
    Type              string  // "dependencies" | "devDependencies"
    Version           string  // 声明版本 "^18.0.0"
    NodeModuleVersion string  // 实际安装版本 "18.2.0"
}
```

---

#### 🔎 第4层: 语言服务能力 (lsp)

**能力描述**: TypeScript 官方语言服务协议集成

**核心 API**:

```go
// 符号分析
symbol := lsp.GetSymbolAt(filePath, line, column)
// 返回: 符号名称、类型、声明位置、作用域

// 引用查找
refs := lsp.FindReferences(filePath, line, column)
// 返回: 所有引用位置 (跨文件)

// 定义跳转
def := lsp.GotoDefinition(filePath, line, column)
// 返回: 定义位置 (文件、行、列)

// 类型提示
info := lsp.GetQuickInfoAtPosition(filePath, line, column)
// 返回: TypeText (完整类型)、Documentation (JSDoc)
```

**应用场景**:
- 🔍 **精确引用查找**: 跨文件查找符号的所有引用
- 🏷️ **类型推断**: 获取任意位置的类型信息
- 📖 **文档生成**: 自动提取 JSDoc 注释
- 🔀 **重构支持**: 基于引用的重命名、移动

---

#### 🎨 第5层: 高级 API 能力 (tsmorphgo)

**能力描述**: ts-morph 风格的类型安全 Go API

##### A. Project API - 项目级操作

```go
project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
    RootPath: "/path/to/project",
    UseTsConfig: true,
    IsMonorepo: false,
})

// 文件管理
file := project.GetSourceFile("/path/to/file.ts")
files := project.GetSourceFiles()

// 动态文件操作
project.CreateSourceFile("/new/file.ts", sourceCode)
project.UpdateSourceFile("/existing/file.ts", newSourceCode)
project.RemoveSourceFile("/old/file.ts")

// 节点查找
node := project.FindNodeAt(filePath, line, column)
```

##### B. SourceFile API - 文件级操作

```go
// 获取解析结果
result := sourceFile.GetFileResult()
fmt.Println(result.ImportDeclarations)
fmt.Println(result.FunctionDeclarations)

// 节点遍历
sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
    // 处理每个节点
})
```

##### C. Node API - 节点操作 (核心能力)

**基础信息 API**:
```go
text := node.GetText()              // 源码文本
startLine := node.GetStartLineNumber()  // 1-based
kind := node.GetKind()               // SyntaxKind 枚举
```

**导航 API**:
```go
parent := node.GetParent()
ancestors := node.GetAncestors()
children := node.GetChildren()
```

**类型检查 API**:
```go
isIdentifier := node.IsIdentifier()
isFunction := node.IsFunctionDeclaration()
isCall := node.IsCallExpression()
```

**透传 API** (核心创新):
```go
// 类型安全的获取底层解析数据
if importDecl, ok := node.AsImportDeclaration(); ok {
    fmt.Println(importDecl.Source)  // 导入源
    for _, module := range importDecl.ImportModules {
        fmt.Println(module.Identifier)
    }
}

if funcDecl, ok := node.AsFunctionDeclaration(); ok {
    fmt.Println(funcDecl.Parameters)  // 参数列表
    fmt.Println(funcDecl.ReturnType)  // 返回类型
}
```

##### D. References API - 引用查找

```go
// 查找所有引用
refs, err := node.FindReferences()

// 带缓存的查找 (性能优化)
refs, fromCache, err := node.FindReferencesWithCache()

// 统计引用数量
count, err := node.CountReferences()
```

**缓存机制**:
```go
type ReferenceCache struct {
    maxSize         int              // 最大缓存条目
    ttl             time.Duration    // 生存时间
    cleanupInterval time.Duration    // 清理间隔
}
```

---

### 能力组合示例

#### 示例 1: 依赖分析工具

```go
// 组合: scanProject → parser → projectParser
func analyzeProjectDependencies(rootPath string) {
    // 1. 扫描文件
    scanner := scanProject.NewProjectResult(rootPath, ignore, false)
    scanner.ScanProject()

    // 2. 解析导入
    for path := range scanner.FileList {
        p := parser.NewParser(path)
        p.Traverse()
        // 获取: p.Result.ImportDeclarations
    }

    // 3. 解析路径别名
    source := projectParser.MatchImportSource(
        path,
        importSource,
        tsconfig.Alias,
        baseUrl,
    )
    // source.Type: "file" | "npm"
}
```

#### 示例 2: 类型安全检查工具

```go
// 组合: parser → tsmorphgo
func checkTypeSafety(project *tsmorphgo.Project) {
    files := project.GetSourceFiles()

    for _, file := range files {
        result := file.GetFileResult()

        // 分析 any 类型使用
        for _, anyInfo := range result.ExtractedNodes.AnyDeclarations {
            fmt.Printf("any found at %v\n", anyInfo.SourceLocation)
        }

        // 分析 as 断言
        for _, asExpr := range result.ExtractedNodes.AsExpressions {
            fmt.Printf("as assertion: %s\n", asExpr.Raw)
        }
    }
}
```

#### 示例 3: 重命名重构工具

```go
// 组合: tsmorphgo → lsp → tsmorphgo
func renameSymbol(project *tsmorphgo.Project, filePath string, line, col int, newName string) {
    // 1. 找到目标节点
    node := project.FindNodeAt(filePath, line, col)

    // 2. 查找所有引用
    refs, err := node.FindReferences()

    // 3. 应用重命名
    for _, ref := range refs {
        file := ref.GetSourceFile()
        newSource := rewriteSymbol(file, ref, newName)
        project.UpdateSourceFile(file.GetFilePath(), newSource)
    }
}
```

---

### 🚀 基于核心能力的扩展方向

#### 1. 代码质量工具

- **类型安全分析**: 统计 `any` 类型、`as` 断言使用
- **死代码检测**: 查找未使用的导出和文件
- **复杂度分析**: 基于函数声明的圈复杂度计算
- **代码重复检测**: 基于 AST 的相似代码查找

#### 2. 文档生成工具

- **API 文档**: 提取函数、接口、类型的 JSDoc
- **依赖图可视化**: 生成模块依赖关系图
- **架构文档**: 分析项目的层次结构和模块划分
- **接口契约**: 从 TypeScript 类型生成 API 规范

#### 3. 重构工具

- **符号重命名**: 基于 LSP 的跨文件重命名
- **模块移动**: 自动更新导入路径
- **内联函数/提取函数**: 基于 AST 的代码重构
- **类型推断**: 自动添加类型注解

#### 4. 测试工具

- **测试覆盖率**: 分析哪些导出没有测试
- **Mock 生成**: 基于接口自动生成 Mock 对象
- **测试用例生成**: 基于函数签名生成测试模板
- **快照测试**: 生成组件的输出快照

#### 5. 架构分析工具

- **循环依赖检测**: 检测模块间的循环引用
- **调用链分析**: 追踪函数的完整调用链路
- **层次分析**: 识别项目的分层架构
- **耦合度分析**: 计算模块间的耦合度

#### 6. 性能分析工具

- **热点函数分析**: 统计函数调用频率
- **Bundle 分析**: 分析打包体积和优化建议
- **懒加载分析**: 识别可以懒加载的模块
- **依赖优化**: 找出可以优化的依赖关系

#### 7. AI 辅助编程

- **代码补全**: 基于类型系统的智能补全
- **代码搜索**: 语义级别的代码搜索 (不是文本搜索)
- **代码理解**: 自动解释代码的功能
- **重构建议**: 基于最佳实践的重构建议

---

### 核心技术优势

| 维度 | 优势 | 说明 |
|------|------|------|
| **性能** | Go 语言 + 一次解析多次使用 | 处理大型项目速度快，内存占用低 |
| **准确性** | TypeScript 官方解析器 | 100% 兼容 TypeScript 语法 |
| **完整性** | 19 种节点类型 + LSP | 覆盖所有语法元素 + 完整类型信息 |
| **可扩展性** | 五层架构 + 透传 API | 可在任意层级扩展功能 |
| **易用性** | ts-morph 风格 API | 熟悉的接口设计，学习成本低 |
| **类型安全** | 完整的 Go 类型系统 | 编译时类型检查，减少运行时错误 |

---

### 性能优化策略

1. **解析缓存**: 项目解析结果可被多个分析器共享
2. **引用查找缓存**: LRU 缓存优化引用查找性能
3. **并发处理**: 支持并发执行多个文件分析
4. **惰性加载**: LSP 服务按需初始化
5. **智能路径匹配**: Monorepo 中最优 tsconfig 选择

---

### 核心理念: 解析一次，分析多次

**传统做法**:
```go
// 每个分析器都需要重新解析项目
analyzer1.ParseProject()  // 解析成本高
analyzer1.Analyze()

analyzer2.ParseProject()  // 重复解析！
analyzer2.Analyze()
```

**analyzer-ts 做法**:
```go
// 只解析一次
parsingResult := ParseProject()  // 只执行一次

// 多个分析器共享结果
analyzer1.Analyze(parsingResult)  // 零成本
analyzer2.Analyze(parsingResult)  // 零成本
analyzer3.Analyze(parsingResult)  // 零成本
```

**性能提升**:
- 解析阶段: O(n) 文件数量
- 分析阶段: O(1) 相对于解析成本
- **10 个分析器 ≈ 1.1x 解析时间 (而非 10x)**

---

## 🎨 插件系统 (可选使用)

**注意**: 插件系统是基于核心解析能力的高级应用层。如果您只需要使用底层 API 进行自定义开发，可以跳过此章节。

所有分析器实现统一的接口：

```go
// Analyzer 接口
type Analyzer interface {
    Name() string                                     // 分析器唯一标识
    Configure(params map[string]string) error         // 配置分析器
    Analyze(ctx *ProjectContext) (Result, error)      // 执行分析
}

// Result 接口
type Result interface {
    Name() string                  // 结果名称
    Summary() string               // 人类可读摘要
    ToJSON(indent bool) ([]byte, error)  // JSON 序列化
    ToConsole() string             // 控制台格式化输出
}

// ProjectContext - 分析器共享上下文
type ProjectContext struct {
    ProjectRoot   string
    Exclude       []string
    IsMonorepo    bool
    ParsingResult *ProjectParserResult  // 共享的解析结果
}
```

**开发新分析器只需 3 步**:

1. 实现 `Analyzer` 和 `Result` 接口
2. 注册到命令行
3. 添加到分析器注册表

### 技术栈

| 依赖包 | 版本 | 用途 |
|--------|------|------|
| `github.com/Zzzen/typescript-go` | v0.0.2 | TypeScript 官方解析器 Go 绑定 |
| `github.com/spf13/cobra` | v1.9.1 | 命令行接口框架 |
| `github.com/samber/lo` | v1.50.0 | Go 高效函数式编程库 |
| `github.com/gobwas/glob` | v0.2.3 | Glob 模式匹配 |
| `github.com/jmespath/go-jmespath` | v0.4.0 | JMESPath 查询语言 |

### 核心技术亮点

1. **TypeScript 官方解析器**: 基于 `github.com/Zzzen/typescript-go`，保证 100% 兼容
2. **智能路径解析**: 自动处理 `tsconfig.json` 的 `paths` 和 `baseUrl`
3. **Monorepo 原生支持**: 自动查找子包的 `tsconfig`
4. **JSX/React 支持**: 自动识别 JSX 组件的隐式导入
5. **并发处理**: 支持并发版本检查等操作
6. **智能缓存**: LRU 缓存优化性能
7. **精确位置信息**: 保留行号、列号、偏移量等详细信息

---

## 🔧 扩展开发

### 开发新分析器

**步骤 1**: 创建分析器目录

```bash
mkdir -p analyzer_plugin/project_analyzer/my_analyzer
```

**步骤 2**: 实现接口

```go
// analyzer_plugin/project_analyzer/my_analyzer/my_analyzer.go
package my_analyzer

import (
    "github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
    projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

type MyAnalyzer struct {
    config map[string]string
}

type MyResult struct {
    // 结果数据
}

func (m *MyAnalyzer) Name() string {
    return "my-analyzer"
}

func (m *MyAnalyzer) Configure(params map[string]string) error {
    m.config = params
    return nil
}

func (m *MyAnalyzer) Analyze(ctx *projectanalyzer.ProjectContext) (projectanalyzer.Result, error) {
    parseResult := ctx.ParsingResult

    // 执行分析逻辑
    // ...

    return &MyResult{
        // 返回结果
    }, nil
}

func (r *MyResult) Name() string {
    return "my-analyzer-result"
}

func (r *MyResult) Summary() string {
    return "分析完成"
}

func (r *MyResult) ToJSON(indent bool) ([]byte, error) {
    // JSON 序列化
}

func (r *MyResult) ToConsole() string {
    // 控制台输出
}
```

**步骤 3**: 注册到命令

```go
// analyzer_plugin/project_analyzer/cmd/analyze.go
var availableAnalyzers = map[string]projectanalyzer.Analyzer{
    "existing-analyzer": &ExistingAnalyzer{},
    "my-analyzer":       &MyAnalyzer{},  // 新增
}
```

### 测试分析器

```go
// analyzer_plugin/project_analyzer/my_analyzer/my_analyzer_test.go
package my_analyzer

import (
    "testing"
    // ...
)

func TestMyAnalyzer(t *testing.T) {
    // 测试逻辑
}
```

详细开发指南请参阅: [分析器架构详解](./analyzer_plugin/project_analyzer/README.md)

---

## 📂 项目结构

```
analyzer-ts/
├── main.go                          # 程序入口
├── go.mod / go.sum                  # Go 依赖管理
├── README.md                        # 项目文档
│
├── cmd/                             # 命令行接口层
│   ├── root.go                      # 根命令定义
│   ├── impact.go                    # impact 子命令（代码影响分析）
│   ├── scan.go                      # scan 子命令
│   └── version.go                   # 版本信息
│
├── pkg/                             # 核心能力包
│   └── pipeline/                    # 代码影响分析管道
│       ├── README.md                # 架构设计文档
│       ├── INTEGRATION.md           # 接入文档
│       ├── pipeline.go              # 管道核心
│       ├── gitlab_pipeline.go       # GitLab MR 管道
│       ├── diff_parser_stage.go     # Diff 解析阶段
│       ├── symbol_analysis_stage.go # 符号分析阶段
│       └── stage.go                 # 阶段接口
│
├── analyzer/                        # 核心解析引擎
│   ├── scanProject/                 # 第1层: 文件扫描
│   ├── parser/                      # 第2层: 单文件解析
│   │   ├── parser.go                # 主解析器
│   │   ├── typeAnalyzer.go          # 类型分析器
│   │   ├── extractedNodes.go        # 提取的节点信息
│   │   └── test/                    # 单元测试
│   ├── projectParser/               # 第3层: 项目级解析
│   │   ├── projectParser.go         # 主项目解析器
│   │   └── utils.go                 # 工具函数
│   └── lsp/                         # LSP 服务集成
│
├── analyzer_plugin/                 # 插件系统
│   ├── project_analyzer/            # 项目分析器插件集
│   │   ├── README.md                # 插件开发指南
│   │   ├── projectanalyzer.go       # 核心接口定义
│   │   ├── cmd/                     # 命令行集成
│   │   │   ├── analyze.go           # analyze 命令
│   │   │   └── query.go             # query 命令
│   │   │
│   │   ├── countAny/                # 统计 any 类型
│   │   ├── countAs/                 # 统计 as 断言
│   │   ├── unconsumed/              # 查找未使用的导出
│   │   ├── unreferenced/            # 查找未引用的文件
│   │   ├── dependency/              # NPM 依赖检查
│   │   ├── trace/                   # NPM 包使用追踪
│   │   ├── api_tracer/              # API 调用链追踪
│   │   ├── component_deps/          # 组件依赖分析
│   │   └── component_deps/       # 组件依赖分析 v2（基于 manifest）
│   │
│   └── ts_bundle/                   # TypeScript 类型打包工具
│       ├── README.md                # 详细文档
│       ├── main.go                  # API 入口
│       ├── bundle.go                # 单类型打包器
│       ├── collect.go               # 依赖收集器
│       └── batch_collect.go         # 批量收集器
│
├── tsmorphgo/                       # ts-morph 风格的 API 封装
│   ├── README.md                    # API 文档
│   ├── project.go                   # 项目 API
│   ├── sourcefile.go                # 源文件 API
│   ├── node.go                      # 节点 API
│   ├── symbol.go                    # 符号 API
│   ├── references.go                # 引用查找 API
│   └── examples/                    # 使用示例
│
└── typescript-go/                   # TypeScript 官方解析器子模块
```

---

## 🎯 适用场景

### 代码质量提升

- **类型安全性改进**: 统计 `any` 类型使用，逐步改进类型定义
- **死代码清理**: 查找未使用的导出和文件，减少代码维护成本
- **依赖健康检查**: 检查隐式依赖、未使用依赖和过期依赖

### 项目重构

- **影响分析**: 查找调用者，了解重构影响范围
- **依赖关系追踪**: 追踪 NPM 包使用，评估替换/移除影响
- **组件依赖分析**: 优化组件结构，减少循环依赖

### 微服务架构

- **类型共享**: 提取 API 类型，为其他服务生成类型定义
- **API 调用链追踪**: 文档化 API 使用情况
- **批量类型打包**: 为多个服务生成类型定义

### CI/CD 集成

- **质量门禁**: 设置代码质量标准，阻止低质量代码合并
- **自动化报告**: 在每次构建后生成分析报告
- **持续监控**: 跟踪代码质量趋势
- **影响范围检查**: 使用 `impact` 命令在 MR/PR 时自动评估变更影响

### 代码变更影响分析

- **Code Review 辅助**: 在 Review 前了解变更的完整影响范围
- **回归测试范围**: 基于影响分析确定需要回归测试的模块
- **风险评估**: 根据影响层级和风险等级决定是否需要额外测试
- **发布决策**: 评估组件库变更对下游项目的影响
- **重构规划**: 使用 `component-deps-v2` 和 `impact` 命令规划重构策略

### 大型项目迁移

- **JavaScript → TypeScript**: 统计类型使用情况，追踪迁移进展
- **构建工具迁移**: 分析模块依赖，规划迁移策略

---

## 📚 更多资源

### 核心文档
- **[架构详解](./analyzer/README.md)**: 深入了解核心解析引擎
- **[代码影响分析管道](./pkg/pipeline/README.md)**: Pipeline 架构设计与数据流向
- **[Pipeline 接入文档](./pkg/pipeline/INTEGRATION.md)**: 业务方接入指南

### 插件开发
- **[插件开发指南](./analyzer_plugin/project_analyzer/README.md)**: 开发自定义分析器
- **[component_deps 文档](./analyzer_plugin/project_analyzer/component_deps/README.md)**: 组件依赖分析 v2

### API 文档
- **[ts_bundle 文档](./analyzer_plugin/ts_bundle/README.md)**: 类型打包工具详解
- **[TSMorphGo API](./tsmorphgo/README.md)**: ts-morph 风格的 Go API

---

## 📄 License

MIT License - 详见 [LICENSE](LICENSE) 文件

---

## 🤝 贡献

欢迎贡献代码、提出问题或建议！

---

<div align="center">

**[⬆ 返回顶部](#analyzer-ts)**

Made with ❤️ by [Flying-Bird1999](https://github.com/Flying-Bird1999)

</div>

