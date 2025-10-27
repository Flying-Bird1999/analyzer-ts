package tsmorphgo

import (
	"strings"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
	"github.com/stretchr/testify/assert"
)

// declaration_test.go
//
// 这个文件包含了 TypeScript 声明处理功能的测试用例，专注于验证 tsmorphgo 对各种
// TypeScript 声明节点的识别、转换和信息提取能力。
//
// 主要测试场景：
// 1. 变量声明处理 - 测试变量名称获取和节点识别
// 2. 函数声明处理 - 验证函数名称提取和声明结构分析
// 3. 接口声明处理 - 测试接口定义的识别和转换
// 4. 类型别名声明处理 - 验证类型别名的识别和信息提取
// 5. 枚举声明处理 - 测试枚举定义的识别和转换
// 6. 导入声明处理 - 验证各种导入语法的识别和别名处理
// 7. 类型转换API - 测试 AsXXX 系列函数的类型安全性
// 8. 边缘情况处理 - 验证无效节点和异常情况的处理
//
// 测试目标：
// - 验证各种声明类型的正确识别和分类
// - 确保类型转换 API 的类型安全和错误处理
// - 测试声明节点的信息提取准确性
// - 验证在异常情况下的系统稳定性

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
			name := strings.TrimSpace(node.Node.AsImportSpecifier().Name().Text())
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

