// package ts_bundle 实现了 TypeScript 类型依赖收集的核心逻辑。
// 主要流程为：
// 1. 以一个或多个入口文件和类型为起点。
// 2. 解析文件，查找目标类型的声明。
// 3. 如果在当前文件中未找到目标类型的声明，则分析该文件的导入（import）和重新导出（export ... from）语句。
// 4. 根据导入/导出路径，递归地进入下一个文件继续查找，直到找到目标类型的声明为止。
// 5. 找到声明后，记录其源码，并分析该类型自身依赖的其他类型，然后为这些依赖启动新一轮的递归收集。
// 6. 最终，所有找到的类型源码都被收集到 SourceCodeMap 中，等待后续处理（例如，打包成一个文件）。
package ts_bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"
)

// GlobalDeclaration 用于存储在 .d.ts 文件中找到的全局类型声明的信息。
// 全局声明通常指不通过 import/export，但在整个项目中都可用的类型。
type GlobalDeclaration struct {
	RawSource string // 类型的原始源码文本。
	FilePath  string // 类型声明所在的文件路径。
}

// CollectResult 是类型依赖收集器，它封装了收集过程中需要的所有状态和数据。
type CollectResult struct {
	RootPath                  string                          // RootPath 项目根目录的绝对路径，用于解析模块和别名。
	Alias                     map[string]string               // Alias 从 tsconfig.json 中解析出的路径别名，例如 "@/*": ["src/*"]。
	BaseUrl                   string                          // BaseUrl 从 tsconfig.json 中解析出的基础 URL，用于路径解析。
	Extensions                []string                        // Extensions 需要解析的文件扩展名列表，例如 [".ts", ".tsx", ".d.ts"]。
	SourceCodeMap             map[string]string               // SourceCodeMap 存储已收集到的类型源码，键是文件路径和类型名的组合，值是源码。
	visited                   map[string]bool                 // visited 用于在递归收集过程中检测和避免循环依赖，键是 "文件路径::类型名"。
	fileCache                 map[string]*parser.ParserResult // fileCache 缓存已解析过的文件结果（AST），避免重复解析，提高性能。
	globalDeclarationsCache   map[string]GlobalDeclaration    // globalDeclarationsCache 缓存从所有 .d.ts 文件中解析出的全局类型声明，用于快速查找。
	globalDeclarationsScanned bool                            // globalDeclarationsScanned 标记是否已执行过 .d.ts 文件的扫描。
}

// findProjectRoot 通过从入口文件开始向上遍历目录，查找项目根目录。
// 判断依据是是否存在 "tsconfig.json", "package.json" 或 ".git" 目录。
// 如果指定了 projectRootPath，则直接使用它。
func findProjectRoot(inputAnalyzeFile string, projectRootPath string) string {
	if projectRootPath != "" {
		return projectRootPath
	}
	absFilePath, err := filepath.Abs(inputAnalyzeFile)
	if err != nil {
		// 如果获取绝对路径失败，则退回到使用输入文件的目录
		dir := filepath.Dir(inputAnalyzeFile)
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return dir
		}
		return absDir
	}
	currentDir := filepath.Dir(absFilePath)
	// 最多向上查找10层，防止无限循环
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(currentDir, "tsconfig.json")); err == nil {
			return currentDir
		}
		if _, err := os.Stat(filepath.Join(currentDir, "package.json")); err == nil {
			return currentDir
		}
		if _, err := os.Stat(filepath.Join(currentDir, ".git")); err == nil {
			return currentDir
		}
		parentDir := filepath.Dir(currentDir)
		// 到达文件系统根部
		if parentDir == currentDir {
			break
		}
		currentDir = parentDir
	}
	// 如果未找到标志性文件，则使用入口文件所在的目录作为根目录
	return filepath.Dir(absFilePath)
}

// buildSourceCodeMapKey 为 SourceCodeMap 构建一个唯一的键。
// 使用空字符（\x00）作为分隔符，因为它在文件路径或类型名中几乎不可能出现。
func buildSourceCodeMapKey(filePath, typeName string) string {
	return fmt.Sprintf("%s\x00%s", filePath, typeName)
}

