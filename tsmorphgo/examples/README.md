# TSMorphGo Examples

🎯 **基于真实React项目的TSMorphGo使用示例集合**，展示了如何使用TSMorphGo进行TypeScript代码分析。所有示例都基于真实的前端项目，确保实用性和准确性。

## 📁 目录结构

```
examples/
├── README.md                      # 📖 使用说明（本文档）
├── run-examples.sh               # 🚀 运行脚本
├── demo-react-app/                # 📁 真实React项目（用作分析素材）
│   ├── src/                       # 源代码目录
│   │   ├── components/            # React组件
│   │   ├── hooks/                 # 自定义Hook
│   │   ├── types/                 # 类型定义
│   │   ├── services/              # 服务模块
│   │   └── forms/                 # 表单组件
│   └── public/                    # 静态资源
└── pkg/                           # 📦 统一的示例目录
    ├── project-management.go      # 🏗️ 项目管理和内存文件系统示例
    ├── node-navigation.go         # 🔍 节点导航和位置信息示例
    ├── type-detection.go          # 🏷️ 类型检测和代码质量分析示例
    ├── reference-finding.go       # 🔗 引用查找和符号系统示例
    └── specialized-apis.go        # 🛠️ 专用API示例
```

## 🚀 快速开始

### 🎯 方法1：使用Shell脚本（推荐）

Shell脚本提供了丰富的功能、美观的输出和强大的错误处理能力。

```bash
# 📋 查看所有可用命令
./run-examples.sh help

# 🚀 运行所有示例（推荐）
./run-examples.sh all         # 运行所有5个示例

# 🎯 运行单个示例
./run-examples.sh project-management      # 项目管理和内存文件系统
./run-examples.sh node-navigation         # 节点导航和位置信息
./run-examples.sh type-detection          # 类型检测和代码质量分析
./run-examples.sh reference-finding       # 引用查找和符号系统
./run-examples.sh specialized-apis        # 专用API

# 🛠️ 开发工具
./run-examples.sh check       # 检查环境配置
./run-examples.sh install     # 安装项目依赖
./run-examples.sh report      # 生成项目报告
./run-examples.sh clean       # 清理临时文件
```

### 🎬 方法2：直接运行（不推荐）

如果只想快速运行特定示例，也可以直接使用go命令：

```bash
# 进入pkg目录
cd pkg

# 运行单个示例
go run -tags project_management project-management.go
go run -tags node_navigation node-navigation.go
go run -tags type_detection type-detection.go
go run -tags reference_finding reference-finding.go
go run -tags specialized_apis specialized-apis.go
```

## 📋 示例详细介绍

### 1. 项目管理和内存文件系统示例 (`project-management.go`)

**功能**：演示项目创建、源文件管理、内存文件系统和动态文件创建

**涵盖API**：
- `NewProject()` - 基于真实项目创建
- `NewProjectFromSources()` - 内存文件系统
- `CreateSourceFile()` - 动态文件创建
- `GetSourceFiles()` - 文件枚举
- `Close()` - 资源管理

**应用场景**：
- 测试环境搭建
- 动态代码生成
- 原型开发
- 配置文件管理

**学习级别**：初级 → 高级 | **预计时间**：30-45分钟

### 2. 节点导航和位置信息示例 (`node-navigation.go`)

**功能**：演示AST节点遍历、位置信息计算和IDE集成

**涵盖API**：
- `ForEachDescendant()` - 节点遍历
- `GetParent()` / `GetAncestors()` - 父节点导航
- `GetFirstAncestorByKind()` - 条件祖先查找
- `GetStart()` / `GetStartLineNumber()` / `GetStartColumnNumber()` - 位置信息
- `GetPositionInfo()` - 完整位置信息

**应用场景**：
- 代码高亮
- 错误定位
- 跳转定义
- IDE插件开发

**学习级别**：初级 → 高级 | **预计时间**：40-60分钟

### 3. 类型检测和代码质量分析示例 (`type-detection.go`)

**功能**：演示TypeScript类型识别、代码质量分析和依赖关系分析

**涵盖API**：
- `IsInterfaceDeclaration()` / `IsTypeAliasDeclaration()` - 类型识别
- `IsFunctionDeclaration()` / `IsVariableDeclaration()` - 声明检测
- `IsCallExpression()` / `IsPropertyAccessExpression()` - 表达式分析
- `node.Kind == KindXxx` - 精确类型匹配
- 复合类型守卫 - 复杂模式识别

**应用场景**：
- 静态代码分析
- 重构工具
- 代码质量检查
- 依赖关系图

**学习级别**：初级 → 高级 | **预计时间**：35-50分钟

### 4. 引用查找和符号系统示例 (`reference-finding.go`)

**功能**：演示符号引用查找、缓存优化、跳转定义和重命名安全分析

