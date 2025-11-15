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
	fmt.Println("🚀 TSMorphGo Symbol 验证 - 改进版")
	fmt.Println("==============================")

	// --- 项目设置 ---
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("❌ 获取工作目录失败: %v", err)
	}
	demoAppPath := filepath.Join(workDir, "demo-react-app")

	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:    demoAppPath,
		UseTsConfig: true,
	})
	if project == nil {
		log.Fatal("❌ 项目创建失败")
	}
	defer project.Close()

	fmt.Println("✅ 项目初始化成功")

	// --- 执行验证 ---
	verifyDifferentScopeSameName(project, demoAppPath)
	verifySameScopeMultipleReferences(project, demoAppPath)
	verifyClassMemberSymbols(project, demoAppPath)
	verifyCrossFileSymbol(project, demoAppPath)
}

// 验证1: 不同作用域的同名变量
func verifyDifferentScopeSameName(project *tsmorphgo.Project, demoAppPath string) {
	fmt.Println("\n🔍 验证1: 不同作用域的同名变量")
	fmt.Println("--------------------------------")
	fmt.Println("目标: 比较 test-symbol.ts 中 outerFunction 和 innerFunction 内的 'counter' 变量")

	file := project.GetSourceFile(filepath.Join(demoAppPath, "src/test-symbol.ts"))
	if file == nil {
		log.Fatal("❌ 未找到 test-symbol.ts 文件")
	}

	var outerCounter, innerCounter tsmorphgo.Node

	file.ForEachDescendant(func(node tsmorphgo.Node) {
		// outerFunction 内的 counter (根据调试，实际在第 16 行)
		if node.GetStartLineNumber() == 16 && node.IsIdentifier() && node.GetText() == "counter" {
			outerCounter = node
		}
		// innerFunction 内的 counter (根据调试，实际在第 23 行)
		if node.GetStartLineNumber() == 23 && node.IsIdentifier() && node.GetText() == "counter" {
			innerCounter = node
		}
	})

	if !outerCounter.IsValid() || !innerCounter.IsValid() {
		log.Fatal("❌ 未能定位到所有 'counter' 节点")
	}

	fmt.Println("✅ 已定位到两个 'counter' 节点")

	outerSymbol, _ := tsmorphgo.GetSymbol(outerCounter)
	innerSymbol, _ := tsmorphgo.GetSymbol(innerCounter)

	if outerSymbol == nil || innerSymbol == nil {
		log.Fatal("❌ 获取 Symbol 失败")
	}

	fmt.Printf("   - Outer Symbol: %s\n", outerSymbol.String())
	fmt.Printf("   - Inner Symbol: %s\n", innerSymbol.String())

	// 使用改进的比较方法
	fmt.Printf("   - Outer Symbol ID: %d\n", outerSymbol.GetId())
	fmt.Printf("   - Inner Symbol ID: %d\n", innerSymbol.GetId())
	fmt.Printf("   - Symbol Equals: %t\n", outerSymbol.Equals(innerSymbol))

	if !outerSymbol.Equals(innerSymbol) {
		fmt.Println("✅ 验证成功: 不同作用域的同名变量具有不同的 Symbol。")
	} else {
		fmt.Println("❌ 验证失败: 不同作用域的同名变量 Symbol 相同。")
	}
}

