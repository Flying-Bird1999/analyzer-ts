package tsmorphgo

import (
	"fmt"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// SyntaxKind 统一透传底层 typescript-go 的语法节点类型
// 这样用户不需要直接导入 ast 包，保持 API 的一致性
type SyntaxKind = ast.Kind

// =============================================================================
// 常用语法节点类型常量（透传自底层 ast.Kind）
// 这些常量提供了类型安全和可读性，避免使用魔法数字
// =============================================================================

// 语句类型 (Statement Types)
const (
	KindVariableStatement        SyntaxKind = ast.KindVariableStatement        // 变量语句
	KindFunctionDeclaration      SyntaxKind = ast.KindFunctionDeclaration      // 函数声明
	KindInterfaceDeclaration     SyntaxKind = ast.KindInterfaceDeclaration     // 接口声明
	KindTypeAliasDeclaration     SyntaxKind = ast.KindTypeAliasDeclaration     // 类型别名声明
	KindClassDeclaration         SyntaxKind = ast.KindClassDeclaration         // 类声明
	KindEnumDeclaration          SyntaxKind = ast.KindEnumDeclaration          // 枚举声明
	KindImportDeclaration        SyntaxKind = ast.KindImportDeclaration        // 导入声明
	KindExportDeclaration        SyntaxKind = ast.KindExportDeclaration        // 导出声明
	KindExportAssignment         SyntaxKind = ast.KindExportAssignment         // export default 赋值
	KindReturnStatement          SyntaxKind = ast.KindReturnStatement          // return语句
	KindIfStatement              SyntaxKind = ast.KindIfStatement              // if语句
	KindForStatement             SyntaxKind = ast.KindForStatement             // for语句
	KindWhileStatement           SyntaxKind = ast.KindWhileStatement           // while语句
	KindTryStatement             SyntaxKind = ast.KindTryStatement             // try语句
	KindCatchClause              SyntaxKind = ast.KindCatchClause              // catch子句
)

// 表达式类型 (Expression Types)
const (
	KindCallExpression           SyntaxKind = ast.KindCallExpression           // 函数调用表达式
	KindPropertyAccessExpression SyntaxKind = ast.KindPropertyAccessExpression // 属性访问表达式
	KindPropertyAssignment       SyntaxKind = ast.KindPropertyAssignment       // 属性赋值
	KindPropertyDeclaration      SyntaxKind = ast.KindPropertyDeclaration      // 属性声明
	KindConditionalExpression    SyntaxKind = ast.KindConditionalExpression    // 条件表达式
	KindBinaryExpression         SyntaxKind = ast.KindBinaryExpression         // 二元表达式
	KindUnaryExpression          SyntaxKind = ast.KindPrefixUnaryExpression     // 一元表达式
	KindObjectLiteralExpression  SyntaxKind = ast.KindObjectLiteralExpression  // 对象字面量
	KindArrayLiteralExpression   SyntaxKind = ast.KindArrayLiteralExpression   // 数组字面量
	KindTemplateExpression       SyntaxKind = ast.KindTemplateExpression       // 模板字符串
	KindSpreadElement            SyntaxKind = ast.KindSpreadElement            // 展开运算符
	KindYieldExpression          SyntaxKind = ast.KindYieldExpression          // yield表达式
	KindAwaitExpression          SyntaxKind = ast.KindAwaitExpression          // await表达式
	KindTypeAssertionExpression  SyntaxKind = ast.KindTypeAssertionExpression  // 类型断言
)

// 基础类型 (Basic Types)
const (
	KindIdentifier               SyntaxKind = ast.KindIdentifier               // 标识符
	KindStringLiteral            SyntaxKind = ast.KindStringLiteral            // 字符串字面量
	KindNumericLiteral           SyntaxKind = ast.KindNumericLiteral           // 数字字面量
	KindTrueKeyword              SyntaxKind = ast.KindTrueKeyword              // true关键字
	KindFalseKeyword             SyntaxKind = ast.KindFalseKeyword             // false关键字
	KindNullKeyword              SyntaxKind = ast.KindNullKeyword              // null关键字
	KindUndefinedKeyword         SyntaxKind = ast.KindUndefinedKeyword         // undefined关键字
	KindThisKeyword              SyntaxKind = ast.KindThisKeyword              // this关键字
	KindSuperKeyword             SyntaxKind = ast.KindSuperKeyword             // super关键字
)

// 结构类型 (Structural Types)
const (
	KindVariableDeclaration      SyntaxKind = ast.KindVariableDeclaration      // 变量声明
	KindVariableDeclarationList  SyntaxKind = ast.KindVariableDeclarationList  // 变量声明列表
	KindParameter                SyntaxKind = ast.KindParameter                // 函数参数
	// KindParameterList            SyntaxKind = 0  // 参数列表 - 暂时注释，typescript-go 中未找到对应常量
	// KindArgument                 SyntaxKind = 0  // 函数调用参数 - 暂时注释，typescript-go 中未找到对应常量
	KindPropertySignature        SyntaxKind = ast.KindPropertySignature        // 属性签名
	KindMethodSignature          SyntaxKind = ast.KindMethodSignature          // 方法签名
	KindTypeParameter            SyntaxKind = ast.KindTypeParameter            // 类型参数
	KindTypeReference            SyntaxKind = ast.KindTypeReference            // 类型引用
)

// 关键字 (Keywords)
const (
	KindAsyncKeyword             SyntaxKind = ast.KindAsyncKeyword             // async关键字
	KindAwaitKeyword             SyntaxKind = ast.KindAwaitKeyword             // await关键字
	KindTypeKeyword              SyntaxKind = ast.KindTypeKeyword              // type关键字
	KindInterfaceKeyword         SyntaxKind = ast.KindInterfaceKeyword         // interface关键字
	KindConstKeyword             SyntaxKind = ast.KindConstKeyword             // const关键字
	KindLetKeyword               SyntaxKind = ast.KindLetKeyword               // let关键字
	KindVarKeyword               SyntaxKind = ast.KindVarKeyword               // var关键字
	KindImportKeyword            SyntaxKind = ast.KindImportKeyword            // import关键字
	KindExportKeyword            SyntaxKind = ast.KindExportKeyword            // export关键字
	KindFunctionKeyword          SyntaxKind = ast.KindFunctionKeyword          // function关键字
	KindClassKeyword             SyntaxKind = ast.KindClassKeyword             // class关键字
	KindExtendsKeyword           SyntaxKind = ast.KindExtendsKeyword           // extends关键字
	KindImplementsKeyword        SyntaxKind = ast.KindImplementsKeyword        // implements关键字
)

// 运算符 (Operators)
const (
	KindPlusToken                SyntaxKind = ast.KindPlusToken                // + 运算符
	KindMinusToken               SyntaxKind = ast.KindMinusToken               // - 运算符
	KindAsteriskToken            SyntaxKind = ast.KindAsteriskToken            // * 运算符
	KindSlashToken               SyntaxKind = ast.KindSlashToken               // / 运算符
	KindEqualsToken              SyntaxKind = ast.KindEqualsToken              // = 运算符
	KindEqualsEqualsEqualsToken  SyntaxKind = ast.KindEqualsEqualsEqualsToken  // === 运算符
	KindExclamationEqualsEqualsToken SyntaxKind = ast.KindExclamationEqualsEqualsToken // !== 运算符
)

// 类相关 (Class Related)
const (
	KindConstructor              SyntaxKind = ast.KindConstructor              // 构造函数
	KindMethodDeclaration        SyntaxKind = ast.KindMethodDeclaration        // 方法声明
	KindGetAccessor              SyntaxKind = ast.KindGetAccessor              // getter访问器
	KindSetAccessor              SyntaxKind = ast.KindSetAccessor              // setter访问器
)

// 模块相关 (Module Related)
const (
	KindImportClause             SyntaxKind = ast.KindImportClause             // 导入子句
	KindImportSpecifier          SyntaxKind = ast.KindImportSpecifier          // 导入说明符
	KindExportSpecifier          SyntaxKind = ast.KindExportSpecifier          // 导出说明符
)

// JSX 相关 (JSX Related)
const (
	KindJsxElement               SyntaxKind = ast.KindJsxElement               // JSX 元素
	KindJsxSelfClosingElement    SyntaxKind = ast.KindJsxSelfClosingElement    // JSX 自闭合元素
	KindJsxOpeningElement        SyntaxKind = ast.KindJsxOpeningElement        // JSX 开始元素
	KindJsxClosingElement        SyntaxKind = ast.KindJsxClosingElement        // JSX 结束元素
	KindJsxAttribute             SyntaxKind = ast.KindJsxAttribute             // JSX 属性
)

// 其他 (Other)
const (
	KindSourceFile               SyntaxKind = ast.KindSourceFile               // 源文件
	KindBlock                    SyntaxKind = ast.KindBlock                    // 代码块
	// KindEndOfFileToken           SyntaxKind = 0  // 文件结束标记 - 暂时注释，typescript-go 中未找到对应常量
)

// =============================================================================
// 语法节点类型分类和工具函数
// =============================================================================

// SyntaxKindCategories 语法节点类型分类
type SyntaxKindCategories struct {
	Statements        []SyntaxKind
	Expressions       []SyntaxKind
	Keywords          []SyntaxKind
	Operators         []SyntaxKind
	ClassRelated      []SyntaxKind
	ModuleRelated     []SyntaxKind
	Declarations      []SyntaxKind
	Literals          []SyntaxKind
}

// GetSyntaxKindCategories 获取所有语法节点分类
func GetSyntaxKindCategories() SyntaxKindCategories {
	return SyntaxKindCategories{
		Statements: []SyntaxKind{
			KindVariableStatement,
			KindFunctionDeclaration,
			KindInterfaceDeclaration,
			KindTypeAliasDeclaration,
			KindClassDeclaration,
			KindEnumDeclaration,
			KindImportDeclaration,
			KindExportDeclaration,
			KindReturnStatement,
			KindIfStatement,
			KindForStatement,
			KindWhileStatement,
			KindTryStatement,
			KindCatchClause,
		},
		Expressions: []SyntaxKind{
			KindCallExpression,
			KindPropertyAccessExpression,
			KindPropertyAssignment,
			KindPropertyDeclaration,
			KindConditionalExpression,
			KindBinaryExpression,
			KindUnaryExpression,
			KindObjectLiteralExpression,
			KindArrayLiteralExpression,
			KindTemplateExpression,
			KindSpreadElement,
			KindYieldExpression,
			KindAwaitExpression,
			KindTypeAssertionExpression,
		},
		Keywords: []SyntaxKind{
			KindAsyncKeyword,
			KindAwaitKeyword,
			KindTypeKeyword,
			KindInterfaceKeyword,
			KindConstKeyword,
			KindLetKeyword,
			KindVarKeyword,
			KindImportKeyword,
			KindExportKeyword,
			KindFunctionKeyword,
			KindClassKeyword,
			KindExtendsKeyword,
			KindImplementsKeyword,
			KindTrueKeyword,
			KindFalseKeyword,
			KindNullKeyword,
			KindUndefinedKeyword,
			KindThisKeyword,
			KindSuperKeyword,
		},
		Operators: []SyntaxKind{
			KindPlusToken,
			KindMinusToken,
			KindAsteriskToken,
			KindSlashToken,
			KindEqualsEqualsEqualsToken,
			KindExclamationEqualsEqualsToken,
		},
		ClassRelated: []SyntaxKind{
			KindClassDeclaration,
			KindConstructor,
			KindMethodDeclaration,
			KindGetAccessor,
			KindSetAccessor,
			KindPropertyDeclaration,
		},
		ModuleRelated: []SyntaxKind{
			KindImportDeclaration,
			KindImportClause,
			KindImportSpecifier,
			KindExportDeclaration,
			KindExportSpecifier,
		},
		Declarations: []SyntaxKind{
			KindVariableDeclaration,
			KindFunctionDeclaration,
			KindInterfaceDeclaration,
			KindTypeAliasDeclaration,
			KindClassDeclaration,
			KindEnumDeclaration,
			KindParameter,
			KindTypeParameter,
			KindPropertySignature,
			KindMethodSignature,
		},
		Literals: []SyntaxKind{
			KindStringLiteral,
			KindNumericLiteral,
			KindTrueKeyword,
			KindFalseKeyword,
			KindNullKeyword,
			KindUndefinedKeyword,
		},
	}
}

// IsStatement 检查是否为语句类型
func IsStatement(kind SyntaxKind) bool {
	categories := GetSyntaxKindCategories()
	for _, statementKind := range categories.Statements {
		if kind == statementKind {
			return true
		}
	}
	return false
}

// IsExpression 检查是否为表达式类型
func IsExpression(kind SyntaxKind) bool {
	categories := GetSyntaxKindCategories()
	for _, exprKind := range categories.Expressions {
		if kind == exprKind {
			return true
		}
	}
	return false
}

// IsKeyword 检查是否为关键字
func IsKeyword(kind SyntaxKind) bool {
	categories := GetSyntaxKindCategories()
	for _, keywordKind := range categories.Keywords {
		if kind == keywordKind {
			return true
		}
	}
	return false
}

// IsOperator 检查是否为运算符
func IsOperator(kind SyntaxKind) bool {
	categories := GetSyntaxKindCategories()
	for _, opKind := range categories.Operators {
		if kind == opKind {
			return true
		}
	}
	return false
}

// IsClassRelated 检查是否为类相关类型
func IsClassRelated(kind SyntaxKind) bool {
	categories := GetSyntaxKindCategories()
	for _, classKind := range categories.ClassRelated {
		if kind == classKind {
			return true
		}
	}
	return false
}

// IsModuleRelated 检查是否为模块相关类型
func IsModuleRelated(kind SyntaxKind) bool {
	categories := GetSyntaxKindCategories()
	for _, moduleKind := range categories.ModuleRelated {
		if kind == moduleKind {
			return true
		}
	}
	return false
}

// IsDeclaration 检查是否为声明类型
func IsDeclaration(kind SyntaxKind) bool {
	categories := GetSyntaxKindCategories()
	for _, declKind := range categories.Declarations {
		if kind == declKind {
			return true
		}
	}
	return false
}

// IsLiteral 检查是否为字面量类型
func IsLiteral(kind SyntaxKind) bool {
	categories := GetSyntaxKindCategories()
	for _, literalKind := range categories.Literals {
		if kind == literalKind {
			return true
		}
	}
	return false
}

// GetCategory 获取语法节点类型的分类
func GetCategory(kind SyntaxKind) string {
	if IsStatement(kind) {
		return "Statement"
	}
	if IsExpression(kind) {
		return "Expression"
	}
	if IsKeyword(kind) {
		return "Keyword"
	}
	if IsOperator(kind) {
		return "Operator"
	}
	if IsClassRelated(kind) {
		return "ClassRelated"
	}
	if IsModuleRelated(kind) {
		return "ModuleRelated"
	}
	if IsDeclaration(kind) {
		return "Declaration"
	}
	if IsLiteral(kind) {
		return "Literal"
	}
	return "Other"
}

// GetKindName 获取语法节点类型的名称
// 这个函数提供了友好的名称显示，用于调试和日志
func GetKindName(kind SyntaxKind) string {
	// 使用 ast 包的字符串转换
	return kind.String()
}

// IsValidSyntaxKind 检查是否为有效的语法节点类型
func IsValidSyntaxKind(kind SyntaxKind) bool {
	// 基本的有效性检查
	return kind >= 0 && kind <= 1000 // 根据实际情况调整范围
}

// =============================================================================
// 常用语法节点类型组合
// =============================================================================

// FunctionRelatedKinds 函数相关的语法节点类型
var FunctionRelatedKinds = []SyntaxKind{
	KindFunctionDeclaration,
	KindCallExpression,
	KindParameter,
	KindReturnStatement,
	KindAsyncKeyword,
	KindAwaitKeyword,
}

// TypeRelatedKinds 类型相关的语法节点类型
var TypeRelatedKinds = []SyntaxKind{
	KindInterfaceDeclaration,
	KindTypeAliasDeclaration,
	KindTypeParameter,
	KindTypeReference,
	KindTypeKeyword,
	KindInterfaceKeyword,
	KindPropertySignature,
	KindMethodSignature,
}

// VariableRelatedKinds 变量相关的语法节点类型
var VariableRelatedKinds = []SyntaxKind{
	KindVariableDeclaration,
	KindVariableDeclarationList,
	KindVariableStatement,
	KindLetKeyword,
	KindConstKeyword,
	KindVarKeyword,
	KindIdentifier,
}

// ControlFlowKinds 控制流相关的语法节点类型
var ControlFlowKinds = []SyntaxKind{
	KindIfStatement,
	KindForStatement,
	KindWhileStatement,
	KindConditionalExpression,
	KindTryStatement,
	KindCatchClause,
	KindReturnStatement,
	KindYieldExpression,
}

// InFunctionRelated 检查是否为函数相关类型
func InFunctionRelated(kind SyntaxKind) bool {
	for _, k := range FunctionRelatedKinds {
		if kind == k {
			return true
		}
	}
	return false
}

// InTypeRelated 检查是否为类型相关类型
func InTypeRelated(kind SyntaxKind) bool {
	for _, k := range TypeRelatedKinds {
		if kind == k {
			return true
		}
	}
	return false
}

// InVariableRelated 检查是否为变量相关类型
func InVariableRelated(kind SyntaxKind) bool {
	for _, k := range VariableRelatedKinds {
		if kind == k {
			return true
		}
	}
	return false
}

// InControlFlow 检查是否为控制流相关类型
func InControlFlow(kind SyntaxKind) bool {
	for _, k := range ControlFlowKinds {
		if kind == k {
			return true
		}
	}
	return false
}

// =============================================================================
// 调试和诊断工具
// =============================================================================

// PrintSyntaxKindInfo 打印语法节点类型的详细信息（用于调试）
func PrintSyntaxKindInfo(kind SyntaxKind) {
	fmt.Printf("语法节点类型信息:\n")
	fmt.Printf("  类型值: %d\n", kind)
	fmt.Printf("  类型名: %s\n", GetKindName(kind))
	fmt.Printf("  分类: %s\n", GetCategory(kind))

	fmt.Printf("  特征:\n")
	fmt.Printf("    - 是语句: %v\n", IsStatement(kind))
	fmt.Printf("    - 是表达式: %v\n", IsExpression(kind))
	fmt.Printf("    - 是关键字: %v\n", IsKeyword(kind))
	fmt.Printf("    - 是运算符: %v\n", IsOperator(kind))
	fmt.Printf("    - 是声明: %v\n", IsDeclaration(kind))
	fmt.Printf("    - 是字面量: %v\n", IsLiteral(kind))
	fmt.Printf("    - 类相关: %v\n", IsClassRelated(kind))
	fmt.Printf("    - 模块相关: %v\n", IsModuleRelated(kind))

	fmt.Printf("  关联性:\n")
	fmt.Printf("    - 函数相关: %v\n", InFunctionRelated(kind))
	fmt.Printf("    - 类型相关: %v\n", InTypeRelated(kind))
	fmt.Printf("    - 变量相关: %v\n", InVariableRelated(kind))
	fmt.Printf("    - 控制流: %v\n", InControlFlow(kind))
}

// GetCommonSyntaxKinds 获取最常用的语法节点类型（用于快速查找）
func GetCommonSyntaxKinds() []SyntaxKind {
	return []SyntaxKind{
		// 最常见的类型
		KindIdentifier,
		KindStringLiteral,
		KindNumericLiteral,

		// 常见的语句
		KindVariableStatement,
		KindFunctionDeclaration,
		KindReturnStatement,
		KindIfStatement,

		// 常见的表达式
		KindCallExpression,
		KindPropertyAccessExpression,
		KindBinaryExpression,
		KindObjectLiteralExpression,

		// 常见的声明
		KindVariableDeclaration,
		KindInterfaceDeclaration,
		KindClassDeclaration,

		// 常见的关键字
		KindConstKeyword,
		KindLetKeyword,
		KindFunctionKeyword,
		KindClassKeyword,
		KindImportKeyword,
		KindExportKeyword,
	}
}