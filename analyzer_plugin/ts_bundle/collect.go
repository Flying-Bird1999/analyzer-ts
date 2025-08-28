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

type CollectResult struct {
	RootPath      string            // 项目根目录
	Alias         map[string]string // tsconfig.json 中的路径别名
	BaseUrl       string            // tsconfig.json 中的 baseUrl
	Extensions    []string          // 支持的文件扩展名
	SourceCodeMap map[string]string // 已收集的类型源码
	visited       map[string]bool   // 用于检测循环依赖的访问记录
}

// findProjectRoot 向上查找项目根目录
// 查找顺序：
// 1. 如果提供了 projectRootPath，直接使用
// 2. 查找 tsconfig.json
// 3. 查找 package.json
// 4. 查找 .git 目录
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

// NewCollectResult 构造函数，初始化 CollectResult。
// 通过入口文件路径自动推断项目根目录、npm 列表、alias、扩展名等信息。
func NewCollectResult(inputAnalyzeFile string, inputAnalyzeType string, projectRootPath string) CollectResult {
	// 通过更智能的方式查找项目根目录
	rootPath := findProjectRoot(inputAnalyzeFile, projectRootPath)

	// 获取 tsconfig.json 中的 alias 列表
	config := projectParser.NewProjectParserConfig(rootPath, []string{}, false)
	ar := projectParser.NewProjectParserResult(config)

	return CollectResult{
		RootPath:      rootPath,
		Alias:         ar.Config.RootTsConfig.Alias,
		BaseUrl:       ar.Config.RootTsConfig.BaseUrl,
		Extensions:    ar.Config.Extensions,
		SourceCodeMap: make(map[string]string),
		visited:       make(map[string]bool),
	}
}

// collectFileType 递归解析指定文件中的类型依赖。
// absFilePath: 当前解析的文件绝对路径path:/Users/bird/Desktop/sp/smart-push-new/tsconfig.base.json
// typeName: 当前要查找的类型名
// replaceTypeName: 类型重命名（如 import {A as B}）时的替换名
// parentTypeName: 父类型名（用于命名空间类型替换）
func (br *CollectResult) collectFileType(absFilePath string, typeName string, replaceTypeName string, parentTypeName string) {
	// 创建一个唯一的键来标识这次调用
	visitKey := fmt.Sprintf("%s::%s::%s::%s", absFilePath, typeName, replaceTypeName, parentTypeName)
	
	// 检查是否已经访问过这个键，如果是则直接返回以避免循环依赖
	if br.visited[visitKey] {
		fmt.Printf("检测到循环依赖，跳过: %s \n", visitKey)
		return
	}
	
	// 标记为已访问
	br.visited[visitKey] = true
	defer func() {
		// 在函数退出时取消标记（允许在不同路径上重新访问）
		delete(br.visited, visitKey)
	}()
	
	fmt.Printf("开始解析当前文件: %s \n", absFilePath)

	// 解析当前文件
	pr := parser.NewParserResult(absFilePath)
	pr.Traverse()
	parserResult := pr.GetResult()

	// 查找类型声明 type
	if typeDecl, found := parserResult.TypeDeclarations[typeName]; found {
		realRaw := typeDecl.Raw
		if replaceTypeName != "" {
			realRaw = strings.ReplaceAll(typeDecl.Raw, typeName, replaceTypeName)
		}

		br.SourceCodeMap[absFilePath+"_"+typeName] = realRaw
		for ref := range typeDecl.Reference {
			br.collectFileType(absFilePath, ref, "", typeName)
		}
		return
	}

	// 查找接口声明 interface
	if interfaceDecl, found := parserResult.InterfaceDeclarations[typeName]; found {
		realRaw := interfaceDecl.Raw
		if replaceTypeName != "" {
			realRaw = strings.ReplaceAll(interfaceDecl.Raw, typeName, replaceTypeName)
		}
		br.SourceCodeMap[absFilePath+"_"+typeName] = realRaw
		for ref := range interfaceDecl.Reference {
			br.collectFileType(absFilePath, ref, "", typeName)
		}
		return
	}

	// 查找枚举声明 enum
	if enumDecl, found := parserResult.EnumDeclarations[typeName]; found {
		realRaw := enumDecl.Raw
		if replaceTypeName != "" {
			realRaw = strings.ReplaceAll(enumDecl.Raw, typeName, replaceTypeName)
		}
		br.SourceCodeMap[absFilePath+"_"+typeName] = realRaw
		return
	}

	// 查找 import 依赖
	for _, importDecl := range parserResult.ImportDeclarations {
		for _, module := range importDecl.ImportModules {
			// 普通命名导入
			if module.Identifier == typeName {
				realTypeName := typeName
				var replaceTypeName string
				if module.Type == "named" && module.ImportModule != typeName {
					realTypeName = module.ImportModule
					replaceTypeName = typeName
				}
				sourceData := projectParser.MatchImportSource(absFilePath, importDecl.Source, br.RootPath, br.Alias, br.Extensions, br.BaseUrl)

				nextFile := ""
				if sourceData.Type == "file" {
					nextFile = sourceData.FilePath
				} else if sourceData.Type == "npm" {
					nextFile = utils.ResolveNpmPath(absFilePath, br.RootPath, importDecl.Source, true)
					// 检查结尾是否有文件后缀，如果没有后缀，需要基于Extensions尝试去匹配
					if !utils.HasExtension(nextFile, br.Extensions) {
						nextFile = utils.FindRealFilePath(nextFile, br.Extensions)
					}
				}
				br.collectFileType(nextFile, realTypeName, replaceTypeName, typeName)
			}

			// case: import * as allTypes from './type';
			if module.Type == "namespace" {
				// 解析typeName: allTypes.MerchantData。提取出 allTypes.MerchantData 中的 MerchantData
				refNameArr := strings.Split(typeName, ".")
				realRefName := refNameArr[len(refNameArr)-1] // MerchantData
				if refNameArr[0] == module.Identifier {      // allTypes
					var replaceTypeName = module.Identifier + "_" + realRefName // allTypes_MerchantData
					// 替换源码的类型，PreloadedState中的 allTypes.MerchantData -> allTypes_MerchantData
					realTargetTypeRaw := strings.ReplaceAll(br.SourceCodeMap[absFilePath+"_"+parentTypeName], typeName, replaceTypeName)
					br.SourceCodeMap[absFilePath+"_"+parentTypeName] = realTargetTypeRaw

					sourceData := projectParser.MatchImportSource(absFilePath, importDecl.Source, br.RootPath, br.Alias, br.Extensions, br.BaseUrl)
					nextFile := ""
					if sourceData.Type == "file" {
						nextFile = sourceData.FilePath
					} else {
						nextFile = utils.ResolveNpmPath(absFilePath, br.RootPath, importDecl.Source, true)
						// 检查结尾是否有文件后缀，如果没有后缀，需要基于Extensions尝试去匹配
						if !utils.HasExtension(nextFile, br.Extensions) {
							nextFile = utils.FindRealFilePath(nextFile, br.Extensions)
						}
					}
					br.collectFileType(nextFile, realRefName, replaceTypeName, typeName)
				}
			}
		}
	}
}
