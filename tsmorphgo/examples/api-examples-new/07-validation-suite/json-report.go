// +build validation-suite

package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)


// ReportGenerator JSON报告生成器
type ReportGenerator struct {
	outputDir string
	verbose   bool
}

// NewReportGenerator 创建新的报告生成器
func NewReportGenerator(outputDir string, verbose bool) *ReportGenerator {
	return &ReportGenerator{
		outputDir: outputDir,
		verbose:   verbose,
	}
}

// GenerateReport 生成综合验证报告
func (rg *ReportGenerator) GenerateReport(suite *ValidationSuite, project *tsmorphgo.Project, config *ValidationConfig) error {
	timestamp := time.Now().Format("20060102-150405")

	// 生成主报告
	mainReport := rg.generateMainReport(suite, project, config)
	mainReportPath := filepath.Join(rg.outputDir, fmt.Sprintf("validation-report-%s.json", timestamp))

	if err := SaveTestResults(mainReport, mainReportPath); err != nil {
		return fmt.Errorf("保存主报告失败: %w", err)
	}

	// 生成分类报告
	if err := rg.generateCategoryReports(suite, timestamp); err != nil {
		return fmt.Errorf("生成分类报告失败: %w", err)
	}

	// 生成摘要报告
	if err := rg.generateSummaryReport(suite, project, timestamp); err != nil {
		return fmt.Errorf("生成摘要报告失败: %w", err)
	}

	if rg.verbose {
		fmt.Printf("📊 报告已生成到: %s\n", mainReportPath)
	}

	return nil
}

// MainReport 主报告结构
type MainReport struct {
	Metadata    *ReportMetadata   `json:"metadata"`
	Suite       *ValidationSuite  `json:"suite"`
	ProjectInfo *ProjectInfo      `json:"projectInfo"`
	Config      *ValidationConfig `json:"config"`
	Analysis    *ReportAnalysis   `json:"analysis"`
	Timestamp   time.Time         `json:"timestamp"`
}

// ReportMetadata 报告元数据
type ReportMetadata struct {
	ReportID     string    `json:"reportId"`
	GeneratedAt  time.Time `json:"generatedAt"`
	GeneratedBy  string    `json:"generatedBy"`
	Version      string    `json:"version"`
	Format       string    `json:"format"`
	TotalTests   int       `json:"totalTests"`
	TestDuration string    `json:"testDuration"`
}

// ReportAnalysis 报告分析
type ReportAnalysis struct {
	OverallHealth    string                       `json:"overallHealth"`
	CriticalIssues   []*AnalysisIssue             `json:"criticalIssues"`
	Recommendations  []*Recommendation            `json:"recommendations"`
	CategoryAnalysis map[string]*CategoryAnalysis `json:"categoryAnalysis"`
	TrendAnalysis    map[string]*TrendData        `json:"trendAnalysis"`
}

// AnalysisIssue 分析问题
type AnalysisIssue struct {
	Type        string                 `json:"type"`     // "critical", "warning", "info"
	Severity    string                 `json:"severity"` // "high", "medium", "low"
	Category    string                 `json:"category"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details"`
}

// Recommendation 推荐建议
type Recommendation struct {
	Priority string `json:"priority"` // "high", "medium", "low"`
	Category string `json:"category"`
	Title    string `json:"title"`
	Action   string `json:"action"`
	Impact   string `json:"impact"`
}

// CategoryAnalysis 类别分析
type CategoryAnalysis struct {
	Category         string            `json:"category"`
	TestCount        int               `json:"testCount"`
	PassRate         float64           `json:"passRate"`
	TotalDuration    float64           `json:"totalDuration"`    // 毫秒
	PerformanceScore float64           `json:"performanceScore"` // 0-100
	StabilityScore   float64           `json:"stabilityScore"`   // 0-100
	Recommendations  []*Recommendation `json:"recommendations"`
}

// TrendData 趋势数据
type TrendData struct {
	Current float64 `json:"current"`
	Target  float64 `json:"target"`
	Trend   string  `json:"trend"` // "improving", "stable", "declining"
}

// generateMainReport 生成主报告
func (rg *ReportGenerator) generateMainReport(suite *ValidationSuite, project *tsmorphgo.Project, config *ValidationConfig) *MainReport {
	return &MainReport{
		Metadata:    rg.generateMetadata(suite),
		Suite:       suite,
		ProjectInfo: rg.extractProjectInfo(project, config),
		Config:      config,
		Analysis:    rg.analyzeResults(suite),
		Timestamp:   time.Now(),
	}
}

