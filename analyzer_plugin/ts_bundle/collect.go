package ts_bundle

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"
)

// UniqueTypeID 定义了一个全局唯一的类型标识符，通常由“文件路径:类型名”构成。
// 这样可以区分不同文件中定义的同名类型。
type UniqueTypeID string

// FileScopeResolutionMap 存储了单个文件作用域内所有标识符到其全局唯一ID的映射。
// 例如，`{ "MyType": "path/to/file:MyType", "AnotherType": "path/to/another/file:AnotherType" }`
// 这包括了本地声明的类型、导入的类型（以及它们的别名）等。
type FileScopeResolutionMap map[string]UniqueTypeID

// CollectResult 是依赖收集过程的最终产出。
// 它包含了整个项目（从入口文件开始）中所有相关的类型声明和它们的引用关系。
type CollectResult struct {
	// Declarations 存储所有收集到的类型声明。
	// 使用 UniqueTypeID 作为键，确保每个类型只存储一次，避免了重复。
	Declarations map[UniqueTypeID]GenericDeclaration
	// ResolutionMaps 存储每个文件（以绝对路径为键）的 FileScopeResolutionMap。
	// 这使得在处理每个文件时，可以快速查找其内部任何标识符对应的全局唯一类型。
	ResolutionMaps map[string]FileScopeResolutionMap
	// ProjectConfig 包含了项目级别的配置信息，如 tsconfig.json 中的路径别名（alias）等，
	// 这对于正确解析模块路径至关重要。
	ProjectConfig *projectParser.ProjectParserConfig
}

// NewCollectResult 初始化一个新的 CollectResult 实例。
// 它需要一个项目根路径来设置项目解析器的配置。
func NewCollectResult(projectRootPath string) *CollectResult {
	config := projectParser.NewProjectParserConfig(projectRootPath, []string{}, false)
	return &CollectResult{
		Declarations:   make(map[UniqueTypeID]GenericDeclaration),
		ResolutionMaps: make(map[string]FileScopeResolutionMap),
		ProjectConfig:  &config,
	}
}

// CollectDependencies 是依赖收集过程的入口函数。
// 它接收一个入口文件路径，并从该文件开始递归地解析所有依赖。
func (cr *CollectResult) CollectDependencies(entryFile string) error {
	absEntryFile, err := filepath.Abs(entryFile)
	if err != nil {
		return fmt.Errorf("无法获取入口文件的绝对路径: %w", err)
	}
	_, err = cr.resolveFile(absEntryFile)
	return err
}

// convertReferences 是一个辅助函数，用于将分析器（parser）生成的复杂引用映射
// 转换为一个简单的字符串集合（map[string]bool），以便于后续处理。
func convertReferences(refs map[string]parser.TypeReference) map[string]bool {
	newRefs := make(map[string]bool)
	for key := range refs {
		newRefs[key] = true
	}
	return newRefs
}

