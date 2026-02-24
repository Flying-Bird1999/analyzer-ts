// Package export_call 引用关系构建器
package export_call

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
)

// ReferenceBuilder 引用关系构建器
type ReferenceBuilder struct {
	assets []AssetItem
	jsData map[string]projectParser.JsFileParserResult

	// exportIndex: (assetName, nodeName, exportType) -> ExportNode
	exportIndex map[string]*ExportNode

	// assetPathMap: filePath -> AssetItem
	assetPathMap map[string]*AssetItem
}

// NewReferenceBuilder 创建引用关系构建器
func NewReferenceBuilder(
	assets []AssetItem,
	jsData map[string]projectParser.JsFileParserResult,
) *ReferenceBuilder {
	rb := &ReferenceBuilder{
		assets:       assets,
		jsData:       jsData,
		exportIndex: make(map[string]*ExportNode),
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
func (b *ReferenceBuilder) BuildReferences(exportNodes []*ExportNode) map[string][]string {
	// 构建导出节点索引
	for _, node := range exportNodes {
		key := b.getExportKey(node.AssetName, node.Name, node.ExportType)
		b.exportIndex[key] = node
	}

	// 构建引用关系
	refMap := make(map[string][]string)

	for filePath, fileData := range b.jsData {
		// 遍历该文件的所有 import
		for _, importDecl := range fileData.ImportDeclarations {
			// 只处理文件类型的 import
			if importDecl.Source.Type != "file" {
				continue
			}

			sourceFile := importDecl.Source.FilePath
			if sourceFile == "" {
				continue
			}

			// 找到 import 来源所属的资产
			asset, ok := b.assetPathMap[sourceFile]
			if !ok {
				continue
			}

			// 根据每个 import module 的类型匹配对应的导出节点
			for _, module := range importDecl.ImportModules {
				var node *ExportNode

				// 根据 ImportModule.Type 匹配不同的 ExportType
				switch module.Type {
				case "named":
					// import { foo } from './mod' → 匹配 named export 的 foo
					key := b.getExportKey(asset.Name, module.ImportModule, ExportTypeNamed)
					node = b.exportIndex[key]

				case "default":
					// import foo from './mod' → 匹配 sourceFile 对应的 default export
					// 由于 default export 的 name 可能是任意值（如 "wrapperRaf"），需要遍历查找
					for _, exportNode := range exportNodes {
						if exportNode.AssetName == asset.Name &&
						   exportNode.ExportType == ExportTypeDefault &&
						   exportNode.SourceFile == sourceFile {
							node = exportNode
							break
						}
					}

				case "namespace":
					// import * as foo from './mod' → 匹配该资产的所有 named export
					for _, exportNode := range exportNodes {
						if exportNode.AssetName == asset.Name && exportNode.ExportType == ExportTypeNamed {
							refMap[exportNode.ID] = append(refMap[exportNode.ID], filePath)
						}
					}
					continue // namespace 特殊处理，跳过后续逻辑
				}

				// 记录引用关系
				if node != nil {
					refMap[node.ID] = append(refMap[node.ID], filePath)
				}
			}
		}
	}

	return refMap
}

// getExportKey 构建导出节点的唯一键
func (b *ReferenceBuilder) getExportKey(assetName, nodeName string, exportType ExportType) string {
	return fmt.Sprintf("%s:%s:%s", assetName, nodeName, exportType)
}

// isInAssetDirectory 判断文件是否在资产目录下
func (b *ReferenceBuilder) isInAssetDirectory(filePath, assetPath string) bool {
	normalizedPath := filepath.ToSlash(filePath)
	normalizedAssetPath := filepath.ToSlash(assetPath)

	return strings.HasPrefix(normalizedPath, normalizedAssetPath+"/")
}
