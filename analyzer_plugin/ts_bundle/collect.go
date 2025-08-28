// 该文件实现了基于入口文件和类型名的 TypeScript 类型依赖递归收集与输出。
// 主要流程为：
// 1. 以入口文件和类型为起点，递归解析类型、接口及 import 依赖，收集所有相关类型声明源码。
// 2. 支持 alias、npm 包、命名空间导入等常见 TypeScript 导入场景。
// 3. 最终将所有依赖类型源码合并输出到指定文件。

package ts_bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"
)

// CollectResult 类型依赖收集器
// 负责递归收集 TypeScript 文件中的类型依赖关系
type CollectResult struct {
	RootPath      string                          // 项目根目录
	Alias         map[string]string               // tsconfig.json 中的路径别名
	BaseUrl       string                          // tsconfig.json 中的 baseUrl
	Extensions    []string                        // 支持的文件扩展名
	SourceCodeMap map[string]string               // 已收集的类型源码，键为文件路径和类型名的组合
	visited       map[string]bool                 // 用于检测循环依赖的访问记录
	fileCache     map[string]*parser.ParserResult // 文件解析缓存，避免重复解析同一文件
}

// findProjectRoot 向上查找项目根目录
// 通过查找项目标识文件来确定项目根目录，查找顺序：
// 1. 如果提供了 projectRootPath，直接使用
// 2. 查找 tsconfig.json（TypeScript 项目标识）
// 3. 查找 package.json（Node.js 项目标识）
// 4. 查找 .git 目录（Git 仓库标识）
// 5. 如果都找不到，使用入口文件所在目录
func findProjectRoot(inputAnalyzeFile string, projectRootPath string) string {
	// 如果显式提供了项目根路径，直接使用
	if projectRootPath != "" {
		return projectRootPath
	}

	// 获取入口文件的绝对路径
	absFilePath, err := filepath.Abs(inputAnalyzeFile)
	if err != nil {
		// 如果无法获取绝对路径，使用文件所在目录
		dir := filepath.Dir(inputAnalyzeFile)
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return dir // 最坏情况下返回原始目录
		}
		return absDir
	}

	// 从入口文件所在目录开始向上查找
	currentDir := filepath.Dir(absFilePath)

	// 限制向上查找的深度，避免无限循环
	for i := 0; i < 10; i++ {
		// 检查是否存在 tsconfig.json
		if _, err := os.Stat(filepath.Join(currentDir, "tsconfig.json")); err == nil {
			return currentDir
		}

		// 检查是否存在 package.json
		if _, err := os.Stat(filepath.Join(currentDir, "package.json")); err == nil {
			return currentDir
		}

		// 检查是否存在 .git 目录
		if _, err := os.Stat(filepath.Join(currentDir, ".git")); err == nil {
			return currentDir
		}

		// 向上移动一级目录
		parentDir := filepath.Dir(currentDir)

		// 如果已经到达文件系统根目录，停止查找
		if parentDir == currentDir {
			break
		}

		currentDir = parentDir
	}

	// 如果找不到项目根目录，使用入口文件所在目录
	return filepath.Dir(absFilePath)
}

// buildSourceCodeMapKey 构建 SourceCodeMap 的键
// 使用不容易冲突的分隔符来组合文件路径和类型名
func buildSourceCodeMapKey(filePath, typeName string) string {
	return fmt.Sprintf("%s\x00%s", filePath, typeName) // 使用空字符作为分隔符
}

// parseFile 解析文件并返回解析结果
// 使用缓存机制避免重复解析同一文件，提高性能
func (br *CollectResult) parseFile(absFilePath string) *parser.ParserResult {
	// 检查缓存中是否已存在解析结果
	if cached, exists := br.fileCache[absFilePath]; exists {
		return cached
	}

	// 创建新的解析器并解析文件
	pr := parser.NewParserResult(absFilePath)
	pr.Traverse()
	result := pr.GetResult()

	// 将解析结果存入缓存
	br.fileCache[absFilePath] = &result
	return &result
}