**涵盖API**：
- `FindReferences()` - 引用查找
- `FindReferencesWithCache()` - 缓存优化
- `GotoDefinition()` - 跳转定义
- `GetSymbol()` / `symbol.GetName()` - 符号系统
- 重命名安全性分析 - 影响范围评估

**应用场景**：
- IDE功能开发
- 重构工具
- 代码导航
- 影响分析

**学习级别**：中级 → 高级 | **预计时间**：45-60分钟

### 5. 专用API示例 (`specialized-apis.go`)

**功能**：演示特定语法结构的分析，包括函数声明、调用表达式、属性访问和导入别名

**涵盖API**：
- `IsFunctionDeclaration()` / `GetFunctionDeclarationNameNode()` - 函数声明
- `IsCallExpression()` / `GetCallExpressionExpression()` - 调用表达式
- `IsPropertyAccessExpression()` / `GetPropertyAccessName()` - 属性访问
- `IsVariableDeclaration()` / `GetVariableName()` - 变量声明
- `IsImportSpecifier()` / `GetImportSpecifierAliasNode()` - 导入别名
- `IsBinaryExpression()` / `GetBinaryExpressionLeft/Right/OperatorToken()` - 二元表达式

**应用场景**：
- 语法分析器
- 代码转换器
- AST处理工具
- 特定模式识别

**学习级别**：中级 → 高级 | **预计时间**：40-55分钟

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

## 🎯 学习路径

### 🔰 初学者路径

1. **项目管理** → 理解项目创建和基本概念
2. **节点导航** → 学习AST遍历和位置信息
3. **类型检测** → 掌握类型识别和基础分析

### 🚀 进阶路径

4. **引用查找** → 学习符号系统和语义分析
5. **专用API** → 掌握特定语法结构的处理

### 💼 专业开发者

- **性能优化**：学习缓存机制和批量处理
- **IDE集成**：掌握位置计算和链接生成
- **代码质量**：理解静态分析和质量检查
- **重构工具**：学习影响分析和安全操作

### 💡 使用建议

#### 对于新手用户

1. **从基础开始**: 先运行 `./run-examples.sh all` 了解整体功能
2. **逐个学习**: 使用单个示例命令深入学习每个API
3. **查看输出**: 仔细分析每个示例的输出，理解TSMorphGo的功能
4. **环境检查**: 使用 `./run-examples.sh check` 确保环境正确配置

#### 对于开发者

1. **性能测试**: 观察引用查找的缓存效果和性能提升
2. **真实项目**: 所有示例都基于真实React项目，可直接应用到实际工作
3. **扩展功能**: 参考示例代码，开发自己的分析工具
4. **集成CI/CD**: 使用脚本功能进行自动化测试和部署

#### 对于团队协作

1. **统一工具**: 推荐团队成员使用相同的构建工具
2. **文档同步**: 参考本文档和代码注释，保持知识同步
3. **版本兼容**: 确保Go版本和依赖库版本一致

## 📊 示例统计

| 类别               | 示例数量 | 主要功能                     |
| ------------------ | -------- | ---------------------------- |
| **基础示例** | 3个      | 项目管理、节点导航、类型检测 |
| **高级示例** | 2个      | 引用查找、专用API            |
| **总计**     | 5个      | 覻盖TSMorphGo主要功能        |

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

# 2. 运行所有示例
./run-examples.sh all

# 3. 查看项目状态
./run-examples.sh status

# 4. 查看项目报告
./run-examples.sh report
```

---

## 🏗️ API对齐说明

所有示例都严格按照 [ts-morph API](../doc/ts-morph.md) 设计，确保：

1. **命名一致性**：Go版本API与ts-morph保持相似的命名模式
2. **功能对齐**：每个API都对应ts-morph中的相应功能
3. **使用模式**：演示正确的使用姿势和最佳实践
4. **错误处理**：展示适当的错误处理和边界情况

### API映射示例

| ts-morph API | TSMorphGo API | 说明 |
|-------------|---------------|------|
| `node.isFunctionDeclaration()` | `IsFunctionDeclaration(node)` | 函数声明检测 |
| `functionDeclaration.getName()` | `GetFunctionDeclarationNameNode(node)` | 获取函数名 |
| `node.forEachDescendant()` | `node.ForEachDescendant()` | 节点遍历 |
| `node.getParent()` | `node.GetParent()` | 获取父节点 |
| `identifier.findReferences()` | `FindReferences(node)` | 引用查找 |

## 📝 注意事项

1. **项目路径**：所有示例都基于 `demo-react-app` 目录
2. **构建标签**：运行时需要指定正确的构建标签
3. **资源管理**：使用 `defer project.Close()` 确保资源释放
4. **错误处理**：检查返回值，处理不存在的文件或符号
5. **性能考虑**：大项目分析时注意内存和CPU使用

---

✨ **使用 TSMorphGo 构建强大的TypeScript代码分析工具！** 🚀
