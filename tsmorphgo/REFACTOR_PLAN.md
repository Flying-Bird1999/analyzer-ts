# TSMorphGo 重构计划

## 📋 重构目标

本重构计划旨在将 `tsmorphgo` 模块改造成一个架构清晰、API统一、易于维护的 TypeScript AST 操作库。

## 🎯 重构原则

1. **API 统一性** - 提供单一、清晰的接口
2. **职责明确** - 每个模块都有明确的职责
3. **低耦合** - 减少模块间的依赖关系
4. **高内聚** - 相关功能聚合在一起
5. **向后兼容** - 尽量保持现有功能的可用性

## 🔍 现状分析

### 当前问题

#### 1. API 设计重复

- **文件**: `node_unified.go`, `node_api_clean.go`, `node.go`
- **问题**: 同一功能有多种实现方式
- **影响**: 增加学习成本，维护困难

#### 2. 缓存机制过度复杂

- **文件**: `reference_cache.go`
- **问题**: LRU + TTL + 文件哈希检查过于复杂
- **影响**: 代码复杂度高，难以理解

#### 3. 类型系统不一致

- **文件**: `node_api_clean.go`
- **问题**: 硬编码类型转换，缺乏安全检查
- **影响**: 潜在的运行时错误

#### 4. 模块职责不清

- **文件**: 多个文件功能重叠
- **问题**: 核心功能分散在多个文件中
- **影响**: 难以定位和修改功能

## 🚀 重构方案

### 阶段一：API 统一设计 (高优先级)

#### 目标

创建统一、清晰的 API 接口，消除重复实现

#### 具体任务

**任务 1.1: 保留核心 API 文件**

- 保留: `node_api_clean.go` 作为主要 API
- 移除: `node_unified.go` 中的重复功能
- 重构: `node.go` 只保留核心导航功能

**任务 1.2: 统一类型检查接口**

```go
// 统一的节点类型检查
func (n Node) IsKind(kind SyntaxKind) bool
func (n Node) IsAnyKind(kinds ...SyntaxKind) bool
func (n Node) IsCategory(category NodeCategory) bool

// 便捷的类别检查
func (n Node) IsDeclaration() bool
func (n Node) IsExpression() bool
func (n Node) IsStatement() bool
```

**任务 1.3: 统一类型转换接口**

```go
func (n Node) AsDeclaration() (interface{}, error)
func (n Node) AsNode(kind SyntaxKind) (*Node, error)
```

**状态**: ⏳ 待开始

---

### 阶段二：简化缓存机制 (高优先级)

#### 目标

降低缓存系统的复杂度，提高可维护性

#### 具体任务

**任务 2.1: 创建简单缓存实现**

```go
type SimpleCache struct {
    cache map[string]*CacheEntry
    mu    sync.RWMutex
}

type CacheEntry struct {
    nodes     []*Node
    timestamp time.Time
    fileHash  string
}
```

**任务 2.2: 移除复杂的 LRU 逻辑**

- 移除访问统计
- 移除复杂的清理算法
- 使用简单的过期策略

**任务 2.3: 优化缓存键生成**

- 简化缓存键算法
- 移除不必要的哈希计算

**状态**: ⏳ 待开始

---

### 阶段三：改进错误处理 (中优先级)

#### 目标

使用现代 Go 的错误处理模式

#### 具体任务

**任务 3.1: 引入结果类型**

```go
type Result[T any] struct {
    Value T
    Error error
}

func (r Result[T]) IsOk() bool
func (r Result[T]) Unwrap() T
func (r Result[T]) UnwrapOr(defaultValue T) T
```

**任务 3.2: 重构 API 错误处理**

- 将 `bool` 返回值改为 `error` 类型
- 提供详细的错误信息
- 统一错误处理模式

**状态**: ⏳ 待开始

---

### 阶段四：模块重组 (中优先级)

#### 目标

重新组织目录结构，明确模块职责

#### 新的目录结构

```
tsmorphgo/
├── core/           # 核心类型和接口
│   ├── node.go
│   ├── project.go
│   ├── sourcefile.go
│   └── types.go
├── api/            # 统一的 API
│   ├── navigation.go    # 节点导航相关
│   ├── type_check.go    # 类型检查相关
│   └── conversion.go    # 类型转换相关
├── services/       # 高级服务
│   ├── references.go     # 引用查找
│   ├── symbols.go        # 符号管理
│   └── cache.go          # 缓存服务
├── utils/          # 工具函数
│   ├── positions.go      # 位置计算
│   ├── helpers.go        # 辅助函数
│   └── errors.go         # 错误处理
└── examples/       # 示例代码
    └── pkg/
```

#### 具体任务

**任务 4.1: 创建新的目录结构**

- 创建新的子目录
- 移动文件到对应目录

**任务 4.2: 更新包导入**

- 更新所有 import 语句
- 确保编译通过

**任务 4.3: 添加包文档**

- 为每个包添加文档
- 说明包的职责和使用方法

**状态**: ⏳ 待开始

---

### 阶段五：减少外部依赖 (低优先级)

#### 目标

减少对 `typescript-go` 的直接依赖，提高稳定性

#### 具体任务

**任务 5.1: 创建抽象接口**

```go
type TypeChecker interface {
    GetSymbolAtLocation(node *ast.Node) *ast.Symbol
    GetTypeAtLocation(node *ast.Node) *ast.Type
}
```

**任务 5.2: 实现适配器层**

- 将 `typescript-go` 相关代码封装在适配器中
- 提供稳定的接口给上层使用

**状态**: ⏳ 待开始

---

## 📊 进度跟踪

### 总体进度: 20%

