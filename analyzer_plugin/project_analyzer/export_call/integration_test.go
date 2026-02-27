package export_call

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExportCallIntegration 测试基于真实 test_project 的集成场景
// 这是一个端到端测试，会真实解析项目并分析导出引用关系
func TestExportCallIntegration(t *testing.T) {
	// 获取测试项目路径
	testProjectPath, err := filepath.Abs("../../../testdata/test_project")
	require.NoError(t, err, "无法获取测试项目路径")

	manifestPath := filepath.Join(testProjectPath, ".analyzer", "component-manifest.json")

	// 1. 解析项目
	t.Log("开始解析测试项目...")
	config := projectParser.NewProjectParserConfig(testProjectPath, []string{}, false, []string{})
	parsingResult := projectParser.NewProjectParserResult(config)
	parsingResult.ProjectParser()
	t.Logf("项目解析完成，共解析 %d 个文件", len(parsingResult.Js_Data))

	// 调试：打印所有文件路径
	t.Log("解析到的文件路径:")
	for filePath := range parsingResult.Js_Data {
		t.Logf("  - %s", filePath)
	}

	// 2. 创建分析器并配置
	analyzer := &ExportCallAnalyzer{}
	err = analyzer.Configure(map[string]string{"manifest": manifestPath})
	require.NoError(t, err, "配置分析器失败")

	// 3. 执行分析
	ctx := &projectanalyzer.ProjectContext{
		ProjectRoot:   testProjectPath,
		Exclude:       []string{},
		IsMonorepo:    false,
		ParsingResult: parsingResult,
	}

	result, err := analyzer.Analyze(ctx)
	require.NoError(t, err, "分析执行失败")

	// 4. 验证结果
	exportCallResult, ok := result.(*ExportCallResult)
	require.True(t, ok, "结果类型错误")

	// 打印结果用于调试
	t.Logf("\n%s", exportCallResult.ToConsole())

	// 5. 验证模块导出
	assert.Equal(t, 3, len(exportCallResult.ModuleExports), "应该有 3 个 function 模块")

	// 6. 验证具体的导出节点
	// 查找 utils 相关的导出
	var utilsFormatRecord *FileExportRecord
	for _, module := range exportCallResult.ModuleExports {
		for i := range module.Files {
			if strings.Contains(module.Files[i].File, "utils") &&
				strings.Contains(module.Files[i].File, "format.ts") {
				utilsFormatRecord = &module.Files[i]
				break
			}
		}
		if utilsFormatRecord != nil {
			break
		}
	}

	require.NotNil(t, utilsFormatRecord, "应该找到 utils/format.ts 的导出记录")

	// 调试：打印所有节点信息
	t.Logf("format.ts 导出节点:")
	for i, node := range utilsFormatRecord.Nodes {
		t.Logf("  [%d] %s - %s, %s", i, node.Name, node.NodeType, node.ExportType)
	}

	// 验证 format.ts 中的导出节点
	nodeNames := make([]string, 0, len(utilsFormatRecord.Nodes))
	for _, node := range utilsFormatRecord.Nodes {
		nodeNames = append(nodeNames, node.Name)
	}

	// format.ts 导出了: formatDate, formatCurrency, formatNumber, truncate
	expectedExports := []string{"formatDate", "formatCurrency", "formatNumber", "truncate"}
	sort.Strings(nodeNames)
	sort.Strings(expectedExports)
	assert.Equal(t, expectedExports, nodeNames, "format.ts 应该导出预期的函数")

	// 验证节点类型
	// export const foo = () => {} 会被识别为 variable 类型（这是正确的）
	for _, node := range utilsFormatRecord.Nodes {
		assert.Equal(t, NodeTypeVariable, node.NodeType, "const 箭头函数应该是 variable 类型")
		assert.Equal(t, ExportTypeNamed, node.ExportType, "所有导出应该是 named 类型")
	}

	// 7. 验证引用关系
	// 注意：re-export (export * from './format') 不会建立引用关系
	// 只有真正的 import 才会建立引用关系
	// 由于当前测试项目中没有文件直接 import format.ts 的内容，所以没有引用
	// 这是一个预期的结果
}

