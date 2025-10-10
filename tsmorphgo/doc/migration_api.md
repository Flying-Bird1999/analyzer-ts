## 1. 项目初始化与管理

### 场景 1.1：基于 tsconfig.json 创建项目

**目前如何解决：**

```typescript
// 在 TsParser 构造函数中创建项目
const tsMorphProject = new TsMorph.Project({
    tsConfigFilePath,  // 传入 tsconfig.json 的路径
});

// 为什么这么用：
// 1. ts-morph 需要 TypeScript 的编译选项来正确解析代码
// 2. tsconfig.json 包含了项目的所有配置（paths、include、exclude 等）
// 3. 自动加载 tsconfig 中指定的所有源文件
```

**期望的 API 能力：**

* **基于 tsconfig.json 创建项目**
* **自动加载配置中的文件**
* **支持 TypeScript 编译选项**

**期望的使用姿势：**

```typescript
interface ProjectOptions {
    tsConfigFilePath: string;
}

class Project {
    constructor(options: ProjectOptions);
}
```

---

### 场景 1.2：创建测试用的内存文件系统项目

**目前如何解决：**

```typescript
class TestTsProject extends TsProject {
    constructor({ logger }: { logger: Logger }) {
        const tsMorphProject = new TsMorph.Project({
            useInMemoryFileSystem: true,        // 使用内存文件系统，不读取真实文件
            skipAddingFilesFromTsConfig: true,  // 不从 tsconfig 加载文件
        });
        // ...
    }
}

// 为什么这么用：
// 1. 单元测试不应该依赖真实的文件系统
// 2. 内存文件系统速度更快，测试执行效率高
// 3. skipAddingFilesFromTsConfig 避免加载不需要的文件
```

**期望的 API 能力：**

* **支持内存文件系统（用于测试）**
* **可选择是否自动加载 tsconfig 中的文件**

**期望的使用姿势：**

```typescript
interface ProjectOptions {
    tsConfigFilePath?: string;
    useInMemoryFileSystem?: boolean;
    skipAddingFilesFromTsConfig?: boolean;
}

class Project {
    constructor(options: ProjectOptions);
}
```

---

## 2. 源文件操作

### 场景 2.1：获取项目中的所有源文件

**目前如何解决：**

```typescript
findNodes(
    tapNode: (n: Node) => { isAnalyzeEntry: boolean; node: Node },
    opts: { ignores: string[] }
): Node[] {
    const sourceFiles = this.project.getSourceFiles();  // 获取所有源文件
  
    for (const sourceFile of sourceFiles) {
        // 遍历每个源文件，进行分析
        if (this.shouldIgnoreFile(sourceFile.getFilePath(), opts.ignores)) {
            continue;
        }
        // ...
    }
}

// 为什么这么用：
// 1. 需要遍历项目中所有的 TypeScript 文件
// 2. 对每个文件进行 AST 分析，查找目标节点
// 3. 配合 ignores 选项过滤不需要分析的文件
```

**期望的 API 能力：**

* **获取项目中所有源文件列表**
* **返回可迭代的文件集合**

**期望的使用姿势：**

```typescript
interface Project {
    getSourceFiles(): SourceFile[];
}
```

---

### 场景 2.2：动态创建源文件（测试场景）

**目前如何解决：**

```typescript
class TestTsProject extends TsProject {
    addTestSourceFile(fileName: string, content: string): void {
        // 在内存中创建一个新的源文件
        this.project.createSourceFile(fileName, content);
    }
}

// 使用示例（假设在测试文件中）：
const testProject = createTestTsProject({ debug: true });
testProject.addTestSourceFile('test.ts', `
    const KEY_A = 'keyA';
    const obj = { key_a: KEY_A };
`);

// 为什么这么用：
// 1. 单元测试需要构造特定的代码场景
// 2. 动态创建文件比维护测试文件更灵活
// 3. 在内存中创建，不污染文件系统
```

**期望的 API 能力：**

* **动态创建源文件**
* **支持内存文件系统**

**期望的使用姿势：**

```typescript
interface Project {
    createSourceFile(fileName: string, content: string): SourceFile;
}
```

---

### 场景 2.3：获取源文件的路径信息

**目前如何解决：**

```typescript
// 在多个地方使用
sourceFile.getFilePath()  // 获取文件的完整路径

// 示例 1：判断是否应该忽略文件
if (this.shouldIgnoreFile(sourceFile.getFilePath(), opts.ignores)) {
    this.logger.debug(`Ignoring file: ${sourceFile.getFilePath()}`);
    continue;
}

// 示例 2：生成节点 ID 时包含文件路径
static generateNodeId(astNode: Node["astNode"]): string {
    const filePath = astNode.getSourceFile().getFilePath();
    const line = astNode.getStartLineNumber();
    const col = astNode.getStart() - astNode.getStartLinePos() + 1;
    const name = getAstNodeName(astNode);
    const kind = astNode.getKindName();
    // 生成唯一 ID: "path/to/file.ts:10:5 Identifier:variableName"
    return `${filePath}:${line}:${col} ${kind}:${name}`;
}

// 为什么这么用：
// 1. 文件路径是节点唯一标识的一部分
// 2. 用于日志输出，方便定位问题
// 3. 用于文件过滤（ignores 配置）
```

