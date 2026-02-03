package component_deps_v2

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// =============================================================================
// 配置文件数据结构
// =============================================================================

// ComponentManifest 配置文件结构
// 对应 component-manifest.json 的格式
type ComponentManifest struct {
	Components []ComponentDefinition `json:"components"`
	Rules      *ManifestRules        `json:"rules,omitempty"`
}

// ManifestMeta 配置文件元数据
type ManifestMeta struct {
	Version     string `json:"version"`     // 配置协议版本
	LibraryName string `json:"libraryName"` // 组件库名称
}

// ComponentDefinition 组件定义
type ComponentDefinition struct {
	Name  string `json:"name"`  // 组件名称
	Entry string `json:"entry"` // 组件入口文件
	// Scope 自动推断为 entry 文件所在目录
}

// ManifestRules 可选的规则配置
type ManifestRules struct {
	IgnorePatterns []string `json:"ignorePatterns,omitempty"` // 忽略的文件模式
}

// =============================================================================
// 配置加载
// =============================================================================

// LoadManifest 从指定路径加载组件配置文件
func LoadManifest(manifestPath string) (*ComponentManifest, error) {
	// 检查文件是否存在
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", manifestPath)
	}

	// 读取文件内容
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析 JSON
	var manifest ComponentManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证配置
	if err := validateManifest(&manifest); err != nil {
		return nil, fmt.Errorf("配置文件验证失败: %w", err)
	}

	return &manifest, nil
}

// LoadManifestFromProjectRoot 从项目根目录加载配置文件
// 尝试多个可能的配置文件位置
func LoadManifestFromProjectRoot(projectRoot string) (*ComponentManifest, error) {
	// 尝试的配置文件路径（按优先级）
	candidatePaths := []string{
		filepath.Join(projectRoot, "component-manifest.json"),
		filepath.Join(projectRoot, ".analyzer", "component-manifest.json"),
		filepath.Join(projectRoot, ".components", "manifest.json"),
	}

	for _, path := range candidatePaths {
		if _, err := os.Stat(path); err == nil {
			return LoadManifest(path)
		}
	}

	return nil, fmt.Errorf("未找到配置文件，已尝试以下路径: %v",
		candidatePaths)
}

// =============================================================================
// 配置验证
// =============================================================================

// validateManifest 验证配置文件的有效性
func validateManifest(manifest *ComponentManifest) error {
	// 验证组件列表
	if len(manifest.Components) == 0 {
		return fmt.Errorf("components 列表不能为空")
	}

	componentNames := make(map[string]bool)
	for i, comp := range manifest.Components {
		// 验证组件名称
		if comp.Name == "" {
			return fmt.Errorf("components[%d].name 不能为空", i)
		}
		// 检查组件名称重复
		if componentNames[comp.Name] {
			return fmt.Errorf("组件名称重复: %s", comp.Name)
		}
		componentNames[comp.Name] = true

		// 验证入口文件
		if comp.Entry == "" {
			return fmt.Errorf("components[%d].entry 不能为空", i)
		}
	}

	return nil
}

// =============================================================================
// 辅助函数
// =============================================================================

// GetComponentByName 根据名称获取组件定义
func (m *ComponentManifest) GetComponentByName(name string) (*ComponentDefinition, bool) {
	for i := range m.Components {
		if m.Components[i].Name == name {
			return &m.Components[i], true
		}
	}
	return nil, false
}

// GetComponentCount 获取组件数量
func (m *ComponentManifest) GetComponentCount() int {
	return len(m.Components)
}

// GetComponentNames 获取所有组件名称列表
func (m *ComponentManifest) GetComponentNames() []string {
	names := make([]string, len(m.Components))
	for i, comp := range m.Components {
		names[i] = comp.Name
	}
	return names
}
