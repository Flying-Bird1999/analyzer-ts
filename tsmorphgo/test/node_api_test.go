package tsmorphgo_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

// TestNewNodeAPIs 测试新实现的 Node 相关 API
func TestNewNodeAPIs(t *testing.T) {
	project := tsmorphgo.NewProjectFromSources(map[string]string{
		"/test.ts": `
			const x = 1;
			function test() {
				const y = x + 1;
				return y;
			}
			import { Something } from './module';
			const obj = { key: 'value' };
			test();
		`,
	})

	sf := project.GetSourceFile("/test.ts")
	assert.NotNil(t, sf)

	// 测试位置信息 API
	sf.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsIdentifier() && node.GetText() == "x" {
			// 测试新添加的位置信息 API
			assert.Greater(t, node.GetEndLineNumber(), 0, "应该有结束行号")
			assert.Greater(t, node.GetEndColumnNumber(), 0, "应该有结束列号")
			assert.GreaterOrEqual(t, node.GetWidth(), 0, "应该有宽度")
			assert.NotEmpty(t, node.GetKindName(), "应该有类型名称")
			t.Logf("节点 '%s': 行 %d-%d, 列 %d-%d, 宽度: %d, 类型: %s",
				node.GetText(),
				node.GetStartLineNumber(), node.GetEndLineNumber(),
				node.GetStartColumnNumber(), node.GetEndColumnNumber(),
				node.GetWidth(),
				node.GetKindName(),
			)
		}
	})

	// 测试节点导航 API
	sf.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsFunctionDeclaration() {
			// 测试 GetChildren
			children := node.GetChildren()
			assert.NotEmpty(t, children, "函数声明应该有子节点")

			// 测试 GetFirstChild
			firstChild := node.GetFirstChild(func(child tsmorphgo.Node) bool {
				return child.IsIdentifier()
			})
			assert.NotNil(t, firstChild, "应该找到第一个标识符子节点")
			t.Logf("函数 '%s' 的第一个标识符: %s", node.GetText(), firstChild.GetText())

			// 测试 ForEachChild
			childCount := 0
			node.ForEachChild(func(child tsmorphgo.Node) bool {
				childCount++
				return false // 继续遍历
			})
			assert.Greater(t, childCount, 0, "应该有子节点")
		}
	})

	// 测试类型判断 API
	sf.ForEachDescendant(func(node tsmorphgo.Node) {
		// 测试新增的类型判断
		if node.GetText() == "key" {
			assert.True(t, node.IsPropertyAssignment(), "key 应该是属性赋值")
		}
		if node.GetText() == "Something" {
			assert.True(t, node.IsImportSpecifier(), "Something 应该是导入指定符")
		}
	})

	// 测试特定节点类型的专有 API
	sf.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsVariableDeclaration() {
			// 测试 VariableDeclaration 的类型安全转换
			if varDecl, ok := node.AsVariableDeclaration(); ok {
				nameNode := varDecl.GetNameNode()
				assert.NotNil(t, nameNode, "应该有名称节点")
				name := varDecl.GetName()
				assert.NotEmpty(t, name, "应该有名称")

				// 检查是否有初始值
				if initializer := varDecl.GetInitializer(); initializer != nil {
					t.Logf("变量 '%s' 的初始值: %s", name, initializer.GetText())
				}
			}
		}

		if node.IsCallExpression() {
			// 测试 CallExpression 的类型安全转换
			if callExpr, ok := node.AsCallExpression(); ok {
				expr := callExpr.GetExpression()
				assert.NotNil(t, expr, "调用表达式应该有被调用部分")
				args := callExpr.GetArguments()
				t.Logf("函数调用: %s, 参数数量: %d", expr.GetText(), len(args))
			}
		}
	})

	project.Close()
}

