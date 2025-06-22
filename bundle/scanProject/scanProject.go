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
	NpmList  ProjectNpmList      // npm列表
}

func NewProjectResult(root string, ignore []string, IsMonorepo bool) *ProjectResult {
	return &ProjectResult{
		Root:       root,
		Ignore:     ignore,
		IsMonorepo: IsMonorepo,
		FileList:   make(map[string]FileItem),
		NpmList:    make(ProjectNpmList),
	}
}

func (pr *ProjectResult) GetFileList() map[string]FileItem {
	return pr.FileList
}

func (pr *ProjectResult) GetNpmList() ProjectNpmList {
	return pr.NpmList
}

func (pr *ProjectResult) ScanProject() {
	pr.ScanNpmList()
	pr.ScanFileList()
}

func (pr *ProjectResult) ScanNpmList() {
	if pr.IsMonorepo {
		// 扫描项目目录下所有的 package.json 文件
		err := filepath.Walk(pr.Root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("访问路径 %s 时出错: %s\n", path, err)
				return nil
			}

			// 明确跳过 node_modules 目录
			if info.IsDir() && info.Name() == "node_modules" {
				return filepath.SkipDir // 跳过整个 node_modules 目录
			}

			// 检查是否是 package.json 文件
			if info.Name() == "package.json" {
				// 解析 package.json 文件内容
				packageJsonInfo, err := GetPackageJson(path)
				if err != nil {
					fmt.Printf("解析 package.json 文件失败: %v\n", err)
					return nil
				}
				// 如果是最外层的 package.json，位于根目录下，则Workspace为root
				if filepath.Dir(path) == pr.Root {
					pr.NpmList["root"] = NpmPackage{
						Workspace: "root",
						Path:      path,
						Namespace: packageJsonInfo.Name,
						Version:   packageJsonInfo.Version,
						NpmList:   packageJsonInfo.NpmList,
					}
				} else {
					pr.NpmList[filepath.Base(filepath.Dir(path))] = NpmPackage{
						Workspace: filepath.Base(filepath.Dir(path)),
						Path:      path,
						Namespace: packageJsonInfo.Name,
						Version:   packageJsonInfo.Version,
						NpmList:   packageJsonInfo.NpmList,
					}
				}
			}
			return nil
		})

		if err != nil {
			fmt.Printf("扫描 package.json 文件时出错: %s\n", err)
		}
	} else {
		// 定义 package.json 文件路径
		packageJsonPath := fmt.Sprintf("%s/package.json", pr.Root)
		// 解析 package.json 文件内容
		packageJsonInfo, err := GetPackageJson(packageJsonPath)

		if err != nil {
			fmt.Printf("解析 package.json 文件失败: %v\n", err)
			return
		}

		pr.NpmList = ProjectNpmList{
			"root": NpmPackage{
				Workspace: "root",
				Path:      packageJsonPath,
				Namespace: packageJsonInfo.Name,
				Version:   packageJsonInfo.Version,
				NpmList:   packageJsonInfo.NpmList,
			},
		}
	}
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
