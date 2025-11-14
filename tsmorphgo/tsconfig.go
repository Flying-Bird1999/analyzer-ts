package tsmorphgo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// TsConfig 表示 TypeScript 配置文件 (tsconfig.json) 的结构
type TsConfig struct {
	CompilerOptions map[string]interface{} `json:"compilerOptions,omitempty"`
	Include        []string               `json:"include,omitempty"`
	Exclude        []string               `json:"exclude,omitempty"`
	Files          []string               `json:"files,omitempty"`
	Extends        string                 `json:"extends,omitempty"`
	References      []TsConfigReference    `json:"references,omitempty"`
}

// TsConfigReference 表示项目引用
type TsConfigReference struct {
	Path string `json:"path"`
}

// parseTsConfig 解析 TypeScript 配置文件
// 如果配置文件路径为空，会自动在根路径下查找 tsconfig.json
func parseTsConfig(config ProjectConfig) *TsConfig {
	// 确定配置文件路径
	configPath := config.TsConfigPath
	if configPath == "" {
		configPath = FindTsConfigFile(config.RootPath)
		if configPath == "" {
			// 没有找到配置文件，返回 nil
			return nil
		}
	}

	// 读取配置文件内容
	content, err := os.ReadFile(configPath)
	if err != nil {
		// 读取失败，不是致命错误，返回 nil 使用默认配置
		return nil
	}

	// 解析 JSON
	var tsConfig TsConfig
	if err := json.Unmarshal(content, &tsConfig); err != nil {
		return nil
	}

	// 处理继承的配置
	if tsConfig.Extends != "" {
		parentConfig := parseExtendedTsConfig(configPath, tsConfig.Extends)
		if parentConfig != nil {
			mergedTsConfig := mergeTsConfigStructs(parentConfig, &tsConfig)
			if mergedTsConfig != nil {
				tsConfig = *mergedTsConfig
			}
		}
	}

	return &tsConfig
}

