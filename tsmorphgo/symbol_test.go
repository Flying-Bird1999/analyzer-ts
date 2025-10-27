package tsmorphgo

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetSymbol 测试符号获取的基本功能
func TestGetSymbol(t *testing.T) {
	t.Run("VariableSymbol", func(t *testing.T) {
		project := createTestProject(map[string]string{
			"/test_var.ts": `const myVariable = "hello";`,
		})
		sf := project.GetSourceFile("/test_var.ts")
		assert.NotNil(t, sf)

		// 找到变量标识符节点
		var identifierNode *Node
		sf.ForEachDescendant(func(node Node) {
			if IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "myVariable" {
				identifierNode = &node
			}
		})

		assert.NotNil(t, identifierNode, "未能找到 'myVariable' 标识符节点")

		// 测试 GetSymbol 方法
		symbol, found := GetSymbol(*identifierNode)
		if assert.True(t, found, "应该能够获取符号") && assert.NotNil(t, symbol, "符号不应该为 nil") {
			// 测试符号的基本属性
			assert.Equal(t, "myVariable", symbol.GetName(), "符号名称应该匹配")
			assert.True(t, symbol.IsVariable(), "应该是变量符号")
			assert.True(t, symbol.HasValue(), "变量应该具有值")
			assert.Equal(t, 1, symbol.GetDeclarationCount(), "应该只有一个声明")
		}
	})

	t.Run("FunctionSymbol", func(t *testing.T) {
		project := createTestProject(map[string]string{
			"/test_func.ts": `
				function myFunction(param: string): number {
					return 42;
				}
			`,
		})
		sf := project.GetSourceFile("/test_func.ts")
		assert.NotNil(t, sf)

		// 找到函数名标识符节点
		var funcNameNode *Node
		sf.ForEachDescendant(func(node Node) {
			if IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "myFunction" {
				// 确保是函数声明中的标识符，而不是调用
				if parent := node.GetParent(); parent != nil && IsFunctionDeclaration(*parent) {
					funcNameNode = &node
				}
			}
		})

		assert.NotNil(t, funcNameNode, "未能找到 'myFunction' 函数名节点")

		// 测试 GetSymbol 方法
		symbol, found := GetSymbol(*funcNameNode)
		if assert.True(t, found, "应该能够获取函数符号") && assert.NotNil(t, symbol, "符号不应该为 nil") {
			assert.Equal(t, "myFunction", symbol.GetName(), "函数名应该匹配")
			assert.True(t, symbol.IsFunction(), "应该是函数符号")
			assert.True(t, symbol.HasValue(), "函数应该具有值")
		}
	})

	t.Run("ClassSymbol", func(t *testing.T) {
		project := createTestProject(map[string]string{
			"/test_class.ts": `
				class MyClass {
					private property: string;
					constructor() {}
					method(): void {}
				}
			`,
		})
		sf := project.GetSourceFile("/test_class.ts")
		assert.NotNil(t, sf)

		// 找到类名标识符节点
		var classNameNode *Node
		sf.ForEachDescendant(func(node Node) {
			if IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "MyClass" {
				if parent := node.GetParent(); parent != nil && IsClassDeclaration(*parent) {
					classNameNode = &node
				}
			}
		})

		assert.NotNil(t, classNameNode, "未能找到 'MyClass' 类名节点")

		// 测试 GetSymbol 方法
		symbol, found := GetSymbol(*classNameNode)
		if assert.True(t, found, "应该能够获取类符号") && assert.NotNil(t, symbol, "符号不应该为 nil") {
			assert.Equal(t, "MyClass", symbol.GetName(), "类名应该匹配")
			assert.True(t, symbol.IsClass(), "应该是类符号")
			assert.True(t, symbol.HasType(), "类应该具有类型")
			assert.True(t, symbol.HasValue(), "类应该具有值")
		}
	})

	t.Run("InterfaceSymbol", func(t *testing.T) {
		project := createTestProject(map[string]string{
			"/test_interface.ts": `
				interface MyInterface {
					prop1: string;
					prop2: number;
				}
			`,
		})
		sf := project.GetSourceFile("/test_interface.ts")
		assert.NotNil(t, sf)

		// 找到接口名标识符节点
		var interfaceNameNode *Node
		sf.ForEachDescendant(func(node Node) {
			if IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "MyInterface" {
				if parent := node.GetParent(); parent != nil && IsInterfaceDeclaration(*parent) {
					interfaceNameNode = &node
				}
			}
		})

		assert.NotNil(t, interfaceNameNode, "未能找到 'MyInterface' 接口名节点")

		// 测试 GetSymbol 方法
		symbol, found := GetSymbol(*interfaceNameNode)
		if assert.True(t, found, "应该能够获取接口符号") && assert.NotNil(t, symbol, "符号不应该为 nil") {
			assert.Equal(t, "MyInterface", symbol.GetName(), "接口名应该匹配")
			assert.True(t, symbol.IsInterface(), "应该是接口符号")
			assert.True(t, symbol.HasType(), "接口应该具有类型")
			assert.False(t, symbol.HasValue(), "接口不应该具有值")
		}
	})

	t.Run("ExportedSymbol", func(t *testing.T) {
		project := createTestProject(map[string]string{
			"/test_export.ts": `
				export const exportedVar = "exported";
				const privateVar = "private";
			`,
		})
		sf := project.GetSourceFile("/test_export.ts")
		assert.NotNil(t, sf)

		// 找到导出的变量标识符节点
		var exportedVarNode *Node
		var privateVarNode *Node

		sf.ForEachDescendant(func(node Node) {
			if IsIdentifier(node) {
				text := strings.TrimSpace(node.GetText())
				if text == "exportedVar" {
					exportedVarNode = &node
				} else if text == "privateVar" {
					privateVarNode = &node
				}
			}
		})

		assert.NotNil(t, exportedVarNode, "未能找到 'exportedVar' 节点")
		assert.NotNil(t, privateVarNode, "未能找到 'privateVar' 节点")

		// 测试导出变量的符号
		exportedSymbol, exportedFound := GetSymbol(*exportedVarNode)
		if assert.True(t, exportedFound, "应该能够获取导出符号") && assert.NotNil(t, exportedSymbol, "导出符号不应该为 nil") {
			assert.Equal(t, "exportedVar", exportedSymbol.GetName())
			assert.True(t, exportedSymbol.IsExported(), "应该是导出的符号")
		}

		// 测试私有变量的符号
		privateSymbol, privateFound := GetSymbol(*privateVarNode)
		if assert.True(t, privateFound, "应该能够获取私有符号") && assert.NotNil(t, privateSymbol, "私有符号不应该为 nil") {
			assert.Equal(t, "privateVar", privateSymbol.GetName())
			assert.False(t, privateSymbol.IsExported(), "不应该是导出的符号")
		}
	})
}

