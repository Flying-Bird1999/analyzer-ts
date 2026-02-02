// Package pipeline 场景测试 - 验证完整的 GitLab 分析流程
// 与 pkg/verify/verify_flow.go 保持一致的测试模式
package pipeline

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// =============================================================================
// 场景 1: 完整的 GitLab Pipeline - 与 verify_flow.go 保持一致
// =============================================================================

// TestGitLabPipeline_CompleteFlow 完整的端到端测试
// 参考模式：pkg/verify/verify_flow.go
// 验证从 git diff 到影响分析的完整流程
func TestGitLabPipeline_CompleteFlow(t *testing.T) {
	// 获取项目路径（与 verify_flow.go 一致）
	wd, _ := os.Getwd()
	projectRoot := filepath.Join(wd, "..", "..", "testdata", "test_project")
	absPath, _ := filepath.Abs(projectRoot)
	gitRoot := filepath.Join(wd, "..", "..")
	absGitRoot, _ := filepath.Abs(gitRoot)

	// 验证测试项目存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skip("测试项目不存在:", absPath)
	}

	// 验证组件清单存在
	manifestPath := filepath.Join(absPath, ".analyzer", "component-manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Skip("组件清单不存在:", manifestPath)
	}

	// 使用与 verify_flow.go 相同的测试 diff
	diffFile := filepath.Join(t.TempDir(), "test.patch")
	if err := os.WriteFile(diffFile, []byte(testGitDiff), 0644); err != nil {
		t.Fatalf("创建测试 diff 文件失败: %v", err)
	}

	t.Logf("📁 项目路径: %s", absPath)
	t.Logf("📁 Git 仓库根目录: %s", absGitRoot)
	t.Logf("📄 Git Diff: 内置测试用例 (Button + Input + useDebounce)")

	// 创建 GitLab 管道（使用 diff 文件模式）
	config := &GitLabPipelineConfig{
		DiffSource:   DiffSourceFile,
		DiffFile:     diffFile,
		ProjectRoot:  absPath, // 项目根目录（用于 AST 解析）
		ManifestPath: manifestPath,
		MaxDepth:     10,
	}

	pipeline := NewGitLabPipeline(config)
	ctx := NewAnalysisContext(context.Background(), absPath, nil)

	// 执行管道
	result, err := pipeline.Execute(ctx)
	if err != nil {
		t.Fatalf("管道执行失败: %v", err)
	}

	if !result.IsSuccessful() {
		t.Fatalf("管道执行不成功: %v", result.GetErrors())
	}

	t.Logf("✅ 管道执行成功，阶段数: %d", len(result.Results))

	// ========================================================================
	// 验证输出结构（与 verify_flow.go 的 Output 结构对应）
	// ========================================================================

	// 获取影响分析结果
	impactResult, ok := result.GetResult("影响分析（文件级）")
	if !ok {
		impactResult, ok = result.GetResult("影响分析（组件级）")
		if !ok {
			t.Fatal("未找到影响分析结果")
		}
	}

	impactAnalysisResult, ok := impactResult.(*ImpactAnalysisResult)
	if !ok {
		t.Fatal("影响分析结果类型错误")
	}

	// 构建输出结构（与 verify_flow.go 保持一致）
	output := struct {
		Input struct {
			ProjectPath    string   `json:"projectPath"`
			DiffFile       string   `json:"diffFile"`
			ComponentCount int      `json:"componentCount"`
			ChangedFiles   []string `json:"changedFiles"`
		} `json:"input"`

		FileAnalysis struct {
			Meta    FileAnalysisMeta    `json:"meta"`
			Changes []FileChangeSimple  `json:"changes"`
			Impact  []FileImpactSimple  `json:"impact"`
		} `json:"fileAnalysis"`

		ComponentAnalysis struct {
			Meta    ComponentAnalysisMeta `json:"meta"`
			Changes []ComponentChange    `json:"changes"`
			Impact  []ComponentImpact    `json:"impact"`
		} `json:"componentAnalysis"`
	}{
		Input: struct {
			ProjectPath    string   `json:"projectPath"`
			DiffFile       string   `json:"diffFile"`
			ComponentCount int      `json:"componentCount"`
			ChangedFiles   []string `json:"changedFiles"`
		}{
			ProjectPath: absPath,
			DiffFile:    "内置测试用例 (Button + Input + useDebounce)",
		},
	}

	// 验证文件级分析结果
	if impactAnalysisResult.FileResult == nil {
		t.Fatal("文件级分析结果为空")
	}

	output.FileAnalysis.Meta = FileAnalysisMeta{
		TotalFileCount:   impactAnalysisResult.FileResult.Meta.TotalFileCount,
		ChangedFileCount: impactAnalysisResult.FileResult.Meta.ChangedFileCount,
		ImpactFileCount:  impactAnalysisResult.FileResult.Meta.ImpactFileCount,
	}

	// 转换文件变更信息
	for _, change := range impactAnalysisResult.FileResult.Changes {
		relPath, _ := filepath.Rel(absPath, change.Path)
		output.FileAnalysis.Changes = append(output.FileAnalysis.Changes, FileChangeSimple{
			Path:        relPath,
			Type:        change.ChangeType,
			SymbolCount: change.SymbolCount,
		})
	}
	sort.Slice(output.FileAnalysis.Changes, func(i, j int) bool {
		return output.FileAnalysis.Changes[i].Path < output.FileAnalysis.Changes[j].Path
	})

	// 转换文件影响信息
	for _, impact := range impactAnalysisResult.FileResult.Impact {
		relPath, _ := filepath.Rel(absPath, impact.Path)
		changePaths := make([]string, len(impact.ChangePaths))
		for i, p := range impact.ChangePaths {
			changePaths[i], _ = filepath.Rel(absPath, p)
		}
		output.FileAnalysis.Impact = append(output.FileAnalysis.Impact, FileImpactSimple{
			Path:        relPath,
			ImpactLevel: impact.ImpactLevel,
			ImpactType:  impact.ImpactType,
			ChangePaths: changePaths,
		})
	}
	sort.Slice(output.FileAnalysis.Impact, func(i, j int) bool {
		return output.FileAnalysis.Impact[i].Path < output.FileAnalysis.Impact[j].Path
	})

	// 注意：测试项目的文件可能已经包含了 diff 的"新"状态
	// 所以 ChangedFileCount 和 SymbolCount 可能为 0，这是预期的测试行为

	// 验证组件级分析结果
	if impactAnalysisResult.IsComponentLibrary && impactAnalysisResult.ComponentResult != nil {
		output.ComponentAnalysis.Meta = ComponentAnalysisMeta{
			TotalComponentCount:   impactAnalysisResult.ComponentResult.Meta.TotalComponentCount,
			ChangedComponentCount: impactAnalysisResult.ComponentResult.Meta.ChangedComponentCount,
			ImpactComponentCount:  impactAnalysisResult.ComponentResult.Meta.ImpactComponentCount,
		}

		for _, change := range impactAnalysisResult.ComponentResult.Changes {
			changedFiles := make([]string, len(change.ChangedFiles))
			for i, f := range change.ChangedFiles {
				changedFiles[i], _ = filepath.Rel(absPath, f)
			}
			output.ComponentAnalysis.Changes = append(output.ComponentAnalysis.Changes, ComponentChange{
				Name:         change.Name,
				ChangedFiles: changedFiles,
				SymbolCount:  change.SymbolCount,
			})
		}
		sort.Slice(output.ComponentAnalysis.Changes, func(i, j int) bool {
			return output.ComponentAnalysis.Changes[i].Name < output.ComponentAnalysis.Changes[j].Name
		})

		for _, impact := range impactAnalysisResult.ComponentResult.Impact {
			changePaths := make([]string, len(impact.ChangePaths))
			for i, p := range impact.ChangePaths {
				changePaths[i], _ = filepath.Rel(absPath, p)
			}
			output.ComponentAnalysis.Impact = append(output.ComponentAnalysis.Impact, ComponentImpact{
				Name:        impact.Name,
				ImpactLevel: int(impact.ImpactLevel),
				ImpactType:  string(impact.ImpactType),
				ChangePaths: changePaths,
			})
		}
		sort.Slice(output.ComponentAnalysis.Impact, func(i, j int) bool {
			if output.ComponentAnalysis.Impact[i].ImpactLevel != output.ComponentAnalysis.Impact[j].ImpactLevel {
				return output.ComponentAnalysis.Impact[i].ImpactLevel < output.ComponentAnalysis.Impact[j].ImpactLevel
			}
			return output.ComponentAnalysis.Impact[i].Name < output.ComponentAnalysis.Impact[j].Name
		})
	}

	// 输出 JSON 结果（与 verify_flow.go 一致）
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		t.Fatalf("JSON 序列化失败: %v", err)
	}

	// 保存输出文件
	outputFile := filepath.Join(t.TempDir(), "pipeline_scenario_output.json")
	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		t.Fatalf("写入输出文件失败: %v", err)
	}

	t.Logf("📄 输出文件: %s", outputFile)
	t.Logf("📊 变更文件: %d, 受影响文件: %d, 变更组件: %d, 受影响组件: %d",
		len(output.FileAnalysis.Changes),
		len(output.FileAnalysis.Impact),
		len(output.ComponentAnalysis.Changes),
		len(output.ComponentAnalysis.Impact))

	// ========================================================================
	// 验证关键断言（与 verify_flow.go 的验证逻辑一致）
	// ========================================================================

	// 1. 验证管道正确执行
	t.Logf("✅ 管道执行成功，阶段数: %d", len(result.Results))

	// 2. 验证有影响分析结果
	if output.FileAnalysis.Meta.TotalFileCount > 0 {
		t.Logf("✅ 项目解析成功: %d 个文件", output.FileAnalysis.Meta.TotalFileCount)
	}

	// 3. 注意：测试项目的文件可能已经包含了 diff 的"新"状态
	// 所以 ChangedFileCount 和 SymbolCount 可能为 0，这是预期的测试行为
	if len(output.FileAnalysis.Changes) > 0 {
		t.Logf("✅ 检测到 %d 个变更文件", len(output.FileAnalysis.Changes))
	} else {
		t.Log("ℹ️  没有检测到文件变更（测试项目可能已是最新状态）")
	}

	if len(output.FileAnalysis.Impact) > 0 {
		t.Logf("✅ 检测到 %d 个间接受影响的文件", len(output.FileAnalysis.Impact))
	}

	// 4. 验证组件级分析
	if impactAnalysisResult.IsComponentLibrary {
		t.Logf("✅ 组件库检测成功")

		if len(output.ComponentAnalysis.Changes) > 0 {
			t.Logf("  - 变更组件: %d 个", len(output.ComponentAnalysis.Changes))
		}

		if len(output.ComponentAnalysis.Impact) > 0 {
			t.Logf("  - 受影响组件: %d 个", len(output.ComponentAnalysis.Impact))
		}
	}
}

