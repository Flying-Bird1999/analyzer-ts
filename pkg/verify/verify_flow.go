// Package main 影响分析完整流程验证 - 接受 git diff 作为输入
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/Flying-Bird1999/analyzer-ts/pkg/gitlab"
	"github.com/Flying-Bird1999/analyzer-ts/pkg/impact_analysis"
	"github.com/Flying-Bird1999/analyzer-ts/pkg/impact_analysis/component_analyzer"
	"github.com/Flying-Bird1999/analyzer-ts/pkg/impact_analysis/file_analyzer"
	"github.com/Flying-Bird1999/analyzer-ts/pkg/symbol_analysis"
	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

// Output 最终输出结构
type Output struct {
	Input struct {
		ProjectPath    string             `json:"projectPath"`
		DiffFile       string             `json:"diffFile"`
		ComponentCount int                `json:"componentCount"`
		Components     []string           `json:"components"`
		ChangedFiles   []FileChangeSimple `json:"changedFiles"`
	} `json:"input"`

	SymbolAnalysis struct {
		Meta     SymbolAnalysisMetaSimple  `json:"meta"`
		Analysis []SymbolFileResultSimple  `json:"analysis"`
	} `json:"symbolAnalysis"`

	FileAnalysis struct {
		Meta    FileAnalysisMetaSimple `json:"meta"`
		Changes []FileChangeInfoSimple `json:"changes"`
		Impact  []FileImpactInfoSimple `json:"impact"`
	} `json:"fileAnalysis"`

	ComponentAnalysis struct {
		Meta    ComponentAnalysisMetaSimple `json:"meta"`
		Changes []ComponentChangeSimple     `json:"changes"`
		Impact  []ComponentImpactSimple     `json:"impact"`
	} `json:"componentAnalysis"`
}

type FileChangeSimple struct {
	Path         string `json:"path"`
	Type         string `json:"type"`
	ChangedLines []int  `json:"changedLines"`
}

type FileAnalysisMetaSimple struct {
	TotalFileCount   int `json:"totalFileCount"`
	ChangedFileCount int `json:"changedFileCount"`
	ImpactFileCount  int `json:"impactFileCount"`
}

type FileChangeInfoSimple struct {
	Path        string `json:"path"`
	Type        string `json:"type"`
	SymbolCount int    `json:"symbolCount"`
}

type FileImpactInfoSimple struct {
	Path        string   `json:"path"`
	ImpactLevel int      `json:"impactLevel"`
	ImpactType  string   `json:"impactType"`
	ChangePaths []string `json:"changePaths"`
}

type ComponentAnalysisMetaSimple struct {
	TotalComponentCount   int `json:"totalComponentCount"`
	ChangedComponentCount int `json:"changedComponentCount"`
	ImpactComponentCount  int `json:"impactComponentCount"`
}

type ComponentChangeSimple struct {
	Name         string   `json:"name"`
	Entry        string   `json:"entry"`
	ChangedFiles []string `json:"changedFiles"`
	SymbolCount  int      `json:"symbolCount"`
}

type ComponentImpactSimple struct {
	Name        string   `json:"name"`
	ImpactLevel int      `json:"impactLevel"`
	ImpactType  string   `json:"impactType"`
	ChangePaths []string `json:"changePaths"`
}

type SymbolAnalysisMetaSimple struct {
	AnalyzedFileCount int `json:"analyzedFileCount"`
	AffectedFileCount int `json:"affectedFileCount"`
	TotalSymbolCount  int `json:"totalSymbolCount"`
}

type SymbolFileResultSimple struct {
	FilePath         string                  `json:"filePath"`
	FileType         string                  `json:"fileType"`
	IsSymbolFile     bool                    `json:"isSymbolFile"`
	AffectedSymbols  []SymbolChangeSimple    `json:"affectedSymbols"`
	TotalSymbolCount int                     `json:"totalSymbolCount"`
	ChangedLines     []int                   `json:"changedLines"`
}

