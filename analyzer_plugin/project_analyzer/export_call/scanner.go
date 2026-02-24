// Package export_call 导出节点扫描器
package export_call

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
)

// ExportScanner 导出节点扫描器
type ExportScanner struct {
	assets   []AssetItem
	resolver *SymbolResolver
}

// NewExportScanner 创建导出节点扫描器
func NewExportScanner(assets []AssetItem) *ExportScanner {
	return &ExportScanner{
		assets: assets,
	}
}

// ScanAll 扫描所有资产目录，提取所有导出节点
func (s *ExportScanner) ScanAll(jsData map[string]projectParser.JsFileParserResult) []*ExportNode {
	s.resolver = NewSymbolResolver(jsData)
	var nodes []*ExportNode

	// 用于去重：key = filePath:nodeName:exportType
	seen := make(map[string]bool)

	for _, asset := range s.assets {
		for filePath, fileData := range jsData {
			if !s.isInAssetDirectory(filePath, asset.Path) {
				continue
			}

			// 1. 先提取 default export（包括 ExportAssignment 和 IsDefaultExport 的函数）
			for _, node := range s.extractDefaultExports(&fileData, asset, filePath) {
				key := fmt.Sprintf("%s:%s:%s", filePath, node.Name, node.ExportType)
				if !seen[key] {
					seen[key] = true
					nodes = append(nodes, node)
				}
			}

			// 2. 提取 export {} 声明
			for _, node := range s.extractExportDeclarations(&fileData, asset, filePath) {
				key := fmt.Sprintf("%s:%s:%s", filePath, node.Name, node.ExportType)
				if !seen[key] {
					seen[key] = true
					nodes = append(nodes, node)
				}
			}

			// 3. 提取直接导出的声明（排除已经是 default 的）
			for _, node := range s.extractNamedExports(&fileData, asset, filePath) {
				key := fmt.Sprintf("%s:%s:%s", filePath, node.Name, node.ExportType)
				if !seen[key] {
					seen[key] = true
					nodes = append(nodes, node)
				}
			}
		}
	}

	return nodes
}

// extractNamedExports 提取 named export（直接导出的声明）
func (s *ExportScanner) extractNamedExports(
	fileData *projectParser.JsFileParserResult,
	asset AssetItem,
	filePath string,
) []*ExportNode {
	var nodes []*ExportNode

	// VariableDeclarations - 先处理，因为 const foo = () => {} 应该是 variable 类型
	for _, v := range fileData.VariableDeclarations {
		if v.Exported {
			for _, decl := range v.Declarators {
				if decl.Identifier != "" {
					nodes = append(nodes, &ExportNode{
						ID:         fmt.Sprintf("%s:%s:named", asset.Name, decl.Identifier),
						Name:       decl.Identifier,
						AssetName:  asset.Name,
						NodeType:   NodeTypeVariable,
						ExportType: ExportTypeNamed,
						SourceFile: filePath,
					})
				}
			}
		}
	}

	// FunctionDeclarations - 只处理非 default export 的函数
	for _, fn := range fileData.FunctionDeclarations {
		if fn.Exported && !fn.IsDefaultExport {
			nodes = append(nodes, &ExportNode{
				ID:         fmt.Sprintf("%s:%s:named", asset.Name, fn.Identifier),
				Name:       fn.Identifier,
				AssetName:  asset.Name,
				NodeType:   NodeTypeFunction,
				ExportType: ExportTypeNamed,
				SourceFile: filePath,
			})
		}
	}

	// TypeDeclarations
	for name, t := range fileData.TypeDeclarations {
		if t.Exported {
			nodes = append(nodes, &ExportNode{
				ID:         fmt.Sprintf("%s:%s:named", asset.Name, name),
				Name:       name,
				AssetName:  asset.Name,
				NodeType:   NodeTypeType,
				ExportType: ExportTypeNamed,
				SourceFile: filePath,
			})
		}
	}

	// InterfaceDeclarations
	for name, iface := range fileData.InterfaceDeclarations {
		if iface.Exported {
			nodes = append(nodes, &ExportNode{
				ID:         fmt.Sprintf("%s:%s:named", asset.Name, name),
				Name:       name,
				AssetName:  asset.Name,
				NodeType:   NodeTypeInterface,
				ExportType: ExportTypeNamed,
				SourceFile: filePath,
			})
		}
	}

	// EnumDeclarations
	for name, enum := range fileData.EnumDeclarations {
		if enum.Exported {
			nodes = append(nodes, &ExportNode{
				ID:         fmt.Sprintf("%s:%s:named", asset.Name, name),
				Name:       name,
				AssetName:  asset.Name,
				NodeType:   NodeTypeEnum,
				ExportType: ExportTypeNamed,
				SourceFile: filePath,
			})
		}
	}

	return nodes
}

