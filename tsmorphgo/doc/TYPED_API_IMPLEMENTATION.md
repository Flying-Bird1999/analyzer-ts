# 特定节点类型专有 API 实现思路

## 概述

本文档总结了 tsmorphgo 特定节点类型专有 API 的实现思路和设计理念，该系统为 TypeScript AST 操作提供了类型安全的高级接口。

## 设计目标

### 核心目标
1. **类型安全**: 提供编译时类型检查，避免运行时错误
2. **需求导向**: 严格基于 ts-morph.md 文档中的实际需求场景
3. **简洁高效**: 移除冗余 API，专注于核心功能
4. **透传集成**: 无缝集成底层解析器数据
5. **易用性**: 提供直观的使用范式

### 文档场景基础
基于 ts-morph.md 中的 4 个核心场景设计：
- **场景 7.1**: CallExpression - 获取被调用表达式和参数
- **场景 7.2**: PropertyAccessExpression - 获取属性名和对象表达式
- **场景 7.3**: VariableDeclaration - 获取名称节点和初始值
- **场景 7.4**: FunctionDeclaration - 获取函数名节点

## 架构设计

### 1. 类型转换系统

```go
// 基础接口，统一所有特定类型的操作
type NodeWrapper interface {
    GetNode() *Node
    GetKind() SyntaxKind
}

// 类型安全转换模式
func (n *Node) AsVariableDeclaration() (*VariableDeclaration, bool) {
    if !n.IsVariableDeclaration() {
        return nil, false
    }
    return &VariableDeclaration{Node: n}, true
}
```

**设计原理**:
- 编译时类型检查：只有匹配的类型才能转换成功
- 返回布尔值表示转换是否成功，符合 Go 语言习惯
- 使用结构体嵌入模式，避免数据复制

### 2. 特定类型结构体设计

#### VariableDeclaration
```go
type VariableDeclaration struct {
    *Node  // 嵌入基础Node，继承所有基础功能
}
```

**核心方法**:
```go
func (v *VariableDeclaration) GetNameNode() *Node {
    // 查找第一个标识符子节点
    return v.Node.getFirstChildByKind(KindIdentifier)
}

func (v *VariableDeclaration) GetName() string {
    // 便利方法，自动处理空格等格式问题
    return strings.TrimSpace(v.GetNameNode().GetText())
}

func (v *VariableDeclaration) HasInitializer() bool {
    return v.GetInitializer() != nil
}

func (v *VariableDeclaration) GetInitializer() *Node {
    // 查找等号后的子节点作为初始值
    children := v.Node.GetChildren()
    for i, child := range children {
        if strings.TrimSpace(child.GetText()) == "=" {
            if i+1 < len(children) {
                return children[i+1]
            }
        }
    }
    return nil
}
```

#### CallExpression
```go
type CallExpression struct {
    *Node
}

func (c *CallExpression) GetExpression() *Node {
    // 第一个子节点是被调用的表达式
    children := c.Node.GetChildren()
    if len(children) > 0 {
        return children[0]
    }
    return nil
}

func (c *CallExpression) GetArguments() []*Node {
    // 跳过第一个子节点，其余为参数
    children := c.Node.GetChildren()
    if len(children) > 1 {
        return children[1:]
    }
    return nil
}

func (c *CallExpression) GetArgumentCount() int {
    return len(c.GetArguments())
}
```

#### PropertyAccessExpression
```go
type PropertyAccessExpression struct {
    *Node
}

func (p *PropertyAccessExpression) GetName() string {
    // 从右向左查找第一个标识符作为属性名
    children := p.Node.GetChildren()
    for i := len(children) - 1; i >= 0; i-- {
        if children[i].IsIdentifier() {
            return strings.TrimSpace(children[i].GetText())
        }
    }
    return ""
}

func (p *PropertyAccessExpression) GetExpression() *Node {
    // 第一个子节点是被访问的对象
    children := p.Node.GetChildren()
    if len(children) >= 2 {
        return children[0]
    }
    return nil
}
```

#### FunctionDeclaration
```go
type FunctionDeclaration struct {
    *Node
}

func (f *FunctionDeclaration) GetNameNode() *Node {
    // 查找第一个标识符子节点作为函数名
    return f.Node.getFirstChildByKind(KindIdentifier)
}

func (f *FunctionDeclaration) GetName() string {
    nameNode := f.GetNameNode()
    if nameNode != nil {
        return strings.TrimSpace(nameNode.GetText())
    }
    return ""
}

func (f *FunctionDeclaration) IsAnonymous() bool {
    return f.GetName() == ""
}
```

### 3. 透传API集成

```go
// 每个特定类型都提供透传API访问
func (v *VariableDeclaration) GetParserData() (parser.VariableDeclaration, bool) {
    return GetParserData[parser.VariableDeclaration](*v.Node)
}

// 通用泛型函数，提供类型安全的底层访问
func GetParserData[T any](node Node) (T, bool) {
    var zero T
    data, ok := node.GetParserData()
    if !ok {
        return zero, false
    }
    if typed, ok := data.(T); ok {
        return typed, true
    }
    return zero, false
}
```

**设计优势**:
- 类型安全：泛型确保编译时类型正确
- 降级策略：当透传数据不可用时，提供基础实现
- 性能优化：直接访问已解析的数据，避免重复解析

## 使用范式

