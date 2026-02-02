package gitlab

import (
	"context"
	"fmt"
	"testing"
)

// =============================================================================
// Demo: GitLab 包使用示例
//
// 本文件演示如何使用 pkg/gitlab 包的各种能力
// 运行: go test -v -run TestDemo ./pkg/gitlab/
// =============================================================================

// DemoConfig 配置参数（请根据实际情况填写）
var DemoConfig = struct {
	// GitLab API 配置
	GitLabURL   string // 例如: "https://gitlab.example.com"
	GitLabToken string // GitLab Personal Access Token
	ProjectID   int    // 项目 ID
	MRIID       int    // Merge Request IID

	// Git 命令配置
	ProjectRoot string   // 项目根目录路径
	BaseSHA     string   // 基础 SHA（例如: "abc123"）
	HeadSHA     string   // 目标 SHA（例如: "def456"）
	DiffFile    string   // diff 文件路径
}{
	// ========== 请在此处填写您的配置 ==========
	// GitLabURL:   "https://gitlab.example.com",
	// GitLabToken: "your-token-here",
	// ProjectID:   123,
	// MRIID:       456,
	// ProjectRoot: "/path/to/your/project",
	// BaseSHA:     "abc123",
	// HeadSHA:     "def456",
	// DiffFile:    "/path/to/changes.patch",
}

// =============================================================================
// Demo 1: 从字符串解析 diff
// =============================================================================

// Demo1_ParseDiffFromString 从字符串解析 diff
func Demo1_ParseDiffFromString() {
	fmt.Println("=== Demo 1: 从字符串解析 diff ===")

	parser := NewParser("")

	diffContent := `diff --git a/src/Button.tsx b/src/Button.tsx
index 1234567..abcdefg 100644
--- a/src/Button.tsx
+++ b/src/Button.tsx
@@ -1,5 +1,7 @@
 export const Button = () => {
-  return <button>Click</button>;
+  return <button>{props.label}</button>;
 }
diff --git a/src/utils.ts b/src/utils.ts
index 1234567..abcdefg 100644
--- a/src/utils.ts
+++ b/src/utils.ts
@@ -1,3 +1,5 @@
 export const add = (a: number, b: number) => a + b;
+export const subtract = (a: number, b: number) => a - b;
+export const multiply = (a: number, b: number) => a * b;
`

	lineSet, err := parser.ParseDiffString(diffContent)
	if err != nil {
		fmt.Printf("❌ 解析失败: %v\n", err)
		return
	}

	// 打印结果
	fmt.Printf("✅ 解析成功！共 %d 个文件发生变更\n\n", len(lineSet))
	for file, lines := range lineSet {
		fmt.Printf("📄 %s\n", file)
		fmt.Printf("   变更行数: %d 行\n", len(lines))
		fmt.Printf("   行号: ")
		for line := range lines {
			fmt.Printf("%d ", line)
		}
		fmt.Println()
	}
}

// =============================================================================
// Demo 2: 从文件解析 diff
// =============================================================================

