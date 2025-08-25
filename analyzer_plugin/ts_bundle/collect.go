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
	"regexp"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"
)

// safeReplace 使用正则表达式确保只替换完整的单词，避免错误修改如 `PageData` 中的 `Page`。
func safeReplace(source, oldName, newName string) string {
	if oldName == "" || newName == "" || oldName == newName {
		return source
	}
	// \b 是单词边界，确保我们匹配的是一个独立的单词
	re := regexp.MustCompile(`\b` + regexp.QuoteMeta(oldName) + `\b`)
	return re.ReplaceAllString(source, newName)
}

type CollectResult struct {
	RootPath      string            // 项目根目录
	Alias         map[string]string // tsconfig.json 中的路径别名
	BaseUrl       string            // tsconfig.json 中的 baseUrl
	Extensions    []string          // 支持的文件扩展名
	SourceCodeMap map[string]string // 已收集的类型源码
	visited       map[string]bool   // 记录已访问的 "文件:类型" 对，防止循环依赖
}

// findProjectRoot searches upwards from a starting directory for a tsconfig.json file.
func findProjectRoot(startDir string) (string, error) {
	dir := startDir
	for {
		tsconfigPath := filepath.Join(dir, "tsconfig.json")
		if _, err := os.Stat(tsconfigPath); err == nil {
			return dir, nil // Found it
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			// Reached the root of the filesystem
			return "", fmt.Errorf("tsconfig.json not found in any parent directory")
		}
		dir = parentDir
	}
}

// NewCollectResult 构造函数，初始化 CollectResult。
func NewCollectResult(inputAnalyzeFile string, inputAnalyzeType string, projectRootPath string) CollectResult {
	var rootPath string = projectRootPath
	if rootPath == "" {
		absFilePath, err := filepath.Abs(inputAnalyzeFile)
		if err != nil {
			absFilePath = inputAnalyzeFile
		}
		foundRoot, err := findProjectRoot(filepath.Dir(absFilePath))
		if err == nil {
			rootPath = foundRoot
		} else {
			if strings.Contains(absFilePath, "/src") {
				rootPath = strings.Split(absFilePath, "/src")[0]
			} else {
				rootPath = filepath.Dir(absFilePath)
			}
		}
	}

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
func (br *CollectResult) collectFileType(absFilePath string, typeName string, replaceTypeName string, parentTypeName string) {
	visitedKey := absFilePath + ":" + typeName
	if br.visited[visitedKey] {
		return
	}
	br.visited[visitedKey] = true

	fmt.Printf("Parsing file: %s for type: %s\n", absFilePath, typeName)

	pr := parser.NewParserResult(absFilePath)
	pr.Traverse()
	parserResult := pr.GetResult()

	// 查找类型声明 type
	if typeDecl, found := parserResult.TypeDeclarations[typeName]; found {
		rawSource := typeDecl.Raw
		for ref := range typeDecl.Reference {
			if strings.Contains(ref, ".") {
				flatName := strings.ReplaceAll(ref, ".", "_")
				rawSource = safeReplace(rawSource, ref, flatName)
			}
		}
		rawSource = safeReplace(rawSource, typeName, replaceTypeName)
		br.SourceCodeMap[absFilePath+"_"+typeName] = rawSource
		for ref := range typeDecl.Reference {
			br.collectFileType(absFilePath, ref, "", typeName)
		}
		return
	}

	// 查找接口声明 interface
	if interfaceDecl, found := parserResult.InterfaceDeclarations[typeName]; found {
		rawSource := interfaceDecl.Raw
		for ref := range interfaceDecl.Reference {
			if strings.Contains(ref, ".") {
				flatName := strings.ReplaceAll(ref, ".", "_")
				rawSource = safeReplace(rawSource, ref, flatName)
			}
		}
		rawSource = safeReplace(rawSource, typeName, replaceTypeName)
		br.SourceCodeMap[absFilePath+"_"+typeName] = rawSource
		for ref := range interfaceDecl.Reference {
			br.collectFileType(absFilePath, ref, "", typeName)
		}
		return
	}

	// 查找枚举声明 enum
	if enumDecl, found := parserResult.EnumDeclarations[typeName]; found {
		rawSource := enumDecl.Raw
		rawSource = safeReplace(rawSource, typeName, replaceTypeName)
		br.SourceCodeMap[absFilePath+"_"+typeName] = rawSource
		return
	}

	// 如果在本地声明中没找到，则在 import 语句中查找
	for _, importDecl := range parserResult.ImportDeclarations {
		for _, module := range importDecl.ImportModules {
			// 处理 `import {A as B}` 或 `import {A}`
			if module.Identifier == typeName {
				realTypeName := module.ImportModule
				alias := typeName // B是A的别名

				sourceData := projectParser.MatchImportSource(absFilePath, importDecl.Source, br.RootPath, br.Alias, br.Extensions, br.BaseUrl)
				nextFile := ""
				if sourceData.Type == "file" {
					nextFile = sourceData.FilePath
				} else if sourceData.Type == "npm" {
					nextFile = utils.ResolveNpmPath(absFilePath, br.RootPath, importDecl.Source, true)
					if !utils.HasExtension(nextFile, br.Extensions) {
						nextFile = utils.FindRealFilePath(nextFile, br.Extensions)
					}
				}
				// 递归到下一个文件，查找真实类型A，并告知它需要被替换为别名B
				br.collectFileType(nextFile, realTypeName, alias, typeName)
			}

			// 处理 `import * as ns from '...'`
			if module.Type == "namespace" && strings.HasPrefix(typeName, module.Identifier+".") {
				realTypeName := strings.TrimPrefix(typeName, module.Identifier+".")
				alias := strings.ReplaceAll(typeName, ".", "_") // ns.Type -> ns_Type

				sourceData := projectParser.MatchImportSource(absFilePath, importDecl.Source, br.RootPath, br.Alias, br.Extensions, br.BaseUrl)
				nextFile := ""
				if sourceData.Type == "file" {
					nextFile = sourceData.FilePath
				} else {
					nextFile = utils.ResolveNpmPath(absFilePath, br.RootPath, importDecl.Source, true)
					if !utils.HasExtension(nextFile, br.Extensions) {
							nextFile = utils.FindRealFilePath(nextFile, br.Extensions)
					}
				}
				// 递归到下一个文件，查找真实类型Type，并告知它需要被替换为扁平化的名称ns_Type
				br.collectFileType(nextFile, realTypeName, alias, typeName)
			}
		}
	}
}
