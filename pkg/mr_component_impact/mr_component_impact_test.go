// Package mr_component_impact MR 组件影响分析 - 统一测试文件
package mr_component_impact

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/component_deps"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/export_call"
)

// =============================================================================
// 简化 API 调用测试
// =============================================================================

// TestE2E_AnalyzeFromDiff 测试简化的 API 调用方式
func TestE2E_AnalyzeFromDiff(t *testing.T) {
	projectRoot := "../../testdata/test_project"
	absProjectRoot, _ := filepath.Abs(projectRoot)

	t.Run("Button组件变更", func(t *testing.T) {
		// 创建临时 diff 文件
		diffFile := createTempDiffFile(t, []string{
			"src/components/Button/Button.tsx",
		})

		// 使用简化的 API 调用
		result, err := AnalyzeFromDiff(&AnalyzeConfig{
			ProjectRoot:  absProjectRoot,
			DiffFilePath: diffFile,
		})

		if err != nil {
			t.Fatalf("分析失败: %v", err)
		}

		// 验证结果
		if len(result.ChangedComponents) != 1 {
			t.Errorf("期望 1 个变更组件，实际 %d", len(result.ChangedComponents))
		}

		if _, exists := result.ChangedComponents["Button"]; !exists {
			t.Error("Button 应该在变更组件列表中")
		}

		// 验证受影响组件
		if len(result.ImpactedComponents) == 0 {
			t.Error("期望有受影响的组件")
		}

		expectedImpacted := []string{"Card", "Form", "Input", "Modal", "Select", "Table"}
		for _, comp := range expectedImpacted {
			if _, exists := result.ImpactedComponents[comp]; !exists {
				t.Errorf("期望 %s 在受影响组件列表中", comp)
			}
		}

		t.Logf("Button 变更影响分析:\n%s", result.ToConsole())
	})

	t.Run("Utils函数变更", func(t *testing.T) {
		diffFile := createTempDiffFile(t, []string{
			"src/utils/validation.ts",
		})

		result, err := AnalyzeFromDiff(&AnalyzeConfig{
			ProjectRoot:  absProjectRoot,
			DiffFilePath: diffFile,
		})

		if err != nil {
			t.Fatalf("分析失败: %v", err)
		}

		if len(result.ChangedFunctions) != 1 {
			t.Errorf("期望 1 个变更函数，实际 %d", len(result.ChangedFunctions))
		}

		if _, exists := result.ChangedFunctions["utils"]; !exists {
			t.Error("utils 应该在变更函数列表中")
		}

		t.Logf("Utils 变更影响分析:\n%s", result.ToConsole())
	})

	t.Run("Hooks函数变更(useDebounce)", func(t *testing.T) {
		diffFile := createTempDiffFile(t, []string{
			"src/hooks/useDebounce.ts",
		})

		result, err := AnalyzeFromDiff(&AnalyzeConfig{
			ProjectRoot:  absProjectRoot,
			DiffFilePath: diffFile,
		})

		if err != nil {
			t.Fatalf("分析失败: %v", err)
		}
		// 验证受影响组件
		if len(result.ImpactedComponents) != 2 {
			t.Error("受影响的组件为2")
		}

		expectedImpacted := []string{"Select", "Table"}
		for _, comp := range expectedImpacted {
			if _, exists := result.ImpactedComponents[comp]; !exists {
				t.Errorf("期望 %s 在受影响组件列表中", comp)
			}
		}

		t.Logf("Hooks 变更影响分析:\n%s", result.ToConsole())
	})

	t.Run("Hooks函数变更(useCounter)", func(t *testing.T) {
		// 验证 Counter 组件添加到 manifest 后，useCounter 的影响分析正常工作
		diffFile := createTempDiffFile(t, []string{
			"src/hooks/useCounter.ts",
		})

		result, err := AnalyzeFromDiff(&AnalyzeConfig{
			ProjectRoot:  absProjectRoot,
			DiffFilePath: diffFile,
		})

		if err != nil {
			t.Fatalf("分析失败: %v", err)
		}

		// 验证受影响组件 - Counter 组件应该被检测到
		if len(result.ImpactedComponents) != 1 {
			t.Errorf("期望 1 个受影响组件，实际 %d", len(result.ImpactedComponents))
		}

		if _, exists := result.ImpactedComponents["Counter"]; !exists {
			t.Error("Counter 应该在受影响组件列表中")
		}

		// 验证影响原因
		for _, impact := range result.ImpactedComponents["Counter"] {
			if impact.Relation != RelationImports {
				t.Errorf("期望关系类型为 'imports'，实际 '%s'", impact.Relation)
			}
		}

		t.Logf("useCounter 变更影响分析:\n%s", result.ToConsole())
	})

	t.Run("混合变更", func(t *testing.T) {
		diffFile := createTempDiffFile(t, []string{
			"src/components/Button/Button.tsx",
			"src/utils/validation.ts",
		})

		result, err := AnalyzeFromDiff(&AnalyzeConfig{
			ProjectRoot:  absProjectRoot,
			DiffFilePath: diffFile,
		})

		if err != nil {
			t.Fatalf("分析失败: %v", err)
		}

		if len(result.ChangedComponents) == 0 {
			t.Error("期望有变更组件")
		}

		if len(result.ChangedFunctions) == 0 {
			t.Error("期望有变更函数")
		}

		t.Logf("混合变更影响分析:\n%s", result.ToConsole())
	})

	t.Run("类型文件变更", func(t *testing.T) {
		diffFile := createTempDiffFile(t, []string{
			"src/types/common.ts",
		})

		result, err := AnalyzeFromDiff(&AnalyzeConfig{
			ProjectRoot:  absProjectRoot,
			DiffFilePath: diffFile,
		})

		if err != nil {
			t.Fatalf("分析失败: %v", err)
		}

		// 类型文件被归类为 functions
		if len(result.ChangedFunctions) != 1 {
			t.Errorf("期望 1 个变更函数（types），实际 %d", len(result.ChangedFunctions))
		}

		t.Logf("类型文件变更影响分析:\n%s", result.ToConsole())
	})

	t.Run("空diff文件", func(t *testing.T) {
		diffFile := createTempDiffFile(t, []string{})

		result, err := AnalyzeFromDiff(&AnalyzeConfig{
			ProjectRoot:  absProjectRoot,
			DiffFilePath: diffFile,
		})

		if err != nil {
			t.Fatalf("分析失败: %v", err)
		}

		if len(result.ChangedComponents) != 0 || len(result.ChangedFunctions) != 0 {
			t.Error("空 diff 文件应该返回空结果")
		}

		t.Logf("空 diff 文件分析:\n%s", result.ToConsole())
	})
}

