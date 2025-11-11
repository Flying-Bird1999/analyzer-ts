package tsmorphgo_test

// import (
// 	"testing"

// 	"github.com/stretchr/testify/assert"

// 	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
// )

// // TestVariableDeclaration_Comprehensive 测试VariableDeclaration专有API的全面功能
// func TestVariableDeclaration_Comprehensive(t *testing.T) {
// 	project := tsmorphgo.NewProjectFromSources(map[string]string{
// 		"/test.ts": `
// 			// 有初始值的变量
// 			const x = 42;
// 			let y = "hello";
// 			var z = true;

// 			// 无初始值的变量
// 			const a;
// 			let b;
// 			var c;

// 			// 复杂初始值
// 			const obj = { key: "value" };
// 			const arr = [1, 2, 3];
// 			const func = function() { return 1; };
// 			const arrow = () => 2;
// 			const result = x + y * z;

// 			// 解构赋值
// 			const { prop1, prop2 } = obj;
// 			const [item1, item2] = arr;
// 		`,
// 	})
// 	defer project.Close()

// 	sf := project.GetSourceFile("/test.ts")
// 	assert.NotNil(t, sf)

// 	var withInitializer, withoutInitializer int
// 	var complexInitializers []string

// 	sf.ForEachDescendant(func(node tsmorphgo.Node) {
// 		if varDecl, ok := node.AsVariableDeclaration(); ok {
// 			name := varDecl.GetName()
// 			assert.NotEmpty(t, name, "变量名不应为空")

// 			// 测试GetNameNode
// 			nameNode := varDecl.GetNameNode()
// 			assert.True(t, nameNode.IsValid(), "名称节点应该有效")
// 			assert.Equal(t, tsmorphgo.KindIdentifier, nameNode.GetKind())
// 			assert.Equal(t, name, nameNode.GetText())

// 			// 测试HasInitializer和GetInitializer
// 			hasInit := varDecl.HasInitializer()
// 			initializer := varDecl.GetInitializer()

// 			if hasInit {
// 				withInitializer++
// 				assert.True(t, initializer.IsValid(), "有初始值的变量应该返回有效的初始值节点")

// 				// 记录复杂的初始值
// 				initText := initializer.GetText()
// 				if len(initText) > 10 { // 假设长度超过10的是复杂表达式
// 					complexInitializers = append(complexInitializers, initText)
// 					t.Logf("复杂初始值 %s: %s", name, initText)
// 				}
// 			} else {
// 				withoutInitializer++
// 				assert.False(t, initializer.IsValid(), "无初始值的变量应该返回无效的初始值节点")
// 			}

// 			t.Logf("变量 %s: HasInitializer=%t", name, hasInit)
// 		}
// 	})

// 	assert.Greater(t, withInitializer, 0, "应该找到有初始值的变量")
// 	assert.Greater(t, withoutInitializer, 0, "应该找到无初始值的变量")
// 	assert.GreaterOrEqual(t, len(complexInitializers), 3, "应该找到复杂初始值表达式")

// 	t.Logf("统计: 有初始值=%d, 无初始值=%d, 复杂表达式=%d",
// 		withInitializer, withoutInitializer, len(complexInitializers))
// }

// // TestCallExpression_Comprehensive 测试CallExpression专有API的全面功能
// func TestCallExpression_Comprehensive(t *testing.T) {
// 	project := tsmorphgo.NewProjectFromSources(map[string]string{
// 		"/test.ts": `
// 			// 简单函数调用
// 			test();
// 			obj.method();
// 			obj.nested.method();

// 			// 带参数的调用
// 			test(1);
// 			test(1, "hello", true);
// 			obj.method(param1, param2);

// 			// 复杂表达式作为函数
// 			(obj.test)();
// 			(result || defaultValue)();
// 			(callback ? callback : defaultCallback)(arg);

// 			// 链式调用
// 			getData().process().format();

