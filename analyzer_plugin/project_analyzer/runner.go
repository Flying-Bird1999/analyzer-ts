// Package project_analyzer 提供了项目分析器，用于 Go 项目直接调用分析器
package project_analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
)

// =============================================================================
// 分析器类型枚举
// =============================================================================
// 注意：这些常量的值必须与各 Analyzer.Name() 的返回值保持一致
// 这是 Go 中提供类型提示的唯一方式，虽然需要一定程度的重复维护

// AnalyzerType 分析器类型枚举
type AnalyzerType string

const (
	AnalyzerPkgDeps       AnalyzerType = "pkg-deps"
	AnalyzerComponentDeps AnalyzerType = "component-deps"
	AnalyzerExportCall    AnalyzerType = "export-call"
	AnalyzerUnconsumed    AnalyzerType = "unconsumed"
	AnalyzerCountAny      AnalyzerType = "count-any"
	AnalyzerCountAs       AnalyzerType = "count-as"
	AnalyzerNpmCheck      AnalyzerType = "npm-check"
	AnalyzerUnreferenced  AnalyzerType = "find-unreferenced-files"
	AnalyzerTrace         AnalyzerType = "trace"
	AnalyzerApiTracer     AnalyzerType = "api-tracer"
	AnalyzerCssFile       AnalyzerType = "css-file"
	AnalyzerMdFile        AnalyzerType = "md-file"
)

// analyzerRegistry 分析器注册表（名称 -> 工厂函数）
var analyzerRegistry = struct {
	sync.RWMutex
	factories map[string]func() Analyzer
}{
	factories: make(map[string]func() Analyzer),
}

// RegisterAnalyzer 注册分析器工厂函数
// 各个 analyzer 包在 init() 时调用此方法注册自己
func RegisterAnalyzer(name string, factory func() Analyzer) {
	analyzerRegistry.Lock()
	defer analyzerRegistry.Unlock()
	analyzerRegistry.factories[name] = factory
}

// String 实现 Stringer 接口
func (t AnalyzerType) String() string {
	return string(t)
}

// =============================================================================
// 配置类型定义
// =============================================================================

// PkgDepsConfig pkg-deps 分析器配置（无需配置）
type PkgDepsConfig struct{}

// ComponentDepsConfig component-deps 分析器配置
type ComponentDepsConfig struct {
	Manifest string
}

// ExportCallConfig export-call 分析器配置
type ExportCallConfig struct {
	Manifest string
}

// UnconsumedConfig unconsumed 分析器配置（无需配置）
type UnconsumedConfig struct{}

// CountAnyConfig count-any 分析器配置（无需配置）
type CountAnyConfig struct{}

// CountAsConfig count-as 分析器配置（无需配置）
type CountAsConfig struct{}

// NpmCheckConfig npm-check 分析器配置（无需配置）
type NpmCheckConfig struct{}

// UnreferencedConfig unreferenced 分析器配置
type UnreferencedConfig struct {
	Entrypoint string
}

// TraceConfig trace 分析器配置
type TraceConfig struct {
	TargetPkgs string
}

// ApiTracerConfig api-tracer 分析器配置
type ApiTracerConfig struct {
	ApiPaths string
}

// CssFileConfig css-file 分析器配置（无需配置）
type CssFileConfig struct{}

// MdFileConfig md-file 分析器配置（无需配置）
type MdFileConfig struct{}

// =============================================================================
// 配置接口
// =============================================================================

// AnalyzerConfig 分析器配置接口
type AnalyzerConfig interface {
	// ToMap 将配置转换为 map[string]string
	ToMap() map[string]string
}