// generateMetadata 生成报告元数据
func (rg *ReportGenerator) generateMetadata(suite *ValidationSuite) *ReportMetadata {
	testDuration := suite.Duration.String()
	return &ReportMetadata{
		ReportID:     fmt.Sprintf("val-%d", time.Now().Unix()),
		GeneratedAt:  time.Now(),
		GeneratedBy:  "TSMorphGo Validation Suite",
		Version:      "1.0.0",
		Format:       "json",
		TotalTests:   suite.Summary.TotalTests,
		TestDuration: testDuration,
	}
}

// extractProjectInfo 提取项目信息
func (rg *ReportGenerator) extractProjectInfo(project *tsmorphgo.Project, config *ValidationConfig) *ProjectInfo {
	// 收集源文件统计信息
	sourceFiles := project.GetSourceFiles()
	totalNodes := 0
	totalSymbols := 0

	// 统计节点和符号数量（示例实现）
	for _, sf := range sourceFiles {
		sf.ForEachDescendant(func(node tsmorphgo.Node) {
			totalNodes++
		})
		// 这里可以添加符号统计逻辑
	}

	return &ProjectInfo{
		Path:           config.ProjectPath,
		SourceFiles:    len(sourceFiles),
		TotalNodes:     totalNodes,
		TotalSymbols:   totalSymbols,
		APIVersions:    map[string]string{"tsmorphgo": "current"},
		FileExtensions: config.TargetExtensions,
		IgnorePatterns: config.IgnorePatterns,
	}
}

// analyzeResults 分析验证结果
func (rg *ReportGenerator) analyzeResults(suite *ValidationSuite) *ReportAnalysis {
	analysis := &ReportAnalysis{
		OverallHealth:    rg.calculateOverallHealth(suite),
		CriticalIssues:   rg.identifyCriticalIssues(suite),
		Recommendations:  rg.generateRecommendations(suite),
		CategoryAnalysis: rg.analyzeCategories(suite),
		TrendAnalysis:    rg.analyzeTrends(suite),
	}

	return analysis
}

// calculateOverallHealth 计算整体健康度
func (rg *ReportGenerator) calculateOverallHealth(suite *ValidationSuite) string {
	if suite.Summary.PassRate >= 95.0 {
		return "excellent"
	} else if suite.Summary.PassRate >= 80.0 {
		return "good"
	} else if suite.Summary.PassRate >= 60.0 {
		return "fair"
	} else {
		return "poor"
	}
}

// identifyCriticalIssues 识别关键问题
func (rg *ReportGenerator) identifyCriticalIssues(suite *ValidationSuite) []*AnalysisIssue {
	issues := make([]*AnalysisIssue, 0)

	// 检查失败率过高的类别
	for category := range suite.Summary.CategoryStats {
		categoryTests := rg.getTestsByCategory(suite, category)
		if len(categoryTests) > 0 {
			failures := rg.countFailedTests(categoryTests)
			failRate := float64(failures) / float64(len(categoryTests)) * 100

			if failRate >= 50.0 {
				issues = append(issues, &AnalysisIssue{
					Type:        "critical",
					Severity:    "high",
					Category:    category,
					Title:       "高失败率类别",
					Description: fmt.Sprintf("类别 %s 的失败率 %.1f%% 过高", category, failRate),
					Details: map[string]interface{}{
						"totalTests":  len(categoryTests),
						"failedTests": failures,
						"failRate":    failRate,
					},
				})
			}
		}
	}

	return issues
}

// generateRecommendations 生成推荐建议
func (rg *ReportGenerator) generateRecommendations(suite *ValidationSuite) []*Recommendation {
	recommendations := make([]*Recommendation, 0)

	// 基于通过率生成建议
	if suite.Summary.PassRate < 80.0 {
		recommendations = append(recommendations, &Recommendation{
			Priority: "high",
			Category: "general",
			Title:    "提高整体测试通过率",
			Action:   "检查失败测试并修复相关问题",
			Impact:   "显著提高API稳定性",
		})
	}

	// 基于性能生成建议
	if suite.Summary.TotalDuration > 5*time.Minute {
		recommendations = append(recommendations, &Recommendation{
			Priority: "medium",
			Category: "performance",
			Title:    "优化测试性能",
			Action:   "检查性能瓶颈并优化测试执行时间",
			Impact:   "减少验证时间，提高开发效率",
		})
	}

	return recommendations
}

