// package ts_bundle 批量类型收集扩展
// 支持一次处理多个入口文件和类型，通过缓存优化避免重复解析
package ts_bundle

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"
)

// TypeEntryPoint 定义类型入口点
type TypeEntryPoint struct {
	FilePath string `json:"filePath"` // 文件路径
	TypeName string `json:"typeName"` // 类型名称
	Alias    string `json:"alias"`    // 可选的类型别名，用于解决命名冲突
}

// BatchCollectResult 批量类型收集器
// 继承自 CollectResult，扩展支持多类型多文件处理
type BatchCollectResult struct {
	// 继承原有字段
	RootPath                  string
	Alias                     map[string]string
	BaseUrl                   string
	Extensions                []string
	SourceCodeMap             map[string]string
	visited                   map[string]bool
	fileCache                 map[string]*parser.ParserResult
	globalDeclarationsCache   map[string]GlobalDeclaration
	globalDeclarationsScanned bool

	// 新增字段：批量处理相关
	EntryPoints         []TypeEntryPoint  // 入口点列表
	EntryPointSourceMap map[string]string // 入口点键到源码的映射 (filePath_typeName -> sourceCode)
	processedFiles      map[string]bool   // 已处理的文件缓存 (用于批量优化)
	typeAliasMap        map[string]string // 类型别名映射 (originalType -> aliasType)
}

// NewBatchCollectResult 创建批量收集器
// 以第一个入口文件的项目根目录为准
func NewBatchCollectResult(entryPoints []TypeEntryPoint, projectRootPath string) BatchCollectResult {
	if len(entryPoints) == 0 {
		panic("entryPoints 不能为空")
	}

	// 使用第一个入口文件确定项目根目录
	rootPath := findProjectRoot(entryPoints[0].FilePath, projectRootPath)

	// 解析项目配置
	config := projectParser.NewProjectParserConfig(rootPath, []string{}, false, nil)
	ar := projectParser.NewProjectParserResult(config)

	bcr := BatchCollectResult{
		RootPath:                  rootPath,
		Alias:                     ar.Config.RootTsConfig.Alias,
		BaseUrl:                   ar.Config.RootTsConfig.BaseUrl,
		Extensions:                ar.Config.Extensions,
		SourceCodeMap:             make(map[string]string),
		visited:                   make(map[string]bool),
		fileCache:                 make(map[string]*parser.ParserResult),
		globalDeclarationsCache:   make(map[string]GlobalDeclaration),
		globalDeclarationsScanned: false,

		// 批量处理相关
		EntryPoints:         entryPoints,
		EntryPointSourceMap: make(map[string]string),
		processedFiles:      make(map[string]bool),
		typeAliasMap:        make(map[string]string),
	}

	return bcr
}

// CollectBatch 批量收集所有入口类型及其依赖
// 核心优化：同一文件只解析一次，结果复用
func (bcr *BatchCollectResult) CollectBatch() error {
	// 按文件分组入口点，优化文件解析
	fileEntryGroups := make(map[string][]TypeEntryPoint)
	for _, entry := range bcr.EntryPoints {
		absPath, err := filepath.Abs(entry.FilePath)
		if err != nil {
			return fmt.Errorf("无法解析文件路径 %s: %v", entry.FilePath, err)
		}
		entry.FilePath = absPath
		fileEntryGroups[absPath] = append(fileEntryGroups[absPath], entry)

		// 设置类型别名映射
		if entry.Alias != "" && entry.Alias != entry.TypeName {
			key := buildSourceCodeMapKey(absPath, entry.TypeName)
			bcr.typeAliasMap[key] = entry.Alias
		}
	}

	// 按文件批量处理
	for filePath, entries := range fileEntryGroups {
		if err := bcr.collectFileTypes(filePath, entries); err != nil {
			return fmt.Errorf("处理文件 %s 时出错: %v", filePath, err)
		}
	}

	return nil
}

// collectFileTypes 收集单个文件中的多个类型
// 优化点：同一文件只解析一次，处理多个类型
func (bcr *BatchCollectResult) collectFileTypes(filePath string, entries []TypeEntryPoint) error {
	// 检查文件是否已处理过（批量优化）
	if bcr.processedFiles[filePath] {
		return nil
	}

	// 解析文件（这里进行缓存优化）
	_, err := bcr.parseFile(filePath)
	if err != nil {
		return fmt.Errorf("解析文件 %s 失败: %v", filePath, err)
	}

	// 标记文件已处理
	bcr.processedFiles[filePath] = true

	// 处理该文件中的所有入口类型
	for _, entry := range entries {
		if err := bcr.collectSingleType(filePath, entry.TypeName, entry.Alias); err != nil {
			return fmt.Errorf("收集类型 %s 失败: %v", entry.TypeName, err)
		}
	}

	return nil
}

// collectSingleType 收集单个类型（复用原有逻辑）
func (bcr *BatchCollectResult) collectSingleType(absFilePath string, typeName string, alias string) error {
	// 构建最终的类型名称
	finalTypeName := typeName
	if alias != "" {
		finalTypeName = alias
	}

	// 调用原有的收集逻辑，但使用批量收集器的状态
	return bcr.collectFileType(absFilePath, typeName, finalTypeName, "")
}

