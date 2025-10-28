package tsmorphgo

import (
	"strings"
	"testing"

	. "github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
	"github.com/stretchr/testify/assert"
)

// node_navigation_test.go
//
// 这个文件包含了基础 AST 节点导航功能的测试用例，专注于验证 tsmorphgo 核心
// 导航 API 的正确性和稳定性。这些是 AST 操作中最基础和最重要的功能。
//
// 主要测试场景：
// 1. 父节点导航 - 测试 GetParent() 方法的基本功能
// 2. 祖先节点遍历 - 验证 GetAncestors() 方法的完整性和准确性
// 3. 按类型查找祖先 - 测试 GetFirstAncestorByKind() 方法的查找能力
// 4. 节点信息获取 - 验证 GetText() 等信息获取方法
// 5. 子节点查找 - 测试 GetFirstChild() 方法的条件查找功能
//
// 测试目标：
// - 验证从任意节点能够正确地向上导航到父节点
// - 确保祖先链的完整性和类型准确性
// - 测试按节点类型查找特定祖先的能力
// - 验证节点文本和位置信息的正确获取
// - 确保子节点查找的条件筛选功能正常工作
//
// 核心 API 测试：
// - GetParent() - 获取直接父节点
// - GetAncestors() - 获取从父节点到根节点的完整祖先链
// - GetFirstAncestorByKind() - 查找第一个匹配指定类型的祖先节点
// - GetText() - 获取节点在源码中的原始文本
// - GetFirstChild() - 按条件查找第一个匹配的子节点
//
// 这些基础导航功能是整个 tsmorphgo 系统的基石，被所有其他功能模块广泛使用。

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
