package project_analyzer

import (
	"fmt"
)

// AnalyzerWithConfig 带配置的分析器包装
type AnalyzerWithConfig struct {
	Analyzer
	Config map[string]string
}

// Config 配置执行参数
type Config struct {
	// Analyzers 要执行的分析器列表及其配置
	Analyzers []*AnalyzerWithConfig
}

// ExecuteWithConfig 使用配置对象执行分析
// 这是推荐的 Go 项目调用方式，避免了注册和配置的分离
func (r *Runner) ExecuteWithConfig(config *Config) (map[string]Result, error) {
	if len(config.Analyzers) == 0 {
		return nil, fmt.Errorf("no analyzers specified")
	}

	// 1. 注册所有分析器
	for _, awc := range config.Analyzers {
		if err := r.Register(awc.Analyzer); err != nil {
			return nil, err
		}
	}

	// 2. 构建配置映射
	configs := make(map[string]map[string]string)
	for _, awc := range config.Analyzers {
		name := awc.Analyzer.Name()
		if awc.Config != nil {
			configs[name] = awc.Config
		} else {
			configs[name] = make(map[string]string)
		}
	}

	// 3. 执行分析
	return r.RunBatch(configs)
}

// AddAnalyzer 添加分析器（链式调用）
func (c *Config) AddAnalyzer(analyzer Analyzer, config map[string]string) *Config {
	c.Analyzers = append(c.Analyzers, &AnalyzerWithConfig{
		Analyzer: analyzer,
		Config:   config,
	})
	return c
}

// NewConfig 创建配置对象
func NewConfig() *Config {
	return &Config{
		Analyzers: make([]*AnalyzerWithConfig, 0),
	}
}