// TestExportCallIntegration_Hooks 测试 hooks 模块的导出分析
func TestExportCallIntegration_Hooks(t *testing.T) {
	testProjectPath, err := filepath.Abs("../../../testdata/test_project")
	require.NoError(t, err)

	manifestPath := filepath.Join(testProjectPath, ".analyzer", "component-manifest.json")

	// 解析项目
	config := projectParser.NewProjectParserConfig(testProjectPath, []string{}, false, []string{})
	parsingResult := projectParser.NewProjectParserResult(config)
	parsingResult.ProjectParser()

	// 创建分析器
	analyzer := &ExportCallAnalyzer{}
	err = analyzer.Configure(map[string]string{"manifest": manifestPath})
	require.NoError(t, err)

	ctx := &projectanalyzer.ProjectContext{
		ProjectRoot:   testProjectPath,
		ParsingResult: parsingResult,
	}

	result, err := analyzer.Analyze(ctx)
	require.NoError(t, err)

	exportCallResult, ok := result.(*ExportCallResult)
	require.True(t, ok)

	// 查找 hooks/useDebounce.ts
	var debounceRecord *FileExportRecord
	for _, module := range exportCallResult.ModuleExports {
		for i := range module.Files {
			if strings.Contains(module.Files[i].File, "hooks") &&
				strings.Contains(module.Files[i].File, "useDebounce.ts") {
				debounceRecord = &module.Files[i]
				break
			}
		}
		if debounceRecord != nil {
			break
		}
	}

	require.NotNil(t, debounceRecord, "应该找到 hooks/useDebounce.ts 的导出记录")

	// useDebounce.ts 导出了 useDebounce 函数和 UseDebounceOptions 接口
	nodeMap := make(map[string]NodeWithRefs)
	for _, node := range debounceRecord.Nodes {
		nodeMap[node.Name] = node
	}

	// 验证函数导出
	// useDebounce 是 const 声明的箭头函数，所以是 variable 类型
	if debounceFn, ok := nodeMap["useDebounce"]; ok {
		assert.Equal(t, NodeTypeVariable, debounceFn.NodeType, "const 箭头函数应该是 variable 类型")
		assert.Equal(t, ExportTypeNamed, debounceFn.ExportType)
	} else {
		t.Error("应该找到 useDebounce 导出")
	}

	// 验证类型导出
	if optionsType, ok := nodeMap["UseDebounceOptions"]; ok {
		assert.Equal(t, NodeTypeInterface, optionsType.NodeType)
		assert.Equal(t, ExportTypeNamed, optionsType.ExportType)
	} else {
		t.Error("应该找到 UseDebounceOptions 导出")
	}
}

// TestExportCallIntegration_Types 测试 types 模块的导出分析
func TestExportCallIntegration_Types(t *testing.T) {
	testProjectPath, err := filepath.Abs("../../../testdata/test_project")
	require.NoError(t, err)

	manifestPath := filepath.Join(testProjectPath, ".analyzer", "component-manifest.json")

	// 解析项目
	config := projectParser.NewProjectParserConfig(testProjectPath, []string{}, false, []string{})
	parsingResult := projectParser.NewProjectParserResult(config)
	parsingResult.ProjectParser()

	// 创建分析器
	analyzer := &ExportCallAnalyzer{}
	err = analyzer.Configure(map[string]string{"manifest": manifestPath})
	require.NoError(t, err)

	ctx := &projectanalyzer.ProjectContext{
		ProjectRoot:   testProjectPath,
		ParsingResult: parsingResult,
	}

	result, err := analyzer.Analyze(ctx)
	require.NoError(t, err)

	exportCallResult, ok := result.(*ExportCallResult)
	require.True(t, ok)

	// 验证 types 模块的导出
	var typeCount int
	var interfaceCount int
	var enumCount int

	for _, module := range exportCallResult.ModuleExports {
		for _, record := range module.Files {
			if strings.Contains(record.File, "src/types/") {
				for _, node := range record.Nodes {
					switch node.NodeType {
					case NodeTypeType:
						typeCount++
					case NodeTypeInterface:
						interfaceCount++
					case NodeTypeEnum:
						enumCount++
					}
				}
			}
		}
	}

	// 验证至少有类型导出
	assert.Greater(t, typeCount+interfaceCount+enumCount, 0, "types 模块应该有类型导出")
	t.Logf("找到 %d 个 type 别名, %d 个 interface, %d 个 enum", typeCount, interfaceCount, enumCount)
}