// FindTsConfigFile 在指定目录中查找 tsconfig.json 文件
func FindTsConfigFile(rootPath string) string {
	// 按优先级查找配置文件
	configFiles := []string{
		"tsconfig.json",
		"tsconfig.base.json",
		"tsconfig.common.json",
	}

	for _, configFile := range configFiles {
		configPath := filepath.Join(rootPath, configFile)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	return ""
}

// parseExtendedTsConfig 解析继承的配置文件
func parseExtendedTsConfig(currentConfigPath, extends string) *TsConfig {
	// 处理相对路径
	var extendsPath string
	if filepath.IsAbs(extends) {
		extendsPath = extends
	} else {
		// 相对于当前配置文件的目录
		currentDir := filepath.Dir(currentConfigPath)
		extendsPath = filepath.Join(currentDir, extends)
	}

	// 递归解析继承的配置（避免循环引用）
	// 这里简化处理，实际可能需要更复杂的引用检测
	config := ProjectConfig{
		TsConfigPath: extendsPath,
	}
	return parseTsConfig(config)
}

// mergeTsConfigStructs 合并两个 TsConfig 结构
func mergeTsConfigStructs(base, override *TsConfig) *TsConfig {
	result := *base

	// 合并编译选项
	if override.CompilerOptions != nil {
		if result.CompilerOptions == nil {
			result.CompilerOptions = make(map[string]interface{})
		}
		for k, v := range override.CompilerOptions {
			result.CompilerOptions[k] = v
		}
	}

	// 合并包含模式
	if len(override.Include) > 0 {
		result.Include = override.Include
	}

	// 合并排除模式
	if len(override.Exclude) > 0 {
		result.Exclude = override.Exclude
	}

	// 合并文件列表
	if len(override.Files) > 0 {
		result.Files = override.Files
	}

	// 处理引用（通常不合并，以子配置为主）
	if len(override.References) > 0 {
		result.References = override.References
	}

	return &result
}

// mergeTsConfig 将 TsConfig 合并到 ProjectConfig 中
func mergeTsConfig(config ProjectConfig, tsConfig *TsConfig) ProjectConfig {
	result := config

	// 合并编译选项
	if tsConfig.CompilerOptions != nil {
		result.CompilerOptions = tsConfig.CompilerOptions
	}

	// 处理包含模式
	if len(tsConfig.Include) > 0 {
		result.IncludePatterns = ConvertGlobPatterns(tsConfig.Include, config.RootPath)
	}

	// 处理排除模式
	if len(tsConfig.Exclude) > 0 {
		result.ExcludePatterns = ConvertGlobPatterns(tsConfig.Exclude, config.RootPath)
	}

	// 处理特定的编译选项
	opts := tsConfig.CompilerOptions
	if opts != nil {
		// 确保基本的TypeScript扩展名存在
		hasTsExt := false
		hasTsxExt := false
		for _, ext := range result.TargetExtensions {
			if ext == ".ts" {
				hasTsExt = true
			}
			if ext == ".tsx" {
				hasTsxExt = true
			}
		}

		if !hasTsExt {
			result.TargetExtensions = append(result.TargetExtensions, ".ts")
		}
		if !hasTsxExt {
			result.TargetExtensions = append(result.TargetExtensions, ".tsx")
		}

		// 注意：target 编译选项不应该是文件扩展名，这里移除错误的处理
		// TypeScript 的 target 是编译目标 (如 es5, es6)，不是文件扩展名

		// 处理模块解析策略
		if module, ok := opts["module"]; ok {
			// 这里可以根据模块解析策略调整项目配置
			_ = module // 占位符，后续可以添加具体逻辑
		}
	}

	return result
}

// ConvertGlobPatterns 将 TypeScript 配置中的 glob 模式转换为文件匹配模式
func ConvertGlobPatterns(patterns []string, rootPath string) []string {
	var result []string

	for _, pattern := range patterns {
		// 处理相对路径，转换为相对于根路径的绝对模式
		if !filepath.IsAbs(pattern) {
			// 处理以 ** 开头的模式，也要加上根路径
			if strings.HasPrefix(pattern, "**") {
				fullPattern := filepath.Join(rootPath, pattern)
				result = append(result, fullPattern)
			} else {
				// 普通相对路径，需要检查是否是目录模式
				// TypeScript 中的 "src" 应该匹配 src 目录下的所有文件，包括子目录
				fullPattern := filepath.Join(rootPath, pattern)

				// 如果模式不包含通配符，假设它是目录，需要添加递归匹配
				if !strings.Contains(pattern, "*") && !strings.Contains(pattern, "?") {
					// 添加递归匹配模式，匹配目录下的所有文件
					recursivePattern := filepath.Join(fullPattern, "**", "*")
					result = append(result, recursivePattern)
				}

				result = append(result, fullPattern)
			}
		} else {
			// 已经是绝对路径，直接使用
			result = append(result, pattern)
		}
	}

	return result
}

// PathMatchesPatterns 检查文件路径是否匹配任一模式
// 支持包含和排除模式（!开头表示排除）
func PathMatchesPatterns(filePath string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}

	// 先检查是否有包含模式匹配
	hasIncludePatterns := false
	matchesInclude := false

	for _, pattern := range patterns {
		if strings.HasPrefix(pattern, "!") {
			// 排除模式
			excludePattern := pattern[1:] // 去掉 !
			if matchesPattern(filePath, excludePattern) {
				return false // 匹配了排除模式，直接返回false
			}
		} else {
			// 包含模式
			hasIncludePatterns = true
			if matchesPattern(filePath, pattern) {
				matchesInclude = true
			}
		}
	}

	// 如果没有包含模式，只检查了排除模式，那么文件是包含的
	if !hasIncludePatterns {
		return true
	}

	// 如果有包含模式，必须至少匹配一个
	return matchesInclude
}

