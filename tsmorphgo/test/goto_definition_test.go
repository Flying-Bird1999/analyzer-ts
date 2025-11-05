package tsmorphgo

import (
	"strings"
	"testing"

	. "github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
	"github.com/stretchr/testify/assert"
)

// TestGotoDefinitionBasic 测试基本的跳转到定义功能
func TestGotoDefinitionBasic(t *testing.T) {
	project := createTestProject(map[string]string{
		"/test.ts": `
			const message = "Hello World";
			function greet() {
				console.log(message);
			}
			export { greet };
		`,
	})
	sf := project.GetSourceFile("/test.ts")
	assert.NotNil(t, sf)

	// 找到函数调用中的 'message' 标识符
	var usageNode *Node
	sf.ForEachDescendant(func(node Node) {
		if IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "message" {
			// 检查父节点是函数调用，确认这是 usage 而不是 definition
			parent := node.GetParent()
			if parent != nil && parent.Kind == ast.KindCallExpression {
				usageNode = &node
			}
		}
	})

	assert.NotNil(t, usageNode, "应该找到 message 的使用位置")

	// 测试跳转到定义
	definitions, err := GotoDefinition(*usageNode)
	assert.NoError(t, err, "跳转到定义应该成功")
	assert.Len(t, definitions, 1, "应该找到一个定义位置")

	// 验证定义位置
	definition := definitions[0]
	assert.NotNil(t, definition)
	assert.True(t, IsIdentifier(*definition), "定义应该是标识符")
	assert.Equal(t, "message", strings.TrimSpace(definition.GetText()), "定义应该是 message 变量")

	// 验证定义位置是变量声明，而不是使用
	parent := definition.GetParent()
	assert.NotNil(t, parent)
	assert.Equal(t, ast.KindVariableDeclaration, parent.Kind, "定义应该在变量声明中")
}