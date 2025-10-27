# `tsmorphgo` API 迁移指南与实战场景

本文档描述了从 `ts-morph` 到 `tsmorphgo` 的 API 迁移方案，以及各种实际使用场景的解决方案。

## 🚀 快速迁移对照表

| ts-morph API | tsmorphgo 等价 API | 状态 | 备注 |
|-------------|---------------------|------|------|
| `project.getSourceFiles()` | `project.GetSourceFile(path)` | ✅ | tsmorphgo 按路径获取文件 |
| `sourceFile.getFilePath()` | `sourceFile.GetFilePath()` | ✅ | 完全兼容 |
| `node.getText()` | `node.GetText()` | ✅ | 完全兼容 |
| `node.getParent()` | `node.GetParent()` | ✅ | 完全兼容 |
| `node.getKind()` | `node.Kind` | ✅ | 使用属性而非方法 |
| `findReferences(node)` | `FindReferences(node)` | ✅ | 函数式调用 |
| `node.isKind(SyntaxKind.XXX)` | `IsXXX(node)` | ✅ | 类型判断函数 |
| `node.asKind(SyntaxKind.XXX)` | `AsXXX(node)` | ✅ | 类型转换函数 |

---

## 1. 项目初始化与管理

### 场景 1.1：基于内存源码创建项目

**ts-morph 原有方式：**

```typescript
// ts-morph: 使用内存文件系统
const project = new Project({
    useInMemoryFileSystem: true,
});

// 添加源文件
project.createSourceFile("test.ts", `
    interface User { id: number; name: string; }
    function getUser(id: number): User {
        return { id, name: `User${id}` };
    }
`);
```

**tsmorphgo 解决方案：**

```go
// tsmorphgo: 直接从内存源码创建
project := tsmorphgo.NewProjectFromSources(map[string]string{
    "test.ts": `
        interface User { id: number; name: string; }
        function getUser(id: number): User {
            return { id, name: "User" + id };
        }
    `,

    // 可选：包含 tsconfig.json 以支持路径别名等高级功能
    "/tsconfig.json": `{
        "compilerOptions": {
            "baseUrl": ".",
            "paths": { "@/*": ["src/*"] }
        }
    }`,
})

// 获取源文件
testFile := project.GetSourceFile("test.ts")
if testFile == nil {
    panic("源文件创建失败")
}
```

**迁移要点：**
- ✅ tsmorphgo 直接支持内存项目创建，无需特殊配置
- ✅ 内置支持 TypeScript 配置和路径别名
- ⚠️ 使用 `map[string]string` 而非文件系统 API
- ⚠️ 按路径获取文件而非批量获取所有文件

---

### 场景 1.2：包含复杂配置的项目创建

**ts-morph 原有方式：**

```typescript
// ts-morph: 复杂项目配置
const project = new Project({
    tsConfigFilePath: "./tsconfig.json",
    skipAddingFilesFromTsConfig: true,
    manipulationSettings: {
        indentationText: "  ",
    },
});

// 手动添加文件
project.addSourceFileAtPath("./src/utils.ts");
project.addSourceFileAtPath("./src/index.ts");
```

**tsmorphgo 解决方案：**

```go
// tsmorphgo: 完整项目配置
project := tsmorphgo.NewProjectFromSources(map[string]string{
    // TypeScript 配置（支持完整语法）
    "/tsconfig.json": `{
        "compilerOptions": {
            "target": "es2018",
            "module": "commonjs",
            "lib": ["es2018", "dom"],
            "declaration": true,
            "outDir": "./dist",
            "rootDir": "./src",
            "strict": true,
            "esModuleInterop": true,
            "skipLibCheck": true,
            "forceConsistentCasingInFileNames": true,
            "baseUrl": ".",
            "paths": {
                "@/*": ["src/*"],
                "@components/*": ["src/components/*"],
                "@utils/*": ["src/utils/*"]
            }
        },
        "include": ["src/**/*"],
        "exclude": ["node_modules", "dist", "**/*.test.ts"]
    }`,

    // 源文件（支持 .ts 和 .tsx）
    "/src/utils.ts": `
        import { Logger } from '@/types';

        export const logger: Logger = {
            log: (message: string) => console.log('[LOG]', message),
            error: (error: Error) => console.error('[ERROR]', error)
        };

        export function formatDate(date: Date): string {
            return date.toISOString().split('T')[0];
        }
    `,

    "/src/index.ts": `
        import { logger, formatDate } from '@/utils';
        import { AppConfig } from '@/config';

        function main() {
            logger.log('Application started');
            const today = formatDate(new Date());
            console.log('Today:', today);
        }

        main();
    `,

    "/src/types.ts": `
        export interface Logger {
            log: (message: string) => void;
            error: (error: Error) => void;
        }

        export interface AppConfig {
            appName: string;
            version: string;
            debug: boolean;
        }
    `,

    "/src/config.ts": `
        import { AppConfig } from '@/types';

        export const config: AppConfig = {
            appName: 'MyApp',
            version: '1.0.0',
            debug: process.env.NODE_ENV === 'development'
        };
    `,
})

// 验证项目结构
utilsFile := project.GetSourceFile("/src/utils.ts")
indexFile := project.GetSourceFile("/src/index.ts")
typesFile := project.GetSourceFile("/src/types.ts")

fmt.Printf("项目创建成功，包含 %d 个源文件\n",
    map[bool]int{true: 1, false: 0}[utilsFile != nil] +
    map[bool]int{true: 1, false: 0}[indexFile != nil] +
    map[bool]int{true: 1, false: 0}[typesFile != nil])
```

**迁移要点：**
- ✅ tsmorphgo 支持 TypeScript 完整配置语法
- ✅ 自动处理路径别名和模块解析
- ⚠️ 一次性提供所有源码，而非动态添加
- ⚠️ 文件路径必须以 `/` 开头

---

## 2. 源文件与节点操作

### 场景 2.1：遍历和分析所有节点

**ts-morph 原有方式：**

```typescript
// ts-morph: 遍历所有节点
function analyzeProject(project: Project): AnalysisResult {
    const result: AnalysisResult = {
        functions: [],
        classes: [],
        interfaces: []
    };

    for (const sourceFile of project.getSourceFiles()) {
        // 获取所有函数声明
        const functions = sourceFile.getFunctions();
        result.functions.push(...functions.map(fn => ({
            name: fn.getName(),
            filePath: sourceFile.getFilePath(),
            line: fn.getStartLineNumber()
        })));

        // 获取所有类声明
        const classes = sourceFile.getClasses();
        result.classes.push(...classes.map(cls => ({
            name: cls.getName(),
            filePath: sourceFile.getFilePath(),
            line: cls.getStartLineNumber(),
            methods: cls.getMethods().map(m => m.getName())
        })));

        // 获取所有接口声明
        const interfaces = sourceFile.getInterfaces();
        result.interfaces.push(...interfaces.map(iface => ({
            name: iface.getName(),
            filePath: sourceFile.getFilePath(),
            line: iface.getStartLineNumber(),
            properties: iface.getProperties().map(p => p.getName())
        })));
    }

    return result;
}
```

**tsmorphgo 解决方案：**

