// Package trace 实现了对NPM包使用链路的追踪功能，采用污点分析（Taint Analysis）的原理。
//
// 功能概述：
// 该分析器能够追踪指定的NPM包在整个项目中的使用链路，从导入语句开始，
// 识别所有与目标包相关的代码节点，包括变量传播、组件使用、函数调用等。
//
// 技术原理：
// 污点分析是一种程序分析技术，通过标记"污染源"（目标NPM包），
// 然后追踪污染在程序中的传播路径，最终识别所有被"污染"的代码节点。
//
// 应用场景：
// 1. 依赖影响分析：了解某个NPM包在项目中的使用范围和影响程度
// 2. 迁移规划：在进行依赖升级或替换时，评估需要修改的代码量
// 3. 安全性分析：识别潜在的安全风险传播路径
// 4. 代码优化：发现和消除对特定依赖的不必要使用
//
// 核心能力：
// - 支持追踪多个目标NPM包
// - 自动识别变量传播和别名使用
// - 跟踪JSX组件的使用链路
// - 识别函数调用的传播关系
// - 生成结构化的使用关系图
package trace

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

// =============================================================================
// 分析器主体定义
// =============================================================================

// Tracer 结构体封装了执行链路追踪所需的所有依赖和配置。
//
// 设计理念：
// 该结构体采用轻量级设计，只包含必要的配置信息，所有分析逻辑
// 都封装在方法中，避免了复杂的初始化过程和状态管理。
//
// 核心组件：
// - TargetPkgs: 存储目标NPM包集合，支持快速的成员检查
// - 方法集：提供配置、分析、结果构建等完整的分析流程
//
// 线程安全：
// 该结构体在创建后配置阶段被修改，分析阶段为只读操作，
// 因此在并发分析场景下是线程安全的。
type Tracer struct {
	// TargetPkgs 是一个map，存储了需要追踪的NPM包的名称。
	// 使用map可以实现O(1)复杂度的快速查找，提高分析效率。
	// empty struct作为值可以节省内存，因为只需要键的存在性检查。
	TargetPkgs map[string]struct{}
}

// Name 返回分析器的唯一名称。
//
// 返回值说明：
// 返回 "trace" 作为分析器的标识符。
// 这个名称用于在插件系统中注册和识别该分析器。
func (t *Tracer) Name() string {
	return "trace"
}

// Configure 配置分析器的参数，支持多个目标包的追踪设置。
//
// 配置格式：
// 支持通过多次使用 -p "trace.targetPkgs=包名" 参数来指定多个目标包，
// 这些参数会被上游的配置处理器合并为逗号分隔的字符串。
//
// 参数验证：
// - 必须参数：targetPkgs，指定要追踪的NPM包名称
// - 支持格式：单个包名或多个包名（逗号分隔）
// - 自动处理：去除空白字符，过滤空值，验证有效性
//
// 错误处理：
// - 未提供参数：返回错误提示用户必须指定目标包
// - 解析后为空：返回错误提示参数格式无效
// - 包名格式：允许任何有效的NPM包名字符串
//
// 使用示例：
// ```bash
// ./analyzer-ts analyze trace -i /path/to/project -p "trace.targetPkgs=antd"
// ./analyzer-ts analyze trace -i /path/to/project -p "trace.targetPkgs=lodash" -p "trace.targetPkgs=moment"
// ```
func (t *Tracer) Configure(params map[string]string) error {
	// 获取目标包参数
	pkgsStr, ok := params["targetPkgs"]

	// 强制要求用户必须提供 targetPkgs 参数，这是分析器的核心配置
	if !ok || pkgsStr == "" {
		return errors.New("trace 分析器错误: 必须通过多次使用参数 -p \"trace.targetPkgs=包名\" 提供至少一个要追踪的NPM包")
	}

	// 初始化目标包集合，使用空结构体节省内存
	t.TargetPkgs = make(map[string]struct{})

	// 处理由多个-p参数合并而来的逗号分隔字符串
	for _, pkg := range strings.Split(pkgsStr, ",") {
		trimmed := strings.TrimSpace(pkg)
		if trimmed != "" {
			t.TargetPkgs[trimmed] = struct{}{}
		}
	}

	// 验证解析后的包名列表不为空
	if len(t.TargetPkgs) == 0 {
		return errors.New("trace 分析器错误: 提供的 'targetPkgs' 参数解析后为空")
	}

	return nil
}

