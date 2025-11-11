package tsmorphgo_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	. "github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

// TestNode_BasicAPIs 测试 Node 基础信息获取 API
// 测试 API: GetSourceFile(), GetText(), IsValid()
func TestNode_BasicAPIs(t *testing.T) {
	project := NewProjectFromSources(map[string]string{
		"/test.ts": `
			const x = 42;
			function test(param: string): number {
				return x;
			}
			console.log("hello");
		`,
	})
	defer project.Close()

	sf := project.GetSourceFile("/test.ts")
	require.NotNil(t, sf)

	// 测试获取源文件
	sf.ForEachDescendant(func(node Node) {
		sourceFile := node.GetSourceFile()
		assert.NotNil(t, sourceFile)
		assert.Equal(t, sf.GetFilePath(), sourceFile.GetFilePath())
	})

	// 测试获取文本内容
	var identifiers []Node
	sf.ForEachDescendant(func(node Node) {
		if node.IsIdentifier() {
			identifiers = append(identifiers, node)
			text := strings.TrimSpace(node.GetText())
			assert.NotEmpty(t, text)
			assert.Contains(t, []string{"x", "test", "param", "console", "log"}, text)
		}
	})

	assert.Greater(t, len(identifiers), 0, "应该找到标识符")
}

// TestNode_PositionAPIs 测试 Node 位置信息 API
// 测试 API: GetStartLineNumber(), GetEndLineNumber(), GetStartColumnNumber(),
//
//	GetEndColumnNumber(), GetStart(), GetEnd(), GetStartLinePos(), GetWidth()
func TestNode_PositionAPIs(t *testing.T) {
	source := `const x = 1;
const y = 2;`

	project := NewProjectFromSources(map[string]string{
		"/test.ts": source,
	})
	defer project.Close()

	sf := project.GetSourceFile("/test.ts")
	require.NotNil(t, sf)

	// 查找 x 变量
	var xNode *Node
	sf.ForEachDescendant(func(node Node) {
		text := strings.TrimSpace(node.GetText())
		if text == "x" {
			xNode = &node
		}
	})
	require.NotNil(t, xNode, "应该找到 x 变量")

	// 验证位置信息
	assert.Equal(t, 1, xNode.GetStartLineNumber(), "x 应该在第1行")
	assert.Equal(t, 1, xNode.GetEndLineNumber(), "x 应该在第1行结束")
	assert.Greater(t, xNode.GetStartColumnNumber(), 0, "x 的起始列号应该大于0")
	assert.Greater(t, xNode.GetEndColumnNumber(), xNode.GetStartColumnNumber(), "x 的结束列号应该大于起始列号")
	assert.Equal(t, 0, xNode.GetStartLinePos(), "行起始位置应该正确")
	assert.Equal(t, "x", strings.TrimSpace(xNode.GetText()), "文本应该是 x")
	assert.Greater(t, xNode.GetWidth(), 0, "宽度应该大于0")
	t.Logf("x 节点位置: 行 %d, 列 %d-%d, 宽度 %d",
		xNode.GetStartLineNumber(), xNode.GetStartColumnNumber(),
		xNode.GetEndColumnNumber(), xNode.GetWidth())

	// 查找 y 变量
	var yNode *Node
	sf.ForEachDescendant(func(node Node) {
		text := strings.TrimSpace(node.GetText())
		if text == "y" {
			yNode = &node
		}
	})
	require.NotNil(t, yNode, "应该找到 y 变量")

	// 验证位置信息
	assert.Equal(t, 2, yNode.GetStartLineNumber(), "y 应该在第2行")
	assert.Equal(t, "y", strings.TrimSpace(yNode.GetText()), "文本应该是 y")
}

