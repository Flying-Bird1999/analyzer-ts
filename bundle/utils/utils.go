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

// 检查路径是否包含后缀
func HasExtension(filePath string) bool {
	return strings.Contains(filepath.Base(filePath), ".")
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

// 解析npm的真实路径
// 1，如果是npm包内部路径，则拼接上 ${rootPath}/node_modules/${npmFile} 即可
//   - 例如： import { IProductSetSearchParams } from '@sl/sc-product/dist/types/src/ProductSetPicker/type';
//
// 2. 如果是从npm包名直接导入，则需要进行依赖分析，找到真实的路径
//   - 例如： import { IProductSetSearchParams } from '@sl/sc-product';
//
// 2.1 找到 @sl/sc-product 对应的 package.json 路径
// 2.3 如果是查找类型，则检查 package.json 的 types / typing 字段。
// 2.4 如果是查找模块，则检查 package.json 的 main 字段，如果没有则默认 index.js。
// 2.3 拼接上入口文件，返回真实路径
func ResolveNpmPath(rootPath string, npmFile string, isImportTsType bool) string {
	// 1. 检查是否是 npm 包内部路径
	if strings.Contains(npmFile, "/") {
		// 拼接路径 ${rootPath}/node_modules/${npmFile}
		return filepath.Join(rootPath, "node_modules", npmFile)
	}

	// 2. 如果是直接导入 npm 包名
	packageJsonPath := filepath.Join(rootPath, "node_modules", npmFile, "package.json")
	if _, err := os.Stat(packageJsonPath); os.IsNotExist(err) {
		// 如果 package.json 不存在，返回默认路径
		return filepath.Join(rootPath, "node_modules", npmFile)
	}

	// 2.1 解析 package.json 文件
	packageJson, err := ReadPackageJson(packageJsonPath)
	if err != nil {
		fmt.Printf("解析 package.json 文件失败: %v\n", err)
		return filepath.Join(rootPath, "node_modules", npmFile)
	}

	// 2.3 如果是查找类型，检查 types / typings 字段
	if isImportTsType {
		if typesPath, exists := packageJson["types"]; exists {
			return filepath.Join(rootPath, "node_modules", npmFile, typesPath)
		}
		if typingsPath, exists := packageJson["typings"]; exists {
			return filepath.Join(rootPath, "node_modules", npmFile, typingsPath)
		}
	}

	// 2.4 如果是查找JS模块，检查 main 字段
	if mainPath, exists := packageJson["main"]; exists {
		return filepath.Join(rootPath, "node_modules", npmFile, mainPath)
	}

	// 如果都没有，默认返回 index.js
	return filepath.Join(rootPath, "node_modules", npmFile, "index.js")
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