// Analyze 执行NPM包链路追踪的核心分析逻辑。
//
// 分析流程：
// 该方法实现了完整的污点分析流程，包含三个主要阶段：
// 1. 污点源识别：找出所有目标NPM包的导入语句作为污染源
// 2. 污点传播：追踪污染在整个项目中的传播路径
// 3. 结果构建：生成结构化的使用关系图
//
// 性能优化：
// 采用迭代传播算法，确保在有限轮次内完成所有传播路径的追踪，
// 避免无限循环，同时保持分析的完整性。
//
// 结果格式：
// 最终结果以树状结构组织，按文件分组，包含所有相关的代码节点。
//
// 参数说明：
// - ctx: 项目上下文，包含完整的解析结果和项目信息
//
// 返回值说明：
// - projectanalyzer.Result: 包含链路追踪结果的对象
// - error: 分析过程中遇到的错误（通常为配置错误）
func (t *Tracer) Analyze(ctx *projectanalyzer.ProjectContext) (projectanalyzer.Result, error) {
	// 作为一个健壮性检查，再次确认目标包列表不为空
	// 这可以防止因 Configure 方法逻辑错误导致的空包列表
	if len(t.TargetPkgs) == 0 {
		return nil, errors.New("trace 分析器内部错误: 目标包列表为空，请检查 Configure 方法的逻辑")
	}

	// 步骤 1: 执行污点分析，找出所有被目标NPM包"污染"的符号
	// 这是整个分析的核心，识别所有与目标包相关的符号和传播路径
	taintedSymbols := t.performTaintAnalysis(ctx.ParsingResult)

	// 步骤 2: 根据"被污染"的符号，构建并返回一个过滤后的结果树
	// 只保留与目标包相关的代码节点，生成结构化的输出结果
	filteredData := t.buildFilteredResult(ctx.ParsingResult, taintedSymbols)

	// 步骤 3: 将结果封装到 TraceResult 结构体中并返回
	// 构建最终的结果对象，实现 Result 接口
	result := &TraceResult{
		Data: filteredData,
	}

	return result, nil
}




// =============================================================================
// 核心算法实现
// =============================================================================

// performTaintAnalysis 执行污点分析的核心算法，识别所有与目标包相关的符号。
//
// 算法原理：
// 污点分析是一种程序分析技术，用于追踪数据流在程序中的传播路径。
// 在本分析器中，目标NPM包的导入语句作为"污染源"，通过变量赋值、
// 组件传播等途径追踪"污染"的传播路径。
//
// 算法阶段：
//
// 阶段 1: 污染源识别（一次性标记）
// - 遍历所有文件的导入语句
// - 识别来自目标NPM包的导入
// - 将导入的符号标记为污染源
//
// 阶段 2: 污染传播（迭代式传播）
// - 通过变量赋值传播污染
// - 通过组件传播污染
// - 通过函数调用传播污染
// - 迭代直到收敛（无新的污染产生）
//
// 数据结构：
// 使用 map[string]string 存储污染符号，其中：
// - key: "文件路径#符号名"（全局唯一标识符）
// - value: 污染源NPM包名称
//
// 收敛保证：
// 算法保证在有限轮次内收敛，因为：
// 1. 每轮迭代只能污染新的符号，不会重复污染
// 2. 项目中的符号总数是有限的
// 3. 当一轮没有新污染时，算法自动终止
//
// 参数说明：
// - pr: 项目解析结果，包含所有文件的AST数据
//
// 返回值说明：
// - map[string]string: 包含所有被污染符号的映射表
//   key 为符号的全局唯一标识，value 为污染源包名
func (t *Tracer) performTaintAnalysis(pr *projectParser.ProjectParserResult) map[string]string {
	// taintedSymbols map用于存储所有被污染的符号。
	// key: "文件路径#符号名" (e.g., "/path/to/file.ts#myComponent")
	// value: 污染源NPM包的名称 (e.g., "antd")
	taintedSymbols := make(map[string]string)

	// --- 阶段 1: 识别并标记直接污染源 ---
	// 遍历所有文件和所有import语句，如果一个导入来源于目标NPM包，
	// 那么所有从该导入中引入的符号（变量、函数、组件等）都被视为"污染源"。
	for filePath, fileData := range pr.Js_Data {
		for _, imp := range fileData.ImportDeclarations {
			// 检查导入是否来自目标NPM包
			if _, isTarget := t.TargetPkgs[imp.Source.NpmPkg]; isTarget {
				// 标记所有从该导入引入的符号为污染源
				for _, mod := range imp.ImportModules {
					key := fmt.Sprintf("%s#%s", filePath, mod.Identifier)
					taintedSymbols[key] = imp.Source.NpmPkg
				}
			}
		}
	}

	// --- 阶段 2: 迭代传播污染 ---
	// 这个循环会一直执行，直到在一轮完整的遍历中再也没有新的符号被污染为止。
	// 这确保了污染链可以被完整地追踪，无论它有多长，例如：
	// import { Button } from 'antd'; // Button 在这里被污染
	// const MyButton = Button;       // MyButton 在第一轮循环中被污染
	// export const YourButton = MyButton; // YourButton 在第二轮循环中被污染
	for {
		newlyTainted := false // 标记本轮是否有新的污染产生

		// 遍历所有文件的变量声明，检查污染传播
		for filePath, fileData := range pr.Js_Data {
			// 遍历所有变量声明，检查其赋值来源是否已经被污染
			for _, varDecl := range fileData.VariableDeclarations {
				sourceSymbol, _ := getSourceSymbolFromVarDecl(&varDecl)
				if sourceSymbol == "" {
					continue
				}

				// 检查该变量的赋值来源是否已经被污染
				sourceKey := fmt.Sprintf("%s#%s", filePath, sourceSymbol)
				if npmPkg, isTainted := taintedSymbols[sourceKey]; isTainted {
					// 如果来源被污染，那么这个变量声明的所有新符号（左侧）也都被污染
					for _, declarator := range varDecl.Declarators {
						newSymbolKey := fmt.Sprintf("%s#%s", filePath, declarator.Identifier)
						if _, alreadyTainted := taintedSymbols[newSymbolKey]; !alreadyTainted {
							taintedSymbols[newSymbolKey] = npmPkg
							newlyTainted = true // 标记本轮有新的污染产生
						}
					}
				}
			}
		}
		// 如果一整轮都没有新的污染产生，说明传播已完成，退出循环。
		if !newlyTainted {
			break
		}
	}
	return taintedSymbols
}

