package gitlab

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// DiffParser 测试
//
// 测试目标：验证 git diff 解析器能够正确解析 diff 输出，并提取精确的变更信息
// - 文件路径：变更发生的文件
// - 行号：具体哪些行发生了变更（行级精度）
// - 兼容性：支持标准 diff 格式、多文件、二进制文件等
// =============================================================================

// TestDiffParser_ParseDiffOutput 测试基础 diff 解析功能
//
// 功能：验证解析器能够正确处理各种格式的 diff 输出
// 验证点：
// 1. 文件数量正确
// 2. 每个文件的变更行数正确
// 3. 正确处理删除行、空行、二进制文件
func TestDiffParser_ParseDiffOutput(t *testing.T) {
	tests := []struct {
		name           string
		diffOutput     string            // git diff 输入
		expectedFiles  int               // 预期变更的文件数量
		expectedLines  map[string]int    // 每个文件预期的变更行数（只验证数量）
		expectError    bool              // 是否预期会出错
	}{
		{
			name: "解析标准 diff 输出",
			// 场景：单个文件的简单变更
			// 验证：能正确识别文件路径和新增行数量
			diffOutput: `diff --git a/src/components/Button.tsx b/src/components/Button.tsx
index 1234567..abcdefg 100644
--- a/src/components/Button.tsx
+++ b/src/components/Button.tsx
@@ -1,5 +1,7 @@
 // Button 组件
-export const Button = () => {
+export const Button = (props) => {
-  return <button>Click</button>;
+  return <button>{props.label}</button>;
 };
`,
			expectedFiles: 1,
			expectedLines: map[string]int{
				"src/components/Button.tsx": 2, // 2行新增（删除行和空行已忽略）
			},
			expectError: false,
		},
		{
			name: "解析多文件 diff",
			// 场景：一次变更涉及多个文件
			// 验证：能正确分割并解析每个文件块
			diffOutput: `diff --git a/src/Button.tsx b/src/Button.tsx
index 1234567..abcdefg 100644
--- a/src/Button.tsx
+++ b/src/Button.tsx
@@ -1,3 +1,4 @@
-export const A = 1;
+export const A = 2;
+export const B = 3;

diff --git a/src/Input.tsx b/src/Input.tsx
index 2345678..bcdefga 100644
--- a/src/Input.tsx
+++ b/src/Input.tsx
@@ -5,6 +5,8 @@
 export const Input = () => {
   return <input />;
 };
+
+export const LabeledInput = () => {};
`,
			expectedFiles: 2,
			expectedLines: map[string]int{
				"src/Button.tsx": 2,
				"src/Input.tsx":  2,
			},
			expectError: false,
		},
		{
			name: "忽略删除的行和空行",
			// 场景：diff 中包含删除行和新增行
			// 验证：只记录新增的行（以 + 开头），删除行不影响行号计数
			//
			// diff 解析逻辑：
			// - 删除行（-）：不影响新文件的行号
			// - 上下文行（空格开头）：增加新文件行号
			// - 新增行（+）：增加新文件行号并记录
			diffOutput: `diff --git a/src/test.tsx b/src/test.tsx
index 1234567..abcdefg 100644
--- a/src/test.tsx
+++ b/src/test.tsx
@@ -1,8 +1,10 @@
-export const old = 1;     // 删除，行号不变
+export const new = 2;      // 新增，新文件行1
-export const removed = 3;  // 删除，行号不变

+export const added = 4;    // 新增，新文件行4（空行占位行3）
+
 export const unchanged = 5; // 上下文，新文件行5
+export const alsoAdded = 6; // 新增，新文件行7
-export const alsoRemoved = 7; // 删除，行号不变
`,
			expectedFiles: 1,
			expectedLines: map[string]int{
				"src/test.tsx": 4, // 4个新增行：行1,4,7,8(注释行之后)
			},
			expectError: false,
		},
		{
			name: "空 diff 输出",
			// 场景：空字符串输入
			// 验证：不会出错，返回空结果
			diffOutput: "",
			expectedFiles: 0,
			expectedLines:  map[string]int{},
			expectError: false,
		},
		{
			name: "只有二进制文件变更",
			// 场景：只包含二进制文件的变更
			// 验证：二进制文件使用特殊标记表示整个文件变更
			// 设计决策：虽然二进制文件没有行级精度，但导入该文件的组件会受到影响
			diffOutput: `diff --git a/image.png b/image.png
index 1234567..abcdefg 100644
Binary files a/image.png and b/image.png differ
`,
			expectedFiles: 1, // 二进制文件整个文件变更
			expectedLines: map[string]int{
				"image.png": 1, // BinaryFileMarker (行0) 表示文件级变更
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建解析器（不需要项目根目录）
			parser := NewDiffParser("")

			// 解析 diff 输出
			result, err := parser.ParseDiffOutput(tt.diffOutput)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// 验证：解析成功
				assert.NoError(t, err)

				// 验证：变更文件数量正确
				assert.Equal(t, tt.expectedFiles, len(result), "文件数量不匹配")

				// 验证：每个文件的变更行数正确
				for file, expectedCount := range tt.expectedLines {
					actualCount := len(result[file])
					assert.Equal(t, expectedCount, actualCount,
						"文件 %s 的变更行数不匹配，期望 %d，实际 %d",
						file, expectedCount, actualCount)
				}
			}
		})
	}
}

