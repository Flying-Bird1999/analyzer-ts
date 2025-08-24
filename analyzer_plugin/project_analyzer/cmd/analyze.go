package cmd

// example: go run main.go analyze unconsumed find-unreferenced-files count-any npm-check structure-simple -i /Users/bird/company/sc1.0/live/shopline-live-sale -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin -x "node_modules/**" -x "bffApiDoc/**"

// example: go run main.go analyze find-callers -i /Users/bird/company/sc1.0/live/shopline-live-sale -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin -x "node_modules/**" -x "bffApiDoc/**" -p "find-callers.targetFiles=/Users/bird/company/sc1.0/live/shopline-live-sale/src/feature/ActivityPage/index.tsx" -p "find-callers.targetFiles=/Users/bird/company/sc1.0/live/shopline-live-sale/src/feature/SettingPage/index.tsx"

import (
	"fmt"
	"os"
	"strings"

	projectanalyzer "main/analyzer_plugin/project_analyzer"
	"main/analyzer_plugin/project_analyzer/callgraph"
	countany "main/analyzer_plugin/project_analyzer/countAny"
	"main/analyzer_plugin/project_analyzer/dependency"
	"main/analyzer_plugin/project_analyzer/internal/filenamer"
	"main/analyzer_plugin/project_analyzer/internal/parser"
	"main/analyzer_plugin/project_analyzer/internal/writer"
	structuresimple "main/analyzer_plugin/project_analyzer/structureSimple"
	"main/analyzer_plugin/project_analyzer/unconsumed"
	"main/analyzer_plugin/project_analyzer/unreferenced"

	"github.com/spf13/cobra"
)

// availableAnalyzers 定义了所有可用的分析器
var availableAnalyzers = map[string]projectanalyzer.Analyzer{
	"unconsumed":              &unconsumed.Finder{},
	"find-callers":            &callgraph.Finder{},
	"find-unreferenced-files": &unreferenced.Finder{},
	"count-any":               &countany.Counter{},
	"npm-check":               &dependency.Checker{},
	"structure-simple":        &structuresimple.StructureSimpleAnalyzer{},
}