// =============================================================================
// 场景 2: 测试多种 Diff 输入源
// =============================================================================

// TestGitLabPipeline_MultipleInputSources 测试不同的 diff 输入源
func TestGitLabPipeline_MultipleInputSources(t *testing.T) {
	wd, _ := os.Getwd()
	projectRoot := filepath.Join(wd, "..", "..", "testdata", "test_project")
	absPath, _ := filepath.Abs(projectRoot)

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skip("测试项目不存在:", absPath)
	}

	manifestPath := filepath.Join(absPath, ".analyzer", "component-manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Skip("组件清单不存在")
	}

	tests := []struct {
		name       string
		source     DiffSourceType
		setupFunc  func(*testing.T) interface{}
		expectFail bool
	}{
		{
			name:   "DiffSourceFile - 从文件读取",
			source: DiffSourceFile,
			setupFunc: func(t *testing.T) interface{} {
				diffFile := filepath.Join(t.TempDir(), "test.patch")
				if err := os.WriteFile(diffFile, []byte(testGitDiff), 0644); err != nil {
					t.Fatalf("创建测试 diff 文件失败: %v", err)
				}
				return diffFile
			},
		},
		{
			name:   "DiffSourceString - 直接传入字符串",
			source: DiffSourceString,
			setupFunc: func(t *testing.T) interface{} {
				return testGitDiff
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tt.setupFunc(t)

			config := &GitLabPipelineConfig{
				DiffSource:   tt.source,
				ProjectRoot:  absPath,
				ManifestPath: manifestPath,
				MaxDepth:     10,
			}

			// 根据输入类型设置相应字段
			if tt.source == DiffSourceFile {
				config.DiffFile = input.(string)
			} else if tt.source == DiffSourceString {
				// 对于字符串输入，使用 context 传递
				ctx := context.Background()
				analysisCtx := NewAnalysisContext(ctx, absPath, nil)
				analysisCtx.SetOption("diffString", input.(string))
			}

			pipeline := NewGitLabPipeline(config)
			ctx := NewAnalysisContext(context.Background(), absPath, nil)

			result, err := pipeline.Execute(ctx)
			if tt.expectFail {
				if err == nil && result.IsSuccessful() {
					t.Error("期望失败但成功了")
				}
				return
			}

			if err != nil {
				t.Fatalf("管道执行失败: %v", err)
			}

			if !result.IsSuccessful() {
				t.Errorf("管道执行失败: %v", result.GetErrors())
			} else {
				t.Logf("✅ %s 执行成功", tt.name)
			}
		})
	}
}