// TestNode_NavigationAPIs 测试 Node 导航 API
// 测试 API: GetParent(), GetAncestors(), GetFirstAncestorByKind(),
//
//	GetChildren(), GetFirstChild(), ForEachChild()
func TestNode_NavigationAPIs(t *testing.T) {
	project := NewProjectFromSources(map[string]string{
		"/test.ts": `
			function outer() {
				function inner() {
					const x = 1;
					return x;
				}
				return inner();
			}
		`,
	})
	defer project.Close()

	sf := project.GetSourceFile("/test.ts")
	require.NotNil(t, sf)

	// 查找 inner 函数
	var innerFunctionNode *Node
	sf.ForEachDescendant(func(node Node) {
		if node.IsFunctionDeclaration() && strings.Contains(node.GetText(), "inner") {
			innerFunctionNode = &node
			t.Logf("找到inner函数: %s", node.GetText())
		}
	})
	require.NotNil(t, innerFunctionNode, "应该找到inner函数")

	// 查找 inner 函数中的 x 标识符
	var innerXNode *Node
	sf.ForEachDescendant(func(node Node) {
		if node.IsIdentifier() && node.GetText() == "x" {
			// 确保这个 x 在 inner 函数中
			ancestors := node.GetAncestors()
			for _, ancestor := range ancestors {
				if ancestor.IsFunctionDeclaration() && strings.Contains(ancestor.GetText(), "inner") {
					innerXNode = &node
					t.Logf("使用inner函数中的标识符: %s", node.GetText())
					break
				}
			}
		}
	})

	require.True(t, innerXNode != nil && innerXNode.IsValid(), "应该找到inner函数中的标识符")

	// 测试 getParent
	parent := innerXNode.GetParent()
	assert.True(t, parent.IsValid(), "x应该有父节点")

	// 测试 getAncestors
	ancestors := innerXNode.GetAncestors()
	assert.Greater(t, len(ancestors), 0, "x应该有祖先节点")

	// 测试 getFirstAncestorByKind
	ancestorFunc, found := innerXNode.GetFirstAncestorByKind(KindFunctionDeclaration)
	assert.True(t, found, "应该找到函数声明祖先")
	assert.True(t, ancestorFunc.IsValid(), "找到的祖先应该有效")

	// 测试 getChildren
	children := innerFunctionNode.GetChildren()
	assert.Greater(t, len(children), 0, "函数应该有子节点")

	// 测试 getFirstChild
	firstChild := innerFunctionNode.GetFirstChild(func(node Node) bool {
		return node.IsIdentifier()
	})
	if firstChild.IsValid() {
		assert.Equal(t, KindIdentifier, firstChild.GetKind())
	}

	// 测试 forEachChild
	childCount := 0
	innerFunctionNode.ForEachChild(func(child Node) bool {
		childCount++
		return false // 继续遍历，返回true会停止遍历
	})
	assert.Equal(t, len(children), childCount, "forEachChild应该遍历所有子节点")

	t.Logf("inner函数中的x节点找到了%d个祖先", len(ancestors))
}

// TestNode_TypeCheckingAPIs 测试 Node 类型判断 API
// 测试 API: IsKind(), IsAnyKind(), IsIdentifier(), IsFunctionDeclaration(),
//
//	IsVariableDeclaration(), IsCallExpression(), IsPropertyAccessExpression(),
//	IsObjectLiteralExpression(), IsArrayLiteralExpression(), IsBinaryExpression()
func TestNode_TypeCheckingAPIs(t *testing.T) {
	project := NewProjectFromSources(map[string]string{
		"/test.ts": `
			const x = 1;
			function test() { return x; }
			const obj = { key: 'value' };
			const arr = [1, 2, 3];
			test();
			obj.key;
			x + 1;
		`,
	})
	defer project.Close()

	sf := project.GetSourceFile("/test.ts")
	require.NotNil(t, sf)

	var foundTypes map[SyntaxKind]bool = make(map[SyntaxKind]bool)

	sf.ForEachDescendant(func(node Node) {
		// 测试基础类型检查
		if node.IsVariableDeclaration() {
			foundTypes[KindVariableDeclaration] = true
		}
		if node.IsFunctionDeclaration() {
			foundTypes[KindFunctionDeclaration] = true
		}
		if node.IsObjectLiteralExpression() {
			foundTypes[KindObjectLiteralExpression] = true
		}
		if node.IsArrayLiteralExpression() {
			foundTypes[KindArrayLiteralExpression] = true
		}
		if node.IsCallExpression() {
			foundTypes[KindCallExpression] = true
		}
		if node.IsPropertyAccessExpression() {
			foundTypes[KindPropertyAccessExpression] = true
		}
		if node.IsBinaryExpression() {
			foundTypes[KindBinaryExpression] = true
		}
		if node.IsIdentifier() {
			foundTypes[KindIdentifier] = true
		}

		// 测试具体的文本匹配和类型判断
		text := strings.TrimSpace(node.GetText())
		if text == "key" && node.GetParent().IsKind(KindPropertyAssignment) {
			assert.True(t, node.IsIdentifier(), "key 应该是标识符")
			t.Logf("找到 key 节点: 类型=%s, 父节点类型=%s", node.GetKindName(), node.GetParent().GetKindName())
		}
		if text == "test" && node.IsKind(KindCallExpression) {
			assert.True(t, node.IsCallExpression(), "test 应该是调用表达式")
			t.Logf("找到 test 调用表达式: 类型=%s", node.GetKindName())
		}
	})

	// 验证找到了预期类型
	assert.True(t, foundTypes[KindVariableDeclaration], "应该找到变量声明")
	assert.True(t, foundTypes[KindFunctionDeclaration], "应该找到函数声明")
	assert.True(t, foundTypes[KindObjectLiteralExpression], "应该找到对象字面量")
	assert.True(t, foundTypes[KindArrayLiteralExpression], "应该找到数组字面量")
	assert.True(t, foundTypes[KindCallExpression], "应该找到调用表达式")
	assert.True(t, foundTypes[KindPropertyAccessExpression], "应该找到属性访问表达式")
	assert.True(t, foundTypes[KindBinaryExpression], "应该找到二元表达式")
	assert.True(t, foundTypes[KindIdentifier], "应该找到标识符")
}

