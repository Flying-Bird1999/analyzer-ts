# TSMorphGo 迁移指南
## 从 ts-morph 到 TSMorphGo 的完整迁移手册

> **目标读者**: 正在使用 ts-morph 并希望迁移到 TSMorphGo 的开发者
> **适用版本**: TSMorphGo v1.0.0+
> **更新日期**: 2025-11-12

---

## 📋 目录

- [快速开始](#快速开始)
- [API 对比参考](#api-对比参考)
- [项目迁移](#项目迁移)
- [语法对比](#语法对比)
- [最佳实践](#最佳实践)
- [常见问题](#常见问题)
- [完整示例](#完整示例)

---

## 🚀 快速开始

### 安装和基础设置

```bash
// ts-morph (TypeScript)
npm install ts-morph

// TSMorphGo (Go)
go get github.com/Flying-Bird1999/analyzer-ts/tsmorphgo
```

### 基础代码结构对比

#### ts-morph (TypeScript)
```typescript
import { Project, Node, SyntaxKind } from "ts-morph";

const project = new Project({
  tsConfigFilePath: "./tsconfig.json",
});

const sourceFile = project.addSourceFileAtPath("./example.ts");
const nodes = sourceFile.getDescendantsOfKind(SyntaxKind.FunctionDeclaration);
```

#### TSMorphGo (Go)
```go
package main

import (
    "github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
    . "github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
    // 创建项目配置
    config := tsmorphgo.ProjectConfig{
        RootPath:    ".",
        UseTsConfig: true,
    }

    // 初始化项目
    project := NewProject(config)
    defer project.Close()

    // 获取源文件
    sourceFile := project.GetSourceFile("./example.ts")
    if sourceFile == nil {
        return
    }

    // 查找函数声明节点
    var functionDeclarations []*Node
    sourceFile.ForEachDescendant(func(node Node) {
        if node.IsFunctionDeclaration() {
            functionDeclarations = append(functionDeclarations, &node)
        }
    })
}
```

---

## 📊 API 对比参考

### 1. 项目初始化与管理

#### ts-morph → TSMorphGo

| 功能 | ts-morph | TSMorphGo | 说明 |
|------|----------|-----------|------|
| 创建项目 | `new Project()` | `NewProject()` | ✅ 完全兼容 |
| 内存项目 | `new Project({ useInMemoryFileSystem: true })` | `NewProject(ProjectConfig{ UseInMemoryFileSystem: true })` | ✅ 完全兼容 |
| 添加源文件 | `project.addSourceFileAtPath()` | `project.CreateSourceFile()` | ✅ 完全兼容 |
| 获取源文件 | `project.getSourceFile()` | `project.GetSourceFile()` | ✅ 完全兼容 |
| 获取所有源文件 | `project.getSourceFiles()` | `project.GetSourceFiles()` | ✅ 完全兼容 |

#### 🎯 实际示例：项目创建

```typescript
// ts-morph
const project = new Project({
    tsConfigFilePath: "./tsconfig.json",
    useInMemoryFileSystem: false,
    skipAddingFilesFromTsConfig: false,
});

const sourceFile = project.addSourceFileAtPath("example.ts", `
    export function hello(name: string): string {
        return `Hello, ${name}!`;
    }
`);
```

```go
// TSMorphGo
config := ProjectConfig{
    RootPath: ".",
    UseTsConfig: true,
    UseInMemoryFileSystem: false,
    SkipAddingFilesFromTsConfig: false,
}

project := NewProject(config)
defer project.Close()

sourceFile, err := project.CreateSourceFile("example.ts", `
export function hello(name: string): string {
    return \`Hello, ${name}!\`;
}
`)
if err != nil {
    panic(err)
}
```

---

### 2. 节点导航和遍历

#### ts-morph → TSMorphGo

| 功能 | ts-morph | TSMorphGo | 状态 |
|------|----------|-----------|------|
| 获取父节点 | `node.getParent()` | `node.GetParent()` | ✅ 完全兼容 |
| 获取祖先节点 | `node.getAncestors()` | `node.GetAncestors()` | ✅ 完全兼容 |
| 按类型找祖先 | `node.getFirstAncestorByKind()` | `node.GetFirstAncestorByKind()` | ✅ 完全兼容 |
| 获取子节点 | `node.getChildren()` | `node.GetChildren()` | ✅ 完全兼容 |
| 遍历后代 | `node.forEachDescendant()` | `sourceFile.ForEachDescendant()` | ✅ 完全兼容 |

#### 🎯 实际示例：节点导航

```typescript
// ts-morph
const functionNode = someNode.getFirstAncestorByKind(SyntaxKind.FunctionDeclaration);
const parentNode = someNode.getParent();

functionNode.forEachDescendant(child => {
    if (child.isKind(SyntaxKind.Identifier)) {
        console.log(child.getText());
    }
});
```

```go
// TSMorphGo
var functionNode *Node
ancestors := someNode.GetAncestors()
for _, ancestor := range ancestors {
    if ancestor.IsFunctionDeclaration() {
        functionNode = &ancestor
        break
    }
}

parentNode := someNode.GetParent()

sourceFile := someNode.GetSourceFile()
sourceFile.ForEachDescendant(func(node Node) {
    if node.IsIdentifier() {
        fmt.Println(node.GetText())
    }
})
```

---

### 3. 节点类型判断

#### ts-morph → TSMorphGo

| 功能 | ts-morph | TSMorphGo | 状态 |
|------|----------|-----------|------|
| 是否为标识符 | `node.isIdentifier()` | `node.IsIdentifier()` | ✅ 完全兼容 |
| 是否为函数声明 | `node.isFunctionDeclaration()` | `node.IsFunctionDeclaration()` | ✅ 完全兼容 |
| 是否为变量声明 | `node.isVariableDeclaration()` | `node.IsVariableDeclaration()` | ✅ 完全兼容 |
| 是否为导入说明符 | `node.isImportSpecifier()` | `node.IsImportSpecifier()` | ✅ 完全兼容 |
| 是否为调用表达式 | `node.isCallExpression()` | `node.IsCallExpression()` | ✅ 完全兼容 |
| 是否为属性访问表达式 | `node.isPropertyAccessExpression()` | `node.IsPropertyAccessExpression()` | ✅ 完全兼容 |

#### 🎯 实际示例：类型判断

```typescript
// ts-morph
node.forEachDescendant(child => {
    if (child.isIdentifier()) {
        console.log("Identifier:", child.getText());
    } else if (child.isFunctionDeclaration()) {
        console.log("Function:", child.getName());
    } else if (child.isImportSpecifier()) {
        const importSpec = child.asImportSpecifier();
        console.log("Import:", importSpec.getLocalName());
    }
});
```

```go
// TSMorphGo
sourceFile.ForEachDescendant(func(node Node) {
    switch {
    case node.IsIdentifier():
        fmt.Println("Identifier:", node.GetText())
    case node.IsFunctionDeclaration():
        if funcDecl, ok := node.AsFunctionDeclaration(); ok {
            fmt.Println("Function:", funcDecl.GetName())
        }
    case node.IsImportSpecifier():
        if importSpec, ok := node.AsImportSpecifier(); ok {
            fmt.Println("Import:", importSpec.GetLocalName())
        }
    }
})
```

---

### 4. ImportSpecifier 专用 API

#### ts-morph → TSMorphGo

| 功能 | ts-morph | TSMorphGo | 状态 |
|------|----------|-----------|------|
| 获取别名节点 | `importSpec.getAliasNode()` | `importSpec.GetAliasNode()` | ✅ 完全兼容 |
| 获取原始名称 | (无直接API) | `importSpec.GetOriginalName()` | 🎯 **增强功能** |
| 获取本地名称 | `importSpec.getName()` | `importSpec.GetLocalName()` | ✅ 完全兼容 |
| 判断是否有别名 | (手动检查) | `importSpec.HasAlias()` | 🎯 **增强功能** |
| 类型安全转换 | `node.asImportSpecifier()` | `node.AsImportSpecifier()` | ✅ 完全兼容 |

#### 🎯 实际示例：导入说明符处理

```typescript
// ts-morph
node.forEachDescendant(child => {
    if (child.isImportSpecifier()) {
        const importSpec = child.asImportSpecifier();

        // 获取别名
        const aliasNode = importSpec.getAliasNode();
        if (aliasNode) {
            console.log("Alias:", aliasNode.getText());
        }

        // 获取名称
        const localName = importSpec.getName();
        console.log("Local name:", localName);

        // 检查是否有别名 (手动方式)
        const hasAlias = aliasNode !== undefined;
    }
});
```

```go
// TSMorphGo
sourceFile.ForEachDescendant(func(node Node) {
    if node.IsImportSpecifier() {
        if importSpec, ok := node.AsImportSpecifier(); ok {
            // 获取别名节点
            aliasNode := importSpec.GetAliasNode()
            if aliasNode != nil {
                fmt.Println("Alias:", aliasNode.GetText())
            }

            // 获取本地名称
            localName := importSpec.GetLocalName()
            fmt.Println("Local name:", localName)

            // 获取原始名称 (增强功能)
            originalName := importSpec.GetOriginalName()
            fmt.Println("Original name:", originalName)

            // 判断是否有别名 (增强功能)
            hasAlias := importSpec.HasAlias()
            fmt.Printf("Has alias: %v\n", hasAlias)
        }
    }
})
```

**🌟 TSMorphGo 增强功能示例：**

```go
// TSMorphGo 独有的便利功能
if importSpec.HasAlias() {
    fmt.Printf("Import: %s as %s\n",
        importSpec.GetOriginalName(),
        importSpec.GetLocalName())
} else {
    fmt.Printf("Import: %s\n", importSpec.GetLocalName())
}

// 获取底层parser数据 (增强功能)
if importModule, success := importSpec.GetParserData(); success {
    fmt.Printf("Parser data: %+v\n", importModule)
}
```

---

### 5. 引用查找

#### ts-morph → TSMorphGo

| 功能 | ts-morph | TSMorphGo | 状态 |
|------|----------|-----------|------|
| 查找引用节点 | `node.findReferencesAsNodes()` | `FindReferences(node)` | ✅ 完全兼容 |
| 带缓存的引用查找 | (手动实现) | `FindReferencesWithCache(node)` | 🎯 **增强功能** |
| 重试机制 | (手动实现) | `FindReferencesWithCacheAndRetry()` | 🎯 **增强功能** |

#### 🎯 实际示例：引用查找

```typescript
// ts-morph
const references = someNode.findReferencesAsNodes();
console.log(`Found ${references.length} references`);

// 手动缓存管理
const cache = new Map<string, Node[]>();
const getCachedReferences = (node: Node) => {
    const key = node.getText();
    return cache.get(key) || node.findReferencesAsNodes();
};
```

```go
// TSMorphGo
references, err := FindReferences(someNode)
if err != nil {
    return err
}
fmt.Printf("Found %d references\n", len(references))

// 内置缓存机制 (增强功能)
cachedRefs, cached, err := FindReferencesWithCache(someNode)
if err != nil {
    return err
}
fmt.Printf("Cached: %v, References: %d\n", cached, len(cachedRefs))

// 内置重试机制 (增强功能)
retryConfig := &DefaultRetryConfig()
refs, _, err := FindReferencesWithCacheAndRetry(someNode, retryConfig)
if err != nil {
    return err
}
```

---

### 6. 特定节点类型API

#### CallExpression

| 功能 | ts-morph | TSMorphGo | 状态 |
|------|----------|-----------|------|
| 获取调用表达式 | `node.getExpression()` | `callExpr.GetExpression()` | ✅ 完全兼容 |
| 获取参数列表 | (遍历子节点) | `callExpr.GetArguments()` | 🎯 **增强功能** |
| 获取参数数量 | (手动计数) | `callExpr.GetArgumentCount()` | 🎯 **增强功能** |
| 判断是否为方法调用 | (手动分析) | `callExpr.IsMethodCall()` | 🎯 **增强功能** |

#### 🎯 实际示例：调用表达式

```typescript
// ts-morph
if (node.isCallExpression()) {
    const callExpr = node.asCallExpression();
    const expression = callExpr.getExpression();
    const args = callExpr getArguments();

    console.log("Call:", expression.getText());
    console.log("Arguments:", args.map(a => a.getText()));
}
```

```go
// TSMorphGo
if node.IsCallExpression() {
    if callExpr, ok := node.AsCallExpression(); ok {
        expression := callExpr.GetExpression()
        args := callExpr.GetArguments()

        fmt.Println("Call:", expression.GetText())

        // 增强功能
        fmt.Printf("Argument count: %d\n", callExpr.GetArgumentCount())
        fmt.Printf("Is method call: %v\n", callExpr.IsMethodCall())
        fmt.Printf("Is constructor call: %v\n", callExpr.IsConstructorCall())

        for i, arg := range args {
            fmt.Printf("Arg[%d]: %s\n", i, arg.GetText())
        }
    }
}
```

---

## 🔧 项目迁移

### 基础项目结构迁移

#### ts-morph 项目结构
```
src/
├── analyzer.ts          // 分析器主文件
├── types.ts             // 类型定义
├── utils.ts             // 工具函数
└── tests/
    └── analyzer.test.ts
```

#### TSMorphGo 项目结构
```
cmd/
├── analyzer/            // 主程序入口
│   └── main.go
internal/
├── analyzer/            // 分析器逻辑
│   ├── analyzer.go
│   ├── types.go
│   └── utils.go
pkg/
└── tsmorphgo/           // 可以单独发布的包
    ├── node.go
    ├── project.go
    └── symbol.go
```

### 核心函数迁移模式

#### 1. AST 节点处理

```typescript
// ts-morph 原代码
function processFunction(node: Node) {
    const funcDecl = node.asFunctionDeclaration();
    const name = funcDecl.getName();
    const body = funcDecl.getBody();

    // 处理参数
    const params = funcDecl.getParameters();
    params.forEach(param => {
        console.log("Parameter:", param.getName());
    });

    // 处理返回类型
    const returnType = funcDecl.getReturnType();
    console.log("Return type:", returnType?.getText());
}
```

```go
// TSMorphGo 迁移后代码
func ProcessFunction(node Node) {
    funcDecl, ok := node.AsFunctionDeclaration()
    if !ok {
        return
    }

    name := funcDecl.GetName()
    // body 处理需要通过AST遍历

    // 处理参数 (通过遍历AST子节点)
    funcDecl.GetNode().ForEachChild(func(child *ast.Node) bool {
        // 实现参数处理逻辑
        return false
    })

    fmt.Println("Function name:", name)
}
```

#### 2. 导入语句分析

```typescript
// ts-morph 原代码
function analyzeImports(sourceFile: SourceFile) {
    const imports: ImportInfo[] = [];

    sourceFile.forEachDescendant(node => {
        if (node.isImportSpecifier()) {
            const importSpec = node.asImportSpecifier();
            const localName = importSpec.getName();
            const hasAlias = importSpec.getAliasNode() !== undefined;

            imports.push({ localName, hasAlias });
        }
    });

    return imports;
}
```

```go
// TSMorphGo 迁移后代码
type ImportInfo struct {
    LocalName     string
    OriginalName  string
    HasAlias      bool
}

func AnalyzeImports(sourceFile *SourceFile) []ImportInfo {
    var imports []ImportInfo

    sourceFile.ForEachDescendant(func(node Node) {
        if node.IsImportSpecifier() {
            if importSpec, ok := node.AsImportSpecifier(); ok {
                localName := importSpec.GetLocalName()
                originalName := importSpec.GetOriginalName()
                hasAlias := importSpec.HasAlias()

                imports = append(imports, ImportInfo{
                    LocalName:    localName,
                    OriginalName: originalName,
                    HasAlias:     hasAlias,
                })
            }
        }
    })

    return imports
}
```

---

## 🔄 语法对比

### 类型系统映射

#### TypeScript 类型 → Go 类型

| TypeScript | Go | 说明 |
|------------|-----|------|
| `string` | `string` | 字符串类型 |
| `boolean` | `bool` | 布尔类型 |
| `number` | `int`, `float64` | 数值类型 |
| `T[]` | `[]T` | 数组/切片 |
| `T | null` | `*T` | 指针/可选类型 |
| `Promise<T>` | `(T, error)` | 错误处理模式 |
| `void` | 无返回值 | 函数返回类型 |

#### 错误处理模式

```typescript
// ts-morph (异常/undefined 模式)
function getFunctionName(node: Node): string | undefined {
    if (!node.isFunctionDeclaration()) {
        return undefined;
    }

    const funcDecl = node.asFunctionDeclaration();
    return funcDecl.getName();
}

// 使用方式
const name = getFunctionName(someNode);
if (name) {
    console.log("Function name:", name);
}
```

```go
// TSMorphGo (错误返回模式)
func GetFunctionName(node Node) (string, bool) {
    if !node.IsFunctionDeclaration() {
        return "", false
    }

    funcDecl, ok := node.AsFunctionDeclaration()
    if !ok {
        return "", false
    }

    return funcDecl.GetName(), true
}

// 使用方式
if name, ok := GetFunctionName(someNode); ok {
    fmt.Println("Function name:", name)
}
```

---

## 💡 最佳实践

### 1. 性能优化

#### ts-morph 性能考虑
```typescript
// ts-morph - 可能的性能陷阱
function slowProcessing(project: Project) {
    // 多次遍历可能影响性能
    const allFiles = project.getSourceFiles();
    for (const file of allFiles) {
        file.forEachDescendant(node => {
            // 处理逻辑
        });
    }
}
```

#### TSMorphGo 性能优化
```go
// TSMorphGo - 利用内置缓存
func FastProcessing(project *Project) error {
    // 利用内置的引用缓存
    for _, sourceFile := range project.GetSourceFiles() {
        sourceFile.ForEachDescendant(func(node Node) {
            if node.IsIdentifier() {
                // 使用缓存引用查找
                if refs, cached, err := FindReferencesWithCache(node); err == nil {
                    fmt.Printf("Node: %s, Cached: %v, References: %d\n",
                        node.GetText(), cached, len(refs))
                }
            }
        })
    }
    return nil
}
```

### 2. 内存管理

```typescript
// ts-morph - 自动垃圾回收
function createAnalyzer() {
    const project = new Project();
    // 无需手动清理
}

// TSMorphGo - 显式资源管理
func CreateAnalyzer() {
    project := NewProject(config)
    defer project.Close() // 必须调用，否则内存泄漏

    // 分析逻辑
}
```

### 3. 并发处理

```typescript
// ts-morph - 单线程为主
function processFiles(files: SourceFile[]) {
    files.forEach(file => processFile(file));
}

// TSMorphGo - 原生并发支持
func ProcessFiles(files []*SourceFile) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(files))

    for _, file := range files {
        wg.Add(1)
        go func(sf *SourceFile) {
            defer wg.Done()
            if err := ProcessFile(sf); err != nil {
                errChan <- err
            }
        }(file)
    }

    wg.Wait()
    close(errChan)

    // 检查错误
    return <-errChan
}
```

---

## ❓ 常见问题

### Q1: 如何处理复杂的类型检查？

```typescript
// ts-morph
const type = node.getType();
if (type.isString()) {
    console.log("String type");
}
```

```go
// TSMorphGo
// 注意：当前版本类型检查功能有限，建议通过AST分析实现
func CheckType(node Node) {
    if node.IsKind(KindStringLiteral) {
        fmt.Println("String literal")
    }
    // 或者获取节点文本进行启发式判断
    text := node.GetText()
    if strings.Contains(text, `"`) || strings.Contains(text, `'`) {
        fmt.Println("Likely string literal")
    }
}
```

### Q2: 如何重构代码？

```typescript
// ts-morph
someNode.rename("newName");
someNode.remove();
```

```go
// TSMorphGo
// 注意：重构功能当前有限，建议通过文件级操作
func RefactorNode(node Node, newName string) error {
    sourceFile := node.GetSourceFile()

    // 获取原始代码
    originalText := sourceFile.GetFileResult().Raw

    // 简单的文本替换 (注意：只适用于简单场景)
    newText := strings.ReplaceAll(originalText, node.GetText(), newName)

    // 重新创建源文件 (需要谨慎使用)
    _, err := sourceFile.GetProject().CreateSourceFile(
        sourceFile.GetFilePath(),
        newText,
    )

    return err
}
```

### Q3: 如何处理大文件？

```typescript
// ts-morph
const sourceFile = project.getSourceFile("large-file.ts");

// TypeScript 运行时限制可能导致内存问题
```

```go
// TSMorphGo - 更好的大文件处理
func ProcessLargeFile(filePath string) error {
    config := ProjectConfig{
        RootPath:              "",
        TargetExtensions:      []string{".ts"},
        UseInMemoryFileSystem: false,
        IgnorePatterns:        []string{"node_modules", ".git"},
    }

    project := NewProject(config)
    defer project.Close()

    sourceFile := project.GetSourceFile(filePath)
    if sourceFile == nil {
        return fmt.Errorf("file not found: %s", filePath)
    }

    // 分批处理
    batchSize := 100
    var nodes []Node

    sourceFile.ForEachDescendant(func(node Node) {
        nodes = append(nodes, node)

        if len(nodes) >= batchSize {
            ProcessBatch(nodes)
            nodes = nodes[:0] // 重置切片
        }
    })

    // 处理剩余节点
    if len(nodes) > 0 {
        ProcessBatch(nodes)
    }

    return nil
}
```

---

## 🎯 完整示例

### 示例1: 代码分析器

#### ts-morph 版本
```typescript
// analyzer.ts
import { Project, Node, SyntaxKind } from "ts-morph";

interface FunctionInfo {
    name: string;
    parameters: string[];
    returnTypes: string[];
    callsites: Node[];
}

export class TypeScriptAnalyzer {
    private project: Project;

    constructor(tsConfigPath: string) {
        this.project = new Project({ tsConfigFilePath: tsConfigPath });
    }

    public analyzeFunctions(): FunctionInfo[] {
        const functions: FunctionInfo[] = [];

        this.project.getSourceFiles().forEach(file => {
            file.forEachDescendant(node => {
                if (node.isFunctionDeclaration()) {
                    const funcDecl = node.asFunctionDeclaration();
                    const info: FunctionInfo = {
                        name: funcDecl.getName() || "anonymous",
                        parameters: funcDecl.getParameters().map(p => p.getName()),
                        returnTypes: funcDecl.getReturnType() ? [funcDecl.getReturnType().getText()] : [],
                        callsites: this.findCallSites(funcDecl)
                    };
                    functions.push(info);
                }
            });
        });

        return functions;
    }

    private findCallSites(funcDecl: Node): Node[] {
        const callSites: Node[] = [];

        this.project.getSourceFiles().forEach(file => {
            file.forEachDescendant(node => {
                if (node.isCallExpression()) {
                    const callExpr = node.asCallExpression();
                    if (this.referencesFunction(callExpr, funcDecl)) {
                        callSites.push(node);
                    }
                }
            });
        });

        return callSites;
    }

    private referencesFunction(callExpr: CallExpression, funcDecl: Node): boolean {
        const references = funcDecl.findReferencesAsNodes();
        return references.some(ref =>
            ref.getParent() === callExpr.getExpression()
        );
    }
}

// 使用示例
const analyzer = new TypeScriptAnalyzer("./tsconfig.json");
const functions = analyzer.analyzeFunctions();
console.log(`Found ${functions.length} functions`);
```

#### TSMorphGo 版本
```go
// analyzer.go
package main

import (
    "fmt"
    "sync"

    "github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
    . "github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

type FunctionInfo struct {
    Name        string
    Parameters  []string
    ReturnTypes []string
    CallSites   []*Node
}

type TypeScriptAnalyzer struct {
    project *tsmorphgo.Project
}

func NewTypeScriptAnalyzer(tsConfigPath string) *TypeScriptAnalyzer {
    config := tsmorphgo.ProjectConfig{
        RootPath:    ".",
        UseTsConfig: true,
    }

    project := tsmorphgo.NewProject(config)

    return &TypeScriptAnalyzer{
        project: project,
    }
}

func (a *TypeScriptAnalyzer) Close() {
    a.project.Close()
}

func (a *TypeScriptAnalyzer) AnalyzeFunctions() ([]FunctionInfo, error) {
    var functions []FunctionInfo

    sourceFiles := a.project.GetSourceFiles()
    for _, sourceFile := range sourceFiles {
        sourceFile.ForEachDescendant(func(node Node) {
            if node.IsFunctionDeclaration() {
                if funcDecl, ok := node.AsFunctionDeclaration(); ok {
                    info := FunctionInfo{
                        Name:        funcDecl.GetName(),
                        Parameters:  a.extractParameters(funcDecl),
                        ReturnTypes: a.extractReturnTypes(funcDecl),
                        CallSites:   a.findCallSites(funcDecl),
                    }
                    functions = append(functions, info)
                }
            }
        })
    }

    return functions, nil
}

func (a *TypeScriptAnalyzer) extractParameters(funcDecl *tsmorphgo.FunctionDeclaration) []string {
    var parameters []string

    funcDecl.GetNode().ForEachChild(func(child *ast.Node) bool {
        // 实现参数提取逻辑
        return false
    })

    return parameters
}

func (a *TypeScriptAnalyzer) extractReturnTypes(funcDecl *tsmorphgo.FunctionDeclaration) []string {
    var returnTypes []string

    funcDecl.GetNode().ForEachChild(func(child *ast.Node) bool {
        // 实现返回类型提取逻辑
        return false
    })

    return returnTypes
}

func (a *TypeScriptAnalyzer) findCallSites(funcDecl *tsmorphgo.FunctionDeclaration) []*Node {
    var callSites []*Node

    // 使用内置的引用查找功能
    if references, err := tsmorphgo.FindReferences(*funcDecl.GetNode()); err == nil {
        for _, ref := range references {
            if a.isCallSite(ref) {
                callSites = append(callSites, ref)
            }
        }
    }

    return callSites
}

func (a *TypeScriptAnalyzer) isCallSite(node *tsmorphgo.Node) bool {
    parent := node.GetParent()
    return parent != nil && parent.IsCallExpression()
}

func main() {
    analyzer := NewTypeScriptAnalyzer("./tsconfig.json")
    defer analyzer.Close()

    functions, err := analyzer.AnalyzeFunctions()
    if err != nil {
        panic(err)
    }

    fmt.Printf("Found %d functions\n", len(functions))

    for _, funcInfo := range functions {
        fmt.Printf("Function: %s\n", funcInfo.Name)
        fmt.Printf("  Parameters: %v\n", funcInfo.Parameters)
        fmt.Printf("  Call sites: %d\n", len(funcInfo.CallSites))
    }
}
```

### 示例2: 导入依赖分析器

```go
// dependency_analyzer.go
package main

import (
    "fmt"
    "sort"

    "github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

type DependencyInfo struct {
    ModulePath string
    LocalName  string
    OriginalName string
    HasAlias   bool
    UsedInCode bool
}

type ImportAnalyzer struct {
    project *tsmorphgo.Project
}

func NewImportAnalyzer() *ImportAnalyzer {
    config := tsmorphgo.ProjectConfig{
        RootPath: ".",
        TargetExtensions: []string{".ts", ".tsx"},
    }

    project := tsmorphgo.NewProject(config)

    return &ImportAnalyzer{
        project: project,
    }
}

func (a *ImportAnalyzer) Close() {
    a.project.Close()
}

func (a *ImportAnalyzer) AnalyzeImports() ([]DependencyInfo, error) {
    var dependencies []DependencyInfo

    sourceFiles := a.project.GetSourceFiles()

    for _, sourceFile := range sourceFiles {
        sourceFile.ForEachDescendant(func(node Node) {
            if node.IsImportSpecifier() {
                if importSpec, ok := node.AsImportSpecifier(); ok {
                    dep := DependencyInfo{
                        ModulePath:   a.getModulePath(importSpec),
                        LocalName:    importSpec.GetLocalName(),
                        OriginalName: importSpec.GetOriginalName(),
                        HasAlias:     importSpec.HasAlias(),
                        UsedInCode:   a.isUsedInCode(sourceFile, importSpec.GetLocalName()),
                    }
                    dependencies = append(dependencies, dep)
                }
            }
        })
    }

    return dependencies, nil
}

func (a *ImportAnalyzer) getModulePath(importSpec *tsmorphgo.ImportSpecifier) string {
    // 通过父级ImportDeclaration获取模块路径
    ancestor := importSpec.GetParent()
    for ancestor != nil {
        if ancestor.IsKind(KindImportDeclaration) {
            // 实现模块路径提取逻辑
            return "extracted-module-path"
        }
        ancestor = ancestor.GetParent()
    }
    return ""
}

func (a *ImportAnalyzer) isUsedInCode(sourceFile *tsmorphgo.SourceFile, localName string) bool {
    var used bool
    sourceFile.ForEachDescendant(func(node Node) {
        if node.IsIdentifier() && node.GetText() == localName {
            // 确保这不是导入语句本身的标识符
            if !a.isInImportStatement(node) {
                used = true
            }
        }
    })
    return used
}

func (a *ImportAnalyzer) isInImportStatement(node Node) bool {
    ancestor := node.GetParent()
    for ancestor != nil {
        if ancestor.IsKind(KindImportDeclaration) || ancestor.IsImportSpecifier() {
            return true
        }
        ancestor = ancestor.GetParent()
    }
    return false
}

func main() {
    analyzer := NewImportAnalyzer()
    defer analyzer.Close()

    dependencies, err := analyzer.AnalyzeImports()
    if err != nil {
        panic(err)
    }

    // 按模块路径分组
    moduleGroups := make(map[string][]DependencyInfo)
    for _, dep := range dependencies {
        moduleGroups[dep.ModulePath] = append(moduleGroups[dep.ModulePath], dep)
    }

    fmt.Printf("Import Analysis Report\n")
    fmt.Printf("===================\n\n")

    // 按模块名称排序
    var moduleNames []string
    for moduleName := range moduleGroups {
        moduleNames = append(moduleNames, moduleName)
    }
    sort.Strings(moduleNames)

    totalImports := 0
    totalUnused := 0

    for _, moduleName := range moduleNames {
        deps := moduleGroups[moduleName]
        fmt.Printf("Module: %s\n", moduleName)
        fmt.Printf("  Imports:\n")

        for _, dep := range deps {
            status := "✅ Used"
            if !dep.UsedInCode {
                status = "❌ Unused"
                totalUnused++
            }

            aliasInfo := ""
            if dep.HasAlias {
                aliasInfo = fmt.Sprintf(" (as %s from %s)", dep.LocalName, dep.OriginalName)
            }

            fmt.Printf("    - %s%s %s\n", dep.LocalName, aliasInfo, status)
        }

        fmt.Printf("\n")
        totalImports += len(deps)
    }

    fmt.Printf("Summary:\n")
    fmt.Printf("========\n")
    fmt.Printf("Total imports: %d\n", totalImports)
    fmt.Printf("Used imports: %d\n", totalImports-totalUnused)
    fmt.Printf("Unused imports: %d\n", totalUnused)
    fmt.Printf("Utilization rate: %.1f%%\n",
        float64(totalImports-totalUnused)/float64(totalImports)*100)
}
```

---

## 🎯 总结

### 迁移优势

1. **性能提升**: Go 的高性能编译型语言特性
2. **内存效率**: 更精确的内存控制和垃圾回收
3. **并发支持**: 原生的并发处理能力
4. **类型安全**: 编译时类型检查
5. **部署简单**: 单一可执行文件部署

### 迁移成本

1. **学习成本**: Go 语言的语法和惯用法
2. **生态差异**: JavaScript/TypeScript vs Go 生态系统
3. **调试工具**: 需要适应 Go 的调试工具链

### 建议的迁移策略

1. **渐进式迁移**: 从小型工具或分析脚本开始
2. **并行开发**: 保持原有 TypeScript 代码的同时开发 Go 版本
3. **重点关注**: CPU密集型和内存密集型的分析任务优先迁移

---

**🎉 祝您迁移顺利！如有任何问题，请参考 [TSMorphGo GitHub](https://github.com/Flying-Bird1999/analyzer-ts) 获取更多支持。**