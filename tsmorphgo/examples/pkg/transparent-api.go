//go:build transparent_api
// +build transparent_api

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🔍 TSMorphGo 透传API - Parser数据验证")
	fmt.Println("=" + strings.Repeat("=", 50))

	// 创建内存项目进行深度验证
	project := tsmorphgo.NewProjectFromSources(map[string]string{
		"/src/complex.ts": `
			// 复杂的函数调用
			const result = calculateSum(1, 2, 3);

			// 变量声明
			const API_URL = 'https://api.example.com';
			const TIMEOUT = 5000;

			// 复杂的变量解构
			const { name, email: userEmail } = user;

			// 接口声明
			interface User {
				id: number;
				name: string;
				email?: string;
			}

			// 函数声明
			function calculateSum(...numbers: number[]): number {
				return numbers.reduce((a, b) => a + b, 0);
			}

			// 导入声明
			import { useEffect, useState } from 'react';
			import { add } from './utils';
			import axios from 'axios';
		`,
	})
	defer project.Close()

	sourceFiles := project.GetSourceFiles()
	if len(sourceFiles) == 0 {
		log.Fatal("❌ 未找到源文件")
	}

	sourceFile := sourceFiles[0]
	fmt.Printf("📁 深度验证文件: %s\n\n", sourceFile.GetFilePath())

	// 详细验证每种类型的解析数据
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if !node.HasParserData() {
			return
		}

		// 获取通用解析数据
		data, ok := node.GetParserData()
		if !ok {
			return
		}

		// 详细分析每种类型
		switch node.GetKind() {
		case tsmorphgo.KindCallExpression:
			verifyCallExpression(node, data)

		case tsmorphgo.KindVariableStatement:
			verifyVariableDeclaration(node, data)

		case tsmorphgo.KindInterfaceDeclaration:
			verifyInterfaceDeclaration(node, data)

		case tsmorphgo.KindFunctionDeclaration:
			verifyFunctionDeclaration(node, data)

		case tsmorphgo.KindImportDeclaration:
			verifyImportDeclaration(node, data)
		}
	})

	fmt.Println("\n✅ Parser数据结构验证完成!")
}

// verifyCallExpression 验证函数调用表达式的解析数据
func verifyCallExpression(node tsmorphgo.Node, data interface{}) {
	fmt.Printf("🔍 验证 CallExpression 节点:\n")
	fmt.Printf("   节点文本: %s\n", node.GetText())
	fmt.Printf("   数据类型: %T\n", data)

	// 使用便利方法获取具体结构
	if callExpr, ok := node.AsCallExpression(); ok {
		fmt.Printf("   ✅ 成功转换为 parser.CallExpression\n")
		fmt.Printf("   📞 调用链: %v\n", callExpr.CallChain)
		fmt.Printf("   🔢 参数数量: %d\n", len(callExpr.Arguments))

		// 详细分析每个参数
		for i, arg := range callExpr.Arguments {
			fmt.Printf("      参数%d: Type=%s, Expression=%s\n",
				i+1, arg.Type, arg.Expression)
		}

		// 检查内联函数
		if len(callExpr.InlineFunctions) > 0 {
			fmt.Printf("   🔧 内联函数数量: %d\n", len(callExpr.InlineFunctions))
			for i, inlineFn := range callExpr.InlineFunctions {
				fmt.Printf("      内联函数%d: %s\n", i+1, inlineFn.Identifier)
			}
		}

		// 验证数据来源和属性
		fmt.Printf("   📍 原始文本: %s\n", callExpr.Raw)
		if callExpr.SourceLocation != nil {
			fmt.Printf("   📐 位置信息: %+v\n", callExpr.SourceLocation)
		}

	} else {
		fmt.Printf("   ❌ 转换失败\n")
	}
	fmt.Println()
}