```go
// tsmorphgo: 节点遍历和分析
type AnalysisResult struct {
    Functions  []FunctionInfo
    Classes    []ClassInfo
    Interfaces []InterfaceInfo
}

type FunctionInfo struct {
    Name     string
    FilePath string
    Line     int
}

type ClassInfo struct {
    Name     string
    FilePath string
    Line     int
    Methods  []string
}

type InterfaceInfo struct {
    Name       string
    FilePath   string
    Line       int
    Properties []string
}

func AnalyzeProject(project *tsmorphgo.Project) *AnalysisResult {
    result := &AnalysisResult{}

    // 注意：这里需要获取项目的所有源文件
    // 当前 API 设计为按路径获取，可以根据实际项目结构调整
    filePaths := []string{
        "/src/index.ts",
        "/src/utils.ts",
        "/src/types.ts",
        "/src/config.ts",
    }

    for _, filePath := range filePaths {
        sourceFile := project.GetSourceFile(filePath)
        if sourceFile == nil {
            continue
        }

        // 遍历所有节点进行分类
        sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
            switch {
            case tsmorphgo.IsFunctionDeclaration(node):
                // 处理函数声明
                if nameNode, ok := tsmorphgo.GetFunctionDeclarationNameNode(node); ok {
                    result.Functions = append(result.Functions, FunctionInfo{
                        Name:     strings.TrimSpace(nameNode.GetText()),
                        FilePath: sourceFile.GetFilePath(),
                        Line:     node.GetStartLineNumber(),
                    })
                }

            case tsmorphgo.IsClassDeclaration(node):
                // 处理类声明
                classInfo := ClassInfo{
                    FilePath: sourceFile.GetFilePath(),
                    Line:     node.GetStartLineNumber(),
                }

                // 获取类名
                if nameNode, ok := tsmorphgo.GetFirstChild(node, func(child tsmorphgo.Node) bool {
                    return tsmorphgo.IsIdentifier(child)
                }); ok {
                    classInfo.Name = strings.TrimSpace(nameNode.GetText())
                }

                // 获取方法
                node.ForEachDescendant(func(descendant tsmorphgo.Node) {
                    if tsmorphgo.IsMethodDeclaration(descendant) {
                        if methodName, ok := getMethodName(descendant); ok {
                            classInfo.Methods = append(classInfo.Methods, methodName)
                        }
                    }
                })

                result.Classes = append(result.Classes, classInfo)

            case tsmorphgo.IsInterfaceDeclaration(node):
                // 处理接口声明
                interfaceInfo := InterfaceInfo{
                    FilePath: sourceFile.GetFilePath(),
                    Line:     node.GetStartLineNumber(),
                }

                // 获取接口名
                if nameNode, ok := tsmorphgo.GetFirstChild(node, func(child tsmorphgo.Node) bool {
                    return tsmorphgo.IsIdentifier(child)
                }); ok {
                    interfaceInfo.Name = strings.TrimSpace(nameNode.GetText())
                }

                // 获取属性
                node.ForEachDescendant(func(descendant tsmorphgo.Node) {
                    if descendant.Kind == ast.KindPropertySignature {
                        if propName, ok := getPropertyName(descendant); ok {
                            interfaceInfo.Properties = append(interfaceInfo.Properties, propName)
                        }
                    }
                })

                result.Interfaces = append(result.Interfaces, interfaceInfo)
            }
        })
    }

    return result
}

// 辅助函数
func getMethodName(node tsmorphgo.Node) (string, bool) {
    if nameNode, ok := tsmorphgo.GetFirstChild(node, func(child tsmorphgo.Node) bool {
        return tsmorphgo.IsIdentifier(child)
    }); ok {
        return nameNode.GetText() + "()", true
    }
    return "", false
}

func getPropertyName(node tsmorphgo.Node) (string, bool) {
    if nameNode, ok := tsmorphgo.GetFirstChild(node, func(child tsmorphgo.Node) bool {
        return tsmorphgo.IsIdentifier(child)
    }); ok {
        return nameNode.GetText(), true
    }
    return "", false
}

// 使用示例
func main() {
    project := tsmorphgo.NewProjectFromSources(map[string]string{
        "/src/app.ts": `
            interface UserService {
                getUser(id: number): User;
                saveUser(user: User): void;
            }

            class UserServiceImpl implements UserService {
                getUser(id: number): User {
                    return { id, name: "User" + id };
                }

                saveUser(user: User): void {
                    console.log("Saving user:", user);
                }

                private log(message: string): void {
                    console.log("[LOG]", message);
                }
            }
        `,
    })

    result := AnalyzeProject(project)
    fmt.Printf("分析结果:\n")
    fmt.Printf("  函数: %d 个\n", len(result.Functions))
    fmt.Printf("  类: %d 个\n", len(result.Classes))
    fmt.Printf("  接口: %d 个\n", len(result.Interfaces))

    for _, class := range result.Classes {
        fmt.Printf("  类 %s 有 %d 个方法\n", class.Name, len(class.Methods))
    }
}
```

**迁移要点：**
- ✅ 使用 `ForEachDescendant` 统一遍历 API
- ✅ 使用类型判断函数 `IsXXX` 替代 `getXXX()` 方法
- ✅ 使用专用 API 如 `GetFunctionDeclarationNameNode`
- ⚠️ 需要手动实现节点分类和属性提取
- ⚠️ 当前设计按路径获取文件，需要明确文件列表

---

### 场景 2.2：高级节点查找与过滤

**ts-morph 原有方式：**

```typescript
// ts-morph: 查找特定条件的节点
function findUnusedVariables(project: Project): UnusedVariable[] {
    const unused: UnusedVariable[] = [];

    for (const sourceFile of project.getSourceFiles()) {
        // 获取所有变量声明
        const variables = sourceFile.getVariableDeclarations();

        for (const variable of variables) {
            const varName = variable.getName();

            // 查找该变量的所有引用
            const references = variable.findReferences();
            const usageCount = references.length;

            // 排除导出的变量和类型引用
            const isExported = variable.isExported();
            const typeReferences = references.filter(ref =>
                ref.getNode().getParent()?.getKind() === SyntaxKind.TypeReference
            );

            if (!isExported && usageCount - typeReferences.length <= 1) {
                unused.push({
                    name: varName,
                    filePath: sourceFile.getFilePath(),
                    line: variable.getStartLineNumber(),
                    isTypeOnly: usageCount === typeReferences.length
                });
            }
        }
    }

    return unused;
}
```

**tsmorphgo 解决方案：**