// handleNamespaceImport 处理命名空间导入
// 专门处理 import * as allTypes from './type' 这样的命名空间导入场景
func (br *CollectResult) handleNamespaceImport(absFilePath, typeName, parentTypeName string, importDecl parser.ImportDeclarationResult, module parser.ImportModule) {
	// 解析 typeName: allTypes.MerchantData，提取出 MerchantData 部分
	refNameArr := strings.Split(typeName, ".")
	realRefName := refNameArr[len(refNameArr)-1] // MerchantData

	// 检查是否匹配当前命名空间标识符
	if refNameArr[0] == module.Identifier { // allTypes
		// 生成替换类型名，如 allTypes_MerchantData
		replaceTypeName := module.Identifier + "_" + realRefName

		// 替换源码中的类型引用
		// 将 PreloadedState 中的 allTypes.MerchantData 替换为 allTypes_MerchantData
		key := buildSourceCodeMapKey(absFilePath, parentTypeName)
		realTargetTypeRaw := strings.ReplaceAll(br.SourceCodeMap[key], typeName, replaceTypeName)
		br.SourceCodeMap[key] = realTargetTypeRaw

		// 解析导入源文件路径
		sourceData := projectParser.MatchImportSource(absFilePath, importDecl.Source, br.RootPath, br.Alias, br.Extensions, br.BaseUrl)
		nextFile := ""
		if sourceData.Type == "file" {
			nextFile = sourceData.FilePath
		} else {
			// 处理 npm 包导入
			nextFile = utils.ResolveNpmPath(absFilePath, br.RootPath, importDecl.Source, true)
			// 检查结尾是否有文件后缀，如果没有后缀，需要基于 Extensions 尝试去匹配
			if !utils.HasExtension(nextFile, br.Extensions) {
				nextFile = utils.FindRealFilePath(nextFile, br.Extensions)
			}
		}

		// 递归收集依赖类型
		br.collectFileType(nextFile, realRefName, replaceTypeName, typeName)
	}
}

// NewCollectResult 构造函数，初始化 CollectResult。
// 通过入口文件路径自动推断项目根目录、npm 列表、alias、扩展名等信息。
func NewCollectResult(inputAnalyzeFile string, inputAnalyzeType string, projectRootPath string) CollectResult {
	// 通过更智能的方式查找项目根目录
	rootPath := findProjectRoot(inputAnalyzeFile, projectRootPath)

	// 获取 tsconfig.json 中的 alias 列表
	config := projectParser.NewProjectParserConfig(rootPath, []string{}, false)
	ar := projectParser.NewProjectParserResult(config)

	// 返回初始化的 CollectResult 实例
	return CollectResult{
		RootPath:      rootPath,
		Alias:         ar.Config.RootTsConfig.Alias,
		BaseUrl:       ar.Config.RootTsConfig.BaseUrl,
		Extensions:    ar.Config.Extensions,
		SourceCodeMap: make(map[string]string),
		visited:       make(map[string]bool),
		fileCache:     make(map[string]*parser.ParserResult), // 初始化文件解析缓存
	}
}

