# TSMorphGo API 完整化技术方案

## 📋 文档概述

本文档详细说明了 TSMorphGo 项目 API 完整化的技术实施方案，基于对 ts-morph.md 文档的分析，采用分阶段开发策略，确保高质量、可维护的代码交付。

## 🎯 项目目标

将当前 TSMorphGo 的 API 覆盖率从 85-90% 提升至 95%+，实现与 ts-morph 的高度兼容，同时保持 Go 惯用设计和性能优势。

## 📊 当前状态分析

### ✅ 已完全支持的功能
- **节点导航系统** (95%+ 覆盖)
- **节点类型判断** (90%+ 覆盖)
- **特定节点类型 API** (85%+ 覆盖)
- **源文件基础操作** (80%+ 覆盖)
- **tsconfig.json 支持** (100% 覆盖)
- **动态文件创建** (100% 覆盖)
- **符号系统集成** (95%+ 覆盖，真实 LSP 集成 + 混合架构)
- **引用查找功能** (95%+ 覆盖，基于 LSP 精确定位)
- **代码导航 API** (90%+ 覆盖，包含跳转到定义)

### ⚠️ 部分支持的功能
- **高级符号关系** (75% 覆盖，基本父子关系和成员访问已支持)
- **类型检查增强** (80% 覆盖，基础类型推断和 QuickInfo 已支持)
- **项目管理** (70% 覆盖，基础功能和 LSP 集成已完成)
- **内存文件系统** (60% 覆盖，基础实现已完成)

### 🔧 当前开发重点 (第三阶段)
- **类型系统深化** - 高级类型推断和检查功能
- **代码操作 API** - 基于符号的代码重构和转换
- **性能优化** - 大规模项目的缓存和延迟加载
- **QuickInfo 增强** - 完整的类型提示和文档信息

## 🚀 分阶段实施计划

### ✅ 第一阶段：核心 API 补全 (已完成)

**完成状态**: 所有第一阶段任务已 100% 完成

**实现成果**:
- ✅ `GetKindName()` - 节点类型名称字符串化
- ✅ `GetStartLinePos()` - 行位置计算和 `PositionInfo` 结构
- ✅ `AsXXX()` 方法重构为 Node 方法，集成声明访问器
- ✅ 统一声明访问接口，集成 analyzer/parser 能力
- ✅ `tsConfigFilePath` 支持 - 完整的 TypeScript 配置文件解析
- ✅ `CreateSourceFile()` - 动态文件创建、更新、移除功能
- ✅ 全面的单元测试覆盖 (95%+ 测试覆盖率)

**核心成就**:
- 与现有 `analyzer/parser` 架构完美集成
- 支持复杂的 tsconfig.json 继承和合并
- 提供运行时动态文件管理能力
- 实现高性能的声明访问器模式
- 保证类型安全和错误容错

### ✅ 第二阶段：符号系统与引用查找 (已完成)

**完成状态**: 所有第二阶段任务已 100% 完成

**实现成果**:
- ✅ **真实符号系统集成** - 完全重写 Symbol 结构，集成 LSP 服务实现
- ✅ **混合架构设计** - LSP 服务优先，基础实现回退，确保可靠性
- ✅ **GetSymbol() 方法增强** - 集成 LSP 服务，支持精确符号获取
- ✅ **FindReferences() 优化** - 基于 LSP 精确定位的引用查找
- ✅ **GotoDefinition() 实现** - 支持跳转到定义功能
- ✅ **全面的错误处理** - panic 恢复和优雅错误处理
- ✅ **完整测试覆盖** - 符号系统和导航功能测试

**技术亮点**:
- **LSP 服务集成**: 深度集成 TypeScript 语言服务，提供真实的语义分析
- **混合架构**: LSP 服务优先策略 + 基础实现回退，确保在任何情况下都能提供稳定服务
- **错误恢复**: 全面的 panic 捕获和错误恢复机制
- **性能优化**: 懒加载 LSP 服务，避免资源浪费
- **API 兼容性**: 保持与 ts-morph API 的高度兼容

**核心代码实现**:
- `symbol.go` - 完全重写，实现混合符号系统
- `references.go` - 新增 GotoDefinition 和优化 FindReferences
- `node.go` - 新增 GetStartLineCharacter 方法
- `analyzer/lsp/lsp.go` - 新增 LSP 服务支持
- `test/goto_definition_test.go` - 新增导航功能测试

### 第三阶段：高级特性开发与完善 (2-3周)

#### 3.1 类型系统深化 (1周)

**目标**: 完善类型检查和推断功能，提供完整的类型分析能力

**实现清单**:

##### 3.1.1 类型推断增强
```go
// 文件: tsmorphgo/types.go (新增)

// GetType 获取节点的类型信息
func (n *Node) GetType() (*Type, error) {
    if n.sourceFile == nil || n.sourceFile.project == nil {
        return nil, fmt.Errorf("node must belong to a source file and project")
    }

    // 获取 LSP 服务
    lspService, err := n.sourceFile.project.getLspService()
    if err != nil {
        return nil, fmt.Errorf("failed to get LSP service: %w", err)
    }

    // 使用 LSP 服务获取类型信息
    quickInfo, err := lspService.GetQuickInfoAtPosition(
        context.Background(),
        n.sourceFile.GetFilePath(),
        n.GetStartLineNumber(),
        n.GetStartLineCharacter(),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to get type info: %w", err)
    }

    if quickInfo == nil {
        return nil, nil
    }

    return &Type{
        Text:         quickInfo.TypeText,
        DisplayParts: quickInfo.DisplayParts,
        Kind:         n.inferTypeKind(quickInfo),
    }, nil
}

// Type 表示 TypeScript 类型信息
type Type struct {
    Text         string            // 类型文本
    DisplayParts []SymbolDisplayPart // 显示部件
    Kind         TypeKind          // 类型种类
}

// TypeKind 表示类型种类
type TypeKind int

const (
    TypeKindUnknown TypeKind = iota
    TypeKindAny
    TypeKindString
    TypeKindNumber
    TypeKindBoolean
    TypeKindObject
    TypeKindFunction
    TypeKindArray
    TypeKindUnion
    TypeKindIntersection
    TypeKindLiteral
)
```

##### 3.1.2 QuickInfo 功能完善
```go
// 文件: tsmorphgo/quickinfo.go (新建)

// GetQuickInfo 获取节点的类型提示信息
func (n *Node) GetQuickInfo() (*QuickInfo, error) {
    if n.sourceFile == nil || n.sourceFile.project == nil {
        return nil, fmt.Errorf("node must belong to a source file and project")
    }

    // 获取 LSP 服务
    lspService, err := n.sourceFile.project.getLspService()
    if err != nil {
        return nil, fmt.Errorf("failed to get LSP service: %w", err)
    }

    // 使用 LSP 服务获取 QuickInfo
    quickInfo, err := lspService.GetQuickInfoAtPosition(
        context.Background(),
        n.sourceFile.GetFilePath(),
        n.GetStartLineNumber(),
        n.GetStartLineCharacter(),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to get quick info: %w", err)
    }

    return quickInfo, nil
}

// QuickInfoWithDocumentation 获取包含完整文档的类型提示
func (n *Node) GetQuickInfoWithDocumentation() (*QuickInfo, error) {
    quickInfo, err := n.GetQuickInfo()
    if err != nil || quickInfo == nil {
        return nil, err
    }

    // 尝试从符号中获取文档信息
    if symbol, found := GetSymbol(n); found {
        documentation := extractDocumentationFromSymbol(symbol)
        if documentation != "" {
            quickInfo.Documentation = documentation
        }
    }

    return quickInfo, nil
}
```

