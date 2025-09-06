// package parser 提供了对单个 TypeScript/TSX 文件进行 AST（抽象语法树）解析的功能。
// 本文件（functionDeclaration.go）专门负责处理和解析函数声明。
package parser

import (
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// ParameterResult 存储一个解析后的函数参数信息。
// 结构进行了扩展，以捕获更丰富的参数属性。
type ParameterResult struct {
	Name         string `json:"name"`                   // 参数名称
	Type         string `json:"type"`                   // 参数的类型文本
	Raw          string `json:"raw,omitempty"`                    // 参数在源码中的原始文本
	Optional     bool   `json:"optional"`               // 新增：标记此参数是否可选 (e.g., name?: string)
	DefaultValue string `json:"defaultValue,omitempty"` // 新增：存储参数的默认值 (e.g., port = 3000)
	IsRest       bool   `json:"isRest"`                 // 新增：标记此参数是否为 rest 参数 (e.g., ...args)
}

// FunctionDeclarationResult 存储一个完整的函数声明的解析结果。
// 结构进行了扩展，以支持泛型和更广泛的函数类型。
type FunctionDeclarationResult struct {
	Identifier     string            `json:"identifier"`     // 函数的名称。对于匿名函数或表达式，这通常是变量名。
	Exported       bool              `json:"exported"`       // 标记此函数是否被导出。
	IsAsync        bool              `json:"isAsync"`        // 标记此函数是否为异步函数 (async)。
	IsGenerator    bool              `json:"isGenerator"`    // 新增：标记此函数是否为生成器函数 (function*)。
	Generics       []string          `json:"generics,omitempty"`       // 新增：存储泛型参数列表 (e.g., ["T", "K"])。
	Parameters     []ParameterResult `json:"parameters"`     // 函数的参数列表。
	ReturnType     string            `json:"returnType,omitempty"`     // 函数的返回类型文本。
	Raw            string            `json:"raw,omitempty"`            // 节点在源码中的原始文本。
	SourceLocation *SourceLocation    `json:"sourceLocation,omitempty"` // 节点在源码中的位置信息。
}

// NewFunctionDeclarationResult 是基于 ast.FunctionDeclaration 节点创建函数解析结果的构造函数。
func NewFunctionDeclarationResult(node *ast.FunctionDeclaration, sourceCode string) *FunctionDeclarationResult {
	result := &FunctionDeclarationResult{
		Raw:            utils.GetNodeText(node.AsNode(), sourceCode),
		Parameters:     []ParameterResult{},
		Generics:       []string{},
		SourceLocation: NewSourceLocation(node.AsNode(), sourceCode),
	}

	if node.Name() != nil {
		result.Identifier = node.Name().Text()
	}

	extractFunctionDetails(result, node.AsNode(), sourceCode)

	return result
}

// NewFunctionDeclarationResultFromExpression 是一个更通用的构造函数，
// 它可以从函数表达式（如箭头函数、匿名函数）创建解析结果。
// identifier: 函数的标识符（通常是赋值的变量名）。
// isExported: 该函数是否被导出。
// node: 函数表达式的 AST 节点。
func NewFunctionDeclarationResultFromExpression(identifier string, isExported bool, node *ast.Node, sourceCode string) *FunctionDeclarationResult {
	result := &FunctionDeclarationResult{
		Identifier:     identifier,
		Exported:       isExported,
		Raw:            utils.GetNodeText(node, sourceCode),
		Parameters:     []ParameterResult{},
		Generics:       []string{},
		SourceLocation: NewSourceLocation(node, sourceCode),
	}

	// 调用核心解析逻辑函数
	extractFunctionDetails(result, node, sourceCode)

	return result
}

// extractFunctionDetails 是核心的函数信息提取逻辑。
// 它可以处理任何符合函数特征的 AST 节点 (FunctionDeclaration, ArrowFunction, FunctionExpression)。
// 它会修改传入的 `result` 指针，为其填充详细信息。
// 第二次修复：根据 AST 库的实际 API 调整字段和方法的调用方式。
func extractFunctionDetails(result *FunctionDeclarationResult, node *ast.Node, sourceCode string) {
	// 根据不同的函数节点类型，分别处理
	switch n := node.AsNode(); n.Kind {
	case ast.KindFunctionDeclaration:
		fnNode := n.AsFunctionDeclaration()
		// 1. 解析修饰符 (export, async)
		if fnNode.Modifiers() != nil {
			for _, modifier := range fnNode.Modifiers().Nodes {
				switch modifier.Kind {
				case ast.KindExportKeyword:
					result.Exported = true
				case ast.KindAsyncKeyword:
					result.IsAsync = true
				}
			}
		}
		// 2. 解析泛型参数
		if fnNode.TypeParameters != nil {
			for _, param := range fnNode.TypeParameters.Nodes {
				result.Generics = append(result.Generics, utils.GetNodeText(param, sourceCode))
			}
		}
		// 3. 解析函数参数
		parseParameters(result, fnNode.Parameters, sourceCode)
		// 4. 解析返回类型
		if fnNode.Type != nil {
			result.ReturnType = strings.TrimSpace(utils.GetNodeText(fnNode.Type, sourceCode))
		} else {
			result.ReturnType = ""
		}
		// 5. 检查是否为生成器函数
		if fnNode.AsteriskToken != nil {
			result.IsGenerator = true
		}

	case ast.KindArrowFunction:
		fnNode := n.AsArrowFunction()
		// 1. 解析修饰符 (async)
		if fnNode.Modifiers() != nil {
			for _, modifier := range fnNode.Modifiers().Nodes {
				if modifier.Kind == ast.KindAsyncKeyword {
					result.IsAsync = true
				}
			}
		}
		// 2. 解析泛型参数
		if fnNode.TypeParameters != nil {
			for _, param := range fnNode.TypeParameters.Nodes {
				result.Generics = append(result.Generics, utils.GetNodeText(param, sourceCode))
			}
		}
		// 3. 解析函数参数
		parseParameters(result, fnNode.Parameters, sourceCode)
		// 4. 解析返回类型
		if fnNode.Type != nil {
			result.ReturnType = strings.TrimSpace(utils.GetNodeText(fnNode.Type, sourceCode))
		} else {
			result.ReturnType = ""
		}

	case ast.KindFunctionExpression:
		fnNode := n.AsFunctionExpression()
		// 1. 解析修饰符 (async)
		if fnNode.Modifiers() != nil {
			for _, modifier := range fnNode.Modifiers().Nodes {
				if modifier.Kind == ast.KindAsyncKeyword {
					result.IsAsync = true
				}
			}
		}
		// 2. 解析泛型参数
		if fnNode.TypeParameters != nil {
			for _, param := range fnNode.TypeParameters.Nodes {
				result.Generics = append(result.Generics, utils.GetNodeText(param, sourceCode))
			}
		}
		// 3. 解析函数参数
		parseParameters(result, fnNode.Parameters, sourceCode)
		// 4. 解析返回类型
		if fnNode.Type != nil {
			result.ReturnType = strings.TrimSpace(utils.GetNodeText(fnNode.Type, sourceCode))
		} else {
			result.ReturnType = ""
		}
		// 5. 检查是否为生成器函数
		if fnNode.AsteriskToken != nil {
			result.IsGenerator = true
		}
	}
}

// parseParameters 是一个辅助函数，用于从参数列表中提取详细信息。
// 这个函数被 `extractFunctionDetails` 调用，以减少代码重复。
func parseParameters(result *FunctionDeclarationResult, params *ast.NodeList, sourceCode string) {
	if params == nil {
		return
	}
	for _, paramNode := range params.Nodes {
		param := paramNode.AsParameterDeclaration()
		paramName := ""
		nameNode := param.Name()

		// 获取参数名，需要处理解构等复杂情况
		if nameNode != nil {
			if nameNode.Kind == ast.KindObjectBindingPattern || nameNode.Kind == ast.KindArrayBindingPattern {
				paramName = utils.GetNodeText(nameNode, sourceCode)
			} else {
				paramName = nameNode.Text()
			}
		}

		// 获取参数类型
		paramType := ""
		if param.Type != nil {
			paramType = strings.TrimSpace(utils.GetNodeText(param.Type, sourceCode))
		}

		// 获取参数默认值
		defaultValue := ""
		if param.Initializer != nil {
			defaultValue = strings.TrimSpace(utils.GetNodeText(param.Initializer, sourceCode))
		}

		result.Parameters = append(result.Parameters, ParameterResult{
			Name:         paramName,
			Type:         paramType,
			Raw:          utils.GetNodeText(param.AsNode(), sourceCode),
			Optional:     param.QuestionToken != nil,
			IsRest:       param.DotDotDotToken != nil,
			DefaultValue: defaultValue,
		})
	}
}

// VisitFunctionDeclaration 解析函数声明。
// 此函数不仅提取函数的基本信息，还负责检查参数和返回类型中的显式 any。
func (p *Parser) VisitFunctionDeclaration(node *ast.FunctionDeclaration) {
	// 1. 解析函数声明本身的信息
	fr := NewFunctionDeclarationResult(node, p.SourceCode)
	// 2. 将解析结果存入
	p.Result.FunctionDeclarations = append(p.Result.FunctionDeclarations, *fr)
}