```go
// tsmorphgo: 高级节点查找与过滤
type UnusedVariable struct {
    Name       string
    FilePath   string
    Line       int
    IsTypeOnly bool
}

func FindUnusedVariables(project *tsmorphgo.Project) []UnusedVariable {
    var unused []UnusedVariable

    // 获取项目中的所有源文件
    filePaths := []string{"/src/index.ts", "/src/utils.ts", /* 其他文件路径 */}

    for _, filePath := range filePaths {
        sourceFile := project.GetSourceFile(filePath)
        if sourceFile == nil {
            continue
        }

        // 收集文件中的所有变量声明
        var declarations []struct {
            node      tsmorphgo.Node
            name      string
            isExported bool
        }

        sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
            if tsmorphgo.IsVariableDeclaration(node) {
                if name, ok := tsmorphgo.GetVariableName(node); ok {
                    declarations = append(declarations, struct {
                        node      tsmorphgo.Node
                        name      string
                        isExported bool
                    }{
                        node:      node,
                        name:      name,
                        isExported: isExportedDeclaration(node),
                    })
                }
            }
        })

        // 检查每个变量的使用情况
        for _, decl := range declarations {
            // 查找所有引用
            refs, err := tsmorphgo.FindReferences(decl.node)
            if err != nil {
                fmt.Printf("查找引用失败: %v\n", err)
                continue
            }

            // 分析引用类型
            typeReferenceCount := 0
            totalUsageCount := len(refs)

            for _, ref := range refs {
                // 排除声明本身的引用
                if isDeclarationPosition(ref, decl.node) {
                    totalUsageCount--
                    continue
                }

                // 检查是否是类型引用
                if isTypeReference(ref) {
                    typeReferenceCount++
                }
            }

            // 判断是否未使用
            usageCount := totalUsageCount - typeReferenceCount
            if !decl.isExported && usageCount <= 0 {
                unused = append(unused, UnusedVariable{
                    Name:       decl.name,
                    FilePath:   filePath,
                    Line:       decl.node.GetStartLineNumber(),
                    IsTypeOnly: typeReferenceCount > 0 && usageCount == 0,
                })
            }
        }
    }

    return unused
}

// 辅助函数：检查是否是声明位置
func isDeclarationPosition(ref, decl tsmorphgo.Node) bool {
    refAncestors := ref.GetAncestors()
    for _, ancestor := range refAncestors {
        if ancestor.Kind == decl.Kind {
            // 简化处理：如果找到相同类型的祖先，认为是声明位置
            return strings.TrimSpace(ancestor.GetText()) == strings.TrimSpace(decl.GetText())
        }
    }
    return false
}

// 辅助函数：检查是否是导出声明
func isExportedDeclaration(node tsmorphgo.Node) bool {
    parent := node.GetParent()
    if parent == nil {
        return false
    }

    // 检查父节点是否有 export 关键字
    var hasExport bool
    parent.ForEachDescendant(func(descendant tsmorphgo.Node) {
        if tsmorphgo.IsExportKeyword(descendant) {
            hasExport = true
        }
    })

    return hasExport
}

// 辅助函数：检查是否是类型引用
func isTypeReference(ref tsmorphgo.Node) bool {
    parent := ref.GetParent()
    if parent == nil {
        return false
    }

    // 检查是否在类型注解、类型参数等上下文中
    grandParent := parent.GetParent()
    return parent.Kind == ast.KindTypeReference ||
           parent.Kind == ast.KindTypeParameter ||
           (grandParent != nil && grandParent.Kind == ast.KindTypeAnnotation)
}

// 使用示例
func main() {
    project := tsmorphgo.NewProjectFromSources(map[string]string{
        "/src/test.ts": `
            import { SomeType } from './types';

            export const usedVar = "used";
            const unusedVar = "unused";
            const typeOnlyVar: SomeType = null as any;

            function test() {
                console.log(usedVar);
                // unusedVar 从未被使用
                const localVar: SomeType = "test";
            }
        `,
    })

    unused := FindUnusedVariables(project)
    fmt.Printf("发现 %d 个未使用的变量:\n", len(unused))
    for _, u := range unused {
        status := "完全未使用"
        if u.IsTypeOnly {
            status = "仅用于类型"
        }
        fmt.Printf("  - %s (%s:%d) - %s\n", u.Name, u.FilePath, u.Line, status)
    }
}
```

**迁移要点：**
- ✅ 使用 `FindReferences` 实现引用查找
- ✅ 使用 `ForEachDescendant` 进行节点遍历
- ✅ 使用专用 API 如 `GetVariableName`
- ⚠️ 需要手动实现复杂的引用分析逻辑
- ⚠️ 导出状态和类型引用判断需要自定义实现

---

## 3. 节点导航与关系

### 场景 3.1：复杂节点导航与祖先查找

**ts-morph 原有方式：**

```typescript
// ts-morph: 复杂的节点导航
function analyzeCallChain(callExpr: CallExpression): CallChainAnalysis {
    const analysis: CallChainAnalysis = {
        fullExpression: callExpr.getText(),
        parts: [],
        rootObject: null,
        finalMethod: null
    };

    let current = callExpr.getExpression();

    // 解析调用链：obj.method1().method2().method3()
    while (true) {
        if (current.isKind(SyntaxKind.PropertyAccessExpression)) {
            const propAccess = current.asKindOrThrow(SyntaxKind.PropertyAccessExpression);
            const propName = propAccess.getName();

            analysis.parts.unshift({
                type: 'property',
                name: propName,
                text: propAccess.getText()
            });

            current = propAccess.getExpression();
        } else if (current.isKind(SyntaxKind.CallExpression)) {
            const innerCall = current.asKindOrThrow(SyntaxKind.CallExpression);
            analysis.parts.unshift({
                type: 'call',
                text: innerCall.getText()
            });

            current = innerCall.getExpression();
        } else if (current.isKind(SyntaxKind.Identifier)) {
            analysis.rootObject = {
                name: current.getText(),
                text: current.getText()
            };
            break;
        } else {
            // 对象字面量、this等
            analysis.rootObject = {
                type: 'expression',
                text: current.getText()
            };
            break;
        }
    }

    // 获取最终调用的方法名
    const finalProp = analysis.parts[analysis.parts.length - 1];
    analysis.finalMethod = finalProp.name;

    return analysis;
}
```

**tsmorphgo 解决方案：**