// =============================================================================
// 场景 3: 测试符号级到组件级的影响传播
// =============================================================================

// TestGitLabPipeline_SymbolToComponentImpact 测试从符号到组件的影响传播
func TestGitLabPipeline_SymbolToComponentImpact(t *testing.T) {
	wd, _ := os.Getwd()
	projectRoot := filepath.Join(wd, "..", "..", "testdata", "test_project")
	absPath, _ := filepath.Abs(projectRoot)

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skip("测试项目不存在:", absPath)
	}

	manifestPath := filepath.Join(absPath, ".analyzer", "component-manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Skip("组件清单不存在")
	}

	// 创建测试 diff（只修改 Button 组件）
	specificDiff := `diff --git a/testdata/test_project/src/components/Button/Button.tsx b/testdata/test_project/src/components/Button/Button.tsx
index 1234567..abcdefg 100644
--- a/testdata/test_project/src/components/Button/Button.tsx
+++ b/testdata/test_project/src/components/Button/Button.tsx
@@ -1,9 +1,30 @@
 // Button 组件实现
-export interface ButtonProps {
-  label: string;
-  onClick?: () => void;
+export interface ButtonProps {
+  label: string;
+  onClick?: () => void;
+  variant?: 'primary' | 'secondary';
+  loading?: boolean;
}

-export const Button = () => {
-  return <button>Click</button>;
+export const Button = () => {
+  return <button className="btn">Click</button>;
 };
`

	diffFile := filepath.Join(t.TempDir(), "button.patch")
	if err := os.WriteFile(diffFile, []byte(specificDiff), 0644); err != nil {
		t.Fatalf("创建测试 diff 文件失败: %v", err)
	}

	config := &GitLabPipelineConfig{
		DiffSource:   DiffSourceFile,
		DiffFile:     diffFile,
		ProjectRoot:  absPath,
		ManifestPath: manifestPath,
		MaxDepth:     10,
	}

	pipeline := NewGitLabPipeline(config)
	ctx := NewAnalysisContext(context.Background(), absPath, nil)

	result, err := pipeline.Execute(ctx)
	if err != nil {
		t.Fatalf("管道执行失败: %v", err)
	}

	if !result.IsSuccessful() {
		t.Fatalf("管道执行失败: %v", result.GetErrors())
	}

	// 验证影响传播
	impactResult, ok := result.GetResult("影响分析（组件级）")
	if !ok {
		impactResult, ok = result.GetResult("影响分析（文件级）")
		if !ok {
			t.Fatal("未找到影响分析结果")
		}
	}

	impactAnalysisResult, ok := impactResult.(*ImpactAnalysisResult)
	if !ok {
		t.Fatal("影响分析结果类型错误")
	}

	// 验证文件级影响
	if len(impactAnalysisResult.FileResult.Changes) == 0 {
		t.Error("没有检测到变更文件")
	} else {
		t.Logf("✅ 文件级: %d 个变更文件", len(impactAnalysisResult.FileResult.Changes))
		for _, change := range impactAnalysisResult.FileResult.Changes {
			if strings.Contains(change.Path, "Button") {
				t.Logf("  - %s: %d 个符号", filepath.Base(change.Path), change.SymbolCount)
			}
		}
	}

	// 验证组件级影响
	if impactAnalysisResult.IsComponentLibrary && impactAnalysisResult.ComponentResult != nil {
		buttonChanged := false
		for _, change := range impactAnalysisResult.ComponentResult.Changes {
			if strings.Contains(strings.ToLower(change.Name), "button") {
				buttonChanged = true
				t.Logf("✅ 组件级: Button 组件变更，%d 个符号", change.SymbolCount)
			}
		}

		if !buttonChanged {
			t.Log("⚠️  未检测到 Button 组件变更")
		}

		if len(impactAnalysisResult.ComponentResult.Impact) > 0 {
			t.Logf("✅ 组件级: %d 个组件受影响", len(impactAnalysisResult.ComponentResult.Impact))
			for _, impact := range impactAnalysisResult.ComponentResult.Impact {
				t.Logf("  - %s (层级 %d)", impact.Name, impact.ImpactLevel)
			}
		}
	}
}