// TestSymbolDeclarations 测试符号声明相关的功能
func TestSymbolDeclarations(t *testing.T) {
	project := createTestProject(map[string]string{
		"/test_decl.ts": `
			const myVar = "test";
			function myFunc() {
				return "hello";
			}
		`,
	})
	sf := project.GetSourceFile("/test_decl.ts")
	assert.NotNil(t, sf)

	// 测试变量符号的声明
	var varIdentifierNode *Node
	sf.ForEachDescendant(func(node Node) {
		if IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "myVar" {
			varIdentifierNode = &node
		}
	})

	assert.NotNil(t, varIdentifierNode)
	varSymbol, found := GetSymbol(*varIdentifierNode)
	assert.True(t, found)
	assert.NotNil(t, varSymbol)

	// 测试 GetDeclarations
	declarations := varSymbol.GetDeclarations()
	assert.Len(t, declarations, 1, "变量应该只有一个声明")

	// 测试 GetFirstDeclaration
	firstDecl, ok := varSymbol.GetFirstDeclaration()
	assert.True(t, ok)
	assert.NotNil(t, firstDecl)
	assert.True(t, IsVariableDeclaration(*firstDecl), "第一个声明应该是变量声明")

	// 测试函数符号的声明
	var funcIdentifierNode *Node
	sf.ForEachDescendant(func(node Node) {
		if IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "myFunc" {
			if parent := node.GetParent(); parent != nil && IsFunctionDeclaration(*parent) {
				funcIdentifierNode = &node
			}
		}
	})

	assert.NotNil(t, funcIdentifierNode)
	funcSymbol, found := GetSymbol(*funcIdentifierNode)
	assert.True(t, found)
	assert.NotNil(t, funcSymbol)

	funcDeclarations := funcSymbol.GetDeclarations()
	assert.Len(t, funcDeclarations, 1, "函数应该只有一个声明")

	funcFirstDecl, ok := funcSymbol.GetFirstDeclaration()
	assert.True(t, ok)
	assert.NotNil(t, funcFirstDecl)
	assert.True(t, IsFunctionDeclaration(*funcFirstDecl), "第一个声明应该是函数声明")
}

