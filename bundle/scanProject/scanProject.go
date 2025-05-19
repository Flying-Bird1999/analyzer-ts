package scanProject

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gobwas/glob"
)

type ProjectResult struct {
	Root   string   // 入口
	Ignore []string // 指定忽略的文件/文件夹

	FileList map[string]FileItem // 文件列表
	NpmList  map[string]NpmItem  // npm列表
}

func NewProjectResult(root string, ignore []string) *ProjectResult {
	return &ProjectResult{
		Root:     root,
		Ignore:   ignore,
		FileList: make(map[string]FileItem),
		NpmList:  make(map[string]NpmItem),
	}
}

func (pr *ProjectResult) GetFileList() map[string]FileItem {
	return pr.FileList
}

func (pr *ProjectResult) GetNpmList() map[string]NpmItem {
	return pr.NpmList
}

func (pr *ProjectResult) ScanProject() {
	pr.scanNpmList()
	pr.scanFileList()
}

func (pr *ProjectResult) scanNpmList() {
	// 定义 package.json 文件路径
	packageJsonPath := fmt.Sprintf("%s/package.json", pr.Root)
	// 解析 package.json 文件内容
	packageJsonMap, err := GetPackageJson(packageJsonPath)

	if err != nil {
		fmt.Printf("解析 package.json 文件失败: %v\n", err)
		pr.NpmList = make(map[string]NpmItem)
	}

	pr.NpmList = packageJsonMap
}

func (pr *ProjectResult) scanFileList() {
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
			if g.Match(unixRelPath) {
				if info.IsDir() {
					return filepath.SkipDir // 跳过整个目录
				}
				return nil // 跳过文件
			}
		}

		// 检查是否是文件
		if !info.IsDir() {
			pr.FileList[path] = FileItem{Path: path}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("扫描文件列表时出错: %s\n", err)
	}
}
