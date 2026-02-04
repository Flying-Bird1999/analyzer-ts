// Package pipeline 场景测试 - 与 pkg/verify/verify_flow.go 一一对齐
package pipeline

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/pkg/symbol_analysis"
)

// testGitDiff 测试用的 git diff 内容
// 场景：修改了 Button 组件接口（添加 loading 状态）和 useDebounce hook（新增文件）
const testGitDiff = `diff --git a/testdata/test_project/src/components/Button/Button.tsx b/testdata/test_project/src/components/Button/Button.tsx
index 340a1b6..d192cfd 100644
--- a/testdata/test_project/src/components/Button/Button.tsx
+++ b/testdata/test_project/src/components/Button/Button.tsx
@@ -1,9 +1,32 @@
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

-const Button: React.FC<{ label: string; onClick?: () => void }> = ({ label, onClick }) => {
-  return <button onClick={onClick}>{label}</button>;
+const Button: React.FC<ButtonProps> = ({ label, onClick, variant = 'primary', loading = false }) => {
+  return (
+    <button
+      className={"btn btn-" + variant + (loading ? " btn-loading" : "")}
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
+};
+
+export default Button;
diff --git a/testdata/test_project/src/hooks/useDebounce.ts b/testdata/test_project/src/hooks/useDebounce.ts
new file mode 100644
index 0000000..1e738aa
--- /dev/null
+++ b/testdata/test_project/src/hooks/useDebounce.ts
@@ -0,0 +1,34 @@
++// useDebounce hook
++import { useEffect, useState, useRef } from 'react';
++
++export interface UseDebounceOptions {
++  immediate?: boolean;  // 新增：是否立即执行第一次回调
++}
++
++export const useDebounce = <T,>(
++  value: T,
++  delay: number,
++  options?: UseDebounceOptions
++): T => {
++  const [debouncedValue, setDebouncedValue] = useState<T>(value);
++  const firstUpdate = useRef(true);
++
++  useEffect(() => {
++    // 如果启用 immediate 选项，首次变更立即生效
++    if (options?.immediate && firstUpdate.current) {
++      setDebouncedValue(value);
++      firstUpdate.current = false;
++      return;
++    }
++
++    const handler = setTimeout(() => {
++      setDebouncedValue(value);
++    }, delay);
++
++    return () => {
++      clearTimeout(handler);
++    };
++  }, [value, delay, options?.immediate]);
++
++  return debouncedValue;
++};
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

++// 新增：带标签的输入框
++export const LabeledInput: React.FC<InputProps & { label: string }> = ({ label, ...inputProps }) => {
++  return (
++    <div className="labeled-input">
++      <label>{label}</label>
++      <Input {...inputProps} />
++      {inputProps.error && <span className="error-message">{inputProps.error}</span>}
++    </div>
++  );
++};
diff --git a/testdata/test_project/src/assets/logo.png b/testdata/test_project/src/assets/logo.png
index 1234567..abcdefg 100644
Binary files a/testdata/test_project/src/assets/logo.png and b/testdata/test_project/src/assets/logo.png differ
diff --git a/testdata/test_project/src/assets/modal.css b/testdata/test_project/src/assets/modal.css
new file mode 100644
index 0000000..1234567
--- /dev/null
+++ b/testdata/test_project/src/assets/modal.css
@@ -0,0 +1,13 @@
++/* Modal 组件样式 */
++.modal-overlay {
++  position: fixed;
++  top: 0;
++  left: 0;
++  right: 0;
++  bottom: 0;
++  background: rgba(0, 0, 0, 0.5);
++}
++
++.modal-content {
++  position: fixed;
++  top: 50%;
++  left: 50%;
++  transform: translate(-50%, -50%);
++  background: white;
++  padding: 20px;
++  border-radius: 8px;
++}
diff --git a/testdata/test_project/src/types/enums.ts b/testdata/test_project/src/types/enums.ts
index 1234567..abcdefg 100644
--- a/testdata/test_project/src/types/enums.ts
+++ b/testdata/test_project/src/types/enums.ts
@@ -1,11 +1,18 @@
 // 枚举类型定义

 export enum ButtonSize {
   Small = 'small',
   Medium = 'medium',
   Large = 'large'
+  ExtraLarge = 'xlarge'  // 新增：超大尺寸
 }

 export enum ThemeColor {
   Primary = 'primary',
   Secondary = 'secondary',
   Success = 'success',
   Warning = 'warning',
   Danger = 'danger',
-  Info = 'info'
+  Info = 'info',
+  Light = 'light',       // 新增：浅色主题
+  Dark = 'dark'          // 新增：深色主题
 }

 export enum Direction {
   Horizontal = 'horizontal',
   Vertical = 'vertical'
+  Diagonal = 'diagonal'  // 新增：对角线方向
 }

 export enum Align {
   Left = 'left',
   Center = 'center',
   Right = 'right',
   Justify = 'justify'
 }
`