**期望的 API 能力：**

* **获取源文件的完整路径**

**期望的使用姿势：**

```typescript
interface SourceFile {
    getFilePath(): string;
}
```

---

## 3. 节点遍历

### 场景 3.1：深度优先遍历源文件的所有子节点

**目前如何解决：**

```typescript
findNodes(
    tapNode: (n: Node) => { isAnalyzeEntry: boolean; node: Node },
    opts: { ignores: string[] }
): Node[] {
    const nodes: Node[] = [];
  
    for (const sourceFile of sourceFiles) {
        // 遍历源文件的所有后代节点（深度优先）
        sourceFile.forEachDescendant((node) => {
            const customNode: Node = TsNode.of(node);
          
            // tapNode 是一个回调函数，用于判断节点是否是分析入口
            const result = tapNode(customNode);
            if (result.isAnalyzeEntry) {
                this.logger.debug(
                    `Found analyze entry node: ${customNode.id}, ` +
                    `kind: ${node.getKind()} (${node.getKindName()})`
                );
                nodes.push(result.node);
            }
        });
    }
  
    return nodes;
}

// 为什么这么用：
// 1. 需要遍历整个 AST 树查找目标节点
// 2. forEachDescendant 提供深度优先遍历
// 3. 回调函数模式允许在遍历过程中执行自定义逻辑
// 4. 不需要手动管理递归，API 内部处理
```

**期望的 API 能力：**

* **深度优先遍历所有后代节点**
* **提供回调函数处理每个节点**
* **自动处理递归逻辑**

**期望的使用姿势：**

```typescript
interface SourceFile {
    forEachDescendant(callback: (node: Node) => void): void;
}
```

---

### 场景 3.2：获取节点的父节点

**目前如何解决：**

```typescript
// 在 TsNode 类中
static getParent(node: Node): TsNode | undefined {
    const parent = node.astNode.getParent();  // 获取父节点
    if (parent) {
        return TsNode.of(parent);  // 包装成自定义 Node 类型
    }
    return undefined;
}

// 使用示例 1：判断父节点类型来决定如何查找引用
if (
    TsMorph.Node.isIdentifier(targetAstNode) &&
    (isAssignmentNode(targetAstNode.getParent()) ||           // 父节点是赋值
     TsMorph.Node.isVariableDeclaration(targetAstNode.getParent()) ||  // 父节点是变量声明
     TsMorph.Node.isFunctionDeclaration(targetAstNode.getParent()) ||  // 父节点是函数声明
     TsMorph.Node.isImportSpecifier(targetAstNode.getParent()))        // 父节点是 import
) {
    // 对于标识符节点，只有在特定父节点类型下才查找引用
    const directRefAstNodes = targetAstNode.findReferencesAsNodes();
    // ...
}

// 使用示例 2：判断节点在对象属性赋值中的位置
if (TsMorph.Node.isPropertyAssignment(targetNode.getParent())) {
    // 如果父节点是属性赋值（如 key_a: KEY_A），进行特殊处理
    // ...
}

// 为什么这么用：
// 1. 父节点的类型决定了当前节点的语义上下文
// 2. 不同的父节点类型需要不同的分析策略
// 3. 向上查找可以确定节点在代码结构中的位置
```

**期望的 API 能力：**

* **获取当前节点的直接父节点**
* **可能返回 undefined（根节点）**

**期望的使用姿势：**

```typescript
interface Node {
    getParent(): Node | undefined;
}
```

---

### 场景 3.3：获取节点的所有祖先节点

**目前如何解决：**

```typescript
// 用于判断节点是否在类型相关的上下文中
const isTypeRelatedNode = (node: Node["astNode"]) => {
    const predicators = [
        TsMorph.Node.isTypeAliasDeclaration,    // type Foo = ...
        TsMorph.Node.isInterfaceDeclaration,    // interface Foo { ... }
    ];
    const predicate = (node: Node["astNode"]) =>
        predicators.some((pred) => pred(node));
  
    // 检查当前节点或任一祖先节点是否是类型声明
    return predicate(node) || node.getAncestors().some(predicate);
};

if (isTypeRelatedNode(node.astNode)) {
    return [];  // 跳过类型节点，因为类型不影响运行时行为
}

// 为什么这么用：
// 1. 需要判断节点是否在类型声明的上下文中
// 2. 例如：interface Foo { bar: string } 中的 bar 不应该被分析
// 3. getAncestors() 返回从父节点到根节点的所有祖先
// 4. 使用 some() 检查是否有任一祖先满足条件
```

**期望的 API 能力：**

* **获取从当前节点到根节点的所有祖先**
* **返回祖先节点数组**

**期望的使用姿势：**

```typescript
interface Node {
    getAncestors(): Node[];
}
```

---

### 场景 3.4：按语法类型查找特定的祖先节点

**目前如何解决：**

