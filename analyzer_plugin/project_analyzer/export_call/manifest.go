// Package export_call 配置文件解析
package export_call

import (
	"encoding/json"
	"fmt"
	"os"
)

// AssetManifest 统一的资产清单配置文件
type AssetManifest struct {
	Components map[string]AssetItem `json:"components"` // 组件名 -> 组件定义
	Functions  map[string]AssetItem `json:"functions"`  // 函数名 -> 函数定义
}

// AssetItem 单个资产项（组件或函数组）
type AssetItem struct {
	Name string `json:"name"` // 资产名称（从 map key 复制，用于结果返回）
	Type string `json:"type"` // "component" | "functions"
	Path string `json:"path"` // 目录路径
}

// LoadAssetManifest 从指定路径加载资产清单配置文件
func LoadAssetManifest(manifestPath string) (*AssetManifest, error) {
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
	var manifest AssetManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证配置
	if err := validateManifest(&manifest); err != nil {
		return nil, fmt.Errorf("配置文件验证失败: %w", err)
	}

	return &manifest, nil
}

// validateManifest 验证配置文件的有效性
func validateManifest(manifest *AssetManifest) error {
	// 验证 components
	for name, comp := range manifest.Components {
		if name == "" {
			return fmt.Errorf("组件名不能为空")
		}
		if comp.Type != "component" {
			return fmt.Errorf("组件 '%s' 的 type 必须为 'component'", name)
		}
		if comp.Path == "" {
			return fmt.Errorf("组件 '%s' 的 path 不能为空", name)
		}
	}

	// 验证 functions
	for name, fn := range manifest.Functions {
		if name == "" {
			return fmt.Errorf("函数名不能为空")
		}
		if fn.Type != "functions" {
			return fmt.Errorf("函数 '%s' 的 type 必须为 'functions'", name)
		}
		if fn.Path == "" {
			return fmt.Errorf("函数 '%s' 的 path 不能为空", name)
		}
	}

	return nil
}

// GetComponentNames 获取所有组件名称列表
func (m *AssetManifest) GetComponentNames() []string {
	names := make([]string, 0, len(m.Components))
	for name := range m.Components {
		names = append(names, name)
	}
	return names
}

// GetFunctionNames 获取所有函数名称列表
func (m *AssetManifest) GetFunctionNames() []string {
	names := make([]string, 0, len(m.Functions))
	for name := range m.Functions {
		names = append(names, name)
	}
	return names
}
