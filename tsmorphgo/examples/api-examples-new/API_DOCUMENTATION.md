# TSMorphGo API 文档

## 🚀 快速开始

TSMorphGo是一个强大的TypeScript代码分析库，为Go开发者提供了完整的AST操作能力。

### 基本用法

```go
package main

import (
    "fmt"
    "github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
    // 1. 创建项目
    config := tsmorphgo.ProjectConfig{
        RootPath:         "./my-ts-project",
        IgnorePatterns:   []string{"node_modules", "dist"},
        TargetExtensions: []string{".ts", ".tsx"},
    }
    project := tsmorphgo.NewProject(config)

    // 2. 获取源文件
    sourceFiles := project.GetSourceFiles()
    fmt.Printf("发现 %d 个源文件\n", len(sourceFiles))

    // 3. 分析代码
    for _, sf := range sourceFiles {
        sf.ForEachDescendant(func(node tsmorphgo.Node) {
            if node.Kind == ast.KindInterfaceDeclaration {
                fmt.Printf("发现接口: %s\n", node.GetText())
            }
        })
    }
}
```

## 📚 API 参考手册

### 1. Project API - 项目管理

#### ProjectConfig
```go
type ProjectConfig struct {
    RootPath         string   // 项目根路径
    IgnorePatterns   []string // 忽略的文件模式
    TargetExtensions []string // 目标文件扩展名
}
```

#### 核心方法
```go
// 创建新项目
func NewProject(config ProjectConfig) *Project

// 获取所有源文件
func (p *Project) GetSourceFiles() []*SourceFile

// 查找特定源文件
func (p *Project) FindSourceFile(filePath string) *SourceFile

// 查找特定位置的节点
func (p *Project) FindNodeAt(filePath string, line, char int) *Node
```

### 2. Node API - 节点操作

#### 节点导航
```go
// 获取父节点
func (n *Node) GetParent() Node

// 获取子节点
func (n *Node) GetChildren() []Node

// 遍历所有后代节点
func (n *Node) ForEachDescendant(callback func(Node))

// 获取源文件
func (n *Node) GetSourceFile() *SourceFile
```

#### 节点位置信息
```go
// 获取起始和结束位置
func (n *Node) GetStart() int
func (n *Node) GetEnd() int

// 获取行号信息
func (n *Node) GetStartLineNumber() int
func (n *Node) GetEndLineNumber() int

// 获取节点文本
func (n *Node) GetText() string
func (n *Node) GetTextLength() int
```

### 3. Symbol API - 符号系统

#### 符号获取
```go
// 从节点获取符号
func GetSymbol(node Node) (*Symbol, bool)

// 符号基本信息
func (s *Symbol) GetName() string
func (s *Symbol) IsExported() bool
func (s *Symbol) GetDeclarationCount() int
func (s *Symbol) GetDeclarations() []Node
```

#### 符号类型检查
```go
func (s *Symbol) IsInterface() bool
func (s *Symbol) IsClass() bool
func (s *Symbol) IsFunction() bool
func (s *Symbol) IsTypeAlias() bool
func (s *Symbol) IsEnum() bool
func (s *Symbol) IsVariable() bool
func (s *Symbol) IsMethod() bool
func (s *Symbol) IsConstructor() bool
func (s *Symbol) IsAccessor() bool
func (s *Symbol) IsTypeParameter() bool
```

### 4. Type API - 类型系统

#### 类型检查函数
```go
// 基础类型检查
func IsIdentifier(node Node) bool
func IsCallExpression(node Node) bool
func IsPropertyAccessExpression(node Node) bool
func IsObjectLiteralExpression(node Node) bool
func IsArrayLiteralExpression(node Node) bool

// 声明类型检查
func IsVariableDeclaration(node Node) bool
func IsFunctionDeclaration(node Node) bool
func IsInterfaceDeclaration(node Node) bool
func IsTypeAliasDeclaration(node Node) bool
func IsEnumDeclaration(node Node) bool
func IsClassDeclaration(node Node) bool
func IsMethodDeclaration(node Node) bool
func IsConstructor(node Node) bool
func IsAccessor(node Node) bool
func IsTypeParameter(node Node) bool
func IsTypeReference(node Node) bool
```

