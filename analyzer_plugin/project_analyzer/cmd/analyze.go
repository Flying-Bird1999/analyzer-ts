// package cmd 定义了所有子命令的实现。
package cmd

// example: go run main.go analyze unconsumed find-unreferenced-files count-any npm-check -i /Users/bird/company/sc1.0/live/shopline-live-sale -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin -x "node_modules/**" -x "bffApiDoc/**"

// example: go run main.go analyze find-callers -i /Users/bird/company/sc1.0/live/shopline-live-sale -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin -x "node_modules/**" -x "bffApiDoc/**" -p "find-callers.targetFiles=/Users/bird/company/sc1.0/live/shopline-live-sale/src/feature/ActivityPage/index.tsx" -p "find-callers.targetFiles=/Users/bird/company/sc1.0/live/shopline-live-sale/src/feature/SettingPage/index.tsx"

// example: go run main.go analyze trace -i /Users/bird/company/sc1.0/live/shopline-live-sale -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin -x "node_modules/**" -x "bffApiDoc/**" -p "trace.targetPkgs=antd" -p "trace.targetPkgs=@yy/sl-admin-components"

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
	apit "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/api_tracer"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/component_deps"
	countany "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/countAny"
	countas "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/countAs"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/dependency"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/trace"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/unconsumed"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/unreferenced"

	"github.com/spf13/cobra"
)

// availableAnalyzers 是一个中央注册表，用于存储所有可用的分析器。
// key 是用户在命令行中使用的分析器名称 (例如 "unconsumed")。
// value 是实现了 projectanalyzer.Analyzer 接口的分析器实例。
// 这种设计使得添加新的分析器变得非常简单，只需在此 map 中增加一个条目即可。
var availableAnalyzers = map[string]projectanalyzer.Analyzer{
	"unconsumed":              &unconsumed.Finder{},
	"find-unreferenced-files": &unreferenced.Finder{},
	"count-any":               &countany.Counter{},
	"count-as":                &countas.Counter{},
	"npm-check":               &dependency.Checker{},
	"trace":                   &trace.Tracer{},
	"api-tracer":              &apit.Tracer{},
	"component-deps":          &component_deps.ComponentDependencyAnalyzer{},
}

