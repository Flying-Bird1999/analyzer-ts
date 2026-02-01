package gitlab

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 集成测试 - 完整流程
// =============================================================================

// TestGitLabIntegration_EndToEnd 端到端集成测试
func TestGitLabIntegration_EndToEnd(t *testing.T) {
	// 跳过 CI 环境（没有 GitLab Token）
	if os.Getenv("CI") == "true" {
		t.Skip("跳过 CI 环境测试")
	}

	testProject := "../../testdata/test_project"
	if _, err := os.Stat(testProject); os.IsNotExist(err) {
		t.Skip("测试项目不存在:", testProject)
	}

	t.Run("完整流程：从 diff 到影响分析", func(t *testing.T) {
		// 1. 准备配置
		config := &GitLabConfig{
			DiffSource:   "file",
			DiffFile:     "testdata/sample.patch",
			MaxDepth:     10,
			ManifestPath: filepath.Join(testProject, ".analyzer", "component-manifest.json"),
		}

		integration := NewGitLabIntegration(config)

		// 2. 测试 diff 解析
		t.Run("步骤1: 解析 diff", func(t *testing.T) {
			changeInput, err := integration.getChangeInput(context.Background())
			require.NoError(t, err)
			assert.NotNil(t, changeInput)
			assert.Greater(t, changeInput.GetFileCount(), 0, "应该检测到变更文件")
		})

		// 3. 测试组件依赖分析（使用预生成的数据）
		t.Run("步骤2: 组件依赖分析", func(t *testing.T) {
			// 先运行 component-deps-v2 生成依赖数据
			ctx := context.Background()
			depData, err := integration.runComponentDepsV2(ctx, testProject)
			require.NoError(t, err, "组件依赖分析应该成功")
			assert.NotNil(t, depData)
			assert.Greater(t, depData.Meta.ComponentCount, 0, "应该有组件")
			assert.NotEmpty(t, depData.DepGraph, "依赖图不应为空")
			assert.NotEmpty(t, depData.RevDepGraph, "反向依赖图不应为空")

			// 验证预期的组件存在
			expectedComponents := []string{"Button", "Input", "Select"}
			for _, comp := range expectedComponents {
				assert.Contains(t, depData.DepGraph, comp, "应该包含组件: "+comp)
			}
		})

		// 4. 测试影响分析
		t.Run("步骤3: 影响分析", func(t *testing.T) {
			ctx := context.Background()

			// 首先获取依赖数据
			depData, err := integration.runComponentDepsV2(ctx, testProject)
			require.NoError(t, err)

			// 从 diff 文件获取变更
			changeInput, err := integration.getChangeInput(ctx)
			require.NoError(t, err)

			// 运行影响分析
			impactResult, err := integration.runImpactAnalysis(ctx, changeInput, depData)
			require.NoError(t, err, "影响分析应该成功")
			assert.NotNil(t, impactResult)

			// 验证结果结构
			assert.NotEmpty(t, impactResult.Meta.AnalyzedAt, "分析时间不应为空")
			assert.Greater(t, impactResult.Meta.ComponentCount, 0, "组件总数应大于0")
			assert.Greater(t, impactResult.Meta.ChangedFileCount, 0, "变更文件数应大于0")

			// 验证变更和影响
			assert.NotEmpty(t, impactResult.Changes, "应该有变更组件")
			assert.NotEmpty(t, impactResult.Impact, "应该有受影响组件")
		})
	})
}

// TestGitLabIntegration_ComponentDepsV2 测试组件依赖分析
func TestGitLabIntegration_ComponentDepsV2(t *testing.T) {
	testProject := "../../testdata/test_project"
	if _, err := os.Stat(testProject); os.IsNotExist(err) {
		t.Skip("测试项目不存在:", testProject)
	}

	// 检查 manifest 文件是否存在
	manifestPath := filepath.Join(testProject, ".analyzer", "component-manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Skip("component-manifest.json 不存在:", manifestPath)
	}

	t.Run("直接运行 component-deps-v2", func(t *testing.T) {
		config := &GitLabConfig{
			ManifestPath: manifestPath,
			DepsFile:     "", // 不使用预生成文件，直接分析
		}

		integration := NewGitLabIntegration(config)
		ctx := context.Background()

		depData, err := integration.runComponentDepsV2(ctx, testProject)
		require.NoError(t, err)
		assert.NotNil(t, depData)

		// 验证依赖图结构（验证组件存在，但不验证具体依赖关系）
		assert.Contains(t, depData.DepGraph, "Button", "应该有 Button 组件")
		assert.Contains(t, depData.DepGraph, "Input", "应该有 Input 组件")
		assert.Contains(t, depData.DepGraph, "Select", "应该有 Select 组件")
		assert.Greater(t, depData.Meta.ComponentCount, 0, "应该有组件")

		// 注意：具体的依赖关系取决于测试项目的实际 import 语句
		// 如果测试项目没有正确的 import，依赖关系可能为空
		t.Logf("Button dependencies: %v", depData.DepGraph["Button"])
		t.Logf("Input dependencies: %v", depData.DepGraph["Input"])
		t.Logf("Select dependencies: %v", depData.DepGraph["Select"])
	})
}