// Demo2_ParseDiffFromFile 从文件解析 diff
func Demo2_ParseDiffFromFile() {
	fmt.Println("=== Demo 2: 从文件解析 diff ===")

	if DemoConfig.DiffFile == "" {
		fmt.Println("⚠️  请先在 DemoConfig 中设置 DiffFile 参数")
		fmt.Println("   示例: DiffFile: \"/path/to/changes.patch\"")
		return
	}

	parser := NewParser(DemoConfig.ProjectRoot)

	lineSet, err := parser.ParseDiffFile(DemoConfig.DiffFile)
	if err != nil {
		fmt.Printf("❌ 解析失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 从文件解析成功！文件: %s\n", DemoConfig.DiffFile)
	fmt.Printf("   共 %d 个文件发生变更\n\n", len(lineSet))

	for file, lines := range lineSet {
		fmt.Printf("📄 %s: %d 行变更\n", file, len(lines))
	}
}

// =============================================================================
// Demo 3: 从 GitLab API 获取并解析 diff
// =============================================================================

// Demo3_ParseDiffFromGitLabAPI 从 GitLab API 获取 diff
func Demo3_ParseDiffFromGitLabAPI() {
	fmt.Println("=== Demo 3: 从 GitLab API 获取 diff ===")

	// 检查配置
	if DemoConfig.GitLabURL == "" || DemoConfig.GitLabToken == "" {
		fmt.Println("⚠️  请先在 DemoConfig 中设置 GitLab 配置:")
		fmt.Println("   GitLabURL:   \"https://gitlab.example.com\"")
		fmt.Println("   GitLabToken: \"your-token-here\"")
		fmt.Println("   ProjectID:   123")
		fmt.Println("   MRIID:       456")
		return
	}

	ctx := context.Background()

	// 创建客户端
	client := NewClient(DemoConfig.GitLabURL, DemoConfig.GitLabToken)

	// 获取 diff
	fmt.Printf("📡 正在获取 MR diff...\n")
	fmt.Printf("   URL: %s\n", DemoConfig.GitLabURL)
	fmt.Printf("   Project ID: %d\n", DemoConfig.ProjectID)
	fmt.Printf("   MR IID: %d\n\n", DemoConfig.MRIID)

	diffFiles, err := client.GetMergeRequestDiff(ctx, DemoConfig.ProjectID, DemoConfig.MRIID)
	if err != nil {
		fmt.Printf("❌ 获取 diff 失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 获取成功！共 %d 个文件\n\n", len(diffFiles))

	// 解析 diff
	parser := NewParser(DemoConfig.ProjectRoot)
	lineSet, err := parser.ParseDiffFiles(diffFiles)
	if err != nil {
		fmt.Printf("❌ 解析失败: %v\n", err)
		return
	}

	// 打印结果
	fmt.Println("📊 解析结果:")
	for file, lines := range lineSet {
		fmt.Printf("📄 %s\n", file)
		fmt.Printf("   变更行数: %d\n", len(lines))

		// 打印前 10 个行号
		count := 0
		for line := range lines {
			if count < 10 {
				fmt.Printf("   行 %d\n", line)
				count++
			}
		}
		if len(lines) > 10 {
			fmt.Printf("   ... (共 %d 行)\n", len(lines))
		}
		fmt.Println()
	}
}

// =============================================================================
// Demo 4: 从 Git 命令解析 diff
// =============================================================================

// Demo4_ParseDiffFromGitCommand 从 git 命令获取 diff
func Demo4_ParseDiffFromGitCommand() {
	fmt.Println("=== Demo 4: 从 Git 命令获取 diff ===")

	if DemoConfig.ProjectRoot == "" {
		fmt.Println("⚠️  请先在 DemoConfig 中设置 ProjectRoot 参数")
		fmt.Println("   示例: ProjectRoot: \"/path/to/your/project\"")
		return
	}

	if DemoConfig.BaseSHA == "" {
		fmt.Println("⚠️  请先在 DemoConfig 中设置 BaseSHA 参数")
		fmt.Println("   示例: BaseSHA: \"abc123\"")
		return
	}

	parser := NewParser(DemoConfig.ProjectRoot)

	headSHA := DemoConfig.HeadSHA
	if headSHA == "" {
		headSHA = "HEAD"
	}

	fmt.Printf("🔧 执行 git diff %s...%s\n", DemoConfig.BaseSHA[:8], headSHA)
	fmt.Printf("   项目目录: %s\n\n", DemoConfig.ProjectRoot)

	lineSet, err := parser.ParseFromGit(DemoConfig.BaseSHA, headSHA)
	if err != nil {
		fmt.Printf("❌ 执行 git diff 失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 解析成功！共 %d 个文件发生变更\n\n", len(lineSet))

	for file, lines := range lineSet {
		fmt.Printf("📄 %s: %d 行变更\n", file, len(lines))
	}
}

// =============================================================================
// Demo 5: 使用 Provider 接口
// =============================================================================

// Demo5_UseProviderInterface 使用 Provider 接口
func Demo5_UseProviderInterface() {
	fmt.Println("=== Demo 5: 使用 Provider 接口 ===")

	ctx := context.Background()
	parser := NewParser("")

	// 演示 StringProvider
	fmt.Println("📦 StringProvider:")
	stringProvider := NewStringProvider(`diff --git a/test.ts b/test.ts
@@ -1,1 +1,2 @@
-const a = 1;
+const a = 2;
+const b = 3;
`)
	lineSet, err := parser.ParseProvider(ctx, stringProvider)
	if err != nil {
		fmt.Printf("❌ 失败: %v\n", err)
	} else {
		fmt.Printf("✅ 成功: %v\n\n", lineSet)
	}

	// 演示 APIProvider（需要配置）
	if DemoConfig.GitLabURL != "" && DemoConfig.GitLabToken != "" {
		fmt.Println("📦 APIProvider:")
		client := NewClient(DemoConfig.GitLabURL, DemoConfig.GitLabToken)
		apiProvider := NewAPIProvider(client, DemoConfig.ProjectID, DemoConfig.MRIID)
		lineSet, err := parser.ParseProvider(ctx, apiProvider)
		if err != nil {
			fmt.Printf("❌ 失败: %v\n", err)
		} else {
			fmt.Printf("✅ 成功: %d 个文件\n\n", len(lineSet))
		}
	}

	// 演示 FileProvider（需要配置）
	if DemoConfig.DiffFile != "" {
		fmt.Println("📦 FileProvider:")
		fileProvider := NewFileProvider(DemoConfig.DiffFile)
		lineSet, err := parser.ParseProvider(ctx, fileProvider)
		if err != nil {
			fmt.Printf("❌ 失败: %v\n", err)
		} else {
			fmt.Printf("✅ 成功: %d 个文件\n\n", len(lineSet))
		}
	}
}

// =============================================================================
// Demo 6: 发布 MR 评论
// =============================================================================

// Demo6_PostMRComment 发布 MR 评论
func Demo6_PostMRComment() {
	fmt.Println("=== Demo 6: 发布 MR 评论 ===")

	if DemoConfig.GitLabURL == "" || DemoConfig.GitLabToken == "" {
		fmt.Println("⚠️  请先在 DemoConfig 中设置 GitLab 配置")
		return
	}

	ctx := context.Background()
	client := NewClient(DemoConfig.GitLabURL, DemoConfig.GitLabToken)
	service := NewService(client, DemoConfig.ProjectID, DemoConfig.MRIID)

	commentBody := `## 🤖 自动化分析报告

本评论由 analyzer-ts 自动生成。

### 分析摘要
- 变更文件: 5 个
- 变更行数: 23 行
- 风险等级: 低

---

💡 *提示: 本工具仅提供参考，请以实际代码审查为准*`

	fmt.Printf("📝 正在发布评论...\n")
	fmt.Printf("   URL: %s\n", DemoConfig.GitLabURL)
	fmt.Printf("   Project ID: %d\n", DemoConfig.ProjectID)
	fmt.Printf("   MR IID: %d\n\n", DemoConfig.MRIID)

	fmt.Println("📄 评论内容:")
	fmt.Println("---")
	fmt.Println(commentBody)
	fmt.Println("---")

	err := service.PostComment(ctx, commentBody)
	if err != nil {
		fmt.Printf("❌ 发布失败: %v\n", err)
		return
	}

	fmt.Println("✅ 评论发布成功！")
}

// =============================================================================
// Demo 7: 完整流程示例
// =============================================================================

// Demo7_CompleteFlow 完整的分析流程
func Demo7_CompleteFlow() {
	fmt.Println("=== Demo 7: 完整分析流程 ===")

	if DemoConfig.GitLabURL == "" || DemoConfig.GitLabToken == "" {
		fmt.Println("⚠️  请先配置 GitLab 参数")
		return
	}

	ctx := context.Background()

	// 步骤 1: 创建客户端
	fmt.Println("步骤 1️⃣: 创建 GitLab 客户端")
	client := NewClient(DemoConfig.GitLabURL, DemoConfig.GitLabToken)

	// 步骤 2: 获取 MR 详情
	fmt.Println("\n步骤 2️⃣: 获取 MR 详情")
	mr, err := client.GetMergeRequest(ctx, DemoConfig.ProjectID, DemoConfig.MRIID)
	if err != nil {
		fmt.Printf("❌ 获取 MR 详情失败: %v\n", err)
		return
	}
	fmt.Printf("✅ MR: !%d - %s\n", mr.IID, mr.Title)
	fmt.Printf("   源分支: %s → 目标分支: %s\n", mr.SourceBranch, mr.TargetBranch)

	// 步骤 3: 获取 diff
	fmt.Println("\n步骤 3️⃣: 获取 diff")
	diffFiles, err := client.GetMergeRequestDiff(ctx, DemoConfig.ProjectID, DemoConfig.MRIID)
	if err != nil {
		fmt.Printf("❌ 获取 diff 失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 获取到 %d 个文件的 diff\n", len(diffFiles))

	// 步骤 4: 解析 diff
	fmt.Println("\n步骤 4️⃣: 解析 diff")
	parser := NewParser(DemoConfig.ProjectRoot)
	lineSet, err := parser.ParseDiffFiles(diffFiles)
	if err != nil {
		fmt.Printf("❌ 解析失败: %v\n", err)
		return
	}

	totalFiles := len(lineSet)
	totalLines := 0
	for _, lines := range lineSet {
		totalLines += len(lines)
	}
	fmt.Printf("✅ 解析完成: %d 个文件，%d 行变更\n", totalFiles, totalLines)

	// 步骤 5: 生成报告
	fmt.Println("\n步骤 5️⃣: 生成报告")
	report := fmt.Sprintf(`## 📊 代码变更分析报告

### MR 信息
- **标题**: %s
- **分支**: %s → %s
- **链接**: %s

### 变更统计
- **变更文件**: %d 个
- **变更行数**: %d 行

### 变更文件列表
%s

---
*由 analyzer-ts 自动生成*
`,
		mr.Title,
		mr.SourceBranch,
		mr.TargetBranch,
		mr.WebURL,
		totalFiles,
		totalLines,
		formatFileList(lineSet),
	)

	// 打印报告预览
	fmt.Println("\n📄 报告预览:")
	fmt.Println("---")
	fmt.Println(report)
	fmt.Println("---")

	// 步骤 6: 发布评论（可选）
	fmt.Println("\n步骤 6️⃣: 发布评论到 MR")
	fmt.Println("⚠️  跳过实际发布（如需发布，请取消注释下面的代码）")
	// service := NewService(client, DemoConfig.ProjectID, DemoConfig.MRIID)
	// err = service.PostComment(ctx, report)
	// if err != nil {
	//     fmt.Printf("❌ 发布失败: %v\n", err)
	//     return
	// }
	// fmt.Println("✅ 评论发布成功！")

	fmt.Println("\n=== 流程完成 ===")
}

// formatFileList 格式化文件列表
func formatFileList(lineSet ChangedLineSetOfFiles) string {
	var result string
	for file, lines := range lineSet {
		result += fmt.Sprintf("- `%s`: %d 行\n", file, len(lines))
	}
	return result
}

// =============================================================================
// 测试入口
// =============================================================================

// TestDemo 运行所有 Demo
// 使用方法: go test -v -run TestDemo ./pkg/gitlab/
func TestDemo(t *testing.T) {
	fmt.Print(`
╔══════════════════════════════════════════════════════════════════════════════╗
║                    pkg/gitlab 使用演示                                       ║
╚══════════════════════════════════════════════════════════════════════════════╝
`)

	fmt.Println("\n运行演示前，请先在 example_test.go 中配置 DemoConfig 参数")

	// Demo 1: 字符串解析（无需配置）
	fmt.Println("\n" + stringsRepeat("=", 70))
	Demo1_ParseDiffFromString()

	// Demo 2: 文件解析（需要配置 DiffFile）
	fmt.Println("\n" + stringsRepeat("=", 70))
	Demo2_ParseDiffFromFile()

	// Demo 3: GitLab API（需要配置 GitLab 参数）
	fmt.Println("\n" + stringsRepeat("=", 70))
	Demo3_ParseDiffFromGitLabAPI()

	// Demo 4: Git 命令（需要配置 ProjectRoot 和 BaseSHA）
	fmt.Println("\n" + stringsRepeat("=", 70))
	Demo4_ParseDiffFromGitCommand()

	// Demo 5: Provider 接口
	fmt.Println("\n" + stringsRepeat("=", 70))
	Demo5_UseProviderInterface()

	// Demo 6: 发布评论
	fmt.Println("\n" + stringsRepeat("=", 70))
	Demo6_PostMRComment()

	// Demo 7: 完整流程
	fmt.Println("\n" + stringsRepeat("=", 70))
	Demo7_CompleteFlow()

	fmt.Println("\n" + stringsRepeat("=", 70))
	fmt.Println("\n✨ 所有演示完成！")
}

// =============================================================================
// 辅助函数
// =============================================================================

// stringsRepeat 重复字符串
func stringsRepeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// =============================================================================
// 独立运行的 Demo 函数
// =============================================================================

// DemoAll 运行所有 Demo（独立运行，非测试）
// 使用方法: 在 main 包中调用 gitlab.DemoAll()
func DemoAll() {
	fmt.Print(`
╔══════════════════════════════════════════════════════════════════════════════╗
║                    pkg/gitlab 使用演示                                       ║
╚══════════════════════════════════════════════════════════════════════════════╝
`)

	fmt.Println("\n运行演示前，请先在 example_test.go 中配置 DemoConfig 参数")

	// Demo 1
	fmt.Println("\n" + stringsRepeat("=", 70))
	Demo1_ParseDiffFromString()

	// 其他 Demo...
}
