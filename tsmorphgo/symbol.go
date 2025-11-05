package tsmorphgo

import (
	"fmt"
	"strings"
)

// Symbol 表示一个简单的符号信息。
// 根据ts-morph.md的需求，只需要提供getName()方法用于符号比较。
type Symbol struct {
	name string // 符号名称
}

// GetName 返回符号的名称。
// 这是ts-morph.md中要求的唯一方法，用于符号比较。
func (s *Symbol) GetName() string {
	if s == nil {
		return ""
	}
	return s.name
}

// String 返回符号的字符串表示，用于调试。
func (s *Symbol) String() string {
	if s == nil {
		return "<nil symbol>"
	}
	return fmt.Sprintf("Symbol{name: %s}", s.GetName())
}

// GetSymbol 获取给定节点关联的语义符号。
// 简化实现：基于节点文本创建符号，满足基本的符号比较需求。
func GetSymbol(node Node) (*Symbol, bool) {
	// 基本验证
	if !node.IsValid() {
		return nil, false
	}

	// 获取节点文本作为符号名称
	nodeText := strings.TrimSpace(node.GetText())
	if nodeText == "" {
		return nil, false
	}

	// 创建简单的符号
	symbol := &Symbol{
		name: nodeText,
	}

	return symbol, true
}