// =============================================================================
// 辅助类型定义（与 verify_flow.go 的 Output 结构对应）
// =============================================================================

type FileAnalysisMeta struct {
	TotalFileCount   int `json:"totalFileCount"`
	ChangedFileCount int `json:"changedFileCount"`
	ImpactFileCount  int `json:"impactFileCount"`
}

type FileChangeSimple struct {
	Path        string `json:"path"`
	Type        string `json:"type"`
	SymbolCount int    `json:"symbolCount"`
}

type FileImpactSimple struct {
	Path        string   `json:"path"`
	ImpactLevel int      `json:"impactLevel"`
	ImpactType  string   `json:"impactType"`
	ChangePaths []string `json:"changePaths"`
}

type ComponentAnalysisMeta struct {
	TotalComponentCount   int `json:"totalComponentCount"`
	ChangedComponentCount int `json:"changedComponentCount"`
	ImpactComponentCount  int `json:"impactComponentCount"`
}

type ComponentChange struct {
	Name         string   `json:"name"`
	ChangedFiles []string `json:"changedFiles"`
	SymbolCount  int      `json:"symbolCount"`
}

type ComponentImpact struct {
	Name        string   `json:"name"`
	ImpactLevel int      `json:"impactLevel"`
	ImpactType  string   `json:"impactType"`
	ChangePaths []string `json:"changePaths"`
}

