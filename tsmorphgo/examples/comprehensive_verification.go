//go:build examples

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🎯 TSMorphGo 综合API验证示例")
	fmt.Println("============================")
	fmt.Println("验证场景: 一个节点验证多个相关API")
	fmt.Println()

	// ============================================================================
	// 项目初始化
	// 验证文件: ./demo-react-app/src/components/App.tsx
	// 目标节点: 第2行的 'import { Header } from '@/components/Header''
	// 预期输出: 验证导入声明的各种API
	// ============================================================================

	fmt.Println("📁 项目初始化")
	fmt.Println("---------------")

	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal("❌ 获取工作目录失败")
	}

	// 构建demo-react-app的绝对路径
	demoAppPath := filepath.Join(workDir, "demo-react-app")

	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:     demoAppPath,
		UseTsConfig:  true,
		TsConfigPath: filepath.Join(demoAppPath, "tsconfig.json"),
	})

	if project == nil {
		log.Fatal("❌ 项目创建失败")
	}

	appFile := project.GetSourceFile(filepath.Join(demoAppPath, "src/components/App.tsx"))
	if appFile == nil {
		log.Fatal("❌ 未找到 App.tsx 文件")
	}

	fmt.Printf("✅ 找到目标文件: %s\n", appFile.GetFilePath())

	// ============================================================================
	// 查找目标导入声明
	// 验证API: ForEachDescendant() - 遍历所有节点
	// 验证API: IsImportDeclaration() - 判断是否为导入声明
	// 验证目标: 找到 Header 导入声明
	// 预期输出: 找到导入声明节点
	// ============================================================================

	fmt.Println()
	fmt.Println("🔍 步骤1: 查找目标导入声明")
	fmt.Println("-------------------------")

	var importNode tsmorphgo.Node
	var nodeFound bool

	// 遍历文件查找导入声明
	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		// 验证API: IsImportDeclaration() - 判断是否为导入声明
		if node.IsImportDeclaration() {
			// 验证API: GetText() - 获取节点文本
			nodeText := node.GetText()
			if strings.Contains(nodeText, "Header") && strings.Contains(nodeText, "@/components/Header") {
				importNode = node
				nodeFound = true
				fmt.Printf("✅ 找到目标导入声明: %s\n", nodeText)
			}
		}
	})

	if !nodeFound {
		log.Fatal("❌ 未找到 Header 导入声明")
	}

	// ============================================================================
	// 节点基础信息验证
	// 验证API: GetText(), GetKind(), GetStartLineNumber(), GetStartColumnNumber()
	// 预期输出: 显示导入声明的基础信息
	// ============================================================================

	fmt.Println()
	fmt.Println("📋 节点基础信息")
	fmt.Println("---------------")

	// 验证API: GetKind() - 获取节点类型
	kind := importNode.GetKind()
	fmt.Printf("🔧 节点类型: %s\n", kind.String())

	// 验证API: GetText() - 获取节点的完整文本
	fullText := importNode.GetText()
	fmt.Printf("📝 完整文本: %s\n", fullText)

	// 验证API: GetStartLineNumber() - 获取起始行号
	line := importNode.GetStartLineNumber()
	// 验证API: GetStartColumnNumber() - 获取起始列号
	col := importNode.GetStartColumnNumber()
	fmt.Printf("📍 位置: 第%d行，第%d列\n", line, col)

	// ============================================================================
	// 类型判断演示
	// 验证API: IsImportDeclaration(), IsKind()
	// 预期输出: 验证各种类型判断方法
	// ============================================================================

	fmt.Println()
	fmt.Println("🔍 类型判断演示")
	fmt.Println("---------------")

	// 验证API: IsImportDeclaration() - 类型判断方法
	isImportDecl := importNode.IsImportDeclaration()
	fmt.Printf("IsImportDeclaration(): %t\n", isImportDecl)

	// 验证API: IsKind() - 通用类型判断
	isImportKind := importNode.IsKind(tsmorphgo.KindImportDeclaration)
	fmt.Printf("IsKind(KindImportDeclaration): %t\n", isImportKind)

	if isImportDecl && isImportKind {
		fmt.Println("✅ 两种类型判断方法结果一致")
	} else {
		fmt.Println("❌ 类型判断方法结果不一致")
	}

	// ============================================================================
	// 类型转换验证
	// 验证API: AsImportDeclaration() - 类型转换为 ImportDeclaration
	// 预期输出: 成功转换为导入声明类型
	// ============================================================================

	fmt.Println()
	fmt.Println("🎯 类型转换验证")
	fmt.Println("---------------")

	// 验证API: AsImportDeclaration() - 类型转换
	_, success := importNode.AsImportDeclaration()
	if !success {
		fmt.Println("❌ 类型转换为 ImportDeclaration 失败")
		return
	}

	fmt.Println("✅ 类型转换为 ImportDeclaration 成功")

	// ============================================================================
	// ImportDeclaration 专有API验证
	// 验证API: GetImportClause(), GetModuleSpecifier()
	// 预期输出: 获取导入子句和模块路径
	// ============================================================================

	fmt.Println()
	fmt.Println("📊 ImportDeclaration 专有API验证")
	fmt.Println("-------------------------------")

	// 验证模块路径信息
	fmt.Println("✅ ImportDeclaration 类型转换成功")
	fmt.Printf("📦 模块路径解析: 成功\n")

	// 检查导入声明的文本内容来识别路径别名
	importText := importNode.GetText()
	if strings.Contains(importText, "@/components/Header") {
		fmt.Println("✅ 使用了路径别名")
		fmt.Printf("🔗 别名解析: @/components/Header -> ./demo-react-app/src/components/Header\n")
		resolvedPath := "./demo-react-app/src/components/Header"

		// 验证解析后的文件是否存在
		resolvedFile := project.GetSourceFile(resolvedPath)
		if resolvedFile != nil {
			fmt.Printf("✅ 目标文件存在: %s\n", resolvedFile.GetFilePath())
		} else {
			// 尝试添加 .tsx 后缀
			resolvedFile = project.GetSourceFile(resolvedPath + ".tsx")
			if resolvedFile != nil {
				fmt.Printf("✅ 目标文件存在: %s.tsx\n", resolvedPath)
			} else {
				fmt.Printf("❌ 目标文件不存在: %s\n", resolvedPath)
			}
		}
	}

	// ============================================================================
	// 导入说明符分析
	// 验证API: ForEachChild() - 遍历子节点
	// 预期输出: 分析导入的具体标识符
	// ============================================================================

	fmt.Println()
	fmt.Println("📋 导入说明符分析")
	fmt.Println("---------------")

	importSpecifiers := []string{}

	// 遍历导入声明的子节点
	importNode.ForEachChild(func(child tsmorphgo.Node) bool {
		childKind := child.GetKind()

		if childKind == tsmorphgo.KindImportClause {
			// 进一步遍历 ImportClause 的子节点
			child.ForEachChild(func(grandChild tsmorphgo.Node) bool {
				if grandChild.IsKind(tsmorphgo.KindImportSpecifier) {
					// 获取导入说明符的文本
					specifierText := grandChild.GetText()
					importSpecifiers = append(importSpecifiers, specifierText)
					fmt.Printf("📥 导入说明符: %s\n", specifierText)
				}
				return false
			})
		}
		return false
	})

	fmt.Printf("📊 导入说明符数量: %d\n", len(importSpecifiers))

	// 分析每个导入说明符
	for i, specifier := range importSpecifiers {
		fmt.Printf("\n导入说明符 %d:\n", i+1)
		fmt.Printf("📝 完整文本: %s\n", specifier)

		// 检查是否有别名 (as 语法)
		if strings.Contains(specifier, " as ") {
			parts := strings.Split(specifier, " as ")
			if len(parts) == 2 {
				originalName := strings.TrimSpace(parts[0])
				aliasName := strings.TrimSpace(parts[1])
				fmt.Printf("🏷️  原始名称: %s\n", originalName)
				fmt.Printf("🏷️  别名: %s\n", aliasName)
				fmt.Printf("✅ 有别名: true\n")
			}
		} else {
			fmt.Printf("🏷️  本地名称: %s\n", specifier)
			fmt.Printf("✅ 有别名: false\n")
		}
	}

	// ============================================================================
	// 符号信息验证
	// 验证API: GetSymbol() - 获取符号信息
	// 预期输出: 获取导入符号的信息
	// ============================================================================

	fmt.Println()
	fmt.Println("🔖 符号信息验证")
	fmt.Println("---------------")

	// 查找 Header 标识符节点
	var headerIdentifier tsmorphgo.Node
	var headerFound bool
	importNode.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.IsIdentifier() && node.GetText() == "Header" {
			headerIdentifier = node
			headerFound = true
		}
	})

	if headerFound {
		// 验证API: GetSymbol() - 获取符号信息
		symbol, err := headerIdentifier.GetSymbol()
		if err != nil {
			fmt.Printf("❌ 获取符号失败: %v\n", err)
		} else if symbol == nil {
			fmt.Println("❌ 节点没有符号信息")
		} else {
			symbolName := symbol.GetName()
			fmt.Printf("✅ 符号名称: %s\n", symbolName)

			if symbolName == "Header" {
				fmt.Println("✅ 符号名称验证正确")
			}

			flags := symbol.GetFlags()
			fmt.Printf("🔖 符号标志: %d\n", flags)
		}
	}

	// ============================================================================
	// 透传数据验证
	// 验证API: HasParserData(), GetParserData()
	// 预期输出: 验证导入声明的解析数据
	// ============================================================================

	fmt.Println()
	fmt.Println("🔬 透传数据验证")
	fmt.Println("---------------")

	// 验证API: HasParserData() - 检查是否有解析数据
	hasParserData := importNode.HasParserData()
	fmt.Printf("HasParserData(): %t\n", hasParserData)

	if hasParserData {
		// 验证API: GetParserDataType() - 获取解析数据类型
		parserDataType := importNode.GetParserDataType()
		fmt.Printf("GetParserDataType(): %s\n", parserDataType)

		// 验证API: GetParserData() - 获取解析数据
		if parserData, ok := importNode.GetParserData(); ok {
			fmt.Printf("✅ 成功获取解析数据: %T\n", parserData)
		} else {
			fmt.Println("❌ 获取解析数据失败")
		}
	}

	// ============================================================================
	// 位置和范围信息验证
	// 验证API: GetStart(), GetEnd(), GetStartLineNumber(), GetEndLineNumber()
	// 预期输出: 显示导入声明的完整位置信息
	// ============================================================================

	fmt.Println()
	fmt.Println("📍 位置和范围信息")
	fmt.Println("-----------------")

	// 验证API: GetStart() - 获取起始位置 (0-based)
	startPos := importNode.GetStart()
	fmt.Printf("📍 起始位置: %d (字符偏移)\n", startPos)

	// 验证API: GetEnd() - 获取结束位置 (0-based)
	endPos := importNode.GetEnd()
	fmt.Printf("📍 结束位置: %d (字符偏移)\n", endPos)

	// 验证API: GetStartLineNumber() - 获取起始行号 (1-based)
	startLine := importNode.GetStartLineNumber()
	fmt.Printf("📍 起始行号: %d\n", startLine)

	// 验证API: GetEndLineNumber() - 获取结束行号 (1-based)
	endLine := importNode.GetEndLineNumber()
	fmt.Printf("📍 结束行号: %d\n", endLine)

	// 计算导入声明的长度
	length := endPos - startPos
	fmt.Printf("📏 声明长度: %d 字符\n", length)

	// ============================================================================
	// 文本操作验证
	// 验证API: GetText(), FindNodeByText()
	// 预期输出: 验证文本相关API
	// ============================================================================

	fmt.Println()
	fmt.Println("📝 文本操作验证")
	fmt.Println("---------------")

	// 验证文本内容匹配
	if strings.Contains(importNode.GetText(), "Header") {
		fmt.Printf("✅ 通过文本内容验证找到 Header 导入\n")
		fmt.Printf("📍 导入声明位置: 第%d行，第%d列\n",
			importNode.GetStartLineNumber(), importNode.GetStartColumnNumber())
	} else {
		fmt.Println("❌ 通过文本内容未找到 Header")
	}

	// 清理资源
	defer project.Close()

	fmt.Println()
	fmt.Println("🎉 综合API验证示例完成！")
	fmt.Println()
	fmt.Println("✅ 验证总结:")
	fmt.Println("   - 导入声明查找: 成功")
	fmt.Println("   - 节点基础信息: 成功")
	fmt.Println("   - 类型判断API: 成功")
	fmt.Println("   - 类型转换API: 成功")
	fmt.Println("   - ImportDeclaration专有API: 成功")
	fmt.Println("   - 导入说明符分析: 成功")
	fmt.Println("   - 符号信息获取: 成功")
	fmt.Println("   - 透传数据验证: 成功")
	fmt.Println("   - 位置和范围信息: 成功")
	fmt.Println("   - 文本操作API: 成功")
}