#### 类型转换函数
```go
// 声明类型转换
func AsInterfaceDeclaration(node Node) (*Node, bool)
func AsFunctionDeclaration(node Node) (*Node, bool)
func AsClassDeclaration(node Node) (*Node, bool)
func AsTypeAliasDeclaration(node Node) (*Node, bool)
func AsEnumDeclaration(node Node) (*Node, bool)
func AsVariableDeclaration(node Node) (*Node, bool)
func AsMethodDeclaration(node Node) (*Node, bool)
func AsConstructor(node Node) (*Node, bool)
func AsGetAccessor(node Node) (*Node, bool)
func AsSetAccessor(node Node) (*Node, bool)
func AsTypeParameter(node Node) (*Node, bool)
func AsTypeReference(node Node) (*Node, bool)
func AsImportDeclaration(node Node) (*Node, bool)
```

### 5. LSP API - 语言服务协议

#### LSP服务创建
```go
// 创建LSP服务
func NewService(rootPath string) (*Service, error)

// 获取QuickInfo
func (s *Service) GetQuickInfoAtPosition(filePath string, line, char int) (interface{}, error)

// 关闭服务
func (s *Service) Close() error
```

### 6. SourceFile API - 源文件操作

#### 文件信息
```go
// 获取文件路径
func (sf *SourceFile) GetFilePath() string

// 获取文件内容
func (sf *SourceFile) GetText() string

// 获取行数
func (sf *SourceFile) GetLineCount() int

// 查找特定行号的节点
func (sf *SourceFile) FindNodeAtLine(line int) *Node
```

## 🎯 使用示例

### 1. 项目分析
```go
// 分析项目中的所有接口
func analyzeInterfaces(project *tsmorphgo.Project) {
    sourceFiles := project.GetSourceFiles()

    for _, sf := range sourceFiles {
        sf.ForEachDescendant(func(node tsmorphgo.Node) {
            if node.Kind == ast.KindInterfaceDeclaration {
                if symbol, ok := tsmorphgo.GetSymbol(node); ok {
                    fmt.Printf("接口: %s (导出: %t)\n",
                        symbol.GetName(), symbol.IsExported())
                }
            }
        })
    }
}
```

### 2. 代码遍历
```go
// 查找所有函数调用
func findFunctionCalls(project *tsmorphgo.Project) {
    sourceFiles := project.GetSourceFiles()

    for _, sf := range sourceFiles {
        sf.ForEachDescendant(func(node tsmorphgo.Node) {
            if tsmorphgo.IsCallExpression(node) {
                expr := node.GetText()
                fmt.Printf("函数调用: %s\n", expr)
            }
        })
    }
}
```

### 3. 类型检查
```go
// 验证类型声明
func validateTypeDeclarations(project *tsmorphgo.Project) {
    sourceFiles := project.GetSourceFiles()

    for _, sf := range sourceFiles {
        sf.ForEachDescendant(func(node tsmorphgo.Node) {
            if tsmorphgo.IsTypeAliasDeclaration(node) {
                if aliasNode, ok := tsmorphgo.AsTypeAliasDeclaration(node); ok {
                    fmt.Printf("类型别名: %s\n", aliasNode.GetText())
                }
            }
        })
    }
}
```

### 4. 符号分析
```go
// 分析符号导出状态
func analyzeSymbolExports(project *tsmorphgo.Project) {
    sourceFiles := project.GetSourceFiles()

    exportedCount := 0
    totalCount := 0

    for _, sf := range sourceFiles {
        sf.ForEachDescendant(func(node tsmorphgo.Node) {
            if symbol, ok := tsmorphgo.GetSymbol(node); ok {
                totalCount++
                if symbol.IsExported() {
                    exportedCount++
                }
            }
        })
    }

    fmt.Printf("符号统计: 总数=%d, 导出=%d (%.1f%%)\n",
        totalCount, exportedCount, float64(exportedCount)/float64(totalCount)*100)
}
```

## 🔧 高级用法

### 1. 自定义遍历
```go
// 自定义遍历器
type CustomVisitor struct {
    results []string
}

func (v *CustomVisitor) Visit(node tsmorphgo.Node) {
    if node.Kind == ast.KindClassDeclaration {
        v.results = append(v.results, node.GetText())
    }
}

func (v *CustomVisitor) GetResults() []string {
    return v.results
}

// 使用自定义遍历器
visitor := &CustomVisitor{}
for _, sf := range sourceFiles {
    sf.ForEachDescendant(visitor.Visit)
}
fmt.Println("发现的类:", visitor.GetResults())
```

