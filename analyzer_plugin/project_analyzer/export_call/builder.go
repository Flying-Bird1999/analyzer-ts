// Package export_call 引用关系构建器
package export_call

import (
	"path/filepath"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
)

// ReferenceBuilder 引用关系构建器
type ReferenceBuilder struct {
	assets []AssetItem
	jsData map[string]projectParser.JsFileParserResult

	// assetPathMap: filePath -> AssetItem
	assetPathMap map[string]*AssetItem
}

// NewReferenceBuilder 创建引用关系构建器
func NewReferenceBuilder(
	assets []AssetItem,
	jsData map[string]projectParser.JsFileParserResult,
) *ReferenceBuilder {
	rb := &ReferenceBuilder{
		assets:      assets,
		jsData:      jsData,
		assetPathMap: make(map[string]*AssetItem),
	}

	// 构建 assetPathMap
	for i := range assets {
		asset := &assets[i]
		for filePath := range jsData {
			if rb.isInAssetDirectory(filePath, asset.Path) {
				rb.assetPathMap[filePath] = asset
			}
		}
	}

	return rb
}

// BuildReferences 构建引用关系
// 支持重导出解析：解析重导出入口，找到实际定义文件和对应的导出节点
func (b *ReferenceBuilder) BuildReferences(exportNodes []*ExportNode) map[string][]string {
	// 构建引用关系
	refMap := make(map[string][]string)

	for filePath, fileData := range b.jsData {
		// 遍历该文件的所有 import
		for _, importDecl := range fileData.ImportDeclarations {
			// 只处理文件类型的 import（包括经过路径别名解析后的npm包）
			if importDecl.Source.Type != "file" {
				continue
			}

			sourceFile := importDecl.Source.FilePath
			if sourceFile == "" {
				continue
			}

			// 根据每个 import module 的类型匹配对应的导出节点
			for _, module := range importDecl.ImportModules {
				// 解析重导出，获取实际源文件
				actualSourceFile := b.resolveSourceFile(sourceFile, module)
				if actualSourceFile == "" {
					continue
				}

				// 根据导入类型匹配导出节点
				switch module.Type {
				case "namespace":
					// import * as foo from './mod'
					// 匹配该文件的所有 named export
					for _, exportNode := range exportNodes {
						if exportNode.SourceFile == actualSourceFile &&
							exportNode.ExportType == ExportTypeNamed {
							refMap[exportNode.ID] = append(refMap[exportNode.ID], filePath)
						}
					}
				default:
					// named 或 default import
					node := b.matchExportNode(exportNodes, actualSourceFile, module)
					if node != nil {
						refMap[node.ID] = append(refMap[node.ID], filePath)
					}
				}
			}
		}
	}

	return refMap
}

// resolveSourceFile 解析源文件，处理重导出情况
func (b *ReferenceBuilder) resolveSourceFile(sourceFile string, module projectParser.ImportModule) string {
	// 如果 sourceFile 在资产下，直接返回（没有重导出）
	if _, ok := b.assetPathMap[sourceFile]; ok {
		return sourceFile
	}

	// 不在资产下，可能是重导出入口，尝试解析
	moduleName := module.ImportModule
	if module.Type == "namespace" {
		moduleName = "*"
	}
	return b.resolveReExportFile(sourceFile, moduleName)
}

// matchExportNode 匹配导出节点
func (b *ReferenceBuilder) matchExportNode(exportNodes []*ExportNode, sourceFile string, module projectParser.ImportModule) *ExportNode {
	for _, exportNode := range exportNodes {
		if exportNode.SourceFile != sourceFile {
			continue
		}

		// 根据 import 类型匹配 export 类型
		switch module.Type {
		case "named":
			// import { foo } from './mod'
			// 导出可能是 named export，也可能是 re-exported default export
			// 例如：export { default as useCallbackState } from './path'
			if exportNode.Name == module.ImportModule &&
				(exportNode.ExportType == ExportTypeNamed || exportNode.ExportType == ExportTypeDefault) {
				return exportNode
			}
		case "default":
			// import foo from './mod'
			if exportNode.ExportType == ExportTypeDefault {
				return exportNode
			}
		}
	}
	return nil
}

// isInAssetDirectory 判断文件是否在资产目录下
func (b *ReferenceBuilder) isInAssetDirectory(filePath, assetPath string) bool {
	normalizedPath := filepath.ToSlash(filePath)
	normalizedAssetPath := filepath.ToSlash(assetPath)

	return strings.HasPrefix(normalizedPath, normalizedAssetPath+"/")
}

// resolveReExportFile 解析重导出链，找到模块的实际定义文件
// 例如：sourceFile = "packages/atlas/src/index.ts", moduleName = "useCallbackState"
// 返回 "packages/atlas/src/core/helper/_utils/useCallbackState.ts"
func (b *ReferenceBuilder) resolveReExportFile(sourceFile, moduleName string) string {
	// 标准化路径：确保路径有文件扩展名
	normalizedSource := b.normalizeFilePath(sourceFile)

	// 获取 sourceFile 的解析数据
	fileData, exists := b.jsData[normalizedSource]
	if !exists {
		// 尝试通过后缀匹配查找文件（处理相对路径问题）
		for key := range b.jsData {
			if strings.HasSuffix(key, sourceFile) || strings.HasSuffix(key, normalizedSource) {
				sourceFile = key
				break
			}
		}
		fileData, exists = b.jsData[sourceFile]
		if !exists {
			return ""
		}
	}

	// 检查该文件是否有重导出声明
	for _, exportDecl := range fileData.ExportDeclarations {
		if exportDecl.Source == nil || exportDecl.Source.Type != "file" {
			continue
		}

		// 检查是否导出了目标模块
		for _, module := range exportDecl.ExportModules {
			// 使用外部名称（Identifier）进行匹配
			exportName := module.Identifier
			if exportName == "" {
				exportName = module.ModuleName
			}

			// 处理 export * 的情况
			if module.ModuleName == "*" {
				// 递归查找目标模块
				targetFile := b.normalizeFilePath(exportDecl.Source.FilePath)
				return b.resolveReExportFile(targetFile, moduleName)
			}

			// 匹配模块名称
			if exportName == moduleName {
				// 找到了，返回实际定义文件（确保有扩展名）
				return b.normalizeFilePath(exportDecl.Source.FilePath)
			}
		}
	}

	return ""
}

// normalizeFilePath 标准化文件路径，确保有文件扩展名
func (b *ReferenceBuilder) normalizeFilePath(filePath string) string {
	if filePath == "" {
		return filePath
	}

	// 如果已经有扩展名，直接返回
	ext := filepath.Ext(filePath)
	if ext != "" {
		return filePath
	}

	// 没有扩展名，尝试添加 .ts 或 .tsx
	extensions := []string{".ts", ".tsx", ".js", ".jsx"}
	for _, ext := range extensions {
		pathWithExt := filePath + ext
		if _, exists := b.jsData[pathWithExt]; exists {
			return pathWithExt
		}
	}

	return filePath
}