```go
// tsmorphgo: 复杂节点导航
type CallChainPart struct {
    Type  string // "property", "call", "identifier", "expression"
    Name  string // 属性名或标识符名
    Text  string // 完整文本
}

type CallChainAnalysis struct {
    FullExpression string
    Parts          []CallChainPart
    RootObject     *CallChainPart
    FinalMethod    string
}

func AnalyzeCallChain(callExpr tsmorphgo.Node) (*CallChainAnalysis, error) {
    if !tsmorphgo.IsCallExpression(callExpr) {
        return nil, fmt.Errorf("期望调用表达式，实际: %v", callExpr.Kind)
    }

    analysis := &CallChainAnalysis{
        FullExpression: strings.TrimSpace(callExpr.GetText()),
        Parts:          []CallChainPart{},
    }

    // 获取调用的表达式
    expr, ok := tsmorphgo.GetCallExpressionExpression(callExpr)
    if !ok {
        return nil, fmt.Errorf("无法获取调用表达式")
    }

    // 解析调用链
    current := *expr
    parts := []CallChainPart{}

    for {
        switch {
        case tsmorphgo.IsPropertyAccessExpression(current):
            // 处理属性访问
            propName, ok := tsmorphgo.GetPropertyAccessName(current)
            if !ok {
                return nil, fmt.Errorf("无法获取属性名")
            }

            part := CallChainPart{
                Type: "property",
                Name: propName,
                Text: strings.TrimSpace(current.GetText()),
            }
            parts = append([]CallChainPart{part}, parts...) // 前置插入

            objExpr, ok := tsmorphgo.GetPropertyAccessExpression(current)
            if !ok {
                return nil, fmt.Errorf("无法获取属性访问表达式")
            }
            current = *objExpr

        case tsmorphgo.IsCallExpression(current):
            // 处理内部调用
            part := CallChainPart{
                Type: "call",
                Text: strings.TrimSpace(current.GetText()),
            }
            parts = append([]CallChainPart{part}, parts...) // 前置插入

            innerExpr, ok := tsmorphgo.GetCallExpressionExpression(current)
            if !ok {
                return nil, fmt.Errorf("无法获取内部调用表达式")
            }
            current = *innerExpr

        case tsmorphgo.IsIdentifier(current):
            // 根对象是标识符
            analysis.RootObject = &CallChainPart{
                Type: "identifier",
                Name: strings.TrimSpace(current.GetText()),
                Text: strings.TrimSpace(current.GetText()),
            }
            break

        default:
            // 其他类型（对象字面量、this等）
            analysis.RootObject = &CallChainPart{
                Type: "expression",
                Text: strings.TrimSpace(current.GetText()),
            }
            break
        }

        // 检查循环终止条件
        if current.Kind == expr.Kind && strings.TrimSpace(current.GetText()) == strings.TrimSpace(expr.GetText()) {
            break
        }
    }

    analysis.Parts = parts

    // 获取最终调用的方法名
    if len(parts) > 0 {
        finalPart := parts[len(parts)-1]
        if finalPart.Type == "property" {
            analysis.FinalMethod = finalPart.Name
        }
    }

    return analysis, nil
}

// 使用示例
func main() {
    project := tsmorphgo.NewProjectFromSources(map[string]string{
        "/src/test.ts": `
            class UserService {
                getUsers(): User[] { return []; }
                findById(id: number): User { return {} as User; }
            }

            class Cache {
                get(key: string): any { return null; }
            }

            const userService = new UserService();
            const cache = new Cache();

            // 复杂调用链
            const result = cache.get("user").findById(123).name;

            // 简单调用
            userService.getUsers();
        `,
    })

    sourceFile := project.GetSourceFile("/src/test.ts")
    var callExprs []*tsmorphgo.Node

    sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
        if tsmorphgo.IsCallExpression(node) {
            // 过滤掉简单的函数声明
            text := strings.TrimSpace(node.GetText())
            if !strings.Contains(text, "function") && !strings.Contains(text, "=>") {
                callExprs = append(callExprs, &node)
            }
        }
    })

    fmt.Printf("找到 %d 个函数调用:\n", len(callExprs))
    for _, call := range callExprs {
        analysis, err := AnalyzeCallChain(*call)
        if err != nil {
            fmt.Printf("分析失败: %v\n", err)
            continue
        }

        fmt.Printf("调用表达式: %s\n", analysis.FullExpression)
        if analysis.RootObject != nil {
            fmt.Printf("  根对象: %s (%s)\n", analysis.RootObject.Name, analysis.RootObject.Type)
        }
        fmt.Printf("  调用链:\n")
        for i, part := range analysis.Parts {
            fmt.Printf("    %d. %s: %s\n", i+1, part.Type, part.Text)
        }
        fmt.Printf("  最终方法: %s\n\n", analysis.FinalMethod)
    }
}
```

**迁移要点：**
- ✅ 使用专用 API 处理表达式：`GetCallExpressionExpression`, `GetPropertyAccessName`
- ✅ 使用节点导航：`GetParent`, `GetAncestors`
- ✅ 使用类型判断：`IsXXX` 函数
- ⚠️ 需要手动实现复杂的调用链解析逻辑
- ⚠️ 递归和循环处理需要仔细设计

---

### 场景 3.2：类型安全的节点转换与操作

**ts-morph 原有方式：**

```typescript
// ts-morph: 类型安全节点操作
function safeProcessDeclarations(sourceFile: SourceFile): ProcessingResult {
    const result: ProcessingResult = {
        imports: [],
        functions: [],
        classes: [],
        interfaces: [],
        errors: []
    };

    // 类型安全的方式遍历声明
    sourceFile.forEachChild(child => {
        try {
            if (child.isKind(SyntaxKind.ImportDeclaration)) {
                const importDecl = child.asKindOrThrow(SyntaxKind.ImportDeclaration);
                result.imports.push(processImport(importDecl));
            } else if (child.isKind(SyntaxKind.FunctionDeclaration)) {
                const funcDecl = child.asKindOrThrow(SyntaxKind.FunctionDeclaration);
                result.functions.push(processFunction(funcDecl));
            } else if (child.isKind(SyntaxKind.ClassDeclaration)) {
                const classDecl = child.asKindOrThrow(SyntaxKind.ClassDeclaration);
                result.classes.push(processClass(classDecl));
            } else if (child.isKind(SyntaxKind.InterfaceDeclaration)) {
                const interfaceDecl = child.asKindOrThrow(SyntaxKind.InterfaceDeclaration);
                result.interfaces.push(processInterface(interfaceDecl));
            }
        } catch (error) {
            result.errors.push({
                node: child,
                error: error instanceof Error ? error.message : String(error),
                line: child.getStartLineNumber()
            });
        }
    });

    return result;
}

function processImport(importDecl: ImportDeclaration): ImportInfo {
    const moduleSpecifier = importDecl.getModuleSpecifier().getText();
    const defaultImport = importDecl.getDefaultImport()?.getText() || null;
    const namedImports = importDecl.getNamedImports().map(specifier => ({
        name: specifier.getName(),
        alias: specifier.getAliasNode()?.getText() || null
    }));

    return {
        moduleSpecifier,
        defaultImport,
        namedImports,
        isTypeOnly: importDecl.isTypeOnly(),
        text: importDecl.getText()
    };
}
```

**tsmorphgo 解决方案：**