// 实现 ToMap 方法
func (c PkgDepsConfig) ToMap() map[string]string { return nil }
func (c ComponentDepsConfig) ToMap() map[string]string {
	if c.Manifest == "" {
		panic("ComponentDepsConfig.Manifest is required")
	}
	return map[string]string{"manifest": c.Manifest}
}
func (c ExportCallConfig) ToMap() map[string]string {
	if c.Manifest == "" {
		panic("ExportCallConfig.Manifest is required")
	}
	return map[string]string{"manifest": c.Manifest}
}
func (c UnconsumedConfig) ToMap() map[string]string { return nil }
func (c CountAnyConfig) ToMap() map[string]string   { return nil }
func (c CountAsConfig) ToMap() map[string]string    { return nil }
func (c NpmCheckConfig) ToMap() map[string]string   { return nil }
func (c UnreferencedConfig) ToMap() map[string]string {
	m := make(map[string]string)
	if c.Entrypoint != "" {
		m["entrypoint"] = c.Entrypoint
	}
	return m
}
func (c TraceConfig) ToMap() map[string]string {
	if c.TargetPkgs == "" {
		panic("TraceConfig.TargetPkgs is required")
	}
	return map[string]string{"targetPkgs": c.TargetPkgs}
}
func (c ApiTracerConfig) ToMap() map[string]string {
	if c.ApiPaths == "" {
		panic("ApiTracerConfig.ApiPaths is required")
	}
	return map[string]string{"apiPaths": c.ApiPaths}
}
func (c CssFileConfig) ToMap() map[string]string { return nil }
func (c MdFileConfig) ToMap() map[string]string  { return nil }

// toConfigMap 将任意配置类型转换为 map[string]string
func toConfigMap(config any) map[string]string {
	if config == nil {
		return nil
	}

	// 如果实现了 AnalyzerConfig 接口
	if ac, ok := config.(AnalyzerConfig); ok {
		return ac.ToMap()
	}

	// 否则使用反射
	v := reflect.ValueOf(config)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	result := make(map[string]string)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// 跳过零值
		if fieldValue.IsZero() {
			continue
		}

		// 使用 json tag 或字段名
		name := strings.ToLower(field.Name)
		if tag := field.Tag.Get("json"); tag != "" {
			if idx := strings.Index(tag, ","); idx != -1 {
				name = tag[:idx]
			} else {
				name = tag
			}
		}

		// 转换值
		switch fieldValue.Kind() {
		case reflect.String:
			result[name] = fieldValue.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			result[name] = fmt.Sprintf("%d", fieldValue.Int())
		case reflect.Bool:
			result[name] = fmt.Sprintf("%t", fieldValue.Bool())
		case reflect.Slice:
			if fieldValue.Type().Elem().Kind() == reflect.String {
				slice := fieldValue.Interface().([]string)
				result[name] = strings.Join(slice, ",")
			}
		}
	}

	return result
}

// =============================================================================
// 项目配置类型
// =============================================================================

// Config 项目分析器配置
type Config struct {
	// 项目根目录（必需）
	ProjectRoot string
	// 排除规则（可选）
	Exclude []string
	// 是否为 monorepo（可选，默认 false）
	IsMonorepo bool
}

// AnalyzerWithConfig 带配置的分析器包装（内部使用）
type AnalyzerWithConfig struct {
	Analyzer Analyzer
	Config   map[string]string
}

// ExecutionConfig 执行配置，包含要运行的分析器及其配置
type ExecutionConfig struct {
	// Analyzers 要执行的分析器列表及其配置
	Analyzers []*AnalyzerWithConfig
}

// =============================================================================
// 分析器核心类型
// =============================================================================

// ProjectAnalyzer 项目分析器，支持 Go 项目直接调用
// 封装了项目解析和分析器执行的完整流程
type ProjectAnalyzer struct {
	// 项目根目录
	ProjectRoot string
	// 排除规则
	Exclude []string
	// 是否为 monorepo
	IsMonorepo bool
	// 已注册的分析器
	analyzers map[string]Analyzer
	mu        sync.RWMutex
	// 分析上下文（NewProjectAnalyzer 时自动创建）
	context *ProjectContext
}

// =============================================================================
// 构造函数
// =============================================================================