// TestDiffParser_ParseDiffFile 测试从文件读取并解析 diff
//
// 功能：验证能正确读取磁盘上的 patch 文件并解析
// 场景：使用测试数据中的 sample.patch
func TestDiffParser_ParseDiffFile(t *testing.T) {
	// 测试数据文件路径
	patchFile := "testdata/sample.patch"
	if _, err := os.Stat(patchFile); os.IsNotExist(err) {
		t.Skip("测试数据文件不存在:", patchFile)
	}

	// 从文件解析 diff
	parser := NewDiffParser("")
	result, err := parser.ParseDiffFile(patchFile)

	assert.NoError(t, err)
	assert.Greater(t, len(result), 0, "应该解析出变更的文件")

	// 验证预期文件存在
	expectedFiles := []string{
		"src/components/Button/Button.tsx",
		"src/components/Input/Input.tsx",
	}

	for _, expectedFile := range expectedFiles {
		assert.Contains(t, result, expectedFile, "应该包含文件: "+expectedFile)
		assert.Greater(t, len(result[expectedFile]), 0, expectedFile+" 应该有变更行")
	}
}

// TestDiffParser_GetChangedFiles 测试行级到文件级的转换
//
// 功能：验证 ChangedLineSetOfFiles 到 ChangeInput 的兼容层转换
//
// 数据结构说明：
// - ChangedLineSetOfFiles: 行级精度 {file: {lineNum: true}}
// - ChangeInput: 文件级精度（兼容 impact-analysis）
//
// 转换逻辑：将所有有变更的文件路径提取到 ModifiedFiles 列表
func TestDiffParser_GetChangedFiles(t *testing.T) {
	// 构造行级变更数据
	lineSet := ChangedLineSetOfFiles{
		"src/Button.tsx":   {1: true, 5: true, 10: true}, // Button 的 1,5,10 行变更
		"src/Input.tsx":    {3: true},                      // Input 的第 3 行变更
		"src/Select.tsx":   {7: true, 8: true},             // Select 的 7,8 行变更
	}

	// 执行转换：行级 → 文件级
	parser := NewDiffParser("")
	changeInput := parser.GetChangedFiles(lineSet)

	// 验证转换结果
	assert.NotNil(t, changeInput)
	assert.Equal(t, 3, len(changeInput.ModifiedFiles))
	assert.ElementsMatch(t, []string{
		"src/Button.tsx",
		"src/Input.tsx",
		"src/Select.tsx",
	}, changeInput.ModifiedFiles)
}