// TestExportCallIntegration_DefaultExport 测试 default export 的分析
func TestExportCallIntegration_DefaultExport(t *testing.T) {
	testProjectPath, err := filepath.Abs("../../../testdata/test_project")
	require.NoError(t, err)

	manifestPath := filepath.Join(testProjectPath, ".analyzer", "component-manifest.json")

	// 解析项目
	config := projectParser.NewProjectParserConfig(testProjectPath, []string{}, false, []string{})
	parsingResult := projectParser.NewProjectParserResult(config)
	parsingResult.ProjectParser()

	// 创建分析器
	analyzer := &ExportCallAnalyzer{}
	err = analyzer.Configure(map[string]string{"manifest": manifestPath})
	require.NoError(t, err)

	ctx := &projectanalyzer.ProjectContext{
		ProjectRoot:   testProjectPath,
		ParsingResult: parsingResult,
	}

	result, err := analyzer.Analyze(ctx)
	require.NoError(t, err)

	exportCallResult, ok := result.(*ExportCallResult)
	require.True(t, ok)

	// 1. 验证具名 default export (useCounter)
	var useCounterRecord *FileExportRecord
	for _, module := range exportCallResult.ModuleExports {
		if module.ModuleName == "hooks" {
			for i := range module.Files {
				if strings.Contains(module.Files[i].File, "useCounter.ts") {
					useCounterRecord = &module.Files[i]
					break
				}
			}
		}
		if useCounterRecord != nil {
			break
		}
	}

	// 调试：打印 hooks 模块的所有文件
	t.Log("=== Hooks 模块的所有文件 ===")
	for _, module := range exportCallResult.ModuleExports {
		if module.ModuleName == "hooks" {
			t.Logf("模块: %s, 路径: %s", module.ModuleName, module.Path)
			for _, f := range module.Files {
				t.Logf("  文件: %s", f.File)
				for _, n := range f.Nodes {
					t.Logf("    - %s [%s, %s]", n.Name, n.NodeType, n.ExportType)
				}
			}
		}
	}

	require.NotNil(t, useCounterRecord, "应该找到 hooks/useCounter.ts 的导出记录")

	// 验证 default export 的 name 是原始名称 "useCounter"，而不是 "default"
	var defaultNode *NodeWithRefs
	for _, node := range useCounterRecord.Nodes {
		if node.ExportType == ExportTypeDefault {
			defaultNode = &node
			break
		}
	}

	require.NotNil(t, defaultNode, "useCounter.ts 应该有 default export")
	assert.Equal(t, "useCounter", defaultNode.Name, "default export 的 name 应该是原始名称，不是 'default'")
	assert.Equal(t, NodeTypeFunction, defaultNode.NodeType, "useCounter 应该是 function 类型")
	t.Logf("✅ default export name 正确: '%s'", defaultNode.Name)

	// 2. 验证具名 default export 有引用关系
	if defaultNode.RefFiles != nil && len(defaultNode.RefFiles) > 0 {
		t.Logf("✅ default export '%s' 有 %d 个引用:", defaultNode.Name, len(defaultNode.RefFiles))
		for _, ref := range defaultNode.RefFiles {
			t.Logf("   - %s", ref)
		}
		// 验证引用来自 Counter 组件
		foundCounterRef := false
		for _, ref := range defaultNode.RefFiles {
			if strings.Contains(ref, "Counter") {
				foundCounterRef = true
				break
			}
		}
		assert.True(t, foundCounterRef, "应该有来自 Counter 组件的引用")
	} else {
		t.Error("default export 'useCounter' 应该有引用关系")
	}

	// 3. 验证匿名 default export (anonymousDefault.ts)
	var anonymousRecord *FileExportRecord
	for _, module := range exportCallResult.ModuleExports {
		if module.ModuleName == "utils" {
			for i := range module.Files {
				if strings.Contains(module.Files[i].File, "anonymousDefault.ts") {
					anonymousRecord = &module.Files[i]
					break
				}
			}
		}
		if anonymousRecord != nil {
			break
		}
	}

	require.NotNil(t, anonymousRecord, "应该找到 utils/anonymousDefault.ts 的导出记录")

	// 验证匿名 default export 的 name 是 "default"
	var anonymousDefaultNode *NodeWithRefs
	for _, node := range anonymousRecord.Nodes {
		if node.ExportType == ExportTypeDefault {
			anonymousDefaultNode = &node
			break
		}
	}

	require.NotNil(t, anonymousDefaultNode, "anonymousDefault.ts 应该有 default export")
	assert.Equal(t, "default", anonymousDefaultNode.Name, "匿名 default export 的 name 应该是 'default'")
	t.Logf("✅ 匿名 default export name 正确: '%s'", anonymousDefaultNode.Name)

	// 4. 验证 named export 也被正确识别
	var namedNode *NodeWithRefs
	for _, node := range anonymousRecord.Nodes {
		if node.Name == "namedExport" && node.ExportType == ExportTypeNamed {
			namedNode = &node
			break
		}
	}

	require.NotNil(t, namedNode, "anonymousDefault.ts 应该有 namedExport")
	assert.Equal(t, ExportTypeNamed, namedNode.ExportType, "namedExport 应该是 named 类型")
	t.Logf("✅ named export 正确: '%s', type: %s", namedNode.Name, namedNode.ExportType)
}