### 类型安全的使用模式
```go
// 推荐使用方式：类型检查 + 安全转换
sf.ForEachDescendant(func(node tsmorphgo.Node) {
    if varDecl, ok := node.AsVariableDeclaration(); ok {
        // 此时 varDecl 是 *VariableDeclaration 类型，类型安全
        nameNode := varDecl.GetNameNode()
        name := varDecl.GetName()

        if varDecl.HasInitializer() {
            initializer := varDecl.GetInitializer()
            // 处理初始值...
        }
    }
})
```

### 错误处理策略
```go
// 安全的访问模式
func processVariable(node tsmorphgo.Node) {
    varDecl, ok := node.AsVariableDeclaration()
    if !ok {
        return // 不是变量声明，直接返回
    }

    // 此时可以安全使用 VariableDeclaration 的所有方法
    name := varDecl.GetName()
    if name == "" {
        // 处理异常情况...
    }
}
```

## 实现优势

### 1. 类型安全
- **编译时检查**: 只有匹配的类型才能成功转换
- **零拷贝**: 使用结构体嵌入，避免数据复制
- **接口统一**: 所有特定类型都实现 NodeWrapper 接口

### 2. 性能优化
- **透传访问**: 直接使用已解析的数据，避免重复解析
- **缓存友好**: 可以在特定类型中实现复杂的缓存逻辑
- **内存效率**: 不创建额外的数据结构

### 3. 可扩展性
- **统一模式**: 新的节点类型可以按照相同模式实现
- **渐进增强**: 可以先实现基础功能，后续添加高级特性
- **向后兼容**: 保持原有 Node API 的完整性

## 与其他方案的对比

### 与原始 ts-morph 对比
| 特性 | ts-morph (TypeScript) | tsmorphgo (Go) |
|------|---------------------|----------------|
| 类型检查 | 运行时 duck typing | 编译时静态类型 |
| 方法调用 | obj.getName() | varDecl.GetName() |
| 类型转换 | 类型断言 | 安全转换方法 |
| 性能 | 动态分发 | 静态分发 + 透传 |

### 与传统方法对比
| 方式 | 传统方式 | 专有API方式 |
|------|---------|------------|
| 类型安全 | ❌ 运行时错误 | ✅ 编译时检查 |
| API数量 | 🔧 大量重复方法 | 📈 专注核心功能 |
| 使用复杂度 | 🔍 需要类型判断 | ✅ 直接使用 |
| 可维护性 | ⚠️ 难以重构 | ✅ 清晰分离 |

## 最佳实践

### 1. 使用优先级
1. **优先使用类型安全转换**: `if typed, ok := node.AsXXX(); ok`
2. **检查节点有效性**: 使用 `IsValid()` 方法
3. **处理边界情况**: 检查 nil 和空值

### 2. 错误处理
```go
// 推荐的错误处理模式
func safeGetName(node tsmorphgo.Node) (string, error) {
    if !node.IsValid() {
        return "", fmt.Errorf("invalid node")
    }

    varDecl, ok := node.AsVariableDeclaration()
    if !ok {
        return "", fmt.Errorf("not a variable declaration")
    }

    name := varDecl.GetName()
    if name == "" {
        return "", fmt.Errorf("variable has no name")
    }

    return name, nil
}
```

### 3. 性能优化
```go
// 避免重复转换
varDecl, ok := node.AsVariableDeclaration()
if ok {
    // 在同一个作用域内重复使用 varDecl，避免重复转换
    name := varDecl.GetName()
    initializer := varDecl.GetInitializer()
    // ...
}
```

## 测试策略

### 1. 功能测试
- **场景覆盖**: 基于 ts-morph.md 的场景进行测试
- **边界测试**: 测试 nil、空值、异常情况
- **集成测试**: 测试与现有 API 的兼容性

### 2. 性能测试
- **内存使用**: 确保不增加内存占用
- **执行效率**: 对比传统方法的时间复杂度
- **并发安全**: 测试多线程环境下的安全性

### 3. 类型安全测试
```go
func TestTypeSafety(t *testing.T) {
    // 确保类型转换的安全性
    node := getNode()

    // 正确的转换应该成功
    if varDecl, ok := node.AsVariableDeclaration(); ok {
        assert.NotNil(t, varDecl)
    }

    // 错误的转换应该安全失败
    if callExpr, ok := node.AsCallExpression(); ok {
        // 在 node 不是 CallExpression 时，ok 应该为 false
        t.Errorf("unexpected successful conversion")
    }
}
```

## 未来扩展

### 1. 新节点类型支持
- InterfaceDeclaration
- ClassDeclaration
- EnumDeclaration
- TypeAliasDeclaration

### 2. 高级功能
- 批量操作 API
- 节点修改 API
- 代码生成辅助

### 3. 工具集成
- IDE 插件支持
- 代码重构工具
- 静态分析集成

## 总结

tsmorphgo 的特定节点类型专有 API 系统成功实现了以下目标：

1. **提供了类型安全的 TypeScript AST 操作接口**
2. **严格基于文档需求，避免过度设计**
3. **保持了高性能和内存效率**
4. **提供了直观易用的使用范式**
5. **实现了与底层解析器的无缝集成**

这个实现为 Go 语言中的 TypeScript 代码分析提供了强大而实用的工具，特别适合构建代码分析、重构和生成工具。