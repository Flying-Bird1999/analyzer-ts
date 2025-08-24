package cmd

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
	"main/analyzer_plugin/project_analyzer/unconsumed"
	"main/analyzer_plugin/project_analyzer/unreferenced"

	"github.com/spf13/cobra"
)

var availableAnalyzers = map[string]projectanalyzer.Analyzer{
	"unconsumed":              &unconsumed.Finder{},
	"find-callers":            &callgraph.Finder{},
	"find-unreferenced-files": &unreferenced.Finder{},
	"count-any":               &countany.Counter{},
	"npm-check":               &dependency.Checker{},
}

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
		Short: "运行一个或多个代码分析器",
		Long:  `...`,
		Run: func(cmd *cobra.Command, args []string) {
			// 1. 解析特定参数
			paramsForAnalyzers := parseAnalyzerParams(analyzerParams)

			// 2. 选择和配置分析器
			analyzersToRun := selectAnalyzers(args)
			configureAnalyzers(analyzersToRun, paramsForAnalyzers)

			// 3. 创建初始上下文并执行一次性解析
			if inputPath == "" {
				fmt.Println("错误: 请使用 -i 或 --input 标志提供项目路径。")
				os.Exit(1)
			}
			fmt.Println("开始解析项目，这可能需要一些时间...")
			parsingResult, err := parser.ParseProject(inputPath, excludePath, isMonorepo)
			if err != nil {
				fmt.Printf("错误: 解析项目失败: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("项目解析完成。")

			ctx := &projectanalyzer.ProjectContext{
				ProjectRoot:   inputPath,
				Exclude:       excludePath,
				IsMonorepo:    isMonorepo,
				ParsingResult: parsingResult,
			}

			// 4. 执行分析器
			fmt.Printf("\n将在项目 %s 中运行 %d 个分析器...\n", ctx.ProjectRoot, len(analyzersToRun))
			allResults := executeAnalyzers(analyzersToRun, ctx)

			// 5. 处理结果
			handleResults(allResults, outputPath, inputPath)
		},
	}

	analyzeCmd.Flags().StringVarP(&inputPath, "input", "i", "", "项目根目录")
	analyzeCmd.Flags().StringVarP(&outputPath, "output", "o", "", "输出文件目录")
	analyzeCmd.Flags().StringSliceVarP(&excludePath, "exclude", "x", []string{}, "排除的 glob 模式")
	analyzeCmd.Flags().BoolVarP(&isMonorepo, "monorepo", "m", false, "是否为 monorepo")
	analyzeCmd.Flags().StringSliceVarP(&analyzerParams, "param", "p", []string{}, "特定分析器参数 (e.g., 'analyzer.param=value')")
	analyzeCmd.MarkFlagRequired("input")
	return analyzeCmd
}

func parseAnalyzerParams(params []string) map[string]map[string]string {
	paramsForAnalyzers := make(map[string]map[string]string)
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
			paramsForAnalyzers[analyzerName] = make(map[string]string)
		}
		paramsForAnalyzers[analyzerName][paramName] = parts[1]
	}
	return paramsForAnalyzers
}

func selectAnalyzers(args []string) []projectanalyzer.Analyzer {
	var analyzersToRun []projectanalyzer.Analyzer
	if len(args) > 0 {
		for _, name := range args {
			if analyzer, ok := availableAnalyzers[name]; ok {
				analyzersToRun = append(analyzersToRun, analyzer)
			} else {
				fmt.Printf("错误: 未知的分析器 '%s'\n", name)
				os.Exit(1)
			}
		}
	} else {
		for _, analyzer := range availableAnalyzers {
			analyzersToRun = append(analyzersToRun, analyzer)
		}
	}
	return analyzersToRun
}

func configureAnalyzers(analyzers []projectanalyzer.Analyzer, params map[string]map[string]string) {
	for _, analyzer := range analyzers {
		if p, ok := params[analyzer.Name()]; ok {
			if err := analyzer.Configure(p); err != nil {
				fmt.Printf("错误: 配置分析器 '%s' 失败: %v\n", analyzer.Name(), err)
				os.Exit(1)
			}
		}
	}
}

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

func handleResults(results map[string]projectanalyzer.Result, path string, inputPath string) {
	if path != "" {
		fmt.Printf("\n分析完成，正在将 %d 个分析结果写入 %s...\n", len(results), path)
		outputFileName := filenamer.GenerateOutputFileName(inputPath, "analyzer_data")
		err := writer.WriteJSONResult(path, outputFileName, results)
		if err != nil {
			fmt.Printf("错误: 无法将结果写入文件 %s: %v\n", path, err)
			os.Exit(1)
		}
		fmt.Println("✅ 结果写入成功！")
	} else {
		fmt.Println("\n--- 分析完成，正在打印结果 ---")
		for _, res := range results {
			fmt.Printf("\n===== 结果来自: %s =====\n", res.Name())
			fmt.Println(res.ToConsole())
		}
	}
}
