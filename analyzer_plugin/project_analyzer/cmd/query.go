// package cmd 存放了所有命令行工具的实现。
package cmd

// go run main.go query 'js_data' --omit-fields="*.raw,*.sourceLocation,*.expression" -i /Users/bird/company/sc1.0/live/shopline-live-sale -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin -x "node_modules/**" -x "bffApiDoc/**"

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	jmespath "github.com/jmespath/go-jmespath"
	"github.com/spf13/cobra"
)

// GetQueryCmd 返回新增的 query 命令的 Cobra 命令对象。
// 这个命令允许用户使用 JMESPath 对项目解析结果进行通用查询。
func GetQueryCmd() *cobra.Command {
	var (
		inputPath   string
		outputPath  string
		excludePath []string
		isMonorepo  bool
		omitFields  []string
		noCache     bool
	)

	queryCmd := &cobra.Command{
		Use:   "query <jmespath_query>",
		Short: "使用 JMESPath 对项目解析结果进行通用查询。",
		Long: "该命令提供了一个强大的、通用的方式来查询和过滤项目解析后的数据。\n\n" +
			"核心功能:\n" +
			"- 使用 JMESPath 语言进行灵活的数据查询和筛选。\n" +
			"- 通过 --omit-fields 标志对查询结果进行精细的“瘦身”，移除不需要的字段。\n\n" +
			"JMESPath 查询示例:\n" +
			"  # 查询所有从 'react' 库的导入\n" +
			"  query 'js_data.*.importDeclarations[?source.npmPkg==`'react'`]'\n\n" +
			"  # 同时查询函数和导入声明，并压平到一个列表\n" +
			"  query 'js_data.*.[functionsDeclarations, importDeclarations][][]'\n\n" +
			"--omit-fields 使用示例:\n" +
			"  # 1. 全局模式: 移除所有节点下的 'raw' 字段\n" +
			"  --omit-fields=\"*.raw\"\n\n" +
			"  # 2. 精确模式: 移除特定节点下的特定字段\n" +
			"  --omit-fields=\"FunctionDeclaration.SourceLocation.Start\"\n\n" +
			"高级查询示例:\n" +
			"  # 1. 提取所有 package.json 的数据\n" +
			"  query 'package_data'\n\n" +
			"  # 2. 提取所有NPM依赖包的名称\n" +
			"  query 'package_data.*.npmList.*.name'\n\n" +
			"  # 3. 输出未经任何处理的完整原始JSON数据\n" +
			"  query '@'",
		Args: cobra.ExactArgs(1), // 要求必须且只有一个参数，即 JMESPath 查询语句
		Run: func(cmd *cobra.Command, args []string) {
			jmespathQuery := args[0]

			// 1. 检查输入路径
			if inputPath == "" {
				fmt.Println("错误: 请使用 -i 或 --input 标志提供项目路径。")
				return
			}

			// 2. 解析项目
			parsingResult, err := ParseProjectWithCache(inputPath, excludePath, isMonorepo, !noCache)
			if err != nil {
				fmt.Printf("错误: 解析项目失败: %v\n", err)
				return
			}

			// 3. 为节点注入类型信息，为后续处理做准备
			typedData, err := injectNodeTypes(parsingResult)
			if err != nil {
				fmt.Printf("错误: 注入节点类型信息失败: %v\n", err)
				return
			}

			// 4. 执行 JMESPath 查询
			jmespathResult, err := jmespath.Search(jmespathQuery, typedData)
			if err != nil {
				fmt.Printf("错误: 执行 JMESPath 查询失败: %v\n", err)
				return
			}

			// 5. 如果定义了 --omit-fields，则对结果进行处理
			if len(omitFields) > 0 {
				rules := parseOmitRules(omitFields)
				recursiveOmit(jmespathResult, rules)
			}

			// 6. 输出最终结果
			outputData, err := json.MarshalIndent(jmespathResult, "", "  ")
			if err != nil {
				fmt.Printf("错误: 序列化最终结果失败: %v\n", err)
				return
			}

			if outputPath == "" {
				// 如果未指定输出路径，则直接打印到控制台
				fmt.Println(string(outputData))
			} else {
				// 如果指定了输出路径，则写入文件
				outputFileName := GenerateOutputFileName(inputPath, "query_result")
				err := WriteJSONResult(outputPath, outputFileName, jmespathResult)
				if err != nil {
					fmt.Printf("错误: 无法将结果写入文件: %v\n", err)
				}
			}
		},
	}

	queryCmd.Flags().StringVarP(&inputPath, "input", "i", "", "项目根目录 (必需)")
	queryCmd.Flags().StringVarP(&outputPath, "output", "o", "", "输出文件目录 (默认为打印到控制台)")
	queryCmd.Flags().StringSliceVarP(&excludePath, "exclude", "x", []string{}, "排除的 glob 模式")
	queryCmd.Flags().BoolVarP(&isMonorepo, "monorepo", "m", false, "是否为 monorepo")
	queryCmd.Flags().StringSliceVarP(&omitFields, "omit-fields", "", []string{}, "需要从结果中剔除的字段 (例如: '*.raw', 'FunctionDeclaration.Parameters)")
	queryCmd.Flags().BoolVar(&noCache, "no-cache", false, "禁用缓存，强制重新解析整个项目")
	queryCmd.MarkFlagRequired("input")

	return queryCmd
}