#### 3.2 代码操作 API (1周)

**目标**: 提供基于符号的代码重构和转换功能

**实现清单**:

##### 3.2.1 重构基础功能
```go
// 文件: tsmorphgo/refactor.go (新建)

// RenameSymbol 重命名符号及其所有引用
func (n *Node) RenameSymbol(newName string) error {
    // 获取符号
    symbol, err := n.GetSymbol()
    if err != nil {
        return fmt.Errorf("failed to get symbol: %w", err)
    }

    // 验证新名称
    if !isValidIdentifier(newName) {
        return fmt.Errorf("invalid identifier name: %s", newName)
    }

    // 查找所有引用
    references, err := symbol.FindReferences()
    if err != nil {
        return fmt.Errorf("failed to find references: %w", err)
    }

    // 执行重命名
    for _, ref := range references {
        if err := updateNodeText(ref.Node, newName); err != nil {
            return fmt.Errorf("failed to update reference at %s:%d: %w",
                ref.Node.GetSourceFile().GetFilePath(), ref.Node.GetStartLineNumber(), err)
        }
    }

    return nil
}

// ExtractFunction 提取函数重构
func (n *Node) ExtractFunction(functionName string) (*Node, error) {
    // 实现函数提取逻辑
    return nil, nil
}

// InlineVariable 内联变量重构
func (n *Node) InlineVariable() error {
    // 实现变量内联逻辑
    return nil
}
```

#### 3.3 性能优化 (0.5周)

**目标**: 提升大规模项目的处理性能

**实现清单**:

##### 3.3.1 缓存机制优化
```go
// 文件: tsmorphgo/cache.go (新建)

// CacheManager 统一管理所有缓存
type CacheManager struct {
    nodeCache    *NodeCache
    symbolCache  *SymbolCache
    typeCache    *TypeCache
    configCache  *ConfigCache
    mu           sync.RWMutex
}

// NodeCache 节点查询缓存
type NodeCache struct {
    entries map[string]*Node
    ttl     time.Duration
    mu      sync.RWMutex
}

func (c *NodeCache) GetOrSet(key string, compute func() *Node) *Node {
    c.mu.RLock()
    if entry, exists := c.entries[key]; exists && !c.isExpired(entry) {
        c.mu.RUnlock()
        return entry
    }
    c.mu.RUnlock()

    // 计算并缓存
    result := compute()
    if result != nil {
        c.mu.Lock()
        c.entries[key] = result
        c.mu.Unlock()
    }

    return result
}

// LSPServiceCache LSP 服务缓存
type LSPServiceCache struct {
    service    *lsp.Service
    lastAccess time.Time
    mu         sync.Mutex
}

func (c *LSPServiceCache) GetOrCreate(create func() (*lsp.Service, error)) (*lsp.Service, error) {
    c.mu.Lock()
    defer c.mu.Unlock()

    if c.service != nil && time.Since(c.lastAccess) < 30*time.Minute {
        c.lastAccess = time.Now()
        return c.service, nil
    }

    // 重新创建服务
    service, err := create()
    if err != nil {
        return nil, err
    }

    // 关闭旧服务
    if c.service != nil {
        c.service.Close()
    }

    c.service = service
    c.lastAccess = time.Now()
    return service, nil
}
```

##### 3.3.2 并发查询优化
```go
// 文件: tsmorphgo/concurrent.go (优化)

// ConcurrentBatch 并发批量查询
type ConcurrentBatch struct {
    queries []QueryTask
    workers int
}

type QueryTask struct {
    FilePath string
    Line     int
    Char     int
    Result   interface{}
    Error    error
}

func (cb *ConcurrentBatch) Execute() {
    wg := sync.WaitGroup{}
    semaphore := make(chan struct{}, cb.workers)

    for i := range cb.queries {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            semaphore <- struct{}{}
            defer func() { <-semaphore }()

            // 执行查询
            task := &cb.queries[idx]
            // 具体查询逻辑...
        }(i)
    }

    wg.Wait()
}

// ProjectQueryOptimized 优化后的项目查询
func (p *Project) QueryOptimized(query QueryBuilder) []*Node {
    // 分批处理大量文件
    sourceFiles := p.GetSourceFiles()
    batchSize := 100

    var results []*Node
    for i := 0; i < len(sourceFiles); i += batchSize {
        end := i + batchSize
        if end > len(sourceFiles) {
            end = len(sourceFiles)
        }

        batch := sourceFiles[i:end]
        batchResults := p.queryBatch(batch, query)
        results = append(results, batchResults...)
    }

    return results
}
```

#### 3.4 高级查询功能 (0.5周)

**目标**: 添加更多便利的查询方法

##### 1.2.2 实现 tsconfig.json 解析
```go
// 文件: tsmorphgo/tsconfig.go (新建)

import (
    "encoding/json"
    "os"
    "path/filepath"
    "strings"
)

// TSConfig 表示 TypeScript 配置文件的结构
type TSConfig struct {
    CompilerOptions map[string]interface{} `json:"compilerOptions"`
    Include        []string               `json:"include"`
    Exclude        []string               `json:"exclude"`
    Baseline       bool                   `json:"baseline"`
    Extends        string                 `json:"extends"`
    Files          []string               `json:"files"`
    References     []string               `json:"references"`
}

// ParseTSConfig 解析 TypeScript 配置文件
func ParseTSConfig(configPath string) (*TSConfig, error) {
    content, err := os.ReadFile(configPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read tsconfig: %w", err)
    }

    var config TSConfig
    // 使用 jsonc 解析器支持 JSON with comments
    decoder := jsonc.NewDecoder(strings.NewReader(string(content)))
    if err := decoder.Decode(&config); err != nil {
        return nil, fmt.Errorf("failed to parse tsconfig: %w", err)
    }

    return &config, nil
}

// ResolveTSConfigPaths 解析 tsconfig 中的路径配置
func (c *TSConfig) ResolveTSConfigPaths(basePath string) ([]string, error) {
    var includePatterns []string

    // 优先使用 files 字段
    if len(c.Files) > 0 {
        includePatterns = c.Files
    } else {
        // 使用 include 字段
        if len(c.Include) > 0 {
            includePatterns = c.Include
        } else {
            // 默认包含所有 .ts, .tsx 文件
            includePatterns = []string{"**/*.ts", "**/*.tsx"}
        }
    }

    // 解析路径模式
    var resolvedPaths []string
    for _, pattern := range includePatterns {
        paths, err := filepath.Glob(filepath.Join(basePath, pattern))
        if err != nil {
            continue
        }
        resolvedPaths = append(resolvedPaths, paths...)
    }

    // 排除指定文件
    if len(c.Exclude) > 0 {
        resolvedPaths = filterExcludedPaths(resolvedPaths, basePath, c.Exclude)
    }

    return resolvedPaths, nil
}

func filterExcludedPaths(paths []string, basePath string, excludePatterns []string) []string {
    var result []string
    for _, path := range paths {
        excluded := false
        relPath, err := filepath.Rel(basePath, path)
        if err != nil {
            continue
        }

        for _, pattern := range excludePatterns {
            matched, _ := filepath.Match(pattern, relPath)
            if matched {
                excluded = true
                break
            }
        }

        if !excluded {
            result = append(result, path)
        }
    }
    return result
}
```