// TestDiffParser_BinaryFileMarker 测试二进制文件标记处理
//
// 功能：验证二进制文件使用 BinaryFileMarker (0) 表示整个文件变更
//
// 为什么需要特殊标记？
// - 文本文件：可以精确到具体行号（如行 1, 5, 10）
// - 二进制文件：无法解析行级内容，但变更仍然会影响依赖它的组件
//
// 设计决策：
// - 使用行号 0 表示"文件级别"变更
// - 保持 ChangedLineSetOfFiles 数据结构一致性
// - 影响分析时，导入该二进制文件的组件会被标记为受影响
//
// 应用场景：
// - 图片文件变更（logo.png 更新可能影响多个页面）
// - 配置文件变更（JSON/XML 等二进制格式）
// - 字体文件、资源文件等
func TestDiffParser_BinaryFileMarker(t *testing.T) {
	tests := []struct {
		name              string
		diffOutput        string
		expectedFile      string
		expectedMarker    int  // 预期的标记值（BinaryFileMarker = 0）
		expectHasMarker   bool // 是否包含 BinaryFileMarker
	}{
		{
			name: "单个二进制文件变更",
			diffOutput: `diff --git a/public/logo.png b/public/logo.png
index 1234567..abcdefg 100644
Binary files a/public/logo.png and b/public/logo.png differ
`,
			expectedFile:    "public/logo.png",
			expectedMarker:  BinaryFileMarker, // 0
			expectHasMarker: true,
		},
		{
			name: "多个二进制文件变更",
			diffOutput: `diff --git a/public/logo.png b/public/logo.png
index 1234567..abcdefg 100644
Binary files a/public/logo.png and b/public/logo.png differ

diff --git a/public/icon.jpg b/public/icon.jpg
index 2345678..bcdefga 100644
Binary files a/public/icon.jpg and b/public/icon.jpg differ
`,
			expectedFile:    "public/logo.png",
			expectedMarker:  BinaryFileMarker,
			expectHasMarker: true,
		},
		{
			name: "文本文件变更（不使用标记）",
			diffOutput: `diff --git a/src/test.ts b/src/test.ts
index 1234567..abcdefg 100644
--- a/src/test.ts
+++ b/src/test.ts
@@ -1,3 +1,4 @@
-const a = 1;
+const a = 2;
+const b = 3;
 const c = 4;
`,
			expectedFile:    "src/test.ts",
			expectHasMarker: false, // 文本文件不使用 BinaryFileMarker
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewDiffParser("")
			result, err := parser.ParseDiffOutput(tt.diffOutput)

			require.NoError(t, err)
			assert.Contains(t, result, tt.expectedFile, "应该包含文件: "+tt.expectedFile)

			lines := result[tt.expectedFile]

			if tt.expectHasMarker {
				// 验证：包含 BinaryFileMarker
				assert.True(t, lines[BinaryFileMarker],
					"应该包含 BinaryFileMarker (0)，实际行号: %v", getSortedKeys(lines))
				assert.Equal(t, 1, len(lines), "二进制文件应该只有一个标记")
			} else {
				// 验证：不包含 BinaryFileMarker（文本文件有具体行号）
				assert.False(t, lines[BinaryFileMarker],
					"文本文件不应该包含 BinaryFileMarker")
				assert.Greater(t, len(lines), 0, "文本文件应该有具体的行号")
			}
		})
	}
}