```go
// tsmorphgo: 类型安全的节点转换与操作
type ProcessingResult struct {
    Imports    []ImportInfo
    Functions  []FunctionInfo
    Classes    []ClassInfo
    Interfaces []InterfaceInfo
    Errors     []ProcessingError
}

type ProcessingError struct {
    NodeText string
    Error    string
    Line     int
    Kind     ast.Kind
}

type ImportInfo struct {
    ModuleSpecifier string
    DefaultImport   string
    NamedImports   []NamedImportInfo
    IsTypeOnly     bool
    Text           string
}

type NamedImportInfo struct {
    Name  string
    Alias string
}

type FunctionInfo struct {
    Name       string
    Parameters []ParameterInfo
    ReturnType string
    IsAsync    bool
    IsExported  bool
    Text       string
}

type ParameterInfo struct {
    Name     string
    Type     string
    Optional bool
}

func SafeProcessDeclarations(sourceFile *tsmorphgo.SourceFile) *ProcessingResult {
    result := &ProcessingResult{}

    sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
        // 只处理顶层声明
        if !isTopLevelDeclaration(node) {
            return
        }

        var err error
        switch {
        case tsmorphgo.IsImportDeclaration(node):
            importInfo, processErr := processImportDeclaration(node)
            if processErr != nil {
                result.Errors = append(result.Errors, ProcessingError{
                    NodeText: strings.TrimSpace(node.GetText()),
                    Error:    processErr.Error(),
                    Line:     node.GetStartLineNumber(),
                    Kind:     node.Kind,
                })
            } else {
                result.Imports = append(result.Imports, *importInfo)
            }

        case tsmorphgo.IsFunctionDeclaration(node):
            funcInfo, processErr := processFunctionDeclaration(node)
            if processErr != nil {
                result.Errors = append(result.Errors, ProcessingError{
                    NodeText: strings.TrimSpace(node.GetText()),
                    Error:    processErr.Error(),
                    Line:     node.GetStartLineNumber(),
                    Kind:     node.Kind,
                })
            } else {
                result.Functions = append(result.Functions, *funcInfo)
            }

        case tsmorphgo.IsClassDeclaration(node):
            classInfo, processErr := processClassDeclaration(node)
            if processErr != nil {
                result.Errors = append(result.Errors, ProcessingError{
                    NodeText: strings.TrimSpace(node.GetText()),
                    Error:    processErr.Error(),
                    Line:     node.GetStartLineNumber(),
                    Kind:     node.Kind,
                })
            } else {
                result.Classes = append(result.Classes, *classInfo)
            }

        case tsmorphgo.IsInterfaceDeclaration(node):
            interfaceInfo, processErr := processInterfaceDeclaration(node)
            if processErr != nil {
                result.Errors = append(result.Errors, ProcessingError{
                    NodeText: strings.TrimSpace(node.GetText()),
                    Error:    processErr.Error(),
                    Line:     node.GetStartLineNumber(),
                    Kind:     node.Kind,
                })
            } else {
                result.Interfaces = append(result.Interfaces, *interfaceInfo)
            }
        }
    })

    return result
}

// 处理导入声明
func processImportDeclaration(node tsmorphgo.Node) (*ImportInfo, error) {
    importDecl, ok := tsmorphgo.AsImportDeclaration(node)
    if !ok {
        return nil, fmt.Errorf("节点不是导入声明")
    }

    info := &ImportInfo{
        Text: strings.TrimSpace(node.GetText()),
    }

    // 提取模块说明符（简化处理）
    if strings.Contains(info.Text, "from") {
        parts := strings.Split(info.Text, "from")
        if len(parts) >= 2 {
            info.ModuleSpecifier = strings.TrimSpace(strings.Trim(parts[1], `'"'`))
        }
    }

    // 提取默认导入
    if strings.Contains(info.Text, "import") && !strings.Contains(info.Text, "{") {
        // 简化处理：默认导入
        importPart := strings.Split(info.Text, "import")[1]
        importPart = strings.Split(importPart, "from")[0]
        info.DefaultImport = strings.TrimSpace(importPart)
    }

    // 检查是否是类型导入
    info.IsTypeOnly = strings.Contains(info.Text, "import type")

    // 提取命名导入（简化处理）
    if braceStart := strings.Index(info.Text, "{"); braceStart >= 0 {
        braceEnd := strings.Index(info.Text[braceStart:], "}")
        if braceEnd >= 0 {
            namedImportsText := info.Text[braceStart+1 : braceStart+braceEnd]
            namedImports := strings.Split(namedImportsText, ",")
            for _, namedImport := range namedImports {
                namedImport = strings.TrimSpace(namedImport)
                if namedImport != "" {
                    if strings.Contains(namedImport, " as ") {
                        parts := strings.Split(namedImport, " as ")
                        info.NamedImports = append(info.NamedImports, NamedImportInfo{
                            Name:  strings.TrimSpace(parts[0]),
                            Alias: strings.TrimSpace(parts[1]),
                        })
                    } else {
                        info.NamedImports = append(info.NamedImports, NamedImportInfo{
                            Name: namedImport,
                        })
                    }
                }
            }
        }
    }

    return info, nil
}

// 处理函数声明
func processFunctionDeclaration(node tsmorphgo.Node) (*FunctionInfo, error) {
    funcInfo := &FunctionInfo{
        Text: strings.TrimSpace(node.GetText()),
    }

    // 获取函数名
    if nameNode, ok := tsmorphgo.GetFunctionDeclarationNameNode(node); ok {
        funcInfo.Name = strings.TrimSpace(nameNode.GetText())
    }

    // 检查是否是异步函数
    funcInfo.IsAsync = strings.Contains(funcInfo.Text, "async")

    // 检查是否是导出函数
    funcInfo.IsExported = strings.Contains(funcInfo.Text, "export")

    // 简化处理：提取返回类型
    if colonPos := strings.Index(funcInfo.Text, ":"); colonPos > 0 {
        parenPos := strings.Index(funcInfo.Text, ")")
        if parenPos > 0 && colonPos > parenPos {
            returnPart := funcInfo.Text[colonPos+1:]
            bracePos := strings.Index(returnPart, "{")
            if bracePos > 0 {
                funcInfo.ReturnType = strings.TrimSpace(returnPart[:bracePos])
            } else {
                funcInfo.ReturnType = strings.TrimSpace(returnPart)
            }
        }
    }

    return funcInfo, nil
}

// 处理类声明（简化版本）
func processClassDeclaration(node tsmorphgo.Node) (*ClassInfo, error) {
    // 实现省略，结构类似
    return nil, nil
}

// 处理接口声明（简化版本）
func processInterfaceDeclaration(node tsmorphgo.Node) (*InterfaceInfo, error) {
    // 实现省略，结构类似
    return nil, nil
}

// 辅助函数：检查是否是顶层声明
func isTopLevelDeclaration(node tsmorphgo.Node) bool {
    // 检查祖先链长度，简化判断
    ancestors := node.GetAncestors()
    // 如果直接在 SourceFile 下一层，认为是顶层声明
    return len(ancestors) <= 2
}

// 使用示例
func main() {
    project := tsmorphgo.NewProjectFromSources(map[string]string{
        "/src/complex.ts": `
            import { Logger, type Config } from './types';
            import * as fs from 'fs';
            import React from 'react';

            export interface DataService {
                getData(): Promise<Data[]>;
            }

            class DataServiceImpl implements DataService {
                constructor(private logger: Logger) {}

                async getData(): Promise<Data[]> {
                    this.logger.log('Fetching data...');
                    return [];
                }
            }

            function createService(logger: Logger): DataService {
                return new DataServiceImpl(logger);
            }
        `,
    })

    sourceFile := project.GetSourceFile("/src/complex.ts")
    if sourceFile == nil {
        panic("源文件未找到")
    }

    result := SafeProcessDeclarations(sourceFile)

    fmt.Printf("处理结果:\n")
    fmt.Printf("  导入: %d 个\n", len(result.Imports))
    fmt.Printf("  函数: %d 个\n", len(result.Functions))
    fmt.Printf("  类: %d 个\n", len(result.Classes))
    fmt.Printf("  接口: %d 个\n", len(result.Interfaces))
    fmt.Printf("  错误: %d 个\n", len(result.Errors))

    for _, imp := range result.Imports {
        fmt.Printf("  导入 %s from %s\n", imp.ModuleSpecifier, imp.NamedImports)
    }

    for _, err := range result.Errors {
        fmt.Printf("  错误 [%v]: %s (行 %d)\n", err.Kind, err.Error, err.Line)
    }
}
```

**迁移要点：**
- ✅ 使用类型转换函数：`AsXXX` 配合类型判断
- ✅ 使用专用 API：`GetFunctionDeclarationNameNode`, `GetPropertyAccessName`
- ✅ 健壮的错误处理和类型安全
- ⚠️ 需要手动实现复杂的声明解析逻辑
- ⚠️ 导入/导出信息提取需要字符串处理

---

## 4. 符号与引用查找

### 场景 4.1：符号系统与语义分析

**ts-morph 原有方式：**

```typescript
// ts-morph: 符号系统使用
function analyzeSymbolUsage(project: Project): SymbolAnalysis {
    const analysis: SymbolAnalysis = {
        variables: [],
        functions: [],
        classes: [],
        interfaces: []
    };

    for (const sourceFile of project.getSourceFiles()) {
        // 获取文件中的所有变量符号
        const variableSymbols = sourceFile.getVariableSymbols();

        for (const symbol of variableSymbols) {
            const declarations = symbol.getDeclarations();
            const references = symbol.findReferences();

            analysis.variables.push({
                name: symbol.getName(),
                isExported: symbol.isExported(),
                declarationCount: declarations.length,
                referenceCount: references.length,
                valueType: symbol.getType().getText(),
                declarations: declarations.map(d => ({
                    file: d.getSourceFile().getFilePath(),
                    line: d.getStartLineNumber()
                })),
                references: references.map(r => ({
                    file: r.getSourceFile().getFilePath(),
                    line: r.getStartLineNumber(),
                    isDefinition: r.isDefinition()
                }))
            });
        }

        // 类似地处理函数、类、接口符号...
    }

    return analysis;
}
```