func ParseProjectWithCache(rootPath string, ignore []string, isMonorepo bool, useCache bool) (*projectParser.ProjectParserResult, error) {
	cacheDir := filepath.Join(rootPath, ".analyzer_cache")
	cacheFile := filepath.Join(cacheDir, "parser_result.gob")

	if useCache {
		stale, err := isQueryCacheStale(cacheFile, rootPath, ignore)
		if err == nil && !stale {
			fmt.Println("缓存有效，正在从缓存加载解析结果...")
			file, err := os.Open(cacheFile)
			if err == nil {
				defer file.Close()
				decoder := gob.NewDecoder(file)
				var result projectParser.ProjectParserResult
				if err := decoder.Decode(&result); err == nil {
					fmt.Println("成功从缓存加载。")
					return &result, nil
				}
			}
			fmt.Println("无法读取或解码缓存，将执行完整解析。")
		}
	}

	result, err := ParseProject(rootPath, ignore, isMonorepo)
	if err != nil {
		return nil, err
	}

	if useCache {
		fmt.Println("正在保存解析结果到缓存...")
		if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
			return nil, fmt.Errorf("创建缓存目录失败: %w", err)
		}
		file, err := os.Create(cacheFile)
		if err != nil {
			return nil, fmt.Errorf("创建缓存文件失败: %w", err)
		}
		defer file.Close()

		encoder := gob.NewEncoder(file)
		if err := encoder.Encode(result); err != nil {
			return nil, fmt.Errorf("写入缓存失败: %w", err)
		}
		fmt.Println("缓存写入成功: ", cacheFile)
	}

	return result, nil
}

func isQueryCacheStale(cacheFile string, rootPath string, ignorePatterns []string) (bool, error) {
	cacheInfo, err := os.Stat(cacheFile)
	if err != nil {
		return true, err
	}
	cacheModTime := cacheInfo.ModTime()

	var isStale bool
	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if isStale {
			return filepath.SkipDir
		}

		for _, pattern := range ignorePatterns {
			if matched, _ := filepath.Match(pattern, path); matched {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		ext := filepath.Ext(path)
		if ext == ".ts" || ext == ".tsx" || ext == ".js" || ext == ".jsx" || ext == "package.json" {
			if info.ModTime().After(cacheModTime) {
				fmt.Printf("缓存失效: 文件被修改过 %s\n", path)
				isStale = true
			}
		}
		return nil
	})

	return isStale, err
}