// =============================================================================
// TestGitLabPipeline_CompleteFlow
// 与 verify_flow.go 的 main 函数一一对应
// =============================================================================

// TestGitLabPipeline_CompleteFlow 完整的端到端测试
// 参考模式：pkg/verify/verify_flow.go
// 验证从 git diff 到影响分析的完整流程
func TestGitLabPipeline_CompleteFlow(t *testing.T) {
	// 项目路径（与 verify_flow.go 一致）
	wd, _ := os.Getwd()
	projectRoot := filepath.Join(wd, "..", "..", "testdata", "test_project")
	absPath, _ := filepath.Abs(projectRoot)

	// git 仓库根目录（diff 路径是相对于 git 仓库根的）
	// GitRoot 与 ProjectRoot 不同时，需要显式传入绝对路径
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

	t.Logf("📁 项目路径: %s", absPath)
	t.Logf("📁 Git 仓库根目录: %s", absGitRoot)
	t.Logf("📄 Git Diff: 内置测试用例 (Button + Input + useDebounce + 二进制文件 + 新增CSS + 枚举类型)")

	// 创建 GitLab 管道
	// GitRoot 显式传入绝对路径（与 ProjectRoot 不同）
	config := &GitLabPipelineConfig{
		DiffSource:   DiffSourceString, // 使用字符串输入，与 verify_flow.go 一致
		ProjectRoot:  absPath,          // 项目根目录（用于 AST 解析）
		GitRoot:      absGitRoot,       // Git 仓库根目录（用于 diff 解析）- 显式传入绝对路径
		ManifestPath: manifestPath,
		MaxDepth:     10,
	}

	// 创建分析上下文，传入 diff 字符串
	ctx := context.Background()
	analysisCtx := NewAnalysisContext(ctx, absPath, nil)
	analysisCtx.SetOption("diffString", testGitDiff)

	pipeline := NewGitLabPipeline(config)

	// 执行管道
	result, err := pipeline.Execute(analysisCtx)
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
			Meta    FileAnalysisMeta   `json:"meta"`
			Changes []FileChangeSimple `json:"changes"`
			Impact  []FileImpactSimple `json:"impact"`
		} `json:"fileAnalysis"`

		ComponentAnalysis struct {
			Meta    ComponentAnalysisMeta `json:"meta"`
			Changes []ComponentChange     `json:"changes"`
			Impact  []ComponentImpact     `json:"impact"`
		} `json:"componentAnalysis"`
	}{}

	// 填充输出数据
	output.Input.ProjectPath = absPath
	output.Input.DiffFile = "内置测试用例 (Button + Input + useDebounce + 二进制文件 + 新增CSS + 枚举类型)"

	// 文件级分析结果
	if impactAnalysisResult.FileResult != nil {
		output.FileAnalysis.Meta = FileAnalysisMeta{
			TotalFileCount:   impactAnalysisResult.FileResult.Meta.TotalFileCount,
			ChangedFileCount: impactAnalysisResult.FileResult.Meta.ChangedFileCount,
			ImpactFileCount:  impactAnalysisResult.FileResult.Meta.ImpactFileCount,
		}

		for _, change := range impactAnalysisResult.FileResult.Changes {
			relPath, _ := filepath.Rel(absPath, change.Path)
			output.FileAnalysis.Changes = append(output.FileAnalysis.Changes, FileChangeSimple{
				Path:        relPath,
				Type:        string(change.ChangeType),
				SymbolCount: change.SymbolCount,
			})
			output.Input.ChangedFiles = append(output.Input.ChangedFiles, relPath)
		}
		sort.Slice(output.FileAnalysis.Changes, func(i, j int) bool {
			return output.FileAnalysis.Changes[i].Path < output.FileAnalysis.Changes[j].Path
		})

		for _, impact := range impactAnalysisResult.FileResult.Impact {
			relPath, _ := filepath.Rel(absPath, impact.Path)
			changePaths := make([]string, len(impact.ChangePaths))
			for i, p := range impact.ChangePaths {
				changePaths[i], _ = filepath.Rel(absPath, p)
			}
			output.FileAnalysis.Impact = append(output.FileAnalysis.Impact, FileImpactSimple{
				Path:        relPath,
				ImpactLevel: impact.ImpactLevel,
				ImpactType:  string(impact.ImpactType),
				ChangePaths: changePaths,
			})
		}
		sort.Slice(output.FileAnalysis.Impact, func(i, j int) bool {
			return output.FileAnalysis.Impact[i].Path < output.FileAnalysis.Impact[j].Path
		})
	}

	// 组件级分析结果
	if impactAnalysisResult.ComponentResult != nil {
		output.ComponentAnalysis.Meta = ComponentAnalysisMeta{
			TotalComponentCount:   impactAnalysisResult.ComponentResult.Meta.TotalComponentCount,
			ChangedComponentCount: impactAnalysisResult.ComponentResult.Meta.ChangedComponentCount,
			ImpactComponentCount:  impactAnalysisResult.ComponentResult.Meta.ImpactComponentCount,
		}
		output.Input.ComponentCount = impactAnalysisResult.ComponentResult.Meta.TotalComponentCount

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

	// 输出 JSON 到文件（与 verify_flow.go 一致）
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		t.Fatalf("JSON 序列化失败: %v", err)
	}

	outputFile := filepath.Join(t.TempDir(), "pipeline_scenario_output.json")
	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		t.Fatalf("写入输出文件失败: %v", err)
	}

	t.Logf("📄 输出文件: %s", outputFile)
	t.Logf("📊 变更文件: %d, 受影响文件: %d, 变更组件: %d, 受影响组件: %d",
		len(output.Input.ChangedFiles),
		len(output.FileAnalysis.Impact),
		len(output.ComponentAnalysis.Changes),
		len(output.ComponentAnalysis.Impact))

	// ========================================================================
	// 验证结果（与 verify_flow.go 的预期一致）
	// ========================================================================

	// 1. 验证管道执行成功
	if !result.IsSuccessful() {
		t.Errorf("管道执行失败: %v", result.GetErrors())
	}

	// 2. 验证阶段结果
	if _, ok := result.GetResult("Diff解析"); !ok {
		t.Error("未找到 Diff解析 结果")
	}
	if _, ok := result.GetResult("项目解析"); !ok {
		t.Error("未找到项目解析 结果")
	}
	if _, ok := result.GetResult("符号分析"); !ok {
		t.Error("未找到符号分析 结果")
	}
	if _, ok := result.GetResult("影响分析（文件级）"); !ok {
		t.Error("未找到影响分析结果")
	}

	// 3. 验证项目解析成功
	if _, ok := result.GetResult("符号分析"); !ok {
		t.Fatal("未找到符号分析结果")
	}

	// 4. 验证检测到变更文件
	if len(output.FileAnalysis.Changes) == 0 {
		t.Error("未检测到变更文件")
	} else {
		t.Logf("✅ 检测到 %d 个变更文件", len(output.FileAnalysis.Changes))
	}

	// 5. 验证检测到间接受影响的文件
	if len(output.FileAnalysis.Impact) == 0 {
		t.Error("未检测到间接受影响的文件")
	} else {
		t.Logf("✅ 检测到 %d 个间接受影响的文件", len(output.FileAnalysis.Impact))
	}

	// 6. 验证组件库检测
	if !impactAnalysisResult.IsComponentLibrary {
		t.Error("未检测到组件库")
	} else {
		t.Logf("✅ 组件库检测成功")
		t.Logf("  - 变更组件: %d 个", len(output.ComponentAnalysis.Changes))
		t.Logf("  - 受影响组件: %d 个", len(output.ComponentAnalysis.Impact))
	}

	// 7. 验证特定文件包含预期符号
	expectedFiles := []string{
		"src/components/Button/Button.tsx",
		"src/components/Input/Input.tsx",
		"src/hooks/useDebounce.ts",
	}

	foundFiles := make(map[string]bool)
	for _, change := range output.FileAnalysis.Changes {
		for _, expected := range expectedFiles {
			if strings.HasSuffix(change.Path, expected) {
				foundFiles[expected] = true
				break
			}
		}
	}

	for _, expected := range expectedFiles {
		if !foundFiles[expected] {
			t.Errorf("未找到预期文件: %s", expected)
		}
	}
}

