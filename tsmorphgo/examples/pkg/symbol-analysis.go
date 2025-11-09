//go:build symbol_analysis
// +build symbol_analysis

package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🧬 TSMorphGo - 符号系统深度分析")
	fmt.Println("=" + strings.Repeat("=", 50))

	// =============================================================================
	// 本文件演示如何利用 tsmorphgo 的符号系统进行高级的、语义级别的代码分析。
	//
	// 核心 API:
	// - GetSymbol(node): 从一个节点获取其关联的符号。
	// - symbol.GetName(): 获取符号的名称。
	// - symbol.GetDeclarations(): 获取符号的所有声明节点。
	//
	// 为什么使用符号?
	// 符号是 TypeScript 编译器对代码实体的语义理解（如变量、函数、类）。
	// 与文本匹配不同，符号分析能够准确地区分同名但不同作用域的实体，
	// 是实现精确的代码重构、导航和分析的基础。
	// =============================================================================

	// 1. 初始化项目
	realProjectPath, err := filepath.Abs("../demo-react-app")
	if err != nil {
		log.Fatalf("无法解析项目路径: %v", err)
	}

	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:         realProjectPath,
		TargetExtensions: []string{".ts", ".tsx"},
		IgnorePatterns:   []string{"node_modules", "dist", ".git", "build"},
		UseTsConfig:      true,
	})
	defer project.Close()

	fmt.Printf("✅ 成功加载项目: %s\n", realProjectPath)

	// 2. 示例 1: 识别并分析文件中的符号
	fmt.Println("\n" + strings.Repeat("-", 20) + " 示例 1: 识别符号 " + strings.Repeat("-", 20))
	analyzeSymbolsInFile(project, realProjectPath)

	// 3. 示例 2: 基于符号进行重命名安全性分析
	fmt.Println("\n" + strings.Repeat("-", 20) + " 示例 2: 重命名安全性分析 " + strings.Repeat("-", 20))
	performRenameSafetyAnalysis(project, realProjectPath)
}

// analyzeSymbolsInFile 演示如何在一个文件中查找并分析符号
func analyzeSymbolsInFile(project *tsmorphgo.Project, basePath string) {
	// 我们分析 `src/types.ts` 文件中的符号
	typesFilePath := filepath.Join(basePath, "src/types.ts")
	typesFile := project.GetSourceFile(typesFilePath)
	if typesFile == nil {
		log.Printf("警告: 未找到 types.ts 文件，跳过符号识别示例。\n")
		return
	}

	fmt.Printf("📋 分析文件: %s\n", filepath.Base(typesFilePath))

	// 使用 map 来存储唯一的符号
	symbolMap := make(map[string]*tsmorphgo.Symbol)

	typesFile.ForEachDescendant(func(node tsmorphgo.Node) {
		// 尝试从每个标识符节点获取符号
		if node.IsIdentifierNode() {
			symbol, err := tsmorphgo.GetSymbol(node)
			if err == nil && symbol != nil {
				// 使用第一个声明的位置作为唯一键
				declarations := symbol.GetDeclarations()
				if len(declarations) > 0 {
					firstDecl := declarations[0]
					key := fmt.Sprintf("%s:%d", firstDecl.GetSourceFile().GetFilePath(), firstDecl.GetStart())
					if _, exists := symbolMap[key]; !exists {
						symbolMap[key] = symbol
					}
				}
			}
		}
	})

	fmt.Printf("📊 在该文件中找到 %d 个唯一符号。\n", len(symbolMap))
	fmt.Println("🔍 部分符号列表:")

	i := 0
	for _, symbol := range symbolMap {
		if i >= 5 { // 只显示前5个
			break
		}
		// 获取符号的声明位置
	
declarations := symbol.GetDeclarations()
		var declInfo string
		if len(declarations) > 0 {
			decl := declarations[0]
			declInfo = fmt.Sprintf("(声明于 %s:%d)",
				filepath.Base(decl.GetSourceFile().GetFilePath()),
				decl.GetStartLineNumber())
		}

		fmt.Printf("  - 符号: '%s' %s\n", symbol.GetName(), declInfo)
		i++
	}
}

// performRenameSafetyAnalysis 演示如何使用符号来评估重命名的影响范围
func performRenameSafetyAnalysis(project *tsmorphgo.Project, basePath string) {
	// 我们将分析 `src/App.tsx` 中的 `users` 状态变量
	appFilePath := filepath.Join(basePath, "src/App.tsx")
	appFile := project.GetSourceFile(appFilePath)
	if appFile == nil {
		log.Printf("警告: 未找到 App.tsx 文件，跳过重命名分析示例。\n")
		return
	}

	var targetSymbol *tsmorphgo.Symbol
	var targetIdentifier tsmorphgo.Node

	// 找到 `const [users, setUsers] = useState<User[]>([]);` 中的 `users`
	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if targetSymbol != nil {
			return
		}
		if node.IsIdentifierNode() && strings.TrimSpace(node.GetText()) == "users" {
			parent := node.GetParent()
			if parent != nil {
				fmt.Printf("DEBUG: Found 'users' identifier. Parent Kind: %s\n", parent.GetKind().String())
				// 确认其祖先节点是变量声明的一部分
				if ancestor, ok := node.GetFirstAncestorByKind(tsmorphgo.KindVariableDeclaration); ok && ancestor != nil {
					symbol, err := tsmorphgo.GetSymbol(node)
					if err == nil && symbol != nil {
						targetSymbol = symbol
						targetIdentifier = node
					}
				}
			}
		}
	})

	if targetSymbol == nil {
		log.Printf("警告: 在 App.tsx 中未找到 'users' 状态变量的符号。\n")
		return
	}

	fmt.Printf("🎯 分析目标: '%s' 变量 (声明于 %s:%d)\n",
		targetSymbol.GetName(),
		filepath.Base(targetIdentifier.GetSourceFile().GetFilePath()),
		targetIdentifier.GetStartLineNumber())

	// 使用 FindReferences 找到所有语义相关的引用
	refs, _, err := tsmorphgo.FindReferencesWithCache(targetIdentifier)
	if err != nil {
		log.Printf("查找 '%s' 的引用失败: %v\n", targetSymbol.GetName(), err)
		return
	}

	refCount := len(refs)
	filesAffected := make(map[string]int)
	for _, ref := range refs {
		path := ref.GetSourceFile().GetFilePath()
		filesAffected[path]++
	}

	fmt.Println("\n📊 重命名影响分析:")
	fmt.Printf("  - 总引用数: %d (这才是准确的引用数)\n", refCount)
	fmt.Printf("  - 影响文件数: %d\n", len(filesAffected))

	fmt.Println("  - 文件引用分布:")
	for path, count := range filesAffected {
		relPath, _ := filepath.Rel(basePath, path)
		fmt.Printf("    - %s: %d 个引用\n", relPath, count)
	}

	// 基于引用数量给出重构建议
	fmt.Println("\n🛡️ 安全性评估:")
	switch {
	case refCount > 10:
		fmt.Println("  - 🔴 高风险: 引用分布广泛，重命名需谨慎，务必进行全量回归测试。\n")
	case refCount > 5:
		fmt.Println("  - 🟡 中风险: 建议在IDE中执行重构，并测试相关功能。\n")
	default:
		fmt.Println("  - 🟢 低风险: 影响范围可控，可以安全地进行重命名。\n")
	}
}