// 			// 构造函数调用
// 			new Constructor();
// 		 new Constructor(arg1, arg2);

// 			// 其他函数调用
// 			someOtherFunction();

// 			// 立即执行函数
// 			(function() { return 1; })();
// 			(() => 2)();
// 		`,
// 	})
// 	defer project.Close()

// 	sf := project.GetSourceFile("/test.ts")
// 	assert.NotNil(t, sf)

// 	var calls []struct {
// 		expr string
// 		args int
// 	}

// 	sf.ForEachDescendant(func(node tsmorphgo.Node) {
// 		if callExpr, ok := node.AsCallExpression(); ok {
// 			expr := callExpr.GetExpression()
// 			assert.True(t, expr.IsValid(), "调用表达式应该有效")

// 			args := callExpr.GetArguments()
// 			argCount := callExpr.GetArgumentCount()
// 			assert.Equal(t, len(args), argCount, "参数数量应该一致")

// 			exprText := expr.GetText()
// 			calls = append(calls, struct {
// 				expr string
// 				args int
// 			}{expr: exprText, args: argCount})

// 			t.Logf("函数调用: %s (参数数量: %d)", exprText, argCount)

// 			// 详细检查每个参数
// 			for i, arg := range args {
// 				assert.True(t, arg.IsValid(), "参数节点应该有效")
// 				t.Logf("  参数 %d: %s (类型: %s)", i, arg.GetText(), arg.GetKindName())
// 			}
// 		}
// 	})

// 	assert.Greater(t, len(calls), 10, "应该找到多个函数调用")

// 	// 验证特定调用模式
// 	var zeroArgCalls, multiArgCalls int
// 	for _, call := range calls {
// 		if call.args == 0 {
// 			zeroArgCalls++
// 		} else if call.args > 1 {
// 			multiArgCalls++
// 		}
// 	}

// 	assert.Greater(t, zeroArgCalls, 0, "应该找到无参数调用")
// 	assert.Greater(t, multiArgCalls, 0, "应该找到多参数调用")

// 	t.Logf("调用统计: 总数=%d, 无参数=%d, 多参数=%d", len(calls), zeroArgCalls, multiArgCalls)
// }

// // TestPropertyAccessExpression_Comprehensive 测试PropertyAccessExpression专有API的全面功能
// func TestPropertyAccessExpression_Comprehensive(t *testing.T) {
// 	project := tsmorphgo.NewProjectFromSources(map[string]string{
// 		"/test.ts": `
// 			// 简单属性访问
// 			obj.prop;
// 			obj.method;
// 			obj.nested.prop;

// 			// 复杂对象表达式
// 			(result || defaultValue).property;
// 			(callback ? callback : defaultCallback).method;
// 			(items[0]).value;

// 			// 链式属性访问
// 			a.b.c.d.e;

// 			// 计算属性访问（不应该被识别为PropertyAccessExpression）
// 			obj["key"];
// 			obj[variable];

// 			// 带调用的属性访问
// 			obj.method();
// 			obj.nested.method().prop;

// 			// this的属性访问
// 			this.property;
// 			this.method();

// 			// 模块访问
// 			module.exportedFunction;
// 			module.CONSTANT;

// 			// 条件表达式中的属性访问
// 			condition ? obj.prop1 : obj.prop2;
// 		`,
// 	})
// 	defer project.Close()

// 	sf := project.GetSourceFile("/test.ts")
// 	assert.NotNil(t, sf)

// 	var accesses []struct {
// 		name  string
// 		expr  string
// 		valid bool
// 	}

// 	sf.ForEachDescendant(func(node tsmorphgo.Node) {
// 		if propAccess, ok := node.AsPropertyAccessExpression(); ok {
// 			name := propAccess.GetName()
// 			expr := propAccess.GetExpression()