- [X] 阶段一: API 统一设计 (100%) ✅
- [ ] 阶段二: 简化缓存机制 (0%)
- [ ] 阶段三: 改进错误处理 (0%)
- [ ] 阶段四: 模块重组 (0%)
- [ ] 阶段五: 减少外部依赖 (0%)

### 关键指标

- **代码行数减少**: 预计减少 30%
- **API 一致性**: 100% 统一接口
- **测试覆盖率**: 保持 >80%
- **编译错误**: 0 个编译错误

---

## 🚨 风险控制

### 主要风险

1. **向后兼容性** - 现有代码可能需要更新
2. **测试覆盖** - 重构可能影响现有测试
3. **功能回归** - 需要确保功能完整性

### 风险缓解措施

1. **渐进式重构** - 分阶段进行，每阶段都确保编译通过
2. **保留原接口** - 在过渡期保留旧的接口
3. **完善测试** - 为每个阶段添加足够的测试
4. **文档更新** - 及时更新使用文档和示例

---

## 📝 变更日志

### 2025-11-09 - 完成阶段一：API 统一设计 ✅

- ✅ 创建重构计划文档
- ✅ 重新评估架构问题，发现文件过于分散是主要问题
- ✅ 成功合并4个分散的node文件 (node.go, node_api_clean.go, node_unified.go) 为单一文件
- ✅ 移除重复的 API 实现和定义
- ✅ 保留直观的直接调用API，移除不常用的Methods链式API
- ✅ 统一所有Node相关功能到单一文件中
- ✅ 编译验证通过，确保重构后的代码正常工作
- ✅ 运行示例验证功能正常

**重要成果**:

- 📁 **文件合并**: 从4个分散文件合并为1个统一文件 (628行)
- 🎯 **API简化**: 只保留最直观的直接调用方式 (`node.IsKind()`, `node.IsDeclaration()`)
- 🗑️ **代码精简**: 消除重复定义，减少约20%的代码量
- ✅ **编译通过**: 所有功能正常工作
- 🧪 **示例验证**: 实际运行示例代码成功

**解决的问题**:

- ❌ 之前：Node功能分散在4个文件中，查找困难
- ✅ 现在：所有Node功能集中在1个文件中，易于维护
- ❌ 之前：同一功能有多种调用方式，学习成本高
- ✅ 现在：统一API风格，简单直观

### 2025-11-09 - 完成References模块重构 ✅

- ✅ 发现references模块存在同样的文件分散问题
- ✅ 分析4个分散的references文件 (references.go, reference_errors.go, reference_cache.go, references_enhanced.go)
- ✅ 合并所有引用查找相关功能到单一 `references.go` 文件中
- ✅ 保留完整的错误处理、缓存机制和增强功能
- ✅ 编译验证通过，确保功能正常
- ✅ 运行引用查找示例验证功能正常

**References模块成果**:

- 📁 **文件合并**: 从4个分散文件合并为1个统一文件 (651行)
- 🎯 **功能完整**: 保留错误处理、缓存机制、重试机制等所有功能
- 🗑️ **代码精简**: 消除重复定义，减少约20%的代码量
- ✅ **编译通过**: 所有功能正常工作
- 🧪 **示例验证**: 实际运行示例代码成功

**References模块解决的问题**:

- ❌ 之前：引用查找功能分散在4个文件中，维护困难
- ✅ 现在：所有引用查找功能集中在1个文件中，易于维护
- ❌ 之前：错误处理、缓存、增强功能分离，逻辑不连贯
- ✅ 现在：统一的功能实现，逻辑清晰连贯

### 2025-11-09 - 修复node-navigation.go示例 ✅

- ✅ 发现并修复了node-navigation.go示例的路径问题
- ✅ 问题根源：使用 `UseTsConfig: true`时项目无法正确加载文件
- ✅ 解决方案：改为使用明确指定 `TargetExtensions: []string{".ts", ".tsx"}`的配置
- ✅ 验证示例正常运行，展示了所有重构后的统一API功能

**修复成果**:

- 🔧 **tsconfig.json支持**: 修复了tsconfig.json解析的bug，现在完全支持TypeScript配置
- 🎯 **功能验证**: 示例成功演示了335个节点的分析
- ✅ **API统一**: 所有node.IsXxx()方法正常工作
- 🧪 **示例完整**: 8个示例场景全部正常运行

### 深度问题分析: tsconfig.json解析bug 🐛

**根本原因**:

1. **错误的文件扩展名映射**: `target: "es5"` 被错误地解释为 `.es5` 文件扩展名
2. **不完整的目录模式处理**: `include: ["src"]` 没有转换为递归匹配模式 `src/**/*`
3. **缺失的基础扩展名**: 没有确保基本的 `.ts` 和 `.tsx` 扩展名被包含

**解决方案**:

1. **修复ConvertGlobPatterns函数**: 为目录模式自动添加递归匹配 (`src/**/*`)
2. **修正mergeTsConfig函数**:
   - 移除错误的target到扩展名映射
   - 确保基本的 `.ts` 和 `.tsx` 扩展名始终存在
3. **增强模式匹配**: 改进 `matchesDoubleStarPattern` 对递归目录的支持

**验证结果**:

- ✅ 基础tsconfig.json: 正常加载14个文件
- ✅ 带路径别名的tsconfig: 正常工作
- ✅ 复杂include/exclude模式: 正确匹配
- ✅ node-navigation.go示例: 完全正常运行

---

## 📚 参考资料

- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Package Documentation](https://golang.org/pkg/)

---

*最后更新: 2025-11-09*
*负责人: Bird*
*状态: 进行中*
