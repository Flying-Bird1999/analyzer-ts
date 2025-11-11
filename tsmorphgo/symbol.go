package tsmorphgo

import (
	"fmt"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// GetSymbol 从 AST 节点获取符号信息
// 直接使用 TypeScript-Go 的原生 Symbol() 方法
func GetSymbol(node Node) (*Symbol, error) {
	if !node.IsValid() {
		return nil, fmt.Errorf("invalid node")
	}

	// 直接使用 TypeScript-Go 原生 Symbol() 方法
	nativeSymbol := node.Node.Symbol()
	if nativeSymbol == nil {
		return nil, nil // 返回 nil 表示没有找到符号，不使用回退机制
	}

	return &Symbol{
		nativeSymbol: nativeSymbol,
	}, nil
}

// Symbol 包装 TypeScript 的原生符号
type Symbol struct {
	nativeSymbol *ast.Symbol
}

// GetName 获取符号名称
func (s *Symbol) GetName() string {
	if s == nil || s.nativeSymbol == nil {
		return ""
	}
	return s.nativeSymbol.Name
}

// String 返回符号的字符串表示（用于调试）
func (s *Symbol) String() string {
	if s == nil || s.nativeSymbol == nil {
		return "Symbol{nil}"
	}
	return fmt.Sprintf("Symbol{name: %s, flags: %d}",
		s.nativeSymbol.Name, int(s.nativeSymbol.Flags))
}
