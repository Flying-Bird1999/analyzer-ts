// package trace 实现了对NPM包使用链路的追踪功能，采用污点分析的原理。
package trace

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	projectanalyzer "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer"
)

// --- Analyzer ---

// Tracer 结构体封装了执行链路追踪所需的所有依赖和配置。
type Tracer struct {
	// TargetPkgs 是一个map，存储了需要追踪的NPM包的名称。
	// 使用map可以实现O(1)复杂度的快速查找。
	TargetPkgs map[string]struct{}
}

// Name 返回分析器的唯一名称。
func (t *Tracer) Name() string {
	return "trace"
}

// Configure 从一个map中解析和设置分析器所需的参数。
// 对于Tracer来说，它强制要求提供 `targetPkgs` 参数。
// 该分析器通过多次使用 -p "trace.targetPkgs=pkg" 的方式来接收多个包，
// 在此方法中，这些值会被上游的 `configureAnalyzers` 函数用逗号连接成一个单一的字符串。
func (t *Tracer) Configure(params map[string]string) error {
	pkgsStr, ok := params["targetPkgs"]
	// 强制要求用户必须提供 targetPkgs 参数。
	if !ok || pkgsStr == "" {
		return errors.New("trace 分析器错误: 必须通过多次使用参数 -p \"trace.targetPkgs=包名\" 提供至少一个要追踪的NPM包")
	}

	// 初始化map，准备接收解析后的包名
	t.TargetPkgs = make(map[string]struct{})
	// 用逗号分割由多个-p参数合并而来的字符串
	for _, pkg := range strings.Split(pkgsStr, ",") {
		trimmed := strings.TrimSpace(pkg)
		if trimmed != "" {
			t.TargetPkgs[trimmed] = struct{}{}
		}
	}

	// 如果解析后列表为空（例如，用户只传入了逗号或空格），同样报错。
	if len(t.TargetPkgs) == 0 {
		return errors.New("trace 分析器错误: 提供的 'targetPkgs' 参数解析后为空")
	}

	return nil
}

// Analyze 是实现 Analyzer 接口的主方法。
// 它接收项目上下文，执行追踪分析，并返回结果。
func (t *Tracer) Analyze(ctx *projectanalyzer.ProjectContext) (projectanalyzer.Result, error) {
	// 作为一个健壮性检查，再次确认目标包列表不为空。
	if len(t.TargetPkgs) == 0 {
		return nil, errors.New("trace 分析器内部错误: 目标包列表为空，请检查 Configure 方法的逻辑")
	}

	// 步骤 1: 执行污点分析，找出所有被目标NPM包“污染”的符号。
	taintedSymbols := t.performTaintAnalysis(ctx.ParsingResult)

	// 步骤 2: 根据“被污染”的符号，构建并返回一个过滤后的结果树。
	filteredData := t.buildFilteredResult(ctx.ParsingResult, taintedSymbols)

	// 步骤 3: 将结果封装到 TraceResult 结构体中并返回。
	result := &TraceResult{
		Data: filteredData,
	}

	return result, nil
}

// --- Result ---

// TraceResult 封装了链路追踪的分析结果，并实现了 Result 接口。
type TraceResult struct {
	// Data 存储了最终过滤后的分析数据，其结构是一个以文件路径为键，
	// 以包含该文件内相关代码节点（如imports, jsx等）的map为值的嵌套map。
	Data map[string]interface{}
}

// Name 返回结果的名称，与分析器名称一致。
func (r *TraceResult) Name() string {
	return "trace"
}

// Summary 返回对结果的简短描述。
func (r *TraceResult) Summary() string {
	return fmt.Sprintf("成功追踪到 %d 个文件中存在相关的使用链路。", len(r.Data))
}

// ToJSON 将结果序列化为 JSON 格式的字节数组。
func (r *TraceResult) ToJSON(indent bool) ([]byte, error) {
	// 直接将核心的Data字段进行序列化，而不是整个TraceResult结构体，
	// 以便输出更纯净的JSON结果。
	return projectanalyzer.ToJSONBytes(r.Data, indent)
}

// ToConsole 将结果转换为适合在控制台输出的字符串格式。
func (r *TraceResult) ToConsole() string {
	// 对于 trace 这种复杂的树状结果，直接输出格式化的JSON是最清晰的，所以我们复用 ToJSON。
	jsonData, err := r.ToJSON(true)
	if err != nil {
		return fmt.Sprintf("无法将结果序列化为JSON: %v", err)
	}
	return string(jsonData)
}


// --- Core Logic ---

// performTaintAnalysis 执行污点分析，找出所有与目标包相关的符号。
// 这是整个追踪功能的核心，包含污染源识别和污染传播两个阶段。
func (t *Tracer) performTaintAnalysis(pr *projectParser.ProjectParserResult) map[string]string {
	// taintedSymbols map用于存储所有被污染的符号。
	// key: "文件路径#符号名" (e.g., "/path/to/file.ts#myComponent")
	// value: 污染源NPM包的名称 (e.g., "antd")
	taintedSymbols := make(map[string]string)

	// --- 阶段 1: 识别并标记直接污染源 ---
	// 遍历所有文件和所有import语句，如果一个导入来源于目标NPM包，
	// 那么所有从该导入中引入的符号（变量、函数、组件等）都被视为“污染源”。
	for filePath, fileData := range pr.Js_Data {
		for _, imp := range fileData.ImportDeclarations {
			if _, isTarget := t.TargetPkgs[imp.Source.NpmPkg]; isTarget {
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
		newlyTainted := false
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

// buildFilteredResult 根据污点分析的结果，动态构建一个只包含相关节点的map，用于最终的JSON输出。
func (t *Tracer) buildFilteredResult(pr *projectParser.ProjectParserResult, taintedSymbols map[string]string) map[string]interface{} {
	// 最终返回的结果，key是文件路径
	filteredJsData := make(map[string]interface{})

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

// getSourceSymbolFromVarDecl 是一个辅助函数，用于从一个变量声明中提取其赋值的来源符号。
// 例如，对于 `const A = B`，它会返回 "B"。
func getSourceSymbolFromVarDecl(varDecl *parser.VariableDeclaration) (string, *parser.VariableValue) {
	if varDecl.Source != nil && varDecl.Source.Type == "identifier" {
		return varDecl.Source.Expression, varDecl.Source
	}
	if len(varDecl.Declarators) == 1 && varDecl.Declarators[0].InitValue != nil {
		initVal := varDecl.Declarators[0].InitValue
		if initVal.Type == "identifier" {
			return initVal.Expression, initVal
		}
	}
	return "", nil
}