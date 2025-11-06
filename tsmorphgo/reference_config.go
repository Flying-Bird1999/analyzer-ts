package tsmorphgo

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ReferencesConfig 引用查找功能的配置
type ReferencesConfig struct {
	// CacheSettings 缓存设置
	CacheSettings CacheSettings `json:"cacheSettings"`

	// RetrySettings 重试设置
	RetrySettings RetrySettings `json:"retrySettings"`

	// PerformanceSettings 性能设置
	PerformanceSettings PerformanceSettings `json:"performanceSettings"`

	// FallbackSettings 降级设置
	FallbackSettings FallbackSettings `json:"fallbackSettings"`

	// LoggingSettings 日志设置
	LoggingSettings LoggingSettings `json:"loggingSettings"`
}

// CacheSettings 缓存配置
type CacheSettings struct {
	// Enabled 是否启用缓存
	Enabled bool `json:"enabled"`

	// MaxEntries 最大缓存条目数
	MaxEntries int `json:"maxEntries"`

	// TTL 缓存生存时间
	TTL string `json:"ttl"`

	// CleanupInterval 缓存清理间隔
	CleanupInterval string `json:"cleanupInterval"`

	// EnableFileHashCheck 是否启用文件内容哈希检查
	EnableFileHashCheck bool `json:"enableFileHashCheck"`
}

// RetrySettings 重试配置
type RetrySettings struct {
	// Enabled 是否启用重试
	Enabled bool `json:"enabled"`

	// MaxRetries 最大重试次数
	MaxRetries int `json:"maxRetries"`

	// BaseDelay 基础延迟
	BaseDelay string `json:"baseDelay"`

	// MaxDelay 最大延迟
	MaxDelay string `json:"maxDelay"`

	// BackoffFactor 退避因子
	BackoffFactor float64 `json:"backoffFactor"`

	// RetryableErrors 可重试的错误类型
	RetryableErrors []string `json:"retryableErrors"`
}

// PerformanceSettings 性能配置
type PerformanceSettings struct {
	// EnableMetrics 是否启用性能指标收集
	EnableMetrics bool `json:"enableMetrics"`

	// MetricsInterval 指标收集间隔
	MetricsInterval string `json:"metricsInterval"`

	// EnableBatching 是否启用批量处理
	EnableBatching bool `json:"enableBatching"`

	// BatchSize 批量处理大小
	BatchSize int `json:"batchSize"`

	// Timeout 操作超时时间
	Timeout string `json:"timeout"`
}

// FallbackSettings 降级配置
type FallbackSettings struct {
	// Enabled 是否启用降级策略
	Enabled bool `json:"enabled"`

	// EnableContextAnalysis 是否启用上下文分析
	EnableContextAnalysis bool `json:"enableContextAnalysis"`

	// FallbackTimeout 降级策略超时时间
	FallbackTimeout string `json:"fallbackTimeout"`

	// MaxFallbackResults 最大降级结果数
	MaxFallbackResults int `json:"maxFallbackResults"`
}

// LoggingSettings 日志配置
type LoggingSettings struct {
	// Enabled 是否启用日志
	Enabled bool `json:"enabled"`

	// Level 日志级别
	Level string `json:"level"`

	// IncludeTimestamp 是否包含时间戳
	IncludeTimestamp bool `json:"includeTimestamp"`

	// LogCacheOperations 是否记录缓存操作
	LogCacheOperations bool `json:"logCacheOperations"`

	// LogRetryAttempts 是否记录重试尝试
	LogRetryAttempts bool `json:"logRetryAttempts"`

	// LogFallbackUsage 是否记录降级策略使用
	LogFallbackUsage bool `json:"logFallbackUsage"`
}