```typescript
// 示例 1：查找节点所在的对象字面量
const targetObjectLiteral = targetNode.getFirstAncestorByKind(
    TsMorph.SyntaxKind.ObjectLiteralExpression
);

// 使用场景：判断节点是否在对象字面量中
// 例如：const obj = { key_a: KEY_A }
//       当前节点是 KEY_A，需要找到外层的 { key_a: KEY_A }

// 示例 2：查找节点所在的变量声明
const targetVariableDecl = targetObjectLiteral.getFirstAncestorByKind(
    TsMorph.SyntaxKind.VariableDeclaration
);

// 使用场景：确定对象字面量是否赋值给某个变量
// 例如：const obj = { key_a: KEY_A }
//       从 { key_a: KEY_A } 向上查找到 const obj = ...

// 示例 3：在 findNodesWhomUse 中的复杂使用
if (TsMorph.Node.isPropertyAssignment(ancestor)) {
    // 对于属性赋值节点，查找最外层的对象字面量
    const outmostLiteral = getOutmostAncestorByKind(
        ancestor.getFirstAncestorByKind(
            TsMorph.SyntaxKind.ObjectLiteralExpression
        ),
        TsMorph.SyntaxKind.ArrayLiteralExpression,
        TsMorph.SyntaxKind.ObjectLiteralExpression
    );
  
    // 使用场景：处理嵌套的对象/数组字面量
    // 例如：const arr = [{ inner: { key_a: KEY_A } }]
    //       需要找到最外层的数组 [...]
}

// 为什么这么用：
// 1. 向上查找特定类型的祖先节点，确定节点的语义上下文
// 2. getFirstAncestorByKind 只返回第一个匹配的祖先，效率高
// 3. 常用于判断节点在什么结构中（对象、数组、函数、类等）
```

**期望的 API 能力：**

* **根据语法类型查找第一个匹配的祖先节点**
* **返回找到的节点或 undefined**

**期望的使用姿势：**

```typescript
enum SyntaxKind {
    ObjectLiteralExpression,
    ArrayLiteralExpression,
    VariableDeclaration,
    // ... 其他类型
}

interface Node {
    getFirstAncestorByKind(kind: SyntaxKind): Node | undefined;
}
```

---

### 场景 3.5：按自定义条件查找子节点

**目前如何解决：**

```typescript
// 在对象属性赋值场景中，查找属性名（标识符）
if (TsMorph.Node.isPropertyAssignment(ancestor)) {
    // ...
  
    // 查找第一个标识符类型的子节点
    const userAstNode = ancestor.getFirstChild((n) =>
        TsMorph.Node.isIdentifier(n)
    )!;
  
    return {
        continue: false,
        result: [userAstNode, maybeVarDecl],
    };
}

// 使用场景：{ key_a: KEY_A } 中的 PropertyAssignment 有两个关键子节点
// - 左侧：key_a (标识符)
// - 右侧：KEY_A (标识符)
// 通过 getFirstChild 找到第一个标识符（即左侧的 key_a）

// 为什么这么用：
// 1. 属性赋值节点包含多个子节点，需要精确定位
// 2. 使用谓词函数灵活筛选
// 3. 只返回第一个匹配的子节点，避免过度遍历
```

**期望的 API 能力：**

* **根据自定义条件查找第一个匹配的子节点**
* **支持谓词函数**

**期望的使用姿势：**

```typescript
interface Node {
    getFirstChild(predicate: (node: Node) => boolean): Node | undefined;
}
```

---

## 4. 节点类型判断

### 场景：判断节点的具体语法类型

**目前如何解决：**

```typescript
// 示例 1：判断是否是标识符节点
if (TsMorph.Node.isIdentifier(targetAstNode)) {
    // 只有标识符才能查找引用
    // 例如：变量名、函数名、属性名等
    const directRefAstNodes = targetAstNode.findReferencesAsNodes();
}

// 示例 2：判断是否是调用表达式
case TsMorph.Node.isCallExpression(astNode):
    // 对于函数调用，取表达式部分作为名称
    // 例如：foo() -> 取 "foo"
    //      obj.method() -> 取 "obj.method"
    name = astNode.getExpression().getText();
    break;

// 示例 3：判断是否是属性访问表达式
case TsMorph.Node.isPropertyAccessExpression(astNode):
    // 对于属性访问，取属性名
    // 例如：obj.key_a -> 取 "key_a"
    name = astNode.getName();
    break;

// 示例 4：组合判断 - 检查节点是否是类型相关
const isTypeRelatedNode = (node: Node["astNode"]) => {
    const predicators = [
        TsMorph.Node.isTypeAliasDeclaration,    // type Foo = ...
        TsMorph.Node.isInterfaceDeclaration,    // interface Bar { ... }
    ];
    const predicate = (node: Node["astNode"]) =>
        predicators.some((pred) => pred(node));
    return predicate(node) || node.getAncestors().some(predicate);
};

// 示例 5：在 findNodesWhomUse 中的多种类型判断
if (TsMorph.Node.isImportSpecifier(ancestor)) {
    // import { foo as bar } from 'module'
    // 处理 import 语句的别名
}

if (TsMorph.Node.isCallExpression(ancestor)) {
    // foo() 或 obj.method()
    // 处理函数调用表达式
}

if (TsMorph.Node.isPropertyAccessExpression(ancestor)) {
    // obj.key_a
    // 处理属性访问
}

if (TsMorph.Node.isFunctionDeclaration(ancestor)) {
    // function foo() { }
    // 处理函数声明
}

if (TsMorph.Node.isVariableDeclaration(ancestor)) {
    // const foo = ...
    // 处理变量声明
}

if (TsMorph.Node.isBinaryExpression(maybeVarDecl)) {
    // obj.key = value
    // 处理二元表达式（可能是赋值）
}

// 为什么这么用：
// 1. TypeScript AST 有几十种节点类型，需要精确判断
// 2. 不同类型的节点有不同的 API 和处理逻辑
// 3. 类型守卫函数提供 TypeScript 类型收窄，IDE 有提示
// 4. 采用命名空间组织（Node.isXxx），避免全局污染
```