// NewProjectAnalyzer 创建项目分析器并自动解析项目
// 这是项目分析的入口点，会立即执行项目解析（耗时操作）
func NewProjectAnalyzer(config Config) (*ProjectAnalyzer, error) {
	if config.ProjectRoot == "" {
		return nil, fmt.Errorf("ProjectRoot is required")
	}

	// 转换为绝对路径
	absPath, err := filepath.Abs(config.ProjectRoot)
	if err != nil {
		return nil, fmt.Errorf("invalid project root: %w", err)
	}

	// 检查目录是否存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("project root does not exist: %s", absPath)
	}

	analyzer := &ProjectAnalyzer{
		ProjectRoot: absPath,
		Exclude:     config.Exclude,
		IsMonorepo:  config.IsMonorepo,
		analyzers:   make(map[string]Analyzer),
	}

	// 立即解析项目（耗时操作）
	pr := analyzer.parseInternal()

	// 创建并缓存分析上下文
	analyzer.context = &ProjectContext{
		ProjectRoot:   analyzer.ProjectRoot,
		Exclude:       analyzer.Exclude,
		IsMonorepo:    analyzer.IsMonorepo,
		ParsingResult: pr.Result,
	}

	return analyzer, nil
}

// =============================================================================
// 执行方法
// =============================================================================

// ExecuteWithConfig 使用配置对象执行分析
// 这是推荐的 Go 项目调用方式，一步完成注册、配置和执行
func (p *ProjectAnalyzer) ExecuteWithConfig(config *ExecutionConfig) (map[string]Result, error) {
	// 验证配置
	if config == nil {
		return nil, fmt.Errorf("execution config cannot be nil")
	}
	if len(config.Analyzers) == 0 {
		return nil, fmt.Errorf("no analyzers specified")
	}

	// 注册并验证分析器
	seenNames := make(map[string]bool)
	for _, awc := range config.Analyzers {
		if awc.Analyzer == nil {
			return nil, fmt.Errorf("analyzer cannot be nil")
		}

		name := awc.Analyzer.Name()
		if name == "" {
			return nil, fmt.Errorf("analyzer name cannot be empty")
		}

		// 检测重复
		if seenNames[name] {
			return nil, fmt.Errorf("duplicate analyzer name: %s", name)
		}
		seenNames[name] = true

		// 注册
		if err := p.registerAnalyzer(awc.Analyzer); err != nil {
			return nil, err
		}
	}

	// 构建配置映射
	configs := make(map[string]map[string]string)
	for _, awc := range config.Analyzers {
		name := awc.Analyzer.Name()
		if awc.Config != nil {
			configs[name] = awc.Config
		} else {
			configs[name] = make(map[string]string)
		}
	}

	// 执行分析
	return p.runBatch(configs)
}

// =============================================================================
// 内部方法
// =============================================================================

// registerAnalyzer 注册单个分析器（内部方法）
func (p *ProjectAnalyzer) registerAnalyzer(analyzer Analyzer) error {
	name := analyzer.Name()
	if name == "" {
		return fmt.Errorf("analyzer name cannot be empty")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.analyzers[name] = analyzer
	return nil
}

// getAnalyzer 获取已注册的分析器（内部方法）
func (p *ProjectAnalyzer) getAnalyzer(name string) (Analyzer, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	analyzer, ok := p.analyzers[name]
	return analyzer, ok
}

// runBatch 批量运行多个分析器（内部方法）
func (p *ProjectAnalyzer) runBatch(configs map[string]map[string]string) (map[string]Result, error) {
	if len(configs) == 0 {
		return nil, fmt.Errorf("no analyzers specified")
	}

	// 使用已缓存的上下文（在 NewProjectAnalyzer 中已解析）
	ctx := p.context

	// 并发执行所有分析器
	results := make(map[string]Result)
	var mu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, len(configs))

	for analyzerName, config := range configs {
		// 获取分析器
		analyzer, ok := p.getAnalyzer(analyzerName)
		if !ok {
			return nil, fmt.Errorf("analyzer '%s' not found", analyzerName)
		}

		wg.Add(1)
		go func(name string, a Analyzer, cfg map[string]string) {
			defer wg.Done()

			// 配置分析器
			if cfg != nil && len(cfg) > 0 {
				if err := a.Configure(cfg); err != nil {
					errChan <- fmt.Errorf("configure analyzer '%s' failed: %w", name, err)
					return
				}
			}

			// 执行分析
			result, err := a.Analyze(ctx)
			if err != nil {
				errChan <- fmt.Errorf("analyze '%s' failed: %w", name, err)
				return
			}

			mu.Lock()
			results[name] = result
			mu.Unlock()
		}(analyzerName, analyzer, config)
	}

	wg.Wait()
	close(errChan)

	// 检查是否有错误
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return results, fmt.Errorf("analysis completed with %d errors: %v", len(errs), errs)
	}

	return results, nil
}