// TestGitLabIntegration_ImpactAnalysis 测试影响分析
func TestGitLabIntegration_ImpactAnalysis(t *testing.T) {
	testProject := "../../testdata/test_project"
	if _, err := os.Stat(testProject); os.IsNotExist(err) {
		t.Skip("测试项目不存在:", testProject)
	}

	t.Run("分析 Button 组件变更的影响", func(t *testing.T) {
		config := &GitLabConfig{
			ManifestPath: filepath.Join(testProject, ".analyzer", "component-manifest.json"),
			MaxDepth:     10,
		}

		integration := NewGitLabIntegration(config)
		ctx := context.Background()

		// 1. 获取依赖数据
		depData, err := integration.runComponentDepsV2(ctx, testProject)
		require.NoError(t, err)

		// 2. 模拟 Button 组件变更
		changeInput := &ChangeInput{
			ModifiedFiles: []string{
				"src/components/Button/Button.tsx",
				"src/components/Button/index.tsx",
			},
			AddedFiles:   []string{},
			DeletedFiles: []string{},
		}

		// 3. 运行影响分析
		impactResult, err := integration.runImpactAnalysis(ctx, changeInput, depData)
		require.NoError(t, err)
		assert.NotNil(t, impactResult)

		// 验证基本结构
		assert.NotEmpty(t, impactResult.Meta.AnalyzedAt, "分析时间不应为空")
		assert.Greater(t, impactResult.Meta.ComponentCount, 0, "组件总数应大于0")
		assert.Equal(t, 2, impactResult.Meta.ChangedFileCount, "变更文件数应为2")

		// 验证变更组件列表存在
		assert.NotEmpty(t, impactResult.Changes, "应该有变更组件")

		// 验证影响列表存在
		assert.NotEmpty(t, impactResult.Impact, "应该有受影响组件")

		// 记录影响分析结果（用于调试）
		t.Logf("Changed components: %d", len(impactResult.Changes))
		for _, change := range impactResult.Changes {
			t.Logf("  - %s: %s", change.Name, change.Action)
		}
		t.Logf("Affected components: %d", len(impactResult.Impact))
		for _, impact := range impactResult.Impact {
			t.Logf("  - %s: level=%d, risk=%s", impact.Name, impact.ImpactLevel, impact.RiskLevel)
		}

		// 注意：具体的影响范围取决于测试项目的实际依赖关系
		// 如果测试项目的组件之间没有 import 依赖关系，影响范围会很小
	})
}

// TestGitLabIntegration_PreloadDepsFile 测试使用预生成依赖文件
func TestGitLabIntegration_PreloadDepsFile(t *testing.T) {
	testProject := "../../testdata/test_project"
	if _, err := os.Stat(testProject); os.IsNotExist(err) {
		t.Skip("测试项目不存在:", testProject)
	}

	// 首先生成依赖数据文件
	t.Run("生成依赖数据文件", func(t *testing.T) {
		config := &GitLabConfig{
			ManifestPath: filepath.Join(testProject, ".analyzer", "component-manifest.json"),
		}

		integration := NewGitLabIntegration(config)
		ctx := context.Background()

		depData, err := integration.runComponentDepsV2(ctx, testProject)
		require.NoError(t, err)

		// 保存到临时文件
		tmpDir := t.TempDir()
		depsFile := filepath.Join(tmpDir, "deps-data.json")

		wrappedData := map[string]interface{}{"component-deps-v2": depData}
		data, err := json.MarshalIndent(wrappedData, "", "  ")
		require.NoError(t, err)
		err = os.WriteFile(depsFile, data, 0644)
		require.NoError(t, err)

		// 使用预生成的文件进行分析
		t.Run("使用预生成文件", func(t *testing.T) {
			config2 := &GitLabConfig{
				DepsFile: depsFile,
				MaxDepth: 10,
			}

			integration2 := NewGitLabIntegration(config2)

			depData2, err := integration2.runComponentDepsV2(ctx, testProject)
			require.NoError(t, err)
			assert.NotNil(t, depData2)

			// 验证数据一致性
			assert.Equal(t, depData.Meta.ComponentCount, depData2.Meta.ComponentCount)
			assert.Equal(t, len(depData.DepGraph), len(depData2.DepGraph))
		})
	})
}

