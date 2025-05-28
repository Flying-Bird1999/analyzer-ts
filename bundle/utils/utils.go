package utils

import (
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