type SymbolChangeSimple struct {
	Name         string   `json:"name"`
	Kind         string   `json:"kind"`
	StartLine    int      `json:"startLine"`
	EndLine      int      `json:"endLine"`
	ChangedLines []int    `json:"changedLines"`
	ChangeType   string   `json:"changeType"`
	IsExported   bool     `json:"isExported"`
	ExportType   string   `json:"exportType"`
}

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
diff --git a/testdata/test_project/src/assets/logo.png b/testdata/test_project/src/assets/logo.png
index 1234567..abcdefg 100644
Binary files a/testdata/test_project/src/assets/logo.png and b/testdata/test_project/src/assets/logo.png differ
diff --git a/testdata/test_project/src/assets/modal.css b/testdata/test_project/src/assets/modal.css
new file mode 100644
index 0000000..1234567
--- /dev/null
+++ b/testdata/test_project/src/assets/modal.css
@@ -0,0 +1,13 @@
+/* Modal 组件样式 */
+.modal-overlay {
+  position: fixed;
+  top: 0;
+  left: 0;
+  right: 0;
+  bottom: 0;
+  background: rgba(0, 0, 0, 0.5);
+}
+
+.modal-content {
+  position: fixed;
+  top: 50%;
+  left: 50%;
+  transform: translate(-50%, -50%);
+  background: white;
+  padding: 20px;
+  border-radius: 8px;
+}
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