// TestGitLabIntegration_Formatter 测试格式化器
func TestGitLabIntegration_Formatter(t *testing.T) {
	t.Run("格式化影响分析结果", func(t *testing.T) {
		result := &ImpactAnalysisResult{
			Meta: ImpactMeta{
				AnalyzedAt:      "2024-01-31T12:00:00+08:00",
				ComponentCount:  3,
				ChangedFileCount: 2,
				ChangeSource:    "test",
			},
			Changes: []ComponentChange{
				{
					Name:         "Button",
					Action:       "modified",
					ChangedFiles: []string{"src/Button.tsx"},
				},
			},
			Impact: []ImpactComponent{
				{
					Name:        "Button",
					ImpactLevel: 0,
					RiskLevel:   "low",
					ChangePaths: []string{"Button"},
				},
				{
					Name:        "Input",
					ImpactLevel: 1,
					RiskLevel:   "medium",
					ChangePaths: []string{"Input → Button"},
				},
			},
			Recommendations: []Recommendation{
				{
					Type:        "test",
					Priority:    "medium",
					Description: "建议补充单元测试",
				},
			},
		}

		formatter := NewFormatter(CommentStyleDetailed)
		markdown, err := formatter.FormatImpactResult(result)
		require.NoError(t, err)
		assert.NotEmpty(t, markdown)

		// 验证 Markdown 包含关键内容
		assert.Contains(t, markdown, "代码影响分析报告")
		assert.Contains(t, markdown, "概要")
		assert.Contains(t, markdown, "Button")
		assert.Contains(t, markdown, "medium")
		assert.Contains(t, markdown, "建议")
	})

	t.Run("紧凑模式", func(t *testing.T) {
		result := &ImpactAnalysisResult{
			Meta: ImpactMeta{
				AnalyzedAt:      "2024-01-31T12:00:00+08:00",
				ComponentCount:  3,
				ChangedFileCount: 2,
				ChangeSource:    "test",
			},
			Changes: []ComponentChange{
				{Name: "Button", Action: "modified", ChangedFiles: []string{"src/Button.tsx"}},
			},
			Impact: []ImpactComponent{
				{Name: "Button", ImpactLevel: 0, RiskLevel: "low", ChangePaths: []string{"Button"}},
				{Name: "Input", ImpactLevel: 1, RiskLevel: "critical", ChangePaths: []string{"Input → Button"}},
			},
			Recommendations: []Recommendation{},
		}

		formatter := NewFormatter(CommentStyleCompact)
		markdown := formatter.FormatSummary(result)

		assert.Contains(t, markdown, "代码影响分析")
		assert.Contains(t, markdown, "变更组件")
		assert.Contains(t, markdown, "严重风险")
	})
}

// TestChangeInput_Serialization 测试 ChangeInput 序列化
func TestChangeInput_Serialization(t *testing.T) {
	changeInput := &ChangeInput{
		ModifiedFiles: []string{"src/A.tsx", "src/B.tsx"},
		AddedFiles:    []string{"src/C.tsx"},
		DeletedFiles:  []string{"src/D.tsx"},
	}

	// 测试序列化
	data, err := json.Marshal(changeInput)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// 测试反序列化
	var decoded ChangeInput
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, changeInput.ModifiedFiles, decoded.ModifiedFiles)
	assert.Equal(t, changeInput.AddedFiles, decoded.AddedFiles)
	assert.Equal(t, changeInput.DeletedFiles, decoded.DeletedFiles)
}

// TestDetectConfigFromFlags 测试配置检测
func TestDetectConfigFromFlags(t *testing.T) {
	// 这个测试需要模拟 cobra.Command，暂时跳过
	t.Skip("需要 cobra.Command 模拟")
}

// TestValidateConfig 测试配置验证
func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *GitLabConfig
		expectError bool
	}{
		{
			name: "完整配置",
			config: &GitLabConfig{
				URL:        "https://gitlab.example.com",
				Token:      "test-token",
				ProjectID:  123,
				MRIID:      456,
				DiffSource: "auto",
			},
			expectError: false,
		},
		{
			name: "缺少 URL",
			config: &GitLabConfig{
				Token:     "test-token",
				ProjectID: 123,
				MRIID:     456,
			},
			expectError: true,
		},
		{
			name: "缺少 Token",
			config: &GitLabConfig{
				URL:        "https://gitlab.example.com",
				ProjectID:  123,
				MRIID:      456,
			},
			expectError: true,
		},
		{
			name: "缺少 ProjectID",
			config: &GitLabConfig{
				URL:    "https://gitlab.example.com",
				Token:  "test-token",
				MRIID:  456,
			},
			expectError: true,
		},
		{
			name: "缺少 MRIID",
			config: &GitLabConfig{
				URL:        "https://gitlab.example.com",
				Token:      "test-token",
				ProjectID:  123,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