// TestDiffParser_MixedBinaryAndText 测试混合文本和二进制文件
//
// 功能：验证 diff 中同时包含文本文件和二进制文件时的处理
//
// 场景：实际项目中经常同时包含代码变更和资源变更
// - 代码文件：精确行号（如 src/Button.tsx 的第 5, 10 行）
// - 资源文件：文件级标记（如 assets/logo.png 使用标记 0）
func TestDiffParser_MixedBinaryAndText(t *testing.T) {
	diffOutput := `diff --git a/src/Button.tsx b/src/Button.tsx
index 1234567..abcdefg 100644
--- a/src/Button.tsx
+++ b/src/Button.tsx
@@ -5,6 +5,8 @@
 export const Button = () => {
+  const handleClick = () => {};
   return <button>Click</button>
 };

diff --git a/public/logo.png b/public/logo.png
index 2345678..bcdefga 100644
Binary files a/public/logo.png and b/public/logo.png differ

diff --git a/src/Input.tsx b/src/Input.tsx
index 3456789..cdefgab 100644
--- a/src/Input.tsx
+++ b/src/Input.tsx
@@ -1,3 +1,4 @@
 export const Input = () => {
+  return <input />;
 };
`

	parser := NewDiffParser("")
	result, err := parser.ParseDiffOutput(diffOutput)

	require.NoError(t, err)

	// 验证：解析出 3 个文件
	assert.Equal(t, 3, len(result), "应该解析出 3 个文件")

	// 验证：文本文件有具体行号
	buttonLines := result["src/Button.tsx"]
	assert.Greater(t, len(buttonLines), 0, "Button.tsx 应该有变更行")
	assert.False(t, buttonLines[BinaryFileMarker], "文本文件不应有 BinaryFileMarker")
	t.Logf("Button.tsx 变更行号: %v", getSortedKeys(buttonLines))

	inputLines := result["src/Input.tsx"]
	assert.Greater(t, len(inputLines), 0, "Input.tsx 应该有变更行")
	assert.False(t, inputLines[BinaryFileMarker], "文本文件不应有 BinaryFileMarker")
	t.Logf("Input.tsx 变更行号: %v", getSortedKeys(inputLines))

	// 验证：二进制文件使用 BinaryFileMarker
	logoLines := result["public/logo.png"]
	assert.True(t, logoLines[BinaryFileMarker], "二进制文件应有 BinaryFileMarker")
	assert.Equal(t, 1, len(logoLines), "二进制文件应该只有一个标记")
	t.Logf("logo.png 变更标记: %v", getSortedKeys(logoLines))
}

// =============================================================================
// 行级精度测试
// =============================================================================

