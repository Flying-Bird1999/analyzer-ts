package utils

import (
	"fmt"
	"os"
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
