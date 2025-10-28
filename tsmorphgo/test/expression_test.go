package tsmorphgo

import (
	"strings"
	"testing"

	. "github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
	"github.com/stretchr/testify/assert"
)

// expression_test.go
//
// 这个文件包含了 TypeScript 表达式处理功能的测试用例，专注于验证 tsmorphgo 对各种
// TypeScript 表达式节点的解析、访问和操作能力。
//
// 主要测试场景：
// 1. 函数调用表达式 - 测试 myObj.method() 等函数调用的解析
// 2. 属性访问表达式 - 验证 obj.property 和 obj.nested.prop 等属性访问
// 3. 二元表达式 - 测试算术运算、逻辑运算等二元操作
// 4. 复杂调用链 - 验证 obj.method().anotherMethod() 等链式调用
// 5. 深度嵌套属性 - 测试 config.database.connection.timeout 等深度访问
// 6. 各种操作符 - 测试 +、-、*、/、===、!== 等操作符的识别
// 7. 无效输入处理 - 验证对非表达式节点的错误处理
//
// 测试目标：
// - 验证表达式节点的正确识别和分类
// - 确保表达式访问 API 的准确性和类型安全
// - 测试复杂表达式结构的解析能力
// - 验证在异常情况下的系统稳定性
//
// 核心 API 测试：
// - GetCallExpressionExpression() - 获取调用表达式的被调用对象
// - GetPropertyAccessName() - 获取属性访问的属性名
// - GetPropertyAccessExpression() - 获取属性访问的表达式对象
// - GetBinaryExpressionLeft/Right() - 获取二元表达式的左右操作数
// - GetBinaryExpressionOperatorToken() - 获取二元表达式的操作符

// TestExpressionAPIs 测试基础表达式API功能
func TestExpressionAPIs(t *testing.T) {
	// 测试用例 1: myObj.method()
	t.Run("PropertyAccessInCall", func(t *testing.T) {
		project := createTestProject(map[string]string{"/test_expr1.ts": `myObj.method();`})
		sf := project.GetSourceFile("/test_expr1.ts")
		assert.NotNil(t, sf)

		var callExprNode *Node
		sf.ForEachDescendant(func(node Node) {
			if node.Kind == ast.KindCallExpression {
				callExprNode = &node
			}
		})

		assert.NotNil(t, callExprNode, "未能找到 CallExpression 节点")

		// 1. 测试 GetCallExpressionExpression
		exprNode, ok := GetCallExpressionExpression(*callExprNode)
		assert.True(t, ok)
		assert.NotNil(t, exprNode)
		assert.Equal(t, ast.KindPropertyAccessExpression, exprNode.Kind)
		assert.Equal(t, "myObj.method", strings.TrimSpace(exprNode.GetText()))

		// 2. 测试 GetPropertyAccessName
		name, ok := GetPropertyAccessName(*exprNode)
		assert.True(t, ok)
		assert.Equal(t, "method", name)

		// 3. 测试 GetPropertyAccessExpression
		objNode, ok := GetPropertyAccessExpression(*exprNode)
		assert.True(t, ok)
		assert.NotNil(t, objNode)
		assert.Equal(t, ast.KindIdentifier, objNode.Kind)
		assert.Equal(t, "myObj", strings.TrimSpace(objNode.GetText()))
	})

	// 测试用例 2: a + b
	t.Run("BinaryExpression", func(t *testing.T) {
		project := createTestProject(map[string]string{"/test_expr2.ts": `const x = a + b;`})
		sf := project.GetSourceFile("/test_expr2.ts")
		assert.NotNil(t, sf)

		var binaryExprNode *Node
		sf.ForEachDescendant(func(node Node) {
			if node.Kind == ast.KindBinaryExpression {
				binaryExprNode = &node
			}
		})

		assert.NotNil(t, binaryExprNode, "未能找到 BinaryExpression 节点")

		// 1. 测试 GetBinaryExpressionLeft
		left, ok := GetBinaryExpressionLeft(*binaryExprNode)
		assert.True(t, ok)
		assert.NotNil(t, left)
		assert.Equal(t, "a", strings.TrimSpace(left.GetText()))

		// 2. 测试 GetBinaryExpressionRight
		right, ok := GetBinaryExpressionRight(*binaryExprNode)
		assert.True(t, ok)
		assert.NotNil(t, right)
		assert.Equal(t, "b", strings.TrimSpace(right.GetText()))

		// 3. 测试 GetBinaryExpressionOperatorToken
		op, ok := GetBinaryExpressionOperatorToken(*binaryExprNode)
		assert.True(t, ok)
		assert.NotNil(t, op)
		assert.Equal(t, ast.KindPlusToken, op.Kind)
	})
}

