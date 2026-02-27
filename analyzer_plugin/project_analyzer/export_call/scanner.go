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

	// 用于去重：key = sourceFile:nodeName:exportType
	// 注意：使用节点的 SourceFile 而不是当前扫描的 filePath
	// 这样可以正确处理重导出的情况
	seen := make(map[string]bool)

	for _, asset := range s.assets {
		for filePath, fileData := range jsData {
			if !s.isInAssetDirectory(filePath, asset.Path) {
				continue
			}

			// 1. 先提取 default export（包括 ExportAssignment 和 IsDefaultExport 的函数）
			for _, node := range s.extractDefaultExports(&fileData, asset, filePath) {
				key := fmt.Sprintf("%s:%s:%s", node.SourceFile, node.Name, node.ExportType)
				if !seen[key] {
					seen[key] = true
					nodes = append(nodes, node)
				}
			}

			// 2. 提取 export {} 声明（包括重导出）
			for _, node := range s.extractExportDeclarations(&fileData, asset, filePath) {
				key := fmt.Sprintf("%s:%s:%s", node.SourceFile, node.Name, node.ExportType)
				if !seen[key] {
					seen[key] = true
					nodes = append(nodes, node)
				}
			}

			// 3. 提取直接导出的声明（排除已经是 default 的）
			for _, node := range s.extractNamedExports(&fileData, asset, filePath) {
				key := fmt.Sprintf("%s:%s:%s", node.SourceFile, node.Name, node.ExportType)
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
// 支持解析重导出（export { xxx } from './path'），追踪到真实定义源
func (s *ExportScanner) extractExportDeclarations(
	fileData *projectParser.JsFileParserResult,
	asset AssetItem,
	filePath string,
) []*ExportNode {
	var nodes []*ExportNode

	for _, exportDecl := range fileData.ExportDeclarations {
		// 处理重导出（Source 不为 nil 表示是 re-export）
		if exportDecl.Source != nil {
			// 重导出：export { xxx } from './path'
			// 追踪到真实定义源
			reexportedNodes := s.resolveReExport(&exportDecl, asset)
			nodes = append(nodes, reexportedNodes...)
			continue
		}

		// 直接导出：export { xxx }
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

// resolveReExport 解析重导出，追踪到真实定义源
// 支持递归追踪重导出链
func (s *ExportScanner) resolveReExport(
	exportDecl *projectParser.ExportDeclarationResult,
	asset AssetItem,
) []*ExportNode {
	var nodes []*ExportNode

	// 只处理文件类型的重导出
	if exportDecl.Source.Type != "file" {
		return nodes
	}

	sourceFile := exportDecl.Source.FilePath
	if sourceFile == "" {
		return nodes
	}

	// 获取源文件的解析结果
	sourceFileData, exists := s.resolver.jsData[sourceFile]
	if !exists {
		// 源文件未被解析，无法追踪
		return nodes
	}

	// 防止循环重导出
	visited := make(map[string]bool)
	return s.resolveReExportRecursive(sourceFileData, exportDecl, asset, sourceFile, visited)
}

// resolveReExportRecursive 递归解析重导出
func (s *ExportScanner) resolveReExportRecursive(
	sourceFileData projectParser.JsFileParserResult,
	exportDecl *projectParser.ExportDeclarationResult,
	asset AssetItem,
	sourceFilePath string,
	visited map[string]bool,
) []*ExportNode {
	var nodes []*ExportNode

	// 防止循环
	if visited[sourceFilePath] {
		return nodes
	}
	visited[sourceFilePath] = true

	for _, module := range exportDecl.ExportModules {
		moduleName := module.ModuleName

		// 处理 export * 的情况
		if moduleName == "*" {
			// 递归获取源文件的所有导出
			nodes = append(nodes, s.extractAllExportsFromFile(sourceFileData, asset, sourceFilePath, visited)...)
			continue
		}

		// 普通导出：export { xxx as yyy } from './path'
		// 在源文件中查找该符号的定义
		nodeType := s.findSymbolInFile(sourceFileData, moduleName)

		// 使用外部名称（Identifier）作为导出名称
		exportName := moduleName
		if module.Identifier != "" && module.Identifier != moduleName {
			exportName = module.Identifier
		}

		// 判断是 default export 还是 named export
		// export { default as foo } from './path' 中，ModuleName 是 "default"
		exportType := ExportTypeNamed
		if moduleName == "default" {
			exportType = ExportTypeDefault
		}

		nodes = append(nodes, &ExportNode{
			ID:         fmt.Sprintf("%s:%s:%s", asset.Name, exportName, exportType),
			Name:       exportName,
			AssetName:  asset.Name,
			NodeType:   nodeType,
			ExportType: exportType,
			SourceFile: sourceFilePath,
		})
	}

	return nodes
}

// extractAllExportsFromFile 从文件中提取所有导出（用于 export * 的情况）
func (s *ExportScanner) extractAllExportsFromFile(
	fileData projectParser.JsFileParserResult,
	asset AssetItem,
	filePath string,
	visited map[string]bool,
) []*ExportNode {
	var nodes []*ExportNode

	// 1. 提取 default export
	nodes = append(nodes, s.extractDefaultExportsFromFile(fileData, asset, filePath)...)

	// 2. 提取 named export
	nodes = append(nodes, s.extractNamedExportsFromFile(fileData, asset, filePath)...)

	// 3. 递归处理该文件中的重导出
	for _, exportDecl := range fileData.ExportDeclarations {
		if exportDecl.Source != nil && exportDecl.Source.Type == "file" {
			reexportedPath := exportDecl.Source.FilePath
			if reexportedPath != "" && !visited[reexportedPath] {
				reexportedData, exists := s.resolver.jsData[reexportedPath]
				if exists {
					nodes = append(nodes, s.resolveReExportRecursive(reexportedData, &exportDecl, asset, reexportedPath, visited)...)
				}
			}
		}
	}

	return nodes
}

// extractDefaultExportsFromFile 从文件中提取 default export
func (s *ExportScanner) extractDefaultExportsFromFile(
	fileData projectParser.JsFileParserResult,
	asset AssetItem,
	filePath string,
) []*ExportNode {
	var nodes []*ExportNode

	// 1. 处理 ExportAssignment
	for _, exportAssign := range fileData.ExportAssignments {
		nodeType := s.resolver.ResolveExportAssignment(&fileData, &exportAssign)
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

	// 2. 处理 IsDefaultExport == true 的函数声明
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

// extractNamedExportsFromFile 从文件中提取 named export
func (s *ExportScanner) extractNamedExportsFromFile(
	fileData projectParser.JsFileParserResult,
	asset AssetItem,
	filePath string,
) []*ExportNode {
	var nodes []*ExportNode

	// 1. VariableDeclarations
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

	// 2. FunctionDeclarations（非 default）
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

	// 3. TypeDeclarations
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

	// 4. InterfaceDeclarations
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

	// 5. EnumDeclarations
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

// findSymbolInFile 在文件中查找符号定义
func (s *ExportScanner) findSymbolInFile(
	fileData projectParser.JsFileParserResult,
	symbolName string,
) NodeType {
	return s.resolver.findSymbolDefinition(&fileData, symbolName)
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