// analyzeCategories 分析各个类别
func (rg *ReportGenerator) analyzeCategories(suite *ValidationSuite) map[string]*CategoryAnalysis {
	analysis := make(map[string]*CategoryAnalysis)

	for category := range suite.Summary.CategoryStats {
		categoryTests := rg.getTestsByCategory(suite, category)
		passed := rg.countPassedTests(categoryTests)
		passRate := 0.0
		if len(categoryTests) > 0 {
			passRate = float64(passed) / float64(len(categoryTests)) * 100
		}

		// 计算总执行时间
		totalDuration := 0.0
		for _, test := range categoryTests {
			totalDuration += float64(test.Duration.Milliseconds())
		}

		analysis[category] = &CategoryAnalysis{
			Category:         category,
			TestCount:        len(categoryTests),
			PassRate:         passRate,
			TotalDuration:    totalDuration,
			PerformanceScore: rg.calculatePerformanceScore(categoryTests),
			StabilityScore:   rg.calculateStabilityScore(categoryTests),
			Recommendations:  rg.generateCategoryRecommendations(category, passRate, totalDuration),
		}
	}

	return analysis
}

// analyzeTrends 分析趋势（简化版本）
func (rg *ReportGenerator) analyzeTrends(suite *ValidationSuite) map[string]*TrendData {
	trends := make(map[string]*TrendData)

	// 基于当前通过率设置趋势
	currentRate := suite.Summary.PassRate
	var trend string
	if currentRate >= 90.0 {
		trend = "improving"
	} else if currentRate >= 70.0 {
		trend = "stable"
	} else {
		trend = "declining"
	}

	trends["passRate"] = &TrendData{
		Current: currentRate,
		Target:  95.0,
		Trend:   trend,
	}

	return trends
}

// Helper functions

func (rg *ReportGenerator) getTestsByCategory(suite *ValidationSuite, category string) []*ValidationResult {
	tests := make([]*ValidationResult, 0)
	for _, test := range suite.Tests {
		if test.Category == category {
			tests = append(tests, test)
		}
	}
	return tests
}

func (rg *ReportGenerator) countFailedTests(tests []*ValidationResult) int {
	count := 0
	for _, test := range tests {
		if test.Status == "failed" {
			count++
		}
	}
	return count
}

func (rg *ReportGenerator) countPassedTests(tests []*ValidationResult) int {
	count := 0
	for _, test := range tests {
		if test.Status == "passed" {
			count++
		}
	}
	return count
}

func (rg *ReportGenerator) calculatePerformanceScore(tests []*ValidationResult) float64 {
	if len(tests) == 0 {
		return 0.0
	}

	totalScore := 0.0
	for _, test := range tests {
		durationMs := float64(test.Duration.Milliseconds())
		score := 100.0
		if durationMs > 1000.0 {
			score = 80.0
		}
		if durationMs > 5000.0 {
			score = 60.0
		}
		if durationMs > 10000.0 {
			score = 40.0
		}
		totalScore += score
	}

	return totalScore / float64(len(tests))
}

func (rg *ReportGenerator) calculateStabilityScore(tests []*ValidationResult) float64 {
	if len(tests) == 0 {
		return 0.0
	}

	passed := rg.countPassedTests(tests)
	return float64(passed) / float64(len(tests)) * 100
}

func (rg *ReportGenerator) generateCategoryRecommendations(category string, passRate float64, duration float64) []*Recommendation {
	recommendations := make([]*Recommendation, 0)

	if passRate < 80.0 {
		recommendations = append(recommendations, &Recommendation{
			Priority: "high",
			Category: category,
			Title:    "提高类别通过率",
			Action:   "检查失败测试并修复API问题",
			Impact:   "提高API稳定性",
		})
	}

	if duration > 10000.0 {
		recommendations = append(recommendations, &Recommendation{
			Priority: "medium",
			Category: category,
			Title:    "优化执行性能",
			Action:   "优化测试逻辑或减少测试范围",
			Impact:   "减少执行时间",
		})
	}

	return recommendations
}

// generateCategoryReports 生成分类报告
func (rg *ReportGenerator) generateCategoryReports(suite *ValidationSuite, timestamp string) error {
	for category := range suite.Summary.CategoryStats {
		categoryTests := rg.getTestsByCategory(suite, category)
		categoryReport := map[string]interface{}{
			"category":   category,
			"timestamp":  timestamp,
			"totalTests": len(categoryTests),
			"tests":      categoryTests,
		}

		reportPath := filepath.Join(rg.outputDir, fmt.Sprintf("category-%s-report-%s.json", category, timestamp))
		if err := SaveTestResults(categoryReport, reportPath); err != nil {
			return err
		}
	}
	return nil
}

// generateSummaryReport 生成摘要报告
func (rg *ReportGenerator) generateSummaryReport(suite *ValidationSuite, project *tsmorphgo.Project, timestamp string) error {
	summaryReport := map[string]interface{}{
		"timestamp":       timestamp,
		"summary":         suite.Summary,
		"health":          rg.calculateOverallHealth(suite),
		"recommendations": rg.generateRecommendations(suite),
	}

	reportPath := filepath.Join(rg.outputDir, fmt.Sprintf("summary-report-%s.json", timestamp))
	return SaveTestResults(summaryReport, reportPath)
}
