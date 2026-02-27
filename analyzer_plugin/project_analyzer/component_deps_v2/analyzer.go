package component_deps_v2

import (
	"fmt"
	"path/filepath"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

func init() {
	// 注册分析器到工厂
	projectanalyzer.RegisterAnalyzer("component-deps-v2", func() projectanalyzer.Analyzer {
		return &ComponentDepsV2Analyzer{}
	})
}

// =============================================================================
// 分析器实现
// =============================================================================

// ComponentDepsV2Analyzer 组件依赖分析器（V2版本）
//
// 使用方式：
//
//	analyzer-ts analyze component-deps-v2 \
//	  -i /path/to/project \
//	  -p "component-deps-v2.manifest=path/to/component-manifest.json"
type ComponentDepsV2Analyzer struct {
	// ManifestPath 配置文件路径
	// 可以是绝对路径或相对于项目根目录的路径
	ManifestPath string

	// manifest 加载后的配置对象
	manifest *ComponentManifest
}

// Name 返回分析器标识符
func (a *ComponentDepsV2Analyzer) Name() string {
	return "component-deps-v2"
}

// Configure 配置分析器参数
// 支持的参数：
//   - manifest: 配置文件路径（必需）
func (a *ComponentDepsV2Analyzer) Configure(params map[string]string) error {
	// 获取配置文件路径
	manifestPath, ok := params["manifest"]
	if !ok {
		return fmt.Errorf("缺少必需参数: manifest\n" +
			"请使用 -p 'component-deps-v2.manifest=path/to/component-manifest.json' 指定配置文件")
	}
	a.ManifestPath = manifestPath

	return nil
}

// Analyze 执行组件依赖分析
// 分析流程：
// 1. 加载配置文件
// 2. 分析外部依赖
// 3. 生成分析结果
func (a *ComponentDepsV2Analyzer) Analyze(ctx *projectanalyzer.ProjectContext) (projectanalyzer.Result, error) {
	// 步骤 1: 加载配置文件
	if err := a.loadManifest(ctx.ProjectRoot); err != nil {
		return nil, fmt.Errorf("加载配置文件失败: %w", err)
	}

	// 步骤 2: 分析外部依赖
	depAnalyzer := NewDependencyAnalyzer(a.manifest)
	fileResults := ctx.ParsingResult.Js_Data
	dependencies := depAnalyzer.AnalyzeAllComponents(fileResults)

	// 步骤 3: 构建结果（传递 depAnalyzer 以便重导出解析器可用）
	result := &ComponentDepsV2Result{
		Meta: Meta{
			ComponentCount: len(a.manifest.Components),
		},
		Components: a.buildComponentInfo(dependencies, depAnalyzer),
	}

	return result, nil
}

// buildComponentInfo 构建组件信息
func (a *ComponentDepsV2Analyzer) buildComponentInfo(
	dependencies map[string][]projectParser.ImportDeclarationResult,
	depAnalyzer *DependencyAnalyzer,
) map[string]ComponentInfo {
	result := make(map[string]ComponentInfo)

	for _, comp := range a.manifest.Components {
		deps := dependencies[comp.Name]

		// 分类依赖
		classified := depAnalyzer.ClassifyDependencies(deps)

		// 转换 ComponentDepDetail 为 ComponentDep
		componentDeps := make([]ComponentDep, 0, len(classified.ComponentDeps))
		for _, detail := range classified.ComponentDeps {
			componentDeps = append(componentDeps, ComponentDep{
				Name:     detail.Name,
				Path:     detail.Path,
				DepFiles: detail.DepFiles,
			})
		}

		result[comp.Name] = ComponentInfo{
			Name:         comp.Name,
			Path:         comp.Path,
			Dependencies: deps,
			NpmDeps:      classified.NpmDeps,
			ComponentDeps: componentDeps,
		}
	}

	return result
}

// =============================================================================
// 私有方法
// =============================================================================

// loadManifest 加载配置文件
// 支持绝对路径和相对路径
func (a *ComponentDepsV2Analyzer) loadManifest(projectRoot string) error {
	var manifestPath string

	// 判断是否为绝对路径
	if filepath.IsAbs(a.ManifestPath) {
		manifestPath = a.ManifestPath
	} else {
		// 相对路径，基于项目根目录
		manifestPath = filepath.Join(projectRoot, a.ManifestPath)
	}

	// 加载配置文件
	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		return err
	}

	a.manifest = manifest
	return nil
}