// 			isValid := name != "" && expr.IsValid()
// 			accesses = append(accesses, struct {
// 				name  string
// 				expr  string
// 				valid bool
// 			}{name: name, expr: expr.GetText(), valid: isValid})

// 			if isValid {
// 				t.Logf("属性访问: %s.%s", expr.GetText(), name)
// 			} else {
// 				t.Logf("无效的属性访问: name='%s', expr valid=%t", name, expr.IsValid())
// 			}
// 		}
// 	})

// 	assert.Greater(t, len(accesses), 10, "应该找到多个属性访问")

// 	// 验证属性名的有效性
// 	var validAccesses int
// 	for _, access := range accesses {
// 		if access.valid {
// 			validAccesses++
// 			assert.NotEmpty(t, access.name, "有效的属性访问应该有非空属性名")
// 			assert.NotEmpty(t, access.expr, "有效的属性访问应该有非空对象表达式")
// 		}
// 	}

// 	assert.Greater(t, validAccesses, 5, "应该找到多个有效的属性访问")

// 	// 检查常见属性名
// 	commonProps := map[string]bool{
// 		"prop":     false,
// 		"method":   false,
// 		"property": false,
// 		"value":    false,
// 	}
// 	for _, access := range accesses {
// 		if _, exists := commonProps[access.name]; exists {
// 			commonProps[access.name] = true
// 		}
// 	}

// 	for propName, found := range commonProps {
// 		if found {
// 			t.Logf("找到常见属性: %s", propName)
// 		}
// 	}

// 	t.Logf("属性访问统计: 总数=%d, 有效=%d", len(accesses), validAccesses)
// }

// // TestFunctionDeclaration_Comprehensive 测试FunctionDeclaration专有API的全面功能
// func TestFunctionDeclaration_Comprehensive(t *testing.T) {
// 	project := tsmorphgo.NewProjectFromSources(map[string]string{
// 		"/test.ts": `
// 			// 命名函数声明
// 			function namedFunction() {
// 				return 1;
// 			}

// 			// 带参数的函数
// 			function withParams(a: string, b: number) {
// 				return a + b;
// 			}

// 			// 带返回类型的函数
// 			function withReturnType(): boolean {
// 				return true;
// 			}

// 			// 异步函数
// 			async function asyncFunction(): Promise<string> {
// 				return "async";
// 			}

// 			// 生成器函数
// 			function* generatorFunction(): Iterator<number> {
// 				yield 1;
// 			}

// 			// 方法声明（在类中）
// 			class TestClass {
// 				method() {
// 					return "method";
// 				}
// 			}

// 			// 导出函数
// 			export function exportedFunction() {
// 				return "exported";
// 			}

// 			// 匿名函数表达式（不是函数声明）
// 			const anonymous = function() {
// 				return "anonymous";
// 			};

// 			// 箭头函数（不是函数声明）
// 			const arrow = () => "arrow";
// 		`,
// 	})
// 	defer project.Close()

// 	sf := project.GetSourceFile("/test.ts")
// 	assert.NotNil(t, sf)

// 	var declarations []struct {
// 		name        string
// 		isAnonymous bool
// 	}

// 	sf.ForEachDescendant(func(node tsmorphgo.Node) {
// 		if funcDecl, ok := node.AsFunctionDeclaration(); ok {
// 			nameNode := funcDecl.GetNameNode()
// 			name := funcDecl.GetName()
// 			isAnonymous := funcDecl.IsAnonymous()

// 			// 验证名称节点和名称的一致性
// 			if nameNode.IsValid() {
// 				assert.Equal(t, name, nameNode.GetText(), "名称和名称节点文本应该一致")
// 			}

// 			// 验证匿名函数的逻辑
// 			if isAnonymous {
// 				assert.Empty(t, name, "匿名函数名称应该为空")
// 				assert.False(t, nameNode.IsValid(), "匿名函数名称节点应该无效")
// 			} else {
// 				assert.NotEmpty(t, name, "命名函数名称不应为空")
// 				assert.True(t, nameNode.IsValid(), "命名函数名称节点应该有效")
// 			}