// TestTypeConversionAPIs 测试类型转换API的全面功能
func TestTypeConversionAPIs(t *testing.T) {
	// 测试用例 1: 导入声明转换
	t.Run("ImportDeclarationConversion", func(t *testing.T) {
		project := createTestProject(map[string]string{"/test_import.ts": `
			import { Component, PropTypes as PT } from 'react';
			import * as React from 'react';
			import DefaultComponent from './component';
		`})
		sf := project.GetSourceFile("/test_import.ts")
		assert.NotNil(t, sf)

		var importDecls []*Node
		sf.ForEachDescendant(func(node Node) {
			if node.Kind == ast.KindImportDeclaration {
				importDecls = append(importDecls, &node)
			}
		})

		// 测试每个导入声明的转换
		for _, importDecl := range importDecls {
			result, ok := AsImportDeclaration(*importDecl)
			if ok {
				// 验证结果不为空
				assert.NotNil(t, result)
				// 导入声明应该有相关的信息
				assert.NotNil(t, result)
			}
		}
	})

	// 测试用例 2: 变量声明转换
	t.Run("VariableDeclarationConversion", func(t *testing.T) {
		project := createTestProject(map[string]string{"/test_vars.ts": `
			const x = 1;
			let y = "hello";
			var z = true;
			const { a, b } = obj;
		`})
		sf := project.GetSourceFile("/test_vars.ts")
		assert.NotNil(t, sf)

		var varDecls []*Node
		sf.ForEachDescendant(func(node Node) {
			if IsVariableDeclaration(node) {
				varDecls = append(varDecls, &node)
			}
		})

		// 应该找到4个变量声明
		assert.Len(t, varDecls, 4)

		// 测试每个变量声明的转换
		for _, varDecl := range varDecls {
			result, ok := AsVariableDeclaration(*varDecl)
			if ok {
				// 验证结果不为空
				assert.NotNil(t, result)
			}
		}

		// 测试 GetVariableDeclarationNameNode API
		assert.NotNil(t, varDecls[0])
		nameNode, ok := GetVariableDeclarationNameNode(*varDecls[0])
		assert.True(t, ok)
		assert.NotNil(t, nameNode)
		assert.Equal(t, "x", strings.TrimSpace(nameNode.GetText()))
	})

	// 测试用例 3: 函数声明转换
	t.Run("FunctionDeclarationConversion", func(t *testing.T) {
		project := createTestProject(map[string]string{"/test_funcs.ts": `
			function regularFunction(param: string): number {
				return 42;
			}

			export function exportedFunction(): void {
				console.log("exported");
			}
		`})
		sf := project.GetSourceFile("/test_funcs.ts")
		assert.NotNil(t, sf)

		var funcDecls []*Node
		sf.ForEachDescendant(func(node Node) {
			if IsFunctionDeclaration(node) {
				funcDecls = append(funcDecls, &node)
			}
		})

		// 应该找到2个函数声明
		assert.Len(t, funcDecls, 2)

		// 测试每个函数声明的转换
		for _, funcDecl := range funcDecls {
			result, ok := AsFunctionDeclaration(*funcDecl)
			if ok {
				// 验证结果不为空
				assert.NotNil(t, result)
			}
		}
	})

	// 测试用例 4: 接口声明转换
	t.Run("InterfaceDeclarationConversion", func(t *testing.T) {
		project := createTestProject(map[string]string{"/test_interfaces.ts": `
		 interface User {
				id: number;
				name: string;
				email?: string;
			}

			export interface Admin extends User {
				permissions: string[];
			}
		`})
		sf := project.GetSourceFile("/test_interfaces.ts")
		assert.NotNil(t, sf)

		var interfaceDecls []*Node
		sf.ForEachDescendant(func(node Node) {
			if IsInterfaceDeclaration(node) {
				interfaceDecls = append(interfaceDecls, &node)
			}
		})

		// 应该找到2个接口声明
		assert.Len(t, interfaceDecls, 2)

		// 测试每个接口声明的转换
		for _, interfaceDecl := range interfaceDecls {
			result, ok := AsInterfaceDeclaration(*interfaceDecl)
			if ok {
				// 验证结果不为空
				assert.NotNil(t, result)
			}
		}
	})

	// 测试用例 5: 类型别名声明转换
	t.Run("TypeAliasDeclarationConversion", func(t *testing.T) {
		project := createTestProject(map[string]string{"/test_types.ts": `
		 type UserID = number;
		 type UserName = string;
		 type User = {
			 id: UserID;
			 name: UserName;
			};
		`})
		sf := project.GetSourceFile("/test_types.ts")
		assert.NotNil(t, sf)

		var typeAliasDecls []*Node
		sf.ForEachDescendant(func(node Node) {
			if IsTypeAliasDeclaration(node) {
				typeAliasDecls = append(typeAliasDecls, &node)
			}
		})

		// 应该找到3个类型别名声明
		assert.Len(t, typeAliasDecls, 3)

		// 测试每个类型别名声明的转换
		for _, typeAliasDecl := range typeAliasDecls {
			result, ok := AsTypeAliasDeclaration(*typeAliasDecl)
			if ok {
				// 验证结果不为空
				assert.NotNil(t, result)
			}
		}
	})

	// 测试用例 6: 枚举声明转换
	t.Run("EnumDeclarationConversion", func(t *testing.T) {
		project := createTestProject(map[string]string{"/test_enums.ts": `
			enum Color {
				Red,
				Green,
				Blue
			}

			enum Status {
				Active = "ACTIVE",
				Inactive = "INACTIVE"
			}
		`})
		sf := project.GetSourceFile("/test_enums.ts")
		assert.NotNil(t, sf)

		var enumDecls []*Node
		sf.ForEachDescendant(func(node Node) {
			if IsEnumDeclaration(node) {
				enumDecls = append(enumDecls, &node)
			}
		})

		// 应该找到2个枚举声明
		assert.Len(t, enumDecls, 2)

		// 测试每个枚举声明的转换
		for _, enumDecl := range enumDecls {
			result, ok := AsEnumDeclaration(*enumDecl)
			if ok {
				// 验证结果不为空
				assert.NotNil(t, result)
			}
		}
	})

	// 测试用例 7: 无效节点的转换
	t.Run("InvalidNodeConversions", func(t *testing.T) {
		project := createTestProject(map[string]string{"/test_invalid_conv.ts": `const x = 1;`})
		sf := project.GetSourceFile("/test_invalid_conv.ts")
		assert.NotNil(t, sf)

		var identifierNode *Node
		sf.ForEachDescendant(func(node Node) {
			if IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "x" {
				identifierNode = &node
			}
		})
		assert.NotNil(t, identifierNode)

		// 测试在标识符节点上调用各种转换API（应该都失败）
		_, ok := AsImportDeclaration(*identifierNode)
		assert.False(t, ok)

		_, ok = AsVariableDeclaration(*identifierNode)
		assert.False(t, ok)

		_, ok = AsFunctionDeclaration(*identifierNode)
		assert.False(t, ok)

		_, ok = AsInterfaceDeclaration(*identifierNode)
		assert.False(t, ok)

		_, ok = AsTypeAliasDeclaration(*identifierNode)
		assert.False(t, ok)

		_, ok = AsEnumDeclaration(*identifierNode)
		assert.False(t, ok)
	})

	// 测试用例 8: 没有sourceFile的节点
	t.Run("NodeWithoutSourceFile", func(t *testing.T) {
		project := createTestProject(map[string]string{"/test_no_source.ts": `const x = 1;`})
		sf := project.GetSourceFile("/test_no_source.ts")
		assert.NotNil(t, sf)

		// 创建一个没有sourceFile的节点
		invalidNode := Node{
			Node:       sf.astNode,
			sourceFile: nil,
		}

		// 所有转换都应该失败
		_, ok := AsImportDeclaration(invalidNode)
		assert.False(t, ok)

		_, ok = AsVariableDeclaration(invalidNode)
		assert.False(t, ok)

		_, ok = AsFunctionDeclaration(invalidNode)
		assert.False(t, ok)

		_, ok = AsInterfaceDeclaration(invalidNode)
		assert.False(t, ok)
	})
}