// DefaultReferencesConfig 返回默认配置
func DefaultReferencesConfig() *ReferencesConfig {
	return &ReferencesConfig{
		CacheSettings: CacheSettings{
			Enabled:             true,
			MaxEntries:          1000,
			TTL:                 "10m",
			CleanupInterval:     "5m",
			EnableFileHashCheck: true,
		},
		RetrySettings: RetrySettings{
			Enabled:     true,
			MaxRetries:  3,
			BaseDelay:   "100ms",
			MaxDelay:    "5s",
			BackoffFactor: 2.0,
			RetryableErrors: []string{
				"timeout",
				"unavailable",
				"service",
				"connection",
			},
		},
		PerformanceSettings: PerformanceSettings{
			EnableMetrics:   true,
			MetricsInterval: "1m",
			EnableBatching:  true,
			BatchSize:       50,
			Timeout:         "30s",
		},
		FallbackSettings: FallbackSettings{
			Enabled:               true,
			EnableContextAnalysis: true,
			FallbackTimeout:       "2s",
			MaxFallbackResults:    100,
		},
		LoggingSettings: LoggingSettings{
			Enabled:            true,
			Level:              "info",
			IncludeTimestamp:   true,
			LogCacheOperations: true,
			LogRetryAttempts:   true,
			LogFallbackUsage:   true,
		},
	}
}

// LoadReferencesConfig 从文件加载配置
func LoadReferencesConfig(configPath string) (*ReferencesConfig, error) {
	// 如果配置文件不存在，返回默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultReferencesConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	config := DefaultReferencesConfig()
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return config, nil
}