func injectNodeTypes(result *projectParser.ProjectParserResult) (interface{}, error) {
	// 1. 先通过 JSON 序列化和反序列化，将整个结构体转换为 map[string]interface{}
	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	var data interface{}
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return nil, err
	}

	// 2. 递归地为特定节点添加 __type 字段
	if rootMap, ok := data.(map[string]interface{}); ok {
		if jsData, ok := rootMap["js_data"].(map[string]interface{}); ok {
			for _, fileData := range jsData {
				if fileMap, ok := fileData.(map[string]interface{}); ok {
					// 遍历文件中的所有可能的声明类型
					for key, value := range fileMap {
						// key 是 "functionsDeclarations", "importDeclarations" 等
						// value 是这些声明的数组
						if declarations, ok := value.([]interface{}); ok {
							var typeName string
							// 特殊处理 JsxElements
							if key == "jsxElements" {
								typeName = "JSXElement"
							} else {
								typeName = strings.TrimSuffix(key, "s")
								typeName = strings.TrimSuffix(typeName, "Declaration")
								typeName = strings.Title(typeName) + "Declaration"
							}

							// 为数组中的每个对象注入 __type
							for _, decl := range declarations {
								if declMap, ok := decl.(map[string]interface{}); ok {
									declMap["__type"] = typeName
								}
							}
						}
					}
				}
			}
		}
	}
	return data, nil
}

// omitRule 表示一条剔除规则
type omitRule struct {
	// IsGlobal 标记这是否是一条全局规则 (例如 *.raw)
	IsGlobal bool
	// NodeType 是规则应用的目标节点类型 (例如 FunctionDeclaration)
	NodeType string
	// Path 是要剔除的字段路径 (例如 Parameters.raw)
	Path string
}

// parseOmitRules 解析来自 --omit-fields 标志的字符串，并将其转换为结构化的规则。
func parseOmitRules(fields []string) []omitRule {
	var rules []omitRule
	// --omit-fields 支持逗号分割的多个规则
	fieldParts := strings.Split(strings.Join(fields, ","), ",")
	for _, field := range fieldParts {
		trimmedField := strings.TrimSpace(field)
		if trimmedField == "" {
			continue
		}
		parts := strings.SplitN(trimmedField, ".", 2)
		if len(parts) == 2 {
			nodeType := parts[0]
			path := parts[1]
			if nodeType == "*" {
				rules = append(rules, omitRule{IsGlobal: true, Path: path})
			} else {
				rules = append(rules, omitRule{NodeType: nodeType, Path: path})
			}
		}
	}
	return rules
}

// recursiveOmit 递归地遍历查询结果，并根据规则剔除字段。
func recursiveOmit(current interface{}, rules []omitRule) {
	switch node := current.(type) {
	case map[string]interface{}:
		// 规则应用: 对当前对象应用所有匹配的规则
		typeName, _ := node["__type"].(string)
		for _, rule := range rules {
			if rule.IsGlobal {
				// 应用全局规则
				deleteFieldByPath(node, rule.Path)
			} else if typeName != "" && rule.NodeType == typeName {
				// 应用特定类型规则
				deleteFieldByPath(node, rule.Path)
			}
		}

		// 递归深入: 继续遍历对象的子字段
		for _, value := range node {
			recursiveOmit(value, rules)
		}

	case []interface{}:
		// 如果是数组，则递归遍历其所有元素
		for _, item := range node {
			recursiveOmit(item, rules)
		}
	}
}

// deleteFieldByPath 根据点路径 (dot path) 从一个对象中删除字段。
// 它能够处理嵌套对象和数组。
func deleteFieldByPath(data interface{}, path string) {
	parts := strings.Split(path, ".")
	current := data

	// 遍历路径的每一部分，除了最后一部分（因为最后一部分是要删除的键）
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		// 将 current 断言为 map 类型，以便访问其字段
		mapCurrent, ok := current.(map[string]interface{})
		if !ok {
			return // 如果路径中的某个环节不是 map，则无法继续，路径无效
		}

		next, exists := mapCurrent[part]
		if !exists {
			return // 路径不存在
		}

		// 如果路径的下一部分是数组，则递归地对数组中每个元素应用剩余的路径
		if reflect.TypeOf(next).Kind() == reflect.Slice {
			if nextSlice, ok := next.([]interface{}); ok {
				remainingPath := strings.Join(parts[i+1:], ".")
				for _, item := range nextSlice {
					deleteFieldByPath(item, remainingPath)
				}
				return // 递归调用已经处理了剩余路径，直接返回
			}
		}
		current = next
	}

	// 删除目标字段
	if mapCurrent, ok := current.(map[string]interface{}); ok {
		delete(mapCurrent, parts[len(parts)-1])
	}
}
