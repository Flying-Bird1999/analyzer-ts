package tsmorphgo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

// TestTypedAPI_CoreFunctionality 测试特定节点类型专有API的核心功能
// 基于 ts-morph.md 中定义的需求场景
func TestTypedAPI_CoreFunctionality(t *testing.T) {
	project := tsmorphgo.NewProjectFromSources(map[string]string{
		"/test.ts": `
			// 场景 7.3: VariableDeclaration
			const x = 1;
			const y;

			// 场景 7.1: CallExpression
			test();
			obj.method(param);

			// 场景 7.2: PropertyAccessExpression
			console.log(obj.prop);

			// 场景 7.4: FunctionDeclaration
			function foo() { return 1; }
		`,
	})

	sf := project.GetSourceFile("/test.ts")
	assert.NotNil(t, sf)

	t.Run("场景7.3-VariableDeclaration", func(t *testing.T) {
		var varDecls []*tsmorphgo.VariableDeclaration
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			if varDecl, ok := node.AsVariableDeclaration(); ok {
				varDecls = append(varDecls, varDecl)
				name := varDecl.GetName()
				t.Logf("变量 %q: HasInitializer=%t", name, varDecl.HasInitializer())
				if varDecl.HasInitializer() {
					initializer := varDecl.GetInitializer()
					if initializer != nil {
						t.Logf("  初始值: %q", initializer.GetText())
					} else {
						t.Logf("  初始值: nil")
					}
				}
			}
		})

		assert.GreaterOrEqual(t, len(varDecls), 2, "应该找到变量声明")

		// 调试：检查所有找到的变量
		for i, varDecl := range varDecls {
			name := varDecl.GetName()
			hasInit := varDecl.HasInitializer()
			t.Logf("变量 %d: name=%q, HasInitializer=%t", i, name, hasInit)

			// 调试：打印所有子节点
			children := varDecl.GetChildren()
			t.Logf("  子节点数量: %d", len(children))
			for j, child := range children {
				t.Logf("    子节点 %d: kind=%s, text=%q", j, child.GetKindName(), child.GetText())
			}

			// 调试：检查初始值
			initializer := varDecl.GetInitializer()
			if initializer != nil {
				t.Logf("  初始值: %q (kind=%s)", initializer.GetText(), initializer.GetKindName())
			} else {
				t.Logf("  初始值: nil")
			}
		}

		// 测试有初始值的变量
		for _, varDecl := range varDecls {
			name := varDecl.GetName()
			if name == "x" {
				assert.True(t, varDecl.HasInitializer(), "x应该有初始值")
				assert.NotNil(t, varDecl.GetInitializer(), "应该能获取初始值")
			}
			if name == "y" {
				assert.False(t, varDecl.HasInitializer(), "y不应该有初始值")
			}
		}
	})

	t.Run("场景7.1-CallExpression", func(t *testing.T) {
		var callExprs []*tsmorphgo.CallExpression
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			if callExpr, ok := node.AsCallExpression(); ok {
				callExprs = append(callExprs, callExpr)
			}
		})

		assert.GreaterOrEqual(t, len(callExprs), 1, "应该找到函数调用")

		// 测试核心功能：获取表达式和参数
		for _, callExpr := range callExprs {
			expr := callExpr.GetExpression()
			assert.NotNil(t, expr, "应该能获取被调用表达式")

			argCount := callExpr.GetArgumentCount()
			t.Logf("函数调用: %s, 参数数量: %d", expr.GetText(), argCount)
		}
	})

	t.Run("场景7.2-PropertyAccessExpression", func(t *testing.T) {
		var propAccesses []*tsmorphgo.PropertyAccessExpression
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			if propAccess, ok := node.AsPropertyAccessExpression(); ok {
				propAccesses = append(propAccesses, propAccess)
			}
		})

		assert.GreaterOrEqual(t, len(propAccesses), 1, "应该找到属性访问")

		// 测试核心功能：获取属性名和对象表达式
		for _, propAccess := range propAccesses {
			name := propAccess.GetName()
			assert.NotEmpty(t, name, "应该能获取属性名")

			expr := propAccess.GetExpression()
			assert.NotNil(t, expr, "应该能获取对象表达式")

			t.Logf("属性访问: %s.%s", expr.GetText(), name)
		}
	})

	t.Run("场景7.4-FunctionDeclaration", func(t *testing.T) {
		var funcDecls []*tsmorphgo.FunctionDeclaration
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			if funcDecl, ok := node.AsFunctionDeclaration(); ok {
				funcDecls = append(funcDecls, funcDecl)
			}
		})

		assert.GreaterOrEqual(t, len(funcDecls), 1, "应该找到函数声明")

		// 测试核心功能：获取函数名节点
		for _, funcDecl := range funcDecls {
			nameNode := funcDecl.GetNameNode()
			assert.NotNil(t, nameNode, "应该能获取函数名节点")

			name := funcDecl.GetName()
			assert.NotEmpty(t, name, "应该能获取函数名")

			t.Logf("函数声明: %s", name)
		}
	})

	project.Close()
}

// TestTypedAPI_TypeSafety 测试类型安全性
func TestTypedAPI_TypeSafety(t *testing.T) {
	project := tsmorphgo.NewProjectFromSources(map[string]string{
		"/test.ts": `const x = 1; test();`,
	})

	sf := project.GetSourceFile("/test.ts")
	assert.NotNil(t, sf)

	var successfulConversions int
	sf.ForEachDescendant(func(node tsmorphgo.Node) {
		// 类型安全转换：只有匹配的类型才能成功转换
		if varDecl, ok := node.AsVariableDeclaration(); ok {
			successfulConversions++
			assert.NotNil(t, varDecl.GetNode())
			assert.Equal(t, tsmorphgo.KindVariableDeclaration, varDecl.GetKind())
		}

		if callExpr, ok := node.AsCallExpression(); ok {
			successfulConversions++
			assert.NotNil(t, callExpr.GetNode())
			assert.Equal(t, tsmorphgo.KindCallExpression, callExpr.GetKind())
		}
	})

	assert.Greater(t, successfulConversions, 0, "应该有成功的类型转换")

	project.Close()
}