// GetAnalyzeCmd 构建并返回 `analyze` 命令。
// 这个命令是整个工具链的核心，负责执行一个或多个分析任务。
func GetAnalyzeCmd() *cobra.Command {
	// 使用变量来接收命令行标志的值
	var (
		inputPath      string
		outputPath     string
		excludePath    []string
		isMonorepo     bool
		analyzerParams []string
		stripFields    []string // 用于存储用户指定的、需要剔除的字段
	)

	analyzeCmd := &cobra.Command{
		Use:   "analyze [analyzer_name...]",
		Short: "对 TypeScript/JavaScript 项目进行代码分析。",
		Long: `该命令是分析器的主要入口点，能够对 TypeScript/JavaScript 项目进行深度分析。

` +
			`您可以选择运行一个或多个内置的分析器，只需在命令后附上它们的名称即可。

` +
			`数据预处理 (--strip-fields):
` +
			`在将解析数据传递给分析器之前，您可以使用 --strip-fields (-s) 标志来预先剔除不需要的字段。
` +
			`这对于简化输入、提升特定分析器的性能或聚焦于特定数据非常有用。
` +
			`  -s raw                 (剔除所有名为 'raw' 的字段)
` +
			`  -s sourceLocation      (剔除所有名为 'sourceLocation' 的字段)
` +
			`  -s "raw,sourceLocation"  (同时剔除两者)

` +
			`可用分析器列表:
` +
			`  - unconsumed: 查找项目中所有已导出但从未被导入的符号.
` +
			`  - find-callers: 查找一个或多个指定文件的所有上游调用方.
` +
			`  - find-unreferenced-files: 查找项目中从未被任何其他文件引用的"孤岛"文件.
` +
			`  - count-any: 统计项目中所有 'any' 类型的使用情况.
` +
			`  - count-as: 统计项目中所有 'as' 类型断言的使用情况.
` +
			`  - npm-check: 检查 NPM 依赖，识别隐式、未使用和过期依赖.
` +
			`  - trace: 追踪一个或多个NPM包的使用链路 (例如 antd).
` +
			`  - api-tracer: 追踪一个或多个接口的调用链路.
` +
			`  - component-deps: 分析组件之间的依赖关系. (必须使用 -p 'component-deps.entryPoint=path/to/entry.ts' 指定入口文件，支持 glob 模式)
` +
			`如果未指定任何分析器，命令将仅解析项目并输出完整的、未经处理的（但可能被剔除过的）原始AST结构.

` +
			`特定分析器参数 (-p, --param) 使用示例:
` +
			`'trace' 分析器需要 'trace.targetPkgs' 参数来指定要追踪的NPM包:
` +
			`'api-tracer' 分析器需要 'api-tracer.apiPaths' 参数来指定要追踪的接口:
` +
			`analyze trace -i . -p "trace.targetPkgs=antd" -p "trace.targetPkgs=@yy/sl-admin-components"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// --- 步骤 0: 初始化和验证路径参数 ---
			if outputPath == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("错误: 无法获取当前工作目录: %w", err)
				}
				outputPath = cwd
			}
			if inputPath == "" {
				return fmt.Errorf("错误: 请使用 -i 或 --input 标志提供项目路径")
			}

			// --- 步骤 1: 快速失败校验 ---
			// 在执行任何耗时操作之前，首先校验用户请求的分析器名称是否都存在。
			var analyzersToRun []projectanalyzer.Analyzer
			if len(args) > 0 {
				analyzersToRun = selectAnalyzers(args)
			}

			// --- 步骤 2: 执行核心解析逻辑 ---
			// 调用公共函数，该函数会负责项目解析以及根据 --strip-fields 参数进行预处理。
			parsingResult, err := ParseAndStripFields(inputPath, excludePath, isMonorepo, stripFields)
			if err != nil {
				return fmt.Errorf("错误: 解析或剔除字段失败: %w", err)
			}

			// --- 步骤 3: 处理不同执行场景 ---
			// 如果没有指定分析器，则直接输出预处理后的项目数据。
			if len(analyzersToRun) == 0 {
				fmt.Println("\n未指定分析器，将直接输出项目解析结果。")
				outputFileName := GenerateOutputFileName(inputPath, "analyzer_data")
				return WriteJSONResult(outputPath, outputFileName, parsingResult)
			}

			// 如果指定了分析器，则配置并执行它们。
			paramsForAnalyzers := parseAnalyzerParams(analyzerParams)
			configureAnalyzers(analyzersToRun, paramsForAnalyzers)

			// 为分析器创建执行上下文，传入可能已被裁剪过的解析结果。
			ctx := &projectanalyzer.ProjectContext{
				ProjectRoot:   inputPath,
				Exclude:       excludePath,
				IsMonorepo:    isMonorepo,
				ParsingResult: parsingResult,
			}

			fmt.Printf("\n将在项目 %s 中运行 %d 个分析器...\n", ctx.ProjectRoot, len(analyzersToRun))
			allResults := executeAnalyzers(analyzersToRun, ctx)
			// --- 步骤 4: 处理并输出最终结果 ---
			handleResults(allResults, outputPath, inputPath)
			return nil
		},
	}

	// --- Flag 定义区 ---
	// 为命令定义所有可接受的命令行标志。
	analyzeCmd.Flags().StringVarP(&inputPath, "input", "i", "", "项目根目录")
	analyzeCmd.Flags().StringVarP(&outputPath, "output", "o", "", "输出文件目录 (默认为当前目录)")
	analyzeCmd.Flags().StringSliceVarP(&excludePath, "exclude", "x", []string{}, "排除的 glob 模式")
	analyzeCmd.Flags().BoolVarP(&isMonorepo, "monorepo", "m", false, "是否为 monorepo")
	analyzeCmd.Flags().StringSliceVarP(&analyzerParams, "param", "p", []string{}, "为特定分析器传递参数 (例如 'trace.targetPkgs=antd')")
	analyzeCmd.Flags().StringSliceVarP(&stripFields, "strip-fields", "s", []string{}, "在分析前，从解析结果中递归删除的字段名或路径")
	analyzeCmd.MarkFlagRequired("input")
	return analyzeCmd
}

// parseAnalyzerParams 解析提供给特定分析器的参数。
// 它将 `-p "analyzer.key=value"` 格式的字符串转换为一个嵌套的 map，
// 以便 `configureAnalyzers` 函数可以轻松地为每个分析器查找其配置。
func parseAnalyzerParams(params []string) map[string]map[string][]string {
	paramsForAnalyzers := make(map[string]map[string][]string)
	for _, p := range params {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) != 2 {
			continue
		}
		keyParts := strings.SplitN(parts[0], ".", 2)
		if len(keyParts) != 2 {
			continue
		}
		analyzerName, paramName := keyParts[0], keyParts[1]
		if _, ok := paramsForAnalyzers[analyzerName]; !ok {
			paramsForAnalyzers[analyzerName] = make(map[string][]string)
		}
		// 支持多次使用同一参数 (例如 -p "A=B" -p "A=C")
		paramsForAnalyzers[analyzerName][paramName] = append(paramsForAnalyzers[analyzerName][paramName], parts[1])
	}
	return paramsForAnalyzers
}

// selectAnalyzers 根据用户在命令行中提供的名称，从 `availableAnalyzers` 注册表中查找并返回一个分析器列表。
func selectAnalyzers(args []string) []projectanalyzer.Analyzer {
	var analyzersToRun []projectanalyzer.Analyzer
	for _, name := range args {
		if analyzer, ok := availableAnalyzers[name]; ok {
			analyzersToRun = append(analyzersToRun, analyzer)
		} else {
			fmt.Printf("错误: 未知的分析器 '%s'\n", name)
			os.Exit(1)
		}
	}
	return analyzersToRun
}

// configureAnalyzers 遍历所有待运行的分析器，并调用它们的 Configure 方法。
// 它将从命令行解析出的、属于各个分析器的参数传递给它们，以完成初始化。
func configureAnalyzers(analyzers []projectanalyzer.Analyzer, params map[string]map[string][]string) {
	for _, analyzer := range analyzers {
		analyzerParams := make(map[string]string)
		// 检查是否存在当前分析器的参数
		if p, ok := params[analyzer.Name()]; ok {
			// 将多个相同key的参数值用逗号连接，以兼容需要接收列表的分析器（如 `trace`）。
			for paramName, paramValues := range p {
				analyzerParams[paramName] = strings.Join(paramValues, ",")
			}
		}
		// 调用分析器自己的配置方法，无论有无参数，都应调用，以便分析器自行处理默认值或报错。
		if err := analyzer.Configure(analyzerParams); err != nil {
			fmt.Printf("错误: 配置分析器 '%s' 失败: %v\n", analyzer.Name(), err)
			os.Exit(1)
		}
	}
}

// executeAnalyzers 遍历并执行所有已配置好的分析器。
// 它为每个分析器调用 Analyze 方法，并收集结果。
func executeAnalyzers(analyzers []projectanalyzer.Analyzer, ctx *projectanalyzer.ProjectContext) map[string]projectanalyzer.Result {
	allResults := make(map[string]projectanalyzer.Result)
	for _, analyzer := range analyzers {
		fmt.Printf("===== 正在运行: %s =====\n", analyzer.Name())
		res, err := analyzer.Analyze(ctx)
		if err != nil {
			fmt.Printf("分析器 '%s' 执行失败: %v\n\n", analyzer.Name(), err)
			continue
		}
		allResults[analyzer.Name()] = res
	}
	return allResults
}

// handleResults 将所有分析器的结果合并到一个map中，并写入到最终的输出文件。
func handleResults(results map[string]projectanalyzer.Result, path string, inputPath string) {
	fmt.Printf("\n分析完成，正在将 %d 个分析结果写入 %s...\n", len(results), path)
	outputFileName := GenerateOutputFileName(inputPath, "analyzer_data")
	err := WriteJSONResult(path, outputFileName, results)
	if err != nil {
		fmt.Printf("错误: 无法将结果写入文件 %s: %v\n", path, err)
		os.Exit(1)
	}
	fmt.Println("✅ 结果写入成功！")
}

// GenerateOutputFileName 是一个公共函数，用于根据输入目录和分析类型生成标准化的输出文件名。
func GenerateOutputFileName(inputPath, suffix string) string {
	baseName := filepath.Base(inputPath)
	safeBaseName := strings.ReplaceAll(baseName, " ", "_")
	return fmt.Sprintf("%s_%s.json", safeBaseName, suffix)
}

// WriteJSONResult 是一个公共函数，用于将分析结果序列化为 JSON 并写入文件。
func WriteJSONResult(outputDir, fileName string, result interface{}) error {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化JSON时出错: %w", err)
	}
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}
	outputFile := filepath.Join(outputDir, fileName)
	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		return fmt.Errorf("写入JSON文件失败: %w", err)
	}
	fmt.Printf("结果已成功写入到 %s\n", outputFile)
	return nil
}
