package tsmorphgo

import (
	"context"
	"strings"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/lsp/lsproto"
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

	// 2. 从项目获取共享的 lsp.Service
	lspService, err := node.GetSourceFile().project.getLspService()
	if err != nil {
		return nil, err
	}

	resp, err := lspService.FindReferences(context.Background(), filePath, startLine, startChar)
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

// GotoDefinition 查找给定节点所代表的符号的定义位置。
// 此功能通过 LSP 服务提供精确的跳转到定义能力。
func GotoDefinition(node Node) ([]*Node, error) {
	// 1. 获取节点的位置信息
	startLine := node.GetStartLineNumber()
	_, startChar := utils.GetLineAndCharacterOfPosition(node.GetSourceFile().fileResult.Raw, node.Pos())
	startChar += 1
	filePath := node.GetSourceFile().filePath

	// 2. 从项目获取共享的 lsp.Service
	lspService, err := node.GetSourceFile().project.getLspService()
	if err != nil {
		return nil, err
	}

	// 3. 使用 LSP 服务查找定义
	resp, err := lspService.GotoDefinition(context.Background(), filePath, startLine, startChar)
	if err != nil {
		return nil, err
	}

	// 4. 将返回的 LSP 位置转换为 Node 列表
	var results []*Node

	// 处理定义响应
	if resp.Locations != nil {
		// 处理 Location 数组
		for _, loc := range *resp.Locations {
			if converted := convertLspLocationToNode(loc, node.GetSourceFile().project); converted != nil {
				results = append(results, converted)
			}
		}
	}

	return results, nil
}

// convertLspLocationToNode 辅助函数：将 LSP Location 转换为 Node
func convertLspLocationToNode(loc lsproto.Location, project *Project) *Node {
	// 清理 file URI 到项目内的虚拟路径
	refPath := strings.TrimPrefix(string(loc.Uri), "file://")

	// 使用 project 上的辅助方法来根据位置查找节点
	foundNode := project.findNodeAt(refPath, int(loc.Range.Start.Line)+1, int(loc.Range.Start.Character)+1)
	if foundNode != nil {
		return &Node{
			Node:       foundNode,
			sourceFile: project.sourceFiles[refPath],
		}
	}
	return nil
}
