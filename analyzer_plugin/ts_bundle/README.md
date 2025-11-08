# TypeScript 类型声明打包工具 (ts_bundle)

一个强大的 TypeScript 类型声明打包工具，用于从 TypeScript 项目中提取、打包和类型声明。支持单类型打包和批量类型处理，能够智能处理复杂的类型依赖关系和命名冲突。

## 🚀 功能特性

### 核心功能

- **智能依赖收集**: 自动递归收集类型的所有依赖项
- **命名冲突解决**: 智能处理跨文件的同名类型冲突
- **类型别名支持**: 支持为类型定义别名，便于自定义输出
- **两种输出模式**: 单文件合并和批量独立文件输出

### TypeScript 特性支持

- **完整类型系统**: 支持 `interface`、`type`、`class`、`enum`
- **高级类型**: 正确处理 `Omit<>`、`Pick<>`、`keyof`、索引类型等
- **多种导入方式**:
  - 命名导入: `import { Type } from './file'`
  - 默认导入: `import Type from './file'`
  - 命名空间导入: `import * as ns from './file'`
  - 重新导出: `export { Type } from './file'`
- **路径别名解析**: 完整支持 tsconfig.json 中的 `paths` 配置
- **NPM 包支持**: 能够解析 node_modules 中的类型定义

## 📦 安装和构建

```bash
# 确保依赖完整
go mod tidy

# 构建主程序
go build -o analyzer-ts

# 或直接运行
go run main.go
```

## 🔧 使用方法

### 单类型打包 (`bundle` 命令)

从单个入口文件收集指定类型及其所有依赖，输出到一个文件：

```bash
analyzer-ts bundle \
  -i ./src/index.ts \           # 入口文件
  -t MyType \                   # 要打包的类型名
  -o ./dist/types.d.ts          # 输出文件
```

**命令参数:**

- `-i, --input`: 入口文件路径（必需）
- `-t, --type`: 要分析的根类型名称（必需）
- `-o, --output`: 输出文件路径（默认：`./output.ts`）
- `-r, --root`: 项目根路径（可选，自动检测）

### 批量类型打包 (`batch-bundle` 命令)

批量处理多个类型，每个类型生成独立的 `.d.ts` 文件，完美解决命名冲突：

```bash
# 基础批量打包
analyzer-ts batch-bundle \
  -e "./src/user.ts:User" \
  -e "./src/product.ts:Product" \
  --output-dir ./dist/types/

# 支持逗号分隔（简写形式）
analyzer-ts batch-
bundle \
  -e "./src/user.ts:User,./src/product.ts:Product" \
  --output-dir ./dist/types/

# 带别名 - 自定义输出类型名
analyzer-ts batch-bundle \
  -e "./src/user.ts:User:UserDTO" \
  -e "./src/common.ts:CommonType:ConfigType" \
  --output-dir ./dist/types/
```

**命令参数:**

- `-e, --entries`: 入口点列表，格式为 `文件路径:类型名[:别名]`（必需，可多次使用或逗号分隔）
- `--output-dir`: 输出目录路径（必需，每个类型生成独立文件）
- `-r, --root`: 项目根路径（可选，自动检测）

## 📝 入口点格式

批量模式支持以下格式：

```bash
# 基础格式：文件路径:类型名
./src/user.ts:User

# 带别名格式：文件路径:类型名:别名
./src/user.ts:User:UserDTO

# 别名作用：输出的类型会被重命名
# 原: interface User { ... }
# 输出: interface UserDTO { ... }
```

## 🎯 使用场景

### 1. API 类型打包

```bash
# 将复杂的 API 类型及其依赖打包为单一文件
analyzer-ts bundle -i ./src/api/user.ts -t UserProfile -o ./types/user.d.ts
```

### 2. 微服务类型共享

```bash
# 批量导出多个服务类型，每个独立文件
analyzer-ts batch-bundle \
  -e "./services/user.ts:User:UserDTO" \
  -e "./services/product.ts:Product:ProductDTO" \
  -e "./services/order.ts:Order:OrderDTO" \
  --output-dir ./shared-types/
```

### 3. 库开发类型导出

```bash
# 为 TypeScript 库生成类型定义文件
analyzer-ts bundle -i ./src/index.ts -t Library -o ./dist/index.d.ts
```

## 🏗️ 架构设计

### 核心组件

1. **CollectResult**: 类型依赖收集器

   - 递归分析文件依赖
   - 支持复杂的模块解析
   - 缓存优化性能
2. **TypeBundler**: 类型打包器

   - 智能命名冲突解决
   - 精确的类型引用更新
   - 支持批量处理
3. **BatchCollectResult**: 批量收集器

   - 文件级缓存优化
   - 支持类型别名
   - 独立文件输出

### 设计特点

- **分离式处理架构**: 先收集，后检测冲突，最后统一更新
- **全局唯一性保证**: 所有类型最终名称唯一，避免命名污染
- **上下文感知的引用更新**: 区分类型声明和类型引用
- **高性能缓存**: 文件级缓存避免重复解析

## 🧪 测试

运行所有测试：

```bash
go test -v
```

运行特定测试：

```bash
# 测试批量功能
go test -v -run "TestGenerateBatch"

# 测试文件输出功能
go test -v -run "TestGenerateBatchBundlesToFiles"
```

运行示例：

```bash
go run example_simple.go
```

## 📁 项目结构

```
analyzer_plugin/ts_bundle/
├── README.md                   # 本文档
├── main.go                     # 主要 API 入口
├── bundle.go                   # TypeBundler 类型打包器
├── collect.go                  # CollectResult 收集器
├── batch_collect.go            # BatchCollectResult 批量收集器
├── cmd/
│   └── bundle.go               # 命令行接口
├── collect_test.go             # 单类型测试
├── batch_bundle_test.go        # 批量功能测试
├── example_simple.go           # 使用示例
└── testdata/                   # 测试数据
```

## 🔍 支持的 TypeScript 特性

### ✅ 完全支持

- [X] 基础类型声明（interface, type, class, enum）
- [X] 类型继承和扩展（extends）
- [X] 泛型类型
- [X] 工具类型（Omit, Pick, Record 等）
- [X] 条件类型
- [X] 映射类型
- [X] 索引类型和索引访问
- [X] 联合类型和交叉类型
- [X] 模板字面量类型
- [X] 各种导入导出语法
- [X] 路径别名和 baseUrl 配置
- [X] NPM 包类型解析
- [X] .d.ts 全局类型声明

### 🔄 部分支持

- [ ] 动态 import()（实验性支持）
- [ ] 命名空间合并（有限支持）

## 🚨 注意事项

1. **文件路径**: 使用绝对路径或相对于项目根目录的路径
2. **类型存在性**: 如果指定的类型不存在，会跳过而不会报错
3. **循环依赖**: 工具能检测并安全处理循环依赖
4. **性能优化**: 大型项目建议使用批量模式，利用缓存优化

## 🤝 贡献

欢迎提交 Issue 和 Pull Request 来改进这个工具！

## 📄 许可证

本项目采用与主项目相同的许可证。
