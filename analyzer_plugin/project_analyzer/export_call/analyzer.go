// Package export_call 实现了工具函数模块导出节点引用关系分析插件
//
// 核心功能：
// 1. 扫描指定的工具函数目录，采集所有导出节点（function/variable/type/interface/enum）
// 2. 通过 ImportDeclarations 建立引用关系
// 3. 按文件分组输出导出节点和引用信息
//
// 注意：本插件只处理 manifest.json 中的 functions 配置项，
//       components 由 component_deps_v2 插件处理
package export_call

import (
	"fmt"
	"path/filepath"
	"sort"

	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

// ExportCallAnalyzer 导出节点引用分析器
type ExportCallAnalyzer struct {
	// ManifestPath 配置文件路径
	ManifestPath string

	// manifest 加载后的配置对象
	manifest *AssetManifest
}

// Name 返回分析器标识符
func (a *ExportCallAnalyzer) Name() string {
	return "export-call"
}

// Configure 配置分析器参数
// 支持的参数：
//   - manifest: 配置文件路径（必需）
func (a *ExportCallAnalyzer) Configure(params map[string]string) error {
	manifestPath, ok := params["manifest"]
	if !ok {
		return fmt.Errorf("缺少必需参数: manifest\n" +
			"请使用 -p 'export-call.manifest=path/to/manifest.json' 指定配置文件")
	}
	a.ManifestPath = manifestPath

	return nil
}

// Analyze 执行导出节点引用分析
// 分析流程：
// 1. 加载配置文件
// 2. 扫描目录提取导出节点
// 3. 分析引用关系
// 4. 构建结果
func (a *ExportCallAnalyzer) Analyze(ctx *projectanalyzer.ProjectContext) (projectanalyzer.Result, error) {
	// 步骤 1: 加载配置文件
	if err := a.loadManifest(ctx.ProjectRoot); err != nil {
		return nil, fmt.Errorf("加载配置文件失败: %w", err)
	}

	// 步骤 2: 合并 components 和 functions 为统一资产列表
	assets := a.buildAssetList()

	// 步骤 2.5: 将相对路径转换为绝对路径
	assets = a.resolveAssetPaths(assets, ctx.ProjectRoot)

	// 步骤 3: 扫描导出节点
	scanner := NewExportScanner(assets)
	exportNodes := scanner.ScanAll(ctx.ParsingResult.Js_Data)

	// 步骤 4: 建立引用关系
	builder := NewReferenceBuilder(assets, ctx.ParsingResult.Js_Data)
	refMap := builder.BuildReferences(exportNodes)

	// 步骤 5: 按模块分组构建结果
	result := &ExportCallResult{
		ModuleExports: a.buildModuleExports(assets, exportNodes, refMap),
	}

	return result, nil
}

// buildAssetList 仅使用 functions 构建资产列表
// Components 由 component_deps_v2 插件处理，本插件只关注工具函数模块
func (a *ExportCallAnalyzer) buildAssetList() []AssetItem {
	assets := make([]AssetItem, len(a.manifest.Functions))

	// 仅添加 functions
	for i, fn := range a.manifest.Functions {
		assets[i] = AssetItem{
			Name: fn.Name,
			Type: fn.Type,
			Path: fn.Path,
		}
	}

	return assets
}

// buildModuleExports 按模块分组构建导出记录
func (a *ExportCallAnalyzer) buildModuleExports(
	assets []AssetItem,
	exportNodes []*ExportNode,
	refMap map[string][]string,
) []ModuleExportRecord {
	// 构建 assetName -> assetPath 映射（使用原始相对路径）
	assetPathMap := make(map[string]string)
	for _, asset := range assets {
		// 这里需要获取原始配置的相对路径，而不是解析后的绝对路径
		// 由于 resolveAssetPaths 已经将路径转换为绝对路径，我们需要从 manifest 中获取原始路径
		for _, fn := range a.manifest.Functions {
			if fn.Name == asset.Name {
				assetPathMap[asset.Name] = fn.Path
				break
			}
		}
	}

	// 按模块分组
	// moduleMap: assetName -> (fileMap: filePath -> *FileExportRecord)
	moduleMap := make(map[string]map[string]*FileExportRecord)

	for _, node := range exportNodes {
		// 获取该节点所属资产的文件映射
		fileMap, ok := moduleMap[node.AssetName]
		if !ok {
			fileMap = make(map[string]*FileExportRecord)
			moduleMap[node.AssetName] = fileMap
		}

		// 获取或创建文件记录
		record, ok := fileMap[node.SourceFile]
		if !ok {
			record = &FileExportRecord{
				File:  node.SourceFile,
				Nodes: []NodeWithRefs{},
			}
			fileMap[node.SourceFile] = record
		}

		record.Nodes = append(record.Nodes, NodeWithRefs{
			Name:       node.Name,
			NodeType:   node.NodeType,
			ExportType: node.ExportType,
			RefFiles:   refMap[node.ID],
		})
	}

	// 构建模块记录
	var result []ModuleExportRecord

	// 按资产名称排序
	assetNames := make([]string, 0, len(assets))
	for _, asset := range assets {
		assetNames = append(assetNames, asset.Name)
	}
	sort.Strings(assetNames)

	for _, assetName := range assetNames {
		fileMap := moduleMap[assetName]
		if fileMap == nil {
			continue
		}

		// 转换文件映射为切片并按文件路径排序
		var files []FileExportRecord
		for filePath := range fileMap {
			files = append(files, *fileMap[filePath])
		}
		sort.Slice(files, func(i, j int) bool {
			return files[i].File < files[j].File
		})

		result = append(result, ModuleExportRecord{
			ModuleName: assetName,
			Path:       assetPathMap[assetName],
			Files:      files,
		})
	}

	return result
}

// loadManifest 加载配置文件
func (a *ExportCallAnalyzer) loadManifest(projectRoot string) error {
	var manifestPath string

	if filepath.IsAbs(a.ManifestPath) {
		manifestPath = a.ManifestPath
	} else {
		manifestPath = filepath.Join(projectRoot, a.ManifestPath)
	}

	manifest, err := LoadAssetManifest(manifestPath)
	if err != nil {
		return err
	}

	a.manifest = manifest
	return nil
}

// resolveAssetPaths 将资产路径转换为绝对路径
// 如果路径已经是绝对路径，则保持不变
func (a *ExportCallAnalyzer) resolveAssetPaths(assets []AssetItem, projectRoot string) []AssetItem {
	resolved := make([]AssetItem, len(assets))

	for i, asset := range assets {
		resolved[i] = asset
		if filepath.IsAbs(asset.Path) {
			// 已经是绝对路径，直接使用
			resolved[i].Path = asset.Path
		} else {
			// 相对路径，转换为绝对路径
			resolved[i].Path = filepath.Join(projectRoot, asset.Path)
		}
	}

	return resolved
}