// 验证2: 同一作用域下的多次引用
func verifySameScopeMultipleReferences(project *tsmorphgo.Project, demoAppPath string) {
	fmt.Println("\n🔍 验证2: 同一作用域下的多次引用")
	fmt.Println("--------------------------------")
	fmt.Println("目标: 比较 test-symbol.ts 中 'sharedVar' 的声明和第一次使用")

	file := project.GetSourceFile(filepath.Join(demoAppPath, "src/test-symbol.ts"))
	if file == nil {
		log.Fatal("❌ 未找到 test-symbol.ts 文件")
	}

	var declaration, firstUse tsmorphgo.Node

	file.ForEachDescendant(func(node tsmorphgo.Node) {
		// sharedVar 的声明 (第 70 行)
		if node.GetStartLineNumber() == 70 && node.IsIdentifier() && node.GetText() == "sharedVar" {
			declaration = node
		}
		// sharedVar 的第一次使用 (第 73 行)
		if node.GetStartLineNumber() == 73 && node.IsIdentifier() && node.GetText() == "sharedVar" {
			firstUse = node
		}
	})

	if !declaration.IsValid() || !firstUse.IsValid() {
		log.Fatal("❌ 未能定位到 'sharedVar' 的声明和使用节点")
	}

	fmt.Println("✅ 已定位到 'sharedVar' 的声明和使用节点")

	declarationSymbol, _ := tsmorphgo.GetSymbol(declaration)
	useSymbol, _ := tsmorphgo.GetSymbol(firstUse)

	if declarationSymbol == nil || useSymbol == nil {
		log.Fatal("❌ 获取 Symbol 失败")
	}

	fmt.Printf("   - Declaration Symbol: %s\n", declarationSymbol.String())
	fmt.Printf("   - First Use Symbol:   %s\n", useSymbol.String())

	// 使用改进的比较方法
	fmt.Printf("   - Declaration Symbol ID: %d\n", declarationSymbol.GetId())
	fmt.Printf("   - First Use Symbol ID:   %d\n", useSymbol.GetId())
	fmt.Printf("   - Symbol Equals: %t\n", declarationSymbol.Equals(useSymbol))

	if declarationSymbol.Equals(useSymbol) {
		fmt.Println("✅ 验证成功: 同一变量的声明和使用具有相同的 Symbol。")
	} else {
		fmt.Println("❌ 验证失败: 同一变量的声明和使用 Symbol 不同。")
	}
}

// 验证3: 类成员的Symbol比较
func verifyClassMemberSymbols(project *tsmorphgo.Project, demoAppPath string) {
	fmt.Println("\n🔍 验证3: 类成员的Symbol比较")
	fmt.Println("----------------------------")
	fmt.Println("目标: 验证 SymbolTest 类中同名标识符在不同上下文中的Symbol")

	file := project.GetSourceFile(filepath.Join(demoAppPath, "src/test-symbol.ts"))
	if file == nil {
		log.Fatal("❌ 未找到 test-symbol.ts 文件")
	}

	var classProperty, localVariable, thisUsage tsmorphgo.Node

	file.ForEachDescendant(func(node tsmorphgo.Node) {
		// SymbolTest 类的 counter 属性声明 (第 42 行)
		if node.GetStartLineNumber() == 42 && node.IsIdentifier() && node.GetText() == "counter" {
			classProperty = node
		}
		// method 方法内的局部 counter 变量 (第 50 行)
		if node.GetStartLineNumber() == 50 && node.IsIdentifier() && node.GetText() == "counter" {
			localVariable = node
		}
		// console.log(this.counter) 中的 counter (第 54 行)
		if node.GetStartLineNumber() == 54 && node.GetText() == "counter" {
			thisUsage = node
		}
	})

	if !classProperty.IsValid() || !localVariable.IsValid() || !thisUsage.IsValid() {
		log.Fatal("❌ 未能定位到所有类成员节点")
	}

	fmt.Println("✅ 已定位到所有类成员节点")

	classPropertySymbol, _ := tsmorphgo.GetSymbol(classProperty)
	localVariableSymbol, _ := tsmorphgo.GetSymbol(localVariable)
	thisUsageSymbol, _ := tsmorphgo.GetSymbol(thisUsage)

	if classPropertySymbol == nil || localVariableSymbol == nil || thisUsageSymbol == nil {
		log.Fatal("❌ 获取 Symbol 失败")
	}

	fmt.Printf("   - Class Property Symbol: %s\n", classPropertySymbol.String())
	fmt.Printf("   - Local Variable Symbol: %s\n", localVariableSymbol.String())
	fmt.Printf("   - This Usage Symbol:     %s\n", thisUsageSymbol.String())

	// 使用改进的比较方法
	fmt.Printf("   - Class Property ID: %d\n", classPropertySymbol.GetId())
	fmt.Printf("   - Local Variable ID: %d\n", localVariableSymbol.GetId())
	fmt.Printf("   - This Usage ID:     %d\n", thisUsageSymbol.GetId())

	// 验证类属性和局部变量有不同的Symbol
	propertyDifferentFromLocal := classPropertySymbol.GetId() != localVariableSymbol.GetId()
	fmt.Printf("   - Class Property != Local Variable: %t\n", propertyDifferentFromLocal)

	// 验证this引用正确指向类属性（这是正确的TypeScript行为）
	thisPointsToClass := thisUsageSymbol.Equals(classPropertySymbol)
	fmt.Printf("   - This Usage Equals Class Property: %t\n", thisPointsToClass)

	if propertyDifferentFromLocal && thisPointsToClass {
		fmt.Println("✅ 验证成功: 类属性与局部变量Symbol不同，this引用正确指向类属性。")
	} else {
		fmt.Println("❌ 验证失败: 类成员Symbol比较出现问题。")
		if !propertyDifferentFromLocal {
			fmt.Println("   - 问题: 类属性与局部变量Symbol相同")
		}
		if !thisPointsToClass {
			fmt.Println("   - 问题: this引用未正确指向类属性")
		}
	}

	// 额外说明
	fmt.Println("\n💡 说明:")
	fmt.Println("   - this.counter 与类属性 counter 共享Symbol是正确的TypeScript行为")
	fmt.Println("   - 方法内局部变量 counter 有不同的Symbol，避免了与类属性冲突")
	fmt.Println("   - 这证明了TypeScript作用域系统正在正确工作")
}