**tsmorphgo 解决方案：**

```go
// tsmorphgo: 符号系统与语义分析
type SymbolAnalysis struct {
    Variables []VariableSymbolInfo
    Functions []FunctionSymbolInfo
    Classes   []ClassSymbolInfo
    Interfaces []InterfaceSymbolInfo
}

type VariableSymbolInfo struct {
    Name            string
    IsExported      bool
    DeclarationCount int
    ReferenceCount   int
    ValueType       string
    Declarations    []DeclarationLocation
    References      []ReferenceLocation
}

type FunctionSymbolInfo struct {
    Name            string
    IsExported      bool
    DeclarationCount int
    ReferenceCount   int
    ReturnType      string
    Parameters      []ParameterSymbolInfo
    Declarations    []DeclarationLocation
    References      []ReferenceLocation
}

type DeclarationLocation struct {
    FilePath string
    Line     int
}

type ReferenceLocation struct {
    FilePath    string
    Line        int
    IsDefinition bool
}

func AnalyzeSymbolUsage(project *tsmorphgo.Project) (*SymbolAnalysis, error) {
    analysis := &SymbolAnalysis{}

    // 获取项目中的所有源文件
    filePaths := []string{"/src/index.ts", "/src/utils.ts", /* 其他文件 */}

    for _, filePath := range filePaths {
        sourceFile := project.GetSourceFile(filePath)
        if sourceFile == nil {
            continue
        }

        // 分析文件中的所有标识符符号
        sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
            if !tsmorphgo.IsIdentifier(node) {
                return
            }

            // 只分析声明位置的符号
            if !isDeclarationIdentifier(node) {
                return
            }

            symbol, found := tsmorphgo.GetSymbol(node)
            if !found {
                return
            }

            // 避免重复分析同一符号
            if isSymbolAlreadyAnalyzed(symbol, analysis) {
                return
            }

            // 根据符号类型进行分析
            if symbol.IsVariable() {
                varInfo := analyzeVariableSymbol(symbol, node)
                analysis.Variables = append(analysis.Variables, varInfo)
            } else if symbol.IsFunction() {
                funcInfo := analyzeFunctionSymbol(symbol, node)
                analysis.Functions = append(analysis.Functions, funcInfo)
            } else if symbol.IsClass() {
                classInfo := analyzeClassSymbol(symbol, node)
                analysis.Classes = append(analysis.Classes, classInfo)
            } else if symbol.IsInterface() {
                interfaceInfo := analyzeInterfaceSymbol(symbol, node)
                analysis.Interfaces = append(analysis.Interfaces, interfaceInfo)
            }
        })
    }

    return analysis, nil
}

// 分析变量符号
func analyzeVariableSymbol(symbol *tsmorphgo.Symbol, node tsmorphgo.Node) VariableSymbolInfo {
    info := VariableSymbolInfo{
        Name:            symbol.GetName(),
        IsExported:      symbol.IsExported(),
        DeclarationCount: symbol.GetDeclarationCount(),
    }

    // 获取声明位置
    if firstDecl, ok := symbol.GetFirstDeclaration(); ok {
        info.Declarations = append(info.Declarations, DeclarationLocation{
            FilePath: firstDecl.GetSourceFile().GetFilePath(),
            Line:     firstDecl.GetStartLineNumber(),
        })
    }

    // 获取所有引用
    if refs, err := symbol.FindReferences(); err == nil {
        info.ReferenceCount = len(refs)
        for _, ref := range refs {
            info.References = append(info.References, ReferenceLocation{
                FilePath:    ref.GetSourceFile().GetFilePath(),
                Line:        ref.GetStartLineNumber(),
                IsDefinition: isDefinitionPosition(ref, node),
            })
        }
    }

    // 提取值类型（简化处理）
    parent := node.GetParent()
    if parent != nil && tsmorphgo.IsVariableDeclaration(*parent) {
        if varDecl, ok := tsmorphgo.AsVariableDeclaration(*parent); ok {
            // 尝试从变量声明文本中提取类型信息
            declText := strings.TrimSpace(varDecl.GetText())
            if colonPos := strings.Index(declText, ":"); colonPos > 0 {
                equalPos := strings.Index(declText, "=")
                if equalPos > colonPos {
                    typePart := strings.TrimSpace(declText[colonPos+1 : equalPos])
                    info.ValueType = typePart
                }
            }
        }
    }

    return info
}

// 分析函数符号（简化版本）
func analyzeFunctionSymbol(symbol *tsmorphgo.Symbol, node tsmorphgo.Node) FunctionSymbolInfo {
    info := FunctionSymbolInfo{
        Name:            symbol.GetName(),
        IsExported:      symbol.IsExported(),
        DeclarationCount: symbol.GetDeclarationCount(),
    }

    // 获取声明位置
    if firstDecl, ok := symbol.GetFirstDeclaration(); ok {
        info.Declarations = append(info.Declarations, DeclarationLocation{
            FilePath: firstDecl.GetSourceFile().GetFilePath(),
            Line:     firstDecl.GetStartLineNumber(),
        })
    }

    // 获取所有引用
    if refs, err := symbol.FindReferences(); err == nil {
        info.ReferenceCount = len(refs)
        for _, ref := range refs {
            info.References = append(info.References, ReferenceLocation{
                FilePath:    ref.GetSourceFile().GetFilePath(),
                Line:        ref.GetStartLineNumber(),
                IsDefinition: isDefinitionPosition(ref, node),
            })
        }
    }

    // 提取返回类型（简化处理）
    funcDeclText := strings.TrimSpace(node.GetParent().GetText())
    if colonPos := strings.Index(funcDeclText, ":"); colonPos > 0 {
        bracePos := strings.Index(funcDeclText[colonPos:], "{")
        if bracePos > 0 {
            info.ReturnType = strings.TrimSpace(funcDeclText[colonPos+1 : colonPos+bracePos])
        }
    }

    return info
}

// 分析类符号（简化版本）
func analyzeClassSymbol(symbol *tsmorphgo.Symbol, node tsmorphgo.Node) ClassSymbolInfo {
    // 类似函数符号分析，省略实现
    return ClassSymbolInfo{}
}

// 分析接口符号（简化版本）
func analyzeInterfaceSymbol(symbol *tsmorphgo.Symbol, node tsmorphgo.Node) InterfaceSymbolInfo {
    // 类似函数符号分析，省略实现
    return InterfaceSymbolInfo{}
}

// 辅助函数
func isDeclarationIdentifier(node tsmorphgo.Node) bool {
    parent := node.GetParent()
    if parent == nil {
        return false
    }

    return tsmorphgo.IsVariableDeclaration(*parent) ||
           tsmorphgo.IsFunctionDeclaration(*parent) ||
           tsmorphgo.IsClassDeclaration(*parent) ||
           tsmorphgo.IsInterfaceDeclaration(*parent)
}

func isSymbolAlreadyAnalyzed(symbol *tsmorphgo.Symbol, analysis *SymbolAnalysis) bool {
    symbolName := symbol.GetName()

    for _, varInfo := range analysis.Variables {
        if varInfo.Name == symbolName {
            return true
        }
    }

    for _, funcInfo := range analysis.Functions {
        if funcInfo.Name == symbolName {
            return true
        }
    }

    // 类似检查其他类型...

    return false
}

func isDefinitionPosition(ref, definition tsmorphgo.Node) bool {
    return strings.TrimSpace(ref.GetText()) == strings.TrimSpace(definition.GetText()) &&
           ref.GetSourceFile().GetFilePath() == definition.GetSourceFile().GetFilePath()
}

// 使用示例
func main() {
    project := tsmorphgo.NewProjectFromSources(map[string]string{
        "/src/symbols.ts": `
            interface Logger {
                log(message: string): void;
            }

            const logger: Logger = {
                log: (msg) => console.log(msg)
            };

            export function processData(data: string): string {
                logger.log('Processing: ' + data);
                return data.toUpperCase();
            }

            class Service {
                constructor(private logger: Logger) {}

                doWork(): void {
                    this.logger.log('Working...');
                    processData('test');
                }
            }

            // 使用各种符号
            const result = processData('hello');
            logger.log('Done');
        `,
    })

    analysis, err := AnalyzeSymbolUsage(project)
    if err != nil {
        panic(err)
    }

    fmt.Printf("符号分析结果:\n")
    fmt.Printf("  变量符号: %d 个\n", len(analysis.Variables))
    fmt.Printf("  函数符号: %d 个\n", len(analysis.Functions))
    fmt.Printf("  类符号: %d 个\n", len(analysis.Classes))
    fmt.Printf("  接口符号: %d 个\n\n", len(analysis.Interfaces))

    for _, varInfo := range analysis.Variables {
        fmt.Printf("变量 %s:\n", varInfo.Name)
        fmt.Printf("  导出: %v\n", varInfo.IsExported)
        fmt.Printf("  声明数: %d\n", varInfo.DeclarationCount)
        fmt.Printf("  引用数: %d\n", varInfo.ReferenceCount)
        fmt.Printf("  类型: %s\n", varInfo.ValueType)
        fmt.Printf("  声明位置: %s:%d\n", varInfo.Declarations[0].FilePath, varInfo.Declarations[0].Line)
        fmt.Println()
    }

    for _, funcInfo := range analysis.Functions {
        fmt.Printf("函数 %s:\n", funcInfo.Name)
        fmt.Printf("  导出: %v\n", funcInfo.IsExported)
        fmt.Printf("  返回类型: %s\n", funcInfo.ReturnType)
        fmt.Printf("  引用数: %d\n", funcInfo.ReferenceCount)
        fmt.Println()
    }
}
```