// =============================================================================
// 测试辅助类型定义（与 verify_flow.go 的 Output 结构对应）
// =============================================================================

// FileAnalysisMeta 文件分析元数据
type FileAnalysisMeta struct {
	TotalFileCount   int `json:"totalFileCount"`
	ChangedFileCount int `json:"changedFileCount"`
	ImpactFileCount  int `json:"impactFileCount"`
}

// FileChangeSimple 文件变更简化信息
type FileChangeSimple struct {
	Path        string `json:"path"`
	Type        string `json:"type"`
	SymbolCount int    `json:"symbolCount"`
}

// FileImpactSimple 文件影响简化信息
type FileImpactSimple struct {
	Path        string   `json:"path"`
	ImpactLevel int      `json:"impactLevel"`
	ImpactType  string   `json:"impactType"`
	ChangePaths []string `json:"changePaths"`
}

// ComponentAnalysisMeta 组件分析元数据
type ComponentAnalysisMeta struct {
	TotalComponentCount   int `json:"totalComponentCount"`
	ChangedComponentCount int `json:"changedComponentCount"`
	ImpactComponentCount  int `json:"impactComponentCount"`
}

// ComponentChange 组件变更信息
type ComponentChange struct {
	Name         string   `json:"name"`
	ChangedFiles []string `json:"changedFiles"`
	SymbolCount  int      `json:"symbolCount"`
}