// TestExportCallIntegration_ReExportTracking 测试重导出追踪功能
// 验证 export * from './path' 和 export { xxx } from './path' 会被正确追踪
func TestExportCallIntegration_ReExportTracking(t *testing.T) {
	testProjectPath, err := filepath.Abs("../../../testdata/test_project")
	require.NoError(t, err)

	manifestPath := filepath.Join(testProjectPath, ".analyzer", "component-manifest.json")

	// 解析项目
	config := projectParser.NewProjectParserConfig(testProjectPath, []string{}, false, []string{})
	parsingResult := projectParser.NewProjectParserResult(config)
	parsingResult.ProjectParser()

	// 创建分析器
	analyzer := &ExportCallAnalyzer{}
	err = analyzer.Configure(map[string]string{"manifest": manifestPath})
	require.NoError(t, err)

	ctx := &projectanalyzer.ProjectContext{
		ProjectRoot:   testProjectPath,
		ParsingResult: parsingResult,
	}

	result, err := analyzer.Analyze(ctx)
	require.NoError(t, err)

	exportCallResult, ok := result.(*ExportCallResult)
	require.True(t, ok)

	// utils/index.ts 有重导出:
	//   export * from './validation'
	//   export * from './format'
	// 验证这些重导出的内容被正确追踪

	// 查找 utils 模块
	var utilsModule *ModuleExportRecord
	for _, module := range exportCallResult.ModuleExports {
		if module.ModuleName == "utils" {
			utilsModule = &module
			break
		}
	}
	require.NotNil(t, utilsModule, "应该找到 utils 模块")

	// 验证 format.ts 的导出被正确追踪（通过 index.ts 的 export *）
	var formatRecord *FileExportRecord
	for i := range utilsModule.Files {
		if strings.Contains(utilsModule.Files[i].File, "format.ts") {
			formatRecord = &utilsModule.Files[i]
			break
		}
	}
	require.NotNil(t, formatRecord, "应该找到 format.ts 的导出记录")

	// format.ts 的导出节点应该被扫描到（通过 index.ts 的重导出）
	nodeNames := make([]string, 0, len(formatRecord.Nodes))
	for _, node := range formatRecord.Nodes {
		nodeNames = append(nodeNames, node.Name)
	}
	sort.Strings(nodeNames)

	// format.ts 导出的函数
	expectedExports := []string{"formatDate", "formatCurrency", "formatNumber", "truncate"}
	sort.Strings(expectedExports)
	assert.Equal(t, expectedExports, nodeNames, "format.ts 的导出应该被正确追踪（通过重导出）")

	// 验证 validation.ts 的导出也被正确追踪
	var validationRecord *FileExportRecord
	for i := range utilsModule.Files {
		if strings.Contains(utilsModule.Files[i].File, "validation.ts") {
			validationRecord = &utilsModule.Files[i]
			break
		}
	}
	require.NotNil(t, validationRecord, "应该找到 validation.ts 的导出记录")

	// validation.ts 导出的函数和类型
	expectedValidationExports := []string{
		"validateField",
		"validateEmail",
		"validatePhone",
		"ValidationResult",
		"ValidationRule",
	}
	sort.Strings(expectedValidationExports)

	validationNodeNames := make([]string, 0, len(validationRecord.Nodes))
	for _, node := range validationRecord.Nodes {
		validationNodeNames = append(validationNodeNames, node.Name)
	}
	sort.Strings(validationNodeNames)

	assert.Equal(t, expectedValidationExports, validationNodeNames, "validation.ts 的导出应该被正确追踪（通过重导出）")

	t.Logf("✅ 重导出追踪功能正常工作！")
	t.Logf("  - format.ts: 找到 %d 个导出", len(formatRecord.Nodes))
	t.Logf("  - validation.ts: 找到 %d 个导出", len(validationRecord.Nodes))
}
