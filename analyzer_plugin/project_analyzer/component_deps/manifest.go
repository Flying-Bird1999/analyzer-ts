package component_deps

import (
	"encoding/json"
	"fmt"
	"os"
)

// =============================================================================
// 配置文件数据结构
// =============================================================================

// ComponentManifest 配置文件结构
// 对应 component-manifest.json 的格式
type ComponentManifest struct {
	Components map[string]ComponentDefinition `json:"components"` // 组件名 -> 组件定义
	Rules      *ManifestRules                   `json:"rules,omitempty"`
}

// ComponentDefinition 组件定义
type ComponentDefinition struct {
	Type string `json:"type"` // 资产类型: "component"
	Path string `json:"path"` // 组件目录路径
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

// =============================================================================
// 配置验证
// =============================================================================

// validateManifest 验证配置文件的有效性
func validateManifest(manifest *ComponentManifest) error {
	// 验证组件列表
	if len(manifest.Components) == 0 {
		return fmt.Errorf("components 不能为空")
	}

	for name, comp := range manifest.Components {
		// 验证组件名称
		if name == "" {
			return fmt.Errorf("组件名不能为空")
		}

		// 验证类型
		if comp.Type == "" {
			return fmt.Errorf("组件 '%s' 的 type 不能为空", name)
		}
		if comp.Type != "component" {
			return fmt.Errorf("组件 '%s' 的 type 必须为 'component'", name)
		}

		// 验证目录路径
		if comp.Path == "" {
			return fmt.Errorf("组件 '%s' 的 path 不能为空", name)
		}
	}

	return nil
}

// =============================================================================
// 辅助函数
// =============================================================================

// GetComponentByName 根据名称获取组件定义
func (m *ComponentManifest) GetComponentByName(name string) (*ComponentDefinition, bool) {
	comp, ok := m.Components[name]
	return &comp, ok
}

// GetComponentCount 获取组件数量
func (m *ComponentManifest) GetComponentCount() int {
	return len(m.Components)
}

// GetComponentNames 获取所有组件名称列表
func (m *ComponentManifest) GetComponentNames() []string {
	names := make([]string, 0, len(m.Components))
	for name := range m.Components {
		names = append(names, name)
	}
	return names
}

// GetAllComponents 返回所有组件的名称和定义
func (m *ComponentManifest) GetAllComponents() map[string]ComponentDefinition {
	return m.Components
}
