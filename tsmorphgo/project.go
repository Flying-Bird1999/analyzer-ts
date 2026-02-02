package tsmorphgo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/lsp"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// Project 代表一个完整的 TypeScript 项目的视图，提供了与 ts-morph 类似的 API。
type Project struct {
	parserResult *projectParser.ProjectParserResult
	sourceFiles  map[string]*SourceFile
	lspService   *lsp.Service
	lspOnce      sync.Once
	// referenceCache 用于缓存 FindReferences 和 GotoDefinition 的结果
	// 使用简化的缓存实现
	referenceCache *ReferenceCache
	// cacheOnce 确保缓存只初始化一次
	cacheOnce sync.Once
}

// ProjectConfig 定义了初始化一个新项目所需的配置。
// API兼容性：对应ts-morph的项目配置选项
type ProjectConfig struct {
	RootPath         string
	IgnorePatterns   []string
	IsMonorepo       bool
	TargetExtensions []string
	// TypeScript 配置文件路径，如果为空则自动查找
	TsConfigPath string
	// 是否使用 tsconfig.json 中的配置覆盖其他设置
	UseTsConfig bool
	// 编译选项映射（从 tsconfig.json 解析而来）
	CompilerOptions map[string]interface{}
	// 包含的文件模式（从 tsconfig.json 解析而来）
	IncludePatterns []string
	// 排除的文件模式（从 tsconfig.json 解析而来）
	ExcludePatterns []string

	// API兼容性新增配置项
	// UseInMemoryFileSystem 是否使用内存文件系统（主要用于测试场景）
	// API兼容性：对应ts-morph的useInMemoryFileSystem选项
	UseInMemoryFileSystem bool

	// SkipAddingFilesFromTsConfig 是否跳过从tsconfig.json自动加载文件
	// API兼容性：对应ts-morph的skipAddingFilesFromTsConfig选项
	SkipAddingFilesFromTsConfig bool
}

