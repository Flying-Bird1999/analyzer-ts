// package projectParser 包含项目级别解析的辅助工具函数。
package projectParser

import (
	"encoding/json"
	"fmt"
	"main/analyzer/scanProject"
	"main/analyzer/utils"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// FindAllTsConfigsAndAliases 在给定的根路径下查找所有名为 "tsconfig.json" 的文件，
// 并为每一个文件解析其路径别名配置。
// 它会利用 scanProject 的能力来智能地忽略被 ignore 规则匹配的目录。
func FindAllTsConfigsAndAliases(rootPath string, ignore []string) map[string]map[string]string {
	allAliases := make(map[string]map[string]string)

	// 使用 scanProject 来获取所有未被忽略的文件列表
	scanner := scanProject.NewProjectResult(rootPath, ignore, true) // isMonorepo is true
	scanner.ScanProject()
	fileList := scanner.GetFileList()

	// 遍历文件列表，找到所有的 tsconfig.json
	for path, fileDetail := range fileList {
		if fileDetail.FileName == "tsconfig.json" {
			// 解析该 tsconfig 文件及其 `extends` 链
			aliases := readAliasRecursive(path, rootPath)
			if len(aliases) > 0 {
				// 使用 tsconfig 文件所在的目录作为键
				dir := filepath.Dir(path)
				allAliases[dir] = aliases
			}
		}
	}

	return allAliases
}

// --- tsconfig.json 路径别名解析 ---

// ReadAliasFromTsConfig 是解析路径别名的入口函数。
// 它从项目根目录下的 tsconfig.json 开始，递归地读取和合并所有 `extends` 链上的路径别名配置。
func ReadAliasFromTsConfig(rootPath string) map[string]string {
	return readAliasRecursive(filepath.Join(rootPath, "tsconfig.json"), rootPath)
}

// readAliasRecursive 递归地解析 tsconfig.json 文件。
// 它首先解析父配置文件（通过 `extends` 字段指定），然后将当前文件的别名配置覆盖到父配置之上。
func readAliasRecursive(configPath, rootPath string) map[string]string {
	// 检查 tsconfig.json 文件是否存在，如果不存在则返回空映射。
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return make(map[string]string)
	}

	// 解析当前 tsconfig 文件，获取其 `paths` 和 `extends` 字段。
	paths, extendsFile := parseSingleTsConfig(configPath)

	// 如果 `extends` 字段存在，则递归解析父配置文件。
	parentAlias := make(map[string]string)
	if extendsFile != "" {
		extendsPath := extendsFile
		// 将 `extends` 的相对路径转换为绝对路径。
		if !filepath.IsAbs(extendsPath) {
			extendsPath = filepath.Join(filepath.Dir(configPath), extendsFile)
		}
		parentAlias = readAliasRecursive(filepath.Clean(extendsPath), rootPath)
	}

	// 将当前文件的别名合并到父别名中。子配置会覆盖父配置中的同名别名。
	for key, path := range paths {
		parentAlias[key] = path
	}

	// 格式化最终的别名映射，移除路径中的星号。
	return FormatAlias(parentAlias)
}

// parseSingleTsConfig 解析单个 tsconfig.json 文件。
// 它不处理递归 `extends`，仅返回当前文件的 `paths` 别名和 `extends` 字段值。
func parseSingleTsConfig(configPath string) (map[string]string, string) {
	data, err := utils.ReadFileContent(configPath)
	if err != nil {
		return nil, ""
	}

	// data = sanitizeJson(data)

	// 定义一个结构体来匹配 tsconfig.json 的关键字段。
	var tsConfig struct {
		Extends         string `json:"extends"`
		CompilerOptions struct {
			Paths map[string][]string `json:"paths"`
		}
	}

	// 解析 JSON 数据。注意：为了简化，这里假设 JSON 是标准的，没有注释。
	// 一个更健壮的实现会先使用 sanitizeJson 清理 JSON。
	if err := json.Unmarshal([]byte(data), &tsConfig); err != nil {
		fmt.Printf("解析 tsconfig.json 失败: path:%s, err: %v\n", configPath, err)
		return nil, ""
	}

	// `paths` 的值是一个数组，这里我们只取每个别名对应的第一个路径。
	paths := make(map[string]string)
	for key, p := range tsConfig.CompilerOptions.Paths {
		if len(p) > 0 {
			paths[key] = p[0]
		}
	}

	return paths, tsConfig.Extends
}