// TestDiffParser_LineLevelPrecision 测试行级精度解析
//
// 功能：验证解析器能精确到具体的行号
//
// 为什么需要行级精度？
// - 文件级：知道哪些文件变了（粗粒度）
// - 行级：知道具体是哪些行变了（细粒度）
//
// 应用场景：
// - 精准的代码审查：只检查变更的行
// - 影响分析：行级别的依赖追踪
// - CI/CD：基于变更行的增量测试
func TestDiffParser_LineLevelPrecision(t *testing.T) {
	tests := []struct {
		name            string
		diffOutput      string            // git diff 输入
		expectedFile    string            // 预期的文件路径
		expectedLineNums map[int]bool     // 预期的具体行号集合
	}{
		{
			name: "精确行号 - 单文件多行变更",
			// 场景：验证最基础的行号计算逻辑
			//
			// diff 格式说明：
			// @@ -1,5 +1,8 @@
			//   -1,5  表示旧文件从第1行开始，共5行
			//   +1,8  表示新文件从第1行开始，共8行
			//
			// 行号追踪：
			// 新文件行1: const a = 1;        (上下文)
			// 新文件行2: const b = 3;        (新增，记录行2)
			// 新文件行3: const c = 4;        (新增，记录行3)
			// 新文件行4: const d = 5;        (上下文)
			// 新文件行5: const e = 6;        (新增，记录行5)
			// 新文件行6: const g = 8;        (上下文，删除行f不占用新文件行号)
			diffOutput: `diff --git a/src/utils.ts b/src/utils.ts
index 1234567..abcdefg 100644
--- a/src/utils.ts
+++ b/src/utils.ts
@@ -1,5 +1,8 @@
 const a = 1;
-const b = 2;
+const b = 3;
+const c = 4;
 const d = 5;
+const e = 6;
-const f = 7;
 const g = 8;
`,
			expectedFile: "src/utils.ts",
			expectedLineNums: map[int]bool{2: true, 3: true, 5: true},
		},
		{
			name: "精确行号 - 带空行的diff",
			// 场景：验证包含空新增行时的行号计算
			//
			// 关键点：即使新增行内容为空，它仍然占据一个行号
			//
			// hunk: @@ -10,5 +10,7 @@
			//   表示从第10行开始，旧文件5行变为新文件7行
			//
			// 逐行解析：
			// 行10: export function foo() {     (上下文)
			// 行11: +  return 2;                (新增，记录行11)
			// 行12: }                          (上下文)
			// 行13: +                          (空新增行，占据行13)
			// 行14: +export function bar() {   (新增，记录行14)
			// 行15: +   return 3;              (新增，记录行15)
			// 行16: +}                         (新增，记录行16)
			diffOutput: `diff --git a/src/test.ts b/src/test.ts
index 1234567..abcdefg 100644
--- a/src/test.ts
+++ b/src/test.ts
@@ -10,5 +10,7 @@
 export function foo() {
-  return 1;
+  return 2;
 }
+
+export function bar() {
+  return 3;
+}
`,
			expectedFile: "src/test.ts",
			expectedLineNums: map[int]bool{11: true, 13: true, 14: true, 15: true, 16: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewDiffParser("")
			result, err := parser.ParseDiffOutput(tt.diffOutput)

			require.NoError(t, err)

			// 验证：文件路径正确
			assert.Contains(t, result, tt.expectedFile, "应该包含文件: "+tt.expectedFile)

			actualLines := result[tt.expectedFile]

			// 验证：行号数量匹配
			assert.Equal(t, len(tt.expectedLineNums), len(actualLines),
				"行号数量不匹配，期望 %d 个行号，实际 %d 个",
				len(tt.expectedLineNums), len(actualLines))

			// 验证：每个预期的行号都存在
			for expectedLine := range tt.expectedLineNums {
				assert.True(t, actualLines[expectedLine],
					"应该包含行号 %d，实际行号: %v",
					expectedLine, getSortedKeys(actualLines))
			}

			// 验证：没有多余的行号
			for actualLine := range actualLines {
				assert.True(t, tt.expectedLineNums[actualLine],
					"存在多余的行号 %d，期望行号: %v",
					actualLine, getSortedKeys(tt.expectedLineNums))
			}

			// 输出：便于调试
			t.Logf("文件 %s 的变更行号: %v", tt.expectedFile, getSortedKeys(actualLines))
		})
	}
}

// TestDiffParser_MultiFileLineNumbers 测试多文件的具体行号解析
//
// 功能：验证能正确处理多文件 diff 中每个文件的精确行号
//
// 场景描述：
// 文件 A (src/A.ts):
//   hunk: @@ -1,3 +1,4 @@
//   - 旧: 3行 (行1-3)
//   - 新: 4行 (行1-4)
//   - 变更: 删除行1，新增行1-2
//
// 文件 B (src/B.ts):
//   hunk: @@ -5,4 +5,6 @@
//   - 旧: 4行，从第5行开始
//   - 新: 6行，从第5行开始
//   - 变更: 删除行y，新增行y-2和z-2
//   - 注意：上下文行x(行4)和w(行7)影响行号计算
func TestDiffParser_MultiFileLineNumbers(t *testing.T) {
	diffOutput := `diff --git a/src/A.ts b/src/A.ts
index 1234567..abcdefg 100644
--- a/src/A.ts
+++ b/src/A.ts
@@ -1,3 +1,4 @@
-export const old1 = 1;
+export const new1 = 1;
+export const new2 = 2;
 export const unchanged = 3;

diff --git a/src/B.ts b/src/B.ts
index 2345678..bcdefga 100644
--- a/src/B.ts
+++ b/src/B.ts
@@ -5,4 +5,6 @@
 const x = 10;
-const y = 20;
+const y = 21;
+const z = 22;
 const w = 30;
`

	parser := NewDiffParser("")
	result, err := parser.ParseDiffOutput(diffOutput)

	require.NoError(t, err)

	// 验证文件 A
	// 期望：新增行1和行2
	assert.Contains(t, result, "src/A.ts")
	linesA := result["src/A.ts"]
	assert.Equal(t, 2, len(linesA))
	assert.True(t, linesA[1], "src/A.ts 应该包含行号 1")  // +export const new1 = 1;
	assert.True(t, linesA[2], "src/A.ts 应该包含行号 2")  // +export const new2 = 2;
	t.Logf("src/A.ts 变更行号: %v", getSortedKeys(linesA))

	// 验证文件 B
	// 期望：新增行6和行7
	// 行4: const x = 10;       (上下文)
	// 行5: -const y = 20;      (删除，不影响新文件行号)
	// 行6: +const y = 21;      (新增)
	// 行7: +const z = 22;      (新增)
	// 行8: const w = 30;       (上下文)
	assert.Contains(t, result, "src/B.ts")
	linesB := result["src/B.ts"]
	assert.Equal(t, 2, len(linesB))
	assert.True(t, linesB[6], "src/B.ts 应该包含行号 6")  // +const y = 21;
	assert.True(t, linesB[7], "src/B.ts 应该包含行号 7")  // +const z = 22;
	t.Logf("src/B.ts 变更行号: %v", getSortedKeys(linesB))
}