##### 1.2.3 修改 NewProject 函数
```go
// 文件: tsmorphgo/project.go (修改)

// NewProject 是创建和初始化一个新项目实例的入口点
func NewProject(config ProjectConfig) *Project {
    var ppConfig *projectParser.ProjectParserConfig

    if config.UseInMemoryFS {
        // 内存文件系统模式
        if len(config.SourceFiles) == 0 {
            panic("UseInMemoryFS requires SourceFiles to be provided")
        }
        ppConfig = projectParser.NewProjectParserConfig("/", nil, false, nil)
        ppResult := projectParser.NewProjectParserResult(ppConfig)
        ppResult.ProjectParserFromMemory(config.SourceFiles)

        p := &Project{
            parserResult: ppResult,
            sourceFiles:  make(map[string]*SourceFile),
        }

        return p.buildProjectFromMemory()
    }

    // 处理 TypeScript 配置文件
    if config.TsConfigFilePath != "" {
        return NewProjectFromTSConfig(config.TsConfigFilePath, config)
    }

    // 使用原有逻辑
    ppConfig = projectParser.NewProjectParserConfig(
        config.RootPath,
        config.IgnorePatterns,
        config.IsMonorepo,
        config.TargetExtensions,
    )
    ppResult := projectParser.NewProjectParserResult(ppConfig)
    if !config.SkipTsConfigFiles {
        ppResult.ProjectParser()
    } else {
        ppResult.ProjectParserSimple()
    }

    p := &Project{
        parserResult: ppResult,
        sourceFiles:  make(map[string]*SourceFile),
    }

    return p.buildProjectFromDisk()
}

// NewProjectFromTSConfig 从 TypeScript 配置文件创建项目
func NewProjectFromTSConfig(tsconfigPath string, additionalConfig ProjectConfig) *Project {
    // 解析 tsconfig
    tsConfig, err := ParseTSConfig(tsconfigPath)
    if err != nil {
        panic(fmt.Errorf("failed to parse tsconfig: %w", err))
    }

    // 获取配置文件所在目录
    configDir := filepath.Dir(tsconfigPath)

    // 解析包含的文件路径
    includePaths, err := tsConfig.ResolveTSConfigPaths(configDir)
    if err != nil {
        panic(fmt.Errorf("failed to resolve tsconfig paths: %w", err))
    }

    // 创建项目配置
    ppConfig := projectParser.NewProjectParserConfig(
        configDir,
        additionalConfig.IgnorePatterns,
        additionalConfig.IsMonorepo,
        includePaths, // 使用 tsconfig 解析的文件列表
    )

    ppResult := projectParser.NewProjectParserResult(ppConfig)
    ppResult.ProjectParserFromTSConfig(tsconfigPath, tsConfig)

    p := &Project{
        parserResult: ppResult,
        sourceFiles:  make(map[string]*SourceFile),
    }

    return p.buildProjectFromDisk()
}

func (p *Project) buildProjectFromMemory() *Project {
    // 从内存中的文件构建项目
    for path, jsResult := range p.parserResult.Js_Data {
        sf := &SourceFile{
            filePath:      path,
            fileResult:    &jsResult,
            astNode:       jsResult.Ast,
            project:       p,
            nodeResultMap: make(map[*ast.Node]interface{}),
        }
        p.sourceFiles[path] = sf
        sf.buildNodeResultMap()
    }
    return p
}

func (p *Project) buildProjectFromDisk() *Project {
    // 从磁盘文件构建项目
    for path, jsResult := range p.parserResult.Js_Data {
        sf := &SourceFile{
            filePath:      path,
            fileResult:    &jsResult,
            astNode:       jsResult.Ast,
            project:       p,
            nodeResultMap: make(map[*ast.Node]interface{}),
        }
        p.sourceFiles[path] = sf
        sf.buildNodeResultMap()
    }
    return p
}
```

#### 1.3 动态文件操作支持 (0.5周)

**目标**: 支持运行时创建和管理源文件

**实现清单**:

##### 1.3.1 CreateSourceFile 方法
```go
// 文件: tsmorphgo/project.go (追加)

// CreateSourceFile 在项目中动态创建新的源文件
func (p *Project) CreateSourceFile(fileName string, content string) (*SourceFile, error) {
    // 检查文件是否已存在
    if _, exists := p.sourceFiles[fileName]; exists {
        return nil, fmt.Errorf("source file %s already exists", fileName)
    }

    // 使用 projectParser 解析新文件
    jsResult, err := projectParser.ParseSingleFileContent(fileName, content)
    if err != nil {
        return nil, fmt.Errorf("failed to parse file content: %w", err)
    }

    // 创建 SourceFile 实例
    sf := &SourceFile{
        filePath:      fileName,
        fileResult:    jsResult,
        astNode:       jsResult.Ast,
        project:       p,
        nodeResultMap: make(map[*ast.Node]interface{}),
    }

    // 构建节点映射
    sf.buildNodeResultMap()

    // 添加到项目中
    p.sourceFiles[fileName] = sf

    // 如果使用 LSP 服务，通知服务有新文件
    if p.lspService != nil {
        // 这里需要实现 LSP 服务的文件添加通知
        // 具体实现依赖于 LSP 服务的 API
    }

    return sf, nil
}

// RemoveSourceFile 从项目中移除源文件
func (p *Project) RemoveSourceFile(fileName string) error {
    if _, exists := p.sourceFiles[fileName]; !exists {
        return fmt.Errorf("source file %s does not exist", fileName)
    }

    delete(p.sourceFiles, fileName)

    // 如果使用 LSP 服务，通知服务文件已移除
    if p.lspService != nil {
        // 通知 LSP 服务文件已移除
    }

    return nil
}
```