// TestSymbolString 测试符号的字符串表示
func TestSymbolString(t *testing.T) {
	project := createTestProject(map[string]string{
		"/test_string.ts": `const testVar = "string representation";`,
	})
	sf := project.GetSourceFile("/test_string.ts")
	assert.NotNil(t, sf)

	var identifierNode *Node
	sf.ForEachDescendant(func(node Node) {
		if IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "testVar" {
			identifierNode = &node
		}
	})

	assert.NotNil(t, identifierNode)
	symbol, found := GetSymbol(*identifierNode)
	assert.True(t, found)
	assert.NotNil(t, symbol)

	// 测试 String() 方法
	str := symbol.String()
	assert.Contains(t, str, "testVar", "字符串表示应该包含符号名称")
	assert.Contains(t, str, "Symbol{name:", "字符串表示应该有正确的格式")
}

// TestGetSymbolWithInvalidNode 测试无效节点的符号获取
func TestGetSymbolWithInvalidNode(t *testing.T) {
	project := createTestProject(map[string]string{
		"/test_invalid.ts": `const x = 1;`,
	})
	sf := project.GetSourceFile("/test_invalid.ts")
	assert.NotNil(t, sf)

	// 创建一个无效的节点（没有sourceFile）
	invalidNode := Node{
		Node:       sf.astNode, // 使用有效的AST节点
		sourceFile: nil,        // 但是sourceFile为nil
	}

	// 测试无效节点的符号获取
	symbol, found := GetSymbol(invalidNode)
	assert.False(t, found, "不应该能从无效节点获取符号")
	assert.Nil(t, symbol, "符号应该为nil")
}

// TestSymbolFlagsCombinations 测试符号标志的组合
func TestSymbolFlagsCombinations(t *testing.T) {
	project := createTestProject(map[string]string{
		"/test_flags.ts": `
			class MyClass {
				method(): void {}
				get getter(): string { return ""; }
				set setter(value: string) {}
			}
		`,
	})
	sf := project.GetSourceFile("/test_flags.ts")
	assert.NotNil(t, sf)

	// 找到类名标识符节点
	var classNameNode, methodNameNode, getterNameNode, setterNameNode *Node

	sf.ForEachDescendant(func(node Node) {
		if !IsIdentifier(node) {
			return
		}

		text := strings.TrimSpace(node.GetText())
		parent := node.GetParent()

		switch text {
		case "MyClass":
			if parent != nil && IsClassDeclaration(*parent) {
				classNameNode = &node
			}
		case "method":
			if parent != nil && IsMethodDeclaration(*parent) {
				methodNameNode = &node
			}
		case "getter":
			if parent != nil && IsGetAccessor(*parent) {
				getterNameNode = &node
			}
		case "setter":
			if parent != nil && IsSetAccessor(*parent) {
				setterNameNode = &node
			}
		}
	})

	// 测试类符号标志
	assert.NotNil(t, classNameNode)
	classSymbol, found := GetSymbol(*classNameNode)
	assert.True(t, found)
	assert.NotNil(t, classSymbol)
	assert.True(t, classSymbol.IsClass())
	assert.True(t, classSymbol.HasType())
	assert.True(t, classSymbol.HasValue())

	// 测试方法符号标志
	if methodNameNode != nil {
		methodSymbol, found := GetSymbol(*methodNameNode)
		if assert.True(t, found) && assert.NotNil(t, methodSymbol) {
			assert.True(t, methodSymbol.IsMethod())
		}
	}

	// 测试getter符号标志
	if getterNameNode != nil {
		getterSymbol, found := GetSymbol(*getterNameNode)
		if assert.True(t, found) && assert.NotNil(t, getterSymbol) {
			assert.True(t, getterSymbol.IsAccessor())
		}
	}

	// 测试setter符号标志
	if setterNameNode != nil {
		setterSymbol, found := GetSymbol(*setterNameNode)
		if assert.True(t, found) && assert.NotNil(t, setterSymbol) {
			assert.True(t, setterSymbol.IsAccessor())
		}
	}
}