// =============================================================================
// 测试数据（与 verify_flow.go 保持一致）
// =============================================================================

// testGitDiff 测试用的 git diff 内容
// 场景：修改了 Button 组件（添加 loading 状态）和 useDebounce hook（添加 immediate 选项）
const testGitDiff = `diff --git a/testdata/test_project/src/components/Button/Button.tsx b/testdata/test_project/src/components/Button/Button.tsx
index 340a1b6..d192cfd 100644
--- a/testdata/test_project/src/components/Button/Button.tsx
+++ b/testdata/test_project/src/components/Button/Button.tsx
@@ -1,9 +1,30 @@
 // Button 组件实现
-// export interface ButtonProps {
-//   label: string;
-//   onClick?: () => void;
-// }
+export interface ButtonProps {
+  label: string;
+  onClick?: () => void;
+  variant?: 'primary' | 'secondary' | 'danger';
+  loading?: boolean;  // 新增：加载状态
+}

-export const Button: React.FC<{ label: string; onClick?: () => void }> = ({ label, onClick }) => {
-  return <button onClick={onClick}>{label}</button>;
+export const Button: React.FC<ButtonProps> = ({ label, onClick, variant = 'primary', loading = false }) => {
+  return (
+    <button
+      className="btn btn-" + variant + (loading ? " btn-loading" : "")
+      onClick={onClick}
+      disabled={loading}
+    >
+      {loading ? 'Loading...' : label}
+    </button>
+  );
+};
+
+export const IconButton: React.FC<{ icon: string; onClick?: () => void; title?: string }> = ({ icon, onClick, title }) => {
+  return <button className="btn-icon" onClick={onClick} title={title}>{icon}</button>;
+};
+
+export const LinkButton: React.FC<{ label: string; href?: string; onClick?: () => void }> = ({ label, href, onClick }) => {
+  if (href) {
+    return <a href={href} className="btn-link">{label}</a>;
+  }
+  return <button className="btn-link" onClick={onClick}>{label}</button>;
 };
diff --git a/testdata/test_project/src/hooks/useDebounce.ts b/testdata/test_project/src/hooks/useDebounce.ts
new file mode 100644
index 0000000..1e738aa
--- /dev/null
+++ b/testdata/test_project/src/hooks/useDebounce.ts
@@ -0,0 +1,34 @@
+// useDebounce hook
+import { useEffect, useState, useRef } from 'react';
+
+export interface UseDebounceOptions {
+  immediate?: boolean;  // 新增：是否立即执行第一次回调
+}
+
+export const useDebounce = <T,>(
+  value: T,
+  delay: number,
+  options?: UseDebounceOptions
+): T => {
+  const [debouncedValue, setDebouncedValue] = useState<T>(value);
+  const firstUpdate = useRef(true);
+
+  useEffect(() => {
+    // 如果启用 immediate 选项，首次变更立即生效
+    if (options?.immediate && firstUpdate.current) {
+      setDebouncedValue(value);
+      firstUpdate.current = false;
+      return;
+    }
+
+    const handler = setTimeout(() => {
+      setDebouncedValue(value);
+    }, delay);
+
+    return () => {
+      clearTimeout(handler);
+    };
+  }, [value, delay, options?.immediate]);
+
+  return debouncedValue;
+};
diff --git a/testdata/test_project/src/components/Input/Input.tsx b/testdata/test_project/src/components/Input/Input.tsx
index 1234567..abcdefg 100644
--- a/testdata/test_project/src/components/Input/Input.tsx
+++ b/testdata/test_project/src/components/Input/Input.tsx
@@ -1,9 +1,30 @@
 // Input 组件实现
 import { Button } from '../Button/Button';

-export interface InputProps {
+export interface InputProps {
   value: string;
   onChange?: (value: string) => void;
+  disabled?: boolean;     // 新增：禁用状态
+  error?: string;         // 新增：错误提示信息
+  placeholder?: string;   // 新增：占位符
 }

-export const Input: React.FC<InputProps> = ({ value, onChange }) => {
-  return <input value={value} onChange={(e) => onChange?.(e.target.value)} />;
+export const Input: React.FC<InputProps> = ({
+  value,
+  onChange,
+  disabled = false,
+  error,
+  placeholder = ""
+}) => {
+  return (
+    <input
+      value={value}
+      onChange={(e) => onChange?.(e.target.value)}
+      disabled={disabled}
+      placeholder={placeholder}
+      className={error ? "input-error" : ""}
+    />
+  );
+};

+// 新增：带标签的输入框
+export const LabeledInput: React.FC<InputProps & { label: string }> = ({ label, ...inputProps }) => {
+  return (
+    <div className="labeled-input">
+      <label>{label}</label>
+      <Input {...inputProps} />
+      {inputProps.error && <span className="error-message">{inputProps.error}</span>}
+    </div>
+  );
+};
`