// ComponentImpact 组件影响信息
type ComponentImpact struct {
	Name        string   `json:"name"`
	ImpactLevel int      `json:"impactLevel"`
	ImpactType  string   `json:"impactType"`
	ChangePaths []string `json:"changePaths"`
}

// =============================================================================
// export default () => {} 场景测试
// =============================================================================

// TestGitLabPipeline_ExportDefaultArrowFunction 测试 export default 箭头函数的场景
// 这是用户报告的问题：当改动在 export default () => {} 内部时，应该检测到符号变更
func TestGitLabPipeline_ExportDefaultArrowFunction(t *testing.T) {
	// 模拟的 git diff：修改了 export default () => {} 内部的一行
	// 使用 ButtonExportDefault.tsx 专门测试此场景
	const exportDefaultDiff = `diff --git a/testdata/test_project/src/components/Button/ButtonExportDefault.tsx b/testdata/test_project/src/components/Button/ButtonExportDefault.tsx
index 1234567..abcdefg 100644
--- a/testdata/test_project/src/components/Button/ButtonExportDefault.tsx
+++ b/testdata/test_project/src/components/Button/ButtonExportDefault.tsx
@@ -9,6 +9,6 @@
 export default () => {
-  return <button>Click</button>
+  return <button className="btn-primary">Click</button>
 }
`

	wd, _ := os.Getwd()
	projectRoot := filepath.Join(wd, "..", "..", "testdata", "test_project")
	absPath, _ := filepath.Abs(projectRoot)
	gitRoot := filepath.Join(wd, "..", "..")
	absGitRoot, _ := filepath.Abs(gitRoot)

	// 验证测试项目存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skip("测试项目不存在:", absPath)
	}

	t.Logf("📁 项目路径: %s", absPath)
	t.Logf("📁 Git 仓库根目录: %s", absGitRoot)
	t.Logf("📄 Git Diff: export default () => {} 内部变更")

	// 创建 GitLab 管道
	config := &GitLabPipelineConfig{
		DiffSource:   DiffSourceString,
		ProjectRoot:  absPath,
		GitRoot:      absGitRoot,
		MaxDepth:     10,
		// 不使用 manifest，只测试文件级分析
	}

	ctx := context.Background()
	analysisCtx := NewAnalysisContext(ctx, absPath, nil)
	analysisCtx.SetOption("diffString", exportDefaultDiff)

	pipeline := NewGitLabPipeline(config)

	// 执行管道
	result, err := pipeline.Execute(analysisCtx)
	if err != nil {
		t.Fatalf("管道执行失败: %v", err)
	}

	// 获取符号分析结果
	symbolResult, ok := result.GetResult("符号分析")
	if !ok {
		t.Fatal("未找到符号分析结果")
	}

	symbolResults, ok := symbolResult.(map[string]*symbol_analysis.FileAnalysisResult)
	if !ok {
		t.Fatalf("符号分析结果格式错误: %T", symbolResult)
	}

	// 验证：ButtonExportDefault.tsx 应该检测到符号变更
	// 注意：路径可能是绝对路径，需要灵活匹配
	var buttonResult *symbol_analysis.FileAnalysisResult

	for path, result := range symbolResults {
		if strings.HasSuffix(path, "src/components/Button/ButtonExportDefault.tsx") ||
		   strings.HasSuffix(path, "components/Button/ButtonExportDefault.tsx") {
			buttonResult = result
			break
		}
	}

	if buttonResult == nil {
		t.Errorf("未找到 ButtonExportDefault.tsx 的分析结果")
		t.Errorf("已分析的文件:")
		for path := range symbolResults {
			t.Errorf("  - %s", path)
		}
		return
	}

	t.Logf("ButtonExportDefault.tsx 分析结果:")
	t.Logf("  - IsSymbolFile: %v", buttonResult.IsSymbolFile)
	t.Logf("  - AffectedSymbols 数量: %d", len(buttonResult.AffectedSymbols))

	// 核心验证：应该检测到符号变更
	if len(buttonResult.AffectedSymbols) == 0 {
		t.Errorf("❌ 预期检测到符号变更，但得到 0 个")
		return
	}

	symbol := buttonResult.AffectedSymbols[0]
	t.Logf("  - 符号名称: %s", symbol.Name)
	t.Logf("  - 是否导出: %v", symbol.IsExported)
	t.Logf("  - 导出类型: %s", symbol.ExportType)

	// 核心验证：对于 export default () => {}，符号名应该是 "default"
	if symbol.Name != "default" {
		t.Errorf("预期符号名称为 'default'，但得到 '%s'（这是用户报告的问题）", symbol.Name)
	}

	// 验证符号已导出
	if !symbol.IsExported {
		t.Errorf("预期符号已导出，但 IsExported = false")
	}

	// 验证导出类型是 "default"
	if symbol.ExportType != symbol_analysis.ExportTypeDefault {
		t.Errorf("预期导出类型为 ExportTypeDefault，但得到 %v", symbol.ExportType)
	}

	// 获取影响分析结果
	impactResult, ok := result.GetResult("影响分析（文件级）")
	if !ok {
		t.Fatal("未找到影响分析结果")
	}

	impact, ok := impactResult.(*ImpactAnalysisResult)
	if !ok {
		t.Fatalf("影响分析结果格式错误: %T", impactResult)
	}

	// 验证：App.tsx 应该被检测为受影响的文件
	if impact.FileResult == nil {
		t.Error("未找到文件级影响分析结果")
		return
	}

	t.Logf("文件级影响分析:")
	t.Logf("  - 变更文件数: %d", impact.FileResult.Meta.ChangedFileCount)
	t.Logf("  - 受影响文件数: %d", impact.FileResult.Meta.ImpactFileCount)

	// 注意：ButtonExportDefault.tsx 没有被其他文件导入，所以不会有受影响文件
	if impact.FileResult.Meta.ImpactFileCount > 0 {
		t.Logf("受影响的文件:")
		for _, imp := range impact.FileResult.Impact {
			t.Logf("  - %s (层级: %d)", imp.Path, imp.ImpactLevel)
		}
	}

	t.Log("✅ export default 场景测试通过")
}