// =============================================================================
// Diff 文件解析测试
// =============================================================================

// TestParseDiffFile 测试 diff 文件解析功能
func TestParseDiffFile(t *testing.T) {
	tmpDir := t.TempDir()
	diffFilePath := filepath.Join(tmpDir, "test.diff")

	// 写入标准的 git diff 格式内容
	diffContent := `diff --git a/src/components/Button/Button.tsx b/src/components/Button/Button.tsx
index 1234567..abcdef 100644
--- a/src/components/Button/Button.tsx
+++ b/src/components/Button/Button.tsx
@@ -1,5 +1,5 @@
 export function Button() {
-  return <button>Old</button>;
+  return <button>New</button>;
 }
diff --git a/src/utils/validation.ts b/src/utils/validation.ts
index 2345678..bcdef01 100644
--- a/src/utils/validation.ts
+++ b/src/utils/validation.ts
@@ -1,3 +1,4 @@
 export function validate() {
+  return true;
 }
diff --git a/src/hooks/useCounter.ts b/src/hooks/useCounter.ts
new file mode 100644
index 0000000..1234567
--- /dev/null
+++ b/src/hooks/useCounter.ts
@@ -0,0 +1,5 @@
+export function useCounter() {
+  return 0;
+}
`

	if err := os.WriteFile(diffFilePath, []byte(diffContent), 0644); err != nil {
		t.Fatalf("写入 diff 文件失败: %v", err)
	}

	// 测试解析
	files, err := parseDiffFile(diffFilePath)
	if err != nil {
		t.Fatalf("解析 diff 文件失败: %v", err)
	}

	// 验证解析结果
	if len(files) != 3 {
		t.Errorf("期望解析出 3 个文件，实际 %d", len(files))
	}

	expectedFiles := map[string]bool{
		"src/components/Button/Button.tsx": false,
		"src/utils/validation.ts":          false,
		"src/hooks/useCounter.ts":          false,
	}

	for _, file := range files {
		if _, exists := expectedFiles[file]; !exists {
			t.Errorf("未预期的文件: %s", file)
		} else {
			expectedFiles[file] = true
		}
	}

	// 检查是否所有期望的文件都被解析出来
	for file, found := range expectedFiles {
		if !found {
			t.Errorf("未找到期望的文件: %s", file)
		}
	}
}

// TestParseDiffFileEmpty 测试空 diff 文件
func TestParseDiffFileEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	diffFilePath := filepath.Join(tmpDir, "empty.diff")

	if err := os.WriteFile(diffFilePath, []byte(""), 0644); err != nil {
		t.Fatalf("写入 diff 文件失败: %v", err)
	}

	files, err := parseDiffFile(diffFilePath)
	if err != nil {
		t.Fatalf("解析空 diff 文件失败: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("期望解析出 0 个文件，实际 %d", len(files))
	}
}

