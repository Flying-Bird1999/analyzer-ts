# TSMorphGo API 需求总结

基于 `ts-morph.md` 文档分析，以下是 TSMorphGo 需要支持的完整 API 列表：

## 1. 项目和文件管理

### Project 接口
```go
type ProjectOptions struct {
    TsConfigFilePath             string  // tsconfig.json 路径
    UseInMemoryFileSystem        bool    // 使用内存文件系统（测试用）
    SkipAddingFilesFromTsConfig  bool    // 跳过自动加载 tsconfig 文件
}

type Project interface {
    constructor(options *ProjectOptions)
    GetSourceFiles() []SourceFile
    CreateSourceFile(fileName, content string) SourceFile
}
```

### SourceFile 接口
```go
type SourceFile interface {
    GetFilePath() string
    ForEachDescendant(callback func(Node) bool) // 返回 bool 用于控制遍历
}
```

## 2. 节点基础接口

### Node 核心接口
```go
type Node interface {
    // 导航相关
    GetParent() Node
    GetAncestors() []Node
    GetFirstAncestorByKind(kind SyntaxKind) Node
    GetFirstChild(predicate func(Node) bool) Node

    // 信息获取
    GetSymbol() Symbol
    GetText() string
    GetSourceFile() SourceFile
    GetStartLineNumber() int    // 1-based
    GetStart() int              // 0-based
    GetStartLinePos() int       // 0-based
    GetKind() SyntaxKind
    GetKindName() string
}
```

### Symbol 接口
```go
type Symbol interface {
    GetName() string
}
```

## 3. 节点类型判断

### 类型守卫函数
```go
// 命名空间形式组织类型判断函数
var Node = struct {
    IsIdentifier func(Node) bool
    IsCallExpression func(Node) bool
    IsPropertyAccessExpression func(Node) bool
    IsPropertyAssignment func(Node) bool
    IsVariableDeclaration func(Node) bool
    IsFunctionDeclaration func(Node) bool
    IsImportSpecifier func(Node) bool
    IsObjectLiteralExpression func(Node) bool
    IsBinaryExpression func(Node) bool
    IsInterfaceDeclaration func(Node) bool
    IsTypeAliasDeclaration func(Node) bool
}{}
```

## 4. 特定节点类型接口

### IdentifierNode
```go
type IdentifierNode interface {
    Node
    FindReferencesAsNodes() []Node
}
```

### CallExpressionNode
```go
type CallExpressionNode interface {
    Node
    GetExpression() Node
}
```

### PropertyAccessExpressionNode
```go
type PropertyAccessExpressionNode interface {
    Node
    GetName() string
    GetExpression() Node
}
```

### VariableDeclarationNode
```go
type VariableDeclarationNode interface {
    Node
    GetNameNode() Node
    GetName() string
}
```

### FunctionDeclarationNode
```go
type FunctionDeclarationNode interface {
    Node
    GetNameNode() Node
}
```

### ImportSpecifierNode
```go
type ImportSpecifierNode interface {
    Node
    GetAliasNode() Node
}
```

### BinaryExpressionNode
```go
type BinaryExpressionNode interface {
    Node
    GetOperatorToken() OperatorToken
    GetLeft() Node
    GetRight() Node
}
```

### OperatorToken
```go
type OperatorToken interface {
    GetKind() SyntaxKind
}
```

## 5. 语法类型枚举

```go
type SyntaxKind int

const (
    // 字面量
    ObjectLiteralExpression SyntaxKind = iota
    ArrayLiteralExpression

    // 声明
    VariableDeclaration
    FunctionDeclaration
    InterfaceDeclaration
    TypeAliasDeclaration

    // 表达式
    CallExpression
    PropertyAccessExpression
    BinaryExpression

    // 其他
    Identifier
    PropertyAssignment
    ImportSpecifier
    EqualsToken
)
```

## 6. 使用场景和优先级

### 高优先级（核心功能）
1. **项目初始化**: `Project` 构造函数和配置
2. **基础节点操作**: `Node` 接口的导航和信息获取方法
3. **类型判断**: 所有 `Node.IsXxx()` 类型守卫函数
4. **源文件操作**: `SourceFile` 的文件遍历功能

### 中优先级（常用功能）
1. **引用查找**: `IdentifierNode.FindReferencesAsNodes()`
2. **特定节点API**: 各个节点类型的专有方法
3. **符号系统**: `Symbol` 接口和相关方法

### 低优先级（高级功能）
1. **内存文件系统**: 测试相关的配置选项
2. **复杂表达式**: 二元表达式、操作符等高级特性

## 7. 关键设计要求

1. **类型安全**: 使用 Go 的接口和类型断言确保类型安全
2. **性能优化**: 节点遍历和查找需要考虑大型项目的性能
3. **错误处理**: API 需要处理各种边界情况（如 nil 返回值）
4. **可扩展性**: 设计应该便于添加新的节点类型和功能
5. **内存管理**: 合理管理内存，避免大量对象创建

## 8. 与 TypeScript 引擎的集成

- 底层需要与 `github.com/Zzzen/typescript-go` 集成
- 需要将 Go 的 API 调用转换为 TypeScript 编译器调用
- 缓存机制以提高性能
- 错误处理和日志记录

这个 API 设计将作为 TSMorphGo 实现的蓝图，确保覆盖所有 TypeScript AST 分析的核心需求。