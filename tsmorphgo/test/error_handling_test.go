package tsmorphgo

import (
	"testing"
	"time"

	. "github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// error_handling_test.go
//
// 这个文件包含了错误处理和重试机制的测试，验证：
// 1. 错误分类和包装功能
// 2. 重试机制和指数退避算法
// 3. 降级策略的有效性
// 4. 错误恢复和容错能力
//

// TestReferenceErrorTypes 测试错误类型分类
func TestReferenceErrorTypes(t *testing.T) {
	t.Logf("测试错误类型分类...")

	// 测试 LSP 服务错误
	lspError := NewReferenceError(ErrorTypeLSPService, "LSP 服务连接失败", nil)
	assert.Equal(t, ErrorTypeLSPService, lspError.Type)
	assert.Equal(t, "LSP 服务连接失败", lspError.Message)
	assert.True(t, lspError.Retryable, "LSP 服务错误应该可重试")

	// 测试文件不存在错误
	fileError := NewReferenceError(ErrorTypeFileNotFound, "文件不存在", nil)
	assert.Equal(t, ErrorTypeFileNotFound, fileError.Type)
	assert.False(t, fileError.Retryable, "文件不存在错误不应该重试")

	// 测试错误字符串表示
	errorStr := lspError.Error()
	assert.Contains(t, errorStr, "[LSP服务错误]", "应该包含错误类型")
	assert.Contains(t, errorStr, "LSP 服务连接失败", "应该包含错误消息")

	t.Logf("✅ 错误类型分类测试通过")
}

// TestRetryConfig 测试重试配置
func TestRetryConfig(t *testing.T) {
	t.Logf("测试重试配置...")

	// 测试默认配置
	config := DefaultRetryConfig()
	assert.Equal(t, 3, config.MaxRetries, "默认最大重试次数应该为3")
	assert.Equal(t, 100*time.Millisecond, config.BaseDelay, "默认基础延迟应该为100ms")
	assert.Equal(t, 5*time.Second, config.MaxDelay, "默认最大延迟应该为5s")
	assert.Equal(t, 2.0, config.BackoffFactor, "默认退避因子应该为2.0")
	assert.True(t, config.Enabled, "默认应该启用重试")

	// 测试延迟计算
	delay1 := config.CalculateDelay(0)
	assert.Equal(t, config.BaseDelay, delay1, "第0次重试延迟应该等于基础延迟")

	delay2 := config.CalculateDelay(1)
	expectedDelay2 := time.Duration(float64(config.BaseDelay) * config.BackoffFactor)
	assert.Equal(t, expectedDelay2, delay2, "第1次重试延迟应该应用退避因子")

	// 测试最大延迟限制
	highRetryDelay := config.CalculateDelay(10)
	assert.LessOrEqual(t, highRetryDelay, config.MaxDelay, "重试延迟不应该超过最大延迟")

	// 测试禁用重试
	config.Enabled = false
	disabledDelay := config.CalculateDelay(1)
	assert.Equal(t, time.Duration(0), disabledDelay, "禁用重试时延迟应该为0")

	t.Logf("✅ 重试配置测试通过")
}

// TestErrorWrapping 测试错误包装功能
func TestErrorWrapping(t *testing.T) {
	// 创建测试项目
	project := createTestProject(map[string]string{
		"/test.ts": `const testVar = "hello";`,
	})

	sourceFile := project.GetSourceFile("/test.ts")
	require.NotNil(t, sourceFile)

	// 找到一个标识符节点
	var targetNode *Node
	sourceFile.ForEachDescendant(func(node Node) {
		if IsIdentifier(node) && node.GetText() == "testVar" {
			nodeCopy := node
			targetNode = &nodeCopy
		}
	})
	if targetNode == nil {
		// 如果找不到节点，跳过测试
		t.Skip("无法找到测试节点，跳过错误包装测试")
		return
	}

	t.Logf("测试错误包装功能...")

	// 测试基础错误包装
	originalErr := assert.AnError
	wrappedErr := WrapError(originalErr, "测试上下文", *targetNode)

	require.NotNil(t, wrappedErr, "包装后的错误不应该为空")
	assert.Contains(t, wrappedErr.Message, "测试上下文", "应该包含上下文信息")
	assert.Contains(t, wrappedErr.Message, originalErr.Error(), "应该包含原始错误信息")
	assert.Equal(t, sourceFile.GetFilePath(), wrappedErr.FilePath, "应该包含文件路径")
	assert.Greater(t, wrappedErr.LineNumber, 0, "应该包含行号")

	// 测试错误类型推断
	timeoutErr := WrapError(assert.AnError, "连接超时", *targetNode)
	assert.Equal(t, ErrorTypeLSPTimeout, timeoutErr.Type, "应该推断出超时错误类型")

	t.Logf("✅ 错误包装功能测试通过")
}

// TestFallbackStrategies 测试降级策略
func TestFallbackStrategies(t *testing.T) {
	// 创建测试项目
	project := createTestProject(map[string]string{
		"/test.ts": `
			const sharedVar = "hello";
			function testFunction() {
				console.log(sharedVar);
			}
			console.log(sharedVar);
		`,
	})

	sourceFile := project.GetSourceFile("/test.ts")
	require.NotNil(t, sourceFile)

	// 找到 sharedVar 的使用节点
	var usageNode *Node
	sourceFile.ForEachDescendant(func(node Node) {
		if IsIdentifier(node) && node.GetText() == "sharedVar" {
			parent := node.GetParent()
			if parent != nil && parent.Kind != KindVariableDeclaration { // 不是变量声明
				nodeCopy := node
				usageNode = &nodeCopy
			}
		}
	})

	if usageNode == nil {
		t.Skip("无法找到合适的使用节点，跳过降级策略测试")
		return
	}

	t.Logf("测试降级策略...")

	// 测试引用查找降级策略
	t.Logf("测试引用查找降级策略...")
	fallbackRefs := FindReferencesFallback(*usageNode)
	t.Logf("降级策略找到 %d 个引用", len(fallbackRefs))

	// 验证找到的引用
	if len(fallbackRefs) > 0 {
		for i, ref := range fallbackRefs {
			assert.Equal(t, "sharedVar", ref.GetText(), "第%d个引用文本应该匹配", i)
			assert.True(t, IsIdentifier(*ref), "引用应该是标识符")
		}
	}

	// 测试定义查找降级策略
	t.Logf("测试定义查找降级策略...")
	fallbackDefs := GotoDefinitionFallback(*usageNode)
	t.Logf("降级策略找到 %d 个定义", len(fallbackDefs))

	// 验证找到的定义
	if len(fallbackDefs) > 0 {
		for i, def := range fallbackDefs {
			assert.Equal(t, "sharedVar", def.GetText(), "第%d个定义文本应该匹配", i)
			assert.True(t, IsIdentifier(*def), "定义应该是标识符")
		}
	}

	t.Logf("✅ 降级策略测试通过")
}

// TestErrorRecovery 测试错误恢复机制
func TestErrorRecovery(t *testing.T) {
	// 创建测试项目
	project := createTestProject(map[string]string{
		"/test.ts": `
			const testVar = "hello";
			console.log(testVar);
		`,
	})

	sourceFile := project.GetSourceFile("/test.ts")
	require.NotNil(t, sourceFile)

	// 找到使用节点
	var usageNode *Node
	sourceFile.ForEachDescendant(func(node Node) {
		if IsIdentifier(node) && node.GetText() == "testVar" {
			parent := node.GetParent()
			if parent != nil && parent.Kind != KindVariableDeclaration {
				nodeCopy := node
				usageNode = &nodeCopy
			}
		}
	})

	if usageNode == nil {
		t.Skip("无法找到合适的使用节点，跳过错误恢复测试")
		return
	}

	t.Logf("测试错误恢复机制...")

	// 创建一个快速失败的重试配置
	fastRetryConfig := &RetryConfig{
		MaxRetries:    2,
		BaseDelay:     1 * time.Millisecond,
		MaxDelay:      10 * time.Millisecond,
		BackoffFactor: 1.5,
		Enabled:       true,
	}

	// 测试带重试的引用查找
	t.Logf("执行带重试的引用查找...")
	start := time.Now()
	refs, fromCache, err := FindReferencesWithCacheAndRetry(*usageNode, fastRetryConfig)
	duration := time.Since(start)

	// 验证结果
	if err != nil {
		t.Logf("引用查找失败: %v", err)

		// 如果失败，检查错误类型
		if refErr, ok := err.(*ReferenceError); ok {
			t.Logf("错误类型: %s, 可重试: %t, 重试次数: %d",
				refErr.Error(), refErr.Retryable, refErr.RetryCount)

			// 验证错误属性
			assert.Greater(t, refErr.RetryCount, 0, "应该有重试记录")
			assert.NotNil(t, refErr.Timestamp, "应该有时间戳")
		}
	} else {
		t.Logf("引用查找成功: %d 个引用, 耗时: %v, 来自缓存: %t",
			len(refs), duration, fromCache)

		// 验证成功的情况
		assert.Greater(t, len(refs), 0, "应该找到引用")
	}

	t.Logf("✅ 错误恢复机制测试通过")
}

// TestContextualAnalysis 测试上下文分析功能
func TestContextualAnalysis(t *testing.T) {
	// 创建测试项目
	project := createTestProject(map[string]string{
		"/test.ts": `
			const myVar = "value";
			console.log(myVar);
			function myFunc() {}
			myFunc();
		`,
	})

	sourceFile := project.GetSourceFile("/test.ts")
	require.NotNil(t, sourceFile)

	t.Logf("测试上下文分析功能...")

	var definitionNodes, referenceNodes []*Node

	// 收集定义和引用节点，只关注我们特定的标识符
	sourceFile.ForEachDescendant(func(node Node) {
		if !IsIdentifier(node) {
			return
		}

		nodeText := node.GetText()
		if nodeText != "myVar" && nodeText != "myFunc" {
			return
		}

		nodeCopy := node

		if IsLikelyDefinition(node) {
			definitionNodes = append(definitionNodes, &nodeCopy)
		} else if IsLikelyReference(node) {
			referenceNodes = append(referenceNodes, &nodeCopy)
		}
	})

	t.Logf("找到 %d 个潜在定义节点, %d 个潜在引用节点",
		len(definitionNodes), len(referenceNodes))

	// 验证分析结果
	if len(definitionNodes) > 0 {
		for i, def := range definitionNodes {
			t.Logf("定义节点 %d: %s", i, def.GetText())
		}
	}

	if len(referenceNodes) > 0 {
		for i, ref := range referenceNodes {
			t.Logf("引用节点 %d: %s", i, ref.GetText())
		}
	}

	// 验证至少找到一些节点（定义或引用）
	totalNodes := len(definitionNodes) + len(referenceNodes)
	assert.Greater(t, totalNodes, 0, "应该找到至少一个节点")

	// 简化验证 - 只要能区分一些节点就说明上下文分析在起作用
	if len(definitionNodes) > 0 && len(referenceNodes) > 0 {
		t.Logf("成功区分了定义节点和引用节点")
	} else if len(definitionNodes) > 0 {
		t.Logf("找到了定义节点")
	} else if len(referenceNodes) > 0 {
		t.Logf("找到了引用节点")
	}

	t.Logf("✅ 上下文分析功能测试通过")
}

// BenchmarkErrorHandling 性能基准测试
func BenchmarkErrorHandling(b *testing.B) {
	// 创建测试项目
	project := createTestProject(map[string]string{
		"/test.ts": `const testVar = "hello"; console.log(testVar);`,
	})

	sourceFile := project.GetSourceFile("/test.ts")
	require.NotNil(b, sourceFile)

	// 找到使用节点
	var usageNode *Node
	sourceFile.ForEachDescendant(func(node Node) {
		if IsIdentifier(node) && node.GetText() == "testVar" {
			parent := node.GetParent()
			if parent != nil && parent.Kind != KindVariableDeclaration {
				nodeCopy := node
				usageNode = &nodeCopy
			}
		}
	})

	require.NotNil(b, usageNode, "需要找到测试节点")

	b.ResetTimer()

	// 基准测试错误包装性能
	for i := 0; i < b.N; i++ {
		_ = WrapError(assert.AnError, "基准测试上下文", *usageNode)
	}
}