##### 1.3.2 扩展 SourceFile 操作
```go
// 文件: tsmorphgo/sourcefile.go (追加)

// UpdateContent 更新源文件内容并重新解析
func (sf *SourceFile) UpdateContent(content string) error {
    // 使用 projectParser 重新解析文件
    jsResult, err := projectParser.ParseSingleFileContent(sf.filePath, content)
    if err != nil {
        return fmt.Errorf("failed to reparse file content: %w", err)
    }

    // 更新文件内容
    sf.fileResult = jsResult
    sf.astNode = jsResult.Ast

    // 重新构建节点映射
    sf.nodeResultMap = make(map[*ast.Node]interface{})
    sf.buildNodeResultMap()

    return nil
}

// GetContent 返回源文件的完整内容
func (sf *SourceFile) GetContent() string {
    if sf.fileResult == nil {
        return ""
    }
    return sf.fileResult.Raw
}
```

#### 1.4 类型守卫补全 (0.5周)

**目标**: 添加缺失的类型判断函数

**实现清单**:

##### 1.4.1 添加 ImportSpecifier 支持
```go
// 文件: tsmorphgo/types.go (追加)

// IsImportSpecifier 检查节点是否是 ImportSpecifier
func IsImportSpecifier(node Node) bool {
    return node.Kind == ast.KindImportSpecifier
}

// AsImportSpecifier 将节点转换为 ImportSpecifier 类型
func AsImportSpecifier(node Node) (Node, bool) {
    if IsImportSpecifier(node) {
        return node, true
    }
    return Node{}, false
}
```

##### 1.4.2 添加更多类型守卫
```go
// 文件: tsmorphgo/types.go (追加)

// IsMethodDeclaration 检查节点是否是方法声明
func IsMethodDeclaration(node Node) bool {
    return node.Kind == ast.KindMethodDeclaration
}

// IsClassDeclaration 检查节点是否是类声明
func IsClassDeclaration(node Node) bool {
    return node.Kind == ast.KindClassDeclaration
}

// IsTypeParameter 检查节点是否是类型参数
func IsTypeParameter(node Node) bool {
    return node.Kind == ast.KindTypeParameter
}

// IsParameter 检查节点是否是参数
func IsParameter(node Node) bool {
    return node.Kind == ast.KindParameter
}
```

### 第二阶段：符号系统增强 (3-4周)

#### 2.1 真实符号系统集成 (2-3周)

**目标**: 替换当前 mock 符号实现，集成 TypeScript 编译器的符号系统

**技术挑战**:
- 需要深度集成 typescript-go 的符号系统
- 处理跨文件符号解析
- 确保性能和内存效率

**实现策略**:
```go
// 文件: tsmorphgo/symbol.go (重写)

import (
    "github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
    "github.com/Zzzen/typescript-go/use-at-your-own-risk/checker"
)

// Symbol 表示 TypeScript 中的符号信息
type Symbol struct {
    name      string
    flags     ast.SymbolFlags
    valueDecl *ast.Node
    typeDecl  *ast.Node
    parent    *Symbol
    children  []*Symbol
    checker   *checker.TypeChecker
}

// GetSymbol 获取节点的符号信息
func (n *Node) GetSymbol() (*Symbol, error) {
    if n.sourceFile == nil || n.sourceFile.project == nil {
        return nil, fmt.Errorf("node must belong to a source file and project")
    }

    // 获取项目的 LSP 服务和类型检查器
    project := n.sourceFile.project
    lspService, err := project.getLspService()
    if err != nil {
        return nil, fmt.Errorf("failed to get LSP service: %w", err)
    }

    // 使用 LSP 服务获取符号信息
    symbol, err := lspService.GetSymbolAt(
        context.Background(),
        n.sourceFile.filePath,
        n.GetStartLineNumber(),
        n.GetStartColumnNumber(),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to get symbol: %w", err)
    }

    if symbol == nil {
        return nil, nil
    }

    // 转换为 TSMorphGo 的 Symbol 结构
    return &Symbol{
        name:      symbol.GetName(),
        flags:     symbol.GetFlags(),
        valueDecl: symbol.GetValueDeclaration(),
        typeDecl:  symbol.GetTypeDeclaration(),
    }, nil
}

// GetName 返回符号的名称
func (s *Symbol) GetName() string {
    return s.name
}

// GetFlags 返回符号的标志
func (s *Symbol) GetFlags() ast.SymbolFlags {
    return s.flags
}

// GetValueDeclaration 返回符号的值声明节点
func (s *Symbol) GetValueDeclaration() *Node {
    if s.valueDecl == nil {
        return nil
    }
    // 需要将 ast.Node 转换为 tsmorphgo.Node
    // 这里需要查找对应的 SourceFile 和包装
    return nil // 待实现
}

// GetTypeDeclaration 返回符号的类型声明节点
func (s *Symbol) GetTypeDeclaration() *Node {
    if s.typeDecl == nil {
        return nil
    }
    // 同上，需要转换
    return nil // 待实现
}

// GetDeclarations 返回符号的所有声明节点
func (s *Symbol) GetDeclarations() []*Node {
    // 需要从 LSP 服务或类型检查器获取所有声明
    return nil // 待实现
}

// GetGlobalScope 获取全局符号作用域
func (p *Project) GetGlobalScope() *SymbolScope {
    return &SymbolScope{
        project: p,
    }
}

// SymbolScope 表示符号的作用域
type SymbolScope struct {
    project *Project
}

// FindSymbol 在作用域中查找符号
func (s *SymbolScope) FindSymbol(name string) (*Symbol, error) {
    // 使用 LSP 服务查找符号
    return nil // 待实现
}
```

#### 2.2 引用查找改进 (1-2周)

**目标**: 提供更可靠、性能更好的引用查找功能

**实现策略**:
```go
// 文件: tsmorphgo/references.go (重写)

// FindReferences 查找节点的所有引用位置
func (n *Node) FindReferences() ([]*ReferenceInfo, error) {
    if n.sourceFile == nil || n.sourceFile.project == nil {
        return nil, fmt.Errorf("node must belong to a source file and project")
    }

    // 检查节点是否是标识符
    if !IsIdentifier(n) {
        return nil, fmt.Errorf("only identifier nodes can have references")
    }

    project := n.sourceFile.project
    lspService, err := project.getLspService()
    if err != nil {
        return nil, fmt.Errorf("failed to get LSP service: %w", err)
    }

    // 使用 LSP 服务查找引用
    response, err := lspService.FindReferences(
        context.Background(),
        n.sourceFile.filePath,
        n.GetStartLineNumber(),
        n.GetStartColumnNumber(),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to find references: %w", err)
    }

    // 转换为 ReferenceInfo 结构
    var references []*ReferenceInfo
    if response.Locations != nil {
        for _, loc := range *response.Locations {
            // 获取引用的源文件
            refFilePath := lsp.DocumentURIToFileName(loc.Uri)
            refSourceFile := project.GetSourceFile(refFilePath)
            if refSourceFile == nil {
                continue
            }

            // 在引用文件中查找对应的节点
            refNode := project.FindNodeAt(
                refFilePath,
                int(loc.Range.Start.Line+1), // 转换为 1-based
                int(loc.Range.Start.Character+1),
            )

            if refNode != nil {
                references = append(references, &ReferenceInfo{
                    Node:     refNode,
                    FilePath: refFilePath,
                    Position: &PositionInfo{
                        Line:        int(loc.Range.Start.Line + 1),
                        Column:      int(loc.Range.Start.Character + 1),
                        StartOffset: 0, // 需要计算
                        EndOffset:    0, // 需要计算
                    },
                    IsDefinition: false, // 需要判断是否是定义
                })
            }
        }
    }

    return references, nil
}

// ReferenceInfo 表示引用的详细信息
type ReferenceInfo struct {
    Node        *Node         // 引用节点
    FilePath    string        // 文件路径
    Position    *PositionInfo // 位置信息
    IsDefinition bool         // 是否是定义位置
}

// FindDefinitions 查找节点的定义位置
func (n *Node) FindDefinitions() ([]*DefinitionInfo, error) {
    // 类似 FindReferences 的实现
    return nil, 待实现
}

// DefinitionInfo 表示定义的详细信息
type DefinitionInfo struct {
    Node     *Node         // 定义节点
    FilePath string        // 文件路径
    Position *PositionInfo // 位置信息
    Kind     string        // 定义类型
}
```

