// package cmd 定义了分析器的所有命令行接口。
// 本文件 (query.go) 实现了一个最终版的、高度灵活的一站式数据查询和重塑命令。
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/jmespath/go-jmespath"
	"github.com/spf13/cobra"
)

// GetQueryCmd 返回 'query' 子命令的 Cobra 命令对象。
// 这是该命令的最终实现，提供了最大的灵活性。
func GetQueryCmd() *cobra.Command {
	var queryCmd = &cobra.Command{
		Use:   "query -i <project-path> [flags]",
		Short: "一站式地分析项目，并可选地使用字段名/路径和 JMESPath 进行精确的数据提取与重塑。",
		Long: `'query' 命令是一个强大的一站式工具，它按以下顺序执行操作：

1.  **项目分析**: 完整地分析指定的 TypeScript 项目，生成一个包含所有代码结构信息的 JSON 对象。
2.  **字段剔除 (可选)**: 根据 --strip-fields 标志，精确地移除 JSON 中所有指定的字段，以简化输出。
3.  **JMESPath 查询 (可选)**: 如果提供了 --jmespath 表达式，则对 JSON 结果进行过滤和重塑。
4.  **输出**: 将最终处理后的数据输出到指定文件或标准输出。

**标志 (Flags):**

- **-i, --input (必需)**: 指定要分析的 TypeScript 项目的根目录。
- **-o, --output**: 指定输出文件的目录。如果留空，结果将打印到标准输出。
- **-m, --monorepo**: 如果项目是一个 monorepo 仓库，请使用此标志以确保正确解析依赖关系。
- **-x, --exclude**: 指定要从分析中排除的目录或文件的 glob 模式 (例如 'src/**/*.test.ts', 'node_modules')。可多次使用。
- **-s, --strip-fields**: 指定要从结果中递归删除的字段名或字段路径。这对于清理和简化输出非常有用。
    - **按名称剔除**: '-s raw' 会删除所有名为 "raw" 的字段。
    - **按路径剔除**: '-s importDeclarations.raw' 只会删除 importDeclarations 下的 "raw" 字段。
- **-j, --jmespath**: 提供一个 JMESPath 表达式来查询和重塑最终的 JSON 数据。

**数据结构:**

分析结果的顶层是一个 JSON 对象，其关键字段是 "js_data"。"js_data" 是一个以文件绝对路径为键，以该文件的解析结果为值的 map/对象。

**实用查询案例请参考: https://jmespath.org/tutorial.html:**
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// --- 步骤 1: 从标志中获取所有用户提供的参数 ---
			// 通过 Cobra 的 Flags() 方法安全地获取用户通过命令行传入的各个参数值。
			inputPath, _ := cmd.Flags().GetString("input")
			outputPath, _ := cmd.Flags().GetString("output")
			isMonorepo, _ := cmd.Flags().GetBool("monorepo")
			excludePaths, _ := cmd.Flags().GetStringSlice("exclude")
			stripPaths, _ := cmd.Flags().GetStringSlice("strip-fields")
			jmespathExpr, _ := cmd.Flags().GetString("jmespath")

			// --- 步骤 2: 执行项目解析 ---
			// 这是核心分析步骤。它会遍历项目文件，解析 AST，并构建一个包含所有信息的结构体。
			fmt.Fprintln(os.Stderr, "开始解析项目，这可能需要一些时间...")
			config := projectParser.NewProjectParserConfig(inputPath, excludePaths, isMonorepo, []string{})
			parsingResult := projectParser.NewProjectParserResult(config)
			parsingResult.ProjectParser()
			fmt.Fprintln(os.Stderr, "项目解析完成。")

			// --- 步骤 3: 将 Go 结构体转换为通用的 interface{} ---
			// 为了让 JMESPath 和递归字段剔除能够处理数据，需要将强类型的 Go 结构体转换为
			// 由 map[string]interface{} 和 []interface{} 组成的通用数据结构。
			var data interface{}
			fullJSON, err := json.Marshal(parsingResult)
			if err != nil {
				return fmt.Errorf("序列化解析结果失败: %w", err)
			}
			if err := json.Unmarshal(fullJSON, &data); err != nil {
				return fmt.Errorf("反序列化至通用接口失败: %w", err)
			}

			// --- 步骤 4: (可选) 执行递归字段剔除 ---
			// 如果用户指定了 --strip-fields，则在此处清理数据。
			if len(stripPaths) > 0 {
				fmt.Fprintf(os.Stderr, "正在按名称/路径剔除指定的 %d 个字段...\n", len(stripPaths))
				pathsToStrip := make(map[string]struct{}, len(stripPaths))
				for _, path := range stripPaths {
					pathsToStrip[path] = struct{}{}
				}
				stripRecursive(data, "", pathsToStrip)
			}

			// --- 步骤 5: (可选) 执行 JMESPath 查询与重塑 ---
			// 如果用户提供了 --jmespath 表达式，则使用它来过滤和重塑数据。
			var finalData interface{}
			if jmespathExpr != "" {
				fmt.Fprintf(os.Stderr, "正在应用 JMESPath 表达式: %s \n", jmespathExpr)
				result, err := jmespath.Search(jmespathExpr, data)
				if err != nil {
					return fmt.Errorf("执行 JMESPath 表达式失败: %w", err)
				}
				finalData = result
			} else {
				// 如果没有提供表达式，则直接使用（可能已被剔除字段的）原始数据。
				finalData = data
			}

			// --- 步骤 6: 格式化最终结果 ---
			// 将最终数据格式化为易于阅读的 JSON 格式。
			outputJSON, err := json.MarshalIndent(finalData, "", "  ")
			if err != nil {
				return fmt.Errorf("格式化最终输出 JSON 失败: %w", err)
			}

			// --- 步骤 7: 输出结果 ---
			// 根据用户是否指定 --output 路径，将结果写入文件或打印到控制台。
			if outputPath != "" {
				return writeOutputToFile(outputPath, inputPath, outputJSON)
			} else {
				fmt.Println(string(outputJSON))
			}

			return nil
		},
	}

	// --- Flag 定义区 ---
	// 定义了所有此命令接受的命令行标志。
	queryCmd.Flags().StringP("input", "i", "", "要分析的项目根目录")
	queryCmd.Flags().BoolP("monorepo", "m", false, "是否将项目作为 monorepo 进行解析")
	queryCmd.Flags().StringSliceP("exclude", "x", []string{}, "要排除的目录或文件的 glob 模式 (可多次使用)")
	queryCmd.Flags().StringP("output", "o", "", "输出文件目录 (如果为空，则输出到标准输出)")
	queryCmd.Flags().StringSliceP("strip-fields", "s", []string{}, "要递归剔除的字段名或路径 (可多次使用)")
	queryCmd.Flags().StringP("jmespath", "j", "", "(可选) 用于查询和重塑 JSON 数据的 JMESPath 表达式")

	// 将 input 标志标记为必需，如果用户没有提供 -i 或 --input，Cobra 会自动报错。
	if err := queryCmd.MarkFlagRequired("input"); err != nil {
		// 在开发阶段，如果标记失败，直接 panic 以便快速发现问题。
		panic(err)
	}

	return queryCmd
}

// stripRecursive 递归地遍历一个 interface{} 并根据一个“键名”或“父键.子键”的路径映射来删除字段。
// 这个最终版本同时支持按名称剔除和按路径剔除两种模式。
// data: 当前正在处理的数据片段 (map 或 slice)。
// parentKey: 当前数据片段的父键，用于构建完整的访问路径。
// fieldsToStrip: 一个包含所有需要被删除的键名和路径的 set。
func stripRecursive(data interface{}, parentKey string, fieldsToStrip map[string]struct{}) {
	switch value := data.(type) {
	case map[string]interface{}:
		// 如果当前是 map (JSON object)
		for k, v := range value {
			// 检查1：直接按键名匹配 (例如 "raw")
			_, stripByKey := fieldsToStrip[k]

			// 检查2：按完整路径匹配 (例如 "importDeclarations.raw")
			// 只有在 parentKey 非空时才构建路径，以避免在顶层产生如 ".field" 这样的无效路径。
			checkPath := ""
			if parentKey != "" {
				checkPath = parentKey + "." + k
			}
			_, stripByPath := fieldsToStrip[checkPath]

			// 如果键名或完整路径任意一个匹配，则从 map 中删除该键值对。
			if stripByKey || stripByPath {
				delete(value, k)
			} else {
				// 否则，继续向下一层递归。下一层的父键就是当前的键 k。
				// 如果当前键是 "fileInfos"，下一层递归时 parentKey 就是 "fileInfos"。
				newParentKey := k
				if parentKey != "" {
					newParentKey = parentKey + "." + k
				}
				stripRecursive(v, newParentKey, fieldsToStrip)
			}
		}
	case []interface{}:
		// 如果是 slice (JSON array)
		// 遍历切片中的所有元素并递归处理。
		// 将当前切片的父键(parentKey)直接传递下去，这样可以检查到 "arrayKey.field" 这样的路径。
		for _, v := range value {
			stripRecursive(v, parentKey, fieldsToStrip)
		}
	}
}

// writeOutputToFile 将结果写入指定目录下的一个自动生成的文件中。
func writeOutputToFile(outputDir, inputPath string, content []byte) error {
	// 确保输出目录存在，如果不存在则创建它。
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}
	// 从输入路径中提取基本名称（例如，从 "/path/to/my-project" 得到 "my-project"）。
	baseName := filepath.Base(inputPath)
	// 将基本名称中的空格替换为下划线，以创建更安全的文件名。
	safeBaseName := strings.ReplaceAll(baseName, " ", "_")
	// 构建一个唯一的输出文件名。
	outputFileName := fmt.Sprintf("%s_query_result.json", safeBaseName)
	outputFile := filepath.Join(outputDir, outputFileName)

	// 将内容写入文件。
	if err := os.WriteFile(outputFile, content, 0644); err != nil {
		return fmt.Errorf("写入 JSON 文件失败: %w", err)
	}
	// 在标准错误流中打印成功消息，告知用户文件已保存的位置。
	fmt.Fprintf(os.Stderr, "✅ 结果已成功写入到 %s\n", outputFile)
	return nil
}