// FormatAlias 格式化路径别名映射。
// 它会移除别名键和路径值末尾的 `/*` 或 `*`，以便于后续的路径替换。
func FormatAlias(alias map[string]string) map[string]string {
	formattedAlias := make(map[string]string)
	for key, path := range alias {
		key = strings.TrimSuffix(key, "/*")
		key = strings.TrimSuffix(key, "*")
		path = strings.TrimSuffix(path, "/*")
		path = strings.TrimSuffix(path, "*")
		formattedAlias[key] = path
	}
	return formattedAlias
}

// sanitizeJson (当前未使用) 是一个工具函数，用于清理 JSON 字符串。
// 它可以移除多行注释、单行注释以及对象和数组中末尾多余的逗号，
// 使得非标准的 JSON 文件（如 tsconfig.json）也能被成功解析。
func sanitizeJson(data string) string {
	// 移除多行注释 (/*...*/)
	multiLineComment := regexp.MustCompile(`/*[\s\S]*?\*/`)
	data = multiLineComment.ReplaceAllString(data, "")

	// 移除单行注释 (//...)
	singleLineComment := regexp.MustCompile(`(?m)//.*$`)
	data = singleLineComment.ReplaceAllString(data, "")

	// 移除对象中末尾的逗号
	trailingCommaObject := regexp.MustCompile(`,\s*}`)
	data = trailingCommaObject.ReplaceAllString(data, "}")

	// 移除数组中末尾的逗号
	trailingCommaArray := regexp.MustCompile(`,\s*]`)
	data = trailingCommaArray.ReplaceAllString(data, "]")

	return data
}

// --- 导入路径解析 ---

// MatchImportSource 是解析导入路径的核心函数。
// 它接收一个导入语句的路径，并尝试按照以下顺序将其解析为最终的来源信息：
// 1. 路径别名 (Alias)
// 2. 相对路径 (Relative Path)
// 3. NPM 包 (NPM Package)
func MatchImportSource(
	importerPath string,      // 包含导入语句的文件的绝对路径
	importPath string,      // 导入语句中的原始路径 (e.g., "@/components/Button", "./utils", "react")
	basePath string,        // 用于解析路径别名的基准目录 (通常是 tsconfig.json 所在的目录)
	alias map[string]string, // 从 tsconfig.json 解析出的路径别名映射
	extensions []string,     // 需要尝试的文件扩展名列表 (e.g., [".ts", ".tsx"])
) SourceData {
	// 1. 尝试解析为路径别名
	resolvedPath, isAliasMatch := resolveAlias(importPath, alias)
	if isAliasMatch {
		// 如果是别名匹配，则基于 basePath (tsconfig.json 所在目录) 来构建绝对路径。
		if finalPath, ok := resolveAsFile(filepath.Join(basePath, resolvedPath), extensions); ok {
			return SourceData{FilePath: finalPath, Type: "file"}
		}
	}

	// 2. 尝试解析为相对路径
	if isRelativePath(importPath) {
		// 将相对路径转换为绝对路径。
		absPath := filepath.Join(filepath.Dir(importerPath), importPath)
		if finalPath, ok := resolveAsFile(absPath, extensions); ok {
			return SourceData{FilePath: finalPath, Type: "file"}
		}
	}

	// 3. 如果既不是别名也不是相对路径，则假定为 NPM 包。
	return SourceData{
		FilePath: importPath, // 对于NPM包，保留原始路径
		NpmPkg:   extractNpmPackageName(importPath),
		Type:     "npm",
	}
}

