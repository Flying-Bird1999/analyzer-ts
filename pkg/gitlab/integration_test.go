package gitlab

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// GitLab Client 测试
//
// 测试目标：验证 GitLab API 客户端的正确性
// 使用 httptest 模拟 GitLab API 服务器，无需真实 GitLab 环境
// =============================================================================

// TestGitLabClient_GetMergeRequestDiff 测试获取 MR diff
func TestGitLabClient_GetMergeRequestDiff(t *testing.T) {
	// 创建模拟的 GitLab API 服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求
		assert.Equal(t, http.MethodGet, r.Method)
		assert.True(t, strings.Contains(r.URL.Path, "/merge_requests"))

		// GitLab API 直接返回 DiffFile 数组，不需要包装在对象中
		diffResponse := []DiffFile{
			{
				OldPath: "src/components/Button.tsx",
				NewPath: "src/components/Button.tsx",
				Diff:    "@@ -1,5 +1,7 @@\n export const Button",
				NewFile: false,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(diffResponse)
	}))
	defer server.Close()

	// 创建客户端并测试
	client := NewClient(server.URL, "test-token")

	diffs, err := client.GetMergeRequestDiff(context.Background(), 123, 456)
	require.NoError(t, err)
	require.Len(t, diffs, 1)

	assert.Equal(t, "src/components/Button.tsx", diffs[0].OldPath)
	assert.Contains(t, diffs[0].Diff, "export const Button")
}

// TestGitLabClient_CreateMRComment 测试创建 MR 评论
func TestGitLabClient_CreateMRComment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.True(t, strings.Contains(r.URL.Path, "/merge_requests"))

		// 验证请求体
		body, _ := io.ReadAll(r.Body)
		var payload map[string]string
		json.Unmarshal(body, &payload)

		assert.Equal(t, "Test comment body", payload["body"])

		w.WriteHeader(http.StatusCreated)
		createResponse := map[string]interface{}{
			"id":      "789",
			"note_id": 123,
		}
		json.NewEncoder(w).Encode(createResponse)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	err := client.CreateMRComment(context.Background(), 123, 456, "Test comment body")
	require.NoError(t, err)
}

// =============================================================================
// DiffProvider 测试
// =============================================================================

// TestDiffProvider_FromFile 测试从文件读取 diff
func TestDiffProvider_FromFile(t *testing.T) {
	// 创建临时 diff 文件
	tmpDir := t.TempDir()
	diffFile := filepath.Join(tmpDir, "changes.patch")

	diffContent := `diff --git a/src/Button.tsx b/src/Button.tsx
index 1234567..abcdefg 100644
--- a/src/Button.tsx
+++ b/src/Button.tsx
@@ -1,5 +1,7 @@
 export const Button = () => {
-  return <button>Click</button>;
+  return <button>New Text</button>;
 }
`
	err := os.WriteFile(diffFile, []byte(diffContent), 0o644)
	require.NoError(t, err)

	// 创建 Parser
	parser := NewParser(tmpDir)

	// 解析 diff
	lineSet, err := parser.ParseDiffFile(diffFile)
	require.NoError(t, err)
	require.NotNil(t, lineSet)

	// 验证结果 - Parser 返回 map[int]bool
	assert.Contains(t, lineSet, "src/Button.tsx")
	// 第2行是新增的
	// @@ -1,5 +1,7 @@ 表示新文件从第1行开始
	//  export const Button = () => { (空格开头的上下文行，第1行)
	// -  return <button>Click</button>; (删除行，不增加行号)
	// +  return <button>New Text</button>; (第2行，新增)
	assert.Equal(t, map[int]bool{2: true}, lineSet["src/Button.tsx"])
}

