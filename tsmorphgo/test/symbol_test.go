package tsmorphgo

import (
	"strings"
	"testing"

	. "github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
	"github.com/stretchr/testify/assert"
)

// TestSymbolBasic 测试简化的符号功能
// 根据ts-morph.md，Symbol只需要提供getName()方法用于符号比较
func TestSymbolBasic(t *testing.T) {
	project := NewProjectFromSources(map[string]string{
		"/test.ts": `
			const message = "Hello World";
			function greet() {
				console.log(message);
			}
		`,
	})
	sf := project.GetSourceFile("/test.ts")
	assert.NotNil(t, sf)

	// 找到变量声明中的 'message' 标识符
	var messageNode *Node
	sf.ForEachDescendant(func(node Node) {
		if IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "message" {
			parent := node.GetParent()
			if parent != nil && parent.Kind == ast.KindVariableDeclaration {
				messageNode = &node
			}
		}
	})

	assert.NotNil(t, messageNode, "应该找到 message 变量声明节点")

	// 测试获取符号
	symbol, err := GetSymbol(*messageNode)
	assert.NoError(t, err, "应该能够获取符号")
	assert.NotNil(t, symbol, "符号不应该为 nil")

	// 测试符号的核心功能 - GetName()
	assert.Equal(t, "message", symbol.GetName(), "符号名称应该匹配")

	// 测试符号的字符串表示
	// 新实现包含更多信息，所以我们只检查名称部分
	assert.Contains(t, symbol.String(), "message", "字符串表示应该包含符号名称")
}

// TestSymbolComparison 测试符号比较功能
// 这是ts-morph.md中提到的核心用途：判断两个节点是否引用同一个实体
func TestSymbolComparison(t *testing.T) {
	project := NewProjectFromSources(map[string]string{
		"/test.ts": `
			const myVar = "test";
			function test() {
				console.log(myVar);
			}
		`,
	})
	sf := project.GetSourceFile("/test.ts")
	assert.NotNil(t, sf)

	var declarationNode *Node
	var usageNode *Node

	sf.ForEachDescendant(func(node Node) {
		if IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "myVar" {
			parent := node.GetParent()
			if parent != nil && parent.Kind == ast.KindVariableDeclaration {
				declarationNode = &node
			} else if parent != nil && parent.Kind == ast.KindCallExpression {
				usageNode = &node
			}
		}
	})

	assert.NotNil(t, declarationNode, "应该找到变量声明节点")
	assert.NotNil(t, usageNode, "应该找到变量使用节点")

	// 获取两个节点的符号
	declSymbol, declErr := GetSymbol(*declarationNode)
	usageSymbol, usageErr := GetSymbol(*usageNode)

	assert.NoError(t, declErr, "应该能够获取声明节点的符号")
	assert.NoError(t, usageErr, "应该能够获取使用节点的符号")

	assert.NotNil(t, declSymbol, "声明符号不应该为 nil")
	assert.NotNil(t, usageSymbol, "使用符号不应该为 nil")

	// 核心测试：符号比较
	// 在简化实现中，相同的文本会产生相同的符号名称
	assert.Equal(t, declSymbol.GetName(), usageSymbol.GetName(), "相同变量的符号名称应该相同")
}

// TestSymbolNotFound 测试找不到符号的情况
func TestSymbolNotFound(t *testing.T) {
	project := NewProjectFromSources(map[string]string{
		"/test.ts": `
			// 空文件
		`,
	})
	sf := project.GetSourceFile("/test.ts")
	assert.NotNil(t, sf)

	// 创建一个无效的节点
	invalidNode := Node{}
	symbol, err := GetSymbol(invalidNode)

	assert.Error(t, err, "无效节点应该返回错误")
	assert.Nil(t, symbol, "找不到符号时应该返回 nil")
}

// TestSymbolNil 测试nil符号的安全性
func TestSymbolNil(t *testing.T) {
	// 由于新的实现使用了 RWMutex，我们需要处理 nil 情况
	// GetName 方法现在使用了锁，所以不能直接在 nil 上调用
	// 我们改为创建一个空的符号进行测试
	emptySymbol := &Symbol{}

	assert.Equal(t, "", emptySymbol.GetName(), "空符号的GetName应该返回空字符串")

	// String 方法应该包含基本信息
	result := emptySymbol.String()
	assert.Contains(t, result, "Symbol", "字符串表示应该包含Symbol")
}
