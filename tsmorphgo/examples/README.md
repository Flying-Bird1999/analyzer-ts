# TSMorphGo Examples

🎯 **基于真实React项目的TSMorphGo使用示例集合**，展示了如何使用TSMorphGo进行TypeScript代码分析。所有示例都基于真实的前端项目，确保实用性和准确性。

## 📁 目录结构

```
examples/
├── README.md                      # 📖 使用说明（本文档）
├── run-examples.sh               # 🚀 推荐的Shell脚本构建工具
├── Makefile                       # 🔧 备选的Makefile构建工具
├── BUILD_SYSTEM_GUIDE.md          # 📚 构建系统对比指南
├── demo-react-app/                # 📁 真实React项目（用作分析素材）
│   ├── src/                       # 源代码目录
│   │   ├── components/            # React组件
│   │   ├── hooks/                 # 自定义Hook
│   │   ├── types/                 # 类型定义
│   │   ├── services/              # 服务模块
│   │   └── forms/                 # 表单组件
│   └── public/                    # 静态资源
├── basic-usage/                   # 📦 基础API使用示例
│   ├── project-management.go      # 🏗️ 项目管理示例
│   ├── node-navigation.go         # 🔍 节点导航示例
│   └── type-detection.go          # 🏷️ 类型检测示例
└── advanced-usage/                 # ⚡ 高级API使用示例
    ├── reference-finding.go       # 🔗 引用查找示例
    └── specialized-apis.go         # 🛠️ 专用API示例
```

## 🚀 快速开始

### 🎯 方法1：使用Shell脚本（推荐）

Shell脚本提供了丰富的功能、美观的输出和强大的错误处理能力。

```bash
# 📋 查看所有可用命令
./run-examples.sh help

# 🚀 运行所有示例
./run-examples.sh all

# 📦 分类运行示例
./run-examples.sh basic      # 运行基础示例
./run-examples.sh advanced   # 运行高级示例

# 🎯 运行单个示例
./run-examples.sh project-management
./run-examples.sh node-navigation
./run-examples.sh type-detection
./run-examples.sh reference-finding
./run-examples.sh specialized-apis

# 🛠️ 开发工具
./run-examples.sh check       # 检查环境配置
./run-examples.sh install     # 安装项目依赖
./run-examples.sh report      # 生成项目报告
./run-examples.sh clean       # 清理临时文件
```

### 🔧 方法2：使用Makefile（备选）

如果您更喜欢传统的构建系统，也可以使用改进的Makefile：

```bash
# 📋 查看所有可用命令
make help

# 🚀 运行所有示例
make all

# 📦 分类运行示例
make basic      # 运行基础示例
make advanced   # 运行高级示例

# 🎯 运行单个示例
make project-management
make node-navigation
make type-detection
make reference-finding
make specialized-apis

# 🛠️ 开发工具
make deps        # 检查依赖
make clean       # 清理临时文件
make test        # 运行测试
```

### 🎬 方法3：直接运行（不推荐）

如果只想快速运行特定示例，也可以直接使用go命令：

```bash
# 运行项目管理示例
cd basic-usage
go run -tags project_management project-management.go

# 运行引用查找示例
cd advanced-usage
go run -tags reference_finding reference-finding.go
```

## 📋 示例详细介绍

### 🏗️ 基础示例 (basic-usage/)

#### 1. 项目管理示例 (`project-management.go`)
- **功能**: 展示项目创建、文件分析、动态添加文件
- **API**: `tsmorphgo.NewProject()`, `project.GetSourceFiles()`, `project.CreateSourceFile()`
- **输出**: 统计项目文件数量、分类、创建动态文件

#### 2. 节点导航示例 (`node-navigation.go`)
- **功能**: 展示AST节点遍历、父子关系查找、React组件结构分析
- **API**: `node.ForEachDescendant()`, `node.GetParent()`, `node.GetAncestors()`
- **输出**: 遍历函数声明、查找特定标识符、分析组件结构

#### 3. 类型检测示例 (`type-detection.go`)
- **功能**: 展示TypeScript类型识别、接口分析、导入导出统计
- **API**: `node.Kind`, `node.IsInterfaceDeclaration()`, `node.IsEnumDeclaration()`
- **输出**: 类型统计、接口分析、导入导出信息

### ⚡ 高级示例 (advanced-usage/)

#### 1. 引用查找示例 (`reference-finding.go`)
- **功能**: 展示符号引用查找、性能缓存优化、跳转到定义
- **API**: `tsmorphgo.FindReferences()`, `tsmorphgo.FindReferencesWithCache()`, `tsmorphgo.GotoDefinition()`
- **输出**: 引用列表、缓存性能对比、定义跳转结果