// parseFile 解析文件（带缓存优化）
func (bcr *BatchCollectResult) parseFile(absFilePath string) (*parser.ParserResult, error) {
	if cached, exists := bcr.fileCache[absFilePath]; exists {
		return cached, nil
	}

	pr := parser.NewParserResult(absFilePath)
	if err := pr.Traverse(); err != nil {
		return nil, err
	}

	result := pr.GetResult()
	bcr.fileCache[absFilePath] = &result
	return &result, nil
}

// collectFileType 原有的类型收集逻辑，适配到批量收集器
// 这里基本复用原有的实现，但使用批量收集器的状态
func (bcr *BatchCollectResult) collectFileType(absFilePath string, typeName string, replaceTypeName string, parentTypeName string) error {
	// 使用 "文件路径::类型名" 作为键，防止循环依赖
	visitKey := fmt.Sprintf("%s::%s", absFilePath, typeName)
	if bcr.visited[visitKey] {
		return nil
	}
	bcr.visited[visitKey] = true
	defer func() { delete(bcr.visited, visitKey) }()

	parserResult, err := bcr.parseFile(absFilePath)
	if err != nil {
		return err
	}

	// --- 步骤 1: 在当前文件中查找类型的直接声明 ---
	var rawSource string
	var references map[string]parser.TypeReference
	found := false

	if decl, ok := parserResult.TypeDeclarations[typeName]; ok {
		rawSource, references, found = decl.Raw, decl.Reference, true
	} else if decl, ok := parserResult.InterfaceDeclarations[typeName]; ok {
		rawSource, references, found = decl.Raw, decl.Reference, true
	} else if decl, ok := parserResult.EnumDeclarations[typeName]; ok {
		rawSource, references, found = decl.Raw, nil, true
	}

	if found {
		finalTypeName := typeName
		if replaceTypeName != "" {
			rawSource = strings.ReplaceAll(rawSource, typeName, replaceTypeName)
			finalTypeName = replaceTypeName
		}

		// 将源码存储到映射中
		key := buildSourceCodeMapKey(absFilePath, finalTypeName)
		bcr.SourceCodeMap[key] = rawSource

		// 如果是入口类型，额外存储到入口点映射中
		entryKey := buildSourceCodeMapKey(absFilePath, typeName)
		if bcr.isEntryPoint(absFilePath, typeName) {
			bcr.EntryPointSourceMap[entryKey] = rawSource
		}

		// 递归收集依赖
		for ref := range references {
			bcr.collectFileType(absFilePath, ref, "", finalTypeName)
		}
		return nil
	}

	// --- 步骤 2: 查找 import 和 re-export 语句 ---
	// (这里复用原有的导入处理逻辑)
	for _, importDecl := range parserResult.ImportDeclarations {
		for _, module := range importDecl.ImportModules {
			if module.Identifier == typeName {
				realTypeName := module.ImportModule
				if module.Type == "default" {
					realTypeName = typeName
				}

				sourceData := projectParser.MatchImportSource(absFilePath, importDecl.Source, bcr.RootPath, bcr.Alias, bcr.Extensions, bcr.BaseUrl)
				nextFile := ""
				if sourceData.Type == "file" {
					nextFile = sourceData.FilePath
				} else if sourceData.Type == "npm" {
					nextFile = utils.ResolveNpmPath(absFilePath, bcr.RootPath, importDecl.Source, true)
					if !utils.HasExtension(nextFile, bcr.Extensions) {
						nextFile = utils.FindRealFilePath(nextFile, bcr.Extensions)
					}
				}

				if nextFile != "" {
					return bcr.collectFileType(nextFile, realTypeName, typeName, typeName)
				}
			}

			if module.Type == "namespace" && strings.HasPrefix(typeName, module.Identifier+".") {
				return bcr.handleNamespaceImport(absFilePath, typeName, parentTypeName, importDecl, module)
			}
		}
	}

	// --- 步骤 3: 处理 export 语句 ---
	for _, exportDecl := range parserResult.ExportDeclarations {
		if exportDecl.Source != "" {
			sourceData := projectParser.MatchImportSource(absFilePath, exportDecl.Source, bcr.RootPath, bcr.Alias, bcr.Extensions, bcr.BaseUrl)
			nextFile := ""
			if sourceData.Type == "file" {
				nextFile = sourceData.FilePath
			} else if sourceData.Type == "npm" {
				nextFile = utils.ResolveNpmPath(absFilePath, bcr.RootPath, exportDecl.Source, true)
				if !utils.HasExtension(nextFile, bcr.Extensions) {
					nextFile = utils.FindRealFilePath(nextFile, bcr.Extensions)
				}
			}

			if nextFile != "" {
				for _, module := range exportDecl.ExportModules {
					if module.Identifier == "*" {
						if err := bcr.collectFileType(nextFile, typeName, "", parentTypeName); err == nil {
							keyToCheck := buildSourceCodeMapKey(nextFile, typeName)
							if _, ok := bcr.SourceCodeMap[keyToCheck]; ok {
								return nil
							}
						}
					}
					if module.Identifier == typeName {
						return bcr.collectFileType(nextFile, module.ModuleName, typeName, typeName)
					}
				}
			}
		}
	}

	// --- 步骤 4: 查找全局声明 ---
	bcr.scanGlobalDeclarationsIfNeeded()

	if globalDecl, ok := bcr.globalDeclarationsCache[typeName]; ok {
		key := buildSourceCodeMapKey(globalDecl.FilePath, typeName)
		if _, alreadyCollected := bcr.SourceCodeMap[key]; !alreadyCollected {
			bcr.SourceCodeMap[key] = globalDecl.RawSource

			p, err := parser.NewParserFromSource(globalDecl.FilePath, globalDecl.RawSource)
			if err != nil {
				return err
			}
			p.Traverse()
			parsedGlobal := p.Result.GetResult()

			var refs map[string]parser.TypeReference
			if decl, ok := parsedGlobal.InterfaceDeclarations[typeName]; ok {
				refs = decl.Reference
			} else if decl, ok := parsedGlobal.TypeDeclarations[typeName]; ok {
				refs = decl.Reference
			}

			for ref := range refs {
				if err := bcr.collectFileType(globalDecl.FilePath, ref, "", typeName); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// handleNamespaceImport 处理命名空间导入（复用原有逻辑）
func (bcr *BatchCollectResult) handleNamespaceImport(absFilePath, typeName, parentTypeName string, importDecl parser.ImportDeclarationResult, module parser.ImportModule) error {
	refNameArr := strings.Split(typeName, ".")
	realRefName := refNameArr[len(refNameArr)-1]

	if refNameArr[0] == module.Identifier {
		replaceTypeName := module.Identifier + "_" + realRefName

		key := buildSourceCodeMapKey(absFilePath, parentTypeName)
		if _, ok := bcr.SourceCodeMap[key]; ok {
			realTargetTypeRaw := strings.ReplaceAll(bcr.SourceCodeMap[key], typeName, replaceTypeName)
			bcr.SourceCodeMap[key] = realTargetTypeRaw
		}

		sourceData := projectParser.MatchImportSource(absFilePath, importDecl.Source, bcr.RootPath, bcr.Alias, bcr.Extensions, bcr.BaseUrl)
		nextFile := ""
		if sourceData.Type == "file" {
			nextFile = sourceData.FilePath
		} else {
			nextFile = utils.ResolveNpmPath(absFilePath, bcr.RootPath, importDecl.Source, true)
			if !utils.HasExtension(nextFile, bcr.Extensions) {
				nextFile = utils.FindRealFilePath(nextFile, bcr.Extensions)
			}
		}

		return bcr.collectFileType(nextFile, realRefName, replaceTypeName, typeName)
	}
	return nil
}

// scanGlobalDeclarationsIfNeeded 扫描全局声明（复用原有逻辑）
func (bcr *BatchCollectResult) scanGlobalDeclarationsIfNeeded() {
	if bcr.globalDeclarationsScanned {
		return
	}
	bcr.preCacheGlobalDeclarations()
	bcr.globalDeclarationsScanned = true
}

// preCacheGlobalDeclarations 预缓存全局声明（复用原有逻辑）
func (bcr *BatchCollectResult) preCacheGlobalDeclarations() {
	config := projectParser.NewProjectParserConfig(bcr.RootPath, []string{}, false, []string{".d.ts"})
	ppr := projectParser.NewProjectParserResult(config)
	ppr.ProjectParser()

	for filePath, fileData := range ppr.Js_Data {
		for typeName, decl := range fileData.TypeDeclarations {
			bcr.globalDeclarationsCache[typeName] = GlobalDeclaration{
				RawSource: decl.Raw,
				FilePath:  filePath,
			}
		}
		for typeName, decl := range fileData.InterfaceDeclarations {
			bcr.globalDeclarationsCache[typeName] = GlobalDeclaration{
				RawSource: decl.Raw,
				FilePath:  filePath,
			}
		}
		for typeName, decl := range fileData.EnumDeclarations {
			bcr.globalDeclarationsCache[typeName] = GlobalDeclaration{
				RawSource: decl.Raw,
				FilePath:  filePath,
			}
		}
	}
}

// isEntryPoint 检查指定类型是否为入口类型
func (bcr *BatchCollectResult) isEntryPoint(filePath string, typeName string) bool {
	for _, entry := range bcr.EntryPoints {
		if entry.FilePath == filePath && entry.TypeName == typeName {
			return true
		}
	}
	return false
}

// GetEntryPointTypes 获取所有入口类型的源码
func (bcr *BatchCollectResult) GetEntryPointTypes() map[string]string {
	result := make(map[string]string)
	for key, source := range bcr.EntryPointSourceMap {
		result[key] = source
	}
	return result
}

// GetAllTypes 获取所有收集到的类型源码
func (bcr *BatchCollectResult) GetAllTypes() map[string]string {
	return bcr.SourceCodeMap
}
