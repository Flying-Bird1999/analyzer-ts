package cmd

// example: go run main.go analyze unconsumed find-unreferenced-files count-any npm-check structure-simple -i /Users/bird/company/sc1.0/live/shopline-live-sale -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin -x "node_modules/**" -x "bffApiDoc/**"

// example: go run main.go analyze find-callers -i /Users/bird/company/sc1.0/live/shopline-live-sale -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin -x "node_modules/**" -x "bffApiDoc/**" -p "find-callers.targetFiles=/Users/bird/company/sc1.0/live/shopline-live-sale/src/feature/ActivityPage/index.tsx" -p "find-callers.targetFiles=/Users/bird/company/sc1.0/live/shopline-live-sale/src/feature/SettingPage/index.tsx"

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/callgraph"
	countany "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/countAny"
	countas "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/countAs"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/dependency"
	structuresimple "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/structureSimple"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/unconsumed"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/unreferenced"

	"github.com/spf13/cobra"
)

// availableAnalyzers 定义了所有可用的分析器
var availableAnalyzers = map[string]projectanalyzer.Analyzer{
	"unconsumed":              &unconsumed.Finder{},
	"find-callers":            &callgraph.Finder{},
	"find-unreferenced-files": &unreferenced.Finder{},
	"count-any":               &countany.Counter{},
	"count-as":                &countas.Counter{},
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
			"  - count-as: 统计项目中所有 `as` 类型断言的使用情况.\n" +
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
			parsingResult, err := ParseProject(inputPath, excludePath, isMonorepo)
			if err != nil {
				fmt.Printf("错误: 解析项目失败: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("项目解析完成。")

			// 打印解析错误
			var totalErrors int
			for filePath, fileResult := range parsingResult.Js_Data {
				if len(fileResult.Errors) > 0 {
					fmt.Fprintf(os.Stderr, "\n--- Errors in %s ---\n", filePath)
					for _, err := range fileResult.Errors {
						fmt.Fprintf(os.Stderr, "- %v\n", err)
					}
					totalErrors += len(fileResult.Errors)
				}
			}
			if totalErrors > 0 {
				fmt.Fprintf(os.Stderr, "\nFound a total of %d parsing errors.\n", totalErrors)
			}

			// 3. 如果没有指定分析器，则直接输出解析结果
			if len(args) == 0 {
				fmt.Println("\n未指定分析器，将直接输出项目解析结果。")
				outputFileName := GenerateOutputFileName(inputPath, "analyzer_data")
				err := WriteJSONResult(outputPath, outputFileName, parsingResult)
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
	outputFileName := GenerateOutputFileName(inputPath, "analyzer_data")
	err := WriteJSONResult(path, outputFileName, results)
	if err != nil {
		fmt.Printf("错误: 无法将结果写入文件 %s: %v\n", path, err)
		os.Exit(1)
	}
	fmt.Println("✅ 结果写入成功！")
}

// GenerateOutputFileName 是一个公共函数，用于根据输入目录和分析类型生成标准化的输出文件名。
// 它封装了文件名生成的通用逻辑，供各个分析器调用，确保输出文件名的一致性。
// 参数:
//   - inputPath: 输入项目目录的路径。
//   - suffix:    - string: 生成的标准化文件名（不包含路径）。
func GenerateOutputFileName(inputPath, suffix string) string {
	// 1. 获取输入目录的基名（例如，"/path/to/my-project" -> "my-project"）。
	baseName := filepath.Base(inputPath)

	// 2. 将基名中的空格替换为下划线，以确保文件名的有效性。
	//    虽然不太常见，但 defensive programming 是好的。
	safeBaseName := strings.ReplaceAll(baseName, " ", "_")

	// 3. 拼接基名和后缀，并加上 .json 扩展名。
	//    例如，"my-project" 和 "analyze" 会生成 "my-project_analyze.json"。
	return fmt.Sprintf("%s_%s.json", safeBaseName, suffix)
}

// ParseProject 是一个公共函数，用于解析整个项目。
// 它封装了项目解析的通用逻辑，供各个分析器调用，避免代码重复。
// 参数:
//   - rootPath: 项目根目录的绝对路径。
//   - ignore: 需要从分析中排除的文件/目录的 glob 模式列表。
//   - isMonorepo: 指示项目是否为 monorepo。
//
// 返回值:
//   - *projectParser.ProjectParserResult: 解析后的项目结果。
//   - error: 解析过程中可能发生的错误。
func ParseProject(rootPath string, ignore []string, isMonorepo bool) (*projectParser.ProjectParserResult, error) {
	// 步骤 1: 创建项目解析器配置。
	config := projectParser.NewProjectParserConfig(rootPath, ignore, isMonorepo, []string{})

	// 步骤 2: 创建用于存储解析结果的容器。
	ar := projectParser.NewProjectParserResult(config)

	// 步骤 3: 运行主解析逻辑。
	ar.ProjectParser()

	// 在未来的版本中，这里可以增加更详细的错误处理逻辑。
	// 目前，ProjectParser 方法内部的错误处理比较简单。
	return ar, nil
}

// WriteJSONResult 是一个公共函数，用于将分析结果序列化为 JSON 并写入文件。
// 它封装了 JSON 序列化和文件写入的通用逻辑，供各个分析器调用，避免代码重复。
// 参数:
//   - outputDir: 用于存储 JSON 输出文件的目录路径。
//   - fileName: 输出文件的名称（不包含路径）。
//   - result:   - error: 写入过程中可能发生的错误。
func WriteJSONResult(outputDir, fileName string, result interface{}) error {
	// 步骤 1: 将结果序列化为格式化的 JSON 字节。
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化JSON时出错: %w", err)
	}

	// 步骤 2: 确保输出目录存在。
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 步骤 3: 构建完整的输出文件路径。
	outputFile := filepath.Join(outputDir, fileName)

	// 步骤 4: 将 JSON 数据写入文件。
	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		return fmt.Errorf("写入JSON文件失败: %w", err)
	}

	// 步骤 5: 打印成功信息到标准输出。
	fmt.Printf("结果已成功写入到 %s\n", outputFile)
	return nil
}