// =============================================================================
// 解析方法
// =============================================================================

// ParseResult 项目解析结果缓存
// 避免多次执行时重复解析
type ParseResult struct {
	Result *projectParser.ProjectParserResult
	Error  error
}

// parseInternal 内部解析方法
func (p *ProjectAnalyzer) parseInternal() *ParseResult {
	config := projectParser.NewProjectParserConfig(p.ProjectRoot, p.Exclude, p.IsMonorepo, nil)
	result := projectParser.NewProjectParserResult(config)
	result.ProjectParser()

	return &ParseResult{
		Result: result,
		Error:  nil,
	}
}

// Context 返回分析上下文
// 用于在业务代码中传递和复用解析结果
// NewProjectAnalyzer 构造时会自动解析并创建 context
func (p *ProjectAnalyzer) Context() *ProjectContext {
	return p.context
}

// =============================================================================
// 单独执行 analyzer 的方法（支持按需调用）
// =============================================================================

// runOne 内部方法：运行单个分析器
func (p *ProjectAnalyzer) runOne(analyzerType AnalyzerType, config any) (Result, error) {
	// 检查是否已解析
	if p.context == nil {
		return nil, fmt.Errorf("project not parsed, please call Parse() first")
	}

	// 使用缓存的上下文
	ctx := p.context
	name := string(analyzerType)

	// 从注册表获取 analyzer
	analyzerRegistry.RLock()
	factory, ok := analyzerRegistry.factories[name]
	analyzerRegistry.RUnlock()

	if !ok {
		return nil, fmt.Errorf("analyzer '%s' not registered", name)
	}

	analyzer := factory()

	// 转换配置
	var configMap map[string]string
	if config != nil {
		configMap = toConfigMap(config)
	}

	// 配置 analyzer
	if configMap != nil && len(configMap) > 0 {
		if err := analyzer.Configure(configMap); err != nil {
			return nil, fmt.Errorf("configure analyzer '%s' failed: %w", name, err)
		}
	}

	// 执行分析
	return analyzer.Analyze(ctx)
}

// =============================================================================
// 辅助函数
// =============================================================================

// NewConfig 创建配置对象
func NewConfig() *Config {
	return &Config{
		Exclude: []string{"node_modules/**", "dist/**", "**/*.test.ts", "**/*.spec.ts"},
	}
}

// NewExecutionConfig 创建执行配置对象
func NewExecutionConfig() *ExecutionConfig {
	return &ExecutionConfig{
		Analyzers: make([]*AnalyzerWithConfig, 0),
	}
}

// AddAnalyzer 添加分析器到执行配置
//
// 参数说明:
//   - analyzerType: 使用 AnalyzerType 常量（IDE 会自动补全）
//   - config: 分析器配置，使用对应的配置结构体
//
// 可用的 AnalyzerType 常量（IDE 会自动提示）:
//   - AnalyzerPkgDeps
//   - AnalyzerComponentDeps
//   - AnalyzerExportCall
//   - AnalyzerUnconsumed
//   - AnalyzerCountAny
//   - AnalyzerCountAs
//   - AnalyzerNpmCheck
//   - AnalyzerUnreferenced
//   - AnalyzerTrace
//   - AnalyzerApiTracer
//   - AnalyzerCssFile
//   - AnalyzerMdFile
//
// 使用示例:
//
//	import _ "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/pkg_deps"
//	// 注意：需要 import 各个 analyzer 包以触发其 init() 注册
//
//	execConfig := project_analyzer.NewExecutionConfig().
//	    AddAnalyzer(project_analyzer.AnalyzerPkgDeps, project_analyzer.PkgDepsConfig{}).
//	    AddAnalyzer(project_analyzer.AnalyzerComponentDeps, project_analyzer.ComponentDepsConfig{
//	        Manifest: path,
//	    })
func (c *ExecutionConfig) AddAnalyzer(analyzerType AnalyzerType, config any) *ExecutionConfig {
	// 验证名称是否有效
	if !isValidAnalyzerName(string(analyzerType)) {
		panic(fmt.Sprintf("unknown analyzer: %q\n\nAvailable analyzers:\n  - %s",
			analyzerType, strings.Join(getRegisteredAnalyzerNames(), "\n  - ")))
	}

	name := string(analyzerType)

	analyzerRegistry.RLock()
	factory, ok := analyzerRegistry.factories[name]
	analyzerRegistry.RUnlock()

	if !ok {
		panic(fmt.Sprintf("analyzer %q not registered. Did you import its package?\n\nRegistered analyzers:\n  - %s",
			name, strings.Join(getRegisteredAnalyzerNames(), "\n  - ")))
	}

	analyzer := factory()

	// 转换配置
	var configMap map[string]string
	if config != nil {
		configMap = toConfigMap(config)
	}

	c.Analyzers = append(c.Analyzers, &AnalyzerWithConfig{
		Analyzer: analyzer,
		Config:   configMap,
	})

	return c
}

