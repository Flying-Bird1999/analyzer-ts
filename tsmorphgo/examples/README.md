# TSMorphGo 示例总览

本目录包含了 TSMorphGo 的完整使用示例，基于真实的 React 项目演示了 TypeScript 代码分析的各种场景。

## 📁 示例文件列表

| 示例文件 | 功能描述 | 验证目标 | 运行方式 |
|---------|---------|---------|---------|
| [`basic_usage.go`](./basic_usage.go) | 基础项目操作和节点查找 | 项目初始化、多种节点查找方式 | `go run -tags=examples basic_usage.go` |
| [`node_navigation.go`](./node_navigation.go) | 节点导航和类型收窄 | 节点关系、类型安全的API访问 | `go run -tags=examples node_navigation.go` |
| [`parser_data.go`](./parser_data.go) | 透传API验证 | 解析数据获取和使用 | `go run -tags=examples parser_data.go` |
| [`comprehensive_verification.go`](./comprehensive_verification.go) | 综合API验证 | 导入声明的完整API链 | `go run -tags=examples comprehensive_verification.go` |
| [`path_aliases.go`](./path_aliases.go) | 路径别名解析 | tsconfig.json路径别名配置 | `go run -tags=examples path_aliases.go` |
| [`references.go`](./references.go) | 综合引用查找 | Hook函数、类型、工具函数的引用分析 | `go run -tags=examples references.go` |

## 🚀 快速开始

### 运行单个示例

```bash
# 运行基础使用示例
go run -tags=examples basic_usage.go

# 运行节点导航示例
go run -tags=examples node_navigation.go

# 运行引用查找示例
go run -tags=examples references.go
```

### 运行所有示例

```bash
# 使用脚本运行所有示例
./run-all-examples.sh
```

## 📊 验证结果总览

所有示例均已通过完整验证：

### ✅ 基础功能验证
- **项目初始化**: 成功读取 tsconfig.json 并扫描源文件
- **节点查找**: 验证了多种节点查找方式（遍历、位置查找等）
- **类型判断**: 正确识别各种 AST 节点类型
- **位置信息**: 准确获取节点的行列号和字符偏移

### ✅ 高级功能验证
- **节点导航**: 成功实现父子节点关系遍历
- **类型收窄**: 安全的类型转换和专有API访问
- **透传数据**: 获取原生解析器的详细信息
- **引用查找**: 跨文件的符号引用分析

### ✅ 实际应用场景
- **路径别名**: 解析 tsconfig.json 中的 paths 配置
- **Hook分析**: React Hook 函数的定义和使用查找
- **类型分析**: 接口定义和类型引用的完整追踪
- **工具函数**: 跨文件的工具函数调用分析
- **综合引用**: 三种引用类型的统一演示 (Hook函数、类型、工具函数)

## 🎯 核心验证结果

- **6个示例全部通过验证** ✅
- **总计发现引用**: 11个（包含Hook函数、类型、工具函数引用）
- **路径别名配置**: 7个
- **路径别名使用**: 9处
- **跨文件引用**: 全部成功识别

## 📖 学习路径

建议按以下顺序学习：

1. **基础入门**: `basic_usage.go` → 了解项目初始化和基础操作
2. **节点操作**: `node_navigation.go` → 学习节点导航和类型收窄
3. **数据获取**: `parser_data.go` → 理解透传API的使用
4. **综合应用**: `comprehensive_verification.go` → 掌握完整API链
5. **路径解析**: `path_aliases.go` → 学习路径别名处理
6. **引用分析**: `references.go` → 综合引用查找 (Hook函数、类型、工具函数)

## 🔧 技术特点

### 统一的代码结构
- 详细的中文注释说明验证目标
- 明确的预期输出描述
- 统一的错误处理和资源清理
- 动态路径构建，避免硬编码

### 完整的API覆盖
- 项目操作：`NewProject`, `GetSourceFile`, `GetSourceFiles`
- 节点遍历：`ForEachDescendant`, `ForEachChild`
- 类型判断：`IsKind`, `IsXxx`, `AsXxx`
- 位置信息：`GetStartLineNumber`, `GetStartColumnNumber`
- 引用查找：`FindReferences`, `GetSymbol`
- 透传数据：`HasParserData`, `GetParserData`

### 实际项目驱动
所有示例都基于真实的 React 项目结构：
- React 组件 (App.tsx, UserProfile.tsx, ProductCard.tsx)
- 自定义 Hooks (useUserData.ts)
- 工具函数 (helpers.ts, dateUtils.ts)
- 类型定义 (types.ts)
- 路径别名配置 (tsconfig.json)

## 📝 详细文档

完整的技术方案文档请参考：[TSMorphGo_示例重构技术方案.md](./TSMorphGo_示例重构技术方案.md)

## 🤝 贡献指南

如果你想要添加新的示例或改进现有示例：

1. 确保示例基于真实的 demo-react-app 项目
2. 提供明确的验证目标和预期输出
3. 添加详细的中文注释
4. 遵循统一的代码结构
5. 确保代码能够成功编译和运行

## 📄 许可证

本示例代码遵循与主项目相同的许可证。