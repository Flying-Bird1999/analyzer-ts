package pipeline

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

// =============================================================================
// Pipeline 核心测试
// =============================================================================

// mockStage 模拟阶段
type mockStage struct {
	name          string
	shouldSkip    bool
	executeResult interface{}
	executeError  error
	executed      bool
}

func (m *mockStage) Name() string {
	return m.name
}

func (m *mockStage) Execute(ctx *AnalysisContext) (interface{}, error) {
	m.executed = true
	return m.executeResult, m.executeError
}

func (m *mockStage) Skip(ctx *AnalysisContext) bool {
	return m.shouldSkip
}

// TestAnalysisPipeline_BasicExecution 测试基本执行流程
func TestAnalysisPipeline_BasicExecution(t *testing.T) {
	ctx := context.Background()
	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{RootPath: "."})
	defer project.Close()

	analysisCtx := NewAnalysisContext(ctx, ".", project)

	stage1 := &mockStage{name: "阶段1", executeResult: "result1"}
	stage2 := &mockStage{name: "阶段2", executeResult: "result2"}

	pipe := NewPipeline("测试管道")
	pipe.AddStage(stage1)
	pipe.AddStage(stage2)

	result, err := pipe.Execute(analysisCtx)

	require.NoError(t, err)
	assert.True(t, result.IsSuccessful())
	assert.True(t, stage1.executed, "阶段1应该被执行")
	assert.True(t, stage2.executed, "阶段2应该被执行")
}

// TestAnalysisPipeline_StageSkip 测试阶段跳过
func TestAnalysisPipeline_StageSkip(t *testing.T) {
	ctx := context.Background()
	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{RootPath: "."})
	defer project.Close()

	analysisCtx := NewAnalysisContext(ctx, ".", project)

	stage1 := &mockStage{name: "阶段1", executeResult: "result1"}
	stage2 := &mockStage{name: "阶段2", shouldSkip: true}
	stage3 := &mockStage{name: "阶段3", executeResult: "result3"}

	// 验证初始状态
	assert.False(t, stage2.executed, "阶段2初始状态应该是未执行")

	pipe := NewPipeline("测试管道")
	pipe.AddStage(stage1)
	pipe.AddStage(stage2)
	pipe.AddStage(stage3)

	result, err := pipe.Execute(analysisCtx)

	require.NoError(t, err)
	assert.True(t, result.IsSuccessful())
	assert.True(t, stage1.executed, "阶段1应该被执行")
	assert.False(t, stage2.executed, "阶段2应该被跳过 (executed=%v)", stage2.executed)
	assert.True(t, stage3.executed, "阶段3应该被执行")

	// 检查结果中是否有跳过的阶段
	_, hasSkipped := result.GetResult("阶段2")
	assert.False(t, hasSkipped, "跳过的阶段不应该有结果")
}

// TestAnalysisPipeline_StageError 测试阶段错误处理
func TestAnalysisPipeline_StageError(t *testing.T) {
	ctx := context.Background()
	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{RootPath: "."})
	defer project.Close()

	analysisCtx := NewAnalysisContext(ctx, ".", project)

	stage1 := &mockStage{name: "阶段1", executeResult: "result1"}
	stage2 := &mockStage{name: "阶段2", executeError: assert.AnError}
	stage3 := &mockStage{name: "阶段3", executeResult: "result3"}

	pipe := NewPipeline("测试管道")
	pipe.AddStage(stage1)
	pipe.AddStage(stage2)
	pipe.AddStage(stage3)

	result, err := pipe.Execute(analysisCtx)

	// 当阶段失败时，pipeline 返回 nil 和 error
	assert.Error(t, err)
	assert.Nil(t, result, "失败时结果应该是 nil")
	assert.True(t, stage1.executed, "阶段1应该被执行")
	assert.True(t, stage2.executed, "阶段2应该被执行并失败")
	assert.False(t, stage3.executed, "阶段3不应该被执行（前面阶段失败）")
}

// TestAnalysisPipeline_GetResult 测试获取结果
func TestAnalysisPipeline_GetResult(t *testing.T) {
	ctx := context.Background()
	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{RootPath: "."})
	defer project.Close()

	analysisCtx := NewAnalysisContext(ctx, ".", project)

	expectedResult := map[string]string{"key": "value"}
	stage1 := &mockStage{name: "阶段1", executeResult: expectedResult}

	pipe := NewPipeline("测试管道")
	pipe.AddStage(stage1)

	result, err := pipe.Execute(analysisCtx)
	require.NoError(t, err)

	// 通过 PipelineResult 获取结果
	retrievedResult, exists := result.GetResult("阶段1")
	assert.True(t, exists, "应该找到阶段1的结果")
	assert.Equal(t, expectedResult, retrievedResult)
}

// =============================================================================
// AnalysisContext 测试
// =============================================================================

// TestAnalysisContext_Options 测试配置选项
func TestAnalysisContext_Options(t *testing.T) {
	ctx := context.Background()
	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{RootPath: "."})
	defer project.Close()

	analysisCtx := NewAnalysisContext(ctx, ".", project)

	// 测试设置和获取选项
	analysisCtx.SetOption("projectID", 123)
	analysisCtx.SetOption("mrIID", 456)

	assert.Equal(t, 123, analysisCtx.GetOption("projectID", 0))
	assert.Equal(t, 456, analysisCtx.GetOption("mrIID", 0))
	assert.Equal(t, "default", analysisCtx.GetOption("nonexistent", "default"))
}

