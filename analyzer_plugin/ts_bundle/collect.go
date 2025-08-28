// 该文件实现了 ts_bundle 功能的核心逻辑：递归地收集类型依赖。
// 主要流程为：
// 1. 以一个入口文件和类型为起点。
// 2. 解析文件，查找目标类型的声明。
// 3. 如果未找到，则在该文件中查找导入（import）和重新导出（export ... from）语句。
// 4. 根据导入/导出路径，递归地到下一个文件中继续查找，直到找到目标类型的声明为止。
// 5. 找到声明后，记录其源码，并分析该类型自身依赖的其他类型，再为这些依赖启动新一轮的递归收集。
// 6. 最终，所有找到的类型源码都被收集到 SourceCodeMap 中，等待后续处理。

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

// CollectResult 是类型依赖收集器，它封装了收集过程中需要的所有状态和数据。
type CollectResult struct {
	RootPath      string                          // RootPath 项目根目录的绝对路径。
	Alias         map[string]string               // Alias tsconfig.json 中定义的路径别名，例如："@/*": ["src/*"]。
	BaseUrl       string                          // BaseUrl tsconfig.json 中定义的 baseUrl。
	Extensions    []string                        // Extensions 需要解析的文件扩展名列表，例如：[".ts", ".tsx", ".d.ts"]。
	SourceCodeMap map[string]string               // SourceCodeMap 存储已收集到的类型源码。键由文件路径和类型名组合而成，值是类型的原始代码字符串。
	visited       map[string]bool                 // visited 用于在递归过程中检测和避免循环依赖。
	fileCache     map[string]*parser.ParserResult // fileCache 缓存已解析过的文件结果，避免对同一文件进行重复的 AST 解析，提高性能。
}

// findProjectRoot 向上查找并确定项目根目录。
// 查找逻辑：优先使用用户指定的 `projectRootPath`，如果未指定，则从入口文件 `inputAnalyzeFile`
// 所在目录开始向上查找，依次寻找 `tsconfig.json`、`package.json` 或 `.git` 作为项目根目录的标志。
// 如果都找不到，则使用入口文件所在的目录作为根目录。
func findProjectRoot(inputAnalyzeFile string, projectRootPath string) string {
	// 如果显式提供了项目根路径，直接使用。
	if projectRootPath != "" {
		return projectRootPath
	}

	// 获取入口文件的绝对路径。
	absFilePath, err := filepath.Abs(inputAnalyzeFile)
	if err != nil {
		// 如果无法获取绝对路径，使用文件所在目录。
		dir := filepath.Dir(inputAnalyzeFile)
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return dir // 最坏情况下返回原始目录。
		}
		return absDir
	}

	// 从入口文件所在目录开始向上查找。
	currentDir := filepath.Dir(absFilePath)

	// 限制向上查找的深度，避免在异常文件结构中无限循环。
	for i := 0; i < 10; i++ {
		// 检查是否存在 tsconfig.json。
		if _, err := os.Stat(filepath.Join(currentDir, "tsconfig.json")); err == nil {
			return currentDir
		}

		// 检查是否存在 package.json。
		if _, err := os.Stat(filepath.Join(currentDir, "package.json")); err == nil {
			return currentDir
		}

		// 检查是否存在 .git 目录。
		if _, err := os.Stat(filepath.Join(currentDir, ".git")); err == nil {
			return currentDir
		}

		// 向上移动一级目录。
		parentDir := filepath.Dir(currentDir)

		// 如果已经到达文件系统根目录，停止查找。
		if parentDir == currentDir {
			break
		}

		currentDir = parentDir
	}

	// 如果找不到项目根目录，使用入口文件所在目录。
	return filepath.Dir(absFilePath)
}

// buildSourceCodeMapKey 为 SourceCodeMap 构建一个唯一的键。
// 使用空字符 `\x00` 作为文件路径和类型名的分隔符，因为它在文件路径或标识符中几乎不会出现，从而避免键冲突。
func buildSourceCodeMapKey(filePath, typeName string) string {
	return fmt.Sprintf("%s\x00%s", filePath, typeName)
}

// parseFile 解析单个 TS 文件并返回其 AST 解析结果。
// 为了提高性能，此函数使用了缓存（fileCache），确保每个文件只被解析一次。
func (br *CollectResult) parseFile(absFilePath string) (*parser.ParserResult, error) {
	// 检查缓存中是否已存在解析结果。
	if cached, exists := br.fileCache[absFilePath]; exists {
		return cached, nil
	}

	// 如果缓存未命中，则创建新的解析器并解析文件。
	pr := parser.NewParserResult(absFilePath)
	if err := pr.Traverse(); err != nil {
		return nil, err // 如果解析失败（例如，文件不存在），则向上传递错误。
	}
	result := pr.GetResult()

	// 将解析结果存入缓存。
	br.fileCache[absFilePath] = &result
	return &result, nil
}