**期望的 API 能力：**

* **提供类型守卫函数判断节点类型**
* **支持 TypeScript 类型收窄**
* **涵盖常用的节点类型**

**期望的使用姿势：**

```typescript
namespace Node {
    export function isIdentifier(node: Node): node is IdentifierNode;
    export function isCallExpression(node: Node): node is CallExpressionNode;
    export function isPropertyAccessExpression(node: Node): node is PropertyAccessExpressionNode;
    export function isPropertyAssignment(node: Node): node is PropertyAssignmentNode;
    export function isVariableDeclaration(node: Node): node is VariableDeclarationNode;
    export function isFunctionDeclaration(node: Node): node is FunctionDeclarationNode;
    export function isImportSpecifier(node: Node): node is ImportSpecifierNode;
    export function isObjectLiteralExpression(node: Node): node is ObjectLiteralExpressionNode;
    export function isBinaryExpression(node: Node): node is BinaryExpressionNode;
    export function isInterfaceDeclaration(node: Node): node is InterfaceDeclarationNode;
    export function isTypeAliasDeclaration(node: Node): node is TypeAliasDeclarationNode;
}
```

---

## 5. 节点信息获取

### 场景 5.1：获取节点的符号和名称

**目前如何解决：**

```typescript
const getAstNodeName = (astNode: Node["astNode"]): string => {
    // 尝试通过符号获取名称（最准确的方式）
    let name = astNode.getSymbol()?.getName();
  
    if (!name) {
        // 如果没有符号，根据节点类型提取名称
        switch (true) {
            case TsMorph.Node.isCallExpression(astNode):
                // 函数调用：取被调用的表达式
                name = astNode.getExpression().getText();
                break;
            case TsMorph.Node.isPropertyAccessExpression(astNode):
                // 属性访问：取属性名
                name = astNode.getName();
                break;
            default:
                // 兜底：取节点文本的前 20 个字符
                name = astNode.getText().slice(0, 20);
                break;
        }
    }
    return name;
};

// 使用示例：比较两个变量的符号是否相同
const targetSymbol = targetVariableDecl.getNameNode().getSymbol();
const refSymbol = refObjectExpression.getSymbol();

if (targetSymbol === refSymbol) {
    // 符号相同，说明引用的是同一个变量
    return true;
}

// 为什么这么用：
// 1. Symbol 是 TypeScript 的语义概念，比文本更准确
// 2. 同一个符号可能在不同位置有不同的文本（如重命名、别名）
// 3. 符号比较可以准确判断两个节点是否引用同一实体
// 4. getName() 获取符号的名称，这是最可靠的名称来源
```

**期望的 API 能力：**

* **获取节点的符号（Symbol）信息**
* **符号可以获取名称**
* **可能返回 undefined（某些节点没有符号）**

**期望的使用姿势：**

```typescript
interface Symbol {
    getName(): string;
}

interface Node {
    getSymbol(): Symbol | undefined;
}
```

---

### 场景 5.2：获取节点的源码文本

**目前如何解决：**

```typescript
// 示例 1：在 getAstNodeName 中获取节点文本作为兜底
default:
    // 截断过长的文本，避免日志过长
    name = astNode.getText().slice(0, 20);
    break;

// 示例 2：获取调用表达式的完整文本
case TsMorph.Node.isCallExpression(astNode):
    // 例如：foo() -> getText() 返回 "foo()"
    //      astNode.getExpression().getText() 返回 "foo"
    name = astNode.getExpression().getText();
    break;

// 示例 3：比较变量名是否相同
const targetVarName = targetVariableDecl.getName();  // 通过 API 获取
const refVarName = refObjectExpression.getText();    // 通过 getText 获取

if (targetVarName === refVarName) {
    return true;
}

// 为什么这么用：
// 1. getText() 返回节点在源码中的完整文本
// 2. 包含所有空格、注释、换行等格式
// 3. 用于日志输出、调试、简单的文本比较
// 4. 对于复杂节点，文本可能很长，需要截断处理
```

**期望的 API 能力：**

* **获取节点在源码中的完整文本**
* **保留原始格式**

**期望的使用姿势：**

```typescript
interface Node {
    getText(): string;
}
```

---

### 场景 5.3：获取节点的位置信息（用于生成唯一 ID）

**目前如何解决：**