// resolveAlias 检查给定路径是否与任何一个别名匹配。
// 如果匹配，它会用别名对应的真实路径替换掉别名部分，并返回替换后的路径和 true。
func resolveAlias(filePath string, alias map[string]string) (string, bool) {
	for key, realPath := range alias {
		if strings.HasPrefix(filePath, key) {
			return strings.Replace(filePath, key, realPath, 1), true
		}
	}
	return filePath, false
}

// isRelativePath 检查路径是否是相对路径（以 "./" 或 "../" 开头）。
func isRelativePath(path string) bool {
	return strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../")
}

// resolveAsFile 尝试将一个基本路径解析为一个实际存在的文件。
// 它会按以下顺序尝试：
// a) 路径本身是否就是一个文件（可能已包含扩展名）。
// b) 路径 + 列表中的每个扩展名。
// c) 将路径视为目录，并尝试其下的 index 文件（路径/index + 扩展名）。
func resolveAsFile(path string, extensions []string) (string, bool) {
	// a) 检查路径本身
	if _, err := os.Stat(path); err == nil {
		return path, true
	}

	// b) 尝试添加扩展名
	for _, ext := range extensions {
		fullPath := path + ext
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, true
		}
	}

	// c) 尝试作为目录下的 index 文件
	for _, ext := range extensions {
		fullPath := filepath.Join(path, "index"+ext)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, true
		}
	}

	return "", false
}

// extractNpmPackageName 从导入路径中提取NPM包的名称。
// 例如，"react/jsx-runtime" -> "react", "@scope/pkg/sub" -> "@scope/pkg"。
func extractNpmPackageName(path string) string {
	parts := strings.Split(path, "/")
	// 处理带 scope 的包，例如 @babel/core
	if len(parts) > 0 && strings.HasPrefix(parts[0], "@") && len(parts) > 1 {
		return parts[0] + "/" + parts[1]
	}
	return parts[0]
}

// --- package.json 解析 ---

// PackageJsonInfo 存储从 package.json 文件中解析出的关键信息。
type PackageJsonInfo struct {
	Name    string
	Version string
	NpmList map[string]NpmItem
}

// GetPackageJson 解析指定的 package.json 文件。
// 它读取文件内容，提取包名、版本以及所有类型的依赖，
// 并尝试获取每个依赖在 node_modules 中的实际安装版本。
func GetPackageJson(packageJsonPath string) (*PackageJsonInfo, error) {
	// 检查文件是否存在
	if _, err := os.Stat(packageJsonPath); os.IsNotExist(err) {
		fmt.Printf("package.json 文件不存在: %s\n", packageJsonPath)
		return nil, err
	}

	// 读取文件内容
	data, err := utils.ReadFileContent(packageJsonPath)
	if err != nil {
		fmt.Printf("读取 package.json 文件失败: %s\n", err)
		return nil, err
	}

	// 定义用于解析 JSON 的匿名结构体
	var packageJson struct {
		Name             string            `json:"name"`
		Version          string            `json:"version"`
		Dependencies     map[string]string `json:"dependencies"`
		DevDependencies  map[string]string `json:"devDependencies"`
		PeerDependencies map[string]string `json:"peerDependencies"`
	}

	// 解析 JSON
	if err := json.Unmarshal([]byte(data), &packageJson); err != nil {
		fmt.Printf("解析 package.json 文件失败: %s\n", err)
		return nil, err
	}

	info := &PackageJsonInfo{
		Name:    packageJson.Name,
		Version: packageJson.Version,
		NpmList: make(map[string]NpmItem),
	}

	// 遍历所有类型的依赖，填充 NpmList
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

// getPackageRealVersion 尝试获取在 node_modules 中实际安装的NPM包的版本号。
// 它通过查找相对于当前 `package.json` 目录的 `node_modules/<packageName>/package.json` 文件来实现。
func getPackageRealVersion(packageJsonPath string, packageName string) string {
	nodeModuleVersion := ""
	packageDir := filepath.Dir(packageJsonPath)
	nodeModulePkgJson := filepath.Join(packageDir, "node_modules", packageName, "package.json")

	// 读取并解析 node_modules 中包的 package.json
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
