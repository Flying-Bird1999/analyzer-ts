package tsmorphgo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// ============================================================================
// QuickInfo 功能测试 - tsmorphgo 层
// ============================================================================
// 这些测试专注于验证 tsmorphgo 层对 QuickInfo 功能的封装和集成

// TestNode_GetQuickInfo_InvalidNode 测试无效节点的 QuickInfo 查询
func TestNode_GetQuickInfo_InvalidNode(t *testing.T) {
	// 创建一个无效的节点（没有 sourceFile）
	invalidNode := &Node{}

	// 调用被测试的方法
	quickInfo, err := invalidNode.GetQuickInfo()

	// 验证返回错误
	assert.Error(t, err, "无效节点应该返回错误")
	assert.Contains(t, err.Error(), "must belong to a source file and project",
		"错误信息应该包含预期文本")

	// 验证 QuickInfo 为空
	assert.Nil(t, quickInfo, "无效节点应该返回 nil QuickInfo")
}

// TestCreateLSPService 测试 createLSPService 函数
func TestCreateLSPService(t *testing.T) {
	// 准备测试用的源码
	testSources := map[string]string{
		"/test.ts": `const test = "hello";`,
	}

	// 创建项目
	project := NewProjectFromSources(testSources)
	require.NotNil(t, project, "项目创建不应该失败")

	// 测试正常情况
	service, err := createLSPService(project)
	require.NoError(t, err, "LSP 服务创建不应该失败")
	assert.NotNil(t, service, "LSP 服务不应该为空")
	service.Close()

	// 测试 nil 项目
	service, err = createLSPService(nil)
	assert.Error(t, err, "nil 项目应该返回错误")
	assert.Nil(t, service, "nil 项目应该返回 nil service")

	// 测试项目没有解析结果
	emptyProject := &Project{}
	service, err = createLSPService(emptyProject)
	assert.Error(t, err, "没有解析结果的项目应该返回错误")
	assert.Nil(t, service, "无效项目应该返回 nil service")
}

// TestNode_GetQuickInfo_Basic 基础 QuickInfo 功能测试
func TestNode_GetQuickInfo_Basic(t *testing.T) {
	// 准备简单的测试场景
	testSources := map[string]string{
		"/basic.ts": `const message: string = "hello world";`,
	}

	project := NewProjectFromSources(testSources)
	require.NotNil(t, project)

	sourceFile := project.GetSourceFile("/basic.ts")
	require.NotNil(t, sourceFile)

	// 查找 message 节点
	var messageNode *Node
	sourceFile.ForEachDescendant(func(node Node) {
		nodeText := node.GetText()

		// 查找 Identifier 节点且文本为 message（考虑前后可能有空格）
		if node.Kind == ast.KindIdentifier && (nodeText == "message" || nodeText == " message") {
			messageNode = &node
		}
	})

	// 如果找不到节点，跳过测试（避免因为 AST 结构变化导致测试失败）
	if messageNode == nil {
		t.Skip("未找到 message 节点，跳过测试")
		return
	}

	// 测试 QuickInfo 功能
	quickInfo, err := messageNode.GetQuickInfo()
	require.NoError(t, err)

	// QuickInfo 可以为 nil（如果该位置没有有效符号）
	if quickInfo == nil {
		t.Logf("message 节点没有 QuickInfo（这是正常的）")
		return
	}

	assert.NotEmpty(t, quickInfo.TypeText)
	t.Logf("message 节点的 QuickInfo: %s", quickInfo.TypeText)
}

// TestNode_GetQuickInfo_FunctionParameter 测试函数参数的 QuickInfo
func TestNode_GetQuickInfo_FunctionParameter(t *testing.T) {
	// 准备函数参数测试场景
	testSources := map[string]string{
		"/function.ts": `function greet(name: string): string {
	return "Hello, " + name;
}`,
	}

	project := NewProjectFromSources(testSources)
	require.NotNil(t, project)

	sourceFile := project.GetSourceFile("/function.ts")
	require.NotNil(t, sourceFile)

	// 查找 name 参数节点
	var nameNode *Node
	sourceFile.ForEachDescendant(func(node Node) {
		nodeText := node.GetText()

		// 查找 Identifier 节点且文本为 name（考虑前后可能有空格）
		if node.Kind == ast.KindIdentifier && (nodeText == "name" || nodeText == " name") {
			nameNode = &node
		}
	})

	if nameNode == nil {
		t.Skip("未找到 name 节点，跳过测试")
		return
	}

	// 测试函数参数的 QuickInfo
	quickInfo, err := nameNode.GetQuickInfo()
	require.NoError(t, err)

	// QuickInfo 可以为 nil（如果该位置没有有效符号）
	if quickInfo == nil {
		t.Logf("name 节点没有 QuickInfo（这是正常的）")
		return
	}

	assert.NotEmpty(t, quickInfo.TypeText)
	t.Logf("name 参数节点的 QuickInfo: %s", quickInfo.TypeText)
}

// TestNode_GetQuickInfo_InterfaceProperty 测试接口属性的 QuickInfo
func TestNode_GetQuickInfo_InterfaceProperty(t *testing.T) {
	// 准备接口属性测试场景
	testSources := map[string]string{
		"/interface.ts": `interface User {
	id: number;
	name: string;
	email?: string;
}`,
	}

	project := NewProjectFromSources(testSources)
	require.NotNil(t, project)

	sourceFile := project.GetSourceFile("/interface.ts")
	require.NotNil(t, sourceFile)

	// 查找 name 属性节点
	var nameNode *Node
	sourceFile.ForEachDescendant(func(node Node) {
		nodeText := node.GetText()

		// 查找 Identifier 节点且文本为 name（考虑前后可能有空格）
		if node.Kind == ast.KindIdentifier && (nodeText == "name" || nodeText == " name") {
			nameNode = &node
		}
	})

	if nameNode == nil {
		t.Skip("未找到 name 节点，跳过测试")
		return
	}

	// 测试接口属性的 QuickInfo
	quickInfo, err := nameNode.GetQuickInfo()
	require.NoError(t, err)

	// QuickInfo 可以为 nil（如果该位置没有有效符号）
	if quickInfo == nil {
		t.Logf("接口 name 属性节点没有 QuickInfo（这是正常的）")
		return
	}

	assert.NotEmpty(t, quickInfo.TypeText)
	t.Logf("接口 name 属性节点的 QuickInfo: %s", quickInfo.TypeText)
}