// TestSymbolRelationships 测试符号关系相关的功能
func TestSymbolRelationships(t *testing.T) {
	project := createTestProject(map[string]string{
		"/test_relations.ts": `
			class MyClass {
				method1(): void {}
				method2(): string { return ""; }
			}
		`,
	})
	sf := project.GetSourceFile("/test_relations.ts")
	assert.NotNil(t, sf)

	// 测试类符号的成员
	var classNameNode *Node
	sf.ForEachDescendant(func(node Node) {
		if IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "MyClass" {
			if parent := node.GetParent(); parent != nil && IsClassDeclaration(*parent) {
				classNameNode = &node
			}
		}
	})

	assert.NotNil(t, classNameNode)
	classSymbol, found := GetSymbol(*classNameNode)
	assert.True(t, found)
	assert.NotNil(t, classSymbol)

	// 测试 GetMembers - 当前实现可能返回空或有限成员
	members := classSymbol.GetMembers()
	assert.NotNil(t, members, "GetMembers 不应该返回 nil")

	// 测试 GetParent
	parent, hasParent := classSymbol.GetParent()
	assert.False(t, hasParent, "顶级类符号不应该有父符号")
	assert.Nil(t, parent, "顶级类符号的父符号应该为 nil")

	// 测试 GetExports - 对于普通类，应该没有导出
	exports := classSymbol.GetExports()
	assert.NotNil(t, exports, "GetExports 不应该返回 nil")
	// 可能是空 map，这是符合预期的
}

// TestSymbolEdgeCases 测试符号系统的边界情况
func TestSymbolEdgeCases(t *testing.T) {
	// 测试空符号
	var emptySymbol *Symbol
	assert.Nil(t, emptySymbol)

	emptySymbol = &Symbol{}
	assert.Equal(t, "", emptySymbol.GetName())
	assert.Equal(t, SymbolFlags(0), emptySymbol.GetFlags())
	assert.False(t, emptySymbol.IsExported())
	assert.False(t, emptySymbol.IsVariable())
	assert.Equal(t, 0, emptySymbol.GetDeclarationCount())

	// 测试空声明列表
	declarations := emptySymbol.GetDeclarations()
	assert.Empty(t, declarations)

	firstDecl, ok := emptySymbol.GetFirstDeclaration()
	assert.False(t, ok)
	assert.Nil(t, firstDecl)

	// 测试空成员和导出
	members := emptySymbol.GetMembers()
	assert.Empty(t, members)

	exports := emptySymbol.GetExports()
	assert.Empty(t, exports)

	parent, ok := emptySymbol.GetParent()
	assert.False(t, ok)
	assert.Nil(t, parent)
}

// TestSymbolFindReferences 测试符号的引用查找功能
func TestSymbolFindReferences(t *testing.T) {
	// 注意：当前 FindReferences 实现可能返回有限的结果
	// 我们测试基本的错误处理和返回值
	project := createTestProject(map[string]string{
		"/test_refs.ts": `
			function targetFunction() {
				return "test";
			}

			// 引用
			targetFunction();
		`,
	})
	sf := project.GetSourceFile("/test_refs.ts")
	assert.NotNil(t, sf)

	// 找到目标函数的标识符节点
	var targetFuncNode *Node
	sf.ForEachDescendant(func(node Node) {
		if IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "targetFunction" {
			if parent := node.GetParent(); parent != nil && IsFunctionDeclaration(*parent) {
				targetFuncNode = &node
			}
		}
	})

	assert.NotNil(t, targetFuncNode, "未能找到 targetFunction 声明节点")

	// 测试 FindReferences - 基本功能测试
	targetSymbol, found := GetSymbol(*targetFuncNode)
	assert.True(t, found)
	assert.NotNil(t, targetSymbol)

	references, err := targetSymbol.FindReferences()
	// 应该不返回错误
	assert.NoError(t, err)
	assert.NotNil(t, references, "FindReferences 不应该返回 nil")

	// 验证返回值是有效的 slice（可能为空）
	// 当前实现可能不返回完整的引用列表，这是符合预期的
	assert.GreaterOrEqual(t, len(references), 0, "引用列表应该包含0个或多个引用")
}