// verifyVariableDeclaration 验证变量声明的解析数据
func verifyVariableDeclaration(node tsmorphgo.Node, data interface{}) {
	fmt.Printf("📦 验证 VariableDeclaration 节点:\n")
	fmt.Printf("   节点文本: %s\n", node.GetText())
	fmt.Printf("   数据类型: %T\n", data)

	if varDecl, ok := node.AsVariableDeclaration(); ok {
		fmt.Printf("   ✅ 成功转换为 parser.VariableDeclaration\n")
		fmt.Printf("   🔖 声明类型: %s\n", varDecl.Kind)
		fmt.Printf("   📤 是否导出: %t\n", varDecl.Exported)
		fmt.Printf("   🔢 声明器数量: %d\n", len(varDecl.Declarators))

		// 详细分析每个声明器
		for i, decl := range varDecl.Declarators {
			fmt.Printf("      声明器%d:\n", i+1)
			fmt.Printf("        变量名: %s\n", decl.Identifier)
			if decl.PropName != decl.Identifier {
				fmt.Printf("        属性名: %s (有别名)\n", decl.PropName)
			}

			if decl.Type != nil {
				fmt.Printf("        类型注解: %s (%s)\n", decl.Type.Type, decl.Type.Expression)
			}

			if decl.InitValue != nil {
				fmt.Printf("        初始值: %s (%s)\n", decl.InitValue.Type, decl.InitValue.Expression)
			}
		}

		// 检查解构赋值源
		if varDecl.Source != nil {
			fmt.Printf("   🔗 解构源: %s (%s)\n", varDecl.Source.Type, varDecl.Source.Expression)
		}

	} else {
		fmt.Printf("   ❌ 转换失败\n")
	}
	fmt.Println()
}

// verifyInterfaceDeclaration 验证接口声明的解析数据
func verifyInterfaceDeclaration(node tsmorphgo.Node, data interface{}) {
	fmt.Printf("🔌 验证 InterfaceDeclaration 节点:\n")
	fmt.Printf("   节点文本: %s\n", node.GetText())
	fmt.Printf("   数据类型: %T\n", data)

	if interfaceDecl, ok := node.AsInterfaceDeclaration(); ok {
		fmt.Printf("   ✅ 成功转换为 parser.InterfaceDeclarationResult\n")

		// 注意：根据实际的结构调整字段访问
		fmt.Printf("   🏷️ 接口信息: %+v\n", interfaceDecl)

		// 检查常见字段
		if interfaceDecl.Raw != "" {
			fmt.Printf("   📄 原始文本长度: %d\n", len(interfaceDecl.Raw))
		}

	} else {
		fmt.Printf("   ❌ 转换失败\n")
	}
	fmt.Println()
}

// verifyFunctionDeclaration 验证函数声明的解析数据
func verifyFunctionDeclaration(node tsmorphgo.Node, data interface{}) {
	fmt.Printf("🔧 验证 FunctionDeclaration 节点:\n")
	fmt.Printf("   节点文本: %s\n", node.GetText())
	fmt.Printf("   数据类型: %T\n", data)

	if funcDecl, ok := node.AsFunctionDeclaration(); ok {
		fmt.Printf("   ✅ 成功转换为 parser.FunctionDeclarationResult\n")

		// 注意：根据实际的结构调整字段访问
		fmt.Printf("   🔧 函数信息: %+v\n", funcDecl)

	} else {
		fmt.Printf("   ❌ 转换失败\n")
	}
	fmt.Println()
}

// verifyImportDeclaration 验证导入声明的解析数据
func verifyImportDeclaration(node tsmorphgo.Node, data interface{}) {
	fmt.Printf("📥 验证 ImportDeclaration 节点:\n")
	fmt.Printf("   节点文本: %s\n", node.GetText())
	fmt.Printf("   数据类型: %T\n", data)

	if importDecl, ok := node.AsImportDeclaration(); ok {
		fmt.Printf("   ✅ 成功转换为 projectParser.ImportDeclarationResult\n")

		// 注意：根据实际的结构调整字段访问
		fmt.Printf("   📥 导入信息: %+v\n", importDecl)

	} else {
		fmt.Printf("   ❌ 转换失败\n")
	}
	fmt.Println()
}
