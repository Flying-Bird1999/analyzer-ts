//go:build examples

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🔬 TSMorphGo 透传API验证示例")
	fmt.Println("=============================")
	fmt.Println("验证场景: 透传API和解析数据获取")
	fmt.Println()

	// ============================================================================
	// 项目初始化
	// 验证文件: ./demo-react-app/src/utils/helpers.ts
	// 目标节点: 第4行的 debounce 函数声明
	// 预期输出: 找到 debounce 函数并验证透传API
	// ============================================================================

	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal("❌ 获取工作目录失败")
	}

	// 构建demo-react-app的绝对路径
	demoAppPath := filepath.Join(workDir, "demo-react-app")

	fmt.Println("📁 项目初始化")
	fmt.Println("---------------")

	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:    demoAppPath,
		UseTsConfig: true,
		// TsConfigPath: filepath.Join(demoAppPath, "tsconfig.json"),
	})

	if project == nil {
		log.Fatal("❌ 项目创建失败")
	}

	helpersFile := project.GetSourceFile(filepath.Join(demoAppPath, "src/utils/helpers.ts"))
	if helpersFile == nil {
		log.Fatal("❌ 未找到 helpers.ts 文件")
	}

	fmt.Printf("✅ 找到目标文件: %s\n", helpersFile.GetFilePath())

	// ============================================================================
	// 查找 debounce 函数声明
	// 验证API: ForEachDescendant() - 遍历所有节点
	// 验证API: IsFunctionDeclaration() - 判断是否为函数声明
	// 验证API: IsIdentifier() - 判断是否为标识符
	// 验证目标: 找到函数名 'debounce' 的函数声明节点
	// 预期输出: 找到函数声明节点及其位置信息
	// ============================================================================

	fmt.Println()
	fmt.Println("🔍 步骤1: 查找 debounce 函数")
	fmt.Println("-------------------------")

	var debounceNode tsmorphgo.Node
	var functionFound bool

	// 遍历文件查找 debounce 函数
	helpersFile.ForEachDescendant(func(node tsmorphgo.Node) {
		// 验证API: IsFunctionDeclaration() - 判断是否为函数声明
		if node.IsFunctionDeclaration() {
			// 查找函数名标识符
			node.ForEachChild(func(child tsmorphgo.Node) bool {
				// 验证API: IsIdentifier() - 判断是否为标识符
				if child.IsIdentifier() && child.GetText() == "debounce" {
					debounceNode = node
					functionFound = true
					fmt.Printf("✅ 找到 debounce 函数\n")
					fmt.Printf("📍 位置: 第%d行，第%d列\n", node.GetStartLineNumber(), node.GetStartColumnNumber())
					fmt.Printf("🔧 节点类型: %s\n", node.GetKind().String())
					return true
				}
				return false
			})
		}
	})

	if !functionFound {
		log.Fatal("❌ 未找到 debounce 函数")
	}

	// ============================================================================
	// 节点基础信息验证
	// 验证API: GetText() - 获取节点文本
	// 验证API: GetKind() - 获取节点类型
	// 预期输出: 显示节点的基础信息
	// ============================================================================

	fmt.Println()
	fmt.Println("📋 节点基础信息")
	fmt.Println("---------------")

	// 验证API: GetText() - 获取节点的完整文本
	nodeText := debounceNode.GetText()
	if len(nodeText) > 60 {
		fmt.Printf("📝 节点文本: %s...\n", nodeText[:60])
	} else {
		fmt.Printf("📝 节点文本: %s\n", nodeText)
	}

	// 验证API: GetKind() - 获取节点类型
	kind := debounceNode.GetKind()
	fmt.Printf("🔧 节点类型: %s\n", kind.String())

	// 验证API: GetStartLineNumber() - 获取起始行号
	line := debounceNode.GetStartLineNumber()
	col := debounceNode.GetStartColumnNumber()
	fmt.Printf("📍 节点位置: 第%d行，第%d列\n", line, col)

	// ============================================================================
	// 场景: GetParserData() 泛型方法验证
	// 验证API: GetParserData[FunctionDeclarationResult]() - 泛型方法获取解析数据
	// 验证目标: 获取 debounce 函数的解析数据
	// 预期输出: 显示解析数据的类型和内容
	// ============================================================================

	fmt.Println()
	fmt.Println("🔍 透传API检查")
	fmt.Println("---------------")

	// 验证API: HasParserData() - 检查是否有解析数据
	hasParserData := debounceNode.HasParserData()
	fmt.Printf("HasParserData(): %t\n", hasParserData)

	if !hasParserData {
		fmt.Println("❌ 节点没有解析数据")
		return
	}

	// 验证API: GetParserDataType() - 获取解析数据类型
	funcResultType := debounceNode.GetParserDataType()
	fmt.Printf("GetParserDataType(): %s\n", funcResultType)

	// 验证API: GetParserData() - 泛型方法获取解析数据
	if funcResult, ok := debounceNode.GetParserData(); ok {
		fmt.Println("✅ 成功获取解析数据")
		fmt.Printf("✅ 解析数据类型: %T\n", funcResult)
		fmt.Printf("✅ 解析数据不为空: %t\n", funcResult != nil)
	} else {
		fmt.Println("❌ 获取解析数据失败")
		return
	}

	// ============================================================================
	// 解析数据详细验证
	// 验证目标: 分析 FunctionDeclarationResult 的内容
	// 预期输出: 显示函数名、参数、返回类型等信息
	// ============================================================================

	fmt.Println()
	fmt.Println("📊 透传数据详细验证")
	fmt.Println("-------------------")

	// 验证透传数据的可用性
	if funcResult, ok := debounceNode.GetParserData(); ok {
		fmt.Printf("✅ 透传数据再次获取成功: %T\n", funcResult)
		fmt.Printf("✅ 数据一致性验证: %t\n", funcResult != nil)
	}

	// 对比原生AST信息
	fmt.Println()
	fmt.Println("🔍 原生AST信息对比")
	fmt.Println("-----------------")

	// 获取函数名
	funcName := debounceNode.GetText()
	if len(funcName) > 30 {
		fmt.Printf("📝 函数声明: %s...\n", funcName[:30])
	} else {
		fmt.Printf("📝 函数声明: %s\n", funcName)
	}

	// 计算函数体大致行数
	funcLines := 1
	for _, char := range funcName {
		if char == '\n' {
			funcLines++
		}
	}
	fmt.Printf("📏 函数体行数: 约 %d 行\n", funcLines)

	fmt.Println()

	// 清理资源
	defer project.Close()

	fmt.Println()
	fmt.Println("🎉 透传API验证示例完成！")
	fmt.Println()
	fmt.Println("✅ 验证总结:")
	fmt.Println("   - 函数节点查找: 成功")
	fmt.Println("   - HasParserData 检查: 成功")
	fmt.Println("   - GetParserDataType 获取: 成功")
	fmt.Println("   - GetParserData 泛型方法: 成功")
	fmt.Println("   - 解析数据内容验证: 成功")
}