// 			declarations = append(declarations, struct {
// 				name        string
// 				isAnonymous bool
// 			}{name: name, isAnonymous: isAnonymous})

// 			t.Logf("函数声明: name='%s', isAnonymous=%t", name, isAnonymous)
// 		}
// 	})

// 	assert.Greater(t, len(declarations), 0, "应该找到函数声明")

// 	// 验证特定函数的存在
// 	expectedFunctions := []string{
// 		"namedFunction",
// 		"withParams",
// 		"withReturnType",
// 		"asyncFunction",
// 		"generatorFunction",
// 		"method",
// 		"exportedFunction",
// 	}

// 	foundFunctions := make(map[string]bool)
// 	for _, decl := range declarations {
// 		if decl.name != "" {
// 			foundFunctions[decl.name] = true
// 		}
// 	}

// 	for _, expected := range expectedFunctions {
// 		if foundFunctions[expected] {
// 			t.Logf("找到预期函数: %s", expected)
// 		}
// 	}

// 	// 统计命名函数和匿名函数
// 	var namedCount, anonymousCount int
// 	for _, decl := range declarations {
// 		if decl.isAnonymous {
// 			anonymousCount++
// 		} else {
// 			namedCount++
// 		}
// 	}

// 	assert.Greater(t, namedCount, 0, "应该找到命名函数")
// 	t.Logf("函数声明统计: 总数=%d, 命名=%d, 匿名=%d", len(declarations), namedCount, anonymousCount)
// }

// // TestBinaryExpression_Comprehensive 测试BinaryExpression专有API的全面功能
// func TestBinaryExpression_Comprehensive(t *testing.T) {
// 	project := tsmorphgo.NewProjectFromSources(map[string]string{
// 		"/test.ts": `
// 			// 算术运算
// 			const a = x + y;
// 			const b = x - y;
// 			const c = x * y;
// 			const d = x / y;
// 			const e = x % y;
// 			const f = x ** y;

// 			// 比较运算
// 			const g = x === y;
// 			const h = x !== y;
// 			const i = x == y;
// 			const j = x != y;
// 			const k = x > y;
// 			const l = x >= y;
// 			const m = x < y;
// 			const n = x <= y;

// 			// 逻辑运算
// 			const o = x && y;
// 			const p = x || y;
// 			const q = x ?? y;

// 			// 位运算
// 			const r = x & y;
// 			const s = x | y;
// 			const t = x ^ y;
// 			const u = x << y;
// 			const v = x >> y;
// 			const w = x >>> y;

// 			// 赋值运算（虽然是BinaryExpression，但通常不作为这类处理）
// 			// const z = x = y;

// 			// 复杂表达式
// 			const complex = (a + b) * (c - d) / (e + f);
// 		`,
// 	})
// 	defer project.Close()

// 	sf := project.GetSourceFile("/test.ts")
// 	assert.NotNil(t, sf)

// 	var expressions []struct {
// 		operator  string
// 		leftText  string
// 		rightText string
// 	}

// 	sf.ForEachDescendant(func(node tsmorphgo.Node) {
// 		if binaryExpr, ok := node.AsBinaryExpression(); ok {
// 			operatorToken := binaryExpr.GetOperatorToken()
// 			left := binaryExpr.GetLeft()
// 			right := binaryExpr.GetRight()

// 			assert.True(t, operatorToken.IsValid(), "操作符节点应该有效")
// 			assert.True(t, left.IsValid(), "左操作数应该有效")
// 			assert.True(t, right.IsValid(), "右操作数应该有效")

// 			operatorText := operatorToken.GetText()
// 			leftText := left.GetText()
// 			rightText := right.GetText()

