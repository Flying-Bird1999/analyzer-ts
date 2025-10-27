package tsmorphgo

import (
	"strings"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
	"github.com/stretchr/testify/assert"
)

func TestFindReferences(t *testing.T) {
	// 1. 创建一个包含 tsconfig.json 和路径别名的项目
	project := createTestProject(map[string]string{
		"/tsconfig.json": `{
			"compilerOptions": {
				"baseUrl": ".",
				"paths": {
					"@/*": ["src/*"]
				}
			}
		}`,
		"/src/utils.ts": `export const myVar = 123;`,
		"/src/index.ts": `
			import { myVar } from '@/utils';
			console.log(myVar);
		`,
	})

	// 2. 找到使用处的节点
	indexFile := project.GetSourceFile("/src/index.ts")
	assert.NotNil(t, indexFile)

	var usageNode *Node
	indexFile.ForEachDescendant(func(node Node) {
		// 找到 console.log(myVar) 中的 myVar
		if IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "myVar" {
			if parent := node.GetParent(); parent != nil && parent.Kind == ast.KindCallExpression {
				usageNode = &node
			}
		}
	})
	assert.NotNil(t, usageNode, "未能找到 myVar 的使用节点")

	// 3. 执行 FindReferences
	refs, err := FindReferences(*usageNode)
	assert.NoError(t, err)

	// 4. 验证结果
	t.Logf("FindReferences found %d locations:", len(refs))
	for _, refNode := range refs {
		t.Logf("  - Path: %s, Line: %d, Text: [%s]", refNode.GetSourceFile().filePath, refNode.GetStartLineNumber(), refNode.GetText())
	}

	// 我们期望至少找到 3 个引用：定义、导入、使用
	assert.GreaterOrEqual(t, len(refs), 3, "期望至少找到 3 个引用")

	// 验证每个引用是否都正确
	locations := map[string]bool{
		"/src/utils.ts": false, // 定义处
		"/src/index.ts": false, // 导入和使用处
	}

	for _, refNode := range refs {
		path := refNode.GetSourceFile().filePath
		if _, ok := locations[path]; ok {
			assert.Equal(t, "myVar", strings.TrimSpace(refNode.GetText()))
			locations[path] = true
		}
	}

	for path, found := range locations {
		assert.True(t, found, "应该在 %s 文件中找到引用", path)
	}
}