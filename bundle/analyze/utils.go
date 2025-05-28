package analyze

import (
	"encoding/json"
	"fmt"
	"main/bundle/utils"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// 新方法：从 tsconfig.json 中读取 alias
func ReadAliasFromTsConfig(rootPath string) map[string]string {
	var alias *map[string]string = &map[string]string{}
	tsConfigPath := filepath.Join(rootPath, "tsconfig.json")

	// 检查 tsconfig.json 是否存在
	if _, err := os.Stat(tsConfigPath); os.IsNotExist(err) {
		return *alias // 如果文件不存在，返回空的 alias
	}

	// 解析 tsconfig.json
	parseTsConfig(tsConfigPath, rootPath, alias)

	return FormatAlias(*alias)
}

// 递归解析 tsconfig.json
func parseTsConfig(configPath, rootPath string, alias *map[string]string) {
	// 读取 tsconfig.json 文件内容
	data, err := utils.ReadFileContent(configPath)
	if err != nil {
		return
	}

	// 移除注释
	data = removeJSONComments(data)

	// 解析 tsconfig.json
	var tsConfig struct {
		Extends         string `json:"extends"`
		CompilerOptions struct {
			Paths map[string][]string `json:"paths"`
		} `json:"compilerOptions"`
	}

	// 解析 JSON 数据
	if err := json.Unmarshal([]byte(data), &tsConfig); err != nil {
		fmt.Printf("解析 tsconfig.json 失败: %v\n", err)
		return
	}

	// 如果存在 extends，递归解析父配置文件
	if tsConfig.Extends != "" {
		extendsPath := tsConfig.Extends
		if !filepath.IsAbs(extendsPath) {
			extendsPath = filepath.Join(filepath.Dir(configPath), extendsPath)
		}
		extendsPath = filepath.Clean(extendsPath)
		if _, err := os.Stat(extendsPath); err == nil {
			parseTsConfig(extendsPath, rootPath, alias)
		}
	}

	// 先默认读取第一个path，后续再考虑优化
	for key, paths := range tsConfig.CompilerOptions.Paths {
		(*alias)[key] = paths[0]
	}
}

// 格式化 alias，默认alias结尾带*，需要去掉
func FormatAlias(alias map[string]string) map[string]string {
	formattedAlias := make(map[string]string)
	for key, path := range alias {
		// 如果有星号(*)需要去掉，读取tsconfig的case可能有
		if strings.HasSuffix(key, "*") {
			key = strings.TrimSuffix(key, "*")
		}

		if strings.HasSuffix(path, "*") {
			path = strings.TrimSuffix(path, "*")
		}
		formattedAlias[key] = path
	}
	return formattedAlias
}

// 移除 JSON 文件中的注释
func removeJSONComments(data string) string {
	// 匹配单行注释 (//...)
	singleLineComment := regexp.MustCompile(`(?m)^\s*//.*$`)
	data = singleLineComment.ReplaceAllString(data, "")

	// TODO: 匹配多行注释 (/*...*/)

	// 移除多余的空行
	emptyLines := regexp.MustCompile(`(?m)^\s*\n`)
	data = emptyLines.ReplaceAllString(data, "")

	return data
}
