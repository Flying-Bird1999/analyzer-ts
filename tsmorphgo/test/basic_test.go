package tsmorphgo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

// createTestProject 创建测试项目
func createTestProject(sources map[string]string) *tsmorphgo.Project {
	return tsmorphgo.NewProjectFromSources(sources)
}

// TestBasicNodeOperations 测试基本的节点操作
func TestBasicNodeOperations(t *testing.T) {
	project := createTestProject(map[string]string{
		"/test.ts": `
			const x = 1;
			function test() {}
			import { Something } from './module';
			"hello";
			interface Test {}
		`,
	})

	sf := project.GetSourceFile("/test.ts")
	assert.NotNil(t, sf)

	var foundVariable, foundFunction, foundImport, foundLiteral, foundInterface bool

	sf.ForEachDescendant(func(node tsmorphgo.Node) {
		// 测试新API的基本类型检查方法
		if node.IsVariableDeclaration() {
			foundVariable = true
		}
		if node.IsFunctionDeclaration() {
			foundFunction = true
		}
		if node.IsImportDeclaration() {
			foundImport = true
		}
		if node.IsInterfaceDeclaration() {
			foundInterface = true
		}

		// 测试类别检查
		if node.IsLiteral() {
			foundLiteral = true
		}
	})

	assert.True(t, foundVariable, "应该找到变量声明")
	assert.True(t, foundFunction, "应该找到函数声明")
	assert.True(t, foundImport, "应该找到导入声明")
	assert.True(t, foundInterface, "应该找到接口声明")
	assert.True(t, foundLiteral, "应该找到字面量")
}

// TestCategoryChecking 测试类别检查功能
func TestCategoryChecking(t *testing.T) {
	project := createTestProject(map[string]string{
		"/test.ts": `
			const x = 1;
			function test() {}
			let y = x + 1;
			import { Something } from './module';
			"hello";
			interface Test {}
		`,
	})

	sf := project.GetSourceFile("/test.ts")
	assert.NotNil(t, sf)

	var declarations, expressions, modules, types, literals int

	sf.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsDeclaration() {
			declarations++
		}
		if node.IsExpression() {
			expressions++
		}
		if node.IsModule() {
			modules++
		}
		if node.IsType() {
			types++
		}
		if node.IsLiteral() {
			literals++
		}
	})

	assert.Greater(t, declarations, 0, "应该找到声明")
	assert.Greater(t, expressions, 0, "应该找到表达式")
	assert.Greater(t, modules, 0, "应该找到模块")
	assert.Greater(t, types, 0, "应该找到类型")
	assert.Greater(t, literals, 0, "应该找到字面量")
}

// TestNodeKindChecking 测试节点类型检查
func TestNodeKindChecking(t *testing.T) {
	project := createTestProject(map[string]string{
		"/test.ts": `
			const x = 1;
			function test() {}
			import { Something } from './module';
		`,
	})

	sf := project.GetSourceFile("/test.ts")
	assert.NotNil(t, sf)

	var foundConstDecl, foundFuncDecl, foundImportDecl bool

	sf.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsKind(tsmorphgo.KindVariableDeclaration) {
			foundConstDecl = true
		}
		if node.IsKind(tsmorphgo.KindFunctionDeclaration) {
			foundFuncDecl = true
		}
		if node.IsKind(tsmorphgo.KindImportDeclaration) {
			foundImportDecl = true
		}
	})

	assert.True(t, foundConstDecl, "应该找到变量声明")
	assert.True(t, foundFuncDecl, "应该找到函数声明")
	assert.True(t, foundImportDecl, "应该找到导入声明")
}

// TestMultipleKindChecking 测试多类型检查
func TestMultipleKindChecking(t *testing.T) {
	project := createTestProject(map[string]string{
		"/test.ts": `
			const x = 1;
			function test() {}
			let y = "hello";
		`,
	})

	sf := project.GetSourceFile("/test.ts")
	assert.NotNil(t, sf)

	var variableOrFunctionCount int
	declarationKinds := []tsmorphgo.SyntaxKind{
		tsmorphgo.KindVariableDeclaration,
		tsmorphgo.KindFunctionDeclaration,
	}

	sf.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsAnyKind(declarationKinds...) {
			variableOrFunctionCount++
		}
	})

	assert.Equal(t, 3, variableOrFunctionCount, "应该找到3个变量或函数声明")
}