// TestNodeNavigationAPI 测试节点导航功能
func TestNodeNavigationAPI(t *testing.T) {
	project := tsmorphgo.NewProjectFromSources(map[string]string{
		"/test.ts": `
			function outer() {
				const x = 1;
				return x;
			}
		`,
	})

	sf := project.GetSourceFile("/test.ts")
	assert.NotNil(t, sf)

	// 调试：打印所有节点信息
	t.Logf("开始遍历所有节点...")
	sf.ForEachDescendant(func(node tsmorphgo.Node) {
		text := strings.TrimSpace(node.GetText())
		if text == "x" {
			t.Logf("找到 x 节点: 类型=%s, 是否变量声明=%t", node.GetKindName(), node.IsVariableDeclaration())
		}
		// 打印前10个节点的信息用于调试
		if text != "" && len(text) < 20 {
			t.Logf("节点: [%s] (%s)", text, node.GetKindName())
		}
	})

	// 查找标识符节点
	var identifierNode *tsmorphgo.Node
	sf.ForEachDescendant(func(node tsmorphgo.Node) {
		if strings.TrimSpace(node.GetText()) == "x" && node.IsIdentifier() {
			identifierNode = &node
			t.Logf("设置 x 标识符节点: 类型=%s", node.GetKindName())
		}
	})
	assert.NotNil(t, identifierNode, "应该找到 x 标识符节点")

	var varNode *tsmorphgo.Node = identifierNode

	if varNode != nil {
		// 测试祖先导航
		ancestors := varNode.GetAncestors()
		t.Logf("祖先节点数量: %d", len(ancestors))
		assert.Greater(t, len(ancestors), 0, "应该有祖先节点")

		// 测试按类型查找祖先
		funcDecl, found := varNode.GetFirstAncestorByKind(tsmorphgo.KindFunctionDeclaration)
		t.Logf("找到函数声明: %t", found)
		if found && funcDecl != nil {
			t.Logf("函数声明文本: %s", funcDecl.GetText())
		}
		assert.True(t, found, "应该找到函数声明祖先")
		assert.NotNil(t, funcDecl, "应该找到 outer 函数声明")

		// 测试父节点导航
		parent := varNode.GetParent()
		if parent != nil {
			t.Logf("父节点类型: %s", parent.GetKindName())
		}
		assert.NotNil(t, parent, "应该有父节点")

		// 测试子节点导航
		children := varNode.GetChildren()
		t.Logf("子节点数量: %d", len(children))
		for i, child := range children {
			t.Logf("子节点 %d: %s (%s)", i, child.GetText(), child.GetKindName())
		}
	}

	project.Close()
}

// TestPositionCalculation 测试位置计算准确性
func TestPositionCalculation(t *testing.T) {
	source := `const x = 1;
const y = 2;`

	project := tsmorphgo.NewProjectFromSources(map[string]string{
		"/test.ts": source,
	})

	sf := project.GetSourceFile("/test.ts")
	assert.NotNil(t, sf)

	// 查找 x 变量
	var xNode *tsmorphgo.Node
	sf.ForEachDescendant(func(node tsmorphgo.Node) {
		text := strings.TrimSpace(node.GetText())
		if text == "x" {
			xNode = &node
		}
	})
	assert.NotNil(t, xNode, "应该找到 x 变量")

	// 验证位置信息
	assert.Equal(t, 1, xNode.GetStartLineNumber(), "x 应该在第1行")
	assert.Equal(t, 1, xNode.GetEndLineNumber(), "x 应该在第1行结束")
	assert.Equal(t, 0, xNode.GetStartLinePos(), "行起始位置应该正确")
	assert.Equal(t, "x", strings.TrimSpace(xNode.GetText()), "文本应该是 x")
	assert.Equal(t, 2, xNode.GetWidth(), "宽度应该是2（包含空格）")

	// 查找 y 变量
	var yNode *tsmorphgo.Node
	sf.ForEachDescendant(func(node tsmorphgo.Node) {
		text := strings.TrimSpace(node.GetText())
		if text == "y" {
			yNode = &node
		}
	})
	assert.NotNil(t, yNode, "应该找到 y 变量")

	// 验证位置信息
	assert.Equal(t, 2, yNode.GetStartLineNumber(), "y 应该在第2行")
	assert.Equal(t, "y", strings.TrimSpace(yNode.GetText()), "文本应该是 y")

	project.Close()
}

// TestTypeChecking 测试类型检查功能
func TestTypeChecking(t *testing.T) {
	project := tsmorphgo.NewProjectFromSources(map[string]string{
		"/test.ts": `
			const x = 1;
			function test() { return x; }
			const obj = { key: 'value' };
			import { lib } from './library';
		`,
	})

	sf := project.GetSourceFile("/test.ts")
	assert.NotNil(t, sf)

	var foundTypes map[tsmorphgo.SyntaxKind]bool = make(map[tsmorphgo.SyntaxKind]bool)

	sf.ForEachDescendant(func(node tsmorphgo.Node) {
		// 测试基础类型检查
		if node.IsVariableDeclaration() {
			foundTypes[tsmorphgo.KindVariableDeclaration] = true
		}
		if node.IsFunctionDeclaration() {
			foundTypes[tsmorphgo.KindFunctionDeclaration] = true
		}
		if node.IsObjectLiteralExpression() {
			foundTypes[tsmorphgo.KindObjectLiteralExpression] = true
		}
		if node.IsImportDeclaration() {
			foundTypes[tsmorphgo.KindImportDeclaration] = true
		}
		if node.IsIdentifier() {
			foundTypes[tsmorphgo.KindIdentifier] = true
		}
	})

	// 验证找到了预期类型
	assert.True(t, foundTypes[tsmorphgo.KindVariableDeclaration], "应该找到变量声明")
	assert.True(t, foundTypes[tsmorphgo.KindFunctionDeclaration], "应该找到函数声明")
	assert.True(t, foundTypes[tsmorphgo.KindObjectLiteralExpression], "应该找到对象字面量")
	assert.True(t, foundTypes[tsmorphgo.KindImportDeclaration], "应该找到导入声明")
	assert.True(t, foundTypes[tsmorphgo.KindIdentifier], "应该找到标识符")

	project.Close()
}