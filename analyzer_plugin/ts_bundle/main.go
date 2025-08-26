package ts_bundle

import (
	"path/filepath"
)

// GenerateBundle 是 TypeScript 类型打包的入口方法。
// 它协调了依赖收集和类型打包的整个过程。
//
// 参数:
//   inputAnalyzeFile: 入口 TypeScript 文件路径。
//   inputAnalyzeType: 要分析的特定类型名称（目前可能未完全使用，但保留）。
//   outputFile: 最终打包内容的输出文件路径（此函数返回字符串，实际写入由调用者处理）。
//   projectRoot: 项目的根路径。如果为空，则尝试从入口文件推断。
//
// 返回值:
//   包含所有打包类型声明的字符串内容，如果发生错误则返回错误信息。
func GenerateBundle(inputAnalyzeFile string, inputAnalyzeType string, outputFile string, projectRoot string) string {
	// 如果未提供项目根路径，则尝试从入口文件的绝对路径推断。
	// 这是一个简单的启发式方法，更健壮的方法是搜索 tsconfig.json。
	if projectRoot == "" {
		abs, err := filepath.Abs(inputAnalyzeFile)
		if err == nil {
			projectRoot = filepath.Dir(abs)
		} else {
			projectRoot = "."
		}
	}

	// 步骤 1: 收集所有依赖的类型声明及其引用关系。
	collector := NewCollectResult(projectRoot)
	err := collector.CollectDependencies(inputAnalyzeFile)
	if err != nil {
		return "依赖收集过程中发生错误: " + err.Error()
	}

	// 步骤 2: 使用收集到的信息进行类型打包。
	// 这包括解决名称冲突、更新引用等。
	bundler := NewTypeBundler(collector)
	bundledContent, err := bundler.Bundle(inputAnalyzeFile, inputAnalyzeType)
	if err != nil {
		return "打包过程中发生错误: " + err.Error()
	}

	return bundledContent
}