#### 2. 专用API示例 (`specialized-apis.go`)
- **功能**: 展示特定语法结构的深度分析、函数声明、调用表达式
- **API**: `tsmorphgo.IsFunctionDeclaration()`, `tsmorphgo.IsCallExpression()`, `node.Kind`
- **输出**: 函数分析、方法调用统计、属性访问统计

## 💻 运行结果示例

### 基础示例输出
```
🏗️ TSMorphGo 项目管理示例
==================================================

📁 示例1: 基于文件系统创建项目
真实React项目包含 14 个TypeScript文件:
  - App.tsx
  - index.ts
  - types.ts
  ...

项目文件分类:
  - Components: 3 个
  - Utils/Services: 2 个
  - Types: 3 个
  - Other: 6 个

✅ 项目管理示例完成!
```

### 高级示例输出
```
🔗 TSMorphGo 引用查找示例
==================================================

🔍 示例1: 基础引用查找
找到 16 个APP_CONFIG引用:
  1. /src/config/app.ts:3 - APP_CONFIG = {...
  2. /src/services/api.ts:6 - private config = APP_CONFIG;...

⚡ 示例2: 带缓存的引用查找性能对比
第一次查找:
  - 耗时: 1.302ms
  - 来源: LSP服务
  - 引用数: 16
第二次查找:
  - 耗时: 137µs
  - 来源: 缓存
  - 引用数: 16
  - 性能提升: 9.5x 倍

✅ 引用查找示例完成!
```

## 🛠️ 技术栈

- **TSMorphGo**: 核心 TypeScript 分析库
- **typescript-go**: AST 解析和遍历引擎
- **Go 1.19+**: 编程语言和构建工具
- **Go Build Tags**: 条件编译，支持不同示例
- **Shell Script**: 现代化构建和自动化工具

## 💡 使用建议

### 🔰 对于新手用户
1. **从基础开始**: 先运行 `./run-examples.sh basic` 了解基本概念
2. **逐个学习**: 使用单个示例命令深入学习每个API
3. **查看输出**: 仔细分析每个示例的输出，理解TSMorphGo的功能
4. **环境检查**: 使用 `./run-examples.sh check` 确保环境正确配置

### 🚀 对于开发者
1. **性能测试**: 观察引用查找的缓存效果和性能提升
2. **真实项目**: 所有示例都基于真实React项目，可直接应用到实际工作
3. **扩展功能**: 参考示例代码，开发自己的分析工具
4. **集成CI/CD**: 使用脚本功能进行自动化测试和部署

### 🏢 对于团队协作
1. **统一工具**: 推荐团队成员使用相同的构建工具
2. **文档同步**: 参考本文档和代码注释，保持知识同步
3. **版本兼容**: 确保Go版本和依赖库版本一致

## 📊 示例统计

| 类别 | 示例数量 | 主要功能 |
|------|----------|----------|
| **基础示例** | 3个 | 项目管理、节点导航、类型检测 |
| **高级示例** | 2个 | 引用查找、专用API |
| **总计** | 5个 | 覻盖TSMorphGo主要功能 |

## 🔧 环境要求

### 必需
- **Go 1.19+**: 运行环境
- **Git**: 版本控制（可选）
- **Terminal**: 命令行终端

### 项目文件
- `demo-react-app/`: 真实React项目（14个TypeScript文件）
- Go模块依赖：自动通过 `go mod` 管理

## 📚 扩展资源

- **构建系统指南**: [BUILD_SYSTEM_GUIDE.md](BUILD_SYSTEM_GUIDE.md) - Makefile vs Shell Script 对比
- **TSMorphGo文档**: 项目根目录的API文档
- **TypeScript规范**: TypeScript官方文档

## 🐛 问题反馈

遇到问题时，请按以下步骤排查：

1. **检查环境**: 运行 `./run-examples.sh check` 检查配置
2. **查看日志**: 仔细阅读错误信息和输出
3. **检查文件**: 确认示例文件存在且可读
4. **提交Issue**: [GitHub Issues](https://github.com/Flying-Bird1999/analyzer-ts/issues)

## 🎉 开始使用

```bash
# 1. 检查环境
./run-examples.sh check

# 2. 运行基础示例
./run-examples.sh basic

# 3. 运行高级示例
./run-examples.sh advanced

# 4. 查看项目报告
./run-examples.sh report
```

---

✨ **使用 TSMorphGo 构建强大的TypeScript代码分析工具！** 🚀