```typescript
static generateNodeId(astNode: Node["astNode"]): string {
    // 1. 获取文件路径
    const filePath = astNode.getSourceFile().getFilePath();
    // 例如："/Users/xxx/project/src/index.ts"
  
    // 2. 获取起始行号（1-based，符合编辑器习惯）
    const line = astNode.getStartLineNumber();
    // 例如：10
  
    // 3. 计算列号（1-based）
    // getStart() 返回节点在文件中的字符偏移量（0-based）
    // getStartLinePos() 返回该行起始位置的字符偏移量
    // 两者相减得到列的偏移量，+1 转为 1-based
    const col = astNode.getStart() - astNode.getStartLinePos() + 1;
    // 例如：5
  
    // 4. 获取节点名称和类型
    const name = getAstNodeName(astNode);
    const kind = astNode.getKindName();
  
    // 5. 生成唯一 ID
    // 格式：文件路径:行号:列号 节点类型:节点名称
    return `${filePath}:${line}:${col} ${kind}:${name}`;
    // 例如："/Users/xxx/project/src/index.ts:10:5 Identifier:KEY_A"
}

// 为什么这么用：
// 1. 节点的唯一标识需要包含文件路径和位置信息
// 2. 行号+列号可以精确定位到源码中的具体位置
// 3. 便于在编辑器中跳转（很多编辑器支持 file:line:col 格式）
// 4. 便于调试和日志输出
```

**期望的 API 能力：**

* **获取节点所在的文件**
* **获取节点的起始行号（1-based）**
* **获取节点在文件中的起始位置（0-based）**
* **获取节点所在行的起始位置**

**期望的使用姿势：**

```typescript
interface Node {
    getSourceFile(): SourceFile;
    getStartLineNumber(): number;  // 1-based
    getStart(): number;            // 0-based, 在文件中的字符偏移
    getStartLinePos(): number;     // 0-based, 该行起始位置
}

interface SourceFile {
    getFilePath(): string;
}
```

---

### 场景 5.4：获取节点的语法类型

**目前如何解决：**

```typescript
// 示例 1：在日志中输出节点类型信息
this.logger.debug(
    `Found analyze entry node: ${customNode.id}, ` +
    `kind: ${node.getKind()} (${node.getKindName()})`
);
// 输出示例：
// kind: 79 (Identifier)
// kind: 206 (CallExpression)

// 示例 2：在生成节点 ID 时包含类型名称
static generateNodeId(astNode: Node["astNode"]): string {
    // ...
    const kind = astNode.getKindName();  // 获取类型的字符串名称
    return `${filePath}:${line}:${col} ${kind}:${name}`;
}

// 示例 3：比较操作符类型
if (TsMorph.Node.isBinaryExpression(maybeVarDecl) &&
    maybeVarDecl.getOperatorToken().getKind() === TsMorph.SyntaxKind.EqualsToken) {
    // 判断是否是赋值表达式（obj.key = value）
    // getKind() 返回枚举值，用于精确比较
}

// 为什么这么用：
// 1. getKind() 返回数字枚举，用于程序逻辑判断（效率高）
// 2. getKindName() 返回字符串，用于日志、调试（可读性好）
// 3. 两者配合使用：逻辑用 getKind()，输出用 getKindName()
```

**期望的 API 能力：**

* **获取节点的语法类型枚举值**
* **获取语法类型的字符串名称**

**期望的使用姿势：**

```typescript
interface Node {
    getKind(): SyntaxKind;      // 返回枚举值，用于判断
    getKindName(): string;      // 返回字符串，用于输出
}
```

---

## 6. 引用查找

### 场景：查找标识符的所有引用位置

**目前如何解决：**

```typescript
findReferences(node: Node, opts: { ignores: string[] }): Node[] {
    // ...
  
    if (TsMorph.Node.isIdentifier(targetAstNode)) {
        // 对于标识符节点，使用 findReferencesAsNodes() 查找所有引用
        const directRefAstNodes = targetAstNode.findReferencesAsNodes();
      
        // 示例场景：
        // 定义：const KEY_A = 'keyA';
        // 使用1：const obj = { key_a: KEY_A };
        // 使用2：console.log(KEY_A);
        // findReferencesAsNodes() 会返回所有使用 KEY_A 的位置
      
        // 过滤引用节点，只保留真正语义相关的引用
        const filteredRefAstNodes = directRefAstNodes.filter((refNode) => {
            return this.isSemanticReference(targetAstNode, refNode);
        });
      
        // ...
    }
}

// 为什么这么用：
// 1. TypeScript 编译器提供了完整的引用查找能力
// 2. findReferencesAsNodes() 基于符号系统，比文本搜索准确
// 3. 可以跨文件查找引用（考虑了 import/export）
// 4. 需要额外过滤，因为有些"引用"只是名字相同，语义无关
```

**期望的 API 能力：**

* **查找标识符在项目中的所有引用位置**
* **返回所有引用节点的列表**
* **基于符号系统，不是简单的文本匹配**

**期望的使用姿势：**

```typescript
interface IdentifierNode extends Node {
    findReferencesAsNodes(): Node[];
}
```

---

## 7. 特定节点类型的专有 API

### 场景 7.1：CallExpression - 获取被调用的表达式

**目前如何解决：**

