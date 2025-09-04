// package cmd 定义了所有命令
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/spf13/cobra"
)

// GetTraceCmd 返回 'trace' 子命令
func GetTraceCmd() *cobra.Command {
	var (
		inputPath  string
		outputPath string
		exclude    []string
		isMonorepo bool
		pkgsStr    string
	)

	traceCmd := &cobra.Command{
		Use:   "trace",
		Short: "追踪项目中特定NPM包的使用链路，并输出过滤后的项目解析数据。",
		Long: `'trace' 命令会深度分析一个项目，并根据指定的NPM包列表，追踪其在代码中的完整使用链路.\n\n` +
			`它会识别从目标包的导入开始，经过变量传递、解构，直到最终在JSX或函数调用中的使用.\n` +
			`最终的输出是一个与原始解析结果结构完全相同但内容经过严格过滤的JSON，只包含构成使用链路的节点。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// --- 1. 解析参数 ---
			targetPkgs := make(map[string]struct{})
			if pkgsStr == "" {
				// 使用默认包列表
				targetPkgs = map[string]struct{}{
					"@yy/sl-admin-components": {},
					"@sl/admin-components":    {},
					"antd":                    {},
				}
			} else {
				for _, pkg := range strings.Split(pkgsStr, ",") {
					trimmed := strings.TrimSpace(pkg)
					if trimmed != "" {
						targetPkgs[trimmed] = struct{}{}
					}
				}
			}
			if len(targetPkgs) == 0 {
				return errors.New("目标NPM包列表为空")
			}

			// --- 2. 执行完整项目解析 ---
			fmt.Fprintln(os.Stderr, "开始完整解析项目...")
			config := projectParser.NewProjectParserConfig(inputPath, exclude, isMonorepo, []string{})
			parsingResult := projectParser.NewProjectParserResult(config)
			parsingResult.ProjectParser()
			fmt.Fprintln(os.Stderr, "项目解析完成。")

			// --- 3. 污点分析 ---
			fmt.Fprintln(os.Stderr, "开始追踪使用链路...")
			taintedSymbols := performTaintAnalysis(parsingResult, targetPkgs)
			fmt.Fprintf(os.Stderr, "追踪完成，共发现 %d 个相关符号。\n", len(taintedSymbols))

			// --- 4. 构建过滤后的结果 ---
			fmt.Fprintln(os.Stderr, "正在生成过滤后的结果...")
			filteredResult := buildFilteredResult(parsingResult, taintedSymbols, targetPkgs)

			// --- 5. 输出JSON ---
			if outputPath != "" {
				// 写入到文件，并进行编码处理
				fileName := fmt.Sprintf("%s_trace_result.json", filepath.Base(inputPath))
				filePath := filepath.Join(outputPath, fileName)

				if err := os.MkdirAll(outputPath, os.ModePerm); err != nil {
					return fmt.Errorf("创建输出目录失败: %w", err)
				}

				file, err := os.Create(filePath)
				if err != nil {
					return fmt.Errorf("创建文件失败: %w", err)
				}
				defer file.Close()

				encoder := json.NewEncoder(file)
				encoder.SetEscapeHTML(false) // 防止将 <, >, & 等符号转义
				encoder.SetIndent("", "  ")  // 设置缩进，实现 pretty-print

				if err := encoder.Encode(filteredResult); err != nil {
					return fmt.Errorf("编码并写入JSON失败: %w", err)
				}
				fmt.Fprintf(os.Stderr, "✅ 结果已成功写入到 %s\n", filePath)
			} else {
				// 输出到标准输出
				outputJSON, err := json.MarshalIndent(filteredResult, "", "  ")
				if err != nil {
					return fmt.Errorf("序列化最终结果失败: %w", err)
				}
				fmt.Println(string(outputJSON))
			}

			return nil
		},
	}

	traceCmd.Flags().StringVarP(&inputPath, "input", "i", "", "要分析的项目根目录 (必需)")
	traceCmd.Flags().StringVarP(&outputPath, "output", "o", "", "输出文件目录 (默认为标准输出)")
	traceCmd.Flags().StringSliceVarP(&exclude, "exclude", "x", []string{}, "要排除的目录或文件的 glob 模式 (可多次使用)")
	traceCmd.Flags().BoolVarP(&isMonorepo, "monorepo", "m", false, "是否将项目作为 monorepo 进行解析")
	traceCmd.Flags().StringVarP(&pkgsStr, "target-pkgs", "p", "", "要追踪的NPM包名列表，用逗号分隔 (默认为 antd, @sl/* 等)")
	traceCmd.MarkFlagRequired("input")

	return traceCmd
}

// performTaintAnalysis 执行污点分析，找出所有与目标包相关的符号
func performTaintAnalysis(pr *projectParser.ProjectParserResult, targetPkgs map[string]struct{}) map[string]string {
	taintedSymbols := make(map[string]string) // key: filePath#symbolName, value: npmPackage

	// 阶段 1: 识别污染源
	for filePath, fileData := range pr.Js_Data {
		for _, imp := range fileData.ImportDeclarations {
			if _, isTarget := targetPkgs[imp.Source.NpmPkg]; isTarget {
				for _, mod := range imp.ImportModules {
					key := fmt.Sprintf("%s#%s", filePath, mod.Identifier)
					taintedSymbols[key] = imp.Source.NpmPkg
				}
			}
		}
	}

	// 阶段 2: 传播污染
	for {
		newlyTainted := false
		for filePath, fileData := range pr.Js_Data {
			for _, varDecl := range fileData.VariableDeclarations {
				sourceSymbol, _ := getSourceSymbolFromVarDecl(&varDecl)
				if sourceSymbol == "" {
					continue
				}
				sourceKey := fmt.Sprintf("%s#%s", filePath, sourceSymbol)
				if npmPkg, isTainted := taintedSymbols[sourceKey]; isTainted {
					for _, declarator := range varDecl.Declarators {
						newSymbolKey := fmt.Sprintf("%s#%s", filePath, declarator.Identifier)
						if _, alreadyTainted := taintedSymbols[newSymbolKey]; !alreadyTainted {
							taintedSymbols[newSymbolKey] = npmPkg
							newlyTainted = true
						}
					}
				}
			}
		}
		if !newlyTainted {
			break
		}
	}
	return taintedSymbols
}

// buildFilteredResult 根据污点分析的结果，动态构建一个只包含相关节点的map，用于最终的JSON输出。
func buildFilteredResult(pr *projectParser.ProjectParserResult, taintedSymbols map[string]string, targetPkgs map[string]struct{}) map[string]interface{} {
	filteredJsData := make(map[string]interface{})

	for filePath, fileData := range pr.Js_Data {
		// 每个文件一个map，只添加非空字段
		filteredFileData := make(map[string]interface{})

		// 过滤 Imports
		var relevantImports []projectParser.ImportDeclarationResult
		for _, imp := range fileData.ImportDeclarations {
			if _, isTarget := targetPkgs[imp.Source.NpmPkg]; isTarget {
				relevantImports = append(relevantImports, imp)
			}
		}
		if len(relevantImports) > 0 {
			filteredFileData["importDeclarations"] = relevantImports
		}

		// 过滤 Variable Declarations
		var relevantVars []parser.VariableDeclaration
		for _, varDecl := range fileData.VariableDeclarations {
			sourceSymbol, _ := getSourceSymbolFromVarDecl(&varDecl)
			if sourceSymbol != "" {
				if _, isTainted := taintedSymbols[fmt.Sprintf("%s#%s", filePath, sourceSymbol)]; isTainted {
					relevantVars = append(relevantVars, varDecl)
				}
			}
		}
		if len(relevantVars) > 0 {
			filteredFileData["variableDeclarations"] = relevantVars
		}

		// 过滤 Jsx Elements
		var relevantJsx []projectParser.JSXElementResult
		for _, jsx := range fileData.JsxElements {
			if len(jsx.ComponentChain) > 0 {
				if _, isTainted := taintedSymbols[fmt.Sprintf("%s#%s", filePath, jsx.ComponentChain[0])]; isTainted {
					relevantJsx = append(relevantJsx, jsx)
				}
			}
		}
		if len(relevantJsx) > 0 {
			filteredFileData["jsxElements"] = relevantJsx
		}

		// 过滤 Call Expressions
		var relevantCalls []parser.CallExpression
		for _, call := range fileData.CallExpressions {
			if len(call.CallChain) > 0 {
				if _, isTainted := taintedSymbols[fmt.Sprintf("%s#%s", filePath, call.CallChain[0])]; isTainted {
					relevantCalls = append(relevantCalls, call)
				}
			}
		}
		if len(relevantCalls) > 0 {
			filteredFileData["callExpressions"] = relevantCalls
		}

		// 如果该文件包含任何相关节点，则将其添加到最终结果中
		if len(filteredFileData) > 0 {
			filteredJsData[filePath] = filteredFileData
		}
	}

	return filteredJsData
}

// getSourceSymbolFromVarDecl 从一个变量声明中提取其赋值的来源符号。
func getSourceSymbolFromVarDecl(varDecl *parser.VariableDeclaration) (string, *parser.VariableValue) {
	if varDecl.Source != nil && varDecl.Source.Type == "identifier" {
		return varDecl.Source.Expression, varDecl.Source
	}
	if len(varDecl.Declarators) == 1 && varDecl.Declarators[0].InitValue != nil {
		initVal := varDecl.Declarators[0].InitValue
		if initVal.Type == "identifier" {
			return initVal.Expression, initVal
		}
	}
	return "", nil
}