// TestAnalysisContext_Results 测试中间结果存储
func TestAnalysisContext_Results(t *testing.T) {
	ctx := context.Background()
	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{RootPath: "."})
	defer project.Close()

	analysisCtx := NewAnalysisContext(ctx, ".", project)

	// 测试存储和获取结果
	result1 := map[string]int{"count": 42}
	result2 := []string{"a", "b", "c"}

	analysisCtx.SetResult("stage1", result1)
	analysisCtx.SetResult("stage2", result2)

	// 验证获取结果
	retrieved1, exists := analysisCtx.GetResult("stage1")
	assert.True(t, exists)
	assert.Equal(t, result1, retrieved1)

	retrieved2, exists := analysisCtx.GetResult("stage2")
	assert.True(t, exists)
	assert.Equal(t, result2, retrieved2)

	// 验证不存在的结果
	_, exists = analysisCtx.GetResult("nonexistent")
	assert.False(t, exists)

	// 测试 MustGetResult
	assert.Equal(t, result1, analysisCtx.MustGetResult("stage1"))

	// 测试 MustGetResult panic
	assert.Panics(t, func() {
		analysisCtx.MustGetResult("nonexistent")
	})
}

// TestAnalysisContext_Cancellation 测试取消信号
func TestAnalysisContext_Cancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{RootPath: "."})
	defer project.Close()

	analysisCtx := NewAnalysisContext(ctx, ".", project)

	// 未取消时
	assert.False(t, analysisCtx.IsCanceled())

	// 取消后
	cancel()
	assert.True(t, analysisCtx.IsCanceled())
}

// =============================================================================
// DiffParserStage 测试
// =============================================================================

// mockGitLabClient 模拟 GitLab 客户端
type mockGitLabClient struct {
	diffFiles []DiffFile
	err       error
}

func (m *mockGitLabClient) GetMergeRequestDiff(ctx context.Context, projectID, mrIID int) ([]DiffFile, error) {
	return m.diffFiles, m.err
}

// TestDiffParserStage_ParseDiffFile 测试解析 diff 文件
func TestDiffParserStage_ParseDiffFile(t *testing.T) {
	// 创建测试 diff 文件
	diffContent := `diff --git a/src/test.ts b/src/test.ts
index 1234567..abcdef 100644
--- a/src/test.ts
+++ b/src/test.ts
@@ -1,3 +1,4 @@
 export function test() {
+  console.log("added");
   return true;
 }
`

	tmpDir := t.TempDir()
	diffFile := tmpDir + "/test.patch"
	require.NoError(t, writeFile(diffFile, diffContent))

	ctx := context.Background()
	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{RootPath: tmpDir})
	defer project.Close()

	analysisCtx := NewAnalysisContext(ctx, tmpDir, project)

	stage := NewDiffParserStage(
		nil, // 不需要 client
		DiffSourceFile,
		diffFile,
		"",
		tmpDir,
		0,
		0,
	)

	result, err := stage.Execute(analysisCtx)

	require.NoError(t, err)
	lineSet, ok := result.(map[string]map[int]bool)
	require.True(t, ok)

	// 验证解析结果
	assert.Contains(t, lineSet, "src/test.ts")
	// hunk 头是 @@ -1,3 +1,4 @@ 表示新文件从第1行开始
	// + 开头的行是第2行（在 hunk 内计数）
	// 实际行号是 2 (hunk 起始位置1 + 偏移量1)
	assert.Contains(t, lineSet["src/test.ts"], 2)
}

// TestDiffParserStage_ParseFromGit 测试从 git 解析
func TestDiffParserStage_ParseFromGit(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要 git 的测试")
	}

	// 这个测试需要在一个 git 仓库中运行
	// 跳过如果没有在 git 仓库中
	// 实际项目中应该创建临时 git 仓库进行测试
	t.Skip("需要设置 git 仓库环境")
}

// TestDiffParserStage_API 测试从 API 获取
func TestDiffParserStage_API(t *testing.T) {
	mockClient := &mockGitLabClient{
		diffFiles: []DiffFile{
			{
				Diff:    "@@ -1,3 +1,4 @@\n export function test() {\n+  console.log(\"added\");\n   return true;\n }\n",
				OldPath: "src/test.ts",
				NewPath: "src/test.ts",
			},
		},
	}

	ctx := context.Background()
	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{RootPath: "."})
	defer project.Close()

	analysisCtx := NewAnalysisContext(ctx, ".", project)

	stage := NewDiffParserStage(
		mockClient,
		DiffSourceAPI,
		"",
		"",
		".",
		123,
		456,
	)

	result, err := stage.Execute(analysisCtx)

	require.NoError(t, err)
	lineSet, ok := result.(map[string]map[int]bool)
	require.True(t, ok)

	assert.Contains(t, lineSet, "src/test.ts")
}

// TestDiffSourceType 测试 DiffSourceType
func TestDiffSourceType(t *testing.T) {
	tests := []struct {
		name   string
		source DiffSourceType
		valid  bool
	}{
		{"auto", DiffSourceAuto, true},
		{"file", DiffSourceFile, true},
		{"api", DiffSourceAPI, true},
		{"diff", DiffSourceSHA, true},
		{"invalid", DiffSourceType("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.source == DiffSourceAuto ||
				tt.source == DiffSourceFile ||
				tt.source == DiffSourceAPI ||
				tt.source == DiffSourceSHA
			assert.Equal(t, tt.valid, valid)
		})
	}
}

// =============================================================================
// 辅助函数
// =============================================================================

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