**迁移要点：**
- ✅ 使用 `GetSymbol` 获取符号信息
- ✅ 使用 `FindReferences` 查找引用
- ✅ 使用符号方法：`GetName()`, `IsExported()`, `GetDeclarationCount()`
- ✅ 使用符号导航：`GetFirstDeclaration()`, `GetDeclarations()`
- ⚠️ 符号系统相对底层，需要手动实现高级分析逻辑
- ⚠️ 类型信息获取有限，需要文本解析作为补充

---

## 5. 迁移最佳实践

### 5.1 性能优化策略

#### 避免重复遍历

**❌ ts-morph 低效模式：**

```typescript
// ts-morph: 多次遍历相同节点
const functions = sourceFile.getFunctions();
const classes = sourceFile.getClasses();
const variables = sourceFile.getVariableDeclarations();
// 每次调用都会遍历整个 AST
```

**✅ tsmorphgo 高效模式：**

```go
// tsmorphgo: 单次遍历收集多种信息
func efficientAnalysis(sourceFile *tsmorphgo.SourceFile) (*AnalysisResult, error) {
    result := &AnalysisResult{}

    sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
        switch {
        case tsmorphgo.IsFunctionDeclaration(node):
            // 一次遍历处理所有类型
            if funcInfo := processFunction(node); funcInfo != nil {
                result.Functions = append(result.Functions, *funcInfo)
            }

        case tsmorphgo.IsClassDeclaration(node):
            if classInfo := processClass(node); classInfo != nil {
                result.Classes = append(result.Classes, *classInfo)
            }

        case tsmorphgo.IsVariableDeclaration(node):
            if varInfo := processVariable(node); varInfo != nil {
                result.Variables = append(result.Variables, *varInfo)
            }
        }
    })

    return result, nil
}
```

#### 使用条件筛选提前终止

**❌ ts-morph 全量处理：**

```typescript
// ts-morph: 总是处理所有节点
const allNodes = sourceFile.getDescendants();
allNodes.forEach(node => {
    // 即使找到了目标节点，也会继续遍历
});
```

**✅ tsmorphgo 条件终止：**

```go
// tsmorphgo: 条件满足时提前终止
targetNode, found := sourceFile.ForEachDescendantUntil(func(node tsmorphgo.Node) bool {
    return tsmorphgo.IsIdentifier(node) &&
           strings.TrimSpace(node.GetText()) == "targetFunction" &&
           tsmorphgo.IsFunctionDeclaration(node.GetParent())
})

if found {
    // 处理找到的目标节点
    processTargetFunction(targetNode)
}
```

### 5.2 错误处理与健壮性

#### 类型安全转换

**❌ ts-morph 不安全转换：**

```typescript
// ts-morph: 可能运行时错误
const funcDecl = node.asKind(SyntaxKind.FunctionDeclaration); // 可能抛出异常
const name = funcDecl.getName();
```

**✅ tsmorphgo 安全转换：**

```go
// tsmorphgo: 安全的类型检查和转换
if tsmorphgo.IsFunctionDeclaration(node) {
    if funcDecl, ok := tsmorphgo.AsFunctionDeclaration(node); ok {
        if nameNode, ok := tsmorphgo.GetFunctionDeclarationNameNode(*funcDecl); ok {
            name := strings.TrimSpace(nameNode.GetText())
            // 安全使用 name
        }
    }
}
```

#### 引用查找错误处理

**❌ ts-morph 简单处理：**

```typescript
// ts-morph: 可能忽略错误
const references = node.findReferences(); // 可能失败但没有处理
references.forEach(ref => {
    // 使用 ref，可能无效
});
```

**✅ tsmorphgo 健壮处理：**

```go
// tsmorphgo: 完整的错误处理
refs, err := tsmorphgo.FindReferences(node)
if err != nil {
    log.Printf("查找引用失败: %v", err)
    return nil, fmt.Errorf("符号分析失败: %w", err)
}

if len(refs) == 0 {
    log.Printf("警告: 符号 %s 没有找到引用", node.GetText())
    // 可以继续处理，只是没有引用
}

validRefs := make([]tsmorphgo.Node, 0, len(refs))
for _, ref := range refs {
    if ref.GetSourceFile() != nil {
        validRefs = append(validRefs, ref)
    }
}
```

### 5.3 代码组织和可维护性

#### 封装常用操作