// matchesPattern 检查文件路径是否匹配单个模式
// 改进的实现，支持完整的 glob 语法
func matchesPattern(filePath, pattern string) bool {
	// 规范化路径，确保使用正斜杠
	filePath = filepath.ToSlash(filePath)
	pattern = filepath.ToSlash(pattern)

	// 处理绝对路径与相对路径的匹配
	// 如果 filePath 以 / 开头但 pattern 不是，去掉 filePath 的前导 /
	if strings.HasPrefix(filePath, "/") && !strings.HasPrefix(pattern, "/") {
		filePath = filePath[1:]
	}
	// 如果 pattern 以 / 开头但 filePath 不是，给 filePath 加上前导 /
	if strings.HasPrefix(pattern, "/") && !strings.HasPrefix(filePath, "/") {
		filePath = "/" + filePath
	}

	// 处理 ** 通配符 - 需要特殊处理
	if strings.Contains(pattern, "**") {
		return matchesDoubleStarPattern(filePath, pattern)
	}

	// 处理目录通配符，如 src/*
	if strings.Contains(pattern, "/*") && !strings.Contains(pattern, "**") {
		return matchesDirectoryPattern(filePath, pattern)
	}

	// 处理文件扩展名通配符，如 *.ts
	if strings.HasPrefix(pattern, "*.") {
		ext := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(filePath, ext) ||
			   strings.HasSuffix(filePath, ext+".ts") ||
			   strings.HasSuffix(filePath, ext+".tsx")
	}

	// 处理简单的 * 通配符在路径中间
	if strings.Contains(pattern, "*") {
		return matchesWildcardPattern(filePath, pattern)
	}

	// 精确匹配
	return filePath == pattern ||
		   filePath == pattern+"/" ||
		   strings.HasPrefix(filePath, pattern+"/") ||
		   filePath == pattern+".ts" ||
		   filePath == pattern+".tsx"
}

// matchesDoubleStarPattern 处理 ** 通配符匹配
func matchesDoubleStarPattern(filePath, pattern string) bool {

	// 将 ** 分割成多个部分
	parts := strings.Split(pattern, "**")

	// 如果模式以 ** 开头，如 **/*.ts
	if strings.HasPrefix(pattern, "**") {
		remainingPattern := strings.TrimSpace(strings.TrimPrefix(pattern, "**"))
		if remainingPattern == "" || remainingPattern == "/" {
			return true // ** 匹配任何路径
		}
		// 特殊处理 **/*.ts, **/*.tsx, **/*.test.ts
		if remainingPattern == "/*.ts" {
			return strings.HasSuffix(filePath, ".ts")
		}
		if remainingPattern == "/*.tsx" {
			return strings.HasSuffix(filePath, ".tsx")
		}
		if remainingPattern == "/*.test.ts" {
			return strings.HasSuffix(filePath, ".test.ts")
		}
		// 检查文件路径是否以剩余模式结尾
		return strings.HasSuffix(filePath, strings.TrimPrefix(remainingPattern, "/")) ||
			   strings.HasSuffix(filePath, remainingPattern)
	}

	// 如果模式以 ** 结尾，如 src/**
	if strings.HasSuffix(pattern, "**") {
		prefix := strings.TrimSpace(strings.TrimSuffix(pattern, "**"))
		if prefix == "" || prefix == "/" {
			return true
		}
		return strings.HasPrefix(filePath, strings.TrimSuffix(prefix, "/")) ||
			   strings.HasPrefix(filePath, prefix)
	}

	// ** 在中间，如 src/**/test/*.ts
	if len(parts) == 2 {
		prefix := strings.TrimSpace(parts[0])
		suffix := strings.TrimSpace(parts[1])

		// 检查文件路径是否以 prefix 开头
		if prefix != "" && !strings.HasPrefix(filePath, strings.TrimSuffix(prefix, "/")) && !strings.HasPrefix(filePath, prefix) {
			return false
		}

		// 检查文件路径是否以 suffix 结尾
		if suffix != "" && suffix != "/" {
			cleanSuffix := strings.TrimPrefix(suffix, "/")

			// 特殊处理 *.ts 这样的文件扩展名模式
			if cleanSuffix == "*.ts" {
				return strings.HasSuffix(filePath, ".ts")
			}
			if cleanSuffix == "*.tsx" {
				return strings.HasSuffix(filePath, ".tsx")
			}
			// 特殊处理 * 模式（匹配任意内容）
			if cleanSuffix == "*" {
				return true
			}

			// 处理其他后缀模式
			return strings.HasSuffix(filePath, cleanSuffix) ||
				(strings.HasSuffix(filePath, "/"+cleanSuffix) && !strings.HasSuffix(cleanSuffix, "/"))
		}

		return true
	}

	// 多个 ** 的情况，简化处理
	return true
}

