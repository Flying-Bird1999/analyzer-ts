# TSMorphGo API 参考文档

## 目录

- [核心类型](#核心类型)
- [Node API](#node-api)
- [Project API](#project-api)
- [SourceFile API](#sourcefile-api)
- [Reference API](#reference-api)
- [类型转换API](#类型转换api)
- [语法类型](#语法类型)
- [示例代码](#示例代码)

## 核心类型

### Node

AST节点的统一包装器，提供一致的访问接口。

```go
type Node struct {
    *ast.Node                    // 底层AST节点
    sourceFile *SourceFile       // 所属源文件
    declarationAccessor DeclarationAccessor // 声明访问器
}
```

### Project

TypeScript项目管理器，处理文件集合和配置。

```go
type Project struct {
    parserResult   *projectParser.ProjectParserResult
    sourceFiles    map[string]*SourceFile
    lspService     *lsp.Service
    symbolManager  *SymbolManager
    referenceCache *ReferenceCache
}
```

### SourceFile

源文件的抽象表示。

```go
type SourceFile struct {
    filePath     string
    fileResult   *projectParser.JsFileParserResult
    astNode      *ast.Node
    project      *Project
    nodeResultMap map[*ast.Node]interface{}
}
```

## Node API

### 类型检查方法

#### 精确类型检查

```go
// 检查具体的语法类型
node.IsKind(KindFunctionDeclaration)      // 函数声明
node.IsKind(KindInterfaceDeclaration)     // 接口声明
node.IsKind(KindClassDeclaration)         // 类声明
node.IsKind(KindVariableDeclaration)      // 变量声明
node.IsKind(KindImportDeclaration)        // 导入声明
node.IsKind(KindExportDeclaration)        // 导出声明
node.IsKind(KindCallExpression)           // 函数调用
node.IsKind(KindStringLiteral)            // 字符串字面量
node.IsKind(KindNumericLiteral)           // 数字字面量
```

#### 便捷类型检查

```go
// 常用类型的便捷检查方法
node.IsFunctionDeclaration()      // 函数声明
node.IsInterfaceDeclaration()     // 接口声明
node.IsClassDeclaration()         // 类声明
node.IsVariableDeclaration()      // 变量声明
node.IsCallExpr()                 // 函数调用
node.IsImportDeclaration()        // 导入声明
node.IsExportDeclaration()        // 导出声明
node.IsIdentifierNode()           // 标识符节点
node.IsPropertyAccessExpression() // 属性访问表达式
```

#### 类别检查

```go
// 批量检查节点类别
node.IsDeclaration()    // 所有声明类型
node.IsExpression()     // 所有表达式类型
node.IsType()          // 所有类型相关
node.IsModule()        // 所有模块相关
node.IsLiteral()       // 所有字面量类型
```

#### 多类型检查

```go
// 一次检查多种类型
kinds := []SyntaxKind{
    KindFunctionDeclaration,
    KindInterfaceDeclaration,
    KindClassDeclaration,
}

if node.IsAnyKind(kinds...) {
    // 处理声明类型节点
}
```

### 信息获取方法

```go
// 获取节点基本信息
node.GetNodeName()           // 获取节点名称 (string, bool)
node.GetText()              // 获取节点文本内容
node.GetKind()              // 获取语法类型
node.IsValid()              // 检查节点是否有效

// 获取位置信息
node.GetStartLineNumber()   // 获取起始行号 (1-based)
node.GetStartColumnNumber() // 获取起始列号 (1-based)
node.GetStartLineCharacter() // 获取起始列号 (0-based)
node.GetStartLinePos()      // 获取行起始位置
node.GetStart()             // 获取起始位置
node.GetEnd()               // 获取结束位置

// 获取关联信息
node.GetSourceFile()        // 获取所属源文件
```

### 导航方法

```go
// 节点导航
node.GetParent()            // 获取直接父节点
node.GetAncestors()         // 获取所有祖先节点
node.GetFirstAncestorByKind(kind SyntaxKind) (*Node, bool) // 查找特定类型祖先

// 节点遍历
node.ForEachDescendant(callback func(Node)) // 遍历所有子孙节点
```

### 类型转换方法

```go
// 统一类型转换
if result, ok := node.AsDeclaration(); ok {
    // 处理声明类型结果
}

// 具体类型转换函数
if result, ok := AsVariableDeclaration(node); ok {
    // 处理变量声明结果
}

if result, ok := AsFunctionDeclaration(node); ok {
    // 处理函数声明结果
}
```

## Project API

### 创建项目

```go
// 从文件系统创建项目
project := NewProject(ProjectConfig{
    RootPath:         "./src",                    // 项目根路径
    TargetExtensions: []string{".ts", ".tsx"},   // 目标文件扩展名
    IgnorePatterns:   []string{"node_modules"},  // 忽略模式
    UseTsConfig:      true,                      // 使用tsconfig.json
    TsConfigPath:     "./tsconfig.json",         // tsconfig.json路径
})
defer project.Close() // 确保资源释放

// 从内存源码创建项目
project := NewProjectFromSources(map[string]string{
    "/src/types.ts": "export interface User { id: number; }",
    "/src/utils.ts": "export function helper() { return true; }",
})
defer project.Close()
```

### 文件操作

```go
// 获取所有源文件
sourceFiles := project.GetSourceFiles()

// 获取特定文件
sourceFile := project.GetSourceFile("/src/main.ts")

// 动态创建文件
newFile := project.CreateSourceFile("/src/generated.ts", content)

// 获取项目LSP服务
lspService, err := project.GetLSPService()
```

### 配置选项

```go
type ProjectConfig struct {
    RootPath         string   // 项目根路径
    IgnorePatterns   []string // 忽略的文件/目录模式
    IsMonorepo       bool     // 是否为单仓库项目
    TargetExtensions []string // 目标文件扩展名
    TsConfigPath     string   // TypeScript配置文件路径
    UseTsConfig      bool     // 是否使用tsconfig.json
    CompilerOptions  map[string]interface{} // 编译选项
    IncludePatterns  []string // 包含的文件模式
    ExcludePatterns  []string // 排除的文件模式
}
```

## SourceFile API

### 文件信息

```go
// 获取文件基本信息
filePath := sourceFile.GetFilePath()      // 文件路径
fileResult := sourceFile.GetFileResult()  // 解析结果
astNode := sourceFile.GetAstNode()        // AST根节点
project := sourceFile.GetProject()        // 所属项目

// 遍历文件中的所有节点
sourceFile.ForEachDescendant(func(node Node) {
    // 处理每个节点
})
```

### 解析结果访问

```go
// 访问解析结果
if fileResult := sourceFile.GetFileResult(); fileResult != nil {
    // 导入声明
    for _, importDecl := range fileResult.ImportDeclarations {
        fmt.Printf("导入: %s\n", importDecl.ModuleSpecifier.Text)
    }

    // 导出声明
    for _, exportDecl := range fileResult.ExportDeclarations {
        fmt.Printf("导出: %s\n", exportDecl.Text)
    }

    // 变量声明
    for _, varDecl := range fileResult.VariableDeclarations {
        fmt.Printf("变量: %s\n", varDecl.Name)
    }
}
```

## Reference API

### 基础引用查找

```go
// 查找符号的所有引用
refs, err := FindReferences(node)
if err != nil {
    log.Printf("查找引用失败: %v", err)
    return
}

for _, ref := range refs {
    fmt.Printf("引用: %s (文件: %s, 行: %d)\n",
        ref.GetText(),
        ref.GetSourceFile().GetFilePath(),
        ref.GetStartLineNumber())
}
```

### 带缓存的引用查找

```go
// 使用缓存提升性能
refs, fromCache, err := FindReferencesWithCache(node)
if err == nil {
    if fromCache {
        fmt.Println("结果来自缓存")
    }
    fmt.Printf("找到 %d 个引用\n", len(refs))
}
```

### 带重试的引用查找

```go
// 配置重试参数
config := &RetryConfig{
    MaxRetries: 3,
    Delay:      time.Second,
}

refs, fromCache, err := FindReferencesWithCacheAndRetry(node, config)
```

### 引用缓存配置

```go
// 获取项目缓存
cache := project.getReferenceCache()

// 清空缓存
cache.Clear()

// 缓存统计
stats := cache.GetStats()
fmt.Printf("缓存大小: %d, 命中率: %.2f%%\n",
    stats.Size, stats.HitRate*100)
```

## 类型转换API

### 声明类型转换

```go
// 转换为导入声明
if result, ok := AsImportDeclaration(node); ok {
    fmt.Printf("导入模块: %s\n", result.ModuleSpecifier.Text)
}

// 转换为变量声明
if result, ok := AsVariableDeclaration(node); ok {
    fmt.Printf("变量名: %s, 类型: %s\n", result.Name, result.Type)
}

// 转换为函数声明
if result, ok := AsFunctionDeclaration(node); ok {
    fmt.Printf("函数名: %s\n", result.Name)
}

// 转换为接口声明
if result, ok := AsInterfaceDeclaration(node); ok {
    fmt.Printf("接口名: %s\n", result.Name)
}

// 转换为类声明
if result, ok := AsClassDeclaration(node); ok {
    fmt.Printf("类名: %s\n", result.Name)
}
```

### 表达式类型转换

```go
// 转换为调用表达式
if result, ok := AsCallExpression(node); ok {
    fmt.Printf("调用表达式: %s\n", result.Expression)
}

// 转换为属性访问表达式
if result, ok := AsPropertyAccessExpression(node); ok {
    fmt.Printf("属性访问: %s.%s\n", result.Expression, result.Name)
}

// 转换为标识符
if result, ok := AsIdentifier(node); ok {
    fmt.Printf("标识符: %s\n", result.Text)
}
```

## 语法类型

### 语句类型

```go
const (
    KindVariableStatement        SyntaxKind = ast.KindVariableStatement
    KindFunctionDeclaration      SyntaxKind = ast.KindFunctionDeclaration
    KindInterfaceDeclaration     SyntaxKind = ast.KindInterfaceDeclaration
    KindTypeAliasDeclaration     SyntaxKind = ast.KindTypeAliasDeclaration
    KindClassDeclaration         SyntaxKind = ast.KindClassDeclaration
    KindEnumDeclaration          SyntaxKind = ast.KindEnumDeclaration
    KindImportDeclaration        SyntaxKind = ast.KindImportDeclaration
    KindExportDeclaration        SyntaxKind = ast.KindExportDeclaration
    KindReturnStatement          SyntaxKind = ast.KindReturnStatement
    KindIfStatement              SyntaxKind = ast.KindIfStatement
    KindForStatement             SyntaxKind = ast.KindForStatement
    KindWhileStatement           SyntaxKind = ast.KindWhileStatement
)
```

### 表达式类型

```go
const (
    KindCallExpression           SyntaxKind = ast.KindCallExpression
    KindPropertyAccessExpression SyntaxKind = ast.KindPropertyAccessExpression
    KindBinaryExpression         SyntaxKind = ast.KindBinaryExpression
    KindUnaryExpression          SyntaxKind = ast.KindUnaryExpression
    KindConditionalExpression    SyntaxKind = ast.KindConditionalExpression
    KindArrayLiteralExpression   SyntaxKind = ast.KindArrayLiteralExpression
    KindObjectLiteralExpression  SyntaxKind = ast.KindObjectLiteralExpression
)
```

### 类型相关

```go
const (
    KindTypeReference            SyntaxKind = ast.KindTypeReference
    KindArrayType                SyntaxKind = ast.KindArrayType
    KindUnionType                SyntaxKind = ast.KindUnionType
    KindIntersectionType         SyntaxKind = ast.KindIntersectionType
    KindTypeParameter            SyntaxKind = ast.KindTypeParameter
)
```

### 字面量类型

```go
const (
    KindStringLiteral            SyntaxKind = ast.KindStringLiteral
    KindNumericLiteral           SyntaxKind = ast.KindNumericLiteral
    KindBooleanLiteral           SyntaxKind = ast.KindBooleanLiteral
    KindNullLiteral              SyntaxKind = ast.KindNullLiteral
    KindUndefinedLiteral         SyntaxKind = ast.KindUndefinedLiteral
)
```

## 示例代码

### 完整项目分析示例

```go
package main

import (
    "fmt"
    "log"
    "github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
    // 创建项目
    project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
        RootPath:         "./src",
        TargetExtensions: []string{".ts", ".tsx"},
        IgnorePatterns:   []string{"node_modules", "dist"},
        UseTsConfig:      true,
    })
    defer project.Close()

    // 分析项目
    analyzeProject(project)
}

func analyzeProject(project *tsmorphgo.Project) {
    fmt.Println("=== TSMorphGo 项目分析 ===")

    // 1. 文件统计
    files := project.GetSourceFiles()
    fmt.Printf("📁 项目文件: %d 个\n", len(files))

    // 2. 类型统计
    typeStats := make(map[string]int)
    functionCount := 0
    interfaceCount := 0

    for _, file := range files {
        file.ForEachDescendant(func(node tsmorphgo.Node) {
            switch {
            case node.IsFunctionDeclaration():
                functionCount++
                if name, ok := node.GetNodeName(); ok {
                    fmt.Printf("  函数: %s (行 %d)\n", name, node.GetStartLineNumber())
                }
            case node.IsInterfaceDeclaration():
                interfaceCount++
                if name, ok := node.GetNodeName(); ok {
                    fmt.Printf("  接口: %s (行 %d)\n", name, node.GetStartLineNumber())
                }
            case node.IsClassDeclaration():
                typeStats["类"]++
            case node.IsVariableDeclaration():
                typeStats["变量"]++
            case node.IsImportDeclaration():
                typeStats["导入"]++
            case node.IsCallExpr():
                typeStats["调用"]++
            }
        })
    }

    // 3. 输出统计结果
    fmt.Printf("\n📊 统计结果:\n")
    fmt.Printf("  函数声明: %d\n", functionCount)
    fmt.Printf("  接口声明: %d\n", interfaceCount)
    for kind, count := range typeStats {
        fmt.Printf("  %s: %d\n", kind, count)
    }

    // 4. 符号分析
    fmt.Printf("\n🔍 符号分析:\n")
    analyzeSymbols(project)
}

func analyzeSymbols(project *tsmorphgo.Project) {
    // 查找第一个函数
    var firstFunction *tsmorphgo.Node
    for _, file := range project.GetSourceFiles() {
        file.ForEachDescendant(func(node tsmorphgo.Node) {
            if node.IsFunctionDeclaration() && firstFunction == nil {
                firstFunction = &node
            }
        })
        if firstFunction != nil {
            break
        }
    }

    if firstFunction != nil {
        if name, ok := firstFunction.GetNodeName(); ok {
            fmt.Printf("分析函数: %s\n", name)

            // 查找引用
            refs, err := tsmorphgo.FindReferences(*firstFunction)
            if err != nil {
                fmt.Printf("  查找引用失败: %v\n", err)
                return
            }

            fmt.Printf("  引用数量: %d\n", len(refs))
            for i, ref := range refs {
                if i >= 3 { // 只显示前3个
                    fmt.Printf("    ... (还有 %d 个)\n", len(refs)-3)
                    break
                }
                fmt.Printf("    %d. %s:%d\n",
                    i+1,
                    ref.GetSourceFile().GetFilePath(),
                    ref.GetStartLineNumber())
            }
        }
    }
}
```

### 内存项目示例

```go
package main

import (
    "fmt"
    "github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
    // 创建内存项目
    project := tsmorphgo.NewProjectFromSources(map[string]string{
        "/src/types.ts": `
            export interface User {
                id: number;
                name: string;
                email?: string;
            }

            export type UserRole = 'admin' | 'user' | 'guest';

            export enum UserStatus {
                Active = 'active',
                Inactive = 'inactive',
                Suspended = 'suspended'
            }
        `,
        "/src/services.ts": `
            import { User, UserRole } from './types';

            export class UserService {
                private users: User[] = [];

                addUser(user: Omit<User, 'id'>): User {
                    const newUser: User = {
                        id: Math.random(),
                        ...user
                    };
                    this.users.push(newUser);
                    return newUser;
                }

                findUser(id: number): User | undefined {
                    return this.users.find(u => u.id === id);
                }

                getAllUsers(): User[] {
                    return [...this.users];
                }
            }
        `,
        "/src/main.ts": `
            import { UserService } from './services';

            const service = new UserService();
            const user = service.addUser({
                name: 'John Doe',
                email: 'john@example.com'
            });

            console.log('用户已创建:', user);
        `,
    })
    defer project.Close()

    // 分析内存项目
    analyzeMemoryProject(project)
}

func analyzeMemoryProject(project *tsmorphgo.Project) {
    fmt.Println("=== 内存项目分析 ===")

    // 1. 获取所有文件
    files := project.GetSourceFiles()
    fmt.Printf("📁 内存项目文件: %d 个\n", len(files))

    // 2. 分析每个文件
    for _, file := range files {
        fmt.Printf("\n📄 分析文件: %s\n", file.GetFilePath())

        // 分析类型定义
        interfaces := 0
        enums := 0
        typeAliases := 0
        classes := 0
        functions := 0

        file.ForEachDescendant(func(node tsmorphgo.Node) {
            switch {
            case node.IsInterfaceDeclaration():
                interfaces++
                if name, ok := node.GetNodeName(); ok {
                    fmt.Printf("  接口: %s\n", name)
                }
            case node.IsKind(KindEnumDeclaration):
                enums++
                if name, ok := node.GetNodeName(); ok {
                    fmt.Printf("  枚举: %s\n", name)
                }
            case node.IsKind(KindTypeAliasDeclaration):
                typeAliases++
                if name, ok := node.GetNodeName(); ok {
                    fmt.Printf("  类型别名: %s\n", name)
                }
            case node.IsClassDeclaration():
                classes++
                if name, ok := node.GetNodeName(); ok {
                    fmt.Printf("  类: %s\n", name)
                }
            case node.IsFunctionDeclaration():
                functions++
                if name, ok := node.GetNodeName(); ok {
                    fmt.Printf("  函数: %s\n", name)
                }
            }
        })

        fmt.Printf("  统计: 接口=%d, 枚举=%d, 类型=%d, 类=%d, 函数=%d\n",
            interfaces, enums, typeAliases, classes, functions)
    }

    // 3. 动态创建新文件
    fmt.Println("\n➕ 创建动态文件...")
    newFile := project.CreateSourceFile("/src/generated.ts", `
        // 自动生成的文件
        export const VERSION = "1.0.0";
        export const BUILD_DATE = new Date().toISOString();

        export function getConfig() {
            return {
                version: VERSION,
                buildDate: BUILD_DATE
            };
        }
    `)

    if newFile != nil {
        fmt.Printf("✅ 文件创建成功: %s\n", newFile.GetFilePath())

        // 分析新文件
        newFile.ForEachDescendant(func(node tsmorphgo.Node) {
            if node.IsVariableDeclaration() {
                if name, ok := node.GetNodeName(); ok {
                    fmt.Printf("  新增变量: %s\n", name)
                }
            }
        })
    }
}
```

### 引用分析示例

```go
package main

import (
    "fmt"
    "time"
    "github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
    project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
        RootPath:         "./src",
        TargetExtensions: []string{".ts", ".tsx"},
        UseTsConfig:      true,
    })
    defer project.Close()

    // 引用分析
    analyzeReferences(project)
}

func analyzeReferences(project *tsmorphgo.Project) {
    fmt.Println("=== 引用分析 ===")

    // 查找所有接口定义
    var interfaces []*tsmorphgo.Node
    for _, file := range project.GetSourceFiles() {
        file.ForEachDescendant(func(node tsmorphgo.Node) {
            if node.IsInterfaceDeclaration() {
                interfaces = append(interfaces, &node)
            }
        })
    }

    // 分析每个接口的引用
    for _, iface := range interfaces {
        if name, ok := iface.GetNodeName(); ok {
            fmt.Printf("\n🎭 分析接口: %s\n", name)

            // 查找引用
            start := time.Now()
            refs, fromCache, err := tsmorphgo.FindReferencesWithCache(*iface)
            duration := time.Since(start)

            if err != nil {
                fmt.Printf("  ❌ 查找失败: %v\n", err)
                continue
            }

            fmt.Printf("  📊 引用统计:\n")
            fmt.Printf("    - 引用数量: %d\n", len(refs))
            fmt.Printf("    - 查找耗时: %v\n", duration)
            if fromCache {
                fmt.Printf("    - 结果来源: 缓存\n")
            } else {
                fmt.Printf("    - 结果来源: 实时计算\n")
            }

            // 按文件分组显示引用
            refsByFile := make(map[string][]*tsmorphgo.Node)
            for _, ref := range refs {
                filePath := ref.GetSourceFile().GetFilePath()
                refsByFile[filePath] = append(refsByFile[filePath], ref)
            }

            fmt.Printf("  📁 引用分布:\n")
            for filePath, fileRefs := range refsByFile {
                fmt.Printf("    %s (%d个):\n", filePath, len(fileRefs))
                for i, ref := range fileRefs {
                    if i >= 2 { // 每个文件最多显示2个引用
                        fmt.Printf("      ... (还有%d个)\n", len(fileRefs)-2)
                        break
                    }
                    fmt.Printf("      %d. 行%d: %s\n",
                        i+1,
                        ref.GetStartLineNumber(),
                        truncateString(ref.GetText(), 50))
                }
            }
        }
    }
}

func truncateString(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen-3] + "..."
}
```

这个API参考文档提供了TSMorphGo所有核心功能的详细说明，包括类型定义、方法签名、使用示例和最佳实践。开发者可以根据需要查找特定的API使用方法。