// parseFile 解析单个 TypeScript 文件并返回其 AST 解析结果。
// 它首先检查缓存，如果文件已被解析，则直接返回缓存结果。
func (br *CollectResult) parseFile(absFilePath string) (*parser.ParserResult, error) {
	if cached, exists := br.fileCache[absFilePath]; exists {
		return cached, nil
	}
	pr := parser.NewParserResult(absFilePath)
	if err := pr.Traverse(); err != nil {
		return nil, err
	}
	result := pr.GetResult()
	br.fileCache[absFilePath] = &result
	return &result, nil
}

// preCacheGlobalDeclarations 扫描项目中的所有 .d.ts 文件，并将其中的类型、接口、枚举声明预先缓存起来。
// 这使得在后续的依赖解析中，可以快速查找到全局可用的类型，而无需再次解析文件。
func (br *CollectResult) preCacheGlobalDeclarations() {
	// 使用项目解析器来查找所有 .d.ts 文件
	config := projectParser.NewProjectParserConfig(br.RootPath, []string{}, false, []string{".d.ts"})
	ppr := projectParser.NewProjectParserResult(config)
	ppr.ProjectParser()

	// 遍历解析结果，将找到的声明存入缓存
	for filePath, fileData := range ppr.Js_Data {
		for typeName, decl := range fileData.TypeDeclarations {
			br.globalDeclarationsCache[typeName] = GlobalDeclaration{
				RawSource: decl.Raw,
				FilePath:  filePath,
			}
		}
		for typeName, decl := range fileData.InterfaceDeclarations {
			br.globalDeclarationsCache[typeName] = GlobalDeclaration{
				RawSource: decl.Raw,
				FilePath:  filePath,
			}
		}
		for typeName, decl := range fileData.EnumDeclarations {
			br.globalDeclarationsCache[typeName] = GlobalDeclaration{
				RawSource: decl.Raw,
				FilePath:  filePath,
			}
		}
	}
}

// scanGlobalDeclarationsIfNeeded 确保全局声明只在需要时被扫描一次。
// 它采用懒加载策略，避免在初始化时进行不必要的昂贵操作。
func (br *CollectResult) scanGlobalDeclarationsIfNeeded() {
	if br.globalDeclarationsScanned {
		return
	}
	br.preCacheGlobalDeclarations()
	br.globalDeclarationsScanned = true
}

// NewCollectResult 是 CollectResult 的构造函数。
// 它初始化收集器，确定项目根目录，解析 tsconfig.json。
// 全局声明的扫描被延迟执行。
func NewCollectResult(inputAnalyzeFile string, inputAnalyzeType string, projectRootPath string) CollectResult {
	rootPath := findProjectRoot(inputAnalyzeFile, projectRootPath)
	// 解析项目配置，获取路径别名、baseUrl等信息
	config := projectParser.NewProjectParserConfig(rootPath, []string{}, false, nil)
	ar := projectParser.NewProjectParserResult(config)

	br := CollectResult{
		RootPath:                  rootPath,
		Alias:                     ar.Config.RootTsConfig.Alias,
		BaseUrl:                   ar.Config.RootTsConfig.BaseUrl,
		Extensions:                ar.Config.Extensions,
		SourceCodeMap:             make(map[string]string),
		visited:                   make(map[string]bool),
		fileCache:                 make(map[string]*parser.ParserResult),
		globalDeclarationsCache:   make(map[string]GlobalDeclaration),
		globalDeclarationsScanned: false, // 初始化时，标记为未扫描
	}

	return br
}