// isValidAnalyzerName 检查名称是否在已注册的分析器列表中
func isValidAnalyzerName(name string) bool {
	analyzerRegistry.RLock()
	defer analyzerRegistry.RUnlock()
	_, ok := analyzerRegistry.factories[name]
	return ok
}

// getRegisteredAnalyzerNames 获取所有已注册的分析器名称
func getRegisteredAnalyzerNames() []string {
	analyzerRegistry.RLock()
	defer analyzerRegistry.RUnlock()

	names := make([]string, 0, len(analyzerRegistry.factories))
	for name := range analyzerRegistry.factories {
		names = append(names, name)
	}
	return names
}

// AvailableAnalyzers 返回所有可用的分析器列表（用于文档等）
func AvailableAnalyzers() []AnalyzerType {
	analyzerRegistry.RLock()
	defer analyzerRegistry.RUnlock()

	names := make([]AnalyzerType, 0, len(analyzerRegistry.factories))
	for name := range analyzerRegistry.factories {
		names = append(names, AnalyzerType(name))
	}
	return names
}

// GetAvailableAnalyzersMap 返回所有已注册的分析器映射
// key 是从 Analyzer.Name() 方法动态获取的，确保单一数据源
// 这样 CLI 的 availableAnalyzers 不需要硬编码 key，而是直接使用 analyzer 自己报告的名称
func GetAvailableAnalyzersMap() map[string]Analyzer {
	analyzerRegistry.RLock()
	defer analyzerRegistry.RUnlock()

	result := make(map[string]Analyzer)
	for _, factory := range analyzerRegistry.factories {
		analyzer := factory()
		name := analyzer.Name()
		result[name] = analyzer
	}
	return result
}

// GetResult 泛型方法获取强类型结果（无需传入名称）
// 通过遍历 results 找到类型匹配的结果
//
// 使用示例:
//
//	result, err := project_analyzer.GetResult[*export_call.ExportCallResult](results)
func GetResult[T Result](results map[string]Result) (T, error) {
	var zero T

	// 遍历所有结果，找到类型匹配的
	for _, result := range results {
		if typed, ok := result.(T); ok {
			return typed, nil
		}
	}

	return zero, fmt.Errorf("result not found for type %T", zero)
}

// RunOneT 类型安全的泛型函数，直接返回具体类型
// 无需手动进行类型断言，编译器会自动推导返回类型
//
// 注意：由于 Go 泛型的限制，这是包级别的函数而不是方法
//
// 使用示例:
//
//	listResult, err := project_analyzer.RunOneT[*pkg_deps.PkgDepsResult](
//	    s.analyzer,
//	    project_analyzer.AnalyzerPkgDeps,
//	    project_analyzer.PkgDepsConfig{},
//	)
//	// listResult 直接是 *pkg_deps.PkgDepsResult 类型，无需类型断言
func RunOneT[T Result](analyzer *ProjectAnalyzer, analyzerType AnalyzerType, config any) (T, error) {
	var zero T

	// 调用内部 runOne 方法
	result, err := analyzer.runOne(analyzerType, config)
	if err != nil {
		return zero, err
	}

	// 类型断言
	typed, ok := result.(T)
	if !ok {
		return zero, fmt.Errorf("analyzer '%s' returned unexpected type %T, expected %T", analyzerType, result, zero)
	}

	return typed, nil
}
