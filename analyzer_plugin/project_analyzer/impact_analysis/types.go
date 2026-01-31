package impact_analysis

import (
	"encoding/json"
	"os"
)

// =============================================================================
// 输入类型定义
// =============================================================================

// ChangeInput 变更输入
// 用于描述代码变更信息，支持多种输入方式
type ChangeInput struct {
	ModifiedFiles []string `json:"modifiedFiles"` // 修改的文件列表
	AddedFiles    []string `json:"addedFiles"`    // 新增的文件列表
	DeletedFiles  []string `json:"deletedFiles"`  // 删除的文件列表
}

// ChangeSourceType 变更来源类型
type ChangeSourceType string

const (
	ChangeSourceFile   ChangeSourceType = "file"   // 来自 JSON 文件
	ChangeSourceParams ChangeSourceType = "params" // 来自命令行参数
)

// ImpactSource 影响分析数据源配置
type ImpactSource struct {
	Type     ChangeSourceType `json:"type"`
	Value    string          `json:"value"` // 文件路径或 JSON 字符串
}

// DepsDataSource 依赖数据源配置
type DepsDataSource struct {
	Type  string `json:"type"`  // "file" - 从文件加载
	Value string `json:"value"` // 文件路径
}

// =============================================================================
// 配置解析
// =============================================================================

// ParseChangeInput 从字符串解析变更输入
func ParseChangeInput(jsonStr string) (*ChangeInput, error) {
	var input ChangeInput
	if err := json.Unmarshal([]byte(jsonStr), &input); err != nil {
		return nil, err
	}
	return &input, nil
}

// LoadChangeInput 从文件加载变更输入
func LoadChangeInput(filePath string) (*ChangeInput, error) {
	data, err := readFile(filePath)
	if err != nil {
		return nil, err
	}
	return ParseChangeInput(string(data))
}

// readFile 读取文件内容
func readFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

// =============================================================================
// 辅助函数
// =============================================================================

// GetAllFiles 获取变更涉及的所有文件
func (c *ChangeInput) GetAllFiles() []string {
	files := make([]string, 0)
	files = append(files, c.ModifiedFiles...)
	files = append(files, c.AddedFiles...)
	files = append(files, c.DeletedFiles...)
	return files
}

// GetFileCount 获取变更文件总数
func (c *ChangeInput) GetFileCount() int {
	return len(c.ModifiedFiles) + len(c.AddedFiles) + len(c.DeletedFiles)
}

// IsEmpty 检查是否为空变更
func (c *ChangeInput) IsEmpty() bool {
	return len(c.ModifiedFiles) == 0 &&
		   len(c.AddedFiles) == 0 &&
		   len(c.DeletedFiles) == 0
}