// handleNamespaceImport 专门处理命名空间导入（例如 `import * as allTypes from './type'`）的场景。
// 当一个类型以命名空间的形式被使用时（如 `allTypes.MyType`），此函数被调用。
func (br *CollectResult) handleNamespaceImport(absFilePath, typeName, parentTypeName string, importDecl parser.ImportDeclarationResult, module parser.ImportModule) error {
	refNameArr := strings.Split(typeName, ".")
	realRefName := refNameArr[len(refNameArr)-1] // 获取真实的类型名，例如从 `allTypes.MyType` 中获取 `MyType`

	if refNameArr[0] == module.Identifier {
		// 创建一个新的、唯一的类型名，以避免和其它同名类型冲突，例如 `allTypes_MyType`
		replaceTypeName := module.Identifier + "_" + realRefName

		// 如果父类型的源码已经被收集，需要将其对旧命名空间形式的引用替换为新的唯一类型名
		key := buildSourceCodeMapKey(absFilePath, parentTypeName)
		if _, ok := br.SourceCodeMap[key]; ok {
			realTargetTypeRaw := strings.ReplaceAll(br.SourceCodeMap[key], typeName, replaceTypeName)
			br.SourceCodeMap[key] = realTargetTypeRaw
		}

		// 解析导入的模块路径，找到对应的文件
		sourceData := projectParser.MatchImportSource(absFilePath, importDecl.Source, br.RootPath, br.Alias, br.Extensions, br.BaseUrl)
		nextFile := ""
		if sourceData.Type == "file" {
			nextFile = sourceData.FilePath
		} else {
			// 如果是 npm 包，则解析其路径
			nextFile = utils.ResolveNpmPath(absFilePath, br.RootPath, importDecl.Source, true)
			if !utils.HasExtension(nextFile, br.Extensions) {
				nextFile = utils.FindRealFilePath(nextFile, br.Extensions)
			}
		}

		// 递归地去下一个文件收集真实的类型
		return br.collectFileType(nextFile, realRefName, replaceTypeName, typeName)
	}
	return nil
}