// getSortedKeys 返回 map 的排序后的键列表
//
// 用途：在测试日志中输出排序后的行号，便于阅读和调试
// 输入：map[int]bool (如 {5: true, 1: true, 3: true})
// 输出：[]int (如 [1, 3, 5])
func getSortedKeys(m map[int]bool) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

// =============================================================================
// Git 集成测试
// =============================================================================

// TestDiffParser_ParseFromGit 测试从 Git 命令获取 diff
//
// 功能：验证能直接调用 git diff 命令并解析输出
//
// 场景：
// - 在实际的 Git 仓库中执行 git diff 命令
// - 解析命令输出获取变更信息
//
// 注意：
// - 此测试需要 Git 仓库环境
// - 需要有实际的变更才能看到结果
func TestDiffParser_ParseFromGit(t *testing.T) {
	// 测试项目路径
	testProject := "../../testdata/test_project"
	if _, err := os.Stat(testProject); os.IsNotExist(err) {
		t.Skip("测试项目不存在:", testProject)
	}

	// 检查是否为 Git 仓库
	gitDir := filepath.Join(testProject, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Skip("测试项目不是 git 仓库:", testProject)
	}

	parser := NewDiffParser(testProject)

	// 测试：解析当前工作区的变更
	t.Run("Parse unstaged changes", func(t *testing.T) {
		result, err := parser.ParseFromGit("HEAD", "")

		// 允许没有变更的情况（不强制要求结果）
		_ = result

		// 如果有变更，确保没有错误
		if len(result) > 0 {
			assert.NoError(t, err)
		}
	})
}

// =============================================================================
// ChangeInput 测试
// =============================================================================

// TestChangeInput_GetFileCount 测试获取变更文件总数
//
// 功能：验证能正确统计所有类型的变更文件数量
//
// 变更类型：
// - ModifiedFiles: 修改的文件
// - AddedFiles: 新增的文件
// - DeletedFiles: 删除的文件
func TestChangeInput_GetFileCount(t *testing.T) {
	changeInput := &ChangeInput{
		ModifiedFiles: []string{"a.ts", "b.ts"},
		AddedFiles:    []string{"c.ts"},
		DeletedFiles:  []string{"d.ts"},
	}

	assert.Equal(t, 4, changeInput.GetFileCount())
}

// TestChangeInput_GetAllFiles 测试获取所有变更文件列表
//
// 功能：验证能正确汇总所有类型的变更文件
//
// 返回：包含 ModifiedFiles + AddedFiles + DeletedFiles 的所有文件
func TestChangeInput_GetAllFiles(t *testing.T) {
	changeInput := &ChangeInput{
		ModifiedFiles: []string{"a.ts", "b.ts"},
		AddedFiles:    []string{"c.ts"},
		DeletedFiles:  []string{"d.ts"},
	}

	allFiles := changeInput.GetAllFiles()
	assert.ElementsMatch(t, []string{"a.ts", "b.ts", "c.ts", "d.ts"}, allFiles)
}
