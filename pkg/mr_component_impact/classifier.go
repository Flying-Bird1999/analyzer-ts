// Package mr_component_impact 提供 MR 组件影响分析功能
package mr_component_impact

import (
	"path/filepath"
	"strings"
)

// =============================================================================
// 文件分类器
// =============================================================================

// Classifier 文件分类器
// 用于判断文件属于组件、函数还是其他类型
type Classifier struct {
	manifest      *ComponentManifest
	functionPaths []string
}

// NewClassifier 创建文件分类器
func NewClassifier(manifest *ComponentManifest, functionPaths []string) *Classifier {
	return &Classifier{
		manifest:      manifest,
		functionPaths: functionPaths,
	}
}

// ClassifyFile 分类单个文件
// 返回: (category, name)
// - category: 文件类型（component/functions/other）
// - name: 组件名称或函数名称（如果是 other 类型则为空）
func (c *Classifier) ClassifyFile(filePath string) (FileCategory, string) {
	// 1. 检查是否为组件文件
	if compName := c.isComponentFile(filePath); compName != "" {
		return CategoryComponent, compName
	}

	// 2. 检查是否为 functions 文件
	if funcName := c.isFunctionFile(filePath); funcName != "" {
		return CategoryFunctions, funcName
	}

	// 3. 其他类型
	return CategoryOther, ""
}

// isComponentFile 判断是否为组件文件
func (c *Classifier) isComponentFile(filePath string) string {
	if c.manifest == nil {
		return ""
	}

	for _, comp := range c.manifest.Components {
		// 检查文件是否在组件路径下
		if isFileInPath(filePath, comp.Path) {
			return comp.Name
		}
	}

	return ""
}

// isFunctionFile 判断是否为 functions 文件
func (c *Classifier) isFunctionFile(filePath string) string {
	// 优先检查 manifest 中的 functions 配置（直接返回配置的名称）
	if c.manifest != nil {
		for _, funcInfo := range c.manifest.Functions {
			if isFileInPath(filePath, funcInfo.Path) {
				return funcInfo.Name
			}
		}
	}

	// 其次检查 functionPaths 列表（需要从路径提取名称）
	for _, funcPath := range c.functionPaths {
		if isFileInPath(filePath, funcPath) {
			// 从路径提取 function 名称
			return extractFunctionName(filePath, funcPath)
		}
	}

	return ""
}

// isFileInPath 检查文件是否在指定路径下
func isFileInPath(filePath, targetPath string) bool {
	relPath, err := filepath.Rel(targetPath, filePath)
	if err != nil {
		return false
	}
	// 如果相对路径不以 ".." 开头，说明文件在目标路径下
	return !strings.HasPrefix(relPath, "..") && !filepath.IsAbs(relPath)
}

// extractFunctionName 从文件路径提取 function 名称
// 例如: src/functions/utils/date.ts → utils
func extractFunctionName(filePath, functionsRoot string) string {
	relPath, err := filepath.Rel(functionsRoot, filePath)
	if err != nil {
		return ""
	}

	// 获取第一级目录作为 function 名称
	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) > 0 {
		return parts[0]
	}

	return filepath.Base(filePath)
}