// TestNode_ForEachDescendant 测试 Node 遍历 API
// 测试 API: ForEachDescendant()
func TestNode_ForEachDescendant(t *testing.T) {
	project := NewProjectFromSources(map[string]string{
		"/test.ts": `
			function outer() {
				const x = 1;
				function inner() {
					return x;
				}
				return inner();
			}
		`,
	})
	defer project.Close()

	sf := project.GetSourceFile("/test.ts")
	require.NotNil(t, sf)

	// 测试深度优先遍历
	var nodeCount int
	var identifiers []string
	var functions []string

	sf.ForEachDescendant(func(node Node) {
		nodeCount++
		if node.IsIdentifier() {
			identifiers = append(identifiers, strings.TrimSpace(node.GetText()))
		}
		if node.IsFunctionDeclaration() {
			functions = append(functions, strings.TrimSpace(node.GetText()))
		}
	})

	assert.Greater(t, nodeCount, 0, "应该遍历到节点")
	assert.Greater(t, len(identifiers), 0, "应该找到标识符")
	assert.Greater(t, len(functions), 0, "应该找到函数")

	t.Logf("总共遍历了 %d 个节点", nodeCount)
	t.Logf("找到的标识符: %v", identifiers)
	t.Logf("找到的函数: %v", functions)
}

// TestNode_TransparentAPIs 测试 Node 透明数据访问 API
// 测试 API: GetParserData(), TryGetParserData(), AsVariableDeclaration(),
//
//	AsCallExpression(), AsPropertyAccessExpression(), AsFunctionDeclaration()
func TestNode_TransparentAPIs(t *testing.T) {
	project := NewProjectFromSources(map[string]string{
		"/test.ts": `
			const x = 1;
			function test() {
				return x;
			}
			test();
			obj.key;
		`,
	})
	defer project.Close()

	sf := project.GetSourceFile("/test.ts")
	require.NotNil(t, sf)

	sf.ForEachDescendant(func(node Node) {
		// 测试 GetParserData
		data, ok := node.GetParserData()
		if ok {
			assert.NotNil(t, data, "透传数据不应为nil")
		}

		// 测试 AsXXX 方法
		if varDecl, ok := node.AsVariableDeclaration(); ok {
			assert.NotNil(t, varDecl.GetNode())
			nameNode := varDecl.GetNameNode()
			assert.NotNil(t, nameNode)
			name := varDecl.GetName()
			assert.NotEmpty(t, name)
		}

		if callExpr, ok := node.AsCallExpression(); ok {
			assert.NotNil(t, callExpr.GetNode())
			expr := callExpr.GetExpression()
			assert.NotNil(t, expr)
			args := callExpr.GetArguments()
			// 参数可能为nil（当函数调用没有参数时）
			if args != nil {
				t.Logf("调用表达式 %s 有 %d 个参数", callExpr.GetExpression().GetText(), len(args))
			}
		}

		if propAccess, ok := node.AsPropertyAccessExpression(); ok {
			assert.NotNil(t, propAccess.GetNode())
			name := propAccess.GetName()
			assert.NotEmpty(t, name)
			expr := propAccess.GetExpression()
			assert.NotNil(t, expr)
		}

		if funcDecl, ok := node.AsFunctionDeclaration(); ok {
			assert.NotNil(t, funcDecl.GetNode())
			nameNode := funcDecl.GetNameNode()
			if nameNode != nil {
				name := funcDecl.GetName()
				assert.NotEmpty(t, name)
			}
		}
	})
}

// TestNode_EdgeCases 测试 Node 边界情况
// 测试 API: IsValid() 在各种边界情况下的行为
func TestNode_EdgeCases(t *testing.T) {
	project := NewProjectFromSources(map[string]string{
		"/test.ts": ``,
	})
	defer project.Close()

	sf := project.GetSourceFile("/test.ts")
	require.NotNil(t, sf)

	// 测试空文件中的节点遍历
	nodeCount := 0
	sf.ForEachDescendant(func(node Node) {
		nodeCount++
		assert.True(t, node.IsValid(), "节点应该是有效的")
		assert.NotEmpty(t, node.GetKindName(), "应该有类型名称")
	})

	// 即使是空文件，也应该有一些基本的 AST 节点
	t.Logf("空文件中找到 %d 个节点", nodeCount)
}
