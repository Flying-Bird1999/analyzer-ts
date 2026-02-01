package symbol_analysis

import (
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

// =============================================================================
// 复杂场景测试 (Complex Scenario Tests)
// =============================================================================
// 这些测试验证在复杂真实场景中的符号识别能力
// 特别是：当变更发生在声明内部时，能否正确识别其祖先声明节点

// TestAnalyzer_ComplexScenarios 测试各种复杂真实场景
func TestAnalyzer_ComplexScenarios(t *testing.T) {
	sources := map[string]string{
		"/src/complex.ts": `
// 多行数组声明
export const CONFIGURATION = [
	{
		key: 'value1',
		nested: {
			items: ['a', 'b', 'c']
		}
	},
	{
		key: 'value2',
		nested: {
			items: ['d', 'e', 'f']
		}
	},
]

// 多参数的长函数
export function processUserData(
	userName: string,
	userAge: number,
	userEmail: string,
	userAddress: string,
	userPhone: string,
): Promise<UserData> {
	// Line 28 - 函数内部的变更
	const validated = validateUser(userName, userEmail)
	const profile = await fetchProfile(userAge)
	return {
		name: validated.name,
		age: profile.age,
	}
}

// React 组件与 useEffect
export default function UserProfileComponent(props: Props) {
	const [data, setData] = useState(null)
	const [loading, setLoading] = useState(false)

	// Line 50 - useEffect 内部的变更
	useEffect(() => {
		const fetchUser = async () => {
			setLoading(true)
			const result = await api.getUser(props.userId)
			setData(result)
			setLoading(false)
		}
		fetchUser()
	}, [props.userId])

	if (loading) return <Spinner />

	return (
		<div className="profile">
			<h1>{data?.name}</h1>
			<p>{data?.email}</p>
		</div>
	)
}

// 带回调的嵌套函数
export function withRetry<T>(
	operation: () => Promise<T>,
	maxRetries: number = 3,
): Promise<T> {
	return new Promise((resolve, reject) => {
		let attempts = 0

		const execute = async () => {
			try {
				// Line 85 - 回调内部的变更
				const result = await operation()
				resolve(result)
			} catch (error) {
				attempts++
				if (attempts >= maxRetries) {
					reject(error)
				} else {
					// Line 92 - 重试逻辑中的变更
					setTimeout(execute, 1000 * attempts)
				}
			}
		}

		execute()
	})
}

// 多个方法的类
export class DataService {
	private config: Configuration
	private cache: Map<string, any>

	constructor(config: Configuration) {
		this.config = config
		this.cache = new Map()
	}

	// Line 110 - 方法内部的变更
	async fetchData(endpoint: string): Promise<Data> {
		if (this.cache.has(endpoint)) {
			return this.cache.get(endpoint)
		}

		const response = await fetch(endpoint, {
			headers: this.buildHeaders(),
		})

		const data = await response.json()
		this.cache.set(endpoint, data)
		return data
	}

	private buildHeaders(): HeadersInit {
		return {
			'Content-Type': 'application/json',
			'Authorization': "Bearer " + this.config.token,
		}
	}
}
`,
	}

	project := tsmorphgo.NewProjectFromSources(sources)
	analyzer := NewAnalyzerWithDefaults(project)

	tests := []struct {
		name           string
		changedLines   map[int]bool
		expectedSymbol string
		expectedKind   SymbolKind
	}{
		{
			name: "多行数组内部的变更",
			changedLines: map[int]bool{
				6: true, // CONFIGURATION 数组内部
			},
			expectedSymbol: "CONFIGURATION",
			expectedKind:   SymbolKindVariable,
		},
		{
			name: "长函数内部的变更",
			changedLines: map[int]bool{
				28: true, // processUserData 函数内部
			},
			expectedSymbol: "processUserData",
			expectedKind:   SymbolKindFunction,
		},
		{
			name: "useEffect 回调内部的变更",
			changedLines: map[int]bool{
				50: true, // useEffect 回调内部
			},
			expectedSymbol: "UserProfileComponent",
			expectedKind:   SymbolKindFunction,
		},
		{
			name: "嵌套回调内部的变更",
			changedLines: map[int]bool{
				85: true, // execute 回调内部
			},
			expectedSymbol: "withRetry",
			expectedKind:   SymbolKindFunction,
		},
		{
			name: "类方法内部的变更",
			changedLines: map[int]bool{
				110: true, // fetchData 方法内部
			},
			expectedSymbol: "DataService",
			expectedKind:   SymbolKindClass,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeFile("/src/complex.ts", tt.changedLines)
			if err != nil {
				t.Fatalf("预期没有错误，但得到: %v", err)
			}

			if len(result.AffectedSymbols) == 0 {
				t.Errorf("预期找到受影响的符号 %s，但没有找到任何符号", tt.expectedSymbol)
				return
			}

			// 检查是否找到预期的符号
			found := false
			for _, symbol := range result.AffectedSymbols {
				if symbol.Name == tt.expectedSymbol {
					found = true
					if symbol.Kind != tt.expectedKind {
						t.Errorf("预期类型 %s，但得到 %s", tt.expectedKind, symbol.Kind)
					}
					break
				}
			}

			if !found {
				t.Errorf("预期找到符号 %s，但得到: %v", tt.expectedSymbol, result.AffectedSymbols)
			}

			// 验证所有找到的符号都是顶层声明
			for _, symbol := range result.AffectedSymbols {
				if symbol.Kind == SymbolKindVariable {
					// 变量应该只在顶层
					if symbol.StartLine > 10 { // 内部变量在第 10 行之后
						t.Errorf("找到内部变量: %s (行 %d)", symbol.Name, symbol.StartLine)
					}
				}
			}
		})
	}
}