// TestParseDiffFileInvalid 测试不存在的 diff 文件
func TestParseDiffFileInvalid(t *testing.T) {
	_, err := parseDiffFile("/nonexistent/file.diff")
	if err == nil {
		t.Error("期望返回错误，但返回了 nil")
	}
}

// TestGetChangedFilesFromDiff 测试公开的辅助函数
func TestGetChangedFilesFromDiff(t *testing.T) {
	tmpDir := t.TempDir()
	diffFilePath := filepath.Join(tmpDir, "test.diff")

	diffContent := `diff --git a/test/file.ts b/test/file.ts
index 123..456 789
--- a/test/file.ts
+++ b/test/file.ts
`

	if err := os.WriteFile(diffFilePath, []byte(diffContent), 0644); err != nil {
		t.Fatalf("写入 diff 文件失败: %v", err)
	}

	files, err := GetChangedFilesFromDiff(diffFilePath)
	if err != nil {
		t.Fatalf("GetChangedFilesFromDiff 失败: %v", err)
	}

	if len(files) != 1 || files[0] != "test/file.ts" {
		t.Errorf("期望解析出 [test/file.ts]，实际 %v", files)
	}
}

// TestParseDiffFileWithChinesePath 测试包含中文路径的 diff 文件
func TestParseDiffFileWithChinesePath(t *testing.T) {
	tmpDir := t.TempDir()
	diffFilePath := filepath.Join(tmpDir, "test.diff")

	// 注意：git diff 实际上会对中文路径进行编码，这里仅测试基本功能
	diffContent := `diff --git a/src/组件/Button.tsx b/src/组件/Button.tsx
index 123..456 789
--- a/src/组件/Button.tsx
+++ b/src/组件/Button.tsx
`

	if err := os.WriteFile(diffFilePath, []byte(diffContent), 0644); err != nil {
		t.Fatalf("写入 diff 文件失败: %v", err)
	}

	files, err := parseDiffFile(diffFilePath)
	if err != nil {
		t.Fatalf("解析 diff 文件失败: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("期望解析出 1 个文件，实际 %d", len(files))
	}

	// 验证中文路径是否正确解析
	if !strings.Contains(files[0], "src/组件/Button.tsx") {
		t.Errorf("中文路径解析不正确，期望包含 'src/组件/Button.tsx'，实际 %s", files[0])
	}
}

// =============================================================================
// 文件分类器测试
// =============================================================================

// TestE2E_Classifier 测试文件分类器
func TestE2E_Classifier(t *testing.T) {
	projectRoot := "../../testdata/test_project"
	manifestPath := filepath.Join(projectRoot, ".analyzer", "component-manifest.json")

	// 加载 manifest
	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("加载 manifest 失败: %v", err)
	}

	// 创建分类器
	classifier := NewClassifier(manifest, []string{
		filepath.Join(projectRoot, "src/utils"),
		filepath.Join(projectRoot, "src/hooks"),
		filepath.Join(projectRoot, "src/types"),
	})

	// 测试用例
	tests := []struct {
		file         string
		expectedCat  FileCategory
		expectedName string
	}{
		{
			file:         filepath.Join(projectRoot, "src/components/Button/Button.tsx"),
			expectedCat:  CategoryComponent,
			expectedName: "Button",
		},
		{
			file:         filepath.Join(projectRoot, "src/components/Form/Form.tsx"),
			expectedCat:  CategoryComponent,
			expectedName: "Form",
		},
		{
			file:         filepath.Join(projectRoot, "src/utils/validation.ts"),
			expectedCat:  CategoryFunctions,
			expectedName: "utils",
		},
		{
			file:         filepath.Join(projectRoot, "src/hooks/useForm.ts"),
			expectedCat:  CategoryFunctions,
			expectedName: "hooks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			cat, name := classifier.ClassifyFile(tt.file)
			if cat != tt.expectedCat {
				t.Errorf("期望分类 %s，实际 %s", tt.expectedCat, cat)
			}
			if name != tt.expectedName {
				t.Errorf("期望名称 %s，实际 %s", tt.expectedName, name)
			}
		})
	}
}

// =============================================================================
// Mock 数据测试
// =============================================================================