// SaveReferencesConfig 保存配置到文件
func (config *ReferencesConfig) SaveReferencesConfig(configPath string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// ToRetryConfig 将重试设置转换为 RetryConfig
func (rs *RetrySettings) ToRetryConfig() *RetryConfig {
	if !rs.Enabled {
		return &RetryConfig{Enabled: false}
	}

	baseDelay, _ := time.ParseDuration(rs.BaseDelay)
	maxDelay, _ := time.ParseDuration(rs.MaxDelay)

	return &RetryConfig{
		MaxRetries:    rs.MaxRetries,
		BaseDelay:     baseDelay,
		MaxDelay:      maxDelay,
		BackoffFactor: rs.BackoffFactor,
		Enabled:       true,
	}
}

// ToCacheTTL 将TTL字符串转换为 time.Duration
func (cs *CacheSettings) ToCacheTTL() time.Duration {
	if !cs.Enabled {
		return 0
	}

	ttl, _ := time.ParseDuration(cs.TTL)
	return ttl
}

// ToCacheMaxEntries 获取最大缓存条目数
func (cs *CacheSettings) ToCacheMaxEntries() int {
	if !cs.Enabled {
		return 0
	}

	return cs.MaxEntries
}

// ToPerformanceTimeout 将超时字符串转换为 time.Duration
func (ps *PerformanceSettings) ToPerformanceTimeout() time.Duration {
	timeout, _ := time.ParseDuration(ps.Timeout)
	return timeout
}

// ToFallbackTimeout 将降级超时字符串转换为 time.Duration
func (fs *FallbackSettings) ToFallbackTimeout() time.Duration {
	timeout, _ := time.ParseDuration(fs.FallbackTimeout)
	return timeout
}

// Validate 验证配置的有效性
func (config *ReferencesConfig) Validate() error {
	// 验证缓存设置
	if config.CacheSettings.Enabled {
		if config.CacheSettings.MaxEntries <= 0 {
			return fmt.Errorf("缓存最大条目数必须大于0")
		}

		if _, err := time.ParseDuration(config.CacheSettings.TTL); err != nil {
			return fmt.Errorf("无效的缓存TTL格式: %w", err)
		}

		if _, err := time.ParseDuration(config.CacheSettings.CleanupInterval); err != nil {
			return fmt.Errorf("无效的缓存清理间隔格式: %w", err)
		}
	}

	// 验证重试设置
	if config.RetrySettings.Enabled {
		if config.RetrySettings.MaxRetries < 0 {
			return fmt.Errorf("最大重试次数不能为负数")
		}

		if config.RetrySettings.BackoffFactor <= 1.0 {
			return fmt.Errorf("退避因子必须大于1.0")
		}

		if _, err := time.ParseDuration(config.RetrySettings.BaseDelay); err != nil {
			return fmt.Errorf("无效的基础延迟格式: %w", err)
		}

		if _, err := time.ParseDuration(config.RetrySettings.MaxDelay); err != nil {
			return fmt.Errorf("无效的最大延迟格式: %w", err)
		}
	}

	// 验证性能设置
	if _, err := time.ParseDuration(config.PerformanceSettings.Timeout); err != nil {
		return fmt.Errorf("无效的超时时间格式: %w", err)
	}

	if config.PerformanceSettings.BatchSize <= 0 {
		return fmt.Errorf("批量处理大小必须大于0")
	}

	// 验证降级设置
	if _, err := time.ParseDuration(config.FallbackSettings.FallbackTimeout); err != nil {
		return fmt.Errorf("无效的降级超时时间格式: %w", err)
	}

	if config.FallbackSettings.MaxFallbackResults <= 0 {
		return fmt.Errorf("最大降级结果数必须大于0")
	}

	// 验证日志设置
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLogLevels[config.LoggingSettings.Level] {
		return fmt.Errorf("无效的日志级别: %s", config.LoggingSettings.Level)
	}

	return nil
}

// Clone 克隆配置
func (config *ReferencesConfig) Clone() *ReferencesConfig {
	clone := *config
	return &clone
}

// String 返回配置的字符串表示
func (config *ReferencesConfig) String() string {
	data, _ := json.MarshalIndent(config, "", "  ")
	return string(data)
}

// ApplyDefaults 应用默认值到未设置的字段
func (config *ReferencesConfig) ApplyDefaults() {
	defaultConfig := DefaultReferencesConfig()

	// 应用缓存设置的默认值
	if config.CacheSettings.MaxEntries == 0 {
		config.CacheSettings.MaxEntries = defaultConfig.CacheSettings.MaxEntries
	}
	if config.CacheSettings.TTL == "" {
		config.CacheSettings.TTL = defaultConfig.CacheSettings.TTL
	}
	if config.CacheSettings.CleanupInterval == "" {
		config.CacheSettings.CleanupInterval = defaultConfig.CacheSettings.CleanupInterval
	}

	// 应用重试设置的默认值
	if config.RetrySettings.MaxRetries == 0 {
		config.RetrySettings.MaxRetries = defaultConfig.RetrySettings.MaxRetries
	}
	if config.RetrySettings.BaseDelay == "" {
		config.RetrySettings.BaseDelay = defaultConfig.RetrySettings.BaseDelay
	}
	if config.RetrySettings.MaxDelay == "" {
		config.RetrySettings.MaxDelay = defaultConfig.RetrySettings.MaxDelay
	}
	if config.RetrySettings.BackoffFactor == 0 {
		config.RetrySettings.BackoffFactor = defaultConfig.RetrySettings.BackoffFactor
	}

	// 应用性能设置的默认值
	if config.PerformanceSettings.BatchSize == 0 {
		config.PerformanceSettings.BatchSize = defaultConfig.PerformanceSettings.BatchSize
	}
	if config.PerformanceSettings.Timeout == "" {
		config.PerformanceSettings.Timeout = defaultConfig.PerformanceSettings.Timeout
	}
	if config.PerformanceSettings.MetricsInterval == "" {
		config.PerformanceSettings.MetricsInterval = defaultConfig.PerformanceSettings.MetricsInterval
	}

	// 应用降级设置的默认值
	if config.FallbackSettings.MaxFallbackResults == 0 {
		config.FallbackSettings.MaxFallbackResults = defaultConfig.FallbackSettings.MaxFallbackResults
	}
	if config.FallbackSettings.FallbackTimeout == "" {
		config.FallbackSettings.FallbackTimeout = defaultConfig.FallbackSettings.FallbackTimeout
	}

	// 应用日志设置的默认值
	if config.LoggingSettings.Level == "" {
		config.LoggingSettings.Level = defaultConfig.LoggingSettings.Level
	}
}

// IsRetryableError 检查错误是否在可重试错误列表中
func (rs *RetrySettings) IsRetryableError(errorType string) bool {
	for _, retryableError := range rs.RetryableErrors {
		if retryableError == errorType {
			return true
		}
	}
	return false
}

// GetLogLevelInt 将日志级别转换为数值
func (ls *LoggingSettings) GetLogLevelInt() int {
	levelMap := map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
	}

	if level, exists := levelMap[ls.Level]; exists {
		return level
	}
	return 1 // 默认为 info 级别
}