// TestExpressionAPIsComprehensive 测试表达式API的全面功能
func TestExpressionAPIsComprehensive(t *testing.T) {
	// 测试用例 1: 复杂的函数调用链
	t.Run("ComplexCallChains", func(t *testing.T) {
		project := createTestProject(map[string]string{"/test_complex_call.ts": `
			obj.nested.method(arg1, arg2);
			ns.service.getdata().then(callback);
		`})
		sf := project.GetSourceFile("/test_complex_call.ts")
		assert.NotNil(t, sf)

		// 测试第一个调用: obj.nested.method(arg1, arg2)
		var call1, call2 *Node
		sf.ForEachDescendant(func(node Node) {
			if IsCallExpression(node) {
				if call1 == nil {
					call1 = &node
				} else if call2 == nil {
					call2 = &node
				}
			}
		})

		// 测试第一个调用
		assert.NotNil(t, call1)
		expr1, ok := GetCallExpressionExpression(*call1)
		assert.True(t, ok)
		assert.True(t, IsPropertyAccessExpression(*expr1))
		assert.Equal(t, "obj.nested.method", strings.TrimSpace(expr1.GetText()))

		propName1, ok := GetPropertyAccessName(*expr1)
		assert.True(t, ok)
		assert.Equal(t, "method", propName1)

		propExpr1, ok := GetPropertyAccessExpression(*expr1)
		assert.True(t, ok)
		assert.Equal(t, "obj.nested", strings.TrimSpace(propExpr1.GetText()))

		// 测试第二个调用: ns.service.getdata().then(callback)
		if call2 != nil {
			expr2, ok := GetCallExpressionExpression(*call2)
			assert.True(t, ok)
			assert.True(t, IsPropertyAccessExpression(*expr2))
			assert.Equal(t, "ns.service.getdata().then", strings.TrimSpace(expr2.GetText()))
		}
	})

	// 测试用例 2: 各种二元操作符
	t.Run("VariousBinaryOperators", func(t *testing.T) {
		testCases := []struct {
			name      string
			code      string
			opKind    ast.Kind
			leftText  string
			rightText string
		}{
			{"Addition", "const x = a + b;", ast.KindPlusToken, "a", "b"},
			{"Subtraction", "const x = a - b;", ast.KindMinusToken, "a", "b"},
			{"Multiplication", "const x = a * b;", ast.KindAsteriskToken, "a", "b"},
			{"Division", "const x = a / b;", ast.KindSlashToken, "a", "b"},
			{"Equality", "const x = a === b;", ast.KindEqualsEqualsEqualsToken, "a", "b"},
			{"Inequality", "const x = a !== b;", ast.KindExclamationEqualsEqualsToken, "a", "b"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				project := createTestProject(map[string]string{"/test_binary_" + tc.name + ".ts": tc.code})
				sf := project.GetSourceFile("/test_binary_" + tc.name + ".ts")
				assert.NotNil(t, sf)

				var binaryExprNode *Node
				sf.ForEachDescendant(func(node Node) {
					if IsBinaryExpression(node) {
						binaryExprNode = &node
					}
				})

				assert.NotNil(t, binaryExprNode, "未能找到 BinaryExpression 节点")

				left, ok := GetBinaryExpressionLeft(*binaryExprNode)
				assert.True(t, ok)
				assert.Equal(t, tc.leftText, strings.TrimSpace(left.GetText()))

				right, ok := GetBinaryExpressionRight(*binaryExprNode)
				assert.True(t, ok)
				assert.Equal(t, tc.rightText, strings.TrimSpace(right.GetText()))

				op, ok := GetBinaryExpressionOperatorToken(*binaryExprNode)
				assert.True(t, ok)
				assert.Equal(t, tc.opKind, op.Kind)
			})
		}
	})

	// 测试用例 3: 深度嵌套的属性访问
	t.Run("DeepPropertyAccess", func(t *testing.T) {
		project := createTestProject(map[string]string{"/test_deep_prop.ts": `
			const result = config.database.connection.timeout;
		`})
		sf := project.GetSourceFile("/test_deep_prop.ts")
		assert.NotNil(t, sf)

		var propAccessNodes []*Node
		sf.ForEachDescendant(func(node Node) {
			if IsPropertyAccessExpression(node) {
				propAccessNodes = append(propAccessNodes, &node)
			}
		})

		// 应该找到3个属性访问: config.database, database.connection, connection.timeout
		assert.GreaterOrEqual(t, len(propAccessNodes), 3)

		// 找到timeout属性的访问节点
		var timeoutAccess *Node
		for _, node := range propAccessNodes {
			if name, ok := GetPropertyAccessName(*node); ok && name == "timeout" {
				timeoutAccess = node
				break
			}
		}

		assert.NotNil(t, timeoutAccess, "应该找到timeout属性访问节点")

		name, ok := GetPropertyAccessName(*timeoutAccess)
		assert.True(t, ok)
		assert.Equal(t, "timeout", name)

		expr, ok := GetPropertyAccessExpression(*timeoutAccess)
		assert.True(t, ok)
		assert.Equal(t, "config.database.connection", strings.TrimSpace(expr.GetText()))
	})

	// 测试用例 4: 无效输入测试
	t.Run("InvalidInputs", func(t *testing.T) {
		project := createTestProject(map[string]string{"/test_invalid.ts": `const x = 1;`})
		sf := project.GetSourceFile("/test_invalid.ts")
		assert.NotNil(t, sf)

		var identifierNode *Node
		sf.ForEachDescendant(func(node Node) {
			if IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "x" {
				identifierNode = &node
			}
		})
		assert.NotNil(t, identifierNode)

		// 测试在非调用表达式上调用 GetCallExpressionExpression
		_, ok := GetCallExpressionExpression(*identifierNode)
		assert.False(t, ok)

		// 测试在非属性访问表达式上调用 GetPropertyAccessName
		_, ok = GetPropertyAccessName(*identifierNode)
		assert.False(t, ok)

		// 测试在非二元表达式上调用 GetBinaryExpressionLeft
		_, ok = GetBinaryExpressionLeft(*identifierNode)
		assert.False(t, ok)
	})
}