// extractExportDeclarations 提取 export {} 声明中的导出
func (s *ExportScanner) extractExportDeclarations(
	fileData *projectParser.JsFileParserResult,
	asset AssetItem,
	filePath string,
) []*ExportNode {
	var nodes []*ExportNode

	for _, exportDecl := range fileData.ExportDeclarations {
		// 跳过 re-export (Source 不为 nil 表示是 re-export)
		if exportDecl.Source != nil {
			continue
		}

		// 解析符号，获取每个导出的真实节点类型
		symbolTypes := s.resolver.ResolveExportDeclaration(fileData, &exportDecl)

		for _, module := range exportDecl.ExportModules {
			nodeType := symbolTypes[module.ModuleName]
			if nodeType == "" {
				nodeType = NodeTypeVariable
			}

			nodes = append(nodes, &ExportNode{
				ID:         fmt.Sprintf("%s:%s:named", asset.Name, module.ModuleName),
				Name:       module.ModuleName,
				AssetName:  asset.Name,
				NodeType:   nodeType,
				ExportType: ExportTypeNamed,
				SourceFile: filePath,
			})
		}
	}

	return nodes
}

// extractDefaultExports 提取 default export
func (s *ExportScanner) extractDefaultExports(
	fileData *projectParser.JsFileParserResult,
	asset AssetItem,
	filePath string,
) []*ExportNode {
	var nodes []*ExportNode

	// 1. 处理 ExportAssignment（export default foo; 形式）
	for _, exportAssign := range fileData.ExportAssignments {
		nodeType := s.resolver.ResolveExportAssignment(fileData, &exportAssign)

		// 优先使用 parser 预先提取的 Name
		// 如果 Name 为空（匿名导出），则使用 "default"
		name := exportAssign.Name
		if name == "" {
			name = "default"
		}

		nodes = append(nodes, &ExportNode{
			ID:         fmt.Sprintf("%s:%s:default", asset.Name, name),
			Name:       name,
			AssetName:  asset.Name,
			NodeType:   nodeType,
			ExportType: ExportTypeDefault,
			SourceFile: filePath,
		})
	}

	// 2. 处理 IsDefaultExport == true 的函数声明（export default function foo() {} 形式）
	for _, fn := range fileData.FunctionDeclarations {
		if fn.IsDefaultExport && fn.Identifier != "" {
			nodes = append(nodes, &ExportNode{
				ID:         fmt.Sprintf("%s:%s:default", asset.Name, fn.Identifier),
				Name:       fn.Identifier,
				AssetName:  asset.Name,
				NodeType:   NodeTypeFunction,
				ExportType: ExportTypeDefault,
				SourceFile: filePath,
			})
		}
	}

	return nodes
}

// isInAssetDirectory 判断文件是否在资产目录下
func (s *ExportScanner) isInAssetDirectory(filePath, assetPath string) bool {
	normalizedPath := filepath.ToSlash(filePath)
	normalizedAssetPath := filepath.ToSlash(assetPath)

	return strings.HasPrefix(normalizedPath, normalizedAssetPath+"/")
}
