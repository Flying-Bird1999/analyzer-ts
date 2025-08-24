// Package parser 提供项目解析相关的公共逻辑，以供各个分析器复用。
package parser

import (
	"main/analyzer/projectParser"
)

// ParseProject 是一个公共函数，用于解析整个项目。
// 它封装了项目解析的通用逻辑，供各个分析器调用，避免代码重复。
// 参数:
//   - rootPath: 项目根目录的绝对路径。
//   - ignore: 需要从分析中排除的文件/目录的 glob 模式列表。
//   - isMonorepo: 指示项目是否为 monorepo。
// 返回值:
//   - *projectParser.ProjectParserResult: 解析后的项目结果。
//   - error: 解析过程中可能发生的错误。
func ParseProject(rootPath string, ignore []string, isMonorepo bool) (*projectParser.ProjectParserResult, error) {
	// 步骤 1: 创建项目解析器配置。
	config := projectParser.NewProjectParserConfig(rootPath, ignore, isMonorepo)
	
	// 步骤 2: 创建用于存储解析结果的容器。
	ar := projectParser.NewProjectParserResult(config)
	
	// 步骤 3: 运行主解析逻辑。
	ar.ProjectParser()
	
	// 在未来的版本中，这里可以增加更详细的错误处理逻辑。
	// 目前，ProjectParser 方法内部的错误处理比较简单。
	return ar, nil
}