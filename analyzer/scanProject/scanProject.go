// 从项目指定入口扫描文件
package scanProject

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gobwas/glob"
)

type ProjectResult struct {
	Root       string   // 入口
	Ignore     []string // 指定忽略的文件/文件夹
	IsMonorepo bool     // 是否为 monorepo 项目

	FileList map[string]FileItem // 文件列表
}

func NewProjectResult(root string, ignore []string, IsMonorepo bool) *ProjectResult {
	return &ProjectResult{
		Root:       root,
		Ignore:     ignore,
		IsMonorepo: IsMonorepo,
		FileList:   make(map[string]FileItem),
	}
}

func (pr *ProjectResult) GetFileList() map[string]FileItem {
	return pr.FileList
}

func (pr *ProjectResult) ScanProject() {
	pr.ScanFileList()
}

func (pr *ProjectResult) ScanFileList() {
	// 如果 pr.Ignore 为空，则使用默认的忽略规则
	var Ignore []string
	if len(pr.Ignore) > 0 {
		Ignore = pr.Ignore
	} else {
		Ignore = []string{
			"node_modules/**/*",
			".git/**/*",
			"**/__test__/**",
			"**/*.test.{ts,tsx,js,jsx}",
		}
	}

	// 编译忽略规则
	var ignoreGlobs []glob.Glob
	for _, pattern := range Ignore {
		g, _ := glob.Compile(pattern, '/')
		ignoreGlobs = append(ignoreGlobs, g)
	}

	// 遍历项目目录，获取所有文件列表
	err := filepath.Walk(pr.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("访问路径 %s 时出错: %s\n", path, err)
			return nil
		}

		relPath, _ := filepath.Rel(pr.Root, path)
		unixRelPath := filepath.ToSlash(relPath)

		// 检查是否匹配忽略规则
		for _, g := range ignoreGlobs {
			// 多包的case有点问题，这里先手动忽略掉 ode_modules
			if g.Match(unixRelPath) || info.Name() == "node_modules" {
				if info.IsDir() {
					return filepath.SkipDir // 跳过整个目录
				}
				return nil // 跳过文件
			}
		}

		// 检查是否是文件
		if !info.IsDir() {
			pr.FileList[path] = FileItem{
				FileName: info.Name(),
				Size:     info.Size(),        // 文件大小（字节）
				Ext:      filepath.Ext(path), // 文件后缀
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("扫描文件列表时出错: %s\n", err)
	}
}