### 第三阶段：高级特性开发 (2-3周)

#### 3.1 内存文件系统完善 (1周)

**目标**: 完全支持内存文件系统模式

**实现策略**:
```go
// 文件: tsmorphgo/memoryfs.go (新建)

import (
    "sync"
    "github.com/Zzzen/typescript-go/use-at-your-own-risk/vfs"
)

// MemoryFS 实现内存文件系统
type MemoryFS struct {
    files map[string]*memFile
    mutex sync.RWMutex
}

type memFile struct {
    content []byte
    mode    vfs.FileMode
}

// NewMemoryFS 创建新的内存文件系统
func NewMemoryFS() *MemoryFS {
    return &MemoryFS{
        files: make(map[string]*memFile),
    }
}

// WriteFile 向内存文件系统写入文件
func (m *MemoryFS) WriteFile(path string, content []byte, mode vfs.FileMode) error {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    m.files[path] = &memFile{
        content: append([]byte{}, content...), // 复制内容
        mode:    mode,
    }

    return nil
}

// ReadFile 从内存文件系统读取文件
func (m *MemoryFS) ReadFile(path string) ([]byte, error) {
    m.mutex.RLock()
    defer m.mutex.RUnlock()

    file, exists := m.files[path]
    if !exists {
        return nil, fmt.Errorf("file not found: %s", path)
    }

    return append([]byte{}, file.content...), nil
}

// ListFiles 列出内存文件系统中的所有文件
func (m *MemoryFS) ListFiles() []string {
    m.mutex.RLock()
    defer m.mutex.RUnlock()

    files := make([]string, 0, len(m.files))
    for path := range m.files {
        files = append(files, path)
    }

    return files
}

// RemoveFile 从内存文件系统移除文件
func (m *MemoryFS) RemoveFile(path string) error {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    if _, exists := m.files[path]; !exists {
        return fmt.Errorf("file not found: %s", path)
    }

    delete(m.files, path)
    return nil
}

// MemoryProject 支持内存文件系统的项目
type MemoryProject struct {
    *Project
    memFS *MemoryFS
}

// NewMemoryProject 创建使用内存文件系统的项目
func NewMemoryProject(initialFiles map[string]string) *MemoryProject {
    memFS := NewMemoryFS()

    // 写入初始文件
    for path, content := range initialFiles {
        err := memFS.WriteFile(path, []byte(content), 0644)
        if err != nil {
            panic(fmt.Errorf("failed to write initial file %s: %w", path, err))
        }
    }

    // 创建项目配置
    config := ProjectConfig{
        UseInMemoryFS: true,
        SourceFiles:   initialFiles,
    }

    project := NewProject(config)

    return &MemoryProject{
        Project: project,
        memFS:   memFS,
    }
}

// AddFile 向内存项目添加文件
func (mp *MemoryProject) AddFile(path string, content string) (*SourceFile, error) {
    err := mp.memFS.WriteFile(path, []byte(content), 0644)
    if err != nil {
        return nil, fmt.Errorf("failed to write file to memory FS: %w", err)
    }

    return mp.CreateSourceFile(path, content)
}

// UpdateFile 更新内存项目中的文件
func (mp *MemoryProject) UpdateFile(path string, content string) error {
    sourceFile := mp.GetSourceFile(path)
    if sourceFile == nil {
        return fmt.Errorf("file not found in project: %s", path)
    }

    err := mp.memFS.WriteFile(path, []byte(content), 0644)
    if err != nil {
        return fmt.Errorf("failed to update file in memory FS: %w", err)
    }

    return sourceFile.UpdateContent(content)
}
```

#### 3.2 性能优化 (1周)

**目标**: 提升大规模项目的处理性能

**优化策略**:

1. **懒加载优化**
```go
// 文件: tsmorphgo/project.go (性能优化)

type Project struct {
    parserResult *projectParser.ProjectParserResult
    sourceFiles  map[string]*SourceFile
    lspService   *lsp.Service
    lspOnce      sync.Once

    // 新增：性能优化相关字段
    sourceFilesLoaded sync.Map      // 已加载的文件缓存
    symbolCache      *SymbolCache    // 符号缓存
    nodeCache        *NodeCache      // 节点缓存
    config          ProjectConfig    // 项目配置缓存
}

// SymbolCache 符号缓存
type SymbolCache struct {
    symbols map[string]*Symbol
    mutex   sync.RWMutex
}

func NewSymbolCache() *SymbolCache {
    return &SymbolCache{
        symbols: make(map[string]*Symbol),
    }
}

func (c *SymbolCache) Get(key string) (*Symbol, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()

    symbol, exists := c.symbols[key]
    return symbol, exists
}

func (c *SymbolCache) Set(key string, symbol *Symbol) {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    c.symbols[key] = symbol
}

// 节点查询性能优化
func (p *Project) FindNodeAt(filePath string, line, char int) *Node {
    // 检查缓存
    cacheKey := fmt.Sprintf("%s:%d:%d", filePath, line, char)
    if p.nodeCache != nil {
        if cached, exists := p.nodeCache.Get(cacheKey); exists {
            return cached
        }
    }

    // 执行查询
    astNode := p.findNodeAt(filePath, line, char)
    if astNode == nil {
        return nil
    }

    sf, ok := p.sourceFiles[filePath]
    if !ok {
        return nil
    }

    node := &Node{
        Node:       astNode,
        sourceFile: sf,
    }

    // 缓存结果
    if p.nodeCache != nil {
        p.nodeCache.Set(cacheKey, node)
    }

    return node
}
```