// handleNamespaceImport 专门处理命名空间导入（例如 `import * as allTypes from './type'`）的场景。
func (br *CollectResult) handleNamespaceImport(absFilePath, typeName, parentTypeName string, importDecl parser.ImportDeclarationResult, module parser.ImportModule) error {
	// 当我们遇到的类型名是点状访问（如 `allTypes.MerchantData`）时，此函数被触发。
	refNameArr := strings.Split(typeName, ".")
	realRefName := refNameArr[len(refNameArr)-1] // 提取出真正的类型名，例如 `MerchantData`。

	// 检查点状访问的根对象是否与命名空间导入的标识符匹配。
	if refNameArr[0] == module.Identifier { // 例如，检查 `allTypes` 是否匹配。
		// 为了避免冲突，生成一个新的、唯一的类型名，例如 `allTypes_MerchantData`。
		replaceTypeName := module.Identifier + "_" + realRefName

		// 在父类型的源码中，将点状访问替换为新的唯一类型名。
		key := buildSourceCodeMapKey(absFilePath, parentTypeName)
		realTargetTypeRaw := strings.ReplaceAll(br.SourceCodeMap[key], typeName, replaceTypeName)
		br.SourceCodeMap[key] = realTargetTypeRaw

		// 解析命名空间导入的源文件路径。
		sourceData := projectParser.MatchImportSource(absFilePath, importDecl.Source, br.RootPath, br.Alias, br.Extensions, br.BaseUrl)
		nextFile := ""
		if sourceData.Type == "file" {
			nextFile = sourceData.FilePath
		} else {
			// 处理 npm 包导入。
			nextFile = utils.ResolveNpmPath(absFilePath, br.RootPath, importDecl.Source, true)
			// 如果解析出的路径没有文件扩展名，则尝试根据 `Extensions` 列表补全。
			if !utils.HasExtension(nextFile, br.Extensions) {
				nextFile = utils.FindRealFilePath(nextFile, br.Extensions)
			}
		}

		// 递归地到源文件中收集真正的类型定义。
		return br.collectFileType(nextFile, realRefName, replaceTypeName, typeName)
	}
	return nil
}

// NewCollectResult 是 CollectResult 的构造函数。
// 它负责初始化收集器，并根据入口文件自动推断出项目根目录、tsconfig.json 中的路径别名等配置信息。
func NewCollectResult(inputAnalyzeFile string, inputAnalyzeType string, projectRootPath string) CollectResult {
	// 通过智能方式查找项目根目录。
	rootPath := findProjectRoot(inputAnalyzeFile, projectRootPath)

	// 解析项目根目录下的 tsconfig.json 文件以获取路径别名等配置。
	config := projectParser.NewProjectParserConfig(rootPath, []string{}, false)
	ar := projectParser.NewProjectParserResult(config)

	// 返回一个初始化完毕的 CollectResult 实例。
	return CollectResult{
		RootPath:      rootPath,
		Alias:         ar.Config.RootTsConfig.Alias,
		BaseUrl:       ar.Config.RootTsConfig.BaseUrl,
		Extensions:    ar.Config.Extensions,
		SourceCodeMap: make(map[string]string),
		visited:       make(map[string]bool),
		fileCache:     make(map[string]*parser.ParserResult),
	}
}

