// +build validation-suite

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ValidationResult 单个验证测试的结果
type ValidationResult struct {
	Name          string        `json:"name"`          // 测试名称
	Category      string        `json:"category"`      // 测试类别
	Description   string        `json:"description"`   // 测试描述
	Status        string        `json:"status"`        // 测试状态 (passed/failed/skipped)
	Message       string        `json:"message"`       // 测试消息
	Error         string        `json:"error"`         // 错误信息（如果有）
	Duration      time.Duration `json:"duration"`      // 执行时间
	Timestamp     time.Time     `json:"timestamp"`     // 执行时间戳
	Metrics       *TestMetrics  `json:"metrics"`       // 测试指标（可选）
}

// TestMetrics 测试指标信息
type TestMetrics struct {
	TotalItems    int     `json:"totalItems"`    // 总项目数
	SuccessItems  int     `json:"successItems"`  // 成功项目数
	FailedItems   int     `json:"failedItems"`   // 失败项目数
	AccuracyRate  float64 `json:"accuracyRate"`  // 准确率百分比
	PerformanceMs float64 `json:"performanceMs"` // 性能指标（毫秒）
	ExtraInfo     map[string]interface{} `json:"extraInfo"` // 额外信息
}

// ValidationSuite 验证套件
type ValidationSuite struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Tests       []*ValidationResult  `json:"tests"`
	StartTime   time.Time            `json:"startTime"`
	EndTime     time.Time            `json:"endTime"`
	Duration    time.Duration        `json:"duration"`
	Summary     *ValidationSummary   `json:"summary"`
}

// ValidationSummary 验证摘要信息
type ValidationSummary struct {
	TotalTests      int               `json:"totalTests"`
	PassedTests     int               `json:"passedTests"`
	FailedTests     int               `json:"failedTests"`
	SkippedTests    int               `json:"skippedTests"`
	PassRate        float64           `json:"passRate"`
	TotalDuration   time.Duration     `json:"totalDuration"`
	StartTime       time.Time         `json:"startTime"`
	EndTime         time.Time         `json:"endTime"`
	CategoryStats   map[string]int    `json:"categoryStats"`   // 按类别统计
	ProjectInfo     *ProjectInfo      `json:"projectInfo"`     // 项目信息
}

// ProjectInfo 项目信息
type ProjectInfo struct {
	Path          string            `json:"path"`
	SourceFiles   int               `json:"sourceFiles"`
	TotalNodes    int               `json:"totalNodes"`
	TotalSymbols  int               `json:"totalSymbols"`
	APIVersions   map[string]string `json:"apiVersions"`
	FileExtensions []string         `json:"fileExtensions"`
	IgnorePatterns []string         `json:"ignorePatterns"`
}

// ValidationConfig 验证配置
type ValidationConfig struct {
	ProjectPath       string        `json:"projectPath"`
	IgnorePatterns    []string      `json:"ignorePatterns"`
	TargetExtensions []string       `json:"targetExtensions"`
	OutputDir         string        `json:"outputDir"`
	EnableJSON        bool          `json:"enableJSON"`
	EnableConsole     bool          `json:"enableConsole"`
	TestCategories    []string      `json:"testCategories"`
	Timeout           time.Duration `json:"timeout"`
	Verbose           bool          `json:"verbose"`
}

// TestResult 测试结果基础类型
type TestResult struct {
	Status   string `json:"status"`   // "passed", "failed", "skipped"
	Message  string `json:"message"`  // 结果消息
	Error    string `json:"error"`    // 错误信息
	Metadata map[string]interface{} `json:"metadata"` // 元数据
}

// NewValidationSuite 创建新的验证套件
func NewValidationSuite(name, description string) *ValidationSuite {
	return &ValidationSuite{
		Name:        name,
		Description: description,
		Tests:       make([]*ValidationResult, 0),
		StartTime:   time.Now(),
		Summary: &ValidationSummary{
			CategoryStats: make(map[string]int),
		},
	}
}

// AddTest 添加测试结果到验证套件
func (suite *ValidationSuite) AddTest(result *ValidationResult) {
	suite.Tests = append(suite.Tests, result)
	suite.Summary.CategoryStats[result.Category]++
}

// Finish 完成验证套件
func (suite *ValidationSuite) Finish() *ValidationSuite {
	suite.EndTime = time.Now()
	suite.Duration = suite.EndTime.Sub(suite.StartTime)

	// 计算摘要统计
	suite.Summary.TotalTests = len(suite.Tests)
	suite.Summary.StartTime = suite.StartTime
	suite.Summary.EndTime = suite.EndTime
	suite.Summary.TotalDuration = suite.Duration

	// 计算通过率
	for _, test := range suite.Tests {
		switch test.Status {
		case "passed":
			suite.Summary.PassedTests++
		case "failed":
			suite.Summary.FailedTests++
		case "skipped":
			suite.Summary.SkippedTests++
		}
	}

	if suite.Summary.TotalTests > 0 {
		suite.Summary.PassRate = float64(suite.Summary.PassedTests) / float64(suite.Summary.TotalTests) * 100
	}

	return suite
}