// TestAnalyzer_MultipleChangesInSameSymbol 测试同一符号的多行变更
func TestAnalyzer_MultipleChangesInSameSymbol(t *testing.T) {
	sources := map[string]string{
		"/src/test.ts": `
export function complexFunction(a: string, b: number): boolean {
	const x = a + b
	const y = x * 2
	const z = y / 3
	return z > 0
}
`,
	}

	project := tsmorphgo.NewProjectFromSources(sources)
	analyzer := NewAnalyzerWithDefaults(project)

	// 同一函数内的多行变更
	changedLines := map[int]bool{
		4: true,
		5: true,
		6: true,
	}

	result, err := analyzer.AnalyzeFile("/src/test.ts", changedLines)
	if err != nil {
		t.Fatalf("预期没有错误，但得到: %v", err)
	}

	// 应该只找到一个符号（函数）
	if len(result.AffectedSymbols) != 1 {
		t.Errorf("预期 1 个受影响符号，但得到 %d", len(result.AffectedSymbols))
	}

	symbol := result.AffectedSymbols[0]
	if symbol.Name != "complexFunction" {
		t.Errorf("预期符号名称 'complexFunction'，但得到 '%s'", symbol.Name)
	}

	if symbol.Kind != SymbolKindFunction {
		t.Errorf("预期类型 Function，但得到 %s", symbol.Kind)
	}

	// 应该有 3 行变更
	if len(symbol.ChangedLines) != 3 {
		t.Errorf("预期 3 行变更，但得到 %d", len(symbol.ChangedLines))
	}
}

// TestAnalyzer_ChangesInMultipleSymbols 测试多个符号的变更
// TODO by bird: "2: true" 为何会影响config？
func TestAnalyzer_ChangesInMultipleSymbols(t *testing.T) {
	sources := map[string]string{
		"/src/test.ts": `
export const config = { key: 'value' }

export function helper() {
	return 'help'
}

export class Service {
	method() {}
}
`,
	}

	project := tsmorphgo.NewProjectFromSources(sources)
	analyzer := NewAnalyzerWithDefaults(project)

	// 多个符号中的变更
	changedLines := map[int]bool{
		2: true, // config
		4: true, // helper
		8: true, // Service.method
	}

	result, err := analyzer.AnalyzeFile("/src/test.ts", changedLines)
	if err != nil {
		t.Fatalf("预期没有错误，但得到: %v", err)
	}

	// 应该找到 3 个符号（注意：方法不会单独列出，其父类会被识别）
	if len(result.AffectedSymbols) != 3 {
		t.Errorf("预期 3 个受影响符号，但得到 %d: %v", len(result.AffectedSymbols), result.AffectedSymbols)
	}

	names := make(map[string]bool)
	for _, symbol := range result.AffectedSymbols {
		names[symbol.Name] = true
	}

	expected := []string{"config", "helper", "Service"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("预期找到符号 '%s'", name)
		}
	}
}

// TestAnalyzer_ExportedAncestorTracking 测试导出状态的祖先追踪
// 验证：当变更发生在声明内部时，能否正确追踪其父节点的导出状态
func TestAnalyzer_ExportedAncestorTracking(t *testing.T) {
	sources := map[string]string{
		"/src/test.ts": `
export const publicConfig = { key: 'value' }

const privateConfig = { key: 'private' }

export function publicFunction() {
	const internal = 'local'
	return internal
}

function privateFunction() {
	return 'private'
}

export default class PublicClass {
	method() {}
}

class PrivateClass {
	method() {}
}

export { privateConfig }
`,
	}

	project := tsmorphgo.NewProjectFromSources(sources)
	analyzer := NewAnalyzerWithDefaults(project)

	tests := []struct {
		name             string
		changedLine      int
		expectedSymbol   string
		expectedExported bool
	}{
		{
			name:             "导出的常量中的变更",
			changedLine:      2, // publicConfig 变量声明
			expectedSymbol:   "publicConfig",
			expectedExported: true,
		},
		{
			name:             "私有常量中的变更",
			changedLine:      4, // privateConfig 变量声明
			expectedSymbol:   "privateConfig",
			expectedExported: true,
		},
		{
			name:             "导出函数内部的变更",
			changedLine:      6, // publicFunction 内部的变量
			expectedSymbol:   "publicFunction",
			expectedExported: true,
		},
		{
			name:             "私有函数内部的变更",
			changedLine:      11, // privateFunction 内部的 return
			expectedSymbol:   "privateFunction",
			expectedExported: false,
		},
		{
			name:             "导出类内部的变更",
			changedLine:      16, // PublicClass 内部的方法
			expectedSymbol:   "PublicClass",
			expectedExported: true,
		},
		{
			name:             "私有类内部的变更",
			changedLine:      20, // PrivateClass 内部的方法
			expectedSymbol:   "PrivateClass",
			expectedExported: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changedLines := map[int]bool{tt.changedLine: true}
			result, err := analyzer.AnalyzeFile("/src/test.ts", changedLines)
			if err != nil {
				t.Fatalf("预期没有错误，但得到: %v", err)
			}

			if len(result.AffectedSymbols) == 0 {
				t.Errorf("预期找到受影响的符号，但没有找到")
				return
			}

			symbol := result.AffectedSymbols[0]
			if symbol.Name != tt.expectedSymbol {
				t.Errorf("预期符号 '%s'，但得到 '%s'", tt.expectedSymbol, symbol.Name)
			}

			if symbol.IsExported != tt.expectedExported {
				t.Errorf("预期 IsExported=%v，但得到 %v", tt.expectedExported, symbol.IsExported)
			}
		})
	}
}
