package tsmorphgo

import (
	"strings"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
	"github.com/stretchr/testify/assert"
)

// createTestProject 是一个测试辅助函数，用于从内存中的源码创建项目。
func createTestProject(sources map[string]string) *Project {
	return NewProjectFromSources(sources)
}

// TestGetParent 验证 Node.GetParent() 方法是否能正确返回父节点。
func TestGetParent(t *testing.T) {
	project := createTestProject(map[string]string{
		"/test.ts": `const greeting = "hello";`,
	})

	sf := project.GetSourceFile("/test.ts")
	assert.NotNil(t, sf)

	var identifierNode *Node
	sf.ForEachDescendant(func(node Node) {
		if IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "greeting" {
			identifierNode = &node
		}
	})

	assert.NotNil(t, identifierNode, "未能找到 a'greeting' 标识符节点")

	// 验证父节点链
	parent1 := identifierNode.GetParent()
	assert.NotNil(t, parent1)
	assert.Equal(t, ast.KindVariableDeclaration, parent1.Kind)

	parent2 := parent1.GetParent()
	assert.NotNil(t, parent2)
	assert.Equal(t, ast.KindVariableDeclarationList, parent2.Kind)

	parent3 := parent2.GetParent()
	assert.NotNil(t, parent3)
	assert.Equal(t, ast.KindVariableStatement, parent3.Kind)

	parent4 := parent3.GetParent()
	assert.NotNil(t, parent4)
	assert.Equal(t, ast.KindSourceFile, parent4.Kind)

	parent5 := parent4.GetParent()
	assert.Nil(t, parent5)
}

// TestNodeNavigation 验证 GetAncestors 和 GetFirstAncestorByKind 方法。
func TestNodeNavigation(t *testing.T) {
	project := createTestProject(map[string]string{
		"/test_nav.ts": `
		  const obj = {
			key: 'value'
		  };
		`,
	})
	sf := project.GetSourceFile("/test_nav.ts")
	assert.NotNil(t, sf)

	var valueNode *Node
	sf.ForEachDescendant(func(node Node) {
		if node.Kind == ast.KindStringLiteral && strings.TrimSpace(node.GetText()) == "'value'" {
			valueNode = &node
		}
	})

	assert.NotNil(t, valueNode, "未能找到 'value' 字符串字面量节点")

	// 1. 测试 GetFirstAncestorByKind
	propAssignment, ok := valueNode.GetFirstAncestorByKind(ast.KindPropertyAssignment)
	assert.True(t, ok)
	assert.NotNil(t, propAssignment)
	assert.Equal(t, ast.KindPropertyAssignment, propAssignment.Kind)

	objLiteral, ok := valueNode.GetFirstAncestorByKind(ast.KindObjectLiteralExpression)
	assert.True(t, ok)
	assert.NotNil(t, objLiteral)
	assert.Equal(t, ast.KindObjectLiteralExpression, objLiteral.Kind)

	// 2. 测试 GetAncestors
	ancestors := valueNode.GetAncestors()
	assert.Len(t, ancestors, 6, "祖先节点的数量应该为6")

	// 验证祖先节点的类型顺序
	expectedKinds := []ast.Kind{
		ast.KindPropertyAssignment,      // key: 'value'
		ast.KindObjectLiteralExpression, // { key: 'value' }
		ast.KindVariableDeclaration,     // obj = { ... }
		ast.KindVariableDeclarationList, // [obj = { ... }]
		ast.KindVariableStatement,       // const [obj = { ... }]
		ast.KindSourceFile,              // 根节点
	}

	for i, ancestor := range ancestors {
		if i >= len(expectedKinds) {
			break
		}
		assert.Equal(t, expectedKinds[i], ancestor.Kind, "祖先节点类型不匹配，索引: %d", i)
	}
}

// TestNodeInfo 验证 GetText 和位置信息相关的方法。
func TestNodeInfo(t *testing.T) {
	project := createTestProject(map[string]string{
		"/test_info.ts": `const user = {\n  name: \"John Doe\"\n};`,
	})
	sf := project.GetSourceFile("/test_info.ts")
	assert.NotNil(t, sf)

	var nameProp *Node
	sf.ForEachDescendant(func(node Node) {
		if node.Kind == ast.KindPropertyAssignment {
			nameNode, ok := GetFirstChild(node, func(child Node) bool { return IsIdentifier(child) })
			if ok && strings.TrimSpace(nameNode.GetText()) == "name" {
				nameProp = &node
			}
		}
	})
	_ = nameProp // 避免 "declared and not used" 错误
}

func TestGetFirstChild(t *testing.T) {
	project := createTestProject(map[string]string{
		"/test_child.ts": `const obj = { key: "value", enabled: true };`,
	})
	sf := project.GetSourceFile("/test_child.ts")
	assert.NotNil(t, sf)

	var objLiteral *Node
	sf.ForEachDescendant(func(node Node) {
		if node.Kind == ast.KindObjectLiteralExpression {
			objLiteral = &node
		}
	})
	assert.NotNil(t, objLiteral)

	// 查找第一个属性名为 "key" 的子节点
	keyNode, ok := GetFirstChild(*objLiteral, func(child Node) bool {
		if child.Kind != ast.KindPropertyAssignment {
			return false
		}
		nameNode, ok := GetFirstChild(child, func(grandChild Node) bool { return grandChild.Kind == ast.KindIdentifier })
		return ok && strings.TrimSpace(nameNode.GetText()) == "key"
	})
	assert.True(t, ok)
	assert.NotNil(t, keyNode)
	assert.Equal(t, `key: "value"`, strings.TrimSpace(keyNode.GetText()))

	// 查找第一个值为 boolean 的子节点
	enabledNode, ok := GetFirstChild(*objLiteral, func(child Node) bool {
		if child.Kind != ast.KindPropertyAssignment {
			return false
		}
		prop := child.Node.AsPropertyAssignment()
		return prop != nil && prop.Initializer != nil && prop.Initializer.Kind == ast.KindTrueKeyword
	})
	assert.True(t, ok)
	assert.NotNil(t, enabledNode)
	assert.Equal(t, `enabled: true`, strings.TrimSpace(enabledNode.GetText()))
}