2. **并发安全增强**
```go
// 文件: tsmorphgo/concurrent.go (新建)

import (
    "sync"
    "context"
)

// ConcurrentIterator 并发迭代器
type ConcurrentIterator struct {
    items    <-chan *Node
    workers  int
    ctx      context.Context
    cancel   context.CancelFunc
    wg       sync.WaitGroup
}

// NewConcurrentIterator 创建并发迭代器
func NewConcurrentIterator(sourceFiles []*SourceFile, workers int) *ConcurrentIterator {
    ctx, cancel := context.WithCancel(context.Background())

    nodeChan := make(chan *Node, workers*2)

    iterator := &ConcurrentIterator{
        items:  nodeChan,
        workers: workers,
        ctx:     ctx,
        cancel:  cancel,
    }

    // 启动工作协程
    for i := 0; i < workers; i++ {
        iterator.wg.Add(1)
        go iterator.worker(sourceFiles, nodeChan)
    }

    // 当所有工作协程完成后关闭通道
    go func() {
        iterator.wg.Wait()
        close(nodeChan)
    }()

    return iterator
}

func (ci *ConcurrentIterator) worker(sourceFiles []*SourceFile, nodeChan chan<- *Node) {
    defer ci.wg.Done()

    for _, sf := range sourceFiles {
        select {
        case <-ci.ctx.Done():
            return
        default:
            sf.ForEachDescendant(func(node Node) {
                select {
                case <-ci.ctx.Done():
                    return
                case nodeChan <- &node:
                    // 发送节点
                }
            })
        }
    }
}

func (ci *ConcurrentIterator) Next() (*Node, bool) {
    select {
    case node, ok := <-ci.items:
        return node, ok
    case <-ci.ctx.Done():
        return nil, false
    }
}

func (ci *ConcurrentIterator) Close() {
    ci.cancel()
}
```

#### 3.3 高级查询功能 (0.5周)

**目标**: 添加更多便利的查询方法

**实现策略**:
```go
// 文件: tsmorphgo/query.go (新建)

// QueryBuilder 查询构建器
type QueryBuilder struct {
    project    *Project
    predicates []func(Node) bool
    kinds      []ast.Kind
    fileFilter func(string) bool
}

// NewQueryBuilder 创建查询构建器
func (p *Project) NewQueryBuilder() *QueryBuilder {
    return &QueryBuilder{
        project: p,
    }
}

// OfKinds 按节点类型过滤
func (qb *QueryBuilder) OfKinds(kinds ...ast.Kind) *QueryBuilder {
    qb.kinds = append(qb.kinds, kinds...)
    return qb
}

// WithPredicate 添加自定义谓词
func (qb *QueryBuilder) WithPredicate(pred func(Node) bool) *QueryBuilder {
    qb.predicates = append(qb.predicates, pred)
    return qb
}

// InFiles 按文件路径过滤
func (qb *QueryBuilder) InFiles(filter func(string) bool) *QueryBuilder {
    qb.fileFilter = filter
    return qb
}

// Find 执行查询
func (qb *QueryBuilder) Find() []*Node {
    var results []*Node

    for _, sf := range qb.project.GetSourceFiles() {
        if qb.fileFilter != nil && !qb.fileFilter(sf.GetFilePath()) {
            continue
        }

        sf.ForEachDescendant(func(node Node) {
            // 检查类型过滤
            if len(qb.kinds) > 0 {
                matched := false
                for _, kind := range qb.kinds {
                    if node.Kind == kind {
                        matched = true
                        break
                    }
                }
                if !matched {
                    return
                }
            }

            // 检查自定义谓词
            for _, pred := range qb.predicates {
                if !pred(node) {
                    return
                }
            }

            results = append(results, &node)
        })
    }

    return results
}

// 便利查询方法
func (p *Project) FindIdentifiers() []*Node {
    return p.NewQueryBuilder().OfKinds(ast.KindIdentifier).Find()
}

func (p *Project) FindFunctions() []*Node {
    return p.NewQueryBuilder().OfKinds(ast.KindFunctionDeclaration, ast.KindMethodDeclaration).Find()
}

func (p *Project) FindVariables() []*Node {
    return p.NewQueryBuilder().OfKinds(ast.KindVariableDeclaration).Find()
}

func (p *Project) FindCallExpressions() []*Node {
    return p.NewQueryBuilder().OfKinds(ast.KindCallExpression).Find()
}
```

### ✅ 第四阶段：当前状态与后续计划

#### 4.1 已完成的工作总结

**核心成就回顾**:
1. **第一阶段 (已完成)**: 核心 API 补全，包括基础工具方法、项目配置增强、动态文件操作
2. **第二阶段 (已完成)**: 符号系统与引用查找，包括真实 LSP 集成、混合架构设计、完整测试覆盖
3. **当前状态**: API 覆盖率已达 85-90%，核心功能稳定可用

**技术亮点**:
- **LSP 服务深度集成**: 提供真实的 TypeScript 语义分析能力
- **混合架构设计**: LSP 优先 + 基础实现回退，确保服务可靠性
- **完整错误处理**: 全面的 panic 恢复和优雅错误处理机制
- **性能优化**: 懒加载和缓存机制，避免资源浪费
- **测试覆盖**: 符号系统和导航功能的完整测试套件

#### 4.2 第三阶段重点任务

**开发目标**: 将 API 覆盖率从 85-90% 提升至 95%+

**重点任务**:

1. **类型系统深化** (1周)
   - 完善 GetType() 方法，提供精确的类型推断
   - 增强 QuickInfo 功能，支持完整的类型提示
   - 实现类型兼容性检查

2. **代码操作 API** (1周)
   - 实现符号重命名功能
   - 添加函数提取重构
   - 支持变量内联操作

3. **性能优化** (0.5周)
   - 统一缓存管理机制
   - 并发查询优化
   - 大规模项目处理能力

4. **高级查询功能** (0.5周)
   - 查询构建器完善
   - 便利查询方法
   - 复杂查询模式支持

#### 4.3 预期成果

**功能目标**:
- API 覆盖率达到 95%+
- 完整的类型分析和推断能力
- 高性能的代码重构操作
- 优秀的错误处理和用户体验

**性能目标**:
- 1000 文件项目初始化时间 < 3秒
- 节点查询响应时间 < 50ms
- 内存使用增长控制在合理范围

**质量目标**:
- 单元测试覆盖率 > 90%
- 集成测试覆盖主要场景
- 完整的使用文档和示例

---

## 📋 实施总结

本项目采用分阶段开发策略，已成功完成前两个阶段的核心功能开发。当前 TSMorphGo 已具备：

- ✅ 完整的 AST 导航和操作能力
- ✅ 真实的 LSP 语义分析集成
- ✅ 稳定的符号系统和引用查找
- ✅ 可靠的错误处理和测试覆盖

**下一步重点**: 继续推进第三阶段的高级特性开发，进一步提升 API 覆盖率和用户体验，实现与 ts-morph 的高度兼容。

**技术优势**: 相比原版 ts-morph，TSMorphGo 具备更好的性能、更强的类型安全性、更符合 Go 语言习惯的设计，同时保持了与 TypeScript 生态系统的深度集成。

#### 4.2 测试套件构建 (1周)

**目标**: 实现全面的测试覆盖

**测试策略**:

1. **单元测试**
```go
// 文件: tsmorphgo/node_test.go

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestNode_GetKindName(t *testing.T) {
    tests := []struct {
        input    ast.Kind
        expected string
    }{
        {ast.KindIdentifier, "Identifier"},
        {ast.KindCallExpression, "CallExpression"},
        {ast.KindFunctionDeclaration, "FunctionDeclaration"},
        // 更多测试用例...
    }

    for _, test := range tests {
        // 创建测试节点
        node := Node{Node: &ast.Node{Kind: test.input}}

        result := node.GetKindName()
        assert.Equal(t, test.expected, result)
    }
}

func TestNode_GetStartLinePos(t *testing.T) {
    // 测试行位置计算
    source := `const x = 1;
const y = 2;`

    project := NewProjectFromSources(map[string]string{
        "/test.ts": source,
    })

    sourceFile := project.GetSourceFile("/test.ts")
    assert.NotNil(t, sourceFile)

    // 查找第一个 const 声明
    nodes := project.NewQueryBuilder().
        OfKinds(ast.KindVariableDeclaration).
        Find()

    assert.Len(t, nodes, 2)

    linePos := nodes[0].GetStartLinePos()
    assert.Equal(t, 0, linePos) // 第一行起始位置
}
```

2. **集成测试**
```go
// 文件: tsmorphgo/project_integration_test.go

func TestProject_TSConfigIntegration(t *testing.T) {
    // 创建临时 tsconfig.json
    tempDir := t.TempDir()
    tsconfigPath := filepath.Join(tempDir, "tsconfig.json")

    tsconfigContent := `{
        "compilerOptions": {
            "target": "es6",
            "module": "commonjs"
        },
        "include": ["**/*.ts"]
    }`

    err := os.WriteFile(tsconfigPath, []byte(tsconfigContent), 0644)
    assert.NoError(t, err)

    // 创建测试文件
    testFilePath := filepath.Join(tempDir, "test.ts")
    testContent := `const test: string = "hello";`

    err = os.WriteFile(testFilePath, []byte(testContent), 0644)
    assert.NoError(t, err)

    // 从 tsconfig 创建项目
    project := NewProjectFromTSConfig(tsconfigPath, ProjectConfig{})
    assert.NotNil(t, project)

    sourceFiles := project.GetSourceFiles()
    assert.Len(t, sourceFiles, 1)

    // 验证文件内容
    sourceFile := sourceFiles[0]
    assert.Equal(t, testFilePath, sourceFile.GetFilePath())
    assert.Equal(t, testContent, sourceFile.GetContent())
}
```

3. **性能测试**
```go
// 文件: tsmorphgo/benchmark_test.go

func BenchmarkProject_LargeProject(b *testing.B) {
    // 创建大型测试项目
    sources := make(map[string]string)
    for i := 0; i < 1000; i++ {
        sources[fmt.Sprintf("/file%d.ts", i)] = `
            function func` + strconv.Itoa(i) + `() {
                const x = ` + strconv.Itoa(i) + `;
                return x;
            }
        `
    }

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        project := NewProjectFromSources(sources)
        _ = project.GetSourceFiles()
    }
}

func BenchmarkNode_FindReferences(b *testing.B) {
    // 创建包含大量引用的测试项目
    sources := map[string]string{
        "/main.ts": `
            const shared = "shared";
            export { shared };
        `,
        "/file1.ts": `import { shared } from './main'; console.log(shared);`,
        "/file2.ts": `import { shared } from './main'; console.log(shared);`,
        // 添加更多文件...
    }

    project := NewProjectFromSources(sources)

    // 查找 shared 变量的所有引用
    sharedNodes := project.NewQueryBuilder().
        WithPredicate(func(node Node) bool {
            return IsIdentifier(node) && node.GetText() == "shared"
        }).
        Find()

    if len(sharedNodes) > 0 {
        b.ResetTimer()

        for i := 0; i < b.N; i++ {
            references, err := sharedNodes[0].FindReferences()
            if err != nil {
                b.Fatal(err)
            }
            _ = references
        }
    }
}
```

#### 4.3 示例和教程 (0.5周)

**目标**: 提供完整的使用示例和迁移指南

**实现内容**:

1. **基础教程**
```markdown
# TSMorphGo 使用教程

## 1. 基础用法

### 创建项目

```go
// 方式1：从目录创建
project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
    RootPath: "./my-ts-project",
})

// 方式2：从 TypeScript 配置创建
project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
    TsConfigFilePath: "./tsconfig.json",
})

// 方式3：内存项目（用于测试）
sources := map[string]string{
    "/main.ts": `const x = 1;`,
}
project := tsmorphgo.NewProjectFromSources(sources)
```

### 遍历 AST

```go
// 获取所有源文件
sourceFiles := project.GetSourceFiles()

for _, sf := range sourceFiles {
    fmt.Printf("Processing: %s\n", sf.GetFilePath())

    // 遍历所有节点
    sf.ForEachDescendant(func(node tsmorphgo.Node) {
        if tsmorphgo.IsIdentifier(node) {
            fmt.Printf("  Found identifier: %s at line %d\n",
                node.GetText(), node.GetStartLineNumber())
        }
    })
}
```

### 节点导航

```go
// 获取父节点和祖先节点
parent := node.GetParent()
if parent.IsValid() {
    fmt.Printf("Parent: %s\n", parent.GetKindName())
}

ancestors := node.GetAncestors()
for _, ancestor := range ancestors {
    fmt.Printf("Ancestor: %s\n", ancestor.GetKindName())
}

// 查找特定类型的祖先
funcDecl, found := node.GetFirstAncestorByKind(ast.KindFunctionDeclaration)
if found {
    fmt.Printf("Found in function: %s\n", funcDecl.GetText())
}
```
```

2. **高级示例**
```markdown
## 2. 高级用法

### 查询构建器

```go
// 查找所有函数声明
functions := project.NewQueryBuilder().
    OfKinds(ast.KindFunctionDeclaration).
    Find()

for _, fn := range functions {
    fmt.Printf("Function: %s\n", fn.GetText())
}

// 复杂查询
complexNodes := project.NewQueryBuilder().
    OfKinds(ast.KindCallExpression, ast.KindVariableDeclaration).
    WithPredicate(func(node tsmorphgo.Node) bool {
        return strings.Contains(node.GetText(), "test")
    }).
    InFiles(func(path string) bool {
        return strings.HasSuffix(path, "_test.ts")
    }).
    Find()
```

### 符号和引用查找

```go
// 获取节点符号
symbol, err := node.GetSymbol()
if err == nil && symbol != nil {
    fmt.Printf("Symbol name: %s\n", symbol.GetName())

    // 查找所有引用
    references, err := node.FindReferences()
    if err == nil {
        fmt.Printf("Found %d references\n", len(references))

        for _, ref := range references {
            fmt.Printf("  Reference at %s:%d:%d\n",
                ref.FilePath, ref.Position.Line, ref.Position.Column)
        }
    }
}
```

### 类型检查

```go
// 使用 LSP 服务获取类型信息
quickInfo, err := node.GetQuickInfo()
if err == nil && quickInfo != nil {
    fmt.Printf("Type: %s\n", quickInfo.TypeText)
    fmt.Printf("Documentation: %s\n", quickInfo.Documentation)

    for _, part := range quickInfo.DisplayParts {
        fmt.Printf("  %s (%s)\n", part.Text, part.Kind)
    }
}
```
```