// TestTypeConversions 测试类型转换
func TestTypeConversions(t *testing.T) {
	project := createTestProject(map[string]string{
		"/test.ts": `
			import { Component } from 'react';
			const x = 1;
			function test() {}
			interface Test {}
		`,
	})

	sf := project.GetSourceFile("/test.ts")
	assert.NotNil(t, sf)

	var conversionSuccess int

	sf.ForEachDescendant(func(node tsmorphgo.Node) {
		// 测试统一声明转换
		if node.IsDeclaration() {
			if result, ok := node.AsDeclaration(); ok {
				conversionSuccess++
				// 确保结果不为nil
				assert.NotNil(t, result)
			}
		}
	})

	assert.Greater(t, conversionSuccess, 0, "应该有成功的类型转换")
}

// TestLiteralValueExtraction 测试字面量值提取
func TestLiteralValueExtraction(t *testing.T) {
	project := createTestProject(map[string]string{
		"/test.ts": `
			const str = "hello";
			const num = 42;
			const boolTrue = true;
		`,
	})

	sf := project.GetSourceFile("/test.ts")
	assert.NotNil(t, sf)

	var literalCount int

	sf.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsLiteral() {
			if value, ok := node.GetLiteralValue(); ok {
				literalCount++
				assert.NotNil(t, value, "字面量值不应该为nil")
			}
		}
	})

	assert.Greater(t, literalCount, 0, "应该找到字面量")
}

// TestNameExtraction 测试名称提取
func TestNameExtraction(t *testing.T) {
	project := createTestProject(map[string]string{
		"/test.ts": `
			const myVariable = 1;
			function myFunction() {}
			interface MyInterface {}
		`,
	})

	sf := project.GetSourceFile("/test.ts")
	assert.NotNil(t, sf)

	var nameExtractionSuccess int

	sf.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsDeclaration() {
			if name, ok := node.GetNodeName(); ok {
				nameExtractionSuccess++
				assert.NotEmpty(t, name, "提取的名称不应该为空")
			}
		}
	})

	assert.Greater(t, nameExtractionSuccess, 0, "应该成功提取名称")
}

// TestProjectCreation 测试项目创建
func TestProjectCreation(t *testing.T) {
	// 测试内存项目创建
	project := tsmorphgo.NewProjectFromSources(map[string]string{
		"/test.ts": `const x = 1;`,
	})

	assert.NotNil(t, project)

	sf := project.GetSourceFile("/test.ts")
	assert.NotNil(t, sf)

	// 清理
	project.Close()
}

// TestSourceFileOperations 测试源文件操作
func TestSourceFileOperations(t *testing.T) {
	project := createTestProject(map[string]string{
		"/test.ts": `
			const x = 1;
			function test() { return x; }
		`,
	})

	sf := project.GetSourceFile("/test.ts")
	assert.NotNil(t, sf)

	// 测试获取文件路径
	assert.Equal(t, "/test.ts", sf.GetFilePath())

	// 测试获取文件内容
	fileResult := sf.GetFileResult()
	assert.NotNil(t, fileResult)
	assert.Contains(t, fileResult.Raw, "const x = 1;")
	assert.Contains(t, fileResult.Raw, "function test()")

	// 测试节点遍历
	var nodeCount int
	sf.ForEachDescendant(func(node tsmorphgo.Node) {
		nodeCount++
	})
	assert.Greater(t, nodeCount, 0, "应该找到节点")

	// 清理
	project.Close()
}

// TestNodeNavigation 测试节点导航
func TestNodeNavigation(t *testing.T) {
	project := createTestProject(map[string]string{
		"/test.ts": `
			function outer() {
				const x = 1;
				return x;
			}
		`,
	})

	sf := project.GetSourceFile("/test.ts")
	assert.NotNil(t, sf)

	var functionNode *tsmorphgo.Node
	var variableNode *tsmorphgo.Node

	sf.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsFunctionDeclaration() {
			functionNode = &node
		}
		if node.IsVariableDeclaration() {
			variableNode = &node
		}
	})

	assert.NotNil(t, functionNode, "应该找到函数节点")
	assert.NotNil(t, variableNode, "应该找到变量节点")

	// 测试父节点导航
	if variableNode != nil {
		parent := variableNode.GetParent()
		assert.NotNil(t, parent, "变量节点应该有父节点")
	}

	// 清理
	project.Close()
}