// 			expressions = append(expressions, struct {
// 				operator  string
// 				leftText  string
// 				rightText string
// 			}{operator: operatorText, leftText: leftText, rightText: rightText})

// 			t.Logf("二元表达式: %s %s %s", leftText, operatorText, rightText)
// 		}
// 	})

// 	assert.Greater(t, len(expressions), 15, "应该找到多个二元表达式")

// 	// 验证不同类型的操作符
// 	operators := make(map[string]int)
// 	for _, expr := range expressions {
// 		operators[expr.operator]++
// 	}

// 	expectedOperators := []string{"+", "-", "*", "/", "==", "===", "!=", "!==", ">", "<", "&&", "||"}
// 	for _, op := range expectedOperators {
// 		if count, exists := operators[op]; exists && count > 0 {
// 			t.Logf("找到操作符 %s: %d次", op, count)
// 		}
// 	}

// 	t.Logf("二元表达式统计: 总数=%d", len(expressions))
// 	for op, count := range operators {
// 		t.Logf("  %s: %d", op, count)
// 	}
// }

// // TestTypeSafety_Comprehensive 测试类型安全转换的全面性
// func TestTypeSafety_Comprehensive(t *testing.T) {
// 	project := tsmorphgo.NewProjectFromSources(map[string]string{
// 		"/test.ts": `
// 			const x = 42;
// 			function test() { return x; }
// 			obj.method();
// 			const result = a + b;
// 		`,
// 	})
// 	defer project.Close()

// 	sf := project.GetSourceFile("/test.ts")
// 	assert.NotNil(t, sf)

// 	var totalNodes, successfulConversions int
// 	conversionResults := make(map[string]int)

// 	sf.ForEachDescendant(func(node tsmorphgo.Node) {
// 		totalNodes++

// 		// 测试所有类型转换方法的安全性
// 		if _, ok := node.AsVariableDeclaration(); ok {
// 			successfulConversions++
// 			conversionResults["VariableDeclaration"]++
// 		}
// 		if _, ok := node.AsCallExpression(); ok {
// 			successfulConversions++
// 			conversionResults["CallExpression"]++
// 		}
// 		if _, ok := node.AsPropertyAccessExpression(); ok {
// 			successfulConversions++
// 			conversionResults["PropertyAccessExpression"]++
// 		}
// 		if _, ok := node.AsFunctionDeclaration(); ok {
// 			successfulConversions++
// 			conversionResults["FunctionDeclaration"]++
// 		}
// 		if _, ok := node.AsBinaryExpression(); ok {
// 			successfulConversions++
// 			conversionResults["BinaryExpression"]++
// 		}
// 	})

// 	assert.Greater(t, totalNodes, 0, "应该找到节点")
// 	assert.Greater(t, successfulConversions, 0, "应该有成功的类型转换")

// 	t.Logf("类型转换统计:")
// 	t.Logf("  总节点数: %d", totalNodes)
// 	t.Logf("  成功转换: %d", successfulConversions)
// 	t.Logf("  转换结果:")
// 	for conversionType, count := range conversionResults {
// 		t.Logf("    %s: %d", conversionType, count)
// 	}

// 	// 验证每个成功转换都返回了有效的结构体
// 	sf.ForEachDescendant(func(node tsmorphgo.Node) {
// 		if varDecl, ok := node.AsVariableDeclaration(); ok {
// 			assert.NotNil(t, varDecl.GetNode(), "转换后的VariableDeclaration应该有有效的Node")
// 			assert.Equal(t, node.GetKind(), varDecl.GetKind(), "转换后应该保持相同的Kind")
// 		}
// 		if callExpr, ok := node.AsCallExpression(); ok {
// 			assert.NotNil(t, callExpr.GetNode(), "转换后的CallExpression应该有有效的Node")
// 			assert.Equal(t, node.GetKind(), callExpr.GetKind(), "转换后应该保持相同的Kind")
// 		}
// 	})
// }
