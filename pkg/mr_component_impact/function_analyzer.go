// Package mr_component_impact 提供 MR 组件影响分析功能
package mr_component_impact

import (
	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/export_call"
)

// =============================================================================
// 函数影响分析器
// =============================================================================

// FunctionImpactAnalyzer 函数影响分析器
// 基于 export_call 的结果分析函数变更的影响
// export_call 已原生支持组件级引用（RefComponents 字段）
type FunctionImpactAnalyzer struct {
	exportCall *export_call.ExportCallResult
}

// NewFunctionImpactAnalyzer 创建函数影响分析器
func NewFunctionImpactAnalyzer(
	exportCall *export_call.ExportCallResult,
) *FunctionImpactAnalyzer {
	return &FunctionImpactAnalyzer{
		exportCall: exportCall,
	}
}

// AnalyzeFunctionChange 分析函数文件变更的影响
// 返回受影响的所有组件信息列表
func (a *FunctionImpactAnalyzer) AnalyzeFunctionChange(
	functionFile string,
	functionName string,
) []ComponentImpact {
	if a.exportCall == nil {
		return nil
	}

	impacts := make([]ComponentImpact, 0)

	// 遍历所有模块导出记录
	for _, module := range a.exportCall.ModuleExports {
		// 遍历每个文件的导出节点
		for _, fileRecord := range module.Files {
			// 跳过非变更的函数文件
			if fileRecord.File != functionFile {
				continue
			}

			// 遍历该文件的所有导出节点
			for _, node := range fileRecord.Nodes {
				// 直接使用 export_call 提供的组件级引用
				for _, compRef := range node.RefComponents {
					impacts = append(impacts, ComponentImpact{
						Component:    compRef.ComponentName,
						ChangeSource: functionName, // 使用函数名
						Relation:     RelationImports,
						Level:        1,
					})
				}
			}
		}
	}

	return impacts
}