// collectFileType 递归解析指定文件中的类型依赖。
// absFilePath: 当前解析的文件绝对路径
// typeName: 当前要查找的类型名
// replaceTypeName: 类型重命名（如 import {A as B}）时的替换名
// parentTypeName: 父类型名（用于命名空间类型替换）
func (br *CollectResult) collectFileType(absFilePath string, typeName string, replaceTypeName string, parentTypeName string) error {
	// 创建一个唯一的键来标识这次调用，用于循环依赖检测
	visitKey := fmt.Sprintf("%s::%s::%s::%s", absFilePath, typeName, replaceTypeName, parentTypeName)

	// 检查是否已经访问过这个键，如果是则直接返回以避免循环依赖
	if br.visited[visitKey] {
		fmt.Printf("检测到循环依赖，跳过: %s \n", visitKey)
		return nil
	}

	// 标记为已访问
	br.visited[visitKey] = true
	defer func() {
		// 在函数退出时取消标记（允许在不同路径上重新访问）
		delete(br.visited, visitKey)
	}()

	fmt.Printf("开始解析当前文件: %s \n", absFilePath)

	// 使用缓存解析文件，避免重复解析
	parserResult := br.parseFile(absFilePath)

	// 查找类型声明 type
	if typeDecl, found := parserResult.TypeDeclarations[typeName]; found {
		realRaw := typeDecl.Raw
		if replaceTypeName != "" {
			// 如果需要重命名类型，则替换源码中的类型名
			realRaw = strings.ReplaceAll(typeDecl.Raw, typeName, replaceTypeName)
		}

		// 将类型声明存入源码映射
		br.SourceCodeMap[buildSourceCodeMapKey(absFilePath, typeName)] = realRaw
		// 递归收集该类型引用的其他类型
		for ref := range typeDecl.Reference {
			br.collectFileType(absFilePath, ref, "", typeName)
		}
		return nil
	}

	// 查找接口声明 interface
	if interfaceDecl, found := parserResult.InterfaceDeclarations[typeName]; found {
		realRaw := interfaceDecl.Raw
		if replaceTypeName != "" {
			// 如果需要重命名类型，则替换源码中的类型名
			realRaw = strings.ReplaceAll(interfaceDecl.Raw, typeName, replaceTypeName)
		}
		// 将接口声明存入源码映射
		br.SourceCodeMap[buildSourceCodeMapKey(absFilePath, typeName)] = realRaw
		// 递归收集该接口引用的其他类型
		for ref := range interfaceDecl.Reference {
			br.collectFileType(absFilePath, ref, "", typeName)
		}
		return nil
	}

	// 查找枚举声明 enum
	if enumDecl, found := parserResult.EnumDeclarations[typeName]; found {
		realRaw := enumDecl.Raw
		if replaceTypeName != "" {
			// 如果需要重命名类型，则替换源码中的类型名
			realRaw = strings.ReplaceAll(enumDecl.Raw, typeName, replaceTypeName)
		}
		// 将枚举声明存入源码映射
		br.SourceCodeMap[buildSourceCodeMapKey(absFilePath, typeName)] = realRaw
		return nil
	}

	// 查找 import 依赖
	for _, importDecl := range parserResult.ImportDeclarations {
		for _, module := range importDecl.ImportModules {
			// 普通命名导入处理
			if module.Identifier == typeName {
				realTypeName := typeName
				var replaceTypeName string
				// 处理 import { A as B } 这样的重命名导入
				if module.Type == "named" && module.ImportModule != typeName {
					realTypeName = module.ImportModule
					replaceTypeName = typeName
				}

				// 解析导入源文件路径
				sourceData := projectParser.MatchImportSource(absFilePath, importDecl.Source, br.RootPath, br.Alias, br.Extensions, br.BaseUrl)

				nextFile := ""
				if sourceData.Type == "file" {
					// 本地文件导入
					nextFile = sourceData.FilePath
				} else if sourceData.Type == "npm" {
					// npm 包导入
					nextFile = utils.ResolveNpmPath(absFilePath, br.RootPath, importDecl.Source, true)
					// 检查结尾是否有文件后缀，如果没有后缀，需要基于 Extensions 尝试去匹配
					if !utils.HasExtension(nextFile, br.Extensions) {
						nextFile = utils.FindRealFilePath(nextFile, br.Extensions)
					}
				}
				// 递归收集依赖类型
				br.collectFileType(nextFile, realTypeName, replaceTypeName, typeName)
			}

			// 命名空间导入处理: import * as allTypes from './type';
			if module.Type == "namespace" {
				// 委托给专门的处理方法
				br.handleNamespaceImport(absFilePath, typeName, parentTypeName, importDecl, module)
			}
		}
	}
	return nil
}
