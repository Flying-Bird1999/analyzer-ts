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

func TestGetVariableName(t *testing.T) {
	project := createTestProject(map[string]string{
		"/test_var.ts": `const hello = "world";`,
	})
	sf := project.GetSourceFile("/test_var.ts")
	assert.NotNil(t, sf)

	var varDeclNode *Node
	sf.ForEachDescendant(func(node Node) {
		if node.Kind == ast.KindVariableDeclaration {
			varDeclNode = &node
		}
	})

	assert.NotNil(t, varDeclNode, "未能找到 VariableDeclaration 节点")

	name, ok := GetVariableName(*varDeclNode)
	assert.True(t, ok)
	assert.Equal(t, "hello", name)
}

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

func TestDeclarationAPIs(t *testing.T) {
	sourceCode := `
		function myFunc() {}
		import { foo, bar as baz } from './mod';
	`
	project := createTestProject(map[string]string{"/test_decl.ts": sourceCode})
	sf := project.GetSourceFile("/test_decl.ts")
	assert.NotNil(t, sf)

	var fnDeclNode *Node
	var importSpecFoo, importSpecBaz *Node

	sf.ForEachDescendant(func(node Node) {
		if node.Kind == ast.KindFunctionDeclaration {
			fnDeclNode = &node
		}
		if node.Kind == ast.KindImportSpecifier {
			name := strings.TrimSpace(node.AsImportSpecifier().Name().Text())
			if name == "foo" {
				importSpecFoo = &node
			} else if name == "baz" {
				importSpecBaz = &node
			}
		}
	})

	// 1. 测试 GetFunctionDeclarationNameNode
	assert.NotNil(t, fnDeclNode)
	fnNameNode, ok := GetFunctionDeclarationNameNode(*fnDeclNode)
	assert.True(t, ok)
	assert.NotNil(t, fnNameNode)
	assert.Equal(t, "myFunc", strings.TrimSpace(fnNameNode.GetText()))

	// 2. 测试 GetImportSpecifierAliasNode
	// 对于 `foo`，没有别名
	assert.NotNil(t, importSpecFoo)
	_, ok = GetImportSpecifierAliasNode(*importSpecFoo)
	assert.False(t, ok, "foo 不应该有别名")

	// 对于 `bar as baz`，别名是 `baz`
	assert.NotNil(t, importSpecBaz)
	aliasNode, ok := GetImportSpecifierAliasNode(*importSpecBaz)
	assert.True(t, ok, "baz 应该有别名")
	assert.NotNil(t, aliasNode)
	assert.Equal(t, "baz", strings.TrimSpace(aliasNode.GetText()))
}

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
		prop := child.AsPropertyAssignment()
		return prop != nil && prop.Initializer != nil && prop.Initializer.Kind == ast.KindTrueKeyword
	})
	assert.True(t, ok)
	assert.NotNil(t, enabledNode)
	assert.Equal(t, `enabled: true`, strings.TrimSpace(enabledNode.GetText()))
}