// TestE2E_HooksWithMockData 使用模拟数据测试 hooks 函数变更的影响分析
func TestE2E_HooksWithMockData(t *testing.T) {
	projectRoot := "../../testdata/test_project"

	// 创建模拟的 export_call 结果
	// 模拟 Counter 组件引用了 useCounter hook
	mockExportCallResult := &export_call.ExportCallResult{
		ModuleExports: []export_call.ModuleExportRecord{
			{
				ModuleName: "hooks",
				Path:       filepath.Join(projectRoot, "src/hooks"),
				Files: []export_call.FileExportRecord{
					{
						File: filepath.Join(projectRoot, "src/hooks/useCounter.ts"),
						Nodes: []export_call.NodeWithRefs{
							{
								Name:       "useCounter",
								NodeType:   "function",
								ExportType: "default",
								RefFiles: []string{
									filepath.Join(projectRoot, "src/components/Counter/index.tsx"),
								},
								RefComponents: []export_call.ComponentRef{
									{
										ComponentName: "Counter",
										RefFiles: []string{
											filepath.Join(projectRoot, "src/components/Counter/index.tsx"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// 创建模拟的 component_deps 结果
	mockComponentDeps := createMockComponentDepsResult(&ComponentManifest{
		Components: map[string]ComponentInfo{
			"Button":  {Name: "Button", Path: filepath.Join(projectRoot, "src/components/Button"), Type: "component"},
			"Counter": {Name: "Counter", Path: filepath.Join(projectRoot, "src/components/Counter"), Type: "component"},
		},
	})

	// 创建 MR 组件影响分析器
	analyzer := NewAnalyzer(&AnalyzerConfig{
		Manifest: &ComponentManifest{
			Components: map[string]ComponentInfo{
				"Button":  {Name: "Button", Path: filepath.Join(projectRoot, "src/components/Button"), Type: "component"},
				"Counter": {Name: "Counter", Path: filepath.Join(projectRoot, "src/components/Counter"), Type: "component"},
			},
			Functions: map[string]FunctionInfo{
				"hooks": {Name: "hooks", Path: filepath.Join(projectRoot, "src/hooks"), Type: "functions"},
			},
		},
		FunctionPaths: []string{filepath.Join(projectRoot, "src/hooks")},
		ComponentDeps: mockComponentDeps,
		ExportCall:    mockExportCallResult,
	})

	// 测试：useCounter hook 变更
	changedFiles := []string{
		filepath.Join(projectRoot, "src/hooks/useCounter.ts"),
	}

	result := analyzer.Analyze(changedFiles)

	// 验证结果
	if len(result.ChangedFunctions) != 1 {
		t.Errorf("期望 1 个变更函数，实际 %d", len(result.ChangedFunctions))
	}

	if _, exists := result.ChangedFunctions["hooks"]; !exists {
		t.Error("hooks 应该在变更函数列表中")
	}

	// 验证受影响的组件
	// Counter 组件引用了 useCounter，应该被识别为受影响的组件
	if len(result.ImpactedComponents) == 0 {
		t.Error("期望有受影响的组件（Counter）")
	} else {
		if _, exists := result.ImpactedComponents["Counter"]; !exists {
			t.Errorf("Counter 应该在受影响组件列表中，实际: %v", result.GetImpactedComponentNames())
		} else {
			t.Logf("✅ 成功检测到 Counter 组件受影响!")
			// 检查影响原因
			for _, impact := range result.ImpactedComponents["Counter"] {
				t.Logf("  影响原因: %s", impact.DisplayReason())
				if impact.Relation == RelationImports {
					t.Logf("  ✅ 关系类型正确: imports")
				}
			}
		}
	}

	t.Logf("\n完整分析结果:\n%s", result.ToConsole())
}

// =============================================================================
// 辅助函数
// =============================================================================

// createTempDiffFile 创建临时 diff 文件用于测试
func createTempDiffFile(t *testing.T, files []string) string {
	t.Helper()

	tmpDir := t.TempDir()
	diffFilePath := filepath.Join(tmpDir, "test.diff")

	// 生成标准 git diff 格式
	var diffContent string
	for _, file := range files {
		diffContent += fmt.Sprintf("diff --git a/%s b/%s\n", file, file)
		diffContent += "index 1234567..abcdef 100644\n"
		diffContent += fmt.Sprintf("--- a/%s\n", file)
		diffContent += fmt.Sprintf("+++ b/%s\n", file)
		diffContent += "@@ -1,5 +1,5 @@\n"
		diffContent += "- old content\n"
		diffContent += "+ new content\n"
	}

	if err := os.WriteFile(diffFilePath, []byte(diffContent), 0644); err != nil {
		t.Fatalf("创建 diff 文件失败: %v", err)
	}

	return diffFilePath
}

// createMockComponentDepsResult 创建模拟的组件依赖结果
func createMockComponentDepsResult(manifest *ComponentManifest) *component_deps.ComponentDepsResult {
	components := make(map[string]component_deps.ComponentInfo)
	for name, comp := range manifest.Components {
		components[name] = component_deps.ComponentInfo{
			Name:          comp.Name,
			Path:          comp.Path,
			Dependencies:  []projectParser.ImportDeclarationResult{},
			ComponentDeps: []component_deps.ComponentDep{},
		}
	}

	return &component_deps.ComponentDepsResult{
		Meta:       component_deps.Meta{ComponentCount: len(manifest.Components)},
		Components: components,
	}
}