// CreateValidationResult 创建验证结果
func CreateValidationResult(name, category, description string) *ValidationResult {
	return &ValidationResult{
		Name:        name,
		Category:    category,
		Description: description,
		Status:      "skipped", // 默认为跳过
		Timestamp:   time.Now(),
	}
}

// PassResult 创建通过的验证结果
func PassResult(name, category, description string) *ValidationResult {
	result := CreateValidationResult(name, category, description)
	result.Status = "passed"
	result.Message = "测试通过"
	return result
}

// FailResult 创建失败的验证结果
func FailResult(name, category, description, message string) *ValidationResult {
	result := CreateValidationResult(name, category, description)
	result.Status = "failed"
	result.Message = message
	return result
}

// FailResultWithError 创建包含错误信息的失败验证结果
func FailResultWithError(name, category, description, message string, err error) *ValidationResult {
	result := FailResult(name, category, description, message)
	if err != nil {
		result.Error = err.Error()
	}
	return result
}

// SkipResult 创建跳过的验证结果
func SkipResult(name, category, description, reason string) *ValidationResult {
	result := CreateValidationResult(name, category, description)
	result.Status = "skipped"
	result.Message = reason
	return result
}

// WithMetrics 为验证结果添加指标
func (result *ValidationResult) WithMetrics(metrics *TestMetrics) *ValidationResult {
	result.Metrics = metrics
	return result
}

// WithDuration 为验证结果添加执行时间
func (result *ValidationResult) WithDuration(duration time.Duration) *ValidationResult {
	result.Duration = duration
	return result
}

// RunValidationWithMetrics 执行带指标的验证函数
func RunValidationWithMetrics(name, category, description string, validationFunc func() (*TestMetrics, error)) *ValidationResult {
	startTime := time.Now()
	result := CreateValidationResult(name, category, description)

	metrics, err := validationFunc()
	duration := time.Since(startTime)

	if err != nil {
		return result.WithDuration(duration).
			WithStatus("failed").
			WithError("验证函数执行失败", err)
	}

	return result.WithDuration(duration).
		WithStatus("passed").
		WithMetrics(metrics).
		WithMessage("验证通过")
}

// WithStatus 设置验证结果状态
func (result *ValidationResult) WithStatus(status string) *ValidationResult {
	result.Status = status
	return result
}

// WithMessage 设置验证结果消息
func (result *ValidationResult) WithMessage(message string) *ValidationResult {
	result.Message = message
	return result
}

// WithError 设置验证结果错误信息
func (result *ValidationResult) WithError(message string, err error) *ValidationResult {
	result.Message = message
	if err != nil {
		result.Error = err.Error()
	}
	return result
}

// CreateTestMetrics 创建测试指标
func CreateTestMetrics(total, success int) *TestMetrics {
	failed := total - success
	var accuracy float64
	if total > 0 {
		accuracy = float64(success) / float64(total) * 100
	}

	return &TestMetrics{
		TotalItems:   total,
		SuccessItems: success,
		FailedItems:  failed,
		AccuracyRate: accuracy,
		ExtraInfo:    make(map[string]interface{}),
	}
}

// WithPerformance 添加性能指标
func (metrics *TestMetrics) WithPerformance(performance float64) *TestMetrics {
	metrics.PerformanceMs = performance
	return metrics
}

// WithExtraInfo 添加额外信息
func (metrics *TestMetrics) WithExtraInfo(key string, value interface{}) *TestMetrics {
	if metrics.ExtraInfo == nil {
		metrics.ExtraInfo = make(map[string]interface{})
	}
	metrics.ExtraInfo[key] = value
	return metrics
}

// DefaultConfig 创建默认验证配置
func DefaultConfig(projectPath string) *ValidationConfig {
	return &ValidationConfig{
		ProjectPath:       projectPath,
		IgnorePatterns:    []string{"node_modules", "dist", "build", ".git"},
		TargetExtensions: []string{".ts", ".tsx"},
		OutputDir:         "../../validation-results",
		EnableJSON:        true,
		EnableConsole:     true,
		TestCategories:    []string{"project-api", "node-api", "symbol-api", "type-api", "lsp-api", "accuracy-validation"},
		Timeout:           30 * time.Second,
		Verbose:           true,
	}
}

// LoadTestCases 从JSON文件加载测试用例
func LoadTestCases(filePath string, testCaseType interface{}) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取测试用例文件失败: %w", err)
	}

	if err := json.Unmarshal(data, testCaseType); err != nil {
		return fmt.Errorf("解析测试用例JSON失败: %w", err)
	}

	return nil
}

// SaveTestResults 保存测试结果到JSON文件
func SaveTestResults(results interface{}, outputPath string) error {
	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 序列化结果
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化测试结果失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("写入测试结果文件失败: %w", err)
	}

	return nil
}

// RunSafe 安全执行函数并捕获错误
func RunSafe(name string, fn func() error) (success bool, duration time.Duration, err error) {
	start := time.Now()
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("执行函数 %s 时发生panic: %v", name, r)
			success = false
		}
		duration = time.Since(start)
	}()

	err = fn()
	success = err == nil
	return success, duration, err
}