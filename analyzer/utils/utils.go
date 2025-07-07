package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/core"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/parser"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/scanner"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/tspath"
)

// 读取文件内容
func ReadFileContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// 解析TypeScript文件为AST
func ParseTypeScriptFile(filePath string, sourceCode string) *ast.SourceFile {
	// 创建路径对象
	path := tspath.Path(filePath)

	// 使用ParseSourceFile函数解析源代码
	sourceFile := parser.ParseSourceFile(
		filePath,
		path,
		sourceCode,
		core.ScriptTargetES2015,
		scanner.JSDocParsingModeParseAll,
	)

	return sourceFile
}

// 获取ast节点的原始源代码文本
func GetNodeText(node *ast.Node, sourceCode string) string {
	start := node.Pos()
	end := node.End()
	if start >= 0 && end >= start && end <= len(sourceCode) {
		return sourceCode[start:end]
	}
	return ""
}

// 判断是否为基本类型
func IsBasicType(typeName string) bool {
	basicTypes := []string{
		"string", "number", "boolean", "any", "void", "null", "undefined",
		"object", "unknown", "never", "bigint", "symbol", "Function",
		"Date", "RegExp", "Error", "Array", "Map", "Set", "Promise",
	}

	for _, basicType := range basicTypes {
		if strings.EqualFold(typeName, basicType) {
			return true
		}
	}

	return false
}

// 打印简易版AST结构
func PrintAST(node *ast.Node) {
	if node == nil {
		return
	}

	// 打印当前节点信息
	fmt.Printf("当前节点 Kind: %s\n", node.Kind)
	if ast.IsIdentifier(node) {
		fmt.Printf(" Text: %s\n", node.Text())
	}
	fmt.Printf("\n")

	// 递归打印子节点
	node.ForEachChild(func(child *ast.Node) bool {
		PrintAST(child)
		return false // 继续遍历
	})
}

// 判断切片中是否包含指定元素
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// 创建文件，写入内容
func WriteResultToFile(filePath string, result string) error {
	// 打开或创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(result)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	fmt.Printf("文件已生成: %s\n", filePath)
	return nil
}

// 检查路径是否包含已知扩展名
func HasExtension(filePath string, extensions []string) bool {
	base := filepath.Base(filePath)
	for _, ext := range extensions {
		if strings.HasSuffix(base, ext) {
			return true
		}
	}
	return false
}

// 根据基础路径和扩展名列表查找真实存在的文件路径
func FindRealFilePath(basePath string, extensions []string) string {
	// 先尝试 basePath + ext
	for _, ext := range extensions {
		extendedPath := basePath + ext
		if _, err := os.Stat(extendedPath); err == nil {
			return extendedPath
		}
	}
	// 再尝试 basePath + "/index" + ext
	for _, ext := range extensions {
		extendedPath := basePath + "/index" + ext
		if _, err := os.Stat(extendedPath); err == nil {
			return extendedPath
		}
	}
	// 如果都找不到，返回原始路径
	return basePath
}

// 向上查找最近的 node_modules 并判断是否含有目标包
func findNodeModulesWithPackage(startPath string, npmFile string) (string, bool) {
	currPath := startPath
	for {
		nodeModulesPath := filepath.Join(currPath, "node_modules")
		packagePath := filepath.Join(nodeModulesPath, npmFile)
		if stat, err := os.Stat(packagePath); err == nil && stat.IsDir() {
			return nodeModulesPath, true
		}
		parent := filepath.Dir(currPath)
		if parent == currPath {
			break
		}
		currPath = parent
	}
	return "", false
}

// ResolveNpmPath 解析 npm 包的真实路径，支持从当前路径向上查找 node_modules。
// curPath: 当前文件所在路径，用于向上查找 node_modules。
// rootPath: 项目根目录路径，作为查找 node_modules 的备用路径。
// npmFile: 导入的 npm 包名或包内文件路径。
// isImportTsType: 是否为导入 TypeScript 类型，用于判断查找 package.json 中的 types 或 typings 字段。
func ResolveNpmPath(curPath string, rootPath string, npmFile string, isImportTsType bool) string {
	// 先尝试从 curPath 向上查找最近的包含目标 npm 包的 node_modules 目录
	node_modules_path, found := findNodeModulesWithPackage(curPath, npmFile)
	if !found {
		// 若未找到，使用 rootPath 下的 node_modules 目录作为备用
		node_modules_path = filepath.Join(rootPath, "node_modules")
	}

	// 1. 检查是否是 npm 包内部路径（即包含斜杠 /）
	if strings.Contains(npmFile, "/") {
		// 若是包内部路径，直接拼接 node_modules 路径和 npmFile 并返回
		return filepath.Join(node_modules_path, npmFile)
	}

	// 2. 如果是直接导入 npm 包名
	packageJsonPath := filepath.Join(node_modules_path, npmFile, "package.json")
	// 检查 package.json 文件是否存在
	if _, err := os.Stat(packageJsonPath); os.IsNotExist(err) {
		// 若 package.json 不存在，返回 node_modules 下该 npm 包的路径
		return filepath.Join(node_modules_path, npmFile)
	}

	// 读取 package.json 文件内容
	packageJson, err := ReadPackageJson(packageJsonPath)
	if err != nil {
		// 若读取失败，打印错误信息并返回 node_modules 下该 npm 包的路径
		fmt.Printf("解析 package.json 文件失败: %v\n", err)
		return filepath.Join(node_modules_path, npmFile)
	}

	// 如果是导入 TypeScript 类型
	if isImportTsType {
		// 检查 package.json 中的 types 字段
		if typesPath, exists := packageJson["types"]; exists {
			// 若存在，拼接并返回对应路径
			return filepath.Join(node_modules_path, npmFile, typesPath)
		}
		// 检查 package.json 中的 typings 字段
		if typingsPath, exists := packageJson["typings"]; exists {
			// 若存在，拼接并返回对应路径
			return filepath.Join(node_modules_path, npmFile, typingsPath)
		}
	}

	// 检查 package.json 中的 main 字段
	if mainPath, exists := packageJson["main"]; exists {
		// 若存在，拼接并返回对应路径
		return filepath.Join(node_modules_path, npmFile, mainPath)
	}

	// 若以上字段都不存在，默认返回 node_modules 下该 npm 包的 index.js 文件路径
	return filepath.Join(node_modules_path, npmFile, "index.js")
}

// 辅助方法：读取 package.json 文件
func ReadPackageJson(packageJsonPath string) (map[string]string, error) {
	file, err := os.Open(packageJsonPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var packageJson map[string]string
	if err := json.NewDecoder(file).Decode(&packageJson); err != nil {
		return nil, err
	}

	return packageJson, nil
}
