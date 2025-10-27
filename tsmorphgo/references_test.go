package tsmorphgo

import (
	"strings"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
	"github.com/stretchr/testify/assert"
)

// references_test.go
//
// 这个文件包含了符号引用查找功能的测试用例，专注于验证 tsmorphgo 在 TypeScript
// 代码中查找符号所有引用位置的能力。这是代码重构、导航和分析的核心功能。
//
// 主要功能：
// FindReferences 提供了查找 TypeScript 符号在项目中所有引用位置的能力，
// 类似于 IDE 中的 "Find All References" 功能，对于代码重构和依赖分析非常重要。
//
// 主要测试场景：
// 1. 跨文件引用查找 - 在多文件项目中查找符号的所有引用
// 2. 路径别名处理 - 验证对 TypeScript 路径别名（如 @/*）的正确处理
// 3. 定义处识别 - 确保能够正确识别符号的定义位置
// 4. 导入语句识别 - 验证能够识别 import 语句中的符号引用
// 5. 使用处识别 - 测试在表达式和语句中的符号使用识别
// 6. 项目配置支持 - 验证对 tsconfig.json 等项目配置的支持
//
// 测试目标：
// - 验证跨文件引用查找的准确性
// - 确保能够处理复杂的导入导出关系
// - 测试对 TypeScript 项目配置的正确理解
// - 验证引用查找功能的完整性和可靠性
//
// 核心 API 测试：
// - FindReferences() - 查找指定符号节点的所有引用位置
//
// 技术挑战：
// 该功能需要正确处理 TypeScript 的模块系统、路径映射、类型解析等复杂特性，
// 同时还要保证在大型项目中的性能表现。

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