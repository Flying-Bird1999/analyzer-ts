package tsmorphgo

import (
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// Symbol 代表一个语义符号，它是连接代码中多个引用的核心。
// 例如，一个变量的声明和它的所有使用之处，都指向同一个 Symbol。
type Symbol struct {
	inner *ast.Symbol
}

// GetName 返回符号的名称。
func (s *Symbol) GetName() string {
	if s.inner == nil {
		return ""
	}
	return s.inner.Name
}

// GetSymbol 获取给定节点关联的语义符号。
//
// # 当前状态: 未实现
//
// 原因: `GetSymbol` 依赖底层的 `typescript-go` 库来提供准确的符号信息。
// 经过多次尝试，我们发现稳定地从 `typescript-go` 的会话中获取类型检查器 (TypeChecker)
// 并调用其 `GetSymbolAtLocation` 方法存在困难，该方法在我们的测试场景下始终返回 nil。
//
// 这似乎是 `typescript-go` 库在项目管理和语义分析初始化方面的复杂性或潜在 bug 导致的。
// 在不直接修改 `typescript-go` 源码的前提下，目前没有找到可靠的实现路径。
//
// 后续计划: 此功能将被搁置，直到找到更可靠的底层 API，或 `typescript-go` 库的未来版本
// 提供了更清晰、更稳定的符号访问方式。
func GetSymbol(node Node) (*Symbol, bool) {
	// TODO: 待 `typescript-go` 提供更稳定的符号获取 API 后实现。
	return nil, false
}