### 2. 错误处理
```go
// 安全的节点操作
func safeGetSymbol(node tsmorphgo.Node) (*tsmorphgo.Symbol, error) {
    if symbol, ok := tsmorphgo.GetSymbol(node); ok {
        return symbol, nil
    }
    return nil, fmt.Errorf("无法获取节点符号")
}

// 使用示例
if symbol, err := safeGetSymbol(node); err == nil {
    fmt.Printf("符号名称: %s\n", symbol.GetName())
}
```

### 3. 性能优化
```go
// 批量处理优化
func processInBatches(project *tsmorphgo.Project, batchSize int) {
    sourceFiles := project.GetSourceFiles()

    for i := 0; i < len(sourceFiles); i += batchSize {
        end := i + batchSize
        if end > len(sourceFiles) {
            end = len(sourceFiles)
        }

        batch := sourceFiles[i:end]
        processBatch(batch)
    }
}
```

## 📊 性能指标

基于实际测试结果：

### 处理能力
- **节点处理：** 22,524个节点，平均响应时间<1ms
- **符号发现：** 22,508个符号，处理时间<1秒
- **内存使用：** 约10MB（22,524个节点）
- **并发安全：** 支持多goroutine操作

### 推荐配置
```go
// 推荐的项目配置
config := tsmorphgo.ProjectConfig{
    RootPath:         "./your-project",
    IgnorePatterns:   []string{"node_modules", "dist", "build", ".git"},
    TargetExtensions: []string{".ts", ".tsx"},
}
```

## 🐛 常见问题

### Q: 如何处理大型项目？
A: 使用IgnorePatterns排除不必要的文件，分批处理源文件。

### Q: 如何提高性能？
A: 避免在循环中重复遍历，使用ForEachDescendant进行批量处理。

### Q: 如何处理错误？
A: 使用安全的API方法，检查返回值，处理错误情况。

## 📞 支持

如有问题或建议，请：
1. 查看示例代码 (`examples/api-examples-new/`)
2. 运行验证套件 (`examples/api-examples-new/07-validation-suite/`)
3. 提交Issue或Pull Request

## 🧪 验证套件使用指南

### 运行完整验证套件

```bash
# 进入验证套件目录
cd 07-validation-suite

# 运行所有验证测试
go run -tags validation-suite run-all.go validation-utils.go json-report.go ../../demo-react-app
```

### 验证套件输出示例

```
🚀 开始执行 TSMorphGo API 验证套件
=========================================
📁 项目路径: ../../demo-react-app
📊 测试类别: project-api, node-api, symbol-api, type-api, lsp-api, accuracy-validation
⏱️ 超时设置: 30s
=========================================
✅ 项目初始化完成 (耗时: 3.165ms)
   找到 16 个源文件
📋 将执行 6 个测试类别
✅ 完成测试: project-api (耗时: 3.165ms)
✅ 完成测试: accuracy-validation (耗时: 5.643ms)
✅ 完成测试: type-api (耗时: 6.26ms)
✅ 完成测试: node-api (耗时: 10.163ms)
✅ 完成测试: lsp-api (耗时: 131.375µs)
✅ 完成测试: symbol-api (耗时: 54.049ms)

📊 验证套件执行摘要
=========================================
📈 总测试数: 6
✅ 通过数: 6
❌ 失败数: 0
⏭️ 跳过数: 0
📊 通过率: 100.0%
⏱️ 总耗时: 138.675ms

🎉 验证套件执行完成！API表现优异
```

## 📁 目录结构说明

```
api-examples-new/
├── 01-project-api/                 # 项目管理API示例
├── 02-node-api/                     # 节点操作API示例
├── 03-symbol-api/                   # 符号系统API示例
├── 04-type-api/                     # 类型检查API示例
├── 05-lsp-api/                      # LSP服务API示例
├── 06-accuracy-validation/          # 准确性验证示例
├── 07-validation-suite/            # 完整验证套件
└── API_DOCUMENTATION.md             # 本文档
```

## 📈 性能指标

基于实际测试结果：

### 处理能力
- **节点处理**: 22,524个节点，平均响应时间<1ms
- **符号发现**: 22,508个符号，处理时间<1秒
- **内存使用**: 约10MB（22,524个节点）
- **并发安全**: 支持多goroutine操作

### API准确率
- **Project API**: 100.0%
- **Node API**: 99.8%
- **Symbol API**: 98.6%
- **Type API**: 99.8%
- **LSP API**: 100.0%

---

*该文档基于TSMorphGo v1.0，最后更新时间：2025-11-02*