package projectParser

import (
	"encoding/json"
	"fmt"
	"main/analyzer/utils"
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

// 匹配别名
func IsHitAlias(filePath string, Alias map[string]string) (string, bool) {
	for alias, realPath := range Alias {
		// 检查路径是否以 alias 开头
		if strings.HasPrefix(filePath, alias) {
			return strings.Replace(filePath, alias, realPath, 1), true
		}
	}
	return filePath, false // 未命中别名
}

// 检查路径是否为相对路径
func isRelativePath(path string) bool {
	return strings.HasPrefix(path, "./") ||
		strings.HasPrefix(path, "../") ||
		(!filepath.IsAbs(path) && !strings.HasPrefix(path, "/"))
}

// 匹配导入的文件路径
func MatchImportSource(
	targetPath string, // 目标文件路径
	filePath string, // 导入的文件路径
	rootPath string, // 项目根目录
	// Npm map[string]scanProject.NpmItem, // npm列表
	Alias map[string]string, //	别名映射，key: 别名, value: 实际路径
	Extensions []string, // 扩展名列表，例如: [".ts", ".tsx",".js", ".jsx"]
) SourceData {
	// 匹配 alias，替换为真实的路径
	realPath, matched := IsHitAlias(filePath, Alias)

	// // 匹配 npm 包
	// for npmName, npmItem := range Npm {
	// 	// 检查 realPath 是否包含 npm 包名
	// 	if strings.HasPrefix(realPath, npmName) {
	// 		return SourceData{
	// 			FilePath: realPath,
	// 			NpmPkg:   npmItem.Name,
	// 			Type:     "npm",
	// 		}
	// 	}
	// }

	// 替换为真实的绝对路径
	if matched || !isRelativePath(filePath) {
		// 如果匹配到别名/非相对路径，基于项目根目录拼接
		realPath = filepath.Join(rootPath, realPath)
	} else {
		// 如果没有匹配到别名，基于当前文件目录进行拼接
		realPath, _ = filepath.Abs(filepath.Join(filepath.Dir(targetPath), realPath))
	}

	// 检查结尾是否有文件后缀，如果没有后缀，需要基于Extensions尝试去匹配
	if !utils.HasExtension(realPath, Extensions) {
		for _, ext := range Extensions {
			// 尝试直接拼接扩展名
			extendedPath := realPath + ext
			if _, err := os.Stat(extendedPath); err == nil {
				realPath = extendedPath
				break
			}
		}
		// 如果拼接上扩展名后还是没有找到文件，尝试在目录下查找
		for _, ext := range Extensions {
			extendedPath := realPath + "/index" + ext
			if _, err := os.Stat(extendedPath); err == nil {
				realPath = extendedPath
				break
			}
		}
	}

	// 5. 如果存在，则返回 SourceData
	if _, err := os.Stat(realPath); err == nil {
		return SourceData{
			FilePath: realPath,
			NpmPkg:   "",
			Type:     "file",
		}
	}

	return SourceData{
		FilePath: filePath,
		NpmPkg:   "",
		Type:     "npm",
	}
}

type PackageJsonInfo struct {
	Name    string
	Version string
	NpmList map[string]NpmItem
}

// 解析 package.json 文件，获取包的基本信息和依赖信息
// GetPackageJson 主要逻辑如下：
// 1. 检查 package.json 文件是否存在
// 2. 读取 package.json 文件内容
// 3. 解析 JSON，获取 name、version、dependencies、devDependencies、peerDependencies 字段
// 4. 对每个依赖（dependencies/devDependencies/peerDependencies）：
//   - 读取 node_modules 下对应包的 package.json，获取实际安装的版本号
//   - 组装 NpmItem，包含依赖名、类型、声明版本、实际安装版本
//
// 5. 返回包含所有依赖信息的 PackageJsonInfo 结构体
func GetPackageJson(packageJsonPath string) (*PackageJsonInfo, error) {
	// 检查文件是否存在
	if _, err := os.Stat(packageJsonPath); os.IsNotExist(err) {
		fmt.Printf("package.json 文件不存在: %s\n", packageJsonPath)
		return nil, err
	}

	// 读取 package.json 文件内容
	data, err := utils.ReadFileContent(packageJsonPath)
	if err != nil {
		fmt.Printf("读取 package.json 文件失败: %s\n", err)
		return nil, err
	}

	// 定义结构体解析 package.json
	var packageJson struct {
		Name             string            `json:"name"`
		Version          string            `json:"version"`
		Dependencies     map[string]string `json:"dependencies"`
		DevDependencies  map[string]string `json:"devDependencies"`
		PeerDependencies map[string]string `json:"peerDependencies"`
	}

	// 解析 JSON 数据
	if err := json.Unmarshal([]byte(data), &packageJson); err != nil {
		fmt.Printf("解析 package.json 文件失败: %s\n", err)
		return nil, err
	}

	info := &PackageJsonInfo{
		Name:    packageJson.Name,
		Version: packageJson.Version,
		NpmList: make(map[string]NpmItem),
	}

	// 将 npm 包添加到 NpmList
	for name, version := range packageJson.Dependencies {
		info.NpmList[name] = NpmItem{
			Name:              name,
			Type:              "dependencies",
			Version:           version,
			NodeModuleVersion: getPackageRealVersion(packageJsonPath, name),
		}
	}
	for name, version := range packageJson.DevDependencies {
		info.NpmList[name] = NpmItem{
			Name:              name,
			Type:              "devDependencies",
			Version:           version,
			NodeModuleVersion: getPackageRealVersion(packageJsonPath, name),
		}
	}
	for name, version := range packageJson.PeerDependencies {
		info.NpmList[name] = NpmItem{
			Name:              name,
			Type:              "peerDependencies",
			Version:           version,
			NodeModuleVersion: getPackageRealVersion(packageJsonPath, name),
		}
	}

	return info, nil
}

// 根据当前package.json的位置去读取当前目录下的node_modules对应的包的版本号
func getPackageRealVersion(packageJsonPath string, packageName string) string {
	nodeModuleVersion := ""
	packageDir := filepath.Dir(packageJsonPath)
	nodeModulePkgJson := filepath.Join(packageDir, "node_modules", packageName, "package.json")
	if data, err := utils.ReadFileContent(nodeModulePkgJson); err == nil {
		var modPkg struct {
			Version string `json:"version"`
		}
		if err := json.Unmarshal([]byte(data), &modPkg); err == nil {
			nodeModuleVersion = modPkg.Version
		}
	}
	return nodeModuleVersion
}
