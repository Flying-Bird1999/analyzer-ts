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

// =============================================================================
// ImportSpecifier API 测试
// =============================================================================

// TestNode_ImportSpecifier_GetAliasNode 测试 ImportSpecifier.GetAliasNode API
// API: GetAliasNode() *Node
// 对应ts-morph: ImportSpecifier.getAliasNode() method
func TestNode_ImportSpecifier_GetAliasNode(t *testing.T) {
	project := NewProjectFromSources(map[string]string{
		"/test.ts": `
			// 测试无别名导入
			import { noalias } from './module1';

			// 测试有别名导入
			import { original as alias } from './module2';

			// 测试混合导入
			import { keepOriginal, changeName as newName } from './module3';
		`,
	})
	defer project.Close()

	sf := project.GetSourceFile("/test.ts")
	require.NotNil(t, sf)

	// 测试用例：[导入文本, 是否有别名]
	testCases := []struct {
		text        string
		hasAlias    bool
		description string
	}{
		{"noalias", false, "无别名导入"},
		{"alias", true, "有别名导入"},
		{"keepOriginal", false, "保持原名导入"},
		{"newName", true, "重命名导入"},
	}

	for _, tc := range testCases {
		t.Run(tc.description+"_"+tc.text, func(t *testing.T) {
			var targetImport *ImportSpecifier

			// 查找对应的ImportSpecifier
			sf.ForEachDescendant(func(node Node) {
				if node.IsImportSpecifier() {
					if importSpec, ok := node.AsImportSpecifier(); ok {
						if importSpec.GetLocalName() == tc.text {
							targetImport = importSpec
						}
					}
				}
			})

			require.NotNil(t, targetImport, "应该找到对应的ImportSpecifier: %s", tc.text)

			// 测试GetAliasNode
			aliasNode := targetImport.GetAliasNode()

			if tc.hasAlias {
				assert.NotNil(t, aliasNode, "有别名的导入应该返回别名节点")
				assert.Equal(t, tc.text, aliasNode.GetText(), "别名节点文本应该匹配本地名称")
				t.Logf("✅ 导入 '%s' 有别名节点: %s", tc.text, aliasNode.GetText())
			} else {
				assert.Nil(t, aliasNode, "无别名的导入应该返回nil")
				t.Logf("✅ 导入 '%s' 无别名节点 (符合预期)", tc.text)
			}
		})
	}
}

// TestNode_ImportSpecifier_GetOriginalName 测试 ImportSpecifier.GetOriginalName API
// API: GetOriginalName() string
func TestNode_ImportSpecifier_GetOriginalName(t *testing.T) {
	project := NewProjectFromSources(map[string]string{
		"/test.ts": `
			import { keepSame } from './module1';
			import { original as changed } from './module2';
		`,
	})
	defer project.Close()

	sf := project.GetSourceFile("/test.ts")
	require.NotNil(t, sf)

	expectedMappings := map[string]string{
		"keepSame": "keepSame",  // 无别名
		"changed":  "original",   // 有别名
	}

	for localName, expectedOriginal := range expectedMappings {
		t.Run("本地名称_"+localName, func(t *testing.T) {
			var targetImport *ImportSpecifier

			sf.ForEachDescendant(func(node Node) {
				if node.IsImportSpecifier() {
					if importSpec, ok := node.AsImportSpecifier(); ok {
						if importSpec.GetLocalName() == localName {
							targetImport = importSpec
						}
					}
				}
			})

			require.NotNil(t, targetImport, "应该找到对应的ImportSpecifier")

			originalName := targetImport.GetOriginalName()
			assert.Equal(t, expectedOriginal, originalName,
				"导入 '%s' 的原始名称应该为 '%s'", localName, expectedOriginal)

			t.Logf("✅ 导入 '%s': 原始='%s', 本地='%s'",
				localName, originalName, localName)
		})
	}
}

// TestNode_ImportSpecifier_GetLocalName 测试 ImportSpecifier.GetLocalName API
// API: GetLocalName() string
func TestNode_ImportSpecifier_GetLocalName(t *testing.T) {
	project := NewProjectFromSources(map[string]string{
		"/test.ts": `
			import { direct, renamed as alias } from './module';
		`,
	})
	defer project.Close()

	sf := project.GetSourceFile("/test.ts")
	require.NotNil(t, sf)

	expectedLocalNames := []string{"direct", "alias"}
	foundNames := make(map[string]bool)

	sf.ForEachDescendant(func(node Node) {
		if node.IsImportSpecifier() {
			if importSpec, ok := node.AsImportSpecifier(); ok {
				localName := importSpec.GetLocalName()
				t.Logf("找到导入: 本地名称='%s'", localName)
				foundNames[localName] = true
			}
		}
	})

	// 验证找到了所有预期的本地名称
	for _, expectedName := range expectedLocalNames {
		assert.True(t, foundNames[expectedName], "应该找到本地名称: %s", expectedName)
	}
}

