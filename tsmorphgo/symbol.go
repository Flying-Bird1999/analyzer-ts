package tsmorphgo

import (
	"context"
	"fmt"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// GetSymbol 从 AST 节点获取符号信息
// 优先使用 TypeChecker.GetSymbolAtLocation 方法，如果失败则回退到节点的 Symbol() 方法
func GetSymbol(node Node) (*Symbol, error) {
	if !node.IsValid() {
		return nil, fmt.Errorf("invalid node")
	}

	// 尝试使用 TypeChecker.GetSymbolAtLocation 获取更准确的符号信息
	if nativeSymbol := getSymbolViaTypeChecker(node); nativeSymbol != nil {
		return &Symbol{nativeSymbol: nativeSymbol}, nil
	}

	// 回退到直接使用节点的 Symbol() 方法
	nativeSymbol := node.Node.Symbol()
	if nativeSymbol == nil {
		return nil, nil
	}

	return &Symbol{nativeSymbol: nativeSymbol}, nil
}

// GetSymbolAtLocation 使用 TypeChecker.GetSymbolAtLocation 获取符号
// 这是 TypeScript-Go 推荐的标准方法，能提供更准确的符号信息
func GetSymbolAtLocation(node Node) (*Symbol, error) {
	if !node.IsValid() {
		return nil, fmt.Errorf("invalid node")
	}

	nativeSymbol := getSymbolViaTypeChecker(node)
	if nativeSymbol == nil {
		return nil, fmt.Errorf("no symbol found at location")
	}

	return &Symbol{nativeSymbol: nativeSymbol}, nil
}

// getSymbolViaTypeChecker 通过 TypeChecker 获取符号，使用 TypeScript-Go 底层方法
func getSymbolViaTypeChecker(node Node) *ast.Symbol {
	sourceFile := node.GetSourceFile()
	if sourceFile == nil {
		return nil
	}

	project := sourceFile.GetProject()
	if project == nil {
		return nil
	}

	// 获取 LSP 服务
	lspService, err := project.getLspService()
	if err != nil {
		return nil
	}

	// 使用 LSP 服务的符号获取方法，这内部使用了 TypeChecker.GetSymbolAtLocation
	filePath := sourceFile.GetFilePath()

	// LSP 服务需要行号和列号，而不是绝对位置
	line := node.GetStartLineNumber()
	column := node.GetStartColumnNumber()

	symbol, err := lspService.GetSymbolAt(context.Background(), filePath, line, column)
	if err != nil || symbol == nil {
		return nil
	}

	return symbol
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

// GetFlags 获取符号的标志值（用于高级分析）
func (s *Symbol) GetFlags() uint32 {
	if s == nil || s.nativeSymbol == nil {
		return 0
	}
	return uint32(s.nativeSymbol.Flags)
}