```typescript
const getAstNodeName = (astNode: Node["astNode"]): string => {
    // ...
    switch (true) {
        case TsMorph.Node.isCallExpression(astNode):
            // 获取被调用的表达式部分
            name = astNode.getExpression().getText();
            break;
    }
}

// 使用场景示例：
// 1. 简单函数调用：foo() -> getExpression() 返回 foo
// 2. 方法调用：obj.method() -> getExpression() 返回 obj.method
// 3. 链式调用：obj.a().b() -> getExpression() 返回 obj.a().b

// 为什么这么用：
// 1. CallExpression 由两部分组成：表达式 + 参数列表
// 2. getExpression() 获取被调用的函数/方法部分
// 3. 用于获取调用的名称，判断调用了什么函数
```

**期望的 API 能力：**

* **获取调用表达式的被调用部分（函数名或表达式）**

**期望的使用姿势：**

```typescript
interface CallExpressionNode extends Node {
    getExpression(): Node;
}
```

---

### 场景 7.2：PropertyAccessExpression - 获取属性名和对象

**目前如何解决：**

```typescript
// 使用示例 1：获取属性名
const getAstNodeName = (astNode: Node["astNode"]): string => {
    // ...
    case TsMorph.Node.isPropertyAccessExpression(astNode):
        // 对于 obj.key_a，getName() 返回 "key_a"
        name = astNode.getName();
        break;
}

// 使用示例 2：获取被访问的对象表达式
private hasValueFlowConnection(
    targetNode: TsMorph.Node,
    refNode: TsMorph.Node
): boolean {
    const refPropertyAccess = refNode.getParent();
    if (!TsMorph.Node.isPropertyAccessExpression(refPropertyAccess)) {
        return false;
    }
  
    // 对于 obj.key_a，getExpression() 返回 obj 节点
    const refObjectExpression = refPropertyAccess.getExpression();
  
    // 检查对象是否是标识符（变量名）
    if (TsMorph.Node.isIdentifier(refObjectExpression)) {
        const refVarName = refObjectExpression.getText();
        // 比较变量名...
    }
}

// 使用场景：
// 代码：obj.key_a
// - getName() 返回 "key_a"（属性名）
// - getExpression() 返回 obj 节点（被访问的对象）

// 为什么这么用：
// 1. PropertyAccessExpression 表示对象属性访问
// 2. 需要分别获取对象部分和属性名部分
// 3. 用于分析属性访问的数据流（哪个对象的哪个属性）
```

**期望的 API 能力：**

* **获取属性访问表达式的属性名**
* **获取被访问的对象表达式**

**期望的使用姿势：**

```typescript
interface PropertyAccessExpressionNode extends Node {
    getName(): string;
    getExpression(): Node;
}
```

---

### 场景 7.3：VariableDeclaration - 获取变量名

**目前如何解决：**

```typescript
// 使用示例 1：获取变量名节点
private hasValueFlowConnection(...): boolean {
    const targetVariableDecl = targetObjectLiteral.getFirstAncestorByKind(
        TsMorph.SyntaxKind.VariableDeclaration
    );
  
    if (targetVariableDecl && TsMorph.Node.isIdentifier(refObjectExpression)) {
        // 获取变量名（字符串）
        const targetVarName = targetVariableDecl.getName();
      
        // 获取变量名节点（用于获取符号）
        const targetSymbol = targetVariableDecl.getNameNode().getSymbol();
      
        // ...
    }
}

// 使用示例 2：在 findNodesWhomUse 中
if (TsMorph.Node.isVariableDeclaration(ancestor)) {
    // 获取变量名节点
    const userAstNode = ancestor.getNameNode();
  
    if (TsMorph.Node.isIdentifier(userAstNode)) {
        // 是简单的标识符（const foo = ...）
        return { continue: false, result: userAstNode };
    } else {
        // 复杂的变量声明，如解构赋值（const { a, b } = ...）
        return { continue: false, result: ancestor };
    }
}

// 使用场景：
// const foo = 123;
// - getName() 返回 "foo"（字符串）
// - getNameNode() 返回 foo 标识符节点

// const { a, b } = obj;
// - getName() 返回 "{a, b}"（字符串）
// - getNameNode() 返回解构模式节点（不是 Identifier）

// 为什么这么用：
// 1. getName() 快速获取变量名字符串，用于比较
// 2. getNameNode() 获取节点，用于获取符号或进一步分析
// 3. 需要判断是否是简单标识符，因为解构赋值需要特殊处理
```

**期望的 API 能力：**

* **获取变量声明的名称节点**
* **获取变量名字符串**

**期望的使用姿势：**

```typescript
interface VariableDeclarationNode extends Node {
    getNameNode(): Node;      // 可能是 Identifier 或其他模式节点
    getName(): string;         // 返回名称的字符串形式
}
```

---

### 场景 7.4：FunctionDeclaration - 获取函数名

**目前如何解决：**