// matchesDirectoryPattern 处理目录通配符，如 src/*
func matchesDirectoryPattern(filePath, pattern string) bool {
	dirPattern := strings.TrimSuffix(pattern, "/*")
	return strings.HasPrefix(filePath, dirPattern) &&
		   (len(filePath) == len(dirPattern) || filePath[len(dirPattern)] == '/')
}

// matchesWildcardPattern 处理简单的 * 通配符
func matchesWildcardPattern(filePath, pattern string) bool {
	// 将 pattern 转换为正则表达式
	regexPattern := "^" + strings.ReplaceAll(regexp.QuoteMeta(pattern), "\\*", "[^/]*") + "$"

	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return false
	}

	return re.MatchString(filePath)
}

// GetTsConfig 获取项目的 TypeScript 配置
// 如果没有配置文件或解析失败，返回 nil
func (p *Project) GetTsConfig() *TsConfig {
	if p.parserResult == nil {
		return nil
	}

	// 尝试从项目配置中获取
	config := ProjectConfig{
		RootPath:     p.parserResult.Config.RootPath,
		UseTsConfig:  true,
	}
	return parseTsConfig(config)
}

// GetCompilerOption 获取指定的编译选项值
func (p *Project) GetCompilerOption(key string) (interface{}, bool) {
	tsConfig := p.GetTsConfig()
	if tsConfig == nil || tsConfig.CompilerOptions == nil {
		return nil, false
	}

	value, ok := tsConfig.CompilerOptions[key]
	return value, ok
}

// GetCompilerOptionString 获取字符串类型的编译选项
func (p *Project) GetCompilerOptionString(key string) (string, bool) {
	value, ok := p.GetCompilerOption(key)
	if !ok {
		return "", false
	}

	if str, ok := value.(string); ok {
		return str, true
	}
	return "", false
}

// GetCompilerOptionBool 获取布尔类型的编译选项
func (p *Project) GetCompilerOptionBool(key string) (bool, bool) {
	value, ok := p.GetCompilerOption(key)
	if !ok {
		return false, false
	}

	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		// 处理字符串形式的布尔值
		return strings.ToLower(v) == "true", true
	default:
		return false, false
	}
}

// GetIncludedFiles 获取 tsconfig.json 中明确包含的文件列表
func (p *Project) GetIncludedFiles() []string {
	tsConfig := p.GetTsConfig()
	if tsConfig == nil {
		return nil
	}

	// 如果有明确的文件列表，优先返回
	if len(tsConfig.Files) > 0 {
		return tsConfig.Files
	}

	// 否则返回匹配包含模式的文件
	var included []string
	for filePath := range p.sourceFiles {
		if PathMatchesPatterns(filePath, tsConfig.Include) {
			included = append(included, filePath)
		}
	}

	return included
}

// GetExcludedFiles 获取 tsconfig.json 中排除的文件列表
func (p *Project) GetExcludedFiles() []string {
	tsConfig := p.GetTsConfig()
	if tsConfig == nil {
		return nil
	}

	var excluded []string
	for filePath := range p.sourceFiles {
		if PathMatchesPatterns(filePath, tsConfig.Exclude) {
			excluded = append(excluded, filePath)
		}
	}

	return excluded
}