// collectFileType 是依赖收集的核心递归函数。
// 它按照“直接声明 -> 导入/导出 -> 全局声明”的顺序在文件中查找指定类型。
// absFilePath: 当前要搜索的文件绝对路径。
// typeName: 要查找的类型名称。
// replaceTypeName: 如果找到，是否需要替换类型名（用于处理命名空间导入和别名导入）。
// parentTypeName: 引用当前类型的父类型名称，用于源码修改。
func (br *CollectResult) collectFileType(absFilePath string, typeName string, replaceTypeName string, parentTypeName string) error {
	// 使用 "文件路径::类型名" 作为键，防止在递归中重复处理同一个目标
	visitKey := fmt.Sprintf("%s::%s", absFilePath, typeName)
	if br.visited[visitKey] {
		return nil // 已在当前递归路径上访问过，说明存在循环依赖，直接返回
	}
	br.visited[visitKey] = true
	defer func() { delete(br.visited, visitKey) }() // 退出函数时，从访问记录中移除，以便其他递归路径可以访问

	parserResult, err := br.parseFile(absFilePath)
	if err != nil {
		return err
	}

	// --- 步骤 1: 在当前文件中查找类型的直接声明（type, interface, enum）---
	var rawSource string
	var references map[string]parser.TypeReference
	found := false
	if decl, ok := parserResult.TypeDeclarations[typeName]; ok {
		rawSource, references, found = decl.Raw, decl.Reference, true
	} else if decl, ok := parserResult.InterfaceDeclarations[typeName]; ok {
		rawSource, references, found = decl.Raw, decl.Reference, true
	} else if decl, ok := parserResult.EnumDeclarations[typeName]; ok {
		rawSource, references, found = decl.Raw, nil, true // 枚举类型没有引用
	}

	if found {
		finalTypeName := typeName
		// 如果需要重命名（例如 `import { MyType as YourType }`），则替换源码中的名称
		if replaceTypeName != "" {
			rawSource = strings.ReplaceAll(rawSource, typeName, replaceTypeName)
			finalTypeName = replaceTypeName
		}
		// 将找到的源码存入结果集
		br.SourceCodeMap[buildSourceCodeMapKey(absFilePath, finalTypeName)] = rawSource
		// 递归收集该类型自身依赖的其他类型
		for ref := range references {
			br.collectFileType(absFilePath, ref, "", finalTypeName)
		}
		return nil // 找到后即可返回
	}

	// --- 步骤 2: 如果没找到直接声明，则查找 import 和 re-export 语句 ---
	// 遍历所有 import 语句
	for _, importDecl := range parserResult.ImportDeclarations {
		for _, module := range importDecl.ImportModules {
			// 检查是否是 `import { typeName } from ...` 或 `import { other as typeName } from ...`
			if module.Identifier == typeName {
				realTypeName := module.ImportModule
				// 处理默认导出 `import typeName from ...`
				if module.Type == "default" {
					realTypeName = typeName
				}
				// 解析导入的模块路径
				sourceData := projectParser.MatchImportSource(absFilePath, importDecl.Source, br.RootPath, br.Alias, br.Extensions, br.BaseUrl)
				nextFile := ""
				if sourceData.Type == "file" {
					nextFile = sourceData.FilePath
				} else if sourceData.Type == "npm" {
					nextFile = utils.ResolveNpmPath(absFilePath, br.RootPath, importDecl.Source, true)
					if !utils.HasExtension(nextFile, br.Extensions) {
						nextFile = utils.FindRealFilePath(nextFile, br.Extensions)
					}
				}
				if nextFile != "" {
					// 递归到下一个文件继续查找
					return br.collectFileType(nextFile, realTypeName, typeName, typeName)
				}
			}
			// 检查是否是命名空间导入 `import * as ns from ...` 且使用 `ns.typeName`
			if module.Type == "namespace" && strings.HasPrefix(typeName, module.Identifier+".") {
				return br.handleNamespaceImport(absFilePath, typeName, parentTypeName, importDecl, module)
			}
		}
	}

	// 遍历所有 export 语句（主要处理 `export ... from ...`）
	for _, exportDecl := range parserResult.ExportDeclarations {
		if exportDecl.Source != "" { // 只处理重新导出的情况
			sourceData := projectParser.MatchImportSource(absFilePath, exportDecl.Source, br.RootPath, br.Alias, br.Extensions, br.BaseUrl)
			nextFile := ""
			if sourceData.Type == "file" {
				nextFile = sourceData.FilePath
			} else if sourceData.Type == "npm" {
				nextFile = utils.ResolveNpmPath(absFilePath, br.RootPath, exportDecl.Source, true)
				if !utils.HasExtension(nextFile, br.Extensions) {
					nextFile = utils.FindRealFilePath(nextFile, br.Extensions)
				}
			}
			if nextFile != "" {
				for _, module := range exportDecl.ExportModules {
					// 处理 `export * from ...`
					if module.Identifier == "*" {
						if err := br.collectFileType(nextFile, typeName, "", parentTypeName); err == nil {
							// 检查是否在下一个文件中成功找到了类型
							keyToCheck := buildSourceCodeMapKey(nextFile, typeName)
							if _, ok := br.SourceCodeMap[keyToCheck]; ok {
								return nil // 成功找到，结束当前路径的搜索
							}
						}
					}
					// 处理 `export { typeName } from ...` 或 `export { other as typeName } from ...`
					if module.Identifier == typeName {
						return br.collectFileType(nextFile, module.ModuleName, typeName, typeName)
					}
				}
			}
		}
	}

	// --- 步骤 3: 如果在本地和导入中都找不到，则回退到全局声明缓存中查找 ---
	// 在访问全局缓存前，按需执行一次扫描操作
	br.scanGlobalDeclarationsIfNeeded()

	if globalDecl, ok := br.globalDeclarationsCache[typeName]; ok {
		key := buildSourceCodeMapKey(globalDecl.FilePath, typeName)
		if _, alreadyCollected := br.SourceCodeMap[key]; !alreadyCollected {
			// 将全局声明的源码直接添加到结果集中
			br.SourceCodeMap[key] = globalDecl.RawSource

			// 为了找到该全局类型自身的依赖，需要即时解析它的源码
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

			// 递归收集该全局类型的依赖
			for ref := range refs {
				if err := br.collectFileType(globalDecl.FilePath, ref, "", typeName); err != nil {
					return err
				}
			}
		}
	}

	return nil // 在所有地方都未找到，正常返回
}