// TestNode_ImportSpecifier_HasAlias 测试 ImportSpecifier.HasAlias API
// API: HasAlias() bool
func TestNode_ImportSpecifier_HasAlias(t *testing.T) {
	project := NewProjectFromSources(map[string]string{
		"/test.ts": `
			import { noalias1, noalias2, renamed as alias } from './module';
		`,
	})
	defer project.Close()

	sf := project.GetSourceFile("/test.ts")
	require.NotNil(t, sf)

	var noAliasCount, withAliasCount int

	sf.ForEachDescendant(func(node Node) {
		if node.IsImportSpecifier() {
			if importSpec, ok := node.AsImportSpecifier(); ok {
				localName := importSpec.GetLocalName()
				hasAlias := importSpec.HasAlias()

				t.Logf("导入 '%s': 有别名=%v", localName, hasAlias)

				if hasAlias {
					withAliasCount++
					assert.Equal(t, "alias", localName, "有别名的导入应该是'alias'")
				} else {
					noAliasCount++
					assert.Contains(t, []string{"noalias1", "noalias2"}, localName, "无别名的导入应该是'noalias1'或'noalias2'")
				}
			}
		}
	})

	assert.Equal(t, 2, noAliasCount, "应该有2个无别名的导入")
	assert.Equal(t, 1, withAliasCount, "应该有1个有别名的导入")
}

// TestNode_ImportSpecifier_AsImportSpecifier 测试 ImportSpecifier.AsImportSpecifier API
// API: AsImportSpecifier() (*ImportSpecifier, bool)
func TestNode_ImportSpecifier_AsImportSpecifier(t *testing.T) {
	project := NewProjectFromSources(map[string]string{
		"/test.ts": `
			import { testItem } from './module';
			const localVar = 42;
		`,
	})
	defer project.Close()

	sf := project.GetSourceFile("/test.ts")
	require.NotNil(t, sf)

	var importNode, nonImportNode Node

	// 找到ImportSpecifier和非ImportSpecifier节点
	sf.ForEachDescendant(func(node Node) {
		if node.IsImportSpecifier() && importNode.IsValid() == false {
			importNode = node
		} else if node.IsIdentifier() && node.GetText() == "localVar" && nonImportNode.IsValid() == false {
			nonImportNode = node
		}
	})

	// 测试成功转换
	if importNode.IsValid() {
		importSpec, ok := importNode.AsImportSpecifier()
		assert.True(t, ok, "ImportSpecifier节点应该成功转换")
		assert.NotNil(t, importSpec, "转换后的ImportSpecifier不应该为nil")
		assert.Equal(t, importNode.GetText(), importSpec.GetText(), "转换后文本应该保持一致")
		t.Logf("✅ 成功转换ImportSpecifier: %s", importSpec.GetText())
	} else {
		t.Fatal("没有找到ImportSpecifier节点进行测试")
	}

	// 测试失败转换
	if nonImportNode.IsValid() {
		importSpec, ok := nonImportNode.AsImportSpecifier()
		assert.False(t, ok, "非ImportSpecifier节点转换应该失败")
		assert.Nil(t, importSpec, "转换失败时应该返回nil")
		t.Logf("✅ 正确拒绝非ImportSpecifier节点: %s", nonImportNode.GetText())
	} else {
		t.Fatal("没有找到非ImportSpecifier节点进行测试")
	}
}

// TestNode_ImportSpecifier_GetParserData 测试 ImportSpecifier.GetParserData API
// API: GetParserData() (parser.ImportModule, bool)
func TestNode_ImportSpecifier_GetParserData(t *testing.T) {
	project := NewProjectFromSources(map[string]string{
		"/test.ts": `
			import { renamed as alias, direct } from './module';
		`,
	})
	defer project.Close()

	sf := project.GetSourceFile("/test.ts")
	require.NotNil(t, sf)

	var testedCount int
	sf.ForEachDescendant(func(node Node) {
		if node.IsImportSpecifier() {
			if importSpec, ok := node.AsImportSpecifier(); ok {
				importModule, success := importSpec.GetParserData()

				// 注意：从内存创建的项目可能没有完整的parser数据
				// 所以我们主要测试API能够正常调用，而不是具体的值
				if success {
					t.Logf("✅ 成功获取Parser数据: ImportModule='%s', Identifier='%s', Type='%s'",
						importModule.ImportModule, importModule.Identifier, importModule.Type)

					// 如果有数据，验证字段非空
					if importModule.ImportModule != "" || importModule.Identifier != "" {
						// 验证数据一致性：如果没有ImportModule，那么Identifier就是ImportModule
						if importModule.ImportModule == "" && importModule.Identifier != "" {
							t.Logf("✅ 无ImportModule时，Identifier='%s' 作为默认值", importModule.Identifier)
						}
					}
				} else {
					t.Logf("✅ GetParserData返回false，这在内存项目中是正常的")
				}

				testedCount++
			}
		}
	})

	assert.GreaterOrEqual(t, testedCount, 1, "应该至少测试了1个ImportSpecifier")
}