// resolveFile 是核心的递归函数，负责解析单个文件并处理其所有依赖。
// 它会缓存已解析过的文件，避免重复工作。
func (cr *CollectResult) resolveFile(absFilePath string) (FileScopeResolutionMap, error) {
	// 如果文件已经解析过，直接返回缓存的结果。
	if resMap, ok := cr.ResolutionMaps[absFilePath]; ok {
		return resMap, nil
	}

	// 初始化当前文件的解析图。
	cr.ResolutionMaps[absFilePath] = make(FileScopeResolutionMap)

	// 使用 AST 分析器解析文件内容。
	pr := parser.NewParserResult(absFilePath)
	pr.Traverse()
	parserResult := pr.GetResult()

	// 步骤 1: 处理并存储当前文件中定义的所有类型声明（type, interface, enum）。
	// 为每个声明创建一个全局唯一的ID，并将其添加到 Declarations 和当前文件的 ResolutionMap 中。
	for _, decl := range parserResult.TypeDeclarations {
		uid := UniqueTypeID(fmt.Sprintf("%s:%s", absFilePath, decl.Identifier))
		cr.Declarations[uid] = GenericDeclaration{Name: decl.Identifier, Raw: decl.Raw, Reference: convertReferences(decl.Reference), FilePath: absFilePath}
		cr.ResolutionMaps[absFilePath][decl.Identifier] = uid
	}
	for _, decl := range parserResult.InterfaceDeclarations {
		uid := UniqueTypeID(fmt.Sprintf("%s:%s", absFilePath, decl.Identifier))
		cr.Declarations[uid] = GenericDeclaration{Name: decl.Identifier, Raw: decl.Raw, Reference: convertReferences(decl.Reference), FilePath: absFilePath}
		cr.ResolutionMaps[absFilePath][decl.Identifier] = uid
	}
	for _, decl := range parserResult.EnumDeclarations {
		uid := UniqueTypeID(fmt.Sprintf("%s:%s", absFilePath, decl.Identifier))
		cr.Declarations[uid] = GenericDeclaration{Name: decl.Identifier, Raw: decl.Raw, Reference: make(map[string]bool), FilePath: absFilePath}
		cr.ResolutionMaps[absFilePath][decl.Identifier] = uid
	}

	// 步骤 2: 处理 import 声明，递归解析依赖文件。
	for _, importDecl := range parserResult.ImportDeclarations {
		nextFile := cr.resolveModulePath(absFilePath, importDecl.Source)
		if nextFile == "" {
			continue // 如果无法解析模块路径，则跳过。
		}
		// 递归调用 resolveFile 来解析依赖文件。
		depResMap, err := cr.resolveFile(nextFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "警告: 无法解析文件 '%s' 中的依赖 '%s': %v", absFilePath, importDecl.Source, err)
			continue
		}

		// 将导入的模块添加到当前文件的作用域解析图中。
		for _, module := range importDecl.ImportModules {
			// 处理命名空间导入 (e.g., `import * as ns from ...`)
			if module.Type == "namespace" {
				// 遍历被导入模块的所有导出，并以 "命名空间.成员" 的形式添加到当前文件的解析图中。
				for name, uid := range depResMap {
					cr.ResolutionMaps[absFilePath][fmt.Sprintf("%s.%s", module.Identifier, name)] = uid
				}
			} else if depUID, ok := depResMap[module.ImportModule]; ok {
				// 处理常规的命名导入或默认导入，将本地标识符映射到其全局唯一ID。
				cr.ResolutionMaps[absFilePath][module.Identifier] = depUID
			}
		}
	}

	// 步骤 3: 处理默认导出 (export default ...)。
	// 将 "default" 关键字映射到被导出的表达式对应的全局唯一ID。
	for _, exportAssign := range parserResult.ExportAssignments {
		if exportAssign.Expression != "" {
			if uid, ok := cr.ResolutionMaps[absFilePath][exportAssign.Expression]; ok {
				cr.ResolutionMaps[absFilePath]["default"] = uid
			}
		}
	}

	// 步骤 4: 处理 export 声明。
	for _, exportDecl := range parserResult.ExportDeclarations {
		if exportDecl.Source != "" { // 处理 `export ... from '...'` 的情况
			nextFile := cr.resolveModulePath(absFilePath, exportDecl.Source)
			if nextFile == "" {
				continue
			}
			depResMap, err := cr.resolveFile(nextFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "警告: 无法解析文件 '%s' 中的依赖 '%s': %v", absFilePath, exportDecl.Source, err)
				continue
			}
			// 将从其他模块重新导出的类型添加到当前文件的解析图中。
			for _, module := range exportDecl.ExportModules {
				if depUID, ok := depResMap[module.ModuleName]; ok {
					cr.ResolutionMaps[absFilePath][module.Identifier] = depUID
				}
			}
		} else { // 处理 `export { Name }` 或 `export default Name` 的情况
			// 将本地导出的类型（可能带有别名）添加到解析图中。
			for _, module := range exportDecl.ExportModules {
				if realUID, ok := cr.ResolutionMaps[absFilePath][module.ModuleName]; ok {
					cr.ResolutionMaps[absFilePath][module.Identifier] = realUID
				}
			}
		}
	}

	return cr.ResolutionMaps[absFilePath], nil
}

// resolveModulePath 将模块导入的相对路径或别名路径解析为最终的绝对文件路径。
// 它利用 ProjectConfig 中的 tsconfig.json 信息（如 alias, baseUrl）来正确处理非相对路径。
func (cr *CollectResult) resolveModulePath(currentFilePath, moduleSource string) string {
	// 使用 projectParser 的功能来匹配导入源。
	sourceData := projectParser.MatchImportSource(currentFilePath, moduleSource, cr.ProjectConfig.RootPath, cr.ProjectConfig.RootTsConfig.Alias, cr.ProjectConfig.Extensions, cr.ProjectConfig.RootTsConfig.BaseUrl)

	nextFile := ""
	if sourceData.Type == "file" {
		nextFile = sourceData.FilePath
	} else if sourceData.Type == "npm" {
		// 如果是 npm 包，尝试解析其类型定义文件。
		nextFile = utils.ResolveNpmPath(currentFilePath, cr.ProjectConfig.RootPath, moduleSource, true)
		if !utils.HasExtension(nextFile, cr.ProjectConfig.Extensions) {
			// 如果解析到的路径没有扩展名，尝试查找实际的文件（例如，补全 .ts, .d.ts 等）。
			nextFile = utils.FindRealFilePath(nextFile, cr.ProjectConfig.Extensions)
		}
	}

	// 确保最终路径是绝对路径。
	if nextFile != "" && !filepath.IsAbs(nextFile) {
		nextFile, _ = filepath.Abs(nextFile)
	}
	return nextFile
}