// collectFileType 是依赖收集的核心函数，它以递归方式在文件中查找并收集指定类型的声明及其所有依赖项。
// absFilePath: 当前正在解析的文件的绝对路径。
// typeName: 当前需要查找的类型名称。
// replaceTypeName: 如果此类型是通过别名导入的（如 `import { A as B }`），此参数会持有别名（`B`），用于后续重命名。
// parentTypeName: 引用当前类型的父类型的名称，主要用于处理命名空间导入。
func (br *CollectResult) collectFileType(absFilePath string, typeName string, replaceTypeName string, parentTypeName string) error {
	// 使用文件路径和类型名创建一个唯一的访问键，用于避免循环依赖。
	visitKey := fmt.Sprintf("%s::%s", absFilePath, typeName)
	if br.visited[visitKey] {
		return nil // 如果已经访问过，则直接返回，中断递归。
	}
	br.visited[visitKey] = true
	defer func() { delete(br.visited, visitKey) }() // 函数结束时移除访问标记。

	fmt.Printf("开始解析当前文件: %s \n", absFilePath)

	// 解析文件，获取其 AST 和声明信息。
	parserResult, err := br.parseFile(absFilePath)
	if err != nil {
		return err // 如果文件解析失败（例如文件不存在），则将错误向上传递。
	}

	// --- 步骤 1: 在当前文件中查找类型的直接声明 --- 
	var foundDecl interface{}
	var isFound bool
	if decl, ok := parserResult.TypeDeclarations[typeName]; ok {
		foundDecl, isFound = decl, true
	} else if decl, ok := parserResult.InterfaceDeclarations[typeName]; ok {
		foundDecl, isFound = decl, true
	} else if decl, ok := parserResult.EnumDeclarations[typeName]; ok {
		foundDecl, isFound = decl, true
	}

	// 如果找到了直接声明...
	if isFound {
		var raw string
		var refs map[string]parser.TypeReference // 存储该类型引用的其他类型。

		// 根据找到的声明类型，提取其原始代码和引用列表。
		switch d := foundDecl.(type) {
		case parser.TypeDeclarationResult:
			raw, refs = d.Raw, d.Reference
		case parser.InterfaceDeclarationResult:
			raw, refs = d.Raw, d.Reference
		case parser.EnumDeclarationResult:
			raw, refs = d.Raw, nil // 枚举类型通常不包含对其他类型的引用。
		}

		finalTypeName := typeName
		// 如果该类型是通过别名导入的，需要将其重命名。
		if replaceTypeName != "" {
			raw = strings.ReplaceAll(raw, typeName, replaceTypeName)
			finalTypeName = replaceTypeName
		}
		// 将处理后的源码存入 SourceCodeMap。
		br.SourceCodeMap[buildSourceCodeMapKey(absFilePath, finalTypeName)] = raw
		
		// 递归地为该类型引用的所有其他类型启动新一轮的依赖收集。
		for ref := range refs {
			if err := br.collectFileType(absFilePath, ref, "", finalTypeName); err != nil {
				return err
			}
		}
		return nil // 成功找到并处理完一个类型，结束当前递归分支。
	}

	// --- 步骤 2: 如果没有直接声明，则查找 import 语句，尝试从其他文件追踪类型 --- 
	for _, importDecl := range parserResult.ImportDeclarations {
		for _, module := range importDecl.ImportModules {
			// 如果导入的标识符与我们正在寻找的类型名匹配...
			if module.Identifier == typeName {
				realTypeName := module.ImportModule
				// 对于默认导出 `import T from './t'`，其真实模块名是 `default`，但我们应继续寻找 `T`。
				if module.Type == "default" {
					realTypeName = typeName
				}
				
				// 解析导入语句的源路径，将其转换为绝对文件路径。
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
					// 递归到下一个文件，继续查找。
					if err := br.collectFileType(nextFile, realTypeName, typeName, typeName); err != nil {
						return err
					}
					// 如果在下一个文件中成功找到了类型，它会被加入 SourceCodeMap，此时可以提前结束当前文件的查找。
					if _, ok := br.SourceCodeMap[buildSourceCodeMapKey(nextFile, typeName)]; ok {
						return nil
					}
				}
			}
			// 处理命名空间导入。
			if module.Type == "namespace" {
				if err := br.handleNamespaceImport(absFilePath, typeName, parentTypeName, importDecl, module); err != nil {
					return err
				}
			}
		}
	}

	// --- 步骤 3: 如果仍未找到，则处理重新导出（export ... from ...）语句 --- 
	for _, exportDecl := range parserResult.ExportDeclarations {
		if exportDecl.Source == "" {
			continue // 这不是一个重新导出语句，跳过。
		}
		// 解析重新导出的源文件路径。
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
				// Case A: 通配符重新导出 `export * from './source'`
				if module.Identifier == "*" {
					if err := br.collectFileType(nextFile, typeName, "", parentTypeName); err != nil {
						return err
					}
					// 如果在通配符导出的模块中找到了，就返回。
					if _, ok := br.SourceCodeMap[buildSourceCodeMapKey(nextFile, typeName)]; ok {
						return nil
					}
				}
				// Case B: 命名重新导出 `export { A as B } from './source'`
				if module.Identifier == typeName {
					if err := br.collectFileType(nextFile, module.ModuleName, typeName, typeName); err != nil {
						return err
					}
					// 如果在命名导出的模块中找到了，就返回。
					if _, ok := br.SourceCodeMap[buildSourceCodeMapKey(nextFile, module.ModuleName)]; ok {
						return nil
					}
				}
			}
		}
	}

	return nil
}