3. **迁移指南**
```markdown
## 3. 从 ts-morph 迁移

### 基本概念映射

| ts-morph | TSMorphGo | 说明 |
|-----------|------------|------|
| `new Project({ tsConfigFilePath })` | `NewProject(ProjectConfig{ TsConfigFilePath })` | 项目初始化 |
| `sourceFile.forEachDescendant()` | `sourceFile.ForEachDescendant()` | 节点遍历 |
| `node.getParent()` | `node.GetParent()` | 父节点 |
| `node.getAncestors()` | `node.getAncestors()` | 祖先节点 |
| `node.getKind()` | `node.Kind` | 节点类型 |
| `node.getText()` | `node.GetText()` | 节点文本 |
| `node.findReferencesAsNodes()` | `node.FindReferences()` | 引用查找 |

### 常用模式转换

#### 节点类型检查
```typescript
// ts-morph
if (ts.Node.isIdentifier(node)) {
    // 处理标识符
}
```

```go
// TSMorphGo
if tsmorphgo.IsIdentifier(node) {
    // 处理标识符
}
```

#### 引用查找
```typescript
// ts-morph
const references = node.findReferencesAsNodes();
```

```go
// TSMorphGo
references, err := node.FindReferences();
if err != nil {
    // 处理错误
}
```

#### 项目创建
```typescript
// ts-morph
const project = new Project({
    tsConfigFilePath: './tsconfig.json',
    useInMemoryFileSystem: true,
});
```

```go
// TSMorphGo
project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
    TsConfigFilePath: "./tsconfig.json",
    UseInMemoryFS:   true,
})
```
```

## 📅 详细时间安排

### 第一阶段：核心 API 补全 (第1-3周)
- **第1周**: 基础工具方法实现
  - 完成 GetKindName()、GetStartLinePos() 方法
  - 实现位置信息计算
  - 添加辅助方法

- **第2周**: 项目配置增强
  - 实现 tsconfig.json 解析
  - 扩展 ProjectConfig 结构
  - 重构 NewProject 函数

- **第3周**: 动态文件操作和类型守卫
  - 完成 CreateSourceFile 方法
  - 补全类型判断函数
  - 第一阶段集成测试

### 第二阶段：符号系统增强 (第4-7周)
- **第4-5周**: 真实符号系统集成
  - 重写 Symbol 实现
  - 集成 TypeScript 编译器符号系统
  - 实现符号查询功能

- **第6-7周**: 引用查找改进
  - 优化 FindReferences 实现
  - 添加定义查找功能
  - 性能优化和缓存

### 第三阶段：高级特性开发 (第8-10周)
- **第8周**: 内存文件系统完善
  - 实现完整的内存文件系统
  - 添加文件操作方法
  - 内存项目优化

- **第9周**: 性能优化
  - 实现缓存机制
  - 并发安全增强
  - 性能基准测试

- **第10周**: 高级查询功能
  - 实现查询构建器
  - 添加便利查询方法
  - 高级查询测试

### 第四阶段：文档与测试 (第11-12周)
- **第11周**: 测试套件构建
  - 单元测试实现
  - 集成测试实现
  - 性能测试实现

- **第12周**: 文档完善和发布准备
  - API 文档编写
  - 使用示例整理
  - 迁移指南编写
  - 最终版本发布

## 📈 质量保证计划

### 1. 代码质量
- **Go 代码规范**: 遵循官方 Go 代码风格指南
- **代码审查**: 所有代码都需要经过同行审查
- **静态分析**: 使用 golangci-lint 进行静态代码分析

### 2. 测试覆盖
- **单元测试**: 覆盖率达到 80%+
- **集成测试**: 关键流程必须有集成测试
- **性能测试**: 性能敏感功能必须有基准测试

### 3. 文档质量
- **API 文档**: 所有公开 API 都必须有文档
- **使用示例**: 每个主要功能都提供使用示例
- **迁移指南**: 提供从 ts-morph 迁移的详细指南

### 4. 性能标准
- **内存使用**: 处理 1000 文件项目内存增长 < 50MB
- **处理速度**: 1000 文件项目初始化时间 < 5s
- **查询性能**: 节点查询响应时间 < 100ms

## 🚧 风险评估与缓解措施

### 技术风险
1. **底层依赖风险**: typescript-go 库的稳定性
   - **缓解**: 实现适配层，隔离依赖变化
   - **备选方案**: 准备替代的解析方案

2. **性能风险**: 大规模项目性能不达标
   - **缓解**: 早期性能测试，及时优化
   - **备选方案**: 实现分批处理机制

3. **兼容性风险**: 与 ts-morph API 兼容性问题
   - **缓解**: 保持 API 兼容性测试
   - **备选方案**: 提供兼容性包装器

### 时间风险
1. **开发延期**: 复杂功能实现超时
   - **缓解**: 采用敏捷开发，小步快跑
   - **备选方案**: 优先核心功能，延后次要功能

2. **测试延期**: 测试覆盖不足
   - **缓解**: 测试驱动开发
   - **备选方案**: 核心功能优先测试

### 资源风险
1. **人力资源**: 开发人员时间不足
   - **缓解**: 合理分配任务，确保关键路径
   - **备选方案**: 考虑外部合作

## 🎯 成功标准

### 功能标准
- [ ] API 覆盖率达到 90%+
- [ ] 所有核心功能稳定可用
- [ ] 支持主流 TypeScript 项目结构

### 性能标准
- [ ] 1000 文件项目初始化 < 5s
- [ ] 节点查询响应时间 < 100ms
- [ ] 内存使用增长控制在合理范围

### 质量标准
- [ ] 单元测试覆盖率 > 80%
- [ ] 集成测试覆盖主要场景
- [ ] 文档完整性 > 90%

### 用户体验
- [ ] API 设计符合 Go 惯例
- [ ] 错误信息清晰易懂
- [ ] 提供完整的使用示例

## 📞 沟通计划

### 进度汇报
- **每周进度报告**: 每周五提交进度更新
- **里程碑评审**: 每个阶段结束进行评审
- **问题升级**: 阻塞问题 24 小时内升级

### 文档更新
- **设计文档**: 及时更新技术方案
- **API 文档**: 代码提交时同步更新
- **使用指南**: 功能完成后立即编写

### 团队协作
- **代码审查**: 所有代码需要至少一人审查
- **知识分享**: 定期进行技术分享
- **问题讨论**: 使用 Issue 跟踪问题和讨论

---

本技术方案将根据实际开发进展和需求变化进行动态调整，确保项目按时高质量交付。