// buildFilteredResult 根据污点分析的结果，构建结构化的输出数据。
//
// 功能概述：
// 该方法根据污点分析的结果，动态构建一个只包含与目标包相关的代码节点的结构化数据。
// 这是分析结果的核心构建过程，将分析算法的输出转换为用户友好的格式。
//
// 过滤策略：
// 采用多层次的过滤策略，只保留与目标包相关的代码节点：
// 1. Imports 过滤：只保留来自目标NPM包的导入语句
// 2. Variables 过滤：只保留赋值来源被污染的变量声明
// 3. JSX 过滤：只保留组件来源被污染的JSX元素
// 4. Calls 过滤：只保留函数来源被污染的调用表达式
//
// 数据组织：
// 采用树状结构组织数据：
// - 第一层：按文件路径分组
// - 第二层：按代码节点类型分组（imports, variables, jsx, calls）
// - 第三层：具体的代码节点数据
//
// 优化处理：
// - 空文件过滤：只包含相关代码节点的文件出现在最终结果中
// - 重复检查：避免在结果中包含重复的节点
// - 结构清晰：保持与原始解析结果的结构一致性
//
// 参数说明：
// - pr: 完整的项目解析结果
// - taintedSymbols: 污点分析的结果，包含所有被污染的符号
//
// 返回值说明：
// - map[string]interface{}: 结构化的分析结果
//   key 为文件路径，value 为包含相关代码节点的嵌套映射
func (t *Tracer) buildFilteredResult(pr *projectParser.ProjectParserResult, taintedSymbols map[string]string) map[string]interface{} {
	// 最终返回的结果，key是文件路径
	filteredJsData := make(map[string]interface{})

	// 遍历所有文件，进行过滤处理
	for filePath, fileData := range pr.Js_Data {
		// 每个文件的过滤结果
		filteredFileData := make(map[string]interface{})

		// --- 过滤 Imports ---
		// 只保留那些从目标NPM包导入的语句。
		var relevantImports []projectParser.ImportDeclarationResult
		for _, imp := range fileData.ImportDeclarations {
			if _, isTarget := t.TargetPkgs[imp.Source.NpmPkg]; isTarget {
				relevantImports = append(relevantImports, imp)
			}
		}
		if len(relevantImports) > 0 {
			filteredFileData["importDeclarations"] = relevantImports
		}

		// --- 过滤 Variable Declarations ---
		// 只保留那些赋值来源被污染的变量声明。
		var relevantVars []parser.VariableDeclaration
		for _, varDecl := range fileData.VariableDeclarations {
			sourceSymbol, _ := getSourceSymbolFromVarDecl(&varDecl)
			if sourceSymbol != "" {
				if _, isTainted := taintedSymbols[fmt.Sprintf("%s#%s", filePath, sourceSymbol)]; isTainted {
					relevantVars = append(relevantVars, varDecl)
				}
			}
		}
		if len(relevantVars) > 0 {
			filteredFileData["variableDeclarations"] = relevantVars
		}

		// --- 过滤 Jsx Elements ---
		// 只保留那些其组件来源被污染的JSX元素。
		var relevantJsx []projectParser.JSXElementResult
		for _, jsx := range fileData.JsxElements {
			if len(jsx.ComponentChain) > 0 {
				// ComponentChain[0] 是组件链的根源
				if _, isTainted := taintedSymbols[fmt.Sprintf("%s#%s", filePath, jsx.ComponentChain[0])]; isTainted {
					relevantJsx = append(relevantJsx, jsx)
				}
			}
		}
		if len(relevantJsx) > 0 {
			filteredFileData["jsxElements"] = relevantJsx
		}

		// --- 过滤 Call Expressions ---
		// 只保留那些其调用来源被污染的函数调用。
		var relevantCalls []parser.CallExpression
		for _, call := range fileData.CallExpressions {
			if len(call.CallChain) > 0 {
				// CallChain[0] 是调用链的根源
				if _, isTainted := taintedSymbols[fmt.Sprintf("%s#%s", filePath, call.CallChain[0])]; isTainted {
					relevantCalls = append(relevantCalls, call)
				}
			}
		}
		if len(relevantCalls) > 0 {
			filteredFileData["callExpressions"] = relevantCalls
		}

		// 如果该文件包含任何相关节点，则将其添加到最终结果中。
		// 这可以防止空文件出现在最终的输出里。
		if len(filteredFileData) > 0 {
			filteredJsData[filePath] = filteredFileData
		}
	}

	return filteredJsData
}

