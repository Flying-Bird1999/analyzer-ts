package demo

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/core"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/parser"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/scanner"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/tspath"
)

// 读取指定目录中的所有TypeScript文件
func ListTSFiles(dirPath string, recursive bool) ([]string, error) {
	// 创建一个正则表达式，匹配.ts和.tsx文件
	tsFileRegex := regexp.MustCompile(`\.(ts|tsx)$`)

	var files []string

	// 获取目录中的所有条目
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	// 遍历每个条目
	for _, entry := range entries {
		path := filepath.Join(dirPath, entry.Name())

		fmt.Printf("path: %s\n", path)

		if !entry.IsDir() {
			// 如果是文件且匹配正则表达式，则添加到文件列表
			if tsFileRegex.MatchString(path) {
				files = append(files, path)
			}
		} else if recursive {
			// 如果是目录且需要递归，则递归处理该目录
			subFiles, err := ListTSFiles(path, recursive)
			if err != nil {
				return nil, err
			}
			files = append(files, subFiles...)
		}
	}

	return files, nil
}

// 读取文件内容
func ReadFileContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// 解析TypeScript文件为AST
func ParseTypeScriptFile(filePath string, sourceText string) *ast.SourceFile {
	// 创建路径对象
	path := tspath.Path(filePath)

	// 使用ParseSourceFile函数解析源代码
	sourceFile := parser.ParseSourceFile(
		filePath,
		path,
		sourceText,
		core.ScriptTargetES2015,
		scanner.JSDocParsingModeParseAll,
	)

	return sourceFile
}

func main() {
	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("用法: go run main.go <目录路径> [是否递归(true/false)]")
		os.Exit(1)
	}

	// 获取目录路径
	dirPath := os.Args[1]

	// 确定是否递归
	recursive := true
	if len(os.Args) > 2 {
		recursive = strings.ToLower(os.Args[2]) == "true"
	}

	// 获取所有TypeScript文件
	files, err := ListTSFiles(dirPath, recursive)
	if err != nil {
		fmt.Printf("读取目录失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("找到 %d 个TypeScript文件\n\n", len(files))

	// 遍历每个文件并解析
	for _, file := range files {
		fmt.Printf("处理文件: %s\n", file)
		// 读取文件内容
		sourceText, err := ReadFileContent(file)
		if err != nil {
			fmt.Printf("读取文件失败: %v\n", err)
			continue
		}

		// 解析文件
		sourceFile := ParseTypeScriptFile(file, sourceText)

		// 打印AST的基本信息
		fmt.Printf("  文件名: %s\n", sourceFile.FileName)
		fmt.Printf("  语句数量: %d\n", len(sourceFile.Statements.Nodes))

		// 打印每个顶层语句的类型
		fmt.Println("\n顶层语句:")
		for i, stmt := range sourceFile.Statements.Nodes {
			fmt.Printf("  %d. 类型: %s\n", i+1, (stmt.Kind))
			// if stmt.Kind == ast.KindFunctionDeclaration {
			// 	// 转换为函数声明节点
			// 	funcDecl := stmt.AsFunctionDeclaration()
			// 	// 访问函数名
			// 	funcName := funcDecl.Name().AsIdentifier().Text
			// 	// 访问函数参数
			// 	params := funcDecl.Parameters
			// 	fmt.Printf("    funcName: %s\n", funcName)

			// 	for _, paramsNode := range params.Nodes {
			// 		fmt.Printf("入参name: %s\n", (paramsNode.AsParameterDeclaration().Name().AsIdentifier().Text))

			// 		typeNode := paramsNode.AsParameterDeclaration().Type
			// 		fmt.Printf("入参Kind: %s\n", typeNode.Kind)

			// 		var typeName string

			// 		if typeNode.Kind == ast.KindTypeReference {
			// 			typeName = typeNode.AsTypeReferenceNode().TypeName.AsIdentifier().Text
			// 		} else {
			// 			// 这个给力
			// 			typeName = scanner.TokenToString(typeNode.Kind)

			// 		}
			// 		fmt.Printf("入参type: %s\n", typeName)
			// 	}
			// }

			if stmt.Kind == ast.KindInterfaceDeclaration {
				fmt.Print("hahahha%v", stmt)
				// stmt.AsInferTypeNode()
				fmt.Printf("    Interface: %s\n", stmt.AsInterfaceDeclaration().Name().Text())
			}
		}

		// 检查是否有解析错误
		diagnostics := sourceFile.Diagnostics()
		if len(diagnostics) > 0 {
			fmt.Printf("  解析错误: %d 个\n", len(diagnostics))
			for i, diag := range diagnostics {
				if i < 3 { // 只显示前3个错误
					fmt.Printf("    - %s\n", diag.Message())
				} else {
					fmt.Printf("    - ... 还有 %d 个错误\n", len(diagnostics)-3)
					break
				}
			}
		} else {
			fmt.Println("  解析成功，没有错误")
		}
	}
}