```go
// tsmorphgo: 封装高级分析逻辑
type CodeAnalyzer struct {
    project *tsmorphgo.Project
    cache   *AnalysisCache
}

func NewCodeAnalyzer(project *tsmorphgo.Project) *CodeAnalyzer {
    return &CodeAnalyzer{
        project: project,
        cache:   NewAnalysisCache(),
    }
}

func (a *CodeAnalyzer) AnalyzeFile(filePath string) (*FileAnalysis, error) {
    // 检查缓存
    if cached := a.cache.Get(filePath); cached != nil {
        return cached, nil
    }

    sourceFile := a.project.GetSourceFile(filePath)
    if sourceFile == nil {
        return nil, fmt.Errorf("文件不存在: %s", filePath)
    }

    analysis := a.analyzeSourceFile(sourceFile)
    a.cache.Set(filePath, analysis)

    return analysis, nil
}

func (a *CodeAnalyzer) analyzeSourceFile(sourceFile *tsmorphgo.SourceFile) *FileAnalysis {
    // 集中实现分析逻辑
    // ...
}
```

#### 使用接口抽象

```go
// tsmorphgo: 使用接口实现可扩展的处理
type NodeProcessor interface {
    CanProcess(node tsmorphgo.Node) bool
    Process(node tsmorphgo.Node) error
    GetResults() interface{}
}

type FunctionProcessor struct {
    results []FunctionInfo
}

func (p *FunctionProcessor) CanProcess(node tsmorphgo.Node) bool {
    return tsmorphgo.IsFunctionDeclaration(node)
}

func (p *FunctionProcessor) Process(node tsmorphgo.Node) error {
    // 处理函数节点
    return nil
}

func (p *FunctionProcessor) GetResults() interface{} {
    return p.results
}

// 使用处理器管道
func ProcessWithPipeline(sourceFile *tsmorphgo.SourceFile, processors []NodeProcessor) error {
    sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
        for _, processor := range processors {
            if processor.CanProcess(node) {
                if err := processor.Process(node); err != nil {
                    log.Printf("处理节点失败: %v", err)
                }
            }
        }
    })
    return nil
}
```

---

## 6. 迁移检查清单

### 6.1 基础 API 迁移检查

- [ ] 项目创建：`NewProjectFromSources` 替代 `new Project()`
- [ ] 文件获取：`GetSourceFile(path)` 替代 `getSourceFiles()`
- [ ] 节点遍历：`ForEachDescendant` 替代 `getDescendants()`
- [ ] 节点导航：`GetParent()`, `GetAncestors()` 兼容
- [ ] 节点文本：`GetText()` 兼容
- [ ] 节点类型：`node.Kind` 替代 `node.getKind()`

### 6.2 类型系统迁移检查

- [ ] 类型判断：`IsXXX()` 函数替代 `isKind(SyntaxKind.XXX)`
- [ ] 类型转换：`AsXXX()` 函数替代 `asKind(SyntaxKind.XXX)`
- [ ] 专用 API：使用 `GetVariableName()`, `GetPropertyAccessName()` 等
- [ ] 错误处理：使用 `ok` 模式进行类型安全转换

### 6.3 符号系统迁移检查

- [ ] 符号获取：`GetSymbol()` 替代 `node.getSymbol()`
- [ ] 引用查找：`FindReferences()` 替代 `findReferences()`
- [ ] 符号信息：`GetName()`, `IsExported()` 等方法使用
- [ ] 声明访问：`GetFirstDeclaration()`, `GetDeclarations()` 使用

### 6.4 高级功能迁移检查

- [ ] 表达式处理：使用 `GetCallExpressionExpression()` 等 API
- [ ] 声明处理：使用 `GetFunctionDeclarationNameNode()` 等 API
- [ ] 复杂分析：实现自定义的分析逻辑
- [ ] 性能优化：实现缓存和批量处理

---

## 7. 常见问题与解决方案

### 7.1 编译错误

**问题：** `cannot use node (type Node) as type *Node in argument`

**解决方案：**
```go
// ❌ 错误：传递值类型
FindReferences(node)

// ✅ 正确：传递指针
FindReferences(&node)

// 或者直接使用值（根据 API 设计）
FindReferences(node)
```

**问题：** `undefined: AsXXX` 或 `undefined: IsXXX`

**解决方案：**
```go
// 确保导入正确的包
import "github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"

// 检查函数名拼写
// 应该是 AsImportDeclaration, AsVariableDeclaration 等
```

### 7.2 逻辑错误

**问题：** 找不到期望的节点

**调试方法：**
```go
// 添加调试信息
sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
    fmt.Printf("节点: %v, 文本: %s\n", node.Kind, strings.TrimSpace(node.GetText()[:50]))
})

// 检查节点类型是否匹配
if tsmorphgo.IsIdentifier(node) {
    fmt.Println("找到标识符:", strings.TrimSpace(node.GetText()))
}
```

**问题：** `FindReferences` 返回空或错误

**调试方法：**
```go
// 检查输入节点是否有效
if node.GetSourceFile() == nil {
    fmt.Println("节点没有关联的源文件")
    return
}

// 检查节点是否是符号声明
if !isDeclarationNode(node) {
    fmt.Println("节点不是声明节点")
    return
}

// 尝试获取符号
symbol, found := tsmorphgo.GetSymbol(node)
if !found {
    fmt.Println("无法获取节点符号")
    return
}
```

### 7.3 性能问题

**问题：** 分析大型项目时性能较慢

**优化方案：**
```go
// 1. 使用缓存机制
type AnalysisCache struct {
    data map[string]interface{}
    mu   sync.RWMutex
}

// 2. 批量处理节点
func batchProcess(sourceFile *tsmorphgo.SourceFile) {
    var batch []tsmorphgo.Node
    batchSize := 100

    sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
        batch = append(batch, node)
        if len(batch) >= batchSize {
            processBatch(batch)
            batch = batch[:0] // 清空 batch
        }
    })

    if len(batch) > 0 {
        processBatch(batch)
    }
}

// 3. 并行处理（如果线程安全）
func parallelProcess(sourceFiles []*tsmorphgo.SourceFile) {
    var wg sync.WaitGroup
    for _, sf := range sourceFiles {
        wg.Add(1)
        go func(file *tsmorphgo.SourceFile) {
            defer wg.Done()
            analyzeFile(file)
        }(sf)
    }
    wg.Wait()
}
```

---

## 8. 总结

### 8.1 迁移成功的关键

1. **理解 API 设计差异**：tsmorphgo 采用函数式 API 而非面向对象
2. **掌握类型安全模式**：使用 `IsXXX` + `AsXXX` 的安全转换模式
3. **实现自定义逻辑**：一些高级功能需要手动实现
4. **注重性能优化**：避免重复遍历，使用缓存和批量处理
5. **完善错误处理**：Go 的错误处理模式与 TypeScript 不同

### 8.2 后续优化方向

1. **API 增强**：基于使用反馈完善高级 API
2. **性能优化**：实现更智能的缓存和索引机制
3. **功能扩展**：添加更多 TypeScript 语言特性的支持
4. **工具集成**：与现有工具链的深度集成

### 8.3 支持与资源

- **API 文档**：详细的 API 参考和使用示例
- **测试用例**：覆盖所有主要使用场景的测试
- **示例代码**：各种迁移场景的完整示例
- **社区支持**：通过 issue 和讨论获得帮助

通过遵循本迁移指南，您可以成功将基于 ts-morph 的项目迁移到 tsmorphgo，并充分利用 Go 语言的优势构建高性能的代码分析工具。

---

**最后更新**: 2024-10-27
**版本**: tsmorphgo v0.1
**维护者**: Flying-Bird1999