```typescript
// 在 findNodesWhomUse 中
if (TsMorph.Node.isFunctionDeclaration(ancestor)) {
    // 获取函数名节点
    const userAstNode = ancestor.getNameNode()!;
    return { continue: false, result: userAstNode };
}

// 使用场景：
// function foo() { ... }
// - getNameNode() 返回 foo 标识符节点

// function() { ... }  // 匿名函数
// - getNameNode() 返回 undefined

// 为什么这么用：
// 1. 函数声明可能是匿名的，需要处理 undefined 情况
// 2. 获取函数名节点，用于后续的引用查找
// 3. 这里使用 ! 断言，说明在这个上下文中确定有名称
```

**期望的 API 能力：**

* **获取函数声明的名称节点**
* **可能返回 undefined（匿名函数）**

**期望的使用姿势：**

```typescript
interface FunctionDeclarationNode extends Node {
    getNameNode(): Node | undefined;
}
```

---

### 场景 7.5：ImportSpecifier - 获取导入别名

**目前如何解决：**

```typescript
// 在 findNodesWhomUse 中
if (
    TsMorph.Node.isImportSpecifier(ancestor) &&
    ancestor.getAliasNode() != node  // 确保当前节点不是别名本身
) {
    // 返回别名节点作为使用者
    return { continue: false, result: ancestor.getAliasNode() };
}

// 使用场景：
// import { foo } from 'module';
// - getAliasNode() 返回 undefined（没有别名）

// import { foo as bar } from 'module';
// - getAliasNode() 返回 bar 标识符节点

// 分析逻辑：
// 当分析 foo 时，如果它在 import { foo as bar } 中
// 那么实际使用的是 bar，所以返回 bar 作为"使用者"

// 为什么这么用：
// 1. import 语句中可以重命名导入项
// 2. 需要区分原始名称和别名
// 3. 在代码中实际使用的是别名，所以别名才是"使用者"
```

**期望的 API 能力：**

* **获取 import 语句中的别名节点（as xxx）**

**期望的使用姿势：**

```typescript
interface ImportSpecifierNode extends Node {
    getAliasNode(): Node | undefined;
}
```

---

### 场景 7.6：BinaryExpression - 获取操作符和操作数

**目前如何解决：**

```typescript
// 使用示例 1：判断是否是赋值表达式
if (
    TsMorph.Node.isBinaryExpression(maybeVarDecl) &&
    maybeVarDecl.getOperatorToken().getKind() === TsMorph.SyntaxKind.EqualsToken
) {
    // 这是一个赋值表达式：obj.key = value
    const userAstNode = ancestor.getFirstChild((n) =>
        TsMorph.Node.isIdentifier(n)
    )!;
    return {
        continue: false,
        result: [userAstNode, maybeVarDecl],
    };
}

// 使用示例 2：在 findNodesWhomUse 的最后一个处理器中
if (isAssignmentNode(ancestor)) {
    // isAssignmentNode 内部也是检查操作符类型
    // 获取赋值的左侧节点
    const userAstNode = ancestor.getLeft();
    return { continue: false, result: userAstNode };
}

// 使用场景：
// obj.key = value
// - getOperatorToken().getKind() 返回 EqualsToken（赋值）
// - getLeft() 返回 obj.key 节点
// - getRight() 返回 value 节点

// a + b
// - getOperatorToken().getKind() 返回 PlusToken（加法）
// - getLeft() 返回 a 节点
// - getRight() 返回 b 节点

// 为什么这么用：
// 1. 二元表达式包含很多操作符（=, +, -, *, /, ==, ===, 等等）
// 2. 需要判断操作符类型来决定如何处理
// 3. 赋值操作符（=）特别重要，因为涉及到值的流动
// 4. getLeft/getRight 获取操作数，用于分析数据流
```

**期望的 API 能力：**

* **获取二元表达式的操作符类型**
* **获取左右操作数**

**期望的使用姿势：**

```typescript
interface BinaryExpressionNode extends Node {
    getOperatorToken(): OperatorToken;
    getLeft(): Node;
    getRight(): Node;
}

interface OperatorToken {
    getKind(): SyntaxKind;
}
```

---

## 8. 完整的类型系统

### 场景：支持完整的 TypeScript 语法类型枚举

**目前如何解决：**

```typescript
// 在代码中使用的 SyntaxKind 枚举值：

// 1. 字面量表达式
TsMorph.SyntaxKind.ObjectLiteralExpression  // { ... }
TsMorph.SyntaxKind.ArrayLiteralExpression   // [ ... ]

// 2. 声明语句
TsMorph.SyntaxKind.VariableDeclaration      // const/let/var foo = ...
TsMorph.SyntaxKind.FunctionDeclaration      // function foo() { ... }
TsMorph.SyntaxKind.InterfaceDeclaration     // interface Foo { ... }
TsMorph.SyntaxKind.TypeAliasDeclaration     // type Foo = ...

// 3. 表达式
TsMorph.SyntaxKind.CallExpression           // foo()
TsMorph.SyntaxKind.PropertyAccessExpression // obj.key
TsMorph.SyntaxKind.BinaryExpression         // a + b, obj.key = value

// 4. 其他
TsMorph.SyntaxKind.Identifier               // 标识符（变量名、函数名等）
TsMorph.SyntaxKind.PropertyAssignment       // 对象中的属性赋值
TsMorph.SyntaxKind.ImportSpecifier          // import 语句中的导入项
TsMorph.SyntaxKind.EqualsToken              // = 操作符

// 使用场景示例：
// 1. 在 getFirstAncestorByKind 中查找特定类型的祖先
targetNode.getFirstAncestorByKind(TsMorph.SyntaxKind.ObjectLiteralExpression)

// 2. 在 getOperatorToken().getKind() 中比较操作符类型
if (node.getOperatorToken().getKind() === TsMorph.SyntaxKind.EqualsToken) {
    // 是赋值操作
}

// 为什么这么用：
// 1. TypeScript 定义了完整的语法类型枚举
// 2. 使用枚举值进行类型判断比字符串更准确、高效
// 3. IDE 有完整的类型提示和补全
// 4. 编译期类型检查，避免拼写错误
```

