package analyze

import (
	"encoding/json"
	"fmt"
	"main/bundle/utils"
	"os"
	"path/filepath"
	"strings"
)

// 新方法：从 tsconfig.json 中读取 alias
func ReadAliasFromTsConfig(rootPath string) map[string]string {
	alias := make(map[string]string)
	tsConfigPath := filepath.Join(rootPath, "tsconfig.json")

	// 检查 tsconfig.json 是否存在
	if _, err := os.Stat(tsConfigPath); os.IsNotExist(err) {
		return alias // 如果文件不存在，返回空的 alias
	}

	// 解析 tsconfig.json
	parseTsConfig(tsConfigPath, rootPath, alias)

	return alias
}

// 递归解析 tsconfig.json
func parseTsConfig(configPath, rootPath string, alias map[string]string) {
	// 读取 tsconfig.json 文件内容
	data, err := utils.ReadFileContent(configPath)
	if err != nil {
		return
	}

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

	// 合并当前配置文件的 paths 到 alias
	for key, paths := range tsConfig.CompilerOptions.Paths {
		// 如果有星号(*)需要去掉，读取tsconfig的case可能有
		if strings.HasSuffix(key, "*") {
			key = strings.TrimSuffix(key, "*")
		}

		// 先读取第一个path即可，后续再考虑优化
		if len(paths) > 0 {
			if strings.HasSuffix(paths[0], "*") {
				realPath := strings.TrimSuffix(paths[0], "*")
				alias[key] = realPath
				break
			}
			alias[key] = paths[0]
		}
	}
}