func main() {
	// 项目路径
	wd, _ := os.Getwd()
	projectRoot := filepath.Join(wd, "..", "..", "testdata", "test_project")
	absPath, _ := filepath.Abs(projectRoot)

	// git 仓库根目录（diff 路径是相对于 git 仓库根的）
	gitRoot := filepath.Join(wd, "..", "..")
	absGitRoot, _ := filepath.Abs(gitRoot)

	fmt.Printf("📁 项目路径: %s\n", absPath)
	fmt.Printf("📁 Git 仓库根目录: %s\n", absGitRoot)
	fmt.Printf("📄 Git Diff: 内置测试用例 (Button + Input + useDebounce + 二进制文件 + 新增CSS + 枚举类型)\n\n")

	// ============================================================
	// 1. 加载组件清单
	// ============================================================
	manifestPath := filepath.Join(absPath, ".analyzer", "component-manifest.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		fmt.Printf("❌ 读取组件清单失败: %v\n", err)
		os.Exit(1)
	}

	var manifest impact_analysis.ComponentManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		fmt.Printf("❌ 解析组件清单失败: %v\n", err)
		os.Exit(1)
	}

	// 转换为绝对路径
	componentNames := make([]string, len(manifest.Components))
	for i := range manifest.Components {
		if !filepath.IsAbs(manifest.Components[i].Entry) {
			manifest.Components[i].Entry = filepath.Join(absPath, manifest.Components[i].Entry)
		}
		componentNames[i] = manifest.Components[i].Name
	}
	sort.Strings(componentNames)

	fmt.Printf("📦 组件总数: %d\n", len(manifest.Components))
	fmt.Printf("📋 组件列表: %s\n\n", strings.Join(componentNames, ", "))

	// ============================================================
	// 2. 解析项目
	// ============================================================
	config := projectParser.NewProjectParserConfig(absPath, nil, false, nil)
	parsingResult := projectParser.NewProjectParserResult(config)
	parsingResult.ProjectParser()

	// ============================================================
	// 3. 创建符号分析项目
	// ============================================================
	tsProject := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:   absPath,
		UseTsConfig: true,
	})

	symbolAnalyzer := symbol_analysis.NewAnalyzerWithDefaults(tsProject)

	// ============================================================
	// 4. 解析 Git Diff
	// ============================================================
	diffParser := gitlab.NewDiffParser(absGitRoot)
	changedLineSet, err := diffParser.ParseDiffOutput(testGitDiff)
	if err != nil {
		fmt.Printf("❌ 解析 git diff 失败: %v\n", err)
		os.Exit(1)
	}

	// ============================================================
	// 5. 执行符号分析
	// ============================================================
	// changedLineSet 的键是相对于 git 仓库根的路径（如 "testdata/test_project/src/components/Button/Button.tsx"）
	// 但 symbol_analysis 期望绝对路径，需要转换
	absChangedLineSet := make(symbol_analysis.ChangedLineSetOfFiles)
	for filePath, lines := range changedLineSet {
		absFilePath := filepath.Join(absGitRoot, filePath)
		absChangedLineSet[absFilePath] = lines
	}

	symbolResults := symbolAnalyzer.AnalyzeChangedLines(absChangedLineSet)

	// 从符号分析结果构建 ChangedSymbol（用于文件级影响分析）
	// 同时保留每个文件的 changedLines 信息
	type FileInfo struct {
		AbsPath      string
		ChangedLines []int
	}
	fileInfoMap := make(map[string]FileInfo)

	changedSymbols := make([]file_analyzer.ChangedSymbol, 0)
	changedNonSymbolFiles := make([]string, 0) // 非符号文件列表

	for filePath, result := range symbolResults {
		// filePath 已经是绝对路径了
		absFilePath := filePath

		// 保存文件的变更行信息（直接使用 symbol_analysis 的结果）
		fileInfoMap[absFilePath] = FileInfo{
			AbsPath:      absFilePath,
			ChangedLines: result.ChangedLines,
		}

		// 根据文件类型分别处理
		if result.IsSymbolFile {
			// 为每个受影响的符号创建 ChangedSymbol 条目
			for _, sym := range result.AffectedSymbols {
				// 只处理导出的符号（因为只有导出的符号才能被其他文件导入）
				if sym.IsExported {
					changedSymbols = append(changedSymbols, file_analyzer.ChangedSymbol{
						Name:       sym.Name,
						FilePath:   absFilePath,
						ExportType: sym.ExportType,
					})
				}
			}

			// 如果没有导出符号，为文件本身创建一个条目
			if len(result.AffectedSymbols) > 0 && len(changedSymbols) == 0 {
				symName := extractSymbolNameFromPath(absFilePath)
				changedSymbols = append(changedSymbols, file_analyzer.ChangedSymbol{
					Name:       symName,
					FilePath:   absFilePath,
					ExportType: symbol_analysis.ExportTypeDefault,
				})
			}
		} else {
			// 非符号文件：添加到非符号文件列表
			changedNonSymbolFiles = append(changedNonSymbolFiles, absFilePath)
		}
	}

	// ============================================================
	// 6. 执行文件级影响分析
	// ============================================================
	fileAnalyzer := file_analyzer.NewAnalyzer(parsingResult)
	fileInput := &file_analyzer.Input{
		ChangedSymbols:        changedSymbols,
		ChangedNonSymbolFiles: changedNonSymbolFiles,
	}

	fileResult, err := fileAnalyzer.Analyze(fileInput)
	if err != nil {
		fmt.Printf("❌ 文件级分析失败: %v\n", err)
		os.Exit(1)
	}

	// ============================================================
	// 5. 执行组件级影响分析
	// ============================================================
	compInput := &component_analyzer.Input{
		FileResult: &component_analyzer.FileAnalysisResultProxy{
			Changes:      convertFileChangeInfos(fileResult.Changes),
			Impact:       convertFileImpactInfos(fileResult.Impact),
			DepGraph:     buildFileDepGraph(parsingResult),
			RevDepGraph:  buildFileRevDepGraph(parsingResult),
			ExternalDeps: buildExternalDeps(parsingResult),
		},
	}

	compAnalyzer := component_analyzer.NewAnalyzer(&manifest, parsingResult, 10)
	compResult, err := compAnalyzer.Analyze(compInput)
	if err != nil {
		fmt.Printf("❌ 组件级分析失败: %v\n", err)
		os.Exit(1)
	}

	// ============================================================
	// 6. 输出 JSON 结果
	// ============================================================
	output := &Output{}
	output.Input.ProjectPath = absPath
	output.Input.DiffFile = "内置测试用例 (Button + Input + useDebounce + 二进制文件 + 新增CSS + 枚举类型)"
	output.Input.ComponentCount = len(manifest.Components)
	output.Input.Components = componentNames

	// 构建输入文件列表（使用 fileInfoMap 获取 changedLines）
	for absFilePath, info := range fileInfoMap {
		relPath, _ := filepath.Rel(absPath, absFilePath)
		output.Input.ChangedFiles = append(output.Input.ChangedFiles, FileChangeSimple{
			Path:         relPath,
			Type:         "modified",
			ChangedLines: info.ChangedLines,
		})
	}
	sort.Slice(output.Input.ChangedFiles, func(i, j int) bool {
		return output.Input.ChangedFiles[i].Path < output.Input.ChangedFiles[j].Path
	})

	// ============================================================
	// 符号分析结果
	// ============================================================
	totalSymbolCount := 0
	symbolFileCount := 0
	nonSymbolFileCount := 0
	for filePath, result := range symbolResults {
		relPath, _ := filepath.Rel(absGitRoot, filePath)
		symbols := make([]SymbolChangeSimple, 0)
		for _, sym := range result.AffectedSymbols {
			symbols = append(symbols, SymbolChangeSimple{
				Name:         sym.Name,
				Kind:         string(sym.Kind),
				StartLine:    sym.StartLine,
				EndLine:      sym.EndLine,
				ChangedLines: sym.ChangedLines,
				ChangeType:   string(sym.ChangeType),
				IsExported:   sym.IsExported,
				ExportType:   string(sym.ExportType),
			})
		}
		totalSymbolCount += len(result.AffectedSymbols)

		if result.IsSymbolFile {
			symbolFileCount++
		} else {
			nonSymbolFileCount++
		}

		output.SymbolAnalysis.Analysis = append(output.SymbolAnalysis.Analysis, SymbolFileResultSimple{
			FilePath:         relPath,
			FileType:         string(result.FileType),
			IsSymbolFile:     result.IsSymbolFile,
			AffectedSymbols:  symbols,
			TotalSymbolCount: len(result.AffectedSymbols),
			ChangedLines:     result.ChangedLines,
		})
	}
	sort.Slice(output.SymbolAnalysis.Analysis, func(i, j int) bool {
		return output.SymbolAnalysis.Analysis[i].FilePath < output.SymbolAnalysis.Analysis[j].FilePath
	})

	output.SymbolAnalysis.Meta = SymbolAnalysisMetaSimple{
		AnalyzedFileCount: len(symbolResults),
		AffectedFileCount: len(symbolResults),
		TotalSymbolCount:  totalSymbolCount,
	}

	// ============================================================
	// 文件级分析结果
	// ============================================================
	output.FileAnalysis.Meta = FileAnalysisMetaSimple{
		TotalFileCount:   fileResult.Meta.TotalFileCount,
		ChangedFileCount: fileResult.Meta.ChangedFileCount,
		ImpactFileCount:  fileResult.Meta.ImpactFileCount,
	}

	for _, change := range fileResult.Changes {
		relPath, _ := filepath.Rel(absPath, change.Path)
		output.FileAnalysis.Changes = append(output.FileAnalysis.Changes, FileChangeInfoSimple{
			Path:        relPath,
			Type:        change.ChangeType,
			SymbolCount: change.SymbolCount,
		})
	}
	sort.Slice(output.FileAnalysis.Changes, func(i, j int) bool {
		return output.FileAnalysis.Changes[i].Path < output.FileAnalysis.Changes[j].Path
	})

	for _, impact := range fileResult.Impact {
		relPath, _ := filepath.Rel(absPath, impact.Path)
		changePaths := make([]string, len(impact.ChangePaths))
		for i, p := range impact.ChangePaths {
			changePaths[i], _ = filepath.Rel(absPath, p)
		}
		output.FileAnalysis.Impact = append(output.FileAnalysis.Impact, FileImpactInfoSimple{
			Path:        relPath,
			ImpactLevel: impact.ImpactLevel,
			ImpactType:  impact.ImpactType,
			ChangePaths: changePaths,
		})
	}
	sort.Slice(output.FileAnalysis.Impact, func(i, j int) bool {
		return output.FileAnalysis.Impact[i].Path < output.FileAnalysis.Impact[j].Path
	})

	output.ComponentAnalysis.Meta = ComponentAnalysisMetaSimple{
		TotalComponentCount:   compResult.Meta.TotalComponentCount,
		ChangedComponentCount: compResult.Meta.ChangedComponentCount,
		ImpactComponentCount:  compResult.Meta.ImpactComponentCount,
	}

	for _, change := range compResult.Changes {
		// 从 manifest 中获取组件 entry
		var entryRel string
		for _, comp := range manifest.Components {
			if comp.Name == change.Name {
				entryRel, _ = filepath.Rel(absPath, comp.Entry)
				break
			}
		}
		changedFiles := make([]string, len(change.ChangedFiles))
		for i, f := range change.ChangedFiles {
			changedFiles[i], _ = filepath.Rel(absPath, f)
		}
		output.ComponentAnalysis.Changes = append(output.ComponentAnalysis.Changes, ComponentChangeSimple{
			Name:         change.Name,
			Entry:        entryRel,
			ChangedFiles: changedFiles,
			SymbolCount:  change.SymbolCount,
		})
	}
	sort.Slice(output.ComponentAnalysis.Changes, func(i, j int) bool {
		return output.ComponentAnalysis.Changes[i].Name < output.ComponentAnalysis.Changes[j].Name
	})

	for _, impact := range compResult.Impact {
		changePaths := make([]string, len(impact.ChangePaths))
		for i, p := range impact.ChangePaths {
			changePaths[i], _ = filepath.Rel(absPath, p)
		}
		output.ComponentAnalysis.Impact = append(output.ComponentAnalysis.Impact, ComponentImpactSimple{
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

	// 输出 JSON 到文件
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Printf("❌ JSON 序列化失败: %v\n", err)
		os.Exit(1)
	}

	// 输出文件保存到当前工作目录
	outputFile := filepath.Join(wd, "verify_output.json")

	// 写入文件
	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		fmt.Printf("❌ 写入输出文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ 分析完成！")
	fmt.Printf("📄 输出文件: %s\n", outputFile)
	fmt.Printf("📊 变更文件: %d, 受影响文件: %d, 变更组件: %d, 受影响组件: %d\n",
		len(output.Input.ChangedFiles),
		len(output.FileAnalysis.Impact),
		len(output.ComponentAnalysis.Changes),
		len(output.ComponentAnalysis.Impact))
}

// 辅助转换函数
func convertFileChangeInfos(changes []file_analyzer.FileChangeInfo) []component_analyzer.FileChangeInfoProxy {
	result := make([]component_analyzer.FileChangeInfoProxy, len(changes))
	for i, c := range changes {
		result[i] = component_analyzer.FileChangeInfoProxy{
			Path:        c.Path,
			ChangeType:  impact_analysis.ChangeType(c.ChangeType),
			SymbolCount: c.SymbolCount,
		}
	}
	return result
}

func convertFileImpactInfos(impacts []file_analyzer.FileImpactInfo) []component_analyzer.FileImpactInfoProxy {
	result := make([]component_analyzer.FileImpactInfoProxy, len(impacts))
	for i, imp := range impacts {
		result[i] = component_analyzer.FileImpactInfoProxy{
			Path:        imp.Path,
			ImpactLevel: impact_analysis.ImpactLevel(imp.ImpactLevel),
			ImpactType:  impact_analysis.ImpactType(imp.ImpactType),
			ChangePaths: imp.ChangePaths,
		}
	}
	return result
}

// extractSymbolNameFromPath 从文件路径提取符号名称
// 例如: src/components/Button/Button.tsx -> Button
func extractSymbolNameFromPath(filePath string) string {
	// 获取文件名（不含扩展名）
	base := filepath.Base(filePath)
	parts := strings.Split(base, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return "Unknown"
}

func buildFileDepGraph(result *projectParser.ProjectParserResult) map[string][]string {
	depGraph := make(map[string][]string)
	for sourceFile, fileResult := range result.Js_Data {
		for _, imp := range fileResult.ImportDeclarations {
			if imp.Source.FilePath != "" {
				depGraph[sourceFile] = append(depGraph[sourceFile], imp.Source.FilePath)
			}
		}
	}
	return depGraph
}

func buildFileRevDepGraph(result *projectParser.ProjectParserResult) map[string][]string {
	revDepGraph := make(map[string][]string)
	for sourceFile, fileResult := range result.Js_Data {
		for _, imp := range fileResult.ImportDeclarations {
			if imp.Source.FilePath != "" {
				revDepGraph[imp.Source.FilePath] = append(revDepGraph[imp.Source.FilePath], sourceFile)
			}
		}
	}
	return revDepGraph
}

func buildExternalDeps(result *projectParser.ProjectParserResult) map[string][]string {
	externalDeps := make(map[string][]string)
	for sourceFile, fileResult := range result.Js_Data {
		for _, imp := range fileResult.ImportDeclarations {
			if imp.Source.NpmPkg != "" {
				externalDeps[sourceFile] = append(externalDeps[sourceFile], imp.Source.NpmPkg)
			}
		}
	}
	return externalDeps
}