// GetAnalyzeCmd 返回分析命令的 Cobra 命令对象
func GetAnalyzeCmd() *cobra.Command {
	var (
		inputPath      string
		outputPath     string
		excludePath    []string
		isMonorepo     bool
		analyzerParams []string
	)

	analyzeCmd := &cobra.Command{
		Use:   "analyze [analyzer_name...]",
		Short: "对 TypeScript/JavaScript 项目进行代码分析。",
		Long: "该命令是分析器的主要入口点，能够对 TypeScript/JavaScript 项目进行深度分析.\n\n" +
			"您可以选择运行一个或多个内置的分析器，只需在命令后附上它们的名称即可.\n\n" +
			"可用分析器列表:\n" +
			"  - structure-simple: 输出一个简化的项目整体结构报告.\n" +
			"  - unconsumed: 查找项目中所有已导出但从未被导入的符号.\n" +
			"  - find-callers: 查找一个或多个指定文件的所有上游调用方.\n" +
			"  - find-unreferenced-files: 查找项目中从未被任何其他文件引用的\"孤岛\"文件.\n" +
			"  - count-any: 统计项目中所有 `any` 类型的使用情况.\n" +
			"  - npm-check: 检查 NPM 依赖，识别隐式、未使用和过期依赖.\n\n" +
			"如果未指定任何分析器，命令将仅解析项目并输出完整的、未经处理的原始 AST 结构.\n\n" +
			"参数 (-p, --param) 使用示例:\n" +
			"某些分析器需要额外的参数。例如，`find-callers` 需要知道要追踪哪个文件.\n" +
			"analyze find-callers -i . -p \"find-callers.targetFiles=src/utils.ts\" -p \"find-callers.targetFiles=src/helper.ts\"",
		Run: func(cmd *cobra.Command, args []string) {
			// 0. 如果未提供输出路径，则默认为当前工作目录
			if outputPath == "" {
				cwd, err := os.Getwd()
				if err != nil {
					fmt.Printf("错误: 无法获取当前工作目录: %v\n", err)
					os.Exit(1)
				}
				outputPath = cwd
			}

			// 1. 检查输入路径
			if inputPath == "" {
				fmt.Println("错误: 请使用 -i 或 --input 标志提供项目路径。")
				os.Exit(1)
			}

			// 2. 解析项目
			fmt.Println("开始解析项目，这可能需要一些时间...")
			parsingResult, err := parser.ParseProject(inputPath, excludePath, isMonorepo)
			if err != nil {
				fmt.Printf("错误: 解析项目失败: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("项目解析完成。")

			// 3. 如果没有指定分析器，则直接输出解析结果
			if len(args) == 0 {
				fmt.Println("\n未指定分析器，将直接输出项目解析结果。")
				outputFileName := filenamer.GenerateOutputFileName(inputPath, "analyzer_data")
				err := writer.WriteJSONResult(outputPath, outputFileName, parsingResult)
				if err != nil {
					fmt.Printf("错误: 无法将结果写入文件 %s: %v\n", outputPath, err)
					os.Exit(1)
				}
				fmt.Println("✅ 结果写入成功！")
				return // 完成执行
			}

			// 4. 如果指定了分析器，则继续执行分析流程
			// 4.1. 解析特定参数
			paramsForAnalyzers := parseAnalyzerParams(analyzerParams)

			// 4.2. 选择和配置分析器
			analyzersToRun := selectAnalyzers(args)
			configureAnalyzers(analyzersToRun, paramsForAnalyzers)

			// 4.3. 创建上下文
			ctx := &projectanalyzer.ProjectContext{
				ProjectRoot:   inputPath,
				Exclude:       excludePath,
				IsMonorepo:    isMonorepo,
				ParsingResult: parsingResult,
			}

			// 4.4. 执行分析器
			fmt.Printf("\n将在项目 %s 中运行 %d 个分析器...\n", ctx.ProjectRoot, len(analyzersToRun))
			allResults := executeAnalyzers(analyzersToRun, ctx)

			// 4.5. 处理结果
			handleResults(allResults, outputPath, inputPath)
		},
	}

	// 定义命令行标志
	analyzeCmd.Flags().StringVarP(&inputPath, "input", "i", "", "项目根目录")
	analyzeCmd.Flags().StringVarP(&outputPath, "output", "o", "", "输出文件目录 (默认为当前目录)")
	analyzeCmd.Flags().StringSliceVarP(&excludePath, "exclude", "x", []string{}, "排除的 glob 模式")
	analyzeCmd.Flags().BoolVarP(&isMonorepo, "monorepo", "m", false, "是否为 monorepo")
	analyzeCmd.Flags().StringSliceVarP(&analyzerParams, "param", "p", []string{}, "特定分析器参数 (e.g., 'analyzer.param=value')")
	analyzeCmd.MarkFlagRequired("input")
	return analyzeCmd
}

// parseAnalyzerParams 解析分析器参数，支持同一参数名的多个值
// 输入: ["find-callers.targetFiles=path1", "find-callers.targetFiles=path2"]
// 输出: {"find-callers": {"targetFiles": ["path1", "path2"]}}
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
		// 将参数值添加到切片中，而不是直接赋值，以支持同一参数名的多个值
		paramsForAnalyzers[analyzerName][paramName] = append(paramsForAnalyzers[analyzerName][paramName], parts[1])
	}
	return paramsForAnalyzers
}

// selectAnalyzers 根据名称选择要运行的分析器
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

// configureAnalyzers 配置选定的分析器
// 为了向后兼容，将 []string 转换为字符串（用逗号连接）
// 输入: map[分析器名称]map[参数名][]参数值
// 例如: {"find-callers": {"targetFiles": ["/path/to/file1.ts", "/path/to/file2.ts"]}}
// 输出: 传递给分析器的参数 map[参数名]参数值
// 例如: {"targetFiles": "/path/to/file1.ts,/path/to/file2.ts"}
func configureAnalyzers(analyzers []projectanalyzer.Analyzer, params map[string]map[string][]string) {
	for _, analyzer := range analyzers {
		if p, ok := params[analyzer.Name()]; ok {
			// 将 map[string][]string 转换为 map[string]string 以保持向后兼容
			// 对于有多个值的参数，我们用逗号将它们连接起来
			compatParams := make(map[string]string)
			for paramName, paramValues := range p {
				compatParams[paramName] = strings.Join(paramValues, ",")
			}

			if err := analyzer.Configure(compatParams); err != nil {
				fmt.Printf("错误: 配置分析器 '%s' 失败: %v\n", analyzer.Name(), err)
				os.Exit(1)
			}
		}
	}
}

// executeAnalyzers 执行所有选定的分析器
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

// handleResults 处理并写入所有分析结果
func handleResults(results map[string]projectanalyzer.Result, path string, inputPath string) {
	fmt.Printf("\n分析完成，正在将 %d 个分析结果写入 %s...\n", len(results), path)
	outputFileName := filenamer.GenerateOutputFileName(inputPath, "analyzer_data")
	err := writer.WriteJSONResult(path, outputFileName, results)
	if err != nil {
		fmt.Printf("错误: 无法将结果写入文件 %s: %v\n", path, err)
		os.Exit(1)
	}
	fmt.Println("✅ 结果写入成功！")
}
