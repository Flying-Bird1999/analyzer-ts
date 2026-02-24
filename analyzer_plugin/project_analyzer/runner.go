// Package project_analyzer 提供了 Runner，用于 Go 项目直接调用分析器
package project_analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
)

// Runner 分析器执行器，支持 Go 项目直接调用
// 封装了项目解析和分析器执行的完整流程
type Runner struct {
	// 项目根目录
	ProjectRoot string
	// 排除规则
	Exclude []string
	// 是否为 monorepo
	IsMonorepo bool
	// 已注册的分析器
	analyzers map[string]Analyzer
	mu        sync.RWMutex
}

// RunnerConfig Runner 配置
type RunnerConfig struct {
	// 项目根目录（必需）
	ProjectRoot string
	// 排除规则（可选）
	Exclude []string
	// 是否为 monorepo（可选，默认 false）
	IsMonorepo bool
}

// NewRunner 创建分析器执行器
func NewRunner(config RunnerConfig) (*Runner, error) {
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

	return &Runner{
		ProjectRoot: absPath,
		Exclude:     config.Exclude,
		IsMonorepo:  config.IsMonorepo,
		analyzers:   make(map[string]Analyzer),
	}, nil
}

// Register 注册分析器
func (r *Runner) Register(analyzer Analyzer) error {
	if analyzer == nil {
		return fmt.Errorf("analyzer cannot be nil")
	}

	name := analyzer.Name()
	if name == "" {
		return fmt.Errorf("analyzer name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.analyzers[name] = analyzer
	return nil
}

// RegisterBatch 批量注册分析器
func (r *Runner) RegisterBatch(analyzers ...Analyzer) error {
	for _, analyzer := range analyzers {
		if err := r.Register(analyzer); err != nil {
			return err
		}
	}
	return nil
}

// Get 获取已注册的分析器
func (r *Runner) Get(name string) (Analyzer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	analyzer, ok := r.analyzers[name]
	return analyzer, ok
}

// List 列出所有已注册的分析器名称
func (r *Runner) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.analyzers))
	for name := range r.analyzers {
		names = append(names, name)
	}
	return names
}

// ParseResult 项目解析结果缓存
// 避免多次执行时重复解析
type ParseResult struct {
	Result *projectParser.ProjectParserResult
	Error  error
}

// parseInternal 内部解析方法
func (r *Runner) parseInternal() *ParseResult {
	config := projectParser.NewProjectParserConfig(r.ProjectRoot, r.Exclude, r.IsMonorepo, nil)
	result := projectParser.NewProjectParserResult(config)
	result.ProjectParser()

	return &ParseResult{
		Result: result,
		Error:  nil,
	}
}

// Parse 解析项目（可复用结果）
func (r *Runner) Parse() (*projectParser.ProjectParserResult, error) {
	pr := r.parseInternal()
	// 注意：ProjectParser() 方法不会返回错误，错误会通过其他方式处理
	return pr.Result, nil
}

// RunSingle 运行单个分析器
// analyzerName: 分析器名称
// config: 分析器配置，通过 Configure(map[string]string) 传递
//
// 返回分析结果的 struct，调用方可以直接使用，也可以调用 ToJSON() 序列化
func (r *Runner) RunSingle(analyzerName string, config map[string]string) (Result, error) {
	// 1. 获取分析器
	analyzer, ok := r.Get(analyzerName)
	if !ok {
		return nil, fmt.Errorf("analyzer '%s' not found", analyzerName)
	}

	// 2. 配置分析器
	if config != nil {
		if err := analyzer.Configure(config); err != nil {
			return nil, fmt.Errorf("configure analyzer '%s' failed: %w", analyzerName, err)
		}
	}

	// 3. 解析项目
	pr := r.parseInternal()

	// 4. 创建项目上下文
	ctx := &ProjectContext{
		ProjectRoot:   r.ProjectRoot,
		Exclude:       r.Exclude,
		IsMonorepo:    r.IsMonorepo,
		ParsingResult: pr.Result,
	}

	// 5. 执行分析
	result, err := analyzer.Analyze(ctx)
	if err != nil {
		return nil, fmt.Errorf("analyze failed: %w", err)
	}

	return result, nil
}

// RunBatch 批量运行多个分析器
// configs: 分析器名称到配置的映射
//
// 返回分析器名称到结果的映射
func (r *Runner) RunBatch(configs map[string]map[string]string) (map[string]Result, error) {
	if len(configs) == 0 {
		return nil, fmt.Errorf("no analyzers specified")
	}

	// 1. 解析项目（只解析一次）
	pr := r.parseInternal()

	// 2. 创建项目上下文（所有分析器共享）
	ctx := &ProjectContext{
		ProjectRoot:   r.ProjectRoot,
		Exclude:       r.Exclude,
		IsMonorepo:    r.IsMonorepo,
		ParsingResult: pr.Result,
	}

	// 3. 并发执行所有分析器
	results := make(map[string]Result)
	var mu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, len(configs))

	for analyzerName, config := range configs {
		// 获取分析器
		analyzer, ok := r.Get(analyzerName)
		if !ok {
			return nil, fmt.Errorf("analyzer '%s' not found", analyzerName)
		}

		wg.Add(1)
		go func(name string, a Analyzer, cfg map[string]string) {
			defer wg.Done()

			// 配置分析器
			if cfg != nil {
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
