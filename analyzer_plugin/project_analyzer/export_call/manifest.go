// Package export_call 配置文件解析
package export_call

import (
	"encoding/json"
	"fmt"
	"os"
)

// AssetManifest 统一的资产清单配置文件
type AssetManifest struct {
	Components []AssetItem `json:"components"`
	Functions  []AssetItem `json:"functions"`
}

// AssetItem 单个资产项（组件或函数组）
type AssetItem struct {
	Name string `json:"name"` // 资产名称
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
	for i, comp := range manifest.Components {
		if comp.Name == "" {
			return fmt.Errorf("components[%d].name 不能为空", i)
		}
		if comp.Type != "component" {
			return fmt.Errorf("components[%d].type 必须为 'component'", i)
		}
		if comp.Path == "" {
			return fmt.Errorf("components[%d].path 不能为空", i)
		}
	}

	// 验证 functions
	for i, fn := range manifest.Functions {
		if fn.Name == "" {
			return fmt.Errorf("functions[%d].name 不能为空", i)
		}
		if fn.Type != "functions" {
			return fmt.Errorf("functions[%d].type 必须为 'functions'", i)
		}
		if fn.Path == "" {
			return fmt.Errorf("functions[%d].path 不能为空", i)
		}
	}

	return nil
}