// NewProject 是创建和初始化一个新项目实例的入口点。
// API兼容性：支持ts-morph的项目配置选项
func NewProject(config ProjectConfig) *Project {
	// 处理内存文件系统模式
	if config.UseInMemoryFileSystem {
		// 内存文件系统模式：不读取文件系统，创建空项目
		// 用户可以通过 CreateSourceFile 手动添加文件
		ppConfig := projectParser.NewProjectParserConfig(config.RootPath, nil, false, nil)
		ppResult := projectParser.NewProjectParserResult(ppConfig)
		// 不调用 ProjectParser()，避免扫描文件系统

		p := &Project{
			parserResult: ppResult,
			sourceFiles:  make(map[string]*SourceFile),
		}

		return p
	}

	// 处理 TypeScript 配置
	enhancedConfig := config
	if config.UseTsConfig && !config.SkipAddingFilesFromTsConfig {
		tsConfig := parseTsConfig(config)
		if tsConfig != nil {
			enhancedConfig = mergeTsConfig(config, tsConfig)
		}
	}

	// 构建忽略模式列表
	ignorePatterns := enhancedConfig.IgnorePatterns
	// 将 ExcludePatterns 添加到忽略模式中
	if len(enhancedConfig.ExcludePatterns) > 0 {
		ignorePatterns = append(ignorePatterns, enhancedConfig.ExcludePatterns...)
	}

	ppConfig := projectParser.NewProjectParserConfig(enhancedConfig.RootPath, ignorePatterns, enhancedConfig.IsMonorepo, enhancedConfig.TargetExtensions)
	ppResult := projectParser.NewProjectParserResult(ppConfig)
	ppResult.ProjectParser()

	p := &Project{
		parserResult: ppResult,
		sourceFiles:  make(map[string]*SourceFile),
	}

	for path, jsResult := range ppResult.Js_Data {
		// 如果有 IncludePatterns，只处理匹配的文件
		if len(enhancedConfig.IncludePatterns) > 0 && !PathMatchesPatterns(path, enhancedConfig.IncludePatterns) {
			continue
		}

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

	// 强制初始化 LSP 服务以确保类型检查器和符号绑定已准备就绪
	// p.getLspService()

	return p
}

// NewProjectFromSources 从内存中的源码 map 创建一个新项目。
func NewProjectFromSources(sources map[string]string) *Project {
	ppConfig := projectParser.NewProjectParserConfig("/", nil, false, nil)
	ppResult := projectParser.NewProjectParserResult(ppConfig)
	ppResult.ProjectParserFromMemory(sources)

	p := &Project{
		parserResult: ppResult,
		sourceFiles:  make(map[string]*SourceFile),
	}

	for path, jsResult := range ppResult.Js_Data {
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

// getLspService 返回项目唯一的 LSP 服务实例，如果需要则进行初始化。
func (p *Project) getLspService() (*lsp.Service, error) {
	var err error
	p.lspOnce.Do(func() {
		sources := make(map[string]any, len(p.parserResult.Js_Data))

		for k, v := range p.parserResult.Js_Data {
			// 关键修复：将绝对路径转换为LSP服务期望的相对路径格式
			// 从: /Users/xxx/demo-react-app/src/hooks/useUserData.ts
			// 转换为: /src/hooks/useUserData.ts
			lspPath := p.convertToLspPath(k)
			sources[lspPath] = v.Raw
		}

		// 显式传递 tsconfig.json 到 LSP 服务
		tsconfigPath := p.findProjectTsConfig()
		if tsconfigPath != "" && !p.hasTsConfigInSources(sources) {
			// 如果项目有 tsconfig.json 但 sources 中没有，显式添加
			if tsconfigContent, err := p.readTsConfigFile(tsconfigPath); err == nil {
				sources["/tsconfig.json"] = tsconfigContent
			}
		}

		p.lspService, err = lsp.NewServiceForTest(sources)
	})
	return p.lspService, err
}

// hasTsConfigInSources 检查 sources 中是否已包含 tsconfig.json
func (p *Project) hasTsConfigInSources(sources map[string]any) bool {
	for path := range sources {
		if strings.Contains(path, "tsconfig.json") {
			return true
		}
	}
	return false
}

// findProjectTsConfig 查找项目的 tsconfig.json 文件
func (p *Project) findProjectTsConfig() string {
	if p.parserResult == nil || p.parserResult.Config.RootPath == "" {
		return ""
	}

	rootPath := p.parserResult.Config.RootPath

	// 优先级列表：按常见程度排序
	tsconfigFiles := []string{
		"tsconfig.json",
		"tsconfig.base.json",
		"tsconfig.common.json",
	}

	// 首先检查项目根目录
	for _, configFile := range tsconfigFiles {
		configPath := filepath.Join(rootPath, configFile)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	return ""
}

// readTsConfigFile 读取 tsconfig.json 文件内容
func (p *Project) readTsConfigFile(configPath string) (string, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}

	// 验证是否是有效的 JSON
	var config interface{}
	if err := json.Unmarshal(content, &config); err != nil {
		return "", fmt.Errorf("JSON 格式无效: %w", err)
	}

	return string(content), nil
}

// convertToLspPath 将项目解析器的绝对路径转换为LSP服务期望的相对路径格式
// 这是连接TSMorphGo项目解析器和LSP服务的关键路径转换层
func (p *Project) convertToLspPath(projectParserPath string) string {
	// 如果路径已经是相对路径格式（以/开头但不是完整的系统路径），直接返回
	if !strings.HasPrefix(projectParserPath, "/") {
		return "/" + projectParserPath
	}

	// 检查是否为绝对路径（包含系统路径特征，如冒号或/Users）
	isAbsolutePath := strings.Contains(projectParserPath, ":") ||
		len(strings.Split(projectParserPath, "/")) > 4 // 如/Users/zxc/xxx/demo-react-app/src/file.ts

	if !isAbsolutePath {
		// 已经是期望的格式，直接返回
		return projectParserPath
	}

	// 将绝对路径转换为相对于项目根目录的路径
	if p.parserResult.Config.RootPath != "" {
		// 使用filepath.Rel计算相对路径
		relPath, err := filepath.Rel(p.parserResult.Config.RootPath, projectParserPath)
		if err == nil {
			// 确保以/开头，符合LSP服务的期望
			if !strings.HasPrefix(relPath, "/") {
				relPath = "/" + relPath
			}
			return filepath.ToSlash(relPath)
		}
	}

	// 如果无法计算相对路径，尝试从路径中提取最后的几个部分
	pathParts := strings.Split(projectParserPath, "/")
	if len(pathParts) >= 3 {
		// 假设最后几部分是 src/components/App.tsx 这样的格式
		// 查找常见的源码目录
		for i, part := range pathParts {
			if part == "src" && i < len(pathParts)-1 {
				// 找到src目录，返回从src开始的路径
				result := "/" + strings.Join(pathParts[i:], "/")
				return result
			}
		}
	}

	// 如果所有方法都失败，返回原始路径（但转换为/分隔符）
	return filepath.ToSlash(projectParserPath)
}

// getReferenceCache 获取或初始化引用缓存
// 使用单例模式确保每个项目只有一个缓存实例
// 现在使用简化的缓存实现
func (p *Project) getReferenceCache() *ReferenceCache {
	p.cacheOnce.Do(func() {
		// 创建简化缓存，最大500条目，TTL为5分钟
		p.referenceCache = NewSimpleReferenceCache(500, 5*time.Minute)
	})
	return p.referenceCache
}

// Close 关闭并释放与项目关联的资源，特别是 LSP 服务和缓存。
func (p *Project) Close() {
	if p.lspService != nil {
		p.lspService.Close()
	}
	// 清空缓存
	if p.referenceCache != nil {
		p.referenceCache.Clear()
	}
}

// GetSourceFile 根据文件路径从项目中获取一个 SourceFile 实例。
func (p *Project) GetSourceFile(path string) *SourceFile {
	normalizedPath := p.normalizeFilePath(path)
	return p.sourceFiles[normalizedPath]
}

// GetSourceFiles 返回项目中的所有源文件
func (p *Project) GetSourceFiles() []*SourceFile {
	files := make([]*SourceFile, 0, len(p.sourceFiles))
	for _, file := range p.sourceFiles {
		files = append(files, file)
	}
	return files
}

// GetParserResult 返回项目的解析结果
// 这个方法允许访问底层的 ProjectParserResult，用于需要直接访问解析数据的场景
// 例如：file_analyzer 需要访问 Js_Data 来构建依赖图
func (p *Project) GetParserResult() *projectParser.ProjectParserResult {
	return p.parserResult
}

// FindNodeAt 在指定的源文件中，根据行列号查找最精确匹配的 AST 节点。
// 返回包装后的Node类型，便于API使用。
func (p *Project) FindNodeAt(filePath string, line, char int) *Node {
	astNode := p.findNodeAt(filePath, line, char)
	if astNode == nil {
		return nil
	}

	sf, ok := p.sourceFiles[filePath]
	if !ok {
		return nil
	}

	return &Node{
		Node:       astNode,
		sourceFile: sf,
	}
}

// findNodeAt 在指定的源文件中，根据行列号查找最精确匹配的 AST 节点。
func (p *Project) findNodeAt(filePath string, line, char int) *ast.Node {
	sf, ok := p.sourceFiles[filePath]
	if !ok {
		return nil
	}

	lines := strings.Split(sf.fileResult.Raw, "\n")
	if line-1 >= len(lines) {
		return nil
	}
	offset := 0
	for i := 0; i < line-1; i++ {
		offset += len(lines[i]) + 1
	}
	offset += char - 1

	var foundNode *ast.Node
	var smallestSpan int = -1

	var walk func(node *ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}

		start, end := node.Pos(), node.End()
		if start <= offset && offset < end {
			span := end - start
			if smallestSpan == -1 || span < smallestSpan {
				smallestSpan = span
				foundNode = node
			}
			node.ForEachChild(func(child *ast.Node) bool {
				walk(child)
				return false
			})
		}
	}

	walk(sf.astNode)
	return foundNode
}

// CreateSourceFile 在项目中动态创建一个新的源文件。
// 这个方法允许在运行时向项目添加新的源文件，非常适合代码生成和动态内容创建。
// 参数:
//   - filePath: 新文件的路径，可以是相对路径或绝对路径
//   - sourceCode: 文件的源代码内容
//   - options: 可选的创建选项，如是否覆盖已存在文件等
//
// 返回值:
//   - *SourceFile: 新创建的源文件实例
//   - error: 操作过程中的错误信息
func (p *Project) CreateSourceFile(filePath string, sourceCode string, options ...CreateSourceFileOptions) (*SourceFile, error) {
	// 处理文件路径，确保使用规范化路径
	normalizedPath := p.normalizeFilePath(filePath)

	// 解析创建选项
	opts := CreateSourceFileOptions{
		Overwrite:  false,
		ScriptKind: "", // 自动检测
	}
	if len(options) > 0 {
		opts = options[0]
	}

	// 检查文件是否已存在
	if existingFile, exists := p.sourceFiles[normalizedPath]; exists {
		if !opts.Overwrite {
			return existingFile, fmt.Errorf("文件已存在: %s", normalizedPath)
		}
		// 如果允许覆盖，先移除现有文件
		delete(p.sourceFiles, normalizedPath)
		// 同时从 parserResult 中移除
		if p.parserResult.Js_Data != nil {
			delete(p.parserResult.Js_Data, normalizedPath)
		}
	}

	// 使用 parser 解析源代码
	fileParser, err := parser.NewParserFromSource(normalizedPath, sourceCode)
	if err != nil {
		return nil, fmt.Errorf("创建解析器失败: %w", err)
	}

	// 执行 AST 遍历和解析
	fileParser.Traverse()

	// 获取解析结果
	parserResult := fileParser.Result.GetResult()

	// 转换为 projectParser 的 JsFileParserResult 格式
	jsFileResult := projectParser.JsFileParserResult{
		Ast:                   fileParser.Ast,
		Raw:                   fileParser.SourceCode,
		ImportDeclarations:    make([]projectParser.ImportDeclarationResult, 0), // 暂时简化处理
		ExportDeclarations:    make([]projectParser.ExportDeclarationResult, 0), // 暂时简化处理
		ExportAssignments:     parserResult.ExportAssignments,
		InterfaceDeclarations: parserResult.InterfaceDeclarations,
		TypeDeclarations:      parserResult.TypeDeclarations,
		EnumDeclarations:      parserResult.EnumDeclarations,
		VariableDeclarations:  parserResult.VariableDeclarations,
		CallExpressions:       parserResult.CallExpressions,
		FunctionDeclarations:  parserResult.FunctionDeclarations,
		ExtractedNodes:        parserResult.ExtractedNodes,
		Errors:                fileParser.Result.Errors,
	}

	// 创建 SourceFile 实例
	sourceFile := &SourceFile{
		filePath:      normalizedPath,
		fileResult:    &jsFileResult,
		astNode:       fileParser.Ast,
		project:       p,
		nodeResultMap: make(map[*ast.Node]interface{}),
	}

	// 构建 node-result 映射
	sourceFile.buildNodeResultMap()

	// 将新文件添加到项目中
	p.sourceFiles[normalizedPath] = sourceFile

	// 同步更新 parserResult 中的数据
	if p.parserResult.Js_Data == nil {
		p.parserResult.Js_Data = make(map[string]projectParser.JsFileParserResult)
	}
	p.parserResult.Js_Data[normalizedPath] = jsFileResult

	return sourceFile, nil
}

// RemoveSourceFile 从项目中移除指定的源文件。
// 这个方法提供了动态文件管理能力，可以清理不再需要的文件。
// 参数:
//   - filePath: 要移除的文件路径
//
// 返回值:
//   - bool: 是否成功移除
//   - error: 操作过程中的错误信息
func (p *Project) RemoveSourceFile(filePath string) (bool, error) {
	normalizedPath := p.normalizeFilePath(filePath)

	// 检查文件是否存在
	if _, exists := p.sourceFiles[normalizedPath]; !exists {
		return false, fmt.Errorf("文件不存在: %s", normalizedPath)
	}

	// 从项目中移除文件
	delete(p.sourceFiles, normalizedPath)

	// 同步更新 parserResult
	delete(p.parserResult.Js_Data, normalizedPath)

	return true, nil
}

// UpdateSourceFile 更新项目中已存在的源文件内容。
// 这个方法支持动态内容更新，适用于实时编辑和代码重构场景。
// 参数:
//   - filePath: 要更新的文件路径
//   - newSourceCode: 新的源代码内容
//
// 返回值:
//   - *SourceFile: 更新后的源文件实例
//   - error: 操作过程中的错误信息
func (p *Project) UpdateSourceFile(filePath string, newSourceCode string) (*SourceFile, error) {
	normalizedPath := p.normalizeFilePath(filePath)

	// 检查文件是否存在
	if _, exists := p.sourceFiles[normalizedPath]; !exists {
		return nil, fmt.Errorf("文件不存在，无法更新: %s", normalizedPath)
	}

	// 使用覆盖选项重新创建文件
	updatedFile, err := p.CreateSourceFile(normalizedPath, newSourceCode, CreateSourceFileOptions{
		Overwrite: true,
	})
	if err != nil {
		return nil, fmt.Errorf("更新文件失败: %w", err)
	}

	// 确保返回的是同一个文件实例
	return updatedFile, nil
}

// CreateSourceFileOptions 定义了创建源文件时的可选参数。
// 这些选项提供了对文件创建过程的精细控制。
type CreateSourceFileOptions struct {
	// Overwrite 指示是否覆盖已存在的同名文件
	Overwrite bool
	// ScriptKind 指定脚本的种类（如 TypeScript、JavaScript 等）
	// 如果为空，将根据文件扩展名自动检测
	ScriptKind string
	// AdditionalOptions 其他自定义创建选项的预留字段
	AdditionalOptions map[string]interface{}
}

// normalizeFilePath 规范化文件路径，确保路径的一致性。
// 这个辅助方法处理各种路径格式，确保文件路径在项目中的统一表示。
// 参数:
//   - filePath: 输入的文件路径
//
// 返回值:
//   - string: 规范化后的文件路径
func (p *Project) normalizeFilePath(filePath string) string {
	// 如果 filePath 已经是绝对路径，直接返回
	if filepath.IsAbs(filePath) {
		return filepath.ToSlash(filePath)
	}
	// 如果是相对路径，和 RootPath 拼接
	return filepath.ToSlash(filepath.Join(p.parserResult.Config.RootPath, filePath))
}

// GetFileCount 返回项目中当前包含的源文件数量。
// 这个方法提供了项目规模的快速概览。
func (p *Project) GetFileCount() int {
	return len(p.sourceFiles)
}

// ContainsFile 检查项目中是否包含指定路径的文件。
// 这个方法用于文件存在性检查，避免重复创建或访问不存在的文件。
func (p *Project) ContainsFile(filePath string) bool {
	normalizedPath := p.normalizeFilePath(filePath)
	_, exists := p.sourceFiles[normalizedPath]
	return exists
}

// GetFilePaths 返回项目中所有源文件的路径列表。
// 这个方法提供了文件级别的项目概览，用于批量处理和统计分析。
func (p *Project) GetFilePaths() []string {
	paths := make([]string, 0, len(p.sourceFiles))
	for path := range p.sourceFiles {
		paths = append(paths, path)
	}
	return paths
}
