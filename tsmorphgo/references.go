package tsmorphgo

import (
	"context"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/lsp"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"
)

// FindReferences 查找给定节点所代表的符号的所有引用。
// 注意：此功能依赖的底层 `typescript-go` 库可能存在 bug，导致结果不完全准确。
func FindReferences(node Node) ([]*Node, error) {
	// 1. 获取节点的位置信息
	startLine := node.GetStartLineNumber()
	_, startChar := utils.GetLineAndCharacterOfPosition(node.GetSourceFile().fileResult.Raw, node.Pos())
	startChar += 1 // a a new field to store the ast of the file
	filePath := node.GetSourceFile().filePath

	// 2. 创建并使用 lsp.Service
	// 注意：每次调用都创建一个新服务可能效率不高，未来可以优化为在 Project 级别缓存。
	sources := make(map[string]any, len(node.GetSourceFile().project.parserResult.Js_Data))
	for k, v := range node.GetSourceFile().project.parserResult.Js_Data {
		sources[k] = v.Raw
	}
	q, err := lsp.NewServiceForTest(sources)
	if err != nil {
		return nil, err
	}
	defer q.Close()

	resp, err := q.FindReferences(context.Background(), filePath, startLine, startChar)
	if err != nil {
		return nil, err
	}

	// 3. 将返回的 LSP 位置转换为 sdk.Node 列表
	var results []*Node
	if resp.Locations != nil {
		for _, loc := range *resp.Locations {
			// 清理和转换 file URI 到项目内的虚拟路径
			refPath := strings.TrimPrefix(string(loc.Uri), "file://")

			// 使用 project 上的辅助方法来根据位置查找节点
			foundNode := node.GetSourceFile().project.findNodeAt(refPath, int(loc.Range.Start.Line)+1, int(loc.Range.Start.Character)+1)
			if foundNode != nil {
				results = append(results, &Node{
					Node:       foundNode,
					sourceFile: node.GetSourceFile().project.sourceFiles[refPath],
				})
			}
		}
	}

	return results, nil
}