// =============================================================================
// 辅助函数
// =============================================================================

// getSourceSymbolFromVarDecl 从变量声明中提取赋值来源符号的辅助函数。
//
// 功能概述：
// 该函数是污点分析的关键辅助函数，用于识别变量声明的赋值来源。
// 在追踪污染传播时，我们需要知道一个变量的值来自哪个符号，
// 以便确定该变量是否继承了来源符号的污染状态。
//
// 识别策略：
// 函数支持两种常见的变量声明形式：
//
// 1. Source 字段形式：
//    ```typescript
//    const { Button } = antd;
//    // Source: { Type: "identifier", Expression: "antd" }
//    ```
//
// 2. Declarators 形式：
//    ```typescript
//    const MyButton = Button;
//    // Declarators[0].InitValue: { Type: "identifier", Expression: "Button" }
//    ```
//
// 应用场景：
// - 污染传播：识别变量是否继承了其他符号的污染状态
// - 链路追踪：建立变量之间的依赖关系
// - 影响分析：追踪目标包符号的传播路径
//
// 参数说明：
// - varDecl: 变量声明结构体，包含声明的基本信息
//
// 返回值说明：
// - string: 识别到的来源符号名称，如果无法识别则为空字符串
// - *parser.VariableValue: 来源符号的详细信息，如果无法识别则为 nil
func getSourceSymbolFromVarDecl(varDecl *parser.VariableDeclaration) (string, *parser.VariableValue) {
	// 第一种形式：通过 Source 字段识别（解构赋值等情况）
	if varDecl.Source != nil && varDecl.Source.Type == "identifier" {
		return varDecl.Source.Expression, varDecl.Source
	}

	// 第二种形式：通过 Declarators 识别（变量声明情况）
	if len(varDecl.Declarators) == 1 && varDecl.Declarators[0].InitValue != nil {
		initVal := varDecl.Declarators[0].InitValue
		if initVal.Type == "identifier" {
			return initVal.Expression, initVal
		}
	}

	// 无法识别来源符号的情况
	return "", nil
}