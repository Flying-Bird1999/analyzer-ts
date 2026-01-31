package component_deps_v2

import (
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
	"github.com/samber/lo"
)

// =============================================================================
// 组件作用域管理
// =============================================================================

// ComponentScope 组件作用域管理器
// 组件作用域基于 entry 文件所在目录自动推断
// 例如: entry = "src/Button/index.tsx" => 作用域 = "src/Button/**"
type ComponentScope struct {
	componentName string        // 组件名称
	componentDir  string        // 组件目录（entry 所在目录）
	pattern       glob.Glob     // 编译后的 glob 模式
	matchedFiles  map[string]bool // 已匹配的文件缓存
}

// NewComponentScope 创建组件作用域管理器
// 基于 entry 文件自动推断组件作用域
func NewComponentScope(comp *ComponentDefinition) *ComponentScope {
	// 获取 entry 所在目录作为组件目录
	componentDir := filepath.Dir(comp.Entry)

	// 创建 glob 模式：组件目录/**
	pattern := filepath.Join(componentDir, "**")
	g, err := glob.Compile(pattern, '/')
	if err != nil {
		// 如果编译失败，使用简单匹配
		g = glob.MustCompile(pattern)
	}

	return &ComponentScope{
		componentName: comp.Name,
		componentDir:  componentDir,
		pattern:       g,
		matchedFiles:  make(map[string]bool),
	}
}

// GetComponentDir 获取组件目录
func (s *ComponentScope) GetComponentDir() string {
	return s.componentDir
}

// Contains 检查文件是否在该组件的作用域内
// 作用域为组件目录下的所有文件
// filePath 需要是相对路径（相对于项目根目录）
func (s *ComponentScope) Contains(filePath string) bool {
	// 检查缓存
	if cached, ok := s.matchedFiles[filePath]; ok {
		return cached
	}

	// 使用 glob 模式匹配
	matched := s.pattern.Match(filePath)
	s.matchedFiles[filePath] = matched
	return matched
}

// =============================================================================
// 多组件作用域管理
// =============================================================================

// MultiComponentScope 多组件作用域管理器
// 管理多个组件的作用域，支持快速查找文件归属
type MultiComponentScope struct {
	components map[string]*ComponentScope // 组件名 -> 作用域
	projectRoot string                   // 项目根目录（用于路径转换）
	fileToComp map[string]string         // 文件路径 -> 组件名（缓存）
}

// NewMultiComponentScope 创建多组件作用域管理器
func NewMultiComponentScope(manifest *ComponentManifest, projectRoot string) *MultiComponentScope {
	m := &MultiComponentScope{
		components: make(map[string]*ComponentScope),
		projectRoot: projectRoot,
		fileToComp: make(map[string]string),
	}

	// 为每个组件创建作用域管理器
	for i := range manifest.Components {
		comp := &manifest.Components[i]
		m.components[comp.Name] = NewComponentScope(comp)
	}

	return m
}

// toRelativePath 将绝对路径转换为相对路径（相对于项目根目录）
func (m *MultiComponentScope) toRelativePath(absPath string) string {
	// 如果已经是相对路径（不是以 / 开头），直接返回
	if !strings.HasPrefix(absPath, "/") && !strings.Contains(absPath, ":") {
		return absPath
	}

	// 去掉项目根目录前缀
	if len(absPath) > len(m.projectRoot) && absPath[:len(m.projectRoot)] == m.projectRoot {
		remaining := absPath[len(m.projectRoot):]
		// 去掉开头的 /
		if strings.HasPrefix(remaining, "/") || strings.HasPrefix(remaining, "\\") {
			return remaining[1:]
		}
		return remaining
	}

	// 如果无法去掉前缀，返回原路径
	return absPath
}

// FindComponentByFile 查找文件所属的组件
// 如果文件属于多个组件作用域，返回第一个匹配的
// 如果文件不属于任何组件，返回 ("", false)
func (m *MultiComponentScope) FindComponentByFile(filePath string) (string, bool) {
	// 转换为相对路径
	relPath := m.toRelativePath(filePath)

	// 检查缓存
	if compName, ok := m.fileToComp[relPath]; ok {
		if compName == "" {
			return "", false
		}
		return compName, true
	}

	// 遍历所有组件作用域
	for compName, scope := range m.components {
		if scope.Contains(relPath) {
			m.fileToComp[relPath] = compName
			return compName, true
		}
	}

	// 未找到
	m.fileToComp[relPath] = ""
	return "", false
}

// GetScope 获取指定组件的作用域管理器
func (m *MultiComponentScope) GetScope(componentName string) (*ComponentScope, bool) {
	scope, ok := m.components[componentName]
	return scope, ok
}

// GetAllComponentNames 获取所有组件名称
func (m *MultiComponentScope) GetAllComponentNames() []string {
	names := make([]string, 0, len(m.components))
	for name := range m.components {
		names = append(names, name)
	}
	return names
}

// GetComponentDir 获取指定组件的目录
func (m *MultiComponentScope) GetComponentDir(componentName string) (string, bool) {
	if scope, ok := m.components[componentName]; ok {
		return scope.GetComponentDir(), true
	}
	return "", false
}

// GetMatchedFiles 从文件列表中获取属于该组件的文件
// 输入的文件路径可以是绝对路径或相对路径
func (m *MultiComponentScope) GetMatchedFiles(files []string) []string {
	return lo.Filter(files, func(path string, _ int) bool {
		compName, _ := m.FindComponentByFile(path)
		return compName != ""
	})
}

// =============================================================================
// 跨组件依赖检测
// =============================================================================

// DetectCrossComponentImports 检测导入语句是否为跨组件导入
// 返回: (目标组件名, 是否跨组件, 是否外部导入)
func (m *MultiComponentScope) DetectCrossComponentImports(importPath string, sourceFilePath string) (string, bool, bool) {
	// 查找源文件所属组件
	sourceComp, _ := m.FindComponentByFile(sourceFilePath)
	if sourceComp == "" {
		return "", false, true // 源文件不属于任何组件，视为外部
	}

	// 查找目标文件所属组件
	targetComp, ok := m.FindComponentByFile(importPath)
	if !ok {
		return "", false, true // 目标文件不属于任何组件，视为外部导入
	}

	// 同一个组件内
	if targetComp == sourceComp {
		return targetComp, false, false
	}

	// 跨组件导入
	return targetComp, true, false
}