// TestDiffProvider_FromString 测试从字符串解析 diff
func TestDiffProvider_FromString(t *testing.T) {
	parser := NewParser("")

	diffContent := `diff --git a/src/A.tsx b/src/A.tsx
index 1234567..abcdefg 100644
--- a/src/A.tsx
+++ b/src/A.tsx
@@ -1,3 +1,5 @@
-const a = 1;
+const a = 2;
+const b = 3;
`

	lineSet, err := parser.ParseDiffString(diffContent)
	require.NoError(t, err)

	// 验证：第1行和第2行是新增的
	// @@ -1,3 +1,5 @@ 表示新文件从第1行开始
	// -const a = 1; 是删除行，不增加行号
	// +const a = 2; 是新增的第1行
	// +const b = 3; 是新增的第2行
	assert.Equal(t, map[int]bool{1: true, 2: true}, lineSet["src/A.tsx"])
}

// TestDiffProvider_MultiFile 测试解析多文件 diff
func TestDiffProvider_MultiFile(t *testing.T) {
	parser := NewParser("")

	diffContent := `diff --git a/src/A.tsx b/src/A.tsx
index 1234567..abcdefg 100644
--- a/src/A.tsx
+++ b/src/A.tsx
@@ -1,3 +1,5 @@
 const a = 1;
+const b = 2;
diff --git a/src/B.tsx b/src/B.tsx
index 1234567..abcdefg 100644
--- a/src/B.tsx
+++ b/src/B.tsx
@@ -1,3 +1,4 @@
 const c = 3;
+const d = 4;
`

	lineSet, err := parser.ParseDiffString(diffContent)
	require.NoError(t, err)

	// 验证两个文件都被解析
	assert.Len(t, lineSet, 2)
	assert.Contains(t, lineSet, "src/A.tsx")
	assert.Contains(t, lineSet, "src/B.tsx")
}

// TestDiffProvider_BinaryFile 测试二进制文件
func TestDiffProvider_BinaryFile(t *testing.T) {
	parser := NewParser("")

	diffContent := `diff --git a/src/assets/logo.png b/src/assets/logo.png
index 1234567..abcdefg 100644
Binary files a/src/assets/logo.png and b/src/assets/logo.png differ
`

	lineSet, err := parser.ParseDiffString(diffContent)
	require.NoError(t, err)

	// 二进制文件使用 0 标记
	assert.Equal(t, map[int]bool{0: true}, lineSet["src/assets/logo.png"])
}

// =============================================================================
// 完整流程集成测试
// =============================================================================

// TestEndToEnd_AnalysisFlow 测试完整分析流程
func TestEndToEnd_AnalysisFlow(t *testing.T) {
	// 跳过 CI 环境
	if os.Getenv("CI") == "true" {
		t.Skip("跳过 CI 环境")
	}

	// 1. 准备测试数据
	diffContent := `diff --git a/src/Button.tsx b/src/Button.tsx
index 1234567..abcdefg 100644
--- a/src/Button.tsx
+++ b/src/Button.tsx
@@ -1,5 +1,7 @@
 export const Button = () => {
-  return <button>Click</button>;
+  return <button>{props.label}</button>;
 }`

	// 2. 解析 diff
	parser := NewParser("")
	lineSet, err := parser.ParseDiffString(diffContent)
	require.NoError(t, err)

	// 3. 验证解析结果
	assert.NotEmpty(t, lineSet)
	assert.Contains(t, lineSet, "src/Button.tsx")
	assert.Equal(t, map[int]bool{2: true}, lineSet["src/Button.tsx"])
}

// =============================================================================
// 辅助函数
// =============================================================================

// getTestProjectRoot 获取测试项目根目录
func getTestProjectRoot() string {
	// 从当前文件路径向上查找 testdata 目录
	wd, _ := os.Getwd()
	for {
		testdataPath := filepath.Join(wd, "testdata", "test_project")
		if info, err := os.Stat(testdataPath); err == nil && info.IsDir() {
			return testdataPath
		}

		parent := filepath.Dir(wd)
		if parent == wd {
			break // 到达根目录
		}
		wd = parent
	}
	return ""
}