**期望的 API 能力：**

* **提供完整的 SyntaxKind 枚举**
* **至少包含代码中使用到的类型**
* **支持 TypeScript 的类型检查**

**期望的使用姿势：**

```typescript
enum SyntaxKind {
    // 字面量
    ObjectLiteralExpression,
    ArrayLiteralExpression,
  
    // 声明
    VariableDeclaration,
    FunctionDeclaration,
    InterfaceDeclaration,
    TypeAliasDeclaration,
  
    // 表达式
    CallExpression,
    PropertyAccessExpression,
    BinaryExpression,
  
    // 其他
    Identifier,
    PropertyAssignment,
    ImportSpecifier,
    EqualsToken,
  
    // ... 其他常用类型
}
```

---

## 总结：核心 API 接口设计（完整版）

```typescript
// ========== 项目和文件 ==========
interface ProjectOptions {
    tsConfigFilePath?: string;
    useInMemoryFileSystem?: boolean;
    skipAddingFilesFromTsConfig?: boolean;
}

interface Project {
    constructor(options: ProjectOptions);
    getSourceFiles(): SourceFile[];
    createSourceFile(fileName: string, content: string): SourceFile;
}

interface SourceFile {
    getFilePath(): string;
    forEachDescendant(callback: (node: Node) => void): void;
}

// ========== 节点基础接口 ==========
interface Node {
    // 导航
    getParent(): Node | undefined;
    getAncestors(): Node[];
    getFirstAncestorByKind(kind: SyntaxKind): Node | undefined;
    getFirstChild(predicate: (node: Node) => boolean): Node | undefined;
  
    // 信息获取
    getSymbol(): Symbol | undefined;
    getText(): string;
    getSourceFile(): SourceFile;
    getStartLineNumber(): number;  // 1-based
    getStart(): number;            // 0-based
    getStartLinePos(): number;     // 0-based
    getKind(): SyntaxKind;
    getKindName(): string;
}

interface Symbol {
    getName(): string;
}

// ========== 节点类型判断 ==========
namespace Node {
    export function isIdentifier(node: Node): node is IdentifierNode;
    export function isCallExpression(node: Node): node is CallExpressionNode;
    export function isPropertyAccessExpression(node: Node): node is PropertyAccessExpressionNode;
    export function isPropertyAssignment(node: Node): node is PropertyAssignmentNode;
    export function isVariableDeclaration(node: Node): node is VariableDeclarationNode;
    export function isFunctionDeclaration(node: Node): node is FunctionDeclarationNode;
    export function isImportSpecifier(node: Node): node is ImportSpecifierNode;
    export function isObjectLiteralExpression(node: Node): node is ObjectLiteralExpressionNode;
    export function isBinaryExpression(node: Node): node is BinaryExpressionNode;
    export function isInterfaceDeclaration(node: Node): node is InterfaceDeclarationNode;
    export function isTypeAliasDeclaration(node: Node): node is TypeAliasDeclarationNode;
}

// ========== 特定节点类型 ==========
interface IdentifierNode extends Node {
    findReferencesAsNodes(): Node[];
}

interface CallExpressionNode extends Node {
    getExpression(): Node;
}

interface PropertyAccessExpressionNode extends Node {
    getName(): string;
    getExpression(): Node;
}

interface VariableDeclarationNode extends Node {
    getNameNode(): Node;
    getName(): string;
}

interface FunctionDeclarationNode extends Node {
    getNameNode(): Node | undefined;
}

interface ImportSpecifierNode extends Node {
    getAliasNode(): Node | undefined;
}

interface BinaryExpressionNode extends Node {
    getOperatorToken(): OperatorToken;
    getLeft(): Node;
    getRight(): Node;
}

interface OperatorToken {
    getKind(): SyntaxKind;
}

interface PropertyAssignmentNode extends Node {}
interface ObjectLiteralExpressionNode extends Node {}
interface InterfaceDeclarationNode extends Node {}
interface TypeAliasDeclarationNode extends Node {}

// ========== 语法类型枚举 ==========
enum SyntaxKind {
    ObjectLiteralExpression,
    ArrayLiteralExpression,
    VariableDeclaration,
    FunctionDeclaration,
    CallExpression,
    PropertyAccessExpression,
    PropertyAssignment,
    ImportSpecifier,
    BinaryExpression,
    EqualsToken,
    Identifier,
    InterfaceDeclaration,
    TypeAliasDeclaration,
}
```