// 验证4: 跨文件 Symbol 比较
func verifyCrossFileSymbol(project *tsmorphgo.Project, demoAppPath string) {
	fmt.Println("\n🔍 验证4: 跨文件 Symbol 比较")
	fmt.Println("--------------------------------")
	fmt.Println("目标: 比较 App.tsx 中导入的 'formatDate' 和其在 utils/dateUtils.ts 中的原始定义")

	appFile := project.GetSourceFile(filepath.Join(demoAppPath, "src/components/App.tsx"))
	if appFile == nil {
		log.Fatal("❌ 未找到 App.tsx 文件")
	}
	utilsFile := project.GetSourceFile(filepath.Join(demoAppPath, "src/utils/dateUtils.ts"))
	if utilsFile == nil {
		log.Fatal("❌ 未找到 utils/dateUtils.ts 文件")
	}

	var importNode, exportNode tsmorphgo.Node

	// 在 App.tsx 中找到导入的 formatDate (第 5 行)
	appFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.GetStartLineNumber() == 5 && node.IsIdentifier() && node.GetText() == "formatDate" {
			importNode = node
		}
	})

	// 在 dateUtils.ts 中找到导出的 formatDate (第 5 行)
	utilsFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if node.GetStartLineNumber() == 5 && node.IsIdentifier() && node.GetText() == "formatDate" {
			exportNode = node
		}
	})

	if !importNode.IsValid() || !exportNode.IsValid() {
		log.Fatal("❌ 未能定位到 'formatDate' 的导入和导出节点")
	}

	fmt.Println("✅ 已定位到 'formatDate' 的导入和导出节点")

	importSymbol, _ := tsmorphgo.GetSymbol(importNode)
	exportSymbol, _ := tsmorphgo.GetSymbol(exportNode)

	if importSymbol == nil || exportSymbol == nil {
		log.Fatal("❌ 获取 Symbol 失败")
	}

	fmt.Printf("   - Import Symbol: %s\n", importSymbol.String())
	fmt.Printf("   - Export Symbol: %s\n", exportSymbol.String())

	// 使用改进的比较方法
	fmt.Printf("   - Import Symbol ID: %d\n", importSymbol.GetId())
	fmt.Printf("   - Export Symbol ID: %d\n", exportSymbol.GetId())
	fmt.Printf("   - Symbol Equals: %t\n", importSymbol.Equals(exportSymbol))

	if importSymbol.Equals(exportSymbol) {
		fmt.Println("✅ 验证成功: 跨文件的导入和导出指向同一个 Symbol。")
	} else {
		fmt.Println("❌ 验证失败: 跨文件的导入和导出 Symbol 不同。")
		// 添加额外的调试信息
		fmt.Println("🔍 调试信息:")
		fmt.Printf("   - Import Symbol ID: %d\n", importSymbol.GetId())
		fmt.Printf("   - Export Symbol ID: %d\n", exportSymbol.GetId())
		fmt.Printf("   - Import Symbol Flags: %d\n", importSymbol.GetFlags())
		fmt.Printf("   - Export Symbol Flags: %d\n", exportSymbol.GetFlags())

		// 尝试使用TypeChecker直接获取符号进行比较
		fmt.Println("   - 尝试直接比较TypeChecker获取的符号...")
		fmt.Println("     这个功能需要进一步实现LSP服务的改进")

		// 说明跨文件Symbol比较的复杂性
		fmt.Println("\n💡 说明:")
		fmt.Println("   在TypeScript模块系统中，导入和导出可能有不同的Symbol实例")
		fmt.Println("   但通过TypeChecker.getSymbolIfSameReference可以确定它们指向同一个引用")
		fmt.Println("   这需要更深入的LSP服务集成来实现准确的跨文件Symbol比较")
	}
}
