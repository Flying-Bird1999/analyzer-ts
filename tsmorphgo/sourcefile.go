package tsmorphgo

import (
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// SourceFile 代表一个源文件及其所有分析结果。
type SourceFile struct {
	filePath string
	// projectParser 为该文件生成的、已处理过的高级结果
	fileResult *projectParser.JsFileParserResult
	// astNode 是该文件的 AST 根节点
	astNode *ast.Node
	project *Project // 指向所属项目

	// nodeResultMap 从 ast.Node 指针快速定位到其对应的、被 parser 解析出的具体结果结构体
	nodeResultMap map[*ast.Node]interface{}
}

// GetFilePath 返回此源文件的绝对路径。
func (sf *SourceFile) GetFilePath() string {
	return sf.filePath
}

// ForEachDescendant 深度优先遍历该文件的所有后代节点。
// 它为每个访问到的节点调用提供的回调函数。
func (sf *SourceFile) ForEachDescendant(callback func(node Node)) {
	var walk func(node *ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}
		// 调用回调
		callback(Node{Node: node, sourceFile: sf})

		// 递归遍历子节点
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(sf.astNode)
}

// buildNodeResultMap 遍历文件解析结果，构建 ast.Node 到其具体结果结构体的映射。
func (sf *SourceFile) buildNodeResultMap() {
	if sf.fileResult == nil {
		return
	}

	// 注意：我们在这里存储的是值的副本，因为 fileResult 中的切片成员本身就是值类型。
	// 这对于并发访问是安全的，并且因为这些结构体不大，性能影响可以忽略不计。

	for _, decl := range sf.fileResult.ImportDeclarations {
		if decl.Node != nil {
			sf.nodeResultMap[decl.Node] = decl
		}
	}

	for _, decl := range sf.fileResult.ExportDeclarations {
		if decl.Node != nil {
			sf.nodeResultMap[decl.Node] = decl
		}
	}

	for _, decl := range sf.fileResult.ExportAssignments {
		if decl.Node != nil {
			sf.nodeResultMap[decl.Node] = decl
		}
	}

	for _, decl := range sf.fileResult.InterfaceDeclarations {
		if decl.Node != nil {
			sf.nodeResultMap[decl.Node] = decl
		}
	}

	for _, decl := range sf.fileResult.TypeDeclarations {
		if decl.Node != nil {
			sf.nodeResultMap[decl.Node] = decl
		}
	}

	for _, decl := range sf.fileResult.EnumDeclarations {
		if decl.Node != nil {
			sf.nodeResultMap[decl.Node] = decl
		}
	}

	for _, decl := range sf.fileResult.VariableDeclarations {
		if decl.Node != nil {
			sf.nodeResultMap[decl.Node] = decl
		}
	}

	for _, decl := range sf.fileResult.CallExpressions {
		if decl.Node != nil {
			sf.nodeResultMap[decl.Node] = decl
		}
	}

	for _, decl := range sf.fileResult.JsxElements {
		if decl.Node != nil {
			sf.nodeResultMap[decl.Node] = decl
		}
	}

	for _, decl := range sf.fileResult.FunctionDeclarations {
		if decl.Node != nil {
			sf.nodeResultMap[decl.Node] = decl
		}
	}

	for _, decl := range sf.fileResult.ExtractedNodes.AnyDeclarations {
		if decl.Node != nil {
			sf.nodeResultMap[decl.Node] = decl
		}
	}

	for _, decl := range sf.fileResult.ExtractedNodes.AsExpressions {
		if decl.Node != nil {
			sf.nodeResultMap[decl.Node] = decl
		}
	}

	// ReturnStatementResult 是在 FunctionDeclarationResult 内部，暂时不直接映射
}
